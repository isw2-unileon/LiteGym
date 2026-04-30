package repository

import (
	"context"
	"errors"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TicketRepository defines the interface for interacting with the support_tickets table.
type TicketRepository interface {
	Create(ctx context.Context, ticket *model.Ticket) error
	ListAll(ctx context.Context) ([]*model.Ticket, error)
	GetByID(ctx context.Context, id string) (*model.Ticket, error)
	Close(ctx context.Context, id string) error
}

type ticketRepository struct {
	db *pgxpool.Pool
}

// NewTicketRepository creates a new instance of TicketRepository.
func NewTicketRepository(db *pgxpool.Pool) TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	query := `
		INSERT INTO support_tickets (user_id, title, description, status, created_at, updated_at)
		VALUES ($1::uuid, $2, $3, 'open', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id::text, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query, ticket.UserID, ticket.Title, ticket.Description).
		Scan(&ticket.ID, &ticket.CreatedAt, &ticket.UpdatedAt)
	if err != nil {
		return err
	}
	ticket.Status = "open"
	return nil
}

func (r *ticketRepository) ListAll(ctx context.Context) ([]*model.Ticket, error) {
	query := `
		SELECT id::text, user_id::text, title, description, status::text, created_at, updated_at
		FROM support_tickets
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*model.Ticket
	for rows.Next() {
		var t model.Ticket
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, &t)
	}
	return tickets, rows.Err()
}

func (r *ticketRepository) GetByID(ctx context.Context, id string) (*model.Ticket, error) {
	query := `SELECT id::text, status::text FROM support_tickets WHERE id = $1::uuid`
	var t model.Ticket
	err := r.db.QueryRow(ctx, query, id).Scan(&t.ID, &t.Status)

	// FIX: errorlint pide usar errors.Is
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}
	return &t, err
}

func (r *ticketRepository) Close(ctx context.Context, id string) error {
	query := `UPDATE support_tickets SET status = 'closed', updated_at = CURRENT_TIMESTAMP WHERE id = $1::uuid`
	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
