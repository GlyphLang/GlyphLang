package websocket

import (
	"encoding/json"
	"errors"
	"time"
)

// VMHandler implements the vm.WebSocketHandler interface
// It bridges the VM's WebSocket operations to the actual WebSocket Connection and Hub
type VMHandler struct {
	conn      *Connection
	hub       *Hub
	startTime time.Time
}

// NewVMHandler creates a new VMHandler for a specific connection
func NewVMHandler(conn *Connection, hub *Hub) *VMHandler {
	return &VMHandler{
		conn:      conn,
		hub:       hub,
		startTime: hub.metrics.GetStartTime(),
	}
}

// VMStatsHandler implements vm.WebSocketHandler for stats-only access from HTTP routes
// It provides read-only access to WebSocket server statistics without requiring a connection
type VMStatsHandler struct {
	hub *Hub
}

// NewVMStatsHandler creates a stats-only handler for HTTP routes
func NewVMStatsHandler(hub *Hub) *VMStatsHandler {
	return &VMStatsHandler{hub: hub}
}

// Send is not available in stats-only mode
func (h *VMStatsHandler) Send(message interface{}) error {
	return errors.New("ws.send not available in HTTP routes")
}

// Broadcast is not available in stats-only mode
func (h *VMStatsHandler) Broadcast(message interface{}) error {
	return errors.New("ws.broadcast not available in HTTP routes")
}

// BroadcastToRoom is not available in stats-only mode
func (h *VMStatsHandler) BroadcastToRoom(room string, message interface{}) error {
	return errors.New("ws.broadcast_to_room not available in HTTP routes")
}

// JoinRoom is not available in stats-only mode
func (h *VMStatsHandler) JoinRoom(room string) error {
	return errors.New("ws.join not available in HTTP routes")
}

// LeaveRoom is not available in stats-only mode
func (h *VMStatsHandler) LeaveRoom(room string) error {
	return errors.New("ws.leave not available in HTTP routes")
}

// Close is not available in stats-only mode
func (h *VMStatsHandler) Close(reason string) error {
	return errors.New("ws.close not available in HTTP routes")
}

// GetRooms returns the list of all rooms
func (h *VMStatsHandler) GetRooms() []string {
	rm := h.hub.GetRoomManager()
	return rm.GetRoomNames()
}

// GetRoomClients returns the list of client IDs in a room
func (h *VMStatsHandler) GetRoomClients(room string) []string {
	rm := h.hub.GetRoomManager()
	r, exists := rm.GetRoom(room)
	if !exists {
		return []string{}
	}

	conns := r.Connections()
	clients := make([]string, len(conns))
	for i, conn := range conns {
		clients[i] = conn.ID
	}
	return clients
}

// GetConnectionID returns empty string in stats-only mode
func (h *VMStatsHandler) GetConnectionID() string {
	return ""
}

// GetConnectionCount returns the total number of active connections
func (h *VMStatsHandler) GetConnectionCount() int {
	return h.hub.GetConnectionCount()
}

// GetUptime returns the server uptime in seconds
func (h *VMStatsHandler) GetUptime() int64 {
	return int64(h.hub.metrics.GetUptime().Seconds())
}

// Send sends a message to the current WebSocket connection
func (h *VMHandler) Send(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return h.conn.Send(data)
}

// Broadcast sends a message to all connections
func (h *VMHandler) Broadcast(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	h.hub.Broadcast(data)
	return nil
}

// BroadcastToRoom sends a message to all connections in a room
func (h *VMHandler) BroadcastToRoom(room string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	h.hub.BroadcastToRoom(room, data, nil)
	return nil
}

// JoinRoom adds the current connection to a room
func (h *VMHandler) JoinRoom(room string) error {
	h.conn.JoinRoom(room)
	return nil
}

// LeaveRoom removes the current connection from a room
func (h *VMHandler) LeaveRoom(room string) error {
	h.conn.LeaveRoom(room)
	return nil
}

// Close closes the current WebSocket connection
func (h *VMHandler) Close(reason string) error {
	// Send close reason if provided
	if reason != "" {
		h.conn.SendJSON(map[string]string{
			"type":   "close",
			"reason": reason,
		})
	}
	return h.conn.Close()
}

// GetRooms returns the list of rooms the current connection is in
func (h *VMHandler) GetRooms() []string {
	return h.conn.GetRooms()
}

// GetRoomClients returns the list of client IDs in a room
func (h *VMHandler) GetRoomClients(room string) []string {
	rm := h.hub.GetRoomManager()
	r, exists := rm.GetRoom(room)
	if !exists {
		return []string{}
	}

	conns := r.Connections()
	clients := make([]string, len(conns))
	for i, conn := range conns {
		clients[i] = conn.ID
	}
	return clients
}

// GetConnectionID returns the current connection's ID
func (h *VMHandler) GetConnectionID() string {
	return h.conn.ID
}

// GetConnectionCount returns the total number of active connections
func (h *VMHandler) GetConnectionCount() int {
	return h.hub.GetConnectionCount()
}

// GetUptime returns the server uptime in seconds
func (h *VMHandler) GetUptime() int64 {
	return int64(h.hub.metrics.GetUptime().Seconds())
}
