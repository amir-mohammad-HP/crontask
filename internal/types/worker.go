package types

import "time"

type WorkerConfig struct {
	Interval      time.Duration `mapstructure:"interval"`
	MaxJobs       int           `mapstructure:"max_jobs"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
}
