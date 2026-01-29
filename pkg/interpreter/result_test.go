package interpreter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ResultValue unit tests ---

func TestResultValue_NewOk(t *testing.T) {
	r := NewOk("hello")
	assert.True(t, r.IsOk())
	assert.False(t, r.IsErr())
	assert.Equal(t, "hello", r.OkValue())
	assert.Nil(t, r.ErrValue())
}

func TestResultValue_NewErr(t *testing.T) {
	r := NewErr("something failed")
	assert.False(t, r.IsOk())
	assert.True(t, r.IsErr())
	assert.Equal(t, "something failed", r.ErrValue())
	assert.Nil(t, r.OkValue())
}

func TestResultValue_Unwrap(t *testing.T) {
	ok := NewOk(42)
	val, err := ok.Unwrap()
	require.NoError(t, err)
	assert.Equal(t, 42, val)

	errResult := NewErr("fail")
	_, err = errResult.Unwrap()
	assert.Error(t, err)
}

func TestResultValue_UnwrapOr(t *testing.T) {
	ok := NewOk(42)
	assert.Equal(t, 42, ok.UnwrapOr(0))

	errResult := NewErr("fail")
	assert.Equal(t, 99, errResult.UnwrapOr(99))
}

func TestResultValue_UnwrapErr(t *testing.T) {
	errResult := NewErr("fail")
	val, err := errResult.UnwrapErr()
	require.NoError(t, err)
	assert.Equal(t, "fail", val)

	ok := NewOk(42)
	_, err = ok.UnwrapErr()
	assert.Error(t, err)
}

func TestResultValue_String(t *testing.T) {
	ok := NewOk("hello")
	assert.Equal(t, "Ok(hello)", ok.String())

	errResult := NewErr("oops")
	assert.Equal(t, "Err(oops)", errResult.String())
}

// --- Interpreter integration tests ---

func TestInterpreter_OkErrBuiltins(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test Ok() constructor
	okExpr := FunctionCallExpr{
		Name: "Ok",
		Args: []Expr{LiteralExpr{Value: IntLiteral{Value: 42}}},
	}
	result, err := interp.EvaluateExpression(okExpr, env)
	require.NoError(t, err)

	rv, ok := result.(*ResultValue)
	require.True(t, ok, "expected *ResultValue, got %T", result)
	assert.True(t, rv.IsOk())
	assert.Equal(t, int64(42), rv.OkValue())

	// Test Err() constructor
	errExpr := FunctionCallExpr{
		Name: "Err",
		Args: []Expr{LiteralExpr{Value: StringLiteral{Value: "not found"}}},
	}
	result, err = interp.EvaluateExpression(errExpr, env)
	require.NoError(t, err)

	rv, ok = result.(*ResultValue)
	require.True(t, ok)
	assert.True(t, rv.IsErr())
	assert.Equal(t, "not found", rv.ErrValue())
}

func TestInterpreter_ResultFieldAccess(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("okResult", NewOk("success"))
	env.Define("errResult", NewErr("failure"))

	// Test .ok field on Ok value
	result, err := interp.EvaluateExpression(FieldAccessExpr{
		Object: VariableExpr{Name: "okResult"},
		Field:  "ok",
	}, env)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Test .value field on Ok value
	result, err = interp.EvaluateExpression(FieldAccessExpr{
		Object: VariableExpr{Name: "okResult"},
		Field:  "value",
	}, env)
	require.NoError(t, err)
	assert.Equal(t, "success", result)

	// Test .ok on Err value
	result, err = interp.EvaluateExpression(FieldAccessExpr{
		Object: VariableExpr{Name: "errResult"},
		Field:  "ok",
	}, env)
	require.NoError(t, err)
	assert.Equal(t, false, result)

	// Test .error on Err value
	result, err = interp.EvaluateExpression(FieldAccessExpr{
		Object: VariableExpr{Name: "errResult"},
		Field:  "error",
	}, env)
	require.NoError(t, err)
	assert.Equal(t, "failure", result)

	// Test .error on Ok value (nil)
	result, err = interp.EvaluateExpression(FieldAccessExpr{
		Object: VariableExpr{Name: "okResult"},
		Field:  "error",
	}, env)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestInterpreter_ResultMethodIsOk(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk(42))

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.isOk",
		Args: []Expr{},
	}, env)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestInterpreter_ResultMethodIsErr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("fail"))

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.isErr",
		Args: []Expr{},
	}, env)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestInterpreter_ResultMethodUnwrap(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk("value"))

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.unwrap",
		Args: []Expr{},
	}, env)
	require.NoError(t, err)
	assert.Equal(t, "value", result)
}

func TestInterpreter_ResultMethodUnwrapErr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("fail"))

	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.unwrap",
		Args: []Expr{},
	}, env)
	assert.Error(t, err)
}

func TestInterpreter_ResultMethodUnwrapOr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("fail"))

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.unwrapOr",
		Args: []Expr{LiteralExpr{Value: IntLiteral{Value: 99}}},
	}, env)
	require.NoError(t, err)
	assert.Equal(t, int64(99), result)
}

func TestInterpreter_ResultPatternMatch_Ok(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk(int64(42)))

	// match r { {ok: val} => val, _ => 0 }
	expr := MatchExpr{
		Value: VariableExpr{Name: "r"},
		Cases: []MatchCase{
			{
				Pattern: ObjectPattern{
					Fields: []ObjectPatternField{
						{Key: "ok", Pattern: VariablePattern{Name: "val"}},
					},
				},
				Body: VariableExpr{Name: "val"},
			},
			{
				Pattern: WildcardPattern{},
				Body:    LiteralExpr{Value: IntLiteral{Value: 0}},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestInterpreter_ResultPatternMatch_Err(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("not found"))

	// match r { {ok: val} => val, {error: e} => e }
	expr := MatchExpr{
		Value: VariableExpr{Name: "r"},
		Cases: []MatchCase{
			{
				Pattern: ObjectPattern{
					Fields: []ObjectPatternField{
						{Key: "ok", Pattern: VariablePattern{Name: "val"}},
					},
				},
				Body: VariableExpr{Name: "val"},
			},
			{
				Pattern: ObjectPattern{
					Fields: []ObjectPatternField{
						{Key: "error", Pattern: VariablePattern{Name: "e"}},
					},
				},
				Body: VariableExpr{Name: "e"},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "not found", result)
}

func TestInterpreter_ResultInRoute(t *testing.T) {
	interp := NewInterpreter()

	// Route that returns Ok(result)
	route := &Route{
		Path:   "/api/test",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "result",
				Value: FunctionCallExpr{
					Name: "Ok",
					Args: []Expr{
						ObjectExpr{
							Fields: []ObjectField{
								{Key: "id", Value: LiteralExpr{Value: IntLiteral{Value: 1}}},
								{Key: "name", Value: LiteralExpr{Value: StringLiteral{Value: "test"}}},
							},
						},
					},
				},
			},
			ReturnStatement{
				Value: VariableExpr{Name: "result"},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, map[string]string{})
	require.NoError(t, err)

	rv, ok := result.(*ResultValue)
	require.True(t, ok, "expected *ResultValue, got %T", result)
	assert.True(t, rv.IsOk())
}

func TestInterpreter_ResultOkErrArgValidation(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Ok with no args
	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "Ok",
		Args: []Expr{},
	}, env)
	assert.Error(t, err)

	// Err with no args
	_, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "Err",
		Args: []Expr{},
	}, env)
	assert.Error(t, err)

	// Ok with too many args
	_, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "Ok",
		Args: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 1}},
			LiteralExpr{Value: IntLiteral{Value: 2}},
		},
	}, env)
	assert.Error(t, err)
}
