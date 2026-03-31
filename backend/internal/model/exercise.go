package model

import "time"

type Exercise struct {
	ID                   int       `json:"id"`
	Name                 string    `json:"name"`
	Description          string    `json:"description,omitempty"`
	MuscleGroup          string    `json:"muscle_group"`
	SecondaryMuscleGroup string    `json:"secondary_muscle_group,omitempty"`
	ExerciseType         string    `json:"exercise_type,omitempty"`
	IsOfficial           bool      `json:"is_official"`
	CreatedAt            time.Time `json:"created_at"`
}
