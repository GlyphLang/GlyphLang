package main

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
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
		input    ast.HttpMethod
		expected string
	}{
		{ast.Get, "GET"},
		{ast.Post, "POST"},
		{ast.Put, "PUT"},
		{ast.Delete, "DELETE"},
		{ast.Patch, "PATCH"},
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
	route, ok := module.Items[0].(*ast.Route)
	require.True(t, ok)
	assert.Equal(t, "/hello", route.Path)
	assert.Equal(t, ast.Get, route.Method)
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
	route := &ast.Route{
		Path:   "/test",
		Method: ast.Get,
		Body:   []ast.Statement{},
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

// Tests for expand/compact watch mode functionality

func TestExpandFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a test .glyph file
	inputFile := filepath.Join(tmpDir, "test.glyph")
	content := `@ GET /hello
  $ message = "Hello"
  > {text: message}
`
	err := os.WriteFile(inputFile, []byte(content), 0644)
	require.NoError(t, err)

	// Expand the file
	outputFile := filepath.Join(tmpDir, "test.glyphx")
	err = expandFile(inputFile, outputFile)
	require.NoError(t, err)

	// Verify output file exists
	assert.FileExists(t, outputFile)

	// Verify content was expanded
	expanded, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(expanded), "route")
	assert.Contains(t, string(expanded), "let")
	assert.Contains(t, string(expanded), "return")
}

func TestExpandFileDefaultOutput(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a test .glyph file
	inputFile := filepath.Join(tmpDir, "test.glyph")
	content := `@ GET /hello
  > {status: "ok"}
`
	err := os.WriteFile(inputFile, []byte(content), 0644)
	require.NoError(t, err)

	// Expand the file with empty output (should use default)
	err = expandFile(inputFile, "")
	require.NoError(t, err)

	// Verify default output file exists (same name with .glyphx extension)
	expectedOutput := filepath.Join(tmpDir, "test.glyphx")
	assert.FileExists(t, expectedOutput)
}

func TestCompactFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a test .glyphx file (expanded syntax)
	inputFile := filepath.Join(tmpDir, "test.glyphx")
	content := `route GET /hello
  let message = "Hello"
  return {text: message}
`
	err := os.WriteFile(inputFile, []byte(content), 0644)
	require.NoError(t, err)

	// Compact the file
	outputFile := filepath.Join(tmpDir, "test.glyph")
	err = compactFile(inputFile, outputFile)
	require.NoError(t, err)

	// Verify output file exists
	assert.FileExists(t, outputFile)

	// Verify content was compacted
	compacted, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(compacted), "@")
	assert.Contains(t, string(compacted), "$")
	assert.Contains(t, string(compacted), ">")
}

func TestCompactFileDefaultOutput(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a test .glyphx file
	inputFile := filepath.Join(tmpDir, "test.glyphx")
	content := `route GET /hello
  return {status: "ok"}
`
	err := os.WriteFile(inputFile, []byte(content), 0644)
	require.NoError(t, err)

	// Compact the file with empty output (should use default)
	err = compactFile(inputFile, "")
	require.NoError(t, err)

	// Verify default output file exists (same name with .glyph extension)
	expectedOutput := filepath.Join(tmpDir, "test.glyph")
	assert.FileExists(t, expectedOutput)
}

func TestExpandDirectory(t *testing.T) {
	// Create temp directory with subdirectories
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "src")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	// Create test .glyph files
	file1 := filepath.Join(tmpDir, "main.glyph")
	file2 := filepath.Join(subDir, "routes.glyph")

	content := `@ GET /test
  > {ok: true}
`
	err = os.WriteFile(file1, []byte(content), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(content), 0644)
	require.NoError(t, err)

	// Expand the directory
	err = expandDirectory(tmpDir, "")
	require.NoError(t, err)

	// Verify output files exist
	assert.FileExists(t, filepath.Join(tmpDir, "main.glyphx"))
	assert.FileExists(t, filepath.Join(subDir, "routes.glyphx"))
}

func TestExpandDirectoryWithOutputDir(t *testing.T) {
	// Create temp directories
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// Create a test .glyph file
	inputFile := filepath.Join(srcDir, "test.glyph")
	content := `@ GET /test
  > {ok: true}
`
	err := os.WriteFile(inputFile, []byte(content), 0644)
	require.NoError(t, err)

	// Expand to output directory
	err = expandDirectory(srcDir, outDir)
	require.NoError(t, err)

	// Verify output file exists in output directory
	assert.FileExists(t, filepath.Join(outDir, "test.glyphx"))
}

func TestCompactDirectory(t *testing.T) {
	// Create temp directory with subdirectories
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "src")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	// Create test .glyphx files
	file1 := filepath.Join(tmpDir, "main.glyphx")
	file2 := filepath.Join(subDir, "routes.glyphx")

	content := `route GET /test
  return {ok: true}
`
	err = os.WriteFile(file1, []byte(content), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(content), 0644)
	require.NoError(t, err)

	// Compact the directory
	err = compactDirectory(tmpDir, "")
	require.NoError(t, err)

	// Verify output files exist
	assert.FileExists(t, filepath.Join(tmpDir, "main.glyph"))
	assert.FileExists(t, filepath.Join(subDir, "routes.glyph"))
}

func TestCompactDirectoryWithOutputDir(t *testing.T) {
	// Create temp directories
	srcDir := t.TempDir()
	outDir := t.TempDir()

	// Create a test .glyphx file
	inputFile := filepath.Join(srcDir, "test.glyphx")
	content := `route GET /test
  return {ok: true}
`
	err := os.WriteFile(inputFile, []byte(content), 0644)
	require.NoError(t, err)

	// Compact to output directory
	err = compactDirectory(srcDir, outDir)
	require.NoError(t, err)

	// Verify output file exists in output directory
	assert.FileExists(t, filepath.Join(outDir, "test.glyph"))
}

func TestExpandCompactRoundTrip(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Original compact content
	original := `@ GET /hello
  $ name = "World"
  > {greeting: "Hello, " + name}
`
	originalFile := filepath.Join(tmpDir, "original.glyph")
	err := os.WriteFile(originalFile, []byte(original), 0644)
	require.NoError(t, err)

	// Expand
	expandedFile := filepath.Join(tmpDir, "expanded.glyphx")
	err = expandFile(originalFile, expandedFile)
	require.NoError(t, err)

	// Compact back
	compactedFile := filepath.Join(tmpDir, "compacted.glyph")
	err = compactFile(expandedFile, compactedFile)
	require.NoError(t, err)

	// Read both files
	originalContent, err := os.ReadFile(originalFile)
	require.NoError(t, err)
	compactedContent, err := os.ReadFile(compactedFile)
	require.NoError(t, err)

	// Normalize whitespace for comparison
	normalizeWs := func(s string) string {
		return strings.Join(strings.Fields(s), " ")
	}

	// They should be equivalent (ignoring whitespace differences)
	assert.Equal(t, normalizeWs(string(originalContent)), normalizeWs(string(compactedContent)))
}

func TestExpandFileNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does-not-exist.glyph")

	err := expandFile(nonExistent, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestCompactFileNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does-not-exist.glyphx")

	err := compactFile(nonExistent, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}
