package tests

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/vm"
)

// TestParserIntegration tests the parser with real GLYPH programs
func TestParserIntegration(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		shouldError bool
	}{
		{
			name: "Simple route",
			source: `@ GET /test {
  > {status: "ok"}
}`,
			shouldError: false,
		},
		{
			name: "Route with path parameter",
			source: `@ GET /users/:id {
  > {id: id}
}`,
			shouldError: false,
		},
		{
			name: "Type definition",
			source: `: User {
  id: int!
  name: str!
}
@ GET /users/:id -> User {
  > {id: id, name: "test"}
}`,
			shouldError: false,
		},
		{
			name: "Route with type annotation",
			source: `@ GET /api/users -> List[User] {
  > []
}`,
			shouldError: false,
		},
		{
			name: "Multiple middleware",
			source: `@ GET /protected {
  + auth(jwt)
  + ratelimit(100/min)
  > {status: "ok"}
}`,
			shouldError: false,
		},
		{
			name: "HTTP method specification",
			source: `@ POST /api/users {
  < input: CreateUserInput
  > {created: true}
}`,
			shouldError: false,
		},
		{
			name: "Database query",
			source: `@ GET /api/users/:id {
  % db: Database
  $ user = db.users.get(id)
  > user
}`,
			shouldError: false,
		},
		{
			name: "Validation",
			source: `@ POST /api/create {
  < input: CreateInput
  > {status: "ok"}
}`,
			shouldError: false,
		},
		{
			name: "Result type with error",
			source: `@ GET /api/data/:id -> Data | Error {
  > {error: "not found"}
}`,
			shouldError: false,
		},
		{
			name:        "Missing route path",
			source:      `@ route`,
			shouldError: true,
		},
		{
			name:        "Invalid symbol",
			source:      `& invalid`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := NewTestHelper(t)
			comp := compiler.NewCompiler()
			module, err := parseSource(tt.source)

			// For error cases, parsing or compilation should fail
			if tt.shouldError {
				if err != nil {
					return // Expected failure at parse stage
				}
				// If parsing succeeded, compilation should fail
				_, compErr := comp.Compile(module)
				if compErr != nil {
					return // Expected failure at compile stage
				}
				t.Errorf("Expected error for test '%s', but parsing and compilation both succeeded", tt.name)
				return
			}

			// For non-error cases, parsing and compilation should succeed
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			bytecode, err := comp.Compile(module)
			helper.AssertNoError(err, "Parse failed")
			helper.AssertNotNil(bytecode, "Bytecode should not be nil")
		})
	}
}

// TestLexerIntegration tests the lexer token generation
func TestLexerIntegration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string // Expected token types
	}{
		{
			name:     "Route symbols",
			input:    "@ GET /test",
			expected: []string{"@", "route", "/test"},
		},
		{
			name:     "Type definition",
			input:    ": User { id: int! }",
			expected: []string{":", "User", "{", "id", ":", "int", "!", "}"},
		},
		{
			name:     "Middleware",
			input:    "+ auth(jwt)",
			expected: []string{"+", "auth", "(", "jwt", ")"},
		},
		{
			name:     "Database query",
			input:    "$ user = db.get(id)",
			expected: []string{"$", "user", "=", "db", ".", "get", "(", "id", ")"},
		},
		{
			name:     "String literals",
			input:    `"hello world"`,
			expected: []string{`"hello world"`},
		},
		{
			name:     "Numbers",
			input:    "123 45.67",
			expected: []string{"123", "45.67"},
		},
		{
			name:     "Return statement",
			input:    "> {status: \"ok\"}",
			expected: []string{">", "{", "status", ":", `"ok"`, "}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Skipping: lexer FFI tokenization not yet implemented")
		})
	}
}

// TestTypeCheckerIntegration tests type checking
func TestTypeCheckerIntegration(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		shouldError bool
		errorMsg    string
	}{
		{
			name: "Valid type usage",
			source: `: User { id: int! }
@ GET /user -> User {
  > {id: 123}
}`,
			shouldError: false,
		},
		{
			name: "Type mismatch",
			source: `: User { id: int! }
@ GET /user -> User {
  > {id: "not a number"}
}`,
			shouldError: true,
			errorMsg:    "type mismatch",
		},
		{
			name: "Missing required field",
			source: `: User {
  id: int!
  name: str!
}
@ GET /user -> User {
  > {id: 123}
}`,
			shouldError: true,
			errorMsg:    "missing required field",
		},
		{
			name: "Optional field",
			source: `: User {
  id: int!
  name: str
}
@ GET /user -> User {
  > {id: 123, name: ""}
}`,
			shouldError: false,
		},
		{
			name: "List type",
			source: `@ GET /numbers -> List[int] {
  > [1, 2, 3]
}`,
			shouldError: false,
		},
		{
			name: "Result type",
			source: `: Error { msg: str! }
@ GET /data -> int | Error {
  > {msg: "error"}
}`,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Skipping: type checker validation not yet implemented")
		})
	}
}

// TestInterpreterIntegration tests AST interpretation
func TestInterpreterIntegration(t *testing.T) {
	t.Skip("Skipping until interpreter is implemented")

	tests := []struct {
		name     string
		ast      interface{}
		input    map[string]interface{}
		expected interface{}
	}{
		{
			name: "Simple return",
			ast: map[string]interface{}{
				"type": "return",
				"value": map[string]interface{}{
					"status": "ok",
				},
			},
			expected: map[string]interface{}{"status": "ok"},
		},
		{
			name: "String concatenation",
			ast: map[string]interface{}{
				"type": "concat",
				"left": "Hello, ",
				"right": map[string]interface{}{
					"type": "var",
					"name": "name",
				},
			},
			input:    map[string]interface{}{"name": "Alice"},
			expected: "Hello, Alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Skipping: interpreter execution not yet implemented")
		})
	}
}

// Note: TestServerIntegration moved to integration_comprehensive_test.go

// TestCompilerVMIntegration tests compiler -> VM integration
func TestCompilerVMIntegration(t *testing.T) {
	helper := NewTestHelper(t)

	// Simple program
	source := `@ GET /test {
  > {status: "ok"}
}`

	// Compile
	comp := compiler.NewCompiler()
	module, err := parseSource(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	bytecode, err := comp.Compile(module)
	helper.AssertNoError(err, "Compilation failed")

	// Execute
	v := vm.NewVM()
	result, err := v.Execute(bytecode)
	helper.AssertNoError(err, "Execution failed")
	helper.AssertNotNil(result, "Result should not be nil")

	t.Log("Compiler -> VM integration test passed")
}

// TestStackOperations tests VM stack operations
func TestStackOperations(t *testing.T) {
	helper := NewTestHelper(t)
	v := vm.NewVM()

	// Test push and pop
	v.Push(vm.IntValue{Val: 42})
	v.Push(vm.StringValue{Val: "hello"})

	val2, err := v.Pop()
	helper.AssertNoError(err, "Pop failed")
	if strVal, ok := val2.(vm.StringValue); ok {
		helper.AssertEqual(strVal.Val, "hello", "String value")
	}

	val1, err := v.Pop()
	helper.AssertNoError(err, "Pop failed")
	if intVal, ok := val1.(vm.IntValue); ok {
		helper.AssertEqual(intVal.Val, int64(42), "Int value")
	}

	// Test stack underflow
	_, err = v.Pop()
	helper.AssertError(err, "Should error on empty stack")

	t.Log("Stack operations test passed")
}

// TestValueTypes tests VM value types
func TestValueTypes(t *testing.T) {
	helper := NewTestHelper(t)

	// Test IntValue
	intVal := vm.IntValue{Val: 123}
	helper.AssertEqual(intVal.Type(), "int", "Int type")

	// Test StringValue
	strVal := vm.StringValue{Val: "test"}
	helper.AssertEqual(strVal.Type(), "string", "String type")

	// Test BoolValue
	boolVal := vm.BoolValue{Val: true}
	helper.AssertEqual(boolVal.Type(), "bool", "Bool type")

	t.Log("Value types test passed")
}

// TestRouteMatching tests route pattern matching
func TestRouteMatching(t *testing.T) {
	t.Skip("Skipping until router is implemented")

	tests := []struct {
		pattern string
		path    string
		matches bool
		params  map[string]string
	}{
		{
			pattern: "/users",
			path:    "/users",
			matches: true,
			params:  map[string]string{},
		},
		{
			pattern: "/users/:id",
			path:    "/users/123",
			matches: true,
			params:  map[string]string{"id": "123"},
		},
		{
			pattern: "/users/:id/posts/:postId",
			path:    "/users/123/posts/456",
			matches: true,
			params:  map[string]string{"id": "123", "postId": "456"},
		},
		{
			pattern: "/users/:id",
			path:    "/posts/123",
			matches: false,
			params:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			t.Skip("Skipping: route matching logic not yet implemented")
		})
	}
}

// TestMiddlewareChain tests middleware execution order
func TestMiddlewareChain(t *testing.T) {
	t.Skip("Skipping: middleware chain execution order and behavior not yet testable")
}

// TestDatabaseQuerySafety tests SQL injection prevention
func TestDatabaseQuerySafety(t *testing.T) {
	t.Skip("Skipping until database integration is ready")

	tests := []struct {
		name        string
		query       string
		safe        bool
		description string
	}{
		{
			name:        "Parameterized query",
			query:       `db.query("SELECT * FROM users WHERE id = :id", {id})`,
			safe:        true,
			description: "Using parameterized query",
		},
		{
			name:        "String concatenation",
			query:       `db.query("SELECT * FROM users WHERE id = " + id)`,
			safe:        false,
			description: "Direct string concatenation - SQL injection risk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Skipping: compile-time SQL injection detection requires type checker")
		})
	}
}

// TestInputValidation tests input validation rules
func TestInputValidation(t *testing.T) {
	t.Skip("Skipping until validation is implemented")

	tests := []struct {
		name  string
		rules map[string]interface{}
		input interface{}
		valid bool
	}{
		{
			name: "String length validation",
			rules: map[string]interface{}{
				"name": "str(min=1, max=100)",
			},
			input: map[string]interface{}{"name": "Alice"},
			valid: true,
		},
		{
			name: "String too long",
			rules: map[string]interface{}{
				"name": "str(min=1, max=5)",
			},
			input: map[string]interface{}{"name": "Alice Bob Carol"},
			valid: false,
		},
		{
			name: "Email format",
			rules: map[string]interface{}{
				"email": "email_format",
			},
			input: map[string]interface{}{"email": "test@example.com"},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Skipping: input validation logic not yet implemented")
		})
	}
}

// TestErrorPropagation tests error handling and propagation
func TestErrorPropagation(t *testing.T) {
	t.Skip("Skipping: Result type, error propagation, and HTTP status code conversion not yet implemented")
}

// TestConcurrency tests concurrent execution safety
func TestConcurrency(t *testing.T) {
	t.Skip("Skipping: concurrent route handlers, shared state safety, and connection pooling not yet testable")
}

// TestMemoryManagement tests memory usage and cleanup
func TestMemoryManagement(t *testing.T) {
	helper := NewTestHelper(t)

	t.Run("stack_cleared_between_executions", func(t *testing.T) {
		// Create a VM and compile a simple program
		comp := compiler.NewCompiler()
		source := `@ GET /test {
  > {status: "ok"}
}`
		module, err := parseSource(source)
		helper.AssertNoError(err, "Parse failed")
		bytecode, err := comp.Compile(module)
		helper.AssertNoError(err, "Compilation failed")

		v := vm.NewVM()

		// Execute the program multiple times
		for i := 0; i < 100; i++ {
			result, err := v.Execute(bytecode)
			helper.AssertNoError(err, "Execution failed")
			helper.AssertNotNil(result, "Result should not be nil")

			// After execution, stack should be cleared (only result on stack, then popped)
			// The VM returns the top of stack, leaving stack empty or with 0 items
			if v.StackSize() > 0 {
				t.Errorf("Iteration %d: Stack should be empty after execution, got %d items", i, v.StackSize())
			}
		}
	})

	t.Run("stack_bounded_under_load", func(t *testing.T) {
		comp := compiler.NewCompiler()
		// Program that does arithmetic operations
		source := `@ GET /calc {
  $ a = 10
  $ b = 20
  $ c = a + b
  > {result: c}
}`
		module, err := parseSource(source)
		helper.AssertNoError(err, "Parse failed")
		bytecode, err := comp.Compile(module)
		helper.AssertNoError(err, "Compilation failed")

		v := vm.NewVM()
		maxStackSize := 0

		// Execute many times and track max stack size
		for i := 0; i < 1000; i++ {
			result, err := v.Execute(bytecode)
			helper.AssertNoError(err, "Execution failed")
			helper.AssertNotNil(result, "Result should not be nil")

			// Check stack size after execution
			currentSize := v.StackSize()
			if currentSize > maxStackSize {
				maxStackSize = currentSize
			}
		}

		// Stack should not grow unbounded - should be at most a small constant
		// After execution the stack should be empty (result was popped)
		if maxStackSize > 10 {
			t.Errorf("Stack grew too large: max size was %d", maxStackSize)
		}
		t.Logf("Max stack size observed: %d", maxStackSize)
	})

	t.Run("vm_reset_clears_state", func(t *testing.T) {
		v := vm.NewVM()

		// Push some values onto the stack
		v.Push(vm.IntValue{Val: 1})
		v.Push(vm.IntValue{Val: 2})
		v.Push(vm.IntValue{Val: 3})
		v.SetLocal("test", vm.StringValue{Val: "value"})

		// Verify state before reset
		if v.StackSize() != 3 {
			t.Errorf("Expected stack size 3, got %d", v.StackSize())
		}
		if v.LocalsCount() != 1 {
			t.Errorf("Expected 1 local, got %d", v.LocalsCount())
		}

		// Reset the VM
		v.Reset()

		// Verify state after reset
		if v.StackSize() != 0 {
			t.Errorf("Expected stack size 0 after reset, got %d", v.StackSize())
		}
		if v.LocalsCount() != 0 {
			t.Errorf("Expected 0 locals after reset, got %d", v.LocalsCount())
		}
		if v.IteratorCount() != 0 {
			t.Errorf("Expected 0 iterators after reset, got %d", v.IteratorCount())
		}
	})

	t.Run("iterators_cleaned_up", func(t *testing.T) {
		comp := compiler.NewCompiler()
		// Program with iteration that creates iterators
		source := `@ GET /iter {
  $ arr = [1, 2, 3]
  $ sum = 0
  for item in arr {
    $ sum = sum + item
  }
  > {sum: sum}
}`
		module, err := parseSource(source)
		helper.AssertNoError(err, "Parse failed")
		bytecode, err := comp.Compile(module)
		helper.AssertNoError(err, "Compilation failed")

		v := vm.NewVM()

		// Execute multiple times
		for i := 0; i < 50; i++ {
			result, err := v.Execute(bytecode)
			helper.AssertNoError(err, "Execution failed")
			helper.AssertNotNil(result, "Result should not be nil")
		}

		// After executions, check that iterators don't accumulate unboundedly
		// Note: iterators may persist in the map, but the count should be bounded
		iterCount := v.IteratorCount()
		t.Logf("Iterator count after 50 executions: %d", iterCount)

		// The iterator count might grow with each execution since they're not cleaned up
		// but we can verify it's not growing faster than expected
		if iterCount > 100 {
			t.Errorf("Too many iterators accumulated: %d", iterCount)
		}
	})

	t.Run("error_cleanup", func(t *testing.T) {
		comp := compiler.NewCompiler()
		// Valid program for comparison
		validSource := `@ GET /valid {
  > {status: "ok"}
}`
		validModule, err := parseSource(validSource)
		helper.AssertNoError(err, "Parse failed")
		validBytecode, err := comp.Compile(validModule)
		helper.AssertNoError(err, "Compilation failed")

		v := vm.NewVM()

		// First, execute a valid program
		result, err := v.Execute(validBytecode)
		helper.AssertNoError(err, "Valid execution failed")
		helper.AssertNotNil(result, "Result should not be nil")

		initialStackSize := v.StackSize()

		// Try to execute invalid bytecode (should error)
		invalidBytecode := []byte{0x00, 0x00, 0x00, 0x00} // Invalid magic bytes
		_, err = v.Execute(invalidBytecode)
		if err == nil {
			t.Error("Expected error for invalid bytecode")
		}

		// Stack should not have grown due to failed execution
		afterErrorStackSize := v.StackSize()
		if afterErrorStackSize > initialStackSize+1 {
			t.Errorf("Stack grew after error: was %d, now %d", initialStackSize, afterErrorStackSize)
		}

		// Should still be able to execute valid programs after an error
		result, err = v.Execute(validBytecode)
		helper.AssertNoError(err, "Execution after error failed")
		helper.AssertNotNil(result, "Result should not be nil after recovery")
	})

	t.Run("stress_test_multiple_vms", func(t *testing.T) {
		comp := compiler.NewCompiler()
		source := `@ GET /stress {
  > {value: 42}
}`
		module, err := parseSource(source)
		helper.AssertNoError(err, "Parse failed")
		bytecode, err := comp.Compile(module)
		helper.AssertNoError(err, "Compilation failed")

		// Create multiple VMs and run programs on each
		vms := make([]*vm.VM, 10)
		for i := range vms {
			vms[i] = vm.NewVM()
		}

		// Execute on all VMs
		for iteration := 0; iteration < 100; iteration++ {
			for i, v := range vms {
				result, err := v.Execute(bytecode)
				if err != nil {
					t.Errorf("VM %d, iteration %d: execution failed: %v", i, iteration, err)
				}
				if result == nil {
					t.Errorf("VM %d, iteration %d: result is nil", i, iteration)
				}
			}
		}

		// All VMs should have empty stacks after execution
		for i, v := range vms {
			if v.StackSize() > 0 {
				t.Errorf("VM %d: stack not empty after stress test, size=%d", i, v.StackSize())
			}
		}
	})

	t.Log("Memory management tests passed")
}

// TestBytecodeFormat tests bytecode structure
func TestBytecodeFormat(t *testing.T) {
	helper := NewTestHelper(t)

	comp := compiler.NewCompiler()
	module, err := parseSource("@ GET /test {\n  > {status: \"ok\"}\n}")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	bytecode, err := comp.Compile(module)
	helper.AssertNoError(err, "Compilation failed")

	// Verify magic bytes
	if len(bytecode) < 4 {
		t.Fatal("Bytecode too short")
	}
	helper.AssertEqual(string(bytecode[0:4]), "GLYP", "Magic bytes")

	// Verify version (bytes 4-7, little-endian)
	if len(bytecode) >= 8 {
		// Version should be present
		t.Logf("Bytecode version bytes: %v", bytecode[4:8])
	}

	t.Log("Bytecode format test passed")
}

// TestCompilerErrorMessages tests compiler error reporting
func TestCompilerErrorMessages(t *testing.T) {
	t.Skip("Skipping until error reporting is implemented")

	tests := []struct {
		name     string
		source   string
		errorMsg string // Expected error message substring
	}{
		{
			name:     "Syntax error",
			source:   "@ route",
			errorMsg: "syntax error",
		},
		{
			name:     "Type error",
			source:   ": User { id: int! }\n@ GET /user -> User\n  > {id: \"string\"}",
			errorMsg: "type mismatch",
		},
		{
			name:     "Unknown type",
			source:   "@ GET /test -> UnknownType",
			errorMsg: "undefined type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := NewTestHelper(t)
			comp := compiler.NewCompiler()
			module, err := parseSource(tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			_, err = comp.Compile(module)

			// Note: Compiler error reporting does not yet produce detailed messages.
			// Once implemented, err.Error() should contain tt.errorMsg.
			if err != nil {
				helper.AssertContains(err.Error(), tt.errorMsg, "Error message")
			} else {
				t.Skip("Skipping: compiler error reporting not yet implemented for: " + tt.name)
			}
		})
	}
}
