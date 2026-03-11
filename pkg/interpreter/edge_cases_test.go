package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Nested break/continue tests
// ---------------------------------------------------------------------------

func TestBreak_InnerForLoop_DoesNotBreakOuter(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// $outerCount = 0
	// for i in [1, 2, 3] {
	//   for j in [10, 20, 30] {
	//     if j == 20 { break }   // breaks inner only
	//   }
	//   $outerCount = $outerCount + 1
	// }
	// outerCount should be 3 (outer loop completes all iterations)
	env.Define("outerCount", int64(0))

	stmts := []Statement{
		ForStatement{
			ValueVar: "i",
			Iterable: ArrayExpr{
				Elements: []Expr{
					LiteralExpr{Value: IntLiteral{Value: 1}},
					LiteralExpr{Value: IntLiteral{Value: 2}},
					LiteralExpr{Value: IntLiteral{Value: 3}},
				},
			},
			Body: []Statement{
				ForStatement{
					ValueVar: "j",
					Iterable: ArrayExpr{
						Elements: []Expr{
							LiteralExpr{Value: IntLiteral{Value: 10}},
							LiteralExpr{Value: IntLiteral{Value: 20}},
							LiteralExpr{Value: IntLiteral{Value: 30}},
						},
					},
					Body: []Statement{
						IfStatement{
							Condition: BinaryOpExpr{
								Left:  VariableExpr{Name: "j"},
								Op:    Eq,
								Right: LiteralExpr{Value: IntLiteral{Value: 20}},
							},
							ThenBlock: []Statement{
								BreakStatement{},
							},
						},
					},
				},
				ReassignStatement{
					Target: "outerCount",
					Value: BinaryOpExpr{
						Left:  VariableExpr{Name: "outerCount"},
						Op:    Add,
						Right: LiteralExpr{Value: IntLiteral{Value: 1}},
					},
				},
			},
		},
	}

	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	outerCount, err := env.Get("outerCount")
	require.NoError(t, err)
	assert.Equal(t, int64(3), outerCount)
}

func TestContinue_InnerForLoop_DoesNotAffectOuter(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// $sum = 0
	// for i in [1, 2] {
	//   for j in [10, 20, 30] {
	//     if j == 20 { continue }  // skip j==20 in inner loop only
	//     $sum = $sum + j
	//   }
	// }
	// Each inner iteration adds 10 + 30 = 40, outer runs twice => sum = 80
	env.Define("sum", int64(0))

	stmts := []Statement{
		ForStatement{
			ValueVar: "i",
			Iterable: ArrayExpr{
				Elements: []Expr{
					LiteralExpr{Value: IntLiteral{Value: 1}},
					LiteralExpr{Value: IntLiteral{Value: 2}},
				},
			},
			Body: []Statement{
				ForStatement{
					ValueVar: "j",
					Iterable: ArrayExpr{
						Elements: []Expr{
							LiteralExpr{Value: IntLiteral{Value: 10}},
							LiteralExpr{Value: IntLiteral{Value: 20}},
							LiteralExpr{Value: IntLiteral{Value: 30}},
						},
					},
					Body: []Statement{
						IfStatement{
							Condition: BinaryOpExpr{
								Left:  VariableExpr{Name: "j"},
								Op:    Eq,
								Right: LiteralExpr{Value: IntLiteral{Value: 20}},
							},
							ThenBlock: []Statement{
								ContinueStatement{},
							},
						},
						ReassignStatement{
							Target: "sum",
							Value: BinaryOpExpr{
								Left:  VariableExpr{Name: "sum"},
								Op:    Add,
								Right: VariableExpr{Name: "j"},
							},
						},
					},
				},
			},
		},
	}

	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	sum, err := env.Get("sum")
	require.NoError(t, err)
	assert.Equal(t, int64(80), sum)
}

func TestBreak_NestedThreeLevelForLoops(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// $deepCount = 0
	// $midCount = 0
	// $outerCount = 0
	// for i in [1, 2] {
	//   $outerCount = $outerCount + 1
	//   for j in [1, 2] {
	//     $midCount = $midCount + 1
	//     for k in [1, 2, 3] {
	//       if k == 2 { break }  // only breaks innermost
	//       $deepCount = $deepCount + 1
	//     }
	//   }
	// }
	// innermost: break at k==2, so only k==1 runs => deepCount increments once per mid*outer = 2*2 = 4
	// midCount = 2 per outer * 2 outer = 4
	// outerCount = 2
	env.Define("deepCount", int64(0))
	env.Define("midCount", int64(0))
	env.Define("outerCount", int64(0))

	stmts := []Statement{
		ForStatement{
			ValueVar: "i",
			Iterable: ArrayExpr{
				Elements: []Expr{
					LiteralExpr{Value: IntLiteral{Value: 1}},
					LiteralExpr{Value: IntLiteral{Value: 2}},
				},
			},
			Body: []Statement{
				ReassignStatement{
					Target: "outerCount",
					Value: BinaryOpExpr{
						Left:  VariableExpr{Name: "outerCount"},
						Op:    Add,
						Right: LiteralExpr{Value: IntLiteral{Value: 1}},
					},
				},
				ForStatement{
					ValueVar: "j",
					Iterable: ArrayExpr{
						Elements: []Expr{
							LiteralExpr{Value: IntLiteral{Value: 1}},
							LiteralExpr{Value: IntLiteral{Value: 2}},
						},
					},
					Body: []Statement{
						ReassignStatement{
							Target: "midCount",
							Value: BinaryOpExpr{
								Left:  VariableExpr{Name: "midCount"},
								Op:    Add,
								Right: LiteralExpr{Value: IntLiteral{Value: 1}},
							},
						},
						ForStatement{
							ValueVar: "k",
							Iterable: ArrayExpr{
								Elements: []Expr{
									LiteralExpr{Value: IntLiteral{Value: 1}},
									LiteralExpr{Value: IntLiteral{Value: 2}},
									LiteralExpr{Value: IntLiteral{Value: 3}},
								},
							},
							Body: []Statement{
								IfStatement{
									Condition: BinaryOpExpr{
										Left:  VariableExpr{Name: "k"},
										Op:    Eq,
										Right: LiteralExpr{Value: IntLiteral{Value: 2}},
									},
									ThenBlock: []Statement{
										BreakStatement{},
									},
								},
								ReassignStatement{
									Target: "deepCount",
									Value: BinaryOpExpr{
										Left:  VariableExpr{Name: "deepCount"},
										Op:    Add,
										Right: LiteralExpr{Value: IntLiteral{Value: 1}},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	outerCount, err := env.Get("outerCount")
	require.NoError(t, err)
	assert.Equal(t, int64(2), outerCount)

	midCount, err := env.Get("midCount")
	require.NoError(t, err)
	assert.Equal(t, int64(4), midCount)

	deepCount, err := env.Get("deepCount")
	require.NoError(t, err)
	assert.Equal(t, int64(4), deepCount)
}

func TestContinue_NestedWhileAndFor(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// $result = 0
	// $i = 0
	// while $i < 3 {
	//   $i = $i + 1
	//   for j in [1, 2, 3] {
	//     if j == 2 { continue }  // skip j==2 in inner for loop
	//     $result = $result + j
	//   }
	// }
	// inner loop adds 1 + 3 = 4 per outer iteration, outer runs 3 times => result = 12
	env.Define("result", int64(0))
	env.Define("i", int64(0))

	stmts := []Statement{
		WhileStatement{
			Condition: BinaryOpExpr{
				Left:  VariableExpr{Name: "i"},
				Op:    Lt,
				Right: LiteralExpr{Value: IntLiteral{Value: 3}},
			},
			Body: []Statement{
				ReassignStatement{
					Target: "i",
					Value: BinaryOpExpr{
						Left:  VariableExpr{Name: "i"},
						Op:    Add,
						Right: LiteralExpr{Value: IntLiteral{Value: 1}},
					},
				},
				ForStatement{
					ValueVar: "j",
					Iterable: ArrayExpr{
						Elements: []Expr{
							LiteralExpr{Value: IntLiteral{Value: 1}},
							LiteralExpr{Value: IntLiteral{Value: 2}},
							LiteralExpr{Value: IntLiteral{Value: 3}},
						},
					},
					Body: []Statement{
						IfStatement{
							Condition: BinaryOpExpr{
								Left:  VariableExpr{Name: "j"},
								Op:    Eq,
								Right: LiteralExpr{Value: IntLiteral{Value: 2}},
							},
							ThenBlock: []Statement{
								ContinueStatement{},
							},
						},
						ReassignStatement{
							Target: "result",
							Value: BinaryOpExpr{
								Left:  VariableExpr{Name: "result"},
								Op:    Add,
								Right: VariableExpr{Name: "j"},
							},
						},
					},
				},
			},
		},
	}

	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, int64(12), result)
}

func TestBreak_InnerWhileLoop_OuterForContinues(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// $outerIterations = 0
	// for i in [1, 2, 3] {
	//   $outerIterations = $outerIterations + 1
	//   $counter = 0
	//   while true {
	//     $counter = $counter + 1
	//     if $counter == 2 { break }
	//   }
	// }
	// outerIterations should be 3 (outer completes all iterations despite inner breaks)
	env.Define("outerIterations", int64(0))
	env.Define("counter", int64(0))

	stmts := []Statement{
		ForStatement{
			ValueVar: "i",
			Iterable: ArrayExpr{
				Elements: []Expr{
					LiteralExpr{Value: IntLiteral{Value: 1}},
					LiteralExpr{Value: IntLiteral{Value: 2}},
					LiteralExpr{Value: IntLiteral{Value: 3}},
				},
			},
			Body: []Statement{
				ReassignStatement{
					Target: "outerIterations",
					Value: BinaryOpExpr{
						Left:  VariableExpr{Name: "outerIterations"},
						Op:    Add,
						Right: LiteralExpr{Value: IntLiteral{Value: 1}},
					},
				},
				ReassignStatement{
					Target: "counter",
					Value:  LiteralExpr{Value: IntLiteral{Value: 0}},
				},
				WhileStatement{
					Condition: LiteralExpr{Value: BoolLiteral{Value: true}},
					Body: []Statement{
						ReassignStatement{
							Target: "counter",
							Value: BinaryOpExpr{
								Left:  VariableExpr{Name: "counter"},
								Op:    Add,
								Right: LiteralExpr{Value: IntLiteral{Value: 1}},
							},
						},
						IfStatement{
							Condition: BinaryOpExpr{
								Left:  VariableExpr{Name: "counter"},
								Op:    Eq,
								Right: LiteralExpr{Value: IntLiteral{Value: 2}},
							},
							ThenBlock: []Statement{
								BreakStatement{},
							},
						},
					},
				},
			},
		},
	}

	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	outerIterations, err := env.Get("outerIterations")
	require.NoError(t, err)
	assert.Equal(t, int64(3), outerIterations)
}

// ---------------------------------------------------------------------------
// Array callback error propagation tests
// ---------------------------------------------------------------------------

// helper: creates a lambda closure that divides 10 by n (will error on n==0)
func divByNLambda(env *Environment) *LambdaClosure {
	return &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "n", Required: true}},
			Body: BinaryOpExpr{
				Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
				Op:    Div,
				Right: VariableExpr{Name: "n"},
			},
		},
		Env: env,
	}
}

// helper: creates a lambda closure that references an undefined variable
func undefinedVarLambda(env *Environment) *LambdaClosure {
	return &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "n", Required: true}},
			Body:   VariableExpr{Name: "nonexistent"},
		},
		Env: env,
	}
}

func TestMap_DivisionByZeroInCallback(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(2), int64(0), int64(5)}
	env.Define("arr", arr)
	env.Define("fn", divByNLambda(env))

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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "map() callback error at index 1")
	assert.Contains(t, err.Error(), "division by zero")
}

func TestFilter_ErrorInCallback(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(2), int64(3)}
	env.Define("arr", arr)
	env.Define("fn", undefinedVarLambda(env))

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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "filter() callback error at index 0")
}

func TestReduce_ErrorInCallback(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Lambda that divides acc by elem. When elem is 0, it will error.
	divLambda := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{
				{Name: "acc", Required: true},
				{Name: "elem", Required: true},
			},
			Body: BinaryOpExpr{
				Left:  VariableExpr{Name: "acc"},
				Op:    Div,
				Right: VariableExpr{Name: "elem"},
			},
		},
		Env: env,
	}

	arr := []interface{}{int64(2), int64(0), int64(3)}
	env.Define("arr", arr)
	env.Define("fn", divLambda)

	stmts := []Statement{
		AssignStatement{
			Target: "result",
			Value: FunctionCallExpr{
				Name: "reduce",
				Args: []Expr{
					VariableExpr{Name: "arr"},
					VariableExpr{Name: "fn"},
					LiteralExpr{Value: IntLiteral{Value: 100}},
				},
			},
		},
	}
	_, err := interp.executeStatements(stmts, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reduce() callback error at index 1")
	assert.Contains(t, err.Error(), "division by zero")
}

func TestFind_ErrorInCallback(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(2)}
	env.Define("arr", arr)
	env.Define("fn", undefinedVarLambda(env))

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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "find() callback error at index 0")
}

func TestSome_ErrorInCallback(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(2)}
	env.Define("arr", arr)
	env.Define("fn", undefinedVarLambda(env))

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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "some() callback error at index 0")
}

func TestEvery_ErrorInCallback(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(1), int64(2)}
	env.Define("arr", arr)
	env.Define("fn", undefinedVarLambda(env))

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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "every() callback error at index 0")
}

// ---------------------------------------------------------------------------
// Pattern matching guard edge cases
// ---------------------------------------------------------------------------

func TestMatchExpr_GuardNonBoolean_ReturnsError(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// match 5 { x when 42 => "yes", _ => "no" }
	// Guard evaluates to 42 (an integer, not a boolean) - should error
	expr := MatchExpr{
		Value: LiteralExpr{Value: IntLiteral{Value: 5}},
		Cases: []MatchCase{
			{
				Pattern: VariablePattern{Name: "x"},
				Guard:   LiteralExpr{Value: IntLiteral{Value: 42}},
				Body:    LiteralExpr{Value: StringLiteral{Value: "yes"}},
			},
			{
				Pattern: WildcardPattern{},
				Body:    LiteralExpr{Value: StringLiteral{Value: "no"}},
			},
		},
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "match guard must evaluate to boolean")
}

func TestMatchExpr_GuardNonBoolean_String_ReturnsError(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// match 10 { x when "truthy" => "yes", _ => "no" }
	// Guard evaluates to a string, not boolean
	expr := MatchExpr{
		Value: LiteralExpr{Value: IntLiteral{Value: 10}},
		Cases: []MatchCase{
			{
				Pattern: VariablePattern{Name: "x"},
				Guard:   LiteralExpr{Value: StringLiteral{Value: "truthy"}},
				Body:    LiteralExpr{Value: StringLiteral{Value: "yes"}},
			},
			{
				Pattern: WildcardPattern{},
				Body:    LiteralExpr{Value: StringLiteral{Value: "no"}},
			},
		},
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "match guard must evaluate to boolean")
}

func TestMatchExpr_GuardWithRuntimeError(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// match 5 { x when undefinedVar > 3 => "yes", _ => "no" }
	// Guard references an undefined variable - should propagate the error
	expr := MatchExpr{
		Value: LiteralExpr{Value: IntLiteral{Value: 5}},
		Cases: []MatchCase{
			{
				Pattern: VariablePattern{Name: "x"},
				Guard: BinaryOpExpr{
					Left:  VariableExpr{Name: "undefinedVar"},
					Op:    Gt,
					Right: LiteralExpr{Value: IntLiteral{Value: 3}},
				},
				Body: LiteralExpr{Value: StringLiteral{Value: "yes"}},
			},
			{
				Pattern: WildcardPattern{},
				Body:    LiteralExpr{Value: StringLiteral{Value: "no"}},
			},
		},
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "undefinedVar")
}

func TestMatchExpr_GuardDivisionByZero(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// match 5 { x when (x / 0) == 1 => "yes", _ => "no" }
	// Guard triggers a division by zero error
	expr := MatchExpr{
		Value: LiteralExpr{Value: IntLiteral{Value: 5}},
		Cases: []MatchCase{
			{
				Pattern: VariablePattern{Name: "x"},
				Guard: BinaryOpExpr{
					Left: BinaryOpExpr{
						Left:  VariableExpr{Name: "x"},
						Op:    Div,
						Right: LiteralExpr{Value: IntLiteral{Value: 0}},
					},
					Op:    Eq,
					Right: LiteralExpr{Value: IntLiteral{Value: 1}},
				},
				Body: LiteralExpr{Value: StringLiteral{Value: "yes"}},
			},
			{
				Pattern: WildcardPattern{},
				Body:    LiteralExpr{Value: StringLiteral{Value: "no"}},
			},
		},
	}

	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "division by zero")
}
