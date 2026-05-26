package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProfileRepository defines methods for profile data and stats.
type ProfileRepository interface {
	GetStats(ctx context.Context, userID string, timeRange string, year int, month int) (*model.ProfileStats, error)
	UpsertGoals(ctx context.Context, goal *model.UserGoal) error
	AddBodyMetric(ctx context.Context, userID string, req *model.AddBodyMetricRequest) error
}

type profileRepository struct {
	db *pgxpool.Pool
}

// NewProfileRepository creates a new ProfileRepository.
func NewProfileRepository(db *pgxpool.Pool) ProfileRepository {
	return &profileRepository{db: db}
}

// GetStats aggregates all the user's data for the profile dashboard.
//nolint:gocognit,gocyclo
func (r *profileRepository) GetStats(ctx context.Context, userID string, timeRange string, year int, month int) (*model.ProfileStats, error) {
	stats := &model.ProfileStats{
		TopExercises:  make([]model.ExerciseStat, 0),
		MuscleRadar:   make([]model.MuscleRadarStat, 0),
		WeightHistory: make([]model.BodyMetric, 0),
		StreakDays:    make([]string, 0),
	}

	// Dynamic date filters
	dateFilterSes := ""
	dateFilterWeight := ""
	if timeRange == "month" && year > 0 && month > 0 {
		dateFilterSes = fmt.Sprintf(
			" AND DATE_TRUNC('month', ses.performed_at) = DATE_TRUNC('month', make_date(%d, %d, 1))",
			year, month,
		)
		dateFilterWeight = fmt.Sprintf(
			" AND DATE_TRUNC('month', recorded_at) = DATE_TRUNC('month', make_date(%d, %d, 1))",
			year, month,
		)
	}

	// 1. Summary stats
	summaryQuery := `
		SELECT 
			COUNT(DISTINCT ses.id) as total_workouts,
			COALESCE(SUM(ses.duration_minutes), 0) as total_duration,
			COUNT(ws.id) as total_sets,
			COALESCE(SUM(ws.reps * ws.weight_kg), 0) as total_volume
		FROM workout_sessions ses
		LEFT JOIN workout_exercises we ON ses.id = we.workout_session_id
		LEFT JOIN workout_sets ws ON we.id = ws.workout_exercise_id AND ws.completed = true
		WHERE ses.user_id = $1::uuid` + dateFilterSes
	err := r.db.QueryRow(ctx, summaryQuery, userID).Scan(
		&stats.TotalWorkouts, &stats.TotalDuration, &stats.TotalSets, &stats.TotalVolume,
	)
	if err != nil {
		fmt.Printf("Error scanning summary: %v\n", err)
	}

	// 2. Top exercises
	topQuery := `
		SELECT e.name, COUNT(ws.id) as sets
		FROM workout_sets ws
		JOIN workout_exercises we ON ws.workout_exercise_id = we.id
		JOIN exercises e ON we.exercise_id = e.id
		JOIN workout_sessions ses ON we.workout_session_id = ses.id
		WHERE ses.user_id = $1::uuid AND ws.completed = true` + dateFilterSes + `
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

	// 3. Muscle radar
	radarQuery := `
		SELECT e.muscle_group, COUNT(ws.id) as value
		FROM workout_sets ws
		JOIN workout_exercises we ON ws.workout_exercise_id = we.id
		JOIN exercises e ON we.exercise_id = e.id
		JOIN workout_sessions ses ON we.workout_session_id = ses.id
		WHERE ses.user_id = $1::uuid AND ws.completed = true` + dateFilterSes + `
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

	// 4. Weight history for the chart (with height from user_profiles)
	weightQuery := `
		SELECT bm.id, bm.weight_kg, bm.body_fat_percentage, bm.recorded_at, up.height_cm
		FROM body_metrics bm
		LEFT JOIN user_profiles up ON bm.user_id = up.user_id
		WHERE bm.user_id = $1::uuid` + dateFilterWeight + `
		ORDER BY bm.recorded_at ASC
	`
	rows3, _ := r.db.Query(ctx, weightQuery, userID)
	for rows3.Next() {
		var w model.BodyMetric
		if err := rows3.Scan(&w.ID, &w.WeightKg, &w.BodyFatPercentage, &w.RecordedAt, &w.HeightCm); err == nil {
			stats.WeightHistory = append(stats.WeightHistory, w)
		}
	}
	rows3.Close()

	// 5. Streak days with extra info for the calendar
	streakQuery := `
		SELECT 
			TO_CHAR(COALESCE(performed_at, planned_at), 'YYYY-MM-DD') as date_str,
			COALESCE(name, 'Entrenamiento Libre') as name,
			COALESCE(duration_minutes, 0) as duration,
			(performed_at IS NULL AND planned_at IS NOT NULL) as is_planned
		FROM workout_sessions 
		WHERE user_id = $1::uuid AND (performed_at IS NOT NULL OR planned_at IS NOT NULL)
	`
	rows4, _ := r.db.Query(ctx, streakQuery, userID)
	for rows4.Next() {
		var dateStr, name string
		var duration int
		var isPlanned bool
		if err := rows4.Scan(&dateStr, &name, &duration, &isPlanned); err == nil {
			stats.StreakDays = append(stats.StreakDays, dateStr)
			stats.StreakActivities = append(stats.StreakActivities, model.CalendarActivity{
				Date:        dateStr,
				WorkoutName: name,
				Duration:    duration,
				IsPlanned:   isPlanned,
			})
		}
	}
	rows4.Close()

	// 6. Goals
	goalQuery := `SELECT short_term, long_term, target_days_per_week FROM user_goals WHERE user_id = $1::uuid`
	var goals model.UserGoal
	err = r.db.QueryRow(ctx, goalQuery, userID).Scan(&goals.ShortTerm, &goals.LongTerm, &goals.TargetDaysPerWeek)
	if err == nil {
		stats.Goals = &goals
	}

	// 7. Weekly streak: how many consecutive weeks the user met their goal
	weeklyGoal := 2 // default goal
	if stats.Goals != nil && stats.Goals.TargetDaysPerWeek > 0 {
		weeklyGoal = stats.Goals.TargetDaysPerWeek
	}
	streakRangeStart := time.Now().AddDate(0, 0, -90)
	weeklyStreakQuery := `
		SELECT performed_at
		FROM workout_sessions
		WHERE user_id = $1::uuid
		  AND performed_at IS NOT NULL
		  AND performed_at >= $2
		ORDER BY performed_at DESC
	`
	var streakDates []time.Time
	rows5, _ := r.db.Query(ctx, weeklyStreakQuery, userID, streakRangeStart)
	for rows5.Next() {
		var t time.Time
		if rows5.Scan(&t) == nil {
			streakDates = append(streakDates, t)
		}
	}
	rows5.Close()

	stats.WeeklyGoal = weeklyGoal
	stats.WeeklyStreak = profileCalcWeeklyStreak(time.Now(), streakDates, weeklyGoal)

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

// AddBodyMetric inserts a new body metric record and updates the user profile height if provided.
func (r *profileRepository) AddBodyMetric(ctx context.Context, userID string, req *model.AddBodyMetricRequest) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Insert into body_metrics
	metricQuery := `
		INSERT INTO body_metrics (user_id, weight_kg, body_fat_percentage, recorded_at)
		VALUES ($1::uuid, $2, $3, CURRENT_TIMESTAMP)
	`
	_, err = tx.Exec(ctx, metricQuery, userID, req.WeightKg, req.BodyFatPercentage)
	if err != nil {
		return err
	}

	// Upsert user_profiles (to update weight and possibly height)
	profileQuery := `
		INSERT INTO user_profiles (user_id, weight_kg, height_cm, updated_at)
		VALUES ($1::uuid, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO UPDATE 
		SET weight_kg = EXCLUDED.weight_kg,
			height_cm = COALESCE(EXCLUDED.height_cm, user_profiles.height_cm),
			updated_at = CURRENT_TIMESTAMP
	`
	_, err = tx.Exec(ctx, profileQuery, userID, req.WeightKg, req.HeightCm)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// profileCalcWeeklyStreak counts how many consecutive weeks (going backwards from now)
// the user met their weekly training goal. Uses Monday as week start.
func profileCalcWeeklyStreak(referenceDate time.Time, dates []time.Time, weeklyGoal int) int {
	if len(dates) == 0 || weeklyGoal <= 0 {
		return 0
	}

	weekCounts := make(map[string]int)
	for _, d := range dates {
		ws := profileStartOfWeek(d)
		weekCounts[ws.Format("2006-01-02")]++
	}

	currentWeek := profileStartOfWeek(referenceDate)
	// If current week not yet fulfilled, start checking from the previous one
	if weekCounts[currentWeek.Format("2006-01-02")] < weeklyGoal {
		currentWeek = currentWeek.AddDate(0, 0, -7)
	}

	streak := 0
	for w := currentWeek; ; w = w.AddDate(0, 0, -7) {
		if weekCounts[w.Format("2006-01-02")] < weeklyGoal {
			break
		}
		streak++
	}
	return streak
}

func profileStartOfWeek(t time.Time) time.Time {
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	offset := (int(day.Weekday()) + 6) % 7 // Monday = 0
	return day.AddDate(0, 0, -offset)
}
