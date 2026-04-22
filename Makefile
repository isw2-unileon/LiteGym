.PHONY: install run-backend run-frontend build-backend build-frontend test lint e2e start-app-snapshot down-app-snapshot delete-app-snapshot start-postges-db stop-postges-db delete-postgres-db

COMPOSE ?= docker compose
ifeq ($(shell command -v docker >/dev/null 2>&1; echo $$?),1)
COMPOSE := podman compose
endif

COMPOSE ?= docker compose
ifeq ($(shell command -v docker >/dev/null 2>&1; echo $$?),1)
COMPOSE := podman compose
endif

## Install all dependencies
install:
	go install github.com/air-verse/air@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	go mod download
	go get github.com/jackc/pgx/v5
	cd frontend && npm ci
	cd e2e && npm ci

## Run backend with hot reload
run-backend:
	$(shell go env GOPATH)/bin/air -c backend/.air.toml

## Run frontend dev server
run-frontend:
	cd frontend && npm run dev

## Build backend binary
build-backend:
	go build -o backend/bin/server ./backend/cmd/server

## Build frontend for production
build-frontend:
	cd frontend && npm run build

#
# App snapshot (compose.yaml) management
# These targets operate the compose.yaml found at the repository root.
#
start-app-snapshot:
	@echo "Starting application stack using compose.yaml..."
	$(COMPOSE) -f compose.yaml up -d

down-app-snapshot:
	@echo "Stopping application stack defined in compose.yaml..."
	$(COMPOSE) -f compose.yaml down

delete-app-snapshot:
	@echo "Deleting application stack (volumes & local images) defined in compose.yaml..."
	$(COMPOSE) -f compose.yaml down --volumes --rmi local --remove-orphans

#
# Postgres-local compose management
# These targets operate the docker-compose.yml inside postgress-local directory.
#
start-postges-db:
	@echo "Starting postgres-local stack..."
	$(COMPOSE) -f postgress-local/docker-compose.yml up -d

stop-postges-db:
	@echo "Stopping postgres-local stack..."
	$(COMPOSE) -f postgress-local/docker-compose.yml down

delete-postgres-db:
	@echo "Deleting postgres-local stack (volumes & local images)..."
	$(COMPOSE) -f postgress-local/docker-compose.yml down --volumes --rmi local --remove-orphans

## Run all tests
test:
	go test -v -race ./...
	cd frontend && npm run test

## Run linters
lint:
	$(shell go env GOPATH)/bin/golangci-lint run
	cd frontend && npm run lint

## Run E2E tests (requires backend + frontend running)
e2e:
	cd e2e && npx playwright test
