# LiteGym

LiteGym is a full-stack fitness application for managing routines, tracking workouts, reviewing performance history, and generating AI-assisted training plans. The project is organized as a monorepo with a Go backend, a React frontend, a PostgreSQL database, and a Playwright E2E test package.

The application currently focuses on these product areas:

- authentication and session-based access control
- exercise catalog management with official and user-owned exercises
- routine browsing, routine detail, and AI-generated routine preview/save flows
- workout session execution with exercises and sets
- dashboard and exercise insights
- support tickets and basic admin views

## Repository map

```text
.
|-- backend/                 Go API server
|   |-- cmd/server/          bootstrap entry point
|   `-- internal/            config, model, repository, service, transport
|-- frontend/                React + TypeScript + Vite application
|-- e2e/                     Playwright end-to-end tests
|-- postgress-local/         local PostgreSQL image, schema, seed data
|-- docs/                    project documentation
|-- compose.yaml             full local stack
`-- Makefile                 common developer commands
```

## Main stack

- Backend: Go, Gin, pgx, PostgreSQL
- Frontend: React 19, TypeScript, React Router, Vite
- Styling: CSS utility-first approach with Tailwind tooling in the frontend stack
- Testing: Go unit/integration tests, Vitest, Playwright
- AI: Gemini API for AI-generated routines

## Quick start

### Prerequisites

- Go
- Node.js and npm
- Docker or Podman with compose support

### Local setup

1. Copy the example environment file and fill in the values you need.
2. Start the local database.
3. Run backend and frontend in separate terminals.

Typical commands:

```bash
make start-postgres-db
make run-backend
make run-frontend
```

Frontend runs on `http://localhost:5173` and the backend on `http://localhost:8080`.

## Documentation

The detailed project documentation lives in [`docs/`](docs/index.md).

- [Documentation index](docs/index.md)
- [Getting started](docs/getting-started.md)
- [Architecture](docs/architecture.md)
- [Configuration](docs/configuration.md)
- [Backend guide](docs/backend.md)
- [Frontend guide](docs/frontend.md)
- [API reference](docs/api-reference.md)
- [Database guide](docs/database.md)
- [AI integration](docs/ai-integration.md)
- [Testing guide](docs/testing.md)

## Development commands

Common commands are exposed through the root `Makefile`:

```bash
make install
make run-backend
make run-frontend
make build-backend
make build-frontend
make test
make test-integration
make lint
make e2e
```

There are also stack management targets for the local database and the full application snapshot:

```bash
make start-postgres-db
make reset-postgres-db
make start-app-snapshot
make down-app-snapshot
```

## Current notes

- AI routine generation currently uses a preview-and-confirm flow.
- The AI service creates user-owned exercises automatically when Gemini proposes a valid new exercise that does not already exist in the catalog.
- AI routine endpoints are rate-limited by an in-memory transport middleware (see `RateLimiter.AI()`); the older DB-backed limiter in the AI service is currently disabled.
