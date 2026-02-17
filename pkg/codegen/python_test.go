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
