package interpreter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Environment

func TestEnvironment_DefineAndGet(t *testing.T) {
	env := NewEnvironment()

	env.Define("x", int64(42))
	val, err := env.Get("x")

	require.NoError(t, err)
	assert.Equal(t, int64(42), val)
}

func TestEnvironment_GetUndefinedVariable(t *testing.T) {
	env := NewEnvironment()

	_, err := env.Get("undefined")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undefined variable")
}

func TestEnvironment_ChildScope(t *testing.T) {
	parent := NewEnvironment()
	parent.Define("x", int64(10))

	child := NewChildEnvironment(parent)
	child.Define("y", int64(20))

	// Child can access parent variables
	val, err := child.Get("x")
	require.NoError(t, err)
	assert.Equal(t, int64(10), val)

	// Child can access its own variables
	val, err = child.Get("y")
	require.NoError(t, err)
	assert.Equal(t, int64(20), val)

	// Parent cannot access child variables
	_, err = parent.Get("y")
	assert.Error(t, err)
}

func TestEnvironment_Set(t *testing.T) {
	env := NewEnvironment()
	env.Define("x", int64(10))

	err := env.Set("x", int64(20))
	require.NoError(t, err)

	val, err := env.Get("x")
	require.NoError(t, err)
	assert.Equal(t, int64(20), val)
}

func TestEnvironment_SetUndefined(t *testing.T) {
	env := NewEnvironment()

	err := env.Set("undefined", int64(10))
	assert.Error(t, err)
}

// Test Expression Evaluation

func TestEvaluateLiteral_Int(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := LiteralExpr{Value: IntLiteral{Value: 42}}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestEvaluateLiteral_String(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := LiteralExpr{Value: StringLiteral{Value: "hello"}}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestEvaluateLiteral_Bool(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := LiteralExpr{Value: BoolLiteral{Value: true}}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateLiteral_Float(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := LiteralExpr{Value: FloatLiteral{Value: 3.14}}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, 3.14, result)
}

func TestEvaluateVariable(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("x", int64(42))

	expr := VariableExpr{Name: "x"}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestEvaluateBinaryOp_Add_Int(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Add,
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Right: LiteralExpr{Value: IntLiteral{Value: 20}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, int64(30), result)
}

func TestEvaluateBinaryOp_Add_String(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Add,
		Left:  LiteralExpr{Value: StringLiteral{Value: "Hello, "}},
		Right: LiteralExpr{Value: StringLiteral{Value: "World!"}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", result)
}

func TestEvaluateBinaryOp_Sub(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Sub,
		Left:  LiteralExpr{Value: IntLiteral{Value: 20}},
		Right: LiteralExpr{Value: IntLiteral{Value: 10}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, int64(10), result)
}

func TestEvaluateBinaryOp_Mul(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Mul,
		Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
		Right: LiteralExpr{Value: IntLiteral{Value: 6}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, int64(30), result)
}

func TestEvaluateBinaryOp_Div(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Div,
		Left:  LiteralExpr{Value: IntLiteral{Value: 20}},
		Right: LiteralExpr{Value: IntLiteral{Value: 4}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, int64(5), result)
}

func TestEvaluateBinaryOp_Div_ByZero(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Div,
		Left:  LiteralExpr{Value: IntLiteral{Value: 20}},
		Right: LiteralExpr{Value: IntLiteral{Value: 0}},
	}
	_, err := interp.EvaluateExpression(expr, env)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "division by zero")
}

func TestEvaluateBinaryOp_Eq(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Eq,
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Right: LiteralExpr{Value: IntLiteral{Value: 10}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateBinaryOp_Ne(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Ne,
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Right: LiteralExpr{Value: IntLiteral{Value: 20}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateBinaryOp_Lt(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Lt,
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Right: LiteralExpr{Value: IntLiteral{Value: 20}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateBinaryOp_Gt(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Gt,
		Left:  LiteralExpr{Value: IntLiteral{Value: 20}},
		Right: LiteralExpr{Value: IntLiteral{Value: 10}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateFieldAccess(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	obj := map[string]interface{}{
		"name": "John",
		"age":  int64(30),
	}
	env.Define("person", obj)

	expr := FieldAccessExpr{
		Object: VariableExpr{Name: "person"},
		Field:  "name",
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, "John", result)
}

func TestEvaluateFunctionCall_TimeNow(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := FunctionCallExpr{
		Name: "time.now",
		Args: []Expr{},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, int64(1234567890), result)
}

func TestEvaluateObject_Empty(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := ObjectExpr{
		Fields: []ObjectField{},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	obj, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, obj, 0)
}

func TestEvaluateObject_WithLiterals(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := ObjectExpr{
		Fields: []ObjectField{
			{Key: "text", Value: LiteralExpr{Value: StringLiteral{Value: "Hello, World!"}}},
			{Key: "timestamp", Value: LiteralExpr{Value: IntLiteral{Value: 1234567890}}},
		},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	obj, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Hello, World!", obj["text"])
	assert.Equal(t, int64(1234567890), obj["timestamp"])
}

func TestEvaluateObject_WithVariables(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("name", "Alice")
	env.Define("age", int64(30))

	expr := ObjectExpr{
		Fields: []ObjectField{
			{Key: "name", Value: VariableExpr{Name: "name"}},
			{Key: "age", Value: VariableExpr{Name: "age"}},
		},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	obj, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Alice", obj["name"])
	assert.Equal(t, int64(30), obj["age"])
}

func TestEvaluateObject_WithExpressions(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("name", "Alice")

	expr := ObjectExpr{
		Fields: []ObjectField{
			{
				Key: "text",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  LiteralExpr{Value: StringLiteral{Value: "Hello, "}},
					Right: VariableExpr{Name: "name"},
				},
			},
			{
				Key: "count",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
					Right: LiteralExpr{Value: IntLiteral{Value: 5}},
				},
			},
		},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	obj, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Hello, Alice", obj["text"])
	assert.Equal(t, int64(15), obj["count"])
}

func TestEvaluateObject_Nested(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := ObjectExpr{
		Fields: []ObjectField{
			{Key: "name", Value: LiteralExpr{Value: StringLiteral{Value: "Alice"}}},
			{
				Key: "address",
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "city", Value: LiteralExpr{Value: StringLiteral{Value: "NYC"}}},
						{Key: "zip", Value: LiteralExpr{Value: IntLiteral{Value: 10001}}},
					},
				},
			},
		},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	obj, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Alice", obj["name"])

	address, ok := obj["address"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "NYC", address["city"])
	assert.Equal(t, int64(10001), address["zip"])
}

// Test Statement Execution

func TestExecuteAssign(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	stmt := AssignStatement{
		Target: "x",
		Value:  LiteralExpr{Value: IntLiteral{Value: 42}},
	}
	result, err := interp.ExecuteStatement(stmt, env)

	require.NoError(t, err)
	assert.Equal(t, int64(42), result)

	val, err := env.Get("x")
	require.NoError(t, err)
	assert.Equal(t, int64(42), val)
}

func TestExecuteAssign_RedeclarationError(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// First assignment should succeed
	stmt1 := AssignStatement{
		Target: "x",
		Value:  LiteralExpr{Value: IntLiteral{Value: 1}},
	}
	_, err := interp.ExecuteStatement(stmt1, env)
	require.NoError(t, err)

	// Second assignment to same variable in same scope should fail
	stmt2 := AssignStatement{
		Target: "x",
		Value:  LiteralExpr{Value: IntLiteral{Value: 2}},
	}
	_, err = interp.ExecuteStatement(stmt2, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot redeclare variable 'x' in the same scope")
}

func TestExecuteAssign_UpdateParentScope(t *testing.T) {
	interp := NewInterpreter()
	parentEnv := NewEnvironment()

	// Define x in parent scope
	parentEnv.Define("x", int64(0))

	// Create child scope
	childEnv := NewChildEnvironment(parentEnv)

	// Assignment in child scope should update parent's x (not create new)
	stmt := AssignStatement{
		Target: "x",
		Value:  LiteralExpr{Value: IntLiteral{Value: 42}},
	}
	_, err := interp.ExecuteStatement(stmt, childEnv)
	require.NoError(t, err)

	// Verify parent's x was updated
	val, err := parentEnv.Get("x")
	require.NoError(t, err)
	assert.Equal(t, int64(42), val)

	// Verify x is not defined locally in child
	assert.False(t, childEnv.HasLocal("x"))
}

func TestExecuteReturn(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	stmt := ReturnStatement{
		Value: LiteralExpr{Value: IntLiteral{Value: 42}},
	}
	result, err := interp.ExecuteStatement(stmt, env)

	require.Error(t, err)
	assert.IsType(t, &returnValue{}, err)
	assert.Equal(t, int64(42), result)
}

func TestExecuteIf_TrueCondition(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	stmt := IfStatement{
		Condition: LiteralExpr{Value: BoolLiteral{Value: true}},
		ThenBlock: []Statement{
			AssignStatement{
				Target: "result",
				Value:  LiteralExpr{Value: StringLiteral{Value: "then"}},
			},
		},
		ElseBlock: []Statement{
			AssignStatement{
				Target: "result",
				Value:  LiteralExpr{Value: StringLiteral{Value: "else"}},
			},
		},
	}
	_, err := interp.ExecuteStatement(stmt, env)

	require.NoError(t, err)

	// Note: result is in child scope, so we can't access it here
	// This tests that the if statement executes without error
}

func TestExecuteIf_FalseCondition(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	stmt := IfStatement{
		Condition: LiteralExpr{Value: BoolLiteral{Value: false}},
		ThenBlock: []Statement{
			AssignStatement{
				Target: "result",
				Value:  LiteralExpr{Value: StringLiteral{Value: "then"}},
			},
		},
		ElseBlock: []Statement{
			AssignStatement{
				Target: "result",
				Value:  LiteralExpr{Value: StringLiteral{Value: "else"}},
			},
		},
	}
	_, err := interp.ExecuteStatement(stmt, env)

	require.NoError(t, err)
}

// Test Route Execution

func TestExecuteRoute_Simple(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/hello",
		Method: Get,
		Body: []Statement{
			ReturnStatement{
				Value: LiteralExpr{
					Value: StringLiteral{Value: "Hello, World!"},
				},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)

	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", result)
}

func TestExecuteRoute_WithPathParam(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/greet/:name",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "greeting",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  LiteralExpr{Value: StringLiteral{Value: "Hello, "}},
					Right: VariableExpr{Name: "name"},
				},
			},
			ReturnStatement{
				Value: VariableExpr{Name: "greeting"},
			},
		},
	}

	params := map[string]string{
		"name": "Alice",
	}
	result, err := interp.ExecuteRouteSimple(route, params)

	require.NoError(t, err)
	assert.Equal(t, "Hello, Alice", result)
}

func TestExecuteRoute_WithMultipleStatements(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/calculate",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "a",
				Value:  LiteralExpr{Value: IntLiteral{Value: 10}},
			},
			AssignStatement{
				Target: "b",
				Value:  LiteralExpr{Value: IntLiteral{Value: 20}},
			},
			AssignStatement{
				Target: "sum",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "a"},
					Right: VariableExpr{Name: "b"},
				},
			},
			ReturnStatement{
				Value: VariableExpr{Name: "sum"},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)

	require.NoError(t, err)
	assert.Equal(t, int64(30), result)
}

func TestExecuteRoute_ComplexExpression(t *testing.T) {
	interp := NewInterpreter()

	// (10 + 20) * 2
	route := &Route{
		Path:   "/complex",
		Method: Get,
		Body: []Statement{
			ReturnStatement{
				Value: BinaryOpExpr{
					Op: Mul,
					Left: BinaryOpExpr{
						Op:    Add,
						Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
						Right: LiteralExpr{Value: IntLiteral{Value: 20}},
					},
					Right: LiteralExpr{Value: IntLiteral{Value: 2}},
				},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)

	require.NoError(t, err)
	assert.Equal(t, int64(60), result)
}

// Test Module Loading

func TestLoadModule_Function(t *testing.T) {
	interp := NewInterpreter()

	fn := Function{
		Name: "add",
		Params: []Field{
			{Name: "a", TypeAnnotation: IntType{}, Required: true},
			{Name: "b", TypeAnnotation: IntType{}, Required: true},
		},
		ReturnType: IntType{},
		Body: []Statement{
			ReturnStatement{
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "a"},
					Right: VariableExpr{Name: "b"},
				},
			},
		},
	}

	module := Module{
		Items: []Item{&fn},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	// Verify function was loaded
	loadedFn, ok := interp.GetFunction("add")
	assert.True(t, ok)
	assert.Equal(t, "add", loadedFn.Name)
}

func TestLoadModule_TypeDef(t *testing.T) {
	interp := NewInterpreter()

	typeDef := TypeDef{
		Name: "User",
		Fields: []Field{
			{Name: "id", TypeAnnotation: IntType{}, Required: true},
			{Name: "name", TypeAnnotation: StringType{}, Required: true},
		},
	}

	module := Module{
		Items: []Item{&typeDef},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	// Verify type was loaded
	loadedType, ok := interp.GetTypeDef("User")
	assert.True(t, ok)
	assert.Equal(t, "User", loadedType.Name)
	assert.Len(t, loadedType.Fields, 2)
}

// Test User-Defined Function Execution

func TestExecuteUserDefinedFunction(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Define a simple add function
	fn := Function{
		Name: "add",
		Params: []Field{
			{Name: "a", TypeAnnotation: IntType{}, Required: true},
			{Name: "b", TypeAnnotation: IntType{}, Required: true},
		},
		ReturnType: IntType{},
		Body: []Statement{
			ReturnStatement{
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "a"},
					Right: VariableExpr{Name: "b"},
				},
			},
		},
	}

	env.Define("add", fn)

	// Call the function
	expr := FunctionCallExpr{
		Name: "add",
		Args: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 5}},
			LiteralExpr{Value: IntLiteral{Value: 7}},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, int64(12), result)
}

// Test Path Parameter Extraction

func TestExtractPathParams_Simple(t *testing.T) {
	params, err := extractPathParams("/greet/:name", "/greet/Alice")

	require.NoError(t, err)
	assert.Equal(t, "Alice", params["name"])
}

func TestExtractPathParams_Multiple(t *testing.T) {
	params, err := extractPathParams("/users/:id/posts/:postId", "/users/123/posts/456")

	require.NoError(t, err)
	assert.Equal(t, "123", params["id"])
	assert.Equal(t, "456", params["postId"])
}

func TestExtractPathParams_Mismatch(t *testing.T) {
	_, err := extractPathParams("/greet/:name", "/hello/Alice")

	assert.Error(t, err)
}

// Test Complete Hello World with Object Literals

func TestExecuteRoute_HelloWorldWithObject(t *testing.T) {
	interp := NewInterpreter()

	// @ GET /hello
	//   > {text: "Hello, World!", timestamp: 1234567890}
	route := &Route{
		Path:   "/hello",
		Method: Get,
		Body: []Statement{
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "text", Value: LiteralExpr{Value: StringLiteral{Value: "Hello, World!"}}},
						{Key: "timestamp", Value: LiteralExpr{Value: IntLiteral{Value: 1234567890}}},
					},
				},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)

	require.NoError(t, err)
	obj, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Hello, World!", obj["text"])
	assert.Equal(t, int64(1234567890), obj["timestamp"])
}

func TestExecuteRoute_GreetWithObjectAndParams(t *testing.T) {
	interp := NewInterpreter()

	// @ GET /greet/:name -> Message
	//   $ message = {
	//     text: "Hello, " + name + "!",
	//     timestamp: time.now()
	//   }
	//   > message
	route := &Route{
		Path:   "/greet/:name",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "message",
				Value: ObjectExpr{
					Fields: []ObjectField{
						{
							Key: "text",
							Value: BinaryOpExpr{
								Op: Add,
								Left: BinaryOpExpr{
									Op:    Add,
									Left:  LiteralExpr{Value: StringLiteral{Value: "Hello, "}},
									Right: VariableExpr{Name: "name"},
								},
								Right: LiteralExpr{Value: StringLiteral{Value: "!"}},
							},
						},
						{
							Key:   "timestamp",
							Value: FunctionCallExpr{Name: "time.now", Args: []Expr{}},
						},
					},
				},
			},
			ReturnStatement{
				Value: VariableExpr{Name: "message"},
			},
		},
	}

	params := map[string]string{"name": "World"}
	result, err := interp.ExecuteRouteSimple(route, params)

	require.NoError(t, err)
	obj, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Hello, World!", obj["text"])
	assert.Equal(t, int64(1234567890), obj["timestamp"])
}

func TestExecuteRoute_ObjectFieldAccess(t *testing.T) {
	interp := NewInterpreter()

	// @ GET /user-name
	//   $ user = {name: "Alice", age: 30}
	//   $ username = user.name
	//   > username
	route := &Route{
		Path:   "/user-name",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "user",
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "name", Value: LiteralExpr{Value: StringLiteral{Value: "Alice"}}},
						{Key: "age", Value: LiteralExpr{Value: IntLiteral{Value: 30}}},
					},
				},
			},
			AssignStatement{
				Target: "username",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "user"},
					Field:  "name",
				},
			},
			ReturnStatement{
				Value: VariableExpr{Name: "username"},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)

	require.NoError(t, err)
	assert.Equal(t, "Alice", result)
}

func TestExecuteRoute_ComplexObjectWithFieldAccess(t *testing.T) {
	interp := NewInterpreter()

	// @ GET /user-info
	//   $ user = {
	//     name: "Alice",
	//     profile: {
	//       bio: "Developer",
	//       score: 100
	//     }
	//   }
	//   $ bio = user.profile.bio
	//   > bio
	route := &Route{
		Path:   "/user-info",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "user",
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "name", Value: LiteralExpr{Value: StringLiteral{Value: "Alice"}}},
						{
							Key: "profile",
							Value: ObjectExpr{
								Fields: []ObjectField{
									{Key: "bio", Value: LiteralExpr{Value: StringLiteral{Value: "Developer"}}},
									{Key: "score", Value: LiteralExpr{Value: IntLiteral{Value: 100}}},
								},
							},
						},
					},
				},
			},
			AssignStatement{
				Target: "profile",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "user"},
					Field:  "profile",
				},
			},
			AssignStatement{
				Target: "bio",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "profile"},
					Field:  "bio",
				},
			},
			ReturnStatement{
				Value: VariableExpr{Name: "bio"},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)

	require.NoError(t, err)
	assert.Equal(t, "Developer", result)
}

// Test Logical Operators

func TestEvaluateBinaryOp_And_True(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    And,
		Left:  LiteralExpr{Value: BoolLiteral{Value: true}},
		Right: LiteralExpr{Value: BoolLiteral{Value: true}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateBinaryOp_And_False(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    And,
		Left:  LiteralExpr{Value: BoolLiteral{Value: true}},
		Right: LiteralExpr{Value: BoolLiteral{Value: false}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestEvaluateBinaryOp_Or_True(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Or,
		Left:  LiteralExpr{Value: BoolLiteral{Value: true}},
		Right: LiteralExpr{Value: BoolLiteral{Value: false}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateBinaryOp_Or_False(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Op:    Or,
		Left:  LiteralExpr{Value: BoolLiteral{Value: false}},
		Right: LiteralExpr{Value: BoolLiteral{Value: false}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestEvaluateBinaryOp_ComplexLogical(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// (true && false) || true
	expr := BinaryOpExpr{
		Op: Or,
		Left: BinaryOpExpr{
			Op:    And,
			Left:  LiteralExpr{Value: BoolLiteral{Value: true}},
			Right: LiteralExpr{Value: BoolLiteral{Value: false}},
		},
		Right: LiteralExpr{Value: BoolLiteral{Value: true}},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

// Test Array Literals

func TestEvaluateArray_Empty(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := ArrayExpr{
		Elements: []Expr{},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	arr, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Len(t, arr, 0)
}

func TestEvaluateArray_Integers(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := ArrayExpr{
		Elements: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 1}},
			LiteralExpr{Value: IntLiteral{Value: 2}},
			LiteralExpr{Value: IntLiteral{Value: 3}},
		},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	arr, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Len(t, arr, 3)
	assert.Equal(t, int64(1), arr[0])
	assert.Equal(t, int64(2), arr[1])
	assert.Equal(t, int64(3), arr[2])
}

func TestEvaluateArray_Strings(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := ArrayExpr{
		Elements: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "Alice"}},
			LiteralExpr{Value: StringLiteral{Value: "Bob"}},
			LiteralExpr{Value: StringLiteral{Value: "Charlie"}},
		},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	arr, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Len(t, arr, 3)
	assert.Equal(t, "Alice", arr[0])
	assert.Equal(t, "Bob", arr[1])
	assert.Equal(t, "Charlie", arr[2])
}

func TestEvaluateArray_Mixed(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := ArrayExpr{
		Elements: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 42}},
			LiteralExpr{Value: StringLiteral{Value: "test"}},
			LiteralExpr{Value: BoolLiteral{Value: true}},
			LiteralExpr{Value: FloatLiteral{Value: 3.14}},
		},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	arr, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Len(t, arr, 4)
	assert.Equal(t, int64(42), arr[0])
	assert.Equal(t, "test", arr[1])
	assert.Equal(t, true, arr[2])
	assert.Equal(t, 3.14, arr[3])
}

func TestEvaluateArray_WithExpressions(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("x", int64(10))

	expr := ArrayExpr{
		Elements: []Expr{
			BinaryOpExpr{
				Op:    Add,
				Left:  LiteralExpr{Value: IntLiteral{Value: 1}},
				Right: LiteralExpr{Value: IntLiteral{Value: 1}},
			},
			BinaryOpExpr{
				Op:    Mul,
				Left:  VariableExpr{Name: "x"},
				Right: LiteralExpr{Value: IntLiteral{Value: 2}},
			},
		},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	arr, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Len(t, arr, 2)
	assert.Equal(t, int64(2), arr[0])
	assert.Equal(t, int64(20), arr[1])
}

func TestEvaluateArray_Nested(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := ArrayExpr{
		Elements: []Expr{
			ArrayExpr{
				Elements: []Expr{
					LiteralExpr{Value: IntLiteral{Value: 1}},
					LiteralExpr{Value: IntLiteral{Value: 2}},
				},
			},
			ArrayExpr{
				Elements: []Expr{
					LiteralExpr{Value: IntLiteral{Value: 3}},
					LiteralExpr{Value: IntLiteral{Value: 4}},
				},
			},
		},
	}
	result, err := interp.EvaluateExpression(expr, env)

	require.NoError(t, err)
	arr, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Len(t, arr, 2)

	arr1, ok := arr[0].([]interface{})
	assert.True(t, ok)
	assert.Len(t, arr1, 2)
	assert.Equal(t, int64(1), arr1[0])
	assert.Equal(t, int64(2), arr1[1])

	arr2, ok := arr[1].([]interface{})
	assert.True(t, ok)
	assert.Len(t, arr2, 2)
	assert.Equal(t, int64(3), arr2[0])
	assert.Equal(t, int64(4), arr2[1])
}

func TestExecuteRoute_WithIfElse(t *testing.T) {
	interp := NewInterpreter()

	// @ GET /check
	//   $ age = 20
	//   if age >= 18 {
	//     > {status: "adult"}
	//   } else {
	//     > {status: "minor"}
	//   }
	route := &Route{
		Path:   "/check",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "age",
				Value:  LiteralExpr{Value: IntLiteral{Value: 20}},
			},
			IfStatement{
				Condition: BinaryOpExpr{
					Op:    Ge,
					Left:  VariableExpr{Name: "age"},
					Right: LiteralExpr{Value: IntLiteral{Value: 18}},
				},
				ThenBlock: []Statement{
					ReturnStatement{
						Value: ObjectExpr{
							Fields: []ObjectField{
								{Key: "status", Value: LiteralExpr{Value: StringLiteral{Value: "adult"}}},
							},
						},
					},
				},
				ElseBlock: []Statement{
					ReturnStatement{
						Value: ObjectExpr{
							Fields: []ObjectField{
								{Key: "status", Value: LiteralExpr{Value: StringLiteral{Value: "minor"}}},
							},
						},
					},
				},
			},
		},
	}

	// Test with age >= 18
	result, err := interp.ExecuteRouteSimple(route, nil)
	require.NoError(t, err)

	obj, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "adult", obj["status"])

	// Test with age < 18
	route2 := &Route{
		Path:   "/check2",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "age",
				Value:  LiteralExpr{Value: IntLiteral{Value: 15}},
			},
			IfStatement{
				Condition: BinaryOpExpr{
					Op:    Ge,
					Left:  VariableExpr{Name: "age"},
					Right: LiteralExpr{Value: IntLiteral{Value: 18}},
				},
				ThenBlock: []Statement{
					ReturnStatement{
						Value: ObjectExpr{
							Fields: []ObjectField{
								{Key: "status", Value: LiteralExpr{Value: StringLiteral{Value: "adult"}}},
							},
						},
					},
				},
				ElseBlock: []Statement{
					ReturnStatement{
						Value: ObjectExpr{
							Fields: []ObjectField{
								{Key: "status", Value: LiteralExpr{Value: StringLiteral{Value: "minor"}}},
							},
						},
					},
				},
			},
		},
	}

	result, err = interp.ExecuteRouteSimple(route2, nil)
	require.NoError(t, err)

	obj, ok = result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "minor", obj["status"])
}

func TestExecuteRoute_ArrayInRoute(t *testing.T) {
	interp := NewInterpreter()

	// @ GET /numbers
	//   $ nums = [1, 2, 3, 4, 5]
	//   > {numbers: nums}
	route := &Route{
		Path:   "/numbers",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "nums",
				Value: ArrayExpr{
					Elements: []Expr{
						LiteralExpr{Value: IntLiteral{Value: 1}},
						LiteralExpr{Value: IntLiteral{Value: 2}},
						LiteralExpr{Value: IntLiteral{Value: 3}},
						LiteralExpr{Value: IntLiteral{Value: 4}},
						LiteralExpr{Value: IntLiteral{Value: 5}},
					},
				},
			},
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "numbers", Value: VariableExpr{Name: "nums"}},
					},
				},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)

	require.NoError(t, err)
	obj, ok := result.(map[string]interface{})
	assert.True(t, ok)

	nums, ok := obj["numbers"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, nums, 5)
	assert.Equal(t, int64(1), nums[0])
	assert.Equal(t, int64(5), nums[4])
}

// ========================================
// Additional Coverage Tests - Arithmetic Operators
// ========================================

func TestEvaluateSub_Float(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: FloatLiteral{Value: 10.5}},
		Op:    Sub,
		Right: LiteralExpr{Value: FloatLiteral{Value: 3.2}},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.InDelta(t, 7.3, result.(float64), 0.001)
}

func TestEvaluateSub_TypeMismatch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Op:    Sub,
		Right: LiteralExpr{Value: FloatLiteral{Value: 3.2}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot subtract")
}

func TestEvaluateSub_UnsupportedType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: StringLiteral{Value: "hello"}},
		Op:    Sub,
		Right: LiteralExpr{Value: StringLiteral{Value: "world"}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}

func TestEvaluateMul_Float(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: FloatLiteral{Value: 2.5}},
		Op:    Mul,
		Right: LiteralExpr{Value: FloatLiteral{Value: 4.0}},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.InDelta(t, 10.0, result.(float64), 0.001)
}

func TestEvaluateMul_TypeMismatch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
		Op:    Mul,
		Right: LiteralExpr{Value: FloatLiteral{Value: 2.0}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}

func TestEvaluateMul_UnsupportedType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: BoolLiteral{Value: true}},
		Op:    Mul,
		Right: LiteralExpr{Value: IntLiteral{Value: 5}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}

func TestEvaluateDiv_Float(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: FloatLiteral{Value: 10.0}},
		Op:    Div,
		Right: LiteralExpr{Value: FloatLiteral{Value: 2.5}},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.InDelta(t, 4.0, result.(float64), 0.001)
}

func TestEvaluateDiv_FloatByZero(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: FloatLiteral{Value: 10.0}},
		Op:    Div,
		Right: LiteralExpr{Value: FloatLiteral{Value: 0.0}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "division by zero")
}

func TestEvaluateDiv_TypeMismatch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Op:    Div,
		Right: LiteralExpr{Value: FloatLiteral{Value: 2.0}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}

// Comparison operators with floats

func TestEvaluateLt_Float(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: FloatLiteral{Value: 3.5}},
		Op:    Lt,
		Right: LiteralExpr{Value: FloatLiteral{Value: 4.2}},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateLt_TypeMismatch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
		Op:    Lt,
		Right: LiteralExpr{Value: FloatLiteral{Value: 3.0}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}

func TestEvaluateLt_UnsupportedType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: BoolLiteral{Value: true}},
		Op:    Lt,
		Right: LiteralExpr{Value: BoolLiteral{Value: false}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}

func TestEvaluateLe_Float(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: FloatLiteral{Value: 4.2}},
		Op:    Le,
		Right: LiteralExpr{Value: FloatLiteral{Value: 4.2}},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateLe_TypeMismatch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: FloatLiteral{Value: 5.0}},
		Op:    Le,
		Right: LiteralExpr{Value: IntLiteral{Value: 5}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}

func TestEvaluateGt_Float(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: FloatLiteral{Value: 5.5}},
		Op:    Gt,
		Right: LiteralExpr{Value: FloatLiteral{Value: 3.3}},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateGt_TypeMismatch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Op:    Gt,
		Right: LiteralExpr{Value: StringLiteral{Value: "5"}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}

func TestEvaluateGe_Float(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: FloatLiteral{Value: 5.0}},
		Op:    Ge,
		Right: LiteralExpr{Value: FloatLiteral{Value: 5.0}},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateGe_TypeMismatch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: StringLiteral{Value: "hello"}},
		Op:    Ge,
		Right: LiteralExpr{Value: IntLiteral{Value: 5}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}

// Logical operators edge cases

func TestEvaluateAnd_NonBoolLeft(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
		Op:    And,
		Right: LiteralExpr{Value: BoolLiteral{Value: true}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires boolean operands")
}

func TestEvaluateAnd_NonBoolRight(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: BoolLiteral{Value: true}},
		Op:    And,
		Right: LiteralExpr{Value: StringLiteral{Value: "not a bool"}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}

func TestEvaluateOr_NonBoolLeft(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: FloatLiteral{Value: 3.14}},
		Op:    Or,
		Right: LiteralExpr{Value: BoolLiteral{Value: false}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}

func TestEvaluateOr_NonBoolRight(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := BinaryOpExpr{
		Left:  LiteralExpr{Value: BoolLiteral{Value: false}},
		Op:    Or,
		Right: LiteralExpr{Value: IntLiteral{Value: 42}},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
}

// Array and object edge cases

func TestEvaluateArray_MixedTypes(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := ArrayExpr{
		Elements: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 1}},
			LiteralExpr{Value: StringLiteral{Value: "two"}},
			LiteralExpr{Value: BoolLiteral{Value: true}},
			LiteralExpr{Value: FloatLiteral{Value: 4.0}},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)

	arr, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Len(t, arr, 4)
	assert.Equal(t, int64(1), arr[0])
	assert.Equal(t, "two", arr[1])
	assert.Equal(t, true, arr[2])
	assert.Equal(t, 4.0, arr[3])
}

func TestEvaluateFieldAccess_NonObject(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("x", int64(42))

	expr := FieldAccessExpr{
		Object: VariableExpr{Name: "x"},
		Field:  "value",
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot access field")
}

func TestEvaluateFieldAccess_MissingField(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("obj", map[string]interface{}{
		"name": "Alice",
	})

	expr := FieldAccessExpr{
		Object: VariableExpr{Name: "obj"},
		Field:  "age",
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field age not found")
}

func TestEvaluateFunctionCall_NonCallable(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("notAFunction", int64(42))

	expr := FunctionCallExpr{
		Name: "notAFunction",
		Args: []Expr{},
	}

	_, err := interp.EvaluateExpression(expr, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a function")
}

// Test While Loops

func TestExecuteWhile_SimpleLoop(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Initialize counter
	env.Define("i", int64(0))

	// Create while loop: while i < 5 { i = i + 1 }
	whileStmt := WhileStatement{
		Condition: BinaryOpExpr{
			Op:    Lt,
			Left:  VariableExpr{Name: "i"},
			Right: LiteralExpr{Value: IntLiteral{Value: 5}},
		},
		Body: []Statement{
			AssignStatement{
				Target: "i",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "i"},
					Right: LiteralExpr{Value: IntLiteral{Value: 1}},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(whileStmt, env)
	require.NoError(t, err)

	// Check final value
	val, err := env.Get("i")
	require.NoError(t, err)
	assert.Equal(t, int64(5), val)
}

func TestExecuteWhile_SumLoop(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Initialize variables
	env.Define("sum", int64(0))
	env.Define("i", int64(0))

	// Create while loop: while i < 5 { sum = sum + i; i = i + 1 }
	whileStmt := WhileStatement{
		Condition: BinaryOpExpr{
			Op:    Lt,
			Left:  VariableExpr{Name: "i"},
			Right: LiteralExpr{Value: IntLiteral{Value: 5}},
		},
		Body: []Statement{
			AssignStatement{
				Target: "sum",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "sum"},
					Right: VariableExpr{Name: "i"},
				},
			},
			AssignStatement{
				Target: "i",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "i"},
					Right: LiteralExpr{Value: IntLiteral{Value: 1}},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(whileStmt, env)
	require.NoError(t, err)

	// Check final values: sum should be 0+1+2+3+4 = 10
	sum, err := env.Get("sum")
	require.NoError(t, err)
	assert.Equal(t, int64(10), sum)

	i, err := env.Get("i")
	require.NoError(t, err)
	assert.Equal(t, int64(5), i)
}

func TestExecuteWhile_NestedLoops(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Initialize variables
	env.Define("count", int64(0))
	env.Define("i", int64(0))

	// Create nested while loops
	innerWhile := WhileStatement{
		Condition: BinaryOpExpr{
			Op:    Lt,
			Left:  VariableExpr{Name: "j"},
			Right: LiteralExpr{Value: IntLiteral{Value: 2}},
		},
		Body: []Statement{
			AssignStatement{
				Target: "count",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "count"},
					Right: LiteralExpr{Value: IntLiteral{Value: 1}},
				},
			},
			AssignStatement{
				Target: "j",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "j"},
					Right: LiteralExpr{Value: IntLiteral{Value: 1}},
				},
			},
		},
	}

	outerWhile := WhileStatement{
		Condition: BinaryOpExpr{
			Op:    Lt,
			Left:  VariableExpr{Name: "i"},
			Right: LiteralExpr{Value: IntLiteral{Value: 3}},
		},
		Body: []Statement{
			AssignStatement{
				Target: "j",
				Value:  LiteralExpr{Value: IntLiteral{Value: 0}},
			},
			innerWhile,
			AssignStatement{
				Target: "i",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "i"},
					Right: LiteralExpr{Value: IntLiteral{Value: 1}},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(outerWhile, env)
	require.NoError(t, err)

	// count should be 3 * 2 = 6
	count, err := env.Get("count")
	require.NoError(t, err)
	assert.Equal(t, int64(6), count)
}

func TestExecuteWhile_WithComplexCondition(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Initialize variables
	env.Define("i", int64(0))
	env.Define("max", int64(7))

	// Create while loop: while i < max && i < 5
	whileStmt := WhileStatement{
		Condition: BinaryOpExpr{
			Op: And,
			Left: BinaryOpExpr{
				Op:    Lt,
				Left:  VariableExpr{Name: "i"},
				Right: VariableExpr{Name: "max"},
			},
			Right: BinaryOpExpr{
				Op:    Lt,
				Left:  VariableExpr{Name: "i"},
				Right: LiteralExpr{Value: IntLiteral{Value: 5}},
			},
		},
		Body: []Statement{
			AssignStatement{
				Target: "i",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "i"},
					Right: LiteralExpr{Value: IntLiteral{Value: 1}},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(whileStmt, env)
	require.NoError(t, err)

	// i should stop at 5, not 7
	i, err := env.Get("i")
	require.NoError(t, err)
	assert.Equal(t, int64(5), i)
}

func TestExecuteWhile_ImmediatelyFalse(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("i", int64(10))

	// Create while loop that never executes
	whileStmt := WhileStatement{
		Condition: BinaryOpExpr{
			Op:    Lt,
			Left:  VariableExpr{Name: "i"},
			Right: LiteralExpr{Value: IntLiteral{Value: 5}},
		},
		Body: []Statement{
			AssignStatement{
				Target: "i",
				Value:  LiteralExpr{Value: IntLiteral{Value: 0}},
			},
		},
	}

	_, err := interp.ExecuteStatement(whileStmt, env)
	require.NoError(t, err)

	// i should remain unchanged
	i, err := env.Get("i")
	require.NoError(t, err)
	assert.Equal(t, int64(10), i)
}

func TestExecuteWhile_WithIfStatement(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Initialize variables
	env.Define("sum", int64(0))
	env.Define("i", int64(0))

	// Create while loop with if statement inside
	whileStmt := WhileStatement{
		Condition: BinaryOpExpr{
			Op:    Lt,
			Left:  VariableExpr{Name: "i"},
			Right: LiteralExpr{Value: IntLiteral{Value: 10}},
		},
		Body: []Statement{
			IfStatement{
				Condition: BinaryOpExpr{
					Op:    Lt,
					Left:  VariableExpr{Name: "i"},
					Right: LiteralExpr{Value: IntLiteral{Value: 5}},
				},
				ThenBlock: []Statement{
					AssignStatement{
						Target: "sum",
						Value: BinaryOpExpr{
							Op:    Add,
							Left:  VariableExpr{Name: "sum"},
							Right: VariableExpr{Name: "i"},
						},
					},
				},
			},
			AssignStatement{
				Target: "i",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "i"},
					Right: LiteralExpr{Value: IntLiteral{Value: 1}},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(whileStmt, env)
	require.NoError(t, err)

	// sum should be 0+1+2+3+4 = 10
	sum, err := env.Get("sum")
	require.NoError(t, err)
	assert.Equal(t, int64(10), sum)

	// i should be 10
	i, err := env.Get("i")
	require.NoError(t, err)
	assert.Equal(t, int64(10), i)
}

func TestExecuteWhile_NonBooleanCondition(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Create while loop with non-boolean condition
	whileStmt := WhileStatement{
		Condition: LiteralExpr{Value: IntLiteral{Value: 1}},
		Body: []Statement{
			AssignStatement{
				Target: "i",
				Value:  LiteralExpr{Value: IntLiteral{Value: 0}},
			},
		},
	}

	_, err := interp.ExecuteStatement(whileStmt, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a boolean")
}

// Test for loop execution

func TestExecuteFor_ArrayIteration(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Create an array
	env.Define("items", []interface{}{int64(1), int64(2), int64(3)})
	env.Define("sum", int64(0))

	// for item in items { sum = sum + item }
	forStmt := ForStatement{
		ValueVar: "item",
		Iterable: VariableExpr{Name: "items"},
		Body: []Statement{
			AssignStatement{
				Target: "sum",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "sum"},
					Right: VariableExpr{Name: "item"},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(forStmt, env)
	require.NoError(t, err)

	// sum should be 6
	sum, err := env.Get("sum")
	require.NoError(t, err)
	assert.Equal(t, int64(6), sum)
}

func TestExecuteFor_ArrayWithIndex(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Create an array
	env.Define("items", []interface{}{"a", "b", "c"})
	env.Define("lastIndex", int64(-1))

	// for index, value in items { lastIndex = index }
	forStmt := ForStatement{
		KeyVar:   "index",
		ValueVar: "value",
		Iterable: VariableExpr{Name: "items"},
		Body: []Statement{
			AssignStatement{
				Target: "lastIndex",
				Value:  VariableExpr{Name: "index"},
			},
		},
	}

	_, err := interp.ExecuteStatement(forStmt, env)
	require.NoError(t, err)

	// lastIndex should be 2 (the last index)
	lastIndex, err := env.Get("lastIndex")
	require.NoError(t, err)
	assert.Equal(t, int64(2), lastIndex)
}

func TestExecuteFor_ObjectIteration(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Create an object
	obj := map[string]interface{}{
		"name": "Alice",
		"age":  int64(30),
	}
	env.Define("user", obj)
	env.Define("keys", []interface{}{})
	env.Define("lastKey", "")

	// for key, value in user { keys = [...keys, key] }
	// Note: This is simplified - we'll just test that keys are captured
	forStmt := ForStatement{
		KeyVar:   "key",
		ValueVar: "value",
		Iterable: VariableExpr{Name: "user"},
		Body: []Statement{
			AssignStatement{
				Target: "lastKey",
				Value:  VariableExpr{Name: "key"},
			},
		},
	}

	_, err := interp.ExecuteStatement(forStmt, env)
	require.NoError(t, err)

	// lastKey should be set to one of the keys
	lastKey, err := env.Get("lastKey")
	require.NoError(t, err)
	assert.Contains(t, []string{"name", "age"}, lastKey)
}

func TestExecuteFor_ArrayLiteral(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("count", int64(0))

	// for item in [1, 2, 3] { count = count + 1 }
	forStmt := ForStatement{
		ValueVar: "item",
		Iterable: ArrayExpr{
			Elements: []Expr{
				LiteralExpr{Value: IntLiteral{Value: 1}},
				LiteralExpr{Value: IntLiteral{Value: 2}},
				LiteralExpr{Value: IntLiteral{Value: 3}},
			},
		},
		Body: []Statement{
			AssignStatement{
				Target: "count",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "count"},
					Right: LiteralExpr{Value: IntLiteral{Value: 1}},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(forStmt, env)
	require.NoError(t, err)

	// count should be 3
	count, err := env.Get("count")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestExecuteFor_NestedLoops(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Create nested arrays
	env.Define("matrix", []interface{}{
		[]interface{}{int64(1), int64(2)},
		[]interface{}{int64(3), int64(4)},
	})
	env.Define("sum", int64(0))

	// for row in matrix { for cell in row { sum = sum + cell } }
	forStmt := ForStatement{
		ValueVar: "row",
		Iterable: VariableExpr{Name: "matrix"},
		Body: []Statement{
			ForStatement{
				ValueVar: "cell",
				Iterable: VariableExpr{Name: "row"},
				Body: []Statement{
					AssignStatement{
						Target: "sum",
						Value: BinaryOpExpr{
							Op:    Add,
							Left:  VariableExpr{Name: "sum"},
							Right: VariableExpr{Name: "cell"},
						},
					},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(forStmt, env)
	require.NoError(t, err)

	// sum should be 10 (1+2+3+4)
	sum, err := env.Get("sum")
	require.NoError(t, err)
	assert.Equal(t, int64(10), sum)
}

func TestExecuteFor_InvalidIterable(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Try to iterate over a number
	env.Define("num", int64(42))

	forStmt := ForStatement{
		ValueVar: "item",
		Iterable: VariableExpr{Name: "num"},
		Body: []Statement{
			AssignStatement{
				Target: "x",
				Value:  LiteralExpr{Value: IntLiteral{Value: 1}},
			},
		},
	}

	_, err := interp.ExecuteStatement(forStmt, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an array or object")
}

func TestExecuteFor_ScopeIsolation(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("items", []interface{}{int64(1), int64(2), int64(3)})

	// for item in items { $ temp = item * 2 }
	forStmt := ForStatement{
		ValueVar: "item",
		Iterable: VariableExpr{Name: "items"},
		Body: []Statement{
			AssignStatement{
				Target: "temp",
				Value: BinaryOpExpr{
					Op:    Mul,
					Left:  VariableExpr{Name: "item"},
					Right: LiteralExpr{Value: IntLiteral{Value: 2}},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(forStmt, env)
	require.NoError(t, err)

	// item should not be accessible in parent scope
	_, err = env.Get("item")
	assert.Error(t, err)

	// temp should not be accessible in parent scope
	_, err = env.Get("temp")
	assert.Error(t, err)
}

// Test Switch Statement Execution

func TestExecuteSwitch_IntegerMatch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("status", int64(1))
	env.Define("result", "")

	// switch status { case 1 { result = "one" } case 2 { result = "two" } default { result = "other" } }
	switchStmt := SwitchStatement{
		Value: VariableExpr{Name: "status"},
		Cases: []SwitchCase{
			{
				Value: LiteralExpr{Value: IntLiteral{Value: 1}},
				Body: []Statement{
					AssignStatement{
						Target: "result",
						Value:  LiteralExpr{Value: StringLiteral{Value: "one"}},
					},
				},
			},
			{
				Value: LiteralExpr{Value: IntLiteral{Value: 2}},
				Body: []Statement{
					AssignStatement{
						Target: "result",
						Value:  LiteralExpr{Value: StringLiteral{Value: "two"}},
					},
				},
			},
		},
		Default: []Statement{
			AssignStatement{
				Target: "result",
				Value:  LiteralExpr{Value: StringLiteral{Value: "other"}},
			},
		},
	}

	_, err := interp.ExecuteStatement(switchStmt, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, "one", result)
}

func TestExecuteSwitch_StringMatch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("status", "pending")
	env.Define("message", "")

	// switch status { case "pending" { message = "Order pending" } case "shipped" { message = "Order shipped" } }
	switchStmt := SwitchStatement{
		Value: VariableExpr{Name: "status"},
		Cases: []SwitchCase{
			{
				Value: LiteralExpr{Value: StringLiteral{Value: "pending"}},
				Body: []Statement{
					AssignStatement{
						Target: "message",
						Value:  LiteralExpr{Value: StringLiteral{Value: "Order pending"}},
					},
				},
			},
			{
				Value: LiteralExpr{Value: StringLiteral{Value: "shipped"}},
				Body: []Statement{
					AssignStatement{
						Target: "message",
						Value:  LiteralExpr{Value: StringLiteral{Value: "Order shipped"}},
					},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(switchStmt, env)
	require.NoError(t, err)

	message, err := env.Get("message")
	require.NoError(t, err)
	assert.Equal(t, "Order pending", message)
}

func TestExecuteSwitch_DefaultCase(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("value", int64(99))
	env.Define("result", "")

	// switch value { case 1 { result = "one" } default { result = "unknown" } }
	switchStmt := SwitchStatement{
		Value: VariableExpr{Name: "value"},
		Cases: []SwitchCase{
			{
				Value: LiteralExpr{Value: IntLiteral{Value: 1}},
				Body: []Statement{
					AssignStatement{
						Target: "result",
						Value:  LiteralExpr{Value: StringLiteral{Value: "one"}},
					},
				},
			},
		},
		Default: []Statement{
			AssignStatement{
				Target: "result",
				Value:  LiteralExpr{Value: StringLiteral{Value: "unknown"}},
			},
		},
	}

	_, err := interp.ExecuteStatement(switchStmt, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, "unknown", result)
}

func TestExecuteSwitch_NoDefault_NoMatch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("value", int64(99))

	// switch value { case 1 { $ x = 1 } }
	switchStmt := SwitchStatement{
		Value: VariableExpr{Name: "value"},
		Cases: []SwitchCase{
			{
				Value: LiteralExpr{Value: IntLiteral{Value: 1}},
				Body: []Statement{
					AssignStatement{
						Target: "x",
						Value:  LiteralExpr{Value: IntLiteral{Value: 1}},
					},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(switchStmt, env)
	require.NoError(t, err)

	// x should not exist since no case matched
	_, err = env.Get("x")
	assert.Error(t, err)
}

func TestExecuteSwitch_NoFallthrough(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("value", int64(1))
	env.Define("count", int64(0))

	// switch value { case 1 { count = count + 1 } case 1 { count = count + 10 } }
	// Only first match should execute
	switchStmt := SwitchStatement{
		Value: VariableExpr{Name: "value"},
		Cases: []SwitchCase{
			{
				Value: LiteralExpr{Value: IntLiteral{Value: 1}},
				Body: []Statement{
					AssignStatement{
						Target: "count",
						Value: BinaryOpExpr{
							Op:    Add,
							Left:  VariableExpr{Name: "count"},
							Right: LiteralExpr{Value: IntLiteral{Value: 1}},
						},
					},
				},
			},
			{
				Value: LiteralExpr{Value: IntLiteral{Value: 1}},
				Body: []Statement{
					AssignStatement{
						Target: "count",
						Value: BinaryOpExpr{
							Op:    Add,
							Left:  VariableExpr{Name: "count"},
							Right: LiteralExpr{Value: IntLiteral{Value: 10}},
						},
					},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(switchStmt, env)
	require.NoError(t, err)

	count, err := env.Get("count")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count) // Should be 1, not 11
}

func TestExecuteSwitch_WithReturnStatement(t *testing.T) {
	interp := NewInterpreter()

	// @ GET /test
	//   switch 1 {
	//     case 1 { > "matched" }
	//     default { > "no match" }
	//   }
	route := &Route{
		Path:   "/test",
		Method: Get,
		Body: []Statement{
			SwitchStatement{
				Value: LiteralExpr{Value: IntLiteral{Value: 1}},
				Cases: []SwitchCase{
					{
						Value: LiteralExpr{Value: IntLiteral{Value: 1}},
						Body: []Statement{
							ReturnStatement{
								Value: LiteralExpr{Value: StringLiteral{Value: "matched"}},
							},
						},
					},
				},
				Default: []Statement{
					ReturnStatement{
						Value: LiteralExpr{Value: StringLiteral{Value: "no match"}},
					},
				},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)
	require.NoError(t, err)
	assert.Equal(t, "matched", result)
}

func TestExecuteSwitch_BooleanValues(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("flag", true)
	env.Define("message", "")

	// switch flag { case true { message = "on" } case false { message = "off" } }
	switchStmt := SwitchStatement{
		Value: VariableExpr{Name: "flag"},
		Cases: []SwitchCase{
			{
				Value: LiteralExpr{Value: BoolLiteral{Value: true}},
				Body: []Statement{
					AssignStatement{
						Target: "message",
						Value:  LiteralExpr{Value: StringLiteral{Value: "on"}},
					},
				},
			},
			{
				Value: LiteralExpr{Value: BoolLiteral{Value: false}},
				Body: []Statement{
					AssignStatement{
						Target: "message",
						Value:  LiteralExpr{Value: StringLiteral{Value: "off"}},
					},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(switchStmt, env)
	require.NoError(t, err)

	message, err := env.Get("message")
	require.NoError(t, err)
	assert.Equal(t, "on", message)
}

func TestExecuteSwitch_MultipleStatements(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("value", int64(1))
	env.Define("x", int64(0))
	env.Define("y", int64(0))

	// switch value { case 1 { x = 10; y = 20 } }
	switchStmt := SwitchStatement{
		Value: VariableExpr{Name: "value"},
		Cases: []SwitchCase{
			{
				Value: LiteralExpr{Value: IntLiteral{Value: 1}},
				Body: []Statement{
					AssignStatement{
						Target: "x",
						Value:  LiteralExpr{Value: IntLiteral{Value: 10}},
					},
					AssignStatement{
						Target: "y",
						Value:  LiteralExpr{Value: IntLiteral{Value: 20}},
					},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(switchStmt, env)
	require.NoError(t, err)

	x, err := env.Get("x")
	require.NoError(t, err)
	assert.Equal(t, int64(10), x)

	y, err := env.Get("y")
	require.NoError(t, err)
	assert.Equal(t, int64(20), y)
}

func TestExecuteSwitch_NestedSwitch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("outer", int64(1))
	env.Define("inner", int64(2))
	env.Define("result", "")

	// switch outer { case 1 { switch inner { case 2 { result = "matched" } } } }
	switchStmt := SwitchStatement{
		Value: VariableExpr{Name: "outer"},
		Cases: []SwitchCase{
			{
				Value: LiteralExpr{Value: IntLiteral{Value: 1}},
				Body: []Statement{
					SwitchStatement{
						Value: VariableExpr{Name: "inner"},
						Cases: []SwitchCase{
							{
								Value: LiteralExpr{Value: IntLiteral{Value: 2}},
								Body: []Statement{
									AssignStatement{
										Target: "result",
										Value:  LiteralExpr{Value: StringLiteral{Value: "matched"}},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(switchStmt, env)
	require.NoError(t, err)

	result, err := env.Get("result")
	require.NoError(t, err)
	assert.Equal(t, "matched", result)
}

// Test CLI Command Execution

func TestInterpreter_ExecuteCommand_Simple(t *testing.T) {
	interp := NewInterpreter()

	cmd := &Command{
		Name: "greet",
		Params: []CommandParam{
			{
				Name:     "name",
				Type:     StringType{},
				Required: true,
			},
		},
		Body: []Statement{
			ReturnStatement{
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  LiteralExpr{Value: StringLiteral{Value: "Hello, "}},
					Right: VariableExpr{Name: "name"},
				},
			},
		},
	}

	args := map[string]interface{}{
		"name": "World",
	}

	result, err := interp.ExecuteCommand(cmd, args)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World", result)
}

func TestInterpreter_ExecuteCommand_WithDefaultValue(t *testing.T) {
	interp := NewInterpreter()

	cmd := &Command{
		Name: "greet",
		Params: []CommandParam{
			{
				Name:     "name",
				Type:     StringType{},
				Required: true,
			},
			{
				Name:    "greeting",
				Type:    StringType{},
				Default: LiteralExpr{Value: StringLiteral{Value: "Hello"}},
			},
		},
		Body: []Statement{
			ReturnStatement{
				Value: BinaryOpExpr{
					Op: Add,
					Left: BinaryOpExpr{
						Op:    Add,
						Left:  VariableExpr{Name: "greeting"},
						Right: LiteralExpr{Value: StringLiteral{Value: ", "}},
					},
					Right: VariableExpr{Name: "name"},
				},
			},
		},
	}

	// Test with explicit greeting
	args := map[string]interface{}{
		"name":     "Alice",
		"greeting": "Hi",
	}
	result, err := interp.ExecuteCommand(cmd, args)
	require.NoError(t, err)
	assert.Equal(t, "Hi, Alice", result)

	// Test with default greeting
	args = map[string]interface{}{
		"name": "Bob",
	}
	result, err = interp.ExecuteCommand(cmd, args)
	require.NoError(t, err)
	assert.Equal(t, "Hello, Bob", result)
}

func TestInterpreter_ExecuteCommand_MissingRequiredParam(t *testing.T) {
	interp := NewInterpreter()

	cmd := &Command{
		Name: "greet",
		Params: []CommandParam{
			{
				Name:     "name",
				Type:     StringType{},
				Required: true,
			},
		},
		Body: []Statement{
			ReturnStatement{
				Value: LiteralExpr{Value: StringLiteral{Value: "Hello"}},
			},
		},
	}

	args := map[string]interface{}{}

	_, err := interp.ExecuteCommand(cmd, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required argument")
}

func TestInterpreter_ExecuteCommand_WithMultipleStatements(t *testing.T) {
	interp := NewInterpreter()

	cmd := &Command{
		Name: "calculate",
		Params: []CommandParam{
			{Name: "x", Type: IntType{}, Required: true},
			{Name: "y", Type: IntType{}, Required: true},
		},
		Body: []Statement{
			AssignStatement{
				Target: "sum",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "x"},
					Right: VariableExpr{Name: "y"},
				},
			},
			AssignStatement{
				Target: "product",
				Value: BinaryOpExpr{
					Op:    Mul,
					Left:  VariableExpr{Name: "x"},
					Right: VariableExpr{Name: "y"},
				},
			},
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "sum", Value: VariableExpr{Name: "sum"}},
						{Key: "product", Value: VariableExpr{Name: "product"}},
					},
				},
			},
		},
	}

	args := map[string]interface{}{
		"x": int64(5),
		"y": int64(3),
	}

	result, err := interp.ExecuteCommand(cmd, args)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(8), resultMap["sum"])
	assert.Equal(t, int64(15), resultMap["product"])
}

// Test Cron Task Execution

func TestInterpreter_ExecuteCronTask_Simple(t *testing.T) {
	interp := NewInterpreter()

	task := &CronTask{
		Name:     "daily_cleanup",
		Schedule: "0 0 * * *",
		Body: []Statement{
			AssignStatement{
				Target: "count",
				Value:  LiteralExpr{Value: IntLiteral{Value: 10}},
			},
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "deleted", Value: VariableExpr{Name: "count"}},
					},
				},
			},
		},
	}

	result, err := interp.ExecuteCronTask(task)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(10), resultMap["deleted"])
}

func TestInterpreter_ExecuteCronTask_WithInjections(t *testing.T) {
	interp := NewInterpreter()

	// Set up a mock database handler
	mockDB := map[string]interface{}{
		"recordCount": int64(42),
	}
	interp.SetDatabaseHandler(mockDB)

	task := &CronTask{
		Name:     "backup",
		Schedule: "0 0 * * *",
		Injections: []Injection{
			{Name: "db", Type: DatabaseType{}},
		},
		Body: []Statement{
			AssignStatement{
				Target: "count",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "db"},
					Field:  "recordCount",
				},
			},
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "backed_up", Value: VariableExpr{Name: "count"}},
					},
				},
			},
		},
	}

	result, err := interp.ExecuteCronTask(task)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(42), resultMap["backed_up"])
}

func TestInterpreter_ExecuteCronTask_ComplexLogic(t *testing.T) {
	interp := NewInterpreter()

	task := &CronTask{
		Name:     "hourly_sync",
		Schedule: "0 */1 * * *",
		Body: []Statement{
			AssignStatement{
				Target: "total",
				Value:  LiteralExpr{Value: IntLiteral{Value: 0}},
			},
			AssignStatement{
				Target: "i",
				Value:  LiteralExpr{Value: IntLiteral{Value: 0}},
			},
			WhileStatement{
				Condition: BinaryOpExpr{
					Op:    Lt,
					Left:  VariableExpr{Name: "i"},
					Right: LiteralExpr{Value: IntLiteral{Value: 5}},
				},
				Body: []Statement{
					AssignStatement{
						Target: "total",
						Value: BinaryOpExpr{
							Op:    Add,
							Left:  VariableExpr{Name: "total"},
							Right: VariableExpr{Name: "i"},
						},
					},
					AssignStatement{
						Target: "i",
						Value: BinaryOpExpr{
							Op:    Add,
							Left:  VariableExpr{Name: "i"},
							Right: LiteralExpr{Value: IntLiteral{Value: 1}},
						},
					},
				},
			},
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "synced", Value: VariableExpr{Name: "total"}},
					},
				},
			},
		},
	}

	result, err := interp.ExecuteCronTask(task)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(10), resultMap["synced"]) // 0+1+2+3+4 = 10
}

// Test Event Handler Execution

func TestInterpreter_ExecuteEventHandler_Simple(t *testing.T) {
	interp := NewInterpreter()

	handler := &EventHandler{
		EventType: "user.created",
		Body: []Statement{
			AssignStatement{
				Target: "userId",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "event"},
					Field:  "userId",
				},
			},
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "handled", Value: LiteralExpr{Value: BoolLiteral{Value: true}}},
						{Key: "userId", Value: VariableExpr{Name: "userId"}},
					},
				},
			},
		},
	}

	eventData := map[string]interface{}{
		"userId": int64(123),
		"email":  "user@example.com",
	}

	result, err := interp.ExecuteEventHandler(handler, eventData)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, resultMap["handled"])
	assert.Equal(t, int64(123), resultMap["userId"])
}

func TestInterpreter_ExecuteEventHandler_WithInput(t *testing.T) {
	interp := NewInterpreter()

	handler := &EventHandler{
		EventType: "order.paid",
		Body: []Statement{
			AssignStatement{
				Target: "orderId",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "input"},
					Field:  "orderId",
				},
			},
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "orderId", Value: VariableExpr{Name: "orderId"}},
					},
				},
			},
		},
	}

	eventData := map[string]interface{}{
		"orderId": int64(456),
		"amount":  100.50,
	}

	result, err := interp.ExecuteEventHandler(handler, eventData)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(456), resultMap["orderId"])
}

func TestInterpreter_ExecuteEventHandler_WithInjections(t *testing.T) {
	interp := NewInterpreter()

	mockDB := map[string]interface{}{
		"connected": true,
	}
	interp.SetDatabaseHandler(mockDB)

	handler := &EventHandler{
		EventType: "notification.send",
		Injections: []Injection{
			{Name: "db", Type: DatabaseType{}},
		},
		Body: []Statement{
			AssignStatement{
				Target: "isConnected",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "db"},
					Field:  "connected",
				},
			},
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "sent", Value: VariableExpr{Name: "isConnected"}},
					},
				},
			},
		},
	}

	eventData := map[string]interface{}{
		"message": "Hello",
	}

	result, err := interp.ExecuteEventHandler(handler, eventData)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, resultMap["sent"])
}

func TestInterpreter_EmitEvent_SingleHandler(t *testing.T) {
	interp := NewInterpreter()

	// Register event handler
	module := Module{
		Items: []Item{
			&EventHandler{
				EventType: "test.event",
				Body: []Statement{
					ReturnStatement{
						Value: LiteralExpr{Value: StringLiteral{Value: "handled"}},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	eventData := map[string]interface{}{
		"data": "test",
	}

	err = interp.EmitEvent("test.event", eventData)
	assert.NoError(t, err)
}

func TestInterpreter_EmitEvent_MultipleHandlers(t *testing.T) {
	interp := NewInterpreter()

	// Register multiple event handlers for the same event
	module := Module{
		Items: []Item{
			&EventHandler{
				EventType: "user.created",
				Body: []Statement{
					ReturnStatement{
						Value: LiteralExpr{Value: StringLiteral{Value: "handler1"}},
					},
				},
			},
			&EventHandler{
				EventType: "user.created",
				Body: []Statement{
					ReturnStatement{
						Value: LiteralExpr{Value: StringLiteral{Value: "handler2"}},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	eventData := map[string]interface{}{
		"userId": int64(1),
	}

	err = interp.EmitEvent("user.created", eventData)
	assert.NoError(t, err)
}

// Test Queue Worker Execution

func TestInterpreter_ExecuteQueueWorker_Simple(t *testing.T) {
	interp := NewInterpreter()

	worker := &QueueWorker{
		QueueName: "email.send",
		Body: []Statement{
			AssignStatement{
				Target: "to",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "message"},
					Field:  "to",
				},
			},
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "sent", Value: LiteralExpr{Value: BoolLiteral{Value: true}}},
						{Key: "to", Value: VariableExpr{Name: "to"}},
					},
				},
			},
		},
	}

	message := map[string]interface{}{
		"to":      "user@example.com",
		"subject": "Hello",
		"body":    "Test message",
	}

	result, err := interp.ExecuteQueueWorker(worker, message)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, resultMap["sent"])
	assert.Equal(t, "user@example.com", resultMap["to"])
}

func TestInterpreter_ExecuteQueueWorker_WithInput(t *testing.T) {
	interp := NewInterpreter()

	worker := &QueueWorker{
		QueueName: "image.resize",
		Body: []Statement{
			AssignStatement{
				Target: "imageId",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "input"},
					Field:  "imageId",
				},
			},
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "processed", Value: LiteralExpr{Value: BoolLiteral{Value: true}}},
						{Key: "imageId", Value: VariableExpr{Name: "imageId"}},
					},
				},
			},
		},
	}

	message := map[string]interface{}{
		"imageId": int64(789),
		"width":   800,
		"height":  600,
	}

	result, err := interp.ExecuteQueueWorker(worker, message)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, resultMap["processed"])
	assert.Equal(t, int64(789), resultMap["imageId"])
}

func TestInterpreter_ExecuteQueueWorker_WithInjections(t *testing.T) {
	interp := NewInterpreter()

	mockDB := map[string]interface{}{
		"status": "connected",
	}
	interp.SetDatabaseHandler(mockDB)

	worker := &QueueWorker{
		QueueName: "report.generate",
		Injections: []Injection{
			{Name: "db", Type: DatabaseType{}},
		},
		Body: []Statement{
			AssignStatement{
				Target: "dbStatus",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "db"},
					Field:  "status",
				},
			},
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "generated", Value: LiteralExpr{Value: BoolLiteral{Value: true}}},
						{Key: "dbStatus", Value: VariableExpr{Name: "dbStatus"}},
					},
				},
			},
		},
	}

	message := map[string]interface{}{
		"reportId": int64(100),
	}

	result, err := interp.ExecuteQueueWorker(worker, message)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, resultMap["generated"])
	assert.Equal(t, "connected", resultMap["dbStatus"])
}

func TestInterpreter_ExecuteQueueWorker_ComplexProcessing(t *testing.T) {
	interp := NewInterpreter()

	worker := &QueueWorker{
		QueueName: "data.process",
		Body: []Statement{
			AssignStatement{
				Target: "items",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "message"},
					Field:  "items",
				},
			},
			AssignStatement{
				Target: "count",
				Value:  LiteralExpr{Value: IntLiteral{Value: 0}},
			},
			ForStatement{
				ValueVar: "item",
				Iterable: VariableExpr{Name: "items"},
				Body: []Statement{
					AssignStatement{
						Target: "count",
						Value: BinaryOpExpr{
							Op:    Add,
							Left:  VariableExpr{Name: "count"},
							Right: LiteralExpr{Value: IntLiteral{Value: 1}},
						},
					},
				},
			},
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "processed", Value: VariableExpr{Name: "count"}},
					},
				},
			},
		},
	}

	message := map[string]interface{}{
		"items": []interface{}{
			int64(1), int64(2), int64(3), int64(4), int64(5),
		},
	}

	result, err := interp.ExecuteQueueWorker(worker, message)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(5), resultMap["processed"])
}

// Test GetCommand, GetCronTasks, GetEventHandlers, GetQueueWorker functions

func TestInterpreter_GetCommand(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&Command{
				Name: "test",
				Body: []Statement{},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	cmd, ok := interp.GetCommand("test")
	assert.True(t, ok)
	assert.Equal(t, "test", cmd.Name)

	_, ok = interp.GetCommand("nonexistent")
	assert.False(t, ok)
}

func TestInterpreter_GetCronTasks(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&CronTask{
				Name:     "task1",
				Schedule: "0 0 * * *",
			},
			&CronTask{
				Name:     "task2",
				Schedule: "*/5 * * * *",
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	tasks := interp.GetCronTasks()
	assert.Len(t, tasks, 2)
	assert.Equal(t, "task1", tasks[0].Name)
	assert.Equal(t, "task2", tasks[1].Name)
}

func TestInterpreter_GetEventHandlers(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&EventHandler{
				EventType: "user.created",
				Body:      []Statement{},
			},
			&EventHandler{
				EventType: "user.created",
				Body:      []Statement{},
			},
			&EventHandler{
				EventType: "order.paid",
				Body:      []Statement{},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	handlers := interp.GetEventHandlers("user.created")
	assert.Len(t, handlers, 2)

	handlers = interp.GetEventHandlers("order.paid")
	assert.Len(t, handlers, 1)

	handlers = interp.GetEventHandlers("nonexistent")
	assert.Len(t, handlers, 0)
}

func TestInterpreter_GetQueueWorker(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&QueueWorker{
				QueueName: "email.send",
				Body:      []Statement{},
			},
		},
	}

	err := interp.LoadModule(module)
	require.NoError(t, err)

	worker, ok := interp.GetQueueWorker("email.send")
	assert.True(t, ok)
	assert.Equal(t, "email.send", worker.QueueName)

	_, ok = interp.GetQueueWorker("nonexistent")
	assert.False(t, ok)
}

// Test built-in string functions

func TestBuiltinFunction_Upper(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := FunctionCallExpr{
		Name: "upper",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "hello world"}},
		},
	}

	result, err := interp.evaluateFunctionCall(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "HELLO WORLD", result)
}

func TestBuiltinFunction_Lower(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := FunctionCallExpr{
		Name: "lower",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "HELLO WORLD"}},
		},
	}

	result, err := interp.evaluateFunctionCall(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestBuiltinFunction_Trim(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := FunctionCallExpr{
		Name: "trim",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "  hello world  "}},
		},
	}

	result, err := interp.evaluateFunctionCall(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestBuiltinFunction_Split(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := FunctionCallExpr{
		Name: "split",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "a,b,c"}},
			LiteralExpr{Value: StringLiteral{Value: ","}},
		},
	}

	result, err := interp.evaluateFunctionCall(expr, env)
	require.NoError(t, err)

	arr, ok := result.([]interface{})
	require.True(t, ok)
	assert.Len(t, arr, 3)
	assert.Equal(t, "a", arr[0])
	assert.Equal(t, "b", arr[1])
	assert.Equal(t, "c", arr[2])
}

func TestBuiltinFunction_Join(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := FunctionCallExpr{
		Name: "join",
		Args: []Expr{
			ArrayExpr{
				Elements: []Expr{
					LiteralExpr{Value: StringLiteral{Value: "a"}},
					LiteralExpr{Value: StringLiteral{Value: "b"}},
					LiteralExpr{Value: StringLiteral{Value: "c"}},
				},
			},
			LiteralExpr{Value: StringLiteral{Value: ","}},
		},
	}

	result, err := interp.evaluateFunctionCall(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "a,b,c", result)
}

func TestBuiltinFunction_Contains(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	tests := []struct {
		name     string
		str      string
		substr   string
		expected bool
	}{
		{"contains_yes", "hello world", "world", true},
		{"contains_no", "hello world", "xyz", false},
		{"contains_empty", "hello", "", true},
		{"contains_exact", "test", "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := FunctionCallExpr{
				Name: "contains",
				Args: []Expr{
					LiteralExpr{Value: StringLiteral{Value: tt.str}},
					LiteralExpr{Value: StringLiteral{Value: tt.substr}},
				},
			}

			result, err := interp.evaluateFunctionCall(expr, env)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuiltinFunction_Replace(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := FunctionCallExpr{
		Name: "replace",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "hello world world"}},
			LiteralExpr{Value: StringLiteral{Value: "world"}},
			LiteralExpr{Value: StringLiteral{Value: "universe"}},
		},
	}

	result, err := interp.evaluateFunctionCall(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "hello universe universe", result)
}

func TestBuiltinFunction_Substring(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	tests := []struct {
		name     string
		str      string
		start    int64
		end      int64
		expected string
	}{
		{"normal_range", "hello world", 0, 5, "hello"},
		{"middle_range", "hello world", 6, 11, "world"},
		{"full_string", "test", 0, 4, "test"},
		{"empty_string", "test", 2, 2, ""},
		{"beyond_length", "test", 0, 10, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := FunctionCallExpr{
				Name: "substring",
				Args: []Expr{
					LiteralExpr{Value: StringLiteral{Value: tt.str}},
					LiteralExpr{Value: IntLiteral{Value: tt.start}},
					LiteralExpr{Value: IntLiteral{Value: tt.end}},
				},
			}

			result, err := interp.evaluateFunctionCall(expr, env)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuiltinFunction_Length(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	t.Run("string_length", func(t *testing.T) {
		expr := FunctionCallExpr{
			Name: "length",
			Args: []Expr{
				LiteralExpr{Value: StringLiteral{Value: "hello"}},
			},
		}

		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.Equal(t, int64(5), result)
	})

	t.Run("array_length", func(t *testing.T) {
		expr := FunctionCallExpr{
			Name: "length",
			Args: []Expr{
				ArrayExpr{
					Elements: []Expr{
						LiteralExpr{Value: IntLiteral{Value: 1}},
						LiteralExpr{Value: IntLiteral{Value: 2}},
						LiteralExpr{Value: IntLiteral{Value: 3}},
					},
				},
			},
		}

		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result)
	})
}

func TestBuiltinFunction_StartsWith(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	tests := []struct {
		str      string
		prefix   string
		expected bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", false},
		{"hello", "hello", true},
		{"hello", "h", true},
		{"hello", "", true},
	}

	for _, tt := range tests {
		expr := FunctionCallExpr{
			Name: "startsWith",
			Args: []Expr{
				LiteralExpr{Value: StringLiteral{Value: tt.str}},
				LiteralExpr{Value: StringLiteral{Value: tt.prefix}},
			},
		}

		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, result)
	}
}

func TestBuiltinFunction_EndsWith(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	tests := []struct {
		str      string
		suffix   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", false},
		{"hello", "hello", true},
		{"hello", "o", true},
		{"hello", "", true},
	}

	for _, tt := range tests {
		expr := FunctionCallExpr{
			Name: "endsWith",
			Args: []Expr{
				LiteralExpr{Value: StringLiteral{Value: tt.str}},
				LiteralExpr{Value: StringLiteral{Value: tt.suffix}},
			},
		}

		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, result)
	}
}

func TestBuiltinFunction_IndexOf(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	tests := []struct {
		str      string
		substr   string
		expected int64
	}{
		{"hello world", "world", 6},
		{"hello world", "hello", 0},
		{"hello world", "not found", -1},
		{"hello", "l", 2},
	}

	for _, tt := range tests {
		expr := FunctionCallExpr{
			Name: "indexOf",
			Args: []Expr{
				LiteralExpr{Value: StringLiteral{Value: tt.str}},
				LiteralExpr{Value: StringLiteral{Value: tt.substr}},
			},
		}

		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, result)
	}
}

func TestBuiltinFunction_CharAt(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := FunctionCallExpr{
		Name: "charAt",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "hello"}},
			LiteralExpr{Value: IntLiteral{Value: 1}},
		},
	}

	result, err := interp.evaluateFunctionCall(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "e", result)
}

func TestBuiltinFunction_ParseInt(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	tests := []struct {
		input    string
		expected int64
	}{
		{"42", 42},
		{"-10", -10},
		{"  123  ", 123},
	}

	for _, tt := range tests {
		expr := FunctionCallExpr{
			Name: "parseInt",
			Args: []Expr{
				LiteralExpr{Value: StringLiteral{Value: tt.input}},
			},
		}

		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, result)
	}
}

func TestBuiltinFunction_ParseFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := FunctionCallExpr{
		Name: "parseFloat",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "3.14"}},
		},
	}

	result, err := interp.evaluateFunctionCall(expr, env)
	require.NoError(t, err)
	assert.InDelta(t, 3.14, result.(float64), 0.001)
}

func TestBuiltinFunction_ToString(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	tests := []struct {
		name     string
		input    Literal
		expected string
	}{
		{"int", IntLiteral{Value: 42}, "42"},
		{"float", FloatLiteral{Value: 3.14}, "3.14"},
		{"bool", BoolLiteral{Value: true}, "true"},
		{"string", StringLiteral{Value: "hello"}, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := FunctionCallExpr{
				Name: "toString",
				Args: []Expr{
					LiteralExpr{Value: tt.input},
				},
			}

			result, err := interp.evaluateFunctionCall(expr, env)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuiltinFunction_Abs(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	t.Run("positive_int", func(t *testing.T) {
		expr := FunctionCallExpr{
			Name: "abs",
			Args: []Expr{
				LiteralExpr{Value: IntLiteral{Value: 42}},
			},
		}
		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)
	})

	t.Run("negative_int", func(t *testing.T) {
		expr := FunctionCallExpr{
			Name: "abs",
			Args: []Expr{
				LiteralExpr{Value: IntLiteral{Value: -42}},
			},
		}
		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)
	})

	t.Run("negative_float", func(t *testing.T) {
		expr := FunctionCallExpr{
			Name: "abs",
			Args: []Expr{
				LiteralExpr{Value: FloatLiteral{Value: -3.14}},
			},
		}
		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.InDelta(t, 3.14, result.(float64), 0.001)
	})
}

func TestBuiltinFunction_Min(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	t.Run("int_min", func(t *testing.T) {
		expr := FunctionCallExpr{
			Name: "min",
			Args: []Expr{
				LiteralExpr{Value: IntLiteral{Value: 5}},
				LiteralExpr{Value: IntLiteral{Value: 3}},
			},
		}
		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result)
	})

	t.Run("float_min", func(t *testing.T) {
		expr := FunctionCallExpr{
			Name: "min",
			Args: []Expr{
				LiteralExpr{Value: FloatLiteral{Value: 5.5}},
				LiteralExpr{Value: FloatLiteral{Value: 3.3}},
			},
		}
		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.InDelta(t, 3.3, result.(float64), 0.001)
	})
}

func TestBuiltinFunction_Max(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	t.Run("int_max", func(t *testing.T) {
		expr := FunctionCallExpr{
			Name: "max",
			Args: []Expr{
				LiteralExpr{Value: IntLiteral{Value: 5}},
				LiteralExpr{Value: IntLiteral{Value: 3}},
			},
		}
		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.Equal(t, int64(5), result)
	})

	t.Run("float_max", func(t *testing.T) {
		expr := FunctionCallExpr{
			Name: "max",
			Args: []Expr{
				LiteralExpr{Value: FloatLiteral{Value: 5.5}},
				LiteralExpr{Value: FloatLiteral{Value: 3.3}},
			},
		}
		result, err := interp.evaluateFunctionCall(expr, env)
		require.NoError(t, err)
		assert.InDelta(t, 5.5, result.(float64), 0.001)
	})
}

// Test Pattern Matching

func TestMatchExpr_LiteralPattern_Int(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// match 42 { 42 => "found", _ => "not found" }
	expr := MatchExpr{
		Value: LiteralExpr{Value: IntLiteral{Value: 42}},
		Cases: []MatchCase{
			{
				Pattern: LiteralPattern{Value: IntLiteral{Value: 42}},
				Body:    LiteralExpr{Value: StringLiteral{Value: "found"}},
			},
			{
				Pattern: WildcardPattern{},
				Body:    LiteralExpr{Value: StringLiteral{Value: "not found"}},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "found", result)
}

func TestMatchExpr_LiteralPattern_String(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// match "hello" { "hello" => "greeting", "bye" => "farewell", _ => "unknown" }
	expr := MatchExpr{
		Value: LiteralExpr{Value: StringLiteral{Value: "hello"}},
		Cases: []MatchCase{
			{
				Pattern: LiteralPattern{Value: StringLiteral{Value: "hello"}},
				Body:    LiteralExpr{Value: StringLiteral{Value: "greeting"}},
			},
			{
				Pattern: LiteralPattern{Value: StringLiteral{Value: "bye"}},
				Body:    LiteralExpr{Value: StringLiteral{Value: "farewell"}},
			},
			{
				Pattern: WildcardPattern{},
				Body:    LiteralExpr{Value: StringLiteral{Value: "unknown"}},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "greeting", result)
}

func TestMatchExpr_WildcardPattern(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// match 99 { 1 => "one", 2 => "two", _ => "other" }
	expr := MatchExpr{
		Value: LiteralExpr{Value: IntLiteral{Value: 99}},
		Cases: []MatchCase{
			{
				Pattern: LiteralPattern{Value: IntLiteral{Value: 1}},
				Body:    LiteralExpr{Value: StringLiteral{Value: "one"}},
			},
			{
				Pattern: LiteralPattern{Value: IntLiteral{Value: 2}},
				Body:    LiteralExpr{Value: StringLiteral{Value: "two"}},
			},
			{
				Pattern: WildcardPattern{},
				Body:    LiteralExpr{Value: StringLiteral{Value: "other"}},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "other", result)
}

func TestMatchExpr_VariablePattern(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// match 42 { x => x + 10 }
	expr := MatchExpr{
		Value: LiteralExpr{Value: IntLiteral{Value: 42}},
		Cases: []MatchCase{
			{
				Pattern: VariablePattern{Name: "x"},
				Body: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "x"},
					Right: LiteralExpr{Value: IntLiteral{Value: 10}},
				},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, int64(52), result)
}

func TestMatchExpr_WithGuard(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// match 10 { x when x > 5 => "big", x => "small" }
	expr := MatchExpr{
		Value: LiteralExpr{Value: IntLiteral{Value: 10}},
		Cases: []MatchCase{
			{
				Pattern: VariablePattern{Name: "x"},
				Guard: BinaryOpExpr{
					Op:    Gt,
					Left:  VariableExpr{Name: "x"},
					Right: LiteralExpr{Value: IntLiteral{Value: 5}},
				},
				Body: LiteralExpr{Value: StringLiteral{Value: "big"}},
			},
			{
				Pattern: VariablePattern{Name: "x"},
				Body:    LiteralExpr{Value: StringLiteral{Value: "small"}},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "big", result)
}

func TestMatchExpr_WithGuard_NotMatching(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// match 3 { x when x > 5 => "big", x => "small" }
	expr := MatchExpr{
		Value: LiteralExpr{Value: IntLiteral{Value: 3}},
		Cases: []MatchCase{
			{
				Pattern: VariablePattern{Name: "x"},
				Guard: BinaryOpExpr{
					Op:    Gt,
					Left:  VariableExpr{Name: "x"},
					Right: LiteralExpr{Value: IntLiteral{Value: 5}},
				},
				Body: LiteralExpr{Value: StringLiteral{Value: "big"}},
			},
			{
				Pattern: VariablePattern{Name: "x"},
				Body:    LiteralExpr{Value: StringLiteral{Value: "small"}},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "small", result)
}

func TestMatchExpr_ObjectPattern(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("person", map[string]interface{}{
		"name": "Alice",
		"age":  int64(30),
	})

	// match person { {name, age} => name }
	expr := MatchExpr{
		Value: VariableExpr{Name: "person"},
		Cases: []MatchCase{
			{
				Pattern: ObjectPattern{
					Fields: []ObjectPatternField{
						{Key: "name"},
						{Key: "age"},
					},
				},
				Body: VariableExpr{Name: "name"},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "Alice", result)
}

func TestMatchExpr_ObjectPattern_NoMatch(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("person", map[string]interface{}{
		"name": "Alice",
	})

	// match person { {name, age} => name, _ => "no match" }
	// Should fall through to wildcard since "age" is missing
	expr := MatchExpr{
		Value: VariableExpr{Name: "person"},
		Cases: []MatchCase{
			{
				Pattern: ObjectPattern{
					Fields: []ObjectPatternField{
						{Key: "name"},
						{Key: "age"},
					},
				},
				Body: VariableExpr{Name: "name"},
			},
			{
				Pattern: WildcardPattern{},
				Body:    LiteralExpr{Value: StringLiteral{Value: "no match"}},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "no match", result)
}

func TestMatchExpr_ArrayPattern(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("numbers", []interface{}{int64(1), int64(2), int64(3)})

	// match numbers { [first, second, third] => first + second + third }
	expr := MatchExpr{
		Value: VariableExpr{Name: "numbers"},
		Cases: []MatchCase{
			{
				Pattern: ArrayPattern{
					Elements: []Pattern{
						VariablePattern{Name: "first"},
						VariablePattern{Name: "second"},
						VariablePattern{Name: "third"},
					},
				},
				Body: BinaryOpExpr{
					Op: Add,
					Left: BinaryOpExpr{
						Op:    Add,
						Left:  VariableExpr{Name: "first"},
						Right: VariableExpr{Name: "second"},
					},
					Right: VariableExpr{Name: "third"},
				},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, int64(6), result)
}

func TestMatchExpr_ArrayPattern_WithRest(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("numbers", []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5)})

	rest := "rest"
	// match numbers { [first, ...rest] => first }
	expr := MatchExpr{
		Value: VariableExpr{Name: "numbers"},
		Cases: []MatchCase{
			{
				Pattern: ArrayPattern{
					Elements: []Pattern{
						VariablePattern{Name: "first"},
					},
					Rest: &rest,
				},
				Body: VariableExpr{Name: "first"},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result)
}

func TestMatchExpr_BooleanPattern(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// match true { true => "yes", false => "no" }
	expr := MatchExpr{
		Value: LiteralExpr{Value: BoolLiteral{Value: true}},
		Cases: []MatchCase{
			{
				Pattern: LiteralPattern{Value: BoolLiteral{Value: true}},
				Body:    LiteralExpr{Value: StringLiteral{Value: "yes"}},
			},
			{
				Pattern: LiteralPattern{Value: BoolLiteral{Value: false}},
				Body:    LiteralExpr{Value: StringLiteral{Value: "no"}},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Equal(t, "yes", result)
}

func TestMatchExpr_NoMatch_ReturnsNil(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// match 5 { 1 => "one", 2 => "two" }
	// No wildcard, no match, returns nil
	expr := MatchExpr{
		Value: LiteralExpr{Value: IntLiteral{Value: 5}},
		Cases: []MatchCase{
			{
				Pattern: LiteralPattern{Value: IntLiteral{Value: 1}},
				Body:    LiteralExpr{Value: StringLiteral{Value: "one"}},
			},
			{
				Pattern: LiteralPattern{Value: IntLiteral{Value: 2}},
				Body:    LiteralExpr{Value: StringLiteral{Value: "two"}},
			},
		},
	}

	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)
	assert.Nil(t, result)
}
