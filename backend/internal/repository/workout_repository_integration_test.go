//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
)

func TestWorkoutRepositoryCreateSessionPlannedWithRoutineIntegration(t *testing.T) {
	db := setupExerciseTestDB(t)
	cleanupExercisesRepository(t, db)

	userID := insertUserRaw(t, db, "plannedworkout", "plannedworkout@example.com")
	routineID := insertRoutineRawRepository(t, db, userID, "Push Day", "Rutina para planificar", time.Now())
	benchID := insertExerciseRawRepository(t, db, model.Exercise{
		Name:         "Bench Press",
		Description:  "Flat bench press",
		MuscleGroup:  "chest",
		ExerciseType: "strength",
		IsOfficial:   true,
	})
	flyID := insertExerciseRawRepository(t, db, model.Exercise{
		Name:         "Chest Fly",
		Description:  "Cable fly",
		MuscleGroup:  "chest",
		ExerciseType: "hypertrophy",
		IsOfficial:   true,
	})

	var primaryRoutineExerciseID string
	err := db.QueryRow(context.Background(), `
		INSERT INTO public.routine_exercises (routine_id, exercise_id, exercise_order, notes)
		VALUES ($1::uuid, $2::uuid, 1, 'principal')
		RETURNING id::text
	`, routineID, benchID).Scan(&primaryRoutineExerciseID)
	if err != nil {
		t.Fatalf("error insertando primer routine_exercise: %v", err)
	}

	var secondaryRoutineExerciseID string
	err = db.QueryRow(context.Background(), `
		INSERT INTO public.routine_exercises (routine_id, exercise_id, exercise_order, notes)
		VALUES ($1::uuid, $2::uuid, 2, 'accesorio')
		RETURNING id::text
	`, routineID, flyID).Scan(&secondaryRoutineExerciseID)
	if err != nil {
		t.Fatalf("error insertando segundo routine_exercise: %v", err)
	}

	if _, err := db.Exec(context.Background(), `
		INSERT INTO public.routine_exercise_sets (
			routine_exercise_id,
			set_number,
			target_reps_min,
			target_reps_max,
			target_weight_kg,
			target_rir,
			rest_seconds,
			notes
		)
		VALUES
			($1::uuid, 1, 6, 8, 75, 2, 120, 'heavy set'),
			($1::uuid, 2, 8, 10, 72.5, 2, 90, 'backoff set'),
			($2::uuid, 1, 10, 12, NULL, 2, 60, 'pump set')
	`, primaryRoutineExerciseID, secondaryRoutineExerciseID); err != nil {
		t.Fatalf("error insertando routine_exercise_sets: %v", err)
	}

	repo := NewWorkoutRepository(db)
	plannedAt := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
	parsedUserID := uuid.MustParse(userID)
	parsedRoutineID := uuid.MustParse(routineID)
	workout := &model.WorkoutSession{
		UserID:    parsedUserID,
		RoutineID: &parsedRoutineID,
		Name:      "Planned Push Day",
		PlannedAt: &plannedAt,
	}

	if err := repo.CreateSession(context.Background(), workout); err != nil {
		t.Fatalf("no se esperaba error en CreateSession, pero se obtuvo: %v", err)
	}

	if workout.ID == uuid.Nil {
		t.Fatal("se esperaba que CreateSession asignase un ID")
	}

	var (
		storedPlannedAt time.Time
		storedRoutineID uuid.UUID
	)
	err = db.QueryRow(context.Background(), `
		SELECT planned_at, routine_id
		FROM public.workout_sessions
		WHERE id = $1
	`, workout.ID).Scan(&storedPlannedAt, &storedRoutineID)
	if err != nil {
		t.Fatalf("error comprobando workout_session creada: %v", err)
	}

	if !storedPlannedAt.Equal(plannedAt) {
		t.Fatalf("se esperaba planned_at %s, pero se obtuvo %s", plannedAt, storedPlannedAt)
	}
	if storedRoutineID != parsedRoutineID {
		t.Fatalf("se esperaba routine_id %s, pero se obtuvo %s", parsedRoutineID, storedRoutineID)
	}

	var exerciseCount int
	err = db.QueryRow(context.Background(), `
		SELECT COUNT(*)::int
		FROM public.workout_exercises
		WHERE workout_session_id = $1
	`, workout.ID).Scan(&exerciseCount)
	if err != nil {
		t.Fatalf("error contando workout_exercises: %v", err)
	}

	if exerciseCount != 2 {
		t.Fatalf("se esperaban 2 workout_exercises copiados, pero se obtuvieron %d", exerciseCount)
	}

	var setCount int
	err = db.QueryRow(context.Background(), `
		SELECT COUNT(*)::int
		FROM public.workout_sets ws
		INNER JOIN public.workout_exercises we ON we.id = ws.workout_exercise_id
		WHERE we.workout_session_id = $1
	`, workout.ID).Scan(&setCount)
	if err != nil {
		t.Fatalf("error contando workout_sets: %v", err)
	}

	if setCount != 3 {
		t.Fatalf("se esperaban 3 workout_sets copiados, pero se obtuvieron %d", setCount)
	}
}
