package transport

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/handlers"
)

type MockDBPinger struct {
	pingFunc func(ctx context.Context) error
}

func (m *MockDBPinger) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

type MockUserRepository struct{}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return &model.User{
		ID:           1,
		Username:     "testuser",
		Email:        email,
		PasswordHash: "$2a$10$8RQm8JDJr2jP6T4Sl7dQ4eWmWj6nAcsoEhW8Y2r7H5P7D6Jj4Lw3S",
	}, nil
}

type MockExerciseRepository struct {
	createFunc  func(ctx context.Context, exercise *model.Exercise) error
	getByIDFunc func(ctx context.Context, id int64) (*model.Exercise, error)
	listFunc    func(ctx context.Context) ([]model.Exercise, error)
}

func (m *MockExerciseRepository) Create(ctx context.Context, exercise *model.Exercise) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, exercise)
	}
	return nil
}

func (m *MockExerciseRepository) GetByID(ctx context.Context, id int64) (*model.Exercise, error) {
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

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{
		pingFunc: func(ctx context.Context) error {
			return nil
		},
	}

	mockRepo := &MockUserRepository{}
	userService := service.NewUserService(mockRepo)
	userHandler := handlers.NewUserHandler(userService)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(mockDB, userHandler, authHandler, exerciseHandler, healthHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHelloEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	mockRepo := &MockUserRepository{}
	userService := service.NewUserService(mockRepo)
	userHandler := handlers.NewUserHandler(userService)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(mockDB, userHandler, authHandler, exerciseHandler, healthHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/hello", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDBHealthEndpointSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{
		pingFunc: func(ctx context.Context) error {
			return nil
		},
	}

	mockRepo := &MockUserRepository{}
	userService := service.NewUserService(mockRepo)
	userHandler := handlers.NewUserHandler(userService)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(mockDB, userHandler, authHandler, exerciseHandler, healthHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/db/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDBHealthEndpointError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{
		pingFunc: func(ctx context.Context) error {
			return context.DeadlineExceeded
		},
	}

	mockRepo := &MockUserRepository{}
	userService := service.NewUserService(mockRepo)
	userHandler := handlers.NewUserHandler(userService)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(mockDB, userHandler, authHandler, exerciseHandler, healthHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/db/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestExerciseCreateRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	mockUserRepo := &MockUserRepository{}
	userService := service.NewUserService(mockUserRepo)
	userHandler := handlers.NewUserHandler(userService)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)

	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)

	healthHandler := handlers.NewHealthHandler()
	router := SetupRouter(mockDB, userHandler, authHandler, exerciseHandler, healthHandler)

	body := bytes.NewBufferString(`{"name":"Bench Press","muscle_group":"chest"}`)
	req := httptest.NewRequest("POST", "/api/exercises", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestExerciseListRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	mockUserRepo := &MockUserRepository{}
	userService := service.NewUserService(mockUserRepo)
	userHandler := handlers.NewUserHandler(userService)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)

	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)

	healthHandler := handlers.NewHealthHandler()
	router := SetupRouter(mockDB, userHandler, authHandler, exerciseHandler, healthHandler)

	req := httptest.NewRequest("GET", "/api/exercises", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExerciseGetByIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	mockUserRepo := &MockUserRepository{}
	userService := service.NewUserService(mockUserRepo)
	userHandler := handlers.NewUserHandler(userService)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)

	exerciseRepo := &MockExerciseRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*model.Exercise, error) {
			return &model.Exercise{
				ID:          1,
				Name:        "Bench Press",
				MuscleGroup: "chest",
				IsOfficial:  true,
				CreatedAt:   time.Now(),
			}, nil
		},
	}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)

	healthHandler := handlers.NewHealthHandler()
	router := SetupRouter(mockDB, userHandler, authHandler, exerciseHandler, healthHandler)

	req := httptest.NewRequest("GET", "/api/exercises/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestLoginRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	mockUserRepo := &MockUserRepository{}
	userService := service.NewUserService(mockUserRepo)
	userHandler := handlers.NewUserHandler(userService)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	router := SetupRouter(mockDB, userHandler, authHandler, exerciseHandler, healthHandler)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized && w.Code != http.StatusOK {
		t.Errorf("expected auth route to be registered, got status %d", w.Code)
	}
}
