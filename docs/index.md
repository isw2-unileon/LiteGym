# LiteGym documentation

This folder holds the main project docs for LiteGym. Each file stays focused on one part of the system so it is easier to understand, run, extend, and debug.

## Reading order

If you are new to the project, this order is the easiest way in:

1. [Getting started](getting-started.md)
2. [Architecture](architecture.md)
3. [Configuration](configuration.md)
4. [Backend guide](backend.md)
5. [Authentication and authorization](authentication.md)
6. [Frontend guide](frontend.md)
7. [Database guide](database.md)
8. [AI integration](ai-integration.md)
9. [Testing guide](testing.md)
10. [Troubleshooting](troubleshooting.md)
11. [Contributing guide](contributing.md)

## Reference documents

- [API reference](api-reference.md)
- [Monorepo notes](monorepo.md)
- [Go notes](golang.md)
- [ADR directory](adr/)

## Document scope

- `getting-started.md`: setup, local execution, common commands
- `architecture.md`: system shape, runtime flow, layering, major modules
- `configuration.md`: environment variables and configuration loading behavior
- `backend.md`: backend structure, bootstrap, repositories, services, handlers
- `authentication.md`: login flow, token format, cookie model, role and session checks
- `frontend.md`: route structure, layout, main pages, data flow
- `api-reference.md`: HTTP endpoints and payload expectations
- `database.md`: schema, core tables, relationships, seed behavior
- `ai-integration.md`: Gemini integration, request flow, preview/save flow
- `testing.md`: unit, integration, frontend, and E2E testing strategy
- `troubleshooting.md`: common local failures and how to diagnose them
- `contributing.md`: branch, commit, PR, and local verification guidance
