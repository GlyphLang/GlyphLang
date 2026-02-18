package codegen

import (
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/ir"
)

func TestTSGenerateEmpty(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 0)
	service := &ir.ServiceIR{}
	output := gen.Generate(service)

	if !strings.Contains(output, "import express") {
		t.Error("expected express import")
	}
	if !strings.Contains(output, "const app = express()") {
		t.Error("expected express app creation")
	}
	if !strings.Contains(output, "app.listen(3000") {
		t.Error("expected app.listen in main block")
	}
}

func TestTSGenerateModel(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 0)
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

	if !strings.Contains(output, "interface User {") {
		t.Error("expected User interface")
	}
	if !strings.Contains(output, "id: number;") {
		t.Error("expected 'id: number' field")
	}
	if !strings.Contains(output, "name: string;") {
		t.Error("expected 'name: string' field")
	}
	if !strings.Contains(output, "age?: number | null;") {
		t.Error("expected optional age field")
	}
}

func TestTSGenerateRoute(t *testing.T) {
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

	gen := NewTypeScriptServerGenerator("", 3000)
	output := gen.Generate(service)

	if !strings.Contains(output, "app.get('/api/users/:id'") {
		t.Errorf("expected Express route, got:\n%s", output)
	}
	if !strings.Contains(output, "const id = req.params.id") {
		t.Error("expected path parameter extraction")
	}
	if !strings.Contains(output, "const db = getDb()") {
		t.Error("expected database provider injection")
	}
	if !strings.Contains(output, "db.users.Get(id)") {
		t.Error("expected method call")
	}
	if !strings.Contains(output, "return res.json(user)") {
		t.Error("expected return res.json statement")
	}
}

func TestTSGeneratePostRoute(t *testing.T) {
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

	gen := NewTypeScriptServerGenerator("", 3000)
	output := gen.Generate(service)

	if !strings.Contains(output, "app.post('/api/users'") {
		t.Errorf("expected POST route, got:\n%s", output)
	}
	if !strings.Contains(output, "const input: CreateUser = req.body") {
		t.Error("expected input body extraction with type")
	}
}

func TestTSGenerateCronJob(t *testing.T) {
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

	gen := NewTypeScriptServerGenerator("", 3000)
	output := gen.Generate(service)

	if !strings.Contains(output, "import cron from 'node-cron'") {
		t.Error("expected node-cron import")
	}
	if !strings.Contains(output, "cron.schedule('0 0 * * *'") {
		t.Error("expected cron schedule call")
	}
	if !strings.Contains(output, "const db = getDb()") {
		t.Error("expected provider injection in cron job")
	}
}

func TestTSGeneratePackageJSON(t *testing.T) {
	service := &ir.ServiceIR{
		Providers: []ir.ProviderRef{
			{ProviderType: "Database", IsStandard: true},
			{ProviderType: "Redis", IsStandard: true},
		},
		CronJobs: []ir.CronBinding{
			{Name: "test", Schedule: "* * * * *"},
		},
	}

	gen := NewTypeScriptServerGenerator("", 3000)
	pkg := gen.GeneratePackageJSON(service)

	if !strings.Contains(pkg, `"express"`) {
		t.Error("expected express dependency")
	}
	if !strings.Contains(pkg, `"pg"`) {
		t.Error("expected pg dependency for Database provider")
	}
	if !strings.Contains(pkg, `"redis"`) {
		t.Error("expected redis dependency")
	}
	if !strings.Contains(pkg, `"node-cron"`) {
		t.Error("expected node-cron dependency for cron jobs")
	}
	if !strings.Contains(pkg, `"typescript"`) {
		t.Error("expected typescript dev dependency")
	}
}

func TestTSCustomProvider(t *testing.T) {
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

	gen := NewTypeScriptServerGenerator("", 3000)
	output := gen.Generate(service)

	if !strings.Contains(output, "class ImageProcessorProvider") {
		t.Error("expected custom provider class")
	}
	if !strings.Contains(output, "getImageProcessor()") {
		t.Error("expected custom provider getter call")
	}
}

func TestIrTypeToTypeScript(t *testing.T) {
	tests := []struct {
		name     string
		input    ir.TypeRef
		expected string
	}{
		{"int", ir.TypeRef{Kind: ir.TypeInt}, "number"},
		{"float", ir.TypeRef{Kind: ir.TypeFloat}, "number"},
		{"string", ir.TypeRef{Kind: ir.TypeString}, "string"},
		{"bool", ir.TypeRef{Kind: ir.TypeBool}, "boolean"},
		{"any", ir.TypeRef{Kind: ir.TypeAny}, "any"},
		{"named", ir.TypeRef{Kind: ir.TypeNamed, Name: "User"}, "User"},
		{"array", ir.TypeRef{Kind: ir.TypeArray, Inner: &ir.TypeRef{Kind: ir.TypeInt}}, "number[]"},
		{"optional", ir.TypeRef{Kind: ir.TypeOptional, Inner: &ir.TypeRef{Kind: ir.TypeString}}, "string | null"},
		{"provider", ir.TypeRef{Kind: ir.TypeProvider, Name: "Database"}, "Database"},
		{"union", ir.TypeRef{Kind: ir.TypeUnion, Elements: []ir.TypeRef{
			{Kind: ir.TypeString},
			{Kind: ir.TypeInt},
		}}, "string | number"},
		{"array_nil_inner", ir.TypeRef{Kind: ir.TypeArray}, "any[]"},
		{"optional_nil_inner", ir.TypeRef{Kind: ir.TypeOptional}, "any | null"},
		{"union_empty", ir.TypeRef{Kind: ir.TypeUnion}, "any"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := irTypeToTypeScript(tt.input)
			if result != tt.expected {
				t.Errorf("irTypeToTypeScript(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGlyphPathToExpress(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/users", "/api/users"},
		{"/api/users/:id", "/api/users/:id"},
		{"/api/users/:id/posts/:postId", "/api/users/:id/posts/:postId"},
	}

	for _, tt := range tests {
		result := glyphPathToExpress(tt.input)
		if result != tt.expected {
			t.Errorf("glyphPathToExpress(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestTSBinOps(t *testing.T) {
	tests := []struct {
		op       ir.BinOp
		expected string
	}{
		{ir.OpEq, "==="},
		{ir.OpNe, "!=="},
		{ir.OpAnd, "&&"},
		{ir.OpOr, "||"},
		{ir.OpAdd, "+"},
		{ir.OpSub, "-"},
		{ir.OpMul, "*"},
		{ir.OpDiv, "/"},
		{ir.OpMod, "%"},
		{ir.OpLt, "<"},
		{ir.OpLe, "<="},
		{ir.OpGt, ">"},
		{ir.OpGe, ">="},
	}

	for _, tt := range tests {
		result := binOpToTypeScript(tt.op)
		if result != tt.expected {
			t.Errorf("binOpToTypeScript(%v) = %q, want %q", tt.op, result, tt.expected)
		}
	}
}

func TestTSUnaryOps(t *testing.T) {
	if result := unaryOpToTypeScript(ir.OpNot); result != "!" {
		t.Errorf("expected '!', got %q", result)
	}
	if result := unaryOpToTypeScript(ir.OpNeg); result != "-" {
		t.Errorf("expected '-', got %q", result)
	}
}

func TestTSGenerateTSConfig(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 0)
	config := gen.GenerateTSConfig()

	if !strings.Contains(config, `"target": "ES2020"`) {
		t.Error("expected ES2020 target")
	}
	if !strings.Contains(config, `"strict": true`) {
		t.Error("expected strict mode")
	}
	if !strings.Contains(config, `"outDir": "./dist"`) {
		t.Error("expected dist output directory")
	}
}

func TestTSStatementTypes(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)

	// Test all statement kinds via direct IR
	service := &ir.ServiceIR{
		Routes: []ir.RouteHandler{
			{
				Method: ir.MethodGet,
				Path:   "/api/test",
				Body: []ir.StmtIR{
					// Reassign
					{Kind: ir.StmtAssign, Assign: &ir.AssignStmt{
						Target: "x",
						Value:  ir.ExprIR{Kind: ir.ExprInt, IntVal: 1},
					}},
					{Kind: ir.StmtReassign, Assign: &ir.AssignStmt{
						Target: "x",
						Value:  ir.ExprIR{Kind: ir.ExprInt, IntVal: 2},
					}},
					// For loop (value only)
					{Kind: ir.StmtFor, For: &ir.ForStmt{
						ValueVar: "item",
						Iterable: ir.ExprIR{Kind: ir.ExprVar, VarName: "items"},
						Body: []ir.StmtIR{
							{Kind: ir.StmtBreak, Break: true},
						},
					}},
					// For loop (key+value)
					{Kind: ir.StmtFor, For: &ir.ForStmt{
						KeyVar:   "i",
						ValueVar: "item",
						Iterable: ir.ExprIR{Kind: ir.ExprVar, VarName: "items"},
						Body: []ir.StmtIR{
							{Kind: ir.StmtContinue, Continue: true},
						},
					}},
					// While loop
					{Kind: ir.StmtWhile, While: &ir.WhileStmt{
						Condition: ir.ExprIR{Kind: ir.ExprBool, BoolVal: true},
						Body: []ir.StmtIR{
							{Kind: ir.StmtBreak, Break: true},
						},
					}},
					// Expr statement
					{Kind: ir.StmtExpr, ExprStmt: &ir.ExprIR{
						Kind: ir.ExprCall,
						Call: &ir.CallExpr{Name: "doSomething", Args: []ir.ExprIR{}},
					}},
					// Validate
					{Kind: ir.StmtValidate, Validate: &ir.ValidateStmt{
						Call: ir.ExprIR{Kind: ir.ExprCall, Call: &ir.CallExpr{Name: "check", Args: []ir.ExprIR{}}},
					}},
					// Return
					{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{
						Value: ir.ExprIR{Kind: ir.ExprVar, VarName: "x"},
					}},
				},
			},
		},
	}

	output := gen.Generate(service)

	if !strings.Contains(output, "const x = 1") {
		t.Error("expected const assignment")
	}
	if !strings.Contains(output, "x = 2") {
		t.Error("expected reassignment")
	}
	if !strings.Contains(output, "for (const item of items)") {
		t.Error("expected for-of loop")
	}
	if !strings.Contains(output, "for (const [i, item] of items.entries())") {
		t.Error("expected for-of with key via .entries()")
	}
	if !strings.Contains(output, "while (true)") {
		t.Error("expected while loop")
	}
	if !strings.Contains(output, "break;") {
		t.Error("expected break statement")
	}
	if !strings.Contains(output, "continue;") {
		t.Error("expected continue statement")
	}
	if !strings.Contains(output, "doSomething();") {
		t.Error("expected expression statement")
	}
	if !strings.Contains(output, "// validate: check()") {
		t.Error("expected validate comment")
	}
}

func TestTSExpressionTypes(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)

	service := &ir.ServiceIR{
		Routes: []ir.RouteHandler{
			{
				Method: ir.MethodGet,
				Path:   "/api/expr",
				Body: []ir.StmtIR{
					// Int literal
					{Kind: ir.StmtAssign, Assign: &ir.AssignStmt{
						Target: "a", Value: ir.ExprIR{Kind: ir.ExprInt, IntVal: 42},
					}},
					// Float literal
					{Kind: ir.StmtAssign, Assign: &ir.AssignStmt{
						Target: "b", Value: ir.ExprIR{Kind: ir.ExprFloat, FloatVal: 3.14},
					}},
					// Bool literals
					{Kind: ir.StmtAssign, Assign: &ir.AssignStmt{
						Target: "c", Value: ir.ExprIR{Kind: ir.ExprBool, BoolVal: true},
					}},
					{Kind: ir.StmtAssign, Assign: &ir.AssignStmt{
						Target: "d", Value: ir.ExprIR{Kind: ir.ExprBool, BoolVal: false},
					}},
					// Unary expression
					{Kind: ir.StmtAssign, Assign: &ir.AssignStmt{
						Target: "e", Value: ir.ExprIR{Kind: ir.ExprUnary, UnaryOp: &ir.UnaryExpr{
							Op:    ir.OpNeg,
							Right: ir.ExprIR{Kind: ir.ExprInt, IntVal: 5},
						}},
					}},
					// Field access
					{Kind: ir.StmtAssign, Assign: &ir.AssignStmt{
						Target: "f", Value: ir.ExprIR{Kind: ir.ExprFieldAccess, FieldAccess: &ir.FieldAccessExpr{
							Object: ir.ExprIR{Kind: ir.ExprVar, VarName: "user"},
							Field:  "name",
						}},
					}},
					// Index access
					{Kind: ir.StmtAssign, Assign: &ir.AssignStmt{
						Target: "g", Value: ir.ExprIR{Kind: ir.ExprIndexAccess, IndexAccess: &ir.IndexAccessExpr{
							Object: ir.ExprIR{Kind: ir.ExprVar, VarName: "arr"},
							Index:  ir.ExprIR{Kind: ir.ExprInt, IntVal: 0},
						}},
					}},
					// Object expression
					{Kind: ir.StmtAssign, Assign: &ir.AssignStmt{
						Target: "h", Value: ir.ExprIR{Kind: ir.ExprObject, Object: &ir.ObjectExpr{
							Fields: []ir.ObjectFieldIR{
								{Key: "name", Value: ir.ExprIR{Kind: ir.ExprString, StringVal: "test"}},
								{Key: "count", Value: ir.ExprIR{Kind: ir.ExprInt, IntVal: 1}},
							},
						}},
					}},
					// Array expression
					{Kind: ir.StmtAssign, Assign: &ir.AssignStmt{
						Target: "i", Value: ir.ExprIR{Kind: ir.ExprArray, Array: &ir.ArrayExpr{
							Elements: []ir.ExprIR{
								{Kind: ir.ExprInt, IntVal: 1},
								{Kind: ir.ExprInt, IntVal: 2},
								{Kind: ir.ExprInt, IntVal: 3},
							},
						}},
					}},
					// Lambda expression
					{Kind: ir.StmtAssign, Assign: &ir.AssignStmt{
						Target: "j", Value: ir.ExprIR{Kind: ir.ExprLambda, Lambda: &ir.LambdaExpr{
							Params: []ir.FieldSchema{{Name: "x"}, {Name: "y"}},
							Body:   ir.ExprIR{Kind: ir.ExprVar, VarName: "x"},
						}},
					}},
					// Return
					{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{
						Value: ir.ExprIR{Kind: ir.ExprNull, IsNull: true},
					}},
				},
			},
		},
	}

	output := gen.Generate(service)

	if !strings.Contains(output, "const a = 42") {
		t.Error("expected int literal")
	}
	if !strings.Contains(output, "const b = 3.14") {
		t.Error("expected float literal")
	}
	if !strings.Contains(output, "const c = true") {
		t.Error("expected true literal")
	}
	if !strings.Contains(output, "const d = false") {
		t.Error("expected false literal")
	}
	if !strings.Contains(output, "const e = -5") {
		t.Error("expected unary negation")
	}
	if !strings.Contains(output, "const f = user.name") {
		t.Error("expected field access")
	}
	if !strings.Contains(output, "const g = arr[0]") {
		t.Error("expected index access")
	}
	if !strings.Contains(output, `name: "test"`) {
		t.Error("expected object key-value")
	}
	if !strings.Contains(output, "[1, 2, 3]") {
		t.Error("expected array literal")
	}
	if !strings.Contains(output, "(x, y) => x") {
		t.Error("expected lambda expression")
	}
}

func TestTSEventHandler(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
	service := &ir.ServiceIR{
		Events: []ir.EventBinding{
			{
				EventType: "user.created",
				Async:     true,
				Providers: []ir.InjectionRef{
					{Name: "db", ProviderType: "Database"},
				},
				Body: []ir.StmtIR{
					{Kind: ir.StmtAssign, Assign: &ir.AssignStmt{
						Target: "result",
						Value:  ir.ExprIR{Kind: ir.ExprString, StringVal: "handled"},
					}},
					{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{
						Value: ir.ExprIR{Kind: ir.ExprVar, VarName: "result"},
					}},
				},
			},
		},
	}

	output := gen.Generate(service)

	if !strings.Contains(output, "import { EventEmitter } from 'events'") {
		t.Error("expected EventEmitter import")
	}
	if !strings.Contains(output, "const eventBus = new EventEmitter()") {
		t.Error("expected eventBus creation")
	}
	if !strings.Contains(output, "eventBus.on('user.created'") {
		t.Error("expected event listener registration")
	}
	if !strings.Contains(output, "const db = getDb()") {
		t.Error("expected provider injection in event handler")
	}
	// Non-route return should be plain return, not res.json()
	if !strings.Contains(output, "return result") {
		t.Error("expected plain return (not res.json) in event handler")
	}
	if strings.Contains(output, "res.json(result)") {
		t.Error("event handler should not use res.json()")
	}
}

func TestTSQueueWorker(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
	service := &ir.ServiceIR{
		Queues: []ir.QueueBinding{
			{
				QueueName:   "email.send",
				Concurrency: 5,
				Providers: []ir.InjectionRef{
					{Name: "db", ProviderType: "Database"},
				},
				Body: []ir.StmtIR{
					{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{
						Value: ir.ExprIR{Kind: ir.ExprBool, BoolVal: true},
					}},
				},
			},
		},
	}

	output := gen.Generate(service)

	if !strings.Contains(output, "// --- Queue Workers ---") {
		t.Error("expected queue workers section")
	}
	if !strings.Contains(output, "async function worker_email_send(message: any)") {
		t.Error("expected queue worker function")
	}
	if !strings.Contains(output, "const db = getDb()") {
		t.Error("expected provider injection in queue worker")
	}
	// Non-route return should be plain return
	if !strings.Contains(output, "return true") {
		t.Error("expected plain return in queue worker")
	}
}

func TestTSRouteWithAuthAndRateLimit(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
	service := &ir.ServiceIR{
		Routes: []ir.RouteHandler{
			{
				Method: ir.MethodPost,
				Path:   "/api/admin",
				Auth: &ir.AuthRequirement{
					AuthType: "jwt",
					Required: true,
				},
				RateLimit: &ir.RateLimitConfig{
					Requests: 100,
					Window:   "1m",
				},
				Body: []ir.StmtIR{
					{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{
						Value: ir.ExprIR{Kind: ir.ExprNull, IsNull: true},
					}},
				},
			},
		},
	}

	output := gen.Generate(service)

	if !strings.Contains(output, "// Requires jwt authentication") {
		t.Error("expected auth comment")
	}
	if !strings.Contains(output, "// Rate limited: 100 requests per 1m") {
		t.Error("expected rate limit comment")
	}
}

func TestTSMongoDBProvider(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
	service := &ir.ServiceIR{
		Providers: []ir.ProviderRef{
			{ProviderType: "MongoDB", IsStandard: true},
		},
	}

	output := gen.Generate(service)

	if !strings.Contains(output, "import { MongoClient, Db } from 'mongodb'") {
		t.Error("expected MongoDB import")
	}

	pkg := gen.GeneratePackageJSON(service)
	if !strings.Contains(pkg, `"mongodb"`) {
		t.Error("expected mongodb dependency in package.json")
	}
}

func TestTSLLMProvider(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
	service := &ir.ServiceIR{
		Providers: []ir.ProviderRef{
			{ProviderType: "LLM", IsStandard: true},
		},
	}

	pkg := gen.GeneratePackageJSON(service)
	if !strings.Contains(pkg, `"@anthropic-ai/sdk"`) {
		t.Error("expected Anthropic SDK dependency for LLM provider")
	}
}

func TestTSRedisProviderGetter(t *testing.T) {
	result := tsProviderToGetterFunc("Redis")
	if result != "getRedis" {
		t.Errorf("expected 'getRedis', got %q", result)
	}
}

func TestTSCronJobNonRouteReturn(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
	service := &ir.ServiceIR{
		CronJobs: []ir.CronBinding{
			{
				Name:     "task",
				Schedule: "* * * * *",
				Body: []ir.StmtIR{
					{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{
						Value: ir.ExprIR{Kind: ir.ExprString, StringVal: "done"},
					}},
				},
			},
		},
	}

	output := gen.Generate(service)

	// Cron job returns should use plain return, not res.json()
	if !strings.Contains(output, `return "done"`) {
		t.Errorf("expected plain return in cron job, got:\n%s", output)
	}
	if strings.Contains(output, `res.json("done")`) {
		t.Error("cron job should not use res.json()")
	}
}

func TestTSSwitchStatement(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
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

	if !strings.Contains(output, "switch (code)") {
		t.Error("expected switch statement")
	}
	if !strings.Contains(output, "case 200:") {
		t.Error("expected case 200")
	}
	if !strings.Contains(output, "case 404:") {
		t.Error("expected case 404")
	}
	if !strings.Contains(output, "default:") {
		t.Error("expected default case")
	}
}

func TestTSPipeExpr(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
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

func TestTSWebSocket(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
	service := &ir.ServiceIR{
		WebSocket: []ir.WebSocketDef{
			{
				Path: "/ws/chat",
				Events: []ir.WSEventDef{
					{EventType: ir.WSConnect, Body: []ir.StmtIR{}},
					{
						EventType: ir.WSMessage,
						Body: []ir.StmtIR{
							{Kind: ir.StmtExpr, ExprStmt: &ir.ExprIR{Kind: ir.ExprVar, VarName: "handle_msg"}},
						},
					},
					{EventType: ir.WSDisconnect, Body: []ir.StmtIR{}},
				},
			},
		},
	}
	output := gen.Generate(service)

	if !strings.Contains(output, "WebSocketServer") {
		t.Error("expected WebSocketServer import")
	}
	if !strings.Contains(output, "createServer") {
		t.Error("expected createServer import")
	}
	if !strings.Contains(output, `path: '/ws/chat'`) {
		t.Error("expected WebSocket path")
	}
	if !strings.Contains(output, "wss.on('connection'") {
		t.Error("expected connection handler")
	}
	if !strings.Contains(output, "ws.on('message'") {
		t.Error("expected message handler")
	}
	if !strings.Contains(output, "ws.on('close'") {
		t.Error("expected close handler")
	}
	// With WebSocket, should use server.listen instead of app.listen
	if !strings.Contains(output, "server.listen(3000") {
		t.Error("expected server.listen for WebSocket support")
	}
}

func TestTSWebSocketPackageJSON(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
	service := &ir.ServiceIR{
		WebSocket: []ir.WebSocketDef{
			{Path: "/ws"},
		},
	}
	pkg := gen.GeneratePackageJSON(service)

	if !strings.Contains(pkg, `"ws"`) {
		t.Error("expected ws dependency")
	}
	if !strings.Contains(pkg, `"@types/ws"`) {
		t.Error("expected @types/ws dev dependency")
	}
}

func TestTSGraphQL(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
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

	if !strings.Contains(output, "ApolloServer") {
		t.Error("expected ApolloServer import")
	}
	if !strings.Contains(output, "type Query {") {
		t.Error("expected Query type definition")
	}
	if !strings.Contains(output, "type Mutation {") {
		t.Error("expected Mutation type definition")
	}
	if !strings.Contains(output, "getUser") {
		t.Error("expected getUser resolver")
	}
	if !strings.Contains(output, "createUser") {
		t.Error("expected createUser resolver")
	}
	if !strings.Contains(output, "apolloServer.start()") {
		t.Error("expected Apollo server start")
	}
	if !strings.Contains(output, "expressMiddleware") {
		t.Error("expected expressMiddleware mount")
	}
}

func TestTSGraphQLPackageJSON(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
	service := &ir.ServiceIR{
		GraphQL: []ir.GraphQLDef{
			{Operation: ir.GraphQLQuery, FieldName: "test"},
		},
	}
	pkg := gen.GeneratePackageJSON(service)

	if !strings.Contains(pkg, `"@apollo/server"`) {
		t.Error("expected @apollo/server dependency")
	}
	if !strings.Contains(pkg, `"graphql"`) {
		t.Error("expected graphql dependency")
	}
}

func TestTSGRPC(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
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

	if !strings.Contains(output, "@grpc/grpc-js") {
		t.Error("expected grpc-js import")
	}
	if !strings.Contains(output, "userServiceHandlers") {
		t.Error("expected UserService handlers object")
	}
	if !strings.Contains(output, "getUser:") {
		t.Error("expected getUser handler")
	}
	if !strings.Contains(output, "listUsers:") {
		t.Error("expected listUsers stub handler")
	}
	if !strings.Contains(output, "throw new Error('Not implemented')") {
		t.Error("expected Not implemented error for stubs")
	}
	if !strings.Contains(output, "stream") {
		t.Error("expected stream annotation in proto comment")
	}
}

func TestTSGRPCPackageJSON(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
	service := &ir.ServiceIR{
		GRPC: []ir.GRPCServiceDef{
			{Name: "TestService"},
		},
	}
	pkg := gen.GeneratePackageJSON(service)

	if !strings.Contains(pkg, `"@grpc/grpc-js"`) {
		t.Error("expected @grpc/grpc-js dependency")
	}
	if !strings.Contains(pkg, `"@grpc/proto-loader"`) {
		t.Error("expected @grpc/proto-loader dependency")
	}
}

func TestTSAsyncAwait(t *testing.T) {
	gen := NewTypeScriptServerGenerator("", 3000)
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
								Await: &ir.AwaitExprIR{Expr: ir.ExprIR{Kind: ir.ExprCall, Call: &ir.CallExpr{Name: "fetchData"}}},
							},
						},
					},
					{Kind: ir.StmtReturn, Return: &ir.ReturnStmt{Value: ir.ExprIR{Kind: ir.ExprVar, VarName: "result"}}},
				},
			},
		},
	}
	output := gen.Generate(service)

	if !strings.Contains(output, "await fetchData()") {
		t.Error("expected await expression")
	}
}

func TestTSGraphQLSchemaTypes(t *testing.T) {
	// Verify GraphQL schema type mapping
	tests := []struct {
		name     string
		input    ir.TypeRef
		expected string
	}{
		{"int", ir.TypeRef{Kind: ir.TypeInt}, "Int"},
		{"float", ir.TypeRef{Kind: ir.TypeFloat}, "Float"},
		{"string", ir.TypeRef{Kind: ir.TypeString}, "String"},
		{"bool", ir.TypeRef{Kind: ir.TypeBool}, "Boolean"},
		{"named", ir.TypeRef{Kind: ir.TypeNamed, Name: "User"}, "User"},
		{"array", ir.TypeRef{Kind: ir.TypeArray, Inner: &ir.TypeRef{Kind: ir.TypeString}}, "[String]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := irTypeToGraphQLSchema(tt.input)
			if result != tt.expected {
				t.Errorf("irTypeToGraphQLSchema(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
