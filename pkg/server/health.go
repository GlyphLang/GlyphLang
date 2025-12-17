package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	// StatusHealthy indicates the component is functioning properly
	StatusHealthy HealthStatus = "healthy"
	// StatusDegraded indicates the component is functioning but with issues
	StatusDegraded HealthStatus = "degraded"
	// StatusUnhealthy indicates the component is not functioning
	StatusUnhealthy HealthStatus = "unhealthy"
)

// CheckResult represents the result of a single health check
type CheckResult struct {
	Status    HealthStatus `json:"status"`
	LatencyMs int64        `json:"latency_ms,omitempty"`
	Message   string       `json:"message,omitempty"`
	Error     string       `json:"error,omitempty"`
}

// HealthResponse represents the aggregated health check response
type HealthResponse struct {
	Status    HealthStatus            `json:"status"`
	Checks    map[string]*CheckResult `json:"checks,omitempty"`
	Timestamp time.Time               `json:"timestamp"`
}

// HealthChecker is an interface for components that can report their health
type HealthChecker interface {
	// Check performs a health check and returns the result
	// The context may contain a timeout
	Check(ctx context.Context) *CheckResult
	// Name returns the name of this health checker
	Name() string
}

// HealthCheckFunc is a function type that implements HealthChecker
type HealthCheckFunc struct {
	name     string
	checkFn  func(ctx context.Context) *CheckResult
}

// Check implements HealthChecker
func (f *HealthCheckFunc) Check(ctx context.Context) *CheckResult {
	return f.checkFn(ctx)
}

// Name implements HealthChecker
func (f *HealthCheckFunc) Name() string {
	return f.name
}

// NewHealthCheckFunc creates a new HealthCheckFunc
func NewHealthCheckFunc(name string, fn func(ctx context.Context) *CheckResult) *HealthCheckFunc {
	return &HealthCheckFunc{
		name:    name,
		checkFn: fn,
	}
}

// HealthManager manages health checks and provides endpoints
type HealthManager struct {
	checkers map[string]HealthChecker
	mu       sync.RWMutex
	timeout  time.Duration
}

// HealthManagerOption is a functional option for configuring the health manager
type HealthManagerOption func(*HealthManager)

// WithHealthCheckTimeout sets the timeout for health checks
func WithHealthCheckTimeout(timeout time.Duration) HealthManagerOption {
	return func(hm *HealthManager) {
		hm.timeout = timeout
	}
}

// NewHealthManager creates a new health manager
func NewHealthManager(options ...HealthManagerOption) *HealthManager {
	hm := &HealthManager{
		checkers: make(map[string]HealthChecker),
		timeout:  5 * time.Second, // Default 5 second timeout
	}

	for _, opt := range options {
		opt(hm)
	}

	return hm
}

// RegisterChecker registers a new health checker
func (hm *HealthManager) RegisterChecker(checker HealthChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.checkers[checker.Name()] = checker
}

// UnregisterChecker removes a health checker
func (hm *HealthManager) UnregisterChecker(name string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.checkers, name)
}

// LivenessHandler handles liveness probe requests
// Liveness probes check if the application is running
// Returns 200 if the application is alive
func (hm *HealthManager) LivenessHandler() RouteHandler {
	return func(ctx *Context) error {
		response := &HealthResponse{
			Status:    StatusHealthy,
			Timestamp: time.Now().UTC(),
		}

		return SendJSON(ctx, http.StatusOK, response)
	}
}

// ReadinessHandler handles readiness probe requests
// Readiness probes check if the application is ready to serve traffic
// Returns 200 if ready, 503 if not ready
func (hm *HealthManager) ReadinessHandler() RouteHandler {
	return func(ctx *Context) error {
		// Create context with timeout
		checkCtx, cancel := context.WithTimeout(context.Background(), hm.timeout)
		defer cancel()

		// Get all checkers
		hm.mu.RLock()
		checkers := make([]HealthChecker, 0, len(hm.checkers))
		for _, checker := range hm.checkers {
			checkers = append(checkers, checker)
		}
		hm.mu.RUnlock()

		// Perform all checks in parallel
		results := hm.performChecks(checkCtx, checkers)

		// Aggregate status
		aggregatedStatus := hm.aggregateStatus(results)

		response := &HealthResponse{
			Status:    aggregatedStatus,
			Checks:    results,
			Timestamp: time.Now().UTC(),
		}

		// Return 503 if not healthy
		statusCode := http.StatusOK
		if aggregatedStatus == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		return SendJSON(ctx, statusCode, response)
	}
}

// HealthHandler handles detailed health check requests
// Similar to readiness but always returns 200 with detailed status
func (hm *HealthManager) HealthHandler() RouteHandler {
	return func(ctx *Context) error {
		// Create context with timeout
		checkCtx, cancel := context.WithTimeout(context.Background(), hm.timeout)
		defer cancel()

		// Get all checkers
		hm.mu.RLock()
		checkers := make([]HealthChecker, 0, len(hm.checkers))
		for _, checker := range hm.checkers {
			checkers = append(checkers, checker)
		}
		hm.mu.RUnlock()

		// Perform all checks in parallel
		results := hm.performChecks(checkCtx, checkers)

		// Aggregate status
		aggregatedStatus := hm.aggregateStatus(results)

		response := &HealthResponse{
			Status:    aggregatedStatus,
			Checks:    results,
			Timestamp: time.Now().UTC(),
		}

		return SendJSON(ctx, http.StatusOK, response)
	}
}

// performChecks executes all health checks in parallel
func (hm *HealthManager) performChecks(ctx context.Context, checkers []HealthChecker) map[string]*CheckResult {
	results := make(map[string]*CheckResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, checker := range checkers {
		wg.Add(1)
		go func(c HealthChecker) {
			defer wg.Done()

			start := time.Now()
			result := c.Check(ctx)

			// Calculate latency if not already set
			if result.LatencyMs == 0 {
				result.LatencyMs = time.Since(start).Milliseconds()
			}

			mu.Lock()
			results[c.Name()] = result
			mu.Unlock()
		}(checker)
	}

	wg.Wait()
	return results
}

// aggregateStatus determines overall health from individual check results
func (hm *HealthManager) aggregateStatus(results map[string]*CheckResult) HealthStatus {
	if len(results) == 0 {
		return StatusHealthy
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, result := range results {
		switch result.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}

// RegisterHealthRoutes registers all health check routes with the server
func (s *Server) RegisterHealthRoutes(hm *HealthManager) error {
	routes := []*Route{
		{
			Method:  GET,
			Path:    "/health/live",
			Handler: hm.LivenessHandler(),
		},
		{
			Method:  GET,
			Path:    "/health/ready",
			Handler: hm.ReadinessHandler(),
		},
		{
			Method:  GET,
			Path:    "/health",
			Handler: hm.HealthHandler(),
		},
	}

	return s.RegisterRoutes(routes)
}

// Common health checker implementations

// DatabaseHealthChecker checks database connectivity
type DatabaseHealthChecker struct {
	name    string
	pingFn  func(ctx context.Context) error
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(name string, pingFn func(ctx context.Context) error) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		name:   name,
		pingFn: pingFn,
	}
}

// Name returns the checker name
func (d *DatabaseHealthChecker) Name() string {
	return d.name
}

// Check performs the database health check
func (d *DatabaseHealthChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()

	err := d.pingFn(ctx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &CheckResult{
			Status:    StatusUnhealthy,
			LatencyMs: latency,
			Error:     err.Error(),
		}
	}

	// Check for degraded performance (> 100ms is considered slow)
	status := StatusHealthy
	message := ""
	if latency > 100 {
		status = StatusDegraded
		message = "high latency detected"
	}

	return &CheckResult{
		Status:    status,
		LatencyMs: latency,
		Message:   message,
	}
}

// HTTPHealthChecker checks external HTTP service availability
type HTTPHealthChecker struct {
	name   string
	url    string
	client *http.Client
}

// NewHTTPHealthChecker creates a new HTTP service health checker
func NewHTTPHealthChecker(name, url string) *HTTPHealthChecker {
	return &HTTPHealthChecker{
		name: name,
		url:  url,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

// Name returns the checker name
func (h *HTTPHealthChecker) Name() string {
	return h.name
}

// Check performs the HTTP health check
func (h *HTTPHealthChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "GET", h.url, nil)
	if err != nil {
		return &CheckResult{
			Status: StatusUnhealthy,
			Error:  err.Error(),
		}
	}

	resp, err := h.client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &CheckResult{
			Status:    StatusUnhealthy,
			LatencyMs: latency,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	// Check response status
	status := StatusHealthy
	message := ""
	if resp.StatusCode >= 500 {
		status = StatusUnhealthy
		message = "service returned 5xx error"
	} else if resp.StatusCode >= 400 {
		status = StatusDegraded
		message = "service returned 4xx error"
	}

	return &CheckResult{
		Status:    status,
		LatencyMs: latency,
		Message:   message,
	}
}

// StaticHealthChecker always returns a fixed status (useful for testing)
type StaticHealthChecker struct {
	name   string
	status HealthStatus
}

// NewStaticHealthChecker creates a health checker that always returns the same status
func NewStaticHealthChecker(name string, status HealthStatus) *StaticHealthChecker {
	return &StaticHealthChecker{
		name:   name,
		status: status,
	}
}

// Name returns the checker name
func (s *StaticHealthChecker) Name() string {
	return s.name
}

// Check returns the static status
func (s *StaticHealthChecker) Check(ctx context.Context) *CheckResult {
	return &CheckResult{
		Status: s.status,
	}
}

// Helper function to create HTTP handlers for standalone use
// (without using the GLYPHLANG routing system)

// LivenessHTTPHandler creates a standard http.Handler for liveness checks
func LivenessHTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := &HealthResponse{
			Status:    StatusHealthy,
			Timestamp: time.Now().UTC(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// ReadinessHTTPHandler creates a standard http.Handler for readiness checks
func ReadinessHTTPHandler(hm *HealthManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), hm.timeout)
		defer cancel()

		// Get all checkers
		hm.mu.RLock()
		checkers := make([]HealthChecker, 0, len(hm.checkers))
		for _, checker := range hm.checkers {
			checkers = append(checkers, checker)
		}
		hm.mu.RUnlock()

		// Perform all checks
		results := hm.performChecks(ctx, checkers)
		aggregatedStatus := hm.aggregateStatus(results)

		response := &HealthResponse{
			Status:    aggregatedStatus,
			Checks:    results,
			Timestamp: time.Now().UTC(),
		}

		statusCode := http.StatusOK
		if aggregatedStatus == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	}
}
