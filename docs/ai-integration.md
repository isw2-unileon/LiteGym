# AI integration

This document describes the current Gemini-based AI routine generation flow in LiteGym.

## Current feature scope

The AI integration currently covers:

- routine generation preview
- user confirmation before persistence
- resolving existing exercises during save
- creating missing user-owned exercises when needed

The AI feature is exposed through the routines area of the application.

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

## Current request contract

Generation request fields:

- `objective`
- `target_muscle_groups`
- `mandatory_exercises`
- `notes`
- `duration_minutes`

This is intentionally user-friendly:

- mandatory exercises are sent by name, not by raw id
- notes let the user add free-text instructions such as preferences or restrictions

## High-level generation flow

### Step 1: preview generation

When a request hits `POST /api/routines/ai/generate`, the backend:

1. validates the authenticated user
2. validates the request body
3. loads available exercises for the user context
4. builds compact recent-training context
5. builds a Gemini prompt
6. calls Gemini over HTTP
7. parses and normalizes the returned JSON
8. returns a preview payload to the frontend

### Step 2: explicit save

When the user confirms the preview through `POST /api/routines/ai/save`, the backend:

1. iterates over generated exercises
2. tries to reuse the returned `exercise_id` if it already exists
3. otherwise searches existing exercises by normalized name, muscle group, and type
4. if still not found, creates a new private exercise owned by the authenticated user
5. persists the routine and its planned sets

## What is sent to Gemini

The backend sends a structured JSON prompt rather than a natural-language-only prompt.

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

The context is compact and intentionally token-aware. It includes:

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

This helps constrain the model toward valid exercises known by the application.

## Prompt behavior

The current system instruction tells Gemini to:

- use recent training history as the main history signal
- respect user notes and mandatory exercises strongly
- build the most complete and sensible routine for the available time
- choose exercise count freely
- avoid a fixed one-to-one mapping between target muscle groups and exercises
- return valid JSON only

This was added to avoid a simplistic pattern where the model returned one exercise per selected target muscle.

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

The service still contains per-user rate-limit logic based on `ai_routine_generation_logs`, but:

- `aiRoutineRateLimitEnabled` is currently `false`

So at the moment the internal application-side limit is disabled, even though the provider may still enforce external quotas.

## Frontend flow

The frontend page responsible for AI routine generation is:

- `frontend/src/pages/UserRoutinesPage.tsx`

Current UX behavior:

- user opens the AI form
- user sets objective and duration
- user selects mandatory exercises from the available catalog
- user adds optional free-text notes
- frontend requests a preview
- preview is shown in a dedicated modal
- user confirms save only after reviewing the result

## Testing

AI-related testing exists at multiple levels:

- service tests in `backend/internal/service/routine_ai_service_test.go`
- integration-style handler test in `backend/internal/transport/handlers/routine_handler_integration_test.go`
- frontend behavior test in `frontend/src/pages/UserRoutinesPage.test.tsx`

The real Gemini integration test is defensive:

- it reads credentials from backend env
- it can skip when no key is configured
- it now also skips when provider quota is exhausted
