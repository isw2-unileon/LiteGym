package model

import (
	"time"

	"github.com/google/uuid"
)

// WorkoutExercise represents the association between a workout and an exercise,
// including the order of the exercise in the workout and any notes.
type WorkoutExercise struct {
	ID               uuid.UUID `json:"id"`
	WorkoutSessionID uuid.UUID `json:"workout_session_id"`
	ExerciseID       uuid.UUID `json:"exercise_id"`
	ExerciseOrder    int    `json:"exercise_order"`
	Notes            string `json:"notes"`
	CreatedAt        time.Time `json:"created_at"`
}
