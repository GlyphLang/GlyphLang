package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The notation's symbol vocabulary is what the spec leads with, but several
// symbols turned out to be parsed and then ignored at runtime - providers,
// auth and ratelimit each shipped that way. These tests exercise one small
// program per symbol and assert something observable, so a symbol that stops
// doing anything fails here instead of being discovered in production.
//
// Symbols that do nothing yet are skipped with the reason, the same convention
// examples_runtime_test.go uses for broken examples.

var inertSymbols = map[string]string{
	"*": "cron tasks are parsed and listed but no scheduler runs them",
	"~": "event handlers are registered but EmitEvent has no caller at runtime",
	"&": "queue workers are registered but no queue is consumed",
}

// writeProgram puts a .glyph program in a temp dir and returns its path.
func writeProgram(t *testing.T, source string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "main.glyph")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("writing program: %v", err)
	}
	return path
}

// serve boots a program and returns a request helper plus the server log.
func serve(t *testing.T, binary, source string) (func(method, path, body string) (int, string), func() string) {
	t.Helper()

	file := writeProgram(t, source)
	port := freePort(t)
	stop, logs := startExample(t, binary, file, port)
	t.Cleanup(stop)

	request := func(method, path, body string) (int, string) {
		t.Helper()

		var reader *strings.Reader
		if body == "" {
			reader = strings.NewReader("")
		} else {
			reader = strings.NewReader(body)
		}
		req, err := http.NewRequest(method, fmt.Sprintf("http://127.0.0.1:%d%s", port, path), reader)
		if err != nil {
			t.Fatalf("building request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s %s: %v\nserver log:\n%s", method, path, err, logs())
		}
		defer resp.Body.Close()

		var out bytes.Buffer
		_, _ = out.ReadFrom(resp.Body)
		return resp.StatusCode, out.String()
	}

	return request, logs
}

// TestSymbolRoutesAndReturns covers @ (route), : (type), $ (binding) and
// > (return, with an explicit status).
func TestSymbolRoutesAndReturns(t *testing.T) {
	if testing.Short() {
		t.Skip("boots a server; skipped in -short")
	}

	request, _ := serve(t, buildGlyphBinary(t), `: Reply {
  message: str!
}

@ GET /hello -> Reply {
  $ greeting = "hello"
  > {message: greeting} :: 201
}`)

	status, body := request("GET", "/hello", "")
	if status != 201 {
		t.Errorf("explicit status suffix ignored: got %d, want 201", status)
	}
	if !strings.Contains(body, `"message":"hello"`) {
		t.Errorf("binding did not reach the response: %s", body)
	}
}

// TestSymbolGuard covers ? used as a guard, which must short-circuit with the
// declared status rather than running the rest of the route.
func TestSymbolGuard(t *testing.T) {
	if testing.Short() {
		t.Skip("boots a server; skipped in -short")
	}

	request, _ := serve(t, buildGlyphBinary(t), `@ GET /guarded {
  $ allowed = false
  ? allowed :: 403 "not allowed"
  > {reached: true}
}`)

	status, body := request("GET", "/guarded", "")
	if status != 403 {
		t.Errorf("guard did not fire: got %d, want 403\nbody: %s", status, body)
	}
	if strings.Contains(body, "reached") {
		t.Errorf("guard did not short-circuit the route body: %s", body)
	}
}

// TestSymbolInputValidation covers < (expects), which must reject a body that
// does not satisfy the declared type.
func TestSymbolInputValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("boots a server; skipped in -short")
	}

	request, _ := serve(t, buildGlyphBinary(t), `: NewUser {
  name: str!
  age: int!
}

@ POST /users {
  < input: NewUser
  > {name: input.name}
}`)

	status, body := request("POST", "/users", `{"name":"ada","age":36}`)
	if status != 200 {
		t.Errorf("valid body rejected: got %d\nbody: %s", status, body)
	}

	status, body = request("POST", "/users", `{"name":"ada"}`)
	if status < 400 || status >= 500 {
		t.Errorf("body missing a required field should be a client error: got %d\nbody: %s", status, body)
	}
}

// TestSymbolMiddleware covers + for both directives it supports, which were
// inert until they were wired: an unconfigured auth must deny, and a rate
// limit must reject once the budget is spent.
func TestSymbolMiddleware(t *testing.T) {
	if testing.Short() {
		t.Skip("boots a server; skipped in -short")
	}

	request, _ := serve(t, buildGlyphBinary(t), `@ GET /guarded {
  + auth(apikey)
  > {ok: true}
}

@ GET /limited {
  + ratelimit(2/min)
  > {ok: true}
}`)

	if status, body := request("GET", "/guarded", ""); status != 401 {
		t.Errorf("auth(apikey) with no keys configured must deny: got %d\nbody: %s", status, body)
	}

	statuses := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		status, _ := request("GET", "/limited", "")
		statuses = append(statuses, status)
	}
	if statuses[0] != 200 || statuses[1] != 200 || statuses[2] != 429 {
		t.Errorf("ratelimit(2/min) should allow two then reject: got %v", statuses)
	}
}

// TestSymbolProviderInjection covers % by writing through the mock database
// and reading the record back.
func TestSymbolProviderInjection(t *testing.T) {
	if testing.Short() {
		t.Skip("boots a server; skipped in -short")
	}

	request, _ := serve(t, buildGlyphBinary(t), `: Note {
  id: int!
  body: str!
}

@ POST /notes {
  < input: Note
  % db: Database
  $ saved = db.notes.Create(input)
  > saved :: 201
}

@ GET /notes/:id {
  % db: Database
  $ note = db.notes.Get(id)
  ? note != null :: 404 "not found"
  > note
}`)

	if status, body := request("POST", "/notes", `{"id":1,"body":"written"}`); status != 201 {
		t.Fatalf("provider write failed: got %d\nbody: %s", status, body)
	}

	status, body := request("GET", "/notes/1", "")
	if status != 200 {
		t.Fatalf("provider read failed: got %d\nbody: %s", status, body)
	}
	if !strings.Contains(body, "written") {
		t.Errorf("record did not round-trip through the provider: %s", body)
	}
}

// TestSymbolCommand covers ! by invoking a declared command through
// `glyph exec`, which needs no server.
func TestSymbolCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the binary; skipped in -short")
	}

	binary := buildGlyphBinary(t)
	file := writeProgram(t, `! greet name: str! {
  > {message: "hello, " + name}
}`)

	out, err := exec.Command(binary, "exec", file, "greet", "--name=ada").CombinedOutput()
	if err != nil {
		t.Fatalf("glyph exec failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "hello, ada") {
		t.Errorf("command did not run: %s", out)
	}
}

// TestInertSymbolsWarnAtStartup is the honest half. Cron tasks, event handlers
// and queue workers do nothing at runtime, so the one behaviour to guarantee is
// that the server says so instead of implying the work is scheduled.
func TestInertSymbolsWarnAtStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("boots a server; skipped in -short")
	}

	for symbol, reason := range inertSymbols {
		t.Logf("symbol %q is inert: %s", symbol, reason)
	}

	_, logs := serve(t, buildGlyphBinary(t), `* "0 0 * * *" nightly {
  $ x = 1
}

~ "user.created" {
  $ y = 2
}

& "emails" {
  $ z = 3
}

@ GET /ping {
  > {ok: true}
}`)

	output := logs()
	for _, expected := range []string{
		"cron task declared",
		"event handler declared",
		"queue worker declared",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("startup did not warn %q; an operator would assume it runs\nserver log:\n%s",
				expected, output)
		}
	}
}
