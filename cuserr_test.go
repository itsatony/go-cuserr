package cuserr

import (
	"errors"
	"testing"
)

func TestNewCustomError(t *testing.T) {
	tests := []struct {
		name         string
		sentinel     error
		wrapped      error
		message      string
		wantCategory ErrorCategory
		wantCode     string
	}{
		{
			name:         "not found error",
			sentinel:     ErrNotFound,
			wrapped:      errors.New("db error"),
			message:      "user not found",
			wantCategory: ErrorCategoryNotFound,
			wantCode:     ERROR_CODE_NOT_FOUND,
		},
		{
			name:         "validation error",
			sentinel:     ErrInvalidInput,
			wrapped:      nil,
			message:      "invalid email format",
			wantCategory: ErrorCategoryValidation,
			wantCode:     ERROR_CODE_INVALID_INPUT,
		},
		{
			name:         "unauthorized error",
			sentinel:     ErrUnauthorized,
			wrapped:      nil,
			message:      "authentication required",
			wantCategory: ErrorCategoryUnauthorized,
			wantCode:     ERROR_CODE_UNAUTHORIZED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCustomError(tt.sentinel, tt.wrapped, tt.message)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Category != tt.wantCategory {
				t.Errorf("Category = %v, want %v", err.Category, tt.wantCategory)
			}

			if err.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", err.Code, tt.wantCode)
			}

			if err.Message != tt.message {
				t.Errorf("Message = %v, want %v", err.Message, tt.message)
			}

			if err.Wrapped != tt.wrapped {
				t.Errorf("Wrapped = %v, want %v", err.Wrapped, tt.wrapped)
			}

			if err.Sentinel != tt.sentinel {
				t.Errorf("Sentinel = %v, want %v", err.Sentinel, tt.sentinel)
			}

			if err.Timestamp.IsZero() {
				t.Error("Timestamp should not be zero")
			}
		})
	}
}

func TestNewCustomErrorWithCategory(t *testing.T) {
	category := ErrorCategoryInternal
	code := "CUSTOM_ERROR"
	message := "custom error message"

	err := NewCustomErrorWithCategory(category, code, message)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Category != category {
		t.Errorf("Category = %v, want %v", err.Category, category)
	}

	if err.Code != code {
		t.Errorf("Code = %v, want %v", err.Code, code)
	}

	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
}

func TestCustomError_WithMetadata(t *testing.T) {
	err := NewCustomError(ErrNotFound, nil, "test error")

	err.WithMetadata("user_id", "123").WithMetadata("email", "test@example.com")

	userID, exists := err.GetMetadata("user_id")
	if !exists || userID != "123" {
		t.Errorf("GetMetadata(user_id) = %v, %v, want 123, true", userID, exists)
	}

	email, exists := err.GetMetadata("email")
	if !exists || email != "test@example.com" {
		t.Errorf("GetMetadata(email) = %v, %v, want test@example.com, true", email, exists)
	}

	_, exists = err.GetMetadata("nonexistent")
	if exists {
		t.Error("GetMetadata(nonexistent) should return false")
	}
}

func TestCustomError_WithRequestID(t *testing.T) {
	err := NewCustomError(ErrNotFound, nil, "test error")
	requestID := "req-123"

	err.WithRequestID(requestID)

	if err.RequestID != requestID {
		t.Errorf("RequestID = %v, want %v", err.RequestID, requestID)
	}
}

func TestCustomError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *CustomError
		want string
	}{
		{
			name: "error without wrapped",
			err:  NewCustomError(ErrNotFound, nil, "user not found"),
			want: "user not found",
		},
		{
			name: "error with wrapped",
			err:  NewCustomError(ErrNotFound, errors.New("db error"), "user not found"),
			want: "user not found: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomError_Unwrap(t *testing.T) {
	wrapped := errors.New("original error")
	err := NewCustomError(ErrInternal, wrapped, "wrapper error")

	unwrapped := err.Unwrap()
	if unwrapped != wrapped {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, wrapped)
	}
}

func TestCustomError_Is(t *testing.T) {
	wrapped := errors.New("original error")
	err := NewCustomError(ErrNotFound, wrapped, "test error")

	// Test sentinel error checking
	if !errors.Is(err, ErrNotFound) {
		t.Error("errors.Is(err, ErrNotFound) should be true")
	}

	// Test wrapped error checking
	if !errors.Is(err, wrapped) {
		t.Error("errors.Is(err, wrapped) should be true")
	}

	// Test negative case
	if errors.Is(err, ErrUnauthorized) {
		t.Error("errors.Is(err, ErrUnauthorized) should be false")
	}
}

func TestCustomError_ToHTTPStatus(t *testing.T) {
	tests := []struct {
		category   ErrorCategory
		wantStatus int
	}{
		{ErrorCategoryValidation, HTTP_STATUS_BAD_REQUEST},
		{ErrorCategoryUnauthorized, HTTP_STATUS_UNAUTHORIZED},
		{ErrorCategoryForbidden, HTTP_STATUS_FORBIDDEN},
		{ErrorCategoryNotFound, HTTP_STATUS_NOT_FOUND},
		{ErrorCategoryTimeout, HTTP_STATUS_REQUEST_TIMEOUT},
		{ErrorCategoryConflict, HTTP_STATUS_CONFLICT},
		{ErrorCategoryRateLimit, HTTP_STATUS_TOO_MANY_REQUESTS},
		{ErrorCategoryExternal, HTTP_STATUS_BAD_GATEWAY},
		{ErrorCategoryInternal, HTTP_STATUS_INTERNAL_SERVER_ERROR},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			err := NewCustomErrorWithCategory(tt.category, "TEST", "test error")
			got := err.ToHTTPStatus()
			if got != tt.wantStatus {
				t.Errorf("ToHTTPStatus() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func TestCustomError_ToJSON(t *testing.T) {
	err := NewCustomError(ErrNotFound, nil, "user not found")
	err.WithMetadata("user_id", "123").WithRequestID("req-456")

	jsonData := err.ToJSON()

	errorData, ok := jsonData[JSON_FIELD_ERROR].(map[string]interface{})
	if !ok {
		t.Fatal("JSON should contain error field")
	}

	if errorData[JSON_FIELD_CODE] != ERROR_CODE_NOT_FOUND {
		t.Errorf("JSON code = %v, want %v", errorData[JSON_FIELD_CODE], ERROR_CODE_NOT_FOUND)
	}

	if errorData[JSON_FIELD_MESSAGE] != "user not found" {
		t.Errorf("JSON message = %v, want %v", errorData[JSON_FIELD_MESSAGE], "user not found")
	}

	if errorData[JSON_FIELD_CATEGORY] != ErrorCategoryNotFound {
		t.Errorf("JSON category = %v, want %v", errorData[JSON_FIELD_CATEGORY], ErrorCategoryNotFound)
	}

	if errorData[JSON_FIELD_REQUEST_ID] != "req-456" {
		t.Errorf("JSON request_id = %v, want %v", errorData[JSON_FIELD_REQUEST_ID], "req-456")
	}
}

func TestHelperFunctions(t *testing.T) {
	err := NewCustomError(ErrNotFound, nil, "test error")
	err.WithMetadata("test_key", "test_value")

	// Test IsErrorCategory
	if !IsErrorCategory(err, ErrorCategoryNotFound) {
		t.Error("IsErrorCategory should return true for matching category")
	}

	if IsErrorCategory(err, ErrorCategoryValidation) {
		t.Error("IsErrorCategory should return false for non-matching category")
	}

	// Test IsErrorCode
	if !IsErrorCode(err, ERROR_CODE_NOT_FOUND) {
		t.Error("IsErrorCode should return true for matching code")
	}

	if IsErrorCode(err, ERROR_CODE_UNAUTHORIZED) {
		t.Error("IsErrorCode should return false for non-matching code")
	}

	// Test GetErrorCategory
	if GetErrorCategory(err) != ErrorCategoryNotFound {
		t.Error("GetErrorCategory should return correct category")
	}

	// Test GetErrorCode
	if GetErrorCode(err) != ERROR_CODE_NOT_FOUND {
		t.Error("GetErrorCode should return correct code")
	}

	// Test GetErrorMetadata
	value, exists := GetErrorMetadata(err, "test_key")
	if !exists || value != "test_value" {
		t.Error("GetErrorMetadata should return correct value")
	}
}

func TestWrapWithCustomError(t *testing.T) {
	originalErr := errors.New("original error")

	wrapped := WrapWithCustomError(
		originalErr,
		ErrorCategoryInternal,
		"WRAP_TEST",
		"wrapped error message")

	if wrapped == nil {
		t.Fatal("WrapWithCustomError should not return nil")
	}

	if wrapped.Category != ErrorCategoryInternal {
		t.Error("Wrapped error should have correct category")
	}

	if wrapped.Code != "WRAP_TEST" {
		t.Error("Wrapped error should have correct code")
	}

	if !errors.Is(wrapped, originalErr) {
		t.Error("Wrapped error should be identifiable with errors.Is")
	}
}

func TestErrorWithContext(t *testing.T) {
	original := errors.New("original error")
	contextual := ErrorWithContext(original, "operation failed")

	expected := "operation failed: original error"
	if contextual.Error() != expected {
		t.Errorf("ErrorWithContext() = %v, want %v", contextual.Error(), expected)
	}

	// Test with nil error
	if ErrorWithContext(nil, "context") != nil {
		t.Error("ErrorWithContext with nil should return nil")
	}
}

func TestConfig(t *testing.T) {
	// Save original config
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	// Test default config
	defaultConfig := DefaultConfig()
	if !defaultConfig.EnableStackTrace {
		t.Error("Default config should enable stack trace")
	}

	if defaultConfig.MaxStackDepth != DEFAULT_STACK_DEPTH {
		t.Error("Default config should have correct max stack depth")
	}

	// Test setting custom config
	customConfig := &Config{
		EnableStackTrace: false,
		MaxStackDepth:    5,
		ProductionMode:   true,
	}

	SetConfig(customConfig)
	currentConfig := GetConfig()

	if currentConfig.EnableStackTrace {
		t.Error("Custom config should disable stack trace")
	}

	if currentConfig.MaxStackDepth != 5 {
		t.Error("Custom config should have correct max stack depth")
	}

	if !currentConfig.ProductionMode {
		t.Error("Custom config should enable production mode")
	}
}
