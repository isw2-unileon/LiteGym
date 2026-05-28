package transport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/handlers"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

func newRateLimitTestRouter() (*gin.Engine, *service.TokenService) {
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
	authHandler := handlers.NewAuthHandler(userService, tokenService, "auth_token", false)
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
		newTestRoutineHandler(),
		newTestOverviewHandler(),
		healthHandler,
		newTestTicketHandler(userService),
		workoutHandler,
	)

	return router, tokenService
}

func addAuthCookieForUserID(t *testing.T, req *http.Request, tokenService *service.TokenService, userID string) {
	t.Helper()

	token, err := tokenService.GenerateToken(userID, "test@example.com", "testuser", "user")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})
}

func TestLoginRateLimitByIP(t *testing.T) {
	router, _ := newRateLimitTestRouter()

	ipOne := "203.0.113.10:1234"
	var lastStatus int
	var lastRetryAfter string

	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":"test@example.com","password":"password123"}`))
		req.RemoteAddr = ipOne
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		lastStatus = w.Code
		lastRetryAfter = w.Header().Get("Retry-After")
	}

	if lastStatus != http.StatusTooManyRequests {
		t.Fatalf("expected %d after repeated login attempts, got %d", http.StatusTooManyRequests, lastStatus)
	}

	if lastRetryAfter == "" {
		t.Fatalf("expected Retry-After header on rate-limited login response")
	}

	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":"test@example.com","password":"password123"}`))
	req.RemoteAddr = "203.0.113.11:1234"
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusTooManyRequests {
		t.Fatalf("expected a different IP to have its own login budget, got %d", w.Code)
	}
}

func TestRegisterRateLimitByIP(t *testing.T) {
	router, _ := newRateLimitTestRouter()

	ip := "203.0.113.20:1234"
	var lastStatus int

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest("POST", "/api/users", strings.NewReader(`{"email":"test@example.com"}`))
		req.RemoteAddr = ip
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		lastStatus = w.Code
	}

	if lastStatus != http.StatusTooManyRequests {
		t.Fatalf("expected %d after repeated registration attempts, got %d", http.StatusTooManyRequests, lastStatus)
	}
}

func TestProtectedRateLimitByUserID(t *testing.T) {
	router, tokenService := newRateLimitTestRouter()

	ip := "203.0.113.30:1234"
	var lastStatus int
	var lastRetryAfter string

	for i := 0; i < 21; i++ {
		req := httptest.NewRequest("GET", "/api/hello", nil)
		req.RemoteAddr = ip
		addAuthCookieForUserID(t, req, tokenService, "550e8400-e29b-41d4-a716-446655440000")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		lastStatus = w.Code
		lastRetryAfter = w.Header().Get("Retry-After")
	}

	if lastStatus != http.StatusTooManyRequests {
		t.Fatalf("expected %d after repeated authenticated requests, got %d", http.StatusTooManyRequests, lastStatus)
	}

	if lastRetryAfter == "" {
		t.Fatalf("expected Retry-After header on rate-limited protected response")
	}

	req := httptest.NewRequest("GET", "/api/hello", nil)
	req.RemoteAddr = ip
	addAuthCookieForUserID(t, req, tokenService, "550e8400-e29b-41d4-a716-446655440999")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusTooManyRequests {
		t.Fatalf("expected a different user to have its own protected budget, got %d", w.Code)
	}
}
