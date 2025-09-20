package cuserr

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Value   string `json:"value,omitempty"`
}

// ErrorCollection represents a collection of related errors
type ErrorCollection struct {
	// Errors contains the individual errors
	Errors []*CustomError `json:"errors"`
	// ValidationErrors contains field-specific validation errors
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
	// Summary provides a high-level description
	Summary string `json:"summary"`
	// RequestID for tracing
	RequestID string `json:"request_id,omitempty"`
	// Context for additional metadata
	Context map[string]string `json:"context,omitempty"`
	// mu protects concurrent access
	mu sync.RWMutex
}

// NewErrorCollection creates a new error collection
func NewErrorCollection(summary string) *ErrorCollection {
	return &ErrorCollection{
		Errors:           make([]*CustomError, 0),
		ValidationErrors: make([]ValidationError, 0),
		Summary:          summary,
		Context:          make(map[string]string),
	}
}

// NewValidationErrorCollection creates a collection specifically for validation errors
func NewValidationErrorCollection() *ErrorCollection {
	return NewErrorCollection("validation failed")
}

// Add appends an error to the collection
func (ec *ErrorCollection) Add(err *CustomError) *ErrorCollection {
	if err == nil {
		return ec
	}

	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.Errors = append(ec.Errors, err)
	return ec
}

// AddValidation adds a field validation error
func (ec *ErrorCollection) AddValidation(field, message string) *ErrorCollection {
	return ec.AddValidationWithCode(field, message, "")
}

// AddValidationWithCode adds a field validation error with a specific code
func (ec *ErrorCollection) AddValidationWithCode(field, message, code string) *ErrorCollection {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.ValidationErrors = append(ec.ValidationErrors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
	return ec
}

// AddValidationWithValue adds a field validation error with the invalid value
func (ec *ErrorCollection) AddValidationWithValue(field, message, value string) *ErrorCollection {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.ValidationErrors = append(ec.ValidationErrors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
	return ec
}

// AddFieldError adds a validation error for a specific field from an existing CustomError
func (ec *ErrorCollection) AddFieldError(field string, err *CustomError) *ErrorCollection {
	if err == nil {
		return ec
	}

	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Extract field and value from error metadata if available
	value, _ := err.GetMetadata("value")
	code := err.Code

	ec.ValidationErrors = append(ec.ValidationErrors, ValidationError{
		Field:   field,
		Message: err.Message,
		Code:    code,
		Value:   value,
	})

	return ec
}

// WithRequestID sets the request ID for the collection
func (ec *ErrorCollection) WithRequestID(requestID string) *ErrorCollection {
	ec.RequestID = requestID
	return ec
}

// WithContext adds context metadata
func (ec *ErrorCollection) WithContext(key, value string) *ErrorCollection {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.Context[key] = value
	return ec
}

// Count returns the total number of errors (both regular and validation)
func (ec *ErrorCollection) Count() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	return len(ec.Errors) + len(ec.ValidationErrors)
}

// HasErrors returns true if the collection contains any errors
func (ec *ErrorCollection) HasErrors() bool {
	return ec.Count() > 0
}

// IsEmpty returns true if the collection is empty
func (ec *ErrorCollection) IsEmpty() bool {
	return ec.Count() == 0
}

// ValidationCount returns the number of validation errors
func (ec *ErrorCollection) ValidationCount() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	return len(ec.ValidationErrors)
}

// ErrorCount returns the number of regular errors
func (ec *ErrorCollection) ErrorCount() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	return len(ec.Errors)
}

// GetFieldErrors returns all validation errors for a specific field
func (ec *ErrorCollection) GetFieldErrors(field string) []ValidationError {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	var fieldErrors []ValidationError
	for _, err := range ec.ValidationErrors {
		if err.Field == field {
			fieldErrors = append(fieldErrors, err)
		}
	}
	return fieldErrors
}

// GetFields returns all fields that have validation errors
func (ec *ErrorCollection) GetFields() []string {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	fieldSet := make(map[string]bool)
	for _, err := range ec.ValidationErrors {
		fieldSet[err.Field] = true
	}

	fields := make([]string, 0, len(fieldSet))
	for field := range fieldSet {
		fields = append(fields, field)
	}

	return fields
}

// ToCustomError converts the collection to a single CustomError
func (ec *ErrorCollection) ToCustomError() *CustomError {
	if ec.IsEmpty() {
		return nil
	}

	ec.mu.RLock()
	defer ec.mu.RUnlock()

	// Determine the primary error type
	var sentinel error = ErrInvalidInput // Default to validation error
	if len(ec.ValidationErrors) == 0 && len(ec.Errors) > 0 {
		// Use the first error's sentinel if no validation errors
		sentinel = ec.Errors[0].Sentinel
	}

	// Create summary message
	message := ec.Summary
	if message == "" {
		message = fmt.Sprintf("multiple errors occurred (%d total)", ec.Count())
	}

	err := NewCustomError(sentinel, nil, message)

	if ec.RequestID != "" {
		err.WithRequestID(ec.RequestID)
	}

	// Add error counts to metadata
	err.WithMetadata("error_count", fmt.Sprintf("%d", len(ec.Errors)))
	err.WithMetadata("validation_error_count", fmt.Sprintf("%d", len(ec.ValidationErrors)))
	err.WithMetadata("total_error_count", fmt.Sprintf("%d", ec.Count()))

	// Add context metadata
	for key, value := range ec.Context {
		err.WithMetadata(key, value)
	}

	// Add fields with errors
	if len(ec.ValidationErrors) > 0 {
		fields := ec.GetFields()
		err.WithMetadata("validation_fields", strings.Join(fields, ","))
	}

	return err
}

// Error implements the error interface
func (ec *ErrorCollection) Error() string {
	if ec.IsEmpty() {
		return "no errors"
	}

	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if ec.Summary != "" {
		return fmt.Sprintf("%s (%d errors)", ec.Summary, ec.Count())
	}

	return fmt.Sprintf("multiple errors occurred (%d total)", ec.Count())
}

// ToJSON returns the collection as a JSON-serializable map
func (ec *ErrorCollection) ToJSON() map[string]interface{} {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	result := map[string]interface{}{
		"error": map[string]interface{}{
			"category":    string(ErrorCategoryValidation),
			"code":        "MULTIPLE_ERRORS",
			"message":     ec.Error(),
			"summary":     ec.Summary,
			"error_count": ec.Count(),
		},
	}

	if ec.RequestID != "" {
		result["error"].(map[string]interface{})["request_id"] = ec.RequestID
	}

	// Add validation errors if any
	if len(ec.ValidationErrors) > 0 {
		result["error"].(map[string]interface{})["validation_errors"] = ec.ValidationErrors
		result["error"].(map[string]interface{})["validation_count"] = len(ec.ValidationErrors)
	}

	// Add regular errors if any
	if len(ec.Errors) > 0 {
		errorDetails := make([]map[string]interface{}, len(ec.Errors))
		for i, err := range ec.Errors {
			errorDetails[i] = err.ToJSON()["error"].(map[string]interface{})
		}
		result["error"].(map[string]interface{})["errors"] = errorDetails
	}

	// Add context
	if len(ec.Context) > 0 {
		result["error"].(map[string]interface{})["context"] = ec.Context
	}

	return result
}

// ToClientJSON returns a client-safe JSON representation
func (ec *ErrorCollection) ToClientJSON() map[string]interface{} {
	result := ec.ToJSON()

	// Remove internal error details if in production mode
	config := GetConfig()
	if config.ProductionMode {
		errorData := result["error"].(map[string]interface{})

		// Keep only safe fields for validation errors
		safeResult := map[string]interface{}{
			"error": map[string]interface{}{
				"category": errorData["category"],
				"code":     errorData["code"],
				"message":  errorData["message"],
			},
		}

		if ec.RequestID != "" {
			safeResult["error"].(map[string]interface{})["request_id"] = ec.RequestID
		}

		// Keep validation errors as they're usually safe to expose
		if validationErrors, exists := errorData["validation_errors"]; exists {
			safeResult["error"].(map[string]interface{})["validation_errors"] = validationErrors
		}

		return safeResult
	}

	return result
}

// ToHTTPStatus returns the appropriate HTTP status code
func (ec *ErrorCollection) ToHTTPStatus() int {
	if ec.IsEmpty() {
		return 200 // OK
	}

	// If we have validation errors, return 400 (Bad Request)
	if ec.ValidationCount() > 0 {
		return 400
	}

	// Otherwise, use the status from the first error
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if len(ec.Errors) > 0 {
		return ec.Errors[0].ToHTTPStatus()
	}

	return 400 // Default to Bad Request
}

// Builder pattern for error collections

// ErrorCollectionBuilder provides a fluent interface for building error collections
type ErrorCollectionBuilder struct {
	collection *ErrorCollection
	ctx        context.Context
}

// NewErrorCollectionBuilder creates a new error collection builder
func NewErrorCollectionBuilder(summary string) *ErrorCollectionBuilder {
	return &ErrorCollectionBuilder{
		collection: NewErrorCollection(summary),
	}
}

// NewValidationCollectionBuilder creates a builder for validation errors
func NewValidationCollectionBuilder() *ErrorCollectionBuilder {
	return NewErrorCollectionBuilder("validation failed")
}

// WithContext sets the context for the builder
func (b *ErrorCollectionBuilder) WithContext(ctx context.Context) *ErrorCollectionBuilder {
	b.ctx = ctx
	if ctx != nil {
		// Auto-extract common context values
		if requestID := GetRequestIDFromContext(ctx); requestID != "" {
			b.collection.WithRequestID(requestID)
		}
		if userID := GetUserIDFromContext(ctx); userID != "" {
			b.collection.WithContext("user_id", userID)
		}
		if traceID := GetTraceIDFromContext(ctx); traceID != "" {
			b.collection.WithContext("trace_id", traceID)
		}
	}
	return b
}

// AddError adds a custom error to the collection
func (b *ErrorCollectionBuilder) AddError(err *CustomError) *ErrorCollectionBuilder {
	b.collection.Add(err)
	return b
}

// AddValidation adds a validation error
func (b *ErrorCollectionBuilder) AddValidation(field, message string) *ErrorCollectionBuilder {
	b.collection.AddValidation(field, message)
	return b
}

// AddFieldError adds a field error from a CustomError
func (b *ErrorCollectionBuilder) AddFieldError(field string, err *CustomError) *ErrorCollectionBuilder {
	b.collection.AddFieldError(field, err)
	return b
}

// AddValidationWithCode adds a validation error with a specific code
func (b *ErrorCollectionBuilder) AddValidationWithCode(field, message, code string) *ErrorCollectionBuilder {
	b.collection.AddValidationWithCode(field, message, code)
	return b
}

// AddValidationWithValue adds a validation error with the invalid value
func (b *ErrorCollectionBuilder) AddValidationWithValue(field, message, value string) *ErrorCollectionBuilder {
	b.collection.AddValidationWithValue(field, message, value)
	return b
}

// Build returns the completed error collection
func (b *ErrorCollectionBuilder) Build() *ErrorCollection {
	return b.collection
}

// BuildIfHasErrors returns the collection only if it has errors, nil otherwise
func (b *ErrorCollectionBuilder) BuildIfHasErrors() *ErrorCollection {
	if b.collection.HasErrors() {
		return b.collection
	}
	return nil
}

// Utility functions for common validation patterns

// ValidateRequired checks if a field is required and not empty
func ValidateRequired(field, value string, collection *ErrorCollection) {
	if strings.TrimSpace(value) == "" {
		collection.AddValidationWithCode(field, fmt.Sprintf("%s is required", field), "REQUIRED")
	}
}

// ValidateLength checks field length constraints
func ValidateLength(field, value string, min, max int, collection *ErrorCollection) {
	length := len(value)
	if length < min {
		collection.AddValidationWithCode(field,
			fmt.Sprintf("%s must be at least %d characters", field, min),
			"TOO_SHORT")
	}
	if max > 0 && length > max {
		collection.AddValidationWithCode(field,
			fmt.Sprintf("%s must be at most %d characters", field, max),
			"TOO_LONG")
	}
}

// ValidateEmail checks if a value is a valid email format
func ValidateEmail(field, value string, collection *ErrorCollection) {
	if value != "" && !strings.Contains(value, "@") {
		collection.AddValidationWithCode(field,
			fmt.Sprintf("%s must be a valid email address", field),
			"INVALID_EMAIL")
	}
}

// JSON marshaling support

// MarshalJSON implements json.Marshaler interface
func (ec *ErrorCollection) MarshalJSON() ([]byte, error) {
	return json.Marshal(ec.ToJSON())
}
