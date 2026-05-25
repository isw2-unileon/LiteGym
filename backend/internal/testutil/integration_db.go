package testutil

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const integrationDBAdvisoryLockKey int64 = 20260424

// LoadIntegrationDBURL loads environment variables from local env files and returns
// the database URL to use for integration tests.
//
// Safety rule: integration tests only run against TEST_DB_URL and the target
// database must be named "test_db". If TEST_DB_URL is absent, DATABASE_URL is
// used as a source to derive it. When the URL points to the docker-compose
// service host "postgres", it is rewritten to "localhost" so tests can run
// from the host shell against the published port.
func LoadIntegrationDBURL(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if ok {
		helperDir := filepath.Dir(filename)
		repoRoot := filepath.Clean(filepath.Join(helperDir, "..", "..", ".."))
		_ = godotenv.Load(
			filepath.Join(repoRoot, ".env.local"),
			filepath.Join(repoRoot, ".env"),
			filepath.Join(repoRoot, "backend", ".env"),
		)
	} else {
		_ = godotenv.Load(".env.local", ".env", "backend/.env")
	}

	testDBURL := strings.TrimSpace(os.Getenv("TEST_DB_URL"))
	if testDBURL == "" {
		testDBURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if testDBURL == "" {
		skipIntegrationTest(t, "TEST_DB_URL/DATABASE_URL not set. To run them, start Postgres with `make start-postgres-db` and run `make test-integration`.")
	}

	parsed, err := url.Parse(testDBURL)
	if err != nil {
		t.Fatalf("integration database URL is invalid: %v", err)
	}

	if parsed.Hostname() == "postgres" {
		port := parsed.Port()
		if port == "" {
			port = "5432"
		}
		parsed.Host = "localhost:" + port
	}

	dbName := strings.TrimPrefix(parsed.Path, "/")
	if dbName != "test_db" {
		t.Fatalf("integration database URL must point to test_db, got %q. Use TEST_DB_URL=postgres://test_user:test_password@localhost:5432/test_db?sslmode=disable", dbName)
	}

	resolvedURL := parsed.String()
	_ = os.Setenv("TEST_DB_URL", resolvedURL)

	return resolvedURL
}

// NewIntegrationTestPool returns a pgx pool for integration tests and holds a
// session-level advisory lock for the full duration of the test so packages do
// not mutate the shared test database concurrently.
func NewIntegrationTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	testDBURL := LoadIntegrationDBURL(t)

	db, err := pgxpool.New(t.Context(), testDBURL)
	if err != nil {
		t.Fatalf("error conectando a la base de test: %v", err)
	}

	if err := db.Ping(t.Context()); err != nil {
		db.Close()
		if canSkipIntegrationDBError(err) {
			skipIntegrationTest(t, "test database is not reachable (%v). Start it with `make start-postgres-db`, then run `make test-integration`.", err)
		}
		t.Fatalf("error haciendo ping a la base de test: %v", err)
	}

	lockConn, err := db.Acquire(t.Context())
	if err != nil {
		db.Close()
		t.Fatalf("error acquiring connection for integration lock: %v", err)
	}

	if _, err := lockConn.Exec(t.Context(), "SELECT pg_advisory_lock($1)", integrationDBAdvisoryLockKey); err != nil {
		lockConn.Release()
		db.Close()
		t.Fatalf("error acquiring integration lock: %v", err)
	}

	t.Cleanup(func() {
		_, _ = lockConn.Exec(t.Context(), "SELECT pg_advisory_unlock($1)", integrationDBAdvisoryLockKey)
		lockConn.Release()
		db.Close()
	})

	return db
}

func canSkipIntegrationDBError(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	return errors.Is(err, pgx.ErrNoRows) ||
		strings.Contains(message, "connection refused") ||
		strings.Contains(message, "socket: operation not permitted") ||
		strings.Contains(message, "dial error") ||
		strings.Contains(message, "timeout")
}

func skipIntegrationTest(t *testing.T, format string, args ...any) {
	t.Helper()

	message := fmt.Sprintf(format, args...)
	if strings.EqualFold(strings.TrimSpace(os.Getenv("CI")), "true") {
		t.Fatalf("integration test cannot be skipped in CI: %s", message)
	}

	t.Logf("warning: skipping integration test: %s", message)
	t.Skipf("skipping integration test: %s", message)
}
