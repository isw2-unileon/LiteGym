package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

// ExerciseHandler handles exercise-related HTTP requests.
type ExerciseHandler struct {
	service *service.ExerciseService
}

type createExerciseRequest struct {
	Name                  string   `json:"name"`
	Description           string   `json:"description"`
	MuscleGroup           string   `json:"muscle_group"`
	SecondaryMuscleGroups []string `json:"secondary_muscle_groups"`
	ExerciseType          string   `json:"exercise_type"`
	IsOfficial            bool     `json:"is_official"`
}

const adminRole = "admin"

// NewExerciseHandler creates a new ExerciseHandler.
func NewExerciseHandler(svc *service.ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{
		service: svc,
	}
}

// GetExerciseByID retrieves an exercise by its ID.
func (h *ExerciseHandler) GetExerciseByID(c *gin.Context) {
	respondWithResourceByID(c, h.service.GetByID, getByIDConfig{
		invalidIDMessage: "invalid exercise id",
		notFoundMessage:  "exercise not found",
		logMessage:       "failed to retrieve exercise",
		internalMessage:  "failed to retrieve exercise",
		invalidInputErr:  service.ErrInvalidExerciseInput,
		notFoundErr:      service.ErrExerciseNotFound,
	})
}

// ListExercises returns all exercises.
func (h *ExerciseHandler) ListExercises(c *gin.Context) {
	page, err := strconv.Atoi(c.Query("page"))
	limit, err2 := strconv.Atoi(c.Query("limit"))

	official := c.Query("official")

	if err != nil || err2 != nil || page < 1 || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination parameters"})
		return
	}

	var officialFilter *bool
	if official != "" {
		isOfficial, err := strconv.ParseBool(official)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'official' query parameter"})
			return
		}
		officialFilter = &isOfficial
	}

	filters := model.ExerciseFilter{
		Search:      c.Query("search"),
		Type:        c.Query("type"),
		MuscleGroup: c.Query("muscle_group"),
		Official:    officialFilter,
		Page:        page,
		Limit:       limit,
	}

	response, err := h.service.List(c.Request.Context(), filters)
	if err != nil {
		slog.Error("failed to list exercises", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list exercises"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetMetadata returns the valid exercise options used by clients.
func (h *ExerciseHandler) GetMetadata(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.GetMetadata())
}

// ListWorkoutSessionsByExercise returns the authenticated user's workout sessions containing an exercise.
func (h *ExerciseHandler) ListWorkoutSessionsByExercise(c *gin.Context) {
	exerciseID, err := uuid.Parse(c.Param("id"))
	if err != nil || exerciseID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exercise id"})
		return
	}

	userIDValue, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, _ := userIDValue.(string)
	sessions, err := h.service.ListWorkoutSessionsByExercise(c.Request.Context(), exerciseID.String(), userID, 10)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidExerciseInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exercise id"})
		case errors.Is(err, service.ErrInvalidUserInput):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		default:
			slog.Error("failed to list workout sessions for exercise", "error", err, "exercise_id", exerciseID.String(), "user_id", userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve workout sessions for exercise"})
		}
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// GetExerciseInsights returns performance history for the authenticated user and selected exercise.
func (h *ExerciseHandler) GetExerciseInsights(c *gin.Context) {
	exerciseID, err := uuid.Parse(c.Param("id"))
	if err != nil || exerciseID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exercise id"})
		return
	}

	userIDValue, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, _ := userIDValue.(string)
	insights, err := h.service.GetInsights(c.Request.Context(), exerciseID.String(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidExerciseInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exercise id"})
		case errors.Is(err, service.ErrInvalidUserInput):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		default:
			slog.Error("failed to get exercise insights", "error", err, "exercise_id", exerciseID.String(), "user_id", userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve exercise insights"})
		}
		return
	}

	c.JSON(http.StatusOK, insights)
}

// CreateExercise creates a new exercise.
func (h *ExerciseHandler) CreateExercise(c *gin.Context) {
	var req createExerciseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	role := userRoleFromContext(c)
	isOfficial, ok := requestedOfficialFlag(c, req.IsOfficial, role, "create")
	if !ok {
		return
	}

	exercise := model.Exercise{
		Name:                 req.Name,
		Description:          req.Description,
		MuscleGroup:          req.MuscleGroup,
		SecondaryMuscleGroup: strings.Join(req.SecondaryMuscleGroups, ", "),
		ExerciseType:         req.ExerciseType,
		IsOfficial:           isOfficial,
	}

	err := h.service.Create(c.Request.Context(), &exercise)
	if err != nil {
		if errors.Is(err, service.ErrInvalidExerciseInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid exercise input",
			})
			return
		}

		slog.Error("failed to create exercise", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create exercise",
		})
		return
	}

	c.JSON(http.StatusCreated, exercise)
}

// UpdateExercise updates an existing exercise by its ID.
func (h *ExerciseHandler) UpdateExercise(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exercise id"})
		return
	}

	var req createExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	role := userRoleFromContext(c)
	existingExercise, ok := h.getExerciseForMutation(c, id, "update")
	if !ok {
		return
	}

	if !allowOfficialMutation(c, existingExercise, role, "update") {
		return
	}

	isOfficial, ok := requestedOfficialFlag(c, req.IsOfficial, role, "update")
	if !ok {
		return
	}

	exercise := model.Exercise{
		ID:                   id,
		Name:                 req.Name,
		Description:          req.Description,
		MuscleGroup:          req.MuscleGroup,
		SecondaryMuscleGroup: strings.Join(req.SecondaryMuscleGroups, ", "),
		ExerciseType:         req.ExerciseType,
		IsOfficial:           isOfficial,
	}

	err := h.service.UpdateExercise(c.Request.Context(), &exercise)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidExerciseInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exercise input"})
		case errors.Is(err, service.ErrExerciseNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "exercise not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update exercise"})
		}
		return
	}

	c.JSON(http.StatusOK, exercise)
}

// DeleteExercise soft-deletes an existing exercise by its ID.
func (h *ExerciseHandler) DeleteExercise(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exercise id"})
		return
	}

	role := userRoleFromContext(c)
	existingExercise, ok := h.getExerciseForMutation(c, id, "delete")
	if !ok {
		return
	}

	if !allowOfficialMutation(c, existingExercise, role, "delete") {
		return
	}

	if err := h.service.DeleteExercise(c.Request.Context(), id); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidExerciseInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exercise id"})
		case errors.Is(err, service.ErrExerciseNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "exercise not found"})
		default:
			slog.Error("failed to delete exercise", "error", err, "id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete exercise"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "exercise deleted successfully"})
}

func userRoleFromContext(c *gin.Context) string {
	userRole, _ := c.Get(middleware.ContextUserRoleKey)
	role, _ := userRole.(string)
	return role
}

func allowOfficialMutation(c *gin.Context, exercise *model.Exercise, role, action string) bool {
	if exercise.IsOfficial && role != adminRole {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "only admins can " + action + " official exercises",
		})
		return false
	}
	return true
}

func requestedOfficialFlag(c *gin.Context, requestedOfficial bool, role, action string) (bool, bool) {
	isOfficial := requestedOfficial && role == adminRole
	if requestedOfficial && role != adminRole {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "only admins can " + action + " official exercises",
		})
		return false, false
	}
	return isOfficial, true
}

func (h *ExerciseHandler) getExerciseForMutation(c *gin.Context, id, action string) (*model.Exercise, bool) {
	existingExercise, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidExerciseInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exercise id"})
		case errors.Is(err, service.ErrExerciseNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "exercise not found"})
		default:
			slog.Error("failed to retrieve exercise before "+action, "error", err, "id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to " + action + " exercise"})
		}
		return nil, false
	}

	return existingExercise, true
}
