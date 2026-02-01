# GlyphLang Code Review Issues

> Generated from comprehensive review of the `dev` branch on 2026-01-30.
> All P0 issues have been verified against the source code directly.
> Issues are ordered by priority (P0 > P1 > P2). Each issue contains enough
> context for an agent to locate, understand, and fix the problem independently.

---

## P0 -- Release Blockers

These must be fixed before any v1.x release. They represent functional bugs,
security vulnerabilities that directly expose user applications, or broken
showcase features.

---

### P0-1: `time.now()` and `now()` return hardcoded value instead of real time

**Files:**
- `pkg/vm/vm.go:1292-1298` (VM implementation)
- `pkg/interpreter/evaluator.go:594-601` (interpreter implementation)

**Problem:**
Both execution modes return a hardcoded Unix timestamp `1234567890` (Feb 2009)
instead of the actual current time. Any GlyphLang program using timestamps
gets incorrect data.

**Verified current code (VM, line 1296-1298):**
```go
// For now, return a mock timestamp
// TODO: Return actual time.Now().Unix() when time package is integrated
return IntValue{Val: 1234567890}, nil
```

**Verified current code (interpreter, line 595-601):**
```go
case "time.now":
    // Return a mock timestamp for now
    return int64(1234567890), nil
case "now":
    // Return a mock timestamp
    return int64(1234567890), nil
```

**Fix:**
- VM: Replace `IntValue{Val: 1234567890}` with `IntValue{Val: time.Now().Unix()}`
  for both `time.now` and `now` builtins. Add `"time"` to the import block if
  not already present.
- Interpreter: Replace `int64(1234567890)` with `time.Now().Unix()` for both
  cases.
- Remove the TODO comments.

**Acceptance criteria:**
- `time.now()` returns current Unix timestamp (within 1 second of `time.Now().Unix()`)
- Both `time.now()` and `now()` behave identically in VM and interpreter modes
- Existing tests still pass
- Add a test that verifies the returned value is within 2 seconds of `time.Now().Unix()`

---

### P0-2: Async execution data race in VM

**File:** `pkg/vm/vm.go:1948-1974`

**Problem:**
`execAsync` spawns a goroutine that shares `vm.constants` by reference (not
copied), reads `vm.locals` and `vm.globals` without synchronization, and writes
`future.Resolved` without atomic operations while `Await()` may be reading it.

**Verified current code (line 1948-1974):**
```go
go func() {
    defer close(future.Done)
    asyncVM := NewVM()
    asyncVM.constants = vm.constants       // Shared slice reference, not a copy
    for k, v := range vm.locals {          // Unsynchronized read of parent VM
        asyncVM.locals[k] = v
    }
    for k, v := range vm.globals {
        asyncVM.globals[k] = v
    }
    for k, v := range vm.builtins {
        asyncVM.builtins[k] = v
    }
    result, execErr := asyncVM.Execute(asyncBody)
    if execErr != nil {
        future.Error = execErr
    } else {
        future.Result = result
    }
    future.Resolved = true                 // Non-atomic write
}()
```

**Fix:**
1. Copy `vm.constants` instead of sharing the reference:
   ```go
   asyncVM.constants = make([]Value, len(vm.constants))
   copy(asyncVM.constants, vm.constants)
   ```
2. The maps (`locals`, `globals`, `builtins`) are already being copied by
   iterating key-value pairs into a new map, which is safe as long as the
   parent VM does not modify them concurrently during the range loop. If the
   parent VM can execute concurrently, protect the reads with a mutex. At
   minimum, document this assumption.
3. Remove `future.Resolved` field entirely and rely only on the `future.Done`
   channel (which is already closed via `defer close(future.Done)`). In
   `Await()`, use `<-future.Done` to detect completion instead of checking
   `future.Resolved`.

**Acceptance criteria:**
- `go test -race ./pkg/vm/...` passes with a test that exercises async execution
- `future.Resolved` field is either removed or accessed via `sync/atomic`
- `vm.constants` is copied, not shared
- Add a test: launch multiple async blocks concurrently and verify correct results

---

### P0-3: Parser does not support struct field default values

**File:** `pkg/parser/parser.go` -- type definition parsing section

**Problem:**
The parser rejects `=` in struct field definitions, so syntax like
`field: type! = defaultValue` causes a parse error. This breaks 4 example
files in the project's own examples directory.

**Verified failing examples (all ran and confirmed to produce parse errors):**
- `examples/defaults-demo/main.glyph` (line 12: `theme: str! = "light"`)
- `examples/todo-api/main.glyph` (line 9: `completed: bool! = false`)
- `examples/auth-demo/main.glyph` (line 9: `role: str! = "user"`)
- `examples/feature-showcase/main.glyph` (line 16: `role: str! = "user"`)

**Error produced:**
```
Expected identifier, but found =
Hint: Identifiers must start with a letter or underscore
```

**Fix:**
In the type definition parsing code (look for where struct fields are parsed --
the loop that reads `fieldName: fieldType` pairs), after parsing the field type,
check for an `=` token. If present, parse the expression following it as the
field's default value. Store this in the AST node (likely `FieldDefinition` or
equivalent struct). The interpreter and compiler must then use this default
when constructing instances where that field is not explicitly provided.

**Acceptance criteria:**
- All 4 failing example files parse and compile without errors
- A test verifies struct fields with defaults: creating an instance without
  specifying the field uses the default, and specifying the field overrides it
- Default values work for at least: `str`, `bool`, `int`, `float` types

---

### P0-4: Parser does not support variable reassignment without `$` prefix

**File:** `pkg/parser/parser.go` -- statement parsing section

**Problem:**
The parser requires `$ varName = value` for all assignments, including
reassignments of existing variables. Bare `varName = value` produces a parse
error. This breaks loop-based examples that need to update variables.

**Verified failing examples (all ran and confirmed to produce parse errors):**
- `examples/for-loop-demo.glyph` (line 10: `greetings = greetings + [greeting]`)
- `examples/while-loop-demo.glyph` (line 9: `result = result + i`)

**Error produced:**
```
Unexpected identifier in statement position: =
Hint: Did you mean to assign a variable? Use '$ varName = value'
```

**Fix:**
In `parseStatement` (around line 1784 of `parser.go`), when the parser
encounters an identifier token that is NOT a keyword and the next token is `=`,
`+=`, `-=`, etc., parse it as a variable reassignment statement. The key
distinction:
- `$ x = 5` declares a new variable
- `x = 10` reassigns an existing variable (should NOT create a new binding)

Create an `AssignStatement` (or `ReassignStatement`) AST node for the bare
form. The interpreter/compiler should verify the variable exists in the current
scope and emit an error if it doesn't (to prevent accidental creation).

**Acceptance criteria:**
- Both failing example files parse and compile without errors
- `$ x = 5` followed by `x = 10` updates x to 10
- `y = 10` without a prior `$ y` declaration produces a clear error:
  "undefined variable: y"
- Compound assignment operators work: `x += 1`, `x -= 1`

---

### P0-5: No request body size limit in core server handler

**File:** `pkg/server/handler.go:94`

**Problem:**
`parseJSONBody` calls `io.ReadAll(r.Body)` without any size limit. An attacker
can send an arbitrarily large request body to exhaust server memory. This is a
denial-of-service vulnerability.

Note: `cmd/glyph/main.go:1510` correctly uses `io.LimitReader` with a 10MB cap
for the compiled route handler, but the core `pkg/server/handler.go` does not.

**Verified current code (handler.go, line 86-94):**
```go
func parseJSONBody(r *http.Request, ctx *Context) error {
    contentType := r.Header.Get("Content-Type")
    if !strings.Contains(contentType, "application/json") && contentType != "" {
        return fmt.Errorf("expected application/json content type, got %s", contentType)
    }
    body, err := io.ReadAll(r.Body)  // No size limit
```

**Fix:**
Wrap `r.Body` with `http.MaxBytesReader` before reading:
```go
const maxBodySize = 10 << 20 // 10 MB
r.Body = http.MaxBytesReader(ctx.ResponseWriter, r.Body, maxBodySize)
body, err := io.ReadAll(r.Body)
```

Alternatively, make the max size configurable via server config. Handle the
`http.MaxBytesError` to return a 413 Payload Too Large response.

**Acceptance criteria:**
- Requests with bodies exceeding the limit return HTTP 413
- Normal requests with reasonable body sizes continue to work
- Add a test that sends a body over the limit and verifies rejection

---

### P0-6: Internal error details exposed to clients

**Files:**
- `pkg/server/handler.go:144-164`
- `pkg/server/errors.go:49-63`
- `pkg/interpreter/interpreter.go:474-480`

**Problem:**
Raw Go `err.Error()` strings are included in JSON responses sent to clients.
These can expose file paths, database connection details, internal stack
information, and other implementation details useful to attackers.

**Verified current code (handler.go, line 145-164):**
```go
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, statusCode int, message string, err error) {
    log.Printf("[ERROR] %s %s: %s - %v", r.Method, r.URL.Path, message, err)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    response := map[string]interface{}{
        "error":   true,
        "message": message,
        "code":    statusCode,
    }
    if err != nil {
        response["details"] = err.Error()  // Leaks internal error to client
    }
    json.NewEncoder(w).Encode(response)
}
```

**Verified current code (errors.go, line 49-63):**
```go
func (e *BaseError) ToResponse() *ErrorResponse {
    resp := &ErrorResponse{
        Status:  e.Code,
        Error:   e.Type,
        Message: e.Msg,
    }
    if e.Detail != "" {
        resp.Details = e.Detail
    } else if e.Cause != nil {
        resp.Details = e.Cause.Error()  // Leaks cause to client
    }
    return resp
}
```

**Verified current code (interpreter.go, line 474-480):**
```go
return &Response{
    StatusCode: 500,
    Body: map[string]interface{}{
        "error": err.Error(),  // Leaks internal error to client
    },
}, err
```

**Fix:**
1. In `handler.go`, log the full error detail server-side (already done via
   `log.Printf`) but remove the `details` field from client responses for
   5xx errors. For 4xx errors, include only the user-facing `message`.
2. In `errors.go`, `ToResponse()` should not include `e.Cause.Error()`. Use
   `e.Detail` (the explicit developer-set detail) only. If no detail is set,
   omit the field.
3. In `interpreter.go`, return a generic "Internal server error" message in
   the response body. Log the actual error.

**Acceptance criteria:**
- No 5xx response body contains raw Go error strings
- 4xx responses contain only the developer-specified message
- Full error details are logged server-side via `log.Printf`
- Add a test that triggers an internal error and verifies the response does
  not contain file paths or Go error formatting

---

### P0-7: WebSocket upgrader accepts all origins

**File:** `pkg/websocket/server.go:17-24`

**Problem:**
The WebSocket upgrader's `CheckOrigin` always returns `true`, enabling
cross-site WebSocket hijacking (CSWSH). A malicious website can open a
WebSocket to the GlyphLang server and interact with it using the victim's
session cookies.

**Verified current code (server.go, line 17-24):**
```go
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        // Allow all origins for now (should be configurable in production)
        return true
    },
}
```

**Fix:**
1. Remove the package-level `upgrader` variable.
2. Add an `AllowedOrigins` field to the WebSocket `Config` struct (in
   `pkg/websocket/config.go`).
3. Create the upgrader per-hub or per-server with a `CheckOrigin` function
   that validates the request's `Origin` header against the allowed list.
4. Default behavior: if no origins are configured, only allow same-origin
   requests (compare `Origin` header to `Host` header).

**Acceptance criteria:**
- By default, cross-origin WebSocket connections are rejected
- Configuring `AllowedOrigins: ["https://example.com"]` allows that origin
- A wildcard `"*"` allows all origins (with a log warning)
- Add tests for: same-origin allowed, cross-origin rejected, configured origin
  allowed

---

## P1 -- Security & Correctness

These are security issues and correctness problems that should be fixed before
a stable release but are not as immediately exploitable as P0 items.

---

### P1-1: VM string builtins are ASCII-only, diverge from interpreter

**File:** `pkg/vm/vm.go:1474-1621`

**Problem:**
The VM reimplements `toUpper`, `toLower`, `trimSpace`, `splitString`,
`joinStrings`, `stringContains`, and `replaceAll` as custom functions that only
handle ASCII characters. The interpreter uses Go's `strings` package which
handles full Unicode. This means `upper("cafe")` can produce different results
in compiled vs interpreted mode for non-ASCII input.

Additionally, `toUpper`/`toLower` allocate `make([]rune, len(s))` using byte
length instead of rune count, producing trailing zero runes for multi-byte
strings. `splitString` with empty delimiter slices at byte offsets, corrupting
multi-byte characters.

**Fix:**
Replace all custom string helper functions with Go stdlib equivalents:
- `toUpper(s)` -> `strings.ToUpper(s)`
- `toLower(s)` -> `strings.ToLower(s)`
- `trimSpace(s)` -> `strings.TrimSpace(s)`
- `splitString(s, d)` -> `strings.Split(s, d)` (for empty delim, use
  `strings.Split(s, "")` which splits on rune boundaries)
- `joinStrings(p, d)` -> `strings.Join(p, d)`
- `stringContains(s, sub)` -> `strings.Contains(s, sub)`
- `replaceAll(s, old, new)` -> `strings.ReplaceAll(s, old, new)`

Delete all the custom helper functions (lines 1474-1621) and the `isWhitespace`
helper. Add `"strings"` to the import block if not already present.

Also delete the duplicate `joinStrings` in `cmd/glyph/main.go` (search for
`func joinStrings`).

**Acceptance criteria:**
- `upper("cafe")` produces identical output in VM and interpreter modes
- Multi-byte Unicode strings (e.g., CJK, emoji, accented chars) handled correctly
- `go test -race ./pkg/vm/...` passes
- Custom helper functions are removed, stdlib is used
- The duplicate `joinStrings` in `main.go` is removed

---

### P1-2: `SanitizeSQL` function is fundamentally flawed

**File:** `pkg/security/sql_injection.go:133-140`

**Problem:**
`SanitizeSQL` uses regex-based blacklisting to sanitize SQL, which is a
known-broken approach. It doesn't handle backslash escaping, Unicode
normalization, double-encoded characters, or alternative comment syntax.
The ORM correctly uses parameterized queries, making this function both
misleading and dangerous -- developers might rely on it for safety.

**Verified current code (line 134-140):**
```go
func SanitizeSQL(input string) string {
    input = regexp.MustCompile(`--.*$`).ReplaceAllString(input, "")
    input = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(input, "")
    input = strings.ReplaceAll(input, "'", "''")
    input = strings.ReplaceAll(input, "\x00", "")
    return input
}
```

**Fix:**
Option A (preferred): Delete the function entirely. Search for callers of
`SanitizeSQL` across the codebase. If callers exist, refactor them to use
parameterized queries instead.

Option B: If the function must remain for some reason, rename it to
`EscapeSQLString` and add a prominent doc comment:
```go
// EscapeSQLString performs basic escaping of a string value for SQL contexts.
// WARNING: This is NOT a security measure. Always use parameterized queries
// ($1, $2 placeholders) for user-provided values. This function only escapes
// single quotes for contexts where parameterized queries are not available
// (e.g., identifiers). It does NOT protect against SQL injection.
```
And simplify it to only double single quotes and remove null bytes (drop the
regex-based comment stripping which is a false sense of security).

**Acceptance criteria:**
- `SanitizeSQL` is either deleted or clearly documented as non-security
- No caller relies on it for security against injection
- All SQL queries involving user data use parameterized queries
- Tests updated accordingly

---

### P1-3: `EscapeHTML` uses non-deterministic map iteration order

**File:** `pkg/security/xss.go:293-307`

**Problem:**
`EscapeHTML` uses a `map[string]string` for replacement pairs and iterates with
`for old, new := range replacements`. Go map iteration is non-deterministic.
If `&` is not replaced before `<` (which becomes `&lt;`), a subsequent `&`
replacement would double-escape it to `&amp;lt;`.

**Verified current code (line 293-307):**
```go
func EscapeHTML(s string) string {
    replacements := map[string]string{
        "&":  "&amp;",
        "<":  "&lt;",
        ">":  "&gt;",
        "\"": "&quot;",
        "'":  "&#39;",
    }
    result := s
    for old, new := range replacements {
        result = strings.ReplaceAll(result, old, new)
    }
    return result
}
```

**Fix:**
Replace the entire function body with Go's standard library:
```go
func EscapeHTML(s string) string {
    return html.EscapeString(s)
}
```
Add `"html"` to the import block. Note that `html.EscapeString` escapes `&`,
`<`, `>`, `"`, and `'` in the correct deterministic order.

Also check `EscapeJS` (line 310) for the same map iteration pattern and fix
if present.

**Acceptance criteria:**
- `EscapeHTML` uses `html.EscapeString` or a deterministic ordered replacement
- Input `<script>&"hello"</script>` produces correct output on every call
- No double-escaping occurs
- Existing tests pass

---

### P1-4: `SendFile` lacks path traversal protection

**File:** `pkg/web/web.go:401-421`

**Problem:**
`ResponseHelper.SendFile` opens `targetPath` directly with `os.Open` without
any path validation or root directory confinement. If `targetPath` is derived
from user input (e.g., a route parameter), an attacker can read arbitrary files.

The `StaticFileServer` in the same file (lines 252-283) correctly implements
path traversal protection with `path.Clean`, `filepath.EvalSymlinks`, and
`isSubPath`. `SendFile` has none of this.

**Verified current code (line 401-421):**
```go
func (rh *ResponseHelper) SendFile(w http.ResponseWriter, r *http.Request, targetPath string) error {
    file, err := os.Open(targetPath)
    // ... no path validation
}
```

**Fix:**
Add path validation before opening the file:
1. Accept a `rootDir` parameter (or get it from `ResponseHelper` config) that
   defines the allowed base directory.
2. Clean and resolve the path: `cleanPath := filepath.Clean(targetPath)`
3. Resolve symlinks: `realPath, err := filepath.EvalSymlinks(cleanPath)`
4. Verify the resolved path is under `rootDir` using prefix matching with a
   trailing path separator (same pattern as `StaticFileServer.isSubPath`).
5. Reject paths that escape the root.

If changing the function signature is not feasible, at minimum:
- Resolve the path and reject absolute paths or paths containing `..`
- Log a warning if the path appears to be a traversal attempt

**Acceptance criteria:**
- `SendFile` rejects paths containing `..` that escape the root directory
- `SendFile` rejects symlinks pointing outside the root
- A test verifies: paths like `../../etc/passwd` are rejected
- Normal relative paths within the root continue to work

---

### P1-5: `X-Forwarded-For` header trusted unconditionally

**File:** `pkg/server/middleware.go:318-332`

**Problem:**
`getClientIP` trusts `X-Forwarded-For` and `X-Real-IP` headers without
verifying they come from a trusted proxy. Any client can set these headers to
spoof their IP, bypassing the rate limiter and auth failure lockout.

**Verified current code (line 318-332):**
```go
func getClientIP(r *http.Request) string {
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        parts := strings.Split(xff, ",")
        return strings.TrimSpace(parts[0])
    }
    if xri := r.Header.Get("X-Real-IP"); xri != "" {
        return xri
    }
    return r.RemoteAddr
}
```

**Fix:**
Add a configurable list of trusted proxy IPs/CIDRs. Only use `X-Forwarded-For`
and `X-Real-IP` when the request's `RemoteAddr` matches a trusted proxy. When
trusted, take the rightmost non-trusted IP from `X-Forwarded-For` (not the
leftmost, which is client-controlled).

If a full implementation is too complex, the minimal fix is:
1. Add a `TrustProxy bool` field to the server/middleware configuration.
2. When `TrustProxy` is `false` (the default), ignore `X-Forwarded-For` and
   `X-Real-IP` entirely and use `r.RemoteAddr`.
3. When `TrustProxy` is `true`, use the current behavior.

**Acceptance criteria:**
- By default (no proxy config), `RemoteAddr` is used directly
- When proxy trust is configured, `X-Forwarded-For` is respected
- An attacker setting `X-Forwarded-For` without a trusted proxy has no effect
  on rate limiting
- Add a test for both proxy-trusted and untrusted scenarios

---

### P1-6: Rate limiter maps grow unbounded

**File:** `pkg/server/middleware.go:342-398`

**Problem:**
The `limits` map in `RateLimitMiddleware` and the `failureTrackers` map in
`AuthFailureLockout` (search for it in the same file) grow indefinitely. Every
unique IP address creates a new entry that is never evicted. An attacker making
requests from many IPs (or spoofing X-Forwarded-For per P1-5) can cause
unbounded memory growth.

**Verified current code (line 342-398) -- `limits` map created at line 351:**
```go
var mu sync.Mutex
limits := make(map[string]*clientLimit)
// ... entries added, never removed
```

**Fix:**
Add periodic cleanup of stale entries. Options:
1. (Simple) Add a max map size. When exceeded, clear the oldest half of entries.
2. (Better) Run a background goroutine that periodically (e.g., every 60s)
   iterates the map and removes entries older than a threshold (e.g., 10
   minutes since last request).
3. (Best) Use a sync.Map with expiring entries or a third-party LRU/TTL cache.

For the `failureTrackers` map, entries should expire after the lockout duration
has passed.

**Acceptance criteria:**
- After a configurable duration with no requests, an IP's rate limit entry is
  cleaned up
- Memory does not grow linearly with the number of unique IPs over time
- Add a test that creates many entries and verifies cleanup occurs

---

### P1-7: No CSRF protection middleware

**Files:** `pkg/server/middleware.go` (absent)

**Problem:**
There is no CSRF token generation, validation, or middleware anywhere in the
codebase. For a web backend framework that supports cookie-based authentication
(via middleware), this is a significant gap. An attacker can forge state-changing
requests from a malicious site.

**Fix:**
Add a `CSRFMiddleware` in `pkg/server/middleware.go` that:
1. Generates a random CSRF token and sets it as a cookie
   (`csrf_token`, `SameSite=Strict`, `HttpOnly=false` so JS can read it)
2. On state-changing requests (POST, PUT, PATCH, DELETE), validates the token
   from either:
   - `X-CSRF-Token` request header, OR
   - `csrf_token` form field
3. Compares using `crypto/subtle.ConstantTimeCompare`
4. Returns 403 Forbidden if the token is missing or invalid

This should be opt-in middleware, not enforced by default (some APIs are
token-authenticated and don't need CSRF).

**Acceptance criteria:**
- `CSRFMiddleware` exists and can be added to routes
- GET/HEAD/OPTIONS requests pass through without CSRF check
- POST/PUT/PATCH/DELETE without valid token return 403
- Token is generated with `crypto/rand`
- Add comprehensive tests

---

### P1-8: `GetLastInsertID` uses original values instead of sanitized values

**File:** `pkg/database/postgres.go:354-376`

**Problem:**
`GetLastInsertID` validates `table` and `idColumn` via `SanitizeIdentifier()`
but then uses the original unsanitized values in the query. The sanitized
values are assigned to `_`. While the regex validation ensures the values are
safe, this pattern is fragile -- a future refactor could remove the validation
without updating the query.

**Verified current code (line 357-371):**
```go
sanitizedTable, err := SanitizeIdentifier(table)
// ...
sanitizedColumn, err := SanitizeIdentifier(idColumn)
// ...
query := fmt.Sprintf("SELECT CURRVAL(pg_get_serial_sequence('%s', '%s'))", table, idColumn)
_ = sanitizedTable  // Used for validation only
_ = sanitizedColumn // Used for validation only
```

**Fix:**
Use the sanitized values in the query:
```go
query := fmt.Sprintf("SELECT CURRVAL(pg_get_serial_sequence('%s', '%s'))", sanitizedTable, sanitizedColumn)
```
Remove the `_ = sanitizedTable` and `_ = sanitizedColumn` lines.

**Acceptance criteria:**
- The sanitized values are used in the query string
- No `_ =` assignments for sanitized identifiers
- Existing database tests pass

---

### P1-9: Raw SQL `Query()` exposed to interpreter without safeguards

**File:** `pkg/database/handler.go:142-145`

**Problem:**
`TableHandler.Query()` accepts raw SQL strings and is accessible from GlyphLang
programs via the interpreter's allowed methods whitelist. If a GlyphLang
developer constructs SQL via string concatenation, this enables SQL injection.

**Verified current code (line 142-145):**
```go
// Query executes a raw SQL query
func (t *TableHandler) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
    return t.orm.Query(t.ctx, query, args...)
}
```

**Fix:**
Option A: Remove `Query` from the allowed methods whitelist in
`pkg/interpreter/database.go` (search for `"Query": true`) and force all
queries through the query builder.

Option B: If raw queries must be supported, add static analysis in the
`SQLInjectionDetector` (pkg/security/sql_injection.go) to detect when the
query argument to `Query()` is built via string concatenation (as opposed to
a string literal). Emit a compile-time warning.

Option C: Add a `PreparedQuery` method that requires the query to be a string
literal (validated at parse time) and pass all dynamic values as args.

**Acceptance criteria:**
- GlyphLang programs cannot accidentally pass concatenated strings as SQL
- Either the method is restricted or a warning is emitted at compile time
- Add a test showing that string concatenation in Query() is detected/prevented

---

### P1-10: Unbounded LLM response body read

**File:** `pkg/llm/handler.go:569`

**Problem:**
When reading responses from LLM providers, `io.ReadAll(resp.Body)` is called
without a size limit. A misconfigured or malicious LLM endpoint could return
an extremely large response to exhaust memory.

**Verified current code (line 569):**
```go
respBody, err := io.ReadAll(resp.Body)
```

**Fix:**
Add a size limit:
```go
const maxLLMResponseSize = 50 << 20 // 50 MB
limitedReader := io.LimitReader(resp.Body, maxLLMResponseSize)
respBody, err := io.ReadAll(limitedReader)
```

Make the limit configurable via the LLM handler config if possible.

**Acceptance criteria:**
- LLM responses exceeding the limit produce a clear error
- Normal LLM responses continue to work
- Add a test with an oversized response body

---

### P1-11: WebSocket default config allows unlimited connections

**File:** `pkg/websocket/config.go:53-54`

**Problem:**
Default WebSocket configuration has `MaxConnectionsPerHub: 0` and
`MaxConnectionsPerRoom: 0`, meaning no connection limits. An attacker can open
unlimited connections to exhaust server resources (file descriptors, memory).

**Fix:**
Set reasonable defaults:
```go
MaxConnectionsPerHub:  10000,
MaxConnectionsPerRoom: 1000,
```
Enforce these limits in the connection acceptance code. When the limit is
reached, reject new connections with an HTTP 503 before upgrading.

**Acceptance criteria:**
- Default config has non-zero connection limits
- Connections exceeding the limit are rejected
- Limits are configurable (0 means "use default", -1 could mean "unlimited")
- Add a test verifying rejection when limit is reached

---

## P2 -- Code Quality & Maintainability

These issues affect long-term maintainability, developer experience, and code
quality. They should be addressed before v1.0 but are not security or
correctness blockers.

---

### P2-1: Run `gofmt` across entire codebase

**Problem:**
72 Go files have formatting issues. `gofmt` is not enforced.

**Fix:**
```bash
gofmt -w .
```

Then add a CI check that fails if `gofmt -l .` produces any output.

**Acceptance criteria:**
- `gofmt -l .` returns no output
- All files are consistently formatted

---

### P2-2: Split `cmd/glyph/main.go` (2,905 lines) into multiple files

**File:** `cmd/glyph/main.go`

**Problem:**
A single file contains CLI setup, server startup, hot reload management, route
registration, request handling, WebSocket bytecode execution, value conversion,
template generation, context generation, validation, file watching, and 20+
cobra command implementations.

**Fix:**
Split into separate files in `cmd/glyph/`:
- `main.go` -- only `main()`, root command setup, version
- `commands.go` -- cobra command definitions (build, run, test, etc.)
- `server.go` -- server startup, route registration, request handling
- `hotreload.go` -- hot reload manager and related types
- `handlers.go` -- HTTP handler helpers, value conversion
- `templates.go` -- project template generation

Each file should be in `package main`. No behavior change needed.

**Acceptance criteria:**
- No file in `cmd/glyph/` exceeds 500 lines
- `go build ./cmd/glyph/...` still works
- All tests pass
- No behavior change

---

### P2-3: Decompose `evaluateFunctionCall` (618 lines)

**File:** `pkg/interpreter/evaluator.go:592-1210`

**Problem:**
Every built-in function is a case in a single 618-line switch statement. This
is difficult to read, test individually, and extend.

**Fix:**
Refactor to a dispatch table pattern (similar to the VM's `builtins` map):
```go
type builtinFunc func(i *Interpreter, args []Expr, env *Environment) (interface{}, error)

var builtinFuncs = map[string]builtinFunc{
    "time.now": builtinTimeNow,
    "now":      builtinTimeNow,
    "upper":    builtinUpper,
    "lower":    builtinLower,
    // ...
}
```
Then `evaluateFunctionCall` becomes:
```go
if fn, ok := builtinFuncs[expr.Name]; ok {
    return fn(i, expr.Args, env)
}
```

Each builtin becomes a separate function that can be tested individually.

**Acceptance criteria:**
- `evaluateFunctionCall` is under 100 lines
- Each builtin is a separate named function
- All interpreter tests pass
- No behavior change

---

### P2-4: Move AST types out of `pkg/interpreter`

**File:** `pkg/interpreter/ast.go` (90+ exported types)

**Problem:**
AST types (every node, type, expression, statement) live in `pkg/interpreter`,
forcing the parser to import the interpreter. This creates tight coupling and
means `pkg/interpreter` exports ~90 types that are really AST concerns.

**Fix:**
Create `pkg/ast/` package and move all AST types there. Update imports in:
- `pkg/parser/` (currently imports `pkg/interpreter` for AST types)
- `pkg/compiler/` (compiles AST nodes)
- `pkg/interpreter/` (evaluates AST nodes)
- Any other package importing AST types from `pkg/interpreter`

This is a large refactor. Consider doing it as the final P2 item.

**Acceptance criteria:**
- `pkg/ast/` package exists with all AST types
- `pkg/interpreter/` no longer exports AST types
- `pkg/parser/` imports `pkg/ast/` instead of `pkg/interpreter/`
- All tests pass
- No behavior change

---

### P2-5: Remove `MockInterpreter` from production code

**File:** `pkg/server/server.go:162`

**Problem:**
`MockInterpreter` is defined in the `server` package's production source file,
not in a test file. Mock types should not be in production code.

**Fix:**
Move the `MockInterpreter` type definition to `pkg/server/server_test.go` or
to a `pkg/server/servertest/` package if it's needed by external tests.

**Acceptance criteria:**
- `MockInterpreter` does not appear in any non-test `.go` file
- Tests that use it still compile and pass

---

### P2-6: Replace stub tests with real assertions

**Files:**
- `tests/e2e_test.go` -- at least 6 tests assign bytecode to `_` and log
  "passed" without asserting behavior
- `tests/integration_test.go` -- at least 10 tests log expected values without
  asserting them

**Problem:**
These tests create the appearance of coverage without testing behavior. They
pass unconditionally regardless of correctness. Examples:

```go
// e2e_test.go
_ = bytecode // TODO: Use bytecode when server is ready
t.Log("Path parameters test passed (compilation)")

// integration_test.go
t.Logf("Expected tokens: %v", tt.expected)  // Never asserted
```

**Fix:**
For each stub test, either:
1. Implement the actual assertion (preferred)
2. Mark with `t.Skip("not yet implemented: <reason>")` so they show as skipped
   rather than passed -- this makes the test report honest

Priority stubs to implement assertions for:
- Path parameter E2E tests (`tests/e2e_test.go`, around line 183)
- Middleware E2E tests (`tests/e2e_test.go`, around line 209)
- Auth E2E tests (`tests/e2e_test.go`, around line 287)
- Error handling E2E tests (`tests/e2e_test.go`, around line 312)
- Lexer integration tests (`tests/integration_test.go`, around line 181)
- Type checker integration tests (`tests/integration_test.go`, around line 256)
- Route matching integration tests (`tests/integration_test.go`, around line 433)
- Validation integration tests (`tests/integration_test.go`, around line 522)

**Acceptance criteria:**
- No test function logs "passed" without asserting anything
- Unimplemented tests use `t.Skip()` instead of `t.Log()`
- At least the compilation-verification tests assert bytecode is non-nil
- Test output clearly distinguishes between passing and skipped tests

---

### P2-7: Resolve all 31 TODO comments

**Problem:**
31 TODO comments remain in the codebase. Locations:

| File | Count | Key TODOs |
|------|-------|-----------|
| `tests/e2e_test.go` | 16 | Server/middleware/auth not ready |
| `tests/integration_test.go` | 10 | Various features not ready |
| `pkg/vm/vm.go` | 1 | Return actual time (covered by P0-1) |
| `pkg/parser/parser.go` | 1 | Add proper method call support |
| `pkg/debug/repl.go` | 1 | VM locals/globals exposure |
| `tests/bytecode_integration_test.go` | 1 | Pure Go implementation |
| `tests/parser_comprehensive_test.go` | 1 | Statement types not implemented |

**Fix:**
For each TODO:
1. If the feature is now implemented, remove the TODO and write the code
2. If the feature is not yet available, convert to a `t.Skip()` in tests or
   create a GitHub issue and reference it: `// TODO(#123): description`
3. No bare TODOs should remain -- they should all reference an issue number

**Acceptance criteria:**
- Zero bare `TODO` comments (without issue references) in production code
  (`pkg/` and `cmd/`)
- Test TODOs are converted to `t.Skip()` with explanatory message
- Any remaining TODOs reference a GitHub issue number

---

### P2-8: Remove dead code

**Files:**
- `cmd/glyph/main.go` (around line 1674) -- unused `watchFile` function
- `pkg/debug/example_usage.go` -- entire file is a block comment, compiles to
  nothing

**Fix:**
1. Delete the `watchFile` function from `main.go` (the hot reload system uses
   `hotReloadManager.watchForChanges()` instead)
2. Delete `pkg/debug/example_usage.go` or convert it to a `_test.go` example
   file if the content is valuable as documentation

**Acceptance criteria:**
- No unused unexported functions remain
- `go build ./...` passes
- No empty/comment-only production source files

---

### P2-9: Consistent pointer vs value types in compiler switch

**File:** `pkg/compiler/compiler.go:298-338`

**Problem:**
`compileStatement` has paired cases for both pointer and value types:
```go
case *interpreter.AssignStatement:
    return c.compileAssignStatement(s)
case interpreter.AssignStatement:
    return c.compileAssignStatement(&s)
```
This is repeated for 8 statement types (16 case branches), indicating the AST
inconsistently uses pointers vs values.

**Fix:**
Standardize the AST to always use pointer types for statements and expressions
(the more common Go convention for interface implementations with methods).
Then the compiler only needs one case per type. This is related to P2-4 (AST
package extraction) and could be done as part of that refactor.

If a full AST refactor is deferred, at minimum add a comment explaining why
both forms are handled.

**Acceptance criteria:**
- Either: AST consistently uses pointers (removing half the case branches)
- Or: a comment explains the dual handling and it's tested
- All tests pass

---

### P2-10: Make version string injectable via ldflags

**File:** `cmd/glyph/main.go:41`

**Current code:**
```go
var version = "0.4.0"
```

**Fix:**
The variable is already `var` (not `const`), so it can be overridden with
`-ldflags`. Update the `Makefile` build targets to inject the version:
```makefile
VERSION := $(shell git describe --tags --always --dirty)
build:
	go build -ldflags "-X main.version=$(VERSION)" -o glyph ./cmd/glyph
```

**Acceptance criteria:**
- `Makefile` build target injects version via ldflags
- `glyph --version` shows the injected version
- Default fallback still works when building without ldflags

---

### P2-11: Add CI enforcement for formatting and linting

**Problem:**
No evidence of CI enforcement for `gofmt`, `go vet`, or linting. The 72
formatting violations suggest no pre-commit hooks or CI checks exist.

**Fix:**
Add a GitHub Actions workflow (`.github/workflows/ci.yml`) or equivalent:
```yaml
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Check formatting
        run: test -z "$(gofmt -l .)"
      - name: Vet
        run: go vet ./...
      - name: Build
        run: go build ./...
      - name: Test
        run: go test -race ./...
```

**Acceptance criteria:**
- CI workflow exists and runs on pull requests
- PRs with formatting issues fail CI
- PRs with `go vet` warnings fail CI
- All tests must pass for PR to merge

---

### P2-12: `os.Exit` called from non-main functions

**File:** `cmd/glyph/main.go` (around line 1182)

**Problem:**
`os.Exit(1)` is called from within `setupRoutes`, making the function hard to
test. Only `main()` should call `os.Exit`.

**Current code:**
```go
if compiler.IsSemanticError(compileErr) {
    printError(fmt.Errorf("compilation error for %s: %v", route.Path, compileErr))
    os.Exit(1)
}
```

**Fix:**
Return an error from `setupRoutes` instead:
```go
if compiler.IsSemanticError(compileErr) {
    return fmt.Errorf("compilation error for %s: %v", route.Path, compileErr)
}
```
Handle the error in the caller and let `main()` decide whether to exit.

Search for other `os.Exit` calls outside of `main()` and cobra `Run` handlers
and apply the same pattern.

**Acceptance criteria:**
- No `os.Exit` calls outside of `main()` or `cobra.Command.Run` handlers
- Functions return errors instead of exiting
- Behavior is unchanged from the user's perspective

---

### P2-13: Password included in connection string (logging risk)

**File:** `pkg/database/database.go:110-121`

**Problem:**
`Config.ConnectionString()` builds a connection string with the password in
plaintext. If this string appears in logs (e.g., via connection error messages),
the password is exposed.

**Fix:**
Add a `SafeConnectionString()` method that masks the password for logging:
```go
func (c *Config) SafeConnectionString() string {
    // Same as ConnectionString() but with password replaced by "****"
}
```
Use `SafeConnectionString()` in all log statements. Keep `ConnectionString()`
for the actual database driver.

**Acceptance criteria:**
- No log statement includes the raw password
- Connection errors logged with masked password
- The database driver still receives the real password
- Add a test verifying password is masked in `SafeConnectionString()`
