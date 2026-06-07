# API reference

This document summarizes the HTTP surface exposed by the Go backend router.

All API routes live under `/api`, except the plain health endpoint `/health`.

## Health endpoints

### `GET /health`

Simple process health check.

### `GET /api/db/health`

Protected endpoint that pings the configured database.

## Auth endpoints

### `POST /api/auth/register`

Registers a new user. Rate-limited in transport middleware.

### `POST /api/auth/login`

Authenticates a user and sets the auth cookie. Rate-limited in transport middleware.

### `POST /api/auth/logout`

Clears the auth cookie.

### `GET /api/auth/me`

Returns the current authenticated user session payload.

## User endpoints

### `POST /api/users`

Creates a user.

### `GET /api/users`

Protected. Lists users.

### `GET /api/users/me`

Protected. Returns the authenticated user's profile view.

### `GET /api/users/:id`

Protected. Returns user detail by id.

### `DELETE /api/users/:id`

Protected. Deletes a user.

## Profile endpoints

### `GET /api/profile/dashboard`

Protected. Returns the profile dashboard view.

### `PUT /api/profile/goals`

Protected. Updates the user's goals.

### `POST /api/profile/metrics`

Protected. Adds a body metric entry.

### `POST /api/profile/ai-analysis`

Protected. Returns an AI-based profile analysis. Rate-limited in transport middleware.

## Exercise endpoints

### `GET /api/exercises`

Protected. Paginated exercise list.

Supported query parameters:

- `page`
- `limit`
- `search`
- `type`
- `muscle_group`
- `official`

### `POST /api/exercises`

Protected. Creates an exercise.

Important rule:

- official exercises are admin-controlled
- non-official exercises are tied to the authenticated user

### `GET /api/exercises/metadata`

Protected. Returns valid exercise types and muscle groups.

### `GET /api/exercises/:id`

Protected. Exercise detail by id.

### `PUT /api/exercises/:id`

Protected. Updates an exercise.

### `DELETE /api/exercises/:id`

Protected. Deletes an exercise.

### `GET /api/exercises/:id/insights`

Protected. Returns aggregated exercise performance insights for the authenticated user.

### `GET /api/exercises/:id/workout-sessions`

Protected. Returns recent workout sessions that include the selected exercise.

## Routine endpoints

### `GET /api/routines`

Protected. Lists routines for the authenticated user.

### `GET /api/routines/:id`

Protected. Returns full routine detail including planned sets.

### `POST /api/routines`

Protected. Creates a manual routine.

### `PUT /api/routines/:id`

Protected. Updates a routine.

### `DELETE /api/routines/:id`

Protected. Deletes a routine.

### `POST /api/routines/:id/duplicate`

Protected. Duplicates an existing routine, including its exercises and sets.

### `POST /api/routines/ai/generate`

Protected. Generates an AI routine preview.

Current request body:

```json
{
  "objective": "Ganar fuerza",
  "target_muscle_groups": ["chest", "back", "legs"],
  "mandatory_exercises": ["Press banca", "Sentadilla"],
  "notes": "Prioriza ejercicios compuestos y evita mucho volumen de hombro.",
  "duration_minutes": 60
}
```

Current response body shape:

```json
{
  "routine_json": {
    "name": "Rutina generada",
    "objective": "Ganar fuerza",
    "duration_minutes": 60,
    "target_muscles": ["chest", "back", "legs"],
    "mandatory_count": 2,
    "generated_at": "2026-05-25T10:00:00Z",
    "generation_source": "gemini",
    "exercises": []
  },
  "rate_limit": {
    "limit": 2,
    "remaining": 1,
    "used_in_current_window": 1,
    "window_seconds": 3600,
    "reset_at": "2026-05-25T11:00:00Z"
  }
}
```

Notes:

- the preview does not persist the routine
- provider failures are surfaced as `503`
- AI generation is rate-limited per user (2 requests/hour); exceeding it returns `429`

### `POST /api/routines/ai/save`

Protected. Persists a previously generated preview.

Request body:

```json
{
  "routine_json": {
    "...": "preview payload returned by /api/routines/ai/generate"
  }
}
```

Behavior:

- resolves existing exercises by id or by name/domain match
- creates missing user-owned exercises when needed
- persists routine, routine exercises, and planned sets

### `POST /api/routines/:id/ai/upgrade`

Protected. Generates an AI-improved version of an existing routine as a preview.
Rate-limited per user like `ai/generate`.

### `POST /api/routines/:id/ai/save-as-new`

Protected. Saves an AI-upgraded preview as a new routine.

### `PUT /api/routines/:id/ai/overwrite`

Protected. Overwrites an existing routine with an AI-upgraded version.

## Dashboard endpoints

### `GET /api/dashboard`

Protected. Returns dashboard-oriented overview data.

## Support endpoints

### `POST /api/tickets`

Protected. Creates a support ticket.

### `GET /api/tickets`

Protected. Lists support tickets.

### `PATCH /api/tickets/:id/close`

Protected. Closes a support ticket.

## Workout endpoints

### `POST /api/workouts/planned`

Protected. Creates a planned workout session.

### `POST /api/workout/start`

Protected. Starts a workout session.

### `GET /api/workout/:id`

Protected. Returns workout detail.

### `GET /api/workout/:id/detail`

Protected. Returns the aggregated workout detail (session, exercises, and sets).

### `POST /api/workout/:id/finish`

Protected. Finishes a workout session.

### `DELETE /api/workout/:id`

Protected. Deletes a workout session.

### `POST /api/workout/:id/exercise`

Protected. Adds an exercise to a workout.

### `GET /api/workout/:id/exercises`

Protected. Lists workout exercises for a workout session.

### `POST /api/workout/:id/exercises/:exercise_id/set`

Protected. Creates a workout set.

### `GET /api/workout/:id/exercises/:exercise_id/sets`

Protected. Lists sets for a workout exercise.

### `POST /api/workout/:id/exercises/:exercise_id/sets/:set_id`

Protected. Updates a workout set.

### `DELETE /api/workout/:id/exercises/:exercise_id/sets/:set_id`

Protected. Removes a workout set.
