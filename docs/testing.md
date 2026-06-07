# Testing guide

LiteGym uses a few testing layers so backend logic, frontend behavior, persistence, and browser-level flows can be checked independently.

## Test layers

- Go unit tests
- Go integration tests against PostgreSQL
- frontend component and page tests with Vitest
- browser-level E2E tests with Playwright

## Backend tests

Backend tests are spread across:

- `backend/internal/service/*_test.go`
- `backend/internal/repository/*_test.go`
- `backend/internal/repository/*_integration_test.go`
- `backend/internal/transport/handlers/*_test.go`
- `backend/internal/transport/router_test.go`

### What backend tests usually cover

- service validation and business rules
- repository query correctness
- handler status-code and payload behavior
- auth middleware behavior
- AI routine generation flow

## Integration database

Backend integration tests rely on the local PostgreSQL stack and helper utilities in:

- `backend/internal/testutil/integration_db.go`

Typical workflow:

```bash
make start-postgres-db
make test-integration
```

Some tests also use the general Go test command directly when targeting specific packages.

## Frontend tests

Frontend tests use Vitest and Testing Library.

Important files:

- `frontend/vitest.config.ts`
- `frontend/src/test/setup.ts`

Examples of tested pages:

- login
- dashboard
- exercises
- profile
- admin
- support
- routines

Run them with:

```bash
cd frontend
npm run test
```

Or for a single file:

```bash
cd frontend
npm test -- --run src/pages/UserRoutinesPage.test.tsx
```

## E2E tests

The E2E suite is still a work in progress. Playwright configuration lives in:

- `e2e/playwright.config.ts`

Current test (the suite currently ships only this one):

- `e2e/tests/health.spec.ts`

Run with:

```bash
make e2e
```

Make sure the backend and frontend are already running before you execute E2E tests.

## Useful commands

### All main tests

```bash
make test
```

### Backend integration only

```bash
make test-integration
```

### Linting

```bash
make lint
```

## AI-specific testing notes

The AI feature has three main kinds of coverage:

- service-level prompt and behavior tests
- HTTP integration tests for generate/save flows
- frontend preview/save flow tests

The real Gemini integration test is intentionally cautious because external provider limits are unstable:

- it requires credentials
- it can skip when the key is missing
- it can skip when Gemini quota is exhausted

That keeps the suite useful without making the whole run flaky because of third-party quotas.

## Suggested testing workflow for feature work

For backend-heavy changes:

1. run focused Go package tests
2. run integration tests when persistence changed
3. run frontend tests only if request or response contracts changed

For frontend-heavy changes:

1. run the page/component Vitest target
2. run `npm run build`
3. optionally validate manually in the browser

For AI flow changes:

1. run `go test ./backend/internal/service -count=1`
2. run `go test ./backend/internal/transport/handlers -count=1`
3. run `npm test -- --run src/pages/UserRoutinesPage.test.tsx`
