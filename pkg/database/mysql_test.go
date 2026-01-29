package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMySQLDB(t *testing.T) {
	config := &Config{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "user",
		Password: "pass",
	}

	db := NewMySQLDB(config)
	assert.NotNil(t, db)
	assert.Equal(t, "mysql", db.Driver())
	assert.Nil(t, db.db)
}

func TestMySQLDB_DriverName(t *testing.T) {
	db := NewMySQLDB(&Config{Driver: "mysql"})
	assert.Equal(t, "mysql", db.Driver())
}

func TestMySQLDB_CloseNilDB(t *testing.T) {
	db := NewMySQLDB(&Config{Driver: "mysql"})
	err := db.Close()
	assert.NoError(t, err)
}

func TestMySQLDB_PingNotConnected(t *testing.T) {
	ctx := context.Background()
	db := NewMySQLDB(&Config{Driver: "mysql"})
	err := db.Ping(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestMySQLDB_QueryNotConnected(t *testing.T) {
	ctx := context.Background()
	db := NewMySQLDB(&Config{Driver: "mysql"})
	_, err := db.Query(ctx, "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestMySQLDB_ExecNotConnected(t *testing.T) {
	ctx := context.Background()
	db := NewMySQLDB(&Config{Driver: "mysql"})
	_, err := db.Exec(ctx, "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestMySQLDB_BeginNotConnected(t *testing.T) {
	ctx := context.Background()
	db := NewMySQLDB(&Config{Driver: "mysql"})
	_, err := db.Begin(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestMySQLDB_BeginTxNotConnected(t *testing.T) {
	ctx := context.Background()
	db := NewMySQLDB(&Config{Driver: "mysql"})
	_, err := db.BeginTx(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestMySQLDB_PrepareNotConnected(t *testing.T) {
	ctx := context.Background()
	db := NewMySQLDB(&Config{Driver: "mysql"})
	_, err := db.Prepare(ctx, "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestMySQLDB_StatsNotConnected(t *testing.T) {
	db := NewMySQLDB(&Config{Driver: "mysql"})
	stats := db.Stats()
	assert.Equal(t, 0, stats.OpenConnections)
}

func TestMySQLDB_QueryRowNotConnected(t *testing.T) {
	ctx := context.Background()
	db := NewMySQLDB(&Config{Driver: "mysql"})
	row := db.QueryRow(ctx, "SELECT 1")
	assert.NotNil(t, row)
}

// SanitizeMySQLIdentifier tests

func TestSanitizeMySQLIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid identifier",
			input: "users",
			want:  "`users`",
		},
		{
			name:  "identifier with underscore",
			input: "user_name",
			want:  "`user_name`",
		},
		{
			name:  "identifier starting with underscore",
			input: "_id",
			want:  "`_id`",
		},
		{
			name:    "empty identifier",
			input:   "",
			wantErr: true,
		},
		{
			name:    "identifier with spaces",
			input:   "user name",
			wantErr: true,
		},
		{
			name:    "identifier with SQL injection",
			input:   "users; DROP TABLE users",
			wantErr: true,
		},
		{
			name:    "identifier starting with number",
			input:   "123users",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeMySQLIdentifier(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeMySQLIdentifiers(t *testing.T) {
	names := []string{"id", "name", "email"}
	result, err := SanitizeMySQLIdentifiers(names)
	require.NoError(t, err)
	assert.Equal(t, []string{"`id`", "`name`", "`email`"}, result)

	// Test with invalid identifier
	_, err = SanitizeMySQLIdentifiers([]string{"valid", "invalid name"})
	assert.Error(t, err)
}

// sanitizeMySQLColumnType tests (same-package access to unexported function)

func TestSanitizeMySQLColumnType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "INT", input: "INT"},
		{name: "VARCHAR with length", input: "VARCHAR(255)"},
		{name: "TEXT", input: "TEXT"},
		{name: "DATETIME", input: "DATETIME"},
		{name: "BOOLEAN", input: "BOOLEAN"},
		{name: "JSON", input: "JSON"},
		{name: "BIGINT", input: "BIGINT"},
		{name: "DECIMAL with precision", input: "DECIMAL(10, 2)"},
		{name: "empty", input: "", wantErr: true},
		{name: "unsupported type", input: "FAKETYPE", wantErr: true},
		{name: "injection attempt", input: "INT; DROP TABLE", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizeMySQLColumnType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.input, result)
		})
	}
}

// Factory and connection string tests

func TestNewDatabase_MySQLDriver(t *testing.T) {
	config := &Config{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "user",
		Password: "pass",
	}

	db, err := NewDatabase(config)
	require.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, "mysql", db.Driver())
}

func TestMySQLConnectionString(t *testing.T) {
	config := &Config{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "secret",
	}

	connStr := config.ConnectionString()
	assert.Equal(t, "root:secret@tcp(localhost:3306)/testdb", connStr)
}

func TestParseMySQLConnectionString(t *testing.T) {
	cfg, err := ParseConnectionString("mysql://user:pass@localhost:3306/mydb")
	require.NoError(t, err)
	assert.Equal(t, "mysql", cfg.Driver)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 3306, cfg.Port)
	assert.Equal(t, "mydb", cfg.Database)
	assert.Equal(t, "user", cfg.Username)
	assert.Equal(t, "pass", cfg.Password)
}

func TestParseMySQLConnectionString_DefaultPort(t *testing.T) {
	cfg, err := ParseConnectionString("mysql://user:pass@localhost/mydb")
	require.NoError(t, err)
	assert.Equal(t, 3306, cfg.Port)
}

// Validation and edge case tests

func TestMySQLDB_BulkInsertEmpty(t *testing.T) {
	db := NewMySQLDB(&Config{Driver: "mysql"})
	err := db.BulkInsert(context.Background(), "users", []string{"id", "name"}, [][]interface{}{})
	assert.NoError(t, err)
}

func TestMySQLDB_BulkInsertInvalidTable(t *testing.T) {
	db := NewMySQLDB(&Config{Driver: "mysql"})
	err := db.BulkInsert(context.Background(), "invalid table", []string{"id"}, [][]interface{}{{1}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
}

func TestMySQLDB_BulkInsertInvalidColumn(t *testing.T) {
	db := NewMySQLDB(&Config{Driver: "mysql"})
	err := db.BulkInsert(context.Background(), "users", []string{"invalid col"}, [][]interface{}{{1}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid column name")
}

func TestMySQLDB_CreateTableInvalidTable(t *testing.T) {
	db := NewMySQLDB(&Config{Driver: "mysql"})
	err := db.CreateTable(context.Background(), "invalid table", map[string]string{"id": "INT"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
}

func TestMySQLDB_CreateTableInvalidColumn(t *testing.T) {
	db := NewMySQLDB(&Config{Driver: "mysql"})
	err := db.CreateTable(context.Background(), "users", map[string]string{"invalid col": "INT"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid column name")
}

func TestMySQLDB_CreateTableInvalidType(t *testing.T) {
	db := NewMySQLDB(&Config{Driver: "mysql"})
	err := db.CreateTable(context.Background(), "users", map[string]string{"id": "FAKETYPE"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid column type")
}

func TestMySQLDB_DropTableInvalidTable(t *testing.T) {
	db := NewMySQLDB(&Config{Driver: "mysql"})
	err := db.DropTable(context.Background(), "invalid table")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
}

func TestMySQLDB_GetLastInsertIDInvalidTable(t *testing.T) {
	db := NewMySQLDB(&Config{Driver: "mysql"})
	_, err := db.GetLastInsertID(context.Background(), "invalid table", "id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
}

func TestMySQLDB_GetLastInsertIDInvalidColumn(t *testing.T) {
	db := NewMySQLDB(&Config{Driver: "mysql"})
	_, err := db.GetLastInsertID(context.Background(), "users", "invalid col")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid column name")
}
