package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Handler wraps a MongoDB client and provides database operations for GlyphLang.
type Handler struct {
	client   *mongo.Client
	database *mongo.Database
	ctx      context.Context
}

// NewHandler creates a Handler from an existing MongoDB client and database name.
func NewHandler(client *mongo.Client, dbName string) *Handler {
	return &Handler{
		client:   client,
		database: client.Database(dbName),
		ctx:      context.Background(),
	}
}

// NewHandlerFromURI creates a Handler by connecting to the given MongoDB URI.
func NewHandlerFromURI(uri string, dbName string) (*Handler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return NewHandler(client, dbName), nil
}

// Close disconnects the MongoDB client.
func (h *Handler) Close() error {
	return h.client.Disconnect(h.ctx)
}

// Ping verifies the MongoDB connection is alive.
func (h *Handler) Ping() error {
	ctx, cancel := context.WithTimeout(h.ctx, 5*time.Second)
	defer cancel()
	return h.client.Ping(ctx, nil)
}

// CollectionHandler provides operations on a single MongoDB collection.
type CollectionHandler struct {
	coll *mongo.Collection
	ctx  context.Context
}

// Collection returns a CollectionHandler for the named collection.
func (h *Handler) Collection(name string) *CollectionHandler {
	return &CollectionHandler{
		coll: h.database.Collection(name),
		ctx:  h.ctx,
	}
}

// FindOne returns the first document matching the filter.
func (c *CollectionHandler) FindOne(filter map[string]interface{}) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := c.coll.FindOne(ctx, toBsonDoc(filter)).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("findOne failed: %w", err)
	}
	return result, nil
}

// Find returns all documents matching the filter.
func (c *CollectionHandler) Find(filter map[string]interface{}) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	cursor, err := c.coll.Find(ctx, toBsonDoc(filter))
	if err != nil {
		return nil, fmt.Errorf("find failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("cursor decode failed: %w", err)
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	return results, nil
}

// InsertOne inserts a single document and returns the inserted ID.
func (c *CollectionHandler) InsertOne(doc map[string]interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	result, err := c.coll.InsertOne(ctx, toBsonDoc(doc))
	if err != nil {
		return nil, fmt.Errorf("insertOne failed: %w", err)
	}
	return result.InsertedID, nil
}

// InsertMany inserts multiple documents and returns the inserted IDs.
func (c *CollectionHandler) InsertMany(docs []map[string]interface{}) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	bsonDocs := make([]interface{}, len(docs))
	for i, doc := range docs {
		bsonDocs[i] = toBsonDoc(doc)
	}

	result, err := c.coll.InsertMany(ctx, bsonDocs)
	if err != nil {
		return nil, fmt.Errorf("insertMany failed: %w", err)
	}
	return result.InsertedIDs, nil
}

// UpdateOne updates the first document matching the filter.
func (c *CollectionHandler) UpdateOne(filter map[string]interface{}, update map[string]interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	result, err := c.coll.UpdateOne(ctx, toBsonDoc(filter), bson.M{"$set": toBsonDoc(update)})
	if err != nil {
		return 0, fmt.Errorf("updateOne failed: %w", err)
	}
	return result.ModifiedCount, nil
}

// UpdateMany updates all documents matching the filter.
func (c *CollectionHandler) UpdateMany(filter map[string]interface{}, update map[string]interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	result, err := c.coll.UpdateMany(ctx, toBsonDoc(filter), bson.M{"$set": toBsonDoc(update)})
	if err != nil {
		return 0, fmt.Errorf("updateMany failed: %w", err)
	}
	return result.ModifiedCount, nil
}

// DeleteOne deletes the first document matching the filter.
func (c *CollectionHandler) DeleteOne(filter map[string]interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	result, err := c.coll.DeleteOne(ctx, toBsonDoc(filter))
	if err != nil {
		return 0, fmt.Errorf("deleteOne failed: %w", err)
	}
	return result.DeletedCount, nil
}

// DeleteMany deletes all documents matching the filter.
func (c *CollectionHandler) DeleteMany(filter map[string]interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	result, err := c.coll.DeleteMany(ctx, toBsonDoc(filter))
	if err != nil {
		return 0, fmt.Errorf("deleteMany failed: %w", err)
	}
	return result.DeletedCount, nil
}

// CountDocuments counts documents matching the filter.
func (c *CollectionHandler) CountDocuments(filter map[string]interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	count, err := c.coll.CountDocuments(ctx, toBsonDoc(filter))
	if err != nil {
		return 0, fmt.Errorf("countDocuments failed: %w", err)
	}
	return count, nil
}

// Aggregate executes an aggregation pipeline.
func (c *CollectionHandler) Aggregate(pipeline []map[string]interface{}) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	bsonPipeline := make([]bson.M, len(pipeline))
	for i, stage := range pipeline {
		bsonPipeline[i] = toBsonDoc(stage)
	}

	cursor, err := c.coll.Aggregate(ctx, bsonPipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("aggregate cursor decode failed: %w", err)
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	return results, nil
}

// CreateIndex creates an index on the collection.
func (c *CollectionHandler) CreateIndex(keys map[string]interface{}, unique bool) (string, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    toBsonDoc(keys),
		Options: options.Index().SetUnique(unique),
	}

	name, err := c.coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return "", fmt.Errorf("createIndex failed: %w", err)
	}
	return name, nil
}

// DropIndex drops an index by name.
func (c *CollectionHandler) DropIndex(name string) error {
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	err := c.coll.Indexes().DropOne(ctx, name)
	if err != nil {
		return fmt.Errorf("dropIndex failed: %w", err)
	}
	return nil
}

// toBsonDoc converts a map[string]interface{} to bson.M for the MongoDB driver.
func toBsonDoc(m map[string]interface{}) bson.M {
	if m == nil {
		return bson.M{}
	}
	doc := bson.M{}
	for k, v := range m {
		doc[k] = v
	}
	return doc
}
