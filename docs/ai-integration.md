# AI System in LiteGym

This document explains in detail how Gemini-based routine generation works inside LiteGym.

The current implementation lives entirely in the Go backend. The frontend does not yet expose a dedicated screen that triggers this feature directly, so the important part today is the API -> service -> Gemini -> response flow.

## Executive Summary

Routine generation follows this path:

1. An authenticated user calls `POST /api/routines/ai/generate`.
2. The handler validates the session and request body.
3. The service checks the per-user usage limit.
4. The service gathers a compact summary of the user, the latest workout history at session/exercise/set level, and a filtered exercise catalog.
5. It builds a structured JSON prompt.
6. It calls Gemini over HTTP.
7. Gemini returns JSON.
8. The backend parses, normalizes, and saves the generated routine as a permanent user routine.
9. The backend returns the generated JSON, the saved `routine_id`, and rate-limit metadata.
10. A generation log is also persisted, so usage can be counted.

## Where It Starts

Dependency injection happens at server startup.

- [backend/cmd/server/main.go](../backend/cmd/server/main.go)

At that point the application builds:

- the routine, workout session, and body metric repositories
- `RoutineAIService`
- `RoutineHandler`

The relevant piece is:

```go
routineAIService := service.NewRoutineAIService(
    routineRepo,
    workoutSessionRepo,
    bodyMetricRepo,
    cfg.GeminiAPIKey,
    cfg.GeminiModel,
)
```

That means the AI service is not standalone. It depends on the database to build useful context before contacting the model, including the latest workout sessions and their set-level detail.

## Configuration

The environment variables that control this feature are:

- `GEMINI_API_KEY`
- `GEMINI_MODEL`

They are loaded in:

- [backend/internal/config/config.go](../backend/internal/config/config.go)

If no model is set, the backend falls back to:

- `gemini-1.5-flash`

If there is no API key, generation cannot run and the backend returns `503`.

## Exposed Endpoint

The protected route is registered in:

- [backend/internal/transport/router.go](../backend/internal/transport/router.go)

Route:

```http
POST /api/routines/ai/generate
```

It is authenticated. If there is no user in context, the handler returns `401`.

## Input Contract

The request body expects:

- `objective`: training goal
- `target_muscle_groups`: list of muscle groups
- `mandatory_exercise_ids`: list of required exercises
- `duration_minutes`: target duration

The model is defined in:

- [backend/internal/model/routine_ai.go](../backend/internal/model/routine_ai.go)

## What the Handler Does

The handler does not talk to Gemini directly. It only coordinates validation and error translation.

File:

- [backend/internal/transport/handlers/routine_handler.go](../backend/internal/transport/handlers/routine_handler.go)

Responsibilities:

- check authentication
- parse JSON
- call `RoutineAIService.GenerateRoutineJSON`
- translate errors into HTTP responses

Error mapping:

- invalid input -> `400`
- rate limit exceeded -> `429`
- provider unavailable -> `503`
- missing API key -> `503`
- any other failure -> `500`

## How Usage Is Limited

There is a simple per-user limit:

- maximum `2` generations per hour

That logic lives in:

- [backend/internal/service/routine_ai_service.go](../backend/internal/service/routine_ai_service.go)
- [backend/internal/repository/routine_repository.go](../backend/internal/repository/routine_repository.go)

The sequence is:

1. compute the time window
2. count previous generations in `ai_routine_generation_logs`
3. if the limit is already reached, stop before calling Gemini
4. if generation succeeds, insert a new log

This avoids spending provider calls when the user has already exhausted their quota.

## What User Data Is Sent to Gemini

The key design decision is that the backend does not send raw history. It sends a compact summary designed to use fewer tokens and be easier for the model to understand.

That summary is built in:

- [backend/internal/service/routine_ai_service.go](../backend/internal/service/routine_ai_service.go)

### `user_context`

The prompt includes a `user_context` object with:

- `training_days_30d`
- `current_streak_days`
- `recent_workouts`
- `recent_training_history`
- `recent_routines`
- `top_muscle_groups`
- `body_metrics`

### What Each Block Means

- `training_days_30d`: number of days trained in the last 30 days.
- `current_streak_days`: current training streak.
- `recent_workouts`: recent sessions with name, associated routine, duration, and number of exercises.
- `recent_training_history`: the latest sessions expanded into exercises and sets, including `reps` and `weight_kg`.
- `recent_routines`: recent routines with name and exercise count.
- `top_muscle_groups`: most trained muscle groups, with count and percentage.
- `body_metrics`: the latest body metrics and deltas when available.

### Why This Reduces Tokens

Because Gemini does not need:

- every historical session
- every set from the whole account history
- every exercise the user has ever done
- the full user profile from the database

Instead, it receives a compact signal that carries the useful information. That reduces:

- prompt size
- cost
- latency
- semantic noise

## What Exercise Catalog Is Sent

Before calling Gemini, the backend loads a filtered catalog of candidate exercises.

The query lives in:

- [backend/internal/repository/routine_repository.go](../backend/internal/repository/routine_repository.go)

Included exercises are:

- official exercises
- or exercises owned by the user
- not deleted
- optionally filtered by target muscle groups

The goal is to prevent the model from inventing arbitrary exercises. It should choose from a known valid list.

Each catalog entry contains:

- `id`
- `name`
- `muscle_group`
- `exercise_type`

## How the Prompt Is Built

The prompt has two parts:

### 1. `system_instruction`

This is a short instruction that tells the model to:

- act as a workout planner
- use `user_context`, especially `recent_training_history`, as the main history signal
- return only valid JSON
- avoid Markdown

### 2. `contents`

A single JSON-serialized block is sent with:

- objective
- duration
- target muscle groups
- mandatory exercise IDs
- `user_context`
- `exercise_catalog`
- `output_contract`

File:

- [backend/internal/service/routine_ai_service.go](../backend/internal/service/routine_ai_service.go)

## What `output_contract` Is

`output_contract` tells Gemini what shape the output must follow.

It includes fields such as:

- `name`
- `objective`
- `duration_minutes`
- `target_muscles`
- `mandatory_count`
- `generated_at`
- `generation_source`
- `exercises`

And each exercise must include:

- `exercise_id`
- `name`
- `muscle_group`
- `exercise_type`
- `is_mandatory`
- `recommended_sets`
- `recommended_reps`

This is not a hard schema validator by itself, but it makes the model much more likely to return parseable structured output.

## Recent Workout History

This is the new part of the prompt that was added after the initial AI integration.

The backend now sends a compact representation of the latest workouts, not just a high-level summary. Each session includes:

- session identity and display name
- routine name
- start time
- duration
- exercises performed in that session
- the order of each exercise
- sets per exercise
- repetitions per set
- weight per set

This data is produced by the workout session repository and folded into `user_context.recent_training_history`.

### Why This Matters

The model can now see not only that the user trained chest or legs recently, but also:

- which movements they already used
- the set and rep ranges they actually handled
- the weight progression between recent sessions

That gives Gemini more signal for:

- progressive overload
- exercise selection
- balancing repeated movements
- avoiding overly repetitive routines

### Token Tradeoff

This extra context does increase prompt size. In local real requests, the prompt token count moved from roughly `1117` tokens for the lighter summary to roughly `1640` tokens once the training history block was added.

That is still a controlled prompt size for the amount of signal being sent, but it is materially larger than the earlier version.

## Gemini Request

The HTTP request goes to:

```text
https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={apiKey}
```

The request body uses:

- `temperature: 0.3`
- `responseMimeType: application/json`

That combination aims for:

- more stable output
- less unnecessary creativity
- JSON that can be parsed directly

## How the Response Is Processed

Gemini returns a structure like `candidates -> content -> parts -> text`.

The backend:

1. parses the HTTP response
2. extracts the first `candidate`
3. extracts the first `part.text`
4. tries to unmarshal that text into `AIRoutineJSON`

If the text is missing or the JSON does not parse, generation fails.

File:

- [backend/internal/service/routine_ai_service.go](../backend/internal/service/routine_ai_service.go)

## What the Backend Normalizes Afterwards

Even if Gemini returns a valid routine, the service fills missing or incomplete fields:

- if `objective` is empty, it uses the original objective
- if `duration_minutes` is zero, it uses the requested duration
- if `target_muscles` is empty, it uses the normalized input muscle groups
- `generated_at` is always set by the backend to the current generation time
- `generation_source` is always set by the backend to `"gemini"`

That gives the frontend a more consistent response.

## What Is Stored in the Database

The generated routine is persisted as a real user routine.

The backend writes:

- a row in `public.routines`
- one row per generated exercise in `public.routine_exercises`
- a usage log in `public.ai_routine_generation_logs`

The saved routine is private to the requesting user:

- `user_id` is the authenticated user
- `is_predefined` is `false`
- `is_public` is `false`
- `name` comes from Gemini, with a backend fallback if it is missing
- `description` records that the routine was generated by Gemini and includes the requested objective and duration

The saved exercise rows contain:

- `routine_id`
- `exercise_id`
- `exercise_order`
- `notes`

The `notes` field stores the generated prescription in compact text form, for example sets, repetitions, and whether the exercise was mandatory. The current schema does not store sets, reps, or target weight as structured routine prescription columns.

Before saving, the backend checks that every generated `exercise_id` exists in the exercise catalog that was sent to Gemini. Unknown exercise IDs are rejected instead of saving an invalid routine.

The generation log is stored in:

- `public.ai_routine_generation_logs`

Schema:

- [postgress-local/initdb/01-schema.sql](../postgress-local/initdb/01-schema.sql)

That log is used to enforce the usage limit.

## Response Format Returned to the Client

The backend returns:

- `routine_json`
- `routine_id`
- `rate_limit`

The contract is defined in:

- [backend/internal/model/routine_ai.go](../backend/internal/model/routine_ai.go)

### `routine_json`

Contains the generated routine.

### `routine_id`

Contains the database ID of the permanent routine created from the AI response.

### `rate_limit`

Contains:

- total limit
- remaining calls
- current usage
- window size
- reset time

This lets the frontend display usage information without recalculating anything.

## Internal Flow Summary

```text
Authenticated user
  -> POST /api/routines/ai/generate
  -> handler validates session and JSON
  -> service validates input
  -> service checks rate limit
  -> service loads compact user summary
  -> service loads recent workout history with exercises and sets
  -> service loads exercise catalog
  -> service builds JSON prompt
  -> service calls Gemini
  -> service parses the response
  -> service normalizes fields
  -> service saves routine + routine exercises
  -> service logs the generation
  -> responds with routine + routine_id + rate limit
```

## What This Design Solves

This design balances four goals:

1. Enough context for Gemini to produce a useful routine.
2. Enough token efficiency so the call stays practical.
3. Structured output so post-processing stays simple.
4. Abuse control so provider usage does not grow without limits.

## Current Limitations

There are a few clear limitations today:

- no deep semantic validation of the generated JSON
- if Gemini returns valid JSON but a poor routine, the backend does not catch it
- the frontend does not yet expose a dedicated UI for this feature
- the user context is summarized, not exhaustive
- the recent training history increases prompt size, so the context needs to stay intentionally compact
- the saved routine exercise prescription is stored in `notes`, not as structured sets/reps/weight fields

## Possible Improvements

If this needs to be hardened later, reasonable next steps would be:

- validate the generated JSON against an internal schema
- add richer validation for exercise order, duration, volume, and generated notes
- enrich the context with a compact per-exercise performance summary
- persist the raw AI payload and generation metadata if auditing is needed
- add a frontend UI for launching generation and opening the saved routine
- tune the workout history window and session count if prompt size grows too much

## Conclusion

LiteGym currently uses the AI layer as a controlled backend routine generator:

- it authenticates the user
- it limits usage
- it summarizes the user history
- it filters valid exercises
- it calls Gemini with a compact prompt
- it expects pure JSON
- it saves the generated routine permanently
- it returns a structured routine and `routine_id` to the client

The most important design choice is not sending full history, but a compact behavioral summary instead. That reduces cost, simplifies the prompt, and makes the output more predictable.
