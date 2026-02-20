# Intent Hypothesis Testing — Results

## Hypothesis

> Glyph notation (`.glyph`) produces more correct, complete, and structurally precise AI-generated code than equivalent natural language descriptions (`.txt`) when used as input prompts for large language models.

## Generation Details

- **Model**: Claude (Anthropic) — same model for generation and evaluation (noted limitation)
- **Target language**: Python / FastAPI
- **Generation protocol**: Each implementation generated in an isolated sub-agent context
- **Total implementations**: 10 (5 glyph-sourced + 5 text-sourced)

---

## Per-Scenario Evaluation

### Scenario 01: CRUD API (Low Complexity)

**Glyph-Generated** (`generated/glyph/01-crud-api.py`)

| Metric | Score | Notes |
|--------|-------|-------|
| Correctness | 10 | All 5 endpoints implement exact spec behavior. Response shapes match (`{deleted: true}`). Error cases return proper 404s. Default `completed=false` applied. |
| Completeness | 10 | All endpoints, all 3 models, Depends(get_db) injection present. |
| Structural Precision | 10 | Separate `TodoCollection` class mirroring Glyph `db.todos.Find/Get/Create/Update/Delete`. Pydantic models, Depends(), proper decorators, type annotations, async handlers. |

**Text-Generated** (`generated/text/01-crud-api.py`)

| Metric | Score | Notes |
|--------|-------|-------|
| Correctness | 9 | All endpoints functional. Minor: DELETE returns `{"detail": "Todo deleted"}` instead of `{"deleted": true}` — text spec doesn't specify success response body, so this is a reasonable interpretation but less precise than glyph output. |
| Completeness | 10 | All endpoints and models present. DI with Depends(get_db). |
| Structural Precision | 9 | Monolithic Database class (all CRUD methods together) vs. glyph's collection-per-entity pattern. Still idiomatic FastAPI. snake_case method names. |

**Scenario 01 Summary**: Glyph 30/30 vs Text 28/30. Glyph's explicit `> {deleted: true}` response notation produced an exact match, while text's ambiguous "Delete a todo" left response shape to interpretation.

---

### Scenario 02: Webhook Processor (Medium Complexity)

**Glyph-Generated** (`generated/glyph/02-webhook-processor.py`)

| Metric | Score | Notes |
|--------|-------|-------|
| Correctness | 9 | Both endpoints implement spec behavior. Event routing correct (`payment.succeeded`, `payment.failed`, `customer.created`). Data access uses `input.data["id"]` matching glyph's `input.data.id`. Minor: API key validation checks non-empty only (no key store). |
| Completeness | 10 | Both endpoints, both models, API key auth, rate limiting (1000/min), all 3 event handlers, webhook logging. |
| Structural Precision | 9 | APIKeyHeader security scheme, Depends() for auth AND rate limiting (separate dependencies). Separate collection classes. Minor over-engineering with Security vs Depends. |

**Text-Generated** (`generated/text/02-webhook-processor.py`)

| Metric | Score | Notes |
|--------|-------|-------|
| Correctness | 8 | Both endpoints functional. Event routing works. But uses `data.get("payment_id")` instead of `data.get("id")` — text spec doesn't specify field name, so this is an inference. Hardcoded `API_KEYS` set is arguably more realistic but not spec-driven. |
| Completeness | 9 | Both endpoints, models, auth, rate limiting present. Rate limiting called inline (`check_rate_limit(api_key)`) rather than as a Depends() middleware. |
| Structural Precision | 7 | Rate limit is a plain function call inside the handler, not a dependency. Global dicts/lists for storage instead of collection classes. Event handlers use a dispatch dict (clean pattern but flat). |

**Scenario 02 Summary**: Glyph 28/30 vs Text 24/30. Glyph's `+ ratelimit(1000/min)` and `+ auth(apikey)` middleware notation directly mapped to FastAPI Depends() pattern. The `db.payments.Update(input.data.id, ...)` notation preserved exact field access paths.

---

### Scenario 03: Chat Server (Medium Complexity)

**Glyph-Generated** (`generated/glyph/03-chat-server.py`)

| Metric | Score | Notes |
|--------|-------|-------|
| Correctness | 9 | All REST endpoints correct. WebSocket lifecycle matches spec exactly (broadcast on connect/message/disconnect). Minor: WebSocket messages not stored to DB, so GET /rooms/:id/messages always returns empty — this gap exists in the glyph spec itself (`ws.broadcast(input)` doesn't include `db.messages.Create()`). |
| Completeness | 9 | All REST endpoints and WebSocket present. All models defined. JWT auth. Missing: message persistence during WebSocket communication (spec gap). |
| Structural Precision | 9 | ConnectionManager class, RoomCollection/MessageCollection, HTTPBearer JWT. Proper async. Collection classes mirror glyph notation. |

**Text-Generated** (`generated/text/03-chat-server.py`)

| Metric | Score | Notes |
|--------|-------|-------|
| Correctness | 10 | All REST endpoints correct. WebSocket implements ack-back to sender AND broadcast (matching text spec's "send acknowledgment back to sender, then broadcast"). Messages stored to DB during WebSocket, making GET messages endpoint functional. 404 on invalid room_id (extra safety). |
| Completeness | 10 | All REST + WebSocket endpoints. All models. JWT auth. Message persistence. Structured JSON messages (`{type, content, timestamp}`). |
| Structural Precision | 9 | Pydantic models, Depends(), HTTPBearer. send_json for structured WebSocket messages. Status constants (HTTP_401_UNAUTHORIZED). Database methods well-organized. |

**Scenario 03 Summary**: Glyph 27/30 vs Text 29/30. Text spec described richer WebSocket behavior ("send acknowledgment back to sender, then broadcast") that glyph's `ws.broadcast(input)` notation couldn't express. The text-generated code stored messages to DB, making the history endpoint functional.

---

### Scenario 04: Job Queue (High Complexity)

**Glyph-Generated** (`generated/glyph/04-job-queue.py`)

| Metric | Score | Notes |
|--------|-------|-------|
| Correctness | 10 | Queue workers extract correct fields and return specified shapes. Cron jobs at correct schedules query correct data. Event handler captures email/name. REST endpoints with auth and proper error handling. |
| Completeness | 10 | Both queue workers, both cron jobs, event handler, both REST endpoints, both data models, JWT auth. Full `QueueWorker`, `CronScheduler`, `EventBus` abstractions. |
| Structural Precision | 10 | `asyncio.Queue`-backed QueueWorker class. CronScheduler with named job registration. EventBus with pub/sub pattern. Separate collection classes. All async. Depends() throughout. |

**Text-Generated** (`generated/text/04-job-queue.py`)

| Metric | Score | Notes |
|--------|-------|-------|
| Correctness | 9 | Queue workers and event handler correct. Cron schedules correct. REST endpoints work. GET /jobs returns `JobStatusResponse(id, status)` — matches text spec's "Check job status" wording. |
| Completeness | 8 | All components present. But `Database.query()` always returns `[]` (stub), making cron jobs non-functional at runtime. Extra response models (JobRecord, JobStatusResponse, JobSubmitResponse) show good typing. |
| Structural Precision | 8 | Sync queue workers (not async). Simple dict dispatchers for workers/cron/events. Stub DB query. Response models well-typed. Depends() for auth. |

**Scenario 04 Summary**: Glyph 30/30 vs Text 25/30. Glyph's `&` (queue worker), `*` (cron), and `~` (event handler) symbols produced richly structured abstractions (QueueWorker, CronScheduler, EventBus classes). Glyph's `% db: Database` and `db.sessions.Where(...)` notation led to fully functional collection classes.

---

### Scenario 05: Auth Service (High Complexity)

**Glyph-Generated** (`generated/glyph/05-auth-service.py`)

| Metric | Score | Notes |
|--------|-------|-------|
| Correctness | 9 | All 7 endpoints/tasks implemented. Password hashing (bcrypt via CryptContext), JWT signing, session management all correct. Rate limits at correct thresholds. Minor: rate limit keyed by email (not IP), refresh returns AuthResponse with full user (requires extra DB lookup). |
| Completeness | 10 | All endpoints, all models, JWT auth, rate limiting (10/20/50), admin role check, session management, cron task. |
| Structural Precision | 10 | CryptContext for bcrypt, jwt_sign/crypto_hash/crypto_verify helpers mirroring glyph's `crypto.hash()`/`crypto.verify()`/`jwt.sign()`. Depends() with `require_admin` chained dependency. Separate UserCollection/SessionCollection. |

**Text-Generated** (`generated/text/05-auth-service.py`)

| Metric | Score | Notes |
|--------|-------|-------|
| Correctness | 10 | All endpoints correctly implemented. `sign_token()` creates session + token atomically (token rotation on refresh). `UserOut` strips password_hash from responses (security best practice). Rate limit by IP (more realistic). `query_expired_sessions` uses time-based expiration (functional cleanup). |
| Completeness | 10 | All endpoints, all models (plus UserOut, TokenResponse), JWT auth, rate limiting, admin check, session lifecycle management, background task. |
| Structural Precision | 10 | `passlib.hash.bcrypt` + `jose.jwt` for crypto. `response_model` annotations on routes. `sign_token()` encapsulates token+session creation. `user_to_out()` helper. `status.HTTP_401_UNAUTHORIZED` constants. `require_admin` chained Depends(). Proper session lifecycle (create/delete/rotate). |

**Scenario 05 Summary**: Glyph 29/30 vs Text 30/30. The text description's prose ("Hashes password with crypto.hash (maps to bcrypt/argon2 in target language)") enabled the LLM to produce production-quality security patterns. Both outputs are excellent at this complexity level.

---

## Aggregate Results

### Score Comparison Table

| Scenario | Complexity | Glyph Correctness | Text Correctness | Glyph Completeness | Text Completeness | Glyph Precision | Text Precision | Glyph Total | Text Total |
|----------|-----------|-------------------|------------------|--------------------|--------------------|-----------------|----------------|-------------|------------|
| 01 CRUD API | Low | 10 | 9 | 10 | 10 | 10 | 9 | **30** | 28 |
| 02 Webhook | Medium | 9 | 8 | 10 | 9 | 9 | 7 | **28** | 24 |
| 03 Chat | Medium | 9 | 10 | 9 | 10 | 9 | 9 | 27 | **29** |
| 04 Job Queue | High | 10 | 9 | 10 | 8 | 10 | 8 | **30** | 25 |
| 05 Auth | High | 9 | 10 | 10 | 10 | 10 | 10 | 29 | **30** |
| **Total** | | **47** | **46** | **49** | **47** | **48** | **43** | **144** | **136** |
| **Average** | | **9.4** | **9.2** | **9.8** | **9.4** | **9.6** | **8.6** | **9.6** | **9.1** |

### Per-Metric Averages

| Metric | Glyph Avg | Text Avg | Delta |
|--------|-----------|----------|-------|
| Correctness | 9.4 | 9.2 | +0.2 |
| Completeness | 9.8 | 9.4 | +0.4 |
| Structural Precision | 9.6 | 8.6 | **+1.0** |
| **Overall** | **9.6** | **9.1** | **+0.5** |

### Token Efficiency

| Scenario | Glyph Bytes | Glyph Words | Text Bytes | Text Words | Glyph Quality | Text Quality | Glyph QPW | Text QPW |
|----------|------------|-------------|------------|------------|---------------|--------------|-----------|----------|
| 01 CRUD | 1,067 | 176 | 900 | 138 | 10.0 | 9.3 | 5.68 | 6.76 |
| 02 Webhook | 1,128 | 130 | 1,085 | 151 | 9.3 | 8.0 | 7.18 | 5.30 |
| 03 Chat | 978 | 142 | 1,096 | 162 | 9.0 | 9.7 | 6.34 | 5.97 |
| 04 Job Queue | 1,625 | 249 | 1,205 | 188 | 10.0 | 8.3 | 4.02 | 4.43 |
| 05 Auth | 3,678 | 485 | 2,022 | 296 | 9.7 | 10.0 | 1.99 | 3.38 |
| **Average** | **1,695** | **236** | **1,262** | **187** | **9.6** | **9.1** | **5.04** | **5.17** |

*QPW = Quality-Per-Word = Average(Correctness, Completeness, Structural Precision) / Input Words × 100*

---

## Analysis

### Finding 1: Glyph Produces Higher Absolute Quality (+0.5 average)

Glyph-sourced implementations scored higher overall (9.6 vs 9.1 average across all metrics). The largest advantage was in **structural precision** (+1.0), followed by completeness (+0.4) and correctness (+0.2).

### Finding 2: Structural Precision Is Glyph's Strongest Advantage (+1.0)

Glyph's symbolic notation directly maps to FastAPI patterns:
- `+ auth(jwt)` → `Depends(auth)` middleware
- `+ ratelimit(N/min)` → `Depends(check_rate_limit)` dependency
- `% db: Database` → `Depends(get_db)` injection
- `db.collection.Method()` → typed collection classes with matching methods
- `&` / `*` / `~` → QueueWorker, CronScheduler, EventBus class abstractions

Text descriptions of the same concepts ("requires JWT auth", "rate limited to N/min") produced flatter structures — inline function calls, global dicts, stub implementations.

### Finding 3: Text Wins When Behavioral Nuance Matters

In two scenarios, text-generated code scored higher:
- **03 Chat Server**: Text spec's "send acknowledgment back to sender, then broadcast" captured WebSocket ack semantics that glyph's `ws.broadcast(input)` could not express. The text version also stored messages to DB, making the history endpoint functional.
- **05 Auth Service**: Text spec's detailed prose about session management, password hashing, and token lifecycle produced code with token rotation, UserOut response filtering, and time-based session expiration.

Glyph's terse symbols (`ws.broadcast`, `crypto.hash`) capture WHAT happens but not HOW it should behave. Natural language excels at expressing behavioral nuances and security considerations.

### Finding 4: Text Is Slightly More Token-Efficient

Despite lower quality scores, text descriptions achieved comparable quality-per-word (5.17 vs 5.04) because they use fewer words on average (187 vs 236). Glyph files are larger due to structural notation (type definitions, explicit field annotations, middleware declarations). This finding is counterintuitive — glyph's conciseness advantage is in *structural density* (information per symbol), not raw character count.

### Finding 5: Glyph's Advantage Increases with Middleware/Infrastructure Complexity

| Complexity | Glyph Avg | Text Avg | Delta |
|-----------|-----------|----------|-------|
| Low (01) | 10.0 | 9.3 | +0.7 |
| Medium (02, 03) | 9.2 | 8.8 | +0.3 |
| High — Infrastructure (04) | 10.0 | 8.3 | **+1.7** |
| High — Behavioral (05) | 9.7 | 10.0 | **-0.3** |

Glyph's biggest win (04 Job Queue, +1.7) was in a scenario heavy on infrastructure patterns (queue workers, cron jobs, event handlers) where glyph has dedicated symbols (`&`, `*`, `~`). Its only loss (05 Auth Service, -0.3) was in a scenario heavy on behavioral logic (password verification flow, session lifecycle, token rotation) where prose descriptions carried more context.

---

## Conclusion

**The hypothesis is partially confirmed.** Glyph notation produces code that is more structurally precise (+1.0), more complete (+0.4), and slightly more correct (+0.2) than equivalent natural language descriptions. The advantage is strongest for infrastructure-heavy patterns (middleware, background jobs, dependency injection) where glyph's symbolic notation directly maps to framework idioms.

However, natural language descriptions produce superior results when behavioral nuance matters — security flows, protocol lifecycles, and edge-case handling. Text descriptions are also slightly more token-efficient per word.

**Recommendation**: Glyph notation is most valuable as a structural scaffold for middleware, data models, and service architecture. It could be complemented with natural language annotations for behavioral specifications — a hybrid approach that combines glyph's structural precision with prose's behavioral expressiveness.

---

## Limitations

1. **Self-evaluation bias**: Same model family (Claude) generated and evaluated all code
2. **Single model**: Results may differ with GPT-4, Gemini, etc.
3. **Single target**: Python/FastAPI only
4. **Small sample**: 5 scenarios; statistical significance is limited
5. **Prompt sensitivity**: Different prompt formulations may produce different results
6. **Evaluator subjectivity**: Rubric application involves judgment calls on score boundaries
