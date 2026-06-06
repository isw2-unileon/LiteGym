# LiteGym

LiteGym is a full-stack fitness app for routines, workouts, progress tracking, and AI-assisted training plans. The repository is a monorepo with a Go API, a React frontend, PostgreSQL, and end-to-end tests.

## What It Covers

- authentication and session cookies
- exercise catalog management with official and user-owned exercises
- routine browsing, routine detail, and AI preview/save flows
- workout sessions with exercises and sets
- dashboard insights and support tickets
- basic admin views

## Requirements

- Go `1.25.0`
- Node.js and npm; the repo does not pin a specific Node version, so use a current LTS release compatible with Vite 6
- Docker or Podman with compose support
- `make` and `curl`

## Local Setup

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd LiteGym
   ```
2. Create a local environment override file from the example:
   ```bash
   cp .env.example .env.local
   ```
3. Set `DATABASE_URL` in `.env.local` to match how you will run PostgreSQL.
4. Install dependencies:
   ```bash
   make install
   ```
5. Start PostgreSQL:
   ```bash
   make start-postgres-db
   ```
6. Run the backend and frontend in separate terminals:
   ```bash
   make run-backend
   make run-frontend
   ```

If you run the backend on your host machine, use this `DATABASE_URL` in `.env.local`:

```env
DATABASE_URL=postgres://test_user:test_password@localhost:5432/test_db?sslmode=disable
```

If you run the whole stack with compose, the backend can keep the container-based database URL from `backend/.env`.

The usual local URLs are:

- frontend: `http://localhost:5173`
- backend: `http://localhost:8080`

## Tests

Run all standard tests:

```bash
make test
```

Run backend integration tests against the local PostgreSQL stack:

```bash
make test-integration
```

Run linters:

```bash
make lint
```

Run E2E tests:

```bash
make e2e
```

## Contributing

- Branch names: use short, descriptive branches such as `feature/...`, `fix/...`, or `chore/...`.
- Commit messages: keep them imperative and specific, for example `fix ai routine save rate limit`.
- Pull requests: include a concise summary, the relevant test output, and screenshots or API examples when the change affects UI or payloads.
- Before opening a PR, run the relevant backend, frontend, and integration tests for the area you changed.

See the technical documentation in [`docs/index.md`](docs/index.md).

## Documentation

The detailed project documentation lives in [`docs/`](docs/index.md):

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
- [Contributing guide](docs/contributing.md)

## Useful Commands

```bash
make build-backend
make build-frontend
make start-app-snapshot
make down-app-snapshot
make reset-postgres-db
```

## Notes

- AI routine generation uses a preview-and-confirm flow.
- The AI service can create user-owned exercises automatically when Gemini proposes a valid exercise that does not already exist.
- AI routine endpoints are rate-limited in transport middleware.
