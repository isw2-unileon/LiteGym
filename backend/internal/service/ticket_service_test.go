package service

import (
	"context"
	"errors"
	"testing"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
)

// --- MOCK DEL REPOSITORIO ---
type MockTicketRepository struct {
	CreateFunc  func(ctx context.Context, ticket *model.Ticket) error
	ListAllFunc func(ctx context.Context) ([]*model.Ticket, error)
	GetByIDFunc func(ctx context.Context, id string) (*model.Ticket, error)
	CloseFunc   func(ctx context.Context, id string) error
}

func (m *MockTicketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, ticket)
	}
	return nil
}

func (m *MockTicketRepository) ListAll(ctx context.Context) ([]*model.Ticket, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockTicketRepository) GetByID(ctx context.Context, id string) (*model.Ticket, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, pgx.ErrNoRows
}

func (m *MockTicketRepository) Close(ctx context.Context, id string) error {
	if m.CloseFunc != nil {
		return m.CloseFunc(ctx, id)
	}
	return nil
}

// --- TESTS ---

func TestTicketService_Create(t *testing.T) {
	mockRepo := &MockTicketRepository{}
	svc := NewTicketService(mockRepo)
	ctx := context.Background()

	t.Run("Empty Title or Description", func(t *testing.T) {
		err := svc.Create(ctx, &model.Ticket{Title: "", Description: "test"})
		if !errors.Is(err, ErrInvalidTicketInput) {
			t.Errorf("expected ErrInvalidTicketInput, got %v", err)
		}

		err = svc.Create(ctx, &model.Ticket{Title: "test", Description: ""})
		if !errors.Is(err, ErrInvalidTicketInput) {
			t.Errorf("expected ErrInvalidTicketInput, got %v", err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		mockRepo.CreateFunc = func(ctx context.Context, ticket *model.Ticket) error {
			ticket.ID = "123"
			return nil
		}

		ticket := &model.Ticket{Title: "Valid Title", Description: "Valid Desc"}
		err := svc.Create(ctx, ticket)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if ticket.ID != "123" {
			t.Errorf("expected ticket ID to be mutated, got %v", ticket.ID)
		}
	})
}

func TestTicketService_Close(t *testing.T) {
	mockRepo := &MockTicketRepository{}
	svc := NewTicketService(mockRepo)
	ctx := context.Background()

	t.Run("Ticket Not Found", func(t *testing.T) {
		mockRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Ticket, error) {
			return nil, pgx.ErrNoRows
		}

		err := svc.Close(ctx, "999")
		if !errors.Is(err, ErrTicketNotFound) {
			t.Errorf("expected ErrTicketNotFound, got %v", err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		mockRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.Ticket, error) {
			return &model.Ticket{ID: id, Status: "open"}, nil
		}
		mockRepo.CloseFunc = func(ctx context.Context, id string) error {
			return nil
		}

		err := svc.Close(ctx, "123")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
