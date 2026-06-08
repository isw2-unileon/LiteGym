package repository

import (
	"context"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VerificationTokenRepository defines methods for managing user verification tokens.
type VerificationTokenRepository interface {
	Create(ctx context.Context, token *model.UserVerificationToken) error
	GetByToken(ctx context.Context, tokenStr string) (*model.UserVerificationToken, error)
	DeleteByUserID(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string) error
}

type verificationTokenRepository struct {
	db *pgxpool.Pool
}

// NewVerificationTokenRepository creates a new postgres-based VerificationTokenRepository.
func NewVerificationTokenRepository(db *pgxpool.Pool) VerificationTokenRepository {
	return &verificationTokenRepository{db: db}
}

func (r *verificationTokenRepository) Create(ctx context.Context, token *model.UserVerificationToken) error {
	query := `
		INSERT INTO user_verification_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id::text, created_at
	`

	return r.db.QueryRow(
		ctx,
		query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}

func (r *verificationTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*model.UserVerificationToken, error) {
	query := `
		SELECT id::text, user_id::text, token, expires_at, created_at
		FROM user_verification_tokens
		WHERE token = $1
	`

	var token model.UserVerificationToken
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

func (r *verificationTokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM user_verification_tokens WHERE user_id = $1::uuid`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *verificationTokenRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM user_verification_tokens WHERE id = $1::uuid`
	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
