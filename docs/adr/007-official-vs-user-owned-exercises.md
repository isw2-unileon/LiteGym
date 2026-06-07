# ADR-007: Official versus user-owned exercises with soft delete

## Status

Accepted

## Date

26-05-2026

## Context

The exercise catalog mixes curated, shared content with exercises that individual
users (or the AI save flow) create. We need to distinguish the two, enforce naming
rules, restrict who can manage official content, and avoid breaking historical
references when an exercise is removed.

## Decision

Model ownership and lifecycle directly on the `exercises` table.

- `is_official` plus `owner_user_id` express ownership: official exercises have
  `owner_user_id = NULL`; private exercises must have a non-null `owner_user_id`.
- Uniqueness is enforced per scope: unique official names by lowercased name, and
  unique private names per owner.
- Visibility and authorization rules follow ownership: private exercises render only
  to their owner (and admins), and only admins manage official exercises.
- Removal uses a `deleted_at` soft-delete pattern in business logic instead of hard
  deletes, so historical routines and workouts keep their references.

## Consequences

- The AI save flow can create personalized private exercises without polluting the
  official catalog (see [ADR-008](008-ai-routine-generation.md)).
- Listing and authorization must consistently filter by ownership and `deleted_at`.
- Soft delete keeps history intact but means queries must exclude deleted rows and
  the table accumulates inactive entries over time.
