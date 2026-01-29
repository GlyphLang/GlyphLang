package graphql

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseSimpleQuery verifies parsing a basic query with a single field
func TestParseSimpleQuery(t *testing.T) {
	q, err := ParseQuery(`{ users }`)
	require.NoError(t, err)
	assert.Equal(t, "query", q.OperationType)
	assert.Len(t, q.Selections, 1)
	assert.Equal(t, "users", q.Selections[0].Name)
}

// TestParseExplicitQuery verifies parsing with explicit "query" keyword
func TestParseExplicitQuery(t *testing.T) {
	q, err := ParseQuery(`query { users }`)
	require.NoError(t, err)
	assert.Equal(t, "query", q.OperationType)
	assert.Len(t, q.Selections, 1)
}

// TestParseNamedQuery verifies parsing a named query operation
func TestParseNamedQuery(t *testing.T) {
	q, err := ParseQuery(`query GetUsers { users }`)
	require.NoError(t, err)
	assert.Equal(t, "query", q.OperationType)
	assert.Equal(t, "GetUsers", q.Name)
	assert.Len(t, q.Selections, 1)
}

// TestParseMutation verifies parsing a mutation operation
func TestParseMutation(t *testing.T) {
	q, err := ParseQuery(`mutation { createUser }`)
	require.NoError(t, err)
	assert.Equal(t, "mutation", q.OperationType)
	assert.Len(t, q.Selections, 1)
	assert.Equal(t, "createUser", q.Selections[0].Name)
}

// TestParseNestedSelections verifies parsing nested field selections
func TestParseNestedSelections(t *testing.T) {
	q, err := ParseQuery(`{
		user {
			name
			email
			posts {
				title
			}
		}
	}`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 1)

	user := q.Selections[0]
	assert.Equal(t, "user", user.Name)
	require.Len(t, user.Selections, 3)
	assert.Equal(t, "name", user.Selections[0].Name)
	assert.Equal(t, "email", user.Selections[1].Name)

	posts := user.Selections[2]
	assert.Equal(t, "posts", posts.Name)
	require.Len(t, posts.Selections, 1)
	assert.Equal(t, "title", posts.Selections[0].Name)
}

// TestParseArguments verifies parsing field arguments
func TestParseArguments(t *testing.T) {
	q, err := ParseQuery(`{ user(id: 42) { name } }`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 1)

	user := q.Selections[0]
	assert.Equal(t, "user", user.Name)
	assert.Equal(t, int64(42), user.Args["id"])
	require.Len(t, user.Selections, 1)
}

// TestParseStringArgument verifies parsing string arguments
func TestParseStringArgument(t *testing.T) {
	q, err := ParseQuery(`{ user(name: "Alice") }`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 1)
	assert.Equal(t, "Alice", q.Selections[0].Args["name"])
}

// TestParseBooleanArgument verifies parsing boolean arguments
func TestParseBooleanArgument(t *testing.T) {
	q, err := ParseQuery(`{ users(active: true) }`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 1)
	assert.Equal(t, true, q.Selections[0].Args["active"])
}

// TestParseNullArgument verifies parsing null arguments
func TestParseNullArgument(t *testing.T) {
	q, err := ParseQuery(`{ user(cursor: null) }`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 1)
	assert.Nil(t, q.Selections[0].Args["cursor"])
}

// TestParseFloatArgument verifies parsing float arguments
func TestParseFloatArgument(t *testing.T) {
	q, err := ParseQuery(`{ products(minPrice: 9.99) }`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 1)
	assert.Equal(t, 9.99, q.Selections[0].Args["minPrice"])
}

// TestParseMultipleArguments verifies parsing multiple field arguments
func TestParseMultipleArguments(t *testing.T) {
	q, err := ParseQuery(`{ users(limit: 10, offset: 20) }`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 1)
	assert.Equal(t, int64(10), q.Selections[0].Args["limit"])
	assert.Equal(t, int64(20), q.Selections[0].Args["offset"])
}

// TestParseAlias verifies parsing field aliases
func TestParseAlias(t *testing.T) {
	q, err := ParseQuery(`{ admin: user(id: 1) { name } }`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 1)

	sel := q.Selections[0]
	assert.Equal(t, "user", sel.Name)
	assert.Equal(t, "admin", sel.Alias)
	assert.Equal(t, "admin", sel.EffectiveName())
}

// TestParseMultipleTopLevelFields verifies parsing multiple top-level fields
func TestParseMultipleTopLevelFields(t *testing.T) {
	q, err := ParseQuery(`{
		user(id: 1) { name }
		posts { title }
	}`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 2)
	assert.Equal(t, "user", q.Selections[0].Name)
	assert.Equal(t, "posts", q.Selections[1].Name)
}

// TestParseComments verifies that comments are skipped
func TestParseComments(t *testing.T) {
	q, err := ParseQuery(`{
		# This is a comment
		user { name }
	}`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 1)
	assert.Equal(t, "user", q.Selections[0].Name)
}

// TestParseError verifies error on invalid query
func TestParseError(t *testing.T) {
	_, err := ParseQuery(`invalid`)
	assert.Error(t, err)
}

// TestParseEmptyBraces verifies an empty selection set
func TestParseEmptyBraces(t *testing.T) {
	q, err := ParseQuery(`{ }`)
	require.NoError(t, err)
	assert.Empty(t, q.Selections)
}

// TestBuildSchema verifies schema construction from type defs and resolvers
func TestBuildSchema(t *testing.T) {
	typeDefs := map[string]interpreter.TypeDef{
		"User": {
			Name: "User",
			Fields: []interpreter.Field{
				{Name: "id", TypeAnnotation: interpreter.IntType{}, Required: true},
				{Name: "name", TypeAnnotation: interpreter.StringType{}, Required: true},
				{Name: "email", TypeAnnotation: interpreter.StringType{}},
			},
		},
	}

	resolvers := map[string]interpreter.GraphQLResolver{
		"query.user": {
			Operation:  interpreter.GraphQLQuery,
			FieldName:  "user",
			Params:     []interpreter.Field{{Name: "id", TypeAnnotation: interpreter.IntType{}, Required: true}},
			ReturnType: interpreter.NamedType{Name: "User"},
		},
		"query.users": {
			Operation:  interpreter.GraphQLQuery,
			FieldName:  "users",
			Params:     []interpreter.Field{{Name: "limit", TypeAnnotation: interpreter.IntType{}}},
			ReturnType: interpreter.ArrayType{ElementType: interpreter.NamedType{Name: "User"}},
		},
		"mutation.createUser": {
			Operation:  interpreter.GraphQLMutation,
			FieldName:  "createUser",
			Params:     []interpreter.Field{{Name: "name", TypeAnnotation: interpreter.StringType{}, Required: true}},
			ReturnType: interpreter.NamedType{Name: "User"},
		},
	}

	schema := BuildSchema(typeDefs, resolvers)
	require.NotNil(t, schema)
	require.NotNil(t, schema.Query, "schema.Query should not be nil")
	require.NotNil(t, schema.Mutation, "schema.Mutation should not be nil")
	assert.Len(t, schema.Query.Fields, 2)
	assert.Len(t, schema.Mutation.Fields, 1)

	// Verify User type was created
	userType, ok := schema.Types["User"]
	require.True(t, ok, "User type should exist in schema")
	assert.Len(t, userType.Fields, 3)
	assert.Equal(t, "Int", userType.Fields["id"].Type)
	assert.Equal(t, "String", userType.Fields["name"].Type)

	// Verify query fields
	userField, ok := schema.Query.Fields["user"]
	require.True(t, ok, "user field should exist on Query type")
	assert.Equal(t, "User", userField.Type)
	require.Len(t, userField.Args, 1)
	assert.Equal(t, "id", userField.Args[0].Name)

	usersField, ok := schema.Query.Fields["users"]
	require.True(t, ok, "users field should exist on Query type")
	assert.Equal(t, "[User]", usersField.Type)
}

// TestGenerateSDL verifies SDL generation from schema
func TestGenerateSDL(t *testing.T) {
	typeDefs := map[string]interpreter.TypeDef{
		"User": {
			Name: "User",
			Fields: []interpreter.Field{
				{Name: "id", TypeAnnotation: interpreter.IntType{}, Required: true},
				{Name: "name", TypeAnnotation: interpreter.StringType{}},
			},
		},
	}
	resolvers := map[string]interpreter.GraphQLResolver{
		"query.user": {
			Operation:  interpreter.GraphQLQuery,
			FieldName:  "user",
			Params:     []interpreter.Field{{Name: "id", TypeAnnotation: interpreter.IntType{}, Required: true}},
			ReturnType: interpreter.NamedType{Name: "User"},
		},
	}

	schema := BuildSchema(typeDefs, resolvers)
	require.NotNil(t, schema)
	sdl := schema.GenerateSDL()

	assert.Contains(t, sdl, "type User")
	assert.Contains(t, sdl, "type Query")
	assert.Contains(t, sdl, "id: Int")
	assert.Contains(t, sdl, "user(id: Int!): User")
}

// TestTypeToGraphQL verifies Glyph type to GraphQL type conversion
func TestTypeToGraphQL(t *testing.T) {
	tests := []struct {
		input    interpreter.Type
		expected string
	}{
		{interpreter.IntType{}, "Int"},
		{interpreter.StringType{}, "String"},
		{interpreter.BoolType{}, "Boolean"},
		{interpreter.FloatType{}, "Float"},
		{interpreter.NamedType{Name: "User"}, "User"},
		{interpreter.ArrayType{ElementType: interpreter.StringType{}}, "[String]"},
		{interpreter.OptionalType{InnerType: interpreter.IntType{}}, "Int"},
		{nil, "String"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, typeToGraphQL(tt.input))
	}
}

// TestExecutorSimpleQuery verifies end-to-end query execution
func TestExecutorSimpleQuery(t *testing.T) {
	interp := interpreter.NewInterpreter()

	// Load a module with a query resolver that returns a literal
	module := interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.GraphQLResolver{
				Operation:  interpreter.GraphQLQuery,
				FieldName:  "hello",
				ReturnType: interpreter.StringType{},
				Body: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.LiteralExpr{
							Value: interpreter.StringLiteral{Value: "Hello, GraphQL!"},
						},
					},
				},
			},
		},
	}
	err := interp.LoadModule(module)
	require.NoError(t, err, "LoadModule should not return an error")

	resolvers := interp.GetGraphQLResolvers()
	schema := BuildSchema(interp.GetTypeDefs(), resolvers)
	require.NotNil(t, schema)
	executor := NewExecutor(schema, interp)

	resp := executor.Execute(GraphQLRequest{
		Query: `{ hello }`,
	}, nil)

	require.Empty(t, resp.Errors, "Execute should not return errors")
	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok, "response data should be a map")
	assert.Equal(t, "Hello, GraphQL!", data["hello"])
}

// TestExecutorQueryWithArgs verifies resolver execution with arguments
func TestExecutorQueryWithArgs(t *testing.T) {
	interp := interpreter.NewInterpreter()

	module := interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.GraphQLResolver{
				Operation:  interpreter.GraphQLQuery,
				FieldName:  "greet",
				Params:     []interpreter.Field{{Name: "name", TypeAnnotation: interpreter.StringType{}, Required: true}},
				ReturnType: interpreter.StringType{},
				Body: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "Hello, "}},
							Right: interpreter.VariableExpr{Name: "name"},
						},
					},
				},
			},
		},
	}
	err := interp.LoadModule(module)
	require.NoError(t, err, "LoadModule should not return an error")

	resolvers := interp.GetGraphQLResolvers()
	schema := BuildSchema(interp.GetTypeDefs(), resolvers)
	require.NotNil(t, schema)
	executor := NewExecutor(schema, interp)

	resp := executor.Execute(GraphQLRequest{
		Query: `{ greet(name: "World") }`,
	}, nil)

	require.Empty(t, resp.Errors, "Execute should not return errors")
	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok, "response data should be a map")
	assert.Equal(t, "Hello, World", data["greet"])
}

// TestExecutorMutation verifies mutation execution
func TestExecutorMutation(t *testing.T) {
	interp := interpreter.NewInterpreter()

	module := interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.GraphQLResolver{
				Operation:  interpreter.GraphQLMutation,
				FieldName:  "createItem",
				Params:     []interpreter.Field{{Name: "name", TypeAnnotation: interpreter.StringType{}, Required: true}},
				ReturnType: interpreter.NamedType{Name: "Item"},
				Body: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.ObjectExpr{
							Fields: []interpreter.ObjectField{
								{Key: "id", Value: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}}},
								{Key: "name", Value: interpreter.VariableExpr{Name: "name"}},
							},
						},
					},
				},
			},
		},
	}
	err := interp.LoadModule(module)
	require.NoError(t, err, "LoadModule should not return an error")

	resolvers := interp.GetGraphQLResolvers()
	schema := BuildSchema(interp.GetTypeDefs(), resolvers)
	require.NotNil(t, schema)
	executor := NewExecutor(schema, interp)

	resp := executor.Execute(GraphQLRequest{
		Query: `mutation { createItem(name: "Widget") { id name } }`,
	}, nil)

	require.Empty(t, resp.Errors, "Execute should not return errors")
	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok, "response data should be a map")
	item, ok := data["createItem"].(map[string]interface{})
	require.True(t, ok, "createItem should be a map")
	assert.Equal(t, int64(1), item["id"])
	assert.Equal(t, "Widget", item["name"])
}

// TestExecutorMissingResolver verifies error on missing resolver
func TestExecutorMissingResolver(t *testing.T) {
	interp := interpreter.NewInterpreter()

	// Empty schema with no resolvers but with a Query type
	schema := &Schema{
		Types:     make(map[string]*ObjectType),
		Resolvers: make(map[string]interpreter.GraphQLResolver),
		Query:     &ObjectType{Name: "Query", Fields: map[string]*FieldDef{}},
	}

	executor := NewExecutor(schema, interp)

	resp := executor.Execute(GraphQLRequest{
		Query: `{ nonexistent }`,
	}, nil)

	require.NotEmpty(t, resp.Errors)
	assert.Contains(t, resp.Errors[0].Message, "no resolver")
}

// TestExecutorUnsupportedOperation verifies error on subscription (not yet supported)
func TestExecutorUnsupportedOperation(t *testing.T) {
	interp := interpreter.NewInterpreter()
	schema := &Schema{
		Types:     make(map[string]*ObjectType),
		Resolvers: make(map[string]interpreter.GraphQLResolver),
		Query:     &ObjectType{Name: "Query", Fields: map[string]*FieldDef{}},
	}
	executor := NewExecutor(schema, interp)

	resp := executor.Execute(GraphQLRequest{
		Query: `subscription { events }`,
	}, nil)

	require.NotEmpty(t, resp.Errors)
	assert.Contains(t, resp.Errors[0].Message, "unsupported operation")
}

// TestExecutorFieldSelection verifies that field selections filter the result
func TestExecutorFieldSelection(t *testing.T) {
	interp := interpreter.NewInterpreter()

	module := interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.GraphQLResolver{
				Operation:  interpreter.GraphQLQuery,
				FieldName:  "user",
				ReturnType: interpreter.NamedType{Name: "User"},
				Body: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.ObjectExpr{
							Fields: []interpreter.ObjectField{
								{Key: "id", Value: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}}},
								{Key: "name", Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "Alice"}}},
								{Key: "email", Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "alice@example.com"}}},
								{Key: "secret", Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "hidden"}}},
							},
						},
					},
				},
			},
		},
	}
	err := interp.LoadModule(module)
	require.NoError(t, err, "LoadModule should not return an error")

	resolvers := interp.GetGraphQLResolvers()
	schema := BuildSchema(interp.GetTypeDefs(), resolvers)
	require.NotNil(t, schema)
	executor := NewExecutor(schema, interp)

	// Only request name and email fields - id and secret should be excluded
	resp := executor.Execute(GraphQLRequest{
		Query: `{ user { name email } }`,
	}, nil)

	require.Empty(t, resp.Errors, "Execute should not return errors")
	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok, "response data should be a map")
	user, ok := data["user"].(map[string]interface{})
	require.True(t, ok, "user should be a map")

	assert.Equal(t, "Alice", user["name"])
	assert.Equal(t, "alice@example.com", user["email"])
	_, hasSecret := user["secret"]
	assert.False(t, hasSecret, "secret field should not be present")
	_, hasID := user["id"]
	assert.False(t, hasID, "id field should not be present (not requested)")
}

// TestExecutorAlias verifies that aliases are applied in the response
func TestExecutorAlias(t *testing.T) {
	interp := interpreter.NewInterpreter()

	module := interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.GraphQLResolver{
				Operation:  interpreter.GraphQLQuery,
				FieldName:  "hello",
				ReturnType: interpreter.StringType{},
				Body: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "world"}},
					},
				},
			},
		},
	}
	err := interp.LoadModule(module)
	require.NoError(t, err, "LoadModule should not return an error")

	resolvers := interp.GetGraphQLResolvers()
	schema := BuildSchema(interp.GetTypeDefs(), resolvers)
	require.NotNil(t, schema)
	executor := NewExecutor(schema, interp)

	resp := executor.Execute(GraphQLRequest{
		Query: `{ greeting: hello }`,
	}, nil)

	require.Empty(t, resp.Errors, "Execute should not return errors")
	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok, "response data should be a map")
	assert.Equal(t, "world", data["greeting"])
	_, hasHello := data["hello"]
	assert.False(t, hasHello, "should use alias, not original field name")
}

// TestExecutorNoQueryType verifies error when no Query type is defined
func TestExecutorNoQueryType(t *testing.T) {
	interp := interpreter.NewInterpreter()
	schema := &Schema{
		Types:     make(map[string]*ObjectType),
		Resolvers: make(map[string]interpreter.GraphQLResolver),
	}
	executor := NewExecutor(schema, interp)

	resp := executor.Execute(GraphQLRequest{
		Query: `{ hello }`,
	}, nil)

	require.NotEmpty(t, resp.Errors)
	assert.Contains(t, resp.Errors[0].Message, "no query type defined")
}

// TestIntrospect verifies schema introspection returns type information
func TestIntrospect(t *testing.T) {
	typeDefs := map[string]interpreter.TypeDef{
		"User": {
			Name: "User",
			Fields: []interpreter.Field{
				{Name: "id", TypeAnnotation: interpreter.IntType{}},
			},
		},
	}
	resolvers := map[string]interpreter.GraphQLResolver{
		"query.user": {
			Operation:  interpreter.GraphQLQuery,
			FieldName:  "user",
			ReturnType: interpreter.NamedType{Name: "User"},
		},
	}

	schema := BuildSchema(typeDefs, resolvers)
	require.NotNil(t, schema)
	interp := interpreter.NewInterpreter()
	executor := NewExecutor(schema, interp)

	result := executor.Introspect()
	assert.Equal(t, "Query", result["queryType"])

	types, ok := result["types"].([]interface{})
	require.True(t, ok, "types should be a slice")
	assert.GreaterOrEqual(t, len(types), 2) // User + Query at minimum
}

// TestSelectionEffectiveName verifies EffectiveName returns alias or name
func TestSelectionEffectiveName(t *testing.T) {
	s1 := Selection{Name: "user"}
	assert.Equal(t, "user", s1.EffectiveName())

	s2 := Selection{Name: "user", Alias: "admin"}
	assert.Equal(t, "admin", s2.EffectiveName())
}

// TestGraphQLOperationTypeString verifies string representation of operation types
func TestGraphQLOperationTypeString(t *testing.T) {
	assert.Equal(t, "query", interpreter.GraphQLQuery.String())
	assert.Equal(t, "mutation", interpreter.GraphQLMutation.String())
	assert.Equal(t, "subscription", interpreter.GraphQLSubscription.String())
}

// TestApplySelectionsArray verifies field selection on array results
func TestApplySelectionsArray(t *testing.T) {
	executor := &Executor{schema: &Schema{}}

	input := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 30},
		map[string]interface{}{"name": "Bob", "age": 25},
	}

	result := executor.applySelections(input, []Selection{
		{Name: "name"},
	})

	arr, ok := result.([]interface{})
	require.True(t, ok, "result should be a slice")
	require.Len(t, arr, 2)

	first, ok := arr[0].(map[string]interface{})
	require.True(t, ok, "first element should be a map")
	assert.Equal(t, "Alice", first["name"])
	_, hasAge := first["age"]
	assert.False(t, hasAge, "age should not be present")
}

// TestApplySelectionsScalar verifies that scalar values pass through unchanged
func TestApplySelectionsScalar(t *testing.T) {
	executor := &Executor{schema: &Schema{}}
	result := executor.applySelections("hello", nil)
	assert.Equal(t, "hello", result)
}
