package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/jackc/pgx/v5"
)

// maxRoutinesPerUser caps how many routines a single user can own.
const maxRoutinesPerUser = 50

const defaultRoutineType = "Sin clasificar"

var validRoutineTypes = map[string]struct{}{
	"Fuerza":           {},
	"Movilidad":        {},
	"Resistencia":      {},
	defaultRoutineType: {},
}

// ManualRoutineSetInput defines the target metrics for one set in a manual routine.
type ManualRoutineSetInput struct {
	SetNumber             int      `json:"set_number"`
	TargetRepsMin         *int     `json:"target_reps_min"`
	TargetRepsMax         *int     `json:"target_reps_max"`
	TargetRepsText        string   `json:"target_reps_text"`
	TargetWeightKg        *float64 `json:"target_weight_kg"`
	TargetDurationSeconds *int     `json:"target_duration_seconds"`
	TargetDistanceKm      *float64 `json:"target_distance_km"`
	TargetRir             *int     `json:"target_rir"`
	RestSeconds           *int     `json:"rest_seconds"`
	Notes                 string   `json:"notes"`
}

// ManualRoutineExerciseInput defines one exercise and its sets for a manual routine.
type ManualRoutineExerciseInput struct {
	ExerciseID string                  `json:"exercise_id"`
	Sets       []ManualRoutineSetInput `json:"sets"`
}

// ManualRoutineInput holds the data accepted when creating or editing a manual routine.
type ManualRoutineInput struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	RoutineType string                       `json:"routine_type"`
	Exercises   []ManualRoutineExerciseInput `json:"exercises"`
}

// ManualRoutineService provides business logic for manually managed routines.
type ManualRoutineService struct {
	repo repository.ManualRoutineRepository
}

// NewManualRoutineService creates a new ManualRoutineService.
func NewManualRoutineService(repo repository.ManualRoutineRepository) *ManualRoutineService {
	return &ManualRoutineService{repo: repo}
}

// Create persists a new manual routine for the user after validating the input and enforcing the routine limit.
func (s *ManualRoutineService) Create(ctx context.Context, userID string, input ManualRoutineInput) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", ErrInvalidRoutineInput
	}

	routine, err := s.buildRoutineToSave(userID, input)
	if err != nil {
		return "", err
	}

	if err := s.ensureUnderLimit(ctx, userID); err != nil {
		return "", err
	}

	return s.repo.CreateManualRoutine(ctx, routine)
}

// Update replaces the contents of an existing manual routine owned by the user.
func (s *ManualRoutineService) Update(ctx context.Context, userID, routineID string, input ManualRoutineInput) error {
	userID = strings.TrimSpace(userID)
	routineID = strings.TrimSpace(routineID)
	if userID == "" || routineID == "" {
		return ErrInvalidRoutineInput
	}

	routine, err := s.buildRoutineToSave(userID, input)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateManualRoutine(ctx, routineID, userID, routine); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRoutineNotFound
		}
		return err
	}
	return nil
}

// Delete removes a routine owned by the user.
func (s *ManualRoutineService) Delete(ctx context.Context, userID, routineID string) error {
	userID = strings.TrimSpace(userID)
	routineID = strings.TrimSpace(routineID)
	if userID == "" || routineID == "" {
		return ErrInvalidRoutineInput
	}

	if err := s.repo.DeleteRoutine(ctx, routineID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRoutineNotFound
		}
		return err
	}
	return nil
}

// Duplicate copies an existing routine into a new one owned by the user, enforcing the routine limit.
func (s *ManualRoutineService) Duplicate(ctx context.Context, userID, routineID string) (string, error) {
	userID = strings.TrimSpace(userID)
	routineID = strings.TrimSpace(routineID)
	if userID == "" || routineID == "" {
		return "", ErrInvalidRoutineInput
	}

	if err := s.ensureUnderLimit(ctx, userID); err != nil {
		return "", err
	}

	newID, err := s.repo.DuplicateRoutine(ctx, routineID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrRoutineNotFound
		}
		return "", err
	}
	return newID, nil
}

func (s *ManualRoutineService) buildRoutineToSave(userID string, input ManualRoutineInput) (model.ManualRoutineToSave, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.ManualRoutineToSave{}, ErrRoutineNameRequired
	}

	exercises := make([]model.AIRoutineExerciseToSave, 0, len(input.Exercises))
	for i, ex := range input.Exercises {
		exerciseID := strings.TrimSpace(ex.ExerciseID)
		if _, err := uuid.Parse(exerciseID); err != nil {
			return model.ManualRoutineToSave{}, ErrInvalidRoutineInput
		}

		sets := make([]model.AIRoutineExerciseSetToSave, 0, len(ex.Sets))
		for _, st := range ex.Sets {
			sets = append(sets, model.AIRoutineExerciseSetToSave{
				SetNumber:             st.SetNumber,
				TargetRepsMin:         st.TargetRepsMin,
				TargetRepsMax:         st.TargetRepsMax,
				TargetRepsText:        st.TargetRepsText,
				TargetWeightKg:        st.TargetWeightKg,
				TargetDurationSeconds: st.TargetDurationSeconds,
				TargetDistanceKm:      st.TargetDistanceKm,
				TargetRir:             st.TargetRir,
				RestSeconds:           st.RestSeconds,
				Notes:                 st.Notes,
			})
		}

		exercises = append(exercises, model.AIRoutineExerciseToSave{
			ExerciseID: exerciseID,
			Order:      i + 1,
			Sets:       sets,
		})
	}

	return model.ManualRoutineToSave{
		UserID:      userID,
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		RoutineType: normalizeRoutineType(input.RoutineType),
		Exercises:   exercises,
	}, nil
}

func (s *ManualRoutineService) ensureUnderLimit(ctx context.Context, userID string) error {
	count, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return err
	}
	if count >= maxRoutinesPerUser {
		return ErrRoutineLimitReached
	}
	return nil
}

func normalizeRoutineType(value string) string {
	value = strings.TrimSpace(value)
	if _, ok := validRoutineTypes[value]; !ok {
		return defaultRoutineType
	}
	return value
}
