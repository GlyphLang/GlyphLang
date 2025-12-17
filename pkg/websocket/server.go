package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now (should be configurable in production)
		return true
	},
}

// Hub maintains the set of active connections and broadcasts messages
type Hub struct {
	// Registered connections
	connections map[*Connection]bool

	// Inbound messages from connections
	handleMessage chan *MessageContext

	// Register requests from connections
	register chan *Connection

	// Unregister requests from connections
	unregister chan *Connection

	// Broadcast messages to all connections
	broadcast chan []byte

	// Broadcast messages to specific room
	broadcastToRoom chan *RoomMessage

	// Join room requests
	joinRoom chan *RoomAction

	// Leave room requests
	leaveRoom chan *RoomAction

	// Room manager
	roomManager *RoomManager

	// Message handler
	handler *Handler

	// Connection lifecycle handlers
	onConnect    []EventHandler
	onDisconnect []EventHandler

	// Mutex for handlers
	handlerMu sync.RWMutex

	// Shutdown channel
	shutdown chan struct{}

	// WaitGroup for graceful shutdown (hub.Run)
	wg sync.WaitGroup

	// WaitGroup for connection goroutines (ReadPump/WritePump)
	connWg sync.WaitGroup

	// Configuration
	config *Config

	// Metrics
	metrics *Metrics

	// Connection states for reconnection
	connectionStates map[string]*ConnectionState
	stateMu          sync.RWMutex

	// Mutex for connections map
	connMu sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return NewHubWithConfig(DefaultConfig())
}

// NewHubWithConfig creates a new Hub with custom configuration
func NewHubWithConfig(config *Config) *Hub {
	if config == nil {
		config = DefaultConfig()
	}
	config.Validate()

	return &Hub{
		connections:      make(map[*Connection]bool),
		handleMessage:    make(chan *MessageContext, 256),
		register:         make(chan *Connection),
		unregister:       make(chan *Connection),
		broadcast:        make(chan []byte, 256),
		broadcastToRoom:  make(chan *RoomMessage, 256),
		joinRoom:         make(chan *RoomAction, 256),
		leaveRoom:        make(chan *RoomAction, 256),
		roomManager:      NewRoomManagerWithConfig(config),
		handler:          NewHandler(),
		onConnect:        make([]EventHandler, 0),
		onDisconnect:     make([]EventHandler, 0),
		shutdown:         make(chan struct{}),
		config:           config,
		metrics:          NewMetrics(),
		connectionStates: make(map[string]*ConnectionState),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	h.wg.Add(1)
	defer h.wg.Done()

	for {
		select {
		case conn := <-h.register:
			h.connMu.Lock()
			// Check connection limit
			if h.config.MaxConnectionsPerHub > 0 && len(h.connections) >= h.config.MaxConnectionsPerHub {
				h.connMu.Unlock()
				log.Printf("[WS] Connection rejected (limit reached): %s", conn.ID)
				h.metrics.IncrementRejectedConnections()
				conn.conn.Close()
				continue
			}

			h.connections[conn] = true
			h.connMu.Unlock()

			h.metrics.IncrementConnections()
			h.metrics.RegisterConnection(conn.ID)
			log.Printf("[WS] Connection registered: %s (total: %d)", conn.ID, len(h.connections))

			// Call onConnect handlers
			h.handlerMu.RLock()
			for _, handler := range h.onConnect {
				if err := handler(conn); err != nil {
					log.Printf("[WS] onConnect handler error: %v", err)
					h.metrics.IncrementHandlerErrors()
				}
			}
			h.handlerMu.RUnlock()

		case conn := <-h.unregister:
			h.connMu.Lock()
			if _, ok := h.connections[conn]; ok {
				delete(h.connections, conn)
				h.connMu.Unlock()

				close(conn.send)
				h.roomManager.RemoveConnectionFromAllRooms(conn)
				h.metrics.DecrementConnections()
				h.metrics.UnregisterConnection(conn.ID)

				// Save connection state for reconnection
				if h.config.EnableReconnection && h.config.PreserveClientState {
					h.saveConnectionState(conn)
				}

				log.Printf("[WS] Connection unregistered: %s (total: %d)", conn.ID, len(h.connections))

				// Call onDisconnect handlers
				h.handlerMu.RLock()
				for _, handler := range h.onDisconnect {
					if err := handler(conn); err != nil {
						log.Printf("[WS] onDisconnect handler error: %v", err)
						h.metrics.IncrementHandlerErrors()
					}
				}
				h.handlerMu.RUnlock()
			} else {
				h.connMu.Unlock()
			}

		case msgCtx := <-h.handleMessage:
			// Route message to handler
			if err := h.handler.HandleMessage(msgCtx); err != nil {
				log.Printf("[WS] Message handling error: %v", err)
				h.metrics.IncrementHandlerErrors()
			}

		case message := <-h.broadcast:
			for conn := range h.connections {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(h.connections, conn)
					h.roomManager.RemoveConnectionFromAllRooms(conn)
				}
			}

		case roomMsg := <-h.broadcastToRoom:
			if room, exists := h.roomManager.GetRoom(roomMsg.RoomName); exists {
				room.Broadcast(roomMsg.Message, roomMsg.ExcludeConn)
			}

		case action := <-h.joinRoom:
			if err := h.roomManager.AddConnectionToRoom(action.Conn, action.RoomName); err != nil {
				log.Printf("[WS] Failed to join room %s: %v", action.RoomName, err)
			} else {
				log.Printf("[WS] Connection %s joined room %s", action.Conn.ID, action.RoomName)
			}

		case action := <-h.leaveRoom:
			h.roomManager.RemoveConnectionFromRoom(action.Conn, action.RoomName)
			log.Printf("[WS] Connection %s left room %s", action.Conn.ID, action.RoomName)

		case <-h.shutdown:
			log.Printf("[WS] Hub shutting down...")
			return
		}
	}
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown() {
	// Close all WebSocket connections first (while hub is still running)
	for conn := range h.connections {
		if conn.conn != nil {
			conn.conn.Close() // Close the underlying websocket, not conn.Close() which sends to unregister
		}
	}

	// Wait for all connection goroutines to finish (they will unregister themselves)
	h.connWg.Wait()

	// Now shutdown the hub
	close(h.shutdown)
	h.wg.Wait()
}

// GetConnections returns all active connections
func (h *Hub) GetConnections() []*Connection {
	conns := make([]*Connection, 0, len(h.connections))
	for conn := range h.connections {
		conns = append(conns, conn)
	}
	return conns
}

// GetConnectionCount returns the number of active connections
func (h *Hub) GetConnectionCount() int {
	h.connMu.RLock()
	defer h.connMu.RUnlock()
	return len(h.connections)
}

// GetConnection returns a connection by ID
func (h *Hub) GetConnection(id string) (*Connection, bool) {
	h.connMu.RLock()
	defer h.connMu.RUnlock()
	for conn := range h.connections {
		if conn.ID == id {
			return conn, true
		}
	}
	return nil, false
}

// Broadcast sends a message to all connections
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// BroadcastJSON sends a JSON message to all connections
func (h *Hub) BroadcastJSON(v interface{}) error {
	msg := NewJSONMessage(v)
	data, err := msg.ToJSON()
	if err != nil {
		return err
	}
	h.Broadcast(data)
	return nil
}

// BroadcastToRoom sends a message to all connections in a room
func (h *Hub) BroadcastToRoom(roomName string, message []byte, exclude *Connection) {
	h.broadcastToRoom <- &RoomMessage{
		RoomName:    roomName,
		Message:     message,
		ExcludeConn: exclude,
	}
}

// BroadcastJSONToRoom sends a JSON message to all connections in a room
func (h *Hub) BroadcastJSONToRoom(roomName string, v interface{}, exclude *Connection) error {
	msg := NewJSONMessage(v)
	data, err := msg.ToJSON()
	if err != nil {
		return err
	}
	h.BroadcastToRoom(roomName, data, exclude)
	return nil
}

// OnConnect registers a handler for connection events
func (h *Hub) OnConnect(handler EventHandler) {
	h.handlerMu.Lock()
	defer h.handlerMu.Unlock()
	h.onConnect = append(h.onConnect, handler)
}

// OnDisconnect registers a handler for disconnection events
func (h *Hub) OnDisconnect(handler EventHandler) {
	h.handlerMu.Lock()
	defer h.handlerMu.Unlock()
	h.onDisconnect = append(h.onDisconnect, handler)
}

// OnMessage registers a handler for message events
func (h *Hub) OnMessage(msgType MessageType, handler MessageHandler) {
	h.handler.On(msgType, handler)
}

// OnEvent registers a handler for custom events
func (h *Hub) OnEvent(event string, handler MessageHandler) {
	h.handler.OnEvent(event, handler)
}

// GetRoomManager returns the room manager
func (h *Hub) GetRoomManager() *RoomManager {
	return h.roomManager
}

// Server represents a WebSocket server
type Server struct {
	hub      *Hub
	upgrader websocket.Upgrader
}

// NewServer creates a new WebSocket server
func NewServer() *Server {
	hub := NewHub()
	go hub.Run()

	return &Server{
		hub:      hub,
		upgrader: upgrader,
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade error: %v", err)
		return
	}

	// Generate unique connection ID
	id := generateConnectionID()

	// Create connection wrapper
	wsConn := NewConnection(id, conn, s.hub)

	// Register connection
	s.hub.register <- wsConn

	// Track connection goroutines for graceful shutdown
	s.hub.connWg.Add(2)

	// Start read/write pumps
	go wsConn.WritePump()
	go wsConn.ReadPump()
}

// GetHub returns the hub
func (s *Server) GetHub() *Hub {
	return s.hub
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	s.hub.Shutdown()
}

// OnConnect registers a connection event handler
func (s *Server) OnConnect(handler EventHandler) {
	s.hub.OnConnect(handler)
}

// OnDisconnect registers a disconnection event handler
func (s *Server) OnDisconnect(handler EventHandler) {
	s.hub.OnDisconnect(handler)
}

// OnMessage registers a message event handler
func (s *Server) OnMessage(msgType MessageType, handler MessageHandler) {
	s.hub.OnMessage(msgType, handler)
}

// OnEvent registers a custom event handler
func (s *Server) OnEvent(event string, handler MessageHandler) {
	s.hub.OnEvent(event, handler)
}

// generateConnectionID generates a unique connection ID
func generateConnectionID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp if random generation fails
		return fmt.Sprintf("conn-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// GetConfig returns the hub configuration
func (h *Hub) GetConfig() *Config {
	return h.config
}

// GetMetrics returns the hub metrics
func (h *Hub) GetMetrics() *Metrics {
	return h.metrics
}

// saveConnectionState saves connection state for reconnection
func (h *Hub) saveConnectionState(conn *Connection) {
	h.stateMu.Lock()
	defer h.stateMu.Unlock()

	// Get client ID from connection data
	clientID, ok := conn.GetData("clientID")
	if !ok {
		// Use connection ID as fallback
		clientID = conn.ID
	}

	state := &ConnectionState{
		ClientID: clientID.(string),
		LastSeen: time.Now(),
		Data:     make(map[string]interface{}),
		Rooms:    conn.GetRooms(),
	}

	// Copy connection data
	conn.mu.RLock()
	for k, v := range conn.Data {
		state.Data[k] = v
	}
	conn.mu.RUnlock()

	h.connectionStates[state.ClientID] = state

	// Start cleanup timer
	go h.cleanupConnectionState(state.ClientID)
}

// cleanupConnectionState removes connection state after timeout
func (h *Hub) cleanupConnectionState(clientID string) {
	timeout := h.config.ReconnectionTimeout
	time.Sleep(timeout)

	h.stateMu.Lock()
	defer h.stateMu.Unlock()

	if state, ok := h.connectionStates[clientID]; ok {
		if time.Since(state.LastSeen) >= timeout {
			delete(h.connectionStates, clientID)
			log.Printf("[WS] Connection state expired for client: %s", clientID)
		}
	}
}

// RestoreConnectionState restores connection state from a previous connection
func (h *Hub) RestoreConnectionState(conn *Connection, clientID string) bool {
	// Get and remove state with lock
	h.stateMu.Lock()
	state, ok := h.connectionStates[clientID]
	if !ok {
		h.stateMu.Unlock()
		return false
	}

	// Check if state is not too old
	if time.Since(state.LastSeen) > h.config.MaxReconnectionTime {
		delete(h.connectionStates, clientID)
		h.stateMu.Unlock()
		return false
	}

	// Make a copy of the data and rooms before releasing the lock
	dataCopy := make(map[string]interface{})
	for k, v := range state.Data {
		dataCopy[k] = v
	}
	roomsCopy := make([]string, len(state.Rooms))
	copy(roomsCopy, state.Rooms)

	// Remove state after copying
	delete(h.connectionStates, clientID)
	h.stateMu.Unlock()

	// Restore data (without holding the hub lock)
	conn.mu.Lock()
	for k, v := range dataCopy {
		conn.Data[k] = v
	}
	conn.Data["clientID"] = clientID
	conn.mu.Unlock()

	// Restore rooms
	for _, room := range roomsCopy {
		conn.JoinRoom(room)
	}

	log.Printf("[WS] Connection state restored for client: %s", clientID)
	return true
}
