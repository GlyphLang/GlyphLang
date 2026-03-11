package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModulo_IntegerBasic(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Op:    Mod,
		Right: LiteralExpr{Value: IntLiteral{Value: 3}},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result)
}

func TestModulo_IntegerExact(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: IntLiteral{Value: 9}},
		Op:    Mod,
		Right: LiteralExpr{Value: IntLiteral{Value: 3}},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

func TestModulo_FloatBasic(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: FloatLiteral{Value: 10.5}},
		Op:    Mod,
		Right: LiteralExpr{Value: FloatLiteral{Value: 3.0}},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.InDelta(t, 1.5, result.(float64), 0.001)
}

func TestModulo_IntAndFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// int % float should coerce int to float
	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Op:    Mod,
		Right: LiteralExpr{Value: FloatLiteral{Value: 3.0}},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.InDelta(t, 1.0, result.(float64), 0.001)
}

func TestModulo_ByZero_Int(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Op:    Mod,
		Right: LiteralExpr{Value: IntLiteral{Value: 0}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "modulo by zero")
}

func TestModulo_ByZero_Float(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: FloatLiteral{Value: 10.5}},
		Op:    Mod,
		Right: LiteralExpr{Value: FloatLiteral{Value: 0.0}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "modulo by zero")
}

func TestModulo_NegativeInt(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Go's % operator preserves the sign of the dividend
	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: IntLiteral{Value: -10}},
		Op:    Mod,
		Right: LiteralExpr{Value: IntLiteral{Value: 3}},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, int64(-1), result)
}

func TestModulo_UnsupportedType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: StringLiteral{Value: "hello"}},
		Op:    Mod,
		Right: LiteralExpr{Value: IntLiteral{Value: 3}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}
