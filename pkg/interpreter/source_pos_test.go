package interpreter

import (
	// Dot-imported: all AST types (VariableExpr, BinaryOpExpr, UnaryOpExpr,
	// LiteralExpr, IntLiteral, StringLiteral, FieldAccessExpr, ArrayIndexExpr,
	// FunctionCallExpr, Pos, BinOp operators, UnOp operators) come from this package.
	. "github.com/glyphlang/glyph/pkg/ast"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPosError_WithPosition verifies posError wraps errors with position info.
func TestPosError_WithPosition(t *testing.T) {
	pos := Pos{Line: 5, Column: 10}
	err := posError(pos, assert.AnError)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at 5:10")
}

// TestPosError_WithoutPosition verifies posError passes through when no position is set.
func TestPosError_WithoutPosition(t *testing.T) {
	pos := Pos{}
	err := posError(pos, assert.AnError)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "at ")
}

// TestPosError_NilError verifies posError returns nil for nil errors.
func TestPosError_NilError(t *testing.T) {
	pos := Pos{Line: 1, Column: 1}
	err := posError(pos, nil)
	assert.NoError(t, err)
}

// TestSourcePos_UndefinedVariable verifies that undefined variable errors include position info.
func TestSourcePos_UndefinedVariable(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := VariableExpr{
		Name: "nonexistent",
		Pos:  Pos{Line: 3, Column: 7},
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at 3:7")
	assert.Contains(t, err.Error(), "undefined variable")
}

// TestSourcePos_BinaryOpTypeError verifies that binary op type errors include position info.
func TestSourcePos_BinaryOpTypeError(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Try to subtract a string from an int
	expr := BinaryOpExpr{
		Op:    Sub,
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Right: LiteralExpr{Value: StringLiteral{Value: "hello"}},
		Pos:   Pos{Line: 5, Column: 12},
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at 5:12")
}

// TestSourcePos_UnaryOpTypeError verifies that unary op type errors include position info.
func TestSourcePos_UnaryOpTypeError(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Try to negate a string
	expr := UnaryOpExpr{
		Op:    Neg,
		Right: LiteralExpr{Value: StringLiteral{Value: "hello"}},
		Pos:   Pos{Line: 2, Column: 1},
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at 2:1")
	assert.Contains(t, err.Error(), "unary negation")
}

// TestSourcePos_FieldAccessMissingReturnsNil verifies that missing fields return nil (null).
// This is intentional: API routes need input.optional_field to return null when absent,
// rather than crashing the route with an error.
func TestSourcePos_FieldAccessMissingReturnsNil(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("obj", map[string]interface{}{"name": "test"})

	expr := FieldAccessExpr{
		Object: VariableExpr{Name: "obj"},
		Field:  "missing",
		Pos:    Pos{Line: 10, Column: 5},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Nil(t, result)
}

// TestSourcePos_ArrayIndexOutOfBounds verifies that array index errors include position info.
func TestSourcePos_ArrayIndexOutOfBounds(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("arr", []interface{}{int64(1), int64(2)})

	expr := ArrayIndexExpr{
		Array: VariableExpr{Name: "arr"},
		Index: LiteralExpr{Value: IntLiteral{Value: 5}},
		Pos:   Pos{Line: 7, Column: 3},
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at 7:3")
	assert.Contains(t, err.Error(), "out of bounds")
}

// TestSourcePos_UndefinedFunction verifies that undefined function errors include position info.
func TestSourcePos_UndefinedFunction(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := FunctionCallExpr{
		Name: "nonexistent_fn",
		Args: []Expr{},
		Pos:  Pos{Line: 15, Column: 2},
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at 15:2")
	assert.Contains(t, err.Error(), "undefined function")
}

// TestSourcePos_NoPosition verifies that errors without positions still work correctly.
func TestSourcePos_NoPosition(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// VariableExpr with zero-value Pos (no position set)
	expr := VariableExpr{
		Name: "nonexistent",
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "at ")
	assert.Contains(t, err.Error(), "undefined variable")
}

// TestSourcePos_DivisionByZero verifies that division by zero errors include position info.
func TestSourcePos_DivisionByZero(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Div,
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Right: LiteralExpr{Value: IntLiteral{Value: 0}},
		Pos:   Pos{Line: 4, Column: 8},
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at 4:8")
}

// TestSourcePos_LogicalNotTypeError verifies that logical NOT type errors include position info.
func TestSourcePos_LogicalNotTypeError(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := UnaryOpExpr{
		Op:    Not,
		Right: LiteralExpr{Value: IntLiteral{Value: 42}},
		Pos:   Pos{Line: 6, Column: 3},
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at 6:3")
	assert.Contains(t, err.Error(), "logical NOT")
}

// TestSourcePos_CannotIndexNonArray verifies position info on indexing non-array values.
func TestSourcePos_CannotIndexNonArray(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("val", int64(42))

	expr := ArrayIndexExpr{
		Array: VariableExpr{Name: "val"},
		Index: LiteralExpr{Value: IntLiteral{Value: 0}},
		Pos:   Pos{Line: 8, Column: 4},
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at 8:4")
	assert.Contains(t, err.Error(), "cannot index")
}

// TestPos_HasPos verifies the HasPos method.
func TestPos_HasPos(t *testing.T) {
	assert.False(t, Pos{}.HasPos())
	assert.False(t, Pos{Line: 0, Column: 5}.HasPos())
	assert.True(t, Pos{Line: 1, Column: 0}.HasPos())
	assert.True(t, Pos{Line: 1, Column: 1}.HasPos())
}

// TestPos_String verifies the String method.
func TestPos_String(t *testing.T) {
	assert.Equal(t, "", Pos{}.String())
	assert.Equal(t, "1:0", Pos{Line: 1, Column: 0}.String())
	assert.Equal(t, "5:10", Pos{Line: 5, Column: 10}.String())
}
