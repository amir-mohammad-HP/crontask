package logger

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/amir-mohammad-HP/crontask/internal/types"
)

// DefaultConfig returns the default logger configuration
func DefaultConfig() *types.LoggerConfig {
	return &types.LoggerConfig{
		Level:           "info",
		Format:          "text",
		Output:          "stdout",
		FilePath:        "",
		MaxSize:         10, // MB
		MaxBackups:      5,
		MaxAge:          30, // days
		Compress:        true,
		TimestampFormat: "2006-01-02 15:04:05.000",
		ShowCaller:      false,
		Colors:          true,
		Async:           false,
		BufferSize:      1000,
	}
}

// getDefaultLogPath returns the default log file path based on OS
func getDefaultLogPath() string {
	switch runtime.GOOS {
	case "windows":
		programData := os.Getenv("PROGRAMDATA")
		if programData == "" {
			programData = "C:\\ProgramData"
		}
		return filepath.Join(programData, "crontask", "logs", "crontaskd.log")
	case "darwin":
		return "/var/log/crontaskd.log"
	default: // linux and other unix-like
		return "/var/log/crontaskd.log"
	}
}
