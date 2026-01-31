package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockHandler_PingAndClose(t *testing.T) {
	h := NewMockHandler()
	assert.NoError(t, h.Ping())
	assert.NoError(t, h.Close())
}

func TestMockHandler_InsertOneAndFindOne(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	id, err := coll.InsertOne(map[string]interface{}{"name": "Alice", "age": 30})
	require.NoError(t, err)
	assert.NotNil(t, id)

	doc, err := coll.FindOne(map[string]interface{}{"name": "Alice"})
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, "Alice", doc["name"])
	assert.Equal(t, 30, doc["age"])
}

func TestMockHandler_FindOneNotFound(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	doc, err := coll.FindOne(map[string]interface{}{"name": "Nobody"})
	require.NoError(t, err)
	assert.Nil(t, doc)
}

func TestMockHandler_Find(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	_, _ = coll.InsertOne(map[string]interface{}{"name": "Alice", "role": "admin"})
	_, _ = coll.InsertOne(map[string]interface{}{"name": "Bob", "role": "user"})
	_, _ = coll.InsertOne(map[string]interface{}{"name": "Charlie", "role": "admin"})

	// Find all admins
	docs, err := coll.Find(map[string]interface{}{"role": "admin"})
	require.NoError(t, err)
	assert.Len(t, docs, 2)

	// Find all with empty filter
	allDocs, err := coll.Find(map[string]interface{}{})
	require.NoError(t, err)
	assert.Len(t, allDocs, 3)
}

func TestMockHandler_InsertMany(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("items")

	docs := []map[string]interface{}{
		{"name": "item1"},
		{"name": "item2"},
		{"name": "item3"},
	}

	ids, err := coll.InsertMany(docs)
	require.NoError(t, err)
	assert.Len(t, ids, 3)

	count, err := coll.CountDocuments(map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestMockHandler_UpdateOne(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	_, _ = coll.InsertOne(map[string]interface{}{"name": "Alice", "age": 30})

	modified, err := coll.UpdateOne(
		map[string]interface{}{"name": "Alice"},
		map[string]interface{}{"age": 31},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(1), modified)

	doc, _ := coll.FindOne(map[string]interface{}{"name": "Alice"})
	assert.Equal(t, 31, doc["age"])
}

func TestMockHandler_UpdateOneNoMatch(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	modified, err := coll.UpdateOne(
		map[string]interface{}{"name": "Nobody"},
		map[string]interface{}{"age": 99},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(0), modified)
}

func TestMockHandler_UpdateMany(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	_, _ = coll.InsertOne(map[string]interface{}{"name": "Alice", "role": "user"})
	_, _ = coll.InsertOne(map[string]interface{}{"name": "Bob", "role": "user"})
	_, _ = coll.InsertOne(map[string]interface{}{"name": "Charlie", "role": "admin"})

	modified, err := coll.UpdateMany(
		map[string]interface{}{"role": "user"},
		map[string]interface{}{"active": true},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(2), modified)
}

func TestMockHandler_DeleteOne(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	_, _ = coll.InsertOne(map[string]interface{}{"name": "Alice"})
	_, _ = coll.InsertOne(map[string]interface{}{"name": "Bob"})

	deleted, err := coll.DeleteOne(map[string]interface{}{"name": "Alice"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	count, _ := coll.CountDocuments(map[string]interface{}{})
	assert.Equal(t, int64(1), count)
}

func TestMockHandler_DeleteOneNoMatch(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	deleted, err := coll.DeleteOne(map[string]interface{}{"name": "Nobody"})
	require.NoError(t, err)
	assert.Equal(t, int64(0), deleted)
}

func TestMockHandler_DeleteMany(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	_, _ = coll.InsertOne(map[string]interface{}{"name": "Alice", "role": "admin"})
	_, _ = coll.InsertOne(map[string]interface{}{"name": "Bob", "role": "user"})
	_, _ = coll.InsertOne(map[string]interface{}{"name": "Charlie", "role": "admin"})

	deleted, err := coll.DeleteMany(map[string]interface{}{"role": "admin"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), deleted)

	count, _ := coll.CountDocuments(map[string]interface{}{})
	assert.Equal(t, int64(1), count)
}

func TestMockHandler_CountDocuments(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	count, err := coll.CountDocuments(map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	_, _ = coll.InsertOne(map[string]interface{}{"name": "Alice"})
	_, _ = coll.InsertOne(map[string]interface{}{"name": "Bob"})

	count, err = coll.CountDocuments(map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestMockHandler_Aggregate(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("orders")

	_, _ = coll.InsertOne(map[string]interface{}{"product": "A", "qty": 10})
	_, _ = coll.InsertOne(map[string]interface{}{"product": "B", "qty": 20})

	results, err := coll.Aggregate([]map[string]interface{}{})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockHandler_CreateIndex(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	name, err := coll.CreateIndex(map[string]interface{}{"email": 1}, true)
	require.NoError(t, err)
	assert.Contains(t, name, "email")
}

func TestMockHandler_DropIndex(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	err := coll.DropIndex("some_index")
	assert.NoError(t, err)
}

func TestMockHandler_SeparateCollections(t *testing.T) {
	h := NewMockHandler()

	users := h.Collection("users")
	posts := h.Collection("posts")

	_, _ = users.InsertOne(map[string]interface{}{"name": "Alice"})
	_, _ = posts.InsertOne(map[string]interface{}{"title": "Hello"})
	_, _ = posts.InsertOne(map[string]interface{}{"title": "World"})

	userCount, _ := users.CountDocuments(map[string]interface{}{})
	postCount, _ := posts.CountDocuments(map[string]interface{}{})

	assert.Equal(t, int64(1), userCount)
	assert.Equal(t, int64(2), postCount)
}

func TestMockHandler_InsertOneWithExplicitID(t *testing.T) {
	h := NewMockHandler()
	coll := h.Collection("users")

	id, err := coll.InsertOne(map[string]interface{}{"_id": "custom-id", "name": "Alice"})
	require.NoError(t, err)
	assert.Equal(t, "custom-id", id)

	doc, _ := coll.FindOne(map[string]interface{}{"_id": "custom-id"})
	assert.Equal(t, "Alice", doc["name"])
}

func TestMockHandler_CollectionReuse(t *testing.T) {
	h := NewMockHandler()

	coll1 := h.Collection("users")
	_, _ = coll1.InsertOne(map[string]interface{}{"name": "Alice"})

	coll2 := h.Collection("users")
	count, _ := coll2.CountDocuments(map[string]interface{}{})
	assert.Equal(t, int64(1), count)
}
