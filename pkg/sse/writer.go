package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

// Event represents a Server-Sent Event.
type Event struct {
	ID    string      // Optional event ID
	Type  string      // Optional event type (defaults to "message")
	Data  interface{} // Event data (will be JSON-encoded if not a string)
	Retry int         // Optional reconnection time in milliseconds
}

// Writer writes Server-Sent Events to an http.ResponseWriter.
type Writer struct {
	w       http.ResponseWriter
	flusher http.Flusher
	mu      sync.Mutex
	eventID int64
}

// NewWriter creates a new SSE Writer and sets the required response headers.
// Returns an error if the ResponseWriter does not support flushing.
func NewWriter(w http.ResponseWriter) (*Writer, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("response writer does not support flushing (required for SSE)")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	return &Writer{
		w:       w,
		flusher: flusher,
	}, nil
}

// Send writes a single SSE event and flushes the response.
func (sw *Writer) Send(event Event) error {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	var sb strings.Builder

	// Auto-assign ID if not provided
	if event.ID == "" {
		id := atomic.AddInt64(&sw.eventID, 1)
		event.ID = fmt.Sprintf("%d", id)
	}
	fmt.Fprintf(&sb, "id: %s\n", event.ID)

	if event.Type != "" {
		fmt.Fprintf(&sb, "event: %s\n", event.Type)
	}

	if event.Retry > 0 {
		fmt.Fprintf(&sb, "retry: %d\n", event.Retry)
	}

	dataStr, err := formatData(event.Data)
	if err != nil {
		return fmt.Errorf("failed to format event data: %w", err)
	}

	// SSE spec requires each line of data to be prefixed with "data: "
	for _, line := range strings.Split(dataStr, "\n") {
		fmt.Fprintf(&sb, "data: %s\n", line)
	}

	// Blank line terminates the event
	sb.WriteString("\n")

	if _, err := fmt.Fprint(sw.w, sb.String()); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	sw.flusher.Flush()
	return nil
}

// SendData sends a simple data-only event.
func (sw *Writer) SendData(data interface{}) error {
	return sw.Send(Event{Data: data})
}

// SendComment sends an SSE comment (used for keep-alive).
func (sw *Writer) SendComment(text string) error {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	if _, err := fmt.Fprintf(sw.w, ": %s\n\n", text); err != nil {
		return fmt.Errorf("failed to write comment: %w", err)
	}
	sw.flusher.Flush()
	return nil
}

// SendEvent implements the interpreter.SSEWriter interface.
// It sends an event with the given data and optional event type.
func (sw *Writer) SendEvent(data interface{}, eventType string) error {
	return sw.Send(Event{
		Type: eventType,
		Data: data,
	})
}

// formatData converts the data to a string suitable for SSE.
func formatData(data interface{}) (string, error) {
	switch v := data.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(jsonBytes), nil
	}
}
