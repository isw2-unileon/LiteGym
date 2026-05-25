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
	savedRoutine *model.AIRoutineToSave
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

func (r *routineAITestRoutineRepository) ListByUser(ctx context.Context, userID string) ([]model.OverviewRoutineSummary, error) {
	return []model.OverviewRoutineSummary{}, nil
}

func (r *routineAITestRoutineRepository) GetByID(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
	return nil, nil
}

func (r *routineAITestRoutineRepository) CountAIGenerationsInWindow(ctx context.Context, userID string, since time.Time) (int, error) {
	return 0, nil
}

func (r *routineAITestRoutineRepository) SaveGeneratedAIRoutine(ctx context.Context, routine model.AIRoutineToSave) (string, error) {
	r.savedRoutine = &routine
	return "saved-routine-1", nil
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
	if response.RoutineID != "" {
		t.Fatalf("expected preview response to omit routine id, got %q", response.RoutineID)
	}
	if routineRepo.savedRoutine != nil {
		t.Fatal("did not expect preview generation to save a routine")
	}

	promptPayload := decodeCapturedPromptPayload(t, capturedPrompt)
	userContext := mapField(t, promptPayload, "user_context")
	assertCompactUserContext(t, userContext)
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
