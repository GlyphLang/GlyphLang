// Package memory provides memory management utilities including object pooling
package memory

import (
	"bytes"
	"sync"
	"sync/atomic"
	"time"
)

// ========================================
// Generic Object Pool
// ========================================

// Pool is a generic object pool that manages reusable objects
type Pool[T any] struct {
	pool     sync.Pool
	factory  func() T
	reset    func(*T)
	stats    PoolStats
	maxSize  int64
	curSize  int64
}

// PoolStats tracks pool usage statistics
type PoolStats struct {
	Gets      uint64
	Puts      uint64
	Misses    uint64 // New allocations
	Hits      uint64 // Reused objects
	Discards  uint64 // Objects discarded (pool full)
}

// NewPool creates a new object pool
func NewPool[T any](factory func() T, reset func(*T)) *Pool[T] {
	p := &Pool[T]{
		factory: factory,
		reset:   reset,
		maxSize: 1000, // Default max size
	}
	p.pool.New = func() interface{} {
		atomic.AddUint64(&p.stats.Misses, 1)
		return factory()
	}
	return p
}

// Get retrieves an object from the pool
func (p *Pool[T]) Get() T {
	atomic.AddUint64(&p.stats.Gets, 1)
	obj := p.pool.Get().(T)
	if atomic.LoadUint64(&p.stats.Misses) < atomic.LoadUint64(&p.stats.Gets) {
		atomic.AddUint64(&p.stats.Hits, 1)
	}
	return obj
}

// Put returns an object to the pool
func (p *Pool[T]) Put(obj T) {
	atomic.AddUint64(&p.stats.Puts, 1)

	// Check if pool is full
	if atomic.LoadInt64(&p.curSize) >= p.maxSize {
		atomic.AddUint64(&p.stats.Discards, 1)
		return
	}

	// Reset the object before returning to pool
	if p.reset != nil {
		p.reset(&obj)
	}

	atomic.AddInt64(&p.curSize, 1)
	p.pool.Put(obj)
}

// Stats returns the pool statistics
func (p *Pool[T]) Stats() PoolStats {
	return PoolStats{
		Gets:     atomic.LoadUint64(&p.stats.Gets),
		Puts:     atomic.LoadUint64(&p.stats.Puts),
		Misses:   atomic.LoadUint64(&p.stats.Misses),
		Hits:     atomic.LoadUint64(&p.stats.Hits),
		Discards: atomic.LoadUint64(&p.stats.Discards),
	}
}

// SetMaxSize sets the maximum pool size
func (p *Pool[T]) SetMaxSize(size int64) {
	p.maxSize = size
}

// ========================================
// Buffer Pool
// ========================================

// BufferPool manages reusable byte buffers
type BufferPool struct {
	small  *sync.Pool // < 4KB
	medium *sync.Pool // < 64KB
	large  *sync.Pool // >= 64KB
	stats  BufferPoolStats
}

// BufferPoolStats tracks buffer pool statistics
type BufferPoolStats struct {
	SmallGets   uint64
	SmallPuts   uint64
	MediumGets  uint64
	MediumPuts  uint64
	LargeGets   uint64
	LargePuts   uint64
}

// NewBufferPool creates a new buffer pool
func NewBufferPool() *BufferPool {
	return &BufferPool{
		small: &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, 4*1024) // 4KB
				return &buf
			},
		},
		medium: &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, 64*1024) // 64KB
				return &buf
			},
		},
		large: &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, 1024*1024) // 1MB
				return &buf
			},
		},
	}
}

// Get retrieves a buffer of at least the specified size
func (bp *BufferPool) Get(size int) *[]byte {
	if size <= 4*1024 {
		atomic.AddUint64(&bp.stats.SmallGets, 1)
		return bp.small.Get().(*[]byte)
	} else if size <= 64*1024 {
		atomic.AddUint64(&bp.stats.MediumGets, 1)
		return bp.medium.Get().(*[]byte)
	} else {
		atomic.AddUint64(&bp.stats.LargeGets, 1)
		return bp.large.Get().(*[]byte)
	}
}

// Put returns a buffer to the pool
func (bp *BufferPool) Put(buf *[]byte) {
	if buf == nil {
		return
	}

	// Reset buffer
	*buf = (*buf)[:0]

	cap := cap(*buf)
	if cap <= 4*1024 {
		atomic.AddUint64(&bp.stats.SmallPuts, 1)
		bp.small.Put(buf)
	} else if cap <= 64*1024 {
		atomic.AddUint64(&bp.stats.MediumPuts, 1)
		bp.medium.Put(buf)
	} else {
		atomic.AddUint64(&bp.stats.LargePuts, 1)
		bp.large.Put(buf)
	}
}

// Stats returns buffer pool statistics
func (bp *BufferPool) Stats() BufferPoolStats {
	return BufferPoolStats{
		SmallGets:  atomic.LoadUint64(&bp.stats.SmallGets),
		SmallPuts:  atomic.LoadUint64(&bp.stats.SmallPuts),
		MediumGets: atomic.LoadUint64(&bp.stats.MediumGets),
		MediumPuts: atomic.LoadUint64(&bp.stats.MediumPuts),
		LargeGets:  atomic.LoadUint64(&bp.stats.LargeGets),
		LargePuts:  atomic.LoadUint64(&bp.stats.LargePuts),
	}
}

// ========================================
// Bytes Buffer Pool
// ========================================

// BytesBufferPool manages reusable bytes.Buffer instances
type BytesBufferPool struct {
	pool  sync.Pool
	stats PoolStats
}

// NewBytesBufferPool creates a new bytes.Buffer pool
func NewBytesBufferPool() *BytesBufferPool {
	return &BytesBufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// Get retrieves a bytes.Buffer from the pool
func (p *BytesBufferPool) Get() *bytes.Buffer {
	atomic.AddUint64(&p.stats.Gets, 1)
	return p.pool.Get().(*bytes.Buffer)
}

// Put returns a bytes.Buffer to the pool
func (p *BytesBufferPool) Put(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	atomic.AddUint64(&p.stats.Puts, 1)
	buf.Reset()
	p.pool.Put(buf)
}

// Stats returns pool statistics
func (p *BytesBufferPool) Stats() PoolStats {
	return PoolStats{
		Gets: atomic.LoadUint64(&p.stats.Gets),
		Puts: atomic.LoadUint64(&p.stats.Puts),
	}
}

// ========================================
// Map Pool (for JSON-like data)
// ========================================

// MapPool manages reusable map[string]interface{} instances
type MapPool struct {
	pool  sync.Pool
	stats PoolStats
}

// NewMapPool creates a new map pool
func NewMapPool() *MapPool {
	return &MapPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{})
			},
		},
	}
}

// Get retrieves a map from the pool
func (p *MapPool) Get() map[string]interface{} {
	atomic.AddUint64(&p.stats.Gets, 1)
	return p.pool.Get().(map[string]interface{})
}

// Put returns a map to the pool
func (p *MapPool) Put(m map[string]interface{}) {
	if m == nil {
		return
	}
	atomic.AddUint64(&p.stats.Puts, 1)
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	p.pool.Put(m)
}

// Stats returns pool statistics
func (p *MapPool) Stats() PoolStats {
	return PoolStats{
		Gets: atomic.LoadUint64(&p.stats.Gets),
		Puts: atomic.LoadUint64(&p.stats.Puts),
	}
}

// ========================================
// Slice Pool
// ========================================

// SlicePool manages reusable []interface{} slices
type SlicePool struct {
	pool  sync.Pool
	stats PoolStats
}

// NewSlicePool creates a new slice pool
func NewSlicePool() *SlicePool {
	return &SlicePool{
		pool: sync.Pool{
			New: func() interface{} {
				s := make([]interface{}, 0, 16)
				return &s
			},
		},
	}
}

// Get retrieves a slice from the pool
func (p *SlicePool) Get() *[]interface{} {
	atomic.AddUint64(&p.stats.Gets, 1)
	return p.pool.Get().(*[]interface{})
}

// Put returns a slice to the pool
func (p *SlicePool) Put(s *[]interface{}) {
	if s == nil {
		return
	}
	atomic.AddUint64(&p.stats.Puts, 1)
	*s = (*s)[:0]
	p.pool.Put(s)
}

// Stats returns pool statistics
func (p *SlicePool) Stats() PoolStats {
	return PoolStats{
		Gets: atomic.LoadUint64(&p.stats.Gets),
		Puts: atomic.LoadUint64(&p.stats.Puts),
	}
}

// ========================================
// Request Context Pool
// ========================================

// RequestContext represents a reusable request context
type RequestContext struct {
	Values    map[string]interface{}
	Headers   map[string]string
	Params    map[string]string
	Query     map[string]string
	Body      []byte
	StartTime time.Time
	RequestID string
}

// RequestContextPool manages reusable request contexts
type RequestContextPool struct {
	pool  sync.Pool
	stats PoolStats
}

// NewRequestContextPool creates a new request context pool
func NewRequestContextPool() *RequestContextPool {
	return &RequestContextPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &RequestContext{
					Values:  make(map[string]interface{}),
					Headers: make(map[string]string),
					Params:  make(map[string]string),
					Query:   make(map[string]string),
				}
			},
		},
	}
}

// Get retrieves a request context from the pool
func (p *RequestContextPool) Get() *RequestContext {
	atomic.AddUint64(&p.stats.Gets, 1)
	ctx := p.pool.Get().(*RequestContext)
	ctx.StartTime = time.Now()
	return ctx
}

// Put returns a request context to the pool
func (p *RequestContextPool) Put(ctx *RequestContext) {
	if ctx == nil {
		return
	}
	atomic.AddUint64(&p.stats.Puts, 1)

	// Clear maps
	for k := range ctx.Values {
		delete(ctx.Values, k)
	}
	for k := range ctx.Headers {
		delete(ctx.Headers, k)
	}
	for k := range ctx.Params {
		delete(ctx.Params, k)
	}
	for k := range ctx.Query {
		delete(ctx.Query, k)
	}

	// Reset other fields
	ctx.Body = ctx.Body[:0]
	ctx.RequestID = ""

	p.pool.Put(ctx)
}

// Stats returns pool statistics
func (p *RequestContextPool) Stats() PoolStats {
	return PoolStats{
		Gets: atomic.LoadUint64(&p.stats.Gets),
		Puts: atomic.LoadUint64(&p.stats.Puts),
	}
}

// ========================================
// Global Pools (Singletons)
// ========================================

var (
	globalBufferPool   *BufferPool
	globalBytesPool    *BytesBufferPool
	globalMapPool      *MapPool
	globalSlicePool    *SlicePool
	globalContextPool  *RequestContextPool
	initOnce           sync.Once
)

// initGlobalPools initializes all global pools
func initGlobalPools() {
	initOnce.Do(func() {
		globalBufferPool = NewBufferPool()
		globalBytesPool = NewBytesBufferPool()
		globalMapPool = NewMapPool()
		globalSlicePool = NewSlicePool()
		globalContextPool = NewRequestContextPool()
	})
}

// GetBuffer retrieves a buffer from the global pool
func GetBuffer(size int) *[]byte {
	initGlobalPools()
	return globalBufferPool.Get(size)
}

// PutBuffer returns a buffer to the global pool
func PutBuffer(buf *[]byte) {
	initGlobalPools()
	globalBufferPool.Put(buf)
}

// GetBytesBuffer retrieves a bytes.Buffer from the global pool
func GetBytesBuffer() *bytes.Buffer {
	initGlobalPools()
	return globalBytesPool.Get()
}

// PutBytesBuffer returns a bytes.Buffer to the global pool
func PutBytesBuffer(buf *bytes.Buffer) {
	initGlobalPools()
	globalBytesPool.Put(buf)
}

// GetMap retrieves a map from the global pool
func GetMap() map[string]interface{} {
	initGlobalPools()
	return globalMapPool.Get()
}

// PutMap returns a map to the global pool
func PutMap(m map[string]interface{}) {
	initGlobalPools()
	globalMapPool.Put(m)
}

// GetSlice retrieves a slice from the global pool
func GetSlice() *[]interface{} {
	initGlobalPools()
	return globalSlicePool.Get()
}

// PutSlice returns a slice to the global pool
func PutSlice(s *[]interface{}) {
	initGlobalPools()
	globalSlicePool.Put(s)
}

// GetRequestContext retrieves a request context from the global pool
func GetRequestContext() *RequestContext {
	initGlobalPools()
	return globalContextPool.Get()
}

// PutRequestContext returns a request context to the global pool
func PutRequestContext(ctx *RequestContext) {
	initGlobalPools()
	globalContextPool.Put(ctx)
}

// GlobalStats returns statistics for all global pools
func GlobalStats() map[string]PoolStats {
	initGlobalPools()
	return map[string]PoolStats{
		"bytes_buffer": globalBytesPool.Stats(),
		"map":          globalMapPool.Stats(),
		"slice":        globalSlicePool.Stats(),
		"context":      globalContextPool.Stats(),
	}
}
