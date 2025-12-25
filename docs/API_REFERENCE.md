# GlyphLang API Reference

This document provides a comprehensive reference for all built-in functions, operators, and APIs available in GlyphLang.

## Table of Contents

- [String Functions](#string-functions)
- [Numeric Functions](#numeric-functions)
- [Type Conversion Functions](#type-conversion-functions)
- [Date/Time Functions](#datetime-functions)
- [Crypto Functions](#crypto-functions)
- [Database API](#database-api)
- [WebSocket API](#websocket-api)
- [Request/Response Objects](#requestresponse-objects)
- [Collection Operations](#collection-operations)
- [Operators](#operators)
- [Control Flow](#control-flow)

---

## String Functions

### length

Returns the length of a string or array.

**Signature:**
```
length(value: str | List) -> int
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| value | str or List | The string or array to measure |

**Return Type:** `int`

**Example:**
```glyph
$ text = "Hello, World!"
$ len = length(text)  # Returns 13

$ items = [1, 2, 3, 4, 5]
$ count = length(items)  # Returns 5
```

---

### upper

Converts a string to uppercase.

**Signature:**
```
upper(str: str) -> str
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The string to convert |

**Return Type:** `str`

**Example:**
```glyph
$ text = "hello world"
$ result = upper(text)  # Returns "HELLO WORLD"
```

---

### lower

Converts a string to lowercase.

**Signature:**
```
lower(str: str) -> str
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The string to convert |

**Return Type:** `str`

**Example:**
```glyph
$ text = "Hello World"
$ result = lower(text)  # Returns "hello world"
```

---

### trim

Removes leading and trailing whitespace from a string.

**Signature:**
```
trim(str: str) -> str
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The string to trim |

**Return Type:** `str`

**Example:**
```glyph
$ text = "  hello world  "
$ result = trim(text)  # Returns "hello world"
```

---

### contains

Checks if a string contains a substring.

**Signature:**
```
contains(str: str, substr: str) -> bool
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The string to search in |
| substr | str | The substring to search for |

**Return Type:** `bool`

**Example:**
```glyph
$ text = "hello world"
$ hasHello = contains(text, "hello")  # Returns true
$ hasFoo = contains(text, "foo")      # Returns false
```

---

### split

Splits a string into an array using a delimiter.

**Signature:**
```
split(str: str, delimiter: str) -> List[str]
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The string to split |
| delimiter | str | The delimiter to split on |

**Return Type:** `List[str]`

**Example:**
```glyph
$ text = "apple,banana,cherry"
$ fruits = split(text, ",")  # Returns ["apple", "banana", "cherry"]
```

---

### join

Joins an array of strings with a delimiter.

**Signature:**
```
join(arr: List, delimiter: str) -> str
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| arr | List | The array to join |
| delimiter | str | The delimiter to use between elements |

**Return Type:** `str`

**Example:**
```glyph
$ words = ["hello", "world"]
$ result = join(words, " ")  # Returns "hello world"

$ nums = [1, 2, 3]
$ csv = join(nums, ",")  # Returns "1,2,3"
```

---

### replace

Replaces all occurrences of a substring with another string.

**Signature:**
```
replace(str: str, old: str, new: str) -> str
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The original string |
| old | str | The substring to replace |
| new | str | The replacement string |

**Return Type:** `str`

**Example:**
```glyph
$ text = "hello world"
$ result = replace(text, "world", "universe")  # Returns "hello universe"
```

---

### substring

Extracts a portion of a string between start and end indices.

**Signature:**
```
substring(str: str, start: int, end: int) -> str
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The string to extract from |
| start | int | The starting index (inclusive, 0-based) |
| end | int | The ending index (exclusive) |

**Return Type:** `str`

**Example:**
```glyph
$ text = "hello world"
$ result = substring(text, 0, 5)  # Returns "hello"
$ result2 = substring(text, 6, 11)  # Returns "world"
```

---

### charAt

Gets the character at a specific index.

**Signature:**
```
charAt(str: str, index: int) -> str
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The string to get the character from |
| index | int | The index of the character (0-based) |

**Return Type:** `str`

**Example:**
```glyph
$ text = "hello"
$ first = charAt(text, 0)  # Returns "h"
$ last = charAt(text, 4)   # Returns "o"
```

---

### startsWith

Checks if a string starts with a specific prefix.

**Signature:**
```
startsWith(str: str, prefix: str) -> bool
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The string to check |
| prefix | str | The prefix to look for |

**Return Type:** `bool`

**Example:**
```glyph
$ text = "hello world"
$ result = startsWith(text, "hello")  # Returns true
$ result2 = startsWith(text, "world")  # Returns false
```

---

### endsWith

Checks if a string ends with a specific suffix.

**Signature:**
```
endsWith(str: str, suffix: str) -> bool
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The string to check |
| suffix | str | The suffix to look for |

**Return Type:** `bool`

**Example:**
```glyph
$ text = "hello world"
$ result = endsWith(text, "world")  # Returns true
$ result2 = endsWith(text, "hello")  # Returns false
```

---

### indexOf

Finds the first occurrence of a substring and returns its index.

**Signature:**
```
indexOf(str: str, substr: str) -> int
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The string to search in |
| substr | str | The substring to find |

**Return Type:** `int` (Returns -1 if not found)

**Example:**
```glyph
$ text = "hello world"
$ pos = indexOf(text, "world")  # Returns 6
$ notFound = indexOf(text, "foo")  # Returns -1
```

---

## Numeric Functions

### abs

Returns the absolute value of a number.

**Signature:**
```
abs(value: int | float) -> int | float
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| value | int or float | The number to get the absolute value of |

**Return Type:** `int` or `float` (same as input type)

**Example:**
```glyph
$ positive = abs(-42)    # Returns 42
$ float_val = abs(-3.14) # Returns 3.14
$ already_pos = abs(10)  # Returns 10
```

---

### min

Returns the minimum of two values.

**Signature:**
```
min(a: int | float, b: int | float) -> int | float
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| a | int or float | First value |
| b | int or float | Second value |

**Return Type:** `int` or `float` (same as input type)

**Example:**
```glyph
$ smaller = min(10, 5)     # Returns 5
$ smaller_float = min(3.14, 2.71)  # Returns 2.71
```

---

### max

Returns the maximum of two values.

**Signature:**
```
max(a: int | float, b: int | float) -> int | float
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| a | int or float | First value |
| b | int or float | Second value |

**Return Type:** `int` or `float` (same as input type)

**Example:**
```glyph
$ larger = max(10, 5)      # Returns 10
$ larger_float = max(3.14, 2.71)  # Returns 3.14
```

---

## Type Conversion Functions

### parseInt

Parses a string to an integer.

**Signature:**
```
parseInt(str: str) -> int
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The string to parse |

**Return Type:** `int`

**Example:**
```glyph
$ num = parseInt("42")     # Returns 42
$ negative = parseInt("-10")  # Returns -10
```

---

### parseFloat

Parses a string to a floating-point number.

**Signature:**
```
parseFloat(str: str) -> float
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| str | str | The string to parse |

**Return Type:** `float`

**Example:**
```glyph
$ pi = parseFloat("3.14159")  # Returns 3.14159
$ negative = parseFloat("-2.5")  # Returns -2.5
```

---

### toString

Converts any value to a string representation.

**Signature:**
```
toString(value: any) -> str
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| value | any | The value to convert |

**Return Type:** `str`

**Example:**
```glyph
$ numStr = toString(42)      # Returns "42"
$ floatStr = toString(3.14)  # Returns "3.14"
$ boolStr = toString(true)   # Returns "true"
```

---

## Date/Time Functions

### now

Returns the current Unix timestamp in seconds.

**Signature:**
```
now() -> int
```

**Parameters:** None

**Return Type:** `int` (Unix timestamp)

**Example:**
```glyph
$ timestamp = now()  # Returns current Unix timestamp, e.g., 1234567890

$ user = db.users.create({
  username: "alice",
  created_at: now()
})
```

---

### time.now

Returns the current Unix timestamp (alias for `now`).

**Signature:**
```
time.now() -> int
```

**Parameters:** None

**Return Type:** `int` (Unix timestamp)

**Example:**
```glyph
$ timestamp = time.now()
```

---

## Crypto Functions

### crypto.hash

Hashes a value using a secure hashing algorithm (typically bcrypt for passwords).

**Signature:**
```
crypto.hash(value: str) -> str
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| value | str | The value to hash |

**Return Type:** `str` (hashed value)

**Example:**
```glyph
$ password = input.password
$ hashedPassword = crypto.hash(password)

$ newUser = db.users.create({
  username: input.username,
  password: hashedPassword
})
```

---

### crypto.verify

Verifies a plain value against a hashed value.

**Signature:**
```
crypto.verify(plain: str, hashed: str) -> bool
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| plain | str | The plain text value to verify |
| hashed | str | The hashed value to compare against |

**Return Type:** `bool`

**Example:**
```glyph
$ user = db.users.findOne("username", input.username)
$ passwordValid = crypto.verify(input.password, user.password)

if passwordValid == false {
  > {success: false, message: "Invalid credentials"}
}
```

---

### jwt.sign

Creates a signed JWT token with the given payload and expiration.

**Signature:**
```
jwt.sign(payload: object, expiration: str) -> str
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| payload | object | The data to encode in the token |
| expiration | str | Token expiration duration (e.g., "24h", "7d") |

**Return Type:** `str` (JWT token)

**Example:**
```glyph
$ token = jwt.sign({
  user_id: user.id,
  username: user.username,
  role: user.role
}, "7d")

> {
  success: true,
  token: token,
  expires_in: "7 days"
}
```

---

### jwt.verify

Verifies and decodes a JWT token.

**Signature:**
```
jwt.verify(token: str) -> object | null
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| token | str | The JWT token to verify |

**Return Type:** `object` (decoded payload) or `null` if invalid

**Example:**
```glyph
$ decoded = jwt.verify(authToken)

if decoded == null {
  > {success: false, message: "Invalid token"}
} else {
  $ userId = decoded.user_id
}
```

---

## Database API

The database API is accessed through dependency injection using `% db: Database`.

### db.{table}.all

Retrieves all records from a table.

**Signature:**
```
db.{table}.all() -> List[object]
```

**Return Type:** `List[object]`

**Example:**
```glyph
@ route /api/users [GET]
  % db: Database

  $ users = db.users.all()
  > {users: users, count: users.length()}
```

---

### db.{table}.get

Retrieves a single record by ID.

**Signature:**
```
db.{table}.get(id: int | str) -> object | null
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| id | int or str | The record ID |

**Return Type:** `object` or `null` if not found

**Example:**
```glyph
@ route /api/users/:id [GET]
  % db: Database

  $ user = db.users.get(id)

  if user == null {
    > {success: false, message: "User not found"}
  } else {
    > user
  }
```

---

### db.{table}.create

Creates a new record in the table.

**Signature:**
```
db.{table}.create(data: object) -> object
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| data | object | The record data to insert |

**Return Type:** `object` (the created record with ID)

**Example:**
```glyph
@ route /api/users [POST]
  % db: Database

  $ newUser = db.users.create({
    username: input.username,
    email: input.email,
    active: true,
    created_at: now()
  })

  > {success: true, user: newUser}
```

---

### db.{table}.update

Updates an existing record.

**Signature:**
```
db.{table}.update(id: int | str, data: object) -> object
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| id | int or str | The record ID to update |
| data | object | The fields to update |

**Return Type:** `object` (the updated record)

**Example:**
```glyph
@ route /api/users/:id [PUT]
  % db: Database

  $ updated = db.users.update(id, {
    username: input.username,
    email: input.email,
    updated_at: now()
  })

  > {success: true, user: updated}
```

---

### db.{table}.delete

Deletes a record by ID.

**Signature:**
```
db.{table}.delete(id: int | str) -> bool
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| id | int or str | The record ID to delete |

**Return Type:** `bool`

**Example:**
```glyph
@ route /api/users/:id [DELETE]
  % db: Database

  $ result = db.users.delete(id)
  > {success: true, message: "User deleted"}
```

---

### db.{table}.filter

Filters records by a field value.

**Signature:**
```
db.{table}.filter(field: str, value: any) -> List[object]
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| field | str | The field name to filter on |
| value | any | The value to match |

**Return Type:** `List[object]`

**Example:**
```glyph
@ route /api/users/active [GET]
  % db: Database

  $ activeUsers = db.users.filter("active", true)
  > {users: activeUsers}
```

---

### db.{table}.findOne

Finds a single record matching a field value.

**Signature:**
```
db.{table}.findOne(field: str, value: any) -> object | null
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| field | str | The field name to search |
| value | any | The value to match |

**Return Type:** `object` or `null`

**Example:**
```glyph
$ user = db.users.findOne("email", input.email)

if user != null {
  > {success: false, message: "Email already registered"}
}
```

---

### db.{table}.count

Counts records matching a field value.

**Signature:**
```
db.{table}.count(field: str, value: any) -> int
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| field | str | The field name to count |
| value | any | The value to match |

**Return Type:** `int`

**Example:**
```glyph
$ activeCount = db.users.count("active", true)
$ inactiveCount = db.users.count("active", false)

> {active: activeCount, inactive: inactiveCount}
```

---

### db.{table}.length

Returns the total number of records in a table.

**Signature:**
```
db.{table}.length() -> int
```

**Return Type:** `int`

**Example:**
```glyph
$ totalUsers = db.users.length()
> {total: totalUsers}
```

---

### db.{table}.nextId

Gets the next available ID for new records.

**Signature:**
```
db.{table}.nextId() -> int
```

**Return Type:** `int`

**Example:**
```glyph
$ newId = db.users.nextId()
$ user = db.users.create({
  id: newId,
  username: input.username
})
```

---

### db.{table}.deleteWhere

Deletes all records matching a field value.

**Signature:**
```
db.{table}.deleteWhere(field: str, value: any) -> int
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| field | str | The field name to match |
| value | any | The value to match |

**Return Type:** `int` (number of deleted records)

**Example:**
```glyph
$ deletedCount = db.users.deleteWhere("active", false)
> {deleted: deletedCount}
```

---

## WebSocket API

WebSocket functionality is available in WebSocket route handlers using the `ws` object.

### ws.send

Sends a message to the current client.

**Signature:**
```
ws.send(message: object | str)
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| message | object or str | The message to send |

**Example:**
```glyph
@ ws /chat {
  on connect {
    ws.send({
      type: "system",
      message: "Welcome to the chat!"
    })
  }
}
```

---

### ws.broadcast

Broadcasts a message to all connected clients.

**Signature:**
```
ws.broadcast(message: object | str)
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| message | object or str | The message to broadcast |

**Example:**
```glyph
@ ws /chat {
  on message {
    ws.broadcast(input)
  }
}
```

---

### ws.join

Adds the current client to a room.

**Signature:**
```
ws.join(room: str)
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| room | str | The room name to join |

**Example:**
```glyph
@ ws /chat/:room {
  on connect {
    ws.join(room)
    ws.send({message: "You joined " + room})
  }
}
```

---

### ws.leave

Removes the current client from a room.

**Signature:**
```
ws.leave(room: str)
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| room | str | The room name to leave |

**Example:**
```glyph
@ ws /chat/:room {
  on disconnect {
    ws.leave(room)
  }
}
```

---

### ws.broadcast_to_room

Broadcasts a message to all clients in a specific room.

**Signature:**
```
ws.broadcast_to_room(room: str, message: object | str)
```

**Parameters:**
| Name | Type | Description |
|------|------|-------------|
| room | str | The room to broadcast to |
| message | object or str | The message to broadcast |

**Example:**
```glyph
@ ws /chat/:room {
  on message {
    ws.broadcast_to_room(room, {
      type: "chat",
      text: input.text,
      sender: input.username
    })
  }
}
```

---

### ws.get_rooms

Returns a list of all active rooms.

**Signature:**
```
ws.get_rooms() -> List[str]
```

**Return Type:** `List[str]`

**Example:**
```glyph
@ route /api/ws-status [GET]
  > {
    rooms: ws.get_rooms(),
    room_count: ws.get_room_count()
  }
```

---

## Request/Response Objects

### Input Object

The `input` object contains the request body for POST/PUT/PATCH requests and is automatically available in route handlers.

**Properties:**
| Property | Type | Description |
|----------|------|-------------|
| input.{field} | any | Access any field from the request body |

**Example:**
```glyph
@ route /api/users [POST]
  $ username = input.username
  $ email = input.email
  $ age = input.age

  > {received: {username: username, email: email}}
```

---

### Query Object

The `query` object contains URL query parameters.

**Properties:**
| Property | Type | Description |
|----------|------|-------------|
| query.{param} | str or int | Access any query parameter |

**Example:**
```glyph
@ route /api/search [GET]
  $ searchTerm = query.q
  $ page = query.page
  $ limit = query.limit

  > {search: searchTerm, page: page}
```

---

### Path Parameters

Path parameters are defined in the route path with `:` prefix and are available as variables.

**Example:**
```glyph
@ route /api/users/:id [GET]
  # 'id' is automatically available as a variable
  $ user = db.users.get(id)
  > user

@ route /api/users/:userId/posts/:postId [GET]
  # Both 'userId' and 'postId' are available
  $ post = db.posts.get(postId)
  > post
```

---

### Auth Object

When using JWT authentication, the `auth` object contains the decoded token data.

**Properties:**
| Property | Type | Description |
|----------|------|-------------|
| auth.user.id | int | The authenticated user's ID |
| auth.user.username | str | The authenticated user's username |
| auth.user.role | str | The authenticated user's role |
| auth.token | str | The raw JWT token |
| auth.expiresAt | int | Token expiration timestamp |

**Example:**
```glyph
@ route /api/profile [GET]
  + auth(jwt)

  $ userId = auth.user.id
  $ user = db.users.get(userId)

  > user
```

---

### HTTP Status Codes

Routes return HTTP 200 by default. To return error responses, structure your response appropriately:

**Example:**
```glyph
@ route /api/users/:id [GET]
  % db: Database

  $ user = db.users.get(id)

  if user == null {
    # This indicates a 404 scenario
    > {
      success: false,
      message: "User not found",
      code: "NOT_FOUND"
    }
  } else {
    > {
      success: true,
      data: user
    }
  }
```

---

## Collection Operations

### Array Concatenation

Arrays can be concatenated using the `+` operator.

**Example:**
```glyph
$ arr1 = [1, 2, 3]
$ arr2 = [4, 5, 6]
$ combined = arr1 + arr2  # [1, 2, 3, 4, 5, 6]

# Adding single element
$ items = []
$ items = items + [newItem]
```

---

### Array Length (Method)

Arrays have a `length()` method.

**Signature:**
```
array.length() -> int
```

**Example:**
```glyph
$ items = [1, 2, 3, 4, 5]
$ count = items.length()  # Returns 5
```

---

### For Loop Iteration

Iterate over arrays and objects using for loops.

**Array Iteration:**
```glyph
$ numbers = [1, 2, 3, 4, 5]
$ sum = 0

for num in numbers {
  $ sum = sum + num
}

# With index
for index, num in numbers {
  > {index: index, value: num}
}
```

**Object Iteration:**
```glyph
$ config = {host: "localhost", port: "8080"}

for key, value in config {
  > {key: key, value: value}
}
```

---

### Array Indexing

Access array elements by index (0-based).

**Example:**
```glyph
$ fruits = ["apple", "banana", "cherry"]
$ first = fruits[0]   # "apple"
$ second = fruits[1]  # "banana"
```

---

## Operators

### Arithmetic Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Addition (or string concatenation) | `a + b` |
| `-` | Subtraction | `a - b` |
| `*` | Multiplication | `a * b` |
| `/` | Division | `a / b` |

**Example:**
```glyph
$ a = 10
$ b = 3

$ sum = a + b      # 13
$ diff = a - b     # 7
$ product = a * b  # 30
$ quotient = a / b # 3 (integer division)
```

---

### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equal to | `a == b` |
| `!=` | Not equal to | `a != b` |
| `<` | Less than | `a < b` |
| `<=` | Less than or equal | `a <= b` |
| `>` | Greater than | `a > b` |
| `>=` | Greater than or equal | `a >= b` |

**Example:**
```glyph
$ age = 25

if age >= 18 {
  > {status: "adult"}
}

if age != 21 {
  > {canDrinkUS: false}
}
```

---

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `&&` | Logical AND | `a && b` |
| `\|\|` | Logical OR | `a \|\| b` |
| `!` | Logical NOT | `!a` |

**Example:**
```glyph
$ isActive = true
$ hasPermission = true

if isActive && hasPermission {
  > {access: "granted"}
}

if isActive || hasPermission {
  > {partialAccess: true}
}

if !isActive {
  > {status: "inactive"}
}
```

---

### Unary Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `-` | Negation | `-value` |
| `!` | Logical NOT | `!condition` |

**Example:**
```glyph
$ value = 42
$ negative = -value  # -42

$ isValid = false
$ isInvalid = !isValid  # true
```

---

### String Concatenation

Strings can be concatenated using the `+` operator.

**Example:**
```glyph
$ firstName = "John"
$ lastName = "Doe"
$ fullName = firstName + " " + lastName  # "John Doe"

$ greeting = "Hello, " + firstName + "!"  # "Hello, John!"
```

---

## Control Flow

### If/Else Statements

**Syntax:**
```glyph
if condition {
  # code if true
} else {
  # code if false
}
```

**Example:**
```glyph
if user.age >= 18 {
  $ status = "adult"
} else {
  $ status = "minor"
}

# Nested if
if score >= 90 {
  $ grade = "A"
} else {
  if score >= 80 {
    $ grade = "B"
  } else {
    $ grade = "C"
  }
}
```

---

### While Loops

**Syntax:**
```glyph
while condition {
  # loop body
}
```

**Example:**
```glyph
$ i = 0
$ sum = 0

while i < 10 {
  $ sum = sum + i
  $ i = i + 1
}

> {sum: sum}  # 45
```

---

### For Loops

**Syntax:**
```glyph
# Simple iteration
for item in collection {
  # process item
}

# With index/key
for index, item in collection {
  # process with index
}
```

**Example:**
```glyph
$ items = [{price: 10}, {price: 20}, {price: 30}]
$ total = 0

for item in items {
  $ total = total + item.price
}

> {total: total}  # 60
```

---

### Switch Statements

**Syntax:**
```glyph
switch value {
  case "option1" {
    # handle option1
  }
  case "option2" {
    # handle option2
  }
  default {
    # handle default case
  }
}
```

**Example:**
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

> {message: message}
```

---

## Route Modifiers

### Authentication

```glyph
@ route /api/protected
  + auth(jwt)              # Require valid JWT
  + auth(jwt, role: admin) # Require admin role
```

### Rate Limiting

```glyph
@ route /api/search
  + ratelimit(10/sec)    # 10 requests per second
  + ratelimit(100/min)   # 100 requests per minute
  + ratelimit(1000/hour) # 1000 requests per hour
```

### Database Injection

```glyph
@ route /api/users
  % db: Database

  $ users = db.users.all()
```

---

## Cron Task Syntax

Scheduled tasks use the `*` symbol with cron expressions.

**Syntax:**
```glyph
* "cron_expression" task_name {
  # task body
}
```

**Cron Expression Format:**
```
* * * * *
| | | | |
| | | | +-- Day of week (0-7, Sunday = 0 or 7)
| | | +---- Month (1-12)
| | +------ Day of month (1-31)
| +-------- Hour (0-23)
+---------- Minute (0-59)
```

**Common Patterns:**
| Pattern | Description |
|---------|-------------|
| `0 0 * * *` | Daily at midnight |
| `0 * * * *` | Every hour |
| `*/5 * * * *` | Every 5 minutes |
| `0 9 * * 0` | Sundays at 9am |
| `0 0 1 * *` | First of every month |

---

## Event Handler Syntax

Event handlers use the `~` symbol.

**Syntax:**
```glyph
~ "event.name" {
  $ eventData = event.field
  # handle event
}

# Async handler
~ "event.name" async {
  # handle asynchronously
}
```

**Built-in Event Data:**
- `event` - The event payload object

---

## Queue Worker Syntax

Queue workers use the `&` symbol.

**Syntax:**
```glyph
& "queue.name" {
  + concurrency(5)  # Number of concurrent workers
  + retries(3)      # Retry attempts
  + timeout(30)     # Timeout in seconds

  $ data = message.field
  # process message
}
```

**Built-in Message Data:**
- `message` - The queue message object

---

## CLI Command Syntax

CLI commands use the `!` symbol.

**Syntax:**
```glyph
! command_name arg1: type! arg2: type = default {
  # command body
}

# With description
! command_name "Description text" {
  # command body
}
```

**Example:**
```glyph
! greet name: str! --formal: bool = false {
  if formal {
    $ msg = "Good day, " + name
  } else {
    $ msg = "Hey " + name + "!"
  }
  > {greeting: msg}
}
```

**Execute with:**
```bash
glyph exec main.glyph greet --name="Alice" --formal
```

---

## Type Definitions

### Basic Types

| Type | Description |
|------|-------------|
| `int` | Integer number |
| `int!` | Required integer |
| `float` | Floating-point number |
| `str` | String |
| `str!` | Required string |
| `bool` | Boolean (true/false) |
| `timestamp` | Unix timestamp |
| `any` | Any type |

### Collection Types

| Type | Description |
|------|-------------|
| `List[T]` | Array of type T |
| `Set[T]` | Set of type T |
| `Map[K, V]` | Map with key K and value V |

### Type Definition Syntax

```glyph
: TypeName {
  field1: type!     # Required field
  field2: type      # Optional field
  field3: List[str] # Collection field
}
```

---

## See Also

- [Language Guide](language-guide.md) - Complete language syntax guide
- [CLI Reference](CLI.md) - Command-line interface documentation
- [Architecture](ARCHITECTURE_DESIGN.md) - System architecture overview
