package repository

import (
	"context"
	"strconv"
	"strings"
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
	return listCompletedMuscleDistributionByUser(ctx, db, userID, &from, &to)
}

func listCompletedMuscleDistributionByUser(
	ctx context.Context,
	db *pgxpool.Pool,
	userID string,
	from, to *time.Time,
) ([]model.OverviewMuscleGroupShare, int, error) {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT
			e.muscle_group,
			COUNT(ws_sets.id)::int
		FROM public.workout_sessions ws
		INNER JOIN public.workout_exercises we ON we.workout_session_id = ws.id
		INNER JOIN public.workout_sets ws_sets ON ws_sets.workout_exercise_id = we.id
		INNER JOIN public.exercises e ON e.id = we.exercise_id
		WHERE ws.user_id = $1::uuid
			AND ws_sets.completed = true
	`)

	args := []any{userID}
	if from != nil {
		args = append(args, *from)
		placeholder := len(args)
		queryBuilder.WriteString(`
			AND ws.performed_at >= $`)
		queryBuilder.WriteString(strconv.Itoa(placeholder))
		queryBuilder.WriteString(`
		`)
	}
	if to != nil {
		args = append(args, *to)
		placeholder := len(args)
		queryBuilder.WriteString(`
			AND ws.performed_at < $`)
		queryBuilder.WriteString(strconv.Itoa(placeholder))
		queryBuilder.WriteString(`
		`)
	}

	queryBuilder.WriteString(`
		GROUP BY e.muscle_group
		ORDER BY COUNT(*) DESC, e.muscle_group ASC
	`)

	rows, err := db.Query(ctx, queryBuilder.String(), args...)
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
