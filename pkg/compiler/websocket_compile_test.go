package compiler

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/vm"
)

// TestCompileWebSocketRoute tests basic WebSocket route compilation
func TestCompileWebSocketRoute(t *testing.T) {
	tests := []struct {
		name        string
		route       *ast.WebSocketRoute
		expectError bool
	}{
		{
			name: "empty WebSocket route",
			route: &ast.WebSocketRoute{
				Path:   "/ws/empty",
				Events: []ast.WebSocketEvent{},
			},
			expectError: false,
		},
		{
			name: "WebSocket route with connect event",
			route: &ast.WebSocketRoute{
				Path: "/ws/chat",
				Events: []ast.WebSocketEvent{
					{
						EventType: ast.WSEventConnect,
						Body: []ast.Statement{
							&ast.AssignStatement{
								Target: "msg",
								Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: "connected"}},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "WebSocket route with all event types",
			route: &ast.WebSocketRoute{
				Path: "/ws/full",
				Events: []ast.WebSocketEvent{
					{
						EventType: ast.WSEventConnect,
						Body: []ast.Statement{
							&ast.AssignStatement{
								Target: "status",
								Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: "connected"}},
							},
						},
					},
					{
						EventType: ast.WSEventMessage,
						Body: []ast.Statement{
							&ast.AssignStatement{
								Target: "received",
								Value:  &ast.VariableExpr{Name: "input"},
							},
						},
					},
					{
						EventType: ast.WSEventDisconnect,
						Body: []ast.Statement{
							&ast.AssignStatement{
								Target: "status",
								Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: "disconnected"}},
							},
						},
					},
					{
						EventType: ast.WSEventError,
						Body: []ast.Statement{
							&ast.AssignStatement{
								Target: "error",
								Value:  &ast.LiteralExpr{Value: ast.StringLiteral{Value: "error occurred"}},
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
		args        []ast.Expr
		expectError bool
		errorMsg    string
	}{
		{
			name:     "ws.send with one argument",
			funcName: "ws.send",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "hello"}},
			},
			expectError: false,
		},
		{
			name:        "ws.send with no arguments",
			funcName:    "ws.send",
			args:        []ast.Expr{},
			expectError: true,
			errorMsg:    "ws.send requires exactly 1 argument",
		},
		{
			name:     "ws.send with too many arguments",
			funcName: "ws.send",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "arg1"}},
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "arg2"}},
			},
			expectError: true,
			errorMsg:    "ws.send requires exactly 1 argument",
		},
		{
			name:     "ws.broadcast with one argument",
			funcName: "ws.broadcast",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "broadcast message"}},
			},
			expectError: false,
		},
		{
			name:     "ws.broadcast with two arguments",
			funcName: "ws.broadcast",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "message"}},
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "except-client"}},
			},
			expectError: false,
		},
		{
			name:        "ws.broadcast with no arguments",
			funcName:    "ws.broadcast",
			args:        []ast.Expr{},
			expectError: true,
			errorMsg:    "ws.broadcast requires 1 or 2 arguments",
		},
		{
			name:     "ws.join with room name",
			funcName: "ws.join",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "room-1"}},
			},
			expectError: false,
		},
		{
			name:        "ws.join with no arguments",
			funcName:    "ws.join",
			args:        []ast.Expr{},
			expectError: true,
			errorMsg:    "ws.join requires exactly 1 argument",
		},
		{
			name:     "ws.leave with room name",
			funcName: "ws.leave",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "room-1"}},
			},
			expectError: false,
		},
		{
			name:        "ws.leave with no arguments",
			funcName:    "ws.leave",
			args:        []ast.Expr{},
			expectError: true,
			errorMsg:    "ws.leave requires exactly 1 argument",
		},
		{
			name:        "ws.close with no arguments",
			funcName:    "ws.close",
			args:        []ast.Expr{},
			expectError: false,
		},
		{
			name:     "ws.close with reason",
			funcName: "ws.close",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "connection timeout"}},
			},
			expectError: false,
		},
		{
			name:     "ws.broadcast_to_room with room and message",
			funcName: "ws.broadcast_to_room",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "room-1"}},
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "hello room"}},
			},
			expectError: false,
		},
		{
			name:     "ws.broadcast_to_room with only one argument",
			funcName: "ws.broadcast_to_room",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "room-1"}},
			},
			expectError: true,
			errorMsg:    "ws.broadcast_to_room requires exactly 2 arguments",
		},
		{
			name:        "ws.get_rooms with no arguments",
			funcName:    "ws.get_rooms",
			args:        []ast.Expr{},
			expectError: false,
		},
		{
			name:     "ws.get_room_users with room argument",
			funcName: "ws.get_room_users",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "room-1"}},
			},
			expectError: false,
		},
		{
			name:        "ws.get_room_users with no argument",
			funcName:    "ws.get_room_users",
			args:        []ast.Expr{},
			expectError: true,
			errorMsg:    "ws.get_room_users requires exactly 1 argument",
		},
		{
			name:     "ws.get_room_clients with room argument",
			funcName: "ws.get_room_clients",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "room-1"}},
			},
			expectError: false,
		},
		{
			name:        "ws.get_connection_count with no arguments",
			funcName:    "ws.get_connection_count",
			args:        []ast.Expr{},
			expectError: false,
		},
		{
			name:        "ws.get_uptime with no arguments",
			funcName:    "ws.get_uptime",
			args:        []ast.Expr{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			c.symbolTable = c.symbolTable.EnterScope(RouteScope)

			expr := &ast.FunctionCallExpr{
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
		stmt        ast.Statement
		expectError bool
	}{
		{
			name: "ws.send statement",
			stmt: &ast.WsSendStatement{
				Message: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "hello"}},
			},
			expectError: false,
		},
		{
			name: "ws.broadcast statement without except",
			stmt: &ast.WsBroadcastStatement{
				Message: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "broadcast"}},
				Except:  nil,
			},
			expectError: false,
		},
		{
			name: "ws.broadcast statement with except",
			stmt: func() *ast.WsBroadcastStatement {
				exceptExpr := ast.Expr(&ast.LiteralExpr{Value: ast.StringLiteral{Value: "client-1"}})
				return &ast.WsBroadcastStatement{
					Message: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "broadcast"}},
					Except:  &exceptExpr,
				}
			}(),
			expectError: false,
		},
		{
			name: "ws.close statement",
			stmt: &ast.WsCloseStatement{
				Reason: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "goodbye"}},
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
		module      *ast.Module
		expectError bool
		errorMsg    string
	}{
		{
			name: "module with single typedef",
			module: &ast.Module{
				Items: []ast.Item{
					&ast.TypeDef{
						Name: "User",
						Fields: []ast.Field{
							{Name: "id", TypeAnnotation: ast.IntType{}},
							{Name: "name", TypeAnnotation: ast.StringType{}},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "module with multiple typedefs",
			module: &ast.Module{
				Items: []ast.Item{
					&ast.TypeDef{
						Name: "User",
						Fields: []ast.Field{
							{Name: "id", TypeAnnotation: ast.IntType{}},
							{Name: "name", TypeAnnotation: ast.StringType{}},
						},
					},
					&ast.TypeDef{
						Name: "Post",
						Fields: []ast.Field{
							{Name: "id", TypeAnnotation: ast.IntType{}},
							{Name: "title", TypeAnnotation: ast.StringType{}},
							{Name: "content", TypeAnnotation: ast.StringType{}},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "module with typedef and no route",
			module: &ast.Module{
				Items: []ast.Item{
					&ast.TypeDef{
						Name: "Config",
						Fields: []ast.Field{
							{Name: "debug", TypeAnnotation: ast.BoolType{}},
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
		module      *ast.Module
		expectError bool
		errorMsg    string
	}{
		{
			name: "completely empty module",
			module: &ast.Module{
				Items: []ast.Item{},
			},
			expectError: true,
			errorMsg:    "empty module",
		},
		{
			name: "nil items module",
			module: &ast.Module{
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
	module := &ast.Module{
		Items: []ast.Item{
			&ast.TypeDef{
				Name: "User",
				Fields: []ast.Field{
					{Name: "id", TypeAnnotation: ast.IntType{}},
					{Name: "name", TypeAnnotation: ast.StringType{}},
				},
			},
			&ast.Route{
				Path:   "/users",
				Method: ast.Get,
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "users"}},
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
	module := &ast.Module{
		Items: []ast.Item{
			&ast.TypeDef{
				Name: "Message",
				Fields: []ast.Field{
					{Name: "content", TypeAnnotation: ast.StringType{}},
					{Name: "sender", TypeAnnotation: ast.StringType{}},
				},
			},
			&ast.Route{
				Path:   "/api/messages",
				Method: ast.Get,
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.ArrayExpr{
							Elements: []ast.Expr{},
						},
					},
				},
			},
			&ast.Route{
				Path:   "/api/health",
				Method: ast.Get,
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.StringLiteral{Value: "ok"}},
					},
				},
			},
			&ast.WebSocketRoute{
				Path: "/ws/chat",
				Events: []ast.WebSocketEvent{
					{
						EventType: ast.WSEventConnect,
						Body: []ast.Statement{
							&ast.AssignStatement{
								Target: "connected",
								Value:  &ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
							},
						},
					},
					{
						EventType: ast.WSEventMessage,
						Body:      []ast.Statement{},
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
		eventType ast.WebSocketEventType
		expected  string
	}{
		{ast.WSEventConnect, "connect"},
		{ast.WSEventMessage, "message"},
		{ast.WSEventDisconnect, "disconnect"},
		{ast.WSEventError, "error"},
		{ast.WebSocketEventType(999), "unknown"},
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
	route := &ast.WebSocketRoute{
		Path: "/ws/test",
		Events: []ast.WebSocketEvent{
			{
				EventType: ast.WSEventMessage,
				Body: []ast.Statement{
					&ast.AssignStatement{
						Target: "received",
						Value:  &ast.VariableExpr{Name: "input"},
					},
					&ast.AssignStatement{
						Target: "clientId",
						Value:  &ast.VariableExpr{Name: "client"},
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

			expr := &ast.FunctionCallExpr{
				Name: tt.funcName,
				Args: []ast.Expr{},
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

	expr := &ast.FunctionCallExpr{
		Name: "ws.get_room_count",
		Args: []ast.Expr{},
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
		args        []ast.Expr
		expectError bool
		errorMsg    string
	}{
		{
			name: "with room argument",
			args: []ast.Expr{
				&ast.LiteralExpr{Value: ast.StringLiteral{Value: "room-1"}},
			},
			expectError: false,
		},
		{
			name:        "without argument",
			args:        []ast.Expr{},
			expectError: true,
			errorMsg:    "ws.get_room_user_count requires exactly 1 argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompiler()
			c.symbolTable = c.symbolTable.EnterScope(RouteScope)

			expr := &ast.FunctionCallExpr{
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
			route := &ast.WebSocketRoute{
				Path:   path,
				Events: []ast.WebSocketEvent{},
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
		route       *ast.WebSocketRoute
		expectError bool
	}{
		{
			name: "single path parameter used in connect handler",
			route: &ast.WebSocketRoute{
				Path: "/chat/:room",
				Events: []ast.WebSocketEvent{
					{
						EventType: ast.WSEventConnect,
						Body: []ast.Statement{
							// ws.join(room) - uses the path parameter 'room'
							&ast.ExpressionStatement{
								Expr: &ast.FunctionCallExpr{
									Name: "ws.join",
									Args: []ast.Expr{
										&ast.VariableExpr{Name: "room"},
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
			route: &ast.WebSocketRoute{
				Path: "/chat/:room",
				Events: []ast.WebSocketEvent{
					{
						EventType: ast.WSEventMessage,
						Body: []ast.Statement{
							// ws.broadcast_to_room(room, input)
							&ast.ExpressionStatement{
								Expr: &ast.FunctionCallExpr{
									Name: "ws.broadcast_to_room",
									Args: []ast.Expr{
										&ast.VariableExpr{Name: "room"},
										&ast.VariableExpr{Name: "input"},
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
			route: &ast.WebSocketRoute{
				Path: "/chat/:room/:user",
				Events: []ast.WebSocketEvent{
					{
						EventType: ast.WSEventConnect,
						Body: []ast.Statement{
							// Use both 'room' and 'user' parameters
							&ast.AssignStatement{
								Target: "roomName",
								Value:  &ast.VariableExpr{Name: "room"},
							},
							&ast.AssignStatement{
								Target: "userName",
								Value:  &ast.VariableExpr{Name: "user"},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "path parameter used in all event types",
			route: &ast.WebSocketRoute{
				Path: "/notifications/:channel",
				Events: []ast.WebSocketEvent{
					{
						EventType: ast.WSEventConnect,
						Body: []ast.Statement{
							&ast.ExpressionStatement{
								Expr: &ast.FunctionCallExpr{
									Name: "ws.join",
									Args: []ast.Expr{
										&ast.VariableExpr{Name: "channel"},
									},
								},
							},
						},
					},
					{
						EventType: ast.WSEventMessage,
						Body: []ast.Statement{
							&ast.ExpressionStatement{
								Expr: &ast.FunctionCallExpr{
									Name: "ws.broadcast_to_room",
									Args: []ast.Expr{
										&ast.VariableExpr{Name: "channel"},
										&ast.VariableExpr{Name: "input"},
									},
								},
							},
						},
					},
					{
						EventType: ast.WSEventDisconnect,
						Body: []ast.Statement{
							&ast.ExpressionStatement{
								Expr: &ast.FunctionCallExpr{
									Name: "ws.leave",
									Args: []ast.Expr{
										&ast.VariableExpr{Name: "channel"},
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
				case ast.WSEventConnect:
					if len(compiled.OnConnect) == 0 {
						t.Error("Expected OnConnect bytecode to be generated")
					}
				case ast.WSEventMessage:
					if len(compiled.OnMessage) == 0 {
						t.Error("Expected OnMessage bytecode to be generated")
					}
				case ast.WSEventDisconnect:
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
	route := &ast.WebSocketRoute{
		Path: "/chat/:room",
		Events: []ast.WebSocketEvent{
			{
				EventType: ast.WSEventConnect,
				Body: []ast.Statement{
					// Use 'undefinedVar' which is NOT a path parameter
					&ast.ExpressionStatement{
						Expr: &ast.FunctionCallExpr{
							Name: "ws.join",
							Args: []ast.Expr{
								&ast.VariableExpr{Name: "undefinedVar"},
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
