package ir

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
)

func TestAnalyzeEmptyModule(t *testing.T) {
	a := NewAnalyzer()
	module := &ast.Module{Items: []ast.Item{}}
	ir, err := a.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ir.Types) != 0 {
		t.Errorf("expected 0 types, got %d", len(ir.Types))
	}
	if len(ir.Routes) != 0 {
		t.Errorf("expected 0 routes, got %d", len(ir.Routes))
	}
}

func TestAnalyzeTypeDef(t *testing.T) {
	a := NewAnalyzer()
	module := &ast.Module{
		Items: []ast.Item{
			&ast.TypeDef{
				Name: "User",
				Fields: []ast.Field{
					{Name: "id", TypeAnnotation: ast.IntType{}, Required: true},
					{Name: "name", TypeAnnotation: ast.StringType{}, Required: true},
					{Name: "email", TypeAnnotation: ast.StringType{}, Required: true},
					{Name: "age", TypeAnnotation: ast.OptionalType{InnerType: ast.IntType{}}},
				},
			},
		},
	}

	ir, err := a.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ir.Types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(ir.Types))
	}
	user := ir.Types[0]
	if user.Name != "User" {
		t.Errorf("expected type name 'User', got %q", user.Name)
	}
	if len(user.Fields) != 4 {
		t.Errorf("expected 4 fields, got %d", len(user.Fields))
	}
	if user.Fields[0].Type.Kind != TypeInt {
		t.Errorf("expected field 'id' to be TypeInt, got %v", user.Fields[0].Type.Kind)
	}
	if user.Fields[3].Type.Kind != TypeOptional {
		t.Errorf("expected field 'age' to be TypeOptional, got %v", user.Fields[3].Type.Kind)
	}
}

func TestAnalyzeRoute(t *testing.T) {
	a := NewAnalyzer()
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
							Name: "db.users.get",
							Args: []ast.Expr{ast.VariableExpr{Name: "id"}},
						},
					},
					ast.ReturnStatement{
						Value: ast.VariableExpr{Name: "user"},
					},
				},
			},
		},
	}

	ir, err := a.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ir.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(ir.Routes))
	}

	route := ir.Routes[0]
	if route.Method != MethodGet {
		t.Errorf("expected GET, got %v", route.Method)
	}
	if route.Path != "/api/users/:id" {
		t.Errorf("expected path '/api/users/:id', got %q", route.Path)
	}
	if len(route.PathParams) != 1 || route.PathParams[0] != "id" {
		t.Errorf("expected path param 'id', got %v", route.PathParams)
	}
	if len(route.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(route.Providers))
	}
	if route.Providers[0].ProviderType != "Database" {
		t.Errorf("expected provider type 'Database', got %q", route.Providers[0].ProviderType)
	}
	if len(route.Body) != 2 {
		t.Errorf("expected 2 statements in body, got %d", len(route.Body))
	}

	// Check that the Database provider was tracked globally
	if len(ir.Providers) != 1 {
		t.Fatalf("expected 1 global provider, got %d", len(ir.Providers))
	}
	if !ir.Providers[0].IsStandard {
		t.Error("expected Database to be a standard provider")
	}
}

func TestAnalyzeRouteWithAuth(t *testing.T) {
	a := NewAnalyzer()
	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Method: ast.Post,
				Path:   "/api/admin/users",
				Auth: &ast.AuthConfig{
					AuthType: "jwt",
					Required: true,
				},
				RateLimit: &ast.RateLimit{
					Requests: 100,
					Window:   "1m",
				},
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.ObjectExpr{
							Fields: []ast.ObjectField{
								{Key: "ok", Value: ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}}},
							},
						},
					},
				},
			},
		},
	}

	ir, err := a.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	route := ir.Routes[0]
	if route.Auth == nil {
		t.Fatal("expected auth requirement")
	}
	if route.Auth.AuthType != "jwt" {
		t.Errorf("expected auth type 'jwt', got %q", route.Auth.AuthType)
	}
	if route.RateLimit == nil {
		t.Fatal("expected rate limit")
	}
	if route.RateLimit.Requests != 100 {
		t.Errorf("expected 100 requests, got %d", route.RateLimit.Requests)
	}
}

func TestAnalyzeCronTask(t *testing.T) {
	a := NewAnalyzer()
	module := &ast.Module{
		Items: []ast.Item{
			&ast.CronTask{
				Name:     "cleanup",
				Schedule: "0 0 * * *",
				Timezone: "UTC",
				Retries:  3,
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

	ir, err := a.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ir.CronJobs) != 1 {
		t.Fatalf("expected 1 cron job, got %d", len(ir.CronJobs))
	}
	cron := ir.CronJobs[0]
	if cron.Name != "cleanup" {
		t.Errorf("expected name 'cleanup', got %q", cron.Name)
	}
	if cron.Schedule != "0 0 * * *" {
		t.Errorf("expected schedule '0 0 * * *', got %q", cron.Schedule)
	}
	if cron.Retries != 3 {
		t.Errorf("expected 3 retries, got %d", cron.Retries)
	}
}

func TestAnalyzeEventHandler(t *testing.T) {
	a := NewAnalyzer()
	module := &ast.Module{
		Items: []ast.Item{
			&ast.EventHandler{
				EventType: "user.created",
				Async:     true,
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

	ir, err := a.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ir.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(ir.Events))
	}
	if ir.Events[0].EventType != "user.created" {
		t.Errorf("expected event type 'user.created', got %q", ir.Events[0].EventType)
	}
	if !ir.Events[0].Async {
		t.Error("expected async=true")
	}
}

func TestAnalyzeMultipleProviders(t *testing.T) {
	a := NewAnalyzer()
	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Method: ast.Get,
				Path:   "/api/data",
				Injections: []ast.Injection{
					{Name: "db", Type: ast.DatabaseType{}},
					{Name: "cache", Type: ast.RedisType{}},
					{Name: "ai", Type: ast.LLMType{}},
				},
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.NullLiteral{}},
					},
				},
			},
		},
	}

	ir, err := a.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ir.Providers) != 3 {
		t.Errorf("expected 3 providers, got %d", len(ir.Providers))
	}

	providerTypes := make(map[string]bool)
	for _, p := range ir.Providers {
		providerTypes[p.ProviderType] = true
		if !p.IsStandard {
			t.Errorf("expected provider %q to be standard", p.ProviderType)
		}
	}
	for _, expected := range []string{"Database", "Redis", "LLM"} {
		if !providerTypes[expected] {
			t.Errorf("expected provider %q to be tracked", expected)
		}
	}
}

func TestAnalyzeCustomProvider(t *testing.T) {
	a := NewAnalyzer()
	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Method: ast.Post,
				Path:   "/api/upload",
				Injections: []ast.Injection{
					{Name: "images", Type: ast.NamedType{Name: "ImageProcessor"}},
				},
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.NullLiteral{}},
					},
				},
			},
		},
	}

	ir, err := a.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ir.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(ir.Providers))
	}
	if ir.Providers[0].ProviderType != "ImageProcessor" {
		t.Errorf("expected provider type 'ImageProcessor', got %q", ir.Providers[0].ProviderType)
	}
	if ir.Providers[0].IsStandard {
		t.Error("expected ImageProcessor to NOT be a standard provider")
	}
}

func TestAnalyzeProviderDef(t *testing.T) {
	a := NewAnalyzer()
	module := &ast.Module{
		Items: []ast.Item{
			&ast.ProviderDef{
				Name: "EmailService",
				Methods: []ast.ProviderMethod{
					{
						Name: "send",
						Params: []ast.Field{
							{Name: "to", TypeAnnotation: ast.StringType{}, Required: true},
							{Name: "subject", TypeAnnotation: ast.StringType{}, Required: true},
						},
						ReturnType: ast.BoolType{},
					},
					{
						Name:       "status",
						Params:     []ast.Field{{Name: "id", TypeAnnotation: ast.StringType{}, Required: true}},
						ReturnType: ast.StringType{},
					},
				},
			},
		},
	}

	ir, err := a.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ir.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(ir.Providers))
	}

	prov := ir.Providers[0]
	if prov.ProviderType != "EmailService" {
		t.Errorf("expected provider type 'EmailService', got %q", prov.ProviderType)
	}
	if prov.IsStandard {
		t.Error("expected EmailService to NOT be a standard provider")
	}
	if len(prov.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(prov.Methods))
	}
	if prov.Methods[0].Name != "send" {
		t.Errorf("expected method name 'send', got %q", prov.Methods[0].Name)
	}
	if len(prov.Methods[0].Params) != 2 {
		t.Errorf("expected 2 params on send, got %d", len(prov.Methods[0].Params))
	}
	if prov.Methods[0].ReturnType.Kind != TypeBool {
		t.Errorf("expected send return type TypeBool, got %v", prov.Methods[0].ReturnType.Kind)
	}
	if prov.Methods[1].Name != "status" {
		t.Errorf("expected method name 'status', got %q", prov.Methods[1].Name)
	}
	if prov.Methods[1].ReturnType.Kind != TypeString {
		t.Errorf("expected status return type TypeString, got %v", prov.Methods[1].ReturnType.Kind)
	}
}

func TestAnalyzeProviderDefWithInjection(t *testing.T) {
	a := NewAnalyzer()
	module := &ast.Module{
		Items: []ast.Item{
			&ast.ProviderDef{
				Name: "PaymentGateway",
				Methods: []ast.ProviderMethod{
					{
						Name:       "charge",
						Params:     []ast.Field{{Name: "amount", TypeAnnotation: ast.IntType{}, Required: true}},
						ReturnType: ast.StringType{},
					},
				},
			},
			&ast.Route{
				Method: ast.Post,
				Path:   "/api/charge",
				Injections: []ast.Injection{
					{Name: "pay", Type: ast.NamedType{Name: "PaymentGateway"}},
				},
				Body: []ast.Statement{
					ast.ReturnStatement{Value: ast.LiteralExpr{Value: ast.NullLiteral{}}},
				},
			},
		},
	}

	ir, err := a.Analyze(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Provider should appear once (not duplicated by both ProviderDef and injection)
	if len(ir.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(ir.Providers))
	}
	prov := ir.Providers[0]
	if prov.ProviderType != "PaymentGateway" {
		t.Errorf("expected 'PaymentGateway', got %q", prov.ProviderType)
	}
	// Methods should be populated from the contract
	if len(prov.Methods) != 1 {
		t.Fatalf("expected 1 method from contract, got %d", len(prov.Methods))
	}
	if prov.Methods[0].Name != "charge" {
		t.Errorf("expected method 'charge', got %q", prov.Methods[0].Name)
	}

	// Route should reference the provider
	if len(ir.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(ir.Routes))
	}
	if len(ir.Routes[0].Providers) != 1 {
		t.Fatalf("expected 1 route provider, got %d", len(ir.Routes[0].Providers))
	}
	if ir.Routes[0].Providers[0].ProviderType != "PaymentGateway" {
		t.Errorf("expected route provider 'PaymentGateway', got %q", ir.Routes[0].Providers[0].ProviderType)
	}
}

func TestConvertTypes(t *testing.T) {
	a := NewAnalyzer()

	tests := []struct {
		name     string
		astType  ast.Type
		expected TypeKind
	}{
		{"int", ast.IntType{}, TypeInt},
		{"float", ast.FloatType{}, TypeFloat},
		{"string", ast.StringType{}, TypeString},
		{"bool", ast.BoolType{}, TypeBool},
		{"database", ast.DatabaseType{}, TypeProvider},
		{"redis", ast.RedisType{}, TypeProvider},
		{"mongodb", ast.MongoDBType{}, TypeProvider},
		{"llm", ast.LLMType{}, TypeProvider},
		{"named", ast.NamedType{Name: "User"}, TypeNamed},
		{"array", ast.ArrayType{ElementType: ast.IntType{}}, TypeArray},
		{"optional", ast.OptionalType{InnerType: ast.StringType{}}, TypeOptional},
		{"nil", nil, TypeAny},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := a.convertType(tt.astType)
			if ref.Kind != tt.expected {
				t.Errorf("expected kind %v, got %v", tt.expected, ref.Kind)
			}
		})
	}
}

func TestHTTPMethodString(t *testing.T) {
	tests := []struct {
		method   HTTPMethod
		expected string
	}{
		{MethodGet, "GET"},
		{MethodPost, "POST"},
		{MethodPut, "PUT"},
		{MethodDelete, "DELETE"},
		{MethodPatch, "PATCH"},
		{MethodWebSocket, "WS"},
		{MethodSSE, "SSE"},
	}
	for _, tt := range tests {
		if got := tt.method.String(); got != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, got)
		}
	}
}
