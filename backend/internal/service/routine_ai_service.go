package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
)

var (
	ErrAIRoutineInvalidInput        = errors.New("invalid ai routine input")
	ErrAIRoutineRateLimited         = errors.New("ai routine generation rate limit exceeded")
	ErrAIRoutineProviderUnavailable = errors.New("ai provider unavailable")
	ErrAIRoutineMissingAPIKey       = errors.New("ai provider missing api key")
)

const (
	aiRoutineRateLimit  = 2
	aiRoutineRateWindow = time.Hour
)

// RoutineAIService generates AI routine JSON and enforces per-user rate limits.
type RoutineAIService struct {
	repo               repository.RoutineRepository
	workoutSessionRepo repository.WorkoutSessionRepository
	bodyMetricRepo     repository.BodyMetricRepository
	apiKey             string
	model              string
	httpClient         *http.Client
}

func NewRoutineAIService(
	repo repository.RoutineRepository,
	workoutSessionRepo repository.WorkoutSessionRepository,
	bodyMetricRepo repository.BodyMetricRepository,
	apiKey, model string,
) *RoutineAIService {
	model = strings.TrimSpace(model)
	if model == "" {
		model = "gemini-1.5-flash"
	}

	return &RoutineAIService{
		repo:               repo,
		workoutSessionRepo: workoutSessionRepo,
		bodyMetricRepo:     bodyMetricRepo,
		apiKey:             strings.TrimSpace(apiKey),
		model:              model,
		httpClient:         &http.Client{Timeout: 25 * time.Second},
	}
}

func (s *RoutineAIService) GenerateRoutineJSON(
	ctx context.Context,
	userID string,
	req model.AIRoutineGenerationRequest,
) (model.AIRoutineGenerateResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return model.AIRoutineGenerateResponse{}, ErrAIRoutineInvalidInput
	}

	req.Objective = strings.TrimSpace(req.Objective)
	if req.Objective == "" || req.DurationMinutes <= 0 {
		return model.AIRoutineGenerateResponse{}, ErrAIRoutineInvalidInput
	}

	now := time.Now().UTC()
	since := now.Add(-aiRoutineRateWindow)
	used, err := s.repo.CountAIGenerationsInWindow(ctx, userID, since)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	resetAt := now.Add(aiRoutineRateWindow)

	if used >= aiRoutineRateLimit {
		return model.AIRoutineGenerateResponse{
			RateLimit: model.AIRoutineRateLimitStatus{
				Limit:               aiRoutineRateLimit,
				Remaining:           0,
				UsedInCurrentWindow: used,
				WindowSeconds:       int(aiRoutineRateWindow.Seconds()),
				ResetAt:             resetAt,
			},
		}, ErrAIRoutineRateLimited
	}

	exercises, err := s.repo.ListAvailableExercisesForAI(ctx, userID, req.TargetMuscleGroups, 200)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	userContext, err := s.buildUserContext(ctx, userID, now)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	generated, err := s.generateWithGemini(ctx, req, exercises, userContext, now)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	if err := s.repo.LogAIGeneration(ctx, userID, now); err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	return model.AIRoutineGenerateResponse{
		RoutineJSON: generated,
		RateLimit: model.AIRoutineRateLimitStatus{
			Limit:               aiRoutineRateLimit,
			Remaining:           aiRoutineRateLimit - (used + 1),
			UsedInCurrentWindow: used + 1,
			WindowSeconds:       int(aiRoutineRateWindow.Seconds()),
			ResetAt:             resetAt,
		},
	}, nil
}

func (s *RoutineAIService) generateWithGemini(
	ctx context.Context,
	req model.AIRoutineGenerationRequest,
	exercises []model.Exercise,
	userContext routineAIUserContext,
	now time.Time,
) (model.AIRoutineJSON, error) {
	if s.apiKey == "" {
		return model.AIRoutineJSON{}, ErrAIRoutineMissingAPIKey
	}

	exerciseCatalog := make([]map[string]string, 0, len(exercises))
	for _, exercise := range exercises {
		exerciseCatalog = append(exerciseCatalog, map[string]string{
			"id":            exercise.ID,
			"name":          exercise.Name,
			"muscle_group":  exercise.MuscleGroup,
			"exercise_type": exercise.ExerciseType,
		})
	}

	inputPayload := map[string]any{
		"objective":              req.Objective,
		"duration_minutes":       req.DurationMinutes,
		"target_muscle_groups":   normalizeTextList(req.TargetMuscleGroups),
		"mandatory_exercise_ids": normalizeTextList(req.MandatoryExerciseIDs),
		"user_context":           userContext,
		"exercise_catalog":       exerciseCatalog,
		"output_contract": map[string]any{
			"name":              "string",
			"objective":         "string",
			"duration_minutes":  "number",
			"target_muscles":    []string{},
			"mandatory_count":   "number",
			"generated_at":      "RFC3339 datetime string",
			"generation_source": "string",
			"exercises": []map[string]string{
				{
					"exercise_id":      "string",
					"name":             "string",
					"muscle_group":     "string",
					"exercise_type":    "string",
					"is_mandatory":     "boolean",
					"recommended_sets": "number",
					"recommended_reps": "string",
				},
			},
		},
	}

	systemInstruction := "You are a workout planner. Use user_context, especially recent_training_history, as the main history signal. Return only valid JSON matching output_contract. Do not include markdown."
	userPromptBytes, _ := json.Marshal(inputPayload)

	requestBody := map[string]any{
		"system_instruction": map[string]any{
			"parts": []map[string]string{
				{"text": systemInstruction},
			},
		},
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": string(userPromptBytes)},
				},
			},
		},
		"generationConfig": map[string]any{
			"temperature":      0.3,
			"responseMimeType": "application/json",
		},
	}

	bodyBytes, _ := json.Marshal(requestBody)
	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		url.PathEscape(s.model),
		url.QueryEscape(s.apiKey),
	)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return model.AIRoutineJSON{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return model.AIRoutineJSON{}, err
	}
	defer httpResp.Body.Close()

	responseBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return model.AIRoutineJSON{}, err
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return model.AIRoutineJSON{}, fmt.Errorf("%w: gemini status %d", ErrAIRoutineProviderUnavailable, httpResp.StatusCode)
	}

	var geminiResp geminiGenerateContentResponse
	if err := json.Unmarshal(responseBody, &geminiResp); err != nil {
		return model.AIRoutineJSON{}, err
	}

	jsonText := extractGeminiText(geminiResp)
	if strings.TrimSpace(jsonText) == "" {
		return model.AIRoutineJSON{}, ErrAIRoutineProviderUnavailable
	}

	var generated model.AIRoutineJSON
	if err := json.Unmarshal([]byte(jsonText), &generated); err != nil {
		return model.AIRoutineJSON{}, err
	}

	if strings.TrimSpace(generated.Objective) == "" {
		generated.Objective = req.Objective
	}
	if generated.DurationMinutes <= 0 {
		generated.DurationMinutes = req.DurationMinutes
	}
	if len(generated.TargetMuscles) == 0 {
		generated.TargetMuscles = normalizeTextList(req.TargetMuscleGroups)
	}
	if generated.GeneratedAt.IsZero() {
		generated.GeneratedAt = now
	}
	if strings.TrimSpace(generated.GenerationSource) == "" {
		generated.GenerationSource = "gemini"
	}

	return generated, nil
}

func (s *RoutineAIService) buildUserContext(
	ctx context.Context,
	userID string,
	now time.Time,
) (routineAIUserContext, error) {
	userContext := routineAIUserContext{}

	recentWorkouts, err := s.workoutSessionRepo.ListRecentByUser(ctx, userID, 2)
	if err != nil {
		return routineAIUserContext{}, err
	}
	userContext.RecentWorkouts = make([]routineAIRecentWorkout, 0, len(recentWorkouts))
	for _, workout := range recentWorkouts {
		userContext.RecentWorkouts = append(userContext.RecentWorkouts, routineAIRecentWorkout{
			Name:            workout.Name,
			RoutineName:     workout.RoutineName,
			DurationMinutes: workout.DurationMinutes,
			ExerciseCount:   workout.ExerciseCount,
		})
	}

	recentWorkoutHistory, err := s.workoutSessionRepo.ListRecentWorkoutHistoryByUser(ctx, userID, 2)
	if err != nil {
		return routineAIUserContext{}, err
	}
	userContext.RecentTrainingHistory = recentWorkoutHistory

	recentRoutines, err := s.repo.ListRecentByUser(ctx, userID, 2)
	if err != nil {
		return routineAIUserContext{}, err
	}
	userContext.RecentRoutines = make([]routineAIRecentRoutine, 0, len(recentRoutines))
	for _, routine := range recentRoutines {
		userContext.RecentRoutines = append(userContext.RecentRoutines, routineAIRecentRoutine{
			Name:          routine.Name,
			ExerciseCount: routine.ExerciseCount,
		})
	}

	last30Start := now.AddDate(0, 0, -30)
	last30WorkoutDates, err := s.workoutSessionRepo.ListTrainingDatesInRange(ctx, userID, last30Start, now.AddDate(0, 0, 1))
	if err != nil {
		return routineAIUserContext{}, err
	}
	userContext.TrainingDays30d = len(last30WorkoutDates)

	streakRangeStart := now.AddDate(0, 0, -45)
	streakWorkoutDates, err := s.workoutSessionRepo.ListTrainingDatesInRange(ctx, userID, streakRangeStart, now.AddDate(0, 0, 1))
	if err != nil {
		return routineAIUserContext{}, err
	}
	userContext.CurrentStreakDays = calculateCurrentStreak(now, streakWorkoutDates)

	yearStart := now.AddDate(0, 0, -90)
	muscleShares, _, err := s.workoutSessionRepo.ListMuscleDistributionByUser(ctx, userID, yearStart, now.AddDate(0, 0, 1))
	if err != nil {
		return routineAIUserContext{}, err
	}
	userContext.TopMuscleGroups = make([]routineAIMuscleGroupShare, 0, minInt(3, len(muscleShares)))
	for index, share := range muscleShares {
		if index >= 3 {
			break
		}
		userContext.TopMuscleGroups = append(userContext.TopMuscleGroups, routineAIMuscleGroupShare{
			Name:       share.Name,
			Count:      share.Count,
			Percentage: share.Percentage,
		})
	}

	bodyMetrics, err := s.bodyMetricRepo.ListRecentByUser(ctx, userID, 2)
	if err != nil {
		return routineAIUserContext{}, err
	}
	userContext.BodyMetrics = buildRoutineAIBodyMetrics(bodyMetrics)

	return userContext, nil
}

func normalizeTextList(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		normalized = append(normalized, trimmed)
		seen[key] = struct{}{}
	}
	return normalized
}

type routineAIUserContext struct {
	TrainingDays30d       int                                   `json:"training_days_30d"`
	CurrentStreakDays     int                                   `json:"current_streak_days"`
	RecentWorkouts        []routineAIRecentWorkout              `json:"recent_workouts,omitempty"`
	RecentTrainingHistory []model.AIRoutineRecentWorkoutSession `json:"recent_training_history,omitempty"`
	RecentRoutines        []routineAIRecentRoutine              `json:"recent_routines,omitempty"`
	TopMuscleGroups       []routineAIMuscleGroupShare           `json:"top_muscle_groups,omitempty"`
	BodyMetrics           *routineAIBodyMetrics                 `json:"body_metrics,omitempty"`
}

type routineAIRecentWorkout struct {
	Name            string `json:"name"`
	RoutineName     string `json:"routine_name,omitempty"`
	DurationMinutes int    `json:"duration_minutes"`
	ExerciseCount   int    `json:"exercise_count"`
}

type routineAIRecentRoutine struct {
	Name          string `json:"name"`
	ExerciseCount int    `json:"exercise_count"`
}

type routineAIMuscleGroupShare struct {
	Name       string `json:"name"`
	Count      int    `json:"count"`
	Percentage int    `json:"percentage"`
}

type routineAIBodyMetrics struct {
	LastRecordedAt         *time.Time `json:"last_recorded_at,omitempty"`
	WeightKg               *float64   `json:"weight_kg,omitempty"`
	WeightKgDelta          *float64   `json:"weight_kg_delta,omitempty"`
	BodyFatPercentage      *float64   `json:"body_fat_percentage,omitempty"`
	BodyFatPercentageDelta *float64   `json:"body_fat_percentage_delta,omitempty"`
	MuscleMassKg           *float64   `json:"muscle_mass_kg,omitempty"`
	MuscleMassKgDelta      *float64   `json:"muscle_mass_kg_delta,omitempty"`
}

type geminiGenerateContentResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func extractGeminiText(response geminiGenerateContentResponse) string {
	if len(response.Candidates) == 0 {
		return ""
	}
	parts := response.Candidates[0].Content.Parts
	if len(parts) == 0 {
		return ""
	}
	return parts[0].Text
}

func buildRoutineAIBodyMetrics(entries []model.OverviewBodyMetricEntry) *routineAIBodyMetrics {
	if len(entries) == 0 {
		return nil
	}

	current := entries[0]
	var previous *model.OverviewBodyMetricEntry
	if len(entries) > 1 {
		previous = &entries[1]
	}

	metrics := &routineAIBodyMetrics{
		LastRecordedAt:    &current.RecordedAt,
		WeightKg:          current.WeightKg,
		BodyFatPercentage: current.BodyFatPercentage,
		MuscleMassKg:      current.MuscleMassKg,
	}

	if current.WeightKg != nil && previous != nil && previous.WeightKg != nil {
		delta := *current.WeightKg - *previous.WeightKg
		metrics.WeightKgDelta = &delta
	}
	if current.BodyFatPercentage != nil && previous != nil && previous.BodyFatPercentage != nil {
		delta := *current.BodyFatPercentage - *previous.BodyFatPercentage
		metrics.BodyFatPercentageDelta = &delta
	}
	if current.MuscleMassKg != nil && previous != nil && previous.MuscleMassKg != nil {
		delta := *current.MuscleMassKg - *previous.MuscleMassKg
		metrics.MuscleMassKgDelta = &delta
	}

	return metrics
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
