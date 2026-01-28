package parser

// TokenType represents the type of token
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	NEWLINE

	// Symbols
	AT         // @
	COLON      // :
	DOLLAR     // $
	PLUS       // +
	MINUS      // -
	STAR       // *
	SLASH      // /
	PERCENT    // %
	GREATER    // >
	GREATER_EQ // >=
	LESS       // <
	LESS_EQ    // <=
	BANG       // !
	NOT_EQ     // !=
	EQ_EQ      // ==
	QUESTION   // ?
	TILDE      // ~
	AMPERSAND  // &
	AND        // &&
	OR         // ||

	// Delimiters
	LPAREN   // (
	RPAREN   // )
	LBRACE   // {
	RBRACE   // }
	LBRACKET // [
	RBRACKET // ]
	COMMA    // ,
	DOT      // .
	ARROW    // ->
	PIPE     // |
	PIPE_OP  // |>
	EQUALS   // =

	// Literals
	IDENT   // identifier or path
	STRING  // "string"
	INTEGER // 123
	FLOAT   // 123.45
	TRUE    // true
	FALSE   // false
	NULL    // null
	WHILE   // while

	// Keywords
	SWITCH   // switch
	CASE     // case
	DEFAULT  // default
	FOR      // for
	IN       // in
	MACRO    // macro
	QUOTE    // quote
	MATCH     // match
	WHEN      // when (for guards in match)
	FATARROW  // =>
	DOTDOTDOT // ...
	ASYNC     // async
	AWAIT     // await
	IMPORT    // import
	FROM      // from
	AS        // as
	MODULE    // module
)

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// String returns a string representation of the token type
func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case NEWLINE:
		return "NEWLINE"
	case AT:
		return "@"
	case COLON:
		return ":"
	case DOLLAR:
		return "$"
	case PLUS:
		return "+"
	case MINUS:
		return "-"
	case STAR:
		return "*"
	case SLASH:
		return "/"
	case PERCENT:
		return "%"
	case GREATER:
		return ">"
	case GREATER_EQ:
		return ">="
	case LESS:
		return "<"
	case LESS_EQ:
		return "<="
	case BANG:
		return "!"
	case NOT_EQ:
		return "!="
	case EQ_EQ:
		return "=="
	case QUESTION:
		return "?"
	case TILDE:
		return "~"
	case AMPERSAND:
		return "&"
	case AND:
		return "&&"
	case OR:
		return "||"
	case LPAREN:
		return "("
	case RPAREN:
		return ")"
	case LBRACE:
		return "{"
	case RBRACE:
		return "}"
	case LBRACKET:
		return "["
	case RBRACKET:
		return "]"
	case COMMA:
		return ","
	case DOT:
		return "."
	case ARROW:
		return "->"
	case PIPE:
		return "|"
	case PIPE_OP:
		return "|>"
	case EQUALS:
		return "="
	case IDENT:
		return "IDENT"
	case STRING:
		return "STRING"
	case INTEGER:
		return "INTEGER"
	case FLOAT:
		return "FLOAT"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	case NULL:
		return "NULL"
	case WHILE:
		return "WHILE"
	case SWITCH:
		return "SWITCH"
	case CASE:
		return "CASE"
	case DEFAULT:
		return "DEFAULT"
	case FOR:
		return "FOR"
	case IN:
		return "IN"
	case MACRO:
		return "MACRO"
	case QUOTE:
		return "QUOTE"
	case MATCH:
		return "MATCH"
	case WHEN:
		return "WHEN"
	case FATARROW:
		return "=>"
	case DOTDOTDOT:
		return "..."
	case ASYNC:
		return "ASYNC"
	case AWAIT:
		return "AWAIT"
	case IMPORT:
		return "IMPORT"
	case FROM:
		return "FROM"
	case AS:
		return "AS"
	case MODULE:
		return "MODULE"
	default:
		return "UNKNOWN"
	}
}
