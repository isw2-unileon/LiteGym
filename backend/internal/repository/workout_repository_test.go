package repository

import "testing"

func TestCalculateDynamicTargetWeightIncreasesAfterStrongPerformance(t *testing.T) {
	repsMin := 8
	repsMax := 10
	targetRir := 2
	baseWeight := 80.0
	reps := 10
	rir := 3

	got := calculateDynamicTargetWeight(
		routineExercisePrescription{MuscleGroup: "chest", ExerciseType: "strength"},
		routineSetPrescription{
			SetNumber:      1,
			TargetRepsMin:  &repsMin,
			TargetRepsMax:  &repsMax,
			TargetWeightKg: &baseWeight,
			TargetRir:      &targetRir,
		},
		[]recentWorkoutSetPerformance{
			{SetNumber: 1, Reps: &reps, WeightKg: &baseWeight, Rir: &rir},
			{SetNumber: 2, Reps: &reps, WeightKg: &baseWeight, Rir: &rir},
		},
	)

	if got == nil || *got != 82.5 {
		t.Fatalf("expected target weight 82.5, got %#v", got)
	}
}

func TestCalculateDynamicTargetWeightReducesAfterMissedReps(t *testing.T) {
	repsMin := 8
	repsMax := 10
	baseWeight := 80.0
	reps := 6
	rir := 1

	got := calculateDynamicTargetWeight(
		routineExercisePrescription{MuscleGroup: "legs", ExerciseType: "strength"},
		routineSetPrescription{
			SetNumber:      1,
			TargetRepsMin:  &repsMin,
			TargetRepsMax:  &repsMax,
			TargetWeightKg: &baseWeight,
		},
		[]recentWorkoutSetPerformance{
			{SetNumber: 1, Reps: &reps, WeightKg: &baseWeight, Rir: &rir},
		},
	)

	if got == nil || *got != 76 {
		t.Fatalf("expected target weight 76, got %#v", got)
	}
}

func TestCalculateDynamicTargetWeightKeepsModeratePerformance(t *testing.T) {
	repsMin := 8
	repsMax := 10
	baseWeight := 80.0
	reps := 9
	rir := 1

	got := calculateDynamicTargetWeight(
		routineExercisePrescription{MuscleGroup: "chest", ExerciseType: "strength"},
		routineSetPrescription{
			SetNumber:      1,
			TargetRepsMin:  &repsMin,
			TargetRepsMax:  &repsMax,
			TargetWeightKg: &baseWeight,
		},
		[]recentWorkoutSetPerformance{
			{SetNumber: 1, Reps: &reps, WeightKg: &baseWeight, Rir: &rir},
		},
	)

	if got == nil || *got != 80 {
		t.Fatalf("expected target weight 80, got %#v", got)
	}
}

func TestCalculateDynamicTargetWeightUsesHistoricalWeightWithoutBaseWeight(t *testing.T) {
	repsMin := 8
	repsMax := 10
	previousWeight := 72.5
	reps := 10
	rir := 3

	got := calculateDynamicTargetWeight(
		routineExercisePrescription{MuscleGroup: "biceps", ExerciseType: "hypertrophy"},
		routineSetPrescription{
			SetNumber:     1,
			TargetRepsMin: &repsMin,
			TargetRepsMax: &repsMax,
		},
		[]recentWorkoutSetPerformance{
			{SetNumber: 1, Reps: &reps, WeightKg: &previousWeight, Rir: &rir},
		},
	)

	if got == nil || *got != 74 {
		t.Fatalf("expected target weight 74, got %#v", got)
	}
}

func TestCalculateDynamicTargetWeightDoesNotForceWeightProgressionForBodyweight(t *testing.T) {
	repsMin := 10
	repsMax := 15
	baseWeight := 20.0
	reps := 15
	rir := 4

	got := calculateDynamicTargetWeight(
		routineExercisePrescription{MuscleGroup: "triceps", ExerciseType: "bodyweight"},
		routineSetPrescription{
			SetNumber:      1,
			TargetRepsMin:  &repsMin,
			TargetRepsMax:  &repsMax,
			TargetWeightKg: &baseWeight,
		},
		[]recentWorkoutSetPerformance{
			{SetNumber: 1, Reps: &reps, WeightKg: &baseWeight, Rir: &rir},
		},
	)

	if got == nil || *got != 20 {
		t.Fatalf("expected bodyweight target to stay at 20, got %#v", got)
	}
}
