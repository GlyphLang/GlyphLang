package parser

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test struct field defaults
func TestParser_StructFieldDefaults(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		fieldName   string
		hasDefault  bool
		expectedReq bool
	}{
		{
			name: "string field with default",
			source: `: User {
  role: str = "user"
}`,
			fieldName:   "role",
			hasDefault:  true,
			expectedReq: false,
		},
		{
			name: "bool field with default",
			source: `: Settings {
  active: bool = true
}`,
			fieldName:   "active",
			hasDefault:  true,
			expectedReq: false,
		},
		{
			name: "int field with default",
			source: `: Config {
  retries: int = 3
}`,
			fieldName:   "retries",
			hasDefault:  true,
			expectedReq: false,
		},
		{
			name: "required field without default",
			source: `: User {
  name: str!
}`,
			fieldName:   "name",
			hasDefault:  false,
			expectedReq: true,
		},
		{
			name: "optional field without default",
			source: `: User {
  nickname: str?
}`,
			fieldName:   "nickname",
			hasDefault:  false,
			expectedReq: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			module, err := parser.Parse()
			require.NoError(t, err)

			typeDef, ok := module.Items[0].(*interpreter.TypeDef)
			require.True(t, ok)
			require.Len(t, typeDef.Fields, 1)

			field := typeDef.Fields[0]
			assert.Equal(t, tt.fieldName, field.Name)
			assert.Equal(t, tt.expectedReq, field.Required)

			if tt.hasDefault {
				assert.NotNil(t, field.Default, "expected field to have default value")
			} else {
				assert.Nil(t, field.Default, "expected field to not have default value")
			}
		})
	}
}

func TestParser_StructWithMixedFields(t *testing.T) {
	source := `: User {
  name: str!
  role: str = "user"
  active: bool = true
  age: int?
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)
	require.Len(t, typeDef.Fields, 4)

	// name: str! - required, no default
	assert.Equal(t, "name", typeDef.Fields[0].Name)
	assert.True(t, typeDef.Fields[0].Required)
	assert.Nil(t, typeDef.Fields[0].Default)

	// role: str = "user" - not marked required, has default
	assert.Equal(t, "role", typeDef.Fields[1].Name)
	assert.False(t, typeDef.Fields[1].Required)
	assert.NotNil(t, typeDef.Fields[1].Default)

	// active: bool = true - not marked required, has default
	assert.Equal(t, "active", typeDef.Fields[2].Name)
	assert.False(t, typeDef.Fields[2].Required)
	assert.NotNil(t, typeDef.Fields[2].Default)

	// age: int? - optional, no default
	assert.Equal(t, "age", typeDef.Fields[3].Name)
	assert.False(t, typeDef.Fields[3].Required)
	assert.Nil(t, typeDef.Fields[3].Default)
}

func TestParser_FunctionParamDefaults(t *testing.T) {
	source := `! greet(name: str!, greeting: str = "Hello"): str {
  > greeting + " " + name
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

	// name: str! - required, no default
	assert.Equal(t, "name", fn.Params[0].Name)
	assert.True(t, fn.Params[0].Required)
	assert.Nil(t, fn.Params[0].Default)

	// greeting: str = "Hello" - has default
	assert.Equal(t, "greeting", fn.Params[1].Name)
	assert.False(t, fn.Params[1].Required)
	assert.NotNil(t, fn.Params[1].Default)
}
