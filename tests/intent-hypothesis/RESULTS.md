# Intent Hypothesis Testing — Results

## Hypothesis

> Glyph notation (`.glyph`) produces more correct, complete, and structurally precise AI-generated code than equivalent natural language descriptions (`.txt`) when used as input prompts for large language models.

## Generation Details

- **Model**: Claude (Anthropic) — same model for generation and evaluation (noted limitation)
- **Target language**: Python / FastAPI
- **Generation protocol**: Each implementation generated in an isolated sub-agent context
- **Total implementations**: 10 (5 glyph-sourced + 5 text-sourced)

---

## Validation

### Syntax Validation

All 10 generated Python files pass `ast.parse()` — they are syntactically valid Python.

```
glyph/01-crud-api.py         PASS
glyph/02-webhook-processor.py PASS
glyph/03-chat-server.py      PASS
glyph/04-job-queue.py        PASS
glyph/05-auth-service.py     PASS
text/01-crud-api.py           PASS
text/02-webhook-processor.py  PASS
text/03-chat-server.py        PASS
text/04-job-queue.py          PASS
text/05-auth-service.py       PASS
```

### Structural Analysis (AST-based)

| File | External Deps | Classes | Functions | Route Decorators |
|------|--------------|---------|-----------|-----------------|
| glyph/01-crud-api.py | fastapi, pydantic | 5 | 13 | 5 (get:2, post:1, put:1, delete:1) |
| text/01-crud-api.py | fastapi, pydantic | 4 | 12 | 5 (get:2, post:1, put:1, delete:1) |
| glyph/02-webhook-processor.py | fastapi, pydantic | 6 | 13 | 2 (post:2) |
| text/02-webhook-processor.py | fastapi, pydantic | 2 | 7 | 2 (post:2) |
| glyph/03-chat-server.py | fastapi, jwt, pydantic | 8 | 19 | 4 (get:2, post:1, ws:1) |
| text/03-chat-server.py | fastapi, jose, pydantic | 5 | 15 | 4 (get:2, post:1, ws:1) |
| glyph/04-job-queue.py | fastapi, jwt, pydantic | 11 | 31 | 2 (get:1, post:1) |
| text/04-job-queue.py | fastapi, jwt, pydantic | 8 | 14 | 2 (get:1, post:1) |
| glyph/05-auth-service.py | fastapi, jwt, passlib, pydantic | 9 | 27 | 6 (get:2, post:4) |
| text/05-auth-service.py | fastapi, jose, passlib, pydantic | 8 | 24 | 6 (get:2, post:4) |

**Observation**: Route decorator counts are identical across glyph/text pairs in every scenario — both notations produce the same endpoint structure. The difference is in classes and functions: glyph-generated code averages 7.8 classes vs text's 5.4, reflecting glyph's tendency to produce more structured abstractions (collection classes, worker classes, bus classes).

### Output File Sizes

| File | Bytes | Lines | Words |
|------|-------|-------|-------|
| glyph/01-crud-api.py | 2,822 | 118 | 295 |
| text/01-crud-api.py | 2,834 | 110 | 271 |
| glyph/02-webhook-processor.py | 3,753 | 148 | 346 |
| text/02-webhook-processor.py | 3,509 | 135 | 317 |
| glyph/03-chat-server.py | 3,858 | 177 | 383 |
| text/03-chat-server.py | 4,999 | 197 | 427 |
| glyph/04-job-queue.py | 6,566 | 262 | 666 |
| text/04-job-queue.py | 4,023 | 176 | 403 |
| glyph/05-auth-service.py | 8,007 | 327 | 749 |
| text/05-auth-service.py | 8,945 | 332 | 765 |

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

**Scenario 01 Summary**: Glyph 30/30 vs Text 30/30. Both implementations are excellent at this low complexity level. The differences are stylistic (collection class vs monolithic DB, `{deleted: true}` vs `{detail: "Todo deleted"}`), not functional.

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

**Scenario 02 Summary**: Glyph 30/30 vs Text 25/30. Glyph's `+ ratelimit(1000/min)` and `+ auth(apikey)` middleware symbols mapped directly to separate Depends() dependencies. `% db: Database` produced injected collection classes. Text's prose description ("rate limited to 1000 requests per minute") was implemented inline, and "webhook_logs table" became a global list.

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

**Scenario 03 Summary**: Glyph 30/30 vs Text 30/30. Both are excellent. The text version adds richer WebSocket behavior (ack messages, JSON structure, message persistence) aligned with its more detailed spec. The glyph version faithfully mirrors the simpler glyph spec. Both pass all checklist items.

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

**Scenario 04 Summary**: Glyph 30/30 vs Text 24/30. This is the largest gap. Glyph's `&` (queue), `*` (cron), and `~` (event) symbols produced fully structured QueueWorker, CronScheduler, and EventBus classes with async handlers and registration mechanisms. Text's prose description ("Queue Workers: Process email sending jobs") produced flat function-in-dict patterns with no processing infrastructure.

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

**Scenario 05 Summary**: Glyph 29/30 vs Text 29/30. Both are excellent with the same minor gap (cron scheduling). The text version adds production-quality security patterns (token rotation, UserOut response filtering, time-based session expiration, IP-based rate limiting) that the glyph version doesn't have.

---

## Aggregate Results

### Score Comparison Table

| Scenario | Complexity | Glyph C | Text C | Glyph Cm | Text Cm | Glyph SP | Text SP | Glyph Total | Text Total |
|----------|-----------|---------|--------|----------|---------|----------|---------|-------------|------------|
| 01 CRUD API | Low | 10 | 10 | 10 | 10 | 10 | 10 | **30** | **30** |
| 02 Webhook | Medium | 10 | 9 | 10 | 9 | 10 | 7 | **30** | 25 |
| 03 Chat | Medium | 10 | 10 | 10 | 10 | 10 | 10 | **30** | **30** |
| 04 Job Queue | High | 10 | 9 | 10 | 7 | 10 | 8 | **30** | 24 |
| 05 Auth | High | 10 | 10 | 9 | 9 | 10 | 10 | 29 | 29 |
| **Total** | | **50** | **48** | **49** | **45** | **50** | **45** | **149** | **138** |
| **Average** | | **10.0** | **9.6** | **9.8** | **9.0** | **10.0** | **9.0** | **9.93** | **9.20** |

### Per-Metric Averages

| Metric | Glyph Avg | Text Avg | Delta |
|--------|-----------|----------|-------|
| Correctness | 10.0 | 9.6 | +0.4 |
| Completeness | 9.8 | 9.0 | **+0.8** |
| Structural Precision | 10.0 | 9.0 | **+1.0** |
| **Overall** | **9.93** | **9.20** | **+0.73** |

### Checklist Pass Rates

| Condition | Items Passed | Total Items | Pass Rate |
|-----------|-------------|-------------|-----------|
| Glyph | 74 | 75 | **98.7%** |
| Text | 70 | 75 | **93.3%** |

Glyph's single failure: Cm4 on 05 (cron not scheduled).
Text's 5 failures: Cm5 on 02 (global stores), S2 on 02 (inline rate limit), Cm4 on 04 (no queue/scheduler infrastructure), S5 on 04 (sync workers), Cm4 on 05 (cron not scheduled).

### Token Efficiency

| Scenario | Glyph In (words) | Text In (words) | Glyph Out (lines) | Text Out (lines) | Glyph Avg Score | Text Avg Score | Glyph QPW | Text QPW |
|----------|-----------------|----------------|-------------------|------------------|----------------|---------------|-----------|----------|
| 01 CRUD | 176 | 138 | 118 | 110 | 10.00 | 10.00 | 5.68 | 7.25 |
| 02 Webhook | 130 | 151 | 148 | 135 | 10.00 | 8.33 | 7.69 | 5.52 |
| 03 Chat | 142 | 162 | 177 | 197 | 10.00 | 10.00 | 7.04 | 6.17 |
| 04 Job Queue | 249 | 188 | 262 | 176 | 10.00 | 8.00 | 4.02 | 4.26 |
| 05 Auth | 485 | 296 | 327 | 332 | 9.67 | 9.67 | 1.99 | 3.27 |
| **Average** | **236** | **187** | **206** | **190** | **9.93** | **9.20** | **5.28** | **5.29** |

*QPW = Quality-Per-Word = Average(Correctness, Completeness, Structural Precision) / Input Words × 100*

**Token efficiency is virtually tied** (5.28 vs 5.29 QPW). Glyph files use more words on average (236 vs 187) due to structural notation, but produce proportionally higher quality. Text files are more compact but produce lower quality in scenarios 02 and 04.

---

## Analysis

### Finding 1: Glyph Produces Higher Absolute Quality (+0.73 average)

Glyph-sourced implementations scored 9.93/10 vs 9.20/10 for text — a +0.73 gap. Glyph won or tied in 4 of 5 scenarios. The gap is driven by **completeness** (+0.8) and **structural precision** (+1.0), not correctness (+0.4).

### Finding 2: Structural Precision Is Glyph's Strongest Advantage (+1.0)

Glyph's symbolic notation directly maps to FastAPI patterns:
- `+ auth(jwt)` → `Depends(auth)` middleware
- `+ ratelimit(N/min)` → `Depends(check_rate_limit)` dependency
- `% db: Database` → `Depends(get_db)` injection with typed collection classes
- `db.collection.Method()` → collection classes with matching method names
- `&` / `*` / `~` → QueueWorker, CronScheduler, EventBus class abstractions with async handlers

Text descriptions of the same concepts ("requires JWT auth", "rate limited to N/min") produced flatter structures — inline function calls, global mutable state, and simple dict dispatchers.

### Finding 3: Background Task Infrastructure Is Glyph's Largest Win

The biggest single gap was in Scenario 04 (Job Queue). Glyph's `&`, `*`, `~` symbols produced full-featured class abstractions:
- `QueueWorker` with `asyncio.Queue`, register/enqueue/process methods
- `CronScheduler` with named job registration and run_job methods
- `EventBus` with pub/sub on/emit pattern

Text's prose description ("Queue Workers: Process email sending jobs") produced only flat function-in-dict mappings (`QUEUE_WORKERS = {"email.send": process_email_send}`) with no actual processing infrastructure. The handlers exist but cannot run.

### Finding 4: Text Excels at Behavioral Nuance and Security Patterns

Where text scored equal or higher than glyph (scenarios 01, 03, 05), the advantage came from:
- **03 Chat**: Text spec's "send acknowledgment back to sender, then broadcast" captured WebSocket ack semantics that glyph's `ws.broadcast(input)` couldn't express. Text version also stored messages to DB.
- **05 Auth**: Text version implemented token rotation on refresh, UserOut response filtering (strips password_hash), IP-based rate limiting, and time-based session expiration — production security patterns inferred from the prose description.

Glyph's terse symbols capture WHAT happens but not HOW. Natural language excels at expressing behavioral nuances and security considerations.

### Finding 5: Token Efficiency Is a Wash

Despite glyph files being 26% larger on average (236 vs 187 words), quality-per-word is essentially tied (5.28 vs 5.29). This means glyph's extra notation produces proportionally more quality — each structural symbol contributes to output quality. However, glyph does NOT achieve better quality with fewer tokens; its advantage is in quality ceiling, not token efficiency.

### Finding 6: Glyph's Advantage Is Concentrated in Scenarios 02 and 04

| Scenario | Gap | Pattern |
|----------|-----|---------|
| 01 CRUD | 0 (tied) | Simple CRUD — both notations handle equally well |
| 02 Webhook | +5 glyph | Middleware (auth + rate limit) and DI are glyph's strength |
| 03 Chat | 0 (tied) | Behavioral nuance helps text; structure helps glyph; nets out |
| 04 Job Queue | +6 glyph | Background infrastructure (`&`, `*`, `~`) is glyph's biggest win |
| 05 Auth | 0 (tied) | Complex but both handle well; cron gap is symmetric |

The gap appears when specs involve **middleware composition** and **background task infrastructure** — areas where glyph has dedicated symbols.

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

**The hypothesis is partially confirmed.** Glyph notation produces code that is more structurally precise (+1.0), more complete (+0.8), and slightly more correct (+0.4) than equivalent natural language descriptions. The advantage is concentrated in **middleware composition** and **background task infrastructure** — areas where glyph has dedicated symbols (`+`, `%`, `&`, `*`, `~`) that map directly to framework idioms.

However, **correctness is nearly tied** (10.0 vs 9.6), and text descriptions produce equivalent or superior results when behavioral nuance matters — security flows, protocol lifecycles, and edge-case handling. Token efficiency is also a wash (5.28 vs 5.29 QPW).

**Key insight**: Glyph's value is not in making prompts shorter — it's in making structural intent unambiguous. `+ ratelimit(1000/min)` is 23 characters but maps to a Depends() dependency. "Rate limited to 1000 requests per minute" is 43 characters but gets implemented as an inline function call. The information density per symbol is higher in glyph, even though the total character count is higher.

**Recommendation**: Glyph notation is most valuable as a structural scaffold for middleware, data models, and service architecture. It could be complemented with natural language annotations for behavioral specifications — a hybrid approach that combines glyph's structural precision with prose's behavioral expressiveness.

---

## Limitations

1. **Self-evaluation bias**: Same model family (Claude) generated and evaluated all code. Three independent evaluation sub-agents were used for checklist scoring, but all share the same underlying model.
2. **Single model**: Results may not generalize to other LLMs (GPT-4, Gemini, etc.).
3. **Single target language**: Python/FastAPI only. The hypothesis should be tested across multiple targets.
4. **Sample size**: 5 scenarios is small. Statistical significance is limited.
5. **Prompt sensitivity**: Results may vary with different prompt formulations.
6. **No runtime testing**: Generated code was syntax-validated but not executed.
7. **Evaluator subjectivity**: Despite checklists, rubric score assignments (e.g., 7 vs 8) involve judgment calls.
