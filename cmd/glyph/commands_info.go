package main

import (
	"encoding/json"
	"fmt"
	"github.com/glyphlang/glyph/pkg/ast"
	"os"
	"path/filepath"
	"strings"

	glyphcontext "github.com/glyphlang/glyph/pkg/context"
	"github.com/glyphlang/glyph/pkg/validate"
	"github.com/spf13/cobra"
)

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
			fmt.Printf("  @ command %s %s\n", name, strings.Join(params, " "))
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
				optsStr = " [" + strings.Join(opts, ", ") + "]"
			}
			fmt.Printf("  @ queue \"%s\"%s\n", queueName, optsStr)
		}
	}

	return nil
}

// typeToString converts an ast.Type to a string representation
func typeToString(t ast.Type) string {
	switch v := t.(type) {
	case ast.IntType:
		return "int"
	case ast.StringType:
		return "str"
	case ast.BoolType:
		return "bool"
	case ast.FloatType:
		return "float"
	case ast.NamedType:
		return v.Name
	case ast.ArrayType:
		return typeToString(v.ElementType) + "[]"
	default:
		return "any"
	}
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
			return fmt.Errorf("validation failed: not all files are valid")
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
