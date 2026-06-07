package model

import "time"

// WorkoutSessionDetail represents a workout session with its exercises and sets.
type WorkoutSessionDetail struct {
	ID             string                         `json:"id"`
	UserID         string                         `json:"user_id"`
	RoutineID      string                         `json:"routine_id,omitempty"`
	Name           string                         `json:"name"`
	PerformedAt    *time.Time                     `json:"performed_at,omitempty"`
	PlannedAt      *time.Time                     `json:"planned_at,omitempty"`
	Duration       *int                           `json:"duration_minutes,omitempty"`
	CaloriesBurned *float64                       `json:"calories_burned,omitempty"`
	Notes          *string                        `json:"notes,omitempty"`
	CreatedAt      time.Time                      `json:"created_at"`
	Exercises      []WorkoutSessionExerciseDetail `json:"exercises"`
}

// WorkoutSessionExerciseDetail represents one workout exercise with display metadata.
type WorkoutSessionExerciseDetail struct {
	ID                   string                    `json:"id"`
	WorkoutSessionID     string                    `json:"workout_session_id"`
	ExerciseID           string                    `json:"exercise_id"`
	RoutineExerciseID    string                    `json:"routine_exercise_id,omitempty"`
	Name                 string                    `json:"name"`
	Description          string                    `json:"description,omitempty"`
	MuscleGroup          string                    `json:"muscle_group"`
	SecondaryMuscleGroup string                    `json:"secondary_muscle_group,omitempty"`
	ExerciseType         string                    `json:"exercise_type,omitempty"`
	ExerciseOrder        int                       `json:"exercise_order"`
	Notes                string                    `json:"notes,omitempty"`
	Sets                 []WorkoutSessionSetDetail `json:"sets"`
}

// WorkoutSessionSetDetail represents one editable set in a workout session.
type WorkoutSessionSetDetail struct {
	ID                    string    `json:"id"`
	RoutineExerciseSetID  string    `json:"routine_exercise_set_id,omitempty"`
	SetNumber             int       `json:"set_number"`
	TargetRepsMin         *int      `json:"target_reps_min,omitempty"`
	TargetRepsMax         *int      `json:"target_reps_max,omitempty"`
	TargetRepsText        string    `json:"target_reps_text,omitempty"`
	TargetWeightKg        *float64  `json:"target_weight_kg,omitempty"`
	TargetDurationSeconds *int      `json:"target_duration_seconds,omitempty"`
	TargetDistanceKm      *float64  `json:"target_distance_km,omitempty"`
	TargetRir             *int      `json:"target_rir,omitempty"`
	RestSeconds           *int      `json:"rest_seconds,omitempty"`
	Repetitions           *int      `json:"reps,omitempty"`
	WeightKg              *float64  `json:"weight_kg,omitempty"`
	Duration              *int      `json:"duration_seconds,omitempty"`
	DistanceKm            *float64  `json:"distance_km,omitempty"`
	Rir                   *int      `json:"rir,omitempty"`
	Completed             *bool     `json:"completed"`
	CreatedAt             time.Time `json:"created_at"`
}
