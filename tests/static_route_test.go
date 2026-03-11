package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/parser"
	"github.com/glyphlang/glyph/pkg/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticRouteParseThenServe(t *testing.T) {
	// Create a temp directory with a test file
	tmpDir := t.TempDir()
	testContent := []byte("hello from static file")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.txt"), testContent, 0644))

	// Parse a static route directive
	source := `@ static /assets "` + tmpDir + `"`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	sr, ok := module.Items[0].(*ast.StaticRoute)
	require.True(t, ok)
	assert.Equal(t, "/assets", sr.Path)
	assert.Equal(t, tmpDir, sr.RootDir)

	// Wire it up the same way routes.go does
	staticServer, err := web.NewStaticFileServer(sr.RootDir, web.WithPrefix(sr.Path))
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.Handle(sr.Path+"/", staticServer)

	// Serve a request
	req := httptest.NewRequest(http.MethodGet, "/assets/test.txt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello from static file", w.Body.String())
}

func TestStaticRoutePathTraversalBlocked(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "safe.txt"), []byte("safe"), 0644))

	staticServer, err := web.NewStaticFileServer(tmpDir, web.WithPrefix("/assets"))
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.Handle("/assets/", staticServer)

	// Attempt path traversal
	req := httptest.NewRequest(http.MethodGet, "/assets/../../../etc/passwd", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Should not return 200
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestStaticRoute404ForMissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	staticServer, err := web.NewStaticFileServer(tmpDir, web.WithPrefix("/assets"))
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.Handle("/assets/", staticServer)

	req := httptest.NewRequest(http.MethodGet, "/assets/nonexistent.txt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
