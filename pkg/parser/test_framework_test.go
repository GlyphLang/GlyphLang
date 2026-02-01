package parser

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLexerTestKeyword verifies the lexer tokenizes test-related keywords
func TestLexerTestKeyword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "test is IDENT to avoid path conflicts",
			input:    "test",
			expected: []TokenType{IDENT},
		},
		{
			name:     "test with string name",
			input:    `test "my test"`,
			expected: []TokenType{IDENT, STRING},
		},
		{
			name:     "assert keyword",
			input:    "assert",
			expected: []TokenType{ASSERT},
		},
		{
			name:     "assert with parens",
			input:    "assert(true)",
			expected: []TokenType{ASSERT, LPAREN, TRUE, RPAREN},
		},
		{
			name:     "test in route path stays as path",
			input:    "@ GET /test",
			expected: []TokenType{AT, IDENT, SLASH, IDENT},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			var actual []TokenType
			for _, tok := range tokens {
				if tok.Type != EOF {
					actual = append(actual, tok.Type)
				}
			}
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestExpandedLexerTestKeyword verifies the expanded lexer tokenizes test/assert
func TestExpandedLexerTestKeyword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "test is IDENT in expanded lexer too",
			input:    "test",
			expected: []TokenType{IDENT},
		},
		{
			name:     "assert keyword",
			input:    "assert",
			expected: []TokenType{ASSERT},
		},
		{
			name:     "test with string",
			input:    `test "example test"`,
			expected: []TokenType{IDENT, STRING},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewExpandedLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			var actual []TokenType
			for _, tok := range tokens {
				if tok.Type != EOF {
					actual = append(actual, tok.Type)
				}
			}
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestParseTestBlock verifies parsing a simple test block
func TestParseTestBlock(t *testing.T) {
	input := `test "should pass" {
	assert(true)
}`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := NewParserWithSource(tokens, input)
	module, err := p.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	tb, ok := module.Items[0].(*ast.TestBlock)
	require.True(t, ok, "expected TestBlock, got %T", module.Items[0])
	assert.Equal(t, "should pass", tb.Name)
	assert.Len(t, tb.Body, 1)

	assertStmt, ok := tb.Body[0].(ast.AssertStatement)
	require.True(t, ok, "expected AssertStatement, got %T", tb.Body[0])
	assert.Nil(t, assertStmt.Message)
}

// TestParseTestBlockWithMessage verifies assert with custom message
func TestParseTestBlockWithMessage(t *testing.T) {
	input := `test "with message" {
	assert(false, "should be true")
}`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := NewParserWithSource(tokens, input)
	module, err := p.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	tb := module.Items[0].(*ast.TestBlock)
	assert.Equal(t, "with message", tb.Name)

	assertStmt := tb.Body[0].(ast.AssertStatement)
	assert.NotNil(t, assertStmt.Message)
}

// TestParseTestBlockWithVariables verifies test with variable assignments
func TestParseTestBlockWithVariables(t *testing.T) {
	input := `test "variable test" {
	$ x = 42
	assert(x == 42)
}`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := NewParserWithSource(tokens, input)
	module, err := p.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	tb := module.Items[0].(*ast.TestBlock)
	assert.Equal(t, "variable test", tb.Name)
	assert.Len(t, tb.Body, 2)

	// First statement is AssignStatement
	_, ok := tb.Body[0].(ast.AssignStatement)
	assert.True(t, ok)

	// Second is AssertStatement
	_, ok = tb.Body[1].(ast.AssertStatement)
	assert.True(t, ok)
}

// TestParseMultipleTestBlocks verifies parsing multiple test blocks
func TestParseMultipleTestBlocks(t *testing.T) {
	input := `test "first" {
	assert(true)
}

test "second" {
	assert(1 == 1)
}`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := NewParserWithSource(tokens, input)
	module, err := p.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 2)

	tb1 := module.Items[0].(*ast.TestBlock)
	assert.Equal(t, "first", tb1.Name)

	tb2 := module.Items[1].(*ast.TestBlock)
	assert.Equal(t, "second", tb2.Name)
}

// TestParseTestBlockWithExpression verifies test with comparison expressions
func TestParseTestBlockWithExpression(t *testing.T) {
	input := `test "math" {
	$ result = 2 + 3
	assert(result == 5)
	assert(result != 0)
	assert(result > 4)
	assert(result < 6)
}`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := NewParserWithSource(tokens, input)
	module, err := p.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	tb := module.Items[0].(*ast.TestBlock)
	assert.Len(t, tb.Body, 5) // 1 assign + 4 asserts
}

// TestParseTestBlockExpandedSyntax verifies parsing test blocks in expanded syntax
func TestParseTestBlockExpandedSyntax(t *testing.T) {
	input := `test "expanded test" {
	let x = 10
	assert(x == 10)
}`
	lexer := NewExpandedLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := NewParserWithSource(tokens, input)
	module, err := p.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	tb := module.Items[0].(*ast.TestBlock)
	assert.Equal(t, "expanded test", tb.Name)
	assert.Len(t, tb.Body, 2)
}

// TestParseTestBlockMissingName verifies error for test without name
func TestParseTestBlockMissingName(t *testing.T) {
	input := `test {
	assert(true)
}`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := NewParserWithSource(tokens, input)
	_, err = p.Parse()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected test name")
}

// TestParseTestBlockMissingBrace verifies error for test without opening brace
func TestParseTestBlockMissingBrace(t *testing.T) {
	input := `test "name" assert(true)`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := NewParserWithSource(tokens, input)
	_, err = p.Parse()
	assert.Error(t, err)
}

// TestParseTestWithFunction verifies test blocks coexist with functions
func TestParseTestWithFunction(t *testing.T) {
	// In compact syntax, functions use ! with parens: ! name(param: type) { body }
	input := `! double(n: int) {
	> n * 2
}

test "double works" {
	$ result = double(5)
	assert(result == 10)
}`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := NewParserWithSource(tokens, input)
	module, err := p.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 2)
	_, ok := module.Items[0].(*ast.Function)
	assert.True(t, ok)
	_, ok = module.Items[1].(*ast.TestBlock)
	assert.True(t, ok)
}

// TestParseAssertMissingParen verifies error for assert without parens
func TestParseAssertMissingParen(t *testing.T) {
	input := `test "bad assert" {
	assert true
}`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := NewParserWithSource(tokens, input)
	_, err = p.Parse()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "(")
}
