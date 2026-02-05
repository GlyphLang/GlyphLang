package codegen

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeScriptGenerator_Interface(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &ast.Module{
		Items: []ast.Item{
			&ast.TypeDef{
				Name: "User",
				Fields: []ast.Field{
					{Name: "id", TypeAnnotation: ast.NamedType{Name: "int"}, Required: true},
					{Name: "name", TypeAnnotation: ast.NamedType{Name: "str"}, Required: true},
					{Name: "email", TypeAnnotation: ast.NamedType{Name: "str"}, Required: false},
				},
			},
		},
	}

	code := gen.Generate(module)
	assert.Contains(t, code, "export interface User {")
	assert.Contains(t, code, "  id: number;")
	assert.Contains(t, code, "  name: string;")
	assert.Contains(t, code, "  email?: string;")
}

func TestTypeScriptGenerator_GetRoute(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Path:       "/api/users/:id",
				Method:     ast.Get,
				ReturnType: ast.NamedType{Name: "User"},
				Body:       []ast.Statement{},
			},
		},
	}

	code := gen.Generate(module)
	assert.Contains(t, code, "async getApiUsers(id: string): Promise<User>")
	assert.Contains(t, code, "${id}")
}

func TestTypeScriptGenerator_PostRoute(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Path:       "/api/users",
				Method:     ast.Post,
				ReturnType: ast.NamedType{Name: "User"},
				InputType:  ast.NamedType{Name: "CreateUserInput"},
				Body:       []ast.Statement{},
			},
		},
	}

	code := gen.Generate(module)
	assert.Contains(t, code, "async createApiUsers(body: CreateUserInput): Promise<User>")
	assert.Contains(t, code, `"POST"`)
}

func TestTypeScriptGenerator_DeleteRoute(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Path:   "/api/users/:id",
				Method: ast.Delete,
				Body:   []ast.Statement{},
			},
		},
	}

	code := gen.Generate(module)
	assert.Contains(t, code, "async deleteApiUsers(id: string): Promise<unknown>")
}

func TestTypeScriptGenerator_ClientClass(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &ast.Module{
		Items: []ast.Item{},
	}

	code := gen.Generate(module)
	assert.Contains(t, code, "export class ApiClient {")
	assert.Contains(t, code, "private baseUrl: string")
	assert.Contains(t, code, `constructor(baseUrl: string = "http://localhost:3000"`)
	assert.Contains(t, code, "private async request<T>")
}

func TestTypeScriptGenerator_FullModule(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &ast.Module{
		Items: []ast.Item{
			&ast.TypeDef{
				Name: "Todo",
				Fields: []ast.Field{
					{Name: "id", TypeAnnotation: ast.NamedType{Name: "int"}, Required: true},
					{Name: "title", TypeAnnotation: ast.NamedType{Name: "str"}, Required: true},
					{Name: "done", TypeAnnotation: ast.NamedType{Name: "bool"}, Required: true},
				},
			},
			&ast.Route{
				Path:       "/api/todos",
				Method:     ast.Get,
				ReturnType: ast.ArrayType{ElementType: ast.NamedType{Name: "Todo"}},
				Body:       []ast.Statement{},
			},
			&ast.Route{
				Path:       "/api/todos/:id",
				Method:     ast.Get,
				ReturnType: ast.NamedType{Name: "Todo"},
				Body:       []ast.Statement{},
			},
			&ast.Route{
				Path:       "/api/todos",
				Method:     ast.Post,
				ReturnType: ast.NamedType{Name: "Todo"},
				Body:       []ast.Statement{},
			},
			&ast.Route{
				Path:   "/api/todos/:id",
				Method: ast.Delete,
				Body:   []ast.Statement{},
			},
		},
	}

	code := gen.Generate(module)

	// Should have the interface
	assert.Contains(t, code, "export interface Todo {")

	// Should have all methods
	assert.Contains(t, code, "getApiTodos")
	assert.Contains(t, code, "createApiTodos")
	assert.Contains(t, code, "deleteApiTodos")

	// Should have auto-generated header
	assert.Contains(t, code, "Auto-generated")
}

func TestGlyphTypeToTS(t *testing.T) {
	tests := []struct {
		input    ast.Type
		expected string
	}{
		{ast.NamedType{Name: "int"}, "number"},
		{ast.NamedType{Name: "float"}, "number"},
		{ast.NamedType{Name: "str"}, "string"},
		{ast.NamedType{Name: "string"}, "string"},
		{ast.NamedType{Name: "bool"}, "boolean"},
		{ast.NamedType{Name: "any"}, "unknown"},
		{ast.NamedType{Name: "User"}, "User"},
		{ast.ArrayType{ElementType: ast.NamedType{Name: "int"}}, "number[]"},
		{ast.OptionalType{InnerType: ast.NamedType{Name: "str"}}, "string | null"},
		{nil, "unknown"},
	}

	for _, tt := range tests {
		result := glyphTypeToTS(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestExtractPathParams(t *testing.T) {
	params := extractPathParams("/api/users/:id/posts/:postId")
	require.Len(t, params, 2)
	assert.Equal(t, "id", params[0])
	assert.Equal(t, "postId", params[1])
}

func TestGlyphPathToTemplate(t *testing.T) {
	result := glyphPathToTemplate("/api/users/:id/posts/:postId")
	assert.Equal(t, "/api/users/${id}/posts/${postId}", result)
}

func TestRouteToMethodName(t *testing.T) {
	tests := []struct {
		method ast.HttpMethod
		path   string
		expect string
	}{
		{ast.Get, "/api/users", "getApiUsers"},
		{ast.Post, "/api/users", "createApiUsers"},
		{ast.Put, "/api/users/:id", "updateApiUsers"},
		{ast.Delete, "/api/users/:id", "deleteApiUsers"},
		{ast.Patch, "/api/users/:id", "patchApiUsers"},
	}

	for _, tt := range tests {
		route := &ast.Route{Method: tt.method, Path: tt.path}
		assert.Equal(t, tt.expect, routeToMethodName(route))
	}
}

func TestCapitalize(t *testing.T) {
	assert.Equal(t, "Hello", capitalize("hello"))
	assert.Equal(t, "A", capitalize("a"))
	assert.Equal(t, "", capitalize(""))
}

func TestTypeScriptGenerator_SkipsWebSocket(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Path:   "/ws/chat",
				Method: ast.WebSocket,
				Body:   []ast.Statement{},
			},
			&ast.Route{
				Path:   "/api/health",
				Method: ast.Get,
				Body:   []ast.Statement{},
			},
		},
	}

	code := gen.Generate(module)
	assert.NotContains(t, code, "WsChat")
	assert.Contains(t, code, "getApiHealth")
}

func TestTypeScriptGenerator_OutputIsValid(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &ast.Module{
		Items: []ast.Item{
			&ast.TypeDef{
				Name: "Item",
				Fields: []ast.Field{
					{Name: "id", TypeAnnotation: ast.NamedType{Name: "int"}, Required: true},
				},
			},
			&ast.Route{
				Path:   "/items",
				Method: ast.Get,
				Body:   []ast.Statement{},
			},
		},
	}

	code := gen.Generate(module)

	// Basic structural validation
	assert.True(t, strings.Contains(code, "export interface"))
	assert.True(t, strings.Contains(code, "export class ApiClient"))
	assert.True(t, strings.Contains(code, "constructor"))
	assert.True(t, strings.Contains(code, "async"))

	// Check balanced braces
	openBraces := strings.Count(code, "{")
	closeBraces := strings.Count(code, "}")
	assert.Equal(t, openBraces, closeBraces, "unbalanced braces in generated code")
}
