package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"
)

// Database represents a generic database interface
type Database interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	// Query execution
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Transaction support
	Begin(ctx context.Context) (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	// Prepared statements
	Prepare(ctx context.Context, query string) (*sql.Stmt, error)

	// Connection info
	Stats() sql.DBStats
	Driver() string
}

// Config represents database configuration
type Config struct {
	Driver          string
	Host            string
	Port            int
	Database        string
	Username        string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// ParseConnectionString parses a database connection string
func ParseConnectionString(connStr string) (*Config, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	// Validate that the connection string has a scheme (driver)
	if u.Scheme == "" {
		return nil, fmt.Errorf("invalid connection string: missing database driver scheme")
	}

	config := &Config{
		Driver:          u.Scheme,
		Host:            u.Hostname(),
		Database:        u.Path[1:], // Remove leading slash
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	// Parse port
	if u.Port() != "" {
		var port int
		if _, err := fmt.Sscanf(u.Port(), "%d", &port); err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
		config.Port = port
	} else {
		// Set default port based on driver
		switch config.Driver {
		case "postgres", "postgresql":
			config.Port = 5432
		case "mysql":
			config.Port = 3306
		default:
			config.Port = 5432
		}
	}

	// Parse username and password
	if u.User != nil {
		config.Username = u.User.Username()
		if password, ok := u.User.Password(); ok {
			config.Password = password
		}
	}

	// Parse query parameters
	query := u.Query()
	if sslMode := query.Get("sslmode"); sslMode != "" {
		config.SSLMode = sslMode
	}

	return config, nil
}

// ConnectionString generates a connection string from config
func (c *Config) ConnectionString() string {
	switch c.Driver {
	case "postgres", "postgresql":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	default:
		return ""
	}
}

// NewDatabase creates a new database instance based on the driver
func NewDatabase(config *Config) (Database, error) {
	switch config.Driver {
	case "postgres", "postgresql":
		return NewPostgresDB(config), nil
	case "mysql":
		return nil, fmt.Errorf("MySQL driver not yet implemented")
	case "sqlite", "sqlite3":
		return nil, fmt.Errorf("SQLite driver not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}
}

// NewDatabaseFromString creates a new database from a connection string
func NewDatabaseFromString(connStr string) (Database, error) {
	config, err := ParseConnectionString(connStr)
	if err != nil {
		return nil, err
	}
	return NewDatabase(config)
}

// HealthCheck performs a health check on the database
func HealthCheck(ctx context.Context, db Database) error {
	// Set a timeout for the health check
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check connection stats
	stats := db.Stats()
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no open database connections")
	}

	return nil
}
