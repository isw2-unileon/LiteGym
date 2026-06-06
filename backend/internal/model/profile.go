package model

import (
	"time"

	"github.com/google/uuid"
)

// BodyMetric represents a single body measurement record.
type BodyMetric struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	WeightKg          float64   `json:"weight_kg"`
	HeightCm          *float64  `json:"height_cm,omitempty"`
	BodyFatPercentage *float64  `json:"body_fat_percentage,omitempty"`
	RecordedAt        time.Time `json:"recorded_at"`
}

// AddBodyMetricRequest is the payload to add a new body metric.
type AddBodyMetricRequest struct {
	WeightKg          float64  `json:"weight_kg" binding:"required,gt=0"`
	BodyFatPercentage *float64 `json:"body_fat_percentage,omitempty" binding:"omitempty,gte=0,lte=100"`
	HeightCm          *float64 `json:"height_cm,omitempty" binding:"omitempty,gt=0"`
}

// UserGoal represents the user's fitness goals.
type UserGoal struct {
	UserID            uuid.UUID `json:"-"`
	ShortTerm         string    `json:"short_term"`
	LongTerm          string    `json:"long_term"`
	TargetDaysPerWeek int       `json:"target_days"`
}

// ExerciseStat represents the frequency of an exercise.
type ExerciseStat struct {
	Name string `json:"name"`
	Sets int    `json:"sets"`
}

// MuscleRadarStat represents the distribution of work across muscle groups.
type MuscleRadarStat struct {
	Muscle string `json:"muscle"`
	Value  int    `json:"value"`
}

// CalendarActivity represents an activity in the user's calendar.
type CalendarActivity struct {
	ID            string `json:"id"`
	Date          string `json:"date"`
	WorkoutName   string `json:"workout_name"`
	Duration      int    `json:"duration"`
	ExerciseCount int    `json:"exercise_count"`
	IsPlanned     bool   `json:"is_planned"`
}

// ProfileStats representa los datos del dashboard
type ProfileStats struct {
	TotalWorkouts    int                `json:"total_workouts"`
	TotalDuration    int                `json:"total_duration_minutes"`
	TotalVolume      float64            `json:"total_volume_kg"`
	TotalSets        int                `json:"total_sets"`
	WeeklyStreak     int                `json:"weekly_streak"`
	WeeklyGoal       int                `json:"weekly_goal"`
	StreakDays       []string           `json:"streak_days"`
	StreakActivities []CalendarActivity `json:"streak_activities"`
	TopExercises     []ExerciseStat     `json:"top_exercises"`
	MuscleRadar      []MuscleRadarStat  `json:"muscle_radar"`
	WeightHistory    []BodyMetric       `json:"weight_history"`
	Goals            *UserGoal          `json:"goals"`
}
