package middleware

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	defaultRateLimitTTL       = 30 * time.Minute
	defaultRateLimitSweepFreq = 5 * time.Minute
)

type rateLimitPolicy struct {
	name      string
	limit     rate.Limit
	burst     int
	keyFunc   func(*gin.Context) string
	errorText string
}

type rateLimitEntry struct {
	limiter  *rate.Limiter
	lastSeen  time.Time
}

type rateLimitStore struct {
	mu           sync.Mutex
	entries      map[string]*rateLimitEntry
	ttl          time.Duration
	sweepFreq    time.Duration
	lastSweep    time.Time
}

func newRateLimitStore(ttl, sweepFreq time.Duration) *rateLimitStore {
	if ttl <= 0 {
		ttl = defaultRateLimitTTL
	}
	if sweepFreq <= 0 {
		sweepFreq = defaultRateLimitSweepFreq
	}

	return &rateLimitStore{
		entries:   make(map[string]*rateLimitEntry),
		ttl:       ttl,
		sweepFreq: sweepFreq,
	}
}

func (s *rateLimitStore) limiterFor(key string, limit rate.Limit, burst int, now time.Time) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	if now.IsZero() {
		now = time.Now()
	}
	s.cleanupLocked(now)

	entry, ok := s.entries[key]
	if !ok {
		entry = &rateLimitEntry{
			limiter: rate.NewLimiter(limit, burst),
		}
		s.entries[key] = entry
	}

	entry.lastSeen = now
	return entry.limiter
}

func (s *rateLimitStore) cleanup(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(now)
}

func (s *rateLimitStore) cleanupLocked(now time.Time) {
	if !s.lastSweep.IsZero() && now.Sub(s.lastSweep) < s.sweepFreq {
		return
	}

	for key, entry := range s.entries {
		if now.Sub(entry.lastSeen) > s.ttl {
			delete(s.entries, key)
		}
	}
	s.lastSweep = now
}

// RateLimiter exposes reusable Gin middlewares for different endpoint groups.
type RateLimiter struct {
	store *rateLimitStore
}

// NewRateLimiter returns the default in-memory rate limiter configuration.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		store: newRateLimitStore(defaultRateLimitTTL, defaultRateLimitSweepFreq),
	}
}

// Login returns a strict IP-based limiter for login attempts.
func (r *RateLimiter) Login() gin.HandlerFunc {
	return r.middleware(rateLimitPolicy{
		name:      "auth_login",
		limit:     rate.Every(12 * time.Second), // 5 req/min
		burst:     3,
		keyFunc:   clientIPKey,
		errorText: "login rate limit exceeded",
	})
}

// Register returns a strict IP-based limiter for account creation.
func (r *RateLimiter) Register() gin.HandlerFunc {
	return r.middleware(rateLimitPolicy{
		name:      "auth_register",
		limit:     rate.Every(20 * time.Second), // 3 req/min
		burst:     2,
		keyFunc:   clientIPKey,
		errorText: "registration rate limit exceeded",
	})
}

// Protected returns the default limiter for authenticated traffic.
func (r *RateLimiter) Protected() gin.HandlerFunc {
	return r.middleware(rateLimitPolicy{
		name:      "protected",
		limit:     rate.Every(time.Second), // 60 req/min
		burst:     20,
		keyFunc:   authenticatedUserOrIPKey,
		errorText: "request rate limit exceeded",
	})
}

// ProtectedReadHeavy returns a tighter limiter for expensive authenticated reads.
func (r *RateLimiter) ProtectedReadHeavy() gin.HandlerFunc {
	return r.middleware(rateLimitPolicy{
		name:      "protected_read_heavy",
		limit:     rate.Every(2 * time.Second), // 30 req/min
		burst:     10,
		keyFunc:   authenticatedUserOrIPKey,
		errorText: "read rate limit exceeded",
	})
}

// AI returns a very strict limiter for AI routine generation and save endpoints.
func (r *RateLimiter) AI() gin.HandlerFunc {
	return r.middleware(rateLimitPolicy{
		name:      "routine_ai",
		limit:     rate.Every(30 * time.Minute), // 2 req/hour
		burst:     1,
		keyFunc:   authenticatedUserOrIPKey,
		errorText: "ai generation rate limit exceeded",
	})
}

// PublicAuth returns a moderate limiter for auth-adjacent public endpoints such as logout.
func (r *RateLimiter) PublicAuth() gin.HandlerFunc {
	return r.middleware(rateLimitPolicy{
		name:      "auth_public",
		limit:     rate.Every(6 * time.Second), // 10 req/min
		burst:     5,
		keyFunc:   clientIPKey,
		errorText: "auth rate limit exceeded",
	})
}

func (r *RateLimiter) middleware(policy rateLimitPolicy) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		key := policy.keyFunc(c)
		if strings.TrimSpace(key) == "" {
			key = clientIPKey(c)
		}

		limiter := r.store.limiterFor(policy.name+"|"+key, policy.limit, policy.burst, now)
		reservation := limiter.ReserveN(now, 1)
		if !reservation.OK() {
			r.reject(c, policy, 0)
			return
		}

		delay := reservation.DelayFrom(now)
		if delay > 0 {
			reservation.CancelAt(now)
			r.reject(c, policy, delay)
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(policy.burst))
		c.Header("X-RateLimit-Remaining", remainingTokensHeader(limiter.TokensAt(now)))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Unix(), 10))

		c.Next()
	}
}

func (r *RateLimiter) reject(c *gin.Context, policy rateLimitPolicy, delay time.Duration) {
	retryAfter := int(math.Ceil(delay.Seconds()))
	if retryAfter < 1 {
		retryAfter = 1
	}

	c.Header("Retry-After", strconv.Itoa(retryAfter))
	c.Header("X-RateLimit-Limit", strconv.Itoa(policy.burst))
	c.Header("X-RateLimit-Remaining", "0")
	c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(delay).Unix(), 10))
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"error":       policy.errorText,
		"retry_after": retryAfter,
	})
}

func remainingTokensHeader(tokens float64) string {
	if tokens <= 0 {
		return "0"
	}

	return strconv.Itoa(int(math.Floor(tokens)))
}

func clientIPKey(c *gin.Context) string {
	ip := strings.TrimSpace(c.ClientIP())
	if ip == "" {
		return "unknown-ip"
	}
	return "ip:" + ip
}

func authenticatedUserOrIPKey(c *gin.Context) string {
	if userIDValue, ok := c.Get(ContextUserIDKey); ok {
		if userID, ok := userIDValue.(string); ok {
			userID = strings.TrimSpace(userID)
			if userID != "" {
				return "user:" + userID
			}
		}
	}

	return clientIPKey(c)
}
