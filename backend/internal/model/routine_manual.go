package model

// ManualRoutineToSave contains the routine data persisted from a manual create or edit.
type ManualRoutineToSave struct {
	UserID      string
	Name        string
	Description string
	RoutineType string
	Exercises   []AIRoutineExerciseToSave
}
