# Glyph CLI - Deployment & Testing Guide

## Prerequisites

Before building and testing the Glyph CLI, ensure you have:

1. **Go 1.21+** installed
   ```bash
   # Check Go version
   go version

   # If not installed, download from https://golang.org/dl/
   ```

2. **Git** for dependency management
   ```bash
   git --version
   ```

3. **Make** (optional, for convenience)
   - Linux/Mac: Usually pre-installed
   - Windows: Install via chocolatey (`choco install make`) or use direct go commands

## Build Instructions

### Method 1: Using Make (Recommended)

```bash
# Navigate to project root (if not already there)
cd /path/to/glyph

# Download Go dependencies
go mod download

# Build the CLI
make build-go

# The binary will be at: build/glyph (or build/glyph.exe on Windows)
```

### Method 2: Direct Go Commands

```bash
# Navigate to project root (if not already there)
cd /path/to/glyph

# Download dependencies
go mod download

# Build CLI
go build -o build/glyph ./cmd/glyph

# On Windows, it will automatically create build/glyph.exe
```

### Method 3: Install to System

```bash
# Install to $GOPATH/bin (adds to your PATH)
go install ./cmd/glyph

# Now you can run 'glyph' from anywhere
glyph --version
```

## Verify Build

After building, verify the CLI works:

```bash
# Check version
./build/glyph --version

# Output should be:
# Glyph version 0.1.0-alpha

# Check help
./build/glyph --help

# Should display available commands
```

## Running Tests

### Unit Tests

```bash
# Test CLI package
go test ./cmd/glyph/...

# Test with verbose output
go test -v ./cmd/glyph/...

# Test with coverage
go test -cover ./cmd/glyph/...

# Generate coverage report
go test -coverprofile=coverage.out ./cmd/glyph/...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Make script executable (Linux/Mac)
chmod +x test_cli.sh

# Run integration tests
./test_cli.sh

# Or run with bash explicitly
bash test_cli.sh
```

### Windows Testing

```powershell
# Build
go build -o build\glyph.exe .\cmd\glyph

# Test version
.\build\glyph.exe --version

# Test help
.\build\glyph.exe --help

# Test init
.\build\glyph.exe init test-project -t hello-world

# Test compile
.\build\glyph.exe compile examples\hello-world\main.glyph

# Test run (Ctrl+C to stop)
.\build\glyph.exe run examples\hello-world\main.glyph

# Test dev (Ctrl+C to stop)
.\build\glyph.exe dev examples\hello-world\main.glyph
```

## Quick Start Guide

### 1. Create a New Project

```bash
./build/glyph init my-first-api -t hello-world
cd my-first-api
```

This creates:
```
my-first-api/
└── main.glyph
```

### 2. Start Development Server

```bash
# From my-first-api directory
../build/glyph dev main.glyph

# Or specify port
../build/glyph dev main.glyph -p 8080
```

You should see:
```
[INFO] Starting development server on port 3000...
[INFO] Watching /path/to/main.glyph for changes...
[SUCCESS] Server listening on http://localhost:3000
[INFO] Press Ctrl+C to stop
```

### 3. Test the API

In another terminal:

```bash
# Test hello endpoint
curl http://localhost:3000/hello

# Should return:
# {"message":"Hello from Glyph!","path":"/hello","method":"GET"}
```

### 4. Edit and Watch Hot Reload

1. Edit `main.glyph` in your text editor
2. Save the file
3. Watch the terminal output:
   ```
   [WARNING] File changed, reloading...
   [INFO] Hot reload triggered (server restart not yet implemented)
   ```

### 5. Compile to Bytecode

```bash
../build/glyph compile main.glyph -o app.glybc -O 3

# Should output:
# [INFO] Compiling main.glyph (optimization level: 3)...
# [SUCCESS] Compiled successfully to app.glybc (8 bytes)
```

### 6. Run in Production Mode

```bash
../build/glyph run main.glyph -p 8080

# Should output:
# [INFO] Running main.glyph...
# [SUCCESS] Server listening on http://localhost:8080
# [INFO] Press Ctrl+C to stop
```

Press Ctrl+C to gracefully shutdown:
```
^C
[WARNING] Shutting down server...
[SUCCESS] Server stopped gracefully
```

## Testing with Example Files

### Hello World Example

```bash
./build/glyph dev examples/hello-world/main.glyph
```

Then test:
```bash
curl http://localhost:3000/hello
curl http://localhost:3000/greet/John
```

### REST API Example

```bash
./build/glyph dev examples/rest-api/main.glyph
```

Then test:
```bash
curl http://localhost:3000/health
curl http://localhost:3000/api/users
curl http://localhost:3000/api/users/123
```

## Troubleshooting

### Issue: "go: command not found"

**Solution:** Install Go from https://golang.org/dl/

### Issue: "package X is not in GOROOT"

**Solution:** Download dependencies:
```bash
go mod download
go mod tidy
```

### Issue: "Port already in use"

**Solution:** Use a different port:
```bash
./build/glyph dev main.glyph -p 3001
```

### Issue: "Permission denied" on Linux/Mac

**Solution:** Make the binary executable:
```bash
chmod +x build/glyph
```

### Issue: Cannot bind to port 80/443

**Solution:** Ports below 1024 require root privileges:
```bash
# Use sudo (not recommended for development)
sudo ./build/glyph run main.glyph -p 80

# Or use a higher port
./build/glyph run main.glyph -p 8080
```

### Issue: File watcher not working

**Solution:**
- Check file system supports inotify (Linux) or FSEvents (Mac)
- Disable watching: `./build/glyph dev main.glyph --watch=false`

### Issue: Build fails with "missing go.sum entry"

**Solution:**
```bash
go mod tidy
go mod download
```

## Performance Testing

### Load Testing with curl

```bash
# Start server
./build/glyph run examples/hello-world/main.glyph &
SERVER_PID=$!

# Simple load test
for i in {1..1000}; do
  curl -s http://localhost:3000/hello > /dev/null
done

# Kill server
kill $SERVER_PID
```

### Load Testing with Apache Bench

```bash
# Install ab (usually comes with apache2-utils)
# Debian/Ubuntu: apt-get install apache2-utils
# Mac: brew install apache2

# Test with 1000 requests, 10 concurrent
ab -n 1000 -c 10 http://localhost:3000/hello
```

### Load Testing with wrk

```bash
# Install wrk
# Mac: brew install wrk
# Linux: build from source

# Test with 10 threads, 100 connections for 30 seconds
wrk -t10 -c100 -d30s http://localhost:3000/hello
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Test Glyph CLI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Download dependencies
        run: go mod download

      - name: Build CLI
        run: go build -o build/glyph ./cmd/glyph

      - name: Run tests
        run: go test -v ./cmd/glyph/...

      - name: Run integration tests
        run: bash test_cli.sh
```

### Docker Example

```dockerfile
# Dockerfile
FROM golang:1.21 AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o glyph ./cmd/glyph

FROM debian:bookworm-slim
COPY --from=builder /app/glyph /usr/local/bin/
COPY --from=builder /app/examples /examples

EXPOSE 3000
CMD ["glyph", "run", "/examples/hello-world/main.glyph"]
```

Build and run:
```bash
docker build -t glyph .
docker run -p 3000:3000 glyph
```

## Development Workflow

### For CLI Development

1. Make changes to `cmd/glyph/main.go`
2. Run tests: `go test ./cmd/glyph/...`
3. Build: `go build ./cmd/glyph`
4. Test manually with examples
5. Commit changes

### Hot Reload During Development

```bash
# Install air for hot reload of Go code
go install github.com/cosmtrek/air@latest

# Run with air
air -c .air.toml
```

Create `.air.toml`:
```toml
[build]
  cmd = "go build -o ./tmp/glyph ./cmd/glyph"
  bin = "./tmp/glyph dev examples/hello-world/main.glyph"
```

## Next Steps

Once you have the CLI working:

1. **Test all commands** - Verify each command works as expected
2. **Check examples** - Make sure both example files work
3. **Integration testing** - Run the test_cli.sh script
4. **Performance testing** - Test with load tools
5. **Parser integration** - Wait for Rust parser FFI to be ready
6. **Interpreter integration** - Wait for full interpreter implementation
7. **Hot reload implementation** - Implement server restart on file change

## Expected Behavior

### Current Working Features ✅

- CLI builds successfully
- All commands parse correctly
- `init` creates new projects
- `compile` generates bytecode (stub)
- `run` starts HTTP server
- `dev` starts server with file watching
- Graceful shutdown with Ctrl+C
- Pretty colored output
- Request logging
- Error handling

### Pending Integration 🔄

- Actual Rust parser calls (currently using stub)
- Full interpreter execution (currently returns mock data)
- Hot reload server restart (currently just logs)
- Database connections
- Middleware execution
- Authentication/authorization

### Not Implemented Yet ⏳

- Bytecode VM execution
- Production optimizations
- Debug mode
- Profiling tools
- Deployment tools

## Support

If you encounter issues:

1. Check this documentation
2. Review the CLI.md for usage details
3. Check the logs for error messages
4. Verify Go version and dependencies
5. Try rebuilding: `go clean && go build ./cmd/glyph`
