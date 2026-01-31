package parser

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that examples/hello-world/main.glyph parses correctly
func TestExamples_HelloWorld(t *testing.T) {
	// Read the hello world example file
	examplePath := filepath.Join("..", "..", "examples", "hello-world", "main.glyph")
	content, err := os.ReadFile(examplePath)
	if err != nil {
		t.Skipf("Could not read example file: %v", err)
		return
	}

	// Lex
	lexer := NewLexer(string(content))
	tokens, err := lexer.Tokenize()
	require.NoError(t, err, "hello-world example should lex without errors")
	require.Greater(t, len(tokens), 0, "should produce tokens")

	// Parse
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err, "hello-world example should parse without errors")
	require.NotNil(t, module)

	// Validate structure - hello-world has 3 routes: /, /hello/:name, /health
	assert.GreaterOrEqual(t, len(module.Items), 3, "should have at least 3 items")

	// Check for routes
	routes := []*ast.Route{}
	for _, item := range module.Items {
		if route, ok := item.(*ast.Route); ok {
			routes = append(routes, route)
		}
	}
	assert.GreaterOrEqual(t, len(routes), 3, "should have at least 3 routes")

	// Verify route paths
	paths := make(map[string]bool)
	for _, route := range routes {
		paths[route.Path] = true
	}
	assert.True(t, paths["/"], "should have / route")
	assert.True(t, paths["/hello/:name"], "should have /hello/:name route")
	assert.True(t, paths["/health"], "should have /health route")
}

// Test that examples/rest-api/main.glyph parses correctly
func TestExamples_RestApi(t *testing.T) {
	// Read the rest-api example file
	examplePath := filepath.Join("..", "..", "examples", "rest-api", "main.glyph")
	content, err := os.ReadFile(examplePath)
	if err != nil {
		t.Skipf("Could not read example file: %v", err)
		return
	}

	// Lex
	lexer := NewLexer(string(content))
	tokens, err := lexer.Tokenize()
	require.NoError(t, err, "rest-api example should lex without errors")
	require.Greater(t, len(tokens), 0, "should produce tokens")

	// Parse
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err, "rest-api example should parse without errors")
	require.NotNil(t, module)

	// Validate structure
	assert.Greater(t, len(module.Items), 0, "should have items")

	// Check for type definitions
	typeDefs := []*ast.TypeDef{}
	for _, item := range module.Items {
		if typeDef, ok := item.(*ast.TypeDef); ok {
			typeDefs = append(typeDefs, typeDef)
		}
	}
	assert.GreaterOrEqual(t, len(typeDefs), 1, "should have at least 1 type definition")

	// Check for User type
	hasUserType := false
	for _, typeDef := range typeDefs {
		if typeDef.Name == "User" {
			hasUserType = true
			assert.GreaterOrEqual(t, len(typeDef.Fields), 3, "User should have at least 3 fields")
		}
	}
	assert.True(t, hasUserType, "should have User type definition")

	// Check for routes
	routes := []*ast.Route{}
	for _, item := range module.Items {
		if route, ok := item.(*ast.Route); ok {
			routes = append(routes, route)
		}
	}
	assert.GreaterOrEqual(t, len(routes), 2, "should have at least 2 routes")

	// Verify at least one route has auth
	hasAuthRoute := false
	for _, route := range routes {
		if route.Auth != nil {
			hasAuthRoute = true
			break
		}
	}
	assert.True(t, hasAuthRoute, "should have at least one route with auth")

	// Verify different HTTP methods are used
	methods := make(map[ast.HttpMethod]bool)
	for _, route := range routes {
		methods[route.Method] = true
	}
	assert.GreaterOrEqual(t, len(methods), 2, "should use at least 2 different HTTP methods")
}

// Test parsing all fixture files
func TestExamples_FixtureFiles(t *testing.T) {
	fixturesPath := filepath.Join("..", "..", "tests", "fixtures")

	// Check if fixtures directory exists
	if _, err := os.Stat(fixturesPath); os.IsNotExist(err) {
		t.Skip("Fixtures directory does not exist")
		return
	}

	// Read all .glyph files in fixtures
	files, err := filepath.Glob(filepath.Join(fixturesPath, "*.glyph"))
	if err != nil {
		t.Fatalf("Failed to read fixtures: %v", err)
	}

	for _, file := range files {
		filename := filepath.Base(file)
		t.Run(filename, func(t *testing.T) {
			content, err := os.ReadFile(file)
			require.NoError(t, err)

			// Lex
			lexer := NewLexer(string(content))
			tokens, err := lexer.Tokenize()

			// Skip files with "invalid" in the name if they produce lex errors
			if filepath.Base(file) == "invalid_syntax.glyph" {
				if err != nil {
					t.Skip("Invalid syntax file - expected to fail")
				}
			} else {
				require.NoError(t, err, "should lex without errors")
			}

			// Parse
			parser := NewParser(tokens)
			module, err := parser.Parse()

			// Skip files with "invalid" in the name if they produce parse errors
			if filepath.Base(file) == "invalid_syntax.glyph" {
				if err != nil {
					t.Skip("Invalid syntax file - expected to fail")
				}
			} else {
				require.NoError(t, err, "should parse without errors")
				require.NotNil(t, module)
			}
		})
	}
}

// Test specific fixture examples
func TestExamples_SimpleRoute(t *testing.T) {
	source := `@ GET /hello {
  > {message: "Hello, World!"}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	assert.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*ast.Route)
	require.True(t, ok)
	assert.Equal(t, "/hello", route.Path)
}

func TestExamples_PathParam(t *testing.T) {
	source := `@ GET /users/:id {
  > {id: id}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	assert.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*ast.Route)
	require.True(t, ok)
	assert.Equal(t, "/users/:id", route.Path)
}

func TestExamples_JsonResponse(t *testing.T) {
	source := `@ GET /data {
  > {
    status: "ok",
    count: 42,
    active: true,
    score: 98.5
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	assert.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*ast.Route)
	require.True(t, ok)
	assert.Equal(t, "/data", route.Path)
}

func TestExamples_MultipleRoutes(t *testing.T) {
	source := `@ GET /first {
  > {msg: "first"}
}

@ GET /second {
  > {msg: "second"}
}

@ GET /third {
  > {msg: "third"}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	assert.Len(t, module.Items, 3)

	paths := []string{}
	for _, item := range module.Items {
		if route, ok := item.(*ast.Route); ok {
			paths = append(paths, route.Path)
		}
	}

	assert.Equal(t, []string{"/first", "/second", "/third"}, paths)
}

func TestExamples_WithAuth(t *testing.T) {
	source := `@ GET /protected {
  + auth(jwt)
  > {msg: "protected"}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	assert.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*ast.Route)
	require.True(t, ok)
	require.NotNil(t, route.Auth)
	assert.Equal(t, "jwt", route.Auth.AuthType)
}

func TestExamples_PostRoute(t *testing.T) {
	source := `@ POST /create {
  > {created: true}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	assert.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*ast.Route)
	require.True(t, ok)
	assert.Equal(t, ast.Post, route.Method)
}

func TestExamples_ErrorHandling(t *testing.T) {
	source := `@ GET /divide/:a/:b {
  if b == 0 {
    > {error: "Division by zero"}
  } else {
    $ result = a / b
    > {result: result}
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	assert.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*ast.Route)
	require.True(t, ok)
	assert.Greater(t, len(route.Body), 0)
}
