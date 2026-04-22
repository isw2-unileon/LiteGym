package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
)

// ExerciseHandler handles exercise-related HTTP requests.
type ExerciseHandler struct {
	service *service.ExerciseService
}

// NewExerciseHandler creates a new ExerciseHandler.
func NewExerciseHandler(svc *service.ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{
		service: svc,
	}
}

// CreateExerciseRequest represents the expected payload for creating an exercise.
type CreateExerciseRequest struct {
	Name                 string `json:"name"`
	Description          string `json:"description"`
	MuscleGroup          string `json:"muscle_group"`
	SecondaryMuscleGroup string `json:"secondary_muscle_group"`
	ExerciseType         string `json:"exercise_type"`
	IsOfficial           *bool  `json:"is_official"`
}

// CreateExercise creates a new exercise.
func (h *ExerciseHandler) CreateExercise(c *gin.Context) {
	var req CreateExerciseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	isOfficial := true
	if req.IsOfficial != nil {
		isOfficial = *req.IsOfficial
	}

	exercise := &model.Exercise{
		Name:                 req.Name,
		Description:          req.Description,
		MuscleGroup:          req.MuscleGroup,
		SecondaryMuscleGroup: req.SecondaryMuscleGroup,
		ExerciseType:         req.ExerciseType,
		IsOfficial:           isOfficial,
	}

	if err := h.service.Create(c.Request.Context(), exercise); err != nil {
		if errors.Is(err, service.ErrInvalidExerciseInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "name and muscle_group are required",
			})
			return
		}

		slog.Error("failed to create exercise", "error", err, "name", req.Name)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create exercise",
		})
		return
	}

	c.JSON(http.StatusCreated, exercise)
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
