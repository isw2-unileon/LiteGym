package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func TestRateLimitMiddlewareReusesLimiterPerCategoryAndIP(t *testing.T) {
	limiter := NewRateLimitMiddleware()

	first := limiter.limiterFor("login", "203.0.113.1", rate.NewLimiter(rate.Every(time.Second), 1))
	second := limiter.limiterFor("login", "203.0.113.1", rate.NewLimiter(rate.Every(time.Second), 1))
	otherIP := limiter.limiterFor("login", "203.0.113.2", rate.NewLimiter(rate.Every(time.Second), 1))

	if first != second {
		t.Fatalf("expected limiter to be reused for the same category and IP")
	}

	if first == otherIP {
		t.Fatalf("expected different IPs to get different limiters")
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

func TestRateLimiter_AIIsStrict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewRateLimiter()

	r := gin.New()
	r.Use(limiter.AI())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for first request, got %d", w.Code)
	}

	req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.2:1234"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429 for second request, got %d", w2.Code)
	}
}
