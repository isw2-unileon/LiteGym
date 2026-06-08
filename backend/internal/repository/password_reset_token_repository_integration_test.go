//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
)

func TestPasswordResetTokenRepositoryIntegration(t *testing.T) {
	pool := setupTestDB(t)

	ctx := context.Background()
	userRepo := NewUserRepository(pool)
	tokenRepo := NewPasswordResetTokenRepository(pool)

	user := &model.User{
		Username:     "reset_token_test_user",
		Email:        "reset_token@example.com",
		PasswordHash: "password123",
	}

	err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token := &model.PasswordResetToken{
		UserID:    user.ID,
		Token:     "secure-reset-token-123",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	err = tokenRepo.Create(ctx, token)
	if err != nil {
		t.Fatalf("failed to create reset token: %v", err)
	}

	fetchedToken, err := tokenRepo.GetByToken(ctx, "secure-reset-token-123")
	if err != nil {
		t.Fatalf("failed to get token: %v", err)
	}

	if fetchedToken.Token != token.Token {
		t.Errorf("expected token %s, got %s", token.Token, fetchedToken.Token)
	}

	if fetchedToken.UserID != user.ID {
		t.Errorf("expected user id %s, got %s", user.ID, fetchedToken.UserID)
	}

	err = tokenRepo.Delete(ctx, fetchedToken.ID)
	if err != nil {
		t.Fatalf("failed to delete token: %v", err)
	}

	_, err = tokenRepo.GetByToken(ctx, "secure-reset-token-123")
	if err == nil {
		t.Fatal("expected error getting deleted token, got nil")
	}

	// Test DeleteByUserID
	token2 := &model.PasswordResetToken{
		UserID:    user.ID,
		Token:     "secure-reset-token-456",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	_ = tokenRepo.Create(ctx, token2)

	err = tokenRepo.DeleteByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to delete tokens by user id: %v", err)
	}

	_, err = tokenRepo.GetByToken(ctx, "secure-reset-token-456")
	if err == nil {
		t.Fatal("expected error getting deleted token by user id, got nil")
	}
}
