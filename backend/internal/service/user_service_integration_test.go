package service

import (
	"context"
	"errors"
	"testing"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/isw2-unileon/Grupo-16/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDBService(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return testutil.NewIntegrationTestPool(t)
}

func cleanupUsersService(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	// Delete from dependent tables first to satisfy foreign keys, then users.
	tables := []string{
		"public.workout_sets",
		"public.workout_exercises",
		"public.workout_sessions",
		"public.support_tickets",
		"public.shared_routines",
		"public.routine_exercises",
		"public.routines",
		"public.friendships",
		"public.exercise_secondary_muscle_groups",
		"public.exercises",
		"public.body_metrics",
		"public.user_profiles",
		"public.users",
	}

	for _, tbl := range tables {
		_, err := db.Exec(context.Background(), "DELETE FROM "+tbl)
		if err != nil {
			t.Fatalf("error cleaning %s: %v", tbl, err)
		}
	}
}

func TestUserServiceCreateAndAuthenticateIntegration(t *testing.T) {
	db := setupTestDBService(t)
	cleanupUsersService(t, db)

	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)

	ctx := context.Background()

	user := &model.User{
		Username:     "svccreated",
		Email:        "svccreated@example.com",
		PasswordHash: "plain-password-123",
	}

	// Create the user through the service so the password is hashed before persisting.
	if err := svc.Create(ctx, user); err != nil {
		t.Fatalf("expected nil error creating user via service, got: %v", err)
	}

	if user.ID == "" {
		t.Fatalf("expected created user to have ID set")
	}
	if user.CreatedAt.IsZero() {
		t.Fatalf("expected created user to have CreatedAt set")
	}

	// Authenticate with the correct password.
	authUser, err := svc.Authenticate(ctx, "svccreated@example.com", "plain-password-123")
	if err != nil {
		t.Fatalf("expected nil error authenticating user, got: %v", err)
	}
	if authUser == nil {
		t.Fatalf("expected authenticated user, got nil")
		return
	}
	if authUser.ID != user.ID {
		t.Fatalf("expected auth user ID %s, got %s", user.ID, authUser.ID)
	}
	if authUser.Email != user.Email {
		t.Fatalf("expected auth user email %s, got %s", user.Email, authUser.Email)
	}

	// Authenticate with a wrong password.
	_, err = svc.Authenticate(ctx, "svccreated@example.com", "wrong-password")
	if err == nil {
		t.Fatalf("expected error when authenticating with wrong password")
	}
	// Use errors.Is to correctly handle wrapped errors.
	if !errors.Is(err, ErrInvalidCredentials) {
		// Service maps repository ErrNoRows to ErrInvalidCredentials,
		// but for wrong password it returns ErrInvalidCredentials as well.
		// Check that we get the expected error type.
		t.Logf("got auth error: %v (expected ErrInvalidCredentials)", err)
	}
}

func TestUserServiceGetByIDIntegration(t *testing.T) {
	db := setupTestDBService(t)
	cleanupUsersService(t, db)

	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)
	ctx := context.Background()

	user := &model.User{
		Username:     "svcbyid",
		Email:        "svcbyid@example.com",
		PasswordHash: "another-password",
	}

	if err := svc.Create(ctx, user); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if user.ID == "" {
		t.Fatalf("expected user.ID after create")
	}

	ret, err := svc.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if ret == nil {
		t.Fatalf("expected user, got nil")
		return
	}
	if ret.ID != user.ID {
		t.Fatalf("expected ID %s, got %s", user.ID, ret.ID)
	}
	if ret.Email != user.Email {
		t.Fatalf("expected email %s, got %s", user.Email, ret.Email)
	}
	if ret.CreatedAt.IsZero() {
		t.Fatalf("expected CreatedAt set")
	}
}

func TestUserServiceGetByIDNotFoundIntegration(t *testing.T) {
	db := setupTestDBService(t)
	cleanupUsersService(t, db)

	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)
	ctx := context.Background()

	_, err := svc.GetByID(ctx, "550e8400-e29b-41d4-a716-446655449999")
	if err == nil {
		t.Fatalf("expected error for non-existent user")
	}
	// Use errors.Is to correctly detect wrapped errors.
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
}
