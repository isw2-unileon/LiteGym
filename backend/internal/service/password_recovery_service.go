package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
)

var (
	// ErrInvalidResetToken is returned when a reset token is invalid or expired.
	ErrInvalidResetToken = errors.New("invalid or expired password reset token")
)

// PasswordRecoveryService manages the logic for recovering user passwords.
type PasswordRecoveryService struct {
	userRepo      repository.UserRepository
	tokenRepo     repository.PasswordResetTokenRepository
	emailService  EmailService
	frontendURL   string
}

// NewPasswordRecoveryService creates a new PasswordRecoveryService.
func NewPasswordRecoveryService(
	userRepo repository.UserRepository,
	tokenRepo repository.PasswordResetTokenRepository,
	emailService EmailService,
	frontendURL string,
) *PasswordRecoveryService {
	return &PasswordRecoveryService{
		userRepo:      userRepo,
		tokenRepo:     tokenRepo,
		emailService:  emailService,
		frontendURL:   frontendURL,
	}
}

// RequestPasswordReset generates a reset token and sends a recovery email.
func (s *PasswordRecoveryService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Prevent email enumeration by returning nil even if the user does not exist.
		return nil
	}

	// Generate a secure random token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	tokenStr := base64.URLEncoding.EncodeToString(b)

	// Save token with 1 hour expiration
	token := &model.PasswordResetToken{
		UserID:    user.ID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Clean up old tokens for this user
	_ = s.tokenRepo.DeleteByUserID(ctx, user.ID)

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	// Send email
	link := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, tokenStr)
	if err := s.emailService.SendPasswordResetEmail(user.Email, user.Username, link); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// ResetPassword updates the user's password if the token is valid.
func (s *PasswordRecoveryService) ResetPassword(ctx context.Context, tokenStr string, newPassword string) error {
	token, err := s.tokenRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		return ErrInvalidResetToken
	}

	if time.Now().After(token.ExpiresAt) {
		_ = s.tokenRepo.Delete(ctx, token.ID)
		return ErrInvalidResetToken
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, token.UserID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Cleanup the token after successful reset
	_ = s.tokenRepo.DeleteByUserID(ctx, token.UserID)

	return nil
}
