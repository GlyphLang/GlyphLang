package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"

	"github.com/glyphlang/glyph/pkg/server"
)

// ExampleServer demonstrates basic server usage
func ExampleServer() {
	// Create a new server with mock interpreter
	srv := server.NewServer(
		server.WithInterpreter(&server.MockInterpreter{
			Response: map[string]interface{}{
				"message": "Hello, World!",
			},
		}),
	)

	// Register a simple route
	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/hello",
	})

	// Make a test request
	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()

	srv.GetHandler().ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

// ExampleServer_pathParams demonstrates path parameter extraction
func ExampleServer_pathParams() {
	srv := server.NewServer(
		server.WithInterpreter(&server.MockInterpreter{}),
	)

	// Register route with path parameters
	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/users/:id",
	})

	// Make request
	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	srv.GetHandler().ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	pathParams := response["pathParams"].(map[string]interface{})
	fmt.Println(pathParams["id"])
	// Output: 123
}

// ExampleServer_customHandler demonstrates custom request handlers
func ExampleServer_customHandler() {
	srv := server.NewServer()

	// Register route with custom handler
	srv.RegisterRoute(&server.Route{
		Method: server.POST,
		Path:   "/api/users",
		Handler: func(ctx *server.Context) error {
			// Access request body
			name := ctx.Body["name"].(string)

			// Send custom response
			return server.SendJSON(ctx, 201, map[string]interface{}{
				"message": fmt.Sprintf("User %s created", name),
				"id":      "new-id-123",
			})
		},
	})

	// Make request
	body, _ := json.Marshal(map[string]string{"name": "John"})
	req := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.GetHandler().ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 201
}

// ExampleServer_middleware demonstrates middleware usage
func ExampleServer_middleware() {
	// Create middleware that adds a header
	headerMiddleware := func(next server.RouteHandler) server.RouteHandler {
		return func(ctx *server.Context) error {
			ctx.ResponseWriter.Header().Set("X-Custom", "value")
			return next(ctx)
		}
	}

	srv := server.NewServer()

	// Register route with middleware
	srv.RegisterRoute(&server.Route{
		Method:      server.GET,
		Path:        "/test",
		Middlewares: []server.Middleware{headerMiddleware},
		Handler: func(ctx *server.Context) error {
			return server.SendJSON(ctx, 200, map[string]string{"status": "ok"})
		},
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	srv.GetHandler().ServeHTTP(w, req)

	fmt.Println(w.Header().Get("X-Custom"))
	// Output: value
}
