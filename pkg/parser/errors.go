package parser

import (
	"fmt"
	"strings"
)

// ParseError represents a parsing error with context
type ParseError struct {
	Message string
	Line    int
	Column  int
	Source  string
	Hint    string
}

func (e *ParseError) Error() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Error at line %d, column %d:\n", e.Line, e.Column))

	// Show the source line with context
	if e.Source != "" {
		lines := strings.Split(e.Source, "\n")
		if e.Line > 0 && e.Line <= len(lines) {
			lineNum := e.Line

			// Show previous line for context if available
			if lineNum > 1 {
				builder.WriteString(fmt.Sprintf("  %4d | %s\n", lineNum-1, lines[lineNum-2]))
			}

			// Show the error line
			errorLine := lines[lineNum-1]
			builder.WriteString(fmt.Sprintf("  %4d | %s\n", lineNum, errorLine))

			// Show caret pointing to error column
			if e.Column > 0 {
				spaces := strings.Repeat(" ", e.Column-1)
				builder.WriteString(fmt.Sprintf("       | %s^\n", spaces))
			}

			// Show next line for context if available
			if lineNum < len(lines) {
				builder.WriteString(fmt.Sprintf("  %4d | %s\n", lineNum+1, lines[lineNum]))
			}
		}
	}

	builder.WriteString(fmt.Sprintf("\n%s", e.Message))

	if e.Hint != "" {
		builder.WriteString(fmt.Sprintf("\n\nHint: %s", e.Hint))
	}

	return builder.String()
}

// LexError represents a lexical analysis error
type LexError struct {
	Message string
	Line    int
	Column  int
	Source  string
	Char    byte
}

func (e *LexError) Error() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Lexer error at line %d, column %d:\n", e.Line, e.Column))

	// Show the source line with context
	if e.Source != "" {
		lines := strings.Split(e.Source, "\n")
		if e.Line > 0 && e.Line <= len(lines) {
			lineNum := e.Line
			errorLine := lines[lineNum-1]

			builder.WriteString(fmt.Sprintf("  %4d | %s\n", lineNum, errorLine))

			// Show caret pointing to error column
			if e.Column > 0 {
				spaces := strings.Repeat(" ", e.Column-1)
				builder.WriteString(fmt.Sprintf("       | %s^\n", spaces))
			}
		}
	}

	builder.WriteString(fmt.Sprintf("\n%s", e.Message))

	return builder.String()
}

// Parser error helper functions

// errorWithContext creates a detailed parse error with source context
func (p *Parser) errorWithContext(msg string, tok Token) error {
	return &ParseError{
		Message: msg,
		Line:    tok.Line,
		Column:  tok.Column,
		Source:  p.source,
	}
}

// errorWithHint creates a parse error with a helpful hint
func (p *Parser) errorWithHint(msg string, tok Token, hint string) error {
	return &ParseError{
		Message: msg,
		Line:    tok.Line,
		Column:  tok.Column,
		Source:  p.source,
		Hint:    hint,
	}
}

// expectError creates a contextual error for expected token mismatches
func (p *Parser) expectError(expected TokenType, got Token) error {
	msg := fmt.Sprintf("Expected %s, but found %s", expected, got.Type)

	// Add helpful hints for common mistakes
	var hint string
	switch {
	case expected == LBRACE && got.Type == NEWLINE:
		hint = "Did you forget to add an opening brace '{'?"
	case expected == LBRACE && got.Type == ARROW:
		hint = "Type definitions and routes require a body enclosed in braces { }"
	case expected == RBRACE && got.Type == EOF:
		hint = "Missing closing brace '}'. Check if all opened braces are properly closed"
	case expected == SLASH && got.Type == IDENT:
		hint = "Route paths must start with '/'"
	case expected == LPAREN && got.Type == NEWLINE:
		hint = "Did you forget to add an opening parenthesis '('?"
	case expected == RPAREN && got.Type == EOF:
		hint = "Missing closing parenthesis ')'. Check if all opened parentheses are properly closed"
	case got.Type == ARROW && (expected == LBRACE || expected == NEWLINE):
		hint = "The '->' symbol is used for return types. If you want to define a route body, use braces { } after the path"
	}

	return p.errorWithHint(msg, got, hint)
}

// routeError creates a contextual error for route parsing
func (p *Parser) routeError(msg string, tok Token) error {
	return p.errorWithHint(msg, tok, "Routes should follow the pattern: @ route /path [METHOD] -> ReturnType")
}

// typeError creates a contextual error for type parsing
func (p *Parser) typeError(msg string, tok Token) error {
	return p.errorWithHint(msg, tok, "Valid types are: int, str, bool, float, or custom type names")
}

// expressionError creates a contextual error for expression parsing
func (p *Parser) expressionError(msg string, tok Token) error {
	return p.errorWithHint(msg, tok, "Expected a valid expression (number, string, variable, or function call)")
}

// Lexer error helper functions

// errorAtPosition creates a detailed lexer error at the current position
func (l *Lexer) errorAtPosition(msg string) error {
	return &LexError{
		Message: msg,
		Line:    l.line,
		Column:  l.column,
		Source:  l.input,
		Char:    l.ch,
	}
}

// unterminatedStringError creates a helpful error for unterminated strings
func (l *Lexer) unterminatedStringError(startLine, startCol int, quote byte) error {
	msg := fmt.Sprintf("Unterminated string literal starting with %c", quote)

	return &LexError{
		Message: msg + "\n\nHint: Make sure to close the string with a matching quote",
		Line:    startLine,
		Column:  startCol,
		Source:  l.input,
		Char:    quote,
	}
}

// invalidCharacterError creates a helpful error for invalid characters
func (l *Lexer) invalidCharacterError() error {
	char := l.ch
	msg := fmt.Sprintf("Unexpected character '%c' (0x%02X)", char, char)

	var hint string
	switch {
	case char >= 'A' && char <= 'Z':
		hint = "Identifiers and keywords in GLYPH are case-sensitive"
	case char == ';':
		hint = "GLYPH uses newlines for statement separation, not semicolons"
	case char == '`':
		hint = "Use double quotes (\") or single quotes (') for strings"
	case char > 127:
		hint = "GLYPH source code must use ASCII characters only"
	default:
		hint = "This character is not valid in GLYPH syntax"
	}

	if hint != "" {
		msg = msg + ". " + hint
	}

	return &LexError{
		Message: msg,
		Line:    l.line,
		Column:  l.column,
		Source:  l.input,
		Char:    char,
	}
}

// Common error message builders

// buildMissingTokenError creates a helpful error message for missing tokens
func buildMissingTokenError(expected string, context string) string {
	return fmt.Sprintf("Missing %s in %s", expected, context)
}

// buildUnexpectedTokenError creates a helpful error message for unexpected tokens
func buildUnexpectedTokenError(got TokenType, context string) string {
	return fmt.Sprintf("Unexpected %s in %s", got, context)
}
