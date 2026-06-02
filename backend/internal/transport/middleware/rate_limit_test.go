package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func TestRateLimitStoreCleanupRemovesStaleEntries(t *testing.T) {
	store := newRateLimitStore(10*time.Millisecond, time.Nanosecond)
	base := time.Now()

	store.limiterFor("auth_login|ip:203.0.113.1", rate.Every(time.Second), 1, base)
	store.limiterFor("auth_login|ip:203.0.113.2", rate.Every(time.Second), 1, base.Add(20*time.Millisecond))

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, ok := store.entries["auth_login|ip:203.0.113.1"]; ok {
		t.Fatalf("expected stale entry to be removed from the store")
	}

	if _, ok := store.entries["auth_login|ip:203.0.113.2"]; !ok {
		t.Fatalf("expected recent entry to be kept in the store")
	}
}

func TestRateLimiter_ProfileAI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewRateLimiter()

	r := gin.New()
	r.Use(limiter.ProfileAI())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Burst is 1, limit is 1 per hour.
	// So first request should succeed, second should fail.

	// Request 1: Should be 200 OK
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	// mock remote IP so rate limit uses IP fallback since auth context is absent
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for first request, got %d", w.Code)
	}

	// Request 2: Should be 429 Too Many Requests
	req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429 for second request, got %d", w2.Code)
	}
}
