package docs

import (
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testModule() *interpreter.Module {
	return &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.TypeDef{
				Name: "User",
				Fields: []interpreter.Field{
					{Name: "id", TypeAnnotation: interpreter.NamedType{Name: "int"}, Required: true},
					{Name: "name", TypeAnnotation: interpreter.NamedType{Name: "str"}, Required: true},
					{Name: "email", TypeAnnotation: interpreter.NamedType{Name: "str"}, Required: false},
				},
			},
			&interpreter.Route{
				Path:       "/api/users",
				Method:     interpreter.Get,
				ReturnType: interpreter.ArrayType{ElementType: interpreter.NamedType{Name: "User"}},
			},
			&interpreter.Route{
				Path:       "/api/users/:id",
				Method:     interpreter.Get,
				ReturnType: interpreter.NamedType{Name: "User"},
			},
			&interpreter.Route{
				Path:       "/api/users",
				Method:     interpreter.Post,
				InputType:  interpreter.NamedType{Name: "CreateUser"},
				ReturnType: interpreter.NamedType{Name: "User"},
			},
			&interpreter.Route{
				Path:   "/api/users/:id",
				Method: interpreter.Delete,
			},
		},
	}
}

func TestExtractDocs(t *testing.T) {
	doc := ExtractDocs(testModule(), "Test API")

	assert.Equal(t, "Test API", doc.Title)
	require.Len(t, doc.Types, 1)
	require.Len(t, doc.Routes, 4)
}

func TestExtractDocs_TypeFields(t *testing.T) {
	doc := ExtractDocs(testModule(), "Test API")

	userType := doc.Types[0]
	assert.Equal(t, "User", userType.Name)
	require.Len(t, userType.Fields, 3)
	assert.Equal(t, "id", userType.Fields[0].Name)
	assert.Equal(t, "int", userType.Fields[0].Type)
	assert.True(t, userType.Fields[0].Required)
	assert.Equal(t, "email", userType.Fields[2].Name)
	assert.False(t, userType.Fields[2].Required)
}

func TestExtractDocs_RouteDetails(t *testing.T) {
	doc := ExtractDocs(testModule(), "Test API")

	getUsers := doc.Routes[0]
	assert.Equal(t, "GET", getUsers.Method)
	assert.Equal(t, "/api/users", getUsers.Path)
	assert.Equal(t, "User[]", getUsers.ReturnType)
	assert.Empty(t, getUsers.PathParams)

	getUserByID := doc.Routes[1]
	assert.Equal(t, "GET", getUserByID.Method)
	assert.Equal(t, "/api/users/:id", getUserByID.Path)
	require.Len(t, getUserByID.PathParams, 1)
	assert.Equal(t, "id", getUserByID.PathParams[0])

	postUser := doc.Routes[2]
	assert.Equal(t, "POST", postUser.Method)
	assert.Equal(t, "CreateUser", postUser.InputType)
}

func TestExtractDocs_SkipsWebSocket(t *testing.T) {
	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:   "/ws/chat",
				Method: interpreter.WebSocket,
			},
			&interpreter.Route{
				Path:   "/api/health",
				Method: interpreter.Get,
			},
		},
	}

	doc := ExtractDocs(module, "API")
	require.Len(t, doc.Routes, 1)
	assert.Equal(t, "/api/health", doc.Routes[0].Path)
}

func TestExtractDocs_Auth(t *testing.T) {
	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:   "/api/admin",
				Method: interpreter.Get,
				Auth:   &interpreter.AuthConfig{AuthType: "jwt"},
			},
		},
	}

	doc := ExtractDocs(module, "API")
	require.Len(t, doc.Routes, 1)
	assert.True(t, doc.Routes[0].HasAuth)
	assert.Equal(t, "jwt", doc.Routes[0].AuthType)
}

func TestExtractDocs_RateLimit(t *testing.T) {
	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:      "/api/data",
				Method:    interpreter.Get,
				RateLimit: &interpreter.RateLimit{Requests: 100, Window: "1m"},
			},
		},
	}

	doc := ExtractDocs(module, "API")
	require.Len(t, doc.Routes, 1)
	assert.Equal(t, "100 requests per 1m", doc.Routes[0].RateLimit)
}

func TestGenerateMarkdown(t *testing.T) {
	doc := ExtractDocs(testModule(), "My API")
	md := GenerateMarkdown(doc)

	assert.Contains(t, md, "# My API")
	assert.Contains(t, md, "## Endpoints")
	assert.Contains(t, md, "## Types")
	assert.Contains(t, md, "### GET `/api/users`")
	assert.Contains(t, md, "### POST `/api/users`")
	assert.Contains(t, md, "### DELETE `/api/users/:id`")
	assert.Contains(t, md, "### User")
	assert.Contains(t, md, "| `id` | int | Yes |")
	assert.Contains(t, md, "| `email` | str | No |")
	assert.Contains(t, md, "**Request Body:** `CreateUser`")
	assert.Contains(t, md, "**Response:** `User[]`")
}

func TestGenerateMarkdown_TableOfContents(t *testing.T) {
	doc := ExtractDocs(testModule(), "API")
	md := GenerateMarkdown(doc)

	assert.Contains(t, md, "## Table of Contents")
	assert.Contains(t, md, "- [Endpoints](#endpoints)")
	assert.Contains(t, md, "- [Types](#types)")
}

func TestGenerateMarkdown_EmptyModule(t *testing.T) {
	module := &interpreter.Module{Items: []interpreter.Item{}}
	doc := ExtractDocs(module, "Empty API")
	md := GenerateMarkdown(doc)

	assert.Contains(t, md, "# Empty API")
	assert.NotContains(t, md, "## Endpoints")
	assert.NotContains(t, md, "## Types")
}

func TestGenerateHTML(t *testing.T) {
	doc := ExtractDocs(testModule(), "My API")
	html := GenerateHTML(doc)

	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "<title>My API</title>")
	assert.Contains(t, html, "class=\"sidebar\"")
	assert.Contains(t, html, "class=\"content\"")
	assert.Contains(t, html, "GET")
	assert.Contains(t, html, "/api/users")
	assert.Contains(t, html, "User")
	assert.Contains(t, html, "filterDocs")
}

func TestGenerateHTML_MethodColors(t *testing.T) {
	doc := ExtractDocs(testModule(), "API")
	html := GenerateHTML(doc)

	assert.Contains(t, html, "method-get")
	assert.Contains(t, html, "method-post")
	assert.Contains(t, html, "method-delete")
}

func TestGenerateHTML_Structure(t *testing.T) {
	doc := ExtractDocs(testModule(), "API")
	h := GenerateHTML(doc)

	// Well-formed HTML
	assert.True(t, strings.HasPrefix(h, "<!DOCTYPE html>"))
	assert.Contains(t, h, "</html>")
	openTags := strings.Count(h, "<table>")
	closeTags := strings.Count(h, "</table>")
	assert.Equal(t, openTags, closeTags, "unbalanced table tags")
}

func TestTypeToString(t *testing.T) {
	tests := []struct {
		input    interpreter.Type
		expected string
	}{
		{interpreter.NamedType{Name: "int"}, "int"},
		{interpreter.ArrayType{ElementType: interpreter.NamedType{Name: "User"}}, "User[]"},
		{interpreter.OptionalType{InnerType: interpreter.NamedType{Name: "str"}}, "str?"},
		{nil, "any"},
	}

	for _, tt := range tests {
		result := typeToString(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestExtractDocs_QueryParams(t *testing.T) {
	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:   "/api/search",
				Method: interpreter.Get,
				QueryParams: []interpreter.QueryParamDecl{
					{Name: "q", Type: interpreter.NamedType{Name: "str"}, Required: true},
					{Name: "page", Type: interpreter.NamedType{Name: "int"}, Required: false},
				},
			},
		},
	}

	doc := ExtractDocs(module, "API")
	require.Len(t, doc.Routes, 1)
	require.Len(t, doc.Routes[0].QueryParams, 2)
	assert.Equal(t, "q", doc.Routes[0].QueryParams[0].Name)
	assert.True(t, doc.Routes[0].QueryParams[0].Required)
	assert.Equal(t, "page", doc.Routes[0].QueryParams[1].Name)
	assert.False(t, doc.Routes[0].QueryParams[1].Required)
}

func TestGenerateMarkdown_QueryParams(t *testing.T) {
	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:   "/api/search",
				Method: interpreter.Get,
				QueryParams: []interpreter.QueryParamDecl{
					{Name: "q", Type: interpreter.NamedType{Name: "str"}, Required: true},
				},
			},
		},
	}

	doc := ExtractDocs(module, "API")
	md := GenerateMarkdown(doc)

	assert.Contains(t, md, "**Query Parameters:**")
	assert.Contains(t, md, "| `q` | str | Yes |")
}
