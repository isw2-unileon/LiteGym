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
	repo      repository.RoutineRepository
	apiKey    string
	model     string
	httpClient *http.Client
}

func NewRoutineAIService(repo repository.RoutineRepository, apiKey, model string) *RoutineAIService {
	model = strings.TrimSpace(model)
	if model == "" {
		model = "gemini-1.5-flash"
	}

	return &RoutineAIService{
		repo:      repo,
		apiKey:    strings.TrimSpace(apiKey),
		model:     model,
		httpClient: &http.Client{Timeout: 25 * time.Second},
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
	remaining := aiRoutineRateLimit - used
	if remaining < 0 {
		remaining = 0
	}

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

	generated, err := s.generateWithGemini(ctx, req, exercises, now)
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
	now time.Time,
) (model.AIRoutineJSON, error) {
	if s.apiKey == "" {
		return model.AIRoutineJSON{}, ErrAIRoutineMissingAPIKey
	}

	exerciseCatalog := make([]map[string]string, 0, len(exercises))
	for _, exercise := range exercises {
		exerciseCatalog = append(exerciseCatalog, map[string]string{
			"id":           exercise.ID,
			"name":         exercise.Name,
			"muscle_group": exercise.MuscleGroup,
			"exercise_type": exercise.ExerciseType,
		})
	}

	inputPayload := map[string]any{
		"objective": req.Objective,
		"duration_minutes": req.DurationMinutes,
		"target_muscle_groups": normalizeTextList(req.TargetMuscleGroups),
		"mandatory_exercise_ids": normalizeTextList(req.MandatoryExerciseIDs),
		"exercise_catalog": exerciseCatalog,
		"output_contract": map[string]any{
			"name": "string",
			"objective": "string",
			"duration_minutes": "number",
			"target_muscles": []string{},
			"mandatory_count": "number",
			"generated_at": "RFC3339 datetime string",
			"generation_source": "string",
			"exercises": []map[string]string{
				{
					"exercise_id": "string",
					"name": "string",
					"muscle_group": "string",
					"exercise_type": "string",
					"is_mandatory": "boolean",
					"recommended_sets": "number",
					"recommended_reps": "string",
				},
			},
		},
	}

	systemInstruction := "You are a workout planner. Return only valid JSON matching output_contract. Do not include markdown."
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
