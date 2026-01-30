package rabbitmq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient() (*Client, *MockConnection) {
	mock := NewMockConnection()
	client := NewClient(mock)
	return client, mock
}

func TestNewClient(t *testing.T) {
	client, _ := newTestClient()
	assert.NotNil(t, client)
	assert.Empty(t, client.GetExchanges())
	assert.Empty(t, client.GetQueues())
}

func TestDeclareExchange(t *testing.T) {
	client, mock := newTestClient()
	err := client.DeclareExchange(ExchangeConfig{
		Name: "orders", Type: Topic, Durable: true,
	})
	require.NoError(t, err)
	assert.Len(t, mock.Exchanges, 1)
	assert.Equal(t, "orders", mock.Exchanges[0].Name)
	assert.Equal(t, Topic, mock.Exchanges[0].Type)
	exchanges := client.GetExchanges()
	assert.Len(t, exchanges, 1)
}

func TestDeclareExchangeEmptyName(t *testing.T) {
	client, _ := newTestClient()
	err := client.DeclareExchange(ExchangeConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestDeclareQueue(t *testing.T) {
	client, mock := newTestClient()
	err := client.DeclareQueue(QueueConfig{
		Name: "order.processing", Durable: true,
	})
	require.NoError(t, err)
	assert.Len(t, mock.Queues, 1)
	assert.Equal(t, "order.processing", mock.Queues[0].Name)
	queues := client.GetQueues()
	assert.Len(t, queues, 1)
}

func TestDeclareQueueEmptyName(t *testing.T) {
	client, _ := newTestClient()
	err := client.DeclareQueue(QueueConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestDeclareQueueWithBindings(t *testing.T) {
	client, mock := newTestClient()
	err := client.DeclareQueue(QueueConfig{
		Name:    "order.processing",
		Durable: true,
		Bindings: []Binding{
			{Exchange: "orders", RoutingKey: "order.created"},
			{Exchange: "orders", RoutingKey: "order.updated"},
		},
	})
	require.NoError(t, err)
	assert.Len(t, mock.Bindings, 2)
	assert.Equal(t, "orders", mock.Bindings[0].Exchange)
	assert.Equal(t, "order.created", mock.Bindings[0].RoutingKey)
}

func TestPublish(t *testing.T) {
	client, mock := newTestClient()
	err := client.Publish("orders", "order.created", []byte(`{"id": 1}`))
	require.NoError(t, err)
	require.Len(t, mock.Published, 1)
	assert.Equal(t, "orders", mock.Published[0].Exchange)
	assert.Equal(t, "order.created", mock.Published[0].RoutingKey)
	assert.Equal(t, []byte(`{"id": 1}`), mock.Published[0].Body)
}

func TestPublishMessage(t *testing.T) {
	client, mock := newTestClient()
	err := client.PublishMessage(Message{
		Exchange:   "orders",
		RoutingKey: "order.created",
		Body:       []byte(`{"id": 1}`),
		Headers:    map[string]interface{}{"x-retry": 0},
		Priority:   5,
	})
	require.NoError(t, err)
	require.Len(t, mock.Published, 1)
	assert.Equal(t, 5, mock.Published[0].Priority)
}

func TestConsume(t *testing.T) {
	client, mock := newTestClient()
	called := false
	err := client.Consume(ConsumerConfig{
		Queue: "order.processing", Prefetch: 10, Concurrency: 5,
	}, func(msg Message) error {
		called = true
		return nil
	})
	require.NoError(t, err)
	require.NotNil(t, mock.ConsumeHandler)

	// Simulate message delivery
	mock.ConsumeHandler(Message{Body: []byte("test")})
	assert.True(t, called)
}

func TestConsumeEmptyQueue(t *testing.T) {
	client, _ := newTestClient()
	err := client.Consume(ConsumerConfig{}, func(msg Message) error { return nil })
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue name is required")
}

func TestAck(t *testing.T) {
	client, mock := newTestClient()
	err := client.Ack(42)
	require.NoError(t, err)
	require.Len(t, mock.Acked, 1)
	assert.Equal(t, uint64(42), mock.Acked[0])
}

func TestNack(t *testing.T) {
	client, mock := newTestClient()
	err := client.Nack(42, true)
	require.NoError(t, err)
	require.Len(t, mock.Nacked, 1)
	assert.Equal(t, uint64(42), mock.Nacked[0].Tag)
	assert.True(t, mock.Nacked[0].Requeue)
}

func TestReject(t *testing.T) {
	client, mock := newTestClient()
	err := client.Reject(42, false)
	require.NoError(t, err)
	require.Len(t, mock.Rejected, 1)
	assert.Equal(t, uint64(42), mock.Rejected[0].Tag)
	assert.False(t, mock.Rejected[0].Requeue)
}

func TestClose(t *testing.T) {
	client, mock := newTestClient()
	err := client.Close()
	require.NoError(t, err)
	assert.True(t, mock.Closed)
}

func TestExchangeTypes(t *testing.T) {
	assert.Equal(t, ExchangeType("direct"), Direct)
	assert.Equal(t, ExchangeType("topic"), Topic)
	assert.Equal(t, ExchangeType("fanout"), Fanout)
	assert.Equal(t, ExchangeType("headers"), Headers)
}

func TestDefaultRetryPolicy(t *testing.T) {
	policy := DefaultRetryPolicy()
	assert.Equal(t, 3, policy.MaxRetries)
	assert.Equal(t, 1*time.Second, policy.InitialBackoff)
	assert.Equal(t, 30*time.Second, policy.MaxBackoff)
	assert.Equal(t, 2.0, policy.Multiplier)
}

func TestBackoffDuration(t *testing.T) {
	policy := RetryPolicy{
		MaxRetries:     5,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
	assert.Equal(t, 1*time.Second, policy.BackoffDuration(0))
	assert.Equal(t, 2*time.Second, policy.BackoffDuration(1))
	assert.Equal(t, 4*time.Second, policy.BackoffDuration(2))
	assert.Equal(t, 8*time.Second, policy.BackoffDuration(3))
	// Should cap at MaxBackoff
	assert.Equal(t, 30*time.Second, policy.BackoffDuration(10))
}

func TestSetRetryPolicy(t *testing.T) {
	client, _ := newTestClient()
	custom := RetryPolicy{MaxRetries: 10, InitialBackoff: 500 * time.Millisecond, MaxBackoff: 60 * time.Second, Multiplier: 3.0}
	client.SetRetryPolicy(custom)
	// Verify through backoff duration
	assert.Equal(t, 500*time.Millisecond, client.retryPolicy.InitialBackoff)
}

func TestDeadLetterQueue(t *testing.T) {
	client, mock := newTestClient()
	err := client.DeclareQueue(QueueConfig{
		Name:               "order.processing",
		Durable:            true,
		DeadLetterExchange: "orders.dlx",
	})
	require.NoError(t, err)
	assert.Equal(t, "orders.dlx", mock.Queues[0].DeadLetterExchange)
}

func TestPriorityQueue(t *testing.T) {
	client, mock := newTestClient()
	err := client.DeclareQueue(QueueConfig{
		Name:        "priority.tasks",
		MaxPriority: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 10, mock.Queues[0].MaxPriority)
}

func TestQueueTTL(t *testing.T) {
	client, mock := newTestClient()
	err := client.DeclareQueue(QueueConfig{
		Name: "temporary.events",
		TTL:  60000, // 60 seconds
	})
	require.NoError(t, err)
	assert.Equal(t, 60000, mock.Queues[0].TTL)
}
