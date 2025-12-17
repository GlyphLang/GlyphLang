package lsp

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
)

func TestDocumentManager(t *testing.T) {
	dm := NewDocumentManager()

	// Test opening a document
	source := `: User {
  name: str!
  email: str!
}

@ route /api/users [GET]
  > {users: []}
`

	doc, err := dm.Open("file:///test.abc", 1, source)
	if err != nil {
		t.Fatalf("Failed to open document: %v", err)
	}

	if doc.URI != "file:///test.abc" {
		t.Errorf("Expected URI 'file:///test.abc', got '%s'", doc.URI)
	}

	if doc.Version != 1 {
		t.Errorf("Expected version 1, got %d", doc.Version)
	}

	if doc.Content != source {
		t.Error("Content mismatch")
	}

	// Test getting a document
	retrieved, exists := dm.Get("file:///test.abc")
	if !exists {
		t.Fatal("Document should exist")
	}

	if retrieved.URI != doc.URI {
		t.Error("Retrieved document URI mismatch")
	}

	// Test updating a document
	newSource := `: User {
  name: str!
  email: str!
  age: int
}
`

	changes := []TextDocumentContentChangeEvent{
		{Text: newSource},
	}

	updated, err := dm.Update("file:///test.abc", 2, changes)
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}

	if updated.Version != 2 {
		t.Errorf("Expected version 2, got %d", updated.Version)
	}

	if updated.Content != newSource {
		t.Error("Updated content mismatch")
	}

	// Test closing a document
	if err := dm.Close("file:///test.abc"); err != nil {
		t.Fatalf("Failed to close document: %v", err)
	}

	_, exists = dm.Get("file:///test.abc")
	if exists {
		t.Error("Document should not exist after closing")
	}
}

func TestDocumentParsing(t *testing.T) {
	dm := NewDocumentManager()

	validSource := `: User {
  name: str!
}

@ route /api/user [GET]
  > {name: "test"}
`

	doc, err := dm.Open("file:///valid.abc", 1, validSource)
	if err != nil {
		t.Fatalf("Failed to open document: %v", err)
	}

	if doc.AST == nil {
		t.Fatal("Expected AST to be parsed")
	}

	if len(doc.Errors) > 0 {
		t.Errorf("Expected no errors, got: %v", doc.Errors)
	}

	// Check AST structure
	if len(doc.AST.Items) != 2 {
		t.Errorf("Expected 2 items in AST, got %d", len(doc.AST.Items))
	}

	// Check for TypeDef
	typeDef, ok := doc.AST.Items[0].(*interpreter.TypeDef)
	if !ok {
		t.Error("Expected first item to be TypeDef")
	} else if typeDef.Name != "User" {
		t.Errorf("Expected TypeDef name 'User', got '%s'", typeDef.Name)
	}

	// Check for Route
	route, ok := doc.AST.Items[1].(*interpreter.Route)
	if !ok {
		t.Error("Expected second item to be Route")
	} else if route.Path != "/api/user" {
		t.Errorf("Expected route path '/api/user', got '%s'", route.Path)
	}
}

func TestDocumentParsingErrors(t *testing.T) {
	dm := NewDocumentManager()

	invalidSource := `: User {
  name str!
}
`

	doc, err := dm.Open("file:///invalid.abc", 1, invalidSource)
	if err != nil {
		t.Fatalf("Failed to open document: %v", err)
	}

	if doc.AST != nil {
		t.Error("Expected AST to be nil for invalid source")
	}

	if len(doc.Errors) == 0 {
		t.Error("Expected parse errors")
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "Empty",
			content:  "",
			expected: 1,
		},
		{
			name:     "Single line",
			content:  "hello",
			expected: 1,
		},
		{
			name:     "Two lines",
			content:  "hello\nworld",
			expected: 2,
		},
		{
			name:     "Three lines with empty",
			content:  "hello\n\nworld",
			expected: 3,
		},
		{
			name:     "Trailing newline",
			content:  "hello\nworld\n",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := splitLines(tt.content)
			if len(lines) != tt.expected {
				t.Errorf("Expected %d lines, got %d", tt.expected, len(lines))
			}
		})
	}
}

func TestGetWordAtPosition(t *testing.T) {
	source := `: User {
  name: str!
  email: str!
}
`

	dm := NewDocumentManager()
	doc, _ := dm.Open("file:///test.abc", 1, source)

	tests := []struct {
		name     string
		pos      Position
		expected string
	}{
		{
			name:     "Type name",
			pos:      Position{Line: 0, Character: 2},
			expected: "User",
		},
		{
			name:     "Field name",
			pos:      Position{Line: 1, Character: 3},
			expected: "name",
		},
		{
			name:     "Type annotation",
			pos:      Position{Line: 1, Character: 9},
			expected: "str",
		},
		{
			name:     "Out of bounds",
			pos:      Position{Line: 100, Character: 0},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := doc.GetWordAtPosition(tt.pos)
			if word != tt.expected {
				t.Errorf("Expected word '%s', got '%s'", tt.expected, word)
			}
		})
	}
}

func TestGetLineContent(t *testing.T) {
	source := `line 1
line 2
line 3`

	dm := NewDocumentManager()
	doc, _ := dm.Open("file:///test.abc", 1, source)

	tests := []struct {
		line     int
		expected string
	}{
		{0, "line 1"},
		{1, "line 2"},
		{2, "line 3"},
		{-1, ""},
		{100, ""},
	}

	for _, tt := range tests {
		content := doc.GetLineContent(tt.line)
		if content != tt.expected {
			t.Errorf("Line %d: expected '%s', got '%s'", tt.line, tt.expected, content)
		}
	}
}

func TestPositionConversion(t *testing.T) {
	source := `abc
def
ghi`

	dm := NewDocumentManager()
	doc, _ := dm.Open("file:///test.abc", 1, source)

	tests := []struct {
		name     string
		pos      Position
		offset   int
		roundTrip bool
	}{
		{
			name:      "Start of file",
			pos:       Position{Line: 0, Character: 0},
			offset:    0,
			roundTrip: true,
		},
		{
			name:      "Middle of first line",
			pos:       Position{Line: 0, Character: 2},
			offset:    2,
			roundTrip: true,
		},
		{
			name:      "Start of second line",
			pos:       Position{Line: 1, Character: 0},
			offset:    4, // "abc\n"
			roundTrip: true,
		},
		{
			name:      "Middle of third line",
			pos:       Position{Line: 2, Character: 1},
			offset:    9, // "abc\ndef\ng"
			roundTrip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := doc.PositionToOffset(tt.pos)
			if offset != tt.offset {
				t.Errorf("PositionToOffset: expected %d, got %d", tt.offset, offset)
			}

			if tt.roundTrip {
				pos := doc.OffsetToPosition(offset)
				if pos.Line != tt.pos.Line || pos.Character != tt.pos.Character {
					t.Errorf("OffsetToPosition: expected %+v, got %+v", tt.pos, pos)
				}
			}
		})
	}
}

func TestIsIdentifierChar(t *testing.T) {
	tests := []struct {
		ch       byte
		expected bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'_', true},
		{'-', false},
		{' ', false},
		{':', false},
		{'{', false},
	}

	for _, tt := range tests {
		result := isIdentifierChar(tt.ch)
		if result != tt.expected {
			t.Errorf("isIdentifierChar('%c'): expected %v, got %v", tt.ch, tt.expected, result)
		}
	}
}

func TestDocumentGetAll(t *testing.T) {
	dm := NewDocumentManager()

	// Open multiple documents
	dm.Open("file:///test1.abc", 1, ": User { name: str! }")
	dm.Open("file:///test2.abc", 1, "@ route /test [GET]\n  > {}")
	dm.Open("file:///test3.abc", 1, ": Product { price: int! }")

	all := dm.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(all))
	}

	// Close one document
	dm.Close("file:///test2.abc")

	all = dm.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 documents after closing one, got %d", len(all))
	}
}

func TestDocumentUpdateNonExistent(t *testing.T) {
	dm := NewDocumentManager()

	changes := []TextDocumentContentChangeEvent{
		{Text: "new content"},
	}

	_, err := dm.Update("file:///nonexistent.abc", 2, changes)
	if err == nil {
		t.Error("Expected error when updating non-existent document")
	}
}

func TestDocumentCloseNonExistent(t *testing.T) {
	dm := NewDocumentManager()

	err := dm.Close("file:///nonexistent.abc")
	if err == nil {
		t.Error("Expected error when closing non-existent document")
	}
}
