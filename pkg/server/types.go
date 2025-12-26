package server

import "net/http"

// HTTPMethod represents an HTTP method
type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
	PATCH  HTTPMethod = "PATCH"
)

// Route represents a parsed GLYPH route definition
type Route struct {
	Method      HTTPMethod
	Path        string
	Handler     RouteHandler
	Middlewares []Middleware
}

// RouteHandler is a function that handles a matched route
type RouteHandler func(ctx *Context) error

// Context represents the request context with parsed parameters
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	PathParams     map[string]string
	QueryParams    map[string][]string // All values for each query param
	Body           map[string]interface{}
	StatusCode     int
}

// Middleware is a function that wraps a handler
type Middleware func(next RouteHandler) RouteHandler

// Interpreter interface for executing GLYPH route logic
// This will be properly implemented later - for now we mock it
type Interpreter interface {
	Execute(route *Route, ctx *Context) (interface{}, error)
}
