package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type routineAIProvider interface {
	Generate(ctx context.Context, request routineAIProviderRequest) (string, error)
}

type routineAIProviderRequest struct {
	SystemInstruction string
	InputPayload      map[string]any
}

type geminiRoutineAIProvider struct {
	model            string
	apiKey           string
	httpClientGetter func() *http.Client
}

func newGeminiRoutineAIProvider(model, apiKey string, httpClientGetter func() *http.Client) routineAIProvider {
	model = strings.TrimSpace(model)
	if model == "" {
		model = "gemini-2.5-flash"
	}

	return &geminiRoutineAIProvider{
		model:            model,
		apiKey:           strings.TrimSpace(apiKey),
		httpClientGetter: httpClientGetter,
	}
}

func (p *geminiRoutineAIProvider) Generate(ctx context.Context, request routineAIProviderRequest) (string, error) {
	if p.apiKey == "" {
		return "", ErrAIRoutineMissingAPIKey
	}

	requestBody, err := buildGeminiRequestBody(request.SystemInstruction, request.InputPayload)
	if err != nil {
		return "", err
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
	}

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		url.PathEscape(p.model),
		url.QueryEscape(p.apiKey),
	)

	requestStartedAt := time.Now()
	slog.Info("ai routine gemini request started", "model", p.model)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := p.httpClient()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
	}
	defer httpResp.Body.Close()

	responseBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
	}
	slog.Info("ai routine gemini response received",
		"model", p.model,
		"status_code", httpResp.StatusCode,
		"duration_ms", time.Since(requestStartedAt).Milliseconds(),
		"response_bytes", len(responseBody),
	)

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		slog.Error("ai routine gemini response status error",
			"model", p.model,
			"status_code", httpResp.StatusCode,
			"response_snippet", truncateProviderError(responseBody),
		)
		return "", fmt.Errorf(
			"%w: gemini status %d: %s",
			ErrAIRoutineProviderUnavailable,
			httpResp.StatusCode,
			truncateProviderError(responseBody),
		)
	}

	var geminiResp geminiGenerateContentResponse
	if err := json.Unmarshal(responseBody, &geminiResp); err != nil {
		return "", fmt.Errorf("%w: %w", ErrAIRoutineProviderUnavailable, err)
	}

	jsonText := extractGeminiText(geminiResp)
	if strings.TrimSpace(jsonText) == "" {
		slog.Error("ai routine gemini response missing text", "model", p.model)
		return "", ErrAIRoutineProviderUnavailable
	}

	if prettyJSON := prettyJSONForLog(jsonText); prettyJSON != "" {
		slog.Info("ai routine gemini raw response",
			"response", truncateLogValue(prettyJSON, 8000),
		)
	}

	if looksLikePromptEcho(jsonText) {
		slog.Error("ai routine gemini response echoed input payload")
		return "", ErrAIRoutineProviderUnavailable
	}

	return jsonText, nil
}

func (p *geminiRoutineAIProvider) httpClient() *http.Client {
	if p.httpClientGetter != nil {
		if client := p.httpClientGetter(); client != nil {
			return client
		}
	}
	return &http.Client{Timeout: 25 * time.Second}
}

func buildGeminiRequestBody(systemInstruction string, inputPayload map[string]any) (map[string]any, error) {
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
