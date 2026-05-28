# Architecture

LiteGym is a layered monorepo application built around a Go backend and a React frontend. The repository structure favors explicit module boundaries and a service-oriented backend rather than direct handler-to-database coupling.

## High-level architecture

```text
Browser
  |
  v
Frontend (React + Vite)
  |
  v
Backend API (Go + Gin)
  |
  +--> PostgreSQL
  |
  `--> Gemini API
```

## Monorepo layout

```text
backend/          API server and business logic
frontend/         client application
e2e/              browser-level tests
postgress-local/  local database schema and seed
docs/             project documentation
```

## Backend layering

The backend follows a clear internal layering pattern:

```text
transport/ -> service/ -> repository/ -> database
```

Each layer has a distinct purpose:

- `transport`: HTTP router, middleware, and request handlers
- `service`: business logic, validation, orchestration, cross-repository flows
- `repository`: persistence queries and database-facing behavior
- `model`: domain and response structs shared inside the backend
- `config`: environment-driven runtime configuration

## Frontend structure

The frontend is route-driven and centered around authenticated application pages:

- `/` login
- `/dashboard`
- `/profile`
- `/routines`
- `/exercises`
- `/admin`
- `/support`

The application shell is provided by:

- `frontend/src/components/AppLayout.tsx`
- `frontend/src/components/AuthenticatedLayoutRoute.tsx`

## Runtime bootstrap

Server startup happens in:

- `backend/cmd/server/main.go`

Bootstrap flow:

1. load environment configuration
2. create PostgreSQL pool
3. initialize repositories
4. initialize services
5. initialize handlers
6. create auth middleware
7. register routes
8. start HTTP server with graceful shutdown

## Request lifecycle

Most authenticated requests flow like this:

1. browser sends a request with cookie credentials
2. Gin router receives the request
3. auth middleware validates the session token
4. handler parses input and calls the service
5. service validates data and coordinates repositories
6. repository runs SQL through pgx
7. response is serialized back to JSON

## AI generation lifecycle

The AI routine flow is slightly richer:

1. frontend submits AI generation request
2. backend validates auth and payload
3. backend loads relevant exercises and compact user context
4. backend builds a Gemini prompt
5. Gemini returns routine JSON
6. frontend displays a preview modal
7. user confirms save
8. backend resolves existing exercises or creates new private ones
9. routine and planned sets are persisted

## Core domains

The system revolves around these domains:

- users and user profiles
- exercises
- routines and routine exercises
- workout sessions, workout exercises, and workout sets
- support tickets
- social sharing and friendships
- AI routine generation

## Storage model

The database is relational and normalized around user-owned and official content.

Important design choices:

- official exercises have `owner_user_id = NULL`
- private exercises must have a non-null `owner_user_id`
- routines and workouts are user-scoped
- routine sets are stored separately from performed workout sets
- AI generation usage is logged in a dedicated table

## Documentation cross-reference

- backend details: [backend.md](backend.md)
- frontend details: [frontend.md](frontend.md)
- database details: [database.md](database.md)
- AI details: [ai-integration.md](ai-integration.md)
