package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// Helper functions for common logger operations

// NewFileLogger creates a logger that writes to a file
func NewFileLogger(filename, level string) (*StdLogger, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// Open or create the log file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return NewWithWriter(file, level), nil
}

// NewMultiWriterLogger creates a logger that writes to multiple outputs
func NewMultiWriterLogger(level string, writers ...io.Writer) *StdLogger {
	multiWriter := io.MultiWriter(writers...)
	return NewWithWriter(multiWriter, level)
}

// DefaultLogger returns a pre-configured logger with reasonable defaults
func DefaultLogger() *StdLogger {
	logger := New("INFO")

	// Use the standard log format (date time)
	logger.logger.SetFlags(log.LstdFlags)

	return logger
}
