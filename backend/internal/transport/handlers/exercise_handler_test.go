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
)

type MockExerciseRepository struct {
	createFunc         func(ctx context.Context, exercise *model.Exercise) error
	getByIDFunc        func(ctx context.Context, id string) (*model.Exercise, error)
	listFunc           func(ctx context.Context) ([]model.Exercise, error)
	updateExerciseFunc func(ctx context.Context, exercise *model.Exercise) error
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

func (m *MockExerciseRepository) UpdateExercise(ctx context.Context, exercise *model.Exercise) error {
	if m.updateExerciseFunc != nil {
		return m.updateExerciseFunc(ctx, exercise)
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
		listFunc: func(ctx context.Context) ([]model.Exercise, error) {
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
			}, nil
		},
	}

	exerciseService := service.NewExerciseService(mockRepo)
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/exercises", nil)

	exerciseHandler.ListExercises(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var exercises []model.Exercise
	if err := json.Unmarshal(w.Body.Bytes(), &exercises); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if len(exercises) != 2 {
		t.Errorf("expected 2 exercises, got %d", len(exercises))
	}
}

func TestListExercisesInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockExerciseRepository{
		listFunc: func(ctx context.Context) ([]model.Exercise, error) {
			return nil, errors.New("database error")
		},
	}

	exerciseService := service.NewExerciseService(mockRepo)
	exerciseHandler := NewExerciseHandler(exerciseService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/exercises", nil)

	exerciseHandler.ListExercises(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
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
