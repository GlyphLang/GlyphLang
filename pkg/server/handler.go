package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

// maxRequestBodySize is the maximum allowed request body size (10 MB).
const maxRequestBodySize = 10 << 20

// parseJSONBody parses JSON request body into the context
func parseJSONBody(r *http.Request, ctx *Context) error {
	defer r.Body.Close()

	// Check Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json" // default for missing Content-Type
	}
	if !strings.Contains(contentType, "application/json") {
		return fmt.Errorf("expected application/json content type, got %s", contentType)
	}

	// Limit request body size to prevent denial-of-service
	r.Body = http.MaxBytesReader(ctx.ResponseWriter, r.Body, maxRequestBodySize)

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

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

// handleError logs and sends an error response.
// Full error details are logged server-side but never exposed to clients.
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, statusCode int, message string, err error) {
	// Log the full error detail server-side
	log.Printf("[ERROR] %s %s: %s - %v", r.Method, r.URL.Path, message, err)

	// Send JSON error response without internal details
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   true,
		"message": message,
		"code":    statusCode,
	}

	// Do not expose internal error details to clients
	json.NewEncoder(w).Encode(response)
}
