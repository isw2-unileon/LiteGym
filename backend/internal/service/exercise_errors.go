package service

import "errors"

// Exercise domain errors returned by the exercise service.
var (
	ErrInvalidExerciseInput   = errors.New("invalid exercise input")
	ErrExerciseNotFound       = errors.New("exercise not found")
	ErrExerciseNameRequired   = errors.New("exercise name is required")
	ErrExerciseNameTooLong    = errors.New("exercise name is too long")
	ErrExerciseTypeRequired   = errors.New("exercise type is required")
	ErrInvalidExerciseType    = errors.New("invalid exercise type")
	ErrMuscleGroupRequired    = errors.New("muscle group is required")
	ErrInvalidMuscleGroup     = errors.New("invalid muscle group")
	ErrDescriptionTooLong     = errors.New("description is too long")
	ErrSecondaryEqualsPrimary = errors.New("secondary muscle group cannot equal primary muscle group")
	ErrExerciseNameTaken      = errors.New("an exercise with this name already exists")
)
