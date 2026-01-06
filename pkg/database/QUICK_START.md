# Database Integration - Quick Start Guide

## 5-Minute Setup

### 1. Install Dependencies

```bash
go get github.com/lib/pq
```

### 2. Create Database Connection

```go
import "github.com/glyphlang/glyph/pkg/database"

// Option A: From connection string
db, err := database.NewDatabaseFromString(
    "postgres://user:pass@localhost:5432/mydb?sslmode=disable")

// Option B: From config
config := &database.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "mydb",
    Username: "user",
    Password: "pass",
    SSLMode:  "disable",
}
db, err := database.NewDatabase(config)

// Connect
ctx := context.Background()
if err := db.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### 3. Use in Go Code

```go
// Create handler for Glyph integration
handler := database.NewHandler(db)

// Or use ORM directly
orm := database.NewORM(db, "users")

// Create a record
user := map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com",
}
created, err := orm.Create(ctx, user)

// Find by ID
user, err := orm.FindByID(ctx, 1)

// Update
updates := map[string]interface{}{"name": "Jane Doe"}
updated, err := orm.Update(ctx, 1, updates)

// Delete
err := orm.Delete(ctx, 1)

// Query
users, err := orm.NewQueryBuilder().
    WhereEq("active", true).
    OrderBy("created_at", "DESC").
    Limit(10).
    Get(ctx)
```

### 4. Use in Glyph Code

Create a route with database injection:

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

### 5. Set Up Interpreter

```go
interpreter := interpreter.NewInterpreter()

// Create database handler
handler, err := database.NewHandlerFromString("postgres://...")
if err != nil {
    log.Fatal(err)
}

// Inject into interpreter
interpreter.SetDatabaseHandler(handler)

// Now routes can use % db: Database
```

## Common Patterns

### CRUD Operations

```glyph
# Create
@ route /api/users [POST] -> User
  % db: Database

  $ user = {
    name: input.name,
    email: input.email,
    created_at: time.now()
  }
  $ created = db.users.create(user)
  > created

# Read
@ route /api/users/:id [GET] -> User
  % db: Database

  $ user = db.users.get(id)
  > user

# Update
@ route /api/users/:id [PUT] -> User
  % db: Database

  $ updates = {
    name: input.name,
    updated_at: time.now()
  }
  $ updated = db.users.update(id, updates)
  > updated

# Delete
@ route /api/users/:id [DELETE] -> Response
  % db: Database

  $ result = db.users.delete(id)
  > { success: true, message: "User deleted" }
```

### Filtering and Counting

```glyph
@ route /api/users/active [GET] -> UserList
  % db: Database

  # Get active users
  $ activeUsers = db.users.filter("active", true)

  # Count active users
  $ count = db.users.count("active", true)

  > {
    users: activeUsers,
    total: count
  }
```

### List All Records

```glyph
@ route /api/users [GET] -> UserList
  % db: Database

  $ users = db.users.all()
  $ total = db.users.length()

  > {
    users: users,
    total: total
  }
```

## Testing

### Using Mock Database

```go
import "testing"

func TestUserOperations(t *testing.T) {
    // Create mock database
    mockDB := database.NewMockDatabase()
    users := mockDB.Table("users")

    // Create test data
    user := map[string]interface{}{
        "id": int64(1),
        "name": "Test User",
    }
    created := users.Create(user)

    // Test operations
    found := users.Get(int64(1))
    assert.NotNil(t, found)

    count := users.Count("name", "Test User")
    assert.Equal(t, int64(1), count)
}
```

### Integration Testing

```go
func TestWithRealDB(t *testing.T) {
    // Skip if no database available
    connStr := os.Getenv("DATABASE_URL")
    if connStr == "" {
        t.Skip("DATABASE_URL not set")
    }

    db, err := database.NewDatabaseFromString(connStr)
    require.NoError(t, err)

    ctx := context.Background()
    err = db.Connect(ctx)
    require.NoError(t, err)
    defer db.Close()

    // Your tests here
}
```

## Common Database Operations

### Query Builder

```go
// Simple query
users, err := orm.NewQueryBuilder().
    WhereEq("status", "active").
    Get(ctx)

// Complex query
users, err := orm.NewQueryBuilder().
    Select("id", "name", "email").
    Where("age", ">", 18).
    Where("status", "=", "active").
    OrderBy("created_at", "DESC").
    Limit(10).
    Offset(0).
    Get(ctx)

// With JOIN
users, err := orm.NewQueryBuilder().
    Select("users.*", "posts.title").
    InnerJoin("posts", "user_id", "id").
    WhereEq("posts.published", true).
    Get(ctx)
```

### Transactions

```go
pgDB := db.(*database.PostgresDB)

err := pgDB.Transaction(ctx, func(tx *sql.Tx) error {
    // Multiple operations in transaction
    _, err := tx.ExecContext(ctx,
        "INSERT INTO users (name) VALUES ($1)", "John")
    if err != nil {
        return err // Automatic rollback
    }

    _, err = tx.ExecContext(ctx,
        "UPDATE accounts SET balance = balance - 100 WHERE id = $1", 1)
    return err // Commit if no error
})
```

### Bulk Insert

```go
pgDB := db.(*database.PostgresDB)

columns := []string{"name", "email"}
values := [][]interface{}{
    {"User1", "user1@example.com"},
    {"User2", "user2@example.com"},
    {"User3", "user3@example.com"},
}

err := pgDB.BulkInsert(ctx, "users", columns, values)
```

## Troubleshooting

### Connection Issues

```go
// Check connection
err := db.Ping(ctx)
if err != nil {
    log.Printf("Connection failed: %v", err)
}

// Check stats
stats := db.Stats()
log.Printf("Open: %d, Idle: %d",
    stats.OpenConnections, stats.Idle)
```

### Query Debugging

```go
// Build query without executing
qb := orm.NewQueryBuilder().
    WhereEq("status", "active")

query, args := qb.Build()
log.Printf("SQL: %s, Args: %v", query, args)
```

### Health Check

```go
err := database.HealthCheck(ctx, db)
if err != nil {
    log.Printf("Database unhealthy: %v", err)
}
```

## Best Practices

1. **Always use context**: Pass context for cancellation and timeout
2. **Close connections**: Use `defer db.Close()`
3. **Use prepared statements**: ORM does this automatically
4. **Handle errors**: Check all database errors
5. **Use transactions**: For multiple related operations
6. **Connection pooling**: Configure based on load
7. **Test with mocks**: Use MockDatabase for unit tests
8. **Validate input**: Check data before database operations
9. **Use query builder**: Safer than raw SQL
10. **Monitor stats**: Check connection pool metrics

## Environment Variables

```bash
# For testing
export DATABASE_URL="postgres://user:pass@localhost:5432/testdb"

# For production
export DATABASE_URL="postgres://user:pass@prod-host:5432/proddb?sslmode=require"
```

## Next Steps

1. See `pkg/database/README.md` for full documentation
2. Check `examples/database-demo/main.glyph` for complete example
3. Review tests in `pkg/database/*_test.go` for usage patterns
4. Read integration guide in README for production deployment

## Support

For issues or questions:
- Check the README.md for detailed documentation
- Review test files for usage examples
- See the example application for real-world patterns
