package cuserr

import (
	"fmt"
	"runtime"
	"strings"
)

// captureStackTrace captures the current stack trace with configurable depth
func captureStackTrace(skip int) []StackFrame {
	config := GetConfig()
	if !config.EnableStackTrace {
		return nil
	}

	var frames []StackFrame
	maxDepth := config.MaxStackDepth
	if maxDepth <= 0 {
		maxDepth = DEFAULT_STACK_DEPTH
	}

	for i := skip; i < skip+maxDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			break
		}

		frame := StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		}

		frames = append(frames, frame)

		// Stop at main or testing functions to avoid noise
		if strings.Contains(fn.Name(), MAIN_FUNCTION_NAME) ||
			strings.Contains(fn.Name(), TESTING_FUNCTION_NAME) {
			break
		}
	}

	return frames
}

// DetailedError returns detailed error information for logging
// This includes full stack trace and metadata information
func (e *CustomError) DetailedError() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(LOG_TEMPLATE_ERROR_DETAIL, e.Message, e.Category, e.Code))
	sb.WriteString("\n")

	if e.RequestID != "" {
		sb.WriteString(fmt.Sprintf(LOG_TEMPLATE_REQUEST_ID, e.RequestID))
		sb.WriteString("\n")
	}

	// Add metadata information
	metadata := e.GetAllMetadata() // Thread-safe access
	if len(metadata) > 0 {
		sb.WriteString("Metadata:\n")
		for k, v := range metadata {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	if e.Wrapped != nil {
		sb.WriteString(fmt.Sprintf(LOG_TEMPLATE_WRAPPED_ERROR, e.Wrapped))
		sb.WriteString("\n")
	}

	// Add stack trace if available
	e.mu.RLock()
	stackTrace := make([]StackFrame, len(e.stackTrace))
	copy(stackTrace, e.stackTrace)
	e.mu.RUnlock()

	if len(stackTrace) > 0 {
		sb.WriteString("Stack Trace:\n")
		for _, frame := range stackTrace {
			sb.WriteString(fmt.Sprintf(LOG_TEMPLATE_STACK_FRAME, frame.Function, frame.File, frame.Line))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ShortError returns a concise error representation for logging
func (e *CustomError) ShortError() string {
	if e.RequestID != "" {
		return fmt.Sprintf("[%s] %s (%s): %s",
			e.RequestID, e.Category, e.Code, e.Message)
	}
	return fmt.Sprintf("%s (%s): %s", e.Category, e.Code, e.Message)
}

// GetStackTrace returns the captured stack trace in a thread-safe manner
func (e *CustomError) GetStackTrace() []StackFrame {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]StackFrame, len(e.stackTrace))
	copy(result, e.stackTrace)
	return result
}

// GetStackTraceString returns stack trace as formatted string in a thread-safe manner
func (e *CustomError) GetStackTraceString() string {
	e.mu.RLock()
	stackTrace := make([]StackFrame, len(e.stackTrace))
	copy(stackTrace, e.stackTrace)
	e.mu.RUnlock()

	if len(stackTrace) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, frame := range stackTrace {
		sb.WriteString(fmt.Sprintf(LOG_TEMPLATE_STACK_FRAME, frame.Function, frame.File, frame.Line))
		sb.WriteString("\n")
	}

	return sb.String()
}

// FilterStackTrace removes frames that match the given patterns in a thread-safe manner
func (e *CustomError) FilterStackTrace(patterns ...string) *CustomError {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.stackTrace) == 0 || len(patterns) == 0 {
		return e
	}

	var filtered []StackFrame
	for _, frame := range e.stackTrace {
		shouldFilter := false
		for _, pattern := range patterns {
			if strings.Contains(frame.Function, pattern) {
				shouldFilter = true
				break
			}
		}
		if !shouldFilter {
			filtered = append(filtered, frame)
		}
	}

	e.stackTrace = filtered
	return e
}

// WithStackTrace manually sets the stack trace (useful for testing) in a thread-safe manner
func (e *CustomError) WithStackTrace(frames []StackFrame) *CustomError {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Make a copy to prevent external modification
	e.stackTrace = make([]StackFrame, len(frames))
	copy(e.stackTrace, frames)
	e.stackTraceCleared = false // Reset cleared flag when manually setting
	return e
}

// ClearStackTrace removes the stack trace to save memory in a thread-safe manner
func (e *CustomError) ClearStackTrace() *CustomError {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stackTrace = nil
	e.stackTraceCleared = true // Mark as explicitly cleared
	return e
}
