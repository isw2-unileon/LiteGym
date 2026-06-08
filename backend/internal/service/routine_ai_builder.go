package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
)

type routineJSONBuilder struct {
	routine model.AIRoutineJSON
}

func newRoutineJSONBuilder(routine model.AIRoutineJSON) *routineJSONBuilder {
	return &routineJSONBuilder{routine: routine}
}

func (b *routineJSONBuilder) withGenerationDefaults(req model.AIRoutineGenerationRequest, now time.Time) *routineJSONBuilder {
	if strings.TrimSpace(b.routine.Objective) == "" {
		b.routine.Objective = req.Objective
	}
	if b.routine.DurationMinutes <= 0 {
		b.routine.DurationMinutes = req.DurationMinutes
	}
	if len(b.routine.TargetMuscles) == 0 {
		b.routine.TargetMuscles = normalizeTextList(req.TargetMuscleGroups)
	}
	normalizeGeneratedRoutineMandatoryFlags(&b.routine, req.MandatoryExercises)
	b.routine.GeneratedAt = now
	if strings.TrimSpace(b.routine.GenerationSource) == "" {
		b.routine.GenerationSource = "gemini"
	}
	b.routine.MandatoryCount = countMandatoryExercises(b.routine.Exercises)
	return b
}

func (b *routineJSONBuilder) withUpgradeDefaults(routine model.Routine, now time.Time) *routineJSONBuilder {
	if strings.TrimSpace(b.routine.Objective) == "" {
		b.routine.Objective = fmt.Sprintf("Upgrade of %s", strings.TrimSpace(routine.Name))
	}
	if b.routine.DurationMinutes <= 0 {
		b.routine.DurationMinutes = estimateRoutineDurationMinutes(routine)
	}
	if len(b.routine.TargetMuscles) == 0 {
		b.routine.TargetMuscles = extractRoutineTargetMuscles(&routine)
	}
	b.routine.GeneratedAt = now
	if strings.TrimSpace(b.routine.GenerationSource) == "" {
		b.routine.GenerationSource = "gemini"
	}
	b.routine.MandatoryCount = countMandatoryExercises(b.routine.Exercises)
	return b
}

func (b *routineJSONBuilder) build() model.AIRoutineJSON {
	return b.routine
}
