# go-cuserr

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/itsatony/go-cuserr)](https://goreportcard.com/report/github.com/itsatony/go-cuserr)
[![Coverage](https://img.shields.io/badge/coverage-92.3%25-brightgreen)](https://github.com/itsatony/go-cuserr)

**Protocol-agnostic Go error handling with 92.3% test coverage. One error type for HTTP, gRPC, CLI, batch jobs, and more.**

## Why go-cuserr?

✅ **Protocol-agnostic design** - Same error works in HTTP APIs, gRPC services, CLI tools, batch jobs, and syslog
✅ **Reduces boilerplate by 50%** - `NewValidationError("email", "invalid")` vs verbose error construction
✅ **Universal error mapping** - Automatic conversion to HTTP status, gRPC codes, CLI exit codes, syslog severity
✅ **Production-safe by default** - Sensitive details filtered in production mode
✅ **Thread-safe & performant** - 92.3% test coverage, comprehensive race detection
✅ **Zero dependencies** - Only uses Go standard library

**Perfect for**: REST APIs, gRPC services, CLI applications, batch jobs, microservices, cron tasks, background workers

## Installation

```bash
go get github.com/itsatony/go-cuserr
```

## Quick Start (30 seconds)

```go
package main

import (
    "fmt"
    "github.com/itsatony/go-cuserr"
)

func main() {
    // Create errors with automatic HTTP status mapping
    err := cuserr.NewNotFoundError("user", "usr_12345")
    
    fmt.Printf("Error: %s\n", err.Error())                    // user with id 'usr_12345' not found
    fmt.Printf("HTTP Status: %d\n", err.ToHTTPStatus())       // 404
    fmt.Printf("JSON: %+v\n", err.ToJSON())                   // {"error": {"code": "NOT_FOUND", ...}}
}
```

**That's it!** Three lines for production-ready error handling with HTTP integration.

## 🌐 Protocol-Agnostic Design

**One error type, every protocol.** Create an error once, use it everywhere:

```go
// Create error once
err := cuserr.NewUnauthorizedError("Invalid API token")
err.OperationID = "op-12345"
err.TraceID = "trace-abc"

// Use in HTTP API
w.WriteHeader(err.ToHTTPStatus())           // 401 Unauthorized
json.NewEncoder(w).Encode(err.ToJSON())

// Use in gRPC service
return nil, status.Error(
    codes.Code(err.ToGRPCCode()),           // codes.Unauthenticated (16)
    err.Message,
)

// Use in CLI application
os.Exit(err.ToExitCode())                   // Exit code 1

// Use in batch jobs / cron tasks
syslog.Write(err.ToSyslogSeverity(), err.Message)  // Severity 3 (Error)

// Use in logging systems
logger.Error(err.DetailedError())           // Full context with metadata
```

**Real-world examples:**
- **[gRPC Service](examples/grpc_service.go)** - Microservices with distributed tracing
- **[Batch Jobs](examples/batch_job.go)** - Cron tasks, background workers, data migrations
- **[HTTP Service](examples/http_service.go)** - REST APIs with full error handling

## 🆕 v0.3.0: Logger Injection & Options Pattern

**Modern, injectable logging - bring your own logger!**

```go
import (
    "log/slog"
    "github.com/itsatony/go-cuserr"
    "github.com/itsatony/go-cuserr/adapters"
)

// Set up your logger (slog, Zap, or Zerolog)
logger := adapters.NewSlogAdapter(slog.Default())

// Inject logger via options pattern
err := cuserr.NewInternalError("database", dbErr,
    cuserr.WithLogger(logger),           // Inject your logger
    cuserr.WithContext(ctx),             // Auto-extract request_id, user_id, trace_id
    cuserr.WithMetadata("query", sql),   // Add custom metadata
)

// Log when you want (uses injected logger)
err.Log(ctx)  // Outputs structured JSON with all error context
```

**Supported loggers:**
- ✅ **slog** (Go 1.21+ standard library) - `adapters.NewSlogAdapter()`
- ✅ **Zap** (Uber's blazing-fast logger) - `adapters.NewZapAdapter()`
- ✅ **Zerolog** (Zero-allocation logger) - `adapters.NewZerologAdapter()`

**Library-friendly:** Errors work perfectly without loggers! No forced dependencies.

[See full example →](examples/modern_api_usage.go)

## Common Patterns

Need more? Pick your use case:

| Use Case | Constructor | Example |
|----------|-------------|---------|
| **Web API validation** | `NewValidationError(field, msg)` | `NewValidationError("email", "invalid format")` |
| **Resource not found** | `NewNotFoundError(resource, id)` | `NewNotFoundError("user", "123")` |
| **Authentication** | `NewUnauthorizedError(reason)` | `NewUnauthorizedError("token expired")` |
| **Authorization** | `NewForbiddenError(action, resource)` | `NewForbiddenError("delete", "admin_user")` |
| **External service** | `NewExternalError(service, op, err)` | `NewExternalError("payment-api", "charge", err)` |
| **Internal errors** | `NewInternalError(component, err)` | `NewInternalError("database", err)` |

**Multiple validation errors?** → [Error Aggregation](#error-aggregation)  
**Need examples?** → [Examples](#examples)  
**Advanced features?** → [Detailed Documentation](#detailed-features)

## Error Aggregation

Handle multiple validation errors in a single response:

```go
// Collect multiple validation errors
collection := cuserr.NewValidationErrorCollection()
collection.AddValidation("email", "required field")
collection.AddValidation("password", "too short")
collection.AddValidation("age", "must be 18+")

// Convert to HTTP response
w.WriteHeader(collection.ToHTTPStatus()) // 400
json.NewEncoder(w).Encode(collection.ToPublicJSON())
```

## Protocol Mappings

**Same error category maps to appropriate codes across all protocols:**

| Category | HTTP | gRPC | CLI Exit | Syslog | When to Use |
|----------|------|------|----------|--------|-------------|
| `Validation` | 400 | 3 (INVALID_ARGUMENT) | 1 | 6 (Info) | Invalid user input |
| `Unauthorized` | 401 | 16 (UNAUTHENTICATED) | 1 | 3 (Error) | Authentication required |
| `Forbidden` | 403 | 7 (PERMISSION_DENIED) | 1 | 3 (Error) | Permission denied |
| `NotFound` | 404 | 5 (NOT_FOUND) | 1 | 6 (Info) | Resource doesn't exist |
| `Conflict` | 409 | 6 (ALREADY_EXISTS) | 1 | 3 (Error) | Resource already exists |
| `Timeout` | 408 | 4 (DEADLINE_EXCEEDED) | 1 | 4 (Warning) | Operation timeout |
| `RateLimit` | 429 | 8 (RESOURCE_EXHAUSTED) | 1 | 4 (Warning) | Too many requests |
| `External` | 502 | 14 (UNAVAILABLE) | 2 | 3 (Error) | External service failed |
| `Internal` | 500 | 13 (INTERNAL) | 2 | 2 (Critical) | Server error |

## Examples

**Protocol-Agnostic:**
- **[gRPC Service](examples/grpc_service.go)** - 🆕 Microservices with distributed tracing, automatic gRPC code mapping
- **[Batch Jobs](examples/batch_job.go)** - 🆕 Cron tasks, background workers, data migrations with syslog integration

**HTTP-Focused:**
- **[Modern API Usage](examples/modern_api_usage.go)** - **v0.3.0**: Logger injection & options pattern
- **[HTTP Service](examples/http_service.go)** - Complete REST API with error handling
- **[Middleware](examples/middleware.go)** - Request tracing, auth, rate limiting

**Fundamentals:**
- **[Basic Usage](examples/basic_usage.go)** - Error creation and handling fundamentals
- **[Enhanced Features](examples/enhanced_usage.go)** - v0.2.0+ advanced patterns

**Run examples:**
```bash
# gRPC service simulation
go run -tags grpc_service examples/grpc_service.go

# Batch job processing
go run -tags batch_job examples/batch_job.go

# HTTP service (runs on :8080)
go run examples/http_service.go
```

## Contributing

1. Fork → Create branch → Add tests → Submit PR
2. See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines
3. All contributions must pass GitHub Actions quality gates

---

## Detailed Documentation

<details>
<summary><strong>📖 Click to expand comprehensive API documentation</strong></summary>

### Core Error Creation

Pre-defined sentinel errors for common scenarios:

```go
var (
    ErrNotFound      = errors.New("resource not found")
    ErrAlreadyExists = errors.New("resource already exists") 
    ErrInvalidInput  = errors.New("invalid input")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrInternal      = errors.New("internal error")
    ErrTimeout       = errors.New("operation timeout")
    ErrRateLimit     = errors.New("rate limit exceeded")
    ErrExternal      = errors.New("external service error")
)
```

## Usage Examples

### Basic Error Creation

```go
// Using sentinel errors (automatic categorization)
err := cuserr.NewCustomError(cuserr.ErrNotFound, nil, "user not found")

// Using explicit categories
err := cuserr.NewCustomErrorWithCategory(
    cuserr.ErrorCategoryValidation,
    "INVALID_EMAIL", 
    "email format is invalid")
```

### Rich Context with Metadata

```go
err := cuserr.NewCustomError(cuserr.ErrInternal, nil, "database connection failed")
err.WithMetadata("database", "users")
err.WithMetadata("operation", "INSERT")
err.WithMetadata("table", "profiles")
err.OperationID = "req_xyz789"

// Thread-safe metadata access
if dbName, exists := err.GetMetadata("database"); exists {
    log.Printf("Database: %s", dbName)
}

// Get all metadata
allMeta := err.GetAllMetadata()
```

### Error Wrapping

```go
originalErr := errors.New("connection refused")

// Wrap with custom error
wrappedErr := cuserr.NewCustomError(
    cuserr.ErrExternal, 
    originalErr, 
    "payment service unavailable")

// Check wrapped error
if errors.Is(wrappedErr, originalErr) {
    // Handle connection issue
}

// Convenience wrapper
wrapped := cuserr.WrapWithCustomError(
    originalErr,
    cuserr.ErrorCategoryExternal,
    "PAYMENT_UNAVAILABLE",
    "payment processing is temporarily unavailable")
```

### HTTP Service Integration

```go
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    userID := extractUserID(r)
    
    user, err := h.service.GetUser(r.Context(), userID)
    if err != nil {
        var customErr *cuserr.CustomError
        if errors.As(err, &customErr) {
            // Log detailed error
            log.Printf("Error: %s", customErr.DetailedError())
            
            // Send appropriate HTTP response
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(customErr.ToHTTPStatus())
            json.NewEncoder(w).Encode(customErr.ToJSON())
            return
        }
        
        // Handle unexpected errors
        http.Error(w, "Internal Server Error", 500)
        return
    }
    
    // Success response
    json.NewEncoder(w).Encode(user)
}
```

### Error Checking Patterns

```go
err := service.DoOperation()

// Check by sentinel error
if errors.Is(err, cuserr.ErrUnauthorized) {
    // Handle unauthorized
}

// Check by category
if cuserr.IsErrorCategory(err, cuserr.ErrorCategoryValidation) {
    // Handle validation errors
}

// Check by error code
if cuserr.IsErrorCode(err, "RATE_LIMIT_EXCEEDED") {
    // Handle rate limiting
}

// Extract error information
category := cuserr.GetErrorCategory(err)
code := cuserr.GetErrorCode(err)
metadata := cuserr.GetErrorMetadata(err, "user_id")
```

## Configuration

Configure global behavior:

```go
// Development configuration
cuserr.SetConfig(&cuserr.Config{
    EnableStackTrace: true,
    MaxStackDepth:    10,
    ProductionMode:   false,
})

// Production configuration  
cuserr.SetConfig(&cuserr.Config{
    EnableStackTrace: false,
    MaxStackDepth:    0,
    ProductionMode:   true,
})
```

### Environment Variables

- `CUSERR_ENABLE_STACK_TRACE`: Enable/disable stack trace capture
- `CUSERR_MAX_STACK_DEPTH`: Maximum stack trace depth
- `CUSERR_PRODUCTION_MODE`: Enable production mode (hides sensitive details)

## Stack Traces

When enabled, stack traces are automatically captured:

```go
err := cuserr.NewCustomError(cuserr.ErrInternal, nil, "something went wrong")

// Access stack trace
frames := err.GetStackTrace()
stackString := err.GetStackTraceString()

// Filter stack trace
err.FilterStackTrace("testing.", "runtime.")

// Manual stack trace management
err.WithStackTrace(customFrames)
err.ClearStackTrace() // Save memory
```

## JSON Serialization

### Standard JSON Output

```go
err := cuserr.NewCustomError(cuserr.ErrValidation, nil, "validation failed")
err.WithMetadata("field", "email")
err.OperationID = "req_123"

jsonData := err.ToJSON()
// Output:
// {
//   "error": {
//     "code": "INVALID_INPUT",
//     "message": "validation failed",
//     "category": "validation",
//     "metadata": {"field": "email"},
//     "operation_id": "req_123",
//     "timestamp": "2023-01-01T12:00:00Z"
//   }
// }
```

### Client-Safe JSON (Production Mode)

```go
// In production mode, sensitive details are filtered
publicJSON := err.ToPublicJSON()
// Filters out internal metadata and provides generic messages for internal errors
```

## Thread Safety

All operations are thread-safe:

```go
// Concurrent metadata access
go func() {
    err.WithMetadata("key1", "value1")
}()

go func() {
    err.WithMetadata("key2", "value2")  
}()

go func() {
    metadata := err.GetAllMetadata() // Returns a copy
}()
```

## Performance

Benchmarked operations (on AMD Ryzen 7 7735HS):

- **Error Creation**: ~1,200 ns/op (896 B/op, 10 allocs/op)
- **Metadata Access**: ~68 ns/op (7 B/op, 1 allocs/op) 
- **HTTP Status**: ~1.1 ns/op (0 allocs/op)
- **JSON Serialization**: ~641 ns/op (1,112 B/op, 12 allocs/op)

Stack trace capture adds ~1,000 ns overhead but can be disabled in production.

## Examples

See the [`examples/`](./examples) directory for comprehensive examples:

- **[basic_usage.go](./examples/basic_usage.go)** - Core functionality demonstration
- **[http_service.go](./examples/http_service.go)** - HTTP service integration
- **[middleware.go](./examples/middleware.go)** - Middleware patterns

Run examples:

```bash
# Basic usage
go run examples/basic_usage.go

# HTTP service (runs on :8080)
go run examples/http_service.go

# Middleware demo (runs on :8081)  
go run examples/middleware.go
```

## Testing

```bash
# Run all tests
go test -v

# Run with race detection
go test -race -v

# Run benchmarks
go test -bench=. -benchmem

# Test coverage
go test -cover
```

## Best Practices

### 1. Use Sentinel Errors for Type Safety

```go
// Good - type safe
if errors.Is(err, cuserr.ErrNotFound) {
    // handle not found
}

// Avoid - string matching  
if err.Error() == "not found" {
    // fragile
}
```

### 2. Add Meaningful Metadata

```go
err := cuserr.NewCustomError(cuserr.ErrValidation, nil, "validation failed")
err.WithMetadata("field", "email")
err.WithMetadata("input", userInput)
err.WithMetadata("rule", "email_format")
```

### 3. Use Operation IDs for Tracing

```go
err := cuserr.NewCustomError(cuserr.ErrInternal, nil, "operation failed")
err.OperationID = getOperationID(ctx)
// Works for HTTP requests, gRPC calls, CLI commands, batch jobs, etc.
```

### 4. Configure for Environment

```go
// Development
cuserr.SetConfig(&cuserr.Config{
    EnableStackTrace: true,
    ProductionMode:   false,
})

// Production
cuserr.SetConfig(&cuserr.Config{
    EnableStackTrace: false,
    ProductionMode:   true,
})
```

### 5. Log Detailed Errors, Return Client-Safe Responses

```go
if err != nil {
    // Log detailed information
    log.Printf("Error: %s", customErr.DetailedError())

    // Return client-safe response (HTTP example)
    w.WriteHeader(customErr.ToHTTPStatus())
    json.NewEncoder(w).Encode(customErr.ToPublicJSON())
}
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make changes and add tests
4. Run tests (`go test -v -race`)
5. Commit changes (`git commit -am 'Add amazing feature'`)
6. Push to branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Setup

```bash
# Clone repository
git clone https://github.com/itsatony/go-cuserr.git
cd go-cuserr

# Run tests
go test -v -race

# Run benchmarks
go test -bench=. -benchmem

# Check formatting
go fmt ./...

# Run linter (if available)
golangci-lint run

# The full local CI gate — hosted GitHub Actions is retired, this is the gate
make ci-local
```

## 🆕 v0.2.0 Detailed Features

### Convenience Constructors
Drastically reduce boilerplate with specialized constructors:

```go
// Before (v0.1.0) - verbose
err := cuserr.NewCustomError(cuserr.ErrInvalidInput, nil, "email format is invalid").
    WithMetadata("field", "email").
    WithMetadata("error_type", "validation")

// After (v0.2.0) - concise  
err := cuserr.NewValidationError("email", "invalid format")
```

Available constructors:
- `NewValidationError(field, message)` 
- `NewNotFoundError(resource, id)`
- `NewUnauthorizedError(reason)`
- `NewForbiddenError(action, resource)`
- `NewInternalError(component, wrapped)`
- `NewExternalError(service, operation, wrapped)`
- `NewTimeoutError(operation, wrapped)`
- `NewRateLimitError(limit, window)`
- `NewConflictError(resource, field, value)`

### Context-Based Configuration
Configure errors per request with automatic context extraction:

```go
// Set context-based configuration
ctx := cuserr.WithProductionMode(ctx)
ctx = context.WithValue(ctx, "operation_id", "req_123")
ctx = context.WithValue(ctx, "user_id", "usr_456")

// Create context-aware errors (auto-extracts metadata)
err := cuserr.NewValidationErrorFromContext(ctx, "email", "invalid")
fmt.Println(err.OperationID) // "req_123"
userID, _ := err.GetMetadata("user_id") // "usr_456"
```

### Error Aggregation
Collect multiple validation errors into a single response:

```go
// Build validation error collection
collection := cuserr.NewValidationCollectionBuilder().
    WithContext(ctx).
    AddValidation("email", "required field").
    AddValidation("password", "too short").
    AddValidationWithCode("age", "must be 18+", "AGE_RESTRICTION").
    Build()

// Convert to HTTP response
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(collection.ToHTTPStatus()) // 400
json.NewEncoder(w).Encode(collection.ToPublicJSON())
```

### 🆕 v0.3.0: Options Pattern & Logger Injection

**Modern dependency injection following Go best practices:**

```go
import (
    "github.com/itsatony/go-cuserr"
    "github.com/itsatony/go-cuserr/adapters"
)

// Available options:
// - WithLogger(logger)      // Inject your logger
// - WithContext(ctx)         // Auto-extract operation_id, user_id, trace_id
// - WithMetadata(k, v)       // Add custom metadata
// - WithOperationID(id)      // Set operation ID
// - WithTraceID(id)          // Set trace ID (for distributed tracing)
// - WithSpanID(id)           // Set span ID (for distributed tracing)
// - WithStack(skip)          // Explicitly capture stack trace
// - WithoutStack()           // Disable stack trace

// Example: Service layer with logger injection
type UserService struct {
    logger cuserr.Logger
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.db.FindUser(id)
    if err != nil {
        // Create error with all options
        return nil, cuserr.NewInternalError("database", err,
            cuserr.WithContext(ctx),           // Auto-extracts context values
            cuserr.WithLogger(s.logger),       // Inject service logger
            cuserr.WithMetadata("user_id", id), // Add custom metadata
        )
    }
    return user, nil
}

// Supported logger adapters:
logger := adapters.NewSlogAdapter(slog.Default())     // Go 1.21+ stdlib
logger := adapters.NewZapAdapter(zapLogger)           // Uber Zap
logger := adapters.NewZerologAdapter(zerologLogger)   // Zerolog

// Log when you want
err.Log(ctx)  // Uses injected logger, outputs structured JSON
```

**Benefits:**
- ✅ **No global state** - Proper dependency injection
- ✅ **Library-friendly** - Works without logger (NoOpLogger default)
- ✅ **Composable** - Mix and match options as needed
- ✅ **Zero-cost** - No logger = no logging overhead

### Typed Metadata
Type-safe metadata operations with IDE support:

```go
err := cuserr.NewExternalError("payment-api", "charge", serviceErr)
tm := err.GetTypedMetadata()

// Fluent, type-safe metadata
tm.WithUserID("usr_123").
    WithHTTPMethod("POST").
    WithURL("https://api.payment.com/charge").
    WithStatusCode(503).
    WithResponseTime(5*time.Second).
    WithRetryCount(3)

// Type-safe retrieval  
if statusCode, exists := tm.GetStatusCode(); exists {
    fmt.Printf("External API returned: %d\n", statusCode)
}
```

### Migration Helpers
Easy migration from existing error handling:

```go
// Migrate from stdlib errors
stdErr := fmt.Errorf("user not found")
customErr := cuserr.FromStdError(stdErr, "failed to get user")
// Automatically categorized as ErrorCategoryNotFound

// Migrate from HTTP status codes
httpErr := cuserr.FromHTTPStatus(404, "resource not found")

// Migrate from SQL errors  
sqlErr := fmt.Errorf("duplicate key constraint violation")
dbErr := cuserr.FromSQLError(sqlErr, "INSERT INTO users...")
// Automatically categorized as ErrorCategoryConflict

// Batch migration with reporting
errors := []error{err1, err2, err3}
collection, report := cuserr.BatchMigrate(errors)
fmt.Println(report.Summary()) // "Migration completed: 3 total, 3 migrated, 0 failed"
```

### Performance Improvements

**Memory Efficiency (30% reduction):**
- Lazy loading for metadata maps
- Smart memory management  
- Preserved stack trace accuracy

**Benchmarks:**
```
BenchmarkErrorCreation-8                 1000000  1125 ns/op   328 B/op   6 allocs/op
BenchmarkMetadataOperations-8           15000000    65.5 ns/op  16 B/op   1 allocs/op
BenchmarkJSONSerialization-8             1500000   631 ns/op   240 B/op   5 allocs/op
BenchmarkHTTPStatus-8                 1000000000   1.08 ns/op   0 B/op    0 allocs/op
```

</details>

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and breaking changes.

---

**Excellence. Always.** - Part of the vAudience.AI Go ecosystem.