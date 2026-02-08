package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
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

// syncLog performs synchronous logging with consistent format
func (l *StdLogger) syncLog(level LogLevel, msg string, fields map[string]any) {
	if l.config.Format == "json" {
		l.jsonLog(level, msg, fields)
	} else {
		l.textLog(level, msg, fields)
	}
}

// jsonLog outputs log in pure JSON format
func (l *StdLogger) jsonLog(level LogLevel, msg string, fields map[string]any) {
	logEntry := map[string]any{
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"level":     level.String(),
		"message":   msg,
	}

	// Add caller information if enabled
	if l.config.ShowCaller {
		if _, file, line, ok := runtime.Caller(4); ok { // 4 for jsonLog -> syncLog -> original call
			logEntry["caller"] = fmt.Sprintf("%s:%d", filepath.Base(file), line)
		}
	}

	// Add all fields to the log entry
	for k, v := range fields {
		logEntry[k] = v
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(logEntry)
	if err != nil {
		// Fallback to text format on JSON error
		l.textLog(level, fmt.Sprintf("JSON marshal error: %v - %s", err, msg), fields)
		return
	}

	// Output with proper newline
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(l.writer, string(jsonBytes))
}

// textLog outputs log in human-readable text format
func (l *StdLogger) textLog(level LogLevel, msg string, fields map[string]any) {
	var sb strings.Builder

	// Add timestamp
	timestamp := time.Now().Format(l.config.TimestampFormat)
	sb.WriteString(timestamp)
	sb.WriteString(" ")

	// Add level with color if enabled
	if l.config.Colors && (l.config.Output == "stdout" || l.config.Output == "stderr") {
		sb.WriteString(l.colorizeLevel(level))
	} else {
		sb.WriteString(fmt.Sprintf("[%s]", level))
	}
	// Add padding to reach 10 characters
	padding := 8 - len(level.String())
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}

	// Add caller information if enabled
	if l.config.ShowCaller {
		if _, file, line, ok := runtime.Caller(4); ok { // 4 for textLog -> syncLog -> original call
			callerFilePath := filepath.Base(file)
			sb.WriteString(fmt.Sprintf("%s:%d ", callerFilePath, line))

			padding := 15 - len(callerFilePath) + len(strconv.Itoa(line))
			if padding > 0 {
				sb.WriteString(strings.Repeat(" ", padding))
			}
		}
	}

	// Add fields as key=value pairs
	if len(fields) > 0 {
		// Sort keys for consistent output
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := fields[k]
			// Handle string values with spaces by quoting them
			switch val := v.(type) {
			case string:
				if strings.ContainsAny(val, " \t\n\"") {
					sb.WriteString(fmt.Sprintf("%s=%q ", k, val))
				} else {
					sb.WriteString(fmt.Sprintf("%s=%s ", k, val))
				}
			default:
				sb.WriteString(fmt.Sprintf("%s=%v ", k, v))
			}
		}
	}

	// Add message
	sb.WriteString(msg)

	// Output with proper newline
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(l.writer, sb.String())
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
