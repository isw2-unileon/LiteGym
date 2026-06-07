package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
)

// PasswordResetTokenRepository defines methods for managing password reset tokens.
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *model.PasswordResetToken) error
	GetByToken(ctx context.Context, tokenStr string) (*model.PasswordResetToken, error)
	DeleteByUserID(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string) error
}

type passwordResetTokenRepository struct {
	db *pgxpool.Pool
}

// NewPasswordResetTokenRepository creates a new postgres-based PasswordResetTokenRepository.
func NewPasswordResetTokenRepository(db *pgxpool.Pool) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(ctx context.Context, token *model.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)

	return err
}

func (r *passwordResetTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*model.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`
	var token model.PasswordResetToken
	err := r.db.QueryRow(ctx, query, tokenStr).Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *passwordResetTokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM password_reset_tokens WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *passwordResetTokenRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM password_reset_tokens WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
