package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test token type string representation
func TestTokenType_String(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		{ILLEGAL, "ILLEGAL"},
		{EOF, "EOF"},
		{NEWLINE, "NEWLINE"},
		{AT, "@"},
		{COLON, ":"},
		{DOLLAR, "$"},
		{PLUS, "+"},
		{MINUS, "-"},
		{STAR, "*"},
		{SLASH, "/"},
		{PERCENT, "%"},
		{GREATER, ">"},
		{GREATER_EQ, ">="},
		{LESS, "<"},
		{LESS_EQ, "<="},
		{BANG, "!"},
		{NOT_EQ, "!="},
		{EQ_EQ, "=="},
		{TILDE, "~"},
		{AMPERSAND, "&"},
		{AND, "&&"},
		{OR, "||"},
		{LPAREN, "("},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{RBRACE, "}"},
		{LBRACKET, "["},
		{RBRACKET, "]"},
		{COMMA, ","},
		{DOT, "."},
		{ARROW, "->"},
		{PIPE, "|"},
		{EQUALS, "="},
		{IDENT, "IDENT"},
		{STRING, "STRING"},
		{INTEGER, "INTEGER"},
		{FLOAT, "FLOAT"},
		{TRUE, "TRUE"},
		{FALSE, "FALSE"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.tokenType.String())
		})
	}
}

// Test token creation
func TestToken_Creation(t *testing.T) {
	token := Token{
		Type:    IDENT,
		Literal: "hello",
		Line:    1,
		Column:  5,
	}

	assert.Equal(t, IDENT, token.Type)
	assert.Equal(t, "hello", token.Literal)
	assert.Equal(t, 1, token.Line)
	assert.Equal(t, 5, token.Column)
}

// Test all token types are defined
func TestTokenType_AllTypesDefined(t *testing.T) {
	// This test ensures we don't accidentally skip token type values
	allTypes := []TokenType{
		ILLEGAL, EOF, NEWLINE,
		AT, COLON, DOLLAR, PLUS, MINUS, STAR, SLASH, PERCENT,
		GREATER, GREATER_EQ, LESS, LESS_EQ,
		BANG, NOT_EQ, EQ_EQ,
		TILDE, AMPERSAND, AND, OR,
		LPAREN, RPAREN, LBRACE, RBRACE, LBRACKET, RBRACKET,
		COMMA, DOT, ARROW, PIPE, EQUALS,
		IDENT, STRING, INTEGER, FLOAT, TRUE, FALSE,
	}

	for _, tokenType := range allTypes {
		str := tokenType.String()
		assert.NotEqual(t, "UNKNOWN", str, "token type %d should have a string representation", tokenType)
	}
}
