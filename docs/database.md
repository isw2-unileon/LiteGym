# Database guide

LiteGym uses PostgreSQL as its main persistence layer. The canonical local schema and demo dataset live in `postgress-local/initdb/`.

## Schema source

- `postgress-local/initdb/01-schema.sql`
- `postgress-local/initdb/02-seed.sql`

The local database image mounts these scripts into `/docker-entrypoint-initdb.d`, so they are executed when the database volume is created from scratch.

## Main schema design

The schema models:

- users and profiles
- body metrics
- exercises
- routines and routine exercise templates
- workouts and performed sets
- friendships and routine sharing
- support tickets
- AI generation logs

## Enumerated types

The schema defines these enums:

- `friendship_status`
- `ticket_status`
- `user_role`

## Core tables

### `users`

Stores:

- identity
- email
- password hash
- role
- active state

### `user_profiles`

One-to-one with users. Stores optional profile attributes such as:

- first name
- last name
- age
- weight
- height
- goal
- experience level

### `body_metrics`

Historical body measurement table per user.

### `exercises`

Central exercise catalog.

Important columns:

- `is_official`
- `owner_user_id`
- `deleted_at`

Important constraint:

- official exercises must not have an owner
- private exercises must have an owner

Important indexes:

- unique official exercise names by lowercased name
- unique private exercise names per owner

### `exercise_secondary_muscle_groups`

Stores additional muscle groups for an exercise.

### `routines`

User routine headers.

Important columns:

- `source`
- `is_predefined`
- `is_public`

### `routine_exercises`

Exercises that belong to a routine with explicit order and optional notes.

### `routine_exercise_sets`

Planned set prescriptions for each routine exercise.

Supported prescription data includes:

- rep ranges
- rep text
- target load
- target duration
- target distance
- RIR
- rest time
- notes

### `workout_sessions`

Concrete workout events. A session can be planned, performed, or tied to a routine.

### `workout_exercises`

Exercises executed during a workout session.

### `workout_sets`

Performed set records for workout exercises. This table supports both target values and actual values.

### `friendships`

Social relationship table between users.

### `shared_routines`

Stores explicit routine sharing entries.

### `support_tickets`

Stores support conversations in a lightweight ticket form.

### `ai_routine_generation_logs`

Stores timestamped AI generation usage entries per user. The backend uses this table for rate-limit counting logic.

## Relationship overview

High-level relationship map:

```text
users
  |-- user_profiles
  |-- body_metrics
  |-- routines
  |-- workout_sessions
  |-- support_tickets
  |-- exercises (private only via owner_user_id)
  `-- ai_routine_generation_logs

routines
  |-- routine_exercises
  `-- shared_routines

routine_exercises
  `-- routine_exercise_sets

workout_sessions
  `-- workout_exercises

workout_exercises
  `-- workout_sets
```

## Seed data

The seed file includes:

- demo users
- user profiles
- body metrics history
- official and private exercises
- sample friendships
- sample routines
- routine exercises and planned sets
- support tickets
- workout history

This is useful for:

- local UI exploration
- integration testing
- exercising dashboard and insights views

## Reset behavior

Important note:

- the init scripts only run when the Postgres volume is initialized
- editing the seed file does not automatically reseed an existing volume

If you want the current schema and seed to be replayed, use:

```bash
make reset-postgres-db
```

## Current design observations

- the schema separates planned routine sets from performed workout sets, which is a good fit for comparing prescription versus execution
- user-owned exercises enable personalized AI routine persistence without polluting the official catalog
- the `deleted_at` pattern is used on exercises instead of hard deletes from business logic
