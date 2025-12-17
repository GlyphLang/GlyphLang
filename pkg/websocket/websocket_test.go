package websocket

import (
	"testing"
	"time"
)

// TestWebSocketPackage is a meta-test to verify all components
func TestWebSocketPackage(t *testing.T) {
	// This test ensures the package compiles and basic structures work
	server := NewServer()
	if server == nil {
		t.Fatal("Failed to create WebSocket server")
	}

	hub := server.GetHub()
	if hub == nil {
		t.Fatal("Hub not initialized")
	}

	if hub.GetConnectionCount() != 0 {
		t.Errorf("Expected 0 connections, got %d", hub.GetConnectionCount())
	}

	roomManager := hub.GetRoomManager()
	if roomManager == nil {
		t.Fatal("Room manager not initialized")
	}

	if roomManager.Count() != 0 {
		t.Errorf("Expected 0 rooms, got %d", roomManager.Count())
	}

	// Test config
	config := hub.GetConfig()
	if config == nil {
		t.Fatal("Config not initialized")
	}

	// Test metrics
	metrics := hub.GetMetrics()
	if metrics == nil {
		t.Fatal("Metrics not initialized")
	}

	server.Shutdown()
}

// TestMessageTypes verifies all message types are defined
func TestMessageTypes(t *testing.T) {
	types := []MessageType{
		MessageTypeText,
		MessageTypeBinary,
		MessageTypeJSON,
		MessageTypeConnect,
		MessageTypeDisconnect,
		MessageTypeError,
		MessageTypeJoinRoom,
		MessageTypeLeaveRoom,
		MessageTypeBroadcast,
		MessageTypePing,
		MessageTypePong,
	}

	for _, msgType := range types {
		msg := NewMessage(msgType, "test")
		if msg.Type != msgType {
			t.Errorf("Message type mismatch: expected %s, got %s", msgType, msg.Type)
		}
	}
}

// TestErrorTypes verifies all error types are defined
func TestErrorTypes(t *testing.T) {
	errors := []error{
		ErrConnectionClosed,
		ErrInvalidMessage,
		ErrRoomNotFound,
		ErrConnectionNotFound,
		ErrRoomFull,
	}

	for _, err := range errors {
		if err == nil {
			t.Error("Error should not be nil")
		}
	}
}

// TestConfig tests configuration functionality
func TestConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()
		if config == nil {
			t.Fatal("DefaultConfig returned nil")
		}

		// Verify default values
		if config.EnableHeartbeat != true {
			t.Errorf("Expected EnableHeartbeat to be true, got %v", config.EnableHeartbeat)
		}

		if config.EnableReconnection != true {
			t.Errorf("Expected EnableReconnection to be true, got %v", config.EnableReconnection)
		}

		if config.EnableMetrics != true {
			t.Errorf("Expected EnableMetrics to be true, got %v", config.EnableMetrics)
		}

		if config.MessageQueueSize != 256 {
			t.Errorf("Expected MessageQueueSize to be 256, got %d", config.MessageQueueSize)
		}

		if config.MessageQueueStrategy != QueueStrategyDropOldest {
			t.Errorf("Expected MessageQueueStrategy to be DropOldest, got %v", config.MessageQueueStrategy)
		}
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		config := &Config{}
		err := config.Validate()
		if err != nil {
			t.Errorf("Validate returned error: %v", err)
		}

		// Check that validation sets defaults
		if config.HeartbeatInterval <= 0 {
			t.Error("HeartbeatInterval should be set to default")
		}

		if config.MessageQueueSize <= 0 {
			t.Error("MessageQueueSize should be set to default")
		}
	})

	t.Run("CustomConfig", func(t *testing.T) {
		config := &Config{
			MaxConnectionsPerHub:  100,
			MaxConnectionsPerRoom: 50,
			EnableHeartbeat:       true,
			HeartbeatInterval:     10 * time.Second,
			MessageQueueSize:      512,
			MessageQueueStrategy:  QueueStrategyBlock,
		}

		hub := NewHubWithConfig(config)
		if hub == nil {
			t.Fatal("Failed to create hub with custom config")
		}

		if hub.config.MaxConnectionsPerHub != 100 {
			t.Errorf("Expected MaxConnectionsPerHub to be 100, got %d", hub.config.MaxConnectionsPerHub)
		}

		hub.Shutdown()
	})
}

// TestMetrics tests metrics functionality
func TestMetrics(t *testing.T) {
	t.Run("BasicMetrics", func(t *testing.T) {
		metrics := NewMetrics()
		if metrics == nil {
			t.Fatal("NewMetrics returned nil")
		}

		// Test connection metrics
		metrics.IncrementConnections()
		if metrics.GetActiveConnections() != 1 {
			t.Errorf("Expected 1 active connection, got %d", metrics.GetActiveConnections())
		}

		metrics.DecrementConnections()
		if metrics.GetActiveConnections() != 0 {
			t.Errorf("Expected 0 active connections, got %d", metrics.GetActiveConnections())
		}

		// Test message metrics
		metrics.IncrementMessagesSent(100)
		if metrics.GetMessagesSent() != 1 {
			t.Errorf("Expected 1 message sent, got %d", metrics.GetMessagesSent())
		}

		if metrics.GetBytesSent() != 100 {
			t.Errorf("Expected 100 bytes sent, got %d", metrics.GetBytesSent())
		}

		metrics.IncrementMessagesReceived(200)
		if metrics.GetMessagesReceived() != 1 {
			t.Errorf("Expected 1 message received, got %d", metrics.GetMessagesReceived())
		}

		if metrics.GetBytesReceived() != 200 {
			t.Errorf("Expected 200 bytes received, got %d", metrics.GetBytesReceived())
		}
	})

	t.Run("ErrorMetrics", func(t *testing.T) {
		metrics := NewMetrics()

		metrics.IncrementReadErrors()
		metrics.IncrementWriteErrors()
		metrics.IncrementHandlerErrors()

		if metrics.GetTotalErrors() != 3 {
			t.Errorf("Expected 3 total errors, got %d", metrics.GetTotalErrors())
		}
	})

	t.Run("HeartbeatMetrics", func(t *testing.T) {
		metrics := NewMetrics()

		metrics.IncrementMissedPongs()
		metrics.IncrementSuccessfulPongs()

		if metrics.GetMissedPongs() != 1 {
			t.Errorf("Expected 1 missed pong, got %d", metrics.GetMissedPongs())
		}

		if metrics.GetSuccessfulPongs() != 1 {
			t.Errorf("Expected 1 successful pong, got %d", metrics.GetSuccessfulPongs())
		}
	})

	t.Run("QueueMetrics", func(t *testing.T) {
		metrics := NewMetrics()

		metrics.IncrementQueueOverflows()
		metrics.IncrementDroppedMessages()

		if metrics.GetQueueOverflows() != 1 {
			t.Errorf("Expected 1 queue overflow, got %d", metrics.GetQueueOverflows())
		}

		if metrics.GetDroppedMessages() != 1 {
			t.Errorf("Expected 1 dropped message, got %d", metrics.GetDroppedMessages())
		}
	})

	t.Run("ConnectionMetrics", func(t *testing.T) {
		metrics := NewMetrics()

		metrics.RegisterConnection("conn1")
		metrics.IncrementConnectionMessagesSent("conn1", 100)
		metrics.IncrementConnectionMessagesReceived("conn1", 200)
		metrics.IncrementConnectionErrors("conn1")
		metrics.IncrementConnectionMissedPongs("conn1")

		connMetrics := metrics.GetConnectionMetrics("conn1")
		if connMetrics == nil {
			t.Fatal("Expected connection metrics, got nil")
		}

		if connMetrics.MessagesSent != 1 {
			t.Errorf("Expected 1 message sent, got %d", connMetrics.MessagesSent)
		}

		if connMetrics.BytesSent != 100 {
			t.Errorf("Expected 100 bytes sent, got %d", connMetrics.BytesSent)
		}

		if connMetrics.MessagesReceived != 1 {
			t.Errorf("Expected 1 message received, got %d", connMetrics.MessagesReceived)
		}

		if connMetrics.BytesReceived != 200 {
			t.Errorf("Expected 200 bytes received, got %d", connMetrics.BytesReceived)
		}

		if connMetrics.Errors != 1 {
			t.Errorf("Expected 1 error, got %d", connMetrics.Errors)
		}

		if connMetrics.MissedPongs != 1 {
			t.Errorf("Expected 1 missed pong, got %d", connMetrics.MissedPongs)
		}

		metrics.UnregisterConnection("conn1")
		if metrics.GetConnectionMetrics("conn1") != nil {
			t.Error("Expected connection metrics to be removed")
		}
	})

	t.Run("Snapshot", func(t *testing.T) {
		metrics := NewMetrics()

		metrics.IncrementConnections()
		metrics.IncrementMessagesSent(100)
		metrics.IncrementMessagesReceived(200)

		snapshot := metrics.GetSnapshot()
		if snapshot == nil {
			t.Fatal("GetSnapshot returned nil")
		}

		if snapshot.ActiveConnections != 1 {
			t.Errorf("Expected 1 active connection, got %d", snapshot.ActiveConnections)
		}

		if snapshot.MessagesSent != 1 {
			t.Errorf("Expected 1 message sent, got %d", snapshot.MessagesSent)
		}

		if snapshot.MessagesReceived != 1 {
			t.Errorf("Expected 1 message received, got %d", snapshot.MessagesReceived)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		metrics := NewMetrics()

		metrics.IncrementConnections()
		metrics.IncrementMessagesSent(100)

		metrics.Reset()

		if metrics.GetActiveConnections() != 0 {
			t.Error("Expected metrics to be reset")
		}

		if metrics.GetMessagesSent() != 0 {
			t.Error("Expected metrics to be reset")
		}
	})

	t.Run("EnableDisable", func(t *testing.T) {
		metrics := NewMetrics()

		if !metrics.IsEnabled() {
			t.Error("Metrics should be enabled by default")
		}

		metrics.Disable()
		if metrics.IsEnabled() {
			t.Error("Metrics should be disabled")
		}

		// Operations should not affect metrics when disabled
		metrics.IncrementConnections()
		if metrics.GetActiveConnections() != 0 {
			t.Error("Metrics should not be updated when disabled")
		}

		metrics.Enable()
		if !metrics.IsEnabled() {
			t.Error("Metrics should be enabled")
		}
	})
}

// TestRoomWithLimits tests room connection limits
func TestRoomWithLimits(t *testing.T) {
	t.Run("MaxConnections", func(t *testing.T) {
		config := &Config{
			MaxConnectionsPerRoom: 2,
		}

		hub := NewHubWithConfig(config)
		go hub.Run()
		defer hub.Shutdown()

		// Create a room
		room := hub.roomManager.CreateRoom("test-room")

		// Create mock connections
		conn1 := &Connection{ID: "conn1", hub: hub, Data: make(map[string]interface{}), rooms: make(map[string]bool)}
		conn2 := &Connection{ID: "conn2", hub: hub, Data: make(map[string]interface{}), rooms: make(map[string]bool)}
		conn3 := &Connection{ID: "conn3", hub: hub, Data: make(map[string]interface{}), rooms: make(map[string]bool)}

		// Add connections up to the limit
		err := room.Add(conn1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		err = room.Add(conn2)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Try to add beyond the limit
		err = room.Add(conn3)
		if err != ErrRoomFull {
			t.Errorf("Expected ErrRoomFull, got %v", err)
		}

		if room.Size() != 2 {
			t.Errorf("Expected room size to be 2, got %d", room.Size())
		}
	})
}

// TestConnectionState tests connection state preservation
func TestConnectionState(t *testing.T) {
	t.Run("StatePreservation", func(t *testing.T) {
		config := &Config{
			EnableReconnection:  true,
			PreserveClientState: true,
			ReconnectionTimeout: 5 * time.Second,
		}

		hub := NewHubWithConfig(config)
		go hub.Run()
		defer hub.Shutdown()

		// Create a mock connection with state
		conn := &Connection{
			ID:    "conn1",
			hub:   hub,
			Data:  make(map[string]interface{}),
			rooms: make(map[string]bool),
		}
		conn.SetData("clientID", "client123")
		conn.SetData("username", "testuser")
		conn.rooms["room1"] = true

		// Save connection state
		hub.saveConnectionState(conn)

		// Verify state was saved
		hub.stateMu.RLock()
		state, exists := hub.connectionStates["client123"]
		hub.stateMu.RUnlock()

		if !exists {
			t.Fatal("Connection state was not saved")
		}

		if state.ClientID != "client123" {
			t.Errorf("Expected clientID to be 'client123', got %s", state.ClientID)
		}

		if len(state.Rooms) != 1 {
			t.Errorf("Expected 1 room, got %d", len(state.Rooms))
		}

		// Create a new connection and restore state
		newConn := &Connection{
			ID:    "conn2",
			hub:   hub,
			Data:  make(map[string]interface{}),
			rooms: make(map[string]bool),
		}

		restored := hub.RestoreConnectionState(newConn, "client123")
		if !restored {
			t.Error("Failed to restore connection state")
		}

		// Verify data was restored
		username, ok := newConn.GetData("username")
		if !ok || username != "testuser" {
			t.Error("Username was not restored correctly")
		}

		clientID, ok := newConn.GetData("clientID")
		if !ok || clientID != "client123" {
			t.Error("ClientID was not restored correctly")
		}
	})

	t.Run("StateExpiration", func(t *testing.T) {
		config := &Config{
			EnableReconnection:  true,
			PreserveClientState: true,
			ReconnectionTimeout: 100 * time.Millisecond,
			MaxReconnectionTime: 200 * time.Millisecond,
		}

		hub := NewHubWithConfig(config)
		go hub.Run()
		defer hub.Shutdown()

		// Save a connection state
		hub.stateMu.Lock()
		hub.connectionStates["client123"] = &ConnectionState{
			ClientID: "client123",
			LastSeen: time.Now().Add(-500 * time.Millisecond), // Old state
			Data:     make(map[string]interface{}),
			Rooms:    []string{},
		}
		hub.stateMu.Unlock()

		// Try to restore - should fail due to age
		newConn := &Connection{
			ID:    "conn2",
			hub:   hub,
			Data:  make(map[string]interface{}),
			rooms: make(map[string]bool),
		}

		restored := hub.RestoreConnectionState(newConn, "client123")
		if restored {
			t.Error("Should not restore expired state")
		}
	})
}

// TestQueueStrategies tests message queueing strategies
func TestQueueStrategies(t *testing.T) {
	t.Run("DropNewest", func(t *testing.T) {
		config := &Config{
			MessageQueueSize:     1,
			MessageQueueStrategy: QueueStrategyDropNewest,
		}

		hub := NewHubWithConfig(config)
		go hub.Run()
		defer hub.Shutdown()

		conn := &Connection{
			ID:   "conn1",
			hub:  hub,
			send: make(chan []byte, 1),
			Data: make(map[string]interface{}),
		}

		// Fill the queue
		conn.send <- []byte("message1")

		// Try to send another message - should be dropped
		err := conn.Send([]byte("message2"))
		if err != nil {
			t.Errorf("Send should not return error with DropNewest, got %v", err)
		}

		// Queue should still have only first message
		select {
		case msg := <-conn.send:
			if string(msg) != "message1" {
				t.Errorf("Expected 'message1', got %s", string(msg))
			}
		default:
			t.Error("Expected message in queue")
		}

		// Queue should be empty now
		select {
		case <-conn.send:
			t.Error("Queue should be empty")
		default:
			// Expected
		}
	})

	t.Run("DropOldest", func(t *testing.T) {
		config := &Config{
			MessageQueueSize:     1,
			MessageQueueStrategy: QueueStrategyDropOldest,
		}

		hub := NewHubWithConfig(config)
		go hub.Run()
		defer hub.Shutdown()

		conn := &Connection{
			ID:   "conn1",
			hub:  hub,
			send: make(chan []byte, 1),
			Data: make(map[string]interface{}),
		}

		// Fill the queue
		conn.send <- []byte("message1")

		// Send another message - should drop oldest
		err := conn.Send([]byte("message2"))
		if err != nil {
			t.Errorf("Send should not return error, got %v", err)
		}

		// Queue should have second message (first was dropped)
		select {
		case msg := <-conn.send:
			if string(msg) != "message2" {
				t.Errorf("Expected 'message2', got %s", string(msg))
			}
		default:
			t.Error("Expected message in queue")
		}
	})
}
