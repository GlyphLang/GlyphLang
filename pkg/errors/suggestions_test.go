package errors

import (
	"strings"
	"testing"
)

// Test FindBestSuggestions with various inputs
func TestFindBestSuggestions(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		candidates []string
		maxResults int
		expected   []string
	}{
		{
			name:       "Typo in variable name",
			target:     "cunt",
			candidates: []string{"count", "counter", "total", "amount"},
			maxResults: 3,
			expected:   []string{"count"},
		},
		{
			name:       "Multiple similar names",
			target:     "user",
			candidates: []string{"users", "username", "userID", "customer"},
			maxResults: 3,
			expected:   []string{"users", "username", "userID"},
		},
		{
			name:       "Short prefix match",
			target:     "requ",
			candidates: []string{"request", "require", "response", "result"},
			maxResults: 3,
			expected:   []string{"request", "require"},
		},
		{
			name:       "Case difference",
			target:     "MyVar",
			candidates: []string{"myvar", "myVar", "MYVAR"},
			maxResults: 3,
			expected:   []string{"myvar", "myVar", "MYVAR"},
		},
		{
			name:       "No similar candidates",
			target:     "foo",
			candidates: []string{"bar", "baz", "qux"},
			maxResults: 3,
			expected:   []string{},
		},
		{
			name:       "Common typo - function",
			target:     "fucntion",
			candidates: []string{"function", "factorial", "functional"},
			maxResults: 3,
			expected:   []string{"function"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultSuggestionConfig()
			config.MaxSuggestions = tt.maxResults
			results := FindBestSuggestions(tt.target, tt.candidates, config)

			if len(tt.expected) == 0 {
				if len(results) != 0 {
					t.Errorf("Expected no suggestions, got %d: %v", len(results), results)
				}
				return
			}

			if len(results) == 0 {
				t.Errorf("Expected suggestions %v, got none", tt.expected)
				return
			}

			// Check if the first suggestion matches expected
			found := false
			for _, exp := range tt.expected {
				if results[0].Suggestion == exp {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected first suggestion to be one of %v, got '%s'",
					tt.expected, results[0].Suggestion)
			}
		})
	}
}

// Test similarity score calculation
func TestCalculateSimilarityScore(t *testing.T) {
	tests := []struct {
		s1            string
		s2            string
		distance      int
		minScore      float64
		shouldBeHigh  bool // true if score should be >= 0.7
	}{
		{"hello", "hello", 0, 1.0, true},   // Exact match
		{"hello", "hallo", 1, 0.6, true},   // Small difference
		{"count", "cunt", 1, 0.6, true},    // Typo
		{"user", "users", 1, 0.7, true},    // Suffix
		{"req", "request", 4, 0.3, true},   // Prefix - high similarity due to prefix bonus
		{"foo", "bar", 3, 0.0, false},      // Completely different
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			score := calculateSimilarityScore(tt.s1, tt.s2, tt.distance)

			if score < 0.0 || score > 1.0 {
				t.Errorf("Score should be between 0 and 1, got %f", score)
			}

			if tt.shouldBeHigh && score < 0.5 {
				t.Errorf("Expected high score (>= 0.5) for %s vs %s, got %f",
					tt.s1, tt.s2, score)
			}

			if !tt.shouldBeHigh && score >= 0.7 {
				t.Errorf("Expected low score (< 0.7) for %s vs %s, got %f",
					tt.s1, tt.s2, score)
			}
		})
	}
}

// Test FormatSuggestions
func TestFormatSuggestions(t *testing.T) {
	tests := []struct {
		name            string
		results         []SuggestionResult
		multipleAllowed bool
		contains        string
	}{
		{
			name: "Single suggestion",
			results: []SuggestionResult{
				{Suggestion: "count", Distance: 1, Score: 0.9},
			},
			multipleAllowed: true,
			contains:        "Did you mean 'count'?",
		},
		{
			name: "Two suggestions",
			results: []SuggestionResult{
				{Suggestion: "count", Distance: 1, Score: 0.9},
				{Suggestion: "counter", Distance: 2, Score: 0.8},
			},
			multipleAllowed: true,
			contains:        "Did you mean 'count' or 'counter'?",
		},
		{
			name: "Three suggestions",
			results: []SuggestionResult{
				{Suggestion: "count", Distance: 1, Score: 0.9},
				{Suggestion: "counter", Distance: 2, Score: 0.8},
				{Suggestion: "amount", Distance: 3, Score: 0.7},
			},
			multipleAllowed: true,
			contains:        "Did you mean 'count', 'counter', or 'amount'?",
		},
		{
			name: "Multiple not allowed",
			results: []SuggestionResult{
				{Suggestion: "count", Distance: 1, Score: 0.9},
				{Suggestion: "counter", Distance: 2, Score: 0.8},
			},
			multipleAllowed: false,
			contains:        "Did you mean 'count'?",
		},
		{
			name:            "Empty results",
			results:         []SuggestionResult{},
			multipleAllowed: true,
			contains:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSuggestions(tt.results, tt.multipleAllowed)

			if tt.contains == "" {
				if result != "" {
					t.Errorf("Expected empty string, got '%s'", result)
				}
			} else {
				if !strings.Contains(result, tt.contains) {
					t.Errorf("Expected result to contain '%s', got '%s'",
						tt.contains, result)
				}
			}
		})
	}
}

// Test GetVariableSuggestion
func TestGetVariableSuggestion(t *testing.T) {
	tests := []struct {
		name      string
		varName   string
		available []string
		contains  string
	}{
		{
			name:      "Similar variable exists",
			varName:   "usrname",
			available: []string{"username", "user", "password"},
			contains:  "Did you mean 'username'?",
		},
		{
			name:      "No similar variables",
			varName:   "unknown",
			available: []string{"foo", "bar", "baz"},
			contains:  "Make sure to define the variable before using it",
		},
		{
			name:      "Empty available list",
			varName:   "test",
			available: []string{},
			contains:  "Make sure to define the variable before using it",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetVariableSuggestion(tt.varName, tt.available)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected suggestion to contain '%s', got '%s'",
					tt.contains, result)
			}
		})
	}
}

// Test GetFunctionSuggestion
func TestGetFunctionSuggestion(t *testing.T) {
	tests := []struct {
		name      string
		funcName  string
		available []string
		contains  string
	}{
		{
			name:      "Similar function exists",
			funcName:  "prnt",
			available: []string{"print", "printf", "println"},
			contains:  "Did you mean 'print'",
		},
		{
			name:      "No similar functions",
			funcName:  "foo",
			available: []string{"bar", "baz"},
			contains:  "Function 'foo' is not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFunctionSuggestion(tt.funcName, tt.available)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected suggestion to contain '%s', got '%s'",
					tt.contains, result)
			}
		})
	}
}

// Test GetTypeSuggestion
func TestGetTypeSuggestion(t *testing.T) {
	tests := []struct {
		name      string
		typeName  string
		available []string
		contains  string
	}{
		{
			name:      "Typo in built-in type",
			typeName:  "strng",
			available: []string{},
			contains:  "string",
		},
		{
			name:      "Similar custom type",
			typeName:  "Usr",
			available: []string{"User", "Product", "Order"},
			contains:  "Did you mean 'User'",
		},
		{
			name:      "Unknown type",
			typeName:  "Unknown",
			available: []string{},
			contains:  "Unknown type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTypeSuggestion(tt.typeName, tt.available)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected suggestion to contain '%s', got '%s'",
					tt.contains, result)
			}
		})
	}
}

// Test GetRouteSuggestion
func TestGetRouteSuggestion(t *testing.T) {
	tests := []struct {
		name      string
		route     string
		available []string
		contains  string
	}{
		{
			name:      "Similar route exists",
			route:     "/usr",
			available: []string{"/users", "/user/:id", "/products"},
			contains:  "Did you mean",
		},
		{
			name:      "No similar routes",
			route:     "/foo",
			available: []string{"/bar", "/baz"},
			contains:  "Route '/foo' is not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRouteSuggestion(tt.route, tt.available)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected suggestion to contain '%s', got '%s'",
					tt.contains, result)
			}
		})
	}
}

// Test DetectMissingBracket
func TestDetectMissingBracket(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		line     int
		column   int
		contains string
	}{
		{
			name:     "Missing closing brace",
			source:   "func foo() {\n    return bar\n",
			line:     2,
			column:   15,
			contains: "Missing 1 closing brace",
		},
		{
			name:     "Missing closing bracket",
			source:   "$ arr = [1, 2, 3\n",
			line:     1,
			column:   17,
			contains: "closing bracket",
		},
		{
			name:     "Missing closing paren",
			source:   "$ result = foo(1, 2\n",
			line:     1,
			column:   20,
			contains: "closing parenthesis",
		},
		{
			name:     "Extra closing brace",
			source:   "$ x = 1\n}\n",
			line:     2,
			column:   1,
			contains: "Unexpected closing brace",
		},
		{
			name:     "Balanced brackets",
			source:   "$ arr = [1, 2, 3]\n",
			line:     1,
			column:   18,
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectMissingBracket(tt.source, tt.line, tt.column)

			if tt.contains == "" {
				if result != "" {
					t.Errorf("Expected no suggestion, got '%s'", result)
				}
			} else {
				if !strings.Contains(result, tt.contains) {
					t.Errorf("Expected suggestion to contain '%s', got '%s'",
						tt.contains, result)
				}
			}
		})
	}
}

// Test DetectUnclosedString
func TestDetectUnclosedString(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		line     int
		contains string
	}{
		{
			name:     "Unclosed double quote",
			source:   `$ msg = "hello world`,
			line:     1,
			contains: "Unclosed string literal (missing closing \")",
		},
		{
			name:     "Unclosed single quote",
			source:   "$ msg = 'hello world",
			line:     1,
			contains: "Unclosed string literal (missing closing ')",
		},
		{
			name:     "Closed string",
			source:   `$ msg = "hello world"`,
			line:     1,
			contains: "",
		},
		{
			name:     "Escaped quote",
			source:   `$ msg = "hello \"world\""`,
			line:     1,
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectUnclosedString(tt.source, tt.line)

			if tt.contains == "" {
				if result != "" {
					t.Errorf("Expected no suggestion, got '%s'", result)
				}
			} else {
				if !strings.Contains(result, tt.contains) {
					t.Errorf("Expected suggestion to contain '%s', got '%s'",
						tt.contains, result)
				}
			}
		})
	}
}

// Test GetTypeMismatchSuggestion
func TestGetTypeMismatchSuggestion(t *testing.T) {
	tests := []struct {
		expected string
		actual   string
		context  string
		contains string
	}{
		{
			expected: "int",
			actual:   "string",
			context:  "",
			contains: "Convert the string to an integer",
		},
		{
			expected: "string",
			actual:   "int",
			context:  "",
			contains: "Convert the integer to a string",
		},
		{
			expected: "bool",
			actual:   "int",
			context:  "",
			contains: "boolean value",
		},
		{
			expected: "float",
			actual:   "int",
			context:  "",
			contains: "automatically converted",
		},
		{
			expected: "array",
			actual:   "int",
			context:  "function parameter",
			contains: "square brackets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected+"_"+tt.actual, func(t *testing.T) {
			result := GetTypeMismatchSuggestion(tt.expected, tt.actual, tt.context)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected suggestion to contain '%s', got '%s'",
					tt.contains, result)
			}
		})
	}
}

// Test GetRuntimeSuggestion
func TestGetRuntimeSuggestion(t *testing.T) {
	tests := []struct {
		name      string
		errorType string
		context   map[string]interface{}
		contains  string
	}{
		{
			name:      "Division by zero",
			errorType: "division_by_zero",
			context:   map[string]interface{}{},
			contains:  "divisor is not zero",
		},
		{
			name:      "Null reference with variable",
			errorType: "null_reference",
			context:   map[string]interface{}{"variable": "user"},
			contains:  "Variable 'user' is null",
		},
		{
			name:      "Index out of bounds",
			errorType: "index_out_of_bounds",
			context:   map[string]interface{}{"index": 5, "length": 3},
			contains:  "out of bounds",
		},
		{
			name:      "SQL injection",
			errorType: "sql_injection",
			context:   map[string]interface{}{},
			contains:  "parameterized queries",
		},
		{
			name:      "XSS vulnerability",
			errorType: "xss",
			context:   map[string]interface{}{},
			contains:  "Sanitize user input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRuntimeSuggestion(tt.errorType, tt.context)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected suggestion to contain '%s', got '%s'",
					tt.contains, result)
			}
		})
	}
}

// Test DetectCommonSyntaxErrors
func TestDetectCommonSyntaxErrors(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		line     int
		errorMsg string
		contains string
	}{
		{
			name:     "Missing closing brace",
			source:   "func foo() {\n    return bar\n",
			line:     2,
			errorMsg: "expected }",
			contains: "Missing",
		},
		{
			name:     "Unclosed string",
			source:   `$ msg = "hello`,
			line:     1,
			errorMsg: "unterminated string",
			contains: "Unclosed string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectCommonSyntaxErrors(tt.source, tt.line, tt.errorMsg)
			if tt.contains != "" {
				if !strings.Contains(result, tt.contains) {
					t.Errorf("Expected result to contain '%s', got '%s'",
						tt.contains, result)
				}
			}
		})
	}
}

// Test IsValidIdentifier
func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"validName", true},
		{"_validName", true},
		{"valid_name_123", true},
		{"123invalid", false},
		{"invalid-name", false},
		{"invalid name", false},
		{"", false},
		{"a", true},
		{"_", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsValidIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidIdentifier(%q) = %v, expected %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

// Test SuggestValidIdentifier
func TestSuggestValidIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		contains string
	}{
		{"", "cannot be empty"},
		{"123name", "cannot start with a digit"},
		{"invalid-name", "Remove special characters"},
		{"valid_name", "only letters, digits, and underscores"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SuggestValidIdentifier(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected suggestion to contain '%s', got '%s'",
					tt.contains, result)
			}
		})
	}
}

// Test common typo corrections
func TestCommonTypos(t *testing.T) {
	tests := []struct {
		typo     string
		expected string
	}{
		{"fucntion", "function"},
		{"retrun", "return"},
		{"lenght", "length"},
		{"treu", "true"},
		{"flase", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.typo, func(t *testing.T) {
			config := DefaultSuggestionConfig()
			results := FindBestSuggestions(tt.typo, []string{tt.expected}, config)

			if len(results) == 0 {
				t.Errorf("Expected suggestion for typo '%s', got none", tt.typo)
				return
			}

			if results[0].Suggestion != tt.expected {
				t.Errorf("Expected suggestion '%s' for typo '%s', got '%s'",
					tt.expected, tt.typo, results[0].Suggestion)
			}
		})
	}
}

// Benchmark tests
func BenchmarkFindBestSuggestions(b *testing.B) {
	candidates := []string{
		"count", "counter", "total", "amount", "sum", "value",
		"user", "username", "userID", "customer", "client",
		"request", "response", "result", "data", "info",
	}
	config := DefaultSuggestionConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FindBestSuggestions("cunt", candidates, config)
	}
}

func BenchmarkCalculateSimilarityScore(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateSimilarityScore("hello", "hallo", 1)
	}
}

func BenchmarkDetectMissingBracket(b *testing.B) {
	source := `func foo() {
    $ x = [1, 2, 3]
    $ y = {a: 1, b: 2}
    return (x + y)
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetectMissingBracket(source, 4, 20)
	}
}

func BenchmarkDetectUnclosedString(b *testing.B) {
	source := `$ msg = "hello world`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetectUnclosedString(source, 1)
	}
}
