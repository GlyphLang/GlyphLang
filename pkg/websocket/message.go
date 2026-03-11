package websocket

import (
	"encoding/json"
	"errors"
	"time"
)

// Common errors
var (
	ErrConnectionClosed   = errors.New("connection closed")
	ErrInvalidMessage     = errors.New("invalid message format")
	ErrRoomNotFound       = errors.New("room not found")
	ErrConnectionNotFound = errors.New("connection not found")
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Message types
	MessageTypeText       MessageType = "text"
	MessageTypeBinary     MessageType = "binary"
	MessageTypeJSON       MessageType = "json"
	MessageTypeConnect    MessageType = "connect"
	MessageTypeDisconnect MessageType = "disconnect"
	MessageTypeError      MessageType = "error"
	MessageTypeJoinRoom   MessageType = "join_room"
	MessageTypeLeaveRoom  MessageType = "leave_room"
	MessageTypeBroadcast  MessageType = "broadcast"
	MessageTypePing       MessageType = "ping"
	MessageTypePong       MessageType = "pong"
)

// Message represents a WebSocket message
type Message struct {
	// Message type
	Type MessageType `json:"type"`

	// Event name (for custom events)
	Event string `json:"event,omitempty"`

	// Message payload
	Data interface{} `json:"data,omitempty"`

	// Target room (for room-specific messages)
	Room string `json:"room,omitempty"`

	// Target connection ID (for direct messages)
	Target string `json:"target,omitempty"`

	// Source connection ID
	ConnectionID string `json:"connection_id,omitempty"`

	// Timestamp
	Timestamp time.Time `json:"timestamp"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewMessage creates a new message
func NewMessage(msgType MessageType, data interface{}) *Message {
	return &Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// NewTextMessage creates a text message
func NewTextMessage(text string) *Message {
	return NewMessage(MessageTypeText, text)
}

// NewJSONMessage creates a JSON message
func NewJSONMessage(data interface{}) *Message {
	return NewMessage(MessageTypeJSON, data)
}

// NewBinaryMessage creates a binary message
func NewBinaryMessage(data []byte) *Message {
	return NewMessage(MessageTypeBinary, data)
}

// NewEventMessage creates an event message
func NewEventMessage(event string, data interface{}) *Message {
	msg := NewMessage(MessageTypeJSON, data)
	msg.Event = event
	return msg
}

// NewErrorMessage creates an error message
func NewErrorMessage(err error) *Message {
	return NewMessage(MessageTypeError, map[string]interface{}{
		"error": err.Error(),
	})
}

// ToJSON converts the message to JSON bytes
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON parses a message from JSON bytes
func FromJSON(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// SetMetadata sets metadata on the message
func (m *Message) SetMetadata(key string, value interface{}) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}
	m.Metadata[key] = value
}

// GetMetadata gets metadata from the message
func (m *Message) GetMetadata(key string) (interface{}, bool) {
	if m.Metadata == nil {
		return nil, false
	}
	value, ok := m.Metadata[key]
	return value, ok
}

// MessageContext represents the context for handling a message
type MessageContext struct {
	// The connection that sent the message
	Conn *Connection

	// The message
	Message *Message

	// Response channel (optional)
	Response chan *Message
}

// Reply sends a reply to the message sender
func (ctx *MessageContext) Reply(msgType MessageType, data interface{}) error {
	msg := NewMessage(msgType, data)
	msg.Target = ctx.Conn.ID
	return ctx.Conn.SendJSON(msg)
}

// ReplyJSON sends a JSON reply to the message sender
func (ctx *MessageContext) ReplyJSON(data interface{}) error {
	return ctx.Reply(MessageTypeJSON, data)
}

// ReplyError sends an error reply to the message sender
func (ctx *MessageContext) ReplyError(err error) error {
	return ctx.Reply(MessageTypeError, map[string]interface{}{
		"error": err.Error(),
	})
}

// Broadcast sends a message to all connections
func (ctx *MessageContext) Broadcast(msgType MessageType, data interface{}) {
	msg := NewMessage(msgType, data)
	msg.ConnectionID = ctx.Conn.ID
	msgData, err := msg.ToJSON()
	if err != nil {
		return
	}
	ctx.Conn.hub.broadcast <- msgData
}

// BroadcastToRoom sends a message to all connections in a room
func (ctx *MessageContext) BroadcastToRoom(room string, msgType MessageType, data interface{}) {
	msg := NewMessage(msgType, data)
	msg.Room = room
	msg.ConnectionID = ctx.Conn.ID
	msgData, err := msg.ToJSON()
	if err != nil {
		return
	}
	ctx.Conn.hub.broadcastToRoom <- &RoomMessage{
		RoomName:    room,
		Message:     msgData,
		ExcludeConn: ctx.Conn,
	}
}

// RoomAction represents an action on a room
type RoomAction struct {
	Conn     *Connection
	RoomName string
}

// RoomMessage represents a message to be sent to a room
type RoomMessage struct {
	RoomName    string
	Message     []byte
	ExcludeConn *Connection
}

// MessageHandler is a function that handles a message
type MessageHandler func(ctx *MessageContext) error

// EventHandler is a function that handles an event
type EventHandler func(conn *Connection) error
