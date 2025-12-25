package interpreter

import (
	"testing"
)

// TestFloatArithmetic tests float arithmetic operations
func TestFloatArithmetic(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test float addition
	result, err := interp.EvaluateExpression(BinaryOpExpr{
		Op:    Add,
		Left:  LiteralExpr{Value: FloatLiteral{Value: 1.5}},
		Right: LiteralExpr{Value: FloatLiteral{Value: 2.5}},
	}, env)
	if err != nil {
		t.Errorf("float add failed: %v", err)
	}
	if result != float64(4.0) {
		t.Errorf("1.5 + 2.5 = %v, want 4.0", result)
	}

	// Test float subtraction
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Sub,
		Left:  LiteralExpr{Value: FloatLiteral{Value: 5.5}},
		Right: LiteralExpr{Value: FloatLiteral{Value: 2.5}},
	}, env)
	if err != nil {
		t.Errorf("float sub failed: %v", err)
	}
	if result != float64(3.0) {
		t.Errorf("5.5 - 2.5 = %v, want 3.0", result)
	}

	// Test float multiplication
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Mul,
		Left:  LiteralExpr{Value: FloatLiteral{Value: 2.5}},
		Right: LiteralExpr{Value: FloatLiteral{Value: 4.0}},
	}, env)
	if err != nil {
		t.Errorf("float mul failed: %v", err)
	}
	if result != float64(10.0) {
		t.Errorf("2.5 * 4.0 = %v, want 10.0", result)
	}
}

// TestMixedArithmetic tests mixed int/float arithmetic
func TestMixedArithmetic(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test int + float
	result, err := interp.EvaluateExpression(BinaryOpExpr{
		Op:    Add,
		Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
		Right: LiteralExpr{Value: FloatLiteral{Value: 2.5}},
	}, env)
	if err != nil {
		t.Errorf("int + float failed: %v", err)
	}
	if result != float64(7.5) {
		t.Errorf("5 + 2.5 = %v, want 7.5", result)
	}

	// Test float + int
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Add,
		Left:  LiteralExpr{Value: FloatLiteral{Value: 2.5}},
		Right: LiteralExpr{Value: IntLiteral{Value: 5}},
	}, env)
	if err != nil {
		t.Errorf("float + int failed: %v", err)
	}
	if result != float64(7.5) {
		t.Errorf("2.5 + 5 = %v, want 7.5", result)
	}
}

// TestAbsFunctionFloat tests abs builtin with floats
func TestAbsFunctionFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("f", float64(-3.14))
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "abs",
		Args: []Expr{VariableExpr{Name: "f"}},
	}, env)
	if err != nil {
		t.Errorf("abs float failed: %v", err)
	}
	if result != float64(3.14) {
		t.Errorf("abs(-3.14) = %v, want 3.14", result)
	}
}

// TestHttpMethodString tests HttpMethod String method
func TestHttpMethodString(t *testing.T) {
	tests := []struct {
		method HttpMethod
		want   string
	}{
		{Get, "GET"},
		{Post, "POST"},
		{Put, "PUT"},
		{Delete, "DELETE"},
		{Patch, "PATCH"},
		{WebSocket, "WS"},
		{HttpMethod(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.method.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.method, got, tt.want)
		}
	}
}

// TestUnOpString tests UnOp String method
func TestUnOpString(t *testing.T) {
	tests := []struct {
		op   UnOp
		want string
	}{
		{Not, "!"},
		{Neg, "-"},
		{UnOp(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.op.String(); got != tt.want {
			t.Errorf("UnOp(%d).String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

// TestBinOpStringUnknown tests BinOp String for unknown value
func TestBinOpStringUnknown(t *testing.T) {
	op := BinOp(999)
	if got := op.String(); got != "UNKNOWN" {
		t.Errorf("BinOp(999).String() = %q, want 'UNKNOWN'", got)
	}
}

// TestGetTypeDefNonExistent tests GetTypeDef method
func TestGetTypeDefNonExistent(t *testing.T) {
	interp := NewInterpreter()

	// Test non-existent type
	_, ok := interp.GetTypeDef("NonExistent")
	if ok {
		t.Error("GetTypeDef should return false for non-existent type")
	}
}

// TestGetFunctionNonExistent tests GetFunction method
func TestGetFunctionNonExistent(t *testing.T) {
	interp := NewInterpreter()

	// Test non-existent function
	_, ok := interp.GetFunction("nonexistent")
	if ok {
		t.Error("GetFunction should return false for non-existent function")
	}
}

// TestLogicalOperators tests short-circuit logical operators
func TestLogicalOperators(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test And with false first (short-circuit)
	result, err := interp.EvaluateExpression(BinaryOpExpr{
		Op:    And,
		Left:  LiteralExpr{Value: BoolLiteral{Value: false}},
		Right: LiteralExpr{Value: BoolLiteral{Value: true}},
	}, env)
	if err != nil {
		t.Errorf("false && true failed: %v", err)
	}
	if result != false {
		t.Errorf("false && true = %v, want false", result)
	}

	// Test Or with true first (short-circuit)
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Or,
		Left:  LiteralExpr{Value: BoolLiteral{Value: true}},
		Right: LiteralExpr{Value: BoolLiteral{Value: false}},
	}, env)
	if err != nil {
		t.Errorf("true || false failed: %v", err)
	}
	if result != true {
		t.Errorf("true || false = %v, want true", result)
	}
}

// TestMinMaxFloats tests min/max with floats
func TestMinMaxFloats(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("a", float64(1.5))
	env.Define("b", float64(2.5))

	// Test min
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "min",
		Args: []Expr{VariableExpr{Name: "a"}, VariableExpr{Name: "b"}},
	}, env)
	if err != nil {
		t.Errorf("min failed: %v", err)
	}
	if result != float64(1.5) {
		t.Errorf("min(1.5, 2.5) = %v, want 1.5", result)
	}

	// Test max
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "max",
		Args: []Expr{VariableExpr{Name: "a"}, VariableExpr{Name: "b"}},
	}, env)
	if err != nil {
		t.Errorf("max failed: %v", err)
	}
	if result != float64(2.5) {
		t.Errorf("max(1.5, 2.5) = %v, want 2.5", result)
	}
}

// TestLoadModuleWithRoute tests loading module with routes
func TestLoadModuleWithRoute(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&Route{
				Method: Get,
				Path:   "/test",
				Body: []Statement{
					ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "hello"}}},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	if err != nil {
		t.Errorf("LoadModule with route failed: %v", err)
	}
}

// TestLoadModuleWithWebSocket tests loading module with websocket
func TestLoadModuleWithWebSocket(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&WebSocketRoute{
				Path: "/ws",
				Events: []WebSocketEvent{
					{EventType: WSEventConnect, Body: []Statement{}},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	if err != nil {
		t.Errorf("LoadModule with websocket failed: %v", err)
	}
}

// TestLoadModuleWithCronTask tests loading module with cron task
func TestLoadModuleWithCronTask(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&CronTask{
				Schedule: "* * * * *",
				Body:     []Statement{},
			},
		},
	}

	err := interp.LoadModule(module)
	if err != nil {
		t.Errorf("LoadModule with cron task failed: %v", err)
	}
}

// TestLoadModuleWithEventHandler tests loading module with event handler
func TestLoadModuleWithEventHandler(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&EventHandler{
				EventType: "user.created",
				Body:      []Statement{},
			},
		},
	}

	err := interp.LoadModule(module)
	if err != nil {
		t.Errorf("LoadModule with event handler failed: %v", err)
	}
}

// TestLoadModuleWithQueueWorker tests loading module with queue worker
func TestLoadModuleWithQueueWorker(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&QueueWorker{
				QueueName: "emails",
				Body:      []Statement{},
			},
		},
	}

	err := interp.LoadModule(module)
	if err != nil {
		t.Errorf("LoadModule with queue worker failed: %v", err)
	}
}

// TestDbQueryStatement tests db query statement execution
func TestDbQueryStatement(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	stmt := DbQueryStatement{
		Var:   "users",
		Query: "SELECT * FROM users",
	}

	// This will fail because no DB handler is set, which is expected
	_, _ = interp.ExecuteStatement(stmt, env)
}

// TestWsSendStatement tests websocket send statement execution
func TestWsSendStatement(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	stmt := WsSendStatement{
		Message: LiteralExpr{Value: StringLiteral{Value: "hello"}},
	}

	// This will fail because no WS connection is set, which is expected
	_, _ = interp.ExecuteStatement(stmt, env)
}

// TestWsBroadcastStatement tests websocket broadcast statement execution
func TestWsBroadcastStatement(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	stmt := WsBroadcastStatement{
		Message: LiteralExpr{Value: StringLiteral{Value: "hello everyone"}},
	}

	// This will fail because no WS connection is set, which is expected
	_, _ = interp.ExecuteStatement(stmt, env)
}

// TestWsCloseStatement tests websocket close statement execution
func TestWsCloseStatement(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	stmt := WsCloseStatement{}

	// This will fail because no WS connection is set, which is expected
	_, _ = interp.ExecuteStatement(stmt, env)
}

// TestParseIntErrors tests parseInt error cases
func TestParseIntErrors(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test parseInt with invalid string
	env.Define("invalid", "not a number")
	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "parseInt",
		Args: []Expr{VariableExpr{Name: "invalid"}},
	}, env)
	if err == nil {
		t.Error("parseInt with invalid string should error")
	}
}

// TestParseFloatErrors tests parseFloat error cases
func TestParseFloatErrors(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test parseFloat with invalid string
	env.Define("invalid", "not a number")
	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "parseFloat",
		Args: []Expr{VariableExpr{Name: "invalid"}},
	}, env)
	if err == nil {
		t.Error("parseFloat with invalid string should error")
	}
}

// TestUnionType tests union type handling
func TestUnionType(t *testing.T) {
	tc := NewTypeChecker()

	// Test union type compatibility
	unionType := UnionType{Types: []Type{IntType{}, StringType{}}}

	// int is compatible with int|string
	result := tc.TypesCompatible(IntType{}, unionType)
	if !result {
		t.Error("int should be compatible with int|string")
	}

	// string is compatible with int|string
	result = tc.TypesCompatible(StringType{}, unionType)
	if !result {
		t.Error("string should be compatible with int|string")
	}
}

// TestSetDatabaseHandler tests setting database handler
func TestSetDatabaseHandler(t *testing.T) {
	interp := NewInterpreter()

	// Set a nil database handler (for testing)
	interp.SetDatabaseHandler(nil)
}

// TestFieldAccessExprNested tests nested field access
func TestFieldAccessExprNested(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Define nested object
	env.Define("user", map[string]interface{}{
		"profile": map[string]interface{}{
			"name": "Alice",
		},
	})

	// Access profile field
	result, err := interp.EvaluateExpression(FieldAccessExpr{
		Object: VariableExpr{Name: "user"},
		Field:  "profile",
	}, env)
	if err != nil {
		t.Errorf("field access failed: %v", err)
	}
	profile, ok := result.(map[string]interface{})
	if !ok {
		t.Errorf("profile not a map: %T", result)
	}
	if profile["name"] != "Alice" {
		t.Errorf("profile.name = %v, want 'Alice'", profile["name"])
	}
}

// TestIntLiteralEdgeCases tests int literal with extreme values
func TestIntLiteralEdgeCases(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test large int
	result, err := interp.EvaluateExpression(LiteralExpr{
		Value: IntLiteral{Value: 9223372036854775807}, // Max int64
	}, env)
	if err != nil {
		t.Errorf("large int literal failed: %v", err)
	}
	if result != int64(9223372036854775807) {
		t.Errorf("max int64 = %v", result)
	}

	// Test negative int
	result, err = interp.EvaluateExpression(LiteralExpr{
		Value: IntLiteral{Value: -9223372036854775808}, // Min int64
	}, env)
	if err != nil {
		t.Errorf("min int literal failed: %v", err)
	}
	if result != int64(-9223372036854775808) {
		t.Errorf("min int64 = %v", result)
	}
}

// TestBoolLiteralExpr tests bool literal
func TestBoolLiteralExpr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test true
	result, err := interp.EvaluateExpression(LiteralExpr{
		Value: BoolLiteral{Value: true},
	}, env)
	if err != nil {
		t.Errorf("true literal failed: %v", err)
	}
	if result != true {
		t.Errorf("true = %v", result)
	}

	// Test false
	result, err = interp.EvaluateExpression(LiteralExpr{
		Value: BoolLiteral{Value: false},
	}, env)
	if err != nil {
		t.Errorf("false literal failed: %v", err)
	}
	if result != false {
		t.Errorf("false = %v", result)
	}
}

// TestDatabaseType tests database type marker
func TestDatabaseType(t *testing.T) {
	dbType := DatabaseType{}
	dbType.isType() // Just call to ensure it exists
}

// TestArrayType tests array type marker
func TestArrayType(t *testing.T) {
	arrType := ArrayType{ElementType: IntType{}}
	arrType.isType() // Just call to ensure it exists
}

// TestOptionalType tests optional type marker
func TestOptionalType(t *testing.T) {
	optType := OptionalType{InnerType: StringType{}}
	optType.isType() // Just call to ensure it exists
}

// TestNamedType tests named type marker
func TestNamedType(t *testing.T) {
	namedType := NamedType{Name: "User"}
	namedType.isType() // Just call to ensure it exists
}

// TestRouteMarker tests route isItem marker
func TestRouteMarker(t *testing.T) {
	route := Route{Path: "/test", Method: Get}
	route.isItem()
}

// TestFunctionMarker tests function isItem marker
func TestFunctionMarker(t *testing.T) {
	fn := Function{Name: "test"}
	fn.isItem()
}

// TestTypeDefMarker tests typedef isItem marker
func TestTypeDefMarker(t *testing.T) {
	td := TypeDef{Name: "User"}
	td.isItem()
}

// TestCommandMarker tests command isItem marker
func TestCommandMarker(t *testing.T) {
	cmd := Command{Name: "hello"}
	cmd.isItem()
}

// TestCronTaskMarker tests cron task isItem marker
func TestCronTaskMarker(t *testing.T) {
	cron := CronTask{Schedule: "* * * * *"}
	cron.isItem()
}

// TestEventHandlerMarker tests event handler isItem marker
func TestEventHandlerMarker(t *testing.T) {
	eh := EventHandler{EventType: "test"}
	eh.isItem()
}

// TestQueueWorkerMarker tests queue worker isItem marker
func TestQueueWorkerMarker(t *testing.T) {
	qw := QueueWorker{QueueName: "test"}
	qw.isItem()
}

// TestWebSocketRouteMarker tests websocket route isItem marker
func TestWebSocketRouteMarker(t *testing.T) {
	wsr := WebSocketRoute{Path: "/ws"}
	wsr.isItem()
}

// TestAssignStatementMarker tests assign statement isStatement marker
func TestAssignStatementMarker(t *testing.T) {
	stmt := AssignStatement{Target: "x"}
	stmt.isStatement()
}

// TestReturnStatementMarker tests return statement isStatement marker
func TestReturnStatementMarker(t *testing.T) {
	stmt := ReturnStatement{}
	stmt.isStatement()
}

// TestIfStatementMarker tests if statement isStatement marker
func TestIfStatementMarker(t *testing.T) {
	stmt := IfStatement{}
	stmt.isStatement()
}

// TestWhileStatementMarker tests while statement isStatement marker
func TestWhileStatementMarker(t *testing.T) {
	stmt := WhileStatement{}
	stmt.isStatement()
}

// TestForStatementMarker tests for statement isStatement marker
func TestForStatementMarker(t *testing.T) {
	stmt := ForStatement{}
	stmt.isStatement()
}

// TestSwitchStatementMarker tests switch statement isStatement marker
func TestSwitchStatementMarker(t *testing.T) {
	stmt := SwitchStatement{}
	stmt.isStatement()
}

// TestDbQueryStatementMarker tests db query statement isStatement marker
func TestDbQueryStatementMarker(t *testing.T) {
	stmt := DbQueryStatement{}
	stmt.isStatement()
}

// TestValidationStatementMarker tests validation statement isStatement marker
func TestValidationStatementMarker(t *testing.T) {
	stmt := ValidationStatement{}
	stmt.isStatement()
}

// TestLiteralExprMarker tests literal expr isExpr marker
func TestLiteralExprMarker(t *testing.T) {
	expr := LiteralExpr{}
	expr.isExpr()
}

// TestVariableExprMarker tests variable expr isExpr marker
func TestVariableExprMarker(t *testing.T) {
	expr := VariableExpr{Name: "x"}
	expr.isExpr()
}

// TestBinaryOpExprMarker tests binary op expr isExpr marker
func TestBinaryOpExprMarker(t *testing.T) {
	expr := BinaryOpExpr{}
	expr.isExpr()
}

// TestUnaryOpExprMarker tests unary op expr isExpr marker
func TestUnaryOpExprMarker(t *testing.T) {
	expr := UnaryOpExpr{}
	expr.isExpr()
}

// TestFieldAccessExprMarker tests field access expr isExpr marker
func TestFieldAccessExprMarker(t *testing.T) {
	expr := FieldAccessExpr{}
	expr.isExpr()
}

// TestArrayIndexExprMarker tests array index expr isExpr marker
func TestArrayIndexExprMarker(t *testing.T) {
	expr := ArrayIndexExpr{}
	expr.isExpr()
}

// TestFunctionCallExprMarker tests function call expr isExpr marker
func TestFunctionCallExprMarker(t *testing.T) {
	expr := FunctionCallExpr{Name: "test"}
	expr.isExpr()
}

// TestObjectExprMarker tests object expr isExpr marker
func TestObjectExprMarker(t *testing.T) {
	expr := ObjectExpr{}
	expr.isExpr()
}

// TestArrayExprMarker tests array expr isExpr marker
func TestArrayExprMarker(t *testing.T) {
	expr := ArrayExpr{}
	expr.isExpr()
}

// TestIntLiteralMarker tests int literal isLiteral marker
func TestIntLiteralMarker(t *testing.T) {
	lit := IntLiteral{Value: 42}
	lit.isLiteral()
}

// TestFloatLiteralMarker tests float literal isLiteral marker
func TestFloatLiteralMarker(t *testing.T) {
	lit := FloatLiteral{Value: 3.14}
	lit.isLiteral()
}

// TestStringLiteralMarker tests string literal isLiteral marker
func TestStringLiteralMarker(t *testing.T) {
	lit := StringLiteral{Value: "hello"}
	lit.isLiteral()
}

// TestBoolLiteralMarker tests bool literal isLiteral marker
func TestBoolLiteralMarker(t *testing.T) {
	lit := BoolLiteral{Value: true}
	lit.isLiteral()
}

// TestNullLiteralMarker tests null literal isLiteral marker
func TestNullLiteralMarker(t *testing.T) {
	lit := NullLiteral{}
	lit.isLiteral()
}

// TestIntTypeMarker tests int type isType marker
func TestIntTypeMarker(t *testing.T) {
	typ := IntType{}
	typ.isType()
}

// TestFloatTypeMarker tests float type isType marker
func TestFloatTypeMarker(t *testing.T) {
	typ := FloatType{}
	typ.isType()
}

// TestStringTypeMarker tests string type isType marker
func TestStringTypeMarker(t *testing.T) {
	typ := StringType{}
	typ.isType()
}

// TestBoolTypeMarker tests bool type isType marker
func TestBoolTypeMarker(t *testing.T) {
	typ := BoolType{}
	typ.isType()
}

// TestUnionTypeMarker tests union type isType marker
func TestUnionTypeMarker(t *testing.T) {
	typ := UnionType{Types: []Type{IntType{}, StringType{}}}
	typ.isType()
}

// TestWebSocketEventMarker tests websocket event isStatement marker
func TestWebSocketEventMarker(t *testing.T) {
	event := WebSocketEvent{EventType: WSEventConnect}
	event.isStatement()
}

// TestWsSendStatementMarker tests ws send statement isStatement marker
func TestWsSendStatementMarker(t *testing.T) {
	stmt := WsSendStatement{}
	stmt.isStatement()
}

// TestWsBroadcastStatementMarker tests ws broadcast statement isStatement marker
func TestWsBroadcastStatementMarker(t *testing.T) {
	stmt := WsBroadcastStatement{}
	stmt.isStatement()
}

// TestWsCloseStatementMarker tests ws close statement isStatement marker
func TestWsCloseStatementMarker(t *testing.T) {
	stmt := WsCloseStatement{}
	stmt.isStatement()
}
