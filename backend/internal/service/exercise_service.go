package service

import (
	"context"
	"errors"
	"strings"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/jackc/pgx/v5"
)

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
func (s *ExerciseService) List(ctx context.Context, filters model.ExerciseFilter) (model.ExerciseListResponse, error) {
	filters.Search = strings.TrimSpace(filters.Search)
	filters.Type = normalizeDomainValue(filters.Type)
	filters.MuscleGroup = normalizeDomainValue(filters.MuscleGroup)

	if filters.Page <= 0 {
		filters.Page = 1
	}

	if filters.Limit <= 0 {
		filters.Limit = 20
	}

	if filters.Limit > 100 {
		filters.Limit = 100
	}

	exercises, total, err := s.repo.List(ctx, filters)
	if err != nil {
		return model.ExerciseListResponse{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + filters.Limit - 1) / filters.Limit
	}

	return model.ExerciseListResponse{
		Items:      exercises,
		Page:       filters.Page,
		Limit:      filters.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// Create creates a new exercise after validating the input data.
func (s *ExerciseService) Create(ctx context.Context, exercise *model.Exercise) error {
	if exercise == nil {
		return ErrInvalidExerciseInput
	}

	if err := s.normalizeAndValidateExercise(exercise); err != nil {
		return err
	}

	return s.repo.Create(ctx, exercise)
}

// GetMetadata returns the valid exercise domain options exposed to clients.
func (s *ExerciseService) GetMetadata() model.ExerciseMetadataResponse {
	return exerciseMetadata()
}

// UpdateExercise validates and updates an existing exercise.
func (s *ExerciseService) UpdateExercise(ctx context.Context, exercise *model.Exercise) error {
	if exercise == nil {
		return ErrInvalidExerciseInput
	}

	exercise.ID = strings.TrimSpace(exercise.ID)
	if exercise.ID == "" {
		return ErrInvalidExerciseInput
	}

	if err := s.normalizeAndValidateExercise(exercise); err != nil {
		return err
	}

	err := s.repo.UpdateExercise(ctx, exercise)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrExerciseNotFound
	}

	return err
}

// DeleteExercise validates and soft-deletes an existing exercise by ID.
func (s *ExerciseService) DeleteExercise(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidExerciseInput
	}

	err := s.repo.DeleteExercise(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrExerciseNotFound
	}

	return err
}

func (s *ExerciseService) normalizeAndValidateExercise(exercise *model.Exercise) error {
	exercise.Name = normalizeName(exercise.Name)
	exercise.Description = strings.TrimSpace(exercise.Description)
	exercise.MuscleGroup = normalizeDomainValue(exercise.MuscleGroup)
	exercise.SecondaryMuscleGroup = normalizeDomainValue(exercise.SecondaryMuscleGroup)
	exercise.ExerciseType = normalizeDomainValue(exercise.ExerciseType)

	if exercise.Name == "" {
		return ErrExerciseNameRequired
	}

	if len([]rune(exercise.Name)) > maxExerciseNameLength {
		return ErrExerciseNameTooLong
	}

	if len([]rune(exercise.Description)) > maxExerciseDescriptionLength {
		return ErrDescriptionTooLong
	}

	if exercise.MuscleGroup == "" {
		return ErrMuscleGroupRequired
	}

	if !isValidMuscleGroup(exercise.MuscleGroup) {
		return ErrInvalidMuscleGroup
	}

	if exercise.SecondaryMuscleGroup != "" {
		if !isValidMuscleGroup(exercise.SecondaryMuscleGroup) {
			return ErrInvalidMuscleGroup
		}
		if exercise.SecondaryMuscleGroup == exercise.MuscleGroup {
			return ErrSecondaryEqualsPrimary
		}
	}

	if exercise.ExerciseType == "" {
		return ErrExerciseTypeRequired
	}

	if !isValidExerciseType(exercise.ExerciseType) {
		return ErrInvalidExerciseType
	}

	return nil
}
