package compiler

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/vm"
)

func TestCompileLiteral(t *testing.T) {
	tests := []struct {
		name     string
		expr     *interpreter.LiteralExpr
		expected vm.Value
	}{
		{
			name: "int literal",
			expr: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
			expected: vm.IntValue{Val: 42},
		},
		{
			name: "float literal",
			expr: &interpreter.LiteralExpr{Value: interpreter.FloatLiteral{Value: 3.14}},
			expected: vm.FloatValue{Val: 3.14},
		},
		{
			name: "string literal",
			expr: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "hello"}},
			expected: vm.StringValue{Val: "hello"},
		},
		{
			name: "bool literal true",
			expr: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
			expected: vm.BoolValue{Val: true},
		},
		{
			name: "bool literal false",
			expr: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}},
			expected: vm.BoolValue{Val: false},
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
		expr     *interpreter.BinaryOpExpr
		expected vm.Value
	}{
		{
			name: "5 + 3",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
			},
			expected: vm.IntValue{Val: 8},
		},
		{
			name: "10 - 4",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Sub,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 4}},
			},
			expected: vm.IntValue{Val: 6},
		},
		{
			name: "6 * 7",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 6}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 7}},
			},
			expected: vm.IntValue{Val: 42},
		},
		{
			name: "42 == 42",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Eq,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
			},
			expected: vm.BoolValue{Val: true},
		},
		{
			name: "5 > 3",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Gt,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "x",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "x"},
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

func TestCompileArithmeticExpression(t *testing.T) {
	// Test: $ result = 5 + 3 * 2, > result
	// Expected: 11 (5 + (3 * 2))
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "result",
				Value: &interpreter.BinaryOpExpr{
					Op: interpreter.Add,
					Left: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
					Right: &interpreter.BinaryOpExpr{
						Op:    interpreter.Mul,
						Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
						Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "result"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.ObjectExpr{
					Fields: []interpreter.ObjectField{
						{
							Key:   "name",
							Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "Alice"}},
						},
						{
							Key:   "age",
							Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 30}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "obj",
				Value: &interpreter.ObjectExpr{
					Fields: []interpreter.ObjectField{
						{
							Key:   "name",
							Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "Alice"}},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.FieldAccessExpr{
					Object: &interpreter.VariableExpr{Name: "obj"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.IfStatement{
				Condition: &interpreter.BinaryOpExpr{
					Op:    interpreter.Gt,
					Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
				},
				ThenBlock: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
					},
				},
				ElseBlock: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "x",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.WhileStatement{
				Condition: &interpreter.BinaryOpExpr{
					Op:    interpreter.Lt,
					Left:  &interpreter.VariableExpr{Name: "x"},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "x",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "x"},
							Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "x"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "sum",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.AssignStatement{
				Target: "i",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
			},
			&interpreter.WhileStatement{
				Condition: &interpreter.BinaryOpExpr{
					Op:    interpreter.Le,
					Left:  &interpreter.VariableExpr{Name: "i"},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "sum",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "sum"},
							Right: &interpreter.VariableExpr{Name: "i"},
						},
					},
					&interpreter.AssignStatement{
						Target: "i",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "i"},
							Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "sum"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "x",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
			},
			&interpreter.WhileStatement{
				Condition: &interpreter.BinaryOpExpr{
					Op:    interpreter.Lt,
					Left:  &interpreter.VariableExpr{Name: "x"},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "x",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "x"},
							Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "x"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "s",
				Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: ""}},
			},
			&interpreter.AssignStatement{
				Target: "i",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.WhileStatement{
				Condition: &interpreter.BinaryOpExpr{
					Op:    interpreter.Lt,
					Left:  &interpreter.VariableExpr{Name: "i"},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "s",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "s"},
							Right: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "x"}},
						},
					},
					&interpreter.AssignStatement{
						Target: "i",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "i"},
							Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "s"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "count",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.AssignStatement{
				Target: "i",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.WhileStatement{
				Condition: &interpreter.BinaryOpExpr{
					Op:    interpreter.Lt,
					Left:  &interpreter.VariableExpr{Name: "i"},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
				},
				Body: []interpreter.Statement{
					&interpreter.IfStatement{
						Condition: &interpreter.BinaryOpExpr{
							Op:    interpreter.Gt,
							Left:  &interpreter.VariableExpr{Name: "i"},
							Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 4}},
						},
						ThenBlock: []interpreter.Statement{
							&interpreter.AssignStatement{
								Target: "count",
								Value: &interpreter.BinaryOpExpr{
									Op:    interpreter.Add,
									Left:  &interpreter.VariableExpr{Name: "count"},
									Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
								},
							},
						},
					},
					&interpreter.AssignStatement{
						Target: "i",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "i"},
							Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "count"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "result",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
			},
			&interpreter.AssignStatement{
				Target: "i",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.WhileStatement{
				Condition: &interpreter.BinaryOpExpr{
					Op:    interpreter.Lt,
					Left:  &interpreter.VariableExpr{Name: "i"},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 4}},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "result",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Mul,
							Left:  &interpreter.VariableExpr{Name: "result"},
							Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
						},
					},
					&interpreter.AssignStatement{
						Target: "i",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "i"},
							Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "result"},
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
	route := &interpreter.Route{
		Path:   "/users/:id",
		Method: interpreter.Get,
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "id"},
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
	route := &interpreter.Route{
		Path:   "/users/:userId/posts/:postId",
		Method: interpreter.Get,
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.BinaryOpExpr{
					Op:    interpreter.Add,
					Left:  &interpreter.VariableExpr{Name: "userId"},
					Right: &interpreter.VariableExpr{Name: "postId"},
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
	route := &interpreter.Route{
		Path:   "/test",
		Method: interpreter.Get,
		Injections: []interpreter.Injection{
			{
				Name: "db",
				Type: interpreter.DatabaseType{},
			},
		},
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "db"},
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
	route := &interpreter.Route{
		Path:   "/users/:id",
		Method: interpreter.Get,
		Injections: []interpreter.Injection{
			{
				Name: "db",
				Type: interpreter.DatabaseType{},
			},
		},
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.ObjectExpr{
					Fields: []interpreter.ObjectField{
						{
							Key:   "id",
							Value: &interpreter.VariableExpr{Name: "id"},
						},
						{
							Key:   "hasDb",
							Value: &interpreter.VariableExpr{Name: "db"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "result",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.ForStatement{
				ValueVar: "item",
				Iterable: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
					},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "result",
						Value:  &interpreter.VariableExpr{Name: "item"},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "result"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "sum",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.ForStatement{
				ValueVar: "item",
				Iterable: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
					},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "sum",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "sum"},
							Right: &interpreter.VariableExpr{Name: "item"},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "sum"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "result",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.ForStatement{
				KeyVar:   "idx",
				ValueVar: "val",
				Iterable: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 20}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 30}},
					},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "result",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "result"},
							Right: &interpreter.VariableExpr{Name: "idx"},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "result"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "result",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.ForStatement{
				KeyVar:   "idx",
				ValueVar: "val",
				Iterable: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 20}},
					},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "result",
						Value: &interpreter.BinaryOpExpr{
							Op: interpreter.Add,
							Left: &interpreter.BinaryOpExpr{
								Op:    interpreter.Add,
								Left:  &interpreter.VariableExpr{Name: "result"},
								Right: &interpreter.VariableExpr{Name: "idx"},
							},
							Right: &interpreter.VariableExpr{Name: "val"},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "result"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "arr",
				Value:  &interpreter.ArrayExpr{Elements: []interpreter.Expr{}},
			},
			&interpreter.ForStatement{
				KeyVar:   "idx",
				ValueVar: "val",
				Iterable: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 20}},
					},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "entry",
						Value: &interpreter.ObjectExpr{
							Fields: []interpreter.ObjectField{
								{Key: "pos", Value: &interpreter.VariableExpr{Name: "idx"}},
								{Key: "val", Value: &interpreter.VariableExpr{Name: "val"}},
							},
						},
					},
					&interpreter.AssignStatement{
						Target: "arr",
						Value: &interpreter.BinaryOpExpr{
							Op:   interpreter.Add,
							Left: &interpreter.VariableExpr{Name: "arr"},
							Right: &interpreter.ArrayExpr{
								Elements: []interpreter.Expr{
									&interpreter.VariableExpr{Name: "entry"},
								},
							},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "arr"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "sum",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
			},
			&interpreter.ForStatement{
				ValueVar: "item",
				Iterable: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "sum",
						Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "sum"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "result",
				Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: ""}},
			},
			&interpreter.ForStatement{
				ValueVar: "s",
				Iterable: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{
						&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "a"}},
						&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "b"}},
						&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "c"}},
					},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "result",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "result"},
							Right: &interpreter.VariableExpr{Name: "s"},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "result"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "arr",
				Value: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{},
				},
			},
			&interpreter.ForStatement{
				ValueVar: "x",
				Iterable: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
					},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "arr",
						Value: &interpreter.BinaryOpExpr{
							Op:   interpreter.Add,
							Left: &interpreter.VariableExpr{Name: "arr"},
							Right: &interpreter.ArrayExpr{
								Elements: []interpreter.Expr{
									&interpreter.VariableExpr{Name: "x"},
								},
							},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "arr"},
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

// Switch statement tests

func TestCompileSwitchSimpleStringMatch(t *testing.T) {
	// Test: $ status = "pending", switch status { case "pending" { > "matched" } default { > "no match" } }
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "status",
				Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "pending"}},
			},
			&interpreter.SwitchStatement{
				Value: &interpreter.VariableExpr{Name: "status"},
				Cases: []interpreter.SwitchCase{
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "pending"}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "matched"}},
							},
						},
					},
				},
				Default: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "no match"}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.SwitchStatement{
				Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "shipped"}},
				Cases: []interpreter.SwitchCase{
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "pending"}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
							},
						},
					},
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "shipped"}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
							},
						},
					},
				},
				Default: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.SwitchStatement{
				Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "unknown"}},
				Cases: []interpreter.SwitchCase{
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "a"}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
							},
						},
					},
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "b"}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
							},
						},
					},
				},
				Default: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 99}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.SwitchStatement{
				Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
				Cases: []interpreter.SwitchCase{
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "one"}},
							},
						},
					},
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "forty-two"}},
							},
						},
					},
				},
				Default: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "other"}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "result",
				Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: ""}},
			},
			&interpreter.AssignStatement{
				Target: "n",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
			},
			&interpreter.SwitchStatement{
				Value: &interpreter.VariableExpr{Name: "n"},
				Cases: []interpreter.SwitchCase{
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						Body: []interpreter.Statement{
							&interpreter.AssignStatement{
								Target: "result",
								Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "one"}},
							},
						},
					},
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
						Body: []interpreter.Statement{
							&interpreter.AssignStatement{
								Target: "result",
								Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "two"}},
							},
						},
					},
				},
				Default: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "result",
						Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "default"}},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "result"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "result",
				Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "unchanged"}},
			},
			&interpreter.SwitchStatement{
				Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "x"}},
				Cases: []interpreter.SwitchCase{
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "a"}},
						Body: []interpreter.Statement{
							&interpreter.AssignStatement{
								Target: "result",
								Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "a"}},
							},
						},
					},
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "b"}},
						Body: []interpreter.Statement{
							&interpreter.AssignStatement{
								Target: "result",
								Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "b"}},
							},
						},
					},
				},
				Default: nil, // No default case
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "result"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.SwitchStatement{
				Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "ok"}},
				Cases: []interpreter.SwitchCase{
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "ok"}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.ObjectExpr{
									Fields: []interpreter.ObjectField{
										{
											Key:   "status",
											Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "success"}},
										},
										{
											Key:   "code",
											Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 200}},
										},
									},
								},
							},
						},
					},
				},
				Default: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.ObjectExpr{
							Fields: []interpreter.ObjectField{
								{
									Key:   "status",
									Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "error"}},
								},
								{
									Key:   "code",
									Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 500}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.SwitchStatement{
				Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
				Cases: []interpreter.SwitchCase{
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "one"}},
							},
						},
					},
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "two"}},
							},
						},
					},
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "three"}},
							},
						},
					},
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 4}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "four"}},
							},
						},
					},
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "five"}},
							},
						},
					},
				},
				Default: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "other"}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.FunctionCallExpr{
					Name: "now",
					Args: []interpreter.Expr{},
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

	// now() returns a mock timestamp
	expected := vm.IntValue{Val: 1234567890}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileFunctionCallWithOneArg(t *testing.T) {
	// Test: > length([1, 2, 3])
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.FunctionCallExpr{
					Name: "length",
					Args: []interpreter.Expr{
						&interpreter.ArrayExpr{
							Elements: []interpreter.Expr{
								&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
								&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
								&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.FunctionCallExpr{
					Name: "length",
					Args: []interpreter.Expr{
						&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "hello"}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.BinaryOpExpr{
					Op: interpreter.Add,
					Left: &interpreter.FunctionCallExpr{
						Name: "length",
						Args: []interpreter.Expr{
							&interpreter.ArrayExpr{
								Elements: []interpreter.Expr{
									&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
									&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
									&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
									&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 4}},
									&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
								},
							},
						},
					},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "arr",
				Value: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 4}},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.FunctionCallExpr{
					Name: "length",
					Args: []interpreter.Expr{
						&interpreter.VariableExpr{Name: "arr"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.FunctionCallExpr{
					Name: "time.now",
					Args: []interpreter.Expr{},
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

	// time.now() returns a mock timestamp
	expected := vm.IntValue{Val: 1234567890}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCompileFunctionCallInObject(t *testing.T) {
	// Test: > {timestamp: now(), items: length([1, 2])}
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.ObjectExpr{
					Fields: []interpreter.ObjectField{
						{
							Key: "timestamp",
							Value: &interpreter.FunctionCallExpr{
								Name: "now",
								Args: []interpreter.Expr{},
							},
						},
						{
							Key: "items",
							Value: &interpreter.FunctionCallExpr{
								Name: "length",
								Args: []interpreter.Expr{
									&interpreter.ArrayExpr{
										Elements: []interpreter.Expr{
											&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
											&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
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

	if tsVal, ok := objVal.Val["timestamp"].(vm.IntValue); !ok || tsVal.Val != 1234567890 {
		t.Errorf("Expected timestamp=1234567890, got %v", objVal.Val["timestamp"])
	}

	if itemsVal, ok := objVal.Val["items"].(vm.IntValue); !ok || itemsVal.Val != 2 {
		t.Errorf("Expected items=2, got %v", objVal.Val["items"])
	}
}

func TestCompileFunctionCallInCondition(t *testing.T) {
	// Test: $ arr = [1, 2, 3], if length(arr) > 2 { > "big" } else { > "small" }
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "arr",
				Value: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
					},
				},
			},
			&interpreter.IfStatement{
				Condition: &interpreter.BinaryOpExpr{
					Op: interpreter.Gt,
					Left: &interpreter.FunctionCallExpr{
						Name: "length",
						Args: []interpreter.Expr{
							&interpreter.VariableExpr{Name: "arr"},
						},
					},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
				},
				ThenBlock: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "big"}},
					},
				},
				ElseBlock: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "small"}},
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
	cmd := &interpreter.Command{
		Name: "greet",
		Params: []interpreter.CommandParam{
			{Name: "name", Type: interpreter.StringType{}, Required: true},
		},
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.BinaryOpExpr{
					Op:    interpreter.Add,
					Left:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "Hello, "}},
					Right: &interpreter.VariableExpr{Name: "name"},
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
	cmd := &interpreter.Command{
		Name: "add",
		Params: []interpreter.CommandParam{
			{Name: "x", Type: interpreter.IntType{}, Required: true},
			{Name: "y", Type: interpreter.IntType{}, Required: true},
		},
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.BinaryOpExpr{
					Op:    interpreter.Add,
					Left:  &interpreter.VariableExpr{Name: "x"},
					Right: &interpreter.VariableExpr{Name: "y"},
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
	cmd := &interpreter.Command{
		Name: "process",
		Params: []interpreter.CommandParam{
			{Name: "input", Type: interpreter.StringType{}, Required: true},
		},
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "result",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.AssignStatement{
				Target: "i",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.WhileStatement{
				Condition: &interpreter.BinaryOpExpr{
					Op:    interpreter.Lt,
					Left:  &interpreter.VariableExpr{Name: "i"},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "result",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "result"},
							Right: &interpreter.VariableExpr{Name: "i"},
						},
					},
					&interpreter.AssignStatement{
						Target: "i",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "i"},
							Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "result"},
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
	task := &interpreter.CronTask{
		Name:     "daily_cleanup",
		Schedule: "0 0 * * *",
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "count",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.ObjectExpr{
					Fields: []interpreter.ObjectField{
						{Key: "deleted", Value: &interpreter.VariableExpr{Name: "count"}},
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
	task := &interpreter.CronTask{
		Name:     "backup",
		Schedule: "0 0 * * *",
		Injections: []interpreter.Injection{
			{Name: "db", Type: interpreter.DatabaseType{}},
		},
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "count",
				Value: &interpreter.FieldAccessExpr{
					Object: &interpreter.VariableExpr{Name: "db"},
					Field:  "count",
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "count"},
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
	task := &interpreter.CronTask{
		Name:     "hourly_sync",
		Schedule: "0 */1 * * *",
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "total",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.AssignStatement{
				Target: "i",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.WhileStatement{
				Condition: &interpreter.BinaryOpExpr{
					Op:    interpreter.Lt,
					Left:  &interpreter.VariableExpr{Name: "i"},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
				},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "total",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "total"},
							Right: &interpreter.VariableExpr{Name: "i"},
						},
					},
					&interpreter.AssignStatement{
						Target: "i",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "i"},
							Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "total"},
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
	handler := &interpreter.EventHandler{
		EventType: "user.created",
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "userId",
				Value: &interpreter.FieldAccessExpr{
					Object: &interpreter.VariableExpr{Name: "event"},
					Field:  "userId",
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.ObjectExpr{
					Fields: []interpreter.ObjectField{
						{Key: "handled", Value: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}}},
						{Key: "userId", Value: &interpreter.VariableExpr{Name: "userId"}},
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
	handler := &interpreter.EventHandler{
		EventType: "order.paid",
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "orderId",
				Value: &interpreter.FieldAccessExpr{
					Object: &interpreter.VariableExpr{Name: "input"},
					Field:  "orderId",
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "orderId"},
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
	handler := &interpreter.EventHandler{
		EventType: "notification.send",
		Injections: []interpreter.Injection{
			{Name: "db", Type: interpreter.DatabaseType{}},
		},
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "user",
				Value: &interpreter.FieldAccessExpr{
					Object: &interpreter.VariableExpr{Name: "db"},
					Field:  "user",
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "user"},
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
	handler := &interpreter.EventHandler{
		EventType: "data.process",
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "data",
				Value: &interpreter.FieldAccessExpr{
					Object: &interpreter.VariableExpr{Name: "event"},
					Field:  "data",
				},
			},
			&interpreter.IfStatement{
				Condition: &interpreter.BinaryOpExpr{
					Op:    interpreter.Ne,
					Left:  &interpreter.VariableExpr{Name: "data"},
					Right: &interpreter.LiteralExpr{Value: interpreter.NullLiteral{}},
				},
				ThenBlock: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
					},
				},
				ElseBlock: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}},
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
	worker := &interpreter.QueueWorker{
		QueueName: "email.send",
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "to",
				Value: &interpreter.FieldAccessExpr{
					Object: &interpreter.VariableExpr{Name: "message"},
					Field:  "to",
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.ObjectExpr{
					Fields: []interpreter.ObjectField{
						{Key: "sent", Value: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}}},
						{Key: "to", Value: &interpreter.VariableExpr{Name: "to"}},
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
	worker := &interpreter.QueueWorker{
		QueueName: "image.resize",
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "imageId",
				Value: &interpreter.FieldAccessExpr{
					Object: &interpreter.VariableExpr{Name: "input"},
					Field:  "imageId",
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "imageId"},
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
	worker := &interpreter.QueueWorker{
		QueueName: "report.generate",
		Injections: []interpreter.Injection{
			{Name: "db", Type: interpreter.DatabaseType{}},
		},
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "data",
				Value: &interpreter.FieldAccessExpr{
					Object: &interpreter.VariableExpr{Name: "db"},
					Field:  "data",
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "data"},
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
	worker := &interpreter.QueueWorker{
		QueueName: "data.process",
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "items",
				Value: &interpreter.FieldAccessExpr{
					Object: &interpreter.VariableExpr{Name: "message"},
					Field:  "items",
				},
			},
			&interpreter.AssignStatement{
				Target: "count",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.ForStatement{
				ValueVar: "item",
				Iterable: &interpreter.VariableExpr{Name: "items"},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "count",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "count"},
							Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "count"},
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
	cmd := &interpreter.Command{
		Name: "check",
		Params: []interpreter.CommandParam{
			{Name: "value", Type: interpreter.IntType{}, Required: true},
		},
		Body: []interpreter.Statement{
			&interpreter.IfStatement{
				Condition: &interpreter.BinaryOpExpr{
					Op:    interpreter.Gt,
					Left:  &interpreter.VariableExpr{Name: "value"},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
				},
				ThenBlock: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "positive"}},
					},
				},
				ElseBlock: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "negative"}},
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
	task := &interpreter.CronTask{
		Name:     "process_items",
		Schedule: "0 * * * *",
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "items",
				Value: &interpreter.ArrayExpr{
					Elements: []interpreter.Expr{
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
						&interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
					},
				},
			},
			&interpreter.AssignStatement{
				Target: "sum",
				Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			&interpreter.ForStatement{
				ValueVar: "item",
				Iterable: &interpreter.VariableExpr{Name: "items"},
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "sum",
						Value: &interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  &interpreter.VariableExpr{Name: "sum"},
							Right: &interpreter.VariableExpr{Name: "item"},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "sum"},
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
	handler := &interpreter.EventHandler{
		EventType: "status.changed",
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "status",
				Value: &interpreter.FieldAccessExpr{
					Object: &interpreter.VariableExpr{Name: "event"},
					Field:  "status",
				},
			},
			&interpreter.SwitchStatement{
				Value: &interpreter.VariableExpr{Name: "status"},
				Cases: []interpreter.SwitchCase{
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "pending"}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
							},
						},
					},
					{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "active"}},
						Body: []interpreter.Statement{
							&interpreter.ReturnStatement{
								Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
							},
						},
					},
				},
				Default: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
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
