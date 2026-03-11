package parser

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"testing"
)

func TestParser_MacroDef(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		macroName  string
		paramCount int
		bodyCount  int
	}{
		{
			name:       "simple macro with one param",
			input:      `macro! log(msg) { > msg }`,
			macroName:  "log",
			paramCount: 1,
			bodyCount:  1,
		},
		{
			name:       "macro with multiple params",
			input:      `macro! add(a, b) { > a + b }`,
			macroName:  "add",
			paramCount: 2,
			bodyCount:  1,
		},
		{
			name:       "macro with no params",
			input:      `macro! hello() { > "Hello" }`,
			macroName:  "hello",
			paramCount: 0,
			bodyCount:  1,
		},
		{
			name: "macro with if statement",
			input: `macro! check(cond, val) {
				if cond {
					> val
				}
			}`,
			macroName:  "check",
			paramCount: 2,
			bodyCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("Lexer error: %v", err)
			}

			parser := NewParser(tokens)
			module, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parser error: %v", err)
			}

			if len(module.Items) != 1 {
				t.Fatalf("Expected 1 item, got %d", len(module.Items))
			}

			macro, ok := module.Items[0].(*ast.MacroDef)
			if !ok {
				t.Fatalf("Expected MacroDef, got %T", module.Items[0])
			}

			if macro.Name != tt.macroName {
				t.Errorf("Expected macro name '%s', got '%s'", tt.macroName, macro.Name)
			}

			if len(macro.Params) != tt.paramCount {
				t.Errorf("Expected %d params, got %d", tt.paramCount, len(macro.Params))
			}

			if len(macro.Body) != tt.bodyCount {
				t.Errorf("Expected %d body nodes, got %d", tt.bodyCount, len(macro.Body))
			}
		})
	}
}

func TestParser_MacroInvocation(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		macroName string
		argCount  int
	}{
		{
			name:      "simple invocation with string arg",
			input:     `log!("Hello")`,
			macroName: "log",
			argCount:  1,
		},
		{
			name:      "invocation with multiple args",
			input:     `add!(1, 2)`,
			macroName: "add",
			argCount:  2,
		},
		{
			name:      "invocation with no args",
			input:     `hello!()`,
			macroName: "hello",
			argCount:  0,
		},
		{
			name:      "invocation with expression args",
			input:     `check!(x > 0, x * 2)`,
			macroName: "check",
			argCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("Lexer error: %v", err)
			}

			parser := NewParser(tokens)
			module, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parser error: %v", err)
			}

			if len(module.Items) != 1 {
				t.Fatalf("Expected 1 item, got %d", len(module.Items))
			}

			inv, ok := module.Items[0].(*ast.MacroInvocation)
			if !ok {
				t.Fatalf("Expected MacroInvocation, got %T", module.Items[0])
			}

			if inv.Name != tt.macroName {
				t.Errorf("Expected macro name '%s', got '%s'", tt.macroName, inv.Name)
			}

			if len(inv.Args) != tt.argCount {
				t.Errorf("Expected %d args, got %d", tt.argCount, len(inv.Args))
			}
		})
	}
}

func TestParser_MacroWithRoute(t *testing.T) {
	input := `macro! crud(resource) {
		@ GET /resource {
			> "get " + resource
		}
	}`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}

	if len(module.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(module.Items))
	}

	macro, ok := module.Items[0].(*ast.MacroDef)
	if !ok {
		t.Fatalf("Expected MacroDef, got %T", module.Items[0])
	}

	if macro.Name != "crud" {
		t.Errorf("Expected macro name 'crud', got '%s'", macro.Name)
	}

	if len(macro.Params) != 1 {
		t.Errorf("Expected 1 param, got %d", len(macro.Params))
	}

	if macro.Params[0] != "resource" {
		t.Errorf("Expected param 'resource', got '%s'", macro.Params[0])
	}

	// Check that body contains a route
	if len(macro.Body) != 1 {
		t.Fatalf("Expected 1 body node, got %d", len(macro.Body))
	}

	_, ok = macro.Body[0].(*ast.Route)
	if !ok {
		t.Errorf("Expected Route in body, got %T", macro.Body[0])
	}
}

func TestParser_MacroDefAndInvocation(t *testing.T) {
	input := `
macro! double(x) {
	> x + x
}

double!(5)
`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}

	if len(module.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(module.Items))
	}

	// First item should be macro def
	_, ok := module.Items[0].(*ast.MacroDef)
	if !ok {
		t.Errorf("Expected first item to be MacroDef, got %T", module.Items[0])
	}

	// Second item should be macro invocation
	_, ok = module.Items[1].(*ast.MacroInvocation)
	if !ok {
		t.Errorf("Expected second item to be MacroInvocation, got %T", module.Items[1])
	}
}

func TestParser_MacroErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "macro without name",
			input: `macro! () { > 1 }`,
		},
		{
			name:  "macro without bang",
			input: `macro log(x) { > x }`,
		},
		{
			name:  "macro without body",
			input: `macro! log(x)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				return // Lexer error is acceptable
			}

			parser := NewParser(tokens)
			_, err = parser.Parse()
			if err == nil {
				t.Error("Expected parser error, but got none")
			}
		})
	}
}

func TestLexer_MacroTokens(t *testing.T) {
	input := `macro! quote`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	// Expected: MACRO, BANG, QUOTE, EOF
	expectedTypes := []TokenType{MACRO, BANG, QUOTE, EOF}

	if len(tokens) != len(expectedTypes) {
		t.Fatalf("Expected %d tokens, got %d", len(expectedTypes), len(tokens))
	}

	for i, expected := range expectedTypes {
		if tokens[i].Type != expected {
			t.Errorf("Token %d: expected %s, got %s", i, expected, tokens[i].Type)
		}
	}
}
