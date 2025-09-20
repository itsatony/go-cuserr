// Package cuserr provides convenience constructors for common error patterns.
// This file contains simplified constructors that reduce boilerplate code.
package cuserr

import (
	"context"
	"fmt"
)

// Convenience constructors for common error patterns
// These reduce boilerplate and improve developer experience

// NewValidationError creates a validation error with field-specific context
func NewValidationError(field, message string) *CustomError {
	err := NewCustomError(ErrInvalidInput, nil, message).
		WithMetadata("field", field).
		WithMetadata("error_type", "validation")
	return err
}

// NewValidationErrorf creates a validation error with formatted message
func NewValidationErrorf(field, format string, args ...interface{}) *CustomError {
	message := fmt.Sprintf(format, args...)
	return NewValidationError(field, message)
}

// NewNotFoundError creates a not found error for a specific resource
func NewNotFoundError(resource, id string) *CustomError {
	message := fmt.Sprintf("%s not found", resource)
	if id != "" {
		message = fmt.Sprintf("%s with id '%s' not found", resource, id)
	}

	err := NewCustomError(ErrNotFound, nil, message).
		WithMetadata("resource", resource).
		WithMetadata("error_type", "not_found")

	if id != "" {
		err.WithMetadata("resource_id", id)
	}

	return err
}

// NewUnauthorizedError creates an unauthorized error with optional reason
func NewUnauthorizedError(reason string) *CustomError {
	message := "authentication required"
	if reason != "" {
		message = fmt.Sprintf("authentication required: %s", reason)
	}

	err := NewCustomError(ErrUnauthorized, nil, message).
		WithMetadata("error_type", "unauthorized")

	if reason != "" {
		err.WithMetadata("reason", reason)
	}

	return err
}

// NewForbiddenError creates a forbidden error with optional details
func NewForbiddenError(action, resource string) *CustomError {
	message := "access denied"
	if action != "" && resource != "" {
		message = fmt.Sprintf("access denied: cannot %s %s", action, resource)
	}

	err := NewCustomError(ErrForbidden, nil, message).
		WithMetadata("error_type", "forbidden")

	if action != "" {
		err.WithMetadata("action", action)
	}
	if resource != "" {
		err.WithMetadata("resource", resource)
	}

	return err
}

// NewConflictError creates a conflict error for resource conflicts
func NewConflictError(resource, field, value string) *CustomError {
	message := fmt.Sprintf("%s already exists", resource)
	if field != "" && value != "" {
		message = fmt.Sprintf("%s with %s '%s' already exists", resource, field, value)
	}

	err := NewCustomError(ErrAlreadyExists, nil, message).
		WithMetadata("resource", resource).
		WithMetadata("error_type", "conflict")

	if field != "" {
		err.WithMetadata("conflict_field", field)
	}
	if value != "" {
		err.WithMetadata("conflict_value", value)
	}

	return err
}

// NewInternalError creates an internal error with optional component context
func NewInternalError(component string, wrapped error) *CustomError {
	message := "internal server error"
	if component != "" {
		message = fmt.Sprintf("internal error in %s", component)
	}

	err := NewCustomError(ErrInternal, wrapped, message).
		WithMetadata("error_type", "internal")

	if component != "" {
		err.WithMetadata("component", component)
	}

	return err
}

// NewExternalError creates an external service error
func NewExternalError(service, operation string, wrapped error) *CustomError {
	message := "external service error"
	if service != "" {
		message = fmt.Sprintf("external service '%s' error", service)
	}

	err := NewCustomError(ErrExternal, wrapped, message).
		WithMetadata("error_type", "external")

	if service != "" {
		err.WithMetadata("service", service)
	}
	if operation != "" {
		err.WithMetadata("operation", operation)
	}

	return err
}

// NewTimeoutError creates a timeout error with optional operation context
func NewTimeoutError(operation string, wrapped error) *CustomError {
	message := "operation timed out"
	if operation != "" {
		message = fmt.Sprintf("%s operation timed out", operation)
	}

	err := NewCustomError(ErrTimeout, wrapped, message).
		WithMetadata("error_type", "timeout")

	if operation != "" {
		err.WithMetadata("operation", operation)
	}

	return err
}

// NewRateLimitError creates a rate limit error with limit information
func NewRateLimitError(limit, window string) *CustomError {
	message := "rate limit exceeded"
	if limit != "" && window != "" {
		message = fmt.Sprintf("rate limit exceeded: %s per %s", limit, window)
	}

	err := NewCustomError(ErrRateLimit, nil, message).
		WithMetadata("error_type", "rate_limit")

	if limit != "" {
		err.WithMetadata("limit", limit)
	}
	if window != "" {
		err.WithMetadata("window", window)
	}

	return err
}

// Context-aware convenience constructors

// NewValidationErrorWithContext creates a validation error with request context
func NewValidationErrorWithContext(ctx context.Context, field, message string) *CustomError {
	err := NewValidationError(field, message)
	return enrichFromContext(ctx, err)
}

// NewNotFoundErrorWithContext creates a not found error with request context
func NewNotFoundErrorWithContext(ctx context.Context, resource, id string) *CustomError {
	err := NewNotFoundError(resource, id)
	return enrichFromContext(ctx, err)
}

// NewInternalErrorWithContext creates an internal error with request context
func NewInternalErrorWithContext(ctx context.Context, component string, wrapped error) *CustomError {
	err := NewInternalError(component, wrapped)
	return enrichFromContext(ctx, err)
}

// NewExternalErrorFromContext creates an external service error with request context
func NewExternalErrorFromContext(ctx context.Context, service, operation string, wrapped error) *CustomError {
	err := NewExternalError(service, operation, wrapped)
	return enrichFromContext(ctx, err)
}

// NewTimeoutErrorFromContext creates a timeout error with request context
func NewTimeoutErrorFromContext(ctx context.Context, operation string, wrapped error) *CustomError {
	err := NewTimeoutError(operation, wrapped)
	return enrichFromContext(ctx, err)
}

// NewRateLimitErrorFromContext creates a rate limit error with request context
func NewRateLimitErrorFromContext(ctx context.Context, limit, window string) *CustomError {
	err := NewRateLimitError(limit, window)
	return enrichFromContext(ctx, err)
}

// NewConflictErrorFromContext creates a conflict error with request context
func NewConflictErrorFromContext(ctx context.Context, resource, field, value string) *CustomError {
	err := NewConflictError(resource, field, value)
	return enrichFromContext(ctx, err)
}

// NewUnauthorizedErrorFromContext creates an unauthorized error with request context
func NewUnauthorizedErrorFromContext(ctx context.Context, reason string) *CustomError {
	err := NewUnauthorizedError(reason)
	return enrichFromContext(ctx, err)
}

// NewForbiddenErrorFromContext creates a forbidden error with request context
func NewForbiddenErrorFromContext(ctx context.Context, action, resource string) *CustomError {
	err := NewForbiddenError(action, resource)
	return enrichFromContext(ctx, err)
}

// Error builder pattern for complex errors

// ErrorBuilder provides a fluent interface for building complex errors
type ErrorBuilder struct {
	sentinel  error
	wrapped   error
	message   string
	metadata  map[string]string
	requestID string
}

// NewErrorBuilder creates a new error builder
func NewErrorBuilder(sentinel error) *ErrorBuilder {
	return &ErrorBuilder{
		sentinel: sentinel,
		metadata: make(map[string]string),
	}
}

// WithMessage sets the error message
func (b *ErrorBuilder) WithMessage(message string) *ErrorBuilder {
	b.message = message
	return b
}

// WithMessagef sets the error message with formatting
func (b *ErrorBuilder) WithMessagef(format string, args ...interface{}) *ErrorBuilder {
	b.message = fmt.Sprintf(format, args...)
	return b
}

// WithWrapped sets the wrapped error
func (b *ErrorBuilder) WithWrapped(err error) *ErrorBuilder {
	b.wrapped = err
	return b
}

// WithMetadata adds metadata
func (b *ErrorBuilder) WithMetadata(key, value string) *ErrorBuilder {
	b.metadata[key] = value
	return b
}

// WithRequestID sets the request ID
func (b *ErrorBuilder) WithRequestID(requestID string) *ErrorBuilder {
	b.requestID = requestID
	return b
}

// WithContext extracts common fields from context
func (b *ErrorBuilder) WithContext(ctx context.Context) *ErrorBuilder {
	if ctx == nil {
		return b
	}

	if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
		b.requestID = requestID
	}

	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		b.metadata["user_id"] = userID
	}

	if traceID, ok := ctx.Value("trace_id").(string); ok && traceID != "" {
		b.metadata["trace_id"] = traceID
	}

	return b
}

// Build creates the CustomError
func (b *ErrorBuilder) Build() *CustomError {
	err := NewCustomError(b.sentinel, b.wrapped, b.message)

	if b.requestID != "" {
		err.WithRequestID(b.requestID)
	}

	for key, value := range b.metadata {
		err.WithMetadata(key, value)
	}

	return err
}
