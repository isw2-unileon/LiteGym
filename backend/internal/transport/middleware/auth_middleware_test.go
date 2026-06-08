package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/jackc/pgx/v5"
)

type staleTokenUserRepository struct{}

func (r *staleTokenUserRepository) Create(ctx context.Context, user *model.User) error {
	return nil
}

func (r *staleTokenUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	return nil, pgx.ErrNoRows
}

func (r *staleTokenUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, pgx.ErrNoRows
}

func (r *staleTokenUserRepository) ListAll(ctx context.Context) ([]*model.User, error) {
	return []*model.User{}, nil
}

func (r *staleTokenUserRepository) Delete(ctx context.Context, id string) error { return nil }
func (r *staleTokenUserRepository) MarkAsVerified(ctx context.Context, id string) error { return nil }
func (r *staleTokenUserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error { return nil }

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

func TestRequireAuthRejectsTokenForMissingUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	userService := service.NewUserService(&staleTokenUserRepository{})
	authMiddleware := NewAuthMiddleware(tokenService, "auth_token", userService)

	token, err := tokenService.GenerateToken("11111111-1111-1111-1111-111111111111", "missing@example.com", "missing", "user")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "user not found") {
		t.Fatalf("expected user not found error, got %s", w.Body.String())
	}
}
