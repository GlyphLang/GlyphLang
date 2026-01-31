package database

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// identifierPattern matches valid SQL identifiers (alphanumeric and underscore only)
var identifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// ErrInvalidIdentifier is returned when an identifier contains invalid characters
var ErrInvalidIdentifier = fmt.Errorf("invalid identifier: must contain only alphanumeric characters and underscores, and start with a letter or underscore")

// SanitizeIdentifier validates and quotes a SQL identifier (table/column name)
// It returns an error if the identifier contains invalid characters
func SanitizeIdentifier(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("identifier cannot be empty")
	}
	if !identifierPattern.MatchString(name) {
		return "", ErrInvalidIdentifier
	}
	// Return double-quoted identifier for PostgreSQL
	return fmt.Sprintf(`"%s"`, name), nil
}

// SanitizeIdentifiers validates and quotes multiple SQL identifiers
func SanitizeIdentifiers(names []string) ([]string, error) {
	result := make([]string, len(names))
	for i, name := range names {
		sanitized, err := SanitizeIdentifier(name)
		if err != nil {
			return nil, fmt.Errorf("invalid identifier %q: %w", name, err)
		}
		result[i] = sanitized
	}
	return result, nil
}

// ValidateIdentifier validates an identifier without quoting it
// Useful when you need to validate but the quoting is done elsewhere
func ValidateIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if !identifierPattern.MatchString(name) {
		return ErrInvalidIdentifier
	}
	return nil
}

// PostgresDB implements the Database interface for PostgreSQL
type PostgresDB struct {
	config *Config
	db     *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database instance
func NewPostgresDB(config *Config) *PostgresDB {
	return &PostgresDB{
		config: config,
	}
}

// Connect establishes a connection to the PostgreSQL database
func (p *PostgresDB) Connect(ctx context.Context) error {
	connStr := p.config.ConnectionString()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(p.config.MaxOpenConns)
	db.SetMaxIdleConns(p.config.MaxIdleConns)
	db.SetConnMaxLifetime(p.config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(p.config.ConnMaxIdleTime)

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.db = db
	return nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	if p.db == nil {
		return nil
	}
	return p.db.Close()
}

// Ping verifies the database connection is alive
func (p *PostgresDB) Ping(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("database not connected")
	}
	return p.db.PingContext(ctx)
}

// Query executes a query that returns rows
func (p *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return p.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (p *PostgresDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if p.db == nil {
		// Return a row that will error when scanned
		return &sql.Row{}
	}
	return p.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (p *PostgresDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return p.db.ExecContext(ctx, query, args...)
}

// Begin starts a new transaction
func (p *PostgresDB) Begin(ctx context.Context) (*sql.Tx, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return p.db.BeginTx(ctx, nil)
}

// BeginTx starts a new transaction with options
func (p *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return p.db.BeginTx(ctx, opts)
}

// Prepare creates a prepared statement
func (p *PostgresDB) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return p.db.PrepareContext(ctx, query)
}

// Stats returns database statistics
func (p *PostgresDB) Stats() sql.DBStats {
	if p.db == nil {
		return sql.DBStats{}
	}
	return p.db.Stats()
}

// Driver returns the driver name
func (p *PostgresDB) Driver() string {
	return "postgres"
}

// Transaction executes a function within a transaction
func (p *PostgresDB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := p.Begin(ctx)
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

// BulkInsert performs a bulk insert operation
func (p *PostgresDB) BulkInsert(ctx context.Context, table string, columns []string, values [][]interface{}) error {
	if len(values) == 0 {
		return nil
	}

	// Validate and sanitize table name
	sanitizedTable, err := SanitizeIdentifier(table)
	if err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	// Validate and sanitize column names
	sanitizedColumns, err := SanitizeIdentifiers(columns)
	if err != nil {
		return fmt.Errorf("invalid column name: %w", err)
	}

	// Build the bulk insert query using strings.Builder for O(n) performance
	var qb strings.Builder
	fmt.Fprintf(&qb, "INSERT INTO %s (%s) VALUES ", sanitizedTable, strings.Join(sanitizedColumns, ", "))

	var args []interface{}
	placeholderIndex := 1

	for i, row := range values {
		if i > 0 {
			qb.WriteString(", ")
		}
		qb.WriteByte('(')
		for j := range row {
			if j > 0 {
				qb.WriteString(", ")
			}
			fmt.Fprintf(&qb, "$%d", placeholderIndex)
			placeholderIndex++
			args = append(args, row[j])
		}
		qb.WriteByte(')')
	}

	_, err = p.Exec(ctx, qb.String(), args...)
	return err
}

// validColumnTypes contains the allowed PostgreSQL column types
var validColumnTypes = map[string]bool{
	// Numeric types
	"SMALLINT": true, "INTEGER": true, "BIGINT": true, "INT": true,
	"DECIMAL": true, "NUMERIC": true, "REAL": true, "DOUBLE PRECISION": true,
	"SMALLSERIAL": true, "SERIAL": true, "BIGSERIAL": true,
	// Character types
	"CHAR": true, "VARCHAR": true, "TEXT": true,
	// Binary types
	"BYTEA": true,
	// Date/Time types
	"DATE": true, "TIME": true, "TIMESTAMP": true, "TIMESTAMPTZ": true,
	"INTERVAL": true,
	// Boolean
	"BOOLEAN": true, "BOOL": true,
	// UUID
	"UUID": true,
	// JSON
	"JSON": true, "JSONB": true,
	// Other common types
	"MONEY": true, "INET": true, "CIDR": true, "MACADDR": true,
}

// columnTypePattern matches valid column type definitions
var columnTypePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_ (),.]*$`)

// sanitizeColumnType validates a column type definition
func sanitizeColumnType(colType string) (string, error) {
	if colType == "" {
		return "", fmt.Errorf("column type cannot be empty")
	}

	// Normalize to uppercase for checking
	upperType := strings.ToUpper(strings.TrimSpace(colType))

	// Check against pattern to prevent injection
	if !columnTypePattern.MatchString(colType) {
		return "", fmt.Errorf("invalid column type: %s", colType)
	}

	// Extract base type (before any parentheses or modifiers)
	baseType := strings.Fields(upperType)[0]
	// Remove parentheses if present (e.g., VARCHAR(255) -> VARCHAR)
	if idx := strings.Index(baseType, "("); idx != -1 {
		baseType = baseType[:idx]
	}

	// Check if the base type is allowed
	if !validColumnTypes[baseType] {
		return "", fmt.Errorf("unsupported column type: %s", baseType)
	}

	return colType, nil
}

// CreateTable creates a table with the given schema
func (p *PostgresDB) CreateTable(ctx context.Context, table string, schema map[string]string) error {
	// Validate and sanitize table name
	sanitizedTable, err := SanitizeIdentifier(table)
	if err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	var columnDefs []string
	for name, colType := range schema {
		// Validate and sanitize column name
		sanitizedName, err := SanitizeIdentifier(name)
		if err != nil {
			return fmt.Errorf("invalid column name %q: %w", name, err)
		}

		// Validate column type
		sanitizedType, err := sanitizeColumnType(colType)
		if err != nil {
			return fmt.Errorf("invalid column type for %q: %w", name, err)
		}

		columnDefs = append(columnDefs, fmt.Sprintf("%s %s", sanitizedName, sanitizedType))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", sanitizedTable, strings.Join(columnDefs, ", "))
	_, err = p.Exec(ctx, query)
	return err
}

// DropTable drops a table
func (p *PostgresDB) DropTable(ctx context.Context, table string) error {
	// Validate and sanitize table name
	sanitizedTable, err := SanitizeIdentifier(table)
	if err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", sanitizedTable)
	_, err = p.Exec(ctx, query)
	return err
}

// TableExists checks if a table exists
func (p *PostgresDB) TableExists(ctx context.Context, table string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = $1
		)
	`
	var exists bool
	err := p.QueryRow(ctx, query, table).Scan(&exists)
	return exists, err
}

// GetLastInsertID retrieves the last inserted ID (PostgreSQL specific)
func (p *PostgresDB) GetLastInsertID(ctx context.Context, table string, idColumn string) (int64, error) {
	// Validate identifiers to prevent SQL injection.
	// SanitizeIdentifier confirms these are alphanumeric-only identifiers.
	// pg_get_serial_sequence takes unquoted text arguments, so we use
	// the validated original names rather than the double-quoted sanitized forms.
	if _, err := SanitizeIdentifier(table); err != nil {
		return 0, fmt.Errorf("invalid table name: %w", err)
	}
	if _, err := SanitizeIdentifier(idColumn); err != nil {
		return 0, fmt.Errorf("invalid column name: %w", err)
	}

	// Identifiers have been validated above by SanitizeIdentifier.
	// pg_get_serial_sequence takes text arguments (not SQL identifiers),
	// so we use the validated original names in a parameterized query.
	query := "SELECT CURRVAL(pg_get_serial_sequence($1, $2))"

	var id int64
	err := p.QueryRow(ctx, query, table, idColumn).Scan(&id)
	return id, err
}

// columnsString is unused dead code - use strings.Join(columns, ", ") instead.

// WithTimeout wraps a context with a timeout
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
