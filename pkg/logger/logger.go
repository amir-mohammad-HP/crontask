package logger

import (
	"log"
	"os"
)

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type StdLogger struct {
	logger *log.Logger
	level  string
}

func New(level string) *StdLogger {
	return &StdLogger{
		logger: log.New(os.Stdout, "[CRONTASK] ", log.LstdFlags),
		level:  level,
	}
}

func (l *StdLogger) Info(msg string, args ...interface{}) {
	l.logger.Printf("INFO: "+msg, args...)
}

func (l *StdLogger) Error(msg string, args ...interface{}) {
	l.logger.Printf("ERROR: "+msg, args...)
}
