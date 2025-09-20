package cuserr

import (
	"fmt"
	"strconv"
	"time"
)

// Common metadata field names as constants for type safety and IDE support
const (
	// Identity fields
	MetaUserID    = "user_id"
	MetaRequestID = "request_id"
	MetaSessionID = "session_id"
	MetaTraceID   = "trace_id"
	MetaSpanID    = "span_id"

	// Operation context
	MetaOperation = "operation"
	MetaComponent = "component"
	MetaService   = "service"
	MetaEndpoint  = "endpoint"
	MetaMethod    = "method"

	// Resource identification
	MetaResource   = "resource"
	MetaResourceID = "resource_id"
	MetaField      = "field"
	MetaEntity     = "entity"

	// Error context
	MetaErrorType    = "error_type"
	MetaFailurePoint = "failure_point"
	MetaRetryCount   = "retry_count"
	MetaAttempt      = "attempt"

	// External service context
	MetaExternalService = "external_service"
	MetaURL             = "url"
	MetaStatusCode      = "status_code"
	MetaResponseTime    = "response_time"

	// Validation context
	MetaValidationField = "validation_field"
	MetaValidationValue = "validation_value"
	MetaValidationRule  = "validation_rule"

	// Security context
	MetaPermission = "permission"
	MetaRole       = "role"
	MetaScope      = "scope"
	MetaIPAddress  = "ip_address"
	MetaUserAgent  = "user_agent"

	// Performance context
	MetaDuration    = "duration"
	MetaMemoryUsage = "memory_usage"
	MetaCPUUsage    = "cpu_usage"
	MetaQueueSize   = "queue_size"

	// Business context
	MetaTenantID       = "tenant_id"
	MetaOrganizationID = "organization_id"
	MetaAccountID      = "account_id"
	MetaProjectID      = "project_id"
)

// TypedMetadata provides type-safe metadata operations
type TypedMetadata struct {
	err *CustomError
}

// NewTypedMetadata wraps a CustomError for typed metadata operations
func NewTypedMetadata(err *CustomError) *TypedMetadata {
	return &TypedMetadata{err: err}
}

// GetTypedMetadata returns a typed metadata wrapper for the error
func (e *CustomError) GetTypedMetadata() *TypedMetadata {
	return NewTypedMetadata(e)
}

// Identity metadata methods

// WithUserID adds user ID metadata with type safety
func (tm *TypedMetadata) WithUserID(userID string) *TypedMetadata {
	tm.err.WithMetadata(MetaUserID, userID)
	return tm
}

// GetUserID retrieves user ID from metadata
func (tm *TypedMetadata) GetUserID() (string, bool) {
	return tm.err.GetMetadata(MetaUserID)
}

// WithRequestID adds request ID metadata
func (tm *TypedMetadata) WithRequestID(requestID string) *TypedMetadata {
	tm.err.WithRequestID(requestID) // Use existing method
	return tm
}

// GetRequestID retrieves request ID
func (tm *TypedMetadata) GetRequestID() string {
	return tm.err.RequestID
}

// WithSessionID adds session ID metadata
func (tm *TypedMetadata) WithSessionID(sessionID string) *TypedMetadata {
	tm.err.WithMetadata(MetaSessionID, sessionID)
	return tm
}

// GetSessionID retrieves session ID from metadata
func (tm *TypedMetadata) GetSessionID() (string, bool) {
	return tm.err.GetMetadata(MetaSessionID)
}

// WithTraceID adds distributed tracing ID
func (tm *TypedMetadata) WithTraceID(traceID string) *TypedMetadata {
	tm.err.WithMetadata(MetaTraceID, traceID)
	return tm
}

// GetTraceID retrieves trace ID from metadata
func (tm *TypedMetadata) GetTraceID() (string, bool) {
	return tm.err.GetMetadata(MetaTraceID)
}

// Operation context methods

// WithOperation adds operation context
func (tm *TypedMetadata) WithOperation(operation string) *TypedMetadata {
	tm.err.WithMetadata(MetaOperation, operation)
	return tm
}

// GetOperation retrieves operation from metadata
func (tm *TypedMetadata) GetOperation() (string, bool) {
	return tm.err.GetMetadata(MetaOperation)
}

// WithComponent adds component context
func (tm *TypedMetadata) WithComponent(component string) *TypedMetadata {
	tm.err.WithMetadata(MetaComponent, component)
	return tm
}

// GetComponent retrieves component from metadata
func (tm *TypedMetadata) GetComponent() (string, bool) {
	return tm.err.GetMetadata(MetaComponent)
}

// WithService adds service context
func (tm *TypedMetadata) WithService(service string) *TypedMetadata {
	tm.err.WithMetadata(MetaService, service)
	return tm
}

// GetService retrieves service from metadata
func (tm *TypedMetadata) GetService() (string, bool) {
	return tm.err.GetMetadata(MetaService)
}

// WithEndpoint adds HTTP endpoint context
func (tm *TypedMetadata) WithEndpoint(endpoint string) *TypedMetadata {
	tm.err.WithMetadata(MetaEndpoint, endpoint)
	return tm
}

// GetEndpoint retrieves endpoint from metadata
func (tm *TypedMetadata) GetEndpoint() (string, bool) {
	return tm.err.GetMetadata(MetaEndpoint)
}

// WithHTTPMethod adds HTTP method context
func (tm *TypedMetadata) WithHTTPMethod(method string) *TypedMetadata {
	tm.err.WithMetadata(MetaMethod, method)
	return tm
}

// GetHTTPMethod retrieves HTTP method from metadata
func (tm *TypedMetadata) GetHTTPMethod() (string, bool) {
	return tm.err.GetMetadata(MetaMethod)
}

// Resource identification methods

// WithResource adds resource type context
func (tm *TypedMetadata) WithResource(resource string) *TypedMetadata {
	tm.err.WithMetadata(MetaResource, resource)
	return tm
}

// GetResource retrieves resource from metadata
func (tm *TypedMetadata) GetResource() (string, bool) {
	return tm.err.GetMetadata(MetaResource)
}

// WithResourceID adds resource ID context
func (tm *TypedMetadata) WithResourceID(resourceID string) *TypedMetadata {
	tm.err.WithMetadata(MetaResourceID, resourceID)
	return tm
}

// GetResourceID retrieves resource ID from metadata
func (tm *TypedMetadata) GetResourceID() (string, bool) {
	return tm.err.GetMetadata(MetaResourceID)
}

// WithField adds field context for validation errors
func (tm *TypedMetadata) WithField(field string) *TypedMetadata {
	tm.err.WithMetadata(MetaField, field)
	return tm
}

// GetField retrieves field from metadata
func (tm *TypedMetadata) GetField() (string, bool) {
	return tm.err.GetMetadata(MetaField)
}

// WithEntity adds entity context
func (tm *TypedMetadata) WithEntity(entity string) *TypedMetadata {
	tm.err.WithMetadata(MetaEntity, entity)
	return tm
}

// GetEntity retrieves entity from metadata
func (tm *TypedMetadata) GetEntity() (string, bool) {
	return tm.err.GetMetadata(MetaEntity)
}

// Error context methods

// WithErrorType adds error type classification
func (tm *TypedMetadata) WithErrorType(errorType string) *TypedMetadata {
	tm.err.WithMetadata(MetaErrorType, errorType)
	return tm
}

// GetErrorType retrieves error type from metadata
func (tm *TypedMetadata) GetErrorType() (string, bool) {
	return tm.err.GetMetadata(MetaErrorType)
}

// WithFailurePoint adds failure point context
func (tm *TypedMetadata) WithFailurePoint(point string) *TypedMetadata {
	tm.err.WithMetadata(MetaFailurePoint, point)
	return tm
}

// GetFailurePoint retrieves failure point from metadata
func (tm *TypedMetadata) GetFailurePoint() (string, bool) {
	return tm.err.GetMetadata(MetaFailurePoint)
}

// WithRetryCount adds retry count information
func (tm *TypedMetadata) WithRetryCount(count int) *TypedMetadata {
	tm.err.WithMetadata(MetaRetryCount, strconv.Itoa(count))
	return tm
}

// GetRetryCount retrieves retry count from metadata
func (tm *TypedMetadata) GetRetryCount() (int, bool) {
	if value, exists := tm.err.GetMetadata(MetaRetryCount); exists {
		if count, err := strconv.Atoi(value); err == nil {
			return count, true
		}
	}
	return 0, false
}

// WithAttempt adds attempt number
func (tm *TypedMetadata) WithAttempt(attempt int) *TypedMetadata {
	tm.err.WithMetadata(MetaAttempt, strconv.Itoa(attempt))
	return tm
}

// GetAttempt retrieves attempt number from metadata
func (tm *TypedMetadata) GetAttempt() (int, bool) {
	if value, exists := tm.err.GetMetadata(MetaAttempt); exists {
		if attempt, err := strconv.Atoi(value); err == nil {
			return attempt, true
		}
	}
	return 0, false
}

// External service context methods

// WithExternalService adds external service context
func (tm *TypedMetadata) WithExternalService(service string) *TypedMetadata {
	tm.err.WithMetadata(MetaExternalService, service)
	return tm
}

// GetExternalService retrieves external service from metadata
func (tm *TypedMetadata) GetExternalService() (string, bool) {
	return tm.err.GetMetadata(MetaExternalService)
}

// WithURL adds URL context
func (tm *TypedMetadata) WithURL(url string) *TypedMetadata {
	tm.err.WithMetadata(MetaURL, url)
	return tm
}

// GetURL retrieves URL from metadata
func (tm *TypedMetadata) GetURL() (string, bool) {
	return tm.err.GetMetadata(MetaURL)
}

// WithStatusCode adds HTTP status code context
func (tm *TypedMetadata) WithStatusCode(statusCode int) *TypedMetadata {
	tm.err.WithMetadata(MetaStatusCode, strconv.Itoa(statusCode))
	return tm
}

// GetStatusCode retrieves HTTP status code from metadata
func (tm *TypedMetadata) GetStatusCode() (int, bool) {
	if value, exists := tm.err.GetMetadata(MetaStatusCode); exists {
		if code, err := strconv.Atoi(value); err == nil {
			return code, true
		}
	}
	return 0, false
}

// WithResponseTime adds response time context
func (tm *TypedMetadata) WithResponseTime(duration time.Duration) *TypedMetadata {
	tm.err.WithMetadata(MetaResponseTime, duration.String())
	return tm
}

// GetResponseTime retrieves response time from metadata
func (tm *TypedMetadata) GetResponseTime() (time.Duration, bool) {
	if value, exists := tm.err.GetMetadata(MetaResponseTime); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration, true
		}
	}
	return 0, false
}

// Validation context methods

// WithValidationField adds validation field context
func (tm *TypedMetadata) WithValidationField(field string) *TypedMetadata {
	tm.err.WithMetadata(MetaValidationField, field)
	return tm
}

// GetValidationField retrieves validation field from metadata
func (tm *TypedMetadata) GetValidationField() (string, bool) {
	return tm.err.GetMetadata(MetaValidationField)
}

// WithValidationValue adds the invalid value for context
func (tm *TypedMetadata) WithValidationValue(value string) *TypedMetadata {
	tm.err.WithMetadata(MetaValidationValue, value)
	return tm
}

// GetValidationValue retrieves validation value from metadata
func (tm *TypedMetadata) GetValidationValue() (string, bool) {
	return tm.err.GetMetadata(MetaValidationValue)
}

// WithValidationRule adds the violated validation rule
func (tm *TypedMetadata) WithValidationRule(rule string) *TypedMetadata {
	tm.err.WithMetadata(MetaValidationRule, rule)
	return tm
}

// GetValidationRule retrieves validation rule from metadata
func (tm *TypedMetadata) GetValidationRule() (string, bool) {
	return tm.err.GetMetadata(MetaValidationRule)
}

// Security context methods

// WithPermission adds required permission context
func (tm *TypedMetadata) WithPermission(permission string) *TypedMetadata {
	tm.err.WithMetadata(MetaPermission, permission)
	return tm
}

// GetPermission retrieves permission from metadata
func (tm *TypedMetadata) GetPermission() (string, bool) {
	return tm.err.GetMetadata(MetaPermission)
}

// WithRole adds user role context
func (tm *TypedMetadata) WithRole(role string) *TypedMetadata {
	tm.err.WithMetadata(MetaRole, role)
	return tm
}

// GetRole retrieves role from metadata
func (tm *TypedMetadata) GetRole() (string, bool) {
	return tm.err.GetMetadata(MetaRole)
}

// WithScope adds authorization scope context
func (tm *TypedMetadata) WithScope(scope string) *TypedMetadata {
	tm.err.WithMetadata(MetaScope, scope)
	return tm
}

// GetScope retrieves scope from metadata
func (tm *TypedMetadata) GetScope() (string, bool) {
	return tm.err.GetMetadata(MetaScope)
}

// WithIPAddress adds client IP address
func (tm *TypedMetadata) WithIPAddress(ipAddress string) *TypedMetadata {
	tm.err.WithMetadata(MetaIPAddress, ipAddress)
	return tm
}

// GetIPAddress retrieves IP address from metadata
func (tm *TypedMetadata) GetIPAddress() (string, bool) {
	return tm.err.GetMetadata(MetaIPAddress)
}

// WithUserAgent adds user agent string
func (tm *TypedMetadata) WithUserAgent(userAgent string) *TypedMetadata {
	tm.err.WithMetadata(MetaUserAgent, userAgent)
	return tm
}

// GetUserAgent retrieves user agent from metadata
func (tm *TypedMetadata) GetUserAgent() (string, bool) {
	return tm.err.GetMetadata(MetaUserAgent)
}

// Performance context methods

// WithDuration adds operation duration
func (tm *TypedMetadata) WithDuration(duration time.Duration) *TypedMetadata {
	tm.err.WithMetadata(MetaDuration, duration.String())
	return tm
}

// GetDuration retrieves duration from metadata
func (tm *TypedMetadata) GetDuration() (time.Duration, bool) {
	if value, exists := tm.err.GetMetadata(MetaDuration); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration, true
		}
	}
	return 0, false
}

// WithMemoryUsage adds memory usage context
func (tm *TypedMetadata) WithMemoryUsage(bytes int64) *TypedMetadata {
	tm.err.WithMetadata(MetaMemoryUsage, fmt.Sprintf("%d", bytes))
	return tm
}

// GetMemoryUsage retrieves memory usage from metadata
func (tm *TypedMetadata) GetMemoryUsage() (int64, bool) {
	if value, exists := tm.err.GetMetadata(MetaMemoryUsage); exists {
		if bytes, err := strconv.ParseInt(value, 10, 64); err == nil {
			return bytes, true
		}
	}
	return 0, false
}

// Business context methods

// WithTenantID adds tenant/organization context for multi-tenancy
func (tm *TypedMetadata) WithTenantID(tenantID string) *TypedMetadata {
	tm.err.WithMetadata(MetaTenantID, tenantID)
	return tm
}

// GetTenantID retrieves tenant ID from metadata
func (tm *TypedMetadata) GetTenantID() (string, bool) {
	return tm.err.GetMetadata(MetaTenantID)
}

// WithOrganizationID adds organization context
func (tm *TypedMetadata) WithOrganizationID(orgID string) *TypedMetadata {
	tm.err.WithMetadata(MetaOrganizationID, orgID)
	return tm
}

// GetOrganizationID retrieves organization ID from metadata
func (tm *TypedMetadata) GetOrganizationID() (string, bool) {
	return tm.err.GetMetadata(MetaOrganizationID)
}

// WithAccountID adds account context
func (tm *TypedMetadata) WithAccountID(accountID string) *TypedMetadata {
	tm.err.WithMetadata(MetaAccountID, accountID)
	return tm
}

// GetAccountID retrieves account ID from metadata
func (tm *TypedMetadata) GetAccountID() (string, bool) {
	return tm.err.GetMetadata(MetaAccountID)
}

// WithProjectID adds project context
func (tm *TypedMetadata) WithProjectID(projectID string) *TypedMetadata {
	tm.err.WithMetadata(MetaProjectID, projectID)
	return tm
}

// GetProjectID retrieves project ID from metadata
func (tm *TypedMetadata) GetProjectID() (string, bool) {
	return tm.err.GetMetadata(MetaProjectID)
}

// Fluent interface methods that return the original error for chaining

// Error returns the underlying CustomError for further chaining
func (tm *TypedMetadata) Error() *CustomError {
	return tm.err
}

// Convenience methods that combine typed metadata with error creation

// NewValidationErrorWithTypedMetadata creates a validation error with typed metadata builder
func NewValidationErrorWithTypedMetadata(field, message string) (*CustomError, *TypedMetadata) {
	err := NewValidationError(field, message)
	tm := err.GetTypedMetadata().WithValidationField(field).WithErrorType("validation")
	return err, tm
}

// NewNotFoundErrorWithTypedMetadata creates a not found error with typed metadata builder
func NewNotFoundErrorWithTypedMetadata(resource, id string) (*CustomError, *TypedMetadata) {
	err := NewNotFoundError(resource, id)
	tm := err.GetTypedMetadata().WithResource(resource).WithResourceID(id).WithErrorType("not_found")
	return err, tm
}

// NewUnauthorizedErrorWithTypedMetadata creates an unauthorized error with typed metadata builder
func NewUnauthorizedErrorWithTypedMetadata(reason string) (*CustomError, *TypedMetadata) {
	err := NewUnauthorizedError(reason)
	tm := err.GetTypedMetadata().WithErrorType("unauthorized")
	return err, tm
}

// NewForbiddenErrorWithTypedMetadata creates a forbidden error with typed metadata builder
func NewForbiddenErrorWithTypedMetadata(action, resource string) (*CustomError, *TypedMetadata) {
	err := NewForbiddenError(action, resource)
	tm := err.GetTypedMetadata().WithResource(resource).WithOperation(action).WithErrorType("forbidden")
	return err, tm
}

// NewInternalErrorWithTypedMetadata creates an internal error with typed metadata builder
func NewInternalErrorWithTypedMetadata(component string, wrapped error) (*CustomError, *TypedMetadata) {
	err := NewInternalError(component, wrapped)
	tm := err.GetTypedMetadata().WithComponent(component).WithErrorType("internal")
	return err, tm
}

// NewExternalErrorWithTypedMetadata creates an external error with typed metadata builder
func NewExternalErrorWithTypedMetadata(service, operation string, wrapped error) (*CustomError, *TypedMetadata) {
	err := NewExternalError(service, operation, wrapped)
	tm := err.GetTypedMetadata().WithExternalService(service).WithOperation(operation).WithErrorType("external")
	return err, tm
}
