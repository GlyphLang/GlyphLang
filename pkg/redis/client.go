package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Nil is returned when key does not exist
var Nil = redis.Nil

// Client implements the Redis interface using go-redis
type Client struct {
	config *Config
	client redis.UniversalClient
}

// NewClient creates a new Redis client
func NewClient(config *Config) *Client {
	return &Client{
		config: config,
	}
}

// NewClientFromString creates a new Redis client from a connection string
func NewClientFromString(connStr string) (*Client, error) {
	config, err := ParseConnectionString(connStr)
	if err != nil {
		return nil, err
	}
	return NewClient(config), nil
}

// Connect establishes the connection to Redis
func (c *Client) Connect(ctx context.Context) error {
	var client redis.UniversalClient

	if c.config.ClusterMode {
		// Cluster mode
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:           c.config.ClusterNodes,
			Password:        c.config.Password,
			PoolSize:        c.config.PoolSize,
			MinIdleConns:    c.config.MinIdleConns,
			ConnMaxLifetime: c.config.ConnMaxLifetime,
			ConnMaxIdleTime: c.config.ConnMaxIdleTime,
			DialTimeout:     c.config.DialTimeout,
			ReadTimeout:     c.config.ReadTimeout,
			WriteTimeout:    c.config.WriteTimeout,
			TLSConfig:       c.getTLSConfig(),
		})
	} else if c.config.SentinelMode {
		// Sentinel mode for HA
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       c.config.SentinelMasterName,
			SentinelAddrs:    c.config.SentinelAddrs,
			SentinelPassword: c.config.SentinelPassword,
			Password:         c.config.Password,
			DB:               c.config.DB,
			PoolSize:         c.config.PoolSize,
			MinIdleConns:     c.config.MinIdleConns,
			ConnMaxLifetime:  c.config.ConnMaxLifetime,
			ConnMaxIdleTime:  c.config.ConnMaxIdleTime,
			DialTimeout:      c.config.DialTimeout,
			ReadTimeout:      c.config.ReadTimeout,
			WriteTimeout:     c.config.WriteTimeout,
			TLSConfig:        c.getTLSConfig(),
		})
	} else {
		// Standard single-node mode
		client = redis.NewClient(&redis.Options{
			Addr:            c.config.Address(),
			Password:        c.config.Password,
			DB:              c.config.DB,
			PoolSize:        c.config.PoolSize,
			MinIdleConns:    c.config.MinIdleConns,
			ConnMaxLifetime: c.config.ConnMaxLifetime,
			ConnMaxIdleTime: c.config.ConnMaxIdleTime,
			DialTimeout:     c.config.DialTimeout,
			ReadTimeout:     c.config.ReadTimeout,
			WriteTimeout:    c.config.WriteTimeout,
			TLSConfig:       c.getTLSConfig(),
		})
	}

	// Test the connection
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	c.client = client
	return nil
}

// getTLSConfig returns TLS configuration if enabled
func (c *Client) getTLSConfig() *tls.Config {
	if !c.config.UseTLS {
		return nil
	}
	return &tls.Config{
		InsecureSkipVerify: c.config.TLSSkipVerify,
	}
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Ping tests the Redis connection
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// String operations

// Get retrieves the value of a key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set sets the value of a key with optional TTL
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// SetNX sets the value only if key does not exist
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, ttl).Result()
}

// Del deletes one or more keys
func (c *Client) Del(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Del(ctx, keys...).Result()
}

// Exists returns the number of keys that exist
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// Expire sets a timeout on a key
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.client.Expire(ctx, key, ttl).Result()
}

// TTL returns the remaining time to live of a key
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Atomic operations

// Incr increments the value of a key by 1
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// IncrBy increments the value of a key by a specific amount
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Decr decrements the value of a key by 1
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// DecrBy decrements the value of a key by a specific amount
func (c *Client) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, key, value).Result()
}

// Hash operations

// HGet gets the value of a hash field
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// HSet sets field(s) in a hash
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.client.HSet(ctx, key, values...).Result()
}

// HGetAll gets all fields and values in a hash
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HDel deletes one or more hash fields
func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return c.client.HDel(ctx, key, fields...).Result()
}

// HExists checks if a hash field exists
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	return c.client.HExists(ctx, key, field).Result()
}

// List operations

// LPush prepends values to a list
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.client.LPush(ctx, key, values...).Result()
}

// RPush appends values to a list
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.client.RPush(ctx, key, values...).Result()
}

// LPop removes and returns the first element of a list
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	return c.client.LPop(ctx, key).Result()
}

// RPop removes and returns the last element of a list
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	return c.client.RPop(ctx, key).Result()
}

// LRange gets a range of elements from a list
func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}

// LLen returns the length of a list
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

// Set operations

// SAdd adds members to a set
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.client.SAdd(ctx, key, members...).Result()
}

// SRem removes members from a set
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.client.SRem(ctx, key, members...).Result()
}

// SMembers returns all members of a set
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// SIsMember checks if a value is a member of a set
func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}

// Pub/Sub

// Publish publishes a message to a channel
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	return c.client.Publish(ctx, channel, message).Result()
}

// Subscribe subscribes to channels
func (c *Client) Subscribe(ctx context.Context, channels ...string) (PubSub, error) {
	pubsub := c.client.Subscribe(ctx, channels...)
	// Wait for subscription to be confirmed
	_, err := pubsub.Receive(ctx)
	if err != nil {
		pubsub.Close()
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}
	return &pubSubWrapper{pubsub: pubsub}, nil
}

// Info returns Redis server information
func (c *Client) Info(ctx context.Context) (string, error) {
	return c.client.Info(ctx).Result()
}

// DBSize returns the number of keys in the database
func (c *Client) DBSize(ctx context.Context) (int64, error) {
	return c.client.DBSize(ctx).Result()
}

// pubSubWrapper wraps go-redis PubSub to implement our PubSub interface
type pubSubWrapper struct {
	pubsub *redis.PubSub
}

// Receive receives a message from the subscription
func (p *pubSubWrapper) Receive(ctx context.Context) (*Message, error) {
	msg, err := p.pubsub.ReceiveMessage(ctx)
	if err != nil {
		return nil, err
	}
	return &Message{
		Channel: msg.Channel,
		Payload: msg.Payload,
	}, nil
}

// Channel returns a channel for receiving messages
func (p *pubSubWrapper) Channel() <-chan *Message {
	ch := make(chan *Message)
	go func() {
		defer close(ch)
		for msg := range p.pubsub.Channel() {
			ch <- &Message{
				Channel: msg.Channel,
				Payload: msg.Payload,
			}
		}
	}()
	return ch
}

// Close closes the subscription
func (p *pubSubWrapper) Close() error {
	return p.pubsub.Close()
}
