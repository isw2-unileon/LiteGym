package service

import (
	"sort"
	"strings"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
)

// Exercise type values accepted by the exercise domain.
const (
	ExerciseTypeStrength    = "strength"
	ExerciseTypeCardio      = "cardio"
	ExerciseTypeMobility    = "mobility"
	ExerciseTypeBalance     = "balance"
	ExerciseTypePlyometric  = "plyometric"
	ExerciseTypeFlexibility = "flexibility"
	ExerciseTypeEndurance   = "endurance"
	ExerciseTypeWarmup      = "warmup"
	ExerciseTypeCooldown    = "cooldown"
	ExerciseTypeTechnique   = "technique"
	ExerciseTypeRehab       = "rehab"

	maxExerciseNameLength        = 100
	maxExerciseDescriptionLength = 500
)

var validExerciseTypes = map[string]struct{}{
	ExerciseTypeStrength:    {},
	ExerciseTypeCardio:      {},
	ExerciseTypeMobility:    {},
	ExerciseTypeBalance:     {},
	ExerciseTypePlyometric:  {},
	ExerciseTypeFlexibility: {},
	ExerciseTypeEndurance:   {},
	ExerciseTypeWarmup:      {},
	ExerciseTypeCooldown:    {},
	ExerciseTypeTechnique:   {},
	ExerciseTypeRehab:       {},
}

var validMuscleGroups = map[string]struct{}{
	"chest":       {},
	"back":        {},
	"legs":        {},
	"shoulders":   {},
	"biceps":      {},
	"triceps":     {},
	"core":        {},
	"glutes":      {},
	"forearms":    {},
	"calves":      {},
	"hamstrings":  {},
	"quadriceps":  {},
	"lats":        {},
	"traps":       {},
	"rear_delts":  {},
	"front_delts": {},
	"side_delts":  {},
	"full_body":   {},
	"cardio":      {},
}

var exerciseTypeLabels = map[string]string{
	ExerciseTypeStrength:    "Fuerza",
	ExerciseTypeCardio:      "Cardio",
	ExerciseTypeMobility:    "Movilidad",
	ExerciseTypeBalance:     "Equilibrio",
	ExerciseTypePlyometric:  "Pliometría",
	ExerciseTypeFlexibility: "Flexibilidad",
	ExerciseTypeEndurance:   "Resistencia",
	ExerciseTypeWarmup:      "Calentamiento",
	ExerciseTypeCooldown:    "Vuelta a la calma",
	ExerciseTypeTechnique:   "Técnica",
	ExerciseTypeRehab:       "Rehabilitación",
}

var muscleGroupLabels = map[string]string{
	"chest":       "Pecho",
	"back":        "Espalda",
	"legs":        "Piernas",
	"shoulders":   "Hombros",
	"biceps":      "Bíceps",
	"triceps":     "Tríceps",
	"core":        "Core",
	"glutes":      "Glúteos",
	"forearms":    "Antebrazos",
	"calves":      "Gemelos",
	"hamstrings":  "Isquiotibiales",
	"quadriceps":  "Cuádriceps",
	"lats":        "Dorsales",
	"traps":       "Trapecios",
	"rear_delts":  "Deltoides posteriores",
	"front_delts": "Deltoides anteriores",
	"side_delts":  "Deltoides laterales",
	"full_body":   "Cuerpo completo",
	"cardio":      "Cardio",
}

func normalizeDomainValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
	value = strings.Join(strings.Fields(value), "_")
	return value
}

func normalizeName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Join(strings.Fields(value), " ")
	return value
}

func isValidExerciseType(value string) bool {
	_, ok := validExerciseTypes[value]
	return ok
}

func isValidMuscleGroup(value string) bool {
	_, ok := validMuscleGroups[value]
	return ok
}

func exerciseMetadata() model.ExerciseMetadataResponse {
	return model.ExerciseMetadataResponse{
		ExerciseTypes: selectOptionsFromCatalog(validExerciseTypes, exerciseTypeLabels),
		MuscleGroups:  selectOptionsFromCatalog(validMuscleGroups, muscleGroupLabels),
	}
}

func selectOptionsFromCatalog(catalog map[string]struct{}, labels map[string]string) []model.SelectOption {
	options := make([]model.SelectOption, 0, len(catalog))

	for value := range catalog {
		options = append(options, model.SelectOption{
			Value: value,
			Label: labelFromDomainValue(value, labels),
		})
	}

	sort.Slice(options, func(i, j int) bool {
		if options[i].Label == options[j].Label {
			return options[i].Value < options[j].Value
		}
		return options[i].Label < options[j].Label
	})

	return options
}

func labelFromDomainValue(value string, labels map[string]string) string {
	if label, ok := labels[value]; ok {
		return label
	}

	words := strings.Split(value, "_")
	for index, word := range words {
		if word == "" {
			continue
		}
		words[index] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}
