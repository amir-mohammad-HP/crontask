package logger

import "context"

// ContextKey is a type for context keys
type ContextKey string

const (
	// LoggerKey is the key used to store logger in context
	LoggerKey ContextKey = "logger"
)

// FromContext extracts logger from context or returns a default one
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(LoggerKey).(Logger); ok {
		return logger
	}
	return DefaultLogger()
}

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}
