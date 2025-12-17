@echo off
REM Test script for Glyph Demo Server (Windows)
REM Make sure the server is running before executing this script

echo Testing Glyph Demo Server Endpoints
echo ====================================
echo.

set BASE_URL=http://localhost:8080

echo 1. GET /hello - Simple hello world
curl -s %BASE_URL%/hello
echo.
echo ---
echo.

echo 2. GET /greet/:name - Greeting with name parameter
curl -s %BASE_URL%/greet/Alice
echo.
echo ---
echo.

echo 3. GET /health - Health check endpoint
curl -s %BASE_URL%/health
echo.
echo ---
echo.

echo 4. GET /api/users - Get all users
curl -s %BASE_URL%/api/users
echo.
echo ---
echo.

echo 5. GET /api/users?page=2^&limit=10 - Get users with pagination
curl -s "%BASE_URL%/api/users?page=2&limit=10"
echo.
echo ---
echo.

echo 6. GET /api/users/:id - Get user by ID
curl -s %BASE_URL%/api/users/42
echo.
echo ---
echo.

echo 7. POST /api/users - Create new user
curl -s -X POST %BASE_URL%/api/users -H "Content-Type: application/json" -d "{\"name\": \"John Doe\", \"email\": \"john@example.com\"}"
echo.
echo ---
echo.

echo 8. PUT /api/users/:id - Update user
curl -s -X PUT %BASE_URL%/api/users/123 -H "Content-Type: application/json" -d "{\"name\": \"Jane Doe\", \"email\": \"jane@example.com\"}"
echo.
echo ---
echo.

echo 9. PATCH /api/users/:id - Partial update user
curl -s -X PATCH %BASE_URL%/api/users/123 -H "Content-Type: application/json" -d "{\"email\": \"newemail@example.com\"}"
echo.
echo ---
echo.

echo 10. DELETE /api/users/:id - Delete user
curl -s -X DELETE %BASE_URL%/api/users/123
echo.
echo ---
echo.

echo 11. GET /api/users/:userId/posts/:postId - Nested resources
curl -s %BASE_URL%/api/users/42/posts/99
echo.
echo ---
echo.

echo 12. GET /invalid - Test 404 error
curl -s %BASE_URL%/invalid
echo.
echo ---
echo.

echo 13. POST /api/users - Test invalid JSON error
curl -s -X POST %BASE_URL%/api/users -H "Content-Type: application/json" -d "{invalid json}"
echo.
echo ---
echo.

echo.
echo Testing complete!
pause
