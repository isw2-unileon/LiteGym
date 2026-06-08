package service

import (
	"context"
	"errors"
	"testing"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
)

const manualTestUserID = "550e8400-e29b-41d4-a716-446655440111"
const manualTestRoutineID = "550e8400-e29b-41d4-a716-446655440000"
const manualTestExerciseID = "550e8400-e29b-41d4-a716-446655440222"

type manualServiceTestRepository struct {
	countFunc     func(ctx context.Context, userID string) (int, error)
	createFunc    func(ctx context.Context, routine model.ManualRoutineToSave) (string, error)
	updateFunc    func(ctx context.Context, routineID, userID string, routine model.ManualRoutineToSave) error
	deleteFunc    func(ctx context.Context, routineID, userID string) error
	duplicateFunc func(ctx context.Context, routineID, userID string) (string, error)
}

func (r *manualServiceTestRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	if r.countFunc != nil {
		return r.countFunc(ctx, userID)
	}
	return 0, nil
}

func (r *manualServiceTestRepository) CreateManualRoutine(ctx context.Context, routine model.ManualRoutineToSave) (string, error) {
	if r.createFunc != nil {
		return r.createFunc(ctx, routine)
	}
	return "new-routine-id", nil
}

func (r *manualServiceTestRepository) UpdateManualRoutine(ctx context.Context, routineID, userID string, routine model.ManualRoutineToSave) error {
	if r.updateFunc != nil {
		return r.updateFunc(ctx, routineID, userID, routine)
	}
	return nil
}

func (r *manualServiceTestRepository) DeleteRoutine(ctx context.Context, routineID, userID string) error {
	if r.deleteFunc != nil {
		return r.deleteFunc(ctx, routineID, userID)
	}
	return nil
}

func (r *manualServiceTestRepository) DuplicateRoutine(ctx context.Context, routineID, userID string) (string, error) {
	if r.duplicateFunc != nil {
		return r.duplicateFunc(ctx, routineID, userID)
	}
	return "duplicated-routine-id", nil
}

func TestManualRoutineServiceCreate(t *testing.T) {
	var captured model.ManualRoutineToSave
	svc := NewManualRoutineService(&manualServiceTestRepository{
		createFunc: func(ctx context.Context, routine model.ManualRoutineToSave) (string, error) {
			captured = routine
			return "routine-123", nil
		},
	})

	weight := 72.5
	id, err := svc.Create(context.Background(), manualTestUserID, ManualRoutineInput{
		Name:        "  Push Day  ",
		Description: "  desc  ",
		RoutineType: "Fuerza",
		Exercises: []ManualRoutineExerciseInput{
			{
				ExerciseID: manualTestExerciseID,
				Sets: []ManualRoutineSetInput{
					{SetNumber: 1, TargetWeightKg: &weight},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "routine-123" {
		t.Fatalf("expected routine-123, got %q", id)
	}
	if captured.Name != "Push Day" {
		t.Fatalf("expected trimmed name 'Push Day', got %q", captured.Name)
	}
	if captured.Description != "desc" {
		t.Fatalf("expected trimmed description 'desc', got %q", captured.Description)
	}
	if captured.RoutineType != "Fuerza" {
		t.Fatalf("expected RoutineType 'Fuerza', got %q", captured.RoutineType)
	}
	if len(captured.Exercises) != 1 {
		t.Fatalf("expected 1 exercise, got %d", len(captured.Exercises))
	}
	if captured.Exercises[0].Order != 1 {
		t.Fatalf("expected order 1, got %d", captured.Exercises[0].Order)
	}
	if len(captured.Exercises[0].Sets) != 1 {
		t.Fatalf("expected 1 set, got %d", len(captured.Exercises[0].Sets))
	}
}

func TestManualRoutineServiceCreateNormalizesUnknownType(t *testing.T) {
	var captured model.ManualRoutineToSave
	svc := NewManualRoutineService(&manualServiceTestRepository{
		createFunc: func(ctx context.Context, routine model.ManualRoutineToSave) (string, error) {
			captured = routine
			return "id", nil
		},
	})

	if _, err := svc.Create(context.Background(), manualTestUserID, ManualRoutineInput{
		Name:        "Routine",
		RoutineType: "Bogus",
	}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if captured.RoutineType != defaultRoutineType {
		t.Fatalf("expected %q, got %q", defaultRoutineType, captured.RoutineType)
	}
}

func TestManualRoutineServiceCreateNameRequired(t *testing.T) {
	svc := NewManualRoutineService(&manualServiceTestRepository{})

	_, err := svc.Create(context.Background(), manualTestUserID, ManualRoutineInput{Name: "   "})
	if !errors.Is(err, ErrRoutineNameRequired) {
		t.Fatalf("expected ErrRoutineNameRequired, got %v", err)
	}
}

func TestManualRoutineServiceCreateInvalidExerciseID(t *testing.T) {
	svc := NewManualRoutineService(&manualServiceTestRepository{})

	_, err := svc.Create(context.Background(), manualTestUserID, ManualRoutineInput{
		Name:      "Routine",
		Exercises: []ManualRoutineExerciseInput{{ExerciseID: "not-a-uuid"}},
	})
	if !errors.Is(err, ErrInvalidRoutineInput) {
		t.Fatalf("expected ErrInvalidRoutineInput, got %v", err)
	}
}

func TestManualRoutineServiceCreateEmptyUser(t *testing.T) {
	svc := NewManualRoutineService(&manualServiceTestRepository{})

	_, err := svc.Create(context.Background(), "  ", ManualRoutineInput{Name: "Routine"})
	if !errors.Is(err, ErrInvalidRoutineInput) {
		t.Fatalf("expected ErrInvalidRoutineInput, got %v", err)
	}
}

func TestManualRoutineServiceCreateLimitReached(t *testing.T) {
	svc := NewManualRoutineService(&manualServiceTestRepository{
		countFunc: func(ctx context.Context, userID string) (int, error) {
			return maxRoutinesPerUser, nil
		},
		createFunc: func(ctx context.Context, routine model.ManualRoutineToSave) (string, error) {
			t.Fatalf("create should not be called when limit reached")
			return "", nil
		},
	})

	_, err := svc.Create(context.Background(), manualTestUserID, ManualRoutineInput{Name: "Routine"})
	if !errors.Is(err, ErrRoutineLimitReached) {
		t.Fatalf("expected ErrRoutineLimitReached, got %v", err)
	}
}

func TestManualRoutineServiceUpdate(t *testing.T) {
	called := false
	svc := NewManualRoutineService(&manualServiceTestRepository{
		updateFunc: func(ctx context.Context, routineID, userID string, routine model.ManualRoutineToSave) error {
			called = true
			if routineID != manualTestRoutineID {
				t.Fatalf("unexpected routineID %q", routineID)
			}
			return nil
		},
	})

	if err := svc.Update(context.Background(), manualTestUserID, manualTestRoutineID, ManualRoutineInput{Name: "Routine"}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected repository update to be called")
	}
}

func TestManualRoutineServiceUpdateNotFound(t *testing.T) {
	svc := NewManualRoutineService(&manualServiceTestRepository{
		updateFunc: func(ctx context.Context, routineID, userID string, routine model.ManualRoutineToSave) error {
			return pgx.ErrNoRows
		},
	})

	err := svc.Update(context.Background(), manualTestUserID, manualTestRoutineID, ManualRoutineInput{Name: "Routine"})
	if !errors.Is(err, ErrRoutineNotFound) {
		t.Fatalf("expected ErrRoutineNotFound, got %v", err)
	}
}

func TestManualRoutineServiceUpdateInvalidInput(t *testing.T) {
	svc := NewManualRoutineService(&manualServiceTestRepository{})

	err := svc.Update(context.Background(), manualTestUserID, "  ", ManualRoutineInput{Name: "Routine"})
	if !errors.Is(err, ErrInvalidRoutineInput) {
		t.Fatalf("expected ErrInvalidRoutineInput, got %v", err)
	}
}

func TestManualRoutineServiceDelete(t *testing.T) {
	svc := NewManualRoutineService(&manualServiceTestRepository{})

	if err := svc.Delete(context.Background(), manualTestUserID, manualTestRoutineID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestManualRoutineServiceDeleteNotFound(t *testing.T) {
	svc := NewManualRoutineService(&manualServiceTestRepository{
		deleteFunc: func(ctx context.Context, routineID, userID string) error {
			return pgx.ErrNoRows
		},
	})

	err := svc.Delete(context.Background(), manualTestUserID, manualTestRoutineID)
	if !errors.Is(err, ErrRoutineNotFound) {
		t.Fatalf("expected ErrRoutineNotFound, got %v", err)
	}
}

func TestManualRoutineServiceDuplicate(t *testing.T) {
	svc := NewManualRoutineService(&manualServiceTestRepository{
		duplicateFunc: func(ctx context.Context, routineID, userID string) (string, error) {
			return "copy-1", nil
		},
	})

	id, err := svc.Duplicate(context.Background(), manualTestUserID, manualTestRoutineID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "copy-1" {
		t.Fatalf("expected copy-1, got %q", id)
	}
}

func TestManualRoutineServiceDuplicateLimitReached(t *testing.T) {
	svc := NewManualRoutineService(&manualServiceTestRepository{
		countFunc: func(ctx context.Context, userID string) (int, error) {
			return maxRoutinesPerUser, nil
		},
		duplicateFunc: func(ctx context.Context, routineID, userID string) (string, error) {
			t.Fatalf("duplicate should not be called when limit reached")
			return "", nil
		},
	})

	_, err := svc.Duplicate(context.Background(), manualTestUserID, manualTestRoutineID)
	if !errors.Is(err, ErrRoutineLimitReached) {
		t.Fatalf("expected ErrRoutineLimitReached, got %v", err)
	}
}

func TestManualRoutineServiceDuplicateNotFound(t *testing.T) {
	svc := NewManualRoutineService(&manualServiceTestRepository{
		duplicateFunc: func(ctx context.Context, routineID, userID string) (string, error) {
			return "", pgx.ErrNoRows
		},
	})

	_, err := svc.Duplicate(context.Background(), manualTestUserID, manualTestRoutineID)
	if !errors.Is(err, ErrRoutineNotFound) {
		t.Fatalf("expected ErrRoutineNotFound, got %v", err)
	}
}
