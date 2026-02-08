package logger

import (
	"fmt"
	"strings"
)

// String returns the string representation of the log level

// Helper method to get level string
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
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

// colorizeLevel adds ANSI color codes to log level
func (l *StdLogger) colorizeLevel(level LogLevel) string {
	var colorCode string
	switch level {
	case DEBUG:
		colorCode = "\033[36m" // Cyan
	case INFO:
		colorCode = "\033[32m" // Green
	case WARN:
		colorCode = "\033[33m" // Yellow
	case ERROR:
		colorCode = "\033[31m" // Red
	case FATAL:
		colorCode = "\033[35m" // Magenta
	default:
		colorCode = "\033[0m" // Reset
	}
	return fmt.Sprintf("%s[%s]\033[0m", colorCode, level.String())
}
