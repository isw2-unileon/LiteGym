package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type routineAITestRoutineRepository struct {
	savedRoutine      *model.AIRoutineToSave
	overwritten       *model.AIRoutineToSave
	overwrittenID     string
	getByIDFunc       func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error)
	countFunc         func(ctx context.Context, userID string, since time.Time) (int, error)
	loggedCount       int
}

type routineAITestExerciseRepository struct {
	exercises        []model.Exercise
	createdExercises []*model.Exercise
}

func (r *routineAITestRoutineRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{
		{ID: "routine-1", Name: "Push Pull Legs", ExerciseCount: 6},
		{ID: "routine-2", Name: "Upper Lower", ExerciseCount: 5},
	}, nil
}

func (r *routineAITestRoutineRepository) ListByUser(ctx context.Context, userID, search string) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

func (r *routineAITestRoutineRepository) GetByID(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
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

func (r *routineAITestRoutineRepository) OverwriteGeneratedAIRoutine(ctx context.Context, routineID, userID string, routine model.AIRoutineToSave) error {
	r.overwrittenID = routineID
	r.overwritten = &routine
	return nil
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

func (r *routineAITestExerciseRepository) GetByID(ctx context.Context, id string) (*model.Exercise, error) {
	for index := range r.exercises {
		if r.exercises[index].ID == id {
			exercise := r.exercises[index]
			return &exercise, nil
		}
	}

	return nil, pgx.ErrNoRows
}

func (r *routineAITestExerciseRepository) List(ctx context.Context, filters model.ExerciseFilter) ([]model.Exercise, int, error) {
	matches := make([]model.Exercise, 0, len(r.exercises))
	for _, exercise := range r.exercises {
		if filters.Search != "" && !containsInsensitive(exercise.Name, filters.Search) {
			continue
		}
		if filters.Type != "" && exercise.ExerciseType != filters.Type {
			continue
		}
		if filters.MuscleGroup != "" && exercise.MuscleGroup != filters.MuscleGroup {
			continue
		}
		matches = append(matches, exercise)
	}

	return matches, len(matches), nil
}

func (r *routineAITestExerciseRepository) ListWorkoutSessionsByExercise(ctx context.Context, exerciseID, userID string, limit int) ([]model.ExerciseWorkoutSessionSummary, error) {
	return nil, nil
}

func (r *routineAITestExerciseRepository) GetInsights(ctx context.Context, exerciseID, userID string) (model.ExerciseInsights, error) {
	return model.ExerciseInsights{}, nil
}

func (r *routineAITestExerciseRepository) Create(ctx context.Context, exercise *model.Exercise) error {
	exercise.ID = "created-exercise-" + time.Now().UTC().Format("150405.000000")
	exercise.CreatedAt = time.Now().UTC()
	exerciseCopy := *exercise
	r.exercises = append(r.exercises, exerciseCopy)
	r.createdExercises = append(r.createdExercises, &exerciseCopy)
	return nil
}

func (r *routineAITestExerciseRepository) UpdateExercise(ctx context.Context, exercise *model.Exercise) error {
	return nil
}

func (r *routineAITestExerciseRepository) DeleteExercise(ctx context.Context, id string) error {
	return nil
}

func containsInsensitive(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
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
	exerciseRepo := &routineAITestExerciseRepository{
		exercises: []model.Exercise{
			{ID: "exercise-1", Name: "Bench Press", MuscleGroup: "chest", ExerciseType: "compound", IsOfficial: true},
			{ID: "exercise-2", Name: "Squat", MuscleGroup: "legs", ExerciseType: "compound", IsOfficial: true},
		},
	}
	svc := NewRoutineAIService(
		routineRepo,
		NewExerciseService(exerciseRepo),
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
		Objective:          "Ganar fuerza",
		TargetMuscleGroups: []string{"chest", "legs"},
		MandatoryExercises: []string{"Bench Press"},
		Notes:              "Prioriza press con barra y descanso amplio.",
		DurationMinutes:    60,
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
	if response.RoutineID != "" {
		t.Fatalf("expected preview response to omit routine id, got %q", response.RoutineID)
	}
	if routineRepo.savedRoutine != nil {
		t.Fatal("did not expect preview generation to save a routine")
	}

	promptPayload := decodeCapturedPromptPayload(t, capturedPrompt)
	userContext := mapField(t, promptPayload, "user_context")
	assertCompactUserContext(t, userContext)

	mandatoryExercises := sliceField(t, promptPayload, "mandatory_exercises", 1)
	if got := mandatoryExercises[0].(string); got != "Bench Press" {
		t.Fatalf("expected mandatory exercise name to be sent, got %q", got)
	}
	if got := promptPayload["user_notes"].(string); got != "Prioriza press con barra y descanso amplio." {
		t.Fatalf("expected user notes to be forwarded, got %q", got)
	}

	systemInstruction := strings.TrimSpace(capturedPromptSystemInstruction(t, capturedPrompt))
	if !strings.Contains(systemInstruction, "Build the most complete and sensible routine possible") {
		t.Fatalf("expected system instruction to encourage adaptive exercise count, got %q", systemInstruction)
	}
	if !strings.Contains(systemInstruction, "Do not force a one-to-one mapping") {
		t.Fatalf("expected system instruction to avoid one-to-one muscle mapping, got %q", systemInstruction)
	}
}

func TestGenerateRoutineJSONResolvesHallucinatedExerciseIDByName(t *testing.T) {
	routineRepo := &routineAITestRoutineRepository{}
	exerciseRepo := &routineAITestExerciseRepository{}
	svc := NewRoutineAIService(
		routineRepo,
		NewExerciseService(exerciseRepo),
		&routineAITestWorkoutSessionRepository{},
		&routineAITestBodyMetricRepository{},
		"test-key",
		"gemini-2.5-flash",
	)

	svc.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			responseJSON := `{
				"candidates": [
					{
						"content": {
							"parts": [
								{
									"text": "{\"name\":\"Rutina de fuerza\",\"exercises\":[{\"exercise_id\":\"66666666-6666-6666-6666-666666666661\",\"name\":\"Bench Press\",\"muscle_group\":\"chest\",\"exercise_type\":\"compound\",\"is_mandatory\":true,\"sets\":[{\"set_number\":1,\"target_reps_min\":5,\"target_reps_max\":5,\"target_rir\":2,\"rest_seconds\":120}]}]}"
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
		Objective:       "Ganar fuerza",
		DurationMinutes: 60,
	})
	if err != nil {
		t.Fatalf("unexpected error generating routine: %v", err)
	}

	if response.RoutineID != "" {
		t.Fatalf("expected preview response to omit routine id, got %q", response.RoutineID)
	}

	saved, err := svc.SaveGeneratedRoutineJSON(context.Background(), "550e8400-e29b-41d4-a716-446655440000", response.RoutineJSON)
	if err != nil {
		t.Fatalf("unexpected error saving generated routine: %v", err)
	}

	if saved.RoutineID != "saved-routine-1" {
		t.Fatalf("expected saved routine id, got %q", saved.RoutineID)
	}
	if routineRepo.savedRoutine == nil {
		t.Fatal("expected generated routine to be saved")
	}
	if len(routineRepo.savedRoutine.Exercises) != 1 {
		t.Fatalf("expected one saved exercise, got %#v", routineRepo.savedRoutine.Exercises)
	}
	if len(exerciseRepo.createdExercises) != 1 {
		t.Fatalf("expected one exercise to be created, got %#v", exerciseRepo.createdExercises)
	}
	if exerciseRepo.createdExercises[0].OwnerUserID == nil || *exerciseRepo.createdExercises[0].OwnerUserID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("expected created exercise to belong to the requesting user, got %#v", exerciseRepo.createdExercises[0].OwnerUserID)
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

func capturedPromptSystemInstruction(t *testing.T, capturedPrompt map[string]any) string {
	t.Helper()

	systemInstruction, ok := capturedPrompt["system_instruction"].(map[string]any)
	if !ok {
		t.Fatalf("expected system_instruction object, got %#v", capturedPrompt["system_instruction"])
	}

	parts := sliceField(t, systemInstruction, "parts", 1)
	part := mapItem(t, parts[0], "system instruction part")
	rawText, ok := part["text"].(string)
	if !ok {
		t.Fatalf("expected system instruction text string, got %#v", part["text"])
	}

	return rawText
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

func assertUpgradeRecentTrainingHistory(t *testing.T, userContext map[string]any) {
	t.Helper()

	recentHistory := sliceField(t, userContext, "recent_training_history", 2)
	firstSession := mapItem(t, recentHistory[0], "first upgrade history session")
	exercises := sliceField(t, firstSession, "exercises", 1)
	if len(exercises) != 1 {
		t.Fatalf("expected filtered related exercises only, got %#v", firstSession["exercises"])
	}
	firstExercise := mapItem(t, exercises[0], "first related exercise")
	if got := firstExercise["exercise_id"]; got != "exercise-1" {
		t.Fatalf("expected only related exercise history, got %#v", firstExercise["exercise_id"])
	}

	secondSession := mapItem(t, recentHistory[1], "second upgrade history session")
	secondExercises := sliceField(t, secondSession, "exercises", 1)
	if len(secondExercises) != 1 {
		t.Fatalf("expected filtered related exercises in second session, got %#v", secondSession["exercises"])
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
	exerciseRepo := &routineAITestExerciseRepository{}
	svc := NewRoutineAIService(
		&routineAITestRoutineRepository{},
		NewExerciseService(exerciseRepo),
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

func TestGenerateRoutineJSONTreatsMalformedGeminiResponseAsProviderUnavailable(t *testing.T) {
	routineRepo := &routineAITestRoutineRepository{}
	exerciseRepo := &routineAITestExerciseRepository{}
	svc := NewRoutineAIService(
		routineRepo,
		NewExerciseService(exerciseRepo),
		&routineAITestWorkoutSessionRepository{},
		&routineAITestBodyMetricRepository{},
		"test-key",
		"gemini-2.5-flash",
	)

	svc.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("{not-json")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	_, err := svc.GenerateRoutineJSON(context.Background(), "550e8400-e29b-41d4-a716-446655440000", model.AIRoutineGenerationRequest{
		Objective:       "Ganar fuerza",
		DurationMinutes: 60,
	})
	if !errors.Is(err, ErrAIRoutineProviderUnavailable) {
		t.Fatalf("expected provider unavailable error, got %v", err)
	}
}

func TestUpgradeRoutineJSONIncludesExistingRoutineAndFeedback(t *testing.T) {
	routineRepo := &routineAITestRoutineRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
			return testRoutineForUpgrade(routineID), nil
		},
	}
	exerciseRepo := &routineAITestExerciseRepository{
		exercises: []model.Exercise{
			{ID: "exercise-1", Name: "Bench Press", MuscleGroup: "chest", ExerciseType: "compound", IsOfficial: true},
			{ID: "exercise-2", Name: "Squat", MuscleGroup: "legs", ExerciseType: "compound", IsOfficial: true},
		},
	}

	svc := NewRoutineAIService(
		routineRepo,
		NewExerciseService(exerciseRepo),
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
	exerciseCatalog := sliceField(t, promptPayload, "exercise_catalog", 2)
	if len(exerciseCatalog) != 2 {
		t.Fatalf("expected full exercise catalog for upgrade, got %#v", promptPayload["exercise_catalog"])
	}
	userContext := mapField(t, promptPayload, "user_context")
	assertUpgradeRecentTrainingHistory(t, userContext)

	systemInstruction := decodeCapturedSystemInstruction(t, capturedPrompt)
	if !strings.Contains(systemInstruction, "Treat existing_routine as the base plan to refine") {
		t.Fatalf("expected refined upgrade instruction, got %q", systemInstruction)
	}
	if !strings.Contains(systemInstruction, "Respect message and feedback_message as strong user instructions") {
		t.Fatalf("expected feedback guidance in instruction, got %q", systemInstruction)
	}
	if !strings.Contains(systemInstruction, "Never include keys such as exercise_catalog, existing_routine, user_context, message, feedback_message, or output_contract in the response") {
		t.Fatalf("expected anti-echo guidance in instruction, got %q", systemInstruction)
	}
}

func testRoutineForUpgrade(routineID string) *model.RoutineDetail {
	repsMin := 6
	repsMax := 8
	weight := 70.0
	rest := 120

	return &model.RoutineDetail{
		ID:          routineID,
		Name:        "Push Day",
		Description: "Improve upper body strength",
		Source:      "manual",
		Exercises: []model.RoutineExerciseDetail{
			{
				ID:            "re-1",
				ExerciseID:    "exercise-1",
				Name:          "Bench Press",
				MuscleGroup:   "chest",
				ExerciseType:  "compound",
				ExerciseOrder: 1,
				Sets: []model.RoutineExerciseSetDetail{
					{
						ID:             "set-1",
						SetNumber:      1,
						TargetRepsMin:  &repsMin,
						TargetRepsMax:  &repsMax,
						TargetWeightKg: &weight,
						RestSeconds:    &rest,
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
	if response.RateLimit.Remaining != aiRoutineRateLimit-1 {
		t.Fatalf("expected %d remaining generations, got %#v", aiRoutineRateLimit-1, response.RateLimit)
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
	exerciseRepo := &routineAITestExerciseRepository{}
	svc := NewRoutineAIService(
		&routineAITestRoutineRepository{
			countFunc: func(ctx context.Context, userID string, since time.Time) (int, error) {
				return aiRoutineRateLimit, nil
			},
		},
		NewExerciseService(exerciseRepo),
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

func TestUpgradeRoutineJSONRejectsEmptyRoutine(t *testing.T) {
	routineRepo := &routineAITestRoutineRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
			return testRoutineForUpgrade(routineID), nil
		},
	}
	exerciseRepo := &routineAITestExerciseRepository{
		exercises: []model.Exercise{
			{ID: "exercise-1", Name: "Bench Press", MuscleGroup: "chest", ExerciseType: "compound", IsOfficial: true},
			{ID: "exercise-2", Name: "Squat", MuscleGroup: "legs", ExerciseType: "compound", IsOfficial: true},
		},
	}

	svc := NewRoutineAIService(
		routineRepo,
		NewExerciseService(exerciseRepo),
		&routineAITestWorkoutSessionRepository{},
		&routineAITestBodyMetricRepository{},
		"test-key",
		"gemini-2.5-flash",
	)

	svc.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			responseJSON := `{
				"candidates": [
					{
						"content": {
							"parts": [
								{
									"text": "{\"name\":\"Push Day Revised\",\"objective\":\"Small refinement\",\"duration_minutes\":45,\"target_muscles\":[\"chest\"],\"mandatory_count\":0,\"exercises\":[]}"
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
		Message: "Improve it a little.",
	})
	if !errors.Is(err, ErrAIRoutineProviderUnavailable) {
		t.Fatalf("expected ErrAIRoutineProviderUnavailable, got %v", err)
	}
	if response.RoutineJSON.Name != "" || len(response.Diff.Exercises) != 0 {
		t.Fatalf("expected empty response payload on provider error, got %#v", response)
	}
}

func TestUpgradeRoutineJSONRejectsPromptEchoResponse(t *testing.T) {
	routineRepo := &routineAITestRoutineRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
			return testRoutineForUpgrade(routineID), nil
		},
	}
	exerciseRepo := &routineAITestExerciseRepository{
		exercises: []model.Exercise{
			{ID: "exercise-1", Name: "Bench Press", MuscleGroup: "chest", ExerciseType: "compound", IsOfficial: true},
			{ID: "exercise-2", Name: "Squat", MuscleGroup: "legs", ExerciseType: "compound", IsOfficial: true},
		},
	}

	svc := NewRoutineAIService(
		routineRepo,
		NewExerciseService(exerciseRepo),
		&routineAITestWorkoutSessionRepository{},
		&routineAITestBodyMetricRepository{},
		"test-key",
		"gemini-2.5-flash",
	)

	svc.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			responseJSON := `{
				"candidates": [
					{
						"content": {
							"parts": [
								{
									"text": "{\"message\":\"Add one exercise\",\"existing_routine\":{\"name\":\"Push Day\"},\"exercise_catalog\":[{\"id\":\"exercise-1\"}],\"output_contract\":{\"name\":\"string\"}}"
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

	_, err := svc.UpgradeRoutineJSON(context.Background(), "550e8400-e29b-41d4-a716-446655440000", "routine-1", model.AIRoutineUpgradeRequest{
		Message: "Add one exercise.",
	})
	if !errors.Is(err, ErrAIRoutineProviderUnavailable) {
		t.Fatalf("expected ErrAIRoutineProviderUnavailable for echoed prompt, got %v", err)
	}
}

func TestSaveUpgradedRoutineAsNew(t *testing.T) {
	routineRepo := &routineAITestRoutineRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
			return testRoutineForUpgrade(routineID), nil
		},
	}
	exerciseRepo := &routineAITestExerciseRepository{
		exercises: []model.Exercise{
			{ID: "exercise-1", Name: "Bench Press", MuscleGroup: "chest", ExerciseType: "compound", IsOfficial: true},
			{ID: "exercise-2", Name: "Squat", MuscleGroup: "legs", ExerciseType: "compound", IsOfficial: true},
		},
	}
	svc := NewRoutineAIService(
		routineRepo,
		NewExerciseService(exerciseRepo),
		&routineAITestWorkoutSessionRepository{},
		&routineAITestBodyMetricRepository{},
		"test-key",
		"gemini-2.5-flash",
	)

	generated := model.AIRoutineJSON{
		Name:            "Push Day Plus",
		Objective:       "Refined push day",
		DurationMinutes: 45,
		Exercises: []model.AIRoutineExercise{
			{
				ExerciseID:   "exercise-1",
				Name:         "Bench Press",
				MuscleGroup:  "chest",
				ExerciseType: "compound",
				IsMandatory:  true,
				Sets: []model.AIRoutineExerciseSet{
					{SetNumber: 1, TargetRepsMin: intPtr(6), TargetRepsMax: intPtr(8)},
				},
			},
		},
	}

	response, err := svc.SaveUpgradedRoutineAsNew(context.Background(), "550e8400-e29b-41d4-a716-446655440000", "routine-1", generated)
	if err != nil {
		t.Fatalf("unexpected error saving upgraded routine as new: %v", err)
	}
	if response.RoutineID != "saved-routine-1" {
		t.Fatalf("expected saved routine id, got %#v", response)
	}
	if routineRepo.savedRoutine == nil || routineRepo.savedRoutine.Name != "Push Day Plus" {
		t.Fatalf("expected saved routine payload, got %#v", routineRepo.savedRoutine)
	}
}

func TestOverwriteRoutineWithGeneratedJSON(t *testing.T) {
	routineRepo := &routineAITestRoutineRepository{
		getByIDFunc: func(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
			return testRoutineForUpgrade(routineID), nil
		},
	}
	exerciseRepo := &routineAITestExerciseRepository{
		exercises: []model.Exercise{
			{ID: "exercise-1", Name: "Bench Press", MuscleGroup: "chest", ExerciseType: "compound", IsOfficial: true},
		},
	}
	svc := NewRoutineAIService(
		routineRepo,
		NewExerciseService(exerciseRepo),
		&routineAITestWorkoutSessionRepository{},
		&routineAITestBodyMetricRepository{},
		"test-key",
		"gemini-2.5-flash",
	)

	generated := model.AIRoutineJSON{
		Name:            "Push Day Reworked",
		Objective:       "Refined push day",
		DurationMinutes: 45,
		Exercises: []model.AIRoutineExercise{
			{
				ExerciseID:   "exercise-1",
				Name:         "Bench Press",
				MuscleGroup:  "chest",
				ExerciseType: "compound",
				IsMandatory:  true,
				Sets: []model.AIRoutineExerciseSet{
					{SetNumber: 1, TargetRepsMin: intPtr(5), TargetRepsMax: intPtr(7)},
				},
			},
		},
	}

	response, err := svc.OverwriteRoutineWithGeneratedJSON(context.Background(), "550e8400-e29b-41d4-a716-446655440000", "routine-1", generated)
	if err != nil {
		t.Fatalf("unexpected error overwriting routine: %v", err)
	}
	if response.RoutineID != "routine-1" {
		t.Fatalf("expected overwritten routine id, got %#v", response)
	}
	if routineRepo.overwritten == nil || routineRepo.overwrittenID != "routine-1" {
		t.Fatalf("expected overwrite payload, got %#v / %q", routineRepo.overwritten, routineRepo.overwrittenID)
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

func decodeCapturedSystemInstruction(t *testing.T, capturedPrompt map[string]any) string {
	t.Helper()

	systemInstruction, ok := capturedPrompt["system_instruction"].(map[string]any)
	if !ok {
		t.Fatalf("expected system_instruction object, got %#v", capturedPrompt["system_instruction"])
	}
	parts, ok := systemInstruction["parts"].([]any)
	if !ok || len(parts) == 0 {
		t.Fatalf("expected system_instruction parts, got %#v", systemInstruction["parts"])
	}
	firstPart, ok := parts[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first system instruction part, got %#v", parts[0])
	}
	text, ok := firstPart["text"].(string)
	if !ok {
		t.Fatalf("expected system instruction text, got %#v", firstPart["text"])
	}

	return text
}

func intPtr(value int) *int {
	return &value
}
