package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

// --- Mock ProfileRepository for handler tests ---

type mockProfileRepo struct {
	getStatsFunc      func(ctx context.Context, userID string, timeRange string, year int, month int) (*model.ProfileStats, error)
	upsertGoalsFunc   func(ctx context.Context, goal *model.UserGoal) error
	addBodyMetricFunc func(ctx context.Context, userID string, req *model.AddBodyMetricRequest) error
}

func (m *mockProfileRepo) GetStats(ctx context.Context, userID string, timeRange string, year int, month int) (*model.ProfileStats, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx, userID, timeRange, year, month)
	}
	return &model.ProfileStats{}, nil
}

func (m *mockProfileRepo) UpsertGoals(ctx context.Context, goal *model.UserGoal) error {
	if m.upsertGoalsFunc != nil {
		return m.upsertGoalsFunc(ctx, goal)
	}
	return nil
}

func (m *mockProfileRepo) AddBodyMetric(ctx context.Context, userID string, req *model.AddBodyMetricRequest) error {
	if m.addBodyMetricFunc != nil {
		return m.addBodyMetricFunc(ctx, userID, req)
	}
	return nil
}

func newTestProfileService(repo *mockProfileRepo) *service.ProfileService {
	return service.NewProfileService(repo)
}

// --- GetDashboard ---

func TestProfileHandlerGetDashboard_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockProfileRepo{
		getStatsFunc: func(_ context.Context, _ string, _ string, _ int, _ int) (*model.ProfileStats, error) {
			return &model.ProfileStats{TotalWorkouts: 7}, nil
		},
	}

	handler := NewProfileHandler(newTestProfileService(repo))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440000")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/profile?range=month&year=2026&month=5", nil)

	handler.GetDashboard(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", w.Code, w.Body.String())
	}

	var result model.ProfileStats
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result.TotalWorkouts != 7 {
		t.Errorf("expected TotalWorkouts 7, got %d", result.TotalWorkouts)
	}
}

func TestProfileHandlerGetDashboard_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewProfileHandler(newTestProfileService(&mockProfileRepo{}))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/profile", nil)

	handler.GetDashboard(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestProfileHandlerGetDashboard_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockProfileRepo{
		getStatsFunc: func(_ context.Context, _ string, _ string, _ int, _ int) (*model.ProfileStats, error) {
			return nil, errors.New("db error")
		},
	}
	handler := NewProfileHandler(newTestProfileService(repo))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440000")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/profile", nil)

	handler.GetDashboard(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// --- UpdateGoals ---

func TestProfileHandlerUpdateGoals_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedGoal *model.UserGoal
	repo := &mockProfileRepo{
		upsertGoalsFunc: func(_ context.Context, goal *model.UserGoal) error {
			capturedGoal = goal
			return nil
		},
	}
	handler := NewProfileHandler(newTestProfileService(repo))

	body := `{"short_term":"Perder 3kg","long_term":"Maratón","target_days":4}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440000")
	c.Request = httptest.NewRequest(http.MethodPut, "/api/profile/goals", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateGoals(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", w.Code, w.Body.String())
	}
	if capturedGoal == nil {
		t.Fatal("expected UpsertGoals to be called")
	}
	if capturedGoal.ShortTerm != "Perder 3kg" {
		t.Errorf("expected ShortTerm 'Perder 3kg', got %q", capturedGoal.ShortTerm)
	}
	if capturedGoal.TargetDaysPerWeek != 4 {
		t.Errorf("expected TargetDaysPerWeek 4, got %d", capturedGoal.TargetDaysPerWeek)
	}
}

func TestProfileHandlerUpdateGoals_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewProfileHandler(newTestProfileService(&mockProfileRepo{}))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/profile/goals", nil)

	handler.UpdateGoals(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestProfileHandlerUpdateGoals_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewProfileHandler(newTestProfileService(&mockProfileRepo{}))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440000")
	c.Request = httptest.NewRequest(http.MethodPut, "/api/profile/goals", bytes.NewBufferString("not json"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateGoals(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProfileHandlerUpdateGoals_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockProfileRepo{
		upsertGoalsFunc: func(_ context.Context, _ *model.UserGoal) error {
			return errors.New("db error")
		},
	}
	handler := NewProfileHandler(newTestProfileService(repo))

	body := `{"short_term":"test","long_term":"test","target_days":3}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440000")
	c.Request = httptest.NewRequest(http.MethodPut, "/api/profile/goals", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateGoals(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// --- AddBodyMetric ---

func TestProfileHandlerAddBodyMetric_ReturnsCreated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedReq *model.AddBodyMetricRequest
	repo := &mockProfileRepo{
		addBodyMetricFunc: func(_ context.Context, _ string, req *model.AddBodyMetricRequest) error {
			capturedReq = req
			return nil
		},
	}
	handler := NewProfileHandler(newTestProfileService(repo))

	body := `{"weight_kg":78.5,"body_fat_percentage":18.5}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440000")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/profile/metrics", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.AddBodyMetric(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", w.Code, w.Body.String())
	}
	if capturedReq == nil {
		t.Fatal("expected AddBodyMetric to be called")
	}
	if capturedReq.WeightKg != 78.5 {
		t.Errorf("expected WeightKg 78.5, got %f", capturedReq.WeightKg)
	}
	if capturedReq.BodyFatPercentage == nil || *capturedReq.BodyFatPercentage != 18.5 {
		t.Errorf("expected BodyFatPercentage 18.5, got %v", capturedReq.BodyFatPercentage)
	}
}

func TestProfileHandlerAddBodyMetric_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewProfileHandler(newTestProfileService(&mockProfileRepo{}))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/profile/metrics", nil)

	handler.AddBodyMetric(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestProfileHandlerAddBodyMetric_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewProfileHandler(newTestProfileService(&mockProfileRepo{}))

	// weight_kg is required and must be > 0; sending 0 fails binding
	body := `{"weight_kg":0}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440000")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/profile/metrics", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.AddBodyMetric(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d – body: %s", w.Code, w.Body.String())
	}
}

func TestProfileHandlerAddBodyMetric_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockProfileRepo{
		addBodyMetricFunc: func(_ context.Context, _ string, _ *model.AddBodyMetricRequest) error {
			return errors.New("insert failed")
		},
	}
	handler := NewProfileHandler(newTestProfileService(repo))

	body := `{"weight_kg":80.0}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440000")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/profile/metrics", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.AddBodyMetric(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// --- GetAIAnalysis ---

func TestProfileHandlerGetAIAnalysis_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockProfileRepo{
		getStatsFunc: func(_ context.Context, _ string, _ string, _ int, _ int) (*model.ProfileStats, error) {
			return &model.ProfileStats{
				TotalWorkouts: 2,
				TotalSets:     10,
				Goals: &model.UserGoal{
					ShortTerm: "Tonificar brazos",
				},
				MuscleRadar: []model.MuscleRadarStat{{Muscle: "Brazos", Value: 6}},
			}, nil
		},
	}

	handler := NewProfileHandler(newTestProfileService(repo))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440000")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/profile/ai-analysis", nil)

	handler.GetAIAnalysis(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", w.Code, w.Body.String())
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	analysis := result["analysis"]
	if analysis == "" {
		t.Error("expected analysis text in response, got empty")
	}
	if !strings.Contains(analysis, "Tonificar brazos") {
		t.Errorf("expected analysis to contain goal 'Tonificar brazos', got: %s", analysis)
	}
}

func TestProfileHandlerGetAIAnalysis_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewProfileHandler(newTestProfileService(&mockProfileRepo{}))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/profile/ai-analysis", nil)

	handler.GetAIAnalysis(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
