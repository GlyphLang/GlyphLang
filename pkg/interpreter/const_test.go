package interpreter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constant is properly loaded and accessible
func TestInterpreter_LoadConstant(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&ConstDecl{
				Name:  "MAX_SIZE",
				Value: LiteralExpr{Value: IntLiteral{Value: 100}},
				Type:  nil,
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	// Check constant is accessible
	val, err := interp.globalEnv.Get("MAX_SIZE")
	require.NoError(t, err)
	assert.Equal(t, int64(100), val)

	// Check it's marked as constant
	assert.True(t, interp.IsConstant("MAX_SIZE"))
}

// Test constant with string value
func TestInterpreter_LoadConstantString(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&ConstDecl{
				Name:  "API_URL",
				Value: LiteralExpr{Value: StringLiteral{Value: "https://api.example.com"}},
				Type:  nil,
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	val, err := interp.globalEnv.Get("API_URL")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", val)
}

// Test constant with expression value
func TestInterpreter_LoadConstantExpression(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&ConstDecl{
				Name: "DOUBLED",
				Value: BinaryOpExpr{
					Op:    Mul,
					Left:  LiteralExpr{Value: IntLiteral{Value: 50}},
					Right: LiteralExpr{Value: IntLiteral{Value: 2}},
				},
				Type: nil,
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	val, err := interp.globalEnv.Get("DOUBLED")
	require.NoError(t, err)
	assert.Equal(t, int64(100), val)
}

// Test that constants cannot be reassigned
func TestInterpreter_ConstantCannotBeReassigned(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&ConstDecl{
				Name:  "MAX_SIZE",
				Value: LiteralExpr{Value: IntLiteral{Value: 100}},
				Type:  nil,
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	// Try to reassign the constant
	stmt := ReassignStatement{
		Target: "MAX_SIZE",
		Value:  LiteralExpr{Value: IntLiteral{Value: 200}},
	}

	env := NewChildEnvironment(interp.globalEnv)
	_, err = interp.executeReassign(stmt, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot reassign constant")
}

// Test regular variables can still be reassigned
func TestInterpreter_VariableCanBeReassigned(t *testing.T) {
	interp := NewInterpreter()

	// Define a regular variable
	interp.globalEnv.Define("x", int64(10))

	// Reassign it
	stmt := ReassignStatement{
		Target: "x",
		Value:  LiteralExpr{Value: IntLiteral{Value: 20}},
	}

	env := NewChildEnvironment(interp.globalEnv)
	_, err := interp.executeReassign(stmt, env)
	require.NoError(t, err)

	val, err := env.Get("x")
	require.NoError(t, err)
	assert.Equal(t, int64(20), val)
}

// Test multiple constants
func TestInterpreter_LoadMultipleConstants(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&ConstDecl{
				Name:  "A",
				Value: LiteralExpr{Value: IntLiteral{Value: 1}},
				Type:  nil,
			},
			&ConstDecl{
				Name:  "B",
				Value: LiteralExpr{Value: IntLiteral{Value: 2}},
				Type:  nil,
			},
			&ConstDecl{
				Name:  "C",
				Value: LiteralExpr{Value: IntLiteral{Value: 3}},
				Type:  nil,
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	// All should be constants
	assert.True(t, interp.IsConstant("A"))
	assert.True(t, interp.IsConstant("B"))
	assert.True(t, interp.IsConstant("C"))

	// All should have correct values
	valA, _ := interp.globalEnv.Get("A")
	valB, _ := interp.globalEnv.Get("B")
	valC, _ := interp.globalEnv.Get("C")
	assert.Equal(t, int64(1), valA)
	assert.Equal(t, int64(2), valB)
	assert.Equal(t, int64(3), valC)
}

// Test constant with type annotation
func TestInterpreter_LoadConstantWithType(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&ConstDecl{
				Name:  "PI",
				Value: LiteralExpr{Value: FloatLiteral{Value: 3.14159}},
				Type:  FloatType{},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	val, err := interp.globalEnv.Get("PI")
	require.NoError(t, err)
	assert.InDelta(t, 3.14159, val, 0.00001)
}

// Test constant available in route execution
func TestInterpreter_ConstantAccessibleInRoute(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&ConstDecl{
				Name:  "DEFAULT_STATUS",
				Value: LiteralExpr{Value: StringLiteral{Value: "active"}},
				Type:  nil,
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	// Create a route that returns the constant
	route := &Route{
		Path:   "/test",
		Method: Get,
		Body: []Statement{
			ReturnStatement{
				Value: VariableExpr{Name: "DEFAULT_STATUS"},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)
	require.NoError(t, err)
	assert.Equal(t, "active", result)
}
