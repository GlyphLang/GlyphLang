package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBreak_WhileLoop(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// $count = 0
	// while true {
	//   $count = $count + 1
	//   if $count == 3 { break }
	// }
	env.Define("count", int64(0))

	stmts := []Statement{
		WhileStatement{
			Condition: LiteralExpr{Value: BoolLiteral{Value: true}},
			Body: []Statement{
				ReassignStatement{
					Target: "count",
					Value: BinaryOpExpr{
						Left:  VariableExpr{Name: "count"},
						Op:    Add,
						Right: LiteralExpr{Value: IntLiteral{Value: 1}},
					},
				},
				IfStatement{
					Condition: BinaryOpExpr{
						Left:  VariableExpr{Name: "count"},
						Op:    Eq,
						Right: LiteralExpr{Value: IntLiteral{Value: 3}},
					},
					ThenBlock: []Statement{
						BreakStatement{},
					},
				},
			},
		},
	}

	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	count, err := env.Get("count")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestContinue_WhileLoop(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// $count = 0
	// $sum = 0
	// while $count < 5 {
	//   $count = $count + 1
	//   if $count == 3 { continue }  // skip adding 3
	//   $sum = $sum + $count
	// }
	env.Define("count", int64(0))
	env.Define("sum", int64(0))

	stmts := []Statement{
		WhileStatement{
			Condition: BinaryOpExpr{
				Left:  VariableExpr{Name: "count"},
				Op:    Lt,
				Right: LiteralExpr{Value: IntLiteral{Value: 5}},
			},
			Body: []Statement{
				ReassignStatement{
					Target: "count",
					Value: BinaryOpExpr{
						Left:  VariableExpr{Name: "count"},
						Op:    Add,
						Right: LiteralExpr{Value: IntLiteral{Value: 1}},
					},
				},
				IfStatement{
					Condition: BinaryOpExpr{
						Left:  VariableExpr{Name: "count"},
						Op:    Eq,
						Right: LiteralExpr{Value: IntLiteral{Value: 3}},
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
						Right: VariableExpr{Name: "count"},
					},
				},
			},
		},
	}

	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	sum, err := env.Get("sum")
	require.NoError(t, err)
	// sum = 1 + 2 + 4 + 5 = 12 (skipping 3)
	assert.Equal(t, int64(12), sum)
}

func TestBreak_ForLoop(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// $result = 0
	// for item in [10, 20, 30, 40, 50] {
	//   if $item == 30 { break }
	//   $result = $result + $item
	// }
	env.Define("result", int64(0))

	stmts := []Statement{
		ForStatement{
			ValueVar: "item",
			Iterable: ArrayExpr{
				Elements: []Expr{
					LiteralExpr{Value: IntLiteral{Value: 10}},
					LiteralExpr{Value: IntLiteral{Value: 20}},
					LiteralExpr{Value: IntLiteral{Value: 30}},
					LiteralExpr{Value: IntLiteral{Value: 40}},
					LiteralExpr{Value: IntLiteral{Value: 50}},
				},
			},
			Body: []Statement{
				IfStatement{
					Condition: BinaryOpExpr{
						Left:  VariableExpr{Name: "item"},
						Op:    Eq,
						Right: LiteralExpr{Value: IntLiteral{Value: 30}},
					},
					ThenBlock: []Statement{
						BreakStatement{},
					},
				},
				ReassignStatement{
					Target: "result",
					Value: BinaryOpExpr{
						Left:  VariableExpr{Name: "result"},
						Op:    Add,
						Right: VariableExpr{Name: "item"},
					},
				},
			},
		},
	}

	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	// result = 10 + 20 = 30 (broke at 30)
	assert.Equal(t, int64(30), result)
}

func TestContinue_ForLoop(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// $result = 0
	// for item in [1, 2, 3, 4, 5] {
	//   if $item == 3 { continue }  // skip 3
	//   $result = $result + $item
	// }
	env.Define("result", int64(0))

	stmts := []Statement{
		ForStatement{
			ValueVar: "item",
			Iterable: ArrayExpr{
				Elements: []Expr{
					LiteralExpr{Value: IntLiteral{Value: 1}},
					LiteralExpr{Value: IntLiteral{Value: 2}},
					LiteralExpr{Value: IntLiteral{Value: 3}},
					LiteralExpr{Value: IntLiteral{Value: 4}},
					LiteralExpr{Value: IntLiteral{Value: 5}},
				},
			},
			Body: []Statement{
				IfStatement{
					Condition: BinaryOpExpr{
						Left:  VariableExpr{Name: "item"},
						Op:    Eq,
						Right: LiteralExpr{Value: IntLiteral{Value: 3}},
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
						Right: VariableExpr{Name: "item"},
					},
				},
			},
		},
	}

	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	// result = 1 + 2 + 4 + 5 = 12 (skipping 3)
	assert.Equal(t, int64(12), result)
}

func TestBreak_ImmediateInWhile(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// while true { break }
	// Should not loop forever
	stmts := []Statement{
		WhileStatement{
			Condition: LiteralExpr{Value: BoolLiteral{Value: true}},
			Body: []Statement{
				BreakStatement{},
			},
		},
	}

	_, err := interp.executeStatements(stmts, env)
	require.NoError(t, err)
}
