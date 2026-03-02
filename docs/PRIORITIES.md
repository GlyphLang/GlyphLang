# GlyphLang Strategic Priorities

> Derived from Phase 5 three-way intent hypothesis testing results (March 2026).
> Glyph scored 9.80/10 vs text 8.47/10 vs OpenAPI 8.00/10 across 5 scenarios.

---

## Priority 1: Fix P0 Runtime Blockers

**Why**: The hypothesis test proves Glyph's notation produces superior code. But
the runtime has critical bugs that would immediately destroy credibility with any
user who actually runs the generated output against the GlyphLang server.

**Issues** (from `CODE_REVIEW_ISSUES.md`):

| ID | Issue | Impact |
|----|-------|--------|
| P0-1 | `time.now()` returns hardcoded `1234567890` (Feb 2009) | Every timestamp is wrong |
| P0-2 | Async `execAsync()` shares state without sync | Race conditions in async code |
| P0-3 | Parser rejects struct field defaults (`= value`) | 4 example files broken |
| P0-4 | Parser rejects variable reassignment without `$` | Loop examples broken |
| P0-5 | No request body size limit in server handler | DoS vulnerability |
| P0-6 | Internal error details exposed to clients | Info leak vulnerability |
| P0-7 | WebSocket upgrader accepts all origins | CSWSH vulnerability |

**Acceptance**: `go test -race ./...` passes, all example files parse, security
issues have regression tests.

---

## Priority 2: Runtime Execution Tests for Generated Code

**Why**: Phase 5 validated generated code via `ast.parse()` and checklist
scoring — the servers have never received an HTTP request. A passing runtime
test suite would be 10x more convincing than static analysis scores.

**Deliverable**: A test harness that:
1. Installs Python dependencies (`fastapi`, `uvicorn`, `pydantic`)
2. Starts each generated FastAPI server on a random port
3. Sends requests to every endpoint
4. Asserts response shapes match the spec
5. Tests error cases (404, 401, 429)
6. Measures actual behavior (rate limiting, auth rejection)

**Scope**: Start with the 5 glyph-generated implementations (highest scoring),
then extend to text and OpenAPI.

---

## Priority 3: Hybrid Approach with OpenAPI

**Why**: Phase 5 showed each format wins at a different layer — Glyph for
architecture, OpenAPI for interface contracts, text for behavior. The strategic
move is "use Glyph WITH OpenAPI", not "instead of."

**Deliverables**:
- `@api: openapi("./spec.yaml")` import syntax — pull in HTTP contracts, layer
  Glyph structural patterns on top
- First-class OpenAPI export (already exists, make it prominent)
- Natural language annotation support for behavioral specs

**Value**: Lowers adoption barrier. Teams add Glyph incrementally to existing
OpenAPI workflows.

---

## Priority 4: Go Code Generation

**Why**: The compiler is written in Go. The team knows Go. Yet there's no Go
codegen target. This is the most natural third language and would:
- Validate Glyph patterns translate to non-Python idioms (interfaces, middleware
  chains, goroutines)
- Enable dogfooding — generate Go from Glyph, run it alongside the compiler
- Appeal to the infrastructure audience most likely to value Glyph's strengths

**Deliverable**: `glyph codegen --target go` producing idiomatic Go with
`net/http`, middleware chains, and dependency injection via function parameters.

---

## Priority 5: Market the Semantic Drift Problem

**Why**: Nobody adopts a language because of benchmark scores. They adopt it
because it solves a pain point. Phase 5 revealed a concrete, relatable one:
natural language prompts produce code with wrong field names, wrong response
shapes, and wrong data types. Every developer using AI code generation has hit
this.

**Key data points**:
- Text condition: `payment_id` instead of `id`, `recipient` instead of `to`,
  `str` timestamps instead of `int`
- Glyph condition: 100% checklist pass rate, zero field-name errors
- OpenAPI condition: correct field names (from schemas) but stub implementations

**The pitch**: "Stop debugging AI-generated field name mismatches." The notation
is the mechanism; drift prevention is the value.

---

## Priority 6: Ship a `glyph init` Experience

**Why**: There's no package manager, no project scaffolding, no quick-start
path. The first-run experience matters more than any feature. A 30-second path
from install to running server would do more for adoption than a third target
language.

**Deliverable**: `glyph init my-service` that:
1. Creates project structure with `main.glyph`
2. Includes a working CRUD example
3. Starts the dev server with hot reload
4. Prints "Your API is running at http://localhost:8080"

---

## Execution Order

```
Priority 1 (P0 fixes)  ──► Priority 2 (runtime tests)  ──► Priority 3 (OpenAPI hybrid)
                                                          └─► Priority 4 (Go codegen)
Priority 5 (marketing) can run in parallel
Priority 6 (glyph init) after Priority 1
```

Priorities 1 and 2 are prerequisites — they establish credibility. Priorities
3-6 build on that foundation.
