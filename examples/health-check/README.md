# Health Check Demo

This example demonstrates the comprehensive health check system for GLYPHLANG applications.

## Features

- **Liveness Probe**: Basic "am I running" check (`/health/live`)
- **Readiness Probe**: "Am I ready to serve traffic" check (`/health/ready`)
- **Detailed Health Status**: Component-level health information (`/health`)
- **Multiple Health Checkers**: Database, cache, and application health
- **Configurable Timeouts**: Prevent slow checks from blocking
- **Parallel Execution**: All health checks run concurrently
- **Aggregated Status**: Automatic rollup of component health

## Quick Start

### Build and Run

```bash
go build -o health-demo
./health-demo
```

### Available Endpoints

#### Health Check Endpoints

```bash
# Liveness probe - Always returns 200 if the app is running
curl http://localhost:8080/health/live

# Readiness probe - Returns 200 if ready, 503 if not ready
curl http://localhost:8080/health/ready

# Detailed health status - Always returns 200 with detailed info
curl http://localhost:8080/health
```

#### Sample API Endpoint

```bash
# Test the application is working
curl http://localhost:8080/api/hello
```

#### Simulation Endpoints (Demo Only)

```bash
# Simulate database failure
curl -X POST http://localhost:8080/admin/simulate/database/down

# Restore database
curl -X POST http://localhost:8080/admin/simulate/database/up

# Simulate slow database (degraded state)
curl -X POST http://localhost:8080/admin/simulate/database/slow

# Restore database speed
curl -X POST http://localhost:8080/admin/simulate/database/fast

# Simulate cache failure
curl -X POST http://localhost:8080/admin/simulate/cache/down

# Restore cache
curl -X POST http://localhost:8080/admin/simulate/cache/up
```

## Response Format

### Liveness Response

```json
{
  "status": "healthy",
  "timestamp": "2025-12-13T10:00:00Z"
}
```

### Readiness/Health Response

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
    },
    "application": {
      "status": "healthy"
    }
  },
  "timestamp": "2025-12-13T10:00:00Z"
}
```

### Status Values

- **healthy**: Component is functioning properly
- **degraded**: Component is functioning but with issues (e.g., high latency)
- **unhealthy**: Component is not functioning

## Testing Scenarios

### 1. Healthy System

```bash
curl http://localhost:8080/health/ready
# Returns: 200 OK with all components healthy
```

### 2. Degraded System

```bash
# Simulate slow database
curl -X POST http://localhost:8080/admin/simulate/database/slow

# Check health
curl http://localhost:8080/health/ready
# Returns: 200 OK with database marked as degraded (latency > 100ms)
```

### 3. Unhealthy System

```bash
# Simulate database failure
curl -X POST http://localhost:8080/admin/simulate/database/down

# Check readiness
curl http://localhost:8080/health/ready
# Returns: 503 Service Unavailable

# Liveness still works
curl http://localhost:8080/health/live
# Returns: 200 OK (app is still running)
```

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

## Implementation Details

### Health Check Components

1. **Database Checker**: Pings the database and measures latency
   - Healthy: < 100ms latency
   - Degraded: >= 100ms latency
   - Unhealthy: Connection failed

2. **Cache Checker**: Verifies cache connectivity
   - Healthy: Connected
   - Unhealthy: Not connected

3. **Application Checker**: Monitors application-level metrics
   - Can check memory usage, goroutines, etc.
   - Customizable based on application needs

### Health Check Execution

- All checks run in parallel for performance
- Configurable timeout (default: 5 seconds)
- Results are aggregated:
  - Any unhealthy → overall unhealthy
  - Any degraded → overall degraded
  - All healthy → overall healthy

## Customization

### Adding Custom Health Checkers

```go
// Create a custom health checker
customChecker := server.NewHealthCheckFunc("my-service", func(ctx context.Context) *server.CheckResult {
    // Your health check logic here
    start := time.Now()

    // Check your service
    err := myService.Check(ctx)

    latency := time.Since(start).Milliseconds()

    if err != nil {
        return &server.CheckResult{
            Status:    server.StatusUnhealthy,
            LatencyMs: latency,
            Error:     err.Error(),
        }
    }

    return &server.CheckResult{
        Status:    server.StatusHealthy,
        LatencyMs: latency,
    }
})

// Register the checker
healthManager.RegisterChecker(customChecker)
```

### Configuring Timeout

```go
healthManager := server.NewHealthManager(
    server.WithHealthCheckTimeout(10 * time.Second),
)
```

## Production Considerations

1. **Liveness vs Readiness**:
   - Liveness should be simple and never fail unless the app is truly dead
   - Readiness should check dependencies and be more strict

2. **Timeout Configuration**:
   - Set timeouts appropriate for your dependencies
   - Consider network latency in distributed systems

3. **Degraded State**:
   - Use degraded status to warn about performance issues
   - Allows traffic to continue while alerting operators

4. **Health Check Frequency**:
   - Balance between quick detection and resource usage
   - Consider caching health check results if needed

## Dependencies

- GLYPHLANG Server (pkg/server)
- Go 1.21+

## License

Part of the GLYPHLANG project
