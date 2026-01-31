package compiler

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/interpreter"

	"github.com/glyphlang/glyph/pkg/server"
	"github.com/glyphlang/glyph/pkg/vm"
)

// CompiledWebSocketRoute represents a compiled WebSocket route
type CompiledWebSocketRoute struct {
	Path         string
	OnConnect    []byte // Bytecode for connect handler
	OnMessage    []byte // Bytecode for message handler
	OnDisconnect []byte // Bytecode for disconnect handler
	OnError      []byte // Bytecode for error handler
}

// CompileWebSocketRoute compiles a WebSocket route to bytecode
func (c *Compiler) CompileWebSocketRoute(route *interpreter.WebSocketRoute) (*CompiledWebSocketRoute, error) {
	compiled := &CompiledWebSocketRoute{
		Path: route.Path,
	}

	// Compile each event handler
	for _, event := range route.Events {
		bytecode, err := c.compileWebSocketEvent(event, route.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to compile %s handler: %w", getEventName(event.EventType), err)
		}

		switch event.EventType {
		case interpreter.WSEventConnect:
			compiled.OnConnect = bytecode
		case interpreter.WSEventMessage:
			compiled.OnMessage = bytecode
		case interpreter.WSEventDisconnect:
			compiled.OnDisconnect = bytecode
		case interpreter.WSEventError:
			compiled.OnError = bytecode
		}
	}

	return compiled, nil
}

// compileWebSocketEvent compiles a WebSocket event handler
func (c *Compiler) compileWebSocketEvent(event interpreter.WebSocketEvent, routePath string) ([]byte, error) {
	// Create a new compiler for this event handler
	eventCompiler := &Compiler{
		constants:    make([]vm.Value, 0),
		code:         make([]byte, 0),
		symbolTable:  NewGlobalSymbolTable(),
		labelCounter: 0,
		optimizer:    c.optimizer,
	}

	// Enter WebSocket scope
	eventCompiler.symbolTable = eventCompiler.symbolTable.EnterScope(RouteScope)

	// Add WebSocket-specific built-in variables
	// ws - the WebSocket connection context
	wsIdx := eventCompiler.addConstant(vm.StringValue{Val: "ws"})
	eventCompiler.symbolTable.Define("ws", wsIdx)

	// input - the incoming message (for message events)
	inputIdx := eventCompiler.addConstant(vm.StringValue{Val: "input"})
	eventCompiler.symbolTable.Define("input", inputIdx)

	// client - the client ID
	clientIdx := eventCompiler.addConstant(vm.StringValue{Val: "client"})
	eventCompiler.symbolTable.Define("client", clientIdx)

	// Extract and define path parameters from route path (e.g., :room from /chat/:room)
	params := server.ExtractRouteParamNames(routePath)
	for _, param := range params {
		nameIdx := eventCompiler.addConstant(vm.StringValue{Val: param})
		eventCompiler.symbolTable.Define(param, nameIdx)
	}

	// Optimize event body before compilation
	optimizedBody := c.optimizer.OptimizeStatements(event.Body)

	// Compile event body
	for _, stmt := range optimizedBody {
		if err := eventCompiler.compileWebSocketStatement(stmt); err != nil {
			return nil, err
		}
	}

	// Add halt at end
	eventCompiler.emit(vm.OpHalt)

	return eventCompiler.buildBytecode(), nil
}

// compileWebSocketStatement compiles a statement in WebSocket context
func (c *Compiler) compileWebSocketStatement(stmt interpreter.Statement) error {
	switch s := stmt.(type) {
	case *interpreter.WsSendStatement:
		return c.compileWsSend(s)
	case interpreter.WsSendStatement:
		return c.compileWsSend(&s)
	case *interpreter.WsBroadcastStatement:
		return c.compileWsBroadcast(s)
	case interpreter.WsBroadcastStatement:
		return c.compileWsBroadcast(&s)
	case *interpreter.WsCloseStatement:
		return c.compileWsClose(s)
	case interpreter.WsCloseStatement:
		return c.compileWsClose(&s)
	default:
		// Fall back to regular statement compilation
		return c.compileStatement(stmt)
	}
}

// compileWsSend compiles a ws.send statement
func (c *Compiler) compileWsSend(stmt *interpreter.WsSendStatement) error {
	// Compile the message expression
	if err := c.compileExpression(stmt.Message); err != nil {
		return fmt.Errorf("failed to compile ws.send message: %w", err)
	}

	// Emit the send opcode
	c.emit(vm.OpWsSend)

	return nil
}

// compileWsBroadcast compiles a ws.broadcast statement
func (c *Compiler) compileWsBroadcast(stmt *interpreter.WsBroadcastStatement) error {
	// Check if we have an "except" clause
	if stmt.Except != nil {
		// Compile the except expression (client to exclude)
		if err := c.compileExpression(*stmt.Except); err != nil {
			return fmt.Errorf("failed to compile ws.broadcast except: %w", err)
		}
	}

	// Compile the message expression
	if err := c.compileExpression(stmt.Message); err != nil {
		return fmt.Errorf("failed to compile ws.broadcast message: %w", err)
	}

	// Emit the broadcast opcode
	c.emit(vm.OpWsBroadcast)

	return nil
}

// compileWsClose compiles a ws.close statement
func (c *Compiler) compileWsClose(stmt *interpreter.WsCloseStatement) error {
	// Compile the reason expression
	if err := c.compileExpression(stmt.Reason); err != nil {
		return fmt.Errorf("failed to compile ws.close reason: %w", err)
	}

	// Emit the close opcode
	c.emit(vm.OpWsClose)

	return nil
}

// compileFunctionCallForWs handles WebSocket-specific function calls like ws.join, ws.leave, etc.
func (c *Compiler) compileFunctionCallForWs(expr *interpreter.FunctionCallExpr) (bool, error) {
	switch expr.Name {
	case "ws.send":
		if len(expr.Args) != 1 {
			return true, fmt.Errorf("ws.send requires exactly 1 argument")
		}
		if err := c.compileExpression(expr.Args[0]); err != nil {
			return true, err
		}
		c.emit(vm.OpWsSend)
		return true, nil

	case "ws.broadcast":
		if len(expr.Args) < 1 || len(expr.Args) > 2 {
			return true, fmt.Errorf("ws.broadcast requires 1 or 2 arguments")
		}
		if err := c.compileExpression(expr.Args[0]); err != nil {
			return true, err
		}
		c.emit(vm.OpWsBroadcast)
		return true, nil

	case "ws.broadcast_to_room":
		if len(expr.Args) != 2 {
			return true, fmt.Errorf("ws.broadcast_to_room requires exactly 2 arguments (room, message)")
		}
		// Push room name first
		if err := c.compileExpression(expr.Args[0]); err != nil {
			return true, err
		}
		// Then message
		if err := c.compileExpression(expr.Args[1]); err != nil {
			return true, err
		}
		c.emit(vm.OpWsBroadcastRoom)
		return true, nil

	case "ws.join":
		if len(expr.Args) != 1 {
			return true, fmt.Errorf("ws.join requires exactly 1 argument (room)")
		}
		if err := c.compileExpression(expr.Args[0]); err != nil {
			return true, err
		}
		c.emit(vm.OpWsJoinRoom)
		return true, nil

	case "ws.leave":
		if len(expr.Args) != 1 {
			return true, fmt.Errorf("ws.leave requires exactly 1 argument (room)")
		}
		if err := c.compileExpression(expr.Args[0]); err != nil {
			return true, err
		}
		c.emit(vm.OpWsLeaveRoom)
		return true, nil

	case "ws.close":
		reason := &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: ""}}
		if len(expr.Args) == 1 {
			if err := c.compileExpression(expr.Args[0]); err != nil {
				return true, err
			}
		} else {
			if err := c.compileExpression(reason); err != nil {
				return true, err
			}
		}
		c.emit(vm.OpWsClose)
		return true, nil

	case "ws.get_rooms":
		c.emit(vm.OpWsGetRooms)
		return true, nil

	case "ws.get_room_users", "ws.get_room_clients":
		if len(expr.Args) != 1 {
			return true, fmt.Errorf("%s requires exactly 1 argument (room)", expr.Name)
		}
		if err := c.compileExpression(expr.Args[0]); err != nil {
			return true, err
		}
		c.emit(vm.OpWsGetClients)
		return true, nil

	case "ws.get_room_count":
		// Push function name first, then argument (rooms array)
		fnNameIdx := c.addConstant(vm.StringValue{Val: "length"})
		c.emitWithOperand(vm.OpPush, uint32(fnNameIdx))
		c.emit(vm.OpWsGetRooms)
		c.emitWithOperand(vm.OpCall, 1)
		return true, nil

	case "ws.get_room_user_count":
		if len(expr.Args) != 1 {
			return true, fmt.Errorf("ws.get_room_user_count requires exactly 1 argument (room)")
		}
		// Push function name first, then argument (clients array)
		fnNameIdx := c.addConstant(vm.StringValue{Val: "length"})
		c.emitWithOperand(vm.OpPush, uint32(fnNameIdx))
		if err := c.compileExpression(expr.Args[0]); err != nil {
			return true, err
		}
		c.emit(vm.OpWsGetClients)
		c.emitWithOperand(vm.OpCall, 1)
		return true, nil

	case "ws.get_connection_count":
		c.emit(vm.OpWsGetConnCount)
		return true, nil

	case "ws.get_uptime":
		c.emit(vm.OpWsGetUptime)
		return true, nil

	default:
		return false, nil // Not a WebSocket function
	}
}

// getEventName returns a string name for a WebSocket event type
func getEventName(eventType interpreter.WebSocketEventType) string {
	switch eventType {
	case interpreter.WSEventConnect:
		return "connect"
	case interpreter.WSEventMessage:
		return "message"
	case interpreter.WSEventDisconnect:
		return "disconnect"
	case interpreter.WSEventError:
		return "error"
	default:
		return "unknown"
	}
}

// CompileModule compiles a module, handling both HTTP routes and WebSocket routes
func (c *Compiler) CompileModule(module *interpreter.Module) (*CompiledModule, error) {
	result := &CompiledModule{
		Routes:          make(map[string][]byte),
		WebSocketRoutes: make(map[string]*CompiledWebSocketRoute),
		TypeDefs:        make(map[string]*interpreter.TypeDef),
	}

	for _, item := range module.Items {
		switch i := item.(type) {
		case *interpreter.Route:
			bytecode, err := c.CompileRoute(i)
			if err != nil {
				return nil, fmt.Errorf("failed to compile route %s %s: %w", i.Method, i.Path, err)
			}
			key := fmt.Sprintf("%s %s", i.Method, i.Path)
			result.Routes[key] = bytecode

		case *interpreter.WebSocketRoute:
			compiled, err := c.CompileWebSocketRoute(i)
			if err != nil {
				return nil, fmt.Errorf("failed to compile WebSocket route %s: %w", i.Path, err)
			}
			result.WebSocketRoutes[i.Path] = compiled

		case *interpreter.TypeDef:
			result.TypeDefs[i.Name] = i
		}
	}

	return result, nil
}

// CompiledModule represents a fully compiled Glyph module
type CompiledModule struct {
	Routes          map[string][]byte                  // HTTP routes: "METHOD /path" -> bytecode
	WebSocketRoutes map[string]*CompiledWebSocketRoute // WS routes: "/path" -> compiled handlers
	TypeDefs        map[string]*interpreter.TypeDef    // Type definitions
}
