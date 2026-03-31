package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
)

type ExerciseHandler struct {
	service *service.ExerciseService
}

func NewExerciseHandler(svc *service.ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{
		service: svc,
	}
}

type CreateExerciseRequest struct {
	Name                 string `json:"name"`
	Description          string `json:"description"`
	MuscleGroup          string `json:"muscle_group"`
	SecondaryMuscleGroup string `json:"secondary_muscle_group"`
	ExerciseType         string `json:"exercise_type"`
	IsOfficial           *bool  `json:"is_official"`
}

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

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create exercise",
		})
		return
	}

	c.JSON(http.StatusCreated, exercise)
}

func (h *ExerciseHandler) GetExerciseByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid exercise id",
		})
		return
	}

	exercise, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidExerciseInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid exercise id",
			})
			return
		}

		if errors.Is(err, service.ErrExerciseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "exercise not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve exercise",
		})
		return
	}

	c.JSON(http.StatusOK, exercise)
}

func (h *ExerciseHandler) ListExercises(c *gin.Context) {
	exercises, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list exercises",
		})
		return
	}

	c.JSON(http.StatusOK, exercises)
}
