package repository

import (
	"context"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func listOverviewWorkoutSummaries(ctx context.Context, db *pgxpool.Pool, userID string, limit int) ([]model.OverviewWorkoutSummary, error) {
	rows, err := db.Query(ctx, `
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

func listMuscleDistributionByUser(ctx context.Context, db *pgxpool.Pool, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error) {
	rows, err := db.Query(ctx, `
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
