package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
)

// MockUserRepository para testing
type MockUserRepository struct {
	createFunc     func(ctx context.Context, user *model.User) error
	getByIDFunc    func(ctx context.Context, id int) (*model.User, error)
	getByEmailFunc func(ctx context.Context, email string) (*model.User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	return nil, nil
}

func TestCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockUserRepository{
		createFunc: func(ctx context.Context, user *model.User) error {
			user.ID = 1
			user.CreatedAt = time.Now()
			return nil
		},
	}

	userService := service.NewUserService(mockRepo)
	userHandler := NewUserHandler(userService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	userHandler.CreateUser(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestCreateUserMissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockUserRepository{}
	userService := service.NewUserService(mockRepo)
	userHandler := NewUserHandler(userService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Request incompleto
	reqBody := CreateUserRequest{
		Username: "testuser",
		// Email vacío
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	userHandler.CreateUser(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetUserByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id int) (*model.User, error) {
			return &model.User{
				ID:           int(id),
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hashed",
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	userService := service.NewUserService(mockRepo)
	userHandler := NewUserHandler(userService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	c.Request = httptest.NewRequest("GET", "/api/users/1", nil)

	userHandler.GetUserByID(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var user model.User
	json.Unmarshal(w.Body.Bytes(), &user)

	if user.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", user.Username)
	}
}

func TestGetUserByIDInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockUserRepository{}
	userService := service.NewUserService(mockRepo)
	userHandler := NewUserHandler(userService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	c.Request = httptest.NewRequest("GET", "/api/users/invalid", nil)

	userHandler.GetUserByID(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id int) (*model.User, error) {
			return nil, service.ErrUserNotFound
		},
	}

	userService := service.NewUserService(mockRepo)
	userHandler := NewUserHandler(userService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	c.Request = httptest.NewRequest("GET", "/api/users/999", nil)

	userHandler.GetUserByID(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
