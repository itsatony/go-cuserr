package cuserr

import (
	"fmt"
	"strings"
)

// ToExitCode maps error category to Unix exit codes for CLI applications.
// Returns standard Unix exit codes that work across all platforms.
//
// Exit Code Reference:
//   - 0: Success (not used for errors)
//   - 1: General errors (validation, not found, state issues)
//   - 69: Service unavailable (EX_UNAVAILABLE)
//   - 70: Internal software error (EX_SOFTWARE)
//   - 74: Data error (EX_IOERR)
//   - 124: Command timed out
//   - 126: Permission denied
//   - 130: Terminated by Ctrl+C
//
// Example:
//
//	err := cuserr.NotFound("user", id)
//	os.Exit(err.ToExitCode())  // exits with code 1
func (e *CustomError) ToExitCode() int {
	return CategoryToExitCode(e.Category)
}

// CategoryToExitCode maps error categories to Unix exit codes.
func CategoryToExitCode(category ErrorCategory) int {
	switch category {
	case ErrorCategoryValidation:
		return 1 // General error - invalid input
	case ErrorCategoryUnauthorized, ErrorCategoryForbidden:
		return 126 // Permission denied
	case ErrorCategoryNotFound:
		return 1 // General error - resource not found
	case ErrorCategoryTimeout:
		return 124 // Command timed out
	case ErrorCategoryConflict:
		return 1 // General error - state conflict
	case ErrorCategoryRateLimit:
		return 1 // General error - too many requests
	case ErrorCategoryExternal:
		return 69 // Service unavailable (EX_UNAVAILABLE)
	case ErrorCategoryInternal:
		return 70 // Internal software error (EX_SOFTWARE)
	default:
		return 1 // General error
	}
}

// ToText returns a plain text representation of the error for CLI output.
// Uses the public message if available, otherwise falls back to the developer message.
//
// Example:
//
//	err := cuserr.Validation("email", "invalid format",
//	    cuserr.WithPublicMessage("Please enter a valid email address"),
//	)
//	fmt.Fprintln(os.Stderr, err.ToText())  // "Error: Please enter a valid email address"
func (e *CustomError) ToText() string {
	var parts []string

	// Use public message if available, otherwise use developer message
	message := e.PublicMessage
	if message == "" {
		message = e.Message
	}

	parts = append(parts, fmt.Sprintf("Error: %s", message))

	// Add error code if available
	if e.Code != "" {
		parts = append(parts, fmt.Sprintf("Code: %s", e.Code))
	}

	// Add operation ID if available
	if e.OperationID != "" {
		parts = append(parts, fmt.Sprintf("Operation: %s", e.OperationID))
	}

	return strings.Join(parts, "\n")
}

// ToCLIOutput returns a formatted text representation for terminal output with optional color support.
// When colored is true, uses ANSI color codes for better readability in terminals.
//
// Example:
//
//	err := cuserr.Validation("email", "invalid",
//	    cuserr.WithPublicMessage("Please enter a valid email address"),
//	    cuserr.WithSuggestions("Check for typos", "Ensure @ symbol is present"),
//	)
//	fmt.Fprintln(os.Stderr, err.ToCLIOutput(true))  // colored output
func (e *CustomError) ToCLIOutput(colored bool) string {
	var output strings.Builder

	// ANSI color codes
	colorReset := ""
	colorRed := ""
	colorYellow := ""
	colorBlue := ""
	colorBold := ""

	if colored {
		colorReset = "\033[0m"
		colorRed = "\033[31m"
		colorYellow = "\033[33m"
		colorBlue = "\033[34m"
		colorBold = "\033[1m"
	}

	// Error header
	output.WriteString(colorRed)
	output.WriteString(colorBold)
	output.WriteString("Error")
	output.WriteString(colorReset)

	// Add category badge
	if colored {
		categoryColor := colorYellow
		if e.Category == ErrorCategoryInternal {
			categoryColor = colorRed
		}
		output.WriteString(fmt.Sprintf(" %s[%s]%s", categoryColor, e.Category, colorReset))
	} else {
		output.WriteString(fmt.Sprintf(" [%s]", e.Category))
	}

	output.WriteString(": ")

	// Use public message if available, otherwise use developer message
	message := e.PublicMessage
	if message == "" {
		message = e.Message
	}
	output.WriteString(message)
	output.WriteString("\n")

	// Add error code
	if e.Code != "" {
		output.WriteString(fmt.Sprintf("\n%sCode:%s %s\n", colorBold, colorReset, e.Code))
	}

	// Add operation ID
	if e.OperationID != "" {
		output.WriteString(fmt.Sprintf("%sOperation ID:%s %s\n", colorBold, colorReset, e.OperationID))
	}

	// Add help URL
	if e.HelpURL != "" {
		output.WriteString(fmt.Sprintf("\n%sHelp:%s %s%s%s\n", colorBold, colorReset, colorBlue, e.HelpURL, colorReset))
	}

	// Add suggestions
	if len(e.Suggestions) > 0 {
		output.WriteString(fmt.Sprintf("\n%sSuggestions:%s\n", colorBold, colorReset))
		for i, suggestion := range e.Suggestions {
			output.WriteString(fmt.Sprintf("  %d. %s\n", i+1, suggestion))
		}
	}

	return output.String()
}

// ToSyslogSeverity maps error category to syslog severity levels (RFC 5424).
//
// Syslog Severity Levels:
//   - 0: Emergency - system is unusable
//   - 1: Alert - action must be taken immediately
//   - 2: Critical - critical conditions
//   - 3: Error - error conditions
//   - 4: Warning - warning conditions
//   - 5: Notice - normal but significant condition
//   - 6: Informational - informational messages
//   - 7: Debug - debug-level messages
//
// Example:
//
//	err := cuserr.Internal("database", dbErr)
//	syslog.Log(err.ToSyslogSeverity(), err.Message)
func (e *CustomError) ToSyslogSeverity() int {
	return CategoryToSyslogSeverity(e.Category)
}

// CategoryToSyslogSeverity maps error categories to syslog severity levels.
func CategoryToSyslogSeverity(category ErrorCategory) int {
	switch category {
	case ErrorCategoryValidation:
		return 6 // Informational - user input issue
	case ErrorCategoryUnauthorized, ErrorCategoryForbidden:
		return 5 // Notice - authentication/authorization issue
	case ErrorCategoryNotFound:
		return 4 // Warning - resource not found
	case ErrorCategoryTimeout:
		return 4 // Warning - operation timeout
	case ErrorCategoryConflict:
		return 4 // Warning - state conflict
	case ErrorCategoryRateLimit:
		return 5 // Notice - rate limiting applied
	case ErrorCategoryExternal:
		return 2 // Critical - external service down
	case ErrorCategoryInternal:
		return 3 // Error - internal error
	default:
		return 3 // Error - unknown category
	}
}

// ToGRPCCode maps error category to gRPC status codes.
// Returns standard gRPC codes as integers compatible with codes.Code from google.golang.org/grpc/codes.
//
// gRPC Code Reference:
//   - 2: Unknown - unknown error
//   - 3: InvalidArgument - client specified invalid argument
//   - 4: DeadlineExceeded - deadline expired before operation completed
//   - 5: NotFound - requested entity not found
//   - 6: AlreadyExists - entity already exists
//   - 7: PermissionDenied - caller does not have permission
//   - 8: ResourceExhausted - resource exhausted (rate limiting)
//   - 13: Internal - internal server error
//   - 14: Unavailable - service unavailable
//   - 16: Unauthenticated - request lacks valid authentication
//
// Example:
//
//	err := cuserr.NotFound("user", id)
//	grpcCode := err.ToGRPCCode()  // returns 5 (codes.NotFound)
func (e *CustomError) ToGRPCCode() int {
	return CategoryToGRPCCode(e.Category)
}

// CategoryToGRPCCode maps error categories to gRPC status codes.
// Returns integer codes compatible with codes.Code from google.golang.org/grpc/codes.
func CategoryToGRPCCode(category ErrorCategory) int {
	switch category {
	case ErrorCategoryValidation:
		return 3 // InvalidArgument
	case ErrorCategoryUnauthorized:
		return 16 // Unauthenticated
	case ErrorCategoryForbidden:
		return 7 // PermissionDenied
	case ErrorCategoryNotFound:
		return 5 // NotFound
	case ErrorCategoryTimeout:
		return 4 // DeadlineExceeded
	case ErrorCategoryConflict:
		return 6 // AlreadyExists
	case ErrorCategoryRateLimit:
		return 8 // ResourceExhausted
	case ErrorCategoryExternal:
		return 14 // Unavailable
	case ErrorCategoryInternal:
		return 13 // Internal
	default:
		return 2 // Unknown
	}
}

// ToGRPCDetails returns a map of error details suitable for gRPC status.WithDetails().
// This provides structured error information compatible with google.rpc.ErrorInfo and other well-known types.
//
// The returned map contains:
//   - code: The error code string
//   - message: Public message if available, otherwise developer message
//   - operation_id: Operation identifier for tracing
//   - trace_id: Distributed tracing trace ID
//   - span_id: Distributed tracing span ID
//   - help_url: URL to documentation/troubleshooting
//   - suggestions: Actionable resolution steps
//   - metadata: Additional context from error metadata
//
// Example:
//
//	err := cuserr.Validation("email", "invalid format",
//	    cuserr.WithPublicMessage("Please enter a valid email address"),
//	    cuserr.WithSuggestions("Check for typos", "Ensure @ symbol is present"),
//	    cuserr.WithOperationID(operationID),
//	)
//	details := err.ToGRPCDetails()
//	// Use with: status.New(...).WithDetails(proto.Message containing details)
func (e *CustomError) ToGRPCDetails() map[string]interface{} {
	details := make(map[string]interface{})

	// Add code
	if e.Code != "" {
		details["code"] = e.Code
	}

	// Use public message if available, otherwise use developer message
	message := e.PublicMessage
	if message == "" {
		message = e.Message
	}
	details["message"] = message

	// Add operation ID
	if e.OperationID != "" {
		details["operation_id"] = e.OperationID
	}

	// Add tracing information
	if e.TraceID != "" {
		details["trace_id"] = e.TraceID
	}
	if e.SpanID != "" {
		details["span_id"] = e.SpanID
	}

	// Add help URL
	if e.HelpURL != "" {
		details["help_url"] = e.HelpURL
	}

	// Add suggestions
	if len(e.Suggestions) > 0 {
		details["suggestions"] = e.Suggestions
	}

	// Add metadata
	e.mu.RLock()
	if len(e.metadata) > 0 {
		metadata := make(map[string]string)
		for k, v := range e.metadata {
			metadata[k] = v
		}
		details["metadata"] = metadata
	}
	e.mu.RUnlock()

	return details
}
