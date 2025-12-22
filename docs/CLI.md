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
#   --bytecode            Execute bytecode (.glybc) file directly
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
$ glyph run build/app.glybc --bytecode
[INFO] Running bytecode build/app.glybc...
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
#   -o, --output <file>      Output file (default: source.glybc)
#   -O, --opt-level <0-3>    Optimization level (default: 2)
```

**Features:**
- Compiles source to optimized bytecode
- Multiple optimization levels
- Custom output path

**Example:**
```bash
$ glyph compile examples/hello-world/main.glyph -o build/hello.glybc -O 3
[INFO] Compiling examples/hello-world/main.glyph (optimization level: 3)...
[SUCCESS] Compiled successfully to build/hello.glybc (8 bytes)
```

### `glyph decompile <file>`

Decompile bytecode back to readable format.

```bash
glyph decompile build/hello.glybc

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
$ glyph decompile build/hello.glybc
[INFO] Decompiling build/hello.glybc...
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
$ glyph decompile --disasm build/hello.glybc
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
Glyph version 0.1.0-alpha
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
  decompile   Decompile bytecode to readable format
  dev         Start development server with hot reload
  init        Initialize new project
  run         Run Glyph source file or bytecode
  exec        Execute a CLI command from a Glyph file
  commands    List all CLI commands in a Glyph file
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
- CLI command structure (compile, decompile, run, dev, init, exec, commands, lsp)
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
glyph compile main.glyph -o build/app.glybc -O 3

# Run in production
glyph run build/app.glybc -p 80
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

The CLI orchestrates three main components:

1. **Parser (Rust)** - Lexical analysis and AST generation
2. **Interpreter (Go)** - AST execution and runtime
3. **Server (Go)** - HTTP routing and middleware

```
┌─────────────────┐
│   Glyph CLI      │
│   (main.go)     │
└────────┬────────┘
         │
    ┌────┴────┬──────────┬──────────┐
    │         │          │          │
┌───▼───┐ ┌──▼──┐ ┌─────▼────┐ ┌──▼──┐
│Parser │ │ VM  │ │Interpreter│ │Server│
│(Rust) │ │(Go) │ │  (Go)     │ │(Go) │
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

The VS Code extension (`vscode-glyph`) automatically starts the LSP server.

## Contributing

When adding new CLI commands:

1. Add the command definition in `main()`
2. Create a `run*` handler function
3. Implement the command logic
4. Add tests in `main_test.go`
5. Update this documentation
6. Add integration test in `test_cli.sh`
