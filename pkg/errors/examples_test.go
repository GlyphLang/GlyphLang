package errors

import (
	"fmt"
	"testing"
)

// This file contains example tests that demonstrate the error formatting
// Run with: go test -v -run Example

func ExampleCompileError_parseError() {
	source := `@ GET /users GET {
    $ users = [1, 2, 3
    return users
}`

	err := NewParseError(
		"Missing closing bracket in array literal",
		2,
		24,
		ExtractSourceSnippet(source, 2),
		"Add a closing bracket ']' to complete the array definition",
	)

	// Print without colors for consistent output
	output := err.FormatError(false)
	fmt.Print(output)
}

func ExampleCompileError_typeError() {
	source := `$ age: int = "twenty-five"`

	err := NewTypeError(
		"Type mismatch: cannot assign 'string' to variable of type 'int'",
		1,
		14,
		ExtractSourceSnippet(source, 1),
		"Convert the string to an integer (e.g., 25) or change the type annotation to 'str'",
	)

	output := err.FormatError(false)
	fmt.Print(output)
}

func ExampleRuntimeError_divisionByZero() {
	err := NewRuntimeError("Division by zero error").
		WithRoute("/calculate POST").
		WithExpression("result = total / count").
		WithScope(map[string]interface{}{
			"total": 100,
			"count": 0,
		}).
		WithSuggestion("Add a check to ensure 'count' is not zero before performing division").
		WithStackFrame("calculateAverage", "/calculate POST", 12).
		WithStackFrame("main", "main.abc", 5)

	output := err.FormatError(false)
	fmt.Print(output)
}

func ExampleRuntimeError_undefinedVariable() {
	err := NewRuntimeError("Variable 'usrName' is not defined").
		WithRoute("/profile GET").
		WithExpression("return { name: usrName, id: userId }").
		WithScope(map[string]interface{}{
			"userId":   123,
			"userName": "john_doe",
		}).
		WithSuggestion("Did you mean 'userName'? Or define the variable before using it: $ usrName = value")

	output := err.FormatError(false)
	fmt.Print(output)
}

// Test to demonstrate before/after comparison
func TestErrorMessageComparison(t *testing.T) {
	// BEFORE: Simple error message
	beforeError := fmt.Sprintf("Error at line 3: undefined variable 'usr'")

	// AFTER: Enhanced error message
	afterError := NewRuntimeError("Variable 'usr' is not defined").
		WithRoute("/users POST").
		WithExpression("$ result = database.save(usr)").
		WithScope(map[string]interface{}{
			"user":     map[string]interface{}{"name": "John", "email": "john@example.com"},
			"request":  map[string]interface{}{"method": "POST"},
			"database": "<database connection>",
		}).
		WithSuggestion("Did you mean 'user'? Variable names are case-sensitive")

	fmt.Println("=== BEFORE (Simple Error) ===")
	fmt.Println(beforeError)
	fmt.Println()

	fmt.Println("=== AFTER (Enhanced Error) ===")
	fmt.Println(afterError.FormatError(false))
}

// Test helper function suggestions
func TestSuggestionHelpers(t *testing.T) {
	tests := []struct {
		name       string
		suggestion string
	}{
		{
			name: "Undefined variable with suggestions",
			suggestion: GetSuggestionForUndefinedVariable("usrName", []string{
				"userName", "userId", "userEmail",
			}),
		},
		{
			name:       "Type mismatch: int expected, string provided",
			suggestion: GetSuggestionForTypeMismatch("int", "string"),
		},
		{
			name:       "Type mismatch: string expected, int provided",
			suggestion: GetSuggestionForTypeMismatch("string", "int"),
		},
		{
			name:       "Type mismatch: bool expected",
			suggestion: GetSuggestionForTypeMismatch("bool", "int"),
		},
		{
			name:       "Division by zero",
			suggestion: GetSuggestionForDivisionByZero(),
		},
		{
			name:       "SQL injection risk",
			suggestion: GetSuggestionForSQLInjection(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.suggestion == "" {
				t.Error("Suggestion should not be empty")
			}
			fmt.Printf("%s:\n  %s\n\n", tt.name, tt.suggestion)
		})
	}
}

// Demonstrate the FormatError function with different error types
func TestFormatErrorExamples(t *testing.T) {
	source := `@ GET /api/data GET {
    $ items = fetchData()
    return items.filter(x => x.value > 0)
}`

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "Parse Error",
			err: NewParseError(
				"Unexpected token '=>' in lambda expression",
				3,
				30,
				ExtractSourceSnippet(source, 3),
				"Lambda expressions are not yet supported. Use a regular function instead",
			),
		},
		{
			name: "Type Error",
			err: NewTypeError(
				"Method 'filter' does not exist on type 'array'",
				3,
				19,
				ExtractSourceSnippet(source, 3),
				"Arrays don't have a filter method. Use a for loop to filter items instead",
			),
		},
		{
			name: "Runtime Error",
			err: NewRuntimeError("Function 'fetchData' is not defined").
				WithRoute("/api/data GET").
				WithExpression("$ items = fetchData()").
				WithSuggestion("Define the fetchData function before calling it"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("\n=== %s ===\n", tt.name)
			fmt.Println(FormatError(tt.err))
		})
	}
}

// Demonstrate error wrapping with context
func TestErrorWrappingExamples(t *testing.T) {
	source := `$ count: int = getUserCount()
$ average = total / count
return average`

	// Start with a simple error
	baseErr := fmt.Errorf("division by zero")

	// Wrap with line information
	err1 := WithLineInfo(baseErr, 2, 13, source)
	fmt.Println("=== Error with line info ===")
	fmt.Println(FormatError(err1))

	// Add a suggestion
	err2 := WithSuggestion(err1, "Check that 'count' is not zero before dividing")
	fmt.Println("\n=== Error with suggestion ===")
	fmt.Println(FormatError(err2))

	// Add a filename
	err3 := WithFileName(err2, "calculations.abc")
	fmt.Println("\n=== Error with filename ===")
	fmt.Println(FormatError(err3))
}

// Demonstrate color vs no-color output
func TestColorFormatting(t *testing.T) {
	source := `$ result: int = "error"`

	err := NewTypeError(
		"Type mismatch: expected int, got string",
		1,
		17,
		ExtractSourceSnippet(source, 1),
		"Convert the string to an integer",
	)

	fmt.Println("=== Without Colors ===")
	fmt.Println(err.FormatError(false))

	fmt.Println("\n=== With Colors (ANSI codes visible) ===")
	fmt.Println(err.FormatError(true))
}
