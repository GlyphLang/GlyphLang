package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test all basic token types
func TestLexer_AllTokenTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "symbols",
			input:    "@ : $ + - * / % > < !",
			expected: []TokenType{AT, COLON, DOLLAR, PLUS, MINUS, STAR, SLASH, PERCENT, GREATER, LESS, BANG},
		},
		{
			name:     "comparison operators",
			input:    ">= <= == !=",
			expected: []TokenType{GREATER_EQ, LESS_EQ, EQ_EQ, NOT_EQ},
		},
		{
			name:     "logical operators",
			input:    "&& ||",
			expected: []TokenType{AND, OR},
		},
		{
			name:     "delimiters",
			input:    "( ) { } [ ] , . ->",
			expected: []TokenType{LPAREN, RPAREN, LBRACE, RBRACE, LBRACKET, RBRACKET, COMMA, DOT, ARROW},
		},
		{
			name:     "pipe and tilde",
			input:    "| ~ &",
			expected: []TokenType{PIPE, TILDE, AMPERSAND},
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

// Test string escaping
func TestLexer_StringEscaping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    `"hello world"`,
			expected: "hello world",
		},
		{
			name:     "single quotes",
			input:    `'hello world'`,
			expected: "hello world",
		},
		{
			name:     "escaped newline",
			input:    `"hello\nworld"`,
			expected: "hello\nworld",
		},
		{
			name:     "escaped tab",
			input:    `"hello\tworld"`,
			expected: "hello\tworld",
		},
		{
			name:     "escaped carriage return",
			input:    `"hello\rworld"`,
			expected: "hello\rworld",
		},
		{
			name:     "escaped double quote",
			input:    `"say \"hello\""`,
			expected: `say "hello"`,
		},
		{
			name:     "escaped single quote",
			input:    `'it\'s working'`,
			expected: `it's working`,
		},
		{
			name:     "escaped backslash",
			input:    `"path\\to\\file"`,
			expected: `path\to\file`,
		},
		{
			name:     "multiple escapes",
			input:    `"line1\nline2\ttab\r\nend"`,
			expected: "line1\nline2\ttab\r\nend",
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(tokens), 1)
			assert.Equal(t, STRING, tokens[0].Type)
			assert.Equal(t, tt.expected, tokens[0].Literal)
		})
	}
}

// Test number parsing (int vs float)
func TestLexer_NumberParsing(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"integer", "42", INTEGER, "42"},
		{"zero", "0", INTEGER, "0"},
		{"large integer", "123456789", INTEGER, "123456789"},
		{"float", "3.14", FLOAT, "3.14"},
		{"float starting with zero", "0.5", FLOAT, "0.5"},
		{"float with many decimals", "123.456789", FLOAT, "123.456789"},
		{"float zero", "0.0", FLOAT, "0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(tokens), 1)
			assert.Equal(t, tt.expectedType, tokens[0].Type)
			assert.Equal(t, tt.expectedLit, tokens[0].Literal)
		})
	}
}

// Test path vs division disambiguation
func TestLexer_PathVsDivision(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "simple path",
			input:    "/api/users",
			expected: []TokenType{IDENT},
		},
		{
			name:     "path with param",
			input:    "/api/users/:id",
			expected: []TokenType{IDENT},
		},
		{
			name:     "path with multiple params",
			input:    "/posts/:postId/comments/:commentId",
			expected: []TokenType{IDENT},
		},
		{
			name:     "division after number",
			input:    "100 / 5",
			expected: []TokenType{INTEGER, SLASH, INTEGER},
		},
		{
			name:     "division after identifier",
			input:    "x / y",
			expected: []TokenType{IDENT, SLASH, IDENT},
		},
		{
			name:     "division in expression",
			input:    "price / discount",
			expected: []TokenType{IDENT, SLASH, IDENT},
		},
		{
			name:     "path in route",
			input:    "@ route /hello",
			expected: []TokenType{AT, IDENT, IDENT},
		},
		{
			name:     "rate limit with division",
			input:    "100/min",
			expected: []TokenType{INTEGER, SLASH, IDENT},
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

// Test comments
func TestLexer_Comments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "single line comment",
			input:    "# This is a comment\n42",
			expected: []TokenType{NEWLINE, INTEGER},
		},
		{
			name:     "comment at end",
			input:    "42 # end comment",
			expected: []TokenType{INTEGER},
		},
		{
			name:     "multiple comments",
			input:    "# comment 1\n42\n# comment 2\n100",
			expected: []TokenType{NEWLINE, INTEGER, NEWLINE, NEWLINE, INTEGER},
		},
		{
			name:     "comment with symbols",
			input:    "# @ $ : + - * /\n42",
			expected: []TokenType{NEWLINE, INTEGER},
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

// Test boolean literals
func TestLexer_BooleanLiterals(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"true", TRUE, "true"},
		{"false", FALSE, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(tokens), 1)
			assert.Equal(t, tt.expectedType, tokens[0].Type)
			assert.Equal(t, tt.expectedLit, tokens[0].Literal)
		})
	}
}

// Test identifiers
func TestLexer_Identifiers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple identifier", "user", "user"},
		{"with underscore", "user_name", "user_name"},
		{"with numbers", "user123", "user123"},
		{"camelCase", "userName", "userName"},
		{"PascalCase", "UserName", "UserName"},
		{"single letter", "x", "x"},
		{"single underscore prefix", "_private", "_private"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(tokens), 1)
			assert.Equal(t, IDENT, tokens[0].Type)
			assert.Equal(t, tt.expected, tokens[0].Literal)
		})
	}
}

// Test newlines and whitespace handling
func TestLexer_WhitespaceAndNewlines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "spaces between tokens",
			input:    "@ route /hello",
			expected: []TokenType{AT, IDENT, IDENT},
		},
		{
			name:     "tabs between tokens",
			input:    "@\troute\t/hello",
			expected: []TokenType{AT, IDENT, IDENT},
		},
		{
			name:     "newlines create NEWLINE tokens",
			input:    "42\n100",
			expected: []TokenType{INTEGER, NEWLINE, INTEGER},
		},
		{
			name:     "multiple newlines",
			input:    "42\n\n\n100",
			expected: []TokenType{INTEGER, NEWLINE, NEWLINE, NEWLINE, INTEGER},
		},
		{
			name:     "mixed whitespace",
			input:    "42  \t  100",
			expected: []TokenType{INTEGER, INTEGER},
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

// Test complex real-world examples
func TestLexer_ComplexExamples(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "route definition",
			input: `@ route /api/users/:id [GET]
  > {id: id, name: "test"}`,
		},
		{
			name: "type definition",
			input: `: User {
  id: int!
  name: str!
  email: str
}`,
		},
		{
			name:  "arithmetic expression",
			input: `$ result = (10 + 20) * 5 / 2 - 1`,
		},
		{
			name: "object literal",
			input: `{
  text: "Hello, World!",
  count: 42,
  active: true,
  score: 98.5
}`,
		},
		{
			name:  "function call",
			input: `$ user = db.users.get(id)`,
		},
		{
			name:  "auth middleware",
			input: `+ auth(jwt, role: admin)`,
		},
		{
			name:  "rate limit",
			input: `+ ratelimit(100/min)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			require.Greater(t, len(tokens), 0, "should produce at least EOF token")
			assert.Equal(t, EOF, tokens[len(tokens)-1].Type, "last token should be EOF")

			// Check no illegal tokens
			for i, tok := range tokens {
				assert.NotEqual(t, ILLEGAL, tok.Type, "token %d should not be ILLEGAL", i)
			}
		})
	}
}

// Test line and column tracking
func TestLexer_LineAndColumnTracking(t *testing.T) {
	input := `@ route /hello
  $ x = 42
  > x`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	// Check that line numbers are tracked
	assert.Equal(t, 1, tokens[0].Line, "first token should be on line 1")

	// Find the DOLLAR token which should be on line 2
	for _, tok := range tokens {
		if tok.Type == DOLLAR {
			assert.Equal(t, 2, tok.Line, "DOLLAR token should be on line 2")
			break
		}
	}
}

// Test edge cases and error conditions
func TestLexer_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{"empty input", "", false},
		{"only whitespace", "   \t\n  ", false},
		{"only comments", "# comment\n# another", false},
		{"unclosed string", `"hello`, true}, // Lexer should error on unterminated strings
		{"single @", "@", false},
		{"single number", "42", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Greater(t, len(tokens), 0, "should produce at least EOF token")
			}
		})
	}
}

// Test literal values are preserved
func TestLexer_LiteralPreservation(t *testing.T) {
	input := `@ route /greet/:name
  $ msg = "Hello, " + name + "!"
  > {text: msg, count: 123, score: 45.67, active: true}`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	// Check that specific literals are preserved
	literals := make(map[TokenType][]string)
	for _, tok := range tokens {
		if tok.Type == STRING || tok.Type == INTEGER || tok.Type == FLOAT || tok.Type == IDENT {
			literals[tok.Type] = append(literals[tok.Type], tok.Literal)
		}
	}

	// Verify some expected literals
	assert.Contains(t, literals[IDENT], "route")
	assert.Contains(t, literals[IDENT], "msg")
	assert.Contains(t, literals[IDENT], "name")
	assert.Contains(t, literals[STRING], "Hello, ")
	assert.Contains(t, literals[STRING], "!")
	assert.Contains(t, literals[INTEGER], "123")
	assert.Contains(t, literals[FLOAT], "45.67")
}

// Test for loop keywords
func TestLexer_ForLoopKeywords(t *testing.T) {
	input := "for item in items"
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()

	require.NoError(t, err)
	require.Equal(t, 5, len(tokens)) // FOR, IDENT, IN, IDENT, EOF

	assert.Equal(t, FOR, tokens[0].Type)
	assert.Equal(t, "for", tokens[0].Literal)

	assert.Equal(t, IDENT, tokens[1].Type)
	assert.Equal(t, "item", tokens[1].Literal)

	assert.Equal(t, IN, tokens[2].Type)
	assert.Equal(t, "in", tokens[2].Literal)

	assert.Equal(t, IDENT, tokens[3].Type)
	assert.Equal(t, "items", tokens[3].Literal)
}

func TestLexer_ForLoopWithComma(t *testing.T) {
	input := "for index, value in array"
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()

	require.NoError(t, err)
	require.Equal(t, 7, len(tokens)) // FOR, IDENT, COMMA, IDENT, IN, IDENT, EOF

	assert.Equal(t, FOR, tokens[0].Type)
	assert.Equal(t, IDENT, tokens[1].Type)
	assert.Equal(t, "index", tokens[1].Literal)
	assert.Equal(t, COMMA, tokens[2].Type)
	assert.Equal(t, IDENT, tokens[3].Type)
	assert.Equal(t, "value", tokens[3].Literal)
	assert.Equal(t, IN, tokens[4].Type)
	assert.Equal(t, IDENT, tokens[5].Type)
	assert.Equal(t, "array", tokens[5].Literal)
}

// Benchmark lexer performance
func BenchmarkLexer_SimpleRoute(b *testing.B) {
	input := `@ route /hello
  > {message: "Hello, World!"}`

	for i := 0; i < b.N; i++ {
		lexer := NewLexer(input)
		_, _ = lexer.Tokenize()
	}
}

func BenchmarkLexer_ComplexRoute(b *testing.B) {
	input := `@ route /api/users/:id [GET]
  + auth(jwt)
  + ratelimit(100/min)
  $ user = db.users.get(id)
  if user == null:
    > {error: "User not found"}
  else:
    > user`

	for i := 0; i < b.N; i++ {
		lexer := NewLexer(input)
		_, _ = lexer.Tokenize()
	}
}
