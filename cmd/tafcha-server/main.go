// Tafcha server - Plain text publishing API
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rayenfassatoui/tafcha-cli/internal/api"
	"github.com/rayenfassatoui/tafcha-cli/internal/config"
	"github.com/rayenfassatoui/tafcha-cli/internal/storage"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("starting tafcha server",
		"host", cfg.Host,
		"port", cfg.Port,
		"base_url", cfg.BaseURL,
	)

	// Initialize database
	ctx := context.Background()
	repo, err := storage.NewPostgresRepository(ctx, storage.PostgresConfig{
		URL:         cfg.DatabaseURL,
		MaxConns:    int32(cfg.MaxDBConns),
		MinConns:    int32(cfg.MinDBConns),
		MaxConnLife: cfg.DBConnMaxLife,
	}, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Run migrations
	if err := repo.Migrate(ctx); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Start cleanup worker
	cleanupWorker := api.NewCleanupWorker(repo, cfg.CleanupInterval, logger)
	cleanupWorker.Start(ctx)
	defer cleanupWorker.Stop()

	// Create API server
	server := api.NewServer(cfg, repo, logger)

	// Configure HTTP server
	httpServer := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      server.Handler(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("server listening", "addr", cfg.Addr())
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}
