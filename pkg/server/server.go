package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	ws "github.com/glyphlang/glyph/pkg/websocket"
)

// Server represents the HTTP server
type Server struct {
	router       *Router
	handler      *Handler
	httpServer   *http.Server
	interpreter  Interpreter
	middlewares  []Middleware
	addr         string
	wsServer     *ws.Server // WebSocket server
}

// ServerOption is a functional option for configuring the server
type ServerOption func(*Server)

// NewServer creates a new HTTP server instance
func NewServer(options ...ServerOption) *Server {
	s := &Server{
		router:      NewRouter(),
		middlewares: make([]Middleware, 0),
		addr:        ":8080",
		wsServer:    ws.NewServer(), // Initialize WebSocket server
	}

	// Apply options
	for _, opt := range options {
		opt(s)
	}

	// Create handler
	s.handler = NewHandler(s.router, s.interpreter)

	return s
}

// WithInterpreter sets the interpreter for the server
func WithInterpreter(interpreter Interpreter) ServerOption {
	return func(s *Server) {
		s.interpreter = interpreter
	}
}

// WithAddr sets the server address
func WithAddr(addr string) ServerOption {
	return func(s *Server) {
		s.addr = addr
	}
}

// WithMiddleware adds a global middleware to the server
func WithMiddleware(middleware Middleware) ServerOption {
	return func(s *Server) {
		s.middlewares = append(s.middlewares, middleware)
	}
}

// RegisterRoute registers a single route
func (s *Server) RegisterRoute(route *Route) error {
	// Add global middlewares to the route
	if len(s.middlewares) > 0 {
		route.Middlewares = append(s.middlewares, route.Middlewares...)
	}

	return s.router.RegisterRoute(route)
}

// RegisterRoutes registers multiple routes
func (s *Server) RegisterRoutes(routes []*Route) error {
	for _, route := range routes {
		if err := s.RegisterRoute(route); err != nil {
			return fmt.Errorf("failed to register route %s %s: %w", route.Method, route.Path, err)
		}
	}
	return nil
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	if addr != "" {
		s.addr = addr
	}

	s.httpServer = &http.Server{
		Addr:         s.addr,
		Handler:      s.handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("[SERVER] Starting HTTP server on %s", s.addr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	log.Printf("[SERVER] Shutting down HTTP server...")

	// Shutdown WebSocket server first
	if s.wsServer != nil {
		s.wsServer.Shutdown()
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Printf("[SERVER] HTTP server stopped")
	return nil
}

// GetRouter returns the server's router
func (s *Server) GetRouter() *Router {
	return s.router
}

// GetHandler returns the server's handler
func (s *Server) GetHandler() *Handler {
	return s.handler
}

// GetWebSocketServer returns the WebSocket server
func (s *Server) GetWebSocketServer() *ws.Server {
	return s.wsServer
}

// RegisterWebSocketRoute registers a WebSocket route
func (s *Server) RegisterWebSocketRoute(path string, handler http.HandlerFunc) error {
	// Register as a GET route that upgrades to WebSocket
	route := &Route{
		Method: GET,
		Path:   path,
		Handler: func(ctx *Context) error {
			handler(ctx.ResponseWriter, ctx.Request)
			return nil
		},
	}
	return s.RegisterRoute(route)
}

// MockInterpreter is a simple mock interpreter for testing
type MockInterpreter struct {
	Response interface{}
	Error    error
}

// Execute implements the Interpreter interface
func (m *MockInterpreter) Execute(route *Route, ctx *Context) (interface{}, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	if m.Response != nil {
		return m.Response, nil
	}

	// Flatten single-value query params for backward compatibility
	query := make(map[string]interface{})
	for k, v := range ctx.QueryParams {
		if len(v) == 1 {
			query[k] = v[0]
		} else {
			query[k] = v
		}
	}

	// Default mock response
	response := map[string]interface{}{
		"message":    "Mock response",
		"path":       ctx.Request.URL.Path,
		"method":     ctx.Request.Method,
		"pathParams": ctx.PathParams,
		"query":      query,
	}

	if ctx.Body != nil && len(ctx.Body) > 0 {
		response["body"] = ctx.Body
	}

	return response, nil
}
