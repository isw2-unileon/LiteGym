package testutil

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

// LoadIntegrationDBURL loads environment variables from local env files and returns
// the database URL to use for integration tests.
//
// Safety rule: integration tests only run against TEST_DB_URL and the target
// database must be named "test_db".
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

	testDBURL := os.Getenv("TEST_DB_URL")
	if testDBURL == "" {
		t.Fatal("TEST_DB_URL is not set")
	}

	parsed, err := url.Parse(testDBURL)
	if err != nil {
		t.Fatalf("TEST_DB_URL is invalid: %v", err)
	}

	dbName := strings.TrimPrefix(parsed.Path, "/")
	if dbName != "test_db" {
		t.Fatal(fmt.Sprintf("TEST_DB_URL must point to test_db, got %q", dbName))
	}

	return testDBURL
}
