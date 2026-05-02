package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
)

// WorkoutHandler handles workout-related HTTP requests.
type WorkoutHandler struct {
	service *service.WorkoutService
}

// NewWorkoutHandler creates a new WorkoutHandler.
func NewWorkoutHandler(svc *service.WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{
		service: svc,
	}
}

type createWorkoutRequest struct {
	UserID    uuid.UUID  `json:"user_id"`
	RoutineID *uuid.UUID `json:"routine_id"`
	Name      string     `json:"name"`
	Notes     *string    `json:"notes"`
}

type finishWorkoutRequest struct {
	Name           string   `json:"name"`
	Duration       *int     `json:"duration_minutes"`
	CaloriesBurned *float64 `json:"calories_burned"`
	Notes          *string  `json:"notes"`
}

type createExerciseToWorkoutRequest struct {
	ExerciseID    uuid.UUID `json:"exercise_id"`
	ExerciseOrder int       `json:"exercise_order"`
	Notes         string    `json:"notes"`
}

type createSetToExerciseRequest struct {
	SetNumber   int      `json:"set_number"`
	Repetitions *int     `json:"reps"`
	WeightKg    *float64 `json:"weight_kg"`
	Duration    *int     `json:"duration_seconds"`
	DistanceKm  *float64 `json:"distance_km"`
	Rir         *int     `json:"rir"`
	Completed   *bool    `json:"completed"`
}

/* Workout Session */

// CreateWorkout handles the HTTP Request for creating a new Workout via the
// API Route /api/workout/start
func (h *WorkoutHandler) CreateWorkout(c *gin.Context) {
	var req createWorkoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	workout := &model.WorkoutSession{
		UserID:    req.UserID,
		RoutineID: req.RoutineID,
		Name:      req.Name,
		Notes:     req.Notes,
	}

	if err := h.service.CreateSession(c.Request.Context(), workout); err != nil {
		if errors.Is(err, service.ErrInvalidWorkoutSessionInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "user_id and name are required",
			})
			return
		}

		slog.Error("failed to create workout session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create workout session",
		})
		return
	}

	c.JSON(http.StatusCreated, workout)
}

// FinishWorkout handles the HTTP Request for finishing a Workout Session via the
// API Route /api/workout/:id/finish
func (h *WorkoutHandler) FinishWorkout(c *gin.Context) {
	workoutID := c.Param("id")
	parsedWorkoutID, ok := ParseUUID(workoutID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workout id is invalid",
		})
		c.Abort()
		return
	}

	var req finishWorkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	updatedSession := &model.WorkoutSession{
		Name:           req.Name,
		EndedAt:        TimePointer(time.Now()),
		Duration:       req.Duration,
		CaloriesBurned: req.CaloriesBurned,
		Notes:          req.Notes,
	}

	if err := h.service.UpdateSessionByID(c.Request.Context(), parsedWorkoutID, updatedSession); err != nil {
		if errors.Is(err, service.ErrInvalidWorkoutSessionInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid workout id or missing required fields",
			})
			return
		}
		if errors.Is(err, service.ErrWorkoutNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "workout session not found",
			})
			return
		}

		slog.Error("failed to finish workout session", "error", err, "workout_id", workoutID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to finish workout session due to a internal error",
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

// GetWorkoutByID handles the HTTP Request for getting a Workout Session using the
// Session ID via the API Route /api/workout/:id
func (h *WorkoutHandler) GetWorkoutByID(c *gin.Context) {
	parsedWorkoutID, ok := ParseUUID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
		return
	}

	session, err := h.service.GetSessionByID(c.Request.Context(), parsedWorkoutID)
	if handleWorkoutError(c, err, "failed to retrieve workout session", "workout_id", c.Param("id")) {
		return
	}
	c.JSON(http.StatusOK, session)
}

// RemoveWorkout handles the HTTP Request for deleting a Workout Session using the
// Session ID via the API Route /api/workout/:id
func (h *WorkoutHandler) RemoveWorkout(c *gin.Context) {
	workoutID := c.Param("id")
	parsedWorkoutID, ok := ParseUUID(workoutID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid workout id",
		})
		c.Abort()
		return
	}

	if err := h.service.RemoveSessionByID(c.Request.Context(), parsedWorkoutID); err != nil {
		if errors.Is(err, service.ErrInvalidWorkoutSessionInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid workout id",
			})
			return
		}
		if errors.Is(err, service.ErrWorkoutNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "workout session not found",
			})
			return
		}

		slog.Error("failed to remove workout session", "error", err, "workout_id", workoutID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to remove workout session",
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

/* Workout Exercises */

// CreateWorkoutExercise handles the HTTP Request for creating a new Workout Exercise
// associated with a Workout Session via the API Route /api/workout/:id/exercise
func (h *WorkoutHandler) CreateWorkoutExercise(c *gin.Context) {
	workoutID := c.Param("id")
	parsedWorkoutID, ok := ParseUUID(workoutID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid workout id",
		})
		c.Abort()
		return
	}

	var req createExerciseToWorkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	workoutExercise := &model.WorkoutExercise{
		WorkoutSessionID: parsedWorkoutID,
		ExerciseID:       req.ExerciseID,
		ExerciseOrder:    req.ExerciseOrder,
		Notes:            req.Notes,
	}

	if err := h.service.CreateWorkoutExercise(c.Request.Context(), workoutExercise); err != nil {
		if errors.Is(err, service.ErrInvalidWorkoutExerciseInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "workout_session_id, exercise_id, notes and exercise_order are required",
			})
			return
		}

		slog.Error("failed to add exercise to workout session", "error", err, "workout_id", workoutID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to add exercise to workout session",
		})
		return
	}

	c.JSON(http.StatusCreated, workoutExercise)
}

// GetExercisesByWorkoutID handles the HTTP Request for getting the Exercises using the
// associated Workout Session ID via the API Route /api/workout/:id/exercises
func (h *WorkoutHandler) GetExercisesByWorkoutID(c *gin.Context) {
	parsedWorkoutID, ok := ParseUUID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
		return
	}

	exercises, err := h.service.GetWorkoutExercisesBySessionID(c.Request.Context(), parsedWorkoutID)
	if handleWorkoutError(c, err, "failed to retrieve exercises for workout session", "workout_id", c.Param("id")) {
		return
	}
	c.JSON(http.StatusOK, exercises)
}

/* Workout Set */

// CreateWorkoutSet handles the HTTP Request for creating a new Workout Set using the
// workout session ID and the workout exercise ID associated with the set via the
// API Route /api/workout/:id/exercices/:exercise_id/set
func (h *WorkoutHandler) CreateWorkoutSet(c *gin.Context) {
	workoutID := c.Param("id")
	_, okSession := ParseUUID(workoutID)
	exerciseID := c.Param("exercise_id")
	parsedExerciseID, okExercise := ParseUUID(exerciseID)
	if !okSession {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workout id is required",
		})
		c.Abort()
	}
	if !okExercise {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "exercise_id is required",
		})
		c.Abort()
	}

	var req createSetToExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	set := &model.WorkoutSet{
		WorkoutExerciseID: parsedExerciseID,
		SetNumber:         req.SetNumber,
		Repetitions:       req.Repetitions,
		WeightKg:          req.WeightKg,
		Duration:          req.Duration,
		DistanceKm:        req.DistanceKm,
		Rir:               req.Rir,
		Completed:         req.Completed,
	}

	if err := h.service.CreateWorkoutSet(c.Request.Context(), set); err != nil {
		if errors.Is(err, service.ErrInvalidWorkoutSetInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "workout_exercise_id and set_number are required, completed must be provided",
			})
			return
		}
		slog.Error("failed to add set to workout session", "error", err, "workout_id", workoutID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to add set to workout session",
		})
		return
	}
	c.JSON(http.StatusCreated, set)
}

// GetWorkoutSetsByExerciseID handles the HTTP Request for getting all the exercise sets
// associated with the workout exercise ID via the
// API Route /api/workout/:id/exercise/:id/sets
func (h *WorkoutHandler) GetWorkoutSetsByExerciseID(c *gin.Context) {
	workoutID := c.Param("id")
	_, okSession := ParseUUID(workoutID)
	exerciseID := c.Param("exercise_id")
	parsedExerciseID, okExercise := ParseUUID(exerciseID)
	if !okSession {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid workout id",
		})
		c.Abort()
	}
	if !okExercise {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid exercise id",
		})
		c.Abort()
	}

	sets, err := h.service.GetWorkoutSetsByWorkoutExerciseID(c.Request.Context(), parsedExerciseID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidWorkoutSetInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid exercise id",
			})
			return
		}
		if errors.Is(err, service.ErrWorkoutSetNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "workout sets not found",
			})
			return
		}
		slog.Error("failed to retrieve sets for workout exercise", "error", err, "workout_id", workoutID, "exercise_id", exerciseID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve sets for workout exercise",
		})
		return
	}
	c.JSON(http.StatusOK, sets)
}

// UpdateWorkoutSet handles the HTTP Request for updating an existing workout set using
// the ID associated via the API Route /api/workout/:id/exercices/:exercise_id/sets/:set_id
func (h *WorkoutHandler) UpdateWorkoutSet(c *gin.Context) {
	workoutID := c.Param("id")
	_, okSession := ParseUUID(workoutID)
	exerciseID := c.Param("exercise_id")
	parsedExerciseID, okExercise := ParseUUID(exerciseID)
	setID := c.Param("set_id")
	parsedSetID, okSet := ParseUUID(setID)
	if !okSession {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid workout id",
		})
		c.Abort()
	}
	if !okExercise {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid exercise id",
		})
		c.Abort()
	}
	if !okSet {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid set id",
		})
		c.Abort()
	}

	var req createSetToExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	updatedSet := &model.WorkoutSet{
		WorkoutExerciseID: parsedExerciseID,
		SetNumber:         req.SetNumber,
		Repetitions:       req.Repetitions,
		WeightKg:          req.WeightKg,
		Duration:          req.Duration,
		DistanceKm:        req.DistanceKm,
		Rir:               req.Rir,
		Completed:         req.Completed,
	}

	if err := h.service.UpdateWorkoutSet(c.Request.Context(), parsedSetID, updatedSet); err != nil {
		if errors.Is(err, service.ErrInvalidWorkoutSetInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "workout_exercise_id and set_number are required, completed must be provided",
			})
			return
		}
		if errors.Is(err, service.ErrWorkoutSetNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "workout set not found",
			})
		}
		slog.Error("failed to update set to workout set", "error", err, "workout_id", workoutID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update set to workout set",
		})
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}

// ParseUUID helps to parse an UUID string and returns the parsed UUID and a
// boolean indicating if the parsing was successful.
func ParseUUID(uuidStr string) (uuid.UUID, bool) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil || parsedUUID == uuid.Nil {
		return uuid.Nil, false
	}
	return parsedUUID, true
}

// TimePointer helps to parse a Time.time to a pointer to use the nil value if needed
func TimePointer(t time.Time) *time.Time {
	return &t
}

// handleWorkoutError writes the appropriate HTTP response for known service errors.
// Returns true if an error was handled (so the caller can return early).
func handleWorkoutError(c *gin.Context, err error, logMsg string, logArgs ...any) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, service.ErrInvalidWorkoutSessionInput),
		errors.Is(err, service.ErrInvalidWorkoutExerciseInput),
		errors.Is(err, service.ErrInvalidWorkoutSetInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrWorkoutNotFound),
		errors.Is(err, service.ErrWorkoutExerciseNotFound),
		errors.Is(err, service.ErrWorkoutSetNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		slog.Error(logMsg, append([]any{"error", err}, logArgs...)...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": logMsg})
	}
	return true
}
