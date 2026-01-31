// Package rabbitmq provides RabbitMQ integration for message publishing,
// consuming, exchange/queue management, dead letter handling, and retry logic.
// It uses a pluggable connection interface for testability.
package rabbitmq

import (
	"fmt"
	"sync"
	"time"
)

// ExchangeType represents AMQP exchange types
type ExchangeType string

const (
	Direct  ExchangeType = "direct"
	Topic   ExchangeType = "topic"
	Fanout  ExchangeType = "fanout"
	Headers ExchangeType = "headers"
)

// ExchangeConfig defines an exchange declaration
type ExchangeConfig struct {
	Name       string
	Type       ExchangeType
	Durable    bool
	AutoDelete bool
}

// QueueConfig defines a queue declaration
type QueueConfig struct {
	Name               string
	Durable            bool
	AutoDelete         bool
	Bindings           []Binding
	DeadLetterExchange string // Dead letter exchange name (empty if none)
	TTL                int    // Message TTL in milliseconds (0 for no TTL)
	MaxPriority        int    // Max priority level (0 for non-priority queue)
}

// Binding represents a queue-to-exchange binding
type Binding struct {
	Exchange   string
	RoutingKey string
}

// Message represents a message to publish or that was consumed
type Message struct {
	Exchange   string
	RoutingKey string
	Body       []byte
	Headers    map[string]interface{}
	Priority   int
	Expiration string // TTL for this message (empty for queue default)

	// Consumer fields (populated on delivery)
	DeliveryTag uint64
	Redelivered bool
}

// ConsumerConfig defines consumer settings
type ConsumerConfig struct {
	Queue       string
	Prefetch    int  // Number of messages to prefetch (0 = no limit)
	AutoAck     bool // Auto-acknowledge messages
	Concurrency int  // Number of concurrent consumers
	MaxRetries  int  // Maximum retry count before dead-lettering
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64 // Backoff multiplier (e.g., 2.0 for exponential)
}

// DefaultRetryPolicy returns a default retry policy with exponential backoff
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
}

// BackoffDuration calculates the backoff duration for a given retry attempt
func (p RetryPolicy) BackoffDuration(attempt int) time.Duration {
	if attempt <= 0 {
		return p.InitialBackoff
	}
	backoff := p.InitialBackoff
	for i := 0; i < attempt; i++ {
		backoff = time.Duration(float64(backoff) * p.Multiplier)
		if backoff > p.MaxBackoff {
			return p.MaxBackoff
		}
	}
	return backoff
}

// Connection defines the interface for RabbitMQ operations.
// This allows mocking for tests and swapping implementations.
type Connection interface {
	DeclareExchange(config ExchangeConfig) error
	DeclareQueue(config QueueConfig) error
	BindQueue(queue, exchange, routingKey string) error
	Publish(msg Message) error
	Consume(config ConsumerConfig, handler func(Message) error) error
	Ack(deliveryTag uint64) error
	Nack(deliveryTag uint64, requeue bool) error
	Reject(deliveryTag uint64, requeue bool) error
	Close() error
}

// Client is the main RabbitMQ client that manages connections, exchanges, and queues
type Client struct {
	mu          sync.RWMutex
	conn        Connection
	exchanges   map[string]ExchangeConfig
	queues      map[string]QueueConfig
	retryPolicy RetryPolicy
}

// NewClient creates a new RabbitMQ client with the given connection
func NewClient(conn Connection) *Client {
	return &Client{
		conn:        conn,
		exchanges:   make(map[string]ExchangeConfig),
		queues:      make(map[string]QueueConfig),
		retryPolicy: DefaultRetryPolicy(),
	}
}

// SetRetryPolicy sets the retry policy for message processing
func (c *Client) SetRetryPolicy(policy RetryPolicy) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.retryPolicy = policy
}

// DeclareExchange declares an exchange on the broker
func (c *Client) DeclareExchange(config ExchangeConfig) error {
	if config.Name == "" {
		return fmt.Errorf("exchange name is required")
	}
	if err := c.conn.DeclareExchange(config); err != nil {
		return fmt.Errorf("failed to declare exchange %s: %w", config.Name, err)
	}
	c.mu.Lock()
	c.exchanges[config.Name] = config
	c.mu.Unlock()
	return nil
}

// DeclareQueue declares a queue on the broker with optional bindings
func (c *Client) DeclareQueue(config QueueConfig) error {
	if config.Name == "" {
		return fmt.Errorf("queue name is required")
	}
	if err := c.conn.DeclareQueue(config); err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", config.Name, err)
	}
	// Set up bindings
	for _, binding := range config.Bindings {
		if err := c.conn.BindQueue(config.Name, binding.Exchange, binding.RoutingKey); err != nil {
			return fmt.Errorf("failed to bind queue %s to %s/%s: %w",
				config.Name, binding.Exchange, binding.RoutingKey, err)
		}
	}
	c.mu.Lock()
	c.queues[config.Name] = config
	c.mu.Unlock()
	return nil
}

// Publish publishes a message to an exchange with a routing key
func (c *Client) Publish(exchange, routingKey string, body []byte) error {
	return c.PublishMessage(Message{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Body:       body,
	})
}

// PublishMessage publishes a full message with headers and priority
func (c *Client) PublishMessage(msg Message) error {
	return c.conn.Publish(msg)
}

// Consume starts consuming messages from a queue with the given handler.
// If AutoAck is false, the handler must acknowledge messages.
func (c *Client) Consume(config ConsumerConfig, handler func(Message) error) error {
	if config.Queue == "" {
		return fmt.Errorf("queue name is required for consumption")
	}
	if config.Concurrency <= 0 {
		config.Concurrency = 1
	}
	return c.conn.Consume(config, handler)
}

// Ack acknowledges a message
func (c *Client) Ack(deliveryTag uint64) error {
	return c.conn.Ack(deliveryTag)
}

// Nack negatively acknowledges a message, optionally requeuing it
func (c *Client) Nack(deliveryTag uint64, requeue bool) error {
	return c.conn.Nack(deliveryTag, requeue)
}

// Reject rejects a message, optionally requeuing it
func (c *Client) Reject(deliveryTag uint64, requeue bool) error {
	return c.conn.Reject(deliveryTag, requeue)
}

// GetExchanges returns a copy of all declared exchanges
func (c *Client) GetExchanges() map[string]ExchangeConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]ExchangeConfig, len(c.exchanges))
	for k, v := range c.exchanges {
		result[k] = v
	}
	return result
}

// GetQueues returns a copy of all declared queues
func (c *Client) GetQueues() map[string]QueueConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]QueueConfig, len(c.queues))
	for k, v := range c.queues {
		result[k] = v
	}
	return result
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// MockConnection is a mock implementation of the Connection interface for testing
type MockConnection struct {
	mu        sync.Mutex
	Exchanges []ExchangeConfig
	Queues    []QueueConfig
	Bindings  []struct{ Queue, Exchange, RoutingKey string }
	Published []Message
	Acked     []uint64
	Nacked    []struct {
		Tag     uint64
		Requeue bool
	}
	Rejected []struct {
		Tag     uint64
		Requeue bool
	}
	ConsumeHandler func(Message) error
	Closed         bool
	// Error injection
	PublishErr error
	ConsumeErr error
}

// NewMockConnection creates a new mock connection
func NewMockConnection() *MockConnection {
	return &MockConnection{}
}

func (m *MockConnection) DeclareExchange(config ExchangeConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Exchanges = append(m.Exchanges, config)
	return nil
}

func (m *MockConnection) DeclareQueue(config QueueConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Queues = append(m.Queues, config)
	return nil
}

func (m *MockConnection) BindQueue(queue, exchange, routingKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Bindings = append(m.Bindings, struct{ Queue, Exchange, RoutingKey string }{queue, exchange, routingKey})
	return nil
}

func (m *MockConnection) Publish(msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.PublishErr != nil {
		return m.PublishErr
	}
	m.Published = append(m.Published, msg)
	return nil
}

func (m *MockConnection) Consume(config ConsumerConfig, handler func(Message) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ConsumeErr != nil {
		return m.ConsumeErr
	}
	m.ConsumeHandler = handler
	return nil
}

func (m *MockConnection) Ack(deliveryTag uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Acked = append(m.Acked, deliveryTag)
	return nil
}

func (m *MockConnection) Nack(deliveryTag uint64, requeue bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Nacked = append(m.Nacked, struct {
		Tag     uint64
		Requeue bool
	}{deliveryTag, requeue})
	return nil
}

func (m *MockConnection) Reject(deliveryTag uint64, requeue bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Rejected = append(m.Rejected, struct {
		Tag     uint64
		Requeue bool
	}{deliveryTag, requeue})
	return nil
}

func (m *MockConnection) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Closed = true
	return nil
}
