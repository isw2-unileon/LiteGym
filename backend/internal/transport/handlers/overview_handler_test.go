package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
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

func (m *mockRoutineRepository) ListByUser(ctx context.Context, userID string) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

type mockOverviewWorkoutRepository struct {
	listRecentByUserFunc             func(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error)
	listTrainingDatesInRangeFunc     func(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error)
	listMuscleDistributionByUserFunc func(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error)
}

func (m *mockOverviewWorkoutRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error) {
	if m.listRecentByUserFunc != nil {
		return m.listRecentByUserFunc(ctx, userID, limit)
	}
	return []model.OverviewWorkoutSummary{}, nil
}

func (m *mockOverviewWorkoutRepository) ListTrainingDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
	if m.listTrainingDatesInRangeFunc != nil {
		return m.listTrainingDatesInRangeFunc(ctx, userID, from, to)
	}
	return []time.Time{}, nil
}

func (m *mockOverviewWorkoutRepository) ListPlannedDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
	return []time.Time{}, nil
}

func (m *mockOverviewWorkoutRepository) ListCalendarWorkoutsByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewCalendarWorkout, error) {
	return []model.OverviewCalendarWorkout{}, nil
}

func (m *mockOverviewWorkoutRepository) ListMuscleDistributionByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error) {
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

func TestOverviewHandlerGetOverview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	routineRepo := &mockRoutineRepository{
		listRecentByUserFunc: func(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error) {
			return []model.OverviewRoutineSummary{{ID: "routine-1", Name: "Push Pull Legs"}}, nil
		},
	}
	overviewWorkoutRepo := &mockOverviewWorkoutRepository{}
	bodyMetricRepo := &mockBodyMetricRepository{}

	svc := service.NewOverviewService(routineRepo, overviewWorkoutRepo, bodyMetricRepo)
	handler := NewOverviewHandler(svc)
	handler.nowFunc = func() time.Time {
		return time.Date(2026, time.April, 29, 10, 0, 0, 0, time.UTC)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "11111111-1111-1111-1111-111111111111")
	c.Request = httptest.NewRequest("GET", "/api/dashboard", nil)

	handler.GetOverview(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result model.Overview
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if result.Calendar.Month != "2026-04" {
		t.Fatalf("expected month 2026-04, got %s", result.Calendar.Month)
	}
}

func TestOverviewHandlerRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewOverviewService(&mockRoutineRepository{}, &mockOverviewWorkoutRepository{}, &mockBodyMetricRepository{})
	handler := NewOverviewHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/dashboard", nil)

	handler.GetOverview(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
