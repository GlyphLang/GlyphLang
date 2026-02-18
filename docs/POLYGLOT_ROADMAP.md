# Polyglot Code Generation Roadmap

## Status: In Progress

The Semantic IR, generalized provider system, and Python/FastAPI codegen are merged to main (PR #170). This document tracks the remaining work to make polyglot code generation fully operational.

## Completed

- [x] **Semantic IR** (`pkg/ir/`) — Language-neutral intermediate representation with full type system, expression/statement IR, and two-pass AST analyzer
- [x] **Generalized provider system** — Generic provider registry replacing hardcoded if/else chain, backward-compatible with existing Database/Redis/MongoDB/LLM types
- [x] **Python/FastAPI codegen** (`pkg/codegen/python.go`) — Generates complete FastAPI apps with Pydantic models, Depends() injection, APScheduler cron, provider stubs
- [x] **Formal notation spec** (`docs/GLYPH_NOTATION_SPEC.md`) — Symbol vocabulary, type mapping table, EBNF grammar
- [x] **Intent-test examples** (`examples/intent-tests/`) — 5 validated .glyph + .txt pairs
- [x] **Provider parser syntax** — `provider` keyword with IR wiring and validation
- [x] **TypeScript/Express codegen** (`pkg/codegen/typescript_server.go`) — Generates complete Express apps with TypeScript interfaces, provider stubs, node-cron
- [x] **Integration tests** (`tests/codegen_integration_test.go`) — End-to-end pipeline tests for all intent-test examples across both targets

## Phase 1: CLI Pipeline (complete)

**Goal**: `glyph codegen app.glyph` outputs a working FastAPI application.

- [x] Add `codegen` subcommand to `cmd/glyph/`
- [x] Wire up: parse .glyph → AST → IR analyzer → Python generator → write output files
- [x] Support `--output` flag for target directory
- [x] Support `--lang python` flag (extensible for future targets)
- [x] Generate project structure: `main.py`, `requirements.txt`

## Phase 2: Provider Parser Syntax (complete)

**Goal**: Parse `provider` declarations in .glyph source files.

- [x] Add `provider` keyword to parser (IDENT-based, no compact symbol needed)
- [x] Add parser rule for `provider Name { method(params) -> ReturnType }` blocks
- [x] Produce `ProviderDef` AST nodes from parsed source (AST types already existed)
- [x] Wire `ProviderDef` through IR analyzer with method signatures
- [x] Add validation in the `validate` command (duplicate detection, method type checking, injection validation)
- [x] Add examples demonstrating custom providers (`examples/custom-provider/`)

## Phase 3: Second Target Language (complete)

**Goal**: Prove polyglot promise with a second codegen backend.

- [x] TypeScript/Express generator (`pkg/codegen/typescript_server.go`)
- [x] Reuse the same `ServiceIR` input, different output
- [x] Extend `glyph codegen --lang typescript` support
- [x] Verify identical behavior from the same .glyph source across both targets

## Phase 4: Integration Tests (complete)

**Goal**: End-to-end tests from .glyph source to generated output.

- [x] Parse each intent-test .glyph → IR → Python, verify output structure (`tests/codegen_integration_test.go`)
- [x] Parse each intent-test .glyph → IR → TypeScript, verify output structure
- [x] Regression tests: both targets tested against all 5 intent-test files + custom-provider example
- [x] Add to CI pipeline (included via `go test ./...` in CI workflow and Makefile)

## Phase 5: Intent Hypothesis Testing

**Goal**: Validate that .glyph notation produces better AI-generated code than natural language.

- [ ] Design evaluation methodology (metrics: correctness, completeness, token efficiency)
- [ ] Feed .glyph files to LLM, generate target-language implementations
- [ ] Feed .txt files to LLM, generate same implementations
- [ ] Compare results across the 5 intent-test scenarios
- [ ] Document findings

## Phase 6: Expand IR Coverage (complete)

**Goal**: Handle all Glyph language features in the IR.

- [x] WebSocket details (connect/message/disconnect handlers, room management)
- [x] GraphQL schema/resolver generation
- [x] gRPC service/method definitions
- [x] Pattern matching in IR → target language switch/match statements
- [x] Async/await mapping per target language

## When Polyglot Is Ready: Documentation & Website Updates

When at least 2 target languages are working end-to-end (Phase 3 complete), update the following:

### GlyphLangSite (`/home/dadams/projects/glyph/GlyphLangSite`)

- [x] **Landing page** (`index.html`): Add polyglot code generation to the features grid and use cases
- [x] **Landing page**: Update "Built with 100% pure Go" messaging to reflect multi-language output
- [x] **Landing page**: Add a code comparison showing the same `.glyph` generating Python and TypeScript
- [x] **Docs page** (`docs.html`): Expand the `codegen` CLI section with all supported languages
- [x] **Docs page**: Add a "Polyglot Code Generation" section explaining one `.glyph` → multiple languages
- [x] **Docs page**: Update the dependency injection / provider documentation to cover custom providers
- [x] **Docs page**: Add examples showing generated output for each target language

### VS Code Extension (`/home/dadams/projects/glyph/vscode-extension`)

- [ ] Add `codegen` command integration (run from editor, pick target language)
- [ ] Add preview panel for generated output
- [ ] Update README with polyglot features

### Main Repo (`/home/dadams/projects/glyph/GlyphLang`)

- [x] Update root README.md with polyglot code generation feature
- [ ] Add generated output examples to `examples/` directory
- [ ] Update CLAUDE.md if the workflow changes
