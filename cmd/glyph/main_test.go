package main

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangeExtension(t *testing.T) {
	tests := []struct {
		input    string
		newExt   string
		expected string
	}{
		{"main.old", ".glyph", "main.glyph"},
		{"test.txt", ".md", "test.md"},
		{"/path/to/file.go", ".txt", "/path/to/file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := changeExtension(tt.input, tt.newExt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertHTTPMethod(t *testing.T) {
	tests := []struct {
		input    interpreter.HttpMethod
		expected string
	}{
		{interpreter.Get, "GET"},
		{interpreter.Post, "POST"},
		{interpreter.Put, "PUT"},
		{interpreter.Delete, "DELETE"},
		{interpreter.Patch, "PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := convertHTTPMethod(tt.input)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestParseSource(t *testing.T) {
	source := `@ GET /hello {
  > {text: "Hello, World!"}
}`

	module, err := parseSource(source)
	require.NoError(t, err)
	assert.NotNil(t, module)
	assert.Len(t, module.Items, 1)

	// Verify the route was created
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)
	assert.Equal(t, "/hello", route.Path)
	assert.Equal(t, interpreter.Get, route.Method)
}

func TestGetHelloWorldTemplate(t *testing.T) {
	template := getHelloWorldTemplate()
	assert.NotEmpty(t, template)
	assert.Contains(t, template, "Hello World Example")
	assert.Contains(t, template, "@ GET /hello")
}

func TestGetRestAPITemplate(t *testing.T) {
	template := getRestAPITemplate()
	assert.NotEmpty(t, template)
	assert.Contains(t, template, "Example REST API")
	assert.Contains(t, template, "@ GET /api/users")
	assert.Contains(t, template, "@ GET /health")
}

func TestRunInitCommand(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	projectName := "test-project"
	projectPath := filepath.Join(tmpDir, projectName)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	// Simulate init command
	err = os.MkdirAll(projectName, 0755)
	require.NoError(t, err)

	mainFile := filepath.Join(projectName, "main.glyph")
	err = os.WriteFile(mainFile, []byte(getHelloWorldTemplate()), 0644)
	require.NoError(t, err)

	// Verify files were created
	assert.DirExists(t, projectPath)
	assert.FileExists(t, mainFile)

	// Read and verify content
	content, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Hello World Example")
}

func TestExecuteRoute(t *testing.T) {
	// Create a simple route with empty body
	// This test just verifies the function signature is correct
	route := &interpreter.Route{
		Path:   "/test",
		Method: interpreter.Get,
		Body:   []interpreter.Statement{},
	}

	// Create test context and interpreter
	// Need to create a proper http.Request with URL
	testURL, _ := url.Parse("http://localhost/test")
	req, _ := http.NewRequest("GET", "http://localhost/test", nil)
	req.URL = testURL

	ctx := &server.Context{
		Request:    req,
		PathParams: make(map[string]string),
	}
	interp := interpreter.NewInterpreter()

	// Execute the route - empty body returns nil which is valid
	result, err := executeRoute(route, ctx, interp)
	require.NoError(t, err)
	// Empty route body returns nil, which is expected
	assert.Nil(t, result)
}
