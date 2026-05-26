package model

import (
	"time"

	"github.com/google/uuid"
)

// WorkoutSet represents a single set of an exercise performed during a workout session.
// It includes details such as the number of repetitions, weight lifted, duration,
// distance covered, and the RIR (Repetitions In Reserve) value.
// Each set is associated with a specific workout exercise and has a unique identifier.
type WorkoutSet struct {
	ID                uuid.UUID    `json:"id"`
	WorkoutExerciseID uuid.UUID    `json:"workout_exercise_id"`
	SetNumber         int       `json:"set_number"`
	Repetitions       *int      `json:"reps,omitempty"`
	WeightKg          *float64  `json:"weight_kg,omitempty"`
	Duration          *int      `json:"duration_seconds,omitempty"`
	DistanceKm        *float64  `json:"distance_km,omitempty"`
	Rir               *int      `json:"rir,omitempty"`
	Completed         *bool      `json:"completed"`
	CreatedAt         time.Time `json:"created_at"`
}
