package ir

import (
	"testing"
)

func TestTypeRefToString(t *testing.T) {
	inner := TypeRef{Kind: TypeString}
	innerInt := TypeRef{Kind: TypeInt}

	tests := []struct {
		name string
		ref  *TypeRef
		want string
	}{
		{"nil", nil, "any"},
		{"int", &TypeRef{Kind: TypeInt}, "int"},
		{"float", &TypeRef{Kind: TypeFloat}, "float"},
		{"string", &TypeRef{Kind: TypeString}, "string"},
		{"bool", &TypeRef{Kind: TypeBool}, "bool"},
		{"any", &TypeRef{Kind: TypeAny}, "any"},
		{"named", &TypeRef{Kind: TypeNamed, Name: "User"}, "User"},
		{"provider", &TypeRef{Kind: TypeProvider, Name: "Database"}, "Database"},
		{"array", &TypeRef{Kind: TypeArray, Inner: &inner}, "[string]"},
		{"optional", &TypeRef{Kind: TypeOptional, Inner: &inner}, "string?"},
		{"future", &TypeRef{Kind: TypeFuture, Inner: &inner}, "Future<string>"},
		{
			"union",
			&TypeRef{Kind: TypeUnion, Elements: []TypeRef{{Kind: TypeString}, {Kind: TypeInt}}},
			"string | int",
		},
		{
			"generic_no_args",
			&TypeRef{Kind: TypeGeneric, Name: "T"},
			"T",
		},
		{
			"generic_with_args",
			&TypeRef{Kind: TypeGeneric, Name: "Map", Elements: []TypeRef{{Kind: TypeString}, {Kind: TypeInt}}},
			"Map<string, int>",
		},
		{
			"function_no_params",
			&TypeRef{Kind: TypeFunction, Inner: &inner},
			"() -> string",
		},
		{
			"function_with_params",
			&TypeRef{Kind: TypeFunction, Elements: []TypeRef{{Kind: TypeString}, {Kind: TypeInt}}, Inner: &innerInt},
			"(string, int) -> int",
		},
		{"unknown_kind", &TypeRef{Kind: TypeKind(99)}, "any"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeRefToString(tt.ref)
			if got != tt.want {
				t.Errorf("typeRefToString(%v) = %q, want %q", tt.ref, got, tt.want)
			}
		})
	}
}

func TestAuthToAuth(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if authToAuth(nil) != nil {
			t.Error("expected nil for nil input")
		}
	})

	t.Run("jwt_with_roles", func(t *testing.T) {
		a := &AuthRequirement{AuthType: "jwt", Required: true, Roles: []string{"admin", "editor"}}
		got := authToAuth(a)
		if got == nil {
			t.Fatal("expected non-nil result")
		}
		if got.Type != "jwt" {
			t.Errorf("Type = %q, want %q", got.Type, "jwt")
		}
		if !got.Required {
			t.Error("Required should be true")
		}
		if len(got.Roles) != 2 || got.Roles[0] != "admin" {
			t.Errorf("Roles = %v, want [admin editor]", got.Roles)
		}
	})
}

func TestToCatalog_Empty(t *testing.T) {
	svc := &ServiceIR{Name: "empty"}
	cat := ToCatalog(svc)
	if cat.Service != "empty" {
		t.Errorf("Service = %q, want %q", cat.Service, "empty")
	}
	if cat.Version != "1.0" {
		t.Errorf("Version = %q, want %q", cat.Version, "1.0")
	}
	if len(cat.Operations) != 0 {
		t.Errorf("expected 0 operations, got %d", len(cat.Operations))
	}
}

func TestToCatalog_Route(t *testing.T) {
	userType := TypeRef{Kind: TypeNamed, Name: "User"}
	svc := &ServiceIR{
		Name: "user-api",
		Routes: []RouteHandler{
			{
				Method:     MethodGet,
				Path:       "/users/:id",
				PathParams: []string{"id"},
				QueryParams: []QueryParam{
					{Name: "verbose", Type: TypeRef{Kind: TypeBool}, Required: false},
				},
				InputType:  nil,
				ReturnType: &userType,
				Auth:       &AuthRequirement{AuthType: "jwt", Required: true},
				Providers: []InjectionRef{
					{Name: "db", ProviderType: "Database"},
				},
			},
		},
	}

	cat := ToCatalog(svc)
	if len(cat.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(cat.Operations))
	}

	op := cat.Operations[0]
	if op.ID != "GET /users/:id" {
		t.Errorf("ID = %q, want %q", op.ID, "GET /users/:id")
	}
	if op.Kind != "route" {
		t.Errorf("Kind = %q, want %q", op.Kind, "route")
	}
	if op.Method != "GET" {
		t.Errorf("Method = %q, want %q", op.Method, "GET")
	}
	if op.Path != "/users/:id" {
		t.Errorf("Path = %q, want %q", op.Path, "/users/:id")
	}
	if op.Returns != "User" {
		t.Errorf("Returns = %q, want %q", op.Returns, "User")
	}
	if op.Auth == nil || op.Auth.Type != "jwt" {
		t.Error("expected jwt auth")
	}
	if len(op.Providers) != 1 || op.Providers[0] != "Database" {
		t.Errorf("Providers = %v, want [Database]", op.Providers)
	}

	// path param + query param
	if len(op.Params) != 2 {
		t.Fatalf("expected 2 params (path+query), got %d", len(op.Params))
	}
	if op.Params[0].Kind != "path" || op.Params[0].Name != "id" {
		t.Errorf("first param should be path param 'id', got %+v", op.Params[0])
	}
	if op.Params[1].Kind != "query" || op.Params[1].Name != "verbose" {
		t.Errorf("second param should be query param 'verbose', got %+v", op.Params[1])
	}
}

func TestToCatalog_RouteWithBody(t *testing.T) {
	bodyType := TypeRef{Kind: TypeNamed, Name: "CreateUserInput"}
	svc := &ServiceIR{
		Routes: []RouteHandler{
			{
				Method:    MethodPost,
				Path:      "/users",
				InputType: &bodyType,
			},
		},
	}

	cat := ToCatalog(svc)
	op := cat.Operations[0]
	if len(op.Params) != 1 || op.Params[0].Kind != "body" {
		t.Errorf("expected one body param, got %v", op.Params)
	}
	if op.Params[0].Type != "CreateUserInput" {
		t.Errorf("body param type = %q, want %q", op.Params[0].Type, "CreateUserInput")
	}
}

func TestToCatalog_Function(t *testing.T) {
	retType := TypeRef{Kind: TypeString}
	svc := &ServiceIR{
		Functions: []FunctionDef{
			{
				Name: "hashPassword",
				Params: []FieldSchema{
					{Name: "input", Type: TypeRef{Kind: TypeString}, Required: true},
				},
				ReturnType: &retType,
			},
		},
	}

	cat := ToCatalog(svc)
	if len(cat.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(cat.Operations))
	}
	op := cat.Operations[0]
	if op.ID != "fn:hashPassword" {
		t.Errorf("ID = %q, want %q", op.ID, "fn:hashPassword")
	}
	if op.Kind != "function" {
		t.Errorf("Kind = %q, want %q", op.Kind, "function")
	}
	if op.Returns != "string" {
		t.Errorf("Returns = %q, want %q", op.Returns, "string")
	}
	if len(op.Params) != 1 || op.Params[0].Name != "input" {
		t.Errorf("unexpected params: %v", op.Params)
	}
}

func TestToCatalog_Command(t *testing.T) {
	svc := &ServiceIR{
		Commands: []CommandDef{
			{
				Name: "deploy",
				Params: []CommandParam{
					{Name: "env", Type: TypeRef{Kind: TypeString}, Required: true, IsFlag: false},
					{Name: "dry-run", Type: TypeRef{Kind: TypeBool}, Required: false, IsFlag: true},
				},
			},
		},
	}

	cat := ToCatalog(svc)
	if len(cat.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(cat.Operations))
	}
	op := cat.Operations[0]
	if op.ID != "cmd:deploy" {
		t.Errorf("ID = %q, want %q", op.ID, "cmd:deploy")
	}
	if op.Kind != "command" {
		t.Errorf("Kind = %q, want %q", op.Kind, "command")
	}
	if len(op.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(op.Params))
	}
	if op.Params[1].Kind != "flag" {
		t.Errorf("expected second param to be flag, got %q", op.Params[1].Kind)
	}
}

func TestToCatalog_Event(t *testing.T) {
	svc := &ServiceIR{
		Events: []EventBinding{
			{
				EventType: "user.created",
				Async:     true,
				Providers: []InjectionRef{{Name: "mail", ProviderType: "Mailer"}},
			},
		},
	}

	cat := ToCatalog(svc)
	op := cat.Operations[0]
	if op.ID != "event:user.created" {
		t.Errorf("ID = %q, want %q", op.ID, "event:user.created")
	}
	if op.Kind != "event" {
		t.Errorf("Kind = %q, want %q", op.Kind, "event")
	}
	if len(op.Providers) != 1 || op.Providers[0] != "Mailer" {
		t.Errorf("Providers = %v, want [Mailer]", op.Providers)
	}
}

func TestToCatalog_Cron(t *testing.T) {
	svc := &ServiceIR{
		CronJobs: []CronBinding{
			{Name: "cleanup", Schedule: "0 0 * * *"},
		},
	}

	cat := ToCatalog(svc)
	op := cat.Operations[0]
	if op.ID != "cron:cleanup" {
		t.Errorf("ID = %q, want %q", op.ID, "cron:cleanup")
	}
	if op.Kind != "cron" {
		t.Errorf("Kind = %q, want %q", op.Kind, "cron")
	}
}

func TestToCatalog_CronNoName(t *testing.T) {
	svc := &ServiceIR{
		CronJobs: []CronBinding{
			{Schedule: "*/5 * * * *"},
		},
	}

	cat := ToCatalog(svc)
	op := cat.Operations[0]
	if op.ID != "cron:*/5 * * * *" {
		t.Errorf("ID = %q, want %q", op.ID, "cron:*/5 * * * *")
	}
}

func TestToCatalog_Queue(t *testing.T) {
	svc := &ServiceIR{
		Queues: []QueueBinding{
			{QueueName: "emails", Providers: []InjectionRef{{Name: "smtp", ProviderType: "Mailer"}}},
		},
	}

	cat := ToCatalog(svc)
	op := cat.Operations[0]
	if op.ID != "queue:emails" {
		t.Errorf("ID = %q, want %q", op.ID, "queue:emails")
	}
	if op.Kind != "queue" {
		t.Errorf("Kind = %q, want %q", op.Kind, "queue")
	}
}

func TestToCatalog_GRPC(t *testing.T) {
	retType := TypeRef{Kind: TypeNamed, Name: "UserResponse"}
	svc := &ServiceIR{
		GRPC: []GRPCServiceDef{
			{
				Name: "UserService",
				Handlers: []GRPCHandlerDef{
					{
						MethodName: "GetUser",
						ReturnType: &retType,
						Auth:       &AuthRequirement{AuthType: "apikey", Required: true},
					},
				},
			},
		},
	}

	cat := ToCatalog(svc)
	op := cat.Operations[0]
	if op.ID != "grpc:UserService.GetUser" {
		t.Errorf("ID = %q, want %q", op.ID, "grpc:UserService.GetUser")
	}
	if op.Kind != "grpc" {
		t.Errorf("Kind = %q, want %q", op.Kind, "grpc")
	}
	if op.Returns != "UserResponse" {
		t.Errorf("Returns = %q, want %q", op.Returns, "UserResponse")
	}
	if op.Auth == nil || op.Auth.Type != "apikey" {
		t.Error("expected apikey auth")
	}
}

func TestToCatalog_GraphQL(t *testing.T) {
	retType := TypeRef{Kind: TypeArray, Inner: &TypeRef{Kind: TypeNamed, Name: "Post"}}
	svc := &ServiceIR{
		GraphQL: []GraphQLDef{
			{
				Operation:  GraphQLQuery,
				FieldName:  "posts",
				ReturnType: &retType,
			},
		},
	}

	cat := ToCatalog(svc)
	op := cat.Operations[0]
	if op.ID != "graphql:query:posts" {
		t.Errorf("ID = %q, want %q", op.ID, "graphql:query:posts")
	}
	if op.Kind != "graphql" {
		t.Errorf("Kind = %q, want %q", op.Kind, "graphql")
	}
	if op.Returns != "[Post]" {
		t.Errorf("Returns = %q, want %q", op.Returns, "[Post]")
	}
}

func TestToCatalog_TypesAndProviders(t *testing.T) {
	svc := &ServiceIR{
		Name: "svc",
		Types: []TypeSchema{
			{
				Name: "Order",
				Fields: []FieldSchema{
					{Name: "id", Type: TypeRef{Kind: TypeInt}, Required: true},
					{Name: "note", Type: TypeRef{Kind: TypeOptional, Inner: &TypeRef{Kind: TypeString}}},
				},
			},
		},
		Providers: []ProviderRef{
			{Name: "db", ProviderType: "Database", IsStandard: true},
			{Name: "cache", ProviderType: "Redis", IsStandard: true},
		},
	}

	cat := ToCatalog(svc)

	if len(cat.Types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(cat.Types))
	}
	typ := cat.Types[0]
	if typ.Name != "Order" {
		t.Errorf("type name = %q, want %q", typ.Name, "Order")
	}
	if len(typ.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(typ.Fields))
	}
	if typ.Fields[0].Type != "int" || !typ.Fields[0].Required {
		t.Errorf("unexpected first field: %+v", typ.Fields[0])
	}
	if typ.Fields[1].Type != "string?" {
		t.Errorf("second field type = %q, want %q", typ.Fields[1].Type, "string?")
	}

	if len(cat.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(cat.Providers))
	}
	if cat.Providers[0].ProviderType != "Database" {
		t.Errorf("first provider type = %q, want %q", cat.Providers[0].ProviderType, "Database")
	}
}

func TestToCatalog_OperationOrder(t *testing.T) {
	svc := &ServiceIR{
		Routes:    []RouteHandler{{Method: MethodGet, Path: "/a"}},
		Functions: []FunctionDef{{Name: "f"}},
		Commands:  []CommandDef{{Name: "c"}},
		Events:    []EventBinding{{EventType: "e"}},
		CronJobs:  []CronBinding{{Name: "j", Schedule: "* * * * *"}},
		Queues:    []QueueBinding{{QueueName: "q"}},
	}

	cat := ToCatalog(svc)
	if len(cat.Operations) != 6 {
		t.Errorf("expected 6 operations, got %d", len(cat.Operations))
	}

	kinds := make([]string, len(cat.Operations))
	for i, op := range cat.Operations {
		kinds[i] = op.Kind
	}
	expected := []string{"route", "function", "command", "event", "cron", "queue"}
	for i, want := range expected {
		if kinds[i] != want {
			t.Errorf("operation[%d].Kind = %q, want %q", i, kinds[i], want)
		}
	}
}
