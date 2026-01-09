package server

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StaticRoute represents a static file serving route
type StaticRoute struct {
	URLPath  string // URL path prefix (e.g., "/assets")
	FilePath string // Filesystem path (e.g., "./public/")
}

// StaticHandler handles static file serving
type StaticHandler struct {
	routes []StaticRoute
}

// NewStaticHandler creates a new static file handler
func NewStaticHandler() *StaticHandler {
	return &StaticHandler{
		routes: []StaticRoute{},
	}
}

// AddRoute adds a static route
func (h *StaticHandler) AddRoute(urlPath, filePath string) {
	// Ensure URL path starts with /
	if !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}
	// Ensure file path is clean
	filePath = filepath.Clean(filePath)

	h.routes = append(h.routes, StaticRoute{
		URLPath:  urlPath,
		FilePath: filePath,
	})
}

// Match tries to match a request path to a static route
// Returns the matched route and remaining path, or nil if no match
func (h *StaticHandler) Match(requestPath string) (*StaticRoute, string) {
	for _, route := range h.routes {
		if strings.HasPrefix(requestPath, route.URLPath) {
			// Get the remaining path after the URL prefix
			remaining := strings.TrimPrefix(requestPath, route.URLPath)
			if remaining == "" {
				remaining = "/"
			}
			return &route, remaining
		}
	}
	return nil, ""
}

// ServeHTTP implements http.Handler for static files
func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route, remainingPath := h.Match(r.URL.Path)
	if route == nil {
		http.NotFound(w, r)
		return
	}

	// Build full filesystem path
	fullPath := filepath.Join(route.FilePath, remainingPath)
	fullPath = filepath.Clean(fullPath)

	// Security: Ensure the path is still within the base directory
	absBase, err := filepath.Abs(route.FilePath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !strings.HasPrefix(absPath, absBase) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if path exists
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If it's a directory, try to serve index.html
	if info.IsDir() {
		indexPath := filepath.Join(fullPath, "index.html")
		if indexInfo, err := os.Stat(indexPath); err == nil && !indexInfo.IsDir() {
			fullPath = indexPath
			info = indexInfo
		} else {
			// No index.html, return forbidden (no directory listing)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	// Serve the file
	h.serveFile(w, r, fullPath, info)
}

// serveFile serves a single file
func (h *StaticHandler) serveFile(w http.ResponseWriter, r *http.Request, path string, info os.FileInfo) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Detect content type
	ext := filepath.Ext(path)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Set headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

	// Handle caching headers
	modTime := info.ModTime()
	w.Header().Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))

	// Check If-Modified-Since header
	if ifModSince := r.Header.Get("If-Modified-Since"); ifModSince != "" {
		if t, err := http.ParseTime(ifModSince); err == nil {
			if modTime.Before(t.Add(1 * 1e9)) { // Within 1 second
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	// Serve content
	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)
}

// CreateStaticMiddleware creates a middleware that serves static files
func CreateStaticMiddleware(handler *StaticHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to match static route
			route, _ := handler.Match(r.URL.Path)
			if route != nil {
				handler.ServeHTTP(w, r)
				return
			}
			// Not a static route, pass to next handler
			next.ServeHTTP(w, r)
		})
	}
}
