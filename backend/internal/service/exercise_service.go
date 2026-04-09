package service

import (
	"context"
	"errors"
	"strings"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidExerciseInput = errors.New("invalid exercise input")
	ErrExerciseNotFound     = errors.New("exercise not found")
)

type ExerciseService struct {
	repo repository.ExerciseRepository
}

func NewExerciseService(repo repository.ExerciseRepository) *ExerciseService {
	return &ExerciseService{
		repo: repo,
	}
}

func (s *ExerciseService) Create(ctx context.Context, exercise *model.Exercise) error {
	if exercise == nil {
		return ErrInvalidExerciseInput
	}

	exercise.Name = strings.TrimSpace(exercise.Name)
	exercise.Description = strings.TrimSpace(exercise.Description)
	exercise.MuscleGroup = strings.TrimSpace(exercise.MuscleGroup)
	exercise.SecondaryMuscleGroup = strings.TrimSpace(exercise.SecondaryMuscleGroup)
	exercise.ExerciseType = strings.TrimSpace(exercise.ExerciseType)

	if exercise.Name == "" || exercise.MuscleGroup == "" {
		return ErrInvalidExerciseInput
	}

	return s.repo.Create(ctx, exercise)
}

func (s *ExerciseService) GetByID(ctx context.Context, id int64) (*model.Exercise, error) {
	if id <= 0 {
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

func (s *ExerciseService) List(ctx context.Context) ([]model.Exercise, error) {
	return s.repo.List(ctx)
}
