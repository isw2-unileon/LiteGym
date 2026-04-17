.PHONY: install run-backend run-frontend build-backend build-frontend test lint e2e start-test-db stop-test-db rm-test-db init-test-db test-integration test-integration-clean

# Podman runtime command
PODMAN ?= podman

# Test DB container config
DB_CONTAINER ?= postgres-test
DB_USER ?= test_user
DB_PASS ?= test_password
DB_NAME ?= test_db
DB_PORT ?= 5432
SCHEMA_FILE ?= sql/schema.sql
POSTGRES_IMAGE ?= postgres:15-alpine

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

## Start a test Postgres container
start-test-db:
	@echo "Starting test database container '$(DB_CONTAINER)'..."
	-$(PODMAN) rm -f $(DB_CONTAINER) 2>/dev/null || true
	$(PODMAN) run -d --name $(DB_CONTAINER) -e POSTGRES_USER=$(DB_USER) -e POSTGRES_PASSWORD=$(DB_PASS) -e POSTGRES_DB=$(DB_NAME) -p $(DB_PORT):5432 $(POSTGRES_IMAGE)

## Stop the test Postgres container
stop-test-db:
	@echo "Stopping test database container '$(DB_CONTAINER)'..."
	-$(PODMAN) stop $(DB_CONTAINER) 2>/dev/null || true

## Remove the test Postgres container
rm-test-db:
	@echo "Removing test database container '$(DB_CONTAINER)'..."
	-$(PODMAN) rm -f $(DB_CONTAINER) 2>/dev/null || true

## Initialize the test DB by copying and executing schema.sql into the container
init-test-db:
	@echo "Waiting for postgres to be ready..."
	@sh -c 'max=60; i=0; until [ $$i -ge $$max ]; do $(PODMAN) exec $(DB_CONTAINER) pg_isready -U $(DB_USER) -d $(DB_NAME) >/dev/null 2>&1 && break; i=$$((i+1)); echo "waiting for postgres ($$i/$$max)..." >&2; sleep 1; done; if [ $$i -ge $$max ]; then echo "Postgres did not become ready in $$max seconds" >&2; $(PODMAN) logs $(DB_CONTAINER) --tail 50 >&2 || true; exit 1; fi'
	@echo "Copying schema file $(SCHEMA_FILE) into container $(DB_CONTAINER)..."
	@$(PODMAN) cp $(SCHEMA_FILE) $(DB_CONTAINER):/schema.sql
	@echo "Applying schema to database $(DB_NAME)..."
	@$(PODMAN) exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -f /schema.sql

## Run integration tests for repository package (starts DB, inits schema, runs tests)
test-integration: start-test-db init-test-db
	@echo "Running repository integration tests..."
	go test ./backend/internal/repository -v

## Run integration tests for service package (starts DB, inits schema, runs tests)
test-integration-service: start-test-db init-test-db
	@echo "Running service integration tests..."
	go test ./backend/internal/service -v

## Run all integration tests (repository + service)
test-integration-all: start-test-db init-test-db
	@echo "Running all integration tests..."
	@$(MAKE) test-integration
	@$(MAKE) test-integration-service

## Run integration tests and then cleanup (stop + remove DB container)
test-integration-clean: test-integration-all
	@echo "Cleaning up test DB container..."
	-$(PODMAN) stop $(DB_CONTAINER) 2>/dev/null || true
	-$(PODMAN) rm -f $(DB_CONTAINER) 2>/dev/null || true

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
