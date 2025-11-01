package cuserr

import (
	"strings"
	"testing"
)

// TestGetPublicMessage tests the GetPublicMessage method
func TestGetPublicMessage(t *testing.T) {
	tests := []struct {
		name           string
		err            *CustomError
		productionMode bool
		expected       string
	}{
		{
			name: "Uses PublicMessage when set",
			err: &CustomError{
				Message:       "Internal database error",
				PublicMessage: "Service temporarily unavailable",
				Category:      ErrorCategoryInternal,
			},
			productionMode: false,
			expected:       "Service temporarily unavailable",
		},
		{
			name: "Falls back to Message when PublicMessage not set (non-production)",
			err: &CustomError{
				Message:  "Invalid email format",
				Category: ErrorCategoryValidation,
			},
			productionMode: false,
			expected:       "Invalid email format",
		},
		{
			name: "Uses generic message for internal errors in production",
			err: &CustomError{
				Message:  "Database connection failed: timeout",
				Category: ErrorCategoryInternal,
			},
			productionMode: true,
			expected:       "An internal error occurred",
		},
		{
			name: "Uses generic message for external errors in production",
			err: &CustomError{
				Message:  "Payment API returned 500",
				Category: ErrorCategoryExternal,
			},
			productionMode: true,
			expected:       "A service is temporarily unavailable",
		},
		{
			name: "Uses actual message for validation in production",
			err: &CustomError{
				Message:  "Email format is invalid",
				Category: ErrorCategoryValidation,
			},
			productionMode: true,
			expected:       "Email format is invalid",
		},
		{
			name: "PublicMessage overrides production mode",
			err: &CustomError{
				Message:       "Database error: connection timeout",
				PublicMessage: "Please try again later",
				Category:      ErrorCategoryInternal,
			},
			productionMode: true,
			expected:       "Please try again later",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set production mode
			originalConfig := GetConfig()
			defer SetConfig(originalConfig)

			SetConfig(&Config{
				EnableStackTrace: true,
				MaxStackDepth:    DEFAULT_STACK_DEPTH,
				ProductionMode:   tt.productionMode,
			})

			result := tt.err.GetPublicMessage()
			if result != tt.expected {
				t.Errorf("GetPublicMessage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestClientSafeMessageBackwardCompat tests that ClientSafeMessage() delegates to GetPublicMessage()
func TestClientSafeMessageBackwardCompat(t *testing.T) {
	err := &CustomError{
		Message:       "Internal error",
		PublicMessage: "Please try again",
		Category:      ErrorCategoryInternal,
	}

	clientSafe := err.ToPublicJSON()[JSON_FIELD_ERROR].(map[string]interface{})[JSON_FIELD_MESSAGE].(string)
	publicMsg := err.GetPublicMessage()

	if clientSafe != publicMsg {
		t.Errorf("ClientSafeMessage() = %q, GetPublicMessage() = %q, should be equal", clientSafe, publicMsg)
	}
}

// TestToPublicJSON tests the ToPublicJSON method
func TestToPublicJSON(t *testing.T) {
	tests := []struct {
		name           string
		err            *CustomError
		productionMode bool
		checkFields    map[string]bool // fields that should exist
	}{
		{
			name: "Includes all universal fields",
			err: &CustomError{
				Code:          "ERR_TEST",
				Message:       "Internal error",
				PublicMessage: "Service unavailable",
				Category:      ErrorCategoryInternal,
				HelpURL:       "https://docs.example.com/errors#test",
				Suggestions:   []string{"Try again later", "Contact support"},
				MessageKey:    "errors.test",
				OperationID:   "op-123",
				TraceID:       "trace-456",
				SpanID:        "span-789",
			},
			productionMode: false,
			checkFields: map[string]bool{
				"public_message": true,
				"help_url":       true,
				"suggestions":    true,
				"message_key":    true,
				"operation_id":   true,
				"trace_id":       true,
				"span_id":        true,
			},
		},
		{
			name: "Filters metadata in production mode",
			err: func() *CustomError {
				err := &CustomError{
					Code:     "ERR_TEST",
					Message:  "Test error",
					Category: ErrorCategoryInternal,
					metadata: make(map[string]string),
				}
				err.metadata["user_id"] = "user-123"      // safe - should be kept
				err.metadata["trace_id"] = "trace-456"    // safe - should be kept
				err.metadata["password"] = "secret"       // sensitive - should be filtered
				err.metadata["credit_card"] = "1234-5678" // sensitive - should be filtered
				return err
			}(),
			productionMode: true,
			checkFields: map[string]bool{
				"metadata": true, // should have metadata but filtered
			},
		},
		{
			name: "Uses OperationID",
			err: &CustomError{
				Code:        "ERR_TEST",
				Message:     "Test error",
				Category:    ErrorCategoryValidation,
				OperationID: "op-123",
			},
			productionMode: false,
			checkFields: map[string]bool{
				"operation_id": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set production mode
			originalConfig := GetConfig()
			defer SetConfig(originalConfig)

			SetConfig(&Config{
				EnableStackTrace: true,
				MaxStackDepth:    DEFAULT_STACK_DEPTH,
				ProductionMode:   tt.productionMode,
			})

			result := tt.err.ToPublicJSON()
			errorData, ok := result[JSON_FIELD_ERROR].(map[string]interface{})
			if !ok {
				t.Fatal("ToPublicJSON() should return error field")
			}

			// Check expected fields
			for field, shouldExist := range tt.checkFields {
				_, exists := errorData[field]
				if shouldExist && !exists {
					t.Errorf("ToPublicJSON() missing expected field: %s", field)
				}
			}

			// Verify metadata filtering in production
			if tt.productionMode && tt.name == "Filters metadata in production mode" {
				metadata, ok := errorData["metadata"].(map[string]string)
				if ok {
					if _, hasSensitive := metadata["password"]; hasSensitive {
						t.Error("ToPublicJSON() should filter sensitive metadata in production")
					}
					if _, hasSafe := metadata["user_id"]; !hasSafe {
						t.Error("ToPublicJSON() should keep safe metadata in production")
					}
				}
			}
		})
	}
}

// TestToClientJSONBackwardCompat tests that ToClientJSON() delegates to ToPublicJSON()
func TestToClientJSONBackwardCompat(t *testing.T) {
	err := &CustomError{
		Code:          "ERR_TEST",
		Message:       "Test error",
		PublicMessage: "User-friendly message",
		Category:      ErrorCategoryValidation,
	}

	clientJSON := err.ToPublicJSON()
	publicJSON := err.ToPublicJSON()

	// Both should return the same structure
	clientErr := clientJSON[JSON_FIELD_ERROR].(map[string]interface{})
	publicErr := publicJSON[JSON_FIELD_ERROR].(map[string]interface{})

	if clientErr["code"] != publicErr["code"] {
		t.Error("ToClientJSON() and ToPublicJSON() should return the same structure")
	}
}

// TestToExitCode tests CLI exit code mapping
func TestToExitCode(t *testing.T) {
	tests := []struct {
		name     string
		category ErrorCategory
		expected int
	}{
		{"Validation error", ErrorCategoryValidation, 1},
		{"Not found error", ErrorCategoryNotFound, 1},
		{"Unauthorized error", ErrorCategoryUnauthorized, 126},
		{"Forbidden error", ErrorCategoryForbidden, 126},
		{"Timeout error", ErrorCategoryTimeout, 124},
		{"Conflict error", ErrorCategoryConflict, 1},
		{"Rate limit error", ErrorCategoryRateLimit, 1},
		{"External error", ErrorCategoryExternal, 69},
		{"Internal error", ErrorCategoryInternal, 70},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &CustomError{
				Code:     "ERR_TEST",
				Message:  "test error",
				Category: tt.category,
			}

			exitCode := err.ToExitCode()
			if exitCode != tt.expected {
				t.Errorf("ToExitCode() = %d, want %d for category %s", exitCode, tt.expected, tt.category)
			}

			// Also test the category function directly
			categoryCode := CategoryToExitCode(tt.category)
			if categoryCode != tt.expected {
				t.Errorf("CategoryToExitCode() = %d, want %d for category %s", categoryCode, tt.expected, tt.category)
			}
		})
	}
}

// TestToText tests plain text output
func TestToText(t *testing.T) {
	tests := []struct {
		name        string
		err         *CustomError
		wantContain []string
	}{
		{
			name: "Basic error with code",
			err: &CustomError{
				Code:          "ERR_VALIDATION",
				Message:       "Invalid input",
				PublicMessage: "Please check your input",
				Category:      ErrorCategoryValidation,
			},
			wantContain: []string{"Error:", "Please check your input", "Code: ERR_VALIDATION"},
		},
		{
			name: "Error with OperationID",
			err: &CustomError{
				Code:        "ERR_TEST",
				Message:     "Test error",
				Category:    ErrorCategoryInternal,
				OperationID: "op-12345",
			},
			wantContain: []string{"Error:", "Test error", "Operation: op-12345"},
		},
		{
			name: "Error with RequestID fallback",
			err: &CustomError{
				Code:      "ERR_TEST",
				Message:   "Test error",
				Category:  ErrorCategoryInternal,
				OperationID: "req-67890",
			},
			wantContain: []string{"Error:", "Test error", "Operation: req-67890"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.err.ToText()
			for _, want := range tt.wantContain {
				if !strings.Contains(output, want) {
					t.Errorf("ToText() output should contain %q, got:\n%s", want, output)
				}
			}
		})
	}
}

// TestToCLIOutput tests colored CLI output
func TestToCLIOutput(t *testing.T) {
	err := &CustomError{
		Code:          "ERR_VALIDATION",
		Message:       "Invalid input",
		PublicMessage: "Please check your input",
		Category:      ErrorCategoryValidation,
		HelpURL:       "https://docs.example.com/errors",
		Suggestions:   []string{"Check spelling", "Verify format"},
		OperationID:   "op-123",
	}

	t.Run("Colored output", func(t *testing.T) {
		output := err.ToCLIOutput(true)

		// Check for ANSI color codes
		if !strings.Contains(output, "\033[") {
			t.Error("ToCLIOutput(true) should contain ANSI color codes")
		}

		// Check content
		wantContain := []string{"Error", "validation", "Please check your input", "ERR_VALIDATION", "https://docs.example.com/errors", "Check spelling", "Verify format", "op-123"}
		for _, want := range wantContain {
			if !strings.Contains(output, want) {
				t.Errorf("ToCLIOutput() should contain %q, got:\n%s", want, output)
			}
		}
	})

	t.Run("Non-colored output", func(t *testing.T) {
		output := err.ToCLIOutput(false)

		// Should not contain color codes
		if strings.Contains(output, "\033[") {
			t.Error("ToCLIOutput(false) should not contain ANSI color codes")
		}

		// Should still contain content
		wantContain := []string{"Error", "validation", "Please check your input"}
		for _, want := range wantContain {
			if !strings.Contains(output, want) {
				t.Errorf("ToCLIOutput() should contain %q", want)
			}
		}
	})
}

// TestToSyslogSeverity tests syslog severity mapping
func TestToSyslogSeverity(t *testing.T) {
	tests := []struct {
		name     string
		category ErrorCategory
		expected int
	}{
		{"Validation error", ErrorCategoryValidation, 6},     // Informational
		{"Unauthorized error", ErrorCategoryUnauthorized, 5}, // Notice
		{"Forbidden error", ErrorCategoryForbidden, 5},       // Notice
		{"Not found error", ErrorCategoryNotFound, 4},        // Warning
		{"Timeout error", ErrorCategoryTimeout, 4},           // Warning
		{"Conflict error", ErrorCategoryConflict, 4},         // Warning
		{"Rate limit error", ErrorCategoryRateLimit, 5},      // Notice
		{"External error", ErrorCategoryExternal, 2},         // Critical
		{"Internal error", ErrorCategoryInternal, 3},         // Error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &CustomError{
				Code:     "ERR_TEST",
				Message:  "test error",
				Category: tt.category,
			}

			severity := err.ToSyslogSeverity()
			if severity != tt.expected {
				t.Errorf("ToSyslogSeverity() = %d, want %d for category %s", severity, tt.expected, tt.category)
			}

			// Also test the category function directly
			categorySeverity := CategoryToSyslogSeverity(tt.category)
			if categorySeverity != tt.expected {
				t.Errorf("CategoryToSyslogSeverity() = %d, want %d for category %s", categorySeverity, tt.expected, tt.category)
			}
		})
	}
}

// TestToGRPCCode tests gRPC code mapping
func TestToGRPCCode(t *testing.T) {
	tests := []struct {
		name     string
		category ErrorCategory
		expected int
	}{
		{"Validation error", ErrorCategoryValidation, 3},      // InvalidArgument
		{"Unauthorized error", ErrorCategoryUnauthorized, 16}, // Unauthenticated
		{"Forbidden error", ErrorCategoryForbidden, 7},        // PermissionDenied
		{"Not found error", ErrorCategoryNotFound, 5},         // NotFound
		{"Timeout error", ErrorCategoryTimeout, 4},            // DeadlineExceeded
		{"Conflict error", ErrorCategoryConflict, 6},          // AlreadyExists
		{"Rate limit error", ErrorCategoryRateLimit, 8},       // ResourceExhausted
		{"External error", ErrorCategoryExternal, 14},         // Unavailable
		{"Internal error", ErrorCategoryInternal, 13},         // Internal
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &CustomError{
				Code:     "ERR_TEST",
				Message:  "test error",
				Category: tt.category,
			}

			grpcCode := err.ToGRPCCode()
			if grpcCode != tt.expected {
				t.Errorf("ToGRPCCode() = %d, want %d for category %s", grpcCode, tt.expected, tt.category)
			}

			// Also test the category function directly
			categoryCode := CategoryToGRPCCode(tt.category)
			if categoryCode != tt.expected {
				t.Errorf("CategoryToGRPCCode() = %d, want %d for category %s", categoryCode, tt.expected, tt.category)
			}
		})
	}
}

// TestToGRPCDetails tests gRPC details generation
func TestToGRPCDetails(t *testing.T) {
	tests := []struct {
		name        string
		err         *CustomError
		checkFields []string
	}{
		{
			name: "Complete error with all fields",
			err: &CustomError{
				Code:          "ERR_TEST",
				Message:       "Internal error",
				PublicMessage: "Service unavailable",
				Category:      ErrorCategoryInternal,
				HelpURL:       "https://docs.example.com/errors",
				Suggestions:   []string{"Try again", "Contact support"},
				OperationID:   "op-123",
				TraceID:       "trace-456",
				SpanID:        "span-789",
			},
			checkFields: []string{"code", "message", "operation_id", "trace_id", "span_id", "help_url", "suggestions"},
		},
		{
			name: "Error with metadata",
			err: func() *CustomError {
				err := &CustomError{
					Code:     "ERR_TEST",
					Message:  "Test error",
					Category: ErrorCategoryValidation,
					metadata: make(map[string]string),
				}
				err.metadata["field"] = "email"
				err.metadata["value"] = "invalid@"
				return err
			}(),
			checkFields: []string{"code", "message", "metadata"},
		},
		{
			name: "Error with RequestID fallback",
			err: &CustomError{
				Code:      "ERR_TEST",
				Message:   "Test error",
				Category:  ErrorCategoryInternal,
				OperationID: "req-123",
			},
			checkFields: []string{"code", "message", "operation_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			details := tt.err.ToGRPCDetails()

			// Check expected fields
			for _, field := range tt.checkFields {
				if _, exists := details[field]; !exists {
					t.Errorf("ToGRPCDetails() missing expected field: %s", field)
				}
			}

			// Verify message uses PublicMessage if available
			if tt.err.PublicMessage != "" {
				if msg, ok := details["message"].(string); ok {
					if msg != tt.err.PublicMessage {
						t.Errorf("ToGRPCDetails() message = %q, want %q", msg, tt.err.PublicMessage)
					}
				}
			}

			// Verify OperationID fallback to RequestID
			if tt.name == "Error with RequestID fallback" {
				if opID, ok := details["operation_id"].(string); ok {
					if opID != tt.err.OperationID {
						t.Errorf("ToGRPCDetails() should use RequestID as operation_id fallback")
					}
				}
			}
		})
	}
}

// TestUniversalFieldOptions tests the new option functions
func TestUniversalFieldOptions(t *testing.T) {
	t.Run("WithPublicMessage", func(t *testing.T) {
		err := NewValidationError("field", "invalid")
		err = NewValidationError("field", "invalid") // Create new with option
		err.PublicMessage = "User-friendly message"  // Set directly for test
		opt := WithPublicMessage("User-friendly message")
		opt(err)
		if err.PublicMessage != "User-friendly message" {
			t.Errorf("WithPublicMessage() failed, got %q", err.PublicMessage)
		}
	})

	t.Run("WithHelpURL", func(t *testing.T) {
		err := NewValidationError("field", "invalid")
		opt := WithHelpURL("https://docs.example.com")
		opt(err)
		if err.HelpURL != "https://docs.example.com" {
			t.Errorf("WithHelpURL() failed, got %q", err.HelpURL)
		}
	})

	t.Run("WithSuggestions", func(t *testing.T) {
		err := NewValidationError("field", "invalid")
		suggestions := []string{"Check spelling", "Verify format"}
		opt := WithSuggestions(suggestions...)
		opt(err)
		if len(err.Suggestions) != 2 {
			t.Errorf("WithSuggestions() failed, got %d suggestions", len(err.Suggestions))
		}
	})

	t.Run("WithMessageKey", func(t *testing.T) {
		err := NewValidationError("field", "invalid")
		opt := WithMessageKey("errors.validation.field")
		opt(err)
		if err.MessageKey != "errors.validation.field" {
			t.Errorf("WithMessageKey() failed, got %q", err.MessageKey)
		}
	})

	t.Run("WithOperationID", func(t *testing.T) {
		err := NewInternalError("component", nil, WithOperationID("op-123"))
		if err.OperationID != "op-123" {
			t.Errorf("WithOperationID() failed, got %q", err.OperationID)
		}
		// Should also set RequestID for backward compatibility
		if err.OperationID != "op-123" {
			t.Errorf("WithOperationID() should set RequestID for backward compat, got %q", err.OperationID)
		}
	})

	t.Run("WithTraceID", func(t *testing.T) {
		err := NewInternalError("component", nil, WithTraceID("trace-456"))
		if err.TraceID != "trace-456" {
			t.Errorf("WithTraceID() failed, got %q", err.TraceID)
		}
	})

	t.Run("WithSpanID", func(t *testing.T) {
		err := NewInternalError("component", nil, WithSpanID("span-789"))
		if err.SpanID != "span-789" {
			t.Errorf("WithSpanID() failed, got %q", err.SpanID)
		}
	})
}

// TestProtocolAgnosticWorkflow tests a complete workflow across protocols
func TestProtocolAgnosticWorkflow(t *testing.T) {
	// Create a rich error with all universal fields
	err := NewValidationError("email", "invalid format")
	WithPublicMessage("Please enter a valid email address")(err)
	WithHelpURL("https://docs.example.com/validation#email")(err)
	WithSuggestions("Check for typos", "Ensure @ symbol is present")(err)
	WithMessageKey("errors.validation.email_invalid")(err)
	WithOperationID("op-12345")(err)
	WithTraceID("trace-67890")(err)
	WithSpanID("span-abcde")(err)

	t.Run("HTTP Protocol", func(t *testing.T) {
		httpStatus := err.ToHTTPStatus()
		if httpStatus != 400 {
			t.Errorf("ToHTTPStatus() = %d, want 400", httpStatus)
		}

		json := err.ToPublicJSON()
		errorData := json[JSON_FIELD_ERROR].(map[string]interface{})

		// Verify key fields are present
		if errorData["public_message"] != "Please enter a valid email address" {
			t.Error("HTTP JSON should include public_message")
		}
	})

	t.Run("CLI Protocol", func(t *testing.T) {
		exitCode := err.ToExitCode()
		if exitCode != 1 {
			t.Errorf("ToExitCode() = %d, want 1", exitCode)
		}

		textOutput := err.ToText()
		if !strings.Contains(textOutput, "Please enter a valid email address") {
			t.Error("CLI text output should include public message")
		}

		coloredOutput := err.ToCLIOutput(true)
		if !strings.Contains(coloredOutput, "Check for typos") {
			t.Error("CLI colored output should include suggestions")
		}
	})

	t.Run("gRPC Protocol", func(t *testing.T) {
		grpcCode := err.ToGRPCCode()
		if grpcCode != 3 { // InvalidArgument
			t.Errorf("ToGRPCCode() = %d, want 3 (InvalidArgument)", grpcCode)
		}

		details := err.ToGRPCDetails()
		if details["help_url"] != "https://docs.example.com/validation#email" {
			t.Error("gRPC details should include help_url")
		}
		if suggestions, ok := details["suggestions"].([]string); ok {
			if len(suggestions) != 2 {
				t.Error("gRPC details should include all suggestions")
			}
		}
	})

	t.Run("Syslog Protocol", func(t *testing.T) {
		severity := err.ToSyslogSeverity()
		if severity != 6 { // Informational
			t.Errorf("ToSyslogSeverity() = %d, want 6 (Informational)", severity)
		}
	})

	t.Run("Tracing", func(t *testing.T) {
		if err.OperationID != "op-12345" {
			t.Error("OperationID should be preserved across protocols")
		}
		if err.TraceID != "trace-67890" {
			t.Error("TraceID should be preserved across protocols")
		}
		if err.SpanID != "span-abcde" {
			t.Error("SpanID should be preserved across protocols")
		}
	})
}

// TestErrorCollectionToPublicJSON tests ErrorCollection protocol-agnostic JSON
func TestErrorCollectionToPublicJSON(t *testing.T) {
	collection := NewValidationErrorCollection()
	collection.AddValidation("email", "Invalid email format")
	collection.AddValidation("age", "Must be at least 18")
	collection.WithOperationID("req-123")

	t.Run("Non-production mode", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    DEFAULT_STACK_DEPTH,
			ProductionMode:   false,
		})

		json := collection.ToPublicJSON()
		errorData := json["error"].(map[string]interface{})

		if errorData["operation_id"] != "req-123" {
			t.Error("ToPublicJSON() should include operation_id")
		}

		if validationErrors, ok := errorData["validation_errors"].([]ValidationError); ok {
			if len(validationErrors) != 2 {
				t.Errorf("ToPublicJSON() should include all validation errors, got %d", len(validationErrors))
			}
		}
	})

	t.Run("Production mode", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    DEFAULT_STACK_DEPTH,
			ProductionMode:   true,
		})

		json := collection.ToPublicJSON()
		errorData := json["error"].(map[string]interface{})

		// Validation errors should still be included in production
		if _, hasValidationErrors := errorData["validation_errors"]; !hasValidationErrors {
			t.Error("ToPublicJSON() should include validation_errors even in production")
		}
	})
}

// TestBackwardCompatibility tests that deprecated methods still work
func TestBackwardCompatibility(t *testing.T) {
	err := NewInternalError("component", nil,
		WithPublicMessage("Public message"),
		WithOperationID("op-123"),
	)

	t.Run("ClientSafeMessage() still works", func(t *testing.T) {
		msg := err.ToPublicJSON()[JSON_FIELD_ERROR].(map[string]interface{})[JSON_FIELD_MESSAGE].(string)
		if msg != "Public message" {
			t.Errorf("ClientSafeMessage() = %q, want %q", msg, "Public message")
		}
	})

	t.Run("ToClientJSON() still works", func(t *testing.T) {
		json := err.ToPublicJSON()
		if _, ok := json[JSON_FIELD_ERROR]; !ok {
			t.Error("ToClientJSON() should return valid JSON structure")
		}
	})

	collection := NewValidationErrorCollection()
	collection.AddValidation("field", "error")

	t.Run("ErrorCollection.ToPublicJSON() still works", func(t *testing.T) {
		json := collection.ToPublicJSON()
		if _, ok := json["error"]; !ok {
			t.Error("ErrorCollection.ToPublicJSON() should return valid JSON structure")
		}
	})
}
