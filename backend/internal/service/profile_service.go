package service

import (
	"context"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
)

// ProfileService provides business logic for user profiles.
type ProfileService struct {
	repo repository.ProfileRepository
}

// NewProfileService creates a new ProfileService.
func NewProfileService(repo repository.ProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
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
