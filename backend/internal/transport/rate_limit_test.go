package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

func newRateLimitTestEngine(handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/", handler, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return r
}

func performRateLimitRequest(t *testing.T, engine *gin.Engine, remoteAddr string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)
	return rec
}

func TestLoginRateLimitAllowsBurstThenBlocks(t *testing.T) {
	engine := newRateLimitTestEngine(middleware.NewRateLimitMiddleware().Login())

	for i := 0; i < 5; i++ {
		resp := performRateLimitRequest(t, engine, "1.2.3.4:1234")
		if resp.Code != http.StatusOK {
			t.Fatalf("expected request %d to pass, got %d", i+1, resp.Code)
		}
	}

	resp := performRateLimitRequest(t, engine, "1.2.3.4:1234")
	if resp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected login rate limit after burst, got %d", resp.Code)
	}

	resp = performRateLimitRequest(t, engine, "5.6.7.8:1234")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected different IP to get a fresh budget, got %d", resp.Code)
	}
}

func TestRegisterRateLimitAllowsBurstThenBlocks(t *testing.T) {
	engine := newRateLimitTestEngine(middleware.NewRateLimitMiddleware().Register())

	for i := 0; i < 5; i++ {
		resp := performRateLimitRequest(t, engine, "1.2.3.4:1234")
		if resp.Code != http.StatusOK {
			t.Fatalf("expected request %d to pass, got %d", i+1, resp.Code)
		}
	}

	resp := performRateLimitRequest(t, engine, "1.2.3.4:1234")
	if resp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected register rate limit after burst, got %d", resp.Code)
	}
}

func TestProtectedRateLimitAllowsBurstThenBlocks(t *testing.T) {
	engine := newRateLimitTestEngine(middleware.NewRateLimitMiddleware().Protected())

	for i := 0; i < 40; i++ {
		resp := performRateLimitRequest(t, engine, "1.2.3.4:1234")
		if resp.Code != http.StatusOK {
			t.Fatalf("expected request %d to pass, got %d", i+1, resp.Code)
		}
	}

	resp := performRateLimitRequest(t, engine, "1.2.3.4:1234")
	if resp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected protected rate limit after burst, got %d", resp.Code)
	}
}

func TestProtectedReadHeavyRateLimitAllowsBurstThenBlocks(t *testing.T) {
	engine := newRateLimitTestEngine(middleware.NewRateLimitMiddleware().ProtectedReadHeavy())

	for i := 0; i < 20; i++ {
		resp := performRateLimitRequest(t, engine, "1.2.3.4:1234")
		if resp.Code != http.StatusOK {
			t.Fatalf("expected request %d to pass, got %d", i+1, resp.Code)
		}
	}

	resp := performRateLimitRequest(t, engine, "1.2.3.4:1234")
	if resp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected protected read-heavy rate limit after burst, got %d", resp.Code)
	}
}

func TestPublicAuthRateLimitAllowsBurstThenBlocks(t *testing.T) {
	engine := newRateLimitTestEngine(middleware.NewRateLimitMiddleware().PublicAuth())

	for i := 0; i < 10; i++ {
		resp := performRateLimitRequest(t, engine, "1.2.3.4:1234")
		if resp.Code != http.StatusOK {
			t.Fatalf("expected request %d to pass, got %d", i+1, resp.Code)
		}
	}

	resp := performRateLimitRequest(t, engine, "1.2.3.4:1234")
	if resp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected public auth rate limit after burst, got %d", resp.Code)
	}
}

func TestAIRateLimitRemainsStrict(t *testing.T) {
	engine := newRateLimitTestEngine(middleware.NewRateLimitMiddleware().AI())

	resp := performRateLimitRequest(t, engine, "1.2.3.4:1234")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected first AI request to pass, got %d", resp.Code)
	}

	resp = performRateLimitRequest(t, engine, "1.2.3.4:1234")
	if resp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected AI rate limit to remain strict, got %d", resp.Code)
	}
}
