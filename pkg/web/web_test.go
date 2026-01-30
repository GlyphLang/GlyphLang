// Tests for pkg/web - web page support including template engine,
// static file server, response helpers, cache config, and page data.
// Implementation verified in pkg/web/web.go.
package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ResponseType and ContentType ---

func TestContentTypeForResponse(t *testing.T) {
	assert.Equal(t, "text/html; charset=utf-8", ContentTypeForResponse(HTML))
	assert.Equal(t, "text/plain; charset=utf-8", ContentTypeForResponse(Text))
	assert.Equal(t, "application/json; charset=utf-8", ContentTypeForResponse(JSON))
	assert.Equal(t, "application/octet-stream", ContentTypeForResponse(File))
	assert.Equal(t, "application/octet-stream", ContentTypeForResponse("unknown"))
}

// --- TemplateEngine ---

func TestNewTemplateEngineFromStrings(t *testing.T) {
	engine := NewTemplateEngineFromStrings()
	assert.NotNil(t, engine)
	assert.Empty(t, engine.TemplateNames())
}

func TestNewTemplateEngineValidatesDir(t *testing.T) {
	_, err := NewTemplateEngine("/nonexistent/dir/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestNewTemplateEngineNotADir(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "notadir")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	_, err = NewTemplateEngine(tmpFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestNewTemplateEngineValid(t *testing.T) {
	tmpDir := t.TempDir()
	engine, err := NewTemplateEngine(tmpDir)
	require.NoError(t, err)
	assert.NotNil(t, engine)
}

func TestLoadTemplateString(t *testing.T) {
	engine := NewTemplateEngineFromStrings()
	err := engine.LoadTemplateString("hello", "<h1>Hello {{.Name}}</h1>")
	require.NoError(t, err)
	assert.True(t, engine.HasTemplate("hello"))
}

func TestLoadTemplateStringInvalid(t *testing.T) {
	engine := NewTemplateEngineFromStrings()
	err := engine.LoadTemplateString("bad", "{{.Invalid")
	assert.Error(t, err)
}

func TestRenderString(t *testing.T) {
	engine := NewTemplateEngineFromStrings()
	err := engine.LoadTemplateString("greeting", "Hello, {{.Name}}!")
	require.NoError(t, err)
	result, err := engine.RenderString("greeting", map[string]interface{}{"Name": "Alice"})
	require.NoError(t, err)
	assert.Equal(t, "Hello, Alice!", result)
}

func TestRenderStringNotFound(t *testing.T) {
	engine := NewTemplateEngineFromStrings()
	_, err := engine.RenderString("missing", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRenderToWriter(t *testing.T) {
	engine := NewTemplateEngineFromStrings()
	err := engine.LoadTemplateString("page", "<p>{{.Content}}</p>")
	require.NoError(t, err)
	var buf strings.Builder
	err = engine.Render(&buf, "page", map[string]interface{}{"Content": "test"})
	require.NoError(t, err)
	assert.Equal(t, "<p>test</p>", buf.String())
}

func TestTemplateAutoEscape(t *testing.T) {
	engine := NewTemplateEngineFromStrings()
	err := engine.LoadTemplateString("xss", "<div>{{.Input}}</div>")
	require.NoError(t, err)
	result, err := engine.RenderString("xss", map[string]interface{}{
		"Input": "<script>alert('xss')</script>",
	})
	require.NoError(t, err)
	assert.NotContains(t, result, "<script>")
	assert.Contains(t, result, "&lt;script&gt;")
}

func TestTemplateNames(t *testing.T) {
	engine := NewTemplateEngineFromStrings()
	_ = engine.LoadTemplateString("a", "A")
	_ = engine.LoadTemplateString("b", "B")
	names := engine.TemplateNames()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "a")
	assert.Contains(t, names, "b")
}

func TestHasTemplate(t *testing.T) {
	engine := NewTemplateEngineFromStrings()
	assert.False(t, engine.HasTemplate("nope"))
	_ = engine.LoadTemplateString("exists", "yes")
	assert.True(t, engine.HasTemplate("exists"))
}

func TestAddFunc(t *testing.T) {
	engine := NewTemplateEngineFromStrings()
	engine.AddFunc("double", func(s string) string { return s + s })
	err := engine.LoadTemplateString("fn", `{{double "ab"}}`)
	require.NoError(t, err)
	result, err := engine.RenderString("fn", nil)
	require.NoError(t, err)
	assert.Equal(t, "abab", result)
}

func TestBuiltinFuncs(t *testing.T) {
	engine := NewTemplateEngineFromStrings()
	err := engine.LoadTemplateString("upper", `{{upper "hello"}}`)
	require.NoError(t, err)
	result, err := engine.RenderString("upper", nil)
	require.NoError(t, err)
	assert.Equal(t, "HELLO", result)
}

func TestLoadTemplateFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	content := "<h1>{{.Title}}</h1>"
	err := os.WriteFile(filepath.Join(tmpDir, "page.html"), []byte(content), 0644)
	require.NoError(t, err)

	engine, err := NewTemplateEngine(tmpDir)
	require.NoError(t, err)
	err = engine.LoadTemplate("page.html")
	require.NoError(t, err)

	result, err := engine.RenderString("page.html", map[string]interface{}{"Title": "Test"})
	require.NoError(t, err)
	assert.Contains(t, result, "<h1>Test</h1>")
}

// --- StaticFileServer ---

func TestStaticFileServerServesFile(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "hello.txt"), []byte("hello world"), 0644)
	require.NoError(t, err)

	srv, err := NewStaticFileServer(tmpDir)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/hello.txt", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "hello world")
}

func TestStaticFileServer404(t *testing.T) {
	tmpDir := t.TempDir()
	srv, err := NewStaticFileServer(tmpDir)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/missing.txt", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestStaticFileServerMethodNotAllowed(t *testing.T) {
	tmpDir := t.TempDir()
	srv, err := NewStaticFileServer(tmpDir)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/file.txt", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestStaticFileServerPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "style.css"), []byte("body{}"), 0644)
	require.NoError(t, err)

	srv, err := NewStaticFileServer(tmpDir, WithPrefix("/assets"))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/assets/style.css", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "body{}")
}

func TestStaticFileServerIndexFile(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("<html>home</html>"), 0644)
	require.NoError(t, err)

	srv, err := NewStaticFileServer(tmpDir)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<html>home</html>")
}

func TestStaticFileServerCustomIndex(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "default.html"), []byte("<html>default</html>"), 0644)
	require.NoError(t, err)

	srv, err := NewStaticFileServer(tmpDir, WithIndex("default.html"))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<html>default</html>")
}

func TestStaticFileServerCacheControl(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "app.js"), []byte("var x=1"), 0644)
	require.NoError(t, err)

	srv, err := NewStaticFileServer(tmpDir, WithMaxAge(3600))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "public, max-age=3600", rec.Header().Get("Cache-Control"))
}

func TestStaticFileServerDirectoryForbidden(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	srv, err := NewStaticFileServer(tmpDir)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/sub", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestStaticFileServerDirectoryListing(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "docs")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir, "readme.txt"), []byte("hi"), 0644)
	require.NoError(t, err)

	srv, err := NewStaticFileServer(tmpDir, WithDirectoryListing(true))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "readme.txt")
	assert.Contains(t, rec.Body.String(), "Index of")
}

func TestStaticFileServerPathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "safe.txt"), []byte("safe"), 0644)
	require.NoError(t, err)

	srv, err := NewStaticFileServer(tmpDir)
	require.NoError(t, err)

	// Attempt path traversal
	req := httptest.NewRequest(http.MethodGet, "/../../../etc/passwd", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	// Should be 404 (file doesn't exist) or 403 (blocked), not 200
	assert.NotEqual(t, http.StatusOK, rec.Code)
}

func TestStaticFileServerRootAndPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	srv, err := NewStaticFileServer(tmpDir, WithPrefix("/static"))
	require.NoError(t, err)
	assert.NotEmpty(t, srv.Root())
	assert.Equal(t, "/static", srv.Prefix())
}

func TestStaticFileServerHeadMethod(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "data.txt"), []byte("test content"), 0644)
	require.NoError(t, err)

	srv, err := NewStaticFileServer(tmpDir)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodHead, "/data.txt", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- ResponseHelper ---

func TestSendHTML(t *testing.T) {
	rh := NewResponseHelper()
	w := NewMockHTTPResponseWriter()
	rh.SendHTML(w, http.StatusOK, "<h1>Hello</h1>")
	assert.Equal(t, http.StatusOK, w.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", w.HeaderMap.Get("Content-Type"))
	assert.Equal(t, "<h1>Hello</h1>", string(w.Body))
}

func TestSendText(t *testing.T) {
	rh := NewResponseHelper()
	w := NewMockHTTPResponseWriter()
	rh.SendText(w, http.StatusOK, "plain text")
	assert.Equal(t, http.StatusOK, w.StatusCode)
	assert.Equal(t, "text/plain; charset=utf-8", w.HeaderMap.Get("Content-Type"))
	assert.Equal(t, "plain text", string(w.Body))
}

func TestSendRaw(t *testing.T) {
	rh := NewResponseHelper()
	w := NewMockHTTPResponseWriter()
	rh.SendRaw(w, http.StatusCreated, "application/xml", []byte("<root/>"))
	assert.Equal(t, http.StatusCreated, w.StatusCode)
	assert.Equal(t, "application/xml", w.HeaderMap.Get("Content-Type"))
	assert.Equal(t, "<root/>", string(w.Body))
}

func TestSendTemplate(t *testing.T) {
	engine := NewTemplateEngineFromStrings()
	err := engine.LoadTemplateString("page", "<div>{{.Text}}</div>")
	require.NoError(t, err)

	rh := NewResponseHelper()
	w := NewMockHTTPResponseWriter()
	err = rh.SendTemplate(w, engine, "page", http.StatusOK, map[string]interface{}{"Text": "hi"})
	require.NoError(t, err)
	assert.Equal(t, "text/html; charset=utf-8", w.HeaderMap.Get("Content-Type"))
	assert.Equal(t, "<div>hi</div>", string(w.Body))
}

func TestSendFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "doc.txt")
	err := os.WriteFile(filePath, []byte("file content"), 0644)
	require.NoError(t, err)

	rh := NewResponseHelper()
	req := httptest.NewRequest(http.MethodGet, "/doc.txt", nil)
	rec := httptest.NewRecorder()
	err = rh.SendFile(rec, req, filePath)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "file content")
}

func TestSendFileNotFound(t *testing.T) {
	rh := NewResponseHelper()
	req := httptest.NewRequest(http.MethodGet, "/missing.txt", nil)
	rec := httptest.NewRecorder()
	err := rh.SendFile(rec, req, "/nonexistent/path/file.txt")
	assert.Error(t, err)
}

// --- PageData ---

func TestPageData(t *testing.T) {
	pd := PageData{}
	pd.Set("title", "Home").Set("count", 42)
	assert.Equal(t, "Home", pd["title"])
	assert.Equal(t, 42, pd["count"])
}

// --- CacheConfig ---

func TestDefaultCacheConfig(t *testing.T) {
	cc := DefaultCacheConfig()
	assert.Equal(t, 3600, cc.MaxAge)
	assert.NotEmpty(t, cc.ExtensionRules)
	assert.Contains(t, cc.NoCache, ".html")
}

func TestMaxAgeForExtension(t *testing.T) {
	cc := DefaultCacheConfig()

	// HTML is in NoCache list
	age, explicit := cc.MaxAgeForExtension(".html")
	assert.Equal(t, 0, age)
	assert.True(t, explicit)

	// CSS has explicit rule
	age, explicit = cc.MaxAgeForExtension(".css")
	assert.Equal(t, 86400, age)
	assert.True(t, explicit)

	// Unknown extension uses default
	age, explicit = cc.MaxAgeForExtension(".xyz")
	assert.Equal(t, 3600, age)
	assert.False(t, explicit)
}

// --- MockHTTPResponseWriter ---

func TestMockHTTPResponseWriter(t *testing.T) {
	w := NewMockHTTPResponseWriter()
	assert.Equal(t, http.StatusOK, w.StatusCode)
	w.Header().Set("X-Test", "value")
	assert.Equal(t, "value", w.Header().Get("X-Test"))
	w.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, w.StatusCode)
	n, err := w.Write([]byte("body"))
	require.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "body", string(w.Body))
}

// --- isSubPath ---

func TestIsSubPath(t *testing.T) {
	assert.True(t, isSubPath("/var/www", "/var/www"))
	assert.True(t, isSubPath("/var/www", "/var/www/index.html"))
	assert.True(t, isSubPath("/var/www", "/var/www/sub/file.txt"))
	assert.False(t, isSubPath("/var/www", "/var/www-evil/file"))
	assert.False(t, isSubPath("/var/www", "/etc/passwd"))
	assert.False(t, isSubPath("/var/www", "/var/ww"))
}
