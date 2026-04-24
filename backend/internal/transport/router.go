package transport

import (
	"context"
	"net/http"
	"strings"
	"time"

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
// This router enables CORS with credential support and a configurable allowed origin.
func SetupRouter(
	db DBPinger,
	userHandler *handlers.UserHandler,
	authHandler *handlers.AuthHandler,
	authMiddleware *middleware.AuthMiddleware,
	exerciseHandler *handlers.ExerciseHandler,
	healthHandler *handlers.HealthHandler,
	corsAllowOrigin ...string,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(corsMiddleware(resolveCORSAllowOrigin(corsAllowOrigin)))

	// Health check endpoints
	r.GET("/health", healthHandler.Health)

	api := r.Group("/api")

	// Public API endpoints
	api.POST("/users", userHandler.CreateUser)
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/logout", authHandler.Logout)

	// Protected API endpoints
	protected := api.Group("")
	protected.Use(authMiddleware.RequireAuth())

	protected.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello from the API"})
	})

	protected.GET("/db/health", func(c *gin.Context) {
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

	protected.GET("/auth/me", authHandler.Me)
	protected.GET("/users/:id", userHandler.GetUserByID)
	
	protected.GET("/exercises/:id", exerciseHandler.GetExerciseByID)
	protected.GET("/exercises", exerciseHandler.ListExercises)

	protected.POST("/exercises", exerciseHandler.CreateExercise)
	return r
}

func resolveCORSAllowOrigin(corsAllowOrigin []string) string {
	if len(corsAllowOrigin) == 0 || strings.TrimSpace(corsAllowOrigin[0]) == "" {
		return "*"
	}

	return strings.TrimSpace(corsAllowOrigin[0])
}

func corsMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestOrigin := c.GetHeader("Origin")
		responseOrigin := allowOrigin

		// Credentialed CORS requests cannot use "*", so reflect the request origin in dev.
		if allowOrigin == "*" && requestOrigin != "" {
			responseOrigin = requestOrigin
		}

		if requestOrigin != "" && (allowOrigin == "*" || requestOrigin == allowOrigin) {
			c.Header("Access-Control-Allow-Origin", responseOrigin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
