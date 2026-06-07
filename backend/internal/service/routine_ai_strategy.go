package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
)

type routineGenerationStrategy interface {
	Generate(ctx context.Context, userID string, req model.AIRoutineGenerationRequest) (model.AIRoutineGenerateResponse, error)
}

type routineUpgradeStrategy interface {
	Upgrade(ctx context.Context, userID, routineID string, req model.AIRoutineUpgradeRequest) (model.AIRoutineUpgradeResponse, error)
}

type geminiRoutineGenerationStrategy struct {
	service  *RoutineAIService
	provider routineAIProvider
}

func (s *geminiRoutineGenerationStrategy) Generate(
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
	slog.Info("ai routine generation started",
		"user_id", userID,
		"objective", req.Objective,
		"duration_minutes", req.DurationMinutes,
		"target_muscle_groups", normalizeTextList(req.TargetMuscleGroups),
		"mandatory_exercises_count", len(normalizeTextList(req.MandatoryExercises)),
		"notes_present", strings.TrimSpace(req.Notes) != "",
	)

	exercises, err := s.service.repo.ListAvailableExercisesForAI(ctx, userID, req.TargetMuscleGroups, 200)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	userContext, err := s.service.buildUserContext(ctx, userID, now)
	if err != nil {
		slog.Error("ai routine user context build failed", "user_id", userID, "error", err)
		return model.AIRoutineGenerateResponse{}, err
	}

	rawJSON, err := s.provider.Generate(ctx, routineAIProviderRequest{
		SystemInstruction: "You are a workout planner. Use user_context, especially recent_training_history, as the main history signal. Respect user_notes and mandatory_exercises as strong instructions from the user. Build the most complete and sensible routine possible for the available time, choosing the exercise count freely based on the objective, time available, and user requests. Do not force a one-to-one mapping between target muscle groups and exercises, and do not use a fixed exercise count. Prefer the best coverage and exercise selection for the routine as a whole. If mandatory_exercises is empty, do not split exercises into optional vs mandatory; treat every exercise in the routine as required. If mandatory_exercises is not empty, mark the requested exercises as mandatory and keep the rest as non-mandatory. Return only valid JSON matching output_contract. Put planned sets in exercises[].sets. Use target_weight_kg only when recent history supports it; otherwise use null or omit it. Do not include markdown.",
		InputPayload:      buildRoutineGenerationInputPayload(req, exercises, userContext),
	})
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	generated, err := finalizeGeneratedAIRoutine(rawJSON, req, now)
	if err != nil {
		slog.Error("ai routine generation failed", "user_id", userID, "error", err)
		return model.AIRoutineGenerateResponse{}, err
	}

	if err := s.service.repo.LogAIGeneration(ctx, userID, now); err != nil {
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
	}, nil
}

type geminiRoutineUpgradeStrategy struct {
	service  *RoutineAIService
	provider routineAIProvider
}

func (s *geminiRoutineUpgradeStrategy) Upgrade(
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

	routine, err := s.service.repo.GetByID(ctx, userID, routineID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AIRoutineUpgradeResponse{}, ErrRoutineNotFound
		}
		return model.AIRoutineUpgradeResponse{}, err
	}
	legacyRoutine := routineDetailToRoutine(*routine, userID)

	exercises, err := s.service.repo.ListAvailableExercisesForAI(ctx, userID, nil, 200)
	if err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}

	userContext, err := s.service.buildUpgradeUserContext(ctx, userID, legacyRoutine, now)
	if err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}

	rawJSON, err := s.provider.Generate(ctx, routineAIProviderRequest{
		SystemInstruction: "You are a workout planner improving an existing routine. Use user_context, especially recent_training_history, as the main history signal. Treat existing_routine as the base plan to refine, not to discard without a good reason. Respect message and feedback_message as strong user instructions. Keep the response fully compatible with the existing routine format. Preserve exercises that are already working unless user guidance or training context suggests a meaningful change. Prefer targeted improvements in exercise selection, order, set structure, fatigue management, and muscle balance over random rewrites. Prefer the provided exercise_catalog and avoid inventing unsupported exercises. Return only valid JSON matching output_contract. Return only the upgraded routine object itself. Never return or repeat the input payload. Never include keys such as exercise_catalog, existing_routine, user_context, message, feedback_message, or output_contract in the response. Put planned sets in exercises[].sets. Use target_weight_kg only when recent history supports it; otherwise use null or omit it. Do not include markdown.",
		InputPayload:      buildRoutineUpgradeInputPayload(legacyRoutine, req, exercises, userContext, now),
	})
	if err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}

	generated, err := finalizeGeneratedAIRoutineUpgrade(rawJSON, legacyRoutine, now)
	if err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}

	if _, err := s.service.buildExercisesToSave(generated, exercises); err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}

	originalRoutineJSON := buildAIRoutineJSONFromRoutine(legacyRoutine, now)
	diff := buildAIRoutineUpgradeDiff(originalRoutineJSON, generated)

	if err := s.service.repo.LogAIGeneration(ctx, userID, now); err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}

	return model.AIRoutineUpgradeResponse{
		RoutineJSON: generated,
		Diff:        diff,
	}, nil
}

type rateLimitedRoutineGenerationStrategy struct {
	inner   routineGenerationStrategy
	limiter routineAIRateLimiter
}

func (s *rateLimitedRoutineGenerationStrategy) Generate(
	ctx context.Context,
	userID string,
	req model.AIRoutineGenerationRequest,
) (model.AIRoutineGenerateResponse, error) {
	now := time.Now().UTC()
	status, err := s.limiter.check(ctx, userID, now)
	if err != nil {
		if errors.Is(err, ErrAIRoutineRateLimited) {
			return model.AIRoutineGenerateResponse{RateLimit: &status}, err
		}
		return model.AIRoutineGenerateResponse{}, err
	}

	response, err := s.inner.Generate(ctx, userID, req)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}
	response.RateLimit = rateLimitStatusPointer(buildAIRoutineRateLimitStatus(status.UsedInCurrentWindow+1, status.ResetAt))
	return response, nil
}

type rateLimitedRoutineUpgradeStrategy struct {
	inner   routineUpgradeStrategy
	limiter routineAIRateLimiter
}

func (s *rateLimitedRoutineUpgradeStrategy) Upgrade(
	ctx context.Context,
	userID, routineID string,
	req model.AIRoutineUpgradeRequest,
) (model.AIRoutineUpgradeResponse, error) {
	now := time.Now().UTC()
	status, err := s.limiter.check(ctx, userID, now)
	if err != nil {
		if errors.Is(err, ErrAIRoutineRateLimited) {
			return model.AIRoutineUpgradeResponse{RateLimit: status}, err
		}
		return model.AIRoutineUpgradeResponse{}, err
	}

	response, err := s.inner.Upgrade(ctx, userID, routineID, req)
	if err != nil {
		return model.AIRoutineUpgradeResponse{}, err
	}
	response.RateLimit = buildAIRoutineRateLimitStatus(status.UsedInCurrentWindow+1, status.ResetAt)
	return response, nil
}

type routineAIRateLimiter interface {
	check(ctx context.Context, userID string, now time.Time) (model.AIRoutineRateLimitStatus, error)
}

type repositoryRoutineAIRateLimiter struct {
	repo repositoryRoutineAICountRepository
}

type repositoryRoutineAICountRepository interface {
	CountAIGenerationsInWindow(ctx context.Context, userID string, since time.Time) (int, error)
}

func (l repositoryRoutineAIRateLimiter) check(ctx context.Context, userID string, now time.Time) (model.AIRoutineRateLimitStatus, error) {
	since := now.Add(-aiRoutineRateWindow)
	used, err := l.repo.CountAIGenerationsInWindow(ctx, userID, since)
	if err != nil {
		return model.AIRoutineRateLimitStatus{}, err
	}

	status := buildAIRoutineRateLimitStatus(used, now.Add(aiRoutineRateWindow))
	if used >= aiRoutineRateLimit {
		return status, ErrAIRoutineRateLimited
	}

	return status, nil
}

func buildRoutineGenerationInputPayload(
	req model.AIRoutineGenerationRequest,
	exercises []model.Exercise,
	userContext routineAIUserContext,
) map[string]any {
	return map[string]any{
		"objective":            req.Objective,
		"duration_minutes":     req.DurationMinutes,
		"target_muscle_groups": normalizeTextList(req.TargetMuscleGroups),
		"mandatory_exercises":  normalizeTextList(req.MandatoryExercises),
		"user_notes":           strings.TrimSpace(req.Notes),
		"user_context":         userContext,
		"exercise_catalog":     buildAIExerciseCatalog(exercises),
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
}

func buildRoutineUpgradeInputPayload(
	routine model.Routine,
	req model.AIRoutineUpgradeRequest,
	exercises []model.Exercise,
	userContext routineAIUserContext,
	now time.Time,
) map[string]any {
	return map[string]any{
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
}

func finalizeGeneratedAIRoutineUpgrade(jsonText string, routine model.Routine, now time.Time) (model.AIRoutineJSON, error) {
	var generated model.AIRoutineJSON
	if err := json.Unmarshal([]byte(jsonText), &generated); err != nil {
		return model.AIRoutineJSON{}, fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
	}

	builder := newRoutineJSONBuilder(generated)
	builder.withUpgradeDefaults(routine, now)
	return builder.build(), nil
}
