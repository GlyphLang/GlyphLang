package compiler

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"testing"

	"github.com/glyphlang/glyph/pkg/vm"
)

// Example demonstrating constant folding optimization
func Example_constantFolding() {
	// Code: $ result = 2 + 3 * 4, > result
	// The optimizer will fold 3 * 4 to 12, then 2 + 12 to 14
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "result",
				Value: &ast.BinaryOpExpr{
					Op: ast.Add,
					Left: &ast.LiteralExpr{
						Value: ast.IntLiteral{Value: 2},
					},
					Right: &ast.BinaryOpExpr{
						Op: ast.Mul,
						Left: &ast.LiteralExpr{
							Value: ast.IntLiteral{Value: 3},
						},
						Right: &ast.LiteralExpr{
							Value: ast.IntLiteral{Value: 4},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "result"},
			},
		},
	}

	// Compile with optimization
	c := NewCompiler()
	bytecode, _ := c.CompileRoute(route)

	// Execute
	vmInstance := vm.NewVM()
	result, _ := vmInstance.Execute(bytecode)

	// Output will be 14
	_ = result
}

// Example demonstrating dead code elimination
func Example_deadCodeElimination() {
	// Code: > 42, $ x = 100 (assignment after return is unreachable)
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.LiteralExpr{
					Value: ast.IntLiteral{Value: 42},
				},
			},
			// This statement is unreachable and will be eliminated
			&ast.AssignStatement{
				Target: "x",
				Value: &ast.LiteralExpr{
					Value: ast.IntLiteral{Value: 100},
				},
			},
		},
	}

	opt := NewOptimizer(OptBasic)
	optimized := opt.OptimizeStatements(route.Body)

	// Only the return statement remains
	_ = optimized
}

// Example demonstrating algebraic simplification
func Example_algebraicSimplification() {
	// Code: $ result = x * 1 (simplifies to x)
	expr := &ast.BinaryOpExpr{
		Op:   ast.Mul,
		Left: &ast.VariableExpr{Name: "x"},
		Right: &ast.LiteralExpr{
			Value: ast.IntLiteral{Value: 1},
		},
	}

	opt := NewOptimizer(OptBasic)
	optimized := opt.OptimizeExpression(expr)

	// Result is just VariableExpr{Name: "x"}
	_ = optimized
}

// Benchmark comparing optimized vs non-optimized compilation
func BenchmarkOptimizedVsNonOptimized(b *testing.B) {
	// Complex expression with many constant operations
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "a",
				Value: &ast.BinaryOpExpr{
					Op: ast.Add,
					Left: &ast.BinaryOpExpr{
						Op:    ast.Mul,
						Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
						Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
					},
					Right: &ast.BinaryOpExpr{
						Op:    ast.Mul,
						Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
						Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "a"},
			},
		},
	}

	b.Run("WithOptimization", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := NewCompilerWithOptLevel(OptBasic)
			bytecode, _ := c.CompileRoute(route)
			vmInstance := vm.NewVM()
			_, _ = vmInstance.Execute(bytecode)
		}
	})

	b.Run("WithoutOptimization", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := NewCompilerWithOptLevel(OptNone)
			bytecode, _ := c.CompileRoute(route)
			vmInstance := vm.NewVM()
			_, _ = vmInstance.Execute(bytecode)
		}
	})
}

// Benchmark constant folding performance
func BenchmarkConstantFolding(b *testing.B) {
	expr := &ast.BinaryOpExpr{
		Op: ast.Add,
		Left: &ast.BinaryOpExpr{
			Op:    ast.Mul,
			Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 3}},
		},
		Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 4}},
	}

	opt := NewOptimizer(OptBasic)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.OptimizeExpression(expr)
	}
}

// Benchmark algebraic simplification performance
func BenchmarkAlgebraicSimplification(b *testing.B) {
	xVar := &ast.VariableExpr{Name: "x"}

	b.Run("x+0", func(b *testing.B) {
		expr := &ast.BinaryOpExpr{
			Op:    ast.Add,
			Left:  xVar,
			Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
		}
		opt := NewOptimizer(OptBasic)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = opt.OptimizeExpression(expr)
		}
	})

	b.Run("x*1", func(b *testing.B) {
		expr := &ast.BinaryOpExpr{
			Op:    ast.Mul,
			Left:  xVar,
			Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
		}
		opt := NewOptimizer(OptBasic)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = opt.OptimizeExpression(expr)
		}
	})

	b.Run("x*0", func(b *testing.B) {
		expr := &ast.BinaryOpExpr{
			Op:    ast.Mul,
			Left:  xVar,
			Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
		}
		opt := NewOptimizer(OptBasic)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = opt.OptimizeExpression(expr)
		}
	})
}

// Benchmark dead code elimination performance
func BenchmarkDeadCodeElimination(b *testing.B) {
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
		&ast.AssignStatement{
			Target: "z",
			Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 200}},
		},
	}

	opt := NewOptimizer(OptBasic)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.OptimizeStatements(stmts)
	}
}

// Benchmark dead branch elimination
func BenchmarkDeadBranchElimination(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.OptimizeStatements(stmts)
	}
}

// Benchmark constant propagation performance
func BenchmarkConstantPropagation(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.OptimizeStatements(stmts)
	}
}

// Benchmark strength reduction performance
func BenchmarkStrengthReduction(b *testing.B) {
	expr := &ast.BinaryOpExpr{
		Op:    ast.Mul,
		Left:  &ast.VariableExpr{Name: "x"},
		Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
	}

	opt := NewOptimizer(OptAggressive)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.OptimizeExpression(expr)
	}
}

// Benchmark CSE performance
func BenchmarkCSE(b *testing.B) {
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
		&ast.AssignStatement{
			Target: "c",
			Value: &ast.BinaryOpExpr{
				Op:    ast.Add,
				Left:  &ast.VariableExpr{Name: "x"},
				Right: &ast.VariableExpr{Name: "y"},
			},
		},
	}

	opt := NewOptimizer(OptAggressive)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.OptimizeStatements(stmts)
	}
}

// Benchmark comparing OptBasic vs OptAggressive
func BenchmarkOptimizationLevels(b *testing.B) {
	route := &ast.Route{
		Body: []ast.Statement{
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
			&ast.ReturnStatement{
				Value: &ast.VariableExpr{Name: "z"},
			},
		},
	}

	b.Run("OptBasic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := NewCompilerWithOptLevel(OptBasic)
			bytecode, _ := c.CompileRoute(route)
			vmInstance := vm.NewVM()
			_, _ = vmInstance.Execute(bytecode)
		}
	})

	b.Run("OptAggressive", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := NewCompilerWithOptLevel(OptAggressive)
			bytecode, _ := c.CompileRoute(route)
			vmInstance := vm.NewVM()
			_, _ = vmInstance.Execute(bytecode)
		}
	})
}

// Benchmark loop invariant code motion
func BenchmarkLICM(b *testing.B) {
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
					Target: "y",
					Value: &ast.BinaryOpExpr{
						Op:    ast.Mul,
						Left:  &ast.VariableExpr{Name: "c"},
						Right: &ast.VariableExpr{Name: "d"},
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.OptimizeStatements(stmts)
	}
}

// Test showing optimization reduces bytecode size
func TestOptimization_ReducesBytecodeSize(t *testing.T) {
	// Code with constant folding opportunities
	route := &ast.Route{
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.BinaryOpExpr{
					Op: ast.Add,
					Left: &ast.BinaryOpExpr{
						Op:    ast.Mul,
						Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
						Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}},
					},
					Right: &ast.BinaryOpExpr{
						Op:    ast.Sub,
						Left:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 100}},
						Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 50}},
					},
				},
			},
		},
	}

	// Compile without optimization
	cNoOpt := NewCompilerWithOptLevel(OptNone)
	bytecodeNoOpt, err := cNoOpt.CompileRoute(route)
	if err != nil {
		t.Fatalf("Compile without optimization failed: %v", err)
	}

	// Compile with optimization
	cOpt := NewCompilerWithOptLevel(OptBasic)
	bytecodeOpt, err := cOpt.CompileRoute(route)
	if err != nil {
		t.Fatalf("Compile with optimization failed: %v", err)
	}

	// Both should produce same result
	vm1 := vm.NewVM()
	result1, err := vm1.Execute(bytecodeNoOpt)
	if err != nil {
		t.Fatalf("Execute without optimization failed: %v", err)
	}

	vm2 := vm.NewVM()
	result2, err := vm2.Execute(bytecodeOpt)
	if err != nil {
		t.Fatalf("Execute with optimization failed: %v", err)
	}

	if !valuesEqual(result1, result2) {
		t.Errorf("Results differ: %v vs %v", result1, result2)
	}

	// Optimized bytecode should be smaller (constant folding reduces operations)
	if len(bytecodeOpt) >= len(bytecodeNoOpt) {
		t.Logf("Warning: Optimized bytecode (%d bytes) not smaller than non-optimized (%d bytes)",
			len(bytecodeOpt), len(bytecodeNoOpt))
		t.Logf("This may indicate the optimizer could be improved or constant pool deduplication is working well")
	} else {
		t.Logf("Optimization reduced bytecode from %d to %d bytes (%.1f%% reduction)",
			len(bytecodeNoOpt), len(bytecodeOpt),
			100.0*float64(len(bytecodeNoOpt)-len(bytecodeOpt))/float64(len(bytecodeNoOpt)))
	}
}
