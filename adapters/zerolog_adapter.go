package adapters

import (
	"context"

	"github.com/itsatony/go-cuserr"
	"github.com/rs/zerolog"
)

// ZerologAdapter wraps a zerolog.Logger to implement cuserr.Logger interface.
// This adapter allows using Zerolog with cuserr errors.
//
// To use this adapter, you need to install Zerolog:
//
//	go get github.com/rs/zerolog
type ZerologAdapter struct {
	logger zerolog.Logger
}

// NewZerologAdapter creates a new Zerolog adapter.
func NewZerologAdapter(logger zerolog.Logger) *ZerologAdapter {
	return &ZerologAdapter{logger: logger}
}

// Error logs an error-level message with structured fields
func (z *ZerologAdapter) Error(ctx context.Context, msg string, fields ...cuserr.Field) {
	event := z.logger.Error()
	for _, f := range fields {
		event = event.Interface(f.Key, f.Value)
	}
	event.Msg(msg)
}

// Warn logs a warning-level message with structured fields
func (z *ZerologAdapter) Warn(ctx context.Context, msg string, fields ...cuserr.Field) {
	event := z.logger.Warn()
	for _, f := range fields {
		event = event.Interface(f.Key, f.Value)
	}
	event.Msg(msg)
}

// Info logs an info-level message with structured fields
func (z *ZerologAdapter) Info(ctx context.Context, msg string, fields ...cuserr.Field) {
	event := z.logger.Info()
	for _, f := range fields {
		event = event.Interface(f.Key, f.Value)
	}
	event.Msg(msg)
}

// Debug logs a debug-level message with structured fields
func (z *ZerologAdapter) Debug(ctx context.Context, msg string, fields ...cuserr.Field) {
	event := z.logger.Debug()
	for _, f := range fields {
		event = event.Interface(f.Key, f.Value)
	}
	event.Msg(msg)
}
