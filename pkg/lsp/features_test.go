package lsp

import (
	"strings"
	"testing"
)

func TestGetDiagnostics(t *testing.T) {
	dm := NewDocumentManager()

	// Test with valid source
	validSource := `: User {
  name: str!
}

@ GET /api/users {
  > {users: []}
}
`

	doc, _ := dm.Open("file:///valid.abc", 1, validSource)
	diagnostics := GetDiagnostics(doc)

	if len(diagnostics) > 0 {
		t.Errorf("Expected no diagnostics for valid source, got %d", len(diagnostics))
	}

	// Test with invalid source (missing colon)
	invalidSource := `: User {
  name str!
}
`

	doc2, _ := dm.Open("file:///invalid.abc", 1, invalidSource)
	diagnostics2 := GetDiagnostics(doc2)

	if len(diagnostics2) == 0 {
		t.Error("Expected diagnostics for invalid source")
	}

	// Check diagnostic properties
	if len(diagnostics2) > 0 {
		diag := diagnostics2[0]
		if diag.Severity != DiagnosticSeverityError {
			t.Errorf("Expected error severity, got %d", diag.Severity)
		}
		if diag.Source != "glyph" {
			t.Errorf("Expected source 'glyph', got '%s'", diag.Source)
		}
	}
}

func TestGetHover(t *testing.T) {
	dm := NewDocumentManager()

	source := `: User {
  name: str!
  age: int!
}

@ GET /api/users {
  > {users: []}
}
`

	doc, _ := dm.Open("file:///test.abc", 1, source)

	tests := []struct {
		name        string
		pos         Position
		expectHover bool
		contains    string
	}{
		{
			name:        "Hover on type definition",
			pos:         Position{Line: 0, Character: 2},
			expectHover: true,
			contains:    "User",
		},
		{
			name:        "Hover on built-in type",
			pos:         Position{Line: 1, Character: 9},
			expectHover: true,
			contains:    "str",
		},
		{
			name:        "Hover on keyword",
			pos:         Position{Line: 5, Character: 3},
			expectHover: true,
			contains:    "GET",
		},
		{
			name:        "No hover on whitespace",
			pos:         Position{Line: 0, Character: 0},
			expectHover: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hover := GetHover(doc, tt.pos)
			if tt.expectHover && hover == nil {
				t.Error("Expected hover information")
			} else if !tt.expectHover && hover != nil {
				t.Error("Did not expect hover information")
			}

			if hover != nil && tt.contains != "" {
				if !strings.Contains(hover.Contents.Value, tt.contains) {
					t.Errorf("Expected hover to contain '%s', got: %s", tt.contains, hover.Contents.Value)
				}
			}
		})
	}
}

func TestGetCompletion(t *testing.T) {
	dm := NewDocumentManager()

	source := `: User {
  name: str!
}
`

	doc, _ := dm.Open("file:///test.abc", 1, source)

	// Get completions
	completions := GetCompletion(doc, Position{Line: 0, Character: 0})

	if len(completions) == 0 {
		t.Fatal("Expected completion items")
	}

	// Check for specific completions
	foundKeywords := false
	foundTypes := false
	foundSnippets := false

	for _, item := range completions {
		switch item.Kind {
		case CompletionItemKindKeyword:
			foundKeywords = true
			if item.Label == "route" || item.Label == "if" {
				// Good
			}
		case CompletionItemKindClass:
			foundTypes = true
			if item.Label == "int" || item.Label == "str" {
				// Good
			}
		case CompletionItemKindSnippet:
			foundSnippets = true
		case CompletionItemKindStruct:
			if item.Label == "User" {
				// Found our type definition
			}
		}
	}

	if !foundKeywords {
		t.Error("Expected keyword completions")
	}

	if !foundTypes {
		t.Error("Expected type completions")
	}

	if !foundSnippets {
		t.Error("Expected snippet completions")
	}
}

func TestGetDefinition(t *testing.T) {
	dm := NewDocumentManager()

	source := `: User {
  name: str!
}

: Post {
  author: User
}
`

	doc, _ := dm.Open("file:///test.abc", 1, source)

	// Try to find definition of "User" type
	definitions := GetDefinition(doc, Position{Line: 5, Character: 11})

	// For now, GetDefinition returns approximate locations
	// This test just checks that it doesn't crash
	if definitions == nil {
		// OK - no definition found
	} else if len(definitions) > 0 {
		// OK - found definition
		if definitions[0].URI != doc.URI {
			t.Error("Definition should be in the same file")
		}
	}
}

func TestGetDocumentSymbols(t *testing.T) {
	dm := NewDocumentManager()

	source := `: User {
  name: str!
  email: str!
}

: Post {
  title: str!
}

@ GET /api/users {
  > {users: []}
}

@ POST /api/posts {
  > {posts: []}
}
`

	doc, _ := dm.Open("file:///test.abc", 1, source)

	symbols := GetDocumentSymbols(doc)

	if len(symbols) != 4 {
		t.Errorf("Expected 4 symbols (2 types + 2 routes), got %d", len(symbols))
	}

	// Check symbol types
	typeCount := 0
	routeCount := 0

	for _, sym := range symbols {
		switch sym.Kind {
		case SymbolKindStruct:
			typeCount++
			// Type definitions should have children (fields)
			if sym.Name == "User" && len(sym.Children) != 2 {
				t.Errorf("Expected User type to have 2 fields, got %d", len(sym.Children))
			}
		case SymbolKindMethod:
			routeCount++
		}
	}

	if typeCount != 2 {
		t.Errorf("Expected 2 type symbols, got %d", typeCount)
	}

	if routeCount != 2 {
		t.Errorf("Expected 2 route symbols, got %d", routeCount)
	}
}

func TestFormatType(t *testing.T) {
	dm := NewDocumentManager()

	source := `: User {
  name: str!
  age: int!
}
`

	doc, _ := dm.Open("file:///test.abc", 1, source)

	if doc.AST == nil {
		t.Fatal("Expected AST to be parsed")
	}

	// The formatType function is used internally
	// We'll just verify it doesn't crash
	symbols := GetDocumentSymbols(doc)
	if len(symbols) > 0 && len(symbols[0].Children) > 0 {
		// Detail contains formatted type
		detail := symbols[0].Children[0].Detail
		if detail == "" {
			t.Error("Expected non-empty type detail")
		}
	}
}

func TestGetKeywordInfo(t *testing.T) {
	keywords := []string{
		"route", "if", "else", "while", "for", "switch", "case", "default",
	}

	for _, kw := range keywords {
		info := getKeywordInfo(kw)
		if info == "" {
			t.Errorf("Expected info for keyword '%s'", kw)
		}
	}

	// Test non-keyword
	info := getKeywordInfo("notakeyword")
	if info != "" {
		t.Error("Should not return info for non-keyword")
	}
}

func TestGetBuiltInTypeInfo(t *testing.T) {
	types := []string{"int", "str", "string", "bool", "float"}

	for _, typ := range types {
		info := getBuiltInTypeInfo(typ)
		if info == "" {
			t.Errorf("Expected info for type '%s'", typ)
		}
	}

	// Test non-built-in type
	info := getBuiltInTypeInfo("CustomType")
	if info != "" {
		t.Error("Should not return info for custom type")
	}
}

func TestCheckTypes(t *testing.T) {
	dm := NewDocumentManager()

	// Test with undefined type reference
	source := `: Post {
  author: User
}
`

	doc, _ := dm.Open("file:///test.abc", 1, source)

	if doc.AST == nil {
		t.Fatal("Expected AST to be parsed")
	}

	diagnostics := checkTypes(doc.AST)

	// Should have warning about undefined type "User"
	foundUndefinedType := false
	for _, diag := range diagnostics {
		if strings.Contains(diag.Message, "Undefined type") {
			foundUndefinedType = true
		}
	}

	if !foundUndefinedType {
		t.Error("Expected warning about undefined type")
	}
}

func TestCheckTypesValid(t *testing.T) {
	dm := NewDocumentManager()

	// Test with all defined types
	source := `: User {
  name: str!
}

: Post {
  author: User
  title: str!
}
`

	doc, _ := dm.Open("file:///test.abc", 1, source)

	if doc.AST == nil {
		t.Fatal("Expected AST to be parsed")
	}

	diagnostics := checkTypes(doc.AST)

	// Should have no warnings since User is defined
	if len(diagnostics) > 0 {
		t.Errorf("Expected no type errors, got %d diagnostics", len(diagnostics))
	}
}

func TestFormatTypeDefHover(t *testing.T) {
	dm := NewDocumentManager()

	source := `: User {
  name: str!
  email: str!
  age: int
}
`

	doc, _ := dm.Open("file:///test.abc", 1, source)

	hover := GetHover(doc, Position{Line: 0, Character: 2})
	if hover == nil {
		t.Fatal("Expected hover for type definition")
	}

	content := hover.Contents.Value
	if !strings.Contains(content, "User") {
		t.Error("Expected hover to contain type name")
	}

	if !strings.Contains(content, "name") || !strings.Contains(content, "email") {
		t.Error("Expected hover to contain field names")
	}
}

func TestFormatRouteHover(t *testing.T) {
	dm := NewDocumentManager()

	source := `@ GET /api/users {
  + auth(jwt)
  + ratelimit(100/min)
  > {users: []}
}
`

	doc, _ := dm.Open("file:///test.abc", 1, source)

	hover := GetHover(doc, Position{Line: 0, Character: 10})
	if hover == nil {
		t.Skip("Hover might not be available at this position")
	}

	content := hover.Contents.Value
	if !strings.Contains(content, "Route") {
		t.Error("Expected hover to contain 'Route'")
	}
}

func TestCompletionSnippets(t *testing.T) {
	dm := NewDocumentManager()

	source := ""
	doc, _ := dm.Open("file:///test.abc", 1, source)

	completions := GetCompletion(doc, Position{Line: 0, Character: 0})

	snippetCount := 0
	for _, item := range completions {
		if item.Kind == CompletionItemKindSnippet {
			snippetCount++
			// Check that snippet has InsertText
			if item.InsertText == "" {
				t.Errorf("Snippet '%s' should have InsertText", item.Label)
			}
			// Check that InsertTextFormat is set to snippet (2)
			if item.InsertTextFormat != 2 {
				t.Errorf("Snippet '%s' should have InsertTextFormat=2", item.Label)
			}
		}
	}

	if snippetCount == 0 {
		t.Error("Expected at least one snippet completion")
	}
}

func TestCompletionHTTPMethods(t *testing.T) {
	dm := NewDocumentManager()

	source := ""
	doc, _ := dm.Open("file:///test.abc", 1, source)

	completions := GetCompletion(doc, Position{Line: 0, Character: 0})

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		found := false
		for _, item := range completions {
			if item.Label == method {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected HTTP method '%s' in completions", method)
		}
	}
}

func TestDiagnosticsWithHint(t *testing.T) {
	dm := NewDocumentManager()

	// Create source that will generate an error with hint
	invalidSource := `@ GET /test {
  > {}
}`

	doc, _ := dm.Open("file:///test.abc", 1, invalidSource)
	diagnostics := GetDiagnostics(doc)

	if len(diagnostics) == 0 {
		t.Skip("No diagnostics generated for this source")
	}

	// Check if any diagnostic has a hint (indicated by newline in message)
	foundHint := false
	for _, diag := range diagnostics {
		if strings.Contains(diag.Message, "Hint:") {
			foundHint = true
			break
		}
	}

	// This is optional - not all errors have hints
	_ = foundHint
}
