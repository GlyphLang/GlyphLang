package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/decompiler"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/lsp"
	"github.com/glyphlang/glyph/pkg/parser"
	"github.com/glyphlang/glyph/pkg/vm"
	"github.com/spf13/cobra"
)

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
		if route, ok := item.(*ast.Route); ok {
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

// runTest handles the test command - executes test blocks in a GLYPH file.
// Argument count is validated by cobra.ExactArgs(1) before this function is called.
// printWarning, printInfo are defined in this file (see helper functions section).
func runTest(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("invalid flag --verbose: %w", err)
	}
	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return fmt.Errorf("invalid flag --filter: %w", err)
	}
	failFast, err := cmd.Flags().GetBool("fail-fast")
	if err != nil {
		return fmt.Errorf("invalid flag --fail-fast: %w", err)
	}

	// Read and parse source file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Determine lexer type based on file extension
	var tokens []parser.Token
	if filepath.Ext(filePath) == ".glyphx" {
		lexer := parser.NewExpandedLexer(string(source))
		tokens, err = lexer.Tokenize()
	} else {
		lexer := parser.NewLexer(string(source))
		tokens, err = lexer.Tokenize()
	}
	if err != nil {
		return fmt.Errorf("lexer error: %w", err)
	}

	p := parser.NewParserWithSource(tokens, string(source))
	module, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Create interpreter and load module
	interp := interpreter.NewInterpreter()
	if err := interp.LoadModule(*module); err != nil {
		return fmt.Errorf("load error: %w", err)
	}

	tests := interp.GetTestBlocks()
	if len(tests) == 0 {
		printWarning("No test blocks found in " + filePath)
		return nil
	}

	// Run tests
	results := interp.RunTests(filter)
	if len(results) == 0 {
		printWarning("No tests matched filter: " + filter)
		return nil
	}

	// Display results
	passed := 0
	failed := 0

	greenCheck := color.New(color.FgGreen).SprintFunc()
	redX := color.New(color.FgRed).SprintFunc()

	for _, r := range results {
		if r.Passed {
			passed++
			if verbose {
				fmt.Printf("  %s %s (%s)\n", greenCheck("PASS"), r.Name, r.Duration)
			}
		} else {
			failed++
			fmt.Printf("  %s %s\n", redX("FAIL"), r.Name)
			if r.Error != "" {
				fmt.Printf("       %s\n", r.Error)
			}
			if failFast {
				break
			}
		}
	}

	// Summary
	fmt.Println()
	total := passed + failed
	if failed > 0 {
		color.New(color.FgRed, color.Bold).Printf("FAIL: %d/%d tests passed\n", passed, total)
		return fmt.Errorf("%d test(s) failed", failed)
	}

	color.New(color.FgGreen, color.Bold).Printf("PASS: %d/%d tests passed\n", passed, total)
	return nil
}

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
func parseCommandArgs(args []string, params []ast.CommandParam) map[string]interface{} {
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
