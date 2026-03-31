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
	"github.com/jackc/pgx/v5/pgxpool"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
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

	// Initialize services
	userService := service.NewUserService(userRepo)
	exerciseService := service.NewExerciseService(exerciseRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService)
	healthHandler := handlers.NewHealthHandler()

	r := transport.SetupRouter(db, userHandler, exerciseHandler, healthHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
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
