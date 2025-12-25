package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

// TestCORSMiddleware tests CORS header handling
func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		expectAllow    bool
	}{
		{
			name:           "allowed origin",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:3000",
			expectAllow:    true,
		},
		{
			name:           "wildcard allows all",
			allowedOrigins: []string{"*"},
			requestOrigin:  "http://example.com",
			expectAllow:    true,
		},
		{
			name:           "disallowed origin",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://evil.com",
			expectAllow:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := CORSMiddleware(tt.allowedOrigins)
			handler := func(ctx *Context) error {
				return nil
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", tt.requestOrigin)
			w := httptest.NewRecorder()
			ctx := &Context{
				Request:        req,
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
			}

			wrappedHandler := middleware(handler)
			wrappedHandler(ctx)

			corsHeader := w.Header().Get("Access-Control-Allow-Origin")
			if tt.expectAllow && corsHeader == "" {
				t.Error("Expected CORS header to be set")
			}
		})
	}
}

// TestCORSMiddleware_Preflight tests CORS preflight handling
func TestCORSMiddleware_Preflight(t *testing.T) {
	middleware := CORSMiddleware([]string{"*"})
	handler := func(ctx *Context) error {
		return nil
	}

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	wrappedHandler := middleware(handler)
	wrappedHandler(ctx)

	if ctx.StatusCode != 204 {
		t.Errorf("Expected status 204 for preflight, got %d", ctx.StatusCode)
	}
}

// TestHeaderMiddleware tests custom header injection
func TestHeaderMiddleware(t *testing.T) {
	headers := map[string]string{
		"X-Custom-Header": "custom-value",
		"X-Another":       "another-value",
	}

	middleware := HeaderMiddleware(headers)
	handler := func(ctx *Context) error {
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	wrappedHandler := middleware(handler)
	wrappedHandler(ctx)

	for key, expected := range headers {
		got := w.Header().Get(key)
		if got != expected {
			t.Errorf("Header %s = %s, want %s", key, got, expected)
		}
	}
}

// TestAuthMiddleware tests authentication middleware
func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		authHeader   string
		validFunc    func(*Context) (bool, error)
		expectStatus int
	}{
		{
			name:         "missing auth header",
			authHeader:   "",
			validFunc:    nil,
			expectStatus: 401,
		},
		{
			name:       "valid token",
			authHeader: "Bearer valid-token",
			validFunc: func(ctx *Context) (bool, error) {
				return true, nil
			},
			expectStatus: 200,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid-token",
			validFunc: func(ctx *Context) (bool, error) {
				return false, nil
			},
			expectStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := AuthMiddleware(tt.validFunc)
			handler := func(ctx *Context) error {
				ctx.StatusCode = 200
				return nil
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			ctx := &Context{
				Request:        req,
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
			}

			wrappedHandler := middleware(handler)
			wrappedHandler(ctx)

			if w.Code != tt.expectStatus && ctx.StatusCode != tt.expectStatus {
				t.Errorf("Expected status %d, got response %d / ctx %d", tt.expectStatus, w.Code, ctx.StatusCode)
			}
		})
	}
}

// TestBasicAuthMiddleware tests basic token authentication
func TestBasicAuthMiddleware(t *testing.T) {
	validTokens := map[string]bool{
		"valid-token":   true,
		"another-token": true,
	}

	tests := []struct {
		name         string
		authHeader   string
		expectStatus int
	}{
		{
			name:         "missing header",
			authHeader:   "",
			expectStatus: 401,
		},
		{
			name:         "valid bearer token",
			authHeader:   "Bearer valid-token",
			expectStatus: 200,
		},
		{
			name:         "valid raw token",
			authHeader:   "valid-token",
			expectStatus: 200,
		},
		{
			name:         "invalid token",
			authHeader:   "Bearer invalid-token",
			expectStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := BasicAuthMiddleware(validTokens)
			handler := func(ctx *Context) error {
				ctx.StatusCode = 200
				return nil
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			ctx := &Context{
				Request:        req,
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
			}

			wrappedHandler := middleware(handler)
			wrappedHandler(ctx)

			if w.Code != tt.expectStatus && ctx.StatusCode != tt.expectStatus {
				t.Errorf("Expected status %d, got response %d / ctx %d", tt.expectStatus, w.Code, ctx.StatusCode)
			}
		})
	}
}

// TestRateLimitMiddleware tests rate limiting
func TestRateLimitMiddleware(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         5,
	}

	middleware := RateLimitMiddleware(config)
	handler := func(ctx *Context) error {
		ctx.StatusCode = 200
		return nil
	}

	// Make requests within burst limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		wrappedHandler := middleware(handler)
		wrappedHandler(ctx)
		if w.Code == 429 || ctx.StatusCode == 429 {
			t.Errorf("Request %d should succeed within burst, got rate limited", i)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	wrappedHandler := middleware(handler)
	wrappedHandler(ctx)
	if w.Code != 429 && ctx.StatusCode != 429 {
		t.Errorf("Expected status 429, got response %d / ctx %d", w.Code, ctx.StatusCode)
	}
}

// TestTimeoutMiddleware tests request timeout
func TestTimeoutMiddleware(t *testing.T) {
	middleware := TimeoutMiddleware(100 * time.Millisecond)

	t.Run("fast handler", func(t *testing.T) {
		handler := func(ctx *Context) error {
			ctx.StatusCode = 200
			return nil
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		wrappedHandler := middleware(handler)
		wrappedHandler(ctx)
		if w.Code == 504 || ctx.StatusCode == 504 {
			t.Error("Fast handler should succeed, got timeout")
		}
	})

	t.Run("slow handler", func(t *testing.T) {
		handler := func(ctx *Context) error {
			time.Sleep(200 * time.Millisecond)
			ctx.StatusCode = 200
			return nil
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		wrappedHandler := middleware(handler)
		wrappedHandler(ctx)
		if w.Code != 504 && ctx.StatusCode != 504 {
			t.Errorf("Expected status 504, got response %d / ctx %d", w.Code, ctx.StatusCode)
		}
	})
}
