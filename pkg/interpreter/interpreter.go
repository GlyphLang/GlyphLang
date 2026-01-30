package interpreter

import (
	"fmt"
	"strings"
	"time"
)

// Interpreter is the main interpreter struct
type Interpreter struct {
	globalEnv       *Environment
	functions       map[string]Function
	typeDefs        map[string]TypeDef
	commands        map[string]Command
	cronTasks       []CronTask
	eventHandlers   map[string][]EventHandler
	queueWorkers    map[string]QueueWorker
	grpcServices    map[string]GRPCService  // key: service name
	grpcHandlers    map[string]GRPCHandler  // key: method name
	testBlocks      []TestBlock
	typeChecker     *TypeChecker
	dbHandler       interface{}              // Database handler for dependency injection
	redisHandler    interface{}              // Redis handler for dependency injection
	mongoDBHandler  interface{}              // MongoDB handler for dependency injection
	llmHandler      interface{}              // LLM handler for AI integration
	moduleResolver  *ModuleResolver          // Module resolver for handling imports
	importedModules map[string]*LoadedModule // Imported modules by alias/name
	constants       map[string]struct{}      // Tracks names that are constants (immutable)
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
		testBlocks:      []TestBlock{},
		eventHandlers:   make(map[string][]EventHandler),
		queueWorkers:    make(map[string]QueueWorker),
		grpcServices:    make(map[string]GRPCService),
		grpcHandlers:    make(map[string]GRPCHandler),
		typeChecker:     typeChecker,
		moduleResolver:  NewModuleResolver(),
		importedModules: make(map[string]*LoadedModule),
		constants:       make(map[string]struct{}),
	}
}

// IsConstant checks if a name refers to a constant (immutable) binding
func (i *Interpreter) IsConstant(name string) bool {
	_, ok := i.constants[name]
	return ok
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
	Path      string
	Method    string
	Params    map[string]string
	Body      interface{}
	Headers   map[string]string
	AuthData  map[string]interface{} // Authenticated user data from JWT
	SSEWriter interface{}            // SSEWriter for SSE routes (implements executor.SSEWriter)
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

		case *GRPCService:
			i.grpcServices[it.Name] = *it

		case *GRPCHandler:
			i.grpcHandlers[it.MethodName] = *it

		case *TestBlock:
			i.testBlocks = append(i.testBlocks, *it)

		case *ConstDecl:
			// Evaluate and store constant at module load time
			value, err := i.EvaluateExpression(it.Value, i.globalEnv)
			if err != nil {
				return fmt.Errorf("error evaluating constant %s: %v", it.Name, err)
			}
			if it.Type != nil {
				if err := i.typeChecker.CheckType(value, it.Type); err != nil {
					return fmt.Errorf("constant %s type mismatch: %v", it.Name, err)
				}
			}
			i.globalEnv.Define(it.Name, value)
			i.constants[it.Name] = struct{}{}

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

// SetRedisHandler sets the Redis handler for dependency injection
func (i *Interpreter) SetRedisHandler(handler interface{}) {
	i.redisHandler = handler
}

// SetMongoDBHandler sets the MongoDB handler for dependency injection
func (i *Interpreter) SetMongoDBHandler(handler interface{}) {
	i.mongoDBHandler = handler
}

// SetLLMHandler sets the LLM handler for AI integration dependency injection
func (i *Interpreter) SetLLMHandler(handler interface{}) {
	i.llmHandler = handler
}

// injectDependency handles a single dependency injection into the given environment.
// It checks for DatabaseType/NamedType{"Database"}, RedisType/NamedType{"Redis"}, MongoDBType/NamedType{"MongoDB"}, and LLMType/NamedType{"LLM"}.
func (i *Interpreter) injectDependency(injection Injection, env *Environment) {
	// Check for DatabaseType or NamedType{Name: "Database"}
	isDB := false
	if _, ok := injection.Type.(DatabaseType); ok {
		isDB = true
	} else if named, ok := injection.Type.(NamedType); ok && named.Name == "Database" {
		isDB = true
	}
	if isDB && i.dbHandler != nil {
		env.Define(injection.Name, i.dbHandler)
		return
	}

	// Check for RedisType or NamedType{Name: "Redis"}
	isRedis := false
	if _, ok := injection.Type.(RedisType); ok {
		isRedis = true
	} else if named, ok := injection.Type.(NamedType); ok && named.Name == "Redis" {
		isRedis = true
	}
	if isRedis && i.redisHandler != nil {
		env.Define(injection.Name, i.redisHandler)
		return
	}

	// Check for MongoDBType or NamedType{Name: "MongoDB"}
	isMongo := false
	if _, ok := injection.Type.(MongoDBType); ok {
		isMongo = true
	} else if named, ok := injection.Type.(NamedType); ok && named.Name == "MongoDB" {
		isMongo = true
	}
	if isMongo && i.mongoDBHandler != nil {
		env.Define(injection.Name, i.mongoDBHandler)
		return
	}

	// Check for LLMType or NamedType{Name: "LLM"}
	isLLM := false
	if _, ok := injection.Type.(LLMType); ok {
		isLLM = true
	} else if named, ok := injection.Type.(NamedType); ok && named.Name == "LLM" {
		isLLM = true
	}
	if isLLM && i.llmHandler != nil {
		env.Define(injection.Name, i.llmHandler)
		return
	}
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
	inputValue := request.Body
	if inputValue != nil {
		// If route has an InputType declared, apply defaults and validate
		if route.InputType != nil {
			if namedType, ok := route.InputType.(NamedType); ok {
				if typeDef, exists := i.typeDefs[namedType.Name]; exists {
					if inputObj, ok := inputValue.(map[string]interface{}); ok {
						// Apply defaults for missing fields
						inputWithDefaults, err := i.ApplyTypeDefaults(inputObj, typeDef, routeEnv)
						if err != nil {
							return &Response{
								StatusCode: 400,
								Body: map[string]interface{}{
									"error": fmt.Sprintf("error applying defaults: %v", err),
								},
							}, err
						}
						inputValue = inputWithDefaults

						// Validate input against the TypeDef
						if err := i.typeChecker.ValidateObjectAgainstTypeDef(inputWithDefaults, typeDef); err != nil {
							return &Response{
								StatusCode: 400,
								Body: map[string]interface{}{
									"error": fmt.Sprintf("input validation failed: %v", err),
								},
							}, err
						}
					}
				}
			}
		}
		routeEnv.Define("input", inputValue)
	} else {
		// Define input as nil/empty map for routes without body
		routeEnv.Define("input", nil)
	}

	// Handle dependency injections
	for _, injection := range route.Injections {
		i.injectDependency(injection, routeEnv)
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

	// For SSE routes (SSE constant defined in ast.go), inject the writer
	// so yield statements can stream events. The writer is provided by the
	// server handler and implements the SSEWriter interface from executor.go.
	// Cleanup/flushing is handled by the server handler after execution.
	if route.Method == SSE && request.SSEWriter != nil {
		routeEnv.Define("__sse_writer", request.SSEWriter)
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

	// SSE routes stream events via yield — no body is returned.
	// Type checking is skipped because the yielded event types are
	// validated individually by the SSEWriter, not as a single return value.
	if route.Method == SSE {
		return &Response{
			StatusCode: 200,
			Headers:    make(map[string]string),
		}, nil
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
		i.injectDependency(injection, routeEnv)
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
		i.injectDependency(injection, taskEnv)
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
		i.injectDependency(injection, handlerEnv)
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
		i.injectDependency(injection, workerEnv)
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

// GetGRPCServices returns a copy of the registered gRPC service definitions
func (i *Interpreter) GetGRPCServices() map[string]GRPCService {
	result := make(map[string]GRPCService, len(i.grpcServices))
	for k, v := range i.grpcServices {
		result[k] = v
	}
	return result
}

// GetGRPCHandlers returns a copy of the registered gRPC handlers
func (i *Interpreter) GetGRPCHandlers() map[string]GRPCHandler {
	result := make(map[string]GRPCHandler, len(i.grpcHandlers))
	for k, v := range i.grpcHandlers {
		result[k] = v
	}
	return result
}

// ExecuteGRPCHandler executes a gRPC handler with the given request.
func (i *Interpreter) ExecuteGRPCHandler(handler *GRPCHandler, args map[string]interface{}, authData map[string]interface{}) (interface{}, error) {
	if handler == nil {
		return nil, fmt.Errorf("handler is nil")
	}

	handlerEnv := NewChildEnvironment(i.globalEnv)

	for _, param := range handler.Params {
		if val, ok := args[param.Name]; ok {
			handlerEnv.Define(param.Name, val)
		} else if param.Required {
			return nil, fmt.Errorf("missing required argument: %s", param.Name)
		} else if param.Default != nil {
			defaultVal, err := i.EvaluateExpression(param.Default, handlerEnv)
			if err != nil {
				return nil, fmt.Errorf("error evaluating default for %s: %v", param.Name, err)
			}
			handlerEnv.Define(param.Name, defaultVal)
		} else {
			handlerEnv.Define(param.Name, nil)
		}
	}

	for _, injection := range handler.Injections {
		i.injectDependency(injection, handlerEnv)
	}

	if handler.Auth != nil && authData != nil {
		handlerEnv.Define("auth", authData)
	}

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

// TestResult represents the result of executing a single test block
type TestResult struct {
	Name     string
	Passed   bool
	Error    string
	Duration time.Duration
}

// GetTestBlocks returns all registered test blocks (returns a copy for safety)
func (i *Interpreter) GetTestBlocks() []TestBlock {
	result := make([]TestBlock, len(i.testBlocks))
	copy(result, i.testBlocks)
	return result
}

// RunTests executes all test blocks and returns results.
func (i *Interpreter) RunTests(filter string) []TestResult {
	var results []TestResult
	for _, test := range i.testBlocks {
		if filter != "" && !matchesFilter(test.Name, filter) {
			continue
		}
		result := i.runSingleTest(test)
		results = append(results, result)
	}
	return results
}

func (i *Interpreter) runSingleTest(test TestBlock) TestResult {
	start := time.Now()
	testEnv := NewChildEnvironment(i.globalEnv)

	_, err := i.executeStatements(test.Body, testEnv)
	duration := time.Since(start)

	if err != nil {
		if _, ok := err.(*returnValue); ok {
			return TestResult{Name: test.Name, Passed: true, Duration: duration}
		}
		return TestResult{Name: test.Name, Passed: false, Error: err.Error(), Duration: duration}
	}

	return TestResult{Name: test.Name, Passed: true, Duration: duration}
}

func matchesFilter(name, filter string) bool {
	if filter == "" {
		return true
	}
	if strings.HasPrefix(filter, "*") && strings.HasSuffix(filter, "*") {
		return strings.Contains(name, filter[1:len(filter)-1])
	}
	if strings.HasPrefix(filter, "*") {
		return strings.HasSuffix(name, filter[1:])
	}
	if strings.HasSuffix(filter, "*") {
		return strings.HasPrefix(name, filter[:len(filter)-1])
	}
	return strings.Contains(name, filter)
}
