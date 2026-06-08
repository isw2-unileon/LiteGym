package service

import (
	"context"
	"errors"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
)

type saveGeneratedRoutineCommand struct {
	service   *RoutineAIService
	userID    string
	generated model.AIRoutineJSON
}

func (c saveGeneratedRoutineCommand) Execute(ctx context.Context) (model.AIRoutineGenerateResponse, error) {
	routineID, err := c.service.saveGeneratedRoutine(ctx, c.userID, c.generated)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	return model.AIRoutineGenerateResponse{
		RoutineJSON: c.generated,
		RoutineID:   routineID,
	}, nil
}

type saveUpgradedRoutineAsNewCommand struct {
	service     *RoutineAIService
	userID      string
	baseRoutine string
	generated   model.AIRoutineJSON
}

func (c saveUpgradedRoutineAsNewCommand) Execute(ctx context.Context) (model.AIRoutineGenerateResponse, error) {
	if _, err := c.service.repo.GetByID(ctx, c.userID, c.baseRoutine); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AIRoutineGenerateResponse{}, ErrRoutineNotFound
		}
		return model.AIRoutineGenerateResponse{}, err
	}

	routineID, err := c.service.saveGeneratedRoutine(ctx, c.userID, c.generated)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	return model.AIRoutineGenerateResponse{
		RoutineJSON: c.generated,
		RoutineID:   routineID,
	}, nil
}

type overwriteRoutineWithAICommand struct {
	service   *RoutineAIService
	userID    string
	routineID string
	generated model.AIRoutineJSON
}

func (c overwriteRoutineWithAICommand) Execute(ctx context.Context) (model.AIRoutineGenerateResponse, error) {
	if _, err := c.service.repo.GetByID(ctx, c.userID, c.routineID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AIRoutineGenerateResponse{}, ErrRoutineNotFound
		}
		return model.AIRoutineGenerateResponse{}, err
	}

	routineToSave, err := c.service.buildGeneratedRoutineToSave(ctx, c.userID, c.generated)
	if err != nil {
		return model.AIRoutineGenerateResponse{}, err
	}

	if err := c.service.repo.OverwriteGeneratedAIRoutine(ctx, c.routineID, c.userID, routineToSave); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AIRoutineGenerateResponse{}, ErrRoutineNotFound
		}
		return model.AIRoutineGenerateResponse{}, err
	}

	return model.AIRoutineGenerateResponse{
		RoutineJSON: c.generated,
		RoutineID:   c.routineID,
	}, nil
}
