# Glyph HTTP Server - Quick Start Guide

Get the Glyph HTTP server up and running in 5 minutes!

## Prerequisites

- Go 1.21 or higher
- curl (for testing)

## Option 1: Run the Demo Server (Fastest)

```bash
# Navigate to demo directory
cd examples/server-demo

# Run the demo server
go run main.go
```

The server starts on `http://localhost:8080` with these endpoints:

- `GET /hello` - Hello world
- `GET /greet/:name` - Greeting with name
- `GET /health` - Health check
- `GET /api/users` - List users
- `GET /api/users/:id` - Get user by ID
- `POST /api/users` - Create user
- `PUT /api/users/:id` - Update user
- `PATCH /api/users/:id` - Partial update
- `DELETE /api/users/:id` - Delete user
- `GET /api/users/:userId/posts/:postId` - Nested resources

### Test the Server

In another terminal:

```bash
# Simple test
curl http://localhost:8080/hello

# Run all tests (Unix/Mac)
cd examples/server-demo
chmod +x test-endpoints.sh
./test-endpoints.sh

# Run all tests (Windows)
test-endpoints.bat
```

## Option 2: Build Your Own Server

### 1. Create a new Go file

```go
package main

import (
    "log"
    "github.com/glyphlang/glyph/pkg/server"
)

func main() {
    // Create server
    srv := server.NewServer(
        server.WithAddr(":8080"),
        server.WithInterpreter(&server.MockInterpreter{}),
    )

    // Register a route
    srv.RegisterRoute(&server.Route{
        Method: server.GET,
        Path:   "/hello",
    })

    // Start server
    log.Println("Server starting on :8080")
    if err := srv.Start(":8080"); err != nil {
        log.Fatal(err)
    }
}
```

### 2. Run it

```bash
go run main.go
```

### 3. Test it

```bash
curl http://localhost:8080/hello
```

## Common Usage Patterns

### Path Parameters

```go
srv.RegisterRoute(&server.Route{
    Method: server.GET,
    Path:   "/users/:id",
    Handler: func(ctx *server.Context) error {
        id := ctx.PathParams["id"]
        return server.SendJSON(ctx, 200, map[string]string{
            "id": id,
        })
    },
})
```

Test:
```bash
curl http://localhost:8080/users/123
```

### POST with JSON

```go
srv.RegisterRoute(&server.Route{
    Method: server.POST,
    Path:   "/users",
    Handler: func(ctx *server.Context) error {
        name := ctx.Body["name"].(string)
        email := ctx.Body["email"].(string)

        return server.SendJSON(ctx, 201, map[string]interface{}{
            "id":    123,
            "name":  name,
            "email": email,
        })
    },
})
```

Test:
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John", "email": "john@example.com"}'
```

### Query Parameters

```go
srv.RegisterRoute(&server.Route{
    Method: server.GET,
    Path:   "/search",
    Handler: func(ctx *server.Context) error {
        query := ctx.QueryParams["q"]
        page := ctx.QueryParams["page"]

        return server.SendJSON(ctx, 200, map[string]string{
            "query": query,
            "page":  page,
        })
    },
})
```

Test:
```bash
curl "http://localhost:8080/search?q=golang&page=1"
```

### Add Middleware

```go
srv := server.NewServer(
    server.WithMiddleware(server.LoggingMiddleware()),
    server.WithMiddleware(server.RecoveryMiddleware()),
)
```

## Testing

### Run Unit Tests

```bash
cd pkg/server
go test -v
```

### Run Benchmarks

```bash
go test -bench=. -benchmem
```

### Test Coverage

```bash
go test -cover
```

## Common Issues

### Port Already in Use

Change the port:
```go
srv.Start(":8081")  // Use different port
```

### Go Command Not Found

Install Go from https://golang.org/dl/

### Import Errors

Make sure you're in the correct directory and run:
```bash
go mod tidy
```

## Next Steps

1. Read the full documentation: `pkg/server/README.md`
2. Check out curl examples: `examples/server-demo/CURL_EXAMPLES.md`
3. Read the implementation report: `pkg/server/IMPLEMENTATION_REPORT.md`
4. Explore the test suite: `pkg/server/server_test.go`

## Getting Help

- Check `pkg/server/README.md` for detailed documentation
- Look at `examples/server-demo/main.go` for complete examples
- Review `pkg/server/server_test.go` for usage patterns
- See `examples/server-demo/CURL_EXAMPLES.md` for curl commands

## Summary

You now have a working HTTP server that can:
- Handle all HTTP methods (GET, POST, PUT, DELETE, PATCH)
- Extract path parameters
- Parse query parameters
- Handle JSON request/response
- Execute middleware chains
- Handle errors gracefully

Happy coding!
