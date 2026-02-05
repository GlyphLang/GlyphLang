package main

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/ast"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/server"
	"github.com/glyphlang/glyph/pkg/websocket"
)

// setupRoutes handles the common logic of determining execution mode, compiling routes,
// and setting up the router. Used by both startServer and startDevServerInternal.
// filePath is the path to the source file, used for resolving relative module imports.
func setupRoutes(module *ast.Module, filePath string, forceInterpreter ...bool) (useCompiler bool, compiledRoutes map[string][]byte, wsServer *websocket.Server, router *server.Router, err error) {
	useCompiler = true
	if len(forceInterpreter) > 0 && forceInterpreter[0] {
		useCompiler = false
	}
	compiledRoutes = make(map[string][]byte)

	// Check if any route has database injection - VM doesn't support db method calls
	for _, item := range module.Items {
		if route, ok := item.(*ast.Route); ok {
			for _, injection := range route.Injections {
				if _, isDB := injection.Type.(ast.DatabaseType); isDB {
					printInfo("Routes use database injection, using interpreter mode")
					useCompiler = false
					break
				}
				if named, ok := injection.Type.(ast.NamedType); ok && named.Name == "Database" {
					printInfo("Routes use database injection, using interpreter mode")
					useCompiler = false
					break
				}
			}
			if !useCompiler {
				break
			}
		}
	}

	// Try to compile routes if using compiler mode
	if useCompiler {
		c := compiler.NewCompilerWithOptLevel(compiler.OptBasic)
		for _, item := range module.Items {
			if route, ok := item.(*ast.Route); ok {
				bytecode, compileErr := c.CompileRoute(route)
				if compileErr != nil {
					// Semantic errors (like redeclaration) should fail completely, not fall back
					if compiler.IsSemanticError(compileErr) {
						err = fmt.Errorf("compilation error for %s: %v", route.Path, compileErr)
						return
					}
					printWarning(fmt.Sprintf("Compilation failed for %s: %v, falling back to interpreter", route.Path, compileErr))
					useCompiler = false
					break
				}
				compiledRoutes[route.Path] = bytecode
			}
		}
	}

	// Create WebSocket server with CORS-aware origin checking
	var wsConfig *websocket.Config
	if corsOrigin := os.Getenv("GLYPH_CORS_ORIGIN"); corsOrigin != "" {
		wsConfig = websocket.DefaultConfig()
		wsConfig.AllowedOrigins = []string{corsOrigin}
	}
	wsServer = websocket.NewServer(wsConfig)

	// Create router and register routes
	router = server.NewRouter()
	interp := newConfiguredInterpreter()

	if useCompiler {
		for _, item := range module.Items {
			if route, ok := item.(*ast.Route); ok {
				bytecode := compiledRoutes[route.Path]
				regErr := registerCompiledRoute(router, route, bytecode, wsServer.GetHub())
				if regErr != nil {
					printWarning(fmt.Sprintf("Failed to register route %s: %v", route.Path, regErr))
				} else {
					printInfo(fmt.Sprintf("Compiled route: %s %s", route.Method, route.Path))
				}
			}
		}

		// Compile and register WebSocket routes
		c := compiler.NewCompilerWithOptLevel(compiler.OptBasic)
		for _, item := range module.Items {
			if wsRoute, ok := item.(*ast.WebSocketRoute); ok {
				compiledWs, compileErr := c.CompileWebSocketRoute(wsRoute)
				if compileErr != nil {
					printWarning(fmt.Sprintf("Failed to compile WebSocket route %s: %v", wsRoute.Path, compileErr))
					continue
				}
				registerCompiledWebSocketRoute(wsServer, wsRoute.Path, compiledWs)
				printInfo(fmt.Sprintf("Compiled WebSocket route: %s", wsRoute.Path))
			}
		}
	} else {
		// Use interpreter mode
		// Pass the directory of the source file for proper module resolution
		basePath := filepath.Dir(filePath)
		if loadErr := interp.LoadModuleWithPath(*module, basePath); loadErr != nil {
			err = fmt.Errorf("failed to load module: %w", loadErr)
			return
		}
		for _, item := range module.Items {
			if route, ok := item.(*ast.Route); ok {
				regErr := registerRoute(router, route, interp)
				if regErr != nil {
					printWarning(fmt.Sprintf("Failed to register route %s: %v", route.Path, regErr))
				}
			}
		}
	}

	return useCompiler, compiledRoutes, wsServer, router, nil
}

// startServer is the unified server startup function used by both 'run' and 'dev' commands.
// It handles database injection detection and automatic fallback to interpreter mode.
func startServer(filePath string, port int, forceInterpreter bool) (*http.Server, error) {
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

	// Use shared logic for route compilation/interpretation
	useCompiler, _, wsServer, router, err := setupRoutes(module, filePath, forceInterpreter)
	if err != nil {
		return nil, err
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", createHandler(router))

	// Register WebSocket routes with HTTP mux
	for _, item := range module.Items {
		if wsRoute, ok := item.(*ast.WebSocketRoute); ok {
			path := wsRoute.Path
			// Convert :param to {param} for Go's http.ServeMux pattern matching
			muxPattern := server.ConvertPatternToMuxFormat(path)
			mux.HandleFunc(muxPattern, wsServer.HandleWebSocketWithPattern(path))
			printInfo(fmt.Sprintf("WebSocket endpoint: ws://localhost:%d%s", port, path))
		}
	}

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        loggingMiddleware(mux),
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in background
	go func() {
		mode := "compiled"
		if !useCompiler {
			mode = "interpreted"
		}
		printSuccess(fmt.Sprintf("Server listening on http://localhost:%d (%s mode)", port, mode))
		printInfo("Press Ctrl+C to stop")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			printError(fmt.Errorf("server error: %w", err))
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	return srv, nil
}
