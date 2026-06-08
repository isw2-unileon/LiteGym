package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

// RoutineHandler handles routine-related HTTP requests.
type RoutineHandler struct {
	service       *service.RoutineService
	aiService     *service.RoutineAIService
	manualService *service.ManualRoutineService
}

type aiRoutineSaveRequest struct {
	RoutineJSON model.AIRoutineJSON `json:"routine_json"`
}

type aiRoutineOverwriteRequest struct {
	RoutineJSON model.AIRoutineJSON `json:"routine_json"`
}

type aiRoutinePersistFunc func(ctx *gin.Context, userID, routineID string, routine model.AIRoutineJSON) (model.AIRoutineGenerateResponse, error)

// NewRoutineHandler creates a new RoutineHandler. The manual routine service is optional so existing
// callers (and tests) that only exercise read/AI endpoints keep working.
func NewRoutineHandler(svc *service.RoutineService, aiService *service.RoutineAIService, manualService ...*service.ManualRoutineService) *RoutineHandler {
	handler := &RoutineHandler{service: svc, aiService: aiService}
	if len(manualService) > 0 {
		handler.manualService = manualService[0]
	}
	return handler
}

// GetRoutineByID returns one authenticated user's routine with exercises and planned sets.
func (h *RoutineHandler) GetRoutineByID(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if _, err := uuid.Parse(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routine id"})
		return
	}

	routine, err := h.service.GetByID(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		if errors.Is(err, service.ErrInvalidRoutineInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routine id"})
			return
		}
		if errors.Is(err, service.ErrRoutineNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "routine not found"})
			return
		}

		slog.Error("failed to retrieve routine", "error", err, "user_id", userID, "routine_id", c.Param("id"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve routine"})
		return
	}

	c.JSON(http.StatusOK, routine)
}

// ListRoutines returns the authenticated user's routines.
func (h *RoutineHandler) ListRoutines(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	routines, err := h.service.ListByUser(c.Request.Context(), userID, c.Query("search"))
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

// GetRoutine returns one authenticated user's routine with exercises and planned sets.
func (h *RoutineHandler) GetRoutine(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	routine, err := h.service.GetByID(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		if errors.Is(err, service.ErrInvalidUserInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routine"})
			return
		}
		if errors.Is(err, service.ErrRoutineNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "routine not found"})
			return
		}

		slog.Error("failed to get routine", "error", err, "user_id", userID, "routine_id", c.Param("id"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get routine"})
		return
	}

	c.JSON(http.StatusOK, routine)
}

// GenerateRoutineJSON generates an AI routine preview for the authenticated user.
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

	slog.Info("ai routine generate handler received",
		"user_id", userID,
		"objective", req.Objective,
		"duration_minutes", req.DurationMinutes,
		"target_muscle_groups", req.TargetMuscleGroups,
		"mandatory_exercises_count", len(req.MandatoryExercises),
		"notes_present", strings.TrimSpace(req.Notes) != "",
	)

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
			slog.Error("ai routine provider unavailable", "error", err, "user_id", userID)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":  "ai provider unavailable",
				"detail": err.Error(),
			})
			return
		}
		if errors.Is(err, service.ErrAIRoutineMissingAPIKey) {
			slog.Error("ai routine missing api key", "user_id", userID)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "ai provider unavailable: configure GEMINI_API_KEY",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate ai routine"})
		return
	}

	slog.Info("ai routine generate handler responding ok",
		"user_id", userID,
		"exercise_count", len(response.RoutineJSON.Exercises),
		"generation_source", response.RoutineJSON.GenerationSource,
	)
	c.JSON(http.StatusOK, response)
}

// SaveAIRoutine persists a generated AI routine after the user confirms it.
func (h *RoutineHandler) SaveAIRoutine(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req aiRoutineSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	slog.Info("ai routine save handler received",
		"user_id", userID,
		"exercise_count", len(req.RoutineJSON.Exercises),
		"routine_name", req.RoutineJSON.Name,
	)

	response, err := h.aiService.SaveGeneratedRoutineJSON(c.Request.Context(), userID, req.RoutineJSON)
	if err != nil {
		if errors.Is(err, service.ErrAIRoutineInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routine preview"})
			return
		}
		if errors.Is(err, service.ErrAIRoutineProviderUnavailable) {
			slog.Error("ai routine persistence unavailable", "error", err, "user_id", userID)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":  "ai provider unavailable",
				"detail": err.Error(),
			})
			return
		}

		slog.Error("failed to save ai routine", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save ai routine"})
		return
	}

	slog.Info("ai routine save handler responding ok",
		"user_id", userID,
		"routine_id", response.RoutineID,
		"exercise_count", len(response.RoutineJSON.Exercises),
	)
	c.JSON(http.StatusOK, response)
}

// UpgradeRoutineJSON generates an AI upgrade proposal for an existing routine without persisting it.
func (h *RoutineHandler) UpgradeRoutineJSON(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	routineID := c.Param("id")

	var req model.AIRoutineUpgradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := h.aiService.UpgradeRoutineJSON(c.Request.Context(), userID, routineID, req)
	if err != nil {
		if errors.Is(err, service.ErrAIRoutineInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "message or feedback_message is required"})
			return
		}
		if errors.Is(err, service.ErrRoutineNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "routine not found"})
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

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upgrade ai routine"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SaveUpgradedRoutineAsNew persists one upgrade proposal as a new routine.
func (h *RoutineHandler) SaveUpgradedRoutineAsNew(c *gin.Context) {
	h.persistUpgradedRoutine(c, "failed to save upgraded routine", func(ctx *gin.Context, userID, routineID string, routine model.AIRoutineJSON) (model.AIRoutineGenerateResponse, error) {
		return h.aiService.SaveUpgradedRoutineAsNew(ctx.Request.Context(), userID, routineID, routine)
	})
}

// OverwriteRoutineWithAI replaces one routine with the upgraded proposal.
func (h *RoutineHandler) OverwriteRoutineWithAI(c *gin.Context) {
	h.persistUpgradedRoutine(c, "failed to overwrite upgraded routine", func(ctx *gin.Context, userID, routineID string, routine model.AIRoutineJSON) (model.AIRoutineGenerateResponse, error) {
		return h.aiService.OverwriteRoutineWithGeneratedJSON(ctx.Request.Context(), userID, routineID, routine)
	})
}

func (h *RoutineHandler) persistUpgradedRoutine(
	c *gin.Context,
	internalErrorMessage string,
	persist aiRoutinePersistFunc,
) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	routineID := c.Param("id")
	var req aiRoutineOverwriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := persist(c, userID, routineID, req.RoutineJSON)
	if err != nil {
		if errors.Is(err, service.ErrAIRoutineInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routine preview"})
			return
		}
		if errors.Is(err, service.ErrRoutineNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "routine not found"})
			return
		}
		if errors.Is(err, service.ErrAIRoutineProviderUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":  "ai provider unavailable",
				"detail": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": internalErrorMessage})
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
