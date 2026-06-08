# Configuration

This page shows how LiteGym loads configuration and which environment variables actually matter.

## Backend configuration loading

Backend configuration is loaded in:

- `backend/internal/config/config.go`

The loader checks environment files in this order:

When the process is launched from the repository root:

1. `.env.local`
2. `.env`
3. `backend/.env`

When the process is launched from inside the `backend/` directory:

1. `../.env.local`
2. `../.env`
3. `.env`

After that, values are read through `os.Getenv(...)` wrappers.

## Backend variables

The backend `Config` struct includes:

- `PORT`
- `GIN_MODE`
- `CORS_ALLOW_ORIGIN`
- `DATABASE_URL`
- `JWT_SECRET`
- `AUTH_COOKIE_NAME`
- `AUTH_COOKIE_SECURE`
- `AUTH_TOKEN_TTL`
- `GEMINI_API_KEY`
- `GEMINI_MODEL`

## Backend variable reference

### `PORT`

- default: `8080`
- backend HTTP listening port

### `GIN_MODE`

- default: `debug`
- controls Gin runtime mode

### `CORS_ALLOW_ORIGIN`

- default: `*`
- comma-separated allowed origins for CORS

### `DATABASE_URL`

- required for a usable backend
- PostgreSQL connection string used by pgxpool

### `JWT_SECRET`

- default: `dev-secret-change-me`
- secret used by the token service

### `AUTH_COOKIE_NAME`

- default: `auth_token`
- cookie name used for session auth

### `AUTH_COOKIE_SECURE`

- default: `false`
- whether the auth cookie requires HTTPS

### `AUTH_TOKEN_TTL`

- default: `24h`
- token expiration duration

### `GEMINI_API_KEY`

- default: empty
- required for AI routine generation

### `GEMINI_MODEL`

- config default (`config.go`): `gemini-2.5-flash`
- service-level fallback when an empty value reaches the AI service: `gemini-2.5-flash`

If you want deterministic behavior, set `GEMINI_MODEL` explicitly rather than relying
on fallback paths.

## Example backend environment

The repository root includes:

- `.env.example`

Its current values are:

```env
PORT=8080
GIN_MODE=debug
CORS_ALLOW_ORIGIN=http://localhost:5173
DATABASE_URL=
JWT_SECRET=change-this-in-production
AUTH_COOKIE_NAME=auth_token
AUTH_COOKIE_SECURE=false
AUTH_TOKEN_TTL=24h
GEMINI_API_KEY=
GEMINI_MODEL=gemini-2.5-flash
```

## Frontend configuration

Frontend configuration is read by Vite through `import.meta.env`.

The main example in the codebase is:

- `frontend/src/lib/api.ts`

Important rule:

- only variables prefixed with `VITE_` are exposed to the browser bundle

If `VITE_API_BASE_URL` is not set, the frontend falls back to relative URLs, which works well with the dev proxy in `frontend/vite.config.ts`.

## Compose configuration

Compose files also reference environment files.

### `compose.yaml`

Uses:

- `.env.local`
- `backend/.env`
- `frontend/.env`

### `postgress-local/docker-compose.yml`

It uses inline Postgres credentials for the local test database container.

## Practical recommendations

- keep shared local overrides in `.env.local`
- keep backend-specific runtime variables in `backend/.env`
- avoid relying on implicit defaults for `DATABASE_URL` and `GEMINI_MODEL`
- when debugging AI issues, confirm both `GEMINI_API_KEY` and `GEMINI_MODEL`
