package parser

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// WebSocket Route Tests
// ============================================================================

// Test 1: WebSocket route with connect event
func TestParser_WebSocket_ConnectEvent(t *testing.T) {
	source := `@ ws /chat {
  on connect {
    $ clientId = connection.id
    > {status: "connected", id: clientId}
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	wsRoute, ok := module.Items[0].(*interpreter.WebSocketRoute)
	require.True(t, ok, "Expected WebSocketRoute, got %T", module.Items[0])

	assert.Equal(t, "/chat", wsRoute.Path)
	require.Len(t, wsRoute.Events, 1)
	assert.Equal(t, interpreter.WSEventConnect, wsRoute.Events[0].EventType)
	assert.Greater(t, len(wsRoute.Events[0].Body), 0)
}

// Test 2: WebSocket route with message event
func TestParser_WebSocket_MessageEvent(t *testing.T) {
	source := `@ ws /messages {
  on message {
    $ content = message.data
    > {received: content}
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	wsRoute, ok := module.Items[0].(*interpreter.WebSocketRoute)
	require.True(t, ok)

	assert.Equal(t, "/messages", wsRoute.Path)
	require.Len(t, wsRoute.Events, 1)
	assert.Equal(t, interpreter.WSEventMessage, wsRoute.Events[0].EventType)
}

// Test 3: WebSocket route with disconnect event
func TestParser_WebSocket_DisconnectEvent(t *testing.T) {
	source := `@ ws /session {
  on disconnect {
    $ userId = session.userId
    > {disconnected: userId}
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	wsRoute, ok := module.Items[0].(*interpreter.WebSocketRoute)
	require.True(t, ok)

	require.Len(t, wsRoute.Events, 1)
	assert.Equal(t, interpreter.WSEventDisconnect, wsRoute.Events[0].EventType)
}

// Test 4: WebSocket route with error event
func TestParser_WebSocket_ErrorEvent(t *testing.T) {
	source := `@ ws /stream {
  on error {
    $ errorMsg = error.message
    > {error: errorMsg, code: 500}
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	wsRoute, ok := module.Items[0].(*interpreter.WebSocketRoute)
	require.True(t, ok)

	require.Len(t, wsRoute.Events, 1)
	assert.Equal(t, interpreter.WSEventError, wsRoute.Events[0].EventType)
}

// Test 5: WebSocket route with all event types
func TestParser_WebSocket_AllEventTypes(t *testing.T) {
	source := `@ ws /realtime {
  on connect {
    > {event: "connect"}
  }
  on message {
    > {event: "message"}
  }
  on disconnect {
    > {event: "disconnect"}
  }
  on error {
    > {event: "error"}
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	wsRoute, ok := module.Items[0].(*interpreter.WebSocketRoute)
	require.True(t, ok)

	assert.Equal(t, "/realtime", wsRoute.Path)
	require.Len(t, wsRoute.Events, 4)

	// Verify all event types are present
	eventTypes := make(map[interpreter.WebSocketEventType]bool)
	for _, event := range wsRoute.Events {
		eventTypes[event.EventType] = true
	}

	assert.True(t, eventTypes[interpreter.WSEventConnect], "missing connect event")
	assert.True(t, eventTypes[interpreter.WSEventMessage], "missing message event")
	assert.True(t, eventTypes[interpreter.WSEventDisconnect], "missing disconnect event")
	assert.True(t, eventTypes[interpreter.WSEventError], "missing error event")
}

// Test 6: WebSocket route with path parameter
func TestParser_WebSocket_PathParameter(t *testing.T) {
	source := `@ ws /chat/:room {
  on connect {
    $ roomId = room
    > {joined: roomId}
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	wsRoute, ok := module.Items[0].(*interpreter.WebSocketRoute)
	require.True(t, ok)

	assert.Equal(t, "/chat/:room", wsRoute.Path)
}

// Test 7: WebSocket route with multiple path parameters
func TestParser_WebSocket_MultiplePathParameters(t *testing.T) {
	source := `@ ws /chat/:room/:userId {
  on message {
    $ data = message.data
    > {room: room, user: userId, data: data}
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	wsRoute, ok := module.Items[0].(*interpreter.WebSocketRoute)
	require.True(t, ok)

	assert.Equal(t, "/chat/:room/:userId", wsRoute.Path)
}

// Test 8: WebSocket route using 'websocket' keyword
func TestParser_WebSocket_WebsocketKeyword(t *testing.T) {
	source := `@ websocket /notifications {
  on message {
    > {ok: true}
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	wsRoute, ok := module.Items[0].(*interpreter.WebSocketRoute)
	require.True(t, ok)

	assert.Equal(t, "/notifications", wsRoute.Path)
}

// ============================================================================
// Generic Type Parsing Tests
// ============================================================================

// Test 9: Generic type List[str]
func TestParser_GenericType_ListOfString(t *testing.T) {
	source := `: UserList {
  names: List[str]!
  count: int!
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)

	assert.Equal(t, "UserList", typeDef.Name)
	require.Len(t, typeDef.Fields, 2)

	// First field should be List[str]
	assert.Equal(t, "names", typeDef.Fields[0].Name)
	assert.True(t, typeDef.Fields[0].Required)
	// The type should be parsed as GenericType with List base and str type argument
	genericType, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.GenericType)
	require.True(t, ok, "Expected GenericType, got %T", typeDef.Fields[0].TypeAnnotation)
	namedType, ok := genericType.BaseType.(interpreter.NamedType)
	require.True(t, ok)
	assert.Equal(t, "List", namedType.Name)
	require.Len(t, genericType.TypeArgs, 1)
}

// Test 10: Generic type Map[str, int]
func TestParser_GenericType_MapOfStringInt(t *testing.T) {
	source := `: Scores {
  playerScores: Map[str, int]!
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)

	assert.Equal(t, "Scores", typeDef.Name)
	require.Len(t, typeDef.Fields, 1)

	assert.Equal(t, "playerScores", typeDef.Fields[0].Name)
	genericType, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.GenericType)
	require.True(t, ok, "Expected GenericType, got %T", typeDef.Fields[0].TypeAnnotation)
	namedType, ok := genericType.BaseType.(interpreter.NamedType)
	require.True(t, ok)
	assert.Equal(t, "Map", namedType.Name)
	require.Len(t, genericType.TypeArgs, 2)
}

// Test 11: Nested generic types
func TestParser_GenericType_NestedGenerics(t *testing.T) {
	source := `: NestedData {
  matrix: List[List[int]]!
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)

	assert.Equal(t, "NestedData", typeDef.Name)
	require.Len(t, typeDef.Fields, 1)
	assert.Equal(t, "matrix", typeDef.Fields[0].Name)
}

// ============================================================================
// Union Type Parsing Tests
// ============================================================================

// Test 12: Simple union type
func TestParser_UnionType_Simple(t *testing.T) {
	source := `: Response {
  result: User | Error!
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)

	require.Len(t, typeDef.Fields, 1)
	assert.Equal(t, "result", typeDef.Fields[0].Name)

	// Check union type
	unionType, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.UnionType)
	require.True(t, ok, "Expected UnionType, got %T", typeDef.Fields[0].TypeAnnotation)
	require.Len(t, unionType.Types, 2)

	// First type should be User
	userType, ok := unionType.Types[0].(interpreter.NamedType)
	require.True(t, ok)
	assert.Equal(t, "User", userType.Name)

	// Second type should be Error
	errorType, ok := unionType.Types[1].(interpreter.NamedType)
	require.True(t, ok)
	assert.Equal(t, "Error", errorType.Name)
}

// Test 13: Union type with multiple types
func TestParser_UnionType_Multiple(t *testing.T) {
	source := `: ApiResponse {
  data: Success | PartialSuccess | Failure!
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)

	unionType, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.UnionType)
	require.True(t, ok)
	require.Len(t, unionType.Types, 3)
}

// Test 14: Union type with primitive types
func TestParser_UnionType_PrimitiveTypes(t *testing.T) {
	source := `: FlexibleValue {
  value: int | str | bool!
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)

	unionType, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.UnionType)
	require.True(t, ok)
	require.Len(t, unionType.Types, 3)

	// Check primitive types
	_, isInt := unionType.Types[0].(interpreter.IntType)
	assert.True(t, isInt, "First type should be int")

	_, isString := unionType.Types[1].(interpreter.StringType)
	assert.True(t, isString, "Second type should be str")

	_, isBool := unionType.Types[2].(interpreter.BoolType)
	assert.True(t, isBool, "Third type should be bool")
}

// ============================================================================
// HTTP Method Shorthand Tests
// ============================================================================

// Test 15: HTTP GET shorthand
func TestParser_HTTPMethod_GETShorthand(t *testing.T) {
	source := `@ GET /users {
  > {users: []}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	assert.Equal(t, "/users", route.Path)
	assert.Equal(t, interpreter.Get, route.Method)
}

// Test 16: HTTP POST shorthand
func TestParser_HTTPMethod_POSTShorthand(t *testing.T) {
	source := `@ POST /users {
  $ user = input
  > {created: user}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	assert.Equal(t, "/users", route.Path)
	assert.Equal(t, interpreter.Post, route.Method)
}

// Test 17: HTTP PUT shorthand
func TestParser_HTTPMethod_PUTShorthand(t *testing.T) {
	source := `@ PUT /users/:id {
  $ updated = input
  > {updated: updated}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	assert.Equal(t, "/users/:id", route.Path)
	assert.Equal(t, interpreter.Put, route.Method)
}

// Test 18: HTTP DELETE shorthand
func TestParser_HTTPMethod_DELETEShorthand(t *testing.T) {
	source := `@ DELETE /users/:id {
  > {deleted: true}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	assert.Equal(t, "/users/:id", route.Path)
	assert.Equal(t, interpreter.Delete, route.Method)
}

// Test 19: HTTP PATCH shorthand
func TestParser_HTTPMethod_PATCHShorthand(t *testing.T) {
	source := `@ PATCH /users/:id {
  $ partial = input
  > {patched: partial}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	assert.Equal(t, "/users/:id", route.Path)
	assert.Equal(t, interpreter.Patch, route.Method)
}

// ============================================================================
// Complex Combined Tests
// ============================================================================

// Test 20: WebSocket with complex event handlers
func TestParser_WebSocket_ComplexEventHandlers(t *testing.T) {
	source := `@ ws /game/:gameId {
  on connect {
    $ playerId = connection.id
    $ game = gameId
    if game == "lobby" {
      > {status: "joined_lobby"}
    } else {
      > {status: "joined_game", gameId: game}
    }
  }
  on message {
    $ action = message.action
    $ data = message.data
    > {processed: action, result: data}
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	wsRoute, ok := module.Items[0].(*interpreter.WebSocketRoute)
	require.True(t, ok)

	assert.Equal(t, "/game/:gameId", wsRoute.Path)
	require.Len(t, wsRoute.Events, 2)

	// Connect event should have if statement
	connectEvent := wsRoute.Events[0]
	assert.Equal(t, interpreter.WSEventConnect, connectEvent.EventType)
	assert.GreaterOrEqual(t, len(connectEvent.Body), 3) // assignments + if statement
}

// Test 21: Route with return type using union
func TestParser_Route_WithUnionReturnType(t *testing.T) {
	source := `@ GET /api/user/:id -> User | Error {
  $ user = db.findUser(id)
  if user == null {
    > {error: "Not found", code: 404}
  } else {
    > user
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	route, ok := module.Items[0].(*interpreter.Route)
	require.True(t, ok)

	assert.Equal(t, "/api/user/:id", route.Path)
	assert.Equal(t, interpreter.Get, route.Method)

	// Check return type is union
	unionType, ok := route.ReturnType.(interpreter.UnionType)
	require.True(t, ok, "Expected UnionType return type, got %T", route.ReturnType)
	require.Len(t, unionType.Types, 2)
}

// Test 22: Type definition with mixed generic and union types
func TestParser_TypeDef_MixedGenericAndUnion(t *testing.T) {
	source := `: ApiResult {
  data: List[User] | Error!
  metadata: Map[str, str]
  pagination: Pagination | int
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)

	assert.Equal(t, "ApiResult", typeDef.Name)
	require.Len(t, typeDef.Fields, 3)

	// First field: List[User] | Error (union with generic)
	assert.Equal(t, "data", typeDef.Fields[0].Name)
	assert.True(t, typeDef.Fields[0].Required)

	// Second field: Map[str, str] (generic)
	assert.Equal(t, "metadata", typeDef.Fields[1].Name)
	genericType, ok := typeDef.Fields[1].TypeAnnotation.(interpreter.GenericType)
	require.True(t, ok, "Expected GenericType, got %T", typeDef.Fields[1].TypeAnnotation)
	namedType, ok := genericType.BaseType.(interpreter.NamedType)
	require.True(t, ok)
	assert.Equal(t, "Map", namedType.Name)
	require.Len(t, genericType.TypeArgs, 2)

	// Third field: Pagination | int (union with named and primitive)
	assert.Equal(t, "pagination", typeDef.Fields[2].Name)
}

// Test 23: Multiple routes with different HTTP methods
func TestParser_MultipleRoutes_AllMethods(t *testing.T) {
	source := `@ GET /items {
  > {items: []}
}

@ POST /items {
  > {created: true}
}

@ PUT /items/:id {
  > {updated: true}
}

@ DELETE /items/:id {
  > {deleted: true}
}

@ PATCH /items/:id {
  > {patched: true}
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 5)

	methods := []interpreter.HttpMethod{
		interpreter.Get,
		interpreter.Post,
		interpreter.Put,
		interpreter.Delete,
		interpreter.Patch,
	}

	for i, expectedMethod := range methods {
		route, ok := module.Items[i].(*interpreter.Route)
		require.True(t, ok, "Item %d should be Route", i)
		assert.Equal(t, expectedMethod, route.Method, "Route %d method mismatch", i)
	}
}

// Test 24: WebSocket error case - unknown event type
func TestParser_WebSocket_UnknownEventType(t *testing.T) {
	source := `@ ws /test {
  on unknownevent {
    > {ok: true}
  }
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	_, err = parser.Parse()
	assert.Error(t, err, "Should error on unknown event type")
	assert.Contains(t, err.Error(), "Unknown WebSocket event")
}

// Test 25: Array type with empty brackets
func TestParser_ArrayType_EmptyBrackets(t *testing.T) {
	source := `: ArrayData {
  numbers: int[]
  strings: str[]!
}`

	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)

	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	require.Len(t, module.Items, 1)
	typeDef, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)

	assert.Equal(t, "ArrayData", typeDef.Name)
	require.Len(t, typeDef.Fields, 2)

	// First field: int[]
	arrayType1, ok := typeDef.Fields[0].TypeAnnotation.(interpreter.ArrayType)
	require.True(t, ok, "Expected ArrayType, got %T", typeDef.Fields[0].TypeAnnotation)
	_, isInt := arrayType1.ElementType.(interpreter.IntType)
	assert.True(t, isInt, "Element type should be int")

	// Second field: str[]!
	arrayType2, ok := typeDef.Fields[1].TypeAnnotation.(interpreter.ArrayType)
	require.True(t, ok)
	_, isString := arrayType2.ElementType.(interpreter.StringType)
	assert.True(t, isString, "Element type should be str")
	assert.True(t, typeDef.Fields[1].Required)
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkParser_WebSocketRoute(b *testing.B) {
	source := `@ ws /chat/:room {
  on connect {
    $ clientId = connection.id
    > {connected: true}
  }
  on message {
    $ data = message.data
    > {received: data}
  }
  on disconnect {
    > {disconnected: true}
  }
}`

	for i := 0; i < b.N; i++ {
		lexer := NewLexer(source)
		tokens, _ := lexer.Tokenize()
		parser := NewParser(tokens)
		_, _ = parser.Parse()
	}
}

func BenchmarkParser_GenericTypes(b *testing.B) {
	source := `: ComplexType {
  users: List[User]!
  scores: Map[str, int]!
  results: List[Map[str, User]]
}`

	for i := 0; i < b.N; i++ {
		lexer := NewLexer(source)
		tokens, _ := lexer.Tokenize()
		parser := NewParser(tokens)
		_, _ = parser.Parse()
	}
}

func BenchmarkParser_UnionTypes(b *testing.B) {
	source := `: Result {
  data: Success | Failure | Pending!
  code: int | str
}`

	for i := 0; i < b.N; i++ {
		lexer := NewLexer(source)
		tokens, _ := lexer.Tokenize()
		parser := NewParser(tokens)
		_, _ = parser.Parse()
	}
}
