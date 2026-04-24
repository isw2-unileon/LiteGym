package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
	"golang.org/x/crypto/bcrypt"
)

type MockAuthUserRepository struct {
	getByIDFunc    func(ctx context.Context, id string) (*model.User, error)
	getByEmailFunc func(ctx context.Context, email string) (*model.User, error)
}

func (m *MockAuthUserRepository) Create(ctx context.Context, user *model.User) error {
	return nil
}

func (m *MockAuthUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockAuthUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	return nil, nil
}

func TestLoginSuccessSetsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mockRepo := &MockAuthUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				Username:     "testuser",
				Email:        email,
				Role:         "user",
				PasswordHash: string(hashedPassword),
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	userService := service.NewUserService(mockRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := NewAuthHandler(userService, tokenService, "auth_token", false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	authHandler.Login(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected auth cookie to be set")
	}

	cookie := cookies[0]
	if cookie.Name != "auth_token" {
		t.Errorf("expected cookie name auth_token, got %s", cookie.Name)
	}
	if !cookie.HttpOnly {
		t.Error("expected auth cookie to be HttpOnly")
	}
	if cookie.Value == "" {
		t.Error("expected auth cookie to contain a token")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mockRepo := &MockAuthUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				Username:     "testuser",
				Email:        email,
				Role:         "user",
				PasswordHash: string(hashedPassword),
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	userService := service.NewUserService(mockRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := NewAuthHandler(userService, tokenService, "auth_token", false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "wrong-password",
	}

	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	authHandler.Login(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestMeReturnsAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userService := service.NewUserService(&MockAuthUserRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{
				ID:        id,
				Username:  "testuser",
				Email:     "test@example.com",
				Role:      "admin",
				CreatedAt: time.Now(),
			}, nil
		},
	})
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := NewAuthHandler(userService, tokenService, "auth_token", false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440000")
	c.Request = httptest.NewRequest("GET", "/api/auth/me", nil)

	authHandler.Me(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var payload struct {
		User model.User `json:"user"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if payload.User.Role != "admin" {
		t.Fatalf("expected role admin, got %s", payload.User.Role)
	}
}

func TestLogoutClearsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userService := service.NewUserService(&MockAuthUserRepository{})
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := NewAuthHandler(userService, tokenService, "auth_token", false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/auth/logout", nil)

	authHandler.Logout(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected auth cookie to be cleared")
	}

	cookie := cookies[0]
	if cookie.Name != "auth_token" {
		t.Errorf("expected cookie name auth_token, got %s", cookie.Name)
	}
	if cookie.MaxAge != -1 {
		t.Errorf("expected MaxAge -1, got %d", cookie.MaxAge)
	}
}
