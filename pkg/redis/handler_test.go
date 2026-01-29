package redis

import (
	"testing"
)

// Tests use MockHandler since they don't require a real Redis server.
// Integration tests with a real server would go in a separate file.

func TestMockHandler_Ping(t *testing.T) {
	m := NewMockHandler()
	result, err := m.Ping()
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
	if result != "PONG" {
		t.Errorf("expected PONG, got %s", result)
	}
}

func TestMockHandler_GetSet(t *testing.T) {
	m := NewMockHandler()

	// Get non-existent key
	val, err := m.Get("missing")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil for missing key, got %v", val)
	}

	// Set and Get
	_, err = m.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err = m.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}

	// Overwrite
	_, err = m.Set("key1", "value2")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val, err = m.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value2" {
		t.Errorf("expected value2, got %v", val)
	}
}

func TestMockHandler_SetWithTTL(t *testing.T) {
	m := NewMockHandler()

	// Set with TTL
	_, err := m.Set("ttl-key", "value", int64(3600))
	if err != nil {
		t.Fatalf("Set with TTL failed: %v", err)
	}

	val, err := m.Get("ttl-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value" {
		t.Errorf("expected value, got %v", val)
	}

	// Check TTL is set
	ttl, err := m.Ttl("ttl-key")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if ttl <= 0 {
		t.Errorf("expected positive TTL, got %d", ttl)
	}
}

func TestMockHandler_Del(t *testing.T) {
	m := NewMockHandler()
	m.Set("key1", "val1")
	m.Set("key2", "val2")

	count, err := m.Del("key1")
	if err != nil {
		t.Fatalf("Del failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 deleted, got %d", count)
	}

	val, _ := m.Get("key1")
	if val != nil {
		t.Errorf("expected nil after delete, got %v", val)
	}

	// key2 should still exist
	val, _ = m.Get("key2")
	if val != "val2" {
		t.Errorf("expected val2 to still exist, got %v", val)
	}
}

func TestMockHandler_Exists(t *testing.T) {
	m := NewMockHandler()
	m.Set("exists-key", "value")

	count, err := m.Exists("exists-key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 for existing key, got %d", count)
	}

	count, err = m.Exists("missing-key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for missing key, got %d", count)
	}
}

func TestMockHandler_IncrDecr(t *testing.T) {
	m := NewMockHandler()

	// Incr on non-existent key
	val, err := m.Incr("counter")
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	// Incr again
	val, err = m.Incr("counter")
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}
	if val != 2 {
		t.Errorf("expected 2, got %d", val)
	}

	// Decr
	val, err = m.Decr("counter")
	if err != nil {
		t.Fatalf("Decr failed: %v", err)
	}
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	// Decr to 0
	val, err = m.Decr("counter")
	if err != nil {
		t.Fatalf("Decr failed: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0, got %d", val)
	}
}

func TestMockHandler_Hash(t *testing.T) {
	m := NewMockHandler()

	// HSet
	added, err := m.HSet("user:1", "name", "Alice", "age", "30")
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	if added != 2 {
		t.Errorf("expected 2 fields added, got %d", added)
	}

	// HGet
	val, err := m.HGet("user:1", "name")
	if err != nil {
		t.Fatalf("HGet failed: %v", err)
	}
	if val != "Alice" {
		t.Errorf("expected Alice, got %v", val)
	}

	// HGet missing field
	val, err = m.HGet("user:1", "missing")
	if err != nil {
		t.Fatalf("HGet failed: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil for missing field, got %v", val)
	}

	// HExists
	exists, err := m.HExists("user:1", "name")
	if err != nil {
		t.Fatalf("HExists failed: %v", err)
	}
	if !exists {
		t.Error("expected name field to exist")
	}

	// HGetAll
	all, err := m.HGetAll("user:1")
	if err != nil {
		t.Fatalf("HGetAll failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 fields, got %d", len(all))
	}

	// HDel
	deleted, err := m.HDel("user:1", "age")
	if err != nil {
		t.Fatalf("HDel failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	val, _ = m.HGet("user:1", "age")
	if val != nil {
		t.Errorf("expected nil after HDel, got %v", val)
	}
}

func TestMockHandler_List(t *testing.T) {
	m := NewMockHandler()

	// LPush
	length, err := m.LPush("queue", "a", "b", "c")
	if err != nil {
		t.Fatalf("LPush failed: %v", err)
	}
	if length != 3 {
		t.Errorf("expected length 3, got %d", length)
	}

	// LLen
	llen, err := m.LLen("queue")
	if err != nil {
		t.Fatalf("LLen failed: %v", err)
	}
	if llen != 3 {
		t.Errorf("expected 3, got %d", llen)
	}

	// LRange (all)
	items, err := m.LRange("queue", 0, -1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}

	// LPop
	val, err := m.LPop("queue")
	if err != nil {
		t.Fatalf("LPop failed: %v", err)
	}
	if val == nil {
		t.Fatal("expected non-nil value from LPop")
	}

	// RPush
	m.RPush("queue2", "x", "y")

	// RPop
	val, err = m.RPop("queue2")
	if err != nil {
		t.Fatalf("RPop failed: %v", err)
	}
	if val != "y" {
		t.Errorf("expected y, got %v", val)
	}

	// Pop from empty list
	val, err = m.LPop("empty")
	if err != nil {
		t.Fatalf("LPop empty failed: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil from empty list, got %v", val)
	}
}

func TestMockHandler_Set_Operations(t *testing.T) {
	m := NewMockHandler()

	// SAdd
	added, err := m.SAdd("tags", "go", "redis", "glyph")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	if added != 3 {
		t.Errorf("expected 3 added, got %d", added)
	}

	// SAdd duplicate
	added, err = m.SAdd("tags", "go")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	if added != 0 {
		t.Errorf("expected 0 for duplicate, got %d", added)
	}

	// SIsMember
	isMember, err := m.SIsMember("tags", "redis")
	if err != nil {
		t.Fatalf("SIsMember failed: %v", err)
	}
	if !isMember {
		t.Error("expected redis to be a member")
	}

	isMember, err = m.SIsMember("tags", "python")
	if err != nil {
		t.Fatalf("SIsMember failed: %v", err)
	}
	if isMember {
		t.Error("expected python not to be a member")
	}

	// SMembers
	members, err := m.SMembers("tags")
	if err != nil {
		t.Fatalf("SMembers failed: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("expected 3 members, got %d", len(members))
	}

	// SRem
	removed, err := m.SRem("tags", "redis")
	if err != nil {
		t.Fatalf("SRem failed: %v", err)
	}
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	members, _ = m.SMembers("tags")
	if len(members) != 2 {
		t.Errorf("expected 2 members after removal, got %d", len(members))
	}
}

func TestMockHandler_FlushAll(t *testing.T) {
	m := NewMockHandler()
	m.Set("key1", "val")
	m.LPush("list1", "a")
	m.HSet("hash1", "f", "v")
	m.SAdd("set1", "x")

	result, err := m.FlushAll()
	if err != nil {
		t.Fatalf("FlushAll failed: %v", err)
	}
	if result != "OK" {
		t.Errorf("expected OK, got %s", result)
	}

	val, _ := m.Get("key1")
	if val != nil {
		t.Error("expected nil after FlushAll")
	}

	llen, _ := m.LLen("list1")
	if llen != 0 {
		t.Errorf("expected empty list after FlushAll, got %d", llen)
	}
}

func TestMockHandler_Expire(t *testing.T) {
	m := NewMockHandler()
	m.Set("key1", "val")

	ok, err := m.Expire("key1", int64(100))
	if err != nil {
		t.Fatalf("Expire failed: %v", err)
	}
	if !ok {
		t.Error("expected true for existing key")
	}

	ttl, _ := m.Ttl("key1")
	if ttl <= 0 || ttl > 100 {
		t.Errorf("expected TTL between 1-100, got %d", ttl)
	}

	// Expire on non-existent key
	ok, _ = m.Expire("missing", int64(100))
	if ok {
		t.Error("expected false for non-existent key")
	}

	// TTL on non-existent key
	ttl, _ = m.Ttl("missing")
	if ttl != -2 {
		t.Errorf("expected -2 for missing key, got %d", ttl)
	}

	// TTL on key without expiry
	m.Set("no-ttl", "val")
	ttl, _ = m.Ttl("no-ttl")
	if ttl != -1 {
		t.Errorf("expected -1 for key without TTL, got %d", ttl)
	}
}

func TestMockHandler_Keys(t *testing.T) {
	m := NewMockHandler()
	m.Set("k1", "v1")
	m.Set("k2", "v2")
	m.LPush("list1", "a")

	keys, err := m.Keys("*")
	if err != nil {
		t.Fatalf("Keys failed: %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}

func TestMockHandler_SetErrors(t *testing.T) {
	m := NewMockHandler()

	// Set with too few args
	_, err := m.Set("only-key")
	if err == nil {
		t.Error("expected error for Set with too few args")
	}

	// Set with non-string key
	_, err = m.Set(123, "value")
	if err == nil {
		t.Error("expected error for Set with non-string key")
	}
}

func TestMockHandler_Publish(t *testing.T) {
	m := NewMockHandler()

	// Publish is a no-op for mock, just verify it doesn't error
	count, err := m.Publish("channel", "message")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for mock publish, got %d", count)
	}
}
