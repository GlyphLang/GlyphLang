package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

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

	// Build the bulk insert query
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", table, columnsString(columns))

	var args []interface{}
	placeholderIndex := 1

	for i, row := range values {
		if i > 0 {
			query += ", "
		}
		query += "("
		for j := range row {
			if j > 0 {
				query += ", "
			}
			query += fmt.Sprintf("$%d", placeholderIndex)
			placeholderIndex++
			args = append(args, row[j])
		}
		query += ")"
	}

	_, err := p.Exec(ctx, query, args...)
	return err
}

// CreateTable creates a table with the given schema
func (p *PostgresDB) CreateTable(ctx context.Context, table string, schema map[string]string) error {
	columns := ""
	first := true
	for name, colType := range schema {
		if !first {
			columns += ", "
		}
		columns += fmt.Sprintf("%s %s", name, colType)
		first = false
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", table, columns)
	_, err := p.Exec(ctx, query)
	return err
}

// DropTable drops a table
func (p *PostgresDB) DropTable(ctx context.Context, table string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
	_, err := p.Exec(ctx, query)
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
	query := fmt.Sprintf("SELECT CURRVAL(pg_get_serial_sequence('%s', '%s'))", table, idColumn)
	var id int64
	err := p.QueryRow(ctx, query).Scan(&id)
	return id, err
}

// Helper function to create a columns string
func columnsString(columns []string) string {
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += col
	}
	return result
}

// WithTimeout wraps a context with a timeout
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
