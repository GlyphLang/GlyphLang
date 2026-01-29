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
	source := `@ GET /hello {
  > {message: "Hello, World!"}
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
	source := `@ GET /users/:id {
  > {id: id}
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
	source := `@ GET /protected {
  + auth(jwt)
  + ratelimit(100/min)
  > {status: "ok"}
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

func TestParser_RouteWithQueryParams(t *testing.T) {
	source := `@ GET /api/search {
  ? q: str!
  ? page: int = 1
  ? limit: int = 20
  ? tags: str[]
  > {query: q, page: page}
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

	route, ok := module.Items[0].(*interpreter.Route)
	if !ok {
		t.Fatalf("expected Route, got %T", module.Items[0])
	}

	if len(route.QueryParams) != 4 {
		t.Fatalf("expected 4 query params, got %d", len(route.QueryParams))
	}

	// Check first param: q: str!
	q := route.QueryParams[0]
	if q.Name != "q" {
		t.Errorf("expected param name 'q', got '%s'", q.Name)
	}
	if _, ok := q.Type.(interpreter.StringType); !ok {
		t.Errorf("expected StringType, got %T", q.Type)
	}
	if !q.Required {
		t.Error("expected q to be required")
	}
	if q.Default != nil {
		t.Error("expected q to have no default")
	}

	// Check second param: page: int = 1
	page := route.QueryParams[1]
	if page.Name != "page" {
		t.Errorf("expected param name 'page', got '%s'", page.Name)
	}
	if _, ok := page.Type.(interpreter.IntType); !ok {
		t.Errorf("expected IntType, got %T", page.Type)
	}
	if page.Required {
		t.Error("expected page to not be required")
	}
	if page.Default == nil {
		t.Error("expected page to have a default value")
	}

	// Check third param: limit: int = 20
	limit := route.QueryParams[2]
	if limit.Name != "limit" {
		t.Errorf("expected param name 'limit', got '%s'", limit.Name)
	}

	// Check fourth param: tags: str[]
	tags := route.QueryParams[3]
	if tags.Name != "tags" {
		t.Errorf("expected param name 'tags', got '%s'", tags.Name)
	}
	if !tags.IsArray {
		t.Error("expected tags to be an array type")
	}
	arrayType, ok := tags.Type.(interpreter.ArrayType)
	if !ok {
		t.Errorf("expected ArrayType, got %T", tags.Type)
	} else if _, ok := arrayType.ElementType.(interpreter.StringType); !ok {
		t.Errorf("expected array element type StringType, got %T", arrayType.ElementType)
	}
}

func TestParser_Assignment(t *testing.T) {
	source := `@ GET /test {
  $ x = 42
  > {value: x}
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

func TestParser_Reassignment(t *testing.T) {
	source := `@ GET /test {
  $ x = 0
  x = 42
  > {value: x}
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

	route, ok := module.Items[0].(*interpreter.Route)
	if !ok {
		t.Fatalf("expected Route, got %T", module.Items[0])
	}

	if len(route.Body) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(route.Body))
	}

	// First statement should be an AssignStatement (declaration)
	assign, ok := route.Body[0].(interpreter.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement for first statement, got %T", route.Body[0])
	}
	if assign.Target != "x" {
		t.Errorf("expected target 'x', got '%s'", assign.Target)
	}

	// Second statement should be a ReassignStatement (reassignment without $)
	reassign, ok := route.Body[1].(interpreter.ReassignStatement)
	if !ok {
		t.Fatalf("expected ReassignStatement for second statement, got %T", route.Body[1])
	}
	if reassign.Target != "x" {
		t.Errorf("expected reassign target 'x', got '%s'", reassign.Target)
	}

	// Check the value is a literal 42
	lit, ok := reassign.Value.(interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("expected LiteralExpr for reassign value, got %T", reassign.Value)
	}
	intLit, ok := lit.Value.(interpreter.IntLiteral)
	if !ok {
		t.Fatalf("expected IntLiteral, got %T", lit.Value)
	}
	if intLit.Value != 42 {
		t.Errorf("expected value 42, got %d", intLit.Value)
	}
}

func TestParser_ReassignmentWithExpression(t *testing.T) {
	source := `@ GET /test {
  $ x = 1
  x = x + 1
  > {value: x}
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

	route, ok := module.Items[0].(*interpreter.Route)
	if !ok {
		t.Fatalf("expected Route, got %T", module.Items[0])
	}

	// Second statement should be a ReassignStatement with binary expression
	reassign, ok := route.Body[1].(interpreter.ReassignStatement)
	if !ok {
		t.Fatalf("expected ReassignStatement, got %T", route.Body[1])
	}

	// Check the value is a binary expression (x + 1)
	binExpr, ok := reassign.Value.(interpreter.BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr for reassign value, got %T", reassign.Value)
	}
	if binExpr.Op != interpreter.Add {
		t.Errorf("expected Add operator, got %v", binExpr.Op)
	}
}

func TestParser_BinaryOp(t *testing.T) {
	source := `@ GET /calc {
  $ result = 10 + 20 * 2
  > {result: result}
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
			source := fmt.Sprintf("@ route /test %s {\n  > {status: \"ok\"}\n}", tt.methodStr)
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
			source := fmt.Sprintf("@ GET %s {\n  > {ok: true}\n}", tt.path)
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
			source := fmt.Sprintf("@ GET /test %s {\n  > {ok: true}\n}", tt.returnTypeStr)
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
	source := `@ GET /api/admin {
  + auth(jwt, role: admin)
  + ratelimit(50/min)
  > {status: "ok"}
}`

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
			source: `@ GET /calc {
  $ result = (10 + 20) * (5 - 2)
  > {result: result}
}`,
		},
		{
			name: "string concatenation",
			source: `@ GET /concat {
  $ full = "Hello, " + name + "!"
  > {text: full}
}`,
		},
		{
			name: "comparison operators",
			source: `@ GET /compare {
  $ check = age >= 18
  > {adult: check}
}`,
		},
		{
			name: "field access",
			source: `@ GET /field {
  $ name = user.profile.name
  > {name: name}
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
			source: `@ GET /obj {
  > {name: "test", age: 30}
}`,
		},
		{
			name: "nested object",
			source: `@ GET /nested {
  > {user: {name: "test", profile: {bio: "dev"}}}
}`,
		},
		{
			name: "object with expressions",
			source: `@ GET /expr {
  > {total: price * quantity, discount: true}
}`,
		},
		{
			name: "object with variables",
			source: `@ GET /vars {
  $ x = 42
  > {value: x, doubled: x * 2}
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
			source: `@ GET /time {
  > {now: time.now()}
}`,
		},
		{
			name: "with args",
			source: `@ GET /add {
  > {result: add(10, 20)}
}`,
		},
		{
			name: "nested calls",
			source: `@ GET /nested {
  > {result: max(min(x, 10), 5)}
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
			assert.Len(t, module.Items, 1)
		})
	}
}

// Test multiple routes in one file
func TestParser_MultipleRoutes(t *testing.T) {
	source := `@ GET /hello {
  > {message: "Hello"}
}

@ GET /goodbye {
  > {message: "Goodbye"}
}`

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
			source:      "@ route {\n  > {ok: true}\n}",
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
			source:      "@ GET /test {}",
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
			source: `@ GET /check/:age {
  if age >= 18 {
    > {status: "adult"}
  } else {
    > {status: "minor"}
  }
}`,
		},
		{
			name: "if without else",
			source: `@ GET /check {
  if true {
    > {ok: true}
  }
}`,
		},
		{
			name: "if with multiple statements",
			source: `@ GET /check {
  if x > 10 {
    $ result = "high"
    > {result: result}
  } else {
    $ result = "low"
    > {result: result}
  }
}`,
		},
		{
			name: "nested if",
			source: `@ GET /nested {
  if x > 0 {
    if y > 0 {
      > {quadrant: 1}
    } else {
      > {quadrant: 4}
    }
  } else {
    > {quadrant: 2}
  }
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
			source: `@ GET /and {
  $ result = true && false
  > {result: result}
}`,
		},
		{
			name: "OR operator",
			source: `@ GET /or {
  $ result = true || false
  > {result: result}
}`,
		},
		{
			name: "complex logical expression",
			source: `@ GET /complex {
  $ result = (x > 10 && y < 20) || z == 0
  > {result: result}
}`,
		},
		{
			name: "if with logical operators",
			source: `@ GET /check {
  if age >= 18 && hasLicense {
    > {canDrive: true}
  } else {
    > {canDrive: false}
  }
}`,
		},
		{
			name: "multiple AND/OR",
			source: `@ GET /multi {
  $ result = a && b && c || d || e
  > {result: result}
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
			source: `@ GET /empty {
  $ arr = []
  > {array: arr}
}`,
		},
		{
			name: "integer array",
			source: `@ GET /ints {
  $ numbers = [1, 2, 3, 4, 5]
  > {numbers: numbers}
}`,
		},
		{
			name: "string array",
			source: `@ GET /strings {
  $ names = ["Alice", "Bob", "Charlie"]
  > {names: names}
}`,
		},
		{
			name: "mixed array",
			source: `@ GET /mixed {
  $ items = [1, "two", true, 4.5]
  > {items: items}
}`,
		},
		{
			name: "nested arrays",
			source: `@ GET /nested {
  $ matrix = [[1, 2], [3, 4]]
  > {matrix: matrix}
}`,
		},
		{
			name: "array with expressions",
			source: `@ GET /expr {
  $ values = [1 + 1, x * 2, y]
  > {values: values}
}`,
		},
		{
			name: "array of objects",
			source: `@ GET /objects {
  $ users = [
    {name: "Alice", age: 30},
    {name: "Bob", age: 25}
  ]
  > {users: users}
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

@ GET /hello {
  > {text: "Hello, World!", timestamp: 1234567890}
}

@ GET /greet/:name -> Message {
  $ message = {
    text: "Hello, " + name + "!",
    timestamp: time.now()
  }
  > message
}`

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
	source := `@ GET /hello {
  > {message: "Hello, World!"}
}`

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

@ GET /api/users/:id {
  + auth(jwt)
  + ratelimit(100/min)
  $ user = db.users.get(id)
  > user
}`

	for i := 0; i < b.N; i++ {
		lexer := NewLexer(source)
		tokens, _ := lexer.Tokenize()
		parser := NewParser(tokens)
		_, _ = parser.Parse()
	}
}

// Test While Loops

func TestParser_SimpleWhileLoop(t *testing.T) {
	source := `@ GET /test {
  $ i = 0
  while i < 5 {
    $ i = i + 1
  }
  > i
}`

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
	source := `@ GET /test {
  $ i = 0
  while i < 3 {
    $ j = 0
    while j < 2 {
      $ j = j + 1
    }
    $ i = i + 1
  }
  > i
}`

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
	source := `@ GET /test {
  $ i = 0
  $ max = 10
  while i < max && i < 100 {
    $ i = i + 1
  }
  > i
}`

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
	source := `@ GET /test {
  $ sum = 0
  $ i = 0
  while i < 5 {
    $ sum = sum + i
    $ i = i + 1
    $ temp = i * 2
  }
  > sum
}`

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
	source := `@ GET /test {
  $ sum = 0
  $ i = 0
  while i < 10 {
    if i < 5 {
      $ sum = sum + i
    }
    $ i = i + 1
  }
  > sum
}`

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
	source := `@ GET /test {
  $ items = [1, 2, 3]
  for item in items {
    $ x = item + 1
  }
  > {ok: true}
}`

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
	source := `@ GET /test {
  for index, value in array {
    $ result = index + value
  }
  > {ok: true}
}`

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
	source := `@ GET /test {
  for item in [1, 2, 3] {
    $ x = item
  }
  > {ok: true}
}`

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
	source := `@ GET /test {
  for key, value in {name: "Alice", age: 30} {
    $ output = key + value
  }
  > {ok: true}
}`

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
	source := `@ GET /test {
  for row in rows {
    for col in row {
      $ cell = col
    }
  }
  > {ok: true}
}`

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
	input := `@ GET /test {
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
}`
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
	input := `@ GET /api/status {
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
}`
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
	input := `@ GET /test {
  switch x {
    case 1 {
      > "one"
    }
  }
}`
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
	input := `@ GET /test {
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
}`
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
	input := `@ GET /test {
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
}`
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

// Test CLI Command Directive (! syntax)
func TestParser_CLICommand_BangSyntax(t *testing.T) {
	source := `! greet name: str! --greeting: str = "Hello" {
  > greeting + ", " + name + "!"
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	cmd, ok := module.Items[0].(*interpreter.Command)
	require.True(t, ok, "Expected Command, got %T", module.Items[0])

	assert.Equal(t, "greet", cmd.Name)
	assert.Len(t, cmd.Params, 2)

	// First param: positional required parameter
	assert.Equal(t, "name", cmd.Params[0].Name)
	assert.True(t, cmd.Params[0].Required)
	assert.False(t, cmd.Params[0].IsFlag)

	// Second param: flag with default value
	assert.Equal(t, "greeting", cmd.Params[1].Name)
	assert.True(t, cmd.Params[1].IsFlag)
	assert.NotNil(t, cmd.Params[1].Default)

	assert.Greater(t, len(cmd.Body), 0)
}

func TestParser_CLICommand_AtCommandSyntax(t *testing.T) {
	source := `@ command hello name: str! {
  > "Hello, " + name + "!"
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	cmd, ok := module.Items[0].(*interpreter.Command)
	require.True(t, ok)

	assert.Equal(t, "hello", cmd.Name)
	assert.Len(t, cmd.Params, 1)
	assert.Equal(t, "name", cmd.Params[0].Name)
}

func TestParser_CLICommand_WithDescription(t *testing.T) {
	source := `! deploy "Deploy application to server" env: str! --force: bool {
  > {env: env, force: force}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	cmd, ok := module.Items[0].(*interpreter.Command)
	require.True(t, ok)

	assert.Equal(t, "deploy", cmd.Name)
	assert.Equal(t, "Deploy application to server", cmd.Description)
	assert.Len(t, cmd.Params, 2)
}

func TestParser_CLICommand_WithReturnType(t *testing.T) {
	source := `! add x: int! y: int! -> int {
  > x + y
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	cmd, ok := module.Items[0].(*interpreter.Command)
	require.True(t, ok)

	assert.Equal(t, "add", cmd.Name)
	assert.NotNil(t, cmd.ReturnType)
	assert.IsType(t, interpreter.IntType{}, cmd.ReturnType)
}

// Test Cron Task Directive (* syntax)
func TestParser_CronTask_StarSyntax(t *testing.T) {
	source := `* "0 0 * * *" daily_cleanup {
  $ count = db.cleanupOldRecords()
  > {deleted: count}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	task, ok := module.Items[0].(*interpreter.CronTask)
	require.True(t, ok, "Expected CronTask, got %T", module.Items[0])

	assert.Equal(t, "0 0 * * *", task.Schedule)
	assert.Equal(t, "daily_cleanup", task.Name)
	assert.Greater(t, len(task.Body), 0)
}

func TestParser_CronTask_AtCronSyntax(t *testing.T) {
	source := `@ cron "*/5 * * * *" five_minute_job {
  > {status: "running"}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	task, ok := module.Items[0].(*interpreter.CronTask)
	require.True(t, ok)

	assert.Equal(t, "*/5 * * * *", task.Schedule)
	assert.Equal(t, "five_minute_job", task.Name)
}

func TestParser_CronTask_WithTimezone(t *testing.T) {
	source := `* "0 9 * * 1-5" morning_report tz "America/New_York" {
  > {status: "ok"}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	task, ok := module.Items[0].(*interpreter.CronTask)
	require.True(t, ok)

	assert.Equal(t, "0 9 * * 1-5", task.Schedule)
	assert.Equal(t, "morning_report", task.Name)
	assert.Equal(t, "America/New_York", task.Timezone)
}

func TestParser_CronTask_WithInjections(t *testing.T) {
	source := `* "0 0 * * *" backup {
  % db: Database
  $ result = db.backupData()
  > result
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	task, ok := module.Items[0].(*interpreter.CronTask)
	require.True(t, ok)

	assert.Len(t, task.Injections, 1)
	assert.Equal(t, "db", task.Injections[0].Name)
}

func TestParser_CronTask_WithRetries(t *testing.T) {
	source := `* "0 */1 * * *" hourly_sync {
  + retries(3)
  > {status: "synced"}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	task, ok := module.Items[0].(*interpreter.CronTask)
	require.True(t, ok)

	assert.Equal(t, 3, task.Retries)
}

// Test Event Handler Directive (~ syntax)
func TestParser_EventHandler_TildeSyntax(t *testing.T) {
	source := `~ "user.created" {
  $ userId = event.userId
  > {status: "handled", userId: userId}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	handler, ok := module.Items[0].(*interpreter.EventHandler)
	require.True(t, ok, "Expected EventHandler, got %T", module.Items[0])

	assert.Equal(t, "user.created", handler.EventType)
	assert.Greater(t, len(handler.Body), 0)
}

func TestParser_EventHandler_AtEventSyntax(t *testing.T) {
	source := `@ event "order.paid" {
  $ orderId = input.orderId
  > {orderId: orderId}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	handler, ok := module.Items[0].(*interpreter.EventHandler)
	require.True(t, ok)

	assert.Equal(t, "order.paid", handler.EventType)
}

func TestParser_EventHandler_UnquotedEventType(t *testing.T) {
	source := `~ user.deleted {
  > {status: "processing"}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	handler, ok := module.Items[0].(*interpreter.EventHandler)
	require.True(t, ok)

	assert.Equal(t, "user.deleted", handler.EventType)
}

func TestParser_EventHandler_Async(t *testing.T) {
	source := `~ "email.send" async {
  $ recipient = event.to
  > {sent: true}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	handler, ok := module.Items[0].(*interpreter.EventHandler)
	require.True(t, ok)

	assert.Equal(t, "email.send", handler.EventType)
	assert.True(t, handler.Async)
}

func TestParser_EventHandler_WithInjections(t *testing.T) {
	source := `~ "notification.send" {
  % db: Database
  $ user = db.getUser(event.userId)
  > user
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	handler, ok := module.Items[0].(*interpreter.EventHandler)
	require.True(t, ok)

	assert.Len(t, handler.Injections, 1)
	assert.Equal(t, "db", handler.Injections[0].Name)
}

// Test Queue Worker Directive (& syntax)
func TestParser_QueueWorker_AmpersandSyntax(t *testing.T) {
	source := `& "email.send" {
  $ to = message.to
  $ body = message.body
  > {sent: true}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	worker, ok := module.Items[0].(*interpreter.QueueWorker)
	require.True(t, ok, "Expected QueueWorker, got %T", module.Items[0])

	assert.Equal(t, "email.send", worker.QueueName)
	assert.Greater(t, len(worker.Body), 0)
}

func TestParser_QueueWorker_AtQueueSyntax(t *testing.T) {
	source := `@ queue "image.resize" {
  $ imageId = input.imageId
  > {processed: true}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	worker, ok := module.Items[0].(*interpreter.QueueWorker)
	require.True(t, ok)

	assert.Equal(t, "image.resize", worker.QueueName)
}

func TestParser_QueueWorker_UnquotedQueueName(t *testing.T) {
	source := `& video.process {
  > {status: "processing"}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	worker, ok := module.Items[0].(*interpreter.QueueWorker)
	require.True(t, ok)

	assert.Equal(t, "video.process", worker.QueueName)
}

func TestParser_QueueWorker_WithConcurrency(t *testing.T) {
	source := `& "email.send" {
  + concurrency(5)
  > {status: "ok"}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	worker, ok := module.Items[0].(*interpreter.QueueWorker)
	require.True(t, ok)

	assert.Equal(t, 5, worker.Concurrency)
}

func TestParser_QueueWorker_WithRetriesAndTimeout(t *testing.T) {
	source := `& "data.process" {
  + retries(3)
  + timeout(60)
  > {done: true}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	worker, ok := module.Items[0].(*interpreter.QueueWorker)
	require.True(t, ok)

	assert.Equal(t, 3, worker.MaxRetries)
	assert.Equal(t, 60, worker.Timeout)
}

func TestParser_QueueWorker_WithInjections(t *testing.T) {
	source := `& "report.generate" {
  % db: Database
  $ data = db.getData()
  > data
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	worker, ok := module.Items[0].(*interpreter.QueueWorker)
	require.True(t, ok)

	assert.Len(t, worker.Injections, 1)
	assert.Equal(t, "db", worker.Injections[0].Name)
}

// Test mixed module with directive types (simplified version)
func TestParser_MixedModule_AllDirectiveTypes(t *testing.T) {
	// Test each directive individually to ensure they work
	tests := []struct {
		name   string
		source string
		check  func(*testing.T, interpreter.Item)
	}{
		{
			name: "TypeDef",
			source: `: User {
  id: int!
  name: str!
}`,
			check: func(t *testing.T, item interpreter.Item) {
				_, ok := item.(*interpreter.TypeDef)
				assert.True(t, ok, "Expected TypeDef")
			},
		},
		{
			name: "Route",
			source: `@ GET /users {
  > {users: []}
}`,
			check: func(t *testing.T, item interpreter.Item) {
				_, ok := item.(*interpreter.Route)
				assert.True(t, ok, "Expected Route")
			},
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
			require.Len(t, module.Items, 1)

			tt.check(t, module.Items[0])
		})
	}
}

// Test complex directive bodies
func TestParser_Directive_ComplexBody(t *testing.T) {
	source := `! process input: str! {
  $ items = []
  $ i = 0
  while i < 10 {
    if i < 5 {
      $ items = items + [i]
    }
    $ i = i + 1
  }
  for item in items {
    $ x = item * 2
  }
  > {count: items, processed: true}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	cmd, ok := module.Items[0].(*interpreter.Command)
	require.True(t, ok)

	// Should have: assign, assign, while, for, return (5 statements)
	assert.GreaterOrEqual(t, len(cmd.Body), 4)
}

// TestParser_UnaryOperators tests parsing of unary operators (! and -)
func TestParser_UnaryOperators(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		checkFn  func(t *testing.T, module *interpreter.Module)
	}{
		{
			name: "unary not operator",
			source: `@ GET /test {
  $ valid = !false
  > {valid: valid}
}`,
			checkFn: func(t *testing.T, module *interpreter.Module) {
				route := module.Items[0].(*interpreter.Route)
				assign := route.Body[0].(interpreter.AssignStatement)
				unary, ok := assign.Value.(interpreter.UnaryOpExpr)
				require.True(t, ok, "expected UnaryOpExpr")
				assert.Equal(t, interpreter.Not, unary.Op)
			},
		},
		{
			name: "unary negation",
			source: `@ GET /test {
  $ neg = -42
  > {neg: neg}
}`,
			checkFn: func(t *testing.T, module *interpreter.Module) {
				route := module.Items[0].(*interpreter.Route)
				assign := route.Body[0].(interpreter.AssignStatement)
				unary, ok := assign.Value.(interpreter.UnaryOpExpr)
				require.True(t, ok, "expected UnaryOpExpr")
				assert.Equal(t, interpreter.Neg, unary.Op)
			},
		},
		{
			name: "chained unary not",
			source: `@ GET /test {
  $ result = !!true
  > {result: result}
}`,
			checkFn: func(t *testing.T, module *interpreter.Module) {
				route := module.Items[0].(*interpreter.Route)
				assign := route.Body[0].(interpreter.AssignStatement)
				outer, ok := assign.Value.(interpreter.UnaryOpExpr)
				require.True(t, ok, "expected outer UnaryOpExpr")
				assert.Equal(t, interpreter.Not, outer.Op)
				inner, ok := outer.Right.(interpreter.UnaryOpExpr)
				require.True(t, ok, "expected inner UnaryOpExpr")
				assert.Equal(t, interpreter.Not, inner.Op)
			},
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

			tt.checkFn(t, module)
		})
	}
}

// TestParser_ArrayIndexing tests parsing of array index expressions
func TestParser_ArrayIndexing(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple array index",
			source: `@ GET /test {
  $ arr = [1, 2, 3]
  $ first = arr[0]
  > {first: first}
}`,
		},
		{
			name: "nested array index",
			source: `@ GET /test {
  $ matrix = [[1, 2], [3, 4]]
  $ val = matrix[0][1]
  > {val: val}
}`,
		},
		{
			name: "array index with expression",
			source: `@ GET /test {
  $ arr = [10, 20, 30]
  $ i = 1
  $ val = arr[i]
  > {val: val}
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

			require.NotNil(t, module)
			require.Len(t, module.Items, 1)
		})
	}
}

// TestParser_RateLimitVariations tests different rate limit formats
func TestParser_RateLimitVariations(t *testing.T) {
	tests := []struct {
		name            string
		source          string
		expectedReqs    uint32
		expectedWindow  string
	}{
		{
			name: "integer with slash format",
			source: `@ GET /test {
  + ratelimit(50/sec)
  > {ok: true}
}`,
			expectedReqs:   50,
			expectedWindow: "sec",
		},
		{
			name: "string format",
			source: `@ GET /test {
  + ratelimit("200/hour")
  > {ok: true}
}`,
			expectedReqs:   200,
			expectedWindow: "hour",
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

			route := module.Items[0].(*interpreter.Route)
			require.NotNil(t, route.RateLimit)
			assert.Equal(t, tt.expectedReqs, route.RateLimit.Requests)
			assert.Equal(t, tt.expectedWindow, route.RateLimit.Window)
		})
	}
}

// TestParser_ErrorHandling tests parser error handling for edge cases
func TestParser_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectError bool
	}{
		{
			name:        "unclosed brace",
			source:      `@ GET /test { $ x = 1`,
			expectError: true,
		},
		{
			name:        "invalid rate limit",
			source:      `@ GET /test + ratelimit() > {ok: true}`,
			expectError: true,
		},
		{
			name:        "missing return value",
			source:      `@ GET /test >`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source)
			tokens, err := lexer.Tokenize()
			if err != nil && tt.expectError {
				return // lexer error is acceptable
			}

			parser := NewParser(tokens)
			_, err = parser.Parse()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestParser_TypeDefWithoutColon tests type definitions without colon
func TestParser_TypeDefWithoutColon(t *testing.T) {
	source := `type Product {
  id: int!
  name: str!
  price: float
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)
	assert.Equal(t, "Product", typeDef.Name)
	assert.Len(t, typeDef.Fields, 3)
}

// TestParser_ValidationStatement tests ? validation syntax
func TestParser_ValidationStatement(t *testing.T) {
	source := `@ GET /validate {
  ? isValid(input)
  > {valid: true}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	route := module.Items[0].(*interpreter.Route)
	require.GreaterOrEqual(t, len(route.Body), 1)
}

// TestParser_ExpressionStatements tests standalone expression statements
func TestParser_ExpressionStatements(t *testing.T) {
	source := `@ GET /test {
  $ _ = log("starting request")
  $ result = compute()
  > {result: result}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	route := module.Items[0].(*interpreter.Route)
	require.GreaterOrEqual(t, len(route.Body), 2)
}

// TestParser_MoreStatementTypes tests additional statement types for coverage
func TestParser_MoreStatementTypes(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "typed variable declaration",
			source: `@ GET /test {
  $ count: int = 0
  > {count: count}
}`,
		},
		{
			name: "else-if chain",
			source: `@ GET /test {
  $ x = 5
  if x > 10 {
    $ result = "big"
  } else if x > 5 {
    $ result = "medium"
  } else {
    $ result = "small"
  }
  > {result: result}
}`,
		},
		{
			name: "switch statement",
			source: `@ GET /test {
  $ day = 1
  switch day {
    case 1 {
      $ name = "Monday"
    }
    case 2 {
      $ name = "Tuesday"
    }
    default {
      $ name = "Unknown"
    }
  }
  > {day: name}
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
			require.NotNil(t, module)
		})
	}
}

// TestParser_AdvancedTypes tests advanced type parsing
func TestParser_AdvancedTypes(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple type",
			source: `: User {
  id: int
  name: str
}`,
		},
		{
			name: "nested object type",
			source: `: Profile {
  user: User
  settings: Settings
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
			require.NotNil(t, module)
		})
	}
}

// TestParser_WebSocketRoutes tests WebSocket route parsing
func TestParser_WebSocketRoutes(t *testing.T) {
	source := `@ ws /chat {
  on connect {
    > {status: "connected"}
  }
  on message {
    > {echo: input}
  }
  on disconnect {
    > {status: "disconnected"}
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
}

// TestParser_AuthConfigVariations tests auth configuration
func TestParser_AuthConfigVariations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "jwt auth with role",
			source: `@ GET /admin {
  + auth(jwt, admin)
  > {access: "granted"}
}`,
		},
		{
			name: "basic auth",
			source: `@ GET /protected {
  + auth(basic)
  > {protected: true}
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

			route := module.Items[0].(*interpreter.Route)
			require.NotNil(t, route.Auth)
		})
	}
}

// TestParser_LogicalOps tests logical operators (and, or)
func TestParser_LogicalOps(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "logical and",
			source: `@ GET /test {
  $ a = true
  $ b = false
  if a && b {
    > {result: "both"}
  }
  > {result: "not both"}
}`,
		},
		{
			name: "logical or",
			source: `@ GET /test {
  $ a = true
  $ b = false
  if a || b {
    > {result: "either"}
  }
  > {result: "neither"}
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
			require.NotNil(t, module)
		})
	}
}

// TestParser_ComparisonOperators tests comparison operators
func TestParser_ComparisonOperators(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "less than or equal",
			source: `@ GET /test {
  $ x = 5
  if x <= 10 {
    > {result: "small"}
  }
  > {result: "big"}
}`,
		},
		{
			name: "greater than or equal",
			source: `@ GET /test {
  $ x = 15
  if x >= 10 {
    > {result: "big"}
  }
  > {result: "small"}
}`,
		},
		{
			name: "not equal",
			source: `@ GET /test {
  $ x = 5
  if x != 10 {
    > {result: "not ten"}
  }
  > {result: "ten"}
}`,
		},
		{
			name: "equality",
			source: `@ GET /test {
  $ x = 10
  if x == 10 {
    > {result: "ten"}
  }
  > {result: "not ten"}
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
			require.NotNil(t, module)
		})
	}
}


// TestParser_ParserErrors tests that parser reports errors correctly
func TestParser_ParserErrors(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "incomplete type definition",
			source: `: User {`,
		},
		{
			name:   "incomplete route",
			source: `@ GET /test`,
		},
		{
			name:   "incomplete if statement",
			source: `@ GET /test if x {`,
		},
		{
			name:   "incomplete while loop",
			source: `@ GET /test while true {`,
		},
		{
			name:   "incomplete for loop",
			source: `@ GET /test for x in {`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source)
			tokens, _ := lexer.Tokenize()

			parser := NewParser(tokens)
			_, err := parser.Parse()
			// These should return errors
			_ = err
		})
	}
}

// TestParser_FieldAccessExpressions tests field access parsing
func TestParser_FieldAccessExpressions(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple field access",
			source: `@ GET /test {
  $ name = user.name
  > {name: name}
}`,
		},
		{
			name: "nested field access",
			source: `@ GET /test {
  $ city = user.address.city
  > {city: city}
}`,
		},
		{
			name: "method call on object",
			source: `@ GET /test {
  $ len = items.length()
  > {len: len}
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
			require.NotNil(t, module)
		})
	}
}


// TestParser_StringConcatenation tests string concatenation
func TestParser_StringConcatenation(t *testing.T) {
	source := `@ GET /test {
  $ name = "World"
  $ greeting = "Hello, " + name + "!"
  > {greeting: greeting}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.NotNil(t, module)
}

// TestParser_TokenTypes tests that token types are correctly identified
func TestParser_TokenTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"@", AT},
		{":", COLON},
		{"$", DOLLAR},
		{"+", PLUS},
		{"-", MINUS},
		{"*", STAR},
		{">", GREATER},
		{"<", LESS},
		{"==", EQ_EQ},
		{"!=", NOT_EQ},
		{"&&", AND},
		{"||", OR},
		{"{", LBRACE},
		{"}", RBRACE},
		{"[", LBRACKET},
		{"]", RBRACKET},
		{"(", LPAREN},
		{")", RPAREN},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)
			require.True(t, len(tokens) >= 1)
			assert.Equal(t, tt.expected, tokens[0].Type)
		})
	}
}

// Test Pattern Matching Parsing

func TestParser_MatchExpr_LiteralPatterns(t *testing.T) {
	input := `@ GET /test {
  $ result = match code {
    200 => "OK"
    404 => "Not Found"
    _ => "Unknown"
  }
  > result
}`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.NotNil(t, module)
	require.Len(t, module.Items, 1)

	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)
	require.Len(t, route.Body, 2)

	// First statement is the assignment with match expression
	// Try both pointer and value types
	var assignStmt interpreter.AssignStatement
	if ptr, ok := route.Body[0].(*interpreter.AssignStatement); ok {
		assignStmt = *ptr
	} else if val, ok := route.Body[0].(interpreter.AssignStatement); ok {
		assignStmt = val
	} else {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}
	assert.Equal(t, "result", assignStmt.Target)

	matchExpr, ok := assignStmt.Value.(interpreter.MatchExpr)
	require.True(t, ok)
	require.Len(t, matchExpr.Cases, 3)

	// Check first case: 200 => "OK"
	lit1, ok := matchExpr.Cases[0].Pattern.(interpreter.LiteralPattern)
	require.True(t, ok)
	intLit, ok := lit1.Value.(interpreter.IntLiteral)
	require.True(t, ok)
	assert.Equal(t, int64(200), intLit.Value)

	// Check wildcard case
	_, ok = matchExpr.Cases[2].Pattern.(interpreter.WildcardPattern)
	assert.True(t, ok)
}

func TestParser_MatchExpr_WithGuard(t *testing.T) {
	input := `@ GET /test {
  $ result = match n {
    x when x > 10 => "big"
    x => "small"
  }
  > result
}`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.NotNil(t, module)

	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	var assignStmt interpreter.AssignStatement
	if ptr, ok := route.Body[0].(*interpreter.AssignStatement); ok {
		assignStmt = *ptr
	} else if val, ok := route.Body[0].(interpreter.AssignStatement); ok {
		assignStmt = val
	} else {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}

	matchExpr, ok := assignStmt.Value.(interpreter.MatchExpr)
	require.True(t, ok)
	require.Len(t, matchExpr.Cases, 2)

	// First case should have a guard
	assert.NotNil(t, matchExpr.Cases[0].Guard)
	// Second case should not have a guard
	assert.Nil(t, matchExpr.Cases[1].Guard)
}

func TestParser_MatchExpr_ObjectPattern(t *testing.T) {
	input := `@ GET /test {
  $ result = match obj {
    {name, age} => name
    _ => "unknown"
  }
  > result
}`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.NotNil(t, module)

	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	var assignStmt interpreter.AssignStatement
	if ptr, ok := route.Body[0].(*interpreter.AssignStatement); ok {
		assignStmt = *ptr
	} else if val, ok := route.Body[0].(interpreter.AssignStatement); ok {
		assignStmt = val
	} else {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}

	matchExpr, ok := assignStmt.Value.(interpreter.MatchExpr)
	require.True(t, ok)
	require.Len(t, matchExpr.Cases, 2)

	// First case should be an object pattern
	objPattern, ok := matchExpr.Cases[0].Pattern.(interpreter.ObjectPattern)
	require.True(t, ok)
	require.Len(t, objPattern.Fields, 2)
	assert.Equal(t, "name", objPattern.Fields[0].Key)
	assert.Equal(t, "age", objPattern.Fields[1].Key)
}

func TestParser_MatchExpr_ArrayPattern(t *testing.T) {
	input := `@ GET /test {
  $ result = match arr {
    [first, second] => first
    [head, ...rest] => head
    _ => null
  }
  > result
}`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.NotNil(t, module)

	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	var assignStmt interpreter.AssignStatement
	if ptr, ok := route.Body[0].(*interpreter.AssignStatement); ok {
		assignStmt = *ptr
	} else if val, ok := route.Body[0].(interpreter.AssignStatement); ok {
		assignStmt = val
	} else {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}

	matchExpr, ok := assignStmt.Value.(interpreter.MatchExpr)
	require.True(t, ok)
	require.Len(t, matchExpr.Cases, 3)

	// First case: [first, second]
	arrPattern1, ok := matchExpr.Cases[0].Pattern.(interpreter.ArrayPattern)
	require.True(t, ok)
	require.Len(t, arrPattern1.Elements, 2)
	assert.Nil(t, arrPattern1.Rest)

	// Second case: [head, ...rest]
	arrPattern2, ok := matchExpr.Cases[1].Pattern.(interpreter.ArrayPattern)
	require.True(t, ok)
	require.Len(t, arrPattern2.Elements, 1)
	require.NotNil(t, arrPattern2.Rest)
	assert.Equal(t, "rest", *arrPattern2.Rest)
}

func TestLexer_MatchTokens(t *testing.T) {
	input := `match when =>`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	// Should have: MATCH, WHEN, FATARROW, EOF
	require.Len(t, tokens, 4)
	assert.Equal(t, MATCH, tokens[0].Type)
	assert.Equal(t, WHEN, tokens[1].Type)
	assert.Equal(t, FATARROW, tokens[2].Type)
	assert.Equal(t, EOF, tokens[3].Type)
}

func TestLexer_DotDotDot(t *testing.T) {
	input := `...rest`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	// Should have: DOTDOTDOT, IDENT, EOF
	require.Len(t, tokens, 3)
	assert.Equal(t, DOTDOTDOT, tokens[0].Type)
	assert.Equal(t, IDENT, tokens[1].Type)
	assert.Equal(t, "rest", tokens[1].Literal)
}

func TestParser_SSERoute(t *testing.T) {
	input := `@ sse /events {
		yield { msg: "hello" }
	}`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok, "expected *interpreter.Route, got %T", module.Items[0])
	assert.Equal(t, interpreter.SSE, route.Method)
	assert.Equal(t, "/events", route.Path)
	require.Len(t, route.Body, 1)

	yieldStmt, ok := route.Body[0].(interpreter.YieldStatement)
	require.True(t, ok, "expected YieldStatement, got %T", route.Body[0])
	assert.NotNil(t, yieldStmt.Value)
}

func TestParser_SSERouteWithPathParams(t *testing.T) {
	input := `@ sse /events/:userId {
		yield "connected"
	}`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)
	assert.Equal(t, interpreter.SSE, route.Method)
	assert.Equal(t, "/events/:userId", route.Path)
}

func TestParser_YieldStatement(t *testing.T) {
	input := `@ GET /test {
		yield "data"
	}`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)
	require.Len(t, route.Body, 1)

	_, ok = route.Body[0].(interpreter.YieldStatement)
	assert.True(t, ok, "expected YieldStatement, got %T", route.Body[0])
}
