package service

import (
	"context"
	"errors"
	"net/mail"
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

	// ErrEmailAlreadyExists indicates that the provided email is already registered.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrInvalidEmail indicates that the provided email does not have a valid format.
	ErrInvalidEmail = errors.New("invalid email")
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
	user.PasswordHash = strings.TrimSpace(user.PasswordHash)

	normalizedEmail, err := normalizeEmail(user.Email)
	if err != nil {
		return err
	}

	user.Email = normalizedEmail

	if user.Username == "" || user.PasswordHash == "" {
		return ErrInvalidUserInput
	}

	existingUser, err := s.repo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
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
	normalizedEmail, err := normalizeEmail(email)
	password = strings.TrimSpace(password)

	if err != nil {
		return nil, err
	}

	if password == "" {
		return nil, ErrInvalidUserInput
	}

	user, err := s.repo.GetByEmail(ctx, normalizedEmail)
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

func normalizeEmail(rawEmail string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(rawEmail))
	if email == "" {
		return "", ErrInvalidUserInput
	}

	parsedAddress, err := mail.ParseAddress(email)
	if err != nil || parsedAddress.Address != email {
		return "", ErrInvalidEmail
	}

	localPart, domain, found := strings.Cut(email, "@")
	if !found || localPart == "" || domain == "" || strings.Contains(domain, "@") {
		return "", ErrInvalidEmail
	}

	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") || strings.Contains(domain, "..") {
		return "", ErrInvalidEmail
	}

	return email, nil
}
