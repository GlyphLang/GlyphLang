package tests

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"testing"

	"github.com/glyphlang/glyph/pkg/compiler"
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
	source := `@ GET /test {
  > {status: "ok"}
}`

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
@ GET /hello {
  > {message: "Hello, World!"}
}

@ GET /greet/:name {
  > {message: "Hello, " + name + "!"}
}
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

@ GET /api/users {
  + auth(jwt)
  + ratelimit(100/min)
  % db: Database
  $ users = db.users.all()
  > users
}

@ GET /api/users/:id {
  + auth(jwt)
  % db: Database
  $ user = db.users.get(id)
  > user
}

@ POST /api/users {
  + auth(jwt, role: admin)
  < input: CreateUserInput
  % db: Database
  $ user = db.users.create(input)
  > user
}
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
	// First compile a simple program to get valid bytecode
	source := `@ GET /test {
  > {status: "ok"}
}`

	module, err := parseSource(source)
	if err != nil {
		b.Fatalf("Parse failed: %v", err)
	}

	c := compiler.NewCompiler()
	bytecode, err := c.Compile(module)
	if err != nil {
		b.Fatalf("Compilation failed: %v", err)
	}

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
	source := `@ GET /test {
  > {status: "ok"}
}`

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
		_ = compiler.NewCompiler()
	}
}

// BenchmarkParallelCompilation benchmarks parallel compilation
func BenchmarkParallelCompilation(b *testing.B) {
	source := `@ GET /test {
  > {status: "ok"}
}`

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
	// First compile a simple program to get valid bytecode
	source := `@ GET /test {
  > {status: "ok"}
}`

	module, err := parseSource(source)
	if err != nil {
		b.Fatalf("Parse failed: %v", err)
	}

	c := compiler.NewCompiler()
	bytecode, err := c.Compile(module)
	if err != nil {
		b.Fatalf("Compilation failed: %v", err)
	}

	b.ResetTimer()
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

// BenchmarkMemoryAllocation tracks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	source := `@ GET /test {
  > {status: "ok"}
}`

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

// BenchmarkComparisonWithExamples benchmarks real example programs
func BenchmarkComparisonWithExamples(b *testing.B) {
	// This will be useful for tracking performance regressions
	examples := []struct {
		name   string
		source string
	}{
		{
			name: "HelloWorld",
			source: `@ GET /hello {
  > {message: "Hello, World!"}
}`,
		},
		{
			name: "PathParam",
			source: `@ GET /greet/:name {
  > {message: "Hello, " + name + "!"}
}`,
		},
		{
			name: "WithAuth",
			source: `@ GET /protected {
  + auth(jwt)
  > {data: "secret"}
}`,
		},
		{
			name: "TypeDefinition",
			source: `: User {
  id: int!
  name: str!
}

@ GET /user {
  > {id: 1, name: "test"}
}`,
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
		program += fmt.Sprintf(`@ GET /api/resource%d/:id {
  + auth(jwt)
  %% db: Database
  $ result = db.resource%d.get(id)
  > result
}

`, i, i)
	}

	return program
}
