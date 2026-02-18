package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/glyphlang/glyph/pkg/config"
	"github.com/glyphlang/glyph/pkg/repl"
	"github.com/spf13/cobra"
)

var version = "0.4.0"

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

	// OpenAPI command - generate OpenAPI 3.0 specification
	var openapiCmd = &cobra.Command{
		Use:   "openapi <file>",
		Short: "Generate OpenAPI 3.0 specification from GLYPH source",
		Long: `Generate an OpenAPI 3.0 specification from your GLYPH source code.

Analyzes route definitions, type definitions, authentication middleware,
and query parameters to produce a complete OpenAPI 3.0 specification.

Output formats:
  - yaml: YAML format (default)
  - json: JSON format

Examples:
  glyph openapi main.glyph                      # Output YAML to stdout
  glyph openapi main.glyph -o openapi.yaml      # Write to file
  glyph openapi main.glyph --format json         # Output as JSON
  glyph openapi main.glyph --title "My API"      # Set API title`,
		Args: cobra.ExactArgs(1),
		RunE: runOpenAPI,
	}
	openapiCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	openapiCmd.Flags().StringP("format", "f", "yaml", "Output format: yaml or json")
	openapiCmd.Flags().String("title", "", "API title (default: derived from filename)")
	openapiCmd.Flags().String("api-version", "1.0.0", "API version")

	// Docs command - generate API documentation
	var docsCmd = &cobra.Command{
		Use:   "docs <file>",
		Short: "Generate API documentation from GLYPH source",
		Long: `Generate API documentation from your GLYPH source code.

Reads route definitions and type definitions to produce documentation
with endpoint listings, request/response schemas, and type definitions.

Output formats:
  - html: Self-contained HTML page with sidebar and search (default)
  - markdown: Markdown documentation

Examples:
  glyph docs main.glyph                      # Output HTML to stdout
  glyph docs main.glyph -o docs.html         # Write to file
  glyph docs main.glyph --format markdown    # Output as Markdown
  glyph docs main.glyph --title "My API"     # Set API title`,
		Args: cobra.ExactArgs(1),
		RunE: runDocs,
	}
	docsCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	docsCmd.Flags().StringP("format", "f", "html", "Output format: html or markdown")
	docsCmd.Flags().String("title", "", "API title (default: derived from filename)")

	// Client command - generate API client code
	var clientCmd = &cobra.Command{
		Use:   "client <file>",
		Short: "Generate API client code from GLYPH source",
		Long: `Generate a typed API client from your GLYPH source code.

Reads route definitions and type definitions to produce a fully typed
API client with methods for each endpoint.

Output languages:
  - typescript: TypeScript client (default)

Examples:
  glyph client main.glyph                          # Output TypeScript to stdout
  glyph client main.glyph -o client.ts              # Write to file
  glyph client main.glyph --base-url http://api.io  # Set base URL`,
		Args: cobra.ExactArgs(1),
		RunE: runClient,
	}
	clientCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	clientCmd.Flags().String("lang", "typescript", "Output language: typescript")
	clientCmd.Flags().String("base-url", "http://localhost:3000", "Base URL for the API client")

	// Codegen command - generate server code for target languages
	var codegenCmd = &cobra.Command{
		Use:   "codegen <file>",
		Short: "Generate server code for a target language from GLYPH source",
		Long: `Generate a complete server application from your GLYPH source code.

Parses the .glyph file, transforms it through the Semantic IR, and generates
a working server application in the target language.

Output languages:
  - python: Python/FastAPI server (default)
  - typescript: TypeScript/Express server

Examples:
  glyph codegen main.glyph                               # Output Python to stdout
  glyph codegen main.glyph --output ./out                 # Write project to directory
  glyph codegen main.glyph --lang python -o ./out         # Python/FastAPI
  glyph codegen main.glyph --lang typescript -o ./out     # TypeScript/Express`,
		Args: cobra.ExactArgs(1),
		RunE: runCodegen,
	}
	codegenCmd.Flags().StringP("output", "o", "", "Output directory for generated project")
	codegenCmd.Flags().String("lang", "python", "Target language: python, typescript")

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

	// Test command
	var testCmd = &cobra.Command{
		Use:   "test <file>",
		Short: "Run tests defined in a GLYPH file",
		Long: `Execute all test blocks defined with 'test' keyword in a GLYPH file.

Example:
  test "should add numbers" {
    assert(1 + 1 == 2)
  }

  glyph test math_test.glyph
  glyph test math_test.glyph --verbose
  glyph test math_test.glyph --filter "add*"`,
		Args: cobra.ExactArgs(1),
		RunE: runTest,
	}
	testCmd.Flags().BoolP("verbose", "v", false, "Show detailed output for each test")
	testCmd.Flags().StringP("filter", "f", "", "Run only tests matching filter pattern")
	testCmd.Flags().Bool("fail-fast", false, "Stop on first test failure")

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
	rootCmd.AddCommand(openapiCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(clientCmd)
	rootCmd.AddCommand(codegenCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		printError(err)
		os.Exit(1)
	}
}
