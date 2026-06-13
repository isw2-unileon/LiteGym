package repository

import (
	"context"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BodyMetricRepository defines the persistence operations for body metrics.
type BodyMetricRepository interface {
	ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewBodyMetricEntry, error)
}

type bodyMetricRepository struct {
	db *pgxpool.Pool
}

// NewBodyMetricRepository creates a new body metric repository backed by PostgreSQL.
func NewBodyMetricRepository(db *pgxpool.Pool) BodyMetricRepository {
	return &bodyMetricRepository{db: db}
}

func (r *bodyMetricRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewBodyMetricEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			recorded_at,
			weight_kg,
			body_fat_percentage,
			muscle_mass_kg
		FROM public.body_metrics
		WHERE user_id = $1::uuid
		ORDER BY recorded_at DESC, id DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]model.OverviewBodyMetricEntry, 0)
	for rows.Next() {
		var (
			entry      model.OverviewBodyMetricEntry
			weight     pgtype.Numeric
			bodyFat    pgtype.Numeric
			muscleMass pgtype.Numeric
		)

		if err := rows.Scan(&entry.RecordedAt, &weight, &bodyFat, &muscleMass); err != nil {
			return nil, err
		}

		entry.WeightKg = numericToFloatPointer(weight)
		entry.BodyFatPercentage = numericToFloatPointer(bodyFat)
		entry.MuscleMassKg = numericToFloatPointer(muscleMass)
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func numericToFloatPointer(value pgtype.Numeric) *float64 {
	if !value.Valid {
		return nil
	}

	floatValue, err := value.Float64Value()
	if err != nil || !floatValue.Valid {
		return nil
	}

	number := floatValue.Float64
	return &number
}
