package app

import (
	"context"
	"sync"

	"github.com/amir-mohammad-HP/crontask/internal/signals"
	"github.com/amir-mohammad-HP/crontask/internal/types"
	"github.com/amir-mohammad-HP/crontask/internal/worker"
	"github.com/amir-mohammad-HP/crontask/pkg/logger"
	"github.com/amir-mohammad-HP/crontask/pkg/shutdown"
)

type App struct {
	config        *types.Config
	logger        *logger.StdLogger
	worker        *worker.Worker
	shutdown      *shutdown.Manager
	signalHandler *signals.Handler
	wg            sync.WaitGroup
}

func New(cfg *types.Config, logger *logger.StdLogger) *App {
	return &App{
		config:        cfg,
		logger:        logger,
		worker:        worker.New(cfg, logger),
		shutdown:      shutdown.NewManager(logger),
		signalHandler: signals.NewHandler(logger),
	}
}

func (a *App) Run() error {
	a.logger.Debug("Starting CronTask application")

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start signal handler
	go a.signalHandler.Handle(ctx, func() {
		a.logger.Debug("Received shutdown signal")
		a.shutdown.Initiate()
	})

	// Register cleanup tasks
	a.shutdown.RegisterTask("worker", a.worker.Stop)
	a.shutdown.RegisterTask("application", a.cleanup)

	// Start worker
	if err := a.worker.Start(ctx, &a.wg); err != nil {
		return err
	}

	// Wait for shutdown
	<-a.shutdown.Done()
	a.wg.Wait()

	a.logger.Debug("Application shutdown complete")
	return nil
}

func (a *App) cleanup() error {
	a.logger.Debug("Performing application cleanup")
	// Add cleanup logic here
	return nil
}
