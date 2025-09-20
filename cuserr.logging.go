package cuserr

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// LogLevel represents logging levels
type LogLevel int

const (
	// LogLevelDebug represents debug level logging for detailed information
	LogLevelDebug LogLevel = iota
	// LogLevelInfo represents info level logging for general information
	LogLevelInfo
	// LogLevelWarn represents warning level logging for concerning events
	LogLevelWarn
	// LogLevelError represents error level logging for error conditions
	LogLevelError
)

// String returns string representation of log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	default:
		return "unknown"
	}
}

// StructuredLogger interface for pluggable logging backends
type StructuredLogger interface {
	Log(ctx context.Context, level LogLevel, message string, fields map[string]interface{})
	LogError(ctx context.Context, err *CustomError)
	LogErrorCollection(ctx context.Context, collection *ErrorCollection)
}

// DefaultSlogLogger implements StructuredLogger using Go's slog package
type DefaultSlogLogger struct {
	logger *slog.Logger
}

// NewDefaultSlogLogger creates a new slog-based structured logger
func NewDefaultSlogLogger(logger *slog.Logger) *DefaultSlogLogger {
	if logger == nil {
		logger = slog.Default()
	}
	return &DefaultSlogLogger{logger: logger}
}

// Log logs a message with structured fields
func (l *DefaultSlogLogger) Log(ctx context.Context, level LogLevel, message string, fields map[string]interface{}) {
	var slogLevel slog.Level
	switch level {
	case LogLevelDebug:
		slogLevel = slog.LevelDebug
	case LogLevelInfo:
		slogLevel = slog.LevelInfo
	case LogLevelWarn:
		slogLevel = slog.LevelWarn
	case LogLevelError:
		slogLevel = slog.LevelError
	}

	args := make([]any, 0, len(fields)*2)
	for key, value := range fields {
		args = append(args, key, value)
	}

	l.logger.LogAttrs(ctx, slogLevel, message, argsToAttrs(args)...)
}

// LogError logs a CustomError with structured fields
func (l *DefaultSlogLogger) LogError(ctx context.Context, err *CustomError) {
	if err == nil {
		return
	}

	fields := err.ToLogFields()
	l.Log(ctx, LogLevelError, err.Message, fields)
}

// LogErrorCollection logs an ErrorCollection with structured fields
func (l *DefaultSlogLogger) LogErrorCollection(ctx context.Context, collection *ErrorCollection) {
	if collection == nil || collection.IsEmpty() {
		return
	}

	fields := collection.ToLogFields()
	l.Log(ctx, LogLevelError, collection.Error(), fields)
}

// Helper function to convert args to slog.Attr
func argsToAttrs(args []any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			attrs = append(attrs, slog.Any(key, args[i+1]))
		}
	}
	return attrs
}

// Structured logging methods for CustomError

// ToLogFields converts error to structured log fields
func (e *CustomError) ToLogFields() map[string]interface{} {
	fields := map[string]interface{}{
		"error_category": string(e.Category),
		"error_code":     e.Code,
		"error_message":  e.Message,
		"timestamp":      e.Timestamp.Format(time.RFC3339),
	}

	if e.RequestID != "" {
		fields["request_id"] = e.RequestID
	}

	// Add metadata
	metadata := e.GetAllMetadata()
	for key, value := range metadata {
		// Prefix metadata fields to avoid conflicts
		fields[fmt.Sprintf("meta_%s", key)] = value
	}

	// Add wrapped error if present
	if e.Wrapped != nil {
		fields["wrapped_error"] = e.Wrapped.Error()
	}

	// Add stack trace info if available
	stackTrace := e.GetStackTrace()
	if len(stackTrace) > 0 {
		fields["has_stack_trace"] = true
		fields["stack_depth"] = len(stackTrace)
		// Add top frame for quick reference
		if len(stackTrace) > 0 {
			fields["top_frame_function"] = stackTrace[0].Function
			fields["top_frame_file"] = stackTrace[0].File
			fields["top_frame_line"] = stackTrace[0].Line
		}
	}

	return fields
}

// LogWith logs the error using the provided logger
func (e *CustomError) LogWith(ctx context.Context, logger StructuredLogger) {
	if logger != nil {
		logger.LogError(ctx, e)
	}
}

// LogDebug logs the error at debug level
func (e *CustomError) LogDebug(ctx context.Context, logger StructuredLogger, message string) {
	if logger != nil {
		fields := e.ToLogFields()
		logger.Log(ctx, LogLevelDebug, message, fields)
	}
}

// LogInfo logs the error at info level
func (e *CustomError) LogInfo(ctx context.Context, logger StructuredLogger, message string) {
	if logger != nil {
		fields := e.ToLogFields()
		logger.Log(ctx, LogLevelInfo, message, fields)
	}
}

// LogWarn logs the error at warn level
func (e *CustomError) LogWarn(ctx context.Context, logger StructuredLogger, message string) {
	if logger != nil {
		fields := e.ToLogFields()
		logger.Log(ctx, LogLevelWarn, message, fields)
	}
}

// Structured logging methods for ErrorCollection

// ToLogFields converts error collection to structured log fields
func (ec *ErrorCollection) ToLogFields() map[string]interface{} {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	fields := map[string]interface{}{
		"error_category":         string(ErrorCategoryValidation),
		"error_code":             "MULTIPLE_ERRORS",
		"error_message":          ec.Error(),
		"error_summary":          ec.Summary,
		"total_error_count":      ec.Count(),
		"validation_error_count": len(ec.ValidationErrors),
		"custom_error_count":     len(ec.Errors),
		"timestamp":              time.Now().UTC().Format(time.RFC3339),
	}

	if ec.RequestID != "" {
		fields["request_id"] = ec.RequestID
	}

	// Add context metadata
	for key, value := range ec.Context {
		fields[fmt.Sprintf("context_%s", key)] = value
	}

	// Add validation fields if any
	if len(ec.ValidationErrors) > 0 {
		fields["validation_fields"] = ec.GetFields()

		// Add detailed validation errors
		validationDetails := make([]map[string]string, len(ec.ValidationErrors))
		for i, vErr := range ec.ValidationErrors {
			validationDetails[i] = map[string]string{
				"field":   vErr.Field,
				"message": vErr.Message,
				"code":    vErr.Code,
				"value":   vErr.Value,
			}
		}
		fields["validation_errors"] = validationDetails
	}

	// Add error categories from custom errors
	if len(ec.Errors) > 0 {
		categories := make(map[string]int)
		for _, err := range ec.Errors {
			categoryStr := string(err.Category)
			categories[categoryStr]++
		}
		fields["error_categories"] = categories
	}

	return fields
}

// LogWith logs the error collection using the provided logger
func (ec *ErrorCollection) LogWith(ctx context.Context, logger StructuredLogger) {
	if logger != nil {
		logger.LogErrorCollection(ctx, ec)
	}
}

// Global structured logger instance
var globalStructuredLogger StructuredLogger

// SetStructuredLogger sets the global structured logger
func SetStructuredLogger(logger StructuredLogger) {
	globalStructuredLogger = logger
}

// GetStructuredLogger returns the global structured logger
func GetStructuredLogger() StructuredLogger {
	if globalStructuredLogger == nil {
		// Initialize with default slog logger if not set
		globalStructuredLogger = NewDefaultSlogLogger(nil)
	}
	return globalStructuredLogger
}

// Convenience functions using global logger

// LogError logs an error using the global structured logger
func LogError(ctx context.Context, err *CustomError) {
	logger := GetStructuredLogger()
	if logger != nil && err != nil {
		logger.LogError(ctx, err)
	}
}

// LogErrorCollection logs an error collection using the global structured logger
func LogErrorCollection(ctx context.Context, collection *ErrorCollection) {
	logger := GetStructuredLogger()
	if logger != nil && collection != nil {
		logger.LogErrorCollection(ctx, collection)
	}
}

// LogErrorWithMessage logs an error with a custom message
func LogErrorWithMessage(ctx context.Context, err *CustomError, level LogLevel, message string) {
	logger := GetStructuredLogger()
	if logger != nil && err != nil {
		fields := err.ToLogFields()
		logger.Log(ctx, level, message, fields)
	}
}

// Integration with popular logging frameworks

// ZapLogger wraps a zap logger to implement StructuredLogger
// Note: This is a stub - in real usage you'd import go.uber.org/zap
type ZapLogger struct {
	// In a real implementation, this would be *zap.Logger
	logger interface{}
}

// NewZapLogger creates a zap-based structured logger
// Note: This is a stub implementation
func NewZapLogger(logger interface{}) *ZapLogger {
	return &ZapLogger{logger: logger}
}

// Log implements StructuredLogger for zap
func (l *ZapLogger) Log(ctx context.Context, level LogLevel, message string, fields map[string]interface{}) {
	// In a real implementation, this would use zap's structured logging
	// For now, we'll use a simple print to demonstrate the pattern
	fmt.Printf("[%s] %s %+v\n", level.String(), message, fields)
}

// LogError implements StructuredLogger for zap
func (l *ZapLogger) LogError(ctx context.Context, err *CustomError) {
	if err != nil {
		fields := err.ToLogFields()
		l.Log(ctx, LogLevelError, err.Message, fields)
	}
}

// LogErrorCollection implements StructuredLogger for zap
func (l *ZapLogger) LogErrorCollection(ctx context.Context, collection *ErrorCollection) {
	if collection != nil && !collection.IsEmpty() {
		fields := collection.ToLogFields()
		l.Log(ctx, LogLevelError, collection.Error(), fields)
	}
}

// LogrusLogger wraps a logrus logger to implement StructuredLogger
// Note: This is a stub - in real usage you'd import github.com/sirupsen/logrus
type LogrusLogger struct {
	// In a real implementation, this would be *logrus.Logger
	logger interface{}
}

// NewLogrusLogger creates a logrus-based structured logger
// Note: This is a stub implementation
func NewLogrusLogger(logger interface{}) *LogrusLogger {
	return &LogrusLogger{logger: logger}
}

// Log implements StructuredLogger for logrus
func (l *LogrusLogger) Log(ctx context.Context, level LogLevel, message string, fields map[string]interface{}) {
	// In a real implementation, this would use logrus's structured logging
	fmt.Printf("[%s] %s %+v\n", level.String(), message, fields)
}

// LogError implements StructuredLogger for logrus
func (l *LogrusLogger) LogError(ctx context.Context, err *CustomError) {
	if err != nil {
		fields := err.ToLogFields()
		l.Log(ctx, LogLevelError, err.Message, fields)
	}
}

// LogErrorCollection implements StructuredLogger for logrus
func (l *LogrusLogger) LogErrorCollection(ctx context.Context, collection *ErrorCollection) {
	if collection != nil && !collection.IsEmpty() {
		fields := collection.ToLogFields()
		l.Log(ctx, LogLevelError, collection.Error(), fields)
	}
}

// Error handler that automatically logs errors
type LoggingErrorHandler struct {
	logger StructuredLogger
	level  LogLevel
}

// NewLoggingErrorHandler creates an error handler that logs errors
func NewLoggingErrorHandler(logger StructuredLogger, level LogLevel) *LoggingErrorHandler {
	return &LoggingErrorHandler{
		logger: logger,
		level:  level,
	}
}

// Handle logs the error
func (h *LoggingErrorHandler) Handle(ctx context.Context, err *CustomError) {
	if h.logger != nil && err != nil {
		fields := err.ToLogFields()
		h.logger.Log(ctx, h.level, err.Message, fields)
	}
}

// Context helpers for automatic error logging

// WithAutoLogging adds automatic error logging to context
func WithAutoLogging(ctx context.Context, logger StructuredLogger, level LogLevel) context.Context {
	handler := NewLoggingErrorHandler(logger, level)
	return WithErrorHandler(ctx, func(ctx context.Context, err *CustomError) {
		handler.Handle(ctx, err)
	})
}

// WithAutoErrorLogging adds automatic error logging at ERROR level to context
func WithAutoErrorLogging(ctx context.Context) context.Context {
	return WithAutoLogging(ctx, GetStructuredLogger(), LogLevelError)
}
