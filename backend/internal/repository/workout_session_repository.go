package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkoutSessionRepository defines the persistence operations for workout sessions.
type WorkoutSessionRepository interface {
	ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error)
	ListRecentWorkoutHistoryByUser(ctx context.Context, userID string, limit int) ([]model.AIRoutineRecentWorkoutSession, error)
	ListTrainingDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error)
	ListMuscleDistributionByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error)
}

type workoutSessionRepository struct {
	db *pgxpool.Pool
}

// NewWorkoutSessionRepository creates a new workout session repository backed by PostgreSQL.
func NewWorkoutSessionRepository(db *pgxpool.Pool) WorkoutSessionRepository {
	return &workoutSessionRepository{db: db}
}

func (r *workoutSessionRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]model.OverviewWorkoutSummary, error) {
	return listOverviewWorkoutSummaries(ctx, r.db, userID, limit)
}

func (r *workoutSessionRepository) ListRecentWorkoutHistoryByUser(
	ctx context.Context,
	userID string,
	limit int,
) ([]model.AIRoutineRecentWorkoutSession, error) {
	rows, err := r.db.Query(ctx, `
		WITH recent_sessions AS (
			SELECT
				ws.id,
				COALESCE(ws.name, '') AS session_name,
				COALESCE(rt.name, '') AS routine_name,
				ws.performed_at,
				COALESCE(ws.duration_minutes, 0) AS duration_minutes
			FROM public.workout_sessions ws
			LEFT JOIN public.routines rt ON rt.id = ws.routine_id
			WHERE ws.user_id = $1::uuid
				AND ws.performed_at IS NOT NULL
			ORDER BY ws.performed_at DESC, ws.id::text DESC
			LIMIT $2
		)
		SELECT
			rs.id::text,
			rs.session_name,
			rs.routine_name,
			rs.performed_at,
			rs.duration_minutes,
			we.id::text,
			we.exercise_order,
			e.id::text,
			e.name,
			e.muscle_group,
			COALESCE(e.exercise_type, ''),
			wset.set_number,
			wset.reps,
			wset.weight_kg
		FROM recent_sessions rs
		LEFT JOIN public.workout_exercises we ON we.workout_session_id = rs.id
		LEFT JOIN public.exercises e ON e.id = we.exercise_id AND e.deleted_at IS NULL
		LEFT JOIN public.workout_sets wset ON wset.workout_exercise_id = we.id
		ORDER BY rs.performed_at DESC, rs.id::text DESC, we.exercise_order ASC, we.id::text ASC, wset.set_number ASC, wset.id::text ASC
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type exerciseAggregate struct {
		exercise model.AIRoutineRecentWorkoutExercise
	}

	type sessionAggregate struct {
		session       model.AIRoutineRecentWorkoutSession
		exercises     map[string]*exerciseAggregate
		exerciseOrder []string
	}

	sessions := make(map[string]*sessionAggregate)
	sessionOrder := make([]string, 0)

	for rows.Next() {
		var (
			sessionID         string
			sessionName       string
			routineName       string
			performedAt       time.Time
			durationMinutes   int
			workoutExerciseID sql.NullString
			exerciseOrder     sql.NullInt64
			exerciseID        sql.NullString
			exerciseName      sql.NullString
			muscleGroup       sql.NullString
			exerciseType      sql.NullString
			setNumber         sql.NullInt64
			reps              sql.NullInt64
			weightKg          sql.NullFloat64
		)

		if err := rows.Scan(
			&sessionID,
			&sessionName,
			&routineName,
			&performedAt,
			&durationMinutes,
			&workoutExerciseID,
			&exerciseOrder,
			&exerciseID,
			&exerciseName,
			&muscleGroup,
			&exerciseType,
			&setNumber,
			&reps,
			&weightKg,
		); err != nil {
			return nil, err
		}

		sessionAgg, ok := sessions[sessionID]
		if !ok {
			sessionAgg = &sessionAggregate{
				session: model.AIRoutineRecentWorkoutSession{
					SessionID:       sessionID,
					SessionName:     sessionName,
					RoutineName:     routineName,
					StartedAt:       performedAt,
					DurationMinutes: durationMinutes,
					Exercises:       []model.AIRoutineRecentWorkoutExercise{},
				},
				exercises:     make(map[string]*exerciseAggregate),
				exerciseOrder: make([]string, 0),
			}
			sessions[sessionID] = sessionAgg
			sessionOrder = append(sessionOrder, sessionID)
		}

		if !exerciseID.Valid || strings.TrimSpace(exerciseID.String) == "" || !workoutExerciseID.Valid {
			continue
		}

		exerciseAgg, exists := sessionAgg.exercises[workoutExerciseID.String]
		if !exists {
			exerciseAgg = &exerciseAggregate{
				exercise: model.AIRoutineRecentWorkoutExercise{
					ExerciseID:    exerciseID.String,
					ExerciseName:  exerciseName.String,
					MuscleGroup:   muscleGroup.String,
					ExerciseType:  exerciseType.String,
					ExerciseOrder: int(exerciseOrder.Int64),
					Sets:          []model.AIRoutineRecentWorkoutSet{},
				},
			}
			sessionAgg.exercises[workoutExerciseID.String] = exerciseAgg
			sessionAgg.exerciseOrder = append(sessionAgg.exerciseOrder, workoutExerciseID.String)
		}

		if setNumber.Valid {
			exerciseAgg.exercise.Sets = append(exerciseAgg.exercise.Sets, model.AIRoutineRecentWorkoutSet{
				SetNumber: int(setNumber.Int64),
				Reps:      intPointerFromNull(reps),
				WeightKg:  floatPointerFromNull(weightKg),
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	history := make([]model.AIRoutineRecentWorkoutSession, 0, len(sessionOrder))
	for _, sessionID := range sessionOrder {
		sessionAgg := sessions[sessionID]
		session := sessionAgg.session
		session.Exercises = make([]model.AIRoutineRecentWorkoutExercise, 0, len(sessionAgg.exerciseOrder))
		for _, workoutExerciseID := range sessionAgg.exerciseOrder {
			session.Exercises = append(session.Exercises, sessionAgg.exercises[workoutExerciseID].exercise)
		}
		history = append(history, session)
	}

	return history, nil
}

func (r *workoutSessionRepository) ListTrainingDatesInRange(ctx context.Context, userID string, from, to time.Time) ([]time.Time, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT DATE(ws.performed_at)
		FROM public.workout_sessions ws
		WHERE ws.user_id = $1::uuid
			AND ws.performed_at >= $2
			AND ws.performed_at < $3
		ORDER BY DATE(ws.performed_at) DESC
	`, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dates := make([]time.Time, 0)
	for rows.Next() {
		var trainedDate time.Time
		if err := rows.Scan(&trainedDate); err != nil {
			return nil, err
		}
		dates = append(dates, trainedDate)
	}

	return dates, rows.Err()
}

func (r *workoutSessionRepository) ListMuscleDistributionByUser(ctx context.Context, userID string, from, to time.Time) ([]model.OverviewMuscleGroupShare, int, error) {
	return listMuscleDistributionByUser(ctx, r.db, userID, from, to)
}
