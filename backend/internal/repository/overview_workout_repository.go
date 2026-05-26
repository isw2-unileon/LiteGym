package repository

import (
	"context"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OverviewWorkoutRepository defines read-only workout queries used by the dashboard overview.
type OverviewWorkoutRepository interface {
	ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error)
	ListTrainingDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error)
	ListPlannedDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error)
	ListCalendarWorkoutsByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewCalendarWorkout, error)
	ListMuscleDistributionByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error)
}

type overviewWorkoutRepository struct {
	db *pgxpool.Pool
}

// NewOverviewWorkoutRepository creates a new dashboard workout repository backed by PostgreSQL.
func NewOverviewWorkoutRepository(db *pgxpool.Pool) OverviewWorkoutRepository {
	return &overviewWorkoutRepository{db: db}
}

func (r *overviewWorkoutRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			ws.id::text,
			COALESCE(ws.name, ''),
			COALESCE(rt.name, ''),
			ws.performed_at,
			COALESCE(ws.duration_minutes, 0),
			COUNT(we.id)::int
		FROM public.workout_sessions ws
		LEFT JOIN public.routines rt ON rt.id = ws.routine_id
		LEFT JOIN public.workout_exercises we ON we.workout_session_id = ws.id
		WHERE ws.user_id = $1::uuid
			AND ws.performed_at IS NOT NULL
		GROUP BY ws.id, ws.name, rt.name, ws.performed_at, ws.duration_minutes
		ORDER BY ws.performed_at DESC, ws.id::text DESC
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
			&workout.PerformedAt,
			&workout.DurationMinutes,
			&workout.ExerciseCount,
		); err != nil {
			return nil, err
		}
		workouts = append(workouts, workout)
	}

	return workouts, rows.Err()
}

func (r *overviewWorkoutRepository) ListTrainingDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
	return r.listDatesInRange(ctx, userID, from, to, `
		SELECT DISTINCT DATE(ws.performed_at)
		FROM public.workout_sessions ws
		WHERE ws.user_id = $1::uuid
			AND ws.performed_at >= $2
			AND ws.performed_at < $3
		ORDER BY DATE(ws.performed_at) DESC
	`)
}

func (r *overviewWorkoutRepository) ListPlannedDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
	return r.listDatesInRange(ctx, userID, from, to, `
		SELECT DISTINCT DATE(ws.planned_at)
		FROM public.workout_sessions ws
		WHERE ws.user_id = $1::uuid
			AND ws.planned_at >= $2
			AND ws.planned_at < $3
			AND ws.performed_at IS NULL
		ORDER BY DATE(ws.planned_at) DESC
	`)
}

func (r *overviewWorkoutRepository) ListCalendarWorkoutsByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewCalendarWorkout, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			ws.id::text,
			COALESCE(ws.name, ''),
			COALESCE(ws.routine_id::text, ''),
			COALESCE(rt.name, ''),
			ws.performed_at,
			ws.planned_at,
			COALESCE(ws.duration_minutes, 0),
			COUNT(we.id)::int
		FROM public.workout_sessions ws
		LEFT JOIN public.routines rt ON rt.id = ws.routine_id
		LEFT JOIN public.workout_exercises we ON we.workout_session_id = ws.id
		WHERE ws.user_id = $1::uuid
			AND (
				(ws.performed_at >= $2 AND ws.performed_at < $3)
				OR (ws.performed_at IS NULL AND ws.planned_at >= $2 AND ws.planned_at < $3)
			)
		GROUP BY ws.id, ws.name, ws.routine_id, rt.name, ws.performed_at, ws.planned_at, ws.duration_minutes
		ORDER BY COALESCE(ws.performed_at, ws.planned_at) ASC, ws.id::text ASC
	`, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workouts := make([]model.OverviewCalendarWorkout, 0)
	for rows.Next() {
		var workout model.OverviewCalendarWorkout
		if err := rows.Scan(
			&workout.ID,
			&workout.Name,
			&workout.RoutineID,
			&workout.RoutineName,
			&workout.PerformedAt,
			&workout.PlannedAt,
			&workout.DurationMinutes,
			&workout.ExerciseCount,
		); err != nil {
			return nil, err
		}
		workouts = append(workouts, workout)
	}

	return workouts, rows.Err()
}

func (r *overviewWorkoutRepository) listDatesInRange(ctx context.Context, userID string, from, to time.Time, query string) ([]time.Time, error) {
	rows, err := r.db.Query(ctx, query, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dates := make([]time.Time, 0)
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err != nil {
			return nil, err
		}
		dates = append(dates, date)
	}

	return dates, rows.Err()
}

func (r *overviewWorkoutRepository) ListMuscleDistributionByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			e.muscle_group,
			COUNT(*)::int
		FROM public.workout_sessions ws
		INNER JOIN public.workout_exercises we ON we.workout_session_id = ws.id
		INNER JOIN public.exercises e ON e.id = we.exercise_id
		WHERE ws.user_id = $1::uuid
			AND ws.performed_at >= $2
			AND ws.performed_at < $3
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
