package compiler

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"math"
	"testing"
	"time"

	"github.com/glyphlang/glyph/pkg/vm"
)

func TestCompileLiteral(t *testing.T) {
	tests := []struct {
		name     string
		expr     *ast.LiteralExpr
		expected vm.Value
	}{
		{
			name:     "int literal",
			expr:     &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
			expected: vm.IntValue{Val: 42},
		},
		{
			name:     "float literal",
			expr:     &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 3.14}},
			expected: vm.FloatValue{Val: 3.14},
		},
		{
			name:     "string literal",
			expr:     &ast.LiteralExpr{Value: ast.StringLiteral{Value: "hello"}},
			expected: vm.StringValue{Val: "hello"},
		},
		{
			name:     "bool literal true",
			expr:     &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
			expected: vm.BoolValue{Val: true},
		},
		{
			name:     "bool literal false",
			expr:     &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
			expected: vm.BoolValue{Val: false},
		},
		{
			name:     "null literal",
			expr:     &ast.LiteralExpr{Value: ast.NullLiteral{}},
			expected: vm.NullValue{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			err := c.compileLiteral(tt.expr)
			if err != nil {
				t.Fatalf("compileLiteral() error: %v", err)
			}

			// Execute and verify
			bytecode := c.buildBytecode()
			vmInstance := vm.NewVM()
			result, err := vmInstance.Execute(bytecode)
			if err != nil {
				t.Fatalf("Execute() error: %v", err)
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCompileBinaryOp(t *testing.T) {
	tests := []struct {
		name     string
		expr     *ast.BinaryOpExpr
		expected vm.Value
	}{
		{
			name: "5 + 3",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
			},
			expected: vm.IntValue{Val: 8},
		},
		{
			name: "10 - 4",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Sub,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
			},
			expected: vm.IntValue{Val: 6},
		},
		{
			name: "6 * 7",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 6}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 7}},
			},
			expected: vm.IntValue{Val: 42},
		},
		{
			name: "42 == 42",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Eq,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
			},
			expected: vm.BoolValue{Val: true},
		},
		{
			name: "5 > 3",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Gt,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
			},
			expected: vm.BoolValue{Val: true},
		},
		{
			name: "null == null",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Eq,
				Left:  &ast.LiteralExpr{Value: ast.NullLiteral{}},
				Right: &ast.LiteralExpr{Value: ast.NullLiteral{}},
			},
			expected: vm.BoolValue{Val: true},
		},
		{
			name: "null != 42",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Ne,
				Left:  &ast.LiteralExpr{Value: ast.NullLiteral{}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
			},
			expected: vm.BoolValue{Val: true},
		},
		{
			name: "42 != null",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Ne,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
				Right: &ast.LiteralExpr{Value: ast.NullLiteral{}},
			},
			expected: vm.BoolValue{Val: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			err := c.compileExpression(tt.expr)
			if err != nil {
				t.Fatalf("compileExpression() error: %v", err)
			}

			c.emit(vm.OpHalt)

			// Execute and verify
			bytecode := c.buildBytecode()
			vmInstance := vm.NewVM()
			result, err := vmInstance.Execute(bytecode)
			if err != nil {
				t.Fatalf("Execute() error: %v", err)
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCompileVariableAssignment(t *testing.T) {
	// Test: $ x = 42, > x
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "x"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 42}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileVariableRedeclaration(t *testing.T) {
	// Test: $ x = 1, $ x = 2 (should fail - redeclaration in same scope)
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
		},
	}

	c := NewCompiler()
	_, err := c.CompileRoute(route)
	if err == nil {
		t.Fatal("Expected redeclaration error, got nil")
	}
	if !IsSemanticError(err) {
		t.Errorf("Expected SemanticError, got %T: %v", err, err)
	}
	expectedMsg := "cannot redeclare variable 'x' in the same scope"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestCompileVariableUpdateInNestedScope(t *testing.T) {
	// Test: $ cond = true, $ x = 0, if (cond) { $ x = 1 }, > x
	// Should work - updating outer variable from nested scope
	// Note: Using OptNone to prevent optimizer from inlining the if block
	// (constant propagation would make the condition always true and inline the block)
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "cond",
				Value:  &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
			},
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.IfStatement{
				Condition: &ast.VariableExpr{Name: "cond"},
				ThenBlock: []ast.Statement{
					&ast.AssignStatement{
						Target: "x",
						Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "x"},
			},
		},
	}

	c := NewCompilerWithOptLevel(OptNone)
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 1}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileReassignment(t *testing.T) {
	// Test: $ x = 0, x = x + 1, x = x + 1, > x
	// Should work - declaration followed by reassignment
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.ReassignStatement{
				Target: "x",
				Value: &ast.BinaryOpExpr{
					Op:    ast.Add,
					Left:  &ast.VariableExpr{Name: "x"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				},
			},
			&ast.ReassignStatement{
				Target: "x",
				Value: &ast.BinaryOpExpr{
					Op:    ast.Add,
					Left:  &ast.VariableExpr{Name: "x"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "x"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 2}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileReassignmentUndeclaredVariable(t *testing.T) {
	// Test: x = 1 (without prior declaration) should fail
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ReassignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
		},
	}

	c := NewCompiler()
	_, err := c.CompileRoute(route)
	if err == nil {
		t.Fatal("Expected undeclared variable error, got nil")
	}
	if !IsSemanticError(err) {
		t.Errorf("Expected SemanticError, got %T: %v", err, err)
	}
	expectedMsg := "cannot assign to undeclared variable 'x'"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestCompileArithmeticExpression(t *testing.T) {
	// Test: $ result = 5 + 3 * 2, > result
	// Expected: 11 (5 + (3 * 2))
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "result",
				Value: &ast.BinaryOpExpr{
					Op:   ast.Add,
					Left: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
					Right: &ast.BinaryOpExpr{
						Op:    ast.Mul,
						Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
						Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
					},
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
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 11}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileObjectLiteral(t *testing.T) {
	// Test: > {name: "Alice", age: 30}
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{
							Key:   "name",
							Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "Alice"}},
						},
						{
							Key:   "age",
							Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 30}},
						},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	objVal, ok := result.(vm.ObjectValue)
	if !ok {
		t.Fatalf("Expected ObjectValue, got %T", result)
	}

	if len(objVal.Val) != 2 {
		t.Errorf("Expected object with 2 fields, got %d", len(objVal.Val))
	}

	if nameVal, ok := objVal.Val["name"].(vm.StringValue); !ok || nameVal.Val != "Alice" {
		t.Errorf("Expected name='Alice', got %v", objVal.Val["name"])
	}

	if ageVal, ok := objVal.Val["age"].(vm.IntValue); !ok || ageVal.Val != 30 {
		t.Errorf("Expected age=30, got %v", objVal.Val["age"])
	}
}

func TestCompileArrayLiteral(t *testing.T) {
	// Test: > [1, 2, 3]
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	arrVal, ok := result.(vm.ArrayValue)
	if !ok {
		t.Fatalf("Expected ArrayValue, got %T", result)
	}

	if len(arrVal.Val) != 3 {
		t.Errorf("Expected array length 3, got %d", len(arrVal.Val))
	}

	expected := []int64{1, 2, 3}
	for i, exp := range expected {
		if intVal, ok := arrVal.Val[i].(vm.IntValue); !ok || intVal.Val != exp {
			t.Errorf("Expected element %d to be %d, got %v", i, exp, arrVal.Val[i])
		}
	}
}

func TestCompileFieldAccess(t *testing.T) {
	// Test: $ obj = {name: "Alice"}, > obj.name
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "obj",
				Value: &ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{
							Key:   "name",
							Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "Alice"}},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.FieldAccessExpr{
					Object: &ast.VariableExpr{Name: "obj"},
					Field:  "name",
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.StringValue{Val: "Alice"}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileIfStatement(t *testing.T) {
	// Test: if 5 > 3 { > 1 } else { > 2 }
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.IfStatement{
				Condition: &ast.BinaryOpExpr{
					Op:    ast.Gt,
					Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
				},
				ThenBlock: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 1} // Should execute then block
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileWhileLoop(t *testing.T) {
	// t.Skip("Skipping - pre-existing VM issue with while loops")
	// Test: $ x = 0, while x < 3 { $ x = x + 1 }, > x
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.WhileStatement{
				Condition: &ast.BinaryOpExpr{
					Op:    ast.Lt,
					Left:  &ast.VariableExpr{Name: "x"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "x",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "x"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "x"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 3} // Should loop 3 times
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileWhileLoopSum(t *testing.T) {
	// Test: $ sum = 0, $ i = 1, while i <= 5 { $ sum = sum + i, $ i = i + 1 }, > sum
	// Expected: 1+2+3+4+5 = 15
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "sum",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.AssignStatement{
				Target: "i",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
			&ast.WhileStatement{
				Condition: &ast.BinaryOpExpr{
					Op:    ast.Le,
					Left:  &ast.VariableExpr{Name: "i"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "sum",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "sum"},
							Right: &ast.VariableExpr{Name: "i"},
						},
					},
					&ast.AssignStatement{
						Target: "i",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "i"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "sum"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 15}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileWhileLoopNoIterations(t *testing.T) {
	// Test: $ x = 10, while x < 5 { $ x = x + 1 }, > x
	// Condition is false from start, so loop never executes
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
			},
			&ast.WhileStatement{
				Condition: &ast.BinaryOpExpr{
					Op:    ast.Lt,
					Left:  &ast.VariableExpr{Name: "x"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "x",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "x"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "x"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 10} // Loop never executed
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileWhileLoopStringConcat(t *testing.T) {
	// Test: $ s = "", $ i = 0, while i < 3 { $ s = s + "x", $ i = i + 1 }, > s
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "s",
				Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: ""}},
			},
			&ast.AssignStatement{
				Target: "i",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.WhileStatement{
				Condition: &ast.BinaryOpExpr{
					Op:    ast.Lt,
					Left:  &ast.VariableExpr{Name: "i"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "s",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "s"},
							Right: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "x"}},
						},
					},
					&ast.AssignStatement{
						Target: "i",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "i"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "s"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.StringValue{Val: "xxx"}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileWhileLoopWithNestedIf(t *testing.T) {
	// Test: $ count = 0, $ i = 0, while i < 10 { if i > 4 { $ count = count + 1 }, $ i = i + 1 }, > count
	// Counts numbers from 5 to 9 (5 numbers)
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "count",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.AssignStatement{
				Target: "i",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.WhileStatement{
				Condition: &ast.BinaryOpExpr{
					Op:    ast.Lt,
					Left:  &ast.VariableExpr{Name: "i"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
				},
				Body: []ast.Statement{
					&ast.IfStatement{
						Condition: &ast.BinaryOpExpr{
							Op:    ast.Gt,
							Left:  &ast.VariableExpr{Name: "i"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
						},
						ThenBlock: []ast.Statement{
							&ast.AssignStatement{
								Target: "count",
								Value: &ast.BinaryOpExpr{
									Op:    ast.Add,
									Left:  &ast.VariableExpr{Name: "count"},
									Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
								},
							},
						},
					},
					&ast.AssignStatement{
						Target: "i",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "i"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "count"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 5} // i=5,6,7,8,9 satisfy i > 4
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileWhileLoopMultiply(t *testing.T) {
	// Test: $ result = 1, $ i = 0, while i < 4 { $ result = result * 2, $ i = i + 1 }, > result
	// 1 * 2 * 2 * 2 * 2 = 16
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "result",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
			&ast.AssignStatement{
				Target: "i",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.WhileStatement{
				Condition: &ast.BinaryOpExpr{
					Op:    ast.Lt,
					Left:  &ast.VariableExpr{Name: "i"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "result",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Mul,
							Left:  &ast.VariableExpr{Name: "result"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						},
					},
					&ast.AssignStatement{
						Target: "i",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "i"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
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
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 16}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileRouteWithPathParameter(t *testing.T) {
	// Test: GET /users/:id -> > id
	route := &ast.Route{
		Path:   "/users/:id",
		Method: ast.Get,
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "id"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute with id="123"
	vmInstance := vm.NewVM()
	vmInstance.SetLocal("id", vm.StringValue{Val: "123"})
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.StringValue{Val: "123"}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileRouteWithMultiplePathParameters(t *testing.T) {
	// Test: GET /users/:userId/posts/:postId -> > userId + postId
	route := &ast.Route{
		Path:   "/users/:userId/posts/:postId",
		Method: ast.Get,
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.BinaryOpExpr{
					Op:    ast.Add,
					Left:  &ast.VariableExpr{Name: "userId"},
					Right: &ast.VariableExpr{Name: "postId"},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute with userId="42" and postId="100"
	vmInstance := vm.NewVM()
	vmInstance.SetLocal("userId", vm.StringValue{Val: "user-"})
	vmInstance.SetLocal("postId", vm.StringValue{Val: "post-1"})
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.StringValue{Val: "user-post-1"}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileRouteWithInjection(t *testing.T) {
	// Test: GET /test % db: Database -> > db
	route := &ast.Route{
		Path:   "/test",
		Method: ast.Get,
		Injections: []ast.Injection{
			{
				Name: "db",
				Type: ast.DatabaseType{},
			},
		},
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "db"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute with db injected
	vmInstance := vm.NewVM()
	// Simulate an injected database object
	dbObject := vm.ObjectValue{Val: map[string]vm.Value{
		"connected": vm.BoolValue{Val: true},
	}}
	vmInstance.SetLocal("db", dbObject)
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// Check that we got the database object back
	resultObj, ok := result.(vm.ObjectValue)
	if !ok {
		t.Fatalf("Expected ObjectValue, got %T", result)
	}

	if connectedVal, ok := resultObj.Val["connected"].(vm.BoolValue); !ok || !connectedVal.Val {
		t.Errorf("Expected database object with connected=true, got %v", resultObj)
	}
}

func TestCompileRouteWithParameterAndInjection(t *testing.T) {
	// Test: GET /users/:id % db: Database -> > {id: id, db: db}
	route := &ast.Route{
		Path:   "/users/:id",
		Method: ast.Get,
		Injections: []ast.Injection{
			{
				Name: "db",
				Type: ast.DatabaseType{},
			},
		},
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{
							Key:   "id",
							Value: &ast.VariableExpr{Name: "id"},
						},
						{
							Key:   "hasDb",
							Value: &ast.VariableExpr{Name: "db"},
						},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute with both parameter and injection
	vmInstance := vm.NewVM()
	vmInstance.SetLocal("id", vm.StringValue{Val: "user123"})
	vmInstance.SetLocal("db", vm.StringValue{Val: "mock-db"})
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// Check result
	resultObj, ok := result.(vm.ObjectValue)
	if !ok {
		t.Fatalf("Expected ObjectValue, got %T", result)
	}

	if idVal, ok := resultObj.Val["id"].(vm.StringValue); !ok || idVal.Val != "user123" {
		t.Errorf("Expected id='user123', got %v", resultObj.Val["id"])
	}

	if dbVal, ok := resultObj.Val["hasDb"].(vm.StringValue); !ok || dbVal.Val != "mock-db" {
		t.Errorf("Expected hasDb='mock-db', got %v", resultObj.Val["hasDb"])
	}
}

func TestCompileForLoopSimpleAssign(t *testing.T) {
	// Simpler test: $ result = 0, for item in [42] { $ result = item }, > result
	// Expected: 42 (just assigns the item, no arithmetic)
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "result",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.ForStatement{
				ValueVar: "item",
				Iterable: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
					},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "result",
						Value:  &ast.VariableExpr{Name: "item"},
					},
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
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 42}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileForLoopSimple(t *testing.T) {
	// Test: $ sum = 0, for item in [1, 2, 3] { $ sum = sum + item }, > sum
	// Expected: 6
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "sum",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.ForStatement{
				ValueVar: "item",
				Iterable: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
					},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "sum",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "sum"},
							Right: &ast.VariableExpr{Name: "item"},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "sum"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 6}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileForLoopWithIndex(t *testing.T) {
	// Test: $ result = 0, for idx, val in [10, 20, 30] { $ result = result + idx }, > result
	// Expected: 0 + 1 + 2 = 3
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "result",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.ForStatement{
				KeyVar:   "idx",
				ValueVar: "val",
				Iterable: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 30}},
					},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "result",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "result"},
							Right: &ast.VariableExpr{Name: "idx"},
						},
					},
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
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 3}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileForLoopWithIndexAndValue(t *testing.T) {
	// Test using BOTH index and value: for idx, val in [10, 20] { $ result = result + idx + val }
	// Expected: 0 + (0 + 10) + (1 + 20) = 31
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "result",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.ForStatement{
				KeyVar:   "idx",
				ValueVar: "val",
				Iterable: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}},
					},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "result",
						Value: &ast.BinaryOpExpr{
							Op: ast.Add,
							Left: &ast.BinaryOpExpr{
								Op:    ast.Add,
								Left:  &ast.VariableExpr{Name: "result"},
								Right: &ast.VariableExpr{Name: "idx"},
							},
							Right: &ast.VariableExpr{Name: "val"},
						},
					},
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
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 31}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileForLoopObjectCreation(t *testing.T) {
	// Test: for idx, val in [10, 20] { $ entry = {pos: idx, val: val}, $ arr = arr + [entry] }
	// Similar to /items endpoint
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "arr",
				Value:  &ast.ArrayExpr{Elements: []ast.Expr{}},
			},
			&ast.ForStatement{
				KeyVar:   "idx",
				ValueVar: "val",
				Iterable: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}},
					},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "entry",
						Value: &ast.ObjectExpr{
							Fields: []ast.ObjectField{
								{Key: "pos", Value: &ast.VariableExpr{Name: "idx"}},
								{Key: "val", Value: &ast.VariableExpr{Name: "val"}},
							},
						},
					},
					&ast.AssignStatement{
						Target: "arr",
						Value: &ast.BinaryOpExpr{
							Op:   ast.Add,
							Left: &ast.VariableExpr{Name: "arr"},
							Right: &ast.ArrayExpr{
								Elements: []ast.Expr{
									&ast.VariableExpr{Name: "entry"},
								},
							},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "arr"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	arrVal, ok := result.(vm.ArrayValue)
	if !ok {
		t.Fatalf("Expected ArrayValue, got %T", result)
	}

	if len(arrVal.Val) != 2 {
		t.Fatalf("Expected array length 2, got %d", len(arrVal.Val))
	}

	// Check first element: {pos: 0, val: 10}
	obj0, ok := arrVal.Val[0].(vm.ObjectValue)
	if !ok {
		t.Fatalf("Expected ObjectValue at index 0, got %T", arrVal.Val[0])
	}
	if pos, ok := obj0.Val["pos"].(vm.IntValue); !ok || pos.Val != 0 {
		t.Errorf("Expected pos=0, got %v", obj0.Val["pos"])
	}
	if val, ok := obj0.Val["val"].(vm.IntValue); !ok || val.Val != 10 {
		t.Errorf("Expected val=10, got %v", obj0.Val["val"])
	}
}

func TestCompileForLoopEmptyArray(t *testing.T) {
	// Test: $ sum = 42, for item in [] { $ sum = 0 }, > sum
	// Expected: 42 (loop body never executes)
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "sum",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
			},
			&ast.ForStatement{
				ValueVar: "item",
				Iterable: &ast.ArrayExpr{
					Elements: []ast.Expr{},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "sum",
						Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "sum"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 42}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileForLoopStringConcat(t *testing.T) {
	// Test: $ result = "", for s in ["a", "b", "c"] { $ result = result + s }, > result
	// Expected: "abc"
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "result",
				Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: ""}},
			},
			&ast.ForStatement{
				ValueVar: "s",
				Iterable: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.StringLiteral{Value: "a"}},
						&ast.LiteralExpr{Value: ast.StringLiteral{Value: "b"}},
						&ast.LiteralExpr{Value: ast.StringLiteral{Value: "c"}},
					},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "result",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "result"},
							Right: &ast.VariableExpr{Name: "s"},
						},
					},
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
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.StringValue{Val: "abc"}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileForLoopArrayConcatenation(t *testing.T) {
	// Test: $ arr = [], for x in [1, 2] { $ arr = arr + [x] }, > arr
	// Expected: [1, 2]
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "arr",
				Value: &ast.ArrayExpr{
					Elements: []ast.Expr{},
				},
			},
			&ast.ForStatement{
				ValueVar: "x",
				Iterable: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
					},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "arr",
						Value: &ast.BinaryOpExpr{
							Op:   ast.Add,
							Left: &ast.VariableExpr{Name: "arr"},
							Right: &ast.ArrayExpr{
								Elements: []ast.Expr{
									&ast.VariableExpr{Name: "x"},
								},
							},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "arr"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	arrVal, ok := result.(vm.ArrayValue)
	if !ok {
		t.Fatalf("Expected ArrayValue, got %T", result)
	}

	if len(arrVal.Val) != 2 {
		t.Fatalf("Expected array length 2, got %d", len(arrVal.Val))
	}

	for i, expected := range []int64{1, 2} {
		if intVal, ok := arrVal.Val[i].(vm.IntValue); !ok || intVal.Val != expected {
			t.Errorf("Expected element %d to be %d, got %v", i, expected, arrVal.Val[i])
		}
	}
}

func TestCompileForLoopNested(t *testing.T) {
	// Test: $ sum = 0, for row in [[1,2],[3,4]] { for cell in row { $ sum = sum + cell } }, > sum
	// Expected: 10 (1+2+3+4)
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "sum",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.ForStatement{
				ValueVar: "row",
				Iterable: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.ArrayExpr{
							Elements: []ast.Expr{
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
							},
						},
						&ast.ArrayExpr{
							Elements: []ast.Expr{
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
							},
						},
					},
				},
				Body: []ast.Statement{
					&ast.ForStatement{
						ValueVar: "cell",
						Iterable: &ast.VariableExpr{Name: "row"},
						Body: []ast.Statement{
							&ast.AssignStatement{
								Target: "sum",
								Value: &ast.BinaryOpExpr{
									Op:    ast.Add,
									Left:  &ast.VariableExpr{Name: "sum"},
									Right: &ast.VariableExpr{Name: "cell"},
								},
							},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "sum"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 10}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Switch statement tests

func TestCompileSwitchSimpleStringMatch(t *testing.T) {
	// Test: $ status = "pending", switch status { case "pending" { > "matched" } default { > "no match" } }
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "status",
				Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: "pending"}},
			},
			&ast.SwitchStatement{
				Value: &ast.VariableExpr{Name: "status"},
				Cases: []ast.SwitchCase{
					{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "pending"}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "matched"}},
							},
						},
					},
				},
				Default: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "no match"}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.StringValue{Val: "matched"}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileSwitchSecondCaseMatch(t *testing.T) {
	// Test: switch "shipped" { case "pending" { > 1 } case "shipped" { > 2 } default { > 3 } }
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.SwitchStatement{
				Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "shipped"}},
				Cases: []ast.SwitchCase{
					{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "pending"}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
							},
						},
					},
					{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "shipped"}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
							},
						},
					},
				},
				Default: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 2}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileSwitchDefaultCase(t *testing.T) {
	// Test: switch "unknown" { case "a" { > 1 } case "b" { > 2 } default { > 99 } }
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.SwitchStatement{
				Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "unknown"}},
				Cases: []ast.SwitchCase{
					{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "a"}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
							},
						},
					},
					{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "b"}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
							},
						},
					},
				},
				Default: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 99}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 99}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileSwitchIntegerValues(t *testing.T) {
	// Test: switch 42 { case 1 { > "one" } case 42 { > "forty-two" } default { > "other" } }
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.SwitchStatement{
				Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
				Cases: []ast.SwitchCase{
					{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "one"}},
							},
						},
					},
					{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "forty-two"}},
							},
						},
					},
				},
				Default: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "other"}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.StringValue{Val: "forty-two"}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileSwitchWithVariableAssignment(t *testing.T) {
	// Test: $ result = "", $ n = 2, switch n { case 1 { $ result = "one" } case 2 { $ result = "two" } }, > result
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "result",
				Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: ""}},
			},
			&ast.AssignStatement{
				Target: "n",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
			&ast.SwitchStatement{
				Value: &ast.VariableExpr{Name: "n"},
				Cases: []ast.SwitchCase{
					{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						Body: []ast.Statement{
							&ast.AssignStatement{
								Target: "result",
								Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: "one"}},
							},
						},
					},
					{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						Body: []ast.Statement{
							&ast.AssignStatement{
								Target: "result",
								Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: "two"}},
							},
						},
					},
				},
				Default: []ast.Statement{
					&ast.AssignStatement{
						Target: "result",
						Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: "default"}},
					},
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
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.StringValue{Val: "two"}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileSwitchNoMatchNoDefault(t *testing.T) {
	// Test: $ result = "unchanged", switch "x" { case "a" { $ result = "a" } case "b" { $ result = "b" } }, > result
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "result",
				Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: "unchanged"}},
			},
			&ast.SwitchStatement{
				Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "x"}},
				Cases: []ast.SwitchCase{
					{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "a"}},
						Body: []ast.Statement{
							&ast.AssignStatement{
								Target: "result",
								Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: "a"}},
							},
						},
					},
					{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "b"}},
						Body: []ast.Statement{
							&ast.AssignStatement{
								Target: "result",
								Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: "b"}},
							},
						},
					},
				},
				Default: nil, // No default case
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "result"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.StringValue{Val: "unchanged"}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileSwitchObjectReturn(t *testing.T) {
	// Test: switch "ok" { case "ok" { > {status: "success", code: 200} } default { > {status: "error", code: 500} } }
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.SwitchStatement{
				Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "ok"}},
				Cases: []ast.SwitchCase{
					{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "ok"}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.ObjectExpr{
									Fields: []ast.ObjectField{
										{
											Key:   "status",
											Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "success"}},
										},
										{
											Key:   "code",
											Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 200}},
										},
									},
								},
							},
						},
					},
				},
				Default: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.ObjectExpr{
							Fields: []ast.ObjectField{
								{
									Key:   "status",
									Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "error"}},
								},
								{
									Key:   "code",
									Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 500}},
								},
							},
						},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	objVal, ok := result.(vm.ObjectValue)
	if !ok {
		t.Fatalf("Expected ObjectValue, got %T", result)
	}

	if statusVal, ok := objVal.Val["status"].(vm.StringValue); !ok || statusVal.Val != "success" {
		t.Errorf("Expected status='success', got %v", objVal.Val["status"])
	}

	if codeVal, ok := objVal.Val["code"].(vm.IntValue); !ok || codeVal.Val != 200 {
		t.Errorf("Expected code=200, got %v", objVal.Val["code"])
	}
}

func TestCompileSwitchMultipleCases(t *testing.T) {
	// Test switch with many cases to verify jump patching works correctly
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.SwitchStatement{
				Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
				Cases: []ast.SwitchCase{
					{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "one"}},
							},
						},
					},
					{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "two"}},
							},
						},
					},
					{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "three"}},
							},
						},
					},
					{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "four"}},
							},
						},
					},
					{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "five"}},
							},
						},
					},
				},
				Default: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "other"}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.StringValue{Val: "five"}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Function call tests

func TestCompileFunctionCallNoArgs(t *testing.T) {
	// Test: > now()
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.FunctionCallExpr{
					Name: "now",
					Args: []ast.Expr{},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// now() returns the current Unix timestamp
	intVal, ok := result.(vm.IntValue)
	if !ok {
		t.Fatalf("Expected IntValue, got %T", result)
	}
	if math.Abs(float64(intVal.Val-time.Now().Unix())) > 2 {
		t.Errorf("Expected timestamp within 2s of now, got %v", intVal.Val)
	}
}

func TestCompileFunctionCallWithOneArg(t *testing.T) {
	// Test: > length([1, 2, 3])
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.FunctionCallExpr{
					Name: "length",
					Args: []ast.Expr{
						&ast.ArrayExpr{
							Elements: []ast.Expr{
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
							},
						},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 3}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileFunctionCallLengthString(t *testing.T) {
	// Test: > length("hello")
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.FunctionCallExpr{
					Name: "length",
					Args: []ast.Expr{
						&ast.LiteralExpr{Value: ast.StringLiteral{Value: "hello"}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 5}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileFunctionCallInExpression(t *testing.T) {
	// Test: > length([1, 2, 3, 4, 5]) + 10
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.BinaryOpExpr{
					Op: ast.Add,
					Left: &ast.FunctionCallExpr{
						Name: "length",
						Args: []ast.Expr{
							&ast.ArrayExpr{
								Elements: []ast.Expr{
									&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
									&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
									&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
									&ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
									&ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
								},
							},
						},
					},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 15} // 5 + 10
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileFunctionCallWithVariable(t *testing.T) {
	// Test: $ arr = [1, 2, 3, 4], > length(arr)
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "arr",
				Value: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.FunctionCallExpr{
					Name: "length",
					Args: []ast.Expr{
						&ast.VariableExpr{Name: "arr"},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 4}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileFunctionCallTimeNow(t *testing.T) {
	// Test: > time.now()
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.FunctionCallExpr{
					Name: "time.now",
					Args: []ast.Expr{},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// time.now() returns the current Unix timestamp
	intVal, ok := result.(vm.IntValue)
	if !ok {
		t.Fatalf("Expected IntValue, got %T", result)
	}
	if math.Abs(float64(intVal.Val-time.Now().Unix())) > 2 {
		t.Errorf("Expected timestamp within 2s of now, got %v", intVal.Val)
	}
}

func TestCompileFunctionCallInObject(t *testing.T) {
	// Test: > {timestamp: now(), items: length([1, 2])}
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{
							Key: "timestamp",
							Value: &ast.FunctionCallExpr{
								Name: "now",
								Args: []ast.Expr{},
							},
						},
						{
							Key: "items",
							Value: &ast.FunctionCallExpr{
								Name: "length",
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
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	objVal, ok := result.(vm.ObjectValue)
	if !ok {
		t.Fatalf("Expected ObjectValue, got %T", result)
	}

	if tsVal, ok := objVal.Val["timestamp"].(vm.IntValue); !ok || math.Abs(float64(tsVal.Val-time.Now().Unix())) > 2 {
		t.Errorf("Expected timestamp within 2s of now, got %v", objVal.Val["timestamp"])
	}

	if itemsVal, ok := objVal.Val["items"].(vm.IntValue); !ok || itemsVal.Val != 2 {
		t.Errorf("Expected items=2, got %v", objVal.Val["items"])
	}
}

func TestCompileFunctionCallInCondition(t *testing.T) {
	// Test: $ arr = [1, 2, 3], if length(arr) > 2 { > "big" } else { > "small" }
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
			&ast.IfStatement{
				Condition: &ast.BinaryOpExpr{
					Op: ast.Gt,
					Left: &ast.FunctionCallExpr{
						Name: "length",
						Args: []ast.Expr{
							&ast.VariableExpr{Name: "arr"},
						},
					},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
				},
				ThenBlock: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "big"}},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "small"}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.StringValue{Val: "big"} // length([1,2,3]) = 3 > 2
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Test CompileCommand

func TestCompileCommand_Simple(t *testing.T) {
	cmd := &ast.Command{
		Name: "greet",
		Params: []ast.CommandParam{
			{Name: "name", Type: ast.StringType{}, Required: true},
		},
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.BinaryOpExpr{
					Op:    ast.Add,
					Left:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: "Hello, "}},
					Right: &ast.VariableExpr{Name: "name"},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileCommand(cmd)
	if err != nil {
		t.Fatalf("CompileCommand() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestCompileCommand_WithMultipleParams(t *testing.T) {
	cmd := &ast.Command{
		Name: "add",
		Params: []ast.CommandParam{
			{Name: "x", Type: ast.IntType{}, Required: true},
			{Name: "y", Type: ast.IntType{}, Required: true},
		},
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.BinaryOpExpr{
					Op:    ast.Add,
					Left:  &ast.VariableExpr{Name: "x"},
					Right: &ast.VariableExpr{Name: "y"},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileCommand(cmd)
	if err != nil {
		t.Fatalf("CompileCommand() error: %v", err)
	}

	// Verify bytecode structure
	if len(bytecode) < 8 { // Magic(4) + Version(4)
		t.Error("Bytecode too short")
	}
}

func TestCompileCommand_WithComplexBody(t *testing.T) {
	cmd := &ast.Command{
		Name: "process",
		Params: []ast.CommandParam{
			{Name: "input", Type: ast.StringType{}, Required: true},
		},
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "result",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.AssignStatement{
				Target: "i",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.WhileStatement{
				Condition: &ast.BinaryOpExpr{
					Op:    ast.Lt,
					Left:  &ast.VariableExpr{Name: "i"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "result",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "result"},
							Right: &ast.VariableExpr{Name: "i"},
						},
					},
					&ast.AssignStatement{
						Target: "i",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "i"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "result"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileCommand(cmd)
	if err != nil {
		t.Fatalf("CompileCommand() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

// Test CompileCronTask

func TestCompileCronTask_Simple(t *testing.T) {
	task := &ast.CronTask{
		Name:     "daily_cleanup",
		Schedule: "0 0 * * *",
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "count",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
			},
			&ast.ReturnStatement{
				Value: &ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{Key: "deleted", Value: &ast.VariableExpr{Name: "count"}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileCronTask(task)
	if err != nil {
		t.Fatalf("CompileCronTask() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestCompileCronTask_WithInjections(t *testing.T) {
	task := &ast.CronTask{
		Name:     "backup",
		Schedule: "0 0 * * *",
		Injections: []ast.Injection{
			{Name: "db", Type: ast.DatabaseType{}},
		},
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "count",
				Value: &ast.FieldAccessExpr{
					Object: &ast.VariableExpr{Name: "db"},
					Field:  "count",
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "count"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileCronTask(task)
	if err != nil {
		t.Fatalf("CompileCronTask() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestCompileCronTask_ComplexLogic(t *testing.T) {
	task := &ast.CronTask{
		Name:     "hourly_sync",
		Schedule: "0 */1 * * *",
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "total",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.AssignStatement{
				Target: "i",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.WhileStatement{
				Condition: &ast.BinaryOpExpr{
					Op:    ast.Lt,
					Left:  &ast.VariableExpr{Name: "i"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
				},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "total",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "total"},
							Right: &ast.VariableExpr{Name: "i"},
						},
					},
					&ast.AssignStatement{
						Target: "i",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "i"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
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
	bytecode, err := c.CompileCronTask(task)
	if err != nil {
		t.Fatalf("CompileCronTask() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

// Test CompileEventHandler

func TestCompileEventHandler_Simple(t *testing.T) {
	handler := &ast.EventHandler{
		EventType: "user.created",
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "userId",
				Value: &ast.FieldAccessExpr{
					Object: &ast.VariableExpr{Name: "event"},
					Field:  "userId",
				},
			},
			&ast.ReturnStatement{
				Value: &ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{Key: "handled", Value: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}}},
						{Key: "userId", Value: &ast.VariableExpr{Name: "userId"}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileEventHandler(handler)
	if err != nil {
		t.Fatalf("CompileEventHandler() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestCompileEventHandler_WithInput(t *testing.T) {
	handler := &ast.EventHandler{
		EventType: "order.paid",
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "orderId",
				Value: &ast.FieldAccessExpr{
					Object: &ast.VariableExpr{Name: "input"},
					Field:  "orderId",
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "orderId"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileEventHandler(handler)
	if err != nil {
		t.Fatalf("CompileEventHandler() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestCompileEventHandler_WithInjections(t *testing.T) {
	handler := &ast.EventHandler{
		EventType: "notification.send",
		Injections: []ast.Injection{
			{Name: "db", Type: ast.DatabaseType{}},
		},
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "user",
				Value: &ast.FieldAccessExpr{
					Object: &ast.VariableExpr{Name: "db"},
					Field:  "user",
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "user"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileEventHandler(handler)
	if err != nil {
		t.Fatalf("CompileEventHandler() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestCompileEventHandler_ComplexProcessing(t *testing.T) {
	handler := &ast.EventHandler{
		EventType: "data.process",
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "data",
				Value: &ast.FieldAccessExpr{
					Object: &ast.VariableExpr{Name: "event"},
					Field:  "data",
				},
			},
			&ast.IfStatement{
				Condition: &ast.BinaryOpExpr{
					Op:    ast.Ne,
					Left:  &ast.VariableExpr{Name: "data"},
					Right: &ast.LiteralExpr{Value: ast.NullLiteral{}},
				},
				ThenBlock: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileEventHandler(handler)
	if err != nil {
		t.Fatalf("CompileEventHandler() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

// Test CompileQueueWorker

func TestCompileQueueWorker_Simple(t *testing.T) {
	worker := &ast.QueueWorker{
		QueueName: "email.send",
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "to",
				Value: &ast.FieldAccessExpr{
					Object: &ast.VariableExpr{Name: "message"},
					Field:  "to",
				},
			},
			&ast.ReturnStatement{
				Value: &ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{Key: "sent", Value: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}}},
						{Key: "to", Value: &ast.VariableExpr{Name: "to"}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileQueueWorker(worker)
	if err != nil {
		t.Fatalf("CompileQueueWorker() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestCompileQueueWorker_WithInput(t *testing.T) {
	worker := &ast.QueueWorker{
		QueueName: "image.resize",
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "imageId",
				Value: &ast.FieldAccessExpr{
					Object: &ast.VariableExpr{Name: "input"},
					Field:  "imageId",
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "imageId"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileQueueWorker(worker)
	if err != nil {
		t.Fatalf("CompileQueueWorker() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestCompileQueueWorker_WithInjections(t *testing.T) {
	worker := &ast.QueueWorker{
		QueueName: "report.generate",
		Injections: []ast.Injection{
			{Name: "db", Type: ast.DatabaseType{}},
		},
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "data",
				Value: &ast.FieldAccessExpr{
					Object: &ast.VariableExpr{Name: "db"},
					Field:  "data",
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "data"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileQueueWorker(worker)
	if err != nil {
		t.Fatalf("CompileQueueWorker() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestCompileQueueWorker_ComplexProcessing(t *testing.T) {
	worker := &ast.QueueWorker{
		QueueName: "data.process",
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "items",
				Value: &ast.FieldAccessExpr{
					Object: &ast.VariableExpr{Name: "message"},
					Field:  "items",
				},
			},
			&ast.AssignStatement{
				Target: "count",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.ForStatement{
				ValueVar: "item",
				Iterable: &ast.VariableExpr{Name: "items"},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "count",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "count"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "count"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileQueueWorker(worker)
	if err != nil {
		t.Fatalf("CompileQueueWorker() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

// Test compilation with multiple statement types

func TestCompile_DirectiveWithIfStatement(t *testing.T) {
	cmd := &ast.Command{
		Name: "check",
		Params: []ast.CommandParam{
			{Name: "value", Type: ast.IntType{}, Required: true},
		},
		Body: []ast.Statement{
			&ast.IfStatement{
				Condition: &ast.BinaryOpExpr{
					Op:    ast.Gt,
					Left:  &ast.VariableExpr{Name: "value"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
				},
				ThenBlock: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "positive"}},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "negative"}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileCommand(cmd)
	if err != nil {
		t.Fatalf("CompileCommand() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestCompile_DirectiveWithForLoop(t *testing.T) {
	task := &ast.CronTask{
		Name:     "process_items",
		Schedule: "0 * * * *",
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "items",
				Value: &ast.ArrayExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
					},
				},
			},
			&ast.AssignStatement{
				Target: "sum",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.ForStatement{
				ValueVar: "item",
				Iterable: &ast.VariableExpr{Name: "items"},
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "sum",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "sum"},
							Right: &ast.VariableExpr{Name: "item"},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "sum"},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileCronTask(task)
	if err != nil {
		t.Fatalf("CompileCronTask() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestCompile_DirectiveWithSwitchStatement(t *testing.T) {
	handler := &ast.EventHandler{
		EventType: "status.changed",
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "status",
				Value: &ast.FieldAccessExpr{
					Object: &ast.VariableExpr{Name: "event"},
					Field:  "status",
				},
			},
			&ast.SwitchStatement{
				Value: &ast.VariableExpr{Name: "status"},
				Cases: []ast.SwitchCase{
					{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "pending"}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
							},
						},
					},
					{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "active"}},
						Body: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
							},
						},
					},
				},
				Default: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.CompileEventHandler(handler)
	if err != nil {
		t.Fatalf("CompileEventHandler() error: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

// TestCompileUnaryOp tests unary operations
func TestCompileUnaryOp(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expr
		expected vm.Value
	}{
		{
			name: "negation",
			expr: &ast.UnaryOpExpr{
				Op:    ast.Neg,
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
			},
			expected: vm.IntValue{Val: -42},
		},
		{
			name: "not true",
			expr: &ast.UnaryOpExpr{
				Op:    ast.Not,
				Right: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
			},
			expected: vm.BoolValue{Val: false},
		},
		{
			name: "not false",
			expr: &ast.UnaryOpExpr{
				Op:    ast.Not,
				Right: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
			},
			expected: vm.BoolValue{Val: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			err := c.compileExpression(tt.expr)
			if err != nil {
				t.Fatalf("compileExpression() error: %v", err)
			}

			bytecode := c.buildBytecode()
			vmInstance := vm.NewVM()
			result, err := vmInstance.Execute(bytecode)
			if err != nil {
				t.Fatalf("Execute() error: %v", err)
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCompileValidationStatement tests validation statement compilation
func TestCompileValidationStatement(t *testing.T) {
	c := NewCompiler()

	stmt := &ast.ValidationStatement{
		Call: ast.FunctionCallExpr{
			Name: "validate",
			Args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
			},
		},
	}

	err := c.compileStatement(stmt)
	// ValidationStatement may be a no-op, just ensure no panic
	_ = err
}

// TestCompileExpressionStatement tests expression statement compilation
func TestCompileExpressionStatement(t *testing.T) {
	c := NewCompiler()

	stmt := &ast.ExpressionStatement{
		Expr: &ast.FunctionCallExpr{
			Name: "print",
			Args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "hello"}},
			},
		},
	}

	err := c.compileStatement(stmt)
	// May or may not return error depending on function
	_ = err
}

// TestSymbolTableFunctions tests symbol table methods
func TestSymbolTableFunctions(t *testing.T) {
	t.Run("DefineConstant", func(t *testing.T) {
		st := NewSymbolTable(nil, GlobalScope)
		st.DefineConstant("PI", 0, 1) // nameIndex, valueIndex

		sym, ok := st.Resolve("PI")
		if !ok {
			t.Error("Expected to resolve constant PI")
		}
		if !sym.IsConstant {
			t.Error("Expected PI to be a constant")
		}
	})

	t.Run("ResolveLocal", func(t *testing.T) {
		st := NewSymbolTable(nil, GlobalScope)
		st.Define("x", 0)

		sym, ok := st.ResolveLocal("x")
		if !ok {
			t.Error("Expected to resolve local x")
		}
		if sym.Name != "x" {
			t.Errorf("Expected name x, got %s", sym.Name)
		}

		// Should not resolve non-existent
		_, ok = st.ResolveLocal("nonexistent")
		if ok {
			t.Error("Should not resolve non-existent variable")
		}
	})

	t.Run("Symbols", func(t *testing.T) {
		st := NewSymbolTable(nil, GlobalScope)
		st.Define("a", 0)
		st.Define("b", 1)

		symbols := st.Symbols()
		if len(symbols) != 2 {
			t.Errorf("Expected 2 symbols, got %d", len(symbols))
		}
	})

	t.Run("Parent", func(t *testing.T) {
		parent := NewSymbolTable(nil, GlobalScope)
		child := NewSymbolTable(parent, FunctionScope)

		if child.Parent() != parent {
			t.Error("Parent() should return parent table")
		}
		if parent.Parent() != nil {
			t.Error("Root table should have nil parent")
		}
	})

	t.Run("Scope", func(t *testing.T) {
		st := NewSymbolTable(nil, GlobalScope)
		scope := st.Scope()
		if scope != GlobalScope {
			t.Errorf("Expected GlobalScope, got %v", scope)
		}
	})
}

// TestFunctionInliner tests function inliner
func TestFunctionInliner(t *testing.T) {
	inliner := NewFunctionInliner()
	if inliner == nil {
		t.Fatal("NewFunctionInliner returned nil")
	}

	// Define a simple function using ast.Function
	funcDef := ast.Function{
		Name:   "add",
		Params: []ast.Field{{Name: "a", TypeAnnotation: ast.IntType{}}, {Name: "b", TypeAnnotation: ast.IntType{}}},
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.BinaryOpExpr{
					Op:    ast.Add,
					Left:  &ast.VariableExpr{Name: "a"},
					Right: &ast.VariableExpr{Name: "b"},
				},
			},
		},
	}

	inliner.AnalyzeFunction(funcDef)

	// Test ShouldInline
	shouldInline := inliner.ShouldInline("add")
	// May or may not inline depending on size
	_ = shouldInline

	// Test InlineCall
	call := &ast.FunctionCallExpr{
		Name: "add",
		Args: []ast.Expr{
			&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
		},
	}

	inlined := inliner.InlineCall(call)
	// May return nil if not inlinable
	_ = inlined
}

// TestOptimizerStrengthReduce tests strength reduction optimization
func TestOptimizerStrengthReduce(t *testing.T) {
	opt := NewOptimizer(OptAggressive)

	tests := []struct {
		name string
		expr ast.Expr
	}{
		{
			name: "multiply by power of 2",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 8}},
			},
		},
		{
			name: "divide by power of 2",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Div,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := opt.StrengthReduce(tt.expr)
			// May return original or optimized expression
			if result == nil {
				t.Error("StrengthReduce should not return nil")
			}
		})
	}
}

// TestOptimizerApplyPeepholeOptimizations tests peephole optimization
func TestOptimizerApplyPeepholeOptimizations(t *testing.T) {
	opt := NewOptimizer(OptAggressive)

	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "x",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
		},
	}

	optimized := opt.ApplyPeepholeOptimizations(stmts)
	if optimized == nil {
		t.Error("ApplyPeepholeOptimizations should not return nil")
	}
}

// TestOptimizerGetStats tests getting optimizer stats
func TestOptimizerGetStats(t *testing.T) {
	opt := NewOptimizer(OptAggressive)

	// Run some optimization
	expr := &ast.BinaryOpExpr{
		Op:    ast.Add,
		Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
		Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
	}
	opt.OptimizeExpression(expr)

	stats := opt.GetStats()
	// Stats should have some data
	_ = stats
}

// =================================================================
// Match expression tests (compileMatchExpr, compilePatternMatch,
// compileLiteralValue)
// =================================================================

func TestCompileMatchExpr_LiteralPatterns(t *testing.T) {
	// match x { 1 => 10, 2 => 20, _ => 0 }
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)

	matchExpr := &interpreter.MatchExpr{
		Value: &interpreter.VariableExpr{Name: "x"},
		Cases: []interpreter.MatchCase{
			{
				Pattern: interpreter.LiteralPattern{
					Value: interpreter.IntLiteral{Value: 1},
				},
				Body: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
			},
			{
				Pattern: interpreter.LiteralPattern{
					Value: interpreter.IntLiteral{Value: 2},
				},
				Body: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 20}},
			},
			{
				Pattern: interpreter.WildcardPattern{},
				Body:    &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
		},
	}

	err := c.compileMatchExpr(matchExpr)
	if err != nil {
		t.Fatalf("compileMatchExpr() error: %v", err)
	}

	// Verify bytecode was generated (non-empty code)
	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from match expression")
	}
}

func TestCompileMatchExpr_VariablePattern(t *testing.T) {
	// match x { y => y + 1 }
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)

	matchExpr := &interpreter.MatchExpr{
		Value: &interpreter.VariableExpr{Name: "x"},
		Cases: []interpreter.MatchCase{
			{
				Pattern: interpreter.VariablePattern{Name: "y"},
				Body: &interpreter.BinaryOpExpr{
					Op:    interpreter.Add,
					Left:  &interpreter.VariableExpr{Name: "y"},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
				},
			},
		},
	}

	err := c.compileMatchExpr(matchExpr)
	if err != nil {
		t.Fatalf("compileMatchExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from match with variable pattern")
	}
}

func TestCompileMatchExpr_WildcardPattern(t *testing.T) {
	// match x { _ => 99 }
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)

	matchExpr := &interpreter.MatchExpr{
		Value: &interpreter.VariableExpr{Name: "x"},
		Cases: []interpreter.MatchCase{
			{
				Pattern: interpreter.WildcardPattern{},
				Body:    &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 99}},
			},
		},
	}

	err := c.compileMatchExpr(matchExpr)
	if err != nil {
		t.Fatalf("compileMatchExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from match with wildcard pattern")
	}
}

func TestCompileMatchExpr_ObjectPattern(t *testing.T) {
	// match x { {name} => name }
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)

	matchExpr := &interpreter.MatchExpr{
		Value: &interpreter.VariableExpr{Name: "x"},
		Cases: []interpreter.MatchCase{
			{
				Pattern: interpreter.ObjectPattern{
					Fields: []interpreter.ObjectPatternField{
						{Key: "name", Pattern: nil}, // bind field to variable
					},
				},
				Body: &interpreter.VariableExpr{Name: "name"},
			},
		},
	}

	err := c.compileMatchExpr(matchExpr)
	if err != nil {
		t.Fatalf("compileMatchExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from match with object pattern")
	}
}

func TestCompileMatchExpr_ObjectPatternWithNestedLiteral(t *testing.T) {
	// match x { {status: 200} => "ok" }
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)

	matchExpr := &interpreter.MatchExpr{
		Value: &interpreter.VariableExpr{Name: "x"},
		Cases: []interpreter.MatchCase{
			{
				Pattern: interpreter.ObjectPattern{
					Fields: []interpreter.ObjectPatternField{
						{
							Key: "status",
							Pattern: interpreter.LiteralPattern{
								Value: interpreter.IntLiteral{Value: 200},
							},
						},
					},
				},
				Body: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "ok"}},
			},
		},
	}

	err := c.compileMatchExpr(matchExpr)
	if err != nil {
		t.Fatalf("compileMatchExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from match with nested object pattern")
	}
}

func TestCompileMatchExpr_ArrayPattern(t *testing.T) {
	// match x { [first, second] => first }
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)

	matchExpr := &interpreter.MatchExpr{
		Value: &interpreter.VariableExpr{Name: "x"},
		Cases: []interpreter.MatchCase{
			{
				Pattern: interpreter.ArrayPattern{
					Elements: []interpreter.Pattern{
						interpreter.VariablePattern{Name: "first"},
						interpreter.VariablePattern{Name: "second"},
					},
				},
				Body: &interpreter.VariableExpr{Name: "first"},
			},
		},
	}

	err := c.compileMatchExpr(matchExpr)
	if err != nil {
		t.Fatalf("compileMatchExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from match with array pattern")
	}
}

func TestCompileMatchExpr_ArrayPatternWithRest(t *testing.T) {
	// match x { [head, ...rest] => head }
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)

	restName := "rest"
	matchExpr := &interpreter.MatchExpr{
		Value: &interpreter.VariableExpr{Name: "x"},
		Cases: []interpreter.MatchCase{
			{
				Pattern: interpreter.ArrayPattern{
					Elements: []interpreter.Pattern{
						interpreter.VariablePattern{Name: "head"},
					},
					Rest: &restName,
				},
				Body: &interpreter.VariableExpr{Name: "head"},
			},
		},
	}

	err := c.compileMatchExpr(matchExpr)
	if err != nil {
		t.Fatalf("compileMatchExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from match with rest pattern")
	}
}

func TestCompileMatchExpr_WithGuard(t *testing.T) {
	// match x { y when y > 0 => y, _ => 0 }
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)

	matchExpr := &interpreter.MatchExpr{
		Value: &interpreter.VariableExpr{Name: "x"},
		Cases: []interpreter.MatchCase{
			{
				Pattern: interpreter.VariablePattern{Name: "y"},
				Guard: &interpreter.BinaryOpExpr{
					Op:    interpreter.Gt,
					Left:  &interpreter.VariableExpr{Name: "y"},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
				},
				Body: &interpreter.VariableExpr{Name: "y"},
			},
			{
				Pattern: interpreter.WildcardPattern{},
				Body:    &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
		},
	}

	err := c.compileMatchExpr(matchExpr)
	if err != nil {
		t.Fatalf("compileMatchExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from match with guard")
	}
}

func TestCompileMatchExpr_ValueTypes(t *testing.T) {
	// Test MatchExpr as value type in compileExpression
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)

	// Pointer type
	matchExpr := &interpreter.MatchExpr{
		Value: &interpreter.VariableExpr{Name: "x"},
		Cases: []interpreter.MatchCase{
			{
				Pattern: interpreter.WildcardPattern{},
				Body:    &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
			},
		},
	}
	err := c.compileExpression(matchExpr)
	if err != nil {
		t.Fatalf("compileExpression(*MatchExpr) error: %v", err)
	}

	// Value type
	c.code = make([]byte, 0)
	matchExprVal := interpreter.MatchExpr{
		Value: &interpreter.VariableExpr{Name: "x"},
		Cases: []interpreter.MatchCase{
			{
				Pattern: interpreter.WildcardPattern{},
				Body:    &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
			},
		},
	}
	err = c.compileExpression(matchExprVal)
	if err != nil {
		t.Fatalf("compileExpression(MatchExpr) error: %v", err)
	}
}

func TestCompileMatchExpr_MultipleLiteralPatterns(t *testing.T) {
	// match x { 1 => "one", 2 => "two", 3 => "three" }
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)

	matchExpr := &interpreter.MatchExpr{
		Value: &interpreter.VariableExpr{Name: "x"},
		Cases: []interpreter.MatchCase{
			{
				Pattern: interpreter.LiteralPattern{Value: interpreter.IntLiteral{Value: 1}},
				Body:    &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "one"}},
			},
			{
				Pattern: interpreter.LiteralPattern{Value: interpreter.IntLiteral{Value: 2}},
				Body:    &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "two"}},
			},
			{
				Pattern: interpreter.LiteralPattern{Value: interpreter.IntLiteral{Value: 3}},
				Body:    &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "three"}},
			},
		},
	}

	err := c.compileMatchExpr(matchExpr)
	if err != nil {
		t.Fatalf("compileMatchExpr() error: %v", err)
	}
}

// Test compileLiteralValue with all literal types
func TestCompileLiteralValue_AllTypes(t *testing.T) {
	tests := []struct {
		name string
		lit  interpreter.Literal
	}{
		{"int", interpreter.IntLiteral{Value: 42}},
		{"float", interpreter.FloatLiteral{Value: 3.14}},
		{"string", interpreter.StringLiteral{Value: "hello"}},
		{"bool true", interpreter.BoolLiteral{Value: true}},
		{"bool false", interpreter.BoolLiteral{Value: false}},
		{"null", interpreter.NullLiteral{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			err := c.compileLiteralValue(tt.lit)
			if err != nil {
				t.Fatalf("compileLiteralValue(%s) error: %v", tt.name, err)
			}
			if len(c.code) == 0 {
				t.Error("expected non-empty bytecode")
			}
		})
	}
}

// =================================================================
// Async/Await expression tests (compileAsyncExpr, compileAwaitExpr)
// =================================================================

func TestCompileAsyncExpr_Basic(t *testing.T) {
	// async { > 42 }
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	asyncExpr := &interpreter.AsyncExpr{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
			},
		},
	}

	err := c.compileAsyncExpr(asyncExpr)
	if err != nil {
		t.Fatalf("compileAsyncExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from async expression")
	}
}

func TestCompileAsyncExpr_WithAssignment(t *testing.T) {
	// async { $ x = 10, > x }
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	asyncExpr := &interpreter.AsyncExpr{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "x",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "x"},
			},
		},
	}

	err := c.compileAsyncExpr(asyncExpr)
	if err != nil {
		t.Fatalf("compileAsyncExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from async expression with assignment")
	}
}

func TestCompileAsyncExpr_EmptyBody(t *testing.T) {
	// async { }
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	asyncExpr := &interpreter.AsyncExpr{
		Body: []interpreter.Statement{},
	}

	err := c.compileAsyncExpr(asyncExpr)
	if err != nil {
		t.Fatalf("compileAsyncExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from empty async expression")
	}
}

func TestCompileAsyncExpr_NoExplicitReturn(t *testing.T) {
	// async { $ x = 42 } -- no explicit return, should add halt
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	asyncExpr := &interpreter.AsyncExpr{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "x",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
			},
		},
	}

	err := c.compileAsyncExpr(asyncExpr)
	if err != nil {
		t.Fatalf("compileAsyncExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode")
	}
}

func TestCompileAsyncExpr_ValueType(t *testing.T) {
	// Test AsyncExpr as value type in compileExpression
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	// Pointer type
	asyncPtrExpr := &interpreter.AsyncExpr{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
			},
		},
	}
	err := c.compileExpression(asyncPtrExpr)
	if err != nil {
		t.Fatalf("compileExpression(*AsyncExpr) error: %v", err)
	}

	// Value type
	c.code = make([]byte, 0)
	asyncValExpr := interpreter.AsyncExpr{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
			},
		},
	}
	err = c.compileExpression(asyncValExpr)
	if err != nil {
		t.Fatalf("compileExpression(AsyncExpr) error: %v", err)
	}
}

func TestCompileAwaitExpr_Basic(t *testing.T) {
	// await someExpr
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	futIdx := c.addConstant(vm.StringValue{Val: "future"})
	c.symbolTable.Define("future", futIdx)

	awaitExpr := &interpreter.AwaitExpr{
		Expr: &interpreter.VariableExpr{Name: "future"},
	}

	err := c.compileAwaitExpr(awaitExpr)
	if err != nil {
		t.Fatalf("compileAwaitExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from await expression")
	}
}

func TestCompileAwaitExpr_WithLiteral(t *testing.T) {
	// await 42
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	awaitExpr := &interpreter.AwaitExpr{
		Expr: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
	}

	err := c.compileAwaitExpr(awaitExpr)
	if err != nil {
		t.Fatalf("compileAwaitExpr() error: %v", err)
	}

	if len(c.code) == 0 {
		t.Error("expected non-empty bytecode from await expression")
	}
}

func TestCompileAwaitExpr_ValueType(t *testing.T) {
	// Test AwaitExpr as value type in compileExpression
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	// Pointer type
	awaitPtrExpr := &interpreter.AwaitExpr{
		Expr: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
	}
	err := c.compileExpression(awaitPtrExpr)
	if err != nil {
		t.Fatalf("compileExpression(*AwaitExpr) error: %v", err)
	}

	// Value type
	c.code = make([]byte, 0)
	awaitValExpr := interpreter.AwaitExpr{
		Expr: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
	}
	err = c.compileExpression(awaitValExpr)
	if err != nil {
		t.Fatalf("compileExpression(AwaitExpr) error: %v", err)
	}
}

// =================================================================
// Additional compileExpression coverage
// =================================================================

func TestCompileExpression_UnsupportedType(t *testing.T) {
	c := NewCompiler()
	// Pass a nil expression which is an unsupported type
	err := c.compileExpression(nil)
	if err == nil {
		t.Fatal("expected error for nil expression, got nil")
	}
}

func TestCompileStatement_UnsupportedType(t *testing.T) {
	c := NewCompiler()
	// Pass a nil statement which is an unsupported type
	err := c.compileStatement(nil)
	if err == nil {
		t.Fatal("expected error for nil statement, got nil")
	}
}

func TestCompileUnaryOp_Not(t *testing.T) {
	// Test: !true => false
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.UnaryOpExpr{
					Op:    interpreter.Not,
					Right: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
				},
			},
		},
	}

	c := NewCompilerWithOptLevel(OptNone)
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.BoolValue{Val: false}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileUnaryOp_Neg(t *testing.T) {
	// Test: -42 => -42
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.UnaryOpExpr{
					Op:    interpreter.Neg,
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
				},
			},
		},
	}

	c := NewCompilerWithOptLevel(OptNone)
	bytecode, err := c.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: -42}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileUnaryOp_UnsupportedOp(t *testing.T) {
	c := NewCompiler()
	// Use an invalid unary op value
	expr := &interpreter.UnaryOpExpr{
		Op:    interpreter.UnOp(99),
		Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
	}
	err := c.compileUnaryOp(expr)
	if err == nil {
		t.Fatal("expected error for unsupported unary operator, got nil")
	}
}

func TestCompileBinaryOp_UnsupportedOp(t *testing.T) {
	c := NewCompiler()
	// Use an invalid binary op value
	expr := &interpreter.BinaryOpExpr{
		Op:    interpreter.BinOp(99),
		Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
		Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
	}
	err := c.compileBinaryOp(expr)
	if err == nil {
		t.Fatal("expected error for unsupported binary operator, got nil")
	}
}

func TestCompileReturnStatement_WithExpression(t *testing.T) {
	// return (x + y)
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)
	yIdx := c.addConstant(vm.StringValue{Val: "y"})
	c.symbolTable.Define("y", yIdx)

	stmt := &interpreter.ReturnStatement{
		Value: &interpreter.BinaryOpExpr{
			Op:    interpreter.Add,
			Left:  &interpreter.VariableExpr{Name: "x"},
			Right: &interpreter.VariableExpr{Name: "y"},
		},
	}

	err := c.compileReturnStatement(stmt)
	if err != nil {
		t.Fatalf("compileReturnStatement() error: %v", err)
	}

	// Verify code has OpReturn
	foundReturn := false
	for _, b := range c.code {
		if b == byte(vm.OpReturn) {
			foundReturn = true
			break
		}
	}
	if !foundReturn {
		t.Error("expected OpReturn in bytecode")
	}
}

func TestCompileReturnStatement_ErrorInExpression(t *testing.T) {
	c := NewCompiler()
	// Reference undefined variable
	stmt := &interpreter.ReturnStatement{
		Value: &interpreter.VariableExpr{Name: "undefined_var"},
	}

	err := c.compileReturnStatement(stmt)
	if err == nil {
		t.Fatal("expected error for undefined variable in return, got nil")
	}
}

func TestCompileArrayIndex_ErrorInArray(t *testing.T) {
	c := NewCompiler()
	// Array expression references undefined variable
	expr := &interpreter.ArrayIndexExpr{
		Array: &interpreter.VariableExpr{Name: "undefined_arr"},
		Index: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
	}

	err := c.compileArrayIndex(expr)
	if err == nil {
		t.Fatal("expected error for undefined array, got nil")
	}
}

func TestCompileArrayIndex_ErrorInIndex(t *testing.T) {
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	arrIdx := c.addConstant(vm.StringValue{Val: "arr"})
	c.symbolTable.Define("arr", arrIdx)

	// Index references undefined variable
	expr := &interpreter.ArrayIndexExpr{
		Array: &interpreter.VariableExpr{Name: "arr"},
		Index: &interpreter.VariableExpr{Name: "undefined_idx"},
	}

	err := c.compileArrayIndex(expr)
	if err == nil {
		t.Fatal("expected error for undefined index, got nil")
	}
}

// =================================================================
// Additional compileStatement coverage
// =================================================================

func TestCompileStatement_SwitchValueType(t *testing.T) {
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)

	// SwitchStatement as value type
	switchStmt := interpreter.SwitchStatement{
		Value: &interpreter.VariableExpr{Name: "x"},
		Cases: []interpreter.SwitchCase{
			{
				Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
				Body: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "one"}},
					},
				},
			},
		},
		Default: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "other"}},
			},
		},
	}

	err := c.compileStatement(switchStmt)
	if err != nil {
		t.Fatalf("compileStatement(SwitchStatement value) failed: %v", err)
	}
}

func TestCompileStatement_ForValueType(t *testing.T) {
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	itemsIdx := c.addConstant(vm.StringValue{Val: "items"})
	c.symbolTable.Define("items", itemsIdx)

	// ForStatement as value type
	forStmt := interpreter.ForStatement{
		ValueVar: "item",
		Iterable: &interpreter.VariableExpr{Name: "items"},
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "x",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
			},
		},
	}

	err := c.compileStatement(forStmt)
	if err != nil {
		t.Fatalf("compileStatement(ForStatement value) failed: %v", err)
	}
}

func TestCompileStatement_ReassignValueType(t *testing.T) {
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)
	xIdx := c.addConstant(vm.StringValue{Val: "x"})
	c.symbolTable.Define("x", xIdx)

	// ReassignStatement as value type
	reassignStmt := interpreter.ReassignStatement{
		Target: "x",
		Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
	}

	err := c.compileStatement(reassignStmt)
	if err != nil {
		t.Fatalf("compileStatement(ReassignStatement value) failed: %v", err)
	}
}

func TestCompileStatement_ValidationValueType(t *testing.T) {
	c := NewCompiler()

	// ValidationStatement as value type
	validationStmt := interpreter.ValidationStatement{}

	err := c.compileStatement(validationStmt)
	if err != nil {
		t.Fatalf("compileStatement(ValidationStatement value) failed: %v", err)
	}
}

func TestCompileStatement_ValidationPtrType(t *testing.T) {
	c := NewCompiler()

	// ValidationStatement as pointer type
	validationStmt := &interpreter.ValidationStatement{}

	err := c.compileStatement(validationStmt)
	if err != nil {
		t.Fatalf("compileStatement(*ValidationStatement) failed: %v", err)
	}
}
