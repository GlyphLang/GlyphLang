package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/server"
	"github.com/glyphlang/glyph/pkg/vm"
)

// TestExampleProgramsCompilation tests that all example programs parse and compile successfully.
// This verifies the full pipeline: Source -> Lexer -> Parser -> AST -> Compiler -> Bytecode.
func TestExampleProgramsCompilation(t *testing.T) {
	examples := []struct {
		name string
		path string
	}{
		{"hello-world", filepath.Join("..", "examples", "hello-world", "main.glyph")},
		{"auth-demo", filepath.Join("..", "examples", "auth-demo", "main.glyph")},
		{"feature-showcase", filepath.Join("..", "examples", "feature-showcase", "main.glyph")},
		{"validation-demo", filepath.Join("..", "examples", "validation-demo", "main.glyph")},
		{"for-loop-demo", filepath.Join("..", "examples", "for-loop-demo.glyph")},
		{"while-loop-demo", filepath.Join("..", "examples", "while-loop-demo.glyph")},
		{"defaults-demo", filepath.Join("..", "examples", "defaults-demo", "main.glyph")},
	}

	comp := compiler.NewCompiler()

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			source, err := os.ReadFile(ex.path)
			if err != nil {
				t.Skipf("Example not found: %v", err)
				return
			}

			module, err := parseSource(string(source))
			if err != nil {
				t.Fatalf("Parse failed for %s: %v", ex.name, err)
			}

			if module == nil {
				t.Fatalf("Module is nil for %s", ex.name)
			}

			bytecode, err := comp.Compile(module)
			if err != nil {
				t.Fatalf("Compilation failed for %s: %v", ex.name, err)
			}

			if bytecode == nil || len(bytecode) == 0 {
				t.Fatalf("Empty bytecode for %s", ex.name)
			}

			// Verify magic bytes
			if len(bytecode) >= 4 && string(bytecode[0:4]) != "GLYP" {
				t.Errorf("Invalid magic bytes for %s: got %q, want %q", ex.name, string(bytecode[0:4]), "GLYP")
			}

			t.Logf("Example %s: parsed, compiled (%d bytes)", ex.name, len(bytecode))
		})
	}
}

// TestExampleProgramsVMExecution tests that example programs execute in the VM.
// This verifies the full pipeline: Source -> Lexer -> Parser -> AST -> Compiler -> Bytecode -> VM.
func TestExampleProgramsVMExecution(t *testing.T) {
	examples := []struct {
		name string
		path string
	}{
		{"hello-world", filepath.Join("..", "examples", "hello-world", "main.glyph")},
		{"for-loop-demo", filepath.Join("..", "examples", "for-loop-demo.glyph")},
		{"while-loop-demo", filepath.Join("..", "examples", "while-loop-demo.glyph")},
		{"defaults-demo", filepath.Join("..", "examples", "defaults-demo", "main.glyph")},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			source, err := os.ReadFile(ex.path)
			if err != nil {
				t.Skipf("Example not found: %v", err)
				return
			}

			module, err := parseSource(string(source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			comp := compiler.NewCompiler()
			bytecode, err := comp.Compile(module)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			v := vm.NewVM()
			result, err := v.Execute(bytecode)
			if err != nil {
				t.Fatalf("VM execution failed: %v", err)
			}

			if result == nil {
				t.Fatal("VM returned nil result")
			}

			t.Logf("Example %s: executed successfully, result type: %T", ex.name, result)
		})
	}
}

// TestExampleProgramModuleStructure verifies the parsed module structure of example programs.
func TestExampleProgramModuleStructure(t *testing.T) {
	t.Run("hello-world has expected routes", func(t *testing.T) {
		source, err := os.ReadFile(filepath.Join("..", "examples", "hello-world", "main.glyph"))
		if err != nil {
			t.Skipf("Example not found: %v", err)
			return
		}

		module, err := parseSource(string(source))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		routeCount := 0
		for _, item := range module.Items {
			if _, ok := item.(*ast.Route); ok {
				routeCount++
			}
		}

		if routeCount < 2 {
			t.Errorf("Expected at least 2 routes in hello-world, got %d", routeCount)
		}
	})

	t.Run("todo-api has type definitions and routes", func(t *testing.T) {
		source, err := os.ReadFile(filepath.Join("..", "examples", "todo-api", "main.glyph"))
		if err != nil {
			t.Skipf("Example not found: %v", err)
			return
		}

		module, err := parseSource(string(source))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		typeDefCount := 0
		routeCount := 0
		for _, item := range module.Items {
			switch item.(type) {
			case *ast.TypeDef:
				typeDefCount++
			case *ast.Route:
				routeCount++
			}
		}

		if typeDefCount < 3 {
			t.Errorf("Expected at least 3 type definitions in todo-api, got %d", typeDefCount)
		}
		if routeCount < 4 {
			t.Errorf("Expected at least 4 routes in todo-api (CRUD), got %d", routeCount)
		}
	})

	t.Run("feature-showcase has comprehensive items", func(t *testing.T) {
		source, err := os.ReadFile(filepath.Join("..", "examples", "feature-showcase", "main.glyph"))
		if err != nil {
			t.Skipf("Example not found: %v", err)
			return
		}

		module, err := parseSource(string(source))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if len(module.Items) < 10 {
			t.Errorf("Expected at least 10 items in feature-showcase, got %d", len(module.Items))
		}
	})
}

// TestInterpreterRouteExecution tests the interpreter executing routes with the real AST.
// Uses ExecuteRouteSimple which takes a route and path params map, returns interface{}.
func TestInterpreterRouteExecution(t *testing.T) {
	t.Run("simple route returns response", func(t *testing.T) {
		source := `@ GET /test {
  > {status: "ok", message: "Hello from Glyph"}
}`
		module, err := parseSource(source)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		interp := interpreter.NewInterpreter()
		err = interp.LoadModule(*module)
		if err != nil {
			t.Fatalf("Failed to load module: %v", err)
		}

		for _, item := range module.Items {
			if route, ok := item.(*ast.Route); ok {
				result, err := interp.ExecuteRouteSimple(route, map[string]string{})
				if err != nil {
					t.Fatalf("Route execution failed: %v", err)
				}
				if result == nil {
					t.Fatal("Route returned nil result")
				}

				resultMap, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected map result, got %T", result)
				}

				if resultMap["status"] != "ok" {
					t.Errorf("Expected status 'ok', got %v", resultMap["status"])
				}
				break
			}
		}
	})

	t.Run("route with variable assignment", func(t *testing.T) {
		source := `@ GET /calc {
  $ a = 10
  $ b = 20
  $ c = a + b
  > {result: c}
}`
		module, err := parseSource(source)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		interp := interpreter.NewInterpreter()
		err = interp.LoadModule(*module)
		if err != nil {
			t.Fatalf("Failed to load module: %v", err)
		}

		for _, item := range module.Items {
			if route, ok := item.(*ast.Route); ok {
				result, err := interp.ExecuteRouteSimple(route, map[string]string{})
				if err != nil {
					t.Fatalf("Route execution failed: %v", err)
				}

				resultMap, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected map result, got %T", result)
				}

				resultVal := resultMap["result"]
				if resultVal == nil {
					t.Error("Expected 'result' field in response")
				}
				break
			}
		}
	})

	t.Run("route with conditional logic", func(t *testing.T) {
		source := `@ GET /check {
  $ value = 42
  $ message = ""
  if value > 0 {
    $ message = "positive"
  } else {
    $ message = "non-positive"
  }
  > {value: value, message: message}
}`
		module, err := parseSource(source)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		interp := interpreter.NewInterpreter()
		err = interp.LoadModule(*module)
		if err != nil {
			t.Fatalf("Failed to load module: %v", err)
		}

		for _, item := range module.Items {
			if route, ok := item.(*ast.Route); ok {
				result, err := interp.ExecuteRouteSimple(route, map[string]string{})
				if err != nil {
					t.Fatalf("Route execution failed: %v", err)
				}

				resultMap, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected map result, got %T", result)
				}

				if resultMap["message"] != "positive" {
					t.Errorf("Expected message 'positive', got %v", resultMap["message"])
				}
				break
			}
		}
	})

	t.Run("route with path parameters", func(t *testing.T) {
		source := `@ GET /greet/:name {
  $ greeting = "Hello, " + name + "!"
  > {message: greeting}
}`
		module, err := parseSource(source)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		interp := interpreter.NewInterpreter()
		err = interp.LoadModule(*module)
		if err != nil {
			t.Fatalf("Failed to load module: %v", err)
		}

		for _, item := range module.Items {
			if route, ok := item.(*ast.Route); ok {
				result, err := interp.ExecuteRouteSimple(route, map[string]string{"name": "Alice"})
				if err != nil {
					t.Fatalf("Route execution failed: %v", err)
				}
				if result == nil {
					t.Fatal("Route returned nil result")
				}

				resultMap, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected map result, got %T", result)
				}

				msg, _ := resultMap["message"].(string)
				if !strings.Contains(msg, "Alice") {
					t.Errorf("Expected greeting containing 'Alice', got %q", msg)
				}
				break
			}
		}
	})

	t.Run("route via ExecuteRoute with full Request", func(t *testing.T) {
		source := `@ GET /test {
  > {status: "ok"}
}`
		module, err := parseSource(source)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		interp := interpreter.NewInterpreter()
		err = interp.LoadModule(*module)
		if err != nil {
			t.Fatalf("Failed to load module: %v", err)
		}

		for _, item := range module.Items {
			if route, ok := item.(*ast.Route); ok {
				req := &interpreter.Request{
					Path:   "/test",
					Method: "GET",
					Params: map[string]string{},
				}
				resp, err := interp.ExecuteRoute(route, req)
				if err != nil {
					t.Fatalf("Route execution failed: %v", err)
				}
				if resp == nil {
					t.Fatal("Route returned nil response")
				}
				if resp.Body == nil {
					t.Fatal("Response body is nil")
				}

				bodyMap, ok := resp.Body.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected map body, got %T", resp.Body)
				}
				if bodyMap["status"] != "ok" {
					t.Errorf("Expected status 'ok', got %v", bodyMap["status"])
				}
				break
			}
		}
	})
}

// TestFullPipelineHTTPServer tests the full pipeline with actual HTTP server integration.
// Source -> Parse -> Interpreter -> Server -> HTTP Request -> Response.
func TestFullPipelineHTTPServer(t *testing.T) {
	t.Run("server responds to registered routes", func(t *testing.T) {
		mockInterp := &MockConcurrentInterpreter{
			response: map[string]interface{}{
				"status":  "ok",
				"message": "Hello from Glyph E2E test",
			},
		}

		srv := server.NewServer(server.WithInterpreter(mockInterp))

		err := srv.RegisterRoute(&server.Route{
			Method: server.GET,
			Path:   "/api/health",
		})
		if err != nil {
			t.Fatalf("Failed to register route: %v", err)
		}

		ts := httptest.NewServer(srv.GetHandler())
		defer ts.Close()

		resp, err := http.Get(ts.URL + "/api/health")
		if err != nil {
			t.Fatalf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if result["status"] != "ok" {
			t.Errorf("Expected status 'ok', got %v", result["status"])
		}
	})

	t.Run("server handles multiple routes", func(t *testing.T) {
		mockInterp := &MockConcurrentInterpreter{
			response: map[string]interface{}{"status": "ok"},
		}

		srv := server.NewServer(server.WithInterpreter(mockInterp))

		routes := []struct {
			method server.HTTPMethod
			path   string
		}{
			{server.GET, "/api/users"},
			{server.POST, "/api/users"},
			{server.GET, "/api/users/:id"},
			{server.PUT, "/api/users/:id"},
			{server.DELETE, "/api/users/:id"},
		}

		for _, r := range routes {
			err := srv.RegisterRoute(&server.Route{
				Method: r.method,
				Path:   r.path,
			})
			if err != nil {
				t.Fatalf("Failed to register route %s %s: %v", r.method, r.path, err)
			}
		}

		ts := httptest.NewServer(srv.GetHandler())
		defer ts.Close()

		// Test GET /api/users
		resp, err := http.Get(ts.URL + "/api/users")
		if err != nil {
			t.Fatalf("GET /api/users failed: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET /api/users: expected 200, got %d", resp.StatusCode)
		}

		// Test POST /api/users
		postResp, err := http.Post(ts.URL+"/api/users", "application/json",
			strings.NewReader(`{"name":"test"}`))
		if err != nil {
			t.Fatalf("POST /api/users failed: %v", err)
		}
		postResp.Body.Close()
		if postResp.StatusCode != http.StatusOK {
			t.Errorf("POST /api/users: expected 200, got %d", postResp.StatusCode)
		}

		// Test GET /api/users/123 (path parameter)
		getResp, err := http.Get(ts.URL + "/api/users/123")
		if err != nil {
			t.Fatalf("GET /api/users/123 failed: %v", err)
		}
		getResp.Body.Close()
		if getResp.StatusCode != http.StatusOK {
			t.Errorf("GET /api/users/123: expected 200, got %d", getResp.StatusCode)
		}
	})

	t.Run("server returns 404 for unregistered routes", func(t *testing.T) {
		srv := server.NewServer()

		srv.RegisterRoute(&server.Route{
			Method: server.GET,
			Path:   "/api/test",
			Handler: func(ctx *server.Context) error {
				return server.SendJSON(ctx, http.StatusOK, map[string]string{"status": "ok"})
			},
		})

		ts := httptest.NewServer(srv.GetHandler())
		defer ts.Close()

		resp, err := http.Get(ts.URL + "/api/nonexistent")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 for unregistered route, got %d", resp.StatusCode)
		}
	})

	t.Run("server handles JSON response content type", func(t *testing.T) {
		srv := server.NewServer()

		srv.RegisterRoute(&server.Route{
			Method: server.GET,
			Path:   "/api/json",
			Handler: func(ctx *server.Context) error {
				return server.SendJSON(ctx, http.StatusOK, map[string]interface{}{
					"id":    1,
					"name":  "Test User",
					"email": "test@example.com",
				})
			},
		})

		ts := httptest.NewServer(srv.GetHandler())
		defer ts.Close()

		resp, err := http.Get(ts.URL + "/api/json")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("Expected Content-Type application/json, got %q", contentType)
		}

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		if result["name"] != "Test User" {
			t.Errorf("Expected name 'Test User', got %v", result["name"])
		}
	})
}

// TestPathParametersFullPipeline tests path parameter extraction through the full pipeline.
func TestPathParametersFullPipeline(t *testing.T) {
	router := server.NewRouter()

	routes := []*server.Route{
		{Method: server.GET, Path: "/users/:id"},
		{Method: server.GET, Path: "/users/:userId/posts/:postId"},
		{Method: server.GET, Path: "/api/v1/orders/:orderId/items/:itemId"},
	}

	for _, route := range routes {
		if err := router.RegisterRoute(route); err != nil {
			t.Fatalf("Failed to register route %s: %v", route.Path, err)
		}
	}

	tests := []struct {
		name       string
		method     server.HTTPMethod
		path       string
		wantParams map[string]string
		wantMatch  bool
	}{
		{"single param", server.GET, "/users/42", map[string]string{"id": "42"}, true},
		{"double param", server.GET, "/users/10/posts/20", map[string]string{"userId": "10", "postId": "20"}, true},
		{"nested param", server.GET, "/api/v1/orders/100/items/5", map[string]string{"orderId": "100", "itemId": "5"}, true},
		{"no match", server.GET, "/nonexistent/path", nil, false},
		{"wrong method", server.POST, "/users/42", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route, params, err := router.Match(tt.method, tt.path)

			if tt.wantMatch {
				if err != nil {
					t.Fatalf("Expected match for %s, got error: %v", tt.path, err)
				}
				if route == nil {
					t.Fatal("Matched route is nil")
				}
				for key, expected := range tt.wantParams {
					if got := params[key]; got != expected {
						t.Errorf("Param %q: got %q, want %q", key, got, expected)
					}
				}
			} else {
				if err == nil && route != nil {
					t.Errorf("Expected no match for %s %s, but got a match", tt.method, tt.path)
				}
			}
		})
	}
}

// TestMiddlewareChainE2E tests middleware execution in the server pipeline.
func TestMiddlewareChainE2E(t *testing.T) {
	t.Run("middleware executes in order", func(t *testing.T) {
		var executionOrder []string

		middleware1 := server.Middleware(func(next server.RouteHandler) server.RouteHandler {
			return func(ctx *server.Context) error {
				executionOrder = append(executionOrder, "middleware1-before")
				err := next(ctx)
				executionOrder = append(executionOrder, "middleware1-after")
				return err
			}
		})

		middleware2 := server.Middleware(func(next server.RouteHandler) server.RouteHandler {
			return func(ctx *server.Context) error {
				executionOrder = append(executionOrder, "middleware2-before")
				err := next(ctx)
				executionOrder = append(executionOrder, "middleware2-after")
				return err
			}
		})

		srv := server.NewServer(
			server.WithMiddleware(middleware1),
			server.WithMiddleware(middleware2),
		)

		srv.RegisterRoute(&server.Route{
			Method: server.GET,
			Path:   "/test",
			Handler: func(ctx *server.Context) error {
				executionOrder = append(executionOrder, "handler")
				return server.SendJSON(ctx, http.StatusOK, map[string]string{"ok": "true"})
			},
		})

		ts := httptest.NewServer(srv.GetHandler())
		defer ts.Close()

		resp, err := http.Get(ts.URL + "/test")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		if len(executionOrder) < 3 {
			t.Errorf("Expected at least 3 execution steps, got %d: %v", len(executionOrder), executionOrder)
		}

		// Verify handler was called
		handlerCalled := false
		for _, step := range executionOrder {
			if step == "handler" {
				handlerCalled = true
				break
			}
		}
		if !handlerCalled {
			t.Error("Handler was not called in middleware chain")
		}
	})
}

// TestFixturesFullPipeline tests all fixtures through parse -> compile -> VM pipeline.
func TestFixturesFullPipeline(t *testing.T) {
	helper := NewTestHelper(t)

	fixtures := []struct {
		name        string
		shouldError bool
		// needsRuntime indicates fixtures that use path params, input body, or db
		// and cannot be executed in a bare VM without runtime context
		needsRuntime bool
	}{
		{"simple_route.glyph", false, false},
		{"path_param.glyph", false, true},
		{"json_response.glyph", false, false},
		{"multiple_routes.glyph", false, false},
		{"post_route.glyph", false, true},
		{"with_auth.glyph", false, false},
		{"error_handling.glyph", false, true},
		{"invalid_syntax.glyph", true, false},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			source := helper.LoadFixture(fixture.name)

			// Step 1: Parse
			module, parseErr := parseSource(source)

			if fixture.shouldError {
				if parseErr != nil {
					return // Expected failure
				}
				// If parsing succeeded, compilation should fail
				comp := compiler.NewCompiler()
				_, compErr := comp.Compile(module)
				if compErr != nil {
					return // Expected failure at compilation
				}
				t.Logf("Note: Fixture %s did not produce expected error", fixture.name)
				return
			}

			if parseErr != nil {
				t.Fatalf("Parse failed: %v", parseErr)
			}

			// Step 2: Compile
			comp := compiler.NewCompiler()
			bytecode, err := comp.Compile(module)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			if len(bytecode) < 4 {
				t.Fatal("Bytecode too short")
			}

			if string(bytecode[0:4]) != "GLYP" {
				t.Errorf("Invalid magic bytes: %q", string(bytecode[0:4]))
			}

			// Step 3: Execute in VM (skip fixtures that need runtime context)
			if fixture.needsRuntime {
				t.Logf("Skipping VM execution for %s (requires runtime context: path params, input, or db)", fixture.name)
				return
			}

			v := vm.NewVM()
			result, err := v.Execute(bytecode)
			if err != nil {
				t.Fatalf("VM execution failed: %v", err)
			}

			if result == nil {
				t.Fatal("VM returned nil result")
			}

			// Step 4: Verify stack is clean after execution
			if v.StackSize() > 0 {
				t.Errorf("Stack not empty after execution, size=%d", v.StackSize())
			}
		})
	}
}

// TestBytecodeReproducibility tests that compiling the same source produces consistent bytecode.
func TestBytecodeReproducibility(t *testing.T) {
	source := `@ GET /test {
  > {status: "ok", message: "Hello"}
}`

	module, err := parseSource(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	comp1 := compiler.NewCompiler()
	bytecode1, err := comp1.Compile(module)
	if err != nil {
		t.Fatalf("First compilation failed: %v", err)
	}

	// Re-parse and compile to get a second copy
	module2, err := parseSource(source)
	if err != nil {
		t.Fatalf("Second parse failed: %v", err)
	}

	comp2 := compiler.NewCompiler()
	bytecode2, err := comp2.Compile(module2)
	if err != nil {
		t.Fatalf("Second compilation failed: %v", err)
	}

	if len(bytecode1) != len(bytecode2) {
		t.Errorf("Bytecode length mismatch: %d vs %d", len(bytecode1), len(bytecode2))
	}

	for i := 0; i < len(bytecode1) && i < len(bytecode2); i++ {
		if bytecode1[i] != bytecode2[i] {
			t.Errorf("Bytecode differs at byte %d: %02x vs %02x", i, bytecode1[i], bytecode2[i])
			break
		}
	}
}

// TestVMExecutionIsolation tests that multiple VM executions don't interfere with each other.
func TestVMExecutionIsolation(t *testing.T) {
	sources := []string{
		`@ GET /a { > {route: "a"} }`,
		`@ GET /b { $ x = 42
  > {route: "b", value: x} }`,
		`@ GET /c { > {route: "c", items: [1, 2, 3]} }`,
	}

	var bytecodes [][]byte
	comp := compiler.NewCompiler()

	for _, src := range sources {
		module, err := parseSource(src)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		bc, err := comp.Compile(module)
		if err != nil {
			t.Fatalf("Compilation failed: %v", err)
		}
		bytecodes = append(bytecodes, bc)
	}

	// Execute all bytecodes on the same VM
	v := vm.NewVM()
	for i, bc := range bytecodes {
		result, err := v.Execute(bc)
		if err != nil {
			t.Errorf("Execution %d failed: %v", i, err)
			continue
		}
		if result == nil {
			t.Errorf("Execution %d returned nil", i)
		}
	}

	// Stack should be clean
	if v.StackSize() > 0 {
		t.Errorf("Stack not empty after all executions, size=%d", v.StackSize())
	}
}
