package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/database"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/parser"
	"github.com/glyphlang/glyph/pkg/server"
	"github.com/glyphlang/glyph/pkg/vm"
	"github.com/glyphlang/glyph/pkg/websocket"
)

// parseSource parses GLYPH source using the Go parser
func parseSource(source string) (*ast.Module, error) {
	// Use Go parser
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("lexer error: %w", err)
	}

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("parser error: %w", err)
	}

	return module, nil
}

// newConfiguredInterpreter creates an interpreter with common configuration
// including mock database for development/demo purposes.
func newConfiguredInterpreter() *interpreter.Interpreter {
	interp := interpreter.NewInterpreter()
	mockDB := database.NewMockDatabase()
	interp.SetDatabaseHandler(mockDB)

	// Set up the parse function for module resolution
	interp.GetModuleResolver().SetParseFunc(func(source string) (*ast.Module, error) {
		lexer := parser.NewLexer(source)
		tokens, err := lexer.Tokenize()
		if err != nil {
			return nil, err
		}
		p := parser.NewParser(tokens)
		return p.Parse()
	})

	return interp
}

// registerRoute registers a route with the router
func registerRoute(router *server.Router, route *ast.Route, interp *interpreter.Interpreter) error {
	handler := createRouteHandler(route, interp)

	serverRoute := &server.Route{
		Method:  convertHTTPMethod(route.Method),
		Path:    route.Path,
		Handler: handler,
	}

	return router.RegisterRoute(serverRoute)
}

// registerCompiledRoute registers a compiled route with the router
func registerCompiledRoute(router *server.Router, route *ast.Route, bytecode []byte, wsHub *websocket.Hub) error {
	handler := createCompiledRouteHandler(route, bytecode, wsHub)

	serverRoute := &server.Route{
		Method:  convertHTTPMethod(route.Method),
		Path:    route.Path,
		Handler: handler,
	}

	return router.RegisterRoute(serverRoute)
}

// createCompiledRouteHandler creates an HTTP handler that executes compiled bytecode
func createCompiledRouteHandler(route *ast.Route, bytecode []byte, wsHub *websocket.Hub) server.RouteHandler {
	return func(ctx *server.Context) error {
		// Create VM instance
		vmInstance := vm.NewVM()

		// Set up WebSocket stats handler if hub is available
		if wsHub != nil {
			wsHandler := websocket.NewVMStatsHandler(wsHub)
			vmInstance.SetWebSocketHandler(wsHandler)
		}

		// Inject path parameters into VM locals
		for key, value := range ctx.PathParams {
			vmInstance.SetLocal(key, vm.StringValue{Val: value})
		}

		// Parse and inject request body as 'input' for POST/PUT/PATCH requests
		if ctx.Request.Method == "POST" || ctx.Request.Method == "PUT" || ctx.Request.Method == "PATCH" {
			contentType := ctx.Request.Header.Get("Content-Type")
			shouldParseJSON := contentType == "" ||
				contentType == "application/json" ||
				(len(contentType) >= 16 && contentType[:16] == "application/json")

			if shouldParseJSON && ctx.Request.Body != nil {
				const maxBodySize = 10 * 1024 * 1024
				limitedReader := io.LimitReader(ctx.Request.Body, maxBodySize)

				var bodyMap map[string]interface{}
				decoder := json.NewDecoder(limitedReader)
				if err := decoder.Decode(&bodyMap); err == nil {
					vmInstance.SetLocal("input", interfaceToValue(bodyMap))
				} else {
					vmInstance.SetLocal("input", vm.NullValue{})
				}
				ctx.Request.Body.Close()
			} else {
				vmInstance.SetLocal("input", vm.NullValue{})
			}
		} else {
			vmInstance.SetLocal("input", vm.NullValue{})
		}

		// Execute compiled bytecode
		result, err := vmInstance.Execute(bytecode)
		if err != nil {
			// Log full error server-side, return generic message to client
			printError(fmt.Errorf("bytecode execution failed: %w", err))
			ctx.StatusCode = http.StatusInternalServerError
			ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
			return json.NewEncoder(ctx.ResponseWriter).Encode(map[string]interface{}{
				"error": "Internal server error",
			})
		}

		// Set response
		ctx.StatusCode = http.StatusOK
		ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(ctx.ResponseWriter).Encode(result)
	}
}

// createRouteHandler creates an HTTP handler for a route
func createRouteHandler(route *ast.Route, interp *interpreter.Interpreter) server.RouteHandler {
	return func(ctx *server.Context) error {
		// Execute route body using the interpreter
		result, err := executeRoute(route, ctx, interp)
		if err != nil {
			// Log full error server-side, return generic message to client
			printError(fmt.Errorf("route execution error: %w", err))
			ctx.StatusCode = http.StatusInternalServerError
			ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
			return json.NewEncoder(ctx.ResponseWriter).Encode(map[string]interface{}{
				"error": "Internal server error",
			})
		}

		// Set response
		ctx.StatusCode = http.StatusOK
		ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(ctx.ResponseWriter).Encode(result)
	}
}

// executeRoute executes a route's body and returns the result
func executeRoute(route *ast.Route, ctx *server.Context, interp *interpreter.Interpreter) (interface{}, error) {
	// Parse request body for POST/PUT/PATCH requests
	var requestBody interface{}
	if ctx.Request.Method == "POST" || ctx.Request.Method == "PUT" || ctx.Request.Method == "PATCH" {
		// Only parse if there's a content type that suggests JSON
		contentType := ctx.Request.Header.Get("Content-Type")
		// Accept empty content-type or application/json
		shouldParseJSON := contentType == "" ||
			contentType == "application/json" ||
			(len(contentType) >= 16 && contentType[:16] == "application/json")

		if shouldParseJSON && ctx.Request.Body != nil {
			// Limit request body size to 10MB to prevent DoS attacks
			const maxBodySize = 10 * 1024 * 1024
			limitedReader := io.LimitReader(ctx.Request.Body, maxBodySize)

			var bodyMap map[string]interface{}
			decoder := json.NewDecoder(limitedReader)
			if err := decoder.Decode(&bodyMap); err != nil {
				// If parsing fails, treat as empty body (could be empty or malformed)
				// Don't return error - just set to nil
				requestBody = nil
			} else {
				requestBody = bodyMap
			}
			ctx.Request.Body.Close()
		}
	}

	// Create request object for interpreter
	request := &interpreter.Request{
		Path:    ctx.Request.URL.Path,
		Method:  ctx.Request.Method,
		Params:  ctx.PathParams,
		Body:    requestBody,
		Headers: make(map[string]string),
	}

	// Copy headers
	for key, values := range ctx.Request.Header {
		if len(values) > 0 {
			request.Headers[key] = values[0]
		}
	}

	// Use ExecuteRoute instead of ExecuteRouteSimple to handle request body
	response, err := interp.ExecuteRoute(route, request)
	if err != nil {
		return nil, fmt.Errorf("route execution failed: %w", err)
	}

	return response.Body, nil
}

// createHandler creates the main HTTP handler.
// CORS is configured via the GLYPH_CORS_ORIGIN environment variable.
// Set it to a specific origin (e.g. "http://localhost:8080") or "*" to
// allow all origins (credentials will be disabled). When unset, no CORS
// headers are added.
func createHandler(router *server.Router) http.HandlerFunc {
	corsOrigin := os.Getenv("GLYPH_CORS_ORIGIN")

	return func(w http.ResponseWriter, r *http.Request) {
		// Apply CORS headers when configured
		if corsOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if corsOrigin == "*" {
				w.Header().Set("Access-Control-Allow-Credentials", "false")
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		method := server.HTTPMethod(r.Method)
		route, params, err := router.Match(method, r.URL.Path)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Route not found",
				"path":  r.URL.Path,
			})
			return
		}

		// Create context
		ctx := &server.Context{
			Request:        r,
			ResponseWriter: w,
			PathParams:     params,
			StatusCode:     http.StatusOK,
		}

		// Execute handler
		if err := route.Handler(ctx); err != nil {
			printError(fmt.Errorf("handler error for %s %s: %w", r.Method, r.URL.Path, err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal server error",
			})
		}
	}
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log request
		printRequest(r.Method, r.URL.Path)

		// Call next handler
		next.ServeHTTP(w, r)

		// Log duration
		duration := time.Since(start)
		printDuration(duration)
	})
}

// signalNotify is a helper to register for interrupt signals.
// Extracted for use by both waitForShutdown and hotReloadManager.waitForShutdown.
func signalNotify(sigChan chan<- os.Signal) {
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
}

// waitForShutdown waits for interrupt signal and gracefully shuts down the server
func waitForShutdown(srv *http.Server) error {
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signalNotify(sigChan)

	// Wait for signal
	<-sigChan

	printWarning("\nShutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	printSuccess("Server stopped gracefully")
	return nil
}

// convertHTTPMethod converts ast.HttpMethod to server.HTTPMethod
func convertHTTPMethod(method ast.HttpMethod) server.HTTPMethod {
	switch method {
	case ast.Get:
		return server.GET
	case ast.Post:
		return server.POST
	case ast.Put:
		return server.PUT
	case ast.Delete:
		return server.DELETE
	case ast.Patch:
		return server.PATCH
	default:
		return server.GET
	}
}

// changeExtension changes the file extension
func changeExtension(path, newExt string) string {
	ext := filepath.Ext(path)
	return path[:len(path)-len(ext)] + newExt
}

// Template content getters
func getHelloWorldTemplate() string {
	return `# Hello World Example
# This is a simple GLYPH program

: Message {
  text: str!
  timestamp: int
}

@ GET /hello
  > {text: "Hello, World!", timestamp: 1234567890}

@ GET /greet/:name -> Message
  $ message = {
    text: "Hello, " + name + "!",
    timestamp: time.now()
  }
  > message
`
}

func getRestAPITemplate() string {
	return `# Example REST API in GLYPH

: User {
  id: int!
  name: str!
  email: str!
}

@ GET /api/users -> List[User]
  + auth(jwt)
  > []

@ GET /health
  > {status: "ok", timestamp: now()}
`
}

// Pretty printing functions
var (
	infoColor    = color.New(color.FgCyan)
	successColor = color.New(color.FgGreen)
	warningColor = color.New(color.FgYellow)
	errorColor   = color.New(color.FgRed)
	requestColor = color.New(color.FgMagenta)
)

func printInfo(msg string) {
	infoColor.Printf("[INFO] %s\n", msg)
}

func printSuccess(msg string) {
	successColor.Printf("[SUCCESS] %s\n", msg)
}

func printWarning(msg string) {
	warningColor.Printf("[WARNING] %s\n", msg)
}

func printError(err error) {
	errorColor.Printf("[ERROR] %s\n", err.Error())
}

func printRequest(method, path string) {
	requestColor.Printf("[%s] %s ", method, path)
}

func printDuration(d time.Duration) {
	fmt.Printf("(%dms)\n", d.Milliseconds())
}

// openInEditor opens the specified file in the default text editor
func openInEditor(filePath string) error {
	// Get absolute path for safety
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// On Windows, try VS Code first, then fall back to notepad
		// We can't use "start" with the file path because .glyph files are associated
		// with glyph.exe itself, which would cause an infinite loop
		if codePath, err := exec.LookPath("code"); err == nil {
			cmd = exec.Command(codePath, absPath) //#nosec G204 -- codePath from LookPath, absPath validated
		} else {
			cmd = exec.Command("notepad", absPath) //#nosec G204 -- absPath validated via filepath.Abs
		}
	case "darwin":
		// On macOS, use 'open -e' for TextEdit or just 'open' for default app
		cmd = exec.Command("open", "-t", absPath) //#nosec G204 -- absPath validated via filepath.Abs
	default:
		// On Linux, try common editors in order of preference
		editors := []string{"code", "gedit", "kate", "xed", "nano", "vi"}
		for _, editor := range editors {
			if _, err := exec.LookPath(editor); err == nil {
				cmd = exec.Command(editor, absPath) //#nosec G204 -- editor from hardcoded list, absPath validated
				break
			}
		}
		if cmd == nil {
			// Fallback to xdg-open which should open in default text editor
			cmd = exec.Command("xdg-open", absPath) //#nosec G204 -- absPath validated via filepath.Abs
		}
	}

	return cmd.Start()
}

// openURL opens the specified URL in the default browser
func openURL(urlStr string) error {
	// Validate URL to prevent command injection
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow http and https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only http and https URLs are allowed, got: %s", parsedURL.Scheme)
	}

	// Ensure the URL has a valid host
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", urlStr)
	case "darwin":
		cmd = exec.Command("open", urlStr)
	default: // Linux and other Unix-like
		cmd = exec.Command("xdg-open", urlStr)
	}

	return cmd.Start()
}
