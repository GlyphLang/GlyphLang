package compiler

import (
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/vm"
)

// TestCompileWebSocketRoute tests basic WebSocket route compilation
func TestCompileWebSocketRoute(t *testing.T) {
	tests := []struct {
		name        string
		route       *interpreter.WebSocketRoute
		expectError bool
	}{
		{
			name: "empty WebSocket route",
			route: &interpreter.WebSocketRoute{
				Path:   "/ws/empty",
				Events: []interpreter.WebSocketEvent{},
			},
			expectError: false,
		},
		{
			name: "WebSocket route with connect event",
			route: &interpreter.WebSocketRoute{
				Path: "/ws/chat",
				Events: []interpreter.WebSocketEvent{
					{
						EventType: interpreter.WSEventConnect,
						Body: []interpreter.Statement{
							&interpreter.AssignStatement{
								Target: "msg",
								Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "connected"}},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "WebSocket route with all event types",
			route: &interpreter.WebSocketRoute{
				Path: "/ws/full",
				Events: []interpreter.WebSocketEvent{
					{
						EventType: interpreter.WSEventConnect,
						Body: []interpreter.Statement{
							&interpreter.AssignStatement{
								Target: "status",
								Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "connected"}},
							},
						},
					},
					{
						EventType: interpreter.WSEventMessage,
						Body: []interpreter.Statement{
							&interpreter.AssignStatement{
								Target: "received",
								Value:  &interpreter.VariableExpr{Name: "input"},
							},
						},
					},
					{
						EventType: interpreter.WSEventDisconnect,
						Body: []interpreter.Statement{
							&interpreter.AssignStatement{
								Target: "status",
								Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "disconnected"}},
							},
						},
					},
					{
						EventType: interpreter.WSEventError,
						Body: []interpreter.Statement{
							&interpreter.AssignStatement{
								Target: "error",
								Value:  &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "error occurred"}},
							},
						},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			compiled, err := c.CompileWebSocketRoute(tt.route)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("CompileWebSocketRoute() error: %v", err)
			}

			if compiled.Path != tt.route.Path {
				t.Errorf("Expected path %s, got %s", tt.route.Path, compiled.Path)
			}
		})
	}
}

// TestCompileWsFunctionCalls tests WebSocket function call compilation
func TestCompileWsFunctionCalls(t *testing.T) {
	tests := []struct {
		name        string
		funcName    string
		args        []interpreter.Expr
		expectError bool
		errorMsg    string
	}{
		{
			name:     "ws.send with one argument",
			funcName: "ws.send",
			args: []interpreter.Expr{
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "hello"}},
			},
			expectError: false,
		},
		{
			name:        "ws.send with no arguments",
			funcName:    "ws.send",
			args:        []interpreter.Expr{},
			expectError: true,
			errorMsg:    "ws.send requires exactly 1 argument",
		},
		{
			name:     "ws.send with too many arguments",
			funcName: "ws.send",
			args: []interpreter.Expr{
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "arg1"}},
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "arg2"}},
			},
			expectError: true,
			errorMsg:    "ws.send requires exactly 1 argument",
		},
		{
			name:     "ws.broadcast with one argument",
			funcName: "ws.broadcast",
			args: []interpreter.Expr{
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "broadcast message"}},
			},
			expectError: false,
		},
		{
			name:     "ws.broadcast with two arguments",
			funcName: "ws.broadcast",
			args: []interpreter.Expr{
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "message"}},
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "except-client"}},
			},
			expectError: false,
		},
		{
			name:        "ws.broadcast with no arguments",
			funcName:    "ws.broadcast",
			args:        []interpreter.Expr{},
			expectError: true,
			errorMsg:    "ws.broadcast requires 1 or 2 arguments",
		},
		{
			name:     "ws.join with room name",
			funcName: "ws.join",
			args: []interpreter.Expr{
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "room-1"}},
			},
			expectError: false,
		},
		{
			name:        "ws.join with no arguments",
			funcName:    "ws.join",
			args:        []interpreter.Expr{},
			expectError: true,
			errorMsg:    "ws.join requires exactly 1 argument",
		},
		{
			name:     "ws.leave with room name",
			funcName: "ws.leave",
			args: []interpreter.Expr{
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "room-1"}},
			},
			expectError: false,
		},
		{
			name:        "ws.leave with no arguments",
			funcName:    "ws.leave",
			args:        []interpreter.Expr{},
			expectError: true,
			errorMsg:    "ws.leave requires exactly 1 argument",
		},
		{
			name:        "ws.close with no arguments",
			funcName:    "ws.close",
			args:        []interpreter.Expr{},
			expectError: false,
		},
		{
			name:     "ws.close with reason",
			funcName: "ws.close",
			args: []interpreter.Expr{
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "connection timeout"}},
			},
			expectError: false,
		},
		{
			name:     "ws.broadcast_to_room with room and message",
			funcName: "ws.broadcast_to_room",
			args: []interpreter.Expr{
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "room-1"}},
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "hello room"}},
			},
			expectError: false,
		},
		{
			name:     "ws.broadcast_to_room with only one argument",
			funcName: "ws.broadcast_to_room",
			args: []interpreter.Expr{
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "room-1"}},
			},
			expectError: true,
			errorMsg:    "ws.broadcast_to_room requires exactly 2 arguments",
		},
		{
			name:        "ws.get_rooms with no arguments",
			funcName:    "ws.get_rooms",
			args:        []interpreter.Expr{},
			expectError: false,
		},
		{
			name:     "ws.get_room_users with room argument",
			funcName: "ws.get_room_users",
			args: []interpreter.Expr{
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "room-1"}},
			},
			expectError: false,
		},
		{
			name:        "ws.get_room_users with no argument",
			funcName:    "ws.get_room_users",
			args:        []interpreter.Expr{},
			expectError: true,
			errorMsg:    "ws.get_room_users requires exactly 1 argument",
		},
		{
			name:     "ws.get_room_clients with room argument",
			funcName: "ws.get_room_clients",
			args: []interpreter.Expr{
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "room-1"}},
			},
			expectError: false,
		},
		{
			name:        "ws.get_connection_count with no arguments",
			funcName:    "ws.get_connection_count",
			args:        []interpreter.Expr{},
			expectError: false,
		},
		{
			name:        "ws.get_uptime with no arguments",
			funcName:    "ws.get_uptime",
			args:        []interpreter.Expr{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			c.symbolTable = c.symbolTable.EnterScope(RouteScope)

			expr := &interpreter.FunctionCallExpr{
				Name: tt.funcName,
				Args: tt.args,
			}

			handled, err := c.compileFunctionCallForWs(expr)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got nil", tt.errorMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("compileFunctionCallForWs() error: %v", err)
			}

			if !handled {
				t.Errorf("Expected function %s to be handled as WebSocket function", tt.funcName)
			}
		})
	}
}

// TestCompileWsStatements tests WebSocket statement compilation
func TestCompileWsStatements(t *testing.T) {
	tests := []struct {
		name        string
		stmt        interpreter.Statement
		expectError bool
	}{
		{
			name: "ws.send statement",
			stmt: &interpreter.WsSendStatement{
				Message: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "hello"}},
			},
			expectError: false,
		},
		{
			name: "ws.broadcast statement without except",
			stmt: &interpreter.WsBroadcastStatement{
				Message: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "broadcast"}},
				Except:  nil,
			},
			expectError: false,
		},
		{
			name: "ws.broadcast statement with except",
			stmt: func() *interpreter.WsBroadcastStatement {
				exceptExpr := interpreter.Expr(&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "client-1"}})
				return &interpreter.WsBroadcastStatement{
					Message: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "broadcast"}},
					Except:  &exceptExpr,
				}
			}(),
			expectError: false,
		},
		{
			name: "ws.close statement",
			stmt: &interpreter.WsCloseStatement{
				Reason: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "goodbye"}},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			c.symbolTable = c.symbolTable.EnterScope(RouteScope)

			err := c.compileWebSocketStatement(tt.stmt)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("compileWebSocketStatement() error: %v", err)
			}

			// Verify bytecode was generated
			if len(c.code) == 0 {
				t.Errorf("Expected bytecode to be generated, but code is empty")
			}
		})
	}
}

// TestCompileModuleTypeDefsOnly tests modules with only type definitions
func TestCompileModuleTypeDefsOnly(t *testing.T) {
	tests := []struct {
		name        string
		module      *interpreter.Module
		expectError bool
		errorMsg    string
	}{
		{
			name: "module with single typedef",
			module: &interpreter.Module{
				Items: []interpreter.Item{
					&interpreter.TypeDef{
						Name: "User",
						Fields: []interpreter.Field{
							{Name: "id", TypeAnnotation: interpreter.IntType{}},
							{Name: "name", TypeAnnotation: interpreter.StringType{}},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "module with multiple typedefs",
			module: &interpreter.Module{
				Items: []interpreter.Item{
					&interpreter.TypeDef{
						Name: "User",
						Fields: []interpreter.Field{
							{Name: "id", TypeAnnotation: interpreter.IntType{}},
							{Name: "name", TypeAnnotation: interpreter.StringType{}},
						},
					},
					&interpreter.TypeDef{
						Name: "Post",
						Fields: []interpreter.Field{
							{Name: "id", TypeAnnotation: interpreter.IntType{}},
							{Name: "title", TypeAnnotation: interpreter.StringType{}},
							{Name: "content", TypeAnnotation: interpreter.StringType{}},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "module with typedef and no route",
			module: &interpreter.Module{
				Items: []interpreter.Item{
					&interpreter.TypeDef{
						Name: "Config",
						Fields: []interpreter.Field{
							{Name: "debug", TypeAnnotation: interpreter.BoolType{}},
						},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			bytecode, err := c.Compile(tt.module)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got nil", tt.errorMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Compile() error: %v", err)
			}

			// Verify bytecode was generated (minimal bytecode for type-only modules)
			if bytecode == nil {
				t.Errorf("Expected bytecode, got nil")
			}
		})
	}
}

// TestCompileEmptyModule tests empty module error handling
func TestCompileEmptyModule(t *testing.T) {
	tests := []struct {
		name        string
		module      *interpreter.Module
		expectError bool
		errorMsg    string
	}{
		{
			name: "completely empty module",
			module: &interpreter.Module{
				Items: []interpreter.Item{},
			},
			expectError: true,
			errorMsg:    "empty module",
		},
		{
			name: "nil items module",
			module: &interpreter.Module{
				Items: nil,
			},
			expectError: true,
			errorMsg:    "empty module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			bytecode, err := c.Compile(tt.module)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got nil", tt.errorMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				if bytecode != nil {
					t.Errorf("Expected nil bytecode for error case, got %v", bytecode)
				}
				return
			}

			if err != nil {
				t.Fatalf("Compile() error: %v", err)
			}
		})
	}
}

// TestCompileModuleWithMixedItems tests modules with both routes and typedefs
func TestCompileModuleWithMixedItems(t *testing.T) {
	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.TypeDef{
				Name: "User",
				Fields: []interpreter.Field{
					{Name: "id", TypeAnnotation: interpreter.IntType{}},
					{Name: "name", TypeAnnotation: interpreter.StringType{}},
				},
			},
			&interpreter.Route{
				Path:   "/users",
				Method: interpreter.Get,
				Body: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "users"}},
					},
				},
			},
		},
	}

	c := NewCompiler()
	bytecode, err := c.Compile(module)
	if err != nil {
		t.Fatalf("Compile() error: %v", err)
	}

	if bytecode == nil {
		t.Errorf("Expected bytecode, got nil")
	}

	// Execute and verify
	vmInstance := vm.NewVM()
	result, err := vmInstance.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	expected := vm.StringValue{Val: "users"}
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestCompileModuleMethod tests the CompileModule method for full module compilation
func TestCompileModuleMethod(t *testing.T) {
	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.TypeDef{
				Name: "Message",
				Fields: []interpreter.Field{
					{Name: "content", TypeAnnotation: interpreter.StringType{}},
					{Name: "sender", TypeAnnotation: interpreter.StringType{}},
				},
			},
			&interpreter.Route{
				Path:   "/api/messages",
				Method: interpreter.Get,
				Body: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.ArrayExpr{
							Elements: []interpreter.Expr{},
						},
					},
				},
			},
			&interpreter.Route{
				Path:   "/api/health",
				Method: interpreter.Get,
				Body: []interpreter.Statement{
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "ok"}},
					},
				},
			},
			&interpreter.WebSocketRoute{
				Path: "/ws/chat",
				Events: []interpreter.WebSocketEvent{
					{
						EventType: interpreter.WSEventConnect,
						Body: []interpreter.Statement{
							&interpreter.AssignStatement{
								Target: "connected",
								Value:  &interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
							},
						},
					},
					{
						EventType: interpreter.WSEventMessage,
						Body:      []interpreter.Statement{},
					},
				},
			},
		},
	}

	c := NewCompiler()
	compiled, err := c.CompileModule(module)
	if err != nil {
		t.Fatalf("CompileModule() error: %v", err)
	}

	// Verify routes
	if len(compiled.Routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(compiled.Routes))
	}

	if _, ok := compiled.Routes["GET /api/messages"]; !ok {
		t.Errorf("Expected route 'GET /api/messages' to exist")
	}

	if _, ok := compiled.Routes["GET /api/health"]; !ok {
		t.Errorf("Expected route 'GET /api/health' to exist")
	}

	// Verify WebSocket routes
	if len(compiled.WebSocketRoutes) != 1 {
		t.Errorf("Expected 1 WebSocket route, got %d", len(compiled.WebSocketRoutes))
	}

	wsRoute, ok := compiled.WebSocketRoutes["/ws/chat"]
	if !ok {
		t.Errorf("Expected WebSocket route '/ws/chat' to exist")
	} else {
		if wsRoute.OnConnect == nil {
			t.Errorf("Expected OnConnect handler to be compiled")
		}
		if wsRoute.OnMessage == nil {
			t.Errorf("Expected OnMessage handler to be compiled")
		}
	}

	// Verify TypeDefs
	if len(compiled.TypeDefs) != 1 {
		t.Errorf("Expected 1 typedef, got %d", len(compiled.TypeDefs))
	}

	if _, ok := compiled.TypeDefs["Message"]; !ok {
		t.Errorf("Expected typedef 'Message' to exist")
	}
}

// TestGetEventName tests the getEventName helper function
func TestGetEventName(t *testing.T) {
	tests := []struct {
		eventType interpreter.WebSocketEventType
		expected  string
	}{
		{interpreter.WSEventConnect, "connect"},
		{interpreter.WSEventMessage, "message"},
		{interpreter.WSEventDisconnect, "disconnect"},
		{interpreter.WSEventError, "error"},
		{interpreter.WebSocketEventType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getEventName(tt.eventType)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestCompileWebSocketEventWithVariables tests WebSocket event compilation with built-in variables
func TestCompileWebSocketEventWithVariables(t *testing.T) {
	// Test that ws, input, and client variables are properly accessible
	route := &interpreter.WebSocketRoute{
		Path: "/ws/test",
		Events: []interpreter.WebSocketEvent{
			{
				EventType: interpreter.WSEventMessage,
				Body: []interpreter.Statement{
					&interpreter.AssignStatement{
						Target: "received",
						Value:  &interpreter.VariableExpr{Name: "input"},
					},
					&interpreter.AssignStatement{
						Target: "clientId",
						Value:  &interpreter.VariableExpr{Name: "client"},
					},
				},
			},
		},
	}

	c := NewCompiler()
	compiled, err := c.CompileWebSocketRoute(route)
	if err != nil {
		t.Fatalf("CompileWebSocketRoute() error: %v", err)
	}

	if compiled.OnMessage == nil {
		t.Errorf("Expected OnMessage bytecode to be generated")
	}

	// Verify bytecode length is reasonable
	if len(compiled.OnMessage) < 5 {
		t.Errorf("OnMessage bytecode seems too short: %d bytes", len(compiled.OnMessage))
	}
}

// TestCompileNonWsFunctionCall tests that non-WS functions are not handled by compileFunctionCallForWs
func TestCompileNonWsFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
	}{
		{"print function", "print"},
		{"len function", "len"},
		{"custom function", "myCustomFunc"},
		{"http.get function", "http.get"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			c.symbolTable = c.symbolTable.EnterScope(RouteScope)

			expr := &interpreter.FunctionCallExpr{
				Name: tt.funcName,
				Args: []interpreter.Expr{},
			}

			handled, err := c.compileFunctionCallForWs(expr)

			if err != nil {
				t.Errorf("Unexpected error for non-WS function: %v", err)
			}

			if handled {
				t.Errorf("Function %s should not be handled as WebSocket function", tt.funcName)
			}
		})
	}
}

// TestCompileWsRoomCount tests the ws.get_room_count function compilation
func TestCompileWsRoomCount(t *testing.T) {
	c := NewCompiler()
	c.symbolTable = c.symbolTable.EnterScope(RouteScope)

	expr := &interpreter.FunctionCallExpr{
		Name: "ws.get_room_count",
		Args: []interpreter.Expr{},
	}

	handled, err := c.compileFunctionCallForWs(expr)

	if err != nil {
		t.Fatalf("compileFunctionCallForWs() error: %v", err)
	}

	if !handled {
		t.Errorf("ws.get_room_count should be handled as WebSocket function")
	}

	// Verify bytecode was generated
	if len(c.code) == 0 {
		t.Errorf("Expected bytecode to be generated")
	}
}

// TestCompileWsRoomUserCount tests the ws.get_room_user_count function compilation
func TestCompileWsRoomUserCount(t *testing.T) {
	tests := []struct {
		name        string
		args        []interpreter.Expr
		expectError bool
		errorMsg    string
	}{
		{
			name: "with room argument",
			args: []interpreter.Expr{
				&interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "room-1"}},
			},
			expectError: false,
		},
		{
			name:        "without argument",
			args:        []interpreter.Expr{},
			expectError: true,
			errorMsg:    "ws.get_room_user_count requires exactly 1 argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			c.symbolTable = c.symbolTable.EnterScope(RouteScope)

			expr := &interpreter.FunctionCallExpr{
				Name: "ws.get_room_user_count",
				Args: tt.args,
			}

			handled, err := c.compileFunctionCallForWs(expr)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got nil", tt.errorMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("compileFunctionCallForWs() error: %v", err)
			}

			if !handled {
				t.Errorf("ws.get_room_user_count should be handled as WebSocket function")
			}
		})
	}
}

// TestWebSocketRouteCompilationPreservesPath tests that the path is preserved after compilation
func TestWebSocketRouteCompilationPreservesPath(t *testing.T) {
	paths := []string{
		"/ws",
		"/ws/chat",
		"/ws/notifications",
		"/api/v1/ws/stream",
		"/socket.io/",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			route := &interpreter.WebSocketRoute{
				Path:   path,
				Events: []interpreter.WebSocketEvent{},
			}

			c := NewCompiler()
			compiled, err := c.CompileWebSocketRoute(route)
			if err != nil {
				t.Fatalf("CompileWebSocketRoute() error: %v", err)
			}

			if compiled.Path != path {
				t.Errorf("Expected path '%s', got '%s'", path, compiled.Path)
			}
		})
	}
}

// TestWebSocketRouteWithPathParams tests that path parameters are available in event handlers
// This is the fix for GitHub issue #64
func TestWebSocketRouteWithPathParams(t *testing.T) {
	tests := []struct {
		name        string
		route       *interpreter.WebSocketRoute
		expectError bool
	}{
		{
			name: "single path parameter used in connect handler",
			route: &interpreter.WebSocketRoute{
				Path: "/chat/:room",
				Events: []interpreter.WebSocketEvent{
					{
						EventType: interpreter.WSEventConnect,
						Body: []interpreter.Statement{
							// ws.join(room) - uses the path parameter 'room'
							&interpreter.ExpressionStatement{
								Expr: &interpreter.FunctionCallExpr{
									Name: "ws.join",
									Args: []interpreter.Expr{
										&interpreter.VariableExpr{Name: "room"},
									},
								},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "single path parameter used in message handler",
			route: &interpreter.WebSocketRoute{
				Path: "/chat/:room",
				Events: []interpreter.WebSocketEvent{
					{
						EventType: interpreter.WSEventMessage,
						Body: []interpreter.Statement{
							// ws.broadcast_to_room(room, input)
							&interpreter.ExpressionStatement{
								Expr: &interpreter.FunctionCallExpr{
									Name: "ws.broadcast_to_room",
									Args: []interpreter.Expr{
										&interpreter.VariableExpr{Name: "room"},
										&interpreter.VariableExpr{Name: "input"},
									},
								},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "multiple path parameters",
			route: &interpreter.WebSocketRoute{
				Path: "/chat/:room/:user",
				Events: []interpreter.WebSocketEvent{
					{
						EventType: interpreter.WSEventConnect,
						Body: []interpreter.Statement{
							// Use both 'room' and 'user' parameters
							&interpreter.AssignStatement{
								Target: "roomName",
								Value:  &interpreter.VariableExpr{Name: "room"},
							},
							&interpreter.AssignStatement{
								Target: "userName",
								Value:  &interpreter.VariableExpr{Name: "user"},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "path parameter used in all event types",
			route: &interpreter.WebSocketRoute{
				Path: "/notifications/:channel",
				Events: []interpreter.WebSocketEvent{
					{
						EventType: interpreter.WSEventConnect,
						Body: []interpreter.Statement{
							&interpreter.ExpressionStatement{
								Expr: &interpreter.FunctionCallExpr{
									Name: "ws.join",
									Args: []interpreter.Expr{
										&interpreter.VariableExpr{Name: "channel"},
									},
								},
							},
						},
					},
					{
						EventType: interpreter.WSEventMessage,
						Body: []interpreter.Statement{
							&interpreter.ExpressionStatement{
								Expr: &interpreter.FunctionCallExpr{
									Name: "ws.broadcast_to_room",
									Args: []interpreter.Expr{
										&interpreter.VariableExpr{Name: "channel"},
										&interpreter.VariableExpr{Name: "input"},
									},
								},
							},
						},
					},
					{
						EventType: interpreter.WSEventDisconnect,
						Body: []interpreter.Statement{
							&interpreter.ExpressionStatement{
								Expr: &interpreter.FunctionCallExpr{
									Name: "ws.leave",
									Args: []interpreter.Expr{
										&interpreter.VariableExpr{Name: "channel"},
									},
								},
							},
						},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			compiled, err := c.CompileWebSocketRoute(tt.route)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("CompileWebSocketRoute() error: %v (path parameters should be accessible)", err)
			}

			// Verify bytecode was generated for the handlers
			if compiled.Path != tt.route.Path {
				t.Errorf("Expected path '%s', got '%s'", tt.route.Path, compiled.Path)
			}

			// Check that we have bytecode for the event handlers
			for _, event := range tt.route.Events {
				switch event.EventType {
				case interpreter.WSEventConnect:
					if len(compiled.OnConnect) == 0 {
						t.Error("Expected OnConnect bytecode to be generated")
					}
				case interpreter.WSEventMessage:
					if len(compiled.OnMessage) == 0 {
						t.Error("Expected OnMessage bytecode to be generated")
					}
				case interpreter.WSEventDisconnect:
					if len(compiled.OnDisconnect) == 0 {
						t.Error("Expected OnDisconnect bytecode to be generated")
					}
				}
			}
		})
	}
}

// TestWebSocketRouteWithUndefinedVariable tests that undefined variables still cause errors
func TestWebSocketRouteWithUndefinedVariable(t *testing.T) {
	route := &interpreter.WebSocketRoute{
		Path: "/chat/:room",
		Events: []interpreter.WebSocketEvent{
			{
				EventType: interpreter.WSEventConnect,
				Body: []interpreter.Statement{
					// Use 'undefinedVar' which is NOT a path parameter
					&interpreter.ExpressionStatement{
						Expr: &interpreter.FunctionCallExpr{
							Name: "ws.join",
							Args: []interpreter.Expr{
								&interpreter.VariableExpr{Name: "undefinedVar"},
							},
						},
					},
				},
			},
		},
	}

	c := NewCompiler()
	_, err := c.CompileWebSocketRoute(route)

	if err == nil {
		t.Error("Expected error for undefined variable, but compilation succeeded")
	}

	if err != nil && !strings.Contains(err.Error(), "undefined") {
		t.Errorf("Expected 'undefined' in error message, got: %v", err)
	}
}
