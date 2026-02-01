package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Array Concatenation Tests ---

func TestArrayConcatenation(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	t.Run("concat two arrays", func(t *testing.T) {
		left := []interface{}{int64(1), int64(2)}
		right := []interface{}{int64(3), int64(4)}
		result, err := interp.evaluateAdd(left, right)
		require.NoError(t, err)
		expected := []interface{}{int64(1), int64(2), int64(3), int64(4)}
		assert.Equal(t, expected, result)
	})

	t.Run("concat empty arrays", func(t *testing.T) {
		left := []interface{}{}
		right := []interface{}{}
		result, err := interp.evaluateAdd(left, right)
		require.NoError(t, err)
		assert.Equal(t, []interface{}{}, result)
	})

	t.Run("concat with empty left", func(t *testing.T) {
		left := []interface{}{}
		right := []interface{}{int64(1)}
		result, err := interp.evaluateAdd(left, right)
		require.NoError(t, err)
		assert.Equal(t, []interface{}{int64(1)}, result)
	})

	t.Run("concat with empty right", func(t *testing.T) {
		left := []interface{}{int64(1)}
		right := []interface{}{}
		result, err := interp.evaluateAdd(left, right)
		require.NoError(t, err)
		assert.Equal(t, []interface{}{int64(1)}, result)
	})

	t.Run("array plus non-array errors", func(t *testing.T) {
		left := []interface{}{int64(1)}
		right := int64(5)
		_, err := interp.evaluateAdd(left, right)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot add array")
	})

	t.Run("concat via binary expression", func(t *testing.T) {
		expr := BinaryOpExpr{
			Op: Add,
			Left: ArrayExpr{Elements: []Expr{
				LiteralExpr{Value: IntLiteral{Value: 1}},
				LiteralExpr{Value: IntLiteral{Value: 2}},
			}},
			Right: ArrayExpr{Elements: []Expr{
				LiteralExpr{Value: IntLiteral{Value: 3}},
			}},
		}
		result, err := interp.EvaluateExpression(expr, env)
		require.NoError(t, err)
		expected := []interface{}{int64(1), int64(2), int64(3)}
		assert.Equal(t, expected, result)
	})

	t.Run("concat does not mutate originals", func(t *testing.T) {
		left := []interface{}{int64(1), int64(2)}
		right := []interface{}{int64(3)}
		leftCopy := make([]interface{}, len(left))
		copy(leftCopy, left)
		_, err := interp.evaluateAdd(left, right)
		require.NoError(t, err)
		assert.Equal(t, leftCopy, left)
	})
}

// --- randomInt and generateId Built-in Tests ---

func TestBuiltinRandomInt(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	t.Run("returns value in range", func(t *testing.T) {
		args := []Expr{
			LiteralExpr{Value: IntLiteral{Value: 1}},
			LiteralExpr{Value: IntLiteral{Value: 10}},
		}
		for range 20 {
			result, err := builtinRandomInt(interp, args, env)
			require.NoError(t, err)
			val, ok := result.(int64)
			require.True(t, ok)
			assert.GreaterOrEqual(t, val, int64(1))
			assert.LessOrEqual(t, val, int64(10))
		}
	})

	t.Run("equal min and max", func(t *testing.T) {
		args := []Expr{
			LiteralExpr{Value: IntLiteral{Value: 5}},
			LiteralExpr{Value: IntLiteral{Value: 5}},
		}
		result, err := builtinRandomInt(interp, args, env)
		require.NoError(t, err)
		assert.Equal(t, int64(5), result)
	})

	t.Run("min greater than max errors", func(t *testing.T) {
		args := []Expr{
			LiteralExpr{Value: IntLiteral{Value: 10}},
			LiteralExpr{Value: IntLiteral{Value: 1}},
		}
		_, err := builtinRandomInt(interp, args, env)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "min <= max")
	})

	t.Run("wrong arg count", func(t *testing.T) {
		args := []Expr{
			LiteralExpr{Value: IntLiteral{Value: 1}},
		}
		_, err := builtinRandomInt(interp, args, env)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expects 2 arguments")
	})

	t.Run("non-integer args error", func(t *testing.T) {
		args := []Expr{
			LiteralExpr{Value: StringLiteral{Value: "abc"}},
			LiteralExpr{Value: IntLiteral{Value: 10}},
		}
		_, err := builtinRandomInt(interp, args, env)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "integer arguments")
	})
}

func TestBuiltinGenerateId(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	t.Run("returns 36-char UUID string", func(t *testing.T) {
		result, err := builtinGenerateId(interp, []Expr{}, env)
		require.NoError(t, err)
		str, ok := result.(string)
		require.True(t, ok)
		assert.Len(t, str, 36)
		assert.Equal(t, 4, strings.Count(str, "-"))
	})

	t.Run("generates unique IDs", func(t *testing.T) {
		result1, err := builtinGenerateId(interp, []Expr{}, env)
		require.NoError(t, err)
		result2, err := builtinGenerateId(interp, []Expr{}, env)
		require.NoError(t, err)
		assert.NotEqual(t, result1, result2)
	})

	t.Run("wrong arg count", func(t *testing.T) {
		args := []Expr{
			LiteralExpr{Value: IntLiteral{Value: 1}},
		}
		_, err := builtinGenerateId(interp, args, env)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expects 0 arguments")
	})
}

// --- Index Assignment Tests ---

func TestIndexAssignArray(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	t.Run("assign to array index", func(t *testing.T) {
		arr := []interface{}{int64(1), int64(2), int64(3)}
		env.Define("arr", arr)

		stmt := IndexAssignStatement{
			Target: ArrayIndexExpr{
				Array: VariableExpr{Name: "arr"},
				Index: LiteralExpr{Value: IntLiteral{Value: 0}},
			},
			Value: LiteralExpr{Value: IntLiteral{Value: 99}},
		}
		_, err := interp.ExecuteStatement(stmt, env)
		require.NoError(t, err)

		val, err := env.Get("arr")
		require.NoError(t, err)
		result := val.([]interface{})
		assert.Equal(t, int64(99), result[0])
		assert.Equal(t, int64(2), result[1])
		assert.Equal(t, int64(3), result[2])
	})

	t.Run("assign to last index", func(t *testing.T) {
		arr := []interface{}{int64(1), int64(2), int64(3)}
		env.Define("arr2", arr)

		stmt := IndexAssignStatement{
			Target: ArrayIndexExpr{
				Array: VariableExpr{Name: "arr2"},
				Index: LiteralExpr{Value: IntLiteral{Value: 2}},
			},
			Value: LiteralExpr{Value: StringLiteral{Value: "replaced"}},
		}
		_, err := interp.ExecuteStatement(stmt, env)
		require.NoError(t, err)

		val, _ := env.Get("arr2")
		result := val.([]interface{})
		assert.Equal(t, "replaced", result[2])
	})

	t.Run("out of bounds error", func(t *testing.T) {
		arr := []interface{}{int64(1)}
		env.Define("arr3", arr)

		stmt := IndexAssignStatement{
			Target: ArrayIndexExpr{
				Array: VariableExpr{Name: "arr3"},
				Index: LiteralExpr{Value: IntLiteral{Value: 5}},
			},
			Value: LiteralExpr{Value: IntLiteral{Value: 99}},
		}
		_, err := interp.ExecuteStatement(stmt, env)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "out of bounds")
	})

	t.Run("non-array target error", func(t *testing.T) {
		env.Define("notArr", "hello")

		stmt := IndexAssignStatement{
			Target: ArrayIndexExpr{
				Array: VariableExpr{Name: "notArr"},
				Index: LiteralExpr{Value: IntLiteral{Value: 0}},
			},
			Value: LiteralExpr{Value: IntLiteral{Value: 99}},
		}
		_, err := interp.ExecuteStatement(stmt, env)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot index-assign")
	})
}

func TestIndexAssignObjectField(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	t.Run("assign to object field via FieldAccessExpr", func(t *testing.T) {
		obj := map[string]interface{}{
			"name": "old",
			"age":  int64(25),
		}
		env.Define("obj", obj)

		stmt := IndexAssignStatement{
			Target: FieldAccessExpr{
				Object: VariableExpr{Name: "obj"},
				Field:  "name",
			},
			Value: LiteralExpr{Value: StringLiteral{Value: "new"}},
		}
		_, err := interp.ExecuteStatement(stmt, env)
		require.NoError(t, err)

		val, _ := env.Get("obj")
		result := val.(map[string]interface{})
		assert.Equal(t, "new", result["name"])
	})
}

func TestIndexAssignNestedArrayInObject(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	t.Run("assign to obj.list[0]", func(t *testing.T) {
		obj := map[string]interface{}{
			"list": []interface{}{int64(10), int64(20), int64(30)},
		}
		env.Define("obj", obj)

		stmt := IndexAssignStatement{
			Target: ArrayIndexExpr{
				Array: FieldAccessExpr{
					Object: VariableExpr{Name: "obj"},
					Field:  "list",
				},
				Index: LiteralExpr{Value: IntLiteral{Value: 0}},
			},
			Value: LiteralExpr{Value: IntLiteral{Value: 99}},
		}
		_, err := interp.ExecuteStatement(stmt, env)
		require.NoError(t, err)

		val, _ := env.Get("obj")
		result := val.(map[string]interface{})
		list := result["list"].([]interface{})
		assert.Equal(t, int64(99), list[0])
	})
}

func TestIndexAssignArrayOfObjects(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	t.Run("assign to arr[0].name", func(t *testing.T) {
		arr := []interface{}{
			map[string]interface{}{"name": "alice"},
			map[string]interface{}{"name": "bob"},
		}
		env.Define("people", arr)

		stmt := IndexAssignStatement{
			Target: FieldAccessExpr{
				Object: ArrayIndexExpr{
					Array: VariableExpr{Name: "people"},
					Index: LiteralExpr{Value: IntLiteral{Value: 0}},
				},
				Field: "name",
			},
			Value: LiteralExpr{Value: StringLiteral{Value: "updated"}},
		}
		_, err := interp.ExecuteStatement(stmt, env)
		require.NoError(t, err)

		val, _ := env.Get("people")
		result := val.([]interface{})
		first := result[0].(map[string]interface{})
		assert.Equal(t, "updated", first["name"])
	})
}

func TestIndexAssignMapWithStringKey(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	t.Run("assign to map with string key via ArrayIndexExpr", func(t *testing.T) {
		obj := map[string]interface{}{
			"key1": "val1",
		}
		env.Define("m", obj)

		stmt := IndexAssignStatement{
			Target: ArrayIndexExpr{
				Array: VariableExpr{Name: "m"},
				Index: LiteralExpr{Value: StringLiteral{Value: "key2"}},
			},
			Value: LiteralExpr{Value: StringLiteral{Value: "val2"}},
		}
		_, err := interp.ExecuteStatement(stmt, env)
		require.NoError(t, err)

		val, _ := env.Get("m")
		result := val.(map[string]interface{})
		assert.Equal(t, "val2", result["key2"])
	})
}
