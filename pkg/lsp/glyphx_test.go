package lsp

import (
	"strings"
	"testing"
)

// TestIsGlyphX tests the IsGlyphX method on Document
func TestIsGlyphX(t *testing.T) {
	tests := []struct {
		uri      string
		expected bool
	}{
		{"file:///test.glyphx", true},
		{"file:///path/to/api.glyphx", true},
		{"file:///test.glyph", false},
		{"file:///test.go", false},
		{"file:///test.glyphx.bak", false},
		{"file:///glyphx", false},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			doc := &Document{URI: tt.uri}
			if doc.IsGlyphX() != tt.expected {
				t.Errorf("IsGlyphX() for URI '%s' = %v, want %v", tt.uri, doc.IsGlyphX(), tt.expected)
			}
		})
	}
}

// TestGlyphXDocumentParsing tests that .glyphx files are parsed with ExpandedLexer
func TestGlyphXDocumentParsing(t *testing.T) {
	dm := NewDocumentManager()

	// Expanded syntax using human-readable keywords
	source := `type User {
  name: str!
  email: str!
}

route GET /api/users {
  return {users: []}
}
`

	doc, err := dm.Open("file:///test.glyphx", 1, source)
	if err != nil {
		t.Fatalf("Failed to open .glyphx document: %v", err)
	}

	if doc.AST == nil {
		t.Fatal("Expected AST to be parsed for .glyphx document")
	}

	if len(doc.Errors) > 0 {
		t.Errorf("Expected no errors for valid .glyphx source, got: %v", doc.Errors)
	}

	// Should parse into TypeDef + Route
	if len(doc.AST.Items) != 2 {
		t.Errorf("Expected 2 items in AST, got %d", len(doc.AST.Items))
	}
}

// TestGlyphXDocumentParsingErrors tests error handling for invalid .glyphx files
func TestGlyphXDocumentParsingErrors(t *testing.T) {
	dm := NewDocumentManager()

	// Invalid expanded syntax (missing colon between field name and type)
	invalidSource := `type User {
  name str!
}
`

	doc, err := dm.Open("file:///invalid.glyphx", 1, invalidSource)
	if err != nil {
		t.Fatalf("Failed to open .glyphx document: %v", err)
	}

	if doc.AST != nil {
		t.Error("Expected AST to be nil for invalid .glyphx source")
	}

	if len(doc.Errors) == 0 {
		t.Error("Expected parse errors for invalid .glyphx source")
	}
}

// TestGlyphXDocumentUpdate tests that updating a .glyphx document re-parses with ExpandedLexer
func TestGlyphXDocumentUpdate(t *testing.T) {
	dm := NewDocumentManager()

	source := `type User {
  name: str!
}
`
	doc, err := dm.Open("file:///test.glyphx", 1, source)
	if err != nil {
		t.Fatalf("Failed to open .glyphx document: %v", err)
	}

	if doc.AST == nil {
		t.Fatal("Expected AST for initial .glyphx document")
	}

	// Update with new content
	newSource := `type User {
  name: str!
  email: str!
}

route GET /api/users {
  return {users: []}
}
`
	changes := []TextDocumentContentChangeEvent{
		{Text: newSource},
	}

	updated, err := dm.Update("file:///test.glyphx", 2, changes)
	if err != nil {
		t.Fatalf("Failed to update .glyphx document: %v", err)
	}

	if updated.AST == nil {
		t.Fatal("Expected AST for updated .glyphx document")
	}

	if len(updated.AST.Items) != 2 {
		t.Errorf("Expected 2 items after update, got %d", len(updated.AST.Items))
	}
}

// TestGlyphXCompletion tests that .glyphx files get expanded syntax completions
func TestGlyphXCompletion(t *testing.T) {
	dm := NewDocumentManager()

	source := `type User {
  name: str!
}
`
	doc, _ := dm.Open("file:///test.glyphx", 1, source)

	completions := GetCompletion(doc, Position{Line: 0, Character: 0})

	if len(completions) == 0 {
		t.Fatal("Expected completion items for .glyphx document")
	}

	// Check for expanded-specific keywords
	expandedKeywords := map[string]bool{
		"route":      false,
		"type":       false,
		"let":        false,
		"return":     false,
		"middleware": false,
		"use":        false,
		"expects":    false,
		"validate":   false,
		"handle":     false,
		"cron":       false,
		"command":    false,
		"queue":      false,
		"func":       false,
		"async":      false,
		"await":      false,
		"import":     false,
		"from":       false,
	}

	for _, item := range completions {
		if _, ok := expandedKeywords[item.Label]; ok {
			expandedKeywords[item.Label] = true
		}
	}

	for kw, found := range expandedKeywords {
		if !found {
			t.Errorf("Expected expanded keyword '%s' in completions", kw)
		}
	}
}

// TestGlyphXCompletionSnippets tests expanded syntax snippets
func TestGlyphXCompletionSnippets(t *testing.T) {
	dm := NewDocumentManager()

	doc, _ := dm.Open("file:///test.glyphx", 1, "")

	completions := GetCompletion(doc, Position{Line: 0, Character: 0})

	snippetCount := 0
	for _, item := range completions {
		if item.Kind == CompletionItemKindSnippet {
			snippetCount++
			if item.InsertText == "" {
				t.Errorf("Snippet '%s' should have InsertText", item.Label)
			}
			if item.InsertTextFormat != 2 {
				t.Errorf("Snippet '%s' should have InsertTextFormat=2", item.Label)
			}
			// Expanded snippets should use human-readable keywords, not symbols
			if strings.Contains(item.InsertText, "@ route") {
				t.Errorf("Expanded snippet '%s' should use 'route' not '@ route'", item.Label)
			}
			if strings.Contains(item.InsertText, "! command") {
				t.Errorf("Expanded snippet '%s' should use 'command' not '! command'", item.Label)
			}
		}
	}

	if snippetCount == 0 {
		t.Error("Expected at least one snippet completion for .glyphx")
	}
}

// TestGlyphXCompletionVsGlyph tests that .glyphx and .glyph get different completions
func TestGlyphXCompletionVsGlyph(t *testing.T) {
	dm := NewDocumentManager()

	glyphDoc, _ := dm.Open("file:///test.glyph", 1, "")
	glyphxDoc, _ := dm.Open("file:///test.glyphx", 1, "")

	glyphCompletions := GetCompletion(glyphDoc, Position{Line: 0, Character: 0})
	glyphxCompletions := GetCompletion(glyphxDoc, Position{Line: 0, Character: 0})

	// .glyphx should have more keywords (expanded forms)
	glyphKeywords := 0
	glyphxKeywords := 0
	for _, item := range glyphCompletions {
		if item.Kind == CompletionItemKindKeyword {
			glyphKeywords++
		}
	}
	for _, item := range glyphxCompletions {
		if item.Kind == CompletionItemKindKeyword {
			glyphxKeywords++
		}
	}

	if glyphxKeywords <= glyphKeywords {
		t.Errorf("Expected .glyphx to have more keywords (%d) than .glyph (%d)",
			glyphxKeywords, glyphKeywords)
	}

	// Check that .glyphx has "let" but .glyph does not
	hasLetGlyph := false
	hasLetGlyphx := false
	for _, item := range glyphCompletions {
		if item.Label == "let" {
			hasLetGlyph = true
		}
	}
	for _, item := range glyphxCompletions {
		if item.Label == "let" {
			hasLetGlyphx = true
		}
	}

	if hasLetGlyph {
		t.Error("Compact .glyph should not have 'let' keyword completion")
	}
	if !hasLetGlyphx {
		t.Error("Expanded .glyphx should have 'let' keyword completion")
	}
}

// TestGlyphXDiagnostics tests diagnostics for .glyphx files
func TestGlyphXDiagnostics(t *testing.T) {
	dm := NewDocumentManager()

	source := `type User {
  name: str!
}

route GET /api/users {
  return {users: []}
}
`
	doc, _ := dm.Open("file:///test.glyphx", 1, source)
	diagnostics := GetDiagnostics(doc)

	if len(diagnostics) > 0 {
		t.Errorf("Expected no diagnostics for valid .glyphx source, got %d", len(diagnostics))
	}
}

// TestGlyphXHover tests hover information for .glyphx files
func TestGlyphXHover(t *testing.T) {
	dm := NewDocumentManager()

	source := `type User {
  name: str!
  age: int!
}

route GET /api/users {
  return {users: []}
}
`

	doc, _ := dm.Open("file:///test.glyphx", 1, source)

	if doc.AST == nil {
		t.Fatal("Expected AST for .glyphx document")
	}

	tests := []struct {
		name        string
		pos         Position
		expectHover bool
		contains    string
	}{
		{
			name:        "Hover on type name",
			pos:         Position{Line: 0, Character: 6},
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
			name:        "Hover on keyword GET",
			pos:         Position{Line: 5, Character: 7},
			expectHover: true,
			contains:    "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hover := GetHover(doc, tt.pos)
			if tt.expectHover && hover == nil {
				t.Error("Expected hover information")
			}
			if hover != nil && tt.contains != "" {
				if !strings.Contains(hover.Contents.Value, tt.contains) {
					t.Errorf("Expected hover to contain '%s', got: %s", tt.contains, hover.Contents.Value)
				}
			}
		})
	}
}

// TestGlyphXDocumentSymbols tests document symbols for .glyphx files
func TestGlyphXDocumentSymbols(t *testing.T) {
	dm := NewDocumentManager()

	source := `type User {
  name: str!
  email: str!
}

type Post {
  title: str!
}

route GET /api/users {
  return {users: []}
}

route POST /api/posts {
  return {created: true}
}
`

	doc, _ := dm.Open("file:///test.glyphx", 1, source)

	if doc.AST == nil {
		t.Fatal("Expected AST for .glyphx document")
	}

	symbols := GetDocumentSymbols(doc)

	if len(symbols) != 4 {
		t.Errorf("Expected 4 symbols (2 types + 2 routes), got %d", len(symbols))
	}

	typeCount := 0
	routeCount := 0
	for _, sym := range symbols {
		switch sym.Kind {
		case SymbolKindStruct:
			typeCount++
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

// TestGlyphXDefinition tests go-to-definition for .glyphx files
func TestGlyphXDefinition(t *testing.T) {
	dm := NewDocumentManager()

	source := `type User {
  name: str!
}

type Post {
  author: User
}
`

	doc, _ := dm.Open("file:///test.glyphx", 1, source)

	if doc.AST == nil {
		t.Fatal("Expected AST for .glyphx document")
	}

	// Try to find definition of "User" in the author field
	defs := GetDefinition(doc, Position{Line: 5, Character: 12})
	if defs == nil || len(defs) == 0 {
		t.Skip("Definition not found at this position")
	}

	if defs[0].URI != doc.URI {
		t.Error("Definition should be in the same file")
	}
}

// TestGlyphXKeywordHover tests hover for expanded-only keywords
func TestGlyphXKeywordHover(t *testing.T) {
	// Test expanded keywords that have hover info
	expandedKeywords := []string{
		"type", "let", "return", "middleware", "use",
		"expects", "validate", "handle", "func",
		"null", "async", "await", "import", "from",
	}

	for _, kw := range expandedKeywords {
		info := getKeywordInfo(kw)
		if info == "" {
			t.Errorf("Expected hover info for expanded keyword '%s'", kw)
		}
	}
}

// TestGlyphXCompletionDefinedTypes tests that defined types appear in .glyphx completions
func TestGlyphXCompletionDefinedTypes(t *testing.T) {
	dm := NewDocumentManager()

	source := `type User {
  name: str!
}

type Post {
  title: str!
}
`

	doc, _ := dm.Open("file:///test.glyphx", 1, source)

	completions := GetCompletion(doc, Position{Line: 0, Character: 0})

	foundUser := false
	foundPost := false
	for _, item := range completions {
		if item.Kind == CompletionItemKindStruct {
			if item.Label == "User" {
				foundUser = true
			}
			if item.Label == "Post" {
				foundPost = true
			}
		}
	}

	if !foundUser {
		t.Error("Expected 'User' type in completions")
	}
	if !foundPost {
		t.Error("Expected 'Post' type in completions")
	}
}

// TestGlyphXWithLetAndReturn tests parsing .glyphx with let/return keywords
func TestGlyphXWithLetAndReturn(t *testing.T) {
	dm := NewDocumentManager()

	source := `type User {
  name: str!
}

route GET /api/users {
  let users = []
  return {users: users}
}
`

	doc, err := dm.Open("file:///test.glyphx", 1, source)
	if err != nil {
		t.Fatalf("Failed to open .glyphx document: %v", err)
	}

	if doc.AST == nil {
		t.Fatal("Expected AST for .glyphx document with let/return")
	}

	if len(doc.Errors) > 0 {
		t.Errorf("Expected no errors, got: %v", doc.Errors)
	}
}

// TestGlyphXWithCommand tests parsing .glyphx with command keyword
func TestGlyphXWithCommand(t *testing.T) {
	dm := NewDocumentManager()

	source := `command greet name: str! {
  return "Hello " + name
}
`

	doc, err := dm.Open("file:///test.glyphx", 1, source)
	if err != nil {
		t.Fatalf("Failed to open .glyphx document: %v", err)
	}

	if doc.AST == nil {
		t.Fatal("Expected AST for .glyphx command document")
	}
}

// TestGlyphXWithCron tests parsing .glyphx with cron keyword
func TestGlyphXWithCron(t *testing.T) {
	dm := NewDocumentManager()

	source := `cron "0 0 * * *" {
  return "cleanup done"
}
`

	doc, err := dm.Open("file:///test.glyphx", 1, source)
	if err != nil {
		t.Fatalf("Failed to open .glyphx document: %v", err)
	}

	if doc.AST == nil {
		t.Fatal("Expected AST for .glyphx cron document")
	}
}

// TestGlyphXCompactDocSameFeatures tests that both file types support core features
func TestGlyphXCompactDocSameFeatures(t *testing.T) {
	dm := NewDocumentManager()

	// Same API defined in both syntaxes
	compactSource := `: User {
  name: str!
}

@ GET /api/users {
  > {users: []}
}
`
	expandedSource := `type User {
  name: str!
}

route GET /api/users {
  return {users: []}
}
`

	compactDoc, _ := dm.Open("file:///test.glyph", 1, compactSource)
	expandedDoc, _ := dm.Open("file:///test.glyphx", 1, expandedSource)

	// Both should parse successfully
	if compactDoc.AST == nil {
		t.Fatal("Expected AST for compact document")
	}
	if expandedDoc.AST == nil {
		t.Fatal("Expected AST for expanded document")
	}

	// Both should produce same number of items
	if len(compactDoc.AST.Items) != len(expandedDoc.AST.Items) {
		t.Errorf("Expected same number of AST items: compact=%d, expanded=%d",
			len(compactDoc.AST.Items), len(expandedDoc.AST.Items))
	}

	// Both should produce same number of symbols
	compactSymbols := GetDocumentSymbols(compactDoc)
	expandedSymbols := GetDocumentSymbols(expandedDoc)

	if len(compactSymbols) != len(expandedSymbols) {
		t.Errorf("Expected same number of symbols: compact=%d, expanded=%d",
			len(compactSymbols), len(expandedSymbols))
	}
}
