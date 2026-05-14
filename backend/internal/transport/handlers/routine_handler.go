package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

// RoutineHandler handles routine-related HTTP requests.
type RoutineHandler struct {
	service *service.RoutineService
}

// NewRoutineHandler creates a new RoutineHandler.
func NewRoutineHandler(svc *service.RoutineService) *RoutineHandler {
	return &RoutineHandler{service: svc}
}

// ListRoutines returns the authenticated user's routines.
func (h *RoutineHandler) ListRoutines(c *gin.Context) {
	userIDValue, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	routines, err := h.service.ListByUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUserInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
			return
		}

		slog.Error("failed to list routines", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list routines"})
		return
	}

	c.JSON(http.StatusOK, routines)
}
