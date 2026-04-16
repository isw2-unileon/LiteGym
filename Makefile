.PHONY: install run-backend run-frontend build-backend build-frontend test lint e2e start-test-db stop-test-db rm-test-db init-test-db test-integration test-integration-clean

# Docker runtime command
DOCKER ?= docker

# Test DB container config
DB_CONTAINER ?= postgres-test
DB_USER ?= test_user
DB_PASS ?= test_password
DB_NAME ?= test_db
DB_PORT ?= 5432
SCHEMA_FILE ?= gym/sql/schema.sql
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
	-$(DOCKER) rm -f $(DB_CONTAINER) 2>/dev/null || true
	$(DOCKER) run -d --name $(DB_CONTAINER) -e POSTGRES_USER=$(DB_USER) -e POSTGRES_PASSWORD=$(DB_PASS) -e POSTGRES_DB=$(DB_NAME) -p $(DB_PORT):5432 $(POSTGRES_IMAGE)

## Stop the test Postgres container
stop-test-db:
	@echo "Stopping test database container '$(DB_CONTAINER)'..."
	-$(DOCKER) stop $(DB_CONTAINER) 2>/dev/null || true

## Remove the test Postgres container
rm-test-db:
	@echo "Removing test database container '$(DB_CONTAINER)'..."
	-$(DOCKER) rm -f $(DB_CONTAINER) 2>/dev/null || true

## Initialize the test DB by copying and executing schema.sql into the container
init-test-db:
	@echo "Waiting for postgres to be ready..."
	@sh -c 'for i in 1 2 3 4 5 6 7 8 9 10; do $(DOCKER) exec $(DB_CONTAINER) pg_isready -U $(DB_USER) -d $(DB_NAME) >/dev/null 2>&1 && break || sleep 1; done'
	@echo "Copying schema file $(SCHEMA_FILE) into container $(DB_CONTAINER)..."
	@$(DOCKER) cp $(SCHEMA_FILE) $(DB_CONTAINER):/schema.sql
	@echo "Applying schema to database $(DB_NAME)..."
	@$(DOCKER) exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -f /schema.sql

## Run integration tests for repository package (starts DB, inits schema, runs tests)
test-integration: start-test-db init-test-db
	@echo "Running repository integration tests..."
	go test ./backend/internal/repository -v

## Run integration tests and then cleanup (stop + remove DB container)
test-integration-clean: test-integration
	@echo "Cleaning up test DB container..."
	-$(DOCKER) stop $(DB_CONTAINER) 2>/dev/null || true
	-$(DOCKER) rm -f $(DB_CONTAINER) 2>/dev/null || true

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
