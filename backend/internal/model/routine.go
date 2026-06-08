package model

import "time"

// Routine represents a persisted routine with its planned exercises and sets.
type Routine struct {
	ID           string            `json:"id"`
	UserID       string            `json:"user_id"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Source       string            `json:"source"`
	IsPredefined bool              `json:"is_predefined"`
	IsPublic     bool              `json:"is_public"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Exercises    []RoutineExercise `json:"exercises"`
}

// RoutineExercise represents one exercise inside a persisted routine.
type RoutineExercise struct {
	ID                  string       `json:"id"`
	RoutineID           string       `json:"routine_id"`
	ExerciseID          string       `json:"exercise_id"`
	ExerciseName        string       `json:"exercise_name"`
	ExerciseDescription string       `json:"exercise_description,omitempty"`
	MuscleGroup         string       `json:"muscle_group"`
	ExerciseType        string       `json:"exercise_type,omitempty"`
	ExerciseOrder       int          `json:"exercise_order"`
	Notes               string       `json:"notes,omitempty"`
	Sets                []RoutineSet `json:"sets"`
}

// RoutineSet represents one planned set inside a routine exercise.
type RoutineSet struct {
	ID                    string    `json:"id"`
	RoutineExerciseID     string    `json:"routine_exercise_id"`
	SetNumber             int       `json:"set_number"`
	TargetRepsMin         *int      `json:"target_reps_min,omitempty"`
	TargetRepsMax         *int      `json:"target_reps_max,omitempty"`
	TargetRepsText        string    `json:"target_reps_text,omitempty"`
	TargetWeightKg        *float64  `json:"target_weight_kg,omitempty"`
	TargetDurationSeconds *int      `json:"target_duration_seconds,omitempty"`
	TargetDistanceKm      *float64  `json:"target_distance_km,omitempty"`
	TargetRir             *int      `json:"target_rir,omitempty"`
	RestSeconds           *int      `json:"rest_seconds,omitempty"`
	Notes                 string    `json:"notes,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
}

// RoutineDetail represents a saved routine with its exercises and planned sets.
type RoutineDetail struct {
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	Description   string                  `json:"description,omitempty"`
	Source        string                  `json:"source"`
	IsPredefined  bool                    `json:"is_predefined"`
	RoutineType   string                  `json:"routine_type"`
	ExerciseCount int                     `json:"exercise_count"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
	Exercises     []RoutineExerciseDetail `json:"exercises"`
}

// RoutineExerciseDetail represents a routine exercise in the saved routine payload.
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

// RoutineExerciseSetDetail represents a planned set for a routine exercise.
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
