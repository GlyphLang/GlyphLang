package parser

import (
	"strings"
	"testing"
)

// TestImprovedErrorMessages demonstrates the improved error messages
func TestImprovedErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name: "missing opening brace in type definition",
			input: `: User
  name: str!`,
			expectError: true,
			errorCheck: func(err error) bool {
				return strings.Contains(err.Error(), "Expected {") &&
					strings.Contains(err.Error(), "Hint:")
			},
		},
		{
			name: "missing route path",
			input: `@ GET [GET] -> User
  > { name: "test" }`,
			expectError: true,
			errorCheck: func(err error) bool {
				return strings.Contains(err.Error(), "Expected route path") ||
					strings.Contains(err.Error(), "Expected identifier")
			},
		},
		{
			name: "invalid HTTP method",
			input: `@ route /users [INVALID] -> User
  > { name: "test" }`,
			expectError: true,
			errorCheck: func(err error) bool {
				return strings.Contains(err.Error(), "Unknown HTTP method") &&
					strings.Contains(err.Error(), "Valid HTTP methods are")
			},
		},
		{
			name: "unexpected token at top level",
			input: `{ invalid: "top level" }`,
			expectError: true,
			errorCheck: func(err error) bool {
				return strings.Contains(err.Error(), "Unexpected token") &&
					strings.Contains(err.Error(), "Top-level items must start with")
			},
		},
		{
			name: "unterminated string",
			input: `@ GET /test
  $ msg = "unterminated string
  > msg`,
			expectError: true,
			errorCheck: func(err error) bool {
				return strings.Contains(err.Error(), "Unterminated string") &&
					strings.Contains(err.Error(), "Hint:")
			},
		},
		{
			name: "invalid character",
			input: `@ GET /test
  $ msg = "hello" ; invalid semicolon`,
			expectError: true,
			errorCheck: func(err error) bool {
				return strings.Contains(err.Error(), "Unexpected character ';'") &&
					strings.Contains(err.Error(), "semicolons")
			},
		},
		{
			name: "type error with hint",
			input: `: User {
  name: 123!
}`,
			expectError: true,
			errorCheck: func(err error) bool {
				return strings.Contains(err.Error(), "Expected type name") &&
					strings.Contains(err.Error(), "Valid types are")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()

			if err != nil {
				if !tt.expectError {
					t.Fatalf("unexpected lexer error: %v", err)
				}
				if tt.errorCheck != nil && !tt.errorCheck(err) {
					t.Errorf("error message doesn't meet expectations:\n%v", err)
				}
				t.Logf("Error message:\n%v", err)
				return
			}

			parser := NewParserWithSource(tokens, tt.input)
			_, err = parser.Parse()

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if tt.errorCheck != nil && !tt.errorCheck(err) {
					t.Errorf("error message doesn't meet expectations:\n%v", err)
				}
				t.Logf("Error message:\n%v", err)
			} else {
				if err != nil {
					t.Fatalf("unexpected parser error: %v", err)
				}
			}
		})
	}
}

// TestErrorContextDisplay tests that errors show proper source context
func TestErrorContextDisplay(t *testing.T) {
	input := `@ route /users [GET] -> User
  + auth(jwt)
  $ users = db.query("SELECT * FROM users")
  > users`

	// Create error by modifying the input to have an error
	badInput := strings.Replace(input, "[GET]", "[INVALID]", 1)

	lexer := NewLexer(badInput)
	tokens, _ := lexer.Tokenize()
	parser := NewParserWithSource(tokens, badInput)
	_, err := parser.Parse()

	if err == nil {
		t.Fatal("expected error but got nil")
	}

	errStr := err.Error()

	// Check that error includes line number
	if !strings.Contains(errStr, "line") {
		t.Error("error message should contain line number")
	}

	// Check that error includes column number
	if !strings.Contains(errStr, "column") {
		t.Error("error message should contain column number")
	}

	// Check that error includes source line with caret
	if !strings.Contains(errStr, "^") {
		t.Error("error message should contain caret pointing to error")
	}

	t.Logf("Full error message:\n%v", err)
}

// TestParserErrorTypes tests different parser error types
func TestParserErrorTypes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "missing route keyword",
			input: `@ notroute /users`,
		},
		{
			name: "missing identifier in field",
			input: `: User {
  : str!
}`,
		},
		{
			name: "invalid statement",
			input: `@ GET /test
  invalid statement here`,
		},
		{
			name: "missing equals in assignment",
			input: `@ GET /test
  $ msg "hello"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Logf("Lexer error:\n%v", err)
				return
			}

			parser := NewParserWithSource(tokens, tt.input)
			_, err = parser.Parse()

			if err == nil {
				t.Fatal("expected error but got nil")
			}

			// Just log the error to see the formatting
			t.Logf("Error message:\n%v", err)
		})
	}
}
