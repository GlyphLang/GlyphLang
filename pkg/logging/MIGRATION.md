# Migration Guide: Basic to Structured Logging

This guide shows how to migrate from the basic `LoggingMiddleware` in `pkg/server/middleware.go` to the new structured logging system.

## Before (Basic Logging)

```go
package main

import (
    "github.com/glyphlang/glyph/pkg/server"
)

func main() {
    router := server.NewRouter()

    // Old basic logging middleware
    loggingMW := server.LoggingMiddleware()
    recoveryMW := server.RecoveryMiddleware()

    handler := func(ctx *server.Context) error {
        // Handler code
        return server.SendJSON(ctx, 200, map[string]interface{}{
            "message": "success",
        })
    }

    router.RegisterRoute(&server.Route{
        Method:      server.GET,
        Path:        "/api/users/:id",
        Handler:     handler,
        Middlewares: []server.Middleware{loggingMW, recoveryMW},
    })
}
```

**Output (Text format, basic):**
```
2025/12/13 16:22:09 [REQUEST] GET /api/users/123
2025/12/13 16:22:09 [RESPONSE] GET /api/users/123 - 200 (5.2ms)
```

## After (Structured Logging)

```go
package main

import (
    "github.com/glyphlang/glyph/pkg/logging"
    "github.com/glyphlang/glyph/pkg/server"
)

func main() {
    // Step 1: Create structured logger
    logger, err := logging.NewLogger(logging.LoggerConfig{
        MinLevel:      logging.INFO,
        Format:        logging.JSONFormat,
        IncludeCaller: true,
        FilePath:      "/var/log/myapp/app.log",
        MaxFileSize:   50 * 1024 * 1024, // 50MB
        MaxBackups:    10,
    })
    if err != nil {
        panic(err)
    }
    defer logger.Close()

    router := server.NewRouter()

    // Step 2: Replace with structured logging middlewares
    loggingMW := logging.StructuredLoggingMiddleware(logger)
    recoveryMW := logging.StructuredRecoveryMiddleware(logger)

    handler := func(ctx *server.Context) error {
        // Step 3: Use logger in handlers
        reqLogger := logging.GetRequestLogger(logger, ctx.Request)

        userID := ctx.PathParams["id"]
        reqLogger.InfoWithFields("Fetching user", map[string]interface{}{
            "user_id": userID,
        })

        // Handler code
        return server.SendJSON(ctx, 200, map[string]interface{}{
            "message": "success",
        })
    }

    router.RegisterRoute(&server.Route{
        Method:      server.GET,
        Path:        "/api/users/:id",
        Handler:     handler,
        Middlewares: []server.Middleware{loggingMW, recoveryMW},
    })
}
```

**Output (JSON format, structured):**
```json
{"timestamp":"2025-12-13T16:22:09.123Z","level":"INFO","message":"request started","request_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567890","fields":{"method":"GET","path":"/api/users/123","remote_ip":"192.168.1.1","user_agent":"Mozilla/5.0","query":""}}
{"timestamp":"2025-12-13T16:22:09.125Z","level":"INFO","message":"Fetching user","request_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567890","fields":{"method":"GET","path":"/api/users/123","user_id":"123"}}
{"timestamp":"2025-12-13T16:22:09.128Z","level":"INFO","message":"request completed","request_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567890","fields":{"method":"GET","path":"/api/users/123","remote_ip":"192.168.1.1","user_agent":"Mozilla/5.0","status":200,"duration_ms":5,"response_size":27}}
```

## Key Benefits of Migration

### 1. Request ID Tracking
Every request gets a unique UUID that's included in all logs for that request.

```go
// Automatic request ID in all logs
reqLogger := logging.GetRequestLogger(logger, ctx.Request)
reqLogger.Info("Step 1")
reqLogger.Info("Step 2")
// Both logs have the same request_id
```

### 2. Structured Fields
Easy to parse and search in log aggregation systems.

```go
// Old way: Hard to parse
log.Printf("User %s performed action %s", userID, action)

// New way: Structured and parseable
reqLogger.InfoWithFields("User action", map[string]interface{}{
    "user_id": userID,
    "action": action,
})
```

### 3. Multiple Output Targets
Write to both console and file simultaneously.

```go
logger, _ := logging.NewLogger(logging.LoggerConfig{
    FilePath: "/var/log/myapp/app.log",
    Outputs:  []io.Writer{os.Stdout}, // Also goes to file via FilePath
})
```

### 4. Log Rotation
Automatic log file rotation based on size.

```go
logger, _ := logging.NewLogger(logging.LoggerConfig{
    FilePath:    "/var/log/myapp/app.log",
    MaxFileSize: 50 * 1024 * 1024, // Rotate at 50MB
    MaxBackups:  10,                // Keep 10 old files
})
```

### 5. Better Error Logging
Enhanced error logging with stack traces.

```go
logger, _ := logging.NewLogger(logging.LoggerConfig{
    IncludeStackTrace: true, // Automatic stack traces for errors
})
```

## Step-by-Step Migration

### Step 1: Create Logger Configuration

Add logger initialization to your main package:

```go
// config/logger.go
package config

import "github.com/glyphlang/glyph/pkg/logging"

func NewLogger() (*logging.Logger, error) {
    return logging.NewLogger(logging.LoggerConfig{
        MinLevel:          logging.INFO,
        Format:            logging.JSONFormat,
        IncludeCaller:     true,
        IncludeStackTrace: true,
        BufferSize:        5000,
        FilePath:          "/var/log/myapp/app.log",
        MaxFileSize:       50 * 1024 * 1024,
        MaxBackups:        10,
    })
}
```

### Step 2: Replace Middleware

Replace in all route registrations:

```go
// Before
loggingMW := server.LoggingMiddleware()
recoveryMW := server.RecoveryMiddleware()

// After
loggingMW := logging.StructuredLoggingMiddleware(logger)
recoveryMW := logging.StructuredRecoveryMiddleware(logger)
```

### Step 3: Update Handlers

Add logging to your handlers:

```go
// Before
func GetUser(ctx *server.Context) error {
    userID := ctx.PathParams["id"]
    // No structured logging
    return server.SendJSON(ctx, 200, userData)
}

// After
func GetUser(logger *logging.Logger) server.RouteHandler {
    return func(ctx *server.Context) error {
        reqLogger := logging.GetRequestLogger(logger, ctx.Request)

        userID := ctx.PathParams["id"]
        reqLogger.InfoWithFields("Fetching user", map[string]interface{}{
            "user_id": userID,
        })

        return server.SendJSON(ctx, 200, userData)
    }
}
```

### Step 4: Replace log.Printf Calls

Replace standard library logging:

```go
// Before
import "log"
log.Printf("User %s logged in", username)

// After
reqLogger.InfoWithFields("User logged in", map[string]interface{}{
    "username": username,
})
```

## Environment-Specific Configuration

### Development

```go
logger, _ := logging.NewLogger(logging.LoggerConfig{
    MinLevel: logging.DEBUG,    // Show all logs
    Format:   logging.TextFormat, // Human-readable
})
```

### Production

```go
logger, _ := logging.NewLogger(logging.LoggerConfig{
    MinLevel:          logging.INFO,      // Only INFO and above
    Format:            logging.JSONFormat, // For log aggregation
    IncludeCaller:     true,
    IncludeStackTrace: true,
    FilePath:          "/var/log/myapp/app.log",
    MaxFileSize:       50 * 1024 * 1024,
    MaxBackups:        10,
})
```

## Common Patterns

### Pattern 1: Request Context Logger

```go
func MyHandler(logger *logging.Logger) server.RouteHandler {
    return func(ctx *server.Context) error {
        // Create request logger once
        reqLogger := logging.GetRequestLogger(logger, ctx.Request)

        // Use throughout handler
        reqLogger.Info("Starting processing")

        if err := doSomething(); err != nil {
            reqLogger.ErrorWithFields("Processing failed", map[string]interface{}{
                "error": err.Error(),
            })
            return err
        }

        reqLogger.Info("Processing completed")
        return server.SendJSON(ctx, 200, result)
    }
}
```

### Pattern 2: Service Layer Logging

```go
type UserService struct {
    logger *logging.Logger
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
    // Extract request ID from context if available
    reqID := ctx.Value("request_id").(string)
    logger := s.logger.WithRequestID(reqID)

    logger.InfoWithFields("Fetching user from database", map[string]interface{}{
        "user_id": userID,
    })

    // Database query...

    return user, nil
}
```

### Pattern 3: Background Job Logging

```go
func ProcessJob(logger *logging.Logger, job *Job) error {
    // Create job-specific logger
    jobLogger := logger.WithFields(map[string]interface{}{
        "job_id":   job.ID,
        "job_type": job.Type,
    })

    jobLogger.Info("Job started")

    // Process job...

    jobLogger.Info("Job completed")
    return nil
}
```

## Troubleshooting

### Logs Not Appearing

Check buffer size and close logger properly:

```go
logger, _ := logging.NewLogger(logging.LoggerConfig{
    BufferSize: 1000, // Increase if needed
})
defer logger.Close() // IMPORTANT: Flushes buffer
```

### Performance Issues

Use async logging and increase buffer:

```go
logger, _ := logging.NewLogger(logging.LoggerConfig{
    BufferSize: 10000, // Larger buffer for high throughput
})
```

### File Permission Errors

Ensure log directory exists and has proper permissions:

```bash
sudo mkdir -p /var/log/myapp
sudo chown myuser:mygroup /var/log/myapp
sudo chmod 755 /var/log/myapp
```

## Testing with Structured Logging

```go
func TestHandler(t *testing.T) {
    var buf bytes.Buffer

    // Create test logger
    logger, _ := logging.NewLogger(logging.LoggerConfig{
        MinLevel: logging.DEBUG,
        Format:   logging.JSONFormat,
        Outputs:  []io.Writer{&buf},
    })
    defer logger.Close()

    // Test handler
    handler := MyHandler(logger)

    // Make request...

    // Verify logs
    time.Sleep(50 * time.Millisecond) // Wait for async logging

    var entry logging.LogEntry
    json.Unmarshal(buf.Bytes(), &entry)

    if entry.Message != "expected message" {
        t.Errorf("Unexpected log message: %s", entry.Message)
    }
}
```

## Rollback Plan

If you need to rollback:

1. Keep both middleware imports
2. Switch back to `server.LoggingMiddleware()`
3. Remove structured logger initialization
4. Remove `logging.GetRequestLogger()` calls

The old middleware will continue to work alongside the new package.
