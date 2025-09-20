// Package main demonstrates the enhanced features of go-cuserr v0.2.0+
// This file contains comprehensive examples of the new v0.2.0 features
// To run: go run enhanced_usage.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/itsatony/go-cuserr"
)

func mainEnhanced() {
	fmt.Println("=== Enhanced go-cuserr Usage Examples ===")

	// Set up structured logging
	logger := cuserr.NewDefaultSlogLogger(nil)
	cuserr.SetStructuredLogger(logger)

	demonstrateConvenienceConstructors()
	demonstrateContextBasedErrors()
	demonstrateErrorAggregation()
	demonstrateTypedMetadata()
	demonstrateStructuredLogging()
	demonstrateMigrationHelpers()
	demonstrateErrorBuilders()
}

// main runs the enhanced examples demonstration
func main() {
	mainEnhanced()
}

// demonstrateConvenienceConstructors shows the new convenience constructor functions
func demonstrateConvenienceConstructors() {
	fmt.Println("1. Convenience Constructors")
	fmt.Println("===========================")

	// Validation errors with field context
	validationErr := cuserr.NewValidationError("email", "invalid email format")
	fmt.Printf("Validation Error: %s\n", validationErr.Error())
	fmt.Printf("  Category: %s\n", validationErr.Category)
	if field, exists := validationErr.GetMetadata("field"); exists {
		fmt.Printf("  Field: %s\n", field)
	}

	// Not found errors with resource context
	notFoundErr := cuserr.NewNotFoundError("user", "usr_12345")
	fmt.Printf("\nNot Found Error: %s\n", notFoundErr.Error())
	if resource, exists := notFoundErr.GetMetadata("resource"); exists {
		fmt.Printf("  Resource: %s\n", resource)
	}

	// Internal errors with component context
	internalErr := cuserr.NewInternalError("database", fmt.Errorf("connection timeout"))
	fmt.Printf("\nInternal Error: %s\n", internalErr.Error())
	if component, exists := internalErr.GetMetadata("component"); exists {
		fmt.Printf("  Component: %s\n", component)
	}

	fmt.Println()
}

// demonstrateContextBasedErrors shows context-aware error creation
func demonstrateContextBasedErrors() {
	fmt.Println("2. Context-Based Error Creation")
	fmt.Println("================================")

	// Create context with request information
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req-ctx-example")
	ctx = context.WithValue(ctx, "user_id", "usr_ctx_demo")
	ctx = context.WithValue(ctx, "trace_id", "trace_abc123")

	// Create error from context - automatically extracts metadata
	err := cuserr.NewValidationErrorFromContext(ctx, "age", "must be positive")
	fmt.Printf("Context Error: %s\n", err.Error())
	fmt.Printf("  Request ID: %s\n", err.RequestID)
	if userID, exists := err.GetMetadata("user_id"); exists {
		fmt.Printf("  User ID: %s\n", userID)
	}
	if traceID, exists := err.GetMetadata("trace_id"); exists {
		fmt.Printf("  Trace ID: %s\n", traceID)
	}

	// Context-based configuration
	prodCtx := cuserr.WithProductionMode(ctx)
	devCtx := cuserr.WithDevelopmentMode(ctx)

	prodConfig := cuserr.GetConfigFromContext(prodCtx)
	devConfig := cuserr.GetConfigFromContext(devCtx)

	fmt.Printf("\nProduction mode from context: %t\n", prodConfig.ProductionMode)
	fmt.Printf("Development mode stack traces: %t\n", devConfig.EnableStackTrace)

	fmt.Println()
}

// demonstrateErrorAggregation shows error collection and aggregation
func demonstrateErrorAggregation() {
	fmt.Println("3. Error Aggregation")
	fmt.Println("====================")

	// Create a validation error collection
	collection := cuserr.NewValidationErrorCollection()
	collection.WithRequestID("req-validation-demo")

	// Add multiple validation errors
	collection.AddValidation("email", "invalid format")
	collection.AddValidation("password", "too short")
	collection.AddValidationWithCode("age", "must be 18 or older", "AGE_RESTRICTION")
	collection.AddValidationWithValue("username", "already taken", "johndoe")

	fmt.Printf("Error Collection Summary: %s\n", collection.Error())
	fmt.Printf("  Total errors: %d\n", collection.Count())
	fmt.Printf("  Validation errors: %d\n", collection.ValidationCount())
	fmt.Printf("  Affected fields: %v\n", collection.GetFields())

	// Convert to single CustomError for easier handling
	aggregatedErr := collection.ToCustomError()
	fmt.Printf("\nAggregated Error: %s\n", aggregatedErr.Error())

	// Builder pattern for collections
	ctx := context.WithValue(context.Background(), "request_id", "req-builder-demo")
	builderCollection := cuserr.NewValidationCollectionBuilder().
		WithContext(ctx).
		AddValidation("name", "required field").
		AddValidation("phone", "invalid format").
		Build()

	fmt.Printf("\nBuilder Collection: %s\n", builderCollection.Error())
	fmt.Printf("  Request ID: %s\n", builderCollection.RequestID)

	fmt.Println()
}

// demonstrateTypedMetadata shows type-safe metadata operations
func demonstrateTypedMetadata() {
	fmt.Println("4. Typed Metadata")
	fmt.Println("=================")

	// Create error with typed metadata
	err := cuserr.NewExternalError("payment-api", "charge", fmt.Errorf("service unavailable"))
	tm := err.GetTypedMetadata()

	// Set various typed metadata fields
	tm.WithUserID("usr_typed_demo").
		WithHTTPMethod("POST").
		WithURL("https://api.payment.com/charge").
		WithStatusCode(503).
		WithResponseTime(5 * time.Second).
		WithRetryCount(3).
		WithExternalService("stripe").
		WithOperation("credit_card_charge")

	fmt.Printf("External Service Error: %s\n", err.Error())

	// Retrieve typed metadata
	if userID, exists := tm.GetUserID(); exists {
		fmt.Printf("  User ID: %s\n", userID)
	}
	if method, exists := tm.GetHTTPMethod(); exists {
		fmt.Printf("  HTTP Method: %s\n", method)
	}
	if statusCode, exists := tm.GetStatusCode(); exists {
		fmt.Printf("  Status Code: %d\n", statusCode)
	}
	if responseTime, exists := tm.GetResponseTime(); exists {
		fmt.Printf("  Response Time: %v\n", responseTime)
	}
	if retryCount, exists := tm.GetRetryCount(); exists {
		fmt.Printf("  Retry Count: %d\n", retryCount)
	}

	// Convenience constructors with typed metadata
	validationErr, validationTM := cuserr.NewValidationErrorWithTypedMetadata("email", "invalid format")
	validationTM.WithValidationRule("email_format").WithValidationValue("invalid@")

	fmt.Printf("\nValidation Error: %s\n", validationErr.Error())
	if rule, exists := validationTM.GetValidationRule(); exists {
		fmt.Printf("  Validation Rule: %s\n", rule)
	}

	fmt.Println()
}

// demonstrateStructuredLogging shows integration with structured logging
func demonstrateStructuredLogging() {
	fmt.Println("5. Structured Logging")
	fmt.Println("=====================")

	// Create error with rich context
	err := cuserr.NewInternalError("auth-service", fmt.Errorf("redis connection failed")).
		WithRequestID("req-logging-demo").
		WithMetadata("user_id", "usr_log_demo").
		WithMetadata("operation", "user_login").
		WithMetadata("retry_count", "2")

	// Get structured log fields
	logFields := err.ToLogFields()
	fmt.Printf("Structured log fields for error:\n")
	for key, value := range logFields {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Log error directly (this would normally go to your structured logger)
	ctx := context.Background()
	err.LogWith(ctx, cuserr.GetStructuredLogger())

	// Error collection logging
	collection := cuserr.NewValidationErrorCollection()
	collection.AddValidation("field1", "error1")
	collection.AddValidation("field2", "error2")
	collection.WithRequestID("req-collection-log")

	collectionFields := collection.ToLogFields()
	fmt.Printf("\nError collection log fields:\n")
	for key, value := range collectionFields {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println()
}

// demonstrateMigrationHelpers shows utilities for migrating from other error systems
func demonstrateMigrationHelpers() {
	fmt.Println("6. Migration Helpers")
	fmt.Println("====================")

	// Migrate from standard library errors
	stdErr := fmt.Errorf("user not found in database")
	migratedErr := cuserr.FromStdError(stdErr, "failed to retrieve user")
	fmt.Printf("Migrated stdlib error: %s\n", migratedErr.Error())
	fmt.Printf("  Category: %s\n", migratedErr.Category)
	if migratedFrom, exists := migratedErr.GetMetadata("migrated_from"); exists {
		fmt.Printf("  Migrated from: %s\n", migratedFrom)
	}

	// Migrate from HTTP status codes
	httpErr := cuserr.FromHTTPStatus(404, "resource not found")
	fmt.Printf("\nHTTP status error: %s\n", httpErr.Error())
	if statusCode, exists := httpErr.GetMetadata("original_status_code"); exists {
		fmt.Printf("  Original status: %s\n", statusCode)
	}

	// Migrate from SQL errors
	sqlErr := fmt.Errorf("duplicate key constraint violation")
	dbErr := cuserr.FromSQLError(sqlErr, "INSERT INTO users (email) VALUES (?)")
	fmt.Printf("\nSQL error: %s\n", dbErr.Error())
	fmt.Printf("  Category: %s\n", dbErr.Category)

	// Batch migration
	errors := []error{
		fmt.Errorf("timeout occurred"),
		fmt.Errorf("permission denied"),
		fmt.Errorf("invalid input data"),
	}

	migrationCollection, report := cuserr.BatchMigrate(errors)
	fmt.Printf("\nBatch migration results:\n")
	fmt.Printf("  %s\n", report.Summary())
	fmt.Printf("  Collection has %d errors\n", migrationCollection.Count())

	fmt.Println()
}

// demonstrateErrorBuilders shows the fluent builder patterns
func demonstrateErrorBuilders() {
	fmt.Println("7. Error Builders")
	fmt.Println("=================")

	// Basic error builder
	err1 := cuserr.NewErrorBuilder(cuserr.ErrInvalidInput).
		WithMessage("validation failed for user input").
		WithMetadata("component", "user-validator").
		WithMetadata("validation_type", "email_format").
		WithRequestID("req-builder-001").
		Build()

	fmt.Printf("Built error: %s\n", err1.Error())
	fmt.Printf("  Category: %s\n", err1.Category)
	fmt.Printf("  Request ID: %s\n", err1.RequestID)

	// Contextual error builder
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user_id", "usr_builder_demo")
	ctx = context.WithValue(ctx, "request_id", "req-contextual-builder")

	err2 := cuserr.NewContextualErrorBuilder(ctx, cuserr.ErrForbidden).
		WithMessage("insufficient permissions").
		WithMetadata("required_role", "admin").
		WithMetadata("user_role", "user").
		Build()

	fmt.Printf("\nContextual built error: %s\n", err2.Error())
	fmt.Printf("  Category: %s\n", err2.Category)
	if requiredRole, exists := err2.GetMetadata("required_role"); exists {
		fmt.Printf("  Required role: %s\n", requiredRole)
	}

	fmt.Println()
}

// Enhanced middleware example showing real-world usage
func enhancedErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "request_id", generateRequestID())
		ctx = context.WithValue(ctx, "method", r.Method)
		ctx = context.WithValue(ctx, "path", r.URL.Path)
		ctx = cuserr.WithAutoErrorLogging(ctx) // Automatically log errors

		// Capture panics with rich context
		defer func() {
			if recovered := recover(); recovered != nil {
				panicErr := cuserr.NewInternalErrorFromContext(ctx, "panic_handler", fmt.Errorf("panic: %v", recovered))

				// Add typed metadata
				tm := panicErr.GetTypedMetadata()
				tm.WithHTTPMethod(r.Method).
					WithEndpoint(r.URL.Path).
					WithUserAgent(r.Header.Get("User-Agent")).
					WithIPAddress(getEnhancedClientIP(r))

				// Write structured error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(panicErr.ToHTTPStatus())

				// Use client-safe JSON in production
				response := panicErr.ToClientJSON()
				if err := json.NewEncoder(w).Encode(response); err != nil {
					log.Printf("Failed to encode error response: %v", err)
				}
			}
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper functions for the enhanced middleware
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func getEnhancedClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// Example of using the enhanced features in a web service
func exampleAPIHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validate input with error collection
	collection := cuserr.NewValidationCollectionBuilder().WithContext(ctx)

	email := r.FormValue("email")
	if email == "" {
		collection.AddValidation("email", "email is required")
	} else if !isValidEmail(email) {
		collection.AddValidationWithCode("email", "invalid email format", "INVALID_EMAIL_FORMAT")
	}

	password := r.FormValue("password")
	if len(password) < 8 {
		collection.AddValidation("password", "password must be at least 8 characters")
	}

	// Check if we have validation errors
	if validationErrors := collection.BuildIfHasErrors(); validationErrors != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(validationErrors.ToHTTPStatus())
		json.NewEncoder(w).Encode(validationErrors.ToClientJSON())
		return
	}

	// Simulate business logic that might fail
	if err := processUser(ctx, email, password); err != nil {
		// Error is already enriched with context and properly categorized
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(err.ToHTTPStatus())
		json.NewEncoder(w).Encode(err.ToClientJSON())
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Simulated business logic with enhanced error handling
func processUser(ctx context.Context, email, password string) *cuserr.CustomError {
	// Check if user exists (simulated)
	if email == "exists@example.com" {
		return cuserr.NewConflictError("user", "email", email).
			GetTypedMetadata().
			WithValidationField("email").
			WithValidationValue(email).
			Error()
	}

	// Simulate external service call
	if err := callExternalService(ctx, email); err != nil {
		// Wrap external error with context
		return cuserr.NewExternalErrorFromContext(ctx, "user-service", "create_user", err)
	}

	// Simulate database error
	if email == "db-error@example.com" {
		dbErr := fmt.Errorf("connection timeout")
		return cuserr.NewInternalErrorFromContext(ctx, "database", dbErr).
			GetTypedMetadata().
			WithOperation("user_creation").
			WithRetryCount(3).
			Error()
	}

	return nil // Success
}

func callExternalService(ctx context.Context, email string) error {
	if email == "external-fail@example.com" {
		return fmt.Errorf("external service unavailable")
	}
	return nil
}

func isValidEmail(email string) bool {
	// Simplified email validation
	return len(email) > 5 && contains(email, "@") && contains(email, ".")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
