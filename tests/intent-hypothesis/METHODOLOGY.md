# Intent Hypothesis Testing — Evaluation Methodology

## Hypothesis

Glyph notation (`.glyph`) produces more correct, complete, and structurally precise AI-generated code than equivalent natural language descriptions (`.txt`) when used as input prompts for large language models.

## Experimental Design

### Test Subjects
- **Model**: Claude (Anthropic) — same model used for evaluation, noted as a limitation
- **Input pairs**: 5 scenarios, each with a `.glyph` and `.txt` version describing identical intent
- **Target output**: Python/FastAPI implementations

### Scenarios

| # | Scenario | Complexity | Key Features |
|---|----------|-----------|--------------|
| 01 | CRUD API | Low | REST endpoints, data models, error handling |
| 02 | Webhook Processor | Medium | Auth middleware, rate limiting, event routing |
| 03 | Chat Server | Medium | REST + WebSocket, JWT auth, real-time messaging |
| 04 | Job Queue | High | Queue workers, cron jobs, event handlers, REST |
| 05 | Auth Service | High | Registration, login, JWT, RBAC, crypto, cron |

### Generation Protocol

For each scenario, the LLM receives:
1. **Glyph condition**: The `.glyph` file + a standard prompt: *"Generate a complete Python/FastAPI implementation from this Glyph specification."*
2. **Text condition**: The `.txt` file + a standard prompt: *"Generate a complete Python/FastAPI implementation from this description."*

Each generation is performed in an isolated context (separate sub-agent) to prevent cross-contamination.

### Control Variables
- Same base prompt structure for both conditions
- Same target language (Python/FastAPI)
- Same model and temperature
- No additional context or examples provided

## Evaluation Metrics

### 1. Correctness (0–10)

Does the generated code implement the specified behavior without errors?

| Score | Criteria |
|-------|----------|
| 10 | All behavior matches spec exactly, no logical errors |
| 8 | Minor issues (e.g., missing a default value) |
| 6 | Moderate issues (e.g., wrong status code, missing validation) |
| 4 | Significant errors in business logic |
| 2 | Major structural errors, wouldn't function correctly |
| 0 | Fundamentally broken or unrelated output |

Checklist per scenario:
- [ ] Endpoints return correct response shapes
- [ ] Error cases handled as specified
- [ ] Data models match field names, types, and optionality
- [ ] Business logic follows specified flow
- [ ] Default values applied correctly

### 2. Completeness (0–10)

Does the code include ALL specified features?

| Score | Criteria |
|-------|----------|
| 10 | Every endpoint, model, middleware, and background task present |
| 8 | Missing 1 minor element |
| 6 | Missing 2–3 elements |
| 4 | Missing several elements or entire category |
| 2 | Only partial implementation |
| 0 | Stub or empty output |

Checklist per scenario:
- [ ] All REST endpoints present with correct HTTP methods and paths
- [ ] All data models/schemas defined
- [ ] All middleware applied (auth, rate limiting)
- [ ] All background tasks (cron, queue workers, event handlers)
- [ ] All dependency injections present

### 3. Structural Precision (0–10)

Does the code match the intended architecture and patterns?

| Score | Criteria |
|-------|----------|
| 10 | Idiomatic FastAPI, proper dependency injection, correct typing |
| 8 | Minor structural deviations (e.g., inline vs. Depends()) |
| 6 | Moderate deviations (e.g., global state instead of DI) |
| 4 | Non-idiomatic patterns, works but poorly structured |
| 2 | Wrong framework patterns entirely |
| 0 | Not recognizable as FastAPI |

Checklist per scenario:
- [ ] Pydantic models used for request/response types
- [ ] FastAPI dependency injection (Depends()) for providers
- [ ] Proper route decorators and path parameters
- [ ] Type annotations present and correct
- [ ] Appropriate use of async/await

### 4. Token Efficiency (measured, not scored)

Quantitative comparison of input size vs. output quality:
- **Input characters**: raw byte count of the input prompt
- **Input words**: word count of the input prompt
- **Quality-per-word**: (Correctness + Completeness + Structural Precision) / 3 / input words × 100

## Scoring Process

1. Generate all 10 implementations (5 × 2 conditions)
2. For each implementation, apply the three rubrics independently
3. Record scores in a comparison table
4. Calculate aggregate statistics
5. Analyze per-scenario and overall trends

## Limitations

1. **Self-evaluation bias**: The same model family generates and evaluates the code. Future work should use human evaluators or a different model for evaluation.
2. **Single model**: Results may not generalize to other LLMs (GPT-4, Gemini, etc.).
3. **Single target language**: Python/FastAPI only. The hypothesis should be tested across multiple targets.
4. **Sample size**: 5 scenarios is small. Statistical significance is limited.
5. **Prompt sensitivity**: Results may vary with different prompt formulations.
