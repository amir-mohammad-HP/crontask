package logger

import (
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"strings"
	"sync"
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

// String returns the string representation of the log level
func (l LogLevel) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[l]
}

// ParseLogLevel converts a string to LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO // Default level
	}
}

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
	logger *log.Logger
	level  LogLevel
	mu     sync.RWMutex
	fields map[string]any
	closer io.Closer
}

// New creates a new StdLogger with the specified log level
func New(level string) *StdLogger {
	logLevel := ParseLogLevel(level)

	return &StdLogger{
		logger: log.New(os.Stdout, "", 0), // No prefix here, we'll handle it in methods
		level:  logLevel,
		fields: make(map[string]any),
	}
}

// NewWithWriter creates a new StdLogger with custom writer
func NewWithWriter(w io.Writer, level string) *StdLogger {
	logLevel := ParseLogLevel(level)

	return &StdLogger{
		logger: log.New(w, "", 0),
		level:  logLevel,
		fields: make(map[string]any),
	}
}

// Debug logs a debug message
func (l *StdLogger) Debug(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.level <= DEBUG {
		l.log(DEBUG, msg, args...)
	}
}

// Info logs an info message
func (l *StdLogger) Info(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.level <= INFO {
		l.log(INFO, msg, args...)
	}
}

// Warn logs a warning message
func (l *StdLogger) Warn(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.level <= WARN {
		l.log(WARN, msg, args...)
	}
}

// Error logs an error message
func (l *StdLogger) Error(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.level <= ERROR {
		l.log(ERROR, msg, args...)
	}
}

// Fatal logs a fatal message and exits the program
func (l *StdLogger) Fatal(msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	l.log(FATAL, msg, args...)
	os.Exit(1)
}

// WithField returns a new logger with an additional field
func (l *StdLogger) WithField(key string, value any) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]any)
	for k, v := range l.fields {
		newFields[k] = v
	}

	newFields[key] = value

	return &StdLogger{
		logger: l.logger,
		level:  l.level,
		fields: newFields,
	}
}

// WithFields returns a new logger with additional fields
func (l *StdLogger) WithFields(fields map[string]any) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]any)
	maps.Copy(l.fields, newFields)
	maps.Copy(fields, newFields)

	return &StdLogger{
		logger: l.logger,
		level:  l.level,
		fields: newFields,
	}
}

// SetLevel changes the log level
func (l *StdLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level
func (l *StdLogger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// SetOutput changes the output writer
func (l *StdLogger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.SetOutput(w)
}

// log is the internal logging method
func (l *StdLogger) log(level LogLevel, msg string, args ...any) {
	// Format the message if there are arguments
	formattedMsg := msg
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(msg, args...)
	}

	// Build prefix with timestamp, level, and fields
	prefix := fmt.Sprintf("[%s] [%s] ", level, "CRONTASK")

	// Add fields if any
	if len(l.fields) > 0 {
		var sb strings.Builder
		for k, v := range l.fields {
			fmt.Fprintf(&sb, " %s=%v", k, v)
		}
		prefix += strings.TrimSpace(sb.String()) + " "
	}

	// Output the log
	l.logger.Printf("%s%s", prefix, formattedMsg)
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
