package compiler

import (
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/parser"
	"github.com/glyphlang/glyph/pkg/vm"
)

func TestArrayIndexCompilation(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantType string
		wantVal  interface{}
	}{
		{
			name: "simple array index",
			code: `@ GET /test
  $ arr = [1, 2, 3]
  $ val = arr[0]
  > val`,
			wantType: "int",
			wantVal:  int64(1),
		},
		{
			name: "array index with variable",
			code: `@ GET /test
  $ arr = [10, 20, 30]
  $ idx = 1
  $ val = arr[idx]
  > val`,
			wantType: "int",
			wantVal:  int64(20),
		},
		{
			name: "array index with expression",
			code: `@ GET /test
  $ arr = [100, 200, 300]
  $ val = arr[1 + 1]
  > val`,
			wantType: "int",
			wantVal:  int64(300),
		},
		{
			name: "string array index",
			code: `@ GET /test
  $ arr = ["a", "b", "c"]
  $ val = arr[2]
  > val`,
			wantType: "string",
			wantVal:  "c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Tokenize the code
			lexer := parser.NewLexer(tt.code)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("Tokenize failed: %v", err)
			}

			// Parse the code
			p := parser.NewParser(tokens)
			module, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Compile the code
			compiler := NewCompiler()
			bytecode, err := compiler.Compile(module)
			if err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			// Execute the bytecode
			vmInstance := vm.NewVM()
			result, err := vmInstance.Execute(bytecode)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			// Check the result type and value
			switch tt.wantType {
			case "int":
				intVal, ok := result.(vm.IntValue)
				if !ok {
					t.Fatalf("Expected IntValue, got %T", result)
				}
				if intVal.Val != tt.wantVal.(int64) {
					t.Errorf("Expected %v, got %v", tt.wantVal, intVal.Val)
				}
			case "string":
				strVal, ok := result.(vm.StringValue)
				if !ok {
					t.Fatalf("Expected StringValue, got %T", result)
				}
				if strVal.Val != tt.wantVal.(string) {
					t.Errorf("Expected %v, got %v", tt.wantVal, strVal.Val)
				}
			}
		})
	}
}

func TestArrayIndexBoundsError(t *testing.T) {
	code := `@ GET /test
  $ arr = [1, 2, 3]
  $ val = arr[10]
  > val`

	// Tokenize the code
	lexer := parser.NewLexer(code)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	// Parse the code
	p := parser.NewParser(tokens)
	module, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Compile the code
	compiler := NewCompiler()
	bytecode, err := compiler.Compile(module)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Execute the bytecode - should error
	vmInstance := vm.NewVM()
	_, err = vmInstance.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected bounds error, got nil")
	}

	// Check that error mentions bounds
	if !strings.Contains(err.Error(), "index out of bounds") || !strings.Contains(err.Error(), "10") {
		t.Errorf("Expected bounds error with index 10, got: %v", err)
	}
}

func TestArrayIndexNegativeError(t *testing.T) {
	code := `@ GET /test
  $ arr = [1, 2, 3]
  $ val = arr[0 - 1]
  > val`

	// Tokenize the code
	lexer := parser.NewLexer(code)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	// Parse the code
	p := parser.NewParser(tokens)
	module, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Compile the code
	compiler := NewCompiler()
	bytecode, err := compiler.Compile(module)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Execute the bytecode - should error
	vmInstance := vm.NewVM()
	_, err = vmInstance.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected bounds error, got nil")
	}

	// Check that error mentions bounds
	if !strings.Contains(err.Error(), "index out of bounds") || !strings.Contains(err.Error(), "-1") {
		t.Errorf("Expected bounds error with index -1, got: %v", err)
	}
}

func TestNestedArrayIndex(t *testing.T) {
	code := `@ GET /test
  $ arr = [[1, 2], [3, 4], [5, 6]]
  $ inner = arr[1]
  $ val = inner[0]
  > val`

	// Tokenize the code
	lexer := parser.NewLexer(code)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	// Parse the code
	p := parser.NewParser(tokens)
	module, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Compile the code
	compiler := NewCompiler()
	bytecode, err := compiler.Compile(module)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Execute the bytecode
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check the result
	intVal, ok := result.(vm.IntValue)
	if !ok {
		t.Fatalf("Expected IntValue, got %T", result)
	}
	if intVal.Val != 3 {
		t.Errorf("Expected 3, got %v", intVal.Val)
	}
}
