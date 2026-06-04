package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/jackc/pgx/v5"
)

var (
	// ErrAIRoutineInvalidInput indicates that the generation request is incomplete or invalid.
	ErrAIRoutineInvalidInput = errors.New("invalid ai routine input")
	// ErrAIRoutineProviderUnavailable indicates that Gemini could not return a usable routine.
	ErrAIRoutineProviderUnavailable = errors.New("ai provider unavailable")
	// ErrAIRoutineMissingAPIKey indicates that Gemini credentials are not configured.
	ErrAIRoutineMissingAPIKey = errors.New("ai provider missing api key")
	// ErrAIRoutineRateLimited indicates that the user exhausted the AI routine quota.
	ErrAIRoutineRateLimited = errors.New("ai routine rate limited")
)

const (
	aiRoutineRateLimit = 20
)

var aiRoutineRateWindow = time.Hour

// RoutineAIService generates AI routine JSON and persists AI routines.
type RoutineAIService struct {
	repo               repository.RoutineRepository
	exerciseService    *ExerciseService
	workoutSessionRepo repository.WorkoutSessionRepository
	bodyMetricRepo     repository.BodyMetricRepository
	apiKey             string
	model              string
	httpClient         *http.Client
}

// NewRoutineAIService creates a service that generates and persists AI routines.
func NewRoutineAIService(
	repo repository.RoutineRepository,
	exerciseService *ExerciseService,
	workoutSessionRepo repository.WorkoutSessionRepository,
	bodyMetricRepo repository.BodyMetricRepository,
	apiKey, model string,
) *RoutineAIService {
	model = strings.TrimSpace(model)
	if model == "" {
		model = "gemini-2.5-flash"
	}

	return &RoutineAIService{
		repo:               repo,
		exerciseService:    exerciseService,
		workoutSessionRepo: workoutSessionRepo,
		bodyMetricRepo:     bodyMetricRepo,
		apiKey:             strings.TrimSpace(apiKey),
		model:              model,
		httpClient:         &http.Client{Timeout: 25 * time.Second},
	}
}

// GenerateRoutineJSON builds user context, calls Gemini, and returns the generated JSON for preview.
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
	used, resetAt, err := s.enforceAIRoutineRateLimit(ctx, userID, now)
	if err != nil {
		if errors.Is(err, ErrAIRoutineRateLimited) {
			rateLimit := buildAIRoutineRateLimitStatus(used, resetAt)
			return model.AIRoutineGenerateResponse{
				RateLimit: &rateLimit,
			}, err
		}
		return model.AIRoutineGenerateResponse{}, err
	}
	slog.Info("ai routine generation started",
		"user_id", userID,
		"objective", req.Objective,
		"duration_minutes", req.DurationMinutes,
		"target_muscle_groups", normalizeTextList(req.TargetMuscleGroups),
		"mandatory_exercises_count", len(normalizeTextList(req.MandatoryExercises)),
		"notes_present", strings.TrimSpace(req.Notes) != "",
	)

	exercises, err := s.repo.ListAvailableExercisesForAI(ctx, userID, req.TargetMuscleGroups, 200)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	userContext, err := s.buildUserContext(ctx, userID, now)
	if err != nil {
		slog.Error("ai routine user context build failed", "user_id", userID, "error", err)
		return model.AIRoutineGenerateResponse{}, err
	}

	generated, err := s.generateWithGemini(ctx, req, exercises, userContext, now)
	if err != nil {
		slog.Error("ai routine generation failed", "user_id", userID, "error", err)
		return model.AIRoutineGenerateResponse{}, err
	}

	if err := s.repo.LogAIGeneration(ctx, userID, now); err != nil {
		slog.Warn("ai routine generation log skipped", "user_id", userID, "error", err)
	}

	slog.Info("ai routine generation finished",
		"user_id", userID,
		"exercise_count", len(generated.Exercises),
		"source", generated.GenerationSource,
		"objective", generated.Objective,
	)

	return model.AIRoutineGenerateResponse{
		RoutineJSON: generated,
		RateLimit:   rateLimitStatusPointer(buildAIRoutineRateLimitStatus(used+1, resetAt)),
	}, nil
}

// SaveGeneratedRoutineJSON resolves or creates exercises from a generated preview and persists the routine.
func (s *RoutineAIService) SaveGeneratedRoutineJSON(
	ctx context.Context,
	userID string,
	generated model.AIRoutineJSON,
) (model.AIRoutineGenerateResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" || s.exerciseService == nil {
		return model.AIRoutineGenerateResponse{}, ErrAIRoutineInvalidInput
	}

	routineID, err := s.saveGeneratedRoutine(ctx, userID, generated)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	return model.AIRoutineGenerateResponse{
		RoutineJSON: generated,
		RoutineID:   routineID,
	}, nil
}

// SaveUpgradedRoutineAsNew persists an upgrade proposal as a brand new routine.
func (s *RoutineAIService) SaveUpgradedRoutineAsNew(
	ctx context.Context,
	userID, baseRoutineID string,
	generated model.AIRoutineJSON,
) (model.AIRoutineGenerateResponse, error) {
	userID = strings.TrimSpace(userID)
	baseRoutineID = strings.TrimSpace(baseRoutineID)
	if userID == "" || baseRoutineID == "" || s.exerciseService == nil {
		return model.AIRoutineGenerateResponse{}, ErrAIRoutineInvalidInput
	}

	if _, err := s.repo.GetByID(ctx, userID, baseRoutineID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AIRoutineGenerateResponse{}, ErrRoutineNotFound
		}
		return model.AIRoutineGenerateResponse{}, err
	}

	routineID, err := s.saveGeneratedRoutine(ctx, userID, generated)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	return model.AIRoutineGenerateResponse{
		RoutineJSON: generated,
		RoutineID:   routineID,
	}, nil
}

// OverwriteRoutineWithGeneratedJSON replaces one existing routine with the upgraded proposal.
func (s *RoutineAIService) OverwriteRoutineWithGeneratedJSON(
	ctx context.Context,
	userID, routineID string,
	generated model.AIRoutineJSON,
) (model.AIRoutineGenerateResponse, error) {
	userID = strings.TrimSpace(userID)
	routineID = strings.TrimSpace(routineID)
	if userID == "" || routineID == "" || s.exerciseService == nil {
		return model.AIRoutineGenerateResponse{}, ErrAIRoutineInvalidInput
	}

	if _, err := s.repo.GetByID(ctx, userID, routineID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AIRoutineGenerateResponse{}, ErrRoutineNotFound
		}
		return model.AIRoutineGenerateResponse{}, err
	}

	routineToSave, err := s.buildGeneratedRoutineToSave(ctx, userID, generated)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	if err := s.repo.OverwriteGeneratedAIRoutine(ctx, routineID, userID, routineToSave); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AIRoutineGenerateResponse{}, ErrRoutineNotFound
		}
		return model.AIRoutineGenerateResponse{}, err
	}

	return model.AIRoutineGenerateResponse{
		RoutineJSON: generated,
		RoutineID:   routineID,
	}, nil
}

// UpgradeRoutineJSON builds a proposed upgrade for one existing routine without persisting it.
func (s *RoutineAIService) UpgradeRoutineJSON(
	ctx context.Context,
	userID, routineID string,
	req model.AIRoutineUpgradeRequest,
) (model.AIRoutineUpgradeResponse, error) {
	userID = strings.TrimSpace(userID)
	routineID = strings.TrimSpace(routineID)
	req.Message = strings.TrimSpace(req.Message)
	req.FeedbackMessage = strings.TrimSpace(req.FeedbackMessage)

	if userID == "" || routineID == "" || (req.Message == "" && req.FeedbackMessage == "") {
		return model.AIRoutineUpgradeResponse{}, ErrAIRoutineInvalidInput
	}

	now := time.Now().UTC()
	used, resetAt, err := s.enforceAIRoutineRateLimit(ctx, userID, now)
	if err != nil {
		if errors.Is(err, ErrAIRoutineRateLimited) {
			return model.AIRoutineUpgradeResponse{
				RateLimit: buildAIRoutineRateLimitStatus(used, resetAt),
			}, err
		}
		return model.AIRoutineUpgradeResponse{}, err
	}

	routine, err := s.repo.GetByID(ctx, userID, routineID)
	if err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}
	legacyRoutine := routineDetailToRoutine(*routine, userID)

	exercises, err := s.repo.ListAvailableExercisesForAI(ctx, userID, nil, 200)
	if err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}

	userContext, err := s.buildUpgradeUserContext(ctx, userID, legacyRoutine, now)
	if err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}

	generated, err := s.generateUpgradeWithGemini(ctx, legacyRoutine, req, exercises, userContext, now)
	if err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}

	if _, err := s.buildExercisesToSave(generated, exercises); err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}

	originalRoutineJSON := buildAIRoutineJSONFromRoutine(legacyRoutine, now)
	diff := buildAIRoutineUpgradeDiff(originalRoutineJSON, generated)

	if err := s.repo.LogAIGeneration(ctx, userID, now); err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}

	return model.AIRoutineUpgradeResponse{
		RoutineJSON: generated,
		Diff:        diff,
		RateLimit:   buildAIRoutineRateLimitStatus(used+1, resetAt),
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

	systemInstruction := "You are a workout planner. Use user_context, especially recent_training_history, as the main history signal. Respect user_notes and mandatory_exercises as strong instructions from the user. Build the most complete and sensible routine possible for the available time, choosing the exercise count freely based on the objective, time available, and user requests. Do not force a one-to-one mapping between target muscle groups and exercises, and do not use a fixed exercise count. Prefer the best coverage and exercise selection for the routine as a whole. If mandatory_exercises is empty, do not split exercises into optional vs mandatory; treat every exercise in the routine as required. If mandatory_exercises is not empty, mark the requested exercises as mandatory and keep the rest as non-mandatory. Return only valid JSON matching output_contract. Put planned sets in exercises[].sets. Use target_weight_kg only when recent history supports it; otherwise use null or omit it. Do not include markdown."
	requestBody, err := buildGeminiRoutineRequestBody(
		systemInstruction,
		req,
		exercises,
		userContext,
	)
	if err != nil {
		return model.AIRoutineJSON{}, err
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return model.AIRoutineJSON{}, fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
	}
	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		url.PathEscape(s.model),
		url.QueryEscape(s.apiKey),
	)

	requestStartedAt := time.Now()
	slog.Info("ai routine gemini request started",
		"model", s.model,
		"exercise_catalog_count", len(exercises),
		"mandatory_exercises_count", len(normalizeTextList(req.MandatoryExercises)),
		"notes_present", strings.TrimSpace(req.Notes) != "",
		"user_context_recent_workouts", len(userContext.RecentWorkouts),
		"user_context_recent_training_sessions", len(userContext.RecentTrainingHistory),
		"user_context_recent_routines", len(userContext.RecentRoutines),
	)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return model.AIRoutineJSON{}, fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return model.AIRoutineJSON{}, fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
	}
	defer httpResp.Body.Close()

	responseBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return model.AIRoutineJSON{}, fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
	}
	slog.Info("ai routine gemini response received",
		"model", s.model,
		"status_code", httpResp.StatusCode,
		"duration_ms", time.Since(requestStartedAt).Milliseconds(),
		"response_bytes", len(responseBody),
	)

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		slog.Error("ai routine gemini response status error",
			"model", s.model,
			"status_code", httpResp.StatusCode,
			"response_snippet", truncateProviderError(responseBody),
		)
		return model.AIRoutineJSON{}, fmt.Errorf(
			"%w: gemini status %d: %s",
			ErrAIRoutineProviderUnavailable,
			httpResp.StatusCode,
			truncateProviderError(responseBody),
		)
	}

	var geminiResp geminiGenerateContentResponse
	if err := json.Unmarshal(responseBody, &geminiResp); err != nil {
		return model.AIRoutineJSON{}, fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
	}

	jsonText := extractGeminiText(geminiResp)
	if strings.TrimSpace(jsonText) == "" {
		slog.Error("ai routine gemini response missing text", "model", s.model)
		return model.AIRoutineJSON{}, ErrAIRoutineProviderUnavailable
	}

	generated, err := finalizeGeneratedAIRoutine(jsonText, req, now)
	if err != nil {
		return model.AIRoutineJSON{}, err
	}
	slog.Info("ai routine gemini response parsed",
		"model", s.model,
		"exercise_count", len(generated.Exercises),
		"duration_minutes", generated.DurationMinutes,
	)

	return generated, nil
}

func (s *RoutineAIService) generateUpgradeWithGemini(
	ctx context.Context,
	routine model.Routine,
	req model.AIRoutineUpgradeRequest,
	exercises []model.Exercise,
	userContext routineAIUserContext,
	now time.Time,
) (model.AIRoutineJSON, error) {
	if s.apiKey == "" {
		return model.AIRoutineJSON{}, ErrAIRoutineMissingAPIKey
	}

	inputPayload := map[string]any{
		"message":          req.Message,
		"feedback_message": req.FeedbackMessage,
		"user_context":     userContext,
		"existing_routine": buildAIRoutineJSONFromRoutine(routine, now),
		"exercise_catalog": buildAIExerciseCatalog(exercises),
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
					"exercise_id":   "string",
					"name":          "string",
					"muscle_group":  "string",
					"exercise_type": "string",
					"is_mandatory":  "boolean",
					"sets":          "array of planned sets with set_number, target_reps_min, target_reps_max, target_reps_text, target_weight_kg, target_rir, rest_seconds, notes",
				},
			},
		},
	}

	systemInstruction := "You are a workout planner improving an existing routine. Use user_context, especially recent_training_history, as the main history signal. Treat existing_routine as the base plan to refine, not to discard without a good reason. Respect message and feedback_message as strong user instructions. Keep the response fully compatible with the existing routine format. Preserve exercises that are already working unless user guidance or training context suggests a meaningful change. Prefer targeted improvements in exercise selection, order, set structure, fatigue management, and muscle balance over random rewrites. Prefer the provided exercise_catalog and avoid inventing unsupported exercises. Return only valid JSON matching output_contract. Return only the upgraded routine object itself. Never return or repeat the input payload. Never include keys such as exercise_catalog, existing_routine, user_context, message, feedback_message, or output_contract in the response. Put planned sets in exercises[].sets. Use target_weight_kg only when recent history supports it; otherwise use null or omit it. Do not include markdown."

	generated, err := s.callGeminiForRoutineJSON(ctx, inputPayload, systemInstruction, now)
	if err != nil {
		return model.AIRoutineJSON{}, err
	}

	if strings.TrimSpace(generated.Objective) == "" {
		generated.Objective = fmt.Sprintf("Upgrade of %s", strings.TrimSpace(routine.Name))
	}
	if generated.DurationMinutes <= 0 {
		generated.DurationMinutes = estimateRoutineDurationMinutes(routine)
	}
	if len(generated.TargetMuscles) == 0 {
		generated.TargetMuscles = extractRoutineTargetMuscles(&routine)
	}
	if len(generated.Exercises) == 0 {
		return model.AIRoutineJSON{}, ErrAIRoutineProviderUnavailable
	}
	generated.MandatoryCount = countMandatoryExercises(generated.Exercises)

	return generated, nil
}

func (s *RoutineAIService) callGeminiForRoutineJSON(
	ctx context.Context,
	inputPayload map[string]any,
	systemInstruction string,
	now time.Time,
) (model.AIRoutineJSON, error) {
	if s.apiKey == "" {
		return model.AIRoutineJSON{}, ErrAIRoutineMissingAPIKey
	}

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

	if prettyJSON := prettyJSONForLog(jsonText); prettyJSON != "" {
		slog.Info("ai routine gemini raw response",
			"response", truncateLogValue(prettyJSON, 8000),
		)
	}

	if looksLikePromptEcho(jsonText) {
		slog.Error("ai routine gemini response echoed input payload")
		return model.AIRoutineJSON{}, ErrAIRoutineProviderUnavailable
	}

	var generated model.AIRoutineJSON
	if err := json.Unmarshal([]byte(jsonText), &generated); err != nil {
		return model.AIRoutineJSON{}, err
	}

	generated.GeneratedAt = now
	if strings.TrimSpace(generated.GenerationSource) == "" {
		generated.GenerationSource = "gemini"
	}

	return generated, nil
}

func (s *RoutineAIService) saveGeneratedRoutine(
	ctx context.Context,
	userID string,
	generated model.AIRoutineJSON,
) (string, error) {
	routineToSave, err := s.buildGeneratedRoutineToSave(ctx, userID, generated)
	if err != nil {
		return "", err
	}

	return s.repo.SaveGeneratedAIRoutine(ctx, routineToSave)
}

func (s *RoutineAIService) buildGeneratedRoutineToSave(
	ctx context.Context,
	userID string,
	generated model.AIRoutineJSON,
) (model.AIRoutineToSave, error) {
	seenExerciseIDs := make(map[string]struct{}, len(generated.Exercises))
	exercisesToSave := make([]model.AIRoutineExerciseToSave, 0, len(generated.Exercises))
	for _, exercise := range generated.Exercises {
		exerciseID, err := s.resolveOrCreateGeneratedAIRoutineExerciseID(ctx, userID, exercise)
		if err != nil {
			return model.AIRoutineToSave{}, err
		}
		if exerciseID == "" {
			continue
		}
		if _, exists := seenExerciseIDs[exerciseID]; exists {
			continue
		}

		exercisesToSave = append(exercisesToSave, model.AIRoutineExerciseToSave{
			ExerciseID: exerciseID,
			Order:      len(exercisesToSave) + 1,
			Notes:      buildAIRoutineExerciseNotes(exercise),
			Sets:       buildAIRoutineExerciseSetsToSave(exercise),
		})
		seenExerciseIDs[exerciseID] = struct{}{}
	}
	if len(exercisesToSave) == 0 {
		return model.AIRoutineToSave{}, ErrAIRoutineProviderUnavailable
	}

	routineName := strings.TrimSpace(generated.Name)
	if routineName == "" {
		routineName = "AI generated routine"
	}

	return model.AIRoutineToSave{
		UserID:      userID,
		Name:        routineName,
		Description: buildAIRoutineDescription(generated),
		Exercises:   exercisesToSave,
	}, nil
}

func (s *RoutineAIService) buildExercisesToSave(
	generated model.AIRoutineJSON,
	availableExercises []model.Exercise,
) ([]model.AIRoutineExerciseToSave, error) {
	availableExerciseIDs := make(map[string]struct{}, len(availableExercises))
	for _, exercise := range availableExercises {
		availableExerciseIDs[exercise.ID] = struct{}{}
	}

	seenExerciseIDs := make(map[string]struct{}, len(generated.Exercises))
	exercisesToSave := make([]model.AIRoutineExerciseToSave, 0, len(generated.Exercises))
	for _, exercise := range generated.Exercises {
		exerciseID := strings.TrimSpace(exercise.ExerciseID)
		if exerciseID == "" {
			continue
		}
		if _, ok := availableExerciseIDs[exerciseID]; !ok {
			return nil, fmt.Errorf("%w: generated unknown exercise id %s", ErrAIRoutineProviderUnavailable, exerciseID)
		}
		if _, exists := seenExerciseIDs[exerciseID]; exists {
			continue
		}

		exercisesToSave = append(exercisesToSave, model.AIRoutineExerciseToSave{
			ExerciseID: exerciseID,
			Order:      len(exercisesToSave) + 1,
			Notes:      buildAIRoutineExerciseNotes(exercise),
			Sets:       buildAIRoutineExerciseSetsToSave(exercise),
		})
		seenExerciseIDs[exerciseID] = struct{}{}
	}

	if len(exercisesToSave) == 0 {
		return nil, ErrAIRoutineProviderUnavailable
	}

	return exercisesToSave, nil
}

func (s *RoutineAIService) resolveOrCreateGeneratedAIRoutineExerciseID(
	ctx context.Context,
	userID string,
	exercise model.AIRoutineExercise,
) (string, error) {
	exerciseID := strings.TrimSpace(exercise.ExerciseID)
	if exerciseID != "" {
		if _, err := s.exerciseService.GetByID(ctx, exerciseID); err == nil {
			return exerciseID, nil
		}
	}

	if existing, err := s.findExistingGeneratedExercise(ctx, exercise); err != nil {
		return "", err
	} else if existing != nil {
		return existing.ID, nil
	}

	name := normalizeName(exercise.Name)
	if name == "" {
		return "", nil
	}

	muscleGroup := normalizeDomainValue(exercise.MuscleGroup)
	newExercise := model.Exercise{
		Name:         name,
		MuscleGroup:  muscleGroup,
		ExerciseType: normalizeAIExerciseTypeForCreation(exercise.ExerciseType),
		IsOfficial:   false,
	}
	newExercise.OwnerUserID = &userID

	if err := s.exerciseService.Create(ctx, &newExercise); err != nil {
		return "", err
	}

	return newExercise.ID, nil
}

func (s *RoutineAIService) findExistingGeneratedExercise(
	ctx context.Context,
	exercise model.AIRoutineExercise,
) (*model.Exercise, error) {
	name := normalizeName(exercise.Name)
	muscleGroup := normalizeDomainValue(exercise.MuscleGroup)
	exerciseType := normalizeAIExerciseTypeForLookup(exercise.ExerciseType)

	if name == "" {
		return nil, nil
	}

	filters := model.ExerciseFilter{
		Search:      name,
		Type:        exerciseType,
		MuscleGroup: muscleGroup,
		Page:        1,
		Limit:       100,
	}

	exercises, err := s.exerciseService.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	for _, candidate := range exercises.Items {
		if normalizeName(candidate.Name) != name {
			continue
		}
		if muscleGroup != "" && normalizeDomainValue(candidate.MuscleGroup) != muscleGroup {
			continue
		}
		if exerciseType != "" && normalizeDomainValue(candidate.ExerciseType) != exerciseType {
			continue
		}
		return &candidate, nil
	}

	return nil, nil
}

func normalizeAIExerciseTypeForLookup(value string) string {
	normalized := normalizeDomainValue(value)
	if normalized == "" {
		return ""
	}
	if !isValidExerciseType(normalized) {
		return ""
	}
	return normalized
}

func normalizeAIExerciseTypeForCreation(value string) string {
	normalized := normalizeDomainValue(value)
	if normalized == "" || !isValidExerciseType(normalized) {
		return ExerciseTypeStrength
	}
	return normalized
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

func (s *RoutineAIService) buildUpgradeUserContext(
	ctx context.Context,
	userID string,
	routine model.Routine,
	now time.Time,
) (routineAIUserContext, error) {
	userContext, err := s.buildUserContext(ctx, userID, now)
	if err != nil {
		return routineAIUserContext{}, err
	}

	signals := buildRoutineRelationSignals(routine)
	if signals.isEmpty() {
		return userContext, nil
	}

	filteredHistory := make([]model.AIRoutineRecentWorkoutSession, 0, len(userContext.RecentTrainingHistory))
	filteredWorkouts := make([]routineAIRecentWorkout, 0, len(userContext.RecentWorkouts))

	for _, session := range userContext.RecentTrainingHistory {
		filteredSession, matched := filterWorkoutSessionByRelation(session, signals)
		if !matched {
			continue
		}
		filteredHistory = append(filteredHistory, filteredSession)
		filteredWorkouts = append(filteredWorkouts, routineAIRecentWorkout{
			Name:            filteredSession.SessionName,
			RoutineName:     filteredSession.RoutineName,
			DurationMinutes: filteredSession.DurationMinutes,
			ExerciseCount:   len(filteredSession.Exercises),
		})
	}

	if len(filteredHistory) > 0 {
		userContext.RecentTrainingHistory = filteredHistory
		userContext.RecentWorkouts = filteredWorkouts
	}

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

func buildGeminiRoutineRequestBody(
	systemInstruction string,
	req model.AIRoutineGenerationRequest,
	exercises []model.Exercise,
	userContext routineAIUserContext,
) (map[string]any, error) {
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
		"objective":            req.Objective,
		"duration_minutes":     req.DurationMinutes,
		"target_muscle_groups": normalizeTextList(req.TargetMuscleGroups),
		"mandatory_exercises":  normalizeTextList(req.MandatoryExercises),
		"user_notes":           strings.TrimSpace(req.Notes),
		"user_context":         userContext,
		"exercise_catalog":     exerciseCatalog,
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
					"exercise_id":   "string",
					"name":          "string",
					"muscle_group":  "string",
					"exercise_type": "string",
					"is_mandatory":  "boolean",
					"sets":          "array of planned sets with set_number, target_reps_min, target_reps_max, target_reps_text, target_weight_kg, target_rir, rest_seconds, notes",
				},
			},
		},
	}

	userPromptBytes, err := json.Marshal(inputPayload)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
	}

	return map[string]any{
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
	}, nil
}

func finalizeGeneratedAIRoutine(jsonText string, req model.AIRoutineGenerationRequest, now time.Time) (model.AIRoutineJSON, error) {
	if prettyJSON := prettyJSONForLog(jsonText); prettyJSON != "" {
		slog.Info("ai routine gemini raw response",
			"response", truncateLogValue(prettyJSON, 8000),
		)
	}

	var generated model.AIRoutineJSON
	if err := json.Unmarshal([]byte(jsonText), &generated); err != nil {
		return model.AIRoutineJSON{}, fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
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
	normalizeGeneratedRoutineMandatoryFlags(&generated, req.MandatoryExercises)
	generated.GeneratedAt = now
	generated.GenerationSource = "gemini"
	generated.MandatoryCount = countMandatoryExercises(generated.Exercises)

	return generated, nil
}

func prettyJSONForLog(jsonText string) string {
	var prettyResponse any
	if err := json.Unmarshal([]byte(jsonText), &prettyResponse); err != nil {
		return strings.TrimSpace(jsonText)
	}
	formatted, err := json.MarshalIndent(prettyResponse, "", "  ")
	if err != nil {
		return strings.TrimSpace(jsonText)
	}
	return string(formatted)
}

func countMandatoryExercises(exercises []model.AIRoutineExercise) int {
	count := 0
	for _, exercise := range exercises {
		if exercise.IsMandatory {
			count++
		}
	}
	return count
}

func looksLikePromptEcho(jsonText string) bool {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonText), &payload); err != nil {
		return false
	}

	echoKeys := []string{
		"exercise_catalog",
		"existing_routine",
		"user_context",
		"message",
		"feedback_message",
		"output_contract",
	}

	for _, key := range echoKeys {
		if _, ok := payload[key]; ok {
			return true
		}
	}

	return false
}

func normalizeGeneratedRoutineMandatoryFlags(generated *model.AIRoutineJSON, mandatoryExercises []string) {
	if generated == nil || len(generated.Exercises) == 0 {
		return
	}

	mandatorySet := make(map[string]struct{}, len(mandatoryExercises))
	for _, value := range normalizeTextList(mandatoryExercises) {
		mandatorySet[normalizeName(value)] = struct{}{}
	}

	// If the user didn't request any mandatory exercises, keep the routine fully required.
	if len(mandatorySet) == 0 {
		for index := range generated.Exercises {
			generated.Exercises[index].IsMandatory = true
		}
		return
	}

	for index := range generated.Exercises {
		_, isMandatory := mandatorySet[normalizeName(generated.Exercises[index].Name)]
		generated.Exercises[index].IsMandatory = isMandatory
	}
}

func buildAIRoutineDescription(generated model.AIRoutineJSON) string {
	objective := strings.TrimSpace(generated.Objective)

	parts := []string{"Generated by Gemini"}
	if objective != "" {
		parts = append(parts, "Objective: "+objective)
	}
	if generated.DurationMinutes > 0 {
		parts = append(parts, fmt.Sprintf("Duration: %d minutes", generated.DurationMinutes))
	}

	return strings.Join(parts, ". ")
}

func buildAIRoutineExerciseNotes(exercise model.AIRoutineExercise) string {
	parts := make([]string, 0, 3)
	if len(exercise.Sets) > 0 {
		parts = append(parts, fmt.Sprintf("%d planned sets", len(exercise.Sets)))
	} else if exercise.RecommendedSets > 0 {
		parts = append(parts, fmt.Sprintf("%d sets", exercise.RecommendedSets))
	}
	if strings.TrimSpace(exercise.RecommendedReps) != "" {
		parts = append(parts, strings.TrimSpace(exercise.RecommendedReps))
	}
	if exercise.IsMandatory {
		parts = append(parts, "mandatory")
	}
	return strings.Join(parts, " | ")
}

func buildAIRoutineExerciseSetsToSave(exercise model.AIRoutineExercise) []model.AIRoutineExerciseSetToSave {
	if len(exercise.Sets) > 0 {
		sets := make([]model.AIRoutineExerciseSetToSave, 0, len(exercise.Sets))
		for index, set := range exercise.Sets {
			setNumber := set.SetNumber
			if setNumber <= 0 {
				setNumber = index + 1
			}
			repsMin, repsMax := normalizedRepsRange(set.TargetRepsMin, set.TargetRepsMax)
			sets = append(sets, model.AIRoutineExerciseSetToSave{
				SetNumber:             setNumber,
				TargetRepsMin:         repsMin,
				TargetRepsMax:         repsMax,
				TargetRepsText:        strings.TrimSpace(set.TargetRepsText),
				TargetWeightKg:        nonNegativeFloatPointer(set.TargetWeightKg),
				TargetDurationSeconds: nonNegativeIntPointer(set.TargetDurationSeconds),
				TargetDistanceKm:      nonNegativeFloatPointer(set.TargetDistanceKm),
				TargetRir:             rirPointer(set.TargetRir),
				RestSeconds:           nonNegativeIntPointer(set.RestSeconds),
				Notes:                 strings.TrimSpace(set.Notes),
			})
		}
		return sets
	}

	if exercise.RecommendedSets <= 0 {
		return nil
	}

	sets := make([]model.AIRoutineExerciseSetToSave, 0, exercise.RecommendedSets)
	for setNumber := 1; setNumber <= exercise.RecommendedSets; setNumber++ {
		sets = append(sets, model.AIRoutineExerciseSetToSave{
			SetNumber:      setNumber,
			TargetRepsText: strings.TrimSpace(exercise.RecommendedReps),
		})
	}
	return sets
}

func normalizedRepsRange(minValue, maxValue *int) (*int, *int) {
	minValue = nonNegativeIntPointer(minValue)
	maxValue = nonNegativeIntPointer(maxValue)
	if minValue != nil && maxValue != nil && *minValue > *maxValue {
		minValue, maxValue = maxValue, minValue
	}
	return minValue, maxValue
}

func nonNegativeIntPointer(value *int) *int {
	if value == nil || *value < 0 {
		return nil
	}
	return value
}

func nonNegativeFloatPointer(value *float64) *float64 {
	if value == nil || *value < 0 {
		return nil
	}
	return value
}

func rirPointer(value *int) *int {
	if value == nil || *value < 0 || *value > 10 {
		return nil
	}
	return value
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

type routineRelationSignals struct {
	exerciseIDs   map[string]struct{}
	exerciseTypes map[string]struct{}
	muscleGroups  map[string]struct{}
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

func truncateProviderError(body []byte) string {
	const maxLength = 600
	message := strings.TrimSpace(string(body))
	if message == "" {
		return "empty response body"
	}
	if len(message) <= maxLength {
		return message
	}
	return message[:maxLength] + "..."
}

func truncateLogValue(value string, maxLength int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= maxLength {
		return value
	}
	return value[:maxLength] + "..."
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

func buildRoutineRelationSignals(routine model.Routine) routineRelationSignals {
	signals := routineRelationSignals{
		exerciseIDs:   make(map[string]struct{}, len(routine.Exercises)),
		exerciseTypes: make(map[string]struct{}, len(routine.Exercises)),
		muscleGroups:  make(map[string]struct{}, len(routine.Exercises)),
	}

	for _, exercise := range routine.Exercises {
		if id := strings.TrimSpace(exercise.ExerciseID); id != "" {
			signals.exerciseIDs[id] = struct{}{}
		}
		if exerciseType := normalizeDomainValue(exercise.ExerciseType); exerciseType != "" {
			signals.exerciseTypes[exerciseType] = struct{}{}
		}
		if muscleGroup := normalizeDomainValue(exercise.MuscleGroup); muscleGroup != "" {
			signals.muscleGroups[muscleGroup] = struct{}{}
		}
	}

	return signals
}

func (s routineRelationSignals) isEmpty() bool {
	return len(s.exerciseIDs) == 0 && len(s.exerciseTypes) == 0 && len(s.muscleGroups) == 0
}

func filterWorkoutSessionByRelation(
	session model.AIRoutineRecentWorkoutSession,
	signals routineRelationSignals,
) (model.AIRoutineRecentWorkoutSession, bool) {
	filteredExercises := make([]model.AIRoutineRecentWorkoutExercise, 0, len(session.Exercises))
	for _, exercise := range session.Exercises {
		if !isRelatedWorkoutExercise(exercise, signals) {
			continue
		}
		filteredExercises = append(filteredExercises, exercise)
	}

	if len(filteredExercises) == 0 {
		return model.AIRoutineRecentWorkoutSession{}, false
	}

	session.Exercises = filteredExercises
	return session, true
}

func isRelatedWorkoutExercise(
	exercise model.AIRoutineRecentWorkoutExercise,
	signals routineRelationSignals,
) bool {
	if exerciseID := strings.TrimSpace(exercise.ExerciseID); exerciseID != "" {
		if _, ok := signals.exerciseIDs[exerciseID]; ok {
			return true
		}
	}

	if exerciseType := normalizeDomainValue(exercise.ExerciseType); exerciseType != "" {
		if _, ok := signals.exerciseTypes[exerciseType]; ok {
			return true
		}
	}

	if muscleGroup := normalizeDomainValue(exercise.MuscleGroup); muscleGroup != "" {
		if _, ok := signals.muscleGroups[muscleGroup]; ok {
			return true
		}
	}

	return false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *RoutineAIService) enforceAIRoutineRateLimit(
	ctx context.Context,
	userID string,
	now time.Time,
) (int, time.Time, error) {
	since := now.Add(-aiRoutineRateWindow)
	used, err := s.repo.CountAIGenerationsInWindow(ctx, userID, since)
	if err != nil {
		return 0, time.Time{}, err
	}

	resetAt := now.Add(aiRoutineRateWindow)
	if used >= aiRoutineRateLimit {
		return used, resetAt, ErrAIRoutineRateLimited
	}

	return used, resetAt, nil
}

func buildAIRoutineRateLimitStatus(used int, resetAt time.Time) model.AIRoutineRateLimitStatus {
	remaining := aiRoutineRateLimit - used
	if remaining < 0 {
		remaining = 0
	}

	return model.AIRoutineRateLimitStatus{
		Limit:               aiRoutineRateLimit,
		Remaining:           remaining,
		UsedInCurrentWindow: used,
		WindowSeconds:       int(aiRoutineRateWindow.Seconds()),
		ResetAt:             resetAt,
	}
}

func rateLimitStatusPointer(status model.AIRoutineRateLimitStatus) *model.AIRoutineRateLimitStatus {
	return &status
}

func buildAIExerciseCatalog(exercises []model.Exercise) []map[string]string {
	exerciseCatalog := make([]map[string]string, 0, len(exercises))
	for _, exercise := range exercises {
		exerciseCatalog = append(exerciseCatalog, map[string]string{
			"id":            exercise.ID,
			"name":          exercise.Name,
			"muscle_group":  exercise.MuscleGroup,
			"exercise_type": exercise.ExerciseType,
		})
	}
	return exerciseCatalog
}

func buildAIRoutineJSONFromRoutine(routine model.Routine, now time.Time) model.AIRoutineJSON {
	exercises := make([]model.AIRoutineExercise, 0, len(routine.Exercises))
	mandatoryCount := 0
	for _, exercise := range routine.Exercises {
		sets := make([]model.AIRoutineExerciseSet, 0, len(exercise.Sets))
		for _, set := range exercise.Sets {
			sets = append(sets, model.AIRoutineExerciseSet{
				SetNumber:             set.SetNumber,
				TargetRepsMin:         set.TargetRepsMin,
				TargetRepsMax:         set.TargetRepsMax,
				TargetRepsText:        set.TargetRepsText,
				TargetWeightKg:        set.TargetWeightKg,
				TargetDurationSeconds: set.TargetDurationSeconds,
				TargetDistanceKm:      set.TargetDistanceKm,
				TargetRir:             set.TargetRir,
				RestSeconds:           set.RestSeconds,
				Notes:                 set.Notes,
			})
		}

		isMandatory := exercise.ExerciseOrder <= 2
		if isMandatory {
			mandatoryCount++
		}

		exercises = append(exercises, model.AIRoutineExercise{
			ExerciseID:      exercise.ExerciseID,
			Name:            exercise.ExerciseName,
			MuscleGroup:     exercise.MuscleGroup,
			ExerciseType:    exercise.ExerciseType,
			IsMandatory:     isMandatory,
			RecommendedSets: len(sets),
			Sets:            sets,
		})
	}

	return model.AIRoutineJSON{
		Name:             routine.Name,
		Objective:        routine.Description,
		DurationMinutes:  estimateRoutineDurationMinutes(routine),
		TargetMuscles:    extractRoutineTargetMuscles(&routine),
		MandatoryCount:   mandatoryCount,
		GeneratedAt:      now,
		Exercises:        exercises,
		GenerationSource: routine.Source,
	}
}

func routineDetailToRoutine(detail model.RoutineDetail, userID string) model.Routine {
	routine := model.Routine{
		ID:          detail.ID,
		UserID:      userID,
		Name:        detail.Name,
		Description: detail.Description,
		Source:      detail.Source,
		CreatedAt:   detail.CreatedAt,
		UpdatedAt:   detail.UpdatedAt,
		Exercises:   make([]model.RoutineExercise, 0, len(detail.Exercises)),
	}

	for _, exercise := range detail.Exercises {
		legacyExercise := model.RoutineExercise{
			ID:                  exercise.ID,
			RoutineID:           detail.ID,
			ExerciseID:          exercise.ExerciseID,
			ExerciseName:        exercise.Name,
			ExerciseDescription: exercise.Description,
			MuscleGroup:         exercise.MuscleGroup,
			ExerciseType:        exercise.ExerciseType,
			ExerciseOrder:       exercise.ExerciseOrder,
			Notes:               exercise.Notes,
			Sets:                make([]model.RoutineSet, 0, len(exercise.Sets)),
		}
		for _, set := range exercise.Sets {
			legacyExercise.Sets = append(legacyExercise.Sets, model.RoutineSet{
				ID:                    set.ID,
				RoutineExerciseID:     exercise.ID,
				SetNumber:             set.SetNumber,
				TargetRepsMin:         set.TargetRepsMin,
				TargetRepsMax:         set.TargetRepsMax,
				TargetRepsText:        set.TargetRepsText,
				TargetWeightKg:        set.TargetWeightKg,
				TargetDurationSeconds: set.TargetDurationSeconds,
				TargetDistanceKm:      set.TargetDistanceKm,
				TargetRir:             set.TargetRir,
				RestSeconds:           set.RestSeconds,
				Notes:                 set.Notes,
			})
		}
		routine.Exercises = append(routine.Exercises, legacyExercise)
	}

	return routine
}

func extractRoutineTargetMuscles(routine *model.Routine) []string {
	if routine == nil {
		return nil
	}

	muscles := make([]string, 0, len(routine.Exercises))
	for _, exercise := range routine.Exercises {
		muscles = append(muscles, exercise.MuscleGroup)
	}
	return normalizeTextList(muscles)
}

func estimateRoutineDurationMinutes(routine model.Routine) int {
	if len(routine.Exercises) == 0 {
		return 0
	}

	totalSets := 0
	for _, exercise := range routine.Exercises {
		if len(exercise.Sets) > 0 {
			totalSets += len(exercise.Sets)
			continue
		}
		totalSets += 3
	}

	estimated := totalSets * 3
	if estimated < 20 {
		estimated = 20
	}
	return estimated
}

func buildAIRoutineUpgradeDiff(
	before model.AIRoutineJSON,
	after model.AIRoutineJSON,
) model.AIRoutineUpgradeDiff {
	beforeByExerciseID, beforeOrderByExerciseID := indexAIRoutineExercises(before.Exercises)
	afterByExerciseID, afterOrderByExerciseID := indexAIRoutineExercises(after.Exercises)
	orderedIDs := orderedAIRoutineExerciseIDs(before.Exercises, after.Exercises)

	diff := model.AIRoutineUpgradeDiff{
		Exercises: make([]model.AIRoutineExerciseDiffEntry, 0, len(orderedIDs)),
	}

	for _, exerciseID := range orderedIDs {
		beforeExercise, hasBefore := beforeByExerciseID[exerciseID]
		afterExercise, hasAfter := afterByExerciseID[exerciseID]

		switch {
		case !hasBefore && hasAfter:
			diff.Summary.AddedExercises++
			diff.Exercises = append(diff.Exercises, buildAddedExerciseDiff(afterExercise, afterOrderByExerciseID[exerciseID]))
		case hasBefore && !hasAfter:
			diff.Summary.RemovedExercises++
			diff.Exercises = append(diff.Exercises, buildRemovedExerciseDiff(beforeExercise, beforeOrderByExerciseID[exerciseID]))
		default:
			entry := buildModifiedOrUnchangedExerciseDiff(
				beforeExercise,
				afterExercise,
				beforeOrderByExerciseID[exerciseID],
				afterOrderByExerciseID[exerciseID],
			)
			if entry.ChangeType == "modified" {
				diff.Summary.ModifiedExercises++
			} else {
				diff.Summary.UnchangedExercises++
			}
			diff.Exercises = append(diff.Exercises, entry)
		}
	}

	return diff
}

func indexAIRoutineExercises(exercises []model.AIRoutineExercise) (map[string]model.AIRoutineExercise, map[string]int) {
	byExerciseID := make(map[string]model.AIRoutineExercise, len(exercises))
	orderByExerciseID := make(map[string]int, len(exercises))
	for index, exercise := range exercises {
		exerciseID := strings.TrimSpace(exercise.ExerciseID)
		if exerciseID == "" {
			continue
		}
		byExerciseID[exerciseID] = exercise
		orderByExerciseID[exerciseID] = index + 1
	}
	return byExerciseID, orderByExerciseID
}

func orderedAIRoutineExerciseIDs(before, after []model.AIRoutineExercise) []string {
	orderedIDs := make([]string, 0, len(before)+len(after))
	seenIDs := make(map[string]struct{}, len(before)+len(after))
	appendUniqueAIRoutineExerciseIDs(&orderedIDs, seenIDs, before)
	appendUniqueAIRoutineExerciseIDs(&orderedIDs, seenIDs, after)
	return orderedIDs
}

func appendUniqueAIRoutineExerciseIDs(
	orderedIDs *[]string,
	seenIDs map[string]struct{},
	exercises []model.AIRoutineExercise,
) {
	for _, exercise := range exercises {
		exerciseID := strings.TrimSpace(exercise.ExerciseID)
		if exerciseID == "" {
			continue
		}
		if _, exists := seenIDs[exerciseID]; exists {
			continue
		}
		*orderedIDs = append(*orderedIDs, exerciseID)
		seenIDs[exerciseID] = struct{}{}
	}
}

func buildAddedExerciseDiff(exercise model.AIRoutineExercise, order int) model.AIRoutineExerciseDiffEntry {
	orderPtr := routineAIIntPointer(order)
	return model.AIRoutineExerciseDiffEntry{
		ChangeType:        "added",
		ExerciseID:        exercise.ExerciseID,
		AfterName:         exercise.Name,
		AfterOrder:        orderPtr,
		AfterMuscleGroup:  exercise.MuscleGroup,
		AfterExerciseType: exercise.ExerciseType,
		Sets:              buildSetDiffEntries(nil, exercise.Sets),
	}
}

func buildRemovedExerciseDiff(exercise model.AIRoutineExercise, order int) model.AIRoutineExerciseDiffEntry {
	orderPtr := routineAIIntPointer(order)
	return model.AIRoutineExerciseDiffEntry{
		ChangeType:         "removed",
		ExerciseID:         exercise.ExerciseID,
		BeforeName:         exercise.Name,
		BeforeOrder:        orderPtr,
		BeforeMuscleGroup:  exercise.MuscleGroup,
		BeforeExerciseType: exercise.ExerciseType,
		Sets:               buildSetDiffEntries(exercise.Sets, nil),
	}
}

func buildModifiedOrUnchangedExerciseDiff(
	before model.AIRoutineExercise,
	after model.AIRoutineExercise,
	beforeOrder int,
	afterOrder int,
) model.AIRoutineExerciseDiffEntry {
	beforeOrderPtr := routineAIIntPointer(beforeOrder)
	afterOrderPtr := routineAIIntPointer(afterOrder)
	setDiffs := buildSetDiffEntries(before.Sets, after.Sets)

	entry := model.AIRoutineExerciseDiffEntry{
		ChangeType:         "unchanged",
		ExerciseID:         after.ExerciseID,
		BeforeName:         before.Name,
		AfterName:          after.Name,
		BeforeOrder:        beforeOrderPtr,
		AfterOrder:         afterOrderPtr,
		BeforeMuscleGroup:  before.MuscleGroup,
		AfterMuscleGroup:   after.MuscleGroup,
		BeforeExerciseType: before.ExerciseType,
		AfterExerciseType:  after.ExerciseType,
		IsMandatoryChanged: before.IsMandatory != after.IsMandatory,
		Sets:               setDiffs,
	}

	if before.Name != after.Name ||
		before.MuscleGroup != after.MuscleGroup ||
		before.ExerciseType != after.ExerciseType ||
		before.IsMandatory != after.IsMandatory ||
		beforeOrder != afterOrder ||
		hasChangedSetDiff(setDiffs) {
		entry.ChangeType = "modified"
	}

	return entry
}

func buildSetDiffEntries(
	beforeSets []model.AIRoutineExerciseSet,
	afterSets []model.AIRoutineExerciseSet,
) []model.AIRoutineExerciseSetDiffEntry {
	maxLen := len(beforeSets)
	if len(afterSets) > maxLen {
		maxLen = len(afterSets)
	}

	diffs := make([]model.AIRoutineExerciseSetDiffEntry, 0, maxLen)
	for index := 0; index < maxLen; index++ {
		beforeSet, afterSet := routineAISetPointersAt(beforeSets, afterSets, index)
		setNumber := resolveRoutineAISetNumber(index, beforeSet, afterSet)
		changeType := classifyRoutineAISetChange(beforeSet, afterSet)

		diffs = append(diffs, model.AIRoutineExerciseSetDiffEntry{
			ChangeType: changeType,
			SetNumber:  setNumber,
			Before:     beforeSet,
			After:      afterSet,
		})
	}

	return diffs
}

func routineAISetPointersAt(
	beforeSets []model.AIRoutineExerciseSet,
	afterSets []model.AIRoutineExerciseSet,
	index int,
) (*model.AIRoutineExerciseSet, *model.AIRoutineExerciseSet) {
	var beforeSet *model.AIRoutineExerciseSet
	var afterSet *model.AIRoutineExerciseSet

	if index < len(beforeSets) {
		beforeCopy := beforeSets[index]
		beforeSet = &beforeCopy
	}
	if index < len(afterSets) {
		afterCopy := afterSets[index]
		afterSet = &afterCopy
	}

	return beforeSet, afterSet
}

func resolveRoutineAISetNumber(index int, beforeSet, afterSet *model.AIRoutineExerciseSet) int {
	if afterSet != nil && afterSet.SetNumber > 0 {
		return afterSet.SetNumber
	}
	if beforeSet != nil && beforeSet.SetNumber > 0 {
		return beforeSet.SetNumber
	}
	return index + 1
}

func classifyRoutineAISetChange(beforeSet, afterSet *model.AIRoutineExerciseSet) string {
	switch {
	case beforeSet == nil && afterSet != nil:
		return "added"
	case beforeSet != nil && afterSet == nil:
		return "removed"
	case beforeSet != nil && afterSet != nil && !equalAIRoutineSets(*beforeSet, *afterSet):
		return "modified"
	default:
		return "unchanged"
	}
}

func equalAIRoutineSets(a, b model.AIRoutineExerciseSet) bool {
	return a.SetNumber == b.SetNumber &&
		equalOptionalInt(a.TargetRepsMin, b.TargetRepsMin) &&
		equalOptionalInt(a.TargetRepsMax, b.TargetRepsMax) &&
		a.TargetRepsText == b.TargetRepsText &&
		equalOptionalFloat(a.TargetWeightKg, b.TargetWeightKg) &&
		equalOptionalInt(a.TargetDurationSeconds, b.TargetDurationSeconds) &&
		equalOptionalFloat(a.TargetDistanceKm, b.TargetDistanceKm) &&
		equalOptionalInt(a.TargetRir, b.TargetRir) &&
		equalOptionalInt(a.RestSeconds, b.RestSeconds) &&
		a.Notes == b.Notes
}

func hasChangedSetDiff(entries []model.AIRoutineExerciseSetDiffEntry) bool {
	for _, entry := range entries {
		if entry.ChangeType != "unchanged" {
			return true
		}
	}
	return false
}

func equalOptionalInt(a, b *int) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}

func equalOptionalFloat(a, b *float64) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}

func routineAIIntPointer(value int) *int {
	return &value
}
