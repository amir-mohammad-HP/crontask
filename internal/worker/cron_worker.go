package worker

import (
	"context"
	"sync"
)

func (w *Worker) runCron(ctx context.Context, wg *sync.WaitGroup) {
	defer w.logger.Debug("cron worker | stopped")
	defer wg.Done()
	defer w.cleanupCron()

	// Start cron scheduler
	w.cron.Start()
	w.logger.Debug("cron worker | cron scheduler started")

	// Wait for shutdown
	select {
	case <-ctx.Done():
		w.logger.Debug("cron worker | received context cancellation")
		return
	case <-w.shutdown:
		w.logger.Debug("cron worker | received shutdown signal")
		return
	}
}

func (w *Worker) cleanupCron() {
	w.logger.Debug("cron worker | cleanup")
	w.cron.Stop()
}
