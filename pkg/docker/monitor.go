// pkg/docker/monitor.go
package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/amir-mohammad-HP/crontask/internal/types"
	"github.com/amir-mohammad-HP/crontask/pkg/logger"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerEvents "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
)

// Event types for communication
type ContainerEvent struct {
	Action      string
	ContainerID string
	Container   *ContainerInfo
}

type ContainerInfo struct {
	ID      string
	Name    string
	State   string
	Image   string
	Labels  map[string]string
	Created time.Time
}

type DockerMonitor struct {
	client     *dockerClient.Client
	logger     *logger.StdLogger
	config     *types.DockerConfig
	eventsChan chan ContainerEvent
	stopChan   chan struct{}
}

func NewMonitor(config *types.DockerConfig, logger *logger.StdLogger) (*DockerMonitor, error) {
	var cli *dockerClient.Client
	var err error

	if config.SocketPath != "" {
		cli, err = dockerClient.NewClientWithOpts(
			dockerClient.WithHost("unix://"+config.SocketPath),
			dockerClient.WithAPIVersionNegotiation(),
		)
	} else {
		// Use platform-specific default socket
		socketPath := getDefaultSocketPath(logger)
		cli, err = dockerClient.NewClientWithOpts(
			dockerClient.WithHost(socketPath),
			dockerClient.WithAPIVersionNegotiation(),
		)
	}

	if err != nil {
		error := fmt.Errorf("failed to create docker client: %w", err)
		logger.Error("%s", error.Error())
		return nil, error
	}

	// Test connection
	_, err = cli.Ping(context.Background())
	if err != nil {
		logger.Warn("Docker connection test failed %s", err.Error())
		logger.Info("Trying alternative Docker socket paths...")

		// Try alternative paths
		cli, err = tryAlternativeSocketPaths(logger)
		if err != nil {
			error := fmt.Errorf("failed to connect to Docker: %w", err)
			logger.Error("%s", error.Error())
			return nil, error
		}
	}

	return &DockerMonitor{
		client:     cli,
		logger:     logger,
		config:     config,
		eventsChan: make(chan ContainerEvent, 100),
		stopChan:   make(chan struct{}),
	}, nil
}

// Start monitoring Docker events
func (dm *DockerMonitor) Start(ctx context.Context) error {
	dm.logger.Debug("Starting Docker monitor")

	// Initial scan of existing containers
	if err := dm.scanExistingContainers(); err != nil {
		dm.logger.Error("Failed to scan existing containers %s", err.Error())
	}

	// Start event monitoring
	go dm.monitorEvents(ctx)

	return nil
}

// Stop monitoring
func (dm *DockerMonitor) Stop() {
	close(dm.stopChan)
	if dm.client != nil {
		dm.client.Close()
	}
}

// GetEvents returns a channel to receive container events
func (dm *DockerMonitor) GetEvents() <-chan ContainerEvent {
	return dm.eventsChan
}

// Scan all existing containers
func (dm *DockerMonitor) scanExistingContainers() error {
	containers, err := dm.client.ContainerList(context.Background(), container.ListOptions{
		All: true,
	})
	if err != nil {
		return err
	}

	for _, c := range containers {
		containerInfo, err := dm.getContainerInfo(c.ID)
		if err != nil {
			dm.logger.Error("Failed to get container info | %s: %s , %s",
				"container", c.ID[:12],
				err.Error())
			continue
		}

		dm.eventsChan <- ContainerEvent{
			Action:      "scan",
			ContainerID: c.ID,
			Container:   containerInfo,
		}
	}

	return nil
}

// Monitor Docker events in real-time
func (dm *DockerMonitor) monitorEvents(ctx context.Context) {
	filter := filters.NewArgs()
	filter.Add("type", "container")
	filter.Add("event", "create")
	filter.Add("event", "start")
	filter.Add("event", "die")
	filter.Add("event", "destroy")
	filter.Add("event", "update")

	eventsChan, errs := dm.client.Events(ctx, dockerTypes.EventsOptions{
		Filters: filter,
	})

	for {
		select {
		case event := <-eventsChan:
			dm.handleEvent(event)
		case err := <-errs:
			if err != nil {
				dm.logger.Error("Docker events error %s", err.Error())
			}
		case <-ctx.Done():
			dm.logger.Debug("Docker monitor context done")
			return
		case <-dm.stopChan:
			dm.logger.Debug("Docker monitor stopped")
			return
		}
	}
}

// Handle individual Docker events
func (dm *DockerMonitor) handleEvent(event dockerEvents.Message) {
	// Give container a moment to fully start
	time.Sleep(500 * time.Millisecond)

	containerInfo, err := dm.getContainerInfo(event.Actor.ID)
	if err != nil {
		dm.logger.Error("Failed to get container info after event | %s: %s, %s: %s, %s",
			"action", string(event.Action), // Convert to string
			"container", event.Actor.ID[:12],
			err.Error())
		return
	}

	dm.eventsChan <- ContainerEvent{
		Action:      string(event.Action), // Convert to string
		ContainerID: event.Actor.ID,
		Container:   containerInfo,
	}
}

// Get detailed container information
func (dm *DockerMonitor) getContainerInfo(containerID string) (*ContainerInfo, error) {
	containerJSON, err := dm.client.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return nil, err
	}

	// Parse the Created time string (ISO 8601 format)
	var createdTime time.Time
	if containerJSON.Created != "" {
		parsedTime, err := time.Parse(time.RFC3339Nano, containerJSON.Created)
		if err != nil {
			// Try parsing without nanoseconds if the first attempt fails
			parsedTime, err = time.Parse(time.RFC3339, containerJSON.Created)
			if err != nil {
				dm.logger.Warn("Failed to parse container creation time | %s: %s, %s: %s, %s",
					"container", containerID[:12],
					"created", containerJSON.Created,
					err)
				createdTime = time.Now() // Fallback to current time
			} else {
				createdTime = parsedTime
			}
		} else {
			createdTime = parsedTime
		}
	}

	return &ContainerInfo{
		ID:      containerJSON.ID,
		Name:    strings.TrimPrefix(containerJSON.Name, "/"),
		State:   containerJSON.State.Status,
		Image:   containerJSON.Config.Image,
		Labels:  containerJSON.Config.Labels,
		Created: createdTime,
	}, nil
}

// Extract cron jobs from container labels
func (dm *DockerMonitor) ExtractCronJobs(container *ContainerInfo) []types.CronJob {
	var cronJobs []types.CronJob

	for labelKey, task := range container.Labels {
		if strings.HasPrefix(labelKey, dm.config.LabelPrefix) {
			cronExpr, err := dm.parseCronExpression(labelKey)
			if err != nil {
				dm.logger.Warn("Failed to parse cron expression | %s: %s, %s",
					"label", labelKey,
					err.Error())
				continue
			}

			cronJobs = append(cronJobs, types.CronJob{
				ContainerID:   container.ID,
				ContainerName: container.Name,
				CronExpr:      cronExpr,
				Task:          task,
				LabelKey:      labelKey,
				IsActive:      container.State == "running",
				CreatedAt:     time.Now(),
			})
		}
	}

	return cronJobs
}

// Parse cron expression from label key
func (dm *DockerMonitor) parseCronExpression(labelKey string) (string, error) {
	// Expected format: prefix.cronjob('* * * * *').task
	start := strings.Index(labelKey, "('")
	if start == -1 {
		return "", fmt.Errorf("invalid cron job format: missing (")
	}

	end := strings.Index(labelKey, "')")
	if end == -1 {
		return "", fmt.Errorf("invalid cron job format: missing )")
	}

	cronExpr := labelKey[start+2 : end]

	// Validate basic cron format (at least 5 fields)
	parts := strings.Fields(cronExpr)
	if len(parts) < 5 {
		return "", fmt.Errorf("invalid cron expression: %s", cronExpr)
	}

	return cronExpr, nil
}

// Execute a task inside a container
func (dm *DockerMonitor) ExecuteTask(containerID string, task string) (string, error) {
	// Create exec instance
	execConfig := dockerTypes.ExecConfig{
		Cmd:          []string{"sh", "-c", task},
		AttachStdout: true,
		AttachStderr: true,
	}

	execID, err := dm.client.ContainerExecCreate(context.Background(), containerID, execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create exec: %w", err)
	}

	// Attach to exec to get output
	resp, err := dm.client.ContainerExecAttach(context.Background(), execID.ID, dockerTypes.ExecStartCheck{})
	if err != nil {
		return "", fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer resp.Close()

	// Read output
	buf := make([]byte, 4096)
	n, err := resp.Reader.Read(buf)
	if err != nil && err.Error() != "EOF" {
		error := fmt.Errorf("failed to read output: %w", err)
		dm.logger.Error("%s", error.Error())
		return "", error
	}

	output := string(buf[:n])

	// Check exec status
	inspect, err := dm.client.ContainerExecInspect(context.Background(), execID.ID)
	if err != nil {
		return output, fmt.Errorf("failed to inspect exec: %w", err)
	}

	if inspect.ExitCode != 0 {
		return output, fmt.Errorf("task exited with code %d", inspect.ExitCode)
	}

	return output, nil
}

// Get all running containers with cron labels
func (dm *DockerMonitor) GetContainersWithCronJobs() ([]ContainerInfo, error) {
	containers, err := dm.client.ContainerList(context.Background(), container.ListOptions{
		All: false, // Only running containers
	})
	if err != nil {
		return nil, err
	}

	var result []ContainerInfo
	for _, c := range containers {
		containerInfo, err := dm.getContainerInfo(c.ID)
		if err != nil {
			continue
		}

		// Check if container has cron labels
		for labelKey := range containerInfo.Labels {
			if strings.HasPrefix(labelKey, dm.config.LabelPrefix) {
				result = append(result, *containerInfo)
				break
			}
		}
	}

	return result, nil
}
