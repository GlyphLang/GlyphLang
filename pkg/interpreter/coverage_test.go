package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

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

func TestLoadModuleWithStaticRoute(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&StaticRoute{
				Path:    "/assets",
				RootDir: "./public",
			},
		},
	}

	err := interp.LoadModule(module)
	if err != nil {
		t.Errorf("LoadModule with static route failed: %v", err)
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

// Compile-time interface satisfaction checks for AST types.
// These replace the previous marker method tests which cannot call
// unexported methods across package boundaries after the AST extraction.
var (
	// Type interface
	_ Type = DatabaseType{}
	_ Type = ArrayType{}
	_ Type = OptionalType{}
	_ Type = NamedType{}
	_ Type = IntType{}
	_ Type = FloatType{}
	_ Type = StringType{}
	_ Type = BoolType{}
	_ Type = UnionType{}

	// Item interface
	_ Item = Route{}
	_ Item = Function{}
	_ Item = TypeDef{}
	_ Item = Command{}
	_ Item = CronTask{}
	_ Item = EventHandler{}
	_ Item = QueueWorker{}
	_ Item = WebSocketRoute{}

	// Statement interface
	_ Statement = AssignStatement{}
	_ Statement = ReturnStatement{}
	_ Statement = IfStatement{}
	_ Statement = WhileStatement{}
	_ Statement = ForStatement{}
	_ Statement = SwitchStatement{}
	_ Statement = DbQueryStatement{}
	_ Statement = ValidationStatement{}
	_ Statement = WebSocketEvent{}
	_ Statement = WsSendStatement{}
	_ Statement = WsBroadcastStatement{}
	_ Statement = WsCloseStatement{}

	// Expr interface
	_ Expr = LiteralExpr{}
	_ Expr = VariableExpr{}
	_ Expr = BinaryOpExpr{}
	_ Expr = UnaryOpExpr{}
	_ Expr = FieldAccessExpr{}
	_ Expr = ArrayIndexExpr{}
	_ Expr = FunctionCallExpr{}
	_ Expr = ObjectExpr{}
	_ Expr = ArrayExpr{}

	// Literal interface
	_ Literal = IntLiteral{}
	_ Literal = FloatLiteral{}
	_ Literal = StringLiteral{}
	_ Literal = BoolLiteral{}
	_ Literal = NullLiteral{}
)
