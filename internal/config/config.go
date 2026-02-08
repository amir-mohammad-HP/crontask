// config/config.go
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/amir-mohammad-HP/crontask/internal/types"
	"github.com/spf13/viper"
)

// Default configuration values
var defaultConfig = types.Config{
	AppName:     "CronTask",
	Environment: "development",
	LogLevel:    "info",
	Worker: types.WorkerConfig{
		Interval:      5 * time.Second,
		MaxJobs:       10,
		RetryAttempts: 3,
	},
	Docker: types.DockerConfig{
		Enabled:      true,
		SocketPath:   "/var/run/docker.sock",
		PollInterval: 5 * time.Second,
		LabelPrefix:  "crontask.",
	},
	Shutdown: types.ShutdownConfig{
		Timeout: 30 * time.Second,
	},
	Logger: types.LoggerConfig{
		Level:           "info",
		Format:          "text",
		Output:          "stdout",
		FilePath:        "",
		MaxSize:         10,
		MaxBackups:      5,
		MaxAge:          30,
		Compress:        true,
		TimestampFormat: "2006-01-02 15:04:05.000",
		ShowCaller:      false,
		Colors:          true,
		Async:           false,
		BufferSize:      1000,
	},
}

// getSystemConfigPath returns the OS-specific configuration directory
func getSystemConfigPath() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %PROGRAMDATA%\crontask
		programData := os.Getenv("PROGRAMDATA")
		if programData == "" {
			programData = "C:\\ProgramData"
		}
		configDir = filepath.Join(programData, "crontask")

	case "darwin":
		// macOS: /Library/Application Support/crontask
		configDir = "/Library/Application Support/crontask"

	case "linux", "freebsd", "openbsd", "netbsd":
		// Unix-like: /etc/crontask
		configDir = "/etc/crontask"

	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return configDir, nil
}

// getConfigPaths returns all possible configuration file paths in order of precedence
func getConfigPaths() ([]string, error) {
	systemConfigDir, err := getSystemConfigPath()
	if err != nil {
		return nil, err
	}

	// Configuration search paths in order of precedence (first found wins):
	paths := []string{}

	// 1. Current directory (for development and testing)
	paths = append(paths, "crontaskd.yaml")

	// 2. User's home directory (~/.config/crontask/)
	if home, err := os.UserHomeDir(); err == nil {
		userConfigDir := filepath.Join(home, ".config", "crontask")
		paths = append(paths, filepath.Join(userConfigDir, "crontaskd.yaml"))
	}

	// 3. System-wide configuration directory
	paths = append(paths, filepath.Join(systemConfigDir, "crontaskd.yaml"))

	return paths, nil
}

// Load loads configuration from file, environment variables, or defaults
func Load() (*types.Config, error) {
	viper.SetConfigName("crontaskd") // Name of config file (without extension)
	viper.SetConfigType("yaml")      // REQUIRED if the config file does not have the extension in the name

	// Set default values
	viper.SetDefault("app_name", defaultConfig.AppName)
	viper.SetDefault("environment", defaultConfig.Environment)
	viper.SetDefault("log_level", defaultConfig.LogLevel)
	viper.SetDefault("worker.interval", defaultConfig.Worker.Interval)
	viper.SetDefault("worker.max_jobs", defaultConfig.Worker.MaxJobs)
	viper.SetDefault("worker.retry_attempts", defaultConfig.Worker.RetryAttempts)
	viper.SetDefault("shutdown.timeout", defaultConfig.Shutdown.Timeout)

	// Add configuration paths
	configPaths, err := getConfigPaths()
	if err != nil {
		return nil, fmt.Errorf("failed to get config paths: %w", err)
	}

	for _, path := range configPaths {
		viper.AddConfigPath(filepath.Dir(path))
	}

	// Try to read configuration file
	if err := viper.ReadInConfig(); err != nil {
		// If file doesn't exist, we'll use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Bind environment variables
	viper.SetEnvPrefix("CRONTASK") // Environment variables will be prefixed with CRONTASK_
	viper.AutomaticEnv()           // Automatically override config with environment variables

	// Unmarshal configuration into struct
	var cfg types.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// GetConfigFileLocation returns the location of the currently loaded config file
func GetConfigFileLocation() string {
	return viper.ConfigFileUsed()
}

// GetSystemConfigDir returns the system-wide configuration directory
func GetSystemConfigDir() (string, error) {
	return getSystemConfigPath()
}

// CreateDefaultConfig creates a default configuration file in the system config directory
func CreateDefaultConfig() error {
	systemConfigDir, err := getSystemConfigPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(systemConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(systemConfigDir, "crontaskd.yaml")

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return errors.New("config file already exists")
	}

	if err := os.WriteFile(configPath, []byte(DEFAULT_CONFIG_YAML), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
