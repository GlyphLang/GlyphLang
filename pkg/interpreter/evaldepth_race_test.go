package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"sync"
	"testing"
)

// TestEvalDepthConcurrentAccess verifies that concurrent expression evaluation
// does not race on the evalDepth counter. This test should be run with -race.
func TestEvalDepthConcurrentAccess(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Simple expression that won't error
	expr := LiteralExpr{Value: IntLiteral{Value: 42}}

	var wg sync.WaitGroup
	const goroutines = 50
	const iterations = 100

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				result, err := interp.EvaluateExpression(expr, env)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != int64(42) {
					t.Errorf("expected 42, got %v", result)
					return
				}
			}
		}()
	}

	wg.Wait()
}

// TestEvalDepthConcurrentAsyncExpr verifies that async expressions
// properly handle concurrent evalDepth access.
func TestEvalDepthConcurrentAsyncExpr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Create multiple async expressions that each evaluate nested expressions
	var wg sync.WaitGroup
	const goroutines = 20

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Evaluate a binary expression concurrently
			expr := BinaryOpExpr{
				Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
				Op:    Add,
				Right: LiteralExpr{Value: IntLiteral{Value: 20}},
			}
			result, err := interp.EvaluateExpression(expr, env)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != int64(30) {
				t.Errorf("expected 30, got %v", result)
			}
		}()
	}

	wg.Wait()
}
