package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/adapters/http/handlers"
)

type DBPinger interface {
	Ping(ctx context.Context) error
}

func SetupRouter(
	db DBPinger,
	userHandler *handlers.UserHandler,
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

	// User routes
	if userHandler != nil {
		api.POST("/users", userHandler.CreateUser)
		api.GET("/users/:id", userHandler.GetUserByID)
	}

	return r
}
