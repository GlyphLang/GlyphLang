package formatter

import (
	"strings"
	"unicode"
)

// Token-level transformer that converts between compact (.glyph) and expanded (.glyphx) syntax
// Preserves comments, whitespace, and formatting

// Symbol to keyword mappings
var symbolToKeyword = map[string]string{
	"@": "route",
	":": "type",
	"$": "let",
	">": "return",
	"+": "middleware",
	"%": "use",
	"<": "expects",
	"?": "validate",
	"~": "handle",
	"*": "cron",
	"!": "command",
	"&": "queue",
	"=": "func",
}

// Keyword to symbol mappings (reverse of above)
var keywordToSymbol = map[string]string{
	"route":      "@",
	"type":       ":",
	"let":        "$",
	"return":     ">",
	"middleware": "+",
	"use":        "%",
	"expects":    "<",
	"validate":   "?",
	"handle":     "~",
	"cron":       "*",
	"command":    "!",
	"queue":      "&",
	"func":       "=",
}

// ExpandSource converts compact .glyph syntax to expanded .glyphx syntax
// Preserves comments and formatting
func ExpandSource(source string) string {
	return transform(source, symbolToKeyword, true)
}

// CompactSource converts expanded .glyphx syntax to compact .glyph syntax
// Preserves comments and formatting
func CompactSource(source string) string {
	return transform(source, keywordToSymbol, false)
}

func transform(source string, mappings map[string]string, expandMode bool) string {
	var result strings.Builder
	i := 0
	n := len(source)

	for i < n {
		ch := source[i]

		// Skip and preserve comments
		if ch == '#' || (ch == '/' && i+1 < n && source[i+1] == '/') {
			start := i
			for i < n && source[i] != '\n' {
				i++
			}
			result.WriteString(source[start:i])
			continue
		}

		// Skip and preserve strings
		if ch == '"' || ch == '\'' {
			quote := ch
			result.WriteByte(ch)
			i++
			for i < n && source[i] != quote {
				if source[i] == '\\' && i+1 < n {
					result.WriteByte(source[i])
					i++
					if i < n {
						result.WriteByte(source[i])
						i++
					}
				} else {
					result.WriteByte(source[i])
					i++
				}
			}
			if i < n {
				result.WriteByte(source[i]) // closing quote
				i++
			}
			continue
		}

		// Handle whitespace
		if unicode.IsSpace(rune(ch)) {
			result.WriteByte(ch)
			i++
			continue
		}

		if expandMode {
			// Expanding: look for symbols to convert to keywords
			sym := string(ch)
			if keyword, ok := mappings[sym]; ok {
				// Check context - only transform at statement start or after certain tokens
				if shouldTransformSymbol(source, i, sym) {
					result.WriteString(keyword)
					// Add space after keyword if next char is not whitespace/newline
					if i+1 < n && !unicode.IsSpace(rune(source[i+1])) && source[i+1] != '{' {
						result.WriteByte(' ')
					}
					i++
					continue
				}
			}
		} else {
			// Compacting: look for keywords to convert to symbols
			if unicode.IsLetter(rune(ch)) {
				// Read the full identifier
				start := i
				for i < n && (unicode.IsLetter(rune(source[i])) || unicode.IsDigit(rune(source[i])) || source[i] == '_') {
					i++
				}
				word := source[start:i]

				if symbol, ok := mappings[word]; ok {
					// Check context - only transform at statement start
					if shouldTransformKeyword(source, start, word) {
						result.WriteString(symbol)
						// Keep exactly one space after symbol if there was whitespace
						if i < n && source[i] == ' ' {
							result.WriteByte(' ')
							i++
							// Skip any additional spaces
							for i < n && source[i] == ' ' {
								i++
							}
						}
						continue
					}
				}

				result.WriteString(word)
				continue
			}
		}

		// Default: copy character as-is
		result.WriteByte(ch)
		i++
	}

	return result.String()
}

// shouldTransformSymbol checks if a symbol should be transformed based on context
func shouldTransformSymbol(source string, pos int, sym string) bool {
	// Find the start of the current line
	lineStart := pos
	for lineStart > 0 && source[lineStart-1] != '\n' {
		lineStart--
	}

	// Check what's before this position on the line (skip whitespace)
	beforeOnLine := strings.TrimSpace(source[lineStart:pos])

	switch sym {
	case "@", ":", "~", "*", "!", "&", "=":
		// These should only appear at the start of a line (possibly after whitespace)
		return beforeOnLine == ""
	case "$", ">", "%", "<", "?":
		// These can appear at line start only (inside blocks)
		return beforeOnLine == ""
	case "+":
		// + is only middleware when at line start, otherwise it's addition
		return beforeOnLine == ""
	}
	return false
}

// shouldTransformKeyword checks if a keyword should be transformed based on context
func shouldTransformKeyword(source string, pos int, keyword string) bool {
	// Find the start of the current line
	lineStart := pos
	for lineStart > 0 && source[lineStart-1] != '\n' {
		lineStart--
	}

	// Check what's before this position on the line (skip whitespace)
	beforeOnLine := strings.TrimSpace(source[lineStart:pos])

	switch keyword {
	case "route", "type", "handle", "cron", "command", "queue", "func":
		// These should only appear at the start of a line
		return beforeOnLine == ""
	case "let", "return", "middleware", "use", "expects", "validate":
		// These can appear at line start or inside blocks
		if beforeOnLine == "" {
			return true
		}
		return isInsideBlock(source, pos)
	}
	return false
}

// isInsideBlock checks if position is inside a block (between { and })
func isInsideBlock(source string, pos int) bool {
	depth := 0
	for i := 0; i < pos; i++ {
		if source[i] == '"' || source[i] == '\'' {
			// Skip strings
			quote := source[i]
			i++
			for i < pos && source[i] != quote {
				if source[i] == '\\' {
					i++
				}
				i++
			}
			continue
		}
		if source[i] == '{' {
			depth++
		} else if source[i] == '}' {
			depth--
		}
	}
	return depth > 0
}
