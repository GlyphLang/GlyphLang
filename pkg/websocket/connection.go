package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Removed constants - now using configuration

// Connection represents a WebSocket connection
type Connection struct {
	// Unique identifier for this connection
	ID string

	// The WebSocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// The hub this connection belongs to
	hub *Hub

	// Custom data associated with the connection
	Data map[string]interface{}

	// Mutex for protecting Data
	mu sync.RWMutex

	// Rooms this connection has joined
	rooms map[string]bool

	// Mutex for protecting rooms
	roomsMu sync.RWMutex

	// Path parameters extracted from the WebSocket route pattern (e.g., :room from /chat/:room)
	PathParams map[string]string

	// routePattern is the original route pattern this connection matched (e.g., /chat/:room)
	// Used internally to filter handlers when multiple WebSocket routes exist
	routePattern string

	// Heartbeat tracking
	missedPongs    int
	lastPongTime   time.Time
	heartbeatMu    sync.RWMutex

	// Message queue for backpressure handling
	messageQueue   [][]byte
	queueMu        sync.Mutex
}

// RoutePattern returns the route pattern this connection matched
func (c *Connection) RoutePattern() string {
	return c.routePattern
}

// SetRoutePattern sets the route pattern for this connection
func (c *Connection) SetRoutePattern(pattern string) {
	c.routePattern = pattern
}

// NewConnection creates a new WebSocket connection
func NewConnection(id string, conn *websocket.Conn, hub *Hub) *Connection {
	queueSize := hub.config.MessageQueueSize
	if queueSize <= 0 {
		queueSize = 256
	}

	return &Connection{
		ID:           id,
		conn:         conn,
		send:         make(chan []byte, queueSize),
		hub:          hub,
		Data:         make(map[string]interface{}),
		rooms:        make(map[string]bool),
		PathParams:   make(map[string]string),
		lastPongTime: time.Now(),
		messageQueue: make([][]byte, 0),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Connection) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		c.hub.connWg.Done()
	}()

	config := c.hub.config
	pongWait := config.PongWaitTimeout
	maxMessageSize := config.MaxMessageSize

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.heartbeatMu.Lock()
		c.lastPongTime = time.Now()
		c.missedPongs = 0
		c.heartbeatMu.Unlock()

		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.hub.metrics.IncrementSuccessfulPongs()
		return nil
	})

	c.conn.SetReadLimit(maxMessageSize)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] Connection %s read error: %v", c.ID, err)
				c.hub.metrics.IncrementReadErrors()
			}
			break
		}

		// Track metrics
		c.hub.metrics.IncrementMessagesReceived(int64(len(message)))
		c.hub.metrics.IncrementConnectionMessagesReceived(c.ID, int64(len(message)))

		// Parse message and route it
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("[WS] Connection %s failed to parse message: %v", c.ID, err)
			c.hub.metrics.IncrementConnectionErrors(c.ID)
			continue
		}

		// Set connection ID in message
		msg.ConnectionID = c.ID

		// Send to hub for routing
		c.hub.handleMessage <- &MessageContext{
			Conn:    c,
			Message: &msg,
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Connection) WritePump() {
	config := c.hub.config
	var ticker *time.Ticker

	if config.EnableHeartbeat {
		ticker = time.NewTicker(config.HeartbeatInterval)
	} else {
		// Use a long interval if heartbeat is disabled
		ticker = time.NewTicker(24 * time.Hour)
	}

	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.hub.connWg.Done()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.hub.metrics.IncrementWriteErrors()
				return
			}
			w.Write(message)

			// Track metrics
			messageSize := int64(len(message))
			c.hub.metrics.IncrementMessagesSent(messageSize)
			c.hub.metrics.IncrementConnectionMessagesSent(c.ID, messageSize)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				msg := <-c.send
				w.Write(msg)
				msgSize := int64(len(msg))
				c.hub.metrics.IncrementMessagesSent(msgSize)
				c.hub.metrics.IncrementConnectionMessagesSent(c.ID, msgSize)
			}

			if err := w.Close(); err != nil {
				c.hub.metrics.IncrementWriteErrors()
				return
			}

		case <-ticker.C:
			if !config.EnableHeartbeat {
				continue
			}

			c.heartbeatMu.Lock()
			c.missedPongs++
			missedCount := c.missedPongs
			c.heartbeatMu.Unlock()

			// Check if connection should be closed due to missed pongs
			if missedCount > config.MaxMissedPongs {
				log.Printf("[WS] Connection %s timeout (missed %d pongs)", c.ID, missedCount)
				c.hub.metrics.IncrementMissedPongs()
				c.hub.metrics.IncrementConnectionMissedPongs(c.ID)
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.metrics.IncrementWriteErrors()
				return
			}
		}
	}
}

// Send sends a message to this connection
func (c *Connection) Send(message []byte) error {
	config := c.hub.config

	select {
	case c.send <- message:
		return nil
	default:
		// Channel is full, apply backpressure strategy
		switch config.MessageQueueStrategy {
		case QueueStrategyDropOldest:
			// Try to drop the oldest message
			select {
			case <-c.send:
				c.hub.metrics.IncrementDroppedMessages()
			default:
			}
			// Try again to send
			select {
			case c.send <- message:
				return nil
			default:
				c.hub.metrics.IncrementQueueOverflows()
				return ErrConnectionClosed
			}

		case QueueStrategyDropNewest:
			// Drop the new message
			c.hub.metrics.IncrementDroppedMessages()
			c.hub.metrics.IncrementQueueOverflows()
			return nil

		case QueueStrategyBlock:
			fallthrough
		default:
			// Block until space is available or connection closes
			c.send <- message
			return nil
		}
	}
}

// SendJSON sends a JSON message to this connection
func (c *Connection) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Send(data)
}

// Close closes the connection
func (c *Connection) Close() error {
	c.hub.unregister <- c
	return c.conn.Close()
}

// SetData sets custom data on the connection
func (c *Connection) SetData(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data[key] = value
}

// GetData gets custom data from the connection
func (c *Connection) GetData(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.Data[key]
	return value, ok
}

// JoinRoom adds this connection to a room
func (c *Connection) JoinRoom(roomName string) {
	c.roomsMu.Lock()
	c.rooms[roomName] = true
	c.roomsMu.Unlock()

	// Add to room manager synchronously to ensure the room exists
	// before any subsequent operations (like broadcast_to_room)
	rm := c.hub.GetRoomManager()
	if err := rm.AddConnectionToRoom(c, roomName); err != nil {
		log.Printf("[WS] Failed to join room %s: %v", roomName, err)
	} else {
		log.Printf("[WS] Connection %s joined room %s", c.ID, roomName)
	}
}

// LeaveRoom removes this connection from a room
func (c *Connection) LeaveRoom(roomName string) {
	c.roomsMu.Lock()
	delete(c.rooms, roomName)
	c.roomsMu.Unlock()

	// Remove from room manager synchronously
	rm := c.hub.GetRoomManager()
	rm.RemoveConnectionFromRoom(c, roomName)
	log.Printf("[WS] Connection %s left room %s", c.ID, roomName)
}

// GetRooms returns all rooms this connection has joined
func (c *Connection) GetRooms() []string {
	c.roomsMu.RLock()
	defer c.roomsMu.RUnlock()

	rooms := make([]string, 0, len(c.rooms))
	for room := range c.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// IsInRoom checks if connection is in a room
func (c *Connection) IsInRoom(roomName string) bool {
	c.roomsMu.RLock()
	defer c.roomsMu.RUnlock()
	return c.rooms[roomName]
}

// GetMissedPongs returns the number of missed pongs
func (c *Connection) GetMissedPongs() int {
	c.heartbeatMu.RLock()
	defer c.heartbeatMu.RUnlock()
	return c.missedPongs
}

// GetLastPongTime returns the time of the last pong
func (c *Connection) GetLastPongTime() time.Time {
	c.heartbeatMu.RLock()
	defer c.heartbeatMu.RUnlock()
	return c.lastPongTime
}

// IsHealthy checks if the connection is healthy based on heartbeat
func (c *Connection) IsHealthy() bool {
	if !c.hub.config.EnableHeartbeat {
		return true
	}

	c.heartbeatMu.RLock()
	defer c.heartbeatMu.RUnlock()

	// Check if we've exceeded max missed pongs
	if c.missedPongs > c.hub.config.MaxMissedPongs {
		return false
	}

	// Check if last pong is within timeout
	if time.Since(c.lastPongTime) > c.hub.config.HeartbeatTimeout {
		return false
	}

	return true
}
