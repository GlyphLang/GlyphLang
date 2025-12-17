# OpenTelemetry Distributed Tracing for GLYPHLANG

This package provides production-ready OpenTelemetry distributed tracing functionality for the GLYPHLANG project. It supports W3C Trace Context propagation, automatic HTTP request tracing, and flexible configuration for both development and production environments.

## Features

- **W3C Trace Context Propagation**: Full support for W3C Trace Context standard
- **Span Creation & Management**: Easy-to-use APIs for creating and managing spans
- **Span Annotations**: Rich attributes, events, and error recording
- **HTTP Request Tracing**: Automatic tracing for both incoming and outgoing HTTP requests
- **Configurable Exporters**:
  - `stdout` exporter for development/debugging
  - `OTLP` exporter for production (works with Jaeger, Zipkin, etc.)
- **TraceID/SpanID Extraction**: Easy access to trace and span IDs for correlation
- **Middleware Integration**: Ready-to-use middleware for GLYPHLANG server
- **Context Propagation**: Automatic trace context injection and extraction
- **Production Ready**: Includes sampling, batching, and resource attribution

## Installation

This package requires OpenTelemetry Go libraries. Update your `go.mod`:

```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/stdout/stdouttrace
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
go get go.opentelemetry.io/otel/sdk
```

## Quick Start

### Basic Initialization

```go
package main

import (
    "context"
    "log"
    "github.com/glyphlang/glyph/pkg/tracing"
)

func main() {
    // Initialize tracing with default config (stdout exporter)
    tp, err := tracing.InitTracing(tracing.DefaultConfig())
    if err != nil {
        log.Fatal(err)
    }
    defer tp.Shutdown(context.Background())

    // Your application code here
}
```

### Production Configuration

```go
config := &tracing.Config{
    ServiceName:    "glyphlang-api",
    ServiceVersion: "1.0.0",
    Environment:    "production",
    ExporterType:   "otlp",
    OTLPEndpoint:   "jaeger:4317", // or your collector endpoint
    SamplingRate:   0.1,            // Sample 10% of traces
    Enabled:        true,
}

tp, err := tracing.InitTracing(config)
if err != nil {
    log.Fatal(err)
}
defer tp.Shutdown(context.Background())
```

## Usage Examples

### Creating Spans

```go
ctx := context.Background()

// Start a span
ctx, span := tracing.StartSpan(ctx, "process-order", tracing.SpanKind.Internal)
defer span.End()

// Add attributes
tracing.SetAttributes(ctx,
    attribute.String("order.id", "12345"),
    attribute.Int("order.items", 3),
    attribute.Float64("order.total", 99.99),
)

// Add events
tracing.AddEvent(ctx, "order-validated")
tracing.AddEvent(ctx, "payment-processed")
```

### Error Handling

```go
ctx, span := tracing.StartSpan(ctx, "risky-operation", tracing.SpanKind.Internal)
defer span.End()

if err := someOperation(); err != nil {
    // Record error with additional context
    tracing.RecordError(ctx, err,
        attribute.String("error.type", "validation"),
        attribute.String("operation", "someOperation"),
    )
    return err
}

// Or use SetError for simpler cases
tracing.SetError(ctx, err)
```

### Using WithSpan Helper

```go
err := tracing.WithSpan(ctx, "database-query", func(ctx context.Context) error {
    // Your code here
    result, err := db.Query(ctx, "SELECT * FROM users")
    if err != nil {
        return err
    }
    // Process result...
    return nil
}, tracing.SpanKind.Internal)
```

### HTTP Server Middleware

```go
import (
    "github.com/glyphlang/glyph/pkg/tracing"
    "net/http"
)

// Initialize tracing
tp, _ := tracing.InitTracing(tracing.DefaultConfig())
defer tp.Shutdown(context.Background())

// Create middleware configuration
config := &tracing.MiddlewareConfig{
    SpanNameFormatter: func(req *http.Request) string {
        return fmt.Sprintf("HTTP %s %s", req.Method, req.URL.Path)
    },
    ExcludePaths: map[string]bool{
        "/health":  true,
        "/metrics": true,
    },
    CustomAttributes: func(req *http.Request) []attribute.KeyValue {
        return []attribute.KeyValue{
            attribute.String("api.version", "v1"),
        }
    },
}

// Create middleware
middleware := tracing.HTTPTracingMiddleware(config)

// Apply to handler
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
})

tracedHandler := middleware(handler)
```

### GLYPHLANG Server Integration

```go
import (
    "github.com/glyphlang/glyph/pkg/server"
    "github.com/glyphlang/glyph/pkg/tracing"
)

// Initialize tracing
tp, err := tracing.InitTracing(tracing.DefaultConfig())
if err != nil {
    log.Fatal(err)
}
defer tp.Shutdown(context.Background())

// Create server with tracing middleware
s := server.NewServer(
    server.WithMiddleware(server.TracingMiddleware(tracing.DefaultMiddlewareConfig())),
    server.WithMiddleware(server.LoggingMiddleware()),
    server.WithMiddleware(server.RecoveryMiddleware()),
)
```

### Tracing Outgoing HTTP Requests

```go
// Start parent span
ctx, span := tracing.StartSpan(ctx, "fetch-user-data", tracing.SpanKind.Internal)
defer span.End()

// Create HTTP request
req, _ := http.NewRequest("GET", "http://api.example.com/users/123", nil)

// Trace the outgoing request
ctx, clientSpan := tracing.TraceOutgoingRequest(ctx, req, "GET /users/123")
defer clientSpan.End()

// Make the request
client := &http.Client{}
resp, err := client.Do(req)

// Record the response
tracing.RecordOutgoingResponse(ctx, resp, err)

// Or use the helper function
resp, err := tracing.WithHTTPClientTrace(ctx, req, client)
```

### Extracting Trace IDs for Logging

```go
ctx, span := tracing.StartSpan(ctx, "operation", tracing.SpanKind.Internal)
defer span.End()

// Get individual IDs
traceID := tracing.GetTraceID(ctx)
spanID := tracing.GetSpanID(ctx)

log.Printf("Processing request [trace_id=%s span_id=%s]", traceID, spanID)

// Or get as a map for structured logging
info := tracing.GetTracingInfo(ctx)
log.Printf("Request info: %+v", info)
```

### Context Propagation

```go
// In your HTTP server handler
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Extract trace context from incoming request
    ctx := tracing.ExtractContext(r.Context(), r)

    // Start a span with the extracted context
    ctx, span := tracing.StartSpan(ctx, "handle-request", tracing.SpanKind.Server)
    defer span.End()

    // When making an outgoing request, inject the context
    outgoingReq, _ := http.NewRequest("GET", "http://api.example.com/data", nil)
    tracing.InjectContext(ctx, outgoingReq)

    // The outgoing request now carries the trace context
    client := &http.Client{}
    resp, _ := client.Do(outgoingReq)
    defer resp.Body.Close()
}
```

## Configuration Options

### Tracing Config

```go
type Config struct {
    ServiceName    string  // Name of your service
    ServiceVersion string  // Version of your service
    Environment    string  // Environment (dev, staging, prod)
    ExporterType   string  // "stdout" or "otlp"
    OTLPEndpoint   string  // Endpoint for OTLP exporter
    SamplingRate   float64 // 0.0 to 1.0 (1.0 = 100% sampling)
    Enabled        bool    // Enable/disable tracing
}
```

### Middleware Config

```go
type MiddlewareConfig struct {
    SpanNameFormatter  func(*http.Request) string
    ExcludePaths       map[string]bool
    RecordRequestBody  bool
    RecordResponseBody bool
    CustomAttributes   func(*http.Request) []attribute.KeyValue
}
```

## Environment Variables

OpenTelemetry SDK respects standard environment variables:

- `OTEL_SERVICE_NAME`: Service name (overrides config)
- `OTEL_SDK_DISABLED`: Set to `true` to disable tracing
- `OTEL_TRACES_EXPORTER`: Exporter type (stdout, otlp, etc.)
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP endpoint
- `OTEL_TRACES_SAMPLER`: Sampler type (always_on, always_off, traceidratio)
- `OTEL_TRACES_SAMPLER_ARG`: Sampler argument (e.g., "0.1" for 10% sampling)

## Deployment

### Development

For development, use the stdout exporter:

```go
config := &tracing.Config{
    ServiceName:  "glyphlang-dev",
    ExporterType: "stdout",
    SamplingRate: 1.0,
    Enabled:      true,
}
```

### Production with Jaeger

1. Deploy Jaeger (or another OTLP-compatible collector):

```bash
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

2. Configure your application:

```go
config := &tracing.Config{
    ServiceName:    "glyphlang",
    ServiceVersion: "1.0.0",
    Environment:    "production",
    ExporterType:   "otlp",
    OTLPEndpoint:   "jaeger:4317",
    SamplingRate:   0.1, // Sample 10% of traces
    Enabled:        true,
}
```

3. Access Jaeger UI at `http://localhost:16686`

### Production with Cloud Providers

#### Google Cloud Trace

```go
config := &tracing.Config{
    ServiceName:  "glyphlang",
    ExporterType: "otlp",
    OTLPEndpoint: "cloudtrace.googleapis.com:443",
    Enabled:      true,
}
```

#### AWS X-Ray

Use the AWS Distro for OpenTelemetry (ADOT) Collector.

#### Azure Monitor

Use the Azure Monitor OpenTelemetry Exporter.

## Best Practices

1. **Initialize Once**: Initialize tracing at application startup
2. **Always Close**: Use `defer tp.Shutdown(context.Background())`
3. **Context Propagation**: Always pass context through your call chain
4. **Meaningful Names**: Use descriptive span names (e.g., "HTTP GET /users/:id")
5. **Attributes Over Events**: Prefer attributes for structured data
6. **Error Recording**: Always record errors in spans
7. **Sampling in Production**: Use sampling to reduce overhead
8. **Exclude Health Checks**: Don't trace health/metrics endpoints
9. **TraceID in Logs**: Include trace IDs in log messages for correlation
10. **Test Tracing**: Use the in-memory exporter for testing

## Testing

The package includes comprehensive tests. Run them with:

```bash
go test ./pkg/tracing/...
```

Run benchmarks:

```bash
go test -bench=. ./pkg/tracing/...
```

## Performance Considerations

- **Overhead**: Tracing adds ~1-5% overhead with proper sampling
- **Sampling**: Use sampling in production (0.1 = 10% is often sufficient)
- **Batching**: The SDK batches spans by default for efficiency
- **Context Size**: Keep span attributes reasonably sized
- **Excluded Paths**: Exclude high-frequency, low-value endpoints

## Troubleshooting

### No Traces Appearing

1. Check if tracing is enabled: `config.Enabled = true`
2. Verify exporter endpoint is reachable
3. Check sampling rate: `config.SamplingRate > 0`
4. Look for initialization errors

### High Memory Usage

1. Reduce sampling rate
2. Check for span leaks (missing `span.End()`)
3. Limit attribute sizes
4. Configure batch size

### Missing Context

1. Ensure context is passed through all function calls
2. Use `ctx = tracing.ExtractContext(ctx, req)` for incoming requests
3. Use `tracing.InjectContext(ctx, req)` for outgoing requests

## API Reference

### Initialization

- `InitTracing(config *Config) (*TracerProvider, error)`
- `DefaultConfig() *Config`

### Span Management

- `StartSpan(ctx, name, opts...) (context.Context, trace.Span)`
- `SpanFromContext(ctx) trace.Span`
- `WithSpan(ctx, name, fn, opts...) error`

### Attributes & Events

- `SetAttributes(ctx, attrs...)`
- `AddEvent(ctx, name, attrs...)`
- `SetStatus(ctx, code, description)`
- `SetError(ctx, err)`
- `RecordError(ctx, err, attrs...)`

### Context Propagation

- `InjectContext(ctx, req)`
- `ExtractContext(ctx, req) context.Context`

### Trace IDs

- `GetTraceID(ctx) string`
- `GetSpanID(ctx) string`
- `GetTracingInfo(ctx) map[string]string`

### HTTP Tracing

- `HTTPTracingMiddleware(config) func(http.Handler) http.Handler`
- `TraceOutgoingRequest(ctx, req, name) (context.Context, trace.Span)`
- `RecordOutgoingResponse(ctx, resp, err)`
- `WithHTTPClientTrace(ctx, req, client) (*http.Response, error)`

## Contributing

When contributing to this package:

1. Add tests for new functionality
2. Update documentation
3. Follow OpenTelemetry semantic conventions
4. Ensure backwards compatibility
5. Run tests and benchmarks

## License

This package is part of the GLYPHLANG project.

## Resources

- [OpenTelemetry Go SDK](https://github.com/open-telemetry/opentelemetry-go)
- [OpenTelemetry Specification](https://opentelemetry.io/docs/specs/otel/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
- [Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/)
