package repository

import (
	"context"
	"strings"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ExerciseRepository defines the persistence operations for exercises.
type ExerciseRepository interface {
	GetByID(ctx context.Context, id string) (*model.Exercise, error)
	List(ctx context.Context) ([]model.Exercise, error)
	Create(ctx context.Context, exercise *model.Exercise) error
	UpdateExercise(ctx context.Context, exercise *model.Exercise) error
	DeleteExercise(ctx context.Context, id string) error
}

type exerciseRepository struct {
	db *pgxpool.Pool
}

// NewExerciseRepository creates a new ExerciseRepository backed by PostgreSQL.
func NewExerciseRepository(db *pgxpool.Pool) ExerciseRepository {
	return &exerciseRepository{
		db: db,
	}
}

func (r *exerciseRepository) GetByID(ctx context.Context, id string) (*model.Exercise, error) {
	query := `
		SELECT
			e.id::text,
			e.name,
			e.description,
			e.muscle_group,
			COALESCE(string_agg(esmg.muscle_group, ', ' ORDER BY esmg.muscle_group), ''),
			e.exercise_type,
			e.is_official,
			e.created_at
		FROM exercises e
		LEFT JOIN exercise_secondary_muscle_groups esmg ON esmg.exercise_id = e.id
		WHERE e.id = $1::uuid AND e.deleted_at IS NULL
		GROUP BY e.id, e.name, e.description, e.muscle_group, e.exercise_type, e.is_official, e.created_at
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
			e.id::text,
			e.name,
			e.description,
			e.muscle_group,
			COALESCE(string_agg(esmg.muscle_group, ', ' ORDER BY esmg.muscle_group), ''),
			e.exercise_type,
			e.is_official,
			e.created_at
		FROM exercises e
		LEFT JOIN exercise_secondary_muscle_groups esmg ON esmg.exercise_id = e.id
		WHERE e.deleted_at IS NULL
		GROUP BY e.id, e.name, e.description, e.muscle_group, e.exercise_type, e.is_official, e.created_at
		ORDER BY e.created_at ASC, e.id::text ASC
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

func (r *exerciseRepository) Create(ctx context.Context, exercise *model.Exercise) error {
	query := `
		INSERT INTO exercises(
			name,
			description,
			muscle_group,
			exercise_type,
			is_official
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, created_at
	`

	if err := r.db.QueryRow(
		ctx,
		query,
		exercise.Name,
		exercise.Description,
		exercise.MuscleGroup,
		exercise.ExerciseType,
		exercise.IsOfficial,
	).Scan(&exercise.ID, &exercise.CreatedAt); err != nil {
		return err
	}

	if exercise.SecondaryMuscleGroup == "" {
		return nil
	}

	secondaryMuscleGroups := strings.Split(exercise.SecondaryMuscleGroup, ",")
	normalized := make([]string, 0, len(secondaryMuscleGroups))

	for _, group := range secondaryMuscleGroups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}

		_, err := r.db.Exec(
			ctx,
			`
				INSERT INTO exercise_secondary_muscle_groups (exercise_id, muscle_group)
				VALUES ($1::uuid, $2)
				ON CONFLICT (exercise_id, muscle_group) DO NOTHING
			`,
			exercise.ID,
			group,
		)
		if err != nil {
			return err
		}

		normalized = append(normalized, group)
	}

	exercise.SecondaryMuscleGroup = strings.Join(normalized, ", ")
	return nil

}

// UpdateExercise updates the contents of an existing exercise.
func (r *exerciseRepository) UpdateExercise(ctx context.Context, exercise *model.Exercise) error {
	// 1. Define the query that updates the exercise core fields.
	query := `
        UPDATE exercises
        SET
            name = $1,
            description = $2,
            muscle_group = $3,
            exercise_type = $4,
            is_official = $5
        WHERE id = $6::uuid AND deleted_at IS NULL
    `

	// 2. Ejecutamos la actualización
	result, err := r.db.Exec(
		ctx,
		query,
		exercise.Name,
		exercise.Description,
		exercise.MuscleGroup,
		exercise.ExerciseType,
		exercise.IsOfficial,
		exercise.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	// 3. Manejo de Músculos Secundarios

	// Borrar actuales
	_, err = r.db.Exec(ctx, "DELETE FROM exercise_secondary_muscle_groups WHERE exercise_id = $1::uuid", exercise.ID)
	if err != nil {
		return err
	}

	// Insertar los nuevos
	if exercise.SecondaryMuscleGroup != "" {
		groups := strings.Split(exercise.SecondaryMuscleGroup, ",")
		for _, g := range groups {
			g = strings.TrimSpace(g)
			if g == "" {
				continue
			}
			_, err = r.db.Exec(ctx,
				"INSERT INTO exercise_secondary_muscle_groups (exercise_id, muscle_group) VALUES ($1::uuid, $2)",
				exercise.ID, g)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteExercise performs a soft-delete of an exercise.
func (r *exerciseRepository) DeleteExercise(ctx context.Context, id string) error {
	query := `
		UPDATE exercises
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1::uuid AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
