package cuserr

import (
	"errors"
	"fmt"
	"testing"
)

// BenchmarkNewCustomError benchmarks error creation
func BenchmarkNewCustomError(b *testing.B) {
	sentinel := ErrNotFound
	wrapped := errors.New("wrapped error")
	message := "benchmark test error"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCustomError(sentinel, wrapped, message)
	}
}

// BenchmarkNewCustomErrorWithCategory benchmarks error creation with category
func BenchmarkNewCustomErrorWithCategory(b *testing.B) {
	category := ErrorCategoryNotFound
	code := "BENCH_TEST"
	message := "benchmark test error"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCustomErrorWithCategory(category, code, message)
	}
}

// BenchmarkWithMetadata benchmarks metadata addition
func BenchmarkWithMetadata(b *testing.B) {
	err := NewCustomError(ErrInternal, nil, "benchmark metadata test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		err.WithMetadata(key, value)
	}
}

// BenchmarkGetMetadata benchmarks metadata retrieval
func BenchmarkGetMetadata(b *testing.B) {
	err := NewCustomError(ErrInternal, nil, "benchmark metadata test")

	// Pre-populate with metadata
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		err.WithMetadata(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%100)
		_, _ = err.GetMetadata(key)
	}
}

// BenchmarkGetAllMetadata benchmarks getting all metadata
func BenchmarkGetAllMetadata(b *testing.B) {
	err := NewCustomError(ErrInternal, nil, "benchmark metadata test")

	// Pre-populate with metadata
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		err.WithMetadata(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.GetAllMetadata()
	}
}

// BenchmarkErrorString benchmarks error string generation
func BenchmarkErrorString(b *testing.B) {
	wrapped := errors.New("wrapped error for benchmarking")
	err := NewCustomError(ErrNotFound, wrapped, "benchmark error message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

// BenchmarkDetailedError benchmarks detailed error string generation
func BenchmarkDetailedError(b *testing.B) {
	wrapped := errors.New("wrapped error for benchmarking")
	err := NewCustomError(ErrNotFound, wrapped, "benchmark detailed error")
	err.WithMetadata("user_id", "123")
	err.WithMetadata("action", "get_user")
	err.WithRequestID("req-456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.DetailedError()
	}
}

// BenchmarkShortError benchmarks short error string generation
func BenchmarkShortError(b *testing.B) {
	err := NewCustomError(ErrNotFound, nil, "benchmark short error")
	err.WithRequestID("req-789")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.ShortError()
	}
}

// BenchmarkToJSON benchmarks JSON serialization
func BenchmarkToJSON(b *testing.B) {
	err := NewCustomError(ErrInternal, nil, "benchmark JSON serialization")
	err.WithMetadata("user_id", "123")
	err.WithMetadata("action", "create_user")
	err.WithRequestID("req-json-test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.ToJSON()
	}
}

// BenchmarkToJSONString benchmarks JSON string serialization
func BenchmarkToJSONString(b *testing.B) {
	err := NewCustomError(ErrInternal, nil, "benchmark JSON string serialization")
	err.WithMetadata("user_id", "123")
	err.WithMetadata("action", "create_user")
	err.WithRequestID("req-json-string-test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.ToJSONString()
	}
}

// BenchmarkToHTTPStatus benchmarks HTTP status code mapping
func BenchmarkToHTTPStatus(b *testing.B) {
	err := NewCustomError(ErrNotFound, nil, "benchmark HTTP status")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.ToHTTPStatus()
	}
}

// BenchmarkStackTraceCapture benchmarks stack trace capture
func BenchmarkStackTraceCapture(b *testing.B) {
	// Enable stack trace for this benchmark
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	SetConfig(&Config{
		EnableStackTrace: true,
		MaxStackDepth:    10,
		ProductionMode:   false,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCustomError(ErrInternal, nil, "stack trace benchmark")
	}
}

// BenchmarkStackTraceDisabled benchmarks error creation with stack trace disabled
func BenchmarkStackTraceDisabled(b *testing.B) {
	// Disable stack trace for this benchmark
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	SetConfig(&Config{
		EnableStackTrace: false,
		MaxStackDepth:    10,
		ProductionMode:   false,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCustomError(ErrInternal, nil, "no stack trace benchmark")
	}
}

// BenchmarkGetStackTraceString benchmarks stack trace string generation
func BenchmarkGetStackTraceString(b *testing.B) {
	// Enable stack trace
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	SetConfig(&Config{
		EnableStackTrace: true,
		MaxStackDepth:    10,
		ProductionMode:   false,
	})

	err := NewCustomError(ErrInternal, nil, "stack trace string benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.GetStackTraceString()
	}
}

// BenchmarkIsErrorCategory benchmarks category checking
func BenchmarkIsErrorCategory(b *testing.B) {
	err := NewCustomError(ErrNotFound, nil, "category benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsErrorCategory(err, ErrorCategoryNotFound)
	}
}

// BenchmarkIsErrorCode benchmarks error code checking
func BenchmarkIsErrorCode(b *testing.B) {
	err := NewCustomError(ErrNotFound, nil, "code benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsErrorCode(err, ERROR_CODE_NOT_FOUND)
	}
}

// BenchmarkErrorsIs benchmarks errors.Is checking
func BenchmarkErrorsIs(b *testing.B) {
	wrapped := errors.New("wrapped error")
	err := NewCustomError(ErrNotFound, wrapped, "errors.Is benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = errors.Is(err, ErrNotFound)
	}
}

// BenchmarkErrorsAs benchmarks errors.As checking
func BenchmarkErrorsAs(b *testing.B) {
	err := NewCustomError(ErrNotFound, nil, "errors.As benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var customErr *CustomError
		_ = errors.As(err, &customErr)
	}
}

// BenchmarkWrapWithCustomError benchmarks error wrapping
func BenchmarkWrapWithCustomError(b *testing.B) {
	originalErr := errors.New("original error for wrapping")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WrapWithCustomError(
			originalErr,
			ErrorCategoryInternal,
			"WRAP_BENCH",
			"wrapped for benchmark")
	}
}

// Comparative benchmarks against standard errors

// BenchmarkStandardError benchmarks standard error creation
func BenchmarkStandardError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = errors.New("standard error for comparison")
	}
}

// BenchmarkStandardErrorf benchmarks standard fmt.Errorf
func BenchmarkStandardErrorf(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fmt.Errorf("formatted error %d", i)
	}
}

// BenchmarkStandardErrorWrap benchmarks standard error wrapping
func BenchmarkStandardErrorWrap(b *testing.B) {
	baseErr := errors.New("base error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fmt.Errorf("wrapped: %w", baseErr)
	}
}

// Memory allocation benchmarks

// BenchmarkMemoryAllocation benchmarks memory usage of CustomError
func BenchmarkMemoryAllocation(b *testing.B) {
	// Disable stack trace to focus on core structure
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	SetConfig(&Config{
		EnableStackTrace: false,
		MaxStackDepth:    0,
		ProductionMode:   true,
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := NewCustomError(ErrInternal, nil, "memory allocation test")
		err.WithMetadata("test", "value")
		_ = err.Error()
	}
}
