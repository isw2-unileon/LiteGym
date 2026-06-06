# Contributing

This guide is for external contributors who want to work on LiteGym locally and open pull requests.

## Branches

- Use short, descriptive branch names.
- Recommended prefixes: `feature/`, `fix/`, `chore/`, `docs/`.
- Keep the scope narrow when possible, especially for backend changes that touch routes, services, or persistence.

## Commits

- Use imperative commit messages.
- Keep each commit focused on one change.
- Good examples: `fix ai routine save rate limit`, `add routine builder`, `update docs index`.

## Pull Requests

- Include a short description of the change.
- Mention the tests you ran and their result.
- Add screenshots for frontend changes and example payloads for API changes when helpful.
- Link related issues or context when the PR depends on a specific bug or task.

## Local Checks

- Run the relevant Go tests for backend changes.
- Run the frontend tests when touching React code.
- Run `make lint` before opening a PR if the change touches shared code.
- Run `make test` or the closest relevant subset for the area you changed.

## Review Expectations

- Keep the diff focused.
- Avoid unrelated refactors in the same PR unless they reduce risk.
- Prefer the existing code style and module boundaries in the repo.

