package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/handlers"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
	"github.com/jackc/pgx/v5"
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

/* User Repository with Functions */
type MockUserRepository struct{}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	user.ID = "550e8400-e29b-41d4-a716-446655440000"
	if user.Role == "" {
		user.Role = "user"
	}
	user.CreatedAt = time.Now()
	user.IsActive = true
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
	return nil, pgx.ErrNoRows
}

func (m *MockUserRepository) ListAll(ctx context.Context) ([]*model.User, error) {
	return nil, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	return nil
}

/* Exercise Repository with Functions */
type MockExerciseRepository struct {
	createFunc                        func(ctx context.Context, exercise *model.Exercise) error
	getByIDFunc                       func(ctx context.Context, id string) (*model.Exercise, error)
	listFunc                          func(ctx context.Context, filters model.ExerciseFilter) ([]model.Exercise, int, error)
	listWorkoutSessionsByExerciseFunc func(ctx context.Context, exerciseID, userID string, limit int) ([]model.ExerciseWorkoutSessionSummary, error)
	getInsightsFunc                   func(ctx context.Context, exerciseID, userID string) (model.ExerciseInsights, error)
	updateExerciseFunc                func(ctx context.Context, exercise *model.Exercise) error
	deleteExerciseFunc                func(ctx context.Context, id string) error
}

func (m *MockExerciseRepository) GetByID(ctx context.Context, id string) (*model.Exercise, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockExerciseRepository) NameExists(ctx context.Context, name string, ownerUserID *string, excludeID string) (bool, error) {
	return false, nil
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

/* Routine Repository */
type MockRoutineRepository struct{}

func (m *MockRoutineRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

func (m *MockRoutineRepository) ListByUser(ctx context.Context, userID, search string) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

func (m *MockRoutineRepository) GetByID(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
	return &model.RoutineDetail{ID: routineID, Name: "Routine"}, nil
}

func (m *MockRoutineRepository) CountAIGenerationsInWindow(ctx context.Context, userID string, since time.Time) (int, error) {
	return 0, nil
}

func (m *MockRoutineRepository) SaveGeneratedAIRoutine(ctx context.Context, routine model.AIRoutineToSave) (string, error) {
	return "", nil
}

func (m *MockRoutineRepository) OverwriteGeneratedAIRoutine(ctx context.Context, routineID, userID string, routine model.AIRoutineToSave) error {
	return nil
}

func (m *MockRoutineRepository) LogAIGeneration(ctx context.Context, userID string, createdAt time.Time) error {
	return nil
}

func (m *MockRoutineRepository) ListAvailableExercisesForAI(
	ctx context.Context,
	userID string,
	targetMuscleGroups []string,
	limit int,
) ([]model.Exercise, error) {
	return []model.Exercise{}, nil
}

/* Workout Session Repository */
type MockWorkoutSessionRepository struct{}

/* Workout Session Repository */
type MockOverviewWorkoutRepository struct{}

func (m *MockOverviewWorkoutRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error) {
	return []model.OverviewWorkoutSummary{}, nil
}

func (m *MockWorkoutSessionRepository) ListRecentWorkoutHistoryByUser(ctx context.Context, userID string, limit int) ([]model.AIRoutineRecentWorkoutSession, error) {
	return []model.AIRoutineRecentWorkoutSession{}, nil
}

func (m *MockOverviewWorkoutRepository) ListTrainingDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
	return []time.Time{}, nil
}

func (m *MockOverviewWorkoutRepository) ListPlannedDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
	return []time.Time{}, nil
}

func (m *MockOverviewWorkoutRepository) ListCalendarWorkoutsByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewCalendarWorkout, error) {
	return []model.OverviewCalendarWorkout{}, nil
}

func (m *MockOverviewWorkoutRepository) ListMuscleDistributionByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error) {
	return []model.OverviewMuscleGroupShare{}, 0, nil
}

/* Body Metrics Repository */
type MockBodyMetricRepository struct{}

func (m *MockBodyMetricRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewBodyMetricEntry, error) {
	return []model.OverviewBodyMetricEntry{}, nil
}

/* Workout Repository */
type MockWorkoutRepository struct {
	CreateSessionFunc                     func(ctx context.Context, workout *model.WorkoutSession) error
	GetSessionByIDFunc                    func(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error)
	GetSessionDetailByIDFunc              func(ctx context.Context, id uuid.UUID) (*model.WorkoutSessionDetail, error)
	UpdateSessionByIDFunc                 func(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error
	RemoveSessionByIDFunc                 func(ctx context.Context, id uuid.UUID) error
	CreateWorkoutExerciseFunc             func(ctx context.Context, workoutExercise *model.WorkoutExercise) error
	GetWorkoutExercisesBySessionIDFunc    func(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error)
	CreateWorkoutSetFunc                  func(ctx context.Context, workoutSet *model.WorkoutSet) error
	GetWorkoutSetsByWorkoutExerciseIDFunc func(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error)
	UpdateWorkoutSetFunc                  func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error
	RemoveWorkoutSetFunc                  func(ctx context.Context, setID uuid.UUID) error
}

func (m *MockWorkoutRepository) CreateSession(ctx context.Context, workout *model.WorkoutSession) error {
	if m.CreateSessionFunc != nil {
		return m.CreateSessionFunc(ctx, workout)
	}
	return nil
}

func (m *MockWorkoutRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
	if m.GetSessionByIDFunc != nil {
		return m.GetSessionByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockWorkoutRepository) GetSessionDetailByID(ctx context.Context, id uuid.UUID) (*model.WorkoutSessionDetail, error) {
	if m.GetSessionDetailByIDFunc != nil {
		return m.GetSessionDetailByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockWorkoutRepository) UpdateSessionByID(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error {
	if m.UpdateSessionByIDFunc != nil {
		return m.UpdateSessionByIDFunc(ctx, id, session)
	}
	return nil
}

func (m *MockWorkoutRepository) RemoveSessionByID(ctx context.Context, id uuid.UUID) error {
	if m.RemoveSessionByIDFunc != nil {
		return m.RemoveSessionByIDFunc(ctx, id)
	}
	return nil
}

func (m *MockWorkoutRepository) CreateWorkoutExercise(ctx context.Context, workoutExercise *model.WorkoutExercise) error {
	if m.CreateWorkoutExerciseFunc != nil {
		return m.CreateWorkoutExerciseFunc(ctx, workoutExercise)
	}
	return nil
}

func (m *MockWorkoutRepository) GetWorkoutExercisesBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error) {
	if m.GetWorkoutExercisesBySessionIDFunc != nil {
		return m.GetWorkoutExercisesBySessionIDFunc(ctx, sessionID)
	}
	return nil, nil
}

func (m *MockWorkoutRepository) CreateWorkoutSet(ctx context.Context, workoutSet *model.WorkoutSet) error {
	if m.CreateWorkoutSetFunc != nil {
		return m.CreateWorkoutSetFunc(ctx, workoutSet)
	}
	return nil
}

func (m *MockWorkoutRepository) GetWorkoutSetsByWorkoutExerciseID(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error) {
	if m.GetWorkoutSetsByWorkoutExerciseIDFunc != nil {
		return m.GetWorkoutSetsByWorkoutExerciseIDFunc(ctx, exerciseID)
	}
	return nil, nil
}

func (m *MockWorkoutRepository) UpdateWorkoutSet(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
	if m.UpdateWorkoutSetFunc != nil {
		return m.UpdateWorkoutSetFunc(ctx, setID, setNumber, set)
	}
	return nil
}

func (m *MockWorkoutRepository) RemoveWorkoutSet(ctx context.Context, setID uuid.UUID) error {
	if m.RemoveWorkoutSetFunc != nil {
		return m.RemoveWorkoutSetFunc(ctx, setID)
	}
	return nil
}

/* Testing */

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
	overviewService := service.NewOverviewService(&MockRoutineRepository{}, &MockOverviewWorkoutRepository{}, &MockBodyMetricRepository{})
	return handlers.NewOverviewHandler(overviewService)
}

func newTestRoutineHandler() *handlers.RoutineHandler {
	routineService := service.NewRoutineService(&MockRoutineRepository{})
	return handlers.NewRoutineHandler(routineService, nil)
}

func newTestTicketHandler(userService *service.UserService) *handlers.TicketHandler {
	return handlers.NewTicketHandler(nil, userService)
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{
		pingFunc: func(ctx context.Context) error {
			return nil
		},
	}

	// Repositories
	mockRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)

	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

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
	// Repositories
	mockRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Service
	userService := service.NewUserService(mockRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handler
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)

	router := SetupRouter(
		mockDB,
		userHandler,
		authHandler,
		authMiddleware,
		exerciseHandler,
		newTestRoutineHandler(), newTestOverviewHandler(),
		healthHandler,
		newTestTicketHandler(userService),
		workoutHandler,
		nil,
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
	// Repositories
	mockRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Service
	userService := service.NewUserService(mockRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handler
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)

	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

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
	// Repositories
	mockRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)

	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

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
	// Repositories
	mockRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)

	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

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
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)

	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

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
	// Repositories
	mockUserRepo := &MockUserRepository{}
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
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

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
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	body := strings.NewReader(`{"email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized && w.Code != http.StatusOK {
		t.Errorf("expected auth route to be registered, got status %d", w.Code)
	}
}

func TestRegisterRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}

	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)

	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	body := strings.NewReader(`{"username":"newuser","email":"new@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/api/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected register route to create user, got status %d", w.Code)
	}
}

func TestExerciseListRouteWithValidCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Service
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

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
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

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
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	addAuthCookie(t, req, tokenService)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestWorkoutStartRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	req := httptest.NewRequest("POST", "/api/workout/start", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestWorkoutFinishRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	req := httptest.NewRequest("POST", "/api/workout/550e8400-e29b-41d4-a716-446655440000/finish", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestWorkoutGetByIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	req := httptest.NewRequest("GET", "/api/workout/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestWorkoutRemoveRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	req := httptest.NewRequest("DELETE", "/api/workout/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestWorkoutExerciseCreateRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	req := httptest.NewRequest("POST", "/api/workout/550e8400-e29b-41d4-a716-446655440000/exercise", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestWorkoutExerciseGetByWorkoutIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	req := httptest.NewRequest("GET", "/api/workout/550e8400-e29b-41d4-a716-446655440000/exercises", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestWorkoutSetCreateRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	req := httptest.NewRequest("POST", "/api/workout/550e8400-e29b-41d4-a716-446655440000/exercises/550e8400-e29b-41d4-a716-446655440000/set", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestWorkoutSetGetByExerciseIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	req := httptest.NewRequest("GET", "/api/workout/550e8400-e29b-41d4-a716-446655440000/exercises/550e8400-e29b-41d4-a716-446655440000/sets", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestWorkoutSetUpdateRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	// Repositories
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	// Services
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, "auth_token", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	req := httptest.NewRequest("POST", "/api/workout/550e8400-e29b-41d4-a716-446655440000/exercises/550e8400-e29b-41d4-a716-446655440000/sets/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func (m *MockUserRepository) MarkAsVerified(ctx context.Context, id string) error { return nil }
func (m *MockUserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error { return nil }
func TestWorkoutGetDetailByIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	mockUserRepo := &MockUserRepository{}
	exerciseRepo := &MockExerciseRepository{}
	workoutRepo := &MockWorkoutRepository{}
	userService := service.NewUserService(mockUserRepo)
	tokenService := service.NewTokenService(""test-secret"", ""test-issuer"", time.Hour)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutService := service.NewWorkoutService(workoutRepo)
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, nil, nil, ""auth_token"", false)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, ""auth_token"")
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	router := SetupRouter(mockDB, userHandler, authHandler, authMiddleware, exerciseHandler, newTestRoutineHandler(), newTestOverviewHandler(), healthHandler, newTestTicketHandler(userService), workoutHandler, nil)

	req := httptest.NewRequest(""GET"", ""/api/workout/550e8400-e29b-41d4-a716-446655440000/detail"", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf(""expected status %d, got %d"", http.StatusUnauthorized, w.Code)
	}
}
