package service

import (
	"context"
	"errors"
	"testing"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
)

type MockExerciseRepository struct {
	createFunc                        func(ctx context.Context, exercise *model.Exercise) error
	getByIDFunc                       func(ctx context.Context, id string) (*model.Exercise, error)
	nameExistsFunc                    func(ctx context.Context, name string, ownerUserID *string, excludeID string) (bool, error)
	listFunc                          func(ctx context.Context, filters model.ExerciseFilter) ([]model.Exercise, int, error)
	listWorkoutSessionsByExerciseFunc func(ctx context.Context, exerciseID, userID string, limit int) ([]model.ExerciseWorkoutSessionSummary, error)
	getInsightsFunc                   func(ctx context.Context, exerciseID, userID string) (model.ExerciseInsights, error)
	updateExerciseFunc                func(ctx context.Context, exercise *model.Exercise) error
	deleteExerciseFunc                func(ctx context.Context, id string) error
}

func (m *MockExerciseRepository) Create(ctx context.Context, exercise *model.Exercise) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, exercise)
	}
	return nil
}

func (m *MockExerciseRepository) GetByID(ctx context.Context, id string) (*model.Exercise, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockExerciseRepository) NameExists(ctx context.Context, name string, ownerUserID *string, excludeID string) (bool, error) {
	if m.nameExistsFunc != nil {
		return m.nameExistsFunc(ctx, name, ownerUserID, excludeID)
	}
	return false, nil
}

func (m *MockExerciseRepository) List(ctx context.Context, filters model.ExerciseFilter) ([]model.Exercise, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filters)
	}
	return []model.Exercise{}, 0, nil
}

func (m *MockExerciseRepository) ListWorkoutSessionsByExercise(ctx context.Context, exerciseID, userID string, limit int) ([]model.ExerciseWorkoutSessionSummary, error) {
	if m.listWorkoutSessionsByExerciseFunc != nil {
		return m.listWorkoutSessionsByExerciseFunc(ctx, exerciseID, userID, limit)
	}
	return []model.ExerciseWorkoutSessionSummary{}, nil
}

func (m *MockExerciseRepository) GetInsights(ctx context.Context, exerciseID, userID string) (model.ExerciseInsights, error) {
	if m.getInsightsFunc != nil {
		return m.getInsightsFunc(ctx, exerciseID, userID)
	}
	return model.ExerciseInsights{}, nil
}

func (m *MockExerciseRepository) UpdateExercise(ctx context.Context, exercise *model.Exercise) error {
	if m.updateExerciseFunc != nil {
		return m.updateExerciseFunc(ctx, exercise)
	}
	return nil
}

func (m *MockExerciseRepository) DeleteExercise(ctx context.Context, id string) error {
	if m.deleteExerciseFunc != nil {
		return m.deleteExerciseFunc(ctx, id)
	}
	return nil
}

func TestExerciseServiceGetByIDInvalidID(t *testing.T) {
	mockRepo := &MockExerciseRepository{}
	svc := NewExerciseService(mockRepo)

	_, err := svc.GetByID(context.Background(), "")

	if !errors.Is(err, ErrInvalidExerciseInput) {
		t.Errorf("expected ErrInvalidExerciseInput, got %v", err)
	}
}

func TestExerciseServiceGetByIDNotFound(t *testing.T) {
	mockRepo := &MockExerciseRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.Exercise, error) {
			return nil, pgx.ErrNoRows
		},
	}

	svc := NewExerciseService(mockRepo)

	_, err := svc.GetByID(context.Background(), "550e8400-e29b-41d4-a716-446655440000")

	if !errors.Is(err, ErrExerciseNotFound) {
		t.Errorf("expected ErrExerciseNotFound, got %v", err)
	}
}

func TestExerciseServiceGetByIDRepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	mockRepo := &MockExerciseRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.Exercise, error) {
			return nil, expectedErr
		},
	}

	svc := NewExerciseService(mockRepo)

	_, err := svc.GetByID(context.Background(), "550e8400-e29b-41d4-a716-446655440001")

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestExerciseServiceGetByIDSuccess(t *testing.T) {
	mockRepo := &MockExerciseRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.Exercise, error) {
			return &model.Exercise{
				ID:          id,
				Name:        "Bench Press",
				MuscleGroup: "chest",
				IsOfficial:  true,
			}, nil
		},
	}

	svc := NewExerciseService(mockRepo)

	exercise, err := svc.GetByID(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if exercise == nil {
		t.Fatal("expected exercise, got nil")
		return
	}

	if exercise.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected ID 550e8400-e29b-41d4-a716-446655440000, got %s", exercise.ID)
	}

	if exercise.Name != "Bench Press" {
		t.Errorf("expected Name Bench Press, got %s", exercise.Name)
	}
}

func TestExerciseServiceListSuccess(t *testing.T) {
	mockRepo := &MockExerciseRepository{
		listFunc: func(ctx context.Context, filters model.ExerciseFilter) ([]model.Exercise, int, error) {
			return []model.Exercise{
				{
					ID:          "550e8400-e29b-41d4-a716-446655440000",
					Name:        "Bench Press",
					MuscleGroup: "chest",
					IsOfficial:  true,
				},
				{
					ID:          "550e8400-e29b-41d4-a716-446655440001",
					Name:        "Squat",
					MuscleGroup: "legs",
					IsOfficial:  true,
				},
			}, 2, nil
		},
	}

	svc := NewExerciseService(mockRepo)

	result, err := svc.List(context.Background(), model.ExerciseFilter{})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if len(result.Items) != 2 {
		t.Errorf("expected 2 exercises, got %d", len(result.Items))
	}
}

func TestExerciseServiceListRepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	mockRepo := &MockExerciseRepository{
		listFunc: func(ctx context.Context, filters model.ExerciseFilter) ([]model.Exercise, int, error) {
			return nil, 0, expectedErr
		},
	}

	svc := NewExerciseService(mockRepo)

	_, err := svc.List(context.Background(), model.ExerciseFilter{})

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestExerciseServiceDeleteExerciseInvalidID(t *testing.T) {
	mockRepo := &MockExerciseRepository{}
	svc := NewExerciseService(mockRepo)

	err := svc.DeleteExercise(context.Background(), "   ")

	if !errors.Is(err, ErrInvalidExerciseInput) {
		t.Errorf("expected ErrInvalidExerciseInput, got %v", err)
	}
}

func TestExerciseServiceDeleteExerciseNotFound(t *testing.T) {
	mockRepo := &MockExerciseRepository{
		deleteExerciseFunc: func(ctx context.Context, id string) error {
			return pgx.ErrNoRows
		},
	}
	svc := NewExerciseService(mockRepo)

	err := svc.DeleteExercise(context.Background(), "550e8400-e29b-41d4-a716-446655440000")

	if !errors.Is(err, ErrExerciseNotFound) {
		t.Errorf("expected ErrExerciseNotFound, got %v", err)
	}
}

func TestExerciseServiceDeleteExerciseSuccess(t *testing.T) {
	mockRepo := &MockExerciseRepository{
		deleteExerciseFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}
	svc := NewExerciseService(mockRepo)

	err := svc.DeleteExercise(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestExerciseServiceCreateMultipleSecondaryMuscleGroups(t *testing.T) {
	var captured *model.Exercise
	mockRepo := &MockExerciseRepository{
		createFunc: func(ctx context.Context, exercise *model.Exercise) error {
			captured = exercise
			return nil
		},
	}
	svc := NewExerciseService(mockRepo)

	ownerID := "550e8400-e29b-41d4-a716-446655440111"
	exercise := &model.Exercise{
		Name:                 "Pull Up",
		MuscleGroup:          "back",
		SecondaryMuscleGroup: "biceps, shoulders",
		ExerciseType:         "strength",
		IsOfficial:           false,
		OwnerUserID:          &ownerID,
	}

	if err := svc.Create(context.Background(), exercise); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if captured == nil {
		t.Fatal("expected repository Create to be called")
	}

	if captured.SecondaryMuscleGroup != "biceps, shoulders" {
		t.Errorf("expected normalized secondary muscle groups %q, got %q", "biceps, shoulders", captured.SecondaryMuscleGroup)
	}
}

func TestExerciseServiceCreateNameTaken(t *testing.T) {
	mockRepo := &MockExerciseRepository{
		nameExistsFunc: func(ctx context.Context, name string, ownerUserID *string, excludeID string) (bool, error) {
			return true, nil
		},
		createFunc: func(ctx context.Context, exercise *model.Exercise) error {
			t.Fatal("Create should not be called when the name is already taken")
			return nil
		},
	}
	svc := NewExerciseService(mockRepo)

	ownerID := "550e8400-e29b-41d4-a716-446655440111"
	exercise := &model.Exercise{
		Name:         "Pull Up",
		MuscleGroup:  "back",
		ExerciseType: "strength",
		IsOfficial:   false,
		OwnerUserID:  &ownerID,
	}

	if err := svc.Create(context.Background(), exercise); !errors.Is(err, ErrExerciseNameTaken) {
		t.Errorf("expected ErrExerciseNameTaken, got %v", err)
	}
}

func TestExerciseServiceCreateSecondaryEqualsPrimary(t *testing.T) {
	svc := NewExerciseService(&MockExerciseRepository{})

	ownerID := "550e8400-e29b-41d4-a716-446655440111"
	exercise := &model.Exercise{
		Name:                 "Pull Up",
		MuscleGroup:          "back",
		SecondaryMuscleGroup: "biceps, back",
		ExerciseType:         "strength",
		IsOfficial:           false,
		OwnerUserID:          &ownerID,
	}

	if err := svc.Create(context.Background(), exercise); !errors.Is(err, ErrSecondaryEqualsPrimary) {
		t.Errorf("expected ErrSecondaryEqualsPrimary, got %v", err)
	}
}
