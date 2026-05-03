package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func insertTicketRaw(t *testing.T, db *pgxpool.Pool, userID, title, description, status string) string {
	t.Helper()

	var id string
	err := db.QueryRow(context.Background(), `
		INSERT INTO support_tickets (user_id, title, description, status, created_at, updated_at)
		VALUES ($1::uuid, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id::text
	`, userID, title, description, status).Scan(&id)
	if err != nil {
		t.Fatalf("error insertando ticket en bd: %v", err)
	}

	return id
}

func TestTicketRepositoryCreateIntegration(t *testing.T) {
	db := setupTestDB(t)
	cleanupUsers(t, db)

	userID := insertUserRaw(t, db, "ticketowner", "ticketowner@example.com")

	repo := NewTicketRepository(db)

	ticket := &model.Ticket{
		UserID:      userID,
		Title:       "[General] Test title",
		Description: "Test description",
	}

	err := repo.Create(context.Background(), ticket)
	if err != nil {
		t.Fatalf("no se esperaba error en Create, pero se obtuvo: %v", err)
	}

	if ticket.ID == "" {
		t.Fatal("se esperaba que el ticket tuviera ID tras el Create")
	}

	if ticket.Status != "open" {
		t.Fatalf("se esperaba estado 'open', obtenido '%s'", ticket.Status)
	}

	var dbStatus string
	err = db.QueryRow(context.Background(), `SELECT status FROM support_tickets WHERE id = $1::uuid`, ticket.ID).Scan(&dbStatus)
	if err != nil {
		t.Fatalf("error comprobando ticket creado en la base: %v", err)
	}

	if dbStatus != "open" {
		t.Fatalf("estado en BD incorrecto: esperado open, obtenido %s", dbStatus)
	}
}

func TestTicketRepositoryListAllIntegration(t *testing.T) {
	db := setupTestDB(t)
	cleanupUsers(t, db)

	userID := insertUserRaw(t, db, "listuser", "listuser@example.com")
	insertTicketRaw(t, db, userID, "Ticket 1", "Desc 1", "open")
	insertTicketRaw(t, db, userID, "Ticket 2", "Desc 2", "closed")

	repo := NewTicketRepository(db)
	tickets, err := repo.ListAll(context.Background())
	if err != nil {
		t.Fatalf("no se esperaba error en ListAll, pero se obtuvo: %v", err)
	}

	if len(tickets) != 2 {
		t.Fatalf("se esperaban 2 tickets, se obtuvieron %d", len(tickets))
	}
}

func TestTicketRepositoryGetByIDIntegration(t *testing.T) {
	db := setupTestDB(t)
	cleanupUsers(t, db)

	userID := insertUserRaw(t, db, "getuser", "getuser@example.com")
	ticketID := insertTicketRaw(t, db, userID, "Get Test", "Desc", "open")

	repo := NewTicketRepository(db)
	ticket, err := repo.GetByID(context.Background(), ticketID)
	if err != nil {
		t.Fatalf("no se esperaba error en GetByID, pero se obtuvo: %v", err)
	}

	if ticket.ID != ticketID {
		t.Fatalf("ID incorrecto: esperado %s, obtenido %s", ticketID, ticket.ID)
	}

	if ticket.Status != "open" {
		t.Fatalf("Estado incorrecto: esperado open, obtenido %s", ticket.Status)
	}
}

func TestTicketRepositoryCloseIntegration(t *testing.T) {
	db := setupTestDB(t)
	cleanupUsers(t, db)

	userID := insertUserRaw(t, db, "closeuser", "closeuser@example.com")
	ticketID := insertTicketRaw(t, db, userID, "Close Test", "Desc", "open")

	repo := NewTicketRepository(db)

	err := repo.Close(context.Background(), ticketID)
	if err != nil {
		t.Fatalf("no se esperaba error al cerrar el ticket: %v", err)
	}

	ticket, _ := repo.GetByID(context.Background(), ticketID)
	if ticket.Status != "closed" {
		t.Fatalf("se esperaba que el ticket estuviera 'closed', pero está '%s'", ticket.Status)
	}
	err = repo.Close(context.Background(), "550e8400-e29b-41d4-a716-446655449999")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("se esperaba pgx.ErrNoRows al cerrar ticket inexistente, obtenido: %v", err)
	}
}
