package interpreter

import (
	"testing"
	"time"

	. "github.com/glyphlang/glyph/pkg/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Environment: GetAll / GetLocal
// =============================================================================

func TestEnvironment_GetAll_Empty(t *testing.T) {
	env := NewEnvironment()
	all := env.GetAll()
	assert.Empty(t, all)
}

func TestEnvironment_GetAll_SingleScope(t *testing.T) {
	env := NewEnvironment()
	env.Define("x", int64(1))
	env.Define("y", "hello")

	all := env.GetAll()
	assert.Equal(t, int64(1), all["x"])
	assert.Equal(t, "hello", all["y"])
	assert.Len(t, all, 2)
}

func TestEnvironment_GetAll_InheritsParent(t *testing.T) {
	parent := NewEnvironment()
	parent.Define("a", int64(10))
	parent.Define("b", "parent")

	child := NewChildEnvironment(parent)
	child.Define("c", true)

	all := child.GetAll()
	assert.Equal(t, int64(10), all["a"])
	assert.Equal(t, "parent", all["b"])
	assert.Equal(t, true, all["c"])
	assert.Len(t, all, 3)
}

func TestEnvironment_GetAll_ChildOverridesParent(t *testing.T) {
	parent := NewEnvironment()
	parent.Define("x", int64(1))
	parent.Define("y", int64(2))

	child := NewChildEnvironment(parent)
	child.Define("x", int64(99)) // Override parent's x

	all := child.GetAll()
	assert.Equal(t, int64(99), all["x"], "child value should override parent")
	assert.Equal(t, int64(2), all["y"])
	assert.Len(t, all, 2)
}

func TestEnvironment_GetAll_ThreeLevels(t *testing.T) {
	grandparent := NewEnvironment()
	grandparent.Define("a", int64(1))

	parent := NewChildEnvironment(grandparent)
	parent.Define("b", int64(2))

	child := NewChildEnvironment(parent)
	child.Define("c", int64(3))

	all := child.GetAll()
	assert.Len(t, all, 3)
	assert.Equal(t, int64(1), all["a"])
	assert.Equal(t, int64(2), all["b"])
	assert.Equal(t, int64(3), all["c"])
}

func TestEnvironment_GetLocal_Empty(t *testing.T) {
	env := NewEnvironment()
	local := env.GetLocal()
	assert.Empty(t, local)
}

func TestEnvironment_GetLocal_OnlyLocalVars(t *testing.T) {
	parent := NewEnvironment()
	parent.Define("a", int64(10))

	child := NewChildEnvironment(parent)
	child.Define("b", int64(20))

	local := child.GetLocal()
	assert.Len(t, local, 1)
	assert.Equal(t, int64(20), local["b"])
	_, hasA := local["a"]
	assert.False(t, hasA, "GetLocal should not include parent vars")
}

func TestEnvironment_GetLocal_MultipleVars(t *testing.T) {
	env := NewEnvironment()
	env.Define("x", int64(1))
	env.Define("y", "two")
	env.Define("z", true)

	local := env.GetLocal()
	assert.Len(t, local, 3)
	assert.Equal(t, int64(1), local["x"])
	assert.Equal(t, "two", local["y"])
	assert.Equal(t, true, local["z"])
}

// =============================================================================
// Future: State / Done
// =============================================================================

func TestFuture_State_Pending(t *testing.T) {
	f := NewFuture()
	assert.Equal(t, FuturePending, f.State())
}

func TestFuture_State_Resolved(t *testing.T) {
	f := NewFuture()
	f.Resolve("value")
	assert.Equal(t, FutureResolved, f.State())
}

func TestFuture_State_Rejected(t *testing.T) {
	f := NewFuture()
	f.Reject(assert.AnError)
	assert.Equal(t, FutureRejected, f.State())
}

func TestFuture_Done_ClosedOnResolve(t *testing.T) {
	f := NewFuture()
	f.Resolve("done")

	select {
	case <-f.Done():
		// Channel should be closed, this path is correct
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Done() channel should be closed after Resolve")
	}
}

func TestFuture_Done_ClosedOnReject(t *testing.T) {
	f := NewFuture()
	f.Reject(assert.AnError)

	select {
	case <-f.Done():
		// Channel should be closed, this path is correct
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Done() channel should be closed after Reject")
	}
}

func TestFuture_Done_NotClosedWhilePending(t *testing.T) {
	f := NewFuture()

	select {
	case <-f.Done():
		t.Fatal("Done() channel should not be closed while pending")
	case <-time.After(10 * time.Millisecond):
		// Correct - channel is still open
	}
}

func TestFuture_Done_SelectMultiple(t *testing.T) {
	f1 := NewFuture()
	f2 := NewFuture()

	// Resolve f1 first
	f1.Resolve("first")

	select {
	case <-f1.Done():
		// Expected
	case <-f2.Done():
		t.Fatal("f2 should not be done yet")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("f1 should be done")
	}
}

func TestFutureState_String(t *testing.T) {
	assert.Equal(t, "pending", FuturePending.String())
	assert.Equal(t, "resolved", FutureResolved.String())
	assert.Equal(t, "rejected", FutureRejected.String())
	assert.Equal(t, "unknown", FutureState(99).String())
}

// =============================================================================
// Interpreter getter functions: GetContract, GetContracts, GetTypeDefs, etc.
// =============================================================================

func TestInterpreter_GetContract(t *testing.T) {
	interp := NewInterpreter()

	// No contracts initially
	_, ok := interp.GetContract("UserService")
	assert.False(t, ok)

	// Load a contract
	contract := &ContractDef{
		Name: "UserService",
		Endpoints: []ContractEndpoint{
			{Method: Get, Path: "/users/:id", ReturnType: NamedType{Name: "User"}},
		},
	}
	err := interp.LoadModule(Module{Items: []Item{contract}})
	require.NoError(t, err)

	// Now should be found
	c, ok := interp.GetContract("UserService")
	assert.True(t, ok)
	assert.Equal(t, "UserService", c.Name)
	assert.Len(t, c.Endpoints, 1)
}

func TestInterpreter_GetContracts(t *testing.T) {
	interp := NewInterpreter()

	// Empty initially
	contracts := interp.GetContracts()
	assert.Empty(t, contracts)

	// Load multiple contracts
	err := interp.LoadModule(Module{
		Items: []Item{
			&ContractDef{Name: "ServiceA", Endpoints: []ContractEndpoint{}},
			&ContractDef{Name: "ServiceB", Endpoints: []ContractEndpoint{
				{Method: Post, Path: "/items", ReturnType: IntType{}},
			}},
		},
	})
	require.NoError(t, err)

	contracts = interp.GetContracts()
	assert.Len(t, contracts, 2)
	_, hasA := contracts["ServiceA"]
	assert.True(t, hasA)
	_, hasB := contracts["ServiceB"]
	assert.True(t, hasB)
}

func TestInterpreter_GetTypeDefs(t *testing.T) {
	interp := NewInterpreter()

	// Empty initially
	defs := interp.GetTypeDefs()
	assert.Empty(t, defs)

	// Load type definitions
	err := interp.LoadModule(Module{
		Items: []Item{
			&TypeDef{
				Name: "User",
				Fields: []Field{
					{Name: "id", TypeAnnotation: IntType{}, Required: true},
					{Name: "name", TypeAnnotation: StringType{}, Required: true},
				},
			},
			&TypeDef{
				Name: "Post",
				Fields: []Field{
					{Name: "title", TypeAnnotation: StringType{}, Required: true},
				},
			},
		},
	})
	require.NoError(t, err)

	defs = interp.GetTypeDefs()
	assert.Len(t, defs, 2)
	user, ok := defs["User"]
	assert.True(t, ok)
	assert.Equal(t, "User", user.Name)
	assert.Len(t, user.Fields, 2)
}

func TestInterpreter_GetGRPCServices(t *testing.T) {
	interp := NewInterpreter()

	// Empty initially
	services := interp.GetGRPCServices()
	assert.Empty(t, services)

	// Load gRPC services
	err := interp.LoadModule(Module{
		Items: []Item{
			&GRPCService{
				Name: "UserService",
				Methods: []GRPCMethod{
					{Name: "GetUser", InputType: NamedType{Name: "GetUserRequest"}, ReturnType: NamedType{Name: "User"}},
				},
			},
			&GRPCService{
				Name: "OrderService",
				Methods: []GRPCMethod{
					{Name: "CreateOrder", InputType: NamedType{Name: "OrderRequest"}, ReturnType: NamedType{Name: "Order"}},
				},
			},
		},
	})
	require.NoError(t, err)

	services = interp.GetGRPCServices()
	assert.Len(t, services, 2)
	_, hasUser := services["UserService"]
	assert.True(t, hasUser)
	_, hasOrder := services["OrderService"]
	assert.True(t, hasOrder)
}

func TestInterpreter_GetGRPCHandlers(t *testing.T) {
	interp := NewInterpreter()

	// Empty initially
	handlers := interp.GetGRPCHandlers()
	assert.Empty(t, handlers)

	// Load gRPC handlers
	err := interp.LoadModule(Module{
		Items: []Item{
			&GRPCHandler{
				MethodName: "GetUser",
				Params:     []Field{{Name: "id", TypeAnnotation: IntType{}, Required: true}},
				ReturnType: NamedType{Name: "User"},
				Body: []Statement{
					ReturnStatement{Value: ObjectExpr{
						Fields: []ObjectField{
							{Key: "id", Value: LiteralExpr{Value: IntLiteral{Value: 1}}},
						},
					}},
				},
			},
		},
	})
	require.NoError(t, err)

	handlers = interp.GetGRPCHandlers()
	assert.Len(t, handlers, 1)
	_, hasGetUser := handlers["GetUser"]
	assert.True(t, hasGetUser)
}

func TestInterpreter_GetGraphQLResolvers(t *testing.T) {
	interp := NewInterpreter()

	// Empty initially
	resolvers := interp.GetGraphQLResolvers()
	assert.Empty(t, resolvers)

	// Load GraphQL resolvers
	err := interp.LoadModule(Module{
		Items: []Item{
			&GraphQLResolver{
				Operation: GraphQLQuery,
				FieldName: "user",
				Params:    []Field{{Name: "id", TypeAnnotation: IntType{}, Required: true}},
				Body: []Statement{
					ReturnStatement{Value: ObjectExpr{
						Fields: []ObjectField{
							{Key: "id", Value: LiteralExpr{Value: IntLiteral{Value: 1}}},
							{Key: "name", Value: LiteralExpr{Value: StringLiteral{Value: "Alice"}}},
						},
					}},
				},
			},
			&GraphQLResolver{
				Operation: GraphQLMutation,
				FieldName: "createUser",
				Params:    []Field{{Name: "name", TypeAnnotation: StringType{}, Required: true}},
				Body: []Statement{
					ReturnStatement{Value: LiteralExpr{Value: BoolLiteral{Value: true}}},
				},
			},
		},
	})
	require.NoError(t, err)

	resolvers = interp.GetGraphQLResolvers()
	assert.Len(t, resolvers, 2)
	_, hasQuery := resolvers["query.user"]
	assert.True(t, hasQuery)
	_, hasMutation := resolvers["mutation.createUser"]
	assert.True(t, hasMutation)
}

// =============================================================================
// SetLLMHandler
// =============================================================================

func TestInterpreter_SetLLMHandler(t *testing.T) {
	interp := NewInterpreter()

	// Set a mock LLM handler
	mockHandler := "mock-llm-handler"
	interp.SetLLMHandler(mockHandler)
	assert.Equal(t, mockHandler, interp.llmHandler)
}

// =============================================================================
// ExecuteGRPCHandler
// =============================================================================

func TestInterpreter_ExecuteGRPCHandler_Nil(t *testing.T) {
	interp := NewInterpreter()
	_, err := interp.ExecuteGRPCHandler(nil, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler is nil")
}

func TestInterpreter_ExecuteGRPCHandler_SimpleReturn(t *testing.T) {
	interp := NewInterpreter()

	handler := &GRPCHandler{
		MethodName: "GetUser",
		Params: []Field{
			{Name: "id", TypeAnnotation: IntType{}, Required: true},
		},
		Body: []Statement{
			ReturnStatement{Value: ObjectExpr{
				Fields: []ObjectField{
					{Key: "id", Value: VariableExpr{Name: "id"}},
					{Key: "name", Value: LiteralExpr{Value: StringLiteral{Value: "Alice"}}},
				},
			}},
		},
	}

	result, err := interp.ExecuteGRPCHandler(handler, map[string]interface{}{"id": int64(42)}, nil)
	require.NoError(t, err)

	obj, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(42), obj["id"])
	assert.Equal(t, "Alice", obj["name"])
}

func TestInterpreter_ExecuteGRPCHandler_MissingRequired(t *testing.T) {
	interp := NewInterpreter()

	handler := &GRPCHandler{
		MethodName: "GetUser",
		Params: []Field{
			{Name: "id", TypeAnnotation: IntType{}, Required: true},
		},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: IntLiteral{Value: 0}}},
		},
	}

	_, err := interp.ExecuteGRPCHandler(handler, map[string]interface{}{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

func TestInterpreter_ExecuteGRPCHandler_DefaultParam(t *testing.T) {
	interp := NewInterpreter()

	handler := &GRPCHandler{
		MethodName: "ListUsers",
		Params: []Field{
			{Name: "limit", TypeAnnotation: IntType{}, Default: LiteralExpr{Value: IntLiteral{Value: 10}}},
		},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "limit"}},
		},
	}

	result, err := interp.ExecuteGRPCHandler(handler, map[string]interface{}{}, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(10), result)
}

func TestInterpreter_ExecuteGRPCHandler_OptionalParamNil(t *testing.T) {
	interp := NewInterpreter()

	handler := &GRPCHandler{
		MethodName: "Search",
		Params: []Field{
			{Name: "query", TypeAnnotation: StringType{}, Required: true},
			{Name: "filter", TypeAnnotation: StringType{}}, // optional, no default
		},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "query"}},
		},
	}

	result, err := interp.ExecuteGRPCHandler(handler, map[string]interface{}{"query": "test"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "test", result)
}

func TestInterpreter_ExecuteGRPCHandler_WithAuth(t *testing.T) {
	interp := NewInterpreter()

	handler := &GRPCHandler{
		MethodName: "SecureMethod",
		Auth:       &AuthConfig{AuthType: "jwt"},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "authorized"}}},
		},
	}

	authData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   int64(1),
			"role": "admin",
		},
	}

	result, err := interp.ExecuteGRPCHandler(handler, map[string]interface{}{}, authData)
	require.NoError(t, err)
	assert.Equal(t, "authorized", result)
}

// =============================================================================
// ExecuteGraphQLResolver
// =============================================================================

func TestInterpreter_ExecuteGraphQLResolver_Nil(t *testing.T) {
	interp := NewInterpreter()
	_, err := interp.ExecuteGraphQLResolver(nil, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resolver is nil")
}

func TestInterpreter_ExecuteGraphQLResolver_SimpleReturn(t *testing.T) {
	interp := NewInterpreter()

	resolver := &GraphQLResolver{
		Operation: GraphQLQuery,
		FieldName: "user",
		Params: []Field{
			{Name: "id", TypeAnnotation: IntType{}, Required: true},
		},
		Body: []Statement{
			ReturnStatement{Value: ObjectExpr{
				Fields: []ObjectField{
					{Key: "id", Value: VariableExpr{Name: "id"}},
					{Key: "name", Value: LiteralExpr{Value: StringLiteral{Value: "Bob"}}},
				},
			}},
		},
	}

	result, err := interp.ExecuteGraphQLResolver(resolver, map[string]interface{}{"id": int64(5)}, nil)
	require.NoError(t, err)

	obj, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(5), obj["id"])
	assert.Equal(t, "Bob", obj["name"])
}

func TestInterpreter_ExecuteGraphQLResolver_MissingRequired(t *testing.T) {
	interp := NewInterpreter()

	resolver := &GraphQLResolver{
		Operation: GraphQLQuery,
		FieldName: "user",
		Params: []Field{
			{Name: "id", TypeAnnotation: IntType{}, Required: true},
		},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: IntLiteral{Value: 0}}},
		},
	}

	_, err := interp.ExecuteGraphQLResolver(resolver, map[string]interface{}{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

func TestInterpreter_ExecuteGraphQLResolver_DefaultParam(t *testing.T) {
	interp := NewInterpreter()

	resolver := &GraphQLResolver{
		Operation: GraphQLQuery,
		FieldName: "users",
		Params: []Field{
			{Name: "limit", TypeAnnotation: IntType{}, Default: LiteralExpr{Value: IntLiteral{Value: 25}}},
		},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "limit"}},
		},
	}

	result, err := interp.ExecuteGraphQLResolver(resolver, map[string]interface{}{}, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(25), result)
}

func TestInterpreter_ExecuteGraphQLResolver_OptionalParamNil(t *testing.T) {
	interp := NewInterpreter()

	resolver := &GraphQLResolver{
		Operation: GraphQLQuery,
		FieldName: "search",
		Params: []Field{
			{Name: "query", TypeAnnotation: StringType{}, Required: true},
			{Name: "filter"}, // optional, no default, no type
		},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "query"}},
		},
	}

	result, err := interp.ExecuteGraphQLResolver(resolver, map[string]interface{}{"query": "find"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "find", result)
}

func TestInterpreter_ExecuteGraphQLResolver_WithAuth(t *testing.T) {
	interp := NewInterpreter()

	resolver := &GraphQLResolver{
		Operation: GraphQLMutation,
		FieldName: "deleteUser",
		Auth:      &AuthConfig{AuthType: "jwt"},
		Params:    []Field{{Name: "id", TypeAnnotation: IntType{}, Required: true}},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: BoolLiteral{Value: true}}},
		},
	}

	authData := map[string]interface{}{
		"user": map[string]interface{}{"id": int64(1), "role": "admin"},
	}

	result, err := interp.ExecuteGraphQLResolver(resolver, map[string]interface{}{"id": int64(99)}, authData)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

// =============================================================================
// capitalizeFirst
// =============================================================================

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"a", "A"},
		{"hello", "Hello"},
		{"Hello", "Hello"},
		{"countWhere", "CountWhere"},
		{"nextId", "NextId"},
		{"ABC", "ABC"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, capitalizeFirst(tt.input))
		})
	}
}

// =============================================================================
// evaluateArrayIndexExpr edge cases
// =============================================================================

func TestEvaluateArrayIndexExpr_ValidIndex(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(10), int64(20), int64(30)}
	env.Define("arr", arr)

	result, err := interp.EvaluateExpression(ArrayIndexExpr{
		Array: VariableExpr{Name: "arr"},
		Index: LiteralExpr{Value: IntLiteral{Value: 1}},
	}, env)
	require.NoError(t, err)
	assert.Equal(t, int64(20), result)
}

func TestEvaluateArrayIndexExpr_OutOfBounds(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(10), int64(20)}
	env.Define("arr", arr)

	_, err := interp.EvaluateExpression(ArrayIndexExpr{
		Array: VariableExpr{Name: "arr"},
		Index: LiteralExpr{Value: IntLiteral{Value: 5}},
	}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "array index out of bounds")
}

func TestEvaluateArrayIndexExpr_NegativeIndex(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(10)}
	env.Define("arr", arr)

	_, err := interp.EvaluateExpression(ArrayIndexExpr{
		Array: VariableExpr{Name: "arr"},
		Index: LiteralExpr{Value: IntLiteral{Value: -1}},
	}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "array index out of bounds")
}

func TestEvaluateArrayIndexExpr_NonIntIndex(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	arr := []interface{}{int64(10)}
	env.Define("arr", arr)

	_, err := interp.EvaluateExpression(ArrayIndexExpr{
		Array: VariableExpr{Name: "arr"},
		Index: LiteralExpr{Value: StringLiteral{Value: "not an int"}},
	}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "array index must be an integer")
}

func TestEvaluateArrayIndexExpr_MapStringKey(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	obj := map[string]interface{}{"name": "Alice", "age": int64(30)}
	env.Define("obj", obj)

	result, err := interp.EvaluateExpression(ArrayIndexExpr{
		Array: VariableExpr{Name: "obj"},
		Index: LiteralExpr{Value: StringLiteral{Value: "name"}},
	}, env)
	require.NoError(t, err)
	assert.Equal(t, "Alice", result)
}

func TestEvaluateArrayIndexExpr_MapNonStringKey(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	obj := map[string]interface{}{"name": "Alice"}
	env.Define("obj", obj)

	_, err := interp.EvaluateExpression(ArrayIndexExpr{
		Array: VariableExpr{Name: "obj"},
		Index: LiteralExpr{Value: IntLiteral{Value: 0}},
	}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "map key must be a string")
}

func TestEvaluateArrayIndexExpr_MapMissingKey(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	obj := map[string]interface{}{"name": "Alice"}
	env.Define("obj", obj)

	result, err := interp.EvaluateExpression(ArrayIndexExpr{
		Array: VariableExpr{Name: "obj"},
		Index: LiteralExpr{Value: StringLiteral{Value: "missing"}},
	}, env)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestEvaluateArrayIndexExpr_NonIndexable(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("x", int64(42))

	_, err := interp.EvaluateExpression(ArrayIndexExpr{
		Array: VariableExpr{Name: "x"},
		Index: LiteralExpr{Value: IntLiteral{Value: 0}},
	}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot index")
}

// =============================================================================
// evaluateResultMethod
// =============================================================================

func TestEvaluateResultMethod_IsOk(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk(int64(42)))

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.isOk",
		Args: []Expr{},
	}, env)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateResultMethod_IsErr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("fail"))

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.isErr",
		Args: []Expr{},
	}, env)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateResultMethod_UnwrapOk(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk("hello"))

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.unwrap",
		Args: []Expr{},
	}, env)
	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestEvaluateResultMethod_UnwrapOnErr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("oops"))

	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.unwrap",
		Args: []Expr{},
	}, env)
	assert.Error(t, err)
}

func TestEvaluateResultMethod_UnwrapOr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("fail"))

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.unwrapOr",
		Args: []Expr{LiteralExpr{Value: IntLiteral{Value: 99}}},
	}, env)
	require.NoError(t, err)
	assert.Equal(t, int64(99), result)
}

func TestEvaluateResultMethod_UnwrapOrWrongArgs(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("fail"))

	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.unwrapOr",
		Args: []Expr{},
	}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unwrapOr() expects 1 argument")
}

func TestEvaluateResultMethod_UnwrapErr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("error msg"))

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.unwrapErr",
		Args: []Expr{},
	}, env)
	require.NoError(t, err)
	assert.Equal(t, "error msg", result)
}

func TestEvaluateResultMethod_UnwrapErrOnOk(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk(int64(42)))

	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.unwrapErr",
		Args: []Expr{},
	}, env)
	assert.Error(t, err)
}

func TestEvaluateResultMethod_Map_OkTransformed(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk(int64(5)))

	// Create a lambda that multiplies by 2: (v -> v * 2)
	lambda := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "v"}},
			Body: BinaryOpExpr{
				Op:    Mul,
				Left:  VariableExpr{Name: "v"},
				Right: LiteralExpr{Value: IntLiteral{Value: 2}},
			},
		},
		Env: env,
	}
	env.Define("double", lambda)

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.map",
		Args: []Expr{VariableExpr{Name: "double"}},
	}, env)
	require.NoError(t, err)

	rv, ok := result.(*ResultValue)
	require.True(t, ok)
	assert.True(t, rv.IsOk())
	val, _ := rv.Unwrap()
	assert.Equal(t, int64(10), val)
}

func TestEvaluateResultMethod_Map_ErrPassedThrough(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("fail"))

	lambda := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "v"}},
			Body:   LiteralExpr{Value: IntLiteral{Value: 99}},
		},
		Env: env,
	}
	env.Define("fn", lambda)

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.map",
		Args: []Expr{VariableExpr{Name: "fn"}},
	}, env)
	require.NoError(t, err)

	rv, ok := result.(*ResultValue)
	require.True(t, ok)
	assert.True(t, rv.IsErr(), "map on Err should return Err unchanged")
}

func TestEvaluateResultMethod_Map_WrongArgCount(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk(int64(5)))

	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.map",
		Args: []Expr{},
	}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "map() expects 1 argument")
}

func TestEvaluateResultMethod_MapErr_TransformsErr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("original error"))

	lambda := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "e"}},
			Body:   LiteralExpr{Value: StringLiteral{Value: "transformed"}},
		},
		Env: env,
	}
	env.Define("fn", lambda)

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.mapErr",
		Args: []Expr{VariableExpr{Name: "fn"}},
	}, env)
	require.NoError(t, err)

	rv, ok := result.(*ResultValue)
	require.True(t, ok)
	assert.True(t, rv.IsErr())
	errVal, _ := rv.UnwrapErr()
	assert.Equal(t, "transformed", errVal)
}

func TestEvaluateResultMethod_MapErr_OkPassedThrough(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk(int64(42)))

	lambda := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "e"}},
			Body:   LiteralExpr{Value: StringLiteral{Value: "should not happen"}},
		},
		Env: env,
	}
	env.Define("fn", lambda)

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.mapErr",
		Args: []Expr{VariableExpr{Name: "fn"}},
	}, env)
	require.NoError(t, err)

	rv, ok := result.(*ResultValue)
	require.True(t, ok)
	assert.True(t, rv.IsOk())
}

func TestEvaluateResultMethod_MapErr_WrongArgCount(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("fail"))

	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.mapErr",
		Args: []Expr{},
	}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mapErr() expects 1 argument")
}

func TestEvaluateResultMethod_AndThen_ChainsOk(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk(int64(10)))

	// andThen with lambda returning a new Result
	lambda := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "v"}},
			Body: BinaryOpExpr{
				Op:    Add,
				Left:  VariableExpr{Name: "v"},
				Right: LiteralExpr{Value: IntLiteral{Value: 5}},
			},
		},
		Env: env,
	}
	env.Define("fn", lambda)

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.andThen",
		Args: []Expr{VariableExpr{Name: "fn"}},
	}, env)
	require.NoError(t, err)

	rv, ok := result.(*ResultValue)
	require.True(t, ok)
	assert.True(t, rv.IsOk())
	val, _ := rv.Unwrap()
	assert.Equal(t, int64(15), val)
}

func TestEvaluateResultMethod_AndThen_SkipsErr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("fail"))

	lambda := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "v"}},
			Body:   LiteralExpr{Value: IntLiteral{Value: 99}},
		},
		Env: env,
	}
	env.Define("fn", lambda)

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.andThen",
		Args: []Expr{VariableExpr{Name: "fn"}},
	}, env)
	require.NoError(t, err)

	rv, ok := result.(*ResultValue)
	require.True(t, ok)
	assert.True(t, rv.IsErr(), "andThen on Err should return Err")
}

func TestEvaluateResultMethod_AndThen_WrongArgCount(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk(int64(1)))

	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.andThen",
		Args: []Expr{},
	}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "andThen() expects 1 argument")
}

func TestEvaluateResultMethod_OrElse_RecoverFromErr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("fail"))

	lambda := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "e"}},
			Body:   LiteralExpr{Value: IntLiteral{Value: 42}},
		},
		Env: env,
	}
	env.Define("fn", lambda)

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.orElse",
		Args: []Expr{VariableExpr{Name: "fn"}},
	}, env)
	require.NoError(t, err)

	rv, ok := result.(*ResultValue)
	require.True(t, ok)
	assert.True(t, rv.IsOk())
	val, _ := rv.Unwrap()
	assert.Equal(t, int64(42), val)
}

func TestEvaluateResultMethod_OrElse_OkPassedThrough(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk(int64(7)))

	lambda := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "e"}},
			Body:   LiteralExpr{Value: IntLiteral{Value: 99}},
		},
		Env: env,
	}
	env.Define("fn", lambda)

	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.orElse",
		Args: []Expr{VariableExpr{Name: "fn"}},
	}, env)
	require.NoError(t, err)

	rv, ok := result.(*ResultValue)
	require.True(t, ok)
	assert.True(t, rv.IsOk())
	val, _ := rv.Unwrap()
	assert.Equal(t, int64(7), val)
}

func TestEvaluateResultMethod_OrElse_WrongArgCount(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewErr("fail"))

	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.orElse",
		Args: []Expr{},
	}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "orElse() expects 1 argument")
}

func TestEvaluateResultMethod_UnknownMethod(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("r", NewOk(int64(1)))

	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "r.nonExistentMethod",
		Args: []Expr{},
	}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Result has no method")
}

// =============================================================================
// callFnArg
// =============================================================================

func TestCallFnArg_WithFunction(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	fn := Function{
		Name: "double",
		Params: []Field{
			{Name: "x", TypeAnnotation: IntType{}, Required: true},
		},
		Body: []Statement{
			ReturnStatement{Value: BinaryOpExpr{
				Op:    Mul,
				Left:  VariableExpr{Name: "x"},
				Right: LiteralExpr{Value: IntLiteral{Value: 2}},
			}},
		},
	}

	result, err := interp.callFnArg(fn, int64(5), env)
	require.NoError(t, err)
	assert.Equal(t, int64(10), result)
}

func TestCallFnArg_WithFunctionPointer(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	fn := &Function{
		Name: "identity",
		Params: []Field{
			{Name: "x", Required: true},
		},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "x"}},
		},
	}

	result, err := interp.callFnArg(fn, "hello", env)
	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestCallFnArg_WithLambdaClosure(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	closure := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "v"}},
			Body: BinaryOpExpr{
				Op:    Add,
				Left:  VariableExpr{Name: "v"},
				Right: LiteralExpr{Value: IntLiteral{Value: 1}},
			},
		},
		Env: env,
	}

	result, err := interp.callFnArg(closure, int64(9), env)
	require.NoError(t, err)
	assert.Equal(t, int64(10), result)
}

func TestCallFnArg_WithLambdaClosureBlock(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	closure := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "v"}},
			Block: []Statement{
				ReturnStatement{Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "v"},
					Right: LiteralExpr{Value: IntLiteral{Value: 100}},
				}},
			},
		},
		Env: env,
	}

	result, err := interp.callFnArg(closure, int64(1), env)
	require.NoError(t, err)
	assert.Equal(t, int64(101), result)
}

func TestCallFnArg_WithLambdaClosureNoBody(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	closure := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "v"}},
			// No Body or Block
		},
		Env: env,
	}

	result, err := interp.callFnArg(closure, int64(1), env)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestCallFnArg_WithInvalidType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	_, err := interp.callFnArg("not a function", int64(1), env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected a function")
}

// =============================================================================
// TypeChecker: GetTypeBinding
// =============================================================================

func TestTypeChecker_GetTypeBinding(t *testing.T) {
	tc := NewTypeChecker()

	// No bindings initially
	_, ok := tc.GetTypeBinding("T")
	assert.False(t, ok)

	// Push a binding
	tc.PushTypeScope(map[string]Type{"T": IntType{}, "U": StringType{}})

	typ, ok := tc.GetTypeBinding("T")
	assert.True(t, ok)
	assert.IsType(t, IntType{}, typ)

	typ, ok = tc.GetTypeBinding("U")
	assert.True(t, ok)
	assert.IsType(t, StringType{}, typ)

	// Non-existent binding
	_, ok = tc.GetTypeBinding("V")
	assert.False(t, ok)

	// Pop the scope
	tc.PopTypeScope([]string{"T", "U"})
	_, ok = tc.GetTypeBinding("T")
	assert.False(t, ok)
}

// =============================================================================
// SubstituteTypeParams - additional coverage
// =============================================================================

func TestSubstituteTypeParams_Nil(t *testing.T) {
	tc := NewTypeChecker()
	assert.Nil(t, tc.SubstituteTypeParams(nil, map[string]Type{}))
}

func TestSubstituteTypeParams_TypeParam_Resolved(t *testing.T) {
	tc := NewTypeChecker()
	typeArgs := map[string]Type{"T": IntType{}}

	result := tc.SubstituteTypeParams(TypeParameterType{Name: "T"}, typeArgs)
	assert.IsType(t, IntType{}, result)
}

func TestSubstituteTypeParams_TypeParam_FromScope(t *testing.T) {
	tc := NewTypeChecker()
	tc.PushTypeScope(map[string]Type{"T": StringType{}})
	defer tc.PopTypeScope([]string{"T"})

	result := tc.SubstituteTypeParams(TypeParameterType{Name: "T"}, map[string]Type{})
	assert.IsType(t, StringType{}, result)
}

func TestSubstituteTypeParams_TypeParam_Unresolved(t *testing.T) {
	tc := NewTypeChecker()

	result := tc.SubstituteTypeParams(TypeParameterType{Name: "T"}, map[string]Type{})
	assert.IsType(t, TypeParameterType{}, result)
	assert.Equal(t, "T", result.(TypeParameterType).Name)
}

func TestSubstituteTypeParams_ArrayType(t *testing.T) {
	tc := NewTypeChecker()
	typeArgs := map[string]Type{"T": IntType{}}

	result := tc.SubstituteTypeParams(ArrayType{ElementType: TypeParameterType{Name: "T"}}, typeArgs)
	arrType, ok := result.(ArrayType)
	require.True(t, ok)
	assert.IsType(t, IntType{}, arrType.ElementType)
}

func TestSubstituteTypeParams_OptionalType(t *testing.T) {
	tc := NewTypeChecker()
	typeArgs := map[string]Type{"T": StringType{}}

	result := tc.SubstituteTypeParams(OptionalType{InnerType: TypeParameterType{Name: "T"}}, typeArgs)
	optType, ok := result.(OptionalType)
	require.True(t, ok)
	assert.IsType(t, StringType{}, optType.InnerType)
}

func TestSubstituteTypeParams_GenericType(t *testing.T) {
	tc := NewTypeChecker()
	typeArgs := map[string]Type{"T": IntType{}}

	input := GenericType{
		BaseType: NamedType{Name: "List"},
		TypeArgs: []Type{TypeParameterType{Name: "T"}},
	}

	result := tc.SubstituteTypeParams(input, typeArgs)
	genType, ok := result.(GenericType)
	require.True(t, ok)
	assert.Len(t, genType.TypeArgs, 1)
	assert.IsType(t, IntType{}, genType.TypeArgs[0])
}

func TestSubstituteTypeParams_FunctionType(t *testing.T) {
	tc := NewTypeChecker()
	typeArgs := map[string]Type{"T": IntType{}, "U": StringType{}}

	input := FunctionType{
		ParamTypes: []Type{TypeParameterType{Name: "T"}},
		ReturnType: TypeParameterType{Name: "U"},
	}

	result := tc.SubstituteTypeParams(input, typeArgs)
	fnType, ok := result.(FunctionType)
	require.True(t, ok)
	assert.IsType(t, IntType{}, fnType.ParamTypes[0])
	assert.IsType(t, StringType{}, fnType.ReturnType)
}

func TestSubstituteTypeParams_UnionType(t *testing.T) {
	tc := NewTypeChecker()
	typeArgs := map[string]Type{"T": IntType{}}

	input := UnionType{
		Types: []Type{TypeParameterType{Name: "T"}, StringType{}},
	}

	result := tc.SubstituteTypeParams(input, typeArgs)
	union, ok := result.(UnionType)
	require.True(t, ok)
	assert.Len(t, union.Types, 2)
	assert.IsType(t, IntType{}, union.Types[0])
	assert.IsType(t, StringType{}, union.Types[1])
}

func TestSubstituteTypeParams_PrimitiveUnchanged(t *testing.T) {
	tc := NewTypeChecker()
	typeArgs := map[string]Type{"T": IntType{}}

	assert.IsType(t, IntType{}, tc.SubstituteTypeParams(IntType{}, typeArgs))
	assert.IsType(t, StringType{}, tc.SubstituteTypeParams(StringType{}, typeArgs))
	assert.IsType(t, BoolType{}, tc.SubstituteTypeParams(BoolType{}, typeArgs))
	assert.IsType(t, FloatType{}, tc.SubstituteTypeParams(FloatType{}, typeArgs))
}

// =============================================================================
// inferFromValue - additional coverage
// =============================================================================

func TestInferTypeArguments_ArrayInference(t *testing.T) {
	tc := NewTypeChecker()

	fn := Function{
		Name: "first",
		TypeParams: []TypeParameter{
			{Name: "T"},
		},
		Params: []Field{
			{Name: "arr", TypeAnnotation: ArrayType{ElementType: TypeParameterType{Name: "T"}}, Required: true},
		},
		ReturnType: TypeParameterType{Name: "T"},
	}

	// Array of int64 values
	args := []interface{}{[]interface{}{int64(1), int64(2), int64(3)}}

	types, err := tc.InferTypeArguments(fn, args)
	require.NoError(t, err)
	assert.Len(t, types, 1)
	assert.IsType(t, IntType{}, types[0])
}

func TestInferFromValue_Nil(t *testing.T) {
	tc := NewTypeChecker()
	inferred := make(map[string]Type)

	// Nil paramType should be a no-op
	tc.inferFromValue(nil, int64(42), inferred)
	assert.Empty(t, inferred)
}

func TestInferFromValue_EmptyArray(t *testing.T) {
	tc := NewTypeChecker()
	inferred := make(map[string]Type)

	// Empty array - cannot infer element type
	tc.inferFromValue(ArrayType{ElementType: TypeParameterType{Name: "T"}}, []interface{}{}, inferred)
	_, found := inferred["T"]
	assert.False(t, found, "empty array should not infer type")
}

func TestInferFromValue_FunctionType(t *testing.T) {
	tc := NewTypeChecker()
	inferred := make(map[string]Type)

	// Function types cannot easily be inferred
	tc.inferFromValue(FunctionType{ParamTypes: []Type{TypeParameterType{Name: "T"}}}, "not a func", inferred)
	assert.Empty(t, inferred)
}

// =============================================================================
// executeValidation - additional edge cases
// =============================================================================

func TestExecuteValidation_NonBooleanNonErrorResult(t *testing.T) {
	interp := NewInterpreter()

	// Load a function that returns a non-bool, non-nil value via LoadModule
	err := interp.LoadModule(Module{
		Items: []Item{
			&Function{
				Name:   "myCheck",
				Params: []Field{},
				Body: []Statement{
					ReturnStatement{Value: LiteralExpr{Value: IntLiteral{Value: 42}}},
				},
			},
		},
	})
	require.NoError(t, err)

	// Use the global env (which has myCheck defined)
	stmt := ValidationStatement{
		Call: FunctionCallExpr{
			Name: "myCheck",
			Args: []Expr{},
		},
	}

	_, err = interp.ExecuteStatement(stmt, interp.globalEnv)
	assert.Error(t, err)
	assert.True(t, IsValidationError(err))
	assert.Contains(t, err.Error(), "unexpected type")
}

func TestExecuteValidation_NilResult(t *testing.T) {
	interp := NewInterpreter()

	// Load a function that returns nil via LoadModule
	err := interp.LoadModule(Module{
		Items: []Item{
			&Function{
				Name:   "nilCheck",
				Params: []Field{},
				Body: []Statement{
					ReturnStatement{Value: LiteralExpr{Value: NullLiteral{}}},
				},
			},
		},
	})
	require.NoError(t, err)

	stmt := ValidationStatement{
		Call: FunctionCallExpr{
			Name: "nilCheck",
			Args: []Expr{},
		},
	}

	_, err = interp.ExecuteStatement(stmt, interp.globalEnv)
	assert.NoError(t, err, "nil result should be treated as validation pass")
}

// =============================================================================
// convertValue - additional coverage for query params
// =============================================================================

func TestConvertValue_Bool(t *testing.T) {
	result, err := ProcessQueryParams(
		map[string][]string{"active": {"true"}},
		[]QueryParamDecl{{Name: "active", Type: BoolType{}}},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result["active"])
}

func TestConvertValue_BoolVariants(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"1", true},
		{"yes", true},
		{"on", true},
		{"false", false},
		{"0", false},
		{"no", false},
		{"off", false},
	}

	for _, tt := range tests {
		result, err := ProcessQueryParams(
			map[string][]string{"flag": {tt.input}},
			[]QueryParamDecl{{Name: "flag", Type: BoolType{}}},
		)
		require.NoError(t, err, "input: %s", tt.input)
		assert.Equal(t, tt.expected, result["flag"], "input: %s", tt.input)
	}
}

func TestConvertValue_InvalidBool(t *testing.T) {
	_, err := ProcessQueryParams(
		map[string][]string{"flag": {"maybe"}},
		[]QueryParamDecl{{Name: "flag", Type: BoolType{}}},
	)
	assert.Error(t, err)
}

func TestConvertValue_Float(t *testing.T) {
	result, err := ProcessQueryParams(
		map[string][]string{"price": {"9.99"}},
		[]QueryParamDecl{{Name: "price", Type: FloatType{}}},
	)
	require.NoError(t, err)
	assert.Equal(t, 9.99, result["price"])
}

func TestConvertValue_InvalidFloat(t *testing.T) {
	_, err := ProcessQueryParams(
		map[string][]string{"price": {"not-a-float"}},
		[]QueryParamDecl{{Name: "price", Type: FloatType{}}},
	)
	assert.Error(t, err)
}

func TestConvertValue_UnknownType(t *testing.T) {
	// Unknown type falls back to string
	result, err := ProcessQueryParams(
		map[string][]string{"data": {"value"}},
		[]QueryParamDecl{{Name: "data", Type: NamedType{Name: "Custom"}}},
	)
	require.NoError(t, err)
	assert.Equal(t, "value", result["data"])
}

func TestConvertValue_ArrayType(t *testing.T) {
	result, err := ProcessQueryParams(
		map[string][]string{"ids": {"1"}},
		[]QueryParamDecl{{Name: "ids", Type: ArrayType{ElementType: IntType{}}}},
	)
	require.NoError(t, err)
	// Single value for array type should be wrapped in a slice
	arr, ok := result["ids"].([]interface{})
	require.True(t, ok)
	assert.Len(t, arr, 1)
	assert.Equal(t, int64(1), arr[0])
}

// =============================================================================
// callWithPipedArg - additional coverage
// =============================================================================

func TestCallWithPipedArg_Function(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	fn := Function{
		Name: "addOne",
		Params: []Field{
			{Name: "x", TypeAnnotation: IntType{}, Required: true},
		},
		Body: []Statement{
			ReturnStatement{Value: BinaryOpExpr{
				Op:    Add,
				Left:  VariableExpr{Name: "x"},
				Right: LiteralExpr{Value: IntLiteral{Value: 1}},
			}},
		},
	}

	result, err := interp.callWithPipedArg(fn, int64(5), nil, env)
	require.NoError(t, err)
	assert.Equal(t, int64(6), result)
}

func TestCallWithPipedArg_LambdaClosure(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	closure := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "v"}},
			Body: BinaryOpExpr{
				Op:    Mul,
				Left:  VariableExpr{Name: "v"},
				Right: LiteralExpr{Value: IntLiteral{Value: 3}},
			},
		},
		Env: env,
	}

	result, err := interp.callWithPipedArg(closure, int64(4), nil, env)
	require.NoError(t, err)
	assert.Equal(t, int64(12), result)
}

func TestCallWithPipedArg_LambdaWithExtraArgs(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	closure := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "a"}, {Name: "b"}},
			Body: BinaryOpExpr{
				Op:    Add,
				Left:  VariableExpr{Name: "a"},
				Right: VariableExpr{Name: "b"},
			},
		},
		Env: env,
	}

	result, err := interp.callWithPipedArg(closure, int64(10),
		[]Expr{LiteralExpr{Value: IntLiteral{Value: 20}}}, env)
	require.NoError(t, err)
	assert.Equal(t, int64(30), result)
}

func TestCallWithPipedArg_NotAFunction(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	_, err := interp.callWithPipedArg("not a function", int64(1), nil, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pipe target is not a function")
}

// =============================================================================
// callLambdaClosure - additional coverage
// =============================================================================

func TestCallLambdaClosure_NoBody(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	closure := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "x"}},
			// No Body or Block
		},
		Env: env,
	}

	result, err := interp.callLambdaClosure(closure, []interface{}{int64(1)})
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestCallLambdaClosure_BlockBody(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	closure := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "x"}},
			Block: []Statement{
				ReturnStatement{Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "x"},
					Right: LiteralExpr{Value: IntLiteral{Value: 100}},
				}},
			},
		},
		Env: env,
	}

	result, err := interp.callLambdaClosure(closure, []interface{}{int64(5)})
	require.NoError(t, err)
	assert.Equal(t, int64(105), result)
}

func TestCallLambdaClosure_MissingRequiredArg(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	closure := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{{Name: "x", Required: true}},
			Body:   VariableExpr{Name: "x"},
		},
		Env: env,
	}

	_, err := interp.callLambdaClosure(closure, []interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

func TestCallLambdaClosure_OptionalParam(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	closure := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{
				{Name: "x", Required: true},
				{Name: "y"}, // optional
			},
			Body: VariableExpr{Name: "x"},
		},
		Env: env,
	}

	// Call with only the required arg
	result, err := interp.callLambdaClosure(closure, []interface{}{int64(42)})
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestCallLambdaClosure_DefaultParam(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	closure := &LambdaClosure{
		Lambda: LambdaExpr{
			Params: []Field{
				{Name: "x", Default: LiteralExpr{Value: IntLiteral{Value: 99}}},
			},
			Body: VariableExpr{Name: "x"},
		},
		Env: env,
	}

	// Call with no args, should use default
	result, err := interp.callLambdaClosure(closure, []interface{}{})
	require.NoError(t, err)
	assert.Equal(t, int64(99), result)
}

// =============================================================================
// Interpreter: LoadModule with various item types
// =============================================================================

func TestLoadModule_TraitDef(t *testing.T) {
	interp := NewInterpreter()

	err := interp.LoadModule(Module{
		Items: []Item{
			&TraitDef{
				Name: "Printable",
				Methods: []TraitMethodSignature{
					{Name: "toString"},
				},
			},
		},
	})
	require.NoError(t, err)

	defs := interp.GetTraitDefs()
	assert.Len(t, defs, 1)
	_, ok := defs["Printable"]
	assert.True(t, ok)
}

func TestLoadModule_UnsupportedItem(t *testing.T) {
	interp := NewInterpreter()

	// ExpressionStatement is a Statement, not an Item. Using something that
	// will trigger the default case requires a custom type.
	// We test the default by directly passing a route (which is handled).
	// Let's test the Module declaration path instead.
	err := interp.LoadModule(Module{
		Items: []Item{
			&ModuleDecl{Name: "mymod"},
		},
	})
	require.NoError(t, err, "ModuleDecl should be silently handled")
}

// =============================================================================
// LambdaExpr evaluation
// =============================================================================

func TestEvaluateLambdaExpr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	result, err := interp.EvaluateExpression(LambdaExpr{
		Params: []Field{{Name: "x"}},
		Body: BinaryOpExpr{
			Op:    Add,
			Left:  VariableExpr{Name: "x"},
			Right: LiteralExpr{Value: IntLiteral{Value: 1}},
		},
	}, env)
	require.NoError(t, err)

	closure, ok := result.(*LambdaClosure)
	require.True(t, ok)
	assert.Equal(t, "x", closure.Lambda.Params[0].Name)
}

// =============================================================================
// NullLiteral evaluation
// =============================================================================

func TestEvaluateNullLiteral(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	result, err := interp.EvaluateExpression(LiteralExpr{Value: NullLiteral{}}, env)
	require.NoError(t, err)
	assert.Nil(t, result)
}

// =============================================================================
// Unsupported expression type
// =============================================================================

func TestEvaluateExpression_Unsupported(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// QuoteExpr is a valid AST Expr that EvaluateExpression does not handle
	_, err := interp.EvaluateExpression(QuoteExpr{}, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported expression type")
}

// =============================================================================
// ExpressionStatement
// =============================================================================

func TestExpressionStatement_Evaluation(t *testing.T) {
	// ExpressionStatement wraps an expression as a statement
	stmt := ExpressionStatement{Expr: LiteralExpr{Value: IntLiteral{Value: 42}}}
	// Verify ExpressionStatement satisfies the Statement interface
	var _ Statement = stmt
	assert.NotNil(t, stmt.Expr)
}

// =============================================================================
// extractModuleName
// =============================================================================

func TestExtractModuleName_Additional(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"./utils", "utils"},
		{"./lib/helpers", "helpers"},
		{"./lib/helpers.glyph", "helpers"},
		{"module", "module"},
		{"path/to/mod.glyph", "mod"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractModuleName(tt.path))
		})
	}
}
