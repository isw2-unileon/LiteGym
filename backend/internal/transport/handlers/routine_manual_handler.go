package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
)

type manualRoutineSetRequest struct {
	SetNumber             int      `json:"set_number"`
	TargetRepsMin         *int     `json:"target_reps_min"`
	TargetRepsMax         *int     `json:"target_reps_max"`
	TargetRepsText        string   `json:"target_reps_text"`
	TargetWeightKg        *float64 `json:"target_weight_kg"`
	TargetDurationSeconds *int     `json:"target_duration_seconds"`
	TargetDistanceKm      *float64 `json:"target_distance_km"`
	TargetRir             *int     `json:"target_rir"`
	RestSeconds           *int     `json:"rest_seconds"`
	Notes                 string   `json:"notes"`
}

type manualRoutineExerciseRequest struct {
	ExerciseID string                   `json:"exercise_id"`
	Sets       []manualRoutineSetRequest `json:"sets"`
}

// manualRoutineRequest is the body accepted when creating or editing a manual routine.
// Objective and TargetMuscleGroups are already composed into Notes by the client and are kept
// here only to document the contract; the persisted description is built from Notes.
type manualRoutineRequest struct {
	Name               string                         `json:"name"`
	Objective          string                         `json:"objective"`
	RoutineType        string                         `json:"routine_type"`
	TargetMuscleGroups []string                       `json:"target_muscle_groups"`
	Notes              string                         `json:"notes"`
	Exercises          []manualRoutineExerciseRequest `json:"exercises"`
}

// CreateRoutine persists a new manual routine for the authenticated user.
func (h *RoutineHandler) CreateRoutine(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if h.manualService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "manual routines unavailable"})
		return
	}

	var req manualRoutineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	routineID, err := h.manualService.Create(c.Request.Context(), userID, manualRoutineInputFrom(req))
	if err != nil {
		h.handleManualRoutineError(c, err, userID, "", "failed to create routine")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"routine_id": routineID})
}

// UpdateRoutine replaces the contents of an existing manual routine owned by the user.
func (h *RoutineHandler) UpdateRoutine(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if h.manualService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "manual routines unavailable"})
		return
	}

	routineID := c.Param("id")
	if _, err := uuid.Parse(routineID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routine id"})
		return
	}

	var req manualRoutineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.manualService.Update(c.Request.Context(), userID, routineID, manualRoutineInputFrom(req)); err != nil {
		h.handleManualRoutineError(c, err, userID, routineID, "failed to update routine")
		return
	}

	c.JSON(http.StatusOK, gin.H{"routine_id": routineID})
}

// DeleteRoutine removes a routine owned by the user.
func (h *RoutineHandler) DeleteRoutine(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if h.manualService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "manual routines unavailable"})
		return
	}

	routineID := c.Param("id")
	if _, err := uuid.Parse(routineID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routine id"})
		return
	}

	if err := h.manualService.Delete(c.Request.Context(), userID, routineID); err != nil {
		h.handleManualRoutineError(c, err, userID, routineID, "failed to delete routine")
		return
	}

	c.Status(http.StatusNoContent)
}

// DuplicateRoutine copies an existing routine into a new one owned by the user.
func (h *RoutineHandler) DuplicateRoutine(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if h.manualService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "manual routines unavailable"})
		return
	}

	routineID := c.Param("id")
	if _, err := uuid.Parse(routineID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routine id"})
		return
	}

	newID, err := h.manualService.Duplicate(c.Request.Context(), userID, routineID)
	if err != nil {
		h.handleManualRoutineError(c, err, userID, routineID, "failed to duplicate routine")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"routine_id": newID})
}

func manualRoutineInputFrom(req manualRoutineRequest) service.ManualRoutineInput {
	exercises := make([]service.ManualRoutineExerciseInput, 0, len(req.Exercises))
	for _, ex := range req.Exercises {
		sets := make([]service.ManualRoutineSetInput, 0, len(ex.Sets))
		for _, st := range ex.Sets {
			sets = append(sets, service.ManualRoutineSetInput{
				SetNumber:             st.SetNumber,
				TargetRepsMin:         st.TargetRepsMin,
				TargetRepsMax:         st.TargetRepsMax,
				TargetRepsText:        st.TargetRepsText,
				TargetWeightKg:        st.TargetWeightKg,
				TargetDurationSeconds: st.TargetDurationSeconds,
				TargetDistanceKm:      st.TargetDistanceKm,
				TargetRir:             st.TargetRir,
				RestSeconds:           st.RestSeconds,
				Notes:                 st.Notes,
			})
		}

		exercises = append(exercises, service.ManualRoutineExerciseInput{
			ExerciseID: ex.ExerciseID,
			Sets:       sets,
		})
	}

	return service.ManualRoutineInput{
		Name:        req.Name,
		Description: req.Notes,
		RoutineType: req.RoutineType,
		Exercises:   exercises,
	}
}

func (h *RoutineHandler) handleManualRoutineError(c *gin.Context, err error, userID, routineID, internalMessage string) {
	switch {
	case errors.Is(err, service.ErrRoutineNameRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "routine name is required"})
	case errors.Is(err, service.ErrInvalidRoutineInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid routine input"})
	case errors.Is(err, service.ErrRoutineLimitReached):
		c.JSON(http.StatusConflict, gin.H{"error": "routine limit reached"})
	case errors.Is(err, service.ErrRoutineNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "routine not found"})
	default:
		slog.Error(internalMessage, "error", err, "user_id", userID, "routine_id", routineID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": internalMessage})
	}
}
