package lsp

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"strings"
	"sync"

	"github.com/glyphlang/glyph/pkg/parser"
)

// Document represents an open document with cached AST
type Document struct {
	URI     string
	Version int
	Content string
	Lines   []string
	AST     *interpreter.Module
	Errors  []parser.ParseError
}

// DocumentManager manages open documents and their cached data
type DocumentManager struct {
	mu        sync.RWMutex
	documents map[string]*Document
}

// NewDocumentManager creates a new document manager
func NewDocumentManager() *DocumentManager {
	return &DocumentManager{
		documents: make(map[string]*Document),
	}
}

// Open opens a new document and parses it
func (dm *DocumentManager) Open(uri string, version int, content string) (*Document, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	doc := &Document{
		URI:     uri,
		Version: version,
		Content: content,
		Lines:   splitLines(content),
	}

	// Parse the document
	dm.parseDocument(doc)

	dm.documents[uri] = doc
	return doc, nil
}

// Update updates an existing document with new content
func (dm *DocumentManager) Update(uri string, version int, changes []TextDocumentContentChangeEvent) (*Document, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	doc, exists := dm.documents[uri]
	if !exists {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	// Apply changes
	for _, change := range changes {
		if change.Range == nil {
			// Full document sync
			doc.Content = change.Text
		} else {
			// Incremental sync (for now, we'll treat as full sync)
			doc.Content = change.Text
		}
	}

	doc.Version = version
	doc.Lines = splitLines(doc.Content)

	// Re-parse the document
	dm.parseDocument(doc)

	return doc, nil
}

// Close closes a document
func (dm *DocumentManager) Close(uri string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if _, exists := dm.documents[uri]; !exists {
		return fmt.Errorf("document not found: %s", uri)
	}

	delete(dm.documents, uri)
	return nil
}

// Get retrieves a document by URI
func (dm *DocumentManager) Get(uri string) (*Document, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	doc, exists := dm.documents[uri]
	return doc, exists
}

// GetAll returns all open documents
func (dm *DocumentManager) GetAll() []*Document {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	docs := make([]*Document, 0, len(dm.documents))
	for _, doc := range dm.documents {
		docs = append(docs, doc)
	}
	return docs
}

// IsGlyphX returns true if the document is a .glyphx file (expanded syntax)
func (doc *Document) IsGlyphX() bool {
	return strings.HasSuffix(doc.URI, ".glyphx")
}

// parseDocument parses a document and updates its AST and errors
func (dm *DocumentManager) parseDocument(doc *Document) {
	// Tokenize using the appropriate lexer based on file extension
	var tokens []parser.Token
	var err error
	if doc.IsGlyphX() {
		lexer := parser.NewExpandedLexer(doc.Content)
		tokens, err = lexer.Tokenize()
	} else {
		lexer := parser.NewLexer(doc.Content)
		tokens, err = lexer.Tokenize()
	}
	if err != nil {
		// Lexer error
		if parseErr, ok := err.(*parser.ParseError); ok {
			doc.Errors = []parser.ParseError{*parseErr}
		} else if lexErr, ok := err.(*parser.LexError); ok {
			doc.Errors = []parser.ParseError{
				{
					Message: lexErr.Message,
					Line:    lexErr.Line,
					Column:  lexErr.Column,
					Source:  doc.Content,
				},
			}
		} else {
			doc.Errors = []parser.ParseError{
				{
					Message: err.Error(),
					Line:    1,
					Column:  1,
					Source:  doc.Content,
				},
			}
		}
		doc.AST = nil
		return
	}

	// Parse
	p := parser.NewParserWithSource(tokens, doc.Content)
	module, err := p.Parse()
	if err != nil {
		// Parse error
		if parseErr, ok := err.(*parser.ParseError); ok {
			doc.Errors = []parser.ParseError{*parseErr}
		} else {
			doc.Errors = []parser.ParseError{
				{
					Message: err.Error(),
					Line:    1,
					Column:  1,
					Source:  doc.Content,
				},
			}
		}
		doc.AST = nil
		return
	}

	// Success
	doc.AST = module
	doc.Errors = nil
}

// GetWordAtPosition returns the word at a given position
func (doc *Document) GetWordAtPosition(pos Position) string {
	if pos.Line < 0 || pos.Line >= len(doc.Lines) {
		return ""
	}

	line := doc.Lines[pos.Line]
	if pos.Character < 0 || pos.Character >= len(line) {
		return ""
	}

	// Find word boundaries
	start := pos.Character
	for start > 0 && isIdentifierChar(line[start-1]) {
		start--
	}

	end := pos.Character
	for end < len(line) && isIdentifierChar(line[end]) {
		end++
	}

	if start >= end {
		return ""
	}

	return line[start:end]
}

// GetLineContent returns the content of a specific line
func (doc *Document) GetLineContent(line int) string {
	if line < 0 || line >= len(doc.Lines) {
		return ""
	}
	return doc.Lines[line]
}

// OffsetToPosition converts a byte offset to a Position
func (doc *Document) OffsetToPosition(offset int) Position {
	if offset < 0 {
		return Position{Line: 0, Character: 0}
	}

	lineNum := 0
	currentOffset := 0

	for lineNum < len(doc.Lines) {
		lineLength := len(doc.Lines[lineNum]) + 1 // +1 for newline
		if currentOffset+lineLength > offset {
			return Position{
				Line:      lineNum,
				Character: offset - currentOffset,
			}
		}
		currentOffset += lineLength
		lineNum++
	}

	// Return last position if offset is beyond document
	if len(doc.Lines) > 0 {
		lastLine := len(doc.Lines) - 1
		return Position{
			Line:      lastLine,
			Character: len(doc.Lines[lastLine]),
		}
	}

	return Position{Line: 0, Character: 0}
}

// PositionToOffset converts a Position to a byte offset
func (doc *Document) PositionToOffset(pos Position) int {
	if pos.Line < 0 || pos.Line >= len(doc.Lines) {
		return 0
	}

	offset := 0
	for i := 0; i < pos.Line; i++ {
		offset += len(doc.Lines[i]) + 1 // +1 for newline
	}

	offset += pos.Character
	return offset
}

// Helper functions

// splitLines splits content into lines
func splitLines(content string) []string {
	lines := []string{}
	currentLine := ""

	for _, ch := range content {
		if ch == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else if ch != '\r' {
			currentLine += string(ch)
		}
	}

	// Always add the last line (even if empty content or empty last line)
	lines = append(lines, currentLine)

	return lines
}

// isIdentifierChar checks if a character is part of an identifier
func isIdentifierChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_'
}

// GetWordRangeAtPosition returns the range of the word at the given position
func (doc *Document) GetWordRangeAtPosition(pos Position) Range {
	if pos.Line < 0 || pos.Line >= len(doc.Lines) {
		return Range{Start: pos, End: pos}
	}

	line := doc.Lines[pos.Line]
	if pos.Character < 0 || pos.Character >= len(line) {
		return Range{Start: pos, End: pos}
	}

	// Find word boundaries
	start := pos.Character
	for start > 0 && isIdentifierChar(line[start-1]) {
		start--
	}

	end := pos.Character
	for end < len(line) && isIdentifierChar(line[end]) {
		end++
	}

	return Range{
		Start: Position{Line: pos.Line, Character: start},
		End:   Position{Line: pos.Line, Character: end},
	}
}

// GetTextInRange returns the text within the given range
func (doc *Document) GetTextInRange(r Range) string {
	if r.Start.Line < 0 || r.Start.Line >= len(doc.Lines) {
		return ""
	}
	if r.End.Line < 0 || r.End.Line >= len(doc.Lines) {
		return ""
	}

	// Single line range
	if r.Start.Line == r.End.Line {
		line := doc.Lines[r.Start.Line]
		if r.Start.Character < 0 || r.End.Character > len(line) {
			return ""
		}
		return line[r.Start.Character:r.End.Character]
	}

	// Multi-line range
	var result string

	// First line (from start character to end of line)
	if r.Start.Character < len(doc.Lines[r.Start.Line]) {
		result = doc.Lines[r.Start.Line][r.Start.Character:]
	}
	result += "\n"

	// Middle lines (complete lines)
	for i := r.Start.Line + 1; i < r.End.Line; i++ {
		result += doc.Lines[i] + "\n"
	}

	// Last line (from beginning to end character)
	if r.End.Character <= len(doc.Lines[r.End.Line]) {
		result += doc.Lines[r.End.Line][:r.End.Character]
	}

	return result
}

// GetLine returns the content of a specific line by index
func (doc *Document) GetLine(lineNum int) string {
	if lineNum < 0 || lineNum >= len(doc.Lines) {
		return ""
	}
	return doc.Lines[lineNum]
}
