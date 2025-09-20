package cuserr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestHTTPIntegrationEndToEnd tests complete HTTP request/response cycle
func TestHTTPIntegrationEndToEnd(t *testing.T) {
	// Create a test HTTP handler that uses cuserr
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = "req-test-123"
		}

		// Simulate different error scenarios based on path
		switch r.URL.Path {
		case "/users/not-found":
			err := NewCustomError(ErrNotFound, nil, "user not found").
				WithMetadata("user_id", "usr_404").
				WithMetadata("operation", "get_user").
				WithRequestID(requestID)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(err.ToHTTPStatus())
			if encodeErr := json.NewEncoder(w).Encode(err.ToClientJSON()); encodeErr != nil {
				http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
				return
			}

		case "/validation-error":
			err := NewCustomError(ErrInvalidInput, nil, "email format is invalid").
				WithMetadata("field", "email").
				WithMetadata("value", "invalid-email").
				WithRequestID(requestID)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(err.ToHTTPStatus())
			if encodeErr := json.NewEncoder(w).Encode(err.ToJSON()); encodeErr != nil {
				http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
				return
			}

		case "/internal-error":
			originalErr := fmt.Errorf("database connection failed")
			err := NewCustomError(ErrInternal, originalErr, "failed to process request").
				WithMetadata("database", "users_db").
				WithMetadata("retry_count", "3").
				WithRequestID(requestID)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(err.ToHTTPStatus())
			if encodeErr := json.NewEncoder(w).Encode(err.ToClientJSON()); encodeErr != nil {
				http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
				return
			}

		case "/rate-limit":
			err := NewCustomError(ErrRateLimit, nil, "API rate limit exceeded").
				WithMetadata("limit", "1000/hour").
				WithMetadata("reset_at", time.Now().Add(time.Hour).Format(time.RFC3339)).
				WithRequestID(requestID)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(err.ToHTTPStatus())
			if encodeErr := json.NewEncoder(w).Encode(err.ToJSON()); encodeErr != nil {
				http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
				return
			}

		default:
			w.WriteHeader(http.StatusOK)
			if _, writeErr := w.Write([]byte(`{"status": "ok"}`)); writeErr != nil {
				http.Error(w, "Write failed", http.StatusInternalServerError)
			}
		}
	})

	// Test scenarios
	testCases := []struct {
		name             string
		path             string
		expectedStatus   int
		expectedCategory string
		expectedCode     string
		checkMetadata    map[string]string
		checkRequestID   bool
	}{
		{
			name:             "Not Found Error",
			path:             "/users/not-found",
			expectedStatus:   404,
			expectedCategory: string(ErrorCategoryNotFound),
			expectedCode:     "NOT_FOUND",
			checkMetadata:    map[string]string{"user_id": "usr_404", "operation": "get_user"},
			checkRequestID:   true,
		},
		{
			name:             "Validation Error",
			path:             "/validation-error",
			expectedStatus:   400,
			expectedCategory: string(ErrorCategoryValidation),
			expectedCode:     "INVALID_INPUT",
			checkMetadata:    map[string]string{"field": "email", "value": "invalid-email"},
			checkRequestID:   true,
		},
		{
			name:             "Internal Error",
			path:             "/internal-error",
			expectedStatus:   500,
			expectedCategory: string(ErrorCategoryInternal),
			expectedCode:     "INTERNAL_ERROR",
			checkMetadata:    map[string]string{"database": "users_db", "retry_count": "3"},
			checkRequestID:   true,
		},
		{
			name:             "Rate Limit Error",
			path:             "/rate-limit",
			expectedStatus:   429,
			expectedCategory: string(ErrorCategoryRateLimit),
			expectedCode:     "RATE_LIMIT",
			checkMetadata:    map[string]string{"limit": "1000/hour"},
			checkRequestID:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("X-Request-ID", "req-integration-test")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			handler(w, req)

			// Verify status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			// Verify content type
			if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			// Parse response body
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse JSON response: %v", err)
			}

			// Verify error structure
			errorData, exists := response["error"].(map[string]interface{})
			if !exists {
				t.Fatal("Response should contain 'error' field")
			}

			// Verify category
			if category, ok := errorData["category"].(string); ok {
				if category != tc.expectedCategory {
					t.Errorf("Expected category %s, got %s", tc.expectedCategory, category)
				}
			} else {
				t.Error("Response should contain category field")
			}

			// Verify code
			if code, ok := errorData["code"].(string); ok {
				if code != tc.expectedCode {
					t.Errorf("Expected code %s, got %s", tc.expectedCode, code)
				}
			} else {
				t.Error("Response should contain code field")
			}

			// Verify request ID if expected
			if tc.checkRequestID {
				if requestID, ok := errorData["request_id"].(string); ok {
					if !strings.Contains(requestID, "req-") {
						t.Errorf("Request ID should contain 'req-', got %s", requestID)
					}
				} else {
					t.Error("Response should contain request_id field")
				}
			}

			// Verify metadata
			if len(tc.checkMetadata) > 0 {
				metadata, ok := errorData["metadata"].(map[string]interface{})
				if !ok {
					t.Error("Response should contain metadata field")
				} else {
					for key, expectedValue := range tc.checkMetadata {
						if value, exists := metadata[key]; exists {
							if value.(string) != expectedValue {
								t.Errorf("Expected metadata[%s] = %s, got %s", key, expectedValue, value)
							}
						} else {
							t.Errorf("Metadata should contain key %s", key)
						}
					}
				}
			}

			// Verify timestamp format
			if timestamp, ok := errorData["timestamp"].(string); ok {
				if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
					t.Errorf("Timestamp should be in RFC3339 format, got %s", timestamp)
				}
			} else {
				t.Error("Response should contain timestamp field")
			}
		})
	}
}

// TestHTTPProductionFiltering tests production mode error filtering
func TestHTTPProductionFiltering(t *testing.T) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	// Test in development mode first
	SetConfig(&Config{
		EnableStackTrace: true,
		ProductionMode:   false,
	})

	err := NewCustomError(ErrInternal, nil, "detailed internal error with sensitive data").
		WithMetadata("database_password", "secret123").
		WithMetadata("user_id", "usr_789").
		WithRequestID("req-dev-test")

	devJSON := err.ToClientJSON()
	devErrorData := devJSON[JSON_FIELD_ERROR].(map[string]interface{})

	// In development mode, should show detailed message
	if devErrorData[JSON_FIELD_MESSAGE] != "detailed internal error with sensitive data" {
		t.Error("Development mode should show detailed error messages")
	}

	// Should contain all metadata in development
	devMetadata := devErrorData[JSON_FIELD_METADATA].(map[string]string)
	if len(devMetadata) != 2 {
		t.Errorf("Development mode should contain all metadata, got %d items", len(devMetadata))
	}

	// Switch to production mode
	SetConfig(&Config{
		EnableStackTrace: false,
		ProductionMode:   true,
	})

	prodJSON := err.ToClientJSON()
	prodErrorData := prodJSON[JSON_FIELD_ERROR].(map[string]interface{})

	// In production mode, should show generic message for internal errors
	if prodErrorData[JSON_FIELD_MESSAGE] != "An internal error occurred" {
		t.Errorf("Production mode should show generic message, got: %s", prodErrorData[JSON_FIELD_MESSAGE])
	}

	// Should filter sensitive metadata in production
	prodMetadata, hasMetadata := prodErrorData[JSON_FIELD_METADATA].(map[string]string)
	if !hasMetadata || len(prodMetadata) != 1 {
		t.Error("Production mode should filter sensitive metadata, keeping only safe identifiers")
	}

	// Should keep safe identifier (user_id)
	if _, exists := prodMetadata["user_id"]; !exists {
		t.Error("Production mode should keep safe identifiers like user_id")
	}

	// Should remove sensitive data
	if _, exists := prodMetadata["database_password"]; exists {
		t.Error("Production mode should remove sensitive metadata")
	}
}

// TestHTTPMiddlewarePattern tests common middleware usage pattern
func TestHTTPMiddlewarePattern(t *testing.T) {
	// Simulate a middleware that handles cuserr errors
	errorMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Capture panic and convert to cuserr
			defer func() {
				if recovered := recover(); recovered != nil {
					err := NewCustomError(ErrInternal, nil, "unexpected server error").
						WithMetadata("panic", fmt.Sprintf("%v", recovered)).
						WithRequestID(r.Header.Get("X-Request-ID"))

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(err.ToHTTPStatus())
					json.NewEncoder(w).Encode(err.ToClientJSON())
				}
			}()

			next(w, r)
		}
	}

	// Test handler that might panic
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("panic") == "true" {
			panic("something went wrong")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap handler with middleware
	wrappedHandler := errorMiddleware(panicHandler)

	// Test normal operation
	t.Run("Normal Operation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if body := w.Body.String(); body != "OK" {
			t.Errorf("Expected body 'OK', got %s", body)
		}
	})

	// Test panic recovery
	t.Run("Panic Recovery", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test?panic=true", nil)
		req.Header.Set("X-Request-ID", "req-panic-test")
		w := httptest.NewRecorder()

		wrappedHandler(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		errorData := response["error"].(map[string]interface{})

		// Verify error structure
		if errorData["category"] != string(ErrorCategoryInternal) {
			t.Error("Panic should be categorized as internal error")
		}

		if errorData["request_id"] != "req-panic-test" {
			t.Error("Should preserve request ID from header")
		}

		// Verify panic information is captured
		metadata := errorData["metadata"].(map[string]interface{})
		panicValue, exists := metadata["panic"]
		if !exists || !strings.Contains(panicValue.(string), "something went wrong") {
			t.Error("Should capture panic information in metadata")
		}
	})
}

// TestContextIntegration tests integration with context.Context
func TestContextIntegration(t *testing.T) {
	// Simulate a service that uses context for request tracing
	processWithContext := func(ctx context.Context, userID string) error {
		// Extract request ID from context
		requestID, ok := ctx.Value("request_id").(string)
		if !ok {
			requestID = "unknown"
		}

		// Simulate timeout
		select {
		case <-ctx.Done():
			return NewCustomError(ErrTimeout, ctx.Err(), "operation cancelled or timed out").
				WithMetadata("user_id", userID).
				WithMetadata("operation", "process_data").
				WithRequestID(requestID)
		case <-time.After(10 * time.Millisecond):
			// Normal processing
			if userID == "invalid" {
				return NewCustomError(ErrInvalidInput, nil, "invalid user ID").
					WithMetadata("user_id", userID).
					WithRequestID(requestID)
			}
			return nil
		}
	}

	t.Run("Normal Processing", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "req-ctx-normal")
		err := processWithContext(ctx, "usr_123")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Context Timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		ctx = context.WithValue(ctx, "request_id", "req-ctx-timeout")

		time.Sleep(2 * time.Millisecond) // Ensure context times out
		err := processWithContext(ctx, "usr_123")

		if err == nil {
			t.Fatal("Expected timeout error")
		}

		var customErr *CustomError
		if !errors.As(err, &customErr) {
			t.Fatal("Expected CustomError")
		}

		if customErr.Category != ErrorCategoryTimeout {
			t.Error("Should be categorized as timeout error")
		}

		if customErr.RequestID != "req-ctx-timeout" {
			t.Error("Should preserve request ID from context")
		}

		// Verify wrapped context error
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Error("Should wrap original context error")
		}
	})

	t.Run("Validation Error with Context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "req-ctx-validation")
		err := processWithContext(ctx, "invalid")

		if err == nil {
			t.Fatal("Expected validation error")
		}

		var customErr *CustomError
		if !errors.As(err, &customErr) {
			t.Fatal("Expected CustomError")
		}

		if customErr.Category != ErrorCategoryValidation {
			t.Error("Should be categorized as validation error")
		}

		if userID, exists := customErr.GetMetadata("user_id"); !exists || userID != "invalid" {
			t.Error("Should capture user_id in metadata")
		}

		if customErr.RequestID != "req-ctx-validation" {
			t.Error("Should preserve request ID from context")
		}
	})
}
