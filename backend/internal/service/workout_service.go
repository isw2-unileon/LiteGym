package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/jackc/pgx/v5"
)

/* Workout Session Errors */

// ErrInvalidWorkoutSessionInput indicates that the provided workout data is invalid.
var ErrInvalidWorkoutSessionInput = errors.New("invalid workout session input")

// ErrWorkoutNotFound indicates that the requested workout does not exist.
var ErrWorkoutNotFound = errors.New("workout session not found")

/* Workout Exercise Errors */

// ErrInvalidWorkoutExerciseInput indicates that the provided workout exercise data is invalid.
var ErrInvalidWorkoutExerciseInput = errors.New("invalid workout exercise input")

// ErrWorkoutExerciseNotFound indicates that the requested workout exercise does not exist.
var ErrWorkoutExerciseNotFound = errors.New("workout exercise not found")

/* Workout Set Errors */

// ErrInvalidWorkoutSetInput indicates that the provided workout set data is invalid.
var ErrInvalidWorkoutSetInput = errors.New("invalid workout set input")

// ErrWorkoutSetNotFound indicates that the requested workout set does not exist.
var ErrWorkoutSetNotFound = errors.New("workout set not found")

// WorkoutService provides business logic for workouts.
type WorkoutService struct {
	repo repository.WorkoutRepository
}

// NewWorkoutService creates a new WorkoutService.
func NewWorkoutService(repo repository.WorkoutRepository) *WorkoutService {
	return &WorkoutService{
		repo: repo,
	}
}

/* Workout Session Methods */

// CreateSession validates and stores a new workoutSession.
func (ws *WorkoutService) CreateSession(ctx context.Context, workoutSession *model.WorkoutSession) error {
	if workoutSession == nil || workoutSession.Name == "" || workoutSession.UserID == uuid.Nil {
		return ErrInvalidWorkoutSessionInput
	}
	return ws.repo.CreateSession(ctx, workoutSession)
}

// GetSessionByID retrieves a workout session by its ID.
func (ws *WorkoutService) GetSessionByID(ctx context.Context, id uuid.UUID) (*model.WorkoutSession, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidWorkoutSessionInput
	}
	session, err := ws.repo.GetSessionByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWorkoutNotFound
	}
	if err != nil {
		return nil, err
	}
	return session, nil
}

// UpdateSessionByID updates an existing workout session by its ID.
func (ws *WorkoutService) UpdateSessionByID(ctx context.Context, id uuid.UUID, session *model.WorkoutSession) error {
	if id == uuid.Nil || session.Name == "" {
		return ErrInvalidWorkoutSessionInput
	}
	err := ws.repo.UpdateSessionByID(ctx, id, session)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrWorkoutNotFound
	}
	if err != nil {
		return err
	}
	return nil
}

// RemoveSessionByID deletes a workout session by its ID.
func (ws *WorkoutService) RemoveSessionByID(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidWorkoutSessionInput
	}
	err := ws.repo.RemoveSessionByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrWorkoutNotFound
	}
	if err != nil {
		return err
	}
	return nil
}

/* Workout Exercise Methods */

// CreateWorkoutExercise validates and stores a new workout exercise.
func (ws *WorkoutService) CreateWorkoutExercise(ctx context.Context, workoutExercise *model.WorkoutExercise) error {
	if workoutExercise == nil || workoutExercise.ExerciseOrder <= 0 || workoutExercise.ExerciseID == uuid.Nil ||
		workoutExercise.WorkoutSessionID == uuid.Nil || workoutExercise.Notes == "" {
		return ErrInvalidWorkoutExerciseInput
	}
	return ws.repo.CreateWorkoutExercise(ctx, workoutExercise)
}

// GetWorkoutExercisesBySessionID retrieves all workout exercises associated with a workout session ID.
func (ws *WorkoutService) GetWorkoutExercisesBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*model.WorkoutExercise, error) {
	if sessionID == uuid.Nil {
		return nil, ErrInvalidWorkoutExerciseInput
	}
	exercise, err := ws.repo.GetWorkoutExercisesBySessionID(ctx, sessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWorkoutExerciseNotFound
	}
	if err != nil {
		return nil, err
	}
	return exercise, nil
}

/* Workout Set Methods */

// CreateWorkoutSet validates and stores a new workout set.
func (ws *WorkoutService) CreateWorkoutSet(ctx context.Context, workoutSet *model.WorkoutSet) error {
	if workoutSet == nil || workoutSet.SetNumber <= 0 || workoutSet.WorkoutExerciseID == uuid.Nil || workoutSet.Completed == nil {
		return ErrInvalidWorkoutSetInput
	}
	return ws.repo.CreateWorkoutSet(ctx, workoutSet)
}

// GetWorkoutSetsByWorkoutExerciseID retrieves all workout sets associated with a workout exercise ID.
func (ws *WorkoutService) GetWorkoutSetsByWorkoutExerciseID(ctx context.Context, exerciseID uuid.UUID) ([]*model.WorkoutSet, error) {
	if exerciseID == uuid.Nil {
		return nil, ErrInvalidWorkoutSetInput
	}
	set, err := ws.repo.GetWorkoutSetsByWorkoutExerciseID(ctx, exerciseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWorkoutSetNotFound
	}
	if err != nil {
		return nil, err
	}
	return set, nil
}

// UpdateWorkoutSet updates an existing workout set by its exercise ID.
func (ws *WorkoutService) UpdateWorkoutSet(ctx context.Context, setID uuid.UUID, set *model.WorkoutSet) error {
	if setID == uuid.Nil || set.Completed == nil || set.SetNumber <= 0 {
		return ErrInvalidWorkoutSetInput
	}
	err := ws.repo.UpdateWorkoutSet(ctx, setID, set.SetNumber, set)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrWorkoutSetNotFound
	}
	if err != nil {
		return err
	}
	return nil
}
