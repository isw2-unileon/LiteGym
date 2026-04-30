package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/handlers"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
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

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	return &model.User{
		ID:       id,
		Email:    "test@example.com",
		Username: "testuser",
		Role:     "user",
	}, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, service.ErrInvalidCredentials
}

func (m *MockUserRepository) ListAll(ctx context.Context) ([]*model.User, error) {
	return nil, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	return nil
}

type MockExerciseRepository struct {
	createFunc         func(ctx context.Context, exercise *model.Exercise) error
	getByIDFunc        func(ctx context.Context, id string) (*model.Exercise, error)
	listFunc           func(ctx context.Context, filters model.ExerciseFilter) ([]model.Exercise, int, error)
	updateExerciseFunc func(ctx context.Context, exercise *model.Exercise) error
	deleteExerciseFunc func(ctx context.Context, id string) error
}

type MockRoutineRepository struct{}

type MockWorkoutSessionRepository struct{}

type MockBodyMetricRepository struct{}

func addAuthCookie(t *testing.T, req *http.Request, tokenService *service.TokenService) {
	t.Helper()

	token, err := tokenService.GenerateToken("550e8400-e29b-41d4-a716-446655440000", "test@example.com", "testuser", "user")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})
}

func newTestOverviewHandler() *handlers.OverviewHandler {
	overviewService := service.NewOverviewService(&MockRoutineRepository{}, &MockWorkoutSessionRepository{}, &MockBodyMetricRepository{})
	return handlers.NewOverviewHandler(overviewService)
}

func newTestTicketHandler(userService *service.UserService) *handlers.TicketHandler {
	return handlers.NewTicketHandler(nil, userService)
}

func (m *MockExerciseRepository) GetByID(ctx context.Context, id string) (*model.Exercise, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockExerciseRepository) Create(ctx context.Context, exercise *model.Exercise) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, exercise)
	}
	return nil
}

func (m *MockExerciseRepository) List(ctx context.Context, filters model.ExerciseFilter) ([]model.Exercise, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filters)
	}
	return []model.Exercise{}, 0, nil
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

func (m *MockRoutineRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

func (m *MockWorkoutSessionRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error) {
	return []model.OverviewWorkoutSummary{}, nil
}

func (m *MockWorkoutSessionRepository) ListTrainingDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
	return []time.Time{}, nil
}

func (m *MockBodyMetricRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewBodyMetricEntry, error) {
	return []model.OverviewBodyMetricEntry{}, nil
}

func (m *MockWorkoutSessionRepository) ListMuscleDistributionByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error) {
	return []model.OverviewMuscleGroupShare{}, 0, nil
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
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCORSPreflightForLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	mockRepo := &MockUserRepository{}
	userService := service.NewUserService(mockRepo)
	userHandler := handlers.NewUserHandler(userService)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(
		mockDB,
		userHandler,
		authHandler,
		authMiddleware,
		exerciseHandler,
		newTestOverviewHandler(),
		healthHandler,
		newTestTicketHandler(userService),
		"http://localhost:5173",
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/api/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:5173, got %q", got)
	}

	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials true, got %q", got)
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
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/hello", nil)
	addAuthCookie(t, req, tokenService)
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
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/db/health", nil)
	addAuthCookie(t, req, tokenService)
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
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/db/health", nil)
	addAuthCookie(t, req, tokenService)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
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
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")

	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)

	healthHandler := handlers.NewHealthHandler()
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService))

	req := httptest.NewRequest("GET", "/api/exercises/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
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
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")

	exerciseRepo := &MockExerciseRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.Exercise, error) {
			return &model.Exercise{
				ID:          "550e8400-e29b-41d4-a716-446655440000",
				Name:        "Bench Press",
				MuscleGroup: "chest",
				IsOfficial:  true,
			}, nil
		},
	}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)

	healthHandler := handlers.NewHealthHandler()
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService))

	req := httptest.NewRequest("GET", "/api/exercises/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
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
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService))

	body := strings.NewReader(`{"email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized && w.Code != http.StatusOK {
		t.Errorf("expected auth route to be registered, got status %d", w.Code)
	}
}

func TestExerciseListRouteWithValidCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	mockUserRepo := &MockUserRepository{}
	userService := service.NewUserService(mockUserRepo)
	userHandler := handlers.NewUserHandler(userService)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService))

	req := httptest.NewRequest("GET", "/api/exercises?page=1&limit=20", nil)
	addAuthCookie(t, req, tokenService)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMeRouteRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	mockUserRepo := &MockUserRepository{}
	userService := service.NewUserService(mockUserRepo)
	userHandler := handlers.NewUserHandler(userService)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService))

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMeRouteWithValidCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	mockUserRepo := &MockUserRepository{}
	userService := service.NewUserService(mockUserRepo)
	userHandler := handlers.NewUserHandler(userService)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseRepo := &MockExerciseRepository{}
	exerciseService := service.NewExerciseService(exerciseRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService))

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	addAuthCookie(t, req, tokenService)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
