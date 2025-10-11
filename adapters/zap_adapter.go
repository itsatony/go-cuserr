package adapters

import (
	"context"

	"github.com/itsatony/go-cuserr"
	"go.uber.org/zap"
)

// ZapAdapter wraps a *zap.Logger to implement cuserr.Logger interface.
// This adapter allows using Uber's Zap logger with cuserr errors.
//
// To use this adapter, you need to install Zap:
//
//	go get go.uber.org/zap
type ZapAdapter struct {
	logger *zap.Logger
}

// NewZapAdapter creates a new Zap adapter.
// If logger is nil, zap.NewNop() will be used (no-op logger).
func NewZapAdapter(logger *zap.Logger) *ZapAdapter {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ZapAdapter{logger: logger}
}

// Error logs an error-level message with structured fields
func (z *ZapAdapter) Error(ctx context.Context, msg string, fields ...cuserr.Field) {
	z.logger.Error(msg, fieldsToZapFields(fields)...)
}

// Warn logs a warning-level message with structured fields
func (z *ZapAdapter) Warn(ctx context.Context, msg string, fields ...cuserr.Field) {
	z.logger.Warn(msg, fieldsToZapFields(fields)...)
}

// Info logs an info-level message with structured fields
func (z *ZapAdapter) Info(ctx context.Context, msg string, fields ...cuserr.Field) {
	z.logger.Info(msg, fieldsToZapFields(fields)...)
}

// Debug logs a debug-level message with structured fields
func (z *ZapAdapter) Debug(ctx context.Context, msg string, fields ...cuserr.Field) {
	z.logger.Debug(msg, fieldsToZapFields(fields)...)
}

// fieldsToZapFields converts cuserr.Field slice to zap.Field slice
func fieldsToZapFields(fields []cuserr.Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	return zapFields
}
