package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// TokenService manages token generation and verification.
type TokenService struct {
	secretKey []byte
	issuer    string
	ttl       time.Duration
}

// TokenClaims represents the JWT claims used by TokenService.
//
// TokenService emits and parses tokens containing these claims. The struct is
// exported because tests and middleware may inspect the parsed claims to
// access the subject, email, username and timing fields.
type TokenClaims struct {
	Subject   string `json:"sub"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Issuer    string `json:"iss"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type tokenHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

// NewTokenService creates a new TokenService.
func NewTokenService(secret, issuer string, ttl time.Duration) *TokenService {
	return &TokenService{
		secretKey: []byte(secret),
		issuer:    issuer,
		ttl:       ttl,
	}
}

// GenerateToken creates a signed token for the given user data.
func (s *TokenService) GenerateToken(userID int, email, username string) (string, error) {
	if len(s.secretKey) == 0 {
		return "", errors.New("token secret is required")
	}

	now := time.Now().UTC()
	header := tokenHeader{
		Algorithm: "HS256",
		Type:      "JWT",
	}
	claims := TokenClaims{
		Subject:   fmt.Sprintf("%d", userID),
		Email:     email,
		Username:  username,
		Issuer:    s.issuer,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(s.ttl).Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerPart := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsPart := base64.RawURLEncoding.EncodeToString(claimsJSON)
	unsignedToken := headerPart + "." + claimsPart

	mac := hmac.New(sha256.New, s.secretKey)
	if _, err := mac.Write([]byte(unsignedToken)); err != nil {
		return "", err
	}

	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return unsignedToken + "." + signature, nil
}

// TTL returns the configured token lifetime.
func (s *TokenService) TTL() time.Duration {
	return s.ttl
}

// VerifyToken validates the token signature and expiration.
func (s *TokenService) VerifyToken(token string) error {
	_, err := s.ParseToken(token)
	return err
}

// ParseToken validates the token and returns its claims.
func (s *TokenService) ParseToken(token string) (*TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	unsignedToken := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, s.secretKey)
	if _, err := mac.Write([]byte(unsignedToken)); err != nil {
		return nil, err
	}

	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return nil, errors.New("invalid token signature")
	}

	var claims TokenClaims
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	if time.Now().UTC().Unix() > claims.ExpiresAt {
		return nil, errors.New("token expired")
	}

	if s.issuer != "" && claims.Issuer != s.issuer {
		return nil, errors.New("invalid token issuer")
	}

	return &claims, nil
}
