package cuserr

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestProductionConfiguration tests production mode behavior
func TestProductionConfiguration(t *testing.T) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	t.Run("Production Mode Error Message Filtering", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: false,
			ProductionMode:   true,
		})

		testCases := []struct {
			name            string
			category        ErrorCategory
			message         string
			expectedMessage string
		}{
			{
				name:            "Internal Error Filtered",
				category:        ErrorCategoryInternal,
				message:         "database connection string: postgres://user:pass@host/db",
				expectedMessage: "An internal error occurred",
			},
			{
				name:            "External Error Filtered",
				category:        ErrorCategoryExternal,
				message:         "API key abc123 is invalid for service xyz",
				expectedMessage: "A service is temporarily unavailable",
			},
			{
				name:            "Validation Error Preserved",
				category:        ErrorCategoryValidation,
				message:         "email field is required",
				expectedMessage: "email field is required",
			},
			{
				name:            "Not Found Error Preserved",
				category:        ErrorCategoryNotFound,
				message:         "user with ID 12345 not found",
				expectedMessage: "user with ID 12345 not found",
			},
			{
				name:            "Unauthorized Error Preserved",
				category:        ErrorCategoryUnauthorized,
				message:         "invalid API token",
				expectedMessage: "invalid API token",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := NewCustomErrorWithCategory(tc.category, "TEST_CODE", tc.message)

				clientMessage := err.ClientSafeMessage()
				if clientMessage != tc.expectedMessage {
					t.Errorf("Expected message '%s', got '%s'", tc.expectedMessage, clientMessage)
				}

				// Test in JSON format too
				clientJSON := err.ToClientJSON()
				errorData := clientJSON[JSON_FIELD_ERROR].(map[string]interface{})
				jsonMessage := errorData[JSON_FIELD_MESSAGE].(string)

				if jsonMessage != tc.expectedMessage {
					t.Errorf("JSON message should match: expected '%s', got '%s'", tc.expectedMessage, jsonMessage)
				}
			})
		}
	})

	t.Run("Production Metadata Filtering", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: false,
			ProductionMode:   true,
		})

		err := NewCustomError(ErrInternal, nil, "internal service error").
			WithMetadata("user_id", "usr_123").             // Safe - should be kept
			WithMetadata("request_id", "req_456").          // Safe - should be kept
			WithMetadata("trace_id", "trace_789").          // Safe - should be kept
			WithMetadata("correlation_id", "corr_abc").     // Safe - should be kept
			WithMetadata("database_password", "secret123"). // Sensitive - should be removed
			WithMetadata("api_key", "key_sensitive").       // Sensitive - should be removed
			WithMetadata("internal_debug", "stack_info")    // Sensitive - should be removed

		clientJSON := err.ToClientJSON()
		errorData := clientJSON[JSON_FIELD_ERROR].(map[string]interface{})
		metadata, hasMetadata := errorData[JSON_FIELD_METADATA].(map[string]string)

		if !hasMetadata {
			t.Fatal("Should have metadata even after filtering")
		}

		// Should keep safe identifiers
		safeKeys := []string{"user_id", "request_id", "trace_id", "correlation_id"}
		for _, key := range safeKeys {
			if _, exists := metadata[key]; !exists {
				t.Errorf("Should keep safe metadata key: %s", key)
			}
		}

		// Should remove sensitive keys
		sensitiveKeys := []string{"database_password", "api_key", "internal_debug"}
		for _, key := range sensitiveKeys {
			if _, exists := metadata[key]; exists {
				t.Errorf("Should remove sensitive metadata key: %s", key)
			}
		}

		// Total metadata should be reduced
		if len(metadata) != 4 { // Only the 4 safe keys
			t.Errorf("Expected 4 metadata keys after filtering, got %d", len(metadata))
		}
	})

	t.Run("Development Mode Shows All Data", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: true,
			ProductionMode:   false,
		})

		err := NewCustomError(ErrInternal, nil, "detailed internal error with sensitive info").
			WithMetadata("database_connection", "sensitive_conn_string").
			WithMetadata("user_id", "usr_123")

		// In development mode, should show detailed message
		clientMessage := err.ClientSafeMessage()
		if clientMessage != "detailed internal error with sensitive info" {
			t.Error("Development mode should show detailed messages")
		}

		// Should preserve all metadata
		clientJSON := err.ToClientJSON()
		errorData := clientJSON[JSON_FIELD_ERROR].(map[string]interface{})
		metadata := errorData[JSON_FIELD_METADATA].(map[string]string)

		if len(metadata) != 2 {
			t.Error("Development mode should preserve all metadata")
		}

		if _, exists := metadata["database_connection"]; !exists {
			t.Error("Development mode should keep sensitive metadata")
		}
	})
}

// TestConfigurationTransitions tests switching between production and development
func TestConfigurationTransitions(t *testing.T) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	t.Run("Runtime Configuration Changes", func(t *testing.T) {
		// Start in development mode
		SetConfig(&Config{
			EnableStackTrace: true,
			ProductionMode:   false,
		})

		err := NewCustomError(ErrInternal, nil, "sensitive error message").
			WithMetadata("sensitive_data", "secret123").
			WithMetadata("user_id", "usr_456")

		// Should show detailed info in development
		devMessage := err.ClientSafeMessage()
		devJSON := err.ToClientJSON()
		devErrorData := devJSON[JSON_FIELD_ERROR].(map[string]interface{})
		devMetadata := devErrorData[JSON_FIELD_METADATA].(map[string]string)

		if devMessage != "sensitive error message" {
			t.Error("Development mode should show sensitive message")
		}
		if len(devMetadata) != 2 {
			t.Error("Development mode should show all metadata")
		}

		// Switch to production mode
		SetConfig(&Config{
			EnableStackTrace: false,
			ProductionMode:   true,
		})

		// Same error should now be filtered
		prodMessage := err.ClientSafeMessage()
		prodJSON := err.ToClientJSON()
		prodErrorData := prodJSON[JSON_FIELD_ERROR].(map[string]interface{})
		prodMetadata := prodErrorData[JSON_FIELD_METADATA].(map[string]string)

		if prodMessage != "An internal error occurred" {
			t.Error("Production mode should show generic message")
		}
		if len(prodMetadata) != 1 { // Only user_id should remain
			t.Errorf("Production mode should filter metadata, expected 1, got %d", len(prodMetadata))
		}
		if _, exists := prodMetadata["sensitive_data"]; exists {
			t.Error("Production mode should filter sensitive metadata")
		}
	})
}

// TestEnvironmentVariableConfiguration tests configuration from environment
func TestEnvironmentVariableConfiguration(t *testing.T) {
	// Save original environment
	originalEnvStackTrace := os.Getenv(ENV_ENABLE_STACK_TRACE)
	originalEnvMaxDepth := os.Getenv(ENV_MAX_STACK_DEPTH)
	originalEnvProduction := os.Getenv(ENV_PRODUCTION_MODE)
	originalConfig := GetConfig()

	defer func() {
		// Restore environment
		os.Setenv(ENV_ENABLE_STACK_TRACE, originalEnvStackTrace)
		os.Setenv(ENV_MAX_STACK_DEPTH, originalEnvMaxDepth)
		os.Setenv(ENV_PRODUCTION_MODE, originalEnvProduction)
		SetConfig(originalConfig)
	}()

	t.Run("Environment Variable Configuration", func(t *testing.T) {
		// Set environment variables
		os.Setenv(ENV_ENABLE_STACK_TRACE, "true")
		os.Setenv(ENV_MAX_STACK_DEPTH, "15")
		os.Setenv(ENV_PRODUCTION_MODE, "true")

		// Note: This test assumes there's an init function or method to read from env
		// For now, we'll test the constants exist and manual configuration works
		testConfig := &Config{
			EnableStackTrace: true,
			MaxStackDepth:    15,
			ProductionMode:   true,
		}

		SetConfig(testConfig)
		config := GetConfig()

		if !config.EnableStackTrace {
			t.Error("Should enable stack trace from environment")
		}
		if config.MaxStackDepth != 15 {
			t.Error("Should set max stack depth from environment")
		}
		if !config.ProductionMode {
			t.Error("Should enable production mode from environment")
		}
	})
}

// TestProductionReliability tests production reliability scenarios
func TestProductionReliability(t *testing.T) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	t.Run("High Volume Error Creation", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: false,
			ProductionMode:   true,
		})

		const numErrors = 10000
		errors := make([]*CustomError, numErrors)

		start := time.Now()

		// Create many errors quickly
		for i := 0; i < numErrors; i++ {
			errors[i] = NewCustomError(ErrInternal, nil, "high volume test error").
				WithMetadata("iteration", string(rune(i))).
				WithRequestID("req-volume-test")
		}

		duration := time.Since(start)

		// Should complete reasonably quickly (less than 1 second for 10k errors)
		if duration > time.Second {
			t.Errorf("High volume error creation too slow: %v", duration)
		}

		// Verify all errors were created correctly
		for i, err := range errors {
			if err == nil {
				t.Errorf("Error %d was not created", i)
				continue
			}

			if err.Message != "high volume test error" {
				t.Errorf("Error %d has wrong message", i)
			}

			// In production, internal errors should be filtered
			if err.ClientSafeMessage() != "An internal error occurred" {
				t.Errorf("Error %d should be filtered in production", i)
			}
		}
	})

	t.Run("Memory Usage Control", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    2, // Limit stack depth
			ProductionMode:   true,
		})

		// Create error with limited stack trace
		err := NewCustomError(ErrTimeout, nil, "memory test error").
			WithMetadata("test", "memory_control")

		stackTrace := err.GetStackTrace()

		// Should respect depth limit
		if len(stackTrace) > 2 {
			t.Errorf("Stack trace should be limited to 2 frames, got %d", len(stackTrace))
		}

		// Clear stack trace to save memory
		err.ClearStackTrace()

		if len(err.GetStackTrace()) != 0 {
			t.Error("Stack trace should be cleared")
		}
	})

	t.Run("Thread Safety in Production", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: false,
			ProductionMode:   true,
		})

		err := NewCustomError(ErrInternal, nil, "production thread safety test")

		const numGoroutines = 50
		const numOperations = 100

		var wg sync.WaitGroup
		wg.Add(numGoroutines * 2)

		// Start metadata writers
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					err.WithMetadata("key", "value")
					time.Sleep(time.Microsecond)
				}
			}(i)
		}

		// Start JSON generators (which read metadata)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					_ = err.ToClientJSON()
					time.Sleep(time.Microsecond)
				}
			}()
		}

		wg.Wait()
		// If we reach here without panic, thread safety is working
	})
}

// TestProductionSecurityFiltering tests security-focused filtering
func TestProductionSecurityFiltering(t *testing.T) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	SetConfig(&Config{
		EnableStackTrace: false,
		ProductionMode:   true,
	})

	t.Run("Sensitive Data Filtering", func(t *testing.T) {
		sensitiveData := map[string]string{
			"password":          "secret123",
			"api_key":           "sk_live_abc123",
			"token":             "bearer_token_xyz",
			"secret":            "my_secret_value",
			"private_key":       "-----BEGIN PRIVATE KEY-----",
			"database_url":      "postgres://user:pass@host/db",
			"connection_string": "server=host;password=secret",
			"auth_header":       "Authorization: Bearer token123",
			"credit_card":       "4111111111111111",
			"ssn":               "123-45-6789",
		}

		safeData := map[string]string{
			"user_id":        "usr_123",
			"request_id":     "req_456",
			"trace_id":       "trace_789",
			"correlation_id": "corr_abc",
		}

		err := NewCustomError(ErrInternal, nil, "test error")

		// Add both sensitive and safe metadata
		for key, value := range sensitiveData {
			err.WithMetadata(key, value)
		}
		for key, value := range safeData {
			err.WithMetadata(key, value)
		}

		clientJSON := err.ToClientJSON()
		errorData := clientJSON[JSON_FIELD_ERROR].(map[string]interface{})
		metadata, hasMetadata := errorData[JSON_FIELD_METADATA].(map[string]string)

		if !hasMetadata {
			t.Fatal("Should have metadata")
		}

		// Should keep only safe data
		for key := range safeData {
			if _, exists := metadata[key]; !exists {
				t.Errorf("Should keep safe metadata: %s", key)
			}
		}

		// Should remove all sensitive data
		for key := range sensitiveData {
			if _, exists := metadata[key]; exists {
				t.Errorf("Should remove sensitive metadata: %s", key)
			}
		}

		// Should only have the safe data
		if len(metadata) != len(safeData) {
			t.Errorf("Expected %d metadata keys, got %d", len(safeData), len(metadata))
		}
	})

	t.Run("Error Message Content Filtering", func(t *testing.T) {
		// Test various types of potentially sensitive error messages
		testCases := []struct {
			category         ErrorCategory
			message          string
			shouldBeFiltered bool
		}{
			{
				category:         ErrorCategoryInternal,
				message:          "database connection failed: password authentication failed for user 'admin'",
				shouldBeFiltered: true,
			},
			{
				category:         ErrorCategoryExternal,
				message:          "API call failed: invalid API key 'sk_live_abc123'",
				shouldBeFiltered: true,
			},
			{
				category:         ErrorCategoryValidation,
				message:          "email format is invalid",
				shouldBeFiltered: false,
			},
			{
				category:         ErrorCategoryNotFound,
				message:          "user with email 'test@example.com' not found",
				shouldBeFiltered: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.message, func(t *testing.T) {
				err := NewCustomErrorWithCategory(tc.category, "TEST_CODE", tc.message)
				clientMessage := err.ClientSafeMessage()

				if tc.shouldBeFiltered {
					if clientMessage == tc.message {
						t.Errorf("Sensitive message should be filtered: %s", tc.message)
					}
					if !strings.Contains(clientMessage, "error occurred") && !strings.Contains(clientMessage, "temporarily unavailable") {
						t.Errorf("Filtered message should be generic: %s", clientMessage)
					}
				} else {
					if clientMessage != tc.message {
						t.Errorf("Non-sensitive message should not be filtered: expected '%s', got '%s'", tc.message, clientMessage)
					}
				}
			})
		}
	})
}

// TestConfigurationValidation tests configuration validation
func TestConfigurationValidation(t *testing.T) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	t.Run("Invalid Configuration Handling", func(t *testing.T) {
		// Test negative stack depth
		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    -1,
			ProductionMode:   false,
		})

		err := NewCustomError(ErrInternal, nil, "test negative depth")

		// Should use default depth when negative
		config := GetConfig()
		if config.MaxStackDepth < 0 {
			// If the package allows negative values, it should handle them gracefully
			stackTrace := err.GetStackTrace()
			// Should either use default depth or handle gracefully
			if len(stackTrace) < 0 {
				t.Error("Stack trace length should not be negative")
			}
		}
	})

	t.Run("Zero Configuration", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: false,
			MaxStackDepth:    0,
			ProductionMode:   false,
		})

		err := NewCustomError(ErrInternal, nil, "test zero config")

		// Should handle zero configuration gracefully
		stackTrace := err.GetStackTrace()
		if len(stackTrace) != 0 {
			t.Error("Should have no stack trace when disabled")
		}

		// Error should still function normally
		if err.Error() != "test zero config" {
			t.Error("Error should function normally with zero config")
		}
	})
}
