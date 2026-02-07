package parser

import (
	"fmt"
	"strings"
	"unicode"
)

// Lexer tokenizes GLYPH source code
type Lexer struct {
	input             string
	position          int
	readPosition      int
	ch                byte
	line              int
	column            int
	lastTokenWasValue bool   // Track if last token was a value (for / disambiguation)
	lastTokenLiteral  string // Track last token literal (for route keyword detection)
}

// New creates a new Lexer
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar advances to the next character
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

// peekChar returns the next character without advancing
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// Tokenize returns all tokens from the input
func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token

	for {
		// Skip whitespace except newlines
		l.skipWhitespaceExceptNewlines()

		if l.ch == 0 {
			break
		}

		// Handle newlines
		if l.ch == '\n' {
			tokens = append(tokens, Token{
				Type:   NEWLINE,
				Line:   l.line,
				Column: l.column,
			})
			l.line++
			l.column = 0
			l.readChar()
			l.lastTokenWasValue = false
			l.lastTokenLiteral = ""
			continue
		}

		// Skip comments (both # and //)
		if l.ch == '#' {
			l.skipComment()
			continue
		}
		if l.ch == '/' && l.peekChar() == '/' {
			l.skipComment()
			continue
		}

		tok := l.nextToken()
		if tok.Type == ILLEGAL {
			// Check if it's an unterminated string error
			if strings.HasPrefix(tok.Literal, "unterminated_string:") {
				quote := tok.Literal[len("unterminated_string:"):]
				return nil, l.unterminatedStringError(tok.Line, tok.Column, quote[0])
			}
			return nil, l.invalidCharacterError()
		}

		// Track if token was a value (for / disambiguation)
		l.lastTokenWasValue = tok.Type == IDENT || tok.Type == INTEGER ||
			tok.Type == FLOAT || tok.Type == STRING || tok.Type == RPAREN ||
			tok.Type == RBRACKET || tok.Type == TRUE || tok.Type == FALSE
		l.lastTokenLiteral = tok.Literal

		tokens = append(tokens, tok)

		if tok.Type == EOF {
			break
		}
	}

	// Add EOF token if not already present
	if len(tokens) == 0 || tokens[len(tokens)-1].Type != EOF {
		tokens = append(tokens, Token{Type: EOF, Line: l.line, Column: l.column})
	}

	return tokens, nil
}

// nextToken returns the next token
func (l *Lexer) nextToken() Token {
	var tok Token
	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case 0:
		tok.Type = EOF
		tok.Literal = ""
	case '@':
		tok.Type = AT
		tok.Literal = string(l.ch)
		l.readChar()
	case ':':
		tok.Type = COLON
		tok.Literal = string(l.ch)
		l.readChar()
	case '$':
		tok.Type = DOLLAR
		tok.Literal = string(l.ch)
		l.readChar()
	case '+':
		tok.Type = PLUS
		tok.Literal = string(l.ch)
		l.readChar()
	case '*':
		tok.Type = STAR
		tok.Literal = string(l.ch)
		l.readChar()
	case '%':
		tok.Type = PERCENT
		tok.Literal = string(l.ch)
		l.readChar()
	case '/':
		// Check if this is a path or division
		// Paths occur after "route" keyword or at statement start
		// Division occurs after values or in rate limits like "100/min"
		nextChar := l.peekChar()
		isPathContext := l.lastTokenLiteral == "route" || !l.lastTokenWasValue
		isPathStart := unicode.IsLetter(rune(nextChar))

		if isPathContext && isPathStart {
			// This is a path like /api/users
			tok = l.readIdentifier()
		} else {
			tok.Type = SLASH
			tok.Literal = string(l.ch)
			l.readChar()
		}
	case '~':
		tok.Type = TILDE
		tok.Literal = string(l.ch)
		l.readChar()
	case '(':
		tok.Type = LPAREN
		tok.Literal = string(l.ch)
		l.readChar()
	case ')':
		tok.Type = RPAREN
		tok.Literal = string(l.ch)
		l.readChar()
	case '{':
		tok.Type = LBRACE
		tok.Literal = string(l.ch)
		l.readChar()
	case '}':
		tok.Type = RBRACE
		tok.Literal = string(l.ch)
		l.readChar()
	case '[':
		tok.Type = LBRACKET
		tok.Literal = string(l.ch)
		l.readChar()
	case ']':
		tok.Type = RBRACKET
		tok.Literal = string(l.ch)
		l.readChar()
	case ',':
		tok.Type = COMMA
		tok.Literal = string(l.ch)
		l.readChar()
	case '.':
		// Check for ... (three dots)
		if l.peekChar() == '.' {
			// Save position to check for third dot
			l.readChar() // consume first dot, now on second dot
			if l.peekChar() == '.' {
				// We have three dots
				l.readChar() // consume second dot
				tok.Type = DOTDOTDOT
				tok.Literal = "..."
				l.readChar() // consume third dot
			} else {
				// Just two dots - rare case, return first as DOT
				// The second dot is already consumed but we'll return DOT
				// This is a simplification - two dots isn't a valid token anyway
				tok.Type = DOT
				tok.Literal = "."
			}
		} else {
			tok.Type = DOT
			tok.Literal = string(l.ch)
			l.readChar()
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = GREATER_EQ
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else {
			tok.Type = GREATER
			tok.Literal = string(l.ch)
			l.readChar()
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = LESS_EQ
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else {
			tok.Type = LESS
			tok.Literal = string(l.ch)
			l.readChar()
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = NOT_EQ
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else {
			tok.Type = BANG
			tok.Literal = string(l.ch)
			l.readChar()
		}
	case '?':
		tok.Type = QUESTION
		tok.Literal = string(l.ch)
		l.readChar()
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok.Type = AND
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else {
			tok.Type = AMPERSAND
			tok.Literal = string(l.ch)
			l.readChar()
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok.Type = OR
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok.Type = PIPE_OP
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else {
			tok.Type = PIPE
			tok.Literal = string(l.ch)
			l.readChar()
		}
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = EQ_EQ
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok.Type = FATARROW
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else {
			tok.Type = EQUALS
			tok.Literal = string(l.ch)
			l.readChar()
		}
	case '-':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok.Type = ARROW
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else {
			tok.Type = MINUS
			tok.Literal = string(l.ch)
			l.readChar()
		}
	case '"', '\'':
		tok = l.readString()
	default:
		if unicode.IsDigit(rune(l.ch)) {
			tok = l.readNumber()
		} else if isIdentifierStart(l.ch) {
			tok = l.readIdentifier()
		} else {
			tok.Type = ILLEGAL
			tok.Literal = string(l.ch)
		}
	}

	return tok
}

// readIdentifier reads an identifier or path
func (l *Lexer) readIdentifier() Token {
	tok := Token{Line: l.line, Column: l.column}
	position := l.position

	// Handle paths like /api/users/:id or /order-status
	if l.ch == '/' {
		for l.ch == '/' || isIdentifierChar(l.ch) || l.ch == ':' || l.ch == '-' {
			l.readChar()
		}
	} else {
		// Regular identifier
		for isIdentifierChar(l.ch) {
			l.readChar()
		}
	}

	tok.Literal = l.input[position:l.position]

	// Check for keywords
	switch tok.Literal {
	case "true":
		tok.Type = TRUE
	case "false":
		tok.Type = FALSE
	case "null":
		tok.Type = NULL
	case "while":
		tok.Type = WHILE
	case "switch":
		tok.Type = SWITCH
	case "case":
		tok.Type = CASE
	case "default":
		tok.Type = DEFAULT
	case "for":
		tok.Type = FOR
	case "in":
		tok.Type = IN
	case "macro":
		tok.Type = MACRO
	case "quote":
		tok.Type = QUOTE
	case "match":
		tok.Type = MATCH
	case "when":
		tok.Type = WHEN
	case "async":
		tok.Type = ASYNC
	case "await":
		tok.Type = AWAIT
	case "import":
		tok.Type = IMPORT
	case "from":
		tok.Type = FROM
	case "as":
		tok.Type = AS
	case "module":
		tok.Type = MODULE
	case "const":
		tok.Type = CONST
	case "assert":
		tok.Type = ASSERT
	case "break":
		tok.Type = BREAK
	case "continue":
		tok.Type = CONTINUE
	default:
		tok.Type = IDENT
	}

	return tok
}

// readNumber reads a number (integer or float)
func (l *Lexer) readNumber() Token {
	tok := Token{Line: l.line, Column: l.column}
	position := l.position

	// Read digits
	for unicode.IsDigit(rune(l.ch)) {
		l.readChar()
	}

	// Check for decimal point
	if l.ch == '.' && unicode.IsDigit(rune(l.peekChar())) {
		tok.Type = FLOAT
		l.readChar() // consume '.'
		for unicode.IsDigit(rune(l.ch)) {
			l.readChar()
		}
	} else {
		tok.Type = INTEGER
	}

	tok.Literal = l.input[position:l.position]
	return tok
}

// readString reads a string literal
func (l *Lexer) readString() Token {
	startLine := l.line
	startColumn := l.column
	quote := l.ch
	l.readChar() // consume opening quote

	var builder strings.Builder
	for l.ch != quote && l.ch != 0 && l.ch != '\n' {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				builder.WriteByte('\n')
			case 't':
				builder.WriteByte('\t')
			case 'r':
				builder.WriteByte('\r')
			case '"':
				builder.WriteByte('"')
			case '\'':
				builder.WriteByte('\'')
			case '\\':
				builder.WriteByte('\\')
			default:
				builder.WriteByte(l.ch)
			}
			l.readChar()
		} else {
			builder.WriteByte(l.ch)
			l.readChar()
		}
	}

	// Check for unterminated string
	if l.ch != quote {
		// Return an ILLEGAL token to signal an error
		return Token{
			Type:    ILLEGAL,
			Literal: fmt.Sprintf("unterminated_string:%c", quote),
			Line:    startLine,
			Column:  startColumn,
		}
	}

	l.readChar() // consume closing quote

	return Token{
		Type:    STRING,
		Literal: builder.String(),
		Line:    startLine,
		Column:  startColumn,
	}
}

// skipWhitespaceExceptNewlines skips spaces and tabs but not newlines
func (l *Lexer) skipWhitespaceExceptNewlines() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// skipComment skips a comment line
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// isIdentifierStart checks if a character can start an identifier
func isIdentifierStart(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

// isIdentifierChar checks if a character can be part of an identifier
func isIdentifierChar(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_'
}
