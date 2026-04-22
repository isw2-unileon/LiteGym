package service

import (
	"context"
	"errors"
	"testing"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
)

type MockExerciseRepository struct {
	createFunc  func(ctx context.Context, exercise *model.Exercise) error
	getByIDFunc func(ctx context.Context, id string) (*model.Exercise, error)
	listFunc    func(ctx context.Context) ([]model.Exercise, error)
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

func (m *MockExerciseRepository) List(ctx context.Context) ([]model.Exercise, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return []model.Exercise{}, nil
}

func TestExerciseServiceCreateNilExercise(t *testing.T) {
	mockRepo := &MockExerciseRepository{}
	svc := NewExerciseService(mockRepo)

	err := svc.Create(context.Background(), nil)

	if !errors.Is(err, ErrInvalidExerciseInput) {
		t.Errorf("expected ErrInvalidExerciseInput, got %v", err)
	}
}

func TestExerciseServiceCreateEmptyName(t *testing.T) {
	mockRepo := &MockExerciseRepository{}
	svc := NewExerciseService(mockRepo)

	exercise := &model.Exercise{
		Name:        "   ",
		MuscleGroup: "chest",
	}

	err := svc.Create(context.Background(), exercise)

	if !errors.Is(err, ErrInvalidExerciseInput) {
		t.Errorf("expected ErrInvalidExerciseInput, got %v", err)
	}
}

func TestExerciseServiceCreateEmptyMuscleGroup(t *testing.T) {
	mockRepo := &MockExerciseRepository{}
	svc := NewExerciseService(mockRepo)

	exercise := &model.Exercise{
		Name:        "Bench Press",
		MuscleGroup: "   ",
	}

	err := svc.Create(context.Background(), exercise)

	if !errors.Is(err, ErrInvalidExerciseInput) {
		t.Errorf("expected ErrInvalidExerciseInput, got %v", err)
	}
}

func TestExerciseServiceCreateTrimsFields(t *testing.T) {
	mockRepo := &MockExerciseRepository{
		createFunc: func(ctx context.Context, exercise *model.Exercise) error {
			if exercise.Name != "Bench Press" {
				t.Errorf("expected trimmed Name, got %q", exercise.Name)
			}
			if exercise.Description != "Flat bench press" {
				t.Errorf("expected trimmed Description, got %q", exercise.Description)
			}
			if exercise.MuscleGroup != "chest" {
				t.Errorf("expected trimmed MuscleGroup, got %q", exercise.MuscleGroup)
			}
			if exercise.SecondaryMuscleGroup != "triceps" {
				t.Errorf("expected trimmed SecondaryMuscleGroup, got %q", exercise.SecondaryMuscleGroup)
			}
			if exercise.ExerciseType != "strength" {
				t.Errorf("expected trimmed ExerciseType, got %q", exercise.ExerciseType)
			}
			return nil
		},
	}

	svc := NewExerciseService(mockRepo)

	exercise := &model.Exercise{
		Name:                 "  Bench Press  ",
		Description:          "  Flat bench press  ",
		MuscleGroup:          "  chest  ",
		SecondaryMuscleGroup: "  triceps  ",
		ExerciseType:         "  strength  ",
		IsOfficial:           true,
	}

	err := svc.Create(context.Background(), exercise)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestExerciseServiceCreateRepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	mockRepo := &MockExerciseRepository{
		createFunc: func(ctx context.Context, exercise *model.Exercise) error {
			return expectedErr
		},
	}

	svc := NewExerciseService(mockRepo)

	exercise := &model.Exercise{
		Name:        "Bench Press",
		MuscleGroup: "chest",
	}

	err := svc.Create(context.Background(), exercise)

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
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
		listFunc: func(ctx context.Context) ([]model.Exercise, error) {
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
			}, nil
		},
	}

	svc := NewExerciseService(mockRepo)

	exercises, err := svc.List(context.Background())
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if len(exercises) != 2 {
		t.Errorf("expected 2 exercises, got %d", len(exercises))
	}
}

func TestExerciseServiceListRepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	mockRepo := &MockExerciseRepository{
		listFunc: func(ctx context.Context) ([]model.Exercise, error) {
			return nil, expectedErr
		},
	}

	svc := NewExerciseService(mockRepo)

	_, err := svc.List(context.Background())

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}
