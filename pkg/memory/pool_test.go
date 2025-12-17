package memory

import (
	"bytes"
	"sync"
	"testing"
)

func TestBufferPool(t *testing.T) {
	pool := NewBufferPool()

	// Test small buffer
	buf := pool.Get(100)
	if buf == nil {
		t.Error("Expected non-nil buffer")
	}
	if cap(*buf) < 100 {
		t.Errorf("Expected capacity >= 100, got %d", cap(*buf))
	}
	pool.Put(buf)

	// Test medium buffer
	buf = pool.Get(10000)
	if buf == nil {
		t.Error("Expected non-nil buffer")
	}
	if cap(*buf) < 10000 {
		t.Errorf("Expected capacity >= 10000, got %d", cap(*buf))
	}
	pool.Put(buf)

	// Test large buffer
	buf = pool.Get(100000)
	if buf == nil {
		t.Error("Expected non-nil buffer")
	}
	if cap(*buf) < 100000 {
		t.Errorf("Expected capacity >= 100000, got %d", cap(*buf))
	}
	pool.Put(buf)

	// Check stats
	stats := pool.Stats()
	if stats.SmallGets != 1 {
		t.Errorf("Expected SmallGets = 1, got %d", stats.SmallGets)
	}
	if stats.MediumGets != 1 {
		t.Errorf("Expected MediumGets = 1, got %d", stats.MediumGets)
	}
	if stats.LargeGets != 1 {
		t.Errorf("Expected LargeGets = 1, got %d", stats.LargeGets)
	}
}

func TestBytesBufferPool(t *testing.T) {
	pool := NewBytesBufferPool()

	// Get and use buffer
	buf := pool.Get()
	if buf == nil {
		t.Error("Expected non-nil buffer")
	}

	buf.WriteString("test data")
	if buf.String() != "test data" {
		t.Error("Buffer write failed")
	}

	// Return to pool
	pool.Put(buf)

	// Get again - should be reset
	buf2 := pool.Get()
	if buf2.Len() != 0 {
		t.Error("Buffer should be reset after Put")
	}
	pool.Put(buf2)

	// Check stats
	stats := pool.Stats()
	if stats.Gets != 2 {
		t.Errorf("Expected Gets = 2, got %d", stats.Gets)
	}
	if stats.Puts != 2 {
		t.Errorf("Expected Puts = 2, got %d", stats.Puts)
	}
}

func TestMapPool(t *testing.T) {
	pool := NewMapPool()

	// Get and use map
	m := pool.Get()
	if m == nil {
		t.Error("Expected non-nil map")
	}

	m["key"] = "value"
	if m["key"] != "value" {
		t.Error("Map assignment failed")
	}

	// Return to pool
	pool.Put(m)

	// Get again - should be cleared
	m2 := pool.Get()
	if len(m2) != 0 {
		t.Error("Map should be cleared after Put")
	}
	pool.Put(m2)
}

func TestSlicePool(t *testing.T) {
	pool := NewSlicePool()

	// Get and use slice
	s := pool.Get()
	if s == nil {
		t.Error("Expected non-nil slice")
	}

	*s = append(*s, "item1", "item2")
	if len(*s) != 2 {
		t.Error("Slice append failed")
	}

	// Return to pool
	pool.Put(s)

	// Get again - should be cleared
	s2 := pool.Get()
	if len(*s2) != 0 {
		t.Error("Slice should be cleared after Put")
	}
	pool.Put(s2)
}

func TestRequestContextPool(t *testing.T) {
	pool := NewRequestContextPool()

	// Get context
	ctx := pool.Get()
	if ctx == nil {
		t.Error("Expected non-nil context")
	}

	// Use context
	ctx.RequestID = "req-123"
	ctx.Values["user"] = "john"
	ctx.Headers["Authorization"] = "Bearer token"
	ctx.Params["id"] = "456"
	ctx.Query["page"] = "1"

	// Return to pool
	pool.Put(ctx)

	// Get again - should be reset
	ctx2 := pool.Get()
	if ctx2.RequestID != "" {
		t.Error("RequestID should be reset")
	}
	if len(ctx2.Values) != 0 {
		t.Error("Values should be reset")
	}
	if len(ctx2.Headers) != 0 {
		t.Error("Headers should be reset")
	}
	pool.Put(ctx2)
}

func TestGlobalPools(t *testing.T) {
	// Test global buffer
	buf := GetBuffer(100)
	if buf == nil {
		t.Error("GetBuffer returned nil")
	}
	PutBuffer(buf)

	// Test global bytes buffer
	bb := GetBytesBuffer()
	if bb == nil {
		t.Error("GetBytesBuffer returned nil")
	}
	PutBytesBuffer(bb)

	// Test global map
	m := GetMap()
	if m == nil {
		t.Error("GetMap returned nil")
	}
	PutMap(m)

	// Test global slice
	s := GetSlice()
	if s == nil {
		t.Error("GetSlice returned nil")
	}
	PutSlice(s)

	// Test global context
	ctx := GetRequestContext()
	if ctx == nil {
		t.Error("GetRequestContext returned nil")
	}
	PutRequestContext(ctx)

	// Test global stats
	stats := GlobalStats()
	if len(stats) != 4 {
		t.Errorf("Expected 4 stat entries, got %d", len(stats))
	}
}

func TestConcurrentAccess(t *testing.T) {
	pool := NewBytesBufferPool()
	var wg sync.WaitGroup

	// Concurrent get/put
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				buf := pool.Get()
				buf.WriteString("concurrent test")
				pool.Put(buf)
			}
		}()
	}

	wg.Wait()

	stats := pool.Stats()
	if stats.Gets != 10000 {
		t.Errorf("Expected Gets = 10000, got %d", stats.Gets)
	}
	if stats.Puts != 10000 {
		t.Errorf("Expected Puts = 10000, got %d", stats.Puts)
	}
}

func BenchmarkBufferPoolGet(b *testing.B) {
	pool := NewBufferPool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := pool.Get(1024)
		pool.Put(buf)
	}
}

func BenchmarkDirectAllocation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, 1024)
		_ = buf
	}
}

func BenchmarkBytesBufferPool(b *testing.B) {
	pool := NewBytesBufferPool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := pool.Get()
		buf.WriteString("benchmark test data")
		pool.Put(buf)
	}
}

func BenchmarkDirectBytesBuffer(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		buf.WriteString("benchmark test data")
		_ = buf
	}
}

func BenchmarkMapPool(b *testing.B) {
	pool := NewMapPool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m := pool.Get()
		m["key1"] = "value1"
		m["key2"] = "value2"
		pool.Put(m)
	}
}

func BenchmarkDirectMap(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m := make(map[string]interface{})
		m["key1"] = "value1"
		m["key2"] = "value2"
		_ = m
	}
}

func BenchmarkRequestContextPool(b *testing.B) {
	pool := NewRequestContextPool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := pool.Get()
		ctx.RequestID = "req-123"
		ctx.Values["user"] = "john"
		pool.Put(ctx)
	}
}
