package logger

import (
	"io"
	"log"
	"sync"

	"github.com/amir-mohammad-HP/crontask/internal/types"
)

// LogLevel represents the severity level of log messages
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger interface defines the contract for all loggers
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)

	// Additional utility methods
	WithField(key string, value any) Logger
	WithFields(fields map[string]any) Logger
	SetLevel(level LogLevel)
	GetLevel() LogLevel
	SetOutput(w io.Writer)
}

// StdLogger implements Logger interface using Go's standard log package
type StdLogger struct {
	config *types.LoggerConfig
	logger *log.Logger
	level  LogLevel
	mu     sync.RWMutex
	fields map[string]any
	closer io.Closer
	writer io.Writer
	async  bool
	buffer chan logMessage
	quit   chan struct{}
	wg     sync.WaitGroup
}

type logMessage struct {
	level  LogLevel
	msg    string
	fields map[string]any
}
