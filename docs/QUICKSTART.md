# GlyphLang Quick Start Guide

GlyphLang is the AI-first language for REST APIs. Its symbol-based syntax uses 45% fewer tokens than Python, making it ideal for AI code generation and agent workflows.

This guide will help you get started with GlyphLang in under 45 minutes. You will learn how to install GlyphLang, create your first API, add authentication, and connect to a database.

## Table of Contents

1. [Installation](#installation)
2. [Hello World (5 minutes)](#hello-world-5-minutes)
3. [Building a REST API (15 minutes)](#building-a-rest-api-15-minutes)
4. [Adding Authentication (10 minutes)](#adding-authentication-10-minutes)
5. [Database Integration (10 minutes)](#database-integration-10-minutes)
6. [AI Agent Integration](#ai-agent-integration)
7. [Next Steps](#next-steps)

---

## Installation

### Building from Source

GlyphLang is built with Go. Make sure you have Go 1.21 or later installed.

```bash
# Clone the repository
git clone https://github.com/glyphlang/glyphlang.git
cd glyphlang

# Build the CLI
go build -o glyph ./cmd/glyph

# Or use make
make build

# Install to your system (optional)
make install
```

After building, verify the installation:

```bash
glyph --version
# GlyphLang v0.1.6
```

### Binary Downloads

Pre-built binaries are available from the [releases page](https://github.com/GlyphLang/GlyphLang/releases/latest):

| Platform | Download |
|----------|----------|
| Windows (Installer) | [glyph-windows-setup.exe](https://github.com/GlyphLang/GlyphLang/releases/latest/download/glyph-windows-setup.exe) |
| Windows (ZIP) | [glyph-windows-amd64.zip](https://github.com/GlyphLang/GlyphLang/releases/latest/download/glyph-windows-amd64.zip) |
| Linux (amd64) | [glyph-linux-amd64.zip](https://github.com/GlyphLang/GlyphLang/releases/latest/download/glyph-linux-amd64.zip) |
| macOS (Intel) | [glyph-darwin-amd64.zip](https://github.com/GlyphLang/GlyphLang/releases/latest/download/glyph-darwin-amd64.zip) |
| macOS (Apple Silicon) | [glyph-darwin-arm64.zip](https://github.com/GlyphLang/GlyphLang/releases/latest/download/glyph-darwin-arm64.zip) |

Download the appropriate binary for your platform and add it to your PATH.

### VS Code Extension

For syntax highlighting and language support in VS Code:

1. Open VS Code
2. Go to Extensions (Ctrl+Shift+X)
3. Search for "GlyphLang"
4. Click Install

Or install from the repository: https://github.com/GlyphLang/vscode-glyph

The extension provides:
- Syntax highlighting
- Code completion
- Hover information
- Go to definition
- Error diagnostics

---

## Hello World (5 minutes)

Let us create your first GlyphLang application.

### Step 1: Create a Project

```bash
# Create a new project directory
glyph init my-first-api -t hello-world

# Navigate to the project
cd my-first-api
```

This creates a `main.glyph` file with a basic hello world template.

### Step 2: Write Your First Route

Open `main.glyph` and replace its contents with:

```glyph
# My First GlyphLang API
# A simple hello world example

@ route / {
  > {
    message: "Hello, World!",
    status: "ok"
  }
}

@ route /hello/:name {
  $ greeting = "Hello, " + name + "!"
  > {message: greeting}
}

@ route /health {
  > {status: "healthy"}
}
```

### Understanding the Syntax

- `#` starts a comment
- `@ route /path` defines an HTTP endpoint
- `>` returns a JSON response
- `$` declares a variable
- `:name` is a path parameter

### Step 3: Run the Development Server

```bash
glyph dev main.glyph
```

You will see output like:

```
[INFO] Starting development server on port 3000...
[SUCCESS] Dev server listening on http://localhost:3000 (compiled mode)
[INFO] Watching main.glyph for changes...
[INFO] Press Ctrl+C to stop
```

### Step 4: Test Your API

Open a new terminal and test with curl:

```bash
# Test the root endpoint
curl http://localhost:3000/
# {"message":"Hello, World!","status":"ok"}

# Test with a path parameter
curl http://localhost:3000/hello/Alice
# {"message":"Hello, Alice!"}

# Test the health endpoint
curl http://localhost:3000/health
# {"status":"healthy"}
```

### Hot Reload

The development server watches for file changes. Edit `main.glyph` and save. The server automatically reloads:

```
[WARNING] File changed, reloading...
[SUCCESS] Hot reload complete (45ms)
```

---

## Building a REST API (15 minutes)

Now let us build a complete Todo API with CRUD operations.

### Step 1: Define Data Types

Create a new file called `main.glyph`:

```glyph
# Todo API
# A complete REST API example

# Define the Todo type
: Todo {
  id: int!
  title: str!
  description: str
  completed: bool!
  priority: str
  created_at: timestamp
}

# Define response types
: TodoResponse {
  success: bool!
  message: str!
  todo: Todo
}

: TodoListResponse {
  success: bool!
  todos: List[Todo]
  total: int!
}
```

### Understanding Types

- `:` defines a new type
- `!` marks a field as required (non-nullable)
- Fields without `!` are optional
- `List[Todo]` is a collection of Todo items

### Step 2: Create CRUD Routes

Add the following routes to your file:

```glyph
# Get all todos
@ route /api/todos [GET] -> TodoListResponse {
  + ratelimit(100/min)
  % db: Database

  $ todos = db.todos.all()
  $ total = todos.length()

  > {
    success: true,
    todos: todos,
    total: total
  }
}

# Get a single todo by ID
@ route /api/todos/:id [GET] -> TodoResponse {
  + ratelimit(200/min)
  % db: Database

  $ todo = db.todos.get(id)

  if todo == null {
    > {
      success: false,
      message: "Todo not found",
      todo: null
    }
  } else {
    > {
      success: true,
      message: "Todo found",
      todo: todo
    }
  }
}

# Create a new todo
@ route /api/todos [POST] -> TodoResponse {
  + ratelimit(50/min)
  % db: Database

  if input.title == null || input.title == "" {
    > {
      success: false,
      message: "Title is required",
      todo: null
    }
  } else {
    $ newTodo = {
      id: db.todos.nextId(),
      title: input.title,
      description: input.description,
      completed: false,
      priority: "medium",
      created_at: now()
    }

    $ saved = db.todos.create(newTodo)

    > {
      success: true,
      message: "Todo created successfully",
      todo: saved
    }
  }
}

# Update a todo
@ route /api/todos/:id [PUT] -> TodoResponse {
  + ratelimit(50/min)
  % db: Database

  $ existingTodo = db.todos.get(id)

  if existingTodo == null {
    > {
      success: false,
      message: "Todo not found",
      todo: null
    }
  } else {
    if input.title != null {
      $ existingTodo.title = input.title
    }

    if input.description != null {
      $ existingTodo.description = input.description
    }

    if input.completed != null {
      $ existingTodo.completed = input.completed
    }

    $ updated = db.todos.update(id, existingTodo)

    > {
      success: true,
      message: "Todo updated successfully",
      todo: updated
    }
  }
}

# Delete a todo
@ route /api/todos/:id [DELETE] -> TodoResponse {
  + ratelimit(50/min)
  % db: Database

  $ todo = db.todos.get(id)

  if todo == null {
    > {
      success: false,
      message: "Todo not found",
      todo: null
    }
  } else {
    $ result = db.todos.delete(id)

    > {
      success: true,
      message: "Todo deleted successfully",
      todo: todo
    }
  }
}
```

### Understanding Route Syntax

- `@ route /path [METHOD]` defines an endpoint with HTTP method
- `-> Type` specifies the return type
- `+ ratelimit(100/min)` adds rate limiting middleware
- `% db: Database` injects the database service
- `input` contains the request body (for POST/PUT)
- `:id` is a path parameter accessible as `id`

### Step 3: Test the API

Start the server:

```bash
glyph dev main.glyph
```

Test the endpoints:

```bash
# Create a todo
curl -X POST http://localhost:3000/api/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn GlyphLang", "description": "Complete the quickstart guide"}'

# Get all todos
curl http://localhost:3000/api/todos

# Get a specific todo
curl http://localhost:3000/api/todos/1

# Update a todo
curl -X PUT http://localhost:3000/api/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"completed": true}'

# Delete a todo
curl -X DELETE http://localhost:3000/api/todos/1
```

---

## Adding Authentication (10 minutes)

Let us protect our API with JWT authentication.

### Step 1: Add Auth Types

Add these types to your file:

```glyph
# Authentication types
: LoginRequest {
  username: str!
  password: str!
}

: AuthResponse {
  success: bool!
  message: str!
  token: str
}
```

### Step 2: Create Login Endpoint

```glyph
# Login endpoint (public)
@ route /api/auth/login [POST] -> AuthResponse {
  + ratelimit(20/min)
  % db: Database

  if input.username == null || input.username == "" {
    > {
      success: false,
      message: "Username is required",
      token: null
    }
  } else {
    if input.password == null || input.password == "" {
      > {
        success: false,
        message: "Password is required",
        token: null
      }
    } else {
      $ user = db.users.findOne("username", input.username)

      if user == null {
        > {
          success: false,
          message: "Invalid username or password",
          token: null
        }
      } else {
        $ passwordValid = crypto.verify(input.password, user.password)

        if passwordValid == false {
          > {
            success: false,
            message: "Invalid username or password",
            token: null
          }
        } else {
          $ token = jwt.sign({
            user_id: user.id,
            username: user.username,
            role: user.role
          }, "7d")

          > {
            success: true,
            message: "Login successful",
            token: token
          }
        }
      }
    }
  }
}
```

### Step 3: Protect Routes with JWT

Add the `+ auth(jwt)` middleware to protect routes:

```glyph
# Protected route - requires valid JWT token
@ route /api/todos [POST] -> TodoResponse {
  + auth(jwt)
  + ratelimit(50/min)
  % db: Database

  # ... rest of the handler
  # The authenticated user is available via auth.user
  $ userId = auth.user.id
}
```

### Step 4: Role-Based Access Control

For admin-only routes, specify the required role:

```glyph
# Admin only - delete any todo
@ route /api/admin/todos/:id [DELETE] -> TodoResponse {
  + auth(jwt, role: admin)
  + ratelimit(10/min)
  % db: Database

  $ todo = db.todos.get(id)

  if todo == null {
    > {
      success: false,
      message: "Todo not found",
      todo: null
    }
  } else {
    $ result = db.todos.delete(id)

    > {
      success: true,
      message: "Todo deleted by admin",
      todo: todo
    }
  }
}
```

### Step 5: Handle Unauthorized Access

When authentication fails, GlyphLang automatically returns a 401 response:

```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired token"
}
```

### Testing Protected Routes

```bash
# Login to get a token
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "secret123"}'

# Use the token in subsequent requests
curl http://localhost:3000/api/todos \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

---

## Database Integration (10 minutes)

GlyphLang provides built-in database support with PostgreSQL.

### Step 1: Configure Database Connection

Database configuration is handled through environment variables or a config file:

```bash
# Set environment variables
export DATABASE_URL="postgres://user:password@localhost:5432/mydb"

# Or use a .env file
echo 'DATABASE_URL=postgres://user:password@localhost:5432/mydb' > .env
```

### Step 2: Define Your Data Model

```glyph
# User model
: User {
  id: int!
  username: str!
  email: str!
  full_name: str
  active: bool!
  created_at: timestamp
  updated_at: timestamp
}
```

### Step 3: Basic CRUD Operations

```glyph
# Inject database service
% db: Database

# Fetch all records
$ users = db.users.all()

# Get by ID
$ user = db.users.get(id)

# Create new record
$ newUser = db.users.create({
  username: "alice",
  email: "alice@example.com",
  full_name: "Alice Smith",
  active: true,
  created_at: now(),
  updated_at: now()
})

# Update record
$ updated = db.users.update(id, {
  full_name: "Alice Johnson",
  updated_at: now()
})

# Delete record
$ result = db.users.delete(id)
```

### Step 4: Filtering and Querying

```glyph
# Filter by field value
$ activeUsers = db.users.filter("active", true)

# Find one record
$ user = db.users.findOne("username", "alice")

# Count records
$ totalUsers = db.users.count()
$ activeCount = db.users.count("active", true)

# Get next ID for new records
$ nextId = db.users.nextId()
```

### Complete Database Example

```glyph
# User API with database integration

: User {
  id: int!
  username: str!
  email: str!
  active: bool!
  created_at: timestamp
}

: UserResponse {
  success: bool!
  message: str!
  user: User
}

: UserListResponse {
  success: bool!
  users: List[User]
  total: int!
}

# Get all users
@ route /api/users [GET] -> UserListResponse {
  + ratelimit(100/min)
  % db: Database

  $ users = db.users.all()
  $ total = users.length()

  > {
    success: true,
    users: users,
    total: total
  }
}

# Get active users only
@ route /api/users/active [GET] -> UserListResponse {
  + ratelimit(100/min)
  % db: Database

  $ activeUsers = db.users.filter("active", true)
  $ total = activeUsers.length()

  > {
    success: true,
    users: activeUsers,
    total: total
  }
}

# Search users by username
@ route /api/users/search/:query [GET] -> UserListResponse {
  + ratelimit(100/min)
  % db: Database

  $ allUsers = db.users.all()
  $ results = []

  for user in allUsers {
    if user.username.contains(query) {
      $ results = results + [user]
    }
  }

  > {
    success: true,
    users: results,
    total: results.length()
  }
}

# Create a new user
@ route /api/users [POST] -> UserResponse {
  + auth(jwt)
  + ratelimit(50/min)
  % db: Database

  if input.username == null || input.username == "" {
    > {
      success: false,
      message: "Username is required",
      user: null
    }
  } else {
    $ existing = db.users.findOne("username", input.username)

    if existing != null {
      > {
        success: false,
        message: "Username already exists",
        user: null
      }
    } else {
      $ newUser = {
        id: db.users.nextId(),
        username: input.username,
        email: input.email,
        active: true,
        created_at: now()
      }

      $ saved = db.users.create(newUser)

      > {
        success: true,
        message: "User created successfully",
        user: saved
      }
    }
  }
}
```

---

## AI Agent Integration

GlyphLang includes built-in commands for AI coding assistants and agents.

### Generate Project Context

The `context` command creates a compact summary of your project that fits in LLM context windows:

```bash
# Generate full project context as JSON
glyph context

# Generate minimal text output (fewer tokens)
glyph context --format compact

# Focus on specific aspects
glyph context --for route    # Routes only
glyph context --for type     # Type definitions only
glyph context --for function # Functions only

# Show only changes since last run
glyph context --changed
```

### Validate with AI-Friendly Output

The `validate` command provides structured errors that AI agents can parse and fix:

```bash
# Validate with structured JSON output
glyph validate main.glyph --ai

# Validate entire directory
glyph validate src/ --ai
```

Example output:

```json
{
  "valid": false,
  "errors": [
    {
      "type": "undefined_reference",
      "message": "undefined variable: userId",
      "file": "main.glyph",
      "line": 15,
      "column": 10,
      "hint": "Did you mean 'user_id'?"
    }
  ]
}
```

### AI Agent Workflow

A typical AI coding workflow with GlyphLang:

```bash
# 1. Agent gets project context
glyph context --format compact > context.txt

# 2. Agent makes changes to .glyph files

# 3. Agent validates changes
glyph validate src/ --ai

# 4. If errors, agent fixes them and re-validates

# 5. Agent checks what changed
glyph context --changed
```

### Why This Matters

- **45% fewer tokens** than Python for equivalent code
- **Structured errors** agents can parse and fix automatically
- **Compact context** fits more project info in context windows
- **Consistent syntax** reduces hallucination errors

---

## Next Steps

Congratulations! You have learned the basics of GlyphLang. Here are some resources to continue your journey:

### Documentation

- [Language Guide](language-guide.md) - Complete syntax reference and language features
- [CLI Reference](CLI.md) - All CLI commands and options
- [Architecture](ARCHITECTURE_DESIGN.md) - Internal architecture and design decisions

### Examples

Explore the `examples/` directory for more complete examples:

- `examples/todo-api/` - Full Todo application with CRUD
- `examples/blog-api/` - Blog with posts and comments
- `examples/auth-demo/` - Complete authentication system
- `examples/database-demo/` - Advanced database operations
- `examples/websocket-chat/` - Real-time WebSocket chat
- `examples/cron-demo/` - Scheduled tasks
- `examples/queue-demo/` - Background job processing

### Key Concepts to Explore

1. **Middleware** - Rate limiting, caching, logging
2. **WebSockets** - Real-time communication
3. **Scheduled Tasks** - Cron-based automation
4. **Event Handlers** - Event-driven architecture
5. **Queue Workers** - Background job processing
6. **CLI Commands** - Build CLI tools with GlyphLang

### Running Examples

```bash
# Run any example
glyph dev examples/todo-api/main.glyph

# Run with a specific port
glyph dev examples/blog-api/main.glyph -p 8080

# Open browser automatically
glyph dev examples/hello-world/main.glyph --open
```

### Production Deployment

```bash
# Compile for production
glyph compile main.glyph -o build/app.glybc -O 3

# Run compiled bytecode
glyph run build/app.glybc -p 80
```

### Getting Help

```bash
# View all commands
glyph --help

# Get help for a specific command
glyph dev --help
glyph compile --help
```

---

Happy coding with GlyphLang!
