package repository

import (
	"context"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ManualRoutineRepository defines the persistence operations for manually managed routines.
type ManualRoutineRepository interface {
	CountByUser(ctx context.Context, userID string) (int, error)
	CreateManualRoutine(ctx context.Context, routine model.ManualRoutineToSave) (string, error)
	UpdateManualRoutine(ctx context.Context, routineID, userID string, routine model.ManualRoutineToSave) error
	DeleteRoutine(ctx context.Context, routineID, userID string) error
	DuplicateRoutine(ctx context.Context, routineID, userID string) (string, error)
}

// NewManualRoutineRepository creates a manual routine repository backed by PostgreSQL.
func NewManualRoutineRepository(db *pgxpool.Pool) ManualRoutineRepository {
	return &routineRepository{db: db}
}

func (r *routineRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(1)::int
		FROM public.routines
		WHERE user_id = $1::uuid
	`, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *routineRepository) CreateManualRoutine(ctx context.Context, routine model.ManualRoutineToSave) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var routineID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO public.routines (user_id, name, description, source, routine_type, is_predefined, is_public, created_at, updated_at)
		VALUES ($1::uuid, $2, $3, 'manual', $4::public.routine_type, false, false, now(), now())
		RETURNING id::text
	`, routine.UserID, routine.Name, routine.Description, routine.RoutineType).Scan(&routineID); err != nil {
		return "", err
	}

	if err := insertRoutineExercisesAndSets(ctx, tx, routineID, routine.Exercises); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return routineID, nil
}

func (r *routineRepository) UpdateManualRoutine(ctx context.Context, routineID, userID string, routine model.ManualRoutineToSave) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	tag, err := tx.Exec(ctx, `
		UPDATE public.routines
		SET
			name = $1,
			description = $2,
			routine_type = $3::public.routine_type,
			updated_at = now()
		WHERE id = $4::uuid
			AND user_id = $5::uuid
	`, routine.Name, routine.Description, routine.RoutineType, routineID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM public.routine_exercises
		WHERE routine_id = $1::uuid
	`, routineID); err != nil {
		return err
	}

	if err := insertRoutineExercisesAndSets(ctx, tx, routineID, routine.Exercises); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *routineRepository) DeleteRoutine(ctx context.Context, routineID, userID string) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM public.routines
		WHERE id = $1::uuid
			AND user_id = $2::uuid
	`, routineID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DuplicateRoutine copies an existing routine (with its exercises and planned sets) into a new routine owned by the same user.
func (r *routineRepository) DuplicateRoutine(ctx context.Context, routineID, userID string) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	exercises, err := loadRoutineExercisesToSave(ctx, tx, routineID)
	if err != nil {
		return "", err
	}

	var newRoutineID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO public.routines (user_id, name, description, source, routine_type, is_predefined, is_public, created_at, updated_at)
		SELECT $1::uuid, name || ' [Duplicado]', description, source, routine_type, false, false, now(), now()
		FROM public.routines
		WHERE id = $2::uuid
			AND (user_id = $1::uuid OR is_predefined = true)
		RETURNING id::text
	`, userID, routineID).Scan(&newRoutineID); err != nil {
		return "", err
	}

	if err := insertRoutineExercisesAndSets(ctx, tx, newRoutineID, exercises); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return newRoutineID, nil
}

// loadRoutineExercisesToSave reads a routine's exercises and planned sets in the shape used to re-insert them.
func loadRoutineExercisesToSave(ctx context.Context, tx pgx.Tx, routineID string) ([]model.AIRoutineExerciseToSave, error) {
	rows, err := tx.Query(ctx, `
		SELECT
			re.id::text,
			re.exercise_id::text,
			re.exercise_order,
			COALESCE(re.notes, ''),
			res.set_number,
			res.target_reps_min,
			res.target_reps_max,
			COALESCE(res.target_reps_text, ''),
			res.target_weight_kg,
			res.target_duration_seconds,
			res.target_distance_km,
			res.target_rir,
			res.rest_seconds,
			COALESCE(res.notes, '')
		FROM public.routine_exercises re
		LEFT JOIN public.routine_exercise_sets res ON res.routine_exercise_id = re.id
		WHERE re.routine_id = $1::uuid
		ORDER BY re.exercise_order ASC, re.id::text ASC, res.set_number ASC, res.id::text ASC
	`, routineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exercises := make([]model.AIRoutineExerciseToSave, 0)
	exerciseIndex := make(map[string]int)

	for rows.Next() {
		var (
			routineExerciseID     string
			exerciseID            string
			exerciseOrder         int
			exerciseNotes         string
			setNumber             *int
			targetRepsMin         *int
			targetRepsMax         *int
			targetRepsText        string
			targetWeightKg        *float64
			targetDurationSeconds *int
			targetDistanceKm      *float64
			targetRir             *int
			restSeconds           *int
			setNotes              string
		)

		if err := rows.Scan(
			&routineExerciseID,
			&exerciseID,
			&exerciseOrder,
			&exerciseNotes,
			&setNumber,
			&targetRepsMin,
			&targetRepsMax,
			&targetRepsText,
			&targetWeightKg,
			&targetDurationSeconds,
			&targetDistanceKm,
			&targetRir,
			&restSeconds,
			&setNotes,
		); err != nil {
			return nil, err
		}

		index, exists := exerciseIndex[routineExerciseID]
		if !exists {
			exercises = append(exercises, model.AIRoutineExerciseToSave{
				ExerciseID: exerciseID,
				Order:      exerciseOrder,
				Notes:      exerciseNotes,
				Sets:       []model.AIRoutineExerciseSetToSave{},
			})
			index = len(exercises) - 1
			exerciseIndex[routineExerciseID] = index
		}

		if setNumber == nil {
			continue
		}

		exercises[index].Sets = append(exercises[index].Sets, model.AIRoutineExerciseSetToSave{
			SetNumber:             *setNumber,
			TargetRepsMin:         targetRepsMin,
			TargetRepsMax:         targetRepsMax,
			TargetRepsText:        targetRepsText,
			TargetWeightKg:        targetWeightKg,
			TargetDurationSeconds: targetDurationSeconds,
			TargetDistanceKm:      targetDistanceKm,
			TargetRir:             targetRir,
			RestSeconds:           restSeconds,
			Notes:                 setNotes,
		})
	}

	return exercises, rows.Err()
}

