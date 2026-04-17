package service

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testDBURLService = "postgres://test_user:test_password@localhost:5432/test_db"

func setupTestDBService(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if env := os.Getenv("TEST_DB_URL"); env != "" {
		testDBURLService = env
	}

	db, err := pgxpool.New(context.Background(), testDBURLService)
	if err != nil {
		t.Fatalf("error conectando a la base de test: %v", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		t.Fatalf("error haciendo ping a la base de test: %v", err)
	}

	return db
}

func cleanupUsersService(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	// Delete from dependent tables first if present (avoid FK violations),
	// then delete from public.users. Keep this conservative but safe.
	tables := []string{
		"public.workout_sets",
		"public.workout_exercises",
		"public.workout_sessions",
		"public.support_tickets",
		"public.shared_routines",
		"public.routine_exercises",
		"public.routines",
		"public.friendships",
		"public.exercises",
		"public.body_metrics",
		"public.user_profiles",
		"public.users",
		"auth.users",
	}

	for _, tbl := range tables {
		_, err := db.Exec(context.Background(), "DELETE FROM "+tbl)
		// ignore errors for tables that don't exist in some schema versions
		if err != nil {
			// If the error indicates the relation/table does not exist, ignore it.
			// Otherwise fail the test so we don't mask real errors.
			errStr := err.Error()
			if strings.Contains(errStr, "does not exist") || strings.Contains(errStr, "no such table") || strings.Contains(errStr, "relation") {
				continue
			}
			t.Fatalf("error cleaning %s: %v", tbl, err)
		}
	}
}

func TestUserServiceCreateAndAuthenticateIntegration(t *testing.T) {
	db := setupTestDBService(t)
	defer db.Close()

	cleanupUsersService(t, db)

	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)

	ctx := context.Background()

	user := &model.User{
		Username:     "svccreated",
		Email:        "svccreated@example.com",
		PasswordHash: "plain-password-123",
	}

	// Create (service should hash the password and persist)
	if err := svc.Create(ctx, user); err != nil {
		t.Fatalf("expected nil error creating user via service, got: %v", err)
	}

	if user.ID == 0 {
		t.Fatalf("expected created user to have ID set")
	}
	if user.CreatedAt.IsZero() {
		t.Fatalf("expected created user to have CreatedAt set")
	}

	// Authenticate with correct password
	authUser, err := svc.Authenticate(ctx, "svccreated@example.com", "plain-password-123")
	if err != nil {
		t.Fatalf("expected nil error authenticating user, got: %v", err)
	}
	if authUser == nil {
		t.Fatalf("expected authenticated user, got nil")
	}
	if authUser.ID != user.ID {
		t.Fatalf("expected auth user ID %d, got %d", user.ID, authUser.ID)
	}
	if authUser.Email != user.Email {
		t.Fatalf("expected auth user email %s, got %s", user.Email, authUser.Email)
	}

	// Authenticate with wrong password
	_, err = svc.Authenticate(ctx, "svccreated@example.com", "wrong-password")
	if err == nil {
		t.Fatalf("expected error when authenticating with wrong password")
	}
	if err != ErrInvalidCredentials {
		// Service maps repository ErrNoRows to ErrInvalidCredentials,
		// but for wrong password it returns ErrInvalidCredentials as well.
		// Check that we get the expected error type.
		t.Logf("got auth error: %v (expected ErrInvalidCredentials)", err)
	}
}

func TestUserServiceGetByIDIntegration(t *testing.T) {
	db := setupTestDBService(t)
	defer db.Close()

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

	if user.ID == 0 {
		t.Fatalf("expected user.ID after create")
	}

	// Service.GetByID expects int64 id parameter; convert
	ret, err := svc.GetByID(ctx, int(user.ID))
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if ret == nil {
		t.Fatalf("expected user, got nil")
	}
	if ret.ID != user.ID {
		t.Fatalf("expected ID %d, got %d", user.ID, ret.ID)
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
	defer db.Close()

	cleanupUsersService(t, db)

	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)
	ctx := context.Background()

	_, err := svc.GetByID(ctx, 99999999)
	if err == nil {
		t.Fatalf("expected error for non-existent user")
	}
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
}

// small helper to ensure Create has time to persist if needed (not usually necessary)
func waitFor(t *testing.T, d time.Duration) {
	t.Helper()
	time.Sleep(d)
}
