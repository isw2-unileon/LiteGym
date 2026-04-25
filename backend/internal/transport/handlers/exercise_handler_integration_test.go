package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupExerciseHandlerTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return testutil.NewIntegrationTestPool(t)
}

func cleanupExercisesHandler(t *testing.T, db *pgxpool.Pool) {
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

func insertExerciseRawHandler(t *testing.T, db *pgxpool.Pool, exercise model.Exercise) string {
	t.Helper()

	var id string
	err := db.QueryRow(context.Background(), `
		INSERT INTO public.exercises (name, description, muscle_group, exercise_type, is_official)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text
	`, exercise.Name, exercise.Description, exercise.MuscleGroup, exercise.ExerciseType, exercise.IsOfficial).Scan(&id)
	if err != nil {
		t.Fatalf("error insertando ejercicio: %v", err)
	}

	if exercise.SecondaryMuscleGroup != "" {
		_, err = db.Exec(context.Background(), `
			INSERT INTO public.exercise_secondary_muscle_groups (exercise_id, muscle_group)
			VALUES ($1::uuid, $2)
		`, id, exercise.SecondaryMuscleGroup)
		if err != nil {
			t.Fatalf("error insertando secondary muscle group: %v", err)
		}
	}

	return id
}

func setupExerciseHandlerRouter(db *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)

	repo := repository.NewExerciseRepository(db)
	svc := service.NewExerciseService(repo)
	handler := NewExerciseHandler(svc)

	router := gin.New()
	router.GET("/api/exercises/:id", handler.GetExerciseByID)
	router.GET("/api/exercises", handler.ListExercises)
	router.DELETE("/api/exercises/:id", func(c *gin.Context) {
		c.Set("user_role", "admin")
		handler.DeleteExercise(c)
	})

	return router
}

func TestExerciseHandlerGetByIDIntegration(t *testing.T) {
	db := setupExerciseHandlerTestDB(t)
	cleanupExercisesHandler(t, db)

	insertedID := insertExerciseRawHandler(t, db, model.Exercise{
		Name:                 "Lat Pulldown",
		Description:          "Cable pulldown",
		MuscleGroup:          "back",
		SecondaryMuscleGroup: "biceps",
		ExerciseType:         "machine",
		IsOfficial:           true,
	})

	router := setupExerciseHandlerRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/api/exercises/"+insertedID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var exercise model.Exercise
	if err := json.Unmarshal(w.Body.Bytes(), &exercise); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if exercise.ID != insertedID {
		t.Fatalf("expected ID %s, got %s", insertedID, exercise.ID)
	}
	if exercise.Name != "Lat Pulldown" {
		t.Fatalf("expected name Lat Pulldown, got %s", exercise.Name)
	}
}

func TestExerciseHandlerGetByIDNotFoundIntegration(t *testing.T) {
	db := setupExerciseHandlerTestDB(t)
	cleanupExercisesHandler(t, db)

	router := setupExerciseHandlerRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/api/exercises/550e8400-e29b-41d4-a716-446655449999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestExerciseHandlerListIntegration(t *testing.T) {
	db := setupExerciseHandlerTestDB(t)
	cleanupExercisesHandler(t, db)

	insertExerciseRawHandler(t, db, model.Exercise{
		Name:         "Bench Press",
		Description:  "Flat bench press",
		MuscleGroup:  "chest",
		ExerciseType: "strength",
		IsOfficial:   true,
	})
	insertExerciseRawHandler(t, db, model.Exercise{
		Name:         "Squat",
		Description:  "Back squat",
		MuscleGroup:  "legs",
		ExerciseType: "strength",
		IsOfficial:   true,
	})

	router := setupExerciseHandlerRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/api/exercises", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var exercises []model.Exercise
	if err := json.Unmarshal(w.Body.Bytes(), &exercises); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if len(exercises) != 2 {
		t.Fatalf("expected 2 exercises, got %d", len(exercises))
	}
}

func TestExerciseHandlerDeleteIntegration(t *testing.T) {
	db := setupExerciseHandlerTestDB(t)
	cleanupExercisesHandler(t, db)

	insertedID := insertExerciseRawHandler(t, db, model.Exercise{
		Name:         "Leg Press",
		Description:  "Machine leg press",
		MuscleGroup:  "legs",
		ExerciseType: "machine",
		IsOfficial:   true,
	})

	router := setupExerciseHandlerRouter(db)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/exercises/"+insertedID, nil)
	deleteW := httptest.NewRecorder()
	router.ServeHTTP(deleteW, deleteReq)

	if deleteW.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, deleteW.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/exercises/"+insertedID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getW.Code)
	}
}

func TestExerciseHandlerDeleteOfficialRequiresAdminIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupExerciseHandlerTestDB(t)
	cleanupExercisesHandler(t, db)

	insertedID := insertExerciseRawHandler(t, db, model.Exercise{
		Name:         "Overhead Press",
		Description:  "Standing barbell press",
		MuscleGroup:  "shoulders",
		ExerciseType: "strength",
		IsOfficial:   true,
	})

	repo := repository.NewExerciseRepository(db)
	svc := service.NewExerciseService(repo)
	handler := NewExerciseHandler(svc)

	router := gin.New()
	router.DELETE("/api/exercises/:id", func(c *gin.Context) {
		c.Set("user_role", "user")
		handler.DeleteExercise(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/exercises/"+insertedID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusForbidden, w.Code, strings.TrimSpace(w.Body.String()))
	}
}
