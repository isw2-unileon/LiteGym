package model

import "time"

// AIRoutineGenerationRequest defines pre-generation filters selected by user.
type AIRoutineGenerationRequest struct {
	Objective            string   `json:"objective"`
	TargetMuscleGroups   []string `json:"target_muscle_groups"`
	MandatoryExerciseIDs []string `json:"mandatory_exercise_ids"`
	DurationMinutes      int      `json:"duration_minutes"`
}

// AIRoutineRecentWorkoutSet represents one set from a recent workout session.
type AIRoutineRecentWorkoutSet struct {
	SetNumber int      `json:"set_number"`
	Reps      *int     `json:"reps,omitempty"`
	WeightKg  *float64 `json:"weight_kg,omitempty"`
}

// AIRoutineRecentWorkoutExercise represents one exercise inside a recent workout session.
type AIRoutineRecentWorkoutExercise struct {
	ExerciseID    string                      `json:"exercise_id"`
	ExerciseName  string                      `json:"exercise_name"`
	MuscleGroup   string                      `json:"muscle_group"`
	ExerciseType  string                      `json:"exercise_type,omitempty"`
	ExerciseOrder int                         `json:"exercise_order"`
	Sets          []AIRoutineRecentWorkoutSet `json:"sets"`
}

// AIRoutineRecentWorkoutSession represents one recent workout session with sets.
type AIRoutineRecentWorkoutSession struct {
	SessionID       string                           `json:"session_id"`
	SessionName     string                           `json:"session_name,omitempty"`
	RoutineName     string                           `json:"routine_name,omitempty"`
	StartedAt       time.Time                        `json:"started_at"`
	DurationMinutes int                              `json:"duration_minutes"`
	Exercises       []AIRoutineRecentWorkoutExercise `json:"exercises"`
}

// AIRoutineExercise defines one exercise entry in generated routine JSON.
type AIRoutineExercise struct {
	ExerciseID      string `json:"exercise_id"`
	Name            string `json:"name"`
	MuscleGroup     string `json:"muscle_group"`
	ExerciseType    string `json:"exercise_type,omitempty"`
	IsMandatory     bool   `json:"is_mandatory"`
	RecommendedSets int    `json:"recommended_sets"`
	RecommendedReps string `json:"recommended_reps"`
}

// AIRoutineJSON is the AI output format returned to client.
type AIRoutineJSON struct {
	Name             string              `json:"name"`
	Objective        string              `json:"objective"`
	DurationMinutes  int                 `json:"duration_minutes"`
	TargetMuscles    []string            `json:"target_muscles"`
	MandatoryCount   int                 `json:"mandatory_count"`
	GeneratedAt      time.Time           `json:"generated_at"`
	Exercises        []AIRoutineExercise `json:"exercises"`
	GenerationSource string              `json:"generation_source"`
}

// AIRoutineRateLimitStatus returns usage information for frontend.
type AIRoutineRateLimitStatus struct {
	Limit               int       `json:"limit"`
	Remaining           int       `json:"remaining"`
	UsedInCurrentWindow int       `json:"used_in_current_window"`
	WindowSeconds       int       `json:"window_seconds"`
	ResetAt             time.Time `json:"reset_at"`
}

// AIRoutineGenerateResponse wraps generated JSON plus rate-limit metadata.
type AIRoutineGenerateResponse struct {
	RoutineJSON AIRoutineJSON            `json:"routine_json"`
	RateLimit   AIRoutineRateLimitStatus `json:"rate_limit"`
}
