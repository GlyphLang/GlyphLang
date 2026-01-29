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
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/config"
	glyphcontext "github.com/glyphlang/glyph/pkg/context"
	"github.com/glyphlang/glyph/pkg/database"
	"github.com/glyphlang/glyph/pkg/decompiler"
	"github.com/glyphlang/glyph/pkg/formatter"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/lsp"
	"github.com/glyphlang/glyph/pkg/parser"
	"github.com/glyphlang/glyph/pkg/repl"
	"github.com/glyphlang/glyph/pkg/server"
	"github.com/glyphlang/glyph/pkg/validate"
	"github.com/glyphlang/glyph/pkg/vm"
	"github.com/glyphlang/glyph/pkg/websocket"
	"github.com/spf13/cobra"
)

var version = "0.3.5"

func main() {
	// Check if invoked with just a .glyph file (e.g., double-click on Windows)
	// In this case, open the file in the default text editor
	if len(os.Args) == 2 && filepath.Ext(os.Args[1]) == ".glyph" {
		filePath := os.Args[1]
		// Check if the file exists
		if _, err := os.Stat(filePath); err == nil {
			if err := openInEditor(filePath); err != nil {
				printError(fmt.Errorf("failed to open editor: %w", err))
				os.Exit(1)
			}
			return
		}
	}

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
	decompileCmd.Flags().BoolP("disasm", "d", false, "Output disassembly only (no pseudo-source generation)")

	// Run command
	var runCmd = &cobra.Command{
		Use:   "run <file>",
		Short: "Run GLYPH source file",
		Args:  cobra.ExactArgs(1),
		RunE:  runRun,
	}
	runCmd.Flags().Uint16P("port", "p", uint16(config.DefaultPort), "Port to listen on")
	runCmd.Flags().Bool("bytecode", false, "Execute bytecode (.glyphc) file")
	runCmd.Flags().Bool("interpret", false, "Use tree-walking interpreter instead of compiler (fallback mode)")

	// Dev command
	var devCmd = &cobra.Command{
		Use:   "dev <file>",
		Short: "Start development server with hot reload",
		Args:  cobra.ExactArgs(1),
		RunE:  runDev,
	}
	devCmd.Flags().Uint16P("port", "p", uint16(config.DefaultPort), "Port to listen on")
	devCmd.Flags().BoolP("watch", "w", true, "Watch for file changes")
	devCmd.Flags().BoolP("open", "o", false, "Open browser automatically")

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

	// Exec command - execute CLI commands defined with @ command
	var execCmd = &cobra.Command{
		Use:   "exec <file> <command> [args...]",
		Short: "Execute a CLI command defined in a GLYPH file",
		Long: `Execute CLI commands defined with @ command in a GLYPH file.

Example:
  # In my-cli.glyph:
  @ command hello name: str!
    $ greeting = "Hello, " + name + "!"
    > greeting

  # Run it:
  glyph exec my-cli.glyph hello --name "World"`,
		Args: cobra.MinimumNArgs(2),
		RunE: runExec,
	}

	// List commands - list all commands in a GLYPH file
	var listCmdsCmd = &cobra.Command{
		Use:   "commands <file>",
		Short: "List all CLI commands defined in a GLYPH file",
		Args:  cobra.ExactArgs(1),
		RunE:  runListCommands,
	}

	// Context command - generate AI-optimized context
	var contextCmd = &cobra.Command{
		Use:   "context [path]",
		Short: "Generate AI-optimized context for the project",
		Long: `Generate compact, cacheable context for AI agents working with this Glyph project.

The context includes:
- Type definitions with field signatures
- Route definitions with methods, paths, and return types
- Function signatures
- CLI command definitions
- Detected patterns (CRUD, auth usage, etc.)

Output formats:
- json: Full structured context (default)
- compact: Minimal text representation for token efficiency
- stubs: Type stub file format (.glyph.d style)

Examples:
  glyph context                    # Generate context for current directory
  glyph context ./src              # Generate context for specific directory
  glyph context --format compact   # Generate compact text output
  glyph context --output .glyph/context.json  # Write to file
  glyph context --file main.glyph  # Generate context for single file`,
		Args: cobra.MaximumNArgs(1),
		RunE: runContext,
	}
	contextCmd.Flags().StringP("format", "f", "json", "Output format: json, compact, stubs")
	contextCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	contextCmd.Flags().String("file", "", "Generate context for a single file only")
	contextCmd.Flags().Bool("pretty", true, "Pretty-print JSON output")
	contextCmd.Flags().Bool("changed", false, "Show only changes since last context generation")
	contextCmd.Flags().Bool("save", false, "Save context to .glyph/context.json for future diffing")
	contextCmd.Flags().String("for", "", "Generate targeted context for a task: route, type, function, command")

	// Validate command - validate source files with structured errors
	var validateCmd = &cobra.Command{
		Use:   "validate <file>",
		Short: "Validate a GLYPH source file",
		Long: `Validate a GLYPH source file and report errors.

By default, outputs human-readable error messages. Use --ai for structured
JSON output optimized for AI agents to parse and fix issues.

The --ai flag outputs:
- Structured error types (syntax_error, undefined_reference, type_mismatch, etc.)
- Precise locations (file, line, column)
- Fix hints for each error
- Source context around the error

Examples:
  glyph validate main.glyph           # Human-readable output
  glyph validate main.glyph --ai      # Structured JSON for AI
  glyph validate src/ --ai            # Validate all files in directory`,
		Args: cobra.ExactArgs(1),
		RunE: runValidate,
	}
	validateCmd.Flags().Bool("ai", false, "Output structured JSON for AI agents")
	validateCmd.Flags().Bool("strict", false, "Treat warnings as errors")
	validateCmd.Flags().Bool("quiet", false, "Only output errors, no stats")

	// Expand command - convert compact glyph to human-readable syntax
	var expandCmd = &cobra.Command{
		Use:   "expand <file|dir>",
		Short: "Convert compact glyph syntax to human-readable expanded syntax",
		Long: `Expand converts compact glyph syntax (using symbols like @, $, >) to
human-readable expanded syntax (using keywords like route, let, return).

This is useful for:
- Making AI-generated code easier to read and modify
- Learning the glyph language by seeing keyword equivalents
- Code review and documentation

The expanded syntax can be converted back using 'glyph compact'.

Examples:
  glyph expand main.glyph                    # Expand single file
  glyph expand main.glyph -o main.glyphx     # Write to specific file
  glyph expand ./src                         # Expand all .glyph files in directory
  glyph expand main.glyph --watch            # Watch file and auto-expand on changes
  glyph expand ./src --watch                 # Watch directory for changes`,
		Args: cobra.ExactArgs(1),
		RunE: runExpand,
	}
	expandCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	expandCmd.Flags().BoolP("watch", "w", false, "Watch for file changes and auto-convert")

	// Compact command - convert human-readable syntax to compact glyph
	var compactCmd = &cobra.Command{
		Use:   "compact <file|dir>",
		Short: "Convert human-readable syntax to compact glyph syntax",
		Long: `Compact converts human-readable expanded syntax (using keywords like route,
let, return) back to compact glyph syntax (using symbols like @, $, >).

This is useful for:
- Converting edited human-readable code back to canonical glyph format
- Minimizing file size for AI token efficiency
- Standardizing code format

Examples:
  glyph compact main.glyphx                   # Compact single file
  glyph compact main.glyphx -o main.glyph     # Write to specific file
  glyph compact ./src                         # Compact all .glyphx files in directory
  glyph compact main.glyphx --watch           # Watch file and auto-compact on changes
  glyph compact ./src --watch                 # Watch directory for changes`,
		Args: cobra.ExactArgs(1),
		RunE: runCompact,
	}
	compactCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	compactCmd.Flags().BoolP("watch", "w", false, "Watch for file changes and auto-convert")

	// REPL command - interactive Read-Eval-Print Loop
	var replCmd = &cobra.Command{
		Use:   "repl",
		Short: "Start an interactive REPL session",
		Long: `Start an interactive Read-Eval-Print Loop (REPL) for GlyphLang.

The REPL allows you to:
- Execute Glyph expressions and statements interactively
- Explore language features and test code snippets
- Inspect variables and their values

Commands:
  :help    - Show available commands
  :quit    - Exit the REPL
  :clear   - Clear the environment
  :vars    - Show all defined variables

Examples:
  glyph repl              # Start REPL`,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := repl.New(os.Stdin, os.Stdout, version)
			return r.Start()
		},
	}

	// Version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GlyphLang v%s\n", version)
		},
	}

	// Version command (built-in, but we can customize)
	rootCmd.SetVersionTemplate(`GlyphLang v{{.Version}}
`)

	// Add commands to root
	rootCmd.AddCommand(compileCmd)
	rootCmd.AddCommand(decompileCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(lspCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(listCmdsCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(expandCmd)
	rootCmd.AddCommand(compactCmd)
	rootCmd.AddCommand(replCmd)
	rootCmd.AddCommand(versionCmd)

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
		output = changeExtension(filePath, ".glyphc")
	}

	// Write bytecode to file with restricted permissions (owner read/write only)
	if err := os.WriteFile(output, bytecode, 0600); err != nil {
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
	disasmOnly, _ := cmd.Flags().GetBool("disasm")

	printInfo(fmt.Sprintf("Decompiling %s...", filePath))

	// Read bytecode file
	bytecode, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Use the decompiler package
	dec := decompiler.NewDecompiler()
	result, err := dec.Decompile(bytecode)
	if err != nil {
		return fmt.Errorf("decompilation failed: %w", err)
	}

	// Print metadata
	printInfo(fmt.Sprintf("Bytecode version: %d", result.Version))
	printInfo(fmt.Sprintf("Constants: %d", len(result.Constants)))
	printInfo(fmt.Sprintf("Instructions: %d", len(result.Instructions)))

	if disasmOnly {
		// Only output disassembly to console
		fmt.Println()
		fmt.Print(result.FormatDisassembly())
		return nil
	}

	// Determine output path
	if output == "" {
		output = changeExtension(filePath, ".glyph")
	}

	// Write decompiled source to file with restricted permissions
	source := result.Format()
	if err := os.WriteFile(output, []byte(source), 0600); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	printSuccess(fmt.Sprintf("Decompiled to %s", output))

	// Also print disassembly to console
	fmt.Println()
	fmt.Print(result.FormatDisassembly())

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
		useBytecode = filepath.Ext(filePath) == ".glyphc"
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

		return nil
	}

	// Running source file - use shared server startup logic
	printInfo(fmt.Sprintf("Starting server for %s...", filePath))
	srv, err := startServer(filePath, int(port), useInterpreter)
	if err != nil {
		return err
	}

	// Setup graceful shutdown
	return waitForShutdown(srv)
}

// runDev handles the dev command
func runDev(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	port, _ := cmd.Flags().GetUint16("port")
	watch, _ := cmd.Flags().GetBool("watch")
	openBrowser, _ := cmd.Flags().GetBool("open")

	printInfo(fmt.Sprintf("Starting development server on port %d...", port))

	// Get absolute path for watching
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create hot reload manager
	manager := &hotReloadManager{
		filePath:        absPath,
		port:            int(port),
		liveReloadConns: make(map[*liveReloadConn]bool),
	}

	// Start initial server
	if err := manager.startServer(); err != nil {
		return err
	}

	// Setup file watching if enabled
	if watch {
		printInfo(fmt.Sprintf("Watching %s for changes...", absPath))
		go manager.watchForChanges()
	}

	// Open browser if requested
	if openBrowser {
		url := fmt.Sprintf("http://localhost:%d", port)
		if err := openURL(url); err != nil {
			printWarning(fmt.Sprintf("Failed to open browser: %v", err))
		} else {
			printInfo(fmt.Sprintf("Opened %s in browser", url))
		}
	}

	// Setup graceful shutdown
	return manager.waitForShutdown()
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
		if wsRoute, ok := item.(*interpreter.WebSocketRoute); ok {
			path := wsRoute.Path
			// Convert :param to {param} for Go's http.ServeMux pattern matching
			muxPattern := server.ConvertPatternToMuxFormat(path)
			mux.HandleFunc(muxPattern, wsServer.HandleWebSocketWithPattern(path))
			printInfo(fmt.Sprintf("WebSocket endpoint: ws://localhost:%d%s", m.port, path))
		}
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", m.port),
		Handler: loggingMiddleware(mux),
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
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

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

	// Create main.glyph file with template
	var content string
	switch template {
	case "hello-world":
		content = getHelloWorldTemplate()
	case "rest-api":
		content = getRestAPITemplate()
	default:
		return fmt.Errorf("unknown template: %s", template)
	}

	mainFile := filepath.Join(name, "main.glyph")
	if err := os.WriteFile(mainFile, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write main.glyph: %w", err)
	}

	printSuccess(fmt.Sprintf("Project created successfully in %s/", name))
	printInfo(fmt.Sprintf("Run: cd %s && glyph dev main.glyph", name))
	return nil
}

// setupRoutes handles the common logic of determining execution mode, compiling routes,
// and setting up the router. Used by both startServer and startDevServerInternal.
// filePath is the path to the source file, used for resolving relative module imports.
func setupRoutes(module *interpreter.Module, filePath string, forceInterpreter ...bool) (useCompiler bool, compiledRoutes map[string][]byte, wsServer *websocket.Server, router *server.Router, err error) {
	useCompiler = true
	if len(forceInterpreter) > 0 && forceInterpreter[0] {
		useCompiler = false
	}
	compiledRoutes = make(map[string][]byte)

	// Check if any route has database injection - VM doesn't support db method calls
	for _, item := range module.Items {
		if route, ok := item.(*interpreter.Route); ok {
			for _, injection := range route.Injections {
				if _, isDB := injection.Type.(interpreter.DatabaseType); isDB {
					printInfo("Routes use database injection, using interpreter mode")
					useCompiler = false
					break
				}
				if named, ok := injection.Type.(interpreter.NamedType); ok && named.Name == "Database" {
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
			if route, ok := item.(*interpreter.Route); ok {
				bytecode, compileErr := c.CompileRoute(route)
				if compileErr != nil {
					// Semantic errors (like redeclaration) should fail completely, not fall back
					if compiler.IsSemanticError(compileErr) {
						printError(fmt.Errorf("compilation error for %s: %v", route.Path, compileErr))
						os.Exit(1)
					}
					printWarning(fmt.Sprintf("Compilation failed for %s: %v, falling back to interpreter", route.Path, compileErr))
					useCompiler = false
					break
				}
				compiledRoutes[route.Path] = bytecode
			}
		}
	}

	// Create WebSocket server
	wsServer = websocket.NewServer()

	// Create router and register routes
	router = server.NewRouter()
	interp := newConfiguredInterpreter()

	if useCompiler {
		for _, item := range module.Items {
			if route, ok := item.(*interpreter.Route); ok {
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
			if wsRoute, ok := item.(*interpreter.WebSocketRoute); ok {
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
			if route, ok := item.(*interpreter.Route); ok {
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
		if wsRoute, ok := item.(*interpreter.WebSocketRoute); ok {
			path := wsRoute.Path
			// Convert :param to {param} for Go's http.ServeMux pattern matching
			muxPattern := server.ConvertPatternToMuxFormat(path)
			mux.HandleFunc(muxPattern, wsServer.HandleWebSocketWithPattern(path))
			printInfo(fmt.Sprintf("WebSocket endpoint: ws://localhost:%d%s", port, path))
		}
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: loggingMiddleware(mux),
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

// parseSource parses GLYPH source using the Go parser
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

// newConfiguredInterpreter creates an interpreter with common configuration
// including mock database for development/demo purposes.
func newConfiguredInterpreter() *interpreter.Interpreter {
	interp := interpreter.NewInterpreter()
	mockDB := database.NewMockDatabase()
	interp.SetDatabaseHandler(mockDB)

	// Set up the parse function for module resolution
	interp.GetModuleResolver().SetParseFunc(func(source string) (*interpreter.Module, error) {
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

// runExec handles the exec command - executes CLI commands defined with @ command
func runExec(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	cmdName := args[1]
	cmdArgs := args[2:]

	// Read source file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse source
	module, err := parseSource(string(source))
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Create interpreter and load module
	interp := newConfiguredInterpreter()
	if err := interp.LoadModule(*module); err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Find the command
	glyphCmd, ok := interp.GetCommand(cmdName)
	if !ok {
		// List available commands
		commands := interp.GetCommands()
		if len(commands) == 0 {
			return fmt.Errorf("no commands found in %s", filePath)
		}
		var available []string
		for name := range commands {
			available = append(available, name)
		}
		return fmt.Errorf("command '%s' not found. Available commands: %v", cmdName, available)
	}

	// Parse command arguments
	argsMap := parseCommandArgs(cmdArgs, glyphCmd.Params)

	// Execute command
	start := time.Now()
	result, err := interp.ExecuteCommand(&glyphCmd, argsMap)
	execTime := time.Since(start)

	if err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	// Print result
	if result != nil {
		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Println(result)
		} else {
			fmt.Println(string(resultJSON))
		}
	}

	printInfo(fmt.Sprintf("Execution time: %s", execTime))
	return nil
}

// parseCommandArgs parses CLI arguments into a map based on command parameters
func parseCommandArgs(args []string, params []interpreter.CommandParam) map[string]interface{} {
	result := make(map[string]interface{})

	positionalIdx := 0
	i := 0
	for i < len(args) {
		arg := args[i]

		// Check for flag arguments (--name value or --name=value)
		if len(arg) > 2 && arg[:2] == "--" {
			flagPart := arg[2:]
			var name, value string

			if eqIdx := indexOf(flagPart, '='); eqIdx != -1 {
				name = flagPart[:eqIdx]
				value = flagPart[eqIdx+1:]
			} else {
				name = flagPart
				if i+1 < len(args) {
					i++
					value = args[i]
				}
			}

			result[name] = value
		} else {
			// Positional argument
			for _, param := range params {
				if !param.IsFlag && positionalIdx == 0 {
					result[param.Name] = arg
					positionalIdx++
					break
				}
			}
			positionalIdx++
		}
		i++
	}

	return result
}

// indexOf returns the index of char c in string s, or -1 if not found
func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// runListCommands lists all commands in a GLYPH file
func runListCommands(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Read source file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse source
	module, err := parseSource(string(source))
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Create interpreter and load module
	interp := newConfiguredInterpreter()
	if err := interp.LoadModule(*module); err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Get commands
	commands := interp.GetCommands()
	cronTasks := interp.GetCronTasks()
	eventHandlers := interp.GetAllEventHandlers()
	queueWorkers := interp.GetQueueWorkers()

	if len(commands) == 0 && len(cronTasks) == 0 && len(eventHandlers) == 0 && len(queueWorkers) == 0 {
		printInfo("No commands, cron tasks, event handlers, or queue workers found in " + filePath)
		return nil
	}

	// Print commands
	if len(commands) > 0 {
		printSuccess("CLI Commands:")
		for name, cmd := range commands {
			var params []string
			for _, p := range cmd.Params {
				paramStr := p.Name
				if p.Type != nil {
					paramStr += ": " + typeToString(p.Type)
				}
				if p.Required {
					paramStr += "!"
				}
				if p.IsFlag {
					paramStr = "--" + paramStr
				}
				params = append(params, paramStr)
			}
			fmt.Printf("  @ command %s %s\n", name, joinStrings(params, " "))
			if cmd.Description != "" {
				fmt.Printf("    %s\n", cmd.Description)
			}
		}
	}

	// Print cron tasks
	if len(cronTasks) > 0 {
		printSuccess("\nCron Tasks:")
		for _, task := range cronTasks {
			name := task.Name
			if name == "" {
				name = "(anonymous)"
			}
			fmt.Printf("  @ cron \"%s\" %s\n", task.Schedule, name)
		}
	}

	// Print event handlers
	if len(eventHandlers) > 0 {
		printSuccess("\nEvent Handlers:")
		for eventType, handlers := range eventHandlers {
			fmt.Printf("  @ event \"%s\" (%d handler(s))\n", eventType, len(handlers))
		}
	}

	// Print queue workers
	if len(queueWorkers) > 0 {
		printSuccess("\nQueue Workers:")
		for queueName, worker := range queueWorkers {
			opts := []string{}
			if worker.Concurrency > 0 {
				opts = append(opts, fmt.Sprintf("concurrency=%d", worker.Concurrency))
			}
			if worker.MaxRetries > 0 {
				opts = append(opts, fmt.Sprintf("retries=%d", worker.MaxRetries))
			}
			optsStr := ""
			if len(opts) > 0 {
				optsStr = " [" + joinStrings(opts, ", ") + "]"
			}
			fmt.Printf("  @ queue \"%s\"%s\n", queueName, optsStr)
		}
	}

	return nil
}

// typeToString converts an interpreter.Type to a string representation
func typeToString(t interpreter.Type) string {
	switch v := t.(type) {
	case interpreter.IntType:
		return "int"
	case interpreter.StringType:
		return "str"
	case interpreter.BoolType:
		return "bool"
	case interpreter.FloatType:
		return "float"
	case interpreter.NamedType:
		return v.Name
	case interpreter.ArrayType:
		return typeToString(v.ElementType) + "[]"
	default:
		return "any"
	}
}

// joinStrings joins strings with a separator (simple helper)
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// runContext handles the context command - generates AI-optimized context
func runContext(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	output, _ := cmd.Flags().GetString("output")
	singleFile, _ := cmd.Flags().GetString("file")
	pretty, _ := cmd.Flags().GetBool("pretty")
	showChanged, _ := cmd.Flags().GetBool("changed")
	saveContext, _ := cmd.Flags().GetBool("save")
	forTask, _ := cmd.Flags().GetString("for")

	// Determine root directory
	rootDir := "."
	if len(args) > 0 {
		rootDir = args[0]
	}

	// Get absolute path
	absPath, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Create generator
	gen := glyphcontext.NewGenerator(absPath)

	// Generate context
	var ctx *glyphcontext.ProjectContext
	if singleFile != "" {
		printInfo(fmt.Sprintf("Generating context for %s...", singleFile))
		ctx, err = gen.GenerateForFile(singleFile)
	} else {
		printInfo(fmt.Sprintf("Generating context for %s...", absPath))
		ctx, err = gen.Generate()
	}

	if err != nil {
		return fmt.Errorf("failed to generate context: %w", err)
	}

	// Handle --for flag (targeted context)
	if forTask != "" {
		var targetedCtx *glyphcontext.TargetedContext

		switch forTask {
		case "route":
			targetedCtx = ctx.ForRoute()
		case "type":
			targetedCtx = ctx.ForType()
		case "function":
			targetedCtx = ctx.ForFunction()
		case "command":
			targetedCtx = ctx.ForCommand()
		default:
			return fmt.Errorf("unknown task: %s (use: route, type, function, command)", forTask)
		}

		// Format targeted output
		var outputData []byte
		switch format {
		case "json":
			outputData, err = targetedCtx.ToJSON(pretty)
			if err != nil {
				return fmt.Errorf("failed to serialize context: %w", err)
			}
		case "compact", "stubs":
			outputData = []byte(targetedCtx.ToCompact())
		default:
			return fmt.Errorf("unknown format: %s (use: json, compact)", format)
		}

		// Output
		if output != "" {
			if err := os.WriteFile(output, outputData, 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			printSuccess(fmt.Sprintf("Context for %s written to %s", forTask, output))
		} else {
			fmt.Println(string(outputData))
		}

		return nil
	}

	// Handle --changed flag
	if showChanged {
		contextPath := glyphcontext.DefaultContextPath(absPath)
		var previousCtx *glyphcontext.ProjectContext

		// Try to load previous context
		previousCtx, loadErr := glyphcontext.LoadContext(contextPath)
		if loadErr != nil {
			printWarning("No previous context found, showing all as new")
		}

		// Generate diff
		diff := ctx.Diff(previousCtx)

		// Format diff output
		var outputData []byte
		switch format {
		case "json":
			outputData, err = diff.ToJSON(pretty)
			if err != nil {
				return fmt.Errorf("failed to serialize diff: %w", err)
			}
		case "compact", "stubs":
			outputData = []byte(diff.ToCompact(ctx))
		default:
			return fmt.Errorf("unknown format: %s (use: json, compact)", format)
		}

		// Output diff
		if output != "" {
			if err := os.WriteFile(output, outputData, 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			printSuccess(fmt.Sprintf("Diff written to %s", output))
		} else {
			fmt.Println(string(outputData))
		}

		// Optionally save the new context
		if saveContext {
			if err := ctx.SaveContext(contextPath); err != nil {
				return fmt.Errorf("failed to save context: %w", err)
			}
			printSuccess(fmt.Sprintf("Context saved to %s", contextPath))
		}

		return nil
	}

	// Format output (normal mode)
	var outputData []byte
	switch format {
	case "json":
		outputData, err = ctx.ToJSON(pretty)
		if err != nil {
			return fmt.Errorf("failed to serialize context: %w", err)
		}
	case "compact":
		outputData = []byte(ctx.ToCompact())
	case "stubs":
		outputData = []byte(ctx.GenerateStubs())
	default:
		return fmt.Errorf("unknown format: %s (use: json, compact, stubs)", format)
	}

	// Write output
	if output != "" {
		// Ensure parent directory exists
		dir := filepath.Dir(output)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}

		if err := os.WriteFile(output, outputData, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		printSuccess(fmt.Sprintf("Context written to %s", output))

		// Print stats
		printInfo(fmt.Sprintf("Files: %d", len(ctx.Files)))
		printInfo(fmt.Sprintf("Types: %d", len(ctx.Types)))
		printInfo(fmt.Sprintf("Routes: %d", len(ctx.Routes)))
		printInfo(fmt.Sprintf("Functions: %d", len(ctx.Functions)))
		printInfo(fmt.Sprintf("Commands: %d", len(ctx.Commands)))
		if len(ctx.Patterns) > 0 {
			printInfo(fmt.Sprintf("Patterns: %v", ctx.Patterns))
		}
	} else {
		// Write to stdout
		fmt.Println(string(outputData))
	}

	// Save context if requested
	if saveContext {
		contextPath := glyphcontext.DefaultContextPath(absPath)
		if err := ctx.SaveContext(contextPath); err != nil {
			return fmt.Errorf("failed to save context: %w", err)
		}
		printSuccess(fmt.Sprintf("Context saved to %s", contextPath))
	}

	return nil
}

// runValidate handles the validate command - validates source files with structured errors
func runValidate(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	aiMode, _ := cmd.Flags().GetBool("ai")
	strict, _ := cmd.Flags().GetBool("strict")
	quiet, _ := cmd.Flags().GetBool("quiet")

	// Check if path is a directory
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to access path: %w", err)
	}

	var results []*validate.ValidationResult

	if info.IsDir() {
		// Validate all .glyph files in directory
		err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".glyph" {
				result := validateFile(path)
				results = append(results, result)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		// Validate single file
		result := validateFile(filePath)
		results = append(results, result)
	}

	// Handle strict mode - treat warnings as errors
	if strict {
		for _, r := range results {
			if len(r.Warnings) > 0 {
				r.Valid = false
				for _, w := range r.Warnings {
					w.Severity = "error"
					r.Errors = append(r.Errors, w)
				}
				r.Warnings = nil
			}
		}
	}

	// Output results
	if aiMode {
		// JSON output for AI
		if len(results) == 1 {
			output, err := results[0].ToJSON(true)
			if err != nil {
				return fmt.Errorf("failed to serialize result: %w", err)
			}
			fmt.Println(string(output))
		} else {
			// Multiple files - wrap in array
			output, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize results: %w", err)
			}
			fmt.Println(string(output))
		}
	} else {
		// Human-readable output
		allValid := true
		for _, r := range results {
			if !quiet || !r.Valid {
				fmt.Print(r.ToHuman())
			}
			if !r.Valid {
				allValid = false
			}
		}

		// Summary for multiple files
		if len(results) > 1 && !quiet {
			validCount := 0
			for _, r := range results {
				if r.Valid {
					validCount++
				}
			}
			fmt.Printf("\nValidation complete: %d/%d files valid\n", validCount, len(results))
		}

		if !allValid {
			os.Exit(1)
		}
	}

	return nil
}

// validateFile validates a single file and returns the result
func validateFile(filePath string) *validate.ValidationResult {
	source, err := os.ReadFile(filePath)
	if err != nil {
		return &validate.ValidationResult{
			Valid:    false,
			FilePath: filePath,
			Errors: []*validate.ValidationError{
				{
					Type:     "file_error",
					Message:  fmt.Sprintf("failed to read file: %v", err),
					Severity: "error",
				},
			},
		}
	}

	validator := validate.NewValidator(string(source), filePath)
	return validator.Validate()
}

// runExpand handles the expand command - converts compact .glyph to human-readable .glyphx
func runExpand(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	output, _ := cmd.Flags().GetString("output")
	watch, _ := cmd.Flags().GetBool("watch")

	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if path is a directory
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("failed to access path: %w", err)
	}

	if watch {
		return runWatchMode(absPath, output, info.IsDir(), "expand")
	}

	// Single file mode
	if info.IsDir() {
		return expandDirectory(absPath, output)
	}

	return expandFile(absPath, output)
}

// expandFile expands a single .glyph file to .glyphx
func expandFile(filePath, output string) error {
	// Validate input file extension
	if filepath.Ext(filePath) != ".glyph" {
		printWarning(fmt.Sprintf("Input file %s is not a .glyph file", filePath))
	}

	printInfo(fmt.Sprintf("Expanding %s to human-readable .glyphx syntax...", filePath))

	// Read source file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Token-level transformation (preserves comments and formatting)
	result := formatter.ExpandSource(string(source))

	// Determine output path (default: change .glyph to .glyphx)
	if output == "" {
		output = changeExtension(filePath, ".glyphx")
	}

	if err := os.WriteFile(output, []byte(result), 0600); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	printSuccess(fmt.Sprintf("Expanded output written to %s", output))
	printInfo(fmt.Sprintf("Original: %d bytes -> Expanded: %d bytes", len(source), len(result)))

	return nil
}

// expandDirectory expands all .glyph files in a directory
func expandDirectory(dirPath, outputDir string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".glyph" {
			return nil
		}

		// Determine output path
		var outPath string
		if outputDir != "" {
			relPath, _ := filepath.Rel(dirPath, path)
			outPath = filepath.Join(outputDir, changeExtension(relPath, ".glyphx"))
			// Ensure output directory exists
			if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		} else {
			outPath = changeExtension(path, ".glyphx")
		}

		return expandFile(path, outPath)
	})
}

// runCompact handles the compact command - converts human-readable .glyphx to compact .glyph
func runCompact(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	output, _ := cmd.Flags().GetString("output")
	watch, _ := cmd.Flags().GetBool("watch")

	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if path is a directory
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("failed to access path: %w", err)
	}

	if watch {
		return runWatchMode(absPath, output, info.IsDir(), "compact")
	}

	// Single file mode
	if info.IsDir() {
		return compactDirectory(absPath, output)
	}

	return compactFile(absPath, output)
}

// compactFile compacts a single .glyphx file to .glyph
func compactFile(filePath, output string) error {
	// Validate input file extension
	if filepath.Ext(filePath) != ".glyphx" {
		printWarning(fmt.Sprintf("Input file %s is not a .glyphx file", filePath))
	}

	printInfo(fmt.Sprintf("Compacting %s to .glyph syntax...", filePath))

	// Read source file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Token-level transformation (preserves comments and formatting)
	result := formatter.CompactSource(string(source))

	// Determine output path (default: change .glyphx to .glyph)
	if output == "" {
		output = changeExtension(filePath, ".glyph")
	}

	if err := os.WriteFile(output, []byte(result), 0600); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	printSuccess(fmt.Sprintf("Compact output written to %s", output))
	printInfo(fmt.Sprintf("Original: %d bytes -> Compact: %d bytes", len(source), len(result)))

	return nil
}

// compactDirectory compacts all .glyphx files in a directory
func compactDirectory(dirPath, outputDir string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".glyphx" {
			return nil
		}

		// Determine output path
		var outPath string
		if outputDir != "" {
			relPath, _ := filepath.Rel(dirPath, path)
			outPath = filepath.Join(outputDir, changeExtension(relPath, ".glyph"))
			// Ensure output directory exists
			if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		} else {
			outPath = changeExtension(path, ".glyph")
		}

		return compactFile(path, outPath)
	})
}

// runWatchMode runs the expand or compact command in watch mode
func runWatchMode(path, outputDir string, isDir bool, mode string) error {
	// Determine source and target extensions based on mode
	var sourceExt, targetExt string
	if mode == "expand" {
		sourceExt = ".glyph"
		targetExt = ".glyphx"
	} else {
		sourceExt = ".glyphx"
		targetExt = ".glyph"
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Function to process a file
	processFile := func(filePath string) {
		var outPath string
		if outputDir != "" {
			if isDir {
				relPath, _ := filepath.Rel(path, filePath)
				outPath = filepath.Join(outputDir, changeExtension(relPath, targetExt))
				// Ensure output directory exists
				if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
					printError(fmt.Errorf("failed to create output directory: %w", err))
					return
				}
			} else {
				outPath = outputDir
			}
		} else {
			outPath = changeExtension(filePath, targetExt)
		}

		var err error
		if mode == "expand" {
			err = expandFile(filePath, outPath)
		} else {
			err = compactFile(filePath, outPath)
		}
		if err != nil {
			printError(err)
		}
	}

	// Do initial conversion
	if isDir {
		if mode == "expand" {
			if err := expandDirectory(path, outputDir); err != nil {
				printError(err)
			}
		} else {
			if err := compactDirectory(path, outputDir); err != nil {
				printError(err)
			}
		}
	} else {
		processFile(path)
	}

	// Add paths to watcher
	if isDir {
		// Watch directory and subdirectories
		err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return watcher.Add(p)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to watch directory: %w", err)
		}
		printInfo(fmt.Sprintf("Watching %s for %s file changes...", path, sourceExt))
	} else {
		// Watch the file's directory (more reliable for atomic saves)
		dir := filepath.Dir(path)
		if err := watcher.Add(dir); err != nil {
			return fmt.Errorf("failed to watch directory: %w", err)
		}
		printInfo(fmt.Sprintf("Watching %s for changes...", path))
	}

	printInfo("Press Ctrl+C to stop")

	// Debounce timer
	var debounceTimer *time.Timer
	debounceDelay := 100 * time.Millisecond
	pendingFiles := make(map[string]bool)
	var pendingMu sync.Mutex

	// Watch loop
	for {
		select {
		case <-sigChan:
			printWarning("\nStopping watch mode...")
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			printSuccess("Watch mode stopped gracefully")
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Check if the file matches our source extension
			if filepath.Ext(event.Name) != sourceExt {
				continue
			}

			// For single file mode, only react to our file
			if !isDir && filepath.Base(event.Name) != filepath.Base(path) {
				continue
			}

			// React to write or create events (create happens with atomic saves)
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				pendingMu.Lock()
				pendingFiles[event.Name] = true
				pendingMu.Unlock()

				// Reset debounce timer
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(debounceDelay, func() {
					pendingMu.Lock()
					filesToProcess := make([]string, 0, len(pendingFiles))
					for f := range pendingFiles {
						filesToProcess = append(filesToProcess, f)
					}
					pendingFiles = make(map[string]bool)
					pendingMu.Unlock()

					for _, f := range filesToProcess {
						processFile(f)
					}
				})
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			printError(fmt.Errorf("watcher error: %w", err))
		}
	}
}

// parseExpandedSource parses .glyphx source using the expanded lexer
func parseExpandedSource(source string) (*interpreter.Module, error) {
	lexer := parser.NewExpandedLexer(source)
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
