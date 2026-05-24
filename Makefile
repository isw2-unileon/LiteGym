.PHONY: install run-backend run-frontend build-backend build-frontend test test-integration lint e2e \
	start-app-snapshot down-app-snapshot delete-app-snapshot \
	start-postgres-db stop-postgres-db down-postgres-db delete-postgres-db reset-postgres-db logs-postgres-db

COMPOSE ?= docker compose
ifeq ($(shell command -v docker >/dev/null 2>&1; echo $$?),1)
COMPOSE := podman compose
endif

TEST_DB_URL ?= postgres://test_user:test_password@postgres:5432/test_db?sslmode=disable
POSTGRES_COMPOSE_FILE := postgress-local/docker-compose.yml

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
# App snapshot compose management
# These targets operate the compose.yaml found at the repository root.
#
start-app-snapshot:
	@echo "Starting application stack using compose.yaml..."
	$(COMPOSE) -f compose.yaml up -d

down-app-snapshot:
	@echo "Stopping application stack defined in compose.yaml..."
	$(COMPOSE) -f compose.yaml down

delete-app-snapshot:
	@echo "Deleting application stack, volumes and local images defined in compose.yaml..."
	$(COMPOSE) -f compose.yaml down --volumes --rmi local --remove-orphans

#
# Postgres local compose management
# These targets operate the docker-compose.yml inside postgress-local directory.
#
start-postgres-db:
	@echo "Starting postgres-local stack..."
	$(COMPOSE) -f $(POSTGRES_COMPOSE_FILE) up -d --build

stop-postgres-db:
	@echo "Stopping postgres-local stack without removing containers or volumes..."
	$(COMPOSE) -f $(POSTGRES_COMPOSE_FILE) stop

down-postgres-db:
	@echo "Stopping and removing postgres-local containers without deleting volumes..."
	$(COMPOSE) -f $(POSTGRES_COMPOSE_FILE) down --remove-orphans

delete-postgres-db:
	@echo "Deleting postgres-local stack, volumes and local images..."
	$(COMPOSE) -f $(POSTGRES_COMPOSE_FILE) down --volumes --rmi local --remove-orphans

reset-postgres-db: delete-postgres-db start-postgres-db

logs-postgres-db:
	$(COMPOSE) -f $(POSTGRES_COMPOSE_FILE) logs -f

## Run all tests
test:
	go test -v -race ./...
	cd frontend && npm run test

## Run backend integration tests against local Postgres
test-integration:
	TEST_DB_URL="$(TEST_DB_URL)" go test -v ./backend/internal/repository ./backend/internal/service

## Run linters
lint:
	$(shell go env GOPATH)/bin/golangci-lint run ./backend/...
	cd frontend && npm run lint

## Run E2E tests, requires backend and frontend running
e2e:
	cd e2e && npx playwright test
