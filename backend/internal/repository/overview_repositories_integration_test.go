//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func insertRoutineRawRepository(t *testing.T, db *pgxpool.Pool, userID, name, description string, updatedAt time.Time) string {
	t.Helper()

	var id string
	err := db.QueryRow(context.Background(), `
		INSERT INTO public.routines (user_id, name, description, updated_at)
		VALUES ($1::uuid, $2, $3, $4)
		RETURNING id::text
	`, userID, name, description, updatedAt).Scan(&id)
	if err != nil {
		t.Fatalf("error insertando rutina en public.routines: %v", err)
	}

	return id
}

func TestRoutineRepositoryListRecentByUserIntegration(t *testing.T) {
	db := setupExerciseTestDB(t)
	cleanupExercisesRepository(t, db)

	userID := insertUserRaw(t, db, "routineuser", "routineuser@example.com")
	exerciseID := insertExerciseRawRepository(t, db, model.Exercise{
		Name:         "Bench Press",
		Description:  "Flat bench press",
		MuscleGroup:  "chest",
		ExerciseType: "strength",
		IsOfficial:   true,
	})

	oldRoutineID := insertRoutineRawRepository(t, db, userID, "Rutina vieja", "Description antigua", time.Now().Add(-48*time.Hour))
	newRoutineID := insertRoutineRawRepository(t, db, userID, "Rutina nueva", "Description reciente", time.Now().Add(-2*time.Hour))

	if _, err := db.Exec(context.Background(), `
		INSERT INTO public.routine_exercises (routine_id, exercise_id, exercise_order, notes)
		VALUES ($1::uuid, $2::uuid, 1, 'serie principal'),
		       ($3::uuid, $2::uuid, 1, 'serie principal')
	`, oldRoutineID, exerciseID, newRoutineID); err != nil {
		t.Fatalf("error insertando routine_exercises: %v", err)
	}

	repo := NewRoutineRepository(db)

	routines, err := repo.ListRecentByUser(context.Background(), userID, 2)
	if err != nil {
		t.Fatalf("no se esperaba error en ListRecentByUser, pero se obtuvo: %v", err)
	}

	if len(routines) != 2 {
		t.Fatalf("se esperaban 2 rutinas, pero se obtuvieron %d", len(routines))
	}

	if routines[0].ID != newRoutineID {
		t.Fatalf("se esperaba la rutina mas reciente primero, pero se obtuvo %s", routines[0].ID)
	}

	if routines[0].ExerciseCount != 1 {
		t.Fatalf("se esperaba ExerciseCount=1, pero se obtuvo %d", routines[0].ExerciseCount)
	}
}

func TestOverviewWorkoutRepositoryIntegration(t *testing.T) {
	db := setupExerciseTestDB(t)
	cleanupExercisesRepository(t, db)

	userID := insertUserRaw(t, db, "workoutuser", "workoutuser@example.com")
	routineID := insertRoutineRawRepository(t, db, userID, "Push Pull Legs", "Rutina de prueba", time.Now())
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	chestExerciseID := insertExerciseRawRepository(t, db, model.Exercise{
		Name:         "Press banca",
		Description:  "Press de pecho",
		MuscleGroup:  "chest",
		ExerciseType: "strength",
		IsOfficial:   true,
	})
	backExerciseID := insertExerciseRawRepository(t, db, model.Exercise{
		Name:         "Remo con barra",
		Description:  "Remo",
		MuscleGroup:  "back",
		ExerciseType: "strength",
		IsOfficial:   true,
	})

	var recentSessionID string
	startedRecent := now.Add(-24 * time.Hour)
	err := db.QueryRow(context.Background(), `
		INSERT INTO public.workout_sessions (user_id, routine_id, name, performed_at, duration_minutes)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5)
		RETURNING id::text
	`, userID, routineID, "Sesion reciente", startedRecent, 70).Scan(&recentSessionID)
	if err != nil {
		t.Fatalf("error insertando workout_session reciente: %v", err)
	}

	var oldSessionID string
	startedOld := monthStart.AddDate(0, 0, -2)
	err = db.QueryRow(context.Background(), `
		INSERT INTO public.workout_sessions (user_id, routine_id, name, performed_at, duration_minutes)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5)
		RETURNING id::text
	`, userID, routineID, "Sesion antigua", startedOld, 60).Scan(&oldSessionID)
	if err != nil {
		t.Fatalf("error insertando workout_session antigua: %v", err)
	}

	if _, err := db.Exec(context.Background(), `
		INSERT INTO public.workout_exercises (workout_session_id, exercise_id, exercise_order, notes)
		VALUES ($1::uuid, $2::uuid, 1, 'principal'),
		       ($1::uuid, $3::uuid, 2, 'secundario'),
		       ($4::uuid, $3::uuid, 1, 'principal')
	`, recentSessionID, chestExerciseID, backExerciseID, oldSessionID); err != nil {
		t.Fatalf("error insertando workout_exercises: %v", err)
	}

	repo := NewOverviewWorkoutRepository(db)

	workouts, err := repo.ListRecentByUser(context.Background(), userID, 2)
	if err != nil {
		t.Fatalf("no se esperaba error en ListRecentByUser, pero se obtuvo: %v", err)
	}

	if len(workouts) != 2 {
		t.Fatalf("se esperaban 2 entrenos, pero se obtuvieron %d", len(workouts))
	}

	if workouts[0].ID != recentSessionID {
		t.Fatalf("se esperaba el entreno mas reciente primero, pero se obtuvo %s", workouts[0].ID)
	}

	if workouts[0].ExerciseCount != 2 {
		t.Fatalf("se esperaba ExerciseCount=2, pero se obtuvo %d", workouts[0].ExerciseCount)
	}

	monthEnd := monthStart.AddDate(0, 1, 0)

	dates, err := repo.ListTrainingDatesInRange(context.Background(), userID, monthStart, monthEnd)
	if err != nil {
		t.Fatalf("no se esperaba error en ListTrainingDatesInRange, pero se obtuvo: %v", err)
	}

	if len(dates) != 1 {
		t.Fatalf("se esperaba 1 fecha de entreno este mes, pero se obtuvieron %d", len(dates))
	}

	distributionStart := monthStart.AddDate(0, 0, -2)
	distributionEnd := now.AddDate(0, 0, 1)

	shares, total, err := repo.ListMuscleDistributionByUser(context.Background(), userID, distributionStart, distributionEnd)
	if err != nil {
		t.Fatalf("no se esperaba error en ListMuscleDistributionByUser, pero se obtuvo: %v", err)
	}

	if total != 3 {
		t.Fatalf("se esperaba total=3, pero se obtuvo %d", total)
	}

	if len(shares) != 2 {
		t.Fatalf("se esperaban 2 grupos musculares, pero se obtuvieron %d", len(shares))
	}
}

func TestBodyMetricRepositoryListRecentByUserIntegration(t *testing.T) {
	db := setupExerciseTestDB(t)
	cleanupExercisesRepository(t, db)

	userID := insertUserRaw(t, db, "metricsuser", "metricsuser@example.com")

	if _, err := db.Exec(context.Background(), `
		INSERT INTO public.body_metrics (user_id, recorded_at, weight_kg, body_fat_percentage, muscle_mass_kg)
		VALUES ($1::uuid, $2, 79.2, 18.5, 36.0),
		       ($1::uuid, $3, 78.5, 17.9, 36.4)
	`, userID, time.Now().Add(-10*24*time.Hour), time.Now().Add(-3*24*time.Hour)); err != nil {
		t.Fatalf("error insertando body_metrics: %v", err)
	}

	repo := NewBodyMetricRepository(db)

	entries, err := repo.ListRecentByUser(context.Background(), userID, 2)
	if err != nil {
		t.Fatalf("no se esperaba error en ListRecentByUser, pero se obtuvo: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("se esperaban 2 medidas, pero se obtuvieron %d", len(entries))
	}

	if entries[0].WeightKg == nil || *entries[0].WeightKg != 78.5 {
		t.Fatalf("se esperaba peso reciente 78.5, pero se obtuvo %#v", entries[0].WeightKg)
	}

	if entries[1].WeightKg == nil || *entries[1].WeightKg != 79.2 {
		t.Fatalf("se esperaba peso anterior 79.2, pero se obtuvo %#v", entries[1].WeightKg)
	}
}
