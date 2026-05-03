package repository

import (
	"context"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkoutSessionRepository defines the persistence operations for workout sessions.
type WorkoutSessionRepository interface {
	ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error)
	ListTrainingDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error)
	ListMuscleDistributionByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error)
}

type workoutSessionRepository struct {
	db *pgxpool.Pool
}

// NewWorkoutSessionRepository creates a new workout session repository backed by PostgreSQL.
func NewWorkoutSessionRepository(db *pgxpool.Pool) WorkoutSessionRepository {
	return &workoutSessionRepository{db: db}
}

func (r *workoutSessionRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			ws.id::text,
			COALESCE(ws.name, ''),
			COALESCE(rt.name, ''),
			ws.started_at,
			COALESCE(ws.duration_minutes, 0),
			COUNT(we.id)::int
		FROM public.workout_sessions ws
		LEFT JOIN public.routines rt ON rt.id = ws.routine_id
		LEFT JOIN public.workout_exercises we ON we.workout_session_id = ws.id
		WHERE ws.user_id = $1::uuid
		GROUP BY ws.id, ws.name, rt.name, ws.started_at, ws.duration_minutes
		ORDER BY ws.started_at DESC, ws.id::text DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workouts := make([]model.OverviewWorkoutSummary, 0)
	for rows.Next() {
		var workout model.OverviewWorkoutSummary
		if err := rows.Scan(
			&workout.ID,
			&workout.Name,
			&workout.RoutineName,
			&workout.StartedAt,
			&workout.DurationMinutes,
			&workout.ExerciseCount,
		); err != nil {
			return nil, err
		}
		workouts = append(workouts, workout)
	}

	return workouts, rows.Err()
}

func (r *workoutSessionRepository) ListTrainingDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT DATE(ws.started_at)
		FROM public.workout_sessions ws
		WHERE ws.user_id = $1::uuid
			AND ws.started_at >= $2
			AND ws.started_at < $3
		ORDER BY DATE(ws.started_at) DESC
	`, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dates := make([]time.Time, 0)
	for rows.Next() {
		var trainedDate time.Time
		if err := rows.Scan(&trainedDate); err != nil {
			return nil, err
		}
		dates = append(dates, trainedDate)
	}

	return dates, rows.Err()
}

func (r *workoutSessionRepository) ListMuscleDistributionByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			e.muscle_group,
			COUNT(*)::int
		FROM public.workout_sessions ws
		INNER JOIN public.workout_exercises we ON we.workout_session_id = ws.id
		INNER JOIN public.exercises e ON e.id = we.exercise_id
		WHERE ws.user_id = $1::uuid
			AND ws.started_at >= $2
			AND ws.started_at < $3
		GROUP BY e.muscle_group
		ORDER BY COUNT(*) DESC, e.muscle_group ASC
	`, userID, from, to)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	shares := make([]model.OverviewMuscleGroupShare, 0)
	total := 0
	for rows.Next() {
		var share model.OverviewMuscleGroupShare
		if err := rows.Scan(&share.Name, &share.Count); err != nil {
			return nil, 0, err
		}
		total += share.Count
		shares = append(shares, share)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	for index := range shares {
		if total == 0 {
			shares[index].Percentage = 0
			continue
		}
		shares[index].Percentage = int(float64(shares[index].Count)*100/float64(total) + 0.5)
	}

	return shares, total, nil
}
