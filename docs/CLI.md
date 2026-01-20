# Glyph CLI Documentation

The Glyph command-line interface (CLI) provides tools for developing, running, and deploying Glyph applications.

## Installation

```bash
# Build from source
make build

# Or install to your system
make install
```

## Commands

### `glyph dev <file>`

Start a development server with hot reload and live browser refresh.

```bash
glyph dev examples/hello-world/main.glyph

# Options:
#   -p, --port <port>     Port to listen on (default: 3000)
#   -w, --watch <bool>    Watch for file changes (default: true)
#   -o, --open            Open browser automatically
```

**Features:**
- Starts HTTP server on specified port
- Watches source file for changes with debounce (100ms)
- Hot reload with automatic server restart on file save
- Live reload via Server-Sent Events (SSE) at `/__livereload`
- JavaScript injection endpoint at `/__livereload.js`
- Browser auto-open with `--open` flag
- Falls back to interpreter mode if compilation fails
- Pretty colored output for requests and errors
- Graceful shutdown with Ctrl+C

**Example:**
```bash
$ glyph dev examples/hello-world/main.glyph -p 8080 --open
[INFO] Starting development server on port 8080...
[SUCCESS] Dev server listening on http://localhost:8080 (compiled mode)
[INFO] Live reload enabled at /__livereload
[INFO] Watching examples/hello-world/main.glyph for changes...
[INFO] Opened http://localhost:8080 in browser
[INFO] Press Ctrl+C to stop
```

**Live Reload Integration:**

Add the live reload script to your HTML for automatic browser refresh:
```html
<script src="/__livereload.js"></script>
```

When you save changes, you'll see:
```
[WARNING] File changed, reloading...
[SUCCESS] Hot reload complete (45ms)
```

### `glyph run <file>`

Run a Glyph source file or bytecode (production mode).

```bash
glyph run examples/rest-api/main.glyph

# Options:
#   -p, --port <port>     Port to listen on (default: 3000)
#   --bytecode            Execute bytecode (.glyphc) file directly
#   --interpret           Use tree-walking interpreter instead of compiler
```

**Features:**
- Compiles and runs Glyph source using VM (default)
- Falls back to interpreter if compilation fails
- Supports direct bytecode execution with `--bytecode`
- Starts HTTP server
- Request logging
- Graceful shutdown

**Example:**
```bash
# Run source file (compiles to bytecode first)
$ glyph run examples/rest-api/main.glyph
[INFO] Compiling and running examples/rest-api/main.glyph...
[SUCCESS] Server listening on http://localhost:3000 (compiled mode)
[INFO] Press Ctrl+C to stop

# Run pre-compiled bytecode
$ glyph run build/app.glyphc --bytecode
[INFO] Running bytecode build/app.glyphc...
[SUCCESS] Bytecode executed successfully

# Force interpreter mode
$ glyph run examples/rest-api/main.glyph --interpret
[INFO] Running examples/rest-api/main.glyph with interpreter...
[SUCCESS] Server listening on http://localhost:3000
```

### `glyph compile <file>`

Compile Glyph source code to bytecode.

```bash
glyph compile examples/hello-world/main.glyph

# Options:
#   -o, --output <file>      Output file (default: source.glyphc)
#   -O, --opt-level <0-3>    Optimization level (default: 2)
```

**Features:**
- Compiles source to optimized bytecode
- Multiple optimization levels
- Custom output path

**Example:**
```bash
$ glyph compile examples/hello-world/main.glyph -o build/hello.glyphc -O 3
[INFO] Compiling examples/hello-world/main.glyph (optimization level: 3)...
[SUCCESS] Compiled successfully to build/hello.glyphc (8 bytes)
```

### `glyph decompile <file>`

Decompile bytecode back to readable format.

```bash
glyph decompile build/hello.glyphc

# Options:
#   -o, --output <file>   Output file (default: source.glyph)
#   -d, --disasm          Output disassembly only (no file generation)
```

**Features:**
- Parses GLYP bytecode format
- Extracts constant pool (null, int, float, bool, string)
- Decodes all 37 VM opcodes
- Generates pseudo-source reconstruction
- Formatted disassembly with comments
- Supports all WebSocket opcodes

**Example:**
```bash
# Decompile to .glyph file with disassembly output
$ glyph decompile build/hello.glyphc
[INFO] Decompiling build/hello.glyphc...
[INFO] Bytecode version: 1
[INFO] Constants: 7
[INFO] Instructions: 7
[SUCCESS] Decompiled to build/hello.glyph

GlyphLang Bytecode v1
==================================================

CONSTANT POOL:
------------------------------
  [  0] string   "text"
  [  1] string   "Hello, World!"

INSTRUCTIONS:
------------------------------
  0000: PUSH               0      ; {text}
  0005: PUSH               1      ; {Hello, World!}
  0010: BUILD_OBJECT       1      ; 1 fields
  0015: RETURN
  0016: HALT

# Show only disassembly (no file output)
$ glyph decompile --disasm build/hello.glyphc
```

### `glyph exec <file> <command> [args...]`

Execute a CLI command defined in a Glyph source file.

```bash
glyph exec main.glyph hello --name="Alice"

# Arguments and flags:
#   <file>        The Glyph source file containing commands
#   <command>     The name of the command to execute
#   [args...]     Command-specific arguments and flags
```

**Features:**
- Execute CLI commands defined with `!` symbol
- Pass arguments and optional flags
- Returns JSON output from command
- Validates required arguments

**Example:**
```bash
# Execute simple command
$ glyph exec examples/cli-demo/main.glyph hello --name="Alice"
{
  "message": "Hello, Alice!"
}

# Execute command with multiple arguments
$ glyph exec examples/cli-demo/main.glyph add --a=5 --b=3
{
  "sum": 8,
  "a": 5,
  "b": 3
}

# Execute command with optional flags
$ glyph exec examples/cli-demo/main.glyph greet --name="Bob" --formal
{
  "greeting": "Good day, Bob. How may I assist you?"
}
```

**Command Definition:**
```glyph
! hello name: str! {
  $ greeting = "Hello, " + name + "!"
  > {message: greeting}
}
```

### `glyph commands <file>`

List all available CLI commands defined in a Glyph source file.

```bash
glyph commands main.glyph
```

**Features:**
- Shows all commands defined with `!` symbol
- Displays command names and descriptions
- Lists required and optional parameters
- Helps discover available commands

**Example:**
```bash
$ glyph commands examples/cli-demo/main.glyph
Available commands in examples/cli-demo/main.glyph:

  hello         Execute: glyph exec main.glyph hello
                Arguments:
                  name: str! (required)

  add           Execute: glyph exec main.glyph add
                Arguments:
                  a: int! (required)
                  b: int! (required)

  greet         Execute: glyph exec main.glyph greet
                Arguments:
                  name: str! (required)
                  --formal: bool (optional, default: false)

  list_users    Execute: glyph exec main.glyph list_users
                No arguments required

  version       Show version information
                No arguments required
```

### `glyph context [directory]`

Generate AI-optimized project context for AI agents working with Glyph codebases.

```bash
glyph context              # Generate context for current directory
glyph context ./src        # Generate for specific directory

# Options:
#   -f, --format <type>     Output format: json, compact, stubs (default: json)
#   -o, --output <file>     Output file (default: stdout)
#   --file <file>           Generate context for single file only
#   --pretty                Pretty-print JSON output (default: true)
#   --changed               Show only changes since last context generation
#   --save                  Save context to .glyph/context.json for future diffing
#   --for <type>            Generate targeted context: route, type, function, command
```

**Features:**
- Compact representations of types, routes, functions, commands
- Hash-based caching for change detection
- Multiple output formats for token efficiency
- Incremental updates with `--changed` flag
- Pattern detection (CRUD, auth usage, database routes)

**Example:**
```bash
# Generate compact context for AI agent
$ glyph context --format compact
Types: User(id:int!, name:str!, email:str?)
Routes: GET /api/users/:id -> User | Error [auth:jwt]
Patterns: crud, auth_usage, database_routes

# Save context and show changes on next run
$ glyph context --save
$ glyph context --changed
Changed: User.email type modified
```

### `glyph validate <file>`

Validate Glyph source files with structured, AI-friendly error output.

```bash
glyph validate main.glyph       # Human-readable output
glyph validate main.glyph --ai  # Structured JSON for AI agents
glyph validate src/ --ai        # Validate all files in directory

# Options:
#   --ai              Output structured JSON for AI agents
#   --strict          Treat warnings as errors
#   --quiet           Only output errors (no success messages)
```

**Features:**
- Structured JSON output with error types
- Precise locations (file, line, column)
- Fix hints for each error
- Semantic validation (type references, duplicate routes)
- Directory validation support

**Error Types:**
- `syntax_error` - Parser syntax errors
- `lexer_error` - Tokenization errors
- `undefined_reference` - Reference to undefined type/variable
- `type_mismatch` - Type incompatibility
- `duplicate_definition` - Duplicate type/route definition
- `invalid_route` - Invalid route configuration
- `missing_required` - Missing required field or parameter

**Example:**
```bash
# Validate with AI-friendly output
$ glyph validate main.glyph --ai
{
  "valid": false,
  "file_path": "main.glyph",
  "errors": [
    {
      "type": "undefined_reference",
      "message": "undefined type: UserProfile",
      "location": {"file": "main.glyph", "line": 15, "column": 12},
      "fix_hint": "define type 'UserProfile' or import it",
      "severity": "error"
    }
  ],
  "stats": {"types": 3, "routes": 5, "functions": 2, "lines": 87}
}
```

### `glyph init <name>`

Initialize a new Glyph project.

```bash
glyph init my-project

# Options:
#   -t, --template <name>    Project template (default: rest-api)
#                           Available: hello-world, rest-api
```

**Features:**
- Creates project directory
- Generates main.glyph with template
- Ready to run with `glyph dev`

**Example:**
```bash
$ glyph init my-api -t rest-api
[INFO] Creating project: my-api
[INFO] Template: rest-api
[SUCCESS] Project created successfully in my-api/
[INFO] Run: cd my-api && glyph dev main.glyph
```

### `glyph --version`

Display Glyph version.

```bash
$ glyph --version
Glyph version 0.3.2
```

### `glyph --help`

Display help information.

```bash
$ glyph --help
Glyph is a programming language specifically designed for AI agents
to rapidly build high-performance, secure backend applications.

Usage:
  glyph [command]

Available Commands:
  compile     Compile source code to bytecode
  context     Generate AI-optimized project context
  decompile   Decompile bytecode to readable format
  dev         Start development server with hot reload
  init        Initialize new project
  run         Run Glyph source file or bytecode
  exec        Execute a CLI command from a Glyph file
  commands    List all CLI commands in a Glyph file
  validate    Validate source with structured errors
  lsp         Start Language Server Protocol server
  help        Help about any command

Flags:
  -h, --help      help for glyph
  -v, --version   version for glyph

Use "glyph [command] --help" for more information about a command.
```

## Features

### Pretty Output

The CLI uses colored output for better readability:

- **[INFO]** - Cyan - General information
- **[SUCCESS]** - Green - Successful operations
- **[WARNING]** - Yellow - Warnings
- **[ERROR]** - Red - Errors
- **[GET/POST/etc]** - Magenta - HTTP requests

### Request Logging

All HTTP requests are logged with method, path, and duration:

```
[GET] /hello (234µs)
[POST] /api/users (1.2ms)
[GET] /api/users/123 (456µs)
```

### Graceful Shutdown

Servers handle Ctrl+C gracefully:

```
^C
[WARNING] Shutting down server...
[SUCCESS] Server stopped gracefully
```

### File Watching and Hot Reload

In dev mode, the CLI watches your source file and automatically restarts the server:

```
[WARNING] File changed, reloading...
[SUCCESS] Hot reload complete (45ms)
```

Browsers connected via the live reload endpoint will automatically refresh.

## Integration Status

### Complete
- CLI command structure (compile, decompile, run, dev, init, exec, commands, context, validate, lsp)
- Go parser with full lexer and AST generation
- Bytecode compiler with 3 optimization levels
- Bytecode VM execution
- Bytecode decompiler with full opcode support
- HTTP server with route registration
- Hot reload with live browser refresh (SSE)
- File watching with debounce
- WebSocket support
- Request logging
- Graceful shutdown
- Pretty error messages
- Project initialization

### Available Features
- Database connections (PostgreSQL)
- Middleware execution
- JWT authentication
- Rate limiting
- Caching (LRU, HTTP)
- Metrics and tracing
- LSP server for IDE integration

## Testing

### Unit Tests

```bash
# Run CLI unit tests
go test ./cmd/glyph/...

# Run with coverage
go test -cover ./cmd/glyph/...
```

### Integration Tests

```bash
# Run integration test script
chmod +x test_cli.sh
./test_cli.sh
```

The integration tests verify:
1. CLI builds successfully
2. Version command works
3. Help command works
4. Init command creates projects
5. Compile command generates bytecode
6. Run command starts server
7. Dev command starts server with watching

## Example Workflows

### Quick Start

```bash
# Create a new project
glyph init my-api -t hello-world

# Start development
cd my-api
glyph dev main.glyph

# In another terminal, test it
curl http://localhost:3000/hello
```

### Development Workflow

```bash
# Start dev server with hot reload
glyph dev main.glyph -p 8080

# Edit main.glyph in your editor
# Changes are automatically detected

# Test your API
curl http://localhost:8080/api/users
```

### CLI Commands Workflow

```bash
# List available commands in a file
glyph commands examples/cli-demo/main.glyph

# Execute a command
glyph exec examples/cli-demo/main.glyph hello --name="Alice"

# Execute with multiple arguments
glyph exec examples/cli-demo/main.glyph add --a=10 --b=20

# Execute with optional flags
glyph exec examples/cli-demo/main.glyph greet --name="Bob" --formal
```

### Production Build

```bash
# Compile to bytecode
glyph compile main.glyph -o build/app.glyphc -O 3

# Run in production
glyph run build/app.glyphc -p 80
```

## Troubleshooting

### Port Already in Use

```bash
# Use a different port
glyph dev main.glyph -p 3001
```

### File Not Found

```bash
# Use absolute path
glyph dev /path/to/main.glyph

# Or relative path from current directory
glyph dev examples/hello-world/main.glyph
```

### Permission Denied

```bash
# Use higher port number (>1024) or run with sudo
glyph dev main.glyph -p 8080
```

## Architecture

The CLI orchestrates four main components:

1. **Parser (Go)** - Lexical analysis and AST generation
2. **VM (Go)** - Bytecode execution with JIT optimization
3. **Interpreter (Go)** - AST execution and runtime (fallback)
4. **Server (Go)** - HTTP routing and middleware

```
┌─────────────────┐
│   Glyph CLI     │
│   (main.go)     │
└────────┬────────┘
         │
    ┌────┴────┬──────────┬──────────┐
    │         │          │          │
┌───▼───┐ ┌──▼──┐ ┌─────▼────┐ ┌──▼──┐
│Parser │ │ VM  │ │Interpreter│ │Server│
│ (Go)  │ │(Go) │ │   (Go)    │ │(Go)  │
└───────┘ └─────┘ └──────────┘ └─────┘
```

## LSP Command

### `glyph lsp`

Start the Language Server Protocol server for IDE integration.

```bash
glyph lsp

# Options:
#   -l, --log <file>    Log file for debugging (optional)
```

**Features:**
- Full LSP protocol support
- Syntax highlighting
- Diagnostics and error reporting
- Hover information
- Go to definition
- Code completion

**VS Code Extension:**

The [VS Code extension](https://github.com/GlyphLang/vscode-glyph) automatically starts the LSP server.

## JIT Compilation

Glyph includes a JIT (Just-In-Time) compiler that automatically optimizes frequently executed routes for better performance.

### How It Works

The JIT compiler uses a tiered compilation strategy:

| Tier | Description | When Applied |
|------|-------------|--------------|
| **Baseline** | Basic compilation, no optimization | First execution |
| **Optimized** | Standard optimizations | After ~50 executions |
| **Highly Optimized** | Aggressive optimizations | After ~100 executions (hot paths) |

### Automatic Optimization

JIT compilation is **enabled by default** in production mode. The runtime automatically:

1. Profiles route execution frequency
2. Identifies "hot paths" (frequently executed routes)
3. Recompiles hot paths with increasing optimization levels
4. Performs type specialization for monomorphic code

### Configuration

JIT behavior can be tuned programmatically when embedding Glyph:

```go
import "github.com/glyphlang/glyph/pkg/jit"

// Create JIT compiler with custom settings
jitCompiler := jit.NewJITCompilerWithConfig(
    100,              // Hot path threshold (execution count)
    10 * time.Second, // Recompile window
)

// Adjust at runtime
jitCompiler.SetHotPathThreshold(50)
jitCompiler.SetRecompileWindow(5 * time.Second)
```

### Monitoring JIT Statistics

Access JIT statistics programmatically:

```go
stats := jitCompiler.GetDetailedStats()
// Returns compilation counts, cache hits/misses, specialization stats
```

### Type Specialization

The JIT compiler tracks runtime types and generates specialized code for monomorphic variables (variables that consistently use the same type). This eliminates type checks in hot paths.

### Best Practices

1. **Let it warm up**: JIT benefits appear after routes are executed multiple times
2. **Avoid polymorphic code in hot paths**: Consistent types enable better optimization
3. **Monitor statistics**: Use `GetDetailedStats()` to understand JIT behavior

---

## Observability

Glyph provides built-in observability through Prometheus metrics and OpenTelemetry tracing.

### Prometheus Metrics

Glyph exposes metrics at the `/metrics` endpoint in Prometheus format.

#### Enabling Metrics

```go
import "github.com/glyphlang/glyph/pkg/metrics"

// Create metrics with default configuration
m := metrics.NewMetrics(metrics.DefaultConfig())

// Expose metrics endpoint
http.Handle("/metrics", m.Handler())
```

#### Available Metrics

**HTTP Metrics:**

| Metric | Type | Description |
|--------|------|-------------|
| `glyphlang_http_requests_total` | Counter | Total HTTP requests (by method, path, status) |
| `glyphlang_http_request_duration_seconds` | Histogram | Request latency in seconds |
| `glyphlang_http_request_errors_total` | Counter | HTTP errors (status >= 400) |

**Runtime Metrics:**

| Metric | Type | Description |
|--------|------|-------------|
| `glyphlang_runtime_goroutines` | Gauge | Current goroutine count |
| `glyphlang_runtime_memory_alloc_bytes` | Gauge | Bytes currently allocated |
| `glyphlang_runtime_memory_sys_bytes` | Gauge | Bytes obtained from system |
| `glyphlang_runtime_gc_pause_ns` | Gauge | Most recent GC pause time |
| `glyphlang_runtime_gc_runs_total` | Gauge | Total GC runs |

#### Custom Configuration

```go
config := metrics.Config{
    Namespace: "myapp",
    Subsystem: "api",
    DurationBuckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
}
m := metrics.NewMetrics(config)
```

#### Middleware Integration

```go
import "github.com/glyphlang/glyph/pkg/metrics"

m := metrics.NewMetrics(metrics.DefaultConfig())
metricsMiddleware := metrics.MetricsMiddleware(m)

// Apply to routes
handler := metricsMiddleware(yourHandler)
```

#### Prometheus Configuration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'glyphlang'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### OpenTelemetry Tracing

Glyph supports distributed tracing via OpenTelemetry with W3C Trace Context propagation.

#### Enabling Tracing

```go
import "github.com/glyphlang/glyph/pkg/tracing"

// Initialize with default config (stdout exporter for development)
tp, err := tracing.InitTracing(tracing.DefaultConfig())
if err != nil {
    log.Fatal(err)
}
defer tp.Shutdown(context.Background())
```

#### Production Configuration

```go
config := &tracing.Config{
    ServiceName:    "my-glyph-app",
    ServiceVersion: "1.0.0",
    Environment:    "production",
    ExporterType:   "otlp",           // "stdout" for dev, "otlp" for production
    OTLPEndpoint:   "jaeger:4317",    // Your collector endpoint
    SamplingRate:   0.1,              // Sample 10% of traces
    Enabled:        true,
}
tp, err := tracing.InitTracing(config)
```

#### Environment Variables

OpenTelemetry SDK respects standard environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `OTEL_SERVICE_NAME` | Service name | `my-glyph-app` |
| `OTEL_SDK_DISABLED` | Disable tracing | `true` |
| `OTEL_TRACES_EXPORTER` | Exporter type | `otlp`, `stdout` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP endpoint | `localhost:4317` |
| `OTEL_TRACES_SAMPLER` | Sampler type | `always_on`, `traceidratio` |
| `OTEL_TRACES_SAMPLER_ARG` | Sampler argument | `0.1` (for 10% sampling) |

#### HTTP Middleware

```go
import "github.com/glyphlang/glyph/pkg/tracing"

config := &tracing.MiddlewareConfig{
    SpanNameFormatter: func(req *http.Request) string {
        return fmt.Sprintf("HTTP %s %s", req.Method, req.URL.Path)
    },
    ExcludePaths: map[string]bool{
        "/health":  true,
        "/metrics": true,
    },
}

middleware := tracing.HTTPTracingMiddleware(config)
tracedHandler := middleware(yourHandler)
```

#### Trace Context Headers

Traced responses include debug headers:
- `X-Trace-ID`: The trace ID for correlation
- `X-Span-ID`: The span ID

#### Deployment with Jaeger

```bash
# Start Jaeger
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  jaegertracing/all-in-one:latest

# Access UI at http://localhost:16686
```

---

## Environment Variables

Glyph applications can be configured via environment variables.

### Core Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `Glyph_ENV` | Environment (development/production) | `production` |
| `Glyph_PORT` | HTTP server port | `8080` |
| `Glyph_LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |

### Database Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost:5432/db` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379` |

### Observability Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OTEL_SDK_DISABLED` | Disable OpenTelemetry | `false` |
| `OTEL_SERVICE_NAME` | Service name for tracing | `glyphlang` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP collector endpoint | `localhost:4317` |

### Example `.env` File

```bash
# Application
Glyph_ENV=production
Glyph_PORT=8080
Glyph_LOG_LEVEL=info

# Database
DATABASE_URL=postgres://glyph:secret@localhost:5432/glyphdb

# Observability
OTEL_SERVICE_NAME=my-glyph-api
OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
```

### Loading Environment Variables

```bash
# Using .env file (requires dotenv or similar)
source .env && glyph run main.glyph

# Or export directly
export Glyph_PORT=3000
export DATABASE_URL="postgres://..."
glyph run main.glyph
```

---

## Contributing

When adding new CLI commands:

1. Add the command definition in `main()`
2. Create a `run*` handler function
3. Implement the command logic
4. Add tests in `main_test.go`
5. Update this documentation
6. Add integration test in `test_cli.sh`
