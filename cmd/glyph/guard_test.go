package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveCompiled compiles the first route in src and performs one request
// against the real compiled-route HTTP path.
func serveCompiled(t *testing.T, src, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	route, bytecode := compileFirstRoute(t, src)
	router := server.NewRouter()
	require.NoError(t, registerCompiledRoute(router, route, bytecode, nil))
	handler := createHandler(router)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w
}

func TestGuardFailureReturnsStatusAndError(t *testing.T) {
	src := `@ GET /items/:id {
  $ found = false
  ? found :: 404 "item not found"
  > {id: id}
}`
	w := serveCompiled(t, src, "GET", "/items/1")
	assert.Equal(t, 404, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "item not found", body["error"])
}

func TestGuardPassContinuesRoute(t *testing.T) {
	src := `@ GET /items/:id {
  $ found = true
  ? found :: 404 "item not found"
  > {id: id}
}`
	w := serveCompiled(t, src, "GET", "/items/7")
	assert.Equal(t, 200, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "7", body["id"])
}

func TestReturnStatusCode(t *testing.T) {
	src := `@ POST /items {
  > {ok: true} :: 201
}`
	w := serveCompiled(t, src, "POST", "/items")
	assert.Equal(t, 201, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, true, body["ok"])
}

// serveInterpreted runs the first route in src through the interpreter path
// (`glyph run --interpret`) rather than compiled bytecode.
func serveInterpreted(t *testing.T, src, rawURL string) *interpreter.Response {
	t.Helper()
	module, err := parseSource(src)
	require.NoError(t, err)

	var route *ast.Route
	for _, item := range module.Items {
		if r, ok := item.(*ast.Route); ok {
			route = r
			break
		}
	}
	require.NotNil(t, route)

	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)
	req, err := http.NewRequest("GET", rawURL, nil)
	require.NoError(t, err)
	req.URL = parsed

	ctx := &server.Context{Request: req, PathParams: map[string]string{"id": "9"}}
	resp, err := executeRoute(route, ctx, interpreter.NewInterpreter())
	require.NoError(t, err)
	return resp
}

func TestGuardInterpretedParity(t *testing.T) {
	src := `@ GET /users/:id {
  $ found = false
  ? found :: 404 "user not found"
  > {id: id}
}`
	resp := serveInterpreted(t, src, "http://localhost/users/9")
	assert.Equal(t, 404, resp.StatusCode)

	body, ok := resp.Body.(map[string]interface{})
	require.True(t, ok, "expected map body, got %T", resp.Body)
	assert.Equal(t, "user not found", body["error"])
}

func TestGuardInterpretedPassContinues(t *testing.T) {
	src := `@ GET /users/:id {
  $ found = true
  ? found :: 404 "user not found"
  > {ok: true} :: 200
}`
	resp := serveInterpreted(t, src, "http://localhost/users/9")
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGuardExpressionCondition(t *testing.T) {
	// The exact shape the generation-accuracy eval produced (issue #294).
	src := `@ GET /users/:id {
  $ user = null
  ? user != null :: 404 "user not found"
  > user :: 200
}`
	w := serveCompiled(t, src, "GET", "/users/9")
	assert.Equal(t, 404, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "user not found", body["error"])
}
