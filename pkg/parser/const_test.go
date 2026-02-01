package parser

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test lexer recognizes const keyword
func TestLexer_ConstKeyword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "simple const keyword",
			input:    "const",
			expected: []TokenType{CONST},
		},
		{
			name:     "const with identifier",
			input:    "const MAX",
			expected: []TokenType{CONST, IDENT},
		},
		{
			name:     "const declaration syntax",
			input:    "const MAX = 100",
			expected: []TokenType{CONST, IDENT, EQUALS, INTEGER},
		},
		{
			name:     "const with type annotation",
			input:    "const PI: float = 3.14",
			expected: []TokenType{CONST, IDENT, COLON, IDENT, EQUALS, FLOAT},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
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

// Test expanded lexer also recognizes const keyword
func TestExpandedLexer_ConstKeyword(t *testing.T) {
	input := "const MAX_SIZE = 100"
	lexer := NewExpandedLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	var actualTokens []TokenType
	for _, tok := range tokens {
		if tok.Type != EOF {
			actualTokens = append(actualTokens, tok.Type)
		}
	}

	expected := []TokenType{CONST, IDENT, EQUALS, INTEGER}
	require.Equal(t, len(expected), len(actualTokens))
	for i, expectedType := range expected {
		assert.Equal(t, expectedType, actualTokens[i])
	}
}

// Test parser parses const declarations
func TestParser_ConstDecl(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkAST func(t *testing.T, module *ast.Module)
	}{
		{
			name:  "simple const declaration",
			input: "const MAX_SIZE = 100",
			checkAST: func(t *testing.T, module *ast.Module) {
				require.Len(t, module.Items, 1)
				constDecl, ok := module.Items[0].(*ast.ConstDecl)
				require.True(t, ok, "expected ConstDecl, got %T", module.Items[0])
				assert.Equal(t, "MAX_SIZE", constDecl.Name)
				assert.Nil(t, constDecl.Type)

				// Check value is integer literal
				lit, ok := constDecl.Value.(ast.LiteralExpr)
				require.True(t, ok)
				intLit, ok := lit.Value.(ast.IntLiteral)
				require.True(t, ok)
				assert.Equal(t, int64(100), intLit.Value)
			},
		},
		{
			name:  "const with string value",
			input: `const API_URL = "https://api.example.com"`,
			checkAST: func(t *testing.T, module *ast.Module) {
				require.Len(t, module.Items, 1)
				constDecl, ok := module.Items[0].(*ast.ConstDecl)
				require.True(t, ok)
				assert.Equal(t, "API_URL", constDecl.Name)

				lit, ok := constDecl.Value.(ast.LiteralExpr)
				require.True(t, ok)
				strLit, ok := lit.Value.(ast.StringLiteral)
				require.True(t, ok)
				assert.Equal(t, "https://api.example.com", strLit.Value)
			},
		},
		{
			name:  "const with type annotation",
			input: "const PI: float = 3.14159",
			checkAST: func(t *testing.T, module *ast.Module) {
				require.Len(t, module.Items, 1)
				constDecl, ok := module.Items[0].(*ast.ConstDecl)
				require.True(t, ok)
				assert.Equal(t, "PI", constDecl.Name)
				assert.NotNil(t, constDecl.Type)

				lit, ok := constDecl.Value.(ast.LiteralExpr)
				require.True(t, ok)
				floatLit, ok := lit.Value.(ast.FloatLiteral)
				require.True(t, ok)
				assert.InDelta(t, 3.14159, floatLit.Value, 0.00001)
			},
		},
		{
			name:  "const with expression value",
			input: "const DOUBLED = 50 * 2",
			checkAST: func(t *testing.T, module *ast.Module) {
				require.Len(t, module.Items, 1)
				constDecl, ok := module.Items[0].(*ast.ConstDecl)
				require.True(t, ok)
				assert.Equal(t, "DOUBLED", constDecl.Name)

				binOp, ok := constDecl.Value.(ast.BinaryOpExpr)
				require.True(t, ok, "expected BinaryOpExpr, got %T", constDecl.Value)
				assert.Equal(t, ast.Mul, binOp.Op)
			},
		},
		{
			name: "multiple const declarations",
			input: `const A = 1
const B = 2
const C = 3`,
			checkAST: func(t *testing.T, module *ast.Module) {
				require.Len(t, module.Items, 3)
				for i, item := range module.Items {
					constDecl, ok := item.(*ast.ConstDecl)
					require.True(t, ok, "item %d: expected ConstDecl, got %T", i, item)
					expected := string('A' + rune(i))
					assert.Equal(t, expected, constDecl.Name)
				}
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

// Test const declarations can coexist with other items
func TestParser_ConstWithOtherItems(t *testing.T) {
	input := `const MAX_RETRIES = 3

: User {
	name: string
	age: int
}

const DEFAULT_TIMEOUT = 30`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 3)

	// First item: const
	constDecl1, ok := module.Items[0].(*ast.ConstDecl)
	require.True(t, ok)
	assert.Equal(t, "MAX_RETRIES", constDecl1.Name)

	// Second item: type
	typeDef, ok := module.Items[1].(*ast.TypeDef)
	require.True(t, ok)
	assert.Equal(t, "User", typeDef.Name)

	// Third item: const
	constDecl2, ok := module.Items[2].(*ast.ConstDecl)
	require.True(t, ok)
	assert.Equal(t, "DEFAULT_TIMEOUT", constDecl2.Name)
}
