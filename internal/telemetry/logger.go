package telemetry

import (
	"log"
	"os"
)

// Logger exposes a minimal structured logger interface to keep the MVP lightweight.
type Logger interface {
	Info(msg string, kv ...any)
	Error(msg string, err error, kv ...any)
}

// StdLogger implements Logger using the standard library log package.
type StdLogger struct {
	logger *log.Logger
}

// NewStdLogger constructs a standard output logger with timestamps.
func NewStdLogger() *StdLogger {
	return &StdLogger{logger: log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)}
}

// Info logs informational messages with optional key/value pairs.
func (l *StdLogger) Info(msg string, kv ...any) {
	l.logger.Println(append([]any{"level", "info", "msg", msg}, kv...)...)
}

// Error logs error messages with optional key/value pairs.
func (l *StdLogger) Error(msg string, err error, kv ...any) {
	l.logger.Println(append([]any{"level", "error", "msg", msg, "error", err}, kv...)...)
}
