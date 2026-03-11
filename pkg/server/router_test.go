package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouterEmptyPath(t *testing.T) {
	router := NewRouter()
	err := router.RegisterRoute(&Route{
		Method: GET,
		Path:   "",
	})
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestRouterEmptyParamName(t *testing.T) {
	router := NewRouter()
	err := router.RegisterRoute(&Route{
		Method: GET,
		Path:   "/api/users/:",
	})
	if err == nil {
		t.Error("Expected error for empty parameter name")
	}
}

func TestRouterDefaultMethod(t *testing.T) {
	router := NewRouter()
	route := &Route{
		Path: "/api/data",
	}
	err := router.RegisterRoute(route)
	require.NoError(t, err)

	assert.Equal(t, GET, route.Method)

	matched, _, err := router.Match(GET, "/api/data")
	require.NoError(t, err)
	assert.NotNil(t, matched)
}

func TestRouterGetRoutes(t *testing.T) {
	router := NewRouter()
	router.RegisterRoute(&Route{Method: GET, Path: "/a"})
	router.RegisterRoute(&Route{Method: GET, Path: "/b"})
	router.RegisterRoute(&Route{Method: POST, Path: "/c"})

	getRoutes := router.GetRoutes(GET)
	assert.Len(t, getRoutes, 2)

	postRoutes := router.GetRoutes(POST)
	assert.Len(t, postRoutes, 1)

	deleteRoutes := router.GetRoutes(DELETE)
	assert.Len(t, deleteRoutes, 0)
}

func TestRouterGetAllRoutes(t *testing.T) {
	router := NewRouter()
	router.RegisterRoute(&Route{Method: GET, Path: "/a"})
	router.RegisterRoute(&Route{Method: POST, Path: "/b"})
	router.RegisterRoute(&Route{Method: PUT, Path: "/c"})

	allRoutes := router.GetAllRoutes()
	assert.Len(t, allRoutes, 3)
	assert.Len(t, allRoutes[GET], 1)
	assert.Len(t, allRoutes[POST], 1)
	assert.Len(t, allRoutes[PUT], 1)
}

func TestRouterPathNormalization(t *testing.T) {
	router := NewRouter()
	err := router.RegisterRoute(&Route{
		Method: GET,
		Path:   "api/users",
	})
	require.NoError(t, err)

	matched, _, err := router.Match(GET, "/api/users")
	require.NoError(t, err)
	assert.NotNil(t, matched)

	matched, _, err = router.Match(GET, "api/users")
	require.NoError(t, err)
	assert.NotNil(t, matched)
}

func TestRouterNoRoutesForMethod(t *testing.T) {
	router := NewRouter()
	router.RegisterRoute(&Route{Method: GET, Path: "/api/users"})

	_, _, err := router.Match(DELETE, "/api/users")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no routes registered")
}

func TestServerNewServerDefaults(t *testing.T) {
	s := NewServer()
	assert.NotNil(t, s)
	assert.NotNil(t, s.GetRouter())
	assert.NotNil(t, s.GetHandler())
	assert.NotNil(t, s.GetWebSocketServer())
}

func TestServerWithAddr(t *testing.T) {
	s := NewServer(WithAddr(":9090"))
	assert.Equal(t, ":9090", s.addr)
}

func TestServerWithGlobalMiddleware(t *testing.T) {
	headerSet := false
	middleware := func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			headerSet = true
			return next(ctx)
		}
	}

	s := NewServer(
		WithInterpreter(&MockInterpreter{Response: "ok"}),
		WithMiddleware(middleware),
	)
	s.RegisterRoute(&Route{Method: GET, Path: "/test"})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	s.GetHandler().ServeHTTP(w, req)

	assert.True(t, headerSet, "Global middleware should have been executed")
}

func TestServerStopWithNoHTTPServer(t *testing.T) {
	s := NewServer()
	err := s.Stop(context.Background())
	assert.NoError(t, err)
}

func TestHandlerEmptyJSONBody(t *testing.T) {
	interpreter := &MockInterpreter{}
	s := NewServer(WithInterpreter(interpreter))
	s.RegisterRoute(&Route{Method: POST, Path: "/api/test"})

	req := httptest.NewRequest("POST", "/api/test", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.GetHandler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerWrongContentType(t *testing.T) {
	interpreter := &MockInterpreter{}
	s := NewServer(WithInterpreter(interpreter))
	s.RegisterRoute(&Route{Method: POST, Path: "/api/test"})

	req := httptest.NewRequest("POST", "/api/test", bytes.NewBufferString("name=test"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.GetHandler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendErrorHelper(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := &Context{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	err := SendError(ctx, http.StatusNotFound, "resource not found")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.True(t, response["error"].(bool))
	assert.Equal(t, "resource not found", response["message"])
	assert.Equal(t, float64(http.StatusNotFound), response["code"])
}

func TestSendJSONHelper(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := &Context{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}

	data := map[string]string{"status": "ok"}
	err := SendJSON(ctx, http.StatusCreated, data)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "ok", response["status"])
}

func TestRegisterRoutesError(t *testing.T) {
	s := NewServer()

	routes := []*Route{
		{Method: GET, Path: "/valid"},
		{Method: GET, Path: ""}, // Invalid - empty path
		{Method: GET, Path: "/another"},
	}

	err := s.RegisterRoutes(routes)
	assert.Error(t, err, "RegisterRoutes should fail on invalid route")
}

func TestRegisterWebSocketRoute(t *testing.T) {
	s := NewServer()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	err := s.RegisterWebSocketRoute("/ws", handler)
	assert.NoError(t, err)

	// The route should be accessible as a GET route
	matched, _, err := s.GetRouter().Match(GET, "/ws")
	assert.NoError(t, err)
	assert.NotNil(t, matched)
}
