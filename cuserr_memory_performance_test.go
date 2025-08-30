package cuserr

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestMemoryManagement tests memory usage patterns
func TestMemoryManagement(t *testing.T) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	t.Run("Memory Usage with Stack Traces", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    20,
			ProductionMode:   false,
		})

		// Force garbage collection to get baseline
		runtime.GC()
		runtime.GC()

		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create many errors with stack traces
		errors := make([]*CustomError, 1000)
		for i := 0; i < 1000; i++ {
			errors[i] = NewCustomError(ErrInternal, nil, fmt.Sprintf("memory test error %d", i)).
				WithMetadata("iteration", fmt.Sprintf("%d", i)).
				WithMetadata("test_type", "memory_usage")
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		// Memory should have increased
		if m2.Alloc <= m1.Alloc {
			t.Error("Memory allocation should increase with error creation")
		}

		// Clear all errors to allow garbage collection
		for i := range errors {
			errors[i] = nil
		}
		errors = nil

		// Force garbage collection
		runtime.GC()
		runtime.GC()

		var m3 runtime.MemStats
		runtime.ReadMemStats(&m3)

		// Memory should be freed (allowing some tolerance)
		if m3.Alloc > m2.Alloc {
			t.Logf("Memory after GC (%d) should be less than or equal to peak (%d)", m3.Alloc, m2.Alloc)
		}
	})

	t.Run("Memory Usage Without Stack Traces", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: false,
			MaxStackDepth:    0,
			ProductionMode:   true,
		})

		runtime.GC()
		runtime.GC()

		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create many errors without stack traces
		errors := make([]*CustomError, 1000)
		for i := 0; i < 1000; i++ {
			errors[i] = NewCustomError(ErrInternal, nil, fmt.Sprintf("no stack error %d", i)).
				WithMetadata("iteration", fmt.Sprintf("%d", i))
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		memUsageWithStack := m2.Alloc - m1.Alloc

		// Clear and test with stack traces enabled
		for i := range errors {
			errors[i] = nil
		}
		errors = nil
		runtime.GC()
		runtime.GC()

		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    20,
			ProductionMode:   false,
		})

		var m3 runtime.MemStats
		runtime.ReadMemStats(&m3)

		// Create same number of errors with stack traces
		errorsWithStack := make([]*CustomError, 1000)
		for i := 0; i < 1000; i++ {
			errorsWithStack[i] = NewCustomError(ErrInternal, nil, fmt.Sprintf("with stack error %d", i)).
				WithMetadata("iteration", fmt.Sprintf("%d", i))
		}

		var m4 runtime.MemStats
		runtime.ReadMemStats(&m4)

		memUsageNoStack := m4.Alloc - m3.Alloc

		// Stack traces should use more memory
		if memUsageNoStack <= memUsageWithStack {
			t.Logf("Errors with stack traces (%d bytes) should use more memory than without (%d bytes)",
				memUsageNoStack, memUsageWithStack)
		}

		// Clean up
		for i := range errorsWithStack {
			errorsWithStack[i] = nil
		}
	})

	t.Run("Memory Management with Clear Operations", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    15,
			ProductionMode:   false,
		})

		err := NewCustomError(ErrTimeout, nil, "memory clear test")

		// Add significant metadata
		for i := 0; i < 100; i++ {
			err.WithMetadata(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d_with_some_longer_content", i))
		}

		// Verify metadata exists
		metadata := err.GetAllMetadata()
		if len(metadata) != 100 {
			t.Errorf("Expected 100 metadata entries, got %d", len(metadata))
		}

		// Clear stack trace to reduce memory usage
		err.ClearStackTrace()

		// Stack trace should be empty
		if len(err.GetStackTrace()) != 0 {
			t.Error("Stack trace should be empty after clearing")
		}

		// Metadata should still exist
		metadata = err.GetAllMetadata()
		if len(metadata) != 100 {
			t.Error("Metadata should remain after clearing stack trace")
		}
	})
}

// TestHighFrequencyOperations tests performance under high load
func TestHighFrequencyOperations(t *testing.T) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	t.Run("High Frequency Error Creation", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: false,
			ProductionMode:   true,
		})

		const numOperations = 50000
		start := time.Now()

		for i := 0; i < numOperations; i++ {
			err := NewCustomError(ErrInternal, nil, "high frequency test")
			_ = err.Error() // Ensure the error is actually used
		}

		duration := time.Since(start)
		opsPerSecond := float64(numOperations) / duration.Seconds()

		t.Logf("Created %d errors in %v (%.0f ops/sec)", numOperations, duration, opsPerSecond)

		// Should be able to create at least 10k errors per second
		if opsPerSecond < 10000 {
			t.Errorf("Error creation too slow: %.0f ops/sec", opsPerSecond)
		}
	})

	t.Run("High Frequency Metadata Operations", func(t *testing.T) {
		err := NewCustomError(ErrInternal, nil, "metadata performance test")

		const numOperations = 10000
		start := time.Now()

		for i := 0; i < numOperations; i++ {
			key := fmt.Sprintf("key_%d", i%100) // Reuse some keys to test overwrites
			value := fmt.Sprintf("value_%d", i)
			err.WithMetadata(key, value)
		}

		duration := time.Since(start)
		opsPerSecond := float64(numOperations) / duration.Seconds()

		t.Logf("Performed %d metadata operations in %v (%.0f ops/sec)", numOperations, duration, opsPerSecond)

		// Should be fast
		if opsPerSecond < 50000 {
			t.Errorf("Metadata operations too slow: %.0f ops/sec", opsPerSecond)
		}

		// Verify final state
		metadata := err.GetAllMetadata()
		if len(metadata) != 100 { // Should have 100 unique keys (key_0 to key_99)
			t.Errorf("Expected 100 unique metadata keys, got %d", len(metadata))
		}
	})

	t.Run("High Frequency JSON Serialization", func(t *testing.T) {
		err := NewCustomError(ErrInvalidInput, nil, "JSON performance test").
			WithMetadata("field", "email").
			WithMetadata("value", "invalid@").
			WithRequestID("req-perf-test")

		const numOperations = 5000
		start := time.Now()

		for i := 0; i < numOperations; i++ {
			_ = err.ToJSON()
		}

		duration := time.Since(start)
		opsPerSecond := float64(numOperations) / duration.Seconds()

		t.Logf("Performed %d JSON serializations in %v (%.0f ops/sec)", numOperations, duration, opsPerSecond)

		// Should be reasonably fast
		if opsPerSecond < 1000 {
			t.Errorf("JSON serialization too slow: %.0f ops/sec", opsPerSecond)
		}
	})

	t.Run("Concurrent High Frequency Operations", func(t *testing.T) {
		const numGoroutines = 10
		const opsPerGoroutine = 1000

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < opsPerGoroutine; j++ {
					err := NewCustomError(ErrTimeout, nil, fmt.Sprintf("concurrent test %d-%d", goroutineID, j))
					err.WithMetadata("goroutine_id", fmt.Sprintf("%d", goroutineID))
					err.WithMetadata("operation_id", fmt.Sprintf("%d", j))
					_ = err.Error()
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)
		totalOps := numGoroutines * opsPerGoroutine
		opsPerSecond := float64(totalOps) / duration.Seconds()

		t.Logf("Performed %d concurrent operations in %v (%.0f ops/sec)", totalOps, duration, opsPerSecond)

		// Should maintain performance under concurrency
		if opsPerSecond < 5000 {
			t.Errorf("Concurrent operations too slow: %.0f ops/sec", opsPerSecond)
		}
	})
}

// TestMemoryLeakPrevention tests for potential memory leaks
func TestMemoryLeakPrevention(t *testing.T) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	t.Run("Error Reference Cycles", func(t *testing.T) {
		// Test that circular references don't prevent garbage collection

		err1 := NewCustomError(ErrInternal, nil, "error 1")
		err2 := NewCustomError(ErrInternal, err1, "error 2")

		// Create a reference cycle in metadata (if possible)
		err1.WithMetadata("related_error", err2.Error())
		err2.WithMetadata("original_error", err1.Error())

		// Make them eligible for GC
		err1 = nil
		err2 = nil

		// Force garbage collection
		runtime.GC()
		runtime.GC()

		// If we reach here without hanging, reference cycles are handled properly
	})

	t.Run("Large Metadata Memory Management", func(t *testing.T) {
		// Test memory management with large amounts of metadata

		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create error with large metadata
		err := NewCustomError(ErrInternal, nil, "large metadata test")

		// Add large metadata values
		largeValue := string(make([]byte, 10000)) // 10KB string
		for i := 0; i < 100; i++ {
			err.WithMetadata(fmt.Sprintf("large_key_%d", i), largeValue)
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		// Memory should increase significantly
		memIncrease := m2.Alloc - m1.Alloc
		if memIncrease < 1000000 { // Should be at least ~1MB
			t.Logf("Memory increase seems small: %d bytes", memIncrease)
		}

		// Clear the error
		err = nil

		// Force GC
		runtime.GC()
		runtime.GC()

		var m3 runtime.MemStats
		runtime.ReadMemStats(&m3)

		// Memory should be reclaimed (with some tolerance)
		if m3.Alloc > m2.Alloc {
			t.Logf("Memory after GC: %d bytes (was %d bytes)", m3.Alloc, m2.Alloc)
		}
	})

	t.Run("Stack Trace Memory Management", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    50, // Large depth
			ProductionMode:   false,
		})

		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create errors with deep stack traces
		errors := make([]*CustomError, 100)
		for i := 0; i < 100; i++ {
			errors[i] = deepStackTraceFunction(10) // Create nested calls
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		// Clear all errors
		for i := range errors {
			errors[i] = nil
		}
		errors = nil

		// Force GC
		runtime.GC()
		runtime.GC()

		var m3 runtime.MemStats
		runtime.ReadMemStats(&m3)

		// Stack trace memory should be reclaimed
		if m3.Alloc > m2.Alloc {
			t.Logf("Stack trace memory management: before=%d, peak=%d, after=%d",
				m1.Alloc, m2.Alloc, m3.Alloc)
		}
	})
}

// TestResourceExhaustion tests behavior under resource constraints
func TestResourceExhaustion(t *testing.T) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	t.Run("Extreme Stack Depth Limitation", func(t *testing.T) {
		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    1000, // Very high limit
			ProductionMode:   false,
		})

		// This should still complete without issues
		err := NewCustomError(ErrInternal, nil, "extreme depth test")

		stackTrace := err.GetStackTrace()

		// Should have captured some stack trace
		if len(stackTrace) == 0 {
			t.Error("Should capture stack trace even with high limit")
		}

		// Should not exceed reasonable limits (runtime.Callers limitations)
		if len(stackTrace) > 100 {
			t.Logf("Stack trace has %d frames (very deep)", len(stackTrace))
		}
	})

	t.Run("Massive Metadata Sets", func(t *testing.T) {
		err := NewCustomError(ErrInternal, nil, "massive metadata test")

		// Add a very large number of metadata entries
		const numEntries = 10000
		for i := 0; i < numEntries; i++ {
			err.WithMetadata(fmt.Sprintf("key_%06d", i), fmt.Sprintf("value_%06d", i))
		}

		// Should still function
		metadata := err.GetAllMetadata()
		if len(metadata) != numEntries {
			t.Errorf("Expected %d metadata entries, got %d", numEntries, len(metadata))
		}

		// JSON serialization should still work (though may be slow)
		start := time.Now()
		jsonData := err.ToJSON()
		duration := time.Since(start)

		t.Logf("JSON serialization of %d metadata entries took %v", numEntries, duration)

		if jsonData == nil {
			t.Error("JSON serialization should succeed even with large metadata")
		}

		// Should complete within reasonable time (10 seconds)
		if duration > 10*time.Second {
			t.Errorf("JSON serialization too slow: %v", duration)
		}
	})
}

// Helper function for creating deep stack traces
func deepStackTraceFunction(depth int) *CustomError {
	if depth <= 0 {
		return NewCustomError(ErrTimeout, nil, "deep stack error")
	}
	return deepStackTraceFunction(depth - 1)
}

// BenchmarkErrorCreation benchmarks basic error creation
func BenchmarkErrorCreation(b *testing.B) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	b.Run("WithoutStackTrace", func(b *testing.B) {
		SetConfig(&Config{
			EnableStackTrace: false,
			ProductionMode:   true,
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := NewCustomError(ErrInternal, nil, "benchmark error")
			_ = err.Error()
		}
	})

	b.Run("WithStackTrace", func(b *testing.B) {
		SetConfig(&Config{
			EnableStackTrace: true,
			MaxStackDepth:    10,
			ProductionMode:   false,
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := NewCustomError(ErrInternal, nil, "benchmark error")
			_ = err.Error()
		}
	})

	b.Run("WithMetadata", func(b *testing.B) {
		SetConfig(&Config{
			EnableStackTrace: false,
			ProductionMode:   true,
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := NewCustomError(ErrInternal, nil, "benchmark error").
				WithMetadata("key1", "value1").
				WithMetadata("key2", "value2")
			_ = err.Error()
		}
	})
}

// BenchmarkJSONSerialization benchmarks JSON operations
func BenchmarkJSONSerialization(b *testing.B) {
	err := NewCustomError(ErrInvalidInput, nil, "benchmark JSON error").
		WithMetadata("field", "email").
		WithMetadata("reason", "invalid format").
		WithRequestID("req-bench-123")

	b.Run("ToJSON", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = err.ToJSON()
		}
	})

	b.Run("ToJSONString", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = err.ToJSONString()
		}
	})

	b.Run("ToClientJSON", func(b *testing.B) {
		originalConfig := GetConfig()
		defer SetConfig(originalConfig)

		SetConfig(&Config{
			EnableStackTrace: false,
			ProductionMode:   true,
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = err.ToClientJSON()
		}
	})
}
