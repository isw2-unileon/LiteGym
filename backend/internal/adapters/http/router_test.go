package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/adapters/http/handlers"
	"github.com/isw2-unileon/Grupo-16/backend/internal/core/domain"
	"github.com/isw2-unileon/Grupo-16/backend/internal/core/services"
)

// MockDBPinger para testing
type MockDBPinger struct {
	pingFunc func(ctx context.Context) error
}

func (m *MockDBPinger) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

// MockUserRepository para testing
type MockUserRepository struct{}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return nil, nil
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{
		pingFunc: func(ctx context.Context) error {
			return nil
		},
	}

	mockRepo := &MockUserRepository{}
	userService := services.NewUserService(mockRepo)
	userHandler := handlers.NewUserHandler(userService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(mockDB, userHandler, healthHandler)

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
	userService := services.NewUserService(mockRepo)
	userHandler := handlers.NewUserHandler(userService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(mockDB, userHandler, healthHandler)

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
	userService := services.NewUserService(mockRepo)
	userHandler := handlers.NewUserHandler(userService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(mockDB, userHandler, healthHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/db/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRouterUserEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDBPinger{}
	mockRepo := &MockUserRepository{}
	userService := services.NewUserService(mockRepo)
	userHandler := handlers.NewUserHandler(userService)
	healthHandler := handlers.NewHealthHandler()

	router := SetupRouter(mockDB, userHandler, healthHandler)

	// Test POST /api/users route exists
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/users", nil)
	router.ServeHTTP(w, req)

	// Should not be 404
	if w.Code == http.StatusNotFound {
		t.Error("POST /api/users should exist")
	}
}
