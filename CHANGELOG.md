# Changelog

All notable changes to GlyphLang will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-12-26

First production-ready release of GlyphLang.

### Added

#### Core Language
- **Type System**: int, string, bool, float, arrays, objects, optional (`T?`), union (`A | B`)
- **Pattern Matching**: match expressions with literal patterns, variable binding, wildcards, guards, object destructuring, array destructuring
- **Async/Await**: async blocks, await expressions, Future type with `All`, `Race`, `Any` combinators
- **Modules**: `import` statements, `from ... import { }` selective imports, module aliases, circular dependency detection
- **Generics**: type parameters on functions and types, type inference, constraints (`T: Numeric`, `T: Comparable`, `T: Any`)
- **Macros**: `macro!` definitions, compile-time expansion, parameter substitution, string interpolation
- **Control Flow**: if/else, while, for loops, switch statements
- **Functions**: user-defined functions with type annotations, lambdas, built-in functions

#### HTTP & WebSocket
- Route definitions with `@route /path [METHOD]` syntax
- Path parameters (`:id`), query parameters, request body handling
- Return type annotations with union types (`-> User | Error`)
- WebSocket routes with event handlers (`on connect`, `on disconnect`, `on message`, `on error`)
- WebSocket operations: `ws.send()`, `ws.broadcast()`, `ws.close()`

#### Middleware & Security
- Authentication middleware (`+ auth(jwt)`) with JWT support and RBAC
- Rate limiting (`+ ratelimit(100/min)`) with exponential backoff lockout
- CORS configuration with security hardening
- SQL injection prevention with identifier sanitization
- Command injection prevention
- XSS protection
- Security headers (X-Frame-Options, X-XSS-Protection, X-Content-Type-Options)
- Panic recovery with generic error messages (no information disclosure)

#### Database
- PostgreSQL integration with connection pooling
- Transaction support
- Migration system
- Query builder with parameterized queries
- Type-safe column validation

#### Compiler & Runtime
- Bytecode compiler with 3 optimization levels
- Dead code elimination
- Constant folding
- Function inlining
- Peephole optimization
- Stack-based virtual machine with 25+ opcodes
- JIT compilation with type specialization and hot path detection
- Hot reload for development

#### Observability
- Structured logging with configurable levels
- Prometheus metrics collection
- OpenTelemetry distributed tracing
- Health check endpoints

#### Tooling
- CLI with `run`, `compile`, `repl`, `lsp`, `decompile` commands
- LSP server with completions, diagnostics, hover, go-to-definition, find references, rename, code actions, formatting, signature help
- Bytecode decompiler
- Debug mode with breakpoints, variable inspection, step execution, REPL

#### Documentation
- Language Specification (docs/LANGUAGE_SPECIFICATION.md)
- API Reference (docs/API_REFERENCE.md)
- Quickstart Guide (docs/QUICKSTART.md)
- CLI Reference (docs/CLI.md)
- Architecture Design (docs/ARCHITECTURE_DESIGN.md)
- Performance Guide (docs/PERFORMANCE.md)
- Binary Format Specification (docs/BINARY_FORMAT.md)

### Quality
- 637+ tests across 22 packages
- 80%+ code coverage on 14 core packages
- 20/20 example compatibility (100%)
- Zero critical security vulnerabilities

### Performance
- Sub-microsecond compilation (~867 ns)
- Nanosecond execution (2.95-37.6 ns/op)
- Zero-allocation arithmetic operations

---

## [Unreleased]

### Planned
- Package manager with registry
- Code formatter
- Linter
- Profiler
- Docker/Kubernetes deployment support
- Serverless deployment (AWS Lambda, Google Cloud Functions)

[1.0.0]: https://github.com/GlyphLang/GlyphLang/releases/tag/v1.0.0
[Unreleased]: https://github.com/GlyphLang/GlyphLang/compare/v1.0.0...HEAD
