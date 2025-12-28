package errors

import (
	"errors"
	"strings"
	"testing"
)

// Test CompileError creation and formatting
func TestNewCompileError(t *testing.T) {
	source := `func add(a: int, b: int) -> int {
    return a + b
}`
	snippet := ExtractSourceSnippet(source, 2)

	err := NewCompileError(
		"Invalid syntax in return statement",
		2,
		12,
		snippet,
		"Check your return statement syntax",
	)

	if err.Message != "Invalid syntax in return statement" {
		t.Errorf("Expected message 'Invalid syntax in return statement', got '%s'", err.Message)
	}

	if err.Line != 2 {
		t.Errorf("Expected line 2, got %d", err.Line)
	}

	if err.Column != 12 {
		t.Errorf("Expected column 12, got %d", err.Column)
	}

	if err.Suggestion != "Check your return statement syntax" {
		t.Errorf("Expected suggestion 'Check your return statement syntax', got '%s'", err.Suggestion)
	}

	if err.ErrorType != "Compile Error" {
		t.Errorf("Expected error type 'Compile Error', got '%s'", err.ErrorType)
	}
}

// Test ParseError creation
func TestNewParseError(t *testing.T) {
	source := `@ GET /users GET {
    $ users = [1, 2, 3
}`
	snippet := ExtractSourceSnippet(source, 2)

	err := NewParseError(
		"Missing closing bracket",
		2,
		24,
		snippet,
		"Add a closing bracket ']' to complete the array",
	)

	if err.ErrorType != "Parse Error" {
		t.Errorf("Expected error type 'Parse Error', got '%s'", err.ErrorType)
	}
}

// Test TypeError creation
func TestNewTypeError(t *testing.T) {
	source := `$ count: int = "hello"`
	snippet := ExtractSourceSnippet(source, 1)

	err := NewTypeError(
		"Type mismatch: expected int, got string",
		1,
		16,
		snippet,
		"Convert the string to an integer or change the type to 'str'",
	)

	if err.ErrorType != "Type Error" {
		t.Errorf("Expected error type 'Type Error', got '%s'", err.ErrorType)
	}

	if !strings.Contains(err.Message, "Type mismatch") {
		t.Errorf("Expected message to contain 'Type mismatch', got '%s'", err.Message)
	}
}

// Test RuntimeError creation and formatting
func TestNewRuntimeError(t *testing.T) {
	err := NewRuntimeError("Division by zero")

	if err.Message != "Division by zero" {
		t.Errorf("Expected message 'Division by zero', got '%s'", err.Message)
	}

	if err.ErrorType != "Runtime Error" {
		t.Errorf("Expected error type 'Runtime Error', got '%s'", err.ErrorType)
	}

	if len(err.StackTrace) != 0 {
		t.Errorf("Expected empty stack trace, got %d frames", len(err.StackTrace))
	}

	if len(err.Scope) != 0 {
		t.Errorf("Expected empty scope, got %d variables", len(err.Scope))
	}
}

// Test RuntimeError with context
func TestRuntimeErrorWithContext(t *testing.T) {
	err := NewRuntimeError("Variable 'count' is undefined").
		WithRoute("/users GET").
		WithExpression("return count + 1").
		WithSuggestion("Define the variable before using it: $ count = 0").
		WithScope(map[string]interface{}{
			"users": []int{1, 2, 3},
			"total": 10,
		}).
		WithStackFrame("handler", "/users GET", 5).
		WithStackFrame("main", "main.abc", 1)

	if err.Route != "/users GET" {
		t.Errorf("Expected route '/users GET', got '%s'", err.Route)
	}

	if err.Expression != "return count + 1" {
		t.Errorf("Expected expression 'return count + 1', got '%s'", err.Expression)
	}

	if err.Suggestion != "Define the variable before using it: $ count = 0" {
		t.Errorf("Expected suggestion, got '%s'", err.Suggestion)
	}

	if len(err.StackTrace) != 2 {
		t.Errorf("Expected 2 stack frames, got %d", len(err.StackTrace))
	}

	if len(err.Scope) != 2 {
		t.Errorf("Expected 2 variables in scope, got %d", len(err.Scope))
	}

	if err.StackTrace[0].Function != "handler" {
		t.Errorf("Expected first frame function 'handler', got '%s'", err.StackTrace[0].Function)
	}

	if err.StackTrace[1].Function != "main" {
		t.Errorf("Expected second frame function 'main', got '%s'", err.StackTrace[1].Function)
	}
}

// Test error formatting with colors
func TestCompileErrorFormatWithColors(t *testing.T) {
	source := `$ result: int = "not a number"`
	snippet := ExtractSourceSnippet(source, 1)

	err := NewTypeError(
		"Type mismatch: expected int, got string",
		1,
		17,
		snippet,
		"Convert the string to an integer or change the type to 'str'",
	)

	formatted := err.FormatError(true)

	// Check that the formatted error contains key elements
	if !strings.Contains(formatted, "Type Error") {
		t.Errorf("Expected formatted error to contain 'Type Error'")
	}

	if !strings.Contains(formatted, "line 1, column 17") {
		t.Errorf("Expected formatted error to contain 'line 1, column 17'")
	}

	if !strings.Contains(formatted, "Type mismatch") {
		t.Errorf("Expected formatted error to contain 'Type mismatch'")
	}

	if !strings.Contains(formatted, "Suggestion:") {
		t.Errorf("Expected formatted error to contain 'Suggestion:'")
	}

	// Check for ANSI color codes
	if !strings.Contains(formatted, "\033[") {
		t.Errorf("Expected formatted error to contain ANSI color codes")
	}
}

// Test error formatting without colors
func TestCompileErrorFormatWithoutColors(t *testing.T) {
	source := `$ result: int = "not a number"`
	snippet := ExtractSourceSnippet(source, 1)

	err := NewTypeError(
		"Type mismatch: expected int, got string",
		1,
		17,
		snippet,
		"Convert the string to an integer or change the type to 'str'",
	)

	formatted := err.FormatError(false)

	// Check that the formatted error contains key elements
	if !strings.Contains(formatted, "Type Error") {
		t.Errorf("Expected formatted error to contain 'Type Error'")
	}

	if !strings.Contains(formatted, "line 1, column 17") {
		t.Errorf("Expected formatted error to contain 'line 1, column 17'")
	}

	// Check that there are no ANSI color codes
	if strings.Contains(formatted, "\033[") {
		t.Errorf("Expected formatted error to NOT contain ANSI color codes")
	}
}

// Test RuntimeError formatting
func TestRuntimeErrorFormat(t *testing.T) {
	err := NewRuntimeError("Cannot divide by zero").
		WithRoute("/calculate POST").
		WithExpression("result = a / b").
		WithSuggestion("Add a check to ensure the divisor is not zero before dividing").
		WithScope(map[string]interface{}{
			"a": 10,
			"b": 0,
		}).
		WithStackFrame("calculate", "/calculate POST", 3)

	formatted := err.FormatError(true)

	// Check that the formatted error contains key elements
	if !strings.Contains(formatted, "Runtime Error") {
		t.Errorf("Expected formatted error to contain 'Runtime Error'")
	}

	if !strings.Contains(formatted, "Cannot divide by zero") {
		t.Errorf("Expected formatted error to contain 'Cannot divide by zero'")
	}

	if !strings.Contains(formatted, "Route:") {
		t.Errorf("Expected formatted error to contain 'Route:'")
	}

	if !strings.Contains(formatted, "/calculate POST") {
		t.Errorf("Expected formatted error to contain '/calculate POST'")
	}

	if !strings.Contains(formatted, "Expression:") {
		t.Errorf("Expected formatted error to contain 'Expression:'")
	}

	if !strings.Contains(formatted, "result = a / b") {
		t.Errorf("Expected formatted error to contain 'result = a / b'")
	}

	if !strings.Contains(formatted, "Variables in scope:") {
		t.Errorf("Expected formatted error to contain 'Variables in scope:'")
	}

	if !strings.Contains(formatted, "Stack trace:") {
		t.Errorf("Expected formatted error to contain 'Stack trace:'")
	}

	if !strings.Contains(formatted, "Suggestion:") {
		t.Errorf("Expected formatted error to contain 'Suggestion:'")
	}
}

// Test source snippet extraction
func TestExtractSourceSnippet(t *testing.T) {
	source := `line 1
line 2
line 3
line 4
line 5`

	tests := []struct {
		line     int
		expected string
	}{
		{1, "line 1\nline 2"},
		{2, "line 1\nline 2\nline 3"},
		{3, "line 2\nline 3\nline 4"},
		{5, "line 4\nline 5\n"}, // Last line includes trailing newline
		{0, ""},
		{10, ""},
	}

	for _, tt := range tests {
		snippet := ExtractSourceSnippet(source, tt.line)
		if snippet != tt.expected {
			t.Errorf("ExtractSourceSnippet(%d) = %q, expected %q", tt.line, snippet, tt.expected)
		}
	}
}

// Test FormatError function
func TestFormatError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name: "CompileError",
			err: NewCompileError(
				"Syntax error",
				1,
				5,
				"test code",
				"Fix the syntax",
			),
			contains: []string{"Compile Error", "line 1, column 5", "Syntax error", "Suggestion:"},
		},
		{
			name: "RuntimeError",
			err: NewRuntimeError("Runtime error").
				WithRoute("/test GET"),
			contains: []string{"Runtime Error", "Runtime error", "Route:", "/test GET"},
		},
		{
			name:     "Generic error",
			err:      errors.New("generic error"),
			contains: []string{"Error:", "generic error"},
		},
		{
			name:     "Nil error",
			err:      nil,
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := FormatError(tt.err)

			if tt.err == nil {
				if formatted != "" {
					t.Errorf("Expected empty string for nil error, got %q", formatted)
				}
				return
			}

			for _, substr := range tt.contains {
				if !strings.Contains(formatted, substr) {
					t.Errorf("Expected formatted error to contain %q, got:\n%s", substr, formatted)
				}
			}
		})
	}
}

// Test WithLineInfo function
func TestWithLineInfo(t *testing.T) {
	source := `$ x = 10
$ y = 20
$ z = x + y`

	tests := []struct {
		name   string
		err    error
		line   int
		col    int
		checks func(*testing.T, error)
	}{
		{
			name: "Wrap generic error",
			err:  errors.New("test error"),
			line: 2,
			col:  5,
			checks: func(t *testing.T, err error) {
				ce, ok := err.(*CompileError)
				if !ok {
					t.Fatalf("Expected *CompileError, got %T", err)
				}
				if ce.Line != 2 {
					t.Errorf("Expected line 2, got %d", ce.Line)
				}
				if ce.Column != 5 {
					t.Errorf("Expected column 5, got %d", ce.Column)
				}
				if ce.Message != "test error" {
					t.Errorf("Expected message 'test error', got '%s'", ce.Message)
				}
			},
		},
		{
			name: "Update existing CompileError",
			err: &CompileError{
				Message: "original error",
				Line:    1,
				Column:  1,
			},
			line: 3,
			col:  10,
			checks: func(t *testing.T, err error) {
				ce, ok := err.(*CompileError)
				if !ok {
					t.Fatalf("Expected *CompileError, got %T", err)
				}
				if ce.Line != 3 {
					t.Errorf("Expected line 3, got %d", ce.Line)
				}
				if ce.Column != 10 {
					t.Errorf("Expected column 10, got %d", ce.Column)
				}
				if ce.Message != "original error" {
					t.Errorf("Expected message 'original error', got '%s'", ce.Message)
				}
			},
		},
		{
			name: "Nil error",
			err:  nil,
			line: 1,
			col:  1,
			checks: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("Expected nil, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WithLineInfo(tt.err, tt.line, tt.col, source)
			tt.checks(t, result)
		})
	}
}

// Test WithSuggestion function
func TestWithSuggestion(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		suggestion string
		checks     func(*testing.T, error)
	}{
		{
			name:       "Add suggestion to CompileError",
			err:        NewCompileError("error", 1, 1, "", ""),
			suggestion: "Try this fix",
			checks: func(t *testing.T, err error) {
				ce, ok := err.(*CompileError)
				if !ok {
					t.Fatalf("Expected *CompileError, got %T", err)
				}
				if ce.Suggestion != "Try this fix" {
					t.Errorf("Expected suggestion 'Try this fix', got '%s'", ce.Suggestion)
				}
			},
		},
		{
			name:       "Add suggestion to RuntimeError",
			err:        NewRuntimeError("runtime error"),
			suggestion: "Check your code",
			checks: func(t *testing.T, err error) {
				re, ok := err.(*RuntimeError)
				if !ok {
					t.Fatalf("Expected *RuntimeError, got %T", err)
				}
				if re.Suggestion != "Check your code" {
					t.Errorf("Expected suggestion 'Check your code', got '%s'", re.Suggestion)
				}
			},
		},
		{
			name:       "Add suggestion to generic error",
			err:        errors.New("generic error"),
			suggestion: "Fix it",
			checks: func(t *testing.T, err error) {
				ce, ok := err.(*CompileError)
				if !ok {
					t.Fatalf("Expected *CompileError, got %T", err)
				}
				if ce.Suggestion != "Fix it" {
					t.Errorf("Expected suggestion 'Fix it', got '%s'", ce.Suggestion)
				}
			},
		},
		{
			name:       "Nil error",
			err:        nil,
			suggestion: "suggestion",
			checks: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("Expected nil, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WithSuggestion(tt.err, tt.suggestion)
			tt.checks(t, result)
		})
	}
}

// Test suggestion helper functions
func TestGetSuggestionForUndefinedVariable(t *testing.T) {
	tests := []struct {
		varName       string
		availableVars []string
		contains      string
	}{
		{
			varName:       "cunt",
			availableVars: []string{"count", "counter", "total"},
			contains:      "count",
		},
		{
			varName:       "userx",
			availableVars: []string{"user", "users", "username"},
			contains:      "user",
		},
		{
			varName:       "unknown",
			availableVars: []string{"foo", "bar"},
			contains:      "Make sure to define the variable before using it",
		},
	}

	for _, tt := range tests {
		t.Run(tt.varName, func(t *testing.T) {
			suggestion := GetSuggestionForUndefinedVariable(tt.varName, tt.availableVars)
			if !strings.Contains(suggestion, tt.contains) {
				t.Errorf("Expected suggestion to contain %q, got %q", tt.contains, suggestion)
			}
		})
	}
}

// Test type mismatch suggestions
func TestGetSuggestionForTypeMismatch(t *testing.T) {
	tests := []struct {
		expected string
		actual   string
		contains string
	}{
		{
			expected: "int",
			actual:   "string",
			contains: "Convert the string to an integer",
		},
		{
			expected: "string",
			actual:   "int",
			contains: "Convert the integer to a string",
		},
		{
			expected: "bool",
			actual:   "int",
			contains: "Use a boolean value",
		},
		{
			expected: "float",
			actual:   "string",
			contains: "Expected type 'float' but got 'string'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected+"_"+tt.actual, func(t *testing.T) {
			suggestion := GetSuggestionForTypeMismatch(tt.expected, tt.actual)
			if !strings.Contains(suggestion, tt.contains) {
				t.Errorf("Expected suggestion to contain %q, got %q", tt.contains, suggestion)
			}
		})
	}
}

// Test Levenshtein distance calculation
func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"hello", "", 5},
		{"", "world", 5},
		{"hello", "hello", 0},
		{"hello", "hallo", 1},
		{"kitten", "sitting", 3},
		{"count", "cunt", 1},
		{"user", "usr", 1},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			distance := levenshteinDistance(tt.s1, tt.s2)
			if distance != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, expected %d", tt.s1, tt.s2, distance, tt.expected)
			}
		})
	}
}

// Test isSimilar function
func TestIsSimilar(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected bool
	}{
		{"count", "count", false},  // Exact match is not similar
		{"count", "cunt", true},    // Typo
		{"userx", "user", true},    // Small edit distance
		{"users", "user", true},    // Prefix
		{"total", "count", false},  // Completely different
		{"hello", "hi", false},     // Too different
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			result := isSimilar(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("isSimilar(%q, %q) = %v, expected %v", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkCompileErrorFormat(b *testing.B) {
	source := `$ result: int = "not a number"`
	snippet := ExtractSourceSnippet(source, 1)
	err := NewTypeError(
		"Type mismatch: expected int, got string",
		1,
		17,
		snippet,
		"Convert the string to an integer",
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.FormatError(true)
	}
}

func BenchmarkRuntimeErrorFormat(b *testing.B) {
	err := NewRuntimeError("Division by zero").
		WithRoute("/calculate POST").
		WithExpression("result = a / b").
		WithScope(map[string]interface{}{
			"a": 10,
			"b": 0,
		})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.FormatError(true)
	}
}

func BenchmarkLevenshteinDistance(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = levenshteinDistance("kitten", "sitting")
	}
}
