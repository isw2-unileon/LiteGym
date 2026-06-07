package model

// WorkoutFinishInput contains the editable fields required to save a completed workout.
type WorkoutFinishInput struct {
	Name           string   `json:"name"`
	Duration       *int     `json:"duration_minutes,omitempty"`
	CaloriesBurned *float64 `json:"calories_burned,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
}
