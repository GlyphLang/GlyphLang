# Todo List API

A comprehensive todo list API demonstrating CRUD operations in Glyph.

## Features

- Complete CRUD operations (Create, Read, Update, Delete)
- Path parameters for resource identification
- Multiple HTTP methods (GET, POST, PUT, DELETE, PATCH)
- Input validation
- Error handling
- Priority-based filtering
- Rate limiting

## API Endpoints

### Get All Todos
```
GET /api/todos
```
Returns a list of all todos with statistics.

**Response:**
```json
{
  "todos": [...],
  "total": 10,
  "completed": 5,
  "pending": 5
}
```

### Get Todo by ID
```
GET /api/todos/:id
```
Returns a specific todo by its ID.

**Response:**
```json
{
  "id": 1,
  "title": "Complete project",
  "description": "Finish the Glyph implementation",
  "completed": false,
  "priority": "high",
  "created_at": 1234567890,
  "updated_at": 1234567890
}
```

### Create Todo
```
POST /api/todos
```
Creates a new todo item.

**Request Body:**
```json
{
  "title": "New task",
  "description": "Task description",
  "priority": "medium"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Todo created successfully",
  "todo": {...}
}
```

### Update Todo
```
PUT /api/todos/:id
```
Updates an existing todo.

**Request Body:**
```json
{
  "title": "Updated title",
  "description": "Updated description",
  "completed": true,
  "priority": "low"
}
```

### Delete Todo
```
DELETE /api/todos/:id
```
Deletes a todo by ID.

**Response:**
```json
{
  "success": true,
  "message": "Todo deleted successfully",
  "todo": {...}
}
```

### Mark as Completed
```
PATCH /api/todos/:id/complete
```
Marks a todo as completed.

### Get Todos by Priority
```
GET /api/todos/priority/:priority
```
Filters todos by priority level (low, medium, high).

## Type Definitions

### Todo
```
: Todo {
  id: int!
  title: str!
  description: str
  completed: bool!
  priority: str
  created_at: timestamp
  updated_at: timestamp
}
```

### TodoList
```
: TodoList {
  todos: List[Todo]
  total: int!
  completed: int!
  pending: int!
}
```

### TodoResponse
```
: TodoResponse {
  success: bool!
  message: str!
  todo: Todo
}
```

## Priority Levels
- `low` - Low priority tasks
- `medium` - Medium priority tasks (default)
- `high` - High priority tasks

## Error Handling

The API returns error responses in this format:
```json
{
  "success": false,
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

Common error codes:
- `NOT_FOUND` - Todo not found
- `VALIDATION_ERROR` - Invalid input data

## Rate Limiting

- GET endpoints: 100-200 requests/minute
- POST/PUT/DELETE endpoints: 50 requests/minute
- PATCH endpoints: 100 requests/minute
