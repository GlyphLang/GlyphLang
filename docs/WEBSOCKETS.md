# WebSockets

GlyphLang provides built-in WebSocket support for real-time APIs.


```glyph
# Basic WebSocket chat
@ ws /chat {
  on connect {
    ws.join("lobby")
    ws.broadcast("User joined the chat")
  }

  on message {
    ws.broadcast(input)
  }

  on disconnect {
    ws.broadcast("User left the chat")
    ws.leave("lobby")
  }
}

# WebSocket with room parameter
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
```
