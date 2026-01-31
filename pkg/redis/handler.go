package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Handler manages Redis connections and operations for the interpreter.
// It wraps a go-redis client and exposes methods that can be called
// via reflection from GlyphLang's dependency injection system.
type Handler struct {
	client *goredis.Client
	ctx    context.Context
}

// NewHandler creates a new Redis handler from a go-redis client.
func NewHandler(client *goredis.Client) *Handler {
	return &Handler{
		client: client,
		ctx:    context.Background(),
	}
}

// NewHandlerFromURL creates a new handler from a Redis URL (e.g., "redis://localhost:6379/0").
func NewHandlerFromURL(redisURL string) (*Handler, error) {
	opts, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}
	client := goredis.NewClient(opts)
	return NewHandler(client), nil
}

// NewHandlerFromOptions creates a new handler from go-redis options.
func NewHandlerFromOptions(opts *goredis.Options) *Handler {
	client := goredis.NewClient(opts)
	return NewHandler(client)
}

// Close closes the Redis connection.
func (h *Handler) Close() error {
	return h.client.Close()
}

// Ping tests the Redis connection.
func (h *Handler) Ping() (string, error) {
	return h.client.Ping(h.ctx).Result()
}

// --- String operations ---

// Get retrieves the value for a key. Returns nil if the key does not exist.
func (h *Handler) Get(key string) (interface{}, error) {
	val, err := h.client.Get(h.ctx, key).Result()
	if err == goredis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Set stores a key-value pair. Accepts an optional TTL in seconds.
func (h *Handler) Set(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("redis.set requires at least key and value arguments")
	}

	key, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("redis.set: key must be a string")
	}
	value := args[1]

	var ttl time.Duration
	if len(args) >= 3 {
		switch v := args[2].(type) {
		case int64:
			ttl = time.Duration(v) * time.Second
		case int:
			ttl = time.Duration(v) * time.Second
		case float64:
			ttl = time.Duration(v) * time.Second
		default:
			return nil, fmt.Errorf("redis.set: ttl must be a number (seconds)")
		}
	}

	err := h.client.Set(h.ctx, key, value, ttl).Err()
	if err != nil {
		return nil, err
	}
	return "OK", nil
}

// Del deletes one or more keys. Returns the number of keys removed.
func (h *Handler) Del(keys ...interface{}) (int64, error) {
	strKeys := make([]string, len(keys))
	for i, k := range keys {
		s, ok := k.(string)
		if !ok {
			return 0, fmt.Errorf("redis.del: key must be a string")
		}
		strKeys[i] = s
	}
	return h.client.Del(h.ctx, strKeys...).Result()
}

// Exists checks if one or more keys exist. Returns the count of existing keys.
func (h *Handler) Exists(keys ...interface{}) (int64, error) {
	strKeys := make([]string, len(keys))
	for i, k := range keys {
		s, ok := k.(string)
		if !ok {
			return 0, fmt.Errorf("redis.exists: key must be a string")
		}
		strKeys[i] = s
	}
	return h.client.Exists(h.ctx, strKeys...).Result()
}

// Expire sets a TTL (in seconds) on a key.
func (h *Handler) Expire(key string, seconds interface{}) (bool, error) {
	var dur time.Duration
	switch v := seconds.(type) {
	case int64:
		dur = time.Duration(v) * time.Second
	case int:
		dur = time.Duration(v) * time.Second
	case float64:
		dur = time.Duration(v) * time.Second
	default:
		return false, fmt.Errorf("redis.expire: seconds must be a number")
	}
	return h.client.Expire(h.ctx, key, dur).Result()
}

// Ttl returns the remaining TTL of a key in seconds. Returns -1 if no TTL, -2 if key doesn't exist.
func (h *Handler) Ttl(key string) (int64, error) {
	dur, err := h.client.TTL(h.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int64(dur.Seconds()), nil
}

// Incr atomically increments a key's integer value by 1.
func (h *Handler) Incr(key string) (int64, error) {
	return h.client.Incr(h.ctx, key).Result()
}

// Decr atomically decrements a key's integer value by 1.
func (h *Handler) Decr(key string) (int64, error) {
	return h.client.Decr(h.ctx, key).Result()
}

// --- Hash operations ---

// HGet retrieves the value of a field in a hash.
func (h *Handler) HGet(key, field string) (interface{}, error) {
	val, err := h.client.HGet(h.ctx, key, field).Result()
	if err == goredis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

// HSet sets one or more fields in a hash. Args: key, field, value [, field, value ...]
func (h *Handler) HSet(args ...interface{}) (int64, error) {
	if len(args) < 3 {
		return 0, fmt.Errorf("redis.hset requires at least key, field, and value arguments")
	}
	key, ok := args[0].(string)
	if !ok {
		return 0, fmt.Errorf("redis.hset: key must be a string")
	}
	return h.client.HSet(h.ctx, key, args[1:]...).Result()
}

// HDel deletes one or more fields from a hash.
func (h *Handler) HDel(key string, fields ...interface{}) (int64, error) {
	strFields := make([]string, len(fields))
	for i, f := range fields {
		s, ok := f.(string)
		if !ok {
			return 0, fmt.Errorf("redis.hdel: field must be a string")
		}
		strFields[i] = s
	}
	return h.client.HDel(h.ctx, key, strFields...).Result()
}

// HGetAll retrieves all fields and values of a hash as a map.
func (h *Handler) HGetAll(key string) (map[string]interface{}, error) {
	result, err := h.client.HGetAll(h.ctx, key).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string]interface{}, len(result))
	for k, v := range result {
		out[k] = v
	}
	return out, nil
}

// HExists checks if a field exists in a hash.
func (h *Handler) HExists(key, field string) (bool, error) {
	return h.client.HExists(h.ctx, key, field).Result()
}

// --- List operations ---

// LPush prepends one or more values to a list.
func (h *Handler) LPush(key string, values ...interface{}) (int64, error) {
	return h.client.LPush(h.ctx, key, values...).Result()
}

// RPush appends one or more values to a list.
func (h *Handler) RPush(key string, values ...interface{}) (int64, error) {
	return h.client.RPush(h.ctx, key, values...).Result()
}

// LPop removes and returns the first element of a list.
func (h *Handler) LPop(key string) (interface{}, error) {
	val, err := h.client.LPop(h.ctx, key).Result()
	if err == goredis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

// RPop removes and returns the last element of a list.
func (h *Handler) RPop(key string) (interface{}, error) {
	val, err := h.client.RPop(h.ctx, key).Result()
	if err == goredis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

// LLen returns the length of a list.
func (h *Handler) LLen(key string) (int64, error) {
	return h.client.LLen(h.ctx, key).Result()
}

// LRange returns a range of elements from a list.
func (h *Handler) LRange(key string, start, stop int64) ([]interface{}, error) {
	result, err := h.client.LRange(h.ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	out := make([]interface{}, len(result))
	for i, v := range result {
		out[i] = v
	}
	return out, nil
}

// --- Set operations ---

// SAdd adds one or more members to a set.
func (h *Handler) SAdd(key string, members ...interface{}) (int64, error) {
	return h.client.SAdd(h.ctx, key, members...).Result()
}

// SRem removes one or more members from a set.
func (h *Handler) SRem(key string, members ...interface{}) (int64, error) {
	return h.client.SRem(h.ctx, key, members...).Result()
}

// SMembers returns all members of a set.
func (h *Handler) SMembers(key string) ([]interface{}, error) {
	result, err := h.client.SMembers(h.ctx, key).Result()
	if err != nil {
		return nil, err
	}
	out := make([]interface{}, len(result))
	for i, v := range result {
		out[i] = v
	}
	return out, nil
}

// SIsMember checks if a value is a member of a set.
func (h *Handler) SIsMember(key string, member interface{}) (bool, error) {
	return h.client.SIsMember(h.ctx, key, member).Result()
}

// --- Pub/Sub operations ---

// Publish sends a message to a channel. Returns the number of clients that received the message.
func (h *Handler) Publish(channel string, message interface{}) (int64, error) {
	return h.client.Publish(h.ctx, channel, message).Result()
}

// --- Key operations ---

// Keys returns all keys matching a pattern.
func (h *Handler) Keys(pattern string) ([]interface{}, error) {
	result, err := h.client.Keys(h.ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	out := make([]interface{}, len(result))
	for i, v := range result {
		out[i] = v
	}
	return out, nil
}

// FlushAll removes all keys from all databases.
func (h *Handler) FlushAll() (string, error) {
	return h.client.FlushAll(h.ctx).Result()
}

// Client returns the underlying go-redis client for advanced usage.
func (h *Handler) Client() *goredis.Client {
	return h.client
}
