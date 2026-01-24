package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractRouteParamNames(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "single parameter",
			path:     "/chat/:room",
			expected: []string{"room"},
		},
		{
			name:     "multiple parameters",
			path:     "/users/:id/:action",
			expected: []string{"id", "action"},
		},
		{
			name:     "no parameters",
			path:     "/static/path",
			expected: []string{},
		},
		{
			name:     "parameter at start",
			path:     "/:version/api",
			expected: []string{"version"},
		},
		{
			name:     "parameter at end",
			path:     "/api/items/:itemId",
			expected: []string{"itemId"},
		},
		{
			name:     "mixed static and parameters",
			path:     "/org/:orgId/project/:projectId/tasks",
			expected: []string{"orgId", "projectId"},
		},
		{
			name:     "empty path",
			path:     "",
			expected: []string{},
		},
		{
			name:     "root path",
			path:     "/",
			expected: []string{},
		},
		{
			name:     "parameter with underscore",
			path:     "/users/:user_id",
			expected: []string{"user_id"},
		},
		{
			name:     "parameter with numbers",
			path:     "/v1/:id123",
			expected: []string{"id123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRouteParamNames(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractPathParamValues(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		actualPath string
		expected   map[string]string
	}{
		{
			name:       "single parameter",
			pattern:    "/chat/:room",
			actualPath: "/chat/general",
			expected:   map[string]string{"room": "general"},
		},
		{
			name:       "multiple parameters",
			pattern:    "/chat/:room/:user",
			actualPath: "/chat/general/alice",
			expected:   map[string]string{"room": "general", "user": "alice"},
		},
		{
			name:       "parameter with numbers",
			pattern:    "/users/:id",
			actualPath: "/users/12345",
			expected:   map[string]string{"id": "12345"},
		},
		{
			name:       "no parameters",
			pattern:    "/static/path",
			actualPath: "/static/path",
			expected:   map[string]string{},
		},
		{
			name:       "URL-encoded value",
			pattern:    "/chat/:room",
			actualPath: "/chat/room%20name",
			expected:   map[string]string{"room": "room name"},
		},
		{
			name:       "URL-encoded special characters",
			pattern:    "/search/:query",
			actualPath: "/search/hello%2Fworld",
			expected:   map[string]string{"query": "hello/world"},
		},
		{
			name:       "URL-encoded unicode",
			pattern:    "/greet/:name",
			actualPath: "/greet/%E4%B8%96%E7%95%8C",
			expected:   map[string]string{"name": "世界"},
		},
		{
			name:       "mismatched path length returns empty",
			pattern:    "/chat/:room",
			actualPath: "/chat/general/extra",
			expected:   map[string]string{},
		},
		{
			name:       "trailing slashes handled",
			pattern:    "/chat/:room/",
			actualPath: "/chat/general/",
			expected:   map[string]string{"room": "general"},
		},
		{
			name:       "empty parameter value",
			pattern:    "/chat/:room",
			actualPath: "/chat/",
			expected:   map[string]string{},
		},
		{
			name:       "parameter with dash",
			pattern:    "/items/:item-id",
			actualPath: "/items/abc-123",
			expected:   map[string]string{"item-id": "abc-123"},
		},
		{
			name:       "parameter with plus sign (encoded)",
			pattern:    "/search/:q",
			actualPath: "/search/a%2Bb",
			expected:   map[string]string{"q": "a+b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPathParamValues(tt.pattern, tt.actualPath)

			if len(tt.expected) == 0 {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestConvertPatternToMuxFormat(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{
			name:     "single parameter",
			pattern:  "/chat/:room",
			expected: "/chat/{room}",
		},
		{
			name:     "multiple parameters",
			pattern:  "/users/:id/:action",
			expected: "/users/{id}/{action}",
		},
		{
			name:     "no parameters",
			pattern:  "/static/path",
			expected: "/static/path",
		},
		{
			name:     "mixed static and parameters",
			pattern:  "/org/:orgId/project/:projectId/tasks",
			expected: "/org/{orgId}/project/{projectId}/tasks",
		},
		{
			name:     "empty path",
			pattern:  "",
			expected: "",
		},
		{
			name:     "root path",
			pattern:  "/",
			expected: "/",
		},
		{
			name:     "parameter with underscore",
			pattern:  "/users/:user_id",
			expected: "/users/{user_id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertPatternToMuxFormat(tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}
