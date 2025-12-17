package interpreter

import (
	"fmt"
)

// Interpreter is the main interpreter struct
type Interpreter struct {
	globalEnv   *Environment
	functions   map[string]Function
	typeDefs    map[string]TypeDef
	typeChecker *TypeChecker
	dbHandler   interface{} // Database handler for dependency injection
}

// NewInterpreter creates a new interpreter instance
func NewInterpreter() *Interpreter {
	typeChecker := NewTypeChecker()
	return &Interpreter{
		globalEnv:   NewEnvironment(),
		functions:   make(map[string]Function),
		typeDefs:    make(map[string]TypeDef),
		typeChecker: typeChecker,
	}
}

// Request represents an HTTP request
type Request struct {
	Path     string
	Method   string
	Params   map[string]string
	Body     interface{}
	Headers  map[string]string
	AuthData map[string]interface{} // Authenticated user data from JWT
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
}

// LoadModule loads a module into the interpreter
func (i *Interpreter) LoadModule(module Module) error {
	for _, item := range module.Items {
		switch it := item.(type) {
		case *TypeDef:
			i.typeDefs[it.Name] = *it

		case *Function:
			i.functions[it.Name] = *it
			i.globalEnv.Define(it.Name, *it)

		case *Route:
			// Routes are not stored in global env
			// They will be executed directly via ExecuteRoute
			continue

		default:
			return fmt.Errorf("unsupported item type: %T", item)
		}
	}

	// Sync typeChecker with loaded types and functions
	i.typeChecker.SetTypeDefs(i.typeDefs)
	i.typeChecker.SetFunctions(i.functions)

	return nil
}

// SetDatabaseHandler sets the database handler for dependency injection
func (i *Interpreter) SetDatabaseHandler(handler interface{}) {
	i.dbHandler = handler
}

// ExecuteRoute executes a route with the given request
func (i *Interpreter) ExecuteRoute(route *Route, request *Request) (*Response, error) {
	// Create a new environment for the route
	routeEnv := NewChildEnvironment(i.globalEnv)

	// Extract path parameters
	params, err := extractPathParams(route.Path, request.Path)
	if err != nil {
		return nil, err
	}

	// Add path parameters to environment
	for key, value := range params {
		routeEnv.Define(key, value)
	}

	// Extract and add query parameters to environment
	queryParams := extractQueryParams(request.Path)
	routeEnv.Define("query", queryParams)

	// Always add request body to environment (even if nil)
	// This ensures 'input' variable is always available in routes
	if request.Body != nil {
		routeEnv.Define("input", request.Body)
	} else {
		// Define input as nil/empty map for routes without body
		routeEnv.Define("input", nil)
	}

	// Handle dependency injections
	for _, injection := range route.Injections {
		if _, ok := injection.Type.(DatabaseType); ok {
			if i.dbHandler != nil {
				routeEnv.Define(injection.Name, i.dbHandler)
			}
		}
	}

	// Handle auth injection when route has auth middleware
	if route.Auth != nil {
		// In a real implementation, this would be extracted from JWT validation
		// For now, provide a structure that allows auth.user.id etc. to work
		authData := map[string]interface{}{
			"user": map[string]interface{}{
				"id":       int64(0), // Would be extracted from JWT
				"username": "",
				"role":     "",
			},
			"token":     "",
			"expiresAt": int64(0),
		}
		// If request has auth data attached, use it
		if request.AuthData != nil {
			authData = request.AuthData
		}
		routeEnv.Define("auth", authData)
	}

	// Execute route body
	result, err := i.executeStatements(route.Body, routeEnv)
	if err != nil {
		// Check if it's a return value
		if retErr, ok := err.(*returnValue); ok {
			result = retErr.value
		} else {
			return &Response{
				StatusCode: 500,
				Body: map[string]interface{}{
					"error": err.Error(),
				},
			}, err
		}
	}

	// Validate return value matches declared return type
	if route.ReturnType != nil {
		if err := i.typeChecker.CheckType(result, route.ReturnType); err != nil {
			return &Response{
				StatusCode: 500,
				Body: map[string]interface{}{
					"error": fmt.Sprintf("return type mismatch: %v", err),
				},
			}, fmt.Errorf("return type mismatch in route %s %s: %v", route.Method, route.Path, err)
		}
	}

	// Create response
	response := &Response{
		StatusCode: 200,
		Body:       result,
		Headers:    make(map[string]string),
	}

	return response, nil
}

// ExecuteRouteSimple is a simplified version for testing
func (i *Interpreter) ExecuteRouteSimple(route *Route, pathParams map[string]string) (interface{}, error) {
	// Create a new environment for the route
	routeEnv := NewChildEnvironment(i.globalEnv)

	// Add path parameters to environment
	for key, value := range pathParams {
		routeEnv.Define(key, value)
	}

	// Handle dependency injections
	for _, injection := range route.Injections {
		if _, ok := injection.Type.(DatabaseType); ok {
			if i.dbHandler != nil {
				routeEnv.Define(injection.Name, i.dbHandler)
			}
		}
	}

	// Handle auth injection when route has auth middleware
	if route.Auth != nil {
		authData := map[string]interface{}{
			"user": map[string]interface{}{
				"id":       int64(0),
				"username": "",
				"role":     "",
			},
			"token":     "",
			"expiresAt": int64(0),
		}
		routeEnv.Define("auth", authData)
	}

	// Execute route body
	result, err := i.executeStatements(route.Body, routeEnv)
	if err != nil {
		// Check if it's a return value
		if retErr, ok := err.(*returnValue); ok {
			result = retErr.value
		} else {
			return nil, err
		}
	}

	// Validate return value matches declared return type
	if route.ReturnType != nil {
		if err := i.typeChecker.CheckType(result, route.ReturnType); err != nil {
			return nil, fmt.Errorf("return type mismatch in route %s %s: %v", route.Method, route.Path, err)
		}
	}

	return result, nil
}

// GetTypeDef retrieves a type definition by name
func (i *Interpreter) GetTypeDef(name string) (TypeDef, bool) {
	typeDef, ok := i.typeDefs[name]
	return typeDef, ok
}

// GetFunction retrieves a function by name
func (i *Interpreter) GetFunction(name string) (Function, bool) {
	fn, ok := i.functions[name]
	return fn, ok
}
