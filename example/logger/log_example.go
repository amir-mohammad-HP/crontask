package main

import (
	"github.com/amir-mohammad-HP/crontask/pkg/logger"
)

func main() {
	// Create logger
	log := logger.New("INFO")

	// Use it
	log.Info("Starting CronTask")
	log.WithField("version", "1.0.0").Info("Application info")

	// In your worker
	log.WithFields(map[string]any{
		"job_id": "job-123",
		"status": "started",
	}).Info("Job execution")
}
