// Package vectordb provides a unified interface for vector database operations
// including embedding storage, similarity search, and metadata filtering.
// It supports multiple backends: in-memory (for testing), Pinecone, Qdrant,
// and pgvector (PostgreSQL extension).
package vectordb

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

// Vector represents an embedding vector
type Vector []float64

// Document represents a stored vector with metadata
type Document struct {
	ID       string
	Vector   Vector
	Metadata map[string]interface{}
	Score    float64 // Populated during search results
}

// QueryRequest represents a similarity search request
type QueryRequest struct {
	Vector   Vector
	TopK     int
	Filter   map[string]interface{} // Metadata filter
	Distance DistanceMetric
}

// UpsertRequest represents a document upsert operation
type UpsertRequest struct {
	ID       string
	Vector   Vector
	Metadata map[string]interface{}
}

// DistanceMetric defines the distance function for similarity search
type DistanceMetric int

const (
	// Cosine computes cosine similarity (1 - cosine distance)
	Cosine DistanceMetric = iota
	// L2 computes Euclidean (L2) distance
	L2
	// DotProduct computes dot product similarity
	DotProduct
)

func (d DistanceMetric) String() string {
	switch d {
	case Cosine:
		return "cosine"
	case L2:
		return "l2"
	case DotProduct:
		return "dot_product"
	default:
		return "unknown"
	}
}

// Provider defines the interface for vector database backends
type Provider interface {
	// Upsert inserts or updates a document
	Upsert(req UpsertRequest) error
	// UpsertBatch inserts or updates multiple documents
	UpsertBatch(reqs []UpsertRequest) error
	// Query performs a similarity search
	Query(req QueryRequest) ([]Document, error)
	// Delete removes a document by ID
	Delete(id string) error
	// Close releases any resources held by the provider
	Close() error
}

// Client is the main vector database client that wraps a provider
type Client struct {
	provider Provider
	dimension int
}

// NewClient creates a new vector database client with the given provider and dimension
func NewClient(provider Provider, dimension int) *Client {
	return &Client{
		provider:  provider,
		dimension: dimension,
	}
}

// Upsert inserts or updates a document, validating vector dimensions
func (c *Client) Upsert(req UpsertRequest) error {
	if len(req.Vector) != c.dimension {
		return fmt.Errorf("vector dimension mismatch: expected %d, got %d", c.dimension, len(req.Vector))
	}
	return c.provider.Upsert(req)
}

// UpsertBatch inserts or updates multiple documents
func (c *Client) UpsertBatch(reqs []UpsertRequest) error {
	for _, req := range reqs {
		if len(req.Vector) != c.dimension {
			return fmt.Errorf("vector dimension mismatch for ID %s: expected %d, got %d",
				req.ID, c.dimension, len(req.Vector))
		}
	}
	return c.provider.UpsertBatch(reqs)
}

// Query performs a similarity search
func (c *Client) Query(req QueryRequest) ([]Document, error) {
	if len(req.Vector) != c.dimension {
		return nil, fmt.Errorf("query vector dimension mismatch: expected %d, got %d",
			c.dimension, len(req.Vector))
	}
	if req.TopK <= 0 {
		req.TopK = 10
	}
	return c.provider.Query(req)
}

// Delete removes a document by ID
func (c *Client) Delete(id string) error {
	return c.provider.Delete(id)
}

// Close releases provider resources
func (c *Client) Close() error {
	return c.provider.Close()
}

// Dimension returns the configured vector dimension
func (c *Client) Dimension() int {
	return c.dimension
}

// MemoryProvider is an in-memory vector database implementation for testing and development
type MemoryProvider struct {
	mu        sync.RWMutex
	documents map[string]*Document
}

// NewMemoryProvider creates a new in-memory vector database
func NewMemoryProvider() *MemoryProvider {
	return &MemoryProvider{
		documents: make(map[string]*Document),
	}
}

// Upsert inserts or updates a document in memory
func (m *MemoryProvider) Upsert(req UpsertRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.documents[req.ID] = &Document{
		ID:       req.ID,
		Vector:   req.Vector,
		Metadata: req.Metadata,
	}
	return nil
}

// UpsertBatch inserts or updates multiple documents in memory
func (m *MemoryProvider) UpsertBatch(reqs []UpsertRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, req := range reqs {
		m.documents[req.ID] = &Document{
			ID:       req.ID,
			Vector:   req.Vector,
			Metadata: req.Metadata,
		}
	}
	return nil
}

// Query performs a similarity search over in-memory documents
func (m *MemoryProvider) Query(req QueryRequest) ([]Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type scored struct {
		doc   *Document
		score float64
	}
	var results []scored

	for _, doc := range m.documents {
		// Apply metadata filter
		if !matchesFilter(doc.Metadata, req.Filter) {
			continue
		}

		score := computeDistance(req.Vector, doc.Vector, req.Distance)
		results = append(results, scored{doc: doc, score: score})
	}

	// Sort by score (lower is better for L2, higher is better for cosine/dot product)
	sort.Slice(results, func(i, j int) bool {
		if req.Distance == L2 {
			return results[i].score < results[j].score
		}
		return results[i].score > results[j].score
	})

	// Return top-K results
	if req.TopK > len(results) {
		req.TopK = len(results)
	}

	docs := make([]Document, req.TopK)
	for i := 0; i < req.TopK; i++ {
		docs[i] = *results[i].doc
		docs[i].Score = results[i].score
	}
	return docs, nil
}

// Delete removes a document by ID from memory
func (m *MemoryProvider) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.documents, id)
	return nil
}

// Close is a no-op for the in-memory provider
func (m *MemoryProvider) Close() error {
	return nil
}

// Count returns the number of documents stored (for testing)
func (m *MemoryProvider) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.documents)
}

// matchesFilter checks if document metadata matches the filter criteria
func matchesFilter(metadata, filter map[string]interface{}) bool {
	if len(filter) == 0 {
		return true
	}
	for key, expected := range filter {
		actual, exists := metadata[key]
		if !exists {
			return false
		}
		if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected) {
			return false
		}
	}
	return true
}

// computeDistance computes the distance/similarity between two vectors
func computeDistance(a, b Vector, metric DistanceMetric) float64 {
	switch metric {
	case Cosine:
		return cosineSimilarity(a, b)
	case L2:
		return l2Distance(a, b)
	case DotProduct:
		return dotProduct(a, b)
	default:
		return cosineSimilarity(a, b)
	}
}

// cosineSimilarity computes cosine similarity between two vectors
func cosineSimilarity(a, b Vector) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}

// l2Distance computes Euclidean distance between two vectors
func l2Distance(a, b Vector) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}
	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// dotProduct computes the dot product between two vectors
func dotProduct(a, b Vector) float64 {
	if len(a) != len(b) {
		return 0
	}
	var sum float64
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}
