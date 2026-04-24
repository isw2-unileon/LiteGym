package repository

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testDBURL = "postgres://test_user:test_password@localhost:5432/test_db"

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if env := os.Getenv("TEST_DB_URL"); env != "" {
		testDBURL = env
	}

	db, err := pgxpool.New(context.Background(), testDBURL)
	if err != nil {
		t.Fatalf("error conectando a la base de test: %v", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		t.Fatalf("error haciendo ping a la base de test: %v", err)
	}

	return db
}

func cleanupUsers(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

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

	for _, table := range tables {
		_, err := db.Exec(context.Background(), "DELETE FROM "+table)
		if err != nil {
			t.Fatalf("error limpiando %s: %v", table, err)
		}
	}
}

func insertUserRaw(t *testing.T, db *pgxpool.Pool, username, email string) int {
	t.Helper()

	var id int64
	err := db.QueryRow(context.Background(), `
		INSERT INTO public.users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, username, email, "").Scan(&id)
	if err != nil {
		t.Fatalf("error insertando en public.users: %v", err)
	}

	return int(id)
}

func TestUserRepositoryCreateIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanupUsers(t, db)

	repo := NewUserRepository(db)

	user := &model.User{
		Username: "createduser",
		Email:    "created@example.com",
	}

	err := repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("no se esperaba error en Create, pero se obtuvo: %v", err)
	}

	if user.ID == 0 {
		t.Fatal("se esperaba que el usuario tuviera ID tras el Create")
	}

	if user.CreatedAt.IsZero() {
		t.Fatal("se esperaba que el usuario tuviera CreatedAt tras el Create")
	}

	var (
		dbID       int64
		dbUsername string
		dbEmail    string
	)

	err = db.QueryRow(context.Background(), `
		SELECT id, username, email
		FROM public.users
		WHERE email = $1
	`, "created@example.com").Scan(&dbID, &dbUsername, &dbEmail)
	if err != nil {
		t.Fatalf("error comprobando usuario creado en la base: %v", err)
	}

	if user.ID != int(dbID) {
		t.Fatalf("ID incorrecto: esperado %d, obtenido %d", dbID, user.ID)
	}

	if dbUsername != "createduser" {
		t.Fatalf("username incorrecto: esperado createduser, obtenido %s", dbUsername)
	}

	if dbEmail != "created@example.com" {
		t.Fatalf("email incorrecto: esperado created@example.com, obtenido %s", dbEmail)
	}
}

func TestUserRepositoryGetByIDIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanupUsers(t, db)

	insertedID := insertUserRaw(t, db, "userbyid", "userbyid@example.com")

	repo := NewUserRepository(db)

	user, err := repo.GetByID(context.Background(), insertedID)
	if err != nil {
		t.Fatalf("no se esperaba error en GetByID, pero se obtuvo: %v", err)
		return
	}

	if user == nil {
		t.Fatal("se esperaba un usuario, pero se obtuvo nil")
		return
	}

	if user.ID != insertedID {
		t.Fatalf("id incorrecto: esperado %d, obtenido %d", insertedID, user.ID)
	}

	if user.Username != "userbyid" {
		t.Fatalf("username incorrecto: esperado userbyid, obtenido %s", user.Username)
	}

	if user.Email != "userbyid@example.com" {
		t.Fatalf("email incorrecto: esperado userbyid@example.com, obtenido %s", user.Email)
	}

	if user.CreatedAt.IsZero() {
		t.Fatal("se esperaba CreatedAt informado")
	}
}

func TestUserRepositoryGetByIDNotFoundIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanupUsers(t, db)

	repo := NewUserRepository(db)

	user, err := repo.GetByID(context.Background(), 999999)
	if err == nil {
		t.Fatal("se esperaba error al buscar un usuario inexistente")
	}

	if user != nil {
		t.Fatalf("se esperaba usuario nil, pero se obtuvo: %#v", user)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("se esperaba pgx.ErrNoRows, pero se obtuvo: %v", err)
	}
}

func TestUserRepositoryGetByEmailIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanupUsers(t, db)

	insertUserRaw(t, db, "userbyemail", "userbyemail@example.com")

	repo := NewUserRepository(db)

	user, err := repo.GetByEmail(context.Background(), "userbyemail@example.com")
	if err != nil {
		t.Fatalf("no se esperaba error en GetByEmail, pero se obtuvo: %v", err)
		return
	}

	if user == nil {
		t.Fatal("se esperaba un usuario, pero se obtuvo nil")
		return
	}

	if user.Username != "userbyemail" {
		t.Fatalf("username incorrecto: esperado userbyemail, obtenido %s", user.Username)
	}

	if user.Email != "userbyemail@example.com" {
		t.Fatalf("email incorrecto: esperado userbyemail@example.com, obtenido %s", user.Email)
	}

	if user.ID == 0 {
		t.Fatal("se esperaba ID informado")
	}

	if user.CreatedAt.IsZero() {
		t.Fatal("se esperaba CreatedAt informado")
	}
}

func TestUserRepositoryGetByEmailNotFoundIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanupUsers(t, db)

	repo := NewUserRepository(db)

	user, err := repo.GetByEmail(context.Background(), "notfound@example.com")
	if err == nil {
		t.Fatal("se esperaba error al buscar un email inexistente")
	}

	if user != nil {
		t.Fatalf("se esperaba usuario nil, pero se obtuvo: %#v", user)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("se esperaba pgx.ErrNoRows, pero se obtuvo: %v", err)
	}
}

func TestUserRepositoryListAllIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanupUsers(t, db)

	insertUserRaw(t, db, "user1", "user1@example.com")
	insertUserRaw(t, db, "user2", "user2@example.com")

	repo := NewUserRepository(db)

	users, err := repo.ListAll(context.Background())
	if err != nil {
		t.Fatalf("no se esperaba error en ListAll, pero se obtuvo: %v", err)
	}

	if users == nil {
		t.Fatal("se esperaba una lista de usuarios, pero se obtuvo nil")
	}

	if len(users) != 2 {
		t.Fatalf("se esperaban 2 usuarios, se obtuvieron %d", len(users))
	}
}

func TestUserRepositoryDeleteIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanupUsers(t, db)

	insertedID := insertUserRaw(t, db, "todelete", "delete@example.com")

	repo := NewUserRepository(db)

	err := repo.Delete(context.Background(), insertedID)
	if err != nil {
		t.Fatalf("no se esperaba error al borrar, se obtuvo: %v", err)
	}

	user, err := repo.GetByID(context.Background(), insertedID)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("se esperaba pgx.ErrNoRows al buscar el usuario borrado, se obtuvo: %v", err)
	}
	if user != nil {
		t.Fatalf("se esperaba que el usuario fuera nil tras borrarlo")
	}
}

func TestUserRepositoryDeleteNotFoundIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanupUsers(t, db)

	repo := NewUserRepository(db)

	err := repo.Delete(context.Background(), 999999)

	if err == nil {
		t.Fatal("se esperaba error al borrar un usuario inexistente")
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("se esperaba pgx.ErrNoRows, pero se obtuvo: %v", err)
	}
}
