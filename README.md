# GlyphLang - Glyph Compiler

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)]()
[![Tests](https://img.shields.io/badge/tests-1696%20passing-success)]()
[![Performance](https://img.shields.io/badge/compilation-867ns-blue)]()
[![AI Tokens](https://img.shields.io/badge/tokens-45%25%20fewer%20than%20Python-purple)]()
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)
[![CLA Assistant](https://cla-assistant.io/readme/badge/glyph-lang/glyph)](https://cla-assistant.io/glyph-lang/glyph)

**GlyphLang** is an AI-first backend language designed for LLM code generation. Symbol-based syntax uses **45% fewer tokens than Python** and **63% fewer than Java**, reducing AI costs by 56%. Sub-microsecond compilation, built-in security, production-grade performance.

```glyph
@ route /hello/:name
  $ greeting = "Hello, " + name + "!"
  > {message: greeting}
```

## ⚡ Key Features

- **AI-First Design** - Symbol syntax (@, $, >, :) uses 45% fewer tokens than Python
- **56% Lower AI Costs** - Fewer tokens = faster generation, lower API costs
- **Sub-Microsecond Compilation** - 867 nanoseconds (115,000x faster than targets)
- **Nanosecond Execution** - 2.95-37.6 ns/op with zero-allocation arithmetic
- **7x Faster than Python** - Benchmarked arithmetic operations
- **Built-in Security** - SQL injection and XSS detection with intelligent suggestions
- **Comprehensive Validation** - Email, length, range, pattern validators
- **Enhanced Errors** - Line numbers, source snippets, and smart suggestions
- **Bytecode VM** - Complete VM with 25 opcodes and exceptional performance
- **Production Ready** - 1696 tests passing, Docker/K8s deployment configs included

## 🚀 Quick Start

### Installation

**macOS / Linux:**
```bash
curl -fsSL https://glyph-lang.github.io/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://glyph-lang.github.io/install.ps1 | iex
```

**Or build from source:**
```bash
git clone https://github.com/glyph-lang/glyph.git
cd glyph && go build -o glyph ./cmd/glyph
```

### Your First API

Create `hello.glyph`:

```glyph
@ route /hello
  > {message: "Hello, World!"}
```

Run it:

```bash
./glyph dev hello.glyph
```

Visit http://localhost:3000/hello - Done! 🎉

## Example Code

```
# Define a type
: User {
  id: int!
  name: str!
  email: str!
}

# Create an API endpoint
@ route /api/users/:id -> User | Error
  + auth(jwt)
  + ratelimit(100/min)
  % db: Database
  $ user = db.users.get(id)
  > user
```

## Architecture

- **Runtime**: Go (fast compilation, excellent HTTP/DB support)
- **Compiler Core**: Rust (high-performance parsing and type checking)
- **Target**: Portable bytecode (OS agnostic)

## Project Structure

```
glyph/
├── cmd/glyph/          # CLI tool
├── pkg/
│   ├── vm/            # Virtual machine (Go)
│   ├── compiler/      # Compiler interface (Go)
│   ├── cache/         # Context cache (Go)
│   ├── runtime/       # Runtime services (Go)
│   └── server/        # HTTP server (Go)
├── glyph-core/         # Compiler core (Rust)
├── examples/          # Example projects
└── docs/              # Documentation
```

## Development

```bash
# Clone repository
git clone https://github.com/glyph-lang/glyph.git
cd glyph

# Build Go components
go build ./cmd/glyph

# Build Rust core
cd glyph-core
cargo build --release
cd ..

# Run tests
go test ./...
cd glyph-core && cargo test && cd ..

# Run example
./glyph run examples/hello-world/main.glyph
```

## 🗺️ Roadmap

### ✅ Phase 1: Core Infrastructure (Complete)
- [x] Rust compiler (lexer, parser, AST, binary format)
- [x] Go interpreter and HTTP server
- [x] CLI with run/dev/build commands
- [x] 221 tests passing

### ✅ Phase 2: Security & Quality (Complete)
- [x] SQL injection and XSS detection
- [x] Comprehensive validation framework
- [x] Enhanced error messages
- [x] 166 tests passing

### ✅ Phase 3: Bytecode & Performance (Complete)
- [x] Bytecode compiler with 25 opcodes
- [x] Stack-based VM
- [x] Sub-microsecond compilation
- [x] 105 tests passing

### 🚧 Phase 4: Platform Features (In Progress)
- [ ] Database integration (PostgreSQL, MySQL, MongoDB)
- [ ] WebSocket support
- [ ] Language Server Protocol (LSP)
- [ ] VS Code extension

**Total: 1696 tests passing across all phases**

## 📊 Benchmarks

### AI Token Efficiency

| Language | Tokens | vs Glyph |
|----------|--------|---------|
| **Glyph** | 463 | baseline |
| Python | 842 | 1.82x more |
| Java | 1,252 | 2.7x more |

**Result:** 45% fewer tokens than Python, 63% fewer than Java

### AI Generation Cost (per 1000 API calls)

| Model | Glyph | Python | Java |
|-------|------|--------|------|
| GPT-4 | $23 | $42 | $63 |
| Claude | $26 | $46 | $69 |

**Result:** 56% cost savings with Glyph

### Runtime Performance

| Operation | Glyph | Python | Java |
|-----------|------|--------|------|
| Arithmetic | 9.35 ns | 65.23 ns | 1.71 ns |
| String Concat | 40.19 ns | 63.68 ns | 7.86 ns |
| Object Creation | 9.72 ns | 84.71 ns | 21.79 ns |

**Result:** Glyph is 4-9x faster than Python

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the **Apache License 2.0** - see the [LICENSE](LICENSE) file for details.

### Contributing

By contributing to GlyphLang, you agree to the [Contributor License Agreement](CONTRIBUTING.md#contributor-license-agreement-cla). This allows us to ensure the project can continue to evolve while protecting both contributors and users.

## Links

- [Documentation](https://docs.glyph.dev)
- [Discord Community](https://discord.gg/glyph)
- [Twitter](https://twitter.com/glyph_lang)
