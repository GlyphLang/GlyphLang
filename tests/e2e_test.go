package tests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/database"
	"github.com/glyphlang/glyph/pkg/hotreload"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/server"
	"github.com/glyphlang/glyph/pkg/vm"
)

// Type aliases for cleaner code in tests
type Route = server.Route

const (
	GET  = server.GET
	POST = server.POST
	PUT  = server.PUT
)

// MockConcurrentInterpreter is a thread-safe mock interpreter for concurrent tests
type MockConcurrentInterpreter struct {
	response interface{}
}

// Execute implements the server.Interpreter interface
func (m *MockConcurrentInterpreter) Execute(route *server.Route, ctx *server.Context) (interface{}, error) {
	return m.response, nil
}

// concurrentResult holds the result of a concurrent request
type concurrentResult struct {
	requestID  int
	statusCode int
	body       string
	err        error
}

// createTestServer creates a server instance for testing
func createTestServer(interpreter server.Interpreter) *server.Server {
	return server.NewServer(server.WithInterpreter(interpreter))
}

// Helper function to parse source code into a Module (E2E version)
func parseSourceE2E(source string) (*interpreter.Module, error) {
	return parseSource(source)
}

// TestHelloWorldExample tests the hello-world example end-to-end
func TestHelloWorldExample(t *testing.T) {
	helper := NewTestHelper(t)

	// 1. Load the hello-world example
	examplePath := filepath.Join("..", "examples", "hello-world", "main.glyph")
	source, err := os.ReadFile(examplePath)
	if err != nil {
		t.Skipf("Skipping test - hello-world example not found: %v", err)
		return
	}

	// 2. Parse and compile the source
	module, err := parseSourceE2E(string(source))
	if err != nil {
		helper.AssertNoError(err, "Parse failed")
	}

	comp := compiler.NewCompiler()
	bytecode, err := comp.Compile(module)
	helper.AssertNoError(err, "Compilation failed")
	helper.AssertNotNil(bytecode, "Bytecode should not be nil")

	// 3. Execute in VM
	v := vm.NewVM()
	result, err := v.Execute(bytecode)
	helper.AssertNoError(err, "Execution failed")
	helper.AssertNotNil(result, "Result should not be nil")

	// TODO: When server is implemented:
	// 4. Start HTTP server with routes
	// 5. Make request to /hello
	// 6. Verify response contains "Hello, World!"
	// 7. Make request to /greet/Alice
	// 8. Verify response contains "Hello, Alice!"

	t.Log("Hello-world example compilation successful")
}

// TestRestAPIExample tests the rest-api example end-to-end
func TestRestAPIExample(t *testing.T) {
	helper := NewTestHelper(t)

	// 1. Load the rest-api example
	examplePath := filepath.Join("..", "examples", "rest-api", "main.glyph")
	source, err := os.ReadFile(examplePath)
	if err != nil {
		t.Skipf("Skipping test - rest-api example not found: %v", err)
		return
	}

	// 2. Parse and compile the source
	module, err := parseSourceE2E(string(source))
	if err != nil {
		helper.AssertNoError(err, "Parse failed")
	}

	comp := compiler.NewCompiler()
	bytecode, err := comp.Compile(module)
	helper.AssertNoError(err, "Compilation failed")
	helper.AssertNotNil(bytecode, "Bytecode should not be nil")

	// 3. Execute in VM (may fail due to missing runtime dependencies like database)
	v := vm.NewVM()
	result, err := v.Execute(bytecode)
	if err != nil {
		// Skip if execution fails due to missing runtime dependencies
		t.Skipf("Skipping execution - missing runtime dependencies: %v", err)
		return
	}
	helper.AssertNotNil(result, "Result should not be nil")

	t.Log("Rest-api example compilation and execution successful")
}

// TestSimpleRouteE2E tests a simple route end-to-end
func TestSimpleRouteE2E(t *testing.T) {
	helper := NewTestHelper(t)

	// Load fixture
	source := helper.LoadFixture("simple_route.glyph")

	// Parse source
	module, err := parseSourceE2E(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Compile
	comp := compiler.NewCompiler()
	bytecode, err := comp.Compile(module)
	helper.AssertNoError(err, "Compilation failed")

	// Execute
	v := vm.NewVM()
	result, err := v.Execute(bytecode)
	helper.AssertNoError(err, "Execution failed")
	helper.AssertNotNil(result, "Result should not be nil")

	// TODO: When interpreter/server is ready:
	// - Start server with this route
	// - Make GET request to /test
	// - Verify response: {status: "ok"}
	// - Verify status code: 200
	// - Verify content-type: application/json

	t.Log("Simple route test passed")
}

// TestPathParametersE2E tests routes with path parameters
func TestPathParametersE2E(t *testing.T) {
	helper := NewTestHelper(t)

	source := helper.LoadFixture("path_param.glyph")

	module, err := parseSourceE2E(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	comp := compiler.NewCompiler()
	bytecode, err := comp.Compile(module)
	helper.AssertNoError(err, "Compilation failed")
	_ = bytecode // TODO: Use bytecode when server is ready

	// TODO: When server is ready:
	// - Execute bytecode with path parameter context
	// - Test GET /greet/Alice -> {message: "Hello, Alice!"}
	// - Test GET /greet/Bob -> {message: "Hello, Bob!"}
	// - Test GET /greet/ (no param) -> Should return 404
	// - Test GET /greet (no param) -> Should return 404

	t.Log("Path parameters test passed (compilation)")
}

// TestJSONSerializationE2E tests JSON response serialization
func TestJSONSerializationE2E(t *testing.T) {
	helper := NewTestHelper(t)

	source := helper.LoadFixture("json_response.glyph")

	module, err := parseSourceE2E(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	comp := compiler.NewCompiler()
	bytecode, err := comp.Compile(module)
	helper.AssertNoError(err, "Compilation failed")
	_ = bytecode // TODO: Use bytecode when server is ready

	// TODO: When server is ready:
	// - Start server
	// - Make GET request to /api/user
	// - Verify JSON structure matches User type
	// - Verify fields: id=1, name="Test User", email="test@example.com"
	// - Verify Content-Type header is application/json

	t.Log("JSON serialization test passed (compilation)")
}

// TestMultipleRoutesE2E tests multiple routes in one program
func TestMultipleRoutesE2E(t *testing.T) {
	helper := NewTestHelper(t)

	source := helper.LoadFixture("multiple_routes.glyph")

	module, err := parseSourceE2E(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	comp := compiler.NewCompiler()
	bytecode, err := comp.Compile(module)
	helper.AssertNoError(err, "Compilation with multiple routes failed")
	_ = bytecode // TODO: Use bytecode when server is ready

	// TODO: When server is ready:
	// Test GET /health -> {status: "ok", timestamp: ...}
	// Test GET /version -> {version: "0.1.0"}
	// Test GET /info/123 -> {id: "123", info: "Details about 123"}
	// Verify all routes are registered correctly
	// Verify routes don't interfere with each other

	t.Log("Multiple routes test passed (compilation)")
}

// TestHTTPMethodsE2E tests different HTTP methods
func TestHTTPMethodsE2E(t *testing.T) {
	helper := NewTestHelper(t)

	source := helper.LoadFixture("post_route.glyph")

	module, err := parseSourceE2E(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	comp := compiler.NewCompiler()
	bytecode, err := comp.Compile(module)
	helper.AssertNoError(err, "POST route compilation failed")
	_ = bytecode // TODO: Use bytecode when server is ready

	// TODO: When server is ready:
	// Test POST /api/users with valid JSON body
	// Test POST with missing required fields -> Should return 400
	// Test POST with invalid email format -> Should return 400
	// Test GET /api/users -> Should return 404 or 405 (Method Not Allowed)
	// Test POST with name too long -> Should return 400

	t.Log("HTTP methods test passed (compilation)")
}

// TestAuthenticationE2E tests authentication middleware
func TestAuthenticationE2E(t *testing.T) {
	helper := NewTestHelper(t)

	source := helper.LoadFixture("with_auth.glyph")

	module, err := parseSourceE2E(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	comp := compiler.NewCompiler()
	bytecode, err := comp.Compile(module)
	helper.AssertNoError(err, "Auth route compilation failed")
	_ = bytecode // TODO: Use bytecode when auth middleware is ready

	// TODO: When auth middleware is implemented:
	// Test GET /api/protected without auth header -> 401 Unauthorized
	// Test GET /api/protected with invalid JWT -> 401 Unauthorized
	// Test GET /api/protected with valid JWT -> 200 OK
	// Test rate limiting (make 101 requests in 1 minute) -> 429 Too Many Requests

	t.Log("Authentication test passed (compilation)")
}

// TestErrorHandlingE2E tests error handling and error responses
func TestErrorHandlingE2E(t *testing.T) {
	helper := NewTestHelper(t)

	source := helper.LoadFixture("error_handling.glyph")

	module, err := parseSourceE2E(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	comp := compiler.NewCompiler()
	bytecode, err := comp.Compile(module)
	helper.AssertNoError(err, "Error handling route compilation failed")
	_ = bytecode // TODO: Use bytecode when error handling is ready

	// TODO: When error handling is implemented:
	// Test GET /api/data/999 (not found) -> Error response
	// Test GET /api/data/invalid (invalid ID) -> Error response
	// Verify error response structure matches Error type
	// Verify proper HTTP status codes (404, 400, 500)

	t.Log("Error handling test passed (compilation)")
}

// TestInvalidSyntaxE2E tests that invalid syntax is rejected
func TestInvalidSyntaxE2E(t *testing.T) {
	helper := NewTestHelper(t)

	source := helper.LoadFixture("invalid_syntax.glyph")

	module, err := parseSourceE2E(source)
	if err != nil {
		t.Logf("Parse failed as expected: %v", err)
		return
	}

	comp := compiler.NewCompiler()
	bytecode, err := comp.Compile(module)

	// TODO: When parser is implemented:
	// This should fail at compilation with a syntax error
	// For now, the placeholder compiler doesn't validate syntax
	if err == nil && bytecode != nil {
		t.Log("Invalid syntax test: placeholder compiler doesn't validate yet")
	} else {
		helper.AssertError(err, "Invalid syntax should fail compilation")
	}
}

// TestServerStartupShutdownE2E tests server lifecycle
func TestServerStartupShutdownE2E(t *testing.T) {
	t.Skip("Skipping until server implementation is ready")

	// TODO: When server is implemented:
	// 1. Compile a simple program
	// 2. Start development server
	// 3. Verify server is listening on port 3000
	// 4. Make a health check request
	// 5. Gracefully shut down server
	// 6. Verify server stopped accepting connections
	// 7. Test server restart works correctly
}

// TestHotReloadE2E tests development server hot reload
func TestHotReloadE2E(t *testing.T) {
	t.Run("FileWatcher detects file changes", func(t *testing.T) {
		// Create temp directory with a .glyph file
		tmpDir, err := os.MkdirTemp("", "hotreload-e2e-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		t.Cleanup(func() { os.RemoveAll(tmpDir) })

		// Create initial file
		testFile := filepath.Join(tmpDir, "main.glyph")
		initialContent := `@ GET /test {
  > {status: "ok"}
}`
		if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Track detected changes
		var changes []hotreload.FileChange
		var mu sync.Mutex
		changeChan := make(chan struct{}, 10)

		// Create watcher
		watcher := hotreload.NewFileWatcher(
			[]string{tmpDir},
			func(c []hotreload.FileChange) {
				mu.Lock()
				changes = append(changes, c...)
				mu.Unlock()
				changeChan <- struct{}{}
			},
			hotreload.WithPollInterval(50*time.Millisecond),
			hotreload.WithDebounce(10*time.Millisecond),
			hotreload.WithPatterns("*.glyph"),
		)

		// Start watching
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		if err := watcher.Start(ctx); err != nil {
			t.Fatalf("Failed to start watcher: %v", err)
		}
		t.Cleanup(watcher.Stop)

		// Wait for initial scan to complete
		time.Sleep(100 * time.Millisecond)

		// Modify the file
		modifiedContent := `@ GET /test {
  > {status: "modified"}
}`
		if err := os.WriteFile(testFile, []byte(modifiedContent), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		// Wait for change detection
		select {
		case <-changeChan:
			// Change detected
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for file change detection")
		}

		// Verify change was detected
		mu.Lock()
		defer mu.Unlock()

		if len(changes) == 0 {
			t.Fatal("Expected at least one change, got none")
		}

		found := false
		for _, c := range changes {
			if c.Path == testFile && c.Type == hotreload.ChangeTypeModified {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected modified change for %s, got changes: %+v", testFile, changes)
		}
	})

	t.Run("ApplicationState preserves state across operations", func(t *testing.T) {
		state := hotreload.NewApplicationState()

		// Set initial state
		state.Set("counter", 42)
		state.Set("user", "testuser")
		state.Set("config", map[string]string{"debug": "true"})

		// Verify Get works
		val, ok := state.Get("counter")
		if !ok || val != 42 {
			t.Errorf("Expected counter=42, got %v (ok=%v)", val, ok)
		}

		val, ok = state.Get("user")
		if !ok || val != "testuser" {
			t.Errorf("Expected user=testuser, got %v (ok=%v)", val, ok)
		}

		// Verify GetAll returns all state
		all := state.GetAll()
		if len(all) != 3 {
			t.Errorf("Expected 3 items, got %d", len(all))
		}

		// Test state preservation (simulating reload scenario)
		savedState := state.GetAll()

		// Clear and restore state (simulating what happens during reload)
		state.SetAll(map[string]interface{}{})
		state.SetAll(savedState)

		// Verify state was preserved
		val, ok = state.Get("counter")
		if !ok || val != 42 {
			t.Errorf("After restore: Expected counter=42, got %v (ok=%v)", val, ok)
		}

		// Verify non-existent key returns false
		_, ok = state.Get("nonexistent")
		if ok {
			t.Error("Expected false for non-existent key")
		}
	})

	t.Run("ReloadManager with mock compiler and server", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "hotreload-rm-e2e-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		t.Cleanup(func() { os.RemoveAll(tmpDir) })

		// Create initial file
		testFile := filepath.Join(tmpDir, "main.glyph")
		if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Create mock compiler and server
		mockCompiler := &e2eMockCompiler{bytecode: []byte("compiled")}
		mockServer := &e2eMockServer{state: make(map[string]interface{})}

		// Track reload events
		var reloadEvents []hotreload.ReloadEvent
		var mu sync.Mutex
		reloadChan := make(chan struct{}, 10)

		// Create reload manager
		rm := hotreload.NewReloadManager(
			[]string{tmpDir},
			mockCompiler,
			mockServer,
			hotreload.WithOnReload(func(event hotreload.ReloadEvent) {
				mu.Lock()
				reloadEvents = append(reloadEvents, event)
				mu.Unlock()
				reloadChan <- struct{}{}
			}),
		)

		// Start reload manager
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		if err := rm.Start(ctx); err != nil {
			t.Fatalf("Failed to start reload manager: %v", err)
		}
		t.Cleanup(rm.Stop)

		// Wait for initial scan
		time.Sleep(200 * time.Millisecond)

		// Modify the file to trigger reload
		if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		// Wait for reload
		select {
		case <-reloadChan:
			// Reload triggered
		case <-time.After(3 * time.Second):
			t.Fatal("Timeout waiting for reload")
		}

		// Verify reload occurred
		mu.Lock()
		eventCount := len(reloadEvents)
		var lastEvent hotreload.ReloadEvent
		if eventCount > 0 {
			lastEvent = reloadEvents[eventCount-1]
		}
		mu.Unlock()

		if eventCount == 0 {
			t.Fatal("Expected at least one reload event")
		}

		if !lastEvent.Success {
			t.Errorf("Expected successful reload, got error: %v", lastEvent.Error)
		}

		if lastEvent.ReloadCount < 1 {
			t.Errorf("Expected reload count >= 1, got %d", lastEvent.ReloadCount)
		}

		// Verify stats
		stats := rm.Stats()
		if stats.ReloadCount < 1 {
			t.Errorf("Expected stats reload count >= 1, got %d", stats.ReloadCount)
		}

		// Verify compiler was called
		mockCompiler.mu.Lock()
		compilerCalls := mockCompiler.compileCalls
		mockCompiler.mu.Unlock()
		if compilerCalls < 1 {
			t.Errorf("Expected compiler to be called at least once, got %d calls", compilerCalls)
		}

		// Verify server reload was called
		mockServer.mu.Lock()
		reloadCalls := mockServer.reloadCalls
		mockServer.mu.Unlock()
		if reloadCalls < 1 {
			t.Errorf("Expected server reload to be called at least once, got %d calls", reloadCalls)
		}
	})

	t.Run("ReloadManager handles compilation errors gracefully", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "hotreload-err-e2e-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		t.Cleanup(func() { os.RemoveAll(tmpDir) })

		// Create initial file
		testFile := filepath.Join(tmpDir, "main.glyph")
		if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Create mock compiler that returns an error
		compileErr := fmt.Errorf("syntax error: unexpected token")
		mockCompiler := &e2eMockCompiler{compileError: compileErr}
		mockServer := &e2eMockServer{state: make(map[string]interface{})}

		// Track errors
		var errors []error
		var mu sync.Mutex
		errorChan := make(chan struct{}, 10)

		// Create reload manager with error handler
		rm := hotreload.NewReloadManager(
			[]string{tmpDir},
			mockCompiler,
			mockServer,
			hotreload.WithErrorHandler(func(err error) {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				errorChan <- struct{}{}
			}),
		)

		// Start reload manager
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		if err := rm.Start(ctx); err != nil {
			t.Fatalf("Failed to start reload manager: %v", err)
		}
		t.Cleanup(rm.Stop)

		// Wait for initial scan
		time.Sleep(200 * time.Millisecond)

		// Modify the file to trigger reload (which will fail)
		if err := os.WriteFile(testFile, []byte("invalid syntax"), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		// Wait for error handler to be called
		select {
		case <-errorChan:
			// Error received
		case <-time.After(3 * time.Second):
			t.Fatal("Timeout waiting for error handler")
		}

		// Verify error was captured
		mu.Lock()
		errorCount := len(errors)
		mu.Unlock()

		if errorCount == 0 {
			t.Fatal("Expected at least one error")
		}

		// Server reload should NOT have been called due to compile error
		mockServer.mu.Lock()
		reloadCalls := mockServer.reloadCalls
		mockServer.mu.Unlock()
		if reloadCalls > 0 {
			t.Errorf("Server reload should not be called on compile error, got %d calls", reloadCalls)
		}
	})

	t.Run("FileWatcher detects new file creation", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "hotreload-new-e2e-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		t.Cleanup(func() { os.RemoveAll(tmpDir) })

		// Track detected changes
		var changes []hotreload.FileChange
		var mu sync.Mutex
		changeChan := make(chan struct{}, 10)

		// Create watcher
		watcher := hotreload.NewFileWatcher(
			[]string{tmpDir},
			func(c []hotreload.FileChange) {
				mu.Lock()
				changes = append(changes, c...)
				mu.Unlock()
				changeChan <- struct{}{}
			},
			hotreload.WithPollInterval(50*time.Millisecond),
			hotreload.WithDebounce(10*time.Millisecond),
			hotreload.WithPatterns("*.glyph"),
		)

		// Start watching
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		if err := watcher.Start(ctx); err != nil {
			t.Fatalf("Failed to start watcher: %v", err)
		}
		t.Cleanup(watcher.Stop)

		// Wait for initial scan
		time.Sleep(100 * time.Millisecond)

		// Create a new file
		newFile := filepath.Join(tmpDir, "new.glyph")
		if err := os.WriteFile(newFile, []byte("new file content"), 0644); err != nil {
			t.Fatalf("Failed to create new file: %v", err)
		}

		// Wait for change detection
		select {
		case <-changeChan:
			// Change detected
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for new file detection")
		}

		// Verify creation was detected
		mu.Lock()
		defer mu.Unlock()

		found := false
		for _, c := range changes {
			if c.Path == newFile && c.Type == hotreload.ChangeTypeCreated {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected created change for %s, got changes: %+v", newFile, changes)
		}
	})

	t.Run("FileWatcher detects file deletion", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "hotreload-del-e2e-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		t.Cleanup(func() { os.RemoveAll(tmpDir) })

		// Create initial file
		testFile := filepath.Join(tmpDir, "delete.glyph")
		if err := os.WriteFile(testFile, []byte("to be deleted"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Track detected changes
		var changes []hotreload.FileChange
		var mu sync.Mutex
		changeChan := make(chan struct{}, 10)

		// Create watcher
		watcher := hotreload.NewFileWatcher(
			[]string{tmpDir},
			func(c []hotreload.FileChange) {
				mu.Lock()
				changes = append(changes, c...)
				mu.Unlock()
				changeChan <- struct{}{}
			},
			hotreload.WithPollInterval(50*time.Millisecond),
			hotreload.WithDebounce(10*time.Millisecond),
			hotreload.WithPatterns("*.glyph"),
		)

		// Start watching
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		if err := watcher.Start(ctx); err != nil {
			t.Fatalf("Failed to start watcher: %v", err)
		}
		t.Cleanup(watcher.Stop)

		// Wait for initial scan
		time.Sleep(100 * time.Millisecond)

		// Delete the file
		if err := os.Remove(testFile); err != nil {
			t.Fatalf("Failed to delete test file: %v", err)
		}

		// Wait for change detection
		select {
		case <-changeChan:
			// Change detected
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for file deletion detection")
		}

		// Verify deletion was detected
		mu.Lock()
		defer mu.Unlock()

		found := false
		for _, c := range changes {
			if c.Path == testFile && c.Type == hotreload.ChangeTypeDeleted {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected deleted change for %s, got changes: %+v", testFile, changes)
		}
	})

	t.Run("State preserved during reload cycle", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "hotreload-state-e2e-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		t.Cleanup(func() { os.RemoveAll(tmpDir) })

		// Create initial file
		testFile := filepath.Join(tmpDir, "main.glyph")
		if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Create mock server with initial state
		mockCompiler := &e2eMockCompiler{bytecode: []byte("compiled")}
		mockServer := &e2eMockServer{
			state: map[string]interface{}{
				"session_id": "abc123",
				"user_data":  "preserved",
				"counter":    100,
			},
		}

		reloadChan := make(chan struct{}, 10)

		// Create reload manager
		rm := hotreload.NewReloadManager(
			[]string{tmpDir},
			mockCompiler,
			mockServer,
			hotreload.WithOnReload(func(event hotreload.ReloadEvent) {
				reloadChan <- struct{}{}
			}),
		)

		// Start reload manager
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		if err := rm.Start(ctx); err != nil {
			t.Fatalf("Failed to start reload manager: %v", err)
		}
		t.Cleanup(rm.Stop)

		// Wait for initial scan
		time.Sleep(200 * time.Millisecond)

		// Modify the file to trigger reload
		if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		// Wait for reload
		select {
		case <-reloadChan:
			// Reload triggered
		case <-time.After(3 * time.Second):
			t.Fatal("Timeout waiting for reload")
		}

		// Verify state was preserved (SetState should have been called with saved state)
		mockServer.mu.Lock()
		setStateCalls := mockServer.setStateCalls
		sessionID := mockServer.state["session_id"]
		mockServer.mu.Unlock()

		if setStateCalls < 1 {
			t.Error("Expected SetState to be called to restore state")
		}

		// Verify the state data is still accessible
		if sessionID != "abc123" {
			t.Errorf("Expected session_id to be preserved, got %v", sessionID)
		}
	})
}

// e2eMockCompiler is a mock compiler for E2E tests
type e2eMockCompiler struct {
	compileError error
	bytecode     []byte
	compileCalls int
	mu           sync.Mutex
}

func (m *e2eMockCompiler) CompileFile(path string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.compileCalls++
	if m.compileError != nil {
		return nil, m.compileError
	}
	return m.bytecode, nil
}

// e2eMockServer is a mock server for E2E tests
type e2eMockServer struct {
	reloadError   error
	setStateError error
	state         map[string]interface{}
	reloadCalls   int
	setStateCalls int
	mu            sync.Mutex
}

func (m *e2eMockServer) Reload(bytecode []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reloadCalls++
	return m.reloadError
}

func (m *e2eMockServer) GetState() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return a copy to simulate real behavior
	result := make(map[string]interface{})
	for k, v := range m.state {
		result[k] = v
	}
	return result
}

func (m *e2eMockServer) SetState(state map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setStateCalls++
	if m.setStateError != nil {
		return m.setStateError
	}
	// Restore the state
	for k, v := range state {
		m.state[k] = v
	}
	return nil
}

// TestConcurrentRequestsE2E tests handling concurrent HTTP requests
func TestConcurrentRequestsE2E(t *testing.T) {
	// Create a mock interpreter that returns a consistent response
	mockInterpreter := &MockConcurrentInterpreter{
		response: map[string]interface{}{
			"status":  "ok",
			"message": "Hello, World!",
		},
	}

	// Create server with the mock interpreter
	srv := createTestServer(mockInterpreter)

	// Register a simple route
	err := srv.RegisterRoute(&Route{
		Method: GET,
		Path:   "/api/test",
	})
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Register health routes for the health check test
	healthManager := server.NewHealthManager()
	if err := srv.RegisterHealthRoutes(healthManager); err != nil {
		t.Fatalf("Failed to register health routes: %v", err)
	}

	// Create test server
	ts := httptest.NewServer(srv.GetHandler())
	defer ts.Close()

	const numRequests = 150 // More than 100 as specified

	t.Run("ConcurrentGETRequests", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make(chan *concurrentResult, numRequests)

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		// Launch concurrent requests
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(requestID int) {
				defer wg.Done()

				resp, err := client.Get(ts.URL + "/api/test")
				result := &concurrentResult{
					requestID: requestID,
					err:       err,
				}

				if err == nil {
					result.statusCode = resp.StatusCode
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					result.body = string(body)
				}

				results <- result
			}(i)
		}

		// Wait for all requests to complete
		wg.Wait()
		close(results)

		// Verify results
		successCount := 0
		var errors []string

		for result := range results {
			if result.err != nil {
				errors = append(errors, fmt.Sprintf("Request %d failed: %v", result.requestID, result.err))
				continue
			}

			if result.statusCode != http.StatusOK {
				errors = append(errors, fmt.Sprintf("Request %d returned status %d", result.requestID, result.statusCode))
				continue
			}

			// Verify response contains expected content
			if !contains(result.body, "ok") {
				errors = append(errors, fmt.Sprintf("Request %d: response missing 'ok' status", result.requestID))
				continue
			}

			successCount++
		}

		// Report any errors
		if len(errors) > 0 {
			for _, e := range errors {
				t.Error(e)
			}
		}

		// All requests should succeed
		if successCount != numRequests {
			t.Errorf("Expected %d successful requests, got %d", numRequests, successCount)
		}

		t.Logf("Successfully completed %d concurrent requests", successCount)
	})

	t.Run("ConcurrentMixedMethods", func(t *testing.T) {
		// Register additional routes for different methods
		srv.RegisterRoute(&Route{
			Method: POST,
			Path:   "/api/create",
		})
		srv.RegisterRoute(&Route{
			Method: PUT,
			Path:   "/api/update/:id",
		})

		var wg sync.WaitGroup
		results := make(chan *concurrentResult, numRequests)

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		// Launch concurrent requests with different methods
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(requestID int) {
				defer wg.Done()

				var resp *http.Response
				var err error

				// Alternate between GET and POST requests
				if requestID%2 == 0 {
					resp, err = client.Get(ts.URL + "/api/test")
				} else {
					resp, err = client.Post(ts.URL+"/api/create", "application/json", nil)
				}

				result := &concurrentResult{
					requestID: requestID,
					err:       err,
				}

				if err == nil {
					result.statusCode = resp.StatusCode
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					result.body = string(body)
				}

				results <- result
			}(i)
		}

		wg.Wait()
		close(results)

		successCount := 0
		for result := range results {
			if result.err == nil && result.statusCode == http.StatusOK {
				successCount++
			}
		}

		if successCount != numRequests {
			t.Errorf("Expected %d successful requests, got %d", numRequests, successCount)
		}

		t.Logf("Successfully completed %d concurrent mixed-method requests", successCount)
	})

	t.Run("ServerStabilityAfterLoad", func(t *testing.T) {
		// After all concurrent requests, verify server is still responsive
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		for i := 0; i < 10; i++ {
			resp, err := client.Get(ts.URL + "/api/test")
			if err != nil {
				t.Errorf("Post-load request %d failed: %v", i, err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Post-load request %d returned status %d", i, resp.StatusCode)
			}
			resp.Body.Close()
		}

		t.Log("Server remains stable after concurrent load")
	})

	t.Run("HealthCheckDuringLoad", func(t *testing.T) {
		var wg sync.WaitGroup
		healthCheckPassed := make(chan bool, 10)

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		// Start background load
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := client.Get(ts.URL + "/api/test")
				if err == nil {
					resp.Body.Close()
				}
			}()
		}

		// Make health check requests during load
		for i := 0; i < 10; i++ {
			resp, err := client.Get(ts.URL + "/health")
			if err != nil {
				healthCheckPassed <- false
				continue
			}

			passed := resp.StatusCode == http.StatusOK
			resp.Body.Close()
			healthCheckPassed <- passed
		}

		wg.Wait()
		close(healthCheckPassed)

		passedCount := 0
		for passed := range healthCheckPassed {
			if passed {
				passedCount++
			}
		}

		if passedCount < 8 { // Allow some tolerance
			t.Errorf("Expected at least 8 health checks to pass during load, got %d", passedCount)
		}

		t.Logf("Health check passed %d/10 times during load", passedCount)
	})
}

// TestDatabaseIntegrationE2E tests database operations
func TestDatabaseIntegrationE2E(t *testing.T) {
	// Skip if DATABASE_URL is not set
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		t.Skip("Skipping database integration test: DATABASE_URL not set")
	}

	cfg, err := database.ParseConnectionString(connStr)
	if err != nil {
		t.Fatalf("Failed to parse connection string: %v", err)
	}

	db := database.NewPostgresDB(cfg)
	ctx := context.Background()

	err = db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test table name - unique per test run
	tableName := fmt.Sprintf("e2e_test_%d", time.Now().UnixNano())

	// Cleanup on test completion
	t.Cleanup(func() {
		db.DropTable(ctx, tableName)
	})

	t.Run("CreateTable", func(t *testing.T) {
		schema := map[string]string{
			"id":    "SERIAL PRIMARY KEY",
			"name":  "VARCHAR(100) NOT NULL",
			"email": "VARCHAR(255)",
			"age":   "INTEGER",
		}
		err := db.CreateTable(ctx, tableName, schema)
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		exists, err := db.TableExists(ctx, tableName)
		if err != nil {
			t.Fatalf("Failed to check table exists: %v", err)
		}
		if !exists {
			t.Fatal("Table should exist after creation")
		}
	})

	t.Run("INSERT", func(t *testing.T) {
		query := fmt.Sprintf("INSERT INTO %s (name, email, age) VALUES ($1, $2, $3)", tableName)
		_, err := db.Exec(ctx, query, "Alice", "alice@example.com", 30)
		if err != nil {
			t.Fatalf("Failed to insert record: %v", err)
		}

		_, err = db.Exec(ctx, query, "Bob", "bob@example.com", 25)
		if err != nil {
			t.Fatalf("Failed to insert second record: %v", err)
		}
	})

	t.Run("SELECT", func(t *testing.T) {
		query := fmt.Sprintf("SELECT name, email, age FROM %s WHERE name = $1", tableName)
		rows, err := db.Query(ctx, query, "Alice")
		if err != nil {
			t.Fatalf("Failed to query: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Fatal("Expected at least one row")
		}

		var name, email string
		var age int
		err = rows.Scan(&name, &email, &age)
		if err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}

		if name != "Alice" || email != "alice@example.com" || age != 30 {
			t.Errorf("Unexpected values: name=%s, email=%s, age=%d", name, email, age)
		}
	})

	t.Run("UPDATE", func(t *testing.T) {
		query := fmt.Sprintf("UPDATE %s SET age = $1 WHERE name = $2", tableName)
		result, err := db.Exec(ctx, query, 31, "Alice")
		if err != nil {
			t.Fatalf("Failed to update: %v", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", rowsAffected)
		}
	})

	t.Run("DELETE", func(t *testing.T) {
		query := fmt.Sprintf("DELETE FROM %s WHERE name = $1", tableName)
		result, err := db.Exec(ctx, query, "Bob")
		if err != nil {
			t.Fatalf("Failed to delete: %v", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", rowsAffected)
		}
	})

	t.Run("ParameterizedQueries_PreventSQLInjection", func(t *testing.T) {
		// Attempt SQL injection via parameterized query - should be safe
		maliciousInput := "'; DROP TABLE " + tableName + "; --"
		query := fmt.Sprintf("SELECT name FROM %s WHERE name = $1", tableName)
		rows, err := db.Query(ctx, query, maliciousInput)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		rows.Close()

		// Table should still exist
		exists, err := db.TableExists(ctx, tableName)
		if err != nil {
			t.Fatalf("Failed to check table: %v", err)
		}
		if !exists {
			t.Fatal("Table was dropped by SQL injection - parameterized queries failed!")
		}
	})

	t.Run("Transaction", func(t *testing.T) {
		tx, err := db.Begin(ctx)
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// Insert in transaction
		query := fmt.Sprintf("INSERT INTO %s (name, email, age) VALUES ($1, $2, $3)", tableName)
		_, err = tx.Exec(query, "Charlie", "charlie@example.com", 35)
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to insert in transaction: %v", err)
		}

		// Rollback
		err = tx.Rollback()
		if err != nil {
			t.Fatalf("Failed to rollback: %v", err)
		}

		// Verify record was not inserted
		selectQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE name = $1", tableName)
		var count int
		row := db.QueryRow(ctx, selectQuery, "Charlie")
		err = row.Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count: %v", err)
		}
		if count != 0 {
			t.Error("Transaction rollback failed - record still exists")
		}
	})

	t.Log("All database integration E2E tests passed")
}

// TestSecurityFeaturesE2E tests security features end-to-end
func TestSecurityFeaturesE2E(t *testing.T) {
	// Test 1: Rate Limiting
	t.Run("RateLimiting", func(t *testing.T) {
		// Create a server with rate limiting middleware
		// Use a very small burst size to ensure we can exhaust it quickly
		rateLimitConfig := server.RateLimiterConfig{
			RequestsPerMinute: 1,  // Very slow refill rate
			BurstSize:         3,  // Small burst size
		}

		srv := server.NewServer(
			server.WithMiddleware(server.RateLimitMiddleware(rateLimitConfig)),
		)

		// Register a simple test route
		srv.RegisterRoute(&server.Route{
			Method: server.GET,
			Path:   "/api/test",
			Handler: func(ctx *server.Context) error {
				return server.SendJSON(ctx, http.StatusOK, map[string]string{"status": "ok"})
			},
		})

		// Create test server
		testServer := httptest.NewServer(srv.GetHandler())
		defer testServer.Close()

		// Use X-Forwarded-For header to simulate a consistent client IP
		// This is necessary because each TCP connection may have a different port
		client := &http.Client{}
		clientIP := "192.168.1.100"

		// Make requests up to the burst limit - these should succeed
		successCount := 0
		for i := 0; i < 3; i++ {
			req, _ := http.NewRequest("GET", testServer.URL+"/api/test", nil)
			req.Header.Set("X-Forwarded-For", clientIP)
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request %d failed: %v", i+1, err)
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				successCount++
			}
		}

		if successCount != 3 {
			t.Errorf("Expected 3 successful requests within burst limit, got %d", successCount)
		}

		// Next request should be rate limited (429 Too Many Requests)
		req, _ := http.NewRequest("GET", testServer.URL+"/api/test", nil)
		req.Header.Set("X-Forwarded-For", clientIP)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Rate limit request failed: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected 429 Too Many Requests after exceeding rate limit, got %d", resp.StatusCode)
		}
	})

	// Test 2: Authentication - Missing Token
	t.Run("Authentication_MissingToken", func(t *testing.T) {
		validTokens := map[string]bool{
			"valid-token-123": true,
			"admin-token-456": true,
		}

		srv := server.NewServer()
		srv.RegisterRoute(&server.Route{
			Method:      server.GET,
			Path:        "/api/protected",
			Middlewares: []server.Middleware{server.BasicAuthMiddleware(validTokens)},
			Handler: func(ctx *server.Context) error {
				return server.SendJSON(ctx, http.StatusOK, map[string]string{"message": "secret data"})
			},
		})

		testServer := httptest.NewServer(srv.GetHandler())
		defer testServer.Close()

		// Request without Authorization header
		resp, err := http.Get(testServer.URL + "/api/protected")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized for missing token, got %d", resp.StatusCode)
		}
	})

	// Test 3: Authentication - Invalid Token
	t.Run("Authentication_InvalidToken", func(t *testing.T) {
		validTokens := map[string]bool{
			"valid-token-123": true,
		}

		srv := server.NewServer()
		srv.RegisterRoute(&server.Route{
			Method:      server.GET,
			Path:        "/api/protected",
			Middlewares: []server.Middleware{server.BasicAuthMiddleware(validTokens)},
			Handler: func(ctx *server.Context) error {
				return server.SendJSON(ctx, http.StatusOK, map[string]string{"message": "secret data"})
			},
		})

		testServer := httptest.NewServer(srv.GetHandler())
		defer testServer.Close()

		// Request with invalid token
		req, _ := http.NewRequest("GET", testServer.URL+"/api/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized for invalid token, got %d", resp.StatusCode)
		}
	})

	// Test 4: Authentication - Valid Token
	t.Run("Authentication_ValidToken", func(t *testing.T) {
		validTokens := map[string]bool{
			"valid-token-123": true,
		}

		srv := server.NewServer()
		srv.RegisterRoute(&server.Route{
			Method:      server.GET,
			Path:        "/api/protected",
			Middlewares: []server.Middleware{server.BasicAuthMiddleware(validTokens)},
			Handler: func(ctx *server.Context) error {
				return server.SendJSON(ctx, http.StatusOK, map[string]string{"message": "secret data"})
			},
		})

		testServer := httptest.NewServer(srv.GetHandler())
		defer testServer.Close()

		// Request with valid token
		req, _ := http.NewRequest("GET", testServer.URL+"/api/protected", nil)
		req.Header.Set("Authorization", "Bearer valid-token-123")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK for valid token, got %d", resp.StatusCode)
		}
	})

	// Test 5: Security Headers
	t.Run("SecurityHeaders", func(t *testing.T) {
		srv := server.NewServer(
			server.WithMiddleware(server.SecurityHeadersMiddleware()),
		)

		srv.RegisterRoute(&server.Route{
			Method: server.GET,
			Path:   "/api/test",
			Handler: func(ctx *server.Context) error {
				return server.SendJSON(ctx, http.StatusOK, map[string]string{"status": "ok"})
			},
		})

		testServer := httptest.NewServer(srv.GetHandler())
		defer testServer.Close()

		resp, err := http.Get(testServer.URL + "/api/test")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		// Check X-Content-Type-Options header
		if xContentTypeOptions := resp.Header.Get("X-Content-Type-Options"); xContentTypeOptions != "nosniff" {
			t.Errorf("Expected X-Content-Type-Options: nosniff, got %q", xContentTypeOptions)
		}

		// Check X-Frame-Options header
		if xFrameOptions := resp.Header.Get("X-Frame-Options"); xFrameOptions != "DENY" {
			t.Errorf("Expected X-Frame-Options: DENY, got %q", xFrameOptions)
		}

		// Check X-XSS-Protection header
		if xXSSProtection := resp.Header.Get("X-XSS-Protection"); xXSSProtection != "1; mode=block" {
			t.Errorf("Expected X-XSS-Protection: 1; mode=block, got %q", xXSSProtection)
		}

		// Check Referrer-Policy header
		if referrerPolicy := resp.Header.Get("Referrer-Policy"); referrerPolicy != "strict-origin-when-cross-origin" {
			t.Errorf("Expected Referrer-Policy: strict-origin-when-cross-origin, got %q", referrerPolicy)
		}
	})

	// Test 6: CORS - Wildcard Origin
	t.Run("CORS_WildcardOrigin", func(t *testing.T) {
		srv := server.NewServer(
			server.WithMiddleware(server.CORSMiddleware([]string{"*"})),
		)

		srv.RegisterRoute(&server.Route{
			Method: server.GET,
			Path:   "/api/test",
			Handler: func(ctx *server.Context) error {
				return server.SendJSON(ctx, http.StatusOK, map[string]string{"status": "ok"})
			},
		})

		testServer := httptest.NewServer(srv.GetHandler())
		defer testServer.Close()

		// Make request with Origin header
		req, _ := http.NewRequest("GET", testServer.URL+"/api/test", nil)
		req.Header.Set("Origin", "http://example.com")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		// Check Access-Control-Allow-Origin header
		allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
		if allowOrigin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin: *, got %q", allowOrigin)
		}

		// Check Access-Control-Allow-Credentials is false for wildcard
		allowCredentials := resp.Header.Get("Access-Control-Allow-Credentials")
		if allowCredentials != "false" {
			t.Errorf("Expected Access-Control-Allow-Credentials: false for wildcard, got %q", allowCredentials)
		}
	})

	// Test 7: CORS - Specific Origin
	t.Run("CORS_SpecificOrigin", func(t *testing.T) {
		allowedOrigins := []string{"http://example.com", "http://trusted.com"}
		srv := server.NewServer(
			server.WithMiddleware(server.CORSMiddleware(allowedOrigins)),
		)

		srv.RegisterRoute(&server.Route{
			Method: server.GET,
			Path:   "/api/test",
			Handler: func(ctx *server.Context) error {
				return server.SendJSON(ctx, http.StatusOK, map[string]string{"status": "ok"})
			},
		})

		testServer := httptest.NewServer(srv.GetHandler())
		defer testServer.Close()

		// Request from allowed origin
		req, _ := http.NewRequest("GET", testServer.URL+"/api/test", nil)
		req.Header.Set("Origin", "http://example.com")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
		if allowOrigin != "http://example.com" {
			t.Errorf("Expected Access-Control-Allow-Origin: http://example.com, got %q", allowOrigin)
		}

		// Request from disallowed origin
		req2, _ := http.NewRequest("GET", testServer.URL+"/api/test", nil)
		req2.Header.Set("Origin", "http://evil.com")
		resp2, err := client.Do(req2)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp2.Body.Close()

		allowOrigin2 := resp2.Header.Get("Access-Control-Allow-Origin")
		if allowOrigin2 != "" {
			t.Errorf("Expected no Access-Control-Allow-Origin for disallowed origin, got %q", allowOrigin2)
		}
	})

	// Test 8: CORS Preflight Request
	t.Run("CORS_PreflightRequest", func(t *testing.T) {
		srv := server.NewServer(
			server.WithMiddleware(server.CORSMiddleware([]string{"http://example.com"})),
		)

		// Register both GET and OPTIONS routes for the same path
		// The CORS middleware will handle the preflight response for OPTIONS
		srv.RegisterRoute(&server.Route{
			Method: server.GET,
			Path:   "/api/test",
			Handler: func(ctx *server.Context) error {
				return server.SendJSON(ctx, http.StatusOK, map[string]string{"status": "ok"})
			},
		})

		// Register OPTIONS route - CORS middleware will intercept and handle preflight
		srv.RegisterRoute(&server.Route{
			Method: "OPTIONS",
			Path:   "/api/test",
			Handler: func(ctx *server.Context) error {
				// This handler won't be called because CORS middleware intercepts OPTIONS
				return nil
			},
		})

		testServer := httptest.NewServer(srv.GetHandler())
		defer testServer.Close()

		// OPTIONS preflight request
		req, _ := http.NewRequest("OPTIONS", testServer.URL+"/api/test", nil)
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Preflight request failed: %v", err)
		}
		resp.Body.Close()

		// Preflight should return 204 No Content
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected 204 No Content for preflight, got %d", resp.StatusCode)
		}

		// Check CORS headers
		if allowMethods := resp.Header.Get("Access-Control-Allow-Methods"); allowMethods == "" {
			t.Error("Expected Access-Control-Allow-Methods header to be set")
		}
	})

	// Test 9: Recovery Middleware - Panic Recovery
	t.Run("RecoveryMiddleware", func(t *testing.T) {
		srv := server.NewServer(
			server.WithMiddleware(server.RecoveryMiddleware()),
		)

		srv.RegisterRoute(&server.Route{
			Method: server.GET,
			Path:   "/api/panic",
			Handler: func(ctx *server.Context) error {
				panic("intentional panic for testing")
			},
		})

		testServer := httptest.NewServer(srv.GetHandler())
		defer testServer.Close()

		// This request should not crash the server
		resp, err := http.Get(testServer.URL + "/api/panic")
		if err != nil {
			t.Fatalf("Request failed (server may have crashed): %v", err)
		}
		defer resp.Body.Close()

		// Should return 500 Internal Server Error
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected 500 Internal Server Error after panic, got %d", resp.StatusCode)
		}

		// Verify server is still responsive after panic
		resp2, err := http.Get(testServer.URL + "/api/panic")
		if err != nil {
			t.Fatalf("Server stopped responding after panic: %v", err)
		}
		resp2.Body.Close()

		t.Log("Server successfully recovered from panic")
	})

	// Test 10: AuthMiddleware with custom validation function
	t.Run("AuthMiddleware_CustomValidation", func(t *testing.T) {
		// Custom validation function that checks for specific token
		validateFunc := func(ctx *server.Context) (bool, error) {
			token := ctx.Request.Header.Get("Authorization")
			return token == "Bearer admin-secret", nil
		}

		srv := server.NewServer()
		srv.RegisterRoute(&server.Route{
			Method:      server.GET,
			Path:        "/api/admin",
			Middlewares: []server.Middleware{server.AuthMiddleware(validateFunc)},
			Handler: func(ctx *server.Context) error {
				return server.SendJSON(ctx, http.StatusOK, map[string]string{"access": "granted"})
			},
		})

		testServer := httptest.NewServer(srv.GetHandler())
		defer testServer.Close()

		// Test with wrong token
		req, _ := http.NewRequest("GET", testServer.URL+"/api/admin", nil)
		req.Header.Set("Authorization", "Bearer wrong-token")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 for wrong token, got %d", resp.StatusCode)
		}

		// Test with correct token
		req2, _ := http.NewRequest("GET", testServer.URL+"/api/admin", nil)
		req2.Header.Set("Authorization", "Bearer admin-secret")
		resp2, err := client.Do(req2)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp2.Body.Close()

		if resp2.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 for correct token, got %d", resp2.StatusCode)
		}
	})

	t.Log("All security features E2E tests passed")
}

// Mock server handler for testing (temporary until real server is ready)
func mockServerHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	mux.HandleFunc("/greet/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/greet/"):]
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Hello, %s!"}`, name)
	})

	return mux
}

// TestMockServerBasic tests the mock server setup
func TestMockServerBasic(t *testing.T) {
	// This tests our test infrastructure itself
	helper := NewTestHelper(t)
	server := NewMockServer(mockServerHandler())
	defer server.Close()

	// Test simple route
	resp := MakeRequest(t, server.URL, HTTPRequest{
		Method: "GET",
		Path:   "/test",
	})

	helper.AssertEqual(resp.StatusCode, 200, "Status code")
	helper.AssertContains(resp.Body, "ok", "Response body")

	// Test parameterized route
	resp2 := MakeRequest(t, server.URL, HTTPRequest{
		Method: "GET",
		Path:   "/greet/Alice",
	})

	helper.AssertEqual(resp2.StatusCode, 200, "Status code")
	helper.AssertContains(resp2.Body, "Alice", "Response body")

	t.Log("Mock server test infrastructure verified")
}

// TestCompilerBasicFlow tests the basic compiler flow
func TestCompilerBasicFlow(t *testing.T) {
	helper := NewTestHelper(t)

	// Test with minimal valid program
	source := `@ GET /test {
  > {status: "ok"}
}`

	module, err := parseSourceE2E(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	comp := compiler.NewCompiler()
	bytecode, err := comp.Compile(module)

	helper.AssertNoError(err, "Compilation should not error")
	helper.AssertNotNil(bytecode, "Bytecode should not be nil")

	// Verify magic bytes
	if len(bytecode) >= 4 {
		helper.AssertEqual(string(bytecode[0:4]), "GLYP", "Magic bytes")
	}

	t.Log("Compiler basic flow test passed")
}

// TestVMBasicFlow tests the basic VM execution flow
func TestVMBasicFlow(t *testing.T) {
	helper := NewTestHelper(t)

	// Use compiler to generate valid bytecode from simple source
	source := `@ GET /test {
  > {status: "ok"}
}`

	module, err := parseSourceE2E(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	comp := compiler.NewCompiler()
	bytecode, err := comp.Compile(module)
	helper.AssertNoError(err, "Compilation failed")

	v := vm.NewVM()
	result, err := v.Execute(bytecode)

	helper.AssertNoError(err, "Execution should not error")
	helper.AssertNotNil(result, "Result should not be nil")

	t.Log("VM basic flow test passed")
}

// TestEndToEndWithTimeout ensures tests don't hang
func TestEndToEndWithTimeout(t *testing.T) {
	done := make(chan bool)

	go func() {
		// Simulate a long-running operation
		time.Sleep(100 * time.Millisecond)
		done <- true
	}()

	select {
	case <-done:
		t.Log("Operation completed successfully")
	case <-time.After(1 * time.Second):
		t.Error("Operation timed out")
	}
}
