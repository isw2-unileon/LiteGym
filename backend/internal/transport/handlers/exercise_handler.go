package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
	exercises, err := h.service.List(c.Request.Context())
	if err != nil {
		slog.Error("failed to list exercises", "error", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list exercises",
		})
		return
	}

	c.JSON(http.StatusOK, exercises)
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

	userRole, _ := c.Get(middleware.ContextUserRoleKey)
	role, _ := userRole.(string)

	isOfficial := req.IsOfficial && role == "admin"
	if req.IsOfficial && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "only admins can create official exercises",
		})
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

	userRole, _ := c.Get(middleware.ContextUserRoleKey)
	role, _ := userRole.(string)

	existingExercise, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidExerciseInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exercise id"})
		case errors.Is(err, service.ErrExerciseNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "exercise not found"})
		default:
			slog.Error("failed to retrieve exercise before update", "error", err, "id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update exercise"})
		}
		return
	}

	if existingExercise.IsOfficial && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "only admins can update official exercises",
		})
		return
	}

	isOfficial := req.IsOfficial && role == "admin"
	if req.IsOfficial && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "only admins can create official exercises",
		})
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

	err = h.service.UpdateExercise(c.Request.Context(), &exercise)

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
