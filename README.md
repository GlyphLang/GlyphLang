# GlyphLang

[![AI Token Savings](https://img.shields.io/badge/AI%20tokens-45%25%20fewer%20than%20Python-blueviolet)]()
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)]()
[![Tests](https://img.shields.io/badge/tests-3503%20passing-success)]()
[![Coverage](https://img.shields.io/badge/coverage-80%25%2B-green)]()
[![Version](https://img.shields.io/badge/version-v0.2.0-blue)](https://github.com/GlyphLang/GlyphLang/releases/latest)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)
[![CLA](https://cla-assistant.io/readme/badge/GlyphLang/GlyphLang)](https://cla-assistant.io/GlyphLang/GlyphLang)

**GlyphLang** is the AI-first language for REST APIs. Symbol-based syntax uses 45% fewer tokens than Python and 63% fewer than Java, making LLM code generation faster, cheaper, and more reliable.

```
Glyph:  @ GET /users/:id -> User     (17 tokens)
Python: @app.route('/users/<id>')... (60 tokens)
Java:   @GetMapping("/users/{id}")...  (74 tokens)
```

**Why AI-first matters:**
- Faster generation (fewer tokens = faster LLM response)
- Lower cost (55% savings on API calls at scale)
- Fewer errors (less code = less hallucination)
- More context (fit more business logic in context windows)

```glyph
@ GET /hello/:name {
  $ greeting = "Hello, " + name + "!"
  > {message: greeting}
}
```

## Features

### AI-First Design
- **Token-Optimized Syntax** - Symbols (`@`, `$`, `>`, `:`) instead of keywords
- **Context Command** - Generate project summaries optimized for LLM context windows
- **Structured Validation** - JSON error output for AI agents to parse and fix
- **Consistent Patterns** - Predictable syntax reduces hallucination

### Core Language
- **Type System** - int, string, bool, float, arrays, objects, optional (`T?`), union (`A | B`), generics
- **Pattern Matching** - match expressions with guards and destructuring
- **Async/Await** - futures with `All`, `Race`, `Any` combinators
- **Modules** - file imports, aliases, selective imports
- **Generics** - type parameters, inference, constraints
- **Macros** - compile-time code generation

### Runtime
- **Bytecode Compiler** - 3 optimization levels
- **JIT Compilation** - type specialization, hot path detection
- **Hot Reload** - instant updates during development
- **Debug Mode** - breakpoints, variable inspection, REPL

### Infrastructure
- **HTTP Server** - routes, middleware, WebSocket support
- **Database** - PostgreSQL with pooling, transactions, migrations
- **Security** - JWT auth, rate limiting, CORS, SQL injection prevention
- **Observability** - logging, Prometheus metrics, OpenTelemetry tracing

### Tooling
- **LSP Server** - full IDE support with completions, diagnostics, rename
- **[VS Code Extension](https://github.com/GlyphLang/vscode-glyph)** - syntax highlighting, error checking
- **CLI** - compile, run, REPL, decompile, AI context commands

## Symbol Reference

GlyphLang uses symbols instead of keywords for token efficiency:

| Symbol | Name | Usage | Example |
|--------|------|-------|---------|
| `@` | Route | HTTP endpoint definition | `@ GET /users` |
| `:` | Type | Type definition | `: User { id: int }` |
| `$` | Variable | Variable declaration | `$ name = "Alice"` |
| `!` | Function | Function/CLI command definition | `! greet(name: string)` |
| `>` | Return | Return statement | `> {message: "ok"}` |
| `+` | Middleware | Apply middleware | `+ auth(jwt)` |
| `%` | Inject | Dependency injection | `% db: Database` |
| `?` | Optional | Optional type modifier | `email: string?` |
| `*` | Cron | Scheduled task definition | `* "0 * * * *" cleanup` |
| `~` | Event | Event handler definition | `~ user.created` |
| `&` | Queue | Queue worker definition | `& emails processEmail` |
| `#` | Comment | Single-line comment | `# This is a comment` |
| `->` | Arrow | Return type annotation | `-> User` |
| `\|` | Union | Union type separator | `string \| int` |

**Type Modifiers:**
- `T!` - Required (non-null)
- `T?` - Optional (nullable)
- `[T]` - Array of T

## Installation

**Windows Installer (Recommended):**

Download and run the installer: [glyph-windows-setup.exe](https://github.com/GlyphLang/GlyphLang/releases/latest/download/glyph-windows-setup.exe)

**Download binary:**
```bash
# Linux
curl -L https://github.com/GlyphLang/GlyphLang/releases/latest/download/glyph-linux-amd64.zip -o glyph.zip
unzip glyph.zip && chmod +x glyph-linux-amd64 && sudo mv glyph-linux-amd64 /usr/local/bin/glyph

# macOS (Intel)
curl -L https://github.com/GlyphLang/GlyphLang/releases/latest/download/glyph-darwin-amd64.zip -o glyph.zip
unzip glyph.zip && chmod +x glyph-darwin-amd64 && sudo mv glyph-darwin-amd64 /usr/local/bin/glyph

# macOS (Apple Silicon)
curl -L https://github.com/GlyphLang/GlyphLang/releases/latest/download/glyph-darwin-arm64.zip -o glyph.zip
unzip glyph.zip && chmod +x glyph-darwin-arm64 && sudo mv glyph-darwin-arm64 /usr/local/bin/glyph
```

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri "https://github.com/GlyphLang/GlyphLang/releases/latest/download/glyph-windows-amd64.zip" -OutFile glyph.zip
Expand-Archive glyph.zip -DestinationPath . ; Move-Item glyph-windows-amd64.exe glyph.exe
```

**Or build from source:**
```bash
git clone https://github.com/GlyphLang/GlyphLang.git
cd GlyphLang && go build -o glyph ./cmd/glyph
```

## Quick Start

Create `hello.glyph`:

```glyph
@ GET /hello {
  > {message: "Hello, World!"}
}

@ GET /greet/:name {
  > {message: "Hello, " + name + "!"}
}
```

Run it:

```bash
glyph run hello.glyph
```

Visit http://localhost:3000/hello

## Examples

### Type Definitions

```glyph
: User {
  id: int!
  name: string!
  email: string?
  roles: [string]!
}

: ApiResponse<T> {
  data: T?
  error: string?
  success: bool!
}
```

### Routes with Authentication

```glyph
@ GET /api/users/:id -> User | Error {
  + auth(jwt)
  + ratelimit(100/min)
  % db: Database

  $ user = db.query("SELECT * FROM users WHERE id = ?", id)
  if user == null {
    > {error: "User not found", code: 404}
  }
  > user
}
```

### Pattern Matching

```glyph
@ GET /status/:code {
  $ result = match code {
    200 => "OK"
    201 => "Created"
    400 => "Bad Request"
    404 => "Not Found"
    n when n >= 500 => "Server Error"
    _ => "Unknown"
  }
  > {status: code, message: result}
}
```

### Async/Await

```glyph
# Basic async block - executes in background, returns Future
@ GET /compute {
  $ future = async {
    $ x = 10
    $ y = 20
    > x + y
  }
  $ result = await future
  > {value: result}
}

# Parallel execution - all requests run concurrently
@ GET /dashboard {
  $ userFuture = async { > db.getUser(userId) }
  $ ordersFuture = async { > db.getOrders(userId) }
  $ statsFuture = async { > db.getStats(userId) }

  # Await blocks until Future resolves
  $ user = await userFuture
  $ orders = await ordersFuture
  $ stats = await statsFuture

  > {user: user, orders: orders, stats: stats}
}
```

### Generics

```glyph
! identity<T>(x: T): T {
  > x
}

! first<T>(a: T, b: T): T {
  > a
}

! map<T, U>(arr: [T], fn: (T) -> U): [U] {
  $ result = []
  for item in arr {
    $ mapped = fn(item)
    result = append(result, mapped)
  }
  > result
}
```

### Modules

```glyph
# utils.glyph
! formatName(first: string, last: string): string {
  > first + " " + last
}

# main.glyph
import "./utils"

@ GET /user/:id {
  $ name = utils.formatName(user.first, user.last)
  > {displayName: name}
}
```

### Macros

```glyph
macro! crud(resource) {
  @ GET /${resource} {
    > db.query("SELECT * FROM ${resource}")
  }

  @ GET /${resource}/:id {
    > db.query("SELECT * FROM ${resource} WHERE id = ?", id)
  }

  @ POST /${resource} {
    > db.insert("${resource}", input)
  }
}

crud!(users)
crud!(posts)
```

## CLI Commands

### AI Agent Commands

These commands are designed for AI coding assistants and agents:

```bash
# Generate context for AI agents (fits in context windows)
glyph context                     # Full project context as JSON
glyph context --format compact    # Minimal text (fewer tokens)
glyph context --changed           # Only changes since last run
glyph context --for route         # Focus on routes only
glyph context --for type          # Focus on type definitions

# Validate with structured output for AI to parse and fix
glyph validate main.glyph --ai    # JSON errors with fix hints
glyph validate src/ --ai          # Validate entire directory
```

Example AI workflow:
```bash
# 1. Agent gets project context
glyph context --format compact > context.txt

# 2. Agent makes changes, then validates
glyph validate src/ --ai | agent-fix-errors

# 3. Agent checks what changed
glyph context --changed
```

### Standard Commands

```bash
glyph run <file>        # Run a Glyph file
glyph dev <file>        # Development server with hot reload
glyph compile <file>    # Compile to bytecode
glyph decompile <file>  # Decompile bytecode
glyph lsp               # Start LSP server
glyph init              # Initialize new project
glyph commands <file>   # List CLI commands in a file
glyph exec <file> <cmd> # Execute a CLI command
glyph version           # Show version
```

## Documentation

- [Language Specification](docs/LANGUAGE_SPECIFICATION.md)
- [API Reference](docs/API_REFERENCE.md)
- [Quickstart Guide](docs/QUICKSTART.md)
- [CLI Reference](docs/CLI.md)
- [Architecture](docs/ARCHITECTURE_DESIGN.md)

## Performance

### AI Token Efficiency

| Comparison | Glyph Tokens | Other Tokens | Savings |
|------------|--------------|--------------|---------|
| vs Python | 463 | 842 | 45% fewer |
| vs Java | 463 | 1252 | 63% fewer |

| Use Case | Glyph | Python | Java |
|----------|-------|--------|------|
| Hello World API | 17 | 60 | 74 |
| CRUD API | 172 | 273 | 285 |
| WebSocket Handler | 111 | 203 | 396 |
| Type Definition | 31 | 55 | 235 |

*Benchmark: 7 equivalent API implementations. Run `python benchmarks/bench_ai_efficiency.py` to reproduce.*

### Runtime Performance

| Metric | Value |
|--------|-------|
| Compilation | ~867 ns |
| Execution | 2.95-37.6 ns/op |
| Test Coverage | 80%+ (23 packages) |
| Examples | 100% compatibility |

## Project Structure

```
GlyphLang/
├── cmd/glyph/           # CLI application
├── pkg/
│   ├── parser/          # Lexer and parser
│   ├── interpreter/     # AST interpreter
│   ├── compiler/        # Bytecode compiler
│   ├── vm/              # Virtual machine
│   ├── jit/             # JIT compiler
│   ├── server/          # HTTP server
│   ├── database/        # Database integration
│   ├── security/        # Auth, rate limiting
│   ├── lsp/             # Language server
│   └── ...              # Other packages
├── examples/            # Example projects
├── docs/                # Documentation
└── tests/               # Integration tests
```

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

By contributing, you agree to the [Contributor License Agreement](CONTRIBUTING.md#contributor-license-agreement-cla).

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
