package config

import (
	"time"
)

type Config struct {
	AppName     string
	Environment string
	LogLevel    string
	Worker      WorkerConfig
	Shutdown    ShutdownConfig
}

type WorkerConfig struct {
	Interval      time.Duration
	MaxJobs       int
	RetryAttempts int
}

type ShutdownConfig struct {
	Timeout time.Duration
}

func Load() (*Config, error) {
	// Load from env vars, config file, or defaults
	return &Config{
		AppName:     "CronTask",
		Environment: "development",
		LogLevel:    "info",
		Worker: WorkerConfig{
			Interval:      5 * time.Second,
			MaxJobs:       10,
			RetryAttempts: 3,
		},
		Shutdown: ShutdownConfig{
			Timeout: 30 * time.Second,
		},
	}, nil
}
