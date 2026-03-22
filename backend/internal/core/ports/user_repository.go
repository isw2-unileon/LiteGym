package ports

import (
	"context"

	"github.com/isw2-unileon/Grupo-16/backend/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}
