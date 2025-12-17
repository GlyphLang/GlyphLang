package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestHealthManager_NewHealthManager tests the creation of a health manager
func TestHealthManager_NewHealthManager(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		hm := NewHealthManager()
		if hm == nil {
			t.Fatal("expected non-nil health manager")
		}
		if hm.timeout != 5*time.Second {
			t.Errorf("expected default timeout of 5s, got %v", hm.timeout)
		}
		if len(hm.checkers) != 0 {
			t.Errorf("expected 0 checkers, got %d", len(hm.checkers))
		}
	})

	t.Run("with custom timeout", func(t *testing.T) {
		customTimeout := 10 * time.Second
		hm := NewHealthManager(WithHealthCheckTimeout(customTimeout))
		if hm.timeout != customTimeout {
			t.Errorf("expected timeout of %v, got %v", customTimeout, hm.timeout)
		}
	})
}

// TestHealthManager_RegisterChecker tests registering health checkers
func TestHealthManager_RegisterChecker(t *testing.T) {
	hm := NewHealthManager()

	checker := NewStaticHealthChecker("test", StatusHealthy)
	hm.RegisterChecker(checker)

	if len(hm.checkers) != 1 {
		t.Errorf("expected 1 checker, got %d", len(hm.checkers))
	}

	if hm.checkers["test"] != checker {
		t.Error("checker not properly registered")
	}
}

// TestHealthManager_UnregisterChecker tests unregistering health checkers
func TestHealthManager_UnregisterChecker(t *testing.T) {
	hm := NewHealthManager()

	checker := NewStaticHealthChecker("test", StatusHealthy)
	hm.RegisterChecker(checker)

	hm.UnregisterChecker("test")

	if len(hm.checkers) != 0 {
		t.Errorf("expected 0 checkers after unregister, got %d", len(hm.checkers))
	}
}

// TestHealthManager_LivenessHandler tests the liveness endpoint
func TestHealthManager_LivenessHandler(t *testing.T) {
	hm := NewHealthManager()

	req := httptest.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()

	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	err := hm.LivenessHandler()(ctx)
	if err != nil {
		t.Fatalf("liveness handler returned error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != StatusHealthy {
		t.Errorf("expected status 'healthy', got '%s'", response.Status)
	}
}

// TestHealthManager_ReadinessHandler tests the readiness endpoint
func TestHealthManager_ReadinessHandler(t *testing.T) {
	t.Run("no checkers", func(t *testing.T) {
		hm := NewHealthManager()

		req := httptest.NewRequest("GET", "/health/ready", nil)
		w := httptest.NewRecorder()

		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		err := hm.ReadinessHandler()(ctx)
		if err != nil {
			t.Fatalf("readiness handler returned error: %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response HealthResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Status != StatusHealthy {
			t.Errorf("expected status 'healthy', got '%s'", response.Status)
		}
	})

	t.Run("all healthy checkers", func(t *testing.T) {
		hm := NewHealthManager()
		hm.RegisterChecker(NewStaticHealthChecker("db", StatusHealthy))
		hm.RegisterChecker(NewStaticHealthChecker("redis", StatusHealthy))

		req := httptest.NewRequest("GET", "/health/ready", nil)
		w := httptest.NewRecorder()

		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		err := hm.ReadinessHandler()(ctx)
		if err != nil {
			t.Fatalf("readiness handler returned error: %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response HealthResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Status != StatusHealthy {
			t.Errorf("expected status 'healthy', got '%s'", response.Status)
		}

		if len(response.Checks) != 2 {
			t.Errorf("expected 2 checks, got %d", len(response.Checks))
		}
	})

	t.Run("degraded checker", func(t *testing.T) {
		hm := NewHealthManager()
		hm.RegisterChecker(NewStaticHealthChecker("db", StatusHealthy))
		hm.RegisterChecker(NewStaticHealthChecker("redis", StatusDegraded))

		req := httptest.NewRequest("GET", "/health/ready", nil)
		w := httptest.NewRecorder()

		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		err := hm.ReadinessHandler()(ctx)
		if err != nil {
			t.Fatalf("readiness handler returned error: %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response HealthResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Status != StatusDegraded {
			t.Errorf("expected status 'degraded', got '%s'", response.Status)
		}
	})

	t.Run("unhealthy checker", func(t *testing.T) {
		hm := NewHealthManager()
		hm.RegisterChecker(NewStaticHealthChecker("db", StatusHealthy))
		hm.RegisterChecker(NewStaticHealthChecker("redis", StatusUnhealthy))

		req := httptest.NewRequest("GET", "/health/ready", nil)
		w := httptest.NewRecorder()

		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		err := hm.ReadinessHandler()(ctx)
		if err != nil {
			t.Fatalf("readiness handler returned error: %v", err)
		}

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}

		var response HealthResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Status != StatusUnhealthy {
			t.Errorf("expected status 'unhealthy', got '%s'", response.Status)
		}
	})
}

// TestHealthManager_HealthHandler tests the detailed health endpoint
func TestHealthManager_HealthHandler(t *testing.T) {
	hm := NewHealthManager()
	hm.RegisterChecker(NewStaticHealthChecker("db", StatusUnhealthy))

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	err := hm.HealthHandler()(ctx)
	if err != nil {
		t.Fatalf("health handler returned error: %v", err)
	}

	// Health endpoint always returns 200
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != StatusUnhealthy {
		t.Errorf("expected status 'unhealthy', got '%s'", response.Status)
	}
}

// TestStaticHealthChecker tests the static health checker
func TestStaticHealthChecker(t *testing.T) {
	tests := []struct {
		name   string
		status HealthStatus
	}{
		{"healthy", StatusHealthy},
		{"degraded", StatusDegraded},
		{"unhealthy", StatusUnhealthy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewStaticHealthChecker("test", tt.status)

			if checker.Name() != "test" {
				t.Errorf("expected name 'test', got '%s'", checker.Name())
			}

			result := checker.Check(context.Background())
			if result.Status != tt.status {
				t.Errorf("expected status '%s', got '%s'", tt.status, result.Status)
			}
		})
	}
}

// TestDatabaseHealthChecker tests the database health checker
func TestDatabaseHealthChecker(t *testing.T) {
	t.Run("healthy database", func(t *testing.T) {
		pingFn := func(ctx context.Context) error {
			return nil
		}

		checker := NewDatabaseHealthChecker("database", pingFn)
		result := checker.Check(context.Background())

		if result.Status != StatusHealthy {
			t.Errorf("expected status 'healthy', got '%s'", result.Status)
		}

		if result.LatencyMs < 0 {
			t.Errorf("expected non-negative latency, got %d", result.LatencyMs)
		}
	})

	t.Run("unhealthy database", func(t *testing.T) {
		pingFn := func(ctx context.Context) error {
			return errors.New("connection failed")
		}

		checker := NewDatabaseHealthChecker("database", pingFn)
		result := checker.Check(context.Background())

		if result.Status != StatusUnhealthy {
			t.Errorf("expected status 'unhealthy', got '%s'", result.Status)
		}

		if result.Error != "connection failed" {
			t.Errorf("expected error message 'connection failed', got '%s'", result.Error)
		}
	})

	t.Run("degraded database (slow)", func(t *testing.T) {
		pingFn := func(ctx context.Context) error {
			time.Sleep(150 * time.Millisecond)
			return nil
		}

		checker := NewDatabaseHealthChecker("database", pingFn)
		result := checker.Check(context.Background())

		if result.Status != StatusDegraded {
			t.Errorf("expected status 'degraded', got '%s'", result.Status)
		}

		if result.LatencyMs < 100 {
			t.Errorf("expected latency >= 100ms, got %d", result.LatencyMs)
		}

		if result.Message == "" {
			t.Error("expected degraded message to be set")
		}
	})
}

// TestHTTPHealthChecker tests the HTTP health checker
func TestHTTPHealthChecker(t *testing.T) {
	t.Run("healthy service", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		checker := NewHTTPHealthChecker("api", server.URL)
		result := checker.Check(context.Background())

		if result.Status != StatusHealthy {
			t.Errorf("expected status 'healthy', got '%s'", result.Status)
		}
	})

	t.Run("service returning 500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		checker := NewHTTPHealthChecker("api", server.URL)
		result := checker.Check(context.Background())

		if result.Status != StatusUnhealthy {
			t.Errorf("expected status 'unhealthy', got '%s'", result.Status)
		}
	})

	t.Run("service returning 404", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		checker := NewHTTPHealthChecker("api", server.URL)
		result := checker.Check(context.Background())

		if result.Status != StatusDegraded {
			t.Errorf("expected status 'degraded', got '%s'", result.Status)
		}
	})

	t.Run("unreachable service", func(t *testing.T) {
		checker := NewHTTPHealthChecker("api", "http://localhost:99999")
		result := checker.Check(context.Background())

		if result.Status != StatusUnhealthy {
			t.Errorf("expected status 'unhealthy', got '%s'", result.Status)
		}

		if result.Error == "" {
			t.Error("expected error to be set")
		}
	})
}

// TestHealthCheckFunc tests the custom health check function
func TestHealthCheckFunc(t *testing.T) {
	checkFn := func(ctx context.Context) *CheckResult {
		return &CheckResult{
			Status:    StatusHealthy,
			LatencyMs: 5,
			Message:   "custom check passed",
		}
	}

	checker := NewHealthCheckFunc("custom", checkFn)

	if checker.Name() != "custom" {
		t.Errorf("expected name 'custom', got '%s'", checker.Name())
	}

	result := checker.Check(context.Background())

	if result.Status != StatusHealthy {
		t.Errorf("expected status 'healthy', got '%s'", result.Status)
	}

	if result.LatencyMs != 5 {
		t.Errorf("expected latency 5ms, got %d", result.LatencyMs)
	}

	if result.Message != "custom check passed" {
		t.Errorf("expected message 'custom check passed', got '%s'", result.Message)
	}
}

// TestHealthManager_AggregateStatus tests status aggregation
func TestHealthManager_AggregateStatus(t *testing.T) {
	hm := NewHealthManager()

	tests := []struct {
		name     string
		results  map[string]*CheckResult
		expected HealthStatus
	}{
		{
			name:     "no results",
			results:  map[string]*CheckResult{},
			expected: StatusHealthy,
		},
		{
			name: "all healthy",
			results: map[string]*CheckResult{
				"db":    {Status: StatusHealthy},
				"redis": {Status: StatusHealthy},
			},
			expected: StatusHealthy,
		},
		{
			name: "one degraded",
			results: map[string]*CheckResult{
				"db":    {Status: StatusHealthy},
				"redis": {Status: StatusDegraded},
			},
			expected: StatusDegraded,
		},
		{
			name: "one unhealthy",
			results: map[string]*CheckResult{
				"db":    {Status: StatusHealthy},
				"redis": {Status: StatusUnhealthy},
			},
			expected: StatusUnhealthy,
		},
		{
			name: "unhealthy takes precedence",
			results: map[string]*CheckResult{
				"db":    {Status: StatusDegraded},
				"redis": {Status: StatusUnhealthy},
			},
			expected: StatusUnhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := hm.aggregateStatus(tt.results)
			if status != tt.expected {
				t.Errorf("expected status '%s', got '%s'", tt.expected, status)
			}
		})
	}
}

// TestHealthManager_ParallelChecks tests that checks run in parallel
func TestHealthManager_ParallelChecks(t *testing.T) {
	hm := NewHealthManager()

	// Create checkers that sleep for different durations
	slowChecker := NewHealthCheckFunc("slow", func(ctx context.Context) *CheckResult {
		time.Sleep(100 * time.Millisecond)
		return &CheckResult{Status: StatusHealthy}
	})

	fastChecker := NewHealthCheckFunc("fast", func(ctx context.Context) *CheckResult {
		return &CheckResult{Status: StatusHealthy}
	})

	checkers := []HealthChecker{slowChecker, fastChecker}

	start := time.Now()
	results := hm.performChecks(context.Background(), checkers)
	duration := time.Since(start)

	// If checks run in parallel, total time should be ~100ms
	// If sequential, it would be > 100ms
	if duration > 150*time.Millisecond {
		t.Errorf("checks appear to be running sequentially, took %v", duration)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

// TestHealthManager_Timeout tests health check timeout
func TestHealthManager_Timeout(t *testing.T) {
	hm := NewHealthManager(WithHealthCheckTimeout(100 * time.Millisecond))

	// Create a checker that takes longer than the timeout
	slowChecker := NewHealthCheckFunc("slow", func(ctx context.Context) *CheckResult {
		select {
		case <-ctx.Done():
			return &CheckResult{
				Status: StatusUnhealthy,
				Error:  "timeout exceeded",
			}
		case <-time.After(1 * time.Second):
			return &CheckResult{Status: StatusHealthy}
		}
	})

	hm.RegisterChecker(slowChecker)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()

	ctx := &Context{
		Request:        req,
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	start := time.Now()
	hm.ReadinessHandler()(ctx)
	duration := time.Since(start)

	// Should complete within timeout period
	if duration > 200*time.Millisecond {
		t.Errorf("health check took too long: %v", duration)
	}
}

// TestRegisterHealthRoutes tests registering health routes with the server
func TestRegisterHealthRoutes(t *testing.T) {
	server := NewServer()
	hm := NewHealthManager()

	err := server.RegisterHealthRoutes(hm)
	if err != nil {
		t.Fatalf("failed to register health routes: %v", err)
	}

	// Verify routes are registered
	routes := server.GetRouter().GetAllRoutes()
	getRoutes := routes[GET]

	expectedPaths := []string{"/health/live", "/health/ready", "/health"}
	foundPaths := make(map[string]bool)

	for _, route := range getRoutes {
		foundPaths[route.Path] = true
	}

	for _, expectedPath := range expectedPaths {
		if !foundPaths[expectedPath] {
			t.Errorf("expected route %s not found", expectedPath)
		}
	}
}

// TestLivenessHTTPHandler tests the standalone liveness HTTP handler
func TestLivenessHTTPHandler(t *testing.T) {
	handler := LivenessHTTPHandler()

	req := httptest.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != StatusHealthy {
		t.Errorf("expected status 'healthy', got '%s'", response.Status)
	}
}

// TestReadinessHTTPHandler tests the standalone readiness HTTP handler
func TestReadinessHTTPHandler(t *testing.T) {
	hm := NewHealthManager()
	hm.RegisterChecker(NewStaticHealthChecker("db", StatusHealthy))

	handler := ReadinessHTTPHandler(hm)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != StatusHealthy {
		t.Errorf("expected status 'healthy', got '%s'", response.Status)
	}

	if len(response.Checks) != 1 {
		t.Errorf("expected 1 check, got %d", len(response.Checks))
	}
}

// TestHealthResponse_JSON tests JSON serialization
func TestHealthResponse_JSON(t *testing.T) {
	response := &HealthResponse{
		Status: StatusHealthy,
		Checks: map[string]*CheckResult{
			"database": {
				Status:    StatusHealthy,
				LatencyMs: 5,
			},
			"redis": {
				Status:    StatusDegraded,
				LatencyMs: 150,
				Message:   "high latency",
			},
		},
		Timestamp: time.Date(2025, 12, 13, 10, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded HealthResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.Status != response.Status {
		t.Errorf("status mismatch: expected '%s', got '%s'", response.Status, decoded.Status)
	}

	if len(decoded.Checks) != len(response.Checks) {
		t.Errorf("checks count mismatch: expected %d, got %d", len(response.Checks), len(decoded.Checks))
	}
}

// Benchmark tests
func BenchmarkHealthManager_LivenessHandler(b *testing.B) {
	hm := NewHealthManager()

	req := httptest.NewRequest("GET", "/health/live", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		hm.LivenessHandler()(ctx)
	}
}

func BenchmarkHealthManager_ReadinessHandler(b *testing.B) {
	hm := NewHealthManager()
	hm.RegisterChecker(NewStaticHealthChecker("db", StatusHealthy))
	hm.RegisterChecker(NewStaticHealthChecker("redis", StatusHealthy))

	req := httptest.NewRequest("GET", "/health/ready", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ctx := &Context{
			Request:        req,
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		hm.ReadinessHandler()(ctx)
	}
}
