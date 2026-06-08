package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
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

func TestRequestPasswordResetSuccess(t *testing.T) {
	userRepo := &mockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "user-id", Email: email, Username: "test"}, nil
		},
	}
	tokenRepo := &mockPasswordResetTokenRepository{
		createFunc: func(ctx context.Context, token *model.PasswordResetToken) error {
			if token.UserID != "user-id" {
				t.Errorf("expected user-id, got %s", token.UserID)
			}
			if len(token.Token) == 0 {
				t.Error("expected token to be generated")
			}
			return nil
		},
	}
	emailSent := false
	emailSvc := &mockEmailService{
		sendPasswordResetFunc: func(toEmail, username, resetLink string) error {
			emailSent = true
			return nil
		},
	}

	svc := NewPasswordRecoveryService(userRepo, tokenRepo, emailSvc, "http://frontend")

	err := svc.RequestPasswordReset(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !emailSent {
		t.Error("expected email to be sent")
	}
}

func TestRequestPasswordResetUserNotFound(t *testing.T) {
	userRepo := &mockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return nil, errors.New("not found")
		},
	}

	svc := NewPasswordRecoveryService(userRepo, &mockPasswordResetTokenRepository{}, &mockEmailService{}, "")

	err := svc.RequestPasswordReset(context.Background(), "unknown@example.com")
	if err != nil {
		t.Fatalf("expected nil error (prevent enumeration), got %v", err)
	}
}

func TestResetPasswordSuccess(t *testing.T) {
	userRepo := &mockUserRepository{
		updatePasswordFunc: func(ctx context.Context, id string, passwordHash string) error {
			if id != "user-id" {
				t.Errorf("expected user-id, got %s", id)
			}
			if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("newpass")); err != nil {
				t.Errorf("expected correct hash: %v", err)
			}
			return nil
		},
	}

	tokenRepo := &mockPasswordResetTokenRepository{
		getByTokenFunc: func(ctx context.Context, tokenStr string) (*model.PasswordResetToken, error) {
			return &model.PasswordResetToken{
				ID:        "token-id",
				UserID:    "user-id",
				Token:     "valid-token",
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}, nil
		},
	}

	svc := NewPasswordRecoveryService(userRepo, tokenRepo, &mockEmailService{}, "")

	err := svc.ResetPassword(context.Background(), "valid-token", "newpass")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestResetPasswordExpiredToken(t *testing.T) {
	tokenRepo := &mockPasswordResetTokenRepository{
		getByTokenFunc: func(ctx context.Context, tokenStr string) (*model.PasswordResetToken, error) {
			return &model.PasswordResetToken{
				ID:        "token-id",
				UserID:    "user-id",
				Token:     "expired-token",
				ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
			}, nil
		},
	}

	svc := NewPasswordRecoveryService(&mockUserRepository{}, tokenRepo, &mockEmailService{}, "")

	err := svc.ResetPassword(context.Background(), "expired-token", "newpass")
	if !errors.Is(err, ErrInvalidResetToken) {
		t.Fatalf("expected ErrInvalidResetToken, got %v", err)
	}
}
