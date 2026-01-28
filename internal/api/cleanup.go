package api

import (
	"context"
	"log/slog"
	"time"

	"github.com/rayenfassatoui/tafcha-cli/internal/storage"
)

// CleanupWorker periodically removes expired snippets.
type CleanupWorker struct {
	repo     storage.Repository
	interval time.Duration
	logger   *slog.Logger
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// NewCleanupWorker creates a new cleanup worker.
func NewCleanupWorker(repo storage.Repository, interval time.Duration, logger *slog.Logger) *CleanupWorker {
	return &CleanupWorker{
		repo:     repo,
		interval: interval,
		logger:   logger,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start begins the cleanup loop in a goroutine.
func (w *CleanupWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

func (w *CleanupWorker) run(ctx context.Context) {
	defer close(w.doneCh)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run once at startup
	w.cleanup()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("cleanup worker stopping due to context cancellation")
			return
		case <-w.stopCh:
			w.logger.Info("cleanup worker stopping")
			return
		case <-ticker.C:
			w.cleanup()
		}
	}
}

func (w *CleanupWorker) cleanup() {
	count, err := w.repo.DeleteExpired()
	if err != nil {
		w.logger.Error("failed to delete expired snippets", "error", err)
		return
	}
	if count > 0 {
		w.logger.Info("cleanup completed", "deleted_count", count)
	}
}

// Stop signals the worker to stop and waits for it to finish.
func (w *CleanupWorker) Stop() {
	close(w.stopCh)
	<-w.doneCh
}
