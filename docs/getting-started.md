# Getting started

This guide walks through the local setup and gives you a map of the parts that matter when you first open the repo.

## Prerequisites

You should have the following tools available:

- Go `1.25.0`
- Node.js and npm; the repo does not pin an exact Node version, so use a current LTS release compatible with Vite 6
- Docker or Podman with compose support
- `make` and `curl`

The root `Makefile` automatically prefers `docker compose`, and falls back to `podman compose` if Docker is not available.

## Repository structure

```text
backend/          Go backend application
frontend/         React frontend application
e2e/              Playwright tests
postgress-local/  local PostgreSQL image, schema, and seed
docs/             project documentation
```

## Environment files

There are two environment layers in this project:

- backend runtime variables
- frontend Vite variables

The example backend-oriented environment file is in the repository root:

- `.env.example`

For day-to-day local work, these are the most relevant files:

- `.env.local`
- `backend/.env`
- `frontend/.env`

Detailed variable behavior is documented in [configuration.md](configuration.md).

If you run the backend on your host machine, point the database URL at `localhost:5432`. The root [`README.md`](../README.md) shows the exact override.

## First-time setup

Install project dependencies:

```bash
make install
```

This command:

- installs Air for Go hot reload
- installs golangci-lint
- downloads Go modules
- installs frontend dependencies
- installs Playwright package dependencies

## Start the local database

The local PostgreSQL stack uses the files in `postgress-local/`.

```bash
make start-postgres-db
```

Useful related commands:

```bash
make stop-postgres-db
make down-postgres-db
make delete-postgres-db
make reset-postgres-db
make logs-postgres-db
```

`reset-postgres-db` recreates the database and replays the schema and seed scripts from scratch.

## Run backend and frontend

Use two terminals.

Terminal 1:

```bash
make run-backend
```

Terminal 2:

```bash
make run-frontend
```

You should see the app at:

- frontend: `http://localhost:5173`
- backend: `http://localhost:8080`

## Full stack with compose

If you want everything in one go, including backend, frontend, and PostgreSQL through a single compose file:

```bash
make start-app-snapshot
```

Related commands:

```bash
make down-app-snapshot
make delete-app-snapshot
```

The full-stack compose file is:

- `compose.yaml`

## Build commands

Backend:

```bash
make build-backend
```

Frontend:

```bash
make build-frontend
```

## Test commands

All main tests:

```bash
make test
```

Backend integration tests against local PostgreSQL:

```bash
make test-integration
```

Frontend lint and backend lint:

```bash
make lint
```

Playwright E2E suite:

```bash
make e2e
```

## Common local issues

### Frontend cannot reach the backend

Check:

- backend is running on port `8080`
- frontend dev server is running on port `5173`
- Vite proxy settings in `frontend/vite.config.ts`

### Backend says database is unavailable

Check:

- local Postgres stack is running
- `DATABASE_URL` points to the correct database
- the database health endpoint `/api/db/health`

### AI routine generation returns `503`

Common causes:

- missing `GEMINI_API_KEY`
- unsupported or rate-limited `GEMINI_MODEL`
- Gemini free tier quota exhausted

See [ai-integration.md](ai-integration.md) for details.
