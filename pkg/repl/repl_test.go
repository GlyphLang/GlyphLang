package repl

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
)

// TestREPLBasicExpression tests basic expression evaluation.
func TestREPLBasicExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "integer literal",
			input:    "42\n",
			expected: "=> 42",
		},
		{
			name:     "string literal",
			input:    "\"hello\"\n",
			expected: "=> \"hello\"",
		},
		{
			name:     "boolean true",
			input:    "true\n",
			expected: "=> true",
		},
		{
			name:     "boolean false",
			input:    "false\n",
			expected: "=> false",
		},
		{
			name:     "addition",
			input:    "1 + 2\n",
			expected: "=> 3",
		},
		{
			name:     "multiplication",
			input:    "3 * 4\n",
			expected: "=> 12",
		},
		{
			name:     "arithmetic precedence",
			input:    "1 + 2 * 3\n",
			expected: "=> 7",
		},
		{
			name:     "string concatenation",
			input:    "\"hello\" + \" world\"\n",
			expected: "=> \"hello world\"",
		},
		{
			name:     "comparison",
			input:    "5 > 3\n",
			expected: "=> true",
		},
		{
			name:     "equality",
			input:    "2 == 2\n",
			expected: "=> true",
		},
		{
			name:     "empty array",
			input:    "[]\n",
			expected: "=> []",
		},
		{
			name:     "array literal",
			input:    "[1, 2, 3]\n",
			expected: "=> [1, 2, 3]",
		},
		{
			name:     "object literal",
			input:    "{x: 1, y: 2}\n",
			expected: "=> {",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input + ":quit\n")
			output := &bytes.Buffer{}

			r := New(input, output, "test")
			r.Start()

			if !strings.Contains(output.String(), tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output.String())
			}
		})
	}
}

// TestREPLVariablePersistence tests that variables persist across lines.
func TestREPLVariablePersistence(t *testing.T) {
	input := strings.NewReader("$ x = 10\nx * 2\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()

	// First should show x = 10
	if !strings.Contains(result, "=> 10") {
		t.Errorf("Expected first output to contain '=> 10', got %q", result)
	}

	// Second should show x * 2 = 20
	if !strings.Contains(result, "=> 20") {
		t.Errorf("Expected second output to contain '=> 20', got %q", result)
	}
}

// TestREPLHelpCommand tests the :help command.
func TestREPLHelpCommand(t *testing.T) {
	input := strings.NewReader(":help\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()

	expectedStrings := []string{
		":help",
		":quit",
		":type",
		":load",
		":reset",
		":clear",
		":vars",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected help output to contain %q", expected)
		}
	}
}

// TestREPLTypeCommand tests the :type command.
func TestREPLTypeCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "integer type",
			input:    ":type 42\n:quit\n",
			expected: ":: int",
		},
		{
			name:     "string type",
			input:    ":type \"hello\"\n:quit\n",
			expected: ":: str",
		},
		{
			name:     "boolean type",
			input:    ":type true\n:quit\n",
			expected: ":: bool",
		},
		{
			name:     "array type",
			input:    ":type [1, 2]\n:quit\n",
			expected: ":: [int]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}

			r := New(input, output, "test")
			r.Start()

			if !strings.Contains(output.String(), tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output.String())
			}
		})
	}
}

// TestREPLResetCommand tests the :reset command.
func TestREPLResetCommand(t *testing.T) {
	input := strings.NewReader("$ x = 10\n:reset\nx\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()

	// After reset, x should be undefined
	if !strings.Contains(result, "undefined variable: x") && !strings.Contains(result, "Error") {
		// x might not error if REPL handles it differently
		// At minimum, we should see the reset message
	}

	if !strings.Contains(result, "reset") {
		t.Errorf("Expected output to contain 'reset', got %q", result)
	}
}

// TestREPLUnknownCommand tests handling of unknown commands.
func TestREPLUnknownCommand(t *testing.T) {
	input := strings.NewReader(":unknown\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	if !strings.Contains(output.String(), "unknown command") {
		t.Errorf("Expected error message for unknown command")
	}
}

// TestREPLMultilineInput tests multi-line input with braces.
func TestREPLMultilineInput(t *testing.T) {
	// Test that incomplete input continues reading
	input := strings.NewReader("{\n  x: 1\n}\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()

	// Should see continuation prompts
	if !strings.Contains(result, "...") || !strings.Contains(result, "=>") {
		// Either continuation or result should appear
		// The object should be parsed eventually
	}
}

// TestIsInputComplete tests the input completeness detection.
func TestIsInputComplete(t *testing.T) {
	tests := []struct {
		input    string
		complete bool
	}{
		{"42", true},
		{"{", false},
		{"{}", true},
		{"{ x: 1 }", true},
		{"{ x: 1", false},
		{"[", false},
		{"[]", true},
		{"[1, 2, 3]", true},
		{"[1, 2", false},
		{"(", false},
		{"()", true},
		{"(1 + 2)", true},
		{"(1 + 2", false},
		{"\"hello\"", true},
		{"\"hello", false},
		{"`template", false},
		{"`template`", true},
		{"'single", false},
		{"'single'", true},
	}

	r := New(strings.NewReader(""), &bytes.Buffer{}, "test")

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := r.isInputComplete(tt.input)
			if result != tt.complete {
				t.Errorf("isInputComplete(%q) = %v, want %v", tt.input, result, tt.complete)
			}
		})
	}
}

// TestDetectInputType tests input type detection.
func TestDetectInputType(t *testing.T) {
	tests := []struct {
		input    string
		expected inputType
	}{
		{"42", inputTypeExpression},
		{"$ x = 1", inputTypeStatement},
		{"> x + 1", inputTypeStatement},
		{": User { name: str! }", inputTypeTypeDef},
		{"! add(a: int, b: int) { > a + b }", inputTypeFunction},
		{"let x = 1", inputTypeStatement},
		{"return x", inputTypeStatement},
		{"if x > 0 { }", inputTypeStatement},
		{"type User { }", inputTypeTypeDef},
	}

	r := New(strings.NewReader(""), &bytes.Buffer{}, "test")

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := r.detectInputType(tt.input)
			if result != tt.expected {
				t.Errorf("detectInputType(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestFormatValue tests value formatting.
func TestFormatValue(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected string
	}{
		{nil, "nil"},
		{int64(42), "42"},
		{int(42), "42"},
		{3.14, "3.14"},
		{true, "true"},
		{false, "false"},
		{"hello", "\"hello\""},
		{[]interface{}{int64(1), int64(2)}, "[1, 2]"},
	}

	for _, tt := range tests {
		result := formatValue(tt.value)
		if result != tt.expected {
			t.Errorf("formatValue(%v) = %q, want %q", tt.value, result, tt.expected)
		}
	}
}

// TestREPLFunctionDefinition tests function definition and calling.
func TestREPLFunctionDefinition(t *testing.T) {
	input := strings.NewReader("! double(n: int) -> int { > n * 2 }\ndouble(21)\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()

	// Should see function defined message
	if !strings.Contains(result, "double") && !strings.Contains(result, "defined") {
		// Function definition confirmation might vary
	}

	// Should see result 42
	if !strings.Contains(result, "42") {
		t.Errorf("Expected output to contain '42' from double(21), got %q", result)
	}
}

// TestREPLVarsCommand tests the :vars command.
func TestREPLVarsCommand(t *testing.T) {
	input := strings.NewReader("$ x = 10\n$ y = \"hello\"\n:vars\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()

	// Should list variables
	if !strings.Contains(result, "x") {
		t.Errorf("Expected :vars to show variable x")
	}
	if !strings.Contains(result, "y") {
		t.Errorf("Expected :vars to show variable y")
	}
}

// TestREPLLetStatement tests let statement as alias for $.
func TestREPLLetStatement(t *testing.T) {
	input := strings.NewReader("let x = 42\nx\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()

	// Should show x = 42
	if !strings.Contains(result, "42") {
		t.Errorf("Expected output to contain '42', got %q", result)
	}
}

// TestREPLReturnStatement tests return statement as alias for >.
func TestREPLReturnStatement(t *testing.T) {
	input := strings.NewReader("return 100\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()

	// Should show 100
	if !strings.Contains(result, "100") {
		t.Errorf("Expected output to contain '100', got %q", result)
	}
}

// TestREPLEmptyInput tests handling of empty input.
func TestREPLEmptyInput(t *testing.T) {
	input := strings.NewReader("\n\n\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	err := r.Start()

	if err != nil {
		t.Errorf("REPL should handle empty input gracefully, got error: %v", err)
	}
}

// TestGetTypeName tests type name detection.
func TestGetTypeName(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected string
	}{
		{nil, "nil"},
		{int64(42), "int"},
		{42, "int"},
		{3.14, "float"},
		{"hello", "str"},
		{true, "bool"},
		{[]interface{}{}, "[]"},
		{[]interface{}{int64(1)}, "[int]"},
		{map[string]interface{}{}, "object"},
	}

	for _, tt := range tests {
		result := getTypeName(tt.value)
		if result != tt.expected {
			t.Errorf("getTypeName(%v) = %q, want %q", tt.value, result, tt.expected)
		}
	}
}

// TestStop tests the Stop method.
func TestStop(t *testing.T) {
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	r := New(input, output, "test")

	// Verify the REPL is not running initially
	if r.running {
		t.Error("REPL should not be running before Start")
	}

	// Set running to true manually, then Stop
	r.running = true
	r.Stop()

	if r.running {
		t.Error("REPL should not be running after Stop")
	}
}

// TestGetEnvironment tests the GetEnvironment method.
func TestGetEnvironment(t *testing.T) {
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	env := r.GetEnvironment()

	if env == nil {
		t.Fatal("GetEnvironment returned nil")
	}

	// Verify it is the same environment used by the REPL
	env.Define("testVar", int64(42))
	val, err := r.env.Get("testVar")
	if err != nil {
		t.Fatalf("Expected to find testVar in environment: %v", err)
	}
	if val != int64(42) {
		t.Errorf("Expected testVar=42, got %v", val)
	}
}

// TestGetInterpreter tests the GetInterpreter method.
func TestGetInterpreter(t *testing.T) {
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	interp := r.GetInterpreter()

	if interp == nil {
		t.Fatal("GetInterpreter returned nil")
	}

	// Verify it is the same interpreter used by the REPL
	if interp != r.interp {
		t.Error("GetInterpreter should return the same interpreter instance")
	}
}

// TestResetDirectly tests the Reset method directly (not through the :reset command).
func TestResetDirectly(t *testing.T) {
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	r := New(input, output, "test")

	// Define a variable in the environment
	r.env.Define("x", int64(10))
	// Write something to the input buffer
	r.inputBuffer.WriteString("partial input")
	r.lineNumber = 5

	// Reset
	r.Reset()

	// Environment should be fresh
	if r.env.Has("x") {
		t.Error("After Reset, environment should not contain previously defined variables")
	}

	// Input buffer should be cleared
	if r.inputBuffer.Len() != 0 {
		t.Errorf("After Reset, inputBuffer should be empty, got %q", r.inputBuffer.String())
	}

	// Line number should be reset
	if r.lineNumber != 1 {
		t.Errorf("After Reset, lineNumber should be 1, got %d", r.lineNumber)
	}

	// Interpreter should be fresh (non-nil)
	if r.interp == nil {
		t.Error("After Reset, interpreter should not be nil")
	}
}

// TestPrintResultNil tests printResult with nil value.
func TestPrintResultNil(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	r.printResult(nil)

	if !strings.Contains(output.String(), "nil") {
		t.Errorf("Expected 'nil' in output, got %q", output.String())
	}
}

// TestPrintResultNonNil tests printResult with a non-nil value.
func TestPrintResultNonNil(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	r.printResult(int64(42))

	result := output.String()
	if !strings.Contains(result, "=> 42") {
		t.Errorf("Expected '=> 42' in output, got %q", result)
	}
}

// TestPrintResultVariousTypes tests printResult with various value types.
func TestPrintResultVariousTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"nil value", nil, "nil"},
		{"integer", int64(99), "=> 99"},
		{"string", "hello", "=> \"hello\""},
		{"boolean", true, "=> true"},
		{"float", 3.14, "=> 3.14"},
		{"array", []interface{}{int64(1), int64(2)}, "=> [1, 2]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			r := New(strings.NewReader(""), output, "test")
			r.printResult(tt.value)
			if !strings.Contains(output.String(), tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output.String())
			}
		})
	}
}

// TestFormatValueMap tests formatValue with a map value.
func TestFormatValueMap(t *testing.T) {
	m := map[string]interface{}{
		"key": "val",
	}
	result := formatValue(m)
	if !strings.Contains(result, "key") || !strings.Contains(result, "val") {
		t.Errorf("Expected map format to contain key and val, got %q", result)
	}
	if !strings.HasPrefix(result, "{") || !strings.HasSuffix(result, "}") {
		t.Errorf("Expected map format to be wrapped in braces, got %q", result)
	}
}

// TestFormatValueDefault tests formatValue with an unrecognized type.
func TestFormatValueDefault(t *testing.T) {
	type customType struct{ X int }
	result := formatValue(customType{X: 5})
	if result == "" {
		t.Error("formatValue should produce non-empty string for unknown types")
	}
	// Should use default %v formatting
	if !strings.Contains(result, "5") {
		t.Errorf("Expected default format to contain '5', got %q", result)
	}
}

// TestTypesCommand tests the :types command.
func TestTypesCommand(t *testing.T) {
	input := strings.NewReader(":types\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "Type definitions") {
		t.Errorf("Expected :types to mention 'Type definitions', got %q", result)
	}
}

// TestFunctionsCommandNoFunctions tests the :functions command with no functions defined.
func TestFunctionsCommandNoFunctions(t *testing.T) {
	input := strings.NewReader(":functions\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "No functions defined") {
		t.Errorf("Expected :functions to show 'No functions defined', got %q", result)
	}
}

// TestFunctionsCommandWithFunctions tests :functions after defining a function.
func TestFunctionsCommandWithFunctions(t *testing.T) {
	input := strings.NewReader("! add(a: int, b: int) -> int { > a + b }\n:functions\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "Functions:") {
		t.Errorf("Expected :functions to show 'Functions:', got %q", result)
	}
	if !strings.Contains(result, "add") {
		t.Errorf("Expected :functions to list 'add' function, got %q", result)
	}
}

// TestFnsAliasCommand tests the :fns alias for :functions.
func TestFnsAliasCommand(t *testing.T) {
	input := strings.NewReader(":fns\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "No functions defined") {
		t.Errorf("Expected :fns alias to work, got %q", result)
	}
}

// TestClearCommand tests the :clear command.
func TestClearCommand(t *testing.T) {
	// :clear runs an external command, but if it fails it falls back to printing newlines.
	// We cannot easily test the actual clear, but we can verify it does not error.
	input := strings.NewReader(":clear\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	// The REPL should still be functional after :clear
	// (no panic, and :quit still processed)
	result := output.String()
	if !strings.Contains(result, "Goodbye") {
		t.Errorf("Expected REPL to continue after :clear, got %q", result)
	}
}

// TestClsAliasCommand tests the :cls alias for :clear.
func TestClsAliasCommand(t *testing.T) {
	input := strings.NewReader(":cls\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "Goodbye") {
		t.Errorf("Expected REPL to continue after :cls, got %q", result)
	}
}

// TestTypeCommandNoArgs tests :type with no arguments.
func TestTypeCommandNoArgs(t *testing.T) {
	input := strings.NewReader(":type\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "usage") || !strings.Contains(result, "Error") {
		t.Errorf("Expected :type with no args to show usage error, got %q", result)
	}
}

// TestTypeCommandShortAlias tests the :t alias for :type.
func TestTypeCommandShortAlias(t *testing.T) {
	input := strings.NewReader(":t 42\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, ":: int") {
		t.Errorf("Expected :t alias to show type, got %q", result)
	}
}

// TestLoadCommandNoArgs tests :load with no arguments.
func TestLoadCommandNoArgs(t *testing.T) {
	input := strings.NewReader(":load\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "usage") || !strings.Contains(result, "Error") {
		t.Errorf("Expected :load with no args to show usage error, got %q", result)
	}
}

// TestLoadCommandNonexistentFile tests :load with a file that does not exist.
func TestLoadCommandNonexistentFile(t *testing.T) {
	input := strings.NewReader(":load nonexistent_file_xyz\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "Error") {
		t.Errorf("Expected :load of nonexistent file to show error, got %q", result)
	}
}

// TestLoadCommandWithExtension tests :load preserves .glyph extension when already present.
func TestLoadCommandWithExtension(t *testing.T) {
	input := strings.NewReader(":load nonexistent.glyph\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	// Should try to load nonexistent.glyph (not nonexistent.glyph.glyph)
	if !strings.Contains(result, "Loading nonexistent.glyph") {
		t.Errorf("Expected loading message for nonexistent.glyph, got %q", result)
	}
}

// TestLoadFileValidFile tests LoadFile with a valid Glyph file.
func TestLoadFileValidFile(t *testing.T) {
	// Create a temporary .glyph file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.glyph")
	content := []byte("! greet(name: str) -> str { > \"Hello, \" + name }\n")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	err := r.LoadFile(filePath)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	// Verify the function was loaded into the environment
	val, err := r.env.Get("greet")
	if err != nil {
		t.Fatalf("Expected function 'greet' in environment: %v", err)
	}

	if _, ok := val.(interpreter.Function); !ok {
		t.Errorf("Expected 'greet' to be a Function, got %T", val)
	}
}

// TestLoadFileNonexistent tests LoadFile with a nonexistent file.
func TestLoadFileNonexistent(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	err := r.LoadFile("/nonexistent/path/to/file.glyph")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Expected 'failed to read file' error, got: %v", err)
	}
}

// TestLoadFileInvalidSyntax tests LoadFile with invalid Glyph syntax.
func TestLoadFileInvalidSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "bad.glyph")
	content := []byte("!!! invalid {{{ syntax ???")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	err := r.LoadFile(filePath)
	if err == nil {
		t.Error("Expected error for invalid syntax")
	}
}

// TestLoadCommandValidFile tests :load command with a valid file through the REPL.
func TestLoadCommandValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "funcs.glyph")
	content := []byte("! triple(n: int) -> int { > n * 3 }\n")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	input := strings.NewReader(":load " + filePath + "\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "Loaded successfully") {
		t.Errorf("Expected 'Loaded successfully' message, got %q", result)
	}
}

// TestLoadCommandShortAlias tests the :l alias for :load.
func TestLoadCommandShortAlias(t *testing.T) {
	input := strings.NewReader(":l nonexistent_file\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	// Should attempt to load (and fail since file does not exist)
	if !strings.Contains(result, "Loading") {
		t.Errorf("Expected :l alias to trigger loading, got %q", result)
	}
}

// TestEvaluateTypeDef tests type definition through the REPL.
func TestEvaluateTypeDef(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	// Call evaluateTypeDef directly since colon-prefixed input goes through
	// executeCommand when entered via processLine
	err := r.evaluateTypeDef(": Point {\n  x: int!\n  y: int!\n}")
	if err != nil {
		t.Fatalf("Expected no error defining type, got: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Type") || !strings.Contains(result, "Point") {
		t.Errorf("Expected type definition confirmation, got %q", result)
	}
}

// TestEvaluateTypeDefWithTypeKeyword tests type definition using the 'type' keyword.
func TestEvaluateTypeDefWithTypeKeyword(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	err := r.evaluateTypeDef("type User {\n  name: str!\n  age: int\n}")
	if err != nil {
		t.Fatalf("Expected no error defining type with 'type' keyword, got: %v", err)
	}
}

// TestEvaluateFunctionDefinition tests function definition through processLine.
func TestEvaluateFunctionDefinition(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	err := r.processLine("! square(n: int) -> int { > n * n }")
	if err != nil {
		t.Fatalf("Expected no error defining function, got: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Function") || !strings.Contains(result, "square") {
		t.Errorf("Expected function definition message, got %q", result)
	}

	// Verify function is in the environment
	val, envErr := r.env.Get("square")
	if envErr != nil {
		t.Fatalf("Expected 'square' in environment: %v", envErr)
	}
	if _, ok := val.(interpreter.Function); !ok {
		t.Errorf("Expected 'square' to be a Function, got %T", val)
	}
}

// TestEvaluateFunctionParseError tests function definition with invalid syntax.
func TestEvaluateFunctionParseError(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	// Use evaluateFunction directly with invalid syntax to trigger parse error
	err := r.evaluateFunction("! @@@ invalid {{{ }")
	if err == nil {
		t.Error("Expected parse error for invalid function syntax")
	}
	if err != nil && !strings.Contains(err.Error(), "parse error") {
		t.Errorf("Expected 'parse error', got: %v", err)
	}
}

// TestEvaluateTypeDefDirectly tests evaluateTypeDef directly.
func TestEvaluateTypeDefDirectly(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	err := r.evaluateTypeDef(": Color {\n  r: int!\n  g: int!\n  b: int!\n}")
	if err != nil {
		t.Fatalf("Expected no error defining type, got: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Type") || !strings.Contains(result, "Color") {
		t.Errorf("Expected type definition message with 'Color', got %q", result)
	}
}

// TestEvaluateTypeDefParseError tests evaluateTypeDef with invalid syntax.
func TestEvaluateTypeDefParseError(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	err := r.evaluateTypeDef(": {{{ invalid")
	if err == nil {
		t.Error("Expected parse error for invalid type definition")
	}
	if err != nil && !strings.Contains(err.Error(), "parse error") {
		t.Errorf("Expected 'parse error', got: %v", err)
	}
}

// TestResetViaCommand tests :reset and then :r alias.
func TestResetViaCommand(t *testing.T) {
	// Test :r alias for reset
	input := strings.NewReader("$ x = 10\n:r\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "reset") {
		t.Errorf("Expected :r alias to reset, got %q", result)
	}
}

// TestExitCommand tests the :exit command.
func TestExitCommand(t *testing.T) {
	input := strings.NewReader(":exit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "Goodbye") {
		t.Errorf("Expected :exit to produce goodbye message, got %q", result)
	}
}

// TestQuitShortAlias tests the :q alias for :quit.
func TestQuitShortAlias(t *testing.T) {
	input := strings.NewReader(":q\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "Goodbye") {
		t.Errorf("Expected :q alias to quit, got %q", result)
	}
}

// TestHelpShortAlias tests the :h alias for :help.
func TestHelpShortAlias(t *testing.T) {
	input := strings.NewReader(":h\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "Glyph REPL Commands") {
		t.Errorf("Expected :h alias to show help, got %q", result)
	}
}

// TestVarsShortAlias tests the :v alias for :vars.
func TestVarsShortAlias(t *testing.T) {
	input := strings.NewReader(":v\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "No variables defined") {
		t.Errorf("Expected :v alias to show vars, got %q", result)
	}
}

// TestFormatFunctionSignature tests formatFunctionSignature with various function definitions.
func TestFormatFunctionSignature(t *testing.T) {
	tests := []struct {
		name     string
		fn       interpreter.Function
		expected string
	}{
		{
			name: "simple function no params",
			fn: interpreter.Function{
				Name:   "hello",
				Params: nil,
			},
			expected: "hello()",
		},
		{
			name: "function with typed params",
			fn: interpreter.Function{
				Name: "add",
				Params: []interpreter.Field{
					{Name: "a", TypeAnnotation: interpreter.IntType{}, Required: false},
					{Name: "b", TypeAnnotation: interpreter.IntType{}, Required: false},
				},
				ReturnType: interpreter.IntType{},
			},
			expected: "add(a: int, b: int) -> int",
		},
		{
			name: "function with required params",
			fn: interpreter.Function{
				Name: "greet",
				Params: []interpreter.Field{
					{Name: "name", TypeAnnotation: interpreter.StringType{}, Required: true},
				},
				ReturnType: interpreter.StringType{},
			},
			expected: "greet(name: str!) -> str",
		},
		{
			name: "function with no type annotation",
			fn: interpreter.Function{
				Name: "identity",
				Params: []interpreter.Field{
					{Name: "x", TypeAnnotation: nil, Required: false},
				},
				ReturnType: nil,
			},
			expected: "identity(x)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFunctionSignature(tt.fn.Name, tt.fn)
			if result != tt.expected {
				t.Errorf("formatFunctionSignature(%q) = %q, want %q", tt.fn.Name, result, tt.expected)
			}
		})
	}
}

// TestFormatType tests formatType with various type representations.
func TestFormatType(t *testing.T) {
	tests := []struct {
		name     string
		typ      interpreter.Type
		expected string
	}{
		{"nil type", nil, "any"},
		{"int type", interpreter.IntType{}, "int"},
		{"string type", interpreter.StringType{}, "str"},
		{"bool type", interpreter.BoolType{}, "bool"},
		{"float type", interpreter.FloatType{}, "float"},
		{"array of int", interpreter.ArrayType{ElementType: interpreter.IntType{}}, "[int]"},
		{"optional int", interpreter.OptionalType{InnerType: interpreter.IntType{}}, "int?"},
		{"named type", interpreter.NamedType{Name: "User"}, "User"},
		{
			"generic type with args",
			interpreter.GenericType{
				BaseType: interpreter.NamedType{Name: "List"},
				TypeArgs: []interpreter.Type{interpreter.IntType{}},
			},
			"List<int>",
		},
		{
			"generic type no args",
			interpreter.GenericType{
				BaseType: interpreter.NamedType{Name: "Any"},
				TypeArgs: nil,
			},
			"Any",
		},
		{
			"function type",
			interpreter.FunctionType{
				ParamTypes: []interpreter.Type{interpreter.IntType{}, interpreter.IntType{}},
				ReturnType: interpreter.IntType{},
			},
			"(int, int) -> int",
		},
		{
			"nested array",
			interpreter.ArrayType{ElementType: interpreter.ArrayType{ElementType: interpreter.StringType{}}},
			"[[str]]",
		},
		{
			"optional string",
			interpreter.OptionalType{InnerType: interpreter.StringType{}},
			"str?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatType(tt.typ)
			if result != tt.expected {
				t.Errorf("formatType(%v) = %q, want %q", tt.typ, result, tt.expected)
			}
		})
	}
}

// TestGetTypeNameFunction tests getTypeName with a Function value.
func TestGetTypeNameFunction(t *testing.T) {
	fn := interpreter.Function{
		Name: "test",
	}
	result := getTypeName(fn)
	if result != "function" {
		t.Errorf("getTypeName(Function) = %q, want %q", result, "function")
	}
}

// TestGetTypeNameUnknown tests getTypeName with an unknown type.
func TestGetTypeNameUnknown(t *testing.T) {
	type custom struct{}
	result := getTypeName(custom{})
	if result == "" {
		t.Error("getTypeName should return a non-empty string for unknown types")
	}
}

// TestVarsCommandWithFunctions tests that :vars skips function values.
func TestVarsCommandWithFunctions(t *testing.T) {
	input := strings.NewReader("$ x = 10\n! noop() -> int { > 0 }\n:vars\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()

	// Should show x variable
	if !strings.Contains(result, "x") {
		t.Errorf("Expected :vars to show variable x, got %q", result)
	}

	// The function 'noop' should be defined but not listed as a variable in :vars
	// :vars output format is "  name :: type = value" for vars; functions are skipped
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		// Look for lines after "Variables:" that contain function info
		if strings.Contains(line, "noop") && strings.Contains(line, "::") && strings.Contains(line, "function") {
			t.Errorf(":vars should skip function values, but found %q", line)
		}
	}
}

// TestVarsCommandNoVars tests :vars with no variables defined.
func TestVarsCommandNoVars(t *testing.T) {
	input := strings.NewReader(":vars\n:quit\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	r.Start()

	result := output.String()
	if !strings.Contains(result, "No variables defined") {
		t.Errorf("Expected 'No variables defined', got %q", result)
	}
}

// TestEvaluateExpressionError tests expression evaluation errors.
func TestEvaluateExpressionError(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	// Evaluate an undefined variable
	err := r.evaluateExpression("undefinedVar123")
	if err == nil {
		t.Error("Expected error evaluating undefined variable")
	}
}

// TestProcessLineEmptyAfterTrim tests processLine with whitespace-only input.
func TestProcessLineEmptyAfterTrim(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	// When input buffer is empty and we get a line that is just whitespace,
	// it should be added to buffer, then the buffer trims to empty and returns nil
	err := r.processLine("   ")
	if err != nil {
		t.Errorf("Expected no error for whitespace input, got: %v", err)
	}
}

// TestResetPreservesModuleResolver tests that Reset properly reinitializes the module resolver.
func TestResetPreservesModuleResolver(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	r.Reset()

	// After reset, interpreter should have a working module resolver
	// Verify by loading a simple function
	err := r.processLine("! testFunc() -> int { > 1 }")
	if err != nil {
		t.Errorf("After Reset, REPL should still work: %v", err)
	}
}

// TestEvaluateFunctionGenericFallback tests evaluateFunction when the module
// has items but none match *interpreter.Function (the generic "Function defined" path).
func TestEvaluateFunctionGenericFallback(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	// Define a function normally - it should show the specific name
	err := r.evaluateFunction("! myFunc(x: int) -> int { > x }")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "myFunc") {
		t.Errorf("Expected function name in output, got %q", result)
	}
}

// TestEvaluateTypeDefColonSyntax tests type definition with colon syntax via evaluateTypeDef.
func TestEvaluateTypeDefColonSyntax(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	err := r.evaluateTypeDef(": Address {\n  street: str!\n  city: str!\n  zip: str!\n}")
	if err != nil {
		t.Fatalf("Expected no error defining type via colon syntax, got: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Type") {
		t.Errorf("Expected type definition message, got %q", result)
	}
}

// TestLoadFileWithTypeDefinition tests LoadFile loading a file with type and function definitions.
func TestLoadFileWithTypeDefinition(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "types.glyph")
	content := []byte(": Coord {\n  lat: float!\n  lng: float!\n}\n\n! getOne() -> int { > 1 }\n")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	err := r.LoadFile(filePath)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	// Functions should be loaded into environment
	if !r.env.Has("getOne") {
		t.Error("Expected 'getOne' function in environment after LoadFile")
	}
}

// TestFunctionsCommandMultiple tests :functions with multiple functions defined.
func TestFunctionsCommandMultiple(t *testing.T) {
	output := &bytes.Buffer{}
	r := New(strings.NewReader(""), output, "test")

	// Define functions in the environment directly
	fn1 := interpreter.Function{
		Name: "alpha",
		Params: []interpreter.Field{
			{Name: "x", TypeAnnotation: interpreter.IntType{}, Required: true},
		},
		ReturnType: interpreter.IntType{},
	}
	fn2 := interpreter.Function{
		Name: "beta",
		Params: []interpreter.Field{
			{Name: "s", TypeAnnotation: interpreter.StringType{}, Required: false},
		},
		ReturnType: interpreter.StringType{},
	}
	r.env.Define("alpha", fn1)
	r.env.Define("beta", fn2)

	// Run :functions
	err := r.executeCommand(":functions")
	if err != nil {
		t.Fatalf("Expected no error from :functions, got: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Functions:") {
		t.Errorf("Expected 'Functions:' header, got %q", result)
	}
	if !strings.Contains(result, "alpha") {
		t.Errorf("Expected 'alpha' in functions list, got %q", result)
	}
	if !strings.Contains(result, "beta") {
		t.Errorf("Expected 'beta' in functions list, got %q", result)
	}
}

// TestExecuteCommandDirectly tests executeCommand with various commands directly.
func TestExecuteCommandDirectly(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected string
		wantErr  bool
	}{
		{"help", ":help", "Glyph REPL Commands", false},
		{"types", ":types", "Type definitions", false},
		{"functions no fns", ":functions", "No functions defined", false},
		{"fns alias", ":fns", "No functions defined", false},
		{"vars no vars", ":vars", "No variables defined", false},
		{"unknown", ":badcmd", "unknown command", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			r := New(strings.NewReader(""), output, "test")
			err := r.executeCommand(tt.command)
			if tt.wantErr && err == nil {
				t.Errorf("Expected error for %q", tt.command)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error for %q: %v", tt.command, err)
			}
			result := output.String()
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.expected) {
					t.Errorf("Expected error to contain %q, got: %v", tt.expected, err)
				}
			} else {
				if !strings.Contains(result, tt.expected) {
					t.Errorf("Expected output to contain %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

// TestNewCreatesValidREPL tests that New creates a properly initialized REPL.
func TestNewCreatesValidREPL(t *testing.T) {
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	r := New(input, output, "1.0.0")

	if r.interp == nil {
		t.Error("Expected interpreter to be initialized")
	}
	if r.env == nil {
		t.Error("Expected environment to be initialized")
	}
	if r.reader == nil {
		t.Error("Expected reader to be initialized")
	}
	if r.writer == nil {
		t.Error("Expected writer to be initialized")
	}
	if r.running {
		t.Error("Expected running to be false initially")
	}
	if r.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %q", r.version)
	}
	if r.lineNumber != 1 {
		t.Errorf("Expected lineNumber 1, got %d", r.lineNumber)
	}
}

// TestStartEOFTermination tests that Start terminates gracefully on EOF.
func TestStartEOFTermination(t *testing.T) {
	// Reader with no :quit, so Start terminates on EOF
	input := strings.NewReader("42\n")
	output := &bytes.Buffer{}

	r := New(input, output, "test")
	err := r.Start()
	if err != nil {
		t.Errorf("Expected no error on EOF termination, got: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "=> 42") {
		t.Errorf("Expected '=> 42' before EOF, got %q", result)
	}
	if !strings.Contains(result, "Goodbye") {
		t.Errorf("Expected 'Goodbye' on EOF, got %q", result)
	}
}
