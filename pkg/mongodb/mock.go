package mongodb

import (
	"fmt"
	"sync"
)

// MockHandler provides an in-memory MongoDB-like handler for testing.
type MockHandler struct {
	mu          sync.RWMutex
	collections map[string]*MockCollectionHandler
	nextID      int64
}

// NewMockHandler creates a new MockHandler.
func NewMockHandler() *MockHandler {
	return &MockHandler{
		collections: make(map[string]*MockCollectionHandler),
		nextID:      1,
	}
}

// Close is a no-op for the mock.
func (m *MockHandler) Close() error {
	return nil
}

// Ping always succeeds for the mock.
func (m *MockHandler) Ping() error {
	return nil
}

// Collection returns a MockCollectionHandler for the named collection.
func (m *MockHandler) Collection(name string) *MockCollectionHandler {
	m.mu.Lock()
	defer m.mu.Unlock()

	if coll, ok := m.collections[name]; ok {
		return coll
	}
	coll := &MockCollectionHandler{
		mock: m,
		docs: make([]map[string]interface{}, 0),
	}
	m.collections[name] = coll
	return coll
}

// MockCollectionHandler provides in-memory document operations.
type MockCollectionHandler struct {
	mock *MockHandler
	mu   sync.RWMutex
	docs []map[string]interface{}
}

// FindOne returns the first document matching the filter.
func (c *MockCollectionHandler) FindOne(filter map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, doc := range c.docs {
		if matchesFilter(doc, filter) {
			return copyDoc(doc), nil
		}
	}
	return nil, nil
}

// Find returns all documents matching the filter.
func (c *MockCollectionHandler) Find(filter map[string]interface{}) ([]map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make([]map[string]interface{}, 0)
	for _, doc := range c.docs {
		if matchesFilter(doc, filter) {
			results = append(results, copyDoc(doc))
		}
	}
	return results, nil
}

// InsertOne inserts a single document.
func (c *MockCollectionHandler) InsertOne(doc map[string]interface{}) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	docCopy := copyDoc(doc)
	if _, ok := docCopy["_id"]; !ok {
		c.mock.mu.Lock()
		docCopy["_id"] = fmt.Sprintf("mock_%d", c.mock.nextID)
		c.mock.nextID++
		c.mock.mu.Unlock()
	}
	c.docs = append(c.docs, docCopy)
	return docCopy["_id"], nil
}

// InsertMany inserts multiple documents.
func (c *MockCollectionHandler) InsertMany(docs []map[string]interface{}) ([]interface{}, error) {
	ids := make([]interface{}, len(docs))
	for i, doc := range docs {
		id, err := c.InsertOne(doc)
		if err != nil {
			return nil, err
		}
		ids[i] = id
	}
	return ids, nil
}

// UpdateOne updates the first matching document.
func (c *MockCollectionHandler) UpdateOne(filter map[string]interface{}, update map[string]interface{}) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, doc := range c.docs {
		if matchesFilter(doc, filter) {
			for k, v := range update {
				doc[k] = v
			}
			c.docs[i] = doc
			return 1, nil
		}
	}
	return 0, nil
}

// UpdateMany updates all matching documents.
func (c *MockCollectionHandler) UpdateMany(filter map[string]interface{}, update map[string]interface{}) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var count int64
	for i, doc := range c.docs {
		if matchesFilter(doc, filter) {
			for k, v := range update {
				doc[k] = v
			}
			c.docs[i] = doc
			count++
		}
	}
	return count, nil
}

// DeleteOne deletes the first matching document.
func (c *MockCollectionHandler) DeleteOne(filter map[string]interface{}) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, doc := range c.docs {
		if matchesFilter(doc, filter) {
			c.docs = append(c.docs[:i], c.docs[i+1:]...)
			return 1, nil
		}
	}
	return 0, nil
}

// DeleteMany deletes all matching documents.
func (c *MockCollectionHandler) DeleteMany(filter map[string]interface{}) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	remaining := make([]map[string]interface{}, 0)
	var count int64
	for _, doc := range c.docs {
		if matchesFilter(doc, filter) {
			count++
		} else {
			remaining = append(remaining, doc)
		}
	}
	c.docs = remaining
	return count, nil
}

// CountDocuments counts matching documents.
func (c *MockCollectionHandler) CountDocuments(filter map[string]interface{}) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var count int64
	for _, doc := range c.docs {
		if matchesFilter(doc, filter) {
			count++
		}
	}
	return count, nil
}

// Aggregate is a simplified mock that returns all documents (real aggregation is not supported in mock).
func (c *MockCollectionHandler) Aggregate(pipeline []map[string]interface{}) ([]map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make([]map[string]interface{}, len(c.docs))
	for i, doc := range c.docs {
		results[i] = copyDoc(doc)
	}
	return results, nil
}

// CreateIndex is a no-op in the mock. Returns a generated index name.
func (c *MockCollectionHandler) CreateIndex(keys map[string]interface{}, unique bool) (string, error) {
	name := "mock_index"
	for k := range keys {
		name += "_" + k
	}
	return name, nil
}

// DropIndex is a no-op in the mock.
func (c *MockCollectionHandler) DropIndex(name string) error {
	return nil
}

// matchesFilter checks if a document matches all key-value pairs in the filter.
func matchesFilter(doc, filter map[string]interface{}) bool {
	if len(filter) == 0 {
		return true
	}
	for k, v := range filter {
		docVal, ok := doc[k]
		if !ok || fmt.Sprintf("%v", docVal) != fmt.Sprintf("%v", v) {
			return false
		}
	}
	return true
}

// copyDoc creates a shallow copy of a document.
func copyDoc(doc map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{}, len(doc))
	for k, v := range doc {
		cp[k] = v
	}
	return cp
}
