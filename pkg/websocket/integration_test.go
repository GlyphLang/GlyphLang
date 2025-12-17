package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullWebSocketFlow tests complete WebSocket communication flow
func TestFullWebSocketFlow(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	// Setup test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	// Connect client
	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client.Close()

	// Wait for connection
	time.Sleep(100 * time.Millisecond)

	// Send message from client
	msg := NewTextMessage("Hello Server")
	msgData, err := msg.ToJSON()
	require.NoError(t, err)

	err = client.WriteMessage(websocket.TextMessage, msgData)
	require.NoError(t, err)

	// Verify connection is registered
	assert.Equal(t, 1, server.GetHub().GetConnectionCount())
}

// TestMultipleClientsAndBroadcast tests multiple clients and broadcasting
func TestMultipleClientsAndBroadcast(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	// Connect multiple clients
	client1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client1.Close()

	client2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client2.Close()

	time.Sleep(100 * time.Millisecond)

	// Verify both clients connected
	assert.Equal(t, 2, server.GetHub().GetConnectionCount())

	// Broadcast message to all clients
	broadcastData := map[string]interface{}{
		"message": "Hello everyone!",
	}
	err = server.GetHub().BroadcastJSON(broadcastData)
	require.NoError(t, err)

	// Both clients should receive the message
	time.Sleep(100 * time.Millisecond)

	// Read from client1
	_, msg1, err := client1.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg1), "Hello everyone!")

	// Read from client2
	_, msg2, err := client2.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg2), "Hello everyone!")
}

// TestRoomBasedMessaging tests room-based message routing
func TestRoomBasedMessaging(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	// Connect two clients
	client1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client1.Close()

	client2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client2.Close()

	time.Sleep(100 * time.Millisecond)

	// Client1 joins room1
	joinMsg := &Message{
		Type: MessageTypeJoinRoom,
		Room: "room1",
	}
	joinData, _ := joinMsg.ToJSON()
	err = client1.WriteMessage(websocket.TextMessage, joinData)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Read join confirmation
	_, _, err = client1.ReadMessage()
	require.NoError(t, err)

	// Broadcast to room1
	roomData := map[string]interface{}{
		"message": "Room1 broadcast",
	}
	err = server.GetHub().BroadcastJSONToRoom("room1", roomData, nil)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Client1 should receive the message
	_, msg, err := client1.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg), "Room1 broadcast")

	// Client2 should not receive the message (not in room)
	client2.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, _, err = client2.ReadMessage()
	assert.Error(t, err) // Timeout expected
}

// TestCustomEventHandling tests custom event handlers
func TestCustomEventHandling(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	eventReceived := false
	var eventData interface{}

	// Register custom event handler
	server.OnEvent("custom_event", func(ctx *MessageContext) error {
		eventReceived = true
		eventData = ctx.Message.Data
		return nil
	})

	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client.Close()

	time.Sleep(100 * time.Millisecond)

	// Send custom event
	msg := NewEventMessage("custom_event", map[string]interface{}{
		"key": "value",
	})
	msgData, _ := msg.ToJSON()
	err = client.WriteMessage(websocket.TextMessage, msgData)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Verify event was handled
	assert.True(t, eventReceived)
	assert.NotNil(t, eventData)
}

// TestConnectionLifecycleHandlers tests onConnect and onDisconnect handlers
func TestConnectionLifecycleHandlers(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	connectCalled := false
	disconnectCalled := false

	server.OnConnect(func(conn *Connection) error {
		connectCalled = true
		return nil
	})

	server.OnDisconnect(func(conn *Connection) error {
		disconnectCalled = true
		return nil
	})

	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.True(t, connectCalled)

	client.Close()
	time.Sleep(100 * time.Millisecond)
	assert.True(t, disconnectCalled)
}

// TestPingPongMechanism tests ping/pong keep-alive
func TestPingPongMechanism(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client.Close()

	time.Sleep(100 * time.Millisecond)

	// Send ping
	pingMsg := &Message{
		Type:      MessageTypePing,
		Timestamp: time.Now(),
	}
	pingData, _ := pingMsg.ToJSON()
	err = client.WriteMessage(websocket.TextMessage, pingData)
	require.NoError(t, err)

	// Receive pong
	_, pongData, err := client.ReadMessage()
	require.NoError(t, err)

	var pongMsg Message
	err = json.Unmarshal(pongData, &pongMsg)
	require.NoError(t, err)
	assert.Equal(t, MessageTypePong, pongMsg.Type)
}

// TestRoomLeaving tests leaving a room
func TestRoomLeaving(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client.Close()

	time.Sleep(100 * time.Millisecond)

	// Join room
	joinMsg := &Message{
		Type: MessageTypeJoinRoom,
		Room: "test-room",
	}
	joinData, _ := joinMsg.ToJSON()
	err = client.WriteMessage(websocket.TextMessage, joinData)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	client.ReadMessage() // Read join confirmation

	// Leave room
	leaveMsg := &Message{
		Type: MessageTypeLeaveRoom,
		Room: "test-room",
	}
	leaveData, _ := leaveMsg.ToJSON()
	err = client.WriteMessage(websocket.TextMessage, leaveData)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	client.ReadMessage() // Read leave confirmation

	// Verify room is empty
	assert.Equal(t, 0, server.GetHub().GetRoomManager().GetRoomSize("test-room"))
}

// TestConcurrentConnections tests handling many concurrent connections
func TestConcurrentConnections(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	numClients := 10
	clients := make([]*websocket.Conn, numClients)

	// Connect multiple clients
	for i := 0; i < numClients; i++ {
		client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		clients[i] = client
		defer client.Close()
	}

	time.Sleep(200 * time.Millisecond)

	// Verify all clients connected
	assert.Equal(t, numClients, server.GetHub().GetConnectionCount())
}

// TestMessageToJSON tests message serialization
func TestMessageToJSON(t *testing.T) {
	msg := NewTextMessage("test")
	msg.SetMetadata("key", "value")

	data, err := msg.ToJSON()
	require.NoError(t, err)

	var parsed Message
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, MessageTypeText, parsed.Type)
	assert.Equal(t, "test", parsed.Data)
	assert.Equal(t, "value", parsed.Metadata["key"])
}

// TestFromJSON tests message deserialization
func TestFromJSON(t *testing.T) {
	jsonData := `{
		"type": "text",
		"data": "hello",
		"metadata": {"key": "value"}
	}`

	msg, err := FromJSON([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, MessageTypeText, msg.Type)
	assert.Equal(t, "hello", msg.Data)
	val, exists := msg.GetMetadata("key")
	assert.True(t, exists)
	assert.Equal(t, "value", val)
}
