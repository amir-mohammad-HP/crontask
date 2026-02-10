package docker

import (
	"context"
	"fmt"
	"runtime"

	"github.com/amir-mohammad-HP/crontask/pkg/logger"
	dockerClient "github.com/docker/docker/client"
)

// Get default Docker socket path based on OS
func getDefaultSocketPath(logger *logger.StdLogger) string {
	logger.Debug("get default docker socket for %s", runtime.GOOS)
	// Linux path
	var namedPipe = "unix:///var/run/docker.sock"
	if runtime.GOOS == "windows" {
		// Windows paths
		// Try WSL2 path first
		// wslPath := "unix://" + `\\wsl$\docker-desktop-data\version-pack-data\community\docker\docker.sock`

		// If using Docker Desktop with named pipe
		namedPipe = "npipe:////./pipe/docker_engine"
		// Try to detect which one exists
	} else if runtime.GOOS == "darwin" {
		// macOS path
		namedPipe = "unix:///var/run/docker.sock"
	}

	logger.Debug("docker default path on %s", namedPipe)
	return namedPipe
}

// Try alternative socket paths
func tryAlternativeSocketPaths(logger *logger.StdLogger) (*dockerClient.Client, error) {
	alternativePaths := []string{
		// Windows paths
		"npipe:////./pipe/docker_engine",
		"unix://" + `\\wsl$\docker-desktop-data\version-pack-data\community\docker\docker.sock`,
		"unix://" + `\\wsl.localhost\docker-desktop-data\version-pack-data\community\docker\docker.sock`,
		"unix:///var/run/docker.sock", // Fallback for WSL

		// Linux/macOS paths
		"unix:///var/run/docker.sock",
	}

	var lastErr error
	for _, path := range alternativePaths {
		logger.Info("try alternative docker sockets: %s", path)
		cli, err := dockerClient.NewClientWithOpts(
			dockerClient.WithHost(path),
			dockerClient.WithAPIVersionNegotiation(),
		)
		if err != nil {
			lastErr = err
			logger.Error("no docker client with alternative path %s", path)
			continue
		}

		// Test connection
		_, err = cli.Ping(context.Background())
		if err != nil {
			cli.Close()
			lastErr = err

			logger.Error("no docker client connecttion using alternative path %s", path)
			continue
		}

		return cli, nil
	}

	return nil, fmt.Errorf("failed to connect to Docker using any socket path. Last error: %w", lastErr)
}
