package codegen

import (
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeScriptGenerator_Interface(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.TypeDef{
				Name: "User",
				Fields: []interpreter.Field{
					{Name: "id", TypeAnnotation: interpreter.NamedType{Name: "int"}, Required: true},
					{Name: "name", TypeAnnotation: interpreter.NamedType{Name: "str"}, Required: true},
					{Name: "email", TypeAnnotation: interpreter.NamedType{Name: "str"}, Required: false},
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

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:       "/api/users/:id",
				Method:     interpreter.Get,
				ReturnType: interpreter.NamedType{Name: "User"},
				Body:       []interpreter.Statement{},
			},
		},
	}

	code := gen.Generate(module)
	assert.Contains(t, code, "async getApiUsers(id: string): Promise<User>")
	assert.Contains(t, code, "${id}")
}

func TestTypeScriptGenerator_PostRoute(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:       "/api/users",
				Method:     interpreter.Post,
				ReturnType: interpreter.NamedType{Name: "User"},
				InputType:  interpreter.NamedType{Name: "CreateUserInput"},
				Body:       []interpreter.Statement{},
			},
		},
	}

	code := gen.Generate(module)
	assert.Contains(t, code, "async createApiUsers(body: CreateUserInput): Promise<User>")
	assert.Contains(t, code, `"POST"`)
}

func TestTypeScriptGenerator_DeleteRoute(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:   "/api/users/:id",
				Method: interpreter.Delete,
				Body:   []interpreter.Statement{},
			},
		},
	}

	code := gen.Generate(module)
	assert.Contains(t, code, "async deleteApiUsers(id: string): Promise<unknown>")
}

func TestTypeScriptGenerator_ClientClass(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &interpreter.Module{
		Items: []interpreter.Item{},
	}

	code := gen.Generate(module)
	assert.Contains(t, code, "export class ApiClient {")
	assert.Contains(t, code, "private baseUrl: string")
	assert.Contains(t, code, `constructor(baseUrl: string = "http://localhost:3000"`)
	assert.Contains(t, code, "private async request<T>")
}

func TestTypeScriptGenerator_FullModule(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.TypeDef{
				Name: "Todo",
				Fields: []interpreter.Field{
					{Name: "id", TypeAnnotation: interpreter.NamedType{Name: "int"}, Required: true},
					{Name: "title", TypeAnnotation: interpreter.NamedType{Name: "str"}, Required: true},
					{Name: "done", TypeAnnotation: interpreter.NamedType{Name: "bool"}, Required: true},
				},
			},
			&interpreter.Route{
				Path:       "/api/todos",
				Method:     interpreter.Get,
				ReturnType: interpreter.ArrayType{ElementType: interpreter.NamedType{Name: "Todo"}},
				Body:       []interpreter.Statement{},
			},
			&interpreter.Route{
				Path:       "/api/todos/:id",
				Method:     interpreter.Get,
				ReturnType: interpreter.NamedType{Name: "Todo"},
				Body:       []interpreter.Statement{},
			},
			&interpreter.Route{
				Path:       "/api/todos",
				Method:     interpreter.Post,
				ReturnType: interpreter.NamedType{Name: "Todo"},
				Body:       []interpreter.Statement{},
			},
			&interpreter.Route{
				Path:   "/api/todos/:id",
				Method: interpreter.Delete,
				Body:   []interpreter.Statement{},
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
		input    interpreter.Type
		expected string
	}{
		{interpreter.NamedType{Name: "int"}, "number"},
		{interpreter.NamedType{Name: "float"}, "number"},
		{interpreter.NamedType{Name: "str"}, "string"},
		{interpreter.NamedType{Name: "string"}, "string"},
		{interpreter.NamedType{Name: "bool"}, "boolean"},
		{interpreter.NamedType{Name: "any"}, "unknown"},
		{interpreter.NamedType{Name: "User"}, "User"},
		{interpreter.ArrayType{ElementType: interpreter.NamedType{Name: "int"}}, "number[]"},
		{interpreter.OptionalType{InnerType: interpreter.NamedType{Name: "str"}}, "string | null"},
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
		method interpreter.HttpMethod
		path   string
		expect string
	}{
		{interpreter.Get, "/api/users", "getApiUsers"},
		{interpreter.Post, "/api/users", "createApiUsers"},
		{interpreter.Put, "/api/users/:id", "updateApiUsers"},
		{interpreter.Delete, "/api/users/:id", "deleteApiUsers"},
		{interpreter.Patch, "/api/users/:id", "patchApiUsers"},
	}

	for _, tt := range tests {
		route := &interpreter.Route{Method: tt.method, Path: tt.path}
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

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:   "/ws/chat",
				Method: interpreter.WebSocket,
				Body:   []interpreter.Statement{},
			},
			&interpreter.Route{
				Path:   "/api/health",
				Method: interpreter.Get,
				Body:   []interpreter.Statement{},
			},
		},
	}

	code := gen.Generate(module)
	assert.NotContains(t, code, "WsChat")
	assert.Contains(t, code, "getApiHealth")
}

func TestTypeScriptGenerator_OutputIsValid(t *testing.T) {
	gen := NewTypeScriptGenerator("http://localhost:3000")

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.TypeDef{
				Name: "Item",
				Fields: []interpreter.Field{
					{Name: "id", TypeAnnotation: interpreter.NamedType{Name: "int"}, Required: true},
				},
			},
			&interpreter.Route{
				Path:   "/items",
				Method: interpreter.Get,
				Body:   []interpreter.Statement{},
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
