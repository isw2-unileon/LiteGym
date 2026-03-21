package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/handler"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRouter(db *pgxpool.Pool) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")

	userRepo := repositories.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

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

	api.POST("/users", userHandler.CreateUser)
	api.GET("/users/:id", userHandler.GetUserByID)

	return r
}