package service

import (
	"context"
	"errors"
	"strings"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/jackc/pgx/v5"
)

// RoutineService provides business logic for routines.
type RoutineService struct {
	repo repository.RoutineRepository
}

// NewRoutineService creates a new RoutineService.
func NewRoutineService(repo repository.RoutineRepository) *RoutineService {
	return &RoutineService{repo: repo}
}

// ListByUser returns all routines owned by a user.
func (s *RoutineService) ListByUser(ctx context.Context, userID string) ([]model.OverviewRoutineSummary, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidUserInput
	}

	return s.repo.ListByUser(ctx, userID)
}

// GetByID returns one routine owned by a user with its exercises and planned sets.
func (s *RoutineService) GetByID(ctx context.Context, userID, routineID string) (*model.RoutineDetail, error) {
	userID = strings.TrimSpace(userID)
	routineID = strings.TrimSpace(routineID)
	if userID == "" || routineID == "" {
		return nil, ErrInvalidRoutineInput
	}

	routine, err := s.repo.GetByID(ctx, userID, routineID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoutineNotFound
		}
		return nil, err
	}

	return routine, nil
}
