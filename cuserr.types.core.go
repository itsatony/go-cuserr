package cuserr

import (
	"sync"
	"time"
)

// ErrorCategory defines the type of error for proper HTTP mapping and handling
type ErrorCategory string

const (
	// ErrorCategoryValidation indicates input validation failures (400)
	ErrorCategoryValidation ErrorCategory = CATEGORY_VALIDATION
	// ErrorCategoryNotFound indicates resource not found (404)
	ErrorCategoryNotFound ErrorCategory = CATEGORY_NOT_FOUND
	// ErrorCategoryConflict indicates resource conflicts (409)
	ErrorCategoryConflict ErrorCategory = CATEGORY_CONFLICT
	// ErrorCategoryUnauthorized indicates authentication failures (401)
	ErrorCategoryUnauthorized ErrorCategory = CATEGORY_UNAUTHORIZED
	// ErrorCategoryForbidden indicates authorization failures (403)
	ErrorCategoryForbidden ErrorCategory = CATEGORY_FORBIDDEN
	// ErrorCategoryInternal indicates internal server errors (500)
	ErrorCategoryInternal ErrorCategory = CATEGORY_INTERNAL
	// ErrorCategoryTimeout indicates operation timeouts (408)
	ErrorCategoryTimeout ErrorCategory = CATEGORY_TIMEOUT
	// ErrorCategoryRateLimit indicates rate limiting (429)
	ErrorCategoryRateLimit ErrorCategory = CATEGORY_RATE_LIMIT
	// ErrorCategoryExternal indicates external service failures (502)
	ErrorCategoryExternal ErrorCategory = CATEGORY_EXTERNAL
)

// StackFrame represents a single frame in the stack trace
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// CustomError provides rich error context for debugging and client responses
// Thread-safe implementation with mutex protection for metadata and stack trace operations
type CustomError struct {
	// Category determines HTTP status code mapping
	Category ErrorCategory `json:"category"`
	// Code is a unique error code for this error type
	Code string `json:"code"`
	// Message is a human-readable error message
	Message string `json:"message"`
	// Metadata contains additional context
	metadata map[string]string
	// RequestID for tracing
	RequestID string `json:"request_id,omitempty"`
	// Timestamp when error occurred
	Timestamp time.Time `json:"timestamp"`
	// StackTrace for debugging (not serialized to JSON)
	stackTrace []StackFrame
	// Wrapped is the underlying error
	Wrapped error `json:"-"`
	// Sentinel is the base error for categorization
	Sentinel error `json:"-"`
	// mu protects metadata and stackTrace access for thread safety
	mu sync.RWMutex
}

// Config holds package-level configuration
type Config struct {
	// EnableStackTrace controls whether to capture stack traces
	EnableStackTrace bool
	// MaxStackDepth controls how many stack frames to capture
	MaxStackDepth int
	// ProductionMode controls error detail exposure
	ProductionMode bool
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		EnableStackTrace: true,
		MaxStackDepth:    DEFAULT_STACK_DEPTH,
		ProductionMode:   false,
	}
}

// Package-level configuration instance with thread safety
var (
	globalConfig   = DefaultConfig()
	globalConfigMu sync.RWMutex
)

// SetConfig updates the global configuration in a thread-safe manner
func SetConfig(config *Config) {
	if config != nil {
		globalConfigMu.Lock()
		globalConfig = config
		globalConfigMu.Unlock()
	}
}

// GetConfig returns the current global configuration in a thread-safe manner
func GetConfig() *Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()

	// Return a copy to prevent external modification
	return &Config{
		EnableStackTrace: globalConfig.EnableStackTrace,
		MaxStackDepth:    globalConfig.MaxStackDepth,
		ProductionMode:   globalConfig.ProductionMode,
	}
}
