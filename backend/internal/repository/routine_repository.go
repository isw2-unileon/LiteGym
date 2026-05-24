package repository

import (
	"context"
	"strings"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RoutineRepository defines the persistence operations for routines.
type RoutineRepository interface {
	ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error)
	SaveGeneratedAIRoutine(ctx context.Context, routine model.AIRoutineToSave) (string, error)
	CountAIGenerationsInWindow(ctx context.Context, userID string, since time.Time) (int, error)
	LogAIGeneration(ctx context.Context, userID string, createdAt time.Time) error
	ListAvailableExercisesForAI(
		ctx context.Context,
		userID string,
		targetMuscleGroups []string,
		limit int,
	) ([]model.Exercise, error)
}

type routineRepository struct {
	db *pgxpool.Pool
}

// NewRoutineRepository creates a new routine repository backed by PostgreSQL.
func NewRoutineRepository(db *pgxpool.Pool) RoutineRepository {
	return &routineRepository{db: db}
}

func (r *routineRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			r.id::text,
			r.name,
			COALESCE(r.description, ''),
			COUNT(re.id)::int,
			r.updated_at
		FROM public.routines r
		LEFT JOIN public.routine_exercises re ON re.routine_id = r.id
		WHERE r.user_id = $1::uuid
		GROUP BY r.id, r.name, r.description, r.updated_at, r.created_at
		ORDER BY r.updated_at DESC, r.created_at DESC, r.id::text DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	routines := make([]model.OverviewRoutineSummary, 0)
	for rows.Next() {
		var routine model.OverviewRoutineSummary
		if err := rows.Scan(
			&routine.ID,
			&routine.Name,
			&routine.Description,
			&routine.ExerciseCount,
			&routine.UpdatedAt,
		); err != nil {
			return nil, err
		}
		routines = append(routines, routine)
	}

	return routines, rows.Err()
}

func (r *routineRepository) SaveGeneratedAIRoutine(ctx context.Context, routine model.AIRoutineToSave) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var routineID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO public.routines (user_id, name, description, source, is_predefined, is_public, created_at, updated_at)
		VALUES ($1::uuid, $2, $3, 'ai', false, false, now(), now())
		RETURNING id::text
	`, routine.UserID, routine.Name, routine.Description).Scan(&routineID); err != nil {
		return "", err
	}

	batch := &pgx.Batch{}
	plannedSetCount := 0
	for _, exercise := range routine.Exercises {
		var routineExerciseID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO public.routine_exercises (routine_id, exercise_id, exercise_order, notes)
			VALUES ($1::uuid, $2::uuid, $3, $4)
			RETURNING id::text
		`, routineID, exercise.ExerciseID, exercise.Order, exercise.Notes).Scan(&routineExerciseID); err != nil {
			return "", err
		}

		for _, set := range exercise.Sets {
			batch.Queue(`
				INSERT INTO public.routine_exercise_sets (
					routine_exercise_id,
					set_number,
					target_reps_min,
					target_reps_max,
					target_reps_text,
					target_weight_kg,
					target_duration_seconds,
					target_distance_km,
					target_rir,
					rest_seconds,
					notes
				)
				VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			`,
				routineExerciseID,
				set.SetNumber,
				set.TargetRepsMin,
				set.TargetRepsMax,
				set.TargetRepsText,
				set.TargetWeightKg,
				set.TargetDurationSeconds,
				set.TargetDistanceKm,
				set.TargetRir,
				set.RestSeconds,
				set.Notes,
			)
			plannedSetCount++
		}
	}

	if plannedSetCount > 0 {
		results := tx.SendBatch(ctx, batch)
		for range plannedSetCount {
			if _, err := results.Exec(); err != nil {
				_ = results.Close()
				return "", err
			}
		}
		if err := results.Close(); err != nil {
			return "", err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return routineID, nil
}

func (r *routineRepository) CountAIGenerationsInWindow(
	ctx context.Context,
	userID string,
	since time.Time,
) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(1)::int
		FROM public.ai_routine_generation_logs
		WHERE user_id = $1::uuid
			AND created_at >= $2
	`, userID, since).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *routineRepository) LogAIGeneration(
	ctx context.Context,
	userID string,
	createdAt time.Time,
) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO public.ai_routine_generation_logs (user_id, created_at)
		VALUES ($1::uuid, $2)
	`, userID, createdAt)
	return err
}

func (r *routineRepository) ListAvailableExercisesForAI(
	ctx context.Context,
	userID string,
	targetMuscleGroups []string,
	limit int,
) ([]model.Exercise, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			e.id::text,
			e.name,
			COALESCE(e.description, ''),
			e.muscle_group,
			COALESCE(e.exercise_type, ''),
			e.is_official
		FROM public.exercises e
		WHERE e.deleted_at IS NULL
			AND (e.is_official = true OR e.owner_user_id = $1::uuid)
			AND (
				$2::text[] IS NULL
				OR cardinality($2::text[]) = 0
				OR e.muscle_group = ANY($2::text[])
			)
		ORDER BY e.is_official DESC, LOWER(e.name) ASC
		LIMIT $3
	`, userID, normalizeStringList(targetMuscleGroups), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exercises := make([]model.Exercise, 0)
	for rows.Next() {
		var exercise model.Exercise
		var exerciseType string
		if err := rows.Scan(
			&exercise.ID,
			&exercise.Name,
			&exercise.Description,
			&exercise.MuscleGroup,
			&exerciseType,
			&exercise.IsOfficial,
		); err != nil {
			return nil, err
		}
		if strings.TrimSpace(exerciseType) != "" {
			exercise.ExerciseType = exerciseType
		}
		exercises = append(exercises, exercise)
	}

	return exercises, rows.Err()
}

func normalizeStringList(items []string) []string {
	if len(items) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return normalized
}
