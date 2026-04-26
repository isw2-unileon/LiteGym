package service

import (
	"sort"
	"strings"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
)

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
		ExerciseTypes: selectOptionsFromCatalog(validExerciseTypes),
		MuscleGroups:  selectOptionsFromCatalog(validMuscleGroups),
	}
}

func selectOptionsFromCatalog(catalog map[string]struct{}) []model.SelectOption {
	options := make([]model.SelectOption, 0, len(catalog))

	for value := range catalog {
		options = append(options, model.SelectOption{
			Value: value,
			Label: labelFromDomainValue(value),
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

func labelFromDomainValue(value string) string {
	words := strings.Split(value, "_")
	for index, word := range words {
		if word == "" {
			continue
		}
		words[index] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}
