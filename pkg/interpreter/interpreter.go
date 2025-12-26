package interpreter

import (
	"fmt"
)

// Interpreter is the main interpreter struct
type Interpreter struct {
	globalEnv      *Environment
	functions      map[string]Function
	typeDefs       map[string]TypeDef
	commands       map[string]Command
	cronTasks      []CronTask
	eventHandlers  map[string][]EventHandler
	queueWorkers   map[string]QueueWorker
	typeChecker    *TypeChecker
	dbHandler      interface{}       // Database handler for dependency injection
	moduleResolver *ModuleResolver   // Module resolver for handling imports
	importedModules map[string]*LoadedModule // Imported modules by alias/name
}

// NewInterpreter creates a new interpreter instance
func NewInterpreter() *Interpreter {
	typeChecker := NewTypeChecker()
	return &Interpreter{
		globalEnv:       NewEnvironment(),
		functions:       make(map[string]Function),
		typeDefs:        make(map[string]TypeDef),
		commands:        make(map[string]Command),
		cronTasks:       []CronTask{},
		eventHandlers:   make(map[string][]EventHandler),
		queueWorkers:    make(map[string]QueueWorker),
		typeChecker:     typeChecker,
		moduleResolver:  NewModuleResolver(),
		importedModules: make(map[string]*LoadedModule),
	}
}

// SetModuleResolver sets a custom module resolver
func (i *Interpreter) SetModuleResolver(resolver *ModuleResolver) {
	i.moduleResolver = resolver
}

// GetModuleResolver returns the module resolver
func (i *Interpreter) GetModuleResolver() *ModuleResolver {
	return i.moduleResolver
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
	return i.LoadModuleWithPath(module, ".")
}

// LoadModuleWithPath loads a module into the interpreter with a base path for imports
func (i *Interpreter) LoadModuleWithPath(module Module, basePath string) error {
	for _, item := range module.Items {
		switch it := item.(type) {
		case *ImportStatement:
			// Handle import statements
			if err := i.processImport(it, basePath); err != nil {
				return err
			}

		case *ModuleDecl:
			// Module declaration - just store it for now
			continue

		case *TypeDef:
			i.typeDefs[it.Name] = *it

		case *Function:
			i.functions[it.Name] = *it
			i.globalEnv.Define(it.Name, *it)

		case *Route:
			// Routes are not stored in global env
			// They will be executed directly via ExecuteRoute
			continue

		case *WebSocketRoute:
			// WebSocket routes are handled by the server
			continue

		case *Command:
			i.commands[it.Name] = *it

		case *CronTask:
			i.cronTasks = append(i.cronTasks, *it)

		case *EventHandler:
			i.eventHandlers[it.EventType] = append(i.eventHandlers[it.EventType], *it)

		case *QueueWorker:
			i.queueWorkers[it.QueueName] = *it

		default:
			return fmt.Errorf("unsupported item type: %T", item)
		}
	}

	// Sync typeChecker with loaded types and functions
	i.typeChecker.SetTypeDefs(i.typeDefs)
	i.typeChecker.SetFunctions(i.functions)

	return nil
}

// processImport handles an import statement
func (i *Interpreter) processImport(importStmt *ImportStatement, basePath string) error {
	// Check if module resolver has a parse function set
	if i.moduleResolver.ParseFunc == nil {
		return fmt.Errorf("cannot process imports: no parser function set on module resolver")
	}

	// Resolve and load the module
	loadedModule, err := i.moduleResolver.ResolveModule(importStmt.Path, basePath)
	if err != nil {
		return fmt.Errorf("failed to import '%s': %w", importStmt.Path, err)
	}

	if importStmt.Selective {
		// Selective import: from "path" import { name1, name2 }
		for _, name := range importStmt.Names {
			exported, exists := loadedModule.Exports[name.Name]
			if !exists {
				return fmt.Errorf("'%s' is not exported from module '%s'", name.Name, importStmt.Path)
			}

			// Determine the name to use in the current scope
			importName := name.Name
			if name.Alias != "" {
				importName = name.Alias
			}

			// Add to global environment based on type
			switch exp := exported.(type) {
			case *Function:
				i.functions[importName] = *exp
				i.globalEnv.Define(importName, *exp)
			case *TypeDef:
				i.typeDefs[importName] = *exp
			case *Command:
				i.commands[importName] = *exp
			default:
				i.globalEnv.Define(importName, exp)
			}
		}
	} else {
		// Full module import: import "path" or import "path" as alias
		// Determine the namespace name
		namespace := importStmt.Alias
		if namespace == "" {
			// Extract name from path (e.g., "./utils" -> "utils")
			namespace = extractModuleName(importStmt.Path)
		}

		// Store the loaded module
		i.importedModules[namespace] = loadedModule

		// Create a namespace object with all exports
		moduleObj := make(map[string]interface{})
		for name, exported := range loadedModule.Exports {
			moduleObj[name] = exported
		}

		// Add the module namespace to the global environment
		i.globalEnv.Define(namespace, moduleObj)
	}

	return nil
}

// extractModuleName extracts the module name from an import path
func extractModuleName(path string) string {
	// Remove .glyph extension if present
	name := path
	if len(name) > 6 && name[len(name)-6:] == ".glyph" {
		name = name[:len(name)-6]
	}

	// Get the last component of the path
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '/' || name[i] == '\\' {
			return name[i+1:]
		}
	}

	// Remove leading ./ if present
	if len(name) > 2 && name[:2] == "./" {
		return name[2:]
	}

	return name
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

	// Extract and process query parameters with type conversion
	rawQueryParams := ExtractRawQueryParams(request.Path)
	queryParams, err := ProcessQueryParams(rawQueryParams, route.QueryParams)
	if err != nil {
		return &Response{
			StatusCode: 400,
			Body: map[string]interface{}{
				"error": err.Error(),
			},
		}, err
	}

	// Apply defaults for declared params not in query string
	for _, decl := range route.QueryParams {
		if _, exists := queryParams[decl.Name]; !exists && decl.Default != nil {
			defaultVal, evalErr := i.EvaluateExpression(decl.Default, routeEnv)
			if evalErr != nil {
				return nil, evalErr
			}
			queryParams[decl.Name] = defaultVal
		}
	}

	// Bind query params as 'query' object
	routeEnv.Define("query", queryParams)

	// Also bind declared query params directly as variables
	for _, decl := range route.QueryParams {
		if val, exists := queryParams[decl.Name]; exists {
			routeEnv.Define(decl.Name, val)
		}
	}

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

// GetCommand retrieves a command by name
func (i *Interpreter) GetCommand(name string) (Command, bool) {
	cmd, ok := i.commands[name]
	return cmd, ok
}

// GetCommands returns all registered commands
func (i *Interpreter) GetCommands() map[string]Command {
	return i.commands
}

// GetCronTasks returns all registered cron tasks
func (i *Interpreter) GetCronTasks() []CronTask {
	return i.cronTasks
}

// GetEventHandlers returns all event handlers for a given event type
func (i *Interpreter) GetEventHandlers(eventType string) []EventHandler {
	return i.eventHandlers[eventType]
}

// GetAllEventHandlers returns all registered event handlers
func (i *Interpreter) GetAllEventHandlers() map[string][]EventHandler {
	return i.eventHandlers
}

// GetQueueWorker retrieves a queue worker by queue name
func (i *Interpreter) GetQueueWorker(queueName string) (QueueWorker, bool) {
	worker, ok := i.queueWorkers[queueName]
	return worker, ok
}

// GetQueueWorkers returns all registered queue workers
func (i *Interpreter) GetQueueWorkers() map[string]QueueWorker {
	return i.queueWorkers
}

// ExecuteCommand executes a CLI command with the given arguments
func (i *Interpreter) ExecuteCommand(cmd *Command, args map[string]interface{}) (interface{}, error) {
	// Create a new environment for the command
	cmdEnv := NewChildEnvironment(i.globalEnv)

	// Add command arguments to environment
	for _, param := range cmd.Params {
		if val, ok := args[param.Name]; ok {
			cmdEnv.Define(param.Name, val)
		} else if param.Default != nil {
			// Use default value
			defaultVal, err := i.EvaluateExpression(param.Default, cmdEnv)
			if err != nil {
				return nil, err
			}
			cmdEnv.Define(param.Name, defaultVal)
		} else if param.Required {
			return nil, fmt.Errorf("missing required argument: %s", param.Name)
		}
	}

	// Execute command body
	result, err := i.executeStatements(cmd.Body, cmdEnv)
	if err != nil {
		if retErr, ok := err.(*returnValue); ok {
			result = retErr.value
		} else {
			return nil, err
		}
	}

	// Validate return type if specified
	if cmd.ReturnType != nil {
		if err := i.typeChecker.CheckType(result, cmd.ReturnType); err != nil {
			return nil, fmt.Errorf("return type mismatch in command %s: %v", cmd.Name, err)
		}
	}

	return result, nil
}

// ExecuteCronTask executes a cron task
func (i *Interpreter) ExecuteCronTask(task *CronTask) (interface{}, error) {
	// Create a new environment for the task
	taskEnv := NewChildEnvironment(i.globalEnv)

	// Handle dependency injections
	for _, injection := range task.Injections {
		if _, ok := injection.Type.(DatabaseType); ok {
			if i.dbHandler != nil {
				taskEnv.Define(injection.Name, i.dbHandler)
			}
		}
	}

	// Execute task body
	result, err := i.executeStatements(task.Body, taskEnv)
	if err != nil {
		if retErr, ok := err.(*returnValue); ok {
			result = retErr.value
		} else {
			return nil, err
		}
	}

	return result, nil
}

// ExecuteEventHandler executes an event handler with the given event data
func (i *Interpreter) ExecuteEventHandler(handler *EventHandler, eventData interface{}) (interface{}, error) {
	// Create a new environment for the handler
	handlerEnv := NewChildEnvironment(i.globalEnv)

	// Add event data to environment
	handlerEnv.Define("event", eventData)
	handlerEnv.Define("input", eventData)

	// Handle dependency injections
	for _, injection := range handler.Injections {
		if _, ok := injection.Type.(DatabaseType); ok {
			if i.dbHandler != nil {
				handlerEnv.Define(injection.Name, i.dbHandler)
			}
		}
	}

	// Execute handler body
	result, err := i.executeStatements(handler.Body, handlerEnv)
	if err != nil {
		if retErr, ok := err.(*returnValue); ok {
			result = retErr.value
		} else {
			return nil, err
		}
	}

	return result, nil
}

// EmitEvent triggers all handlers for a given event type
func (i *Interpreter) EmitEvent(eventType string, eventData interface{}) error {
	handlers := i.eventHandlers[eventType]
	for _, handler := range handlers {
		if handler.Async {
			// In a real implementation, this would be executed asynchronously
			go func(h EventHandler) {
				i.ExecuteEventHandler(&h, eventData)
			}(handler)
		} else {
			_, err := i.ExecuteEventHandler(&handler, eventData)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ExecuteQueueWorker executes a queue worker with the given message
func (i *Interpreter) ExecuteQueueWorker(worker *QueueWorker, message interface{}) (interface{}, error) {
	// Create a new environment for the worker
	workerEnv := NewChildEnvironment(i.globalEnv)

	// Add message to environment
	workerEnv.Define("message", message)
	workerEnv.Define("input", message)

	// Handle dependency injections
	for _, injection := range worker.Injections {
		if _, ok := injection.Type.(DatabaseType); ok {
			if i.dbHandler != nil {
				workerEnv.Define(injection.Name, i.dbHandler)
			}
		}
	}

	// Execute worker body
	result, err := i.executeStatements(worker.Body, workerEnv)
	if err != nil {
		if retErr, ok := err.(*returnValue); ok {
			result = retErr.value
		} else {
			return nil, err
		}
	}

	return result, nil
}
