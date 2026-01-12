package parser

import (
	"testing"
)

func TestExpandedLexerKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"route", AT},
		{"type", COLON},
		{"let", DOLLAR},
		{"return", GREATER},
		{"middleware", PLUS},
		{"inject", PERCENT},
		{"expects", LESS},
		{"validate", QUESTION},
		{"event", TILDE},
		{"cron", STAR},
		{"command", BANG},
		{"queue", AMPERSAND},
		{"func", EQUALS},
	}

	for _, tt := range tests {
		lexer := NewExpandedLexer(tt.input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Errorf("Tokenize(%q) error: %v", tt.input, err)
			continue
		}

		if len(tokens) < 1 {
			t.Errorf("Tokenize(%q) returned no tokens", tt.input)
			continue
		}

		if tokens[0].Type != tt.expected {
			t.Errorf("Tokenize(%q) got token type %v, want %v", tt.input, tokens[0].Type, tt.expected)
		}
	}
}

func TestExpandedLexerRouteDefinition(t *testing.T) {
	input := `route GET /api/users {
  let users = []
  return users
}`

	lexer := NewExpandedLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	// Verify key tokens are present with correct types
	// First token should be AT (route keyword)
	if tokens[0].Type != AT {
		t.Errorf("First token should be AT (route), got %v", tokens[0].Type)
	}

	// Second token should be IDENT (GET)
	if tokens[1].Type != IDENT || tokens[1].Literal != "GET" {
		t.Errorf("Second token should be IDENT 'GET', got %v %q", tokens[1].Type, tokens[1].Literal)
	}

	// Find DOLLAR (let) and GREATER (return) tokens
	foundLet := false
	foundReturn := false
	for _, tok := range tokens {
		if tok.Type == DOLLAR {
			foundLet = true
		}
		if tok.Type == GREATER {
			foundReturn = true
		}
	}

	if !foundLet {
		t.Error("Expected to find DOLLAR token for 'let'")
	}
	if !foundReturn {
		t.Error("Expected to find GREATER token for 'return'")
	}

	// Last non-EOF token should be RBRACE
	if tokens[len(tokens)-2].Type != RBRACE {
		t.Errorf("Second-to-last token should be RBRACE, got %v", tokens[len(tokens)-2].Type)
	}
}

func TestExpandedLexerTypeDefinition(t *testing.T) {
	input := `type User {
  id: int!
  name: str!
}`

	lexer := NewExpandedLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	// First token should be COLON (type keyword)
	if tokens[0].Type != COLON {
		t.Errorf("First token should be COLON (type), got %v", tokens[0].Type)
	}

	// Second token should be IDENT (User)
	if tokens[1].Type != IDENT || tokens[1].Literal != "User" {
		t.Errorf("Second token should be IDENT 'User', got %v %q", tokens[1].Type, tokens[1].Literal)
	}
}

func TestExpandedLexerCommand(t *testing.T) {
	input := `command hello name: str! {
  return {message: "Hello"}
}`

	lexer := NewExpandedLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	// First token should be BANG (command keyword)
	if tokens[0].Type != BANG {
		t.Errorf("First token should be BANG (command), got %v", tokens[0].Type)
	}
}

func TestExpandedLexerCronTask(t *testing.T) {
	input := `cron "0 0 * * *" daily_task {
  return {done: true}
}`

	lexer := NewExpandedLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	// First token should be STAR (cron keyword)
	if tokens[0].Type != STAR {
		t.Errorf("First token should be STAR (cron), got %v", tokens[0].Type)
	}

	// Second token should be STRING (schedule)
	if tokens[1].Type != STRING || tokens[1].Literal != "0 0 * * *" {
		t.Errorf("Second token should be STRING '0 0 * * *', got %v %q", tokens[1].Type, tokens[1].Literal)
	}
}

func TestExpandedLexerEventHandler(t *testing.T) {
	input := `event "user.created" async {
  return {handled: true}
}`

	lexer := NewExpandedLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	// First token should be TILDE (event keyword)
	if tokens[0].Type != TILDE {
		t.Errorf("First token should be TILDE (event), got %v", tokens[0].Type)
	}
}

func TestExpandedLexerQueueWorker(t *testing.T) {
	input := `queue "email.send" {
  middleware concurrency(5)
  return {sent: true}
}`

	lexer := NewExpandedLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	// First token should be AMPERSAND (queue keyword)
	if tokens[0].Type != AMPERSAND {
		t.Errorf("First token should be AMPERSAND (queue), got %v", tokens[0].Type)
	}
}

func TestExpandedLexerMiddlewareAndInject(t *testing.T) {
	input := `route GET /api/admin {
  middleware auth(jwt)
  inject db: Database
  return {}
}`

	lexer := NewExpandedLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	// Find middleware token (should be PLUS)
	foundMiddleware := false
	foundInject := false
	for _, tok := range tokens {
		if tok.Type == PLUS && tok.Literal == "middleware" {
			foundMiddleware = true
		}
		if tok.Type == PERCENT && tok.Literal == "inject" {
			foundInject = true
		}
	}

	if !foundMiddleware {
		t.Error("Expected to find PLUS token for 'middleware'")
	}
	if !foundInject {
		t.Error("Expected to find PERCENT token for 'inject'")
	}
}

func TestExpandedLexerValidate(t *testing.T) {
	input := `validate checkEmail(email)`

	lexer := NewExpandedLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	// First token should be QUESTION (validate keyword)
	if tokens[0].Type != QUESTION {
		t.Errorf("First token should be QUESTION (validate), got %v", tokens[0].Type)
	}
}

func TestExpandedLexerStandardKeywords(t *testing.T) {
	// These keywords should remain unchanged from compact lexer
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"true", TRUE},
		{"false", FALSE},
		{"null", NULL},
		{"while", WHILE},
		{"switch", SWITCH},
		{"case", CASE},
		{"default", DEFAULT},
		{"for", FOR},
		{"in", IN},
		{"if", IDENT}, // 'if' is not a keyword token, handled by parser
		{"else", IDENT},
		{"async", ASYNC},
		{"await", AWAIT},
		{"import", IMPORT},
		{"from", FROM},
		{"as", AS},
		{"module", MODULE},
	}

	for _, tt := range tests {
		lexer := NewExpandedLexer(tt.input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Errorf("Tokenize(%q) error: %v", tt.input, err)
			continue
		}

		if len(tokens) < 1 {
			t.Errorf("Tokenize(%q) returned no tokens", tt.input)
			continue
		}

		if tokens[0].Type != tt.expected {
			t.Errorf("Tokenize(%q) got token type %v, want %v", tt.input, tokens[0].Type, tt.expected)
		}
	}
}

func TestExpandedLexerNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
		literal  string
	}{
		{"42", INTEGER, "42"},
		{"3.14", FLOAT, "3.14"},
		{"0", INTEGER, "0"},
		{"100", INTEGER, "100"},
	}

	for _, tt := range tests {
		lexer := NewExpandedLexer(tt.input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Errorf("Tokenize(%q) error: %v", tt.input, err)
			continue
		}

		if tokens[0].Type != tt.expected {
			t.Errorf("Tokenize(%q) got type %v, want %v", tt.input, tokens[0].Type, tt.expected)
		}
		if tokens[0].Literal != tt.literal {
			t.Errorf("Tokenize(%q) got literal %q, want %q", tt.input, tokens[0].Literal, tt.literal)
		}
	}
}

func TestExpandedLexerStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello world"`, "hello world"},
		{`"escaped\nline"`, "escaped\nline"},
		{`'single quotes'`, "single quotes"},
	}

	for _, tt := range tests {
		lexer := NewExpandedLexer(tt.input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Errorf("Tokenize(%q) error: %v", tt.input, err)
			continue
		}

		if tokens[0].Type != STRING {
			t.Errorf("Tokenize(%q) got type %v, want STRING", tt.input, tokens[0].Type)
		}
		if tokens[0].Literal != tt.expected {
			t.Errorf("Tokenize(%q) got literal %q, want %q", tt.input, tokens[0].Literal, tt.expected)
		}
	}
}

func TestExpandedLexerComments(t *testing.T) {
	input := `route GET /test {
  # This is a comment
  let x = 1
  // Another comment style
  return x
}`

	lexer := NewExpandedLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	// Comments should be skipped, check that we don't have comment content as tokens
	for _, tok := range tokens {
		if tok.Type == IDENT && (tok.Literal == "This" || tok.Literal == "Another") {
			t.Error("Comment content should not appear as tokens")
		}
	}
}

func TestExpandedLexerOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"==", EQ_EQ},
		{"!=", NOT_EQ},
		{"<=", LESS_EQ},
		{">=", GREATER_EQ},
		{"&&", AND},
		{"||", OR},
		{"->", ARROW},
		{"=>", FATARROW},
		{"+", PLUS},
		{"-", MINUS},
		{"*", STAR},
		{"/", SLASH},
	}

	for _, tt := range tests {
		lexer := NewExpandedLexer(tt.input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Errorf("Tokenize(%q) error: %v", tt.input, err)
			continue
		}

		if tokens[0].Type != tt.expected {
			t.Errorf("Tokenize(%q) got type %v, want %v", tt.input, tokens[0].Type, tt.expected)
		}
	}
}
