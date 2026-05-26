package repository

import (
	"context"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RoutineRepository defines the persistence operations for routines.
type RoutineRepository interface {
	ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error)
	ListByUser(ctx context.Context, userID string) ([]model.OverviewRoutineSummary, error)
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
