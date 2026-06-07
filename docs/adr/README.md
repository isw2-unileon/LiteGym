# Architecture Decision Records

This directory records the significant architectural decisions for LiteGym. Each ADR
captures the context, the decision, and its consequences at the time it was made.

Use [`000-template.md`](000-template.md) as the starting point for a new ADR. Number
ADRs sequentially and never rewrite history: to change a decision, add a new ADR and
mark the old one `Superseded by` it.

## Index

| ADR | Title | Status | Date       |
|-----|-------|--------|------------|
| [001](001-monorepo-structure.md) | Monorepo with Go backend and React frontend | Accepted | 26-02-2026 |
| [002](002-layered-backend.md) | Layered backend with transport, service, and repository | Accepted | 26-03-2026 |
| [003](003-custom-hmac-cookie-sessions.md) | Cookie sessions with a custom HMAC-signed token | Accepted | 26-03-2026 |
| [004](004-postgresql-pgx-raw-sql.md) | PostgreSQL with pgx and hand-written SQL | Accepted | 26-03-2026 |
| [005](005-postgres-enums.md) | Postgres ENUM types for closed value sets | Accepted | 12-05-2026 |
| [006](006-planned-vs-performed-sets.md) | Separate planned routine sets from performed workout sets | Accepted | 20-05-2026 |
| [007](007-official-vs-user-owned-exercises.md) | Official versus user-owned exercises with soft delete | Accepted | 26-05-2026 |
| [008](008-ai-routine-generation.md) | AI routine generation via Gemini with preview-and-confirm | Accepted | 26-05-2026 |
| [009](009-frontend-fetch-page-owned-data.md) | Frontend uses direct fetch with page-owned data contracts | Accepted | 26-05-2026 |
