# ADR-002: Layered backend with transport, service, and repository

## Status

Accepted

## Date

26-03-2026

## Context

The Go backend handles HTTP concerns, business rules, and persistence for several
domains (users, exercises, routines, workouts, tickets, AI generation). Without a
clear internal structure, handlers tend to query the database directly and business
rules leak into HTTP code, which makes the system hard to test and change.

## Decision

Split the backend into explicit layers under `backend/internal/`:

```text
transport/ -> service/ -> repository/ -> database
```

- `transport`: Gin router, middleware, and HTTP handlers. Handlers only read
  request input, call a service, and map service errors to status codes.
- `service`: business logic, validation, orchestration, and cross-repository flows.
- `repository`: SQL-backed persistence, one repository per domain.
- `model`: domain and response structs shared inside the backend.
- `config`: environment-driven runtime configuration.

Dependencies point inward only: handlers depend on services, services depend on
repositories. Wiring happens once at startup in `backend/cmd/server/main.go`
(config → pool → repositories → services → handlers → router).

## Consequences

- Each layer can be unit-tested in isolation; services are tested without HTTP and
  repositories have integration tests against PostgreSQL.
- Handlers stay thin and never touch the database directly.
- Adding a feature follows a predictable path: model → repository → service →
  handler → route → tests.
- The extra indirection adds boilerplate for simple endpoints, which is an accepted
  trade-off for consistency.
