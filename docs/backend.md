# Backend guide

The backend is a Go application using Gin for HTTP transport, pgx for PostgreSQL access, and a service/repository architecture for domain logic.

## Directory structure

```text
backend/
|-- cmd/server/                 server bootstrap
`-- internal/
    |-- config/                 environment loading
    |-- model/                  domain and response structs
    |-- repository/             SQL-backed persistence
    |-- service/                business logic
    `-- transport/
        |-- handlers/           HTTP handlers
        `-- middleware/         auth and request middleware
```

## Bootstrap

Bootstrap happens in:

- `backend/cmd/server/main.go`

It wires:

- config loading
- database pool
- repositories
- services
- handlers
- auth middleware
- Gin router
- graceful shutdown

## Repositories

Repositories encapsulate persistence logic per domain.

Main repository files:

- `user_repository.go`
- `exercise_repository.go`
- `routine_repository.go`
- `workout_repository.go`
- `workout_session_repository.go`
- `body_metric_repository.go`
- `ticket_repository.go`
- `overview_workout_repository.go`

Typical responsibilities:

- `UserRepository`: users and profile-related persistence
- `ExerciseRepository`: exercise retrieval, filtering, creation, update, delete, insights history
- `RoutineRepository`: routine lists, detail loading, AI routine persistence, AI usage logging
- `WorkoutRepository`: workout sessions, workout exercises, workout sets
- `BodyMetricRepository`: body metric history
- `TicketRepository`: support ticket persistence

## Services

Services enforce application rules and orchestrate repositories.

Main service files:

- `user_service.go`
- `exercise_service.go`
- `routine_service.go`
- `routine_ai_service.go`
- `workout_service.go`
- `overview_service.go`
- `ticket_service.go`
- `token_service.go`

### `TokenService`

Responsible for session token creation and validation.

### `UserService`

Responsible for user lifecycle operations and lookup by id.

### `ExerciseService`

Responsible for:

- validation
- domain normalization
- official versus private exercise rules
- list filtering and pagination
- insights and related workout-session history

### `RoutineService`

Responsible for:

- listing routines for the authenticated user
- loading routine detail with exercise and planned set structure

### `RoutineAIService`

Responsible for:

- validating AI generation requests
- optional AI rate limit logic
- building compact user training context
- calling Gemini
- preview response generation
- save-confirmation flow
- creating missing user-owned exercises during save

### `WorkoutService`

Responsible for:

- starting workouts
- finishing workouts
- managing workout exercises
- creating and updating workout sets
- planned and performed workout flows

### `OverviewService`

Responsible for dashboard aggregates such as:

- recent routines
- recent workouts
- body metrics summaries
- muscle distribution and activity over time

## Transport layer

The transport layer is centered in:

- `backend/internal/transport/router.go`

It defines:

- public endpoints
- protected endpoints
- auth middleware application
- health and database health checks
- CORS behavior

## Middleware

The main middleware is:

- `backend/internal/transport/middleware/auth_middleware.go`

Current auth behavior:

- reads token from cookie
- validates and extracts claims
- injects user id and role into context
- validates that the user still exists in the database
- rejects stale tokens for deleted or missing users

## Core handlers

Important handlers:

- `auth_handler.go`
- `user_handler.go`
- `exercise_handler.go`
- `routine_handler.go`
- `overview_handler.go`
- `ticket_handler.go`
- `workout_handler.go`
- `health_handler.go`

Each handler keeps HTTP-specific concerns:

- reading request params
- request body binding
- mapping service errors to status codes
- JSON responses

## Logging

The backend currently uses structured logging with `log/slog`.

Examples of useful logging areas:

- server startup and shutdown
- AI generation flow
- provider failures and response status issues
- database connectivity at boot

## Timeouts

The HTTP server currently uses:

- `ReadTimeout: 10s`
- `WriteTimeout: 60s`

The longer write timeout is especially relevant for AI generation endpoints, which can take several seconds depending on Gemini response time.

## Backend extension guidelines

When adding a new feature:

1. add or update models if needed
2. add repository operations
3. implement service rules
4. expose handler endpoints
5. register routes
6. add unit tests
7. add integration tests when persistence behavior matters
