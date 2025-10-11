package cuserr

import "context"

// Logger is the minimal interface for logging errors.
// Libraries should inject this interface, not depend on specific implementations.
// This interface is designed to be zero-cost and easily adaptable to any logging framework.
type Logger interface {
	// Error logs an error-level message with structured fields
	Error(ctx context.Context, msg string, fields ...Field)

	// Warn logs a warning-level message with structured fields
	Warn(ctx context.Context, msg string, fields ...Field)

	// Info logs an info-level message with structured fields
	Info(ctx context.Context, msg string, fields ...Field)

	// Debug logs a debug-level message with structured fields
	Debug(ctx context.Context, msg string, fields ...Field)
}

// Field represents a structured logging field (key-value pair)
type Field struct {
	Key   string
	Value interface{}
}

// NoOpLogger is a logger that does nothing (useful as default for libraries)
type NoOpLogger struct{}

// Error is a no-op implementation that discards all error log messages
func (n NoOpLogger) Error(ctx context.Context, msg string, fields ...Field) {}

// Warn is a no-op implementation that discards all warning log messages
func (n NoOpLogger) Warn(ctx context.Context, msg string, fields ...Field) {}

// Info is a no-op implementation that discards all info log messages
func (n NoOpLogger) Info(ctx context.Context, msg string, fields ...Field) {}

// Debug is a no-op implementation that discards all debug log messages
func (n NoOpLogger) Debug(ctx context.Context, msg string, fields ...Field) {}
