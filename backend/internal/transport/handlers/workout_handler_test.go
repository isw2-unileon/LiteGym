package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

type MockWorkoutRepository struct {
	CreateSessionFunc                     func(ctx context.Context, workout *model.WorkoutSession) error
	GetSessionByIDFunc                    func(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error)
	GetSessionDetailByIDFunc              func(ctx context.Context, id uuid.UUID) (*model.WorkoutSessionDetail, error)
	UpdateSessionByIDFunc                 func(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error
	RemoveSessionByIDFunc                 func(ctx context.Context, id uuid.UUID) error
	CreateWorkoutExerciseFunc             func(ctx context.Context, workoutExercise *model.WorkoutExercise) error
	GetWorkoutExercisesBySessionIDFunc    func(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error)
	CreateWorkoutSetFunc                  func(ctx context.Context, workoutSet *model.WorkoutSet) error
	GetWorkoutSetsByWorkoutExerciseIDFunc func(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error)
	UpdateWorkoutSetFunc                  func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error
	RemoveWorkoutSetFunc                  func(ctx context.Context, setID uuid.UUID) error
	OriginalSessionData                   *model.WorkoutSession
	OriginalSessionDataID                 uuid.UUID
	OriginalSetData                       *model.WorkoutSet
}

/* Mock Repository functions */

/* Workout Session */
func (m *MockWorkoutRepository) CreateSession(ctx context.Context, workout *model.WorkoutSession) error {
	if m.CreateSessionFunc != nil {
		return m.CreateSessionFunc(ctx, workout)
	}
	return nil
}

func (m *MockWorkoutRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
	if m.GetSessionByIDFunc != nil {
		return m.GetSessionByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockWorkoutRepository) GetSessionDetailByID(ctx context.Context, id uuid.UUID) (*model.WorkoutSessionDetail, error) {
	if m.GetSessionDetailByIDFunc != nil {
		return m.GetSessionDetailByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockWorkoutRepository) UpdateSessionByID(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error {
	m.OriginalSessionData = session
	if m.UpdateSessionByIDFunc != nil {
		return m.UpdateSessionByIDFunc(ctx, id, session)
	}
	return nil
}

func (m *MockWorkoutRepository) RemoveSessionByID(ctx context.Context, id uuid.UUID) error {
	m.OriginalSessionDataID = id
	if m.RemoveSessionByIDFunc != nil {
		return m.RemoveSessionByIDFunc(ctx, id)
	}
	return nil
}

/* Workout exercise */
func (m *MockWorkoutRepository) CreateWorkoutExercise(ctx context.Context, workoutExercise *model.WorkoutExercise) error {
	if m.CreateWorkoutExerciseFunc != nil {
		return m.CreateWorkoutExerciseFunc(ctx, workoutExercise)
	}
	return nil
}

func (m *MockWorkoutRepository) GetWorkoutExercisesBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error) {
	if m.GetWorkoutExercisesBySessionIDFunc != nil {
		return m.GetWorkoutExercisesBySessionIDFunc(ctx, sessionID)
	}
	return nil, nil
}

/* Workout Set*/
func (m *MockWorkoutRepository) CreateWorkoutSet(ctx context.Context, workoutSet *model.WorkoutSet) error {
	if m.CreateWorkoutSetFunc != nil {
		return m.CreateWorkoutSetFunc(ctx, workoutSet)
	}
	return nil
}

func (m *MockWorkoutRepository) GetWorkoutSetsByWorkoutExerciseID(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error) {
	if m.GetWorkoutSetsByWorkoutExerciseIDFunc != nil {
		return m.GetWorkoutSetsByWorkoutExerciseIDFunc(ctx, exerciseID)
	}
	return nil, nil
}

func (m *MockWorkoutRepository) UpdateWorkoutSet(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
	m.OriginalSetData = set
	if m.UpdateWorkoutSetFunc != nil {
		return m.UpdateWorkoutSetFunc(ctx, setID, setNumber, set)
	}
	return nil
}

func (m *MockWorkoutRepository) RemoveWorkoutSet(ctx context.Context, setID uuid.UUID) error {
	if m.RemoveWorkoutSetFunc != nil {
		return m.RemoveWorkoutSetFunc(ctx, setID)
	}
	return nil
}

/* Create Workout */

func TestWorkoutHandlerCreateWorkout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateSessionFunc: func(ctx context.Context, workout *model.WorkoutSession) error {
			workout.ID = uuid.New()
			workout.CreatedAt = time.Now()
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	jsonPayload := fmt.Sprintf(`{"user_id":"%s","name":"Routine 1","notes":"This are notes"}`, uuid.New().String())
	ctx.Request = httptest.NewRequest("POST", "/api/workout", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")

	workoutHandler.CreateWorkout(ctx)
	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestWorkoutHandlerCreateWorkoutUsesAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedUserID uuid.UUID
	mockRepo := &MockWorkoutRepository{
		CreateSessionFunc: func(ctx context.Context, workout *model.WorkoutSession) error {
			capturedUserID = workout.UserID
			workout.ID = uuid.New()
			workout.CreatedAt = time.Now()
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	authenticatedUserID := uuid.New()
	ctx.Set(middleware.ContextUserIDKey, authenticatedUserID.String())
	ctx.Request = httptest.NewRequest("POST", "/api/workout", bytes.NewBufferString(`{"name":"Routine 1"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	workoutHandler.CreateWorkout(ctx)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
	if capturedUserID != authenticatedUserID {
		t.Fatalf("expected authenticated user %s, got %s", authenticatedUserID, capturedUserID)
	}
}

func TestWorkoutHandlerCreateWorkoutInvalidJSONName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateSessionFunc: func(ctx context.Context, workout *model.WorkoutSession) error {
			workout.ID = uuid.New()
			workout.CreatedAt = time.Now()
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("POST", "/api/workout", bytes.NewBufferString(`{"name"": }`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	workoutHandler.CreateWorkout(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

}

func TestWorkoutHandlerCreateWorkoutInvalidJSONUserUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateSessionFunc: func(ctx context.Context, workout *model.WorkoutSession) error {
			workout.ID = uuid.New()
			workout.CreatedAt = time.Now()
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("POST", "/api/workout", bytes.NewBufferString(`{"name": "Routine 1", "user_id" : "213"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	workoutHandler.CreateWorkout(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

}

func TestWorkoutHandlerCreateWorkoutInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateSessionFunc: func(ctx context.Context, workout *model.WorkoutSession) error {
			return errors.New("database error")
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	jsonPayload := fmt.Sprintf(`{"user_id":"%s","name":"Routine 1","notes":"This are notes"}`, uuid.New().String())
	ctx.Request = httptest.NewRequest("POST", "/api/workout", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")

	workoutHandler.CreateWorkout(ctx)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

/* Finish Workout */

func TestWorkoutHandlerFinishWorkoutSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetSessionByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
			return &model.WorkoutSession{ID: id, UserID: uuid.New(), Name: "Routine 1"}, nil
		},
		UpdateSessionByIDFunc: func(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/finish/"+workoutID.String(), bytes.NewBufferString(`{"Name": "Routine 1"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.FinishWorkout(ctx)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestWorkoutHandlerFinishWorkoutInvalidName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := &MockWorkoutRepository{
		UpdateSessionByIDFunc: func(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	workoutID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/finish/"+workoutID.String(), bytes.NewBufferString(`{"Name": ""}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.FinishWorkout(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerFinishWorkoutInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := &MockWorkoutRepository{
		UpdateSessionByIDFunc: func(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	workoutID := uuid.Nil
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/finish/"+workoutID.String(), bytes.NewBufferString(`{"Name": ""}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.FinishWorkout(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerFinishWorkoutNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := &MockWorkoutRepository{
		UpdateSessionByIDFunc: func(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error {
			return service.ErrWorkoutNotFound
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	workoutID := uuid.Nil
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/finish/"+workoutID.String(), bytes.NewBufferString(`{"Name": "A"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.FinishWorkout(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerFinishWorkoutInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := &MockWorkoutRepository{
		GetSessionByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
			return &model.WorkoutSession{ID: id, UserID: uuid.New(), Name: "Routine 1"}, nil
		},
		UpdateSessionByIDFunc: func(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error {
			return errors.New("database error")
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	workoutID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/finish/"+workoutID.String(), bytes.NewBufferString(`{"Name": "A"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.FinishWorkout(ctx)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

/* Get Workout by ID */

func TestWorkoutHandlerGetWorkoutByIDSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetSessionByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
			return &model.WorkoutSession{ID: id, UserID: uuid.New(), Name: "Push Day"}, nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String(), nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetWorkoutByID(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestWorkoutHandlerGetWorkoutByIDInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetSessionByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
			return nil, nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.Nil
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String(), nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetWorkoutByID(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerGetWorkoutByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetSessionByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
			return nil, service.ErrWorkoutNotFound
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String(), nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetWorkoutByID(ctx)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestWorkoutHandlerGetWorkoutByIDInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetSessionByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
			return nil, errors.New("database error")
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String(), nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetWorkoutByID(ctx)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

/* Remove Workout */

func TestWorkoutHandlerRemoveWorkoutByIDSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		RemoveSessionByIDFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String(), nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.RemoveWorkout(ctx)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestWorkoutHandlerRemoveWorkoutByIDInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		RemoveSessionByIDFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.Nil
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String(), nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.RemoveWorkout(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerRemoveWorkoutByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		RemoveSessionByIDFunc: func(ctx context.Context, id uuid.UUID) error {
			return service.ErrWorkoutNotFound
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String(), nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.RemoveWorkout(ctx)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestWorkoutHandlerRemoveWorkoutByIDInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		RemoveSessionByIDFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("database error")
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String(), nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.RemoveWorkout(ctx)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

/* Create Workout Exercise */

func TestWorkoutHandlerCreateWorkoutExerciseSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutExerciseFunc: func(ctx context.Context, exercise *model.WorkoutExercise) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()

	jsonPayload := fmt.Sprintf(`{"exercise_id":"%s","exercise_order":3,"notes":"These are notes"}`, exerciseID.String())
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercise", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutExercise(ctx)
	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d. %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestWorkoutHandlerCreateWorkoutExerciseInvalidWorkoutID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutExerciseFunc: func(ctx context.Context, exercise *model.WorkoutExercise) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.Nil
	exerciseID := uuid.New()

	jsonPayload := fmt.Sprintf(`{"exercise_id":"%s","exercise_order":3,"notes":"These are notes"}`, exerciseID.String())
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercise", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutExercise(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d. %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestWorkoutHandlerCreateWorkoutExerciseInvalidExerciseID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutExerciseFunc: func(ctx context.Context, exercise *model.WorkoutExercise) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.Nil

	jsonPayload := fmt.Sprintf(`{"exercise_id":"%s","exercise_order":3,"notes":"These are notes"}`, exerciseID.String())
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercise", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutExercise(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d. %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestWorkoutHandlerCreateWorkoutExerciseInvalidExerciseOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutExerciseFunc: func(ctx context.Context, exercise *model.WorkoutExercise) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()

	jsonPayload := fmt.Sprintf(`{"exercise_id":"%s","exercise_order":0,"notes":"These are notes"}`, exerciseID.String())
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercise", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutExercise(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d. %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestWorkoutHandlerCreateWorkoutExerciseInvalidNotes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutExerciseFunc: func(ctx context.Context, exercise *model.WorkoutExercise) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()

	jsonPayload := fmt.Sprintf(`{"exercise_id":"%s","exercise_order":3,"notes":""}`, exerciseID.String())
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercise", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutExercise(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d. %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestWorkoutHandlerCreateWorkoutExerciseInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutExerciseFunc: func(ctx context.Context, exercise *model.WorkoutExercise) error {
			return errors.New("database error")
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()

	jsonPayload := fmt.Sprintf(`{"exercise_id":"%s","exercise_order":3,"notes":"These are notes"}`, exerciseID.String())
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercise", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutExercise(ctx)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d. %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
}

func TestWorkoutHandlerCreateWorkoutExerciseInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutExerciseFunc: func(ctx context.Context, exercise *model.WorkoutExercise) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()

	jsonPayload := `{"exercise_order":3,"notes":"These are notes"}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercise", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutExercise(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d. %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

/* Get Exercise by Workout ID */

func TestWorkoutHandlerGetExercisesByWorkoutIDSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetWorkoutExercisesBySessionIDFunc: func(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error) {
			return nil, nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String()+"/exercises", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetExercisesByWorkoutID(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestWorkoutHandlerGetWorkoutDetailByIDSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workoutID := uuid.New()
	mockRepo := &MockWorkoutRepository{
		GetSessionDetailByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.WorkoutSessionDetail, error) {
			return &model.WorkoutSessionDetail{
				ID:   id.String(),
				Name: "Push Day",
			}, nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String()+"/detail", nil)

	workoutHandler.GetWorkoutDetailByID(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestWorkoutHandlerGetWorkoutDetailByIDInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workoutService := service.NewWorkoutService(&MockWorkoutRepository{})
	workoutHandler := NewWorkoutHandler(workoutService)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Params = gin.Params{{Key: "id", Value: "invalid"}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/invalid/detail", nil)

	workoutHandler.GetWorkoutDetailByID(ctx)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerGetExercisesByWorkoutIDInvalidSessionID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetWorkoutExercisesBySessionIDFunc: func(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error) {
			return nil, nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.Nil
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String()+"/exercises", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetExercisesByWorkoutID(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerGetExercisesByWorkoutIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetWorkoutExercisesBySessionIDFunc: func(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error) {
			return nil, service.ErrWorkoutExerciseNotFound
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String()+"/exercises", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetExercisesByWorkoutID(ctx)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestWorkoutHandlerGetExercisesByWorkoutIDInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetWorkoutExercisesBySessionIDFunc: func(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error) {
			return nil, errors.New("database error")
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String()+"/exercises", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetExercisesByWorkoutID(ctx)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

/* Create Workout Set */

func TestWorkoutHandlerCreateWorkoutSetSuccessFalse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutSetFunc: func(ctx context.Context, exercise *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()

	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":false}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercises/"+
		exerciseID.String()+"/set", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutSet(ctx)
	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d. %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestWorkoutHandlerCreateWorkoutSetSuccessTrue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutSetFunc: func(ctx context.Context, exercise *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()

	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":true}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercises/"+
		exerciseID.String()+"/set", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutSet(ctx)
	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d. %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestWorkoutHandlerCreateWorkoutSetInvalidWorkoutID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutSetFunc: func(ctx context.Context, exercise *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.Nil
	exerciseID := uuid.New()

	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":true}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercises/"+
		exerciseID.String()+"/set", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutSet(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d. %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestWorkoutHandlerCreateWorkoutSetInvalidExerciseID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutSetFunc: func(ctx context.Context, exercise *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.Nil

	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":true}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercises/"+
		exerciseID.String()+"/set", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutSet(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d. %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestWorkoutHandlerCreateWorkoutSetInvalidSetNumber(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutSetFunc: func(ctx context.Context, exercise *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()

	jsonPayload := `{"set_number":0,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":true}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercises/"+
		exerciseID.String()+"/set", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutSet(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d. %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestWorkoutHandlerCreateWorkoutSetInvalidCompleted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutSetFunc: func(ctx context.Context, exercise *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()

	// Completed omitted
	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercises/"+
		exerciseID.String()+"/set", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutSet(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d. %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestWorkoutHandlerCreateWorkoutSetInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		CreateWorkoutSetFunc: func(ctx context.Context, exercise *model.WorkoutSet) error {
			return errors.New("database error")
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()

	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":true}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+"/exercises/"+
		exerciseID.String()+"/set", bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.CreateWorkoutSet(ctx)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d. %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
}

/* Get Workout Sets by Exercise ID */

func TestWorkoutHandlerGetWorkoutSetsByExerciseIDSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetWorkoutSetsByWorkoutExerciseIDFunc: func(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error) {
			return nil, nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String()+"/exercises/"+
		exerciseID.String()+"/sets", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetWorkoutSetsByExerciseID(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestWorkoutHandlerGetWorkoutSetsByExerciseIDInvalidWorkoutID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetWorkoutSetsByWorkoutExerciseIDFunc: func(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error) {
			return nil, nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.Nil
	exerciseID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String()+"/exercises/"+
		exerciseID.String()+"/sets", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetWorkoutSetsByExerciseID(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerGetWorkoutSetsByExerciseIDInvalidExerciseID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetWorkoutSetsByWorkoutExerciseIDFunc: func(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error) {
			return nil, nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.Nil
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String()+"/exercises/"+
		exerciseID.String()+"/sets", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetWorkoutSetsByExerciseID(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d. %s", http.StatusBadRequest, w.Code, w.Body)
	}
}

func TestWorkoutHandlerGetWorkoutSetsByExerciseIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetWorkoutSetsByWorkoutExerciseIDFunc: func(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error) {
			return nil, service.ErrWorkoutSetNotFound
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String()+"/exercises/"+
		exerciseID.String()+"/sets", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetWorkoutSetsByExerciseID(ctx)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d. %s", http.StatusNotFound, w.Code, w.Body)
	}
}

func TestWorkoutHandlerGetWorkoutSetsByExerciseIDInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		GetWorkoutSetsByWorkoutExerciseIDFunc: func(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error) {
			return nil, errors.New("database error")
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}}
	ctx.Request = httptest.NewRequest("GET", "/api/workout/"+workoutID.String()+"/exercises/"+
		exerciseID.String()+"/sets", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.GetWorkoutSetsByExerciseID(ctx)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d. %s", http.StatusInternalServerError, w.Code, w.Body)
	}
}

/* Update Workout Set */

func TestWorkoutHandlerUpdateWorkoutSetSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		UpdateWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()
	setID := uuid.New()

	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":true}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}, {Key: "set_id", Value: setID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+
		"/exercises/"+exerciseID.String()+"/sets/"+setID.String(), bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.UpdateWorkoutSet(ctx)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestWorkoutHandlerUpdateWorkoutSetInvalidWorkoutID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		UpdateWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.Nil
	exerciseID := uuid.New()
	setID := uuid.New()

	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":true}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}, {Key: "set_id", Value: setID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+
		"/exercises/"+exerciseID.String()+"/sets/"+setID.String(), bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.UpdateWorkoutSet(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerUpdateWorkoutSetInvalidExerciseID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		UpdateWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.Nil
	setID := uuid.New()

	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":true}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}, {Key: "set_id", Value: setID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+
		"/exercises/"+exerciseID.String()+"/sets/"+setID.String(), bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.UpdateWorkoutSet(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerUpdateWorkoutSetInvalidSetID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		UpdateWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()
	setID := uuid.Nil

	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":true}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}, {Key: "set_id", Value: setID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+
		"/exercises/"+exerciseID.String()+"/sets/"+setID.String(), bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.UpdateWorkoutSet(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerUpdateWorkoutSetInvalidSeNumber(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		UpdateWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()
	setID := uuid.New()

	jsonPayload := `{"set_number":0,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":true}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}, {Key: "set_id", Value: setID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+
		"/exercises/"+exerciseID.String()+"/sets/"+setID.String(), bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.UpdateWorkoutSet(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerUpdateWorkoutSetInvalidCompleted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		UpdateWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()
	setID := uuid.New()

	// Completed omitted
	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}, {Key: "set_id", Value: setID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+
		"/exercises/"+exerciseID.String()+"/sets/"+setID.String(), bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.UpdateWorkoutSet(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerUpdateWorkoutSetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		UpdateWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
			return service.ErrWorkoutSetNotFound
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()
	setID := uuid.New()

	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":true}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}, {Key: "set_id", Value: setID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+
		"/exercises/"+exerciseID.String()+"/sets/"+setID.String(), bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.UpdateWorkoutSet(ctx)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestWorkoutHandlerUpdateWorkoutSetInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		UpdateWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()
	setID := uuid.Nil

	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}, {Key: "set_id", Value: setID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+
		"/exercises/"+exerciseID.String()+"/sets/"+setID.String(), bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.UpdateWorkoutSet(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerUpdateWorkoutSetInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		UpdateWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
			return errors.New("internal server error")
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	workoutID := uuid.New()
	exerciseID := uuid.New()
	setID := uuid.New()

	jsonPayload := `{"set_number":1,"repetitions":10,"weight_kg":50.0,"duration":30,"distance_km":0.0,"rir":2,"completed":true}`
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}, {Key: "set_id", Value: setID.String()}}
	ctx.Request = httptest.NewRequest("POST", "/api/workout/"+workoutID.String()+
		"/exercises/"+exerciseID.String()+"/sets/"+setID.String(), bytes.NewBufferString(jsonPayload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	workoutHandler.UpdateWorkoutSet(ctx)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestWorkoutHandlerRemoveWorkoutSetSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		RemoveWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID) error {
			return nil
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	workoutID := uuid.New()
	exerciseID := uuid.New()
	setID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}, {Key: "set_id", Value: setID.String()}}
	ctx.Request = httptest.NewRequest("DELETE", "/api/workout/"+workoutID.String()+"/exercises/"+exerciseID.String()+"/sets/"+setID.String(), nil)

	workoutHandler.RemoveWorkoutSet(ctx)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestWorkoutHandlerRemoveWorkoutSetInvalidSetID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workoutService := service.NewWorkoutService(&MockWorkoutRepository{})
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Params = gin.Params{{Key: "id", Value: uuid.New().String()}, {Key: "exercise_id", Value: uuid.New().String()}, {Key: "set_id", Value: "invalid"}}
	ctx.Request = httptest.NewRequest("DELETE", "/api/workout/x/exercises/y/sets/z", nil)

	workoutHandler.RemoveWorkoutSet(ctx)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWorkoutHandlerRemoveWorkoutSetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockWorkoutRepository{
		RemoveWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID) error {
			return service.ErrWorkoutSetNotFound
		},
	}
	workoutService := service.NewWorkoutService(mockRepo)
	workoutHandler := NewWorkoutHandler(workoutService)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	workoutID := uuid.New()
	exerciseID := uuid.New()
	setID := uuid.New()
	ctx.Params = gin.Params{{Key: "id", Value: workoutID.String()}, {Key: "exercise_id", Value: exerciseID.String()}, {Key: "set_id", Value: setID.String()}}
	ctx.Request = httptest.NewRequest("DELETE", "/api/workout/"+workoutID.String()+"/exercises/"+exerciseID.String()+"/sets/"+setID.String(), nil)

	workoutHandler.RemoveWorkoutSet(ctx)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
