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

type TokenService struct {
	secretKey []byte
	issuer    string
	ttl       time.Duration
}

type tokenHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type tokenClaims struct {
	Subject   string `json:"sub"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Issuer    string `json:"iss"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func NewTokenService(secret, issuer string, ttl time.Duration) *TokenService {
	return &TokenService{
		secretKey: []byte(secret),
		issuer:    issuer,
		ttl:       ttl,
	}
}

func (s *TokenService) GenerateToken(userID int, email, username string) (string, error) {
	if len(s.secretKey) == 0 {
		return "", errors.New("token secret is required")
	}

	now := time.Now().UTC()
	header := tokenHeader{
		Algorithm: "HS256",
		Type:      "JWT",
	}
	claims := tokenClaims{
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

func (s *TokenService) TTL() time.Duration {
	return s.ttl
}

func (s *TokenService) VerifyToken(token string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("invalid token format")
	}

	unsignedToken := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, s.secretKey)
	if _, err := mac.Write([]byte(unsignedToken)); err != nil {
		return err
	}

	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return errors.New("invalid token signature")
	}

	var claims tokenClaims
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return err
	}

	if time.Now().UTC().Unix() > claims.ExpiresAt {
		return errors.New("token expired")
	}

	return nil
}
