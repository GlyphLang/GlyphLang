// Package repl provides an interactive Read-Eval-Print Loop for Glyph.
package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/parser"
)

// REPL provides an interactive programming environment for Glyph.
type REPL struct {
	interp  *interpreter.Interpreter
	env     *interpreter.Environment
	reader  *bufio.Reader
	writer  io.Writer
	running bool
	version string
	// inputBuffer holds incomplete multi-line input
	inputBuffer strings.Builder
	// lineNumber tracks the current input line for prompts
	lineNumber int
}

// New creates a new REPL instance.
func New(reader io.Reader, writer io.Writer, version string) *REPL {
	interp := interpreter.NewInterpreter()
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

	return &REPL{
		interp:     interp,
		env:        interpreter.NewEnvironment(),
		reader:     bufio.NewReader(reader),
		writer:     writer,
		running:    false,
		version:    version,
		lineNumber: 1,
	}
}

// Start begins the REPL loop.
func (r *REPL) Start() error {
	r.running = true
	r.printWelcome()

	for r.running {
		r.printPrompt()
		line, err := r.readLine()
		if err != nil {
			if err == io.EOF {
				r.running = false
				break
			}
			r.printf("Error reading input: %v\n", err)
			continue
		}

		line = strings.TrimRight(line, "\r\n")

		// Handle empty lines
		if line == "" && r.inputBuffer.Len() == 0 {
			continue
		}

		if err := r.processLine(line); err != nil {
			r.printf("Error: %v\n", err)
		}
	}

	r.printGoodbye()
	return nil
}

// Stop stops the REPL loop.
func (r *REPL) Stop() {
	r.running = false
}

// processLine processes a single line of input.
func (r *REPL) processLine(line string) error {
	// Check for REPL commands (lines starting with :)
	if strings.HasPrefix(line, ":") && r.inputBuffer.Len() == 0 {
		return r.executeCommand(line)
	}

	// Add line to input buffer
	if r.inputBuffer.Len() > 0 {
		r.inputBuffer.WriteString("\n")
	}
	r.inputBuffer.WriteString(line)

	// Check if input is complete (balanced braces/parens)
	input := r.inputBuffer.String()
	if !r.isInputComplete(input) {
		// Continue reading - show continuation prompt
		return nil
	}

	// Reset input buffer
	r.inputBuffer.Reset()
	r.lineNumber++

	// Trim input
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	// Try to evaluate the input
	return r.evaluate(input)
}

// evaluate evaluates the given input and prints the result.
func (r *REPL) evaluate(input string) error {
	// Detect what type of input this is
	inputType := r.detectInputType(input)

	switch inputType {
	case inputTypeExpression:
		return r.evaluateExpression(input)
	case inputTypeStatement:
		return r.evaluateStatement(input)
	case inputTypeTypeDef:
		return r.evaluateTypeDef(input)
	case inputTypeFunction:
		return r.evaluateFunction(input)
	default:
		return r.evaluateExpression(input)
	}
}

// inputType represents the type of REPL input.
type inputType int

const (
	inputTypeExpression inputType = iota
	inputTypeStatement
	inputTypeTypeDef
	inputTypeFunction
)

// detectInputType determines what type of input the user has provided.
func (r *REPL) detectInputType(input string) inputType {
	trimmed := strings.TrimSpace(input)

	// Type definition: starts with ":" but not a REPL command
	if strings.HasPrefix(trimmed, ":") {
		// Check for REPL commands first
		lowerTrimmed := strings.ToLower(trimmed)
		if strings.HasPrefix(lowerTrimmed, ":help") || strings.HasPrefix(lowerTrimmed, ":h ") ||
			lowerTrimmed == ":h" ||
			strings.HasPrefix(lowerTrimmed, ":quit") || strings.HasPrefix(lowerTrimmed, ":q ") ||
			lowerTrimmed == ":q" || strings.HasPrefix(lowerTrimmed, ":exit") ||
			strings.HasPrefix(lowerTrimmed, ":type") || strings.HasPrefix(lowerTrimmed, ":t ") ||
			strings.HasPrefix(lowerTrimmed, ":load") || strings.HasPrefix(lowerTrimmed, ":l ") ||
			strings.HasPrefix(lowerTrimmed, ":reset") || strings.HasPrefix(lowerTrimmed, ":r ") ||
			lowerTrimmed == ":r" ||
			strings.HasPrefix(lowerTrimmed, ":clear") || strings.HasPrefix(lowerTrimmed, ":cls") ||
			strings.HasPrefix(lowerTrimmed, ":vars") || strings.HasPrefix(lowerTrimmed, ":v ") ||
			lowerTrimmed == ":v" ||
			strings.HasPrefix(lowerTrimmed, ":types") ||
			strings.HasPrefix(lowerTrimmed, ":functions") || strings.HasPrefix(lowerTrimmed, ":fns") {
			// This is a REPL command, not a type definition
		} else if len(trimmed) > 1 {
			// Type definition: : TypeName { ... }
			// After ":", skip spaces and check for a valid identifier
			afterColon := strings.TrimSpace(trimmed[1:])
			if len(afterColon) > 0 && (afterColon[0] >= 'A' && afterColon[0] <= 'Z') {
				return inputTypeTypeDef
			}
		}
	}

	// Variable declaration: starts with "$"
	if strings.HasPrefix(trimmed, "$") {
		return inputTypeStatement
	}

	// Return statement: starts with ">"
	if strings.HasPrefix(trimmed, ">") {
		return inputTypeStatement
	}

	// Function definition: starts with "!"
	if strings.HasPrefix(trimmed, "!") {
		return inputTypeFunction
	}

	// Let statement
	if strings.HasPrefix(trimmed, "let ") {
		return inputTypeStatement
	}

	// Return statement (keyword)
	if strings.HasPrefix(trimmed, "return ") {
		return inputTypeStatement
	}

	// Control flow keywords
	if strings.HasPrefix(trimmed, "if ") || strings.HasPrefix(trimmed, "while ") ||
		strings.HasPrefix(trimmed, "for ") || strings.HasPrefix(trimmed, "switch ") {
		return inputTypeStatement
	}

	// Type keyword
	if strings.HasPrefix(trimmed, "type ") {
		return inputTypeTypeDef
	}

	// Default to expression
	return inputTypeExpression
}

// evaluateExpression evaluates an expression and prints the result.
func (r *REPL) evaluateExpression(input string) error {
	// Parse the expression
	expr, err := r.parseExpression(input)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Evaluate the expression
	result, err := r.interp.EvaluateExpression(expr, r.env)
	if err != nil {
		return fmt.Errorf("evaluation error: %w", err)
	}

	// Print the result
	r.printResult(result)
	return nil
}

// evaluateStatement evaluates a statement.
func (r *REPL) evaluateStatement(input string) error {
	// Parse the statement
	stmt, err := r.parseStatement(input)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Execute the statement
	result, err := r.interp.ExecuteStatement(stmt, r.env)
	if err != nil {
		// Check for return value (which is normal for > statements)
		if strings.Contains(err.Error(), "return") {
			r.printResult(result)
			return nil
		}
		return fmt.Errorf("execution error: %w", err)
	}

	// For assignment statements, show the assigned value
	if _, ok := stmt.(interpreter.AssignStatement); ok {
		r.printResult(result)
	}

	return nil
}

// evaluateTypeDef evaluates a type definition.
func (r *REPL) evaluateTypeDef(input string) error {
	// Parse and load the type definition as a module item
	module, err := r.parseModule(input)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Load the module into the interpreter
	if err := r.interp.LoadModule(*module); err != nil {
		return fmt.Errorf("load error: %w", err)
	}

	// Get the type name from the module
	for _, item := range module.Items {
		if typeDef, ok := item.(*interpreter.TypeDef); ok {
			r.printf("Type '%s' defined\n", typeDef.Name)
			return nil
		}
	}

	r.printf("Type defined\n")
	return nil
}

// evaluateFunction evaluates a function definition.
func (r *REPL) evaluateFunction(input string) error {
	// Parse and load the function definition as a module item
	module, err := r.parseModule(input)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Load the module into the interpreter
	if err := r.interp.LoadModule(*module); err != nil {
		return fmt.Errorf("load error: %w", err)
	}

	// Get the function name from the module
	for _, item := range module.Items {
		if fn, ok := item.(*interpreter.Function); ok {
			// Also define the function in the REPL environment
			r.env.Define(fn.Name, *fn)
			r.printf("Function '%s' defined\n", fn.Name)
			return nil
		}
	}

	r.printf("Function defined\n")
	return nil
}

// parseExpression parses a string as an expression.
func (r *REPL) parseExpression(input string) (interpreter.Expr, error) {
	lexer := parser.NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}

	p := parser.NewParser(tokens)
	return p.ParseExpression()
}

// parseStatement parses a string as a statement.
func (r *REPL) parseStatement(input string) (interpreter.Statement, error) {
	lexer := parser.NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}

	p := parser.NewParser(tokens)
	return p.ParseStatement()
}

// parseModule parses a string as a module (for type and function definitions).
func (r *REPL) parseModule(input string) (*interpreter.Module, error) {
	lexer := parser.NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}

	p := parser.NewParser(tokens)
	return p.Parse()
}

// isInputComplete checks if the input has balanced braces and parentheses.
func (r *REPL) isInputComplete(input string) bool {
	braceCount := 0
	parenCount := 0
	bracketCount := 0
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(input); i++ {
		ch := input[i]

		// Handle string literals
		if ch == '"' || ch == '\'' || ch == '`' {
			if !inString {
				inString = true
				stringChar = ch
			} else if ch == stringChar && (i == 0 || input[i-1] != '\\') {
				inString = false
			}
			continue
		}

		if inString {
			continue
		}

		switch ch {
		case '{':
			braceCount++
		case '}':
			braceCount--
		case '(':
			parenCount++
		case ')':
			parenCount--
		case '[':
			bracketCount++
		case ']':
			bracketCount--
		}
	}

	return braceCount == 0 && parenCount == 0 && bracketCount == 0 && !inString
}

// printWelcome prints the welcome message.
func (r *REPL) printWelcome() {
	r.printf("Glyph REPL v%s\n", r.version)
	r.printf("Type :help for available commands, :quit to exit\n")
	r.printf("=========================================\n\n")
}

// printGoodbye prints the goodbye message.
func (r *REPL) printGoodbye() {
	r.printf("\nGoodbye!\n")
}

// printPrompt prints the command prompt.
func (r *REPL) printPrompt() {
	if r.inputBuffer.Len() > 0 {
		// Continuation prompt for multi-line input
		r.printf("... ")
	} else {
		r.printf("glyph> ")
	}
}

// readLine reads a line of input.
func (r *REPL) readLine() (string, error) {
	return r.reader.ReadString('\n')
}

// printf writes formatted output.
func (r *REPL) printf(format string, args ...interface{}) {
	fmt.Fprintf(r.writer, format, args...)
}

// printResult prints an evaluation result.
func (r *REPL) printResult(result interface{}) {
	if result == nil {
		r.printf("nil\n")
		return
	}

	r.printf("=> %s\n", formatValue(result))
}

// formatValue formats a value for display.
func formatValue(v interface{}) string {
	if v == nil {
		return "nil"
	}

	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case int:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case []interface{}:
		var parts []string
		for _, elem := range val {
			parts = append(parts, formatValue(elem))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]interface{}:
		var parts []string
		for k, elem := range val {
			parts = append(parts, fmt.Sprintf("%s: %s", k, formatValue(elem)))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// GetEnvironment returns the current REPL environment.
func (r *REPL) GetEnvironment() *interpreter.Environment {
	return r.env
}

// GetInterpreter returns the REPL interpreter.
func (r *REPL) GetInterpreter() *interpreter.Interpreter {
	return r.interp
}

// LoadFile loads and executes a Glyph file.
func (r *REPL) LoadFile(filepath string) error {
	source, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	module, err := r.parseModule(string(source))
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	if err := r.interp.LoadModule(*module); err != nil {
		return fmt.Errorf("load error: %w", err)
	}

	// Copy functions to the REPL environment
	for _, item := range module.Items {
		if fn, ok := item.(*interpreter.Function); ok {
			r.env.Define(fn.Name, *fn)
		}
	}

	return nil
}

// Reset resets the REPL state.
func (r *REPL) Reset() {
	r.env = interpreter.NewEnvironment()
	r.interp = interpreter.NewInterpreter()
	r.interp.GetModuleResolver().SetParseFunc(func(source string) (*interpreter.Module, error) {
		lexer := parser.NewLexer(source)
		tokens, err := lexer.Tokenize()
		if err != nil {
			return nil, err
		}
		p := parser.NewParser(tokens)
		return p.Parse()
	})
	r.inputBuffer.Reset()
	r.lineNumber = 1
}
