package handler

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
)

type mockUserRepository struct {
	createErr   error
	getByIDErr  error
	userToReturn *model.User
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) error {
	if m.createErr != nil {
		return m.createErr
	}

	user.ID = 1
	user.CreatedAt = time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}

	return m.userToReturn, nil
}

func TestCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockUserRepository{}
	handler := NewUserHandler(repo)

	router := gin.New()
	router.POST("/users", handler.CreateUser)

	body := []byte(`{
		"username": "diego",
		"email": "diego@example.com",
		"password": "123456"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if response["username"] != "diego" {
		t.Fatalf("expected username 'diego', got %v", response["username"])
	}

	if response["email"] != "diego@example.com" {
		t.Fatalf("expected email 'diego@example.com', got %v", response["email"])
	}
}

func TestCreateUserInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockUserRepository{}
	handler := NewUserHandler(repo)

	router := gin.New()
	router.POST("/users", handler.CreateUser)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateUserMissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockUserRepository{}
	handler := NewUserHandler(repo)

	router := gin.New()
	router.POST("/users", handler.CreateUser)

	body := []byte(`{
		"username": "diego",
		"email": ""
	}`)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetUserByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockUserRepository{
		userToReturn: &model.User{
			ID:        1,
			Username:  "diego",
			Email:     "diego@example.com",
			CreatedAt: time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC),
		},
	}
	handler := NewUserHandler(repo)

	router := gin.New()
	router.GET("/users/:id", handler.GetUserByID)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if response["username"] != "diego" {
		t.Fatalf("expected username 'diego', got %v", response["username"])
	}
}

func TestGetUserByIDInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockUserRepository{}
	handler := NewUserHandler(repo)

	router := gin.New()
	router.GET("/users/:id", handler.GetUserByID)

	req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockUserRepository{
		getByIDErr: errors.New("not found"),
	}
	handler := NewUserHandler(repo)

	router := gin.New()
	router.GET("/users/:id", handler.GetUserByID)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}