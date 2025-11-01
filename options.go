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

// WithContext enriches the error with context values (operation_id, user_id, trace_id).
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

		// Extract operation ID (try request_id for compatibility with existing context keys)
		if operationID := GetRequestIDFromContext(ctx); operationID != "" {
			e.OperationID = operationID
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

// WithPublicMessage sets a user-friendly message safe for end users.
// Works across all protocols: HTTP responses, CLI output, gRPC details, mobile apps.
//
// Example:
//
//	err := cuserr.Validation("email", "invalid format",
//	    cuserr.WithPublicMessage("Please enter a valid email address"),
//	)
func WithPublicMessage(message string) Option {
	return func(e *CustomError) {
		e.PublicMessage = message
	}
}

// WithHelpURL sets a URL pointing to documentation or troubleshooting guide.
//
// Example:
//
//	err := cuserr.Validation("email", "invalid",
//	    cuserr.WithHelpURL("https://docs.example.com/validation#email"),
//	)
func WithHelpURL(url string) Option {
	return func(e *CustomError) {
		e.HelpURL = url
	}
}

// WithSuggestions provides actionable steps to resolve the error.
//
// Example:
//
//	err := cuserr.Validation("email", "invalid",
//	    cuserr.WithSuggestions(
//	        "Check for typos in the email address",
//	        "Ensure the @ symbol is present",
//	        "Verify the domain name is correct",
//	    ),
//	)
func WithSuggestions(suggestions ...string) Option {
	return func(e *CustomError) {
		e.Suggestions = suggestions
	}
}

// WithMessageKey sets a translation key for i18n support.
// Use with go-cuserr-i18n extension package for multi-language error messages.
//
// Example:
//
//	err := cuserr.Validation("email", "invalid",
//	    cuserr.WithMessageKey("errors.validation.email_invalid"),
//	)
func WithMessageKey(key string) Option {
	return func(e *CustomError) {
		e.MessageKey = key
	}
}

// WithOperationID sets a unique operation identifier.
// Works for: HTTP requests, gRPC calls, CLI commands, batch jobs, cron tasks.
//
// Example:
//
//	err := cuserr.Internal("database", dbErr,
//	    cuserr.WithOperationID(operationID),
//	)
func WithOperationID(operationID string) Option {
	return func(e *CustomError) {
		e.OperationID = operationID
	}
}

// WithTraceID sets a distributed tracing trace ID.
// Compatible with OpenTelemetry, Jaeger, Zipkin, and other tracing systems.
//
// Example:
//
//	err := cuserr.External("payment-api", "charge", apiErr,
//	    cuserr.WithTraceID(span.SpanContext().TraceID().String()),
//	)
func WithTraceID(traceID string) Option {
	return func(e *CustomError) {
		e.TraceID = traceID
	}
}

// WithSpanID sets a distributed tracing span ID.
// Compatible with OpenTelemetry, Jaeger, Zipkin, and other tracing systems.
//
// Example:
//
//	err := cuserr.External("payment-api", "charge", apiErr,
//	    cuserr.WithSpanID(span.SpanContext().SpanID().String()),
//	)
func WithSpanID(spanID string) Option {
	return func(e *CustomError) {
		e.SpanID = spanID
	}
}
