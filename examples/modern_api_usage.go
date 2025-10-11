// Package main demonstrates modern go-cuserr usage with logger injection and options pattern.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/itsatony/go-cuserr"
	"github.com/itsatony/go-cuserr/adapters"
)

func main() {
	fmt.Println("=== Modern go-cuserr API Demo ===")
	fmt.Println()

	// Example 1: Simple error without logger (library-friendly)
	example1SimpleError()

	// Example 2: Error with logger injection
	example2WithLogger()

	// Example 3: Multiple options (context + logger + metadata)
	example3MultipleOptions()

	// Example 4: Service-layer pattern
	example4ServicePattern()
}

// Example 1: Simple error creation (no logging required)
func example1SimpleError() {
	fmt.Println("Example 1: Simple Error (Library-Friendly)")
	fmt.Println("--------------------------------------------")

	// Create a simple error - no logger required
	// This makes the library usable in any context
	err := cuserr.NewNotFoundError("user", "usr_123")

	fmt.Printf("Error: %s\n", err.Error())
	fmt.Printf("Category: %s\n", err.Category)
	fmt.Printf("HTTP Status: %d\n", err.ToHTTPStatus())
	fmt.Printf("Request ID: %s\n\n", err.RequestID)
}

// Example 2: With logger injection
func example2WithLogger() {
	fmt.Println("Example 2: Error with Logger Injection")
	fmt.Println("---------------------------------------")

	// Create a structured logger (slog in this case)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	adapter := adapters.NewSlogAdapter(logger)

	// Create error with logger option
	err := cuserr.NewInternalError("database", nil,
		cuserr.WithLogger(adapter),
		cuserr.WithMetadata("query", "SELECT * FROM users"),
		cuserr.WithMetadata("duration_ms", "250"),
	)

	// Log the error
	ctx := context.Background()
	err.Log(ctx)

	fmt.Printf("\n")
}

// Example 3: Multiple options (realistic scenario)
func example3MultipleOptions() {
	fmt.Println("Example 3: Multiple Options (Realistic Scenario)")
	fmt.Println("------------------------------------------------")

	// Simulate request context
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req-abc-123")
	ctx = context.WithValue(ctx, "user_id", "usr-456")

	// Create logger
	logger := adapters.NewSlogAdapter(slog.Default())

	// Create error with all the bells and whistles
	err := cuserr.NewExternalError("payment-service", "charge", fmt.Errorf("connection timeout"),
		cuserr.WithContext(ctx),      // Extracts request_id, user_id, trace_id
		cuserr.WithLogger(logger),     // Enables logging
		cuserr.WithMetadata("amount", "99.99"),
		cuserr.WithMetadata("currency", "USD"),
		cuserr.WithMetadata("merchant_id", "merch-789"),
	)

	fmt.Printf("Error created with context and logger\n")
	fmt.Printf("Request ID: %s\n", err.RequestID)
	if userID, ok := err.GetMetadata("user_id"); ok {
		fmt.Printf("User ID: %s\n", userID)
	}
	if amount, ok := err.GetMetadata("amount"); ok {
		if currency, ok := err.GetMetadata("currency"); ok {
			fmt.Printf("Amount: %s %s\n", amount, currency)
		}
	}

	// Log it
	err.Log(context.Background())
	fmt.Printf("\n")
}

// Example 4: Typical service layer pattern
func example4ServicePattern() {
	fmt.Println("Example 4: Service Layer Pattern")
	fmt.Println("---------------------------------")

	// Initialize service with logger
	logger := adapters.NewSlogAdapter(slog.Default())
	service := &UserService{logger: logger}

	// Simulate request
	ctx := context.WithValue(context.Background(), "request_id", "req-xyz-999")

	// Call service method
	user, err := service.GetUser(ctx, "usr_404")
	if err != nil {
		fmt.Printf("Service returned error: %s\n", err)

		// Check error type
		if customErr, ok := err.(*cuserr.CustomError); ok {
			fmt.Printf("HTTP Status would be: %d\n", customErr.ToHTTPStatus())
			fmt.Printf("Request ID: %s\n", customErr.RequestID)
		}
	} else {
		fmt.Printf("User: %+v\n", user)
	}

	fmt.Printf("\n")
}

// UserService demonstrates typical service layer usage
type UserService struct {
	logger cuserr.Logger
}

// User represents a user entity
type User struct {
	ID    string
	Email string
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
	// Simulate database lookup
	user := s.findInDatabase(userID)

	if user == nil {
		// Create error with context and logger
		err := cuserr.NewNotFoundError("user", userID,
			cuserr.WithContext(ctx),
			cuserr.WithLogger(s.logger),
			cuserr.WithMetadata("attempted_lookup", "database"),
		)

		// Log the error (happens automatically because logger is configured)
		err.Log(ctx)

		return nil, err
	}

	return user, nil
}

// findInDatabase simulates a database lookup
func (s *UserService) findInDatabase(userID string) *User {
	// Simulate not found
	return nil
}
