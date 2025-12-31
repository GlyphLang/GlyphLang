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
| `!` | CLI Command | `! hello name: str! { ... }` |
| `<` | Input declaration | `< input: Type` |
| `=` | Function definition | `= myFunction() { ... }` |
| `~` | Event Handler | `~ "user.created" { ... }` |
| `*` | Cron Task | `* "0 0 * * *" task { ... }` |
| `&` | Queue Worker | `& "email.send" { ... }` |
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
@ route /api/users/:id -> User | Error {
  % db: Database
  $ user = db.users.get(id)
  > user
}
```

### HTTP Methods

```glyph
@ route /api/users [GET] {       # GET request
  > {users: []}
}
@ route /api/users [POST] {      # POST request
  > {created: true}
}
@ route /api/users/:id [PUT] {   # PUT request
  > {updated: true}
}
@ route /api/users/:id [DELETE] { # DELETE request
  > {deleted: true}
}
```

### Authentication

```glyph
@ route /api/admin/users {
  + auth(jwt)                  # Require JWT
  + auth(jwt, role: admin)     # Require admin role
  > {users: []}
}
```

### Rate Limiting

```glyph
@ route /api/search {
  + ratelimit(10/sec)          # 10 requests per second
  + ratelimit(100/min)         # 100 requests per minute
  + ratelimit(1000/hour)       # 1000 requests per hour
  > {results: []}
}
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
@ route /api/users [POST] {
  < input: CreateUserInput
  ! validate input {
    name: str(min=1, max=100)
    email: email_format
    age: int(min=0, max=150)
  }
  % db: Database
  $ user = db.users.create(input)
  > user
}
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
@ route /api/users/:id {
  % db: Database
  try {
    $ user = db.users.get(id)
    > user
  } catch Error {
    > {error: "User not found"}
  }
}
```

## CLI Commands

GlyphLang allows you to define CLI commands that can be executed using the `glyph exec` command. Commands use the `!` symbol.

### Basic Command

```glyph
! hello name: str! {
  $ greeting = "Hello, " + name + "!"
  > {message: greeting}
}
```

Execute with:
```bash
glyph exec main.glyph hello --name="Alice"
```

### Command with Multiple Arguments

```glyph
! add a: int! b: int! {
  $ result = a + b
  > {sum: result, a: a, b: b}
}
```

### Command with Optional Flags

```glyph
! greet name: str! --formal: bool = false {
  if formal {
    $ msg = "Good day, " + name + ". How may I assist you?"
  } else {
    $ msg = "Hey " + name + "!"
  }
  > {greeting: msg}
}
```

Execute with:
```bash
glyph exec main.glyph greet --name="Bob" --formal
```

### Command with Description

```glyph
! version "Show version information" {
  > {
    name: "GlyphLang CLI Demo",
    version: "1.0.0",
    timestamp: now()
  }
}
```

## Scheduled Tasks

GlyphLang supports cron-based scheduled tasks using the `*` symbol. Tasks run automatically based on cron expressions.

### Daily Task

```glyph
* "0 0 * * *" daily_cleanup {
  % db: Database
  $ records_deleted = db.cleanup_old_records()
  > {task: "cleanup", timestamp: now(), deleted: records_deleted}
}
```

### Hourly Stats Collection

```glyph
* "0 * * * *" hourly_stats {
  % db: Database
  $ stats = {
    timestamp: now(),
    active_users: db.users.count_active()
  }
  > stats
}
```

### Every N Minutes

```glyph
* "*/5 * * * *" health_check {
  > {status: "healthy", checked_at: now()}
}
```

### Weekly Report

```glyph
* "0 9 * * 0" weekly_report {
  + retries(3)
  % db: Database
  $ report = {
    week: "current",
    generated_at: now()
  }
  > report
}
```

**Cron Expression Format:**
- `* * * * *` - minute, hour, day of month, month, day of week
- `0 0 * * *` - Every day at midnight
- `0 * * * *` - Every hour
- `*/5 * * * *` - Every 5 minutes
- `0 9 * * 0` - Every Sunday at 9am

## Event Handlers

GlyphLang provides event-driven capabilities using the `~` symbol. Event handlers respond to application events.

### Basic Event Handler

```glyph
~ "user.created" {
  $ user_id = event.id
  $ email = event.email
  $ timestamp = now()
  > {handled: true, user_id: user_id, email: email}
}
```

### Async Event Handler

```glyph
~ "user.updated" async {
  $ user_id = event.id
  $ changes = event.changes
  > {processed: true, user_id: user_id}
}
```

### Order Completion Handler

```glyph
~ "order.completed" {
  $ order_id = event.order_id
  $ user_id = event.user_id
  $ total = event.total
  > {processed: true, order_id: order_id}
}
```

### Error Handling

```glyph
~ "payment.failed" {
  $ order_id = event.order_id
  $ reason = event.reason
  > {alert: "payment_failure", order_id: order_id, reason: reason}
}
```

**Common Events:**
- `user.created` - New user registration
- `user.updated` - User profile changes
- `order.completed` - Order processing complete
- `payment.failed` - Payment processing error

## Queue Workers

GlyphLang supports background job processing using queue workers with the `&` symbol. Workers process messages from named queues.

### Email Worker

```glyph
& "email.send" {
  + concurrency(5)
  + retries(3)
  + timeout(30)

  $ to = message.to
  $ subject = message.subject
  $ body = message.body
  $ message_id = message.id
  > {sent: true, message_id: message_id, to: to}
}
```

### Image Processing Worker

```glyph
& "image.process" {
  + concurrency(3)
  + timeout(120)

  $ image_id = message.image_id
  $ operations = message.operations
  > {processed: true, image_id: image_id}
}
```

### Report Generation Worker

```glyph
& "report.generate" {
  + concurrency(2)
  + retries(2)
  + timeout(300)

  $ report_type = message.type
  $ user_id = message.requested_by
  $ report_id = now()
  > {report_id: report_id, type: report_type}
}
```

### Webhook Delivery Worker

```glyph
& "webhook.deliver" {
  + concurrency(10)
  + retries(5)
  + timeout(15)

  $ url = message.url
  $ webhook_id = message.webhook_id
  > {delivered: true, webhook_id: webhook_id}
}
```

**Worker Modifiers:**
- `+ concurrency(n)` - Number of concurrent workers
- `+ retries(n)` - Number of retry attempts
- `+ timeout(seconds)` - Timeout in seconds

## Next Steps

- See [API Reference](api-reference.md) for complete function list
- Check [Examples](../examples/) for real-world usage
- Read [Architecture](architecture.md) for internals
