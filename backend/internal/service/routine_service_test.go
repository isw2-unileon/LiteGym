package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
)

type routineServiceTestRepository struct {
	getByIDFunc func(ctx context.Context, userID, routineID string) (*model.Routine, error)
}

func (r *routineServiceTestRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

func (r *routineServiceTestRepository) ListByUser(ctx context.Context, userID string) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

func (r *routineServiceTestRepository) GetByID(ctx context.Context, userID, routineID string) (*model.Routine, error) {
	if r.getByIDFunc != nil {
		return r.getByIDFunc(ctx, userID, routineID)
	}
	return nil, nil
}

func (r *routineServiceTestRepository) SaveGeneratedAIRoutine(ctx context.Context, routine model.AIRoutineToSave) (string, error) {
	return "", nil
}

func (r *routineServiceTestRepository) CountAIGenerationsInWindow(ctx context.Context, userID string, since time.Time) (int, error) {
	return 0, nil
}

func (r *routineServiceTestRepository) LogAIGeneration(ctx context.Context, userID string, createdAt time.Time) error {
	return nil
}

func (r *routineServiceTestRepository) ListAvailableExercisesForAI(ctx context.Context, userID string, targetMuscleGroups []string, limit int) ([]model.Exercise, error) {
	return []model.Exercise{}, nil
}

func TestRoutineServiceGetByID(t *testing.T) {
	svc := NewRoutineService(&routineServiceTestRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.Routine, error) {
			return &model.Routine{ID: routineID, UserID: userID, Name: "Push Day"}, nil
		},
	})

	routine, err := svc.GetByID(context.Background(), "user-1", "routine-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if routine == nil || routine.Name != "Push Day" {
		t.Fatalf("expected routine Push Day, got %#v", routine)
	}
}

func TestRoutineServiceGetByIDInvalidInput(t *testing.T) {
	svc := NewRoutineService(&routineServiceTestRepository{})

	_, err := svc.GetByID(context.Background(), "", "routine-1")
	if !errors.Is(err, ErrInvalidRoutineInput) {
		t.Fatalf("expected ErrInvalidRoutineInput, got %v", err)
	}
}

func TestRoutineServiceGetByIDNotFound(t *testing.T) {
	svc := NewRoutineService(&routineServiceTestRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.Routine, error) {
			return nil, pgx.ErrNoRows
		},
	})

	_, err := svc.GetByID(context.Background(), "user-1", "routine-1")
	if !errors.Is(err, ErrRoutineNotFound) {
		t.Fatalf("expected ErrRoutineNotFound, got %v", err)
	}
}
