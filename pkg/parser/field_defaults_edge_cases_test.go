package parser

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that required parameters must come before optional ones
func TestParser_FunctionParamOrdering_InvalidOrder(t *testing.T) {
	// Required parameter after optional parameter should fail
	source := `! greet(greeting: str = "Hello", name: str!): str {
  > greeting + " " + name
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	_, err = parser.Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required parameter 'name' cannot come after optional parameters")
}

// Test valid parameter ordering with multiple optional params
func TestParser_FunctionParamOrdering_ValidOrder(t *testing.T) {
	source := `! greet(name: str!, greeting: str = "Hello", punctuation: str = "!"): str {
  > greeting + " " + name + punctuation
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	fn, ok := module.Items[0].(*interpreter.Function)
	require.True(t, ok)
	require.Len(t, fn.Params, 3)

	// All required params before optional ones - should pass
	assert.True(t, fn.Params[0].Required)
	assert.Nil(t, fn.Params[0].Default)
	assert.NotNil(t, fn.Params[1].Default)
	assert.NotNil(t, fn.Params[2].Default)
}

// Test default with arithmetic expression
func TestParser_FieldDefaultExpression(t *testing.T) {
	source := `: Config {
  timeout: int = 30 * 60
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)
	require.Len(t, typeDef.Fields, 1)

	// Should parse as a binary expression
	_, ok = typeDef.Fields[0].Default.(interpreter.BinaryOpExpr)
	assert.True(t, ok, "default should be a binary expression")
}

// Test default with array literal
func TestParser_FieldDefaultArrayLiteral(t *testing.T) {
	source := `: Config {
  tags: [str] = []
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)
	require.Len(t, typeDef.Fields, 1)

	// Should parse as an array expression
	_, ok = typeDef.Fields[0].Default.(interpreter.ArrayExpr)
	assert.True(t, ok, "default should be an array expression")
}

// Test float default
func TestParser_FieldDefaultFloat(t *testing.T) {
	source := `: Config {
  rate: float = 0.5
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)

	lit, ok := typeDef.Fields[0].Default.(interpreter.LiteralExpr)
	require.True(t, ok)
	floatLit, ok := lit.Value.(interpreter.FloatLiteral)
	require.True(t, ok)
	assert.Equal(t, 0.5, floatLit.Value)
}

// Test generic function with param defaults
func TestParser_GenericFunctionParamDefaults(t *testing.T) {
	source := `! process<T>(items: [T]!, limit: int = 10): [T] {
  > items
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	fn, ok := module.Items[0].(*interpreter.Function)
	require.True(t, ok)
	require.Len(t, fn.Params, 2)

	// items: [T]! - required (marked with !)
	assert.True(t, fn.Params[0].Required)
	assert.Nil(t, fn.Params[0].Default)

	// limit: int = 10 - has default
	assert.False(t, fn.Params[1].Required)
	assert.NotNil(t, fn.Params[1].Default)
}

// Test generic function with invalid param ordering
func TestParser_GenericFunctionParamOrdering_Invalid(t *testing.T) {
	// Required parameter (items: [T]!) comes after optional parameter (limit: int = 10)
	source := `! process<T>(limit: int = 10, items: [T]!): [T] {
  > items
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	_, err = parser.Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required parameter 'items' cannot come after optional parameters")
}

// Test that optional params (without !) don't need defaults
func TestParser_OptionalParamWithoutDefault(t *testing.T) {
	source := `! test(name: str!, nickname: str): str {
  > name
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	fn, ok := module.Items[0].(*interpreter.Function)
	require.True(t, ok)
	require.Len(t, fn.Params, 2)

	// name: str! - required
	assert.True(t, fn.Params[0].Required)

	// nickname: str - optional (no !) but no default either
	assert.False(t, fn.Params[1].Required)
	assert.Nil(t, fn.Params[1].Default)
}
