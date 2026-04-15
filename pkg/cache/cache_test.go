package cache

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestLRUCache_SetGet(t *testing.T) {
	cache := NewLRUCache(WithCapacity(100))

	// Test basic set/get
	err := cache.Set("key1", "value1", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, ok := cache.Get("key1")
	if !ok {
		t.Fatal("Get returned false for existing key")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test missing key
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("Get returned true for non-existent key")
	}
}

func TestLRUCache_Expiration(t *testing.T) {
	cache := NewLRUCache(WithCapacity(100))

	// Set with short TTL
	err := cache.Set("expiring", "value", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Should be available immediately
	_, ok := cache.Get("expiring")
	if !ok {
		t.Error("Key should exist before expiration")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be gone
	_, ok = cache.Get("expiring")
	if ok {
		t.Error("Key should be expired")
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(WithCapacity(3))

	// Fill cache
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	cache.Set("key3", "value3", 0)

	// Access key1 to make it recently used
	cache.Get("key1")

	// Add another entry, should evict key2 (least recently used)
	cache.Set("key4", "value4", 0)

	// key2 should be gone
	_, ok := cache.Get("key2")
	if ok {
		t.Error("key2 should have been evicted")
	}

	// key1 should still exist
	_, ok = cache.Get("key1")
	if !ok {
		t.Error("key1 should still exist")
	}
}

func TestLRUCache_Delete(t *testing.T) {
	cache := NewLRUCache(WithCapacity(100))

	cache.Set("key1", "value1", 0)

	err := cache.Delete("key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, ok := cache.Get("key1")
	if ok {
		t.Error("Key should be deleted")
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(WithCapacity(100))

	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	cache.Set("key3", "value3", 0)

	err := cache.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	stats := cache.Stats()
	if stats.EntryCount != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", stats.EntryCount)
	}
}

func TestLRUCache_Stats(t *testing.T) {
	cache := NewLRUCache(WithCapacity(100))

	// Generate some hits and misses
	cache.Set("key1", "value1", 0)
	cache.Get("key1")        // Hit
	cache.Get("key1")        // Hit
	cache.Get("nonexistent") // Miss
	cache.Get("nonexistent") // Miss

	stats := cache.Stats()
	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 2 {
		t.Errorf("Expected 2 misses, got %d", stats.Misses)
	}
	if stats.Sets != 1 {
		t.Errorf("Expected 1 set, got %d", stats.Sets)
	}
}

func TestLRUCache_Tags(t *testing.T) {
	cache := NewLRUCache(WithCapacity(100))

	// Set with tags
	cache.SetWithTags("user:1", "Alice", 0, []string{"users"})
	cache.SetWithTags("user:2", "Bob", 0, []string{"users"})
	cache.SetWithTags("post:1", "Hello", 0, []string{"posts"})

	// Delete by tag
	count := cache.DeleteByTag("users")
	if count != 2 {
		t.Errorf("Expected 2 deleted, got %d", count)
	}

	// Users should be gone
	_, ok := cache.Get("user:1")
	if ok {
		t.Error("user:1 should be deleted")
	}

	// Posts should remain
	_, ok = cache.Get("post:1")
	if !ok {
		t.Error("post:1 should still exist")
	}
}

func TestLRUCache_Concurrent(t *testing.T) {
	cache := NewLRUCache(WithCapacity(1000))
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := string(rune(n*100 + j))
				cache.Set(key, j, 0)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := string(rune(n*100 + j))
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()
}

func TestHTTPCache_GetSet(t *testing.T) {
	config := DefaultHTTPCacheConfig()
	hc := NewHTTPCache(config)

	resp := &CachedResponse{
		StatusCode: 200,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"status":"ok"}`),
		ETag:       GenerateETag([]byte(`{"status":"ok"}`)),
		CreatedAt:  time.Now(),
	}

	err := hc.Set("GET:/api/test", resp, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	cached, ok := hc.Get("GET:/api/test")
	if !ok {
		t.Fatal("Get returned false for existing key")
	}
	if cached.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", cached.StatusCode)
	}
	if string(cached.Body) != `{"status":"ok"}` {
		t.Errorf("Unexpected body: %s", cached.Body)
	}
}

func TestGenerateETag(t *testing.T) {
	body := []byte("test content")
	etag := GenerateETag(body)

	// ETag should be in quotes
	if etag[0] != '"' || etag[len(etag)-1] != '"' {
		t.Error("ETag should be quoted")
	}

	// Same content should produce same ETag
	etag2 := GenerateETag(body)
	if etag != etag2 {
		t.Error("Same content should produce same ETag")
	}

	// Different content should produce different ETag
	etag3 := GenerateETag([]byte("different content"))
	if etag == etag3 {
		t.Error("Different content should produce different ETag")
	}
}

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		query    string
		expected string
	}{
		{"GET", "/api/users", "", "GET:/api/users"},
		{"GET", "/api/users", "page=1", "GET:/api/users?page=1"},
		{"POST", "/api/users", "", "POST:/api/users"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path+"?"+tt.query, nil)
		if tt.query == "" {
			req = httptest.NewRequest(tt.method, tt.path, nil)
		}
		key := GenerateCacheKey(req)
		if key != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, key)
		}
	}
}

func TestHTTPCacheMiddleware(t *testing.T) {
	config := DefaultHTTPCacheConfig()
	hc := NewHTTPCache(config)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Wrap with cache middleware
	cachedHandler := hc.Middleware()(handler)

	// First request - should miss cache
	req1 := httptest.NewRequest("GET", "/api/test", nil)
	rec1 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec1.Code)
	}

	stats := hc.Stats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}

	// Second request - should hit cache
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	rec2 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(rec2, req2)

	stats = hc.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
}

func TestHTTPCacheMiddleware_ETagMatch(t *testing.T) {
	config := DefaultHTTPCacheConfig()
	hc := NewHTTPCache(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	cachedHandler := hc.Middleware()(handler)

	// First request to populate cache
	req1 := httptest.NewRequest("GET", "/api/test", nil)
	rec1 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(rec1, req1)

	etag := rec1.Header().Get("ETag")
	if etag == "" {
		t.Fatal("Expected ETag header")
	}

	// Second request with If-None-Match
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	req2.Header.Set("If-None-Match", etag)
	rec2 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusNotModified {
		t.Errorf("Expected 304 Not Modified, got %d", rec2.Code)
	}
}

func TestKeyBuilder(t *testing.T) {
	tests := []struct {
		parts    []string
		expected string
	}{
		{[]string{"user", "123"}, "user:123"},
		{[]string{"api", "v1", "users"}, "api:v1:users"},
		{[]string{"cache"}, "cache"},
	}

	for _, tt := range tests {
		kb := NewKeyBuilder()
		for _, p := range tt.parts {
			kb.Add(p)
		}
		result := kb.Build()
		if result != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, result)
		}
	}
}

func TestGlobalCache(t *testing.T) {
	// Test global cache functions
	err := Set("global:key1", "value1", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, ok := Get("global:key1")
	if !ok {
		t.Fatal("Get returned false")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	err = Delete("global:key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, ok = Get("global:key1")
	if ok {
		t.Error("Key should be deleted")
	}
}

func TestInvalidateByPrefix(t *testing.T) {
	config := DefaultHTTPCacheConfig()
	hc := NewHTTPCache(config)

	// Add entries with common prefix
	hc.Set("GET:/api/users/1", &CachedResponse{StatusCode: 200}, 0)
	hc.Set("GET:/api/users/2", &CachedResponse{StatusCode: 200}, 0)
	hc.Set("GET:/api/posts/1", &CachedResponse{StatusCode: 200}, 0)

	// Invalidate by prefix
	count := hc.InvalidateByPrefix("GET:/api/users")
	if count != 2 {
		t.Errorf("Expected 2 invalidated, got %d", count)
	}

	// Users should be gone
	_, ok := hc.Get("GET:/api/users/1")
	if ok {
		t.Error("users/1 should be invalidated")
	}

	// Posts should remain
	_, ok = hc.Get("GET:/api/posts/1")
	if !ok {
		t.Error("posts/1 should still exist")
	}
}

func TestLRUCache_MaxSizeEviction(t *testing.T) {
	// Set a small max size so entries get evicted based on memory
	cache := NewLRUCache(WithCapacity(1000), WithMaxSize(100))
	defer cache.Close()

	// Add entries until we exceed the max size
	// Each string value's size is estimated by its length
	cache.Set("key1", "aaaaaaaaaa", 0) // ~10 bytes
	cache.Set("key2", "bbbbbbbbbb", 0) // ~10 bytes
	cache.Set("key3", "cccccccccc", 0) // ~10 bytes

	// Verify they all fit
	if _, ok := cache.Get("key1"); !ok {
		t.Error("key1 should exist")
	}

	// Now add a large entry that pushes us over the limit
	// This should cause evictions of oldest entries
	largeValue := make([]byte, 90)
	for i := range largeValue {
		largeValue[i] = 'x'
	}
	cache.Set("big", string(largeValue), 0)

	stats := cache.Stats()
	if stats.Evictions == 0 {
		t.Error("Expected evictions when exceeding max size")
	}

	// The big entry should exist
	if _, ok := cache.Get("big"); !ok {
		t.Error("big key should exist after eviction of older entries")
	}
}

func TestLRUCache_UpdateExistingKey(t *testing.T) {
	cache := NewLRUCache(WithCapacity(100))
	defer cache.Close()

	cache.Set("a", "val1", 0)

	stats := cache.Stats()
	initialCount := stats.EntryCount

	// Update the same key
	cache.Set("a", "val2", 0)

	value, ok := cache.Get("a")
	if !ok {
		t.Fatal("key 'a' should exist after update")
	}
	if value != "val2" {
		t.Errorf("Expected 'val2', got %v", value)
	}

	stats = cache.Stats()
	if stats.EntryCount != initialCount {
		t.Errorf("EntryCount should remain %d after update, got %d", initialCount, stats.EntryCount)
	}
}

func TestLRUCache_OnEvict(t *testing.T) {
	var mu sync.Mutex
	evictedKeys := make([]string, 0)

	cache := NewLRUCache(
		WithCapacity(2),
		WithOnEvict(func(key string, value interface{}) {
			mu.Lock()
			evictedKeys = append(evictedKeys, key)
			mu.Unlock()
		}),
	)
	defer cache.Close()

	cache.Set("a", "1", 0)
	cache.Set("b", "2", 0)

	// This should evict "a" (oldest/least recently used)
	cache.Set("c", "3", 0)

	mu.Lock()
	defer mu.Unlock()
	if len(evictedKeys) == 0 {
		t.Fatal("Expected eviction callback to be called")
	}
	if evictedKeys[0] != "a" {
		t.Errorf("Expected evicted key 'a', got '%s'", evictedKeys[0])
	}
}

func TestLRUCache_DeleteNonExistent(t *testing.T) {
	cache := NewLRUCache(WithCapacity(100))
	defer cache.Close()

	err := cache.Delete("does-not-exist")
	if err != nil {
		t.Errorf("Deleting non-existent key should not error, got: %v", err)
	}
}

func TestLRUCache_ConcurrentMixedOperations(t *testing.T) {
	cache := NewLRUCache(WithCapacity(100))
	defer cache.Close()

	var wg sync.WaitGroup
	const goroutines = 20
	const opsPerGoroutine = 200

	// Concurrent sets
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := string(rune(id*opsPerGoroutine + j))
				cache.Set(key, j, 0)
			}
		}(i)
	}

	// Concurrent gets
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := string(rune(id*opsPerGoroutine + j))
				cache.Get(key)
			}
		}(i)
	}

	// Concurrent deletes
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := string(rune(id*opsPerGoroutine + j))
				cache.Delete(key)
			}
		}(i)
	}

	// Concurrent clears
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				cache.Clear()
			}
		}()
	}

	wg.Wait()
	// If we get here without a race condition panic, the test passes
}

func TestLRUCache_DefaultTTL(t *testing.T) {
	cache := NewLRUCache(
		WithCapacity(100),
		WithDefaultTTL(50*time.Millisecond),
	)
	defer cache.Close()

	// Set with ttl=0, which should use the default TTL of 50ms
	cache.Set("default-ttl-key", "value", 0)

	// Should exist immediately
	if _, ok := cache.Get("default-ttl-key"); !ok {
		t.Error("Key should exist immediately after set")
	}

	// Wait for default TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	if _, ok := cache.Get("default-ttl-key"); ok {
		t.Error("Key should have expired after default TTL")
	}
}

func TestLRUCache_Close(t *testing.T) {
	cache := NewLRUCache(WithCapacity(100))

	cache.Set("key", "value", 0)

	// First close should succeed
	cache.Close()

	// Second close should not panic (idempotent)
	cache.Close()

	// Cache should still be functional after close (cleanup goroutine stops but data remains)
	value, ok := cache.Get("key")
	if !ok {
		t.Error("Cache should still return values after Close")
	}
	if value != "value" {
		t.Errorf("Expected 'value', got %v", value)
	}

	// Set should still work
	err := cache.Set("key2", "value2", 0)
	if err != nil {
		t.Errorf("Set should still work after Close, got: %v", err)
	}
}

func TestEntry_IsExpired(t *testing.T) {
	// Zero time = never expires
	entry := &Entry{ExpiresAt: time.Time{}}
	if entry.IsExpired() {
		t.Error("Entry with zero ExpiresAt should never expire")
	}

	// Past time = expired
	entry = &Entry{ExpiresAt: time.Now().Add(-1 * time.Second)}
	if !entry.IsExpired() {
		t.Error("Entry with past ExpiresAt should be expired")
	}

	// Future time = not expired
	entry = &Entry{ExpiresAt: time.Now().Add(1 * time.Hour)}
	if entry.IsExpired() {
		t.Error("Entry with future ExpiresAt should not be expired")
	}
}

func TestHTTPCache_NonGetRequest(t *testing.T) {
	config := DefaultHTTPCacheConfig()
	hc := NewHTTPCache(config)

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1}`))
	})

	cachedHandler := hc.Middleware()(handler)

	// POST request should pass through without caching
	req := httptest.NewRequest("POST", "/api/users", nil)
	rec := httptest.NewRecorder()
	cachedHandler.ServeHTTP(rec, req)

	if !called {
		t.Error("Handler should have been called for POST request")
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", rec.Code)
	}

	// Verify nothing was cached
	stats := hc.Stats()
	if stats.Sets != 0 {
		t.Errorf("POST requests should not be cached, but got %d sets", stats.Sets)
	}
}

func TestKeyBuilder_AddIfNotEmpty(t *testing.T) {
	kb := NewKeyBuilder()
	kb.Add("prefix")
	kb.AddIfNotEmpty("")     // Should be skipped
	kb.AddIfNotEmpty("part") // Should be added
	kb.AddIfNotEmpty("")     // Should be skipped

	result := kb.Build()
	expected := "prefix:part"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func BenchmarkLRUCache_Set(b *testing.B) {
	cache := NewLRUCache(WithCapacity(10000))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Set(string(rune(i)), i, 0)
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(WithCapacity(10000))
	for i := 0; i < 10000; i++ {
		cache.Set(string(rune(i)), i, 0)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Get(string(rune(i % 10000)))
	}
}

func BenchmarkLRUCache_Concurrent(b *testing.B) {
	cache := NewLRUCache(WithCapacity(10000))
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				cache.Set(string(rune(i)), i, 0)
			} else {
				cache.Get(string(rune(i)))
			}
			i++
		}
	})
}
