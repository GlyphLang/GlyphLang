# Glyph Database Package

Complete database integration layer for Glyph with PostgreSQL driver and ORM capabilities.

## Features

- **PostgreSQL Driver**: Production-ready driver with connection pooling
- **Generic Database Interface**: Support for multiple database drivers
- **ORM Layer**: Type-safe query builder and CRUD operations
- **Connection Management**: Automatic connection pooling and health checks
- **Transaction Support**: Full transaction support with rollback
- **Prepared Statements**: Security through prepared statements
- **Mock Database**: Built-in mock for testing without actual database

## Installation

Add to your go.mod:

```bash
go get github.com/lib/pq
```

## Quick Start

### Basic Connection

```go
import "github.com/glyphlang/glyph/pkg/database"

// From connection string
db, err := database.NewDatabaseFromString("postgres://user:pass@localhost:5432/mydb")
if err != nil {
    log.Fatal(err)
}

// Connect
ctx := context.Background()
if err := db.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### Using the ORM

```go
// Create ORM instance for a table
orm := database.NewORM(db, "users")

// Create a record
user := map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30,
}
created, err := orm.Create(ctx, user)

// Find by ID
user, err := orm.FindByID(ctx, 1)

// Update
updates := map[string]interface{}{
    "name": "Jane Doe",
}
updated, err := orm.Update(ctx, 1, updates)

// Delete
err := orm.Delete(ctx, 1)

// Query with builder
users, err := orm.NewQueryBuilder().
    WhereEq("active", true).
    Where("age", ">", 18).
    OrderBy("created_at", "DESC").
    Limit(10).
    Get(ctx)
```

### Using the Handler (for Glyph integration)

```go
// Create handler
handler, err := database.NewHandlerFromString("postgres://...")

// Access tables
users := handler.Table("users")

// Perform operations
all, err := users.All()
user, err := users.Get(1)
created, err := users.Create(data)
updated, err := users.Update(1, data)
err := users.Delete(1)
count, err := users.Count("status", "active")
```

## Glyph Language Integration

In your Glyph code, inject the database using `% db: Database`:

```glyph
@ route /api/users/:id [GET] -> User
  % db: Database

  $ user = db.users.get(id)

  if user == null {
    > { error: "User not found" }
  } else {
    > user
  }
```

### Available Database Operations

- `db.table.all()` - Get all records
- `db.table.get(id)` - Get record by ID
- `db.table.create(data)` - Create new record
- `db.table.update(id, data)` - Update record
- `db.table.delete(id)` - Delete record
- `db.table.count(column, value)` - Count matching records
- `db.table.filter(column, value)` - Filter records
- `db.table.length()` - Total record count
- `db.table.nextId()` - Get next available ID

## Query Builder

The query builder provides a fluent interface for building complex queries:

```go
qb := orm.NewQueryBuilder()

// Select specific columns
qb.Select("id", "name", "email")

// WHERE conditions
qb.WhereEq("status", "active")
qb.Where("age", ">", 18)
qb.Where("name", "LIKE", "%john%")

// Joins
qb.InnerJoin("posts", "user_id", "id")
qb.LeftJoin("profiles", "user_id", "id")

// Ordering
qb.OrderBy("created_at", "DESC")

// Pagination
qb.Limit(10).Offset(20)

// Execute
results, err := qb.Get(ctx)
first, err := qb.First(ctx)
```

## Transaction Support

```go
pgDB := db.(*database.PostgresDB)

err := pgDB.Transaction(ctx, func(tx *sql.Tx) error {
    // Perform multiple operations
    _, err := tx.ExecContext(ctx, "INSERT INTO ...")
    if err != nil {
        return err // Automatic rollback
    }

    _, err = tx.ExecContext(ctx, "UPDATE ...")
    return err // Commit if no error
})
```

## Bulk Operations

```go
pgDB := db.(*database.PostgresDB)

columns := []string{"name", "email", "age"}
values := [][]interface{}{
    {"John", "john@example.com", 30},
    {"Jane", "jane@example.com", 25},
    {"Bob", "bob@example.com", 35},
}

err := pgDB.BulkInsert(ctx, "users", columns, values)
```

## Mock Database for Testing

```go
// Create mock database
mockDB := database.NewMockDatabase()

// Access tables
users := mockDB.Table("users")

// Perform operations (no real database needed)
users.Create(map[string]interface{}{
    "id": 1,
    "name": "Test User",
})

user := users.Get(1)
```

## Configuration

```go
config := &database.Config{
    Driver:          "postgres",
    Host:            "localhost",
    Port:            5432,
    Database:        "mydb",
    Username:        "user",
    Password:        "pass",
    SSLMode:         "disable",
    MaxOpenConns:    25,        // Maximum open connections
    MaxIdleConns:    5,         // Maximum idle connections
    ConnMaxLifetime: 5 * time.Minute,
    ConnMaxIdleTime: 5 * time.Minute,
}

db, err := database.NewDatabase(config)
```

## Health Checks

```go
err := database.HealthCheck(ctx, db)
if err != nil {
    log.Printf("Database unhealthy: %v", err)
}

// Check connection stats
stats := db.Stats()
fmt.Printf("Open connections: %d\n", stats.OpenConnections)
fmt.Printf("Idle connections: %d\n", stats.Idle)
```

## Schema Management

```go
pgDB := db.(*database.PostgresDB)

// Create table
schema := map[string]string{
    "id": "SERIAL PRIMARY KEY",
    "name": "VARCHAR(100) NOT NULL",
    "email": "VARCHAR(100) UNIQUE",
    "created_at": "TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
}
err := pgDB.CreateTable(ctx, "users", schema)

// Check if table exists
exists, err := pgDB.TableExists(ctx, "users")

// Drop table
err := pgDB.DropTable(ctx, "users")
```

## Error Handling

The package provides comprehensive error handling:

- Connection errors
- Query execution errors
- Transaction rollback errors
- Context timeout errors
- Not found errors (sql.ErrNoRows)

## Testing

Run unit tests:
```bash
go test ./pkg/database/...
```

Run integration tests (requires PostgreSQL):
```bash
DATABASE_URL="postgres://user:pass@localhost:5432/testdb" go test ./pkg/database/... -v
```

## Performance

- Connection pooling for optimal performance
- Prepared statements to prevent SQL injection
- Bulk insert support for large datasets
- Query builder generates optimized SQL

## Security

- Prepared statements prevent SQL injection
- Parameterized queries
- SSL/TLS support
- Connection string password protection

## Examples

See `examples/database-demo/main.abc` for a complete CRUD application example.

## API Reference

### Database Interface

- `Connect(ctx) error` - Establish connection
- `Close() error` - Close connection
- `Ping(ctx) error` - Check connection health
- `Query(ctx, query, args...) (*Rows, error)` - Execute query
- `QueryRow(ctx, query, args...) *Row` - Execute single row query
- `Exec(ctx, query, args...) (Result, error)` - Execute command
- `Begin(ctx) (*Tx, error)` - Start transaction
- `Prepare(ctx, query) (*Stmt, error)` - Prepare statement
- `Stats() DBStats` - Get connection statistics

### ORM Methods

- `FindByID(ctx, id) (map, error)` - Find record by ID
- `FindAll(ctx) ([]map, error)` - Find all records
- `Create(ctx, data) (map, error)` - Create record
- `Update(ctx, id, data) (map, error)` - Update record
- `Delete(ctx, id) error` - Delete record
- `Count(ctx, conditions...) (int64, error)` - Count records
- `Exists(ctx, conditions...) (bool, error)` - Check existence
- `Query(ctx, query, args...) ([]map, error)` - Raw query

### QueryBuilder Methods

- `Select(columns...)` - Select columns
- `Where(col, op, val)` - Add WHERE condition
- `WhereEq(col, val)` - Add equality condition
- `OrderBy(col, dir)` - Add ORDER BY
- `Limit(n)` - Add LIMIT
- `Offset(n)` - Add OFFSET
- `InnerJoin(table, on, with)` - Add INNER JOIN
- `LeftJoin(table, on, with)` - Add LEFT JOIN
- `Build()` - Build SQL query
- `Get(ctx)` - Execute and get results
- `First(ctx)` - Get first result

## License

Part of the Glyph project.
