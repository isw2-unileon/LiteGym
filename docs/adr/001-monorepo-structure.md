# ADR-001: Monorepo with Go backend and React frontend

## Status

Accepted

## Date

26-02-2026

## Context

We need a repository structure that supports developing a Go API backend and a React frontend in the same codebase, with independent CI pipelines and clear boundaries between the two.

## Decision

Use a monorepo with top-level `backend/` and `frontend/` directories. The Go module is declared once via a `go.mod` at the repository root (covering `backend/`), while the frontend manages its own dependencies through `package.json`. Each side has a CI workflow with path filters so changes to one don't trigger the other.

## Consequences

- Atomic commits can span both backend and frontend.
- CI workflows run independently based on path filters, keeping feedback fast.
- Developers need both Go and Node.js toolchains installed locally.
