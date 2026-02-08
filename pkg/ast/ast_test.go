package ast

import "testing"

func TestHttpMethod_String(t *testing.T) {
	tests := []struct {
		method   HttpMethod
		expected string
	}{
		{Get, "GET"},
		{Post, "POST"},
		{Put, "PUT"},
		{Delete, "DELETE"},
		{Patch, "PATCH"},
		{WebSocket, "WS"},
		{SSE, "SSE"},
		{HttpMethod(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.method.String()
			if got != tt.expected {
				t.Errorf("HttpMethod(%d).String() = %q, want %q", tt.method, got, tt.expected)
			}
		})
	}
}

func TestBinOp_String(t *testing.T) {
	tests := []struct {
		op       BinOp
		expected string
	}{
		{Add, "+"},
		{Sub, "-"},
		{Mul, "*"},
		{Div, "/"},
		{Eq, "=="},
		{Ne, "!="},
		{Lt, "<"},
		{Le, "<="},
		{Gt, ">"},
		{Ge, ">="},
		{And, "&&"},
		{Or, "||"},
		{BinOp(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.op.String()
			if got != tt.expected {
				t.Errorf("BinOp(%d).String() = %q, want %q", tt.op, got, tt.expected)
			}
		})
	}
}

func TestUnOp_String(t *testing.T) {
	tests := []struct {
		op       UnOp
		expected string
	}{
		{Not, "!"},
		{Neg, "-"},
		{UnOp(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.op.String()
			if got != tt.expected {
				t.Errorf("UnOp(%d).String() = %q, want %q", tt.op, got, tt.expected)
			}
		})
	}
}

func TestGRPCStreamType_String(t *testing.T) {
	tests := []struct {
		st       GRPCStreamType
		expected string
	}{
		{GRPCUnary, "unary"},
		{GRPCServerStream, "server_stream"},
		{GRPCClientStream, "client_stream"},
		{GRPCBidirectional, "bidirectional"},
		{GRPCStreamType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.st.String()
			if got != tt.expected {
				t.Errorf("GRPCStreamType(%d).String() = %q, want %q", tt.st, got, tt.expected)
			}
		})
	}
}

func TestGraphQLOperationType_String(t *testing.T) {
	tests := []struct {
		op       GraphQLOperationType
		expected string
	}{
		{GraphQLQuery, "query"},
		{GraphQLMutation, "mutation"},
		{GraphQLSubscription, "subscription"},
		{GraphQLOperationType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.op.String()
			if got != tt.expected {
				t.Errorf("GraphQLOperationType(%d).String() = %q, want %q", tt.op, got, tt.expected)
			}
		})
	}
}

// Test interface implementations to ensure all types satisfy their interfaces.

func TestItem_Interface(t *testing.T) {
	items := []Item{
		TypeDef{},
		TraitDef{},
		Route{},
		Function{},
		WebSocketRoute{},
		Command{},
		CronTask{},
		EventHandler{},
		QueueWorker{},
		GRPCService{},
		GRPCHandler{},
		GraphQLResolver{},
		ImportStatement{},
		ModuleDecl{},
		ConstDecl{},
		MacroDef{},
		MacroInvocation{},
		ContractDef{},
		TestBlock{},
	}

	if len(items) == 0 {
		t.Error("expected items to be non-empty")
	}
}

func TestStatement_Interface(t *testing.T) {
	stmts := []Statement{
		AssignStatement{},
		ReassignStatement{},
		IndexAssignStatement{},
		DbQueryStatement{},
		ReturnStatement{},
		IfStatement{},
		WhileStatement{},
		SwitchStatement{},
		ForStatement{},
		WsSendStatement{},
		WsBroadcastStatement{},
		WsCloseStatement{},
		ValidationStatement{},
		ExpressionStatement{},
		YieldStatement{},
		WebSocketEvent{},
		MacroInvocation{},
		AssertStatement{},
	}

	if len(stmts) == 0 {
		t.Error("expected statements to be non-empty")
	}
}

func TestExpr_Interface(t *testing.T) {
	exprs := []Expr{
		LiteralExpr{},
		VariableExpr{},
		BinaryOpExpr{},
		UnaryOpExpr{},
		FieldAccessExpr{},
		ArrayIndexExpr{},
		FunctionCallExpr{},
		ObjectExpr{},
		ArrayExpr{},
		LambdaExpr{},
		PipeExpr{},
		AsyncExpr{},
		AwaitExpr{},
		MatchExpr{},
		QuoteExpr{},
		UnquoteExpr{},
		MacroInvocation{},
	}

	if len(exprs) == 0 {
		t.Error("expected expressions to be non-empty")
	}
}

func TestType_Interface(t *testing.T) {
	types := []Type{
		IntType{},
		StringType{},
		BoolType{},
		FloatType{},
		ArrayType{},
		OptionalType{},
		NamedType{},
		DatabaseType{},
		RedisType{},
		MongoDBType{},
		LLMType{},
		UnionType{},
		GenericType{},
		TypeParameterType{},
		FunctionType{},
		FutureType{},
	}

	if len(types) == 0 {
		t.Error("expected types to be non-empty")
	}
}

func TestPattern_Interface(t *testing.T) {
	patterns := []Pattern{
		LiteralPattern{},
		VariablePattern{},
		WildcardPattern{},
		ObjectPattern{},
		ArrayPattern{},
	}

	if len(patterns) == 0 {
		t.Error("expected patterns to be non-empty")
	}
}

func TestLiteral_Interface(t *testing.T) {
	literals := []Literal{
		IntLiteral{Value: 42},
		StringLiteral{Value: "hello"},
		BoolLiteral{Value: true},
		FloatLiteral{Value: 3.14},
		NullLiteral{},
	}

	if len(literals) == 0 {
		t.Error("expected literals to be non-empty")
	}
}
