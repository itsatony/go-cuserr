package cuserr

import "context"

// Log logs the error using the configured logger.
// If no logger is configured for this error, it does nothing (no-op).
// This allows errors to be created without forcing logging, giving users full control.
func (e *CustomError) Log(ctx context.Context) {
	if e.logger == nil {
		return // No-op if no logger configured
	}

	fields := e.toLogFields()
	e.logger.Error(ctx, e.Message, fields...)
}

// LogWithLevel logs the error at a specific level
func (e *CustomError) LogWithLevel(ctx context.Context, level string) {
	if e.logger == nil {
		return
	}

	fields := e.toLogFields()

	switch level {
	case LOG_LEVEL_DEBUG:
		e.logger.Debug(ctx, e.Message, fields...)
	case LOG_LEVEL_INFO:
		e.logger.Info(ctx, e.Message, fields...)
	case LOG_LEVEL_WARN:
		e.logger.Warn(ctx, e.Message, fields...)
	case LOG_LEVEL_ERROR:
		fallthrough
	default:
		e.logger.Error(ctx, e.Message, fields...)
	}
}

// toLogFields converts error to structured log fields
func (e *CustomError) toLogFields() []Field {
	e.mu.RLock()
	defer e.mu.RUnlock()

	fields := []Field{
		{Key: "error_category", Value: string(e.Category)},
		{Key: "error_code", Value: e.Code},
		{Key: JSON_FIELD_TIMESTAMP, Value: e.Timestamp},
	}

	if e.OperationID != "" {
		fields = append(fields, Field{Key: "operation_id", Value: e.OperationID})
	}

	// Add metadata
	for key, value := range e.metadata {
		fields = append(fields, Field{Key: key, Value: value})
	}

	// Add wrapped error if present
	if e.Wrapped != nil {
		fields = append(fields, Field{Key: "wrapped_error", Value: e.Wrapped.Error()})
	}

	// Add stack trace info if available
	if len(e.stackTrace) > 0 {
		fields = append(fields, Field{Key: "has_stack_trace", Value: true})
		fields = append(fields, Field{Key: "stack_depth", Value: len(e.stackTrace)})

		// Add top frame for quick reference
		if len(e.stackTrace) > 0 {
			fields = append(fields, Field{Key: "top_frame_function", Value: e.stackTrace[0].Function})
			fields = append(fields, Field{Key: "top_frame_file", Value: e.stackTrace[0].File})
			fields = append(fields, Field{Key: "top_frame_line", Value: e.stackTrace[0].Line})
		}
	}

	return fields
}
