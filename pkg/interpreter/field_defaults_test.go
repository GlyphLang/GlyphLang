package interpreter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that fields with defaults are not required
func TestTypeChecker_ValidateObjectAgainstTypeDef_FieldWithDefault(t *testing.T) {
	tc := NewTypeChecker()

	// Create a TypeDef with a field that has a default value
	typeDef := TypeDef{
		Name: "User",
		Fields: []Field{
			{Name: "name", TypeAnnotation: StringType{}, Required: true, Default: nil},
			{Name: "role", TypeAnnotation: StringType{}, Required: false, Default: LiteralExpr{Value: StringLiteral{Value: "user"}}},
		},
	}

	// Object with only the required field (missing field with default)
	obj := map[string]interface{}{
		"name": "Alice",
		// "role" is missing but has a default
	}

	err := tc.ValidateObjectAgainstTypeDef(obj, typeDef)
	assert.NoError(t, err, "field with default should not be required")
}

func TestTypeChecker_ValidateObjectAgainstTypeDef_RequiredFieldNoDefault(t *testing.T) {
	tc := NewTypeChecker()

	// Required field without default - should error if missing
	typeDef := TypeDef{
		Name: "User",
		Fields: []Field{
			{Name: "name", TypeAnnotation: StringType{}, Required: true, Default: nil},
		},
	}

	obj := map[string]interface{}{
		// "name" is missing and required without default
	}

	err := tc.ValidateObjectAgainstTypeDef(obj, typeDef)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field: name")
}

// Test ApplyTypeDefaults function
func TestInterpreter_ApplyTypeDefaults(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	typeDef := TypeDef{
		Name: "User",
		Fields: []Field{
			{Name: "name", TypeAnnotation: StringType{}, Required: true, Default: nil},
			{Name: "role", TypeAnnotation: StringType{}, Required: false, Default: LiteralExpr{Value: StringLiteral{Value: "user"}}},
			{Name: "active", TypeAnnotation: BoolType{}, Required: false, Default: LiteralExpr{Value: BoolLiteral{Value: true}}},
		},
	}

	// Object with only name provided
	obj := map[string]interface{}{
		"name": "Alice",
	}

	result, err := interp.ApplyTypeDefaults(obj, typeDef, env)
	require.NoError(t, err)

	// Original name should be preserved
	assert.Equal(t, "Alice", result["name"])
	// Defaults should be applied
	assert.Equal(t, "user", result["role"])
	assert.Equal(t, true, result["active"])
}

func TestInterpreter_ApplyTypeDefaults_ExistingValuesNotOverwritten(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	typeDef := TypeDef{
		Name: "User",
		Fields: []Field{
			{Name: "role", TypeAnnotation: StringType{}, Required: false, Default: LiteralExpr{Value: StringLiteral{Value: "user"}}},
		},
	}

	// Object with role already provided
	obj := map[string]interface{}{
		"role": "admin",
	}

	result, err := interp.ApplyTypeDefaults(obj, typeDef, env)
	require.NoError(t, err)

	// Existing value should not be overwritten
	assert.Equal(t, "admin", result["role"])
}

func TestInterpreter_ApplyTypeDefaults_NoDefaultsNeeded(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	typeDef := TypeDef{
		Name: "User",
		Fields: []Field{
			{Name: "name", TypeAnnotation: StringType{}, Required: true, Default: nil},
		},
	}

	// Object with all fields provided
	obj := map[string]interface{}{
		"name": "Alice",
	}

	result, err := interp.ApplyTypeDefaults(obj, typeDef, env)
	require.NoError(t, err)

	assert.Equal(t, "Alice", result["name"])
	assert.Len(t, result, 1)
}

func TestInterpreter_ApplyTypeDefaults_IntDefault(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	typeDef := TypeDef{
		Name: "Config",
		Fields: []Field{
			{Name: "retries", TypeAnnotation: IntType{}, Required: false, Default: LiteralExpr{Value: IntLiteral{Value: 3}}},
		},
	}

	obj := map[string]interface{}{}

	result, err := interp.ApplyTypeDefaults(obj, typeDef, env)
	require.NoError(t, err)

	assert.Equal(t, int64(3), result["retries"])
}

// Test executeFunction with optional parameter without default
func TestInterpreter_ExecuteFunction_OptionalParamWithoutDefault(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Function with required param and optional param (no default)
	// The function just returns a literal - we're testing that it can be called
	// with only the required argument (the optional one gets nil)
	fn := Function{
		Name: "test",
		Params: []Field{
			{Name: "name", TypeAnnotation: StringType{}, Required: true, Default: nil},
			{Name: "nickname", TypeAnnotation: StringType{}, Required: false, Default: nil},
		},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "success"}}},
		},
	}

	// Call with only the required argument - optional param should get nil
	args := []Expr{
		LiteralExpr{Value: StringLiteral{Value: "Alice"}},
	}

	result, err := interp.executeFunction(fn, args, env)
	require.NoError(t, err, "should not error when optional param without default is omitted")
	assert.Equal(t, "success", result)
}

// Test executeFunction counts only required params without defaults
func TestInterpreter_ExecuteFunction_RequiredParamCount(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Function with: required param, optional param (no default), param with default
	fn := Function{
		Name: "test",
		Params: []Field{
			{Name: "a", TypeAnnotation: StringType{}, Required: true, Default: nil},
			{Name: "b", TypeAnnotation: StringType{}, Required: false, Default: nil},
			{Name: "c", TypeAnnotation: StringType{}, Required: false, Default: LiteralExpr{Value: StringLiteral{Value: "default"}}},
		},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "ok"}}},
		},
	}

	// Should succeed with just required param
	args := []Expr{
		LiteralExpr{Value: StringLiteral{Value: "value"}},
	}

	_, err := interp.executeFunction(fn, args, env)
	require.NoError(t, err, "should succeed with only required argument")

	// Should fail with no arguments (missing required)
	_, err = interp.executeFunction(fn, []Expr{}, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expects at least 1 argument")
}

// Test that providing all arguments still works correctly
func TestInterpreter_ExecuteFunction_AllArgsProvided(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	fn := Function{
		Name: "test",
		Params: []Field{
			{Name: "a", TypeAnnotation: StringType{}, Required: true, Default: nil},
			{Name: "b", TypeAnnotation: StringType{}, Required: false, Default: nil},
			{Name: "c", TypeAnnotation: StringType{}, Required: false, Default: LiteralExpr{Value: StringLiteral{Value: "default_c"}}},
		},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "ok"}}},
		},
	}

	// Provide all three arguments
	args := []Expr{
		LiteralExpr{Value: StringLiteral{Value: "val_a"}},
		LiteralExpr{Value: StringLiteral{Value: "val_b"}},
		LiteralExpr{Value: StringLiteral{Value: "val_c"}},
	}

	_, err := interp.executeFunction(fn, args, env)
	require.NoError(t, err, "should succeed when all arguments are provided")
}

// Test that too many arguments fails
func TestInterpreter_ExecuteFunction_TooManyArgs(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	fn := Function{
		Name: "test",
		Params: []Field{
			{Name: "a", TypeAnnotation: StringType{}, Required: true, Default: nil},
		},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "ok"}}},
		},
	}

	// Provide too many arguments
	args := []Expr{
		LiteralExpr{Value: StringLiteral{Value: "val_a"}},
		LiteralExpr{Value: StringLiteral{Value: "extra"}},
	}

	_, err := interp.executeFunction(fn, args, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expects at most 1 argument")
}

// Test required param with default is not required
func TestInterpreter_ExecuteFunction_RequiredParamWithDefault(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// A required (!) param with a default should NOT require an argument
	fn := Function{
		Name: "test",
		Params: []Field{
			{Name: "name", TypeAnnotation: StringType{}, Required: true, Default: LiteralExpr{Value: StringLiteral{Value: "default_name"}}},
		},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "ok"}}},
		},
	}

	// Call with no arguments - should work because default exists
	_, err := interp.executeFunction(fn, []Expr{}, env)
	require.NoError(t, err, "required param with default should not require argument")
}

// Test InstantiateGenericType preserves defaults
func TestTypeChecker_InstantiateGenericType_PreservesDefaults(t *testing.T) {
	tc := NewTypeChecker()

	defaultExpr := LiteralExpr{Value: IntLiteral{Value: 0}}
	typeDef := TypeDef{
		Name: "Container",
		TypeParams: []TypeParameter{
			{Name: "T"},
		},
		Fields: []Field{
			{Name: "value", TypeAnnotation: TypeParameterType{Name: "T"}, Required: true, Default: nil},
			{Name: "count", TypeAnnotation: IntType{}, Required: false, Default: defaultExpr},
		},
	}

	instantiated, err := tc.InstantiateGenericType(typeDef, []Type{StringType{}})
	require.NoError(t, err)

	// The count field should still have its default
	require.Len(t, instantiated.Fields, 2)
	assert.NotNil(t, instantiated.Fields[1].Default, "default should be preserved in generic instantiation")
}

// Test ExecuteRoute with InputType and defaults
func TestInterpreter_ExecuteRoute_WithInputTypeDefaults(t *testing.T) {
	interp := NewInterpreter()

	// Register a TypeDef with defaults
	typeDef := TypeDef{
		Name: "CreateUserInput",
		Fields: []Field{
			{Name: "name", TypeAnnotation: StringType{}, Required: true, Default: nil},
			{Name: "role", TypeAnnotation: StringType{}, Required: false, Default: LiteralExpr{Value: StringLiteral{Value: "user"}}},
			{Name: "active", TypeAnnotation: BoolType{}, Required: false, Default: LiteralExpr{Value: BoolLiteral{Value: true}}},
		},
	}
	interp.typeDefs["CreateUserInput"] = typeDef
	interp.typeChecker.SetTypeDefs(interp.typeDefs)

	// Create a route with InputType
	route := &Route{
		Path:      "/api/users",
		Method:    Post,
		InputType: NamedType{Name: "CreateUserInput"},
		Body: []Statement{
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "name", Value: FieldAccessExpr{Object: VariableExpr{Name: "input"}, Field: "name"}},
						{Key: "role", Value: FieldAccessExpr{Object: VariableExpr{Name: "input"}, Field: "role"}},
						{Key: "active", Value: FieldAccessExpr{Object: VariableExpr{Name: "input"}, Field: "active"}},
					},
				},
			},
		},
	}

	// Create a request with only required fields - defaults should be applied
	request := &Request{
		Path:   "/api/users",
		Method: "POST",
		Body: map[string]interface{}{
			"name": "Alice",
			// "role" and "active" are missing - defaults should be applied
		},
	}

	response, err := interp.ExecuteRoute(route, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)

	body, ok := response.Body.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Alice", body["name"])
	assert.Equal(t, "user", body["role"], "Default role should be applied")
	assert.Equal(t, true, body["active"], "Default active should be applied")
}

// Test ExecuteRoute validation fails for missing required field
func TestInterpreter_ExecuteRoute_ValidationFailsMissingRequired(t *testing.T) {
	interp := NewInterpreter()

	// Register a TypeDef with a required field
	typeDef := TypeDef{
		Name: "CreateUserInput",
		Fields: []Field{
			{Name: "name", TypeAnnotation: StringType{}, Required: true, Default: nil},
			{Name: "role", TypeAnnotation: StringType{}, Required: false, Default: LiteralExpr{Value: StringLiteral{Value: "user"}}},
		},
	}
	interp.typeDefs["CreateUserInput"] = typeDef
	interp.typeChecker.SetTypeDefs(interp.typeDefs)

	route := &Route{
		Path:      "/api/users",
		Method:    Post,
		InputType: NamedType{Name: "CreateUserInput"},
		Body: []Statement{
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "name", Value: FieldAccessExpr{Object: VariableExpr{Name: "input"}, Field: "name"}},
					},
				},
			},
		},
	}

	// Create a request missing the required "name" field
	request := &Request{
		Path:   "/api/users",
		Method: "POST",
		Body: map[string]interface{}{
			"role": "admin",
			// "name" is missing and required
		},
	}

	response, err := interp.ExecuteRoute(route, request)
	require.Error(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 400, response.StatusCode, "Should return 400 for validation error")
}

// Test ExecuteRoute without InputType works normally
func TestInterpreter_ExecuteRoute_NoInputType(t *testing.T) {
	interp := NewInterpreter()

	// Create a route without InputType
	route := &Route{
		Path:   "/api/echo",
		Method: Post,
		// InputType is nil
		Body: []Statement{
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "received", Value: VariableExpr{Name: "input"}},
					},
				},
			},
		},
	}

	// Any input should work
	request := &Request{
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

// Test that existing values are not overwritten by defaults
func TestInterpreter_ExecuteRoute_ExistingValuesNotOverwritten(t *testing.T) {
	interp := NewInterpreter()

	typeDef := TypeDef{
		Name: "CreateUserInput",
		Fields: []Field{
			{Name: "name", TypeAnnotation: StringType{}, Required: true, Default: nil},
			{Name: "role", TypeAnnotation: StringType{}, Required: false, Default: LiteralExpr{Value: StringLiteral{Value: "user"}}},
		},
	}
	interp.typeDefs["CreateUserInput"] = typeDef
	interp.typeChecker.SetTypeDefs(interp.typeDefs)

	route := &Route{
		Path:      "/api/users",
		Method:    Post,
		InputType: NamedType{Name: "CreateUserInput"},
		Body: []Statement{
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "name", Value: FieldAccessExpr{Object: VariableExpr{Name: "input"}, Field: "name"}},
						{Key: "role", Value: FieldAccessExpr{Object: VariableExpr{Name: "input"}, Field: "role"}},
					},
				},
			},
		},
	}

	// Provide all values including role
	request := &Request{
		Path:   "/api/users",
		Method: "POST",
		Body: map[string]interface{}{
			"name": "Bob",
			"role": "admin", // This should not be overwritten
		},
	}

	response, err := interp.ExecuteRoute(route, request)
	require.NoError(t, err)
	require.NotNil(t, response)

	body, ok := response.Body.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Bob", body["name"])
	assert.Equal(t, "admin", body["role"], "Provided role should not be overwritten")
}
