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
	getByIDFunc   func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error)
	saveFunc      func(ctx context.Context, routine model.AIRoutineToSave) (string, error)
	overwriteFunc func(ctx context.Context, routineID, userID string, routine model.AIRoutineToSave) error
	countFunc     func(ctx context.Context, userID string, since time.Time) (int, error)
}

type routineHandlerTestWorkoutSessionRepository struct{}

type routineHandlerTestBodyMetricRepository struct{}

type routineHandlerTestExerciseRepository struct{}

func (r *routineHandlerTestRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

func (r *routineHandlerTestRepository) ListByUser(ctx context.Context, userID string) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

func (r *routineHandlerTestRepository) GetByID(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
	if r.getByIDFunc != nil {
		return r.getByIDFunc(ctx, userID, routineID)
	}
	return nil, nil
}

func (r *routineHandlerTestRepository) SaveGeneratedAIRoutine(ctx context.Context, routine model.AIRoutineToSave) (string, error) {
	if r.saveFunc != nil {
		return r.saveFunc(ctx, routine)
	}
	return "", nil
}

func (r *routineHandlerTestRepository) OverwriteGeneratedAIRoutine(ctx context.Context, routineID, userID string, routine model.AIRoutineToSave) error {
	if r.overwriteFunc != nil {
		return r.overwriteFunc(ctx, routineID, userID, routine)
	}
	return nil
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

func (r *routineHandlerTestExerciseRepository) GetByID(ctx context.Context, id string) (*model.Exercise, error) {
	return &model.Exercise{ID: id}, nil
}

func (r *routineHandlerTestExerciseRepository) List(ctx context.Context, filters model.ExerciseFilter) ([]model.Exercise, int, error) {
	return []model.Exercise{}, 0, nil
}

func (r *routineHandlerTestExerciseRepository) ListWorkoutSessionsByExercise(ctx context.Context, exerciseID, userID string, limit int) ([]model.ExerciseWorkoutSessionSummary, error) {
	return nil, nil
}

func (r *routineHandlerTestExerciseRepository) GetInsights(ctx context.Context, exerciseID, userID string) (model.ExerciseInsights, error) {
	return model.ExerciseInsights{}, nil
}

func (r *routineHandlerTestExerciseRepository) Create(ctx context.Context, exercise *model.Exercise) error {
	exercise.ID = "created-exercise"
	return nil
}

func (r *routineHandlerTestExerciseRepository) UpdateExercise(ctx context.Context, exercise *model.Exercise) error {
	return nil
}

func (r *routineHandlerTestExerciseRepository) DeleteExercise(ctx context.Context, id string) error {
	return nil
}

func TestGetRoutineByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	routineService := service.NewRoutineService(&routineHandlerTestRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
			return &model.RoutineDetail{
				ID:          routineID,
				Name:        "Push Day",
				Description: "Rutina de empuje",
				Source:      "manual",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Exercises: []model.RoutineExerciseDetail{
					{
						ID:            "routine-exercise-1",
						ExerciseID:    "exercise-1",
						Name:          "Bench Press",
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

	var routine model.RoutineDetail
	if err := json.Unmarshal(w.Body.Bytes(), &routine); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if routine.Name != "Push Day" {
		t.Fatalf("expected Push Day, got %q", routine.Name)
	}
	if len(routine.Exercises) != 1 || routine.Exercises[0].Name != "Bench Press" {
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
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
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
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
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
			return 20, nil
		},
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
			return &model.RoutineDetail{ID: routineID, Name: "Fallback"}, nil
		},
	}
	aiService := service.NewRoutineAIService(repo, service.NewExerciseService(&routineHandlerTestExerciseRepository{}), &routineHandlerTestWorkoutSessionRepository{}, &routineHandlerTestBodyMetricRepository{}, "test-key", "gemini-2.5-flash")
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

func TestGenerateRoutineJSONRateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &routineHandlerTestRepository{
		countFunc: func(ctx context.Context, userID string, since time.Time) (int, error) {
			return 20, nil
		},
	}
	aiService := service.NewRoutineAIService(repo, service.NewExerciseService(&routineHandlerTestExerciseRepository{}), &routineHandlerTestWorkoutSessionRepository{}, &routineHandlerTestBodyMetricRepository{}, "test-key", "gemini-2.5-flash")
	handler := NewRoutineHandler(service.NewRoutineService(repo), aiService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440111")
	c.Request = httptest.NewRequest("POST", "/api/routines/ai/generate", bytes.NewBufferString(`{"objective":"Strength","duration_minutes":45}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GenerateRoutineJSON(c)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestUpgradeRoutineJSONNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &routineHandlerTestRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
			return nil, service.ErrRoutineNotFound
		},
	}
	aiService := service.NewRoutineAIService(repo, service.NewExerciseService(&routineHandlerTestExerciseRepository{}), &routineHandlerTestWorkoutSessionRepository{}, &routineHandlerTestBodyMetricRepository{}, "test-key", "gemini-2.5-flash")
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
	aiService := service.NewRoutineAIService(repo, service.NewExerciseService(&routineHandlerTestExerciseRepository{}), &routineHandlerTestWorkoutSessionRepository{}, &routineHandlerTestBodyMetricRepository{}, "test-key", "gemini-2.5-flash")
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

func TestSaveUpgradedRoutineAsNew(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &routineHandlerTestRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
			return &model.RoutineDetail{ID: routineID, Name: "Push Day"}, nil
		},
		saveFunc: func(ctx context.Context, routine model.AIRoutineToSave) (string, error) {
			return "new-routine-1", nil
		},
	}
	aiService := service.NewRoutineAIService(repo, service.NewExerciseService(&routineHandlerTestExerciseRepository{}), &routineHandlerTestWorkoutSessionRepository{}, &routineHandlerTestBodyMetricRepository{}, "test-key", "gemini-2.5-flash")
	handler := NewRoutineHandler(service.NewRoutineService(repo), aiService)

	body := `{"routine_json":{"name":"Push Day 2.0","objective":"Refined","duration_minutes":45,"target_muscles":["chest"],"mandatory_count":1,"generated_at":"2026-06-04T10:00:00Z","generation_source":"gemini","exercises":[{"exercise_id":"exercise-1","name":"Bench Press","muscle_group":"chest","exercise_type":"strength","is_mandatory":true,"sets":[{"set_number":1,"target_reps_min":6,"target_reps_max":8}]}]}}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440111")
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Request = httptest.NewRequest("POST", "/api/routines/550e8400-e29b-41d4-a716-446655440000/ai/save-as-new", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SaveUpgradedRoutineAsNew(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestOverwriteRoutineWithAI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &routineHandlerTestRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
			return &model.RoutineDetail{ID: routineID, Name: "Push Day"}, nil
		},
		overwriteFunc: func(ctx context.Context, routineID, userID string, routine model.AIRoutineToSave) error {
			return nil
		},
	}
	aiService := service.NewRoutineAIService(repo, service.NewExerciseService(&routineHandlerTestExerciseRepository{}), &routineHandlerTestWorkoutSessionRepository{}, &routineHandlerTestBodyMetricRepository{}, "test-key", "gemini-2.5-flash")
	handler := NewRoutineHandler(service.NewRoutineService(repo), aiService)

	body := `{"routine_json":{"name":"Push Day 2.0","objective":"Refined","duration_minutes":45,"target_muscles":["chest"],"mandatory_count":1,"generated_at":"2026-06-04T10:00:00Z","generation_source":"gemini","exercises":[{"exercise_id":"exercise-1","name":"Bench Press","muscle_group":"chest","exercise_type":"strength","is_mandatory":true,"sets":[{"set_number":1,"target_reps_min":6,"target_reps_max":8}]}]}}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440111")
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Request = httptest.NewRequest("PUT", "/api/routines/550e8400-e29b-41d4-a716-446655440000/ai/overwrite", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.OverwriteRoutineWithAI(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
