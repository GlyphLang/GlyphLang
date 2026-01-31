package graphql

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
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
	typeDefs := map[string]ast.TypeDef{
		"User": {
			Name: "User",
			Fields: []ast.Field{
				{Name: "id", TypeAnnotation: ast.IntType{}, Required: true},
				{Name: "name", TypeAnnotation: ast.StringType{}, Required: true},
				{Name: "email", TypeAnnotation: ast.StringType{}},
			},
		},
	}

	resolvers := map[string]ast.GraphQLResolver{
		"query.user": {
			Operation:  ast.GraphQLQuery,
			FieldName:  "user",
			Params:     []ast.Field{{Name: "id", TypeAnnotation: ast.IntType{}, Required: true}},
			ReturnType: ast.NamedType{Name: "User"},
		},
		"query.users": {
			Operation:  ast.GraphQLQuery,
			FieldName:  "users",
			Params:     []ast.Field{{Name: "limit", TypeAnnotation: ast.IntType{}}},
			ReturnType: ast.ArrayType{ElementType: ast.NamedType{Name: "User"}},
		},
		"mutation.createUser": {
			Operation:  ast.GraphQLMutation,
			FieldName:  "createUser",
			Params:     []ast.Field{{Name: "name", TypeAnnotation: ast.StringType{}, Required: true}},
			ReturnType: ast.NamedType{Name: "User"},
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
	typeDefs := map[string]ast.TypeDef{
		"User": {
			Name: "User",
			Fields: []ast.Field{
				{Name: "id", TypeAnnotation: ast.IntType{}, Required: true},
				{Name: "name", TypeAnnotation: ast.StringType{}},
			},
		},
	}
	resolvers := map[string]ast.GraphQLResolver{
		"query.user": {
			Operation:  ast.GraphQLQuery,
			FieldName:  "user",
			Params:     []ast.Field{{Name: "id", TypeAnnotation: ast.IntType{}, Required: true}},
			ReturnType: ast.NamedType{Name: "User"},
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
		input    ast.Type
		expected string
	}{
		{ast.IntType{}, "Int"},
		{ast.StringType{}, "String"},
		{ast.BoolType{}, "Boolean"},
		{ast.FloatType{}, "Float"},
		{ast.NamedType{Name: "User"}, "User"},
		{ast.ArrayType{ElementType: ast.StringType{}}, "[String]"},
		{ast.OptionalType{InnerType: ast.IntType{}}, "Int"},
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
	module := ast.Module{
		Items: []ast.Item{
			&ast.GraphQLResolver{
				Operation:  ast.GraphQLQuery,
				FieldName:  "hello",
				ReturnType: ast.StringType{},
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{
							Value: ast.StringLiteral{Value: "Hello, GraphQL!"},
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

	module := ast.Module{
		Items: []ast.Item{
			&ast.GraphQLResolver{
				Operation:  ast.GraphQLQuery,
				FieldName:  "greet",
				Params:     []ast.Field{{Name: "name", TypeAnnotation: ast.StringType{}, Required: true}},
				ReturnType: ast.StringType{},
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  ast.LiteralExpr{Value: ast.StringLiteral{Value: "Hello, "}},
							Right: ast.VariableExpr{Name: "name"},
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

	module := ast.Module{
		Items: []ast.Item{
			&ast.GraphQLResolver{
				Operation:  ast.GraphQLMutation,
				FieldName:  "createItem",
				Params:     []ast.Field{{Name: "name", TypeAnnotation: ast.StringType{}, Required: true}},
				ReturnType: ast.NamedType{Name: "Item"},
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.ObjectExpr{
							Fields: []ast.ObjectField{
								{Key: "id", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}}},
								{Key: "name", Value: ast.VariableExpr{Name: "name"}},
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
		Resolvers: make(map[string]ast.GraphQLResolver),
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
		Resolvers: make(map[string]ast.GraphQLResolver),
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

	module := ast.Module{
		Items: []ast.Item{
			&ast.GraphQLResolver{
				Operation:  ast.GraphQLQuery,
				FieldName:  "user",
				ReturnType: ast.NamedType{Name: "User"},
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.ObjectExpr{
							Fields: []ast.ObjectField{
								{Key: "id", Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}}},
								{Key: "name", Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "Alice"}}},
								{Key: "email", Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "alice@example.com"}}},
								{Key: "secret", Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "hidden"}}},
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

	module := ast.Module{
		Items: []ast.Item{
			&ast.GraphQLResolver{
				Operation:  ast.GraphQLQuery,
				FieldName:  "hello",
				ReturnType: ast.StringType{},
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "world"}},
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
		Resolvers: make(map[string]ast.GraphQLResolver),
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
	typeDefs := map[string]ast.TypeDef{
		"User": {
			Name: "User",
			Fields: []ast.Field{
				{Name: "id", TypeAnnotation: ast.IntType{}},
			},
		},
	}
	resolvers := map[string]ast.GraphQLResolver{
		"query.user": {
			Operation:  ast.GraphQLQuery,
			FieldName:  "user",
			ReturnType: ast.NamedType{Name: "User"},
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
	assert.Equal(t, "query", ast.GraphQLQuery.String())
	assert.Equal(t, "mutation", ast.GraphQLMutation.String())
	assert.Equal(t, "subscription", ast.GraphQLSubscription.String())
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

// TestSkipBalanced verifies parsing of queries with variable definitions
func TestSkipBalanced(t *testing.T) {
	q, err := ParseQuery(`query GetUser($id: Int!) { user }`)
	require.NoError(t, err)
	assert.Equal(t, "query", q.OperationType)
	assert.Equal(t, "GetUser", q.Name)
	require.Len(t, q.Selections, 1)
	assert.Equal(t, "user", q.Selections[0].Name)
}

// TestSkipBalancedNested verifies parsing of queries with nested variable definitions
func TestSkipBalancedNested(t *testing.T) {
	q, err := ParseQuery(`query Search($filter: Filter(inner: String)) { results }`)
	require.NoError(t, err)
	assert.Equal(t, "query", q.OperationType)
	assert.Equal(t, "Search", q.Name)
	require.Len(t, q.Selections, 1)
	assert.Equal(t, "results", q.Selections[0].Name)
}

// TestSkipBalancedUnbalanced verifies error on unbalanced parentheses
func TestSkipBalancedUnbalanced(t *testing.T) {
	_, err := ParseQuery(`query Broken($id: Int { user }`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unbalanced")
}

// TestReadStringEscapeSequences verifies parsing of escape sequences in strings
func TestReadStringEscapeSequences(t *testing.T) {
	// Escaped double quote
	q, err := ParseQuery(`{ user(name: "hello \"world\"") }`)
	require.NoError(t, err)
	assert.Equal(t, `hello "world"`, q.Selections[0].Args["name"])

	// Escaped backslash
	q, err = ParseQuery(`{ user(path: "C:\\Users\\admin") }`)
	require.NoError(t, err)
	assert.Equal(t, `C:\Users\admin`, q.Selections[0].Args["path"])

	// Escaped newline
	q, err = ParseQuery(`{ user(bio: "line1\nline2") }`)
	require.NoError(t, err)
	assert.Equal(t, "line1\nline2", q.Selections[0].Args["bio"])

	// Escaped tab
	q, err = ParseQuery(`{ user(data: "col1\tcol2") }`)
	require.NoError(t, err)
	assert.Equal(t, "col1\tcol2", q.Selections[0].Args["data"])

	// Other escape (unknown, passes through the character after backslash literally)
	q, err = ParseQuery(`{ user(data: "hello\rworld") }`)
	require.NoError(t, err)
	assert.Equal(t, "hellorworld", q.Selections[0].Args["data"])
}

// TestReadStringUnterminated verifies error on unterminated string
func TestReadStringUnterminated(t *testing.T) {
	_, err := ParseQuery(`{ user(name: "unterminated) }`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unterminated string")
}

// TestParseValueFalse verifies parsing false boolean argument
func TestParseValueFalse(t *testing.T) {
	q, err := ParseQuery(`{ users(active: false) }`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 1)
	assert.Equal(t, false, q.Selections[0].Args["active"])
}

// TestParseValueEnumArgument verifies parsing an enum argument (unquoted identifier)
func TestParseValueEnumArgument(t *testing.T) {
	q, err := ParseQuery(`{ users(status: ACTIVE) }`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 1)
	assert.Equal(t, "ACTIVE", q.Selections[0].Args["status"])
}

// TestParseValueUnexpectedCharacter verifies error on unexpected character in value
func TestParseValueUnexpectedCharacter(t *testing.T) {
	_, err := ParseQuery(`{ user(id: @bad) }`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected character")
}

// TestParseValueEndOfInput verifies error on unexpected end of input in value
func TestParseValueEndOfInput(t *testing.T) {
	_, err := ParseQuery(`{ user(id: `)
	assert.Error(t, err)
}

// TestParseValueNegativeNumber verifies parsing negative integer argument
func TestParseValueNegativeNumber(t *testing.T) {
	q, err := ParseQuery(`{ items(offset: -5) }`)
	require.NoError(t, err)
	require.Len(t, q.Selections, 1)
	assert.Equal(t, int64(-5), q.Selections[0].Args["offset"])
}

// TestParseArgumentsMissingColon verifies error when colon is missing after argument name
func TestParseArgumentsMissingColon(t *testing.T) {
	_, err := ParseQuery(`{ user(id 42) }`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected ':'")
}

// TestParseArgumentsUnterminatedArgs verifies error on unterminated arguments
func TestParseArgumentsUnterminatedArgs(t *testing.T) {
	_, err := ParseQuery(`{ user(id: 42 `)
	assert.Error(t, err)
}

// TestParseArgumentsEmptyArgName verifies error when argument name is empty
func TestParseArgumentsEmptyArgName(t *testing.T) {
	_, err := ParseQuery(`{ user(: 42) }`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected argument name")
}

// TestTypeToGraphQLUnionType verifies GraphQL type for union types
func TestTypeToGraphQLUnionType(t *testing.T) {
	// Union with types - uses first type
	unionType := interpreter.UnionType{
		Types: []interpreter.Type{interpreter.StringType{}, interpreter.IntType{}},
	}
	assert.Equal(t, "String", typeToGraphQL(unionType))

	// Empty union - defaults to String
	emptyUnion := interpreter.UnionType{Types: []interpreter.Type{}}
	assert.Equal(t, "String", typeToGraphQL(emptyUnion))
}

// TestConvertFieldArrayType verifies convertField sets IsList for array types
func TestConvertFieldArrayType(t *testing.T) {
	field := interpreter.Field{
		Name:           "tags",
		TypeAnnotation: interpreter.ArrayType{ElementType: interpreter.StringType{}},
		Required:       true,
	}
	fd := convertField(field)
	assert.Equal(t, "tags", fd.Name)
	assert.Equal(t, "[String]", fd.Type)
	assert.True(t, fd.IsList, "IsList should be true for array types")
	assert.False(t, fd.IsNullable, "IsNullable should be false when Required is true")
}

// TestConvertFieldNonArrayType verifies convertField for non-array types
func TestConvertFieldNonArrayType(t *testing.T) {
	field := interpreter.Field{
		Name:           "name",
		TypeAnnotation: interpreter.StringType{},
		Required:       false,
	}
	fd := convertField(field)
	assert.Equal(t, "name", fd.Name)
	assert.Equal(t, "String", fd.Type)
	assert.False(t, fd.IsList, "IsList should be false for non-array types")
	assert.True(t, fd.IsNullable, "IsNullable should be true when Required is false")
}

// TestIntrospectWithFieldArgs verifies Introspect returns argument info for fields
func TestIntrospectWithFieldArgs(t *testing.T) {
	typeDefs := map[string]interpreter.TypeDef{
		"User": {
			Name: "User",
			Fields: []interpreter.Field{
				{Name: "id", TypeAnnotation: interpreter.IntType{}},
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
		"mutation.updateUser": {
			Operation:  interpreter.GraphQLMutation,
			FieldName:  "updateUser",
			Params:     []interpreter.Field{{Name: "id", TypeAnnotation: interpreter.IntType{}, Required: true}, {Name: "name", TypeAnnotation: interpreter.StringType{}, Required: false}},
			ReturnType: interpreter.NamedType{Name: "User"},
		},
	}

	schema := BuildSchema(typeDefs, resolvers)
	require.NotNil(t, schema)
	interp := interpreter.NewInterpreter()
	executor := NewExecutor(schema, interp)

	result := executor.Introspect()

	// Verify queryType and mutationType are set
	assert.Equal(t, "Query", result["queryType"])
	assert.Equal(t, "Mutation", result["mutationType"])

	// Verify types are returned
	types, ok := result["types"].([]interface{})
	require.True(t, ok, "types should be a slice")

	// Find the Query type and verify it has fields with args
	foundQueryWithArgs := false
	for _, typeInfo := range types {
		typeMap, ok := typeInfo.(map[string]interface{})
		if !ok {
			continue
		}
		if typeMap["name"] == "Query" {
			fields, ok := typeMap["fields"].([]interface{})
			if !ok {
				continue
			}
			for _, f := range fields {
				fieldMap, ok := f.(map[string]interface{})
				if !ok {
					continue
				}
				if fieldMap["name"] == "user" {
					args, hasArgs := fieldMap["args"]
					if hasArgs {
						foundQueryWithArgs = true
						argSlice, ok := args.([]interface{})
						require.True(t, ok, "args should be a slice")
						require.Len(t, argSlice, 1)
						argMap, ok := argSlice[0].(map[string]interface{})
						require.True(t, ok, "arg should be a map")
						assert.Equal(t, "id", argMap["name"])
						assert.Equal(t, "Int", argMap["type"])
						assert.Equal(t, true, argMap["required"])
					}
				}
			}
		}
	}
	assert.True(t, foundQueryWithArgs, "should find Query type with field args")
}

// TestIntrospectNoMutation verifies Introspect omits mutationType when not set
func TestIntrospectNoMutation(t *testing.T) {
	typeDefs := map[string]interpreter.TypeDef{}
	resolvers := map[string]interpreter.GraphQLResolver{
		"query.hello": {
			Operation:  interpreter.GraphQLQuery,
			FieldName:  "hello",
			ReturnType: interpreter.StringType{},
		},
	}

	schema := BuildSchema(typeDefs, resolvers)
	require.NotNil(t, schema)
	interp := interpreter.NewInterpreter()
	executor := NewExecutor(schema, interp)

	result := executor.Introspect()
	assert.Equal(t, "Query", result["queryType"])
	_, hasMutation := result["mutationType"]
	assert.False(t, hasMutation, "mutationType should not be present when no mutation resolvers exist")
}

// TestExecutorParseError verifies executor returns error on invalid query syntax
func TestExecutorParseError(t *testing.T) {
	interp := interpreter.NewInterpreter()
	schema := &Schema{
		Types:     make(map[string]*ObjectType),
		Resolvers: make(map[string]interpreter.GraphQLResolver),
		Query:     &ObjectType{Name: "Query", Fields: map[string]*FieldDef{}},
	}
	executor := NewExecutor(schema, interp)

	resp := executor.Execute(GraphQLRequest{
		Query: `not a valid query`,
	}, nil)

	require.NotEmpty(t, resp.Errors)
	assert.Contains(t, resp.Errors[0].Message, "parse error")
}

// TestExecutorVariablesMerge verifies that request variables are merged into the query
func TestExecutorVariablesMerge(t *testing.T) {
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
	require.NoError(t, err)

	resolvers := interp.GetGraphQLResolvers()
	schema := BuildSchema(interp.GetTypeDefs(), resolvers)
	require.NotNil(t, schema)
	executor := NewExecutor(schema, interp)

	resp := executor.Execute(GraphQLRequest{
		Query:     `{ hello }`,
		Variables: map[string]interface{}{"key": "value"},
	}, nil)

	require.Empty(t, resp.Errors)
}

// TestExecutorNoMutationType verifies error when mutation type is not defined
func TestExecutorNoMutationType(t *testing.T) {
	interp := interpreter.NewInterpreter()
	schema := &Schema{
		Types:     make(map[string]*ObjectType),
		Resolvers: make(map[string]interpreter.GraphQLResolver),
		// Query is set but Mutation is nil
		Query: &ObjectType{Name: "Query", Fields: map[string]*FieldDef{}},
	}
	executor := NewExecutor(schema, interp)

	resp := executor.Execute(GraphQLRequest{
		Query: `mutation { createUser }`,
	}, nil)

	require.NotEmpty(t, resp.Errors)
	assert.Contains(t, resp.Errors[0].Message, "no mutation type defined")
}
