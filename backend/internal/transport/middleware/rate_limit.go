package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware provides in-memory request throttling by route category.
type RateLimitMiddleware struct {
	mu         sync.Mutex
	requests   map[string]*rate.Limiter
	login      *rate.Limiter
	register   *rate.Limiter
	protected  *rate.Limiter
	heavyRead  *rate.Limiter
	publicAuth *rate.Limiter
	ai         *rate.Limiter
	profileAI  *rate.Limiter
}

// NewRateLimitMiddleware creates the shared rate limit middleware instance.
func NewRateLimitMiddleware() *RateLimitMiddleware {
	return &RateLimitMiddleware{
		requests:   make(map[string]*rate.Limiter),
		login:      rate.NewLimiter(rate.Every(6*time.Second), 5),         // 10 req/min
		register:   rate.NewLimiter(rate.Every(3*time.Second), 5),         // 20 req/min
		protected:  rate.NewLimiter(rate.Every(500*time.Millisecond), 40), // 120 req/min
		heavyRead:  rate.NewLimiter(rate.Every(time.Second), 20),          // 60 req/min
		publicAuth: rate.NewLimiter(rate.Every(3*time.Second), 10),        // 20 req/min
		ai:         rate.NewLimiter(rate.Every(10*time.Minute), 1),        // 0.1 req/min
		profileAI:  rate.NewLimiter(rate.Every(15*time.Minute), 1),        // 0.066 req/min
	}
}

// NewRateLimiter returns the shared rate limit middleware instance.
func NewRateLimiter() *RateLimitMiddleware {
	return NewRateLimitMiddleware()
}

func (m *RateLimitMiddleware) limiterFor(category, clientIP string, template *rate.Limiter) *rate.Limiter {
	key := category + ":" + clientIP

	m.mu.Lock()
	defer m.mu.Unlock()

	if limiter, ok := m.requests[key]; ok {
		return limiter
	}

	limiter := rate.NewLimiter(template.Limit(), template.Burst())
	m.requests[key] = limiter
	return limiter
}

func (m *RateLimitMiddleware) withLimit(category string, template *rate.Limiter, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.limiterFor(category, c.ClientIP(), template).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": message,
			})
			return
		}

		c.Next()
	}
}

// Login limits login attempts.
func (m *RateLimitMiddleware) Login() gin.HandlerFunc {
	return m.withLimit("login", m.login, "too many login attempts")
}

// Register limits registration attempts.
func (m *RateLimitMiddleware) Register() gin.HandlerFunc {
	return m.withLimit("register", m.register, "too many register attempts")
}

// Protected limits general authenticated traffic.
func (m *RateLimitMiddleware) Protected() gin.HandlerFunc {
	return m.withLimit("protected", m.protected, "too many requests")
}

// ProtectedReadHeavy limits read-heavy authenticated traffic.
func (m *RateLimitMiddleware) ProtectedReadHeavy() gin.HandlerFunc {
	return m.withLimit("protected-read-heavy", m.heavyRead, "too many read requests")
}

// PublicAuth limits public authentication endpoints.
func (m *RateLimitMiddleware) PublicAuth() gin.HandlerFunc {
	return m.withLimit("public-auth", m.publicAuth, "too many authentication requests")
}

// AI limits AI generation requests.
func (m *RateLimitMiddleware) AI() gin.HandlerFunc {
	return m.withLimit("ai", m.ai, "ai rate limit exceeded")
}

// ProfileAI limits AI requests from the profile flow.
func (m *RateLimitMiddleware) ProfileAI() gin.HandlerFunc {
	return m.withLimit("profile-ai", m.profileAI, "profile ai rate limit exceeded")
}
