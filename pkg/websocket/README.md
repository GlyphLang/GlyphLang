# WebSocket Support for Glyph

Production-ready WebSocket implementation for real-time bidirectional communication in Glyph.

## Features

- **Connection Management**: Thread-safe connection tracking and lifecycle management
- **Message Routing**: Event-based message handling with JSON and binary support
- **Room/Channel Support**: Group connections into rooms for targeted broadcasting
- **Broadcast Operations**: Send to all connections, specific rooms, or individual clients
- **Ping/Pong Keep-Alive**: Automatic connection health monitoring
- **Event Handlers**: Custom event handling for connect, disconnect, and messages
- **Integration Ready**: Seamlessly integrates with Glyph HTTP server

## Architecture

```
pkg/websocket/
├── connection.go      # Connection wrapper with send/receive pumps
├── server.go          # WebSocket server and hub for connection management
├── handler.go         # Message routing and event handling
├── message.go         # Message types and structures
├── room.go            # Room/channel management
└── *_test.go          # Comprehensive test suite (30+ tests)
```

## Quick Start

### Go Server Setup

```go
package main

import (
    "net/http"
    "github.com/glyphlang/glyph/pkg/websocket"
)

func main() {
    // Create WebSocket server
    wsServer := websocket.NewServer()

    // Register event handlers
    wsServer.OnConnect(func(conn *websocket.Connection) error {
        log.Printf("Client connected: %s", conn.ID)
        return nil
    })

    wsServer.OnMessage(websocket.MessageTypeText, func(ctx *websocket.MessageContext) error {
        log.Printf("Received: %v", ctx.Message.Data)
        return ctx.ReplyJSON(map[string]interface{}{
            "echo": ctx.Message.Data,
        })
    })

    // Setup HTTP server
    http.HandleFunc("/ws", wsServer.HandleWebSocket)
    http.ListenAndServe(":8080", nil)
}
```

### Glyph Language Syntax

```glyph
// WebSocket route
@ ws /chat {
  on connect {
    ws.send({ type: "welcome", message: "Hello!" })
  }

  on message {
    ws.broadcast_to_room("general", {
      type: "chat",
      text: input.text,
      user: input.user
    })
  }

  on disconnect {
    // Cleanup logic
  }
}
```

## Room Management

```go
// Join a room
conn.JoinRoom("game-lobby")

// Leave a room
conn.LeaveRoom("game-lobby")

// Broadcast to a room
hub.BroadcastJSONToRoom("game-lobby", data, excludeConn)

// Get room info
roomSize := hub.GetRoomManager().GetRoomSize("game-lobby")
```

## Message Types

- `MessageTypeText`: Plain text messages
- `MessageTypeJSON`: Structured JSON data
- `MessageTypeBinary`: Binary data
- `MessageTypeJoinRoom`: Join a room
- `MessageTypeLeaveRoom`: Leave a room
- `MessageTypeBroadcast`: Broadcast message
- `MessageTypePing`/`MessageTypePong`: Keep-alive

## Testing

The package includes 30+ comprehensive tests covering:

- Connection lifecycle (connect, disconnect, send, receive)
- Message routing and handling
- Room operations (join, leave, broadcast)
- Event handlers (onConnect, onDisconnect, onMessage)
- Integration tests with real WebSocket connections
- Concurrent connection handling
- Ping/pong mechanism

Run tests:
```bash
go test ./pkg/websocket/... -v
```

## Example: Chat Application

See `examples/websocket-chat/` for a complete chat application with:
- Real-time messaging
- Multiple chat rooms
- User presence
- HTML/JavaScript client
- Glyph backend

## Production Considerations

### Security
- Configure CORS appropriately in the upgrader
- Validate message origins
- Implement authentication via connection data
- Rate limit message sending

### Performance
- Configurable buffer sizes (currently 256 messages)
- Concurrent read/write pumps per connection
- Efficient room broadcasting
- Connection pooling

### Monitoring
- Connection count tracking
- Room size monitoring
- Message throughput metrics
- Error logging

## API Reference

### Server
- `NewServer()`: Create new WebSocket server
- `HandleWebSocket(w, r)`: HTTP upgrade handler
- `OnConnect(handler)`: Register connect handler
- `OnDisconnect(handler)`: Register disconnect handler
- `OnMessage(msgType, handler)`: Register message handler
- `OnEvent(event, handler)`: Register custom event handler
- `Shutdown()`: Graceful shutdown

### Connection
- `Send([]byte)`: Send raw bytes
- `SendJSON(v)`: Send JSON
- `JoinRoom(name)`: Join a room
- `LeaveRoom(name)`: Leave a room
- `SetData(key, value)`: Store connection metadata
- `GetData(key)`: Retrieve connection metadata
- `Close()`: Close connection

### Hub
- `Broadcast(message)`: Send to all connections
- `BroadcastJSON(v)`: Broadcast JSON to all
- `BroadcastToRoom(room, message, exclude)`: Send to room
- `GetConnectionCount()`: Get active connections
- `GetConnection(id)`: Get specific connection

### Room
- `Add(conn)`: Add connection to room
- `Remove(conn)`: Remove connection from room
- `Broadcast(message, exclude)`: Broadcast to room
- `Size()`: Get room size
- `SetMetadata(key, value)`: Set room metadata

## Integration with Glyph Server

The WebSocket server integrates seamlessly with the Glyph HTTP server:

```go
server := server.NewServer()
wsServer := server.GetWebSocketServer()

// Register WebSocket route
server.RegisterWebSocketRoute("/chat", wsServer.HandleWebSocket)
```

## License

Part of the Glyph project.
