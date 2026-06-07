# ADR-008: AI routine generation via Gemini with preview-and-confirm

## Status

Accepted

## Date

26-05-2026

## Context

LiteGym offers AI-assisted routine creation. Generated content can be wrong,
incomplete, or reference exercises the catalog does not know about. We did not want
to persist model output blindly, nor reject a useful routine just because an exercise
id does not match.

## Decision

Generate routines with the Gemini API behind a two-step preview-and-confirm flow, and
resolve exercises tolerantly on save.

- `POST /api/routines/ai/generate` builds a structured JSON prompt (objective,
  duration, target muscle groups, mandatory exercises, user notes, a compact
  `user_context`, a filtered `exercise_catalog`, and an `output_contract`), calls
  Gemini, normalizes the JSON, and returns a preview. Nothing is persisted yet.
- `POST /api/routines/ai/save` persists only after the user confirms the preview.
- Exercise resolution on save is tolerant: trust the returned `exercise_id` only if
  it exists; otherwise match by normalized name, then narrow by muscle group and
  type; as a last resort create a new private exercise owned by the user (see
  [ADR-007](007-official-vs-user-owned-exercises.md)).
- Configuration: `GEMINI_API_KEY` (missing key → `503`) and `GEMINI_MODEL` (empty →
  fallback `gemini-2.5-flash`).
- Errors map deliberately: invalid input → `400`, rate limit → `429`, provider
  unavailable or malformed → `503`, unexpected → `500`.
- Per-user rate limiting (2 requests per hour, `aiRoutineRateLimit = 2` over
  `aiRoutineRateWindow = time.Hour`) is counted from `ai_routine_generation_logs`; an
  exceeded limit returns `429`. The provider may also enforce its own external quotas.

## Consequences

- The user always reviews AI output before it is saved, keeping bad routines out of
  the database.
- Routines remain usable even when the model invents or mislabels exercises, because
  resolution degrades gracefully to a created private exercise.
- The prompt sends compact context to stay token-friendly, trading some richness for
  cost and latency.
- Generation depends on an external provider and its quotas; failures are surfaced as
  specific status codes rather than hidden.
