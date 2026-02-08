package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/amir-mohammad-HP/crontask/internal/types"
)

// createWriter creates the appropriate writer based on configuration
func createWriter(config *types.LoggerConfig) io.Writer {
	switch config.Output {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	case "file":
		// Create directory if it doesn't exist
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Warning: Failed to create log directory %s: %v", dir, err)
			return os.Stdout
		}

		// Open or create log file
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("Warning: Failed to open log file %s: %v", config.FilePath, err)
			return os.Stdout
		}

		return file
	case "null":
		return io.Discard
	default:
		return os.Stdout
	}
}

// NullLogger is a logger that discards all messages (useful for testing)
type NullLogger struct{}

func (n *NullLogger) Debug(msg string, args ...any)           {}
func (n *NullLogger) Info(msg string, args ...any)            {}
func (n *NullLogger) Warn(msg string, args ...any)            {}
func (n *NullLogger) Error(msg string, args ...any)           {}
func (n *NullLogger) Fatal(msg string, args ...any)           {}
func (n *NullLogger) WithField(key string, value any) Logger  { return n }
func (n *NullLogger) WithFields(fields map[string]any) Logger { return n }
func (n *NullLogger) SetLevel(level LogLevel)                 {}
func (n *NullLogger) GetLevel() LogLevel                      { return INFO }
func (n *NullLogger) SetOutput(w io.Writer)                   {}

// NewNullLogger creates a logger that discards all output
func NewNullLogger() *NullLogger {
	return &NullLogger{}
}
