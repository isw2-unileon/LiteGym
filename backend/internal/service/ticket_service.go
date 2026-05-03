package service

import (
	"context"
	"errors"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/jackc/pgx/v5"
)

var (
	// ErrInvalidTicketInput indicates that the provided ticket data is missing required fields.
	ErrInvalidTicketInput = errors.New("invalid ticket input")
	// ErrTicketNotFound indicates that the requested ticket does not exist in the database.
	ErrTicketNotFound = errors.New("ticket not found")
)

// TicketService provides business logic for support tickets.
type TicketService struct {
	repo repository.TicketRepository
}

// NewTicketService creates a new instance of TicketService.
func NewTicketService(repo repository.TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

// Create handles the business logic for creating a new support ticket.
func (s *TicketService) Create(ctx context.Context, ticket *model.Ticket) error {
	if ticket.Title == "" || ticket.Description == "" {
		return ErrInvalidTicketInput
	}
	return s.repo.Create(ctx, ticket)
}

// ListAll retrieves all support tickets in the system.
func (s *TicketService) ListAll(ctx context.Context) ([]*model.Ticket, error) {
	return s.repo.ListAll(ctx)
}

// Close marks a specific support ticket as closed.
func (s *TicketService) Close(ctx context.Context, id string) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTicketNotFound
		}
		return err
	}
	return s.repo.Close(ctx, id)
}
