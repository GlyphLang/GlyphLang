# Polyglot Code Generation Roadmap

## Status: In Progress

The Semantic IR, generalized provider system, and Python/FastAPI codegen are merged to main (PR #170). This document tracks the remaining work to make polyglot code generation fully operational.

## Completed

- [x] **Semantic IR** (`pkg/ir/`) — Language-neutral intermediate representation with full type system, expression/statement IR, and two-pass AST analyzer
- [x] **Generalized provider system** — Generic provider registry replacing hardcoded if/else chain, backward-compatible with existing Database/Redis/MongoDB/LLM types
- [x] **Python/FastAPI codegen** (`pkg/codegen/python.go`) — Generates complete FastAPI apps with Pydantic models, Depends() injection, APScheduler cron, provider stubs
- [x] **Formal notation spec** (`docs/GLYPH_NOTATION_SPEC.md`) — Symbol vocabulary, type mapping table, EBNF grammar
- [x] **Intent-test examples** (`examples/intent-tests/`) — 5 validated .glyph + .txt pairs

## Phase 1: CLI Pipeline (next)

**Goal**: `glyph codegen python app.glyph` outputs a working FastAPI application.

- [ ] Add `codegen` subcommand to `cmd/glyph/`
- [ ] Wire up: parse .glyph → AST → IR analyzer → Python generator → write output files
- [ ] Support `--output` flag for target directory
- [ ] Support `--lang python` flag (extensible for future targets)
- [ ] Generate project structure: `main.py`, `models.py`, `requirements.txt`

## Phase 2: Provider Parser Syntax

**Goal**: Parse `provider` declarations in .glyph source files.

- [ ] Add `provider` keyword to lexer (both compact and expanded)
- [ ] Add parser rule for `provider Name { method(...): ReturnType }` blocks
- [ ] Produce `ProviderDef` AST nodes from parsed source
- [ ] Add validation in the `validate` command
- [ ] Add examples demonstrating custom providers

## Phase 3: Second Target Language

**Goal**: Prove polyglot promise with a second codegen backend.

- [ ] TypeScript/Express generator (`pkg/codegen/typescript.go`) — or Go/Chi
- [ ] Reuse the same `ServiceIR` input, different output
- [ ] Extend `glyph codegen --lang typescript` support
- [ ] Verify identical behavior from the same .glyph source across both targets

## Phase 4: Integration Tests

**Goal**: End-to-end tests from .glyph source to generated output.

- [ ] Parse each intent-test .glyph → IR → Python, verify output compiles/runs
- [ ] Parse each intent-test .glyph → IR → TypeScript, verify output compiles/runs
- [ ] Regression tests: any .glyph change must still produce valid output for all targets
- [ ] Add to CI pipeline

## Phase 5: Intent Hypothesis Testing

**Goal**: Validate that .glyph notation produces better AI-generated code than natural language.

- [ ] Design evaluation methodology (metrics: correctness, completeness, token efficiency)
- [ ] Feed .glyph files to LLM, generate target-language implementations
- [ ] Feed .txt files to LLM, generate same implementations
- [ ] Compare results across the 5 intent-test scenarios
- [ ] Document findings

## Phase 6: Expand IR Coverage

**Goal**: Handle all Glyph language features in the IR.

- [ ] WebSocket details (connect/message/disconnect handlers, room management)
- [ ] GraphQL schema/resolver generation
- [ ] gRPC service/method definitions
- [ ] Pattern matching in IR → target language switch/match statements
- [ ] Async/await mapping per target language
