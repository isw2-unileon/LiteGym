package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
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

// ListWorkoutSessionsByExercise returns recent workout sessions that include the selected exercise.
func (s *ExerciseService) ListWorkoutSessionsByExercise(ctx context.Context, exerciseID, userID string, limit int) ([]model.ExerciseWorkoutSessionSummary, error) {
	exerciseID = strings.TrimSpace(exerciseID)
	userID = strings.TrimSpace(userID)

	if exerciseID == "" || userID == "" {
		return nil, ErrInvalidExerciseInput
	}

	if _, err := uuid.Parse(exerciseID); err != nil {
		return nil, ErrInvalidExerciseInput
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, ErrInvalidUserInput
	}

	if limit <= 0 {
		limit = 5
	}

	if limit > 20 {
		limit = 20
	}

	return s.repo.ListWorkoutSessionsByExercise(ctx, exerciseID, userID, limit)
}

// GetInsights returns historical performance analytics for a user's exercise.
func (s *ExerciseService) GetInsights(ctx context.Context, exerciseID, userID string) (model.ExerciseInsights, error) {
	exerciseID = strings.TrimSpace(exerciseID)
	userID = strings.TrimSpace(userID)

	if exerciseID == "" {
		return model.ExerciseInsights{}, ErrInvalidExerciseInput
	}

	if userID == "" {
		return model.ExerciseInsights{}, ErrInvalidUserInput
	}

	if _, err := uuid.Parse(exerciseID); err != nil {
		return model.ExerciseInsights{}, ErrInvalidExerciseInput
	}

	if _, err := uuid.Parse(userID); err != nil {
		return model.ExerciseInsights{}, ErrInvalidUserInput
	}

	return s.repo.GetInsights(ctx, exerciseID, userID)
}

// Create creates a new exercise after validating the input data.
func (s *ExerciseService) Create(ctx context.Context, exercise *model.Exercise) error {
	if exercise == nil {
		return ErrInvalidExerciseInput
	}

	if err := validateExerciseOwnership(exercise); err != nil {
		return err
	}

	if err := s.normalizeAndValidateExercise(exercise); err != nil {
		return err
	}

	exists, err := s.repo.NameExists(ctx, exercise.Name, exercise.OwnerUserID, "")
	if err != nil {
		return err
	}
	if exists {
		return ErrExerciseNameTaken
	}

	if err := s.repo.Create(ctx, exercise); err != nil {
		if isUniqueViolation(err) {
			return ErrExerciseNameTaken
		}
		return err
	}

	return nil
}

// validateExerciseOwnership enforces the official/owner invariant: official
// exercises have no owner, custom exercises must belong to a valid user.
func validateExerciseOwnership(exercise *model.Exercise) error {
	if exercise.IsOfficial {
		exercise.OwnerUserID = nil
		return nil
	}

	if exercise.OwnerUserID == nil || strings.TrimSpace(*exercise.OwnerUserID) == "" {
		return ErrInvalidExerciseInput
	}
	if _, err := uuid.Parse(strings.TrimSpace(*exercise.OwnerUserID)); err != nil {
		return ErrInvalidExerciseInput
	}

	return nil
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

	if err := validateExerciseOwnership(exercise); err != nil {
		return err
	}

	if err := s.normalizeAndValidateExercise(exercise); err != nil {
		return err
	}

	exists, err := s.repo.NameExists(ctx, exercise.Name, exercise.OwnerUserID, exercise.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrExerciseNameTaken
	}

	err = s.repo.UpdateExercise(ctx, exercise)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrExerciseNotFound
	}
	if isUniqueViolation(err) {
		return ErrExerciseNameTaken
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

	normalizedSecondary, err := normalizeSecondaryMuscleGroups(
		exercise.SecondaryMuscleGroup,
		exercise.MuscleGroup,
	)
	if err != nil {
		return err
	}
	exercise.SecondaryMuscleGroup = normalizedSecondary

	if exercise.ExerciseType == "" {
		return ErrExerciseTypeRequired
	}

	if !isValidExerciseType(exercise.ExerciseType) {
		return ErrInvalidExerciseType
	}

	return nil
}
