package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
)

type MockWorkoutRepository struct {
	CreateSessionFunc                  func(ctx context.Context, workout *model.WorkoutSession) error
	GetSessionByIDFunc                 func(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error)
	GetSessionDetailByIDFunc           func(ctx context.Context, id uuid.UUID) (*model.WorkoutSessionDetail, error)
	UpdateSessionByIDFunc              func(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error
	RemoveSessionByIDFunc              func(ctx context.Context, id uuid.UUID) error
	CreateWorkoutExerciseFunc          func(ctx context.Context, workoutExercise *model.WorkoutExercise) error
	GetWorkoutExercisesBySessionIDFunc func(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error)
	CreateWorkoutSetFunc               func(ctx context.Context, workoutSet *model.WorkoutSet) error
	GetWorkoutSetsByExerciseIDFunc     func(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error)
	UpdateWorkoutSetFunc               func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error
	RemoveWorkoutSetFunc               func(ctx context.Context, setID uuid.UUID) error
	OriginalSessionData                *model.WorkoutSession
	OriginalSessionDataID              uuid.UUID
	OriginalSetData                    *model.WorkoutSet
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
	if m.GetWorkoutSetsByExerciseIDFunc != nil {
		return m.GetWorkoutSetsByExerciseIDFunc(ctx, exerciseID)
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

/* Testing */

/* Create Workout Session */
func TestWorkoutServiceCreateSessionSuccess(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)

	workoutSession := &model.WorkoutSession{
		UserID:    uuid.New(),
		RoutineID: uuidPointer(uuid.Nil),
		Name:      "Morning Workout",
	}

	err := service.CreateSession(context.Background(), workoutSession)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestWorkoutServiceCreateSessionInvalidInput(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)

	workoutSessionNoUser := &model.WorkoutSession{
		Name: "Morning Routine",
	}

	workoutSessionNoName := &model.WorkoutSession{
		UserID: uuid.Nil,
	}

	err := service.CreateSession(context.Background(), workoutSessionNoUser)

	if !errors.Is(err, ErrInvalidWorkoutSessionInput) {
		t.Fatalf("expected ErrInvalidWorkoutSessionInput, got %v", err)
	}

	err = service.CreateSession(context.Background(), workoutSessionNoName)

	if !errors.Is(err, ErrInvalidWorkoutSessionInput) {
		t.Fatalf("expected ErrInvalidWorkoutSessionInput, got %v", err)
	}
}

/* Get Workout Session by ID */
func TestWorkoutServiceGetSessionByIDSuccess(t *testing.T) {
	id, _ := uuid.Parse("f4622ac3-ed17-463f-98de-6affaa427a59")
	userID, _ := uuid.Parse("ff7e8cca-4103-4ef7-9c92-1d23975f81fc")

	mockRepo := &MockWorkoutRepository{
		GetSessionByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
			return &model.WorkoutSession{
				ID:     id,
				UserID: userID,
				Name:   "Morning Workout",
			}, nil
		},
	}
	service := NewWorkoutService(mockRepo)
	session, err := service.GetSessionByID(context.Background(), id)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session == nil {
		t.Fatal("expected workout session, got nil")
	}

	if session.Name != "Morning Workout" {
		t.Errorf("expected name 'Morning Workout', got '%s'", session.Name)
	}

	if session.UserID != userID {
		t.Errorf("expected user ID '%s', got '%s'", id, session.UserID)
	}

}

func TestWorkoutServiceGetSessionByIDInvalidInput(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)

	_, err := service.GetSessionByID(context.Background(), uuid.Nil)

	if !errors.Is(err, ErrInvalidWorkoutSessionInput) {
		t.Fatalf("expected ErrInvalidWorkoutSessionInput, got %v", err)
	}
}

func TestWorkoutServiceGetSessionByIDNotFound(t *testing.T) {
	mockRepo := &MockWorkoutRepository{
		GetSessionByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
			return nil, pgx.ErrNoRows
		},
	}
	service := NewWorkoutService(mockRepo)

	_, err := service.GetSessionByID(context.Background(), uuid.New())

	if !errors.Is(err, ErrWorkoutNotFound) {
		t.Fatalf("expected ErrWorkoutNotFound, got %v", err)
	}
}

func TestWorkoutServiceGetSessionDetailByIDSuccess(t *testing.T) {
	id := uuid.New()
	mockRepo := &MockWorkoutRepository{
		GetSessionDetailByIDFunc: func(ctx context.Context, receivedID uuid.UUID) (*model.WorkoutSessionDetail, error) {
			return &model.WorkoutSessionDetail{
				ID:   receivedID.String(),
				Name: "Push Day",
			}, nil
		},
	}

	service := NewWorkoutService(mockRepo)
	session, err := service.GetSessionDetailByID(context.Background(), id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if session == nil || session.ID != id.String() {
		t.Fatalf("expected workout detail for %s, got %#v", id, session)
	}
}

func TestWorkoutServiceGetSessionDetailByIDInvalidInput(t *testing.T) {
	service := NewWorkoutService(&MockWorkoutRepository{})
	_, err := service.GetSessionDetailByID(context.Background(), uuid.Nil)
	if !errors.Is(err, ErrInvalidWorkoutSessionInput) {
		t.Fatalf("expected ErrInvalidWorkoutSessionInput, got %v", err)
	}
}

func TestWorkoutServiceGetSessionDetailByIDNotFound(t *testing.T) {
	mockRepo := &MockWorkoutRepository{
		GetSessionDetailByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.WorkoutSessionDetail, error) {
			return nil, pgx.ErrNoRows
		},
	}
	service := NewWorkoutService(mockRepo)
	_, err := service.GetSessionDetailByID(context.Background(), uuid.New())
	if !errors.Is(err, ErrWorkoutNotFound) {
		t.Fatalf("expected ErrWorkoutNotFound, got %v", err)
	}
}

/* Update Workout Session by ID */
func TestWorkoutServiceUpdateSessionByIDInvalidInput(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)

	err := service.UpdateSessionByID(context.Background(), uuid.Nil, &model.WorkoutSession{})

	if !errors.Is(err, ErrInvalidWorkoutSessionInput) {
		t.Fatalf("expected ErrInvalidWorkoutSessionInput, got %v", err)
	}

	err = service.UpdateSessionByID(context.Background(), uuid.Nil, &model.WorkoutSession{
		Name: "",
	})

	if !errors.Is(err, ErrInvalidWorkoutSessionInput) {
		t.Fatalf("expected ErrInvalidWorkoutSessionInput, got %v", err)
	}
}

func TestWorkoutServiceFinishSessionSuccess(t *testing.T) {
	id := uuid.New()
	performedBefore := time.Now().Add(-24 * time.Hour)
	plannedAt := time.Now().Add(-48 * time.Hour)
	mockRepo := &MockWorkoutRepository{
		GetSessionByIDFunc: func(ctx context.Context, receivedID uuid.UUID) (*model.WorkoutSession, error) {
			return &model.WorkoutSession{
				ID:          receivedID,
				UserID:      uuid.New(),
				Name:        "Original",
				PerformedAt: &performedBefore,
				PlannedAt:   &plannedAt,
			}, nil
		},
		UpdateSessionByIDFunc: func(ctx context.Context, receivedID uuid.UUID, session *model.WorkoutSession) error {
			return nil
		},
	}
	service := NewWorkoutService(mockRepo)
	duration := 55
	err := service.FinishSession(context.Background(), id, &model.WorkoutFinishInput{
		Name:     "Finished Workout",
		Duration: &duration,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mockRepo.OriginalSessionData == nil {
		t.Fatal("expected updated session data to be forwarded to repository")
	}
	if mockRepo.OriginalSessionData.Name != "Finished Workout" {
		t.Fatalf("expected updated name, got %q", mockRepo.OriginalSessionData.Name)
	}
	if mockRepo.OriginalSessionData.PlannedAt == nil || !mockRepo.OriginalSessionData.PlannedAt.Equal(plannedAt) {
		t.Fatalf("expected planned_at to be preserved, got %#v", mockRepo.OriginalSessionData.PlannedAt)
	}
	if mockRepo.OriginalSessionData.PerformedAt == nil {
		t.Fatal("expected performed_at to be set")
	}
}

func TestWorkoutServiceUpdateSessionByIDNotFound(t *testing.T) {
	mockRepo := &MockWorkoutRepository{
		UpdateSessionByIDFunc: func(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error {
			return pgx.ErrNoRows
		},
	}
	service := NewWorkoutService(mockRepo)
	err := service.UpdateSessionByID(context.Background(), uuid.New(), &model.WorkoutSession{
		Name: "Morning Routine",
	})

	if !errors.Is(err, ErrWorkoutNotFound) {
		t.Fatalf("expected ErrWorkoutNotFound, got %v", err)
	}
}

func TestWorkoutServiceUpdateSessionByIDSuccess(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)

	updatedData := model.WorkoutSession{
		Name:        "Evening Workout",
		PerformedAt: timePointer(time.Now()),
	}
	err := service.UpdateSessionByID(context.Background(), uuid.New(), &updatedData)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockRepo.OriginalSessionData.Name != updatedData.Name {
		t.Errorf("expected name '%s', got '%s'", updatedData.Name, mockRepo.OriginalSessionData.Name)
	}

	if mockRepo.OriginalSessionData.PerformedAt == nil || !mockRepo.OriginalSessionData.PerformedAt.Equal(*updatedData.PerformedAt) {
		t.Errorf("expected performed_at %v, got %v", updatedData.PerformedAt, mockRepo.OriginalSessionData.PerformedAt)
	}
}

/* Remove Workout Session */
func TestWorkoutServiceRemoveSessionByIDInvalidInput(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)

	err := service.RemoveSessionByID(context.Background(), uuid.Nil)

	if !errors.Is(err, ErrInvalidWorkoutSessionInput) {
		t.Fatalf("expected ErrInvalidWorkoutSessionInput, got %v", err)
	}
}

func TestWorkoutServiceRemoveSessionByIDNotFound(t *testing.T) {
	mockRepo := &MockWorkoutRepository{
		RemoveSessionByIDFunc: func(ctx context.Context, id uuid.UUID) error {
			return pgx.ErrNoRows
		},
	}
	service := NewWorkoutService(mockRepo)

	err := service.RemoveSessionByID(context.Background(), uuid.New())

	if !errors.Is(err, ErrWorkoutNotFound) {
		t.Fatalf("expected ErrWorkoutNotFound, got %v", err)
	}
}

func TestWorkoutServiceRemoveSessionByIDSuccess(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)

	sessionID := uuid.New()

	err := service.RemoveSessionByID(context.Background(), sessionID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockRepo.OriginalSessionDataID != sessionID {
		t.Fatalf("expected ID '550e8400-e29b-41d4-a716-446655440000', got '%s'", mockRepo.OriginalSessionData.ID)
	}

}

/* Create Workout Exercise */
func TestWorkoutServiceCreateWorkoutExerciseInvalidInput(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)

	workoutExerciseNoSessionID := &model.WorkoutExercise{
		ExerciseID:    uuid.New(),
		ExerciseOrder: 1,
		Notes:         "this are notes",
	}

	workoutExerciseNoExerciseID := &model.WorkoutExercise{
		WorkoutSessionID: uuid.New(),
		ExerciseOrder:    2,
		Notes:            "this are notes",
	}

	workoutExerciseOrderZero := &model.WorkoutExercise{
		WorkoutSessionID: uuid.New(),
		ExerciseID:       uuid.New(),
		Notes:            "this are notes",
	}

	workoutExerciseNoNotes := &model.WorkoutExercise{
		WorkoutSessionID: uuid.New(),
		ExerciseID:       uuid.New(),
		ExerciseOrder:    3,
	}

	err := service.CreateWorkoutExercise(context.Background(), workoutExerciseNoSessionID)
	if !errors.Is(err, ErrInvalidWorkoutExerciseInput) {
		t.Fatalf("NoSessionID expected ErrInvalidWorkoutSessionInput, got %v", err)
	}
	err = service.CreateWorkoutExercise(context.Background(), workoutExerciseNoExerciseID)
	if !errors.Is(err, ErrInvalidWorkoutExerciseInput) {
		t.Fatalf("NoExerciseID expected ErrInvalidWorkoutSessionInput, got %v", err)
	}
	err = service.CreateWorkoutExercise(context.Background(), workoutExerciseOrderZero)
	if !errors.Is(err, ErrInvalidWorkoutExerciseInput) {
		t.Fatalf("OrderZero expected ErrInvalidWorkoutSessionInput, got %v", err)
	}
	err = service.CreateWorkoutExercise(context.Background(), workoutExerciseNoNotes)
	if !errors.Is(err, ErrInvalidWorkoutExerciseInput) {
		t.Fatalf("NoNotes expected ErrInvalidWorkoutSessionInput, got %v", err)
	}
}

func TestWorkoutServiceCreateWorkoutExerciseSuccess(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)
	workoutExercise := &model.WorkoutExercise{
		WorkoutSessionID: uuid.New(),
		ExerciseID:       uuid.New(),
		ExerciseOrder:    1,
		Notes:            "this are notes",
	}
	err := service.CreateWorkoutExercise(context.Background(), workoutExercise)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

/* Get Workout Exercises by Session ID */
func TestWorkoutServiceUpdateWorkoutExerciseInvalidInput(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)

	_, err := service.GetWorkoutExercisesBySessionID(context.Background(), uuid.Nil)

	if !errors.Is(err, ErrInvalidWorkoutExerciseInput) {
		t.Fatalf("expected ErrInvalidWorkoutExerciseInput, got %v", err)
	}
}

func TestWorkoutServiceUpdateWorkoutExerciseNotFound(t *testing.T) {
	mockRepo := &MockWorkoutRepository{
		GetWorkoutExercisesBySessionIDFunc: func(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error) {
			return nil, pgx.ErrNoRows
		},
	}
	service := NewWorkoutService(mockRepo)

	_, err := service.GetWorkoutExercisesBySessionID(context.Background(), uuid.New())

	if !errors.Is(err, ErrWorkoutExerciseNotFound) {
		t.Fatalf("expected ErrWorkoutExerciseNotFound, got %v", err)
	}
}

func TestWorkoutServiceUpdateWorkoutExerciseSuccess(t *testing.T) {
	workoutSessionID := uuid.New()
	mockRepo := &MockWorkoutRepository{
		GetWorkoutExercisesBySessionIDFunc: func(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error) {
			return []*model.WorkoutExercise{
				{
					ID:               uuid.New(),
					WorkoutSessionID: workoutSessionID,
					ExerciseID:       uuid.New(),
					ExerciseOrder:    1,
					Notes:            "this are notes",
				},
				{
					ID:               uuid.New(),
					WorkoutSessionID: workoutSessionID,
					ExerciseID:       uuid.New(),
					ExerciseOrder:    2,
					Notes:            "this are notes",
				},
			}, nil
		},
	}
	service := NewWorkoutService(mockRepo)
	workouts, err := service.GetWorkoutExercisesBySessionID(context.Background(), workoutSessionID)
	if len(workouts) != 2 {
		t.Fatalf("expected 2 workout, got %d", len(workouts))
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

/* Create Workout Set */
func TestWorkoutServiceCreateWorkoutSetInvalidInput(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)
	setNoExerciseID := &model.WorkoutSet{
		SetNumber:   1,
		Repetitions: intPointer(10),
		WeightKg:    floatPointer(100.5),
		Duration:    intPointer(10),
		DistanceKm:  floatPointer(0.0),
		Rir:         intPointer(2),
		Completed:   boolPointer(false),
	}
	setNoSetNumber := &model.WorkoutSet{
		WorkoutExerciseID: uuid.New(),
		Repetitions:       intPointer(10),
		WeightKg:          floatPointer(100.5),
		Duration:          intPointer(10),
		DistanceKm:        floatPointer(0.0),
		Rir:               intPointer(2),
		Completed:         boolPointer(false),
	}
	setNoCompletion := &model.WorkoutSet{
		WorkoutExerciseID: uuid.New(),
		SetNumber:         1,
		Repetitions:       intPointer(10),
		WeightKg:          floatPointer(100.5),
		Duration:          intPointer(10),
		DistanceKm:        floatPointer(0.0),
		Rir:               intPointer(2),
	}
	err := service.CreateWorkoutSet(context.Background(), setNoExerciseID)
	if !errors.Is(err, ErrInvalidWorkoutSetInput) {
		t.Fatalf("expected ErrInvalidWorkoutSetInput, got %v", err)
	}
	err = service.CreateWorkoutSet(context.Background(), setNoSetNumber)
	if !errors.Is(err, ErrInvalidWorkoutSetInput) {
		t.Fatalf("expected ErrInvalidWorkoutSetInput, got %v", err)
	}
	err = service.CreateWorkoutSet(context.Background(), setNoCompletion)
	if !errors.Is(err, ErrInvalidWorkoutSetInput) {
		t.Fatalf("expected ErrInvalidWorkoutSetInput, got %v", err)
	}
}

func TestWorkoutServiceCreateWorkoutSetSuccess(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)
	workoutSet := &model.WorkoutSet{
		WorkoutExerciseID: uuid.New(),
		SetNumber:         1,
		Repetitions:       intPointer(10),
		WeightKg:          floatPointer(100.5),
		Duration:          intPointer(10),
		DistanceKm:        floatPointer(0.0),
		Rir:               intPointer(2),
		Completed:         boolPointer(false),
	}
	err := service.CreateWorkoutSet(context.Background(), workoutSet)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

/* Get Workout Sets by Exercise ID */
func TestWorkoutServiceGetWorkoutSetsByExerciseIDInvalidInput(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)
	_, err := service.GetWorkoutSetsByWorkoutExerciseID(context.Background(), uuid.Nil)
	if !errors.Is(err, ErrInvalidWorkoutSetInput) {
		t.Fatalf("expected ErrInvalidWorkoutSetInput, got %v", err)
	}
}

func TestWorkoutServiceGetWorkoutSetsByExerciseIDNotFound(t *testing.T) {
	mockRepo := &MockWorkoutRepository{
		GetWorkoutSetsByExerciseIDFunc: func(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error) {
			return nil, pgx.ErrNoRows
		},
	}
	service := NewWorkoutService(mockRepo)
	_, err := service.GetWorkoutSetsByWorkoutExerciseID(context.Background(), uuid.New())
	if !errors.Is(err, ErrWorkoutSetNotFound) {
		t.Fatalf("expected ErrNoRows, got %v", err)
	}
}

func TestWorkoutServiceGetWorkoutSetsByWorkoutExerciseIDSuccess(t *testing.T) {
	workoutExerciseID := uuid.New()
	mockRepo := &MockWorkoutRepository{
		GetWorkoutSetsByExerciseIDFunc: func(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error) {
			return []*model.WorkoutSet{
				{
					ID:                uuid.New(),
					WorkoutExerciseID: workoutExerciseID,
					SetNumber:         1,
					Repetitions:       intPointer(10),
					WeightKg:          floatPointer(100.5),
					Duration:          intPointer(10),
					DistanceKm:        floatPointer(0.0),
					Rir:               intPointer(2),
					Completed:         boolPointer(false),
				},
				{
					ID:                uuid.New(),
					WorkoutExerciseID: workoutExerciseID,
					SetNumber:         2,
					Repetitions:       intPointer(10),
					WeightKg:          floatPointer(100.5),
					Duration:          intPointer(10),
					DistanceKm:        floatPointer(0.0),
					Rir:               intPointer(2),
					Completed:         boolPointer(false),
				},
			}, nil
		},
	}
	service := NewWorkoutService(mockRepo)
	sets, err := service.GetWorkoutSetsByWorkoutExerciseID(context.Background(), workoutExerciseID)
	if len(sets) != 2 {
		t.Fatalf("expected 2 sets, got %d", len(sets))
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

/* Update Workout Set */
func TestWorkoutServiceUpdateWorkoutSetInvalidInput(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)
	setNoSetNumber := &model.WorkoutSet{
		Repetitions: intPointer(10),
		WeightKg:    floatPointer(100.5),
		Duration:    intPointer(10),
		DistanceKm:  floatPointer(0.0),
		Rir:         intPointer(2),
		Completed:   boolPointer(false),
	}
	setZeroSetNumber := &model.WorkoutSet{
		SetNumber:   0,
		Repetitions: intPointer(10),
		WeightKg:    floatPointer(100.5),
		Duration:    intPointer(10),
		DistanceKm:  floatPointer(0.0),
		Rir:         intPointer(2),
		Completed:   boolPointer(false),
	}
	setNoCompletion := &model.WorkoutSet{
		SetNumber:   1,
		Repetitions: intPointer(10),
		WeightKg:    floatPointer(100.5),
		Duration:    intPointer(10),
		DistanceKm:  floatPointer(0.0),
		Rir:         intPointer(2),
	}

	err := service.UpdateWorkoutSet(context.Background(), uuid.New(), setNoSetNumber)
	if !errors.Is(err, ErrInvalidWorkoutSetInput) {
		t.Fatalf("NoSetID expected ErrInvalidWorkoutSetInput, got %v", err)
	}
	err = service.UpdateWorkoutSet(context.Background(), uuid.New(), setNoSetNumber)
	if !errors.Is(err, ErrInvalidWorkoutSetInput) {
		t.Fatalf("NoSetNumber expected ErrInvalidWorkoutSetInput, got %v", err)
	}
	err = service.UpdateWorkoutSet(context.Background(), uuid.New(), setZeroSetNumber)
	if !errors.Is(err, ErrInvalidWorkoutSetInput) {
		t.Fatalf("ZeroSetNumber expected ErrInvalidWorkoutSetInput, got %v", err)
	}
	err = service.UpdateWorkoutSet(context.Background(), uuid.New(), setNoCompletion)
	if !errors.Is(err, ErrInvalidWorkoutSetInput) {
		t.Fatalf("NoCompletion expected ErrInvalidWorkoutSetInput, got %v", err)
	}
}

func TestWorkoutServiceUpdateWorkoutSetNotFound(t *testing.T) {
	mockRepo := &MockWorkoutRepository{
		UpdateWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID, setNumber int, set *model.WorkoutSet) error {
			return pgx.ErrNoRows
		},
	}
	service := NewWorkoutService(mockRepo)
	set := &model.WorkoutSet{
		SetNumber:   1,
		Repetitions: intPointer(10),
		WeightKg:    floatPointer(100.5),
		Duration:    intPointer(10),
		DistanceKm:  floatPointer(0.0),
		Rir:         intPointer(2),
		Completed:   boolPointer(true),
	}
	err := service.UpdateWorkoutSet(context.Background(), uuid.New(), set)
	if !errors.Is(err, ErrWorkoutSetNotFound) {
		t.Fatalf("expected ErrWorkoutSetNotFound, got %v", err)
	}
}

func TestWorkoutServiceUpdateWorkoutSetSuccess(t *testing.T) {
	mockRepo := &MockWorkoutRepository{}
	service := NewWorkoutService(mockRepo)
	updatedSet := &model.WorkoutSet{
		SetNumber:   3,
		Repetitions: intPointer(12),
		WeightKg:    floatPointer(105.0),
		Duration:    intPointer(12),
		DistanceKm:  floatPointer(0.0),
		Rir:         intPointer(1),
		Completed:   boolPointer(true),
	}
	err := service.UpdateWorkoutSet(context.Background(), uuid.New(), updatedSet)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mockRepo.OriginalSetData.Repetitions != updatedSet.Repetitions {
		t.Fatalf("expected repetitions %d, got %d", *updatedSet.Repetitions, *mockRepo.OriginalSetData.Repetitions)
	}
}

func TestWorkoutServiceRemoveWorkoutSetInvalidInput(t *testing.T) {
	service := NewWorkoutService(&MockWorkoutRepository{})
	err := service.RemoveWorkoutSet(context.Background(), uuid.Nil)
	if !errors.Is(err, ErrInvalidWorkoutSetInput) {
		t.Fatalf("expected ErrInvalidWorkoutSetInput, got %v", err)
	}
}

func TestWorkoutServiceRemoveWorkoutSetNotFound(t *testing.T) {
	mockRepo := &MockWorkoutRepository{
		RemoveWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID) error {
			return pgx.ErrNoRows
		},
	}
	service := NewWorkoutService(mockRepo)
	err := service.RemoveWorkoutSet(context.Background(), uuid.New())
	if !errors.Is(err, ErrWorkoutSetNotFound) {
		t.Fatalf("expected ErrWorkoutSetNotFound, got %v", err)
	}
}

func TestWorkoutServiceRemoveWorkoutSetSuccess(t *testing.T) {
	mockRepo := &MockWorkoutRepository{
		RemoveWorkoutSetFunc: func(ctx context.Context, setID uuid.UUID) error {
			return nil
		},
	}
	service := NewWorkoutService(mockRepo)
	if err := service.RemoveWorkoutSet(context.Background(), uuid.New()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func timePointer(t time.Time) *time.Time {
	return &t
}

func intPointer(i int) *int {
	return &i
}

func floatPointer(f float64) *float64 {
	return &f
}

func boolPointer(b bool) *bool {
	return &b
}

func uuidPointer(u uuid.UUID) *uuid.UUID {
	return &u
}
