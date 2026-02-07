package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestStdLogger_Levels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, "DEBUG")

	tests := []struct {
		name     string
		logFunc  func(string, ...any)
		expected string
	}{
		{"Debug", logger.Debug, "DEBUG"},
		{"Info", logger.Info, "INFO"},
		{"Warn", logger.Warn, "WARN"},
		{"Error", logger.Error, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc("test message")
			if !strings.Contains(buf.String(), tt.expected) {
				t.Errorf("Expected %s in log output, got: %s", tt.expected, buf.String())
			}
		})
	}
}

func TestStdLogger_WithField(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, "INFO")

	logger.WithField("request_id", "123").Info("processing request")

	output := buf.String()
	if !strings.Contains(output, "request_id=123") {
		t.Errorf("Expected field in output, got: %s", output)
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", DEBUG},
		{"INFO", INFO},
		{"warn", WARN},
		{"ERROR", ERROR},
		{"fatal", FATAL},
		{"unknown", INFO}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLogLevel(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
