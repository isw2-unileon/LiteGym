package service

import (
	"context"
	"strings"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
)

// OverviewService provides the aggregated data shown on the dashboard.
type OverviewService struct {
	routineRepo        repository.RoutineRepository
	workoutSessionRepo repository.WorkoutSessionRepository
	bodyMetricRepo     repository.BodyMetricRepository
}

// NewOverviewService creates a new OverviewService.
func NewOverviewService(
	routineRepo repository.RoutineRepository,
	workoutSessionRepo repository.WorkoutSessionRepository,
	bodyMetricRepo repository.BodyMetricRepository,
) *OverviewService {
	return &OverviewService{
		routineRepo:        routineRepo,
		workoutSessionRepo: workoutSessionRepo,
		bodyMetricRepo:     bodyMetricRepo,
	}
}

// GetOverview returns the data required to render the authenticated dashboard.
func (s *OverviewService) GetOverview(ctx context.Context, userID string, referenceDate time.Time) (model.Overview, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return model.Overview{}, ErrInvalidUserInput
	}

	if referenceDate.IsZero() {
		referenceDate = time.Now()
	}

	monthStart := time.Date(referenceDate.Year(), referenceDate.Month(), 1, 0, 0, 0, 0, referenceDate.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	yearStart := referenceDate.AddDate(-1, 0, 0)
	streakRangeStart := referenceDate.AddDate(0, 0, -45)

	recentRoutines, err := s.routineRepo.ListRecentByUser(ctx, userID, 3)
	if err != nil {
		return model.Overview{}, err
	}

	recentWorkouts, err := s.workoutSessionRepo.ListRecentByUser(ctx, userID, 3)
	if err != nil {
		return model.Overview{}, err
	}

	monthWorkoutDates, err := s.workoutSessionRepo.ListTrainingDatesInRange(ctx, userID, monthStart, monthEnd)
	if err != nil {
		return model.Overview{}, err
	}

	streakWorkoutDates, err := s.workoutSessionRepo.ListTrainingDatesInRange(ctx, userID, streakRangeStart, referenceDate.AddDate(0, 0, 1))
	if err != nil {
		return model.Overview{}, err
	}

	bodyMetrics, err := s.bodyMetricRepo.ListRecentByUser(ctx, userID, 2)
	if err != nil {
		return model.Overview{}, err
	}

	yearDistribution, yearExerciseCount, err := s.workoutSessionRepo.ListMuscleDistributionByUser(ctx, userID, yearStart, referenceDate.AddDate(0, 0, 1))
	if err != nil {
		return model.Overview{}, err
	}

	monthDistribution, monthExerciseCount, err := s.workoutSessionRepo.ListMuscleDistributionByUser(ctx, userID, monthStart, monthEnd)
	if err != nil {
		return model.Overview{}, err
	}

	return model.Overview{
		Calendar: model.OverviewCalendar{
			Month:         monthStart.Format("2006-01"),
			TrainedDays:   buildDateStringsAscending(monthWorkoutDates),
			SessionsCount: len(monthWorkoutDates),
			CurrentStreak: calculateCurrentStreak(referenceDate, streakWorkoutDates),
			NextObjective: buildNextObjective(len(monthWorkoutDates)),
		},
		RecentRoutines: recentRoutines,
		RecentWorkouts: recentWorkouts,
		Progress:       buildProgressSummary(bodyMetrics),
		MuscleDistribution: model.OverviewMuscleDistribution{
			Year:               yearDistribution,
			Month:              monthDistribution,
			YearExerciseCount:  yearExerciseCount,
			MonthExerciseCount: monthExerciseCount,
		},
	}, nil
}

func buildProgressSummary(entries []model.OverviewBodyMetricEntry) model.OverviewProgressSummary {
	if len(entries) == 0 {
		return model.OverviewProgressSummary{}
	}

	current := entries[0]
	var previous *model.OverviewBodyMetricEntry
	if len(entries) > 1 {
		previous = &entries[1]
	}

	recordedAt := current.RecordedAt

	return model.OverviewProgressSummary{
		LastRecordedAt:    &recordedAt,
		WeightKg:          buildProgressMetric(current.WeightKg, previousValue(previous, func(entry *model.OverviewBodyMetricEntry) *float64 { return entry.WeightKg })),
		BodyFatPercentage: buildProgressMetric(current.BodyFatPercentage, previousValue(previous, func(entry *model.OverviewBodyMetricEntry) *float64 { return entry.BodyFatPercentage })),
		MuscleMassKg:      buildProgressMetric(current.MuscleMassKg, previousValue(previous, func(entry *model.OverviewBodyMetricEntry) *float64 { return entry.MuscleMassKg })),
	}
}

func buildProgressMetric(current, previous *float64) model.OverviewProgressMetric {
	metric := model.OverviewProgressMetric{
		Current:  current,
		Previous: previous,
	}

	if current != nil && previous != nil {
		delta := *current - *previous
		metric.Delta = &delta
	}

	return metric
}

func previousValue(entry *model.OverviewBodyMetricEntry, selector func(*model.OverviewBodyMetricEntry) *float64) *float64 {
	if entry == nil {
		return nil
	}
	return selector(entry)
}

func buildDateStringsAscending(dates []time.Time) []string {
	if len(dates) == 0 {
		return []string{}
	}

	values := make([]string, 0, len(dates))
	for index := len(dates) - 1; index >= 0; index-- {
		values = append(values, dates[index].Format("2006-01-02"))
	}
	return values
}

func calculateCurrentStreak(referenceDate time.Time, dates []time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	trainingDays := make(map[string]struct{}, len(dates))
	for _, trainedDate := range dates {
		trainingDays[trainedDate.Format("2006-01-02")] = struct{}{}
	}

	startDate := truncateDate(referenceDate)
	if _, ok := trainingDays[startDate.Format("2006-01-02")]; !ok {
		startDate = startDate.AddDate(0, 0, -1)
		if _, ok := trainingDays[startDate.Format("2006-01-02")]; !ok {
			return 0
		}
	}

	streak := 0
	for current := startDate; ; current = current.AddDate(0, 0, -1) {
		if _, ok := trainingDays[current.Format("2006-01-02")]; !ok {
			break
		}
		streak++
	}

	return streak
}

func buildNextObjective(monthSessions int) string {
	switch {
	case monthSessions >= 12:
		return "Mantener el ritmo del mes"
	case monthSessions >= 8:
		return "Completar 12 sesiones este mes"
	case monthSessions >= 4:
		return "Llegar a 8 sesiones este mes"
	default:
		return "Arrancar con 4 sesiones este mes"
	}
}

func truncateDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
