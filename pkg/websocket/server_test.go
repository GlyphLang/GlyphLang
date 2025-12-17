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

func TestNewServer(t *testing.T) {
	server := NewServer()
	assert.NotNil(t, server)
	assert.NotNil(t, server.hub)
	defer server.Shutdown()
}

func TestServerWebSocketUpgrade(t *testing.T) {
	server := NewServer()
	defer server.Shutdown()

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
	time.Sleep(100 * time.Millisecond)

	// Check connection count
	assert.Equal(t, 1, server.GetHub().GetConnectionCount())
}

func TestHubConnectionRegistration(t *testing.T) {
	hub := NewHub()
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
	time.Sleep(100 * time.Millisecond)

	// Check if connection is registered
	assert.Equal(t, 1, hub.GetConnectionCount())
	assert.Contains(t, hub.connections, mockConn)
}

func TestHubConnectionUnregistration(t *testing.T) {
	hub := NewHub()
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
	time.Sleep(100 * time.Millisecond)

	hub.unregister <- mockConn
	time.Sleep(100 * time.Millisecond)

	// Check if connection is unregistered
	assert.Equal(t, 0, hub.GetConnectionCount())
	assert.NotContains(t, hub.connections, mockConn)
}

func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
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
	time.Sleep(100 * time.Millisecond)

	// Broadcast message
	testMsg := []byte("test broadcast")
	hub.Broadcast(testMsg)

	// Wait for broadcast
	time.Sleep(100 * time.Millisecond)

	// Check both connections received the message
	select {
	case msg := <-conn1.send:
		assert.Equal(t, testMsg, msg)
	default:
		t.Fatal("conn1 did not receive broadcast")
	}

	select {
	case msg := <-conn2.send:
		assert.Equal(t, testMsg, msg)
	default:
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
	time.Sleep(100 * time.Millisecond)

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
	go hub.Run()
	defer hub.Shutdown()

	handlerCalled := false
	hub.OnConnect(func(conn *Connection) error {
		handlerCalled = true
		assert.Equal(t, "test-conn", conn.ID)
		return nil
	})

	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	hub.register <- conn
	time.Sleep(100 * time.Millisecond)

	assert.True(t, handlerCalled)
}

func TestHubOnDisconnectHandler(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	handlerCalled := false
	hub.OnDisconnect(func(conn *Connection) error {
		handlerCalled = true
		assert.Equal(t, "test-conn", conn.ID)
		return nil
	})

	conn := &Connection{
		ID:   "test-conn",
		send: make(chan []byte, 256),
		hub:  hub,
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	hub.register <- conn
	time.Sleep(100 * time.Millisecond)

	hub.unregister <- conn
	time.Sleep(100 * time.Millisecond)

	assert.True(t, handlerCalled)
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
	time.Sleep(100 * time.Millisecond)

	data := map[string]interface{}{
		"type":    "notification",
		"message": "test",
	}

	err := hub.BroadcastJSON(data)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	select {
	case msg := <-conn.send:
		assert.Contains(t, string(msg), "notification")
		assert.Contains(t, string(msg), "test")
	case <-time.After(1 * time.Second):
		t.Fatal("No message received")
	}
}
