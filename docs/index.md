# LiteGym documentation

This directory contains the main project documentation for LiteGym. Each document is focused on a specific area of the system so the codebase is easier to understand, run, extend, and debug.

## Reading order

If you are new to the project, this order works well:

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
