package model

import (
	"time"

	"github.com/google/uuid"
)

// WorkoutSession represents a user's workout session,
// including details about the exercises performed, duration, and calories burned.
type WorkoutSession struct {
	ID             uuid.UUID    `json:"id"`
	UserID         uuid.UUID   `json:"user_id"`
	RoutineID      *uuid.UUID   `json:"routine_id,omitempty"`
	Name           string    `json:"name"`
	StartedAt      time.Time `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at,omitempty"`
	Duration       *int       `json:"duration_minutes,omitempty"`
	CaloriesBurned *float64   `json:"calories_burned,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
