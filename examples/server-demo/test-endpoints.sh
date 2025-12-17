#!/bin/bash
# Test script for Glyph Demo Server
# Make sure the server is running before executing this script

echo "Testing Glyph Demo Server Endpoints"
echo "===================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

# Helper function
test_endpoint() {
    echo -e "${BLUE}$1${NC}"
    echo "Command: $2"
    echo "Response:"
    eval $2
    echo ""
    echo "---"
    echo ""
}

# GET /hello
test_endpoint "1. GET /hello - Simple hello world" \
    "curl -s $BASE_URL/hello | jq ."

# GET /greet/:name
test_endpoint "2. GET /greet/:name - Greeting with name parameter" \
    "curl -s $BASE_URL/greet/Alice | jq ."

# GET /health
test_endpoint "3. GET /health - Health check endpoint" \
    "curl -s $BASE_URL/health | jq ."

# GET /api/users
test_endpoint "4. GET /api/users - Get all users" \
    "curl -s $BASE_URL/api/users | jq ."

# GET /api/users with query params
test_endpoint "5. GET /api/users?page=2&limit=10 - Get users with pagination" \
    "curl -s '$BASE_URL/api/users?page=2&limit=10' | jq ."

# GET /api/users/:id
test_endpoint "6. GET /api/users/:id - Get user by ID" \
    "curl -s $BASE_URL/api/users/42 | jq ."

# POST /api/users
test_endpoint "7. POST /api/users - Create new user" \
    "curl -s -X POST $BASE_URL/api/users \
    -H 'Content-Type: application/json' \
    -d '{\"name\": \"John Doe\", \"email\": \"john@example.com\"}' | jq ."

# PUT /api/users/:id
test_endpoint "8. PUT /api/users/:id - Update user" \
    "curl -s -X PUT $BASE_URL/api/users/123 \
    -H 'Content-Type: application/json' \
    -d '{\"name\": \"Jane Doe\", \"email\": \"jane@example.com\"}' | jq ."

# PATCH /api/users/:id
test_endpoint "9. PATCH /api/users/:id - Partial update user" \
    "curl -s -X PATCH $BASE_URL/api/users/123 \
    -H 'Content-Type: application/json' \
    -d '{\"email\": \"newemail@example.com\"}' | jq ."

# DELETE /api/users/:id
test_endpoint "10. DELETE /api/users/:id - Delete user" \
    "curl -s -X DELETE $BASE_URL/api/users/123 | jq ."

# GET /api/users/:userId/posts/:postId
test_endpoint "11. GET /api/users/:userId/posts/:postId - Nested resources" \
    "curl -s $BASE_URL/api/users/42/posts/99 | jq ."

# Test 404 error
test_endpoint "12. GET /invalid - Test 404 error" \
    "curl -s $BASE_URL/invalid | jq ."

# Test invalid JSON
test_endpoint "13. POST /api/users - Test invalid JSON error" \
    "curl -s -X POST $BASE_URL/api/users \
    -H 'Content-Type: application/json' \
    -d '{invalid json}' | jq ."

echo -e "${GREEN}Testing complete!${NC}"
