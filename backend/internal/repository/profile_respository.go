package repository

import (
	"context"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProfileRepository defines methods for profile data and stats.
type ProfileRepository interface {
	GetStats(ctx context.Context, userID string) (*model.ProfileStats, error)
	UpsertGoals(ctx context.Context, goal *model.UserGoal) error
}

type profileRepository struct {
	db *pgxpool.Pool
}

// NewProfileRepository creates a new ProfileRepository.
func NewProfileRepository(db *pgxpool.Pool) ProfileRepository {
	return &profileRepository{db: db}
}

// GetStats aggregates all the user's data for the profile dashboard.
func (r *profileRepository) GetStats(ctx context.Context, userID string) (*model.ProfileStats, error) {
	stats := &model.ProfileStats{
		TopExercises:  make([]model.ExerciseStat, 0),
		MuscleRadar:   make([]model.MuscleRadarStat, 0),
		WeightHistory: make([]model.BodyMetric, 0),
		StreakDays:    make([]string, 0),
	}

	// 1. Resumen Estadístico (Workouts, Duration, Sets, Volume)
	summaryQuery := `
		SELECT 
			COUNT(DISTINCT ses.id) as total_workouts,
			COALESCE(SUM(ses.duration_minutes), 0) as total_duration,
			COUNT(ws.id) as total_sets,
			COALESCE(SUM(ws.reps * ws.weight_kg), 0) as total_volume
		FROM workout_sessions ses
		LEFT JOIN workout_exercises we ON ses.id = we.workout_session_id
		LEFT JOIN workout_sets ws ON we.id = ws.workout_exercise_id AND ws.completed = true
		WHERE ses.user_id = $1::uuid
	`
	_ = r.db.QueryRow(ctx, summaryQuery, userID).Scan(
		&stats.TotalWorkouts, &stats.TotalDuration, &stats.TotalSets, &stats.TotalVolume,
	)

	// 2. Top Ejercicios
	topQuery := `
		SELECT e.name, COUNT(ws.id) as sets
		FROM workout_sets ws
		JOIN workout_exercises we ON ws.workout_exercise_id = we.id
		JOIN exercises e ON we.exercise_id = e.id
		JOIN workout_sessions ses ON we.workout_session_id = ses.id
		WHERE ses.user_id = $1::uuid AND ws.completed = true
		GROUP BY e.name ORDER BY sets DESC LIMIT 3
	`
	rows, _ := r.db.Query(ctx, topQuery, userID)
	for rows.Next() {
		var stat model.ExerciseStat
		if err := rows.Scan(&stat.Name, &stat.Sets); err == nil {
			stats.TopExercises = append(stats.TopExercises, stat)
		}
	}
	rows.Close()

	// 3. Radar Muscular
	radarQuery := `
		SELECT e.muscle_group, COUNT(ws.id) as value
		FROM workout_sets ws
		JOIN workout_exercises we ON ws.workout_exercise_id = we.id
		JOIN exercises e ON we.exercise_id = e.id
		JOIN workout_sessions ses ON we.workout_session_id = ses.id
		WHERE ses.user_id = $1::uuid AND ws.completed = true
		GROUP BY e.muscle_group
	`
	rows2, _ := r.db.Query(ctx, radarQuery, userID)
	for rows2.Next() {
		var radar model.MuscleRadarStat
		if err := rows2.Scan(&radar.Muscle, &radar.Value); err == nil {
			stats.MuscleRadar = append(stats.MuscleRadar, radar)
		}
	}
	rows2.Close()

	// 4. Metas (Goals)
	goalQuery := `SELECT short_term, long_term, target_days_per_week FROM user_goals WHERE user_id = $1::uuid`
	var goals model.UserGoal
	err := r.db.QueryRow(ctx, goalQuery, userID).Scan(&goals.ShortTerm, &goals.LongTerm, &goals.TargetDaysPerWeek)
	if err == nil {
		stats.Goals = &goals
	}

	return stats, nil
}

// UpsertGoals updates or inserts the user's fitness goals.
func (r *profileRepository) UpsertGoals(ctx context.Context, goal *model.UserGoal) error {
	query := `
		INSERT INTO user_goals (user_id, short_term, long_term, target_days_per_week, updated_at)
		VALUES ($1::uuid, $2, $3, $4, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO UPDATE 
		SET short_term = EXCLUDED.short_term, 
		    long_term = EXCLUDED.long_term, 
		    target_days_per_week = EXCLUDED.target_days_per_week,
		    updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.Exec(ctx, query, goal.UserID, goal.ShortTerm, goal.LongTerm, goal.TargetDaysPerWeek)
	return err
}
