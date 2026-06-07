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
)

type mockPasswordResetTokenRepository struct {
	createFunc         func(ctx context.Context, token *model.PasswordResetToken) error
	getByTokenFunc     func(ctx context.Context, tokenStr string) (*model.PasswordResetToken, error)
	deleteByUserIDFunc func(ctx context.Context, userID string) error
	deleteFunc         func(ctx context.Context, id string) error
}

func (m *mockPasswordResetTokenRepository) Create(ctx context.Context, token *model.PasswordResetToken) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, token)
	}
	return nil
}

func (m *mockPasswordResetTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*model.PasswordResetToken, error) {
	if m.getByTokenFunc != nil {
		return m.getByTokenFunc(ctx, tokenStr)
	}
	return nil, nil
}

func (m *mockPasswordResetTokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	if m.deleteByUserIDFunc != nil {
		return m.deleteByUserIDFunc(ctx, userID)
	}
	return nil
}

func (m *mockPasswordResetTokenRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

type mockEmailService struct {
	sendPasswordResetFunc func(toEmail, username, resetLink string) error
}

func (m *mockEmailService) SendVerificationEmail(toEmail, username, verificationLink string) error {
	return nil
}

func (m *mockEmailService) SendPasswordResetEmail(toEmail, username, resetLink string) error {
	if m.sendPasswordResetFunc != nil {
		return m.sendPasswordResetFunc(toEmail, username, resetLink)
	}
	return nil
}

func TestForgotPasswordSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUserRepo := &MockAuthUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "user-id", Email: email, Username: "testuser"}, nil
		},
	}
	mockTokenRepo := &mockPasswordResetTokenRepository{}
	mockEmailSvc := &mockEmailService{}

	recoverySvc := service.NewPasswordRecoveryService(mockUserRepo, mockTokenRepo, mockEmailSvc, "http://frontend")

	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := NewAuthHandler(userService, tokenService, nil, recoverySvc, "auth_token", false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := ForgotPasswordRequest{Email: "test@example.com"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/auth/forgot-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	authHandler.ForgotPassword(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestResetPasswordSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUserRepo := &MockAuthUserRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: "user-id", Email: "test@example.com"}, nil
		},
	}
	mockTokenRepo := &mockPasswordResetTokenRepository{
		getByTokenFunc: func(ctx context.Context, tokenStr string) (*model.PasswordResetToken, error) {
			return &model.PasswordResetToken{
				ID:        "token-id",
				UserID:    "user-id",
				Token:     "valid-token",
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}, nil
		},
	}

	recoverySvc := service.NewPasswordRecoveryService(mockUserRepo, mockTokenRepo, &mockEmailService{}, "")

	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := NewAuthHandler(userService, tokenService, nil, recoverySvc, "auth_token", false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := ResetPasswordRequest{Token: "valid-token", NewPassword: "newpassword123"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/auth/reset-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	authHandler.ResetPassword(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
