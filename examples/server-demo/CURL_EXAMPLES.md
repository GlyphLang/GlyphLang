# Glyph Server - curl Examples

Quick reference for testing the Glyph HTTP server with curl commands.

## Prerequisites

Make sure the server is running:
```bash
go run examples/server-demo/main.go
```

The server will start on `http://localhost:8080`

## Basic Requests

### GET /hello
Simple hello world endpoint:
```bash
curl http://localhost:8080/hello
```

Response:
```json
{
  "message": "Hello, World!",
  "timestamp": 1701234567
}
```

### GET /greet/:name
Greeting with path parameter:
```bash
curl http://localhost:8080/greet/Alice
```

Response:
```json
{
  "text": "Hello, Alice!",
  "timestamp": 1701234567
}
```

### GET /health
Health check endpoint:
```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok",
  "timestamp": 1701234567
}
```

## REST API - Users

### GET /api/users
Get all users:
```bash
curl http://localhost:8080/api/users
```

With query parameters:
```bash
curl "http://localhost:8080/api/users?page=2&limit=10"
```

Response:
```json
[
  {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  },
  {
    "id": 2,
    "name": "Jane Smith",
    "email": "jane@example.com"
  }
]
```

### GET /api/users/:id
Get specific user:
```bash
curl http://localhost:8080/api/users/42
```

Response:
```json
{
  "id": "42",
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": 1701234567
}
```

### POST /api/users
Create new user:
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com"
  }'
```

Response (201 Created):
```json
{
  "id": 123,
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": 1701234567
}
```

### PUT /api/users/:id
Update user (full replacement):
```bash
curl -X PUT http://localhost:8080/api/users/123 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com"
  }'
```

Response:
```json
{
  "id": "123",
  "name": "Jane Doe",
  "email": "jane@example.com",
  "updated_at": 1701234567
}
```

### PATCH /api/users/:id
Partial update user:
```bash
curl -X PATCH http://localhost:8080/api/users/123 \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newemail@example.com"
  }'
```

Response:
```json
{
  "id": "123",
  "updated": {
    "email": "newemail@example.com"
  }
}
```

### DELETE /api/users/:id
Delete user:
```bash
curl -X DELETE http://localhost:8080/api/users/123
```

Response:
```json
{
  "success": true,
  "message": "User deleted"
}
```

## Nested Resources

### GET /api/users/:userId/posts/:postId
Access nested resource:
```bash
curl http://localhost:8080/api/users/42/posts/99
```

Response:
```json
{
  "id": "99",
  "userId": "42",
  "title": "Sample Post",
  "content": "This is a sample post content"
}
```

## Error Handling

### 404 Not Found
Request non-existent route:
```bash
curl http://localhost:8080/invalid
```

Response (404):
```json
{
  "error": true,
  "message": "route not found",
  "code": 404,
  "details": "no route matches path /invalid"
}
```

### 400 Bad Request
Send invalid JSON:
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{invalid json}'
```

Response (400):
```json
{
  "error": true,
  "message": "invalid JSON body",
  "code": 400,
  "details": "failed to parse JSON: ..."
}
```

Missing required fields:
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{}'
```

Response (400):
```json
{
  "error": true,
  "message": "name and email are required",
  "code": 400
}
```

## Advanced Usage

### Pretty Print with jq
Install jq for formatted JSON output:
```bash
curl -s http://localhost:8080/api/users | jq .
```

### Include Response Headers
```bash
curl -i http://localhost:8080/hello
```

### Verbose Output
```bash
curl -v http://localhost:8080/hello
```

### Save Response to File
```bash
curl -o response.json http://localhost:8080/api/users
```

### Custom Headers
```bash
curl -H "Authorization: Bearer token123" \
  http://localhost:8080/api/users
```

### Multiple Headers
```bash
curl -H "Authorization: Bearer token123" \
     -H "X-Custom-Header: value" \
     http://localhost:8080/api/users
```

### Timing Information
```bash
curl -w "\nTotal time: %{time_total}s\n" \
  http://localhost:8080/api/users
```

## Testing Scripts

### Unix/Linux/Mac
```bash
chmod +x examples/server-demo/test-endpoints.sh
./examples/server-demo/test-endpoints.sh
```

### Windows
```batch
examples\server-demo\test-endpoints.bat
```

## Performance Testing

### Simple load test with curl
```bash
for i in {1..100}; do
  curl -s http://localhost:8080/hello > /dev/null
done
```

### Parallel requests
```bash
for i in {1..10}; do
  curl -s http://localhost:8080/hello > /dev/null &
done
wait
```

## Notes

- All POST/PUT/PATCH requests must include `Content-Type: application/json` header
- Path parameters are specified with `:paramName` in the route definition
- Query parameters are appended with `?key=value&key2=value2`
- The server automatically parses JSON request bodies
- All responses are in JSON format
- Use `-s` flag with curl for silent mode (no progress bar)
- Use `-X METHOD` to specify HTTP method (defaults to GET)
