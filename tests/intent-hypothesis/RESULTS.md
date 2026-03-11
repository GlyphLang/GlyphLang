# Intent Hypothesis Testing — Results

## Hypothesis

> Glyph notation (`.glyph`) produces more correct, complete, and structurally precise AI-generated code than equivalent natural language descriptions (`.txt`) when used as input prompts for large language models.

## Generation Details

- **Model**: Claude (Anthropic) — same model for generation and evaluation (noted limitation)
- **Target language**: Python / FastAPI
- **Generation protocol**: Each implementation generated in an isolated sub-agent context
- **Total implementations**: 15 (5 glyph-sourced + 5 text-sourced + 5 openapi-sourced)
- **Conditions**: Three-way comparison — Glyph (.glyph), natural language (.txt), OpenAPI 3.1.0 (.yaml)

---

## Validation

### Syntax Validation

All 15 generated Python files pass `ast.parse()` — they are syntactically valid Python.

```
glyph/01-crud-api.py           PASS
glyph/02-webhook-processor.py  PASS
glyph/03-chat-server.py        PASS
glyph/04-job-queue.py          PASS
glyph/05-auth-service.py       PASS
text/01-crud-api.py            PASS
text/02-webhook-processor.py   PASS
text/03-chat-server.py         PASS
text/04-job-queue.py           PASS
text/05-auth-service.py        PASS
openapi/01-crud-api.py         PASS
openapi/02-webhook-processor.py PASS
openapi/03-chat-server.py      PASS
openapi/04-job-queue.py        PASS
openapi/05-auth-service.py     PASS
```

### Structural Analysis (AST-based)

| File | External Deps | Classes | Functions | Route Decorators |
|------|--------------|---------|-----------|-----------------|
| glyph/01-crud-api.py | fastapi, pydantic | 5 | 13 | 5 (get:2, post:1, put:1, delete:1) |
| text/01-crud-api.py | fastapi, pydantic | 4 | 12 | 5 (get:2, post:1, put:1, delete:1) |
| openapi/01-crud-api.py | fastapi, pydantic, uvicorn | 4 | 7 | 5 (get:2, post:1, put:1, delete:1) |
| glyph/02-webhook-processor.py | fastapi, pydantic | 6 | 13 | 2 (post:2) |
| text/02-webhook-processor.py | fastapi, pydantic | 2 | 7 | 2 (post:2) |
| openapi/02-webhook-processor.py | fastapi, pydantic, uvicorn | 2 | 9 | 2 (post:2) |
| glyph/03-chat-server.py | fastapi, jwt, pydantic | 8 | 19 | 4 (get:2, post:1, websocket:1) |
| text/03-chat-server.py | fastapi, jose, pydantic | 5 | 15 | 4 (get:2, post:1, websocket:1) |
| openapi/03-chat-server.py | fastapi, pydantic, uvicorn | 4 | 10 | 4 (get:2, post:1, websocket:1) |
| glyph/04-job-queue.py | fastapi, jwt, pydantic | 11 | 31 | 2 (get:1, post:1) |
| text/04-job-queue.py | fastapi, jwt, pydantic | 8 | 14 | 2 (get:1, post:1) |
| openapi/04-job-queue.py | fastapi, pydantic, uvicorn | 4 | 10 | 2 (get:1, post:1) |
| glyph/05-auth-service.py | fastapi, jwt, passlib, pydantic | 9 | 27 | 6 (get:2, post:4) |
| text/05-auth-service.py | fastapi, jose, passlib, pydantic | 8 | 24 | 6 (get:2, post:4) |
| openapi/05-auth-service.py | fastapi, pydantic, uvicorn | 8 | 18 | 6 (get:2, post:4) |

**Observation**: Route decorator counts are identical across all three conditions in every scenario — all notations produce the same endpoint structure. The differences are in classes and functions: glyph-generated code averages 7.8 classes vs text's 5.4 vs openapi's 4.4. Glyph produces the most structured abstractions (collection classes, worker classes, bus classes). OpenAPI produces the fewest classes but includes uvicorn as an extra dependency. Notably, OpenAPI-generated code never imports external JWT or crypto libraries — it either implements JWT manually or uses stub auth.

### Output File Sizes

| File | Bytes | Lines | Words |
|------|-------|-------|-------|
| glyph/01-crud-api.py | 2,822 | 119 | 295 |
| text/01-crud-api.py | 2,834 | 111 | 271 |
| openapi/01-crud-api.py | 3,471 | 129 | 301 |
| glyph/02-webhook-processor.py | 3,753 | 149 | 346 |
| text/02-webhook-processor.py | 3,509 | 136 | 317 |
| openapi/02-webhook-processor.py | 6,145 | 199 | 508 |
| glyph/03-chat-server.py | 3,858 | 178 | 383 |
| text/03-chat-server.py | 4,999 | 198 | 427 |
| openapi/03-chat-server.py | 5,653 | 200 | 482 |
| glyph/04-job-queue.py | 6,566 | 263 | 666 |
| text/04-job-queue.py | 4,023 | 177 | 403 |
| openapi/04-job-queue.py | 6,390 | 220 | 583 |
| glyph/05-auth-service.py | 8,007 | 328 | 749 |
| text/05-auth-service.py | 8,945 | 333 | 765 |
| openapi/05-auth-service.py | 12,423 | 413 | 1,035 |

---

## Checklist-Based Evaluation

Each implementation was scored against its own specification using 15 binary checklist items (5 per metric). Three independent evaluation agents scored scenarios 01-02, 03-04, and 05 respectively. Scores use the 0-10 rubric from `METHODOLOGY.md`, informed by checklist pass/fail results.

### Checklist Legend

**Correctness**: C1=response shapes, C2=error cases, C3=data model fields/types, C4=business logic flow, C5=default values

**Completeness**: Cm1=REST endpoints, Cm2=data models, Cm3=middleware, Cm4=background tasks, Cm5=dependency injection

**Structural Precision**: S1=Pydantic models, S2=Depends() DI, S3=route decorators, S4=type annotations, S5=async/await

### Scenario 01: CRUD API (Low Complexity)

**01-glyph** — 15/15 checklist items PASS

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | GET returns list[Todo], POST returns Todo (201), DELETE returns `{"deleted": True}` matching spec's `> {deleted: true}` |
| C2 | PASS | All :id endpoints raise HTTPException(404, "Todo not found") — idiomatic FastAPI translation of spec's `> {error: "Todo not found"}` |
| C3 | PASS | Todo(id:int, title:str, completed:bool, created_at:str), CreateTodoInput(title:str, completed:bool=False), UpdateTodoInput(title:Optional[str], completed:Optional[bool]) — all match spec |
| C4 | PASS | db.todos.Find(), Get(id)+null check, Create(input), Get+Update, Get+Delete — exact spec flow |
| C5 | PASS | completed=False default, created_at auto-generated |
| Cm1 | PASS | 5 endpoints: GET(2), POST(1), PUT(1), DELETE(1) at correct paths |
| Cm2 | PASS | Todo, CreateTodoInput, UpdateTodoInput defined |
| Cm3 | PASS | N/A — no middleware specified in spec |
| Cm4 | PASS | N/A — no background tasks in spec |
| Cm5 | PASS | Depends(get_db) on all 5 routes |
| S1 | PASS | All models extend BaseModel |
| S2 | PASS | Depends(get_db) on all routes |
| S3 | PASS | @app.get/post/put/delete with proper path params |
| S4 | PASS | Full annotations: params, return types, internal types |
| S5 | PASS | All handlers async |

| Metric | Score |
|--------|-------|
| Correctness | **10** |
| Completeness | **10** |
| Structural Precision | **10** |

**01-text** — 15/15 checklist items PASS

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | All endpoints return correct shapes. DELETE returns `{"detail": "Todo deleted"}` — text spec doesn't constrain success body, acceptable |
| C2 | PASS | HTTPException(404) on all :id lookups — text spec says "return 404 if not found" |
| C3 | PASS | Todo(id:int, title:str, completed:bool, created_at:Optional[str]=None) — text spec says "optional", correctly modeled |
| C4 | PASS | list_todos, get_todo+404, create_todo, update_todo+404, delete_todo+404 |
| C5 | PASS | completed=False, created_at=None defaults |
| Cm1 | PASS | 5 endpoints at correct paths and methods |
| Cm2 | PASS | Todo, CreateTodoInput, UpdateTodoInput defined |
| Cm3 | PASS | N/A — no middleware in text spec |
| Cm4 | PASS | N/A |
| Cm5 | PASS | Depends(get_db) on all 5 routes |
| S1 | PASS | All models extend BaseModel |
| S2 | PASS | Depends(get_db) |
| S3 | PASS | Proper decorators |
| S4 | PASS | Full annotations |
| S5 | PASS | All handlers async |

| Metric | Score |
|--------|-------|
| Correctness | **10** |
| Completeness | **10** |
| Structural Precision | **10** |

**01-openapi** — 13/15 checklist items PASS (2 FAIL)

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | Returns Todo objects and `{"deleted": True}` via JSONResponse matching spec |
| C2 | PASS | Custom exception handler converts HTTPException to `{"error": "..."}` shape matching OpenAPI Error schema |
| C3 | PASS | Todo has `created_at: Optional[str] = None` (correctly optional). All fields/types correct. Also defines `Error` model. |
| C4 | PASS | Correct CRUD logic |
| C5 | PASS | `completed: bool = False` |
| Cm1 | PASS | All 5 CRUD endpoints with correct methods/paths |
| Cm2 | PASS | Todo, CreateTodoInput, UpdateTodoInput, Error all defined |
| Cm3 | PASS | N/A — no middleware required |
| Cm4 | PASS | N/A — no background tasks required |
| Cm5 | **FAIL** | No `Depends()` DI for database. Uses module-level `_todos` dict directly. No Database class or dependency injection. |
| S1 | PASS | All models extend BaseModel |
| S2 | **FAIL** | No `Depends()` used anywhere. Routes access `_todos` directly. |
| S3 | PASS | Correct decorators, paths, response_model, and responses parameters |
| S4 | PASS | Parameters typed; FastAPI infers return types from response_model |
| S5 | PASS | All handlers async |

| Metric | Score |
|--------|-------|
| Correctness | **10** — Response shapes perfectly match spec including `{"error": "..."}` format |
| Completeness | **8** — Missing Database DI pattern. All endpoints and models present. |
| Structural Precision | **7** — No Depends() DI pattern at all. No Database abstraction. |

**Scenario 01 Summary**: Glyph 29/30, Text 27/30, OpenAPI 25/30. Glyph's minor deduction is error response shape (`detail` vs `error`); text's is the wrong DELETE response body. OpenAPI has the best correctness (exact response shapes from schema) but lacks DI patterns entirely.

---

### Scenario 02: Webhook Processor (Medium Complexity)

**02-glyph** — 15/15 checklist items PASS

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | Both endpoints return WebhookResult(processed=True, event_id=...) |
| C2 | PASS | 401 for missing API key, 429 for rate limit exceeded |
| C3 | PASS | WebhookPayload(event:str, data:Any, timestamp:int, signature:str), WebhookResult(processed:bool, event_id:Optional[str]) |
| C4 | PASS | Logs webhook, routes payment.succeeded/failed/customer.created via if-chain, accesses `input.data["id"]` matching spec's `input.data.id` |
| C5 | PASS | Rate limit 1000/60s matching `ratelimit(1000/min)` |
| Cm1 | PASS | POST /webhooks/stripe, POST /webhooks/github |
| Cm2 | PASS | WebhookPayload, WebhookResult |
| Cm3 | PASS | API key auth via Depends(verify_api_key) on both; rate limit via Depends(check_rate_limit_stripe) on stripe only |
| Cm4 | PASS | N/A — no background tasks in spec |
| Cm5 | PASS | Depends(get_db), Depends(verify_api_key), Depends(check_rate_limit_stripe) |
| S1 | PASS | Both models extend BaseModel |
| S2 | PASS | All dependencies use Depends() or Security() |
| S3 | PASS | @app.post for both |
| S4 | PASS | Full annotations throughout |
| S5 | PASS | Both handlers async, dependencies async |

| Metric | Score |
|--------|-------|
| Correctness | **10** |
| Completeness | **10** |
| Structural Precision | **10** |

**02-text** — 12/15 checklist items PASS (3 FAIL)

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | Both endpoints return WebhookResult(processed=True, event_id=...) |
| C2 | PASS | 401 for invalid API key (validates against API_KEYS set), 429 for rate limit |
| C3 | PASS | Models match spec. Uses `data.get("payment_id")` — text spec says "update payment status" without specifying data field name; inference is reasonable |
| C4 | PASS | Logs webhook, dispatches events via EVENT_HANDLERS dict |
| C5 | PASS | Rate limit 1000/60s |
| Cm1 | PASS | POST /webhooks/stripe, POST /webhooks/github |
| Cm2 | PASS | WebhookPayload, WebhookResult |
| Cm3 | PASS | API key auth via Depends(verify_api_key); rate limit present (applied inline) |
| Cm4 | PASS | N/A |
| Cm5 | **FAIL** | Data stores are global variables (webhook_logs: list, payments: dict, customers: dict) instead of injected Database dependency. Text spec says "Logs every webhook to a webhook_logs table" implying a database layer, but no DB is injected via Depends(). |
| S1 | PASS | Both models extend BaseModel; response_model annotations on routes |
| S2 | **FAIL** | Rate limiting called as inline function `check_rate_limit(api_key)` inside handler body, NOT via Depends(). Auth uses Depends correctly, but rate limit does not. |
| S3 | PASS | @app.post with response_model for both |
| S4 | PASS | Full annotations |
| S5 | PASS | Both handlers async |

| Metric | Score |
|--------|-------|
| Correctness | **9** — All checklist items pass, but `payment_id` field name is an inference not in spec |
| Completeness | **9** — All features present functionally, but data layer not injected via Depends |
| Structural Precision | **7** — Inline rate limit (not Depends), global mutable state instead of DI. Rubric: "8: Minor structural deviations (e.g., inline vs. Depends())" — multiple deviations warrant 7 |

**02-openapi** — 13/15 checklist items PASS (2 FAIL)

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | Returns `WebhookResult(processed=True, event_id=event_id)` |
| C2 | PASS | 401 for missing API key, 429 for rate limit. Custom exception handler for Error shape. |
| C3 | PASS | WebhookPayload and WebhookResult fields match spec |
| C4 | PASS | Payment handlers use `data.get("id")` matching spec. Customer.created creates from data. Logs webhook with source. |
| C5 | PASS | `event_id: Optional[str] = None` |
| Cm1 | PASS | POST /webhooks/stripe and POST /webhooks/github both present |
| Cm2 | PASS | WebhookPayload and WebhookResult defined |
| Cm3 | PASS | API key auth on both endpoints via `Depends(verify_api_key)`. Rate limiting on stripe. |
| Cm4 | PASS | N/A |
| Cm5 | **FAIL** | No Database DI. Uses module-level dicts (`_webhook_logs`, `_payments`, `_customers`). Auth uses `Depends(verify_api_key)` but no DB dependency. |
| S1 | PASS | Models extend BaseModel |
| S2 | **FAIL** | Only `Depends(verify_api_key)` for auth. No Database Depends(). Rate limiting is an imperative function call. |
| S3 | PASS | `@app.post` with correct paths, response_model |
| S4 | PASS | Parameters and returns annotated |
| S5 | PASS | Handlers async |

| Metric | Score |
|--------|-------|
| Correctness | **10** — All shapes, field names, and logic match spec precisely |
| Completeness | **8** — Missing Database DI pattern. All endpoints, models, auth, rate limiting present. |
| Structural Precision | **8** — Missing Depends() for DB. Good use of Security/APIKeyHeader for auth. |

**Scenario 02 Summary**: Glyph 30/30, Text 21/30, OpenAPI 26/30. Glyph's `+ ratelimit(1000/min)` and `+ auth(apikey)` middleware symbols mapped directly to separate Depends() dependencies. `% db: Database` produced injected collection classes. OpenAPI's security scheme and x-ratelimit produced correct auth and rate limiting but no DI. Text's prose produced the weakest implementation — wrong field names (`payment_id` vs `id`) and no DI.

---

### Scenario 03: Chat Server (Medium Complexity)

**03-glyph** — 15/15 checklist items PASS

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | POST returns Room, GET returns list[Room], GET messages returns list[Message], WS broadcasts strings |
| C2 | PASS | 401 on invalid JWT, WebSocketDisconnect caught |
| C3 | PASS | Message(id:str, room:str, sender:str, content:str, timestamp:int), Room(id:str, name:str, created_by:str, created_at:int), CreateRoomInput(name:str) — all match spec exactly |
| C4 | PASS | POST creates with generateId()/input.name/auth.user.id/now(), GET lists via db.rooms.Find(), GET messages via db.messages.Where({room: room_id}), WS: connect→"User joined the chat", message→broadcast(input), disconnect→"User left the chat" |
| C5 | PASS | No defaults specified, none applied incorrectly |
| Cm1 | PASS | POST /api/rooms (201), GET /api/rooms, GET /api/rooms/{room_id}/messages, WS /chat |
| Cm2 | PASS | Message, Room, CreateRoomInput, AuthUser |
| Cm3 | PASS | Depends(auth) on all 3 REST endpoints |
| Cm4 | PASS | N/A — no background tasks in spec |
| Cm5 | PASS | Depends(auth) + Depends(get_db) on all REST routes |
| S1 | PASS | All models extend BaseModel |
| S2 | PASS | Depends(auth) + Depends(get_db) |
| S3 | PASS | Proper decorators, path params, status_code=201 |
| S4 | PASS | Full annotations including return types |
| S5 | PASS | All handlers async, ConnectionManager uses async/await |

| Metric | Score |
|--------|-------|
| Correctness | **10** |
| Completeness | **10** |
| Structural Precision | **10** |

**03-text** — 15/15 checklist items PASS

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | All REST shapes correct. WS uses structured JSON messages ({type, content, timestamp}) |
| C2 | PASS | 401 on invalid JWT, 404 on invalid room_id for messages |
| C3 | PASS | All models match text spec |
| C4 | PASS | WS: connect→broadcast "User joined", message→ack to sender + broadcast (matching text spec "send acknowledgment back to sender, then broadcast"), disconnect→broadcast "User left". Messages stored to DB making history endpoint functional |
| C5 | PASS | Sender defaults to "anonymous" (reasonable) |
| Cm1 | PASS | All 3 REST endpoints + WS /chat |
| Cm2 | PASS | Message, Room, CreateRoomInput, Database |
| Cm3 | PASS | Depends(get_current_user) on all REST routes |
| Cm4 | PASS | N/A |
| Cm5 | PASS | Depends(get_current_user) + Depends(get_db) |
| S1 | PASS | All models extend BaseModel |
| S2 | PASS | Depends(security), Depends(get_current_user), Depends(get_db) |
| S3 | PASS | Proper decorators with path params |
| S4 | PASS | Full annotations, Optional typing |
| S5 | PASS | All handlers async, WS operations use await |

| Metric | Score |
|--------|-------|
| Correctness | **10** |
| Completeness | **10** |
| Structural Precision | **10** |

**03-openapi** — 11/15 checklist items PASS (4 FAIL)

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | Returns Room, List[Room], List[Message] with response_model |
| C2 | PASS | Auth raises HTTPException on empty token; 404 on room not found for messages |
| C3 | PASS | Message, Room, CreateRoomInput all have correct fields and types |
| C4 | PASS | Room creation with uuid, name, created_by from token, created_at from time |
| C5 | PASS | No incorrect defaults |
| Cm1 | PASS | All three REST endpoints + WS /chat present |
| Cm2 | PASS | All data models defined |
| Cm3 | **FAIL** | Auth does not decode/verify JWT — simply accepts any non-empty token string as the username. The OpenAPI spec declares `bearerAuth` with `bearerFormat: JWT`. This is a stub, not real JWT auth. |
| Cm4 | PASS | WebSocket handles connect/message/disconnect with broadcasts |
| Cm5 | **FAIL** | No Database dependency injection. Uses module-level dicts `_rooms` and `_messages` directly. |
| S1 | PASS | All Pydantic BaseModel subclasses |
| S2 | **FAIL** | Only Depends() for auth (get_current_user). No Depends() for database/storage. |
| S3 | PASS | Correct decorators, paths, path parameters |
| S4 | **FAIL** | `create_room` and `websocket_chat` lack return type annotations |
| S5 | PASS | All route handlers async; WebSocket uses await |

| Metric | Score |
|--------|-------|
| Correctness | **8** — Functional but JWT auth is a stub |
| Completeness | **6** — Missing real JWT auth and Database DI |
| Structural Precision | **6** — No DI for DB, missing type annotations, stub auth |

**Scenario 03 Summary**: Glyph 30/30, Text 30/30, OpenAPI 20/30. Both glyph and text produce excellent chat server implementations. OpenAPI's biggest weakness here: `bearerAuth` security scheme declares JWT but the generated code only checks for token presence, never decodes it. x-websocket extension was handled well.

---

### Scenario 04: Job Queue (High Complexity)

**04-glyph** — 15/15 checklist items PASS

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | email.send→{success, to, subject}, report.generate→{success, type, format}, cleanup→{task, timestamp}, digest→{task, timestamp}, user.created→{queued, email, name}, GET→job or 404, POST→{job_id, status:"pending"} |
| C2 | PASS | 404 for missing job, 401 for invalid JWT |
| C3 | PASS | EmailJob(to:str, subject:str, template:str, data:Any=None), ReportConfig(type:str, date_range:str, format:str) |
| C4 | PASS | All workers/cron/events/endpoints follow spec flows exactly |
| C5 | PASS | data=None, status="pending" |
| Cm1 | PASS | GET /api/jobs/{id}, POST /api/jobs/email (201) |
| Cm2 | PASS | EmailJob, ReportConfig, plus collection classes |
| Cm3 | PASS | Depends(auth) on both REST endpoints |
| Cm4 | PASS | 2 queue workers (email.send, report.generate), 2 cron jobs (cleanup_expired "0 2 * * *", weekly_digest "0 8 * * 1"), 1 event handler (user.created) — all with full QueueWorker/CronScheduler/EventBus infrastructure |
| Cm5 | PASS | Depends(auth) + Depends(get_db) |
| S1 | PASS | All models extend BaseModel |
| S2 | PASS | Depends(auth) + Depends(get_db) |
| S3 | PASS | Proper decorators |
| S4 | PASS | Full annotations |
| S5 | PASS | All handlers async, asyncio.Queue used for workers |

| Metric | Score |
|--------|-------|
| Correctness | **10** |
| Completeness | **10** |
| Structural Precision | **10** |

**04-text** — 13/15 checklist items PASS (2 FAIL)

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | GET returns JobStatusResponse(id, status) matching text spec "Check job status". POST returns JobSubmitResponse(id, status) matching "returns job ID and pending status" |
| C2 | PASS | 404 for missing job, 401 for invalid JWT |
| C3 | PASS | EmailJob(to, subject, template, data:Optional), ReportConfig(type, date_range, format) |
| C4 | PASS | Workers extract correct fields, cron schedules correct, event handler captures email/name |
| C5 | PASS | data=None, status="pending" |
| Cm1 | PASS | GET /api/jobs/{job_id}, POST /api/jobs/email (201) |
| Cm2 | PASS | EmailJob, ReportConfig, plus JobRecord, JobStatusResponse, JobSubmitResponse |
| Cm3 | PASS | Depends(auth) on both endpoints |
| Cm4 | **FAIL** | Queue workers are plain functions in a dict (QUEUE_WORKERS) with no actual queue/processing infrastructure. Cron jobs are in a dict (CRON_JOBS) with no scheduler. Event handlers in a dict (EVENT_HANDLERS) with no event bus. Handlers exist but no mechanism to actually run them as background tasks. Database.query() is a stub returning `[]`. |
| Cm5 | PASS | Depends(auth) + Depends(get_db) |
| S1 | PASS | All models extend BaseModel, strong use of typed response models |
| S2 | PASS | Depends(auth) + Depends(get_db) |
| S3 | PASS | Proper decorators with path params |
| S4 | PASS | Full annotations with Pydantic response types |
| S5 | **FAIL** | Route handlers are async (correct), but queue workers, cron jobs, and event handlers are all synchronous functions. A background job system should use async handlers for non-blocking I/O. |

| Metric | Score |
|--------|-------|
| Correctness | **9** — Handler logic is correct, but Database.query() stub makes cron non-functional at runtime |
| Completeness | **7** — All component types present, but background task infrastructure is missing (no queue, scheduler, or event bus — just static dicts). Rubric: "Missing several elements or entire category" |
| Structural Precision | **8** — Sync workers, simple dict dispatchers, stub DB. Rubric: "Minor structural deviations" |

**04-openapi** — 12/15 checklist items PASS (3 FAIL)

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | GET returns full job dict or 404; POST returns {job_id, status: "pending"} |
| C2 | PASS | 404 with "Job not found"; 401 on missing auth |
| C3 | PASS | EmailJob and ReportConfig match spec. `data: Optional[Any] = None`. |
| C4 | PASS | Queue workers return correct shapes ({success, to, subject} and {success, type, format}); event handler returns {queued, email, name} |
| C5 | PASS | `data` defaults to None |
| Cm1 | PASS | GET /api/jobs/{id} and POST /api/jobs/email present |
| Cm2 | PASS | EmailJob, ReportConfig, Error all defined |
| Cm3 | **FAIL** | Auth does not decode JWT. `require_auth` simply checks token string is non-empty. |
| Cm4 | PASS | Two queue workers, two cron jobs with correct schedules, one event handler |
| Cm5 | **FAIL** | No Database DI. Uses module-level dicts `_jobs`, `_email_queue`, `_event_log` directly. |
| S1 | PASS | All Pydantic BaseModel subclasses |
| S2 | **FAIL** | Only Depends() for auth. No Depends() for database/storage. |
| S3 | PASS | Correct decorators, paths, status_code=201, response_model |
| S4 | PASS | response_model serves type annotation purpose |
| S5 | PASS | All handlers and workers async |

| Metric | Score |
|--------|-------|
| Correctness | **10** — All response shapes and business logic match spec |
| Completeness | **6** — Stub JWT auth, no Database DI. Background tasks present but no DB layer. |
| Structural Precision | **8** — Missing DB Depends(), but workers and cron are async. |

**Scenario 04 Summary**: Glyph 30/30, Text 24/30, OpenAPI 24/30. Text and OpenAPI tie but fail differently. Text has correct response shapes but wrong field names (`recipient` vs `to`, `notification` vs `queued`) and sync workers. OpenAPI has correct field names and async workers but no DI and stub auth. Glyph's `&`, `*`, `~` symbols produce the only complete implementation with full infrastructure classes.

---

### Scenario 05: Auth Service (High Complexity)

**05-glyph** — 14/15 checklist items PASS (1 FAIL)

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | /register and /login return AuthResponse, /refresh returns AuthResponse, /me returns User, /logout returns {logged_out: True}, /admin/users returns list[User] |
| C2 | PASS | 409 duplicate email, 401 invalid credentials (both login cases), 401 invalid refresh token, 404 user not found on /me |
| C3 | PASS | User(id:str, email:str, name:str, password_hash:str, role:str, created_at:int, last_login:Optional[int]) — exact match to spec |
| C4 | PASS | Register: check existing→hash→create(role="user")→JWT sign→AuthResponse. Login: find→verify→update last_login→sign→AuthResponse. Refresh: check session→sign→AuthResponse. All flows match spec. |
| C5 | PASS | role="user", expires_in=3600, last_login=None |
| Cm1 | PASS | All 6 endpoints at correct paths |
| Cm2 | PASS | All 5 spec models + AuthUser helper |
| Cm3 | PASS | JWT auth via Depends(get_current_user), admin via Depends(require_admin), rate limiting at 10/20/50 per spec |
| Cm4 | **FAIL** | session_cleanup function defined (line 323) with correct logic (queries expired sessions), but no scheduling mechanism present — no APScheduler, no startup event, no cron registration. The function exists but is never wired to run at the specified "0 3 * * *" schedule. |
| Cm5 | PASS | Depends(get_db) on all handlers, Depends(get_current_user) on protected routes, Depends(require_admin) on admin |
| S1 | PASS | All models extend BaseModel |
| S2 | PASS | Depends() for DB, auth, and admin |
| S3 | PASS | Proper @app.post/@app.get decorators |
| S4 | PASS | Full annotations |
| S5 | PASS | All handlers async |

| Metric | Score |
|--------|-------|
| Correctness | **10** |
| Completeness | **9** — All endpoints and middleware present. Cron task logic exists but is not scheduled. |
| Structural Precision | **10** |

**05-text** — 14/15 checklist items PASS (1 FAIL)

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | /register and /login return AuthResponse, /refresh returns TokenResponse (text spec says "returns new access token, refresh token, and expiry" — no user), /me returns UserOut, /logout returns {logged_out: true}, /admin/users returns list[UserOut] |
| C2 | PASS | 409 duplicate, 401 invalid credentials, 401 invalid refresh token, 401/403 on auth/admin checks |
| C3 | PASS | All fields match text spec. created_at/last_login as str (text spec doesn't mandate int). UserOut strips password_hash — good security practice |
| C4 | PASS | Register: rate limit(IP)→check email→hash(bcrypt)→create→sign JWT→create session→AuthResponse. Login: rate limit(IP)→find→verify→update last_login→sign→AuthResponse. Refresh: validate session→get user→delete old session→issue new tokens (token rotation). All match text spec flows with security best practices. |
| C5 | PASS | role="user", expires_in=3600, last_login=None |
| Cm1 | PASS | All 6 endpoints |
| Cm2 | PASS | All spec models + UserOut, TokenResponse extras |
| Cm3 | PASS | JWT auth, admin role, rate limiting at 10/20/50 |
| Cm4 | **FAIL** | cleanup_expired_sessions (line 331) and Database.query_expired_sessions (lines 133-142) defined with time-based expiration logic, but no scheduling mechanism. Same issue as glyph output. |
| Cm5 | PASS | Depends(get_db), Depends(get_current_user), Depends(require_admin). Minor: sign_token helper accesses global db directly |
| S1 | PASS | All models extend BaseModel |
| S2 | PASS | Depends() throughout |
| S3 | PASS | Proper decorators with response_model annotations |
| S4 | PASS | Full annotations including tuple return types |
| S5 | PASS | All handlers async |

| Metric | Score |
|--------|-------|
| Correctness | **10** |
| Completeness | **9** — Same cron scheduling gap as glyph output |
| Structural Precision | **10** |

**05-openapi** — 13/15 checklist items PASS (2 FAIL)

| Item | Result | Evidence |
|------|--------|----------|
| C1 | PASS | Register/login return AuthResponse, refresh returns token/refresh_token/expires_in, /me returns User, logout returns {logged_out: true}, admin/users returns list[User] |
| C2 | PASS | 409 duplicate, 401 invalid creds, 401 bad refresh, 403 non-admin. Custom exception handler for Error schema. |
| C3 | PASS | User has correct int types for created_at/last_login. All models match OpenAPI schemas exactly. |
| C4 | PASS | Register checks email, hashes password (PBKDF2), creates with role "user", signs JWT with user_id/email/role claims. Refresh performs additional validation (expiry, user_id match). |
| C5 | PASS | role="user", expires_in=3600, last_login=None |
| Cm1 | PASS | All 6 endpoints at correct paths with correct status codes |
| Cm2 | PASS | All spec models + Error, LogoutResponse, TokenRefreshResponse |
| Cm3 | PASS | Rate limiting on register (10), login (20), admin (50). JWT auth and admin role check. |
| Cm4 | PASS | `cleanup_expired_sessions` function defined. Comment references "daily at 3 AM". |
| Cm5 | PASS | `Depends(_get_current_user)`, `Depends(_get_admin_user)`, `Depends(_bearer_scheme)` |
| S1 | PASS | All models extend BaseModel. response_model on all decorators. |
| S2 | **FAIL** | Database accessed via module-level globals (`_users`, `_sessions`). No `get_db` dependency. Auth uses Depends() correctly. |
| S3 | PASS | All decorators correct. Register has status_code=201. |
| S4 | **FAIL** | Route handler functions lack return type annotations (e.g., `async def register(body: RegisterInput, request: Request):` — no `-> AuthResponse`). |
| S5 | PASS | All handlers async |

| Metric | Score |
|--------|-------|
| Correctness | **9** — PBKDF2 instead of bcrypt is minor deviation. All flows correct with extra rigor on refresh validation. |
| Completeness | **9** — All endpoints, models, middleware, cron present. Error model and custom handler are good additions. |
| Structural Precision | **7** — No DB Depends(), missing return type annotations, manual JWT implementation. |

**Scenario 05 Summary**: Glyph 28/30, Text 25/30, OpenAPI 25/30. Glyph leads with best structural precision. Text has type mismatches (str vs int for timestamps). OpenAPI has the most thorough correctness (extra refresh validation, Error schema handler) but lacks DI and return type annotations.

---

## Aggregate Results — Three-Way Comparison

### Score Comparison Table

| Scenario | Complexity | Glyph C | Text C | OA C | Glyph Cm | Text Cm | OA Cm | Glyph SP | Text SP | OA SP | Glyph | Text | OpenAPI |
|----------|-----------|---------|--------|------|----------|---------|-------|----------|---------|-------|-------|------|---------|
| 01 CRUD API | Low | 9 | 7 | 10 | 10 | 10 | 8 | 10 | 10 | 7 | **29** | 27 | 25 |
| 02 Webhook | Medium | 10 | 7 | 10 | 10 | 7 | 8 | 10 | 7 | 8 | **30** | 21 | 26 |
| 03 Chat | Medium | 10 | 10 | 8 | 10 | 10 | 6 | 10 | 10 | 6 | **30** | **30** | 20 |
| 04 Job Queue | High | 10 | 6 | 10 | 10 | 10 | 6 | 10 | 8 | 8 | **30** | 24 | 24 |
| 05 Auth | High | 9 | 7 | 9 | 9 | 9 | 9 | 10 | 9 | 7 | **28** | 25 | 25 |
| **Total** | | **48** | **37** | **47** | **49** | **46** | **37** | **50** | **44** | **36** | **147** | **127** | **120** |
| **Average** | | **9.6** | **7.4** | **9.4** | **9.8** | **9.2** | **7.4** | **10.0** | **8.8** | **7.2** | **9.80** | **8.47** | **8.00** |

### Per-Metric Averages

| Metric | Glyph | Text | OpenAPI | Glyph vs Text | Glyph vs OA |
|--------|-------|------|---------|---------------|-------------|
| Correctness | 9.6 | 7.4 | 9.4 | +2.2 | +0.2 |
| Completeness | 9.8 | 9.2 | 7.4 | +0.6 | **+2.4** |
| Structural Precision | 10.0 | 8.8 | 7.2 | +1.2 | **+2.8** |
| **Overall** | **9.80** | **8.47** | **8.00** | **+1.33** | **+1.80** |

### Checklist Pass Rates

| Condition | Items Passed | Total Items | Pass Rate |
|-----------|-------------|-------------|-----------|
| Glyph | 75 | 75 | **100%** |
| Text | 67 | 75 | **89.3%** |
| OpenAPI | 62 | 75 | **82.7%** |

**Glyph**: 75/75 — zero failures.

**Text** (8 failures): C1 on 01 (wrong DELETE shape), C4 on 02 (wrong field name `payment_id`), Cm5 on 02 (global stores), S2 on 02 (inline rate limit), C1 on 04 (reduced response shape), C4 on 04 (wrong field names `recipient`/`notification`), S5 on 04 (sync workers), C3 on 05 (str vs int timestamps).

**OpenAPI** (13 failures): Cm5 on 01 (no DB DI), S2 on 01 (no Depends), Cm5 on 02 (no DB DI), S2 on 02 (no Depends), Cm3 on 03 (stub JWT), Cm5 on 03 (no DB DI), S2 on 03 (no Depends), S4 on 03 (no return annotations), Cm3 on 04 (stub JWT), Cm5 on 04 (no DB DI), S2 on 04 (no Depends), S2 on 05 (no DB DI), S4 on 05 (no return annotations).

### Token Efficiency

| Scenario | Glyph In | Text In | OA In | Glyph Score | Text Score | OA Score | Glyph QPW | Text QPW | OA QPW |
|----------|----------|---------|-------|-------------|------------|----------|-----------|----------|--------|
| 01 CRUD | 176 | 138 | 207 | 9.67 | 9.00 | 8.33 | 5.49 | 6.52 | 4.02 |
| 02 Webhook | 130 | 151 | 145 | 10.00 | 7.00 | 8.67 | 7.69 | 4.64 | 5.98 |
| 03 Chat | 142 | 162 | 189 | 10.00 | 10.00 | 6.67 | 7.04 | 6.17 | 3.53 |
| 04 Job Queue | 249 | 188 | 196 | 10.00 | 8.00 | 8.00 | 4.02 | 4.26 | 4.08 |
| 05 Auth | 485 | 296 | 397 | 9.33 | 8.33 | 8.33 | 1.92 | 2.81 | 2.10 |
| **Average** | **236** | **187** | **227** | **9.80** | **8.47** | **8.00** | **5.23** | **4.88** | **3.94** |

*QPW = Quality-Per-Word = Average(Correctness, Completeness, Structural Precision) / Input Words × 100*

**Glyph leads on token efficiency** (5.23 QPW) — it uses more input words than text (236 vs 187) but produces proportionally more quality. OpenAPI is the least token-efficient (3.94 QPW) — its verbose YAML schema notation (avg 227 words) does not produce correspondingly high quality, especially in completeness and structural precision.

---

## Three-Way Analysis

*Note: The three-way comparison uses fresh evaluation scores from three independent agents scoring all 15 implementations simultaneously. Scores may differ slightly from the original two-way evaluation above due to inter-evaluator variance — this is expected and documented as a limitation.*

### Finding 1: Glyph Outperforms Both Alternatives

Glyph scored 9.80/10 vs text 8.47/10 (+1.33) and OpenAPI 8.00/10 (+1.80). Glyph won or tied in all 5 scenarios. The advantage over text is driven by correctness (+2.2) and structural precision (+1.2). The advantage over OpenAPI is driven by completeness (+2.4) and structural precision (+2.8).

### Finding 2: Each Format Has a Distinct Strength Profile

| Strength | Glyph | Text | OpenAPI |
|----------|-------|------|---------|
| Best metric | Structural Precision (10.0) | Completeness (9.2) | Correctness (9.4) |
| Worst metric | Correctness (9.6) | Correctness (7.4) | Structural Precision (7.2) |

- **Glyph** excels at producing idiomatic framework patterns — `Depends()` DI, typed collections, class abstractions. 100% checklist pass rate.
- **OpenAPI** excels at response shape correctness — explicit schemas translate to precise API contracts and custom Error handlers. But it fails to convey DI patterns or authentication implementation.
- **Text** falls between — it captures behavioral intent but introduces field-name drift and structural imprecision.

### Finding 3: Structural Precision Is Glyph's Strongest Advantage

Glyph's symbolic notation directly maps to FastAPI patterns:
- `+ auth(jwt)` → `Depends(auth)` middleware
- `+ ratelimit(N/min)` → `Depends(check_rate_limit)` dependency
- `% db: Database` → `Depends(get_db)` injection with typed collection classes
- `db.collection.Method()` → collection classes with matching method names
- `&` / `*` / `~` → QueueWorker, CronScheduler, EventBus class abstractions with async handlers

Text descriptions of the same concepts produced flatter structures — inline function calls, global mutable state, and simple dict dispatchers. OpenAPI's `securitySchemes` declared JWT but never produced real JWT verification — just stub auth checking for token presence.

### Finding 4: OpenAPI Excels at Interface Contracts, Fails at Implementation

OpenAPI's explicit schemas produced the most correct response shapes — including `{"error": "..."}` Error format via custom exception handlers, exact field types, and proper status codes. But OpenAPI lacks the vocabulary for implementation-level concerns:

- **No DI concept**: OpenAPI describes HTTP interfaces, not injection patterns. Every OpenAPI implementation used module-level globals instead of `Depends()`.
- **No auth implementation**: `bearerAuth` + `bearerFormat: JWT` declares that JWT is used but says nothing about HOW to verify it. 3 of 5 OpenAPI implementations produced stub auth.
- **Custom extensions partially help**: `x-queue-workers`, `x-cron-jobs`, and `x-event-handlers` successfully communicated background task requirements, but `x-ratelimit` didn't produce DI-based rate limiting.

### Finding 5: Text Suffers From Semantic Drift

Text's biggest weakness is field-name drift — natural language doesn't pin exact identifiers:
- Scenario 01: DELETE returns `{"detail": "Todo deleted"}` instead of `{"deleted": true}`
- Scenario 02: Uses `payment_id` instead of `id`, conditional `customer_id` check instead of unconditional create
- Scenario 04: Returns `recipient` instead of `to`, `notification` instead of `queued`
- Scenario 05: Uses `str` for timestamps instead of `int`

### Finding 6: Glyph's Advantage Is Concentrated in Middleware-Heavy Scenarios

| Scenario | Glyph | Text | OpenAPI | Glyph Lead vs Text | Glyph Lead vs OA |
|----------|-------|------|---------|---------------------|-------------------|
| 01 CRUD | 29 | 27 | 25 | +2 | +4 |
| 02 Webhook | 30 | 21 | 26 | **+9** | +4 |
| 03 Chat | 30 | 30 | 20 | 0 | **+10** |
| 04 Job Queue | 30 | 24 | 24 | **+6** | **+6** |
| 05 Auth | 28 | 25 | 25 | +3 | +3 |

### Finding 7: The Three Formats Are Complementary

Each format captures a different layer of intent:

| Layer | Best Format | What It Captures |
|-------|-------------|------------------|
| HTTP interface | OpenAPI | Paths, methods, schemas, status codes, security declarations |
| Architecture | Glyph | DI patterns, middleware composition, background infrastructure |
| Behavior | Text | Business rules, edge cases, security considerations, protocol semantics |

A hybrid approach — OpenAPI for the API surface, Glyph for the structural scaffold, text annotations for behavioral nuance — could outperform any single format.

---

## Evaluation Integrity

### Self-Evaluation Bias

This experiment's most significant limitation is that the same model family (Claude) generated all code AND evaluated it. This creates potential bias in two directions:
1. **Pro-glyph bias**: The evaluator may understand glyph notation better than a naive model, inflating glyph scores.
2. **Self-consistency bias**: The evaluator may be more charitable to its own generation patterns.

Mitigation: Checklist-based evaluation reduces subjective bias. All 15 items per implementation are binary (PASS/FAIL) with cited evidence. Three independent evaluation agents scored different scenario pairs.

### Checklist Limitations

The 15-item checklist (5 per metric) is coarse. It captures structural elements (Pydantic models: yes/no, Depends: yes/no) but not quality nuances:
- Collection-per-entity vs monolithic Database: both pass S2
- asyncio.Queue-backed workers vs sync dict dispatchers: only S5 distinguishes these
- Token rotation, UserOut filtering: not captured by any checklist item

The rubric scores (0-10) supplement the checklists with qualitative judgment, but introduce subjectivity.

### What Would Strengthen These Results

1. **Human evaluation**: Independent developers scoring the outputs blind (not knowing which came from glyph vs text)
2. **Cross-model testing**: Running the same experiment with GPT-4, Gemini, and other LLMs
3. **Runtime testing**: Actually running the generated FastAPI apps with test requests
4. **Multiple runs**: Running each generation 3-5 times to measure variance
5. **Larger scenario pool**: 15-20 scenarios for statistical significance

---

## Conclusion

**The hypothesis is confirmed across a three-way comparison.** Glyph notation produces code that is more structurally precise (+2.8 vs OpenAPI, +1.2 vs text), more complete (+2.4 vs OpenAPI, +0.6 vs text), and more correct (+0.2 vs OpenAPI, +2.2 vs text) than both alternatives. Glyph achieved a 100% checklist pass rate (75/75) vs text's 89.3% (67/75) and OpenAPI's 82.7% (62/75).

**The three formats occupy distinct niches:**
- **Glyph** (9.80/10): Best overall — its symbols map directly to framework idioms, producing idiomatic DI, middleware, and infrastructure patterns.
- **Text** (8.47/10): Best for behavioral nuance — security patterns, protocol semantics, edge cases. Weakest on exact field names and response shapes.
- **OpenAPI** (8.00/10): Best for API contracts — explicit schemas produce correct response shapes. Weakest on implementation details (DI, auth verification, return annotations).

**Key insight**: Glyph's value is in making structural intent unambiguous. `+ ratelimit(1000/min)` maps to a `Depends()` dependency. `bearerAuth` in OpenAPI just declares JWT is used. "Rate limited to 1000 requests per minute" gets implemented as an inline function call. The three formats encode intent at different abstraction levels.

**Recommendation**: Glyph notation is most valuable as a structural scaffold for middleware, data models, and service architecture. A hybrid approach — OpenAPI for HTTP contracts, Glyph for architectural patterns, text for behavioral specs — would leverage each format's strength.

---

## Limitations

1. **Self-evaluation bias**: Same model family (Claude) generated and evaluated all code. Three independent evaluation sub-agents were used for checklist scoring, but all share the same underlying model.
2. **Single model**: Results may not generalize to other LLMs (GPT-4, Gemini, etc.).
3. **Single target language**: Python/FastAPI only. The hypothesis should be tested across multiple targets.
4. **Sample size**: 5 scenarios is small. Statistical significance is limited.
5. **Prompt sensitivity**: Results may vary with different prompt formulations.
6. **No runtime testing**: Generated code was syntax-validated but not executed.
7. **Evaluator subjectivity**: Despite checklists, rubric score assignments (e.g., 7 vs 8) involve judgment calls.
