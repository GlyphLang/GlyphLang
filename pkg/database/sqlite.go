package database

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// sqliteIdentifierPattern matches valid SQLite identifiers
var sqliteIdentifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// SanitizeSQLiteIdentifier validates and quotes a SQLite identifier with double quotes
func SanitizeSQLiteIdentifier(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("identifier cannot be empty")
	}
	if !sqliteIdentifierPattern.MatchString(name) {
		return "", ErrInvalidIdentifier
	}
	return fmt.Sprintf(`"%s"`, name), nil
}

// SanitizeSQLiteIdentifiers validates and quotes multiple SQLite identifiers
func SanitizeSQLiteIdentifiers(names []string) ([]string, error) {
	result := make([]string, len(names))
	for i, name := range names {
		sanitized, err := SanitizeSQLiteIdentifier(name)
		if err != nil {
			return nil, fmt.Errorf("invalid identifier %q: %w", name, err)
		}
		result[i] = sanitized
	}
	return result, nil
}

// SQLiteDB implements the Database interface for SQLite
type SQLiteDB struct {
	config *Config
	db     *sql.DB
}

// NewSQLiteDB creates a new SQLite database instance
func NewSQLiteDB(config *Config) *SQLiteDB {
	return &SQLiteDB{
		config: config,
	}
}

// Connect establishes a connection to the SQLite database
func (s *SQLiteDB) Connect(ctx context.Context) error {
	dsn := s.config.Database
	if dsn == "" {
		dsn = ":memory:"
	}

	// Enable WAL mode and foreign keys via query parameters
	if !strings.Contains(dsn, "?") && dsn != ":memory:" {
		dsn += "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
	} else if dsn == ":memory:" {
		dsn += "?_pragma=foreign_keys(1)"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// SQLite does not benefit from multiple connections in the same way as
	// client-server databases. Limit to 1 open connection to avoid
	// "database is locked" errors during concurrent writes.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // No expiration

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	s.db = db
	return nil
}

// Close closes the database connection
func (s *SQLiteDB) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Ping verifies the database connection is alive
func (s *SQLiteDB) Ping(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("database not connected")
	}
	return s.db.PingContext(ctx)
}

// Query executes a query that returns rows
func (s *SQLiteDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return s.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (s *SQLiteDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if s.db == nil {
		return &sql.Row{}
	}
	return s.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (s *SQLiteDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return s.db.ExecContext(ctx, query, args...)
}

// Begin starts a new transaction
func (s *SQLiteDB) Begin(ctx context.Context) (*sql.Tx, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return s.db.BeginTx(ctx, nil)
}

// BeginTx starts a new transaction with options
func (s *SQLiteDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return s.db.BeginTx(ctx, opts)
}

// Prepare creates a prepared statement
func (s *SQLiteDB) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return s.db.PrepareContext(ctx, query)
}

// Stats returns database statistics
func (s *SQLiteDB) Stats() sql.DBStats {
	if s.db == nil {
		return sql.DBStats{}
	}
	return s.db.Stats()
}

// Driver returns the driver name
func (s *SQLiteDB) Driver() string {
	return "sqlite"
}

// Transaction executes a function within a transaction
func (s *SQLiteDB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// BulkInsert performs a bulk insert operation using SQLite syntax
func (s *SQLiteDB) BulkInsert(ctx context.Context, table string, columns []string, values [][]interface{}) error {
	if len(values) == 0 {
		return nil
	}

	sanitizedTable, err := SanitizeSQLiteIdentifier(table)
	if err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	sanitizedColumns, err := SanitizeSQLiteIdentifiers(columns)
	if err != nil {
		return fmt.Errorf("invalid column name: %w", err)
	}

	// Build the bulk insert query
	var qb strings.Builder
	fmt.Fprintf(&qb, "INSERT INTO %s (%s) VALUES ", sanitizedTable, strings.Join(sanitizedColumns, ", "))

	var args []interface{}
	for i, row := range values {
		if i > 0 {
			qb.WriteString(", ")
		}
		qb.WriteByte('(')
		for j := range row {
			if j > 0 {
				qb.WriteString(", ")
			}
			qb.WriteByte('?')
			args = append(args, row[j])
		}
		qb.WriteByte(')')
	}

	_, err = s.Exec(ctx, qb.String(), args...)
	return err
}

// validSQLiteColumnTypes contains the allowed SQLite column types
var validSQLiteColumnTypes = map[string]bool{
	// SQLite type affinities
	"TEXT": true, "INTEGER": true, "INT": true, "REAL": true,
	"BLOB": true, "NUMERIC": true,
	// Common aliases
	"VARCHAR": true, "CHAR": true, "BOOLEAN": true, "BOOL": true,
	"DATE": true, "DATETIME": true, "TIMESTAMP": true,
	"FLOAT": true, "DOUBLE": true, "DECIMAL": true,
	"BIGINT": true, "SMALLINT": true, "TINYINT": true,
	"DOUBLE PRECISION": true,
}

// sqliteColumnTypePattern matches valid column type definitions
var sqliteColumnTypePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_ (),.]*$`)

// sanitizeSQLiteColumnType validates a SQLite column type definition
func sanitizeSQLiteColumnType(colType string) (string, error) {
	if colType == "" {
		return "", fmt.Errorf("column type cannot be empty")
	}

	upperType := strings.ToUpper(strings.TrimSpace(colType))

	if !sqliteColumnTypePattern.MatchString(colType) {
		return "", fmt.Errorf("invalid column type: %s", colType)
	}

	baseType := strings.Fields(upperType)[0]
	if idx := strings.Index(baseType, "("); idx != -1 {
		baseType = baseType[:idx]
	}

	if !validSQLiteColumnTypes[baseType] {
		return "", fmt.Errorf("unsupported column type: %s", baseType)
	}

	return colType, nil
}

// CreateTable creates a table with the given schema
func (s *SQLiteDB) CreateTable(ctx context.Context, table string, schema map[string]string) error {
	sanitizedTable, err := SanitizeSQLiteIdentifier(table)
	if err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	var columnDefs []string
	for name, colType := range schema {
		sanitizedName, err := SanitizeSQLiteIdentifier(name)
		if err != nil {
			return fmt.Errorf("invalid column name %q: %w", name, err)
		}

		sanitizedType, err := sanitizeSQLiteColumnType(colType)
		if err != nil {
			return fmt.Errorf("invalid column type for %q: %w", name, err)
		}

		columnDefs = append(columnDefs, fmt.Sprintf("%s %s", sanitizedName, sanitizedType))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", sanitizedTable, strings.Join(columnDefs, ", "))
	_, err = s.Exec(ctx, query)
	return err
}

// DropTable drops a table
func (s *SQLiteDB) DropTable(ctx context.Context, table string) error {
	sanitizedTable, err := SanitizeSQLiteIdentifier(table)
	if err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", sanitizedTable)
	_, err = s.Exec(ctx, query)
	return err
}

// TableExists checks if a table exists in SQLite
func (s *SQLiteDB) TableExists(ctx context.Context, table string) (bool, error) {
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`
	var count int
	err := s.QueryRow(ctx, query, table).Scan(&count)
	return count > 0, err
}

// GetLastInsertID retrieves the last inserted row ID
func (s *SQLiteDB) GetLastInsertID(ctx context.Context, table string, idColumn string) (int64, error) {
	if err := ValidateIdentifier(table); err != nil {
		return 0, fmt.Errorf("invalid table name: %w", err)
	}
	if err := ValidateIdentifier(idColumn); err != nil {
		return 0, fmt.Errorf("invalid column name: %w", err)
	}

	var id int64
	err := s.QueryRow(ctx, "SELECT last_insert_rowid()").Scan(&id)
	return id, err
}
