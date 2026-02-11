package main

import (
	"context"
	"fmt"
	"github.com/glyphlang/glyph/pkg/ast"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/server"
	"github.com/glyphlang/glyph/pkg/vm"
	"github.com/glyphlang/glyph/pkg/websocket"
)

// hotReloadManager manages server lifecycle for hot reload
type hotReloadManager struct {
	filePath        string
	port            int
	server          *http.Server
	mu              sync.Mutex
	watcher         *fsnotify.Watcher
	liveReloadConns map[*liveReloadConn]bool
	liveReloadMu    sync.Mutex
}

// liveReloadConn represents a live reload SSE connection
type liveReloadConn struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	done    chan struct{}
}

// startServer starts or restarts the server
func (m *hotReloadManager) startServer() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop existing server if running
	if m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		m.server.Shutdown(ctx)
		time.Sleep(100 * time.Millisecond) // Allow port to be released
	}

	// Start dev server with live reload support
	srv, err := m.startDevServerInternal()
	if err != nil {
		return err
	}

	m.server = srv
	return nil
}

// startDevServerInternal starts the development server with live reload support
func (m *hotReloadManager) startDevServerInternal() (*http.Server, error) {
	// Read source file
	source, err := os.ReadFile(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the source
	module, err := parseSource(string(source))
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// Use shared logic for route compilation/interpretation
	useCompiler, _, wsServer, router, err := setupRoutes(module, m.filePath)
	if err != nil {
		return nil, err
	}

	// Create HTTP server with live reload support
	mux := http.NewServeMux()

	// Live reload SSE endpoint
	mux.HandleFunc("/__livereload", m.handleLiveReload)

	// Live reload script endpoint
	mux.HandleFunc("/__livereload.js", m.handleLiveReloadScript)

	// Main application handler
	mux.HandleFunc("/", createHandler(router))

	// Register WebSocket routes
	for _, item := range module.Items {
		if wsRoute, ok := item.(*ast.WebSocketRoute); ok {
			path := wsRoute.Path
			// Convert :param to {param} for Go's http.ServeMux pattern matching
			muxPattern := server.ConvertPatternToMuxFormat(path)
			mux.HandleFunc(muxPattern, wsServer.HandleWebSocketWithPattern(path))
			printInfo(fmt.Sprintf("WebSocket endpoint: ws://localhost:%d%s", m.port, path))
		}
	}

	// Register static file routes
	if err := registerStaticRoutes(mux, module, m.filePath, m.port); err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", m.port),
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
		printSuccess(fmt.Sprintf("Dev server listening on http://localhost:%d (%s mode)", m.port, mode))
		printInfo("Live reload enabled at /__livereload")
		printInfo("Press Ctrl+C to stop")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			printError(fmt.Errorf("server error: %w", err))
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	return srv, nil
}

// handleLiveReload handles Server-Sent Events for live reload
func (m *hotReloadManager) handleLiveReload(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create connection
	conn := &liveReloadConn{
		writer:  w,
		flusher: flusher,
		done:    make(chan struct{}),
	}

	// Register connection
	m.liveReloadMu.Lock()
	m.liveReloadConns[conn] = true
	m.liveReloadMu.Unlock()

	// Send initial connected event
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	// Wait for disconnect
	<-r.Context().Done()

	// Unregister connection
	m.liveReloadMu.Lock()
	delete(m.liveReloadConns, conn)
	close(conn.done)
	m.liveReloadMu.Unlock()
}

// handleLiveReloadScript serves the live reload JavaScript
func (m *hotReloadManager) handleLiveReloadScript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	script := fmt.Sprintf(`(function() {
    var es = new EventSource('http://localhost:%d/__livereload');
    es.onmessage = function(e) {
        var data = JSON.parse(e.data);
        if (data.action === 'reload') {
            console.log('[LiveReload] Reloading...');
            window.location.reload();
        }
    };
    es.addEventListener('connected', function(e) {
        console.log('[LiveReload] Connected');
    });
    es.onerror = function() {
        console.log('[LiveReload] Connection lost. Retrying...');
    };
})();`, m.port)
	w.Write([]byte(script))
}

// notifyLiveReload sends a reload notification to all connected clients
func (m *hotReloadManager) notifyLiveReload() {
	m.liveReloadMu.Lock()
	defer m.liveReloadMu.Unlock()

	for conn := range m.liveReloadConns {
		select {
		case <-conn.done:
			continue
		default:
			fmt.Fprintf(conn.writer, "data: {\"action\":\"reload\"}\n\n")
			conn.flusher.Flush()
		}
	}
}

// watchForChanges watches the file and triggers reload on changes
func (m *hotReloadManager) watchForChanges() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		printError(fmt.Errorf("failed to create watcher: %w", err))
		return
	}
	m.watcher = watcher

	// Watch the file's directory (more reliable for editors that do atomic saves)
	dir := filepath.Dir(m.filePath)
	filename := filepath.Base(m.filePath)

	if err := watcher.Add(dir); err != nil {
		printError(fmt.Errorf("failed to watch directory: %w", err))
		return
	}

	// Debounce timer to avoid multiple reloads
	var debounceTimer *time.Timer
	debounceDelay := 100 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Only react to our file
			if filepath.Base(event.Name) != filename {
				continue
			}

			// React to write or create events (create happens with atomic saves)
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				// Debounce: reset timer on each event
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(debounceDelay, func() {
					m.reload()
				})
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			printError(fmt.Errorf("watcher error: %w", err))
		}
	}
}

// reload reloads the server with updated code
func (m *hotReloadManager) reload() {
	printWarning("\nFile changed, reloading...")
	start := time.Now()

	if err := m.startServer(); err != nil {
		printError(fmt.Errorf("reload failed: %w", err))
		printWarning("Server still running with previous version")
	} else {
		printSuccess(fmt.Sprintf("Hot reload complete (%s)", time.Since(start)))
		// Notify all connected browsers to reload
		m.notifyLiveReload()
	}
}

// waitForShutdown waits for shutdown signal
func (m *hotReloadManager) waitForShutdown() error {
	sigChan := make(chan os.Signal, 1)
	signalNotify(sigChan)

	<-sigChan

	printWarning("\nShutting down server...")

	// Close watcher
	if m.watcher != nil {
		m.watcher.Close()
	}

	// Shutdown server
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := m.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
	}

	printSuccess("Server stopped gracefully")
	return nil
}

// registerCompiledWebSocketRoute registers a compiled WebSocket route with event handlers
func registerCompiledWebSocketRoute(wsServer *websocket.Server, path string, compiled *compiler.CompiledWebSocketRoute) {
	hub := wsServer.GetHub()

	// Register connect handler for this specific route
	if len(compiled.OnConnect) > 0 {
		hub.OnConnectForRoute(path, func(conn *websocket.Connection) error {
			return executeWebSocketBytecode(compiled.OnConnect, conn, hub, nil)
		})
	}

	// Register disconnect handler for this specific route
	if len(compiled.OnDisconnect) > 0 {
		hub.OnDisconnectForRoute(path, func(conn *websocket.Connection) error {
			return executeWebSocketBytecode(compiled.OnDisconnect, conn, hub, nil)
		})
	}

	// Register message handler with route filtering
	// Message handlers are global, so we filter by route pattern at execution time
	if len(compiled.OnMessage) > 0 {
		routePath := path // Capture for closure
		hub.OnMessage(websocket.MessageTypeJSON, func(ctx *websocket.MessageContext) error {
			if ctx.Conn.RoutePattern() != routePath {
				return nil // Skip - not for this route
			}
			return executeWebSocketBytecode(compiled.OnMessage, ctx.Conn, hub, ctx.Message)
		})
		hub.OnMessage(websocket.MessageTypeText, func(ctx *websocket.MessageContext) error {
			if ctx.Conn.RoutePattern() != routePath {
				return nil // Skip - not for this route
			}
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

	// Inject path parameters from connection (e.g., room from /chat/:room)
	for key, value := range conn.PathParams {
		vmInstance.SetLocal(key, vm.StringValue{Val: value})
	}

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
