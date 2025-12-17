package database

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests require a real PostgreSQL database
// They will be skipped if DATABASE_URL environment variable is not set

func getTestDBConfig() (*Config, bool) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, false
	}

	cfg, err := ParseConnectionString(connStr)
	if err != nil {
		return nil, false
	}

	return cfg, true
}

func TestPostgresDB_Connect(t *testing.T) {
	cfg, ok := getTestDBConfig()
	if !ok {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	db := NewPostgresDB(cfg)
	ctx := context.Background()

	err := db.Connect(ctx)
	require.NoError(t, err)

	defer db.Close()

	// Verify connection is alive
	err = db.Ping(ctx)
	assert.NoError(t, err)
}

func TestPostgresDB_CreateDropTable(t *testing.T) {
	cfg, ok := getTestDBConfig()
	if !ok {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	db := NewPostgresDB(cfg)
	ctx := context.Background()

	err := db.Connect(ctx)
	require.NoError(t, err)
	defer db.Close()

	tableName := "test_users"

	// Drop table if exists (cleanup from previous runs)
	db.DropTable(ctx, tableName)

	// Create table
	schema := map[string]string{
		"id":         "SERIAL PRIMARY KEY",
		"name":       "VARCHAR(100) NOT NULL",
		"email":      "VARCHAR(100)",
		"created_at": "TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}

	err = db.CreateTable(ctx, tableName, schema)
	assert.NoError(t, err)

	// Check if table exists
	exists, err := db.TableExists(ctx, tableName)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Drop table
	err = db.DropTable(ctx, tableName)
	assert.NoError(t, err)

	// Verify table is dropped
	exists, err = db.TableExists(ctx, tableName)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestPostgresDB_CRUD(t *testing.T) {
	cfg, ok := getTestDBConfig()
	if !ok {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	db := NewPostgresDB(cfg)
	ctx := context.Background()

	err := db.Connect(ctx)
	require.NoError(t, err)
	defer db.Close()

	tableName := "test_users"

	// Setup: Create table
	db.DropTable(ctx, tableName)
	schema := map[string]string{
		"id":    "SERIAL PRIMARY KEY",
		"name":  "VARCHAR(100) NOT NULL",
		"email": "VARCHAR(100)",
	}
	err = db.CreateTable(ctx, tableName, schema)
	require.NoError(t, err)

	defer db.DropTable(ctx, tableName)

	// Create ORM instance
	orm := NewORM(db, tableName)

	// Test Create
	user := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	created, err := orm.Create(ctx, user)
	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotNil(t, created["id"])

	userID := created["id"]

	// Test FindByID
	found, err := orm.FindByID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", found["name"])

	// Test Update
	updates := map[string]interface{}{
		"name": "Jane Doe",
	}

	updated, err := orm.Update(ctx, userID, updates)
	require.NoError(t, err)
	assert.Equal(t, "Jane Doe", updated["name"])

	// Test Delete
	err = orm.Delete(ctx, userID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = orm.FindByID(ctx, userID)
	assert.Error(t, err)
}

func TestPostgresDB_Query(t *testing.T) {
	cfg, ok := getTestDBConfig()
	if !ok {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	db := NewPostgresDB(cfg)
	ctx := context.Background()

	err := db.Connect(ctx)
	require.NoError(t, err)
	defer db.Close()

	tableName := "test_products"

	// Setup
	db.DropTable(ctx, tableName)
	schema := map[string]string{
		"id":    "SERIAL PRIMARY KEY",
		"name":  "VARCHAR(100) NOT NULL",
		"price": "DECIMAL(10, 2)",
	}
	err = db.CreateTable(ctx, tableName, schema)
	require.NoError(t, err)

	defer db.DropTable(ctx, tableName)

	orm := NewORM(db, tableName)

	// Insert test data
	products := []map[string]interface{}{
		{"name": "Product A", "price": 10.50},
		{"name": "Product B", "price": 25.00},
		{"name": "Product C", "price": 15.75},
	}

	for _, product := range products {
		_, err := orm.Create(ctx, product)
		require.NoError(t, err)
	}

	// Test QueryBuilder
	results, err := orm.NewQueryBuilder().
		Where("price", ">", 15.00).
		OrderBy("price", "ASC").
		Get(ctx)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)
}

func TestPostgresDB_Transaction(t *testing.T) {
	cfg, ok := getTestDBConfig()
	if !ok {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	db := NewPostgresDB(cfg)
	ctx := context.Background()

	err := db.Connect(ctx)
	require.NoError(t, err)
	defer db.Close()

	tableName := "test_accounts"

	// Setup
	db.DropTable(ctx, tableName)
	schema := map[string]string{
		"id":      "SERIAL PRIMARY KEY",
		"name":    "VARCHAR(100) NOT NULL",
		"balance": "DECIMAL(10, 2)",
	}
	err = db.CreateTable(ctx, tableName, schema)
	require.NoError(t, err)

	defer db.DropTable(ctx, tableName)

	// Test successful transaction
	err = db.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO "+tableName+" (name, balance) VALUES ($1, $2)", "Account1", 100.00)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO "+tableName+" (name, balance) VALUES ($1, $2)", "Account2", 200.00)
		return err
	})

	assert.NoError(t, err)

	// Verify data was committed
	orm := NewORM(db, tableName)
	count, err := orm.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestPostgresDB_BulkInsert(t *testing.T) {
	cfg, ok := getTestDBConfig()
	if !ok {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	db := NewPostgresDB(cfg)
	ctx := context.Background()

	err := db.Connect(ctx)
	require.NoError(t, err)
	defer db.Close()

	tableName := "test_bulk"

	// Setup
	db.DropTable(ctx, tableName)
	schema := map[string]string{
		"id":   "SERIAL PRIMARY KEY",
		"name": "VARCHAR(100) NOT NULL",
		"code": "VARCHAR(10)",
	}
	err = db.CreateTable(ctx, tableName, schema)
	require.NoError(t, err)

	defer db.DropTable(ctx, tableName)

	// Bulk insert
	columns := []string{"name", "code"}
	values := [][]interface{}{
		{"Item1", "A1"},
		{"Item2", "A2"},
		{"Item3", "A3"},
	}

	err = db.BulkInsert(ctx, tableName, columns, values)
	assert.NoError(t, err)

	// Verify data
	orm := NewORM(db, tableName)
	count, err := orm.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestPostgresDB_ConnectionPool(t *testing.T) {
	cfg, ok := getTestDBConfig()
	if !ok {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	cfg.MaxOpenConns = 5
	cfg.MaxIdleConns = 2

	db := NewPostgresDB(cfg)
	ctx := context.Background()

	err := db.Connect(ctx)
	require.NoError(t, err)
	defer db.Close()

	// Check stats
	stats := db.Stats()
	assert.Equal(t, 5, stats.MaxOpenConnections)
}

func TestPostgresDB_ContextTimeout(t *testing.T) {
	cfg, ok := getTestDBConfig()
	if !ok {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	db := NewPostgresDB(cfg)
	ctx := context.Background()

	err := db.Connect(ctx)
	require.NoError(t, err)
	defer db.Close()

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// This should fail due to timeout
	time.Sleep(10 * time.Millisecond)
	err = db.Ping(ctx)
	assert.Error(t, err)
}
