package tests

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
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
	typeDef, ok := module.Items[0].(*ast.TypeDef)
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

	typeDef := module.Items[0].(*ast.TypeDef)

	// Create type checker and register type
	tc := interpreter.NewTypeChecker()
	tc.SetTypeDefs(map[string]ast.TypeDef{
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

	typeDef := module.Items[0].(*ast.TypeDef)

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

	fn := module.Items[0].(*ast.Function)

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

	typeDef := module.Items[0].(*ast.TypeDef)
	assert.Equal(t, "User", typeDef.Name)

	// Verify default expressions are stored correctly
	roleDefault, ok := typeDef.Fields[0].Default.(ast.LiteralExpr)
	require.True(t, ok)
	strLit, ok := roleDefault.Value.(ast.StringLiteral)
	require.True(t, ok)
	assert.Equal(t, "user", strLit.Value)

	activeDefault, ok := typeDef.Fields[1].Default.(ast.LiteralExpr)
	require.True(t, ok)
	boolLit, ok := activeDefault.Value.(ast.BoolLiteral)
	require.True(t, ok)
	assert.Equal(t, true, boolLit.Value)
}

// TestFieldDefaultsE2E_RouteInputType tests that InputType is parsed and stored
func TestFieldDefaultsE2E_RouteInputType(t *testing.T) {
	source := `
: CreateUserInput {
  name: str!
  role: str = "user"
  active: bool = true
}

@ POST /api/users {
  < input: CreateUserInput
  > {
    id: 123,
    name: input.name,
    role: input.role,
    active: input.active
  }
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	// Should have 2 items: TypeDef and Route
	require.Len(t, module.Items, 2)

	// Check TypeDef
	typeDef, ok := module.Items[0].(*ast.TypeDef)
	require.True(t, ok)
	assert.Equal(t, "CreateUserInput", typeDef.Name)

	// Check Route has InputType
	route, ok := module.Items[1].(*ast.Route)
	require.True(t, ok)
	assert.Equal(t, "/api/users", route.Path)
	require.NotNil(t, route.InputType, "Route should have InputType set")

	// Verify InputType is NamedType pointing to CreateUserInput
	namedType, ok := route.InputType.(ast.NamedType)
	require.True(t, ok, "InputType should be a NamedType")
	assert.Equal(t, "CreateUserInput", namedType.Name)
}

// TestFieldDefaultsE2E_AutomaticDefaultsInRoute tests that defaults are automatically applied
// when executing a route with declared InputType
func TestFieldDefaultsE2E_AutomaticDefaultsInRoute(t *testing.T) {
	source := `
: CreateUserInput {
  name: str!
  role: str = "user"
  active: bool = true
}

@ POST /api/users {
  < input: CreateUserInput
  > {
    name: input.name,
    role: input.role,
    active: input.active
  }
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	// Create interpreter and load module
	interp := interpreter.NewInterpreter()
	err = interp.LoadModule(*module)
	require.NoError(t, err)

	// Get the route from the module
	var route *ast.Route
	for _, item := range module.Items {
		if r, ok := item.(*ast.Route); ok {
			route = r
			break
		}
	}
	require.NotNil(t, route)

	// Create a request with only the required field (name)
	// Defaults should be applied for role and active
	request := &interpreter.Request{
		Path:   "/api/users",
		Method: "POST",
		Body: map[string]interface{}{
			"name": "Alice",
			// "role" and "active" are missing - should get defaults
		},
	}

	// Execute the route
	response, err := interp.ExecuteRoute(route, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)

	// Check response body - defaults should have been applied
	body, ok := response.Body.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Alice", body["name"])
	assert.Equal(t, "user", body["role"], "Default role should be applied")
	assert.Equal(t, true, body["active"], "Default active should be applied")
}

// TestFieldDefaultsE2E_DefaultsNotOverwriteProvided tests that provided values are not overwritten
func TestFieldDefaultsE2E_DefaultsNotOverwriteProvided(t *testing.T) {
	source := `
: CreateUserInput {
  name: str!
  role: str = "user"
  active: bool = true
}

@ POST /api/users {
  < input: CreateUserInput
  > {
    name: input.name,
    role: input.role,
    active: input.active
  }
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	interp := interpreter.NewInterpreter()
	err = interp.LoadModule(*module)
	require.NoError(t, err)

	// Get the route from the module
	var route *ast.Route
	for _, item := range module.Items {
		if r, ok := item.(*ast.Route); ok {
			route = r
			break
		}
	}
	require.NotNil(t, route)

	// Create a request with all fields provided
	request := &interpreter.Request{
		Path:   "/api/users",
		Method: "POST",
		Body: map[string]interface{}{
			"name":   "Bob",
			"role":   "admin", // Explicitly provided, should not be overwritten
			"active": false,   // Explicitly provided, should not be overwritten
		},
	}

	response, err := interp.ExecuteRoute(route, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)

	body, ok := response.Body.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Bob", body["name"])
	assert.Equal(t, "admin", body["role"], "Provided role should not be overwritten")
	assert.Equal(t, false, body["active"], "Provided active should not be overwritten")
}

// TestFieldDefaultsE2E_ValidationFailsWithMissingRequired tests that validation fails
// when required fields without defaults are missing
func TestFieldDefaultsE2E_ValidationFailsWithMissingRequired(t *testing.T) {
	source := `
: CreateUserInput {
  name: str!
  role: str = "user"
}

@ POST /api/users {
  < input: CreateUserInput
  > {
    name: input.name,
    role: input.role
  }
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	interp := interpreter.NewInterpreter()
	err = interp.LoadModule(*module)
	require.NoError(t, err)

	// Get the route from the module
	var route *ast.Route
	for _, item := range module.Items {
		if r, ok := item.(*ast.Route); ok {
			route = r
			break
		}
	}
	require.NotNil(t, route)

	// Create a request missing the required "name" field
	request := &interpreter.Request{
		Path:   "/api/users",
		Method: "POST",
		Body: map[string]interface{}{
			"role": "admin",
			// "name" is missing and has no default
		},
	}

	response, err := interp.ExecuteRoute(route, request)
	// Should return an error due to missing required field
	require.Error(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 400, response.StatusCode, "Should return 400 for validation error")
}

// TestFieldDefaultsE2E_RouteWithoutInputType tests that routes without InputType work normally
func TestFieldDefaultsE2E_RouteWithoutInputType(t *testing.T) {
	source := `
@ POST /api/echo {
  > {
    received: input
  }
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	interp := interpreter.NewInterpreter()
	err = interp.LoadModule(*module)
	require.NoError(t, err)

	// Get the route from the module
	var route *ast.Route
	for _, item := range module.Items {
		if r, ok := item.(*ast.Route); ok {
			route = r
			break
		}
	}
	require.NotNil(t, route)

	// Route should not have InputType
	assert.Nil(t, route.InputType, "Route without input declaration should have nil InputType")

	// Execute should work without any type validation/defaults
	request := &interpreter.Request{
		Path:   "/api/echo",
		Method: "POST",
		Body: map[string]interface{}{
			"anything": "goes",
		},
	}

	response, err := interp.ExecuteRoute(route, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)

	body, ok := response.Body.(map[string]interface{})
	require.True(t, ok)
	received, ok := body["received"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "goes", received["anything"])
}
