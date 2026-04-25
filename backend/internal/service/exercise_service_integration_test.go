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

func setupExerciseServiceTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return testutil.NewIntegrationTestPool(t)
}

func cleanupExercisesService(t *testing.T, db *pgxpool.Pool) {
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
		if _, err := db.Exec(context.Background(), "DELETE FROM "+table); err != nil {
			t.Fatalf("error limpiando %s: %v", table, err)
		}
	}
}

func insertExerciseRawService(t *testing.T, db *pgxpool.Pool, exercise model.Exercise) string {
	t.Helper()

	var id string
	err := db.QueryRow(context.Background(), `
		INSERT INTO public.exercises (name, description, muscle_group, exercise_type, is_official)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text
	`, exercise.Name, exercise.Description, exercise.MuscleGroup, exercise.ExerciseType, exercise.IsOfficial).Scan(&id)
	if err != nil {
		t.Fatalf("error insertando ejercicio: %v", err)
	}

	if exercise.SecondaryMuscleGroup != "" {
		_, err = db.Exec(context.Background(), `
			INSERT INTO public.exercise_secondary_muscle_groups (exercise_id, muscle_group)
			VALUES ($1::uuid, $2)
		`, id, exercise.SecondaryMuscleGroup)
		if err != nil {
			t.Fatalf("error insertando secondary muscle group: %v", err)
		}
	}

	return id
}

func TestExerciseServiceGetByIDIntegration(t *testing.T) {
	db := setupExerciseServiceTestDB(t)
	cleanupExercisesService(t, db)

	insertedID := insertExerciseRawService(t, db, model.Exercise{
		Name:                 "Pull Up",
		Description:          "Bodyweight pull up",
		MuscleGroup:          "back",
		SecondaryMuscleGroup: "biceps",
		ExerciseType:         "bodyweight",
		IsOfficial:           true,
	})

	repo := repository.NewExerciseRepository(db)
	svc := NewExerciseService(repo)

	exercise, err := svc.GetByID(context.Background(), insertedID)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if exercise == nil {
		t.Fatal("expected exercise, got nil")
	}
	if exercise.ID != insertedID {
		t.Fatalf("expected ID %s, got %s", insertedID, exercise.ID)
	}
	if exercise.Name != "Pull Up" {
		t.Fatalf("expected name Pull Up, got %s", exercise.Name)
	}
}

func TestExerciseServiceGetByIDNotFoundIntegration(t *testing.T) {
	db := setupExerciseServiceTestDB(t)
	cleanupExercisesService(t, db)

	repo := repository.NewExerciseRepository(db)
	svc := NewExerciseService(repo)

	_, err := svc.GetByID(context.Background(), "550e8400-e29b-41d4-a716-446655449999")
	if err == nil {
		t.Fatal("expected error for non-existent exercise")
	}
	if !errors.Is(err, ErrExerciseNotFound) {
		t.Fatalf("expected ErrExerciseNotFound, got: %v", err)
	}
}

func TestExerciseServiceListIntegration(t *testing.T) {
	db := setupExerciseServiceTestDB(t)
	cleanupExercisesService(t, db)

	insertExerciseRawService(t, db, model.Exercise{
		Name:         "Bench Press",
		Description:  "Flat bench press",
		MuscleGroup:  "chest",
		ExerciseType: "strength",
		IsOfficial:   true,
	})
	insertExerciseRawService(t, db, model.Exercise{
		Name:         "Row",
		Description:  "Barbell row",
		MuscleGroup:  "back",
		ExerciseType: "strength",
		IsOfficial:   true,
	})

	repo := repository.NewExerciseRepository(db)
	svc := NewExerciseService(repo)

	result, err := svc.List(context.Background(), model.ExerciseFilter{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 exercises, got %d", len(result.Items))
	}
}

func TestExerciseServiceDeleteExerciseIntegration(t *testing.T) {
	db := setupExerciseServiceTestDB(t)
	cleanupExercisesService(t, db)

	insertedID := insertExerciseRawService(t, db, model.Exercise{
		Name:         "Cable Row",
		Description:  "Seated cable row",
		MuscleGroup:  "back",
		ExerciseType: "machine",
		IsOfficial:   true,
	})

	repo := repository.NewExerciseRepository(db)
	svc := NewExerciseService(repo)

	if err := svc.DeleteExercise(context.Background(), insertedID); err != nil {
		t.Fatalf("expected nil error deleting exercise, got: %v", err)
	}

	_, err := svc.GetByID(context.Background(), insertedID)
	if !errors.Is(err, ErrExerciseNotFound) {
		t.Fatalf("expected ErrExerciseNotFound after soft-delete, got: %v", err)
	}
}

func TestExerciseServiceDeleteExerciseNotFoundIntegration(t *testing.T) {
	db := setupExerciseServiceTestDB(t)
	cleanupExercisesService(t, db)

	repo := repository.NewExerciseRepository(db)
	svc := NewExerciseService(repo)

	err := svc.DeleteExercise(context.Background(), "550e8400-e29b-41d4-a716-446655449999")
	if !errors.Is(err, ErrExerciseNotFound) {
		t.Fatalf("expected ErrExerciseNotFound, got: %v", err)
	}
}
