// Package web provides web page support for GlyphLang: static file serving,
// HTML template rendering, content-type-aware response helpers, and template
// management with auto-escaping for XSS prevention.
package web

import (
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ResponseType indicates the content type of a response.
type ResponseType string

const (
	JSON ResponseType = "json"
	HTML ResponseType = "html"
	Text ResponseType = "text"
	File ResponseType = "file"
)

// ContentTypeForResponse returns the MIME content-type for a ResponseType.
func ContentTypeForResponse(rt ResponseType) string {
	switch rt {
	case HTML:
		return "text/html; charset=utf-8"
	case Text:
		return "text/plain; charset=utf-8"
	case JSON:
		return "application/json; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

// TemplateEngine manages HTML templates with auto-escaping.
type TemplateEngine struct {
	mu        sync.RWMutex
	templates map[string]*template.Template
	baseDir   string
	funcMap   template.FuncMap
}

// NewTemplateEngine creates a new template engine.
// baseDir is the directory to load template files from.
// It validates that baseDir exists and is a directory.
func NewTemplateEngine(baseDir string) (*TemplateEngine, error) {
	info, err := os.Stat(baseDir)
	if err != nil {
		return nil, fmt.Errorf("template base directory does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("template base path is not a directory: %s", baseDir)
	}
	return &TemplateEngine{
		templates: make(map[string]*template.Template),
		baseDir:   baseDir,
		funcMap:   defaultFuncMap(),
	}, nil
}

// NewTemplateEngineFromStrings creates a template engine that only uses
// string-based templates (no file loading), bypassing directory validation.
func NewTemplateEngineFromStrings() *TemplateEngine {
	return &TemplateEngine{
		templates: make(map[string]*template.Template),
		funcMap:   defaultFuncMap(),
	}
}

// defaultFuncMap returns built-in template functions.
func defaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"join":  strings.Join,
	}
}

// copyFuncMap returns a shallow copy of a template.FuncMap to prevent
// race conditions when AddFunc is called concurrently with template loading.
func copyFuncMap(src template.FuncMap) template.FuncMap {
	dst := make(template.FuncMap, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// AddFunc registers a custom template function.
// Must be called before loading templates that use the function.
func (e *TemplateEngine) AddFunc(name string, fn interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.funcMap[name] = fn
}

// LoadTemplate parses and caches a template file.
func (e *TemplateEngine) LoadTemplate(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	filePath := filepath.Join(e.baseDir, name)
	fmap := copyFuncMap(e.funcMap)
	tmpl, err := template.New(filepath.Base(name)).Funcs(fmap).ParseFiles(filePath)
	if err != nil {
		return fmt.Errorf("failed to load template %s: %w", name, err)
	}
	e.templates[name] = tmpl
	return nil
}

// LoadTemplateString parses and caches a template from a string.
func (e *TemplateEngine) LoadTemplateString(name, content string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	fmap := copyFuncMap(e.funcMap)
	tmpl, err := template.New(name).Funcs(fmap).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	e.templates[name] = tmpl
	return nil
}

// Render executes a template with the given data and writes to w.
func (e *TemplateEngine) Render(w io.Writer, name string, data interface{}) error {
	e.mu.RLock()
	tmpl, ok := e.templates[name]
	e.mu.RUnlock()
	if !ok {
		return fmt.Errorf("template %q not found", name)
	}
	return tmpl.Execute(w, data)
}

// RenderString executes a template and returns the result as a string.
func (e *TemplateEngine) RenderString(name string, data interface{}) (string, error) {
	var buf strings.Builder
	if err := e.Render(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// HasTemplate checks if a template is registered.
func (e *TemplateEngine) HasTemplate(name string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ok := e.templates[name]
	return ok
}

// TemplateNames returns the names of all registered templates.
func (e *TemplateEngine) TemplateNames() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	names := make([]string, 0, len(e.templates))
	for name := range e.templates {
		names = append(names, name)
	}
	return names
}

// StaticFileServer serves static files from a directory with path-traversal protection.
// Uses filepath.EvalSymlinks to resolve symlinks and prevent escape from the root.
type StaticFileServer struct {
	absRoot   string // Absolute, symlink-resolved path to the root directory
	prefix    string
	indexFile string
	maxAge    int // Cache-Control max-age in seconds
	allowList bool
}

// StaticOption configures the static file server.
type StaticOption func(*StaticFileServer)

// WithPrefix sets the URL prefix to strip before serving files.
func WithPrefix(prefix string) StaticOption {
	return func(s *StaticFileServer) {
		s.prefix = strings.TrimSuffix(prefix, "/")
	}
}

// WithIndex sets the default index file name.
func WithIndex(filename string) StaticOption {
	return func(s *StaticFileServer) {
		s.indexFile = filename
	}
}

// WithMaxAge sets the Cache-Control max-age header in seconds.
func WithMaxAge(seconds int) StaticOption {
	return func(s *StaticFileServer) {
		s.maxAge = seconds
	}
}

// WithDirectoryListing enables directory listing.
func WithDirectoryListing(allow bool) StaticOption {
	return func(s *StaticFileServer) {
		s.allowList = allow
	}
}

// NewStaticFileServer creates a new static file server for the given directory.
// The root directory is resolved via filepath.EvalSymlinks to an absolute canonical
// path to prevent path traversal and symlink escape attacks.
func NewStaticFileServer(root string, opts ...StaticOption) (*StaticFileServer, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root path: %w", err)
	}
	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root symlinks: %w", err)
	}
	s := &StaticFileServer{
		absRoot:   realRoot,
		indexFile: "index.html",
	}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

// isSubPath returns true if child is equal to or a subdirectory/file of parent.
// Both paths must be absolute and already symlink-resolved.
func isSubPath(parent, child string) bool {
	// Exact match means child IS the root directory itself
	if child == parent {
		return true
	}
	// child must start with parent followed by a path separator
	// This prevents /var/www matching /var/www-evil
	return strings.HasPrefix(child, parent+string(filepath.Separator))
}

// ServeHTTP serves static files with path-traversal and symlink-escape protection.
// After resolving the file path, filepath.EvalSymlinks verifies the real
// path remains within the root directory.
func (s *StaticFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	urlPath := r.URL.Path
	if s.prefix != "" {
		urlPath = strings.TrimPrefix(urlPath, s.prefix)
	}
	// Clean the URL path to normalize slashes and remove traversal sequences
	urlPath = path.Clean("/" + urlPath)

	// Resolve the full filesystem path
	candidatePath := filepath.Join(s.absRoot, filepath.FromSlash(urlPath))

	// Resolve symlinks to get the real path on disk
	realPath, err := filepath.EvalSymlinks(candidatePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Path traversal protection: verify resolved real path is under absRoot
	if !isSubPath(s.absRoot, realPath) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	fileInfo, err := os.Stat(realPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Handle directory requests
	if fileInfo.IsDir() {
		indexPath := filepath.Join(realPath, s.indexFile)
		if idxInfo, idxErr := os.Stat(indexPath); idxErr == nil && !idxInfo.IsDir() {
			realPath = indexPath
		} else if s.allowList {
			s.serveDirectoryListing(w, urlPath, realPath)
			return
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	// Open and serve the file
	file, err := os.Open(realPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Detect content type from file extension
	fileExt := filepath.Ext(realPath)
	contentType := mime.TypeByExtension(fileExt)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)

	if s.maxAge > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", s.maxAge))
	}

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

// serveDirectoryListing writes a simple HTML directory listing with XSS-safe output.
func (s *StaticFileServer) serveDirectoryListing(w http.ResponseWriter, urlPath, dirPath string) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	sort.Strings(names)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	escapedPath := template.HTMLEscapeString(urlPath)
	fmt.Fprintf(w, "<html><head><title>%s</title></head><body>", escapedPath)
	fmt.Fprintf(w, "<h1>Index of %s</h1><ul>", escapedPath)
	for _, name := range names {
		href := path.Join(urlPath, name)
		fmt.Fprintf(w, "<li><a href=\"%s\">%s</a></li>",
			template.HTMLEscapeString(href), template.HTMLEscapeString(name))
	}
	fmt.Fprint(w, "</ul></body></html>")
}

// Root returns the absolute root directory of the static file server.
func (s *StaticFileServer) Root() string {
	return s.absRoot
}

// Prefix returns the URL prefix.
func (s *StaticFileServer) Prefix() string {
	return s.prefix
}

// ResponseHelper provides helpers for writing typed HTTP responses.
type ResponseHelper struct{}

// NewResponseHelper creates a new ResponseHelper.
func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{}
}

// SendHTML writes an HTML response with the given status code.
func (rh *ResponseHelper) SendHTML(w http.ResponseWriter, status int, htmlContent string) {
	w.Header().Set("Content-Type", ContentTypeForResponse(HTML))
	w.WriteHeader(status)
	io.WriteString(w, htmlContent)
}

// SendText writes a plain text response with the given status code.
func (rh *ResponseHelper) SendText(w http.ResponseWriter, status int, textContent string) {
	w.Header().Set("Content-Type", ContentTypeForResponse(Text))
	w.WriteHeader(status)
	io.WriteString(w, textContent)
}

// SendFile sends a file as the response. Content type is detected from the filename extension.
// The rootDir parameter restricts file access to the given directory tree. If rootDir is empty,
// the current working directory is used.
func (rh *ResponseHelper) SendFile(w http.ResponseWriter, r *http.Request, rootDir, targetPath string) error {
	if rootDir == "" {
		var err error
		rootDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to determine working directory: %w", err)
		}
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("failed to resolve root directory: %w", err)
	}

	// Clean and resolve the target path relative to root
	cleanPath := filepath.Clean(targetPath)
	if !filepath.IsAbs(cleanPath) {
		cleanPath = filepath.Join(absRoot, cleanPath)
	}

	// Resolve symlinks to get the real path on disk
	realPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", targetPath)
		}
		return fmt.Errorf("failed to resolve file path: %w", err)
	}

	// Path traversal protection: verify resolved path is under the root
	if !isSubPath(absRoot, realPath) {
		return fmt.Errorf("access denied: path escapes root directory")
	}

	file, err := os.Open(realPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", targetPath, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", targetPath, err)
	}

	if stat.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", targetPath)
	}

	fileExt := filepath.Ext(realPath)
	contentType := mime.TypeByExtension(fileExt)
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
	return nil
}

// SendTemplate renders a named template and writes it as an HTML response.
func (rh *ResponseHelper) SendTemplate(w http.ResponseWriter, engine *TemplateEngine, name string, status int, data interface{}) error {
	w.Header().Set("Content-Type", ContentTypeForResponse(HTML))
	w.WriteHeader(status)
	return engine.Render(w, name, data)
}

// SendRaw writes raw bytes with a specified content type.
func (rh *ResponseHelper) SendRaw(w http.ResponseWriter, status int, contentType string, body []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	w.Write(body)
}

// PageData is a convenience type for template data.
type PageData map[string]interface{}

// Set adds a key-value pair to the page data and returns the PageData for chaining.
func (pd PageData) Set(key string, value interface{}) PageData {
	pd[key] = value
	return pd
}

// CacheConfig holds cache configuration for static files.
type CacheConfig struct {
	MaxAge         int            // Default max-age in seconds
	ExtensionRules map[string]int // Per-extension max-age overrides
	NoCache        []string       // Extensions that should not be cached
}

// DefaultCacheConfig returns a reasonable default cache configuration.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxAge: 3600,
		ExtensionRules: map[string]int{
			".html":  0,
			".css":   86400,
			".js":    86400,
			".png":   604800,
			".jpg":   604800,
			".gif":   604800,
			".svg":   604800,
			".ico":   604800,
			".woff":  2592000,
			".woff2": 2592000,
		},
		NoCache: []string{".html"},
	}
}

// MaxAgeForExtension returns the max-age for a file extension and whether it was an explicit rule.
func (cc CacheConfig) MaxAgeForExtension(ext string) (int, bool) {
	for _, nc := range cc.NoCache {
		if nc == ext {
			return 0, true
		}
	}
	if age, ok := cc.ExtensionRules[ext]; ok {
		return age, true
	}
	return cc.MaxAge, false
}

// MockHTTPResponseWriter is a minimal mock for testing HTTP responses.
type MockHTTPResponseWriter struct {
	StatusCode int
	HeaderMap  http.Header
	Body       []byte
}

// NewMockHTTPResponseWriter creates a new MockHTTPResponseWriter.
func NewMockHTTPResponseWriter() *MockHTTPResponseWriter {
	return &MockHTTPResponseWriter{
		HeaderMap:  make(http.Header),
		StatusCode: http.StatusOK,
	}
}

func (m *MockHTTPResponseWriter) Header() http.Header {
	return m.HeaderMap
}

func (m *MockHTTPResponseWriter) Write(b []byte) (int, error) {
	m.Body = append(m.Body, b...)
	return len(b), nil
}

func (m *MockHTTPResponseWriter) WriteHeader(statusCode int) {
	m.StatusCode = statusCode
}

// fileModTime returns a fixed time for testing.
func fileModTime() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}
