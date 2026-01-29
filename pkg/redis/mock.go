package redis

import (
	"fmt"
	"sync"
	"time"
)

// MockHandler provides an in-memory Redis mock for testing without a real Redis server.
// It implements the same public methods as Handler so it can be used via reflection.
type MockHandler struct {
	mu     sync.RWMutex
	data   map[string]mockEntry
	lists  map[string][]string
	hashes map[string]map[string]string
	sets   map[string]map[string]struct{}
}

type mockEntry struct {
	value     string
	expiresAt time.Time // zero value means no expiry
}

func (e mockEntry) isExpired() bool {
	if e.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.expiresAt)
}

// NewMockHandler creates a new mock Redis handler.
func NewMockHandler() *MockHandler {
	return &MockHandler{
		data:   make(map[string]mockEntry),
		lists:  make(map[string][]string),
		hashes: make(map[string]map[string]string),
		sets:   make(map[string]map[string]struct{}),
	}
}

// Ping always returns "PONG".
func (m *MockHandler) Ping() (string, error) {
	return "PONG", nil
}

// --- String operations ---

func (m *MockHandler) Get(key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, ok := m.data[key]
	if !ok || entry.isExpired() {
		return nil, nil
	}
	return entry.value, nil
}

func (m *MockHandler) Set(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("redis.set requires at least key and value arguments")
	}

	key, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("redis.set: key must be a string")
	}
	value := fmt.Sprintf("%v", args[1])

	var expiresAt time.Time
	if len(args) >= 3 {
		switch v := args[2].(type) {
		case int64:
			expiresAt = time.Now().Add(time.Duration(v) * time.Second)
		case int:
			expiresAt = time.Now().Add(time.Duration(v) * time.Second)
		case float64:
			expiresAt = time.Now().Add(time.Duration(v) * time.Second)
		default:
			return nil, fmt.Errorf("redis.set: ttl must be a number (seconds)")
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = mockEntry{value: value, expiresAt: expiresAt}
	return "OK", nil
}

func (m *MockHandler) Del(keys ...interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var count int64
	for _, k := range keys {
		key, ok := k.(string)
		if !ok {
			return count, fmt.Errorf("redis.del: key must be a string")
		}
		if _, exists := m.data[key]; exists {
			delete(m.data, key)
			count++
		}
		if _, exists := m.lists[key]; exists {
			delete(m.lists, key)
			count++
		}
		if _, exists := m.hashes[key]; exists {
			delete(m.hashes, key)
			count++
		}
		if _, exists := m.sets[key]; exists {
			delete(m.sets, key)
			count++
		}
	}
	return count, nil
}

func (m *MockHandler) Exists(keys ...interface{}) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int64
	for _, k := range keys {
		key, ok := k.(string)
		if !ok {
			return count, fmt.Errorf("redis.exists: key must be a string")
		}
		if entry, exists := m.data[key]; exists && !entry.isExpired() {
			count++
		} else if _, exists := m.lists[key]; exists {
			count++
		} else if _, exists := m.hashes[key]; exists {
			count++
		} else if _, exists := m.sets[key]; exists {
			count++
		}
	}
	return count, nil
}

func (m *MockHandler) Expire(key string, seconds interface{}) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.data[key]
	if !ok {
		return false, nil
	}

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

	entry.expiresAt = time.Now().Add(dur)
	m.data[key] = entry
	return true, nil
}

func (m *MockHandler) Ttl(key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, ok := m.data[key]
	if !ok {
		return -2, nil
	}
	if entry.expiresAt.IsZero() {
		return -1, nil
	}
	remaining := time.Until(entry.expiresAt)
	if remaining <= 0 {
		return -2, nil
	}
	return int64(remaining.Seconds()), nil
}

func (m *MockHandler) Incr(key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.data[key]
	if !ok || entry.isExpired() {
		m.data[key] = mockEntry{value: "1"}
		return 1, nil
	}

	var val int64
	_, err := fmt.Sscanf(entry.value, "%d", &val)
	if err != nil {
		return 0, fmt.Errorf("redis.incr: value is not an integer")
	}
	val++
	entry.value = fmt.Sprintf("%d", val)
	m.data[key] = entry
	return val, nil
}

func (m *MockHandler) Decr(key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.data[key]
	if !ok || entry.isExpired() {
		m.data[key] = mockEntry{value: "-1"}
		return -1, nil
	}

	var val int64
	_, err := fmt.Sscanf(entry.value, "%d", &val)
	if err != nil {
		return 0, fmt.Errorf("redis.decr: value is not an integer")
	}
	val--
	entry.value = fmt.Sprintf("%d", val)
	m.data[key] = entry
	return val, nil
}

// --- Hash operations ---

func (m *MockHandler) HGet(key, field string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hash, ok := m.hashes[key]
	if !ok {
		return nil, nil
	}
	val, ok := hash[field]
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *MockHandler) HSet(args ...interface{}) (int64, error) {
	if len(args) < 3 {
		return 0, fmt.Errorf("redis.hset requires at least key, field, and value arguments")
	}
	key, ok := args[0].(string)
	if !ok {
		return 0, fmt.Errorf("redis.hset: key must be a string")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.hashes[key]; !ok {
		m.hashes[key] = make(map[string]string)
	}

	var added int64
	for i := 1; i+1 < len(args); i += 2 {
		field := fmt.Sprintf("%v", args[i])
		value := fmt.Sprintf("%v", args[i+1])
		if _, exists := m.hashes[key][field]; !exists {
			added++
		}
		m.hashes[key][field] = value
	}
	return added, nil
}

func (m *MockHandler) HDel(key string, fields ...interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hash, ok := m.hashes[key]
	if !ok {
		return 0, nil
	}

	var count int64
	for _, f := range fields {
		field := fmt.Sprintf("%v", f)
		if _, exists := hash[field]; exists {
			delete(hash, field)
			count++
		}
	}
	return count, nil
}

func (m *MockHandler) HGetAll(key string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hash, ok := m.hashes[key]
	if !ok {
		return map[string]interface{}{}, nil
	}

	out := make(map[string]interface{}, len(hash))
	for k, v := range hash {
		out[k] = v
	}
	return out, nil
}

func (m *MockHandler) HExists(key, field string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hash, ok := m.hashes[key]
	if !ok {
		return false, nil
	}
	_, exists := hash[field]
	return exists, nil
}

// --- List operations ---

func (m *MockHandler) LPush(key string, values ...interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, v := range values {
		m.lists[key] = append([]string{fmt.Sprintf("%v", v)}, m.lists[key]...)
	}
	return int64(len(m.lists[key])), nil
}

func (m *MockHandler) RPush(key string, values ...interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, v := range values {
		m.lists[key] = append(m.lists[key], fmt.Sprintf("%v", v))
	}
	return int64(len(m.lists[key])), nil
}

func (m *MockHandler) LPop(key string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.lists[key]
	if !ok || len(list) == 0 {
		return nil, nil
	}
	val := list[0]
	m.lists[key] = list[1:]
	return val, nil
}

func (m *MockHandler) RPop(key string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.lists[key]
	if !ok || len(list) == 0 {
		return nil, nil
	}
	val := list[len(list)-1]
	m.lists[key] = list[:len(list)-1]
	return val, nil
}

func (m *MockHandler) LLen(key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return int64(len(m.lists[key])), nil
}

func (m *MockHandler) LRange(key string, start, stop int64) ([]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := m.lists[key]
	length := int64(len(list))
	if length == 0 {
		return []interface{}{}, nil
	}

	// Normalize negative indices
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop {
		return []interface{}{}, nil
	}

	out := make([]interface{}, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		out = append(out, list[i])
	}
	return out, nil
}

// --- Set operations ---

func (m *MockHandler) SAdd(key string, members ...interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sets[key]; !ok {
		m.sets[key] = make(map[string]struct{})
	}

	var added int64
	for _, member := range members {
		s := fmt.Sprintf("%v", member)
		if _, exists := m.sets[key][s]; !exists {
			m.sets[key][s] = struct{}{}
			added++
		}
	}
	return added, nil
}

func (m *MockHandler) SRem(key string, members ...interface{}) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	set, ok := m.sets[key]
	if !ok {
		return 0, nil
	}

	var removed int64
	for _, member := range members {
		s := fmt.Sprintf("%v", member)
		if _, exists := set[s]; exists {
			delete(set, s)
			removed++
		}
	}
	return removed, nil
}

func (m *MockHandler) SMembers(key string) ([]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	set, ok := m.sets[key]
	if !ok {
		return []interface{}{}, nil
	}

	out := make([]interface{}, 0, len(set))
	for member := range set {
		out = append(out, member)
	}
	return out, nil
}

func (m *MockHandler) SIsMember(key string, member interface{}) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	set, ok := m.sets[key]
	if !ok {
		return false, nil
	}
	_, exists := set[fmt.Sprintf("%v", member)]
	return exists, nil
}

// --- Pub/Sub (no-op for mock) ---

func (m *MockHandler) Publish(channel string, message interface{}) (int64, error) {
	return 0, nil
}

// --- Key operations ---

func (m *MockHandler) Keys(pattern string) ([]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Simple pattern matching: only support "*" (all keys)
	out := make([]interface{}, 0)
	for key, entry := range m.data {
		if !entry.isExpired() {
			out = append(out, key)
		}
	}
	for key := range m.lists {
		out = append(out, key)
	}
	for key := range m.hashes {
		out = append(out, key)
	}
	for key := range m.sets {
		out = append(out, key)
	}
	return out, nil
}

func (m *MockHandler) FlushAll() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]mockEntry)
	m.lists = make(map[string][]string)
	m.hashes = make(map[string]map[string]string)
	m.sets = make(map[string]map[string]struct{})
	return "OK", nil
}
