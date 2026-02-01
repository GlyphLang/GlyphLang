package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create an interpreter with environment access
func newTestInterpreter() *Interpreter {
	interp := NewInterpreter()
	return interp
}

// Test pipe expression evaluation using direct AST construction
func TestInterpreter_PipeExpression(t *testing.T) {
	t.Run("pipe to function with single arg", func(t *testing.T) {
		// Create a simple function: double(x: int) -> x * 2
		doubleFn := Function{
			Name: "double",
			Params: []Field{
				{Name: "x", TypeAnnotation: IntType{}, Required: true},
			},
			ReturnType: IntType{},
			Body: []Statement{
				ReturnStatement{
					Value: BinaryOpExpr{
						Op:    Mul,
						Left:  VariableExpr{Name: "x"},
						Right: LiteralExpr{Value: IntLiteral{Value: 2}},
					},
				},
			},
		}

		// Create: 5 |> double
		pipeExpr := PipeExpr{
			Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
			Right: VariableExpr{Name: "double"},
		}

		interp := newTestInterpreter()
		env := NewEnvironment()
		env.Define("double", doubleFn)

		result, err := interp.EvaluateExpression(pipeExpr, env)
		require.NoError(t, err)
		assert.Equal(t, int64(10), result)
	})

	t.Run("pipe to function call prepends arg", func(t *testing.T) {
		// Create a function: add(a: int, b: int) -> a + b
		addFn := Function{
			Name: "add",
			Params: []Field{
				{Name: "a", TypeAnnotation: IntType{}, Required: true},
				{Name: "b", TypeAnnotation: IntType{}, Required: true},
			},
			ReturnType: IntType{},
			Body: []Statement{
				ReturnStatement{
					Value: BinaryOpExpr{
						Op:    Add,
						Left:  VariableExpr{Name: "a"},
						Right: VariableExpr{Name: "b"},
					},
				},
			},
		}

		// Create: 5 |> add(3) which should become add(5, 3)
		pipeExpr := PipeExpr{
			Left: LiteralExpr{Value: IntLiteral{Value: 5}},
			Right: FunctionCallExpr{
				Name: "add",
				Args: []Expr{LiteralExpr{Value: IntLiteral{Value: 3}}},
			},
		}

		interp := newTestInterpreter()
		env := NewEnvironment()
		env.Define("add", addFn)

		result, err := interp.EvaluateExpression(pipeExpr, env)
		require.NoError(t, err)
		assert.Equal(t, int64(8), result)
	})

	t.Run("chained pipes", func(t *testing.T) {
		// Create: double(x) -> x * 2
		doubleFn := Function{
			Name: "double",
			Params: []Field{
				{Name: "x", TypeAnnotation: IntType{}, Required: true},
			},
			ReturnType: IntType{},
			Body: []Statement{
				ReturnStatement{
					Value: BinaryOpExpr{
						Op:    Mul,
						Left:  VariableExpr{Name: "x"},
						Right: LiteralExpr{Value: IntLiteral{Value: 2}},
					},
				},
			},
		}

		// Create: inc(x) -> x + 1
		incFn := Function{
			Name: "inc",
			Params: []Field{
				{Name: "x", TypeAnnotation: IntType{}, Required: true},
			},
			ReturnType: IntType{},
			Body: []Statement{
				ReturnStatement{
					Value: BinaryOpExpr{
						Op:    Add,
						Left:  VariableExpr{Name: "x"},
						Right: LiteralExpr{Value: IntLiteral{Value: 1}},
					},
				},
			},
		}

		// Create: 5 |> double |> inc (should be (5 |> double) |> inc = 10 |> inc = 11)
		pipeExpr := PipeExpr{
			Left: PipeExpr{
				Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
				Right: VariableExpr{Name: "double"},
			},
			Right: VariableExpr{Name: "inc"},
		}

		interp := newTestInterpreter()
		env := NewEnvironment()
		env.Define("double", doubleFn)
		env.Define("inc", incFn)

		result, err := interp.EvaluateExpression(pipeExpr, env)
		require.NoError(t, err)
		assert.Equal(t, int64(11), result)
	})

	t.Run("pipe with expression on left", func(t *testing.T) {
		// Create: double(x) -> x * 2
		doubleFn := Function{
			Name: "double",
			Params: []Field{
				{Name: "x", TypeAnnotation: IntType{}, Required: true},
			},
			ReturnType: IntType{},
			Body: []Statement{
				ReturnStatement{
					Value: BinaryOpExpr{
						Op:    Mul,
						Left:  VariableExpr{Name: "x"},
						Right: LiteralExpr{Value: IntLiteral{Value: 2}},
					},
				},
			},
		}

		// Create: (2 + 3) |> double = 5 |> double = 10
		pipeExpr := PipeExpr{
			Left: BinaryOpExpr{
				Op:    Add,
				Left:  LiteralExpr{Value: IntLiteral{Value: 2}},
				Right: LiteralExpr{Value: IntLiteral{Value: 3}},
			},
			Right: VariableExpr{Name: "double"},
		}

		interp := newTestInterpreter()
		env := NewEnvironment()
		env.Define("double", doubleFn)

		result, err := interp.EvaluateExpression(pipeExpr, env)
		require.NoError(t, err)
		assert.Equal(t, int64(10), result)
	})
}

// Test pipe with lambda expressions
func TestInterpreter_PipeWithLambda(t *testing.T) {
	t.Run("pipe to lambda", func(t *testing.T) {
		// Create: 5 |> (v -> v * 2)
		pipeExpr := PipeExpr{
			Left: LiteralExpr{Value: IntLiteral{Value: 5}},
			Right: LambdaExpr{
				Params: []Field{
					{Name: "v", TypeAnnotation: IntType{}, Required: true},
				},
				Body: BinaryOpExpr{
					Op:    Mul,
					Left:  VariableExpr{Name: "v"},
					Right: LiteralExpr{Value: IntLiteral{Value: 2}},
				},
			},
		}

		interp := newTestInterpreter()
		env := NewEnvironment()

		result, err := interp.EvaluateExpression(pipeExpr, env)
		require.NoError(t, err)
		assert.Equal(t, int64(10), result)
	})
}

// Test pipe error handling
func TestInterpreter_PipeErrors(t *testing.T) {
	t.Run("pipe to undefined function", func(t *testing.T) {
		// Create: 5 |> undefined_func
		pipeExpr := PipeExpr{
			Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
			Right: VariableExpr{Name: "undefined_func"},
		}

		interp := newTestInterpreter()
		env := NewEnvironment()

		_, err := interp.EvaluateExpression(pipeExpr, env)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "undefined function")
	})

	t.Run("pipe to non-function", func(t *testing.T) {
		// Define a variable, not a function
		env := NewEnvironment()
		env.Define("notAFunc", int64(42))

		// Create: 5 |> notAFunc
		pipeExpr := PipeExpr{
			Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
			Right: VariableExpr{Name: "notAFunc"},
		}

		interp := newTestInterpreter()

		_, err := interp.EvaluateExpression(pipeExpr, env)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a function")
	})
}

// Test LambdaClosure evaluation
func TestInterpreter_LambdaClosure(t *testing.T) {
	t.Run("evaluate lambda expression creates closure", func(t *testing.T) {
		lambda := LambdaExpr{
			Params: []Field{
				{Name: "x", TypeAnnotation: IntType{}, Required: true},
			},
			Body: BinaryOpExpr{
				Op:    Mul,
				Left:  VariableExpr{Name: "x"},
				Right: LiteralExpr{Value: IntLiteral{Value: 2}},
			},
		}

		interp := newTestInterpreter()
		env := NewEnvironment()
		result, err := interp.EvaluateExpression(lambda, env)
		require.NoError(t, err)

		closure, ok := result.(*LambdaClosure)
		require.True(t, ok, "expected LambdaClosure, got %T", result)
		assert.NotNil(t, closure.Env)
		assert.Equal(t, lambda, closure.Lambda)
	})
}
