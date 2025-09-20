package cuserr

const (
	// Package information constants

	// PACKAGE_NAME defines the name of the cuserr package
	PACKAGE_NAME = "cuserr"
	// PACKAGE_VERSION defines the current version of the cuserr package
	PACKAGE_VERSION = "0.2.0"

	// Error codes for consistent identification

	// ERROR_CODE_NOT_FOUND represents resource not found errors
	ERROR_CODE_NOT_FOUND = "NOT_FOUND"
	// ERROR_CODE_ALREADY_EXISTS represents resource conflict/duplicate errors
	ERROR_CODE_ALREADY_EXISTS = "ALREADY_EXISTS"
	// ERROR_CODE_INVALID_INPUT represents validation and input errors
	ERROR_CODE_INVALID_INPUT = "INVALID_INPUT"
	// ERROR_CODE_UNAUTHORIZED represents authentication errors
	ERROR_CODE_UNAUTHORIZED = "UNAUTHORIZED"
	// ERROR_CODE_FORBIDDEN represents authorization errors
	ERROR_CODE_FORBIDDEN = "FORBIDDEN"
	// ERROR_CODE_TIMEOUT represents operation timeout errors
	ERROR_CODE_TIMEOUT = "TIMEOUT"
	// ERROR_CODE_RATE_LIMIT represents rate limiting errors
	ERROR_CODE_RATE_LIMIT = "RATE_LIMIT"
	// ERROR_CODE_EXTERNAL_ERROR represents external service errors
	ERROR_CODE_EXTERNAL_ERROR = "EXTERNAL_ERROR"
	// ERROR_CODE_INTERNAL_ERROR represents internal server errors
	ERROR_CODE_INTERNAL_ERROR = "INTERNAL_ERROR"

	// Error category string constants

	// CATEGORY_VALIDATION represents input validation error category
	CATEGORY_VALIDATION = "validation"
	// CATEGORY_NOT_FOUND represents resource not found error category
	CATEGORY_NOT_FOUND = "not_found"
	// CATEGORY_CONFLICT represents resource conflict error category
	CATEGORY_CONFLICT = "conflict"
	// CATEGORY_UNAUTHORIZED represents authentication error category
	CATEGORY_UNAUTHORIZED = "unauthorized"
	// CATEGORY_FORBIDDEN represents authorization error category
	CATEGORY_FORBIDDEN = "forbidden"
	// CATEGORY_INTERNAL represents internal server error category
	CATEGORY_INTERNAL = "internal"
	// CATEGORY_TIMEOUT represents operation timeout error category
	CATEGORY_TIMEOUT = "timeout"
	// CATEGORY_RATE_LIMIT represents rate limiting error category
	CATEGORY_RATE_LIMIT = "rate_limit"
	// CATEGORY_EXTERNAL represents external service error category
	CATEGORY_EXTERNAL = "external"

	// Sentinel error message constants

	// SENTINEL_MSG_NOT_FOUND represents default message for not found errors
	SENTINEL_MSG_NOT_FOUND = "resource not found"
	// SENTINEL_MSG_ALREADY_EXISTS represents default message for conflict errors
	SENTINEL_MSG_ALREADY_EXISTS = "resource already exists"
	// SENTINEL_MSG_INVALID_INPUT represents default message for validation errors
	SENTINEL_MSG_INVALID_INPUT = "invalid input"
	// SENTINEL_MSG_UNAUTHORIZED represents default message for authentication errors
	SENTINEL_MSG_UNAUTHORIZED = "unauthorized"
	// SENTINEL_MSG_FORBIDDEN represents default message for authorization errors
	SENTINEL_MSG_FORBIDDEN = "forbidden"
	// SENTINEL_MSG_INTERNAL represents default message for internal errors
	SENTINEL_MSG_INTERNAL = "internal error"
	// SENTINEL_MSG_TIMEOUT represents default message for timeout errors
	SENTINEL_MSG_TIMEOUT = "operation timeout"
	// SENTINEL_MSG_RATE_LIMIT represents default message for rate limit errors
	SENTINEL_MSG_RATE_LIMIT = "rate limit exceeded"
	// SENTINEL_MSG_EXTERNAL represents default message for external service errors
	SENTINEL_MSG_EXTERNAL = "external service error"

	// Stack trace configuration constants
	DEFAULT_STACK_DEPTH = 10
	STACK_SKIP_FRAMES   = 2

	// JSON field names for serialization
	JSON_FIELD_ERROR      = "error"
	JSON_FIELD_CODE       = "code"
	JSON_FIELD_MESSAGE    = "message"
	JSON_FIELD_CATEGORY   = "category"
	JSON_FIELD_METADATA   = "metadata"
	JSON_FIELD_REQUEST_ID = "request_id"
	JSON_FIELD_TIMESTAMP  = "timestamp"

	// HTTP status codes
	HTTP_STATUS_BAD_REQUEST           = 400
	HTTP_STATUS_UNAUTHORIZED          = 401
	HTTP_STATUS_FORBIDDEN             = 403
	HTTP_STATUS_NOT_FOUND             = 404
	HTTP_STATUS_REQUEST_TIMEOUT       = 408
	HTTP_STATUS_CONFLICT              = 409
	HTTP_STATUS_TOO_MANY_REQUESTS     = 429
	HTTP_STATUS_INTERNAL_SERVER_ERROR = 500
	HTTP_STATUS_BAD_GATEWAY           = 502
	HTTP_STATUS_DEFAULT_ERROR         = 500

	// Log message templates for detailed error reporting
	LOG_TEMPLATE_ERROR_CREATED   = "CustomError created: category=%s, code=%s, message=%s"
	LOG_TEMPLATE_ERROR_WITH_META = "CustomError with metadata: %s=%s"
	LOG_TEMPLATE_STACK_FRAME     = "  %s\n    %s:%d"
	LOG_TEMPLATE_ERROR_DETAIL    = "Error: %s\nCategory: %s, Code: %s"
	LOG_TEMPLATE_REQUEST_ID      = "RequestID: %s"
	LOG_TEMPLATE_WRAPPED_ERROR   = "Wrapped: %v"

	// Configuration environment variables
	ENV_ENABLE_STACK_TRACE = "CUSERR_ENABLE_STACK_TRACE"
	ENV_MAX_STACK_DEPTH    = "CUSERR_MAX_STACK_DEPTH"
	ENV_PRODUCTION_MODE    = "CUSERR_PRODUCTION_MODE"

	// Function names for stack trace filtering
	MAIN_FUNCTION_NAME    = "main.main"
	TESTING_FUNCTION_NAME = "testing."
)
