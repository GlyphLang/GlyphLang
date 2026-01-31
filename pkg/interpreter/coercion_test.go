package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test numeric coercion in arithmetic operations

func TestCoercion_Add_IntAndFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 5 + 3.2 should coerce int to float
	left := LiteralExpr{Value: IntLiteral{Value: 5}}
	right := LiteralExpr{Value: FloatLiteral{Value: 3.2}}
	expr := BinaryOpExpr{Left: left, Op: Add, Right: right}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, float64(8.2), result)
}

func TestCoercion_Add_FloatAndInt(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 3.5 + 2 should coerce int to float
	left := LiteralExpr{Value: FloatLiteral{Value: 3.5}}
	right := LiteralExpr{Value: IntLiteral{Value: 2}}
	expr := BinaryOpExpr{Left: left, Op: Add, Right: right}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, float64(5.5), result)
}

func TestCoercion_Sub_IntAndFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 10 - 3.5 should error (strict type checking for subtraction)
	left := LiteralExpr{Value: IntLiteral{Value: 10}}
	right := LiteralExpr{Value: FloatLiteral{Value: 3.5}}
	expr := BinaryOpExpr{Left: left, Op: Sub, Right: right}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot subtract")
}

func TestCoercion_Mul_IntAndFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 4 * 2.5 should error (strict type checking for multiplication)
	left := LiteralExpr{Value: IntLiteral{Value: 4}}
	right := LiteralExpr{Value: FloatLiteral{Value: 2.5}}
	expr := BinaryOpExpr{Left: left, Op: Mul, Right: right}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot multiply")
}

func TestCoercion_Div_IntAndFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 10 / 2.5 should error (strict type checking for division)
	left := LiteralExpr{Value: IntLiteral{Value: 10}}
	right := LiteralExpr{Value: FloatLiteral{Value: 2.5}}
	expr := BinaryOpExpr{Left: left, Op: Div, Right: right}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot divide")
}

// Test numeric coercion in comparison operations

func TestCoercion_Eq_IntAndFloat_Equal(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 5 == 5.0 should return true (with coercion)
	left := LiteralExpr{Value: IntLiteral{Value: 5}}
	right := LiteralExpr{Value: FloatLiteral{Value: 5.0}}
	expr := BinaryOpExpr{Left: left, Op: Eq, Right: right}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestCoercion_Eq_IntAndFloat_NotEqual(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 5 == 5.1 should return false
	left := LiteralExpr{Value: IntLiteral{Value: 5}}
	right := LiteralExpr{Value: FloatLiteral{Value: 5.1}}
	expr := BinaryOpExpr{Left: left, Op: Eq, Right: right}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCoercion_Ne_IntAndFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 5 != 5.0 should return false (they are equal with coercion)
	left := LiteralExpr{Value: IntLiteral{Value: 5}}
	right := LiteralExpr{Value: FloatLiteral{Value: 5.0}}
	expr := BinaryOpExpr{Left: left, Op: Ne, Right: right}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCoercion_Lt_IntAndFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 5 < 5.5 should error (strict type checking for comparisons)
	left := LiteralExpr{Value: IntLiteral{Value: 5}}
	right := LiteralExpr{Value: FloatLiteral{Value: 5.5}}
	expr := BinaryOpExpr{Left: left, Op: Lt, Right: right}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot compare")
}

func TestCoercion_Le_FloatAndInt(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 5.0 <= 5 should error (strict type checking for comparisons)
	left := LiteralExpr{Value: FloatLiteral{Value: 5.0}}
	right := LiteralExpr{Value: IntLiteral{Value: 5}}
	expr := BinaryOpExpr{Left: left, Op: Le, Right: right}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot compare")
}

func TestCoercion_Gt_FloatAndInt(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 5.5 > 5 should error (strict type checking for comparisons)
	left := LiteralExpr{Value: FloatLiteral{Value: 5.5}}
	right := LiteralExpr{Value: IntLiteral{Value: 5}}
	expr := BinaryOpExpr{Left: left, Op: Gt, Right: right}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot compare")
}

func TestCoercion_Ge_IntAndFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 5 >= 4.9 should error (strict type checking for comparisons)
	left := LiteralExpr{Value: IntLiteral{Value: 5}}
	right := LiteralExpr{Value: FloatLiteral{Value: 4.9}}
	expr := BinaryOpExpr{Left: left, Op: Ge, Right: right}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot compare")
}

// Test that same-type operations still work

func TestCoercion_Add_BothInt(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 5 + 3 should stay as int
	left := LiteralExpr{Value: IntLiteral{Value: 5}}
	right := LiteralExpr{Value: IntLiteral{Value: 3}}
	expr := BinaryOpExpr{Left: left, Op: Add, Right: right}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, int64(8), result)
}

func TestCoercion_Add_BothFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 5.5 + 3.2 should stay as float
	left := LiteralExpr{Value: FloatLiteral{Value: 5.5}}
	right := LiteralExpr{Value: FloatLiteral{Value: 3.2}}
	expr := BinaryOpExpr{Left: left, Op: Add, Right: right}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, float64(8.7), result)
}

// Test edge cases

func TestCoercion_ChainedOperations(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// (5 + 2.5) * 3 should work with Add coercion, but then fail on Mul
	// First: 5 + 2.5 = 7.5 (float) - Add allows coercion
	// Then: 7.5 * 3 should error - Mul does not allow type mixing
	inner := BinaryOpExpr{
		Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
		Op:    Add,
		Right: LiteralExpr{Value: FloatLiteral{Value: 2.5}},
	}
	expr := BinaryOpExpr{
		Left:  inner,
		Op:    Mul,
		Right: LiteralExpr{Value: IntLiteral{Value: 3}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot multiply")
}

func TestCoercion_DivisionByZero_Float(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// 5 / 0.0 should error due to type mismatch (not division by zero)
	left := LiteralExpr{Value: IntLiteral{Value: 5}}
	right := LiteralExpr{Value: FloatLiteral{Value: 0.0}}
	expr := BinaryOpExpr{Left: left, Op: Div, Right: right}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot divide")
}

func TestCoercion_NoCoercionForString(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// "hello" + 5 should error (no coercion from int to string)
	left := LiteralExpr{Value: StringLiteral{Value: "hello"}}
	right := LiteralExpr{Value: IntLiteral{Value: 5}}
	expr := BinaryOpExpr{Left: left, Op: Add, Right: right}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}
