package service

import (
	"context"
	"errors"
	"strings"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUserInput   = errors.New("invalid user input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) Create(ctx context.Context, user *model.User) error {
	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)
	user.PasswordHash = strings.TrimSpace(user.PasswordHash)

	if user.Username == "" || user.Email == "" || user.PasswordHash == "" {
		return ErrInvalidUserInput
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)

	return s.repo.Create(ctx, user)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

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
