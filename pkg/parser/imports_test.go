package parser

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test lexer tokens for imports

func TestLexer_ImportKeywords(t *testing.T) {
	input := `import from as module`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()

	require.NoError(t, err)

	expected := []TokenType{IMPORT, FROM, AS, MODULE, EOF}
	require.Len(t, tokens, len(expected))

	for i, expectedType := range expected {
		assert.Equal(t, expectedType, tokens[i].Type, "token %d", i)
	}
}

func TestLexer_ImportStatement(t *testing.T) {
	input := `import "./utils"`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()

	require.NoError(t, err)
	require.Len(t, tokens, 3) // IMPORT, STRING, EOF

	assert.Equal(t, IMPORT, tokens[0].Type)
	assert.Equal(t, STRING, tokens[1].Type)
	assert.Equal(t, "./utils", tokens[1].Literal)
}

func TestLexer_ImportWithAlias(t *testing.T) {
	input := `import "./utils" as u`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()

	require.NoError(t, err)
	require.Len(t, tokens, 5) // IMPORT, STRING, AS, IDENT, EOF

	assert.Equal(t, IMPORT, tokens[0].Type)
	assert.Equal(t, STRING, tokens[1].Type)
	assert.Equal(t, AS, tokens[2].Type)
	assert.Equal(t, IDENT, tokens[3].Type)
	assert.Equal(t, "u", tokens[3].Literal)
}

func TestLexer_FromImport(t *testing.T) {
	input := `from "./utils" import { funcA, funcB }`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()

	require.NoError(t, err)

	expected := []TokenType{FROM, STRING, IMPORT, LBRACE, IDENT, COMMA, IDENT, RBRACE, EOF}
	require.Len(t, tokens, len(expected))

	for i, expectedType := range expected {
		assert.Equal(t, expectedType, tokens[i].Type, "token %d", i)
	}
}

func TestLexer_ModuleDecl(t *testing.T) {
	input := `module "myapp/utils"`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()

	require.NoError(t, err)
	require.Len(t, tokens, 3) // MODULE, STRING, EOF

	assert.Equal(t, MODULE, tokens[0].Type)
	assert.Equal(t, STRING, tokens[1].Type)
	assert.Equal(t, "myapp/utils", tokens[1].Literal)
}

// Test parser for imports

func TestParser_ImportStatement(t *testing.T) {
	input := `import "./utils"`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()

	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	importStmt, ok := module.Items[0].(*interpreter.ImportStatement)
	require.True(t, ok, "expected ImportStatement, got %T", module.Items[0])

	assert.Equal(t, "./utils", importStmt.Path)
	assert.Empty(t, importStmt.Alias)
	assert.False(t, importStmt.Selective)
}

func TestParser_ImportWithAlias(t *testing.T) {
	input := `import "./models" as m`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()

	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	importStmt, ok := module.Items[0].(*interpreter.ImportStatement)
	require.True(t, ok)

	assert.Equal(t, "./models", importStmt.Path)
	assert.Equal(t, "m", importStmt.Alias)
	assert.False(t, importStmt.Selective)
}

func TestParser_FromImport(t *testing.T) {
	input := `from "./utils" import { getAllUsers }`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()

	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	importStmt, ok := module.Items[0].(*interpreter.ImportStatement)
	require.True(t, ok)

	assert.Equal(t, "./utils", importStmt.Path)
	assert.True(t, importStmt.Selective)
	require.Len(t, importStmt.Names, 1)
	assert.Equal(t, "getAllUsers", importStmt.Names[0].Name)
	assert.Empty(t, importStmt.Names[0].Alias)
}

func TestParser_FromImportMultiple(t *testing.T) {
	input := `from "./utils" import { funcA, funcB, funcC }`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()

	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	importStmt, ok := module.Items[0].(*interpreter.ImportStatement)
	require.True(t, ok)

	assert.True(t, importStmt.Selective)
	require.Len(t, importStmt.Names, 3)
	assert.Equal(t, "funcA", importStmt.Names[0].Name)
	assert.Equal(t, "funcB", importStmt.Names[1].Name)
	assert.Equal(t, "funcC", importStmt.Names[2].Name)
}

func TestParser_FromImportWithAliases(t *testing.T) {
	input := `from "./models" import { User as U, Order as Ord }`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()

	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	importStmt, ok := module.Items[0].(*interpreter.ImportStatement)
	require.True(t, ok)

	assert.True(t, importStmt.Selective)
	require.Len(t, importStmt.Names, 2)

	assert.Equal(t, "User", importStmt.Names[0].Name)
	assert.Equal(t, "U", importStmt.Names[0].Alias)

	assert.Equal(t, "Order", importStmt.Names[1].Name)
	assert.Equal(t, "Ord", importStmt.Names[1].Alias)
}

func TestParser_ModuleDecl(t *testing.T) {
	input := `module "myapp/utils"`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()

	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	moduleDecl, ok := module.Items[0].(*interpreter.ModuleDecl)
	require.True(t, ok, "expected ModuleDecl, got %T", module.Items[0])

	assert.Equal(t, "myapp/utils", moduleDecl.Name)
}

func TestParser_MultipleImports(t *testing.T) {
	input := `
import "./utils"
import "./models" as m
from "./helpers" import { formatDate, formatCurrency }
`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()

	require.NoError(t, err)
	require.Len(t, module.Items, 3)

	// First import
	import1, ok := module.Items[0].(*interpreter.ImportStatement)
	require.True(t, ok)
	assert.Equal(t, "./utils", import1.Path)
	assert.False(t, import1.Selective)

	// Second import
	import2, ok := module.Items[1].(*interpreter.ImportStatement)
	require.True(t, ok)
	assert.Equal(t, "./models", import2.Path)
	assert.Equal(t, "m", import2.Alias)

	// Third import
	import3, ok := module.Items[2].(*interpreter.ImportStatement)
	require.True(t, ok)
	assert.True(t, import3.Selective)
	require.Len(t, import3.Names, 2)
}

func TestParser_ImportsWithOtherItems(t *testing.T) {
	input := `
import "./utils"

: User {
  name: str!
}

@ GET /users {
  > {status: "ok"}
}
`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()

	require.NoError(t, err)
	require.Len(t, module.Items, 3)

	// Import
	_, ok := module.Items[0].(*interpreter.ImportStatement)
	require.True(t, ok)

	// Type definition
	_, ok = module.Items[1].(*interpreter.TypeDef)
	require.True(t, ok)

	// Route
	_, ok = module.Items[2].(*interpreter.Route)
	require.True(t, ok)
}

// Test error cases

func TestParser_ImportMissingPath(t *testing.T) {
	input := `import`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	_, err = parser.Parse()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected import path string")
}

func TestParser_ImportMissingAliasName(t *testing.T) {
	input := `import "./utils" as`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	_, err = parser.Parse()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected alias name")
}

func TestParser_FromImportMissingImport(t *testing.T) {
	input := `from "./utils" { funcA }`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	_, err = parser.Parse()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected 'import' after path")
}

func TestParser_FromImportMissingBrace(t *testing.T) {
	input := `from "./utils" import funcA`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	_, err = parser.Parse()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected '{'")
}

func TestParser_ModuleDeclMissingName(t *testing.T) {
	input := `module`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	_, err = parser.Parse()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected module name string")
}

// Test token string representations

func TestTokenType_ImportKeywords(t *testing.T) {
	assert.Equal(t, "IMPORT", IMPORT.String())
	assert.Equal(t, "FROM", FROM.String())
	assert.Equal(t, "AS", AS.String())
	assert.Equal(t, "MODULE", MODULE.String())
}
