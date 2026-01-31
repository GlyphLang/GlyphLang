package main

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/glyphlang/glyph/pkg/codegen"
	"github.com/glyphlang/glyph/pkg/docs"
	"github.com/glyphlang/glyph/pkg/formatter"
	"github.com/glyphlang/glyph/pkg/openapi"
	"github.com/glyphlang/glyph/pkg/parser"
	"github.com/spf13/cobra"
)

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

// runOpenAPI handles the openapi command
func runOpenAPI(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	output, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")
	title, _ := cmd.Flags().GetString("title")
	apiVersion, _ := cmd.Flags().GetString("api-version")

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

	// Default title from filename
	if title == "" {
		base := filepath.Base(filePath)
		ext := filepath.Ext(base)
		title = base[:len(base)-len(ext)] + " API"
	}

	// Generate spec
	spec := openapi.GenerateFromModule(module, title, apiVersion)

	// Format output
	data, err := openapi.FormatSpec(spec, format)
	if err != nil {
		return err
	}

	// Write output
	if output != "" {
		if err := os.WriteFile(output, data, 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		printSuccess(fmt.Sprintf("OpenAPI spec written to %s", output))
		return nil
	}

	fmt.Print(string(data))
	return nil
}

// runDocs handles the docs command
func runDocs(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	output, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")
	title, _ := cmd.Flags().GetString("title")

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

	// Default title from filename
	if title == "" {
		base := filepath.Base(filePath)
		ext := filepath.Ext(base)
		title = base[:len(base)-len(ext)] + " API"
	}

	apiDoc := docs.ExtractDocs(module, title)

	var content string
	switch format {
	case "html":
		content = docs.GenerateHTML(apiDoc)
	case "markdown", "md":
		content = docs.GenerateMarkdown(apiDoc)
	default:
		return fmt.Errorf("unsupported format: %s (supported: html, markdown)", format)
	}

	// Write output
	if output != "" {
		if err := os.WriteFile(output, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		printSuccess(fmt.Sprintf("API documentation written to %s", output))
		return nil
	}

	fmt.Print(content)
	return nil
}

// runClient handles the client command
func runClient(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	output, _ := cmd.Flags().GetString("output")
	lang, _ := cmd.Flags().GetString("lang")
	baseURL, _ := cmd.Flags().GetString("base-url")

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

	var code string
	switch lang {
	case "typescript", "ts":
		gen := codegen.NewTypeScriptGenerator(baseURL)
		code = gen.Generate(module)
	default:
		return fmt.Errorf("unsupported language: %s (supported: typescript)", lang)
	}

	// Write output
	if output != "" {
		if err := os.WriteFile(output, []byte(code), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		printSuccess(fmt.Sprintf("API client written to %s", output))
		return nil
	}

	fmt.Print(code)
	return nil
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
