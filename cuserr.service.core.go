package cuserr

import (
	"errors"
	"fmt"
	"time"
)

// NewCustomError creates a new CustomError with the given sentinel error and context
// This is the primary constructor for creating rich errors with automatic categorization
// Uses lazy loading for metadata but captures stack traces immediately for accuracy
func NewCustomError(sentinel error, wrapped error, message string) *CustomError {
	category := mapSentinelToCategory(sentinel)
	code := generateErrorCode(sentinel)

	err := &CustomError{
		Category:  category,
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC(),
		Wrapped:   wrapped,
		Sentinel:  sentinel,
		// metadata is nil - lazy loaded when needed
	}

	// Capture stack trace immediately if enabled (for accuracy)
	config := GetConfig()
	if config.EnableStackTrace {
		err.stackTrace = captureStackTrace(STACK_SKIP_FRAMES)
	}

	return err
}

// NewCustomErrorWithCategory creates an error with explicit category
// Use when you need direct control over error categorization
// Uses lazy loading for metadata but captures stack traces immediately for accuracy
func NewCustomErrorWithCategory(category ErrorCategory, code, message string) *CustomError {
	err := &CustomError{
		Category:  category,
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC(),
		// metadata is nil - lazy loaded when needed
	}

	// Capture stack trace immediately if enabled (for accuracy)
	config := GetConfig()
	if config.EnableStackTrace {
		err.stackTrace = captureStackTrace(STACK_SKIP_FRAMES)
	}

	return err
}

// WithMetadata adds metadata to the error in a thread-safe manner
// Implements lazy loading - map is created only when first metadata is added
func (e *CustomError) WithMetadata(key, value string) *CustomError {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Lazy initialize metadata map
	if e.metadata == nil {
		e.metadata = make(map[string]string)
	}
	e.metadata[key] = value
	return e
}

// GetMetadata retrieves metadata value by key in a thread-safe manner
func (e *CustomError) GetMetadata(key string) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.metadata == nil {
		return "", false
	}
	value, exists := e.metadata[key]
	return value, exists
}

// GetAllMetadata returns a copy of all metadata in a thread-safe manner
func (e *CustomError) GetAllMetadata() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.metadata == nil {
		return make(map[string]string)
	}

	// Return a copy to prevent external modification
	copy := make(map[string]string)
	for k, v := range e.metadata {
		copy[k] = v
	}
	return copy
}

// WithRequestID adds request ID for tracing
func (e *CustomError) WithRequestID(requestID string) *CustomError {
	e.RequestID = requestID
	return e
}

// Error implements the error interface
func (e *CustomError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Wrapped)
	}
	return e.Message
}

// Unwrap implements the errors.Unwrap interface for error chain unwrapping
func (e *CustomError) Unwrap() error {
	return e.Wrapped
}

// Is implements the errors.Is interface for sentinel error checking
func (e *CustomError) Is(target error) bool {
	return errors.Is(e.Sentinel, target) || errors.Is(e.Wrapped, target)
}

// WrapWithCustomError wraps a standard error into a CustomError
func WrapWithCustomError(err error, category ErrorCategory, code, message string) *CustomError {
	if err == nil {
		return nil
	}

	customErr := NewCustomErrorWithCategory(category, code, message)
	customErr.Wrapped = err
	return customErr
}

// ErrorWithContext wraps an error with additional context
// This is a convenience function for simple error wrapping
func ErrorWithContext(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// Helper Functions for Error Inspection

// IsErrorCategory checks if an error belongs to a specific category
func IsErrorCategory(err error, category ErrorCategory) bool {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.Category == category
	}
	return false
}

// IsErrorCode checks if an error has a specific code
func IsErrorCode(err error, code string) bool {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.Code == code
	}
	return false
}

// GetErrorCategory extracts the category from an error
func GetErrorCategory(err error) ErrorCategory {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.Category
	}
	return ErrorCategoryInternal
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) string {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.Code
	}
	return ERROR_CODE_INTERNAL_ERROR
}

// GetErrorMetadata extracts metadata from an error
func GetErrorMetadata(err error, key string) (string, bool) {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.GetMetadata(key)
	}
	return "", false
}
