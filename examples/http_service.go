// Package main demonstrates HTTP service integration with cuserr
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/itsatony/go-cuserr"
)

// User represents a user in our system
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UserService handles user-related operations
type UserService struct {
	users map[string]*User // Simple in-memory store
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]*User),
	}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
	// Simulate validation
	if userID == "" {
		return nil, cuserr.NewCustomError(
			cuserr.ErrInvalidInput,
			nil,
			"user ID is required").
			WithMetadata("field", "user_id").
			WithRequestID(getRequestIDFromContext(ctx))
	}

	// Look up user
	user, exists := s.users[userID]
	if !exists {
		return nil, cuserr.NewCustomError(
			cuserr.ErrNotFound,
			nil,
			"user not found").
			WithMetadata("user_id", userID).
			WithMetadata("operation", "get_user").
			WithRequestID(getRequestIDFromContext(ctx))
	}

	return user, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, email, name string) (*User, error) {
	// Input validation
	if email == "" {
		return nil, cuserr.NewCustomError(
			cuserr.ErrInvalidInput,
			nil,
			"email is required").
			WithMetadata("field", "email").
			WithRequestID(getRequestIDFromContext(ctx))
	}

	if name == "" {
		return nil, cuserr.NewCustomError(
			cuserr.ErrInvalidInput,
			nil,
			"name is required").
			WithMetadata("field", "name").
			WithRequestID(getRequestIDFromContext(ctx))
	}

	// Check if user already exists (simulate email uniqueness)
	for _, existingUser := range s.users {
		if existingUser.Email == email {
			return nil, cuserr.NewCustomError(
				cuserr.ErrAlreadyExists,
				nil,
				"user with this email already exists").
				WithMetadata("email", email).
				WithMetadata("existing_user_id", existingUser.ID).
				WithRequestID(getRequestIDFromContext(ctx))
		}
	}

	// Create new user
	userID := fmt.Sprintf("usr_%d", len(s.users)+1)
	user := &User{
		ID:    userID,
		Email: email,
		Name:  name,
	}

	s.users[userID] = user
	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, userID, email, name string) (*User, error) {
	// Get existing user first
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		// Re-wrap with update context
		if cuserr.IsErrorCategory(err, cuserr.ErrorCategoryNotFound) {
			return nil, cuserr.NewCustomError(
				cuserr.ErrNotFound,
				err,
				"cannot update non-existent user").
				WithMetadata("user_id", userID).
				WithMetadata("operation", "update_user").
				WithRequestID(getRequestIDFromContext(ctx))
		}
		return nil, err
	}

	// Update fields
	if email != "" {
		user.Email = email
	}
	if name != "" {
		user.Name = name
	}

	return user, nil
}

// HTTP Handlers

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	service *UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := extractPathParam(r, "id")
	ctx := r.Context()

	user, err := h.service.GetUser(ctx, userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSON(w, http.StatusOK, user)
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		customErr := cuserr.NewCustomError(
			cuserr.ErrInvalidInput,
			err,
			"invalid JSON payload").
			WithMetadata("content_type", r.Header.Get("Content-Type")).
			WithRequestID(getRequestIDFromContext(r.Context()))

		h.handleError(w, r, customErr)
		return
	}

	user, err := h.service.CreateUser(r.Context(), req.Email, req.Name)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, user)
}

// UpdateUser handles PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := extractPathParam(r, "id")

	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		customErr := cuserr.NewCustomError(
			cuserr.ErrInvalidInput,
			err,
			"invalid JSON payload").
			WithRequestID(getRequestIDFromContext(r.Context()))

		h.handleError(w, r, customErr)
		return
	}

	user, err := h.service.UpdateUser(r.Context(), userID, req.Email, req.Name)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.writeJSON(w, http.StatusOK, user)
}

// handleError handles custom errors and converts them to HTTP responses
func (h *UserHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var customErr *cuserr.CustomError

	if errors.As(err, &customErr) {
		// Log detailed error for debugging
		log.Printf("[ERROR] %s", customErr.DetailedError())

		// Send appropriate HTTP response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(customErr.ToHTTPStatus())

		// Use client-safe JSON in production
		config := cuserr.GetConfig()
		var response map[string]interface{}
		if config.ProductionMode {
			response = customErr.ToClientJSON()
		} else {
			response = customErr.ToJSON()
		}

		json.NewEncoder(w).Encode(response)
		return
	}

	// Handle unexpected errors
	log.Printf("[ERROR] Unexpected error: %v", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		},
	})
}

// writeJSON writes a JSON response
func (h *UserHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Middleware

// requestIDMiddleware adds a request ID to the context
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d", len(r.URL.Path)+int(r.ContentLength))
		}

		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// errorRecoveryMiddleware recovers from panics and converts them to errors
func errorRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("[PANIC] %v", recovered)

				err := cuserr.NewCustomError(
					cuserr.ErrInternal,
					fmt.Errorf("panic: %v", recovered),
					"internal server error occurred").
					WithMetadata("panic_value", fmt.Sprintf("%v", recovered)).
					WithRequestID(getRequestIDFromContext(r.Context()))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(err.ToClientJSON())
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Helper functions

// getRequestIDFromContext extracts request ID from context
func getRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// extractPathParam extracts path parameter (simplified implementation)
func extractPathParam(r *http.Request, param string) string {
	// In a real implementation, you'd use a router like gorilla/mux or similar
	// This is a simplified version for demonstration
	path := r.URL.Path

	// Extract ID from path like /users/123
	if param == "id" && len(path) > 7 { // "/users/" = 7 chars
		return path[7:]
	}

	return ""
}

// Demo server setup
func runHTTPServiceDemo() {
	fmt.Println("=== HTTP Service with cuserr Integration ===")

	// Configure cuserr for development
	cuserr.SetConfig(&cuserr.Config{
		EnableStackTrace: true,
		MaxStackDepth:    10,
		ProductionMode:   false,
	})

	// Create service and handler
	userService := NewUserService()
	userHandler := NewUserHandler(userService)

	// Add some sample users
	userService.users["usr_1"] = &User{ID: "usr_1", Email: "alice@example.com", Name: "Alice"}
	userService.users["usr_2"] = &User{ID: "usr_2", Email: "bob@example.com", Name: "Bob"}

	// Setup routes with middleware
	mux := http.NewServeMux()

	// Apply middleware
	handler := requestIDMiddleware(errorRecoveryMiddleware(mux))

	// Register routes
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.GetUser(w, r)
		case http.MethodPut:
			userHandler.UpdateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Demonstration endpoints
	mux.HandleFunc("/demo", demoErrorTypes)
	mux.HandleFunc("/health", healthCheck)

	fmt.Println("Server starting on :8080")
	fmt.Println("Try these endpoints:")
	fmt.Println("  GET  /users/usr_1     # Get existing user")
	fmt.Println("  GET  /users/usr_999   # User not found error")
	fmt.Println("  POST /users           # Create user")
	fmt.Println("  PUT  /users/usr_1     # Update user")
	fmt.Println("  GET  /demo            # Error type demonstration")
	fmt.Println("  GET  /health          # Health check")

	log.Fatal(http.ListenAndServe(":8080", handler))
}

// demoErrorTypes demonstrates different error types
func demoErrorTypes(w http.ResponseWriter, r *http.Request) {
	errorType := r.URL.Query().Get("type")
	ctx := r.Context()
	requestID := getRequestIDFromContext(ctx)

	var err error

	switch errorType {
	case "validation":
		err = cuserr.NewCustomError(cuserr.ErrInvalidInput, nil, "validation error example").
			WithMetadata("field", "email").
			WithRequestID(requestID)
	case "not_found":
		err = cuserr.NewCustomError(cuserr.ErrNotFound, nil, "resource not found example").
			WithMetadata("resource", "user").
			WithRequestID(requestID)
	case "unauthorized":
		err = cuserr.NewCustomError(cuserr.ErrUnauthorized, nil, "authentication required").
			WithMetadata("auth_method", "bearer_token").
			WithRequestID(requestID)
	case "conflict":
		err = cuserr.NewCustomError(cuserr.ErrAlreadyExists, nil, "resource conflict example").
			WithMetadata("resource", "email").
			WithRequestID(requestID)
	case "timeout":
		err = cuserr.NewCustomError(cuserr.ErrTimeout, nil, "operation timeout example").
			WithMetadata("operation", "database_query").
			WithMetadata("timeout", "30s").
			WithRequestID(requestID)
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"available_types": []string{"validation", "not_found", "unauthorized", "conflict", "timeout"},
			"usage":           "?type=validation",
		})
		return
	}

	// Handle the error like a real endpoint would
	var customErr *cuserr.CustomError
	if errors.As(err, &customErr) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(customErr.ToHTTPStatus())
		json.NewEncoder(w).Encode(customErr.ToJSON())
	}
}

// healthCheck provides a simple health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"service":   "cuserr-example",
		"timestamp": "now",
	})
}
