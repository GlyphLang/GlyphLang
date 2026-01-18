package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// waitChan waits for a channel to close or times out
func waitChan(ch <-chan struct{}, timeout time.Duration) bool {
	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		return false
	}
}

// pollCondition polls until condition returns true or timeout
func pollCondition(condition func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// TestFullWebSocketFlow tests complete WebSocket communication flow
func TestFullWebSocketFlow(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	connected := make(chan struct{})
	server.OnConnect(func(conn *Connection) error {
		close(connected)
		return nil
	})

	// Setup test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	// Connect client
	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client.Close()

	// Wait for connection
	require.True(t, waitChan(connected, 2*time.Second), "connection timed out")

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

	var connCount int32
	allConnected := make(chan struct{})
	server.OnConnect(func(conn *Connection) error {
		if atomic.AddInt32(&connCount, 1) == 2 {
			close(allConnected)
		}
		return nil
	})

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

	require.True(t, waitChan(allConnected, 2*time.Second), "connections timed out")

	// Verify both clients connected
	assert.Equal(t, 2, server.GetHub().GetConnectionCount())

	// Broadcast message to all clients
	broadcastData := map[string]interface{}{
		"message": "Hello everyone!",
	}
	err = server.GetHub().BroadcastJSON(broadcastData)
	require.NoError(t, err)

	// Read from client1 (with timeout via SetReadDeadline)
	client1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg1, err := client1.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg1), "Hello everyone!")

	// Read from client2
	client2.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg2, err := client2.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg2), "Hello everyone!")
}

// TestRoomBasedMessaging tests room-based message routing
func TestRoomBasedMessaging(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	var connCount int32
	allConnected := make(chan struct{})
	server.OnConnect(func(conn *Connection) error {
		if atomic.AddInt32(&connCount, 1) == 2 {
			close(allConnected)
		}
		return nil
	})

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

	require.True(t, waitChan(allConnected, 2*time.Second), "connections timed out")

	// Client1 joins room1
	joinMsg := &Message{
		Type: MessageTypeJoinRoom,
		Room: "room1",
	}
	joinData, _ := joinMsg.ToJSON()
	err = client1.WriteMessage(websocket.TextMessage, joinData)
	require.NoError(t, err)

	// Read join confirmation with timeout
	client1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err = client1.ReadMessage()
	require.NoError(t, err)

	// Broadcast to room1
	roomData := map[string]interface{}{
		"message": "Room1 broadcast",
	}
	err = server.GetHub().BroadcastJSONToRoom("room1", roomData, nil)
	require.NoError(t, err)

	// Client1 should receive the message
	client1.SetReadDeadline(time.Now().Add(2 * time.Second))
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

	connected := make(chan struct{})
	eventHandled := make(chan struct{})
	var eventData interface{}
	var mu sync.Mutex

	server.OnConnect(func(conn *Connection) error {
		close(connected)
		return nil
	})

	// Register custom event handler
	server.OnEvent("custom_event", func(ctx *MessageContext) error {
		mu.Lock()
		eventData = ctx.Message.Data
		mu.Unlock()
		close(eventHandled)
		return nil
	})

	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client.Close()

	require.True(t, waitChan(connected, 2*time.Second), "connection timed out")

	// Send custom event
	msg := NewEventMessage("custom_event", map[string]interface{}{
		"key": "value",
	})
	msgData, _ := msg.ToJSON()
	err = client.WriteMessage(websocket.TextMessage, msgData)
	require.NoError(t, err)

	// Wait for event to be handled
	require.True(t, waitChan(eventHandled, 2*time.Second), "event handler not called")

	// Verify event data
	mu.Lock()
	assert.NotNil(t, eventData)
	mu.Unlock()
}

// TestConnectionLifecycleHandlers tests onConnect and onDisconnect handlers
func TestConnectionLifecycleHandlers(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	connected := make(chan struct{})
	disconnected := make(chan struct{})

	server.OnConnect(func(conn *Connection) error {
		close(connected)
		return nil
	})

	server.OnDisconnect(func(conn *Connection) error {
		close(disconnected)
		return nil
	})

	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	require.True(t, waitChan(connected, 2*time.Second), "OnConnect not called")

	client.Close()
	require.True(t, waitChan(disconnected, 2*time.Second), "OnDisconnect not called")
}

// TestPingPongMechanism tests ping/pong keep-alive
func TestPingPongMechanism(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	connected := make(chan struct{})
	server.OnConnect(func(conn *Connection) error {
		close(connected)
		return nil
	})

	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client.Close()

	require.True(t, waitChan(connected, 2*time.Second), "connection timed out")

	// Send ping
	pingMsg := &Message{
		Type:      MessageTypePing,
		Timestamp: time.Now(),
	}
	pingData, _ := pingMsg.ToJSON()
	err = client.WriteMessage(websocket.TextMessage, pingData)
	require.NoError(t, err)

	// Receive pong with timeout
	client.SetReadDeadline(time.Now().Add(2 * time.Second))
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

	connected := make(chan struct{})
	server.OnConnect(func(conn *Connection) error {
		close(connected)
		return nil
	})

	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer client.Close()

	require.True(t, waitChan(connected, 2*time.Second), "connection timed out")

	// Join room
	joinMsg := &Message{
		Type: MessageTypeJoinRoom,
		Room: "test-room",
	}
	joinData, _ := joinMsg.ToJSON()
	err = client.WriteMessage(websocket.TextMessage, joinData)
	require.NoError(t, err)

	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	client.ReadMessage() // Read join confirmation

	// Leave room
	leaveMsg := &Message{
		Type: MessageTypeLeaveRoom,
		Room: "test-room",
	}
	leaveData, _ := leaveMsg.ToJSON()
	err = client.WriteMessage(websocket.TextMessage, leaveData)
	require.NoError(t, err)

	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	client.ReadMessage() // Read leave confirmation

	// Wait for room to be empty
	require.True(t, pollCondition(func() bool {
		return server.GetHub().GetRoomManager().GetRoomSize("test-room") == 0
	}, 2*time.Second), "room not empty")
}

// TestConcurrentConnections tests handling many concurrent connections
func TestConcurrentConnections(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	numClients := 10
	var connCount int32
	allConnected := make(chan struct{})
	server.OnConnect(func(conn *Connection) error {
		if atomic.AddInt32(&connCount, 1) == int32(numClients) {
			close(allConnected)
		}
		return nil
	})

	ts := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	clients := make([]*websocket.Conn, numClients)

	// Connect multiple clients
	for i := 0; i < numClients; i++ {
		client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		clients[i] = client
		defer client.Close()
	}

	require.True(t, waitChan(allConnected, 5*time.Second), "not all clients connected")

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
