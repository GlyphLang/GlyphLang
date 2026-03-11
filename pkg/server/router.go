package server

import (
	"fmt"
	"strings"
)

// Router manages route registration and matching
type Router struct {
	routes map[HTTPMethod][]*RouteNode
}

// RouteNode represents a node in the route tree
type RouteNode struct {
	route      *Route
	pattern    string
	segments   []pathSegment
	paramNames []string
	isStatic   bool
}

// pathSegment represents a segment of a path pattern
type pathSegment struct {
	value     string
	isParam   bool
	paramName string
}

// NewRouter creates a new router instance
func NewRouter() *Router {
	return &Router{
		routes: make(map[HTTPMethod][]*RouteNode),
	}
}

// RegisterRoute adds a route to the router
func (r *Router) RegisterRoute(route *Route) error {
	if route.Method == "" {
		route.Method = GET // Default to GET
	}

	node, err := parseRoutePattern(route)
	if err != nil {
		return err
	}

	if r.routes[route.Method] == nil {
		r.routes[route.Method] = make([]*RouteNode, 0)
	}

	r.routes[route.Method] = append(r.routes[route.Method], node)
	return nil
}

// Match finds a matching route for the given method and path
func (r *Router) Match(method HTTPMethod, path string) (*Route, map[string]string, error) {
	routes, exists := r.routes[method]
	if !exists {
		return nil, nil, fmt.Errorf("no routes registered for method %s", method)
	}

	// Clean the path
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	pathSegments := splitPath(path)

	// Try to match routes in order
	for _, node := range routes {
		if params, matched := matchRoute(node, pathSegments); matched {
			return node.route, params, nil
		}
	}

	return nil, nil, fmt.Errorf("no route matches path %s", path)
}

// GetRoutes returns all registered routes for a method
func (r *Router) GetRoutes(method HTTPMethod) []*Route {
	nodes := r.routes[method]
	routes := make([]*Route, len(nodes))
	for i, node := range nodes {
		routes[i] = node.route
	}
	return routes
}

// GetAllRoutes returns all registered routes
func (r *Router) GetAllRoutes() map[HTTPMethod][]*Route {
	result := make(map[HTTPMethod][]*Route)
	for method, nodes := range r.routes {
		routes := make([]*Route, len(nodes))
		for i, node := range nodes {
			routes[i] = node.route
		}
		result[method] = routes
	}
	return result
}

// parseRoutePattern parses a route pattern into a RouteNode
func parseRoutePattern(route *Route) (*RouteNode, error) {
	pattern := route.Path
	if pattern == "" {
		return nil, fmt.Errorf("route pattern cannot be empty")
	}

	// Clean the pattern
	pattern = strings.TrimSpace(pattern)
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	segments := splitPath(pattern)
	pathSegments := make([]pathSegment, len(segments))
	paramNames := make([]string, 0)
	isStatic := true

	for i, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			// Path parameter
			paramName := seg[1:]
			if paramName == "" {
				return nil, fmt.Errorf("empty parameter name in pattern: %s", pattern)
			}
			pathSegments[i] = pathSegment{
				value:     seg,
				isParam:   true,
				paramName: paramName,
			}
			paramNames = append(paramNames, paramName)
			isStatic = false
		} else {
			// Static segment
			pathSegments[i] = pathSegment{
				value:   seg,
				isParam: false,
			}
		}
	}

	return &RouteNode{
		route:      route,
		pattern:    pattern,
		segments:   pathSegments,
		paramNames: paramNames,
		isStatic:   isStatic,
	}, nil
}

// matchRoute tries to match a route node against path segments
func matchRoute(node *RouteNode, pathSegments []string) (map[string]string, bool) {
	// Check segment count matches
	if len(node.segments) != len(pathSegments) {
		return nil, false
	}

	params := make(map[string]string)

	for i, segment := range node.segments {
		if segment.isParam {
			// Capture parameter value
			params[segment.paramName] = pathSegments[i]
		} else {
			// Static segment must match exactly
			if segment.value != pathSegments[i] {
				return nil, false
			}
		}
	}

	return params, true
}

// splitPath splits a path into segments, removing empty segments
func splitPath(path string) []string {
	parts := strings.Split(path, "/")
	segments := make([]string, 0, len(parts))

	for _, part := range parts {
		if part != "" {
			segments = append(segments, part)
		}
	}

	return segments
}
