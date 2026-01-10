package interpreter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example: Simple Hello World Route
func Example_helloWorld() {
	interp := NewInterpreter()

	// @ GET /hello
	//   > {text: "Hello, World!", timestamp: 1234567890}
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
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
	// Output: Hello, World!
}

// Example: Greeting with Path Parameter
func Example_greetWithParam() {
	interp := NewInterpreter()

	// @ GET /greet/:name
	//   $ greeting = "Hello, " + name + "!"
	//   > greeting
	route := &Route{
		Path:   "/greet/:name",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "greeting",
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
			ReturnStatement{
				Value: VariableExpr{Name: "greeting"},
			},
		},
	}

	params := map[string]string{"name": "Alice"}
	result, err := interp.ExecuteRouteSimple(route, params)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
	// Output: Hello, Alice!
}

// Example: Arithmetic Calculation
func Example_arithmetic() {
	interp := NewInterpreter()

	// @ GET /calculate
	//   $ a = 10
	//   $ b = 20
	//   $ sum = a + b
	//   $ product = sum * 2
	//   > product
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
			AssignStatement{
				Target: "product",
				Value: BinaryOpExpr{
					Op:    Mul,
					Left:  VariableExpr{Name: "sum"},
					Right: LiteralExpr{Value: IntLiteral{Value: 2}},
				},
			},
			ReturnStatement{
				Value: VariableExpr{Name: "product"},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
	// Output: 60
}

// Test: Complete Hello World Example from examples/hello-world/main.glyph
func TestInterpreter_HelloWorldExample(t *testing.T) {
	interp := NewInterpreter()

	// Define the Message type
	messageType := TypeDef{
		Name: "Message",
		Fields: []Field{
			{Name: "text", TypeAnnotation: StringType{}, Required: true},
			{Name: "timestamp", TypeAnnotation: IntType{}, Required: false},
		},
	}

	// Load module with type definition
	module := Module{
		Items: []Item{&messageType},
	}
	err := interp.LoadModule(module)
	require.NoError(t, err)

	// Route 1: @ GET /hello
	//   > {text: "Hello, World!", timestamp: 1234567890}
	route1 := &Route{
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

	result1, err := interp.ExecuteRouteSimple(route1, nil)
	require.NoError(t, err)
	obj1, ok := result1.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Hello, World!", obj1["text"])
	assert.Equal(t, int64(1234567890), obj1["timestamp"])

	// Route 2: @ GET /greet/:name -> Message
	//   $ message = {
	//     text: "Hello, " + name + "!",
	//     timestamp: time.now()
	//   }
	//   > message
	route2 := &Route{
		Path:       "/greet/:name",
		Method:     Get,
		ReturnType: NamedType{Name: "Message"},
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
	result2, err := interp.ExecuteRouteSimple(route2, params)
	require.NoError(t, err)
	obj2, ok := result2.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Hello, World!", obj2["text"])
	// time.now returns current Unix timestamp
	timestamp2, ok := obj2["timestamp"].(int64)
	assert.True(t, ok, "timestamp should be int64")
	assert.Greater(t, timestamp2, int64(1700000000), "timestamp should be after 2023")
}

// Test: Conditional Logic
func TestInterpreter_ConditionalExample(t *testing.T) {
	interp := NewInterpreter()

	// @ GET /check/:age
	//   $ age_num = parse_int(age)
	//   if age_num >= 18:
	//     > "adult"
	//   else:
	//     > "minor"
	route := &Route{
		Path:   "/check/:age",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "age",
				Value:  LiteralExpr{Value: IntLiteral{Value: 25}},
			},
			IfStatement{
				Condition: BinaryOpExpr{
					Op:    Ge,
					Left:  VariableExpr{Name: "age"},
					Right: LiteralExpr{Value: IntLiteral{Value: 18}},
				},
				ThenBlock: []Statement{
					ReturnStatement{
						Value: LiteralExpr{Value: StringLiteral{Value: "adult"}},
					},
				},
				ElseBlock: []Statement{
					ReturnStatement{
						Value: LiteralExpr{Value: StringLiteral{Value: "minor"}},
					},
				},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)
	require.NoError(t, err)
	assert.Equal(t, "adult", result)
}

// Test: User-Defined Function with Module
func TestInterpreter_UserDefinedFunctionExample(t *testing.T) {
	interp := NewInterpreter()

	// Define a function: fn add(a: int, b: int) -> int
	addFn := Function{
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

	// Load module with function
	module := Module{
		Items: []Item{&addFn},
	}
	err := interp.LoadModule(module)
	require.NoError(t, err)

	// @ GET /sum/:x/:y
	//   $ result = add(x, y)
	//   > result
	route := &Route{
		Path:   "/sum/:x/:y",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "x_num",
				Value:  LiteralExpr{Value: IntLiteral{Value: 5}},
			},
			AssignStatement{
				Target: "y_num",
				Value:  LiteralExpr{Value: IntLiteral{Value: 7}},
			},
			AssignStatement{
				Target: "result",
				Value: FunctionCallExpr{
					Name: "add",
					Args: []Expr{
						VariableExpr{Name: "x_num"},
						VariableExpr{Name: "y_num"},
					},
				},
			},
			ReturnStatement{
				Value: VariableExpr{Name: "result"},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(12), result)
}

// Test: Complex Expression Evaluation
func TestInterpreter_ComplexExpressionExample(t *testing.T) {
	interp := NewInterpreter()

	// @ GET /complex
	//   $ result = (10 + 20) * 3 - 5
	//   > result
	route := &Route{
		Path:   "/complex",
		Method: Get,
		Body: []Statement{
			AssignStatement{
				Target: "result",
				Value: BinaryOpExpr{
					Op: Sub,
					Left: BinaryOpExpr{
						Op: Mul,
						Left: BinaryOpExpr{
							Op:    Add,
							Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
							Right: LiteralExpr{Value: IntLiteral{Value: 20}},
						},
						Right: LiteralExpr{Value: IntLiteral{Value: 3}},
					},
					Right: LiteralExpr{Value: IntLiteral{Value: 5}},
				},
			},
			ReturnStatement{
				Value: VariableExpr{Name: "result"},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)
	require.NoError(t, err)
	// (10 + 20) * 3 - 5 = 30 * 3 - 5 = 90 - 5 = 85
	assert.Equal(t, int64(85), result)
}

// Test: Comparison Operations
func TestInterpreter_ComparisonExample(t *testing.T) {
	tests := []struct {
		name     string
		op       BinOp
		left     int64
		right    int64
		expected bool
	}{
		{"Equal", Eq, 10, 10, true},
		{"Not Equal", Ne, 10, 20, true},
		{"Less Than", Lt, 10, 20, true},
		{"Less Than or Equal", Le, 10, 10, true},
		{"Greater Than", Gt, 20, 10, true},
		{"Greater Than or Equal", Ge, 10, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewInterpreter()

			route := &Route{
				Path:   "/compare",
				Method: Get,
				Body: []Statement{
					ReturnStatement{
						Value: BinaryOpExpr{
							Op:    tt.op,
							Left:  LiteralExpr{Value: IntLiteral{Value: tt.left}},
							Right: LiteralExpr{Value: IntLiteral{Value: tt.right}},
						},
					},
				},
			}

			result, err := interp.ExecuteRouteSimple(route, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test: Field Access Example
func TestInterpreter_FieldAccessExample(t *testing.T) {
	interp := NewInterpreter()

	// @ GET /user-info
	//   $ user = {name: "Alice", age: 30}
	//   $ name = user.name
	//   > name
	route := &Route{
		Path:   "/user-info",
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
				Target: "name",
				Value: FieldAccessExpr{
					Object: VariableExpr{Name: "user"},
					Field:  "name",
				},
			},
			ReturnStatement{
				Value: VariableExpr{Name: "name"},
			},
		},
	}

	result, err := interp.ExecuteRouteSimple(route, nil)
	require.NoError(t, err)
	assert.Equal(t, "Alice", result)
}
