// internal/worker/worker.go
package worker

import (
	"context"
	"sync"
	"time"

	"github.com/amir-mohammad-HP/crontask/internal/job"
	"github.com/amir-mohammad-HP/crontask/internal/types"
	"github.com/amir-mohammad-HP/crontask/pkg/docker"
	"github.com/amir-mohammad-HP/crontask/pkg/logger"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Worker struct {
	config      *types.Config
	logger      *logger.StdLogger
	shutdown    chan struct{}
	mu          sync.RWMutex
	cron        *cron.Cron
	jobRegistry *job.JobRegistry
	dockerMon   *docker.DockerMonitor
}

func New(cfg *types.Config, logger *logger.StdLogger) *Worker {
	w := &Worker{
		config:   cfg,
		logger:   logger,
		shutdown: make(chan struct{}),
		cron: cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		))),
	}

	// Initialize Docker monitor if enabled
	if cfg.Docker.Enabled {
		zapLogger := zap.NewExample() // Convert your logger or create a new one
		monitor, err := docker.NewMonitor(&cfg.Docker, zapLogger)
		if err != nil {
			logger.Error("Failed to create Docker monitor", err)
		} else {
			w.dockerMon = monitor
			w.jobRegistry = job.NewJobRegistry(monitor)
		}
	}

	return w
}

func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup) error {
	w.logger.Info("Starting worker")

	wg.Add(1)
	go w.run(ctx)

	return nil
}

func (w *Worker) run(ctx context.Context) {
	defer w.logger.Info("Worker stopped")

	// Start Docker monitor if enabled
	if w.dockerMon != nil {
		if err := w.dockerMon.Start(ctx); err != nil {
			w.logger.Error("Failed to start Docker monitor", err)
		} else {
			go w.handleDockerEvents(ctx)
		}
	}

	// Start cron scheduler
	w.cron.Start()
	w.logger.Info("Worker cron scheduler started")

	// Wait for shutdown
	select {
	case <-ctx.Done():
		w.logger.Info("Worker received context cancellation")
		w.cleanup()
	case <-w.shutdown:
		w.logger.Info("Worker received shutdown signal")
		w.cleanup()
	}
}

func (w *Worker) handleDockerEvents(ctx context.Context) {
	if w.dockerMon == nil {
		return
	}

	events := w.dockerMon.GetEvents()
	for {
		select {
		case event := <-events:
			w.processDockerEvent(event)
		case <-ctx.Done():
			return
		case <-w.shutdown:
			return
		}
	}
}

func (w *Worker) processDockerEvent(event docker.ContainerEvent) {
	switch event.Action {
	case "scan", "create", "start", "update":
		if event.Container.State == "running" {
			w.registerContainerJobs(event.Container)
		}
	case "die", "destroy":
		w.unregisterContainerJobs(event.ContainerID)
	}
}

func (w *Worker) registerContainerJobs(container *docker.ContainerInfo) {
	if w.dockerMon == nil || w.jobRegistry == nil {
		return
	}

	// Remove existing jobs for this container first
	w.unregisterContainerJobs(container.ID)

	// Extract and register new jobs
	cronJobs := w.dockerMon.ExtractCronJobs(container)
	for _, cronJob := range cronJobs {
		dockerJob := job.NewDockerJob(
			cronJob.ContainerID,
			cronJob.ContainerName,
			cronJob.CronExpr,
			cronJob.Task,
			w.dockerMon,
		)

		// Add to registry
		if w.jobRegistry.AddJob(dockerJob) {
			// Schedule the job
			entryID, err := w.cron.AddFunc(cronJob.CronExpr, func() {
				w.executeJob(dockerJob)
			})

			if err != nil {
				w.logger.Error("Failed to schedule job",
					err,
					"container", container.ID[:12],
					"cron", cronJob.CronExpr)
				w.jobRegistry.RemoveJob(dockerJob.Name())
			} else {
				dockerJob.SetCronEntryID(entryID)
				w.logger.Info("Job registered",
					"container", container.ID[:12],
					"name", container.Name,
					"cron", cronJob.CronExpr,
					"task", cronJob.Task)
			}
		}
	}
}

func (w *Worker) unregisterContainerJobs(containerID string) {
	if w.jobRegistry == nil {
		return
	}

	removedJobs := w.jobRegistry.RemoveJobsByContainer(containerID)
	for _, jobID := range removedJobs {
		// Note: cron entries are automatically removed when container stops
		w.logger.Info("Job unregistered",
			"container", containerID[:12],
			"job", jobID)
	}
}

func (w *Worker) executeJob(job *job.DockerJob) {
	w.logger.Info("Executing job",
		"job", job.Name(),
		"container", job.GetContainerID()[:12],
		"time", time.Now().Format("2006-01-02 15:04:05"))

	if err := job.Execute(); err != nil {
		w.logger.Error("Job execution failed",
			err,
			"job", job.Name(),
			"container", job.GetContainerID()[:12])
	} else {
		w.logger.Info("Job executed successfully",
			"job", job.Name(),
			"container", job.GetContainerID()[:12])
	}
}

func (w *Worker) cleanup() {
	w.cron.Stop()
	if w.dockerMon != nil {
		w.dockerMon.Stop()
	}
}

func (w *Worker) Stop() error {
	w.logger.Info("Stopping worker")
	close(w.shutdown)
	return nil
}

// GetStats returns worker statistics
func (w *Worker) GetStats() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	stats := map[string]interface{}{
		"cron_entries": len(w.cron.Entries()),
	}

	if w.jobRegistry != nil {
		stats["registered_jobs"] = w.jobRegistry.Count()
	}

	return stats
}

// ListJobs returns all registered jobs
func (w *Worker) ListJobs() []map[string]interface{} {
	if w.jobRegistry == nil {
		return []map[string]interface{}{}
	}

	jobs := w.jobRegistry.GetAllJobs()
	result := make([]map[string]interface{}, 0, len(jobs))

	for _, job := range jobs {
		result = append(result, map[string]interface{}{
			"id":           job.Name(),
			"container_id": job.GetContainerID()[:12],
			"cron_expr":    job.Schedule(),
			"last_run":     job.GetLastRun(),
			"next_run":     job.GetNextRun(),
		})
	}

	return result
}
