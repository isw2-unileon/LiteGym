package service

import "time"

func calculateCurrentStreak(referenceDate time.Time, workoutDates []time.Time) int {
	if len(workoutDates) == 0 {
		return 0
	}

	trained := make(map[string]struct{}, len(workoutDates))
	for _, value := range workoutDates {
		trained[truncateDate(value).Format("2006-01-02")] = struct{}{}
	}

	current := truncateDate(referenceDate)
	if _, ok := trained[current.Format("2006-01-02")]; !ok {
		current = current.AddDate(0, 0, -1)
	}

	streak := 0
	for {
		if _, ok := trained[current.Format("2006-01-02")]; !ok {
			break
		}
		streak++
		current = current.AddDate(0, 0, -1)
	}

	return streak
}
