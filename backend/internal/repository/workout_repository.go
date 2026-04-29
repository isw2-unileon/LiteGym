package repository

import (
	"context"
	"time"

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

// NewWorkoutRepository creates a new WorkoutRepository backed by PostgreSQL.
func NewWorkoutRepository(db *pgxpool.Pool) WorkoutRepository {
	return &workoutRepository{
		db: db,
	}
}

// CreateSession creates a new workout session in the database.
func (wr *workoutRepository) CreateSession(ctx context.Context, workout *model.WorkoutSession) error {
	query := `
		INSERT INTO workout_sessions (user_id, routine_id, name, started_at, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
		`

	err := wr.db.QueryRow(
		ctx,
		query,
		workout.UserID,
		workout.RoutineID,
		workout.Name,
		time.Now(),
		workout.Notes,
	).Scan(&workout.ID, &workout.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

// GetSessionByID retrieves a workout session by its ID.
func (wr *workoutRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
	query := `
		SELECT id, user_id, routine_id, name, started_at, ended_at, 
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
		&workout.StartedAt,
		&workout.EndedAt,
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
	SET name = $1, ended_at = $2, duration_minutes = $3, calories_burned = $4, notes = $5
	WHERE id = $6
	`

	commandTag, err := wr.db.Exec(
		ctx,
		query,
		workout.Name,
		time.Now(),
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
	INSERT INTO workout_exercises (workout_session_id, exercise_id, exercise_order, notes)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at
	`

	err := wr.db.QueryRow(
		ctx,
		query,
		workoutExercise.WorkoutSessionID,
		workoutExercise.ExerciseID,
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
	SELECT id::text, workout_session_id, exercise_id, exercise_order, notes, created_at
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
	INSERT INTO workout_sets (workout_exercise_id, set_number, reps, weight_kg, duration_seconds,
	                          distance_km, rir, completed)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id, created_at
	`

	err := wr.db.QueryRow(
		ctx,
		query,
		workoutSet.WorkoutExerciseID,
		workoutSet.SetNumber,
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
	SELECT id::text, workout_exercise_id, set_number, reps, weight_kg, duration_seconds,
	       distance_km, rir, completed, created_at
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
			&ws.SetNumber,
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
