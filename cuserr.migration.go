package cuserr

import (
	"errors"
	"fmt"
	"strings"
)

// Migration utilities for converting from standard library and pkg/errors

// FromStdError converts a standard library error to a CustomError
// Attempts to categorize the error based on common patterns
func FromStdError(err error, message string) *CustomError {
	if err == nil {
		return nil
	}

	// Determine category based on error content
	errorText := strings.ToLower(err.Error())
	var sentinel error

	switch {
	case strings.Contains(errorText, "not found"):
		sentinel = ErrNotFound
	case strings.Contains(errorText, "permission denied") || strings.Contains(errorText, "forbidden"):
		sentinel = ErrForbidden
	case strings.Contains(errorText, "unauthorized") || strings.Contains(errorText, "authentication"):
		sentinel = ErrUnauthorized
	case strings.Contains(errorText, "timeout") || strings.Contains(errorText, "deadline"):
		sentinel = ErrTimeout
	case strings.Contains(errorText, "invalid") || strings.Contains(errorText, "bad"):
		sentinel = ErrInvalidInput
	case strings.Contains(errorText, "exists") || strings.Contains(errorText, "duplicate"):
		sentinel = ErrAlreadyExists
	case strings.Contains(errorText, "connection") || strings.Contains(errorText, "network"):
		sentinel = ErrExternal
	default:
		sentinel = ErrInternal
	}

	if message == "" {
		message = err.Error()
	}

	return NewCustomError(sentinel, err, message).
		WithMetadata("migrated_from", "stdlib").
		WithMetadata("original_error", err.Error())
}

// FromStdErrorWithCategory converts a standard error with explicit category
func FromStdErrorWithCategory(err error, category ErrorCategory, message string) *CustomError {
	if err == nil {
		return nil
	}

	if message == "" {
		message = err.Error()
	}

	code := fmt.Sprintf("%s_ERROR", strings.ToUpper(string(category)))

	return NewCustomErrorWithCategory(category, code, message).
		WithMetadata("migrated_from", "stdlib").
		WithMetadata("original_error", err.Error())
}

// WrapStdError wraps a standard error with additional context
func WrapStdError(err error, context, message string) *CustomError {
	if err == nil {
		return nil
	}

	wrappedMessage := message
	if context != "" {
		wrappedMessage = fmt.Sprintf("%s: %s", context, message)
	}

	customErr := FromStdError(err, wrappedMessage)
	if context != "" {
		customErr.WithMetadata("context", context)
	}

	return customErr
}

// Migration from pkg/errors patterns

// HasStackTrace checks if an error has stack trace information
// This is useful when migrating from pkg/errors which embeds stack traces
func HasStackTrace(err error) bool {
	// Check if it's already a CustomError with stack trace
	if customErr, ok := err.(*CustomError); ok {
		return len(customErr.GetStackTrace()) > 0
	}

	// Check for common stack trace interfaces
	type stackTracer interface {
		StackTrace() []uintptr
	}

	type causer interface {
		Cause() error
	}

	// Check if error has stack trace method (pkg/errors style)
	if _, hasStack := err.(stackTracer); hasStack {
		return true
	}

	// Walk the error chain looking for stack traces
	for err != nil {
		if _, hasStack := err.(stackTracer); hasStack {
			return true
		}

		// Try pkg/errors Cause method
		if causer, ok := err.(causer); ok {
			err = causer.Cause()
		} else {
			// Try standard library Unwrap
			err = errors.Unwrap(err)
		}
	}

	return false
}

// ExtractCause walks the error chain to find the root cause
// Compatible with both pkg/errors and Go 1.13+ error unwrapping
func ExtractCause(err error) error {
	type causer interface {
		Cause() error
	}

	for {
		// Try pkg/errors Cause method first
		if causer, ok := err.(causer); ok {
			next := causer.Cause()
			if next == nil {
				break
			}
			err = next
			continue
		}

		// Try standard library Unwrap
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			err = unwrapped
			continue
		}

		break
	}

	return err
}

// FromPkgError converts a pkg/errors style error to CustomError
func FromPkgError(err error, message string) *CustomError {
	if err == nil {
		return nil
	}

	// Extract the root cause
	rootCause := ExtractCause(err)

	// Create CustomError based on the original error
	customErr := FromStdError(rootCause, message)
	if customErr == nil {
		return nil
	}

	// Add migration metadata
	customErr.WithMetadata("migrated_from", "pkg/errors")

	// If the original error had stack trace info, note it
	if HasStackTrace(err) {
		customErr.WithMetadata("had_stack_trace", "true")
	}

	// Preserve the full error chain as metadata
	var errorChain []string
	current := err
	for current != nil {
		errorChain = append(errorChain, current.Error())

		// Try both unwrapping methods
		if causer, ok := current.(interface{ Cause() error }); ok {
			current = causer.Cause()
		} else {
			current = errors.Unwrap(current)
		}

		// Prevent infinite loops
		if len(errorChain) > 10 {
			break
		}
	}

	if len(errorChain) > 1 {
		customErr.WithMetadata("error_chain_length", fmt.Sprintf("%d", len(errorChain)))
	}

	return customErr
}

// Batch migration utilities

// MigrateErrorsInSlice converts a slice of standard errors to CustomErrors
func MigrateErrorsInSlice(errors []error) []*CustomError {
	if len(errors) == 0 {
		return nil
	}

	customErrors := make([]*CustomError, 0, len(errors))
	for _, err := range errors {
		if err != nil {
			customErr := FromStdError(err, "")
			customErrors = append(customErrors, customErr)
		}
	}

	return customErrors
}

// MigrateErrorsInMap converts a map of errors to CustomErrors
func MigrateErrorsInMap(errorMap map[string]error) map[string]*CustomError {
	if len(errorMap) == 0 {
		return nil
	}

	customErrorMap := make(map[string]*CustomError)
	for key, err := range errorMap {
		if err != nil {
			customErr := FromStdError(err, "")
			customErr.WithMetadata("original_key", key)
			customErrorMap[key] = customErr
		}
	}

	return customErrorMap
}

// HTTP error migration utilities

// FromHTTPStatus creates a CustomError from an HTTP status code
func FromHTTPStatus(statusCode int, message string) *CustomError {
	var sentinel error

	switch statusCode {
	case 400:
		sentinel = ErrInvalidInput
	case 401:
		sentinel = ErrUnauthorized
	case 403:
		sentinel = ErrForbidden
	case 404:
		sentinel = ErrNotFound
	case 408:
		sentinel = ErrTimeout
	case 409:
		sentinel = ErrAlreadyExists
	case 429:
		sentinel = ErrRateLimit
	case 500:
		sentinel = ErrInternal
	case 502, 503, 504:
		sentinel = ErrExternal
	default:
		if statusCode >= 400 && statusCode < 500 {
			sentinel = ErrInvalidInput
		} else if statusCode >= 500 {
			sentinel = ErrInternal
		} else {
			sentinel = ErrInternal
		}
	}

	if message == "" {
		message = fmt.Sprintf("HTTP %d error", statusCode)
	}

	return NewCustomError(sentinel, nil, message).
		WithMetadata("migrated_from", "http_status").
		WithMetadata("original_status_code", fmt.Sprintf("%d", statusCode))
}

// Framework-specific migration helpers

// FromGinError converts a Gin framework error to CustomError
func FromGinError(err error, message string) *CustomError {
	if err == nil {
		return nil
	}

	customErr := FromStdError(err, message)
	if customErr != nil {
		customErr.WithMetadata("migrated_from", "gin")
		customErr.WithMetadata("framework", "gin")
	}

	return customErr
}

// FromEchoError converts an Echo framework error to CustomError
func FromEchoError(err error, message string) *CustomError {
	if err == nil {
		return nil
	}

	customErr := FromStdError(err, message)
	if customErr != nil {
		customErr.WithMetadata("migrated_from", "echo")
		customErr.WithMetadata("framework", "echo")
	}

	return customErr
}

// FromFiberError converts a Fiber framework error to CustomError
func FromFiberError(err error, message string) *CustomError {
	if err == nil {
		return nil
	}

	customErr := FromStdError(err, message)
	if customErr != nil {
		customErr.WithMetadata("migrated_from", "fiber")
		customErr.WithMetadata("framework", "fiber")
	}

	return customErr
}

// Database migration helpers

// FromSQLError converts database errors to CustomErrors
func FromSQLError(err error, query string) *CustomError {
	if err == nil {
		return nil
	}

	errorText := strings.ToLower(err.Error())
	var sentinel error
	var message string

	switch {
	case strings.Contains(errorText, "duplicate") || strings.Contains(errorText, "unique"):
		sentinel = ErrAlreadyExists
		message = "resource already exists"
	case strings.Contains(errorText, "not found") || strings.Contains(errorText, "no rows"):
		sentinel = ErrNotFound
		message = "resource not found"
	case strings.Contains(errorText, "foreign key") || strings.Contains(errorText, "constraint"):
		sentinel = ErrInvalidInput
		message = "constraint violation"
	case strings.Contains(errorText, "connection") || strings.Contains(errorText, "timeout"):
		sentinel = ErrExternal
		message = "database connection error"
	case strings.Contains(errorText, "syntax") || strings.Contains(errorText, "invalid"):
		sentinel = ErrInvalidInput
		message = "invalid query"
	default:
		sentinel = ErrInternal
		message = "database error"
	}

	customErr := NewCustomError(sentinel, err, message).
		WithMetadata("migrated_from", "sql").
		WithMetadata("error_type", "database")

	if query != "" {
		// Only store the query if it's not too long and doesn't contain sensitive data
		if len(query) < 500 && !containsSensitiveSQL(query) {
			customErr.WithMetadata("query", query)
		} else {
			customErr.WithMetadata("query_length", fmt.Sprintf("%d", len(query)))
		}
	}

	return customErr
}

// containsSensitiveSQL checks if a SQL query contains potentially sensitive keywords
func containsSensitiveSQL(query string) bool {
	lowerQuery := strings.ToLower(query)
	sensitiveKeywords := []string{
		"password", "secret", "token", "key", "credential",
		"ssn", "social_security", "credit_card", "card_number",
	}

	for _, keyword := range sensitiveKeywords {
		if strings.Contains(lowerQuery, keyword) {
			return true
		}
	}

	return false
}

// Validation migration helpers

// FromValidationErrors converts multiple validation errors to ErrorCollection
func FromValidationErrors(validationErrors map[string][]string) *ErrorCollection {
	if len(validationErrors) == 0 {
		return nil
	}

	collection := NewValidationErrorCollection().
		WithContext("migrated_from", "validation_errors")

	for field, errors := range validationErrors {
		for _, errMsg := range errors {
			collection.AddValidation(field, errMsg)
		}
	}

	return collection
}

// FromFieldErrors converts field-specific errors to ErrorCollection
func FromFieldErrors(fieldErrors map[string]error) *ErrorCollection {
	if len(fieldErrors) == 0 {
		return nil
	}

	collection := NewValidationErrorCollection().
		WithContext("migrated_from", "field_errors")

	for field, err := range fieldErrors {
		if err != nil {
			customErr := FromStdError(err, err.Error())
			collection.AddFieldError(field, customErr)
		}
	}

	return collection
}

// Compatibility checking utilities

// IsCompatibleError checks if an error is compatible with CustomError patterns
func IsCompatibleError(err error) bool {
	if err == nil {
		return false
	}

	// Already a CustomError
	if _, ok := err.(*CustomError); ok {
		return true
	}

	// Has Unwrap method (Go 1.13+ compatible)
	if errors.Unwrap(err) != nil {
		return true
	}

	// Has Is method (Go 1.13+ compatible)
	if errors.Is(err, err) { // This will return true for compatible errors
		return true
	}

	// Has As method (Go 1.13+ compatible)
	var target *CustomError
	if errors.As(err, &target) {
		return true
	}

	return false
}

// Migration report utilities

// MigrationReport provides information about error migration
type MigrationReport struct {
	Total    int      `json:"total"`
	Migrated int      `json:"migrated"`
	Skipped  int      `json:"skipped"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}

// NewMigrationReport creates a new migration report
func NewMigrationReport() *MigrationReport {
	return &MigrationReport{
		Errors: make([]string, 0),
	}
}

// AddMigrated increments the migrated count
func (r *MigrationReport) AddMigrated() {
	r.Total++
	r.Migrated++
}

// AddSkipped increments the skipped count
func (r *MigrationReport) AddSkipped() {
	r.Total++
	r.Skipped++
}

// AddFailed increments the failed count and records the error
func (r *MigrationReport) AddFailed(err error) {
	r.Total++
	r.Failed++
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}
}

// Summary returns a human-readable summary of the migration
func (r *MigrationReport) Summary() string {
	return fmt.Sprintf("Migration completed: %d total, %d migrated, %d skipped, %d failed",
		r.Total, r.Migrated, r.Skipped, r.Failed)
}

// BatchMigrate performs batch migration of errors with reporting
func BatchMigrate(errors []error) (*ErrorCollection, *MigrationReport) {
	report := NewMigrationReport()
	collection := NewErrorCollection("batch migration results")

	for _, err := range errors {
		if err == nil {
			report.AddSkipped()
			continue
		}

		customErr := FromStdError(err, "")
		if customErr != nil {
			collection.Add(customErr)
			report.AddMigrated()
		} else {
			report.AddFailed(fmt.Errorf("failed to migrate error: %v", err))
		}
	}

	return collection, report
}
