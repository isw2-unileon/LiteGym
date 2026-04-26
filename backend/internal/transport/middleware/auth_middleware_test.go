package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
)

func TestRequireAuthWithoutCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authMiddleware := NewAuthMiddleware(tokenService, "auth_token")

	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequireAuthWithInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authMiddleware := NewAuthMiddleware(tokenService, "auth_token")

	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: "invalid-token",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequireAuthWithValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	token, err := tokenService.GenerateToken("550e8400-e29b-41d4-a716-446655440000", "test@example.com", "testuser", "user")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	authMiddleware := NewAuthMiddleware(tokenService, "auth_token")

	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		userID, ok := c.Get(ContextUserIDKey)
		if !ok {
			t.Fatal("expected user_id in context")
		}
		role, ok := c.Get(ContextUserRoleKey)
		if !ok {
			t.Fatal("expected user_role in context")
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"role":    role,
		})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if gotRole, _ := body["role"].(string); gotRole != "user" {
		t.Fatalf("expected role user, got %v", body["role"])
	}
}
