package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/lsp"
	"github.com/glyphlang/glyph/pkg/parser"
	"github.com/glyphlang/glyph/pkg/server"
	"github.com/glyphlang/glyph/pkg/vm"
	"github.com/glyphlang/glyph/pkg/websocket"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var version = "1.0.0"

func main() {
	var rootCmd = &cobra.Command{
		Use:   "glyph",
		Short: "AI Backend Compiler - A language for AI-generated backends",
		Long: `GLYPH is a programming language specifically designed for AI agents
to rapidly build high-performance, secure backend applications.`,
		Version: version,
	}

	// Compile command
	var compileCmd = &cobra.Command{
		Use:   "compile <file>",
		Short: "Compile source code to binary format",
		Args:  cobra.ExactArgs(1),
		RunE:  runCompile,
	}
	compileCmd.Flags().StringP("output", "o", "", "Output file")
	compileCmd.Flags().Uint8P("opt-level", "O", 2, "Optimization level (0-3)")

	// Decompile command
	var decompileCmd = &cobra.Command{
		Use:   "decompile <file>",
		Short: "Decompile binary format to source code",
		Args:  cobra.ExactArgs(1),
		RunE:  runDecompile,
	}
	decompileCmd.Flags().StringP("output", "o", "", "Output file")

	// Run command
	var runCmd = &cobra.Command{
		Use:   "run <file>",
		Short: "Run GLYPH source file",
		Args:  cobra.ExactArgs(1),
		RunE:  runRun,
	}
	runCmd.Flags().Uint16P("port", "p", 3000, "Port to listen on")
	runCmd.Flags().Bool("bytecode", false, "Execute bytecode (.glybc) file")
	runCmd.Flags().Bool("interpret", false, "Use tree-walking interpreter instead of compiler (fallback mode)")

	// Dev command
	var devCmd = &cobra.Command{
		Use:   "dev <file>",
		Short: "Start development server with hot reload",
		Args:  cobra.ExactArgs(1),
		RunE:  runDev,
	}
	devCmd.Flags().Uint16P("port", "p", 3000, "Port to listen on")
	devCmd.Flags().BoolP("watch", "w", true, "Watch for file changes")

	// Init command
	var initCmd = &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize new project",
		Args:  cobra.ExactArgs(1),
		RunE:  runInit,
	}
	initCmd.Flags().StringP("template", "t", "rest-api", "Project template")

	// LSP command
	var lspCmd = &cobra.Command{
		Use:   "lsp",
		Short: "Start Language Server Protocol server",
		RunE:  runLSP,
	}
	lspCmd.Flags().StringP("log", "l", "", "Log file for debugging (optional)")

	// Version command (built-in, but we can customize)
	rootCmd.SetVersionTemplate(`GLYPH version {{.Version}}
`)

	// Add commands to root
	rootCmd.AddCommand(compileCmd)
	rootCmd.AddCommand(decompileCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(lspCmd)

	if err := rootCmd.Execute(); err != nil {
		printError(err)
		os.Exit(1)
	}
}

// runCompile handles the compile command
func runCompile(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	output, _ := cmd.Flags().GetString("output")
	optLevel, _ := cmd.Flags().GetUint8("opt-level")

	printInfo(fmt.Sprintf("Compiling %s... (opt-level: %d)", filePath, optLevel))

	// Read source file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Measure compilation time
	start := time.Now()

	// Parse source code
	module, err := parseSource(string(source))
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}

	// Determine optimization level
	var optLevelEnum compiler.OptimizationLevel
	switch optLevel {
	case 0:
		optLevelEnum = compiler.OptNone
	case 1, 2:
		optLevelEnum = compiler.OptBasic
	case 3:
		optLevelEnum = compiler.OptAggressive
	default:
		optLevelEnum = compiler.OptBasic
	}

	// Compile module (for now, just compile the first route)
	c := compiler.NewCompilerWithOptLevel(optLevelEnum)
	var bytecode []byte

	// Find first route and compile it
	for _, item := range module.Items {
		if route, ok := item.(*interpreter.Route); ok {
			bytecode, err = c.CompileRoute(route)
			if err != nil {
				return fmt.Errorf("compilation failed: %w", err)
			}
			break
		}
	}

	if bytecode == nil {
		return fmt.Errorf("no routes found in module")
	}

	compilationTime := time.Since(start)

	// Determine output path
	if output == "" {
		output = changeExtension(filePath, ".glybc")
	}

	// Write bytecode to file
	if err := os.WriteFile(output, bytecode, 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	printSuccess(fmt.Sprintf("Compiled to %s", output))
	printInfo(fmt.Sprintf("Compilation time: %s", compilationTime))
	printInfo(fmt.Sprintf("Output size: %d bytes", len(bytecode)))
	printInfo(fmt.Sprintf("Source size: %d bytes", len(source)))
	compressionRatio := (1.0 - float64(len(bytecode))/float64(len(source))) * 100.0
	if compressionRatio > 0 {
		printInfo(fmt.Sprintf("Compression: %.1f%%", compressionRatio))
	}

	return nil
}

// runDecompile handles the decompile command
func runDecompile(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	output, _ := cmd.Flags().GetString("output")

	printInfo(fmt.Sprintf("Decompiling %s...", filePath))

	// Read bytecode file
	bytecode, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// For now, decompile just returns the source
	// (Full bytecode decompilation coming soon)
	source := string(bytecode)

	// Determine output path
	if output == "" {
		output = changeExtension(filePath, ".abc")
	}

	// Write source to file
	if err := os.WriteFile(output, []byte(source), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	printSuccess(fmt.Sprintf("Decompiled to %s", output))
	printInfo(fmt.Sprintf("Output size: %d bytes", len(source)))

	return nil
}

// runRun handles the run command
func runRun(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	port, _ := cmd.Flags().GetUint16("port")
	useBytecode, _ := cmd.Flags().GetBool("bytecode")
	useInterpreter, _ := cmd.Flags().GetBool("interpret")

	// Check if file is bytecode based on extension or flag
	if !useBytecode {
		useBytecode = filepath.Ext(filePath) == ".glybc"
	}

	if useBytecode {
		printInfo(fmt.Sprintf("Running bytecode %s...", filePath))

		// Read bytecode file
		bytecode, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read bytecode file: %w", err)
		}

		// Execute bytecode using VM
		start := time.Now()
		vmInstance := vm.NewVM()
		result, err := vmInstance.Execute(bytecode)
		execTime := time.Since(start)

		if err != nil {
			return fmt.Errorf("bytecode execution failed: %w", err)
		}

		printSuccess("Bytecode executed successfully")
		printInfo(fmt.Sprintf("Execution time: %s", execTime))
		printInfo(fmt.Sprintf("Result: %v", result))

		// For now, just print the result
		// TODO: Implement server mode for bytecode execution
		return nil
	}

	// Running source file
	if useInterpreter {
		// Use tree-walking interpreter (fallback mode)
		printInfo(fmt.Sprintf("Running %s with interpreter...", filePath))
		srv, err := startServerWithInterpreter(filePath, int(port))
		if err != nil {
			return err
		}
		return waitForShutdown(srv)
	}

	// Default: compile and run with VM
	printInfo(fmt.Sprintf("Compiling and running %s...", filePath))
	srv, err := startServerWithCompiler(filePath, int(port))
	if err != nil {
		// Fall back to interpreter on compilation error
		printWarning(fmt.Sprintf("Compilation failed: %v", err))
		printWarning("Falling back to interpreter mode...")
		srv, err = startServerWithInterpreter(filePath, int(port))
		if err != nil {
			return err
		}
	}

	// Setup graceful shutdown
	return waitForShutdown(srv)
}

// runDev handles the dev command
func runDev(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	port, _ := cmd.Flags().GetUint16("port")
	watch, _ := cmd.Flags().GetBool("watch")

	printInfo(fmt.Sprintf("Starting development server on port %d...", port))

	// Get absolute path for watching
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Start initial server (use compiler by default)
	srv, err := startServerWithCompiler(absPath, int(port))
	if err != nil {
		// Fall back to interpreter if compilation fails
		printWarning(fmt.Sprintf("Compilation failed: %v", err))
		printWarning("Falling back to interpreter mode...")
		srv, err = startServerWithInterpreter(absPath, int(port))
		if err != nil {
			return err
		}
	}

	// Setup file watching if enabled
	if watch {
		printInfo(fmt.Sprintf("Watching %s for changes...", absPath))
		go watchFile(absPath, func() {
			printWarning("File changed, reloading...")
			// In a real implementation, we'd reload the server here
			// For now, just notify the user
			printInfo("Hot reload triggered (server restart not yet implemented)")
		})
	}

	// Setup graceful shutdown
	return waitForShutdown(srv)
}

// runInit handles the init command
func runInit(cmd *cobra.Command, args []string) error {
	name := args[0]
	template, _ := cmd.Flags().GetString("template")

	printInfo(fmt.Sprintf("Creating project: %s", name))
	printInfo(fmt.Sprintf("Template: %s", template))

	// Create project directory
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create main.abc file with template
	var content string
	switch template {
	case "hello-world":
		content = getHelloWorldTemplate()
	case "rest-api":
		content = getRestAPITemplate()
	default:
		return fmt.Errorf("unknown template: %s", template)
	}

	mainFile := filepath.Join(name, "main.abc")
	if err := os.WriteFile(mainFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write main.abc: %w", err)
	}

	printSuccess(fmt.Sprintf("Project created successfully in %s/", name))
	printInfo(fmt.Sprintf("Run: cd %s && glyph dev main.abc", name))
	return nil
}

// startServerWithInterpreter parses the source file and starts the HTTP server using the tree-walking interpreter
func startServerWithInterpreter(filePath string, port int) (*http.Server, error) {
	// Read source file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the source (placeholder - Rust parser integration will come later)
	module, err := parseSource(string(source))
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// Create interpreter and load module
	interp := interpreter.NewInterpreter()
	if err := interp.LoadModule(*module); err != nil {
		return nil, fmt.Errorf("failed to load module: %w", err)
	}

	// Create router and register routes
	router := server.NewRouter()
	for _, item := range module.Items {
		if route, ok := item.(*interpreter.Route); ok {
			err := registerRoute(router, route, interp)
			if err != nil {
				printWarning(fmt.Sprintf("Failed to register route %s: %v", route.Path, err))
			}
		}
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", createHandler(router))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: loggingMiddleware(mux),
	}

	// Start server in background
	go func() {
		printSuccess(fmt.Sprintf("Server listening on http://localhost:%d", port))
		printInfo("Press Ctrl+C to stop")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			printError(fmt.Errorf("server error: %w", err))
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	return srv, nil
}

// startServerWithCompiler parses, compiles, and starts the HTTP server using the VM
func startServerWithCompiler(filePath string, port int) (*http.Server, error) {
	// Read source file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the source
	module, err := parseSource(string(source))
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// Compile routes
	c := compiler.NewCompilerWithOptLevel(compiler.OptBasic)
	compiledRoutes := make(map[string][]byte) // path -> bytecode

	for _, item := range module.Items {
		if route, ok := item.(*interpreter.Route); ok {
			bytecode, err := c.CompileRoute(route)
			if err != nil {
				return nil, fmt.Errorf("failed to compile route %s: %w", route.Path, err)
			}
			compiledRoutes[route.Path] = bytecode
			printInfo(fmt.Sprintf("Compiled route: %s %s (%d bytes)", route.Method, route.Path, len(bytecode)))
		}
	}

	// Create WebSocket server first (so HTTP routes can access stats)
	wsServer := websocket.NewServer()

	// Create router and register compiled routes
	router := server.NewRouter()
	for _, item := range module.Items {
		if route, ok := item.(*interpreter.Route); ok {
			bytecode := compiledRoutes[route.Path]
			err := registerCompiledRoute(router, route, bytecode, wsServer.GetHub())
			if err != nil {
				printWarning(fmt.Sprintf("Failed to register route %s: %v", route.Path, err))
			}
		}
	}

	// Compile and register WebSocket routes
	for _, item := range module.Items {
		if wsRoute, ok := item.(*interpreter.WebSocketRoute); ok {
			compiledWs, err := c.CompileWebSocketRoute(wsRoute)
			if err != nil {
				printWarning(fmt.Sprintf("Failed to compile WebSocket route %s: %v", wsRoute.Path, err))
				continue
			}
			registerCompiledWebSocketRoute(wsServer, wsRoute.Path, compiledWs)
			printInfo(fmt.Sprintf("Compiled WebSocket route: %s", wsRoute.Path))
		}
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", createHandler(router))

	// Register WebSocket routes with HTTP mux
	for _, item := range module.Items {
		if wsRoute, ok := item.(*interpreter.WebSocketRoute); ok {
			path := wsRoute.Path
			mux.HandleFunc(path, wsServer.HandleWebSocket)
			printInfo(fmt.Sprintf("WebSocket endpoint: ws://localhost:%d%s", port, path))
		}
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: loggingMiddleware(mux),
	}

	// Start server in background
	go func() {
		printSuccess(fmt.Sprintf("Server listening on http://localhost:%d (compiled mode)", port))
		printInfo("Press Ctrl+C to stop")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			printError(fmt.Errorf("server error: %w", err))
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	return srv, nil
}

// registerCompiledWebSocketRoute registers a compiled WebSocket route with event handlers
func registerCompiledWebSocketRoute(wsServer *websocket.Server, path string, compiled *compiler.CompiledWebSocketRoute) {
	hub := wsServer.GetHub()

	// Register connect handler
	if len(compiled.OnConnect) > 0 {
		hub.OnConnect(func(conn *websocket.Connection) error {
			return executeWebSocketBytecode(compiled.OnConnect, conn, hub, nil)
		})
	}

	// Register disconnect handler
	if len(compiled.OnDisconnect) > 0 {
		hub.OnDisconnect(func(conn *websocket.Connection) error {
			return executeWebSocketBytecode(compiled.OnDisconnect, conn, hub, nil)
		})
	}

	// Register message handler
	if len(compiled.OnMessage) > 0 {
		hub.OnMessage(websocket.MessageTypeJSON, func(ctx *websocket.MessageContext) error {
			return executeWebSocketBytecode(compiled.OnMessage, ctx.Conn, hub, ctx.Message)
		})
		hub.OnMessage(websocket.MessageTypeText, func(ctx *websocket.MessageContext) error {
			return executeWebSocketBytecode(compiled.OnMessage, ctx.Conn, hub, ctx.Message)
		})
	}
}

// executeWebSocketBytecode executes compiled WebSocket event bytecode
func executeWebSocketBytecode(bytecode []byte, conn *websocket.Connection, hub *websocket.Hub, msg *websocket.Message) error {
	// Create VM instance
	vmInstance := vm.NewVM()

	// Create WebSocket handler adapter
	wsHandler := websocket.NewVMHandler(conn, hub)
	vmInstance.SetWebSocketHandler(wsHandler)

	// Set connection context variables
	vmInstance.SetLocal("client", vm.StringValue{Val: conn.ID})

	// Set input data if message is provided
	if msg != nil {
		vmInstance.SetLocal("input", convertMessageToValue(msg))
	}

	// Execute bytecode
	_, err := vmInstance.Execute(bytecode)
	return err
}

// convertMessageToValue converts a WebSocket message to a VM Value
func convertMessageToValue(msg *websocket.Message) vm.Value {
	if msg == nil {
		return vm.NullValue{}
	}

	// Convert message data to VM value
	if msg.Data != nil {
		return interfaceToValue(msg.Data)
	}

	// Return message type as string if no data
	return vm.StringValue{Val: string(msg.Type)}
}

// interfaceToValue converts a Go interface{} to a VM Value
func interfaceToValue(v interface{}) vm.Value {
	if v == nil {
		return vm.NullValue{}
	}

	switch val := v.(type) {
	case int:
		return vm.IntValue{Val: int64(val)}
	case int64:
		return vm.IntValue{Val: val}
	case float64:
		return vm.FloatValue{Val: val}
	case string:
		return vm.StringValue{Val: val}
	case bool:
		return vm.BoolValue{Val: val}
	case []interface{}:
		arr := make([]vm.Value, len(val))
		for i, elem := range val {
			arr[i] = interfaceToValue(elem)
		}
		return vm.ArrayValue{Val: arr}
	case map[string]interface{}:
		obj := make(map[string]vm.Value)
		for k, elem := range val {
			obj[k] = interfaceToValue(elem)
		}
		return vm.ObjectValue{Val: obj}
	default:
		return vm.NullValue{}
	}
}

// parseSource parses GLYPH source using the Rust FFI parser
func parseSource(source string) (*interpreter.Module, error) {
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

// registerRoute registers a route with the router
func registerRoute(router *server.Router, route *interpreter.Route, interp *interpreter.Interpreter) error {
	handler := createRouteHandler(route, interp)

	serverRoute := &server.Route{
		Method:  convertHTTPMethod(route.Method),
		Path:    route.Path,
		Handler: handler,
	}

	return router.RegisterRoute(serverRoute)
}

// registerCompiledRoute registers a compiled route with the router
func registerCompiledRoute(router *server.Router, route *interpreter.Route, bytecode []byte, wsHub *websocket.Hub) error {
	handler := createCompiledRouteHandler(route, bytecode, wsHub)

	serverRoute := &server.Route{
		Method:  convertHTTPMethod(route.Method),
		Path:    route.Path,
		Handler: handler,
	}

	return router.RegisterRoute(serverRoute)
}

// createCompiledRouteHandler creates an HTTP handler that executes compiled bytecode
func createCompiledRouteHandler(route *interpreter.Route, bytecode []byte, wsHub *websocket.Hub) server.RouteHandler {
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

		// Execute compiled bytecode
		result, err := vmInstance.Execute(bytecode)
		if err != nil {
			// Return error response
			ctx.StatusCode = http.StatusInternalServerError
			ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
			return json.NewEncoder(ctx.ResponseWriter).Encode(map[string]interface{}{
				"error": fmt.Sprintf("bytecode execution failed: %v", err),
			})
		}

		// Set response
		ctx.StatusCode = http.StatusOK
		ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(ctx.ResponseWriter).Encode(result)
	}
}

// createRouteHandler creates an HTTP handler for a route
func createRouteHandler(route *interpreter.Route, interp *interpreter.Interpreter) server.RouteHandler {
	return func(ctx *server.Context) error {
		// Execute route body using the interpreter
		result, err := executeRoute(route, ctx, interp)
		if err != nil {
			// Return error response
			ctx.StatusCode = http.StatusInternalServerError
			ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
			return json.NewEncoder(ctx.ResponseWriter).Encode(map[string]interface{}{
				"error": err.Error(),
			})
		}

		// Set response
		ctx.StatusCode = http.StatusOK
		ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(ctx.ResponseWriter).Encode(result)
	}
}

// executeRoute executes a route's body and returns the result
func executeRoute(route *interpreter.Route, ctx *server.Context, interp *interpreter.Interpreter) (interface{}, error) {
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
			var bodyMap map[string]interface{}
			decoder := json.NewDecoder(ctx.Request.Body)
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

// createHandler creates the main HTTP handler
func createHandler(router *server.Router) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
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

// watchFile watches a file for changes and calls onChange when it changes
func watchFile(filePath string, onChange func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		printError(fmt.Errorf("failed to create watcher: %w", err))
		return
	}
	defer watcher.Close()

	// Add file to watcher
	if err := watcher.Add(filePath); err != nil {
		printError(fmt.Errorf("failed to watch file: %w", err))
		return
	}

	// Watch for events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				onChange()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			printError(fmt.Errorf("watcher error: %w", err))
		}
	}
}

// waitForShutdown waits for interrupt signal and gracefully shuts down the server
func waitForShutdown(srv *http.Server) error {
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

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

// convertHTTPMethod converts interpreter.HttpMethod to server.HTTPMethod
func convertHTTPMethod(method interpreter.HttpMethod) server.HTTPMethod {
	switch method {
	case interpreter.Get:
		return server.GET
	case interpreter.Post:
		return server.POST
	case interpreter.Put:
		return server.PUT
	case interpreter.Delete:
		return server.DELETE
	case interpreter.Patch:
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

@ route /hello
  > {text: "Hello, World!", timestamp: 1234567890}

@ route /greet/:name -> Message
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

@ route /api/users -> List[User]
  + auth(jwt)
  > []

@ route /health
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
	fmt.Printf("(%s)\n", d)
}

// runLSP handles the LSP command
func runLSP(cmd *cobra.Command, args []string) error {
	logFile, _ := cmd.Flags().GetString("log")

	printInfo("Starting GLYPH Language Server...")
	if logFile != "" {
		printInfo(fmt.Sprintf("Logging to: %s", logFile))
	}

	// Create LSP server using stdin/stdout
	server := lsp.NewServer(os.Stdin, os.Stdout, logFile)

	// Start server (blocks until shutdown)
	if err := server.Start(); err != nil {
		return fmt.Errorf("LSP server error: %w", err)
	}

	return nil
}
