package tests

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Examples are the project's showcase surface, but nothing executed them: only
// two were exercised by tests, and CI never ran any. That is how a route that
// answered "Internal server error" on every request, and a provider that was
// never wired at all, both shipped while `glyph validate` reported success.
//
// This boots each example with the real binary and requests its parameter-free
// GET routes. It asserts only that nothing returns 5xx - a 401 from an auth
// directive or a 404 from an empty mock database is a working route.

// knownBroken lists examples whose routes still fail, with the reason each one
// fails. They are skipped so this test can guard the 14 that work; an entry
// here is coverage we do not have yet, not a route that is fine.
//
// Three themes, none of them a wiring problem: examples calling functions that
// do not exist, examples reading a query parameter that was not supplied, and
// examples returning objects missing a required field.
var knownBroken = map[string]string{
	"mongodb-demo":      "mongo.Collection(x).InsertOne(y): chained provider calls are not dispatched",
	"auth-demo":         "GET /api/auth/me returns an object missing the required field id",
	"blog-api":          "calls paginate(), which is not a builtin",
	"blog-api-complete": "calls timestamp(), which is not a builtin",
	"e-commerce-api":    "calls timestamp(), which is not a builtin",
	"macros-demo":       "calls db.insert(), which the database provider does not expose",
	"cart-api":          "calls filter() with 3 arguments; it takes 2 (array, function)",
	"feature-showcase":  "calls filter() with 3 arguments; it takes 2 (array, function)",
	"database-demo":     "calls length() on a table handler rather than a collection",
}

var paramFreeGetRoute = regexp.MustCompile(`(?m)^@\s+GET\s+(/[^\s:{]*)\s*(->[^{]*)?\{`)

func TestExamplesServeWithoutServerErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("boots a server per example; skipped in -short")
	}

	binary := buildGlyphBinary(t)

	files, err := filepath.Glob(filepath.Join("..", "examples", "*", "main.glyph"))
	if err != nil {
		t.Fatalf("globbing examples: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no examples found - the glob or layout changed")
	}

	for _, file := range files {
		name := filepath.Base(filepath.Dir(file))

		if reason, skip := knownBroken[name]; skip {
			t.Run(name, func(t *testing.T) { t.Skipf("known broken: %s", reason) })
			continue
		}

		routes := paramFreeGetRoutes(t, file)
		if len(routes) == 0 {
			continue // nothing this test can drive without inventing inputs
		}

		t.Run(name, func(t *testing.T) {
			port := freePort(t)
			stop, logs := startExample(t, binary, file, port)
			defer stop()

			for _, route := range routes {
				url := fmt.Sprintf("http://127.0.0.1:%d%s", port, route)
				resp, err := http.Get(url)
				if err != nil {
					t.Errorf("GET %s: %v\nserver log:\n%s", route, err, logs())
					continue
				}
				resp.Body.Close()

				if resp.StatusCode >= 500 {
					t.Errorf("GET %s returned %d, want < 500\nserver log:\n%s",
						route, resp.StatusCode, logs())
				}
			}
		})
	}
}

// paramFreeGetRoutes returns GET paths that take no path parameter, so they can
// be requested without inventing an id that exists.
func paramFreeGetRoutes(t *testing.T, file string) []string {
	t.Helper()

	source, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("reading %s: %v", file, err)
	}

	var routes []string
	for _, match := range paramFreeGetRoute.FindAllStringSubmatch(string(source), -1) {
		path := strings.TrimSpace(match[1])
		if path != "" && !strings.Contains(path, ":") {
			routes = append(routes, path)
		}
	}
	return routes
}

func buildGlyphBinary(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "glyph")
	if isWindows() {
		binary += ".exe"
	}

	build := exec.Command("go", "build", "-o", binary, "./cmd/glyph")
	build.Dir = ".."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building glyph: %v\n%s", err, out)
	}
	return binary
}

// startExample runs one example and waits for it to accept connections.
// Returns a stop function and an accessor for the server's output, which is
// what explains a failure.
func startExample(t *testing.T, binary, file string, port int, extraArgs ...string) (func(), func() string) {
	t.Helper()

	var output strings.Builder
	// No cmd.Dir: file is already relative to this package's directory, and
	// examples resolve their module imports from their own path.
	args := append([]string{"run", file, "--port", fmt.Sprint(port)}, extraArgs...)
	cmd := exec.Command(binary, args...)
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		t.Fatalf("starting %s: %v", file, err)
	}

	stop := func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
		}
	}
	logs := func() string { return output.String() }

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return stop, logs
		}
		time.Sleep(100 * time.Millisecond)
	}

	stop()
	t.Fatalf("%s did not listen on port %d within 10s\nserver log:\n%s", file, port, output.String())
	return stop, logs
}

// freePort asks the OS for an unused port and releases it. The server binds it
// moments later; a collision would need another process to take the same port
// in that window.
func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserving a port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}
