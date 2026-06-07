package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
	"github.com/jackc/pgx/v5"
)

type MockTicketRepo struct {
	CreateFunc  func(ctx context.Context, ticket *model.Ticket) error
	ListAllFunc func(ctx context.Context) ([]*model.Ticket, error)
	GetByIDFunc func(ctx context.Context, id string) (*model.Ticket, error)
	CloseFunc   func(ctx context.Context, id string) error
}

func (m *MockTicketRepo) Create(ctx context.Context, ticket *model.Ticket) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, ticket)
	}
	return nil
}

func (m *MockTicketRepo) ListAll(ctx context.Context) ([]*model.Ticket, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockTicketRepo) GetByID(ctx context.Context, id string) (*model.Ticket, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, pgx.ErrNoRows
}

func (m *MockTicketRepo) Close(ctx context.Context, id string) error {
	if m.CloseFunc != nil {
		return m.CloseFunc(ctx, id)
	}
	return nil
}

type MockUserRepo struct {
	GetByIDFunc func(ctx context.Context, id string) (*model.User, error)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *MockUserRepo) Create(ctx context.Context, user *model.User) error { return nil }
func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, nil
}
func (m *MockUserRepo) ListAll(ctx context.Context) ([]*model.User, error) { return nil, nil }
func (m *MockUserRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockUserRepo) MarkAsVerified(ctx context.Context, id string) error {
	return nil
}

func (m *MockUserRepo) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	return nil
}

// --- TESTS ---

func setupTicketHandlerTest() (*gin.Engine, *MockTicketRepo, *MockUserRepo) {
	gin.SetMode(gin.TestMode)
	ticketRepo := &MockTicketRepo{}
	userRepo := &MockUserRepo{}

	ticketService := service.NewTicketService(ticketRepo)
	userService := service.NewUserService(userRepo)
	handler := NewTicketHandler(ticketService, userService)

	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, "user-123")
		c.Next()
	})

	r.POST("/tickets", handler.CreateTicket)
	r.GET("/tickets", handler.ListTickets)
	r.PATCH("/tickets/:id/close", handler.CloseTicket)

	return r, ticketRepo, userRepo
}

func TestTicketHandler_CreateTicket(t *testing.T) {
	r, _, _ := setupTicketHandlerTest()

	body := CreateTicketRequest{
		Category:    "General",
		Title:       "Error",
		Description: "No funciona",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/tickets", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created, got %d", w.Code)
	}
}

func TestTicketHandler_ListTickets_Forbidden(t *testing.T) {
	r, _, userRepo := setupTicketHandlerTest()

	userRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.User, error) {
		return &model.User{Role: "user"}, nil
	}

	req := httptest.NewRequest("GET", "/tickets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403 Forbidden, got %d", w.Code)
	}
}

func TestTicketHandler_ListTickets_AdminSuccess(t *testing.T) {
	r, _, userRepo := setupTicketHandlerTest()

	userRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.User, error) {
		return &model.User{Role: "admin"}, nil
	}

	req := httptest.NewRequest("GET", "/tickets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", w.Code)
	}
}

func TestTicketHandler_CloseTicket_Success(t *testing.T) {
	r, ticketRepo, userRepo := setupTicketHandlerTest()

	userRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.User, error) {
		return &model.User{Role: "admin"}, nil
	}

	ticketRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Ticket, error) {
		return &model.Ticket{ID: id, Status: "open"}, nil
	}

	req := httptest.NewRequest("PATCH", "/tickets/t1/close", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", w.Code)
	}
}
