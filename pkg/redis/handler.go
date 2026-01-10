package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Handler manages Redis connections and operations for the Glyph interpreter
type Handler struct {
	client Redis
	mu     sync.RWMutex
	ctx    context.Context
}

// NewHandler creates a new Redis handler
func NewHandler(client Redis) *Handler {
	return &Handler{
		client: client,
		ctx:    context.Background(),
	}
}

// NewHandlerFromString creates a handler from a connection string
func NewHandlerFromString(connStr string) (*Handler, error) {
	client, err := NewClientFromString(connStr)
	if err != nil {
		return nil, err
	}

	if err := client.Connect(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return NewHandler(client), nil
}

// Close closes the Redis connection
func (h *Handler) Close() error {
	return h.client.Close()
}

// Ping tests the Redis connection
func (h *Handler) Ping() error {
	return h.client.Ping(h.ctx)
}

// String operations

// Get retrieves the value of a key
// Returns nil if the key does not exist (Glyph-friendly)
func (h *Handler) Get(key string) (interface{}, error) {
	result, err := h.client.Get(h.ctx, key)
	if err != nil {
		if err == Nil {
			return nil, nil // Return nil for missing keys
		}
		return nil, err
	}
	return result, nil
}

// Set sets the value of a key with optional TTL in seconds
// Usage: redis.set("key", "value") or redis.set("key", "value", 300)
func (h *Handler) Set(key string, value interface{}, ttlSeconds ...int64) error {
	var ttl time.Duration
	if len(ttlSeconds) > 0 && ttlSeconds[0] > 0 {
		ttl = time.Duration(ttlSeconds[0]) * time.Second
	}
	return h.client.Set(h.ctx, key, value, ttl)
}

// SetNX sets the value only if key does not exist
// Returns true if the key was set, false if it already exists
func (h *Handler) SetNX(key string, value interface{}, ttlSeconds ...int64) (bool, error) {
	var ttl time.Duration
	if len(ttlSeconds) > 0 && ttlSeconds[0] > 0 {
		ttl = time.Duration(ttlSeconds[0]) * time.Second
	}
	return h.client.SetNX(h.ctx, key, value, ttl)
}

// Del deletes one or more keys
func (h *Handler) Del(keys ...string) (int64, error) {
	return h.client.Del(h.ctx, keys...)
}

// Exists checks if a key exists
func (h *Handler) Exists(key string) (bool, error) {
	count, err := h.client.Exists(h.ctx, key)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Expire sets a timeout on a key in seconds
func (h *Handler) Expire(key string, seconds int64) (bool, error) {
	return h.client.Expire(h.ctx, key, time.Duration(seconds)*time.Second)
}

// TTL returns the remaining time to live of a key in seconds
func (h *Handler) TTL(key string) (int64, error) {
	duration, err := h.client.TTL(h.ctx, key)
	if err != nil {
		return 0, err
	}
	return int64(duration.Seconds()), nil
}

// Atomic operations

// Incr increments the value of a key by 1
func (h *Handler) Incr(key string) (int64, error) {
	return h.client.Incr(h.ctx, key)
}

// IncrBy increments the value of a key by a specific amount
func (h *Handler) IncrBy(key string, value int64) (int64, error) {
	return h.client.IncrBy(h.ctx, key, value)
}

// Decr decrements the value of a key by 1
func (h *Handler) Decr(key string) (int64, error) {
	return h.client.Decr(h.ctx, key)
}

// DecrBy decrements the value of a key by a specific amount
func (h *Handler) DecrBy(key string, value int64) (int64, error) {
	return h.client.DecrBy(h.ctx, key, value)
}

// Hash operations

// HGet gets the value of a hash field
// Returns nil if the field does not exist
func (h *Handler) HGet(key, field string) (interface{}, error) {
	result, err := h.client.HGet(h.ctx, key, field)
	if err != nil {
		if err == Nil {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

// HSet sets field(s) in a hash
// Usage: redis.hset("key", "field", "value")
func (h *Handler) HSet(key, field string, value interface{}) (int64, error) {
	return h.client.HSet(h.ctx, key, field, value)
}

// HGetAll gets all fields and values in a hash
func (h *Handler) HGetAll(key string) (map[string]string, error) {
	return h.client.HGetAll(h.ctx, key)
}

// HDel deletes one or more hash fields
func (h *Handler) HDel(key string, fields ...string) (int64, error) {
	return h.client.HDel(h.ctx, key, fields...)
}

// HExists checks if a hash field exists
func (h *Handler) HExists(key, field string) (bool, error) {
	return h.client.HExists(h.ctx, key, field)
}

// List operations

// LPush prepends values to a list
func (h *Handler) LPush(key string, values ...interface{}) (int64, error) {
	return h.client.LPush(h.ctx, key, values...)
}

// RPush appends values to a list
func (h *Handler) RPush(key string, values ...interface{}) (int64, error) {
	return h.client.RPush(h.ctx, key, values...)
}

// LPop removes and returns the first element of a list
// Returns nil if the list is empty
func (h *Handler) LPop(key string) (interface{}, error) {
	result, err := h.client.LPop(h.ctx, key)
	if err != nil {
		if err == Nil {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

// RPop removes and returns the last element of a list
// Returns nil if the list is empty
func (h *Handler) RPop(key string) (interface{}, error) {
	result, err := h.client.RPop(h.ctx, key)
	if err != nil {
		if err == Nil {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

// LRange gets a range of elements from a list
func (h *Handler) LRange(key string, start, stop int64) ([]string, error) {
	return h.client.LRange(h.ctx, key, start, stop)
}

// LLen returns the length of a list
func (h *Handler) LLen(key string) (int64, error) {
	return h.client.LLen(h.ctx, key)
}

// Set operations

// SAdd adds members to a set
func (h *Handler) SAdd(key string, members ...interface{}) (int64, error) {
	return h.client.SAdd(h.ctx, key, members...)
}

// SRem removes members from a set
func (h *Handler) SRem(key string, members ...interface{}) (int64, error) {
	return h.client.SRem(h.ctx, key, members...)
}

// SMembers returns all members of a set
func (h *Handler) SMembers(key string) ([]string, error) {
	return h.client.SMembers(h.ctx, key)
}

// SIsMember checks if a value is a member of a set
func (h *Handler) SIsMember(key string, member interface{}) (bool, error) {
	return h.client.SIsMember(h.ctx, key, member)
}

// Pub/Sub

// Publish publishes a message to a channel
func (h *Handler) Publish(channel string, message interface{}) (int64, error) {
	return h.client.Publish(h.ctx, channel, message)
}

// Subscribe subscribes to channels and returns a subscription
func (h *Handler) Subscribe(channels ...string) (PubSub, error) {
	return h.client.Subscribe(h.ctx, channels...)
}

// JSON serialization helpers

// SetJSON serializes value as JSON and stores it
func (h *Handler) SetJSON(key string, value interface{}, ttlSeconds ...int64) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}
	return h.Set(key, string(data), ttlSeconds...)
}

// GetJSON retrieves a value and deserializes it from JSON
func (h *Handler) GetJSON(key string) (interface{}, error) {
	result, err := h.Get(key)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	var value interface{}
	if err := json.Unmarshal([]byte(result.(string)), &value); err != nil {
		return nil, fmt.Errorf("failed to deserialize value: %w", err)
	}
	return value, nil
}

// GetJSONInto retrieves a value and deserializes it into the provided destination
func (h *Handler) GetJSONInto(key string, dest interface{}) error {
	result, err := h.Get(key)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}

	if err := json.Unmarshal([]byte(result.(string)), dest); err != nil {
		return fmt.Errorf("failed to deserialize value: %w", err)
	}
	return nil
}

// Connection info

// Info returns Redis server information
func (h *Handler) Info() (string, error) {
	return h.client.Info(h.ctx)
}

// DBSize returns the number of keys in the database
func (h *Handler) DBSize() (int64, error) {
	return h.client.DBSize(h.ctx)
}

// MockHandler provides a mock Redis handler for testing
type MockHandler struct {
	data   map[string]interface{}
	hashes map[string]map[string]string
	lists  map[string][]string
	sets   map[string]map[string]bool
	mu     sync.RWMutex
}

// NewMockHandler creates a new mock Redis handler
func NewMockHandler() *MockHandler {
	return &MockHandler{
		data:   make(map[string]interface{}),
		hashes: make(map[string]map[string]string),
		lists:  make(map[string][]string),
		sets:   make(map[string]map[string]bool),
	}
}

// Get retrieves a value from the mock store
func (m *MockHandler) Get(key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, nil
}

// Set stores a value in the mock store
func (m *MockHandler) Set(key string, value interface{}, ttlSeconds ...int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

// Del deletes keys from the mock store
func (m *MockHandler) Del(keys ...string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	for _, key := range keys {
		if _, ok := m.data[key]; ok {
			delete(m.data, key)
			count++
		}
	}
	return count, nil
}

// Exists checks if a key exists in the mock store
func (m *MockHandler) Exists(key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok, nil
}

// Incr increments a value in the mock store
func (m *MockHandler) Incr(key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.data[key]
	if !ok {
		m.data[key] = int64(1)
		return 1, nil
	}
	if intVal, ok := val.(int64); ok {
		intVal++
		m.data[key] = intVal
		return intVal, nil
	}
	return 0, fmt.Errorf("value is not an integer")
}

// Decr decrements a value in the mock store
func (m *MockHandler) Decr(key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.data[key]
	if !ok {
		m.data[key] = int64(-1)
		return -1, nil
	}
	if intVal, ok := val.(int64); ok {
		intVal--
		m.data[key] = intVal
		return intVal, nil
	}
	return 0, fmt.Errorf("value is not an integer")
}

// Expire is a no-op for mock (always returns true)
func (m *MockHandler) Expire(key string, seconds int64) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok, nil
}

// TTL returns -1 for mock (no TTL tracking)
func (m *MockHandler) TTL(key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.data[key]; ok {
		return -1, nil // Key exists but no TTL
	}
	return -2, nil // Key does not exist
}

// SetNX sets a value only if the key does not exist
func (m *MockHandler) SetNX(key string, value interface{}, ttlSeconds ...int64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; exists {
		return false, nil
	}
	m.data[key] = value
	return true, nil
}

// IncrBy increments a value by a specific amount
func (m *MockHandler) IncrBy(key string, increment int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.data[key]
	if !ok {
		m.data[key] = increment
		return increment, nil
	}
	if intVal, ok := val.(int64); ok {
		intVal += increment
		m.data[key] = intVal
		return intVal, nil
	}
	return 0, fmt.Errorf("value is not an integer")
}

// DecrBy decrements a value by a specific amount
func (m *MockHandler) DecrBy(key string, decrement int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.data[key]
	if !ok {
		m.data[key] = -decrement
		return -decrement, nil
	}
	if intVal, ok := val.(int64); ok {
		intVal -= decrement
		m.data[key] = intVal
		return intVal, nil
	}
	return 0, fmt.Errorf("value is not an integer")
}

// Hash operations

// HGet gets a field from a hash
func (m *MockHandler) HGet(key, field string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if hash, ok := m.hashes[key]; ok {
		if val, ok := hash[field]; ok {
			return val, nil
		}
	}
	return nil, nil
}

// HSet sets a field in a hash
func (m *MockHandler) HSet(key, field string, value interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.hashes[key] == nil {
		m.hashes[key] = make(map[string]string)
	}
	_, exists := m.hashes[key][field]
	m.hashes[key][field] = fmt.Sprintf("%v", value)
	if exists {
		return 0, nil // Field was updated
	}
	return 1, nil // Field was created
}

// HGetAll gets all fields from a hash
func (m *MockHandler) HGetAll(key string) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if hash, ok := m.hashes[key]; ok {
		result := make(map[string]string)
		for k, v := range hash {
			result[k] = v
		}
		return result, nil
	}
	return make(map[string]string), nil
}

// HDel deletes fields from a hash
func (m *MockHandler) HDel(key string, fields ...string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	if hash, ok := m.hashes[key]; ok {
		for _, field := range fields {
			if _, exists := hash[field]; exists {
				delete(hash, field)
				count++
			}
		}
	}
	return count, nil
}

// HExists checks if a field exists in a hash
func (m *MockHandler) HExists(key, field string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if hash, ok := m.hashes[key]; ok {
		_, exists := hash[field]
		return exists, nil
	}
	return false, nil
}

// List operations

// LPush prepends values to a list
// In Redis, LPUSH key a b c pushes each value from left to right to the head
// resulting in [c, b, a] (last value ends up at the head)
func (m *MockHandler) LPush(key string, values ...interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, v := range values {
		m.lists[key] = append([]string{fmt.Sprintf("%v", v)}, m.lists[key]...)
	}
	return int64(len(m.lists[key])), nil
}

// RPush appends values to a list
func (m *MockHandler) RPush(key string, values ...interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, v := range values {
		m.lists[key] = append(m.lists[key], fmt.Sprintf("%v", v))
	}
	return int64(len(m.lists[key])), nil
}

// LPop removes and returns the first element
func (m *MockHandler) LPop(key string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if list, ok := m.lists[key]; ok && len(list) > 0 {
		val := list[0]
		m.lists[key] = list[1:]
		return val, nil
	}
	return nil, nil
}

// RPop removes and returns the last element
func (m *MockHandler) RPop(key string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if list, ok := m.lists[key]; ok && len(list) > 0 {
		val := list[len(list)-1]
		m.lists[key] = list[:len(list)-1]
		return val, nil
	}
	return nil, nil
}

// LRange gets a range of elements from a list
func (m *MockHandler) LRange(key string, start, stop int64) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list, ok := m.lists[key]
	if !ok || len(list) == 0 {
		return []string{}, nil
	}

	length := int64(len(list))

	// Handle negative indices
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	// Bounds checking
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop || start >= length {
		return []string{}, nil
	}

	return list[start : stop+1], nil
}

// LLen returns the length of a list
func (m *MockHandler) LLen(key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.lists[key])), nil
}

// Set operations

// SAdd adds members to a set
func (m *MockHandler) SAdd(key string, members ...interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sets[key] == nil {
		m.sets[key] = make(map[string]bool)
	}
	var added int64
	for _, member := range members {
		memberStr := fmt.Sprintf("%v", member)
		if !m.sets[key][memberStr] {
			m.sets[key][memberStr] = true
			added++
		}
	}
	return added, nil
}

// SRem removes members from a set
func (m *MockHandler) SRem(key string, members ...interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var removed int64
	if set, ok := m.sets[key]; ok {
		for _, member := range members {
			memberStr := fmt.Sprintf("%v", member)
			if set[memberStr] {
				delete(set, memberStr)
				removed++
			}
		}
	}
	return removed, nil
}

// SMembers returns all members of a set
func (m *MockHandler) SMembers(key string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var members []string
	if set, ok := m.sets[key]; ok {
		for member := range set {
			members = append(members, member)
		}
	}
	return members, nil
}

// SIsMember checks if a value is a member of a set
func (m *MockHandler) SIsMember(key string, member interface{}) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if set, ok := m.sets[key]; ok {
		return set[fmt.Sprintf("%v", member)], nil
	}
	return false, nil
}

// Publish is a no-op for mock (returns 0 subscribers)
func (m *MockHandler) Publish(channel string, message interface{}) (int64, error) {
	return 0, nil
}

// Close is a no-op for mock
func (m *MockHandler) Close() error {
	return nil
}
