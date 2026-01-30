package websocket

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks WebSocket server metrics
type Metrics struct {
	// Connection metrics
	activeConnections   int64
	totalConnections    int64
	totalDisconnections int64
	rejectedConnections int64

	// Message metrics
	messagesSent     int64
	messagesReceived int64
	messagesFailed   int64
	bytesReceived    int64
	bytesSent        int64

	// Error metrics
	readErrors    int64
	writeErrors   int64
	handlerErrors int64

	// Room metrics
	activeRooms int64

	// Heartbeat metrics
	missedPongs     int64
	successfulPongs int64

	// Queue metrics
	queueOverflows  int64
	droppedMessages int64

	// Timestamp tracking
	lastMessageTime atomic.Value // time.Time
	startTime       time.Time

	// Per-connection metrics
	connMetrics map[string]*ConnectionMetrics
	connMu      sync.RWMutex

	// Enabled flag (accessed atomically for thread safety)
	enabled atomic.Bool
}

// ConnectionMetrics tracks metrics for a specific connection
type ConnectionMetrics struct {
	MessagesSent     int64
	MessagesReceived int64
	BytesSent        int64
	BytesReceived    int64
	Errors           int64
	MissedPongs      int64
	ConnectedAt      time.Time
	LastMessageAt    time.Time
}

// NewMetrics creates a new metrics tracker
func NewMetrics() *Metrics {
	m := &Metrics{
		startTime:   time.Now(),
		connMetrics: make(map[string]*ConnectionMetrics),
	}
	m.enabled.Store(true)
	m.lastMessageTime.Store(time.Time{})
	return m
}

// Enable enables metrics collection
func (m *Metrics) Enable() {
	m.enabled.Store(true)
}

// Disable disables metrics collection
func (m *Metrics) Disable() {
	m.enabled.Store(false)
}

// IsEnabled returns whether metrics are enabled
func (m *Metrics) IsEnabled() bool {
	return m.enabled.Load()
}

// Connection metrics

// IncrementConnections increments the active connections counter
func (m *Metrics) IncrementConnections() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.activeConnections, 1)
	atomic.AddInt64(&m.totalConnections, 1)
}

// DecrementConnections decrements the active connections counter
func (m *Metrics) DecrementConnections() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.activeConnections, -1)
	atomic.AddInt64(&m.totalDisconnections, 1)
}

// IncrementRejectedConnections increments the rejected connections counter
func (m *Metrics) IncrementRejectedConnections() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.rejectedConnections, 1)
}

// GetActiveConnections returns the number of active connections
func (m *Metrics) GetActiveConnections() int64 {
	return atomic.LoadInt64(&m.activeConnections)
}

// GetTotalConnections returns the total number of connections
func (m *Metrics) GetTotalConnections() int64 {
	return atomic.LoadInt64(&m.totalConnections)
}

// GetTotalDisconnections returns the total number of disconnections
func (m *Metrics) GetTotalDisconnections() int64 {
	return atomic.LoadInt64(&m.totalDisconnections)
}

// GetRejectedConnections returns the number of rejected connections
func (m *Metrics) GetRejectedConnections() int64 {
	return atomic.LoadInt64(&m.rejectedConnections)
}

// Message metrics

// IncrementMessagesSent increments the messages sent counter
func (m *Metrics) IncrementMessagesSent(bytes int64) {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.messagesSent, 1)
	atomic.AddInt64(&m.bytesSent, bytes)
	m.lastMessageTime.Store(time.Now())
}

// IncrementMessagesReceived increments the messages received counter
func (m *Metrics) IncrementMessagesReceived(bytes int64) {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.messagesReceived, 1)
	atomic.AddInt64(&m.bytesReceived, bytes)
	m.lastMessageTime.Store(time.Now())
}

// IncrementMessagesFailed increments the failed messages counter
func (m *Metrics) IncrementMessagesFailed() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.messagesFailed, 1)
}

// GetMessagesSent returns the number of messages sent
func (m *Metrics) GetMessagesSent() int64 {
	return atomic.LoadInt64(&m.messagesSent)
}

// GetMessagesReceived returns the number of messages received
func (m *Metrics) GetMessagesReceived() int64 {
	return atomic.LoadInt64(&m.messagesReceived)
}

// GetMessagesFailed returns the number of failed messages
func (m *Metrics) GetMessagesFailed() int64 {
	return atomic.LoadInt64(&m.messagesFailed)
}

// GetBytesReceived returns the number of bytes received
func (m *Metrics) GetBytesReceived() int64 {
	return atomic.LoadInt64(&m.bytesReceived)
}

// GetBytesSent returns the number of bytes sent
func (m *Metrics) GetBytesSent() int64 {
	return atomic.LoadInt64(&m.bytesSent)
}

// Error metrics

// IncrementReadErrors increments the read errors counter
func (m *Metrics) IncrementReadErrors() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.readErrors, 1)
}

// IncrementWriteErrors increments the write errors counter
func (m *Metrics) IncrementWriteErrors() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.writeErrors, 1)
}

// IncrementHandlerErrors increments the handler errors counter
func (m *Metrics) IncrementHandlerErrors() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.handlerErrors, 1)
}

// GetReadErrors returns the number of read errors
func (m *Metrics) GetReadErrors() int64 {
	return atomic.LoadInt64(&m.readErrors)
}

// GetWriteErrors returns the number of write errors
func (m *Metrics) GetWriteErrors() int64 {
	return atomic.LoadInt64(&m.writeErrors)
}

// GetHandlerErrors returns the number of handler errors
func (m *Metrics) GetHandlerErrors() int64 {
	return atomic.LoadInt64(&m.handlerErrors)
}

// GetTotalErrors returns the total number of errors
func (m *Metrics) GetTotalErrors() int64 {
	return m.GetReadErrors() + m.GetWriteErrors() + m.GetHandlerErrors()
}

// Room metrics

// IncrementRooms increments the active rooms counter
func (m *Metrics) IncrementRooms() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.activeRooms, 1)
}

// DecrementRooms decrements the active rooms counter
func (m *Metrics) DecrementRooms() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.activeRooms, -1)
}

// GetActiveRooms returns the number of active rooms
func (m *Metrics) GetActiveRooms() int64 {
	return atomic.LoadInt64(&m.activeRooms)
}

// Heartbeat metrics

// IncrementMissedPongs increments the missed pongs counter
func (m *Metrics) IncrementMissedPongs() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.missedPongs, 1)
}

// IncrementSuccessfulPongs increments the successful pongs counter
func (m *Metrics) IncrementSuccessfulPongs() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.successfulPongs, 1)
}

// GetMissedPongs returns the number of missed pongs
func (m *Metrics) GetMissedPongs() int64 {
	return atomic.LoadInt64(&m.missedPongs)
}

// GetSuccessfulPongs returns the number of successful pongs
func (m *Metrics) GetSuccessfulPongs() int64 {
	return atomic.LoadInt64(&m.successfulPongs)
}

// Queue metrics

// IncrementQueueOverflows increments the queue overflows counter
func (m *Metrics) IncrementQueueOverflows() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.queueOverflows, 1)
}

// IncrementDroppedMessages increments the dropped messages counter
func (m *Metrics) IncrementDroppedMessages() {
	if !m.enabled.Load() {
		return
	}
	atomic.AddInt64(&m.droppedMessages, 1)
}

// GetQueueOverflows returns the number of queue overflows
func (m *Metrics) GetQueueOverflows() int64 {
	return atomic.LoadInt64(&m.queueOverflows)
}

// GetDroppedMessages returns the number of dropped messages
func (m *Metrics) GetDroppedMessages() int64 {
	return atomic.LoadInt64(&m.droppedMessages)
}

// Time metrics

// GetLastMessageTime returns the time of the last message
func (m *Metrics) GetLastMessageTime() time.Time {
	val := m.lastMessageTime.Load()
	if t, ok := val.(time.Time); ok {
		return t
	}
	return time.Time{}
}

// GetStartTime returns the start time of the metrics
func (m *Metrics) GetStartTime() time.Time {
	return m.startTime
}

// GetUptime returns the uptime duration
func (m *Metrics) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

// Per-connection metrics

// RegisterConnection registers a new connection for metrics tracking
func (m *Metrics) RegisterConnection(connID string) {
	if !m.enabled.Load() {
		return
	}
	m.connMu.Lock()
	defer m.connMu.Unlock()

	m.connMetrics[connID] = &ConnectionMetrics{
		ConnectedAt:   time.Now(),
		LastMessageAt: time.Now(),
	}
}

// UnregisterConnection removes a connection from metrics tracking
func (m *Metrics) UnregisterConnection(connID string) {
	if !m.enabled.Load() {
		return
	}
	m.connMu.Lock()
	defer m.connMu.Unlock()

	delete(m.connMetrics, connID)
}

// IncrementConnectionMessagesSent increments messages sent for a connection
func (m *Metrics) IncrementConnectionMessagesSent(connID string, bytes int64) {
	if !m.enabled.Load() {
		return
	}
	m.connMu.Lock()
	defer m.connMu.Unlock()

	if cm, ok := m.connMetrics[connID]; ok {
		cm.MessagesSent++
		cm.BytesSent += bytes
		cm.LastMessageAt = time.Now()
	}
}

// IncrementConnectionMessagesReceived increments messages received for a connection
func (m *Metrics) IncrementConnectionMessagesReceived(connID string, bytes int64) {
	if !m.enabled.Load() {
		return
	}
	m.connMu.Lock()
	defer m.connMu.Unlock()

	if cm, ok := m.connMetrics[connID]; ok {
		cm.MessagesReceived++
		cm.BytesReceived += bytes
		cm.LastMessageAt = time.Now()
	}
}

// IncrementConnectionErrors increments errors for a connection
func (m *Metrics) IncrementConnectionErrors(connID string) {
	if !m.enabled.Load() {
		return
	}
	m.connMu.RLock()
	defer m.connMu.RUnlock()

	if cm, ok := m.connMetrics[connID]; ok {
		atomic.AddInt64(&cm.Errors, 1)
	}
}

// IncrementConnectionMissedPongs increments missed pongs for a connection
func (m *Metrics) IncrementConnectionMissedPongs(connID string) {
	if !m.enabled.Load() {
		return
	}
	m.connMu.RLock()
	defer m.connMu.RUnlock()

	if cm, ok := m.connMetrics[connID]; ok {
		atomic.AddInt64(&cm.MissedPongs, 1)
	}
}

// GetConnectionMetrics returns metrics for a specific connection
func (m *Metrics) GetConnectionMetrics(connID string) *ConnectionMetrics {
	m.connMu.RLock()
	defer m.connMu.RUnlock()

	if cm, ok := m.connMetrics[connID]; ok {
		// Return a copy to avoid concurrent access issues
		return &ConnectionMetrics{
			MessagesSent:     atomic.LoadInt64(&cm.MessagesSent),
			MessagesReceived: atomic.LoadInt64(&cm.MessagesReceived),
			BytesSent:        atomic.LoadInt64(&cm.BytesSent),
			BytesReceived:    atomic.LoadInt64(&cm.BytesReceived),
			Errors:           atomic.LoadInt64(&cm.Errors),
			MissedPongs:      atomic.LoadInt64(&cm.MissedPongs),
			ConnectedAt:      cm.ConnectedAt,
			LastMessageAt:    cm.LastMessageAt,
		}
	}
	return nil
}

// GetAllConnectionMetrics returns metrics for all connections
func (m *Metrics) GetAllConnectionMetrics() map[string]*ConnectionMetrics {
	m.connMu.RLock()
	defer m.connMu.RUnlock()

	result := make(map[string]*ConnectionMetrics)
	for id, cm := range m.connMetrics {
		result[id] = &ConnectionMetrics{
			MessagesSent:     atomic.LoadInt64(&cm.MessagesSent),
			MessagesReceived: atomic.LoadInt64(&cm.MessagesReceived),
			BytesSent:        atomic.LoadInt64(&cm.BytesSent),
			BytesReceived:    atomic.LoadInt64(&cm.BytesReceived),
			Errors:           atomic.LoadInt64(&cm.Errors),
			MissedPongs:      atomic.LoadInt64(&cm.MissedPongs),
			ConnectedAt:      cm.ConnectedAt,
			LastMessageAt:    cm.LastMessageAt,
		}
	}
	return result
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	atomic.StoreInt64(&m.activeConnections, 0)
	atomic.StoreInt64(&m.totalConnections, 0)
	atomic.StoreInt64(&m.totalDisconnections, 0)
	atomic.StoreInt64(&m.rejectedConnections, 0)
	atomic.StoreInt64(&m.messagesSent, 0)
	atomic.StoreInt64(&m.messagesReceived, 0)
	atomic.StoreInt64(&m.messagesFailed, 0)
	atomic.StoreInt64(&m.bytesReceived, 0)
	atomic.StoreInt64(&m.bytesSent, 0)
	atomic.StoreInt64(&m.readErrors, 0)
	atomic.StoreInt64(&m.writeErrors, 0)
	atomic.StoreInt64(&m.handlerErrors, 0)
	atomic.StoreInt64(&m.activeRooms, 0)
	atomic.StoreInt64(&m.missedPongs, 0)
	atomic.StoreInt64(&m.successfulPongs, 0)
	atomic.StoreInt64(&m.queueOverflows, 0)
	atomic.StoreInt64(&m.droppedMessages, 0)
	m.lastMessageTime.Store(time.Time{})
	m.startTime = time.Now()

	m.connMu.Lock()
	m.connMetrics = make(map[string]*ConnectionMetrics)
	m.connMu.Unlock()
}

// Snapshot returns a snapshot of current metrics
type MetricsSnapshot struct {
	ActiveConnections    int64
	TotalConnections     int64
	TotalDisconnections  int64
	RejectedConnections  int64
	MessagesSent         int64
	MessagesReceived     int64
	MessagesFailed       int64
	BytesReceived        int64
	BytesSent            int64
	ReadErrors           int64
	WriteErrors          int64
	HandlerErrors        int64
	TotalErrors          int64
	ActiveRooms          int64
	MissedPongs          int64
	SuccessfulPongs      int64
	QueueOverflows       int64
	DroppedMessages      int64
	LastMessageTime      time.Time
	StartTime            time.Time
	Uptime               time.Duration
	MessagesPerSecond    float64
	ConnectionsPerSecond float64
}

// GetSnapshot returns a snapshot of current metrics
func (m *Metrics) GetSnapshot() *MetricsSnapshot {
	uptime := m.GetUptime()
	uptimeSeconds := uptime.Seconds()

	snapshot := &MetricsSnapshot{
		ActiveConnections:   m.GetActiveConnections(),
		TotalConnections:    m.GetTotalConnections(),
		TotalDisconnections: m.GetTotalDisconnections(),
		RejectedConnections: m.GetRejectedConnections(),
		MessagesSent:        m.GetMessagesSent(),
		MessagesReceived:    m.GetMessagesReceived(),
		MessagesFailed:      m.GetMessagesFailed(),
		BytesReceived:       m.GetBytesReceived(),
		BytesSent:           m.GetBytesSent(),
		ReadErrors:          m.GetReadErrors(),
		WriteErrors:         m.GetWriteErrors(),
		HandlerErrors:       m.GetHandlerErrors(),
		TotalErrors:         m.GetTotalErrors(),
		ActiveRooms:         m.GetActiveRooms(),
		MissedPongs:         m.GetMissedPongs(),
		SuccessfulPongs:     m.GetSuccessfulPongs(),
		QueueOverflows:      m.GetQueueOverflows(),
		DroppedMessages:     m.GetDroppedMessages(),
		LastMessageTime:     m.GetLastMessageTime(),
		StartTime:           m.GetStartTime(),
		Uptime:              uptime,
	}

	// Calculate rates
	if uptimeSeconds > 0 {
		totalMessages := float64(snapshot.MessagesSent + snapshot.MessagesReceived)
		snapshot.MessagesPerSecond = totalMessages / uptimeSeconds
		snapshot.ConnectionsPerSecond = float64(snapshot.TotalConnections) / uptimeSeconds
	}

	return snapshot
}
