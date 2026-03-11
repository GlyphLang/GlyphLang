package security

import (
	"strings"
	"testing"
)

func TestStripSQLComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "line comment removal",
			input:    "SELECT * FROM users -- get all users",
			expected: "SELECT * FROM users ",
		},
		{
			name:     "block comment removal",
			input:    "SELECT * /* columns */ FROM users",
			expected: "SELECT *  FROM users",
		},
		{
			name:     "single quotes escaped",
			input:    "Robert'; DROP TABLE users",
			expected: "Robert''; DROP TABLE users",
		},
		{
			name:     "null bytes removed",
			input:    "test\x00value",
			expected: "testvalue",
		},
		{
			name:     "combined: comments and quotes",
			input:    "SELECT * FROM users WHERE name = 'admin' -- bypass",
			expected: "SELECT * FROM users WHERE name = ''admin'' ",
		},
		{
			name:     "no modifications needed",
			input:    "SELECT id FROM users",
			expected: "SELECT id FROM users",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripSQLComments(tt.input)
			if result != tt.expected {
				t.Errorf("StripSQLComments(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeSQLString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single quote escape",
			input:    "O'Reilly",
			expected: "O''Reilly",
		},
		{
			name:     "null byte removal",
			input:    "test\x00value",
			expected: "testvalue",
		},
		{
			name:     "no change needed",
			input:    "normal string",
			expected: "normal string",
		},
		{
			name:     "multiple single quotes",
			input:    "it's a 'test'",
			expected: "it''s a ''test''",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeSQLString(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeSQLString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeSQLEqualsEscapeSQLString(t *testing.T) {
	inputs := []string{
		"normal",
		"O'Reilly",
		"test\x00value",
		"it's 'complex'",
		"",
	}

	for _, input := range inputs {
		sanitized := SanitizeSQL(input)
		escaped := EscapeSQLString(input)
		if sanitized != escaped {
			t.Errorf("SanitizeSQL(%q) = %q differs from EscapeSQLString(%q) = %q",
				input, sanitized, input, escaped)
		}
	}
}

func TestEscapeJS_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name:        "single quote",
			input:       "it's",
			mustContain: []string{`\'`},
		},
		{
			name:        "tab character",
			input:       "col1\tcol2",
			mustContain: []string{`\t`},
		},
		{
			name:        "carriage return",
			input:       "line1\rline2",
			mustContain: []string{`\r`},
		},
		{
			name:           "ampersand",
			input:          "a&b",
			mustContain:    []string{`\u0026`},
			mustNotContain: []string{"&"},
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "no special chars",
			input: "hello world 123",
		},
		{
			name:           "mixed special chars",
			input:          `<script>alert("xss")</script>`,
			mustContain:    []string{`\u003C`, `\u003E`, `\"`},
			mustNotContain: []string{"<", ">"},
		},
		{
			name:        "backslash not double-escaped with others",
			input:       `\n`,
			mustContain: []string{`\\n`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeJS(tt.input)

			if tt.input == "" && result != "" {
				t.Errorf("Expected empty output for empty input, got %q", result)
			}

			for _, substr := range tt.mustContain {
				if !strings.Contains(result, substr) {
					t.Errorf("Expected output to contain %q, got %q", substr, result)
				}
			}
			for _, substr := range tt.mustNotContain {
				if strings.Contains(result, substr) {
					t.Errorf("Expected output NOT to contain %q, got %q", substr, result)
				}
			}
		})
	}
}

func TestEscapeHTML_SpecificOutputs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "angle brackets",
			input:    "<script>",
			expected: "&lt;script&gt;",
		},
		{
			name:     "ampersand",
			input:    "Tom & Jerry",
			expected: "Tom &amp; Jerry",
		},
		{
			name:     "double quote",
			input:    `say "hello"`,
			expected: "say &#34;hello&#34;",
		},
		{
			name:     "single quote",
			input:    "it's",
			expected: "it&#39;s",
		},
		{
			name:     "plain text unchanged",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "combined",
			input:    `<a href="test">link & stuff</a>`,
			expected: "&lt;a href=&#34;test&#34;&gt;link &amp; stuff&lt;/a&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeHTML(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
