package tests

import (
	"fmt"
	"testing"

	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/parser"
	"github.com/glyphlang/glyph/pkg/vm"
)

// Helper function to parse source code into a Module
func parseSource(source string) (*interpreter.Module, error) {
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}

	p := parser.NewParser(tokens)
	return p.Parse()
}

// BenchmarkCompilation benchmarks the compilation process
func BenchmarkCompilation(b *testing.B) {
	source := `@ GET /test
  > {status: "ok"}`

	module, err := parseSource(source)
	if err != nil {
		b.Fatalf("Parse failed: %v", err)
	}

	c := compiler.NewCompiler()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.Compile(module)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

// BenchmarkCompilationSmallProgram benchmarks small program compilation
func BenchmarkCompilationSmallProgram(b *testing.B) {
	source := `
@ GET /hello
  > {message: "Hello, World!"}

@ GET /greet/:name
  > {message: "Hello, " + name + "!"}
`

	module, err := parseSource(source)
	if err != nil {
		b.Fatalf("Parse failed: %v", err)
	}

	c := compiler.NewCompiler()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.Compile(module)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

// BenchmarkCompilationMediumProgram benchmarks medium program compilation
func BenchmarkCompilationMediumProgram(b *testing.B) {
	source := `
: User {
  id: int!
  name: str!
  email: str!
  created_at: timestamp
}

: Error {
  code: str!
  message: str!
}

@ GET /api/users -> List[User]
  + auth(jwt)
  + ratelimit(100/min)
  % db: Database
  $ users = db.users.all()
  > users

@ GET /api/users/:id -> User | Error
  + auth(jwt)
  % db: Database
  $ user = db.users.get(id)
  > user

@ POST /api/users -> User | Error
  + auth(jwt, role: admin)
  < input: CreateUserInput
  ! validate input {
    name: str(min=1, max=100)
    email: email_format
  }
  % db: Database
  $ user = db.users.create(input)
  > user
`

	module, err := parseSource(source)
	if err != nil {
		b.Fatalf("Parse failed: %v", err)
	}

	c := compiler.NewCompiler()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.Compile(module)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

// BenchmarkCompilationLargeProgram benchmarks large program compilation
func BenchmarkCompilationLargeProgram(b *testing.B) {
	// Generate a large program with many routes
	source := generateLargeProgram(50) // 50 routes

	module, err := parseSource(source)
	if err != nil {
		b.Fatalf("Parse failed: %v", err)
	}

	c := compiler.NewCompiler()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.Compile(module)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

// BenchmarkVMExecution benchmarks VM bytecode execution
func BenchmarkVMExecution(b *testing.B) {
	// Create simple bytecode
	bytecode := []byte{0x41, 0x49, 0x42, 0x43, 0x01, 0x00, 0x00, 0x00}

	v := vm.NewVM()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := v.Execute(bytecode)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkStackOperations benchmarks VM stack push/pop
func BenchmarkStackOperations(b *testing.B) {
	v := vm.NewVM()
	value := vm.IntValue{Val: 42}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Push(value)
		_, err := v.Pop()
		if err != nil {
			b.Fatalf("Stack operation failed: %v", err)
		}
	}
}

// BenchmarkStackPush benchmarks just stack push operations
func BenchmarkStackPush(b *testing.B) {
	v := vm.NewVM()
	value := vm.IntValue{Val: 42}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Push(value)
	}
}

// BenchmarkStackPop benchmarks stack pop operations
func BenchmarkStackPop(b *testing.B) {
	v := vm.NewVM()

	// Pre-populate stack
	for i := 0; i < 1000; i++ {
		v.Push(vm.IntValue{Val: int64(i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i < 1000 {
			v.Pop()
		} else {
			v.Push(vm.IntValue{Val: 42})
		}
	}
}

// BenchmarkCompileAndExecute benchmarks full compile + execute cycle
func BenchmarkCompileAndExecute(b *testing.B) {
	source := `@ GET /test
  > {status: "ok"}`

	module, err := parseSource(source)
	if err != nil {
		b.Fatalf("Parse failed: %v", err)
	}

	c := compiler.NewCompiler()
	v := vm.NewVM()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytecode, err := c.Compile(module)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}

		_, err = v.Execute(bytecode)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkVMCreation benchmarks VM instantiation
func BenchmarkVMCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vm.NewVM()
	}
}

// BenchmarkCompilerCreation benchmarks compiler instantiation
func BenchmarkCompilerCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// No compiler instance needed with FFI
	}
}

// BenchmarkParallelCompilation benchmarks parallel compilation
func BenchmarkParallelCompilation(b *testing.B) {
	source := `@ GET /test
  > {status: "ok"}`

	module, err := parseSource(source)
	if err != nil {
		b.Fatalf("Parse failed: %v", err)
	}

	b.RunParallel(func(pb *testing.PB) {
		c := compiler.NewCompiler()
		for pb.Next() {
			_, err := c.Compile(module)
			if err != nil {
				b.Fatalf("Compilation failed: %v", err)
			}
		}
	})
}

// BenchmarkParallelExecution benchmarks parallel VM execution
func BenchmarkParallelExecution(b *testing.B) {
	bytecode := []byte{0x41, 0x49, 0x42, 0x43, 0x01, 0x00, 0x00, 0x00}

	b.RunParallel(func(pb *testing.PB) {
		v := vm.NewVM()
		for pb.Next() {
			_, err := v.Execute(bytecode)
			if err != nil {
				b.Fatalf("Execution failed: %v", err)
			}
		}
	})
}

// BenchmarkRouteMatching benchmarks route matching (when implemented)
func BenchmarkRouteMatching(b *testing.B) {
	b.Skip("Skipping until router is implemented")

	// TODO: When router is ready:
	// - Benchmark simple route matching
	// - Benchmark parameterized route matching
	// - Benchmark route matching with many routes
	// - Benchmark route matching with nested paths
}

// BenchmarkHTTPHandler benchmarks HTTP request handling (when implemented)
func BenchmarkHTTPHandler(b *testing.B) {
	b.Skip("Skipping until server is implemented")

	// TODO: When server is ready:
	// - Benchmark single request handling
	// - Benchmark concurrent request handling
	// - Benchmark with middleware chain
	// - Benchmark with database queries
	// - Benchmark JSON serialization
}

// BenchmarkJSONSerialization benchmarks JSON response serialization
func BenchmarkJSONSerialization(b *testing.B) {
	b.Skip("Skipping until JSON serialization is implemented")

	// TODO: When JSON serialization is ready:
	// - Benchmark simple object serialization
	// - Benchmark complex nested object
	// - Benchmark array serialization
	// - Benchmark large response bodies
}

// BenchmarkMiddleware benchmarks middleware execution
func BenchmarkMiddleware(b *testing.B) {
	b.Skip("Skipping until middleware is implemented")

	// TODO: When middleware is ready:
	// - Benchmark single middleware
	// - Benchmark middleware chain (3 middlewares)
	// - Benchmark middleware chain (10 middlewares)
	// - Benchmark auth middleware
	// - Benchmark rate limit middleware
}

// BenchmarkDatabaseQuery benchmarks database operations
func BenchmarkDatabaseQuery(b *testing.B) {
	b.Skip("Skipping until database integration is ready")

	// TODO: When database is integrated:
	// - Benchmark SELECT query
	// - Benchmark INSERT query
	// - Benchmark UPDATE query
	// - Benchmark DELETE query
	// - Benchmark query with joins
	// - Benchmark connection pool overhead
}

// BenchmarkTypeChecking benchmarks type checking performance
func BenchmarkTypeChecking(b *testing.B) {
	b.Skip("Skipping until type checker is implemented")

	source := `
: User { id: int!, name: str! }
@ GET /user -> User
  > {id: 123, name: "test"}
`

	// TODO: When type checker is ready:
	// Benchmark type checking overhead
	module, err := parseSource(source)
	if err != nil {
		b.Fatalf("Parse failed: %v", err)
	}

	c := compiler.NewCompiler()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Compile(module)
	}
}

// BenchmarkMemoryAllocation tracks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	source := `@ GET /test
  > {status: "ok"}`

	module, err := parseSource(source)
	if err != nil {
		b.Fatalf("Parse failed: %v", err)
	}

	c := compiler.NewCompiler()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := c.Compile(module)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

// BenchmarkLexing benchmarks just the lexing phase
func BenchmarkLexing(b *testing.B) {
	b.Skip("Skipping until lexer FFI is implemented")

	source := `@ GET /test
  > {status: "ok"}`

	// TODO: When lexer is accessible:
	// Benchmark just tokenization
	_ = source

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// lexer.Tokenize(source)
	}
}

// BenchmarkParsing benchmarks just the parsing phase
func BenchmarkParsing(b *testing.B) {
	b.Skip("Skipping until parser FFI is implemented")

	// TODO: When parser is accessible:
	// Pre-tokenize source
	// Benchmark just AST building

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// parser.Parse(tokens)
	}
}

// BenchmarkComparisonWithExamples benchmarks real example programs
func BenchmarkComparisonWithExamples(b *testing.B) {
	// This will be useful for tracking performance regressions
	examples := []struct {
		name   string
		source string
	}{
		{
			name: "HelloWorld",
			source: `@ GET /hello
  > {message: "Hello, World!"}`,
		},
		{
			name: "PathParam",
			source: `@ GET /greet/:name
  > {message: "Hello, " + name + "!"}`,
		},
		{
			name: "WithAuth",
			source: `@ GET /protected
  + auth(jwt)
  > {data: "secret"}`,
		},
		{
			name: "TypeDefinition",
			source: `: User { id: int!, name: str! }
@ GET /user -> User
  > {id: 1, name: "test"}`,
		},
	}

	for _, ex := range examples {
		b.Run(ex.name, func(b *testing.B) {
			module, err := parseSource(ex.source)
			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}

			c := compiler.NewCompiler()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := c.Compile(module)
				if err != nil {
					b.Fatalf("Compilation failed: %v", err)
				}
			}
		})
	}
}

// Benchmark different sizes to see scaling behavior
func BenchmarkCompilationScaling(b *testing.B) {
	sizes := []int{1, 5, 10, 25, 50, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Routes_%d", size), func(b *testing.B) {
			source := generateLargeProgram(size)
			module, err := parseSource(source)
			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}

			c := compiler.NewCompiler()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := c.Compile(module)
				if err != nil {
					b.Fatalf("Compilation failed: %v", err)
				}
			}
		})
	}
}

// Helper function to generate large programs for benchmarking
func generateLargeProgram(numRoutes int) string {
	program := ""

	// Add type definitions
	program += `: User {
  id: int!
  name: str!
  email: str!
}

: Error {
  code: str!
  message: str!
}

`

	// Add routes
	for i := 0; i < numRoutes; i++ {
		program += fmt.Sprintf(`@ GET /api/resource%d/:id -> User | Error
  + auth(jwt)
  %% db: Database
  $ result = db.resource%d.get(id)
  > result

`, i, i)
	}

	return program
}

// BenchmarkContextSwitching benchmarks goroutine context switching
func BenchmarkContextSwitching(b *testing.B) {
	b.Skip("Skipping until concurrent request handling is implemented")

	// TODO: When server supports concurrent requests:
	// Benchmark overhead of context switching
	// Compare single-threaded vs multi-threaded performance
}

// BenchmarkCachePerformance benchmarks context cache (when implemented)
func BenchmarkCachePerformance(b *testing.B) {
	b.Skip("Skipping until context cache is implemented")

	// TODO: When context cache is ready:
	// - Benchmark cache hit
	// - Benchmark cache miss
	// - Benchmark cache with different sizes
	// - Compare compilation with/without cache
}

// BenchmarkErrorHandling benchmarks error path performance
func BenchmarkErrorHandling(b *testing.B) {
	b.Skip("Skipping until error handling is implemented")

	// TODO: When error handling is ready:
	// - Benchmark success path vs error path
	// - Benchmark error propagation
	// - Benchmark error conversion
}

// BenchmarkValidation benchmarks input validation performance
func BenchmarkValidation(b *testing.B) {
	b.Skip("Skipping until validation is implemented")

	// TODO: When validation is ready:
	// - Benchmark simple validation rules
	// - Benchmark complex validation rules
	// - Benchmark validation with large inputs
}

// BenchmarkCompilationVsInterpretation compares compilation vs interpretation
func BenchmarkCompilationVsInterpretation(b *testing.B) {
	b.Skip("Skipping until both paths are implemented")

	// TODO: When both are available:
	// Compare bytecode compilation + execution vs direct interpretation
	// Measure break-even point
}
