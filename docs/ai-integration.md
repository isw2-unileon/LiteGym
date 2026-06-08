# AI integration

This page explains the Gemini-based routine flow in LiteGym.

## Current feature scope

The AI integration currently covers:

- routine generation preview
- user confirmation before persistence
- resolving existing exercises during save
- creating missing user-owned exercises when needed

You will find the AI feature in the routines area of the app.

## Entry points

Backend wiring starts in:

- `backend/cmd/server/main.go`

The main runtime service is:

- `backend/internal/service/routine_ai_service.go`

The HTTP handler is:

- `backend/internal/transport/handlers/routine_handler.go`

The routes are:

- `POST /api/routines/ai/generate`
- `POST /api/routines/ai/save`

## Configuration

Relevant environment variables:

- `GEMINI_API_KEY`
- `GEMINI_MODEL`

Important behavior:

- if `GEMINI_API_KEY` is missing, generation fails with `503`
- if `GEMINI_MODEL` is empty when it reaches the AI service, the service falls back to `gemini-2.5-flash`

## Request contract

Generation request fields:

- `objective`
- `target_muscle_groups`
- `mandatory_exercises`
- `notes`
- `duration_minutes`

This is intentionally user-friendly:

- mandatory exercises are sent by name, not by raw id
- notes let the user add free-text instructions such as preferences or restrictions

## Generation flow

### Step 1: generate a preview

When a request hits `POST /api/routines/ai/generate`, the backend:

1. validates the authenticated user
2. validates the request body
3. loads available exercises for the user context
4. builds compact recent-training context
5. builds a Gemini prompt
6. calls Gemini over HTTP
7. parses and normalizes the returned JSON
8. returns a preview payload to the frontend

### Step 2: save it after confirmation

When the user confirms the preview through `POST /api/routines/ai/save`, the backend:

1. iterates over generated exercises
2. tries to reuse the returned `exercise_id` if it already exists
3. otherwise searches existing exercises by normalized name, muscle group, and type
4. if still not found, creates a new private exercise owned by the authenticated user
5. persists the routine and its planned sets

## What is sent to Gemini

The backend sends a structured JSON prompt instead of relying only on free-form text.

Main blocks:

- `objective`
- `duration_minutes`
- `target_muscle_groups`
- `mandatory_exercises`
- `user_notes`
- `user_context`
- `exercise_catalog`
- `output_contract`

### `user_context`

The context is kept compact so the prompt stays token-friendly. It includes:

- `training_days_30d`
- `current_streak_days`
- `recent_workouts`
- `recent_training_history`
- `recent_routines`
- `top_muscle_groups`
- `body_metrics`

### `exercise_catalog`

Gemini also receives a filtered exercise catalog with:

- `id`
- `name`
- `muscle_group`
- `exercise_type`

This keeps the model focused on exercises the application actually knows about.

## Prompt behavior

The current system instruction tells Gemini to:

- use recent training history as the main history signal
- respect user notes and mandatory exercises strongly
- build the most complete and sensible routine for the available time
- choose exercise count freely
- avoid a fixed one-to-one mapping between target muscle groups and exercises
- return valid JSON only

This was added to avoid the simple one-exercise-per-target pattern.

## Exercise resolution strategy

Saving a generated routine uses a tolerant resolution flow:

1. trust `exercise_id` only if it exists in the database
2. otherwise search by normalized exercise name
3. narrow by muscle group and exercise type when available
4. create a new private exercise as a last resort

The created exercise:

- is marked as non-official
- receives `owner_user_id` equal to the authenticated user

## Failure modes

Common generation failures:

- missing API key
- Gemini quota exhaustion
- model not available for the active plan
- provider `429` or other non-2xx status
- malformed provider response

Error handling behavior:

- invalid input -> `400`
- rate-limit exceeded -> `429`
- provider unavailable or malformed response -> `503`
- unexpected internal issues -> `500`

## Rate limiting

The service enforces a per-user rate limit based on `ai_routine_generation_logs`:

- `aiRoutineRateLimit = 2` requests
- `aiRoutineRateWindow = time.Hour`

So each user can run at most 2 AI routine requests per hour. Exceeding the limit
returns `429`. The provider may still enforce its own external quotas on top of this.

## Frontend flow

The frontend page responsible for AI routine generation is:

- `frontend/src/pages/UserRoutinesPage.tsx`

Current UX flow:

- user opens the AI form
- user sets objective and duration
- user selects mandatory exercises from the available catalog
- user adds optional free-text notes
- frontend requests a preview
- preview is shown in a dedicated modal
- user confirms save only after reviewing the result

## Testing

AI-related testing exists at a few levels:

- service tests in `backend/internal/service/routine_ai_service_test.go`
- integration-style handler test in `backend/internal/transport/handlers/routine_handler_integration_test.go`
- frontend behavior test in `frontend/src/pages/UserRoutinesPage.test.tsx`

The real Gemini integration test is defensive:

- it reads credentials from backend env
- it can skip when no key is configured
- it now also skips when provider quota is exhausted
