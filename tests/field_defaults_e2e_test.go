package tests

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFieldDefaultsE2E tests the full flow of struct field defaults
func TestFieldDefaultsE2E_StructWithDefaults(t *testing.T) {
	source := `
: User {
  role: str = "user"
  active: bool = true
  name: str!
}

! createUser(name: str!): User {
  > {name: name, role: "admin"}
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	// Should have 2 items: TypeDef and Function
	require.Len(t, module.Items, 2)

	// Check TypeDef
	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)
	assert.Equal(t, "User", typeDef.Name)
	require.Len(t, typeDef.Fields, 3)

	// role: str = "user"
	assert.Equal(t, "role", typeDef.Fields[0].Name)
	assert.NotNil(t, typeDef.Fields[0].Default)
	assert.False(t, typeDef.Fields[0].Required)

	// active: bool = true
	assert.Equal(t, "active", typeDef.Fields[1].Name)
	assert.NotNil(t, typeDef.Fields[1].Default)
	assert.False(t, typeDef.Fields[1].Required)

	// name: str!
	assert.Equal(t, "name", typeDef.Fields[2].Name)
	assert.Nil(t, typeDef.Fields[2].Default)
	assert.True(t, typeDef.Fields[2].Required)
}

func TestFieldDefaultsE2E_ValidationWithDefaults(t *testing.T) {
	source := `
: Config {
  debug: bool = false
  timeout: int = 30
  host: str!
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	typeDef := module.Items[0].(*interpreter.TypeDef)

	// Create type checker and register type
	tc := interpreter.NewTypeChecker()
	tc.SetTypeDefs(map[string]interpreter.TypeDef{
		"Config": *typeDef,
	})

	// Object with only required field should pass validation
	// (since fields with defaults are not required)
	obj := map[string]interface{}{
		"host": "localhost",
	}

	err = tc.ValidateObjectAgainstTypeDef(obj, *typeDef)
	assert.NoError(t, err, "object with only required fields should be valid")
}

func TestFieldDefaultsE2E_ApplyDefaults(t *testing.T) {
	source := `
: User {
  role: str = "user"
  active: bool = true
  name: str!
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	typeDef := module.Items[0].(*interpreter.TypeDef)

	// Create interpreter and environment
	interp := interpreter.NewInterpreter()
	env := interpreter.NewEnvironment()

	// Object with only required field
	obj := map[string]interface{}{
		"name": "Alice",
	}

	// Apply defaults
	result, err := interp.ApplyTypeDefaults(obj, *typeDef, env)
	require.NoError(t, err)

	// Check that defaults were applied
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, "user", result["role"])
	assert.Equal(t, true, result["active"])
}

func TestFieldDefaultsE2E_FunctionWithDefaultParams(t *testing.T) {
	source := `
! greet(name: str!, greeting: str = "Hello", punctuation: str = "!"): str {
  > greeting + ", " + name + punctuation
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	fn := module.Items[0].(*interpreter.Function)

	// Check function parameters
	require.Len(t, fn.Params, 3)

	// name: str! - required, no default
	assert.Equal(t, "name", fn.Params[0].Name)
	assert.True(t, fn.Params[0].Required)
	assert.Nil(t, fn.Params[0].Default)

	// greeting: str = "Hello" - has default
	assert.Equal(t, "greeting", fn.Params[1].Name)
	assert.False(t, fn.Params[1].Required)
	assert.NotNil(t, fn.Params[1].Default)

	// punctuation: str = "!" - has default
	assert.Equal(t, "punctuation", fn.Params[2].Name)
	assert.False(t, fn.Params[2].Required)
	assert.NotNil(t, fn.Params[2].Default)
}

func TestFieldDefaultsE2E_IssueExample(t *testing.T) {
	// Example from issue #63
	source := `
: User {
  role: str = "user"
  active: bool = true
  name: str!
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	typeDef := module.Items[0].(*interpreter.TypeDef)
	assert.Equal(t, "User", typeDef.Name)

	// Verify default expressions are stored correctly
	roleDefault, ok := typeDef.Fields[0].Default.(interpreter.LiteralExpr)
	require.True(t, ok)
	strLit, ok := roleDefault.Value.(interpreter.StringLiteral)
	require.True(t, ok)
	assert.Equal(t, "user", strLit.Value)

	activeDefault, ok := typeDef.Fields[1].Default.(interpreter.LiteralExpr)
	require.True(t, ok)
	boolLit, ok := activeDefault.Value.(interpreter.BoolLiteral)
	require.True(t, ok)
	assert.Equal(t, true, boolLit.Value)
}
