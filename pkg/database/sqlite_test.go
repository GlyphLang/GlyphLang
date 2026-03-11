package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newInMemorySQLite(t *testing.T) *SQLiteDB {
	t.Helper()
	db := NewSQLiteDB(&Config{Driver: "sqlite", Database: ":memory:"})
	err := db.Connect(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSQLiteDB_ConnectInMemory(t *testing.T) {
	db := NewSQLiteDB(&Config{Driver: "sqlite", Database: ":memory:"})
	err := db.Connect(context.Background())
	require.NoError(t, err)
	defer db.Close()

	assert.Equal(t, "sqlite", db.Driver())
}

func TestSQLiteDB_ConnectFileBased(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db := NewSQLiteDB(&Config{Driver: "sqlite", Database: dbPath})
	err := db.Connect(context.Background())
	require.NoError(t, err)
	defer db.Close()

	assert.FileExists(t, dbPath)
	assert.Equal(t, "sqlite", db.Driver())
}

func TestSQLiteDB_ConnectEmptyDatabase(t *testing.T) {
	// Empty database field should default to :memory:
	db := NewSQLiteDB(&Config{Driver: "sqlite", Database: ""})
	err := db.Connect(context.Background())
	require.NoError(t, err)
	defer db.Close()
}

func TestSQLiteDB_Ping(t *testing.T) {
	db := newInMemorySQLite(t)
	err := db.Ping(context.Background())
	require.NoError(t, err)
}

func TestSQLiteDB_PingNotConnected(t *testing.T) {
	db := NewSQLiteDB(&Config{Driver: "sqlite"})
	err := db.Ping(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestSQLiteDB_CloseNotConnected(t *testing.T) {
	db := NewSQLiteDB(&Config{Driver: "sqlite"})
	err := db.Close()
	assert.NoError(t, err)
}

func TestSQLiteDB_Stats(t *testing.T) {
	db := newInMemorySQLite(t)
	stats := db.Stats()
	assert.GreaterOrEqual(t, stats.OpenConnections, 0)
}

func TestSQLiteDB_StatsNotConnected(t *testing.T) {
	db := NewSQLiteDB(&Config{Driver: "sqlite"})
	stats := db.Stats()
	assert.Equal(t, 0, stats.OpenConnections)
}

func TestSQLiteDB_CreateTableAndInsert(t *testing.T) {
	db := newInMemorySQLite(t)
	ctx := context.Background()

	// Create table
	schema := map[string]string{
		"id":   "INTEGER",
		"name": "TEXT",
		"age":  "INTEGER",
	}
	err := db.CreateTable(ctx, "users", schema)
	require.NoError(t, err)

	// Verify table exists
	exists, err := db.TableExists(ctx, "users")
	require.NoError(t, err)
	assert.True(t, exists)

	// Insert data
	_, err = db.Exec(ctx, "INSERT INTO users (id, name, age) VALUES (?, ?, ?)", 1, "Alice", 30)
	require.NoError(t, err)

	// Query data
	var name string
	var age int
	err = db.QueryRow(ctx, "SELECT name, age FROM users WHERE id = ?", 1).Scan(&name, &age)
	require.NoError(t, err)
	assert.Equal(t, "Alice", name)
	assert.Equal(t, 30, age)
}

func TestSQLiteDB_Query(t *testing.T) {
	db := newInMemorySQLite(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE items (id INTEGER PRIMARY KEY, value TEXT)")
	require.NoError(t, err)

	_, err = db.Exec(ctx, "INSERT INTO items (value) VALUES (?), (?), (?)", "a", "b", "c")
	require.NoError(t, err)

	rows, err := db.Query(ctx, "SELECT value FROM items ORDER BY id")
	require.NoError(t, err)
	defer rows.Close()

	var values []string
	for rows.Next() {
		var v string
		err := rows.Scan(&v)
		require.NoError(t, err)
		values = append(values, v)
	}
	assert.Equal(t, []string{"a", "b", "c"}, values)
}

func TestSQLiteDB_QueryNotConnected(t *testing.T) {
	db := NewSQLiteDB(&Config{Driver: "sqlite"})
	_, err := db.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestSQLiteDB_ExecNotConnected(t *testing.T) {
	db := NewSQLiteDB(&Config{Driver: "sqlite"})
	_, err := db.Exec(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestSQLiteDB_Transaction(t *testing.T) {
	db := newInMemorySQLite(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE counter (val INTEGER)")
	require.NoError(t, err)
	_, err = db.Exec(ctx, "INSERT INTO counter (val) VALUES (0)")
	require.NoError(t, err)

	// Use Begin directly for testing
	tx, err := db.Begin(ctx)
	require.NoError(t, err)
	_, err = tx.Exec("UPDATE counter SET val = 1")
	require.NoError(t, err)
	err = tx.Commit()
	require.NoError(t, err)

	var val int
	err = db.QueryRow(ctx, "SELECT val FROM counter").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 1, val)
}

func TestSQLiteDB_TransactionRollback(t *testing.T) {
	db := newInMemorySQLite(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE counter (val INTEGER)")
	require.NoError(t, err)
	_, err = db.Exec(ctx, "INSERT INTO counter (val) VALUES (0)")
	require.NoError(t, err)

	// Transaction with rollback
	tx, err := db.Begin(ctx)
	require.NoError(t, err)
	_, err = tx.Exec("UPDATE counter SET val = 99")
	require.NoError(t, err)
	err = tx.Rollback()
	require.NoError(t, err)

	var val int
	err = db.QueryRow(ctx, "SELECT val FROM counter").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 0, val) // Should still be 0
}

func TestSQLiteDB_BeginNotConnected(t *testing.T) {
	db := NewSQLiteDB(&Config{Driver: "sqlite"})
	_, err := db.Begin(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestSQLiteDB_BeginTxNotConnected(t *testing.T) {
	db := NewSQLiteDB(&Config{Driver: "sqlite"})
	_, err := db.BeginTx(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestSQLiteDB_Prepare(t *testing.T) {
	db := newInMemorySQLite(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	stmt, err := db.Prepare(ctx, "INSERT INTO items (name) VALUES (?)")
	require.NoError(t, err)
	defer stmt.Close()

	_, err = stmt.Exec("test1")
	require.NoError(t, err)
	_, err = stmt.Exec("test2")
	require.NoError(t, err)

	var count int
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM items").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestSQLiteDB_PrepareNotConnected(t *testing.T) {
	db := NewSQLiteDB(&Config{Driver: "sqlite"})
	_, err := db.Prepare(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestSQLiteDB_BulkInsert(t *testing.T) {
	db := newInMemorySQLite(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE users (name TEXT, age INTEGER)")
	require.NoError(t, err)

	err = db.BulkInsert(ctx, "users", []string{"name", "age"}, [][]interface{}{
		{"Alice", 30},
		{"Bob", 25},
		{"Charlie", 35},
	})
	require.NoError(t, err)

	var count int
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestSQLiteDB_BulkInsertEmpty(t *testing.T) {
	db := newInMemorySQLite(t)
	err := db.BulkInsert(context.Background(), "users", []string{"name"}, nil)
	require.NoError(t, err)
}

func TestSQLiteDB_DropTable(t *testing.T) {
	db := newInMemorySQLite(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE temp (id INTEGER)")
	require.NoError(t, err)

	exists, err := db.TableExists(ctx, "temp")
	require.NoError(t, err)
	assert.True(t, exists)

	err = db.DropTable(ctx, "temp")
	require.NoError(t, err)

	exists, err = db.TableExists(ctx, "temp")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestSQLiteDB_TableExistsNotFound(t *testing.T) {
	db := newInMemorySQLite(t)
	exists, err := db.TableExists(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestSQLiteDB_GetLastInsertID(t *testing.T) {
	db := newInMemorySQLite(t)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE items (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	require.NoError(t, err)

	_, err = db.Exec(ctx, "INSERT INTO items (name) VALUES (?)", "first")
	require.NoError(t, err)

	id, err := db.GetLastInsertID(ctx, "items", "id")
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

func TestSQLiteDB_PersistenceAcrossReconnect(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "persist.db")

	// First connection: create and populate
	db1 := NewSQLiteDB(&Config{Driver: "sqlite", Database: dbPath})
	err := db1.Connect(context.Background())
	require.NoError(t, err)

	_, err = db1.Exec(context.Background(), "CREATE TABLE data (key TEXT, value TEXT)")
	require.NoError(t, err)
	_, err = db1.Exec(context.Background(), "INSERT INTO data (key, value) VALUES (?, ?)", "hello", "world")
	require.NoError(t, err)
	db1.Close()

	// Second connection: verify data persists
	db2 := NewSQLiteDB(&Config{Driver: "sqlite", Database: dbPath})
	err = db2.Connect(context.Background())
	require.NoError(t, err)
	defer db2.Close()

	var value string
	err = db2.QueryRow(context.Background(), "SELECT value FROM data WHERE key = ?", "hello").Scan(&value)
	require.NoError(t, err)
	assert.Equal(t, "world", value)
}

func TestNewDatabaseFromString_SQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDatabaseFromString("sqlite://" + dbPath)
	require.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, "sqlite", db.Driver())
}

func TestSQLiteDB_FileCleaned(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "cleanup.db")

	db := NewSQLiteDB(&Config{Driver: "sqlite", Database: dbPath})
	err := db.Connect(context.Background())
	require.NoError(t, err)

	_, err = db.Exec(context.Background(), "CREATE TABLE test (id INTEGER)")
	require.NoError(t, err)

	db.Close()
	// File should still exist after close
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)
}

// --- Sanitization tests ---

func TestSanitizeSQLiteIdentifier(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"users", `"users"`, false},
		{"my_table", `"my_table"`, false},
		{"_private", `"_private"`, false},
		{"", "", true},
		{"table name", "", true},
		{"drop;--", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := SanitizeSQLiteIdentifier(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSanitizeSQLiteIdentifiers(t *testing.T) {
	result, err := SanitizeSQLiteIdentifiers([]string{"name", "age"})
	require.NoError(t, err)
	assert.Equal(t, []string{`"name"`, `"age"`}, result)

	_, err = SanitizeSQLiteIdentifiers([]string{"valid", "drop;--"})
	assert.Error(t, err)
}

// --- Config SQLite support ---

func TestConfig_ConnectionString_SQLite(t *testing.T) {
	config := &Config{Driver: "sqlite", Database: "/data/app.db"}
	assert.Equal(t, "/data/app.db", config.ConnectionString())

	config2 := &Config{Driver: "sqlite", Database: ""}
	assert.Equal(t, ":memory:", config2.ConnectionString())
}

func TestConfig_SafeConnectionString_SQLite(t *testing.T) {
	config := &Config{Driver: "sqlite", Database: "/data/app.db"}
	assert.Equal(t, "sqlite:///data/app.db", config.SafeConnectionString())

	config2 := &Config{Driver: "sqlite", Database: ""}
	assert.Equal(t, "sqlite://:memory:", config2.SafeConnectionString())
}

func TestConfig_String_SQLite(t *testing.T) {
	config := &Config{Driver: "sqlite", Database: "/data/app.db"}
	assert.Equal(t, "sqlite:///data/app.db", config.String())
}
