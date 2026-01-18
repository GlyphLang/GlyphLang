package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// waitWithTimeout waits for a channel to close or times out
func waitWithTimeout(ch <-chan struct{}, timeout time.Duration) bool {
	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		return false
	}
}

func TestNewServer(t *testing.T) {
	server := NewServer()
	assert.NotNil(t, server)
	assert.NotNil(t, server.hub)
	defer server.Shutdown()
}

func TestServerWebSocketUpgrade(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

	// Set up connection signal
	connected := make(chan struct{})
	server.OnConnect(func(conn *Connection) error {
		close(connected)
		return nil
	})

	// Create test HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// Connect client
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Wait for connection to register
	require.True(t, waitWithTimeout(connected, 2*time.Second), "connection registration timed out")

	// Check connection count
	assert.Equal(t, 1, server.GetHub().GetConnectionCount())
}

func TestHubConnectionRegistration(t *testing.T) {
	hub := NewHub()

	// Set up connection signal before starting hub
	connected := make(chan struct{})
	hub.OnConnect(func(conn *Connection) error {
		close(connected)
		return nil
	})

	go hub.Run()
	defer hub.Shutdown()

	// Create mock connection
	mockConn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	// Register connection
	hub.register <- mockConn

	// Wait for registration
	require.True(t, waitWithTimeout(connected, 2*time.Second), "connection registration timed out")

	// Check if connection is registered
	assert.Equal(t, 1, hub.GetConnectionCount())
	hub.connMu.RLock()
	assert.Contains(t, hub.connections, mockConn)
	hub.connMu.RUnlock()
}

func TestHubConnectionUnregistration(t *testing.T) {
	hub := NewHub()

	// Set up connection signals
	connected := make(chan struct{})
	disconnected := make(chan struct{})
	hub.OnConnect(func(conn *Connection) error {
		close(connected)
		return nil
	})
	hub.OnDisconnect(func(conn *Connection) error {
		close(disconnected)
		return nil
	})

	go hub.Run()
	defer hub.Shutdown()

	mockConn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	// Register then unregister
	hub.register <- mockConn
	require.True(t, waitWithTimeout(connected, 2*time.Second), "connection registration timed out")

	hub.unregister <- mockConn
	require.True(t, waitWithTimeout(disconnected, 2*time.Second), "connection unregistration timed out")

	// Check if connection is unregistered
	assert.Equal(t, 0, hub.GetConnectionCount())
	hub.connMu.RLock()
	assert.NotContains(t, hub.connections, mockConn)
	hub.connMu.RUnlock()
}

func TestHubBroadcast(t *testing.T) {
	hub := NewHub()

	// Track connections
	connCount := 0
	allConnected := make(chan struct{})
	hub.OnConnect(func(conn *Connection) error {
		connCount++
		if connCount == 2 {
			close(allConnected)
		}
		return nil
	})

	go hub.Run()
	defer hub.Shutdown()

	// Create two mock connections
	conn1 := &Connection{
		ID:   "conn1",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}
	conn2 := &Connection{
		ID:   "conn2",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	hub.register <- conn1
	hub.register <- conn2
	require.True(t, waitWithTimeout(allConnected, 2*time.Second), "connections registration timed out")

	// Broadcast message
	testMsg := []byte("test broadcast")
	hub.Broadcast(testMsg)

	// Check both connections received the message (with timeout)
	select {
	case msg := <-conn1.send:
		assert.Equal(t, testMsg, msg)
	case <-time.After(2 * time.Second):
		t.Fatal("conn1 did not receive broadcast")
	}

	select {
	case msg := <-conn2.send:
		assert.Equal(t, testMsg, msg)
	case <-time.After(2 * time.Second):
		t.Fatal("conn2 did not receive broadcast")
	}
}

func TestConnectionSetGetData(t *testing.T) {
	conn := &Connection{
		ID:   "test-conn",
		Data: make(map[string]interface{}),
	}

	// Set data
	conn.SetData("user_id", 123)
	conn.SetData("username", "testuser")

	// Get data
	userID, exists := conn.GetData("user_id")
	assert.True(t, exists)
	assert.Equal(t, 123, userID)

	username, exists := conn.GetData("username")
	assert.True(t, exists)
	assert.Equal(t, "testuser", username)

	// Get non-existent data
	_, exists = conn.GetData("nonexistent")
	assert.False(t, exists)
}

func TestConnectionSendJSON(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
	}

	data := map[string]interface{}{
		"message": "hello",
		"count":   42,
	}

	err := conn.SendJSON(data)
	require.NoError(t, err)

	// Verify message in channel
	select {
	case msg := <-conn.send:
		assert.Contains(t, string(msg), "hello")
		assert.Contains(t, string(msg), "42")
	case <-time.After(1 * time.Second):
		t.Fatal("No message received")
	}
}

func TestHubGetConnection(t *testing.T) {
	hub := NewHub()

	connected := make(chan struct{})
	hub.OnConnect(func(conn *Connection) error {
		close(connected)
		return nil
	})

	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:   "test-conn-123",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	hub.register <- conn
	require.True(t, waitWithTimeout(connected, 2*time.Second), "connection registration timed out")

	// Get existing connection
	found, exists := hub.GetConnection("test-conn-123")
	assert.True(t, exists)
	assert.Equal(t, conn, found)

	// Get non-existent connection
	_, exists = hub.GetConnection("nonexistent")
	assert.False(t, exists)
}

func TestHubOnConnectHandler(t *testing.T) {
	hub := NewHub()

	handlerCalled := make(chan struct{})
	hub.OnConnect(func(conn *Connection) error {
		assert.Equal(t, "test-conn", conn.ID)
		close(handlerCalled)
		return nil
	})

	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	hub.register <- conn
	require.True(t, waitWithTimeout(handlerCalled, 2*time.Second), "OnConnect handler not called")
}

func TestHubOnDisconnectHandler(t *testing.T) {
	hub := NewHub()

	connected := make(chan struct{})
	disconnectCalled := make(chan struct{})
	hub.OnConnect(func(conn *Connection) error {
		close(connected)
		return nil
	})
	hub.OnDisconnect(func(conn *Connection) error {
		assert.Equal(t, "test-conn", conn.ID)
		close(disconnectCalled)
		return nil
	})

	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	hub.register <- conn
	require.True(t, waitWithTimeout(connected, 2*time.Second), "connection registration timed out")

	hub.unregister <- conn
	require.True(t, waitWithTimeout(disconnectCalled, 2*time.Second), "OnDisconnect handler not called")
}

func TestGenerateConnectionID(t *testing.T) {
	id1 := generateConnectionID()
	id2 := generateConnectionID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestHubBroadcastJSON(t *testing.T) {
	hub := NewHub()

	connected := make(chan struct{})
	hub.OnConnect(func(conn *Connection) error {
		close(connected)
		return nil
	})

	go hub.Run()
	defer hub.Shutdown()

	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	hub.register <- conn
	require.True(t, waitWithTimeout(connected, 2*time.Second), "connection registration timed out")

	data := map[string]interface{}{
		"type":    "notification",
		"message": "test",
	}

	err := hub.BroadcastJSON(data)
	require.NoError(t, err)

	select {
	case msg := <-conn.send:
		assert.Contains(t, string(msg), "notification")
		assert.Contains(t, string(msg), "test")
	case <-time.After(2 * time.Second):
		t.Fatal("No message received")
	}
}
