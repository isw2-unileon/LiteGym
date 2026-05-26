# Getting started

This guide explains how to boot LiteGym locally and what each part of the repository is responsible for.

## Prerequisites

You should have the following tools available:

- Go
- Node.js and npm
- Docker or Podman with compose support

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

There are two environment stories in this project:

- backend runtime variables
- frontend Vite variables

The example backend-oriented environment file is in the repository root:

- `.env.example`

For day-to-day local work, these are the most relevant files:

- `.env.local`
- `backend/.env`
- `frontend/.env`

Detailed variable behavior is documented in [configuration.md](configuration.md).

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

Open two terminals.

Terminal 1:

```bash
make run-backend
```

Terminal 2:

```bash
make run-frontend
```

The usual local URLs are:

- frontend: `http://localhost:5173`
- backend: `http://localhost:8080`

## Full stack with compose

If you want to run the complete application snapshot, including backend, frontend, and PostgreSQL through one compose file:

```bash
make start-app-snapshot
```

Related commands:

```bash
make down-app-snapshot
make delete-app-snapshot
```

The full stack compose definition is in:

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
