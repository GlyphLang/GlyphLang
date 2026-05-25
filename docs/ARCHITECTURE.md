# Architecture Overview

The canonical architecture reference for GlyphLang is [ARCHITECTURE_DESIGN.md](ARCHITECTURE_DESIGN.md), which contains full mermaid diagrams for:

- Compilation pipeline
- Package dependency graph
- Request execution flow
- AST node hierarchy
- VM architecture
- WebSocket architecture
- JIT compilation tiers
- Database integration

This file provides a quick-reference package map and pointers to the relevant guides for contributors who need to extend the system.

---

## Compilation Pipeline (Summary)

```
.glyph source
     |
     v
  Lexer (pkg/parser/lexer.go)
     |  tokens
     v
  Parser (pkg/parser/parser.go)
     |  *ast.Module
     v
  Type Checker (pkg/interpreter/typechecker.go)
     |
     +-----> Interpreter (pkg/interpreter/)   <-- glyph run / glyph dev
     |
     v
  Compiler (pkg/compiler/compiler.go)
     |  []byte bytecode
     v
  Optimizer (pkg/compiler/optimizer.go)
     |
     v
  VM (pkg/vm/vm.go)  <-- glyph compile + execute
     |
     +-----> JIT (pkg/jit/jit.go)             <-- hot paths (100+ executions)

  Parallel path for codegen:
  *ast.Module --> IR Analyzer (pkg/ir/analyzer.go)
                      |
                      +--> Python/FastAPI (pkg/codegen/python.go)
                      +--> TypeScript/Express (pkg/codegen/typescript_server.go)
```

---

## Package Quick-Reference

| Package | Role |
|---|---|
| `pkg/parser/` | Lexer + recursive descent parser |
| `pkg/ast/` | AST node definitions (~40 node types) |
| `pkg/interpreter/` | Tree-walking interpreter, type checker, built-ins |
| `pkg/compiler/` | Bytecode compiler, macro expander, optimizer (4 levels) |
| `pkg/vm/` | Stack-based virtual machine (37 opcodes) |
| `pkg/jit/` | JIT with type specialization and hot path detection |
| `pkg/ir/` | Language-neutral semantic IR for multi-target codegen |
| `pkg/codegen/` | Python/FastAPI and TypeScript/Express generators |
| `pkg/server/` | Built-in HTTP server, routing, middleware, WebSockets |
| `pkg/database/` | ORM + drivers for PostgreSQL, MySQL, SQLite, MongoDB |
| `pkg/security/` | JWT, rate limiting, CORS, SQL injection prevention |
| `pkg/lsp/` | Language Server Protocol (VS Code, Neovim) |
| `pkg/decompiler/` | Bytecode disassembly |
| `pkg/validate/` | Structured JSON validation output for AI tooling |
| `cmd/glyph/` | CLI entry point and command handlers |

---

## Key Design Decisions

See [DECISIONS.md](../DECISIONS.md) for the rationale behind architectural choices.

---

## Going Deeper

| Topic | Document |
|---|---|
| Adding an opcode or language feature | [DEVELOPMENT.md](DEVELOPMENT.md) |
| Bytecode file format and type encoding | [BINARY_FORMAT.md](BINARY_FORMAT.md) |
| Language syntax and type system | [LANGUAGE_SPECIFICATION.md](LANGUAGE_SPECIFICATION.md) |
| Multi-language code generation | [POLYGLOT_ROADMAP.md](POLYGLOT_ROADMAP.md) |
| Symbol vocabulary and EBNF grammar | [GLYPH_NOTATION_SPEC.md](GLYPH_NOTATION_SPEC.md) |
| Full architecture diagrams | [ARCHITECTURE_DESIGN.md](ARCHITECTURE_DESIGN.md) |
