package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRouteRuntimeErrorReturns500 verifies that a route failing at runtime
// reports HTTP 500 rather than 200. Writing the JSON body without calling
// WriteHeader first makes net/http commit an implicit 200, so a crashed
// handler used to look like a success to every client checking status codes.
func TestRouteRuntimeErrorReturns500(t *testing.T) {
	// Redeclaring a binding in the same scope parses and typechecks, then
	// fails during execution - the same shape as the bug found in
	// examples/todo-api.
	src := `@ GET /api/boom {
  $ count = 0
  $ count = 1
  > {count: count}
}`
	module, err := parseSource(src)
	require.NoError(t, err, "parse source")

	var route *ast.Route
	for _, item := range module.Items {
		if r, ok := item.(*ast.Route); ok {
			route = r
			break
		}
	}
	require.NotNil(t, route, "no route found in source")

	req := httptest.NewRequest("GET", "/api/boom", nil)
	rec := httptest.NewRecorder()
	ctx := &server.Context{
		Request:        req,
		ResponseWriter: rec,
		PathParams:     map[string]string{},
		StatusCode:     http.StatusOK,
	}

	handler := createRouteHandler(route, interpreter.NewInterpreter())
	require.NoError(t, handler(ctx), "handler should write a response, not return an error")

	assert.Equal(t, http.StatusInternalServerError, rec.Code, "body=%s", rec.Body.String())
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "Internal server error", body["error"])
}
