package redis

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Redis represents a generic Redis interface
type Redis interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	// String operations
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)
	Del(ctx context.Context, keys ...string) (int64, error)
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Atomic operations
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	DecrBy(ctx context.Context, key string, value int64) (int64, error)

	// Hash operations
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key string, values ...interface{}) (int64, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) (int64, error)
	HExists(ctx context.Context, key, field string) (bool, error)

	// List operations
	LPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	RPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	LLen(ctx context.Context, key string) (int64, error)

	// Set operations
	SAdd(ctx context.Context, key string, members ...interface{}) (int64, error)
	SRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)

	// Pub/Sub
	Publish(ctx context.Context, channel string, message interface{}) (int64, error)
	Subscribe(ctx context.Context, channels ...string) (PubSub, error)

	// Connection info
	Info(ctx context.Context) (string, error)
	DBSize(ctx context.Context) (int64, error)
}

// PubSub represents a pub/sub subscription
type PubSub interface {
	Receive(ctx context.Context) (*Message, error)
	Channel() <-chan *Message
	Close() error
}

// Message represents a pub/sub message
type Message struct {
	Channel string
	Payload string
}

// Config represents Redis configuration
type Config struct {
	// Connection settings
	Host     string
	Port     int
	Password string
	DB       int

	// TLS settings
	UseTLS        bool
	TLSSkipVerify bool

	// Connection pool settings
	PoolSize        int           // Maximum connections in pool
	MinIdleConns    int           // Minimum idle connections
	MaxIdleConns    int           // Maximum idle connections
	ConnMaxLifetime time.Duration // Maximum connection lifetime
	ConnMaxIdleTime time.Duration // Maximum idle time
	DialTimeout     time.Duration // Connection timeout
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration

	// Cluster mode settings
	ClusterMode  bool
	ClusterNodes []string // e.g., ["node1:6379", "node2:6379"]

	// Sentinel settings (HA)
	SentinelMode       bool
	SentinelAddrs      []string
	SentinelMasterName string
	SentinelPassword   string
}

// DefaultConfig returns sensible defaults for Redis connection
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            6379,
		DB:              0,
		PoolSize:        10,
		MinIdleConns:    2,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
	}
}

// ParseConnectionString parses a Redis connection string
// Format: redis://[:password@]host[:port][/db][?options]
func ParseConnectionString(connStr string) (*Config, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	// Validate scheme
	if u.Scheme != "redis" && u.Scheme != "rediss" {
		return nil, fmt.Errorf("invalid connection string: expected redis:// or rediss:// scheme")
	}

	config := DefaultConfig()

	// Use TLS for rediss:// scheme
	config.UseTLS = u.Scheme == "rediss"

	// Parse host
	if u.Hostname() != "" {
		config.Host = u.Hostname()
	}

	// Parse port
	if u.Port() != "" {
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
		config.Port = port
	}

	// Parse password
	if u.User != nil {
		if password, ok := u.User.Password(); ok {
			config.Password = password
		} else {
			// Username only means password (Redis auth format)
			config.Password = u.User.Username()
		}
	}

	// Parse database number from path
	if len(u.Path) > 1 {
		db, err := strconv.Atoi(u.Path[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid database number: %w", err)
		}
		config.DB = db
	}

	// Parse query parameters
	query := u.Query()
	if poolSize := query.Get("pool_size"); poolSize != "" {
		if size, err := strconv.Atoi(poolSize); err == nil {
			config.PoolSize = size
		}
	}
	if timeout := query.Get("dial_timeout"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.DialTimeout = d
		}
	}
	if timeout := query.Get("read_timeout"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.ReadTimeout = d
		}
	}
	if timeout := query.Get("write_timeout"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.WriteTimeout = d
		}
	}

	return config, nil
}

// ConnectionString generates a connection string from config
func (c *Config) ConnectionString() string {
	scheme := "redis"
	if c.UseTLS {
		scheme = "rediss"
	}

	auth := ""
	if c.Password != "" {
		auth = ":" + c.Password + "@"
	}

	return fmt.Sprintf("%s://%s%s:%d/%d", scheme, auth, c.Host, c.Port, c.DB)
}

// Address returns the host:port address
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// HealthCheck performs a health check on the Redis connection
func HealthCheck(ctx context.Context, r Redis) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := r.Ping(ctx); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	return nil
}
