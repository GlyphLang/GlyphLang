package vectordb

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient() *Client {
	return NewClient(NewMemoryProvider(), 3)
}

func TestNewClient(t *testing.T) {
	client := newTestClient()
	assert.Equal(t, 3, client.Dimension())
}

func TestUpsert(t *testing.T) {
	client := newTestClient()
	err := client.Upsert(UpsertRequest{
		ID:       "doc1",
		Vector:   Vector{0.1, 0.2, 0.3},
		Metadata: map[string]interface{}{"title": "test"},
	})
	require.NoError(t, err)
}

func TestUpsertDimensionMismatch(t *testing.T) {
	client := newTestClient()
	err := client.Upsert(UpsertRequest{
		ID:     "doc1",
		Vector: Vector{0.1, 0.2},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dimension mismatch")
}

func TestUpsertBatch(t *testing.T) {
	client := newTestClient()
	err := client.UpsertBatch([]UpsertRequest{
		{ID: "doc1", Vector: Vector{0.1, 0.2, 0.3}},
		{ID: "doc2", Vector: Vector{0.4, 0.5, 0.6}},
		{ID: "doc3", Vector: Vector{0.7, 0.8, 0.9}},
	})
	require.NoError(t, err)
	mem := client.provider.(*MemoryProvider)
	assert.Equal(t, 3, mem.Count())
}

func TestUpsertBatchDimensionMismatch(t *testing.T) {
	client := newTestClient()
	err := client.UpsertBatch([]UpsertRequest{
		{ID: "doc1", Vector: Vector{0.1, 0.2, 0.3}},
		{ID: "doc2", Vector: Vector{0.4, 0.5}},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dimension mismatch")
}

func TestQueryCosine(t *testing.T) {
	client := newTestClient()
	require.NoError(t, client.UpsertBatch([]UpsertRequest{
		{ID: "doc1", Vector: Vector{1.0, 0.0, 0.0}},
		{ID: "doc2", Vector: Vector{0.0, 1.0, 0.0}},
		{ID: "doc3", Vector: Vector{0.9, 0.1, 0.0}},
	}))
	results, err := client.Query(QueryRequest{
		Vector: Vector{1.0, 0.0, 0.0}, TopK: 2, Distance: Cosine,
	})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "doc1", results[0].ID)
	assert.Equal(t, "doc3", results[1].ID)
	assert.InDelta(t, 1.0, results[0].Score, 0.001)
}

func TestQueryL2(t *testing.T) {
	client := newTestClient()
	require.NoError(t, client.UpsertBatch([]UpsertRequest{
		{ID: "near", Vector: Vector{1.0, 1.0, 1.0}},
		{ID: "far", Vector: Vector{10.0, 10.0, 10.0}},
	}))
	results, err := client.Query(QueryRequest{
		Vector: Vector{1.1, 1.1, 1.1}, TopK: 2, Distance: L2,
	})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "near", results[0].ID)
}

func TestQueryDotProduct(t *testing.T) {
	client := newTestClient()
	require.NoError(t, client.UpsertBatch([]UpsertRequest{
		{ID: "high", Vector: Vector{1.0, 1.0, 1.0}},
		{ID: "low", Vector: Vector{0.1, 0.1, 0.1}},
	}))
	results, err := client.Query(QueryRequest{
		Vector: Vector{1.0, 1.0, 1.0}, TopK: 2, Distance: DotProduct,
	})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "high", results[0].ID)
}

func TestQueryWithFilter(t *testing.T) {
	client := newTestClient()
	require.NoError(t, client.UpsertBatch([]UpsertRequest{
		{ID: "a1", Vector: Vector{1.0, 0.0, 0.0}, Metadata: map[string]interface{}{"category": "books"}},
		{ID: "a2", Vector: Vector{0.9, 0.1, 0.0}, Metadata: map[string]interface{}{"category": "books"}},
		{ID: "b1", Vector: Vector{0.95, 0.05, 0.0}, Metadata: map[string]interface{}{"category": "movies"}},
	}))
	results, err := client.Query(QueryRequest{
		Vector: Vector{1.0, 0.0, 0.0}, TopK: 10, Distance: Cosine,
		Filter: map[string]interface{}{"category": "books"},
	})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "a1", results[0].ID)
}

func TestQueryDimensionMismatch(t *testing.T) {
	client := newTestClient()
	_, err := client.Query(QueryRequest{Vector: Vector{1.0, 0.0}, TopK: 5})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dimension mismatch")
}

func TestQueryDefaultTopK(t *testing.T) {
	client := newTestClient()
	for i := 0; i < 15; i++ {
		require.NoError(t, client.Upsert(UpsertRequest{
			ID: fmt.Sprintf("doc%d", i), Vector: Vector{float64(i), 0.0, 0.0},
		}))
	}
	results, err := client.Query(QueryRequest{Vector: Vector{1.0, 0.0, 0.0}})
	require.NoError(t, err)
	assert.Len(t, results, 10)
}

func TestDelete(t *testing.T) {
	client := newTestClient()
	require.NoError(t, client.Upsert(UpsertRequest{ID: "doc1", Vector: Vector{0.1, 0.2, 0.3}}))
	require.NoError(t, client.Delete("doc1"))
	mem := client.provider.(*MemoryProvider)
	assert.Equal(t, 0, mem.Count())
}

func TestUpsertOverwrite(t *testing.T) {
	client := newTestClient()
	require.NoError(t, client.Upsert(UpsertRequest{
		ID: "doc1", Vector: Vector{0.1, 0.2, 0.3}, Metadata: map[string]interface{}{"v": "1"},
	}))
	require.NoError(t, client.Upsert(UpsertRequest{
		ID: "doc1", Vector: Vector{0.4, 0.5, 0.6}, Metadata: map[string]interface{}{"v": "2"},
	}))
	mem := client.provider.(*MemoryProvider)
	assert.Equal(t, 1, mem.Count())
}

func TestClose(t *testing.T) {
	err := newTestClient().Close()
	assert.NoError(t, err)
}

func TestDistanceMetricString(t *testing.T) {
	assert.Equal(t, "cosine", Cosine.String())
	assert.Equal(t, "l2", L2.String())
	assert.Equal(t, "dot_product", DotProduct.String())
	assert.Equal(t, "unknown", DistanceMetric(99).String())
}

func TestCosineSimilarityDirect(t *testing.T) {
	assert.InDelta(t, 1.0, cosineSimilarity(Vector{1, 0, 0}, Vector{1, 0, 0}), 0.001)
	assert.InDelta(t, 0.0, cosineSimilarity(Vector{1, 0, 0}, Vector{0, 1, 0}), 0.001)
	assert.InDelta(t, -1.0, cosineSimilarity(Vector{1, 0, 0}, Vector{-1, 0, 0}), 0.001)
	assert.Equal(t, 0.0, cosineSimilarity(Vector{}, Vector{}))
	assert.Equal(t, 0.0, cosineSimilarity(Vector{1, 0}, Vector{1, 0, 0}))
}

func TestL2DistanceDirect(t *testing.T) {
	assert.InDelta(t, 0.0, l2Distance(Vector{1, 2, 3}, Vector{1, 2, 3}), 0.001)
	assert.InDelta(t, math.Sqrt(3), l2Distance(Vector{0, 0, 0}, Vector{1, 1, 1}), 0.001)
	assert.True(t, math.IsInf(l2Distance(Vector{1, 0}, Vector{1, 0, 0}), 1))
}

func TestDotProductDirect(t *testing.T) {
	assert.InDelta(t, 3.0, dotProduct(Vector{1, 1, 1}, Vector{1, 1, 1}), 0.001)
	assert.InDelta(t, 0.0, dotProduct(Vector{1, 0, 0}, Vector{0, 1, 0}), 0.001)
	assert.Equal(t, 0.0, dotProduct(Vector{1, 0}, Vector{1, 0, 0}))
}

func TestMatchesFilterDirect(t *testing.T) {
	md := map[string]interface{}{"category": "books", "year": 2024}
	assert.True(t, matchesFilter(md, nil))
	assert.True(t, matchesFilter(md, map[string]interface{}{}))
	assert.True(t, matchesFilter(md, map[string]interface{}{"category": "books"}))
	assert.False(t, matchesFilter(md, map[string]interface{}{"category": "movies"}))
	assert.False(t, matchesFilter(md, map[string]interface{}{"missing": "key"}))
}

func TestMemoryProviderCount(t *testing.T) {
	p := NewMemoryProvider()
	assert.Equal(t, 0, p.Count())
	require.NoError(t, p.Upsert(UpsertRequest{ID: "1", Vector: Vector{1}}))
	assert.Equal(t, 1, p.Count())
	require.NoError(t, p.Delete("1"))
	assert.Equal(t, 0, p.Count())
}
