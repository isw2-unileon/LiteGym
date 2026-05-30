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

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
)

// ProfileService provides business logic for user profiles.
type ProfileService struct {
	repo         repository.ProfileRepository
	geminiAPIKey string
	geminiModel  string
	httpClient   *http.Client
}

// NewProfileService creates a new ProfileService.
func NewProfileService(repo repository.ProfileRepository, aiConfig ...string) *ProfileService {
	var apiKey, modelName string
	if len(aiConfig) > 0 {
		apiKey = strings.TrimSpace(aiConfig[0])
	}
	if len(aiConfig) > 1 {
		modelName = strings.TrimSpace(aiConfig[1])
	}
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	return &ProfileService{
		repo:         repo,
		geminiAPIKey: apiKey,
		geminiModel:  modelName,
		httpClient:   &http.Client{Timeout: 20 * time.Second},
	}
}

// GetDashboardStats retrieves all stats needed for the profile view.
func (s *ProfileService) GetDashboardStats(ctx context.Context, userID string, timeRange string, year int, month int) (*model.ProfileStats, error) {
	return s.repo.GetStats(ctx, userID, timeRange, year, month)
}

// UpdateGoals saves the user's short and long term goals.
func (s *ProfileService) UpdateGoals(ctx context.Context, goal *model.UserGoal) error {
	return s.repo.UpsertGoals(ctx, goal)
}

// AddBodyMetric adds a new body metric record.
func (s *ProfileService) AddBodyMetric(ctx context.Context, userID string, req *model.AddBodyMetricRequest) error {
	return s.repo.AddBodyMetric(ctx, userID, req)
}

// GenerateAIAnalysis queries Gemini to generate a personalized performance coaching feedback in Spanish.
// If the API key is not configured or queries fail, it gracefully falls back to a high-quality offline analysis.
func (s *ProfileService) GenerateAIAnalysis(ctx context.Context, stats *model.ProfileStats) (string, error) {
	if stats == nil {
		return "Aún no tengo suficientes datos de tu perfil para generar un análisis. ¡Empieza a registrar tus entrenamientos y pesajes!", nil
	}

	// 1. Check if Gemini API is available. If not, fallback immediately
	if s.geminiAPIKey == "" {
		return s.generateOfflineFeedback(stats), nil
	}

	// 2. Format the user information for the prompt
	shortTermGoal := "progreso general"
	longTermGoal := "desarrollo constante"
	targetDays := 3
	if stats.Goals != nil {
		if strings.TrimSpace(stats.Goals.ShortTerm) != "" {
			shortTermGoal = stats.Goals.ShortTerm
		}
		if strings.TrimSpace(stats.Goals.LongTerm) != "" {
			longTermGoal = stats.Goals.LongTerm
		}
		if stats.Goals.TargetDaysPerWeek > 0 {
			targetDays = stats.Goals.TargetDaysPerWeek
		}
	}

	var muscleList []string
	for _, m := range stats.MuscleRadar {
		muscleList = append(muscleList, fmt.Sprintf("%s (%d series)", m.Muscle, m.Value))
	}
	muscleDistribution := strings.Join(muscleList, ", ")
	if muscleDistribution == "" {
		muscleDistribution = "Ninguna serie registrada aún"
	}

	var exerciseList []string
	for _, e := range stats.TopExercises {
		exerciseList = append(exerciseList, fmt.Sprintf("%s (%d series)", e.Name, e.Sets))
	}
	topExercises := strings.Join(exerciseList, ", ")
	if topExercises == "" {
		topExercises = "Ninguno registrado"
	}

	var weightList []string
	for _, w := range stats.WeightHistory {
		weightList = append(weightList, fmt.Sprintf("%.1fkg (%s)", w.WeightKg, w.RecordedAt.Format("02/01")))
	}
	weightHistory := strings.Join(weightList, " -> ")
	if weightHistory == "" {
		weightHistory = "Sin registro de peso aún"
	}

	promptText := fmt.Sprintf(`Datos reales de progreso del usuario:
- Objetivos: Corto plazo: "%s". Largo plazo: "%s". Días de entrenamiento semanales objetivo: %d días.
- Actividad mensual: %d entrenamientos completados, %d minutos totales entrenados, %d series totales, %.1f kg de volumen total acumulado.
- Reparto muscular por series: %s.
- Ejercicios preferidos: %s.
- Historial de peso corporal reciente: %s.

Por favor, como mi entrenador personal, analízalos y dame tu feedback personalizado para mi meta de "%s".`,
		shortTermGoal, longTermGoal, targetDays,
		stats.TotalWorkouts, stats.TotalDuration, stats.TotalSets, stats.TotalVolume,
		muscleDistribution, topExercises, weightHistory, shortTermGoal)

	systemInstruction := "Eres un entrenador personal de gimnasio e inteligencia artificial de LiteGym experto y motivador. Analiza los datos de rendimiento, métricas corporales e historial de entrenamiento del usuario y genera un análisis de rendimiento y consejos altamente personalizados, prácticos y motivadores en español. Mantén el tono profesional, cercano, enérgico y optimista. El análisis debe ser breve (máximo 3-4 frases o unos 300 caracteres) y estar en formato de texto plano sin usar ningún tipo de formato markdown (sin asteriscos, sin almohadillas, sin negrita, sin viñetas)."

	requestBody := map[string]any{
		"system_instruction": map[string]any{
			"parts": []map[string]string{
				{"text": systemInstruction},
			},
		},
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": promptText},
				},
			},
		},
		"generationConfig": map[string]any{
			"temperature": 0.7,
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		slog.Error("failed to marshal profile ai request", "error", err)
		return s.generateOfflineFeedback(stats), nil
	}

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		url.PathEscape(s.geminiModel),
		url.QueryEscape(s.geminiAPIKey),
	)

	requestStartedAt := time.Now()
	slog.Info("profile ai gemini request started", "model", s.geminiModel)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		slog.Error("failed to create profile ai request", "error", err)
		return s.generateOfflineFeedback(stats), nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		slog.Warn("gemini api call failed for profile coach, using offline fallback", "error", err)
		return s.generateOfflineFeedback(stats), nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read gemini api response for profile coach", "error", err)
		return s.generateOfflineFeedback(stats), nil
	}

	slog.Info("profile ai gemini response received",
		"model", s.geminiModel,
		"status_code", resp.StatusCode,
		"duration_ms", time.Since(requestStartedAt).Milliseconds(),
		"response_bytes", len(respBody),
	)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Warn("gemini api returned status error for profile coach, using offline fallback",
			"status_code", resp.StatusCode,
			"response_snippet", truncateProviderError(respBody),
		)
		return s.generateOfflineFeedback(stats), nil
	}

	var geminiResp geminiGenerateContentResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		slog.Error("failed to unmarshal gemini api response for profile coach", "error", err)
		return s.generateOfflineFeedback(stats), nil
	}

	analysis := extractGeminiText(geminiResp)
	if strings.TrimSpace(analysis) == "" {
		slog.Error("profile ai gemini response missing text", "model", s.geminiModel)
		return s.generateOfflineFeedback(stats), nil
	}

	slog.Info("profile ai gemini raw response",
		"response", truncateLogValue(analysis, 500),
	)

	return strings.TrimSpace(analysis), nil
}

// generateOfflineFeedback generates high-quality responsive fitness analysis when AI is offline.
func (s *ProfileService) generateOfflineFeedback(stats *model.ProfileStats) string {
	shortTerm := "mejorar tu condición física"
	if stats.Goals != nil && strings.TrimSpace(stats.Goals.ShortTerm) != "" {
		shortTerm = stats.Goals.ShortTerm
	}

	topMuscle := "Pecho"
	maxVal := -1
	for _, m := range stats.MuscleRadar {
		if m.Value > maxVal {
			maxVal = m.Value
			topMuscle = m.Muscle
		}
	}

	if stats.TotalWorkouts == 0 {
		return fmt.Sprintf("¡Bienvenido a LiteGym! Veo que tu meta a corto plazo es \"%s\". Para empezar con fuerza, te recomiendo planificar tu primer entrenamiento en el Dashboard o generar una rutina automática con IA.", shortTerm)
	}

	return fmt.Sprintf("¡Vas genial con tu entrenamiento! He analizado tu historial y veo que dominas el trabajo de %s, con un total de %d series y %d entrenamientos completados. Para lograr tu meta de \"%s\", céntrate en mantener esta consistencia y registrar un nuevo pesaje esta semana.", topMuscle, stats.TotalSets, stats.TotalWorkouts, shortTerm)
}
