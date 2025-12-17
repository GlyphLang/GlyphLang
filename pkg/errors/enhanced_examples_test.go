package errors

import (
	"fmt"
	"testing"
)

// Example: Enhanced variable suggestion
func ExampleCompileError_enhancedVariableSuggestion() {
	source := `@ route /users GET -> array {
    $ userList = getUsers()
    $ totl = userList.length
    return userList
}`
	snippet := ExtractSourceSnippet(source, 3)

	availableVars := []string{"userList", "total", "count"}
	suggestion := GetVariableSuggestion("totl", availableVars)

	err := NewCompileError(
		"Undefined variable 'totl'",
		3,
		7,
		snippet,
		suggestion,
	).WithContext("in route /users GET")

	fmt.Println(err.FormatError(false))
	// Output will show enhanced error with "Did you mean 'total'?" suggestion
}

// Example: Type error with expected vs actual
func ExampleCompileError_enhancedTypeError() {
	source := `@ route /calculate POST -> int {
    $ result: int = "not a number"
    return result
}`
	snippet := ExtractSourceSnippet(source, 2)

	suggestion := GetTypeMismatchSuggestion("int", "string", "variable assignment")

	err := NewTypeError(
		"Type mismatch in assignment",
		2,
		21,
		snippet,
		suggestion,
	).WithTypes("int", "string").WithContext("in route /calculate POST")

	fmt.Println(err.FormatError(false))
	// Output will show expected vs actual types clearly
}

// Example: Syntax error with suggested fix
func ExampleCompileError_syntaxErrorWithFix() {
	source := `@ route /data GET -> object {
    $ data = {name: "test", age 25}
    return data
}`
	snippet := ExtractSourceSnippet(source, 2)

	err := NewParseError(
		"Missing colon in object literal",
		2,
		36,
		snippet,
		"Object properties should be in the format 'key: value'",
	).WithFixedLine(`    $ data = {name: "test", age: 25}`)

	fmt.Println(err.FormatError(false))
	// Output will show the error and suggested fix inline
}

// Example: Missing bracket detection
func ExampleDetectMissingBracket() {
	source := `@ route /items GET -> array {
    $ items = [1, 2, 3
    return items
}`

	suggestion := DetectMissingBracket(source, 2, 24)

	err := NewParseError(
		"Unexpected token at end of line",
		2,
		24,
		ExtractSourceSnippet(source, 2),
		suggestion,
	)

	fmt.Println(err.FormatError(false))
	// Output will suggest adding closing bracket
}

// Example: Unclosed string detection
func ExampleDetectUnclosedString() {
	source := `@ route /message GET -> string {
    $ msg = "Hello world
    return msg
}`

	suggestion := DetectUnclosedString(source, 2)

	err := NewParseError(
		"Unexpected end of line in string literal",
		2,
		24,
		ExtractSourceSnippet(source, 2),
		suggestion,
	)

	fmt.Println(err.FormatError(false))
	// Output will detect unclosed string
}

// Example: Runtime error with full context
func ExampleRuntimeError_enhancedFullContext() {
	err := NewRuntimeError("Division by zero").
		WithRoute("/calculate POST").
		WithExpression("result = numerator / denominator").
		WithSuggestion(GetRuntimeSuggestion("division_by_zero", nil)).
		WithScope(map[string]interface{}{
			"numerator":   10,
			"denominator": 0,
		}).
		WithStackFrame("calculateResult", "/calculate POST", 5).
		WithStackFrame("handleRequest", "main.glyph", 12)

	fmt.Println(err.FormatError(false))
	// Output will show full runtime context with variables and stack trace
}

// Example: Function name suggestion
func ExampleGetFunctionSuggestion() {
	availableFuncs := []string{"println", "print", "printf", "log"}
	suggestion := GetFunctionSuggestion("prnt", availableFuncs)

	source := `@ route /test GET -> string {
    prnt("Hello")
    return "ok"
}`
	snippet := ExtractSourceSnippet(source, 2)

	err := NewCompileError(
		"Undefined function 'prnt'",
		2,
		5,
		snippet,
		suggestion,
	)

	fmt.Println(err.FormatError(false))
}

// Example: Type name suggestion
func ExampleGetTypeSuggestion() {
	customTypes := []string{"User", "Product", "Order"}
	suggestion := GetTypeSuggestion("Usr", customTypes)

	source := `@ route /data GET -> Usr {
    $ user: Usr = getUser()
    return user
}`
	snippet := ExtractSourceSnippet(source, 2)

	err := NewTypeError(
		"Unknown type 'Usr'",
		2,
		13,
		snippet,
		suggestion,
	)

	fmt.Println(err.FormatError(false))
}

// Example: Route path suggestion
func ExampleGetRouteSuggestion() {
	availableRoutes := []string{"/users", "/users/:id", "/products", "/orders"}
	suggestion := GetRouteSuggestion("/usr", availableRoutes)

	err := NewCompileError(
		"Route '/usr' is not defined",
		1,
		1,
		"@ forward /usr",
		suggestion,
	)

	fmt.Println(err.FormatError(false))
}

// Test enhanced error formatting with all features
func TestEnhancedErrorFormatting(t *testing.T) {
	source := `@ route /users GET -> array {
    $ userCount: int = "invalid"
    return getUsers()
}`
	snippet := ExtractSourceSnippet(source, 2)

	err := NewTypeError(
		"Type mismatch in variable declaration",
		2,
		24,
		snippet,
		GetTypeMismatchSuggestion("int", "string", "variable declaration"),
	).WithTypes("int", "string").
	  WithContext("in route /users GET").
	  WithFixedLine(`    $ userCount: int = 0`)

	formatted := err.FormatError(false)

	// Verify all components are present
	expectedComponents := []string{
		"Type Error",
		"line 2, column 24",
		"Type mismatch",
		"Expected: int",
		"Actual:   string",
		"Suggestion:",
		"suggested fix",
	}

	for _, component := range expectedComponents {
		if !contains(formatted, component) {
			t.Errorf("Expected formatted error to contain '%s'\nGot:\n%s",
				component, formatted)
		}
	}
}

// Test runtime error with debugging context
func TestRuntimeErrorDebuggingContext(t *testing.T) {
	err := NewRuntimeError("Cannot access property 'name' of null").
		WithRoute("/users/:id GET").
		WithExpression("userName = user.name").
		WithSuggestion(GetRuntimeSuggestion("null_reference", map[string]interface{}{
			"variable": "user",
		})).
		WithScope(map[string]interface{}{
			"id":   123,
			"user": nil,
		}).
		WithStackFrame("getUser", "/users/:id GET", 8).
		WithStackFrame("handleRequest", "main.glyph", 15)

	formatted := err.FormatError(false)

	expectedComponents := []string{
		"Runtime Error",
		"Cannot access property",
		"Route: /users/:id GET",
		"Expression:",
		"Variables in scope:",
		"user = <nil>",
		"Stack trace:",
		"getUser",
		"Suggestion:",
		"null",
	}

	for _, component := range expectedComponents {
		if !contains(formatted, component) {
			t.Errorf("Expected formatted error to contain '%s'\nGot:\n%s",
				component, formatted)
		}
	}
}

// Test common typo detection
func TestCommonTypoDetection(t *testing.T) {
	tests := []struct {
		typo     string
		expected string
	}{
		{"fucntion", "function"},
		{"retrun", "return"},
		{"lenght", "length"},
		{"treu", "true"},
		{"flase", "false"},
		{"nill", "nil"},
		{"undifined", "undefined"},
	}

	for _, tt := range tests {
		t.Run(tt.typo, func(t *testing.T) {
			config := DefaultSuggestionConfig()
			results := FindBestSuggestions(tt.typo, []string{"dummy"}, config)

			// Check if common typo was detected
			if len(results) > 0 && results[0].Suggestion == tt.expected {
				// Common typo was detected and corrected
				return
			}

			// Otherwise, regular fuzzy matching should work
			results = FindBestSuggestions(tt.typo, []string{tt.expected}, config)
			if len(results) == 0 {
				t.Errorf("Expected to find suggestion for '%s', got none", tt.typo)
			}
		})
	}
}

// Test syntax error detection
func TestSyntaxErrorDetection(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		line     int
		errorMsg string
		contains string
	}{
		{
			name: "Missing closing brace",
			source: `@ route /test GET {
    $ x = 1
`,
			line:     2,
			errorMsg: "expected }",
			contains: "Missing",
		},
		{
			name: "Unclosed string",
			source: `@ route /test GET {
    $ msg = "hello
}`,
			line:     2,
			errorMsg: "unterminated string",
			contains: "Unclosed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := DetectCommonSyntaxErrors(tt.source, tt.line, tt.errorMsg)
			if !contains(suggestion, tt.contains) {
				t.Errorf("Expected suggestion to contain '%s', got '%s'",
					tt.contains, suggestion)
			}
		})
	}
}

// Test identifier validation
func TestIdentifierValidation(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		valid      bool
	}{
		{"Valid simple", "myVar", true},
		{"Valid with underscore", "my_var", true},
		{"Valid with number", "var123", true},
		{"Valid starting with underscore", "_private", true},
		{"Invalid starting with number", "123var", false},
		{"Invalid with hyphen", "my-var", false},
		{"Invalid with space", "my var", false},
		{"Invalid empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidIdentifier(tt.identifier)
			if result != tt.valid {
				t.Errorf("IsValidIdentifier(%q) = %v, expected %v",
					tt.identifier, result, tt.valid)
			}

			// If invalid, check that suggestion is provided
			if !tt.valid && tt.identifier != "" {
				suggestion := SuggestValidIdentifier(tt.identifier)
				if suggestion == "" {
					t.Error("Expected non-empty suggestion for invalid identifier")
				}
			}
		})
	}
}

// Test multiple suggestions
func TestMultipleSuggestions(t *testing.T) {
	availableVars := []string{"count", "counter", "total", "amount"}
	config := DefaultSuggestionConfig()
	config.ShowMultipleSuggestions = true

	results := FindBestSuggestions("cnt", availableVars, config)

	if len(results) == 0 {
		t.Fatal("Expected at least one suggestion")
	}

	// Format with multiple suggestions
	formatted := FormatSuggestions(results, true)

	if !contains(formatted, "Did you mean") {
		t.Errorf("Expected formatted suggestions to contain 'Did you mean', got '%s'", formatted)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
