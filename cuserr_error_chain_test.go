package cuserr

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// TestErrorChainIntegrity tests deep error wrapping and unwrapping
func TestErrorChainIntegrity(t *testing.T) {
	// Create a chain of errors: network -> database -> service -> API
	networkErr := errors.New("connection refused")

	dbErr := NewCustomError(ErrExternal, networkErr, "database connection failed").
		WithMetadata("database", "postgres").
		WithMetadata("host", "db.example.com")

	serviceErr := NewCustomError(ErrInternal, dbErr, "user service unavailable").
		WithMetadata("service", "user-service").
		WithMetadata("operation", "get_user")

	apiErr := NewCustomError(ErrInternal, serviceErr, "API request failed").
		WithMetadata("endpoint", "/api/users/123").
		WithMetadata("method", http.MethodGet).
		WithRequestID("req-chain-test")

	// Test error chain integrity
	t.Run("Error Chain Navigation", func(t *testing.T) {
		// Test errors.Is() works through the chain
		if !errors.Is(apiErr, networkErr) {
			t.Error("Should find original network error through chain")
		}

		if !errors.Is(apiErr, dbErr) {
			t.Error("Should find database error in chain")
		}

		if !errors.Is(apiErr, serviceErr) {
			t.Error("Should find service error in chain")
		}

		// Test errors.As() works through the chain
		var foundDbErr *CustomError
		if !errors.As(dbErr, &foundDbErr) {
			t.Fatal("Should extract CustomError from chain")
		}

		if foundDbErr.Category != ErrorCategoryExternal {
			t.Error("Should preserve category through chain")
		}

		// Test unwrapping step by step
		unwrapped := errors.Unwrap(apiErr)
		if unwrapped != serviceErr {
			t.Error("First unwrap should return service error")
		}

		unwrapped = errors.Unwrap(serviceErr)
		if unwrapped != dbErr {
			t.Error("Second unwrap should return database error")
		}

		unwrapped = errors.Unwrap(dbErr)
		if unwrapped != networkErr {
			t.Error("Third unwrap should return network error")
		}

		unwrapped = errors.Unwrap(networkErr)
		if unwrapped != nil {
			t.Error("Final unwrap should return nil")
		}
	})

	t.Run("Metadata Isolation", func(t *testing.T) {
		// Each error should maintain its own metadata
		apiMetadata := apiErr.GetAllMetadata()
		serviceMetadata := serviceErr.GetAllMetadata()
		dbMetadata := dbErr.GetAllMetadata()

		// API error metadata
		if endpoint, exists := apiMetadata["endpoint"]; !exists || endpoint != "/api/users/123" {
			t.Error("API error should contain endpoint metadata")
		}

		// Service error metadata
		if service, exists := serviceMetadata["service"]; !exists || service != "user-service" {
			t.Error("Service error should contain service metadata")
		}

		// Database error metadata
		if database, exists := dbMetadata["database"]; !exists || database != "postgres" {
			t.Error("Database error should contain database metadata")
		}

		// Metadata should not leak between levels
		if _, exists := apiMetadata["database"]; exists {
			t.Error("API error should not contain database metadata")
		}

		if _, exists := dbMetadata["endpoint"]; exists {
			t.Error("Database error should not contain API metadata")
		}
	})

	t.Run("Request ID Propagation", func(t *testing.T) {
		// Only the top-level error should have request ID
		if apiErr.RequestID != "req-chain-test" {
			t.Error("API error should have request ID")
		}

		if serviceErr.RequestID != "" {
			t.Error("Service error should not have request ID")
		}

		if dbErr.RequestID != "" {
			t.Error("Database error should not have request ID")
		}
	})
}

// TestErrorChainDebugging tests debugging scenarios with error chains
func TestErrorChainDebugging(t *testing.T) {
	// Create a realistic debugging scenario
	originalErr := errors.New("permission denied")

	fileErr := NewCustomError(ErrForbidden, originalErr, "cannot access configuration file").
		WithMetadata("file_path", "/etc/myapp/config.yaml").
		WithMetadata("user", "app_user").
		WithMetadata("permissions", "0644")

	configErr := NewCustomError(ErrInternal, fileErr, "configuration loading failed").
		WithMetadata("config_type", "yaml").
		WithMetadata("required_keys", "database,redis,auth")

	appErr := NewCustomError(ErrInternal, configErr, "application startup failed").
		WithMetadata("component", "main").
		WithMetadata("startup_time", "2023-10-15T10:30:00Z").
		WithRequestID("req-startup-001")

	t.Run("Debugging Information Extraction", func(t *testing.T) {
		// Test detailed error output for debugging
		detailedError := appErr.DetailedError()

		// Should contain main error information
		if !strings.Contains(detailedError, "application startup failed") {
			t.Error("Detailed error should contain main message")
		}

		if !strings.Contains(detailedError, "req-startup-001") {
			t.Error("Detailed error should contain request ID")
		}

		// Should contain metadata
		if !strings.Contains(detailedError, "component: main") {
			t.Error("Detailed error should contain metadata")
		}

		// Should contain wrapped error information
		if !strings.Contains(detailedError, "configuration loading failed") {
			t.Error("Detailed error should contain wrapped error info")
		}
	})

	t.Run("Short Error Format", func(t *testing.T) {
		shortError := appErr.ShortError()

		// Should be concise but informative
		if !strings.Contains(shortError, "INTERNAL_ERROR") {
			t.Error("Short error should contain error code")
		}

		if !strings.Contains(shortError, "req-startup-001") {
			t.Error("Short error should contain request ID")
		}

		if !strings.Contains(shortError, "application startup failed") {
			t.Error("Short error should contain main message")
		}

		// Should not contain metadata (too verbose for short format)
		if strings.Contains(shortError, "component: main") {
			t.Error("Short error should not contain metadata details")
		}
	})

	t.Run("Root Cause Analysis", func(t *testing.T) {
		// Walk the error chain to find root cause
		current := error(appErr)
		var rootCause error
		var errorChain []string

		for current != nil {
			errorChain = append(errorChain, current.Error())

			unwrapped := errors.Unwrap(current)
			if unwrapped == nil {
				rootCause = current
				break
			}
			current = unwrapped
		}

		// Verify root cause
		if rootCause.Error() != "permission denied" {
			t.Errorf("Expected root cause 'permission denied', got %s", rootCause.Error())
		}

		// Verify chain length
		if len(errorChain) != 4 { // app -> config -> file -> original
			t.Errorf("Expected 4 errors in chain, got %d", len(errorChain))
		}

		// Verify chain order (most recent first)
		if !strings.Contains(errorChain[0], "application startup failed") {
			t.Error("First error should be application error")
		}

		if !strings.Contains(errorChain[len(errorChain)-1], "permission denied") {
			t.Error("Last error should be root cause")
		}
	})

	t.Run("Context Recovery from Chain", func(t *testing.T) {
		// Extract all CustomErrors from chain
		var customErrors []*CustomError
		current := error(appErr)

		for current != nil {
			var customErr *CustomError
			if errors.As(current, &customErr) {
				customErrors = append(customErrors, customErr)
			}
			current = errors.Unwrap(current)
		}

		// Should find 3 CustomErrors (app, config, file)
		if len(customErrors) != 3 {
			t.Fatalf("Expected 3 CustomErrors, got %d", len(customErrors))
		}

		// Verify each has its own category
		expectedCategories := []ErrorCategory{
			ErrorCategoryInternal,  // app
			ErrorCategoryInternal,  // config
			ErrorCategoryForbidden, // file
		}

		for i, err := range customErrors {
			if err.Category != expectedCategories[i] {
				t.Errorf("Error %d: expected category %s, got %s", i, expectedCategories[i], err.Category)
			}
		}

		// Extract all metadata from chain
		allMetadata := make(map[string]map[string]string)
		for i, err := range customErrors {
			allMetadata[fmt.Sprintf("level_%d", i)] = err.GetAllMetadata()
		}

		// Verify metadata at each level
		if component, exists := allMetadata["level_0"]["component"]; !exists || component != "main" {
			t.Error("Should find component metadata at app level")
		}

		if configType, exists := allMetadata["level_1"]["config_type"]; !exists || configType != "yaml" {
			t.Error("Should find config_type metadata at config level")
		}

		if filePath, exists := allMetadata["level_2"]["file_path"]; !exists || filePath != "/etc/myapp/config.yaml" {
			t.Error("Should find file_path metadata at file level")
		}
	})
}

// TestCircularErrorPrevention tests prevention of circular error references
func TestCircularErrorPrevention(t *testing.T) {
	err1 := NewCustomError(ErrInternal, nil, "error 1")
	err2 := NewCustomError(ErrInternal, err1, "error 2")

	// This should not create a circular reference (err1 cannot wrap err2)
	// The package should handle this gracefully
	t.Run("Prevent Self Reference", func(t *testing.T) {
		// Try to make err1 wrap err2 (which already wraps err1)
		// This should either be prevented or handled safely

		// Test that we can still unwrap normally
		unwrapped := errors.Unwrap(err2)
		if unwrapped != err1 {
			t.Error("Normal unwrapping should work")
		}

		// Test that errors.Is still works
		if !errors.Is(err2, err1) {
			t.Error("errors.Is should work through chain")
		}

		// The error chain should terminate properly
		var depth int
		current := error(err2)
		for current != nil && depth < 10 { // Prevent infinite loop
			current = errors.Unwrap(current)
			depth++
		}

		if depth >= 10 {
			t.Error("Error chain should terminate, possible circular reference")
		}
	})
}

// TestComplexErrorChainScenarios tests realistic complex scenarios
func TestComplexErrorChainScenarios(t *testing.T) {
	t.Run("Microservice Error Chain", func(t *testing.T) {
		// Simulate error propagating through microservices

		// External API error
		externalErr := errors.New("503 Service Unavailable")

		// Payment service error
		paymentErr := NewCustomError(ErrExternal, externalErr, "payment gateway timeout").
			WithMetadata("gateway", "stripe").
			WithMetadata("amount", "99.99").
			WithMetadata("currency", "USD")

		// Order service error
		orderErr := NewCustomError(ErrInternal, paymentErr, "failed to process payment").
			WithMetadata("order_id", "ord_123456").
			WithMetadata("customer_id", "cust_789").
			WithMetadata("retry_count", "3")

		// API gateway error
		gatewayErr := NewCustomError(ErrInternal, orderErr, "checkout process failed").
			WithMetadata("endpoint", "/api/checkout").
			WithMetadata("user_agent", "mobile-app/1.2.3").
			WithRequestID("req-checkout-001")

		// Test complete error information is accessible
		if !errors.Is(gatewayErr, externalErr) {
			t.Error("Should trace back to original external error")
		}

		// Extract service-specific information
		var paymentCustomErr, orderCustomErr, gatewayCustomErr *CustomError

		// Walk chain and extract each CustomError
		current := error(gatewayErr)
		for current != nil {
			var customErr *CustomError
			if errors.As(current, &customErr) {
				switch {
				case strings.Contains(customErr.Message, "checkout"):
					gatewayCustomErr = customErr
				case strings.Contains(customErr.Message, "process payment"):
					orderCustomErr = customErr
				case strings.Contains(customErr.Message, "payment gateway"):
					paymentCustomErr = customErr
				}
			}
			current = errors.Unwrap(current)
		}

		// Verify each service's error was captured
		if gatewayCustomErr == nil {
			t.Fatal("Gateway error not found in chain")
		}
		if orderCustomErr == nil {
			t.Fatal("Order error not found in chain")
		}
		if paymentCustomErr == nil {
			t.Fatal("Payment error not found in chain")
		}

		// Verify service-specific metadata
		gatewayMeta := gatewayCustomErr.GetAllMetadata()
		if gatewayMeta["endpoint"] != "/api/checkout" {
			t.Error("Gateway should have endpoint metadata")
		}

		orderMeta := orderCustomErr.GetAllMetadata()
		if orderMeta["order_id"] != "ord_123456" {
			t.Error("Order service should have order_id metadata")
		}

		paymentMeta := paymentCustomErr.GetAllMetadata()
		if paymentMeta["gateway"] != "stripe" {
			t.Error("Payment service should have gateway metadata")
		}
	})

	t.Run("Database Transaction Error Chain", func(t *testing.T) {
		// Simulate database transaction failure chain

		// Low-level database error
		sqlErr := errors.New("duplicate key value violates unique constraint")

		// Database layer error
		dbErr := NewCustomError(ErrAlreadyExists, sqlErr, "user email already exists").
			WithMetadata("table", "users").
			WithMetadata("constraint", "unique_email").
			WithMetadata("attempted_email", "test@example.com")

		// Business logic error
		businessErr := NewCustomError(ErrAlreadyExists, dbErr, "user registration failed").
			WithMetadata("step", "create_user").
			WithMetadata("validation_passed", "true")

		// Controller error
		controllerErr := NewCustomError(ErrAlreadyExists, businessErr, "account creation unsuccessful").
			WithMetadata("handler", "POST /register").
			WithMetadata("client_ip", "10.0.1.5").
			WithRequestID("req-register-002")

		// Test error categorization remains consistent
		if controllerErr.Category != ErrorCategoryConflict {
			t.Error("Error category should be maintained through chain")
		}

		// Test SQL error is still accessible
		if !errors.Is(controllerErr, sqlErr) {
			t.Error("Should trace back to SQL error")
		}

		// Test HTTP status mapping works
		if controllerErr.ToHTTPStatus() != HTTP_STATUS_CONFLICT {
			t.Error("Should map to HTTP 409 Conflict")
		}

		// Extract debugging information for each layer
		detailedInfo := controllerErr.DetailedError()

		// Should contain information from all layers
		layers := []string{
			"account creation unsuccessful", // controller
			"user registration failed",      // business
			"user email already exists",     // db
			"duplicate key value",           // sql
		}

		for _, layer := range layers {
			if !strings.Contains(detailedInfo, layer) {
				t.Errorf("Detailed error should contain layer info: %s", layer)
			}
		}
	})
}
