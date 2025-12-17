package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test parsing and executing a complete .abc file
func TestIntegration_HelloWorldFile(t *testing.T) {
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

	// Lex
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)
	require.Greater(t, len(tokens), 0)

	// Parse
	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, module)
	assert.Len(t, module.Items, 3)

	// Load module into interpreter
	interp := interpreter.NewInterpreter()
	err = interp.LoadModule(*module)
	require.NoError(t, err)

	// Execute first route
	route1, ok := module.Items[1].(*interpreter.Route)
	require.True(t, ok)
	result1, err := interp.ExecuteRouteSimple(route1, nil)
	require.NoError(t, err)

	obj1, ok := result1.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Hello, World!", obj1["text"])
	assert.Equal(t, int64(1234567890), obj1["timestamp"])

	// Execute second route with params
	route2, ok := module.Items[2].(*interpreter.Route)
	require.True(t, ok)
	params := map[string]string{"name": "Alice"}
	result2, err := interp.ExecuteRouteSimple(route2, params)
	require.NoError(t, err)

	obj2, ok := result2.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Hello, Alice!", obj2["text"])
}

// Test parsing REST API example file
func TestIntegration_RestApiFile(t *testing.T) {
	source := `: User {
  id: int!
  name: str!
  email: str!
}

@ route /api/users/:id -> User
  + auth(jwt)
  % db: Database
  $ user = db.users.get(id)
  > user

@ route /health
  > {status: "ok", timestamp: now()}`

	// Lex
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	// Parse
	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, module)

	// Should have 1 type def and 2 routes
	assert.Len(t, module.Items, 3)

	// Check type def
	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	assert.True(t, ok)
	assert.Equal(t, "User", typeDef.Name)
	assert.Len(t, typeDef.Fields, 3)

	// Check first route
	route1, ok := module.Items[1].(*interpreter.Route)
	assert.True(t, ok)
	assert.Equal(t, "/api/users/:id", route1.Path)
	assert.NotNil(t, route1.Auth)
	assert.Equal(t, "jwt", route1.Auth.AuthType)

	// Check health route
	route2, ok := module.Items[2].(*interpreter.Route)
	assert.True(t, ok)
	assert.Equal(t, "/health", route2.Path)
}

// Test complete lexer -> parser -> interpreter pipeline
func TestIntegration_FullPipeline(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		routeIndex     int
		params         map[string]string
		expectedResult interface{}
	}{
		{
			name: "simple string return",
			source: `@ route /test
  > "Hello"`,
			routeIndex:     0,
			params:         nil,
			expectedResult: "Hello",
		},
		{
			name: "integer calculation",
			source: `@ route /calc
  $ result = 10 + 20
  > result`,
			routeIndex:     0,
			params:         nil,
			expectedResult: int64(30),
		},
		{
			name: "path parameter",
			source: `@ route /echo/:msg
  > msg`,
			routeIndex:     0,
			params:         map[string]string{"msg": "test"},
			expectedResult: "test",
		},
		{
			name: "object literal",
			source: `@ route /obj
  > {count: 42, active: true}`,
			routeIndex: 0,
			params:     nil,
			expectedResult: map[string]interface{}{
				"count":  int64(42),
				"active": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Lex
			lexer := parser.NewLexer(tt.source)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			// Parse
			p := parser.NewParser(tokens)
			module, err := p.Parse()
			require.NoError(t, err)

			// Execute
			interp := interpreter.NewInterpreter()
			route, ok := module.Items[tt.routeIndex].(*interpreter.Route)
			require.True(t, ok)

			result, err := interp.ExecuteRouteSimple(route, tt.params)
			require.NoError(t, err)

			// Compare results
			if obj, ok := tt.expectedResult.(map[string]interface{}); ok {
				resultObj, ok := result.(map[string]interface{})
				require.True(t, ok)
				for key, val := range obj {
					assert.Equal(t, val, resultObj[key])
				}
			} else {
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

// Test loading from actual example files
func TestIntegration_ExampleFiles(t *testing.T) {
	projectRoot := filepath.Join("..", "..")

	tests := []struct {
		name     string
		filePath string
		validate func(*testing.T, *interpreter.Module)
	}{
		{
			name:     "hello-world example",
			filePath: filepath.Join(projectRoot, "examples", "hello-world", "main.abc"),
			validate: func(t *testing.T, module *interpreter.Module) {
				assert.GreaterOrEqual(t, len(module.Items), 2, "should have at least 2 items")

				// Find routes
				routeCount := 0
				typeCount := 0
				for _, item := range module.Items {
					switch item.(type) {
					case interpreter.Route:
						routeCount++
					case interpreter.TypeDef:
						typeCount++
					}
				}

				assert.GreaterOrEqual(t, routeCount, 1, "should have at least 1 route")
				assert.GreaterOrEqual(t, typeCount, 0, "may have type definitions")
			},
		},
		{
			name:     "rest-api example",
			filePath: filepath.Join(projectRoot, "examples", "rest-api", "main.abc"),
			validate: func(t *testing.T, module *interpreter.Module) {
				assert.Greater(t, len(module.Items), 0, "should have items")

				// Find routes and types
				routeCount := 0
				typeCount := 0
				for _, item := range module.Items {
					switch item.(type) {
					case interpreter.Route:
						routeCount++
					case interpreter.TypeDef:
						typeCount++
					}
				}

				assert.Greater(t, routeCount, 0, "should have routes")
				assert.Greater(t, typeCount, 0, "should have type definitions")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read file
			content, err := os.ReadFile(tt.filePath)
			if err != nil {
				t.Skipf("Could not read example file %s: %v", tt.filePath, err)
				return
			}

			// Lex
			lexer := parser.NewLexer(string(content))
			tokens, err := lexer.Tokenize()
			require.NoError(t, err, "lexer should not error on example file")

			// Parse
			p := parser.NewParser(tokens)
			module, err := p.Parse()
			require.NoError(t, err, "parser should not error on example file")
			require.NotNil(t, module)

			// Validate
			tt.validate(t, module)
		})
	}
}

// Test error handling throughout the pipeline
func TestIntegration_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectLexErr  bool
		expectParseErr bool
	}{
		{
			name:           "valid code",
			source:         "@ route /test\n  > {ok: true}",
			expectLexErr:   false,
			expectParseErr: false,
		},
		{
			name:           "unclosed string (lexer handles it)",
			source:         "@ route /test\n  > {msg: \"unclosed",
			expectLexErr:   true, // Lexer now errors on unterminated strings
			expectParseErr: false,
		},
		{
			name:           "invalid token sequence",
			source:         "@ route\n  > {ok: true}", // missing path
			expectLexErr:   false,
			expectParseErr: true,
		},
		{
			name:           "unclosed brace",
			source:         ": User {\n  id: int!",
			expectLexErr:   false,
			expectParseErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Lex
			lexer := parser.NewLexer(tt.source)
			tokens, err := lexer.Tokenize()

			if tt.expectLexErr {
				assert.Error(t, err, "expected lexer error")
				return
			}
			require.NoError(t, err)

			// Parse
			p := parser.NewParser(tokens)
			_, err = p.Parse()

			if tt.expectParseErr {
				assert.Error(t, err, "expected parser error")
			} else {
				assert.NoError(t, err, "unexpected parser error")
			}
		})
	}
}

// Test multiple routes with different features
func TestIntegration_MultipleRoutesWithFeatures(t *testing.T) {
	source := `: User {
  id: int!
  name: str!
}

@ route /public
  > {message: "Public endpoint"}

@ route /protected [POST]
  + auth(jwt)
  + ratelimit(100/min)
  > {message: "Protected endpoint"}

@ route /users/:id [GET] -> User
  + auth(jwt)
  $ user = {id: id, name: "Test User"}
  > user`

	// Lex
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	// Parse
	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	// Should have 1 type def and 3 routes
	assert.Len(t, module.Items, 4)

	// Check each route
	routes := []*interpreter.Route{}
	for _, item := range module.Items {
		if route, ok := item.(*interpreter.Route); ok {
			routes = append(routes, route)
		}
	}
	require.Len(t, routes, 3)

	// First route - public
	assert.Equal(t, "/public", routes[0].Path)
	assert.Equal(t, interpreter.Get, routes[0].Method)
	assert.Nil(t, routes[0].Auth)
	assert.Nil(t, routes[0].RateLimit)

	// Second route - protected with POST
	assert.Equal(t, "/protected", routes[1].Path)
	assert.Equal(t, interpreter.Post, routes[1].Method)
	assert.NotNil(t, routes[1].Auth)
	assert.NotNil(t, routes[1].RateLimit)

	// Third route - with path param and return type
	assert.Equal(t, "/users/:id", routes[2].Path)
	assert.Equal(t, interpreter.Get, routes[2].Method)
	assert.NotNil(t, routes[2].Auth)
	_, ok := routes[2].ReturnType.(interpreter.NamedType)
	assert.True(t, ok)
}

// Test arithmetic and string operations
func TestIntegration_Expressions(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		params   map[string]string
		expected interface{}
	}{
		{
			name: "addition",
			source: `@ route /add
  $ result = 10 + 5
  > result`,
			params:   nil,
			expected: int64(15),
		},
		{
			name: "multiplication",
			source: `@ route /mul
  $ result = 6 * 7
  > result`,
			params:   nil,
			expected: int64(42),
		},
		{
			name: "string concat",
			source: `@ route /greet/:name
  $ msg = "Hello, " + name
  > msg`,
			params:   map[string]string{"name": "World"},
			expected: "Hello, World",
		},
		{
			name: "complex arithmetic",
			source: `@ route /complex
  $ a = 10
  $ b = 20
  $ c = a + b * 2
  > c`,
			params:   nil,
			expected: int64(50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := parser.NewLexer(tt.source)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err)

			p := parser.NewParser(tokens)
			module, err := p.Parse()
			require.NoError(t, err)

			interp := interpreter.NewInterpreter()
			route, ok := module.Items[0].(*interpreter.Route)
			require.True(t, ok)

			result, err := interp.ExecuteRouteSimple(route, tt.params)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark full pipeline
func BenchmarkIntegration_FullPipeline(b *testing.B) {
	source := `@ route /hello
  > {message: "Hello, World!"}`

	for i := 0; i < b.N; i++ {
		lexer := parser.NewLexer(source)
		tokens, _ := lexer.Tokenize()
		p := parser.NewParser(tokens)
		module, _ := p.Parse()
		interp := interpreter.NewInterpreter()
		if route, ok := module.Items[0].(*interpreter.Route); ok {
			_, _ = interp.ExecuteRouteSimple(route, nil)
		}
	}
}

func BenchmarkIntegration_ComplexRoute(b *testing.B) {
	source := `: User {
  id: int!
  name: str!
}

@ route /api/users/:id [GET]
  + auth(jwt)
  + ratelimit(100/min)
  $ user = {id: id, name: "Test"}
  > user`

	for i := 0; i < b.N; i++ {
		lexer := parser.NewLexer(source)
		tokens, _ := lexer.Tokenize()
		p := parser.NewParser(tokens)
		module, _ := p.Parse()
		interp := interpreter.NewInterpreter()
		_ = interp.LoadModule(*module)
		if len(module.Items) >= 2 {
			if route, ok := module.Items[1].(*interpreter.Route); ok {
				params := map[string]string{"id": "123"}
				_, _ = interp.ExecuteRouteSimple(route, params)
			}
		}
	}
}
