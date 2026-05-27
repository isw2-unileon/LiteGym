package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

type routineHandlerTestRepository struct {
	getByIDFunc func(ctx context.Context, userID, routineID string) (*model.Routine, error)
	countFunc   func(ctx context.Context, userID string, since time.Time) (int, error)
}

type routineHandlerTestWorkoutSessionRepository struct{}

type routineHandlerTestBodyMetricRepository struct{}

func (r *routineHandlerTestRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

func (r *routineHandlerTestRepository) ListByUser(ctx context.Context, userID string) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

func (r *routineHandlerTestRepository) GetByID(ctx context.Context, userID, routineID string) (*model.Routine, error) {
	if r.getByIDFunc != nil {
		return r.getByIDFunc(ctx, userID, routineID)
	}
	return nil, nil
}

func (r *routineHandlerTestRepository) SaveGeneratedAIRoutine(ctx context.Context, routine model.AIRoutineToSave) (string, error) {
	return "", nil
}

func (r *routineHandlerTestRepository) CountAIGenerationsInWindow(ctx context.Context, userID string, since time.Time) (int, error) {
	if r.countFunc != nil {
		return r.countFunc(ctx, userID, since)
	}
	return 0, nil
}

func (r *routineHandlerTestRepository) LogAIGeneration(ctx context.Context, userID string, createdAt time.Time) error {
	return nil
}

func (r *routineHandlerTestRepository) ListAvailableExercisesForAI(ctx context.Context, userID string, targetMuscleGroups []string, limit int) ([]model.Exercise, error) {
	return []model.Exercise{}, nil
}

func (r *routineHandlerTestWorkoutSessionRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error) {
	return []model.OverviewWorkoutSummary{}, nil
}

func (r *routineHandlerTestWorkoutSessionRepository) ListRecentWorkoutHistoryByUser(ctx context.Context, userID string, limit int) ([]model.AIRoutineRecentWorkoutSession, error) {
	return []model.AIRoutineRecentWorkoutSession{}, nil
}

func (r *routineHandlerTestWorkoutSessionRepository) ListTrainingDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
	return []time.Time{}, nil
}

func (r *routineHandlerTestWorkoutSessionRepository) ListMuscleDistributionByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error) {
	return []model.OverviewMuscleGroupShare{}, 0, nil
}

func (r *routineHandlerTestBodyMetricRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewBodyMetricEntry, error) {
	return []model.OverviewBodyMetricEntry{}, nil
}

func TestGetRoutineByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	routineService := service.NewRoutineService(&routineHandlerTestRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.Routine, error) {
			return &model.Routine{
				ID:          routineID,
				UserID:      userID,
				Name:        "Push Day",
				Description: "Rutina de empuje",
				Source:      "manual",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Exercises: []model.RoutineExercise{
					{
						ID:            "routine-exercise-1",
						RoutineID:     routineID,
						ExerciseID:    "exercise-1",
						ExerciseName:  "Bench Press",
						MuscleGroup:   "chest",
						ExerciseOrder: 1,
					},
				},
			}, nil
		},
	})
	handler := NewRoutineHandler(routineService, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440111")
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Request = httptest.NewRequest("GET", "/api/routines/550e8400-e29b-41d4-a716-446655440000", nil)

	handler.GetRoutineByID(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var routine model.Routine
	if err := json.Unmarshal(w.Body.Bytes(), &routine); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if routine.Name != "Push Day" {
		t.Fatalf("expected Push Day, got %q", routine.Name)
	}
	if len(routine.Exercises) != 1 || routine.Exercises[0].ExerciseName != "Bench Press" {
		t.Fatalf("unexpected exercises payload: %#v", routine.Exercises)
	}
}

func TestGetRoutineByIDInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewRoutineHandler(service.NewRoutineService(&routineHandlerTestRepository{}), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440111")
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = httptest.NewRequest("GET", "/api/routines/invalid", nil)

	handler.GetRoutineByID(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetRoutineByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewRoutineHandler(service.NewRoutineService(&routineHandlerTestRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.Routine, error) {
			return nil, service.ErrRoutineNotFound
		},
	}), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440111")
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Request = httptest.NewRequest("GET", "/api/routines/550e8400-e29b-41d4-a716-446655440000", nil)

	handler.GetRoutineByID(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetRoutineByIDInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewRoutineHandler(service.NewRoutineService(&routineHandlerTestRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.Routine, error) {
			return nil, errors.New("db error")
		},
	}), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440111")
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Request = httptest.NewRequest("GET", "/api/routines/550e8400-e29b-41d4-a716-446655440000", nil)

	handler.GetRoutineByID(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestGetRoutineByIDUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewRoutineHandler(service.NewRoutineService(&routineHandlerTestRepository{}), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Request = httptest.NewRequest("GET", "/api/routines/550e8400-e29b-41d4-a716-446655440000", nil)

	handler.GetRoutineByID(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUpgradeRoutineJSONRateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &routineHandlerTestRepository{
		countFunc: func(ctx context.Context, userID string, since time.Time) (int, error) {
			return 2, nil
		},
	}
	aiService := service.NewRoutineAIService(repo, &routineHandlerTestWorkoutSessionRepository{}, &routineHandlerTestBodyMetricRepository{}, "test-key", "gemini-2.5-flash")
	handler := NewRoutineHandler(service.NewRoutineService(repo), aiService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440111")
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Request = httptest.NewRequest("POST", "/api/routines/550e8400-e29b-41d4-a716-446655440000/ai/upgrade", bytes.NewBufferString(`{"message":"Make it better"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpgradeRoutineJSON(c)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestUpgradeRoutineJSONNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &routineHandlerTestRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.Routine, error) {
			return nil, service.ErrRoutineNotFound
		},
	}
	aiService := service.NewRoutineAIService(repo, &routineHandlerTestWorkoutSessionRepository{}, &routineHandlerTestBodyMetricRepository{}, "test-key", "gemini-2.5-flash")
	handler := NewRoutineHandler(service.NewRoutineService(repo), aiService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440111")
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Request = httptest.NewRequest("POST", "/api/routines/550e8400-e29b-41d4-a716-446655440000/ai/upgrade", bytes.NewBufferString(`{"message":"Make it better"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpgradeRoutineJSON(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpgradeRoutineJSONInvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &routineHandlerTestRepository{}
	aiService := service.NewRoutineAIService(repo, &routineHandlerTestWorkoutSessionRepository{}, &routineHandlerTestBodyMetricRepository{}, "test-key", "gemini-2.5-flash")
	handler := NewRoutineHandler(service.NewRoutineService(repo), aiService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440111")
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Request = httptest.NewRequest("POST", "/api/routines/550e8400-e29b-41d4-a716-446655440000/ai/upgrade", bytes.NewBufferString(`{"message":"   ","feedback_message":"  "}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpgradeRoutineJSON(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
