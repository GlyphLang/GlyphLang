package compiler

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/vm"
)

// TestCompileBreakInWhileLoop tests that break exits a while loop early.
// Equivalent to:
//
//	$x = 0
//	while true {
//	  x = x + 1
//	  if x == 3 { break }
//	}
//	> x   // expect 3
func TestCompileBreakInWhileLoop(t *testing.T) {
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.WhileStatement{
				Condition: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
				Body: []ast.Statement{
					&ast.ReassignStatement{
						Target: "x",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "x"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
					&ast.IfStatement{
						Condition: &ast.BinaryOpExpr{
							Op:    ast.Eq,
							Left:  &ast.VariableExpr{Name: "x"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
						},
						ThenBlock: []ast.Statement{
							&ast.BreakStatement{},
						},
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

// TestCompileContinueInWhileLoop tests that continue skips to the next iteration.
// Equivalent to:
//
//	$count = 0
//	$sum = 0
//	while count < 5 {
//	  count = count + 1
//	  if count == 3 { continue }
//	  sum = sum + count
//	}
//	> sum   // expect 1+2+4+5 = 12
func TestCompileContinueInWhileLoop(t *testing.T) {
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "count",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.AssignStatement{
				Target: "sum",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.WhileStatement{
				Condition: &ast.BinaryOpExpr{
					Op:    ast.Lt,
					Left:  &ast.VariableExpr{Name: "count"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
				},
				Body: []ast.Statement{
					&ast.ReassignStatement{
						Target: "count",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "count"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
					&ast.IfStatement{
						Condition: &ast.BinaryOpExpr{
							Op:    ast.Eq,
							Left:  &ast.VariableExpr{Name: "count"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
						},
						ThenBlock: []ast.Statement{
							&ast.ContinueStatement{},
						},
					},
					&ast.ReassignStatement{
						Target: "sum",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "sum"},
							Right: &ast.VariableExpr{Name: "count"},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "sum"},
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

	expected := vm.IntValue{Val: 12}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestCompileBreakInForLoop tests that break exits a for loop early.
// Equivalent to:
//
//	$result = 0
//	for item in [10, 20, 30, 40, 50] {
//	  if item == 30 { break }
//	  result = result + item
//	}
//	> result   // expect 10+20 = 30
func TestCompileBreakInForLoop(t *testing.T) {
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
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 30}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 40}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 50}},
					},
				},
				Body: []ast.Statement{
					&ast.IfStatement{
						Condition: &ast.BinaryOpExpr{
							Op:    ast.Eq,
							Left:  &ast.VariableExpr{Name: "item"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 30}},
						},
						ThenBlock: []ast.Statement{
							&ast.BreakStatement{},
						},
					},
					&ast.ReassignStatement{
						Target: "result",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "result"},
							Right: &ast.VariableExpr{Name: "item"},
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
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 30}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestCompileContinueInForLoop tests that continue skips an iteration in for loop.
// Equivalent to:
//
//	$result = 0
//	for item in [1, 2, 3, 4, 5] {
//	  if item == 3 { continue }
//	  result = result + item
//	}
//	> result   // expect 1+2+4+5 = 12
func TestCompileContinueInForLoop(t *testing.T) {
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
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
					},
				},
				Body: []ast.Statement{
					&ast.IfStatement{
						Condition: &ast.BinaryOpExpr{
							Op:    ast.Eq,
							Left:  &ast.VariableExpr{Name: "item"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
						},
						ThenBlock: []ast.Statement{
							&ast.ContinueStatement{},
						},
					},
					&ast.ReassignStatement{
						Target: "result",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "result"},
							Right: &ast.VariableExpr{Name: "item"},
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
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 12}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestCompileBreakContinueNestedLoops tests that break and continue only
// affect the innermost loop. The outer loop runs 3 times. In the inner
// loop, continue skips item 2 and break exits when item == 4.
// Equivalent to:
//
//	$total = 0
//	$i = 0
//	while i < 3 {
//	  i = i + 1
//	  for item in [1, 2, 3, 4, 5] {
//	    if item == 2 { continue }
//	    if item == 4 { break }
//	    total = total + item
//	  }
//	}
//	> total   // inner loop adds 1+3=4 each outer iteration, 3*4 = 12
func TestCompileBreakContinueNestedLoops(t *testing.T) {
	route := &ast.Route{
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
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
				},
				Body: []ast.Statement{
					&ast.ReassignStatement{
						Target: "i",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "i"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
					&ast.ForStatement{
						ValueVar: "item",
						Iterable: &ast.ArrayExpr{
							Elements: []ast.Expr{
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
								&ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
							},
						},
						Body: []ast.Statement{
							// continue when item == 2
							&ast.IfStatement{
								Condition: &ast.BinaryOpExpr{
									Op:    ast.Eq,
									Left:  &ast.VariableExpr{Name: "item"},
									Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
								},
								ThenBlock: []ast.Statement{
									&ast.ContinueStatement{},
								},
							},
							// break when item == 4
							&ast.IfStatement{
								Condition: &ast.BinaryOpExpr{
									Op:    ast.Eq,
									Left:  &ast.VariableExpr{Name: "item"},
									Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
								},
								ThenBlock: []ast.Statement{
									&ast.BreakStatement{},
								},
							},
							// total = total + item
							&ast.ReassignStatement{
								Target: "total",
								Value: &ast.BinaryOpExpr{
									Op:    ast.Add,
									Left:  &ast.VariableExpr{Name: "total"},
									Right: &ast.VariableExpr{Name: "item"},
								},
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

	// Inner loop: 1 + 3 = 4 (skip 2 via continue, break at 4)
	// Outer loop runs 3 times: 4 * 3 = 12
	expected := vm.IntValue{Val: 12}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestCompileBreakWithSurroundingCode tests that code after a break-containing
// loop executes correctly, verifying that jump targets are patched properly.
// Equivalent to:
//
//	$before = 10
//	$x = 0
//	while true {
//	  x = x + 1
//	  if x == 5 { break }
//	}
//	$after = 20
//	> before + x + after   // expect 10 + 5 + 20 = 35
func TestCompileBreakWithSurroundingCode(t *testing.T) {
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "before",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
			},
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.WhileStatement{
				Condition: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
				Body: []ast.Statement{
					&ast.ReassignStatement{
						Target: "x",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "x"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
					&ast.IfStatement{
						Condition: &ast.BinaryOpExpr{
							Op:    ast.Eq,
							Left:  &ast.VariableExpr{Name: "x"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
						},
						ThenBlock: []ast.Statement{
							&ast.BreakStatement{},
						},
					},
				},
			},
			&ast.AssignStatement{
				Target: "after",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}},
			},
			&ast.ReturnStatement{
				Value: &ast.BinaryOpExpr{
					Op: ast.Add,
					Left: &ast.BinaryOpExpr{
						Op:    ast.Add,
						Left:  &ast.VariableExpr{Name: "before"},
						Right: &ast.VariableExpr{Name: "x"},
					},
					Right: &ast.VariableExpr{Name: "after"},
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

	expected := vm.IntValue{Val: 35}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestCompileImmediateBreakInWhile tests that an immediate break terminates
// the loop without executing any further body statements.
// Equivalent to:
//
//	$x = 0
//	while true { break }
//	> x   // expect 0 (loop body after break is unreachable)
func TestCompileImmediateBreakInWhile(t *testing.T) {
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.WhileStatement{
				Condition: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
				Body: []ast.Statement{
					&ast.BreakStatement{},
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

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 0}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestCompileBreakOutsideLoop tests that break outside of a loop produces
// a semantic error at compile time.
func TestCompileBreakOutsideLoop(t *testing.T) {
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.BreakStatement{},
		},
	}

	c := NewCompilerWithOptLevel(OptNone)
	_, err := c.CompileRoute(route)
	if err == nil {
		t.Fatal("Expected error for break outside loop, got nil")
	}
	if !IsSemanticError(err) {
		t.Errorf("Expected SemanticError, got %T: %v", err, err)
	}
}

// TestCompileContinueOutsideLoop tests that continue outside of a loop produces
// a semantic error at compile time.
func TestCompileContinueOutsideLoop(t *testing.T) {
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ContinueStatement{},
		},
	}

	c := NewCompilerWithOptLevel(OptNone)
	_, err := c.CompileRoute(route)
	if err == nil {
		t.Fatal("Expected error for continue outside loop, got nil")
	}
	if !IsSemanticError(err) {
		t.Errorf("Expected SemanticError, got %T: %v", err, err)
	}
}

// TestCompileMultipleBreaksInWhile tests a loop with multiple break paths.
// Equivalent to:
//
//	$x = 0
//	while true {
//	  x = x + 1
//	  if x == 2 {
//	    if true { break }
//	  }
//	  if x == 10 { break }
//	}
//	> x   // expect 2 (first break is reached)
func TestCompileMultipleBreaksInWhile(t *testing.T) {
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
			},
			&ast.WhileStatement{
				Condition: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
				Body: []ast.Statement{
					&ast.ReassignStatement{
						Target: "x",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "x"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
					&ast.IfStatement{
						Condition: &ast.BinaryOpExpr{
							Op:    ast.Eq,
							Left:  &ast.VariableExpr{Name: "x"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						},
						ThenBlock: []ast.Statement{
							&ast.IfStatement{
								Condition: &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
								ThenBlock: []ast.Statement{
									&ast.BreakStatement{},
								},
							},
						},
					},
					&ast.IfStatement{
						Condition: &ast.BinaryOpExpr{
							Op:    ast.Eq,
							Left:  &ast.VariableExpr{Name: "x"},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
						},
						ThenBlock: []ast.Statement{
							&ast.BreakStatement{},
						},
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

// TestCompileBreakContinueNormalization tests that BreakStatement and
// ContinueStatement pointer types are correctly normalized to value types
// by normalizeStatement, matching the pattern used for all other statement types.
func TestCompileBreakContinueNormalization(t *testing.T) {
	// Test pointer-to-value normalization for BreakStatement
	breakPtr := &ast.BreakStatement{}
	normalized := normalizeStatement(breakPtr)
	if _, ok := normalized.(ast.BreakStatement); !ok {
		t.Errorf("normalizeStatement(*BreakStatement) should return BreakStatement value, got %T", normalized)
	}

	// Test pointer-to-value normalization for ContinueStatement
	continuePtr := &ast.ContinueStatement{}
	normalized = normalizeStatement(continuePtr)
	if _, ok := normalized.(ast.ContinueStatement); !ok {
		t.Errorf("normalizeStatement(*ContinueStatement) should return ContinueStatement value, got %T", normalized)
	}
}

// TestCompileContinueInForLoopSkipsMultiple tests continue with multiple skipped values.
// Equivalent to:
//
//	$result = 0
//	for item in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10] {
//	  if item % 2 == 0 { continue }  // skip even numbers
//	  result = result + item
//	}
//	> result   // expect 1+3+5+7+9 = 25
func TestCompileContinueInForLoopSkipsMultiple(t *testing.T) {
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
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 6}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 7}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 8}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 9}},
						&ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
					},
				},
				Body: []ast.Statement{
					// if item % 2 == 0 { continue }
					&ast.IfStatement{
						Condition: &ast.BinaryOpExpr{
							Op: ast.Eq,
							Left: &ast.BinaryOpExpr{
								Op:    ast.Mod,
								Left:  &ast.VariableExpr{Name: "item"},
								Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
							},
							Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
						},
						ThenBlock: []ast.Statement{
							&ast.ContinueStatement{},
						},
					},
					&ast.ReassignStatement{
						Target: "result",
						Value: &ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  &ast.VariableExpr{Name: "result"},
							Right: &ast.VariableExpr{Name: "item"},
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
		t.Fatalf("CompileRoute() error: %v", err)
	}

	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.IntValue{Val: 25}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
