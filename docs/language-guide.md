# GlyphLang Language Guide

## Syntax Overview

GlyphLang uses symbolic tokens instead of English keywords for efficiency and language-neutrality.

### Core Symbols

| Symbol | Meaning | Example |
|--------|---------|---------|
| `:` | Type definition | `: User { ... }` |
| `@` | Route/endpoint | `@ route /api/users` |
| `$` | Database operation | `$ user = db.query(...)` |
| `+` | Middleware/modifier | `+ auth(jwt)` |
| `%` | Service injection | `% db: Database` |
| `>` | Return statement | `> result` |
| `!` | Validation | `! validate { ... }` |
| `<` | Input declaration | `< input: Type` |
| `=` | Function definition | `= myFunction() { ... }` |
| `~` | Async operation | `~ fetch_data()` |
| `#` | Comment | `# This is a comment` |

## Type System

### Basic Types

```glyph
: User {
  id: int!           # Required integer
  name: str!         # Required string
  age: int?          # Optional integer
  balance: float     # Float number
  active: bool       # Boolean
  created: timestamp # Timestamp
}
```

### Type Modifiers

- `!` - Required (non-nullable)
- `?` - Optional (nullable)

### Collection Types

```glyph
: BlogPost {
  tags: List[str]
  authors: Set[User]
  metadata: Map[str, str]
}
```

### Union Types

```glyph
: Result = User | Error
```

## Routes and Endpoints

### Basic Route

```glyph
@ route /api/users/:id -> User | Error
  % db: Database
  $ user = db.users.get(id)
  > user
```

### HTTP Methods

```glyph
@ route /api/users [GET]       # GET request
@ route /api/users [POST]      # POST request
@ route /api/users/:id [PUT]   # PUT request
@ route /api/users/:id [DELETE] # DELETE request
```

### Authentication

```glyph
@ route /api/admin/users
  + auth(jwt)                  # Require JWT
  + auth(jwt, role: admin)     # Require admin role
```

### Rate Limiting

```glyph
@ route /api/search
  + ratelimit(10/sec)          # 10 requests per second
  + ratelimit(100/min)         # 100 requests per minute
  + ratelimit(1000/hour)       # 1000 requests per hour
```

## Database Operations

### Queries

```glyph
# Simple query
$ users = db.users.all()

# Query with parameters
$ user = db.users.get(id)

# Custom SQL (parameterized)
$ results = db.query(
  "SELECT * FROM users WHERE age > :age",
  {age: 18}
)
```

### Mutations

```glyph
# Insert
$ user = db.users.create({
  name: "Alice",
  email: "alice@example.com"
})

# Update
$ user = db.users.update(id, {name: "Bob"})

# Delete
$ db.users.delete(id)
```

## Input Validation

```glyph
@ route /api/users [POST]
  < input: CreateUserInput
  ! validate input {
    name: str(min=1, max=100)
    email: email_format
    age: int(min=0, max=150)
  }
  % db: Database
  $ user = db.users.create(input)
  > user
```

## Functions

```glyph
= calculate_total(items: List[Item]) -> float {
  total = 0
  for item in items {
    total += item.price
  }
  > total
}
```

## Error Handling

```glyph
@ route /api/users/:id
  % db: Database
  try {
    $ user = db.users.get(id)
    > user
  } catch Error {
    > {error: "User not found"}
  }
```

## Async Operations

```glyph
~ fetch_external_data(url: str) -> Data {
  & response = http.get(url)
  data = parse(response)
  > data
}
```

## Next Steps

- See [API Reference](api-reference.md) for complete function list
- Check [Examples](../examples/) for real-world usage
- Read [Architecture](architecture.md) for internals
