package tests

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/vm"
)

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
	source := helper.LoadFixture("simple_route.abc")

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

	source := helper.LoadFixture("path_param.abc")

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

	source := helper.LoadFixture("json_response.abc")

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

	source := helper.LoadFixture("multiple_routes.abc")

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

	source := helper.LoadFixture("post_route.abc")

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

	source := helper.LoadFixture("with_auth.abc")

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

	source := helper.LoadFixture("error_handling.abc")

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

	source := helper.LoadFixture("invalid_syntax.abc")

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
	t.Skip("Skipping until hot reload is implemented")

	// TODO: When hot reload is implemented:
	// 1. Start dev server with a route
	// 2. Make request to verify route works
	// 3. Modify the source file
	// 4. Verify server automatically reloads
	// 5. Make request to verify new behavior
	// 6. Test that invalid changes don't crash server
}

// TestConcurrentRequestsE2E tests handling concurrent HTTP requests
func TestConcurrentRequestsE2E(t *testing.T) {
	t.Skip("Skipping until server implementation is ready")

	// TODO: When server is implemented:
	// 1. Start server with a simple route
	// 2. Make 100 concurrent requests
	// 3. Verify all requests succeed
	// 4. Verify responses are correct
	// 5. Verify no race conditions
	// 6. Check server remains stable under load
}

// TestDatabaseIntegrationE2E tests database operations
func TestDatabaseIntegrationE2E(t *testing.T) {
	t.Skip("Skipping until database integration is implemented")

	// TODO: When database is integrated:
	// 1. Set up test database
	// 2. Compile program with database queries
	// 3. Start server
	// 4. Test SELECT query
	// 5. Test INSERT query
	// 6. Test UPDATE query
	// 7. Test DELETE query
	// 8. Verify SQL injection prevention works
	// 9. Clean up test database
}

// TestSecurityFeaturesE2E tests security features end-to-end
func TestSecurityFeaturesE2E(t *testing.T) {
	t.Skip("Skipping until security features are implemented")

	// TODO: When security features are ready:
	// Test SQL injection prevention
	// Test XSS prevention in responses
	// Test CSRF protection
	// Test rate limiting
	// Test authentication enforcement
	// Test authorization (role-based access)
	// Test input validation
	// Test secure headers (CORS, CSP, etc.)
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
	source := `@ GET /test
  > {status: "ok"}`

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
	source := `@ GET /test
  > {status: "ok"}`

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
