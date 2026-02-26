# Monorepo Template: Go + React/Vite

A monorepo template for full-stack applications with a **Go** backend and a **React + TypeScript + Vite** frontend.

## Project Structure

```text
├── backend/              Go API server (Gin)
│   ├── cmd/server/       Entry point
│   └── internal/config/  Environment config
│
├── frontend/             React + TypeScript + Vite + Tailwind
│   └── src/
│
├── e2e/                  Playwright E2E tests
├── .github/workflows/    CI/CD pipelines
└── Makefile              Dev commands
```

## Prerequisites

- [Go](https://go.dev/dl/) 1.24+
- [Node.js](https://nodejs.org/) 22+

## Getting Started

```bash
make install

# Terminal 1
make run-backend    # port 8080

# Terminal 2
make run-frontend   # port 5173
```

The Vite dev server proxies `/api` requests to the backend.

## Commands

| Command              | Description                     |
|----------------------|---------------------------------|
| `make install`       | Install all dependencies        |
| `make run-backend`   | Backend with hot reload (Air)   |
| `make run-frontend`  | Frontend dev server (Vite)      |
| `make test`          | Run all tests                   |
| `make lint`          | Run all linters                 |
| `make e2e`           | Run Playwright E2E tests        |

## API

| Method | Path         | Description    |
|--------|------------- |----------------|
| `GET`  | `/health`    | Health check   |
| `GET`  | `/api/hello` | Sample endpoint|
