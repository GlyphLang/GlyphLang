package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMacro_RegisterAndExpand(t *testing.T) {
	interp := NewInterpreter()

	// Define: macro! double(x) { $result = x + x }
	interp.macros["double"] = &MacroDef{
		Name:   "double",
		Params: []string{"x"},
		Body: []Node{
			AssignStatement{
				Target: "result",
				Value: BinaryOpExpr{
					Left:  VariableExpr{Name: "x"},
					Op:    Add,
					Right: VariableExpr{Name: "x"},
				},
			},
		},
	}

	env := NewEnvironment()

	// Invoke: double!(5)
	_, err := interp.executeMacroInvocation(MacroInvocation{
		Name: "double",
		Args: []Expr{LiteralExpr{Value: IntLiteral{Value: 5}}},
	}, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, int64(10), result)
}

func TestMacro_MultipleStatements(t *testing.T) {
	interp := NewInterpreter()

	// macro! swap(a, b) {
	//   $temp = a
	//   $a = b
	//   $b = temp
	// }
	// Note: this macro operates on variable names passed as expressions
	interp.macros["init_pair"] = &MacroDef{
		Name:   "init_pair",
		Params: []string{"x", "y"},
		Body: []Node{
			AssignStatement{
				Target: "first",
				Value:  VariableExpr{Name: "x"},
			},
			AssignStatement{
				Target: "second",
				Value:  VariableExpr{Name: "y"},
			},
		},
	}

	env := NewEnvironment()

	_, err := interp.executeMacroInvocation(MacroInvocation{
		Name: "init_pair",
		Args: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 10}},
			LiteralExpr{Value: IntLiteral{Value: 20}},
		},
	}, env)
	require.NoError(t, err)

	first, err := env.Get("first")
	require.NoError(t, err)
	assert.Equal(t, int64(10), first)

	second, err := env.Get("second")
	require.NoError(t, err)
	assert.Equal(t, int64(20), second)
}

func TestMacro_UndefinedMacro(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	_, err := interp.executeMacroInvocation(MacroInvocation{
		Name: "nonexistent",
		Args: []Expr{},
	}, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "undefined macro: nonexistent")
}

func TestMacro_WrongArgCount(t *testing.T) {
	interp := NewInterpreter()

	interp.macros["one_arg"] = &MacroDef{
		Name:   "one_arg",
		Params: []string{"x"},
		Body:   []Node{},
	}

	env := NewEnvironment()

	_, err := interp.executeMacroInvocation(MacroInvocation{
		Name: "one_arg",
		Args: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 1}},
			LiteralExpr{Value: IntLiteral{Value: 2}},
		},
	}, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expects 1 arguments, got 2")
}

func TestMacro_WithIfStatement(t *testing.T) {
	interp := NewInterpreter()

	// macro! clamp(val, limit) {
	//   if val > limit { result = limit } else { result = val }
	// }
	// Uses ReassignStatement so the variable is set in the parent scope
	interp.macros["clamp"] = &MacroDef{
		Name:   "clamp",
		Params: []string{"val", "limit"},
		Body: []Node{
			IfStatement{
				Condition: BinaryOpExpr{
					Left:  VariableExpr{Name: "val"},
					Op:    Gt,
					Right: VariableExpr{Name: "limit"},
				},
				ThenBlock: []Statement{
					ReassignStatement{
						Target: "result",
						Value:  VariableExpr{Name: "limit"},
					},
				},
				ElseBlock: []Statement{
					ReassignStatement{
						Target: "result",
						Value:  VariableExpr{Name: "val"},
					},
				},
			},
		},
	}

	env := NewEnvironment()
	env.Define("result", int64(0))

	// clamp!(100, 50) => result should be 50
	_, err := interp.executeMacroInvocation(MacroInvocation{
		Name: "clamp",
		Args: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 100}},
			LiteralExpr{Value: IntLiteral{Value: 50}},
		},
	}, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, int64(50), result)
}

func TestMacro_AsExpression(t *testing.T) {
	interp := NewInterpreter()

	// macro! add(a, b) { $result = a + b }
	interp.macros["add"] = &MacroDef{
		Name:   "add",
		Params: []string{"a", "b"},
		Body: []Node{
			AssignStatement{
				Target: "result",
				Value: BinaryOpExpr{
					Left:  VariableExpr{Name: "a"},
					Op:    Add,
					Right: VariableExpr{Name: "b"},
				},
			},
		},
	}

	env := NewEnvironment()

	// Evaluate as expression
	_, err := interp.evaluateMacroInvocation(MacroInvocation{
		Name: "add",
		Args: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 3}},
			LiteralExpr{Value: IntLiteral{Value: 7}},
		},
	}, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, int64(10), result)
}

func TestMacro_NestedExpansion(t *testing.T) {
	interp := NewInterpreter()

	// macro! inc(x) { result = x + 1 }
	// Uses ReassignStatement so it works when called multiple times
	interp.macros["inc"] = &MacroDef{
		Name:   "inc",
		Params: []string{"x"},
		Body: []Node{
			ReassignStatement{
				Target: "result",
				Value: BinaryOpExpr{
					Left:  VariableExpr{Name: "x"},
					Op:    Add,
					Right: LiteralExpr{Value: IntLiteral{Value: 1}},
				},
			},
		},
	}

	// macro! inc_twice(x) { inc!(x) inc!(result) }
	interp.macros["inc_twice"] = &MacroDef{
		Name:   "inc_twice",
		Params: []string{"x"},
		Body: []Node{
			&MacroInvocation{
				Name: "inc",
				Args: []Expr{VariableExpr{Name: "x"}},
			},
			// After first inc, result holds x+1; use it for the second
			&MacroInvocation{
				Name: "inc",
				Args: []Expr{VariableExpr{Name: "result"}},
			},
		},
	}

	env := NewEnvironment()
	env.Define("result", int64(0))

	_, err := interp.executeMacroInvocation(MacroInvocation{
		Name: "inc_twice",
		Args: []Expr{LiteralExpr{Value: IntLiteral{Value: 5}}},
	}, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	// 5 + 1 = 6, then 6 + 1 = 7
	assert.Equal(t, int64(7), result)
}

func TestMacro_StringInterpolation(t *testing.T) {
	interp := NewInterpreter()

	// macro! greet(name) { $msg = "Hello, ${name}!" }
	interp.macros["greet"] = &MacroDef{
		Name:   "greet",
		Params: []string{"name"},
		Body: []Node{
			AssignStatement{
				Target: "msg",
				Value:  LiteralExpr{Value: StringLiteral{Value: "Hello, ${name}!"}},
			},
		},
	}

	env := NewEnvironment()

	_, err := interp.executeMacroInvocation(MacroInvocation{
		Name: "greet",
		Args: []Expr{LiteralExpr{Value: StringLiteral{Value: "World"}}},
	}, env)
	require.NoError(t, err)

	msg, err := env.Get("msg")
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", msg)
}

func TestMacro_LoadModule(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&MacroDef{
				Name:   "square",
				Params: []string{"x"},
				Body: []Node{
					AssignStatement{
						Target: "result",
						Value: BinaryOpExpr{
							Left:  VariableExpr{Name: "x"},
							Op:    Mul,
							Right: VariableExpr{Name: "x"},
						},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	// Verify macro was registered
	_, ok := interp.macros["square"]
	assert.True(t, ok)
}

func TestMacro_LoadModuleWithInvocation(t *testing.T) {
	interp := NewInterpreter()

	// First load the macro definition
	module1 := Module{
		Items: []Item{
			&MacroDef{
				Name:   "set_val",
				Params: []string{"v"},
				Body: []Node{
					AssignStatement{
						Target: "myval",
						Value:  VariableExpr{Name: "v"},
					},
				},
			},
		},
	}
	err := interp.LoadModule(module1)
	require.NoError(t, err)

	// Then load a module with an invocation
	module2 := Module{
		Items: []Item{
			&MacroInvocation{
				Name: "set_val",
				Args: []Expr{LiteralExpr{Value: IntLiteral{Value: 42}}},
			},
		},
	}
	err = interp.LoadModule(module2)
	require.NoError(t, err)

	result, err := interp.globalEnv.Get("myval")
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestMacro_EmptyBody(t *testing.T) {
	interp := NewInterpreter()

	interp.macros["noop"] = &MacroDef{
		Name:   "noop",
		Params: []string{},
		Body:   []Node{},
	}

	env := NewEnvironment()

	result, err := interp.executeMacroInvocation(MacroInvocation{
		Name: "noop",
		Args: []Expr{},
	}, env)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestMacro_ViaExecuteStatement(t *testing.T) {
	interp := NewInterpreter()

	interp.macros["assign42"] = &MacroDef{
		Name:   "assign42",
		Params: []string{},
		Body: []Node{
			AssignStatement{
				Target: "x",
				Value:  LiteralExpr{Value: IntLiteral{Value: 42}},
			},
		},
	}

	env := NewEnvironment()

	// Execute via the main ExecuteStatement dispatch
	_, err := interp.ExecuteStatement(MacroInvocation{
		Name: "assign42",
		Args: []Expr{},
	}, env)
	require.NoError(t, err)

	result, err := env.Get("x")
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestMacro_ViaEvaluateExpression(t *testing.T) {
	interp := NewInterpreter()

	interp.macros["set_flag"] = &MacroDef{
		Name:   "set_flag",
		Params: []string{},
		Body: []Node{
			AssignStatement{
				Target: "flag",
				Value:  LiteralExpr{Value: BoolLiteral{Value: true}},
			},
		},
	}

	env := NewEnvironment()

	// Evaluate via EvaluateExpression dispatch
	_, err := interp.EvaluateExpression(MacroInvocation{
		Name: "set_flag",
		Args: []Expr{},
	}, env)
	require.NoError(t, err)

	result, err := env.Get("flag")
	require.NoError(t, err)
	assert.Equal(t, true, result)
}
