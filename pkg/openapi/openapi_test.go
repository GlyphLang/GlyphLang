package openapi

import (
	"encoding/json"
	"github.com/glyphlang/glyph/pkg/ast"
	"testing"
)

func TestGenerator_EmptyModule(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{Items: []ast.Item{}}
	spec := gen.Generate(module)

	if spec.OpenAPI != "3.0.3" {
		t.Errorf("expected OpenAPI 3.0.3, got %s", spec.OpenAPI)
	}
	if spec.Info.Title != "Test API" {
		t.Errorf("expected title Test API, got %s", spec.Info.Title)
	}
	if spec.Info.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", spec.Info.Version)
	}
	if len(spec.Paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(spec.Paths))
	}
}

func TestGenerator_SimpleGetRoute(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.Route{
				Path:   "/api/health",
				Method: ast.Get,
				Body:   []ast.Statement{},
			},
		},
	}

	spec := gen.Generate(module)

	path, ok := spec.Paths["/api/health"]
	if !ok {
		t.Fatal("expected /api/health path")
	}
	if path.Get == nil {
		t.Fatal("expected GET operation")
	}
	if path.Post != nil || path.Put != nil || path.Delete != nil {
		t.Error("expected only GET operation to be set")
	}
	if _, ok := path.Get.Responses["200"]; !ok {
		t.Error("expected 200 response")
	}
}

func TestGenerator_AllHTTPMethods(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.Route{Path: "/api/users", Method: ast.Get},
			ast.Route{Path: "/api/users", Method: ast.Post},
			ast.Route{Path: "/api/users/:id", Method: ast.Put},
			ast.Route{Path: "/api/users/:id", Method: ast.Delete},
			ast.Route{Path: "/api/users/:id", Method: ast.Patch},
		},
	}

	spec := gen.Generate(module)

	usersPath := spec.Paths["/api/users"]
	if usersPath == nil {
		t.Fatal("expected /api/users path")
	}
	if usersPath.Get == nil {
		t.Error("expected GET on /api/users")
	}
	if usersPath.Post == nil {
		t.Error("expected POST on /api/users")
	}

	usersIdPath := spec.Paths["/api/users/{id}"]
	if usersIdPath == nil {
		t.Fatal("expected /api/users/{id} path")
	}
	if usersIdPath.Put == nil {
		t.Error("expected PUT on /api/users/{id}")
	}
	if usersIdPath.Delete == nil {
		t.Error("expected DELETE on /api/users/{id}")
	}
	if usersIdPath.Patch == nil {
		t.Error("expected PATCH on /api/users/{id}")
	}
}

func TestGenerator_PathParameters(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.Route{
				Path:   "/api/users/:userId/posts/:postId",
				Method: ast.Get,
			},
		},
	}

	spec := gen.Generate(module)

	path := spec.Paths["/api/users/{userId}/posts/{postId}"]
	if path == nil {
		t.Fatal("expected path with OpenAPI-style parameters")
	}

	op := path.Get
	if len(op.Parameters) != 2 {
		t.Fatalf("expected 2 path parameters, got %d", len(op.Parameters))
	}

	if op.Parameters[0].Name != "userId" || op.Parameters[0].In != "path" || !op.Parameters[0].Required {
		t.Errorf("unexpected first param: %+v", op.Parameters[0])
	}
	if op.Parameters[1].Name != "postId" || op.Parameters[1].In != "path" || !op.Parameters[1].Required {
		t.Errorf("unexpected second param: %+v", op.Parameters[1])
	}
}

func TestGenerator_QueryParameters(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.Route{
				Path:   "/api/users",
				Method: ast.Get,
				QueryParams: []ast.QueryParamDecl{
					{Name: "page", Type: ast.IntType{}, Required: false},
					{Name: "q", Type: ast.StringType{}, Required: true},
					{Name: "tags", Type: ast.StringType{}, Required: false, IsArray: true},
				},
			},
		},
	}

	spec := gen.Generate(module)
	op := spec.Paths["/api/users"].Get

	if len(op.Parameters) != 3 {
		t.Fatalf("expected 3 query params, got %d", len(op.Parameters))
	}

	pageParam := op.Parameters[0]
	if pageParam.Name != "page" || pageParam.In != "query" || pageParam.Required {
		t.Errorf("unexpected page param: %+v", pageParam)
	}
	if pageParam.Schema.Type != "integer" {
		t.Errorf("expected integer type for page, got %s", pageParam.Schema.Type)
	}

	qParam := op.Parameters[1]
	if !qParam.Required {
		t.Error("expected q to be required")
	}

	tagsParam := op.Parameters[2]
	if tagsParam.Schema.Type != "array" {
		t.Errorf("expected array type for tags, got %s", tagsParam.Schema.Type)
	}
}

func TestGenerator_TypeDefinitions(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.TypeDef{
				Name: "User",
				Fields: []ast.Field{
					{Name: "id", TypeAnnotation: ast.IntType{}, Required: true},
					{Name: "name", TypeAnnotation: ast.StringType{}, Required: true},
					{Name: "email", TypeAnnotation: ast.OptionalType{InnerType: ast.StringType{}}, Required: false},
					{Name: "roles", TypeAnnotation: ast.ArrayType{ElementType: ast.StringType{}}},
					{Name: "score", TypeAnnotation: ast.FloatType{}, Required: true},
					{Name: "active", TypeAnnotation: ast.BoolType{}, Required: true},
				},
			},
		},
	}

	spec := gen.Generate(module)

	userSchema, ok := spec.Components.Schemas["User"]
	if !ok {
		t.Fatal("expected User schema in components")
	}

	if userSchema.Type != "object" {
		t.Errorf("expected object type, got %s", userSchema.Type)
	}

	// Check required fields
	expectedRequired := []string{"active", "id", "name", "score"}
	if len(userSchema.Required) != len(expectedRequired) {
		t.Fatalf("expected %d required fields, got %d: %v", len(expectedRequired), len(userSchema.Required), userSchema.Required)
	}
	for i, req := range userSchema.Required {
		if req != expectedRequired[i] {
			t.Errorf("expected required[%d] = %s, got %s", i, expectedRequired[i], req)
		}
	}

	// Check individual fields
	if userSchema.Properties["id"].Type != "integer" {
		t.Errorf("expected integer for id, got %s", userSchema.Properties["id"].Type)
	}
	if userSchema.Properties["name"].Type != "string" {
		t.Errorf("expected string for name, got %s", userSchema.Properties["name"].Type)
	}
	if !userSchema.Properties["email"].Nullable {
		t.Error("expected email to be nullable")
	}
	if userSchema.Properties["roles"].Type != "array" {
		t.Errorf("expected array for roles, got %s", userSchema.Properties["roles"].Type)
	}
	if userSchema.Properties["score"].Type != "number" {
		t.Errorf("expected number for score, got %s", userSchema.Properties["score"].Type)
	}
	if userSchema.Properties["active"].Type != "boolean" {
		t.Errorf("expected boolean for active, got %s", userSchema.Properties["active"].Type)
	}
}

func TestGenerator_ReturnTypeRef(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.TypeDef{
				Name: "User",
				Fields: []ast.Field{
					{Name: "id", TypeAnnotation: ast.IntType{}, Required: true},
				},
			},
			ast.Route{
				Path:       "/api/users/:id",
				Method:     ast.Get,
				ReturnType: ast.NamedType{Name: "User"},
			},
		},
	}

	spec := gen.Generate(module)
	op := spec.Paths["/api/users/{id}"].Get
	resp := op.Responses["200"]
	schema := resp.Content["application/json"].Schema

	if schema.Ref != "#/components/schemas/User" {
		t.Errorf("expected $ref to User schema, got %s", schema.Ref)
	}
}

func TestGenerator_UnionReturnType(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.Route{
				Path:   "/api/users/:id",
				Method: ast.Get,
				ReturnType: ast.UnionType{
					Types: []ast.Type{
						ast.NamedType{Name: "User"},
						ast.NamedType{Name: "NotFound"},
					},
				},
			},
		},
	}

	spec := gen.Generate(module)
	op := spec.Paths["/api/users/{id}"].Get

	if _, ok := op.Responses["200"]; !ok {
		t.Error("expected 200 response")
	}
	if _, ok := op.Responses["404"]; !ok {
		t.Error("expected 404 response for NotFound union member")
	}
}

func TestGenerator_AuthJWT(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.Route{
				Path:   "/api/me",
				Method: ast.Get,
				Auth:   &ast.AuthConfig{AuthType: "jwt", Required: true},
			},
		},
	}

	spec := gen.Generate(module)

	// Check security scheme
	scheme, ok := spec.Components.SecuritySchemes["bearerAuth"]
	if !ok {
		t.Fatal("expected bearerAuth security scheme")
	}
	if scheme.Type != "http" || scheme.Scheme != "bearer" || scheme.BearerFormat != "JWT" {
		t.Errorf("unexpected JWT scheme: %+v", scheme)
	}

	// Check operation security
	op := spec.Paths["/api/me"].Get
	if len(op.Security) == 0 {
		t.Fatal("expected security on operation")
	}
	if _, ok := op.Security[0]["bearerAuth"]; !ok {
		t.Error("expected bearerAuth in operation security")
	}
}

func TestGenerator_RequestBody(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.TypeDef{
				Name: "CreateUserRequest",
				Fields: []ast.Field{
					{Name: "name", TypeAnnotation: ast.StringType{}, Required: true},
					{Name: "email", TypeAnnotation: ast.StringType{}, Required: true},
				},
			},
			ast.Route{
				Path:      "/api/users",
				Method:    ast.Post,
				InputType: ast.NamedType{Name: "CreateUserRequest"},
			},
		},
	}

	spec := gen.Generate(module)
	op := spec.Paths["/api/users"].Post

	if op.RequestBody == nil {
		t.Fatal("expected request body for POST")
	}
	if !op.RequestBody.Required {
		t.Error("expected required request body")
	}

	schema := op.RequestBody.Content["application/json"].Schema
	if schema.Ref != "#/components/schemas/CreateUserRequest" {
		t.Errorf("expected ref to CreateUserRequest, got %s", schema.Ref)
	}
}

func TestGenerator_PostWithoutInputType(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.Route{
				Path:   "/api/action",
				Method: ast.Post,
			},
		},
	}

	spec := gen.Generate(module)
	op := spec.Paths["/api/action"].Post

	if op.RequestBody == nil {
		t.Fatal("expected default request body for POST")
	}
	schema := op.RequestBody.Content["application/json"].Schema
	if schema.Type != "object" {
		t.Errorf("expected default object schema, got %s", schema.Type)
	}
}

func TestGenerator_NamedTypeTimestamp(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.TypeDef{
				Name: "Event",
				Fields: []ast.Field{
					{Name: "created_at", TypeAnnotation: ast.NamedType{Name: "timestamp"}, Required: true},
				},
			},
		},
	}

	spec := gen.Generate(module)
	schema := spec.Components.Schemas["Event"]
	prop := schema.Properties["created_at"]

	if prop.Type != "string" || prop.Format != "date-time" {
		t.Errorf("expected string/date-time for timestamp, got %s/%s", prop.Type, prop.Format)
	}
}

func TestGenerator_GenericListType(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.TypeDef{
				Name: "UserList",
				Fields: []ast.Field{
					{
						Name: "users",
						TypeAnnotation: ast.GenericType{
							BaseType: ast.NamedType{Name: "List"},
							TypeArgs: []ast.Type{ast.NamedType{Name: "User"}},
						},
					},
				},
			},
		},
	}

	spec := gen.Generate(module)
	schema := spec.Components.Schemas["UserList"]
	prop := schema.Properties["users"]

	if prop.Type != "array" {
		t.Errorf("expected array for List[User], got %s", prop.Type)
	}
	if prop.Items == nil || prop.Items.Ref != "#/components/schemas/User" {
		t.Error("expected items to ref User schema")
	}
}

func TestGenerator_WebSocketSkipped(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.Route{
				Path:   "/ws/chat",
				Method: ast.WebSocket,
			},
			ast.Route{
				Path:   "/api/health",
				Method: ast.Get,
			},
		},
	}

	spec := gen.Generate(module)

	if _, ok := spec.Paths["/ws/chat"]; ok {
		t.Error("WebSocket routes should be skipped")
	}
	if _, ok := spec.Paths["/api/health"]; !ok {
		t.Error("HTTP routes should be included")
	}
}

func TestGenerator_OperationID(t *testing.T) {
	tests := []struct {
		method   ast.HttpMethod
		path     string
		expected string
	}{
		{ast.Get, "/api/users", "getApi_users"},
		{ast.Post, "/api/users", "postApi_users"},
		{ast.Get, "/api/users/:id", "getApi_users_byId"},
		{ast.Delete, "/api/users/:id/posts/:postId", "deleteApi_users_byId_posts_byPostId"},
	}

	for _, tt := range tests {
		route := &ast.Route{Path: tt.path, Method: tt.method}
		got := generateOperationID(route)
		if got != tt.expected {
			t.Errorf("operationID for %s %s: expected %s, got %s", tt.method, tt.path, tt.expected, got)
		}
	}
}

func TestGenerator_Tags(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.Route{Path: "/api/users", Method: ast.Get},
			ast.Route{Path: "/api/todos", Method: ast.Get},
		},
	}

	spec := gen.Generate(module)

	usersOp := spec.Paths["/api/users"].Get
	if len(usersOp.Tags) == 0 || usersOp.Tags[0] != "users" {
		t.Errorf("expected users tag, got %v", usersOp.Tags)
	}

	todosOp := spec.Paths["/api/todos"].Get
	if len(todosOp.Tags) == 0 || todosOp.Tags[0] != "todos" {
		t.Errorf("expected todos tag, got %v", todosOp.Tags)
	}
}

func TestGlyphPathToOpenAPI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/users", "/api/users"},
		{"/api/users/:id", "/api/users/{id}"},
		{"/api/users/:userId/posts/:postId", "/api/users/{userId}/posts/{postId}"},
		{"/", "/"},
	}

	for _, tt := range tests {
		got := glyphPathToOpenAPI(tt.input)
		if got != tt.expected {
			t.Errorf("glyphPathToOpenAPI(%s): expected %s, got %s", tt.input, tt.expected, got)
		}
	}
}

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{"/api/users", nil},
		{"/api/users/:id", []string{"id"}},
		{"/api/users/:userId/posts/:postId", []string{"userId", "postId"}},
	}

	for _, tt := range tests {
		got := extractPathParams(tt.path)
		if len(got) != len(tt.expected) {
			t.Errorf("extractPathParams(%s): expected %v, got %v", tt.path, tt.expected, got)
			continue
		}
		for i, p := range got {
			if p != tt.expected[i] {
				t.Errorf("extractPathParams(%s)[%d]: expected %s, got %s", tt.path, i, tt.expected[i], p)
			}
		}
	}
}

func TestSpec_ToJSON(t *testing.T) {
	spec := &Spec{
		OpenAPI: "3.0.3",
		Info:    Info{Title: "Test", Version: "1.0.0"},
		Paths:   map[string]*PathItem{},
	}

	data, err := spec.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed["openapi"] != "3.0.3" {
		t.Errorf("expected openapi 3.0.3 in JSON output")
	}
}

func TestSpec_ToYAML(t *testing.T) {
	spec := &Spec{
		OpenAPI: "3.0.3",
		Info:    Info{Title: "Test", Version: "1.0.0"},
		Paths:   map[string]*PathItem{},
	}

	data, err := spec.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}

	yamlStr := string(data)
	if !containsSubstring(yamlStr, "openapi: 3.0.3") {
		t.Error("expected openapi: 3.0.3 in YAML output")
	}
	if !containsSubstring(yamlStr, "title: Test") {
		t.Error("expected title: Test in YAML output")
	}
}

func TestFormatSpec(t *testing.T) {
	spec := &Spec{
		OpenAPI: "3.0.3",
		Info:    Info{Title: "Test", Version: "1.0.0"},
		Paths:   map[string]*PathItem{},
	}

	_, err := FormatSpec(spec, "json")
	if err != nil {
		t.Errorf("FormatSpec json failed: %v", err)
	}

	_, err = FormatSpec(spec, "yaml")
	if err != nil {
		t.Errorf("FormatSpec yaml failed: %v", err)
	}

	_, err = FormatSpec(spec, "xml")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestGenerateFromModule(t *testing.T) {
	module := &ast.Module{
		Items: []ast.Item{
			ast.Route{
				Path:   "/api/test",
				Method: ast.Get,
			},
		},
	}

	spec := GenerateFromModule(module, "My API", "2.0.0")
	if spec.Info.Title != "My API" {
		t.Errorf("expected My API, got %s", spec.Info.Title)
	}
	if len(spec.Paths) != 1 {
		t.Errorf("expected 1 path, got %d", len(spec.Paths))
	}
}

func TestGenerator_MultipleAuthTypes(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.Route{
				Path:   "/api/jwt-route",
				Method: ast.Get,
				Auth:   &ast.AuthConfig{AuthType: "jwt"},
			},
			ast.Route{
				Path:   "/api/basic-route",
				Method: ast.Get,
				Auth:   &ast.AuthConfig{AuthType: "basic"},
			},
		},
	}

	spec := gen.Generate(module)

	if _, ok := spec.Components.SecuritySchemes["bearerAuth"]; !ok {
		t.Error("expected bearerAuth scheme")
	}
	if _, ok := spec.Components.SecuritySchemes["basicAuth"]; !ok {
		t.Error("expected basicAuth scheme")
	}
}

func TestGenerator_NoAuthNoSecuritySchemes(t *testing.T) {
	gen := NewGenerator("Test API", "1.0.0")
	module := &ast.Module{
		Items: []ast.Item{
			ast.Route{
				Path:   "/api/public",
				Method: ast.Get,
			},
		},
	}

	spec := gen.Generate(module)

	if spec.Components != nil && len(spec.Components.SecuritySchemes) > 0 {
		t.Error("expected no security schemes for routes without auth")
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
