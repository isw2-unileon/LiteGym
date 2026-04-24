package repository

import (
	"context"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	ListAll(ctx context.Context) ([]*model.User, error)
	Delete(ctx context.Context, id string) error
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
		INSERT INTO users (username, email, password_hash, role, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, created_at
	`
	// Provide default values if they are empty
	role := user.Role
	if role == "" {
		role = "user"
	}

	err := r.db.QueryRow(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		role,
		true, // is_active by default

	).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return err
	}

	if user.Role == "" {
		user.Role = "user"
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id::text, username, email, password_hash, role, is_active,created_at
		FROM users
		WHERE id = $1::uuid
	`

	var user model.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id::text, username, email, password_hash, created_at, role, is_active
		FROM users
		WHERE email = $1
	`

	var user model.User

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.Role,
		&user.IsActive,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ListAll retrieves all users from the database, ordered by creation date.
func (r *userRepository) ListAll(ctx context.Context) ([]*model.User, error) {
	// 1. Nos aseguramos de pedir el id como texto (id::text)
	query := `
		SELECT id::text, username, email, password_hash, role, is_active, created_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User

	for rows.Next() {
		var user model.User

		// 2. Escaneamos TODO directamente al modelo (sin inventar variables id64)
		err := rows.Scan(
			&user.ID, // Directo a string, como debe ser
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Delete removes a user from the database by their ID.
func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
