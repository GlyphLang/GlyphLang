package parser

import (
	"fmt"
	"strings"
	"unicode"
)

// ExpandedLexer tokenizes .glyphx source code (human-readable expanded syntax)
// It recognizes keywords like "route", "let", "return" and converts them to
// the same token types as their symbol equivalents (@, $, >).
type ExpandedLexer struct {
	input             string
	position          int
	readPosition      int
	ch                byte
	line              int
	column            int
	lastTokenWasValue bool
	lastTokenLiteral  string
}

// NewExpandedLexer creates a new ExpandedLexer for .glyphx files
func NewExpandedLexer(input string) *ExpandedLexer {
	l := &ExpandedLexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

func (l *ExpandedLexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

func (l *ExpandedLexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// Tokenize returns all tokens from the expanded syntax input
func (l *ExpandedLexer) Tokenize() ([]Token, error) {
	var tokens []Token

	for {
		l.skipWhitespaceExceptNewlines()

		if l.ch == 0 {
			break
		}

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

		// Skip comments
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
			if strings.HasPrefix(tok.Literal, "unterminated_string:") {
				quote := tok.Literal[len("unterminated_string:"):]
				return nil, fmt.Errorf("unterminated string at line %d, column %d", tok.Line, tok.Column)
				_ = quote // silence unused variable
			}
			return nil, fmt.Errorf("invalid character '%c' at line %d, column %d", tok.Literal[0], tok.Line, tok.Column)
		}

		l.lastTokenWasValue = tok.Type == IDENT || tok.Type == INTEGER ||
			tok.Type == FLOAT || tok.Type == STRING || tok.Type == RPAREN ||
			tok.Type == RBRACKET || tok.Type == TRUE || tok.Type == FALSE
		l.lastTokenLiteral = tok.Literal

		tokens = append(tokens, tok)

		if tok.Type == EOF {
			break
		}
	}

	if len(tokens) == 0 || tokens[len(tokens)-1].Type != EOF {
		tokens = append(tokens, Token{Type: EOF, Line: l.line, Column: l.column})
	}

	return tokens, nil
}

func (l *ExpandedLexer) nextToken() Token {
	var tok Token
	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case 0:
		tok.Type = EOF
		tok.Literal = ""
	case ':':
		tok.Type = COLON
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
	case '/':
		nextChar := l.peekChar()
		isPathContext := l.lastTokenLiteral == "route" || !l.lastTokenWasValue
		isPathStart := unicode.IsLetter(rune(nextChar))

		if isPathContext && isPathStart {
			tok = l.readIdentifier()
		} else {
			tok.Type = SLASH
			tok.Literal = string(l.ch)
			l.readChar()
		}
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
		if l.peekChar() == '.' {
			l.readChar()
			if l.peekChar() == '.' {
				l.readChar()
				tok.Type = DOTDOTDOT
				tok.Literal = "..."
				l.readChar()
			} else {
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
		} else if l.peekChar() == '-' {
			// Handle -- for flags like --formal
			l.readChar() // consume first -
			l.readChar() // consume second -
			// Now read the flag name
			flagTok := l.readIdentifier()
			tok.Type = flagTok.Type
			tok.Literal = "--" + flagTok.Literal
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

func (l *ExpandedLexer) readIdentifier() Token {
	tok := Token{Line: l.line, Column: l.column}
	position := l.position

	// Handle paths like /api/users/:id
	if l.ch == '/' {
		for l.ch == '/' || isIdentifierChar(l.ch) || l.ch == ':' || l.ch == '-' {
			l.readChar()
		}
	} else {
		for isIdentifierChar(l.ch) {
			l.readChar()
		}
	}

	tok.Literal = l.input[position:l.position]

	// Check for expanded keywords first (these map to symbols)
	switch tok.Literal {
	// Expanded keywords -> symbol tokens
	case "route":
		tok.Type = AT
		tok.Literal = "route" // Keep literal for path detection
	case "type":
		tok.Type = COLON
	case "let":
		tok.Type = DOLLAR
	case "return":
		tok.Type = GREATER
	case "middleware":
		tok.Type = PLUS
	case "inject":
		tok.Type = PERCENT
	case "expects":
		tok.Type = LESS
	case "validate":
		tok.Type = QUESTION
	case "handle":
		// "handle" maps to ~ for event handlers
		// Note: "event" is NOT a keyword because it conflicts with the
		// built-in "event" variable that holds event data in handlers
		tok.Type = TILDE
	case "cron":
		tok.Type = STAR
	case "command":
		tok.Type = BANG
	case "queue":
		tok.Type = AMPERSAND
	case "func":
		tok.Type = EQUALS
		tok.Literal = "func"

	// Standard keywords (same as compact lexer)
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
	default:
		tok.Type = IDENT
	}

	return tok
}

func (l *ExpandedLexer) readNumber() Token {
	tok := Token{Line: l.line, Column: l.column}
	position := l.position

	for unicode.IsDigit(rune(l.ch)) {
		l.readChar()
	}

	if l.ch == '.' && unicode.IsDigit(rune(l.peekChar())) {
		tok.Type = FLOAT
		l.readChar()
		for unicode.IsDigit(rune(l.ch)) {
			l.readChar()
		}
	} else {
		tok.Type = INTEGER
	}

	tok.Literal = l.input[position:l.position]
	return tok
}

func (l *ExpandedLexer) readString() Token {
	startLine := l.line
	startColumn := l.column
	quote := l.ch
	l.readChar()

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

	if l.ch != quote {
		return Token{
			Type:    ILLEGAL,
			Literal: fmt.Sprintf("unterminated_string:%c", quote),
			Line:    startLine,
			Column:  startColumn,
		}
	}

	l.readChar()

	return Token{
		Type:    STRING,
		Literal: builder.String(),
		Line:    startLine,
		Column:  startColumn,
	}
}

func (l *ExpandedLexer) skipWhitespaceExceptNewlines() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *ExpandedLexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}
