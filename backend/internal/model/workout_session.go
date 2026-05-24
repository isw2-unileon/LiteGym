package model

import (
	"time"

	"github.com/google/uuid"
)

// WorkoutSession represents a user's workout session,
// including planning, completion, duration, and calories burned.
type WorkoutSession struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	RoutineID      *uuid.UUID `json:"routine_id,omitempty"`
	Name           string     `json:"name"`
	PerformedAt    *time.Time `json:"performed_at,omitempty"`
	PlannedAt      *time.Time `json:"planned_at,omitempty"`
	Duration       *int       `json:"duration_minutes,omitempty"`
	CaloriesBurned *float64   `json:"calories_burned,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}
