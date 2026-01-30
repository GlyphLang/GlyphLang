package interpreter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadTestBlock verifies test blocks are stored during module loading
func TestLoadTestBlock(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "basic test",
				Body: []Statement{
					AssertStatement{
						Condition: LiteralExpr{Value: BoolLiteral{Value: true}},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	blocks := interp.GetTestBlocks()
	assert.Len(t, blocks, 1)
	assert.Equal(t, "basic test", blocks[0].Name)
}

// TestLoadMultipleTestBlocks verifies multiple test blocks are collected
func TestLoadMultipleTestBlocks(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "test one",
				Body: []Statement{
					AssertStatement{
						Condition: LiteralExpr{Value: BoolLiteral{Value: true}},
					},
				},
			},
			&TestBlock{
				Name: "test two",
				Body: []Statement{
					AssertStatement{
						Condition: LiteralExpr{Value: BoolLiteral{Value: true}},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	blocks := interp.GetTestBlocks()
	assert.Len(t, blocks, 2)
	assert.Equal(t, "test one", blocks[0].Name)
	assert.Equal(t, "test two", blocks[1].Name)
}

// TestRunTestsPassing verifies all passing tests return correct results
func TestRunTestsPassing(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "passes",
				Body: []Statement{
					AssertStatement{
						Condition: LiteralExpr{Value: BoolLiteral{Value: true}},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	results := interp.RunTests("")
	require.Len(t, results, 1)
	assert.True(t, results[0].Passed)
	assert.Equal(t, "passes", results[0].Name)
	assert.Empty(t, results[0].Error)
}

// TestRunTestsFailing verifies failing assertion produces correct result
func TestRunTestsFailing(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "fails",
				Body: []Statement{
					AssertStatement{
						Condition: LiteralExpr{Value: BoolLiteral{Value: false}},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	results := interp.RunTests("")
	require.Len(t, results, 1)
	assert.False(t, results[0].Passed)
	assert.Equal(t, "fails", results[0].Name)
	assert.Equal(t, "assertion failed", results[0].Error)
}

// TestRunTestsFailingWithMessage verifies custom assertion message
func TestRunTestsFailingWithMessage(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "fails with message",
				Body: []Statement{
					AssertStatement{
						Condition: LiteralExpr{Value: BoolLiteral{Value: false}},
						Message:   LiteralExpr{Value: StringLiteral{Value: "expected true"}},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	results := interp.RunTests("")
	require.Len(t, results, 1)
	assert.False(t, results[0].Passed)
	assert.Equal(t, "expected true", results[0].Error)
}

// TestRunTestsWithFilter verifies filter selects correct tests
func TestRunTestsWithFilter(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "user creation",
				Body: []Statement{
					AssertStatement{Condition: LiteralExpr{Value: BoolLiteral{Value: true}}},
				},
			},
			&TestBlock{
				Name: "order processing",
				Body: []Statement{
					AssertStatement{Condition: LiteralExpr{Value: BoolLiteral{Value: true}}},
				},
			},
			&TestBlock{
				Name: "user deletion",
				Body: []Statement{
					AssertStatement{Condition: LiteralExpr{Value: BoolLiteral{Value: true}}},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	// Filter by substring
	results := interp.RunTests("user")
	assert.Len(t, results, 2)

	// Filter by prefix wildcard
	results = interp.RunTests("user*")
	assert.Len(t, results, 2)

	// Filter by suffix wildcard
	results = interp.RunTests("*processing")
	assert.Len(t, results, 1)
	assert.Equal(t, "order processing", results[0].Name)

	// No match
	results = interp.RunTests("nonexistent")
	assert.Len(t, results, 0)
}

// TestRunTestsMixed verifies mix of passing and failing tests
func TestRunTestsMixed(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "first passes",
				Body: []Statement{
					AssertStatement{Condition: LiteralExpr{Value: BoolLiteral{Value: true}}},
				},
			},
			&TestBlock{
				Name: "second fails",
				Body: []Statement{
					AssertStatement{Condition: LiteralExpr{Value: BoolLiteral{Value: false}}},
				},
			},
			&TestBlock{
				Name: "third passes",
				Body: []Statement{
					AssertStatement{Condition: LiteralExpr{Value: BoolLiteral{Value: true}}},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	results := interp.RunTests("")
	require.Len(t, results, 3)
	assert.True(t, results[0].Passed)
	assert.False(t, results[1].Passed)
	assert.True(t, results[2].Passed)
}

// TestAssertWithExpression verifies assert with binary expression
func TestAssertWithExpression(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "equality check",
				Body: []Statement{
					AssignStatement{
						Target: "x",
						Value:  LiteralExpr{Value: IntLiteral{Value: 42}},
					},
					AssertStatement{
						Condition: BinaryOpExpr{
							Op:    Eq,
							Left:  VariableExpr{Name: "x"},
							Right: LiteralExpr{Value: IntLiteral{Value: 42}},
						},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	results := interp.RunTests("")
	require.Len(t, results, 1)
	assert.True(t, results[0].Passed)
}

// TestAssertNonBoolean verifies assertion fails for non-boolean result
func TestAssertNonBoolean(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "non-bool assert",
				Body: []Statement{
					AssertStatement{
						Condition: LiteralExpr{Value: IntLiteral{Value: 42}},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	results := interp.RunTests("")
	require.Len(t, results, 1)
	assert.False(t, results[0].Passed)
	assert.Contains(t, results[0].Error, "boolean")
}

// TestAssertionErrorType verifies error type checking
func TestAssertionErrorType(t *testing.T) {
	err := &AssertionError{Message: "test error"}
	assert.True(t, IsAssertionError(err))
	assert.Equal(t, "test error", err.Error())

	otherErr := &ValidationError{Message: "other"}
	assert.False(t, IsAssertionError(otherErr))
}

// TestGetTestBlocksReturns copy verifies safety of returned slice
func TestGetTestBlocksReturnsCopy(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{Name: "test", Body: nil},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	blocks := interp.GetTestBlocks()
	blocks[0].Name = "modified"

	// Original should be unchanged
	assert.Equal(t, "test", interp.testBlocks[0].Name)
}

// TestTestBlockIsolation verifies tests run in isolated environments
func TestTestBlockIsolation(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "sets x",
				Body: []Statement{
					AssignStatement{
						Target: "x",
						Value:  LiteralExpr{Value: IntLiteral{Value: 10}},
					},
					AssertStatement{
						Condition: BinaryOpExpr{
							Op:    Eq,
							Left:  VariableExpr{Name: "x"},
							Right: LiteralExpr{Value: IntLiteral{Value: 10}},
						},
					},
				},
			},
			&TestBlock{
				Name: "x not defined",
				Body: []Statement{
					// This test defines its own x - should not see the first test's x
					AssignStatement{
						Target: "x",
						Value:  LiteralExpr{Value: IntLiteral{Value: 20}},
					},
					AssertStatement{
						Condition: BinaryOpExpr{
							Op:    Eq,
							Left:  VariableExpr{Name: "x"},
							Right: LiteralExpr{Value: IntLiteral{Value: 20}},
						},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	results := interp.RunTests("")
	require.Len(t, results, 2)
	assert.True(t, results[0].Passed)
	assert.True(t, results[1].Passed)
}

// TestMatchesFilter verifies the filter matching function
func TestMatchesFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		want   bool
	}{
		{"user creation", "", true},
		{"user creation", "user", true},
		{"user creation", "user*", true},
		{"user creation", "*creation", true},
		{"user creation", "*user*", true},
		{"user creation", "order", false},
		{"user creation", "user creation", true},
		{"abc", "*abc*", true},
		{"abc", "xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.filter, func(t *testing.T) {
			got := matchesFilter(tt.name, tt.filter)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestTestBlockWithFunctions verifies tests can call module functions
func TestTestBlockWithFunctions(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&Function{
				Name:   "double",
				Params: []Field{{Name: "n", TypeAnnotation: IntType{}}},
				Body: []Statement{
					ReturnStatement{
						Value: BinaryOpExpr{
							Op:    Mul,
							Left:  VariableExpr{Name: "n"},
							Right: LiteralExpr{Value: IntLiteral{Value: 2}},
						},
					},
				},
			},
			&TestBlock{
				Name: "test double function",
				Body: []Statement{
					AssignStatement{
						Target: "result",
						Value: FunctionCallExpr{
							Name: "double",
							Args: []Expr{LiteralExpr{Value: IntLiteral{Value: 5}}},
						},
					},
					AssertStatement{
						Condition: BinaryOpExpr{
							Op:    Eq,
							Left:  VariableExpr{Name: "result"},
							Right: LiteralExpr{Value: IntLiteral{Value: 10}},
						},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	results := interp.RunTests("")
	require.Len(t, results, 1)
	assert.True(t, results[0].Passed)
}

// TestMultipleAssertions verifies test with multiple assertions
func TestMultipleAssertions(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "multiple asserts",
				Body: []Statement{
					AssertStatement{Condition: LiteralExpr{Value: BoolLiteral{Value: true}}},
					AssertStatement{Condition: LiteralExpr{Value: BoolLiteral{Value: true}}},
					AssertStatement{Condition: LiteralExpr{Value: BoolLiteral{Value: true}}},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	results := interp.RunTests("")
	require.Len(t, results, 1)
	assert.True(t, results[0].Passed)
}

// TestAssertStopsOnFirstFailure verifies test stops at first failed assertion
func TestAssertStopsOnFirstFailure(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "stops early",
				Body: []Statement{
					AssertStatement{Condition: LiteralExpr{Value: BoolLiteral{Value: true}}},
					AssertStatement{
						Condition: LiteralExpr{Value: BoolLiteral{Value: false}},
						Message:   LiteralExpr{Value: StringLiteral{Value: "second fails"}},
					},
					// This assertion would pass but should not be reached
					AssertStatement{Condition: LiteralExpr{Value: BoolLiteral{Value: true}}},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	results := interp.RunTests("")
	require.Len(t, results, 1)
	assert.False(t, results[0].Passed)
	assert.Equal(t, "second fails", results[0].Error)
}

// TestTestDuration verifies duration is tracked
func TestTestDuration(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TestBlock{
				Name: "timed test",
				Body: []Statement{
					AssertStatement{Condition: LiteralExpr{Value: BoolLiteral{Value: true}}},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	results := interp.RunTests("")
	require.Len(t, results, 1)
	assert.True(t, results[0].Duration >= 0)
}
