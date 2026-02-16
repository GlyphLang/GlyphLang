# Glyph Notation Specification

**Version:** 0.1.0-draft
**Status:** Working draft

## 1. Overview

Glyph notation is a compact, typed, language-neutral notation for describing backend services. It is designed to be consumed by both humans and AI systems, and can be compiled, interpreted, or used as a structured intent language for code generation in any target language.

A Glyph program describes **what a service does**, not how it does it in any specific language.

## 2. Symbol Vocabulary

Each glyph symbol maps to a semantic concept, not a language keyword.

| Symbol | Keyword | Semantic Meaning |
|--------|---------|------------------|
| `@` | `route` | HTTP endpoint definition |
| `:` | `type` | Type/schema definition |
| `$` | `let` | Variable binding |
| `>` | `return` | Return a value |
| `+` | `middleware` | Attach middleware/modifier |
| `%` | `use` | Inject a provider dependency |
| `<` | `expects` | Declare expected input type |
| `?` | `validate` | Validation assertion |
| `~` | `handle` | Event handler binding |
| `*` | `cron` | Scheduled task |
| `!` | `command` | CLI command definition |
| `&` | `queue` | Queue worker definition |
| `=` | `func` | Function definition |

### 2.1 Context Rules

Symbols are only interpreted as glyphs when they appear at the **start of a line** (after optional whitespace). In all other positions, they retain their standard operator meaning:

- `+` at line start = middleware; `a + b` mid-expression = addition
- `>` at line start = return; `a > b` mid-expression = comparison
- `*` at line start = cron; `a * b` mid-expression = multiplication
- `%` at line start = use/inject; `a % b` mid-expression = modulo

## 3. Type System

### 3.1 Primitive Types

| Type | Description |
|------|-------------|
| `int` | Integer (64-bit) |
| `float` | Floating point (64-bit) |
| `str` | String |
| `bool` | Boolean |
| `any` | Dynamic/untyped |

### 3.2 Composite Types

| Syntax | Description |
|--------|-------------|
| `[T]` | Array of T |
| `T?` | Optional T (nullable) |
| `T!` | Required T (non-nullable, in type fields) |
| `T \| U` | Union of T and U |
| `Name<T>` | Generic type instantiation |
| `(T) -> U` | Function type |

### 3.3 Type Definitions

```glyph
: User {
  id: int!
  name: str!
  email: str!
  age: int?
  role: str = "user"
}
```

Fields marked with `!` are required. Fields marked with `?` are optional. Fields with `= value` have defaults.

### 3.4 Target Language Type Mapping

| Glyph | Python | TypeScript | Go | Java | Rust |
|-------|--------|------------|-----|------|------|
| `int` | `int` | `number` | `int64` | `long` | `i64` |
| `float` | `float` | `number` | `float64` | `double` | `f64` |
| `str` | `str` | `string` | `string` | `String` | `String` |
| `bool` | `bool` | `boolean` | `bool` | `boolean` | `bool` |
| `[T]` | `List[T]` | `T[]` | `[]T` | `List<T>` | `Vec<T>` |
| `T?` | `Optional[T]` | `T \| null` | `*T` | `T` (nullable) | `Option<T>` |
| `any` | `Any` | `unknown` | `interface{}` | `Object` | `Box<dyn Any>` |

## 4. Providers

Providers are abstract service dependencies injected via the `%` (use) symbol.

### 4.1 Standard Providers

| Provider | Purpose |
|----------|---------|
| `Database` | Relational database access |
| `Redis` | Key-value cache/store |
| `MongoDB` | Document database |
| `LLM` | AI language model |

### 4.2 Provider Usage

```glyph
@ GET /api/users/:id {
  % db: Database
  $ user = db.users.Get(id)
  > user
}
```

### 4.3 Custom Provider Contracts

```glyph
provider ImageProcessor {
  thumbnail(file: file!, width: int!, height: int!) -> file
  resize(file: file!, width: int!, height: int!) -> file
  metadata(file: file!) -> ImageMeta
}
```

Custom providers declare an interface. Implementations are provided per target language.

### 4.4 Provider Method Semantics

Standard database provider methods follow CRUD semantics:

| Method | Meaning |
|--------|---------|
| `Get(id)` | Retrieve by primary key |
| `Find(filter?)` | Retrieve all (optionally filtered) |
| `Create(data)` | Insert a new record |
| `Update(id, data)` | Update an existing record |
| `Delete(id)` | Remove a record |
| `Where(filter)` | Query with filter conditions |
| `Count()` | Count records |

## 5. Route Definitions

```glyph
@ METHOD /path/:param {
  + middleware(args)
  + ratelimit(N/window)
  < InputType
  % provider: ProviderType
  ? condition :: statusCode "message"
  $ variable = expression
  > returnValue :: statusCode
}
```

### 5.1 HTTP Methods

`GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `WS` (WebSocket), `SSE` (Server-Sent Events)

### 5.2 Path Parameters

Prefixed with `:` — e.g., `/users/:id/posts/:postId`

### 5.3 Authentication

```glyph
+ auth(jwt)
+ auth(jwt, role: admin)
+ auth(apikey)
```

### 5.4 Rate Limiting

```glyph
+ ratelimit(100/min)
+ ratelimit(1000/hour)
```

## 6. Background Tasks

### 6.1 Cron Jobs

```glyph
* "cron-expression" task_name {
  % provider: ProviderType
  # body
}
```

### 6.2 Event Handlers

```glyph
~ "event.type" {
  % provider: ProviderType
  $ data = event
  # body
}
```

### 6.3 Queue Workers

```glyph
& "queue.name" {
  % provider: ProviderType
  $ job = message
  # body
}
```

## 7. Semantic IR

The Semantic IR is the language-neutral intermediate representation produced by analyzing a Glyph program. It captures intent independent of any target language.

### 7.1 IR Structure

```
ServiceIR
  Types[]        - Type schemas with fields, methods, traits
  Providers[]    - Required provider dependencies
  Routes[]       - HTTP route handlers
  Events[]       - Event bindings
  CronJobs[]     - Scheduled tasks
  Queues[]       - Queue worker bindings
  Commands[]     - CLI command definitions
  Functions[]    - Standalone functions
  Constants[]    - Module-level constants
```

### 7.2 Provider Resolution

Each `%` injection is resolved to a `ProviderRef` in the IR:

```
InjectionRef {
  Name: "db"              // Local variable name
  ProviderType: "Database" // Provider type for dependency resolution
}
```

The IR tracks all unique provider types required by the service, distinguishing standard providers (Database, Redis, MongoDB, LLM) from custom providers.

### 7.3 IR Expression Model

All expressions are normalized into a flat, discriminated union:

- Literals: int, float, string, bool, null
- Variables: name references
- Binary/Unary operations: with operator enum
- Field access: object.field
- Index access: array[index]
- Function calls: name(args)
- Object/Array literals
- Lambda expressions
- Pipe operations: left |> right
- Pattern matching: match/case

## 8. Compact vs Expanded Syntax

Every Glyph program has two equivalent representations:

**Compact** (`.glyph`): Uses symbols for maximum token efficiency
```
@ GET /users/:id { % db: Database $ user = db.users.Get(id) > user }
```

**Expanded** (`.glyphx`): Uses keywords for readability
```
route GET /users/:id { use db: Database let user = db.users.Get(id) return user }
```

Both are semantically identical and produce the same AST/IR.

## 9. Control Flow

```glyph
# Conditionals
if condition { ... }
if condition { ... } else { ... }

# Loops
for item in collection { ... }
for key, value in collection { ... }
while condition { ... }

# Pattern matching
match value {
  pattern => result
  _ => default
}

# Switch
switch value {
  case x { ... }
  default { ... }
}
```

## 10. Grammar (EBNF Summary)

```ebnf
program     = { item } ;
item        = type_def | route | function | cron | event | queue
            | command | import | const | provider | test ;

type_def    = ":" IDENT [ "<" type_params ">" ] [ "impl" traits ] "{" { field } "}" ;
route       = "@" method path "{" { route_stmt } "}" ;
function    = "=" IDENT "(" params ")" [ "->" type ] "{" { statement } "}" ;
cron        = "*" STRING [ IDENT ] "{" { statement } "}" ;
event       = "~" STRING "{" { statement } "}" ;
queue       = "&" STRING "{" { statement } "}" ;
command     = "!" IDENT { param } "{" { statement } "}" ;
provider    = "provider" IDENT "{" { method_sig } "}" ;

route_stmt  = middleware | injection | input_decl | validate | statement ;
middleware  = "+" IDENT "(" args ")" ;
injection   = "%" IDENT ":" type ;
input_decl  = "<" type ;
validate    = "?" expr "::" INTEGER [ STRING ] ;
statement   = assign | reassign | return | if | for | while | switch | expr_stmt ;
assign      = "$" IDENT "=" expr ;
return      = ">" expr [ "::" INTEGER ] ;

type        = "int" | "float" | "str" | "bool" | "any"
            | "[" type "]" | type "?" | type "!" | IDENT
            | IDENT "<" type { "," type } ">"
            | type "|" type
            | "(" { type } ")" "->" type ;

method      = "GET" | "POST" | "PUT" | "DELETE" | "PATCH" | "ws" | "sse" ;
```

## 11. Design Principles

1. **Intent over implementation**: Glyph describes what a service does, not how.
2. **Provider abstraction**: Services depend on capabilities, not specific libraries.
3. **Two syntaxes, one semantics**: Compact for AI, expanded for humans.
4. **Verifiable**: The type system catches mismatches at analysis time.
5. **Target-neutral IR**: The Semantic IR is the canonical representation, independent of any language.
6. **Minimal escape hatches**: Custom providers handle language-specific logic cleanly.
