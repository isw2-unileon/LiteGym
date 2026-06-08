package model

import "time"

// AIRoutineGenerationRequest defines the user input used to guide AI routine generation.
type AIRoutineGenerationRequest struct {
	Objective          string   `json:"objective"`
	TargetMuscleGroups []string `json:"target_muscle_groups"`
	MandatoryExercises []string `json:"mandatory_exercises"`
	Notes              string   `json:"notes,omitempty"`
	DurationMinutes    int      `json:"duration_minutes"`
}

// AIRoutineUpgradeRequest defines the input required to improve an existing routine with AI.
type AIRoutineUpgradeRequest struct {
	Message         string `json:"message"`
	FeedbackMessage string `json:"feedback_message"`
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

// AIRoutineExerciseSet defines one planned set in a generated routine exercise.
type AIRoutineExerciseSet struct {
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

// AIRoutineExercise defines one exercise entry in generated routine JSON.
type AIRoutineExercise struct {
	ExerciseID      string                 `json:"exercise_id"`
	Name            string                 `json:"name"`
	MuscleGroup     string                 `json:"muscle_group"`
	ExerciseType    string                 `json:"exercise_type,omitempty"`
	IsMandatory     bool                   `json:"is_mandatory"`
	RecommendedSets int                    `json:"recommended_sets,omitempty"`
	RecommendedReps string                 `json:"recommended_reps,omitempty"`
	Sets            []AIRoutineExerciseSet `json:"sets,omitempty"`
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

// AIRoutineExerciseToSave contains the routine exercise data persisted from an AI generation.
type AIRoutineExerciseToSave struct {
	ExerciseID string
	Order      int
	Notes      string
	Sets       []AIRoutineExerciseSetToSave
}

// AIRoutineExerciseSetToSave contains a planned routine set persisted from an AI generation.
type AIRoutineExerciseSetToSave struct {
	SetNumber             int
	TargetRepsMin         *int
	TargetRepsMax         *int
	TargetRepsText        string
	TargetWeightKg        *float64
	TargetDurationSeconds *int
	TargetDistanceKm      *float64
	TargetRir             *int
	RestSeconds           *int
	Notes                 string
}

// AIRoutineToSave contains the routine data persisted from an AI generation.
type AIRoutineToSave struct {
	UserID      string
	Name        string
	Description string
	Exercises   []AIRoutineExerciseToSave
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
	RoutineJSON AIRoutineJSON             `json:"routine_json"`
	RoutineID   string                    `json:"routine_id,omitempty"`
	RateLimit   *AIRoutineRateLimitStatus `json:"rate_limit,omitempty"`
}

// AIRoutineUpgradeResponse wraps an upgraded routine proposal plus rate-limit metadata.
type AIRoutineUpgradeResponse struct {
	RoutineJSON AIRoutineJSON            `json:"routine_json"`
	Diff        AIRoutineUpgradeDiff     `json:"diff"`
	RateLimit   AIRoutineRateLimitStatus `json:"rate_limit"`
}

// AIRoutineUpgradeDiff represents a routine comparison prepared for a git-diff-style UI.
type AIRoutineUpgradeDiff struct {
	Summary   AIRoutineUpgradeDiffSummary  `json:"summary"`
	Exercises []AIRoutineExerciseDiffEntry `json:"exercises"`
}

// AIRoutineUpgradeDiffSummary aggregates the overall changes between routine versions.
type AIRoutineUpgradeDiffSummary struct {
	AddedExercises     int `json:"added_exercises"`
	RemovedExercises   int `json:"removed_exercises"`
	ModifiedExercises  int `json:"modified_exercises"`
	UnchangedExercises int `json:"unchanged_exercises"`
}

// AIRoutineExerciseDiffEntry describes the change for one exercise between two routine versions.
type AIRoutineExerciseDiffEntry struct {
	ChangeType         string                          `json:"change_type"`
	ExerciseID         string                          `json:"exercise_id"`
	BeforeName         string                          `json:"before_name,omitempty"`
	AfterName          string                          `json:"after_name,omitempty"`
	BeforeOrder        *int                            `json:"before_order,omitempty"`
	AfterOrder         *int                            `json:"after_order,omitempty"`
	BeforeMuscleGroup  string                          `json:"before_muscle_group,omitempty"`
	AfterMuscleGroup   string                          `json:"after_muscle_group,omitempty"`
	BeforeExerciseType string                          `json:"before_exercise_type,omitempty"`
	AfterExerciseType  string                          `json:"after_exercise_type,omitempty"`
	IsMandatoryChanged bool                            `json:"is_mandatory_changed,omitempty"`
	Sets               []AIRoutineExerciseSetDiffEntry `json:"sets,omitempty"`
}

// AIRoutineExerciseSetDiffEntry describes the change for one set in one exercise.
type AIRoutineExerciseSetDiffEntry struct {
	ChangeType string                `json:"change_type"`
	SetNumber  int                   `json:"set_number"`
	Before     *AIRoutineExerciseSet `json:"before,omitempty"`
	After      *AIRoutineExerciseSet `json:"after,omitempty"`
}
