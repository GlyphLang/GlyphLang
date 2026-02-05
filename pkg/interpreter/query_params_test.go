package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractRawQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected map[string][]string
	}{
		{
			name:     "no query string",
			path:     "/api/users",
			expected: map[string][]string{},
		},
		{
			name:     "single param",
			path:     "/api/users?page=1",
			expected: map[string][]string{"page": {"1"}},
		},
		{
			name:     "multiple params",
			path:     "/api/users?page=1&limit=10",
			expected: map[string][]string{"page": {"1"}, "limit": {"10"}},
		},
		{
			name:     "multi-value param",
			path:     "/api/posts?tag=go&tag=programming",
			expected: map[string][]string{"tag": {"go", "programming"}},
		},
		{
			name:     "mixed params",
			path:     "/api/search?q=test&tag=a&tag=b&page=1",
			expected: map[string][]string{"q": {"test"}, "tag": {"a", "b"}, "page": {"1"}},
		},
		{
			name:     "empty value",
			path:     "/api/users?active=",
			expected: map[string][]string{"active": {""}},
		},
		{
			name:     "param without value",
			path:     "/api/users?debug",
			expected: map[string][]string{"debug": {""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRawQueryParams(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessQueryParams_NoDeclarations(t *testing.T) {
	raw := map[string][]string{
		"page":   {"5"},
		"limit":  {"10"},
		"active": {"true"},
		"ratio":  {"3.14"},
		"tags":   {"a", "b"},
	}

	result, err := ProcessQueryParams(raw, nil)
	require.NoError(t, err)

	// Auto-conversion for undeclared params
	assert.Equal(t, int64(5), result["page"])
	assert.Equal(t, int64(10), result["limit"])
	assert.Equal(t, true, result["active"])
	assert.Equal(t, 3.14, result["ratio"])
	assert.Equal(t, []string{"a", "b"}, result["tags"])
}

func TestProcessQueryParams_WithDeclarations(t *testing.T) {
	raw := map[string][]string{
		"page":  {"5"},
		"limit": {"abc"}, // Invalid int
	}

	declarations := []QueryParamDecl{
		{Name: "page", Type: IntType{}, Required: false},
	}

	result, err := ProcessQueryParams(raw, declarations)
	require.NoError(t, err)

	assert.Equal(t, int64(5), result["page"])
	// limit is undeclared, so it's auto-converted (stays as string since "abc" is not a number)
	assert.Equal(t, "abc", result["limit"])
}

func TestProcessQueryParams_RequiredMissing(t *testing.T) {
	raw := map[string][]string{}

	declarations := []QueryParamDecl{
		{Name: "q", Type: StringType{}, Required: true, Default: nil},
	}

	_, err := ProcessQueryParams(raw, declarations)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required query parameter missing: q")
}

func TestProcessQueryParams_TypeConversion(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		targetType  Type
		expected    interface{}
		expectError bool
	}{
		{"int valid", "42", IntType{}, int64(42), false},
		{"int invalid", "abc", IntType{}, nil, true},
		{"float valid", "3.14", FloatType{}, 3.14, false},
		{"float invalid", "xyz", FloatType{}, nil, true},
		{"bool true", "true", BoolType{}, true, false},
		{"bool false", "false", BoolType{}, false, false},
		{"bool yes", "yes", BoolType{}, true, false},
		{"bool no", "no", BoolType{}, false, false},
		{"bool 1", "1", BoolType{}, true, false},
		{"bool 0", "0", BoolType{}, false, false},
		{"bool invalid", "maybe", BoolType{}, nil, true},
		{"string", "hello", StringType{}, "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertValue(tt.value, tt.targetType)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestProcessQueryParams_ArrayType(t *testing.T) {
	raw := map[string][]string{
		"ids": {"1", "2", "3"},
	}

	declarations := []QueryParamDecl{
		{
			Name:    "ids",
			Type:    ArrayType{ElementType: IntType{}},
			IsArray: true,
		},
	}

	result, err := ProcessQueryParams(raw, declarations)
	require.NoError(t, err)

	ids, ok := result["ids"].([]interface{})
	require.True(t, ok)
	assert.Len(t, ids, 3)
	assert.Equal(t, int64(1), ids[0])
	assert.Equal(t, int64(2), ids[1])
	assert.Equal(t, int64(3), ids[2])
}

func TestAutoConvert(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"42", int64(42)},
		{"3.14", 3.14},
		{"true", true},
		{"false", false},
		{"hello", "hello"},
		{"-10", int64(-10)},
		{"0", int64(0)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := autoConvert(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
