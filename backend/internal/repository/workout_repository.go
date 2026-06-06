package repository

import (
	"context"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkoutRepository defines the persistence operations for workout sessions, exercises, and sets.
type WorkoutRepository interface {
	CreateSession(ctx context.Context, workout *model.WorkoutSession) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error)
	UpdateSessionByID(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error
	RemoveSessionByID(ctx context.Context, id uuid.UUID) error
	CreateWorkoutExercise(ctx context.Context, workoutExercise *model.WorkoutExercise) error
	GetWorkoutExercisesBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error)
	CreateWorkoutSet(ctx context.Context, workoutSet *model.WorkoutSet) error
	GetWorkoutSetsByWorkoutExerciseID(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error)
	UpdateWorkoutSet(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error
}

type workoutRepository struct {
	db *pgxpool.Pool
}

const (
	// ACSM progression guidance recommends 2-10% load increases when the
	// current workload can be performed above the desired repetition target.
	// These conservative tiers keep increases smaller for small muscle-mass
	// exercises and larger for lower-body compound movements.
	acsmSmallExerciseProgressionRate = 0.02
	acsmUpperCompoundProgressionRate = 0.03
	acsmLowerCompoundProgressionRate = 0.05
	acsmNoExternalLoadProgression    = 0.00

	minimumProgressionAdjustmentKg = 0.5
	weightRoundingIncrementKg      = 0.5
)

// NewWorkoutRepository creates a new WorkoutRepository backed by PostgreSQL.
func NewWorkoutRepository(db *pgxpool.Pool) WorkoutRepository {
	return &workoutRepository{
		db: db,
	}
}

// CreateSession creates a new workout session in the database.
func (wr *workoutRepository) CreateSession(ctx context.Context, workout *model.WorkoutSession) error {
	tx, err := wr.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	query := `
		INSERT INTO workout_sessions (user_id, routine_id, name, performed_at, planned_at, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
		`

	err = tx.QueryRow(
		ctx,
		query,
		workout.UserID,
		workout.RoutineID,
		workout.Name,
		workout.PerformedAt,
		workout.PlannedAt,
		workout.Notes,
	).Scan(&workout.ID, &workout.CreatedAt)

	if err != nil {
		return err
	}

	if workout.RoutineID != nil {
		if err := wr.copyRoutineToWorkout(ctx, tx, workout.UserID, *workout.RoutineID, workout.ID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetSessionByID retrieves a workout session by its ID.
func (wr *workoutRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
	query := `
		SELECT id, user_id, routine_id, name, performed_at, planned_at,
		duration_minutes, calories_burned, notes, created_at
		FROM workout_sessions
		WHERE id = $1
	`

	var workout model.WorkoutSession

	err := wr.db.QueryRow(ctx, query, id).Scan(
		&workout.ID,
		&workout.UserID,
		&workout.RoutineID,
		&workout.Name,
		&workout.PerformedAt,
		&workout.PlannedAt,
		&workout.Duration,
		&workout.CaloriesBurned,
		&workout.Notes,
		&workout.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &workout, nil
}

// UpdateSessionByID updates an existing workout session in the database.
func (wr *workoutRepository) UpdateSessionByID(ctx context.Context, id uuid.UUID, workout *model.WorkoutSession) error {
	query := `
	UPDATE workout_sessions
	SET name = $1, performed_at = $2, planned_at = $3, duration_minutes = $4, calories_burned = $5, notes = $6
	WHERE id = $7
	`

	commandTag, err := wr.db.Exec(
		ctx,
		query,
		workout.Name,
		workout.PerformedAt,
		workout.PlannedAt,
		workout.Duration,
		workout.CaloriesBurned,
		workout.Notes,
		id,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// RemoveSessionByID deletes a workout session from the database by its ID.
func (wr *workoutRepository) RemoveSessionByID(ctx context.Context, id uuid.UUID) error {
	query := `
	DELETE FROM workout_sessions
	WHERE id = $1
	`

	commandTag, err := wr.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// CreateWorkoutExercise creates a new workout exercise associated with a workout session in the database.
func (wr *workoutRepository) CreateWorkoutExercise(ctx context.Context, workoutExercise *model.WorkoutExercise) error {
	query := `
	INSERT INTO workout_exercises (workout_session_id, exercise_id, routine_exercise_id, exercise_order, notes)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at
	`

	err := wr.db.QueryRow(
		ctx,
		query,
		workoutExercise.WorkoutSessionID,
		workoutExercise.ExerciseID,
		workoutExercise.RoutineExerciseID,
		workoutExercise.ExerciseOrder,
		workoutExercise.Notes,
	).Scan(&workoutExercise.ID, &workoutExercise.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

// GetWorkoutExercisesBySessionID retrieves all workout exercises associated with a specific workout session ID, ordered by their exercise order.
func (wr *workoutRepository) GetWorkoutExercisesBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error) {
	query := `
	SELECT id::text, workout_session_id, exercise_id, routine_exercise_id, exercise_order, notes, created_at
	FROM workout_exercises
	WHERE workout_session_id = $1
	ORDER BY exercise_order
	`

	rows, err := wr.db.Query(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workoutExercises []*model.WorkoutExercise

	for rows.Next() {
		var we model.WorkoutExercise

		err := rows.Scan(
			&we.ID,
			&we.WorkoutSessionID,
			&we.ExerciseID,
			&we.RoutineExerciseID,
			&we.ExerciseOrder,
			&we.Notes,
			&we.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		workoutExercises = append(workoutExercises, &we)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return workoutExercises, nil
}

// CreateWorkoutSet creates a new workout set associated with a workout exercise in the database.
func (wr *workoutRepository) CreateWorkoutSet(ctx context.Context, workoutSet *model.WorkoutSet) error {
	query := `
	INSERT INTO workout_sets (
		workout_exercise_id,
		routine_exercise_set_id,
		set_number,
		target_reps_min,
		target_reps_max,
		target_reps_text,
		target_weight_kg,
		target_duration_seconds,
		target_distance_km,
		target_rir,
		rest_seconds,
		reps,
		weight_kg,
		duration_seconds,
		distance_km,
		rir,
		completed
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	RETURNING id, created_at
	`

	err := wr.db.QueryRow(
		ctx,
		query,
		workoutSet.WorkoutExerciseID,
		workoutSet.RoutineExerciseSetID,
		workoutSet.SetNumber,
		workoutSet.TargetRepsMin,
		workoutSet.TargetRepsMax,
		workoutSet.TargetRepsText,
		workoutSet.TargetWeightKg,
		workoutSet.TargetDurationSeconds,
		workoutSet.TargetDistanceKm,
		workoutSet.TargetRir,
		workoutSet.RestSeconds,
		workoutSet.Repetitions,
		workoutSet.WeightKg,
		workoutSet.Duration,
		workoutSet.DistanceKm,
		workoutSet.Rir,
		workoutSet.Completed,
	).Scan(&workoutSet.ID, &workoutSet.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

// GetWorkoutSetsByWorkoutExerciseID retrieves all workout sets associated with a specific workout exercise ID, ordered by their set number.
func (wr *workoutRepository) GetWorkoutSetsByWorkoutExerciseID(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error) {
	query := `
	SELECT
		id::text,
		workout_exercise_id,
		routine_exercise_set_id,
		set_number,
		target_reps_min,
		target_reps_max,
		COALESCE(target_reps_text, ''),
		target_weight_kg,
		target_duration_seconds,
		target_distance_km,
		target_rir,
		rest_seconds,
		reps,
		weight_kg,
		duration_seconds,
		distance_km,
		rir,
		completed,
		created_at
	FROM workout_sets
	WHERE workout_exercise_id = $1
	ORDER BY set_number
	`

	rows, err := wr.db.Query(ctx, query, exerciseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workoutSets []*model.WorkoutSet

	for rows.Next() {
		var ws model.WorkoutSet

		err := rows.Scan(
			&ws.ID,
			&ws.WorkoutExerciseID,
			&ws.RoutineExerciseSetID,
			&ws.SetNumber,
			&ws.TargetRepsMin,
			&ws.TargetRepsMax,
			&ws.TargetRepsText,
			&ws.TargetWeightKg,
			&ws.TargetDurationSeconds,
			&ws.TargetDistanceKm,
			&ws.TargetRir,
			&ws.RestSeconds,
			&ws.Repetitions,
			&ws.WeightKg,
			&ws.Duration,
			&ws.DistanceKm,
			&ws.Rir,
			&ws.Completed,
			&ws.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		workoutSets = append(workoutSets, &ws)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return workoutSets, nil
}

// UpdateWorkoutSet updates an existing workout set in the database by its exercise ID and set number.
func (wr *workoutRepository) UpdateWorkoutSet(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
	query := `
	UPDATE workout_sets
	SET reps = $1, weight_kg = $2, duration_seconds = $3, distance_km = $4, rir = $5, completed = $6
	WHERE id = $7 AND set_number = $8
	`

	commandTag, err := wr.db.Exec(
		ctx,
		query,
		set.Repetitions,
		set.WeightKg,
		set.Duration,
		set.DistanceKm,
		set.Rir,
		set.Completed,
		setID,
		setNumber,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

type routineExercisePrescription struct {
	RoutineExerciseID string
	ExerciseID        uuid.UUID
	MuscleGroup       string
	ExerciseType      string
	ExerciseOrder     int
	Notes             string
}

type routineSetPrescription struct {
	RoutineExerciseSetID  uuid.UUID
	SetNumber             int
	TargetRepsMin         *int
	TargetRepsMax         *int
	TargetRepsText        string
	TargetWeightKg        *float64
	TargetDurationSeconds *int
	TargetDistanceKm      *float64
	TargetRir             *int
	RestSeconds           *int
}

type recentWorkoutSetPerformance struct {
	SetNumber int
	Reps      *int
	WeightKg  *float64
	Rir       *int
}

type workoutTx interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (wr *workoutRepository) copyRoutineToWorkout(
	ctx context.Context,
	tx workoutTx,
	userID uuid.UUID,
	routineID uuid.UUID,
	workoutSessionID uuid.UUID,
) error {
	rows, err := tx.Query(ctx, `
		SELECT
			re.id::text,
			re.exercise_id,
			e.muscle_group,
			COALESCE(e.exercise_type, ''),
			re.exercise_order,
			COALESCE(re.notes, '')
		FROM public.routine_exercises re
		INNER JOIN public.exercises e ON e.id = re.exercise_id
		WHERE re.routine_id = $1
		ORDER BY re.exercise_order, re.id::text
	`, routineID)
	if err != nil {
		return err
	}

	exercises := make([]routineExercisePrescription, 0)

	for rows.Next() {
		var prescription routineExercisePrescription
		if err := rows.Scan(
			&prescription.RoutineExerciseID,
			&prescription.ExerciseID,
			&prescription.MuscleGroup,
			&prescription.ExerciseType,
			&prescription.ExerciseOrder,
			&prescription.Notes,
		); err != nil {
			rows.Close()
			return err
		}
		exercises = append(exercises, prescription)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, prescription := range exercises {
		var workoutExerciseID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO public.workout_exercises (
				workout_session_id,
				exercise_id,
				routine_exercise_id,
				exercise_order,
				notes
			)
			VALUES ($1, $2, $3::uuid, $4, $5)
			RETURNING id
		`,
			workoutSessionID,
			prescription.ExerciseID,
			prescription.RoutineExerciseID,
			prescription.ExerciseOrder,
			prescription.Notes,
		).Scan(&workoutExerciseID); err != nil {
			return err
		}

		if err := wr.copyRoutineSetsToWorkout(ctx, tx, userID, workoutSessionID, prescription, workoutExerciseID); err != nil {
			return err
		}
	}

	return nil
}

func (wr *workoutRepository) copyRoutineSetsToWorkout(
	ctx context.Context,
	tx workoutTx,
	userID uuid.UUID,
	workoutSessionID uuid.UUID,
	exercise routineExercisePrescription,
	workoutExerciseID uuid.UUID,
) error {
	recentSets, err := wr.listLatestExercisePerformance(ctx, tx, userID, exercise.ExerciseID, workoutSessionID)
	if err != nil {
		return err
	}

	rows, err := tx.Query(ctx, `
		SELECT
			id,
			set_number,
			target_reps_min,
			target_reps_max,
			COALESCE(target_reps_text, ''),
			target_weight_kg,
			target_duration_seconds,
			target_distance_km,
			target_rir,
		rest_seconds
		FROM public.routine_exercise_sets
		WHERE routine_exercise_id = $1::uuid
		ORDER BY set_number, id::text
	`, exercise.RoutineExerciseID)
	if err != nil {
		return err
	}

	prescriptions := make([]routineSetPrescription, 0)

	for rows.Next() {
		var prescription routineSetPrescription
		if err := rows.Scan(
			&prescription.RoutineExerciseSetID,
			&prescription.SetNumber,
			&prescription.TargetRepsMin,
			&prescription.TargetRepsMax,
			&prescription.TargetRepsText,
			&prescription.TargetWeightKg,
			&prescription.TargetDurationSeconds,
			&prescription.TargetDistanceKm,
			&prescription.TargetRir,
			&prescription.RestSeconds,
		); err != nil {
			rows.Close()
			return err
		}
		prescriptions = append(prescriptions, prescription)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, prescription := range prescriptions {
		completed := false
		targetWeightKg := calculateDynamicTargetWeight(exercise, prescription, recentSets)
		var workoutSetID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO public.workout_sets (
				workout_exercise_id,
				routine_exercise_set_id,
				set_number,
				target_reps_min,
				target_reps_max,
				target_reps_text,
				target_weight_kg,
				target_duration_seconds,
				target_distance_km,
				target_rir,
				rest_seconds,
				completed
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id
		`,
			workoutExerciseID,
			prescription.RoutineExerciseSetID,
			prescription.SetNumber,
			prescription.TargetRepsMin,
			prescription.TargetRepsMax,
			prescription.TargetRepsText,
			targetWeightKg,
			prescription.TargetDurationSeconds,
			prescription.TargetDistanceKm,
			prescription.TargetRir,
			prescription.RestSeconds,
			completed,
		).Scan(&workoutSetID); err != nil {
			return err
		}
	}

	return nil
}

func (wr *workoutRepository) listLatestExercisePerformance(
	ctx context.Context,
	tx workoutTx,
	userID uuid.UUID,
	exerciseID uuid.UUID,
	excludedWorkoutSessionID uuid.UUID,
) ([]recentWorkoutSetPerformance, error) {
	rows, err := tx.Query(ctx, `
		WITH latest_workout_exercise AS (
			SELECT we.id
			FROM public.workout_exercises we
			INNER JOIN public.workout_sessions ws ON ws.id = we.workout_session_id
			WHERE ws.user_id = $1
				AND we.exercise_id = $2
				AND ws.id <> $3
				AND ws.performed_at IS NOT NULL
				AND EXISTS (
					SELECT 1
					FROM public.workout_sets completed_sets
					WHERE completed_sets.workout_exercise_id = we.id
						AND completed_sets.completed = true
						AND completed_sets.weight_kg IS NOT NULL
				)
			ORDER BY ws.performed_at DESC, ws.id::text DESC, we.exercise_order ASC
			LIMIT 1
		)
		SELECT wset.set_number, wset.reps, wset.weight_kg, wset.rir
		FROM public.workout_sets wset
		INNER JOIN latest_workout_exercise lwe ON lwe.id = wset.workout_exercise_id
		WHERE wset.completed = true
		ORDER BY wset.set_number ASC
	`, userID, exerciseID, excludedWorkoutSessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sets := make([]recentWorkoutSetPerformance, 0)
	for rows.Next() {
		var set recentWorkoutSetPerformance
		if err := rows.Scan(&set.SetNumber, &set.Reps, &set.WeightKg, &set.Rir); err != nil {
			return nil, err
		}
		sets = append(sets, set)
	}

	return sets, rows.Err()
}

func calculateDynamicTargetWeight(
	exercise routineExercisePrescription,
	prescription routineSetPrescription,
	recentSets []recentWorkoutSetPerformance,
) *float64 {
	baseline := historicalWeightBaseline(prescription.SetNumber, recentSets)
	if baseline == nil {
		baseline = prescription.TargetWeightKg
	}
	if baseline == nil {
		return nil
	}

	targetWeight := *baseline
	adjustment := progressionWeightAdjustment(exercise, targetWeight)
	switch {
	case shouldReduceTargetWeight(prescription, recentSets):
		targetWeight = math.Max(0, targetWeight-adjustment)
	case shouldIncreaseTargetWeight(prescription, recentSets):
		targetWeight += adjustment
	}

	rounded := roundWeightToIncrement(targetWeight)
	return &rounded
}

func progressionWeightAdjustment(exercise routineExercisePrescription, baselineWeight float64) float64 {
	if baselineWeight <= 0 {
		return 0
	}

	rate := progressionRate(exercise)
	if rate <= 0 {
		return 0
	}

	adjustment := roundWeightToIncrement(baselineWeight * rate)
	if adjustment < minimumProgressionAdjustmentKg {
		return minimumProgressionAdjustmentKg
	}
	return adjustment
}

func progressionRate(exercise routineExercisePrescription) float64 {
	exerciseType := strings.ToLower(exercise.ExerciseType)
	muscleGroup := strings.ToLower(exercise.MuscleGroup)

	switch {
	case exerciseType == "bodyweight" || exerciseType == "isometric":
		return acsmNoExternalLoadProgression
	case muscleGroup == "legs" || muscleGroup == "glutes":
		return acsmLowerCompoundProgressionRate
	case exerciseType == "strength":
		return acsmUpperCompoundProgressionRate
	case exerciseType == "hypertrophy":
		return acsmSmallExerciseProgressionRate
	case muscleGroup == "biceps" || muscleGroup == "triceps" || muscleGroup == "shoulders":
		return acsmSmallExerciseProgressionRate
	default:
		return acsmUpperCompoundProgressionRate
	}
}

func roundWeightToIncrement(weight float64) float64 {
	return math.Round(weight/weightRoundingIncrementKg) * weightRoundingIncrementKg
}

func historicalWeightBaseline(setNumber int, recentSets []recentWorkoutSetPerformance) *float64 {
	var fallback *float64
	for _, set := range recentSets {
		if set.WeightKg == nil {
			continue
		}
		if fallback == nil || *set.WeightKg > *fallback {
			weight := *set.WeightKg
			fallback = &weight
		}
		if set.SetNumber == setNumber {
			weight := *set.WeightKg
			return &weight
		}
	}
	return fallback
}

func shouldIncreaseTargetWeight(prescription routineSetPrescription, recentSets []recentWorkoutSetPerformance) bool {
	targetReps := targetRepsCeiling(prescription)
	if len(recentSets) == 0 || targetReps == 0 {
		return false
	}

	for _, set := range recentSets {
		if set.Reps == nil || *set.Reps < targetReps {
			return false
		}
		if prescription.TargetRir != nil && set.Rir != nil && *set.Rir < *prescription.TargetRir {
			return false
		}
	}
	return true
}

func shouldReduceTargetWeight(prescription routineSetPrescription, recentSets []recentWorkoutSetPerformance) bool {
	targetReps := targetRepsFloor(prescription)
	if len(recentSets) == 0 {
		return false
	}

	for _, set := range recentSets {
		if targetReps > 0 && set.Reps != nil && *set.Reps < targetReps {
			return true
		}
		if set.Rir != nil && *set.Rir <= 0 {
			return true
		}
	}
	return false
}

func targetRepsFloor(prescription routineSetPrescription) int {
	if prescription.TargetRepsMin != nil {
		return *prescription.TargetRepsMin
	}
	if prescription.TargetRepsMax != nil {
		return *prescription.TargetRepsMax
	}
	return 0
}

func targetRepsCeiling(prescription routineSetPrescription) int {
	if prescription.TargetRepsMax != nil {
		return *prescription.TargetRepsMax
	}
	if prescription.TargetRepsMin != nil {
		return *prescription.TargetRepsMin
	}
	return 0
}
