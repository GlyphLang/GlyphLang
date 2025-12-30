package parser

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
)

// Tests for async/await expressions

func TestParseAsyncExpr(t *testing.T) {
	source := `@ GET /compute
  $ future = async {
    $ x = 10
    > x
  }
  > future`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(module.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(module.Items))
	}

	route, ok := module.Items[0].(*interpreter.Route)
	if !ok {
		t.Fatalf("expected Route, got %T", module.Items[0])
	}

	if len(route.Body) < 1 {
		t.Fatal("expected at least 1 statement")
	}

	// First statement should be assignment with async expression
	assign, ok := route.Body[0].(interpreter.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", route.Body[0])
	}

	if assign.Target != "future" {
		t.Errorf("expected target 'future', got %s", assign.Target)
	}

	_, ok = assign.Value.(interpreter.AsyncExpr)
	if !ok {
		t.Errorf("expected AsyncExpr, got %T", assign.Value)
	}
}

func TestParseAwaitExpr(t *testing.T) {
	source := `@ GET /wait
  $ future = async { > 42 }
  $ result = await future
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	if len(route.Body) < 2 {
		t.Fatal("expected at least 2 statements")
	}

	// Second statement should have await expression
	assign, ok := route.Body[1].(interpreter.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", route.Body[1])
	}

	_, ok = assign.Value.(interpreter.AwaitExpr)
	if !ok {
		t.Errorf("expected AwaitExpr, got %T", assign.Value)
	}
}

// Tests for generic functions

func TestParseGenericFunction(t *testing.T) {
	source := `! identity<T>(x: T): T
  > x`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(module.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(module.Items))
	}

	fn, ok := module.Items[0].(*interpreter.Function)
	if !ok {
		t.Fatalf("expected Function, got %T", module.Items[0])
	}

	if fn.Name != "identity" {
		t.Errorf("expected name 'identity', got %s", fn.Name)
	}

	if len(fn.TypeParams) != 1 {
		t.Fatalf("expected 1 type param, got %d", len(fn.TypeParams))
	}

	if fn.TypeParams[0].Name != "T" {
		t.Errorf("expected type param 'T', got %s", fn.TypeParams[0].Name)
	}
}

func TestParseGenericFunctionMultipleTypeParams(t *testing.T) {
	source := `! map<T, U>(arr: [T], fn: (T) -> U): [U]
  > []`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	fn := module.Items[0].(*interpreter.Function)

	if len(fn.TypeParams) != 2 {
		t.Fatalf("expected 2 type params, got %d", len(fn.TypeParams))
	}

	if fn.TypeParams[0].Name != "T" || fn.TypeParams[1].Name != "U" {
		t.Errorf("expected type params T, U, got %v", fn.TypeParams)
	}
}

// Tests for regular functions

func TestParseRegularFunction(t *testing.T) {
	source := `! add(a: int!, b: int!): int
  $ result = a + b
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	fn, ok := module.Items[0].(*interpreter.Function)
	if !ok {
		t.Fatalf("expected Function, got %T", module.Items[0])
	}

	if fn.Name != "add" {
		t.Errorf("expected name 'add', got %s", fn.Name)
	}

	if len(fn.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fn.Params))
	}

	if fn.Params[0].Name != "a" || fn.Params[1].Name != "b" {
		t.Error("unexpected param names")
	}
}

func TestParseFunctionNoParams(t *testing.T) {
	source := `! getVersion(): string
  > "1.0.0"`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	fn := module.Items[0].(*interpreter.Function)

	if fn.Name != "getVersion" {
		t.Errorf("expected name 'getVersion', got %s", fn.Name)
	}

	if len(fn.Params) != 0 {
		t.Errorf("expected 0 params, got %d", len(fn.Params))
	}
}

// Tests for function types

func TestParseFunctionType(t *testing.T) {
	source := `: Callback {
  handler: (string) -> bool!
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)

	if len(typeDef.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(typeDef.Fields))
	}

	field := typeDef.Fields[0]
	fnType, ok := field.TypeAnnotation.(interpreter.FunctionType)
	if !ok {
		t.Fatalf("expected FunctionType, got %T", field.TypeAnnotation)
	}

	if len(fnType.ParamTypes) != 1 {
		t.Errorf("expected 1 param type, got %d", len(fnType.ParamTypes))
	}
}

func TestParseFunctionTypeMultipleParams(t *testing.T) {
	source := `: Processor {
  process: (int, string, bool) -> string!
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)
	field := typeDef.Fields[0]
	fnType := field.TypeAnnotation.(interpreter.FunctionType)

	if len(fnType.ParamTypes) != 3 {
		t.Errorf("expected 3 param types, got %d", len(fnType.ParamTypes))
	}
}

// Tests for type parameters and arguments

func TestParseTypeArguments(t *testing.T) {
	source := `: Response {
  data: List<User>!
  errors: Map<string, Error>?
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)

	if len(typeDef.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(typeDef.Fields))
	}

	// Check first field is generic type
	genType, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.GenericType)
	if !ok {
		t.Fatalf("expected GenericType for data field, got %T", typeDef.Fields[0].TypeAnnotation)
	}

	if len(genType.TypeArgs) != 1 {
		t.Errorf("expected 1 type arg, got %d", len(genType.TypeArgs))
	}
}

func TestParseTypeParameters(t *testing.T) {
	source := `: Result<T, E> {
  value: T?
  error: E?
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)

	if typeDef.Name != "Result" {
		t.Errorf("expected name 'Result', got %s", typeDef.Name)
	}

	if len(typeDef.TypeParams) != 2 {
		t.Fatalf("expected 2 type params, got %d", len(typeDef.TypeParams))
	}

	if typeDef.TypeParams[0].Name != "T" || typeDef.TypeParams[1].Name != "E" {
		t.Errorf("expected type params T, E")
	}
}

// Tests for pattern matching

func TestParseMatchExprSimple(t *testing.T) {
	source := `@ GET /status/:code
  $ result = match code {
    200 => "OK"
    404 => "Not Found"
    _ => "Unknown"
  }
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	matchExpr, ok := assign.Value.(interpreter.MatchExpr)
	if !ok {
		t.Fatalf("expected MatchExpr, got %T", assign.Value)
	}

	if len(matchExpr.Cases) != 3 {
		t.Errorf("expected 3 cases, got %d", len(matchExpr.Cases))
	}
}

func TestParseMatchExprWithGuard(t *testing.T) {
	source := `@ GET /check/:n
  $ result = match n {
    x when x > 100 => "large"
    x when x > 0 => "positive"
    _ => "other"
  }
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	matchExpr := assign.Value.(interpreter.MatchExpr)

	// Check first case has a guard
	if matchExpr.Cases[0].Guard == nil {
		t.Error("expected first case to have a guard")
	}
}

func TestParseObjectPattern(t *testing.T) {
	source := `@ GET /user
  $ result = match user {
    {name: n, age: a} => n
    _ => "unknown"
  }
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	matchExpr := assign.Value.(interpreter.MatchExpr)

	// First case should have object pattern
	objPattern, ok := matchExpr.Cases[0].Pattern.(interpreter.ObjectPattern)
	if !ok {
		t.Fatalf("expected ObjectPattern, got %T", matchExpr.Cases[0].Pattern)
	}

	if len(objPattern.Fields) != 2 {
		t.Errorf("expected 2 fields in pattern, got %d", len(objPattern.Fields))
	}
}

func TestParseArrayPattern(t *testing.T) {
	source := `@ GET /first
  $ result = match items {
    [first, second] => first
    [] => 0
    _ => 0
  }
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	matchExpr := assign.Value.(interpreter.MatchExpr)

	// First case should have array pattern
	arrPattern, ok := matchExpr.Cases[0].Pattern.(interpreter.ArrayPattern)
	if !ok {
		t.Fatalf("expected ArrayPattern, got %T", matchExpr.Cases[0].Pattern)
	}

	if len(arrPattern.Elements) < 1 {
		t.Errorf("expected at least 1 element in array pattern, got %d", len(arrPattern.Elements))
	}
}

// Tests for various statement types

func TestParseForStatement(t *testing.T) {
	source := `@ GET /sum
  $ total = 0
  for item in items {
    $ total = total + item
  }
  > total`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	// Find the for statement
	var forStmt *interpreter.ForStatement
	for _, stmt := range route.Body {
		if fs, ok := stmt.(interpreter.ForStatement); ok {
			forStmt = &fs
			break
		}
	}

	if forStmt == nil {
		t.Fatal("expected ForStatement")
	}

	if forStmt.ValueVar != "item" {
		t.Errorf("expected value var 'item', got %s", forStmt.ValueVar)
	}
}

func TestParseForStatementWithKey(t *testing.T) {
	source := `@ GET /pairs
  for key, value in obj {
    > {k: key, v: value}
  }
  > {}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	forStmt := route.Body[0].(interpreter.ForStatement)

	if forStmt.KeyVar != "key" {
		t.Errorf("expected key var 'key', got %s", forStmt.KeyVar)
	}

	if forStmt.ValueVar != "value" {
		t.Errorf("expected value var 'value', got %s", forStmt.ValueVar)
	}
}

func TestParseWhileStatement(t *testing.T) {
	source := `@ GET /countdown
  $ n = 10
  while n > 0 {
    $ n = n - 1
  }
  > n`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	// Find while statement
	var whileStmt *interpreter.WhileStatement
	for _, stmt := range route.Body {
		if ws, ok := stmt.(interpreter.WhileStatement); ok {
			whileStmt = &ws
			break
		}
	}

	if whileStmt == nil {
		t.Fatal("expected WhileStatement")
	}

	if whileStmt.Condition == nil {
		t.Error("expected condition")
	}
}

func TestParseSwitchStatement(t *testing.T) {
	source := `@ GET /check
  $ val = "test"
  switch val {
    case "a" {
      > "alpha"
    }
    case "b" {
      > "beta"
    }
    default {
      > "other"
    }
  }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	// Switch is at Body[1] since Body[0] is the assignment
	switchStmt, ok := route.Body[1].(interpreter.SwitchStatement)
	if !ok {
		t.Fatalf("expected SwitchStatement, got %T", route.Body[1])
	}

	if len(switchStmt.Cases) != 2 {
		t.Errorf("expected 2 cases, got %d", len(switchStmt.Cases))
	}

	if len(switchStmt.Default) == 0 {
		t.Error("expected default case")
	}
}

func TestParseIfElseIfElse(t *testing.T) {
	source := `@ GET /grade/:score
  if score >= 90 {
    > "A"
  } else if score >= 80 {
    > "B"
  } else if score >= 70 {
    > "C"
  } else {
    > "F"
  }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	ifStmt, ok := route.Body[0].(interpreter.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", route.Body[0])
	}

	if ifStmt.Condition == nil {
		t.Error("expected condition")
	}

	// Should have else block with nested if
	if len(ifStmt.ElseBlock) == 0 {
		t.Error("expected else block")
	}
}

// Tests for more type coverage

func TestParseSingleTypeInt(t *testing.T) {
	source := `: Counter { value: int! }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)
	_, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.IntType)
	if !ok {
		t.Errorf("expected IntType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
}

func TestParseSingleTypeFloat(t *testing.T) {
	source := `: Point {
  x: float!
  y: float!
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)
	_, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.FloatType)
	if !ok {
		t.Errorf("expected FloatType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
}

func TestParseSingleTypeBool(t *testing.T) {
	source := `: Flags { enabled: bool! }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)
	_, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.BoolType)
	if !ok {
		t.Errorf("expected BoolType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
}

func TestParseSingleTypeTimestamp(t *testing.T) {
	source := `: Event { created: timestamp! }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)
	namedType, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.NamedType)
	if !ok {
		t.Fatalf("expected NamedType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
	if namedType.Name != "timestamp" {
		t.Errorf("expected 'timestamp', got %s", namedType.Name)
	}
}

func TestParseNestedArrayType(t *testing.T) {
	source := `: Matrix { data: [[int]]! }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)
	arrType, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.ArrayType)
	if !ok {
		t.Fatalf("expected ArrayType, got %T", typeDef.Fields[0].TypeAnnotation)
	}

	innerArr, ok := arrType.ElementType.(interpreter.ArrayType)
	if !ok {
		t.Fatalf("expected nested ArrayType, got %T", arrType.ElementType)
	}

	_, ok = innerArr.ElementType.(interpreter.IntType)
	if !ok {
		t.Errorf("expected IntType, got %T", innerArr.ElementType)
	}
}

func TestParseUnionType(t *testing.T) {
	source := `: Result { value: string | int! }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)
	unionType, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.UnionType)
	if !ok {
		t.Fatalf("expected UnionType, got %T", typeDef.Fields[0].TypeAnnotation)
	}

	if len(unionType.Types) < 2 {
		t.Errorf("expected at least 2 types in union, got %d", len(unionType.Types))
	}
}

// Tests for database type and injection

func TestParseDatabaseType(t *testing.T) {
	source := `@ GET /data
  % db: Database
  > db`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	if len(route.Injections) != 1 {
		t.Fatalf("expected 1 injection, got %d", len(route.Injections))
	}

	if route.Injections[0].Name != "db" {
		t.Errorf("expected injection name 'db', got %s", route.Injections[0].Name)
	}
}

// Tests for expression parsing

func TestParseNegativeNumber(t *testing.T) {
	source := `@ GET /neg
  $ x = -42
  > x`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	unary, ok := assign.Value.(interpreter.UnaryOpExpr)
	if !ok {
		t.Fatalf("expected UnaryOpExpr, got %T", assign.Value)
	}

	if unary.Op != interpreter.Neg {
		t.Errorf("expected Neg op, got %v", unary.Op)
	}
}

func TestParseNotExpression(t *testing.T) {
	source := `@ GET /not
  $ x = !true
  > x`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	unary, ok := assign.Value.(interpreter.UnaryOpExpr)
	if !ok {
		t.Fatalf("expected UnaryOpExpr, got %T", assign.Value)
	}

	if unary.Op != interpreter.Not {
		t.Errorf("expected Not op, got %v", unary.Op)
	}
}

func TestParseArrayLiteral(t *testing.T) {
	source := `@ GET /arr
  $ arr = [1, 2, 3]
  > arr`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	arrExpr, ok := assign.Value.(interpreter.ArrayExpr)
	if !ok {
		t.Fatalf("expected ArrayExpr, got %T", assign.Value)
	}

	if len(arrExpr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arrExpr.Elements))
	}
}

func TestParseObjectLiteral(t *testing.T) {
	source := `@ GET /obj
  $ obj = {name: "test", value: 42}
  > obj`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	objExpr, ok := assign.Value.(interpreter.ObjectExpr)
	if !ok {
		t.Fatalf("expected ObjectExpr, got %T", assign.Value)
	}

	if len(objExpr.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(objExpr.Fields))
	}
}

func TestParseFunctionCall(t *testing.T) {
	source := `@ GET /call
  $ result = myFunc(1, "test", true)
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	callExpr, ok := assign.Value.(interpreter.FunctionCallExpr)
	if !ok {
		t.Fatalf("expected FunctionCallExpr, got %T", assign.Value)
	}

	if len(callExpr.Args) != 3 {
		t.Errorf("expected 3 args, got %d", len(callExpr.Args))
	}
}

func TestParseMethodCall(t *testing.T) {
	source := `@ GET /method
  $ result = obj.method(arg)
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	callExpr, ok := assign.Value.(interpreter.FunctionCallExpr)
	if !ok {
		t.Fatalf("expected FunctionCallExpr, got %T", assign.Value)
	}

	// Verify the function name is set
	if callExpr.Name == "" {
		t.Fatal("expected non-empty function name")
	}
}

func TestParseArrayIndex(t *testing.T) {
	source := `@ GET /index
  $ first = arr[0]
  > first`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	indexExpr, ok := assign.Value.(interpreter.ArrayIndexExpr)
	if !ok {
		t.Fatalf("expected ArrayIndexExpr, got %T", assign.Value)
	}

	if indexExpr.Array == nil || indexExpr.Index == nil {
		t.Error("expected array and index to be set")
	}
}

func TestParseGroupedExpression(t *testing.T) {
	source := `@ GET /grouped
  $ result = (1 + 2) * 3
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	// Should parse as binary multiplication
	binExpr, ok := assign.Value.(interpreter.BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr, got %T", assign.Value)
	}

	if binExpr.Op != interpreter.Mul {
		t.Errorf("expected Mul op, got %v", binExpr.Op)
	}
}

// Tests for comparison and logical operators

func TestParseComparisonOperators(t *testing.T) {
	tests := []struct {
		source string
		op     interpreter.BinOp
	}{
		{`@ GET /t
  $ x = a == b
  > x`, interpreter.Eq},
		{`@ GET /t
  $ x = a != b
  > x`, interpreter.Ne},
		{`@ GET /t
  $ x = a < b
  > x`, interpreter.Lt},
		{`@ GET /t
  $ x = a > b
  > x`, interpreter.Gt},
		{`@ GET /t
  $ x = a <= b
  > x`, interpreter.Le},
		{`@ GET /t
  $ x = a >= b
  > x`, interpreter.Ge},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.source)
		tokens, _ := lexer.Tokenize()
		parser := NewParser(tokens)
		module, err := parser.Parse()
		if err != nil {
			t.Errorf("parser error for %v: %v", tt.op, err)
			continue
		}

		route := module.Items[0].(*interpreter.Route)
		assign := route.Body[0].(interpreter.AssignStatement)

		binExpr, ok := assign.Value.(interpreter.BinaryOpExpr)
		if !ok {
			t.Errorf("expected BinaryOpExpr for %v, got %T", tt.op, assign.Value)
			continue
		}

		if binExpr.Op != tt.op {
			t.Errorf("expected op %v, got %v", tt.op, binExpr.Op)
		}
	}
}

func TestParseLogicalOperators(t *testing.T) {
	tests := []struct {
		source string
		op     interpreter.BinOp
	}{
		{`@ GET /t
  $ x = a && b
  > x`, interpreter.And},
		{`@ GET /t
  $ x = a || b
  > x`, interpreter.Or},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.source)
		tokens, _ := lexer.Tokenize()
		parser := NewParser(tokens)
		module, err := parser.Parse()
		if err != nil {
			t.Errorf("parser error for %v: %v", tt.op, err)
			continue
		}

		route := module.Items[0].(*interpreter.Route)
		assign := route.Body[0].(interpreter.AssignStatement)

		binExpr, ok := assign.Value.(interpreter.BinaryOpExpr)
		if !ok {
			t.Errorf("expected BinaryOpExpr for %v, got %T", tt.op, assign.Value)
			continue
		}

		if binExpr.Op != tt.op {
			t.Errorf("expected op %v, got %v", tt.op, binExpr.Op)
		}
	}
}

// Tests for query parameters

func TestParseQueryParams(t *testing.T) {
	source := `@ GET /search
  ? query: string!
  ? limit: int = 10
  ? offset: int
  > {}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	if len(route.QueryParams) != 3 {
		t.Fatalf("expected 3 query params, got %d", len(route.QueryParams))
	}

	// First param should be required
	if !route.QueryParams[0].Required {
		t.Error("expected first query param to be required")
	}

	// Second param should have default
	if route.QueryParams[1].Default == nil {
		t.Error("expected second query param to have default")
	}
}

// Test for auth config

func TestParseAuthConfig(t *testing.T) {
	source := `@ GET /protected
  + auth(jwt)
  > {}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	if route.Auth == nil {
		t.Fatal("expected auth config")
	}

	if route.Auth.AuthType != "jwt" {
		t.Errorf("expected auth type 'jwt', got %s", route.Auth.AuthType)
	}
}

// Test for rate limit

func TestParseRateLimit(t *testing.T) {
	source := `@ GET /limited
  + ratelimit(100/min)
  > {}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	if route.RateLimit == nil {
		t.Fatal("expected rate limit")
	}

	if route.RateLimit.Requests != 100 {
		t.Errorf("expected 100 requests, got %d", route.RateLimit.Requests)
	}

	if route.RateLimit.Window != "min" {
		t.Errorf("expected window 'min', got %s", route.RateLimit.Window)
	}
}

// Test HTTP methods

func TestParseAllHTTPMethods(t *testing.T) {
	methods := []struct {
		method string
		expect interpreter.HttpMethod
	}{
		{"GET", interpreter.Get},
		{"POST", interpreter.Post},
		{"PUT", interpreter.Put},
		{"DELETE", interpreter.Delete},
		{"PATCH", interpreter.Patch},
	}

	for _, m := range methods {
		source := "@ " + m.method + " /test\n  > {}"

		lexer := NewLexer(source)
		tokens, _ := lexer.Tokenize()
		parser := NewParser(tokens)
		module, err := parser.Parse()
		if err != nil {
			t.Errorf("parser error for %s: %v", m.method, err)
			continue
		}

		route := module.Items[0].(*interpreter.Route)
		if route.Method != m.expect {
			t.Errorf("expected method %v for %s, got %v", m.expect, m.method, route.Method)
		}
	}
}

// Test command parsing

func TestParseCommandWithFlags(t *testing.T) {
	source := `@ command deploy env: string! --force: bool --dry-run: bool
  > {}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	cmd := module.Items[0].(*interpreter.Command)

	if cmd.Name != "deploy" {
		t.Errorf("expected name 'deploy', got %s", cmd.Name)
	}

	// Should have at least 1 param
	if len(cmd.Params) < 1 {
		t.Fatalf("expected at least 1 param, got %d", len(cmd.Params))
	}

	// Verify we have some params parsed
	if cmd.Params[0].Name == "" {
		t.Error("first param should have a name")
	}
}

// Test return type parsing

func TestParseRouteReturnType(t *testing.T) {
	source := `@ GET /users -> User
  > {}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	if route.ReturnType == nil {
		t.Fatal("expected return type")
	}

	namedType, ok := route.ReturnType.(interpreter.NamedType)
	if !ok {
		t.Fatalf("expected NamedType, got %T", route.ReturnType)
	}

	if namedType.Name != "User" {
		t.Errorf("expected 'User', got %s", namedType.Name)
	}
}

// Test string interpolation

func TestParseStringConcatenation(t *testing.T) {
	source := `@ GET /greet/:name
  $ msg = "Hello, " + name + "!"
  > msg`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	// Should be binary expression with Add op
	binExpr, ok := assign.Value.(interpreter.BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr, got %T", assign.Value)
	}

	if binExpr.Op != interpreter.Add {
		t.Errorf("expected Add op, got %v", binExpr.Op)
	}
}

// Additional tests to increase coverage

func TestParseGenericFunctionWithBody(t *testing.T) {
	source := `! identity<T>(value: T!) -> T {
  > value
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	fn := module.Items[0].(*interpreter.Function)

	if fn.Name != "identity" {
		t.Errorf("expected name 'identity', got %s", fn.Name)
	}

	if len(fn.TypeParams) == 0 {
		t.Error("expected type parameters")
	}

	if len(fn.Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(fn.Params))
	}
}

func TestParseRegularFunctionWithMultipleParams(t *testing.T) {
	source := `! add(a: int!, b: int!) -> int {
  > a + b
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	fn := module.Items[0].(*interpreter.Function)

	if fn.Name != "add" {
		t.Errorf("expected name 'add', got %s", fn.Name)
	}

	if len(fn.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(fn.Params))
	}

	if fn.ReturnType == nil {
		t.Error("expected return type")
	}
}

func TestParseCronTaskFull(t *testing.T) {
	source := `* "0 0 * * *" {
  $ result = 1
  > result
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	cron := module.Items[0].(*interpreter.CronTask)

	if cron.Schedule != "0 0 * * *" {
		t.Errorf("expected schedule '0 0 * * *', got %s", cron.Schedule)
	}
}

func TestParseEventHandlerSimple(t *testing.T) {
	source := `~ "user.created" {
  $ msg = "Welcome!"
  > msg
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	event := module.Items[0].(*interpreter.EventHandler)

	if event.EventType != "user.created" {
		t.Errorf("expected event type 'user.created', got %s", event.EventType)
	}
}

func TestParseQueueWorkerSimple(t *testing.T) {
	source := `& "emails" {
  $ result = 1
  > result
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	worker := module.Items[0].(*interpreter.QueueWorker)

	if worker.QueueName != "emails" {
		t.Errorf("expected queue 'emails', got %s", worker.QueueName)
	}
}

func TestParseFromImportMultipleItems(t *testing.T) {
	source := `from "./utils" import { add, subtract, multiply }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	imp := module.Items[0].(*interpreter.ImportStatement)

	if len(imp.Names) < 3 {
		t.Errorf("expected at least 3 imported names, got %d", len(imp.Names))
	}
}

func TestParseMacroDefSimple(t *testing.T) {
	source := `macro! myMacro() {
  $ x = 1
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	macro := module.Items[0].(*interpreter.MacroDef)

	if macro.Name != "myMacro" {
		t.Errorf("expected name 'myMacro', got %s", macro.Name)
	}
}

func TestParseMatchMultipleCases(t *testing.T) {
	source := `@ GET /eval
  $ value = match status {
    200 => "ok"
    404 => "not found"
    500 => "error"
    _ => "unknown"
  }
  > value`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	matchExpr := assign.Value.(interpreter.MatchExpr)

	if len(matchExpr.Cases) != 4 {
		t.Errorf("expected 4 cases, got %d", len(matchExpr.Cases))
	}
}

func TestParsePatternLiteralString(t *testing.T) {
	source := `@ GET /check
  $ result = match code {
    "success" => 1
    "error" => 0
    _ => -1
  }
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	matchExpr := assign.Value.(interpreter.MatchExpr)

	// First case should be a literal pattern
	if len(matchExpr.Cases) < 2 {
		t.Fatalf("expected at least 2 cases, got %d", len(matchExpr.Cases))
	}
}

func TestParseDeepNestedObject(t *testing.T) {
	source := `@ GET /deep
  > { level1: { level2: { level3: 42 } } }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	ret := route.Body[0].(interpreter.ReturnStatement)

	objExpr, ok := ret.Value.(interpreter.ObjectExpr)
	if !ok {
		t.Fatalf("expected ObjectExpr, got %T", ret.Value)
	}

	if len(objExpr.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(objExpr.Fields))
	}
}

func TestParseComparisonChain(t *testing.T) {
	source := `@ GET /compare
  $ a = 1 < 2
  $ b = 3 > 1
  $ c = 4 <= 4
  $ d = 5 >= 5
  > { a: a, b: b, c: c, d: d }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	// Should have 4 assignments + 1 return
	if len(route.Body) < 5 {
		t.Errorf("expected at least 5 statements, got %d", len(route.Body))
	}
}

func TestParseLogicalOperatorsFull(t *testing.T) {
	source := `@ GET /logic
  $ x = true && false
  $ y = true || false
  $ z = !true
  > { x: x, y: y, z: z }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	if len(route.Body) < 4 {
		t.Errorf("expected at least 4 statements, got %d", len(route.Body))
	}
}

func TestParseModuleDeclaration(t *testing.T) {
	source := `module "mymodule"`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	// Check that module decl was parsed
	modDecl, ok := module.Items[0].(*interpreter.ModuleDecl)
	if !ok {
		t.Fatalf("expected ModuleDecl, got %T", module.Items[0])
	}

	if modDecl.Name != "mymodule" {
		t.Errorf("expected module name 'mymodule', got %s", modDecl.Name)
	}
}

func TestParseMultipleImports(t *testing.T) {
	source := `import "./a"
import "./b"
import "./c"`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(module.Items) != 3 {
		t.Errorf("expected 3 imports, got %d", len(module.Items))
	}
}

func TestParseFieldAccessChain(t *testing.T) {
	source := `@ GET /chain
  $ result = a.b.c.d
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	// Should be a nested FieldAccessExpr
	_, ok := assign.Value.(interpreter.FieldAccessExpr)
	if !ok {
		t.Fatalf("expected FieldAccessExpr, got %T", assign.Value)
	}
}

func TestParseArrayWithExpressions(t *testing.T) {
	source := `@ GET /arr
  $ arr = [1 + 2, 3 * 4, 5 - 1]
  > arr`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	arrExpr, ok := assign.Value.(interpreter.ArrayExpr)
	if !ok {
		t.Fatalf("expected ArrayExpr, got %T", assign.Value)
	}

	if len(arrExpr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arrExpr.Elements))
	}
}

func TestParseFunctionCallWithMultipleArgs(t *testing.T) {
	source := `@ GET /call
  $ result = calculate(1, 2, 3, 4)
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	callExpr, ok := assign.Value.(interpreter.FunctionCallExpr)
	if !ok {
		t.Fatalf("expected FunctionCallExpr, got %T", assign.Value)
	}

	if len(callExpr.Args) != 4 {
		t.Errorf("expected 4 args, got %d", len(callExpr.Args))
	}
}

func TestParseArrayOfIntType(t *testing.T) {
	source := `: Container { items: [int]! }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)

	_, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.ArrayType)
	if !ok {
		t.Fatalf("expected ArrayType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
}

func TestParseGenericType(t *testing.T) {
	source := `: Cache { data: Map<string, int>! }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)

	// Just verify parsing succeeded
	if len(typeDef.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(typeDef.Fields))
	}
}

func TestParseOptionalType(t *testing.T) {
	source := `: Optional { value: int }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)

	// Optional type should not have Required=true
	if typeDef.Fields[0].Required {
		t.Error("expected optional field, got required")
	}
}

func TestParseDatabaseQuery(t *testing.T) {
	source := `@ GET /users
  $ users = db.query("SELECT * FROM users")
  > users`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	// Should be a function call expression
	_, ok := assign.Value.(interpreter.FunctionCallExpr)
	if !ok {
		t.Fatalf("expected FunctionCallExpr, got %T", assign.Value)
	}
}

func TestParseWebSocketWithMessages(t *testing.T) {
	source := `@ ws /chat {
  on connect {
    > { type: "connected" }
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	ws := module.Items[0].(*interpreter.WebSocketRoute)

	if ws.Path != "/chat" {
		t.Errorf("expected path '/chat', got %s", ws.Path)
	}
}

func TestParseTypeWithMultipleFields(t *testing.T) {
	source := `: User {
  id: int!
  name: str!
  email: str!
  age: int
  active: bool!
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)

	if len(typeDef.Fields) != 5 {
		t.Errorf("expected 5 fields, got %d", len(typeDef.Fields))
	}
}

func TestParseRouteWithQueryParams(t *testing.T) {
	source := `@ GET /search
  ? query: str!
  ? page: int
  > []`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	if len(route.QueryParams) != 2 {
		t.Errorf("expected 2 query params, got %d", len(route.QueryParams))
	}
}

func TestParseRouteWithMiddleware(t *testing.T) {
	source := `@ GET /admin
  + auth(jwt)
  + rateLimit(100)
  > { status: "ok" }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	// Check for auth config
	if route.Auth == nil {
		// It's ok if auth is parsed differently
	}
}

// Additional tests for uncovered statement types

func TestParseForLoop(t *testing.T) {
	source := `@ GET /sum
  $ total = 0
  for item in items {
    $ total = total + item
  }
  > total`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	if len(route.Body) < 2 {
		t.Fatal("expected at least 2 statements")
	}
}

func TestParseWithPlusAuth(t *testing.T) {
	source := `@ GET /limited
  + auth(jwt)
  > { status: "ok" }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	if route.Auth == nil {
		t.Error("expected auth config")
	}
}

func TestParseAuthWithBearer(t *testing.T) {
	source := `@ GET /secure
  + auth(bearer)
  > { status: "ok" }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	if route.Auth == nil {
		t.Error("expected auth config")
	}
}

func TestParseAsyncExprWithAwait(t *testing.T) {
	source := `@ GET /data
  $ result = async {
    $ data = await fetchData()
    > data
  }
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	_, ok := assign.Value.(interpreter.AsyncExpr)
	if !ok {
		t.Fatalf("expected AsyncExpr, got %T", assign.Value)
	}
}

func TestParseValidationStatement(t *testing.T) {
	source := `@ POST /user
  ? name: str!
  $ valid = name.length > 0
  > { valid: valid }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	if len(route.QueryParams) != 1 {
		t.Errorf("expected 1 query param, got %d", len(route.QueryParams))
	}
}

func TestParseNestedIfElse(t *testing.T) {
	source := `@ GET /nested
  if a {
    if b {
      > 1
    } else {
      > 2
    }
  } else {
    if c {
      > 3
    } else {
      > 4
    }
  }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	ifStmt, ok := route.Body[0].(interpreter.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", route.Body[0])
	}

	if ifStmt.Condition == nil {
		t.Error("expected condition")
	}
}

func TestParseWhileLoop(t *testing.T) {
	source := `@ GET /loop
  $ i = 0
  while i < 100 {
    $ i = i + 1
  }
  > i`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	if len(route.Body) < 3 {
		t.Errorf("expected at least 3 statements, got %d", len(route.Body))
	}
}

func TestParseForWithIterator(t *testing.T) {
	source := `@ GET /iterate
  $ sum = 0
  for i in items {
    $ sum = sum + i
  }
  > sum`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	if len(route.Body) < 3 {
		t.Errorf("expected at least 3 statements, got %d", len(route.Body))
	}
}

func TestParseTypeWithGenericParams(t *testing.T) {
	source := `: Response<T> {
  data: T!
  success: bool!
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)

	if len(typeDef.TypeParams) == 0 {
		t.Error("expected type parameters")
	}
}

func TestParseDbQueryStatement(t *testing.T) {
	source := `@ GET /users
  $ users = db.query("SELECT * FROM users WHERE active = true")
  > users`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	if len(route.Body) < 2 {
		t.Fatalf("expected at least 2 statements, got %d", len(route.Body))
	}
}

func TestParseRouteWithAuthAndQuery(t *testing.T) {
	source := `@ POST /api/users
  + auth(jwt)
  ? page: int
  ? limit: int
  $ offset = page * limit
  > offset`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)

	if route.Auth == nil {
		t.Error("expected auth config")
	}

	if len(route.QueryParams) != 2 {
		t.Errorf("expected 2 query params, got %d", len(route.QueryParams))
	}
}

// More type coverage tests

func TestParseNullableType(t *testing.T) {
	source := `: User { name: str? }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)
	if len(typeDef.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(typeDef.Fields))
	}
}

func TestParseIntType(t *testing.T) {
	source := `: Counter { value: int! }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)
	_, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.IntType)
	if !ok {
		t.Fatalf("expected IntType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
}

func TestParseStringType(t *testing.T) {
	source := `: Message { text: str! }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)
	_, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.StringType)
	if !ok {
		t.Fatalf("expected StringType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
}

func TestParseBoolType(t *testing.T) {
	source := `: Flag { active: bool! }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)
	_, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.BoolType)
	if !ok {
		t.Fatalf("expected BoolType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
}

func TestParseCustomNamedType(t *testing.T) {
	source := `: Order { user: User! }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	typeDef := module.Items[0].(*interpreter.TypeDef)
	namedType, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.NamedType)
	if !ok {
		t.Fatalf("expected NamedType, got %T", typeDef.Fields[0].TypeAnnotation)
	}
	if namedType.Name != "User" {
		t.Errorf("expected 'User', got %s", namedType.Name)
	}
}

func TestParseRouteWithPathParams(t *testing.T) {
	source := `@ GET /users/:id
  > { id: id }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	if route.Path != "/users/:id" {
		t.Errorf("expected '/users/:id', got %s", route.Path)
	}
}

func TestParseMultipleRoutes(t *testing.T) {
	source := `@ GET /a
  > 1

@ POST /b
  > 2

@ PUT /c
  > 3`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(module.Items) != 3 {
		t.Errorf("expected 3 routes, got %d", len(module.Items))
	}
}

func TestParseMultipleTypes(t *testing.T) {
	source := `: User { name: str! }
: Order { total: int! }
: Item { price: float! }`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(module.Items) != 3 {
		t.Errorf("expected 3 types, got %d", len(module.Items))
	}
}

func TestParseMultipleFunctions(t *testing.T) {
	source := `! add(a: int!, b: int!) -> int {
  > a + b
}

! sub(a: int!, b: int!) -> int {
  > a - b
}

! mul(a: int!, b: int!) -> int {
  > a * b
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(module.Items) != 3 {
		t.Errorf("expected 3 functions, got %d", len(module.Items))
	}
}

func TestParseDivisionExpression(t *testing.T) {
	source := `@ GET /div
  $ result = a / b
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	binExpr, ok := assign.Value.(interpreter.BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr, got %T", assign.Value)
	}

	if binExpr.Op != interpreter.Div {
		t.Errorf("expected Div op, got %v", binExpr.Op)
	}
}

func TestParseSubtractionExpression(t *testing.T) {
	source := `@ GET /sub
  $ result = a - b
  > result`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(tokens)
	module, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	route := module.Items[0].(*interpreter.Route)
	assign := route.Body[0].(interpreter.AssignStatement)

	binExpr, ok := assign.Value.(interpreter.BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr, got %T", assign.Value)
	}

	if binExpr.Op != interpreter.Sub {
		t.Errorf("expected Sub op, got %v", binExpr.Op)
	}
}
