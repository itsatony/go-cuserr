// Package cuserr provides context-based error configuration and handling.
// This file contains context-aware error creation and configuration utilities.
package cuserr

import (
	"context"
)

// Context keys for configuration and error handling
type contextKey string

const (
	// ConfigContextKey is the context key for error configuration
	ConfigContextKey contextKey = "cuserr_config"
	// ErrorHandlerContextKey is the context key for custom error handlers
	ErrorHandlerContextKey contextKey = "cuserr_error_handler"
)

// ContextConfig holds error configuration that can be passed via context
type ContextConfig struct {
	// EnableStackTrace controls whether to capture stack traces
	EnableStackTrace bool
	// MaxStackDepth controls how many stack frames to capture
	MaxStackDepth int
	// ProductionMode controls error detail exposure
	ProductionMode bool
	// RequestID for this context (extracted automatically)
	RequestID string
	// UserID for this context (extracted automatically)
	UserID string
	// TraceID for distributed tracing (extracted automatically)
	TraceID string
}

// GetConfigFromContext returns configuration from context, falling back to global config
func GetConfigFromContext(ctx context.Context) *Config {
	if ctx == nil {
		return GetConfig()
	}

	if contextConfig, ok := ctx.Value(ConfigContextKey).(*ContextConfig); ok {
		return &Config{
			EnableStackTrace: contextConfig.EnableStackTrace,
			MaxStackDepth:    contextConfig.MaxStackDepth,
			ProductionMode:   contextConfig.ProductionMode,
		}
	}

	return GetConfig()
}

// WithConfig adds error configuration to context
func WithConfig(ctx context.Context, config *ContextConfig) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ConfigContextKey, config)
}

// WithStackTraceDisabled returns a context with stack trace disabled
func WithStackTraceDisabled(ctx context.Context) context.Context {
	config := GetConfigFromContext(ctx)
	contextConfig := &ContextConfig{
		EnableStackTrace: false,
		MaxStackDepth:    config.MaxStackDepth,
		ProductionMode:   config.ProductionMode,
	}
	return WithConfig(ctx, contextConfig)
}

// WithProductionMode returns a context with production mode enabled
func WithProductionMode(ctx context.Context) context.Context {
	config := GetConfigFromContext(ctx)
	contextConfig := &ContextConfig{
		EnableStackTrace: config.EnableStackTrace,
		MaxStackDepth:    config.MaxStackDepth,
		ProductionMode:   true,
	}
	return WithConfig(ctx, contextConfig)
}

// WithDevelopmentMode returns a context with development mode enabled
func WithDevelopmentMode(ctx context.Context) context.Context {
	config := GetConfigFromContext(ctx)
	contextConfig := &ContextConfig{
		EnableStackTrace: true,
		MaxStackDepth:    config.MaxStackDepth,
		ProductionMode:   false,
	}
	return WithConfig(ctx, contextConfig)
}

// Context-aware error creation functions

// NewErrorWithContext creates an error using context-based configuration
func NewErrorWithContext(ctx context.Context, sentinel error, wrapped error, message string) *CustomError {
	err := NewCustomError(sentinel, wrapped, message)

	// Apply context-specific configuration
	if ctx != nil {
		if contextConfig, ok := ctx.Value(ConfigContextKey).(*ContextConfig); ok {
			// Override stack trace behavior if specified in context
			if !contextConfig.EnableStackTrace {
				err.ClearStackTrace()
			}
		}

		// Extract and apply context values
		err = enrichFromContext(ctx, err)
	}

	return err
}

// NewValidationErrorFromContext creates a validation error using context
func NewValidationErrorFromContext(ctx context.Context, field, message string) *CustomError {
	err := NewValidationError(field, message)
	return enrichFromContext(ctx, err)
}

// NewNotFoundErrorFromContext creates a not found error using context
func NewNotFoundErrorFromContext(ctx context.Context, resource, id string) *CustomError {
	err := NewNotFoundError(resource, id)
	return enrichFromContext(ctx, err)
}

// NewInternalErrorFromContext creates an internal error using context
func NewInternalErrorFromContext(ctx context.Context, component string, wrapped error) *CustomError {
	err := NewInternalError(component, wrapped)
	return enrichFromContext(ctx, err)
}

// Error handler middleware support

// ErrorHandler is a function type for custom error handling
type ErrorHandler func(ctx context.Context, err *CustomError)

// WithErrorHandler adds a custom error handler to context
func WithErrorHandler(ctx context.Context, handler ErrorHandler) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ErrorHandlerContextKey, handler)
}

// HandleError processes an error using context-specific handlers
func HandleError(ctx context.Context, err *CustomError) {
	if ctx == nil || err == nil {
		return
	}

	if handler, ok := ctx.Value(ErrorHandlerContextKey).(ErrorHandler); ok {
		handler(ctx, err)
	}
}

// Context utilities for common patterns

// ContextualErrorBuilder provides context-aware error building
type ContextualErrorBuilder struct {
	ctx context.Context
	*ErrorBuilder
}

// NewContextualErrorBuilder creates a context-aware error builder
func NewContextualErrorBuilder(ctx context.Context, sentinel error) *ContextualErrorBuilder {
	return &ContextualErrorBuilder{
		ctx:          ctx,
		ErrorBuilder: NewErrorBuilder(sentinel),
	}
}

// Build creates the CustomError with context enrichment
func (b *ContextualErrorBuilder) Build() *CustomError {
	// First enrich the ErrorBuilder with context values before building
	if b.ctx != nil {
		if userID := GetUserIDFromContext(b.ctx); userID != "" {
			// Add user_id to the ErrorBuilder's metadata map
			if b.ErrorBuilder.metadata == nil {
				b.ErrorBuilder.metadata = make(map[string]string)
			}
			b.ErrorBuilder.metadata["user_id"] = userID
		}
		if requestID := GetRequestIDFromContext(b.ctx); requestID != "" {
			b.ErrorBuilder.requestID = requestID
		}
		if traceID := GetTraceIDFromContext(b.ctx); traceID != "" {
			if b.ErrorBuilder.metadata == nil {
				b.ErrorBuilder.metadata = make(map[string]string)
			}
			b.ErrorBuilder.metadata["trace_id"] = traceID
		}
	}

	err := b.ErrorBuilder.Build()

	// Handle error if handler is set
	HandleError(b.ctx, err)

	return err
}

// Context value extractors

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Try multiple common context keys
	keys := []string{"request_id", "requestID", "req_id", "x-request-id"}
	for _, key := range keys {
		if value, ok := ctx.Value(key).(string); ok && value != "" {
			return value
		}
	}

	return ""
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Try multiple common context keys
	keys := []string{"user_id", "userID", "user", "uid"}
	for _, key := range keys {
		if value, ok := ctx.Value(key).(string); ok && value != "" {
			return value
		}
	}

	return ""
}

// GetTraceIDFromContext extracts trace ID from context
func GetTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Try multiple common context keys (OpenTelemetry, Jaeger, etc.)
	keys := []string{"trace_id", "traceID", "trace-id", "x-trace-id", "span_id"}
	for _, key := range keys {
		if value, ok := ctx.Value(key).(string); ok && value != "" {
			return value
		}
	}

	return ""
}

// Enhanced enrichFromContext with better context value extraction
func enrichFromContext(ctx context.Context, err *CustomError) *CustomError {
	if ctx == nil || err == nil {
		return err
	}

	// Extract request ID
	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		err.WithRequestID(requestID)
	}

	// Extract user ID
	if userID := GetUserIDFromContext(ctx); userID != "" {
		err.WithMetadata("user_id", userID)
	}

	// Extract trace ID
	if traceID := GetTraceIDFromContext(ctx); traceID != "" {
		err.WithMetadata("trace_id", traceID)
	}

	// Handle error if handler is set
	HandleError(ctx, err)

	return err
}
