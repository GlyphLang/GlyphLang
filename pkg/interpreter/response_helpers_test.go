package interpreter

import (
	"testing"

	. "github.com/glyphlang/glyph/pkg/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a function call expression with literal arguments.
func callExpr(name string, args ...Expr) FunctionCallExpr {
	return FunctionCallExpr{Name: name, Args: args}
}

func strLit(s string) Expr {
	return LiteralExpr{Value: StringLiteral{Value: s}}
}

func intLit(n int64) Expr {
	return LiteralExpr{Value: IntLiteral{Value: n}}
}

// --- text() tests ---

func TestTextBuiltin_DefaultStatus(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("text", strLit("ok"))
	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)

	tr, ok := result.(*TextResponse)
	require.True(t, ok, "expected *TextResponse, got %T", result)
	assert.Equal(t, "ok", tr.Body)
	assert.Equal(t, 200, tr.StatusCode)
}

func TestTextBuiltin_CustomStatus(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("text", strLit("not found"), intLit(404))
	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)

	tr, ok := result.(*TextResponse)
	require.True(t, ok, "expected *TextResponse, got %T", result)
	assert.Equal(t, "not found", tr.Body)
	assert.Equal(t, 404, tr.StatusCode)
}

func TestTextBuiltin_WrongType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("text", intLit(123))
	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "text() first argument must be a string")
}

func TestTextBuiltin_NoArgs(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("text")
	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "text() requires 1-2 arguments")
}

func TestTextBuiltin_TooManyArgs(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("text", strLit("a"), intLit(200), intLit(300))
	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "text() requires 1-2 arguments")
}

// --- html() tests ---

func TestHTMLBuiltin_DefaultStatus(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("html", strLit("<h1>Hi</h1>"))
	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)

	hr, ok := result.(*HTMLResponse)
	require.True(t, ok, "expected *HTMLResponse, got %T", result)
	assert.Equal(t, "<h1>Hi</h1>", hr.Body)
	assert.Equal(t, 200, hr.StatusCode)
}

func TestHTMLBuiltin_CustomStatus(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("html", strLit("<h1>Created</h1>"), intLit(201))
	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)

	hr, ok := result.(*HTMLResponse)
	require.True(t, ok, "expected *HTMLResponse, got %T", result)
	assert.Equal(t, "<h1>Created</h1>", hr.Body)
	assert.Equal(t, 201, hr.StatusCode)
}

func TestHTMLBuiltin_WrongType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("html", intLit(123))
	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "html() first argument must be a string")
}

func TestHTMLBuiltin_NoArgs(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("html")
	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "html() requires 1-2 arguments")
}

// --- blob() tests ---

func TestBlobBuiltin_StringData(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("blob", strLit("data"), strLit("application/xml"))
	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)

	br, ok := result.(*BlobResponse)
	require.True(t, ok, "expected *BlobResponse, got %T", result)
	assert.Equal(t, []byte("data"), br.Data)
	assert.Equal(t, "application/xml", br.ContentType)
	assert.Equal(t, 200, br.StatusCode)
}

func TestBlobBuiltin_CustomStatus(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("blob", strLit("data"), strLit("application/xml"), intLit(201))
	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)

	br, ok := result.(*BlobResponse)
	require.True(t, ok, "expected *BlobResponse, got %T", result)
	assert.Equal(t, []byte("data"), br.Data)
	assert.Equal(t, "application/xml", br.ContentType)
	assert.Equal(t, 201, br.StatusCode)
}

func TestBlobBuiltin_MissingContentType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("blob", strLit("data"))
	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "blob() requires 2-3 arguments")
}

func TestBlobBuiltin_WrongDataType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("blob", intLit(123), strLit("text/plain"))
	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "blob() first argument must be a string or bytes")
}

func TestBlobBuiltin_WrongContentTypeType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("blob", strLit("data"), intLit(123))
	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "blob() second argument must be a content type string")
}

func TestBlobBuiltin_TooManyArgs(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("blob", strLit("d"), strLit("text/plain"), intLit(200), intLit(0))
	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "blob() requires 2-3 arguments")
}

// --- ExecuteRoute integration tests ---

func TestExecuteRoute_TextResponse(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Method: Get,
		Path:   "/test",
		Body: []Statement{
			ReturnStatement{
				Value: callExpr("text", strLit("hello world")),
			},
		},
	}

	request := &Request{
		Path:    "/test",
		Method:  "GET",
		Params:  map[string]string{},
		Headers: map[string]string{},
	}

	response, err := interp.ExecuteRoute(route, request)
	require.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "hello world", response.Body)
	assert.Equal(t, "text/plain; charset=utf-8", response.Headers["Content-Type"])
}

func TestExecuteRoute_HTMLResponse(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Method: Get,
		Path:   "/test",
		Body: []Statement{
			ReturnStatement{
				Value: callExpr("html", strLit("<p>hello</p>")),
			},
		},
	}

	request := &Request{
		Path:    "/test",
		Method:  "GET",
		Params:  map[string]string{},
		Headers: map[string]string{},
	}

	response, err := interp.ExecuteRoute(route, request)
	require.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "<p>hello</p>", response.Body)
	assert.Equal(t, "text/html; charset=utf-8", response.Headers["Content-Type"])
}

func TestExecuteRoute_BlobResponse(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Method: Get,
		Path:   "/test",
		Body: []Statement{
			ReturnStatement{
				Value: callExpr("blob", strLit("<xml/>"), strLit("application/xml")),
			},
		},
	}

	request := &Request{
		Path:    "/test",
		Method:  "GET",
		Params:  map[string]string{},
		Headers: map[string]string{},
	}

	response, err := interp.ExecuteRoute(route, request)
	require.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, []byte("<xml/>"), response.Body)
	assert.Equal(t, "application/xml", response.Headers["Content-Type"])
}

func TestExecuteRoute_TextResponseCustomStatus(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Method: Get,
		Path:   "/test",
		Body: []Statement{
			ReturnStatement{
				Value: callExpr("text", strLit("not found"), intLit(404)),
			},
		},
	}

	request := &Request{
		Path:    "/test",
		Method:  "GET",
		Params:  map[string]string{},
		Headers: map[string]string{},
	}

	response, err := interp.ExecuteRoute(route, request)
	require.NoError(t, err)
	assert.Equal(t, 404, response.StatusCode)
	assert.Equal(t, "not found", response.Body)
	assert.Equal(t, "text/plain; charset=utf-8", response.Headers["Content-Type"])
}

func TestTextBuiltin_StatusCodeFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test that float64 status code works (e.g., from parsed JSON numbers)
	expr := callExpr("text", strLit("ok"), LiteralExpr{Value: FloatLiteral{Value: 201.0}})
	result, err := interp.EvaluateExpression(expr, env)
	require.NoError(t, err)

	tr, ok := result.(*TextResponse)
	require.True(t, ok)
	assert.Equal(t, 201, tr.StatusCode)
}

func TestTextBuiltin_InvalidStatusCodeType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("text", strLit("ok"), strLit("not-a-number"))
	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "text() second argument must be an integer status code")
}

func TestHTMLBuiltin_InvalidStatusCodeType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("html", strLit("<p>x</p>"), strLit("not-a-number"))
	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "html() second argument must be an integer status code")
}

func TestBlobBuiltin_InvalidStatusCodeType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	expr := callExpr("blob", strLit("data"), strLit("text/plain"), strLit("not-a-number"))
	_, err := interp.EvaluateExpression(expr, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "blob() third argument must be an integer status code")
}
