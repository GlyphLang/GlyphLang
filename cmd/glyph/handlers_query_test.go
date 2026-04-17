package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// compileFirstRoute parses GLYPH source and returns the first route and its
// compiled bytecode. Used by tests that need to exercise the compiled route
// handler directly.
func compileFirstRoute(t *testing.T, source string) (*ast.Route, []byte) {
	t.Helper()
	module, err := parseSource(source)
	require.NoError(t, err, "parse source")

	var route *ast.Route
	for _, item := range module.Items {
		if r, ok := item.(*ast.Route); ok {
			route = r
			break
		}
	}
	require.NotNil(t, route, "no route found in source")

	c := compiler.NewCompiler()
	bytecode, err := c.CompileRoute(route)
	require.NoError(t, err, "compile route")
	return route, bytecode
}

// invokeCompiledRoute builds a Context from an incoming rawURL and invokes
// the compiled route handler. Returns the HTTP recorder so tests can assert
// on status code and body.
func invokeCompiledRoute(t *testing.T, route *ast.Route, bytecode []byte, method, rawURL string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, rawURL, nil)
	rec := httptest.NewRecorder()
	ctx := &server.Context{
		Request:        req,
		ResponseWriter: rec,
		PathParams:     map[string]string{},
		StatusCode:     http.StatusOK,
	}
	handler := createCompiledRouteHandler(route, bytecode, nil)
	require.NoError(t, handler(ctx), "handler error")
	return rec
}

// TestCompiledRouteQueryAccess verifies that a compiled-mode route with a
// declared query parameter can read `query.X` — the bug reported in #240.
func TestCompiledRouteQueryAccess(t *testing.T) {
	src := `@ GET /api/search {
  ? q: str!
  > {ok: true, q: query.q}
}`
	route, bytecode := compileFirstRoute(t, src)
	rec := invokeCompiledRoute(t, route, bytecode, "GET", "/api/search?q=hello")

	assert.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, true, body["ok"])
	assert.Equal(t, "hello", body["q"])
	assert.NotContains(t, rec.Body.String(), "undefined variable: query")
}

// TestCompiledRouteQueryEmpty verifies that a route with no query param
// declarations still has `query` defined (as an empty object) — i.e. no
// "undefined variable: query" error even when the handler does not touch it.
func TestCompiledRouteQueryEmpty(t *testing.T) {
	src := `@ GET /api/ping {
  > {ok: true}
}`
	route, bytecode := compileFirstRoute(t, src)
	rec := invokeCompiledRoute(t, route, bytecode, "GET", "/api/ping")

	assert.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.NotContains(t, rec.Body.String(), "undefined variable")
}

// TestCompiledRouteQueryTypedParam verifies that declared typed query params
// (e.g. int) are coerced and accessible via `query.X` in compiled mode.
func TestCompiledRouteQueryTypedParam(t *testing.T) {
	src := `@ GET /api/items {
  ? limit: int = 10
  > {limit: query.limit}
}`
	route, bytecode := compileFirstRoute(t, src)
	rec := invokeCompiledRoute(t, route, bytecode, "GET", "/api/items?limit=42")

	assert.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, float64(42), body["limit"])
}

// TestCompiledRouteInvalidQueryParam verifies that a bad type in a declared
// query param surfaces as a 400 Bad Request, matching interpreter behavior.
func TestCompiledRouteInvalidQueryParam(t *testing.T) {
	src := `@ GET /api/items {
  ? limit: int = 10
  > {limit: query.limit}
}`
	route, bytecode := compileFirstRoute(t, src)
	rec := invokeCompiledRoute(t, route, bytecode, "GET", "/api/items?limit=notanumber")

	assert.Equal(t, http.StatusBadRequest, rec.Code, "body=%s", rec.Body.String())
	assert.True(t, strings.Contains(rec.Body.String(), "limit"),
		"error should mention the offending param, got %q", rec.Body.String())
}
