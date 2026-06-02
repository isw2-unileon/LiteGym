package model

import "time"

// Exercise represents an exercise entity in the system.
type Exercise struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Description          string    `json:"description,omitempty"`
	MuscleGroup          string    `json:"muscle_group"`
	SecondaryMuscleGroup string    `json:"secondary_muscle_group,omitempty"`
	ExerciseType         string    `json:"exercise_type,omitempty"`
	IsOfficial           bool      `json:"is_official"`
	OwnerUserID          *string   `json:"owner_user_id,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}

// ExerciseFilter contains the supported filters for listing exercises.
type ExerciseFilter struct {
	Search      string
	Type        string
	MuscleGroup string
	Official    *bool
	Page        int
	Limit       int
}

// ExerciseListResponse represents a paginated exercise list response.
type ExerciseListResponse struct {
	Items      []Exercise `json:"items"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	Total      int        `json:"total"`
	TotalPages int        `json:"total_pages"`
}

// ExerciseWorkoutSessionSummary represents a workout session where an exercise was performed.
type ExerciseWorkoutSessionSummary struct {
	ID              string    `json:"id"`
	Name            string    `json:"name,omitempty"`
	RoutineName     string    `json:"routine_name,omitempty"`
	StartedAt       time.Time `json:"started_at"`
	DurationMinutes int       `json:"duration_minutes"`
	ExerciseOrder   int       `json:"exercise_order"`
	SetCount        int       `json:"set_count"`
}

// ExerciseInsights contains historical performance data for an exercise.
type ExerciseInsights struct {
	Summary         ExerciseInsightSummary          `json:"summary"`
	BestSet         *ExerciseInsightSet             `json:"best_set,omitempty"`
	PersonalRecords []ExercisePersonalRecord        `json:"personal_records"`
	Progression     []ExerciseProgressPoint         `json:"progression"`
	History         []ExerciseInsightSessionHistory `json:"history"`
}

// ExerciseInsightSummary contains aggregated exercise performance metrics.
type ExerciseInsightSummary struct {
	SessionCount       int        `json:"session_count"`
	SetCount           int        `json:"set_count"`
	TotalVolumeKg      float64    `json:"total_volume_kg"`
	MaxWeightKg        *float64   `json:"max_weight_kg,omitempty"`
	MaxReps            *int       `json:"max_reps,omitempty"`
	LastPerformedAt    *time.Time `json:"last_performed_at,omitempty"`
	FirstPerformedAt   *time.Time `json:"first_performed_at,omitempty"`
	AverageDaysBetween *float64   `json:"average_days_between,omitempty"`
	Trend              string     `json:"trend"`
}

// ExerciseInsightSet describes a notable set for the selected exercise.
type ExerciseInsightSet struct {
	SessionID   string    `json:"session_id"`
	SessionName string    `json:"session_name,omitempty"`
	RoutineName string    `json:"routine_name,omitempty"`
	PerformedAt time.Time `json:"performed_at"`
	SetNumber   int       `json:"set_number"`
	Reps        *int      `json:"reps,omitempty"`
	WeightKg    *float64  `json:"weight_kg,omitempty"`
	VolumeKg    float64   `json:"volume_kg"`
	Rir         *int      `json:"rir,omitempty"`
	Completed   bool      `json:"completed"`
}

// ExercisePersonalRecord describes one personal record dimension.
type ExercisePersonalRecord struct {
	Type        string    `json:"type"`
	Label       string    `json:"label"`
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	PerformedAt time.Time `json:"performed_at"`
}

// ExerciseProgressPoint stores per-session progression for charting.
type ExerciseProgressPoint struct {
	SessionID   string    `json:"session_id"`
	Date        time.Time `json:"date"`
	MaxWeightKg *float64  `json:"max_weight_kg,omitempty"`
	MaxReps     *int      `json:"max_reps,omitempty"`
	VolumeKg    float64   `json:"volume_kg"`
	SetCount    int       `json:"set_count"`
}

// ExerciseInsightSessionHistory stores detailed set history grouped by session.
type ExerciseInsightSessionHistory struct {
	SessionID       string               `json:"session_id"`
	SessionName     string               `json:"session_name,omitempty"`
	RoutineName     string               `json:"routine_name,omitempty"`
	PerformedAt     time.Time            `json:"performed_at"`
	DurationMinutes int                  `json:"duration_minutes"`
	VolumeKg        float64              `json:"volume_kg"`
	Sets            []ExerciseInsightSet `json:"sets"`
}

// SelectOption represents one selectable value exposed to clients.
type SelectOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ExerciseMetadataResponse contains the valid exercise domain options.
type ExerciseMetadataResponse struct {
	ExerciseTypes []SelectOption `json:"exercise_types"`
	MuscleGroups  []SelectOption `json:"muscle_groups"`
}
