package compiler

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/vm"
)

// Test constant folding for arithmetic operations
func TestOptimizer_ConstantFoldingArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		expr     interpreter.Expr
		expected interpreter.Expr
	}{
		{
			name: "2 + 3 = 5",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
		},
		{
			name: "10 - 4 = 6",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Sub,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 4}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 6}},
		},
		{
			name: "6 * 7 = 42",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 6}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 7}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
		},
		{
			name: "20 / 4 = 5",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Div,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 20}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 4}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
		},
		{
			name: "3.5 + 2.5 = 6.0",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  &interpreter.LiteralExpr{Value: interpreter.FloatLiteral{Value: 3.5}},
				Right: &interpreter.LiteralExpr{Value: interpreter.FloatLiteral{Value: 2.5}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.FloatLiteral{Value: 6.0}},
		},
	}

	opt := NewOptimizer(OptBasic)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := opt.OptimizeExpression(tt.expr)
			if !exprsEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test constant folding for comparison operations
func TestOptimizer_ConstantFoldingComparison(t *testing.T) {
	tests := []struct {
		name     string
		expr     interpreter.Expr
		expected interpreter.Expr
	}{
		{
			name: "5 > 3 = true",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Gt,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
		},
		{
			name: "2 < 1 = false",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Lt,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}},
		},
		{
			name: "42 == 42 = true",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Eq,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
		},
		{
			name: "5 != 3 = true",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Ne,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
		},
		{
			name: "10 >= 10 = true",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Ge,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
		},
		{
			name: "5 <= 3 = false",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Le,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}},
		},
	}

	opt := NewOptimizer(OptBasic)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := opt.OptimizeExpression(tt.expr)
			if !exprsEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test constant folding for boolean operations
func TestOptimizer_ConstantFoldingBoolean(t *testing.T) {
	tests := []struct {
		name     string
		expr     interpreter.Expr
		expected interpreter.Expr
	}{
		{
			name: "true && false = false",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.And,
				Left:  &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
				Right: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}},
		},
		{
			name: "true || false = true",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Or,
				Left:  &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
				Right: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
		},
	}

	opt := NewOptimizer(OptBasic)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := opt.OptimizeExpression(tt.expr)
			if !exprsEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test algebraic simplifications
func TestOptimizer_AlgebraicSimplifications(t *testing.T) {
	xVar := &interpreter.VariableExpr{Name: "x"}

	tests := []struct {
		name     string
		expr     interpreter.Expr
		expected interpreter.Expr
	}{
		{
			name: "x + 0 = x",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  xVar,
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			expected: xVar,
		},
		{
			name: "0 + x = x",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
				Right: xVar,
			},
			expected: xVar,
		},
		{
			name: "x - 0 = x",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Sub,
				Left:  xVar,
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			expected: xVar,
		},
		{
			name: "x * 1 = x",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  xVar,
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
			},
			expected: xVar,
		},
		{
			name: "1 * x = x",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
				Right: xVar,
			},
			expected: xVar,
		},
		{
			name: "x * 0 = 0",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  xVar,
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
		},
		{
			name: "0 * x = 0",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
				Right: xVar,
			},
			expected: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
		},
		{
			name: "x / 1 = x",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Div,
				Left:  xVar,
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
			},
			expected: xVar,
		},
	}

	opt := NewOptimizer(OptBasic)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := opt.OptimizeExpression(tt.expr)
			if !exprsEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test nested constant folding
func TestOptimizer_NestedConstantFolding(t *testing.T) {
	// Test: (2 + 3) * 4 = 5 * 4 = 20
	expr := &interpreter.BinaryOpExpr{
		Op: interpreter.Mul,
		Left: &interpreter.BinaryOpExpr{
			Op:    interpreter.Add,
			Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
			Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
		},
		Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 4}},
	}

	expected := &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 20}}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeExpression(expr)

	if !exprsEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Test dead code elimination after return
func TestOptimizer_DeadCodeAfterReturn(t *testing.T) {
	stmts := []interpreter.Statement{
		&interpreter.AssignStatement{
			Target: "x",
			Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
		},
		&interpreter.ReturnStatement{
			Value: &interpreter.VariableExpr{Name: "x"},
		},
		&interpreter.AssignStatement{
			Target: "y",
			Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 100}},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Should only have 2 statements (assignment and return)
	if len(result) != 2 {
		t.Errorf("Expected 2 statements after optimization, got %d", len(result))
	}

	// First should be assignment
	if _, ok := result[0].(*interpreter.AssignStatement); !ok {
		t.Errorf("Expected first statement to be AssignStatement, got %T", result[0])
	}

	// Second should be return
	if _, ok := result[1].(*interpreter.ReturnStatement); !ok {
		t.Errorf("Expected second statement to be ReturnStatement, got %T", result[1])
	}
}

// Test dead code elimination in if statement with constant condition
func TestOptimizer_DeadCodeInIfStatement(t *testing.T) {
	// if true { > 1 } else { > 2 }
	// Should optimize to just: > 1
	stmts := []interpreter.Statement{
		&interpreter.IfStatement{
			Condition: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
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
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Should only have 1 statement (the return from then block)
	if len(result) != 1 {
		t.Errorf("Expected 1 statement after optimization, got %d", len(result))
	}

	// Should be a return statement with value 1
	retStmt, ok := result[0].(*interpreter.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", result[0])
	}

	litExpr, ok := retStmt.Value.(*interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", retStmt.Value)
	}

	intLit, ok := litExpr.Value.(interpreter.IntLiteral)
	if !ok || intLit.Value != 1 {
		t.Errorf("Expected literal value 1, got %v", litExpr.Value)
	}
}

// Test dead code elimination with false constant condition
func TestOptimizer_DeadCodeInIfStatementFalse(t *testing.T) {
	// if false { > 1 } else { > 2 }
	// Should optimize to just: > 2
	stmts := []interpreter.Statement{
		&interpreter.IfStatement{
			Condition: &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}},
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
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Should only have 1 statement (the return from else block)
	if len(result) != 1 {
		t.Errorf("Expected 1 statement after optimization, got %d", len(result))
	}

	// Should be a return statement with value 2
	retStmt, ok := result[0].(*interpreter.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", result[0])
	}

	litExpr, ok := retStmt.Value.(*interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", retStmt.Value)
	}

	intLit, ok := litExpr.Value.(interpreter.IntLiteral)
	if !ok || intLit.Value != 2 {
		t.Errorf("Expected literal value 2, got %v", litExpr.Value)
	}
}

// Test that optimization produces same result as non-optimized code
func TestOptimizer_ExecutionEquivalence(t *testing.T) {
	// Test: $ result = 2 + 3, > result
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "result",
				Value: &interpreter.BinaryOpExpr{
					Op:    interpreter.Add,
					Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
					Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "result"},
			},
		},
	}

	// Compile without optimization
	c1 := NewCompiler()
	bytecode1, err := c1.CompileRoute(route)
	if err != nil {
		t.Fatalf("CompileRoute() without optimization error: %v", err)
	}

	// Compile with optimization
	opt := NewOptimizer(OptBasic)
	optimizedRoute := &interpreter.Route{
		Body: opt.OptimizeStatements(route.Body),
	}

	c2 := NewCompiler()
	bytecode2, err := c2.CompileRoute(optimizedRoute)
	if err != nil {
		t.Fatalf("CompileRoute() with optimization error: %v", err)
	}

	// Execute both
	vm1 := vm.NewVM()
	result1, err := vm1.Execute(bytecode1)
	if err != nil {
		t.Fatalf("Execute() without optimization error: %v", err)
	}

	vm2 := vm.NewVM()
	result2, err := vm2.Execute(bytecode2)
	if err != nil {
		t.Fatalf("Execute() with optimization error: %v", err)
	}

	// Results should be equal
	if !valuesEqual(result1, result2) {
		t.Errorf("Optimized and non-optimized code produced different results: %v vs %v", result1, result2)
	}
}

// Test that optimization with constant condition produces correct result
func TestOptimizer_ConstantConditionExecution(t *testing.T) {
	// Test: if 5 > 3 { > 1 } else { > 2 }
	// Should optimize to: > 1
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

	// Optimize
	opt := NewOptimizer(OptBasic)
	optimizedRoute := &interpreter.Route{
		Body: opt.OptimizeStatements(route.Body),
	}

	// Compile optimized code
	c := NewCompiler()
	bytecode, err := c.CompileRoute(optimizedRoute)
	if err != nil {
		t.Fatalf("CompileRoute() error: %v", err)
	}

	// Execute
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// Should return 1 (from then block)
	expected := vm.IntValue{Val: 1}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Test constant propagation
func TestOptimizer_ConstantPropagation(t *testing.T) {
	// Code: $ x = 5, $ y = x + 3, > y
	// Should become: $ x = 5, $ y = 5 + 3, > y
	// Then constant folding: $ x = 5, $ y = 8, > y
	stmts := []interpreter.Statement{
		&interpreter.AssignStatement{
			Target: "x",
			Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
		},
		&interpreter.AssignStatement{
			Target: "y",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
			},
		},
		&interpreter.ReturnStatement{
			Value: &interpreter.VariableExpr{Name: "y"},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Check that we still have 3 statements
	if len(result) != 3 {
		t.Fatalf("Expected 3 statements, got %d", len(result))
	}

	// Check second statement - y should be assigned 8 (constant folded)
	assignStmt, ok := result[1].(*interpreter.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", result[1])
	}

	litExpr, ok := assignStmt.Value.(*interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr after constant propagation and folding, got %T", assignStmt.Value)
	}

	intLit, ok := litExpr.Value.(interpreter.IntLiteral)
	if !ok || intLit.Value != 8 {
		t.Errorf("Expected y = 8, got %v", litExpr.Value)
	}

	// Check return statement - y should be propagated to literal 8
	retStmt, ok := result[2].(*interpreter.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", result[2])
	}

	retLitExpr, ok := retStmt.Value.(*interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr in return (y should be propagated), got %T", retStmt.Value)
	}

	retIntLit, ok := retLitExpr.Value.(interpreter.IntLiteral)
	if !ok || retIntLit.Value != 8 {
		t.Errorf("Expected return 8, got return %v", retLitExpr.Value)
	}
}

// Test constant propagation with multiple uses
func TestOptimizer_ConstantPropagationMultipleUses(t *testing.T) {
	// Code: $ x = 10, $ y = x * 2, $ z = x + y
	// Should become: $ x = 10, $ y = 10 * 2, $ z = 10 + 20
	// Then: $ x = 10, $ y = 20, $ z = 30
	stmts := []interpreter.Statement{
		&interpreter.AssignStatement{
			Target: "x",
			Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
		},
		&interpreter.AssignStatement{
			Target: "y",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
			},
		},
		&interpreter.AssignStatement{
			Target: "z",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.VariableExpr{Name: "y"},
			},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Check z = 30
	assignStmt, ok := result[2].(*interpreter.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", result[2])
	}

	litExpr, ok := assignStmt.Value.(*interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr for z, got %T", assignStmt.Value)
	}

	intLit, ok := litExpr.Value.(interpreter.IntLiteral)
	if !ok || intLit.Value != 30 {
		t.Errorf("Expected z = 30, got %v", litExpr.Value)
	}
}

// Test constant propagation invalidation
func TestOptimizer_ConstantPropagationInvalidation(t *testing.T) {
	// Code: $ x = 5, $ y = x + 1, $ x = y, $ z = x + 1
	// After first assignment: x is constant 5
	// After second: y is constant 6
	// After third: x is non-constant (assigned from y), so invalidated
	// Fourth: z = x + 1 should NOT be folded (x is not constant anymore)
	stmts := []interpreter.Statement{
		&interpreter.AssignStatement{
			Target: "x",
			Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
		},
		&interpreter.AssignStatement{
			Target: "y",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
			},
		},
		&interpreter.AssignStatement{
			Target: "x",
			Value:  &interpreter.VariableExpr{Name: "y"},
		},
		&interpreter.AssignStatement{
			Target: "z",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
			},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Check third statement - x should be assigned literal 6 (y was propagated)
	assignStmt3, ok := result[2].(*interpreter.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", result[2])
	}

	litExpr, ok := assignStmt3.Value.(*interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr for x = y (y should be propagated), got %T", assignStmt3.Value)
	}

	intLit, ok := litExpr.Value.(interpreter.IntLiteral)
	if !ok || intLit.Value != 6 {
		t.Errorf("Expected x = 6 (propagated from y), got %v", litExpr.Value)
	}

	// Check fourth statement - z = x + 1
	// Since x was reassigned with a literal 6, it should be propagated and folded to 7
	assignStmt4, ok := result[3].(*interpreter.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", result[3])
	}

	litExpr2, ok := assignStmt4.Value.(*interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr for z, got %T", assignStmt4.Value)
	}

	intLit2, ok := litExpr2.Value.(interpreter.IntLiteral)
	if !ok || intLit2.Value != 7 {
		t.Errorf("Expected z = 7, got %v", litExpr2.Value)
	}
}

// Test strength reduction (x * 2 -> x + x)
func TestOptimizer_StrengthReduction(t *testing.T) {
	xVar := &interpreter.VariableExpr{Name: "x"}

	tests := []struct {
		name     string
		expr     interpreter.Expr
		level    OptimizationLevel
		expected string // "add" or "mul"
	}{
		{
			name: "x*2 with OptAggressive",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  xVar,
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
			},
			level:    OptAggressive,
			expected: "add",
		},
		{
			name: "2*x with OptAggressive",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
				Right: xVar,
			},
			level:    OptAggressive,
			expected: "add",
		},
		{
			name: "x*2 with OptBasic",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  xVar,
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
			},
			level:    OptBasic,
			expected: "mul", // Basic level doesn't do strength reduction
		},
		{
			name: "x*3 with OptAggressive",
			expr: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  xVar,
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
			},
			level:    OptAggressive,
			expected: "mul", // Only x*2 is reduced
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := NewOptimizer(tt.level)
			result := opt.OptimizeExpression(tt.expr)

			if tt.expected == "add" {
				binOp, ok := result.(*interpreter.BinaryOpExpr)
				if !ok {
					t.Fatalf("Expected BinaryOpExpr, got %T", result)
				}
				if binOp.Op != interpreter.Add {
					t.Errorf("Expected Add operation, got %v", binOp.Op)
				}
			} else if tt.expected == "mul" {
				binOp, ok := result.(*interpreter.BinaryOpExpr)
				if !ok {
					t.Fatalf("Expected BinaryOpExpr, got %T", result)
				}
				if binOp.Op != interpreter.Mul {
					t.Errorf("Expected Mul operation, got %v", binOp.Op)
				}
			}
		})
	}
}

// Test common subexpression elimination
func TestOptimizer_CSE(t *testing.T) {
	// Code: $ a = x + y, $ b = x + y
	// Should become: $ a = x + y, $ b = a
	stmts := []interpreter.Statement{
		&interpreter.AssignStatement{
			Target: "a",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.VariableExpr{Name: "y"},
			},
		},
		&interpreter.AssignStatement{
			Target: "b",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.VariableExpr{Name: "y"},
			},
		},
	}

	opt := NewOptimizer(OptAggressive)
	result := opt.OptimizeStatements(stmts)

	// Check that we have 2 statements
	if len(result) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(result))
	}

	// Check second statement - b should be assigned from a
	assignStmt, ok := result[1].(*interpreter.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", result[1])
	}

	varExpr, ok := assignStmt.Value.(*interpreter.VariableExpr)
	if !ok {
		t.Fatalf("Expected VariableExpr (CSE should replace with 'a'), got %T", assignStmt.Value)
	}

	if varExpr.Name != "a" {
		t.Errorf("Expected b = a (CSE), got b = %s", varExpr.Name)
	}
}

// Test CSE only works at OptAggressive level
func TestOptimizer_CSE_LevelCheck(t *testing.T) {
	stmts := []interpreter.Statement{
		&interpreter.AssignStatement{
			Target: "a",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.VariableExpr{Name: "y"},
			},
		},
		&interpreter.AssignStatement{
			Target: "b",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.VariableExpr{Name: "y"},
			},
		},
	}

	// Test with OptBasic - should NOT do CSE
	optBasic := NewOptimizer(OptBasic)
	resultBasic := optBasic.OptimizeStatements(stmts)

	assignStmt, ok := resultBasic[1].(*interpreter.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", resultBasic[1])
	}

	// Should still be a BinaryOpExpr (not CSE'd)
	if _, ok := assignStmt.Value.(*interpreter.BinaryOpExpr); !ok {
		t.Errorf("OptBasic should NOT perform CSE, but got %T", assignStmt.Value)
	}

	// Test with OptAggressive - should do CSE
	optAgg := NewOptimizer(OptAggressive)
	resultAgg := optAgg.OptimizeStatements(stmts)

	assignStmt2, ok := resultAgg[1].(*interpreter.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", resultAgg[1])
	}

	// Should be a VariableExpr (CSE'd)
	varExpr, ok := assignStmt2.Value.(*interpreter.VariableExpr)
	if !ok {
		t.Errorf("OptAggressive should perform CSE, got %T", assignStmt2.Value)
	} else if varExpr.Name != "a" {
		t.Errorf("Expected CSE to replace with 'a', got '%s'", varExpr.Name)
	}
}

// Test CSE with three identical expressions
func TestOptimizer_CSE_MultipleUses(t *testing.T) {
	// Code: $ a = x * y, $ b = x * y, $ c = x * y
	// Should become: $ a = x * y, $ b = a, $ c = a
	stmts := []interpreter.Statement{
		&interpreter.AssignStatement{
			Target: "a",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.VariableExpr{Name: "y"},
			},
		},
		&interpreter.AssignStatement{
			Target: "b",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.VariableExpr{Name: "y"},
			},
		},
		&interpreter.AssignStatement{
			Target: "c",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Mul,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.VariableExpr{Name: "y"},
			},
		},
	}

	opt := NewOptimizer(OptAggressive)
	result := opt.OptimizeStatements(stmts)

	// Check b = a
	assignB, ok := result[1].(*interpreter.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement for b, got %T", result[1])
	}

	varExprB, ok := assignB.Value.(*interpreter.VariableExpr)
	if !ok || varExprB.Name != "a" {
		t.Errorf("Expected b = a, got b = %v", assignB.Value)
	}

	// Check c = a
	assignC, ok := result[2].(*interpreter.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement for c, got %T", result[2])
	}

	varExprC, ok := assignC.Value.(*interpreter.VariableExpr)
	if !ok || varExprC.Name != "a" {
		t.Errorf("Expected c = a, got c = %v", assignC.Value)
	}
}

// Test copy propagation
func TestOptimizer_CopyPropagation(t *testing.T) {
	// Code: $ x = 10, $ y = x, $ z = y + 5
	// Should become: $ x = 10, $ y = x, $ z = x + 5
	stmts := []interpreter.Statement{
		&interpreter.AssignStatement{
			Target: "x",
			Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
		},
		&interpreter.AssignStatement{
			Target: "y",
			Value:  &interpreter.VariableExpr{Name: "x"},
		},
		&interpreter.AssignStatement{
			Target: "z",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  &interpreter.VariableExpr{Name: "y"},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
			},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Check third statement - should be constant folded to 15
	assignStmt, ok := result[2].(*interpreter.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", result[2])
	}

	// With copy propagation and constant propagation, z = y + 5 becomes z = x + 5 becomes z = 10 + 5 = 15
	litExpr, ok := assignStmt.Value.(*interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr (constant folded to 15), got %T", assignStmt.Value)
	}

	intLit, ok := litExpr.Value.(interpreter.IntLiteral)
	if !ok || intLit.Value != 15 {
		t.Errorf("Expected z = 15, got %v", litExpr.Value)
	}
}

// Test copy propagation with chains
func TestOptimizer_CopyPropagationChain(t *testing.T) {
	// Code: $ a = 5, $ b = a, $ c = b, $ d = c
	// Should track: b->a, c->a (via b), d->a (via c->b->a)
	stmts := []interpreter.Statement{
		&interpreter.AssignStatement{
			Target: "a",
			Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
		},
		&interpreter.AssignStatement{
			Target: "b",
			Value:  &interpreter.VariableExpr{Name: "a"},
		},
		&interpreter.AssignStatement{
			Target: "c",
			Value:  &interpreter.VariableExpr{Name: "b"},
		},
		&interpreter.AssignStatement{
			Target: "d",
			Value:  &interpreter.VariableExpr{Name: "c"},
		},
		&interpreter.ReturnStatement{
			Value: &interpreter.VariableExpr{Name: "d"},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Check that d is assigned from a (or constant 5)
	assignD, ok := result[3].(*interpreter.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement for d, got %T", result[3])
	}

	// Should resolve to either variable 'a' or constant 5
	switch v := assignD.Value.(type) {
	case *interpreter.VariableExpr:
		if v.Name != "a" {
			t.Errorf("Expected d = a (copy propagated), got d = %s", v.Name)
		}
	case *interpreter.LiteralExpr:
		// Constant propagated
		intLit, ok := v.Value.(interpreter.IntLiteral)
		if !ok || intLit.Value != 5 {
			t.Errorf("Expected d = 5 (constant propagated), got d = %v", v.Value)
		}
	default:
		t.Errorf("Expected VariableExpr or LiteralExpr for d, got %T", assignD.Value)
	}

	// Check return statement - should also be resolved
	retStmt, ok := result[4].(*interpreter.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", result[4])
	}

	// Return should be constant 5 or variable 'a'
	switch v := retStmt.Value.(type) {
	case *interpreter.VariableExpr:
		if v.Name != "a" {
			t.Errorf("Expected return a, got return %s", v.Name)
		}
	case *interpreter.LiteralExpr:
		intLit, ok := v.Value.(interpreter.IntLiteral)
		if !ok || intLit.Value != 5 {
			t.Errorf("Expected return 5, got return %v", v.Value)
		}
	default:
		t.Errorf("Expected VariableExpr or LiteralExpr in return, got %T", retStmt.Value)
	}
}

// Test loop invariant code motion (LICM)
func TestOptimizer_LICM_Basic(t *testing.T) {
	// Code: while (i < 10) { $ x = a + b, $ i = i + 1 }
	// Should become: $ x = a + b, while (i < 10) { $ i = i + 1 }
	stmts := []interpreter.Statement{
		&interpreter.WhileStatement{
			Condition: &interpreter.BinaryOpExpr{
				Op:    interpreter.Lt,
				Left:  &interpreter.VariableExpr{Name: "i"},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
			},
			Body: []interpreter.Statement{
				&interpreter.AssignStatement{
					Target: "x",
					Value: &interpreter.BinaryOpExpr{
						Op:    interpreter.Add,
						Left:  &interpreter.VariableExpr{Name: "a"},
						Right: &interpreter.VariableExpr{Name: "b"},
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
	}

	opt := NewOptimizer(OptAggressive)
	result := opt.OptimizeStatements(stmts)

	// Should have 2 statements: the hoisted assignment and the while loop
	if len(result) < 2 {
		t.Fatalf("Expected at least 2 statements (hoisted + while), got %d", len(result))
	}

	// First statement should be the hoisted x = a + b
	assignStmt, ok := result[0].(*interpreter.AssignStatement)
	if !ok {
		t.Fatalf("Expected first statement to be hoisted AssignStatement, got %T", result[0])
	}
	if assignStmt.Target != "x" {
		t.Errorf("Expected hoisted assignment to x, got %s", assignStmt.Target)
	}

	// Second statement should be the while loop
	whileStmt, ok := result[1].(*interpreter.WhileStatement)
	if !ok {
		t.Fatalf("Expected second statement to be WhileStatement, got %T", result[1])
	}

	// Loop body should only have the i = i + 1 assignment
	if len(whileStmt.Body) != 1 {
		t.Errorf("Expected loop body to have 1 statement after LICM, got %d", len(whileStmt.Body))
	}
}

// Test LICM doesn't move condition variables
func TestOptimizer_LICM_ConditionVariable(t *testing.T) {
	// Code: while (i < 10) { $ i = a + b }
	// Should NOT move i assignment out (i is in condition)
	stmts := []interpreter.Statement{
		&interpreter.WhileStatement{
			Condition: &interpreter.BinaryOpExpr{
				Op:    interpreter.Lt,
				Left:  &interpreter.VariableExpr{Name: "i"},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
			},
			Body: []interpreter.Statement{
				&interpreter.AssignStatement{
					Target: "i",
					Value: &interpreter.BinaryOpExpr{
						Op:    interpreter.Add,
						Left:  &interpreter.VariableExpr{Name: "a"},
						Right: &interpreter.VariableExpr{Name: "b"},
					},
				},
			},
		},
	}

	opt := NewOptimizer(OptAggressive)
	result := opt.OptimizeStatements(stmts)

	// Should have 1 statement (the while loop)
	if len(result) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(result))
	}

	whileStmt, ok := result[0].(*interpreter.WhileStatement)
	if !ok {
		t.Fatalf("Expected WhileStatement, got %T", result[0])
	}

	// Loop body should still have the assignment (not hoisted)
	if len(whileStmt.Body) != 1 {
		t.Errorf("Expected loop body to have 1 statement, got %d", len(whileStmt.Body))
	}
}

// Test LICM doesn't move variant code
func TestOptimizer_LICM_VariantExpression(t *testing.T) {
	// Code: while (i < 10) { $ x = i + 1, $ i = i + 1 }
	// Should NOT move x assignment (depends on i which is modified)
	stmts := []interpreter.Statement{
		&interpreter.WhileStatement{
			Condition: &interpreter.BinaryOpExpr{
				Op:    interpreter.Lt,
				Left:  &interpreter.VariableExpr{Name: "i"},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
			},
			Body: []interpreter.Statement{
				&interpreter.AssignStatement{
					Target: "x",
					Value: &interpreter.BinaryOpExpr{
						Op:    interpreter.Add,
						Left:  &interpreter.VariableExpr{Name: "i"},
						Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
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
	}

	opt := NewOptimizer(OptAggressive)
	result := opt.OptimizeStatements(stmts)

	// Should have 1 statement (the while loop)
	if len(result) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(result))
	}

	whileStmt, ok := result[0].(*interpreter.WhileStatement)
	if !ok {
		t.Fatalf("Expected WhileStatement, got %T", result[0])
	}

	// Loop body should still have both assignments
	if len(whileStmt.Body) != 2 {
		t.Errorf("Expected loop body to have 2 statements, got %d", len(whileStmt.Body))
	}
}

// Test LICM only works at OptAggressive
func TestOptimizer_LICM_LevelCheck(t *testing.T) {
	stmts := []interpreter.Statement{
		&interpreter.WhileStatement{
			Condition: &interpreter.BinaryOpExpr{
				Op:    interpreter.Lt,
				Left:  &interpreter.VariableExpr{Name: "i"},
				Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
			},
			Body: []interpreter.Statement{
				&interpreter.AssignStatement{
					Target: "x",
					Value: &interpreter.BinaryOpExpr{
						Op:    interpreter.Add,
						Left:  &interpreter.VariableExpr{Name: "a"},
						Right: &interpreter.VariableExpr{Name: "b"},
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
	}

	// Test with OptBasic - should NOT do LICM
	optBasic := NewOptimizer(OptBasic)
	resultBasic := optBasic.OptimizeStatements(stmts)

	if len(resultBasic) != 1 {
		t.Errorf("OptBasic should not hoist code, expected 1 statement, got %d", len(resultBasic))
	}

	// Test with OptAggressive - should do LICM
	optAgg := NewOptimizer(OptAggressive)
	resultAgg := optAgg.OptimizeStatements(stmts)

	if len(resultAgg) < 2 {
		t.Errorf("OptAggressive should hoist code, expected at least 2 statements, got %d", len(resultAgg))
	}
}

// Test optimization level none
func TestOptimizer_OptNone(t *testing.T) {
	expr := &interpreter.BinaryOpExpr{
		Op:    interpreter.Add,
		Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
		Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
	}

	opt := NewOptimizer(OptNone)
	result := opt.OptimizeExpression(expr)

	// With OptNone, expression should remain unchanged
	binOp, ok := result.(*interpreter.BinaryOpExpr)
	if !ok {
		t.Fatalf("Expected BinaryOpExpr, got %T", result)
	}

	if binOp.Op != interpreter.Add {
		t.Errorf("Expected Add operation, got %v", binOp.Op)
	}
}

// Helper function to compare expressions
func exprsEqual(a, b interpreter.Expr) bool {
	aLit, aIsLit := a.(*interpreter.LiteralExpr)
	bLit, bIsLit := b.(*interpreter.LiteralExpr)

	if aIsLit && bIsLit {
		return literalsEqual(aLit.Value, bLit.Value)
	}

	aVar, aIsVar := a.(*interpreter.VariableExpr)
	bVar, bIsVar := b.(*interpreter.VariableExpr)

	if aIsVar && bIsVar {
		return aVar.Name == bVar.Name
	}

	aBinOp, aIsBinOp := a.(*interpreter.BinaryOpExpr)
	bBinOp, bIsBinOp := b.(*interpreter.BinaryOpExpr)

	if aIsBinOp && bIsBinOp {
		return aBinOp.Op == bBinOp.Op &&
			exprsEqual(aBinOp.Left, bBinOp.Left) &&
			exprsEqual(aBinOp.Right, bBinOp.Right)
	}

	return false
}

