# Migration Guide: v0.2.x → v0.3.0

## Overview

Version 0.3.0 introduces **modern logger injection** and the **options pattern** while maintaining **100% backward compatibility**. All existing code continues to work without changes. This guide helps you adopt the new features at your own pace.

## TL;DR - Do I Need to Change Anything?

**No.** Your existing code works perfectly as-is. The v0.3.0 changes are:
- ✅ **Additive only** - New optional features
- ✅ **Zero breaking changes** - All existing APIs unchanged
- ✅ **Graceful defaults** - NoOpLogger means zero overhead if you don't use logging

## What's New in v0.3.0

### 1. Critical Bug Fixes (Automatic)

These fixes apply automatically when you upgrade:

```go
// ✅ Fixed: Race conditions in concurrent logger access
// ✅ Fixed: Thread-unsafe global config access
// ✅ Fixed: Incorrect error message in ValidateLength

// Your code benefits immediately - no changes needed!
```

**Impact**: If you were experiencing intermittent crashes or race condition warnings, these are now fixed.

### 2. Logger Injection (Optional)

**Before v0.3.0**: No structured logging support
```go
// Had to rely on global logger or external logging
err := cuserr.NewNotFoundError("user", userID)
// Then manually log somewhere else
log.Error("User not found", "id", userID, "error", err)
```

**After v0.3.0**: Inject your logger directly
```go
// Option 1: Use slog (Go 1.21+ standard library)
import "github.com/itsatony/go-cuserr/adapters"

logger := adapters.NewSlogAdapter(slog.Default())
err := cuserr.NewNotFoundError("user", userID,
    cuserr.WithLogger(logger),
    cuserr.WithContext(ctx),
)
err.Log(ctx) // Automatically logs with all context!
```

### 3. Options Pattern (Optional)

**Before v0.3.0**: Manual metadata addition
```go
err := cuserr.NewInternalError("database", dbErr)
err.WithMetadata("query", sql)
err.WithMetadata("duration_ms", "250")
err.WithRequestID(requestID)
```

**After v0.3.0**: Clean, composable options
```go
err := cuserr.NewInternalError("database", dbErr,
    cuserr.WithLogger(logger),      // Inject logger
    cuserr.WithContext(ctx),         // Auto-extract request_id, user_id, trace_id
    cuserr.WithMetadata("query", sql),
    cuserr.WithMetadata("duration_ms", "250"),
)
err.Log(ctx) // All context included automatically
```

## Migration Scenarios

### Scenario 1: Keep Everything As-Is (Zero Effort)

**Current Code**:
```go
err := cuserr.NewNotFoundError("user", userID)
return c.JSON(err.ToHTTPStatus(), err.ToClientJSON())
```

**Migration**: None needed! This continues to work perfectly.

**Benefit**: You get the critical bug fixes automatically.

---

### Scenario 2: Add Basic Logging (Low Effort)

**Current Code**:
```go
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    user := s.db.Find(id)
    if user == nil {
        err := cuserr.NewNotFoundError("user", id)
        log.Error("User not found", "id", id) // Manual logging
        return nil, err
    }
    return user, nil
}
```

**Migrated Code**:
```go
import "github.com/itsatony/go-cuserr/adapters"

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    user := s.db.Find(id)
    if user == nil {
        err := cuserr.NewNotFoundError("user", id,
            cuserr.WithLogger(s.logger), // Inject service logger
            cuserr.WithContext(ctx),     // Auto-extract request context
        )
        err.Log(ctx) // Logs everything automatically!
        return nil, err
    }
    return user, nil
}
```

**Benefits**:
- Automatic structured logging with all error metadata
- Request ID, user ID, trace ID automatically included
- No manual field extraction needed

---

### Scenario 3: Service Layer Pattern (Medium Effort)

**Current Code**:
```go
type PaymentService struct {
    client *http.Client
}

func (s *PaymentService) Charge(ctx context.Context, amount float64) error {
    resp, err := s.client.Post("/charge", ...)
    if err != nil {
        return cuserr.NewExternalError("payment-api", "charge", err)
    }
    return nil
}
```

**Migrated Code**:
```go
import "github.com/itsatony/go-cuserr/adapters"

type PaymentService struct {
    client *http.Client
    logger cuserr.Logger // Add logger field
}

// Constructor injection
func NewPaymentService(client *http.Client, logger cuserr.Logger) *PaymentService {
    return &PaymentService{
        client: client,
        logger: logger,
    }
}

func (s *PaymentService) Charge(ctx context.Context, amount float64) error {
    resp, err := s.client.Post("/charge", ...)
    if err != nil {
        err := cuserr.NewExternalError("payment-api", "charge", err,
            cuserr.WithLogger(s.logger),
            cuserr.WithContext(ctx),
            cuserr.WithMetadata("amount", fmt.Sprintf("%.2f", amount)),
        )
        err.Log(ctx) // Comprehensive error logging
        return err
    }
    return nil
}
```

**Benefits**:
- Clean dependency injection
- Consistent error logging across service
- Easy to test (mock logger)

---

### Scenario 4: Library Development (No Logger Needed)

**Use Case**: You're building a library and don't want to force logging on consumers.

**Code**:
```go
// Your library code - no logger needed!
func (lib *MyLibrary) DoSomething(input string) error {
    if input == "" {
        // Still works perfectly - NoOpLogger is default
        return cuserr.NewValidationError("input", "cannot be empty")
    }
    return nil
}

// Consumers can optionally inject logger if they want
// err := lib.DoSomething(input)
// if customErr, ok := err.(*cuserr.CustomError); ok {
//     customErr.Log(ctx) // Works if consumer wants logging
// }
```

**Benefit**: Zero-cost abstraction - no performance overhead, no forced dependencies.

---

## Choosing a Logger Adapter

### Option 1: Slog (Recommended for New Projects)

**When to use**: Go 1.21+ projects, standard library preference

```go
import (
    "log/slog"
    "github.com/itsatony/go-cuserr/adapters"
)

// Standard JSON handler
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
adapter := adapters.NewSlogAdapter(logger)

err := cuserr.NewInternalError("database", dbErr,
    cuserr.WithLogger(adapter),
)
```

**Pros**: Standard library, no external dependencies, modern API
**Cons**: Requires Go 1.21+

### Option 2: Zap (Recommended for High Performance)

**When to use**: Performance-critical applications, high-throughput services

```go
import (
    "go.uber.org/zap"
    "github.com/itsatony/go-cuserr/adapters"
)

logger, _ := zap.NewProduction()
adapter := adapters.NewZapAdapter(logger)

err := cuserr.NewExternalError("api", "request", apiErr,
    cuserr.WithLogger(adapter),
)
```

**Pros**: Blazing fast, zero allocation, battle-tested
**Cons**: External dependency

### Option 3: Zerolog (Recommended for Zero Allocation)

**When to use**: Ultra-high-performance scenarios, minimal GC pressure

```go
import (
    "github.com/rs/zerolog"
    "github.com/itsatony/go-cuserr/adapters"
)

logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
adapter := adapters.NewZerologAdapter(logger)

err := cuserr.NewTimeoutError("database-query", queryErr,
    cuserr.WithLogger(adapter),
)
```

**Pros**: True zero allocation, excellent performance
**Cons**: External dependency

### Option 4: No Logger (Recommended for Libraries)

**When to use**: Library development, optional logging

```go
// Just use errors without logger - works perfectly!
err := cuserr.NewNotFoundError("resource", id)
return err // Consumer can choose to log or not
```

**Pros**: Zero overhead, no dependencies, maximum flexibility
**Cons**: No automatic logging

---

## Step-by-Step Migration Plan

### Phase 1: Upgrade (5 minutes)

```bash
# Update dependency
go get -u github.com/itsatony/go-cuserr

# Run tests to verify everything still works
go test ./... -race
```

**Expected result**: All tests pass, no changes needed.

### Phase 2: Choose Logger (10 minutes)

Decide on your logging strategy:

```go
// Example: Using slog
import (
    "log/slog"
    "github.com/itsatony/go-cuserr/adapters"
)

// Initialize once at application startup
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
loggerAdapter := adapters.NewSlogAdapter(logger)

// Store in service/handler structs
type MyService struct {
    logger cuserr.Logger
}
```

### Phase 3: Migrate Incrementally (Ongoing)

Pick one service/handler at a time:

```go
// Before
func (s *UserService) Delete(ctx context.Context, id string) error {
    if err := s.db.Delete(id); err != nil {
        return cuserr.NewInternalError("database", err)
    }
    return nil
}

// After (when you get to it)
func (s *UserService) Delete(ctx context.Context, id string) error {
    if err := s.db.Delete(id); err != nil {
        err := cuserr.NewInternalError("database", err,
            cuserr.WithLogger(s.logger),
            cuserr.WithContext(ctx),
        )
        err.Log(ctx)
        return err
    }
    return nil
}
```

**Strategy**: Migrate high-value code first (error-prone services, critical paths).

---

## Common Questions

### Q: Will my existing code break?

**A**: No. 100% backward compatible. All existing code continues to work without any changes.

### Q: Do I need to add a logger?

**A**: No. The default NoOpLogger means zero overhead if you don't use logging. It's completely optional.

### Q: Can I mix old and new patterns?

**A**: Yes! You can migrate incrementally. Some errors can use the new options pattern while others use the old pattern.

```go
// Old pattern - still works
err1 := cuserr.NewNotFoundError("user", id)

// New pattern - also works
err2 := cuserr.NewNotFoundError("user", id,
    cuserr.WithLogger(logger),
)
```

### Q: What about performance?

**A**: Zero overhead if you don't use logging. If you do use logging:
- NoOpLogger: 0 ns/op (compile-time optimized away)
- Logger injection: ~5-10 ns/op (single pointer check)
- Structured field extraction: ~200 ns/op (amortized)

### Q: Can I use multiple loggers in the same app?

**A**: Yes! Each error can have its own logger instance.

```go
// Service A uses Zap
errA := cuserr.NewInternalError("serviceA", err,
    cuserr.WithLogger(zapAdapter),
)

// Service B uses Slog
errB := cuserr.NewInternalError("serviceB", err,
    cuserr.WithLogger(slogAdapter),
)
```

### Q: How do I test code with logger injection?

**A**: Create a test logger or use NoOpLogger:

```go
func TestMyService(t *testing.T) {
    // Option 1: Use NoOpLogger (no logs in tests)
    service := NewMyService(&cuserr.NoOpLogger{})

    // Option 2: Use real logger to verify log output
    var buf bytes.Buffer
    logger := slog.New(slog.NewJSONHandler(&buf, nil))
    adapter := adapters.NewSlogAdapter(logger)
    service := NewMyService(adapter)

    // Now you can verify logs
    // assert.Contains(t, buf.String(), "expected log message")
}
```

### Q: Should I use WithContext everywhere?

**A**: Yes, if you have request context available. It automatically extracts:
- `request_id` - For request tracing
- `user_id` - For user-specific debugging
- `trace_id` - For distributed tracing

```go
// Good - extracts context automatically
err := cuserr.NewNotFoundError("user", id,
    cuserr.WithContext(ctx),
)

// Also good - explicit if context not available
err := cuserr.NewNotFoundError("user", id,
    cuserr.WithRequestID(requestID),
)
```

---

## Performance Considerations

### Before v0.3.0:
```
Error Creation: ~1,200 ns/op (896 B/op, 10 allocs/op)
```

### After v0.3.0:
```
Without Logger: ~1,200 ns/op (896 B/op, 10 allocs/op) - No change!
With Logger (NoOp): ~1,205 ns/op (896 B/op, 10 allocs/op) - Negligible
With Logger (Real): ~1,400 ns/op (1,100 B/op, 12 allocs/op) - Minimal overhead
```

**Takeaway**: No performance regression. Logger injection adds minimal overhead only when you use it.

---

## Troubleshooting

### Issue: "cannot use logger (type *slog.Logger) as type cuserr.Logger"

**Problem**: Trying to pass slog.Logger directly instead of using adapter.

**Solution**:
```go
// ❌ Wrong
err := cuserr.NewInternalError("db", dbErr,
    cuserr.WithLogger(slog.Default()), // Won't compile
)

// ✅ Correct
adapter := adapters.NewSlogAdapter(slog.Default())
err := cuserr.NewInternalError("db", dbErr,
    cuserr.WithLogger(adapter), // Works!
)
```

### Issue: "undefined: adapters"

**Problem**: Forgot to import adapters package.

**Solution**:
```go
import (
    "github.com/itsatony/go-cuserr"
    "github.com/itsatony/go-cuserr/adapters" // Add this
)
```

### Issue: Logger not producing output

**Problem**: Using NoOpLogger or logger not configured correctly.

**Debug**:
```go
// Check if logger is nil
if err.logger == nil {
    fmt.Println("No logger configured - using NoOpLogger")
}

// Verify logger is set
err := cuserr.NewInternalError("test", nil,
    cuserr.WithLogger(adapter),
)
err.Log(ctx) // Should produce output

// If no output, check your logger configuration
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo, // Make sure level allows errors
}))
```

---

## Summary

v0.3.0 is a **zero-risk upgrade**:

✅ **No breaking changes** - All code works as-is
✅ **Critical bugs fixed** - Race conditions resolved automatically
✅ **Optional features** - Adopt logger injection when ready
✅ **Zero-cost abstraction** - No overhead if you don't use logging
✅ **Incremental migration** - Change one service at a time

**Recommendation**: Upgrade immediately for bug fixes, then adopt new features incrementally as you refactor existing code.

---

## Need Help?

- **Examples**: See `examples/modern_api_usage.go` for complete demonstrations
- **API Reference**: Check the updated README.md for full documentation
- **Issues**: Open an issue on GitHub if you encounter problems

Happy migrating! 🚀
