# OpenTelemetry Tracing Quick Start

## 30-Second Setup

```go
import "github.com/glyphlang/glyph/pkg/tracing"

// 1. Initialize (in main.go)
tp, _ := tracing.InitTracing(tracing.DefaultConfig())
defer tp.Shutdown(context.Background())

// 2. Add to server
s := server.NewServer(
    server.WithMiddleware(createTracingMiddleware()),
)

// 3. Use in handlers
ctx, span := tracing.StartSpan(ctx, "operation", tracing.SpanKind.Internal)
defer span.End()
```

## Common Operations

### Start a Span
```go
ctx, span := tracing.StartSpan(ctx, "operation-name", tracing.SpanKind.Internal)
defer span.End()
```

### Add Attributes
```go
tracing.SetAttributes(ctx,
    attribute.String("user.id", "123"),
    attribute.Int("items.count", 5),
)
```

### Record Error
```go
if err != nil {
    tracing.SetError(ctx, err)
}
```

### Add Event
```go
tracing.AddEvent(ctx, "cache_hit")
```

### Get Trace IDs
```go
traceID := tracing.GetTraceID(ctx)
spanID := tracing.GetSpanID(ctx)
```

## Configuration Presets

### Development
```go
config := &tracing.Config{
    ServiceName:  "my-service",
    ExporterType: "stdout",
    SamplingRate: 1.0,
    Enabled:      true,
}
```

### Production (Jaeger)
```go
config := &tracing.Config{
    ServiceName:  "my-service",
    ExporterType: "otlp",
    OTLPEndpoint: "jaeger:4317",
    SamplingRate: 0.1,  // 10%
    Enabled:      true,
}
```

## Middleware Setup

```go
func createTracingMiddleware() server.Middleware {
    return func(next server.RouteHandler) server.RouteHandler {
        return func(ctx *server.Context) error {
            traceCtx := tracing.ExtractContext(ctx.Request.Context(), ctx.Request)
            traceCtx, span := tracing.StartSpan(traceCtx,
                fmt.Sprintf("HTTP %s %s", ctx.Request.Method, ctx.Request.URL.Path),
                tracing.SpanKind.Server)
            defer span.End()

            ctx.Request = ctx.Request.WithContext(traceCtx)

            // Add trace headers
            ctx.ResponseWriter.Header().Set("X-Trace-ID", tracing.GetTraceID(traceCtx))

            err := next(ctx)
            if err != nil {
                tracing.SetError(traceCtx, err)
            }
            return err
        }
    }
}
```

## Docker Compose (Jaeger)

```yaml
version: '3'
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    ports:
      - "16686:16686"  # UI
      - "4317:4317"    # OTLP gRPC
```

```bash
docker-compose up -d
# Access UI: http://localhost:16686
```

## Environment Variables

```bash
export OTEL_SERVICE_NAME=my-service
export OTEL_EXPORTER_TYPE=otlp
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
export OTEL_TRACES_SAMPLER_ARG=0.1
```

## Common Patterns

### Database Query
```go
ctx, span := tracing.StartSpan(ctx, "db.query", tracing.SpanKind.Internal)
defer span.End()
tracing.SetAttributes(ctx,
    attribute.String("db.operation", "SELECT"),
    attribute.String("db.table", "users"),
)
result, err := db.Query(ctx, query)
if err != nil {
    tracing.SetError(ctx, err)
}
```

### HTTP Client Request
```go
req, _ := http.NewRequest("GET", "http://api.example.com/data", nil)
client := &http.Client{}
resp, err := tracing.WithHTTPClientTrace(ctx, req, client)
```

### Logging with Trace IDs
```go
info := tracing.GetTracingInfo(ctx)
log.Printf("[INFO] Processing [trace_id=%s span_id=%s] message",
    info["trace_id"], info["span_id"])
```

## Testing

```bash
# Run tests
go test ./pkg/tracing/...

# With coverage
go test ./pkg/tracing/... -cover

# Benchmarks
go test ./pkg/tracing/... -bench=.
```

## Troubleshooting

### No traces appearing?
1. Check `config.Enabled = true`
2. Verify `SamplingRate > 0`
3. Check exporter endpoint

### High overhead?
1. Reduce sampling: `config.SamplingRate = 0.01`
2. Exclude paths in middleware
3. Check span creation frequency

### Missing context?
Always pass `context.Context` through your function chain

## Resources

- Full Documentation: [README.md](./README.md)
- Integration Guide: [INTEGRATION.md](./INTEGRATION.md)
- Example Server: [example_server.go.txt](./example_server.go.txt)
- Tests: [tracing_test.go](./tracing_test.go)

## Support

For issues or questions:
1. Check the [README.md](./README.md) troubleshooting section
2. Review the [INTEGRATION.md](./INTEGRATION.md) guide
3. Check the example code

---

**Remember**: Always initialize tracing once and shut it down gracefully!
