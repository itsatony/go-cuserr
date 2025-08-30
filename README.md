# go-cuserr

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/itsatony/go-cuserr)](https://goreportcard.com/report/github.com/itsatony/go-cuserr)

A comprehensive, thread-safe custom error handling package for Go applications with rich context, HTTP status mapping, and production-ready features.

## Features

- 🎯 **Sentinel Error Types** - Predefined errors for common scenarios
- 🏷️ **Error Categorization** - Automatic HTTP status code mapping  
- 🔍 **Rich Context** - Metadata, request IDs, and stack traces
- 🔒 **Thread-Safe** - Concurrent access protected with mutexes
- 🌐 **HTTP Integration** - JSON serialization and status code mapping
- 📊 **Stack Traces** - Optional runtime stack capture for debugging
- 🛡️ **Production Ready** - Client-safe error messages and configurable details
- ⚡ **High Performance** - Benchmarked and optimized
- 🔗 **Error Wrapping** - Full compatibility with Go's `errors.Is()` and `errors.As()`

## Installation

```bash
go get github.com/itsatony/go-cuserr
```

## Quick Start

```go
package main

import (
    "errors"
    "fmt"
    "log"
    
    "github.com/itsatony/go-cuserr"
)

func main() {
    // Create a custom error with rich context
    err := cuserr.NewCustomError(
        cuserr.ErrNotFound, 
        nil, 
        "user not found")
    err.WithMetadata("user_id", "usr_12345")
    err.WithRequestID("req_abc123")
    
    // Check error type
    if errors.Is(err, cuserr.ErrNotFound) {
        fmt.Printf("HTTP Status: %d\n", err.ToHTTPStatus()) // 404
    }
    
    // Convert to JSON for API response
    jsonResponse := err.ToJSON()
    fmt.Printf("JSON: %+v\n", jsonResponse)
    
    // Log detailed error information
    log.Printf("Error: %s", err.DetailedError())
}
```

## Error Categories and HTTP Status Mapping

| Category | HTTP Status | Description |
|----------|-------------|-------------|
| `ErrorCategoryValidation` | 400 | Input validation failures |
| `ErrorCategoryUnauthorized` | 401 | Authentication required |
| `ErrorCategoryForbidden` | 403 | Insufficient permissions |
| `ErrorCategoryNotFound` | 404 | Resource not found |
| `ErrorCategoryTimeout` | 408 | Operation timeout |
| `ErrorCategoryConflict` | 409 | Resource conflicts |
| `ErrorCategoryRateLimit` | 429 | Rate limit exceeded |
| `ErrorCategoryExternal` | 502 | External service failures |
| `ErrorCategoryInternal` | 500 | Internal server errors |

## Sentinel Errors

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
err.WithRequestID("req_xyz789")

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
err.WithRequestID("req_123")

jsonData := err.ToJSON()
// Output:
// {
//   "error": {
//     "code": "INVALID_INPUT",
//     "message": "validation failed", 
//     "category": "validation",
//     "metadata": {"field": "email"},
//     "request_id": "req_123",
//     "timestamp": "2023-01-01T12:00:00Z"
//   }
// }
```

### Client-Safe JSON (Production Mode)

```go
// In production mode, sensitive details are filtered
clientJSON := err.ToClientJSON()
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

### 3. Use Request IDs for Tracing

```go
err := cuserr.NewCustomError(cuserr.ErrInternal, nil, "operation failed")
err.WithRequestID(getRequestID(ctx))
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
    
    // Return client-safe response
    w.WriteHeader(customErr.ToHTTPStatus())
    json.NewEncoder(w).Encode(customErr.ToClientJSON())
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
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and breaking changes.

---

**Excellence. Always.** - Part of the vAudience.AI Go ecosystem.