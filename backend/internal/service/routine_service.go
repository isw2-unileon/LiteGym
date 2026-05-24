package service

import (
	"context"
	"strings"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
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
