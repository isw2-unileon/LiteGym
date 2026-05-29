package service

import (
	"context"
	"errors"
	"strings"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidUserInput indicates that the provided user data is invalid.
	ErrInvalidUserInput = errors.New("invalid user input")

	// ErrInvalidCredentials indicates that the provided credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserNotFound indicates that the requested user does not exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists indicates that the provided username or email is already registered.
	ErrUserAlreadyExists = errors.New("user already exists")
)

// UserService handles user-related business logic.
type UserService struct {
	repo repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// Create validates and creates a new user.
func (s *UserService) Create(ctx context.Context, user *model.User) error {
	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.PasswordHash = strings.TrimSpace(user.PasswordHash)

	if user.Username == "" || user.Email == "" || user.PasswordHash == "" {
		return ErrInvalidUserInput
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)

	if err := s.repo.Create(ctx, user); err != nil {
		if isUniqueViolation(err) {
			return ErrUserAlreadyExists
		}

		return err
	}

	return nil
}

// GetByID retrieves a user by ID.
func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidUserInput
	}

	user, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Authenticate validates credentials and returns the authenticated user.
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*model.User, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if email == "" || password == "" {
		return nil, ErrInvalidUserInput
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// ListAll retrieves all users from the system.
// This is primarily used for the Admin View dashboard.
func (s *UserService) ListAll(ctx context.Context) ([]*model.User, error) {
	users, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	if users == nil {
		return []*model.User{}, nil
	}

	return users, nil
}

// Delete removes a user by ID.
func (s *UserService) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}
	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
