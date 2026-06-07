# Why a Monorepo?

LiteGym keeps backend and frontend code in the same repository on purpose. That makes it easier to change both sides together without juggling separate repos.

## Benefits

- **Atomic changes** -- A single PR can update the API and the UI together.
- **Shared CI/CD** -- One place to define workflows, with path filters to keep them fast.
- **Simpler coordination** -- No cross-repo version pinning or release synchronization.
- **Easier onboarding** -- Clone once, run `make install`, and you have everything.

## Trade-offs

- Both toolchains (Go + Node.js) are required locally.
- CI must use path filters to avoid running everything on every change.
- Repository size grows over time (mitigated by `.gitignore` and clean boundaries).

## How It Works Here

```text
go.mod      → single Go module at the repository root (covers backend/)
frontend/   → Node project with its own package.json
e2e/        → Separate Node project for Playwright tests (work in progress)
```

The Go module is declared once at the repository root and contains the backend under
`backend/`. The frontend and e2e suites are self-contained Node projects. The root
`Makefile` provides unified commands across all of them.

## Further Reading

- [Monorepo vs Polyrepo](https://monorepo.tools/) -- comparison of tools and strategies
- [Google's Monorepo](https://research.google/pubs/pub45424/) -- why Google uses a single repository
- [Monorepos: Please do!](https://medium.com/@adamhjk/monorepo-please-do-3657e08a4b70) -- practical argument for monorepos
- [ADR-001](adr/001-monorepo-structure.md) -- our architectural decision record
