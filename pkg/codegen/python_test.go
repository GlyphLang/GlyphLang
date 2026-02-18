package codegen

import (
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/ir"
)

func TestPythonGenerateEmpty(t *testing.T) {
	gen := NewPythonGenerator("", 0)
	service := &ir.ServiceIR{}
	output := gen.Generate(service)

	if !strings.Contains(output, "from fastapi import FastAPI") {
		t.Error("expected FastAPI import")
	}
	if !strings.Contains(output, "app = FastAPI()") {
		t.Error("expected FastAPI app creation")
	}
	if !strings.Contains(output, "uvicorn.run") {
		t.Error("expected uvicorn.run in main block")
	}
}

func TestPythonGenerateModel(t *testing.T) {
	gen := NewPythonGenerator("", 0)
	service := &ir.ServiceIR{
		Types: []ir.TypeSchema{
			{
				Name: "User",
				Fields: []ir.FieldSchema{
					{Name: "id", Type: ir.TypeRef{Kind: ir.TypeInt}, Required: true},
					{Name: "name", Type: ir.TypeRef{Kind: ir.TypeString}, Required: true},
					{Name: "email", Type: ir.TypeRef{Kind: ir.TypeString}, Required: true},
					{Name: "age", Type: ir.TypeRef{Kind: ir.TypeOptional, Inner: &ir.TypeRef{Kind: ir.TypeInt}}},
				},
			},
		},
	}

	output := gen.Generate(service)

	if !strings.Contains(output, "class User(BaseModel):") {
		t.Error("expected User model class")
	}
	if !strings.Contains(output, "id: int") {
		t.Error("expected 'id: int' field")
	}
	if !strings.Contains(output, "name: str") {
		t.Error("expected 'name: str' field")
	}
	if !strings.Contains(output, "age: Optional[int] = None") {
		t.Error("expected optional age field")
	}
}

func TestPythonGenerateRoute(t *testing.T) {
	analyzer := ir.NewAnalyzer()
	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Method: ast.Get,
				Path:   "/api/users/:id",
				Injections: []ast.Injection{
					{Name: "db", Type: ast.DatabaseType{}},
				},
				Body: []ast.Statement{
					ast.AssignStatement{
						Target: "user",
						Value: ast.FunctionCallExpr{
							Name: "Get",
							Args: []ast.Expr{
								ast.FieldAccessExpr{
									Object: ast.VariableExpr{Name: "db"},
									Field:  "users",
								},
								ast.VariableExpr{Name: "id"},
							},
						},
					},
					ast.ReturnStatement{
						Value: ast.VariableExpr{Name: "user"},
					},
				},
			},
		},
	}

	service, err := analyzer.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gen := NewPythonGenerator("", 8000)
	output := gen.Generate(service)

	if !strings.Contains(output, `@app.get("/api/users/{id}")`) {
		t.Errorf("expected FastAPI route decorator, got:\n%s", output)
	}
	if !strings.Contains(output, "id: str") {
		t.Error("expected path parameter 'id: str'")
	}
	if !strings.Contains(output, "Depends(get_db)") {
		t.Error("expected database dependency injection")
	}
	if !strings.Contains(output, "user = db.users.Get(id)") {
		t.Error("expected variable assignment")
	}
	if !strings.Contains(output, "return user") {
		t.Error("expected return statement")
	}
}

func TestPythonGeneratePostRoute(t *testing.T) {
	analyzer := ir.NewAnalyzer()
	module := &ast.Module{
		Items: []ast.Item{
			&ast.TypeDef{
				Name: "CreateUser",
				Fields: []ast.Field{
					{Name: "name", TypeAnnotation: ast.StringType{}, Required: true},
					{Name: "email", TypeAnnotation: ast.StringType{}, Required: true},
				},
			},
			&ast.Route{
				Method:    ast.Post,
				Path:      "/api/users",
				InputType: ast.NamedType{Name: "CreateUser"},
				Injections: []ast.Injection{
					{Name: "db", Type: ast.DatabaseType{}},
				},
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
					},
				},
			},
		},
	}

	service, err := analyzer.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gen := NewPythonGenerator("", 8000)
	output := gen.Generate(service)

	if !strings.Contains(output, `@app.post("/api/users", status_code=201)`) {
		t.Errorf("expected POST decorator with 201 status, got:\n%s", output)
	}
	if !strings.Contains(output, "input: CreateUser") {
		t.Error("expected input body parameter with type")
	}
}

func TestPythonGenerateCronJob(t *testing.T) {
	analyzer := ir.NewAnalyzer()
	module := &ast.Module{
		Items: []ast.Item{
			&ast.CronTask{
				Name:     "cleanup",
				Schedule: "0 0 * * *",
				Injections: []ast.Injection{
					{Name: "db", Type: ast.DatabaseType{}},
				},
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "done"}},
					},
				},
			},
		},
	}

	service, err := analyzer.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gen := NewPythonGenerator("", 8000)
	output := gen.Generate(service)

	if !strings.Contains(output, "scheduler = AsyncIOScheduler()") {
		t.Error("expected scheduler setup")
	}
	if !strings.Contains(output, `CronTrigger.from_crontab("0 0 * * *")`) {
		t.Error("expected cron trigger")
	}
	if !strings.Contains(output, "async def cleanup()") {
		t.Error("expected cleanup function")
	}
}

func TestPythonGenerateRequirements(t *testing.T) {
	service := &ir.ServiceIR{
		Providers: []ir.ProviderRef{
			{ProviderType: "Database", IsStandard: true},
			{ProviderType: "Redis", IsStandard: true},
		},
		CronJobs: []ir.CronBinding{
			{Name: "test", Schedule: "* * * * *"},
		},
	}

	gen := NewPythonGenerator("", 8000)
	reqs := gen.GenerateRequirements(service)

	if !strings.Contains(reqs, "fastapi>=0.100.0") {
		t.Error("expected fastapi dependency")
	}
	if !strings.Contains(reqs, "sqlalchemy>=2.0.0") {
		t.Error("expected sqlalchemy dependency for Database provider")
	}
	if !strings.Contains(reqs, "redis>=5.0.0") {
		t.Error("expected redis dependency")
	}
	if !strings.Contains(reqs, "apscheduler>=3.10.0") {
		t.Error("expected apscheduler dependency for cron jobs")
	}
}

func TestPythonCustomProvider(t *testing.T) {
	service := &ir.ServiceIR{
		Providers: []ir.ProviderRef{
			{ProviderType: "ImageProcessor", IsStandard: false},
		},
		Routes: []ir.RouteHandler{
			{
				Method: ir.MethodPost,
				Path:   "/api/upload",
				Providers: []ir.InjectionRef{
					{Name: "images", ProviderType: "ImageProcessor"},
				},
				Body: []ir.StmtIR{
					{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{Value: ir.ExprIR{Kind: ir.ExprNull, IsNull: true}}},
				},
			},
		},
	}

	gen := NewPythonGenerator("", 8000)
	output := gen.Generate(service)

	if !strings.Contains(output, "class ImageProcessorProvider:") {
		t.Error("expected custom provider class")
	}
	if !strings.Contains(output, "Depends(get_imageprocessor)") {
		t.Error("expected custom provider dependency")
	}
}

func TestGlyphPathToFastAPI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/users", "/api/users"},
		{"/api/users/:id", "/api/users/{id}"},
		{"/api/users/:id/posts/:postId", "/api/users/{id}/posts/{postId}"},
	}

	for _, tt := range tests {
		result := glyphPathToFastAPI(tt.input)
		if result != tt.expected {
			t.Errorf("glyphPathToFastAPI(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestPythonSwitchStatement(t *testing.T) {
	gen := NewPythonGenerator("", 8000)
	service := &ir.ServiceIR{
		Routes: []ir.RouteHandler{
			{
				Method: ir.MethodGet,
				Path:   "/api/status",
				Body: []ir.StmtIR{
					{
						Kind: ir.StmtSwitch,
						Switch: &ir.SwitchStmt{
							Value: ir.ExprIR{Kind: ir.ExprVar, VarName: "code"},
							Cases: []ir.SwitchCase{
								{
									Value: ir.ExprIR{Kind: ir.ExprInt, IntVal: 200},
									Body: []ir.StmtIR{
										{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{Value: ir.ExprIR{Kind: ir.ExprString, StringVal: "ok"}}},
									},
								},
								{
									Value: ir.ExprIR{Kind: ir.ExprInt, IntVal: 404},
									Body: []ir.StmtIR{
										{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{Value: ir.ExprIR{Kind: ir.ExprString, StringVal: "not found"}}},
									},
								},
							},
							Default: []ir.StmtIR{
								{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{Value: ir.ExprIR{Kind: ir.ExprString, StringVal: "error"}}},
							},
						},
					},
				},
			},
		},
	}
	output := gen.Generate(service)

	if !strings.Contains(output, "if code == 200") {
		t.Error("expected first case as if statement")
	}
	if !strings.Contains(output, "elif code == 404") {
		t.Error("expected second case as elif statement")
	}
	if !strings.Contains(output, "else:") {
		t.Error("expected default as else block")
	}
}

func TestPythonPipeExpr(t *testing.T) {
	gen := NewPythonGenerator("", 8000)
	service := &ir.ServiceIR{
		Routes: []ir.RouteHandler{
			{
				Method: ir.MethodGet,
				Path:   "/api/pipe",
				Body: []ir.StmtIR{
					{
						Kind: ir.StmtReturn,
						Return: &ir.ReturnStmt{
							Value: ir.ExprIR{
								Kind: ir.ExprPipe,
								Pipe: &ir.PipeExpr{
									Left:  ir.ExprIR{Kind: ir.ExprVar, VarName: "data"},
									Right: ir.ExprIR{Kind: ir.ExprVar, VarName: "transform"},
								},
							},
						},
					},
				},
			},
		},
	}
	output := gen.Generate(service)

	if !strings.Contains(output, "transform(data)") {
		t.Error("expected pipe expression rendered as function call")
	}
}

func TestPythonWebSocket(t *testing.T) {
	gen := NewPythonGenerator("", 8000)
	service := &ir.ServiceIR{
		WebSocket: []ir.WebSocketDef{
			{
				Path: "/ws/chat",
				Events: []ir.WSEventDef{
					{
						EventType: ir.WSConnect,
						Body:      []ir.StmtIR{},
					},
					{
						EventType: ir.WSMessage,
						Body: []ir.StmtIR{
							{Kind: ir.StmtExpr, ExprStmt: &ir.ExprIR{Kind: ir.ExprVar, VarName: "handle_message"}},
						},
					},
					{
						EventType: ir.WSDisconnect,
						Body:      []ir.StmtIR{},
					},
				},
			},
		},
	}
	output := gen.Generate(service)

	if !strings.Contains(output, "class ConnectionManager") {
		t.Error("expected ConnectionManager class")
	}
	if !strings.Contains(output, `@app.websocket("/ws/chat")`) {
		t.Error("expected websocket decorator")
	}
	if !strings.Contains(output, "await manager.connect(websocket)") {
		t.Error("expected websocket connect")
	}
	if !strings.Contains(output, "WebSocketDisconnect") {
		t.Error("expected WebSocketDisconnect handling")
	}
	if !strings.Contains(output, "from fastapi import WebSocket, WebSocketDisconnect") {
		t.Error("expected WebSocket imports")
	}
}

func TestPythonWebSocketRequirements(t *testing.T) {
	gen := NewPythonGenerator("", 8000)
	service := &ir.ServiceIR{
		WebSocket: []ir.WebSocketDef{
			{Path: "/ws"},
		},
	}
	reqs := gen.GenerateRequirements(service)

	if !strings.Contains(reqs, "websockets>=12.0") {
		t.Error("expected websockets dependency")
	}
}

func TestPythonGraphQL(t *testing.T) {
	gen := NewPythonGenerator("", 8000)
	service := &ir.ServiceIR{
		GraphQL: []ir.GraphQLDef{
			{
				Operation: ir.GraphQLQuery,
				FieldName: "getUser",
				Params: []ir.FieldSchema{
					{Name: "id", Type: ir.TypeRef{Kind: ir.TypeInt}, Required: true},
				},
				ReturnType: &ir.TypeRef{Kind: ir.TypeNamed, Name: "User"},
				Body: []ir.StmtIR{
					{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{Value: ir.ExprIR{Kind: ir.ExprNull, IsNull: true}}},
				},
			},
			{
				Operation: ir.GraphQLMutation,
				FieldName: "createUser",
				Params: []ir.FieldSchema{
					{Name: "name", Type: ir.TypeRef{Kind: ir.TypeString}, Required: true},
				},
				ReturnType: &ir.TypeRef{Kind: ir.TypeNamed, Name: "User"},
				Body: []ir.StmtIR{
					{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{Value: ir.ExprIR{Kind: ir.ExprNull, IsNull: true}}},
				},
			},
		},
	}
	output := gen.Generate(service)

	if !strings.Contains(output, "import strawberry") {
		t.Error("expected strawberry import")
	}
	if !strings.Contains(output, "@strawberry.type") {
		t.Error("expected strawberry type decorator")
	}
	if !strings.Contains(output, "class Query:") {
		t.Error("expected Query class")
	}
	if !strings.Contains(output, "class Mutation:") {
		t.Error("expected Mutation class")
	}
	if !strings.Contains(output, "async def getUser") {
		t.Error("expected getUser resolver")
	}
	if !strings.Contains(output, "async def createUser") {
		t.Error("expected createUser resolver")
	}
	if !strings.Contains(output, `schema = strawberry.Schema(`) {
		t.Error("expected schema creation")
	}
	if !strings.Contains(output, `app.include_router(graphql_app, prefix="/graphql")`) {
		t.Error("expected GraphQL router mount")
	}
}

func TestPythonGraphQLRequirements(t *testing.T) {
	gen := NewPythonGenerator("", 8000)
	service := &ir.ServiceIR{
		GraphQL: []ir.GraphQLDef{
			{Operation: ir.GraphQLQuery, FieldName: "test"},
		},
	}
	reqs := gen.GenerateRequirements(service)

	if !strings.Contains(reqs, "strawberry-graphql>=0.220.0") {
		t.Error("expected strawberry-graphql dependency")
	}
}

func TestPythonGRPC(t *testing.T) {
	gen := NewPythonGenerator("", 8000)
	service := &ir.ServiceIR{
		GRPC: []ir.GRPCServiceDef{
			{
				Name: "UserService",
				Methods: []ir.GRPCMethodDef{
					{
						Name:       "GetUser",
						InputType:  ir.TypeRef{Kind: ir.TypeNamed, Name: "GetUserRequest"},
						ReturnType: ir.TypeRef{Kind: ir.TypeNamed, Name: "UserResponse"},
						StreamType: ir.GRPCUnary,
					},
					{
						Name:       "ListUsers",
						InputType:  ir.TypeRef{Kind: ir.TypeNamed, Name: "ListRequest"},
						ReturnType: ir.TypeRef{Kind: ir.TypeNamed, Name: "UserResponse"},
						StreamType: ir.GRPCServerStream,
					},
				},
				Handlers: []ir.GRPCHandlerDef{
					{
						ServiceName: "UserService",
						MethodName:  "GetUser",
						Body: []ir.StmtIR{
							{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{Value: ir.ExprIR{Kind: ir.ExprNull, IsNull: true}}},
						},
					},
				},
			},
		},
	}
	output := gen.Generate(service)

	if !strings.Contains(output, "import grpc") {
		t.Error("expected grpc import")
	}
	if !strings.Contains(output, "class UserServiceServicer:") {
		t.Error("expected UserServiceServicer class")
	}
	if !strings.Contains(output, "async def GetUser(self, request, context):") {
		t.Error("expected GetUser handler method")
	}
	if !strings.Contains(output, "async def ListUsers(self, request, context):") {
		t.Error("expected ListUsers stub method")
	}
	if !strings.Contains(output, "raise NotImplementedError") {
		t.Error("expected NotImplementedError for unimplemented methods")
	}
	if !strings.Contains(output, "stream") {
		t.Error("expected stream annotation in proto comment")
	}
}

func TestPythonGRPCRequirements(t *testing.T) {
	gen := NewPythonGenerator("", 8000)
	service := &ir.ServiceIR{
		GRPC: []ir.GRPCServiceDef{
			{Name: "TestService"},
		},
	}
	reqs := gen.GenerateRequirements(service)

	if !strings.Contains(reqs, "grpcio>=1.60.0") {
		t.Error("expected grpcio dependency")
	}
	if !strings.Contains(reqs, "protobuf>=4.25.0") {
		t.Error("expected protobuf dependency")
	}
}

func TestPythonAsyncAwait(t *testing.T) {
	gen := NewPythonGenerator("", 8000)
	service := &ir.ServiceIR{
		Routes: []ir.RouteHandler{
			{
				Method: ir.MethodGet,
				Path:   "/api/async",
				Body: []ir.StmtIR{
					{
						Kind: ir.StmtAssign,
						Assign: &ir.AssignStmt{
							Target: "result",
							Value: ir.ExprIR{
								Kind:  ir.ExprAwait,
								Await: &ir.AwaitExprIR{Expr: ir.ExprIR{Kind: ir.ExprCall, Call: &ir.CallExpr{Name: "fetch_data"}}},
							},
						},
					},
					{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{Value: ir.ExprIR{Kind: ir.ExprVar, VarName: "result"}}},
				},
			},
		},
	}
	output := gen.Generate(service)

	if !strings.Contains(output, "await fetch_data()") {
		t.Error("expected await expression")
	}
}

func TestIrTypeToPython(t *testing.T) {
	tests := []struct {
		name     string
		input    ir.TypeRef
		expected string
	}{
		{"int", ir.TypeRef{Kind: ir.TypeInt}, "int"},
		{"float", ir.TypeRef{Kind: ir.TypeFloat}, "float"},
		{"string", ir.TypeRef{Kind: ir.TypeString}, "str"},
		{"bool", ir.TypeRef{Kind: ir.TypeBool}, "bool"},
		{"any", ir.TypeRef{Kind: ir.TypeAny}, "Any"},
		{"named", ir.TypeRef{Kind: ir.TypeNamed, Name: "User"}, "User"},
		{"array", ir.TypeRef{Kind: ir.TypeArray, Inner: &ir.TypeRef{Kind: ir.TypeInt}}, "List[int]"},
		{"optional", ir.TypeRef{Kind: ir.TypeOptional, Inner: &ir.TypeRef{Kind: ir.TypeString}}, "Optional[str]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := irTypeToPython(tt.input)
			if result != tt.expected {
				t.Errorf("irTypeToPython(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
