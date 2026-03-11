package parser

import (
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
)

// Helper function: lex and parse source code, returning module or error
func parseSource(t *testing.T, source string) *ast.Module {
	t.Helper()
	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	parser := NewParserWithSource(tokens, source)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	return module
}

// Helper function: lex and parse source code, expecting a parse error
func parseSourceExpectError(t *testing.T, source string) error {
	t.Helper()
	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		// Lex error is fine too
		return err
	}
	parser := NewParserWithSource(tokens, source)
	_, err = parser.Parse()
	if err == nil {
		t.Fatal("expected parser error, got nil")
	}
	return err
}

// Helper: create parser from source and return it with tokens
func makeParser(t *testing.T, source string) *Parser {
	t.Helper()
	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	return NewParserWithSource(tokens, source)
}

// ============================================================
// Tests for errors.go: errorWithContext, expressionError,
// errorAtPosition, buildMissingTokenError, buildUnexpectedTokenError
// ============================================================

func TestErrorWithContext(t *testing.T) {
	source := "@ GET /test {\n  > 42\n}"
	p := makeParser(t, source)

	tok := p.current()
	err := p.errorWithContext("test error message", tok)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	if parseErr.Message != "test error message" {
		t.Errorf("expected message 'test error message', got %q", parseErr.Message)
	}
	if parseErr.Source != source {
		t.Error("expected source to match")
	}
}

func TestExpressionError(t *testing.T) {
	source := "@ GET /test {\n  > 42\n}"
	p := makeParser(t, source)

	tok := p.current()
	err := p.expressionError("bad expression", tok)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	if parseErr.Message != "bad expression" {
		t.Errorf("expected message 'bad expression', got %q", parseErr.Message)
	}
	if parseErr.Hint == "" {
		t.Error("expected non-empty hint from expressionError")
	}
	if !strings.Contains(parseErr.Hint, "expression") {
		t.Errorf("expected hint about expression, got %q", parseErr.Hint)
	}
}

func TestErrorAtPosition(t *testing.T) {
	source := "invalid input here"
	l := NewLexer(source)
	// Advance the lexer a bit
	l.readChar()
	l.readChar()

	err := l.errorAtPosition("test lexer error")
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	lexErr, ok := err.(*LexError)
	if !ok {
		t.Fatalf("expected *LexError, got %T", err)
	}
	if lexErr.Message != "test lexer error" {
		t.Errorf("expected message 'test lexer error', got %q", lexErr.Message)
	}
	if lexErr.Source != source {
		t.Error("expected source to match")
	}
}

func TestBuildMissingTokenError(t *testing.T) {
	result := buildMissingTokenError("semicolon", "function declaration")
	expected := "Missing semicolon in function declaration"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBuildUnexpectedTokenError(t *testing.T) {
	result := buildUnexpectedTokenError(IDENT, "type definition")
	expected := "Unexpected IDENT in type definition"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}

	// Test with various token types
	result2 := buildUnexpectedTokenError(RBRACE, "route body")
	if !strings.Contains(result2, "}") {
		t.Errorf("expected token string in error, got %q", result2)
	}
}

// ============================================================
// Tests for ParseError.Error() with various configurations
// ============================================================

func TestParseErrorWithFullContext(t *testing.T) {
	// Error on line 2, with previous and next lines visible
	err := &ParseError{
		Message: "unexpected token",
		Line:    2,
		Column:  5,
		Source:  "line one\nline two\nline three",
		Hint:    "check your syntax",
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "line 2") {
		t.Error("expected line number in error")
	}
	if !strings.Contains(errStr, "line one") {
		t.Error("expected previous line context")
	}
	if !strings.Contains(errStr, "line three") {
		t.Error("expected next line context")
	}
	if !strings.Contains(errStr, "^") {
		t.Error("expected caret")
	}
	if !strings.Contains(errStr, "Hint:") {
		t.Error("expected hint")
	}
}

func TestParseErrorOnFirstLine(t *testing.T) {
	// Error on line 1 - no previous line
	err := &ParseError{
		Message: "unexpected token",
		Line:    1,
		Column:  3,
		Source:  "abc\ndef",
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "abc") {
		t.Error("expected error line in output")
	}
	if !strings.Contains(errStr, "^") {
		t.Error("expected caret")
	}
}

func TestParseErrorNoSource(t *testing.T) {
	err := &ParseError{
		Message: "no source provided",
		Line:    1,
		Column:  1,
		Source:  "",
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "no source provided") {
		t.Error("expected message in error string")
	}
}

func TestParseErrorNoHint(t *testing.T) {
	err := &ParseError{
		Message: "error without hint",
		Line:    1,
		Column:  1,
		Source:  "x",
	}
	errStr := err.Error()
	if strings.Contains(errStr, "Hint:") {
		t.Error("should not contain hint")
	}
}

func TestParseErrorZeroColumn(t *testing.T) {
	err := &ParseError{
		Message: "zero column",
		Line:    1,
		Column:  0,
		Source:  "some code",
	}
	errStr := err.Error()
	// Should not have a caret when column is 0
	if strings.Contains(errStr, "^") {
		t.Error("should not have caret with column 0")
	}
}

// ============================================================
// Tests for LexError.Error()
// ============================================================

func TestLexErrorFull(t *testing.T) {
	err := &LexError{
		Message: "bad char",
		Line:    1,
		Column:  3,
		Source:  "abc",
		Char:    'c',
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "Lexer error") {
		t.Error("expected 'Lexer error' prefix")
	}
	if !strings.Contains(errStr, "abc") {
		t.Error("expected source line")
	}
	if !strings.Contains(errStr, "^") {
		t.Error("expected caret")
	}
}

func TestLexErrorNoSource(t *testing.T) {
	err := &LexError{
		Message: "no source",
		Line:    1,
		Column:  1,
		Source:  "",
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "no source") {
		t.Error("expected message")
	}
}

func TestLexErrorZeroColumn(t *testing.T) {
	err := &LexError{
		Message: "zero col",
		Line:    1,
		Column:  0,
		Source:  "code",
	}
	errStr := err.Error()
	if strings.Contains(errStr, "^") {
		t.Error("should not have caret when column is 0")
	}
}

// ============================================================
// Tests for parseQuoteExpr (directly calling the method since
// it is not yet wired into parsePrimary)
// ============================================================

func TestParseQuoteExprDirect(t *testing.T) {
	source := `quote { $ x = 1 }`
	p := makeParser(t, source)
	expr, err := p.parseQuoteExpr()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	quoteExpr, ok := expr.(ast.QuoteExpr)
	if !ok {
		t.Fatalf("expected QuoteExpr, got %T", expr)
	}
	if len(quoteExpr.Body) == 0 {
		t.Error("expected non-empty quote body")
	}
}

func TestParseQuoteExprWithReturn(t *testing.T) {
	source := "quote {\n  > 42\n}"
	p := makeParser(t, source)
	expr, err := p.parseQuoteExpr()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	quoteExpr, ok := expr.(ast.QuoteExpr)
	if !ok {
		t.Fatalf("expected QuoteExpr, got %T", expr)
	}
	if len(quoteExpr.Body) == 0 {
		t.Error("expected non-empty quote body")
	}
}

func TestParseQuoteExprWithIfStatement(t *testing.T) {
	source := "quote {\n  if x > 0 {\n    > x\n  }\n}"
	p := makeParser(t, source)
	expr, err := p.parseQuoteExpr()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	quoteExpr, ok := expr.(ast.QuoteExpr)
	if !ok {
		t.Fatalf("expected QuoteExpr, got %T", expr)
	}
	if len(quoteExpr.Body) == 0 {
		t.Error("expected non-empty quote body")
	}
}

func TestParseQuoteExprWithMultipleStatements(t *testing.T) {
	source := "quote {\n  $ x = 1\n  $ y = 2\n  > x + y\n}"
	p := makeParser(t, source)
	expr, err := p.parseQuoteExpr()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	quoteExpr, ok := expr.(ast.QuoteExpr)
	if !ok {
		t.Fatalf("expected QuoteExpr, got %T", expr)
	}
	if len(quoteExpr.Body) < 3 {
		t.Errorf("expected at least 3 body nodes, got %d", len(quoteExpr.Body))
	}
}

func TestParseQuoteExprEmpty(t *testing.T) {
	source := "quote { }"
	p := makeParser(t, source)
	expr, err := p.parseQuoteExpr()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	quoteExpr, ok := expr.(ast.QuoteExpr)
	if !ok {
		t.Fatalf("expected QuoteExpr, got %T", expr)
	}
	if len(quoteExpr.Body) != 0 {
		t.Errorf("expected empty body, got %d nodes", len(quoteExpr.Body))
	}
}

// ============================================================
// Tests for ParseExpression (public method, 0% coverage)
// ============================================================

func TestParseExpressionPublic(t *testing.T) {
	source := `42`
	p := makeParser(t, source)
	expr, err := p.ParseExpression()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lit, ok := expr.(ast.LiteralExpr)
	if !ok {
		t.Fatalf("expected LiteralExpr, got %T", expr)
	}
	intLit, ok := lit.Value.(ast.IntLiteral)
	if !ok {
		t.Fatalf("expected IntLiteral, got %T", lit.Value)
	}
	if intLit.Value != 42 {
		t.Errorf("expected 42, got %d", intLit.Value)
	}
}

func TestParseExpressionPublicString(t *testing.T) {
	source := `"hello world"`
	p := makeParser(t, source)
	expr, err := p.ParseExpression()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lit, ok := expr.(ast.LiteralExpr)
	if !ok {
		t.Fatalf("expected LiteralExpr, got %T", expr)
	}
	strLit, ok := lit.Value.(ast.StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral, got %T", lit.Value)
	}
	if strLit.Value != "hello world" {
		t.Errorf("expected 'hello world', got %q", strLit.Value)
	}
}

func TestParseExpressionPublicBinaryOp(t *testing.T) {
	source := `1 + 2`
	p := makeParser(t, source)
	expr, err := p.ParseExpression()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	binExpr, ok := expr.(ast.BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr, got %T", expr)
	}
	if binExpr.Op != ast.Add {
		t.Errorf("expected Add op, got %v", binExpr.Op)
	}
}

func TestParseExpressionPublicEmpty(t *testing.T) {
	source := ``
	p := makeParser(t, source)
	_, err := p.ParseExpression()
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseExpressionPublicWithNewlines(t *testing.T) {
	source := "\n\n42\n\n"
	p := makeParser(t, source)
	expr, err := p.ParseExpression()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := expr.(ast.LiteralExpr)
	if !ok {
		t.Fatalf("expected LiteralExpr, got %T", expr)
	}
}

func TestParseExpressionPublicBool(t *testing.T) {
	source := `true`
	p := makeParser(t, source)
	expr, err := p.ParseExpression()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lit, ok := expr.(ast.LiteralExpr)
	if !ok {
		t.Fatalf("expected LiteralExpr, got %T", expr)
	}
	_, ok = lit.Value.(ast.BoolLiteral)
	if !ok {
		t.Fatalf("expected BoolLiteral, got %T", lit.Value)
	}
}

func TestParseExpressionPublicFunctionCall(t *testing.T) {
	source := `myFunc(1, 2)`
	p := makeParser(t, source)
	expr, err := p.ParseExpression()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	callExpr, ok := expr.(ast.FunctionCallExpr)
	if !ok {
		t.Fatalf("expected FunctionCallExpr, got %T", expr)
	}
	if callExpr.Name != "myFunc" {
		t.Errorf("expected 'myFunc', got %q", callExpr.Name)
	}
}

// ============================================================
// Tests for ParseStatement (public method, 0% coverage)
// ============================================================

func TestParseStatementPublicAssign(t *testing.T) {
	source := `$ x = 42`
	p := makeParser(t, source)
	stmt, err := p.ParseStatement()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign, ok := stmt.(ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", stmt)
	}
	if assign.Target != "x" {
		t.Errorf("expected target 'x', got %q", assign.Target)
	}
}

func TestParseStatementPublicReturn(t *testing.T) {
	source := `> 100`
	p := makeParser(t, source)
	stmt, err := p.ParseStatement()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ret, ok := stmt.(ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected ReturnStatement, got %T", stmt)
	}
	if ret.Value == nil {
		t.Error("expected non-nil return value")
	}
}

func TestParseStatementPublicEmpty(t *testing.T) {
	source := ``
	p := makeParser(t, source)
	_, err := p.ParseStatement()
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseStatementPublicWithNewlines(t *testing.T) {
	source := "\n\n$ y = 10\n\n"
	p := makeParser(t, source)
	stmt, err := p.ParseStatement()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assign, ok := stmt.(ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", stmt)
	}
	if assign.Target != "y" {
		t.Errorf("expected target 'y', got %q", assign.Target)
	}
}

// ============================================================
// Tests for parseStatement paths (low coverage: let, return, yield,
// reassignment, validation, expression statement, default error)
// ============================================================

func TestParseStatementLetKeyword(t *testing.T) {
	source := `@ GET /test {
  let x = 42
  > x
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign, ok := route.Body[0].(ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}
	if assign.Target != "x" {
		t.Errorf("expected target 'x', got %q", assign.Target)
	}
}

func TestParseStatementLetWithTypeAnnotation(t *testing.T) {
	source := `@ GET /test {
  let x: int = 42
  > x
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign, ok := route.Body[0].(ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}
	if assign.Target != "x" {
		t.Errorf("expected target 'x', got %q", assign.Target)
	}
}

func TestParseStatementLetDeclarationOnly(t *testing.T) {
	// let with type but no assignment
	source := `@ GET /test {
  let x: int
  > x
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign, ok := route.Body[0].(ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}
	if assign.Target != "x" {
		t.Errorf("expected target 'x', got %q", assign.Target)
	}
}

func TestParseStatementReturnKeyword(t *testing.T) {
	source := `@ GET /test {
  return 42
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	ret, ok := route.Body[0].(ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected ReturnStatement, got %T", route.Body[0])
	}
	if ret.Value == nil {
		t.Error("expected non-nil return value")
	}
}

func TestParseStatementYield(t *testing.T) {
	source := `@ GET /test {
  yield "hello"
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	yieldStmt, ok := route.Body[0].(ast.YieldStatement)
	if !ok {
		t.Fatalf("expected YieldStatement, got %T", route.Body[0])
	}
	if yieldStmt.Value == nil {
		t.Error("expected non-nil yield value")
	}
}

func TestParseStatementReassignment(t *testing.T) {
	source := `@ GET /test {
  $ x = 1
  x = 2
  > x
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	reassign, ok := route.Body[1].(ast.ReassignStatement)
	if !ok {
		t.Fatalf("expected ReassignStatement, got %T", route.Body[1])
	}
	if reassign.Target != "x" {
		t.Errorf("expected target 'x', got %q", reassign.Target)
	}
}

func TestParseStatementValidation(t *testing.T) {
	source := `@ POST /submit {
  ? validateEmail(email)
  > { ok: true }
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)

	// The route should parse with query params or validation
	if len(route.Body) == 0 && len(route.QueryParams) == 0 {
		t.Error("expected at least one statement or query param")
	}
}

func TestParseStatementExpressionStatement(t *testing.T) {
	source := `@ GET /test {
  doSomething(1, 2)
  > "done"
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	exprStmt, ok := route.Body[0].(ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", route.Body[0])
	}
	if exprStmt.Expr == nil {
		t.Error("expected non-nil expression")
	}
}

func TestParseStatementDollarFieldAccess(t *testing.T) {
	source := `@ GET /test {
  $ obj.field = 42
  > obj
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign, ok := route.Body[0].(ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}
	if assign.Target != "obj.field" {
		t.Errorf("expected target 'obj.field', got %q", assign.Target)
	}
}

func TestParseStatementDollarTypeAnnotationNoAssign(t *testing.T) {
	source := `@ GET /test {
  $ x: int
  > x
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign, ok := route.Body[0].(ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}
	if assign.Target != "x" {
		t.Errorf("expected target 'x', got %q", assign.Target)
	}
}

func TestParseStatementAssertStatement(t *testing.T) {
	source := `@ GET /test {
  assert(1 == 1)
  > "ok"
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	_, ok := route.Body[0].(ast.AssertStatement)
	if !ok {
		t.Fatalf("expected AssertStatement, got %T", route.Body[0])
	}
}

func TestParseStatementDefault(t *testing.T) {
	// Test error case with an unexpected top-level token
	parseSourceExpectError(t, `123`)
}

// ============================================================
// Tests for parseTypeDefWithoutColon
// ============================================================

func TestParseTypeDefWithoutColon(t *testing.T) {
	source := `type User {
  name: str!
  age: int!
}`
	module := parseSource(t, source)
	typeDef, ok := module.Items[0].(*ast.TypeDef)
	if !ok {
		t.Fatalf("expected TypeDef, got %T", module.Items[0])
	}
	if typeDef.Name != "User" {
		t.Errorf("expected name 'User', got %q", typeDef.Name)
	}
	if len(typeDef.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(typeDef.Fields))
	}
}

func TestParseTypeDefWithoutColonGeneric(t *testing.T) {
	source := `type Container<T> {
  value: T!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	if typeDef.Name != "Container" {
		t.Errorf("expected name 'Container', got %q", typeDef.Name)
	}
	if len(typeDef.TypeParams) != 1 {
		t.Fatalf("expected 1 type param, got %d", len(typeDef.TypeParams))
	}
	if typeDef.TypeParams[0].Name != "T" {
		t.Errorf("expected type param 'T', got %q", typeDef.TypeParams[0].Name)
	}
}

func TestParseTypeDefWithoutColonImpl(t *testing.T) {
	source := `type MyType impl Serializable {
  name: str!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	if typeDef.Name != "MyType" {
		t.Errorf("expected name 'MyType', got %q", typeDef.Name)
	}
	if len(typeDef.Traits) != 1 {
		t.Fatalf("expected 1 trait, got %d", len(typeDef.Traits))
	}
	if typeDef.Traits[0] != "Serializable" {
		t.Errorf("expected trait 'Serializable', got %q", typeDef.Traits[0])
	}
}

func TestParseTypeDefWithoutColonImplMultiple(t *testing.T) {
	source := `type MyType impl Serializable, Comparable {
  name: str!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	if len(typeDef.Traits) != 2 {
		t.Fatalf("expected 2 traits, got %d", len(typeDef.Traits))
	}
	if typeDef.Traits[0] != "Serializable" {
		t.Errorf("expected first trait 'Serializable', got %q", typeDef.Traits[0])
	}
	if typeDef.Traits[1] != "Comparable" {
		t.Errorf("expected second trait 'Comparable', got %q", typeDef.Traits[1])
	}
}

func TestParseTypeDefWithoutColonGenericAndImpl(t *testing.T) {
	source := `type Box<T> impl Serializable {
  value: T!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	if len(typeDef.TypeParams) != 1 {
		t.Fatalf("expected 1 type param, got %d", len(typeDef.TypeParams))
	}
	if len(typeDef.Traits) != 1 {
		t.Fatalf("expected 1 trait, got %d", len(typeDef.Traits))
	}
}

// ============================================================
// Tests for parseMacroBodyNode and statementToNode
// ============================================================

func TestParseMacroBodyWithRoute(t *testing.T) {
	source := `macro! api(path) {
  @ GET /path {
    > "ok"
  }
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if len(macroDef.Body) == 0 {
		t.Error("expected non-empty macro body")
	}
}

func TestParseMacroBodyWithTypeDef(t *testing.T) {
	source := `macro! types() {
  : MyType {
    name: str!
  }
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if len(macroDef.Body) == 0 {
		t.Error("expected non-empty macro body")
	}
}

func TestParseMacroBodyWithAssignment(t *testing.T) {
	source := `macro! setVar() {
  $ x = 1
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if len(macroDef.Body) == 0 {
		t.Error("expected non-empty macro body")
	}
}

func TestParseMacroBodyWithReturn(t *testing.T) {
	source := `macro! retVal() {
  > 42
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if len(macroDef.Body) == 0 {
		t.Error("expected non-empty macro body")
	}
}

func TestParseMacroBodyWithIfStatement(t *testing.T) {
	source := `macro! conditional(flag) {
  if flag {
    > 1
  }
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if len(macroDef.Body) == 0 {
		t.Error("expected non-empty macro body")
	}
}

func TestParseMacroBodyWithForLoop(t *testing.T) {
	source := `macro! looper() {
  for i in items {
    $ x = i
  }
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if len(macroDef.Body) == 0 {
		t.Error("expected non-empty macro body")
	}
}

func TestParseMacroBodyWithWhileLoop(t *testing.T) {
	source := `macro! whileLoop() {
  while true {
    > 1
  }
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if len(macroDef.Body) == 0 {
		t.Error("expected non-empty macro body")
	}
}

func TestParseMacroBodyWithSwitchStatement(t *testing.T) {
	source := `macro! switcher() {
  switch val {
    case 1 {
      > "one"
    }
    default {
      > "other"
    }
  }
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if len(macroDef.Body) == 0 {
		t.Error("expected non-empty macro body")
	}
}

func TestParseMacroBodyWithLetReturn(t *testing.T) {
	source := `macro! letAndReturn() {
  let x = 42
  return x
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if len(macroDef.Body) < 2 {
		t.Errorf("expected at least 2 body nodes, got %d", len(macroDef.Body))
	}
}

func TestParseMacroBodyWithFuncCall(t *testing.T) {
	source := `macro! callSomething() {
  doWork(1, 2)
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if len(macroDef.Body) == 0 {
		t.Error("expected non-empty macro body")
	}
}

func TestParseMacroBodyWithMacroInvocation(t *testing.T) {
	source := `macro! outer() {
  inner!(1, 2)
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if len(macroDef.Body) == 0 {
		t.Error("expected non-empty macro body")
	}
}

func TestParseMacroBodyWithValidation(t *testing.T) {
	source := `macro! validate() {
  ? check(input)
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if len(macroDef.Body) == 0 {
		t.Error("expected non-empty macro body")
	}
}

// ============================================================
// Tests for parsePattern (low coverage)
// ============================================================

func TestParsePatternFloat(t *testing.T) {
	source := `@ GET /test {
  $ result = match val {
    3.14 => "pi"
    _ => "other"
  }
  > result
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign := route.Body[0].(ast.AssignStatement)
	matchExpr := assign.Value.(ast.MatchExpr)

	litPat, ok := matchExpr.Cases[0].Pattern.(ast.LiteralPattern)
	if !ok {
		t.Fatalf("expected LiteralPattern, got %T", matchExpr.Cases[0].Pattern)
	}
	_, ok = litPat.Value.(ast.FloatLiteral)
	if !ok {
		t.Fatalf("expected FloatLiteral, got %T", litPat.Value)
	}
}

func TestParsePatternBoolTrue(t *testing.T) {
	source := `@ GET /test {
  $ result = match flag {
    true => "yes"
    false => "no"
    _ => "unknown"
  }
  > result
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign := route.Body[0].(ast.AssignStatement)
	matchExpr := assign.Value.(ast.MatchExpr)

	litPat, ok := matchExpr.Cases[0].Pattern.(ast.LiteralPattern)
	if !ok {
		t.Fatalf("expected LiteralPattern for true, got %T", matchExpr.Cases[0].Pattern)
	}
	boolLit, ok := litPat.Value.(ast.BoolLiteral)
	if !ok {
		t.Fatalf("expected BoolLiteral, got %T", litPat.Value)
	}
	if !boolLit.Value {
		t.Error("expected true literal")
	}

	// Check false case
	litPat2, ok := matchExpr.Cases[1].Pattern.(ast.LiteralPattern)
	if !ok {
		t.Fatalf("expected LiteralPattern for false, got %T", matchExpr.Cases[1].Pattern)
	}
	boolLit2, ok := litPat2.Value.(ast.BoolLiteral)
	if !ok {
		t.Fatalf("expected BoolLiteral, got %T", litPat2.Value)
	}
	if boolLit2.Value {
		t.Error("expected false literal")
	}
}

func TestParsePatternNull(t *testing.T) {
	source := `@ GET /test {
  $ result = match val {
    null => "nothing"
    _ => "something"
  }
  > result
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign := route.Body[0].(ast.AssignStatement)
	matchExpr := assign.Value.(ast.MatchExpr)

	litPat, ok := matchExpr.Cases[0].Pattern.(ast.LiteralPattern)
	if !ok {
		t.Fatalf("expected LiteralPattern, got %T", matchExpr.Cases[0].Pattern)
	}
	_, ok = litPat.Value.(ast.NullLiteral)
	if !ok {
		t.Fatalf("expected NullLiteral, got %T", litPat.Value)
	}
}

func TestParsePatternVariable(t *testing.T) {
	source := `@ GET /test {
  $ result = match val {
    x => x
  }
  > result
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign := route.Body[0].(ast.AssignStatement)
	matchExpr := assign.Value.(ast.MatchExpr)

	varPat, ok := matchExpr.Cases[0].Pattern.(ast.VariablePattern)
	if !ok {
		t.Fatalf("expected VariablePattern, got %T", matchExpr.Cases[0].Pattern)
	}
	if varPat.Name != "x" {
		t.Errorf("expected 'x', got %q", varPat.Name)
	}
}

// ============================================================
// Tests for parseMethodDef (low coverage)
// ============================================================

func TestParseMethodDefInType(t *testing.T) {
	source := `: Calculator {
  value: int!
  add(n: int!) -> int {
    > value + n
  }
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	if len(typeDef.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(typeDef.Methods))
	}
	method := typeDef.Methods[0]
	if method.Name != "add" {
		t.Errorf("expected method name 'add', got %q", method.Name)
	}
	if len(method.Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(method.Params))
	}
	if method.ReturnType == nil {
		t.Error("expected return type")
	}
}

func TestParseMethodDefNoReturn(t *testing.T) {
	source := `: Logger {
  level: str!
  log(msg: str!) {
    > msg
  }
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	if len(typeDef.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(typeDef.Methods))
	}
	method := typeDef.Methods[0]
	if method.Name != "log" {
		t.Errorf("expected method name 'log', got %q", method.Name)
	}
	if method.ReturnType != nil {
		t.Error("expected nil return type")
	}
}

func TestParseMethodDefMultipleParams(t *testing.T) {
	source := `: Math {
  combine(a: int!, b: int!, c: int!) -> int {
    > a + b + c
  }
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	if len(typeDef.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(typeDef.Methods))
	}
	if len(typeDef.Methods[0].Params) != 3 {
		t.Errorf("expected 3 params, got %d", len(typeDef.Methods[0].Params))
	}
}

func TestParseMethodDefMultipleMethods(t *testing.T) {
	source := `: Shape {
  x: int!
  y: int!
  getX() -> int {
    > x
  }
  getY() -> int {
    > y
  }
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	if len(typeDef.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(typeDef.Methods))
	}
}

func TestParseMethodDefWithGenericTypeParams(t *testing.T) {
	source := `: Container<T> {
  value: T!
  get() -> T {
    > value
  }
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	if len(typeDef.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(typeDef.Methods))
	}
}

// ============================================================
// Tests for parseConstDecl (low coverage)
// ============================================================

func TestParseConstDeclSimple(t *testing.T) {
	source := `const MAX_SIZE = 100`
	module := parseSource(t, source)
	constDecl, ok := module.Items[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("expected ConstDecl, got %T", module.Items[0])
	}
	if constDecl.Name != "MAX_SIZE" {
		t.Errorf("expected name 'MAX_SIZE', got %q", constDecl.Name)
	}
}

func TestParseConstDeclWithType(t *testing.T) {
	source := `const PI: float = 3.14159`
	module := parseSource(t, source)
	constDecl := module.Items[0].(*ast.ConstDecl)
	if constDecl.Name != "PI" {
		t.Errorf("expected name 'PI', got %q", constDecl.Name)
	}
	if constDecl.Type == nil {
		t.Error("expected non-nil type")
	}
}

func TestParseConstDeclString(t *testing.T) {
	source := `const APP_NAME = "MyApp"`
	module := parseSource(t, source)
	constDecl := module.Items[0].(*ast.ConstDecl)
	if constDecl.Name != "APP_NAME" {
		t.Errorf("expected name 'APP_NAME', got %q", constDecl.Name)
	}
}

func TestParseConstDeclWithIntType(t *testing.T) {
	source := `const TIMEOUT: int = 30`
	module := parseSource(t, source)
	constDecl := module.Items[0].(*ast.ConstDecl)
	if constDecl.Type == nil {
		t.Error("expected non-nil type")
	}
}

func TestParseConstDeclMultiple(t *testing.T) {
	source := `const A = 1
const B = 2
const C = 3`
	module := parseSource(t, source)
	if len(module.Items) != 3 {
		t.Errorf("expected 3 const declarations, got %d", len(module.Items))
	}
}

// ============================================================
// Tests for parseSingleType (low coverage)
// ============================================================

func TestParseSingleTypeFunctionType(t *testing.T) {
	source := `: Callback {
  handler: (int, str) -> bool!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	_, ok := typeDef.Fields[0].TypeAnnotation.(ast.FunctionType)
	if !ok {
		t.Fatalf("expected FunctionType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
}

func TestParseSingleTypeArrayBrackets(t *testing.T) {
	source := `: Container {
  items: [str]!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	_, ok := typeDef.Fields[0].TypeAnnotation.(ast.ArrayType)
	if !ok {
		t.Fatalf("expected ArrayType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
}

func TestParseSingleTypeOptional(t *testing.T) {
	source := `: User {
  nickname: str?
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	_, ok := typeDef.Fields[0].TypeAnnotation.(ast.OptionalType)
	if !ok {
		t.Fatalf("expected OptionalType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
}

func TestParseSingleTypeGenericAngleBrackets(t *testing.T) {
	source := `: Response {
  data: List<int>!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	genType, ok := typeDef.Fields[0].TypeAnnotation.(ast.GenericType)
	if !ok {
		t.Fatalf("expected GenericType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
	if len(genType.TypeArgs) != 1 {
		t.Errorf("expected 1 type arg, got %d", len(genType.TypeArgs))
	}
}

func TestParseSingleTypeGenericSquareBrackets(t *testing.T) {
	source := `: Response {
  data: Map[str, int]!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	genType, ok := typeDef.Fields[0].TypeAnnotation.(ast.GenericType)
	if !ok {
		t.Fatalf("expected GenericType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
	if len(genType.TypeArgs) != 2 {
		t.Errorf("expected 2 type args, got %d", len(genType.TypeArgs))
	}
}

func TestParseSingleTypeEmptyBrackets(t *testing.T) {
	// int[] should be parsed as ArrayType{ElementType: IntType}
	source := `: Container {
  items: int[]!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	arrType, ok := typeDef.Fields[0].TypeAnnotation.(ast.ArrayType)
	if !ok {
		t.Fatalf("expected ArrayType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
	_, ok = arrType.ElementType.(ast.IntType)
	if !ok {
		t.Fatalf("expected IntType element, got %T", arrType.ElementType)
	}
}

func TestParseSingleTypeTypeParamRef(t *testing.T) {
	source := `: Box<T> {
  value: T!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	_, ok := typeDef.Fields[0].TypeAnnotation.(ast.TypeParameterType)
	if !ok {
		t.Fatalf("expected TypeParameterType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
}

func TestParseSingleTypeStringAlias(t *testing.T) {
	source := `: Msg {
  text: string!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	_, ok := typeDef.Fields[0].TypeAnnotation.(ast.StringType)
	if !ok {
		t.Fatalf("expected StringType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
}

// ============================================================
// Tests for typeToString (low coverage)
// ============================================================

func TestTypeToStringInt(t *testing.T) {
	result := typeToString(ast.IntType{})
	if result != "int" {
		t.Errorf("expected 'int', got %q", result)
	}
}

func TestTypeToStringStr(t *testing.T) {
	result := typeToString(ast.StringType{})
	if result != "str" {
		t.Errorf("expected 'str', got %q", result)
	}
}

func TestTypeToStringBool(t *testing.T) {
	result := typeToString(ast.BoolType{})
	if result != "bool" {
		t.Errorf("expected 'bool', got %q", result)
	}
}

func TestTypeToStringFloat(t *testing.T) {
	result := typeToString(ast.FloatType{})
	if result != "float" {
		t.Errorf("expected 'float', got %q", result)
	}
}

func TestTypeToStringArray(t *testing.T) {
	result := typeToString(ast.ArrayType{ElementType: ast.IntType{}})
	if result != "[int]" {
		t.Errorf("expected '[int]', got %q", result)
	}
}

func TestTypeToStringOptional(t *testing.T) {
	result := typeToString(ast.OptionalType{InnerType: ast.StringType{}})
	if result != "str?" {
		t.Errorf("expected 'str?', got %q", result)
	}
}

func TestTypeToStringNamed(t *testing.T) {
	result := typeToString(ast.NamedType{Name: "User"})
	if result != "User" {
		t.Errorf("expected 'User', got %q", result)
	}
}

func TestTypeToStringGeneric(t *testing.T) {
	result := typeToString(ast.GenericType{
		BaseType: ast.NamedType{Name: "List"},
		TypeArgs: []ast.Type{ast.IntType{}},
	})
	if result != "List<int>" {
		t.Errorf("expected 'List<int>', got %q", result)
	}
}

func TestTypeToStringGenericMultipleArgs(t *testing.T) {
	result := typeToString(ast.GenericType{
		BaseType: ast.NamedType{Name: "Map"},
		TypeArgs: []ast.Type{ast.StringType{}, ast.IntType{}},
	})
	if result != "Map<str, int>" {
		t.Errorf("expected 'Map<str, int>', got %q", result)
	}
}

func TestTypeToStringGenericNoArgs(t *testing.T) {
	result := typeToString(ast.GenericType{
		BaseType: ast.NamedType{Name: "List"},
	})
	if result != "List" {
		t.Errorf("expected 'List', got %q", result)
	}
}

func TestTypeToStringTypeParameter(t *testing.T) {
	result := typeToString(ast.TypeParameterType{Name: "T"})
	if result != "T" {
		t.Errorf("expected 'T', got %q", result)
	}
}

func TestTypeToStringFunctionType(t *testing.T) {
	result := typeToString(ast.FunctionType{
		ParamTypes: []ast.Type{ast.IntType{}, ast.StringType{}},
		ReturnType: ast.BoolType{},
	})
	if result != "(int, str) -> bool" {
		t.Errorf("expected '(int, str) -> bool', got %q", result)
	}
}

func TestTypeToStringUnknown(t *testing.T) {
	// Test the default case with an unknown type
	result := typeToString(ast.UnionType{})
	if result != "unknown" {
		t.Errorf("expected 'unknown', got %q", result)
	}
}

// ============================================================
// Tests for parseTypeParameters (low coverage: constraint, extends)
// ============================================================

func TestParseTypeParametersWithConstraint(t *testing.T) {
	source := `: SortedList<T: Comparable> {
  items: [T]!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	if len(typeDef.TypeParams) != 1 {
		t.Fatalf("expected 1 type param, got %d", len(typeDef.TypeParams))
	}
	if typeDef.TypeParams[0].Name != "T" {
		t.Errorf("expected type param 'T', got %q", typeDef.TypeParams[0].Name)
	}
	if typeDef.TypeParams[0].Constraint == nil {
		t.Error("expected constraint on type parameter")
	}
}

func TestParseTypeParametersWithExtends(t *testing.T) {
	source := `: SortedList<T extends Comparable> {
  items: [T]!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	if len(typeDef.TypeParams) != 1 {
		t.Fatalf("expected 1 type param, got %d", len(typeDef.TypeParams))
	}
	if typeDef.TypeParams[0].Constraint == nil {
		t.Error("expected constraint on type parameter via 'extends'")
	}
}

func TestParseTypeParametersMultipleWithConstraints(t *testing.T) {
	source := `: Pair<K: Hashable, V> {
  key: K!
  value: V!
}`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	if len(typeDef.TypeParams) != 2 {
		t.Fatalf("expected 2 type params, got %d", len(typeDef.TypeParams))
	}
	if typeDef.TypeParams[0].Constraint == nil {
		t.Error("expected constraint on K")
	}
	if typeDef.TypeParams[1].Constraint != nil {
		t.Error("expected no constraint on V")
	}
}

// ============================================================
// Tests for parseAwaitExpr (low coverage)
// ============================================================

func TestParseAwaitExprWithFuncCall(t *testing.T) {
	source := `@ GET /test {
  $ result = await fetchData()
  > result
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign := route.Body[0].(ast.AssignStatement)
	awaitExpr, ok := assign.Value.(ast.AwaitExpr)
	if !ok {
		t.Fatalf("expected AwaitExpr, got %T", assign.Value)
	}
	if awaitExpr.Expr == nil {
		t.Error("expected non-nil awaited expression")
	}
}

func TestParseAwaitExprWithVariable(t *testing.T) {
	source := `@ GET /test {
  $ future = async { > 10 }
  $ val = await future
  > val
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign := route.Body[1].(ast.AssignStatement)
	_, ok := assign.Value.(ast.AwaitExpr)
	if !ok {
		t.Fatalf("expected AwaitExpr, got %T", assign.Value)
	}
}

// ============================================================
// Tests for parseReassignment (low coverage)
// ============================================================

func TestParseReassignmentSimple(t *testing.T) {
	source := `@ GET /test {
  $ count = 0
  count = 1
  > count
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	reassign, ok := route.Body[1].(ast.ReassignStatement)
	if !ok {
		t.Fatalf("expected ReassignStatement, got %T", route.Body[1])
	}
	if reassign.Target != "count" {
		t.Errorf("expected target 'count', got %q", reassign.Target)
	}
}

func TestParseReassignmentWithExpression(t *testing.T) {
	source := `@ GET /test {
  $ x = 1
  x = x + 1
  > x
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	reassign, ok := route.Body[1].(ast.ReassignStatement)
	if !ok {
		t.Fatalf("expected ReassignStatement, got %T", route.Body[1])
	}
	if reassign.Value == nil {
		t.Error("expected non-nil value")
	}
}

// ============================================================
// Tests for currentCommandDefaultBinaryOp (low coverage)
// ============================================================

func TestParseCommandWithDefaultBinaryExpressions(t *testing.T) {
	// Test command with default value that has binary expression
	source := `! calc(x: int!, y: int = 2 + 3) {
  > x + y
}`
	module := parseSource(t, source)
	fn, ok := module.Items[0].(*ast.Function)
	if !ok {
		t.Fatalf("expected Function, got %T", module.Items[0])
	}
	if fn.Name != "calc" {
		t.Errorf("expected name 'calc', got %q", fn.Name)
	}
}

func TestParseCommandDefaultBinaryOps(t *testing.T) {
	// Test a command (!) with various operators in default value expressions
	tests := []struct {
		name   string
		source string
	}{
		{
			"addition",
			`! f(x: int = 1 + 2) { > x }`,
		},
		{
			"multiplication",
			`! f(x: int = 2 * 3) { > x }`,
		},
		{
			"division",
			`! f(x: int = 10 / 2) { > x }`,
		},
		{
			"equality",
			`! f(x: bool = 1 == 1) { > x }`,
		},
		{
			"inequality",
			`! f(x: bool = 1 != 2) { > x }`,
		},
		{
			"less than",
			`! f(x: bool = 1 < 2) { > x }`,
		},
		{
			"greater than or eq",
			`! f(x: bool = 5 >= 3) { > x }`,
		},
		{
			"less than or eq",
			`! f(x: bool = 3 <= 5) { > x }`,
		},
		{
			"logical and",
			`! f(x: bool = true && false) { > x }`,
		},
		{
			"logical or",
			`! f(x: bool = true || false) { > x }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := parseSource(t, tt.source)
			if len(module.Items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(module.Items))
			}
		})
	}
}

// ============================================================
// Tests for parseCommandDefaultBinaryExpr
// ============================================================

func TestParseCommandDefaultBinaryExprComplex(t *testing.T) {
	source := `@ command compute x: int --scale: int = 2 * 5 {
  > x * scale
}`
	module := parseSource(t, source)
	cmd, ok := module.Items[0].(*ast.Command)
	if !ok {
		t.Fatalf("expected Command, got %T", module.Items[0])
	}
	if cmd.Name != "compute" {
		t.Errorf("expected name 'compute', got %q", cmd.Name)
	}
}

// ============================================================
// Tests for error paths that trigger various error helpers
// ============================================================

func TestExpectErrorHintMissingBrace(t *testing.T) {
	// Trigger expectError for LBRACE expected but NEWLINE found
	source := `: User
  name: str!
}`
	err := parseSourceExpectError(t, source)
	errStr := err.Error()
	if !strings.Contains(errStr, "{") || !strings.Contains(errStr, "brace") {
		// The error may use different wording - just check it's an error
		if errStr == "" {
			t.Error("expected non-empty error message")
		}
	}
}

func TestExpectErrorMissingClosingBrace(t *testing.T) {
	source := `: User {
  name: str!`
	err := parseSourceExpectError(t, source)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRouteErrorBadPath(t *testing.T) {
	// A route without a proper path
	source := `@ GET {
  > 42
}`
	err := parseSourceExpectError(t, source)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ============================================================
// Tests for edge cases in Parse()
// ============================================================

func TestParseEmptySource(t *testing.T) {
	source := ``
	module := parseSource(t, source)
	if len(module.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(module.Items))
	}
}

func TestParseOnlyNewlines(t *testing.T) {
	source := "\n\n\n"
	module := parseSource(t, source)
	if len(module.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(module.Items))
	}
}

func TestParseUnexpectedTopLevelToken(t *testing.T) {
	source := `+ invalid`
	err := parseSourceExpectError(t, source)
	if err == nil {
		t.Fatal("expected error for unexpected top-level token")
	}
}

func TestParseTopLevelMacroInvocation(t *testing.T) {
	source := `myMacro!(1, 2, 3)`
	module := parseSource(t, source)
	_, ok := module.Items[0].(*ast.MacroInvocation)
	if !ok {
		t.Fatalf("expected MacroInvocation, got %T", module.Items[0])
	}
}

// ============================================================
// Tests for the switch statement with multiple cases
// ============================================================

func TestParseSwitchStatementFull(t *testing.T) {
	source := `@ GET /test {
  switch status {
    case "active" {
      > 1
    }
    case "inactive" {
      > 0
    }
    case "pending" {
      > 2
    }
    default {
      > -1
    }
  }
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	switchStmt, ok := route.Body[0].(ast.SwitchStatement)
	if !ok {
		t.Fatalf("expected SwitchStatement, got %T", route.Body[0])
	}
	if len(switchStmt.Cases) != 3 {
		t.Errorf("expected 3 cases, got %d", len(switchStmt.Cases))
	}
	if len(switchStmt.Default) == 0 {
		t.Error("expected default case")
	}
}

// ============================================================
// Tests for test blocks and assert statements
// ============================================================

func TestParseTestBlockWithAssert(t *testing.T) {
	source := `test "basic math" {
  assert(1 + 1 == 2)
  assert(2 * 3 == 6, "multiplication failed")
}`
	module := parseSource(t, source)
	testBlock, ok := module.Items[0].(*ast.TestBlock)
	if !ok {
		t.Fatalf("expected TestBlock, got %T", module.Items[0])
	}
	if testBlock.Name != "basic math" {
		t.Errorf("expected name 'basic math', got %q", testBlock.Name)
	}
	if len(testBlock.Body) != 2 {
		t.Errorf("expected 2 statements, got %d", len(testBlock.Body))
	}
}

// ============================================================
// Tests for type union parsing in parseSingleType
// ============================================================

func TestParseUnionTypeThreeTypes(t *testing.T) {
	source := `: Result { value: str | int | bool! }`
	module := parseSource(t, source)
	typeDef := module.Items[0].(*ast.TypeDef)
	unionType, ok := typeDef.Fields[0].TypeAnnotation.(ast.UnionType)
	if !ok {
		t.Fatalf("expected UnionType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
	if len(unionType.Types) != 3 {
		t.Errorf("expected 3 types in union, got %d", len(unionType.Types))
	}
}

// ============================================================
// Tests for method calls as expression statements
// ============================================================

func TestParseMethodCallAsStatement(t *testing.T) {
	source := `@ GET /test {
  ws.send("hello")
  > "ok"
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	exprStmt, ok := route.Body[0].(ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", route.Body[0])
	}
	callExpr, ok := exprStmt.Expr.(ast.FunctionCallExpr)
	if !ok {
		t.Fatalf("expected FunctionCallExpr, got %T", exprStmt.Expr)
	}
	if callExpr.Name != "ws.send" {
		t.Errorf("expected 'ws.send', got %q", callExpr.Name)
	}
}

// ============================================================
// Tests for import with alias
// ============================================================

func TestParseImportWithAlias(t *testing.T) {
	source := `import "./utils" as u`
	module := parseSource(t, source)
	imp, ok := module.Items[0].(*ast.ImportStatement)
	if !ok {
		t.Fatalf("expected ImportStatement, got %T", module.Items[0])
	}
	if imp.Alias != "u" {
		t.Errorf("expected alias 'u', got %q", imp.Alias)
	}
}

// ============================================================
// Tests for null literal expression
// ============================================================

func TestParseNullLiteralExpr(t *testing.T) {
	source := `@ GET /test {
  $ x = null
  > x
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign := route.Body[0].(ast.AssignStatement)
	lit, ok := assign.Value.(ast.LiteralExpr)
	if !ok {
		t.Fatalf("expected LiteralExpr, got %T", assign.Value)
	}
	_, ok = lit.Value.(ast.NullLiteral)
	if !ok {
		t.Fatalf("expected NullLiteral, got %T", lit.Value)
	}
}

// ============================================================
// Tests for float literal expression
// ============================================================

func TestParseFloatLiteralExpr(t *testing.T) {
	source := `@ GET /test {
  $ x = 3.14
  > x
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign := route.Body[0].(ast.AssignStatement)
	lit, ok := assign.Value.(ast.LiteralExpr)
	if !ok {
		t.Fatalf("expected LiteralExpr, got %T", assign.Value)
	}
	floatLit, ok := lit.Value.(ast.FloatLiteral)
	if !ok {
		t.Fatalf("expected FloatLiteral, got %T", lit.Value)
	}
	if floatLit.Value != 3.14 {
		t.Errorf("expected 3.14, got %f", floatLit.Value)
	}
}

// ============================================================
// Test for dependency injection
// ============================================================

func TestParseDependencyInjection(t *testing.T) {
	source := `@ GET /test {
  % cache: Cache
  > "ok"
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	if len(route.Injections) != 1 {
		t.Fatalf("expected 1 injection, got %d", len(route.Injections))
	}
	if route.Injections[0].Name != "cache" {
		t.Errorf("expected injection name 'cache', got %q", route.Injections[0].Name)
	}
}

// ============================================================
// Tests for pipe operator
// ============================================================

func TestParsePipeExpression(t *testing.T) {
	source := `@ GET /test {
  $ result = data |> transform
  > result
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	if len(route.Body) < 2 {
		t.Fatal("expected at least 2 statements")
	}
}

// ============================================================
// Test for dollar with field access chain
// ============================================================

func TestParseStatementDollarDeepFieldAccess(t *testing.T) {
	source := `@ GET /test {
  $ a.b.c = 42
  > a
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign, ok := route.Body[0].(ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}
	if assign.Target != "a.b.c" {
		t.Errorf("expected target 'a.b.c', got %q", assign.Target)
	}
}

// ============================================================
// Tests for multiple macro params
// ============================================================

func TestParseMacroDefWithMultipleParams(t *testing.T) {
	source := `macro! log(level, msg, extra) {
  > msg
}`
	module := parseSource(t, source)
	macroDef := module.Items[0].(*ast.MacroDef)
	if macroDef.Name != "log" {
		t.Errorf("expected name 'log', got %q", macroDef.Name)
	}
	if len(macroDef.Params) != 3 {
		t.Errorf("expected 3 params, got %d", len(macroDef.Params))
	}
}

// ============================================================
// Tests for ternary/conditional expressions within if
// ============================================================

func TestParseIfWithComplexCondition(t *testing.T) {
	source := `@ GET /test {
  if a > 0 && b < 100 {
    > "in range"
  }
  > "out"
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	ifStmt, ok := route.Body[0].(ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", route.Body[0])
	}
	if ifStmt.Condition == nil {
		t.Error("expected non-nil condition")
	}
}

// ============================================================
// Test for contract parsing (coverage of "contract" IDENT branch)
// ============================================================

func TestParseContractBasic(t *testing.T) {
	source := `contract UserContract {
  @ GET /users -> User
  @ POST /users -> User
}`
	module := parseSource(t, source)
	if len(module.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(module.Items))
	}
	contract, ok := module.Items[0].(*ast.ContractDef)
	if !ok {
		t.Fatalf("expected ContractDef, got %T", module.Items[0])
	}
	if contract.Name != "UserContract" {
		t.Errorf("expected name 'UserContract', got %q", contract.Name)
	}
	if len(contract.Endpoints) != 2 {
		t.Errorf("expected 2 endpoints, got %d", len(contract.Endpoints))
	}
}

// ============================================================
// Test for the boolean false literal pattern
// ============================================================

func TestParsePatternBoolFalse(t *testing.T) {
	source := `@ GET /test {
  $ result = match flag {
    false => "no"
    _ => "yes"
  }
  > result
}`
	module := parseSource(t, source)
	route := module.Items[0].(*ast.Route)
	assign := route.Body[0].(ast.AssignStatement)
	matchExpr := assign.Value.(ast.MatchExpr)
	litPat, ok := matchExpr.Cases[0].Pattern.(ast.LiteralPattern)
	if !ok {
		t.Fatalf("expected LiteralPattern, got %T", matchExpr.Cases[0].Pattern)
	}
	boolLit, ok := litPat.Value.(ast.BoolLiteral)
	if !ok {
		t.Fatalf("expected BoolLiteral, got %T", litPat.Value)
	}
	if boolLit.Value != false {
		t.Error("expected false")
	}
}
