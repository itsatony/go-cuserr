package cuserr

import (
	"context"
	"testing"
)

// TestOptionsPattern verifies the functional options pattern for v0.3.0
func TestOptionsPattern(t *testing.T) {
	t.Run("WithLogger option", func(t *testing.T) {
		logger := &NoOpLogger{}
		err := NewNotFoundError("user", "123",
			WithLogger(logger),
		)

		if err.logger == nil {
			t.Error("WithLogger option did not set logger")
		}
	})

	t.Run("WithContext option extracts request_id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "req-test-123")
		err := NewInternalError("database", nil,
			WithContext(ctx),
		)

		if err.RequestID != "req-test-123" {
			t.Errorf("WithContext did not extract request_id, got: %s", err.RequestID)
		}
	})

	t.Run("WithContext option extracts user_id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", "user-456")
		err := NewInternalError("database", nil,
			WithContext(ctx),
		)

		userID, ok := err.GetMetadata("user_id")
		if !ok || userID != "user-456" {
			t.Errorf("WithContext did not extract user_id, got: %s, exists: %v", userID, ok)
		}
	})

	t.Run("WithContext option extracts trace_id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "trace_id", "trace-789")
		err := NewInternalError("database", nil,
			WithContext(ctx),
		)

		traceID, ok := err.GetMetadata("trace_id")
		if !ok || traceID != "trace-789" {
			t.Errorf("WithContext did not extract trace_id, got: %s, exists: %v", traceID, ok)
		}
	})

	t.Run("WithMetadata option adds metadata", func(t *testing.T) {
		err := NewInternalError("database", nil,
			WithMetadata("query", "SELECT * FROM users"),
			WithMetadata("duration_ms", "250"),
		)

		query, ok := err.GetMetadata("query")
		if !ok || query != "SELECT * FROM users" {
			t.Errorf("WithMetadata did not add query metadata correctly")
		}

		duration, ok := err.GetMetadata("duration_ms")
		if !ok || duration != "250" {
			t.Errorf("WithMetadata did not add duration_ms metadata correctly")
		}
	})

	t.Run("WithRequestID option sets request ID", func(t *testing.T) {
		err := NewInternalError("database", nil,
			WithRequestID("custom-req-999"),
		)

		if err.RequestID != "custom-req-999" {
			t.Errorf("WithRequestID did not set request ID, got: %s", err.RequestID)
		}
	})

	t.Run("WithoutStack option clears stack trace", func(t *testing.T) {
		err := NewInternalError("database", nil,
			WithoutStack(),
		)

		if err.stackTrace != nil {
			t.Error("WithoutStack did not clear stack trace")
		}
		if !err.stackTraceCleared {
			t.Error("WithoutStack did not set stackTraceCleared flag")
		}
	})

	t.Run("WithStack option captures stack trace", func(t *testing.T) {
		err := NewInternalError("database", nil,
			WithoutStack(), // First clear it
		)
		if err.stackTrace != nil {
			t.Error("Setup failed: stack trace should be nil")
		}

		// Now apply WithStack
		opt := WithStack(1)
		opt(err)

		if err.stackTrace == nil {
			t.Error("WithStack did not capture stack trace")
		}
		if err.stackTraceCleared {
			t.Error("WithStack did not reset stackTraceCleared flag")
		}
	})

	t.Run("Multiple options composition", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "req-multi-test")
		ctx = context.WithValue(ctx, "user_id", "user-multi")
		logger := &NoOpLogger{}

		err := NewExternalError("payment-api", "charge", nil,
			WithLogger(logger),
			WithContext(ctx),
			WithMetadata("amount", "99.99"),
			WithMetadata("currency", "USD"),
		)

		// Verify logger
		if err.logger == nil {
			t.Error("Multiple options: logger not set")
		}

		// Verify request ID from context
		if err.RequestID != "req-multi-test" {
			t.Errorf("Multiple options: request ID not extracted, got: %s", err.RequestID)
		}

		// Verify user ID from context
		userID, ok := err.GetMetadata("user_id")
		if !ok || userID != "user-multi" {
			t.Error("Multiple options: user_id not extracted")
		}

		// Verify custom metadata
		amount, ok := err.GetMetadata("amount")
		if !ok || amount != "99.99" {
			t.Error("Multiple options: amount metadata not set")
		}

		currency, ok := err.GetMetadata("currency")
		if !ok || currency != "USD" {
			t.Error("Multiple options: currency metadata not set")
		}
	})

	t.Run("WithContext handles nil context gracefully", func(t *testing.T) {
		// Should not panic
		err := NewInternalError("database", nil,
			WithContext(nil),
		)

		if err == nil {
			t.Error("Error creation failed with nil context")
		}
	})
}

// TestLoggerMethods verifies the logging functionality for v0.3.0
func TestLoggerMethods(t *testing.T) {
	t.Run("Log with no logger is no-op", func(t *testing.T) {
		err := NewNotFoundError("user", "123")
		// Should not panic
		err.Log(context.Background())
	})

	t.Run("Log with NoOpLogger is no-op", func(t *testing.T) {
		logger := &NoOpLogger{}
		err := NewNotFoundError("user", "123",
			WithLogger(logger),
		)
		// Should not panic
		err.Log(context.Background())
	})

	t.Run("LogWithLevel with no logger is no-op", func(t *testing.T) {
		err := NewNotFoundError("user", "123")
		// Should not panic
		err.LogWithLevel(context.Background(), "error")
		err.LogWithLevel(context.Background(), "warn")
		err.LogWithLevel(context.Background(), "info")
		err.LogWithLevel(context.Background(), "debug")
	})

	t.Run("toLogFields extracts error fields", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "req-log-test")
		err := NewInternalError("database", nil,
			WithContext(ctx),
			WithMetadata("query", "SELECT 1"),
		)

		fields := err.toLogFields()

		// Verify fields contain expected data
		if len(fields) == 0 {
			t.Error("toLogFields returned empty fields")
		}

		// Check that common fields are present
		// Note: message is NOT in fields - it's passed separately to logger
		foundCategory := false
		foundCode := false
		foundTimestamp := false
		foundRequestID := false
		foundQuery := false

		for _, field := range fields {
			switch field.Key {
			case "error_category":
				foundCategory = true
			case "error_code":
				foundCode = true
			case "timestamp":
				foundTimestamp = true
			case "request_id":
				foundRequestID = true
				if field.Value != "req-log-test" {
					t.Errorf("request_id value incorrect: got %v, want req-log-test", field.Value)
				}
			case "query":
				foundQuery = true
				if field.Value != "SELECT 1" {
					t.Errorf("query metadata incorrect: got %v, want SELECT 1", field.Value)
				}
			}
		}

		if !foundCategory {
			t.Error("toLogFields missing error_category field")
		}
		if !foundCode {
			t.Error("toLogFields missing error_code field")
		}
		if !foundTimestamp {
			t.Error("toLogFields missing timestamp field")
		}
		if !foundRequestID {
			t.Error("toLogFields missing request_id field from context")
		}
		if !foundQuery {
			t.Error("toLogFields missing query metadata")
		}
	})
}

// TestNoOpLogger verifies the zero-cost abstraction
func TestNoOpLogger(t *testing.T) {
	logger := &NoOpLogger{}
	ctx := context.Background()

	// These should all be no-ops and not panic
	t.Run("Error is no-op", func(t *testing.T) {
		logger.Error(ctx, "test message", Field{Key: "key", Value: "value"})
	})

	t.Run("Warn is no-op", func(t *testing.T) {
		logger.Warn(ctx, "test message", Field{Key: "key", Value: "value"})
	})

	t.Run("Info is no-op", func(t *testing.T) {
		logger.Info(ctx, "test message", Field{Key: "key", Value: "value"})
	})

	t.Run("Debug is no-op", func(t *testing.T) {
		logger.Debug(ctx, "test message", Field{Key: "key", Value: "value"})
	})
}
