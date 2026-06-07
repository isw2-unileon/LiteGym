# ADR-005: Postgres ENUM types for closed value sets

## Status

Accepted

## Date

12-05-2026

## Context

Several columns hold a small, fixed set of values: user roles, friendship status,
ticket status, and routine type. We needed a way to constrain these values at the
database level. The alternative considered was a plain `VARCHAR` with a `CHECK`
constraint.

## Decision

Model closed value sets as PostgreSQL `ENUM` types in the schema.

Current enums in `postgress-local/initdb/01-schema.sql`:

- `user_role`
- `ticket_status`
- `routine_type` (`'Fuerza'`, `'Movilidad'`, `'Resistencia'`, `'Sin clasificar'`)

This decision originally also covered `friendship_status`, which was removed together
with the `friendships` table when the social feature was dropped. The enum approach
itself is unchanged.

The frontend reads and writes these values 1:1 with the column (for example, the
routine filter keys equal the `routine_type` enum values).

## Consequences

- The database enforces valid values and they are self-documenting in the schema.
- Adding a value is an explicit `ALTER TYPE`, which keeps the value set intentional.
- Enum changes are slightly less flexible than free-text columns, accepted in
  exchange for stronger guarantees.
