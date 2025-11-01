//go:build grpc_service
// +build grpc_service

// Package main demonstrates go-cuserr usage in gRPC services.
// This example shows how protocol-agnostic error handling works seamlessly with gRPC.
//
// Build and run:
//
//	go run -tags grpc_service examples/grpc_service.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/itsatony/go-cuserr"
)

// User represents a user in our system
type User struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
}

// UserService implements a gRPC-like user service
type UserService struct {
	users map[string]*User
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]*User),
	}
}

// GetUser retrieves a user by ID (simulates gRPC GetUser RPC)
func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
	// Extract tracing information from context
	operationID := getOperationIDFromContext(ctx)
	traceID := getTraceIDFromContext(ctx)
	spanID := getSpanIDFromContext(ctx)

	// Check if user exists
	user, exists := s.users[userID]
	if !exists {
		// Create protocol-agnostic error that works perfectly with gRPC
		return nil, cuserr.NewNotFoundError("user", userID,
			cuserr.WithOperationID(operationID),
			cuserr.WithTraceID(traceID),
			cuserr.WithSpanID(spanID),
			cuserr.WithMetadata("service", "user-service"),
			cuserr.WithMetadata("method", "GetUser"),
		)
	}

	return user, nil
}

// CreateUser creates a new user (simulates gRPC CreateUser RPC)
func (s *UserService) CreateUser(ctx context.Context, id, email, name string) (*User, error) {
	// Extract tracing information
	operationID := getOperationIDFromContext(ctx)
	traceID := getTraceIDFromContext(ctx)
	spanID := getSpanIDFromContext(ctx)

	// Validate input
	if email == "" {
		err := cuserr.NewValidationError("email", "Email is required")
		err.OperationID = operationID
		err.TraceID = traceID
		err.SpanID = spanID
		cuserr.WithMetadata("service", "user-service")(err)
		cuserr.WithMetadata("method", "CreateUser")(err)
		return nil, err
	}

	if name == "" {
		err := cuserr.NewValidationError("name", "Name is required")
		err.OperationID = operationID
		err.TraceID = traceID
		err.SpanID = spanID
		cuserr.WithMetadata("service", "user-service")(err)
		cuserr.WithMetadata("method", "CreateUser")(err)
		return nil, err
	}

	// Check for conflicts
	if _, exists := s.users[id]; exists {
		err := cuserr.NewConflictError("user", "id", id)
		err.OperationID = operationID
		err.TraceID = traceID
		err.SpanID = spanID
		cuserr.WithMetadata("service", "user-service")(err)
		cuserr.WithMetadata("method", "CreateUser")(err)
		return nil, err
	}

	// Create user
	user := &User{
		ID:        id,
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
	}
	s.users[id] = user

	return user, nil
}

// DeleteUser deletes a user (simulates gRPC DeleteUser RPC)
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	operationID := getOperationIDFromContext(ctx)
	traceID := getTraceIDFromContext(ctx)
	spanID := getSpanIDFromContext(ctx)

	if _, exists := s.users[userID]; !exists {
		return cuserr.NewNotFoundError("user", userID,
			cuserr.WithOperationID(operationID),
			cuserr.WithTraceID(traceID),
			cuserr.WithSpanID(spanID),
			cuserr.WithMetadata("service", "user-service"),
			cuserr.WithMetadata("method", "DeleteUser"),
		)
	}

	delete(s.users, userID)
	return nil
}

// SimulateDatabaseError simulates a database failure
func (s *UserService) SimulateDatabaseError(ctx context.Context) error {
	operationID := getOperationIDFromContext(ctx)
	traceID := getTraceIDFromContext(ctx)
	spanID := getSpanIDFromContext(ctx)

	// Simulate a database connection error
	return cuserr.NewExternalError("database", "query",
		fmt.Errorf("connection timeout after 30s"),
		cuserr.WithOperationID(operationID),
		cuserr.WithTraceID(traceID),
		cuserr.WithSpanID(spanID),
		cuserr.WithMetadata("service", "user-service"),
		cuserr.WithMetadata("method", "SimulateDatabaseError"),
		cuserr.WithMetadata("database", "postgres"),
		cuserr.WithMetadata("host", "db.example.com"),
		cuserr.WithMetadata("timeout_ms", "30000"),
	)
}

// Context helper functions (simulate extracting from gRPC metadata)
func getOperationIDFromContext(ctx context.Context) string {
	if opID := ctx.Value("operation_id"); opID != nil {
		return opID.(string)
	}
	return fmt.Sprintf("grpc-op-%d", time.Now().UnixNano())
}

func getTraceIDFromContext(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		return traceID.(string)
	}
	return ""
}

func getSpanIDFromContext(ctx context.Context) string {
	if spanID := ctx.Value("span_id"); spanID != nil {
		return spanID.(string)
	}
	return ""
}

// handleGRPCError demonstrates how to handle cuserr errors in gRPC
func handleGRPCError(err error) {
	if err == nil {
		return
	}

	customErr, ok := err.(*cuserr.CustomError)
	if !ok {
		log.Printf("Non-cuserr error: %v\n", err)
		return
	}

	// In a real gRPC service, you would convert this to a status.Status
	// using google.golang.org/grpc/status
	fmt.Println("\n--- gRPC Error Response ---")
	fmt.Printf("gRPC Code: %d (%s)\n", customErr.ToGRPCCode(), grpcCodeToString(customErr.ToGRPCCode()))
	fmt.Printf("Message: %s\n", customErr.Message)
	fmt.Printf("Operation ID: %s\n", customErr.OperationID)
	fmt.Printf("Trace ID: %s\n", customErr.TraceID)
	fmt.Printf("Span ID: %s\n", customErr.SpanID)

	// Show gRPC Details (for rich error responses)
	details := customErr.ToGRPCDetails()
	fmt.Printf("gRPC Details: %+v\n", details)

	// Show how it would look as JSON (for logging/debugging)
	fmt.Printf("\nJSON for logging:\n%s\n", customErr.ToJSONString())
}

// grpcCodeToString converts gRPC codes to human-readable strings
func grpcCodeToString(code int) string {
	codes := map[int]string{
		0:  "OK",
		1:  "CANCELLED",
		2:  "UNKNOWN",
		3:  "INVALID_ARGUMENT",
		4:  "DEADLINE_EXCEEDED",
		5:  "NOT_FOUND",
		6:  "ALREADY_EXISTS",
		7:  "PERMISSION_DENIED",
		8:  "RESOURCE_EXHAUSTED",
		9:  "FAILED_PRECONDITION",
		10: "ABORTED",
		11: "OUT_OF_RANGE",
		12: "UNIMPLEMENTED",
		13: "INTERNAL",
		14: "UNAVAILABLE",
		15: "DATA_LOSS",
		16: "UNAUTHENTICATED",
	}
	if name, exists := codes[code]; exists {
		return name
	}
	return "UNKNOWN"
}

func main() {
	fmt.Println("=== gRPC Service Example: go-cuserr Protocol-Agnostic Error Handling ===\n")

	// Create service
	service := NewUserService()

	// Create context with tracing information (simulates gRPC metadata)
	ctx := context.Background()
	ctx = context.WithValue(ctx, "operation_id", "grpc-123-456")
	ctx = context.WithValue(ctx, "trace_id", "trace-abc-def")
	ctx = context.WithValue(ctx, "span_id", "span-xyz-789")

	// Example 1: Successful user creation
	fmt.Println("Example 1: CreateUser RPC - Success")
	user, err := service.CreateUser(ctx, "user-1", "alice@example.com", "Alice Smith")
	if err != nil {
		handleGRPCError(err)
		return
	}
	fmt.Printf("✓ User created: %s (%s)\n", user.Name, user.Email)

	// Example 2: Not Found error
	fmt.Println("\nExample 2: GetUser RPC - Not Found")
	_, err = service.GetUser(ctx, "user-999")
	handleGRPCError(err)

	// Example 3: Validation error
	fmt.Println("\nExample 3: CreateUser RPC - Validation Error")
	_, err = service.CreateUser(ctx, "user-2", "", "Bob Jones")
	handleGRPCError(err)

	// Example 4: Conflict error
	fmt.Println("\nExample 4: CreateUser RPC - Conflict")
	_, err = service.CreateUser(ctx, "user-1", "duplicate@example.com", "Duplicate User")
	handleGRPCError(err)

	// Example 5: External service error
	fmt.Println("\nExample 5: Database Error - External Dependency")
	err = service.SimulateDatabaseError(ctx)
	handleGRPCError(err)

	// Example 6: Protocol-agnostic demonstration
	fmt.Println("\n=== Protocol-Agnostic Nature ===")
	fmt.Println("The same error works across all protocols:\n")

	// Create a sample error
	err = cuserr.NewUnauthorizedError("Invalid API token")
	customErr := err.(*cuserr.CustomError)
	customErr.OperationID = "multi-protocol-demo"
	customErr.TraceID = "trace-multi-123"
	customErr.SpanID = "span-multi-456"
	cuserr.WithMetadata("service", "auth-service")(customErr)

	fmt.Printf("HTTP Status Code:    %d (Unauthorized)\n", customErr.ToHTTPStatus())
	fmt.Printf("gRPC Code:          %d (%s)\n", customErr.ToGRPCCode(), grpcCodeToString(customErr.ToGRPCCode()))
	fmt.Printf("CLI Exit Code:      %d\n", customErr.ToExitCode())
	fmt.Printf("Syslog Severity:    %d (Error)\n", customErr.ToSyslogSeverity())

	// Example 7: Rich error details for gRPC
	fmt.Println("\n=== Rich gRPC Error Details ===")
	err = cuserr.NewValidationError("user_input", "Multiple validation failures")
	customErr2 := err.(*cuserr.CustomError)
	customErr2.OperationID = "validate-123"
	cuserr.WithMetadata("field", "email")(customErr2)
	cuserr.WithMetadata("value", "invalid@")(customErr2)

	collection := cuserr.NewValidationErrorCollection()
	collection.AddValidationWithCode("email", "Invalid email format", "INVALID_EMAIL")
	collection.AddValidationWithCode("age", "Must be at least 18", "AGE_TOO_LOW")
	collection.AddValidationWithCode("username", "Already taken", "USERNAME_CONFLICT")

	collectionErr := collection.ToCustomError()
	collectionErr.OperationID = "validate-collection-123"
	collectionErr.TraceID = "trace-validation-xyz"

	fmt.Println("\nValidation Collection Error:")
	handleGRPCError(collectionErr)

	// Example 8: Production mode behavior
	fmt.Println("\n=== Production Mode (Sanitized Errors) ===")
	cuserr.SetConfig(&cuserr.Config{
		EnableStackTrace: false,
		MaxStackDepth:    cuserr.DEFAULT_STACK_DEPTH,
		ProductionMode:   true,
	})

	err = cuserr.NewInternalError("payment-service", nil,
		cuserr.WithOperationID("payment-789"),
		cuserr.WithMetadata("sensitive_data", "should-be-filtered"),
		cuserr.WithMetadata("user_id", "usr-123"), // Safe ID kept
	)

	fmt.Println("\nInternal error in production mode:")
	customErr3 := err.(*cuserr.CustomError)
	publicJSON := customErr3.ToPublicJSON()
	fmt.Printf("Public JSON (safe for clients): %+v\n", publicJSON)

	// Reset to development mode
	cuserr.SetConfig(&cuserr.Config{
		EnableStackTrace: true,
		MaxStackDepth:    cuserr.DEFAULT_STACK_DEPTH,
		ProductionMode:   false,
	})

	// Example 9: Error chaining in microservices
	fmt.Println("\n=== Error Chaining Across Microservices ===")

	// Simulate error from downstream service
	downstreamErr := cuserr.NewExternalError("inventory-service", "check_stock",
		fmt.Errorf("HTTP 503: Service Temporarily Unavailable"),
		cuserr.WithOperationID("check-stock-001"),
		cuserr.WithTraceID("trace-chain-abc"),
		cuserr.WithMetadata("downstream_service", "inventory-service"),
	)

	// Wrap it in current service's error
	orderErr := cuserr.NewCustomError(
		cuserr.ErrExternal,
		downstreamErr,
		"Failed to create order: inventory check failed",
	)
	orderErr.OperationID = "create-order-123"
	orderErr.TraceID = "trace-chain-abc"
	orderErr.SpanID = "span-order-service"
	cuserr.WithMetadata("service", "order-service")(orderErr)
	cuserr.WithMetadata("order_id", "ord-789")(orderErr)

	fmt.Println("\nChained error (preserves full context):")
	fmt.Printf("Error: %s\n", orderErr.Error())
	fmt.Printf("Wrapped error: %v\n", orderErr.Wrapped)
	fmt.Printf("Operation ID: %s\n", orderErr.OperationID)
	fmt.Printf("Trace ID: %s (spans multiple services)\n", orderErr.TraceID)

	fmt.Println("\n=== Summary ===")
	fmt.Println("✓ Same error type works across HTTP, gRPC, CLI, and logging")
	fmt.Println("✓ Automatic mapping to appropriate gRPC status codes")
	fmt.Println("✓ Rich error details for structured error responses")
	fmt.Println("✓ Operation/Trace/Span IDs for distributed tracing")
	fmt.Println("✓ Production mode automatically sanitizes sensitive data")
	fmt.Println("✓ Error chaining preserves context across microservices")
	fmt.Println("✓ Protocol-agnostic design = write once, use everywhere")
}
