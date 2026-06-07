# ADR-006: Separate planned routine sets from performed workout sets

## Status

Accepted

## Date

20-05-2026

## Context

A routine prescribes what should be done (target reps, load, RIR, rest), while a
workout records what was actually done. Storing both in one table would force a
single shape to mean two different things and make prescription-versus-execution
comparisons awkward.

## Decision

Keep prescription and execution in separate tables.

- `routine_exercises` and `routine_exercise_sets` hold the planned prescription for a
  routine: rep ranges, rep text, target load, duration, distance, RIR, rest, and
  notes.
- `workout_sessions`, `workout_exercises`, and `workout_sets` hold concrete events.
  `workout_sets` supports both target values and actual values.
- A workout session can be planned, performed, or tied to a routine; "filling" a
  planned workout completes its sets and marks it as completed.

## Consequences

- Prescription versus execution is easy to compare and report on.
- The "fill workout" flow has a natural home: copy the plan into a session, then
  record actuals.
- There is some duplication of target fields across the two table groups, accepted as
  the cost of separating intent from outcome.
- Reads that need both plan and performance must join across the two groups.
