package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRouterBasic tests basic route registration and matching
func TestRouterBasic(t *testing.T) {
	router := NewRouter()

	route := &Route{
		Method: GET,
		Path:   "/api/users",
	}

	err := router.RegisterRoute(route)
	require.NoError(t, err)

	matched, params, err := router.Match(GET, "/api/users")
	require.NoError(t, err)
	assert.NotNil(t, matched)
	assert.Equal(t, "/api/users", matched.Path)
	assert.Empty(t, params)
}

// TestRouterPathParams tests path parameter extraction
func TestRouterPathParams(t *testing.T) {
	tests := []struct {
		name           string
		routePath      string
		requestPath    string
		expectedParams map[string]string
		shouldMatch    bool
	}{
		{
			name:           "single parameter",
			routePath:      "/api/users/:id",
			requestPath:    "/api/users/123",
			expectedParams: map[string]string{"id": "123"},
			shouldMatch:    true,
		},
		{
			name:           "multiple parameters",
			routePath:      "/api/users/:userId/posts/:postId",
			requestPath:    "/api/users/42/posts/99",
			expectedParams: map[string]string{"userId": "42", "postId": "99"},
			shouldMatch:    true,
		},
		{
			name:           "mixed static and params",
			routePath:      "/api/v1/users/:id/profile",
			requestPath:    "/api/v1/users/456/profile",
			expectedParams: map[string]string{"id": "456"},
			shouldMatch:    true,
		},
		{
			name:        "no match - different path",
			routePath:   "/api/users/:id",
			requestPath: "/api/posts/123",
			shouldMatch: false,
		},
		{
			name:        "no match - extra segments",
			routePath:   "/api/users/:id",
			requestPath: "/api/users/123/extra",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			route := &Route{
				Method: GET,
				Path:   tt.routePath,
			}

			err := router.RegisterRoute(route)
			require.NoError(t, err)

			matched, params, err := router.Match(GET, tt.requestPath)

			if tt.shouldMatch {
				require.NoError(t, err)
				assert.NotNil(t, matched)
				assert.Equal(t, tt.expectedParams, params)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestRouterMultipleMethods tests different HTTP methods
func TestRouterMultipleMethods(t *testing.T) {
	router := NewRouter()

	routes := []*Route{
		{Method: GET, Path: "/api/users"},
		{Method: POST, Path: "/api/users"},
		{Method: PUT, Path: "/api/users/:id"},
		{Method: DELETE, Path: "/api/users/:id"},
		{Method: PATCH, Path: "/api/users/:id"},
	}

	for _, route := range routes {
		err := router.RegisterRoute(route)
		require.NoError(t, err)
	}

	// Test GET
	matched, _, err := router.Match(GET, "/api/users")
	require.NoError(t, err)
	assert.Equal(t, GET, matched.Method)

	// Test POST
	matched, _, err = router.Match(POST, "/api/users")
	require.NoError(t, err)
	assert.Equal(t, POST, matched.Method)

	// Test PUT with param
	matched, params, err := router.Match(PUT, "/api/users/123")
	require.NoError(t, err)
	assert.Equal(t, PUT, matched.Method)
	assert.Equal(t, "123", params["id"])

	// Test DELETE with param
	matched, params, err = router.Match(DELETE, "/api/users/456")
	require.NoError(t, err)
	assert.Equal(t, DELETE, matched.Method)
	assert.Equal(t, "456", params["id"])

	// Test PATCH with param
	matched, params, err = router.Match(PATCH, "/api/users/789")
	require.NoError(t, err)
	assert.Equal(t, PATCH, matched.Method)
	assert.Equal(t, "789", params["id"])
}

// TestHandlerJSONResponse tests JSON response serialization
func TestHandlerJSONResponse(t *testing.T) {
	_ = NewRouter() // Placeholder for future use
	interpreter := &MockInterpreter{
		Response: map[string]interface{}{
			"message": "Hello, World!",
			"status":  "ok",
		},
	}

	server := NewServer(WithInterpreter(interpreter))
	server.RegisterRoute(&Route{
		Method: GET,
		Path:   "/hello",
	})

	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", response["message"])
	assert.Equal(t, "ok", response["status"])
}

// TestHandlerJSONRequest tests JSON request body parsing
func TestHandlerJSONRequest(t *testing.T) {
	_ = NewRouter() // Placeholder for future use
	interpreter := &MockInterpreter{}

	server := NewServer(WithInterpreter(interpreter))
	server.RegisterRoute(&Route{
		Method: POST,
		Path:   "/api/users",
	})

	requestBody := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check that body was parsed
	bodyData, ok := response["body"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "John Doe", bodyData["name"])
	assert.Equal(t, "john@example.com", bodyData["email"])
}

// TestHandlerQueryParams tests query parameter parsing
func TestHandlerQueryParams(t *testing.T) {
	_ = NewRouter() // Placeholder for future use
	interpreter := &MockInterpreter{}

	server := NewServer(WithInterpreter(interpreter))
	server.RegisterRoute(&Route{
		Method: GET,
		Path:   "/api/users",
	})

	req := httptest.NewRequest("GET", "/api/users?page=2&limit=10&sort=name", nil)
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	query, ok := response["query"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "2", query["page"])
	assert.Equal(t, "10", query["limit"])
	assert.Equal(t, "name", query["sort"])
}

// TestHandlerPathParams tests path parameter extraction in handler
func TestHandlerPathParams(t *testing.T) {
	_ = NewRouter() // Placeholder for future use
	interpreter := &MockInterpreter{}

	server := NewServer(WithInterpreter(interpreter))
	server.RegisterRoute(&Route{
		Method: GET,
		Path:   "/api/users/:id",
	})

	req := httptest.NewRequest("GET", "/api/users/42", nil)
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	pathParams, ok := response["pathParams"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "42", pathParams["id"])
}

// TestHandlerNotFound tests 404 error handling
func TestHandlerNotFound(t *testing.T) {
	_ = NewRouter() // Placeholder for future use
	interpreter := &MockInterpreter{}

	server := NewServer(WithInterpreter(interpreter))
	server.RegisterRoute(&Route{
		Method: GET,
		Path:   "/api/users",
	})

	req := httptest.NewRequest("GET", "/api/posts", nil)
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["error"].(bool))
	assert.Contains(t, response["message"], "route not found")
}

// TestHandlerMethodNotAllowed tests different methods on same path
func TestHandlerMethodNotAllowed(t *testing.T) {
	_ = NewRouter() // Placeholder for future use
	interpreter := &MockInterpreter{}

	server := NewServer(WithInterpreter(interpreter))
	server.RegisterRoute(&Route{
		Method: GET,
		Path:   "/api/users",
	})

	// Try POST on a GET-only route
	req := httptest.NewRequest("POST", "/api/users", nil)
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestMiddlewareExecution tests middleware chain execution
func TestMiddlewareExecution(t *testing.T) {
	_ = NewRouter() // Placeholder for future use
	interpreter := &MockInterpreter{
		Response: map[string]string{"result": "success"},
	}

	var executionOrder []string

	middleware1 := func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			executionOrder = append(executionOrder, "middleware1-before")
			err := next(ctx)
			executionOrder = append(executionOrder, "middleware1-after")
			return err
		}
	}

	middleware2 := func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			executionOrder = append(executionOrder, "middleware2-before")
			err := next(ctx)
			executionOrder = append(executionOrder, "middleware2-after")
			return err
		}
	}

	server := NewServer(WithInterpreter(interpreter))
	server.RegisterRoute(&Route{
		Method:      GET,
		Path:        "/test",
		Middlewares: []Middleware{middleware1, middleware2},
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{
		"middleware1-before",
		"middleware2-before",
		"middleware2-after",
		"middleware1-after",
	}, executionOrder)
}

// TestCustomHandler tests custom route handlers
func TestCustomHandler(t *testing.T) {
	_ = NewRouter() // Placeholder for future use

	customHandler := func(ctx *Context) error {
		return SendJSON(ctx, http.StatusOK, map[string]string{
			"custom": "response",
			"id":     ctx.PathParams["id"],
		})
	}

	server := NewServer()
	server.RegisterRoute(&Route{
		Method:  GET,
		Path:    "/custom/:id",
		Handler: customHandler,
	})

	req := httptest.NewRequest("GET", "/custom/123", nil)
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "response", response["custom"])
	assert.Equal(t, "123", response["id"])
}

// TestErrorHandling tests error responses
func TestErrorHandling(t *testing.T) {
	_ = NewRouter() // Placeholder for future use
	interpreter := &MockInterpreter{
		Error: assert.AnError,
	}

	server := NewServer(WithInterpreter(interpreter))
	server.RegisterRoute(&Route{
		Method: GET,
		Path:   "/error",
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["error"].(bool))
}

// TestInvalidJSON tests invalid JSON handling
func TestInvalidJSON(t *testing.T) {
	_ = NewRouter() // Placeholder for future use
	interpreter := &MockInterpreter{}

	server := NewServer(WithInterpreter(interpreter))
	server.RegisterRoute(&Route{
		Method: POST,
		Path:   "/api/users",
	})

	req := httptest.NewRequest("POST", "/api/users", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.GetHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["error"].(bool))
	assert.Contains(t, response["message"], "invalid JSON body")
}

// TestRouteRegistration tests bulk route registration
func TestRouteRegistration(t *testing.T) {
	server := NewServer()

	routes := []*Route{
		{Method: GET, Path: "/api/users"},
		{Method: POST, Path: "/api/users"},
		{Method: GET, Path: "/api/users/:id"},
		{Method: PUT, Path: "/api/users/:id"},
		{Method: DELETE, Path: "/api/users/:id"},
	}

	err := server.RegisterRoutes(routes)
	require.NoError(t, err)

	// Verify all routes are registered
	allRoutes := server.GetRouter().GetAllRoutes()
	assert.Len(t, allRoutes[GET], 2)
	assert.Len(t, allRoutes[POST], 1)
	assert.Len(t, allRoutes[PUT], 1)
	assert.Len(t, allRoutes[DELETE], 1)
}

// BenchmarkRouterMatch benchmarks route matching performance
func BenchmarkRouterMatch(b *testing.B) {
	router := NewRouter()

	routes := []*Route{
		{Method: GET, Path: "/api/users"},
		{Method: GET, Path: "/api/users/:id"},
		{Method: GET, Path: "/api/users/:id/posts"},
		{Method: GET, Path: "/api/users/:id/posts/:postId"},
		{Method: POST, Path: "/api/users"},
		{Method: PUT, Path: "/api/users/:id"},
		{Method: DELETE, Path: "/api/users/:id"},
	}

	for _, route := range routes {
		router.RegisterRoute(route)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.Match(GET, "/api/users/123/posts/456")
	}
}

// BenchmarkHandlerJSON benchmarks JSON handling
func BenchmarkHandlerJSON(b *testing.B) {
	server := NewServer(WithInterpreter(&MockInterpreter{
		Response: map[string]interface{}{
			"id":    123,
			"name":  "Test User",
			"email": "test@example.com",
		},
	}))

	server.RegisterRoute(&Route{
		Method: GET,
		Path:   "/api/users/:id",
	})

	req := httptest.NewRequest("GET", "/api/users/123", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.GetHandler().ServeHTTP(w, req)
	}
}
