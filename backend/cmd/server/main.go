package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/config"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repository"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/handlers"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() { //nolint:funlen // Server bootstrap wires all repositories, services, handlers, and graceful shutdown in one place.
	ctx := context.Background()

	cfg := config.Load()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}

	if err := db.Ping(ctx); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	defer db.Close()
	logger.Info("connected to database")

	gin.SetMode(cfg.GinMode)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	exerciseRepo := repository.NewExerciseRepository(db)
	routineRepo := repository.NewRoutineRepository(db)
	manualRoutineRepo := repository.NewManualRoutineRepository(db)
	overviewWorkoutRepo := repository.NewOverviewWorkoutRepository(db)
	bodyMetricRepo := repository.NewBodyMetricRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	workoutRepo := repository.NewWorkoutRepository(db)
	verificationTokenRepo := repository.NewVerificationTokenRepository(db)
	passwordResetTokenRepo := repository.NewPasswordResetTokenRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	tokenService := service.NewTokenService(cfg.JWTSecret, "grupo-16-backend", cfg.AuthTokenTTL)
	exerciseService := service.NewExerciseService(exerciseRepo)
	overviewService := service.NewOverviewService(routineRepo, overviewWorkoutRepo, bodyMetricRepo)
	workoutSessionRepo := repository.NewWorkoutSessionRepository(db)
	routineAIService := service.NewRoutineAIService(
		routineRepo,
		exerciseService,
		workoutSessionRepo,
		bodyMetricRepo,
		cfg.GeminiAPIKey,
		cfg.GeminiModel,
	)
	routineService := service.NewRoutineService(routineRepo)
	manualRoutineService := service.NewManualRoutineService(manualRoutineRepo)
	ticketService := service.NewTicketService(ticketRepo)
	workoutService := service.NewWorkoutService(workoutRepo)

	var emailService service.EmailService
	if cfg.SMTPUser != "" && cfg.SMTPPass != "" {
		emailService = service.NewSMTPEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	} else {
		emailService = service.NewMockEmailService()
	}

	verificationService := service.NewVerificationService(
		userRepo,
		verificationTokenRepo,
		emailService,
		cfg.FrontendURL,
	)

	passwordRecoveryService := service.NewPasswordRecoveryService(
		userRepo,
		passwordResetTokenRepo,
		emailService,
		cfg.FrontendURL,
	)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, tokenService, verificationService, passwordRecoveryService, cfg.AuthCookieName, cfg.AuthCookieSecure)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	routineHandler := handlers.NewRoutineHandler(routineService, routineAIService, manualRoutineService)
	overviewHandler := handlers.NewOverviewHandler(overviewService)
	healthHandler := handlers.NewHealthHandler()
	ticketHandler := handlers.NewTicketHandler(ticketService, userService)
	workoutHandler := handlers.NewWorkoutHandler(workoutService)
	profileHandler := handlers.NewProfileHandler(service.NewProfileService(repository.NewProfileRepository(db), cfg.GeminiAPIKey, cfg.GeminiModel))
	authMiddleware := middleware.NewAuthMiddleware(tokenService, cfg.AuthCookieName, userService)

	r := transport.SetupRouterWithRoutine(
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
		cfg.CORSAllowOrigin,
	)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}

	logger.Info("server stopped")
}
