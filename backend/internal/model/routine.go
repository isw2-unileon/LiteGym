package model

import "time"

type RoutineDetail struct {
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	Description   string                  `json:"description,omitempty"`
	Source        string                  `json:"source"`
	ExerciseCount int                     `json:"exercise_count"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
	Exercises     []RoutineExerciseDetail `json:"exercises"`
}

type RoutineExerciseDetail struct {
	ID                   string                     `json:"id"`
	ExerciseID           string                     `json:"exercise_id"`
	Name                 string                     `json:"name"`
	Description          string                     `json:"description,omitempty"`
	MuscleGroup          string                     `json:"muscle_group"`
	SecondaryMuscleGroup string                     `json:"secondary_muscle_group,omitempty"`
	ExerciseType         string                     `json:"exercise_type,omitempty"`
	ExerciseOrder        int                        `json:"exercise_order"`
	Notes                string                     `json:"notes,omitempty"`
	Sets                 []RoutineExerciseSetDetail `json:"sets"`
}

type RoutineExerciseSetDetail struct {
	ID                    string   `json:"id"`
	SetNumber             int      `json:"set_number"`
	TargetRepsMin         *int     `json:"target_reps_min,omitempty"`
	TargetRepsMax         *int     `json:"target_reps_max,omitempty"`
	TargetRepsText        string   `json:"target_reps_text,omitempty"`
	TargetWeightKg        *float64 `json:"target_weight_kg,omitempty"`
	TargetDurationSeconds *int     `json:"target_duration_seconds,omitempty"`
	TargetDistanceKm      *float64 `json:"target_distance_km,omitempty"`
	TargetRir             *int     `json:"target_rir,omitempty"`
	RestSeconds           *int     `json:"rest_seconds,omitempty"`
	Notes                 string   `json:"notes,omitempty"`
}
