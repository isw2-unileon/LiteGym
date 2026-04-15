package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewTokenServiceTTL(t *testing.T) {
	ttl := 2 * time.Hour
	svc := NewTokenService("my-secret", "my-app", ttl)

	if svc == nil {
		t.Fatal("expected service, got nil")
	}

	if svc.TTL() != ttl {
		t.Errorf("expected TTL %v, got %v", ttl, svc.TTL())
	}
}

func TestGenerateTokenSuccess(t *testing.T) {
	svc := NewTokenService("my-secret", "my-app", time.Hour)

	token, err := svc.GenerateToken(1, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected token with 3 parts, got %d", len(parts))
	}

	if err := svc.VerifyToken(token); err != nil {
		t.Errorf("expected generated token to be valid, got %v", err)
	}
}

func TestGenerateTokenWithoutSecret(t *testing.T) {
	svc := NewTokenService("", "my-app", time.Hour)

	token, err := svc.GenerateToken(1, "test@example.com", "testuser")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if token != "" {
		t.Errorf("expected empty token, got %q", token)
	}
}

func TestGenerateTokenContainsExpectedClaims(t *testing.T) {
	svc := NewTokenService("my-secret", "my-app", time.Hour)

	token, err := svc.GenerateToken(42, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected token with 3 parts, got %d", len(parts))
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("failed to unmarshal claims: %v", err)
	}

	if claims.Subject != "42" {
		t.Errorf("expected subject 42, got %s", claims.Subject)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", claims.Email)
	}

	if claims.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", claims.Username)
	}

	if claims.Issuer != "my-app" {
		t.Errorf("expected issuer my-app, got %s", claims.Issuer)
	}

	if claims.IssuedAt == 0 {
		t.Error("expected issued at to be set")
	}

	if claims.ExpiresAt == 0 {
		t.Error("expected expires at to be set")
	}

	if claims.ExpiresAt <= claims.IssuedAt {
		t.Errorf("expected exp > iat, got exp=%d iat=%d", claims.ExpiresAt, claims.IssuedAt)
	}
}

func TestVerifyTokenInvalidFormat(t *testing.T) {
	svc := NewTokenService("my-secret", "my-app", time.Hour)

	err := svc.VerifyToken("invalid-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVerifyTokenInvalidSignature(t *testing.T) {
	svc := NewTokenService("my-secret", "my-app", time.Hour)

	token, err := svc.GenerateToken(1, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected token with 3 parts, got %d", len(parts))
	}

	tamperedToken := parts[0] + "." + parts[1] + ".invalidsignature"

	err = svc.VerifyToken(tamperedToken)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVerifyTokenInvalidPayloadBase64(t *testing.T) {
	svc := NewTokenService("my-secret", "my-app", time.Hour)

	token, err := svc.GenerateToken(1, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected token with 3 parts, got %d", len(parts))
	}

	invalidPayloadToken := parts[0] + ".###." + parts[2]

	err = svc.VerifyToken(invalidPayloadToken)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVerifyTokenInvalidPayloadJSON(t *testing.T) {
	svc := NewTokenService("my-secret", "my-app", time.Hour)

	header := tokenHeader{
		Algorithm: "HS256",
		Type:      "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("failed to marshal header: %v", err)
	}

	headerPart := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadPart := base64.RawURLEncoding.EncodeToString([]byte("not-json"))
	unsignedToken := headerPart + "." + payloadPart

	mac := hmac.New(sha256.New, []byte("my-secret"))
	if _, err := mac.Write([]byte(unsignedToken)); err != nil {
		t.Fatalf("failed to write HMAC: %v", err)
	}

	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	token := unsignedToken + "." + signature

	err = svc.VerifyToken(token)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVerifyTokenExpired(t *testing.T) {
	svc := NewTokenService("my-secret", "my-app", -1*time.Hour)

	token, err := svc.GenerateToken(1, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	err = svc.VerifyToken(token)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVerifyTokenWithDifferentSecret(t *testing.T) {
	generator := NewTokenService("secret-one", "my-app", time.Hour)
	verifier := NewTokenService("secret-two", "my-app", time.Hour)

	token, err := generator.GenerateToken(1, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	err = verifier.VerifyToken(token)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
