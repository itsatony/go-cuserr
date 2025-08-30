# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is **go-cuserr**, a custom error handling package for Go applications. The project implements a comprehensive error system with rich context, categorization, and HTTP status code mapping, following vAudience.AI's "Excellence. Always." philosophy.

## Architecture

### Core Components

The package provides:
- **Sentinel Errors**: Predefined error types for common scenarios (not found, validation, unauthorized, etc.)
- **Error Categories**: Map to HTTP status codes (validation→400, not_found→404, internal→500, etc.)
- **CustomError Type**: Rich error context with metadata, request IDs, stack traces, and wrapping
- **Stack Trace Capture**: Optional runtime stack trace capture for debugging

### Key Files

- `ecuserr.rough_base.go`: Complete implementation with ~505 lines including examples
- `Agent.md`: Comprehensive development guidelines and coding standards from vAudience.AI
- No `go.mod` yet - needs initialization

## Development Commands

### Initial Setup (Required)
```bash
# Initialize Go module (run first)
go mod init github.com/itsatony/go-cuserr

# Add dependencies (if any external deps are needed)
go mod tidy
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestFunctionName
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run golangci-lint (if configured)
golangci-lint run
```

### Build and Install
```bash
# Build (no main package, library only)
go build

# Install as module dependency
go install
```

## Code Standards (from Agent.md)

### Critical Requirements
1. **Thread Safety by Default**: All code must handle concurrent access safely
2. **No Magic Strings**: Every string literal must be a constant
3. **Comprehensive Error Handling**: Use sentinel errors with rich context
4. **Complete Type Safety**: Full type hints throughout
5. **Production-Ready**: No stubs, mocks, or incomplete implementations

### Error Handling Pattern
```go
// Service-specific constants
const (
    ERR_MSG_USER_NOT_FOUND = "user with ID (%s) not found"
    CLASS_NAME = "UserService"
)

// Usage example
func (s *Service) Method() error {
    methodPrefix := fmt.Sprintf("[%s.%s]", CLASS_NAME, "MethodName")
    
    if err := validate(); err != nil {
        errMsg := fmt.Sprintf(ERR_MSG_USER_NOT_FOUND, userID)
        return NewCustomError(ErrNotFound, err, errMsg).
            WithMetadata("user_id", userID).
            WithRequestID(requestID)
    }
    return nil
}
```

### File Naming Convention
Follow pattern: `{project}.{type}.{module}.{framework}.go`
- Example: `cuserr.errors.user.go`, `cuserr.constants.validation.go`

## Testing Strategy

### Required Test Types
1. **Unit Tests**: Test individual functions and error creation
2. **Edge Cases**: Nil handling, empty strings, invalid inputs  
3. **Concurrency Tests**: Thread safety with `-race` flag
4. **Integration Tests**: HTTP handler integration, JSON serialization
5. **Benchmark Tests**: Performance regression testing

### Example Test Structure
```go
func TestCustomError_Creation(t *testing.T) {
    // Test error creation with all fields
}

func TestCustomError_HTTPMapping(t *testing.T) {
    // Test category to HTTP status mapping
}

func TestCustomError_Concurrency(t *testing.T) {
    // Test concurrent access to metadata
}
```

## Implementation Status

**Current**: Single file `ecuserr.rough_base.go` with complete implementation including:
- All core types and functions
- HTTP status mapping
- JSON serialization
- Stack trace capture
- Extensive usage examples

**Next Steps for Full Implementation**:
1. Initialize `go.mod`
2. Split into proper package structure
3. Add comprehensive test suite
4. Add benchmarks
5. Create examples directory
6. Add proper documentation

## Key Features

- **Rich Error Context**: Metadata, request IDs, timestamps
- **HTTP Integration**: Automatic status code mapping
- **Stack Traces**: Optional runtime stack capture  
- **Error Wrapping**: Supports `errors.Is()` and `errors.As()`
- **JSON Serialization**: Ready for API responses
- **Thread Safe**: Concurrent access safe by design