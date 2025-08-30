package cuserr

import (
	"runtime"
	"strings"
	"testing"
)

// TestDetailedError tests comprehensive detailed error output
func TestDetailedError(t *testing.T) {
	// Enable stack trace for this test
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	SetConfig(&Config{
		EnableStackTrace: true,
		MaxStackDepth:    10,
		ProductionMode:   false,
	})

	t.Run("Basic Detailed Error", func(t *testing.T) {
		err := createTestErrorWithStack("test detailed error").
			WithMetadata("component", "auth").
			WithMetadata("user_id", "usr_123").
			WithRequestID("req-detailed-001")

		detailed := err.DetailedError()

		// Should contain error message
		if !strings.Contains(detailed, "test detailed error") {
			t.Error("Detailed error should contain error message")
		}

		// Should contain category and code
		if !strings.Contains(detailed, string(ErrorCategoryInternal)) {
			t.Error("Detailed error should contain category")
		}
		if !strings.Contains(detailed, "INTERNAL_ERROR") {
			t.Error("Detailed error should contain error code")
		}

		// Should contain request ID
		if !strings.Contains(detailed, "req-detailed-001") {
			t.Error("Detailed error should contain request ID")
		}

		// Should contain metadata
		if !strings.Contains(detailed, "component: auth") {
			t.Error("Detailed error should contain metadata")
		}
		if !strings.Contains(detailed, "user_id: usr_123") {
			t.Error("Detailed error should contain all metadata")
		}

		// Should contain stack trace
		if !strings.Contains(detailed, "Stack Trace:") {
			t.Error("Detailed error should contain stack trace header")
		}

		// Should contain function names in stack trace
		if !strings.Contains(detailed, "createTestErrorWithStack") {
			t.Error("Stack trace should contain function names")
		}
	})

	t.Run("Detailed Error with Wrapped Error", func(t *testing.T) {
		originalErr := NewCustomError(ErrNotFound, nil, "user not found")
		wrappedErr := NewCustomError(ErrInternal, originalErr, "service unavailable").
			WithMetadata("service", "user-service")

		detailed := wrappedErr.DetailedError()

		// Should contain both error messages
		if !strings.Contains(detailed, "service unavailable") {
			t.Error("Should contain wrapper error message")
		}
		if !strings.Contains(detailed, "user not found") {
			t.Error("Should contain wrapped error message")
		}

		// Should show wrapped error section
		if !strings.Contains(detailed, "Wrapped:") {
			t.Error("Should contain wrapped error section")
		}
	})

	t.Run("Detailed Error Without Stack Trace", func(t *testing.T) {
		// Disable stack trace
		SetConfig(&Config{
			EnableStackTrace: false,
			MaxStackDepth:    0,
			ProductionMode:   false,
		})

		err := NewCustomError(ErrTimeout, nil, "operation timeout").
			WithMetadata("timeout_duration", "30s")

		detailed := err.DetailedError()

		// Should contain basic error info
		if !strings.Contains(detailed, "operation timeout") {
			t.Error("Should contain error message")
		}
		if !strings.Contains(detailed, "timeout_duration: 30s") {
			t.Error("Should contain metadata")
		}

		// Should not contain stack trace
		if strings.Contains(detailed, "Stack Trace:") {
			t.Error("Should not contain stack trace when disabled")
		}
	})
}

// TestShortError tests concise error format
func TestShortError(t *testing.T) {
	t.Run("Short Error with Request ID", func(t *testing.T) {
		err := NewCustomError(ErrUnauthorized, nil, "authentication failed").
			WithMetadata("api_key", "key_123").
			WithRequestID("req-short-001")

		short := err.ShortError()

		// Should contain request ID
		if !strings.Contains(short, "[req-short-001]") {
			t.Error("Short error should contain request ID in brackets")
		}

		// Should contain category
		if !strings.Contains(short, string(ErrorCategoryUnauthorized)) {
			t.Error("Short error should contain category")
		}

		// Should contain code
		if !strings.Contains(short, "UNAUTHORIZED") {
			t.Error("Short error should contain error code")
		}

		// Should contain message
		if !strings.Contains(short, "authentication failed") {
			t.Error("Short error should contain error message")
		}

		// Should not contain metadata (too verbose for short format)
		if strings.Contains(short, "api_key") {
			t.Error("Short error should not contain metadata")
		}

		// Should be single line
		if strings.Contains(short, "\n") {
			t.Error("Short error should be single line")
		}
	})

	t.Run("Short Error without Request ID", func(t *testing.T) {
		err := NewCustomError(ErrInvalidInput, nil, "invalid email format")

		short := err.ShortError()

		// Should not contain brackets (no request ID)
		if strings.Contains(short, "[") || strings.Contains(short, "]") {
			t.Error("Short error without request ID should not contain brackets")
		}

		// Should contain category, code, and message
		if !strings.Contains(short, string(ErrorCategoryValidation)) {
			t.Error("Should contain category")
		}
		if !strings.Contains(short, "INVALID_INPUT") {
			t.Error("Should contain error code")
		}
		if !strings.Contains(short, "invalid email format") {
			t.Error("Should contain error message")
		}

		// Should follow format: "category (code): message"
		expected := "validation (INVALID_INPUT): invalid email format"
		if short != expected {
			t.Errorf("Expected format '%s', got '%s'", expected, short)
		}
	})
}

// TestStackTraceCapture tests stack trace functionality
func TestStackTraceCapture(t *testing.T) {
	// Enable stack trace for this test
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	SetConfig(&Config{
		EnableStackTrace: true,
		MaxStackDepth:    5,
		ProductionMode:   false,
	})

	t.Run("Stack Trace Capture", func(t *testing.T) {
		err := createTestErrorWithStack("stack trace test")

		stackTrace := err.GetStackTrace()
		if len(stackTrace) == 0 {
			t.Fatal("Should capture stack trace when enabled")
		}

		// Should contain function information
		found := false
		for _, frame := range stackTrace {
			if strings.Contains(frame.Function, "createTestErrorWithStack") {
				found = true
				if frame.Line <= 0 {
					t.Error("Stack frame should have valid line number")
				}
				if frame.File == "" {
					t.Error("Stack frame should have file path")
				}
				break
			}
		}

		if !found {
			t.Error("Stack trace should contain createTestErrorWithStack function")
		}
	})

	t.Run("Stack Trace String Format", func(t *testing.T) {
		err := createTestErrorWithStack("stack trace string test")

		stackString := err.GetStackTraceString()
		if stackString == "" {
			t.Fatal("Should return non-empty stack trace string")
		}

		// Should contain function names and file paths
		if !strings.Contains(stackString, "createTestErrorWithStack") {
			t.Error("Stack trace string should contain function names")
		}

		// Should contain file paths
		if !strings.Contains(stackString, ".go") {
			t.Error("Stack trace string should contain file paths")
		}

		// Should contain line numbers
		lines := strings.Split(stackString, "\n")
		foundLineNumber := false
		for _, line := range lines {
			if strings.Contains(line, ":") && strings.Contains(line, ".go") {
				foundLineNumber = true
				break
			}
		}
		if !foundLineNumber {
			t.Error("Stack trace string should contain line numbers")
		}
	})

	t.Run("Stack Trace Depth Limit", func(t *testing.T) {
		// Set very low depth limit
		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    2,
			ProductionMode:   false,
		})

		err := deeplyNestedFunction(5)
		stackTrace := err.GetStackTrace()

		// Should respect depth limit
		if len(stackTrace) > 2 {
			t.Errorf("Stack trace should be limited to 2 frames, got %d", len(stackTrace))
		}
	})

	t.Run("Stack Trace Disabled", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: false,
			MaxStackDepth:    10,
			ProductionMode:   false,
		})

		err := createTestErrorWithStack("no stack trace test")

		stackTrace := err.GetStackTrace()
		if len(stackTrace) != 0 {
			t.Error("Should not capture stack trace when disabled")
		}

		stackString := err.GetStackTraceString()
		if stackString != "" {
			t.Error("Should return empty stack trace string when disabled")
		}
	})
}

// TestFilterStackTrace tests stack trace filtering
func TestFilterStackTrace(t *testing.T) {
	// Enable stack trace for this test
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	SetConfig(&Config{
		EnableStackTrace: true,
		MaxStackDepth:    10,
		ProductionMode:   false,
	})

	t.Run("Filter Stack Trace", func(t *testing.T) {
		err := createTestErrorWithStack("filter test")

		originalTrace := err.GetStackTrace()
		originalLength := len(originalTrace)

		// Filter out test functions
		err.FilterStackTrace("testing.")

		filteredTrace := err.GetStackTrace()

		// Should have fewer frames after filtering
		if len(filteredTrace) >= originalLength {
			t.Error("Filtered trace should have fewer frames")
		}

		// Should not contain filtered functions
		for _, frame := range filteredTrace {
			if strings.Contains(frame.Function, "testing.") {
				t.Error("Filtered trace should not contain filtered functions")
			}
		}
	})

	t.Run("Filter Multiple Patterns", func(t *testing.T) {
		err := createTestErrorWithStack("multi filter test")

		// Filter multiple patterns
		err.FilterStackTrace("testing.", "runtime.")

		filteredTrace := err.GetStackTrace()

		// Should not contain any filtered patterns
		for _, frame := range filteredTrace {
			if strings.Contains(frame.Function, "testing.") || strings.Contains(frame.Function, "runtime.") {
				t.Error("Should filter all specified patterns")
			}
		}
	})

	t.Run("Filter with Empty Patterns", func(t *testing.T) {
		err := createTestErrorWithStack("empty filter test")

		originalTrace := err.GetStackTrace()
		originalLength := len(originalTrace)

		// Filter with empty patterns (should not change anything)
		err.FilterStackTrace()

		filteredTrace := err.GetStackTrace()

		if len(filteredTrace) != originalLength {
			t.Error("Empty filter should not change stack trace")
		}
	})
}

// TestStackTraceManipulation tests stack trace manipulation methods
func TestStackTraceManipulation(t *testing.T) {
	t.Run("WithStackTrace", func(t *testing.T) {
		err := NewCustomError(ErrInternal, nil, "manual stack trace")

		// Create manual stack trace
		manualTrace := []StackFrame{
			{
				Function: "main.main",
				File:     "/app/main.go",
				Line:     42,
			},
			{
				Function: "main.processRequest",
				File:     "/app/handler.go",
				Line:     123,
			},
		}

		err.WithStackTrace(manualTrace)

		retrievedTrace := err.GetStackTrace()
		if len(retrievedTrace) != 2 {
			t.Fatalf("Expected 2 stack frames, got %d", len(retrievedTrace))
		}

		// Verify first frame
		if retrievedTrace[0].Function != "main.main" {
			t.Error("Should preserve function name")
		}
		if retrievedTrace[0].File != "/app/main.go" {
			t.Error("Should preserve file path")
		}
		if retrievedTrace[0].Line != 42 {
			t.Error("Should preserve line number")
		}

		// Verify second frame
		if retrievedTrace[1].Function != "main.processRequest" {
			t.Error("Should preserve second frame function")
		}
	})

	t.Run("WithStackTrace Creates Copy", func(t *testing.T) {
		err := NewCustomError(ErrInternal, nil, "copy test")

		originalTrace := []StackFrame{
			{Function: "test", File: "test.go", Line: 1},
		}

		err.WithStackTrace(originalTrace)

		// Modify original slice
		originalTrace[0].Function = "modified"

		// Retrieved trace should not be affected
		retrievedTrace := err.GetStackTrace()
		if retrievedTrace[0].Function != "test" {
			t.Error("WithStackTrace should create a copy of the input")
		}
	})

	t.Run("ClearStackTrace", func(t *testing.T) {
		// Enable stack trace to capture one first
		originalConfig := GetConfig()
		defer SetConfig(originalConfig)

		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    5,
			ProductionMode:   false,
		})

		err := createTestErrorWithStack("clear test")

		// Verify stack trace was captured
		if len(err.GetStackTrace()) == 0 {
			t.Fatal("Should have captured stack trace initially")
		}

		// Clear stack trace
		err.ClearStackTrace()

		// Should be empty after clearing
		if len(err.GetStackTrace()) != 0 {
			t.Error("Stack trace should be empty after clearing")
		}

		// Stack trace string should also be empty
		if err.GetStackTraceString() != "" {
			t.Error("Stack trace string should be empty after clearing")
		}
	})

	t.Run("GetStackTrace Returns Copy", func(t *testing.T) {
		originalConfig := GetConfig()
		defer SetConfig(originalConfig)

		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    5,
			ProductionMode:   false,
		})

		err := createTestErrorWithStack("copy protection test")

		trace1 := err.GetStackTrace()
		trace2 := err.GetStackTrace()

		// Should be different slices
		if &trace1[0] == &trace2[0] {
			t.Error("GetStackTrace should return copies, not the same slice")
		}

		// But should have same content
		if len(trace1) != len(trace2) {
			t.Error("Copies should have same length")
		}

		for i := range trace1 {
			if trace1[i] != trace2[i] {
				t.Error("Copies should have same content")
			}
		}

		// Modifying returned slice should not affect the error
		if len(trace1) > 0 {
			trace1[0].Function = "modified"

			trace3 := err.GetStackTrace()
			if trace3[0].Function == "modified" {
				t.Error("Modifying returned slice should not affect internal stack trace")
			}
		}
	})
}

// Helper functions for testing

// createTestErrorWithStack creates an error and captures its stack
func createTestErrorWithStack(message string) *CustomError {
	return NewCustomError(ErrInternal, nil, message)
}

// deeplyNestedFunction creates errors at various nesting levels
func deeplyNestedFunction(depth int) *CustomError {
	if depth <= 0 {
		return NewCustomError(ErrTimeout, nil, "deep error")
	}
	return deeplyNestedFunction(depth - 1)
}

// TestStackTraceAccuracy tests that stack traces are accurate
func TestStackTraceAccuracy(t *testing.T) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	SetConfig(&Config{
		EnableStackTrace: true,
		MaxStackDepth:    10,
		ProductionMode:   false,
	})

	t.Run("Stack Trace Line Accuracy", func(t *testing.T) {
		// Get current line for comparison
		_, _, currentLine, ok := runtime.Caller(0)
		if !ok {
			t.Fatal("Could not get current line")
		}

		err := NewCustomError(ErrInternal, nil, "line accuracy test") // This line + 1

		stackTrace := err.GetStackTrace()
		if len(stackTrace) == 0 {
			t.Fatal("Should have stack trace")
		}

		// Find the frame for this test function
		var testFrame *StackFrame
		for i := range stackTrace {
			if strings.Contains(stackTrace[i].Function, "TestStackTraceAccuracy") {
				testFrame = &stackTrace[i]
				break
			}
		}

		if testFrame == nil {
			t.Fatal("Should find test function in stack trace")
		}

		// The error creation should be captured at approximately the right line
		// Allow some tolerance for compiler optimizations
		expectedLine := currentLine + 2 // Line where NewCustomError is called
		if testFrame.Line < expectedLine-5 || testFrame.Line > expectedLine+5 {
			t.Errorf("Stack trace line number inaccurate: expected around %d, got %d", expectedLine, testFrame.Line)
		}

		// File should contain this test file
		if !strings.Contains(testFrame.File, "cuserr_stack_trace_test.go") {
			t.Errorf("Stack trace file should be test file, got %s", testFrame.File)
		}
	})
}
