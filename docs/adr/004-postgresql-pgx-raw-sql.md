# ADR-004: PostgreSQL with pgx and hand-written SQL

## Status

Accepted

## Date

26-03-2026

## Context

The backend needs a relational store for users, exercises, routines, workouts, and
related data with non-trivial queries (filtering, pagination, aggregates for the
dashboard, nested routine/workout structures). We had to decide between an ORM and
direct SQL access.

## Decision

Use PostgreSQL accessed through `pgx`, with SQL written by hand in the repository
layer instead of an ORM.

- Each domain has its own repository file under `backend/internal/repository/`.
- The local schema and demo data live in `postgress-local/initdb/01-schema.sql` and
  `02-seed.sql`, mounted into `/docker-entrypoint-initdb.d` so they run when the
  database volume is created from scratch.
- Schema changes are made by editing those scripts; an existing volume is replayed
  with `make reset-postgres-db` (init scripts only run on a fresh volume).

## Consequences

- Full control over SQL, which suits the dashboard aggregates and nested reads.
- No ORM abstraction to learn or fight; queries are explicit and reviewable.
- Schema and seed are versioned in the repo and easy to reproduce locally and in
  integration tests.
- Editing the seed does not reseed an existing volume; developers must reset the
  volume to pick up schema/seed changes.
- Repositories carry more manual mapping code than an ORM would, accepted for the
  control it provides.
