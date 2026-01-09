package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Handler creates an HTTP handler function for routing
type Handler struct {
	router      *Router
	interpreter Interpreter
}

// NewHandler creates a new handler with the given router and interpreter
func NewHandler(router *Router, interpreter Interpreter) *Handler {
	return &Handler{
		router:      router,
		interpreter: interpreter,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Convert HTTP method to our HTTPMethod type
	method := HTTPMethod(strings.ToUpper(r.Method))

	// Try to match the route
	route, pathParams, err := h.router.Match(method, r.URL.Path)
	if err != nil {
		h.handleError(w, r, http.StatusNotFound, "route not found", err)
		return
	}

	// Create context
	ctx := &Context{
		Request:        r,
		ResponseWriter: w,
		PathParams:     pathParams,
		QueryParams:    parseQueryParams(r),
		StatusCode:     http.StatusOK,
	}

	// Parse JSON body if present
	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
		if err := parseJSONBody(r, ctx); err != nil {
			h.handleError(w, r, http.StatusBadRequest, "invalid JSON body", err)
			return
		}
	}

	// Apply middlewares and execute handler
	handler := route.Handler
	if handler == nil && h.interpreter != nil {
		// Use interpreter if no custom handler
		handler = func(ctx *Context) error {
			result, err := h.interpreter.Execute(route, ctx)
			if err != nil {
				return err
			}

			// Check for special response objects (from html(), file(), template(), text() functions)
			if responseType, content, ok := IsSpecialResponse(result); ok {
				return SendResponse(ctx, responseType, content)
			}

			// Check route's response format
			if route.ResponseFormat != "" {
				return SendResponse(ctx, route.ResponseFormat, result)
			}

			// Default to JSON
			return sendJSONResponse(ctx, result)
		}
	}

	// Apply middlewares in reverse order
	for i := len(route.Middlewares) - 1; i >= 0; i-- {
		handler = route.Middlewares[i](handler)
	}

	// Execute the handler
	if err := handler(ctx); err != nil {
		h.handleError(w, r, http.StatusInternalServerError, "handler error", err)
		return
	}
}

// parseQueryParams extracts all query parameter values from the request
func parseQueryParams(r *http.Request) map[string][]string {
	return r.URL.Query()
}

// parseJSONBody parses JSON request body into the context
func parseJSONBody(r *http.Request, ctx *Context) error {
	// Check Content-Type
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") && contentType != "" {
		return fmt.Errorf("expected application/json content type, got %s", contentType)
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}
	defer r.Body.Close()

	// Skip empty bodies
	if len(body) == 0 {
		ctx.Body = make(map[string]interface{})
		return nil
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	ctx.Body = data
	return nil
}

// sendJSONResponse sends a JSON response
func sendJSONResponse(ctx *Context, data interface{}) error {
	ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	ctx.ResponseWriter.WriteHeader(ctx.StatusCode)

	encoder := json.NewEncoder(ctx.ResponseWriter)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON response: %w", err)
	}

	return nil
}

// SendJSON is a helper to send JSON responses from handlers
func SendJSON(ctx *Context, statusCode int, data interface{}) error {
	ctx.StatusCode = statusCode
	return sendJSONResponse(ctx, data)
}

// SendError is a helper to send error responses
func SendError(ctx *Context, statusCode int, message string) error {
	return SendJSON(ctx, statusCode, map[string]interface{}{
		"error":   true,
		"message": message,
		"code":    statusCode,
	})
}

// SendHTML sends an HTML response
func SendHTML(ctx *Context, statusCode int, html string) error {
	ctx.ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	ctx.ResponseWriter.WriteHeader(statusCode)
	_, err := ctx.ResponseWriter.Write([]byte(html))
	if err != nil {
		return fmt.Errorf("failed to write HTML response: %w", err)
	}
	return nil
}

// SendText sends a plain text response
func SendText(ctx *Context, statusCode int, text string) error {
	ctx.ResponseWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.ResponseWriter.WriteHeader(statusCode)
	_, err := ctx.ResponseWriter.Write([]byte(text))
	if err != nil {
		return fmt.Errorf("failed to write text response: %w", err)
	}
	return nil
}

// SendBytes sends a raw byte response with specified content type
func SendBytes(ctx *Context, statusCode int, contentType string, data []byte) error {
	ctx.ResponseWriter.Header().Set("Content-Type", contentType)
	ctx.ResponseWriter.WriteHeader(statusCode)
	_, err := ctx.ResponseWriter.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write bytes response: %w", err)
	}
	return nil
}

// SendFile serves a file from the filesystem
func SendFile(ctx *Context, filePath string) error {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(filePath)

	// Check if file exists
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return SendError(ctx, http.StatusNotFound, "file not found")
		}
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Don't serve directories
	if info.IsDir() {
		return SendError(ctx, http.StatusForbidden, "cannot serve directory")
	}

	// Open the file
	file, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Detect content type from extension
	ext := filepath.Ext(cleanPath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Set headers
	ctx.ResponseWriter.Header().Set("Content-Type", contentType)
	ctx.ResponseWriter.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	ctx.ResponseWriter.WriteHeader(http.StatusOK)

	// Stream the file
	_, err = io.Copy(ctx.ResponseWriter, file)
	if err != nil {
		return fmt.Errorf("failed to send file: %w", err)
	}

	return nil
}

// ResponseType constants for special response handling
const (
	ResponseTypeJSON = "json"
	ResponseTypeHTML = "html"
	ResponseTypeText = "text"
	ResponseTypeFile = "file"
)

// IsSpecialResponse checks if the result is a special response object
func IsSpecialResponse(result interface{}) (string, interface{}, bool) {
	m, ok := result.(map[string]interface{})
	if !ok {
		return "", nil, false
	}

	responseType, ok := m["__responseType"].(string)
	if !ok {
		return "", nil, false
	}

	content := m["content"]
	return responseType, content, true
}

// SendResponse sends a response based on the response type
func SendResponse(ctx *Context, responseType string, result interface{}) error {
	switch responseType {
	case ResponseTypeHTML:
		html, ok := result.(string)
		if !ok {
			return fmt.Errorf("HTML response content must be a string")
		}
		return SendHTML(ctx, ctx.StatusCode, html)

	case ResponseTypeText:
		text, ok := result.(string)
		if !ok {
			return fmt.Errorf("text response content must be a string")
		}
		return SendText(ctx, ctx.StatusCode, text)

	case ResponseTypeFile:
		filePath, ok := result.(string)
		if !ok {
			return fmt.Errorf("file response content must be a file path string")
		}
		return SendFile(ctx, filePath)

	case ResponseTypeJSON:
		fallthrough
	default:
		return sendJSONResponse(ctx, result)
	}
}

// handleError logs and sends an error response
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, statusCode int, message string, err error) {
	// Log the error
	log.Printf("[ERROR] %s %s: %s - %v", r.Method, r.URL.Path, message, err)

	// Send JSON error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   true,
		"message": message,
		"code":    statusCode,
	}

	if err != nil {
		response["details"] = err.Error()
	}

	json.NewEncoder(w).Encode(response)
}
