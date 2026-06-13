package service

import (
	"context"
	"strings"
	"time"
	_ "time/tzdata" // embed the IANA timezone database so LoadLocation works on any host

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
)

const defaultWeeklyWorkoutGoal = 2

// madridLocation is the canonical timezone used to present calendar days and
// workout timestamps, regardless of the server's own timezone.
var madridLocation = loadMadridLocation()

func loadMadridLocation() *time.Location {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		return time.FixedZone("CET", 60*60)
	}
	return loc
}

// OverviewService provides the aggregated data shown on the dashboard.
type OverviewService struct {
	routineRepo         repository.RoutineRepository
	overviewWorkoutRepo repository.OverviewWorkoutRepository
	bodyMetricRepo      repository.BodyMetricRepository
}

// NewOverviewService creates a new OverviewService.
func NewOverviewService(
	routineRepo repository.RoutineRepository,
	overviewWorkoutRepo repository.OverviewWorkoutRepository,
	bodyMetricRepo repository.BodyMetricRepository,
) *OverviewService {
	return &OverviewService{
		routineRepo:         routineRepo,
		overviewWorkoutRepo: overviewWorkoutRepo,
		bodyMetricRepo:      bodyMetricRepo,
	}
}

// GetOverview returns the data required to render the authenticated dashboard.
func (s *OverviewService) GetOverview(ctx context.Context, userID string, referenceDate time.Time) (model.Overview, error) {
	return s.GetOverviewForMonth(ctx, userID, referenceDate, referenceDate)
}

// GetOverviewForMonth returns dashboard data for one calendar month while keeping rolling stats anchored to today.
func (s *OverviewService) GetOverviewForMonth(ctx context.Context, userID string, calendarDate, statsReferenceDate time.Time) (model.Overview, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return model.Overview{}, ErrInvalidUserInput
	}

	calendarDate = nowIfZero(calendarDate)
	statsReferenceDate = nowIfZero(statsReferenceDate)

	// Anchor every calendar/day computation to Madrid time so day boundaries match
	// what the user sees, independent of the server timezone.
	calendarDate = calendarDate.In(madridLocation)
	statsReferenceDate = statsReferenceDate.In(madridLocation)

	monthStart := time.Date(calendarDate.Year(), calendarDate.Month(), 1, 0, 0, 0, 0, madridLocation)
	monthEnd := monthStart.AddDate(0, 1, 0)
	statsMonthStart := time.Date(statsReferenceDate.Year(), statsReferenceDate.Month(), 1, 0, 0, 0, 0, madridLocation)
	statsYearStart := statsReferenceDate.AddDate(-1, 0, 0)
	streakRangeStart := statsReferenceDate.AddDate(0, 0, -90)

	recentRoutines, err := s.routineRepo.ListRecentByUser(ctx, userID, 4)
	if err != nil {
		return model.Overview{}, err
	}

	recentWorkouts, err := s.overviewWorkoutRepo.ListRecentByUser(ctx, userID, 4)
	if err != nil {
		return model.Overview{}, err
	}

	monthWorkoutDates, err := s.overviewWorkoutRepo.ListTrainingDatesInRange(ctx, userID, monthStart.UTC(), monthEnd.UTC())
	if err != nil {
		return model.Overview{}, err
	}

	todayStart := truncateDate(statsReferenceDate)
	plannedRangeStart := todayStart
	if plannedRangeStart.Before(monthStart) {
		plannedRangeStart = monthStart
	}
	plannedWorkoutDates, err := s.overviewWorkoutRepo.ListPlannedDatesInRange(ctx, userID, plannedRangeStart.UTC(), monthEnd.UTC())
	if err != nil {
		return model.Overview{}, err
	}

	calendarWorkouts, err := s.overviewWorkoutRepo.ListCalendarWorkoutsByUser(ctx, userID, monthStart.UTC(), monthEnd.UTC())
	if err != nil {
		return model.Overview{}, err
	}

	streakWorkoutDates, err := s.overviewWorkoutRepo.ListTrainingDatesInRange(ctx, userID, streakRangeStart.UTC(), statsReferenceDate.AddDate(0, 0, 1).UTC())
	if err != nil {
		return model.Overview{}, err
	}

	bodyMetrics, err := s.bodyMetricRepo.ListRecentByUser(ctx, userID, 2)
	if err != nil {
		return model.Overview{}, err
	}

	yearDistribution, yearExerciseCount, err := s.overviewWorkoutRepo.ListMuscleDistributionByUser(ctx, userID, statsYearStart.UTC(), statsReferenceDate.AddDate(0, 0, 1).UTC())
	if err != nil {
		return model.Overview{}, err
	}

	monthDistribution, monthExerciseCount, err := s.overviewWorkoutRepo.ListMuscleDistributionByUser(ctx, userID, statsMonthStart.UTC(), statsMonthStart.AddDate(0, 1, 0).UTC())
	if err != nil {
		return model.Overview{}, err
	}

	// Present every returned timestamp in Madrid time so the frontend renders the
	// same day the backend used to build the calendar.
	calendarWorkouts = workoutsToMadrid(calendarWorkouts)
	recentWorkouts = recentSummariesToMadrid(recentWorkouts)
	recentRoutines = routinesToMadrid(recentRoutines)
	progress := buildMadridProgressSummary(bodyMetrics)

	return model.Overview{
		Calendar: model.OverviewCalendar{
			Month:            monthStart.Format("2006-01"),
			TrainedDays:      buildDateStringsAscending(monthWorkoutDates),
			PlannedDays:      buildDateStringsAscending(plannedWorkoutDates),
			CalendarWorkouts: calendarWorkouts,
			SessionsCount:    len(monthWorkoutDates),
			CurrentStreak:    calculateWeeklyGoalStreak(statsReferenceDate, streakWorkoutDates, defaultWeeklyWorkoutGoal),
			WeeklyGoal:       defaultWeeklyWorkoutGoal,
			NextObjective:    buildNextObjective(len(monthWorkoutDates)),
		},
		RecentRoutines: recentRoutines,
		RecentWorkouts: recentWorkouts,
		Progress:       progress,
		MuscleDistribution: model.OverviewMuscleDistribution{
			Year:               yearDistribution,
			Month:              monthDistribution,
			YearExerciseCount:  yearExerciseCount,
			MonthExerciseCount: monthExerciseCount,
		},
	}, nil
}

func workoutsToMadrid(workouts []model.OverviewCalendarWorkout) []model.OverviewCalendarWorkout {
	for index := range workouts {
		if workouts[index].PerformedAt != nil {
			performedAt := workouts[index].PerformedAt.In(madridLocation)
			workouts[index].PerformedAt = &performedAt
		}
		if workouts[index].PlannedAt != nil {
			plannedAt := workouts[index].PlannedAt.In(madridLocation)
			workouts[index].PlannedAt = &plannedAt
		}
	}
	return workouts
}

func recentSummariesToMadrid(workouts []model.OverviewWorkoutSummary) []model.OverviewWorkoutSummary {
	for index := range workouts {
		workouts[index].PerformedAt = workouts[index].PerformedAt.In(madridLocation)
	}
	return workouts
}

func routinesToMadrid(routines []model.OverviewRoutineSummary) []model.OverviewRoutineSummary {
	for index := range routines {
		routines[index].UpdatedAt = routines[index].UpdatedAt.In(madridLocation)
	}
	return routines
}

func nowIfZero(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}

func buildMadridProgressSummary(entries []model.OverviewBodyMetricEntry) model.OverviewProgressSummary {
	progress := buildProgressSummary(entries)
	if progress.LastRecordedAt != nil {
		recordedAt := progress.LastRecordedAt.In(madridLocation)
		progress.LastRecordedAt = &recordedAt
	}
	return progress
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

func calculateWeeklyGoalStreak(referenceDate time.Time, dates []time.Time, weeklyGoal int) int {
	if len(dates) == 0 || weeklyGoal <= 0 {
		return 0
	}

	trainingWeeks := make(map[string]int)
	for _, trainedDate := range dates {
		weekStart := startOfWeek(trainedDate)
		trainingWeeks[weekStart.Format("2006-01-02")]++
	}

	currentWeekStart := startOfWeek(referenceDate)
	if trainingWeeks[currentWeekStart.Format("2006-01-02")] < weeklyGoal {
		currentWeekStart = currentWeekStart.AddDate(0, 0, -7)
	}

	streak := 0
	for weekStart := currentWeekStart; ; weekStart = weekStart.AddDate(0, 0, -7) {
		if trainingWeeks[weekStart.Format("2006-01-02")] < weeklyGoal {
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

func startOfWeek(value time.Time) time.Time {
	date := truncateDate(value)
	offset := (int(date.Weekday()) + 6) % 7
	return date.AddDate(0, 0, -offset)
}
