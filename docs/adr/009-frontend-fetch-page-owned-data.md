# ADR-009: Frontend uses direct fetch with page-owned data contracts

## Status

Accepted

## Date

26-05-2026

## Context

The React frontend talks to the Go API across several route-driven pages
(dashboard, exercises, routines, profile, admin, support). We had to decide how much
shared infrastructure to build for data access and state: a centralized API client, a
global state/data-fetching library, or something lighter.

## Decision

Keep data access lightweight and page-centric.

- Pages call `fetch(...)` directly with cookie credentials rather than going through
  a centralized API client layer.
- URL construction is shared through `frontend/src/lib/api.ts`. If
  `VITE_API_BASE_URL` is unset, requests stay relative and rely on the Vite dev
  proxy.
- Each page owns the data contract it needs; response types live page-local when they
  are tightly bound to one page, and shared domain types stay in `frontend/src/types/`.
- Route-level state lives in the page; visual subparts are extracted into components
  (control inverted via props/callbacks rather than variant flags).

## Consequences

- Low ceremony: a page can be understood end to end without tracing a shared client
  or global store.
- No global data-fetching/state dependency to learn or maintain.
- Cross-page concerns (auth, base URL) are handled by convention (`credentials:
  "include"`, `lib/api.ts`) rather than enforced by a single client, so consistency
  relies on following the pattern.
- Data contracts can drift between pages; shared shapes must be deliberately promoted
  into `types/` when reused.
