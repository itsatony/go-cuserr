package cuserr

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrentMetadataAccess tests thread-safe metadata operations
func TestConcurrentMetadataAccess(t *testing.T) {
	err := NewCustomError(ErrInternal, nil, "test concurrent access")

	const numGoroutines = 100
	const numOperations = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // Writers + Readers

	// Start writers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				err.WithMetadata(key, value)

				// Small delay to increase chance of race conditions
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	// Start readers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Read all metadata
				metadata := err.GetAllMetadata()

				// Verify we got a copy, not the original
				metadata["test_key"] = "test_value"

				// Try to read a specific key
				key := fmt.Sprintf("key_%d_%d", id, j)
				err.GetMetadata(key)

				// Small delay
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	allMetadata := err.GetAllMetadata()
	expectedKeys := numGoroutines * numOperations

	if len(allMetadata) != expectedKeys {
		t.Errorf("Expected %d keys, got %d", expectedKeys, len(allMetadata))
	}

	// Verify the test key we added to the copy didn't affect original
	if _, exists := err.GetMetadata("test_key"); exists {
		t.Error("Modifying returned metadata copy should not affect original")
	}
}

// TestConcurrentErrorCreation tests concurrent error creation
func TestConcurrentErrorCreation(t *testing.T) {
	const numGoroutines = 50

	var wg sync.WaitGroup
	errors := make([]*CustomError, numGoroutines)

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Create error with metadata
			err := NewCustomError(ErrNotFound, nil, fmt.Sprintf("concurrent error %d", id))
			err.WithMetadata("goroutine_id", fmt.Sprintf("%d", id))
			err.WithRequestID(fmt.Sprintf("req-%d", id))

			errors[id] = err
		}(i)
	}

	wg.Wait()

	// Verify all errors were created correctly
	for i, err := range errors {
		if err == nil {
			t.Errorf("Error %d was not created", i)
			continue
		}

		if err.Category != ErrorCategoryNotFound {
			t.Errorf("Error %d has wrong category: %v", i, err.Category)
		}

		expectedMessage := fmt.Sprintf("concurrent error %d", i)
		if err.Message != expectedMessage {
			t.Errorf("Error %d has wrong message: %v", i, err.Message)
		}

		goroutineID, exists := err.GetMetadata("goroutine_id")
		if !exists || goroutineID != fmt.Sprintf("%d", i) {
			t.Errorf("Error %d has wrong goroutine_id metadata", i)
		}

		expectedRequestID := fmt.Sprintf("req-%d", i)
		if err.RequestID != expectedRequestID {
			t.Errorf("Error %d has wrong request ID: %v", i, err.RequestID)
		}
	}
}

// TestRaceConditionWithConfig tests concurrent config access
func TestRaceConditionWithConfig(t *testing.T) {
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	const numGoroutines = 20
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Start config readers
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				config := GetConfig()
				_ = config.EnableStackTrace
				_ = config.MaxStackDepth
				_ = config.ProductionMode
			}
		}()
	}

	// Start config writers
	configs := []*Config{
		{EnableStackTrace: true, MaxStackDepth: 10, ProductionMode: false},
		{EnableStackTrace: false, MaxStackDepth: 5, ProductionMode: true},
		{EnableStackTrace: true, MaxStackDepth: 15, ProductionMode: false},
	}

	for i := 0; i < numGoroutines; i++ {
		go func(configIndex int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				SetConfig(configs[configIndex%len(configs)])
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	wg.Wait()

	// Verify config is still valid
	finalConfig := GetConfig()
	if finalConfig == nil {
		t.Error("Final config should not be nil")
	}
}

// TestConcurrentJSONSerialization tests concurrent JSON operations
func TestConcurrentJSONSerialization(t *testing.T) {
	err := NewCustomError(ErrInternal, nil, "concurrent JSON test")
	err.WithRequestID("req-123")

	const numGoroutines = 30
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Start JSON generators
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				jsonData := err.ToJSON()
				if jsonData == nil {
					t.Error("ToJSON() should not return nil")
					return
				}

				// Verify basic structure
				if _, exists := jsonData[JSON_FIELD_ERROR]; !exists {
					t.Error("JSON should contain error field")
					return
				}
			}
		}()
	}

	// Start metadata modifiers (while JSON is being generated)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				key := fmt.Sprintf("json_test_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				err.WithMetadata(key, value)
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentStackTraceAccess tests concurrent stack trace operations
func TestConcurrentStackTraceAccess(t *testing.T) {
	// Ensure stack trace is enabled for this test
	originalConfig := GetConfig()
	defer SetConfig(originalConfig)

	SetConfig(&Config{
		EnableStackTrace: true,
		MaxStackDepth:    5,
		ProductionMode:   false,
	})

	const numGoroutines = 25
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make([]*CustomError, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Create error (this captures stack trace)
			err := NewCustomError(ErrTimeout, nil, fmt.Sprintf("stack test %d", id))

			// Access stack trace methods concurrently
			go func() {
				_ = err.GetStackTrace()
				_ = err.GetStackTraceString()
				_ = err.FilterStackTrace("testing")
			}()

			errors[id] = err
		}(i)
	}

	wg.Wait()

	// Verify all errors have stack traces
	for i, err := range errors {
		if err == nil {
			t.Errorf("Error %d was not created", i)
			continue
		}

		stackTrace := err.GetStackTrace()
		if len(stackTrace) == 0 {
			t.Errorf("Error %d should have stack trace", i)
		}

		stackString := err.GetStackTraceString()
		if len(stackString) == 0 {
			t.Errorf("Error %d should have non-empty stack trace string", i)
		}
	}
}

// BenchmarkConcurrentMetadataOperations benchmarks concurrent metadata access
func BenchmarkConcurrentMetadataOperations(b *testing.B) {
	err := NewCustomError(ErrInternal, nil, "benchmark test")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", counter)
			value := fmt.Sprintf("bench_value_%d", counter)

			err.WithMetadata(key, value)
			_, _ = err.GetMetadata(key)
			_ = err.GetAllMetadata()

			counter++
		}
	})
}

// BenchmarkConcurrentErrorCreation benchmarks concurrent error creation
func BenchmarkConcurrentErrorCreation(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			message := fmt.Sprintf("benchmark error %d", counter)
			err := NewCustomError(ErrInternal, nil, message)
			err.WithMetadata("counter", fmt.Sprintf("%d", counter))
			_ = err.Error()
			counter++
		}
	})
}
