package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type routineAITestRoutineRepository struct{}

func (r *routineAITestRoutineRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{
		{ID: "routine-1", Name: "Push Pull Legs", ExerciseCount: 6},
		{ID: "routine-2", Name: "Upper Lower", ExerciseCount: 5},
	}, nil
}

func (r *routineAITestRoutineRepository) CountAIGenerationsInWindow(ctx context.Context, userID string, since time.Time) (int, error) {
	return 0, nil
}

func (r *routineAITestRoutineRepository) LogAIGeneration(ctx context.Context, userID string, createdAt time.Time) error {
	return nil
}

func (r *routineAITestRoutineRepository) ListAvailableExercisesForAI(
	ctx context.Context,
	userID string,
	targetMuscleGroups []string,
	limit int,
) ([]model.Exercise, error) {
	return []model.Exercise{
		{ID: "exercise-1", Name: "Bench Press", MuscleGroup: "chest", ExerciseType: "compound"},
		{ID: "exercise-2", Name: "Squat", MuscleGroup: "legs", ExerciseType: "compound"},
	}, nil
}

type routineAITestWorkoutSessionRepository struct{}

func (r *routineAITestWorkoutSessionRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error) {
	return []model.OverviewWorkoutSummary{
		{ID: "workout-1", Name: "Push Day", RoutineName: "Push Pull Legs", DurationMinutes: 70, ExerciseCount: 4},
		{ID: "workout-2", Name: "Leg Day", RoutineName: "Lower Body", DurationMinutes: 65, ExerciseCount: 5},
	}, nil
}

func (r *routineAITestWorkoutSessionRepository) ListRecentWorkoutHistoryByUser(ctx context.Context, userID string, limit int) ([]model.AIRoutineRecentWorkoutSession, error) {
	benchReps1 := 8
	benchReps2 := 7
	benchWeight1 := 75.0
	benchWeight2 := 77.5
	squatReps1 := 10
	squatWeight1 := 90.0

	return []model.AIRoutineRecentWorkoutSession{
		{
			SessionID:       "workout-1",
			SessionName:     "Push Day",
			RoutineName:     "Push Pull Legs",
			StartedAt:       time.Date(2026, time.May, 21, 0, 0, 0, 0, time.UTC),
			DurationMinutes: 70,
			Exercises: []model.AIRoutineRecentWorkoutExercise{
				{
					ExerciseID:    "exercise-1",
					ExerciseName:  "Bench Press",
					MuscleGroup:   "chest",
					ExerciseType:  "compound",
					ExerciseOrder: 1,
					Sets: []model.AIRoutineRecentWorkoutSet{
						{SetNumber: 1, Reps: &benchReps1, WeightKg: &benchWeight1},
						{SetNumber: 2, Reps: &benchReps2, WeightKg: &benchWeight2},
					},
				},
			},
		},
		{
			SessionID:       "workout-2",
			SessionName:     "Leg Day",
			RoutineName:     "Lower Body",
			StartedAt:       time.Date(2026, time.May, 19, 0, 0, 0, 0, time.UTC),
			DurationMinutes: 65,
			Exercises: []model.AIRoutineRecentWorkoutExercise{
				{
					ExerciseID:    "exercise-2",
					ExerciseName:  "Squat",
					MuscleGroup:   "legs",
					ExerciseType:  "compound",
					ExerciseOrder: 1,
					Sets: []model.AIRoutineRecentWorkoutSet{
						{SetNumber: 1, Reps: &squatReps1, WeightKg: &squatWeight1},
					},
				},
			},
		},
	}, nil
}

func (r *routineAITestWorkoutSessionRepository) ListTrainingDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
	if to.Sub(from) > 40*24*time.Hour {
		return []time.Time{
			time.Date(2026, time.May, 22, 0, 0, 0, 0, time.UTC),
			time.Date(2026, time.May, 21, 0, 0, 0, 0, time.UTC),
			time.Date(2026, time.May, 18, 0, 0, 0, 0, time.UTC),
		}, nil
	}

	return []time.Time{
		time.Date(2026, time.May, 22, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 21, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 18, 0, 0, 0, 0, time.UTC),
	}, nil
}

func (r *routineAITestWorkoutSessionRepository) ListMuscleDistributionByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error) {
	return []model.OverviewMuscleGroupShare{
		{Name: "chest", Count: 10, Percentage: 40},
		{Name: "legs", Count: 8, Percentage: 32},
		{Name: "back", Count: 6, Percentage: 24},
		{Name: "shoulders", Count: 1, Percentage: 4},
	}, 25, nil
}

type routineAITestBodyMetricRepository struct{}

func (r *routineAITestBodyMetricRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewBodyMetricEntry, error) {
	weightNow := 78.4
	weightBefore := 79.1
	bodyFatNow := 17.6
	bodyFatBefore := 18.0
	muscleMassNow := 36.8
	muscleMassBefore := 36.1

	return []model.OverviewBodyMetricEntry{
		{
			RecordedAt:        time.Date(2026, time.May, 22, 0, 0, 0, 0, time.UTC),
			WeightKg:          &weightNow,
			BodyFatPercentage: &bodyFatNow,
			MuscleMassKg:      &muscleMassNow,
		},
		{
			RecordedAt:        time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC),
			WeightKg:          &weightBefore,
			BodyFatPercentage: &bodyFatBefore,
			MuscleMassKg:      &muscleMassBefore,
		},
	}, nil
}

func TestGenerateRoutineJSONIncludesCompactUserContext(t *testing.T) {
	svc := NewRoutineAIService(
		&routineAITestRoutineRepository{},
		&routineAITestWorkoutSessionRepository{},
		&routineAITestBodyMetricRepository{},
		"test-key",
		"gemini-2.5-flash",
	)

	var capturedPrompt map[string]any
	svc.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(body, &capturedPrompt); err != nil {
				return nil, err
			}

			responseJSON := `{
				"candidates": [
					{
						"content": {
							"parts": [
								{
									"text": "{\"name\":\"Rutina de empuje\",\"exercises\":[]}"
								}
							]
						}
					}
				]
			}`

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(responseJSON)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	response, err := svc.GenerateRoutineJSON(context.Background(), "550e8400-e29b-41d4-a716-446655440000", model.AIRoutineGenerationRequest{
		Objective:            "Ganar fuerza",
		TargetMuscleGroups:   []string{"chest", "legs"},
		MandatoryExerciseIDs: []string{"exercise-1"},
		DurationMinutes:      60,
	})
	if err != nil {
		t.Fatalf("unexpected error generating routine: %v", err)
	}

	if response.RoutineJSON.Objective != "Ganar fuerza" {
		t.Fatalf("expected objective to be preserved, got %q", response.RoutineJSON.Objective)
	}
	if response.RoutineJSON.DurationMinutes != 60 {
		t.Fatalf("expected duration to be preserved, got %d", response.RoutineJSON.DurationMinutes)
	}
	if response.RoutineJSON.GenerationSource != "gemini" {
		t.Fatalf("expected generation source to default to gemini, got %q", response.RoutineJSON.GenerationSource)
	}
	if response.RoutineJSON.GeneratedAt.IsZero() {
		t.Fatal("expected generated_at to be populated")
	}
	if response.RateLimit.Limit != aiRoutineRateLimit || response.RateLimit.Remaining != 1 {
		t.Fatalf("unexpected rate limit payload: %#v", response.RateLimit)
	}

	contents, ok := capturedPrompt["contents"].([]any)
	if !ok || len(contents) != 1 {
		t.Fatalf("expected one content item in prompt, got %#v", capturedPrompt["contents"])
	}
	content, ok := contents[0].(map[string]any)
	if !ok {
		t.Fatalf("expected content map, got %#v", contents[0])
	}
	parts, ok := content["parts"].([]any)
	if !ok || len(parts) != 1 {
		t.Fatalf("expected one prompt part, got %#v", content["parts"])
	}
	part, ok := parts[0].(map[string]any)
	if !ok {
		t.Fatalf("expected prompt part map, got %#v", parts[0])
	}
	rawPrompt, ok := part["text"].(string)
	if !ok {
		t.Fatalf("expected prompt text string, got %#v", part["text"])
	}

	var promptPayload map[string]any
	if err := json.Unmarshal([]byte(rawPrompt), &promptPayload); err != nil {
		t.Fatalf("prompt text is not valid JSON: %v", err)
	}

	userContext, ok := promptPayload["user_context"].(map[string]any)
	if !ok {
		t.Fatalf("expected user_context in prompt payload, got %#v", promptPayload["user_context"])
	}
	if got := int(userContext["training_days_30d"].(float64)); got != 3 {
		t.Fatalf("expected 3 training days in last 30d, got %d", got)
	}
	if got := int(userContext["current_streak_days"].(float64)); got != 2 {
		t.Fatalf("expected streak of 2 days, got %d", got)
	}

	recentWorkouts, ok := userContext["recent_workouts"].([]any)
	if !ok || len(recentWorkouts) != 2 {
		t.Fatalf("expected two recent workouts, got %#v", userContext["recent_workouts"])
	}

	recentHistory, ok := userContext["recent_training_history"].([]any)
	if !ok || len(recentHistory) != 2 {
		t.Fatalf("expected two recent training history sessions, got %#v", userContext["recent_training_history"])
	}
	firstSession, ok := recentHistory[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first history session map, got %#v", recentHistory[0])
	}
	exercises, ok := firstSession["exercises"].([]any)
	if !ok || len(exercises) == 0 {
		t.Fatalf("expected exercises in recent history, got %#v", firstSession["exercises"])
	}
	firstExercise, ok := exercises[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first exercise map, got %#v", exercises[0])
	}
	sets, ok := firstExercise["sets"].([]any)
	if !ok || len(sets) != 2 {
		t.Fatalf("expected two sets for first exercise, got %#v", firstExercise["sets"])
	}
	firstSet, ok := sets[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first set map, got %#v", sets[0])
	}
	if _, ok := firstSet["reps"]; !ok {
		t.Fatal("expected reps in workout history set")
	}
	if _, ok := firstSet["weight_kg"]; !ok {
		t.Fatal("expected weight_kg in workout history set")
	}

	topMuscles, ok := userContext["top_muscle_groups"].([]any)
	if !ok || len(topMuscles) != 3 {
		t.Fatalf("expected three muscle groups, got %#v", userContext["top_muscle_groups"])
	}

	bodyMetrics, ok := userContext["body_metrics"].(map[string]any)
	if !ok {
		t.Fatalf("expected body_metrics in prompt payload, got %#v", userContext["body_metrics"])
	}
	if got := bodyMetrics["weight_kg_delta"]; got == nil {
		t.Fatal("expected weight_kg_delta to be present")
	}
}

func TestGenerateRoutineJSONFailsWithoutAPIKey(t *testing.T) {
	svc := NewRoutineAIService(
		&routineAITestRoutineRepository{},
		&routineAITestWorkoutSessionRepository{},
		&routineAITestBodyMetricRepository{},
		"",
		"",
	)

	svc.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("unexpected request")
		}),
	}

	_, err := svc.GenerateRoutineJSON(context.Background(), "550e8400-e29b-41d4-a716-446655440000", model.AIRoutineGenerationRequest{
		Objective:       "Ganar fuerza",
		DurationMinutes: 60,
	})
	if !errors.Is(err, ErrAIRoutineMissingAPIKey) {
		t.Fatalf("expected missing api key error, got %v", err)
	}
}
