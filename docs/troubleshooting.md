# Troubleshooting

This guide collects the local issues you are most likely to hit while running or extending LiteGym.

## Backend does not start

Check:

- `DATABASE_URL` is set correctly
- PostgreSQL is running
- the configured port is free

Helpful commands:

```bash
make start-postgres-db
make logs-postgres-db
```

## Frontend loads but authenticated pages redirect to login

Check:

- the backend is running
- login succeeded and set the auth cookie
- browser requests are using `credentials: "include"`
- the token is not stale relative to the current database contents

If the database was reset but the browser still has an old cookie, log in again so the token matches an existing user.

## `/api/routines/ai/generate` returns `503`

Likely causes:

- missing `GEMINI_API_KEY`
- unsupported or unavailable `GEMINI_MODEL`
- Gemini free-tier quota exhaustion
- malformed provider response

Useful places to inspect:

- backend logs
- `backend/internal/service/routine_ai_service.go`
- `backend/internal/config/config.go`

## Gemini responds with `429`

This is usually a provider-side quota issue, not an application bug.

Check:

- which `GEMINI_MODEL` is configured
- whether the model is available for your plan
- whether your free-tier quota has been exhausted

If a real integration test fails because Gemini quota is exhausted, the handler integration test is already designed to skip that case.

## Frontend shows proxy or socket errors during AI generation

Check:

- backend is actually running on `127.0.0.1:8080`
- Vite proxy configuration in `frontend/vite.config.ts`
- backend `WriteTimeout` is large enough for slow AI responses

The project already includes proxy and timeout tuning for slower AI requests.

## Seed changes are not visible in the database

This is expected if the Postgres volume already existed.

The init scripts only replay when the volume is recreated. Use:

```bash
make reset-postgres-db
```

## Integration tests fail because the database is unreachable

Check:

- local Postgres stack is running on port `5432`
- `TEST_DB_URL` points to the correct instance

Common command:

```bash
make start-postgres-db
```

## Routine AI save fails because an exercise does not exist

The current save flow is designed to be forgiving:

- reuse existing exercise id when valid
- search by normalized name and domain fields
- create a new private exercise if necessary

If this still fails, inspect:

- generated exercise payload
- exercise normalization logic
- exercise creation validation rules

## Frontend tests fail after UI text changes

Many frontend tests use text and role queries. After changing labels, button text, or placeholder text:

- update the related test expectations
- rerun the targeted Vitest file

Useful command:

```bash
cd frontend
npm test -- --run src/pages/UserRoutinesPage.test.tsx
```
