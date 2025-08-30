# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-01-30

### Added

#### Core Features
- **Thread-safe custom error handling** with mutex-protected metadata and stack trace access
- **Sentinel error types** for common scenarios (NotFound, Unauthorized, Validation, etc.)
- **Automatic error categorization** with HTTP status code mapping
- **Rich error context** with metadata, request IDs, and timestamps
- **Optional stack trace capture** with configurable depth and filtering
- **Error wrapping support** compatible with Go's `errors.Is()` and `errors.As()`

#### HTTP Integration
- **JSON serialization** with standard and client-safe modes
- **HTTP status code mapping** from error categories
- **Production-mode filtering** for sensitive error details
- **Request ID propagation** for distributed tracing

#### Configuration & Environment
- **Global configuration management** with thread-safe access
- **Environment variable support** for configuration
- **Development vs Production modes** with different error exposure levels
- **Configurable stack trace capture** (enable/disable, max depth)

#### Error Categories & HTTP Status Mapping
- `ErrorCategoryValidation` → HTTP 400 (Bad Request)
- `ErrorCategoryUnauthorized` → HTTP 401 (Unauthorized)
- `ErrorCategoryForbidden` → HTTP 403 (Forbidden)
- `ErrorCategoryNotFound` → HTTP 404 (Not Found)
- `ErrorCategoryTimeout` → HTTP 408 (Request Timeout)
- `ErrorCategoryConflict` → HTTP 409 (Conflict)
- `ErrorCategoryRateLimit` → HTTP 429 (Too Many Requests)
- `ErrorCategoryExternal` → HTTP 502 (Bad Gateway)
- `ErrorCategoryInternal` → HTTP 500 (Internal Server Error)

#### Sentinel Errors
- `ErrNotFound` - Resource not found
- `ErrAlreadyExists` - Resource already exists
- `ErrInvalidInput` - Invalid input data
- `ErrUnauthorized` - Authentication failure
- `ErrForbidden` - Authorization failure
- `ErrInternal` - Internal error
- `ErrTimeout` - Operation timeout
- `ErrRateLimit` - Rate limit exceeded
- `ErrExternal` - External service error

#### API Functions

**Error Creation:**
- `NewCustomError(sentinel, wrapped, message)` - Create error with automatic categorization
- `NewCustomErrorWithCategory(category, code, message)` - Create error with explicit category
- `WrapWithCustomError(err, category, code, message)` - Wrap existing error

**Error Enhancement:**
- `WithMetadata(key, value)` - Add metadata (chainable, thread-safe)
- `WithRequestID(requestID)` - Add request ID for tracing
- `WithStackTrace(frames)` - Manually set stack trace

**Error Inspection:**
- `IsErrorCategory(err, category)` - Check error category
- `IsErrorCode(err, code)` - Check error code
- `GetErrorCategory(err)` - Extract error category
- `GetErrorCode(err)` - Extract error code
- `GetErrorMetadata(err, key)` - Extract metadata value

**Error Output:**
- `Error()` - Standard error string
- `DetailedError()` - Full error details with stack trace
- `ShortError()` - Concise error for logging
- `ToJSON()` - Standard JSON representation
- `ToClientJSON()` - Client-safe JSON representation
- `ToJSONString()` - Basic JSON string representation
- `ToHTTPStatus()` - HTTP status code

**Stack Trace Management:**
- `GetStackTrace()` - Get stack trace frames
- `GetStackTraceString()` - Get formatted stack trace string
- `FilterStackTrace(patterns...)` - Filter stack trace by patterns
- `ClearStackTrace()` - Remove stack trace to save memory

**Configuration:**
- `SetConfig(config)` - Set global configuration
- `GetConfig()` - Get current configuration (returns copy)
- `DefaultConfig()` - Get default configuration

#### Testing & Quality Assurance
- **Comprehensive unit tests** with >95% coverage
- **Concurrency tests** with race condition detection
- **Benchmark tests** for performance validation
- **Integration examples** for HTTP services and middleware
- **Thread-safety validation** with `-race` flag support

#### Performance Characteristics
- **Error Creation**: ~1,200 ns/op (896 B/op, 10 allocs/op)
- **Metadata Access**: ~68 ns/op (7 B/op, 1 allocs/op)
- **HTTP Status Mapping**: ~1.1 ns/op (0 allocs/op)
- **JSON Serialization**: ~641 ns/op (1,112 B/op, 12 allocs/op)
- **Stack Trace Disabled**: ~277 ns/op (248 B/op, 3 allocs/op)

#### Documentation & Examples
- **Comprehensive README** with usage examples and best practices
- **API documentation** with GoDoc comments
- **Example implementations**:
  - Basic usage patterns (`examples/basic_usage.go`)
  - HTTP service integration (`examples/http_service.go`) 
  - Middleware patterns (`examples/middleware.go`)

#### Code Organization
Following vAudience.AI conventions:
- `cuserr.constants.errors.go` - All string constants and configuration values
- `cuserr.types.core.go` - Core type definitions and configuration
- `cuserr.errors.sentinel.go` - Sentinel error definitions and mapping
- `cuserr.service.core.go` - Primary error creation and manipulation functions
- `cuserr.utils.http.go` - HTTP integration and JSON serialization
- `cuserr.utils.stack.go` - Stack trace capture and formatting

#### Environment Variables
- `CUSERR_ENABLE_STACK_TRACE` - Enable/disable stack trace capture
- `CUSERR_MAX_STACK_DEPTH` - Maximum stack trace depth
- `CUSERR_PRODUCTION_MODE` - Enable production mode

### Technical Implementation

#### Thread Safety
- All metadata operations protected by `sync.RWMutex`
- Stack trace access protected by mutex
- Global configuration access protected by mutex
- Thread-safe metadata copying to prevent external modification
- Comprehensive race condition testing

#### Memory Management
- Stack trace capture can be disabled for performance
- Metadata maps initialized lazily
- Stack trace filtering to remove unnecessary frames
- Optional stack trace clearing for memory optimization

#### Error Chain Compatibility
- Full support for `errors.Is()` checking against sentinel errors
- Full support for `errors.As()` type assertions
- Proper error unwrapping with `errors.Unwrap()`
- Maintains original error context when wrapping

### Breaking Changes
- N/A (Initial release)

### Dependencies
- **Standard Library Only** - No external dependencies
- Minimum Go version: 1.21+

### Migration Guide
- N/A (Initial release)

---

## Release Notes

### v0.1.0 - Foundation Release

This initial release establishes go-cuserr as a production-ready, thread-safe error handling package for Go applications. Built with vAudience.AI's "Excellence. Always." philosophy, it provides comprehensive error context, automatic HTTP integration, and high-performance operation suitable for both development and production environments.

**Key Highlights:**
- **Zero external dependencies** - Uses only Go standard library
- **Thread-safe by design** - All operations protected against race conditions  
- **Production ready** - Configurable error detail exposure and client-safe responses
- **High performance** - Benchmarked and optimized for minimal overhead
- **Comprehensive testing** - Full test coverage including concurrency validation

This release provides everything needed to implement robust, traceable, and user-friendly error handling in Go applications, from simple CLI tools to complex distributed systems.