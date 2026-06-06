package transport

import (
	"context"
	"net/http"
	"net/url"
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
	return setupRouterInternal(
		db,
		userHandler,
		authHandler,
		authMiddleware,
		exerciseHandler,
		nil,
		overviewHandler,
		healthHandler,
		ticketHandler,
		workoutHandler,
		profileHandler,
		corsAllowOrigin...,
	)
}

// SetupRouterWithRoutine configures the HTTP router including routine endpoints.
func SetupRouterWithRoutine(
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
	return setupRouterInternal(
		db,
		userHandler,
		authHandler,
		authMiddleware,
		exerciseHandler,
		routineHandler,
		overviewHandler,
		healthHandler,
		ticketHandler,
		workoutHandler,
		profileHandler,
		corsAllowOrigin...,
	)
}

func setupRouterInternal(
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
	r.Use(corsMiddleware(resolveCORSAllowOrigins(corsAllowOrigin)))
	rateLimiter := middleware.NewRateLimiter()

	// Health check
	r.GET("/health", healthHandler.Health)

	// Base API group
	api := r.Group("/api")

	// --------------------
	// Public endpoints
	// --------------------
	api.POST("/users", rateLimiter.Register(), userHandler.CreateUser)
	api.POST("/auth/register", rateLimiter.Register(), authHandler.Register)
	api.POST("/auth/login", rateLimiter.Login(), authHandler.Login)
	api.POST("/auth/logout", rateLimiter.PublicAuth(), authHandler.Logout)

	// --------------------
	// Protected endpoints
	// --------------------
	protected := api.Group("")
	protected.Use(authMiddleware.RequireAuth())
	protected.Use(rateLimiter.Protected())

	heavyReads := protected.Group("")
	heavyReads.Use(rateLimiter.ProtectedReadHeavy())

	aiLimited := protected.Group("")
	aiLimited.Use(rateLimiter.AI())

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
	if profileHandler != nil {
		protected.GET("/profile/dashboard", profileHandler.GetDashboard)
		protected.PUT("/profile/goals", profileHandler.UpdateGoals)
		protected.POST("/profile/metrics", profileHandler.AddBodyMetric)
		protected.POST("/profile/ai-analysis", rateLimiter.ProfileAI(), profileHandler.GetAIAnalysis)
	}

	registerExerciseRoutes(protected, heavyReads, exerciseHandler)

	// Routines
	if routineHandler != nil {
		heavyReads.GET("/routines", routineHandler.ListRoutines)
		heavyReads.GET("/routines/:id", routineHandler.GetRoutine)
		aiLimited.POST("/routines/ai/generate", routineHandler.GenerateRoutineJSON)
		protected.POST("/routines/ai/save", routineHandler.SaveAIRoutine)
		aiLimited.POST("/routines/:id/ai/upgrade", routineHandler.UpgradeRoutineJSON)
		protected.POST("/routines/:id/ai/save-as-new", routineHandler.SaveUpgradedRoutineAsNew)
		protected.PUT("/routines/:id/ai/overwrite", routineHandler.OverwriteRoutineWithAI)
	}

	// Dashboard
	heavyReads.GET("/dashboard", overviewHandler.GetOverview)

	// Tickets
	protected.POST("/tickets", ticketHandler.CreateTicket)
	protected.GET("/tickets", ticketHandler.ListTickets)
	protected.PATCH("/tickets/:id/close", ticketHandler.CloseTicket)

	registerWorkoutRoutes(protected, workoutHandler)

	return r
}

func registerExerciseRoutes(
	protected, heavyReads gin.IRoutes,
	exerciseHandler *handlers.ExerciseHandler,
) {
	// Exercises
	protected.POST("/exercises", exerciseHandler.CreateExercise)
	protected.GET("/exercises/metadata", exerciseHandler.GetMetadata)
	heavyReads.GET("/exercises/:id/insights", exerciseHandler.GetExerciseInsights)
	heavyReads.GET("/exercises/:id/workout-sessions", exerciseHandler.ListWorkoutSessionsByExercise)
	protected.GET("/exercises/:id", exerciseHandler.GetExerciseByID)
	heavyReads.GET("/exercises", exerciseHandler.ListExercises)
	protected.PUT("/exercises/:id", exerciseHandler.UpdateExercise)
	protected.DELETE("/exercises/:id", exerciseHandler.DeleteExercise)
}

func registerWorkoutRoutes(
	protected gin.IRoutes,
	workoutHandler *handlers.WorkoutHandler,
) {
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
}

func resolveCORSAllowOrigins(corsAllowOrigin []string) []string {
	if len(corsAllowOrigin) == 0 || strings.TrimSpace(corsAllowOrigin[0]) == "" {
		return []string{"*"}
	}

	values := strings.Split(corsAllowOrigin[0], ",")
	origins := make([]string, 0, len(values))
	for _, value := range values {
		origin := strings.TrimSpace(value)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return []string{"*"}
	}
	return origins
}

func corsMiddleware(allowOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestOrigin := c.GetHeader("Origin")
		responseOrigin := ""

		for _, allowOrigin := range allowOrigins {
			if allowOrigin == "*" && requestOrigin != "" {
				responseOrigin = requestOrigin
				break
			}
			if requestOrigin != "" && requestOrigin == allowOrigin {
				responseOrigin = requestOrigin
				break
			}
		}
		if responseOrigin == "" && isLocalDevelopmentOrigin(requestOrigin) {
			responseOrigin = requestOrigin
		}

		if responseOrigin != "" {
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

func isLocalDevelopmentOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	hostname := parsed.Hostname()
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
}
