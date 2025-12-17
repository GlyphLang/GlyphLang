# Health Check System - Quick Reference Guide

## Overview

The GLYPHLANG health check system provides production-ready health monitoring with support for liveness probes, readiness probes, and custom health checkers.

## Quick Start

```go
package main

import (
    "github.com/glyphlang/glyph/pkg/server"
    "time"
)

func main() {
    // 1. Create a health manager
    hm := server.NewHealthManager(
        server.WithHealthCheckTimeout(5 * time.Second),
    )

    // 2. Register health checkers
    hm.RegisterChecker(server.NewStaticHealthChecker("db", server.StatusHealthy))

    // 3. Create server and register routes
    srv := server.NewServer()
    srv.RegisterHealthRoutes(hm)

    // 4. Start server
    srv.Start(":8080")
}
```

## Endpoints

| Endpoint | Purpose | Use Case |
|----------|---------|----------|
| `/health/live` | Liveness probe | Kubernetes liveness checks |
| `/health/ready` | Readiness probe | Kubernetes readiness checks, load balancers |
| `/health` | Detailed health | Monitoring dashboards, debugging |

## Health Status Values

- **`healthy`**: Component is functioning properly
- **`degraded`**: Component is functional but has issues (e.g., high latency)
- **`unhealthy`**: Component is not functioning

## Built-in Health Checkers

### Static Checker
Always returns a fixed status (useful for testing)

```go
checker := server.NewStaticHealthChecker("service", server.StatusHealthy)
```

### Database Checker
Checks database connectivity with latency tracking

```go
checker := server.NewDatabaseHealthChecker("database", func(ctx context.Context) error {
    return db.PingContext(ctx)
})
```

- ✓ Healthy: < 100ms latency
- ⚠ Degraded: >= 100ms latency
- ✗ Unhealthy: Connection failed

### HTTP Checker
Checks external HTTP service availability

```go
checker := server.NewHTTPHealthChecker("api", "https://api.example.com/health")
```

- ✓ Healthy: 2xx/3xx response
- ⚠ Degraded: 4xx response
- ✗ Unhealthy: 5xx response or connection error

### Custom Checker
Create your own health check logic

```go
checker := server.NewHealthCheckFunc("custom", func(ctx context.Context) *server.CheckResult {
    // Your check logic here
    return &server.CheckResult{
        Status:    server.StatusHealthy,
        LatencyMs: 10,
        Message:   "All systems operational",
    }
})
```

## API Reference

### HealthManager

```go
// Create a new health manager
hm := server.NewHealthManager(
    server.WithHealthCheckTimeout(5 * time.Second),
)

// Register a health checker
hm.RegisterChecker(checker)

// Unregister a health checker
hm.UnregisterChecker("checker-name")
```

### CheckResult

```go
type CheckResult struct {
    Status    HealthStatus  // healthy, degraded, or unhealthy
    LatencyMs int64         // Response time in milliseconds
    Message   string        // Optional message
    Error     string        // Error message if check failed
}
```

### HealthResponse

```go
type HealthResponse struct {
    Status    HealthStatus               // Overall aggregated status
    Checks    map[string]*CheckResult    // Individual check results
    Timestamp time.Time                  // When the check was performed
}
```

## Response Examples

### Liveness Response
```json
{
  "status": "healthy",
  "timestamp": "2025-12-13T10:00:00Z"
}
```

### Readiness Response (All Healthy)
```json
{
  "status": "healthy",
  "checks": {
    "database": {
      "status": "healthy",
      "latency_ms": 5
    },
    "cache": {
      "status": "healthy",
      "latency_ms": 2
    }
  },
  "timestamp": "2025-12-13T10:00:00Z"
}
```

### Readiness Response (Degraded)
```json
{
  "status": "degraded",
  "checks": {
    "database": {
      "status": "degraded",
      "latency_ms": 150,
      "message": "high latency detected"
    },
    "cache": {
      "status": "healthy",
      "latency_ms": 2
    }
  },
  "timestamp": "2025-12-13T10:00:00Z"
}
```

### Readiness Response (Unhealthy)
```json
{
  "status": "unhealthy",
  "checks": {
    "database": {
      "status": "unhealthy",
      "latency_ms": 5000,
      "error": "connection refused"
    },
    "cache": {
      "status": "healthy",
      "latency_ms": 2
    }
  },
  "timestamp": "2025-12-13T10:00:00Z"
}
```

## HTTP Status Codes

| Endpoint | Healthy | Degraded | Unhealthy |
|----------|---------|----------|-----------|
| `/health/live` | 200 | 200 | 200 |
| `/health/ready` | 200 | 200 | 503 |
| `/health` | 200 | 200 | 200 |

## Kubernetes Integration

### Liveness Probe
```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### Readiness Probe
```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 5
  failureThreshold: 3
```

## Best Practices

### 1. Liveness Checks
- ✓ Keep simple and lightweight
- ✓ Should rarely fail (only if app is truly dead)
- ✗ Don't check external dependencies
- ✗ Don't perform expensive operations

### 2. Readiness Checks
- ✓ Check critical dependencies (database, cache, etc.)
- ✓ Should fail if the app can't serve requests
- ✓ Can be more expensive than liveness
- ✗ Don't make it so strict that it always fails

### 3. Timeout Configuration
- ✓ Set appropriate timeouts for your dependencies
- ✓ Consider network latency in distributed systems
- ✓ Default is 5 seconds, adjust as needed

### 4. Health Check Frequency
- ✓ Balance between quick detection and resource usage
- ✓ Typical: liveness every 10s, readiness every 5s
- ✗ Don't check too frequently (can overload dependencies)

### 5. Degraded Status
- ✓ Use for warning conditions (slow but working)
- ✓ Helps identify issues before complete failure
- ✓ Can trigger alerts without removing from load balancer

## Advanced Usage

### Standalone HTTP Handlers
Use health checks without the GLYPHLANG routing system:

```go
import "net/http"

hm := server.NewHealthManager()
// ... register checkers ...

// Use with standard http.ServeMux
mux := http.NewServeMux()
mux.HandleFunc("/health/live", server.LivenessHTTPHandler())
mux.HandleFunc("/health/ready", server.ReadinessHTTPHandler(hm))

http.ListenAndServe(":8080", mux)
```

### Dynamic Health Checker Registration
Add/remove health checkers at runtime:

```go
// Add when service becomes available
hm.RegisterChecker(newServiceChecker)

// Remove when service is disabled
hm.UnregisterChecker("service-name")
```

### Custom Aggregation Logic
The default aggregation rules:
- Any unhealthy → overall unhealthy
- Any degraded → overall degraded
- All healthy → overall healthy

For custom logic, implement your own handler using the HealthManager's methods.

## Testing

### Unit Testing Your Health Checkers

```go
func TestMyHealthChecker(t *testing.T) {
    checker := NewMyHealthChecker()
    result := checker.Check(context.Background())

    if result.Status != server.StatusHealthy {
        t.Errorf("expected healthy, got %s", result.Status)
    }
}
```

### Integration Testing

```go
func TestHealthEndpoints(t *testing.T) {
    hm := server.NewHealthManager()
    hm.RegisterChecker(server.NewStaticHealthChecker("test", server.StatusHealthy))

    srv := server.NewServer()
    srv.RegisterHealthRoutes(hm)

    // Test endpoints with httptest
}
```

## Performance

Benchmark results (AMD Ryzen 7 7800X3D):
- Liveness handler: ~700 ns/op
- Readiness handler (2 checkers): ~3,400 ns/op

Health checks run in parallel, so total time is approximately the slowest checker's time, not the sum of all checkers.

## Troubleshooting

### Health checks timing out
- Increase timeout: `server.WithHealthCheckTimeout(10 * time.Second)`
- Check for slow dependencies
- Ensure checkers respect context cancellation

### Readiness always failing
- Check if dependencies are actually available
- Verify timeout is sufficient
- Review individual check results in `/health` endpoint

### Too many false positives
- Adjust latency thresholds in checkers
- Use degraded status instead of unhealthy
- Increase check interval in Kubernetes probes

## Example

See the complete working example in `examples/health-check/`

```bash
cd examples/health-check
go run main.go
```

## API Documentation

Full API documentation is available in the Go package documentation:

```bash
go doc github.com/glyphlang/glyph/pkg/server.HealthManager
go doc github.com/glyphlang/glyph/pkg/server.HealthChecker
```
