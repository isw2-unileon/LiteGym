//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupExerciseTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return testutil.NewIntegrationTestPool(t)
}

func cleanupExercisesRepository(t *testing.T, db *pgxpool.Pool) {
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

func insertExerciseRawRepository(t *testing.T, db *pgxpool.Pool, exercise model.Exercise) string {
	t.Helper()

	var id string
	err := db.QueryRow(context.Background(), `
		INSERT INTO public.exercises (name, description, muscle_group, exercise_type, is_official)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text
	`, exercise.Name, exercise.Description, exercise.MuscleGroup, exercise.ExerciseType, exercise.IsOfficial).Scan(&id)
	if err != nil {
		t.Fatalf("error inserting into public.exercises: %v", err)
	}

	if exercise.SecondaryMuscleGroup != "" {
		_, err = db.Exec(context.Background(), `
			INSERT INTO public.exercise_secondary_muscle_groups (exercise_id, muscle_group)
			VALUES ($1::uuid, $2)
		`, id, exercise.SecondaryMuscleGroup)
		if err != nil {
			t.Fatalf("error inserting secondary muscle group: %v", err)
		}
	}

	return id
}

func TestExerciseRepositoryCreateIntegration(t *testing.T) {
	db := setupExerciseTestDB(t)
	cleanupExercisesRepository(t, db)

	repo := NewExerciseRepository(db)

	exercise := &model.Exercise{
		Name:                 "Bench Press",
		Description:          "Flat bench press",
		MuscleGroup:          "chest",
		SecondaryMuscleGroup: "triceps",
		ExerciseType:         "strength",
		IsOfficial:           true,
	}

	if err := repo.Create(context.Background(), exercise); err != nil {
		t.Fatalf("expected no error from Create, got: %v", err)
	}

	if exercise.ID == "" {
		t.Fatal("expected exercise to have ID after Create")
	}

	if exercise.CreatedAt.IsZero() {
		t.Fatal("expected exercise to have CreatedAt after Create")
	}

	var (
		dbName          string
		dbDescription   string
		dbMuscleGroup   string
		dbExerciseType  string
		dbIsOfficial    bool
		dbSecondaryMusl string
	)

	err := db.QueryRow(context.Background(), `
		SELECT
			e.name,
			e.description,
			e.muscle_group,
			e.exercise_type,
			e.is_official,
			COALESCE(string_agg(esmg.muscle_group, ', ' ORDER BY esmg.muscle_group), '')
		FROM public.exercises e
		LEFT JOIN public.exercise_secondary_muscle_groups esmg ON esmg.exercise_id = e.id
		WHERE e.id = $1::uuid
		GROUP BY e.id, e.name, e.description, e.muscle_group, e.exercise_type, e.is_official
	`, exercise.ID).Scan(
		&dbName,
		&dbDescription,
		&dbMuscleGroup,
		&dbExerciseType,
		&dbIsOfficial,
		&dbSecondaryMusl,
	)
	if err != nil {
		t.Fatalf("error checking created exercise in database: %v", err)
	}

	if dbName != exercise.Name {
		t.Fatalf("incorrect name: expected %s, got %s", exercise.Name, dbName)
	}
	if dbDescription != exercise.Description {
		t.Fatalf("incorrect description: expected %s, got %s", exercise.Description, dbDescription)
	}
	if dbMuscleGroup != exercise.MuscleGroup {
		t.Fatalf("incorrect muscle_group: expected %s, got %s", exercise.MuscleGroup, dbMuscleGroup)
	}
	if dbExerciseType != exercise.ExerciseType {
		t.Fatalf("incorrect exercise_type: expected %s, got %s", exercise.ExerciseType, dbExerciseType)
	}
	if dbIsOfficial != exercise.IsOfficial {
		t.Fatalf("incorrect is_official: expected %t, got %t", exercise.IsOfficial, dbIsOfficial)
	}
	if dbSecondaryMusl != exercise.SecondaryMuscleGroup {
		t.Fatalf("incorrect secondary_muscle_group: expected %s, got %s", exercise.SecondaryMuscleGroup, dbSecondaryMusl)
	}
}

func TestExerciseRepositoryGetByIDIntegration(t *testing.T) {
	db := setupExerciseTestDB(t)
	cleanupExercisesRepository(t, db)

	insertedID := insertExerciseRawRepository(t, db, model.Exercise{
		Name:                 "Squat",
		Description:          "Back squat",
		MuscleGroup:          "legs",
		SecondaryMuscleGroup: "glutes",
		ExerciseType:         "strength",
		IsOfficial:           true,
	})

	repo := NewExerciseRepository(db)

	exercise, err := repo.GetByID(context.Background(), insertedID)
	if err != nil {
		t.Fatalf("expected no error from GetByID, got: %v", err)
	}

	if exercise == nil {
		t.Fatal("expected exercise, got nil")
		return
	}

	if exercise.ID != insertedID {
		t.Fatalf("incorrect id: expected %s, got %s", insertedID, exercise.ID)
	}
	if exercise.Name != "Squat" {
		t.Fatalf("incorrect name: expected Squat, got %s", exercise.Name)
	}
	if exercise.SecondaryMuscleGroup != "glutes" {
		t.Fatalf("incorrect secondary muscle group: expected glutes, got %s", exercise.SecondaryMuscleGroup)
	}
	if exercise.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestExerciseRepositoryGetByIDNotFoundIntegration(t *testing.T) {
	db := setupExerciseTestDB(t)
	cleanupExercisesRepository(t, db)

	repo := NewExerciseRepository(db)

	exercise, err := repo.GetByID(context.Background(), "550e8400-e29b-41d4-a716-446655449999")
	if err == nil {
		t.Fatal("expected error when finding a missing exercise")
	}

	if exercise != nil {
		t.Fatalf("expected nil exercise, got: %#v", exercise)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows, got: %v", err)
	}
}

func TestExerciseRepositoryListIntegration(t *testing.T) {
	db := setupExerciseTestDB(t)
	cleanupExercisesRepository(t, db)

	insertExerciseRawRepository(t, db, model.Exercise{
		Name:                 "Bench Press",
		Description:          "Flat bench press",
		MuscleGroup:          "chest",
		SecondaryMuscleGroup: "triceps",
		ExerciseType:         "strength",
		IsOfficial:           true,
	})
	insertExerciseRawRepository(t, db, model.Exercise{
		Name:         "Deadlift",
		Description:  "Conventional deadlift",
		MuscleGroup:  "back",
		ExerciseType: "strength",
		IsOfficial:   true,
	})

	repo := NewExerciseRepository(db)

	exercises, total, err := repo.List(context.Background(), model.ExerciseFilter{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("no se esperaba error en List, pero se obtuvo: %v", err)
	}

	if len(exercises) != 2 {
		t.Fatalf("se esperaban 2 ejercicios, pero se obtuvieron %d", len(exercises))
	}

	if total != 2 {
		t.Fatalf("se esperaba total 2, pero se obtuvo %d", total)
	}

	if exercises[0].ID == "" || exercises[1].ID == "" {
		t.Fatal("se esperaban IDs informados en la lista")
	}
}

func TestExerciseRepositoryDeleteExerciseSoftDeleteIntegration(t *testing.T) {
	db := setupExerciseTestDB(t)
	cleanupExercisesRepository(t, db)

	insertedID := insertExerciseRawRepository(t, db, model.Exercise{
		Name:         "Incline Press",
		Description:  "Incline bench press",
		MuscleGroup:  "chest",
		ExerciseType: "strength",
		IsOfficial:   true,
	})

	repo := NewExerciseRepository(db)

	if err := repo.DeleteExercise(context.Background(), insertedID); err != nil {
		t.Fatalf("no se esperaba error en DeleteExercise, pero se obtuvo: %v", err)
	}

	var deletedAtValid bool
	err := db.QueryRow(context.Background(), `
		SELECT deleted_at IS NOT NULL
		FROM public.exercises
		WHERE id = $1::uuid
	`, insertedID).Scan(&deletedAtValid)
	if err != nil {
		t.Fatalf("error comprobando deleted_at: %v", err)
	}
	if !deletedAtValid {
		t.Fatal("se esperaba deleted_at informado tras soft-delete")
	}

	_, err = repo.GetByID(context.Background(), insertedID)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("se esperaba pgx.ErrNoRows en GetByID tras soft-delete, se obtuvo: %v", err)
	}

	exercises, _, err := repo.List(context.Background(), model.ExerciseFilter{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("no se esperaba error en List, pero se obtuvo: %v", err)
	}
	if len(exercises) != 0 {
		t.Fatalf("se esperaban 0 ejercicios visibles tras soft-delete, se obtuvieron %d", len(exercises))
	}
}

func TestExerciseRepositoryDeleteExerciseNotFoundIntegration(t *testing.T) {
	db := setupExerciseTestDB(t)
	cleanupExercisesRepository(t, db)

	repo := NewExerciseRepository(db)
	err := repo.DeleteExercise(context.Background(), "550e8400-e29b-41d4-a716-446655449999")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("se esperaba pgx.ErrNoRows, se obtuvo: %v", err)
	}
}
