package interpreter

import (
	"testing"

	. "github.com/glyphlang/glyph/pkg/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectBuiltin_DefaultStatusCode(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	args := []Expr{
		LiteralExpr{Value: StringLiteral{Value: "/new"}},
	}
	result, err := builtinRedirect(interp, args, env)

	require.NoError(t, err)
	redir, ok := result.(*RedirectResponse)
	require.True(t, ok, "expected *RedirectResponse, got %T", result)
	assert.Equal(t, "/new", redir.URL)
	assert.Equal(t, 302, redir.StatusCode)
}

func TestRedirectBuiltin_CustomStatusCode(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	args := []Expr{
		LiteralExpr{Value: StringLiteral{Value: "/new"}},
		LiteralExpr{Value: IntLiteral{Value: 301}},
	}
	result, err := builtinRedirect(interp, args, env)

	require.NoError(t, err)
	redir, ok := result.(*RedirectResponse)
	require.True(t, ok, "expected *RedirectResponse, got %T", result)
	assert.Equal(t, "/new", redir.URL)
	assert.Equal(t, 301, redir.StatusCode)
}

func TestRedirectBuiltin_AllValidCodes(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	validCodes := []int64{301, 302, 307, 308}
	for _, code := range validCodes {
		args := []Expr{
			LiteralExpr{Value: StringLiteral{Value: "/target"}},
			LiteralExpr{Value: IntLiteral{Value: code}},
		}
		result, err := builtinRedirect(interp, args, env)

		require.NoError(t, err, "code %d should be valid", code)
		redir, ok := result.(*RedirectResponse)
		require.True(t, ok, "expected *RedirectResponse for code %d", code)
		assert.Equal(t, int(code), redir.StatusCode)
	}
}

func TestRedirectBuiltin_InvalidStatusCode(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	invalidCodes := []int64{200, 404, 500, 303, 100}
	for _, code := range invalidCodes {
		args := []Expr{
			LiteralExpr{Value: StringLiteral{Value: "/new"}},
			LiteralExpr{Value: IntLiteral{Value: code}},
		}
		_, err := builtinRedirect(interp, args, env)
		assert.Error(t, err, "code %d should be invalid", code)
		assert.Contains(t, err.Error(), "invalid redirect status code")
	}
}

func TestRedirectBuiltin_NoArgs(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	args := []Expr{}
	_, err := builtinRedirect(interp, args, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redirect() requires 1-2 arguments")
}

func TestRedirectBuiltin_TooManyArgs(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	args := []Expr{
		LiteralExpr{Value: StringLiteral{Value: "/a"}},
		LiteralExpr{Value: IntLiteral{Value: 302}},
		LiteralExpr{Value: StringLiteral{Value: "extra"}},
	}
	_, err := builtinRedirect(interp, args, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redirect() requires 1-2 arguments")
}

func TestRedirectBuiltin_NonStringURL(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	args := []Expr{
		LiteralExpr{Value: IntLiteral{Value: 42}},
	}
	_, err := builtinRedirect(interp, args, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redirect() first argument must be a string URL")
}

func TestRedirectBuiltin_NonIntStatusCode(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	args := []Expr{
		LiteralExpr{Value: StringLiteral{Value: "/new"}},
		LiteralExpr{Value: StringLiteral{Value: "301"}},
	}
	_, err := builtinRedirect(interp, args, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redirect() second argument must be an integer status code")
}

func TestRedirectBuiltin_AbsoluteURL(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	args := []Expr{
		LiteralExpr{Value: StringLiteral{Value: "https://example.com/path"}},
	}
	result, err := builtinRedirect(interp, args, env)

	require.NoError(t, err)
	redir, ok := result.(*RedirectResponse)
	require.True(t, ok)
	assert.Equal(t, "https://example.com/path", redir.URL)
	assert.Equal(t, 302, redir.StatusCode)
}

func TestValidateRedirect(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		statusCode int
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid relative URL with 302",
			url:        "/new-page",
			statusCode: 302,
			wantErr:    false,
		},
		{
			name:       "valid absolute URL with 301",
			url:        "https://example.com",
			statusCode: 301,
			wantErr:    false,
		},
		{
			name:       "valid URL with 307",
			url:        "/api/v2/resource",
			statusCode: 307,
			wantErr:    false,
		},
		{
			name:       "valid URL with 308",
			url:        "/permanent",
			statusCode: 308,
			wantErr:    false,
		},
		{
			name:       "invalid status code 200",
			url:        "/new",
			statusCode: 200,
			wantErr:    true,
			errMsg:     "invalid redirect status code 200",
		},
		{
			name:       "invalid status code 404",
			url:        "/new",
			statusCode: 404,
			wantErr:    true,
			errMsg:     "invalid redirect status code 404",
		},
		{
			name:       "invalid status code 500",
			url:        "/new",
			statusCode: 500,
			wantErr:    true,
			errMsg:     "invalid redirect status code 500",
		},
		{
			name:       "empty URL is valid",
			url:        "",
			statusCode: 302,
			wantErr:    false,
		},
		{
			name:       "URL with query params",
			url:        "/search?q=hello&page=1",
			statusCode: 302,
			wantErr:    false,
		},
		{
			name:       "URL with fragment",
			url:        "/page#section",
			statusCode: 302,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRedirect(tt.url, tt.statusCode)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedirectResponse_InExecuteRoute(t *testing.T) {
	interp := NewInterpreter()

	// Create a route that calls redirect()
	route := &Route{
		Method: Get,
		Path:   "/old-page",
		Body: []Statement{
			ReturnStatement{
				Value: FunctionCallExpr{
					Name: "redirect",
					Args: []Expr{
						LiteralExpr{Value: StringLiteral{Value: "/new-page"}},
						LiteralExpr{Value: IntLiteral{Value: 301}},
					},
				},
			},
		},
	}

	request := &Request{
		Path:    "/old-page",
		Method:  "GET",
		Params:  map[string]string{},
		Headers: map[string]string{},
	}

	response, err := interp.ExecuteRoute(route, request)

	require.NoError(t, err)
	assert.Equal(t, 301, response.StatusCode)
	assert.Equal(t, "/new-page", response.Headers["Location"])
	assert.Nil(t, response.Body)
}

func TestRedirectResponse_DefaultCodeInRoute(t *testing.T) {
	interp := NewInterpreter()

	// Create a route that calls redirect() with default status code
	route := &Route{
		Method: Get,
		Path:   "/old",
		Body: []Statement{
			ReturnStatement{
				Value: FunctionCallExpr{
					Name: "redirect",
					Args: []Expr{
						LiteralExpr{Value: StringLiteral{Value: "/new"}},
					},
				},
			},
		},
	}

	request := &Request{
		Path:    "/old",
		Method:  "GET",
		Params:  map[string]string{},
		Headers: map[string]string{},
	}

	response, err := interp.ExecuteRoute(route, request)

	require.NoError(t, err)
	assert.Equal(t, 302, response.StatusCode)
	assert.Equal(t, "/new", response.Headers["Location"])
}
