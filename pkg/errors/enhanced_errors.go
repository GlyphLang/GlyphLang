package errors

import (
	"fmt"
	"strings"
)

// ANSI color codes for terminal output
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
	Bold   = "\033[1m"
)

// CompileError represents a compilation error with context
type CompileError struct {
	Message       string
	Line          int
	Column        int
	SourceSnippet string
	Suggestion    string
	FileName      string
	ErrorType     string
	FixedLine     string // Suggested fix for the error line
	ExpectedType  string // For type errors
	ActualType    string // For type errors
	Context       string // Additional context (e.g., "in function foo", "in route /users")
}

// Error implements the error interface
func (e *CompileError) Error() string {
	return e.FormatError(true)
}

// FormatError formats the error with optional color support
func (e *CompileError) FormatError(useColors bool) string {
	var builder strings.Builder

	// Header
	errorType := e.ErrorType
	if errorType == "" {
		errorType = "Compile Error"
	}

	if useColors {
		builder.WriteString(fmt.Sprintf("%s%s%s", Bold+Red, errorType, Reset))
	} else {
		builder.WriteString(errorType)
	}

	if e.FileName != "" {
		builder.WriteString(fmt.Sprintf(" in %s", e.FileName))
	}

	builder.WriteString(fmt.Sprintf(" at line %d, column %d\n", e.Line, e.Column))

	// Source snippet with context
	if e.SourceSnippet != "" {
		lines := strings.Split(e.SourceSnippet, "\n")
		lineNum := e.Line

		builder.WriteString("\n")

		// Show previous line for context if available
		if len(lines) > 1 && lineNum > 1 {
			prevLineNum := lineNum - 1
			if useColors {
				builder.WriteString(fmt.Sprintf("  %s%4d |%s %s\n", Gray, prevLineNum, Reset, lines[0]))
			} else {
				builder.WriteString(fmt.Sprintf("  %4d | %s\n", prevLineNum, lines[0]))
			}
		}

		// Show the error line
		errorLineIdx := 0
		if len(lines) > 1 {
			errorLineIdx = 1
		}

		if errorLineIdx < len(lines) {
			errorLine := lines[errorLineIdx]
			if useColors {
				builder.WriteString(fmt.Sprintf("  %s%4d |%s %s\n", Cyan, lineNum, Reset, errorLine))
			} else {
				builder.WriteString(fmt.Sprintf("  %4d | %s\n", lineNum, errorLine))
			}

			// Show caret pointing to error column with context
			if e.Column > 0 {
				spaces := strings.Repeat(" ", e.Column-1)
				if useColors {
					builder.WriteString(fmt.Sprintf("       %s|%s %s%s^ error here%s\n", Gray, Reset, Red, spaces, Reset))
				} else {
					builder.WriteString(fmt.Sprintf("       | %s^ error here\n", spaces))
				}
			}

			// Show suggested fix if available
			if e.FixedLine != "" {
				if useColors {
					builder.WriteString(fmt.Sprintf("  %s%4d |%s %s %s(suggested fix)%s\n",
						Green, lineNum, Reset, e.FixedLine, Gray, Reset))
				} else {
					builder.WriteString(fmt.Sprintf("  %4d | %s (suggested fix)\n", lineNum, e.FixedLine))
				}
			}
		}

		// Show next line for context if available
		nextLineIdx := errorLineIdx + 1
		if nextLineIdx < len(lines) && lineNum+1 <= lineNum+1 {
			if useColors {
				builder.WriteString(fmt.Sprintf("  %s%4d |%s %s\n", Gray, lineNum+1, Reset, lines[nextLineIdx]))
			} else {
				builder.WriteString(fmt.Sprintf("  %4d | %s\n", lineNum+1, lines[nextLineIdx]))
			}
		}
	}

	// Error message
	builder.WriteString("\n")
	if useColors {
		builder.WriteString(fmt.Sprintf("%s%s%s", Red, e.Message, Reset))
	} else {
		builder.WriteString(e.Message)
	}

	// Add context if available
	if e.Context != "" {
		if useColors {
			builder.WriteString(fmt.Sprintf(" %s%s%s", Gray, e.Context, Reset))
		} else {
			builder.WriteString(fmt.Sprintf(" %s", e.Context))
		}
	}
	builder.WriteString("\n")

	// For type errors, show expected vs actual clearly
	if e.ErrorType == "Type Error" && e.ExpectedType != "" && e.ActualType != "" {
		builder.WriteString("\n")
		if useColors {
			builder.WriteString(fmt.Sprintf("%sExpected:%s %s%s%s\n", Bold, Reset, Green, e.ExpectedType, Reset))
			builder.WriteString(fmt.Sprintf("%sActual:%s   %s%s%s\n", Bold, Reset, Red, e.ActualType, Reset))
		} else {
			builder.WriteString(fmt.Sprintf("Expected: %s\n", e.ExpectedType))
			builder.WriteString(fmt.Sprintf("Actual:   %s\n", e.ActualType))
		}
	}

	// Suggestion
	if e.Suggestion != "" {
		builder.WriteString("\n")
		if useColors {
			builder.WriteString(fmt.Sprintf("%s%sSuggestion:%s %s\n", Bold, Yellow, Reset, e.Suggestion))
		} else {
			builder.WriteString(fmt.Sprintf("Suggestion: %s\n", e.Suggestion))
		}
	}

	return builder.String()
}

// RuntimeError represents a runtime error with execution context
type RuntimeError struct {
	Message    string
	Route      string
	Expression string
	StackTrace []StackFrame
	Suggestion string
	ErrorType  string
	Scope      map[string]interface{}
}

// StackFrame represents a single frame in the call stack
type StackFrame struct {
	Function string
	Location string
	Line     int
}

// Error implements the error interface
func (e *RuntimeError) Error() string {
	return e.FormatError(true)
}

// FormatError formats the runtime error with optional color support
func (e *RuntimeError) FormatError(useColors bool) string {
	var builder strings.Builder

	// Header
	errorType := e.ErrorType
	if errorType == "" {
		errorType = "Runtime Error"
	}

	if useColors {
		builder.WriteString(fmt.Sprintf("%s%s%s\n", Bold+Red, errorType, Reset))
	} else {
		builder.WriteString(fmt.Sprintf("%s\n", errorType))
	}

	// Error message
	if useColors {
		builder.WriteString(fmt.Sprintf("%s%s%s\n", Red, e.Message, Reset))
	} else {
		builder.WriteString(fmt.Sprintf("%s\n", e.Message))
	}

	// Route context
	if e.Route != "" {
		builder.WriteString("\n")
		if useColors {
			builder.WriteString(fmt.Sprintf("%sRoute:%s %s\n", Bold, Reset, e.Route))
		} else {
			builder.WriteString(fmt.Sprintf("Route: %s\n", e.Route))
		}
	}

	// Expression context
	if e.Expression != "" {
		builder.WriteString("\n")
		if useColors {
			builder.WriteString(fmt.Sprintf("%sExpression:%s\n  %s\n", Bold, Reset, e.Expression))
		} else {
			builder.WriteString(fmt.Sprintf("Expression:\n  %s\n", e.Expression))
		}
	}

	// Scope variables (only show if available)
	if len(e.Scope) > 0 {
		builder.WriteString("\n")
		if useColors {
			builder.WriteString(fmt.Sprintf("%sVariables in scope:%s\n", Bold, Reset))
		} else {
			builder.WriteString("Variables in scope:\n")
		}

		for name, value := range e.Scope {
			if useColors {
				builder.WriteString(fmt.Sprintf("  %s%s%s = %v (%T)\n", Cyan, name, Reset, value, value))
			} else {
				builder.WriteString(fmt.Sprintf("  %s = %v (%T)\n", name, value, value))
			}
		}
	}

	// Stack trace
	if len(e.StackTrace) > 0 {
		builder.WriteString("\n")
		if useColors {
			builder.WriteString(fmt.Sprintf("%sStack trace:%s\n", Bold, Reset))
		} else {
			builder.WriteString("Stack trace:\n")
		}

		for i, frame := range e.StackTrace {
			if useColors {
				builder.WriteString(fmt.Sprintf("  %d. %s%s%s at %s:%d\n",
					i+1, Cyan, frame.Function, Reset, frame.Location, frame.Line))
			} else {
				builder.WriteString(fmt.Sprintf("  %d. %s at %s:%d\n",
					i+1, frame.Function, frame.Location, frame.Line))
			}
		}
	}

	// Suggestion
	if e.Suggestion != "" {
		builder.WriteString("\n")
		if useColors {
			builder.WriteString(fmt.Sprintf("%s%sSuggestion:%s %s\n", Bold, Yellow, Reset, e.Suggestion))
		} else {
			builder.WriteString(fmt.Sprintf("Suggestion: %s\n", e.Suggestion))
		}
	}

	return builder.String()
}

// Helper functions for creating enhanced errors

// NewCompileError creates a new compile error with context
func NewCompileError(message string, line, column int, sourceSnippet, suggestion string) *CompileError {
	return &CompileError{
		Message:       message,
		Line:          line,
		Column:        column,
		SourceSnippet: sourceSnippet,
		Suggestion:    suggestion,
		ErrorType:     "Compile Error",
	}
}

// WithFixedLine adds a suggested fix line to the error
func (e *CompileError) WithFixedLine(fixedLine string) *CompileError {
	e.FixedLine = fixedLine
	return e
}

// WithTypes adds type information for type errors
func (e *CompileError) WithTypes(expected, actual string) *CompileError {
	e.ExpectedType = expected
	e.ActualType = actual
	return e
}

// WithContext adds contextual information to the error
func (e *CompileError) WithContext(context string) *CompileError {
	e.Context = context
	return e
}

// NewParseError creates a parse-specific error
func NewParseError(message string, line, column int, sourceSnippet, suggestion string) *CompileError {
	return &CompileError{
		Message:       message,
		Line:          line,
		Column:        column,
		SourceSnippet: sourceSnippet,
		Suggestion:    suggestion,
		ErrorType:     "Parse Error",
	}
}

// NewTypeError creates a type-checking error
func NewTypeError(message string, line, column int, sourceSnippet, suggestion string) *CompileError {
	return &CompileError{
		Message:       message,
		Line:          line,
		Column:        column,
		SourceSnippet: sourceSnippet,
		Suggestion:    suggestion,
		ErrorType:     "Type Error",
	}
}

// NewRuntimeError creates a new runtime error
func NewRuntimeError(message string) *RuntimeError {
	return &RuntimeError{
		Message:    message,
		ErrorType:  "Runtime Error",
		StackTrace: []StackFrame{},
		Scope:      make(map[string]interface{}),
	}
}

// WithRoute adds route context to a runtime error
func (e *RuntimeError) WithRoute(route string) *RuntimeError {
	e.Route = route
	return e
}

// WithExpression adds expression context to a runtime error
func (e *RuntimeError) WithExpression(expr string) *RuntimeError {
	e.Expression = expr
	return e
}

// WithSuggestion adds a suggestion to a runtime error
func (e *RuntimeError) WithSuggestion(suggestion string) *RuntimeError {
	e.Suggestion = suggestion
	return e
}

// WithScope adds variable scope to a runtime error
func (e *RuntimeError) WithScope(scope map[string]interface{}) *RuntimeError {
	e.Scope = scope
	return e
}

// WithStackFrame adds a stack frame to a runtime error
func (e *RuntimeError) WithStackFrame(function, location string, line int) *RuntimeError {
	e.StackTrace = append(e.StackTrace, StackFrame{
		Function: function,
		Location: location,
		Line:     line,
	})
	return e
}

// Common error suggestions (now delegating to suggestions.go)

// GetSuggestionForUndefinedVariable suggests common fixes for undefined variables
func GetSuggestionForUndefinedVariable(varName string, availableVars []string) string {
	return GetVariableSuggestion(varName, availableVars)
}

// GetSuggestionForTypeMismatch suggests fixes for type mismatches
func GetSuggestionForTypeMismatch(expected, actual string) string {
	return GetTypeMismatchSuggestion(expected, actual, "")
}

// GetSuggestionForDivisionByZero suggests fixes for division by zero
func GetSuggestionForDivisionByZero() string {
	return "Add a check to ensure the divisor is not zero before dividing"
}

// GetSuggestionForSQLInjection suggests fixes for SQL injection vulnerabilities
func GetSuggestionForSQLInjection() string {
	return "Use parameterized queries or prepared statements instead of string concatenation"
}

// Helper function to check if two strings are similar
func isSimilar(s1, s2 string) bool {
	if s1 == s2 {
		return false // Exact match is not a suggestion
	}

	// Simple heuristics for similarity
	if strings.HasPrefix(s1, s2) || strings.HasPrefix(s2, s1) {
		return true
	}

	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		return true
	}

	// Check Levenshtein distance (simplified)
	if len(s1) > 3 && len(s2) > 3 {
		diff := levenshteinDistance(s1, s2)
		return diff <= 2
	}

	return false
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create a 2D slice for dynamic programming
	d := make([][]int, len(s1)+1)
	for i := range d {
		d[i] = make([]int, len(s2)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	// Calculate distances
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			d[i][j] = min(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}

	return d[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// ExtractSourceSnippet extracts a source code snippet around a specific line
func ExtractSourceSnippet(source string, line int) string {
	lines := strings.Split(source, "\n")

	if line <= 0 || line > len(lines) {
		return ""
	}

	var snippet strings.Builder

	// Include previous line if available
	if line > 1 {
		snippet.WriteString(lines[line-2])
		snippet.WriteString("\n")
	}

	// Include error line
	snippet.WriteString(lines[line-1])
	snippet.WriteString("\n")

	// Include next line if available
	if line < len(lines) {
		snippet.WriteString(lines[line])
	}

	return snippet.String()
}
