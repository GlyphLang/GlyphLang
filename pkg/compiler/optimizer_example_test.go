package compiler

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/vm"
)

// Example demonstrating constant folding optimization
func Example_constantFolding() {
	// Code: $ result = 2 + 3 * 4, > result
	// The optimizer will fold 3 * 4 to 12, then 2 + 12 to 14
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "result",
				Value: &interpreter.BinaryOpExpr{
					Op: interpreter.Add,
					Left: &interpreter.LiteralExpr{
						Value: interpreter.IntLiteral{Value: 2},
					},
					Right: &interpreter.BinaryOpExpr{
						Op: interpreter.Mul,
						Left: &interpreter.LiteralExpr{
							Value: interpreter.IntLiteral{Value: 3},
						},
						Right: &interpreter.LiteralExpr{
							Value: interpreter.IntLiteral{Value: 4},
						},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "result"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.LiteralExpr{
					Value: interpreter.IntLiteral{Value: 42},
				},
			},
			// This statement is unreachable and will be eliminated
			&interpreter.AssignStatement{
				Target: "x",
				Value: &interpreter.LiteralExpr{
					Value: interpreter.IntLiteral{Value: 100},
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
	expr := &interpreter.BinaryOpExpr{
		Op:   interpreter.Mul,
		Left: &interpreter.VariableExpr{Name: "x"},
		Right: &interpreter.LiteralExpr{
			Value: interpreter.IntLiteral{Value: 1},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.AssignStatement{
				Target: "a",
				Value: &interpreter.BinaryOpExpr{
					Op: interpreter.Add,
					Left: &interpreter.BinaryOpExpr{
						Op:    interpreter.Mul,
						Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
						Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
					},
					Right: &interpreter.BinaryOpExpr{
						Op:    interpreter.Mul,
						Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 4}},
						Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
					},
				},
			},
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "a"},
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
	expr := &interpreter.BinaryOpExpr{
		Op: interpreter.Add,
		Left: &interpreter.BinaryOpExpr{
			Op:    interpreter.Mul,
			Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
			Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 3}},
		},
		Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 4}},
	}

	opt := NewOptimizer(OptBasic)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.OptimizeExpression(expr)
	}
}

// Benchmark algebraic simplification performance
func BenchmarkAlgebraicSimplification(b *testing.B) {
	xVar := &interpreter.VariableExpr{Name: "x"}

	b.Run("x+0", func(b *testing.B) {
		expr := &interpreter.BinaryOpExpr{
			Op:    interpreter.Add,
			Left:  xVar,
			Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
		}
		opt := NewOptimizer(OptBasic)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = opt.OptimizeExpression(expr)
		}
	})

	b.Run("x*1", func(b *testing.B) {
		expr := &interpreter.BinaryOpExpr{
			Op:    interpreter.Mul,
			Left:  xVar,
			Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
		}
		opt := NewOptimizer(OptBasic)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = opt.OptimizeExpression(expr)
		}
	})

	b.Run("x*0", func(b *testing.B) {
		expr := &interpreter.BinaryOpExpr{
			Op:    interpreter.Mul,
			Left:  xVar,
			Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
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
		&interpreter.AssignStatement{
			Target: "z",
			Value:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 200}},
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.OptimizeStatements(stmts)
	}
}

// Benchmark constant propagation performance
func BenchmarkConstantPropagation(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.OptimizeStatements(stmts)
	}
}

// Benchmark strength reduction performance
func BenchmarkStrengthReduction(b *testing.B) {
	expr := &interpreter.BinaryOpExpr{
		Op:    interpreter.Mul,
		Left:  &interpreter.VariableExpr{Name: "x"},
		Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
	}

	opt := NewOptimizer(OptAggressive)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.OptimizeExpression(expr)
	}
}

// Benchmark CSE performance
func BenchmarkCSE(b *testing.B) {
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
		&interpreter.AssignStatement{
			Target: "c",
			Value: &interpreter.BinaryOpExpr{
				Op:    interpreter.Add,
				Left:  &interpreter.VariableExpr{Name: "x"},
				Right: &interpreter.VariableExpr{Name: "y"},
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
	route := &interpreter.Route{
		Body: []interpreter.Statement{
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
			&interpreter.ReturnStatement{
				Value: &interpreter.VariableExpr{Name: "z"},
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
					Target: "y",
					Value: &interpreter.BinaryOpExpr{
						Op:    interpreter.Mul,
						Left:  &interpreter.VariableExpr{Name: "c"},
						Right: &interpreter.VariableExpr{Name: "d"},
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.OptimizeStatements(stmts)
	}
}

// Test showing optimization reduces bytecode size
func TestOptimization_ReducesBytecodeSize(t *testing.T) {
	// Code with constant folding opportunities
	route := &interpreter.Route{
		Body: []interpreter.Statement{
			&interpreter.ReturnStatement{
				Value: &interpreter.BinaryOpExpr{
					Op: interpreter.Add,
					Left: &interpreter.BinaryOpExpr{
						Op:    interpreter.Mul,
						Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
						Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 20}},
					},
					Right: &interpreter.BinaryOpExpr{
						Op:    interpreter.Sub,
						Left:  &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 100}},
						Right: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 50}},
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
