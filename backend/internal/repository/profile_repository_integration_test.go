//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func insertProfileUser(t *testing.T, db *pgxpool.Pool, username, email string) string {
	t.Helper()
	return insertUserRaw(t, db, username, email)
}

func insertBodyMetricRaw(t *testing.T, db *pgxpool.Pool, userID string, weightKg float64, fatPct *float64, recordedAt time.Time) {
	t.Helper()
	_, err := db.Exec(context.Background(), `
		INSERT INTO public.body_metrics (user_id, weight_kg, body_fat_percentage, recorded_at)
		VALUES ($1::uuid, $2, $3, $4)
	`, userID, weightKg, fatPct, recordedAt)
	if err != nil {
		t.Fatalf("failed to insert body_metric: %v", err)
	}
}

func insertGoalRaw(t *testing.T, db *pgxpool.Pool, userID, short, long string, targetDays int) {
	t.Helper()
	_, err := db.Exec(context.Background(), `
		INSERT INTO public.user_goals (user_id, short_term, long_term, target_days_per_week, updated_at)
		VALUES ($1::uuid, $2, $3, $4, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO UPDATE
		  SET short_term = EXCLUDED.short_term,
		      long_term = EXCLUDED.long_term,
		      target_days_per_week = EXCLUDED.target_days_per_week,
		      updated_at = CURRENT_TIMESTAMP
	`, userID, short, long, targetDays)
	if err != nil {
		t.Fatalf("failed to insert user_goal: %v", err)
	}
}

// ─── profileCalcWeeklyStreak (pure logic – no DB needed) ─────────────────────

func TestProfileCalcWeeklyStreak_EmptyDates(t *testing.T) {
	streak := profileCalcWeeklyStreak(time.Now(), nil, 3)
	if streak != 0 {
		t.Errorf("expected streak=0 with empty dates, got %d", streak)
	}
}

func TestProfileCalcWeeklyStreak_ZeroGoal(t *testing.T) {
	dates := []time.Time{time.Now()}
	streak := profileCalcWeeklyStreak(time.Now(), dates, 0)
	if streak != 0 {
		t.Errorf("expected streak=0 with goal=0, got %d", streak)
	}
}

func TestProfileCalcWeeklyStreak_OneFullWeek(t *testing.T) {
	// reference: Monday of the week
	ref := time.Date(2026, time.May, 25, 12, 0, 0, 0, time.UTC) // Monday
	// three sessions in that week (goal=3)
	dates := []time.Time{
		time.Date(2026, time.May, 25, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 27, 10, 0, 0, 0, time.UTC),
	}
	streak := profileCalcWeeklyStreak(ref, dates, 3)
	if streak != 1 {
		t.Errorf("expected streak=1, got %d", streak)
	}
}

func TestProfileCalcWeeklyStreak_TwoConsecutiveWeeks(t *testing.T) {
	ref := time.Date(2026, time.May, 25, 12, 0, 0, 0, time.UTC) // Monday
	dates := []time.Time{
		// current week
		time.Date(2026, time.May, 25, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC),
		// previous week
		time.Date(2026, time.May, 18, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 19, 10, 0, 0, 0, time.UTC),
	}
	streak := profileCalcWeeklyStreak(ref, dates, 2)
	if streak != 2 {
		t.Errorf("expected streak=2, got %d", streak)
	}
}

func TestProfileCalcWeeklyStreak_BrokenStreak(t *testing.T) {
	ref := time.Date(2026, time.May, 25, 12, 0, 0, 0, time.UTC)
	dates := []time.Time{
		// current week: meets goal (2 sessions, goal=2)
		time.Date(2026, time.May, 25, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC),
		// previous week: only 1 session, does not meet goal=2, breaks streak
		time.Date(2026, time.May, 18, 10, 0, 0, 0, time.UTC),
	}
	streak := profileCalcWeeklyStreak(ref, dates, 2)
	if streak != 1 {
		t.Errorf("expected streak=1 (current week met, previous not), got %d", streak)
	}
}

// ─── profileStartOfWeek ──────────────────────────────────────────────────────

func TestProfileStartOfWeek_Monday(t *testing.T) {
	monday := time.Date(2026, time.May, 25, 15, 30, 0, 0, time.UTC)
	got := profileStartOfWeek(monday)
	expected := time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestProfileStartOfWeek_Sunday(t *testing.T) {
	sunday := time.Date(2026, time.May, 31, 8, 0, 0, 0, time.UTC)
	got := profileStartOfWeek(sunday)
	expected := time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC) // Monday of that week
	if !got.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestProfileStartOfWeek_Wednesday(t *testing.T) {
	wednesday := time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC)
	got := profileStartOfWeek(wednesday)
	expected := time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

// ─── Integration tests ───────────────────────────────────────────────────────

func TestProfileRepositoryGetStatsEmptyUserIntegration(t *testing.T) {
	db := setupTestDB(t)
	cleanupUsers(t, db)

	userID := insertProfileUser(t, db, "profileempty", "profileempty@example.com")
	repo := NewProfileRepository(db)

	stats, err := repo.GetStats(context.Background(), userID, "all", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats.TotalWorkouts != 0 {
		t.Errorf("expected TotalWorkouts=0, got %d", stats.TotalWorkouts)
	}
	if len(stats.WeightHistory) != 0 {
		t.Errorf("expected empty WeightHistory, got %d records", len(stats.WeightHistory))
	}
	if len(stats.TopExercises) != 0 {
		t.Errorf("expected empty TopExercises, got %d records", len(stats.TopExercises))
	}
	if len(stats.MuscleRadar) != 0 {
		t.Errorf("expected empty MuscleRadar, got %d records", len(stats.MuscleRadar))
	}
}

func TestProfileRepositoryGetStatsWithWeightHistoryIntegration(t *testing.T) {
	db := setupTestDB(t)
	cleanupUsers(t, db)

	userID := insertProfileUser(t, db, "profileweight", "profileweight@example.com")
	fat1 := 18.5
	fat2 := 17.9
	insertBodyMetricRaw(t, db, userID, 79.2, &fat1, time.Now().Add(-10*24*time.Hour))
	insertBodyMetricRaw(t, db, userID, 78.5, &fat2, time.Now().Add(-3*24*time.Hour))

	repo := NewProfileRepository(db)
	stats, err := repo.GetStats(context.Background(), userID, "all", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stats.WeightHistory) != 2 {
		t.Fatalf("expected 2 weight records, got %d", len(stats.WeightHistory))
	}
	// ordered ASC: oldest first
	if stats.WeightHistory[0].WeightKg != 79.2 {
		t.Errorf("expected first weight 79.2, got %f", stats.WeightHistory[0].WeightKg)
	}
	if stats.WeightHistory[1].WeightKg != 78.5 {
		t.Errorf("expected second weight 78.5, got %f", stats.WeightHistory[1].WeightKg)
	}
	if stats.WeightHistory[1].BodyFatPercentage == nil || *stats.WeightHistory[1].BodyFatPercentage != fat2 {
		t.Errorf("expected body_fat_percentage %f in last record, got %v", fat2, stats.WeightHistory[1].BodyFatPercentage)
	}
}

func TestProfileRepositoryGetStatsWithGoalsIntegration(t *testing.T) {
	db := setupTestDB(t)
	cleanupUsers(t, db)

	userID := insertProfileUser(t, db, "profilegoals", "profilegoals@example.com")
	insertGoalRaw(t, db, userID, "Lose 5kg", "Run a marathon", 4)

	repo := NewProfileRepository(db)
	stats, err := repo.GetStats(context.Background(), userID, "all", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Goals == nil {
		t.Fatal("expected non-nil goals")
	}
	if stats.Goals.ShortTerm != "Lose 5kg" {
		t.Errorf("expected ShortTerm 'Lose 5kg', got %q", stats.Goals.ShortTerm)
	}
	if stats.Goals.LongTerm != "Run a marathon" {
		t.Errorf("expected LongTerm 'Run a marathon', got %q", stats.Goals.LongTerm)
	}
	if stats.Goals.TargetDaysPerWeek != 4 {
		t.Errorf("expected TargetDaysPerWeek=4, got %d", stats.Goals.TargetDaysPerWeek)
	}
	if stats.WeeklyGoal != 4 {
		t.Errorf("expected WeeklyGoal=4, got %d", stats.WeeklyGoal)
	}
}

func TestProfileRepositoryUpsertGoalsInsertIntegration(t *testing.T) {
	db := setupTestDB(t)
	cleanupUsers(t, db)

	userID := insertProfileUser(t, db, "profileupsert", "profileupsert@example.com")
	uid, err := uuid.Parse(userID)
	if err != nil {
		t.Fatalf("invalid UUID: %v", err)
	}

	repo := NewProfileRepository(db)
	goal := &model.UserGoal{
		UserID:            uid,
		ShortTerm:         "Lose weight",
		LongTerm:          "Fitness competition",
		TargetDaysPerWeek: 5,
	}

	if err := repo.UpsertGoals(context.Background(), goal); err != nil {
		t.Fatalf("unexpected error in UpsertGoals: %v", err)
	}

	var short, long string
	var targetDays int
	err = db.QueryRow(context.Background(),
		`SELECT short_term, long_term, target_days_per_week FROM public.user_goals WHERE user_id = $1::uuid`,
		userID,
	).Scan(&short, &long, &targetDays)
	if err != nil {
		t.Fatalf("error checking inserted goal: %v", err)
	}
	if short != "Lose weight" {
		t.Errorf("expected short_term 'Lose weight', got %q", short)
	}
	if targetDays != 5 {
		t.Errorf("expected target_days_per_week=5, got %d", targetDays)
	}
}

func TestProfileRepositoryUpsertGoalsUpdateIntegration(t *testing.T) {
	db := setupTestDB(t)
	cleanupUsers(t, db)

	userID := insertProfileUser(t, db, "profileupdate", "profileupdate@example.com")
	insertGoalRaw(t, db, userID, "Initial goal", "Initial long term", 3)

	uid, _ := uuid.Parse(userID)
	repo := NewProfileRepository(db)

	updated := &model.UserGoal{
		UserID:            uid,
		ShortTerm:         "Updated goal",
		LongTerm:          "Updated long term",
		TargetDaysPerWeek: 6,
	}
	if err := repo.UpsertGoals(context.Background(), updated); err != nil {
		t.Fatalf("unexpected error updating goals: %v", err)
	}

	var short string
	var days int
	err := db.QueryRow(context.Background(),
		`SELECT short_term, target_days_per_week FROM public.user_goals WHERE user_id = $1::uuid`,
		userID,
	).Scan(&short, &days)
	if err != nil {
		t.Fatalf("error checking updated goal: %v", err)
	}
	if short != "Updated goal" {
		t.Errorf("expected short_term 'Updated goal', got %q", short)
	}
	if days != 6 {
		t.Errorf("expected target_days_per_week=6, got %d", days)
	}
}

func TestProfileRepositoryAddBodyMetricIntegration(t *testing.T) {
	db := setupTestDB(t)
	cleanupUsers(t, db)

	userID := insertProfileUser(t, db, "profilemetric", "profilemetric@example.com")
	fat := 17.5
	height := 175.0
	req := &model.AddBodyMetricRequest{
		WeightKg:          80.0,
		BodyFatPercentage: &fat,
		HeightCm:          &height,
	}

	repo := NewProfileRepository(db)
	if err := repo.AddBodyMetric(context.Background(), userID, req); err != nil {
		t.Fatalf("unexpected error in AddBodyMetric: %v", err)
	}

	// Verify body_metrics row was inserted
	var weightKg float64
	var bodyFat float64
	err := db.QueryRow(context.Background(),
		`SELECT weight_kg, body_fat_percentage FROM public.body_metrics WHERE user_id = $1::uuid ORDER BY recorded_at DESC LIMIT 1`,
		userID,
	).Scan(&weightKg, &bodyFat)
	if err != nil {
		t.Fatalf("error verifying inserted body_metric: %v", err)
	}
	if weightKg != 80.0 {
		t.Errorf("expected weight_kg=80.0, got %f", weightKg)
	}
	if bodyFat != 17.5 {
		t.Errorf("expected body_fat_percentage=17.5, got %f", bodyFat)
	}

	// Verify user_profiles was upserted with weight and height
	var profileWeight, profileHeight float64
	err = db.QueryRow(context.Background(),
		`SELECT weight_kg, height_cm FROM public.user_profiles WHERE user_id = $1::uuid`,
		userID,
	).Scan(&profileWeight, &profileHeight)
	if err != nil {
		t.Fatalf("error verifying updated user_profile: %v", err)
	}
	if profileWeight != 80.0 {
		t.Errorf("expected profile weight_kg=80.0, got %f", profileWeight)
	}
	if profileHeight != 175.0 {
		t.Errorf("expected profile height_cm=175.0, got %f", profileHeight)
	}
}

func TestProfileRepositoryAddBodyMetricWithoutHeightIntegration(t *testing.T) {
	db := setupTestDB(t)
	cleanupUsers(t, db)

	userID := insertProfileUser(t, db, "profilenoheight", "profilenoheight@example.com")

	// Insert an existing profile with a known height
	_, err := db.Exec(context.Background(), `
		INSERT INTO public.user_profiles (user_id, weight_kg, height_cm, updated_at)
		VALUES ($1::uuid, 75.0, 180.0, CURRENT_TIMESTAMP)
	`, userID)
	if err != nil {
		t.Fatalf("error inserting existing profile: %v", err)
	}

	// Add a new metric without height; the existing height must not be overwritten
	req := &model.AddBodyMetricRequest{WeightKg: 77.0}
	repo := NewProfileRepository(db)
	if err := repo.AddBodyMetric(context.Background(), userID, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var profileHeight float64
	err = db.QueryRow(context.Background(),
		`SELECT height_cm FROM public.user_profiles WHERE user_id = $1::uuid`,
		userID,
	).Scan(&profileHeight)
	if err != nil {
		t.Fatalf("error reading profile: %v", err)
	}
	if profileHeight != 180.0 {
		t.Errorf("height should not have changed; expected 180.0, got %f", profileHeight)
	}
}
