package service

import "errors"

// Routine domain errors returned by the routine service.
var (
	ErrInvalidRoutineInput = errors.New("invalid routine input")
	ErrRoutineNotFound     = errors.New("routine not found")
	ErrRoutineNameRequired = errors.New("routine name is required")
	ErrRoutineLimitReached = errors.New("routine limit reached")
)
