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
func SetupRouter(
	db DBPinger,
	userHandler *handlers.UserHandler,
	authHandler *handlers.AuthHandler,
	authMiddleware *middleware.AuthMiddleware,
	exerciseHandler *handlers.ExerciseHandler,
	routineHandler *handlers.RoutineHandler,
	overviewHandler *handlers.OverviewHandler,
	healthHandler *handlers.HealthHandler,
	ticketHandler *handlers.TicketHandler,
	workoutHandler *handlers.WorkoutHandler,
	profileHandler *handlers.ProfileHandler,
	corsAllowOrigin ...string,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(corsMiddleware(resolveCORSAllowOrigin(corsAllowOrigin)))

	// Health check
	r.GET("/health", healthHandler.Health)

	// Base API group
	api := r.Group("/api")

	// --------------------
	// Public endpoints
	// --------------------
	api.POST("/users", userHandler.CreateUser)
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/logout", authHandler.Logout)

	// --------------------
	// Protected endpoints
	// --------------------
	protected := api.Group("")
	protected.Use(authMiddleware.RequireAuth())

	// General
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

	// Auth
	protected.GET("/auth/me", authHandler.Me)

	// Users
	protected.GET("/users", userHandler.ListAllUsers)

	protected.GET("/users/me", userHandler.GetMe)

	protected.GET("/users/:id", userHandler.GetUserByID)
	protected.DELETE("/users/:id", userHandler.DeleteUser)

	// Profile
	protected.GET("/profile/dashboard", profileHandler.GetDashboard)
	protected.PUT("/profile/goals", profileHandler.UpdateGoals)

	// Exercises
	protected.POST("/exercises", exerciseHandler.CreateExercise)
	protected.GET("/exercises/metadata", exerciseHandler.GetMetadata)
	protected.GET("/exercises/:id", exerciseHandler.GetExerciseByID)
	protected.GET("/exercises", exerciseHandler.ListExercises)
	protected.PUT("/exercises/:id", exerciseHandler.UpdateExercise)
	protected.DELETE("/exercises/:id", exerciseHandler.DeleteExercise)

	// Routines
	protected.GET("/routines", routineHandler.ListRoutines)

	// Dashboard
	protected.GET("/dashboard", overviewHandler.GetOverview)

	// Tickets
	protected.POST("/tickets", ticketHandler.CreateTicket)
	protected.GET("/tickets", ticketHandler.ListTickets)
	protected.PATCH("/tickets/:id/close", ticketHandler.CloseTicket)

	// Workout
	protected.POST("/workouts/planned", workoutHandler.CreatePlannedWorkout)
	protected.POST("/workout/start", workoutHandler.CreateWorkout)
	protected.GET("/workout/:id", workoutHandler.GetWorkoutByID)
	protected.POST("/workout/:id/finish", workoutHandler.FinishWorkout)
	protected.DELETE("/workout/:id", workoutHandler.RemoveWorkout)
	protected.POST("/workout/:id/exercise", workoutHandler.CreateWorkoutExercise)
	protected.GET("/workout/:id/exercises", workoutHandler.GetExercisesByWorkoutID)
	protected.POST("/workout/:id/exercises/:exercise_id/set", workoutHandler.CreateWorkoutSet)
	protected.GET("/workout/:id/exercises/:exercise_id/sets", workoutHandler.GetWorkoutSetsByExerciseID)
	protected.POST("/workout/:id/exercises/:exercise_id/sets/:set_id", workoutHandler.UpdateWorkoutSet)

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
