package repositories

import (
	"context"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	var user model.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}