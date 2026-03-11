package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper: creates a lambda closure that doubles a number
func doubleLambda(env *Environment) *LambdaClosure {
	return &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "n", Required: true}},
			Body: BinaryOpExpr{
				Left:  VariableExpr{Name: "n"},
				Op:    Mul,
				Right: LiteralExpr{Value: IntLiteral{Value: 2}},
			},
		},
		Env: env,
	}
}

// helper: creates a lambda closure that checks if n > threshold
func greaterThanLambda(env *Environment, threshold int64) *LambdaClosure {
	return &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "n", Required: true}},
			Body: BinaryOpExpr{
				Left:  VariableExpr{Name: "n"},
				Op:    Gt,
				Right: LiteralExpr{Value: IntLiteral{Value: threshold}},
			},
		},
		Env: env,
	}
}

// helper: creates a lambda closure that adds two values (acc, n) => acc + n
func addLambda(env *Environment) *LambdaClosure {
	return &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{
				{Name: "acc", Required: true},
				{Name: "n", Required: true},
			},
			Body: BinaryOpExpr{
				Left:  VariableExpr{Name: "acc"},
				Op:    Add,
				Right: VariableExpr{Name: "n"},
			},
		},
		Env: env,
	}
}

func TestMap_Basic(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(2), int64(3)}
	env.Define("arr", arr)
	env.Define("fn", doubleLambda(env))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "map",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int64(2), int64(4), int64(6)}, result)
}

func TestMap_EmptyArray(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("arr", []interface{}{})
	env.Define("fn", doubleLambda(env))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "map",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{}, result)
}

func TestFilter_Basic(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5)}
	env.Define("arr", arr)
	env.Define("fn", greaterThanLambda(env, 3))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "filter",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int64(4), int64(5)}, result)
}

func TestFilter_NoneMatch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(3), int64(5)}
	env.Define("arr", arr)
	// All elements are <= 10, so greaterThan(10) matches nothing
	env.Define("fn", greaterThanLambda(env, 10))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "filter",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{}, result)
}

func TestReduce_Sum(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5)}
	env.Define("arr", arr)
	env.Define("fn", addLambda(env))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "reduce",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
					LiteralExpr{Value: IntLiteral{Value: 0}},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, int64(15), result)
}

func TestReduce_EmptyArray(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("arr", []interface{}{})
	env.Define("fn", addLambda(env))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "reduce",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
					LiteralExpr{Value: IntLiteral{Value: 42}},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestFind_Found(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(2), int64(3), int64(4)}
	env.Define("arr", arr)
	env.Define("fn", greaterThanLambda(env, 2))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "find",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, int64(3), result)
}

func TestFind_NotFound(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(2)}
	env.Define("arr", arr)
	env.Define("fn", greaterThanLambda(env, 10))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "find",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestSome_True(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(2), int64(3)}
	env.Define("arr", arr)
	// 2 > 1, so some() should return true
	env.Define("fn", greaterThanLambda(env, 1))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "some",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestSome_False(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(3), int64(5)}
	env.Define("arr", arr)
	// greaterThanLambda is defined at the top of this file; all elements <= 10
	env.Define("fn", greaterThanLambda(env, 10))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "some",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestEvery_True(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// All elements > 0, so greaterThan(0) is true for every element
	arr := []interface{}{int64(2), int64(4), int64(6)}
	env.Define("arr", arr)
	env.Define("fn", greaterThanLambda(env, 0))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "every",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvery_False(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 2 is not > 3, so greaterThan(3) fails for at least one element
	arr := []interface{}{int64(2), int64(5), int64(4)}
	env.Define("arr", arr)
	env.Define("fn", greaterThanLambda(env, 3))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "every",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSort_DefaultInts(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(3), int64(1), int64(4), int64(1), int64(5)}
	env.Define("arr", arr)

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "sort",
				Args: []Expr{
					VariableExpr{Name: "arr"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int64(1), int64(1), int64(3), int64(4), int64(5)}, result)
}

func TestSort_DefaultStrings(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{"banana", "apple", "cherry"}
	env.Define("arr", arr)

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "sort",
				Args: []Expr{
					VariableExpr{Name: "arr"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{"apple", "banana", "cherry"}, result)
}

func TestSort_DoesNotMutateOriginal(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	original := []interface{}{int64(3), int64(1), int64(2)}
	env.Define("arr", original)

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "sort",
				Args: []Expr{
					VariableExpr{Name: "arr"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	arrVal, _ := env.Get("arr")
	assert.Equal(t, []interface{}{int64(3), int64(1), int64(2)}, arrVal)
}

func TestReverse_Basic(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(2), int64(3)}
	env.Define("arr", arr)

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "reverse",
				Args: []Expr{
					VariableExpr{Name: "arr"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int64(3), int64(2), int64(1)}, result)
}

func TestFlat_Basic(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{
		[]interface{}{int64(1), int64(2)},
		[]interface{}{int64(3)},
		int64(4),
	}
	env.Define("arr", arr)

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "flat",
				Args: []Expr{
					VariableExpr{Name: "arr"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int64(1), int64(2), int64(3), int64(4)}, result)
}

func TestSlice_Basic(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(10), int64(20), int64(30), int64(40), int64(50)}
	env.Define("arr", arr)

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "slice",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					LiteralExpr{Value: IntLiteral{Value: 1}},
					LiteralExpr{Value: IntLiteral{Value: 3}},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int64(20), int64(30)}, result)
}

func TestSlice_OutOfBounds(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(2), int64(3)}
	env.Define("arr", arr)

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "slice",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					LiteralExpr{Value: IntLiteral{Value: 0}},
					LiteralExpr{Value: IntLiteral{Value: 100}},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int64(1), int64(2), int64(3)}, result)
}

func TestMap_WrongArgCount(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "map",
				Args: []Expr{
					LiteralExpr{Value: IntLiteral{Value: 1}},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "map() expects 2 arguments")
}

func TestFilter_NotAnArray(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("fn", doubleLambda(env))

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "filter",
				Args: []Expr{
					LiteralExpr{Value: IntLiteral{Value: 42}},
					VariableExpr{Name: "fn"},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "filter() expects first argument to be an array")
}
