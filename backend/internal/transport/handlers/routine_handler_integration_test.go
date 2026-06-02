//go:build integration

package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/testutil"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func setupRoutineAIIntegrationDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return testutil.NewIntegrationTestPool(t)
}

func loadGeminiAPIKeyFromBackendEnv(t *testing.T) {
	t.Helper()

	if apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY")); apiKey != "" {
		return
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file location")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", ".."))
	envPath := filepath.Join(repoRoot, "backend", ".env")

	values, err := godotenv.Read(envPath)
	if err != nil {
		t.Skip("skipping real Gemini integration test: GEMINI_API_KEY is not set and backend/.env is unavailable")
	}

	apiKey := strings.TrimSpace(values["GEMINI_API_KEY"])
	if apiKey == "" {
		t.Skip("skipping real Gemini integration test: GEMINI_API_KEY is not set in backend/.env")
	}

	for key, value := range values {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("failed to set %s for test: %v", key, err)
		}
	}
}

func cleanupRoutineAIIntegration(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	tables := []string{
		"public.workout_sets",
		"public.workout_exercises",
		"public.workout_sessions",
		"public.support_tickets",
		"public.shared_routines",
		"public.routine_exercises",
		"public.routines",
		"public.friendships",
		"public.exercise_secondary_muscle_groups",
		"public.exercises",
		"public.body_metrics",
		"public.user_profiles",
		"public.users",
	}

	for _, table := range tables {
		if _, err := db.Exec(context.Background(), "DELETE FROM "+table); err != nil {
			t.Fatalf("error limpiando %s: %v", table, err)
		}
	}
}

type routineAIGenerateIntegrationPayload struct {
	RoutineJSON model.AIRoutineJSON `json:"routine_json"`
}

type routineAISaveIntegrationPayload struct {
	RoutineID   string              `json:"routine_id"`
	RoutineJSON model.AIRoutineJSON `json:"routine_json"`
}

func insertUserRawRoutineAI(t *testing.T, db *pgxpool.Pool, username, email string) string {
	t.Helper()

	var id string
	err := db.QueryRow(context.Background(), `
		INSERT INTO public.users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id::text
	`, username, email, "").Scan(&id)
	if err != nil {
		t.Fatalf("error insertando usuario: %v", err)
	}

	return id
}

func setupRoutineAIRouter(db *pgxpool.Pool, aiService *service.RoutineAIService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, "auth_token")

	routineRepo := repository.NewRoutineRepository(db)
	routineService := service.NewRoutineService(routineRepo)
	routineHandler := NewRoutineHandler(routineService, aiService)

	userHandler := NewUserHandler(userService)
	authHandler := NewAuthHandler(userService, tokenService, "auth_token", false)
	exerciseHandler := NewExerciseHandler(service.NewExerciseService(repository.NewExerciseRepository(db)))
	overviewHandler := NewOverviewHandler(service.NewOverviewService(
		repository.NewRoutineRepository(db),
		repository.NewOverviewWorkoutRepository(db),
		repository.NewBodyMetricRepository(db),
	))
	ticketHandler := NewTicketHandler(service.NewTicketService(repository.NewTicketRepository(db)), userService)
	workoutHandler := NewWorkoutHandler(service.NewWorkoutService(repository.NewWorkoutRepository(db)))

	return func() *gin.Engine {
		router := gin.New()
		router.Use(gin.Logger(), gin.Recovery())

		api := router.Group("/api")
		api.POST("/users", userHandler.CreateUser)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/logout", authHandler.Logout)

		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		protected.GET("/hello", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Hello from the API"}) })
		protected.GET("/db/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "database"}) })
		protected.GET("/auth/me", authHandler.Me)
		protected.GET("/users", userHandler.ListAllUsers)
		protected.GET("/users/me", userHandler.GetMe)
		protected.GET("/users/:id", userHandler.GetUserByID)
		protected.DELETE("/users/:id", userHandler.DeleteUser)
		protected.POST("/exercises", exerciseHandler.CreateExercise)
		protected.GET("/exercises/metadata", exerciseHandler.GetMetadata)
		protected.GET("/exercises/:id/insights", exerciseHandler.GetExerciseInsights)
		protected.GET("/exercises/:id/workout-sessions", exerciseHandler.ListWorkoutSessionsByExercise)
		protected.GET("/exercises/:id", exerciseHandler.GetExerciseByID)
		protected.GET("/exercises", exerciseHandler.ListExercises)
		protected.PUT("/exercises/:id", exerciseHandler.UpdateExercise)
		protected.DELETE("/exercises/:id", exerciseHandler.DeleteExercise)
		protected.GET("/routines", routineHandler.ListRoutines)
		protected.GET("/routines/:id", routineHandler.GetRoutine)
		protected.POST("/routines/ai/generate", routineHandler.GenerateRoutineJSON)
		protected.POST("/routines/ai/save", routineHandler.SaveAIRoutine)
		protected.GET("/dashboard", overviewHandler.GetOverview)
		protected.POST("/tickets", ticketHandler.CreateTicket)
		protected.GET("/tickets", ticketHandler.ListTickets)
		protected.PATCH("/tickets/:id/close", ticketHandler.CloseTicket)
		protected.POST("/workouts/planned", workoutHandler.CreatePlannedWorkout)
		protected.POST("/workout/start", workoutHandler.CreateWorkout)
		protected.GET("/workout/:id", workoutHandler.GetWorkoutByID)
		protected.POST("/workout/:id/finish", workoutHandler.FinishWorkout)
		protected.DELETE("/workout/:id", workoutHandler.RemoveWorkout)
		protected.POST("/workout/:id/exercise", workoutHandler.CreateWorkoutExercise)
		protected.GET("/workout/:id/exercises", workoutHandler.GetExercisesByWorkoutID)
		protected.POST("/workout/:id/exercises/:exercise_id/set", workoutHandler.CreateWorkoutSet)
		protected.GET("/workout/:id/exercises/:exercise_id/sets", workoutHandler.GetWorkoutSetsByExerciseID)
		protected.POST("/workout/:id/exercises/:exercise_id/sets/:set_id", workoutHandler.UpdateWorkoutSet)

		return router
	}()
}

func TestRoutineAIPreviewAndSaveIntegration(t *testing.T) {
	runRoutineAIPreviewAndSaveIntegration(t)
}

func runRoutineAIPreviewAndSaveIntegration(t *testing.T) {
	db := setupRoutineAIIntegrationDB(t)
	cleanupRoutineAIIntegration(t, db)

	loadGeminiAPIKeyFromBackendEnv(t)

	userID := insertUserRawRoutineAI(t, db, "aiuser", "aiuser@example.com")

	routineRepo := repository.NewRoutineRepository(db)
	exerciseRepo := repository.NewExerciseRepository(db)
	exerciseService := service.NewExerciseService(exerciseRepo)
	workoutSessionRepo := repository.NewWorkoutSessionRepository(db)
	bodyMetricRepo := repository.NewBodyMetricRepository(db)
	aiService := service.NewRoutineAIService(
		routineRepo,
		exerciseService,
		workoutSessionRepo,
		bodyMetricRepo,
		os.Getenv("GEMINI_API_KEY"),
		strings.TrimSpace(os.Getenv("GEMINI_MODEL")),
	)

	router := setupRoutineAIRouter(db, aiService)

	tokenService := service.NewTokenService("test-secret", "test-issuer", time.Hour)
	token, err := tokenService.GenerateToken(userID, "aiuser@example.com", "aiuser", "user")
	if err != nil {
		t.Fatalf("failed to generate auth token: %v", err)
	}

	generatePayload := mustGenerateRoutinePreview(t, router, token)
	for index := range generatePayload.RoutineJSON.Exercises {
		generatePayload.RoutineJSON.Exercises[index].MuscleGroup = "chest"
	}

	savePayload := mustSaveRoutinePreview(t, router, token, generatePayload)
	mustVerifySavedRoutineExercise(t, db, userID, generatePayload.RoutineJSON.Exercises[0].Name)
	mustVerifyRoutineExerciseCount(t, db, savePayload.RoutineID)
}

func mustGenerateRoutinePreview(
	t *testing.T,
	router *gin.Engine,
	token string,
) routineAIGenerateIntegrationPayload {
	t.Helper()

	generateBody := `{"objective":"Ganar fuerza","duration_minutes":50,"target_muscle_groups":["chest"],"mandatory_exercises":["Press banca"],"notes":"Prioriza técnica y control"}`
	generateReq := httptest.NewRequest(http.MethodPost, "/api/routines/ai/generate", strings.NewReader(generateBody))
	generateReq.Header.Set("Content-Type", "application/json")
	generateReq.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	generateW := httptest.NewRecorder()
	router.ServeHTTP(generateW, generateReq)

	if generateW.Code == http.StatusServiceUnavailable &&
		(strings.Contains(generateW.Body.String(), "RESOURCE_EXHAUSTED") ||
			strings.Contains(generateW.Body.String(), "quota exceeded") ||
			strings.Contains(generateW.Body.String(), "gemini status 429")) {
		t.Skip("skipping real Gemini integration test: quota exceeded")
	}

	if generateW.Code != http.StatusOK {
		t.Fatalf("expected generate status %d, got %d body=%s", http.StatusOK, generateW.Code, strings.TrimSpace(generateW.Body.String()))
	}

	var payload routineAIGenerateIntegrationPayload
	if err := json.Unmarshal(generateW.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal generate response: %v", err)
	}
	if len(payload.RoutineJSON.Exercises) == 0 {
		t.Fatalf("expected at least one generated exercise, got %#v", payload.RoutineJSON.Exercises)
	}
	if strings.TrimSpace(payload.RoutineJSON.Exercises[0].Name) == "" {
		t.Fatal("expected the generated exercise to have a name")
	}

	return payload
}

func mustSaveRoutinePreview(
	t *testing.T,
	router *gin.Engine,
	token string,
	generatePayload routineAIGenerateIntegrationPayload,
) routineAISaveIntegrationPayload {
	t.Helper()

	saveBodyBytes, err := json.Marshal(generatePayload)
	if err != nil {
		t.Fatalf("failed to marshal save payload: %v", err)
	}
	saveReq := httptest.NewRequest(http.MethodPost, "/api/routines/ai/save", bytes.NewReader(saveBodyBytes))
	saveReq.Header.Set("Content-Type", "application/json")
	saveReq.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	saveW := httptest.NewRecorder()
	router.ServeHTTP(saveW, saveReq)

	if saveW.Code != http.StatusOK {
		t.Fatalf("expected save status %d, got %d body=%s", http.StatusOK, saveW.Code, strings.TrimSpace(saveW.Body.String()))
	}

	var payload routineAISaveIntegrationPayload
	if err := json.Unmarshal(saveW.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal save response: %v", err)
	}
	if payload.RoutineID == "" {
		t.Fatal("expected saved routine id in response")
	}

	return payload
}

func mustVerifySavedRoutineExercise(t *testing.T, db *pgxpool.Pool, userID, exerciseName string) {
	t.Helper()

	var exerciseOwner sql.NullString
	var exerciseOfficial bool
	if err := db.QueryRow(context.Background(), `
		SELECT owner_user_id::text, is_official
		FROM public.exercises
		WHERE LOWER(name) = LOWER($1)
		ORDER BY created_at DESC
		LIMIT 1
	`, exerciseName).Scan(&exerciseOwner, &exerciseOfficial); err != nil {
		t.Fatalf("failed to load created exercise: %v", err)
	}
	if exerciseOwner.Valid {
		if exerciseOwner.String != userID {
			t.Fatalf("expected created exercise owner_user_id to be %s, got %s", userID, exerciseOwner.String)
		}
	} else {
		t.Logf("exercise %q already existed in the catalogue (official=%t), so owner_user_id was not set by this run", exerciseName, exerciseOfficial)
	}
}

func mustVerifyRoutineExerciseCount(t *testing.T, db *pgxpool.Pool, routineID string) {
	t.Helper()

	var routineExerciseCount int
	if err := db.QueryRow(context.Background(), `
		SELECT COUNT(1)::int
		FROM public.routine_exercises
		WHERE routine_id = $1::uuid
	`, routineID).Scan(&routineExerciseCount); err != nil {
		t.Fatalf("failed to verify saved routine exercises: %v", err)
	}
	if routineExerciseCount < 1 {
		t.Fatalf("expected at least 1 saved routine exercise, got %d", routineExerciseCount)
	}
}
