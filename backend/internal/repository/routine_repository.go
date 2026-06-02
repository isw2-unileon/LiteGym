package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RoutineRepository defines the persistence operations for routines.
type RoutineRepository interface {
	ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error)
	ListByUser(ctx context.Context, userID string) ([]model.OverviewRoutineSummary, error)
	GetByID(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error)
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

	return scanRoutineSummaries(rows)
}

func (r *routineRepository) ListByUser(ctx context.Context, userID string) ([]model.OverviewRoutineSummary, error) {
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
		ORDER BY r.name ASC, r.updated_at DESC, r.id::text DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRoutineSummaries(rows)
}

func (r *routineRepository) GetByID(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			r.id::text,
			r.name,
			COALESCE(r.description, ''),
			r.source,
			r.created_at,
			r.updated_at,
			re.id::text,
			re.exercise_id::text,
			COALESCE(e.name, ''),
			COALESCE(e.description, ''),
			COALESCE(e.muscle_group, ''),
			COALESCE(
				(
					SELECT string_agg(esmg.muscle_group, ', ' ORDER BY esmg.muscle_group)
					FROM public.exercise_secondary_muscle_groups esmg
					WHERE esmg.exercise_id = e.id
				),
				''
			),
			COALESCE(e.exercise_type, ''),
			re.exercise_order,
			COALESCE(re.notes, ''),
			res.id::text,
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
		FROM public.routines r
		LEFT JOIN public.routine_exercises re ON re.routine_id = r.id
		LEFT JOIN public.exercises e ON e.id = re.exercise_id AND e.deleted_at IS NULL
		LEFT JOIN public.routine_exercise_sets res ON res.routine_exercise_id = re.id
		WHERE r.user_id = $1::uuid
			AND r.id = $2::uuid
		ORDER BY re.exercise_order ASC, re.id::text ASC, res.set_number ASC, res.id::text ASC
	`, userID, routineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routine *model.RoutineDetail
	exerciseIndex := make(map[string]int)

	for rows.Next() {
		var (
			routineIDValue        string
			routineName           string
			routineDescription    string
			routineSource         string
			routineCreatedAt      time.Time
			routineUpdatedAt      time.Time
			routineExerciseID     sql.NullString
			exerciseID            sql.NullString
			exerciseName          sql.NullString
			exerciseDescription   sql.NullString
			muscleGroup           sql.NullString
			secondaryMuscleGroup  sql.NullString
			exerciseType          sql.NullString
			exerciseOrder         sql.NullInt64
			exerciseNotes         sql.NullString
			setID                 sql.NullString
			setNumber             sql.NullInt64
			targetRepsMin         sql.NullInt64
			targetRepsMax         sql.NullInt64
			targetRepsText        sql.NullString
			targetWeightKg        sql.NullFloat64
			targetDurationSeconds sql.NullInt64
			targetDistanceKm      sql.NullFloat64
			targetRir             sql.NullInt64
			restSeconds           sql.NullInt64
			setNotes              sql.NullString
		)

		if err := rows.Scan(
			&routineIDValue,
			&routineName,
			&routineDescription,
			&routineSource,
			&routineCreatedAt,
			&routineUpdatedAt,
			&routineExerciseID,
			&exerciseID,
			&exerciseName,
			&exerciseDescription,
			&muscleGroup,
			&secondaryMuscleGroup,
			&exerciseType,
			&exerciseOrder,
			&exerciseNotes,
			&setID,
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

		if routine == nil {
			routine = &model.RoutineDetail{
				ID:          routineIDValue,
				Name:        routineName,
				Description: routineDescription,
				Source:      routineSource,
				CreatedAt:   routineCreatedAt,
				UpdatedAt:   routineUpdatedAt,
				Exercises:   []model.RoutineExerciseDetail{},
			}
		}

		if !routineExerciseID.Valid || strings.TrimSpace(routineExerciseID.String) == "" {
			continue
		}

		index, exists := exerciseIndex[routineExerciseID.String]
		if !exists {
			routine.Exercises = append(routine.Exercises, model.RoutineExerciseDetail{
				ID:                   routineExerciseID.String,
				ExerciseID:           exerciseID.String,
				Name:                 exerciseName.String,
				Description:          exerciseDescription.String,
				MuscleGroup:          muscleGroup.String,
				SecondaryMuscleGroup: secondaryMuscleGroup.String,
				ExerciseType:         exerciseType.String,
				ExerciseOrder:        int(exerciseOrder.Int64),
				Notes:                exerciseNotes.String,
				Sets:                 []model.RoutineExerciseSetDetail{},
			})
			index = len(routine.Exercises) - 1
			exerciseIndex[routineExerciseID.String] = index
		}

		if !setID.Valid || strings.TrimSpace(setID.String) == "" {
			continue
		}

		routine.Exercises[index].Sets = append(routine.Exercises[index].Sets, model.RoutineExerciseSetDetail{
			ID:                    setID.String,
			SetNumber:             int(setNumber.Int64),
			TargetRepsMin:         intPointerFromNull(targetRepsMin),
			TargetRepsMax:         intPointerFromNull(targetRepsMax),
			TargetRepsText:        targetRepsText.String,
			TargetWeightKg:        floatPointerFromNull(targetWeightKg),
			TargetDurationSeconds: intPointerFromNull(targetDurationSeconds),
			TargetDistanceKm:      floatPointerFromNull(targetDistanceKm),
			TargetRir:             intPointerFromNull(targetRir),
			RestSeconds:           intPointerFromNull(restSeconds),
			Notes:                 setNotes.String,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if routine == nil {
		return nil, pgx.ErrNoRows
	}

	routine.ExerciseCount = len(routine.Exercises)
	return routine, nil
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

func scanRoutineSummaries(rows pgx.Rows) ([]model.OverviewRoutineSummary, error) {
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
