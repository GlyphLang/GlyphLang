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
				// Should have error from recovery
				if err == nil {
					t.Error("Expected error from panic recovery, got nil")
				}

				// Response should be 500
				if w.Code != http.StatusInternalServerError {
					t.Errorf("Expected status 500, got %d", w.Code)
				}

				// Response should be JSON error (generic for security)
				var resp map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}

				// Error should be boolean true
				if resp["error"] != true {
					t.Errorf("Response error = %v, want true", resp["error"])
				}
				// Message should be generic (security: don't expose panic details)
				if resp["message"] != "Internal Server Error" {
					t.Errorf("Response message = %v, want Internal Server Error", resp["message"])
				}
				// Code should be 500
				if resp["code"] != float64(500) {
					t.Errorf("Response code = %v, want 500", resp["code"])
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
// and that panic details are NOT exposed to clients (security fix)
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

	// Verify error response returns generic message (security: don't expose panic details)
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	// Error should be a boolean true
	if resp["error"] != true {
		t.Errorf("Expected error=true, got: %v", resp["error"])
	}

	// Message should be generic "Internal Server Error" (not expose panic details)
	if resp["message"] != "Internal Server Error" {
		t.Errorf("Expected generic error message, got: %v", resp["message"])
	}

	// Code should be 500
	if resp["code"] != float64(500) {
		t.Errorf("Expected code=500, got: %v", resp["code"])
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
			// Note: handler continues after timeout but writes are discarded
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
		// Check response code, not ctx.StatusCode (which may be modified by handler after timeout)
		if w.Code != 504 {
			t.Errorf("Expected status 504, got response %d", w.Code)
		}
	})
}

// TestSecurityHeadersMiddleware tests that security headers are added
func TestSecurityHeadersMiddleware(t *testing.T) {
	middleware := SecurityHeadersMiddleware()
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

	// Verify all security headers are set
	tests := []struct {
		header string
		value  string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := w.Header().Get(tt.header)
			if got != tt.value {
				t.Errorf("Header %s = %q, want %q", tt.header, got, tt.value)
			}
		})
	}
}

// TestCORSMiddleware_WildcardSecurity tests CORS wildcard security
func TestCORSMiddleware_WildcardSecurity(t *testing.T) {
	t.Run("wildcard sets literal asterisk", func(t *testing.T) {
		middleware := CORSMiddleware([]string{"*"})
		handler := func(ctx *Context) error {
			return nil
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://attacker.com")
		w := httptest.NewRecorder()
		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		wrappedHandler := middleware(handler)
		wrappedHandler(ctx)

		// Should set literal "*" not reflect the origin
		corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if corsOrigin != "*" {
			t.Errorf("Expected literal '*', got %q", corsOrigin)
		}

		// Credentials should be explicitly disabled when using wildcard
		corsCredentials := w.Header().Get("Access-Control-Allow-Credentials")
		if corsCredentials != "false" {
			t.Errorf("Expected credentials to be 'false' with wildcard, got %q", corsCredentials)
		}
	})

	t.Run("specific origin reflects origin", func(t *testing.T) {
		middleware := CORSMiddleware([]string{"http://trusted.com"})
		handler := func(ctx *Context) error {
			return nil
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://trusted.com")
		w := httptest.NewRecorder()
		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		wrappedHandler := middleware(handler)
		wrappedHandler(ctx)

		corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if corsOrigin != "http://trusted.com" {
			t.Errorf("Expected reflected origin, got %q", corsOrigin)
		}
	})

	t.Run("disallowed origin gets no headers", func(t *testing.T) {
		middleware := CORSMiddleware([]string{"http://trusted.com"})
		handler := func(ctx *Context) error {
			return nil
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://untrusted.com")
		w := httptest.NewRecorder()
		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		wrappedHandler := middleware(handler)
		wrappedHandler(ctx)

		corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if corsOrigin != "" {
			t.Errorf("Expected no CORS header for disallowed origin, got %q", corsOrigin)
		}
	})
}

// TestGetClientIP tests IP extraction from requests
func TestGetClientIP(t *testing.T) {
	// Configure 10.0.0.1 as a trusted proxy for proxy header tests
	SetTrustedProxies([]string{"10.0.0.1"})
	defer SetTrustedProxies(nil)

	tests := []struct {
		name          string
		remoteAddr    string
		xForwardedFor string
		xRealIP       string
		trustProxy    bool
		expectedIP    string
	}{
		{
			name:       "use RemoteAddr when no proxy headers",
			remoteAddr: "192.168.1.100:12345",
			trustProxy: true,
			expectedIP: "192.168.1.100:12345",
		},
		{
			name:          "prefer X-Forwarded-For over RemoteAddr from trusted proxy",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.195",
			trustProxy:    true,
			expectedIP:    "203.0.113.195",
		},
		{
			name:          "X-Forwarded-For with multiple IPs uses first from trusted proxy",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.195, 70.41.3.18, 150.172.238.178",
			trustProxy:    true,
			expectedIP:    "203.0.113.195",
		},
		{
			name:       "prefer X-Real-IP over RemoteAddr from trusted proxy",
			remoteAddr: "10.0.0.1:12345",
			xRealIP:    "203.0.113.100",
			trustProxy: true,
			expectedIP: "203.0.113.100",
		},
		{
			name:          "prefer X-Forwarded-For over X-Real-IP from trusted proxy",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.195",
			xRealIP:       "203.0.113.100",
			trustProxy:    true,
			expectedIP:    "203.0.113.195",
		},
		{
			name:          "ignore proxy headers from untrusted proxy IP",
			remoteAddr:    "192.168.1.50:12345",
			xForwardedFor: "203.0.113.195",
			xRealIP:       "203.0.113.100",
			trustProxy:    true,
			expectedIP:    "192.168.1.50:12345",
		},
		{
			name:          "ignore X-Forwarded-For when trustProxy is false",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.195",
			trustProxy:    false,
			expectedIP:    "10.0.0.1:12345",
		},
		{
			name:       "ignore X-Real-IP when trustProxy is false",
			remoteAddr: "10.0.0.1:12345",
			xRealIP:    "203.0.113.100",
			trustProxy: false,
			expectedIP: "10.0.0.1:12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getClientIP(req, tt.trustProxy)
			if ip != tt.expectedIP {
				t.Errorf("getClientIP() = %q, want %q", ip, tt.expectedIP)
			}
		})
	}
}

// TestAuthRateLimiting tests auth rate limiting with lockout
func TestAuthRateLimiting(t *testing.T) {
	config := AuthRateLimitConfig{
		MaxFailures:     3,
		LockoutDuration: 100 * time.Millisecond,
		MaxLockout:      500 * time.Millisecond,
		ResetAfter:      1 * time.Second,
	}

	validTokens := map[string]bool{
		"valid-token": true,
	}

	middleware := BasicAuthMiddlewareWithConfig(validTokens, config)
	handler := func(ctx *Context) error {
		ctx.StatusCode = 200
		return nil
	}

	makeRequest := func(token string) int {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		wrappedHandler := middleware(handler)
		wrappedHandler(ctx)
		if w.Code != 0 && w.Code != 200 {
			return w.Code
		}
		return ctx.StatusCode
	}

	// First 2 failures should just return 401
	for i := 0; i < 2; i++ {
		status := makeRequest("invalid-token")
		if status != 401 {
			t.Errorf("Failure %d: expected 401, got %d", i+1, status)
		}
	}

	// Third failure should trigger lockout
	status := makeRequest("invalid-token")
	if status != 401 {
		t.Errorf("Third failure: expected 401, got %d", status)
	}

	// Fourth attempt should be rate limited (429)
	status = makeRequest("valid-token")
	if status != 429 {
		t.Errorf("After lockout: expected 429, got %d", status)
	}

	// Wait for lockout to expire
	time.Sleep(150 * time.Millisecond)

	// Should work again with valid token
	status = makeRequest("valid-token")
	if status != 200 {
		t.Errorf("After lockout expires: expected 200, got %d", status)
	}
}

// TestDefaultAuthRateLimitConfig tests default config values
func TestDefaultAuthRateLimitConfig(t *testing.T) {
	config := DefaultAuthRateLimitConfig()

	if config.MaxFailures != 5 {
		t.Errorf("MaxFailures = %d, want 5", config.MaxFailures)
	}
	if config.LockoutDuration != 1*time.Minute {
		t.Errorf("LockoutDuration = %v, want 1 minute", config.LockoutDuration)
	}
	if config.MaxLockout != 15*time.Minute {
		t.Errorf("MaxLockout = %v, want 15 minutes", config.MaxLockout)
	}
	if config.ResetAfter != 15*time.Minute {
		t.Errorf("ResetAfter = %v, want 15 minutes", config.ResetAfter)
	}
}
