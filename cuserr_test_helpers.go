package cuserr

// Common test context keys as strings (library expects string keys)
// The library uses string keys internally, so tests must match.
const (
	testKeyRequestID string = "request_id"
	testKeyUserID    string = "user_id"
	testKeyTraceID   string = "trace_id"
)
