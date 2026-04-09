package transport

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/handlers"
)

// DBPinger defines the behavior required to check database connectivity.
type DBPinger interface {
	Ping(ctx context.Context) error
}

// SetupRouter configures and returns the application HTTP router.
func SetupRouter(
	db DBPinger,
	userHandler *handlers.UserHandler,
	authHandler *handlers.AuthHandler,
	exerciseHandler *handlers.ExerciseHandler,
	healthHandler *handlers.HealthHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

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

	// Exercise endpoints
	r.POST("/api/exercises", exerciseHandler.CreateExercise)
	r.GET("/api/exercises/:id", exerciseHandler.GetExerciseByID)
	r.GET("/api/exercises", exerciseHandler.ListExercises)

	return r
}
