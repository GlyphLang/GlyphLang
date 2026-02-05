package parser

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParser_CLICommand_MultipleFlagDefaults tests CLI commands with multiple flag
// parameters that have default values. This was previously a limitation where the
// parser would incorrectly interpret the `-` in `--flag` as a binary subtraction
// operator after parsing a default value expression.
//
// The fix is in parseCommandDefaultExpr which uses a specialized binary operator
// parser that recognizes `--` or `-` followed by an identifier as the start of a
// new flag parameter rather than a subtraction operation.
func TestParser_CLICommand_MultipleFlagDefaults(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectError   bool
		expectedFlags []string
	}{
		{
			name: "single_flag_with_string_default",
			source: `! start_server --host: str = "localhost" {
  > host
}`,
			expectError:   false,
			expectedFlags: []string{"host"},
		},
		{
			name: "two_flags_with_string_defaults",
			source: `! start_server --host: str = "localhost" --env: str = "dev" {
  > host
}`,
			expectError:   false,
			expectedFlags: []string{"host", "env"},
		},
		{
			name: "two_flags_string_and_int_defaults",
			source: `! start_server --host: str = "localhost" --port: int = 8080 {
  > host
}`,
			expectError:   false,
			expectedFlags: []string{"host", "port"},
		},
		{
			name: "three_flags_with_defaults",
			source: `! server --host: str = "localhost" --port: int = 8080 --debug: bool = false {
  > host
}`,
			expectError:   false,
			expectedFlags: []string{"host", "port", "debug"},
		},
		{
			name: "flag_with_negative_number_default",
			source: `! config --offset: int = -10 {
  > offset
}`,
			expectError:   false,
			expectedFlags: []string{"offset"},
		},
		{
			name: "multiple_flags_with_negative_default",
			source: `! config --offset: int = -10 --limit: int = 100 {
  > offset
}`,
			expectError:   false,
			expectedFlags: []string{"offset", "limit"},
		},
		{
			name: "positional_and_multiple_flags",
			source: `! deploy env: str! --host: str = "localhost" --port: int = 8080 {
  > env
}`,
			expectError:   false,
			expectedFlags: []string{"host", "port"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.source)
			tokens, err := lexer.Tokenize()
			require.NoError(t, err, "Lexer should not error")

			parser := NewParser(tokens)
			module, parseErr := parser.Parse()

			if tt.expectError {
				assert.Error(t, parseErr, "Expected parser error")
				return
			}

			require.NoError(t, parseErr, "Parser should not error")
			require.Len(t, module.Items, 1, "Should have 1 item")

			cmd, ok := module.Items[0].(*ast.Command)
			require.True(t, ok, "Expected Command, got %T", module.Items[0])

			// Check flags
			var flags []string
			for _, param := range cmd.Params {
				if param.IsFlag {
					flags = append(flags, param.Name)
				}
			}

			assert.Equal(t, tt.expectedFlags, flags, "Flag names should match")
		})
	}
}
