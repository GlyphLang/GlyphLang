package websocket

import (
	"errors"
	"sync"
)

// Room represents a chat room or channel
type Room struct {
	// Room name
	Name string

	// Connections in this room
	connections map[*Connection]bool

	// Mutex for thread-safe access
	mu sync.RWMutex

	// Room metadata
	metadata map[string]interface{}

	// Maximum number of connections (0 = unlimited)
	maxConnections int

	// Room creation timestamp
	createdAt int64
}

// NewRoom creates a new room
func NewRoom(name string) *Room {
	return &Room{
		Name:        name,
		connections: make(map[*Connection]bool),
		metadata:    make(map[string]interface{}),
	}
}

// Add adds a connection to the room
func (r *Room) Add(conn *Connection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if room is full
	if r.maxConnections > 0 && len(r.connections) >= r.maxConnections {
		return ErrRoomFull
	}

	r.connections[conn] = true
	return nil
}

// Remove removes a connection from the room
func (r *Room) Remove(conn *Connection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.connections, conn)
}

// Has checks if a connection is in the room
func (r *Room) Has(conn *Connection) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.connections[conn]
}

// Broadcast sends a message to all connections in the room
func (r *Room) Broadcast(message []byte, exclude *Connection) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for conn := range r.connections {
		if exclude != nil && conn == exclude {
			continue
		}
		select {
		case conn.send <- message:
		default:
			// Connection send channel is full, skip it
		}
	}
}

// Size returns the number of connections in the room
func (r *Room) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.connections)
}

// Connections returns a list of all connections in the room
func (r *Room) Connections() []*Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conns := make([]*Connection, 0, len(r.connections))
	for conn := range r.connections {
		conns = append(conns, conn)
	}
	return conns
}

// SetMetadata sets room metadata
func (r *Room) SetMetadata(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metadata[key] = value
}

// GetMetadata gets room metadata
func (r *Room) GetMetadata(key string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.metadata[key]
	return value, ok
}

// GetAllMetadata returns all room metadata
func (r *Room) GetAllMetadata() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata := make(map[string]interface{})
	for k, v := range r.metadata {
		metadata[k] = v
	}
	return metadata
}

// SetMaxConnections sets the maximum number of connections
func (r *Room) SetMaxConnections(max int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.maxConnections = max
}

// GetMaxConnections returns the maximum number of connections
func (r *Room) GetMaxConnections() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.maxConnections
}

// Clear removes all connections from the room
func (r *Room) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connections = make(map[*Connection]bool)
}

// RoomManager manages all rooms
type RoomManager struct {
	rooms  map[string]*Room
	mu     sync.RWMutex
	config *Config
}

// NewRoomManager creates a new room manager
func NewRoomManager() *RoomManager {
	return NewRoomManagerWithConfig(DefaultConfig())
}

// NewRoomManagerWithConfig creates a new room manager with configuration
func NewRoomManagerWithConfig(config *Config) *RoomManager {
	return &RoomManager{
		rooms:  make(map[string]*Room),
		config: config,
	}
}

// CreateRoom creates a new room
func (rm *RoomManager) CreateRoom(name string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, exists := rm.rooms[name]; exists {
		return room
	}

	room := NewRoom(name)
	// Set max connections from config
	if rm.config != nil && rm.config.MaxConnectionsPerRoom > 0 {
		room.SetMaxConnections(rm.config.MaxConnectionsPerRoom)
	}
	rm.rooms[name] = room
	return room
}

// GetRoom gets a room by name
func (rm *RoomManager) GetRoom(name string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	room, exists := rm.rooms[name]
	return room, exists
}

// GetOrCreateRoom gets or creates a room
func (rm *RoomManager) GetOrCreateRoom(name string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, exists := rm.rooms[name]; exists {
		return room
	}

	room := NewRoom(name)
	// Set max connections from config
	if rm.config != nil && rm.config.MaxConnectionsPerRoom > 0 {
		room.SetMaxConnections(rm.config.MaxConnectionsPerRoom)
	}
	rm.rooms[name] = room
	return room
}

// DeleteRoom deletes a room
func (rm *RoomManager) DeleteRoom(name string) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.rooms[name]; !exists {
		return false
	}

	delete(rm.rooms, name)
	return true
}

// GetAllRooms returns all rooms
func (rm *RoomManager) GetAllRooms() []*Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rooms := make([]*Room, 0, len(rm.rooms))
	for _, room := range rm.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// GetRoomNames returns all room names
func (rm *RoomManager) GetRoomNames() []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	names := make([]string, 0, len(rm.rooms))
	for name := range rm.rooms {
		names = append(names, name)
	}
	return names
}

// Count returns the number of rooms
func (rm *RoomManager) Count() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.rooms)
}

// Clear removes all rooms
func (rm *RoomManager) Clear() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.rooms = make(map[string]*Room)
}

// AddConnectionToRoom adds a connection to a room
func (rm *RoomManager) AddConnectionToRoom(conn *Connection, roomName string) error {
	room := rm.GetOrCreateRoom(roomName)
	return room.Add(conn)
}

// RemoveConnectionFromRoom removes a connection from a room
func (rm *RoomManager) RemoveConnectionFromRoom(conn *Connection, roomName string) {
	room, exists := rm.GetRoom(roomName)
	if !exists {
		return
	}
	room.Remove(conn)
}

// RemoveConnectionFromAllRooms removes a connection from all rooms
func (rm *RoomManager) RemoveConnectionFromAllRooms(conn *Connection) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for _, room := range rm.rooms {
		room.Remove(conn)
	}
}

// BroadcastToRoom sends a message to all connections in a room
func (rm *RoomManager) BroadcastToRoom(roomName string, message []byte, exclude *Connection) error {
	room, exists := rm.GetRoom(roomName)
	if !exists {
		return ErrRoomNotFound
	}
	room.Broadcast(message, exclude)
	return nil
}

// GetRoomSize returns the number of connections in a room
func (rm *RoomManager) GetRoomSize(roomName string) int {
	room, exists := rm.GetRoom(roomName)
	if !exists {
		return 0
	}
	return room.Size()
}

// Common errors
var (
	ErrRoomFull = errors.New("room is full")
)
