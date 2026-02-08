package types

import "time"

// CronJob represents a container-based cron job
type CronJob struct {
	ContainerID   string     `json:"container_id"`
	ContainerName string     `json:"container_name"`
	CronExpr      string     `json:"cron_expression"`
	Task          string     `json:"task"`
	LabelKey      string     `json:"label_key"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	LastExecution *time.Time `json:"last_execution,omitempty"`
}
