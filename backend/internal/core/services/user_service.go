package services

import (
	"context"

	"github.com/isw2-unileon/Grupo-16/backend/internal/core/domain"
	"github.com/isw2-unileon/Grupo-16/backend/internal/core/ports"
)

type UserService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) Create(ctx context.Context, user *domain.User) error {
	return s.repo.Create(ctx, user)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}
