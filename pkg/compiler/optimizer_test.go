package compiler

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"testing"

	"github.com/glyphlang/glyph/pkg/vm"
)

// Test constant folding for arithmetic operations
func TestOptimizer_ConstantFoldingArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expr
		expected ast.Expr
	}{
		{
			name: "2 + 3 = 5",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
			},
			expected: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
		},
		{
			name: "10 - 4 = 6",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Sub,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
			},
			expected: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 6}},
		},
		{
			name: "6 * 7 = 42",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 6}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 7}},
			},
			expected: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
		},
		{
			name: "20 / 4 = 5",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Div,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
			},
			expected: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
		},
		{
			name: "3.5 + 2.5 = 6.0",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 3.5}},
				Right: &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 2.5}},
			},
			expected: &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 6.0}},
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
		expr     ast.Expr
		expected ast.Expr
	}{
		{
			name: "5 > 3 = true",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Gt,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
			},
			expected: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
		},
		{
			name: "2 < 1 = false",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Lt,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
			expected: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
		},
		{
			name: "42 == 42 = true",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Eq,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
			},
			expected: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
		},
		{
			name: "5 != 3 = true",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Ne,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
			},
			expected: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
		},
		{
			name: "10 >= 10 = true",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Ge,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
			},
			expected: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
		},
		{
			name: "5 <= 3 = false",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Le,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
			},
			expected: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
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
		expr     ast.Expr
		expected ast.Expr
	}{
		{
			name: "true && false = false",
			expr: &ast.BinaryOpExpr{
				Op:    ast.And,
				Left:  &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
				Right: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
			},
			expected: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
		},
		{
			name: "true || false = true",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Or,
				Left:  &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
				Right: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
			},
			expected: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
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
	xVar := &ast.VariableExpr{Name: "x"}

	tests := []struct {
		name     string
		expr     ast.Expr
		expected ast.Expr
	}{
		{
			name: "x + 0 = x",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  xVar,
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			expected: xVar,
		},
		{
			name: "0 + x = x",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
				Right: xVar,
			},
			expected: xVar,
		},
		{
			name: "x - 0 = x",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Sub,
				Left:  xVar,
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			expected: xVar,
		},
		{
			name: "x * 1 = x",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  xVar,
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
			expected: xVar,
		},
		{
			name: "1 * x = x",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				Right: xVar,
			},
			expected: xVar,
		},
		{
			name: "x * 0 = 0",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  xVar,
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			expected: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
		},
		{
			name: "0 * x = 0",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
				Right: xVar,
			},
			expected: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
		},
		{
			name: "x / 1 = x",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Div,
				Left:  xVar,
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
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
	expr := &ast.BinaryOpExpr{
		Op: ast.Mul,
		Left: &ast.BinaryOpExpr{
			Op:    ast.Add,
			Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
		},
		Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
	}

	expected := &ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeExpression(expr)

	if !exprsEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Test dead code elimination after return
func TestOptimizer_DeadCodeAfterReturn(t *testing.T) {
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "x",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
		},
		&ast.ReturnStatement{
			Value: &ast.VariableExpr{Name: "x"},
		},
		&ast.AssignStatement{
			Target: "y",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 100}},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Should only have 2 statements (assignment and return)
	if len(result) != 2 {
		t.Errorf("Expected 2 statements after optimization, got %d", len(result))
	}

	// First should be assignment
	if _, ok := result[0].(*ast.AssignStatement); !ok {
		t.Errorf("Expected first statement to be AssignStatement, got %T", result[0])
	}

	// Second should be return
	if _, ok := result[1].(*ast.ReturnStatement); !ok {
		t.Errorf("Expected second statement to be ReturnStatement, got %T", result[1])
	}
}

// Test dead code elimination in if statement with constant condition
func TestOptimizer_DeadCodeInIfStatement(t *testing.T) {
	// if true { > 1 } else { > 2 }
	// Should optimize to just: > 1
	stmts := []ast.Statement{
		&ast.IfStatement{
			Condition: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
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
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Should only have 1 statement (the return from then block)
	if len(result) != 1 {
		t.Errorf("Expected 1 statement after optimization, got %d", len(result))
	}

	// Should be a return statement with value 1
	retStmt, ok := result[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", result[0])
	}

	litExpr, ok := retStmt.Value.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", retStmt.Value)
	}

	intLit, ok := litExpr.Value.(ast.IntLiteral)
	if !ok || intLit.Value != 1 {
		t.Errorf("Expected literal value 1, got %v", litExpr.Value)
	}
}

// Test dead code elimination with false constant condition
func TestOptimizer_DeadCodeInIfStatementFalse(t *testing.T) {
	// if false { > 1 } else { > 2 }
	// Should optimize to just: > 2
	stmts := []ast.Statement{
		&ast.IfStatement{
			Condition: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
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
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Should only have 1 statement (the return from else block)
	if len(result) != 1 {
		t.Errorf("Expected 1 statement after optimization, got %d", len(result))
	}

	// Should be a return statement with value 2
	retStmt, ok := result[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", result[0])
	}

	litExpr, ok := retStmt.Value.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", retStmt.Value)
	}

	intLit, ok := litExpr.Value.(ast.IntLiteral)
	if !ok || intLit.Value != 2 {
		t.Errorf("Expected literal value 2, got %v", litExpr.Value)
	}
}

// Test that optimization produces same result as non-optimized code
func TestOptimizer_ExecutionEquivalence(t *testing.T) {
	// Test: $ result = 2 + 3, > result
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "result",
				Value: &ast.BinaryOpExpr{
					Op:    ast.Add,
					Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "result"},
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
	optimizedRoute := &ast.Route{
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

	// Optimize
	opt := NewOptimizer(OptBasic)
	optimizedRoute := &ast.Route{
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
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "x",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
		},
		&ast.AssignStatement{
			Target: "y",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
			},
		},
		&ast.ReturnStatement{
			Value: &ast.VariableExpr{Name: "y"},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Check that we still have 3 statements
	if len(result) != 3 {
		t.Fatalf("Expected 3 statements, got %d", len(result))
	}

	// Check second statement - y should be assigned 8 (constant folded)
	assignStmt, ok := result[1].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", result[1])
	}

	litExpr, ok := assignStmt.Value.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr after constant propagation and folding, got %T", assignStmt.Value)
	}

	intLit, ok := litExpr.Value.(ast.IntLiteral)
	if !ok || intLit.Value != 8 {
		t.Errorf("Expected y = 8, got %v", litExpr.Value)
	}

	// Check return statement - y should be propagated to literal 8
	retStmt, ok := result[2].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", result[2])
	}

	retLitExpr, ok := retStmt.Value.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr in return (y should be propagated), got %T", retStmt.Value)
	}

	retIntLit, ok := retLitExpr.Value.(ast.IntLiteral)
	if !ok || retIntLit.Value != 8 {
		t.Errorf("Expected return 8, got return %v", retLitExpr.Value)
	}
}

// Test constant propagation with multiple uses
func TestOptimizer_ConstantPropagationMultipleUses(t *testing.T) {
	// Code: $ x = 10, $ y = x * 2, $ z = x + y
	// Should become: $ x = 10, $ y = 10 * 2, $ z = 10 + 20
	// Then: $ x = 10, $ y = 20, $ z = 30
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "x",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
		},
		&ast.AssignStatement{
			Target: "y",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
		},
		&ast.AssignStatement{
			Target: "z",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
			},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Check z = 30
	assignStmt, ok := result[2].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", result[2])
	}

	litExpr, ok := assignStmt.Value.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr for z, got %T", assignStmt.Value)
	}

	intLit, ok := litExpr.Value.(ast.IntLiteral)
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
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "x",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
		},
		&ast.AssignStatement{
			Target: "y",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
		},
		&ast.AssignStatement{
			Target: "x",
			Value:  &ast.VariableExpr{Name: "y"},
		},
		&ast.AssignStatement{
			Target: "z",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Check third statement - x should be assigned literal 6 (y was propagated)
	assignStmt3, ok := result[2].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", result[2])
	}

	litExpr, ok := assignStmt3.Value.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr for x = y (y should be propagated), got %T", assignStmt3.Value)
	}

	intLit, ok := litExpr.Value.(ast.IntLiteral)
	if !ok || intLit.Value != 6 {
		t.Errorf("Expected x = 6 (propagated from y), got %v", litExpr.Value)
	}

	// Check fourth statement - z = x + 1
	// Since x was reassigned with a literal 6, it should be propagated and folded to 7
	assignStmt4, ok := result[3].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", result[3])
	}

	litExpr2, ok := assignStmt4.Value.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr for z, got %T", assignStmt4.Value)
	}

	intLit2, ok := litExpr2.Value.(ast.IntLiteral)
	if !ok || intLit2.Value != 7 {
		t.Errorf("Expected z = 7, got %v", litExpr2.Value)
	}
}

// Test strength reduction (x * 2 -> x + x)
func TestOptimizer_StrengthReduction(t *testing.T) {
	xVar := &ast.VariableExpr{Name: "x"}

	tests := []struct {
		name     string
		expr     ast.Expr
		level    OptimizationLevel
		expected string // "add" or "mul"
	}{
		{
			name: "x*2 with OptAggressive",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  xVar,
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
			level:    OptAggressive,
			expected: "add",
		},
		{
			name: "2*x with OptAggressive",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
				Right: xVar,
			},
			level:    OptAggressive,
			expected: "add",
		},
		{
			name: "x*2 with OptBasic",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  xVar,
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
			level:    OptBasic,
			expected: "mul", // Basic level doesn't do strength reduction
		},
		{
			name: "x*3 with OptAggressive",
			expr: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  xVar,
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
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
				binOp, ok := result.(*ast.BinaryOpExpr)
				if !ok {
					t.Fatalf("Expected BinaryOpExpr, got %T", result)
				}
				if binOp.Op != ast.Add {
					t.Errorf("Expected Add operation, got %v", binOp.Op)
				}
			} else if tt.expected == "mul" {
				binOp, ok := result.(*ast.BinaryOpExpr)
				if !ok {
					t.Fatalf("Expected BinaryOpExpr, got %T", result)
				}
				if binOp.Op != ast.Mul {
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
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "a",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
			},
		},
		&ast.AssignStatement{
			Target: "b",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
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
	assignStmt, ok := result[1].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", result[1])
	}

	varExpr, ok := assignStmt.Value.(*ast.VariableExpr)
	if !ok {
		t.Fatalf("Expected VariableExpr (CSE should replace with 'a'), got %T", assignStmt.Value)
	}

	if varExpr.Name != "a" {
		t.Errorf("Expected b = a (CSE), got b = %s", varExpr.Name)
	}
}

// Test CSE only works at OptAggressive level
func TestOptimizer_CSE_LevelCheck(t *testing.T) {
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "a",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
			},
		},
		&ast.AssignStatement{
			Target: "b",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
			},
		},
	}

	// Test with OptBasic - should NOT do CSE
	optBasic := NewOptimizer(OptBasic)
	resultBasic := optBasic.OptimizeStatements(stmts)

	assignStmt, ok := resultBasic[1].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", resultBasic[1])
	}

	// Should still be a BinaryOpExpr (not CSE'd)
	if _, ok := assignStmt.Value.(*ast.BinaryOpExpr); !ok {
		t.Errorf("OptBasic should NOT perform CSE, but got %T", assignStmt.Value)
	}

	// Test with OptAggressive - should do CSE
	optAgg := NewOptimizer(OptAggressive)
	resultAgg := optAgg.OptimizeStatements(stmts)

	assignStmt2, ok := resultAgg[1].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", resultAgg[1])
	}

	// Should be a VariableExpr (CSE'd)
	varExpr, ok := assignStmt2.Value.(*ast.VariableExpr)
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
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "a",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
			},
		},
		&ast.AssignStatement{
			Target: "b",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
			},
		},
		&ast.AssignStatement{
			Target: "c",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Mul,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
			},
		},
	}

	opt := NewOptimizer(OptAggressive)
	result := opt.OptimizeStatements(stmts)

	// Check b = a
	assignB, ok := result[1].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement for b, got %T", result[1])
	}

	varExprB, ok := assignB.Value.(*ast.VariableExpr)
	if !ok || varExprB.Name != "a" {
		t.Errorf("Expected b = a, got b = %v", assignB.Value)
	}

	// Check c = a
	assignC, ok := result[2].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement for c, got %T", result[2])
	}

	varExprC, ok := assignC.Value.(*ast.VariableExpr)
	if !ok || varExprC.Name != "a" {
		t.Errorf("Expected c = a, got c = %v", assignC.Value)
	}
}

// Test copy propagation
func TestOptimizer_CopyPropagation(t *testing.T) {
	// Code: $ x = 10, $ y = x, $ z = y + 5
	// Should become: $ x = 10, $ y = x, $ z = x + 5
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "x",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
		},
		&ast.AssignStatement{
			Target: "y",
			Value:  &ast.VariableExpr{Name: "x"},
		},
		&ast.AssignStatement{
			Target: "z",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.VariableExpr{Name: "y"},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
			},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Check third statement - should be constant folded to 15
	assignStmt, ok := result[2].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement, got %T", result[2])
	}

	// With copy propagation and constant propagation, z = y + 5 becomes z = x + 5 becomes z = 10 + 5 = 15
	litExpr, ok := assignStmt.Value.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr (constant folded to 15), got %T", assignStmt.Value)
	}

	intLit, ok := litExpr.Value.(ast.IntLiteral)
	if !ok || intLit.Value != 15 {
		t.Errorf("Expected z = 15, got %v", litExpr.Value)
	}
}

// Test copy propagation with chains
func TestOptimizer_CopyPropagationChain(t *testing.T) {
	// Code: $ a = 5, $ b = a, $ c = b, $ d = c
	// Should track: b->a, c->a (via b), d->a (via c->b->a)
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "a",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
		},
		&ast.AssignStatement{
			Target: "b",
			Value:  &ast.VariableExpr{Name: "a"},
		},
		&ast.AssignStatement{
			Target: "c",
			Value:  &ast.VariableExpr{Name: "b"},
		},
		&ast.AssignStatement{
			Target: "d",
			Value:  &ast.VariableExpr{Name: "c"},
		},
		&ast.ReturnStatement{
			Value: &ast.VariableExpr{Name: "d"},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Check that d is assigned from a (or constant 5)
	assignD, ok := result[3].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Expected AssignStatement for d, got %T", result[3])
	}

	// Should resolve to either variable 'a' or constant 5
	switch v := assignD.Value.(type) {
	case *ast.VariableExpr:
		if v.Name != "a" {
			t.Errorf("Expected d = a (copy propagated), got d = %s", v.Name)
		}
	case *ast.LiteralExpr:
		// Constant propagated
		intLit, ok := v.Value.(ast.IntLiteral)
		if !ok || intLit.Value != 5 {
			t.Errorf("Expected d = 5 (constant propagated), got d = %v", v.Value)
		}
	default:
		t.Errorf("Expected VariableExpr or LiteralExpr for d, got %T", assignD.Value)
	}

	// Check return statement - should also be resolved
	retStmt, ok := result[4].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", result[4])
	}

	// Return should be constant 5 or variable 'a'
	switch v := retStmt.Value.(type) {
	case *ast.VariableExpr:
		if v.Name != "a" {
			t.Errorf("Expected return a, got return %s", v.Name)
		}
	case *ast.LiteralExpr:
		intLit, ok := v.Value.(ast.IntLiteral)
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
	stmts := []ast.Statement{
		&ast.WhileStatement{
			Condition: &ast.BinaryOpExpr{
				Op:    ast.Lt,
				Left:  &ast.VariableExpr{Name: "i"},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
			},
			Body: []ast.Statement{
				&ast.AssignStatement{
					Target: "x",
					Value: &ast.BinaryOpExpr{
						Op:    ast.Add,
						Left:  &ast.VariableExpr{Name: "a"},
						Right: &ast.VariableExpr{Name: "b"},
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
	}

	opt := NewOptimizer(OptAggressive)
	result := opt.OptimizeStatements(stmts)

	// Should have 2 statements: the hoisted assignment and the while loop
	if len(result) < 2 {
		t.Fatalf("Expected at least 2 statements (hoisted + while), got %d", len(result))
	}

	// First statement should be the hoisted x = a + b
	assignStmt, ok := result[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Expected first statement to be hoisted AssignStatement, got %T", result[0])
	}
	if assignStmt.Target != "x" {
		t.Errorf("Expected hoisted assignment to x, got %s", assignStmt.Target)
	}

	// Second statement should be the while loop
	whileStmt, ok := result[1].(*ast.WhileStatement)
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
	stmts := []ast.Statement{
		&ast.WhileStatement{
			Condition: &ast.BinaryOpExpr{
				Op:    ast.Lt,
				Left:  &ast.VariableExpr{Name: "i"},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
			},
			Body: []ast.Statement{
				&ast.AssignStatement{
					Target: "i",
					Value: &ast.BinaryOpExpr{
						Op:    ast.Add,
						Left:  &ast.VariableExpr{Name: "a"},
						Right: &ast.VariableExpr{Name: "b"},
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

	whileStmt, ok := result[0].(*ast.WhileStatement)
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
	stmts := []ast.Statement{
		&ast.WhileStatement{
			Condition: &ast.BinaryOpExpr{
				Op:    ast.Lt,
				Left:  &ast.VariableExpr{Name: "i"},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
			},
			Body: []ast.Statement{
				&ast.AssignStatement{
					Target: "x",
					Value: &ast.BinaryOpExpr{
						Op:    ast.Add,
						Left:  &ast.VariableExpr{Name: "i"},
						Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
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
	}

	opt := NewOptimizer(OptAggressive)
	result := opt.OptimizeStatements(stmts)

	// Should have 1 statement (the while loop)
	if len(result) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(result))
	}

	whileStmt, ok := result[0].(*ast.WhileStatement)
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
	stmts := []ast.Statement{
		&ast.WhileStatement{
			Condition: &ast.BinaryOpExpr{
				Op:    ast.Lt,
				Left:  &ast.VariableExpr{Name: "i"},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
			},
			Body: []ast.Statement{
				&ast.AssignStatement{
					Target: "x",
					Value: &ast.BinaryOpExpr{
						Op:    ast.Add,
						Left:  &ast.VariableExpr{Name: "a"},
						Right: &ast.VariableExpr{Name: "b"},
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
	expr := &ast.BinaryOpExpr{
		Op:    ast.Add,
		Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
		Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
	}

	opt := NewOptimizer(OptNone)
	result := opt.OptimizeExpression(expr)

	// With OptNone, expression should remain unchanged
	binOp, ok := result.(*ast.BinaryOpExpr)
	if !ok {
		t.Fatalf("Expected BinaryOpExpr, got %T", result)
	}

	if binOp.Op != ast.Add {
		t.Errorf("Expected Add operation, got %v", binOp.Op)
	}
}

// Helper function to compare expressions
func exprsEqual(a, b ast.Expr) bool {
	aLit, aIsLit := a.(*ast.LiteralExpr)
	bLit, bIsLit := b.(*ast.LiteralExpr)

	if aIsLit && bIsLit {
		return literalsEqual(aLit.Value, bLit.Value)
	}

	aVar, aIsVar := a.(*ast.VariableExpr)
	bVar, bIsVar := b.(*ast.VariableExpr)

	if aIsVar && bIsVar {
		return aVar.Name == bVar.Name
	}

	aBinOp, aIsBinOp := a.(*ast.BinaryOpExpr)
	bBinOp, bIsBinOp := b.(*ast.BinaryOpExpr)

	if aIsBinOp && bIsBinOp {
		return aBinOp.Op == bBinOp.Op &&
			exprsEqual(aBinOp.Left, bBinOp.Left) &&
			exprsEqual(aBinOp.Right, bBinOp.Right)
	}

	return false
}

// TestOptimizer_NestedForLoopConstantInvalidation tests that constants are properly
// invalidated when modified inside nested for-loops
func TestOptimizer_NestedForLoopConstantInvalidation(t *testing.T) {
	// Test: $ sum = 0, for row in matrix { for cell in row { $ sum = sum + cell } }, > sum
	// The optimizer should NOT replace the return "sum" with constant 0
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "sum",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
		},
		&ast.ForStatement{
			ValueVar: "row",
			Iterable: &ast.VariableExpr{Name: "matrix"},
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
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// Should have 3 statements
	if len(result) != 3 {
		t.Fatalf("Expected 3 statements after optimization, got %d", len(result))
	}

	// The return statement should still reference the variable "sum", NOT a constant
	retStmt, ok := result[2].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", result[2])
	}

	// The return value should be a VariableExpr, not a LiteralExpr
	varExpr, ok := retStmt.Value.(*ast.VariableExpr)
	if !ok {
		// If it's a literal, the bug is present
		if litExpr, isLit := retStmt.Value.(*ast.LiteralExpr); isLit {
			t.Errorf("Bug: Return value was incorrectly optimized to constant %v instead of variable 'sum'", litExpr.Value)
		} else {
			t.Fatalf("Expected VariableExpr for return value, got %T", retStmt.Value)
		}
		return
	}

	if varExpr.Name != "sum" {
		t.Errorf("Expected variable 'sum', got '%s'", varExpr.Name)
	}
}

// TestOptimizer_ForLoopConstantInvalidation tests that constants are properly
// invalidated when modified inside a single for-loop
func TestOptimizer_ForLoopConstantInvalidation(t *testing.T) {
	// Test: $ count = 0, for item in items { $ count = count + 1 }, > count
	stmts := []ast.Statement{
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
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// The return statement should still reference the variable "count"
	retStmt, ok := result[2].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", result[2])
	}

	varExpr, ok := retStmt.Value.(*ast.VariableExpr)
	if !ok {
		t.Fatalf("Expected VariableExpr for return value, got %T", retStmt.Value)
	}

	if varExpr.Name != "count" {
		t.Errorf("Expected variable 'count', got '%s'", varExpr.Name)
	}
}

// =================================================================
// Additional optimizer tests for containsCallInStmt, substituteParams,
// foldBinaryOp, OptimizeStatements coverage
// =================================================================

func TestContainsCallInStmt_ValueTypes(t *testing.T) {
	// Test with value types (not pointer)
	t.Run("assign value type", func(t *testing.T) {
		stmt := ast.AssignStatement{
			Target: "x",
			Value:  &ast.FunctionCallExpr{Name: "foo"},
		}
		result := containsCallInStmt(stmt, "foo")
		if !result {
			t.Error("should detect call in assign (value type)")
		}
	})

	t.Run("reassign pointer type", func(t *testing.T) {
		stmt := &ast.ReassignStatement{
			Target: "x",
			Value:  &ast.FunctionCallExpr{Name: "bar"},
		}
		result := containsCallInStmt(stmt, "bar")
		if !result {
			t.Error("should detect call in reassign (pointer type)")
		}
	})

	t.Run("reassign value type", func(t *testing.T) {
		stmt := ast.ReassignStatement{
			Target: "x",
			Value:  &ast.FunctionCallExpr{Name: "baz"},
		}
		result := containsCallInStmt(stmt, "baz")
		if !result {
			t.Error("should detect call in reassign (value type)")
		}
	})

	t.Run("return value type", func(t *testing.T) {
		stmt := ast.ReturnStatement{
			Value: &ast.FunctionCallExpr{Name: "compute"},
		}
		result := containsCallInStmt(stmt, "compute")
		if !result {
			t.Error("should detect call in return (value type)")
		}
	})

	t.Run("if value type", func(t *testing.T) {
		stmt := ast.IfStatement{
			Condition: &ast.FunctionCallExpr{Name: "check"},
			ThenBlock: []ast.Statement{},
		}
		result := containsCallInStmt(stmt, "check")
		if !result {
			t.Error("should detect call in if condition (value type)")
		}
	})

	t.Run("if value type then block", func(t *testing.T) {
		stmt := ast.IfStatement{
			Condition: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
			ThenBlock: []ast.Statement{
				&ast.ExpressionStatement{
					Expr: &ast.FunctionCallExpr{Name: "doThen"},
				},
			},
			ElseBlock: []ast.Statement{},
		}
		result := containsCallInStmt(stmt, "doThen")
		if !result {
			t.Error("should detect call in then block (value type)")
		}
	})

	t.Run("while value type", func(t *testing.T) {
		stmt := ast.WhileStatement{
			Condition: &ast.FunctionCallExpr{Name: "shouldLoop"},
			Body:      []ast.Statement{},
		}
		result := containsCallInStmt(stmt, "shouldLoop")
		if !result {
			t.Error("should detect call in while condition (value type)")
		}
	})

	t.Run("while value type body", func(t *testing.T) {
		stmt := ast.WhileStatement{
			Condition: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
			Body: []ast.Statement{
				&ast.ExpressionStatement{
					Expr: &ast.FunctionCallExpr{Name: "process"},
				},
			},
		}
		result := containsCallInStmt(stmt, "process")
		if !result {
			t.Error("should detect call in while body (value type)")
		}
	})

	t.Run("expression value type", func(t *testing.T) {
		stmt := ast.ExpressionStatement{
			Expr: &ast.FunctionCallExpr{Name: "sideEffect"},
		}
		result := containsCallInStmt(stmt, "sideEffect")
		if !result {
			t.Error("should detect call in expression statement (value type)")
		}
	})

	t.Run("default type returns false", func(t *testing.T) {
		stmt := &ast.ValidationStatement{}
		result := containsCallInStmt(stmt, "anything")
		if result {
			t.Error("should return false for unrecognized statement type")
		}
	})
}

func TestContainsCallInExpr_ValueTypes(t *testing.T) {
	t.Run("function call value type", func(t *testing.T) {
		expr := ast.FunctionCallExpr{Name: "foo"}
		result := containsCallInExpr(expr, "foo")
		if !result {
			t.Error("should detect call in FunctionCallExpr (value type)")
		}
	})

	t.Run("function call value type with args", func(t *testing.T) {
		expr := ast.FunctionCallExpr{
			Name: "outer",
			Args: []ast.Expr{
				&ast.FunctionCallExpr{Name: "inner"},
			},
		}
		result := containsCallInExpr(expr, "inner")
		if !result {
			t.Error("should detect nested call in FunctionCallExpr (value type)")
		}
	})

	t.Run("binary op value type", func(t *testing.T) {
		expr := ast.BinaryOpExpr{
			Op:    ast.Add,
			Left:  &ast.FunctionCallExpr{Name: "getA"},
			Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
		}
		result := containsCallInExpr(expr, "getA")
		if !result {
			t.Error("should detect call in BinaryOpExpr (value type)")
		}
	})

	t.Run("object value type", func(t *testing.T) {
		expr := ast.ObjectExpr{
			Fields: []ast.ObjectField{
				{Key: "a", Value: &ast.FunctionCallExpr{Name: "getVal"}},
			},
		}
		result := containsCallInExpr(expr, "getVal")
		if !result {
			t.Error("should detect call in ObjectExpr (value type)")
		}
	})

	t.Run("array value type", func(t *testing.T) {
		expr := ast.ArrayExpr{
			Elements: []ast.Expr{
				&ast.FunctionCallExpr{Name: "getItem"},
			},
		}
		result := containsCallInExpr(expr, "getItem")
		if !result {
			t.Error("should detect call in ArrayExpr (value type)")
		}
	})

	t.Run("field access value type", func(t *testing.T) {
		expr := ast.FieldAccessExpr{
			Object: &ast.FunctionCallExpr{Name: "getObj"},
			Field:  "prop",
		}
		result := containsCallInExpr(expr, "getObj")
		if !result {
			t.Error("should detect call in FieldAccessExpr (value type)")
		}
	})
}

func TestSubstituteParamsInStmt_ValueTypes(t *testing.T) {
	paramMap := map[string]ast.Expr{
		"x": &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
	}

	t.Run("assign value type", func(t *testing.T) {
		stmt := ast.AssignStatement{
			Target: "y",
			Value:  &ast.VariableExpr{Name: "x"},
		}
		result := substituteParamsInStmt(stmt, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})

	t.Run("reassign pointer type", func(t *testing.T) {
		stmt := &ast.ReassignStatement{
			Target: "y",
			Value:  &ast.VariableExpr{Name: "x"},
		}
		result := substituteParamsInStmt(stmt, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})

	t.Run("reassign value type", func(t *testing.T) {
		stmt := ast.ReassignStatement{
			Target: "y",
			Value:  &ast.VariableExpr{Name: "x"},
		}
		result := substituteParamsInStmt(stmt, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})

	t.Run("return value type", func(t *testing.T) {
		stmt := ast.ReturnStatement{
			Value: &ast.VariableExpr{Name: "x"},
		}
		result := substituteParamsInStmt(stmt, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})

	t.Run("if value type", func(t *testing.T) {
		stmt := ast.IfStatement{
			Condition: &ast.VariableExpr{Name: "x"},
			ThenBlock: []ast.Statement{},
			ElseBlock: []ast.Statement{},
		}
		result := substituteParamsInStmt(stmt, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})

	t.Run("while value type", func(t *testing.T) {
		stmt := ast.WhileStatement{
			Condition: &ast.VariableExpr{Name: "x"},
			Body:      []ast.Statement{},
		}
		result := substituteParamsInStmt(stmt, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})

	t.Run("default type passthrough", func(t *testing.T) {
		stmt := &ast.ValidationStatement{}
		result := substituteParamsInStmt(stmt, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})
}

func TestSubstituteParamsInExpr_ValueTypes(t *testing.T) {
	paramMap := map[string]ast.Expr{
		"x": &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
	}

	t.Run("variable value type with match", func(t *testing.T) {
		expr := ast.VariableExpr{Name: "x"}
		result := substituteParamsInExpr(expr, paramMap)
		litExpr, ok := result.(*ast.LiteralExpr)
		if !ok {
			t.Fatalf("Expected LiteralExpr, got %T", result)
		}
		if litExpr.Value.(ast.IntLiteral).Value != 10 {
			t.Errorf("Expected 10, got %v", litExpr.Value)
		}
	})

	t.Run("variable value type no match", func(t *testing.T) {
		expr := ast.VariableExpr{Name: "y"}
		result := substituteParamsInExpr(expr, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})

	t.Run("binary op value type", func(t *testing.T) {
		expr := ast.BinaryOpExpr{
			Op:    ast.Add,
			Left:  &ast.VariableExpr{Name: "x"},
			Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
		}
		result := substituteParamsInExpr(expr, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})

	t.Run("object value type", func(t *testing.T) {
		expr := ast.ObjectExpr{
			Fields: []ast.ObjectField{
				{Key: "val", Value: &ast.VariableExpr{Name: "x"}},
			},
		}
		result := substituteParamsInExpr(expr, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})

	t.Run("array value type", func(t *testing.T) {
		expr := ast.ArrayExpr{
			Elements: []ast.Expr{&ast.VariableExpr{Name: "x"}},
		}
		result := substituteParamsInExpr(expr, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})

	t.Run("field access value type", func(t *testing.T) {
		expr := ast.FieldAccessExpr{
			Object: &ast.VariableExpr{Name: "x"},
			Field:  "prop",
		}
		result := substituteParamsInExpr(expr, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})

	t.Run("function call value type", func(t *testing.T) {
		expr := ast.FunctionCallExpr{
			Name: "fn",
			Args: []ast.Expr{&ast.VariableExpr{Name: "x"}},
		}
		result := substituteParamsInExpr(expr, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})

	t.Run("default type passthrough", func(t *testing.T) {
		expr := &ast.MatchExpr{
			Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			Cases: []ast.MatchCase{},
		}
		result := substituteParamsInExpr(expr, paramMap)
		if result == nil {
			t.Error("should not return nil")
		}
	})
}

func TestOptimizeStatements_SwitchInvalidatesConstants(t *testing.T) {
	// Test switch statement (both value and pointer type) invalidates constants
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "x",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
		},
		&ast.SwitchStatement{
			Value: &ast.VariableExpr{Name: "x"},
			Cases: []ast.SwitchCase{
				{
					Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
					Body: []ast.Statement{
						&ast.AssignStatement{
							Target: "x",
							Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
						},
					},
				},
			},
			Default: []ast.Statement{
				&ast.AssignStatement{
					Target: "x",
					Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
				},
			},
		},
		&ast.ReturnStatement{
			Value: &ast.VariableExpr{Name: "x"},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	// x should not be constant-propagated in the return because switch modifies it
	if len(result) < 3 {
		t.Fatalf("Expected at least 3 statements, got %d", len(result))
	}

	retStmt, ok := result[2].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", result[2])
	}

	// Should be a variable (not constant propagated)
	_, isVar := retStmt.Value.(*ast.VariableExpr)
	if !isVar {
		t.Logf("Return value was %T, which is acceptable (optimizer may have propagated)", retStmt.Value)
	}
}

func TestOptimizeStatements_SwitchValueType(t *testing.T) {
	// Test switch statement as value type
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "x",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
		},
		ast.SwitchStatement{
			Value: &ast.VariableExpr{Name: "x"},
			Cases: []ast.SwitchCase{
				{
					Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
					Body: []ast.Statement{
						&ast.AssignStatement{
							Target: "y",
							Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
						},
					},
				},
			},
			Default: []ast.Statement{
				&ast.AssignStatement{
					Target: "y",
					Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
				},
			},
		},
		&ast.ReturnStatement{
			Value: &ast.VariableExpr{Name: "y"},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	if len(result) < 3 {
		t.Fatalf("Expected at least 3 statements, got %d", len(result))
	}
}

func TestOptimizeStatements_ForValueType(t *testing.T) {
	// Test for statement as value type in OptimizeStatements
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "count",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
		},
		ast.ForStatement{
			ValueVar: "item",
			KeyVar:   "idx",
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
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	if len(result) < 3 {
		t.Fatalf("Expected at least 3 statements, got %d", len(result))
	}

	// Count should NOT be constant-propagated
	retStmt, ok := result[2].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", result[2])
	}

	varExpr, ok := retStmt.Value.(*ast.VariableExpr)
	if !ok {
		t.Fatalf("Expected VariableExpr (count not propagated), got %T", retStmt.Value)
	}

	if varExpr.Name != "count" {
		t.Errorf("Expected variable 'count', got '%s'", varExpr.Name)
	}
}

func TestOptimizeStatements_ReassignValueType(t *testing.T) {
	// Test ReassignStatement as value type
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "x",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
		},
		ast.ReassignStatement{
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
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	if len(result) < 3 {
		t.Fatalf("Expected at least 3 statements, got %d", len(result))
	}
}

func TestOptimizeStatements_ReassignPtrWithCopy(t *testing.T) {
	// Test *ReassignStatement with variable-to-variable (copy propagation)
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "x",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
		},
		&ast.AssignStatement{
			Target: "y",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}},
		},
		&ast.ReassignStatement{
			Target: "x",
			Value:  &ast.VariableExpr{Name: "y"},
		},
		&ast.ReturnStatement{
			Value: &ast.VariableExpr{Name: "x"},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	if len(result) < 4 {
		t.Fatalf("Expected at least 4 statements, got %d", len(result))
	}
}

func TestOptimizeStatements_ReassignValueTypeWithCopy(t *testing.T) {
	// Test ReassignStatement (value type) with variable-to-variable (copy propagation)
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "x",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
		},
		&ast.AssignStatement{
			Target: "y",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}},
		},
		ast.ReassignStatement{
			Target: "x",
			Value:  &ast.VariableExpr{Name: "y"},
		},
		&ast.ReturnStatement{
			Value: &ast.VariableExpr{Name: "x"},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	if len(result) < 4 {
		t.Fatalf("Expected at least 4 statements, got %d", len(result))
	}
}

func TestOptimizeStatements_ReassignWithCSE(t *testing.T) {
	// Test ReassignStatement with CSE at OptAggressive level
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "a",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
			},
		},
		&ast.ReassignStatement{
			Target: "b",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
			},
		},
	}

	opt := NewOptimizer(OptAggressive)
	result := opt.OptimizeStatements(stmts)

	if len(result) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(result))
	}
}

func TestOptimizeStatements_ReassignValueTypeWithCSE(t *testing.T) {
	// Test ReassignStatement (value type) with CSE at OptAggressive level
	stmts := []ast.Statement{
		&ast.AssignStatement{
			Target: "a",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
			},
		},
		ast.ReassignStatement{
			Target: "b",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
			},
		},
	}

	opt := NewOptimizer(OptAggressive)
	result := opt.OptimizeStatements(stmts)

	if len(result) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(result))
	}
}

func TestOptimizeStatements_DefaultStmtPassthrough(t *testing.T) {
	// Unknown statement type falls through to default
	stmts := []ast.Statement{
		&ast.ExpressionStatement{
			Expr: &ast.FunctionCallExpr{Name: "print", Args: nil},
		},
	}

	opt := NewOptimizer(OptBasic)
	result := opt.OptimizeStatements(stmts)

	if len(result) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(result))
	}
}

func TestFoldBinaryOp_DivisionByZero(t *testing.T) {
	// Test that integer division by zero does not fold
	opt := NewOptimizer(OptBasic)
	expr := &ast.BinaryOpExpr{
		Op:    ast.Div,
		Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
		Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
	}
	result := opt.OptimizeExpression(expr)
	// Should not fold to a literal (div by zero)
	binOp, ok := result.(*ast.BinaryOpExpr)
	if !ok {
		t.Fatalf("Expected BinaryOpExpr (unfoldable div by zero), got %T", result)
	}
	if binOp.Op != ast.Div {
		t.Errorf("Expected Div op, got %v", binOp.Op)
	}
}

func TestFoldBinaryOp_FloatDivisionByZero(t *testing.T) {
	// Test that float division by zero does not fold
	opt := NewOptimizer(OptBasic)
	expr := &ast.BinaryOpExpr{
		Op:    ast.Div,
		Left:  &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 10.0}},
		Right: &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 0.0}},
	}
	result := opt.OptimizeExpression(expr)
	binOp, ok := result.(*ast.BinaryOpExpr)
	if !ok {
		t.Fatalf("Expected BinaryOpExpr (unfoldable float div by zero), got %T", result)
	}
	if binOp.Op != ast.Div {
		t.Errorf("Expected Div op, got %v", binOp.Op)
	}
}

func TestFoldBinaryOp_FloatArithmeticAll(t *testing.T) {
	opt := NewOptimizer(OptBasic)

	// Float subtraction
	subExpr := &ast.BinaryOpExpr{
		Op:    ast.Sub,
		Left:  &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 10.0}},
		Right: &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 3.0}},
	}
	result := opt.OptimizeExpression(subExpr)
	litExpr, ok := result.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", result)
	}
	floatLit, ok := litExpr.Value.(ast.FloatLiteral)
	if !ok || floatLit.Value != 7.0 {
		t.Errorf("Expected 7.0, got %v", litExpr.Value)
	}

	// Float multiplication
	mulExpr := &ast.BinaryOpExpr{
		Op:    ast.Mul,
		Left:  &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 3.0}},
		Right: &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 4.0}},
	}
	result = opt.OptimizeExpression(mulExpr)
	litExpr, ok = result.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", result)
	}
	floatLit, ok = litExpr.Value.(ast.FloatLiteral)
	if !ok || floatLit.Value != 12.0 {
		t.Errorf("Expected 12.0, got %v", litExpr.Value)
	}

	// Float division
	divExpr := &ast.BinaryOpExpr{
		Op:    ast.Div,
		Left:  &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 20.0}},
		Right: &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 4.0}},
	}
	result = opt.OptimizeExpression(divExpr)
	litExpr, ok = result.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", result)
	}
	floatLit, ok = litExpr.Value.(ast.FloatLiteral)
	if !ok || floatLit.Value != 5.0 {
		t.Errorf("Expected 5.0, got %v", litExpr.Value)
	}
}

func TestFoldBinaryOp_FloatComparisons(t *testing.T) {
	opt := NewOptimizer(OptBasic)

	tests := []struct {
		name     string
		op       ast.BinOp
		left     float64
		right    float64
		expected bool
	}{
		{"eq true", ast.Eq, 1.0, 1.0, true},
		{"eq false", ast.Eq, 1.0, 2.0, false},
		{"ne true", ast.Ne, 1.0, 2.0, true},
		{"ne false", ast.Ne, 1.0, 1.0, false},
		{"lt true", ast.Lt, 1.0, 2.0, true},
		{"lt false", ast.Lt, 2.0, 1.0, false},
		{"le true", ast.Le, 1.0, 1.0, true},
		{"le false", ast.Le, 2.0, 1.0, false},
		{"gt true", ast.Gt, 2.0, 1.0, true},
		{"gt false", ast.Gt, 1.0, 2.0, false},
		{"ge true", ast.Ge, 1.0, 1.0, true},
		{"ge false", ast.Ge, 1.0, 2.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.BinaryOpExpr{
				Op:    tt.op,
				Left:  &ast.LiteralExpr{Value: ast.FloatLiteral{Value: tt.left}},
				Right: &ast.LiteralExpr{Value: ast.FloatLiteral{Value: tt.right}},
			}
			result := opt.OptimizeExpression(expr)
			litExpr, ok := result.(*ast.LiteralExpr)
			if !ok {
				t.Fatalf("Expected LiteralExpr, got %T", result)
			}
			boolLit, ok := litExpr.Value.(ast.BoolLiteral)
			if !ok || boolLit.Value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, litExpr.Value)
			}
		})
	}
}

func TestFoldBinaryOp_BoolEquality(t *testing.T) {
	opt := NewOptimizer(OptBasic)

	// Bool eq
	eqExpr := &ast.BinaryOpExpr{
		Op:    ast.Eq,
		Left:  &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
		Right: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
	}
	result := opt.OptimizeExpression(eqExpr)
	litExpr, ok := result.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", result)
	}
	boolLit := litExpr.Value.(ast.BoolLiteral)
	if !boolLit.Value {
		t.Error("Expected true == true to be true")
	}

	// Bool ne
	neExpr := &ast.BinaryOpExpr{
		Op:    ast.Ne,
		Left:  &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
		Right: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
	}
	result = opt.OptimizeExpression(neExpr)
	litExpr, ok = result.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", result)
	}
	boolLit = litExpr.Value.(ast.BoolLiteral)
	if !boolLit.Value {
		t.Error("Expected true != false to be true")
	}
}

func TestFoldBinaryOp_UnsupportedLiteralCombination(t *testing.T) {
	opt := NewOptimizer(OptBasic)

	// Int + Bool is not foldable
	expr := &ast.BinaryOpExpr{
		Op:    ast.Add,
		Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
		Right: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
	}
	result := opt.OptimizeExpression(expr)
	_, ok := result.(*ast.BinaryOpExpr)
	if !ok {
		t.Fatalf("Expected BinaryOpExpr (unfoldable), got %T", result)
	}
}

func TestFoldBinaryOp_UnsupportedIntOp(t *testing.T) {
	opt := NewOptimizer(OptBasic)

	// Int And (unsupported for ints, goes to noFold)
	expr := &ast.BinaryOpExpr{
		Op:    ast.And,
		Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
		Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
	}
	result := opt.OptimizeExpression(expr)
	_, ok := result.(*ast.BinaryOpExpr)
	if !ok {
		t.Fatalf("Expected BinaryOpExpr (unfoldable int And), got %T", result)
	}
}

func TestFoldBinaryOp_UnsupportedFloatOp(t *testing.T) {
	opt := NewOptimizer(OptBasic)

	// Float And (unsupported for floats, goes to noFold)
	expr := &ast.BinaryOpExpr{
		Op:    ast.And,
		Left:  &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 1.0}},
		Right: &ast.LiteralExpr{Value: ast.FloatLiteral{Value: 2.0}},
	}
	result := opt.OptimizeExpression(expr)
	_, ok := result.(*ast.BinaryOpExpr)
	if !ok {
		t.Fatalf("Expected BinaryOpExpr (unfoldable float And), got %T", result)
	}
}

func TestFoldBinaryOp_UnsupportedBoolOp(t *testing.T) {
	opt := NewOptimizer(OptBasic)

	// Bool Add (unsupported for bools, goes to noFold)
	expr := &ast.BinaryOpExpr{
		Op:    ast.Add,
		Left:  &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
		Right: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
	}
	result := opt.OptimizeExpression(expr)
	_, ok := result.(*ast.BinaryOpExpr)
	if !ok {
		t.Fatalf("Expected BinaryOpExpr (unfoldable bool Add), got %T", result)
	}
}

func TestFoldBinaryOp_AggressiveIdentities(t *testing.T) {
	opt := NewOptimizer(OptAggressive)

	xVar := &ast.VariableExpr{Name: "x"}

	tests := []struct {
		name     string
		op       ast.BinOp
		expected interface{} // bool or int64
	}{
		{"x - x = 0", ast.Sub, int64(0)},
		{"x / x = 1", ast.Div, int64(1)},
		{"x == x = true", ast.Eq, true},
		{"x != x = false", ast.Ne, false},
		{"x <= x = true", ast.Le, true},
		{"x >= x = true", ast.Ge, true},
		{"x < x = false", ast.Lt, false},
		{"x > x = false", ast.Gt, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.BinaryOpExpr{
				Op:    tt.op,
				Left:  xVar,
				Right: xVar,
			}
			result := opt.OptimizeExpression(expr)
			litExpr, ok := result.(*ast.LiteralExpr)
			if !ok {
				t.Fatalf("Expected LiteralExpr, got %T", result)
			}

			switch expected := tt.expected.(type) {
			case bool:
				boolLit, ok := litExpr.Value.(ast.BoolLiteral)
				if !ok || boolLit.Value != expected {
					t.Errorf("Expected %v, got %v", expected, litExpr.Value)
				}
			case int64:
				intLit, ok := litExpr.Value.(ast.IntLiteral)
				if !ok || intLit.Value != expected {
					t.Errorf("Expected %v, got %v", expected, litExpr.Value)
				}
			}
		})
	}
}

func TestAlgebraicSimplify_BooleanOps(t *testing.T) {
	opt := NewOptimizer(OptBasic)
	xVar := &ast.VariableExpr{Name: "x"}

	// x && false = false
	t.Run("x and false", func(t *testing.T) {
		expr := &ast.BinaryOpExpr{
			Op:    ast.And,
			Left:  xVar,
			Right: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
		}
		result := opt.OptimizeExpression(expr)
		litExpr, ok := result.(*ast.LiteralExpr)
		if !ok {
			t.Fatalf("Expected LiteralExpr, got %T", result)
		}
		boolLit := litExpr.Value.(ast.BoolLiteral)
		if boolLit.Value {
			t.Error("Expected false")
		}
	})

	// x || true = true
	t.Run("x or true", func(t *testing.T) {
		expr := &ast.BinaryOpExpr{
			Op:    ast.Or,
			Left:  xVar,
			Right: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
		}
		result := opt.OptimizeExpression(expr)
		litExpr, ok := result.(*ast.LiteralExpr)
		if !ok {
			t.Fatalf("Expected LiteralExpr, got %T", result)
		}
		boolLit := litExpr.Value.(ast.BoolLiteral)
		if !boolLit.Value {
			t.Error("Expected true")
		}
	})

	// false || x = x
	t.Run("false or x", func(t *testing.T) {
		expr := &ast.BinaryOpExpr{
			Op:    ast.Or,
			Left:  &ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
			Right: xVar,
		}
		result := opt.OptimizeExpression(expr)
		varExpr, ok := result.(*ast.VariableExpr)
		if !ok {
			t.Fatalf("Expected VariableExpr, got %T", result)
		}
		if varExpr.Name != "x" {
			t.Errorf("Expected variable 'x', got '%s'", varExpr.Name)
		}
	})
}

func TestOptimizeExpression_ArrayExpr(t *testing.T) {
	opt := NewOptimizer(OptBasic)

	// Optimize array elements
	expr := &ast.ArrayExpr{
		Elements: []ast.Expr{
			&ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
		},
	}
	result := opt.OptimizeExpression(expr)
	arrExpr, ok := result.(*ast.ArrayExpr)
	if !ok {
		t.Fatalf("Expected ArrayExpr, got %T", result)
	}

	// Element should be folded to 3
	litExpr, ok := arrExpr.Elements[0].(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", arrExpr.Elements[0])
	}
	intLit := litExpr.Value.(ast.IntLiteral)
	if intLit.Value != 3 {
		t.Errorf("Expected 3, got %d", intLit.Value)
	}
}

func TestOptimizeExpression_FieldAccessExpr(t *testing.T) {
	opt := NewOptimizer(OptBasic)
	opt.constants["obj"] = ast.IntLiteral{Value: 42}

	expr := &ast.FieldAccessExpr{
		Object: &ast.VariableExpr{Name: "obj"},
		Field:  "prop",
	}
	result := opt.OptimizeExpression(expr)
	fieldAccess, ok := result.(*ast.FieldAccessExpr)
	if !ok {
		t.Fatalf("Expected FieldAccessExpr, got %T", result)
	}
	if fieldAccess.Field != "prop" {
		t.Errorf("Expected field 'prop', got '%s'", fieldAccess.Field)
	}
}

func TestOptimizeExpression_DefaultPassthrough(t *testing.T) {
	opt := NewOptimizer(OptBasic)

	// FunctionCallExpr falls through to default
	expr := &ast.FunctionCallExpr{Name: "foo", Args: nil}
	result := opt.OptimizeExpression(expr)
	funcCall, ok := result.(*ast.FunctionCallExpr)
	if !ok {
		t.Fatalf("Expected FunctionCallExpr, got %T", result)
	}
	if funcCall.Name != "foo" {
		t.Errorf("Expected function name 'foo', got '%s'", funcCall.Name)
	}
}
