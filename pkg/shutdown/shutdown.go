package shutdown

import (
	"context"
	"time"

	"github.com/amir-mohammad-HP/crontask/pkg/logger"
)

type Task func() error

type Manager struct {
	logger   *logger.StdLogger
	tasks    map[string]Task
	shutdown chan struct{}
	timeout  time.Duration
}

func NewManager(logger *logger.StdLogger) *Manager {
	return &Manager{
		logger:   logger,
		tasks:    make(map[string]Task),
		shutdown: make(chan struct{}),
		timeout:  30 * time.Second,
	}
}

func (m *Manager) RegisterTask(name string, task Task) {
	m.tasks[name] = task
}

func (m *Manager) Initiate() {
	close(m.shutdown)
}

func (m *Manager) Done() <-chan struct{} {
	return m.shutdown
}

func (m *Manager) Wait(ctx context.Context) error {
	select {
	case <-m.shutdown:
		m.logger.Info("Starting shutdown sequence")
		return m.executeTasks()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *Manager) executeTasks() error {
	_, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	for name, task := range m.tasks {
		m.logger.Info("Executing shutdown task %s", name)
		if err := task(); err != nil {
			m.logger.Error("Task failed, task: %s, error: %s", name, err)
		}
	}

	return nil
}
