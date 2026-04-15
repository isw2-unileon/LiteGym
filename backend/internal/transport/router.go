package transport

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/handlers"
)

type DBPinger interface {
	Ping(ctx context.Context) error
}

func SetupRouter(
	db DBPinger,
	userHandler *handlers.UserHandler,
	authHandler *handlers.AuthHandler,
	exerciseHandler *handlers.ExerciseHandler,
	healthHandler *handlers.HealthHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// TEMP CORS CONFIG
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}

	r.Use(cors.New(corsConfig))

	// --------------------------------------

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
