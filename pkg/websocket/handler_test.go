package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// waitForCondition polls until condition returns true or timeout
func waitForCondition(condition func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func TestNewHandler(t *testing.T) {
	handler := NewHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.messageHandlers)
	assert.NotNil(t, handler.eventHandlers)
}

func TestHandlerOnMessage(t *testing.T) {
	handler := NewHandler()

	called := false
	handler.On(MessageTypeText, func(ctx *MessageContext) error {
		called = true
		return nil
	})

	assert.True(t, handler.HasHandler(MessageTypeText))

	hub := NewHub()
	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	msg := NewTextMessage("hello")
	ctx := &MessageContext{
		Conn:    conn,
		Message: msg,
	}

	err := handler.HandleMessage(ctx)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestHandlerOnEvent(t *testing.T) {
	handler := NewHandler()

	called := false
	var receivedData interface{}

	handler.OnEvent("custom_event", func(ctx *MessageContext) error {
		called = true
		receivedData = ctx.Message.Data
		return nil
	})

	assert.True(t, handler.HasEventHandler("custom_event"))

	hub := NewHub()
	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	msg := NewEventMessage("custom_event", map[string]interface{}{"test": "data"})
	ctx := &MessageContext{
		Conn:    conn,
		Message: msg,
	}

	err := handler.HandleMessage(ctx)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.NotNil(t, receivedData)
}

func TestHandlerJoinRoomHandler(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	handler := hub.handler

	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	msg := &Message{
		Type: MessageTypeJoinRoom,
		Room: "test-room",
	}

	ctx := &MessageContext{
		Conn:    conn,
		Message: msg,
	}

	err := handler.HandleMessage(ctx)
	assert.NoError(t, err)

	// Wait for room join with polling
	require.True(t, waitForCondition(func() bool {
		return conn.IsInRoom("test-room")
	}, 2*time.Second), "connection did not join room")
}

func TestHandlerLeaveRoomHandler(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	handler := hub.handler

	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	// Join room first
	conn.JoinRoom("test-room")
	require.True(t, waitForCondition(func() bool {
		return conn.IsInRoom("test-room")
	}, 2*time.Second), "connection did not join room")

	msg := &Message{
		Type: MessageTypeLeaveRoom,
		Room: "test-room",
	}

	ctx := &MessageContext{
		Conn:    conn,
		Message: msg,
	}

	err := handler.HandleMessage(ctx)
	assert.NoError(t, err)

	// Wait for room leave with polling
	require.True(t, waitForCondition(func() bool {
		return !conn.IsInRoom("test-room")
	}, 2*time.Second), "connection did not leave room")
}

func TestHandlerPingPongHandler(t *testing.T) {
	handler := NewHandler()

	hub := NewHub()
	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	msg := &Message{
		Type:      MessageTypePing,
		Timestamp: time.Now(),
	}

	ctx := &MessageContext{
		Conn:    conn,
		Message: msg,
	}

	err := handler.HandleMessage(ctx)
	assert.NoError(t, err)

	// Should receive pong response
	select {
	case response := <-conn.send:
		var respMsg Message
		err := json.Unmarshal(response, &respMsg)
		require.NoError(t, err)
		assert.Equal(t, MessageTypePong, respMsg.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("No pong response received")
	}
}

func TestHandlerRemoveHandler(t *testing.T) {
	handler := NewHandler()

	handler.On(MessageTypeText, func(ctx *MessageContext) error {
		return nil
	})

	assert.True(t, handler.HasHandler(MessageTypeText))

	handler.RemoveHandler(MessageTypeText)
	assert.False(t, handler.HasHandler(MessageTypeText))
}

func TestHandlerRemoveEventHandler(t *testing.T) {
	handler := NewHandler()

	handler.OnEvent("test_event", func(ctx *MessageContext) error {
		return nil
	})

	assert.True(t, handler.HasEventHandler("test_event"))

	handler.RemoveEventHandler("test_event")
	assert.False(t, handler.HasEventHandler("test_event"))
}

func TestHandlerGetMessageTypes(t *testing.T) {
	handler := NewHandler()

	handler.On(MessageTypeText, func(ctx *MessageContext) error {
		return nil
	})
	handler.On(MessageTypeJSON, func(ctx *MessageContext) error {
		return nil
	})

	types := handler.GetMessageTypes()
	assert.Contains(t, types, MessageTypeText)
	assert.Contains(t, types, MessageTypeJSON)
}

func TestHandlerGetEvents(t *testing.T) {
	handler := NewHandler()

	handler.OnEvent("event1", func(ctx *MessageContext) error {
		return nil
	})
	handler.OnEvent("event2", func(ctx *MessageContext) error {
		return nil
	})

	events := handler.GetEvents()
	assert.Contains(t, events, "event1")
	assert.Contains(t, events, "event2")
}

func TestHandlerClear(t *testing.T) {
	handler := NewHandler()

	handler.On(MessageTypeText, func(ctx *MessageContext) error {
		return nil
	})
	handler.OnEvent("test_event", func(ctx *MessageContext) error {
		return nil
	})

	handler.Clear()

	// Custom handlers should be cleared
	assert.False(t, handler.HasHandler(MessageTypeText))
	assert.False(t, handler.HasEventHandler("test_event"))

	// Default handlers should still exist
	assert.True(t, handler.HasHandler(MessageTypeJoinRoom))
	assert.True(t, handler.HasHandler(MessageTypeLeaveRoom))
}

func TestMessageContextReply(t *testing.T) {
	hub := NewHub()
	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	msg := NewTextMessage("hello")
	ctx := &MessageContext{
		Conn:    conn,
		Message: msg,
	}

	err := ctx.Reply(MessageTypeText, "reply message")
	assert.NoError(t, err)

	select {
	case response := <-conn.send:
		var respMsg Message
		err := json.Unmarshal(response, &respMsg)
		require.NoError(t, err)
		assert.Equal(t, MessageTypeText, respMsg.Type)
		assert.Equal(t, "test-conn", respMsg.Target)
	case <-time.After(1 * time.Second):
		t.Fatal("No reply received")
	}
}

func TestMessageContextReplyJSON(t *testing.T) {
	hub := NewHub()
	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	msg := NewTextMessage("hello")
	ctx := &MessageContext{
		Conn:    conn,
		Message: msg,
	}

	data := map[string]interface{}{"status": "ok"}
	err := ctx.ReplyJSON(data)
	assert.NoError(t, err)

	select {
	case response := <-conn.send:
		assert.Contains(t, string(response), "status")
		assert.Contains(t, string(response), "ok")
	case <-time.After(1 * time.Second):
		t.Fatal("No reply received")
	}
}

func TestMessageContextReplyError(t *testing.T) {
	hub := NewHub()
	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	msg := NewTextMessage("hello")
	ctx := &MessageContext{
		Conn:    conn,
		Message: msg,
	}

	err := ctx.ReplyError(assert.AnError)
	assert.NoError(t, err)

	select {
	case response := <-conn.send:
		var respMsg Message
		err := json.Unmarshal(response, &respMsg)
		require.NoError(t, err)
		assert.Equal(t, MessageTypeError, respMsg.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("No reply received")
	}
}
