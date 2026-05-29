package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	createFunc     func(ctx context.Context, user *model.User) error
	getByIDFunc    func(ctx context.Context, id string) (*model.User, error)
	getByEmailFunc func(ctx context.Context, email string) (*model.User, error)
	listAllFunc    func(ctx context.Context) ([]*model.User, error)
	deleteFunc     func(ctx context.Context, id string) error
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

func (m *mockUserRepository) ListAll(ctx context.Context) ([]*model.User, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
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
		Role:         "user",
		IsActive:     true,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestUserServiceCreateReturnsUserAlreadyExistsOnUniqueViolation(t *testing.T) {
	mockRepo := &mockUserRepository{
		createFunc: func(ctx context.Context, user *model.User) error {
			return &pgconn.PgError{Code: "23505"}
		},
	}

	svc := NewUserService(mockRepo)

	err := svc.Create(context.Background(), &model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "password123",
	})
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
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
				Role:         "user",
				IsActive:     true,
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
				Role:         "user",
				IsActive:     true,
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
				Role:      "user",
				IsActive:  true,
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
		return
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

func TestUserServiceListAllSuccess(t *testing.T) {
	// Arrange
	mockRepo := &mockUserRepository{
		listAllFunc: func(ctx context.Context) ([]*model.User, error) {
			return []*model.User{
				{ID: "1", Username: "admin_user", Email: "admin@example.com", Role: "admin"},
				{ID: "2", Username: "normal_user", Email: "user@example.com", Role: "user"},
			}, nil
		},
	}

	svc := NewUserService(mockRepo)

	// Act
	users, err := svc.ListAll(context.Background())
	// Assert
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if users == nil {
		t.Fatal("expected user list, got nil")
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if users[0].Username != "admin_user" || users[1].Username != "normal_user" {
		t.Errorf("the returned users do not match the expected data")
	}
}

func TestUserServiceDeleteSuccess(t *testing.T) {
	mockRepo := &mockUserRepository{
		deleteFunc: func(ctx context.Context, id string) error {
			if id != "1" {
				t.Errorf("expected to delete user ID '1', got '%s'", id)
			}
			return nil
		},
	}

	svc := NewUserService(mockRepo)

	err := svc.Delete(context.Background(), "1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestUserServiceDeleteNotFound(t *testing.T) {
	mockRepo := &mockUserRepository{
		deleteFunc: func(ctx context.Context, id string) error {
			return pgx.ErrNoRows
		},
	}

	svc := NewUserService(mockRepo)

	err := svc.Delete(context.Background(), "99")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
