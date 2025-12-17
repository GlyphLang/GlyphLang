# Structured Logging Package

A production-ready structured logging package for GLYPHLANG with comprehensive features for high-performance applications.

## Features

- **Multiple Log Levels**: DEBUG, INFO, WARN, ERROR, FATAL
- **Structured Logging**: JSON and text formats
- **Request ID Tracking**: Automatic UUID generation for request tracing
- **Async Logging**: Buffered async processing for high performance
- **Contextual Logging**: Add fields to log context for enriched logs
- **Log Rotation**: Automatic file rotation based on size
- **Multiple Outputs**: Write to stdout, files, or custom writers
- **Middleware Integration**: Ready-to-use HTTP middleware
- **Thread-Safe**: Concurrent logging support
- **Caller Information**: Optional file and line number logging
- **Stack Traces**: Automatic stack traces for errors
- **Zero Dependencies**: Uses only standard library (except UUID)

## Quick Start

### Basic Usage

```go
import "github.com/glyphlang/glyph/pkg/logging"

// Create a logger
logger, err := logging.NewLogger(logging.LoggerConfig{
    MinLevel: logging.INFO,
    Format:   logging.TextFormat,
})
if err != nil {
    panic(err)
}
defer logger.Close()

// Simple logging
logger.Info("Application started")
logger.Warn("Low disk space")
logger.Error("Failed to connect to database")
```

### Structured Logging with Fields

```go
logger.InfoWithFields("User logged in", map[string]interface{}{
    "user_id":  12345,
    "username": "john_doe",
    "ip":       "192.168.1.1",
})
```

### Context Logger with Request ID

```go
// Create a context logger with request ID
requestID := logging.NewRequestID()
ctxLogger := logger.WithRequestID(requestID)

// All logs will include the request ID
ctxLogger.Info("Processing request")
ctxLogger.Info("Request completed")
```

### Chaining Context

```go
userLogger := logger.
    WithRequestID("req-123").
    WithField("user_id", 123).
    WithField("action", "update")

userLogger.Info("User action performed")
```

## Configuration

### Logger Configuration

```go
config := logging.LoggerConfig{
    // Minimum log level (default: INFO)
    MinLevel: logging.DEBUG,

    // Output format: TextFormat or JSONFormat (default: TextFormat)
    Format: logging.JSONFormat,

    // Include caller file and line number
    IncludeCaller: true,

    // Include stack traces for ERROR and FATAL
    IncludeStackTrace: true,

    // Async buffer size (default: 1000)
    BufferSize: 5000,

    // File path for file logging (empty = no file logging)
    FilePath: "/var/log/myapp/app.log",

    // Maximum file size before rotation (0 = no rotation)
    MaxFileSize: 50 * 1024 * 1024, // 50MB

    // Maximum number of old files to keep
    MaxBackups: 10,

    // Custom output writers
    Outputs: []io.Writer{os.Stdout},
}

logger, err := logging.NewLogger(config)
```

### Production Configuration

```go
logger, err := logging.NewLogger(logging.LoggerConfig{
    MinLevel:          logging.INFO,
    Format:            logging.JSONFormat,
    IncludeCaller:     true,
    IncludeStackTrace: true,
    BufferSize:        5000,
    FilePath:          "/var/log/myapp/app.log",
    MaxFileSize:       50 * 1024 * 1024, // 50MB
    MaxBackups:        10,
})
```

## HTTP Middleware

### Structured Logging Middleware

```go
import (
    "github.com/glyphlang/glyph/pkg/logging"
    "github.com/glyphlang/glyph/pkg/server"
)

// Create logger
logger, _ := logging.NewLogger(logging.LoggerConfig{
    MinLevel: logging.INFO,
    Format:   logging.JSONFormat,
})

// Create router
router := server.NewRouter()

// Create middlewares
loggingMW := logging.StructuredLoggingMiddleware(logger)
recoveryMW := logging.StructuredRecoveryMiddleware(logger)

// Register route with middlewares
router.RegisterRoute(&server.Route{
    Method:      server.GET,
    Path:        "/users/:id",
    Handler:     handler,
    Middlewares: []server.Middleware{loggingMW, recoveryMW},
})
```

### Body Logging Middleware (Debug Only)

**WARNING**: This logs request and response bodies. Use only in development.

```go
bodyLoggingMW := logging.StructuredLoggingMiddlewareWithBodyLogging(logger, 10240)

router.RegisterRoute(&server.Route{
    Method:      server.POST,
    Path:        "/api/users",
    Handler:     handler,
    Middlewares: []server.Middleware{bodyLoggingMW},
})
```

### Recovery Middleware

Recovers from panics and logs them with structured logging:

```go
recoveryMW := logging.StructuredRecoveryMiddleware(logger)
```

### Using Logger in Handlers

```go
handler := func(ctx *server.Context) error {
    // Get logger for this request
    reqLogger := logging.GetRequestLogger(logger, ctx.Request)

    reqLogger.InfoWithFields("Processing request", map[string]interface{}{
        "user_id": ctx.PathParams["id"],
    })

    return server.SendJSON(ctx, 200, data)
}
```

## Default Logger

For simple applications, use the default logger:

```go
// Initialize once at application start
logging.InitDefaultLogger(logging.LoggerConfig{
    MinLevel: logging.INFO,
    Format:   logging.JSONFormat,
})

// Use anywhere in your code
logging.Info("Application started")
logging.Warn("Low disk space")

// With context
reqLogger := logging.WithRequestID("req-123")
reqLogger.Info("Processing request")
```

## Log Levels

```go
logger.Debug("Detailed debug information")
logger.Info("Informational message")
logger.Warn("Warning: something might be wrong")
logger.Error("Error occurred but application continues")
logger.Fatal("Critical error - application will exit")
```

## Output Formats

### Text Format

```
[2025-12-13 16:22:09.878] [INFO] [req-123] Application started
[2025-12-13 16:22:09.888] [INFO] [req-123] User logged in {user_id=12345, username=john_doe}
```

### JSON Format

```json
{"timestamp":"2025-12-13T16:22:09.878Z","level":"INFO","message":"Application started","request_id":"req-123"}
{"timestamp":"2025-12-13T16:22:09.888Z","level":"INFO","message":"User logged in","request_id":"req-123","fields":{"user_id":12345,"username":"john_doe"}}
```

## Log Rotation

Automatic file rotation based on file size:

```go
logger, _ := logging.NewLogger(logging.LoggerConfig{
    FilePath:    "/var/log/myapp/app.log",
    MaxFileSize: 10 * 1024 * 1024, // 10MB
    MaxBackups:  5,                 // Keep 5 old files
})
```

When the log file reaches 10MB:
- Current file renamed to `app.log.1`
- Previous backups shifted (`.1` → `.2`, `.2` → `.3`, etc.)
- New `app.log` created
- Oldest backup (`.5`) deleted if limit reached

## Request ID Tracking

Every HTTP request gets a unique request ID for distributed tracing:

```go
// Auto-generated if not present
requestID := logging.NewRequestID() // UUID v4

// Or use existing request ID from header
// X-Request-ID header is automatically checked and added
```

## Performance

The logger is designed for high performance:

- **Async Logging**: Non-blocking with buffered channel
- **Efficient Serialization**: Optimized JSON encoding
- **Minimal Allocations**: Reuses buffers where possible

Benchmark results (AMD Ryzen 7 7800X3D):
```
BenchmarkLoggerInfo-16                     381.5 ns/op    577 B/op    13 allocs/op
BenchmarkLoggerInfoWithFields-16           596.7 ns/op    738 B/op    10 allocs/op
BenchmarkContextLogger-16                  565.3 ns/op    721 B/op     8 allocs/op
BenchmarkStructuredLoggingMiddleware-16    2577 ns/op    4773 B/op    59 allocs/op
```

## Best Practices

### 1. Always Close Loggers

```go
logger, err := logging.NewLogger(config)
if err != nil {
    panic(err)
}
defer logger.Close() // Ensures all buffered logs are written
```

### 2. Use Context Loggers for Requests

```go
// Create once per request
reqLogger := logger.WithRequestID(requestID)

// Use throughout request lifecycle
reqLogger.Info("Request started")
reqLogger.Info("Database query executed")
reqLogger.Info("Request completed")
```

### 3. Add Structured Fields

```go
// Good: Structured fields for easy parsing
logger.InfoWithFields("User action", map[string]interface{}{
    "user_id": 123,
    "action": "login",
    "ip": "192.168.1.1",
})

// Bad: String interpolation
logger.Info(fmt.Sprintf("User %d performed action %s from %s", 123, "login", "192.168.1.1"))
```

### 4. Use Appropriate Log Levels

- **DEBUG**: Detailed diagnostic information (disabled in production)
- **INFO**: General informational messages
- **WARN**: Warning conditions that might need attention
- **ERROR**: Error conditions that don't stop the application
- **FATAL**: Critical errors that require application shutdown

### 5. Don't Log Sensitive Data

```go
// Bad: Logging passwords or tokens
logger.InfoWithFields("User login", map[string]interface{}{
    "username": "john",
    "password": "secret123", // NEVER DO THIS
})

// Good: Log only non-sensitive data
logger.InfoWithFields("User login", map[string]interface{}{
    "username": "john",
    "success": true,
})
```

## Thread Safety

All logging operations are thread-safe and can be called from multiple goroutines:

```go
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        logger.InfoWithFields("Concurrent log", map[string]interface{}{
            "goroutine": id,
        })
    }(i)
}
wg.Wait()
```

## Error Handling

The logger handles errors gracefully:

- Buffer overflow: Falls back to synchronous logging
- Write errors: Logs to stderr
- Closed logger: Silently ignores new log entries

## License

Part of the GLYPHLANG project.
