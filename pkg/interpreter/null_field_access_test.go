package interpreter

import (
	"strings"
	"testing"

	// Dot-import matches the existing convention used by all sibling test
	// files in this package (see redirect_test.go, result_test.go, etc.).
	. "github.com/glyphlang/glyph/pkg/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNullFieldAccess_InRoute is a regression test for issue #236.
// Accessing a field on a null value must produce a clean Glyph-level error
// rather than a Go panic from reflect.Value.MethodByName.
func TestNullFieldAccess_InRoute(t *testing.T) {
	interp := NewInterpreter()

	// Route equivalent to:
	//   @ GET /test {
	//     $ x = null
	//     $ y = x.foo
	//   }
	// The second statement must error cleanly.
	route := &Route{
		Method: Get,
		Path:   "/test",
		Body: []Statement{
			AssignStatement{
				Target: "x",
				Value:  LiteralExpr{Value: NullLiteral{}},
			},
			AssignStatement{
				Target: "y",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "x"},
					Field:  "foo",
				},
			},
		},
	}

	request := &Request{
		Path:    "/test",
		Method:  "GET",
		Params:  map[string]string{},
		Headers: map[string]string{},
	}

	// Outer-scope var captures the real error after the panic-guarded call.
	var err error
	assert.NotPanics(t, func() {
		_, err = interp.ExecuteRoute(route, request)
	})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "null")
	assert.Contains(t, err.Error(), "foo")
}

// TestNullFieldAccess_Direct exercises the evaluator's field-access path on a
// null object directly, ensuring the null short-circuit returns an error
// instead of panicking.
func TestNullFieldAccess_Direct(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("x", nil)

	// Outer-scope var captures the real error after the panic-guarded call.
	var err error
	assert.NotPanics(t, func() {
		_, err = interp.EvaluateExpression(FieldAccessExpr{
			Object: VariableExpr{Name: "x"},
			Field:  "foo",
		}, env)
	})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "null")
	assert.Contains(t, err.Error(), "foo")
}
