//go:build cli_tool
// +build cli_tool

// Package main demonstrates go-cuserr usage in CLI applications.
// This example shows how to use protocol-agnostic error handling for terminal applications.
//
// Build and run:
//
//	go run -tags cli_tool examples/cli_tool.go
package main

import (
	"fmt"
	"os"

	"github.com/itsatony/go-cuserr"
)

// User represents a user in our system
type User struct {
	ID    string
	Email string
	Age   int
}

// UserValidator validates user input
type UserValidator struct{}

// ValidateUser validates a user and returns appropriate errors
func (v *UserValidator) ValidateUser(email string, age int) error {
	// Create a validation error collection
	collection := cuserr.NewValidationErrorCollection()

	// Validate email
	if email == "" {
		collection.AddValidation("email", "Email is required")
	} else if !isValidEmail(email) {
		collection.AddValidationWithCode("email", "Invalid email format", "INVALID_EMAIL")
	}

	// Validate age
	if age < 18 {
		collection.AddValidationWithCode("age", "Must be at least 18 years old", "AGE_TOO_LOW")
	} else if age > 150 {
		collection.AddValidationWithCode("age", "Age seems unrealistic", "AGE_TOO_HIGH")
	}

	// Return if we have validation errors
	if collection.HasErrors() {
		return collection.ToCustomError()
	}

	return nil
}

// Simple email validation
func isValidEmail(email string) bool {
	return len(email) > 3 && email[0] != '@' && email[len(email)-1] != '@'
}

// UserService handles user operations
type UserService struct {
	users map[string]*User
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]*User),
	}
}

// CreateUser creates a new user with validation
func (s *UserService) CreateUser(id, email string, age int) error {
	// Validate input
	validator := &UserValidator{}
	if err := validator.ValidateUser(email, age); err != nil {
		// Enhance with public-facing message and suggestions
		if customErr, ok := err.(*cuserr.CustomError); ok {
			cuserr.WithPublicMessage("Please check your input and try again")(customErr)
			cuserr.WithSuggestions(
				"Ensure email is in correct format (user@domain.com)",
				"Age must be between 18 and 150",
			)(customErr)
			cuserr.WithHelpURL("https://docs.example.com/users/create")(customErr)
			cuserr.WithOperationID(fmt.Sprintf("create-user-%s", id))(customErr)
		}
		return err
	}

	// Check if user already exists
	if _, exists := s.users[id]; exists {
		err := cuserr.NewConflictError("user", "id", id)
		cuserr.WithPublicMessage("A user with this ID already exists")(err)
		cuserr.WithSuggestions("Use a different user ID", "Update the existing user instead")(err)
		cuserr.WithOperationID(fmt.Sprintf("create-user-%s", id))(err)
		return err
	}

	// Create user
	s.users[id] = &User{
		ID:    id,
		Email: email,
		Age:   age,
	}

	return nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id string) (*User, error) {
	user, exists := s.users[id]
	if !exists {
		return nil, cuserr.NewNotFoundError("user", id,
			cuserr.WithPublicMessage(fmt.Sprintf("User '%s' not found", id)),
			cuserr.WithSuggestions("Check the user ID for typos", "Create the user first"),
			cuserr.WithOperationID(fmt.Sprintf("get-user-%s", id)),
		)
	}
	return user, nil
}

// handleError processes errors and outputs them appropriately for CLI
func handleError(err error) {
	if err == nil {
		return
	}

	// Check if it's a CustomError
	customErr, ok := err.(*cuserr.CustomError)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
		return
	}

	// Determine if terminal supports colors
	supportsColor := os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb"

	// Output formatted error
	fmt.Fprintf(os.Stderr, "%s\n", customErr.ToCLIOutput(supportsColor))

	// Exit with appropriate code
	os.Exit(customErr.ToExitCode())
}

// handleErrorSimple demonstrates simple text output without colors
func handleErrorSimple(err error) {
	if err == nil {
		return
	}

	customErr, ok := err.(*cuserr.CustomError)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
		return
	}

	// Simple text output
	fmt.Fprintf(os.Stderr, "%s\n", customErr.ToText())
	os.Exit(customErr.ToExitCode())
}

func main() {
	fmt.Println("=== CLI Tool Example: go-cuserr Protocol-Agnostic Error Handling ===\n")

	// Create service
	service := NewUserService()

	// Example 1: Successful user creation
	fmt.Println("Example 1: Creating a valid user...")
	err := service.CreateUser("user-1", "alice@example.com", 25)
	if err != nil {
		handleError(err)
	}
	fmt.Println("✓ User created successfully!\n")

	// Example 2: Validation error - invalid email
	fmt.Println("Example 2: Creating user with invalid email...")
	err = service.CreateUser("user-2", "invalid-email", 30)
	if err != nil {
		customErr := err.(*cuserr.CustomError)
		fmt.Printf("Caught validation error: %s\n", customErr.GetPublicMessage())
		fmt.Printf("Exit code would be: %d\n", customErr.ToExitCode())
		fmt.Printf("Error as text:\n%s\n\n", customErr.ToText())
	}

	// Example 3: Validation error - age too low
	fmt.Println("Example 3: Creating user with invalid age...")
	err = service.CreateUser("user-3", "bob@example.com", 15)
	if err != nil {
		customErr := err.(*cuserr.CustomError)
		fmt.Printf("Caught validation error: %s\n", customErr.GetPublicMessage())
		fmt.Printf("Suggestions:\n")
		for i, suggestion := range customErr.Suggestions {
			fmt.Printf("  %d. %s\n", i+1, suggestion)
		}
		fmt.Println()
	}

	// Example 4: Conflict error - duplicate user
	fmt.Println("Example 4: Creating duplicate user...")
	err = service.CreateUser("user-1", "duplicate@example.com", 28)
	if err != nil {
		customErr := err.(*cuserr.CustomError)
		fmt.Printf("Caught conflict error: %s\n", customErr.GetPublicMessage())
		fmt.Printf("HTTP status would be: %d\n", customErr.ToHTTPStatus())
		fmt.Printf("Exit code would be: %d\n", customErr.ToExitCode())
		fmt.Printf("gRPC code would be: %d\n", customErr.ToGRPCCode())
		fmt.Printf("Syslog severity would be: %d\n\n", customErr.ToSyslogSeverity())
	}

	// Example 5: Not found error
	fmt.Println("Example 5: Getting non-existent user...")
	_, err = service.GetUser("user-999")
	if err != nil {
		customErr := err.(*cuserr.CustomError)
		fmt.Printf("Caught not found error: %s\n", customErr.GetPublicMessage())
		fmt.Printf("Operation ID: %s\n", customErr.OperationID)
		fmt.Println()
	}

	// Example 6: Demonstrating colored output
	fmt.Println("Example 6: Colored CLI output (if terminal supports it)...")
	err = cuserr.NewValidationError("username", "must be at least 3 characters")
	cuserr.WithPublicMessage("Username validation failed")(err.(*cuserr.CustomError))
	cuserr.WithSuggestions(
		"Use at least 3 characters",
		"Only use letters, numbers, and underscores",
		"Choose a unique username",
	)(err.(*cuserr.CustomError))
	cuserr.WithHelpURL("https://docs.example.com/validation#username")(err.(*cuserr.CustomError))

	customErr := err.(*cuserr.CustomError)
	fmt.Println("--- Colored output ---")
	fmt.Println(customErr.ToCLIOutput(true))

	fmt.Println("--- Plain text output ---")
	fmt.Println(customErr.ToText())
	fmt.Println()

	// Example 7: Protocol-agnostic error - works everywhere
	fmt.Println("Example 7: Protocol-agnostic error showcasing multi-protocol support...")
	err = cuserr.NewInternalError("database", nil,
		cuserr.WithPublicMessage("A temporary issue occurred"),
		cuserr.WithHelpURL("https://status.example.com"),
		cuserr.WithSuggestions("Try again in a few moments", "Contact support if issue persists"),
		cuserr.WithOperationID("op-12345"),
		cuserr.WithTraceID("trace-67890"),
		cuserr.WithSpanID("span-abcde"),
	)

	customErr = err.(*cuserr.CustomError)
	fmt.Printf("HTTP Status: %d\n", customErr.ToHTTPStatus())
	fmt.Printf("CLI Exit Code: %d\n", customErr.ToExitCode())
	fmt.Printf("gRPC Code: %d\n", customErr.ToGRPCCode())
	fmt.Printf("Syslog Severity: %d\n", customErr.ToSyslogSeverity())
	fmt.Printf("Public Message: %s\n", customErr.GetPublicMessage())
	fmt.Printf("Operation ID: %s\n", customErr.OperationID)
	fmt.Printf("Trace ID: %s\n", customErr.TraceID)
	fmt.Printf("Span ID: %s\n\n", customErr.SpanID)

	// Example 8: Production mode behavior
	fmt.Println("Example 8: Production mode (hides sensitive details)...")
	cuserr.SetConfig(&cuserr.Config{
		EnableStackTrace: false,
		MaxStackDepth:    cuserr.DEFAULT_STACK_DEPTH,
		ProductionMode:   true,
	})

	err = cuserr.NewInternalError("auth-service", nil)
	customErr = err.(*cuserr.CustomError)
	fmt.Printf("Production-safe message: %s\n", customErr.GetPublicMessage())
	fmt.Printf("Developer message (internal): %s\n\n", customErr.Message)

	// Reset to development mode
	cuserr.SetConfig(&cuserr.Config{
		EnableStackTrace: true,
		MaxStackDepth:    cuserr.DEFAULT_STACK_DEPTH,
		ProductionMode:   false,
	})

	fmt.Println("=== All examples completed successfully! ===")
	fmt.Println("\nKey Takeaways:")
	fmt.Println("✓ Same error works across HTTP, CLI, gRPC, and syslog")
	fmt.Println("✓ Public messages keep sensitive details private")
	fmt.Println("✓ Suggestions guide users to resolution")
	fmt.Println("✓ Operation/Trace IDs enable distributed tracing")
	fmt.Println("✓ Production mode automatically sanitizes errors")
	fmt.Println("✓ Exit codes follow Unix conventions")
}
