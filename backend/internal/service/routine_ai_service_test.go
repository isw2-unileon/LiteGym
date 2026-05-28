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

type routineAITestRoutineRepository struct {
	savedRoutine *model.AIRoutineToSave
	getByIDFunc  func(ctx context.Context, userID, routineID string) (*model.Routine, error)
	countFunc    func(ctx context.Context, userID string, since time.Time) (int, error)
	loggedCount  int
}

func (r *routineAITestRoutineRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{
		{ID: "routine-1", Name: "Push Pull Legs", ExerciseCount: 6},
		{ID: "routine-2", Name: "Upper Lower", ExerciseCount: 5},
	}, nil
}

func (r *routineAITestRoutineRepository) ListByUser(ctx context.Context, userID string) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

func (r *routineAITestRoutineRepository) GetByID(ctx context.Context, userID, routineID string) (*model.Routine, error) {
	if r.getByIDFunc != nil {
		return r.getByIDFunc(ctx, userID, routineID)
	}
	return nil, nil
}

func (r *routineAITestRoutineRepository) CountAIGenerationsInWindow(ctx context.Context, userID string, since time.Time) (int, error) {
	if r.countFunc != nil {
		return r.countFunc(ctx, userID, since)
	}
	return 0, nil
}

func (r *routineAITestRoutineRepository) SaveGeneratedAIRoutine(ctx context.Context, routine model.AIRoutineToSave) (string, error) {
	r.savedRoutine = &routine
	return "saved-routine-1", nil
}

func (r *routineAITestRoutineRepository) LogAIGeneration(ctx context.Context, userID string, createdAt time.Time) error {
	r.loggedCount++
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
	routineRepo := &routineAITestRoutineRepository{}
	svc := NewRoutineAIService(
		routineRepo,
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
									"text": "{\"name\":\"Rutina de empuje\",\"exercises\":[{\"exercise_id\":\"exercise-1\",\"name\":\"Bench Press\",\"muscle_group\":\"chest\",\"exercise_type\":\"compound\",\"is_mandatory\":true,\"sets\":[{\"set_number\":1,\"target_reps_min\":6,\"target_reps_max\":8,\"target_weight_kg\":75,\"target_rir\":2,\"rest_seconds\":120}]}]}"
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

	assertGeneratedRoutineResponse(t, response, routineRepo.savedRoutine)

	promptPayload := decodeCapturedPromptPayload(t, capturedPrompt)
	userContext := mapField(t, promptPayload, "user_context")
	assertCompactUserContext(t, userContext)
}

func assertGeneratedRoutineResponse(
	t *testing.T,
	response model.AIRoutineGenerateResponse,
	savedRoutine *model.AIRoutineToSave,
) {
	t.Helper()

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
	if response.RoutineID != "saved-routine-1" {
		t.Fatalf("expected saved routine id, got %q", response.RoutineID)
	}
	if savedRoutine == nil {
		t.Fatal("expected generated routine to be saved")
	}
	if savedRoutine.Name != "Rutina de empuje" {
		t.Fatalf("expected saved routine name, got %q", savedRoutine.Name)
	}
	if len(savedRoutine.Exercises) != 1 || savedRoutine.Exercises[0].ExerciseID != "exercise-1" {
		t.Fatalf("unexpected saved routine exercises: %#v", savedRoutine.Exercises)
	}
	if len(savedRoutine.Exercises[0].Sets) != 1 {
		t.Fatalf("expected one planned set, got %#v", savedRoutine.Exercises[0].Sets)
	}
	savedSet := savedRoutine.Exercises[0].Sets[0]
	if savedSet.TargetWeightKg == nil || *savedSet.TargetWeightKg != 75 {
		t.Fatalf("expected saved target weight, got %#v", savedSet.TargetWeightKg)
	}
}

func decodeCapturedPromptPayload(t *testing.T, capturedPrompt map[string]any) map[string]any {
	t.Helper()

	contents := sliceField(t, capturedPrompt, "contents", 1)
	content := mapItem(t, contents[0], "content")
	parts := sliceField(t, content, "parts", 1)
	part := mapItem(t, parts[0], "prompt part")
	rawPrompt, ok := part["text"].(string)
	if !ok {
		t.Fatalf("expected prompt text string, got %#v", part["text"])
	}

	var promptPayload map[string]any
	if err := json.Unmarshal([]byte(rawPrompt), &promptPayload); err != nil {
		t.Fatalf("prompt text is not valid JSON: %v", err)
	}

	return promptPayload
}

func assertCompactUserContext(t *testing.T, userContext map[string]any) {
	t.Helper()

	if got := int(userContext["training_days_30d"].(float64)); got != 3 {
		t.Fatalf("expected 3 training days in last 30d, got %d", got)
	}
	if got := int(userContext["current_streak_days"].(float64)); got != 0 {
		t.Fatalf("expected streak of 0 days, got %d", got)
	}

	sliceField(t, userContext, "recent_workouts", 2)
	assertRecentTrainingHistory(t, userContext)

	topMuscles := sliceField(t, userContext, "top_muscle_groups", 3)
	if len(topMuscles) != 3 {
		t.Fatalf("expected three muscle groups, got %#v", userContext["top_muscle_groups"])
	}

	bodyMetrics := mapField(t, userContext, "body_metrics")
	if got := bodyMetrics["weight_kg_delta"]; got == nil {
		t.Fatal("expected weight_kg_delta to be present")
	}
}

func assertRecentTrainingHistory(t *testing.T, userContext map[string]any) {
	t.Helper()

	recentHistory := sliceField(t, userContext, "recent_training_history", 2)
	firstSession := mapItem(t, recentHistory[0], "first history session")
	exercises := sliceField(t, firstSession, "exercises", 1)
	firstExercise := mapItem(t, exercises[0], "first exercise")
	sets := sliceField(t, firstExercise, "sets", 2)
	firstSet := mapItem(t, sets[0], "first set")
	if _, ok := firstSet["reps"]; !ok {
		t.Fatal("expected reps in workout history set")
	}
	if _, ok := firstSet["weight_kg"]; !ok {
		t.Fatal("expected weight_kg in workout history set")
	}
}

func mapField(t *testing.T, values map[string]any, key string) map[string]any {
	t.Helper()

	item, ok := values[key].(map[string]any)
	if !ok {
		t.Fatalf("expected %s map, got %#v", key, values[key])
	}

	return item
}

func mapItem(t *testing.T, value any, label string) map[string]any {
	t.Helper()

	item, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected %s map, got %#v", label, value)
	}

	return item
}

func sliceField(t *testing.T, values map[string]any, key string, minLen int) []any {
	t.Helper()

	items, ok := values[key].([]any)
	if !ok || len(items) < minLen {
		t.Fatalf("expected at least %d item(s) for %s, got %#v", minLen, key, values[key])
	}

	return items
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

func TestUpgradeRoutineJSONIncludesExistingRoutineAndFeedback(t *testing.T) {
	routineRepo := &routineAITestRoutineRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.Routine, error) {
			return testRoutineForUpgrade(routineID, userID), nil
		},
	}

	svc := NewRoutineAIService(
		routineRepo,
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
									"text": "{\"name\":\"Push Day 2.0\",\"objective\":\"More balanced push session\",\"duration_minutes\":45,\"target_muscles\":[\"chest\",\"shoulders\"],\"mandatory_count\":1,\"exercises\":[{\"exercise_id\":\"exercise-1\",\"name\":\"Bench Press\",\"muscle_group\":\"chest\",\"exercise_type\":\"compound\",\"is_mandatory\":true,\"sets\":[{\"set_number\":1,\"target_reps_min\":5,\"target_reps_max\":7,\"target_weight_kg\":77.5,\"target_rir\":2,\"rest_seconds\":150}]},{\"exercise_id\":\"exercise-2\",\"name\":\"Squat\",\"muscle_group\":\"legs\",\"exercise_type\":\"compound\",\"is_mandatory\":false,\"sets\":[{\"set_number\":1,\"target_reps_min\":8,\"target_reps_max\":10,\"rest_seconds\":90}]}]}"
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

	response, err := svc.UpgradeRoutineJSON(context.Background(), "550e8400-e29b-41d4-a716-446655440000", "routine-1", model.AIRoutineUpgradeRequest{
		Message:         "Make it more balanced and slightly harder.",
		FeedbackMessage: "Keep Bench Press but raise intensity.",
	})
	if err != nil {
		t.Fatalf("unexpected error upgrading routine: %v", err)
	}

	assertRoutineUpgradeResponse(t, response)
	if routineRepo.savedRoutine != nil {
		t.Fatal("did not expect upgrade flow to persist the routine")
	}
	if routineRepo.loggedCount != 1 {
		t.Fatalf("expected one generation log, got %d", routineRepo.loggedCount)
	}

	promptPayload := decodeCapturedPromptPayload(t, capturedPrompt)
	if got := promptPayload["message"]; got != "Make it more balanced and slightly harder." {
		t.Fatalf("expected message in prompt, got %#v", got)
	}
	if got := promptPayload["feedback_message"]; got != "Keep Bench Press but raise intensity." {
		t.Fatalf("expected feedback_message in prompt, got %#v", got)
	}
	existingRoutine := mapField(t, promptPayload, "existing_routine")
	if got := existingRoutine["name"]; got != "Push Day" {
		t.Fatalf("expected existing routine name in prompt, got %#v", got)
	}
	existingExercises := sliceField(t, existingRoutine, "exercises", 1)
	firstExercise := mapItem(t, existingExercises[0], "existing routine first exercise")
	if got := firstExercise["exercise_id"]; got != "exercise-1" {
		t.Fatalf("expected existing routine exercise id, got %#v", got)
	}
}

func testRoutineForUpgrade(routineID, userID string) *model.Routine {
	repsMin := 6
	repsMax := 8
	weight := 70.0
	rest := 120

	return &model.Routine{
		ID:          routineID,
		UserID:      userID,
		Name:        "Push Day",
		Description: "Improve upper body strength",
		Source:      "manual",
		Exercises: []model.RoutineExercise{
			{
				ID:            "re-1",
				RoutineID:     routineID,
				ExerciseID:    "exercise-1",
				ExerciseName:  "Bench Press",
				MuscleGroup:   "chest",
				ExerciseType:  "compound",
				ExerciseOrder: 1,
				Sets: []model.RoutineSet{
					{
						ID:                "set-1",
						RoutineExerciseID: "re-1",
						SetNumber:         1,
						TargetRepsMin:     &repsMin,
						TargetRepsMax:     &repsMax,
						TargetWeightKg:    &weight,
						RestSeconds:       &rest,
					},
				},
			},
		},
	}
}

func assertRoutineUpgradeResponse(t *testing.T, response model.AIRoutineUpgradeResponse) {
	t.Helper()

	if response.RoutineJSON.Name != "Push Day 2.0" {
		t.Fatalf("expected upgraded routine name, got %q", response.RoutineJSON.Name)
	}
	if response.RateLimit.Remaining != 1 {
		t.Fatalf("expected one remaining generation, got %#v", response.RateLimit)
	}
	if response.Diff.Summary.AddedExercises != 1 || response.Diff.Summary.ModifiedExercises != 1 {
		t.Fatalf("unexpected diff summary: %#v", response.Diff.Summary)
	}
	if len(response.Diff.Exercises) != 2 {
		t.Fatalf("expected two diff exercise entries, got %#v", response.Diff.Exercises)
	}
	if response.Diff.Exercises[0].ChangeType != "modified" {
		t.Fatalf("expected first exercise diff to be modified, got %#v", response.Diff.Exercises[0])
	}
	if len(response.Diff.Exercises[0].Sets) != 1 || response.Diff.Exercises[0].Sets[0].ChangeType != "modified" {
		t.Fatalf("expected first exercise set diff to be modified, got %#v", response.Diff.Exercises[0].Sets)
	}
	if response.Diff.Exercises[1].ChangeType != "added" {
		t.Fatalf("expected second exercise diff to be added, got %#v", response.Diff.Exercises[1])
	}
}

func TestUpgradeRoutineJSONRateLimited(t *testing.T) {
	svc := NewRoutineAIService(
		&routineAITestRoutineRepository{
			countFunc: func(ctx context.Context, userID string, since time.Time) (int, error) {
				return aiRoutineRateLimit, nil
			},
		},
		&routineAITestWorkoutSessionRepository{},
		&routineAITestBodyMetricRepository{},
		"test-key",
		"gemini-2.5-flash",
	)

	response, err := svc.UpgradeRoutineJSON(context.Background(), "550e8400-e29b-41d4-a716-446655440000", "routine-1", model.AIRoutineUpgradeRequest{
		Message: "Improve it",
	})
	if !errors.Is(err, ErrAIRoutineRateLimited) {
		t.Fatalf("expected ErrAIRoutineRateLimited, got %v", err)
	}
	if response.RateLimit.Limit != aiRoutineRateLimit || response.RateLimit.Remaining != 0 {
		t.Fatalf("unexpected rate limit payload: %#v", response.RateLimit)
	}
}

func TestBuildAIRoutineUpgradeDiff(t *testing.T) {
	before := model.AIRoutineJSON{
		Exercises: []model.AIRoutineExercise{
			{
				ExerciseID:   "exercise-1",
				Name:         "Bench Press",
				MuscleGroup:  "chest",
				ExerciseType: "compound",
				Sets: []model.AIRoutineExerciseSet{
					{SetNumber: 1, TargetRepsText: "6-8"},
				},
			},
			{
				ExerciseID:   "exercise-2",
				Name:         "Row",
				MuscleGroup:  "back",
				ExerciseType: "compound",
			},
		},
	}
	after := model.AIRoutineJSON{
		Exercises: []model.AIRoutineExercise{
			{
				ExerciseID:   "exercise-1",
				Name:         "Bench Press",
				MuscleGroup:  "chest",
				ExerciseType: "compound",
				Sets: []model.AIRoutineExerciseSet{
					{SetNumber: 1, TargetRepsText: "5-7"},
				},
			},
			{
				ExerciseID:   "exercise-3",
				Name:         "Shoulder Press",
				MuscleGroup:  "shoulders",
				ExerciseType: "compound",
			},
		},
	}

	diff := buildAIRoutineUpgradeDiff(before, after)

	if diff.Summary.AddedExercises != 1 {
		t.Fatalf("expected one added exercise, got %#v", diff.Summary)
	}
	if diff.Summary.RemovedExercises != 1 {
		t.Fatalf("expected one removed exercise, got %#v", diff.Summary)
	}
	if diff.Summary.ModifiedExercises != 1 {
		t.Fatalf("expected one modified exercise, got %#v", diff.Summary)
	}
	if len(diff.Exercises) != 3 {
		t.Fatalf("expected three exercise diff entries, got %#v", diff.Exercises)
	}
	if diff.Exercises[0].ChangeType != "modified" || diff.Exercises[0].Sets[0].ChangeType != "modified" {
		t.Fatalf("expected first entry to be modified with modified set, got %#v", diff.Exercises[0])
	}
	if diff.Exercises[1].ChangeType != "removed" {
		t.Fatalf("expected second entry removed, got %#v", diff.Exercises[1])
	}
	if diff.Exercises[2].ChangeType != "added" {
		t.Fatalf("expected third entry added, got %#v", diff.Exercises[2])
	}
}
