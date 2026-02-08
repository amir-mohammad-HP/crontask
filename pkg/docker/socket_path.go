package docker

import (
	"context"
	"fmt"
	"runtime"

	dockerClient "github.com/docker/docker/client"
)

// Get default Docker socket path based on OS
func getDefaultSocketPath() string {
	if runtime.GOOS == "windows" {
		// Windows paths
		// Try WSL2 path first
		// wslPath := "unix://" + `\\wsl$\docker-desktop-data\version-pack-data\community\docker\docker.sock`

		// If using Docker Desktop with named pipe
		namedPipe := "npipe:////./pipe/docker_engine"

		// Try to detect which one exists
		return namedPipe // Docker Desktop usually uses named pipe
	} else if runtime.GOOS == "darwin" {
		// macOS path
		return "unix:///var/run/docker.sock"
	} else {
		// Linux path
		return "unix:///var/run/docker.sock"
	}
}

// Try alternative socket paths
func tryAlternativeSocketPaths() (*dockerClient.Client, error) {
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
		cli, err := dockerClient.NewClientWithOpts(
			dockerClient.WithHost(path),
			dockerClient.WithAPIVersionNegotiation(),
		)
		if err != nil {
			lastErr = err
			continue
		}

		// Test connection
		_, err = cli.Ping(context.Background())
		if err != nil {
			cli.Close()
			lastErr = err
			continue
		}

		return cli, nil
	}

	return nil, fmt.Errorf("failed to connect to Docker using any socket path. Last error: %w", lastErr)
}
