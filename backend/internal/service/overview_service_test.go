package service

import (
	"context"
	"testing"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
)

type mockRoutineRepository struct {
	listRecentByUserFunc func(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error)
}

func (m *mockRoutineRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error) {
	if m.listRecentByUserFunc != nil {
		return m.listRecentByUserFunc(ctx, userID, limit)
	}
	return []model.OverviewRoutineSummary{}, nil
}

func (m *mockRoutineRepository) CountAIGenerationsInWindow(ctx context.Context, userID string, since time.Time) (int, error) {
	return 0, nil
}

func (m *mockRoutineRepository) LogAIGeneration(ctx context.Context, userID string, createdAt time.Time) error {
	return nil
}

func (m *mockRoutineRepository) ListAvailableExercisesForAI(
	ctx context.Context,
	userID string,
	targetMuscleGroups []string,
	limit int,
) ([]model.Exercise, error) {
	return []model.Exercise{}, nil
}

type mockWorkoutSessionRepository struct {
	listRecentByUserFunc             func(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error)
	listTrainingDatesInRangeFunc     func(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error)
	listMuscleDistributionByUserFunc func(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error)
}

func (m *mockWorkoutSessionRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error) {
	if m.listRecentByUserFunc != nil {
		return m.listRecentByUserFunc(ctx, userID, limit)
	}
	return []model.OverviewWorkoutSummary{}, nil
}

func (m *mockWorkoutSessionRepository) ListTrainingDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
	if m.listTrainingDatesInRangeFunc != nil {
		return m.listTrainingDatesInRangeFunc(ctx, userID, from, to)
	}
	return []time.Time{}, nil
}

func (m *mockWorkoutSessionRepository) ListMuscleDistributionByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error) {
	if m.listMuscleDistributionByUserFunc != nil {
		return m.listMuscleDistributionByUserFunc(ctx, userID, from, to)
	}
	return []model.OverviewMuscleGroupShare{}, 0, nil
}

type mockBodyMetricRepository struct {
	listRecentByUserFunc func(ctx context.Context, userID string, limit int) ([]model.OverviewBodyMetricEntry, error)
}

func (m *mockBodyMetricRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewBodyMetricEntry, error) {
	if m.listRecentByUserFunc != nil {
		return m.listRecentByUserFunc(ctx, userID, limit)
	}
	return []model.OverviewBodyMetricEntry{}, nil
}

func newOverviewServiceTestSubject(referenceDate time.Time) (*OverviewService, float64) {
	routineRepo := &mockRoutineRepository{
		listRecentByUserFunc: func(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error) {
			return []model.OverviewRoutineSummary{{ID: "routine-1", Name: "Push Pull Legs", ExerciseCount: 3}}, nil
		},
	}
	workoutRepo := &mockWorkoutSessionRepository{
		listRecentByUserFunc: func(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error) {
			return []model.OverviewWorkoutSummary{{ID: "workout-1", Name: "Push Day", DurationMinutes: 70, ExerciseCount: 3}}, nil
		},
		listTrainingDatesInRangeFunc: func(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
			return []time.Time{
				time.Date(2026, time.April, 28, 0, 0, 0, 0, time.UTC),
				time.Date(2026, time.April, 27, 0, 0, 0, 0, time.UTC),
				time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
			}, nil
		},
		listMuscleDistributionByUserFunc: func(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error) {
			return overviewMuscleDistributionFixture(from)
		},
	}

	currentWeight := 78.5
	previousWeight := 79.2
	currentBodyFat := 17.9
	previousBodyFat := 18.5
	currentMuscleMass := 36.4
	previousMuscleMass := 36.0
	bodyMetricRepo := &mockBodyMetricRepository{
		listRecentByUserFunc: func(ctx context.Context, userID string, limit int) ([]model.OverviewBodyMetricEntry, error) {
			return []model.OverviewBodyMetricEntry{
				{
					RecordedAt:        referenceDate.AddDate(0, 0, -3),
					WeightKg:          &currentWeight,
					BodyFatPercentage: &currentBodyFat,
					MuscleMassKg:      &currentMuscleMass,
				},
				{
					RecordedAt:        referenceDate.AddDate(0, 0, -10),
					WeightKg:          &previousWeight,
					BodyFatPercentage: &previousBodyFat,
					MuscleMassKg:      &previousMuscleMass,
				},
			}, nil
		},
	}

	return NewOverviewService(routineRepo, workoutRepo, bodyMetricRepo), currentWeight
}

func overviewMuscleDistributionFixture(from time.Time) ([]model.OverviewMuscleGroupShare, int, error) {
	if from.Month() == time.April && from.Day() == 1 {
		return []model.OverviewMuscleGroupShare{
			{Name: "back", Count: 2, Percentage: 67},
			{Name: "chest", Count: 1, Percentage: 33},
		}, 3, nil
	}

	return []model.OverviewMuscleGroupShare{
		{Name: "back", Count: 5, Percentage: 50},
		{Name: "chest", Count: 3, Percentage: 30},
		{Name: "legs", Count: 2, Percentage: 20},
	}, 10, nil
}

func assertOverviewPayload(t *testing.T, overview model.Overview, currentWeight float64) {
	t.Helper()

	if overview.Calendar.Month != "2026-04" {
		t.Fatalf("expected month 2026-04, got %s", overview.Calendar.Month)
	}

	if overview.Calendar.SessionsCount != 3 {
		t.Fatalf("expected 3 monthly sessions, got %d", overview.Calendar.SessionsCount)
	}

	if overview.Calendar.CurrentStreak != 2 {
		t.Fatalf("expected streak 2, got %d", overview.Calendar.CurrentStreak)
	}

	if len(overview.RecentRoutines) != 1 || overview.RecentRoutines[0].Name != "Push Pull Legs" {
		t.Fatalf("unexpected routines payload: %#v", overview.RecentRoutines)
	}

	if len(overview.RecentWorkouts) != 1 || overview.RecentWorkouts[0].Name != "Push Day" {
		t.Fatalf("unexpected workouts payload: %#v", overview.RecentWorkouts)
	}

	if overview.Progress.WeightKg.Current == nil || *overview.Progress.WeightKg.Current != currentWeight {
		t.Fatalf("expected current weight %.1f, got %#v", currentWeight, overview.Progress.WeightKg.Current)
	}

	if overview.Progress.WeightKg.Delta == nil || *overview.Progress.WeightKg.Delta >= 0 {
		t.Fatalf("expected negative weight delta, got %#v", overview.Progress.WeightKg.Delta)
	}

	if overview.MuscleDistribution.YearExerciseCount != 10 {
		t.Fatalf("expected year exercise count 10, got %d", overview.MuscleDistribution.YearExerciseCount)
	}

	if overview.MuscleDistribution.MonthExerciseCount != 3 {
		t.Fatalf("expected month exercise count 3, got %d", overview.MuscleDistribution.MonthExerciseCount)
	}
}

func TestOverviewServiceGetOverview(t *testing.T) {
	referenceDate := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)
	svc, currentWeight := newOverviewServiceTestSubject(referenceDate)

	overview, err := svc.GetOverview(context.Background(), "11111111-1111-1111-1111-111111111111", referenceDate)
	if err != nil {
		t.Fatalf("unexpected error building dashboard overview: %v", err)
	}

	assertOverviewPayload(t, overview, currentWeight)
}

func TestOverviewServiceGetOverviewRequiresUserID(t *testing.T) {
	svc := NewOverviewService(&mockRoutineRepository{}, &mockWorkoutSessionRepository{}, &mockBodyMetricRepository{})

	_, err := svc.GetOverview(context.Background(), "   ", time.Now())
	if err == nil {
		t.Fatal("expected error for empty user id")
	}
}
