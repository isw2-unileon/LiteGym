package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockDB struct {
	pingErr error
}

func (m *mockDB) Ping(ctx context.Context) error {
	return m.pingErr
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := SetupRouter(&mockDB{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected status 'ok', got %q", body["status"])
	}
}

func TestHelloEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := SetupRouter(&mockDB{})

	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["message"] != "Hello from the API" {
		t.Fatalf("expected message 'Hello from the API', got %q", body["message"])
	}
}

func TestDatabaseHealthEndpointOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := SetupRouter(&mockDB{})

	req := httptest.NewRequest(http.MethodGet, "/api/db/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected status 'ok', got %q", body["status"])
	}

	if body["service"] != "database" {
		t.Fatalf("expected service 'database', got %q", body["service"])
	}
}

func TestDatabaseHealthEndpointError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := SetupRouter(&mockDB{
		pingErr: errors.New("database unavailable"),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/db/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["status"] != "error" {
		t.Fatalf("expected status 'error', got %q", body["status"])
	}

	if body["service"] != "database" {
		t.Fatalf("expected service 'database', got %q", body["service"])
	}

	if body["error"] != "database unavailable" {
		t.Fatalf("expected error 'database unavailable', got %q", body["error"])
	}
}