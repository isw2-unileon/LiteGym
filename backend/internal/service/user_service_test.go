package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	createFunc     func(ctx context.Context, user *model.User) error
	getByIDFunc    func(ctx context.Context, id string) (*model.User, error)
	getByEmailFunc func(ctx context.Context, email string) (*model.User, error)
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	return nil, nil
}

func TestUserServiceCreateHashesPassword(t *testing.T) {
	mockRepo := &mockUserRepository{
		createFunc: func(ctx context.Context, user *model.User) error {
			if user.PasswordHash == "password123" {
				t.Fatal("expected password to be hashed before persistence")
			}

			if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")); err != nil {
				t.Fatalf("expected stored hash to match password: %v", err)
			}

			return nil
		},
	}

	svc := NewUserService(mockRepo)

	err := svc.Create(context.Background(), &model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "password123",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestUserServiceAuthenticateSuccess(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mockRepo := &mockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				Username:     "testuser",
				Email:        email,
				PasswordHash: string(hashedPassword),
			}, nil
		},
	}

	svc := NewUserService(mockRepo)

	user, err := svc.Authenticate(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if user == nil || user.Email != "test@example.com" {
		t.Fatalf("expected authenticated user, got %#v", user)
	}
}

func TestUserServiceAuthenticateInvalidCredentials(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mockRepo := &mockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				Username:     "testuser",
				Email:        email,
				PasswordHash: string(hashedPassword),
			}, nil
		},
	}

	svc := NewUserService(mockRepo)

	_, err = svc.Authenticate(context.Background(), "test@example.com", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserServiceAuthenticateUserNotFound(t *testing.T) {
	mockRepo := &mockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return nil, pgx.ErrNoRows
		},
	}

	svc := NewUserService(mockRepo)

	_, err := svc.Authenticate(context.Background(), "missing@example.com", "password123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserServiceGetProfileSuccess(t *testing.T) {
	expectedTime := time.Now().UTC()
	expectedID := "550e8400-e29b-41d4-a716-446655440000"

	mockRepo := &mockUserRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{
				ID:        expectedID,
				Username:  "profileuser",
				Email:     "profile@example.com",
				CreatedAt: expectedTime,
			}, nil
		},
	}

	svc := NewUserService(mockRepo)

	user, err := svc.GetByID(context.Background(), expectedID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if user == nil {
		t.Fatal("expected user profile, got nil")
	}

	if user.Username != "profileuser" {
		t.Errorf("expected username 'profileuser', got '%s'", user.Username)
	}

	if user.Email != "profile@example.com" {
		t.Errorf("expected email 'profile@example.com', got '%s'", user.Email)
	}

	if user.CreatedAt != expectedTime {
		t.Errorf("expected created_at %v, got %v", expectedTime, user.CreatedAt)
	}
}
