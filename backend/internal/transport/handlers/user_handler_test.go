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
	"github.com/jackc/pgx/v5"
)

// MockUserRepository is a test double for the user repository.
type MockUserRepository struct {
	createFunc     func(ctx context.Context, user *model.User) error
	getByIDFunc    func(ctx context.Context, id string) (*model.User, error)
	getByEmailFunc func(ctx context.Context, email string) (*model.User, error)
	listAllFunc    func(ctx context.Context) ([]*model.User, error)
	deleteFunc     func(ctx context.Context, id string) error
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
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

func (m *MockUserRepository) ListAll(ctx context.Context) ([]*model.User, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *MockUserRepository) MarkAsVerified(ctx context.Context, id string) error {
	return nil
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	return nil
}

func TestCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockUserRepository{
		createFunc: func(ctx context.Context, user *model.User) error {
			user.ID = "550e8400-e29b-41d4-a716-446655440000"
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
		Role:     "user",
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

	reqBody := CreateUserRequest{
		Username: "testuser",
		
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
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{
				ID:           id,
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
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440000"}}

	c.Request = httptest.NewRequest("GET", "/api/users/550e8400-e29b-41d4-a716-446655440000", nil)

	userHandler.GetUserByID(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var user model.User
	if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

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
	c.Params = gin.Params{{Key: "id", Value: "invalid-user-id"}}

	c.Request = httptest.NewRequest("GET", "/api/users/invalid-user-id", nil)

	userHandler.GetUserByID(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return nil, service.ErrUserNotFound
		},
	}

	userService := service.NewUserService(mockRepo)
	userHandler := NewUserHandler(userService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "550e8400-e29b-41d4-a716-446655440999"}}

	c.Request = httptest.NewRequest("GET", "/api/users/550e8400-e29b-41d4-a716-446655440999", nil)

	userHandler.GetUserByID(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetMe_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{
				ID:       "550e8400-e29b-41d4-a716-446655440999",
				Username: "authenticated_user",
				Email:    "auth@example.com",
			}, nil
		},
	}
	userService := service.NewUserService(mockRepo)
	userHandler := NewUserHandler(userService)

	router := gin.New()

	router.Use(func(ginCtx *gin.Context) {
		ginCtx.Set("user_id", "550e8400-e29b-41d4-a716-446655440999")
		ginCtx.Next()
	})

	router.GET("/api/users/me", userHandler.GetMe)

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "/api/users/me", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var responseUser model.User
	if err := json.Unmarshal(recorder.Body.Bytes(), &responseUser); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if responseUser.Username != "authenticated_user" {
		t.Errorf("expected username 'authenticated_user', got '%s'", responseUser.Username)
	}
}

func TestGetMe_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockUserRepository{}
	userService := service.NewUserService(mockRepo)
	userHandler := NewUserHandler(userService)

	router := gin.New()
	router.GET("/api/users/me", userHandler.GetMe)

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "/api/users/me", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockUserRepository{
		deleteFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	userService := service.NewUserService(mockRepo)
	userHandler := NewUserHandler(userService)

	recorder := httptest.NewRecorder()
	ginCtx, router := gin.CreateTestContext(recorder)

	router.DELETE("/api/users/:id", userHandler.DeleteUser)

	ginCtx.Request = httptest.NewRequest(http.MethodDelete, "/api/users/550e8400-e29b-41d4-a716-446655440000", nil)
	router.ServeHTTP(recorder, ginCtx.Request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockUserRepository{
		deleteFunc: func(ctx context.Context, id string) error {
			return pgx.ErrNoRows
		},
	}

	userService := service.NewUserService(mockRepo)
	userHandler := NewUserHandler(userService)

	recorder := httptest.NewRecorder()
	ginCtx, router := gin.CreateTestContext(recorder)

	router.DELETE("/api/users/:id", userHandler.DeleteUser)

	ginCtx.Request = httptest.NewRequest(http.MethodDelete, "/api/users/00000000-0000-0000-0000-000000000000", nil)
	router.ServeHTTP(recorder, ginCtx.Request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}