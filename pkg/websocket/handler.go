package websocket

import (
	"fmt"
	"log"
	"sync"
)

// Handler routes messages to appropriate handlers
type Handler struct {
	// Message type handlers
	messageHandlers map[MessageType][]MessageHandler

	// Custom event handlers
	eventHandlers map[string][]MessageHandler

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// NewHandler creates a new message handler
func NewHandler() *Handler {
	h := &Handler{
		messageHandlers: make(map[MessageType][]MessageHandler),
		eventHandlers:   make(map[string][]MessageHandler),
	}

	// Register default handlers
	h.registerDefaultHandlers()

	return h
}

// registerDefaultHandlers registers built-in message handlers
func (h *Handler) registerDefaultHandlers() {
	// Handle join room requests
	h.On(MessageTypeJoinRoom, func(ctx *MessageContext) error {
		if ctx.Message.Room == "" {
			return fmt.Errorf("room name is required")
		}

		ctx.Conn.JoinRoom(ctx.Message.Room)

		// Send confirmation
		return ctx.ReplyJSON(map[string]interface{}{
			"type":   "join_room_success",
			"room":   ctx.Message.Room,
			"status": "joined",
		})
	})

	// Handle leave room requests
	h.On(MessageTypeLeaveRoom, func(ctx *MessageContext) error {
		if ctx.Message.Room == "" {
			return fmt.Errorf("room name is required")
		}

		ctx.Conn.LeaveRoom(ctx.Message.Room)

		// Send confirmation
		return ctx.ReplyJSON(map[string]interface{}{
			"type":   "leave_room_success",
			"room":   ctx.Message.Room,
			"status": "left",
		})
	})

	// Handle broadcast requests
	h.On(MessageTypeBroadcast, func(ctx *MessageContext) error {
		if ctx.Message.Room != "" {
			// Broadcast to room
			ctx.BroadcastToRoom(ctx.Message.Room, MessageTypeJSON, ctx.Message.Data)
		} else {
			// Broadcast to all
			ctx.Broadcast(MessageTypeJSON, ctx.Message.Data)
		}
		return nil
	})

	// Handle ping/pong
	h.On(MessageTypePing, func(ctx *MessageContext) error {
		return ctx.Reply(MessageTypePong, map[string]interface{}{
			"timestamp": ctx.Message.Timestamp,
		})
	})
}

// On registers a handler for a specific message type
func (h *Handler) On(msgType MessageType, handler MessageHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.messageHandlers[msgType] == nil {
		h.messageHandlers[msgType] = make([]MessageHandler, 0)
	}
	h.messageHandlers[msgType] = append(h.messageHandlers[msgType], handler)
}

// OnEvent registers a handler for a custom event
func (h *Handler) OnEvent(event string, handler MessageHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.eventHandlers[event] == nil {
		h.eventHandlers[event] = make([]MessageHandler, 0)
	}
	h.eventHandlers[event] = append(h.eventHandlers[event], handler)
}

// HandleMessage routes a message to the appropriate handlers
func (h *Handler) HandleMessage(ctx *MessageContext) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Handle custom events first
	if ctx.Message.Event != "" {
		if handlers, exists := h.eventHandlers[ctx.Message.Event]; exists {
			for _, handler := range handlers {
				if err := handler(ctx); err != nil {
					log.Printf("[WS] Event handler error for %s: %v", ctx.Message.Event, err)
					return err
				}
			}
			return nil
		}
	}

	// Handle by message type
	if handlers, exists := h.messageHandlers[ctx.Message.Type]; exists {
		for _, handler := range handlers {
			if err := handler(ctx); err != nil {
				log.Printf("[WS] Message handler error for %s: %v", ctx.Message.Type, err)
				return err
			}
		}
		return nil
	}

	// No handler found
	log.Printf("[WS] No handler for message type %s from connection %s", ctx.Message.Type, ctx.Conn.ID)
	return nil
}

// RemoveHandler removes all handlers for a message type
func (h *Handler) RemoveHandler(msgType MessageType) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.messageHandlers, msgType)
}

// RemoveEventHandler removes all handlers for an event
func (h *Handler) RemoveEventHandler(event string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.eventHandlers, event)
}

// GetMessageTypes returns all registered message types
func (h *Handler) GetMessageTypes() []MessageType {
	h.mu.RLock()
	defer h.mu.RUnlock()

	types := make([]MessageType, 0, len(h.messageHandlers))
	for msgType := range h.messageHandlers {
		types = append(types, msgType)
	}
	return types
}

// GetEvents returns all registered custom events
func (h *Handler) GetEvents() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	events := make([]string, 0, len(h.eventHandlers))
	for event := range h.eventHandlers {
		events = append(events, event)
	}
	return events
}

// HasHandler checks if a handler exists for a message type
func (h *Handler) HasHandler(msgType MessageType) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.messageHandlers[msgType]
	return exists
}

// HasEventHandler checks if a handler exists for an event
func (h *Handler) HasEventHandler(event string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.eventHandlers[event]
	return exists
}

// Clear removes all handlers
func (h *Handler) Clear() {
	h.mu.Lock()
	h.messageHandlers = make(map[MessageType][]MessageHandler)
	h.eventHandlers = make(map[string][]MessageHandler)
	h.mu.Unlock()

	// Register default handlers (calls On() which needs the lock)
	h.registerDefaultHandlers()
}
