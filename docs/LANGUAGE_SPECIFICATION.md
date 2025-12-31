# GlyphLang Language Specification

Version 1.0

## Overview

GlyphLang is a domain-specific language designed for building web APIs with minimal syntax overhead. It uses symbolic tokens instead of English keywords for efficiency and language-neutrality, making it particularly well-suited for AI code generation.

---

## 1. Lexical Structure

### 1.1 Source Encoding

GlyphLang source files use UTF-8 encoding with the `.glyph` file extension.

### 1.2 Comments

GlyphLang supports two styles of comments:

```glyph
# Hash-style comment (extends to end of line)
// Double-slash comment (extends to end of line)
```

Both comment styles are equivalent and extend from the comment marker to the end of the line. Multi-line comments are not supported; each line must have its own comment marker.

### 1.3 Whitespace

- Spaces and tabs are used as token separators
- Newlines are significant as statement terminators
- Carriage returns (`\r`) are ignored
- Indentation is not syntactically significant but is used for readability

### 1.4 Tokens

#### 1.4.1 Special Symbols

| Symbol | Name | Description |
|--------|------|-------------|
| `@` | At | Route/endpoint definition |
| `:` | Colon | Type definition, type annotations |
| `$` | Dollar | Variable declaration/assignment |
| `+` | Plus | Middleware modifier, addition operator |
| `-` | Minus | Subtraction operator, flags |
| `*` | Star | Cron task definition, multiplication operator |
| `/` | Slash | Division operator, path separator |
| `%` | Percent | Dependency injection |
| `>` | Greater | Return statement, comparison operator |
| `>=` | GreaterEq | Greater than or equal comparison |
| `<` | Less | Input declaration, comparison operator |
| `<=` | LessEq | Less than or equal comparison |
| `!` | Bang | CLI command definition, required modifier, NOT operator |
| `!=` | NotEq | Not equal comparison |
| `?` | Question | Optional type modifier, validation statement |
| `~` | Tilde | Event handler definition |
| `&` | Ampersand | Queue worker definition |
| `&&` | And | Logical AND operator |
| `\|\|` | Or | Logical OR operator |
| `->` | Arrow | Return type annotation |
| `\|` | Pipe | Union type separator |
| `=` | Equals | Assignment |
| `==` | EqEq | Equality comparison |

#### 1.4.2 Delimiters

| Symbol | Name | Description |
|--------|------|-------------|
| `(` | LParen | Function call, grouping |
| `)` | RParen | Function call, grouping |
| `{` | LBrace | Block, object literal |
| `}` | RBrace | Block, object literal |
| `[` | LBracket | Array literal, array access, HTTP method |
| `]` | RBracket | Array literal, array access, HTTP method |
| `,` | Comma | Separator |
| `.` | Dot | Field access |

#### 1.4.3 Keywords

The following identifiers are reserved keywords:

| Keyword | Description |
|---------|-------------|
| `true` | Boolean true literal |
| `false` | Boolean false literal |
| `null` | Null literal |
| `if` | Conditional statement |
| `else` | Alternative branch |
| `while` | While loop |
| `for` | For loop |
| `in` | For loop iterator |
| `switch` | Switch statement |
| `case` | Switch case |
| `default` | Switch default case |
| `let` | Variable declaration (alias for `$`) |
| `return` | Return statement (alias for `>`) |
| `type` | Type definition (alias for `:`) |
| `route` | Route keyword |
| `ws` | WebSocket keyword |
| `websocket` | WebSocket keyword (alias) |
| `on` | WebSocket event handler |
| `async` | Async modifier |

### 1.5 Identifiers

Identifiers must start with a letter (a-z, A-Z) or underscore (`_`), followed by zero or more letters, digits (0-9), or underscores.

```
identifier = letter | "_" { letter | digit | "_" }
letter     = "a"..."z" | "A"..."Z"
digit      = "0"..."9"
```

Examples:
```glyph
user
_private
userName123
MAX_VALUE
```

### 1.6 Path Literals

Path literals are used in route definitions and start with a forward slash. They may contain:
- Alphanumeric characters
- Underscores
- Hyphens
- Colons (for parameters)
- Forward slashes (as separators)

```
path = "/" { segment "/" }
segment = identifier | ":" identifier
```

Examples:
```glyph
/api/users
/api/users/:id
/order-status/:status
/api/v2/products/:category/:id
```

### 1.7 String Literals

String literals are enclosed in either double quotes (`"`) or single quotes (`'`).

```glyph
"Hello, World!"
'Single-quoted string'
"String with \"escaped\" quotes"
```

Escape sequences:
| Sequence | Meaning |
|----------|---------|
| `\n` | Newline |
| `\t` | Tab |
| `\r` | Carriage return |
| `\"` | Double quote |
| `\'` | Single quote |
| `\\` | Backslash |

### 1.8 Number Literals

#### Integer Literals

Integers are sequences of decimal digits.

```glyph
0
42
1000000
```

#### Float Literals

Floats contain a decimal point followed by one or more digits.

```glyph
3.14
0.5
100.0
```

---

## 2. Type System

### 2.1 Primitive Types

| Type | Description | Example Values |
|------|-------------|----------------|
| `int` | 64-bit signed integer | `0`, `42`, `-100` |
| `str` or `string` | UTF-8 string | `"hello"`, `'world'` |
| `bool` | Boolean | `true`, `false` |
| `float` | 64-bit floating point | `3.14`, `0.5` |
| `timestamp` | Unix timestamp | `now()` |
| `any` | Dynamic type | Any value |

### 2.2 Collection Types

#### List Type

Ordered collection of elements of the same type.

```glyph
List[str]       # List of strings
List[int]       # List of integers
List[User]      # List of User objects
List[any]       # List of any type
```

#### Set Type

Unordered collection of unique elements.

```glyph
Set[str]        # Set of strings
Set[User]       # Set of User objects
```

#### Map Type

Key-value pairs.

```glyph
Map[str, str]   # String to string mapping
Map[str, int]   # String to integer mapping
```

### 2.3 Optional Types

The `?` suffix denotes an optional (nullable) type.

```glyph
: User {
  id: int!           # Required
  nickname: str?     # Optional (explicit)
  bio: str           # Optional (implicit, no !)
}
```

### 2.4 Required Modifier

The `!` suffix denotes a required (non-nullable) field.

```glyph
: User {
  id: int!          # Required integer
  name: str!        # Required string
  email: str!       # Required string
}
```

### 2.5 Union Types

Union types allow a value to be one of several types, separated by `|`.

```glyph
: Result = User | Error
: ApiResponse = SuccessResponse | ErrorResponse
```

### 2.6 Custom Type Definitions

Types are defined using the `:` symbol or `type` keyword.

```glyph
# Using colon syntax
: User {
  id: int!
  name: str!
  email: str!
  age: int
  active: bool!
  created_at: timestamp
}

# Using type keyword (alternative syntax)
type ChatMessage {
  id: int
  room: string
  sender: string
  text: string
  timestamp: int
}
```

### 2.7 Type Annotations

Type annotations appear after a colon in field definitions and function parameters.

```glyph
fieldName: Type         # Basic annotation
fieldName: Type!        # Required field
fieldName: Type?        # Optional field (explicit)
fieldName: List[Type]   # Collection type
```

---

## 3. Declarations

### 3.1 Type Definitions (`:`)

Type definitions create custom structured types.

**Syntax:**
```
":" identifier "{" field* "}"
```

**Examples:**
```glyph
: User {
  id: int!
  name: str!
  email: str!
  age: int
  active: bool!
  tags: List[str]
}

: ApiResponse {
  success: bool!
  message: str!
  data: any
  errors: List[ValidationError]
}
```

### 3.2 Route Definitions (`@`)

Route definitions create HTTP endpoints.

**Syntax:**
```
"@" "route" path [ "[" method "]" ] [ "->" returnType ] "{"
  [ middlewares ]
  [ injections ]
  body
"}"
```

All routes require braces:
```
"@" method path "{" body "}"
```

**Examples:**
```glyph
# Basic route
@ route /api/health {
  > {status: "ok"}
}

# Route with method and return type
@ route /api/users/:id [GET] -> User {
  % db: Database
  $ user = db.users.get(id)
  > user
}

# Block syntax
@ GET /api/info {
  $ info = {name: "API", version: "1.0"}
  > info
}
```

### 3.3 Function Definitions

Functions are defined using the `~` symbol for user-defined functions or inline within routes.

**Internal functions** are called using standard function call syntax:
```glyph
$ result = functionName(arg1, arg2)
$ length = input.username.length()
$ upper = upper(text)
```

### 3.4 Variable Declarations (`$`)

Variables are declared using `$` or the `let` keyword.

**Syntax:**
```
"$" identifier "=" expression
"let" identifier "=" expression
```

**Examples:**
```glyph
$ name = "Alice"
$ age = 30
$ active = true
$ users = []
$ config = {host: "localhost", port: 8080}

# Alternative syntax
let name = "Alice"
let age = 30
```

### 3.5 CLI Commands (`!`)

CLI commands define executable command-line operations.

**Syntax:**
```
"!" commandName [ description ] [ params ] "{" body "}"
```

**Examples:**
```glyph
# Simple command
! hello name: str! {
  $ greeting = "Hello, " + name + "!"
  > {message: greeting}
}

# Command with optional flag
! greet name: str! --formal: bool = false {
  if formal {
    $ msg = "Good day, " + name + "."
  } else {
    $ msg = "Hey " + name + "!"
  }
  > {greeting: msg}
}

# Command with description
! version "Show version information" {
  > {name: "MyApp", version: "1.0.0"}
}
```

### 3.6 Cron Tasks (`*`)

Cron tasks define scheduled operations using cron expressions.

**Syntax:**
```
"*" cronExpression [ taskName ] "{" body "}"
```

**Cron Expression Format:** `minute hour day-of-month month day-of-week`

| Field | Values |
|-------|--------|
| Minute | 0-59 |
| Hour | 0-23 |
| Day of Month | 1-31 |
| Month | 1-12 |
| Day of Week | 0-6 (Sunday=0) |

Special characters:
- `*` - Any value
- `*/n` - Every n units
- `,` - Value list separator

**Examples:**
```glyph
# Every minute
* "* * * * *" health_check {
  > {status: "healthy", checked_at: now()}
}

# Every day at midnight
* "0 0 * * *" daily_cleanup {
  % db: Database
  > {task: "cleanup", timestamp: now()}
}

# Every 5 minutes
* "*/5 * * * *" health_check {
  > {status: "ok"}
}

# Sundays at 9am
* "0 9 * * 0" weekly_report {
  + retries(3)
  % db: Database
  > {report: "weekly", generated_at: now()}
}
```

### 3.7 Event Handlers (`~`)

Event handlers respond to application events.

**Syntax:**
```
"~" eventType [ "async" ] "{" body "}"
```

**Examples:**
```glyph
# Synchronous handler
~ "user.created" {
  $ user_id = event.id
  $ email = event.email
  > {handled: true, user_id: user_id}
}

# Asynchronous handler
~ "user.updated" async {
  $ user_id = event.id
  $ changes = event.changes
  > {processed: true}
}

# Order event
~ "order.completed" {
  $ order_id = event.order_id
  $ total = event.total
  > {processed: true, order_id: order_id}
}
```

### 3.8 Queue Workers (`&`)

Queue workers process messages from named queues.

**Syntax:**
```
"&" queueName "{" [ config ] body "}"
```

**Configuration Options:**
- `+ concurrency(n)` - Number of concurrent workers
- `+ retries(n)` - Number of retry attempts
- `+ timeout(seconds)` - Processing timeout

**Examples:**
```glyph
# Email worker
& "email.send" {
  + concurrency(5)
  + retries(3)
  + timeout(30)

  $ to = message.to
  $ subject = message.subject
  $ body = message.body
  > {sent: true, to: to}
}

# Image processing worker
& "image.process" {
  + concurrency(3)
  + timeout(120)

  $ image_id = message.image_id
  $ operations = message.operations
  > {processed: true, image_id: image_id}
}
```

---

## 4. Expressions

### 4.1 Literal Expressions

#### String Literals
```glyph
"Hello, World!"
'Single quotes also work'
"Escaped \"quotes\""
```

#### Number Literals
```glyph
42          # Integer
3.14        # Float
-100        # Negative integer
0.5         # Float less than 1
```

#### Boolean Literals
```glyph
true
false
```

#### Null Literal
```glyph
null
```

#### Object Literals
```glyph
{name: "Alice", age: 30}
{
  id: 1,
  username: "alice",
  email: "alice@example.com",
  active: true
}
```

#### Array Literals
```glyph
[1, 2, 3, 4, 5]
["apple", "banana", "cherry"]
[{name: "Alice"}, {name: "Bob"}]
[[1, 2], [3, 4], [5, 6]]
```

### 4.2 Binary Operators

#### Arithmetic Operators

| Operator | Description | Precedence |
|----------|-------------|------------|
| `+` | Addition, string concatenation | 10 |
| `-` | Subtraction | 10 |
| `*` | Multiplication | 20 |
| `/` | Division | 20 |

```glyph
$ sum = a + b
$ diff = a - b
$ product = a * b
$ quotient = a / b
$ greeting = "Hello, " + name + "!"
```

#### Comparison Operators

| Operator | Description | Precedence |
|----------|-------------|------------|
| `==` | Equal | 5 |
| `!=` | Not equal | 5 |
| `<` | Less than | 5 |
| `<=` | Less than or equal | 5 |
| `>` | Greater than | 5 |
| `>=` | Greater than or equal | 5 |

```glyph
$ isEqual = a == b
$ notEqual = a != b
$ lessThan = a < b
$ lessOrEqual = a <= b
$ greaterThan = a > b
$ greaterOrEqual = a >= b
```

#### Logical Operators

| Operator | Description | Precedence |
|----------|-------------|------------|
| `&&` | Logical AND | 3 |
| `\|\|` | Logical OR | 2 |

```glyph
$ bothTrue = condition1 && condition2
$ eitherTrue = condition1 || condition2
$ complex = (a > 5 && b < 10) || c == 0
```

### 4.3 Unary Operators

| Operator | Description |
|----------|-------------|
| `!` | Logical NOT |
| `-` | Negation |

```glyph
$ notActive = !active
$ negative = -value
$ doubleNot = !!confirmed
```

### 4.4 Field Access

Access object fields using dot notation.

```glyph
$ name = user.name
$ city = user.address.city
$ first = users[0].name
```

### 4.5 Array Indexing

Access array elements using bracket notation.

```glyph
$ first = items[0]
$ second = items[1]
$ nested = matrix[0][1]
$ value = data.items[index]
```

### 4.6 Function Calls

Call functions with parentheses and comma-separated arguments.

```glyph
$ len = length(text)
$ upper = upper(name)
$ result = parseInt(value)
$ max = max(a, b)
```

### 4.7 Method Calls

Call methods on objects using dot notation.

```glyph
$ user = db.users.get(id)
$ all = db.users.all()
$ filtered = db.users.filter("active", true)
$ len = input.username.length()
$ hasAt = email.contains("@")
```

### 4.8 Operator Precedence

From highest to lowest:

1. Parentheses `()`
2. Field access `.`, array index `[]`, function call `()`
3. Unary operators `!`, `-`
4. Multiplicative `*`, `/`
5. Additive `+`, `-`
6. Comparison `<`, `<=`, `>`, `>=`
7. Equality `==`, `!=`
8. Logical AND `&&`
9. Logical OR `||`

---

## 5. Statements

### 5.1 Assignment Statements

Assign values to variables using `$` or `let`.

```glyph
$ name = "Alice"
$ age = 30
$ user.name = "Bob"
let active = true
```

### 5.2 If/Else Statements

Conditional execution with optional else clause.

**Syntax:**
```
"if" expression "{" statements "}" [ "else" "{" statements "}" ]
"if" expression "{" statements "}" [ "else" "if" expression "{" statements "}" ]* [ "else" "{" statements "}" ]
```

**Examples:**
```glyph
if age >= 18 {
  $ category = "adult"
}

if user == null {
  > {error: "User not found"}
} else {
  > user
}

if age < 13 {
  $ category = "child"
} else if age < 20 {
  $ category = "teenager"
} else {
  $ category = "adult"
}
```

### 5.3 While Loops

Repeat statements while a condition is true.

**Syntax:**
```
"while" expression "{" statements "}"
```

**Examples:**
```glyph
$ i = 0
$ sum = 0
while i < 10 {
  $ sum = sum + i
  $ i = i + 1
}

$ count = 0
while count < items.length() {
  $ item = items[count]
  $ count = count + 1
}
```

### 5.4 For Loops

Iterate over arrays or objects.

**Syntax:**
```
"for" identifier "in" expression "{" statements "}"
"for" keyIdentifier "," valueIdentifier "in" expression "{" statements "}"
```

**Examples:**
```glyph
# Simple iteration
for user in users {
  $ name = user.name
}

# Iteration with index
for index, item in items {
  $ entry = {position: index, value: item}
}

# Object iteration (key-value pairs)
for key, value in config {
  $ setting = {key: key, value: value}
}

# Nested loops
for row in matrix {
  for cell in row {
    $ sum = sum + cell
  }
}
```

### 5.5 Switch Statements

Multi-way branching based on value matching.

**Syntax:**
```
"switch" expression "{"
  ( "case" expression "{" statements "}" )*
  [ "default" "{" statements "}" ]
"}"
```

**Examples:**
```glyph
switch status {
  case "pending" {
    $ message = "Order is pending"
  }
  case "shipped" {
    $ message = "Order has shipped"
  }
  case "delivered" {
    $ message = "Order delivered"
  }
  default {
    $ message = "Unknown status"
  }
}

switch score {
  case 90 {
    $ grade = "A"
  }
  case 80 {
    $ grade = "B"
  }
  case 70 {
    $ grade = "C"
  }
  default {
    $ grade = "F"
  }
}
```

### 5.6 Return Statements

Return a value from a route or function using `>` or `return`.

```glyph
> {success: true, data: user}
> result
return {status: "ok"}
```

---

## 6. Routes

### 6.1 Path Syntax

Routes are defined with paths that can include static segments and parameters.

```glyph
# Static path
@ route /api/health {
  > {status: "ok"}
}

# Path with parameter
@ route /api/users/:id {
  > {id: id}
}

# Multiple parameters
@ route /api/users/:userId/posts/:postId {
  > {userId: userId, postId: postId}
}

# Hyphenated segments
@ route /api/order-status/:orderId {
  > {orderId: orderId}
}
```

### 6.2 HTTP Methods

Specify HTTP methods using bracket notation or method keywords.

**Bracket Syntax:**
```glyph
@ route /api/users [GET] {
  > {users: []}
}
@ route /api/users [POST] {
  > {created: true}
}
@ route /api/users/:id [PUT] {
  > {updated: true}
}
@ route /api/users/:id [PATCH] {
  > {patched: true}
}
@ route /api/users/:id [DELETE] {
  > {deleted: true}
}
```

**Method Keyword Syntax:**
```glyph
@ GET /api/users { ... }
@ POST /api/users { ... }
@ PUT /api/users/:id { ... }
@ PATCH /api/users/:id { ... }
@ DELETE /api/users/:id { ... }
```

Supported methods:
- `GET` - Retrieve resources
- `POST` - Create resources
- `PUT` - Replace resources
- `PATCH` - Partially update resources
- `DELETE` - Delete resources

### 6.3 Path Parameters

Path parameters are denoted with a colon prefix and are available as variables in the route body.

```glyph
@ route /api/users/:id [GET] {
  % db: Database
  $ user = db.users.get(id)  # 'id' is the path parameter
  > user
}

@ route /api/users/:userId/posts/:postId [GET] {
  % db: Database
  $ post = db.posts.getByUserAndId(userId, postId)
  > post
}
```

### 6.4 Query Parameters

Query parameters are accessed via the `input` object.

```glyph
@ route /api/search [GET] {
  $ query = input.q          # ?q=search+term
  $ page = input.page        # ?page=1
  $ limit = input.limit      # ?limit=20
  > {query: query, page: page, limit: limit}
}
```

### 6.5 Request Body

Request body data is accessed via the `input` object for POST, PUT, and PATCH requests.

```glyph
@ route /api/users [POST] {
  $ username = input.username
  $ email = input.email
  $ password = input.password
  > {created: true}
}
```

### 6.6 Return Types

Specify the return type using the arrow syntax.

```glyph
@ route /api/users/:id [GET] -> User {
  % db: Database
  $ user = db.users.get(id)
  > user
}

@ route /api/users [GET] -> List[User] {
  % db: Database
  $ users = db.users.all()
  > users
}

@ route /api/auth/login [POST] -> AuthResponse | Error {
  # ...
}
```

---

## 7. Middleware

### 7.1 Auth Middleware (`+auth`)

Require authentication for a route.

**Syntax:**
```
"+" "auth" "(" authType [ "," options ] ")"
```

**Examples:**
```glyph
# Basic JWT authentication
@ route /api/users [GET] {
  + auth(jwt)
  > {users: []}
}

# Role-based authentication
@ route /api/admin/users [GET] {
  + auth(jwt, role: admin)
  > {users: []}
}

# Moderator access
@ route /api/posts/:id [DELETE] {
  + auth(jwt, role: moderator)
  > {deleted: true}
}
```

When auth middleware is applied, the `auth` object is available with user information:
```glyph
$ userId = auth.user.id
$ userRole = auth.user.role
$ username = auth.user.username
```

### 7.2 Rate Limiting (`+ratelimit`)

Limit request frequency.

**Syntax:**
```
"+" "ratelimit" "(" count "/" window ")"
```

**Window units:**
- `sec` - Seconds
- `min` - Minutes
- `hour` - Hours

**Examples:**
```glyph
@ route /api/search [GET] {
  + ratelimit(10/sec)        # 10 requests per second
  > {results: []}
}

@ route /api/users [GET] {
  + ratelimit(100/min)       # 100 requests per minute
  > {users: []}
}

@ route /api/export [POST] {
  + ratelimit(5/hour)        # 5 requests per hour
  > {exported: true}
}
```

### 7.3 Combining Middleware

Multiple middleware can be applied to a single route.

```glyph
@ route /api/admin/reports [GET] {
  + auth(jwt, role: admin)
  + ratelimit(50/min)
  % db: Database
  $ reports = db.reports.all()
  > reports
}
```

---

## 8. Dependency Injection

### 8.1 Service Injection (`%`)

Inject services into routes using the `%` symbol.

**Syntax:**
```
"%" identifier ":" Type
```

**Examples:**
```glyph
@ route /api/users [GET] {
  % db: Database
  $ users = db.users.all()
  > users
}

# Multiple injections
@ route /api/checkout [POST] {
  % db: Database
  % cache: Cache
  % payment: PaymentService
  # ...
}
```

### 8.2 Database Operations

The `Database` type provides standard CRUD operations:

```glyph
% db: Database

# Read operations
$ user = db.users.get(id)           # Get by ID
$ allUsers = db.users.all()         # Get all
$ found = db.users.findOne("email", email)  # Find by field
$ filtered = db.users.filter("active", true) # Filter by field
$ count = db.users.count("active", true)    # Count matching

# Write operations
$ created = db.users.create({name: "Alice", email: "alice@example.com"})
$ updated = db.users.update(id, {name: "Bob"})
$ deleted = db.users.delete(id)
$ deletedCount = db.users.deleteWhere("active", false)

# Utility operations
$ nextId = db.users.nextId()
```

---

## 9. WebSocket Routes

### 9.1 WebSocket Definition

WebSocket routes use the `ws` or `websocket` keyword.

**Syntax:**
```
"@" "ws" path "{"
  [ "on" "connect" "{" statements "}" ]
  [ "on" "message" "{" statements "}" ]
  [ "on" "disconnect" "{" statements "}" ]
  [ "on" "error" "{" statements "}" ]
"}"
```

**Examples:**
```glyph
@ ws /chat {
  on connect {
    ws.join("lobby")
    ws.broadcast("User joined")
  }

  on message {
    ws.broadcast(input)
  }

  on disconnect {
    ws.broadcast("User left")
    ws.leave("lobby")
  }
}
```

### 9.2 WebSocket with Path Parameters

```glyph
@ ws /chat/:room {
  on connect {
    ws.join(room)
    ws.broadcast_to_room(room, "User joined")
  }

  on message {
    ws.broadcast_to_room(room, input)
  }

  on disconnect {
    ws.broadcast_to_room(room, "User left")
    ws.leave(room)
  }
}
```

### 9.3 WebSocket Functions

| Function | Description |
|----------|-------------|
| `ws.send(message)` | Send message to current client |
| `ws.broadcast(message)` | Broadcast to all clients |
| `ws.broadcast_to_room(room, message)` | Broadcast to room members |
| `ws.join(room)` | Join a room |
| `ws.leave(room)` | Leave a room |
| `ws.get_rooms()` | Get all room names |
| `ws.get_room_count()` | Get number of rooms |
| `ws.get_connection_count()` | Get total connections |
| `ws.get_uptime()` | Get server uptime |

---

## 10. Built-in Functions

### 10.1 String Functions

| Function | Description | Example |
|----------|-------------|---------|
| `length(str)` | String length | `length("hello")` returns `5` |
| `upper(str)` | Uppercase | `upper("hello")` returns `"HELLO"` |
| `lower(str)` | Lowercase | `lower("HELLO")` returns `"hello"` |
| `trim(str)` | Remove whitespace | `trim("  hi  ")` returns `"hi"` |
| `contains(str, substr)` | Check substring | `contains("hello", "ell")` returns `true` |
| `startsWith(str, prefix)` | Check prefix | `startsWith("hello", "he")` returns `true` |
| `endsWith(str, suffix)` | Check suffix | `endsWith("hello", "lo")` returns `true` |
| `indexOf(str, substr)` | Find position | `indexOf("hello", "l")` returns `2` |
| `replace(str, old, new)` | Replace substring | `replace("hello", "l", "L")` |
| `substring(str, start, end)` | Extract substring | `substring("hello", 0, 2)` returns `"he"` |
| `charAt(str, index)` | Get character | `charAt("hello", 0)` returns `"h"` |
| `split(str, delimiter)` | Split string | `split("a,b,c", ",")` returns `["a","b","c"]` |
| `join(arr, delimiter)` | Join array | `join(["a","b"], "-")` returns `"a-b"` |

### 10.2 Numeric Functions

| Function | Description | Example |
|----------|-------------|---------|
| `abs(num)` | Absolute value | `abs(-5)` returns `5` |
| `min(a, b)` | Minimum value | `min(3, 7)` returns `3` |
| `max(a, b)` | Maximum value | `max(3, 7)` returns `7` |
| `parseInt(str)` | Parse integer | `parseInt("42")` returns `42` |
| `parseFloat(str)` | Parse float | `parseFloat("3.14")` returns `3.14` |
| `toString(val)` | Convert to string | `toString(42)` returns `"42"` |

### 10.3 Date/Time Functions

| Function | Description |
|----------|-------------|
| `now()` | Current Unix timestamp |

### 10.4 Crypto Functions

| Function | Description |
|----------|-------------|
| `crypto.hash(password)` | Hash a password |
| `crypto.verify(password, hash)` | Verify password against hash |

### 10.5 JWT Functions

| Function | Description |
|----------|-------------|
| `jwt.sign(payload, duration)` | Create JWT token |
| `jwt.verify(token)` | Verify JWT token |

---

## 11. Special Variables

### 11.1 Input Variable

The `input` variable contains request data:
- Query parameters (GET requests)
- Request body (POST, PUT, PATCH requests)

```glyph
$ username = input.username
$ email = input.email
$ page = input.page
```

### 11.2 Auth Variable

Available when `+auth` middleware is applied:

```glyph
$ userId = auth.user.id
$ role = auth.user.role
$ username = auth.user.username
```

### 11.3 Event Variable

Available in event handlers:

```glyph
~ "user.created" {
  $ userId = event.id
  $ email = event.email
}
```

### 11.4 Message Variable

Available in queue workers:

```glyph
& "email.send" {
  $ to = message.to
  $ subject = message.subject
  $ body = message.body
}
```

---

## 12. Grammar Summary

```ebnf
Module      = Item*
Item        = TypeDef | Route | Command | CronTask | EventHandler | QueueWorker

TypeDef     = ":" Identifier "{" Field* "}"
            | "type" Identifier "{" Field* "}"
Field       = Identifier ":" Type ["!"]

Route       = "@" "route" Path ["[" Method "]"] ["->" Type] "{" Middleware* Injection* Statement* "}"
            | "@" Method Path "{" Statement* "}"

Command     = "!" Identifier [String] Param* ["->" Type] "{" Statement* "}"
CronTask    = "*" String [Identifier] "{" Statement* "}"
EventHandler = "~" String ["async"] "{" Statement* "}"
QueueWorker = "&" String "{" Config* Statement* "}"

Path        = "/" PathSegment*
PathSegment = Identifier | ":" Identifier
Method      = "GET" | "POST" | "PUT" | "PATCH" | "DELETE"

Type        = "int" | "str" | "string" | "bool" | "float"
            | "timestamp" | "any"
            | Identifier
            | Identifier "[" Type "]"
            | Type "|" Type

Middleware  = "+" "auth" "(" Identifier ["," options] ")"
            | "+" "ratelimit" "(" Integer "/" Identifier ")"
Injection   = "%" Identifier ":" Type

Statement   = Assignment | Return | If | While | For | Switch | ExprStmt
Assignment  = ("$" | "let") Identifier "=" Expr
Return      = (">" | "return") Expr
If          = "if" Expr "{" Statement* "}" ["else" ("{" Statement* "}" | If)]
While       = "while" Expr "{" Statement* "}"
For         = "for" Identifier ["," Identifier] "in" Expr "{" Statement* "}"
Switch      = "switch" Expr "{" Case* [Default] "}"
Case        = "case" Expr "{" Statement* "}"
Default     = "default" "{" Statement* "}"

Expr        = OrExpr
OrExpr      = AndExpr ("||" AndExpr)*
AndExpr     = EqExpr ("&&" EqExpr)*
EqExpr      = CmpExpr (("==" | "!=") CmpExpr)*
CmpExpr     = AddExpr (("<" | "<=" | ">" | ">=") AddExpr)*
AddExpr     = MulExpr (("+" | "-") MulExpr)*
MulExpr     = UnaryExpr (("*" | "/") UnaryExpr)*
UnaryExpr   = ("!" | "-") UnaryExpr | Primary
Primary     = Integer | Float | String | "true" | "false" | "null"
            | Identifier
            | Identifier "." Identifier
            | Identifier "[" Expr "]"
            | Identifier "(" [Expr ("," Expr)*] ")"
            | "{" [ObjectField ("," ObjectField)*] "}"
            | "[" [Expr ("," Expr)*] "]"
            | "(" Expr ")"
ObjectField = Identifier ":" Expr

Identifier  = Letter (Letter | Digit | "_")*
Integer     = Digit+
Float       = Digit+ "." Digit+
String      = '"' Character* '"' | "'" Character* "'"
```

---

## 13. Examples

### 13.1 Complete REST API Example

```glyph
# User type definition
: User {
  id: int!
  username: str!
  email: str!
  role: str!
  active: bool!
  created_at: timestamp
}

# Health check
@ route /api/health [GET] {
  > {status: "ok", timestamp: now()}
}

# Get all users
@ route /api/users [GET] {
  + auth(jwt)
  + ratelimit(100/min)
  % db: Database

  $ users = db.users.all()
  > {users: users, count: users.length()}
}

# Get user by ID
@ route /api/users/:id [GET] {
  + auth(jwt)
  % db: Database

  $ user = db.users.get(id)
  if user == null {
    > {success: false, message: "User not found"}
  } else {
    > user
  }
}

# Create user
@ route /api/users [POST] {
  + auth(jwt, role: admin)
  + ratelimit(10/min)
  % db: Database

  $ hashedPassword = crypto.hash(input.password)
  $ user = db.users.create({
    username: input.username,
    email: input.email,
    password: hashedPassword,
    role: "user",
    active: true,
    created_at: now()
  })
  > {success: true, user: user}
}

# Update user
@ route /api/users/:id [PUT] {
  + auth(jwt)
  % db: Database

  $ user = db.users.get(id)
  if user == null {
    > {success: false, message: "User not found"}
  } else {
    $ updated = db.users.update(id, {
      username: input.username,
      email: input.email
    })
    > {success: true, user: updated}
  }
}

# Delete user
@ route /api/users/:id [DELETE] {
  + auth(jwt, role: admin)
  % db: Database

  $ user = db.users.get(id)
  if user == null {
    > {success: false, message: "User not found"}
  } else {
    $ result = db.users.delete(id)
    > {success: true, message: "User deleted"}
  }
}
```

### 13.2 Event-Driven Example

```glyph
# User registration event
~ "user.created" {
  $ userId = event.id
  $ email = event.email

  # Send welcome email via queue
  > {queued: "email.send", user_id: userId}
}

# Email worker
& "email.send" {
  + concurrency(5)
  + retries(3)
  + timeout(30)

  $ to = message.to
  $ subject = message.subject
  $ body = message.body
  > {sent: true, to: to}
}

# Scheduled cleanup
* "0 0 * * *" daily_cleanup {
  % db: Database
  $ deleted = db.sessions.deleteWhere("expired", true)
  > {task: "cleanup", deleted_count: deleted}
}
```

---

## Version History

| Version | Date | Description |
|---------|------|-------------|
| 1.0 | 2024 | Initial specification |
