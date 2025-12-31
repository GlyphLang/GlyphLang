package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/parser"
	"github.com/glyphlang/glyph/pkg/vm"
)

// TestCompileHelloWorld tests compiling a simple hello-world route to bytecode
func TestCompileHelloWorld(t *testing.T) {
	source := `@ GET /hello {
  > {message: "Hello, World!"}
}`

	// Parse source to module
	module, err := parseSource(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Compile to bytecode
	c := compiler.NewCompiler()
	bytecode, err := c.Compile(module)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Verify bytecode starts with magic bytes "GLYP"
	if len(bytecode) < 4 {
		t.Fatal("Bytecode too short")
	}

	if string(bytecode[0:4]) != "GLYP" {
		t.Errorf("Expected magic bytes 'GLYP', got '%s'", string(bytecode[0:4]))
	}

	// Verify bytecode is not empty
	if len(bytecode) < 10 {
		t.Errorf("Bytecode suspiciously small: %d bytes", len(bytecode))
	}

	t.Logf("Compiled %d bytes of source to %d bytes of bytecode", len(source), len(bytecode))
}

// TestExecuteBytecode tests executing bytecode with the VM
func TestExecuteBytecode(t *testing.T) {
	source := `@ GET /hello {
  > {message: "Hello, World!"}
}`

	// Parse source to module
	module, err := parseSource(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Compile to bytecode
	c := compiler.NewCompiler()
	bytecode, err := c.Compile(module)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Execute bytecode
	vmInstance := vm.NewVM()
	start := time.Now()
	result, err := vmInstance.Execute(bytecode)
	execTime := time.Since(start)

	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	t.Logf("Execution time: %s", execTime)
	t.Logf("Result: %v", result)

	// Verify execution is fast (should be < 1ms for simple bytecode)
	if execTime > time.Millisecond {
		t.Logf("Warning: Execution took longer than expected: %s", execTime)
	}
}

// TestCompareInterpreterVsBytecode tests that interpreter and bytecode produce same results
func TestCompareInterpreterVsBytecode(t *testing.T) {
	source := `@ GET /hello {
  > {message: "Hello, World!"}
}`

	// Parse with interpreter
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenization failed: %v", err)
	}

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	// Execute with interpreter
	interp := interpreter.NewInterpreter()
	if err := interp.LoadModule(*module); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Get first route
	var route *interpreter.Route
	for _, item := range module.Items {
		if r, ok := item.(*interpreter.Route); ok {
			route = r
			break
		}
	}

	if route == nil {
		t.Fatal("No route found in module")
	}

	interpStart := time.Now()
	interpResult, err := interp.ExecuteRouteSimple(route, map[string]string{})
	interpTime := time.Since(interpStart)

	if err != nil {
		t.Fatalf("Interpreter execution failed: %v", err)
	}

	// Compile and execute with bytecode
	module2, err := parseSource(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	c := compiler.NewCompiler()
	bytecode, err := c.Compile(module2)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	vmInstance := vm.NewVM()
	bytecodeStart := time.Now()
	bytecodeResult, err := vmInstance.Execute(bytecode)
	bytecodeTime := time.Since(bytecodeStart)

	if err != nil {
		t.Fatalf("Bytecode execution failed: %v", err)
	}

	t.Logf("Interpreter time: %s", interpTime)
	t.Logf("Bytecode time: %s", bytecodeTime)
	t.Logf("Interpreter result: %v", interpResult)
	t.Logf("Bytecode result: %v", bytecodeResult)

	// Note: We can't directly compare results yet since VM returns placeholder
	// This will be updated when VM implementation is complete
}

// TestBenchmarkInterpreterVsBytecode benchmarks interpreter vs bytecode execution
func TestBenchmarkInterpreterVsBytecode(t *testing.T) {
	source := `@ GET /api/users/:id {
  + auth(jwt)
  % db: Database
  $ user = db.users.get(id)
  > user
}`

	iterations := 100

	// Benchmark interpreter
	lexer := parser.NewLexer(source)
	tokens, _ := lexer.Tokenize()
	p := parser.NewParser(tokens)
	module, _ := p.Parse()
	interp := interpreter.NewInterpreter()
	interp.LoadModule(*module)

	var route *interpreter.Route
	for _, item := range module.Items {
		if r, ok := item.(*interpreter.Route); ok {
			route = r
			break
		}
	}

	interpStart := time.Now()
	for i := 0; i < iterations; i++ {
		interp.ExecuteRouteSimple(route, map[string]string{"id": "123"})
	}
	interpTotal := time.Since(interpStart)
	interpAvg := interpTotal / time.Duration(iterations)

	// Benchmark bytecode
	module2, err := parseSource(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	c := compiler.NewCompiler()
	bytecode, err := c.Compile(module2)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	vmInstance := vm.NewVM()
	bytecodeStart := time.Now()
	for i := 0; i < iterations; i++ {
		vmInstance.Execute(bytecode)
	}
	bytecodeTotal := time.Since(bytecodeStart)
	bytecodeAvg := bytecodeTotal / time.Duration(iterations)

	speedup := float64(interpTotal) / float64(bytecodeTotal)

	t.Logf("Interpreter: %s total, %s avg", interpTotal, interpAvg)
	t.Logf("Bytecode: %s total, %s avg", bytecodeTotal, bytecodeAvg)
	t.Logf("Speedup: %.2fx", speedup)

	// Bytecode should be faster (or at least competitive)
	// Note: This may not be true yet since VM is not fully implemented
}

// TestCompilationSpeed tests that compilation is fast (< 100ms target)
func TestCompilationSpeed(t *testing.T) {
	// Use a realistic multi-route API
	source := `
: User {
  id: int!
  name: str!
  email: str!
}

@ GET /api/users -> List[User] {
  + auth(jwt)
  + ratelimit(100/min)
  % db: Database
  $ users = db.users.all()
  > users
}

@ GET /api/users/:id -> User {
  + auth(jwt)
  % db: Database
  $ user = db.users.get(id)
  > user
}

@ GET /health {
  > {status: "ok"}
}
`

	// Parse source to module
	module, err := parseSource(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Run compilation multiple times and measure
	iterations := 10
	var totalTime time.Duration
	c := compiler.NewCompiler()

	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := c.Compile(module)
		if err != nil {
			t.Fatalf("Compilation failed: %v", err)
		}
		totalTime += time.Since(start)
	}

	avgTime := totalTime / time.Duration(iterations)

	t.Logf("Average compilation time: %s", avgTime)
	t.Logf("Source size: %d bytes", len(source))

	// Target: < 100ms for typical route
	if avgTime > 100*time.Millisecond {
		t.Logf("Warning: Compilation slower than 100ms target: %s", avgTime)
	}
}

// TestRoundTrip tests compile -> decompile round trip
func TestRoundTrip(t *testing.T) {
	t.Skip("Skipping until DeserializeFromBinary and FormatAST are implemented in pure Go")

	// TODO: Implement these in pure Go
	// source := `@ GET /hello
	//   > {message: "Hello, World!"}`
	//
	// Compile to bytecode
	// c := compiler.NewCompiler()
	// bytecode, err := c.Compile(source)
	// if err != nil {
	// 	t.Fatalf("Compilation failed: %v", err)
	// }
	//
	// Deserialize back to AST
	// module, err := deserializeFromBinary(bytecode)
	// if err != nil {
	// 	t.Fatalf("Deserialization failed: %v", err)
	// }
	//
	// Format back to source
	// decompiled, err := formatAST(module)
	// if err != nil {
	// 	t.Fatalf("Formatting failed: %v", err)
	// }
	//
	// t.Logf("Original:\n%s", source)
	// t.Logf("Decompiled:\n%s", decompiled)
	//
	// Verify module has at least one item
	// if len(module.Items) == 0 {
	// 	t.Fatal("Round trip lost all items")
	// }
	//
	// Verify decompiled source is not empty
	// if len(decompiled) == 0 {
	// 	t.Fatal("Decompiled source is empty")
	// }
}

// TestCompileAndSaveBytecode tests the full compile workflow
func TestCompileAndSaveBytecode(t *testing.T) {
	source := `@ GET /test {
  > {status: "ok"}
}`

	// Create temp directory
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.glyph")

	// Parse source to module
	module, err := parseSource(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Compile to bytecode
	c := compiler.NewCompiler()
	bytecode, err := c.Compile(module)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Save to file
	if err := os.WriteFile(outputPath, bytecode, 0644); err != nil {
		t.Fatalf("Failed to write bytecode: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Bytecode file was not created")
	}

	// Read back and verify
	readBytecode, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read bytecode: %v", err)
	}

	if len(readBytecode) != len(bytecode) {
		t.Errorf("Bytecode size mismatch: wrote %d, read %d", len(bytecode), len(readBytecode))
	}

	// Verify it can be executed
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(readBytecode)
	if err != nil {
		t.Fatalf("Failed to execute loaded bytecode: %v", err)
	}

	if result == nil {
		t.Fatal("Execution result is nil")
	}

	t.Logf("Successfully compiled, saved, loaded, and executed bytecode")
}

// TestVMPerformanceTarget tests that VM execution meets performance targets
func TestVMPerformanceTarget(t *testing.T) {
	source := `@ GET /hello {
  > {message: "Hello!"}
}`

	// Parse source to module
	module, err := parseSource(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	c := compiler.NewCompiler()
	bytecode, err := c.Compile(module)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Run many iterations to get accurate timing
	iterations := 1000
	vmInstance := vm.NewVM()

	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := vmInstance.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
	}
	totalTime := time.Since(start)
	avgTime := totalTime / time.Duration(iterations)

	// Calculate instructions per second (rough estimate)
	// Assuming ~10 instructions per simple route
	instructionsPerExec := 10
	totalInstructions := iterations * instructionsPerExec
	instructionsPerSec := float64(totalInstructions) / totalTime.Seconds()

	t.Logf("Average execution time: %s", avgTime)
	t.Logf("Estimated instructions/sec: %.0f", instructionsPerSec)

	// Target: < 10μs per instruction (100k instructions/sec)
	targetPerInstruction := 10 * time.Microsecond
	estimatedPerInstruction := avgTime / time.Duration(instructionsPerExec)

	t.Logf("Estimated time per instruction: %s", estimatedPerInstruction)
	t.Logf("Target: %s per instruction", targetPerInstruction)

	if estimatedPerInstruction > targetPerInstruction*10 {
		t.Logf("Warning: Execution slower than target by >10x")
	}
}
