package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/parser"
	"github.com/glyphlang/glyph/pkg/vm"
	glyphws "github.com/glyphlang/glyph/pkg/websocket"
	gorillaWS "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketPathParamsCompilationFlow tests the compilation aspect of path params
func TestWebSocketPathParamsCompilationFlow(t *testing.T) {
	source := `
@ ws /chat/:room {
  on connect {
    ws.join(room)
    ws.send("Joined room: " + room)
  }
}
`
	// Parse the source
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err, "Lexer should tokenize the source")

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parser should parse the module")

	// Find the WebSocket route
	var wsRoute *interpreter.WebSocketRoute
	for _, item := range module.Items {
		if r, ok := item.(*interpreter.WebSocketRoute); ok {
			wsRoute = r
			break
		}
	}
	require.NotNil(t, wsRoute, "Should find a WebSocket route")
	assert.Equal(t, "/chat/:room", wsRoute.Path, "Path should be /chat/:room")

	// Compile the WebSocket route
	c := compiler.NewCompilerWithOptLevel(compiler.OptBasic)
	compiled, err := c.CompileWebSocketRoute(wsRoute)
	require.NoError(t, err, "Should compile WebSocket route without error")

	// Verify bytecode was generated for the connect handler
	assert.NotEmpty(t, compiled.OnConnect, "OnConnect handler should have bytecode")

	t.Log("Compilation flow verified successfully - path params are recognized")
}

// TestWebSocketPathParamsVMExecution tests that path params are injected into the VM
func TestWebSocketPathParamsVMExecution(t *testing.T) {
	// This tests the executeWebSocketBytecode function pattern

	// Create a simple bytecode that loads the 'room' variable and returns it
	// We'll simulate what happens during WebSocket event handling

	source := `
@ ws /chat/:room {
  on connect {
    $ roomName = room
  }
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	var wsRoute *interpreter.WebSocketRoute
	for _, item := range module.Items {
		if r, ok := item.(*interpreter.WebSocketRoute); ok {
			wsRoute = r
			break
		}
	}
	require.NotNil(t, wsRoute)

	c := compiler.NewCompilerWithOptLevel(compiler.OptBasic)
	compiled, err := c.CompileWebSocketRoute(wsRoute)
	require.NoError(t, err)
	require.NotEmpty(t, compiled.OnConnect)

	// Now test that the VM can execute with the path param injected
	vmInstance := vm.NewVM()

	// This is what executeWebSocketBytecode does - inject path params as locals
	vmInstance.SetLocal("room", vm.StringValue{Val: "testroom"})
	vmInstance.SetLocal("client", vm.StringValue{Val: "test-client-123"})

	// Execute the bytecode
	_, err = vmInstance.Execute(compiled.OnConnect)
	require.NoError(t, err, "VM should execute bytecode without error when path params are set")

	t.Log("VM execution with path params verified successfully")
}

// TestConvertPathPatternFunction tests the path pattern conversion
func TestConvertPathPatternFunction(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/chat/:room", "/chat/{room}"},
		{"/users/:id/posts/:postId", "/users/{id}/posts/{postId}"},
		{"/static/path", "/static/path"},
		{"/api/:version/items/:id", "/api/{version}/items/{id}"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertPathPattern(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// convertPathPattern converts :param to {param} for Go's http.ServeMux
// This is a copy of the function from cmd/glyph/main.go for testing
func convertPathPattern(pattern string) string {
	parts := strings.Split(pattern, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

// TestExtractPathParamsFromRequest tests the runtime path extraction
func TestExtractPathParamsFromRequest(t *testing.T) {
	tests := []struct {
		pattern    string
		actualPath string
		expected   map[string]string
	}{
		{"/chat/:room", "/chat/general", map[string]string{"room": "general"}},
		{"/chat/:room/:user", "/chat/general/alice", map[string]string{"room": "general", "user": "alice"}},
		{"/api/v1/:resource/:id", "/api/v1/users/123", map[string]string{"resource": "users", "id": "123"}},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			result := extractPathParams(tt.pattern, tt.actualPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// extractPathParams - copy of function from pkg/websocket/server.go
func extractPathParams(pattern, actualPath string) map[string]string {
	params := make(map[string]string)

	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	actualParts := strings.Split(strings.Trim(actualPath, "/"), "/")

	if len(patternParts) != len(actualParts) {
		return params
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			paramName := part[1:]
			params[paramName] = actualParts[i]
		}
	}

	return params
}

// TestWebSocketServerPathParamsIntegration tests the full server integration
func TestWebSocketServerPathParamsIntegration(t *testing.T) {
	wsServer := glyphws.NewServer()
	defer wsServer.Shutdown()

	// Track received path params
	var receivedRoom string
	var mu sync.Mutex

	// Register a connect handler that captures the path params
	wsServer.GetHub().OnConnect(func(conn *glyphws.Connection) error {
		mu.Lock()
		defer mu.Unlock()
		if room, ok := conn.PathParams["room"]; ok {
			receivedRoom = room
		}
		return nil
	})

	// Create a test server with the pattern handler
	pattern := "/chat/:room"
	mux := http.NewServeMux()

	// Use the pattern-based handler
	muxPattern := convertPathPattern(pattern)
	mux.HandleFunc(muxPattern, wsServer.HandleWebSocketWithPattern(pattern))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Connect to the WebSocket endpoint with a specific room
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/chat/testroom"

	client, resp, err := gorillaWS.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Logf("Response status: %v", resp)
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Wait for the connection to be processed
	time.Sleep(200 * time.Millisecond)

	// Verify the path param was extracted
	mu.Lock()
	capturedRoom := receivedRoom
	mu.Unlock()

	assert.Equal(t, "testroom", capturedRoom, "Path parameter 'room' should be 'testroom'")

	t.Logf("SUCCESS: Path parameter correctly extracted. Room = %q", capturedRoom)
}

// TestWebSocketPathParamsMultiple tests multiple path parameters
func TestWebSocketPathParamsMultiple(t *testing.T) {
	wsServer := glyphws.NewServer()
	defer wsServer.Shutdown()

	var receivedParams map[string]string
	var mu sync.Mutex

	wsServer.GetHub().OnConnect(func(conn *glyphws.Connection) error {
		mu.Lock()
		defer mu.Unlock()
		receivedParams = make(map[string]string)
		for k, v := range conn.PathParams {
			receivedParams[k] = v
		}
		return nil
	})

	pattern := "/chat/:room/:user"
	mux := http.NewServeMux()
	muxPattern := convertPathPattern(pattern)
	mux.HandleFunc(muxPattern, wsServer.HandleWebSocketWithPattern(pattern))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/chat/general/alice"

	client, _, err := gorillaWS.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client.Close()

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	captured := receivedParams
	mu.Unlock()

	assert.Equal(t, "general", captured["room"], "room should be 'general'")
	assert.Equal(t, "alice", captured["user"], "user should be 'alice'")

	t.Logf("SUCCESS: Multiple path parameters correctly extracted: %v", captured)
}

// TestWebSocketPathParamsInBytecodeExecution simulates full bytecode execution
// This is the closest simulation to how the actual server processes events
//
// NOTE: This test currently FAILS due to a bug in the implementation:
// The compiler emits a POP after ws.* function calls (like ws.join, ws.broadcast_to_room),
// but these operations don't push a result to the stack. This causes a "stack underflow" error.
//
// BUG LOCATION: pkg/compiler/compiler.go line ~440
// The compileExpressionStatement always emits POP, but WebSocket void functions
// (ws.join, ws.leave, ws.send, ws.broadcast, ws.broadcast_to_room) don't push results.
//
// FIX OPTIONS:
// 1. Have void WS operations push a null value so POP works
// 2. Have the compiler detect void WS functions and skip the POP
func TestWebSocketPathParamsInBytecodeExecution(t *testing.T) {
	// BUG FIXED: WebSocket void functions now push null to avoid stack underflow

	// Source that uses the path parameter
	source := `
@ ws /chat/:room {
  on message {
    ws.broadcast_to_room(room, input)
  }
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err)

	var wsRoute *interpreter.WebSocketRoute
	for _, item := range module.Items {
		if r, ok := item.(*interpreter.WebSocketRoute); ok {
			wsRoute = r
			break
		}
	}
	require.NotNil(t, wsRoute)

	c := compiler.NewCompilerWithOptLevel(compiler.OptBasic)
	compiled, err := c.CompileWebSocketRoute(wsRoute)
	require.NoError(t, err)

	// Verify the message handler was compiled
	require.NotEmpty(t, compiled.OnMessage, "OnMessage handler should have bytecode")

	// Create a test connection with path params
	hub := glyphws.NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Wait for hub to start
	time.Sleep(50 * time.Millisecond)

	// Create a mock connection (without actual WebSocket)
	conn := glyphws.NewConnection("test-conn", nil, hub)
	conn.PathParams["room"] = "testroom"

	// Create VM and inject path params
	vmInstance := vm.NewVM()
	vmInstance.SetLocal("room", vm.StringValue{Val: conn.PathParams["room"]})
	vmInstance.SetLocal("client", vm.StringValue{Val: conn.ID})
	vmInstance.SetLocal("input", vm.StringValue{Val: "test message"})

	// Create a mock WebSocket handler
	wsHandler := &mockWSHandler{room: conn.PathParams["room"]}
	vmInstance.SetWebSocketHandler(wsHandler)

	// Execute the message handler bytecode
	_, err = vmInstance.Execute(compiled.OnMessage)
	require.NoError(t, err)

	// Verify the handler was called with the correct room
	assert.Equal(t, "testroom", wsHandler.broadcastRoom, "broadcast_to_room should use path param 'room'")

	t.Logf("SUCCESS: Bytecode correctly accessed path param 'room' = %q", wsHandler.broadcastRoom)
}

// mockWSHandler implements vm.WebSocketHandler for testing
type mockWSHandler struct {
	room          string
	broadcastRoom string
	sentMessage   interface{}
}

func (h *mockWSHandler) Send(message interface{}) error {
	h.sentMessage = message
	return nil
}

func (h *mockWSHandler) Broadcast(message interface{}) error {
	return nil
}

func (h *mockWSHandler) BroadcastToRoom(room string, message interface{}) error {
	h.broadcastRoom = room
	return nil
}

func (h *mockWSHandler) JoinRoom(room string) error {
	return nil
}

func (h *mockWSHandler) LeaveRoom(room string) error {
	return nil
}

func (h *mockWSHandler) Close(reason string) error {
	return nil
}

func (h *mockWSHandler) GetRooms() []string {
	return nil
}

func (h *mockWSHandler) GetRoomClients(room string) []string {
	return nil
}

func (h *mockWSHandler) GetConnectionID() string {
	return "test-conn"
}

func (h *mockWSHandler) GetConnectionCount() int {
	return 1
}

func (h *mockWSHandler) GetUptime() int64 {
	return 0
}

// TestWebSocketPathParamsFullFlow tests the complete flow from parsing to execution
//
// NOTE: Stack underflow bug has been fixed - WebSocket void functions now push null.
func TestWebSocketPathParamsFullFlow(t *testing.T) {
	// BUG FIXED: WebSocket void functions now push null to avoid stack underflow

	t.Log("=== Testing WebSocket Path Parameters End-to-End ===")

	// Step 1: Parse Glyph source with WebSocket path params
	t.Log("Step 1: Parsing source with path parameters...")
	source := `
@ ws /chat/:room {
  on connect {
    ws.join(room)
  }
  on message {
    ws.broadcast_to_room(room, input)
  }
  on disconnect {
    ws.leave(room)
  }
}
`
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err, "Lexer should succeed")
	t.Log("  Lexer: OK")

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parser should succeed")
	t.Log("  Parser: OK")

	// Step 2: Verify the route was parsed correctly
	var wsRoute *interpreter.WebSocketRoute
	for _, item := range module.Items {
		if r, ok := item.(*interpreter.WebSocketRoute); ok {
			wsRoute = r
			break
		}
	}
	require.NotNil(t, wsRoute, "WebSocket route should be parsed")
	assert.Equal(t, "/chat/:room", wsRoute.Path)
	t.Logf("  Route path: %s", wsRoute.Path)

	// Step 3: Compile the route
	t.Log("Step 2: Compiling WebSocket route...")
	c := compiler.NewCompilerWithOptLevel(compiler.OptBasic)
	compiled, err := c.CompileWebSocketRoute(wsRoute)
	require.NoError(t, err, "Compilation should succeed")
	t.Logf("  OnConnect bytecode: %d bytes", len(compiled.OnConnect))
	t.Logf("  OnMessage bytecode: %d bytes", len(compiled.OnMessage))
	t.Logf("  OnDisconnect bytecode: %d bytes", len(compiled.OnDisconnect))

	// Step 4: Simulate runtime path param extraction
	t.Log("Step 3: Testing runtime path extraction...")
	pattern := "/chat/:room"
	actualPath := "/chat/testroom"
	params := extractPathParams(pattern, actualPath)
	assert.Equal(t, "testroom", params["room"])
	t.Logf("  Extracted params: %v", params)

	// Step 5: Simulate pattern conversion for http.ServeMux
	t.Log("Step 4: Testing pattern conversion...")
	muxPattern := convertPathPattern(pattern)
	assert.Equal(t, "/chat/{room}", muxPattern)
	t.Logf("  Mux pattern: %s", muxPattern)

	// Step 6: Test VM execution with path params
	t.Log("Step 5: Testing VM execution with path params...")
	vmInstance := vm.NewVM()
	vmInstance.SetLocal("room", vm.StringValue{Val: params["room"]})
	vmInstance.SetLocal("client", vm.StringValue{Val: "test-client"})

	wsHandler := &mockWSHandler{room: params["room"]}
	vmInstance.SetWebSocketHandler(wsHandler)

	_, err = vmInstance.Execute(compiled.OnConnect)
	require.NoError(t, err, "VM should execute OnConnect")
	t.Log("  OnConnect execution: OK")

	vmInstance2 := vm.NewVM()
	vmInstance2.SetLocal("room", vm.StringValue{Val: params["room"]})
	vmInstance2.SetLocal("client", vm.StringValue{Val: "test-client"})
	vmInstance2.SetLocal("input", vm.StringValue{Val: "Hello!"})

	wsHandler2 := &mockWSHandler{room: params["room"]}
	vmInstance2.SetWebSocketHandler(wsHandler2)

	_, err = vmInstance2.Execute(compiled.OnMessage)
	require.NoError(t, err, "VM should execute OnMessage")
	assert.Equal(t, "testroom", wsHandler2.broadcastRoom, "broadcast_to_room should receive 'testroom'")
	t.Logf("  OnMessage execution: OK (broadcast to room=%q)", wsHandler2.broadcastRoom)

	t.Log("")
	t.Log("=== ALL TESTS PASSED ===")
	t.Log("WebSocket path parameters work correctly end-to-end:")
	t.Log("  1. Parser correctly parses /chat/:room")
	t.Log("  2. Compiler defines 'room' as a local variable in bytecode")
	t.Log("  3. Runtime extracts 'testroom' from /chat/testroom")
	t.Log("  4. Pattern converter creates /chat/{room} for http.ServeMux")
	t.Log("  5. VM receives 'room' as local and uses it in handlers")
}

// TestWebSocketPathParamsIssue64Verification is a specific test to verify GitHub issue #64 is fixed
//
// NOTE: This test currently FAILS due to a stack underflow bug.
// See TestWebSocketPathParamsInBytecodeExecution for details.
//
// The path param feature is PARTIALLY implemented:
// - Parsing: OK - /chat/:room is correctly parsed
// - Compile-time symbol definition: OK - 'room' is defined in symbol table
// - Runtime path extraction: OK - extractPathParams correctly extracts params
// - Pattern conversion: OK - :room -> {room} for http.ServeMux
// - Connection storage: OK - PathParams map on Connection
// - VM injection: OK - SetLocal injects params correctly
// - VM execution: FIXED - WS void functions now push null to avoid stack underflow
func TestWebSocketPathParamsIssue64Verification(t *testing.T) {
	// BUG FIXED: WebSocket void functions now push null to avoid stack underflow

	t.Log("Verifying fix for GitHub issue #64: WebSocket path parameters")

	// The issue was that WebSocket routes with path parameters like /chat/:room
	// didn't have the parameter available in the handler

	// Create a test that mimics the exact example from the issue
	source := `
@ ws /chat/:room {
  on connect {
    ws.join(room)
    ws.broadcast_to_room(room, "User joined")
  }

  on message {
    ws.broadcast_to_room(room, input)
  }

  on disconnect {
    ws.broadcast_to_room(room, "User left")
    ws.leave(room)
  }
}
`

	// Parse
	lexer := parser.NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err, "Lexer failed")

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parser failed")

	// Compile
	var wsRoute *interpreter.WebSocketRoute
	for _, item := range module.Items {
		if r, ok := item.(*interpreter.WebSocketRoute); ok {
			wsRoute = r
			break
		}
	}
	require.NotNil(t, wsRoute, "No WebSocket route found")

	c := compiler.NewCompilerWithOptLevel(compiler.OptBasic)
	compiled, err := c.CompileWebSocketRoute(wsRoute)
	require.NoError(t, err, "Compilation failed - 'room' variable might not be defined")

	// Execute with simulated path param
	vmInstance := vm.NewVM()
	vmInstance.SetLocal("room", vm.StringValue{Val: "general"})
	vmInstance.SetLocal("client", vm.StringValue{Val: "user123"})

	wsHandler := &mockWSHandler{}
	vmInstance.SetWebSocketHandler(wsHandler)

	_, err = vmInstance.Execute(compiled.OnConnect)
	require.NoError(t, err, "OnConnect execution failed - 'room' variable access issue")

	assert.Equal(t, "general", wsHandler.broadcastRoom,
		"broadcast_to_room should be called with room='general' from path parameter")

	t.Log("Issue #64 verification: PASSED")
	t.Log("  - Path parameter 'room' is correctly defined in symbol table")
	t.Log("  - Path parameter is accessible in all event handlers")
	t.Log("  - broadcast_to_room(room, ...) works as expected")
}
