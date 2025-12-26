# GlyphLang

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)]()
[![Tests](https://img.shields.io/badge/tests-637%2B%20passing-success)]()
[![Coverage](https://img.shields.io/badge/coverage-80%25%2B-green)]()
[![Version](https://img.shields.io/badge/version-v1.0.0-blue)](https://github.com/GlyphLang/GlyphLang/releases/tag/v1.0.0)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)

**GlyphLang** is a domain-specific language for building type-safe REST APIs with bytecode compilation and JIT optimization. Symbol-based syntax reduces token usage by 45% compared to Python, making it ideal for AI code generation.

```glyph
@ route /hello/:name [GET]
  $ greeting = "Hello, " + name + "!"
  > {message: greeting}
```

## Features

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
- **VS Code Extension** - syntax highlighting, error checking
- **CLI** - compile, run, REPL, decompile commands

## Installation

**Download binary:**
```bash
# Linux
curl -L https://github.com/GlyphLang/GlyphLang/releases/download/v1.0.0/glyph-linux-amd64.zip -o glyph.zip
unzip glyph.zip && chmod +x glyph-linux-amd64 && sudo mv glyph-linux-amd64 /usr/local/bin/glyph

# macOS (Intel)
curl -L https://github.com/GlyphLang/GlyphLang/releases/download/v1.0.0/glyph-darwin-amd64.zip -o glyph.zip
unzip glyph.zip && chmod +x glyph-darwin-amd64 && sudo mv glyph-darwin-amd64 /usr/local/bin/glyph

# macOS (Apple Silicon)
curl -L https://github.com/GlyphLang/GlyphLang/releases/download/v1.0.0/glyph-darwin-arm64.zip -o glyph.zip
unzip glyph.zip && chmod +x glyph-darwin-arm64 && sudo mv glyph-darwin-arm64 /usr/local/bin/glyph
```

**Or build from source:**
```bash
git clone https://github.com/GlyphLang/GlyphLang.git
cd GlyphLang && go build -o glyph ./cmd/glyph
```

## Quick Start

Create `hello.glyph`:

```glyph
@ route /hello [GET]
  > {message: "Hello, World!"}

@ route /greet/:name [GET]
  > {message: "Hello, " + name + "!"}
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
@ route /api/users/:id [GET] -> User | Error
  + auth(jwt)
  + ratelimit(100/min)
  % db: Database

  $ user = db.query("SELECT * FROM users WHERE id = ?", id)
  if user == null {
    > {error: "User not found", code: 404}
  }
  > user
```

### Pattern Matching

```glyph
@ route /status/:code [GET]
  $ result = match code {
    200 => "OK"
    201 => "Created"
    400 => "Bad Request"
    404 => "Not Found"
    n when n >= 500 => "Server Error"
    _ => "Unknown"
  }
  > {status: code, message: result}
```

### Async/Await

```glyph
@ route /dashboard [GET]
  $ userFuture = async { db.getUser(userId) }
  $ ordersFuture = async { db.getOrders(userId) }
  $ statsFuture = async { db.getStats(userId) }

  $ user = await userFuture
  $ orders = await ordersFuture
  $ stats = await statsFuture

  > {user: user, orders: orders, stats: stats}
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

@ route /user/:id [GET]
  $ name = utils.formatName(user.first, user.last)
  > {displayName: name}
```

### Macros

```glyph
macro! crud(resource) {
  @ route /${resource} [GET]
    > db.query("SELECT * FROM ${resource}")

  @ route /${resource}/:id [GET]
    > db.query("SELECT * FROM ${resource} WHERE id = ?", id)

  @ route /${resource} [POST]
    > db.insert("${resource}", input)
}

crud!(users)
crud!(posts)
```

## CLI Commands

```bash
glyph run <file>        # Run a Glyph file
glyph compile <file>    # Compile to bytecode
glyph repl              # Interactive REPL
glyph lsp               # Start LSP server
glyph decompile <file>  # Decompile bytecode
glyph version           # Show version
```

## Documentation

- [Language Specification](docs/LANGUAGE_SPECIFICATION.md)
- [API Reference](docs/API_REFERENCE.md)
- [Quickstart Guide](docs/QUICKSTART.md)
- [CLI Reference](docs/CLI.md)
- [Architecture](docs/ARCHITECTURE_DESIGN.md)

## Performance

| Metric | Value |
|--------|-------|
| Compilation | ~867 ns |
| Execution | 2.95-37.6 ns/op |
| Test Coverage | 80%+ (14 packages) |
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
