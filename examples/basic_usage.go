// Package main demonstrates basic usage patterns of the cuserr package
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/itsatony/go-cuserr"
)

func main() {
	fmt.Println("=== Basic cuserr Usage Examples ===")

	// Example 1: Creating basic custom errors
	basicErrorExample()

	// Example 2: Error categorization and HTTP status mapping
	httpStatusExample()

	// Example 3: Rich error context with metadata
	richContextExample()

	// Example 4: Error wrapping and unwrapping
	errorWrappingExample()

	// Example 5: Error checking patterns
	errorCheckingExample()

	// Example 6: JSON serialization
	jsonSerializationExample()
}

func basicErrorExample() {
	fmt.Println("1. Basic Error Creation:")

	// Create a simple validation error
	err := cuserr.NewCustomError(cuserr.ErrInvalidInput, nil, "email address is required")
	fmt.Printf("   Error: %s\n", err.Error())
	fmt.Printf("   Category: %s\n", err.Category)
	fmt.Printf("   Code: %s\n", err.Code)

	// Create error with explicit category
	err2 := cuserr.NewCustomErrorWithCategory(
		cuserr.ErrorCategoryTimeout,
		"DB_TIMEOUT",
		"database connection timeout after 30 seconds")
	fmt.Printf("   Custom Error: %s (HTTP: %d)\n\n", err2.Error(), err2.ToHTTPStatus())
}

func httpStatusExample() {
	fmt.Println("2. HTTP Status Code Mapping:")

	errorTypes := []struct {
		sentinel cuserr.ErrorCategory
		message  string
	}{
		{cuserr.ErrorCategoryValidation, "invalid input data"},
		{cuserr.ErrorCategoryNotFound, "user not found"},
		{cuserr.ErrorCategoryUnauthorized, "authentication required"},
		{cuserr.ErrorCategoryForbidden, "insufficient permissions"},
		{cuserr.ErrorCategoryConflict, "resource already exists"},
		{cuserr.ErrorCategoryTimeout, "operation timeout"},
		{cuserr.ErrorCategoryRateLimit, "rate limit exceeded"},
		{cuserr.ErrorCategoryExternal, "external service unavailable"},
		{cuserr.ErrorCategoryInternal, "internal server error"},
	}

	for _, et := range errorTypes {
		err := cuserr.NewCustomErrorWithCategory(et.sentinel, "EXAMPLE", et.message)
		fmt.Printf("   %s -> HTTP %d\n", et.sentinel, err.ToHTTPStatus())
	}
	fmt.Println()
}

func richContextExample() {
	fmt.Println("3. Rich Error Context:")

	// Create error with metadata and request ID
	err := cuserr.NewCustomError(cuserr.ErrNotFound, nil, "user profile not found")
	err.WithMetadata("user_id", "usr_12345")
	err.WithMetadata("operation", "get_profile")
	err.WithMetadata("client_ip", "192.168.1.100")
	err.WithRequestID("req_abc123")

	fmt.Printf("   Short: %s\n", err.ShortError())
	fmt.Printf("   Metadata: %v\n", err.GetAllMetadata())

	// Access specific metadata
	if userID, exists := err.GetMetadata("user_id"); exists {
		fmt.Printf("   User ID from error: %s\n", userID)
	}
	fmt.Println()
}

func errorWrappingExample() {
	fmt.Println("4. Error Wrapping:")

	// Simulate a database error
	dbErr := errors.New("connection refused")

	// Wrap with custom error
	wrappedErr := cuserr.NewCustomError(cuserr.ErrInternal, dbErr, "failed to save user data")
	wrappedErr.WithMetadata("table", "users")
	wrappedErr.WithMetadata("operation", "INSERT")

	fmt.Printf("   Wrapped: %s\n", wrappedErr.Error())

	// Check if original error is still accessible
	if errors.Is(wrappedErr, dbErr) {
		fmt.Println("   ✓ Original error is accessible via errors.Is()")
	}

	// Unwrap to get original error
	if unwrapped := errors.Unwrap(wrappedErr); unwrapped != nil {
		fmt.Printf("   Unwrapped: %s\n", unwrapped.Error())
	}

	// Use the convenience wrapper function
	wrapped2 := cuserr.WrapWithCustomError(
		dbErr,
		cuserr.ErrorCategoryExternal,
		"DB_CONNECTION_FAILED",
		"database is temporarily unavailable")
	fmt.Printf("   Convenience wrapped: %s\n\n", wrapped2.Error())
}

func errorCheckingExample() {
	fmt.Println("5. Error Checking Patterns:")

	err := cuserr.NewCustomError(cuserr.ErrUnauthorized, nil, "invalid API key")
	err.WithMetadata("api_key_id", "key_789")

	// Check by sentinel error
	if errors.Is(err, cuserr.ErrUnauthorized) {
		fmt.Println("   ✓ Detected as unauthorized error (sentinel)")
	}

	// Check by category
	if cuserr.IsErrorCategory(err, cuserr.ErrorCategoryUnauthorized) {
		fmt.Println("   ✓ Detected as unauthorized error (category)")
	}

	// Check by error code
	if cuserr.IsErrorCode(err, "UNAUTHORIZED") {
		fmt.Println("   ✓ Detected by error code")
	}

	// Extract error information
	category := cuserr.GetErrorCategory(err)
	code := cuserr.GetErrorCode(err)
	fmt.Printf("   Category: %s, Code: %s\n", category, code)

	// Check metadata
	if keyID, exists := cuserr.GetErrorMetadata(err, "api_key_id"); exists {
		fmt.Printf("   API Key ID: %s\n", keyID)
	}
	fmt.Println()
}

func jsonSerializationExample() {
	fmt.Println("6. JSON Serialization:")

	err := cuserr.NewCustomError(cuserr.ErrInvalidInput, nil, "validation failed")
	err.WithMetadata("field", "email")
	err.WithMetadata("reason", "invalid format")
	err.WithRequestID("req_json_example")

	// Convert to JSON map
	jsonMap := err.ToJSON()
	fmt.Printf("   JSON Map: %+v\n", jsonMap)

	// Convert to JSON string (basic implementation)
	jsonString := err.ToJSONString()
	fmt.Printf("   JSON String: %s\n", jsonString)

	// Client-safe JSON (production mode)
	cuserr.SetConfig(&cuserr.Config{
		EnableStackTrace: false,
		ProductionMode:   true,
	})

	clientJSON := err.ToClientJSON()
	fmt.Printf("   Client-safe JSON: %+v\n", clientJSON)
	fmt.Println()
}

// Additional helper functions

// simulateServiceError demonstrates a typical service error pattern
func simulateServiceError() error {
	// This would be your actual service logic
	userID := "usr_404"

	// Simulate user not found
	return cuserr.NewCustomError(cuserr.ErrNotFound, nil, "user not found").
		WithMetadata("user_id", userID).
		WithMetadata("operation", "get_user").
		WithRequestID("req_simulate_123")
}

// handleServiceError demonstrates error handling pattern
func handleServiceError() {
	err := simulateServiceError()
	if err != nil {
		var customErr *cuserr.CustomError
		if errors.As(err, &customErr) {
			log.Printf("Service Error: %s", customErr.ShortError())
			log.Printf("Details: %s", customErr.DetailedError())

			// Handle specific error types
			switch customErr.Category {
			case cuserr.ErrorCategoryNotFound:
				log.Println("Handle not found error")
			case cuserr.ErrorCategoryValidation:
				log.Println("Handle validation error")
			default:
				log.Println("Handle generic error")
			}
		}
	}
}
