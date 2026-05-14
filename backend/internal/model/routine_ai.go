package model

import "time"

// AIRoutineGenerationRequest defines pre-generation filters selected by user.
type AIRoutineGenerationRequest struct {
	Objective            string   `json:"objective"`
	TargetMuscleGroups   []string `json:"target_muscle_groups"`
	MandatoryExerciseIDs []string `json:"mandatory_exercise_ids"`
	DurationMinutes      int      `json:"duration_minutes"`
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
	RoutineJSON AIRoutineJSON         `json:"routine_json"`
	RateLimit   AIRoutineRateLimitStatus `json:"rate_limit"`
}

