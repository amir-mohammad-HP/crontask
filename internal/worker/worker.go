package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/amir-mohammad-HP/crontask/internal/config"
	"github.com/amir-mohammad-HP/crontask/pkg/logger"
)

type Worker struct {
	config   *config.Config
	logger   logger.Logger
	jobs     []Job
	shutdown chan struct{}
	mu       sync.RWMutex
}

type Job interface {
	Execute() error
	Name() string
	Schedule() time.Duration
}

func New(cfg *config.Config, logger logger.Logger) *Worker {
	return &Worker{
		config:   cfg,
		logger:   logger,
		shutdown: make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup) error {
	w.logger.Info("Starting worker")

	wg.Add(1)
	go func() {
		defer wg.Done()
		w.run(ctx)
	}()

	return nil
}

func (w *Worker) run(ctx context.Context) {
	w.logger.Info("Worker main loop started")

	ticker := time.NewTicker(w.config.Worker.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Worker received context cancellation")
			return
		case <-w.shutdown:
			w.logger.Info("Worker received shutdown signal")
			return
		case t := <-ticker.C:
			w.executeJobs(t)
		}
	}
}

func (w *Worker) executeJobs(t time.Time) {
	fmt.Printf("Current datetime: %s\n", t.Format("2006-01-02 15:04:05"))
	// Execute registered jobs here
}

func (w *Worker) Stop() error {
	w.logger.Info("Stopping worker")
	close(w.shutdown)
	return nil
}

func (w *Worker) RegisterJob(job Job) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.jobs = append(w.jobs, job)
}
