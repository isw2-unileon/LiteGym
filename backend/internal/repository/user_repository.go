package repository

import (
	"context"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id::text, role::text, created_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
	).Scan(&user.ID, &user.Role, &user.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id::text, username, email, role::text, password_hash, created_at
		FROM users
		WHERE id = $1::uuid
	`

	var user model.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id::text, username, email, role::text, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	var user model.User

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
