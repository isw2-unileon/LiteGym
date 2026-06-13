package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ExerciseRepository defines the persistence operations for exercises.
type ExerciseRepository interface {
	GetByID(ctx context.Context, id string) (*model.Exercise, error)
	NameExists(ctx context.Context, name string, ownerUserID *string, excludeID string) (bool, error)
	List(ctx context.Context, filter model.ExerciseFilter) ([]model.Exercise, int, error)
	ListWorkoutSessionsByExercise(ctx context.Context, exerciseID, userID string, limit int) ([]model.ExerciseWorkoutSessionSummary, error)
	GetInsights(ctx context.Context, exerciseID, userID string) (model.ExerciseInsights, error)
	Create(ctx context.Context, exercise *model.Exercise) error
	UpdateExercise(ctx context.Context, exercise *model.Exercise) error
	DeleteExercise(ctx context.Context, id string) error
}

type exerciseRepository struct {
	db *pgxpool.Pool
}

// NewExerciseRepository creates a new ExerciseRepository backed by PostgreSQL.
func NewExerciseRepository(db *pgxpool.Pool) ExerciseRepository {
	return &exerciseRepository{
		db: db,
	}
}

func (r *exerciseRepository) GetByID(ctx context.Context, id string) (*model.Exercise, error) {
	query := `
		SELECT
			e.id::text,
			e.name,
			e.description,
			e.muscle_group,
			COALESCE(string_agg(esmg.muscle_group, ', ' ORDER BY esmg.muscle_group), ''),
			e.exercise_type,
			e.is_official,
			e.owner_user_id::text,
			e.created_at
		FROM exercises e
		LEFT JOIN exercise_secondary_muscle_groups esmg ON esmg.exercise_id = e.id
		WHERE e.id = $1::uuid AND e.deleted_at IS NULL
		GROUP BY e.id, e.name, e.description, e.muscle_group, e.exercise_type, e.is_official, e.owner_user_id, e.created_at
	`

	var exercise model.Exercise

	err := r.db.QueryRow(ctx, query, id).Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.Description,
		&exercise.MuscleGroup,
		&exercise.SecondaryMuscleGroup,
		&exercise.ExerciseType,
		&exercise.IsOfficial,
		&exercise.OwnerUserID,
		&exercise.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &exercise, nil
}

// NameExists reports whether an active exercise already uses the given name
// (case-insensitive) within the scope that must stay unique: any official
// exercise, plus the requesting owner's own exercises. ownerUserID is nil for
// official exercises (which then only clash with other official ones). excludeID
// skips a specific exercise (used when updating).
func (r *exerciseRepository) NameExists(ctx context.Context, name string, ownerUserID *string, excludeID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM exercises e
			WHERE e.deleted_at IS NULL
				AND lower(e.name) = lower($1)
				AND ($2::uuid IS NULL OR e.id <> $2::uuid)
				AND (e.is_official = true OR e.owner_user_id = $3::uuid)
		)
	`

	var exists bool
	if err := r.db.QueryRow(ctx, query, name, nullableUUIDParam(excludeID), ownerUserID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

// nullableUUIDParam returns nil for an empty id so a "$n::uuid IS NULL" guard can
// skip the predicate (e.g. an admin owner filter, or an absent exclude id);
// otherwise it returns the value to bind.
func nullableUUIDParam(id string) *string {
	if strings.TrimSpace(id) == "" {
		return nil
	}
	return &id
}

func (r *exerciseRepository) List(ctx context.Context, filters model.ExerciseFilter) ([]model.Exercise, int, error) {
	total, err := r.countExercises(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	offset := (filters.Page - 1) * filters.Limit

	query := `
		SELECT
			e.id::text,
			e.name,
			e.description,
			e.muscle_group,
			COALESCE(string_agg(esmg.muscle_group, ', ' ORDER BY esmg.muscle_group), ''),
			e.exercise_type,
			e.is_official,
			e.owner_user_id::text,
			u.email,
			e.created_at
		FROM exercises e
		LEFT JOIN exercise_secondary_muscle_groups esmg ON esmg.exercise_id = e.id
		LEFT JOIN users u ON u.id = e.owner_user_id
		WHERE e.deleted_at IS NULL
			AND (
				$1 = ''
				OR e.name ILIKE '%' || $1 || '%'
				OR e.exercise_type ILIKE '%' || $1 || '%'
				OR e.muscle_group ILIKE '%' || $1 || '%'
				OR EXISTS (
					SELECT 1
					FROM exercise_secondary_muscle_groups esmg_search
					WHERE esmg_search.exercise_id = e.id
						AND esmg_search.muscle_group ILIKE '%' || $1 || '%'
				)
			)
			AND ($2 = '' OR e.exercise_type = $2)
			AND (
				$3 = ''
				OR e.muscle_group = $3
				OR EXISTS (
					SELECT 1
					FROM exercise_secondary_muscle_groups esmg_filter
					WHERE esmg_filter.exercise_id = e.id
						AND esmg_filter.muscle_group = $3
				)
			)
			AND ($4::boolean IS NULL OR e.is_official = $4)
			AND ($7::uuid IS NULL OR e.is_official = true OR e.owner_user_id = $7::uuid)
		GROUP BY
			e.id,
			e.name,
			e.description,
			e.muscle_group,
			e.exercise_type,
			e.is_official,
			e.owner_user_id,
			u.email,
			e.created_at
		ORDER BY e.created_at ASC, e.id::text ASC
		LIMIT $5 OFFSET $6
	`

	rows, err := r.db.Query(
		ctx,
		query,
		filters.Search,
		filters.Type,
		filters.MuscleGroup,
		filters.Official,
		filters.Limit,
		offset,
		nullableUUIDParam(filters.UserID),
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	exercises := make([]model.Exercise, 0)

	for rows.Next() {
		var exercise model.Exercise

		err := rows.Scan(
			&exercise.ID,
			&exercise.Name,
			&exercise.Description,
			&exercise.MuscleGroup,
			&exercise.SecondaryMuscleGroup,
			&exercise.ExerciseType,
			&exercise.IsOfficial,
			&exercise.OwnerUserID,
			&exercise.OwnerEmail,
			&exercise.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		exercises = append(exercises, exercise)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return exercises, total, nil
}

func (r *exerciseRepository) countExercises(ctx context.Context, filters model.ExerciseFilter) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM exercises e
		WHERE e.deleted_at IS NULL
			AND (
				$1 = ''
				OR e.name ILIKE '%' || $1 || '%'
				OR e.exercise_type ILIKE '%' || $1 || '%'
				OR e.muscle_group ILIKE '%' || $1 || '%'
				OR EXISTS (
					SELECT 1
					FROM exercise_secondary_muscle_groups esmg_search
					WHERE esmg_search.exercise_id = e.id
						AND esmg_search.muscle_group ILIKE '%' || $1 || '%'
				)
			)
			AND ($2 = '' OR e.exercise_type = $2)
			AND (
				$3 = ''
				OR e.muscle_group = $3
				OR EXISTS (
					SELECT 1
					FROM exercise_secondary_muscle_groups esmg_filter
					WHERE esmg_filter.exercise_id = e.id
						AND esmg_filter.muscle_group = $3
				)
			)
			AND ($4::boolean IS NULL OR e.is_official = $4)
			AND ($5::uuid IS NULL OR e.is_official = true OR e.owner_user_id = $5::uuid)
	`

	var total int

	err := r.db.QueryRow(
		ctx,
		query,
		filters.Search,
		filters.Type,
		filters.MuscleGroup,
		filters.Official,
		nullableUUIDParam(filters.UserID),
	).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *exerciseRepository) ListWorkoutSessionsByExercise(ctx context.Context, exerciseID, userID string, limit int) ([]model.ExerciseWorkoutSessionSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			ws.id::text,
			COALESCE(ws.name, ''),
			COALESCE(rt.name, ''),
			ws.performed_at,
			COALESCE(ws.duration_minutes, 0),
			we.exercise_order,
			COUNT(wset.id)::int
		FROM workout_exercises we
		INNER JOIN workout_sessions ws ON ws.id = we.workout_session_id
		INNER JOIN exercises e ON e.id = we.exercise_id AND e.deleted_at IS NULL
		LEFT JOIN routines rt ON rt.id = ws.routine_id
		LEFT JOIN workout_sets wset ON wset.workout_exercise_id = we.id
		WHERE we.exercise_id = $1::uuid
			AND ws.user_id = $2::uuid
			AND ws.performed_at IS NOT NULL
		GROUP BY
			ws.id,
			ws.name,
			rt.name,
			ws.performed_at,
			ws.duration_minutes,
			we.exercise_order
		ORDER BY ws.performed_at DESC, ws.id::text DESC
		LIMIT $3
	`, exerciseID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]model.ExerciseWorkoutSessionSummary, 0)
	for rows.Next() {
		var session model.ExerciseWorkoutSessionSummary
		if err := rows.Scan(
			&session.ID,
			&session.Name,
			&session.RoutineName,
			&session.StartedAt,
			&session.DurationMinutes,
			&session.ExerciseOrder,
			&session.SetCount,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

func (r *exerciseRepository) GetInsights(ctx context.Context, exerciseID, userID string) (model.ExerciseInsights, error) { //nolint:gocognit,gocyclo,funlen // Aggregates exercise history into summary, records, progression, and session history from one ordered query.
	rows, err := r.db.Query(ctx, `
		SELECT
			ws.id::text,
			COALESCE(ws.name, ''),
			COALESCE(rt.name, ''),
			ws.performed_at,
			COALESCE(ws.duration_minutes, 0),
			wset.set_number,
			wset.reps,
			wset.weight_kg,
			wset.rir,
			wset.completed
		FROM workout_exercises we
		INNER JOIN workout_sessions ws ON ws.id = we.workout_session_id
		INNER JOIN exercises e ON e.id = we.exercise_id AND e.deleted_at IS NULL
		LEFT JOIN routines rt ON rt.id = ws.routine_id
		INNER JOIN workout_sets wset ON wset.workout_exercise_id = we.id
		WHERE we.exercise_id = $1::uuid
			AND ws.user_id = $2::uuid
			AND ws.performed_at IS NOT NULL
		ORDER BY ws.performed_at ASC, ws.id::text ASC, wset.set_number ASC
	`, exerciseID, userID)
	if err != nil {
		return model.ExerciseInsights{}, err
	}
	defer rows.Close()

	type sessionAggregate struct {
		history   model.ExerciseInsightSessionHistory
		maxWeight *float64
		maxReps   *int
	}

	insights := model.ExerciseInsights{
		PersonalRecords: []model.ExercisePersonalRecord{},
		Progression:     []model.ExerciseProgressPoint{},
		History:         []model.ExerciseInsightSessionHistory{},
	}
	sessions := make(map[string]*sessionAggregate)
	sessionOrder := make([]string, 0)

	for rows.Next() {
		var (
			sessionID       string
			sessionName     string
			routineName     string
			performedAt     time.Time
			durationMinutes int
			setNumber       int
			reps            sql.NullInt64
			weightKg        sql.NullFloat64
			rir             sql.NullInt64
			completed       bool
		)

		if err := rows.Scan(
			&sessionID,
			&sessionName,
			&routineName,
			&performedAt,
			&durationMinutes,
			&setNumber,
			&reps,
			&weightKg,
			&rir,
			&completed,
		); err != nil {
			return model.ExerciseInsights{}, err
		}

		aggregate, ok := sessions[sessionID]
		if !ok {
			aggregate = &sessionAggregate{
				history: model.ExerciseInsightSessionHistory{
					SessionID:       sessionID,
					SessionName:     sessionName,
					RoutineName:     routineName,
					PerformedAt:     performedAt,
					DurationMinutes: durationMinutes,
					Sets:            []model.ExerciseInsightSet{},
				},
			}
			sessions[sessionID] = aggregate
			sessionOrder = append(sessionOrder, sessionID)
		}

		repsValue := intPointerFromNull(reps)
		weightValue := floatPointerFromNull(weightKg)
		rirValue := intPointerFromNull(rir)
		volume := setVolume(weightValue, repsValue)
		set := model.ExerciseInsightSet{
			SessionID:   sessionID,
			SessionName: sessionName,
			RoutineName: routineName,
			PerformedAt: performedAt,
			SetNumber:   setNumber,
			Reps:        repsValue,
			WeightKg:    weightValue,
			VolumeKg:    volume,
			Rir:         rirValue,
			Completed:   completed,
		}

		aggregate.history.Sets = append(aggregate.history.Sets, set)
		aggregate.history.VolumeKg += volume
		insights.Summary.SetCount++
		insights.Summary.TotalVolumeKg += volume

		if weightValue != nil && (aggregate.maxWeight == nil || *weightValue > *aggregate.maxWeight) {
			value := *weightValue
			aggregate.maxWeight = &value
		}

		if repsValue != nil && (aggregate.maxReps == nil || *repsValue > *aggregate.maxReps) {
			value := *repsValue
			aggregate.maxReps = &value
		}

		if weightValue != nil && (insights.Summary.MaxWeightKg == nil || *weightValue > *insights.Summary.MaxWeightKg) {
			value := *weightValue
			insights.Summary.MaxWeightKg = &value
			insights.PersonalRecords = upsertExerciseRecord(insights.PersonalRecords, model.ExercisePersonalRecord{
				Type:        "max_weight",
				Label:       "Peso maximo",
				Value:       value,
				Unit:        "kg",
				PerformedAt: performedAt,
			})
		}

		if repsValue != nil && (insights.Summary.MaxReps == nil || *repsValue > *insights.Summary.MaxReps) {
			value := *repsValue
			insights.Summary.MaxReps = &value
			insights.PersonalRecords = upsertExerciseRecord(insights.PersonalRecords, model.ExercisePersonalRecord{
				Type:        "max_reps",
				Label:       "Maximas repeticiones",
				Value:       float64(value),
				Unit:        "reps",
				PerformedAt: performedAt,
			})
		}

		if insights.BestSet == nil || volume > insights.BestSet.VolumeKg {
			bestSet := set
			insights.BestSet = &bestSet
		}

		if volume > 0 {
			insights.PersonalRecords = upsertExerciseRecord(insights.PersonalRecords, model.ExercisePersonalRecord{
				Type:        "best_volume_set",
				Label:       "Mejor serie por volumen",
				Value:       volume,
				Unit:        "kg",
				PerformedAt: performedAt,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return model.ExerciseInsights{}, err
	}

	insights.Summary.SessionCount = len(sessionOrder)
	insights.Summary.Trend = "empty"

	for _, sessionID := range sessionOrder {
		aggregate := sessions[sessionID]
		history := aggregate.history
		insights.History = append([]model.ExerciseInsightSessionHistory{history}, insights.History...)

		point := model.ExerciseProgressPoint{
			SessionID:   history.SessionID,
			Date:        history.PerformedAt,
			MaxWeightKg: aggregate.maxWeight,
			MaxReps:     aggregate.maxReps,
			VolumeKg:    history.VolumeKg,
			SetCount:    len(history.Sets),
		}
		insights.Progression = append(insights.Progression, point)
	}

	if len(insights.Progression) > 0 {
		first := insights.Progression[0].Date
		last := insights.Progression[len(insights.Progression)-1].Date
		insights.Summary.FirstPerformedAt = &first
		insights.Summary.LastPerformedAt = &last
		insights.Summary.Trend = calculateExerciseTrend(insights.Progression)

		if len(insights.Progression) > 1 {
			days := last.Sub(first).Hours() / 24
			average := days / float64(len(insights.Progression)-1)
			insights.Summary.AverageDaysBetween = &average
		}
	}

	return insights, nil
}

func (r *exerciseRepository) Create(ctx context.Context, exercise *model.Exercise) error {
	query := `
		INSERT INTO exercises(
			name,
			description,
			muscle_group,
			exercise_type,
			is_official,
			owner_user_id
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text, created_at
	`

	if err := r.db.QueryRow(
		ctx,
		query,
		exercise.Name,
		exercise.Description,
		exercise.MuscleGroup,
		exercise.ExerciseType,
		exercise.IsOfficial,
		exercise.OwnerUserID,
	).Scan(&exercise.ID, &exercise.CreatedAt); err != nil {
		return err
	}

	if exercise.SecondaryMuscleGroup == "" {
		return nil
	}

	secondaryMuscleGroups := strings.Split(exercise.SecondaryMuscleGroup, ",")
	normalized := make([]string, 0, len(secondaryMuscleGroups))

	for _, group := range secondaryMuscleGroups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}

		_, err := r.db.Exec(
			ctx,
			`
				INSERT INTO exercise_secondary_muscle_groups (exercise_id, muscle_group)
				VALUES ($1::uuid, $2)
				ON CONFLICT (exercise_id, muscle_group) DO NOTHING
			`,
			exercise.ID,
			group,
		)
		if err != nil {
			return err
		}

		normalized = append(normalized, group)
	}

	exercise.SecondaryMuscleGroup = strings.Join(normalized, ", ")
	return nil

}

// UpdateExercise updates the contents of an existing exercise.
func (r *exerciseRepository) UpdateExercise(ctx context.Context, exercise *model.Exercise) error {
	// 1. Define the query that updates the exercise core fields.
	query := `
        UPDATE exercises
        SET
            name = $1,
            description = $2,
            muscle_group = $3,
            exercise_type = $4,
            is_official = $5,
            owner_user_id = $6::uuid
        WHERE id = $7::uuid AND deleted_at IS NULL
    `

	// 2. Execute the update.
	result, err := r.db.Exec(
		ctx,
		query,
		exercise.Name,
		exercise.Description,
		exercise.MuscleGroup,
		exercise.ExerciseType,
		exercise.IsOfficial,
		exercise.OwnerUserID,
		exercise.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	// 3. Handle secondary muscles.

	// Delete current secondary muscles.
	_, err = r.db.Exec(ctx, "DELETE FROM exercise_secondary_muscle_groups WHERE exercise_id = $1::uuid", exercise.ID)
	if err != nil {
		return err
	}

	// Insert new secondary muscles.
	if exercise.SecondaryMuscleGroup != "" {
		groups := strings.Split(exercise.SecondaryMuscleGroup, ",")
		for _, g := range groups {
			g = strings.TrimSpace(g)
			if g == "" {
				continue
			}
			_, err = r.db.Exec(ctx,
				"INSERT INTO exercise_secondary_muscle_groups (exercise_id, muscle_group) VALUES ($1::uuid, $2)",
				exercise.ID, g)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteExercise performs a soft-delete of an exercise.
func (r *exerciseRepository) DeleteExercise(ctx context.Context, id string) error {
	query := `
		UPDATE exercises
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1::uuid AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func setVolume(weightKg *float64, reps *int) float64 {
	if weightKg == nil || reps == nil {
		return 0
	}
	return *weightKg * float64(*reps)
}

func intPointerFromNull(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	converted := int(value.Int64)
	return &converted
}

func floatPointerFromNull(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	converted := value.Float64
	return &converted
}

func upsertExerciseRecord(records []model.ExercisePersonalRecord, next model.ExercisePersonalRecord) []model.ExercisePersonalRecord {
	for index, record := range records {
		if record.Type == next.Type {
			records[index] = next
			return records
		}
	}
	return append(records, next)
}

func calculateExerciseTrend(points []model.ExerciseProgressPoint) string {
	if len(points) < 2 {
		return "stable"
	}

	recentCount := 3
	if len(points) < recentCount {
		recentCount = len(points)
	}

	recent := points[len(points)-recentCount:]
	first := recent[0].VolumeKg
	last := recent[len(recent)-1].VolumeKg

	if first == 0 && last == 0 {
		return "stable"
	}

	delta := last - first
	threshold := 0.05
	if first != 0 {
		changeRatio := delta / first
		if changeRatio > threshold {
			return "up"
		}
		if changeRatio < -threshold {
			return "down"
		}
		return "stable"
	}

	if delta > 0 {
		return "up"
	}

	return "stable"
}
