package logger

import (
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/amir-mohammad-HP/crontask/internal/types"
)

// New creates a new StdLogger with the specified log level
func New(level string) *StdLogger {
	config := DefaultConfig()
	config.Level = level
	return NewWithConfig(config)
}

func NewWithConfig(config *types.LoggerConfig) *StdLogger {
	// Ensure file path exists if using file output
	if config.Output == "file" && config.FilePath == "" {
		config.FilePath = getDefaultLogPath()
	}

	writer := createWriter(config)
	logLevel := ParseLogLevel(config.Level)

	logger := &StdLogger{
		config: config,
		logger: log.New(writer, "", 0),
		level:  logLevel,
		fields: make(map[string]any),
		writer: writer,
		async:  config.Async,
	}

	// Initialize async logging if enabled
	if config.Async {
		logger.buffer = make(chan logMessage, config.BufferSize)
		logger.quit = make(chan struct{})
		logger.wg.Add(1)
		go logger.asyncWriter()
	}

	return logger
}

// Debug logs a debug message
func (l *StdLogger) Debug(msg string, args ...any) {
	l.log(DEBUG, msg, args...)
}

// Info logs an info message
func (l *StdLogger) Info(msg string, args ...any) {
	l.log(INFO, msg, args...)
}

// Warn logs a warning message
func (l *StdLogger) Warn(msg string, args ...any) {
	l.log(WARN, msg, args...)
}

// Error logs an error message
func (l *StdLogger) Error(msg string, args ...any) {
	l.log(ERROR, msg, args...)
}

// Fatal logs a fatal message and exits the program
func (l *StdLogger) Fatal(msg string, args ...any) {
	l.log(FATAL, msg, args...)
	os.Exit(1)
}

// WithField returns a new logger with an additional field
func (l *StdLogger) WithField(key string, value any) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]any)
	maps.Copy(newFields, l.fields)
	newFields[key] = value

	return &StdLogger{
		config: l.config,
		logger: l.logger,
		level:  l.level,
		fields: newFields,
		writer: l.writer,
		async:  l.async,
	}
}

// WithFields returns a new logger with additional fields
func (l *StdLogger) WithFields(fields map[string]any) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]any)
	maps.Copy(newFields, l.fields)
	maps.Copy(newFields, fields)

	return &StdLogger{
		config: l.config,
		logger: l.logger,
		level:  l.level,
		fields: newFields,
		writer: l.writer,
		async:  l.async,
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
	l.writer = w
}

// log is the internal logging method
func (l *StdLogger) log(level LogLevel, msg string, args ...any) {
	// Check if we should log at this level
	if level < l.GetLevel() {
		return
	}

	// Format the message if there are arguments
	formattedMsg := msg
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(msg, args...)
	}

	// Prepare fields
	fields := make(map[string]any)
	l.mu.RLock()
	maps.Copy(fields, l.fields)
	l.mu.RUnlock()

	// Handle async or sync logging
	if l.async {
		select {
		case l.buffer <- logMessage{level: level, msg: formattedMsg, fields: fields}:
			// Message queued
		default:
			// Buffer full, fallback to sync logging
			l.syncLog(level, formattedMsg, fields)
		}
	} else {
		l.syncLog(level, formattedMsg, fields)
	}
}

// syncLog performs synchronous logging
func (l *StdLogger) syncLog(level LogLevel, msg string, fields map[string]any) {
	var sb strings.Builder

	// Add timestamp
	timestamp := time.Now().Format(l.config.TimestampFormat)
	sb.WriteString(timestamp)
	sb.WriteString(" ")

	// Add level with color if enabled
	if l.config.Colors && (l.config.Output == "stdout" || l.config.Output == "stderr") {
		sb.WriteString(l.colorizeLevel(level))
	} else {
		fmt.Fprintf(l.writer, "[%s]", level)
	}
	sb.WriteString(" ")

	// Add caller information if enabled
	if l.config.ShowCaller {
		if _, file, line, ok := runtime.Caller(3); ok {
			fmt.Fprintf(l.writer, "%s:%d ", filepath.Base(file), line)
		}
	}

	// Add fields in JSON format
	if len(fields) > 0 && l.config.Format == "json" {
		fieldStr := ""
		for k, v := range fields {
			fieldStr += fmt.Sprintf("\"%s\":\"%v\",", k, v)
		}
		if len(fieldStr) > 0 {
			fieldStr = fieldStr[:len(fieldStr)-1] // Remove trailing comma
			fmt.Fprintf(l.writer, "{%s} ", fieldStr)
		}
	} else if len(fields) > 0 {
		for k, v := range fields {
			fmt.Fprintf(l.writer, "%s=%v ", k, v)
		}
	}

	// Add message
	sb.WriteString(msg)

	// Output the log
	l.logger.Println(sb.String())
}

// colorizeLevel adds ANSI colors to log levels
func (l *StdLogger) colorizeLevel(level LogLevel) string {
	switch level {
	case DEBUG:
		return "\033[36m[DEBUG]\033[0m" // Cyan
	case INFO:
		return "\033[32m[INFO]\033[0m" // Green
	case WARN:
		return "\033[33m[WARN]\033[0m" // Yellow
	case ERROR:
		return "\033[31m[ERROR]\033[0m" // Red
	case FATAL:
		return "\033[35m[FATAL]\033[0m" // Magenta
	default:
		return fmt.Sprintf("[%s]", level)
	}
}

// asyncWriter handles async log writing
func (l *StdLogger) asyncWriter() {
	defer l.wg.Done()

	for {
		select {
		case msg := <-l.buffer:
			l.syncLog(msg.level, msg.msg, msg.fields)
		case <-l.quit:
			// Drain remaining messages
			for {
				select {
				case msg := <-l.buffer:
					l.syncLog(msg.level, msg.msg, msg.fields)
				default:
					return
				}
			}
		}
	}
}

// Close gracefully shuts down the logger
func (l *StdLogger) Close() error {
	if l.async {
		close(l.quit)
		l.wg.Wait()
	}

	if l.closer != nil {
		return l.closer.Close()
	}
	return nil
}
