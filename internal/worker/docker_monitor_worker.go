package worker

import "context"

func (w *Worker) runDockerMon(ctx context.Context) {
	defer w.logger.Debug("docker monitor | Worker stopped")
	defer w.cleanupDockerMon()

	// Start Docker monitor if enabled
	if w.dockerMon != nil {
		if err := w.dockerMon.Start(ctx); err != nil {
			w.logger.Error("docker monitor | Failed to start Docker monitor, %s", err.Error())
		} else {
			go w.handleDockerEvents(ctx)
		}
	}
}

func (w *Worker) cleanupDockerMon() {
	w.logger.Debug("docker monitor | cleanup")
	if w.dockerMon != nil {
		w.dockerMon.Stop()
	}
}
