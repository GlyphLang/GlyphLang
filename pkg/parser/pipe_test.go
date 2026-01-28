package parser

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test lexer tokenizes pipe operator |>
func TestLexer_PipeOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "simple pipe operator",
			input:    "|>",
			expected: []TokenType{PIPE_OP},
		},
		{
			name:     "pipe operator with spaces",
			input:    "x |> f",
			expected: []TokenType{IDENT, PIPE_OP, IDENT},
		},
		{
			name:     "pipe operator chain",
			input:    "x |> f |> g",
			expected: []TokenType{IDENT, PIPE_OP, IDENT, PIPE_OP, IDENT},
		},
		{
			name:     "pipe vs union type",
			input:    "a | b |> f",
			expected: []TokenType{IDENT, PIPE, IDENT, PIPE_OP, IDENT},
		},
		{
			name:     "pipe vs or operator",
			input:    "a || b |> f",
			expected: []TokenType{IDENT, OR, IDENT, PIPE_OP, IDENT},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			// Filter out EOF token for comparison
			var actualTokens []TokenType
			for _, tok := range tokens {
				if tok.Type != EOF {
					actualTokens = append(actualTokens, tok.Type)
				}
			}

			require.Equal(t, len(tt.expected), len(actualTokens), "token count mismatch")

			for i, expectedType := range tt.expected {
				assert.Equal(t, expectedType, actualTokens[i], "token %d type mismatch", i)
			}
		})
	}
}

// Test expanded lexer also tokenizes pipe operator
func TestExpandedLexer_PipeOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "simple pipe operator",
			input:    "|>",
			expected: []TokenType{PIPE_OP},
		},
		{
			name:     "pipe operator in expression",
			input:    "x |> func",
			expected: []TokenType{IDENT, PIPE_OP, EQUALS}, // "func" maps to EQUALS in expanded lexer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewExpandedLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			var actualTokens []TokenType
			for _, tok := range tokens {
				if tok.Type != EOF {
					actualTokens = append(actualTokens, tok.Type)
				}
			}

			require.Equal(t, len(tt.expected), len(actualTokens), "token count mismatch")

			for i, expectedType := range tt.expected {
				assert.Equal(t, expectedType, actualTokens[i], "token %d type mismatch", i)
			}
		})
	}
}

// Test parser parses pipe expressions
func TestParser_PipeExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkAST func(t *testing.T, module *interpreter.Module)
	}{
		{
			name: "simple pipe to function",
			input: `@ GET /test {
				> x |> f
			}`,
			checkAST: func(t *testing.T, module *interpreter.Module) {
				require.Len(t, module.Items, 1)
				route, ok := module.Items[0].(*interpreter.Route)
				require.True(t, ok)
				require.Len(t, route.Body, 1)

				ret, ok := route.Body[0].(interpreter.ReturnStatement)
				require.True(t, ok)

				pipe, ok := ret.Value.(interpreter.PipeExpr)
				require.True(t, ok, "expected PipeExpr, got %T", ret.Value)

				// Left should be variable x
				left, ok := pipe.Left.(interpreter.VariableExpr)
				require.True(t, ok)
				assert.Equal(t, "x", left.Name)

				// Right should be variable f
				right, ok := pipe.Right.(interpreter.VariableExpr)
				require.True(t, ok)
				assert.Equal(t, "f", right.Name)
			},
		},
		{
			name: "pipe to function call with args",
			input: `@ GET /test {
				> x |> f(a, b)
			}`,
			checkAST: func(t *testing.T, module *interpreter.Module) {
				require.Len(t, module.Items, 1)
				route, ok := module.Items[0].(*interpreter.Route)
				require.True(t, ok)
				require.Len(t, route.Body, 1)

				ret, ok := route.Body[0].(interpreter.ReturnStatement)
				require.True(t, ok)

				pipe, ok := ret.Value.(interpreter.PipeExpr)
				require.True(t, ok, "expected PipeExpr, got %T", ret.Value)

				// Left should be variable x
				left, ok := pipe.Left.(interpreter.VariableExpr)
				require.True(t, ok)
				assert.Equal(t, "x", left.Name)

				// Right should be function call f(a, b)
				right, ok := pipe.Right.(interpreter.FunctionCallExpr)
				require.True(t, ok, "expected FunctionCallExpr, got %T", pipe.Right)
				assert.Equal(t, "f", right.Name)
				assert.Len(t, right.Args, 2)
			},
		},
		{
			name: "chained pipes",
			input: `@ GET /test {
				> x |> f |> g |> h
			}`,
			checkAST: func(t *testing.T, module *interpreter.Module) {
				require.Len(t, module.Items, 1)
				route, ok := module.Items[0].(*interpreter.Route)
				require.True(t, ok)
				require.Len(t, route.Body, 1)

				ret, ok := route.Body[0].(interpreter.ReturnStatement)
				require.True(t, ok)

				// Should be: ((x |> f) |> g) |> h (left associative)
				pipe1, ok := ret.Value.(interpreter.PipeExpr)
				require.True(t, ok, "expected PipeExpr, got %T", ret.Value)

				// Right should be h
				rightH, ok := pipe1.Right.(interpreter.VariableExpr)
				require.True(t, ok)
				assert.Equal(t, "h", rightH.Name)

				// Left should be (x |> f) |> g
				pipe2, ok := pipe1.Left.(interpreter.PipeExpr)
				require.True(t, ok)

				rightG, ok := pipe2.Right.(interpreter.VariableExpr)
				require.True(t, ok)
				assert.Equal(t, "g", rightG.Name)

				// Left should be x |> f
				pipe3, ok := pipe2.Left.(interpreter.PipeExpr)
				require.True(t, ok)

				leftX, ok := pipe3.Left.(interpreter.VariableExpr)
				require.True(t, ok)
				assert.Equal(t, "x", leftX.Name)

				rightF, ok := pipe3.Right.(interpreter.VariableExpr)
				require.True(t, ok)
				assert.Equal(t, "f", rightF.Name)
			},
		},
		{
			name: "pipe with arithmetic expression on left",
			input: `@ GET /test {
				> x + 1 |> f
			}`,
			checkAST: func(t *testing.T, module *interpreter.Module) {
				require.Len(t, module.Items, 1)
				route, ok := module.Items[0].(*interpreter.Route)
				require.True(t, ok)
				require.Len(t, route.Body, 1)

				ret, ok := route.Body[0].(interpreter.ReturnStatement)
				require.True(t, ok)

				pipe, ok := ret.Value.(interpreter.PipeExpr)
				require.True(t, ok, "expected PipeExpr, got %T", ret.Value)

				// Left should be binary expression (x + 1)
				binOp, ok := pipe.Left.(interpreter.BinaryOpExpr)
				require.True(t, ok, "expected BinaryOpExpr on left, got %T", pipe.Left)
				assert.Equal(t, interpreter.Add, binOp.Op)

				// Right should be f
				right, ok := pipe.Right.(interpreter.VariableExpr)
				require.True(t, ok)
				assert.Equal(t, "f", right.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			module, err := parser.Parse()
			require.NoError(t, err)

			tt.checkAST(t, module)
		})
	}
}

// Test pipe operator precedence - should be lower than all other operators
func TestParser_PipePrecedence(t *testing.T) {
	// Pipe should have the lowest precedence
	// So "a || b |> f" should parse as "(a || b) |> f"
	// And "a + b |> f" should parse as "(a + b) |> f"

	input := `@ GET /test {
		> a || b |> f
	}`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)
	require.Len(t, route.Body, 1)

	ret, ok := route.Body[0].(interpreter.ReturnStatement)
	require.True(t, ok)

	// Should be (a || b) |> f
	pipe, ok := ret.Value.(interpreter.PipeExpr)
	require.True(t, ok, "expected PipeExpr, got %T", ret.Value)

	// Left should be a || b
	binOp, ok := pipe.Left.(interpreter.BinaryOpExpr)
	require.True(t, ok, "expected BinaryOpExpr, got %T", pipe.Left)
	assert.Equal(t, interpreter.Or, binOp.Op)

	// Right should be f
	right, ok := pipe.Right.(interpreter.VariableExpr)
	require.True(t, ok)
	assert.Equal(t, "f", right.Name)
}
