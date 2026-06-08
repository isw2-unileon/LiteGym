package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
)

var (
	// ErrInvalidToken is returned when a verification token is invalid or expired.
	ErrInvalidToken = errors.New("invalid or expired token")
)

// VerificationService manages the logic for verifying user accounts.
type VerificationService struct {
	userRepo       repository.UserRepository
	tokenRepo      repository.VerificationTokenRepository
	emailService   EmailService
	frontendURL    string
}

// NewVerificationService creates a new VerificationService.
func NewVerificationService(
	userRepo repository.UserRepository,
	tokenRepo repository.VerificationTokenRepository,
	emailService EmailService,
	frontendURL string,
) *VerificationService {
	return &VerificationService{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		emailService:   emailService,
		frontendURL:    frontendURL,
	}
}

// GenerateAndSendVerificationToken creates a new token and sends the verification email.
func (s *VerificationService) GenerateAndSendVerificationToken(ctx context.Context, user *model.User) error {
	// Generate random token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return err
	}
	tokenStr := base64.URLEncoding.EncodeToString(b)

	// Save token
	token := &model.UserVerificationToken{
		UserID:    user.ID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Delete old tokens for the user to prevent buildup
	_ = s.tokenRepo.DeleteByUserID(ctx, user.ID)

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send email
	link := fmt.Sprintf("%s/verify-email?token=%s", s.frontendURL, tokenStr)
	if err := s.emailService.SendVerificationEmail(user.Email, user.Username, link); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// VerifyToken validates a token and marks the associated user as verified.
func (s *VerificationService) VerifyToken(ctx context.Context, tokenStr string) error {
	token, err := s.tokenRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		return ErrInvalidToken
	}

	if time.Now().After(token.ExpiresAt) {
		// Delete expired token
		_ = s.tokenRepo.Delete(ctx, token.ID)
		return ErrInvalidToken
	}

	if err := s.userRepo.MarkAsVerified(ctx, token.UserID); err != nil {
		return fmt.Errorf("failed to mark user as verified: %w", err)
	}

	// Clean up token
	_ = s.tokenRepo.Delete(ctx, token.ID)

	return nil
}

// ResendVerificationEmail resends the verification email to a user if not already verified.
func (s *VerificationService) ResendVerificationEmail(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user.IsVerified {
		return errors.New("user is already verified")
	}

	return s.GenerateAndSendVerificationToken(ctx, user)
}
