package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
)

// --- Mock ProfileRepository ---

type mockProfileRepository struct {
	getStatsFunc      func(ctx context.Context, userID string, timeRange string, year int, month int) (*model.ProfileStats, error)
	upsertGoalsFunc   func(ctx context.Context, goal *model.UserGoal) error
	addBodyMetricFunc func(ctx context.Context, userID string, req *model.AddBodyMetricRequest) error
}

func (m *mockProfileRepository) GetStats(ctx context.Context, userID string, timeRange string, year int, month int) (*model.ProfileStats, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx, userID, timeRange, year, month)
	}
	return &model.ProfileStats{}, nil
}

func (m *mockProfileRepository) UpsertGoals(ctx context.Context, goal *model.UserGoal) error {
	if m.upsertGoalsFunc != nil {
		return m.upsertGoalsFunc(ctx, goal)
	}
	return nil
}

func (m *mockProfileRepository) AddBodyMetric(ctx context.Context, userID string, req *model.AddBodyMetricRequest) error {
	if m.addBodyMetricFunc != nil {
		return m.addBodyMetricFunc(ctx, userID, req)
	}
	return nil
}

// --- GetDashboardStats ---

func TestProfileServiceGetDashboardStats_Success(t *testing.T) {
	expectedStats := &model.ProfileStats{
		TotalWorkouts: 5,
		TotalSets:     30,
		TopExercises:  []model.ExerciseStat{{Name: "Bench Press", Sets: 10}},
	}

	repo := &mockProfileRepository{
		getStatsFunc: func(_ context.Context, userID, timeRange string, year, month int) (*model.ProfileStats, error) {
			if userID != "user-1" {
				t.Errorf("unexpected userID: %s", userID)
			}
			return expectedStats, nil
		},
	}

	svc := NewProfileService(repo)
	stats, err := svc.GetDashboardStats(context.Background(), "user-1", "month", 2026, 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalWorkouts != expectedStats.TotalWorkouts {
		t.Errorf("expected TotalWorkouts %d, got %d", expectedStats.TotalWorkouts, stats.TotalWorkouts)
	}
	if len(stats.TopExercises) != 1 || stats.TopExercises[0].Name != "Bench Press" {
		t.Errorf("unexpected TopExercises: %+v", stats.TopExercises)
	}
}

func TestProfileServiceGetDashboardStats_PropagatesRepoError(t *testing.T) {
	repoErr := errors.New("db connection failed")
	repo := &mockProfileRepository{
		getStatsFunc: func(_ context.Context, _, _ string, _, _ int) (*model.ProfileStats, error) {
			return nil, repoErr
		},
	}

	svc := NewProfileService(repo)
	_, err := svc.GetDashboardStats(context.Background(), "user-1", "month", 2026, 5)

	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got: %v", err)
	}
}

// --- UpdateGoals ---

func TestProfileServiceUpdateGoals_Success(t *testing.T) {
	var capturedGoal *model.UserGoal

	repo := &mockProfileRepository{
		upsertGoalsFunc: func(_ context.Context, goal *model.UserGoal) error {
			capturedGoal = goal
			return nil
		},
	}

	svc := NewProfileService(repo)
	goal := &model.UserGoal{
		ShortTerm:         "Perder 3kg",
		LongTerm:          "Correr una maratón",
		TargetDaysPerWeek: 4,
	}

	if err := svc.UpdateGoals(context.Background(), goal); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedGoal == nil {
		t.Fatal("expected UpsertGoals to be called")
	}
	if capturedGoal.ShortTerm != goal.ShortTerm {
		t.Errorf("expected ShortTerm %q, got %q", goal.ShortTerm, capturedGoal.ShortTerm)
	}
	if capturedGoal.TargetDaysPerWeek != 4 {
		t.Errorf("expected TargetDaysPerWeek 4, got %d", capturedGoal.TargetDaysPerWeek)
	}
}

func TestProfileServiceUpdateGoals_PropagatesRepoError(t *testing.T) {
	repoErr := errors.New("constraint violation")
	repo := &mockProfileRepository{
		upsertGoalsFunc: func(_ context.Context, _ *model.UserGoal) error {
			return repoErr
		},
	}

	svc := NewProfileService(repo)
	err := svc.UpdateGoals(context.Background(), &model.UserGoal{})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got: %v", err)
	}
}

// --- AddBodyMetric ---

func TestProfileServiceAddBodyMetric_Success(t *testing.T) {
	fat := 18.5
	var capturedUserID string
	var capturedReq *model.AddBodyMetricRequest

	repo := &mockProfileRepository{
		addBodyMetricFunc: func(_ context.Context, userID string, req *model.AddBodyMetricRequest) error {
			capturedUserID = userID
			capturedReq = req
			return nil
		},
	}

	svc := NewProfileService(repo)
	req := &model.AddBodyMetricRequest{
		WeightKg:          78.5,
		BodyFatPercentage: &fat,
	}

	if err := svc.AddBodyMetric(context.Background(), "user-1", req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedUserID != "user-1" {
		t.Errorf("expected userID user-1, got %s", capturedUserID)
	}
	if capturedReq.WeightKg != 78.5 {
		t.Errorf("expected WeightKg 78.5, got %f", capturedReq.WeightKg)
	}
	if capturedReq.BodyFatPercentage == nil || *capturedReq.BodyFatPercentage != fat {
		t.Errorf("expected BodyFatPercentage %f, got %v", fat, capturedReq.BodyFatPercentage)
	}
}

func TestProfileServiceAddBodyMetric_PropagatesRepoError(t *testing.T) {
	repoErr := errors.New("insert failed")
	repo := &mockProfileRepository{
		addBodyMetricFunc: func(_ context.Context, _ string, _ *model.AddBodyMetricRequest) error {
			return repoErr
		},
	}

	svc := NewProfileService(repo)
	err := svc.AddBodyMetric(context.Background(), "user-1", &model.AddBodyMetricRequest{WeightKg: 70})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got: %v", err)
	}
}

// --- GenerateAIAnalysis ---

func TestProfileServiceGenerateAIAnalysis_OfflineFallback(t *testing.T) {
	svc := NewProfileService(&mockProfileRepository{}) // No API key -> offline fallback
	stats := &model.ProfileStats{
		TotalWorkouts: 4,
		TotalSets:     12,
		Goals: &model.UserGoal{
			ShortTerm: "Ganar fuerza",
		},
		MuscleRadar: []model.MuscleRadarStat{{Muscle: "Hombros", Value: 8}},
	}

	analysis, err := svc.GenerateAIAnalysis(context.Background(), stats)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if analysis == "" {
		t.Error("expected non-empty analysis feedback")
	}
	if !strings.Contains(analysis, "Ganar fuerza") {
		t.Errorf("expected analysis to contain short-term goal 'Ganar fuerza', got: %s", analysis)
	}
	if !strings.Contains(analysis, "Hombros") {
		t.Errorf("expected analysis to contain top muscle 'Hombros', got: %s", analysis)
	}
}
