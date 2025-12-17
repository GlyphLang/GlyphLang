# GLYPHLANG Server Integration Guide

This guide shows how to integrate OpenTelemetry distributed tracing with the GLYPHLANG server.

## Quick Start

### 1. Initialize Tracing in Your Main Application

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/glyphlang/glyph/pkg/server"
    "github.com/glyphlang/glyph/pkg/tracing"
)

func main() {
    // Initialize tracing
    tracingConfig := &tracing.Config{
        ServiceName:    "glyphlang-api",
        ServiceVersion: "1.0.0",
        Environment:    getEnv("ENVIRONMENT", "development"),
        ExporterType:   getEnv("OTEL_EXPORTER_TYPE", "stdout"),
        OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
        SamplingRate:   0.1, // Sample 10% of traces in production
        Enabled:        getEnv("OTEL_ENABLED", "true") == "true",
    }

    tp, err := tracing.InitTracing(tracingConfig)
    if err != nil {
        log.Fatalf("Failed to initialize tracing: %v", err)
    }

    // Ensure proper shutdown
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := tp.Shutdown(ctx); err != nil {
            log.Printf("Error shutting down tracer provider: %v", err)
        }
    }()

    // Create server with tracing middleware
    s := server.NewServer(
        server.WithAddr(":8080"),
        server.WithMiddleware(createTracingMiddleware()),
        server.WithMiddleware(server.LoggingMiddleware()),
        server.WithMiddleware(server.RecoveryMiddleware()),
    )

    // Register routes...
    // Start server...
}

// createTracingMiddleware creates a tracing middleware compatible with the server
func createTracingMiddleware() server.Middleware {
    return func(next server.RouteHandler) server.RouteHandler {
        return func(ctx *server.Context) error {
            // Extract trace context from incoming request
            traceCtx := tracing.ExtractContext(ctx.Request.Context(), ctx.Request)

            // Generate span name
            spanName := fmt.Sprintf("HTTP %s %s", ctx.Request.Method, ctx.Request.URL.Path)

            // Start span
            traceCtx, span := tracing.StartSpan(traceCtx, spanName, tracing.SpanKind.Server)
            defer span.End()

            // Update request context
            ctx.Request = ctx.Request.WithContext(traceCtx)

            // Add request attributes
            tracing.SetAttributes(traceCtx,
                attribute.String("http.method", ctx.Request.Method),
                attribute.String("http.url", ctx.Request.URL.Path),
                attribute.String("http.host", ctx.Request.Host),
            )

            // Add trace IDs to response headers
            if traceID := tracing.GetTraceID(traceCtx); traceID != "" {
                ctx.ResponseWriter.Header().Set("X-Trace-ID", traceID)
            }
            if spanID := tracing.GetSpanID(traceCtx); spanID != "" {
                ctx.ResponseWriter.Header().Set("X-Span-ID", spanID)
            }

            // Call next handler
            err := next(ctx)

            // Record error if present
            if err != nil {
                tracing.SetError(traceCtx, err)
            }

            // Set status code
            tracing.SetAttributes(traceCtx,
                attribute.Int("http.status_code", ctx.StatusCode),
            )

            return err
        }
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### 2. Use Standard HTTP Middleware (Alternative Approach)

If you prefer to use the standard HTTP middleware directly:

```go
import (
    "net/http"
    "github.com/glyphlang/glyph/pkg/tracing"
)

func main() {
    // Initialize tracing
    tp, err := tracing.InitTracing(tracing.DefaultConfig())
    if err != nil {
        log.Fatal(err)
    }
    defer tp.Shutdown(context.Background())

    // Create middleware config
    middlewareConfig := &tracing.MiddlewareConfig{
        SpanNameFormatter: func(req *http.Request) string {
            return fmt.Sprintf("HTTP %s %s", req.Method, req.URL.Path)
        },
        ExcludePaths: map[string]bool{
            "/health":  true,
            "/metrics": true,
        },
    }

    // Create HTTP handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Your handler logic
        w.WriteHeader(http.StatusOK)
    })

    // Wrap with tracing middleware
    tracedHandler := tracing.HTTPTracingMiddleware(middlewareConfig)(handler)

    // Start server
    http.ListenAndServe(":8080", tracedHandler)
}
```

### 3. Add Tracing to Route Handlers

Within your route handlers, you can create child spans:

```go
func handleUserRequest(ctx *server.Context) error {
    // Extract trace context from request
    traceCtx := ctx.Request.Context()

    // Create a child span for database operation
    dbCtx, dbSpan := tracing.StartSpan(traceCtx, "database.get_user", tracing.SpanKind.Internal)
    defer dbSpan.End()

    // Add custom attributes
    tracing.SetAttributes(dbCtx,
        attribute.String("user.id", ctx.PathParams["id"]),
        attribute.String("db.operation", "SELECT"),
    )

    // Perform database operation
    user, err := getUserFromDB(dbCtx, ctx.PathParams["id"])
    if err != nil {
        tracing.SetError(dbCtx, err)
        return err
    }

    // Add event
    tracing.AddEvent(dbCtx, "user_retrieved")

    // Return response
    return server.SendJSON(ctx, 200, user)
}
```

### 4. Trace Outgoing HTTP Requests

When making HTTP requests to other services:

```go
func callExternalAPI(ctx context.Context) error {
    // Create outgoing request
    req, _ := http.NewRequest("GET", "http://api.example.com/data", nil)

    // Trace the outgoing request
    reqCtx, span := tracing.TraceOutgoingRequest(ctx, req, "GET /data")
    defer span.End()

    // Add custom attributes
    tracing.SetAttributes(reqCtx,
        attribute.String("api.name", "example-api"),
    )

    // Make the request
    client := &http.Client{}
    req = req.WithContext(reqCtx)
    resp, err := client.Do(req)

    // Record the response
    tracing.RecordOutgoingResponse(reqCtx, resp, err)

    return err
}

// Or use the helper function
func callExternalAPISimple(ctx context.Context) error {
    req, _ := http.NewRequest("GET", "http://api.example.com/data", nil)
    client := &http.Client{}

    resp, err := tracing.WithHTTPClientTrace(ctx, req, client)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Process response...
    return nil
}
```

### 5. Add Trace IDs to Structured Logging

Enhance your logs with trace correlation:

```go
import "log"

func handleRequest(ctx *server.Context) error {
    traceCtx := ctx.Request.Context()

    // Get trace info
    info := tracing.GetTracingInfo(traceCtx)

    // Log with trace IDs
    log.Printf("[INFO] Processing request [trace_id=%s span_id=%s] path=%s",
        info["trace_id"],
        info["span_id"],
        ctx.Request.URL.Path,
    )

    // Your handler logic...

    return nil
}
```

## Configuration Options

### Environment Variables

The tracing system respects standard OpenTelemetry environment variables:

```bash
# Enable/disable tracing
export OTEL_SDK_DISABLED=false

# Service name
export OTEL_SERVICE_NAME=glyphlang-api

# Exporter type (stdout or otlp)
export OTEL_TRACES_EXPORTER=otlp

# OTLP endpoint
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317

# Sampling rate (0.0 to 1.0)
export OTEL_TRACES_SAMPLER=traceidratio
export OTEL_TRACES_SAMPLER_ARG=0.1
```

### Code Configuration

```go
config := &tracing.Config{
    ServiceName:    "glyphlang-api",
    ServiceVersion: "1.0.0",
    Environment:    "production",
    ExporterType:   "otlp",         // "stdout" or "otlp"
    OTLPEndpoint:   "jaeger:4317",  // Your OTLP collector
    SamplingRate:   0.1,             // Sample 10% of traces
    Enabled:        true,
}
```

## Deployment Scenarios

### Development

```go
config := &tracing.Config{
    ServiceName:  "glyphlang-dev",
    ExporterType: "stdout",
    SamplingRate: 1.0,  // Trace everything in dev
    Enabled:      true,
}
```

### Production with Jaeger

```yaml
# docker-compose.yml
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    ports:
      - "16686:16686"  # UI
      - "4317:4317"    # OTLP gRPC
      - "4318:4318"    # OTLP HTTP
```

```go
config := &tracing.Config{
    ServiceName:  "glyphlang",
    ExporterType: "otlp",
    OTLPEndpoint: "jaeger:4317",
    SamplingRate: 0.1,
    Enabled:      true,
}
```

## Best Practices

1. **Initialize Once**: Initialize tracing at application startup
2. **Always Shutdown**: Use defer to ensure proper cleanup
3. **Propagate Context**: Always pass context.Context through your call chain
4. **Meaningful Span Names**: Use descriptive names like "HTTP GET /users/:id"
5. **Add Attributes**: Include relevant metadata (user IDs, operation types, etc.)
6. **Record Errors**: Always call `tracing.SetError()` or `tracing.RecordError()`
7. **Sample in Production**: Use sampling to reduce overhead (10% is often sufficient)
8. **Exclude Health Checks**: Don't trace /health, /metrics endpoints
9. **Include Trace IDs in Logs**: Add trace IDs to all log messages
10. **Test Without Tracing**: Ensure your app works with tracing disabled

## Troubleshooting

### No Traces Appearing

```bash
# Check if tracing is enabled
export OTEL_SDK_DISABLED=false

# Verify endpoint is reachable
curl http://localhost:4317

# Check sampling rate
export OTEL_TRACES_SAMPLER_ARG=1.0
```

### High Overhead

```go
// Reduce sampling rate
config.SamplingRate = 0.01  // 1%

// Exclude high-frequency endpoints
excludePaths := map[string]bool{
    "/health": true,
    "/ping": true,
    "/metrics": true,
}
```

### Missing Context

```go
// Always pass context through the chain
func handler(ctx *server.Context) error {
    traceCtx := ctx.Request.Context()

    // Pass to child functions
    result, err := childFunction(traceCtx)
    if err != nil {
        return err
    }

    return nil
}

func childFunction(ctx context.Context) (Result, error) {
    // Create child span
    ctx, span := tracing.StartSpan(ctx, "child-operation", tracing.SpanKind.Internal)
    defer span.End()

    // Do work...
    return result, nil
}
```

## Complete Example

See the `example_test.go` file for complete working examples.

## Resources

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
- [Jaeger](https://www.jaegertracing.io/)
- [GLYPHLANG Tracing Package README](./README.md)
