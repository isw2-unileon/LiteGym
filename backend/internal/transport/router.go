package transport

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/handlers"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

// DBPinger defines the behavior required to check database connectivity.
type DBPinger interface {
	Ping(ctx context.Context) error
}

// SetupRouter configures and returns the application HTTP router.
//
// This router enables CORS with credential support and configurable allowed origins.
// Configuration is read from environment variables:
// - CORS_ALLOW_ORIGIN: comma-separated list of allowed origins (default "*")
// - CORS_ALLOW_CREDENTIALS: whether to allow credentials (default "true")
func SetupRouter(
	db DBPinger,
	userHandler *handlers.UserHandler,
	authHandler *handlers.AuthHandler,
	authMiddleware *middleware.AuthMiddleware,
	exerciseHandler *handlers.ExerciseHandler,
	healthHandler *handlers.HealthHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Read CORS configuration from environment
	allowed := os.Getenv("CORS_ALLOW_ORIGIN")
	if strings.TrimSpace(allowed) == "" {
		allowed = "*"
	}

	allowCreds := true
	if v := os.Getenv("CORS_ALLOW_CREDENTIALS"); strings.TrimSpace(v) != "" {
		if parsed := strings.ToLower(strings.TrimSpace(v)); parsed == "false" || parsed == "0" {
			allowCreds = false
		}
	}

	// Build cors.Config with sensible defaults and credential support
	corsCfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: allowCreds,
		MaxAge:           12 * time.Hour,
	}

	// If allowed is wildcard, accept any origin. When credentials are allowed, we must
	// echo the Origin header back (cannot use literal wildcard in Access-Control-Allow-Origin).
	if strings.TrimSpace(allowed) == "*" {
		// Accept all origins by returning true in AllowOriginFunc.
		corsCfg.AllowOriginFunc = func(origin string) bool {
			// Basic safety: reject empty origins
			return origin != ""
		}
	} else {
		// Parse comma-separated list of origins
		parts := strings.Split(allowed, ",")
		origins := make([]string, 0, len(parts))
		for _, p := range parts {
			if o := strings.TrimSpace(p); o != "" {
				origins = append(origins, o)
			}
		}
		corsCfg.AllowOrigins = origins
	}

	// Register CORS middleware
	r.Use(cors.New(corsCfg))

	// Health check endpoints
	r.GET("/health", healthHandler.Health)

	api := r.Group("/api")

	api.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello from the API"})
	})

	api.GET("/db/health", func(c *gin.Context) {
		pingCtx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		if err := db.Ping(pingCtx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"service": "database",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "database",
		})
	})

	// User endpoints
	r.POST("/api/users", userHandler.CreateUser)
	r.GET("/api/users/:id", userHandler.GetUserByID)
	r.POST("/api/auth/login", authHandler.Login)

	// Exercise endpoints (protected)
	protected := r.Group("/api")
	protected.Use(authMiddleware.RequireAuth())

	protected.POST("/exercises", exerciseHandler.CreateExercise)
	protected.GET("/exercises/:id", exerciseHandler.GetExerciseByID)
	protected.GET("/exercises", exerciseHandler.ListExercises)

	return r
}
