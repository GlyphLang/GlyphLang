package websocket

import "time"

// Config holds WebSocket server configuration
type Config struct {
	// Connection limits
	MaxConnectionsPerHub  int
	MaxConnectionsPerRoom int

	// Heartbeat/Ping-Pong settings
	EnableHeartbeat   bool
	HeartbeatInterval time.Duration
	HeartbeatTimeout  time.Duration
	PongWaitTimeout   time.Duration
	MaxMissedPongs    int

	// Reconnection settings
	EnableReconnection   bool
	ReconnectionTimeout  time.Duration
	MaxReconnectionTime  time.Duration
	PreserveClientState  bool

	// Message queueing
	MessageQueueSize      int
	MessageQueueStrategy  QueueStrategy
	MaxMessageSize        int64
	WriteWait             time.Duration

	// Read settings
	ReadWait time.Duration

	// Metrics
	EnableMetrics bool
}

// QueueStrategy defines how to handle queue overflow
type QueueStrategy string

const (
	// QueueStrategyDropOldest drops the oldest message when queue is full
	QueueStrategyDropOldest QueueStrategy = "drop_oldest"

	// QueueStrategyDropNewest drops the newest message when queue is full
	QueueStrategyDropNewest QueueStrategy = "drop_newest"

	// QueueStrategyBlock blocks until space is available (default)
	QueueStrategyBlock QueueStrategy = "block"
)

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		// Connection limits (0 = unlimited)
		MaxConnectionsPerHub:  0,
		MaxConnectionsPerRoom: 0,

		// Heartbeat settings (enabled by default)
		EnableHeartbeat:   true,
		HeartbeatInterval: 30 * time.Second,
		HeartbeatTimeout:  90 * time.Second,
		PongWaitTimeout:   60 * time.Second,
		MaxMissedPongs:    3,

		// Reconnection settings (enabled by default)
		EnableReconnection:  true,
		ReconnectionTimeout: 30 * time.Second,
		MaxReconnectionTime: 5 * time.Minute,
		PreserveClientState: true,

		// Message queueing
		MessageQueueSize:     256,
		MessageQueueStrategy: QueueStrategyDropOldest,
		MaxMessageSize:       512 * 1024, // 512 KB
		WriteWait:            10 * time.Second,

		// Read settings
		ReadWait: 60 * time.Second,

		// Metrics (enabled by default)
		EnableMetrics: true,
	}
}

// ConnectionState represents stored connection state for reconnection
type ConnectionState struct {
	// Client ID for reconnection
	ClientID string

	// Last seen timestamp
	LastSeen time.Time

	// Custom data
	Data map[string]interface{}

	// Rooms the client was in
	Rooms []string

	// Buffered messages (optional)
	BufferedMessages [][]byte
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.HeartbeatInterval <= 0 {
		c.HeartbeatInterval = 30 * time.Second
	}

	if c.HeartbeatTimeout <= 0 {
		c.HeartbeatTimeout = 90 * time.Second
	}

	if c.PongWaitTimeout <= 0 {
		c.PongWaitTimeout = 60 * time.Second
	}

	if c.MaxMissedPongs <= 0 {
		c.MaxMissedPongs = 3
	}

	if c.MessageQueueSize <= 0 {
		c.MessageQueueSize = 256
	}

	if c.MaxMessageSize <= 0 {
		c.MaxMessageSize = 512 * 1024
	}

	if c.WriteWait <= 0 {
		c.WriteWait = 10 * time.Second
	}

	if c.ReadWait <= 0 {
		c.ReadWait = 60 * time.Second
	}

	if c.MessageQueueStrategy == "" {
		c.MessageQueueStrategy = QueueStrategyDropOldest
	}

	return nil
}
