package database

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// mysqlIdentifierPattern matches valid MySQL identifiers
var mysqlIdentifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// SanitizeMySQLIdentifier validates and quotes a MySQL identifier with backticks
func SanitizeMySQLIdentifier(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("identifier cannot be empty")
	}
	if !mysqlIdentifierPattern.MatchString(name) {
		return "", ErrInvalidIdentifier
	}
	return fmt.Sprintf("`%s`", name), nil
}

// SanitizeMySQLIdentifiers validates and quotes multiple MySQL identifiers
func SanitizeMySQLIdentifiers(names []string) ([]string, error) {
	result := make([]string, len(names))
	for i, name := range names {
		sanitized, err := SanitizeMySQLIdentifier(name)
		if err != nil {
			return nil, fmt.Errorf("invalid identifier %q: %w", name, err)
		}
		result[i] = sanitized
	}
	return result, nil
}

// MySQLDB implements the Database interface for MySQL
type MySQLDB struct {
	config *Config
	db     *sql.DB
}

// NewMySQLDB creates a new MySQL database instance
func NewMySQLDB(config *Config) *MySQLDB {
	return &MySQLDB{
		config: config,
	}
}

// Connect establishes a connection to the MySQL database
func (m *MySQLDB) Connect(ctx context.Context) error {
	connStr := m.config.ConnectionString()

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(m.config.MaxOpenConns)
	db.SetMaxIdleConns(m.config.MaxIdleConns)
	db.SetConnMaxLifetime(m.config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(m.config.ConnMaxIdleTime)

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	m.db = db
	return nil
}

// Close closes the database connection
func (m *MySQLDB) Close() error {
	if m.db == nil {
		return nil
	}
	return m.db.Close()
}

// Ping verifies the database connection is alive
func (m *MySQLDB) Ping(ctx context.Context) error {
	if m.db == nil {
		return fmt.Errorf("database not connected")
	}
	return m.db.PingContext(ctx)
}

// Query executes a query that returns rows
func (m *MySQLDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return m.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (m *MySQLDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.db == nil {
		return &sql.Row{}
	}
	return m.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (m *MySQLDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return m.db.ExecContext(ctx, query, args...)
}

// Begin starts a new transaction
func (m *MySQLDB) Begin(ctx context.Context) (*sql.Tx, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return m.db.BeginTx(ctx, nil)
}

// BeginTx starts a new transaction with options
func (m *MySQLDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return m.db.BeginTx(ctx, opts)
}

// Prepare creates a prepared statement
func (m *MySQLDB) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return m.db.PrepareContext(ctx, query)
}

// Stats returns database statistics
func (m *MySQLDB) Stats() sql.DBStats {
	if m.db == nil {
		return sql.DBStats{}
	}
	return m.db.Stats()
}

// Driver returns the driver name
func (m *MySQLDB) Driver() string {
	return "mysql"
}

// Transaction executes a function within a transaction
func (m *MySQLDB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := m.Begin(ctx)
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

// BulkInsert performs a bulk insert operation using MySQL syntax
func (m *MySQLDB) BulkInsert(ctx context.Context, table string, columns []string, values [][]interface{}) error {
	if len(values) == 0 {
		return nil
	}

	sanitizedTable, err := SanitizeMySQLIdentifier(table)
	if err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	sanitizedColumns, err := SanitizeMySQLIdentifiers(columns)
	if err != nil {
		return fmt.Errorf("invalid column name: %w", err)
	}

	// Build the bulk insert query using strings.Builder for O(n) performance
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

	_, err = m.Exec(ctx, qb.String(), args...)
	return err
}

// validMySQLColumnTypes contains the allowed MySQL column types
var validMySQLColumnTypes = map[string]bool{
	// Numeric types
	"TINYINT": true, "SMALLINT": true, "MEDIUMINT": true, "INT": true,
	"INTEGER": true, "BIGINT": true, "FLOAT": true, "DOUBLE": true,
	"DECIMAL": true, "NUMERIC": true,
	// Character types
	"CHAR": true, "VARCHAR": true, "TEXT": true, "TINYTEXT": true,
	"MEDIUMTEXT": true, "LONGTEXT": true,
	// Binary types
	"BINARY": true, "VARBINARY": true, "BLOB": true, "TINYBLOB": true,
	"MEDIUMBLOB": true, "LONGBLOB": true,
	// Date/Time types
	"DATE": true, "TIME": true, "DATETIME": true, "TIMESTAMP": true,
	"YEAR": true,
	// Boolean
	"BOOLEAN": true, "BOOL": true,
	// JSON
	"JSON": true,
	// Enum and Set
	"ENUM": true, "SET": true,
}

// mysqlColumnTypePattern matches valid column type definitions
var mysqlColumnTypePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_ (),.]*$`)

// sanitizeMySQLColumnType validates a MySQL column type definition
func sanitizeMySQLColumnType(colType string) (string, error) {
	if colType == "" {
		return "", fmt.Errorf("column type cannot be empty")
	}

	upperType := strings.ToUpper(strings.TrimSpace(colType))

	if !mysqlColumnTypePattern.MatchString(colType) {
		return "", fmt.Errorf("invalid column type: %s", colType)
	}

	baseType := strings.Fields(upperType)[0]
	if idx := strings.Index(baseType, "("); idx != -1 {
		baseType = baseType[:idx]
	}

	if !validMySQLColumnTypes[baseType] {
		return "", fmt.Errorf("unsupported column type: %s", baseType)
	}

	return colType, nil
}

// CreateTable creates a table with the given schema
func (m *MySQLDB) CreateTable(ctx context.Context, table string, schema map[string]string) error {
	sanitizedTable, err := SanitizeMySQLIdentifier(table)
	if err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	var columnDefs []string
	for name, colType := range schema {
		sanitizedName, err := SanitizeMySQLIdentifier(name)
		if err != nil {
			return fmt.Errorf("invalid column name %q: %w", name, err)
		}

		sanitizedType, err := sanitizeMySQLColumnType(colType)
		if err != nil {
			return fmt.Errorf("invalid column type for %q: %w", name, err)
		}

		columnDefs = append(columnDefs, fmt.Sprintf("%s %s", sanitizedName, sanitizedType))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", sanitizedTable, strings.Join(columnDefs, ", "))
	_, err = m.Exec(ctx, query)
	return err
}

// DropTable drops a table
func (m *MySQLDB) DropTable(ctx context.Context, table string) error {
	sanitizedTable, err := SanitizeMySQLIdentifier(table)
	if err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", sanitizedTable)
	_, err = m.Exec(ctx, query)
	return err
}

// TableExists checks if a table exists in MySQL
func (m *MySQLDB) TableExists(ctx context.Context, table string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = DATABASE()
		AND table_name = ?
	`
	var count int
	err := m.QueryRow(ctx, query, table).Scan(&count)
	return count > 0, err
}

// GetLastInsertID retrieves the last inserted ID (MySQL specific)
func (m *MySQLDB) GetLastInsertID(ctx context.Context, table string, idColumn string) (int64, error) {
	if err := ValidateIdentifier(table); err != nil {
		return 0, fmt.Errorf("invalid table name: %w", err)
	}
	if err := ValidateIdentifier(idColumn); err != nil {
		return 0, fmt.Errorf("invalid column name: %w", err)
	}

	var id int64
	err := m.QueryRow(ctx, "SELECT LAST_INSERT_ID()").Scan(&id)
	return id, err
}

// WithMySQLTimeout wraps a context with a timeout
func WithMySQLTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
