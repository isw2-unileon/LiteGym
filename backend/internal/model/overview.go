package model

import "time"

// Overview contains the data required to render the authenticated dashboard.
type Overview struct {
	Calendar           OverviewCalendar           `json:"calendar"`
	RecentRoutines     []OverviewRoutineSummary   `json:"recent_routines"`
	RecentWorkouts     []OverviewWorkoutSummary   `json:"recent_workouts"`
	Progress           OverviewProgressSummary    `json:"progress"`
	MuscleDistribution OverviewMuscleDistribution `json:"muscle_distribution"`
}

// OverviewCalendar summarizes the current month's training activity.
type OverviewCalendar struct {
	Month            string                    `json:"month"`
	TrainedDays      []string                  `json:"trained_days"`
	PlannedDays      []string                  `json:"planned_days"`
	CalendarWorkouts []OverviewCalendarWorkout `json:"calendar_workouts"`
	SessionsCount    int                       `json:"sessions_count"`
	CurrentStreak    int                       `json:"current_streak"`
	WeeklyGoal       int                       `json:"weekly_goal"`
	NextObjective    string                    `json:"next_objective"`
}

// OverviewRoutineSummary represents a lightweight routine card in the overview.
type OverviewRoutineSummary struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	ExerciseCount int       `json:"exercise_count"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// OverviewWorkoutSummary represents a lightweight training session card in the overview.
type OverviewWorkoutSummary struct {
	ID              string    `json:"id"`
	Name            string    `json:"name,omitempty"`
	RoutineName     string    `json:"routine_name,omitempty"`
	PerformedAt     time.Time `json:"performed_at"`
	DurationMinutes int       `json:"duration_minutes"`
	ExerciseCount   int       `json:"exercise_count"`
}

// OverviewCalendarWorkout represents a workout displayed in the monthly calendar.
type OverviewCalendarWorkout struct {
	ID              string     `json:"id"`
	Name            string     `json:"name,omitempty"`
	RoutineID       string     `json:"routine_id,omitempty"`
	RoutineName     string     `json:"routine_name,omitempty"`
	PerformedAt     *time.Time `json:"performed_at,omitempty"`
	PlannedAt       *time.Time `json:"planned_at,omitempty"`
	DurationMinutes int        `json:"duration_minutes"`
	ExerciseCount   int        `json:"exercise_count"`
}

// OverviewProgressMetric stores the current and previous values for one body metric.
type OverviewProgressMetric struct {
	Current  *float64 `json:"current,omitempty"`
	Previous *float64 `json:"previous,omitempty"`
	Delta    *float64 `json:"delta,omitempty"`
}

// OverviewProgressSummary contains the latest body metrics used in the progress panel.
type OverviewProgressSummary struct {
	LastRecordedAt    *time.Time             `json:"last_recorded_at,omitempty"`
	WeightKg          OverviewProgressMetric `json:"weight_kg"`
	BodyFatPercentage OverviewProgressMetric `json:"body_fat_percentage"`
	MuscleMassKg      OverviewProgressMetric `json:"muscle_mass_kg"`
}

// OverviewMuscleGroupShare describes how often a muscle group appears in the selected range.
type OverviewMuscleGroupShare struct {
	Name       string `json:"name"`
	Count      int    `json:"count"`
	Percentage int    `json:"percentage"`
}

// OverviewMuscleDistribution contains both yearly and monthly muscle distributions.
type OverviewMuscleDistribution struct {
	Year               []OverviewMuscleGroupShare `json:"year"`
	Month              []OverviewMuscleGroupShare `json:"month"`
	YearExerciseCount  int                        `json:"year_exercise_count"`
	MonthExerciseCount int                        `json:"month_exercise_count"`
}

// OverviewBodyMetricEntry represents one persisted body metrics snapshot.
type OverviewBodyMetricEntry struct {
	RecordedAt        time.Time
	WeightKg          *float64
	BodyFatPercentage *float64
	MuscleMassKg      *float64
}
