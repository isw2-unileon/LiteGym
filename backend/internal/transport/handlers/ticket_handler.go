package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

// TicketHandler handles HTTP requests related to support tickets.
type TicketHandler struct {
	ticketService *service.TicketService
	userService   *service.UserService
}

// NewTicketHandler creates a new instance of TicketHandler.
func NewTicketHandler(ts *service.TicketService, us *service.UserService) *TicketHandler {
	return &TicketHandler{ticketService: ts, userService: us}
}

// CreateTicketRequest represents the JSON payload received from the frontend to create a ticket.
type CreateTicketRequest struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CreateTicket processes the HTTP request to create a new ticket.
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	formattedTitle := fmt.Sprintf("[%s] %s", req.Category, req.Title)

	ticket := &model.Ticket{
		UserID:      userIDVal.(string),
		Title:       formattedTitle,
		Description: req.Description,
	}

	if err := h.ticketService.Create(c.Request.Context(), ticket); err != nil {
		if errors.Is(err, service.ErrInvalidTicketInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create ticket"})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

func (h *TicketHandler) isAdmin(c *gin.Context) bool {
	userIDVal, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		return false
	}
	user, err := h.userService.GetByID(c.Request.Context(), userIDVal.(string))
	if err != nil || user.Role != "admin" {
		return false
	}
	return true
}

// ListTickets processes the HTTP request to retrieve all tickets (admin only).
func (h *TicketHandler) ListTickets(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	tickets, err := h.ticketService.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tickets"})
		return
	}
	c.JSON(http.StatusOK, tickets)
}

// CloseTicket processes the HTTP request to close a specific ticket (admin only).
func (h *TicketHandler) CloseTicket(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	id := c.Param("id")
	if err := h.ticketService.Close(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ticket closed successfully"})
}
