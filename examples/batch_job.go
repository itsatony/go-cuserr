//go:build batch_job
// +build batch_job

// Package main demonstrates go-cuserr usage in batch jobs, cron tasks, and background workers.
// This example shows how protocol-agnostic error handling works in non-interactive contexts.
//
// Build and run:
//
//	go run -tags batch_job examples/batch_job.go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/itsatony/go-cuserr"
)

// JobResult represents the result of a batch job operation
type JobResult struct {
	ID        string
	Success   bool
	Error     error
	Duration  time.Duration
	Processed int
	Failed    int
}

// DataRecord represents a record to be processed
type DataRecord struct {
	ID    string
	Email string
	Data  map[string]string
}

// BatchProcessor processes records in batches
type BatchProcessor struct {
	name         string
	batchSize    int
	errorCount   int
	successCount int
	logger       *SimpleLogger
	jobID        string
	startTime    time.Time
}

// SimpleLogger for demonstration
type SimpleLogger struct{}

func (l *SimpleLogger) Log(level, message string, fields map[string]interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s: %s", timestamp, level, message)
	if len(fields) > 0 {
		fmt.Printf(" | fields=%+v", fields)
	}
	fmt.Println()
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(name string, batchSize int) *BatchProcessor {
	return &BatchProcessor{
		name:      name,
		batchSize: batchSize,
		logger:    &SimpleLogger{},
		jobID:     fmt.Sprintf("job-%d", time.Now().Unix()),
		startTime: time.Now(),
	}
}

// ProcessRecords processes a batch of records
func (bp *BatchProcessor) ProcessRecords(ctx context.Context, records []DataRecord) *JobResult {
	result := &JobResult{
		ID:      bp.jobID,
		Success: true,
	}

	startTime := time.Now()

	bp.logger.Log("INFO", "Starting batch processing", map[string]interface{}{
		"job_id":       bp.jobID,
		"batch_size":   bp.batchSize,
		"record_count": len(records),
	})

	// Process each record
	for i, record := range records {
		operationID := fmt.Sprintf("%s-record-%d", bp.jobID, i)

		if err := bp.processRecord(ctx, record, operationID); err != nil {
			result.Failed++
			bp.errorCount++

			// Log error with cuserr metadata
			if customErr, ok := err.(*cuserr.CustomError); ok {
				bp.logCustomError(customErr, "ERROR")
			} else {
				bp.logger.Log("ERROR", fmt.Sprintf("Failed to process record: %v", err), map[string]interface{}{
					"job_id":       bp.jobID,
					"operation_id": operationID,
					"record_id":    record.ID,
				})
			}

			// Decide whether to continue or abort based on error category
			if customErr, ok := err.(*cuserr.CustomError); ok {
				if bp.shouldAbortBatch(customErr) {
					result.Success = false
					result.Error = cuserr.NewInternalError("batch-processor",
						fmt.Errorf("batch aborted due to critical error: %w", err),
						cuserr.WithOperationID(bp.jobID),
						cuserr.WithMetadata("total_processed", fmt.Sprintf("%d", i)),
						cuserr.WithMetadata("abort_reason", "critical_error"),
					)
					break
				}
			}
		} else {
			result.Processed++
			bp.successCount++
		}
	}

	result.Duration = time.Since(startTime)

	// Log job completion
	bp.logger.Log("INFO", "Batch processing completed", map[string]interface{}{
		"job_id":      bp.jobID,
		"success":     result.Success,
		"processed":   result.Processed,
		"failed":      result.Failed,
		"duration_ms": result.Duration.Milliseconds(),
	})

	return result
}

// processRecord processes a single record
func (bp *BatchProcessor) processRecord(ctx context.Context, record DataRecord, operationID string) error {
	// Simulate validation
	if record.Email == "" {
		err := cuserr.NewValidationError("email", "Email is required")
		err.OperationID = operationID
		cuserr.WithMetadata("job_id", bp.jobID)(err)
		cuserr.WithMetadata("record_id", record.ID)(err)
		cuserr.WithMetadata("batch_processor", bp.name)(err)
		return err
	}

	// Simulate external API call failure
	if record.Email == "fail@example.com" {
		return cuserr.NewExternalError("email-api", "validate",
			fmt.Errorf("API rate limit exceeded"),
			cuserr.WithOperationID(operationID),
			cuserr.WithMetadata("job_id", bp.jobID),
			cuserr.WithMetadata("record_id", record.ID),
			cuserr.WithMetadata("api_host", "api.example.com"),
			cuserr.WithMetadata("retry_after", "60"),
		)
	}

	// Simulate database error (critical - should abort batch)
	if record.Email == "critical@example.com" {
		return cuserr.NewInternalError("database",
			fmt.Errorf("connection pool exhausted"),
			cuserr.WithOperationID(operationID),
			cuserr.WithMetadata("job_id", bp.jobID),
			cuserr.WithMetadata("record_id", record.ID),
			cuserr.WithMetadata("database", "postgres"),
			cuserr.WithMetadata("severity", "critical"),
		)
	}

	// Simulate timeout
	if record.Email == "timeout@example.com" {
		err := cuserr.NewTimeoutError("processing", nil)
		err.OperationID = operationID
		cuserr.WithMetadata("job_id", bp.jobID)(err)
		cuserr.WithMetadata("record_id", record.ID)(err)
		cuserr.WithMetadata("operation", "data_enrichment")(err)
		cuserr.WithMetadata("timeout_duration", "30s")(err)
		return err
	}

	// Simulate successful processing
	time.Sleep(10 * time.Millisecond)
	return nil
}

// shouldAbortBatch determines if batch should be aborted based on error
func (bp *BatchProcessor) shouldAbortBatch(err *cuserr.CustomError) bool {
	// Abort on internal errors or if error rate is too high
	if err.Category == cuserr.ErrorCategoryInternal {
		return true
	}

	errorRate := float64(bp.errorCount) / float64(bp.errorCount+bp.successCount)
	return errorRate > 0.5 // Abort if >50% failure rate
}

// logCustomError logs a cuserr error with all its rich metadata
func (bp *BatchProcessor) logCustomError(err *cuserr.CustomError, level string) {
	fields := map[string]interface{}{
		"error_code":     err.Code,
		"error_category": string(err.Category),
		"operation_id":   err.OperationID,
		"job_id":         bp.jobID,
	}

	// Add all metadata
	for k, v := range err.GetAllMetadata() {
		fields[k] = v
	}

	// Add tracing if available
	if err.TraceID != "" {
		fields["trace_id"] = err.TraceID
	}
	if err.SpanID != "" {
		fields["span_id"] = err.SpanID
	}

	// Add syslog severity for log aggregation systems
	fields["syslog_severity"] = err.ToSyslogSeverity()

	bp.logger.Log(level, err.Message, fields)

	// Log wrapped error if present
	if err.Wrapped != nil {
		fields["wrapped_error"] = err.Wrapped.Error()
		bp.logger.Log(level, "Wrapped error details", fields)
	}
}

// CronJob represents a scheduled task
type CronJob struct {
	name      string
	schedule  string
	processor *BatchProcessor
	logger    *SimpleLogger
}

// NewCronJob creates a new cron job
func NewCronJob(name, schedule string) *CronJob {
	return &CronJob{
		name:      name,
		schedule:  schedule,
		processor: NewBatchProcessor(name, 100),
		logger:    &SimpleLogger{},
	}
}

// Execute runs the cron job
func (cj *CronJob) Execute(ctx context.Context) error {
	executionID := fmt.Sprintf("%s-%d", cj.name, time.Now().Unix())

	cj.logger.Log("INFO", "Cron job starting", map[string]interface{}{
		"job_name":     cj.name,
		"execution_id": executionID,
		"schedule":     cj.schedule,
	})

	// Simulate fetching records
	records := []DataRecord{
		{ID: "1", Email: "user1@example.com", Data: map[string]string{"status": "active"}},
		{ID: "2", Email: "user2@example.com", Data: map[string]string{"status": "active"}},
		{ID: "3", Email: "", Data: map[string]string{"status": "pending"}}, // Will cause validation error
		{ID: "4", Email: "user4@example.com", Data: map[string]string{"status": "active"}},
	}

	// Process the batch
	result := cj.processor.ProcessRecords(ctx, records)

	// Report results
	if !result.Success {
		return cuserr.NewInternalError("cron-job",
			result.Error,
			cuserr.WithOperationID(executionID),
			cuserr.WithMetadata("job_name", cj.name),
			cuserr.WithMetadata("processed", fmt.Sprintf("%d", result.Processed)),
			cuserr.WithMetadata("failed", fmt.Sprintf("%d", result.Failed)),
			cuserr.WithMetadata("duration_ms", fmt.Sprintf("%d", result.Duration.Milliseconds())),
		)
	}

	if result.Failed > 0 {
		// Partial success - log warning but don't fail the job
		cj.logger.Log("WARN", "Cron job completed with errors", map[string]interface{}{
			"job_name":     cj.name,
			"execution_id": executionID,
			"processed":    result.Processed,
			"failed":       result.Failed,
		})
	}

	return nil
}

// simulateBackgroundWorker demonstrates a background worker pattern
func simulateBackgroundWorker() {
	fmt.Println("\n=== Background Worker Pattern ===")

	ctx := context.Background()
	processor := NewBatchProcessor("background-worker", 50)

	// Simulate a queue of tasks
	tasks := []DataRecord{
		{ID: "task-1", Email: "worker1@example.com"},
		{ID: "task-2", Email: "fail@example.com"},    // Will fail
		{ID: "task-3", Email: "timeout@example.com"}, // Will timeout
		{ID: "task-4", Email: "worker4@example.com"},
		{ID: "task-5", Email: "critical@example.com"}, // Critical error
	}

	result := processor.ProcessRecords(ctx, tasks)

	fmt.Printf("\nWorker Results:\n")
	fmt.Printf("  Processed: %d\n", result.Processed)
	fmt.Printf("  Failed: %d\n", result.Failed)
	fmt.Printf("  Duration: %v\n", result.Duration)
	fmt.Printf("  Success: %v\n", result.Success)

	if result.Error != nil {
		if customErr, ok := result.Error.(*cuserr.CustomError); ok {
			fmt.Printf("\nBatch aborted with error:\n")
			fmt.Printf("  Code: %s\n", customErr.Code)
			fmt.Printf("  Category: %s\n", customErr.Category)
			fmt.Printf("  Message: %s\n", customErr.Message)
			fmt.Printf("  Syslog Severity: %d\n", customErr.ToSyslogSeverity())

			// For batch jobs, exit code is important for monitoring
			fmt.Printf("  Exit Code: %d (for shell scripts)\n", customErr.ToExitCode())
		}
	}
}

// simulateDataMigration demonstrates a data migration scenario
func simulateDataMigration() {
	fmt.Println("\n=== Data Migration Pattern ===")

	ctx := context.Background()
	migrationID := fmt.Sprintf("migration-%d", time.Now().Unix())

	// Simulate migration steps
	steps := []struct {
		name string
		fn   func(context.Context, string) error
	}{
		{"Validate source data", validateSourceData},
		{"Transform records", transformRecords},
		{"Load to destination", loadToDestination},
		{"Verify migration", verifyMigration},
	}

	logger := &SimpleLogger{}
	logger.Log("INFO", "Starting data migration", map[string]interface{}{
		"migration_id": migrationID,
		"steps":        len(steps),
	})

	for i, step := range steps {
		operationID := fmt.Sprintf("%s-step-%d", migrationID, i+1)

		logger.Log("INFO", fmt.Sprintf("Executing step %d: %s", i+1, step.name), map[string]interface{}{
			"migration_id": migrationID,
			"operation_id": operationID,
		})

		if err := step.fn(ctx, operationID); err != nil {
			// Migration failed - log error and exit
			if customErr, ok := err.(*cuserr.CustomError); ok {
				logger.Log("ERROR", "Migration failed", map[string]interface{}{
					"migration_id":   migrationID,
					"operation_id":   operationID,
					"step":           step.name,
					"error_code":     customErr.Code,
					"error_category": string(customErr.Category),
					"exit_code":      customErr.ToExitCode(),
				})

				fmt.Printf("\nMigration failed at step '%s'\n", step.name)
				fmt.Printf("Exit code: %d\n", customErr.ToExitCode())
				return
			}
		}
	}

	logger.Log("INFO", "Migration completed successfully", map[string]interface{}{
		"migration_id": migrationID,
	})
}

// Migration step functions
func validateSourceData(ctx context.Context, operationID string) error {
	// Simulate validation
	return nil
}

func transformRecords(ctx context.Context, operationID string) error {
	// Simulate transformation
	return nil
}

func loadToDestination(ctx context.Context, operationID string) error {
	// Simulate a conflict during load
	err := cuserr.NewConflictError("destination_table", "primary_key", "PK-123")
	err.OperationID = operationID
	cuserr.WithMetadata("migration_step", "load")(err)
	cuserr.WithMetadata("table", "users")(err)
	cuserr.WithMetadata("records_loaded", "1543")(err)
	return err
}

func verifyMigration(ctx context.Context, operationID string) error {
	// Simulate verification
	return nil
}

func main() {
	fmt.Println("=== Batch Job Example: go-cuserr Protocol-Agnostic Error Handling ===\n")

	// Example 1: Simple batch processing
	fmt.Println("Example 1: Basic Batch Processing")
	ctx := context.Background()
	processor := NewBatchProcessor("email-processor", 100)

	records := []DataRecord{
		{ID: "1", Email: "alice@example.com"},
		{ID: "2", Email: "bob@example.com"},
		{ID: "3", Email: ""}, // Invalid
		{ID: "4", Email: "carol@example.com"},
		{ID: "5", Email: "fail@example.com"}, // API failure
	}

	result := processor.ProcessRecords(ctx, records)
	fmt.Printf("\nBatch Results: Processed=%d, Failed=%d, Duration=%v\n",
		result.Processed, result.Failed, result.Duration)

	// Example 2: Cron job execution
	fmt.Println("\n\nExample 2: Cron Job Execution")
	cronJob := NewCronJob("daily-user-sync", "0 2 * * *")
	if err := cronJob.Execute(ctx); err != nil {
		if customErr, ok := err.(*cuserr.CustomError); ok {
			fmt.Printf("Cron job failed: %s\n", customErr.Message)
			fmt.Printf("Syslog severity: %d\n", customErr.ToSyslogSeverity())
		}
	}

	// Example 3: Background worker
	simulateBackgroundWorker()

	// Example 4: Data migration
	simulateDataMigration()

	// Example 5: Protocol-agnostic error in batch context
	fmt.Println("\n=== Protocol-Agnostic in Batch Jobs ===")
	batchErr := cuserr.NewExternalError("sftp-server", "upload",
		fmt.Errorf("connection refused"),
		cuserr.WithOperationID("batch-upload-789"),
		cuserr.WithMetadata("job_type", "file_transfer"),
		cuserr.WithMetadata("destination", "sftp.example.com"),
	)

	customErr := batchErr
	fmt.Printf("\nSame error, multiple protocol mappings:\n")
	fmt.Printf("  Exit Code (for scripts):  %d\n", customErr.ToExitCode())
	fmt.Printf("  Syslog Severity:          %d\n", customErr.ToSyslogSeverity())
	fmt.Printf("  HTTP Status (if exposed): %d\n", customErr.ToHTTPStatus())
	fmt.Printf("  gRPC Code (if exposed):   %d\n", customErr.ToGRPCCode())

	fmt.Println("\n=== Summary ===")
	fmt.Println("✓ Same error type works in batch jobs, cron tasks, and workers")
	fmt.Println("✓ Rich metadata for debugging and monitoring")
	fmt.Println("✓ Syslog severity for log aggregation systems")
	fmt.Println("✓ Exit codes for shell script integration")
	fmt.Println("✓ Operation IDs for tracking across distributed systems")
	fmt.Println("✓ Error categorization helps decide retry/abort logic")
	fmt.Println("✓ Protocol-agnostic = works everywhere without modification")
}
