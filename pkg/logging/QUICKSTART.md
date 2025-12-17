# Structured Logging - Quick Start

Get up and running with structured logging in 5 minutes.

## Installation

The logging package is already part of the GLYPHLANG project. Just import it:

```go
import "github.com/glyphlang/glyph/pkg/logging"
```

## 1. Basic Usage (30 seconds)

```go
package main

import "github.com/glyphlang/glyph/pkg/logging"

func main() {
    // Create logger
    logger, _ := logging.NewLogger(logging.LoggerConfig{
        MinLevel: logging.INFO,
        Format:   logging.TextFormat,
    })
    defer logger.Close()

    // Log messages
    logger.Info("Application started")
    logger.Warn("Warning message")
    logger.Error("Error message")
}
```

**Output:**
```
[2025-12-13 16:22:09.123] [INFO] Application started
[2025-12-13 16:22:09.124] [WARN] Warning message
[2025-12-13 16:22:09.125] [ERROR] Error message
```

## 2. Structured Logging (1 minute)

```go
logger, _ := logging.NewLogger(logging.LoggerConfig{
    MinLevel: logging.INFO,
    Format:   logging.JSONFormat,  // Use JSON format
})
defer logger.Close()

// Log with structured fields
logger.InfoWithFields("User logged in", map[string]interface{}{
    "user_id":  12345,
    "username": "john_doe",
    "ip":       "192.168.1.1",
})
```

**Output:**
```json
{"timestamp":"2025-12-13T16:22:09.123Z","level":"INFO","message":"User logged in","fields":{"user_id":12345,"username":"john_doe","ip":"192.168.1.1"}}
```

## 3. Request Tracking (2 minutes)

```go
logger, _ := logging.NewLogger(logging.LoggerConfig{
    MinLevel: logging.INFO,
    Format:   logging.JSONFormat,
})
defer logger.Close()

// Create context logger with request ID
requestID := logging.NewRequestID()
reqLogger := logger.WithRequestID(requestID)

// All logs include the same request ID
reqLogger.Info("Request started")
reqLogger.Info("Processing data")
reqLogger.Info("Request completed")
```

**Output:**
```json
{"timestamp":"2025-12-13T16:22:09.123Z","level":"INFO","message":"Request started","request_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567890"}
{"timestamp":"2025-12-13T16:22:09.124Z","level":"INFO","message":"Processing data","request_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567890"}
{"timestamp":"2025-12-13T16:22:09.125Z","level":"INFO","message":"Request completed","request_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567890"}
```

## 4. HTTP Middleware (3 minutes)

```go
import (
    "github.com/glyphlang/glyph/pkg/logging"
    "github.com/glyphlang/glyph/pkg/server"
)

func main() {
    // Create logger
    logger, _ := logging.NewLogger(logging.LoggerConfig{
        MinLevel: logging.INFO,
        Format:   logging.JSONFormat,
    })
    defer logger.Close()

    // Create router
    router := server.NewRouter()

    // Add structured logging middleware
    loggingMW := logging.StructuredLoggingMiddleware(logger)
    recoveryMW := logging.StructuredRecoveryMiddleware(logger)

    // Create handler
    handler := func(ctx *server.Context) error {
        return server.SendJSON(ctx, 200, map[string]interface{}{
            "message": "success",
        })
    }

    // Register route with middlewares
    router.RegisterRoute(&server.Route{
        Method:      server.GET,
        Path:        "/api/users/:id",
        Handler:     handler,
        Middlewares: []server.Middleware{loggingMW, recoveryMW},
    })

    // Start server...
}
```

## 5. Production Config (5 minutes)

```go
logger, err := logging.NewLogger(logging.LoggerConfig{
    // Log level
    MinLevel: logging.INFO,

    // JSON format for log aggregation tools
    Format: logging.JSONFormat,

    // Include file and line numbers
    IncludeCaller: true,

    // Include stack traces for errors
    IncludeStackTrace: true,

    // Async buffer size for high throughput
    BufferSize: 5000,

    // File logging with rotation
    FilePath:    "/var/log/myapp/app.log",
    MaxFileSize: 50 * 1024 * 1024, // 50MB
    MaxBackups:  10,                // Keep 10 old files
})
if err != nil {
    panic(err)
}
defer logger.Close()

logger.Info("Production application started")
```

## Common Patterns

### Pattern: Use in Handler

```go
func GetUser(logger *logging.Logger) server.RouteHandler {
    return func(ctx *server.Context) error {
        // Get logger for this request
        reqLogger := logging.GetRequestLogger(logger, ctx.Request)

        userID := ctx.PathParams["id"]
        reqLogger.InfoWithFields("Fetching user", map[string]interface{}{
            "user_id": userID,
        })

        // Your logic here...

        return server.SendJSON(ctx, 200, userData)
    }
}
```

### Pattern: Chain Context

```go
// Start with request ID
reqLogger := logger.WithRequestID(requestID)

// Add user context
userLogger := reqLogger.
    WithField("user_id", 123).
    WithField("role", "admin")

// All logs include request_id, user_id, and role
userLogger.Info("User action performed")
```

### Pattern: Default Logger

For simple applications:

```go
// Initialize once at startup
logging.InitDefaultLogger(logging.LoggerConfig{
    MinLevel: logging.INFO,
    Format:   logging.JSONFormat,
})

// Use anywhere in your code
logging.Info("Message")
logging.WithRequestID("req-123").Info("Request message")
```

## Log Levels

```go
logger.Debug("Detailed debug info")    // Development only
logger.Info("Informational message")   // General info
logger.Warn("Warning condition")       // Potential issue
logger.Error("Error occurred")         // Error but app continues
logger.Fatal("Critical error")         // Logs and exits app
```

## Environment Setup

### Development
```go
logging.LoggerConfig{
    MinLevel: logging.DEBUG,      // Show all logs
    Format:   logging.TextFormat, // Human-readable
}
```

### Production
```go
logging.LoggerConfig{
    MinLevel:      logging.INFO,      // INFO and above
    Format:        logging.JSONFormat, // Machine-parseable
    FilePath:      "/var/log/myapp/app.log",
    IncludeCaller: true,
}
```

## Next Steps

- Read [README.md](README.md) for full feature documentation
- Read [MIGRATION.md](MIGRATION.md) for migration from basic logging
- Check [example_test.go](example_test.go) for more examples

## Troubleshooting

### Logs not appearing?
```go
defer logger.Close() // Don't forget to close!
```

### Need more performance?
```go
logging.LoggerConfig{
    BufferSize: 10000, // Increase buffer
}
```

### File permission errors?
```bash
mkdir -p /var/log/myapp
chmod 755 /var/log/myapp
```

## Tips

1. **Always close the logger**: Use `defer logger.Close()` to ensure logs are flushed
2. **Use structured fields**: Better than string formatting for parsing
3. **One logger per application**: Create once, use everywhere
4. **Request IDs**: Track requests across services
5. **Appropriate log levels**: Don't log DEBUG in production

## Help

For issues or questions:
- Check the [README.md](README.md)
- Look at [example_test.go](example_test.go)
- Read the inline code documentation
