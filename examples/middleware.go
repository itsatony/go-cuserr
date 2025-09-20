//go:build middleware

// Package main demonstrates middleware integration with cuserr
//
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/itsatony/go-cuserr"
)

// ErrorMiddleware wraps HTTP handlers with comprehensive error handling
func ErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer to capture status codes
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		defer func() {
			// Handle panics and convert them to structured errors
			if recovered := recover(); recovered != nil {
				log.Printf("[PANIC] Request: %s %s, Panic: %v", r.Method, r.URL.Path, recovered)

				panicErr := cuserr.NewCustomError(
					cuserr.ErrInternal,
					fmt.Errorf("panic: %v", recovered),
					"internal server error occurred").
					WithMetadata("method", r.Method).
					WithMetadata("path", r.URL.Path).
					WithMetadata("panic_value", fmt.Sprintf("%v", recovered)).
					WithRequestID(GetRequestID(r.Context()))

				WriteErrorResponse(w, panicErr)
			}
		}()

		next.ServeHTTP(wrapper, r)
	})
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), len(r.URL.Path))
		}

		ctx := context.WithValue(r.Context(), "request_id", requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TimeoutMiddleware adds request timeout handling
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						log.Printf("[TIMEOUT_PANIC] %v", recovered)
					}
					close(done)
				}()
				next.ServeHTTP(w, r.WithContext(ctx))
			}()

			select {
			case <-done:
				// Request completed normally
			case <-ctx.Done():
				// Request timed out
				if ctx.Err() == context.DeadlineExceeded {
					timeoutErr := cuserr.NewCustomError(
						cuserr.ErrTimeout,
						ctx.Err(),
						"request timeout exceeded").
						WithMetadata("timeout", timeout.String()).
						WithMetadata("method", r.Method).
						WithMetadata("path", r.URL.Path).
						WithRequestID(GetRequestID(r.Context()))

					WriteErrorResponse(w, timeoutErr)
				}
			}
		})
	}
}

// RateLimitMiddleware provides basic rate limiting
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	// Simple in-memory rate limiter (not production ready)
	clientRequests := make(map[string][]time.Time)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			now := time.Now()

			// Clean old requests
			cutoff := now.Add(-time.Minute)
			var validRequests []time.Time
			for _, reqTime := range clientRequests[clientIP] {
				if reqTime.After(cutoff) {
					validRequests = append(validRequests, reqTime)
				}
			}

			// Check rate limit
			if len(validRequests) >= requestsPerMinute {
				rateLimitErr := cuserr.NewCustomError(
					cuserr.ErrRateLimit,
					nil,
					"rate limit exceeded").
					WithMetadata("client_ip", clientIP).
					WithMetadata("limit", fmt.Sprintf("%d/min", requestsPerMinute)).
					WithMetadata("current_requests", fmt.Sprintf("%d", len(validRequests))).
					WithRequestID(GetRequestID(r.Context()))

				WriteErrorResponse(w, rateLimitErr)
				return
			}

			// Record this request
			validRequests = append(validRequests, now)
			clientRequests[clientIP] = validRequests

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware provides authentication checking
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for API key in header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			unauthorizedErr := cuserr.NewCustomError(
				cuserr.ErrUnauthorized,
				nil,
				"API key required").
				WithMetadata("header", "X-API-Key").
				WithMetadata("path", r.URL.Path).
				WithRequestID(GetRequestID(r.Context()))

			WriteErrorResponse(w, unauthorizedErr)
			return
		}

		// Validate API key (simplified)
		if !isValidAPIKey(apiKey) {
			unauthorizedErr := cuserr.NewCustomError(
				cuserr.ErrUnauthorized,
				nil,
				"invalid API key").
				WithMetadata("api_key", maskAPIKey(apiKey)).
				WithRequestID(GetRequestID(r.Context()))

			WriteErrorResponse(w, unauthorizedErr)
			return
		}

		// Add user info to context
		userID := getUserIDFromAPIKey(apiKey)
		ctx := context.WithValue(r.Context(), "user_id", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs requests and responses with error details
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Log request
		log.Printf("[REQUEST] %s %s %s", r.Method, r.URL.Path, GetRequestID(r.Context()))

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)

		// Log response with error details if applicable
		if wrapper.statusCode >= 400 {
			log.Printf("[RESPONSE] %s %s %d %v (ERROR)",
				r.Method, r.URL.Path, wrapper.statusCode, duration)
		} else {
			log.Printf("[RESPONSE] %s %s %d %v",
				r.Method, r.URL.Path, wrapper.statusCode, duration)
		}
	})
}

// Helper types and functions

// responseWriter wraps http.ResponseWriter to capture status codes
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return "unknown"
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// WriteErrorResponse writes a cuserr.CustomError as HTTP response
func WriteErrorResponse(w http.ResponseWriter, err *cuserr.CustomError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.ToHTTPStatus())

	config := cuserr.GetConfig()
	var response map[string]interface{}
	if config.ProductionMode {
		response = err.ToClientJSON()
	} else {
		response = err.ToJSON()
	}

	json.NewEncoder(w).Encode(response)
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// isValidAPIKey validates an API key (simplified)
func isValidAPIKey(apiKey string) bool {
	// In a real implementation, this would check against a database
	validKeys := map[string]bool{
		"key_12345": true,
		"key_67890": true,
		"demo_key":  true,
	}

	return validKeys[apiKey]
}

// getUserIDFromAPIKey gets user ID from API key (simplified)
func getUserIDFromAPIKey(apiKey string) string {
	keyToUser := map[string]string{
		"key_12345": "user_alice",
		"key_67890": "user_bob",
		"demo_key":  "user_demo",
	}

	if userID, exists := keyToUser[apiKey]; exists {
		return userID
	}
	return "unknown_user"
}

// maskAPIKey masks an API key for logging
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 4 {
		return "****"
	}
	return apiKey[:4] + "****"
}

// Example handlers that demonstrate error usage

// HelloHandler demonstrates basic success response
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	requestID := GetRequestID(r.Context())

	response := map[string]interface{}{
		"message":    "Hello, World!",
		"user_id":    userID,
		"request_id": requestID,
		"timestamp":  time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ErrorDemoHandler demonstrates various error types
func ErrorDemoHandler(w http.ResponseWriter, r *http.Request) {
	errorType := r.URL.Query().Get("type")
	requestID := GetRequestID(r.Context())
	userID := GetUserID(r.Context())

	var err *cuserr.CustomError

	switch errorType {
	case "validation":
		err = cuserr.NewCustomError(cuserr.ErrInvalidInput, nil, "validation failed").
			WithMetadata("field", "email").
			WithMetadata("user_id", userID).
			WithRequestID(requestID)

	case "not_found":
		err = cuserr.NewCustomError(cuserr.ErrNotFound, nil, "resource not found").
			WithMetadata("resource", "user").
			WithMetadata("resource_id", "usr_404").
			WithRequestID(requestID)

	case "forbidden":
		err = cuserr.NewCustomError(cuserr.ErrForbidden, nil, "insufficient permissions").
			WithMetadata("required_role", "admin").
			WithMetadata("user_role", "user").
			WithMetadata("user_id", userID).
			WithRequestID(requestID)

	case "conflict":
		err = cuserr.NewCustomError(cuserr.ErrAlreadyExists, nil, "resource already exists").
			WithMetadata("resource", "email").
			WithMetadata("value", "test@example.com").
			WithRequestID(requestID)

	case "external":
		originalErr := errors.New("connection refused")
		err = cuserr.NewCustomError(cuserr.ErrExternal, originalErr, "external service unavailable").
			WithMetadata("service", "payment_gateway").
			WithMetadata("endpoint", "https://api.payment.com/charge").
			WithRequestID(requestID)

	case "internal":
		originalErr := errors.New("database connection failed")
		err = cuserr.NewCustomError(cuserr.ErrInternal, originalErr, "internal server error").
			WithMetadata("component", "database").
			WithRequestID(requestID)

	case "panic":
		// Simulate a panic (will be caught by ErrorMiddleware)
		panic("simulated panic for testing")

	case "slow":
		// Simulate slow operation (will trigger timeout if TimeoutMiddleware is used)
		time.Sleep(6 * time.Second)
		HelloHandler(w, r)
		return

	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"available_types": []string{
				"validation", "not_found", "forbidden", "conflict",
				"external", "internal", "panic", "slow"},
			"usage": "/error?type=validation",
		})
		return
	}

	WriteErrorResponse(w, err)
}

// Demo server setup
func runMiddlewareDemo() {
	fmt.Println("=== Middleware Integration Demo ===")

	// Configure cuserr
	cuserr.SetConfig(&cuserr.Config{
		EnableStackTrace: true,
		MaxStackDepth:    10,
		ProductionMode:   false,
	})

	// Create router
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/hello", HelloHandler)
	mux.HandleFunc("/error", ErrorDemoHandler)
	mux.HandleFunc("/protected", HelloHandler) // Will require auth

	// Chain middleware
	handler := LoggingMiddleware(
		RequestIDMiddleware(
			ErrorMiddleware(
				TimeoutMiddleware(5 * time.Second)(
					RateLimitMiddleware(10)( // 10 requests per minute
						mux)))))

	// Protected route with auth
	protectedMux := http.NewServeMux()
	protectedMux.Handle("/protected", AuthMiddleware(http.HandlerFunc(HelloHandler)))

	// Combine handlers
	finalMux := http.NewServeMux()
	finalMux.Handle("/hello", handler)
	finalMux.Handle("/error", handler)
	finalMux.Handle("/protected", LoggingMiddleware(RequestIDMiddleware(ErrorMiddleware(protectedMux))))

	fmt.Println("\nServer starting on :8081")
	fmt.Println("Try these endpoints:")
	fmt.Println("  GET  /hello                    # Basic hello (no auth required)")
	fmt.Println("  GET  /error?type=validation    # Validation error demo")
	fmt.Println("  GET  /error?type=not_found     # Not found error demo")
	fmt.Println("  GET  /error?type=panic         # Panic recovery demo")
	fmt.Println("  GET  /error?type=slow          # Timeout demo (>5s)")
	fmt.Println("  GET  /protected                # Protected route (needs X-API-Key header)")
	fmt.Println("\nValid API keys: key_12345, key_67890, demo_key")
	fmt.Println("Rate limit: 10 requests per minute per IP")
	fmt.Println("Timeout: 5 seconds")

	log.Fatal(http.ListenAndServe(":8081", finalMux))
}

// main starts the middleware integration demo
func main() {
	runMiddlewareDemo()
}
