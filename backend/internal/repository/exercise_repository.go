package repository

import (
	"context"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExerciseRepository interface {
	Create(ctx context.Context, exercise *model.Exercise) error
	GetByID(ctx context.Context, id int64) (*model.Exercise, error)
	List(ctx context.Context) ([]model.Exercise, error)
}

type exerciseRepository struct {
	db *pgxpool.Pool
}

func NewExerciseRepository(db *pgxpool.Pool) ExerciseRepository {
	return &exerciseRepository{
		db: db,
	}
}

func (r *exerciseRepository) Create(ctx context.Context, exercise *model.Exercise) error {
	query := `
		INSERT INTO exercises (
			name,
			description,
			muscle_group,
			secondary_muscle_group,
			exercise_type,
			is_official
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		exercise.Name,
		exercise.Description,
		exercise.MuscleGroup,
		exercise.SecondaryMuscleGroup,
		exercise.ExerciseType,
		exercise.IsOfficial,
	).Scan(&exercise.ID, &exercise.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *exerciseRepository) GetByID(ctx context.Context, id int64) (*model.Exercise, error) {
	query := `
		SELECT
			id,
			name,
			description,
			muscle_group,
			secondary_muscle_group,
			exercise_type,
			is_official,
			created_at
		FROM exercises
		WHERE id = $1
	`

	var exercise model.Exercise

	err := r.db.QueryRow(ctx, query, id).Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.Description,
		&exercise.MuscleGroup,
		&exercise.SecondaryMuscleGroup,
		&exercise.ExerciseType,
		&exercise.IsOfficial,
		&exercise.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &exercise, nil
}

func (r *exerciseRepository) List(ctx context.Context) ([]model.Exercise, error) {
	query := `
		SELECT
			id,
			name,
			description,
			muscle_group,
			secondary_muscle_group,
			exercise_type,
			is_official,
			created_at
		FROM exercises
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exercises := make([]model.Exercise, 0)
	for rows.Next() {
		var exercise model.Exercise

		err := rows.Scan(
			&exercise.ID,
			&exercise.Name,
			&exercise.Description,
			&exercise.MuscleGroup,
			&exercise.SecondaryMuscleGroup,
			&exercise.ExerciseType,
			&exercise.IsOfficial,
			&exercise.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		exercises = append(exercises, exercise)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return exercises, nil
}
