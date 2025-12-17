package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRecoveryMiddleware_PanicRecovery tests that panics are caught and converted to 500 errors
func TestRecoveryMiddleware_PanicRecovery(t *testing.T) {
	tests := []struct {
		name        string
		panicValue  interface{}
		expectError bool
	}{
		{
			name:        "panic with string",
			panicValue:  "something went wrong",
			expectError: true,
		},
		{
			name:        "panic with error",
			panicValue:  errors.New("database connection failed"),
			expectError: true,
		},
		{
			name:        "panic with nil",
			panicValue:  nil,
			expectError: false,
		},
		{
			name:        "panic with integer",
			panicValue:  42,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a handler that panics
			panicHandler := func(ctx *Context) error {
				if tt.panicValue != nil {
					panic(tt.panicValue)
				}
				return nil
			}

			// Wrap with recovery middleware
			middleware := RecoveryMiddleware()
			wrappedHandler := middleware(panicHandler)

			// Create test request and response
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			ctx := &Context{
				Request:        req,
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
			}

			// Execute handler
			err := wrappedHandler(ctx)

			if tt.expectError {
				// Should have error from SendHTTPError
				if err == nil {
					t.Error("Expected error from panic recovery, got nil")
				}

				// Response should be 500
				if w.Code != http.StatusInternalServerError {
					t.Errorf("Expected status 500, got %d", w.Code)
				}

				// Response should be JSON error
				var resp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}

				if resp.Status != http.StatusInternalServerError {
					t.Errorf("Response status = %d, want %d", resp.Status, http.StatusInternalServerError)
				}
				if resp.Error != "InternalError" {
					t.Errorf("Response error type = %s, want InternalError", resp.Error)
				}
			} else {
				// No panic, should succeed
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestRecoveryMiddleware_NormalExecution tests that non-panicking handlers work normally
func TestRecoveryMiddleware_NormalExecution(t *testing.T) {
	// Create a normal handler that doesn't panic
	normalHandler := func(ctx *Context) error {
		return SendJSON(ctx, http.StatusOK, map[string]string{
			"message": "success",
		})
	}

	// Wrap with recovery middleware
	middleware := RecoveryMiddleware()
	wrappedHandler := middleware(normalHandler)

	// Create test request and response
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	// Execute handler
	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["message"] != "success" {
		t.Errorf("Expected message 'success', got '%s'", resp["message"])
	}
}

// TestRecoveryMiddleware_ErrorHandling tests that regular errors are passed through
func TestRecoveryMiddleware_ErrorHandling(t *testing.T) {
	// Create a handler that returns an error (not panic)
	errorHandler := func(ctx *Context) error {
		return NewValidationError("field", "invalid value")
	}

	// Wrap with recovery middleware
	middleware := RecoveryMiddleware()
	wrappedHandler := middleware(errorHandler)

	// Create test request and response
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	// Execute handler
	err := wrappedHandler(ctx)

	// Should get the original error, not a panic recovery error
	if err == nil {
		t.Fatal("Expected error from handler, got nil")
	}

	if _, ok := err.(*ValidationError); !ok {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

// TestRecoveryMiddleware_ChainedMiddlewares tests recovery works with other middlewares
func TestRecoveryMiddleware_ChainedMiddlewares(t *testing.T) {
	// Create a handler that panics
	panicHandler := func(ctx *Context) error {
		panic("critical error")
	}

	// Chain recovery with logging
	recovery := RecoveryMiddleware()
	logging := LoggingMiddleware()

	// Apply middlewares: logging -> recovery -> handler
	wrappedHandler := logging(recovery(panicHandler))

	// Create test request and response
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	// Execute handler - should not crash
	err := wrappedHandler(ctx)

	// Should have recovered from panic
	if err == nil {
		t.Error("Expected error from panic recovery, got nil")
	}

	// Response should be 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// TestRecoveryMiddleware_RequestContext tests that request context is logged
func TestRecoveryMiddleware_RequestContext(t *testing.T) {
	// Create a handler that panics
	panicHandler := func(ctx *Context) error {
		panic("test panic for context logging")
	}

	// Wrap with recovery middleware
	middleware := RecoveryMiddleware()
	wrappedHandler := middleware(panicHandler)

	// Create test request with various headers
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("User-Agent", "TestClient/1.0")
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.100:12345"

	w := httptest.NewRecorder()
	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	// Execute handler
	err := wrappedHandler(ctx)

	// Should have recovered
	if err == nil {
		t.Error("Expected error from panic recovery, got nil")
	}

	// Verify error response
	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	// Details should mention panic recovery
	if !strings.Contains(resp.Details, "panic recovered") {
		t.Errorf("Expected details to mention panic recovery, got: %s", resp.Details)
	}
}

// TestLoggingMiddleware tests request/response logging
func TestLoggingMiddleware(t *testing.T) {
	handler := func(ctx *Context) error {
		return SendJSON(ctx, http.StatusOK, map[string]string{"status": "ok"})
	}

	middleware := LoggingMiddleware()
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestChainMiddlewares tests middleware chaining
func TestChainMiddlewares(t *testing.T) {
	callOrder := []string{}

	middleware1 := func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			callOrder = append(callOrder, "m1-before")
			err := next(ctx)
			callOrder = append(callOrder, "m1-after")
			return err
		}
	}

	middleware2 := func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			callOrder = append(callOrder, "m2-before")
			err := next(ctx)
			callOrder = append(callOrder, "m2-after")
			return err
		}
	}

	handler := func(ctx *Context) error {
		callOrder = append(callOrder, "handler")
		return nil
	}

	chained := ChainMiddlewares(middleware1, middleware2)
	wrappedHandler := chained(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	wrappedHandler(ctx)

	expectedOrder := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
	if len(callOrder) != len(expectedOrder) {
		t.Fatalf("Call order length mismatch: got %v, want %v", callOrder, expectedOrder)
	}

	for i, expected := range expectedOrder {
		if callOrder[i] != expected {
			t.Errorf("Call order[%d] = %s, want %s", i, callOrder[i], expected)
		}
	}
}
