package cuserr

import "errors"

// Common sentinel errors for use across services
// These provide type-safe error categorization and enable errors.Is() checks
var (
	// ErrNotFound indicates a requested resource was not found
	ErrNotFound = errors.New(SENTINEL_MSG_NOT_FOUND)
	// ErrAlreadyExists indicates a resource already exists
	ErrAlreadyExists = errors.New(SENTINEL_MSG_ALREADY_EXISTS)
	// ErrInvalidInput indicates invalid input data
	ErrInvalidInput = errors.New(SENTINEL_MSG_INVALID_INPUT)
	// ErrUnauthorized indicates authentication failure
	ErrUnauthorized = errors.New(SENTINEL_MSG_UNAUTHORIZED)
	// ErrForbidden indicates authorization failure
	ErrForbidden = errors.New(SENTINEL_MSG_FORBIDDEN)
	// ErrInternal indicates an internal error
	ErrInternal = errors.New(SENTINEL_MSG_INTERNAL)
	// ErrTimeout indicates an operation timeout
	ErrTimeout = errors.New(SENTINEL_MSG_TIMEOUT)
	// ErrRateLimit indicates rate limit exceeded
	ErrRateLimit = errors.New(SENTINEL_MSG_RATE_LIMIT)
	// ErrExternal indicates external service failure
	ErrExternal = errors.New(SENTINEL_MSG_EXTERNAL)
)

// mapSentinelToCategory maps sentinel errors to their appropriate categories
func mapSentinelToCategory(sentinel error) ErrorCategory {
	switch {
	case errors.Is(sentinel, ErrNotFound):
		return ErrorCategoryNotFound
	case errors.Is(sentinel, ErrAlreadyExists):
		return ErrorCategoryConflict
	case errors.Is(sentinel, ErrInvalidInput):
		return ErrorCategoryValidation
	case errors.Is(sentinel, ErrUnauthorized):
		return ErrorCategoryUnauthorized
	case errors.Is(sentinel, ErrForbidden):
		return ErrorCategoryForbidden
	case errors.Is(sentinel, ErrTimeout):
		return ErrorCategoryTimeout
	case errors.Is(sentinel, ErrRateLimit):
		return ErrorCategoryRateLimit
	case errors.Is(sentinel, ErrExternal):
		return ErrorCategoryExternal
	default:
		return ErrorCategoryInternal
	}
}

// generateErrorCode creates consistent error codes from sentinel errors
func generateErrorCode(sentinel error) string {
	switch {
	case errors.Is(sentinel, ErrNotFound):
		return ERROR_CODE_NOT_FOUND
	case errors.Is(sentinel, ErrAlreadyExists):
		return ERROR_CODE_ALREADY_EXISTS
	case errors.Is(sentinel, ErrInvalidInput):
		return ERROR_CODE_INVALID_INPUT
	case errors.Is(sentinel, ErrUnauthorized):
		return ERROR_CODE_UNAUTHORIZED
	case errors.Is(sentinel, ErrForbidden):
		return ERROR_CODE_FORBIDDEN
	case errors.Is(sentinel, ErrTimeout):
		return ERROR_CODE_TIMEOUT
	case errors.Is(sentinel, ErrRateLimit):
		return ERROR_CODE_RATE_LIMIT
	case errors.Is(sentinel, ErrExternal):
		return ERROR_CODE_EXTERNAL_ERROR
	default:
		return ERROR_CODE_INTERNAL_ERROR
	}
}
