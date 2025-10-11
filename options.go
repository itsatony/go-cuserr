package cuserr

import "context"

// Option is a functional option for configuring errors.
// This pattern allows for clean, composable error creation with optional features.
type Option func(*CustomError)

// WithLogger sets a logger for this error.
// The error will use this logger when Log() is called.
//
// Example:
//
//	logger := adapters.NewSlogAdapter(slog.Default())
//	err := cuserr.NotFound("user", id, cuserr.WithLogger(logger))
//	err.Log(ctx) // Logs using the configured logger
func WithLogger(logger Logger) Option {
	return func(e *CustomError) {
		e.logger = logger
	}
}

// WithContext enriches the error with context values (request_id, user_id, trace_id).
// This automatically extracts common identifiers from the context.
//
// Example:
//
//	err := cuserr.Internal("database", dbErr, cuserr.WithContext(ctx))
func WithContext(ctx context.Context) Option {
	return func(e *CustomError) {
		if ctx == nil {
			return
		}

		// Extract request ID
		if requestID := GetRequestIDFromContext(ctx); requestID != "" {
			e.RequestID = requestID
		}

		// Extract user ID
		if userID := GetUserIDFromContext(ctx); userID != "" {
			e.WithMetadata("user_id", userID)
		}

		// Extract trace ID
		if traceID := GetTraceIDFromContext(ctx); traceID != "" {
			e.WithMetadata("trace_id", traceID)
		}
	}
}

// WithMetadata adds a single metadata key-value pair.
//
// Example:
//
//	err := cuserr.Internal("database", dbErr,
//	    cuserr.WithMetadata("query", sql),
//	    cuserr.WithMetadata("duration_ms", "150"),
//	)
func WithMetadata(key, value string) Option {
	return func(e *CustomError) {
		e.WithMetadata(key, value)
	}
}

// WithRequestID sets the request ID for this error.
//
// Example:
//
//	err := cuserr.NotFound("user", id, cuserr.WithRequestID(requestID))
func WithRequestID(requestID string) Option {
	return func(e *CustomError) {
		e.RequestID = requestID
	}
}

// WithoutStack disables stack trace capture for this error.
// Useful for performance-sensitive code or when stack traces aren't needed.
//
// Example:
//
//	err := cuserr.Validation("email", "invalid", cuserr.WithoutStack())
func WithoutStack() Option {
	return func(e *CustomError) {
		e.stackTrace = nil
		e.stackTraceCleared = true
	}
}

// WithStack explicitly captures a stack trace (even if globally disabled).
// The skip parameter determines how many stack frames to skip (typically 1-2).
//
// Example:
//
//	err := cuserr.Internal("worker", err, cuserr.WithStack(2))
func WithStack(skip int) Option {
	return func(e *CustomError) {
		e.stackTrace = captureStackTrace(skip + 1) // +1 to skip WithStack itself
		e.stackTraceCleared = false
	}
}
