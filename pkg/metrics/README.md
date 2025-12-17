# Metrics Package

The metrics package provides Prometheus-based metrics collection for the GLYPHLANG project. It includes automatic HTTP metrics collection, runtime metrics, and support for custom business metrics.

## Features

- **Request Metrics**: Automatic collection of HTTP request rates, latencies, and errors
- **Runtime Metrics**: Monitoring of goroutines, memory usage, and garbage collection
- **Custom Metrics**: Support for custom counters, gauges, and histograms
- **Middleware Integration**: Easy integration with existing HTTP middleware
- **Prometheus Compatible**: Exposes metrics in Prometheus format via `/metrics` endpoint

## Installation

The package is part of the GLYPHLANG project and uses the official Prometheus client library:

```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

## Quick Start

### Basic Usage

```go
package main

import (
    "net/http"
    "time"

    "github.com/glyphlang/glyph/pkg/metrics"
)

func main() {
    // Create metrics with default configuration
    m := metrics.NewMetrics(metrics.DefaultConfig())

    // Expose metrics endpoint
    http.Handle("/metrics", m.Handler())

    // Record requests manually
    m.RecordRequest("GET", "/api/users", 200, 50*time.Millisecond)

    http.ListenAndServe(":8080", nil)
}
```

### Middleware Integration

```go
import (
    "github.com/glyphlang/glyph/pkg/metrics"
    "github.com/glyphlang/glyph/pkg/server"
)

// Create metrics instance
m := metrics.NewMetrics(metrics.DefaultConfig())

// Create metrics middleware
metricsMiddleware := metrics.MetricsMiddleware(m)

// Apply to your routes
handler := metricsMiddleware(yourHandler)
```

### Custom Configuration

```go
config := metrics.Config{
    Namespace: "myapp",
    Subsystem: "api",
    DurationBuckets: []float64{
        0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0,
    },
}

m := metrics.NewMetrics(config)
```

## Metrics Collected

### HTTP Metrics

#### `glyphlang_http_requests_total`
Counter tracking total number of HTTP requests.

**Labels:**
- `method`: HTTP method (GET, POST, etc.)
- `path`: Request path
- `status`: HTTP status code

#### `glyphlang_http_request_duration_seconds`
Histogram tracking HTTP request latency in seconds.

**Labels:**
- `method`: HTTP method
- `path`: Request path
- `status`: HTTP status code

**Default Buckets:** 1ms, 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s

#### `glyphlang_http_request_errors_total`
Counter tracking HTTP errors (status codes >= 400).

**Labels:**
- `method`: HTTP method
- `path`: Request path
- `status`: HTTP status code

### Runtime Metrics

#### `glyphlang_runtime_goroutines`
Gauge showing the current number of goroutines.

#### `glyphlang_runtime_memory_alloc_bytes`
Gauge showing bytes currently allocated and in use.

#### `glyphlang_runtime_memory_total_alloc_bytes`
Gauge showing cumulative bytes allocated.

#### `glyphlang_runtime_memory_sys_bytes`
Gauge showing bytes obtained from the system.

#### `glyphlang_runtime_gc_pause_ns`
Gauge showing the most recent GC pause time in nanoseconds.

#### `glyphlang_runtime_gc_runs_total`
Gauge showing the total number of GC runs.

## Custom Metrics

### Custom Counters

Counters are monotonically increasing values, useful for counting events.

```go
// Register a counter
err := m.RegisterCustomCounter(
    "user_signups_total",
    "Total number of user signups",
    []string{"plan", "region"},
)

// Increment the counter
m.IncrementCustomCounter("user_signups_total", map[string]string{
    "plan":   "premium",
    "region": "us-east",
})
```

### Custom Gauges

Gauges represent values that can go up or down.

```go
// Register a gauge
err := m.RegisterCustomGauge(
    "queue_size",
    "Current size of processing queue",
    []string{"queue_name"},
)

// Set the gauge value
m.SetCustomGauge("queue_size", 42.0, map[string]string{
    "queue_name": "email",
})
```

### Custom Histograms

Histograms track distributions of values, useful for measuring durations.

```go
// Register a histogram
buckets := []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0}
err := m.RegisterCustomHistogram(
    "db_query_duration_seconds",
    "Database query execution time in seconds",
    []string{"query_type"},
    buckets,
)

// Observe values
m.ObserveCustomHistogram("db_query_duration_seconds", 0.023, map[string]string{
    "query_type": "select",
})
```

## API Reference

### NewMetrics

```go
func NewMetrics(config Config) *Metrics
```

Creates a new metrics instance with the specified configuration.

### Config

```go
type Config struct {
    Namespace       string    // Metric namespace prefix (default: "glyphlang")
    Subsystem       string    // Metric subsystem prefix (default: "http")
    DurationBuckets []float64 // Histogram buckets for request duration
}
```

### DefaultConfig

```go
func DefaultConfig() Config
```

Returns the default metrics configuration.

### RecordRequest

```go
func (m *Metrics) RecordRequest(method, path string, statusCode int, duration time.Duration)
```

Records metrics for an HTTP request.

### UpdateRuntimeMetrics

```go
func (m *Metrics) UpdateRuntimeMetrics()
```

Updates runtime metrics (called automatically every 15 seconds).

### RegisterCustomCounter

```go
func (m *Metrics) RegisterCustomCounter(name, help string, labels []string) error
```

Registers a new custom counter metric.

### RegisterCustomGauge

```go
func (m *Metrics) RegisterCustomGauge(name, help string, labels []string) error
```

Registers a new custom gauge metric.

### RegisterCustomHistogram

```go
func (m *Metrics) RegisterCustomHistogram(name, help string, labels []string, buckets []float64) error
```

Registers a new custom histogram metric.

### IncrementCustomCounter

```go
func (m *Metrics) IncrementCustomCounter(name string, labels map[string]string)
```

Increments a custom counter.

### SetCustomGauge

```go
func (m *Metrics) SetCustomGauge(name string, value float64, labels map[string]string)
```

Sets a custom gauge value.

### ObserveCustomHistogram

```go
func (m *Metrics) ObserveCustomHistogram(name string, value float64, labels map[string]string)
```

Observes a value in a custom histogram.

### Handler

```go
func (m *Metrics) Handler() http.Handler
```

Returns an HTTP handler for the `/metrics` endpoint.

### GetRegistry

```go
func (m *Metrics) GetRegistry() *prometheus.Registry
```

Returns the underlying Prometheus registry.

## Middleware

### MetricsMiddleware

```go
func MetricsMiddleware(m *Metrics) server.Middleware
```

Creates middleware that automatically collects HTTP metrics for all requests.

## Integration Example

Complete example integrating metrics with an GLYPHLANG server:

```go
package main

import (
    "net/http"
    "time"

    "github.com/glyphlang/glyph/pkg/metrics"
    "github.com/glyphlang/glyph/pkg/server"
)

func main() {
    // Create metrics
    m := metrics.NewMetrics(metrics.DefaultConfig())

    // Create server and apply metrics middleware
    metricsMiddleware := metrics.MetricsMiddleware(m)

    // Register custom business metrics
    m.RegisterCustomCounter(
        "api_calls_total",
        "Total API calls by endpoint",
        []string{"endpoint", "user_type"},
    )

    // Your route handler
    handler := func(ctx *server.Context) error {
        // Track business metric
        m.IncrementCustomCounter("api_calls_total", map[string]string{
            "endpoint":  "/api/data",
            "user_type": "premium",
        })

        ctx.StatusCode = 200
        return nil
    }

    // Apply middleware
    wrappedHandler := metricsMiddleware(handler)

    // Setup HTTP server
    http.Handle("/api/data", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := &server.Context{
            Request:        r,
            ResponseWriter: w,
            PathParams:     make(map[string]string),
            QueryParams:    make(map[string]string),
            Body:           make(map[string]interface{}),
        }
        wrappedHandler(ctx)
    }))

    // Expose metrics endpoint
    http.Handle("/metrics", m.Handler())

    http.ListenAndServe(":8080", nil)
}
```

## Prometheus Configuration

To scrape metrics with Prometheus, add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'glyphlang'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

## Best Practices

1. **Use Meaningful Labels**: Keep label cardinality low to avoid memory issues
2. **Choose Appropriate Metric Types**:
   - Use counters for monotonically increasing values
   - Use gauges for values that can go up or down
   - Use histograms for distributions and durations
3. **Set Appropriate Buckets**: Configure histogram buckets based on your expected value distribution
4. **Don't Over-Instrument**: Only track metrics that provide actionable insights
5. **Use Consistent Naming**: Follow Prometheus naming conventions (e.g., `_total` suffix for counters)

## Performance

- Metrics collection adds minimal overhead (< 1ms per request)
- Runtime metrics are updated every 15 seconds in a background goroutine
- Thread-safe and suitable for high-concurrency environments
- Benchmark results available in `metrics_test.go`

## License

Part of the GLYPHLANG project.
