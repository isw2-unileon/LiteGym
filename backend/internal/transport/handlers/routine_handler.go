package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

// RoutineHandler handles AI routine generation endpoints.
type RoutineHandler struct {
	aiService *service.RoutineAIService
}

// NewRoutineHandler creates a handler for AI routine generation.
func NewRoutineHandler(aiService *service.RoutineAIService) *RoutineHandler {
	return &RoutineHandler{aiService: aiService}
}

// GenerateRoutineJSON generates and persists a routine for the authenticated user.
func (h *RoutineHandler) GenerateRoutineJSON(c *gin.Context) {
	userIDValue, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
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
