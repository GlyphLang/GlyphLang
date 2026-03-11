package compiler

import (
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/vm"
)

// TestCompileArrayMapCall tests compilation of map(arr, fn) function call
func TestCompileArrayMapCall(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("arr", c.addConstant(vm.StringValue{Val: "arr"}))
	c.symbolTable.Define("fn", c.addConstant(vm.StringValue{Val: "fn"}))

	expr := &ast.FunctionCallExpr{
		Name: "map",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "arr"},
			&ast.VariableExpr{Name: "fn"},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(map) failed: %v", err)
	}

	// Verify opcodes: OpPush (fn name "map"), OpLoadVar (arr), OpLoadVar (fn), OpCall(2)
	verifyOpcodeSequence(t, c.code, []vm.Opcode{
		vm.OpPush,    // push function name "map"
		vm.OpLoadVar, // push arr
		vm.OpLoadVar, // push fn
		vm.OpCall,    // call with 2 args
	})
}

// TestCompileArrayFilterCall tests compilation of filter(arr, fn) function call
func TestCompileArrayFilterCall(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("arr", c.addConstant(vm.StringValue{Val: "arr"}))
	c.symbolTable.Define("predicate", c.addConstant(vm.StringValue{Val: "predicate"}))

	expr := &ast.FunctionCallExpr{
		Name: "filter",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "arr"},
			&ast.VariableExpr{Name: "predicate"},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(filter) failed: %v", err)
	}

	verifyOpcodeSequence(t, c.code, []vm.Opcode{
		vm.OpPush,
		vm.OpLoadVar,
		vm.OpLoadVar,
		vm.OpCall,
	})
}

// TestCompileArrayReduceCall tests compilation of reduce(arr, fn, initial) function call
func TestCompileArrayReduceCall(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("arr", c.addConstant(vm.StringValue{Val: "arr"}))
	c.symbolTable.Define("fn", c.addConstant(vm.StringValue{Val: "fn"}))

	expr := &ast.FunctionCallExpr{
		Name: "reduce",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "arr"},
			&ast.VariableExpr{Name: "fn"},
			&ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(reduce) failed: %v", err)
	}

	// Verify opcodes: OpPush (fn name), OpLoadVar (arr), OpLoadVar (fn), OpPush (0), OpCall(3)
	verifyOpcodeSequence(t, c.code, []vm.Opcode{
		vm.OpPush,    // push function name "reduce"
		vm.OpLoadVar, // push arr
		vm.OpLoadVar, // push fn
		vm.OpPush,    // push initial value 0
		vm.OpCall,    // call with 3 args
	})
}

// TestCompileArrayFindCall tests compilation of find(arr, fn) function call
func TestCompileArrayFindCall(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("items", c.addConstant(vm.StringValue{Val: "items"}))
	c.symbolTable.Define("matcher", c.addConstant(vm.StringValue{Val: "matcher"}))

	expr := &ast.FunctionCallExpr{
		Name: "find",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "items"},
			&ast.VariableExpr{Name: "matcher"},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(find) failed: %v", err)
	}

	verifyOpcodeSequence(t, c.code, []vm.Opcode{
		vm.OpPush,
		vm.OpLoadVar,
		vm.OpLoadVar,
		vm.OpCall,
	})
}

// TestCompileArraySortCall tests compilation of sort(arr) function call (no comparator)
func TestCompileArraySortCall(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("data", c.addConstant(vm.StringValue{Val: "data"}))

	expr := &ast.FunctionCallExpr{
		Name: "sort",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "data"},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(sort) failed: %v", err)
	}

	// sort with 1 arg: OpPush (fn name), OpLoadVar (data), OpCall(1)
	verifyOpcodeSequence(t, c.code, []vm.Opcode{
		vm.OpPush,
		vm.OpLoadVar,
		vm.OpCall,
	})
}

// TestCompileArrayReverseCall tests compilation of reverse(arr) function call
func TestCompileArrayReverseCall(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("list", c.addConstant(vm.StringValue{Val: "list"}))

	expr := &ast.FunctionCallExpr{
		Name: "reverse",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "list"},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(reverse) failed: %v", err)
	}

	verifyOpcodeSequence(t, c.code, []vm.Opcode{
		vm.OpPush,
		vm.OpLoadVar,
		vm.OpCall,
	})
}

// TestCompileArraySomeCall tests compilation of some(arr, fn) function call
func TestCompileArraySomeCall(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("numbers", c.addConstant(vm.StringValue{Val: "numbers"}))
	c.symbolTable.Define("check", c.addConstant(vm.StringValue{Val: "check"}))

	expr := &ast.FunctionCallExpr{
		Name: "some",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "numbers"},
			&ast.VariableExpr{Name: "check"},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(some) failed: %v", err)
	}

	verifyOpcodeSequence(t, c.code, []vm.Opcode{
		vm.OpPush,
		vm.OpLoadVar,
		vm.OpLoadVar,
		vm.OpCall,
	})
}

// TestCompileArrayEveryCall tests compilation of every(arr, fn) function call
func TestCompileArrayEveryCall(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("values", c.addConstant(vm.StringValue{Val: "values"}))
	c.symbolTable.Define("validator", c.addConstant(vm.StringValue{Val: "validator"}))

	expr := &ast.FunctionCallExpr{
		Name: "every",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "values"},
			&ast.VariableExpr{Name: "validator"},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(every) failed: %v", err)
	}

	verifyOpcodeSequence(t, c.code, []vm.Opcode{
		vm.OpPush,
		vm.OpLoadVar,
		vm.OpLoadVar,
		vm.OpCall,
	})
}

// TestCompileArrayFlatCall tests compilation of flat(arr) function call
func TestCompileArrayFlatCall(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("nested", c.addConstant(vm.StringValue{Val: "nested"}))

	expr := &ast.FunctionCallExpr{
		Name: "flat",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "nested"},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(flat) failed: %v", err)
	}

	verifyOpcodeSequence(t, c.code, []vm.Opcode{
		vm.OpPush,
		vm.OpLoadVar,
		vm.OpCall,
	})
}

// TestCompileArraySliceCall tests compilation of slice(arr, start, end) function call
func TestCompileArraySliceCall(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("arr", c.addConstant(vm.StringValue{Val: "arr"}))

	expr := &ast.FunctionCallExpr{
		Name: "slice",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "arr"},
			&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(slice) failed: %v", err)
	}

	verifyOpcodeSequence(t, c.code, []vm.Opcode{
		vm.OpPush,    // push function name "slice"
		vm.OpLoadVar, // push arr
		vm.OpPush,    // push start (1)
		vm.OpPush,    // push end (3)
		vm.OpCall,    // call with 3 args
	})
}

// TestCompileArrayOpsCallArgCount verifies the OpCall operand contains the correct argument count
func TestCompileArrayOpsCallArgCount(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		args     []ast.Expr
		argCount uint32
	}{
		{
			name:     "map has 2 args",
			funcName: "map",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
			argCount: 2,
		},
		{
			name:     "filter has 2 args",
			funcName: "filter",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
			argCount: 2,
		},
		{
			name:     "reduce has 3 args",
			funcName: "reduce",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			argCount: 3,
		},
		{
			name:     "find has 2 args",
			funcName: "find",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
			argCount: 2,
		},
		{
			name:     "sort has 1 arg",
			funcName: "sort",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
			argCount: 1,
		},
		{
			name:     "reverse has 1 arg",
			funcName: "reverse",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
			argCount: 1,
		},
		{
			name:     "some has 2 args",
			funcName: "some",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
			argCount: 2,
		},
		{
			name:     "every has 2 args",
			funcName: "every",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
			argCount: 2,
		},
		{
			name:     "flat has 1 arg",
			funcName: "flat",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
			argCount: 1,
		},
		{
			name:     "slice has 3 args",
			funcName: "slice",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
			argCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()

			expr := &ast.FunctionCallExpr{
				Name: tt.funcName,
				Args: tt.args,
			}

			err := c.compileFunctionCall(expr)
			if err != nil {
				t.Fatalf("compileFunctionCall(%s) failed: %v", tt.funcName, err)
			}

			// Find the OpCall instruction and verify its operand
			argCount := extractCallArgCount(t, c.code)
			if argCount != tt.argCount {
				t.Errorf("Expected OpCall arg count %d, got %d", tt.argCount, argCount)
			}
		})
	}
}

// TestCompileArrayOpsInRoute tests compilation of array operations within a route context
func TestCompileArrayOpsInRoute(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		args     []ast.Expr
	}{
		{
			name:     "map in route",
			funcName: "map",
			args: []ast.Expr{
				&ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
					},
				},
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "double"}},
			},
		},
		{
			name:     "filter in route",
			funcName: "filter",
			args: []ast.Expr{
				&ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
					},
				},
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "isEven"}},
			},
		},
		{
			name:     "reduce in route",
			funcName: "reduce",
			args: []ast.Expr{
				&ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
					},
				},
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "sum"}},
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
		},
		{
			name:     "find in route",
			funcName: "find",
			args: []ast.Expr{
				&ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}},
					},
				},
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "isLarge"}},
			},
		},
		{
			name:     "sort in route",
			funcName: "sort",
			args: []ast.Expr{
				&ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
					},
				},
			},
		},
		{
			name:     "reverse in route",
			funcName: "reverse",
			args: []ast.Expr{
				&ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
					},
				},
			},
		},
		{
			name:     "some in route",
			funcName: "some",
			args: []ast.Expr{
				&ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
					},
				},
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "isPositive"}},
			},
		},
		{
			name:     "every in route",
			funcName: "every",
			args: []ast.Expr{
				&ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
					},
				},
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "isEven"}},
			},
		},
		{
			name:     "flat in route",
			funcName: "flat",
			args: []ast.Expr{
				&ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.ArrayExpr{
							Elements: []ast.Expr{
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
							},
						},
						&ast.ArrayExpr{
							Elements: []ast.Expr{
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
							},
						},
					},
				},
			},
		},
		{
			name:     "slice in route",
			funcName: "slice",
			args: []ast.Expr{
				&ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 30}},
					},
				},
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
				&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &ast.Route{
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "result",
						Value: &ast.FunctionCallExpr{
							Name: tt.funcName,
							Args: tt.args,
						},
					},
					&ast.ReturnStatement{
						Value: &ast.VariableExpr{Name: "result"},
					},
				},
			}

			c := NewCompiler()
			bytecode, err := c.CompileRoute(route)
			if err != nil {
				t.Fatalf("CompileRoute() failed for %s: %v", tt.funcName, err)
			}

			// Verify bytecode is well-formed (has magic header)
			if len(bytecode) < 4 {
				t.Fatal("Bytecode too short")
			}
			if string(bytecode[:4]) != "GLYP" {
				t.Error("Missing GLYP magic header")
			}

			// Execute to verify the VM can decode the bytecode
			// (will fail at runtime because these aren't VM builtins,
			// but the bytecode structure should be valid)
			vmInstance := vm.NewVM()
			_, err = vmInstance.Execute(bytecode)
			if err == nil {
				t.Fatal("Expected undefined function error, got nil")
			}
			if !strings.Contains(err.Error(), "undefined function") {
				t.Errorf("Expected 'undefined function' error, got: %v", err)
			}
		})
	}
}

// TestCompileArrayOpsWithArrayLiteral tests compilation with inline array literal arguments
func TestCompileArrayOpsWithArrayLiteral(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("fn", c.addConstant(vm.StringValue{Val: "fn"}))

	// map([1, 2, 3], fn)
	expr := &ast.FunctionCallExpr{
		Name: "map",
		Args: []ast.Expr{
			&ast.ArrayExpr{
				Elements: []ast.Expr{
					&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
					&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
					&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
				},
			},
			&ast.VariableExpr{Name: "fn"},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(map with array literal) failed: %v", err)
	}

	// Opcodes: OpPush("map"), OpPush(1), OpPush(2), OpPush(3), OpBuildArray(3), OpLoadVar(fn), OpCall(2)
	verifyOpcodeSequence(t, c.code, []vm.Opcode{
		vm.OpPush,       // push function name "map"
		vm.OpPush,       // push 1
		vm.OpPush,       // push 2
		vm.OpPush,       // push 3
		vm.OpBuildArray, // build array [1, 2, 3]
		vm.OpLoadVar,    // push fn
		vm.OpCall,       // call with 2 args
	})
}

// TestCompileArrayOpsWithExpressionArgs tests compilation with expression arguments
func TestCompileArrayOpsWithExpressionArgs(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("arr", c.addConstant(vm.StringValue{Val: "arr"}))

	// slice(arr, 1 + 1, 2 * 3)
	expr := &ast.FunctionCallExpr{
		Name: "slice",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "arr"},
			&ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
			&ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
			},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(slice with expressions) failed: %v", err)
	}

	// Verify OpCall appears with 3 arguments
	argCount := extractCallArgCount(t, c.code)
	if argCount != 3 {
		t.Errorf("Expected OpCall arg count 3, got %d", argCount)
	}
}

// TestCompileArrayMapWithNestedCall tests map with a nested function call as argument
func TestCompileArrayMapWithNestedCall(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("arr", c.addConstant(vm.StringValue{Val: "arr"}))
	c.symbolTable.Define("fn", c.addConstant(vm.StringValue{Val: "fn"}))

	// map(arr, fn) used as argument to length()
	// length(map(arr, fn))
	expr := &ast.FunctionCallExpr{
		Name: "length",
		Args: []ast.Expr{
			&ast.FunctionCallExpr{
				Name: "map",
				Args: []ast.Expr{
					&ast.VariableExpr{Name: "arr"},
					&ast.VariableExpr{Name: "fn"},
				},
			},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(nested map in length) failed: %v", err)
	}

	// Should have two OpCall instructions - one for map, one for length
	callCount := countOpcode(c.code, vm.OpCall)
	if callCount != 2 {
		t.Errorf("Expected 2 OpCall instructions, got %d", callCount)
	}
}

// TestCompileArrayOpsChained tests compilation of chained array operations
func TestCompileArrayOpsChained(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("arr", c.addConstant(vm.StringValue{Val: "arr"}))
	c.symbolTable.Define("filterFn", c.addConstant(vm.StringValue{Val: "filterFn"}))
	c.symbolTable.Define("mapFn", c.addConstant(vm.StringValue{Val: "mapFn"}))

	// map(filter(arr, filterFn), mapFn) - chain filter then map
	expr := &ast.FunctionCallExpr{
		Name: "map",
		Args: []ast.Expr{
			&ast.FunctionCallExpr{
				Name: "filter",
				Args: []ast.Expr{
					&ast.VariableExpr{Name: "arr"},
					&ast.VariableExpr{Name: "filterFn"},
				},
			},
			&ast.VariableExpr{Name: "mapFn"},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(chained filter+map) failed: %v", err)
	}

	// Should have two OpCall instructions (filter then map)
	callCount := countOpcode(c.code, vm.OpCall)
	if callCount != 2 {
		t.Errorf("Expected 2 OpCall instructions for chained operations, got %d", callCount)
	}
}

// TestCompileArrayOpsReduceWithChain tests reduce(filter(arr, fn), accFn, init)
func TestCompileArrayOpsReduceWithChain(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("numbers", c.addConstant(vm.StringValue{Val: "numbers"}))
	c.symbolTable.Define("isPositive", c.addConstant(vm.StringValue{Val: "isPositive"}))
	c.symbolTable.Define("sumFn", c.addConstant(vm.StringValue{Val: "sumFn"}))

	// reduce(filter(numbers, isPositive), sumFn, 0)
	expr := &ast.FunctionCallExpr{
		Name: "reduce",
		Args: []ast.Expr{
			&ast.FunctionCallExpr{
				Name: "filter",
				Args: []ast.Expr{
					&ast.VariableExpr{Name: "numbers"},
					&ast.VariableExpr{Name: "isPositive"},
				},
			},
			&ast.VariableExpr{Name: "sumFn"},
			&ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(reduce with chained filter) failed: %v", err)
	}

	callCount := countOpcode(c.code, vm.OpCall)
	if callCount != 2 {
		t.Errorf("Expected 2 OpCall instructions (filter + reduce), got %d", callCount)
	}
}

// TestCompileLambdaInArrayOp tests that lambda expressions in array operations
// produce the expected unsupported expression error
func TestCompileLambdaInArrayOp(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("arr", c.addConstant(vm.StringValue{Val: "arr"}))

	// map(arr, (n) => n * 2) - lambda not supported in compiler
	expr := &ast.FunctionCallExpr{
		Name: "map",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "arr"},
			&ast.LambdaExpr{
				Params: []ast.Field{{Name: "n", Required: true}},
				Body: &ast.BinaryOpExpr{
					Left:  &ast.VariableExpr{Name: "n"},
					Op:    ast.Mul,
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
				},
			},
		},
	}

	err := c.compileFunctionCall(expr)
	if err == nil {
		t.Fatal("Expected error for lambda expression in array op, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported expression type") {
		t.Errorf("Expected 'unsupported expression type' error, got: %v", err)
	}
}

// TestCompileLambdaInFilterOp tests lambda in filter produces expected error
func TestCompileLambdaInFilterOp(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("items", c.addConstant(vm.StringValue{Val: "items"}))

	// filter(items, (x) => x > 5)
	expr := &ast.FunctionCallExpr{
		Name: "filter",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "items"},
			&ast.LambdaExpr{
				Params: []ast.Field{{Name: "x", Required: true}},
				Body: &ast.BinaryOpExpr{
					Left:  &ast.VariableExpr{Name: "x"},
					Op:    ast.Gt,
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
				},
			},
		},
	}

	err := c.compileFunctionCall(expr)
	if err == nil {
		t.Fatal("Expected error for lambda in filter, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported expression type") {
		t.Errorf("Expected 'unsupported expression type' error, got: %v", err)
	}
}

// TestCompileLambdaInReduceOp tests lambda in reduce produces expected error
func TestCompileLambdaInReduceOp(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("arr", c.addConstant(vm.StringValue{Val: "arr"}))

	// reduce(arr, (acc, n) => acc + n, 0)
	expr := &ast.FunctionCallExpr{
		Name: "reduce",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "arr"},
			&ast.LambdaExpr{
				Params: []ast.Field{
					{Name: "acc", Required: true},
					{Name: "n", Required: true},
				},
				Body: &ast.BinaryOpExpr{
					Left:  &ast.VariableExpr{Name: "acc"},
					Op:    ast.Add,
					Right: &ast.VariableExpr{Name: "n"},
				},
			},
			&ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
		},
	}

	err := c.compileFunctionCall(expr)
	if err == nil {
		t.Fatal("Expected error for lambda in reduce, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported expression type") {
		t.Errorf("Expected 'unsupported expression type' error, got: %v", err)
	}
}

// TestCompileArrayOpsConstantDeduplication verifies that the compiler deduplicates
// constants when the same function name is used multiple times
func TestCompileArrayOpsConstantDeduplication(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("arr1", c.addConstant(vm.StringValue{Val: "arr1"}))
	c.symbolTable.Define("arr2", c.addConstant(vm.StringValue{Val: "arr2"}))
	c.symbolTable.Define("fn", c.addConstant(vm.StringValue{Val: "fn"}))

	// First map call
	expr1 := &ast.FunctionCallExpr{
		Name: "map",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "arr1"},
			&ast.VariableExpr{Name: "fn"},
		},
	}
	err := c.compileFunctionCall(expr1)
	if err != nil {
		t.Fatalf("First map call failed: %v", err)
	}

	constantsBefore := len(c.constants)

	// Second map call (function name "map" should be deduplicated)
	expr2 := &ast.FunctionCallExpr{
		Name: "map",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "arr2"},
			&ast.VariableExpr{Name: "fn"},
		},
	}
	err = c.compileFunctionCall(expr2)
	if err != nil {
		t.Fatalf("Second map call failed: %v", err)
	}

	constantsAfter := len(c.constants)

	// The "map" string constant should not be duplicated
	// Only new constant should be for different args (none in this case since we reuse variables)
	if constantsAfter != constantsBefore {
		t.Errorf("Expected no new constants (deduplication), but went from %d to %d",
			constantsBefore, constantsAfter)
	}
}

// TestCompileArrayOpsFunctionNameInConstants verifies that the function name
// is stored in the constants pool
func TestCompileArrayOpsFunctionNameInConstants(t *testing.T) {
	funcNames := []string{"map", "filter", "reduce", "find", "sort", "reverse", "some", "every", "flat", "slice"}

	for _, name := range funcNames {
		t.Run(name, func(t *testing.T) {
			c := NewCompiler()

			expr := &ast.FunctionCallExpr{
				Name: name,
				Args: []ast.Expr{
					&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				},
			}

			err := c.compileFunctionCall(expr)
			if err != nil {
				t.Fatalf("compileFunctionCall(%s) failed: %v", name, err)
			}

			// Find the function name in the constants pool
			found := false
			for _, constant := range c.constants {
				if sv, ok := constant.(vm.StringValue); ok && sv.Val == name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Function name %q not found in constants pool", name)
			}
		})
	}
}

// TestCompileArrayOpsInExpressionStatement tests array ops used as expression statements
// (called for side effects, result discarded)
func TestCompileArrayOpsInExpressionStatement(t *testing.T) {
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ExpressionStatement{
				Expr: &ast.FunctionCallExpr{
					Name: "sort",
					Args: []ast.Expr{
						&ast.ArrayExpr{
							Elements: []ast.Expr{
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
							},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() failed: %v", err)
	}

	// Verify bytecode has OpPop after the function call (expression statement discards result)
	if len(bytecode) < 4 {
		t.Fatal("Bytecode too short")
	}
	if string(bytecode[:4]) != "GLYP" {
		t.Error("Missing GLYP magic header")
	}

	// The bytecode should be valid enough for the VM to start executing
	vmInstance := vm.NewVM()
	_, err = vmInstance.Execute(bytecode)
	// Expected to fail because sort is not a VM builtin
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "undefined function") {
		t.Errorf("Expected 'undefined function' error, got: %v", err)
	}
}

// TestCompileArrayOpsInConditional tests array operations inside if/else blocks
func TestCompileArrayOpsInConditional(t *testing.T) {
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "flag",
				Value:  &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
			},
			// Declare result in outer scope so it's visible after the if/else
			&ast.AssignStatement{
				Target: "result",
				Value:  &ast.LiteralExpr{Value: ast.NullLiteral{}},
			},
			&ast.IfStatement{
				Condition: &ast.VariableExpr{Name: "flag"},
				ThenBlock: []ast.Statement{
					&ast.ReassignStatement{
						Target: "result",
						Value: &ast.FunctionCallExpr{
							Name: "reverse",
							Args: []ast.Expr{
								&ast.ArrayExpr{
									Elements: []ast.Expr{
										&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
										&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
									},
								},
							},
						},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.ReassignStatement{
						Target: "result",
						Value: &ast.FunctionCallExpr{
							Name: "sort",
							Args: []ast.Expr{
								&ast.ArrayExpr{
									Elements: []ast.Expr{
										&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
										&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
									},
								},
							},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "result"},
			},
		},
	}

	c := NewCompilerWithOptLevel(OptNone)
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() failed: %v", err)
	}

	if len(bytecode) < 4 || string(bytecode[:4]) != "GLYP" {
		t.Error("Invalid bytecode header")
	}
}

// TestCompileArrayOpsInForLoop tests array operations inside a for loop
func TestCompileArrayOpsInForLoop(t *testing.T) {
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "items",
				Value: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.ArrayExpr{
							Elements: []ast.Expr{
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
							},
						},
						&ast.ArrayExpr{
							Elements: []ast.Expr{
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
							},
						},
					},
				},
			},
			&ast.AssignStatement{
				Target: "total",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.ForStatement{
				ValueVar: "item",
				Iterable: &ast.VariableExpr{Name: "items"},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "len",
						Value: &ast.FunctionCallExpr{
							Name: "length",
							Args: []ast.Expr{
								&ast.VariableExpr{Name: "item"},
							},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "total"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() failed: %v", err)
	}

	if len(bytecode) < 4 || string(bytecode[:4]) != "GLYP" {
		t.Error("Invalid bytecode header")
	}
}

// TestCompileArrayOpsMethodCallResolution tests that field access on arrays followed
// by function calls compiles correctly (e.g., arr.length pattern via FieldAccessExpr)
func TestCompileArrayOpsMethodCallResolution(t *testing.T) {
	// Test: $ arr = [1, 2, 3], > arr.length (field access pattern)
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "arr",
				Value: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.FieldAccessExpr{
					Object: &ast.VariableExpr{Name: "arr"},
					Field:  "length",
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() failed: %v", err)
	}

	if len(bytecode) < 4 || string(bytecode[:4]) != "GLYP" {
		t.Error("Invalid bytecode header")
	}
}

// TestCompileArrayOpsAllFunctionsCompile is a comprehensive test that verifies
// all array functional operations compile without error at the AST level
func TestCompileArrayOpsAllFunctionsCompile(t *testing.T) {
	operations := []struct {
		name string
		args int
	}{
		{"map", 2},
		{"filter", 2},
		{"reduce", 3},
		{"find", 2},
		{"some", 2},
		{"every", 2},
		{"sort", 1},
		{"reverse", 1},
		{"flat", 1},
		{"slice", 3},
	}

	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			c := NewCompiler()

			args := make([]ast.Expr, op.args)
			for i := 0; i < op.args; i++ {
				args[i] = &ast.LiteralExpr{Value: ast.IntLiteral{Value: int64(i)}}
			}

			expr := &ast.FunctionCallExpr{
				Name: op.name,
				Args: args,
			}

			err := c.compileFunctionCall(expr)
			if err != nil {
				t.Fatalf("compileFunctionCall(%s) should compile successfully, got: %v", op.name, err)
			}

			// Verify bytecode was generated
			if len(c.code) == 0 {
				t.Errorf("No bytecode generated for %s", op.name)
			}

			// Verify the function name constant exists
			found := false
			for _, constant := range c.constants {
				if sv, ok := constant.(vm.StringValue); ok && sv.Val == op.name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Function name %q not in constants pool", op.name)
			}
		})
	}
}

// TestCompileArrayOpsWithVariableArgs tests that array ops with variable references compile
func TestCompileArrayOpsWithVariableArgs(t *testing.T) {
	c := NewCompiler()

	// Define several variables in the symbol table
	c.symbolTable.Define("arr", c.addConstant(vm.StringValue{Val: "arr"}))
	c.symbolTable.Define("fn", c.addConstant(vm.StringValue{Val: "fn"}))
	c.symbolTable.Define("init", c.addConstant(vm.StringValue{Val: "init"}))

	// Test reduce(arr, fn, init) with all variable references
	expr := &ast.FunctionCallExpr{
		Name: "reduce",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "arr"},
			&ast.VariableExpr{Name: "fn"},
			&ast.VariableExpr{Name: "init"},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(reduce with variable args) failed: %v", err)
	}

	// Verify we have 3 OpLoadVar instructions (one per variable arg)
	loadCount := countOpcode(c.code, vm.OpLoadVar)
	if loadCount != 3 {
		t.Errorf("Expected 3 OpLoadVar instructions, got %d", loadCount)
	}
}

// TestCompileArrayOpsEmptyArray tests compilation with empty array literal
func TestCompileArrayOpsEmptyArray(t *testing.T) {
	c := NewCompiler()
	c.symbolTable.Define("fn", c.addConstant(vm.StringValue{Val: "fn"}))

	// map([], fn)
	expr := &ast.FunctionCallExpr{
		Name: "map",
		Args: []ast.Expr{
			&ast.ArrayExpr{Elements: []ast.Expr{}},
			&ast.VariableExpr{Name: "fn"},
		},
	}

	err := c.compileFunctionCall(expr)
	if err != nil {
		t.Fatalf("compileFunctionCall(map with empty array) failed: %v", err)
	}

	// Should have OpBuildArray with operand 0
	verifyOpcodeSequence(t, c.code, []vm.Opcode{
		vm.OpPush,       // push function name "map"
		vm.OpBuildArray, // build empty array
		vm.OpLoadVar,    // push fn
		vm.OpCall,       // call with 2 args
	})
}

// verifyOpcodeSequence checks that the bytecode contains the expected sequence of opcodes
func verifyOpcodeSequence(t *testing.T, code []byte, expected []vm.Opcode) {
	t.Helper()

	opcodes := extractOpcodes(code)
	if len(opcodes) != len(expected) {
		t.Errorf("Expected %d opcodes, got %d. Opcodes: %v", len(expected), len(opcodes), opcodes)
		return
	}

	for i, exp := range expected {
		if opcodes[i] != exp {
			t.Errorf("Opcode at position %d: expected 0x%02x, got 0x%02x", i, exp, opcodes[i])
		}
	}
}

// extractOpcodes extracts opcodes from raw bytecode (skipping operands)
func extractOpcodes(code []byte) []vm.Opcode {
	var opcodes []vm.Opcode
	i := 0
	for i < len(code) {
		opcode := vm.Opcode(code[i])
		opcodes = append(opcodes, opcode)
		i++
		if hasOperand(code[i-1]) {
			i += 4 // skip 4-byte operand
		}
	}
	return opcodes
}

// extractCallArgCount finds the OpCall instruction and returns its operand (arg count)
func extractCallArgCount(t *testing.T, code []byte) uint32 {
	t.Helper()
	i := 0
	for i < len(code) {
		if vm.Opcode(code[i]) == vm.OpCall {
			if i+4 >= len(code) {
				t.Fatal("OpCall found but operand is truncated")
			}
			return uint32(code[i+1]) | uint32(code[i+2])<<8 | uint32(code[i+3])<<16 | uint32(code[i+4])<<24
		}
		i++
		if hasOperand(code[i-1]) {
			i += 4
		}
	}
	t.Fatal("OpCall instruction not found in bytecode")
	return 0
}

// countOpcode counts the number of times a specific opcode appears in the bytecode
func countOpcode(code []byte, target vm.Opcode) int {
	count := 0
	i := 0
	for i < len(code) {
		if vm.Opcode(code[i]) == target {
			count++
		}
		i++
		if hasOperand(code[i-1]) {
			i += 4
		}
	}
	return count
}
