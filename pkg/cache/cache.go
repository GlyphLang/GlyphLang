// Package cache provides caching functionality including in-memory cache and HTTP caching
package cache

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// ========================================
// Cache Entry
// ========================================

// Entry represents a cached item
type Entry struct {
	Key         string
	Value       interface{}
	ExpiresAt   time.Time
	CreatedAt   time.Time
	AccessedAt  time.Time
	AccessCount uint64
	Size        int64
	Tags        []string
}

// IsExpired checks if the entry has expired
func (e *Entry) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return false // No expiration
	}
	return time.Now().After(e.ExpiresAt)
}

// ========================================
// Cache Interface
// ========================================

// Cache defines the cache interface
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Clear() error
	Stats() Stats
}

// Stats represents cache statistics
type Stats struct {
	Hits       uint64
	Misses     uint64
	Sets       uint64
	Deletes    uint64
	Evictions  uint64
	Size       int64
	MaxSize    int64
	EntryCount int64
}

// ========================================
// In-Memory LRU Cache
// ========================================

// LRUCache implements an LRU (Least Recently Used) cache
type LRUCache struct {
	mu          sync.RWMutex
	capacity    int
	items       map[string]*list.Element
	evictList   *list.List
	ttl         time.Duration
	onEvict     func(key string, value interface{})
	stats       Stats
	maxSize     int64 // Maximum size in bytes
	currentSize int64
}

// LRUOption configures the LRU cache
type LRUOption func(*LRUCache)

// WithCapacity sets the cache capacity
func WithCapacity(capacity int) LRUOption {
	return func(c *LRUCache) {
		c.capacity = capacity
	}
}

// WithDefaultTTL sets the default TTL
func WithDefaultTTL(ttl time.Duration) LRUOption {
	return func(c *LRUCache) {
		c.ttl = ttl
	}
}

// WithOnEvict sets the eviction callback
func WithOnEvict(fn func(key string, value interface{})) LRUOption {
	return func(c *LRUCache) {
		c.onEvict = fn
	}
}

// WithMaxSize sets the maximum size in bytes
func WithMaxSize(size int64) LRUOption {
	return func(c *LRUCache) {
		c.maxSize = size
	}
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(opts ...LRUOption) *LRUCache {
	c := &LRUCache{
		capacity:  1000,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
		ttl:       5 * time.Minute,
		maxSize:   100 * 1024 * 1024, // 100MB default
	}

	for _, opt := range opts {
		opt(c)
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		atomic.AddUint64(&c.stats.Misses, 1)
		return nil, false
	}

	entry := elem.Value.(*Entry)
	if entry.IsExpired() {
		c.removeElement(elem)
		atomic.AddUint64(&c.stats.Misses, 1)
		return nil, false
	}

	// Move to front (most recently used)
	c.evictList.MoveToFront(elem)
	entry.AccessedAt = time.Now()
	entry.AccessCount++

	atomic.AddUint64(&c.stats.Hits, 1)
	return entry.Value, true
}

// Set adds a value to the cache
func (c *LRUCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl == 0 {
		ttl = c.ttl
	}

	// Calculate entry size (rough estimate)
	size := estimateSize(value)

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	entry := &Entry{
		Key:        key,
		Value:      value,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		Size:       size,
	}

	// Check if key already exists
	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		oldEntry := elem.Value.(*Entry)
		c.currentSize -= oldEntry.Size
		c.currentSize += size
		elem.Value = entry
		atomic.AddUint64(&c.stats.Sets, 1)
		return nil
	}

	// Evict if necessary
	for c.evictList.Len() >= c.capacity || (c.maxSize > 0 && c.currentSize+size > c.maxSize) {
		c.evictOldest()
	}

	// Add new entry
	elem := c.evictList.PushFront(entry)
	c.items[key] = elem
	c.currentSize += size
	atomic.AddUint64(&c.stats.Sets, 1)
	atomic.AddInt64(&c.stats.EntryCount, 1)

	return nil
}

// SetWithTags adds a value with tags for grouped invalidation
func (c *LRUCache) SetWithTags(key string, value interface{}, ttl time.Duration, tags []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl == 0 {
		ttl = c.ttl
	}

	size := estimateSize(value)

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	entry := &Entry{
		Key:        key,
		Value:      value,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		Size:       size,
		Tags:       tags,
	}

	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		oldEntry := elem.Value.(*Entry)
		c.currentSize -= oldEntry.Size
		c.currentSize += size
		elem.Value = entry
		return nil
	}

	for c.evictList.Len() >= c.capacity || (c.maxSize > 0 && c.currentSize+size > c.maxSize) {
		c.evictOldest()
	}

	elem := c.evictList.PushFront(entry)
	c.items[key] = elem
	c.currentSize += size
	atomic.AddInt64(&c.stats.EntryCount, 1)

	return nil
}

// Delete removes a value from the cache
func (c *LRUCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.removeElement(elem)
		atomic.AddUint64(&c.stats.Deletes, 1)
	}

	return nil
}

// DeleteByTag removes all entries with the given tag
func (c *LRUCache) DeleteByTag(tag string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for _, elem := range c.items {
		entry := elem.Value.(*Entry)
		for _, t := range entry.Tags {
			if t == tag {
				c.removeElement(elem)
				count++
				break
			}
		}
	}

	return count
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, elem := range c.items {
		if c.onEvict != nil {
			entry := elem.Value.(*Entry)
			c.onEvict(key, entry.Value)
		}
		delete(c.items, key)
	}

	c.evictList.Init()
	c.currentSize = 0
	atomic.StoreInt64(&c.stats.EntryCount, 0)

	return nil
}

// Stats returns cache statistics
func (c *LRUCache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		Hits:       atomic.LoadUint64(&c.stats.Hits),
		Misses:     atomic.LoadUint64(&c.stats.Misses),
		Sets:       atomic.LoadUint64(&c.stats.Sets),
		Deletes:    atomic.LoadUint64(&c.stats.Deletes),
		Evictions:  atomic.LoadUint64(&c.stats.Evictions),
		Size:       c.currentSize,
		MaxSize:    c.maxSize,
		EntryCount: int64(c.evictList.Len()),
	}
}

// evictOldest removes the least recently used entry
func (c *LRUCache) evictOldest() {
	elem := c.evictList.Back()
	if elem != nil {
		c.removeElement(elem)
		atomic.AddUint64(&c.stats.Evictions, 1)
	}
}

// removeElement removes an element from the cache
func (c *LRUCache) removeElement(elem *list.Element) {
	entry := c.evictList.Remove(elem).(*Entry)
	delete(c.items, entry.Key)
	c.currentSize -= entry.Size
	atomic.AddInt64(&c.stats.EntryCount, -1)

	if c.onEvict != nil {
		c.onEvict(entry.Key, entry.Value)
	}
}

// cleanup periodically removes expired entries
func (c *LRUCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for _, elem := range c.items {
			entry := elem.Value.(*Entry)
			if entry.IsExpired() {
				c.removeElement(elem)
			}
		}
		c.mu.Unlock()
	}
}

// estimateSize estimates the memory size of a value
func estimateSize(value interface{}) int64 {
	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case int, int32, int64, float32, float64, bool:
		return 8
	default:
		// Estimate by marshaling to JSON
		data, err := json.Marshal(v)
		if err != nil {
			return 64 // Default estimate
		}
		return int64(len(data))
	}
}

// ========================================
// HTTP Cache Middleware
// ========================================

// HTTPCacheConfig configures HTTP caching behavior
type HTTPCacheConfig struct {
	DefaultTTL     time.Duration
	MaxAge         int
	SharedMaxAge   int
	Private        bool
	NoStore        bool
	NoCache        bool
	MustRevalidate bool
	ETagEnabled    bool
	VaryHeaders    []string
}

// DefaultHTTPCacheConfig returns the default HTTP cache configuration
func DefaultHTTPCacheConfig() HTTPCacheConfig {
	return HTTPCacheConfig{
		DefaultTTL:  5 * time.Minute,
		MaxAge:      300, // 5 minutes
		ETagEnabled: true,
		VaryHeaders: []string{"Accept", "Accept-Encoding"},
	}
}

// HTTPCache provides HTTP response caching
type HTTPCache struct {
	cache  *LRUCache
	config HTTPCacheConfig
}

// NewHTTPCache creates a new HTTP cache
func NewHTTPCache(config HTTPCacheConfig, opts ...LRUOption) *HTTPCache {
	return &HTTPCache{
		cache:  NewLRUCache(opts...),
		config: config,
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	ETag       string
	CreatedAt  time.Time
}

// GenerateETag generates an ETag for the response body
func GenerateETag(body []byte) string {
	hash := sha256.Sum256(body)
	return `"` + hex.EncodeToString(hash[:8]) + `"`
}

// GenerateCacheKey generates a cache key from the request
func GenerateCacheKey(r *http.Request) string {
	// Include method, path, and query
	key := r.Method + ":" + r.URL.Path
	if r.URL.RawQuery != "" {
		key += "?" + r.URL.RawQuery
	}
	return key
}

// Get retrieves a cached response
func (hc *HTTPCache) Get(key string) (*CachedResponse, bool) {
	value, ok := hc.cache.Get(key)
	if !ok {
		return nil, false
	}
	resp, ok := value.(*CachedResponse)
	return resp, ok
}

// Set caches a response
func (hc *HTTPCache) Set(key string, resp *CachedResponse, ttl time.Duration) error {
	if ttl == 0 {
		ttl = hc.config.DefaultTTL
	}
	return hc.cache.Set(key, resp, ttl)
}

// SetCacheHeaders sets appropriate cache headers on the response
func (hc *HTTPCache) SetCacheHeaders(w http.ResponseWriter, etag string) {
	// Cache-Control header
	var cacheControl []string

	if hc.config.NoStore {
		cacheControl = append(cacheControl, "no-store")
	} else if hc.config.NoCache {
		cacheControl = append(cacheControl, "no-cache")
	} else {
		if hc.config.Private {
			cacheControl = append(cacheControl, "private")
		} else {
			cacheControl = append(cacheControl, "public")
		}

		if hc.config.MaxAge > 0 {
			cacheControl = append(cacheControl, fmt.Sprintf("max-age=%d", hc.config.MaxAge))
		}

		if hc.config.SharedMaxAge > 0 {
			cacheControl = append(cacheControl, fmt.Sprintf("s-maxage=%d", hc.config.SharedMaxAge))
		}

		if hc.config.MustRevalidate {
			cacheControl = append(cacheControl, "must-revalidate")
		}
	}

	if len(cacheControl) > 0 {
		w.Header().Set("Cache-Control", joinStrings(cacheControl, ", "))
	}

	// ETag header
	if hc.config.ETagEnabled && etag != "" {
		w.Header().Set("ETag", etag)
	}

	// Vary header
	if len(hc.config.VaryHeaders) > 0 {
		w.Header().Set("Vary", joinStrings(hc.config.VaryHeaders, ", "))
	}
}

// CheckETagMatch checks if the request's If-None-Match header matches the ETag
func CheckETagMatch(r *http.Request, etag string) bool {
	ifNoneMatch := r.Header.Get("If-None-Match")
	return ifNoneMatch != "" && ifNoneMatch == etag
}

// Middleware returns an HTTP middleware that caches responses
func (hc *HTTPCache) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only cache GET requests
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			key := GenerateCacheKey(r)

			// Check cache
			if cached, ok := hc.Get(key); ok {
				// Check ETag
				if CheckETagMatch(r, cached.ETag) {
					w.WriteHeader(http.StatusNotModified)
					return
				}

				// Serve cached response
				for k, v := range cached.Headers {
					for _, val := range v {
						w.Header().Add(k, val)
					}
				}
				hc.SetCacheHeaders(w, cached.ETag)
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
				return
			}

			// Capture response
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           make([]byte, 0),
			}

			next.ServeHTTP(recorder, r)

			// Cache successful responses
			if recorder.statusCode == http.StatusOK {
				etag := ""
				if hc.config.ETagEnabled {
					etag = GenerateETag(recorder.body)
				}

				cached := &CachedResponse{
					StatusCode: recorder.statusCode,
					Headers:    recorder.Header().Clone(),
					Body:       recorder.body,
					ETag:       etag,
					CreatedAt:  time.Now(),
				}

				hc.Set(key, cached, 0)
				hc.SetCacheHeaders(w, etag)
			}
		})
	}
}

// responseRecorder captures the response
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

// Stats returns cache statistics
func (hc *HTTPCache) Stats() Stats {
	return hc.cache.Stats()
}

// Clear clears the cache
func (hc *HTTPCache) Clear() error {
	return hc.cache.Clear()
}

// Invalidate removes a specific key from the cache
func (hc *HTTPCache) Invalidate(key string) error {
	return hc.cache.Delete(key)
}

// InvalidateByPrefix removes all entries with keys starting with the prefix
func (hc *HTTPCache) InvalidateByPrefix(prefix string) int {
	hc.cache.mu.Lock()
	defer hc.cache.mu.Unlock()

	count := 0
	for key, elem := range hc.cache.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			hc.cache.removeElement(elem)
			count++
		}
	}

	return count
}

// Helper function
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// ========================================
// Cache Key Builder
// ========================================

// KeyBuilder helps build cache keys
type KeyBuilder struct {
	parts []string
}

// NewKeyBuilder creates a new key builder
func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{
		parts: make([]string, 0),
	}
}

// Add adds a part to the key
func (kb *KeyBuilder) Add(part string) *KeyBuilder {
	kb.parts = append(kb.parts, part)
	return kb
}

// AddIfNotEmpty adds a part if not empty
func (kb *KeyBuilder) AddIfNotEmpty(part string) *KeyBuilder {
	if part != "" {
		kb.parts = append(kb.parts, part)
	}
	return kb
}

// Build builds the final cache key
func (kb *KeyBuilder) Build() string {
	return joinStrings(kb.parts, ":")
}

// ========================================
// Global Cache Instance
// ========================================

var (
	globalCache     *LRUCache
	globalCacheOnce sync.Once
)

// Global returns the global cache instance
func Global() *LRUCache {
	globalCacheOnce.Do(func() {
		globalCache = NewLRUCache(
			WithCapacity(10000),
			WithDefaultTTL(5*time.Minute),
			WithMaxSize(100*1024*1024), // 100MB
		)
	})
	return globalCache
}

// Get retrieves a value from the global cache
func Get(key string) (interface{}, bool) {
	return Global().Get(key)
}

// Set adds a value to the global cache
func Set(key string, value interface{}, ttl time.Duration) error {
	return Global().Set(key, value, ttl)
}

// Delete removes a value from the global cache
func Delete(key string) error {
	return Global().Delete(key)
}
