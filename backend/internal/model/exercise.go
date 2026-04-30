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
