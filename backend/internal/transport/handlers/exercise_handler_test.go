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

type MockExerciseRepository struct {
	createFunc                        func(ctx context.Context, exercise *model.Exercise) error
	getByIDFunc                       func(ctx context.Context, id string) (*model.Exercise, error)
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

func TestGetExerciseByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockExerciseRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.Exercise, error) {
			return &model.Exercise{
				ID:          id,
				Name:        "Bench Press",
				MuscleGroup: "chest",
				IsOfficial:  true,
				CreatedAt:   time.Now(),
			}, nil
		},
	}

	exerciseService := service.NewExerciseService(mockRepo)
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Request = httptest.NewRequest("GET", "/api/exercises/550e8400-e29b-41d4-a716-446655440000", nil)

	exerciseHandler.GetExerciseByID(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var exercise model.Exercise
	if err := json.Unmarshal(w.Body.Bytes(), &exercise); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if exercise.Name != "Bench Press" {
		t.Errorf("expected name Bench Press, got %s", exercise.Name)
	}
}

func TestGetExerciseByIDInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(mockRepo)
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = httptest.NewRequest("GET", "/api/exercises/invalid", nil)

	exerciseHandler.GetExerciseByID(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetExerciseByIDZeroID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(mockRepo)
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "0"}}
	c.Request = httptest.NewRequest("GET", "/api/exercises/0", nil)

	exerciseHandler.GetExerciseByID(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetExerciseByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockExerciseRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.Exercise, error) {
			return nil, service.ErrExerciseNotFound
		},
	}

	exerciseService := service.NewExerciseService(mockRepo)
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Request = httptest.NewRequest("GET", "/api/exercises/550e8400-e29b-41d4-a716-446655440000", nil)

	exerciseHandler.GetExerciseByID(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetExerciseByIDInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockExerciseRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.Exercise, error) {
			return nil, errors.New("database error")
		},
	}

	exerciseService := service.NewExerciseService(mockRepo)
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Request = httptest.NewRequest("GET", "/api/exercises/550e8400-e29b-41d4-a716-446655440000", nil)

	exerciseHandler.GetExerciseByID(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestListExercises(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockExerciseRepository{
		listFunc: func(ctx context.Context, filters model.ExerciseFilter) ([]model.Exercise, int, error) {
			return []model.Exercise{
				{
					ID:          "550e8400-e29b-41d4-a716-446655440000",
					Name:        "Bench Press",
					MuscleGroup: "chest",
					IsOfficial:  true,
					CreatedAt:   time.Now(),
				},
				{
					ID:          "550e8400-e29b-41d4-a716-446655440001",
					Name:        "Squat",
					MuscleGroup: "legs",
					IsOfficial:  true,
					CreatedAt:   time.Now(),
				},
			}, 2, nil
		},
	}

	exerciseService := service.NewExerciseService(mockRepo)
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/exercises?page=1&limit=20", nil)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440111")

	exerciseHandler.ListExercises(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result model.ExerciseListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if len(result.Items) != 2 {
		t.Errorf("expected 2 exercises, got %d", len(result.Items))
	}
}

func TestListExercisesInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockExerciseRepository{
		listFunc: func(ctx context.Context, filters model.ExerciseFilter) ([]model.Exercise, int, error) {
			return nil, 0, errors.New("database error")
		},
	}

	exerciseService := service.NewExerciseService(mockRepo)
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/exercises?page=1&limit=20", nil)
	c.Set(middleware.ContextUserIDKey, "550e8400-e29b-41d4-a716-446655440111")

	exerciseHandler.ListExercises(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestGetExerciseMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	exerciseService := service.NewExerciseService(&MockExerciseRepository{})
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/exercises/metadata", nil)

	exerciseHandler.GetMetadata(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response model.ExerciseMetadataResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if !hasOptionValue(response.ExerciseTypes, "strength") {
		t.Fatal("expected exercise_types to include strength")
	}

	if !hasOptionValue(response.MuscleGroups, "chest") {
		t.Fatal("expected muscle_groups to include chest")
	}
}

func hasOptionValue(options []model.SelectOption, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}

	return false
}

func TestUpdateExerciseOfficialRequiresAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	exerciseService := service.NewExerciseService(mockRepo)
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Set("user_role", "user")
	c.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/exercises/550e8400-e29b-41d4-a716-446655440000",
		bytes.NewBufferString(`{"name":"Bench Press","muscle_group":"chest","is_official":true}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")

	exerciseHandler.UpdateExercise(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestDeleteExerciseOfficialRequiresAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	exerciseService := service.NewExerciseService(mockRepo)
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Set("user_role", "user")
	c.Request = httptest.NewRequest(
		http.MethodDelete,
		"/api/exercises/550e8400-e29b-41d4-a716-446655440000",
		nil,
	)

	exerciseHandler.DeleteExercise(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestDeleteExerciseSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockExerciseRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.Exercise, error) {
			return &model.Exercise{
				ID:          id,
				Name:        "Custom Curl",
				MuscleGroup: "biceps",
				IsOfficial:  false,
			}, nil
		},
		deleteExerciseFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	exerciseService := service.NewExerciseService(mockRepo)
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}
	c.Set("user_role", "user")
	c.Request = httptest.NewRequest(
		http.MethodDelete,
		"/api/exercises/550e8400-e29b-41d4-a716-446655440000",
		nil,
	)

	exerciseHandler.DeleteExercise(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
