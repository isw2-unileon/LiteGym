package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

// RoutineHandler handles routine-related HTTP requests.
type RoutineHandler struct {
	service   *service.RoutineService
	aiService *service.RoutineAIService
}

// NewRoutineHandler creates a new RoutineHandler.
func NewRoutineHandler(svc *service.RoutineService, aiService *service.RoutineAIService) *RoutineHandler {
	return &RoutineHandler{service: svc, aiService: aiService}
}

// GetRoutineByID returns one authenticated user's routine with exercises and planned sets.
func (h *RoutineHandler) GetRoutineByID(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	respondWithResourceByID(c, func(ctx context.Context, routineID string) (*model.Routine, error) {
		return h.service.GetByID(ctx, userID, routineID)
	}, getByIDConfig{
		invalidIDMessage: "invalid routine id",
		notFoundMessage:  "routine not found",
		logMessage:       "failed to retrieve routine",
		internalMessage:  "failed to retrieve routine",
		invalidInputErr:  service.ErrInvalidRoutineInput,
		notFoundErr:      service.ErrRoutineNotFound,
	})
}

// ListRoutines returns the authenticated user's routines.
func (h *RoutineHandler) ListRoutines(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
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

// GenerateRoutineJSON generates and persists a routine for the authenticated user.
func (h *RoutineHandler) GenerateRoutineJSON(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req model.AIRoutineGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := h.aiService.GenerateRoutineJSON(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrAIRoutineInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "objective and duration_minutes are required"})
			return
		}
		if errors.Is(err, service.ErrAIRoutineRateLimited) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "ai generation rate limit exceeded (max 2 per hour)",
				"rate_limit":  response.RateLimit,
				"retry_after": int(time.Until(response.RateLimit.ResetAt).Seconds()),
			})
			return
		}
		if errors.Is(err, service.ErrAIRoutineProviderUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":  "ai provider unavailable",
				"detail": err.Error(),
			})
			return
		}
		if errors.Is(err, service.ErrAIRoutineMissingAPIKey) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "ai provider unavailable: configure GEMINI_API_KEY",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate ai routine"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func authenticatedUserID(c *gin.Context) (string, bool) {
	userIDValue, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		return "", false
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		return "", false
	}

	return userID, true
}
