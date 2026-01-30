package parser

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lexTraitSource(source string) []Token {
	lexer := NewLexer(source)
	tokens, _ := lexer.Tokenize()
	return tokens
}

func lexExpandedTraitSource(source string) []Token {
	lexer := NewExpandedLexer(source)
	tokens, _ := lexer.Tokenize()
	return tokens
}

// TestTraitDefinition tests parsing a basic trait definition
func TestTraitDefinition(t *testing.T) {
	source := `trait Serializable {
  toJson() -> string
  fromJson(s: string) -> Self
}`
	tokens := lexTraitSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	trait, ok := module.Items[0].(*interpreter.TraitDef)
	require.True(t, ok, "expected TraitDef")
	assert.Equal(t, "Serializable", trait.Name)
	require.Len(t, trait.Methods, 2)

	assert.Equal(t, "toJson", trait.Methods[0].Name)
	assert.Len(t, trait.Methods[0].Params, 0)
	assert.IsType(t, interpreter.StringType{}, trait.Methods[0].ReturnType)

	assert.Equal(t, "fromJson", trait.Methods[1].Name)
	require.Len(t, trait.Methods[1].Params, 1)
	assert.Equal(t, "s", trait.Methods[1].Params[0].Name)
}

// TestTraitDefinitionExpandedSyntax tests trait parsing with expanded lexer
func TestTraitDefinitionExpandedSyntax(t *testing.T) {
	source := `trait Comparable {
  compareTo(other: Self) -> int
}`
	tokens := lexExpandedTraitSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	trait, ok := module.Items[0].(*interpreter.TraitDef)
	require.True(t, ok)
	assert.Equal(t, "Comparable", trait.Name)
	require.Len(t, trait.Methods, 1)
	assert.Equal(t, "compareTo", trait.Methods[0].Name)
}

// TestTraitWithGenericParams tests trait with generic type parameters
func TestTraitWithGenericParams(t *testing.T) {
	source := `trait Container<T> {
  get(key: string) -> T
  set(key: string, val: T) -> bool
}`
	tokens := lexTraitSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	trait, ok := module.Items[0].(*interpreter.TraitDef)
	require.True(t, ok)
	assert.Equal(t, "Container", trait.Name)
	require.Len(t, trait.TypeParams, 1)
	assert.Equal(t, "T", trait.TypeParams[0].Name)
	require.Len(t, trait.Methods, 2)
}

// TestTypeDefImplTrait tests parsing a TypeDef with impl clause (compact syntax)
func TestTypeDefImplTrait(t *testing.T) {
	source := `: User impl Serializable {
  id: int
  name: string
  toJson() -> string {
    > "json"
  }
}`
	tokens := lexTraitSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	td, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok, "expected TypeDef")
	assert.Equal(t, "User", td.Name)
	require.Len(t, td.Traits, 1)
	assert.Equal(t, "Serializable", td.Traits[0])
	require.Len(t, td.Fields, 2)
	assert.Equal(t, "id", td.Fields[0].Name)
	assert.Equal(t, "name", td.Fields[1].Name)
	require.Len(t, td.Methods, 1)
	assert.Equal(t, "toJson", td.Methods[0].Name)
}

// TestTypeDefImplMultipleTraits tests implementing multiple traits
func TestTypeDefImplMultipleTraits(t *testing.T) {
	source := `: User impl Serializable, Hashable {
  id: int
  toJson() -> string {
    > "json"
  }
  hash() -> int {
    > 42
  }
}`
	tokens := lexTraitSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	td, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)
	assert.Equal(t, "User", td.Name)
	require.Len(t, td.Traits, 2)
	assert.Equal(t, "Serializable", td.Traits[0])
	assert.Equal(t, "Hashable", td.Traits[1])
	require.Len(t, td.Methods, 2)
}

// TestTypeDefWithoutImpl tests that regular TypeDefs still work
func TestTypeDefWithoutImpl(t *testing.T) {
	source := `: User {
  id: int
  name: string
}`
	tokens := lexTraitSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	td, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)
	assert.Equal(t, "User", td.Name)
	assert.Empty(t, td.Traits)
	assert.Empty(t, td.Methods)
	require.Len(t, td.Fields, 2)
}

// TestExpandedTypeDefImpl tests type + impl with expanded syntax
func TestExpandedTypeDefImpl(t *testing.T) {
	source := `type User impl Serializable {
  id: int
  toJson() -> string {
    return "json"
  }
}`
	tokens := lexExpandedTraitSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	td, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)
	assert.Equal(t, "User", td.Name)
	require.Len(t, td.Traits, 1)
	assert.Equal(t, "Serializable", td.Traits[0])
	require.Len(t, td.Methods, 1)
	assert.Equal(t, "toJson", td.Methods[0].Name)
}

// TestTraitAndTypeDefTogether tests parsing a trait definition followed by a type implementing it
func TestTraitAndTypeDefTogether(t *testing.T) {
	source := `trait Printable {
  toString() -> string
}

: Message impl Printable {
  text: string
  toString() -> string {
    > "msg"
  }
}`
	tokens := lexTraitSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 2)

	trait, ok := module.Items[0].(*interpreter.TraitDef)
	require.True(t, ok)
	assert.Equal(t, "Printable", trait.Name)

	td, ok := module.Items[1].(*interpreter.TypeDef)
	require.True(t, ok)
	assert.Equal(t, "Message", td.Name)
	require.Len(t, td.Traits, 1)
	assert.Equal(t, "Printable", td.Traits[0])
}

// TestTraitMethodWithMultipleParams tests parsing trait methods with multiple parameters
func TestTraitMethodWithMultipleParams(t *testing.T) {
	source := `trait Mapper {
  map(key: string, value: int) -> bool
}`
	tokens := lexTraitSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	trait, ok := module.Items[0].(*interpreter.TraitDef)
	require.True(t, ok)
	require.Len(t, trait.Methods, 1)
	assert.Equal(t, "map", trait.Methods[0].Name)
	require.Len(t, trait.Methods[0].Params, 2)
	assert.Equal(t, "key", trait.Methods[0].Params[0].Name)
	assert.Equal(t, "value", trait.Methods[0].Params[1].Name)
}

// TestTraitMethodNoReturnType tests trait method with no return type
func TestTraitMethodNoReturnType(t *testing.T) {
	source := `trait Logger {
  log(msg: string)
}`
	tokens := lexTraitSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	trait, ok := module.Items[0].(*interpreter.TraitDef)
	require.True(t, ok)
	require.Len(t, trait.Methods, 1)
	assert.Equal(t, "log", trait.Methods[0].Name)
	assert.Nil(t, trait.Methods[0].ReturnType)
}
