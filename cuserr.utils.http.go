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

// ToJSON converts error to JSON response format
func (e *CustomError) ToJSON() map[string]interface{} {
	metadata := e.GetAllMetadata() // Thread-safe metadata access

	errorData := map[string]interface{}{
		JSON_FIELD_CODE:      e.Code,
		JSON_FIELD_MESSAGE:   e.Message,
		JSON_FIELD_CATEGORY:  e.Category,
		JSON_FIELD_TIMESTAMP: e.Timestamp.Format(time.RFC3339),
	}

	// Only include non-empty fields
	if len(metadata) > 0 {
		errorData[JSON_FIELD_METADATA] = metadata
	}

	if e.RequestID != "" {
		errorData[JSON_FIELD_REQUEST_ID] = e.RequestID
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

	if errorData[JSON_FIELD_REQUEST_ID] != nil {
		result += fmt.Sprintf(`,"%s":"%s"`, JSON_FIELD_REQUEST_ID, errorData[JSON_FIELD_REQUEST_ID])
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

// ClientSafeMessage returns a safe message for client consumption
// In production mode, it may hide sensitive error details
func (e *CustomError) ClientSafeMessage() string {
	if globalConfig.ProductionMode {
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
	return e.Message
}

// ToClientJSON converts error to client-safe JSON format
func (e *CustomError) ToClientJSON() map[string]interface{} {
	metadata := e.GetAllMetadata() // Thread-safe metadata access

	// Filter sensitive metadata in production
	if globalConfig.ProductionMode {
		filteredMetadata := make(map[string]string)
		for k, v := range metadata {
			// Only include non-sensitive metadata keys
			switch k {
			case "user_id", "request_id", "trace_id", "correlation_id":
				// Keep safe identifiers
			default:
				// Skip potentially sensitive data in production
				continue
			}
			filteredMetadata[k] = v
		}
		metadata = filteredMetadata
	}

	errorData := map[string]interface{}{
		JSON_FIELD_CODE:      e.Code,
		JSON_FIELD_MESSAGE:   e.ClientSafeMessage(),
		JSON_FIELD_CATEGORY:  e.Category,
		JSON_FIELD_TIMESTAMP: e.Timestamp.Format(time.RFC3339),
	}

	// Only include non-empty fields
	if len(metadata) > 0 {
		errorData[JSON_FIELD_METADATA] = metadata
	}

	if e.RequestID != "" {
		errorData[JSON_FIELD_REQUEST_ID] = e.RequestID
	}

	return map[string]interface{}{
		JSON_FIELD_ERROR: errorData,
	}
}
