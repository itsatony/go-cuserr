package cuserr

import (
	"fmt"
	"time"
)

// ToHTTPStatus maps error category to HTTP status code
func (e *CustomError) ToHTTPStatus() int {
	return CategoryToHTTPStatus(e.Category)
}

// CategoryToHTTPStatus maps error categories to HTTP status codes
func CategoryToHTTPStatus(category ErrorCategory) int {
	switch category {
	case ErrorCategoryValidation:
		return HTTP_STATUS_BAD_REQUEST
	case ErrorCategoryUnauthorized:
		return HTTP_STATUS_UNAUTHORIZED
	case ErrorCategoryForbidden:
		return HTTP_STATUS_FORBIDDEN
	case ErrorCategoryNotFound:
		return HTTP_STATUS_NOT_FOUND
	case ErrorCategoryTimeout:
		return HTTP_STATUS_REQUEST_TIMEOUT
	case ErrorCategoryConflict:
		return HTTP_STATUS_CONFLICT
	case ErrorCategoryRateLimit:
		return HTTP_STATUS_TOO_MANY_REQUESTS
	case ErrorCategoryExternal:
		return HTTP_STATUS_BAD_GATEWAY
	case ErrorCategoryInternal:
		fallthrough
	default:
		return HTTP_STATUS_INTERNAL_SERVER_ERROR
	}
}

// ToJSON converts error to JSON response format with all available fields.
// This is the internal/developer representation including all error details.
// For public/external consumption, use ToPublicJSON() instead.
func (e *CustomError) ToJSON() map[string]interface{} {
	metadata := e.GetAllMetadata() // Thread-safe metadata access

	errorData := map[string]interface{}{
		JSON_FIELD_CODE:      e.Code,
		JSON_FIELD_MESSAGE:   e.Message,
		JSON_FIELD_CATEGORY:  e.Category,
		JSON_FIELD_TIMESTAMP: e.Timestamp.Format(time.RFC3339),
	}

	// Add universal user-facing fields if set
	if e.PublicMessage != "" {
		errorData["public_message"] = e.PublicMessage
	}

	if e.HelpURL != "" {
		errorData["help_url"] = e.HelpURL
	}

	if len(e.Suggestions) > 0 {
		errorData["suggestions"] = e.Suggestions
	}

	if e.MessageKey != "" {
		errorData["message_key"] = e.MessageKey
	}

	// Add universal tracing fields
	if e.OperationID != "" {
		errorData["operation_id"] = e.OperationID
	}

	if e.TraceID != "" {
		errorData["trace_id"] = e.TraceID
	}

	if e.SpanID != "" {
		errorData["span_id"] = e.SpanID
	}

	// Include metadata
	if len(metadata) > 0 {
		errorData[JSON_FIELD_METADATA] = metadata
	}

	return map[string]interface{}{
		JSON_FIELD_ERROR: errorData,
	}
}

// ToJSONString converts error to JSON string format
func (e *CustomError) ToJSONString() string {
	jsonData := e.ToJSON()

	// Simple JSON serialization without external dependencies
	errorData := jsonData[JSON_FIELD_ERROR].(map[string]interface{})

	result := fmt.Sprintf(`{"%s":{`, JSON_FIELD_ERROR)
	result += fmt.Sprintf(`"%s":"%s",`, JSON_FIELD_CODE, errorData[JSON_FIELD_CODE])
	result += fmt.Sprintf(`"%s":"%s",`, JSON_FIELD_MESSAGE, errorData[JSON_FIELD_MESSAGE])
	result += fmt.Sprintf(`"%s":"%s",`, JSON_FIELD_CATEGORY, errorData[JSON_FIELD_CATEGORY])
	result += fmt.Sprintf(`"%s":"%s"`, JSON_FIELD_TIMESTAMP, errorData[JSON_FIELD_TIMESTAMP])

	// Add operation_id if present
	if errorData["operation_id"] != nil && errorData["operation_id"] != "" {
		result += fmt.Sprintf(`,"%s":"%s"`, "operation_id", errorData["operation_id"])
	}

	if errorData[JSON_FIELD_METADATA] != nil {
		metadata := errorData[JSON_FIELD_METADATA].(map[string]string)
		if len(metadata) > 0 {
			result += fmt.Sprintf(`,"%s":{`, JSON_FIELD_METADATA)
			first := true
			for k, v := range metadata {
				if !first {
					result += ","
				}
				result += fmt.Sprintf(`"%s":"%s"`, k, v)
				first = false
			}
			result += "}"
		}
	}

	result += "}}"
	return result
}

// GetPublicMessage returns a safe, user-friendly message for external consumption.
// This is the universal method that works across all protocols (HTTP, gRPC, CLI, etc.).
//
// Priority order:
//  1. PublicMessage field if explicitly set
//  2. Production-safe message for internal/external errors in production mode
//  3. Developer message (fallback)
//
// Example:
//
//	err := cuserr.Internal("database", dbErr,
//	    cuserr.WithPublicMessage("Service temporarily unavailable"),
//	)
//	msg := err.GetPublicMessage()  // Returns "Service temporarily unavailable"
func (e *CustomError) GetPublicMessage() string {
	// Priority 1: Use explicitly set public message
	if e.PublicMessage != "" {
		return e.PublicMessage
	}

	// Priority 2: Apply production mode safety
	config := GetConfig()
	if config.ProductionMode {
		// In production, return generic messages for internal errors
		switch e.Category {
		case ErrorCategoryInternal:
			return "An internal error occurred"
		case ErrorCategoryExternal:
			return "A service is temporarily unavailable"
		default:
			return e.Message
		}
	}

	// Priority 3: Fallback to developer message
	return e.Message
}

// ToPublicJSON converts error to public-safe JSON format suitable for external consumption.
// This is the universal method that works across all protocols.
//
// Behavior:
//   - Uses GetPublicMessage() for the message field (respects PublicMessage field and production mode)
//   - Includes PublicMessage, HelpURL, Suggestions, OperationID, TraceID, SpanID if set
//   - In production mode, filters sensitive metadata (keeps only safe identifiers)
//   - Thread-safe metadata access
//
// Example:
//
//	err := cuserr.Validation("email", "invalid format",
//	    cuserr.WithPublicMessage("Please enter a valid email address"),
//	    cuserr.WithSuggestions("Check for typos", "Ensure @ symbol is present"),
//	    cuserr.WithHelpURL("https://docs.example.com/validation#email"),
//	)
//	json := err.ToPublicJSON()
func (e *CustomError) ToPublicJSON() map[string]interface{} {
	metadata := e.GetAllMetadata() // Thread-safe metadata access

	// Filter sensitive metadata in production
	config := GetConfig()
	if config.ProductionMode {
		filteredMetadata := make(map[string]string)
		for k, v := range metadata {
			// Only include non-sensitive metadata keys
			switch k {
			case "user_id", "request_id", "trace_id", "span_id", "correlation_id", "operation_id":
				// Keep safe identifiers
				filteredMetadata[k] = v
			default:
				// Skip potentially sensitive data in production
				continue
			}
		}
		metadata = filteredMetadata
	}

	errorData := map[string]interface{}{
		JSON_FIELD_CODE:      e.Code,
		JSON_FIELD_MESSAGE:   e.GetPublicMessage(),
		JSON_FIELD_CATEGORY:  e.Category,
		JSON_FIELD_TIMESTAMP: e.Timestamp.Format(time.RFC3339),
	}

	// Add universal fields if set
	if e.PublicMessage != "" {
		errorData["public_message"] = e.PublicMessage
	}

	if e.HelpURL != "" {
		errorData["help_url"] = e.HelpURL
	}

	if len(e.Suggestions) > 0 {
		errorData["suggestions"] = e.Suggestions
	}

	if e.MessageKey != "" {
		errorData["message_key"] = e.MessageKey
	}

	// Add tracing fields (universal across protocols)
	if e.OperationID != "" {
		errorData["operation_id"] = e.OperationID
	}

	if e.TraceID != "" {
		errorData["trace_id"] = e.TraceID
	}

	if e.SpanID != "" {
		errorData["span_id"] = e.SpanID
	}

	// Include filtered metadata
	if len(metadata) > 0 {
		errorData[JSON_FIELD_METADATA] = metadata
	}

	return map[string]interface{}{
		JSON_FIELD_ERROR: errorData,
	}
}
