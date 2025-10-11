// Package adapters provides logger adapters for popular logging frameworks.
package adapters

import (
	"context"
	"log/slog"

	"github.com/itsatony/go-cuserr"
)

// SlogAdapter wraps a *slog.Logger to implement cuserr.Logger interface.
// This adapter allows using Go's standard library slog with cuserr errors.
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new slog adapter.
// If logger is nil, slog.Default() will be used.
func NewSlogAdapter(logger *slog.Logger) *SlogAdapter {
	if logger == nil {
		logger = slog.Default()
	}
	return &SlogAdapter{logger: logger}
}

// Error logs an error-level message with structured fields
func (s *SlogAdapter) Error(ctx context.Context, msg string, fields ...cuserr.Field) {
	s.logger.ErrorContext(ctx, msg, fieldsToSlogAttrs(fields)...)
}

// Warn logs a warning-level message with structured fields
func (s *SlogAdapter) Warn(ctx context.Context, msg string, fields ...cuserr.Field) {
	s.logger.WarnContext(ctx, msg, fieldsToSlogAttrs(fields)...)
}

// Info logs an info-level message with structured fields
func (s *SlogAdapter) Info(ctx context.Context, msg string, fields ...cuserr.Field) {
	s.logger.InfoContext(ctx, msg, fieldsToSlogAttrs(fields)...)
}

// Debug logs a debug-level message with structured fields
func (s *SlogAdapter) Debug(ctx context.Context, msg string, fields ...cuserr.Field) {
	s.logger.DebugContext(ctx, msg, fieldsToSlogAttrs(fields)...)
}

// fieldsToSlogAttrs converts cuserr.Field slice to slog attributes
func fieldsToSlogAttrs(fields []cuserr.Field) []any {
	attrs := make([]any, 0, len(fields))
	for _, f := range fields {
		attrs = append(attrs, slog.Any(f.Key, f.Value))
	}
	return attrs
}
