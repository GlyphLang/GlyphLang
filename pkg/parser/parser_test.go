package parser

import (
	"fmt"
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexer_SimpleTokens(t *testing.T) {
	input := `@ : $ + - * /`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()

	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	expected := []TokenType{AT, COLON, DOLLAR, PLUS, MINUS, STAR, SLASH, EOF}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, expectedType := range expected {
		if tokens[i].Type != expectedType {
			t.Errorf("token %d: expected %s, got %s", i, expectedType, tokens[i].Type)
		}
	}
}

func TestLexer_String(t *testing.T) {
	input := `"hello world"`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()

	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	if len(tokens) != 2 { // STRING + EOF
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != STRING {
		t.Errorf("expected STRING, got %s", tokens[0].Type)
	}

	if tokens[0].Literal != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", tokens[0].Literal)
	}
}

func TestLexer_Number(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
		literal  string
	}{
		{"123", INTEGER, "123"},
		{"123.45", FLOAT, "123.45"},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		tokens, err := lexer.Tokenize()

		if err != nil {
			t.Fatalf("lexer error for '%s': %v", tt.input, err)
		}

		if tokens[0].Type != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tokens[0].Type)
		}

		if tokens[0].Literal != tt.literal {
			t.Errorf("expected '%s', got '%s'", tt.literal, tokens[0].Literal)
		}
	}
}

func TestLexer_Path(t *testing.T) {
	input := `/api/users/:id`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()

	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	if tokens[0].Type != IDENT {
		t.Errorf("expected IDENT, got %s", tokens[0].Type)
	}

	if tokens[0].Literal != "/api/users/:id" {
		t.Errorf("expected '/api/users/:id', got '%s'", tokens[0].Literal)
	}
}

func TestLexer_Division(t *testing.T) {
	input := `100 / min`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()

	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	// Should be: INTEGER SLASH IDENT EOF
	expected := []TokenType{INTEGER, SLASH, IDENT, EOF}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, expectedType := range expected {
		if tokens[i].Type != expectedType {
			t.Errorf("token %d: expected %s, got %s", i, expectedType, tokens[i].Type)
		}
	}
}

func TestParser_SimpleRoute(t *testing.T) {
	source := `@ route /hello
  > {message: "Hello, World!"}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(module.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(module.Items))
	}

	route, ok := module.Items[0].(*interpreter.Route)
	if !ok {
		t.Fatalf("expected Route, got %T", module.Items[0])
	}

	if route.Path != "/hello" {
		t.Errorf("expected path '/hello', got '%s'", route.Path)
	}

	if route.Method != interpreter.Get {
		t.Errorf("expected GET method, got %s", route.Method)
	}

	if len(route.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(route.Body))
	}
}

func TestParser_RouteWithPathParam(t *testing.T) {
	source := `@ route /users/:id [GET]
  > {id: id}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(module.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(module.Items))
	}

	route, ok := module.Items[0].(*interpreter.Route)
	if !ok {
		t.Fatalf("expected Route, got %T", module.Items[0])
	}

	if route.Path != "/users/:id" {
		t.Errorf("expected path '/users/:id', got '%s'", route.Path)
	}

	if route.Method != interpreter.Get {
		t.Errorf("expected GET method, got %s", route.Method)
	}
}

func TestParser_TypeDef(t *testing.T) {
	source := `: User {
  id: int!
  name: str!
  email: str
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(module.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(module.Items))
	}

	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	if !ok {
		t.Fatalf("expected TypeDef, got %T", module.Items[0])
	}

	if typeDef.Name != "User" {
		t.Errorf("expected name 'User', got '%s'", typeDef.Name)
	}

	if len(typeDef.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(typeDef.Fields))
	}

	// Check first field
	if typeDef.Fields[0].Name != "id" {
		t.Errorf("expected field name 'id', got '%s'", typeDef.Fields[0].Name)
	}

	if !typeDef.Fields[0].Required {
		t.Error("expected field 'id' to be required")
	}
}

func TestParser_RouteWithMiddleware(t *testing.T) {
	source := `@ route /protected
  + auth(jwt)
  + ratelimit(100/min)
  > {status: "ok"}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(module.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(module.Items))
	}

	route, ok := module.Items[0].(*interpreter.Route)
	if !ok {
		t.Fatalf("expected Route, got %T", module.Items[0])
	}

	if route.Auth == nil {
		t.Fatal("expected auth config, got nil")
	}

	if route.Auth.AuthType != "jwt" {
		t.Errorf("expected auth type 'jwt', got '%s'", route.Auth.AuthType)
	}

	if route.RateLimit == nil {
		t.Fatal("expected rate limit, got nil")
	}

	if route.RateLimit.Requests != 100 {
		t.Errorf("expected 100 requests, got %d", route.RateLimit.Requests)
	}

	if route.RateLimit.Window != "min" {
		t.Errorf("expected window 'min', got '%s'", route.RateLimit.Window)
	}
}

func TestParser_Assignment(t *testing.T) {
	source := `@ route /test
  $ x = 42
  > {value: x}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route, ok := module.Items[0].(*interpreter.Route)
	if !ok {
		t.Fatalf("expected Route, got %T", module.Items[0])
	}

	if len(route.Body) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(route.Body))
	}

	assign, ok := route.Body[0].(interpreter.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}

	if assign.Target != "x" {
		t.Errorf("expected target 'x', got '%s'", assign.Target)
	}
}

func TestParser_BinaryOp(t *testing.T) {
	source := `@ route /calc
  $ result = 10 + 20 * 2
  > {result: result}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route, ok := module.Items[0].(*interpreter.Route)
	if !ok {
		t.Fatalf("expected Route, got %T", module.Items[0])
	}

	assign, ok := route.Body[0].(interpreter.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}

	// Should be: 10 + (20 * 2) due to precedence
	binOp, ok := assign.Value.(interpreter.BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr, got %T", assign.Value)
	}

	if binOp.Op != interpreter.Add {
		t.Errorf("expected Add operator, got %s", binOp.Op)
	}
}

// Test all HTTP methods
func TestParser_AllHTTPMethods(t *testing.T) {
	tests := []struct {
		name           string
		methodStr      string
		expectedMethod interpreter.HttpMethod
	}{
		{"GET", "[GET]", interpreter.Get},
		{"POST", "[POST]", interpreter.Post},
		{"PUT", "[PUT]", interpreter.Put},
		{"DELETE", "[DELETE]", interpreter.Delete},
		{"PATCH", "[PATCH]", interpreter.Patch},
		{"default GET", "", interpreter.Get}, // No method specified defaults to GET
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := fmt.Sprintf("@ route /test %s\n  > {status: \"ok\"}", tt.methodStr)
			lexer := NewLexer(source)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			module, err := parser.Parse()
			require.NoError(t, err)

			route, ok := module.Items[0].(*interpreter.Route)
			require.True(t, ok)
			assert.Equal(t, tt.expectedMethod, route.Method)
		})
	}
}

// Test complex path parameters
func TestParser_ComplexPathParameters(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedPath string
	}{
		{"single param", "/users/:id", "/users/:id"},
		{"multiple params", "/posts/:postId/comments/:commentId", "/posts/:postId/comments/:commentId"},
		{"three params", "/api/:version/users/:userId/posts/:postId", "/api/:version/users/:userId/posts/:postId"},
		{"param at end", "/users/:id", "/users/:id"},
		{"param at start", "/:lang/home", "/:lang/home"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := fmt.Sprintf("@ route %s\n  > {ok: true}", tt.path)
			lexer := NewLexer(source)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			module, err := parser.Parse()
			require.NoError(t, err)

			route, ok := module.Items[0].(*interpreter.Route)
			require.True(t, ok)
			assert.Equal(t, tt.expectedPath, route.Path)
		})
	}
}

// Test return types
func TestParser_ReturnTypes(t *testing.T) {
	tests := []struct {
		name          string
		returnTypeStr string
		expectedType  interpreter.Type
	}{
		{"named type", "-> User", interpreter.NamedType{Name: "User"}},
		{"int type", "-> int", interpreter.IntType{}},
		{"str type", "-> str", interpreter.StringType{}},
		{"bool type", "-> bool", interpreter.BoolType{}},
		{"float type", "-> float", interpreter.FloatType{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := fmt.Sprintf("@ route /test %s\n  > {ok: true}", tt.returnTypeStr)
			lexer := NewLexer(source)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			parser := NewParser(tokens)
			module, err := parser.Parse()
			require.NoError(t, err)

			route, ok := module.Items[0].(*interpreter.Route)
			require.True(t, ok)
			assert.Equal(t, tt.expectedType, route.ReturnType)
		})
	}
}

// Test multiple middlewares
func TestParser_MultipleMiddlewares(t *testing.T) {
	source := `@ route /api/admin
  + auth(jwt, role: admin)
  + ratelimit(50/min)
  > {status: "ok"}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	// Check auth
	require.NotNil(t, route.Auth)
	assert.Equal(t, "jwt", route.Auth.AuthType)

	// Check rate limit
	require.NotNil(t, route.RateLimit)
	assert.Equal(t, uint32(50), route.RateLimit.Requests)
	assert.Equal(t, "min", route.RateLimit.Window)
}

// Test complex expressions
func TestParser_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "nested arithmetic",
			source: `@ route /calc
  $ result = (10 + 20) * (5 - 2)
  > {result: result}`,
		},
		{
			name: "string concatenation",
			source: `@ route /concat
  $ full = "Hello, " + name + "!"
  > {text: full}`,
		},
		{
			name: "comparison operators",
			source: `@ route /compare
  $ check = age >= 18
  > {adult: check}`,
		},
		{
			name: "field access",
			source: `@ route /field
  $ name = user.profile.name
  > {name: name}`,
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
			assert.Len(t, module.Items, 1)
		})
	}
}

// Test object literals
func TestParser_ObjectLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple object",
			source: `@ route /obj
  > {name: "test", age: 30}`,
		},
		{
			name: "nested object",
			source: `@ route /nested
  > {user: {name: "test", profile: {bio: "dev"}}}`,
		},
		{
			name: "object with expressions",
			source: `@ route /expr
  > {total: price * quantity, discount: true}`,
		},
		{
			name: "object with variables",
			source: `@ route /vars
  $ x = 42
  > {value: x, doubled: x * 2}`,
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
			assert.Len(t, module.Items, 1)
		})
	}
}

// Test function calls
func TestParser_FunctionCalls(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "no args",
			source: `@ route /time
  > {now: time.now()}`,
		},
		{
			name: "with args",
			source: `@ route /add
  > {result: add(10, 20)}`,
		},
		{
			name: "nested calls",
			source: `@ route /nested
  > {result: max(min(x, 10), 5)}`,
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
			assert.Len(t, module.Items, 1)
		})
	}
}

// Test multiple routes in one file
func TestParser_MultipleRoutes(t *testing.T) {
	source := `@ route /hello
  > {message: "Hello"}

@ route /goodbye
  > {message: "Goodbye"}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	assert.Len(t, module.Items, 2)

	route1, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)
	assert.Equal(t, "/hello", route1.Path)

	route2, ok := module.Items[1].(*interpreter.Route)
	require.True(t, ok)
	assert.Equal(t, "/goodbye", route2.Path)
}

// Test type definitions with different field types
func TestParser_TypeDefFieldTypes(t *testing.T) {
	source := `: ComplexType {
  id: int!
  name: str!
  score: float
  active: bool!
  tags: List[str]
  metadata: Map[str, str]
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)
	assert.Equal(t, "ComplexType", typeDef.Name)
	assert.GreaterOrEqual(t, len(typeDef.Fields), 4)
}

// Test error cases
func TestParser_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		shouldError bool
	}{
		{
			name:        "missing route path",
			source:      "@ route\n  > {ok: true}",
			shouldError: true,
		},
		{
			name:        "unclosed brace",
			source:      ": User {\n  id: int!",
			shouldError: true,
		},
		{
			name:        "invalid token start",
			source:      "invalid start",
			shouldError: true,
		},
		{
			name:        "empty route body",
			source:      "@ route /test",
			shouldError: false, // Empty body might be allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source)
			tokens, err := lexer.Tokenize()
			if err != nil {
				if tt.shouldError {
					return
				}
				t.Fatalf("lexer error: %v", err)
			}

			parser := NewParser(tokens)
			_, err = parser.Parse()

			if tt.shouldError {
				assert.Error(t, err, "expected parser error")
			} else {
				assert.NoError(t, err, "unexpected parser error")
			}
		})
	}
}

// Test if statements
func TestParser_IfStatements(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "if with else",
			source: `@ route /check/:age
  if age >= 18 {
    > {status: "adult"}
  } else {
    > {status: "minor"}
  }`,
		},
		{
			name: "if without else",
			source: `@ route /check
  if true {
    > {ok: true}
  }`,
		},
		{
			name: "if with multiple statements",
			source: `@ route /check
  if x > 10 {
    $ result = "high"
    > {result: result}
  } else {
    $ result = "low"
    > {result: result}
  }`,
		},
		{
			name: "nested if",
			source: `@ route /nested
  if x > 0 {
    if y > 0 {
      > {quadrant: 1}
    } else {
      > {quadrant: 4}
    }
  } else {
    > {quadrant: 2}
  }`,
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

			route, ok := module.Items[0].(*interpreter.Route)
			require.True(t, ok)
			assert.Greater(t, len(route.Body), 0)

			// First statement should be an if statement
			_, ok = route.Body[0].(interpreter.IfStatement)
			assert.True(t, ok, "expected first statement to be IfStatement")
		})
	}
}

// Test logical operators
func TestParser_LogicalOperators(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "AND operator",
			source: `@ route /and
  $ result = true && false
  > {result: result}`,
		},
		{
			name: "OR operator",
			source: `@ route /or
  $ result = true || false
  > {result: result}`,
		},
		{
			name: "complex logical expression",
			source: `@ route /complex
  $ result = (x > 10 && y < 20) || z == 0
  > {result: result}`,
		},
		{
			name: "if with logical operators",
			source: `@ route /check
  if age >= 18 && hasLicense {
    > {canDrive: true}
  } else {
    > {canDrive: false}
  }`,
		},
		{
			name: "multiple AND/OR",
			source: `@ route /multi
  $ result = a && b && c || d || e
  > {result: result}`,
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

			route, ok := module.Items[0].(*interpreter.Route)
			require.True(t, ok)
			assert.Greater(t, len(route.Body), 0)
		})
	}
}

// Test array literals
func TestParser_ArrayLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "empty array",
			source: `@ route /empty
  $ arr = []
  > {array: arr}`,
		},
		{
			name: "integer array",
			source: `@ route /ints
  $ numbers = [1, 2, 3, 4, 5]
  > {numbers: numbers}`,
		},
		{
			name: "string array",
			source: `@ route /strings
  $ names = ["Alice", "Bob", "Charlie"]
  > {names: names}`,
		},
		{
			name: "mixed array",
			source: `@ route /mixed
  $ items = [1, "two", true, 4.5]
  > {items: items}`,
		},
		{
			name: "nested arrays",
			source: `@ route /nested
  $ matrix = [[1, 2], [3, 4]]
  > {matrix: matrix}`,
		},
		{
			name: "array with expressions",
			source: `@ route /expr
  $ values = [1 + 1, x * 2, y]
  > {values: values}`,
		},
		{
			name: "array of objects",
			source: `@ route /objects
  $ users = [
    {name: "Alice", age: 30},
    {name: "Bob", age: 25}
  ]
  > {users: users}`,
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

			route, ok := module.Items[0].(*interpreter.Route)
			require.True(t, ok)
			assert.Greater(t, len(route.Body), 0)

			// First statement should be assignment
			assign, ok := route.Body[0].(interpreter.AssignStatement)
			assert.True(t, ok, "expected AssignStatement")

			// Value should be an array expression
			_, ok = assign.Value.(interpreter.ArrayExpr)
			assert.True(t, ok, "expected ArrayExpr")
		})
	}
}

// Test complete hello world example
func TestParser_HelloWorldExample(t *testing.T) {
	source := `: Message {
  text: str!
  timestamp: int
}

@ route /hello
  > {text: "Hello, World!", timestamp: 1234567890}

@ route /greet/:name -> Message
  $ message = {
    text: "Hello, " + name + "!",
    timestamp: time.now()
  }
  > message`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	// Should have 1 type def and 2 routes
	assert.Len(t, module.Items, 3)

	// Check type def
	_, ok := module.Items[0].(*interpreter.TypeDef)
	assert.True(t, ok, "first item should be TypeDef")

	// Check routes
	route1, ok := module.Items[1].(*interpreter.Route)
	assert.True(t, ok, "second item should be Route")
	assert.Equal(t, "/hello", route1.Path)

	route2, ok := module.Items[2].(*interpreter.Route)
	assert.True(t, ok, "third item should be Route")
	assert.Equal(t, "/greet/:name", route2.Path)
}

// Benchmark parser performance
func BenchmarkParser_SimpleRoute(b *testing.B) {
	source := `@ route /hello
  > {message: "Hello, World!"}`

	for i := 0; i < b.N; i++ {
		lexer := NewLexer(source)
		tokens, _ := lexer.Tokenize()
		parser := NewParser(tokens)
		_, _ = parser.Parse()
	}
}

func BenchmarkParser_ComplexModule(b *testing.B) {
	source := `: User {
  id: int!
  name: str!
}

@ route /api/users/:id [GET]
  + auth(jwt)
  + ratelimit(100/min)
  $ user = db.users.get(id)
  > user`

	for i := 0; i < b.N; i++ {
		lexer := NewLexer(source)
		tokens, _ := lexer.Tokenize()
		parser := NewParser(tokens)
		_, _ = parser.Parse()
	}
}

// Test While Loops

func TestParser_SimpleWhileLoop(t *testing.T) {
	source := `@ route /test
  $ i = 0
  while i < 5 {
    $ i = i + 1
  }
  > i`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	require.Len(t, route.Body, 3)

	// Check second statement is while loop
	whileStmt, ok := route.Body[1].(interpreter.WhileStatement)
	require.True(t, ok)

	// Check condition
	condExpr, ok := whileStmt.Condition.(interpreter.BinaryOpExpr)
	require.True(t, ok)
	assert.Equal(t, interpreter.Lt, condExpr.Op)

	// Check body has one statement
	require.Len(t, whileStmt.Body, 1)
}

func TestParser_NestedWhileLoops(t *testing.T) {
	source := `@ route /test
  $ i = 0
  while i < 3 {
    $ j = 0
    while j < 2 {
      $ j = j + 1
    }
    $ i = i + 1
  }
  > i`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	// Check outer while loop
	outerWhile, ok := route.Body[1].(interpreter.WhileStatement)
	require.True(t, ok)
	require.Len(t, outerWhile.Body, 3)

	// Check inner while loop
	innerWhile, ok := outerWhile.Body[1].(interpreter.WhileStatement)
	require.True(t, ok)
	require.Len(t, innerWhile.Body, 1)
}

func TestParser_WhileWithComplexCondition(t *testing.T) {
	source := `@ route /test
  $ i = 0
  $ max = 10
  while i < max && i < 100 {
    $ i = i + 1
  }
  > i`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	whileStmt, ok := route.Body[2].(interpreter.WhileStatement)
	require.True(t, ok)

	// Check condition is AND operation
	condExpr, ok := whileStmt.Condition.(interpreter.BinaryOpExpr)
	require.True(t, ok)
	assert.Equal(t, interpreter.And, condExpr.Op)
}

func TestParser_WhileWithMultipleStatements(t *testing.T) {
	source := `@ route /test
  $ sum = 0
  $ i = 0
  while i < 5 {
    $ sum = sum + i
    $ i = i + 1
    $ temp = i * 2
  }
  > sum`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	whileStmt, ok := route.Body[2].(interpreter.WhileStatement)
	require.True(t, ok)

	// Check body has three statements
	require.Len(t, whileStmt.Body, 3)
}

func TestParser_WhileWithIfStatement(t *testing.T) {
	source := `@ route /test
  $ sum = 0
  $ i = 0
  while i < 10 {
    if i < 5 {
      $ sum = sum + i
    }
    $ i = i + 1
  }
  > sum`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	whileStmt, ok := route.Body[2].(interpreter.WhileStatement)
	require.True(t, ok)

	// Check body has two statements: if and assignment
	require.Len(t, whileStmt.Body, 2)

	// Check first statement is if
	_, ok = whileStmt.Body[0].(interpreter.IfStatement)
	require.True(t, ok)
}

// Test for loop parsing
func TestParser_ForLoop_SimpleArray(t *testing.T) {
	source := `@ route /test
  $ items = [1, 2, 3]
  for item in items {
    $ x = item + 1
  }
  > {ok: true}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	// Check for statement
	require.Len(t, route.Body, 3) // assignment, for loop, return
	forStmt, ok := route.Body[1].(interpreter.ForStatement)
	require.True(t, ok)

	// Check loop variables
	assert.Equal(t, "", forStmt.KeyVar)
	assert.Equal(t, "item", forStmt.ValueVar)

	// Check iterable is a variable
	iterableVar, ok := forStmt.Iterable.(interpreter.VariableExpr)
	require.True(t, ok)
	assert.Equal(t, "items", iterableVar.Name)

	// Check body has one statement
	require.Len(t, forStmt.Body, 1)
}

func TestParser_ForLoop_WithIndex(t *testing.T) {
	source := `@ route /test
  for index, value in array {
    $ result = index + value
  }
  > {ok: true}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	// Check for statement
	forStmt, ok := route.Body[0].(interpreter.ForStatement)
	require.True(t, ok)

	// Check loop variables
	assert.Equal(t, "index", forStmt.KeyVar)
	assert.Equal(t, "value", forStmt.ValueVar)

	// Check iterable
	iterableVar, ok := forStmt.Iterable.(interpreter.VariableExpr)
	require.True(t, ok)
	assert.Equal(t, "array", iterableVar.Name)
}

func TestParser_ForLoop_ArrayLiteral(t *testing.T) {
	source := `@ route /test
  for item in [1, 2, 3] {
    $ x = item
  }
  > {ok: true}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	forStmt, ok := route.Body[0].(interpreter.ForStatement)
	require.True(t, ok)

	// Check iterable is an array literal
	arrayExpr, ok := forStmt.Iterable.(interpreter.ArrayExpr)
	require.True(t, ok)
	assert.Len(t, arrayExpr.Elements, 3)
}

func TestParser_ForLoop_ObjectIteration(t *testing.T) {
	source := `@ route /test
  for key, value in {name: "Alice", age: 30} {
    $ output = key + value
  }
  > {ok: true}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	forStmt, ok := route.Body[0].(interpreter.ForStatement)
	require.True(t, ok)

	// Check loop variables
	assert.Equal(t, "key", forStmt.KeyVar)
	assert.Equal(t, "value", forStmt.ValueVar)

	// Check iterable is an object
	objExpr, ok := forStmt.Iterable.(interpreter.ObjectExpr)
	require.True(t, ok)
	assert.Len(t, objExpr.Fields, 2)
}

func TestParser_ForLoop_Nested(t *testing.T) {
	source := `@ route /test
  for row in rows {
    for col in row {
      $ cell = col
    }
  }
  > {ok: true}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	// Outer for loop
	outerFor, ok := route.Body[0].(interpreter.ForStatement)
	require.True(t, ok)
	assert.Equal(t, "row", outerFor.ValueVar)

	// Inner for loop
	require.Len(t, outerFor.Body, 1)
	innerFor, ok := outerFor.Body[0].(interpreter.ForStatement)
	require.True(t, ok)
	assert.Equal(t, "col", innerFor.ValueVar)

	// Check inner loop has one statement
	require.Len(t, innerFor.Body, 1)
}

func TestParser_SwitchStatement_IntegerCases(t *testing.T) {
	input := `
@route /test [GET]
  $ status = 1
  switch status {
    case 1 {
      > "one"
    }
    case 2 {
      > "two"
    }
    default {
      > "other"
    }
  }
`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	require.Len(t, route.Body, 2)

	switchStmt, ok := route.Body[1].(interpreter.SwitchStatement)
	require.True(t, ok)

	// Check switch value is a variable reference
	varExpr, ok := switchStmt.Value.(interpreter.VariableExpr)
	require.True(t, ok)
	assert.Equal(t, "status", varExpr.Name)

	// Check cases
	require.Len(t, switchStmt.Cases, 2)

	// Case 1
	case1Literal, ok := switchStmt.Cases[0].Value.(interpreter.LiteralExpr)
	require.True(t, ok)
	intLit1, ok := case1Literal.Value.(interpreter.IntLiteral)
	require.True(t, ok)
	assert.Equal(t, int64(1), intLit1.Value)
	require.Len(t, switchStmt.Cases[0].Body, 1)

	// Case 2
	case2Literal, ok := switchStmt.Cases[1].Value.(interpreter.LiteralExpr)
	require.True(t, ok)
	intLit2, ok := case2Literal.Value.(interpreter.IntLiteral)
	require.True(t, ok)
	assert.Equal(t, int64(2), intLit2.Value)
	require.Len(t, switchStmt.Cases[1].Body, 1)

	// Default
	require.Len(t, switchStmt.Default, 1)
}

func TestParser_SwitchStatement_StringCases(t *testing.T) {
	input := `
@route /api/status [GET]
  switch status {
    case "pending" {
      > {message: "Order is pending"}
    }
    case "shipped" {
      > {message: "Order has shipped"}
    }
    default {
      > {message: "Unknown status"}
    }
  }
`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	switchStmt, ok := route.Body[0].(interpreter.SwitchStatement)
	require.True(t, ok)

	// Check string cases
	require.Len(t, switchStmt.Cases, 2)

	case1Literal, ok := switchStmt.Cases[0].Value.(interpreter.LiteralExpr)
	require.True(t, ok)
	strLit1, ok := case1Literal.Value.(interpreter.StringLiteral)
	require.True(t, ok)
	assert.Equal(t, "pending", strLit1.Value)

	case2Literal, ok := switchStmt.Cases[1].Value.(interpreter.LiteralExpr)
	require.True(t, ok)
	strLit2, ok := case2Literal.Value.(interpreter.StringLiteral)
	require.True(t, ok)
	assert.Equal(t, "shipped", strLit2.Value)
}

func TestParser_SwitchStatement_NoDefault(t *testing.T) {
	input := `
@route /test [GET]
  switch x {
    case 1 {
      > "one"
    }
  }
`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	route := module.Items[0].(*interpreter.Route)
	switchStmt := route.Body[0].(interpreter.SwitchStatement)

	require.Len(t, switchStmt.Cases, 1)
	assert.Nil(t, switchStmt.Default)
}

func TestParser_SwitchStatement_MultipleCases(t *testing.T) {
	input := `
@route /test [GET]
  switch value {
    case 1 {
      > "one"
    }
    case 2 {
      > "two"
    }
    case 3 {
      > "three"
    }
    case 4 {
      > "four"
    }
    case 5 {
      > "five"
    }
  }
`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	route := module.Items[0].(*interpreter.Route)
	switchStmt := route.Body[0].(interpreter.SwitchStatement)

	require.Len(t, switchStmt.Cases, 5)
}

func TestParser_SwitchStatement_NestedStatements(t *testing.T) {
	input := `
@route /test [GET]
  switch x {
    case 1 {
      $ y = 10
      if y == 10 {
        > "matched"
      }
    }
    default {
      $ z = 20
      > z
    }
  }
`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	route := module.Items[0].(*interpreter.Route)
	switchStmt := route.Body[0].(interpreter.SwitchStatement)

	// Check case body has multiple statements
	require.Len(t, switchStmt.Cases[0].Body, 2)

	// First should be assignment
	_, ok := switchStmt.Cases[0].Body[0].(interpreter.AssignStatement)
	require.True(t, ok)

	// Second should be if statement
	_, ok = switchStmt.Cases[0].Body[1].(interpreter.IfStatement)
	require.True(t, ok)

	// Check default has multiple statements
	require.Len(t, switchStmt.Default, 2)
}
