package service

import (
	"context"
	"errors"
	"strings"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/jackc/pgx/v5"
)

// ErrInvalidExerciseInput indicates that the provided exercise data is invalid.
var ErrInvalidExerciseInput = errors.New("invalid exercise input")

// ErrExerciseNotFound indicates that the requested exercise does not exist.
var ErrExerciseNotFound = errors.New("exercise not found")

// ExerciseService provides business logic for exercises.
type ExerciseService struct {
	repo repository.ExerciseRepository
}

// NewExerciseService creates a new ExerciseService.
func NewExerciseService(repo repository.ExerciseRepository) *ExerciseService {
	return &ExerciseService{
		repo: repo,
	}
}

// GetByID retrieves an exercise by its UUID string ID.
func (s *ExerciseService) GetByID(ctx context.Context, id string) (*model.Exercise, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidExerciseInput
	}

	exercise, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrExerciseNotFound
	}
	if err != nil {
		return nil, err
	}

	return exercise, nil
}

// List returns all exercises.
func (s *ExerciseService) List(ctx context.Context) ([]model.Exercise, error) {
	return s.repo.List(ctx)
}

// Create creates a new exercise after validating the input data.
func (s *ExerciseService) Create(ctx context.Context, exercise *model.Exercise) error {
	if exercise == nil {
		return ErrInvalidExerciseInput
	}

	exercise.Name = strings.TrimSpace(exercise.Name)
	exercise.Description = strings.TrimSpace(exercise.Description)
	exercise.MuscleGroup = strings.TrimSpace(exercise.MuscleGroup)
	exercise.SecondaryMuscleGroup = strings.TrimSpace(exercise.SecondaryMuscleGroup)
	exercise.ExerciseType = strings.TrimSpace(exercise.ExerciseType)

	if strings.TrimSpace(exercise.Name) == "" {
		return ErrInvalidExerciseInput
	}

	return s.repo.Create(ctx, exercise)
}

func (s *ExerciseService) UpdateExercise(ctx context.Context, exercise *model.Exercise) error {
	if exercise == nil {
		return ErrInvalidExerciseInput
	}

	exercise.ID = strings.TrimSpace(exercise.ID)
	exercise.Name = strings.TrimSpace(exercise.Name)
	exercise.Description = strings.TrimSpace(exercise.Description)
	exercise.MuscleGroup = strings.TrimSpace(exercise.MuscleGroup)
	exercise.SecondaryMuscleGroup = strings.TrimSpace(exercise.SecondaryMuscleGroup)
	exercise.ExerciseType = strings.TrimSpace(exercise.ExerciseType)

	if exercise.ID == "" || strings.TrimSpace(exercise.Name) == "" {
		return ErrInvalidExerciseInput
	}

	err := s.repo.UpdateExercise(ctx, exercise)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrExerciseNotFound
	}

	return err
}
