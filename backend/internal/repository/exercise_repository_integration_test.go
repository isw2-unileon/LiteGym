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
		t.Fatalf("error insertando en public.exercises: %v", err)
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
		t.Fatalf("no se esperaba error en Create, pero se obtuvo: %v", err)
	}

	if exercise.ID == "" {
		t.Fatal("se esperaba que el ejercicio tuviera ID tras el Create")
	}

	if exercise.CreatedAt.IsZero() {
		t.Fatal("se esperaba que el ejercicio tuviera CreatedAt tras el Create")
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
		t.Fatalf("error comprobando ejercicio creado en la base: %v", err)
	}

	if dbName != exercise.Name {
		t.Fatalf("name incorrecto: esperado %s, obtenido %s", exercise.Name, dbName)
	}
	if dbDescription != exercise.Description {
		t.Fatalf("description incorrecta: esperada %s, obtenida %s", exercise.Description, dbDescription)
	}
	if dbMuscleGroup != exercise.MuscleGroup {
		t.Fatalf("muscle_group incorrecto: esperado %s, obtenido %s", exercise.MuscleGroup, dbMuscleGroup)
	}
	if dbExerciseType != exercise.ExerciseType {
		t.Fatalf("exercise_type incorrecto: esperado %s, obtenido %s", exercise.ExerciseType, dbExerciseType)
	}
	if dbIsOfficial != exercise.IsOfficial {
		t.Fatalf("is_official incorrecto: esperado %t, obtenido %t", exercise.IsOfficial, dbIsOfficial)
	}
	if dbSecondaryMusl != exercise.SecondaryMuscleGroup {
		t.Fatalf("secondary_muscle_group incorrecto: esperado %s, obtenido %s", exercise.SecondaryMuscleGroup, dbSecondaryMusl)
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
		t.Fatalf("no se esperaba error en GetByID, pero se obtuvo: %v", err)
	}

	if exercise == nil {
		t.Fatal("se esperaba un ejercicio, pero se obtuvo nil")
	}

	if exercise.ID != insertedID {
		t.Fatalf("id incorrecto: esperado %s, obtenido %s", insertedID, exercise.ID)
	}
	if exercise.Name != "Squat" {
		t.Fatalf("name incorrecto: esperado Squat, obtenido %s", exercise.Name)
	}
	if exercise.SecondaryMuscleGroup != "glutes" {
		t.Fatalf("secondary muscle group incorrecto: esperado glutes, obtenido %s", exercise.SecondaryMuscleGroup)
	}
	if exercise.CreatedAt.IsZero() {
		t.Fatal("se esperaba CreatedAt informado")
	}
}

func TestExerciseRepositoryGetByIDNotFoundIntegration(t *testing.T) {
	db := setupExerciseTestDB(t)
	cleanupExercisesRepository(t, db)

	repo := NewExerciseRepository(db)

	exercise, err := repo.GetByID(context.Background(), "550e8400-e29b-41d4-a716-446655449999")
	if err == nil {
		t.Fatal("se esperaba error al buscar un ejercicio inexistente")
	}

	if exercise != nil {
		t.Fatalf("se esperaba ejercicio nil, pero se obtuvo: %#v", exercise)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("se esperaba pgx.ErrNoRows, pero se obtuvo: %v", err)
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
