#!/bin/bash

# GlyphLang Demo Script
# Run this script to demonstrate GlyphLang examples in sequence

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
GRAY='\033[0;90m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:3000}"
PAUSE_SHORT=2
PAUSE_MEDIUM=4
PAUSE_LONG=6

# Store token for auth demo
TOKEN=""

# Helper functions
print_header() {
    echo ""
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${MAGENTA}  $1${NC}"
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

print_section() {
    echo ""
    echo -e "${CYAN}── $1 ──${NC}"
    echo ""
}

print_comment() {
    echo -e "${GRAY}# $1${NC}"
}

print_command() {
    echo -e "${YELLOW}$ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_code() {
    echo -e "${WHITE}$1${NC}"
}

pause() {
    local duration=${1:-$PAUSE_MEDIUM}
    sleep $duration
}

wait_for_key() {
    echo ""
    echo -e "${GRAY}Press Enter to continue...${NC}"
    read -r
}

run_curl() {
    local description="$1"
    shift
    local cmd="$*"

    echo ""
    print_comment "$description"
    print_command "curl $cmd"
    echo ""

    # Run curl with nice formatting
    response=$(curl -s -w "\n" "$@" 2>&1)

    # Try to pretty print JSON, fall back to raw output
    if command -v jq &> /dev/null; then
        echo "$response" | jq . 2>/dev/null || echo "$response"
    else
        echo "$response"
    fi

    echo ""
    pause $PAUSE_SHORT
}

# ============================================================================
# INTRO
# ============================================================================

clear
print_header "GlyphLang Demo"

echo -e "${WHITE}Welcome to the GlyphLang demonstration!${NC}"
echo ""
echo -e "GlyphLang is the ${CYAN}AI-first backend language for REST APIs${NC}."
echo ""
echo -e "Key features:"
echo -e "  ${GREEN}@${NC} - Define HTTP routes"
echo -e "  ${GREEN}:${NC} - Define types"
echo -e "  ${GREEN}\$${NC} - Declare variables"
echo -e "  ${GREEN}>${NC} - Return responses"
echo -e "  ${GREEN}+${NC} - Apply middleware (auth, rate limiting)"
echo -e "  ${GREEN}%${NC} - Inject dependencies (database)"
echo ""
echo -e "${BLUE}Token efficiency: 23% fewer tokens than FastAPI, 57% fewer than Java${NC}"
echo ""

wait_for_key

# ============================================================================
# DEMO 1: HELLO WORLD
# ============================================================================

clear
print_header "Demo 1: Hello World"

print_info "This example shows the basic GlyphLang syntax in just 21 lines."
echo ""
print_section "The Code"

echo -e "${WHITE}"
cat << 'EOF'
# Hello World Example

@ GET / {
  > {
    text: "Hello, World!",
    timestamp: 1234567890
  }
}

@ GET /hello/:name {
  $ greeting = "Hello, " + name + "!"
  > {message: greeting}
}

@ GET /health {
  > {status: "ok"}
}
EOF
echo -e "${NC}"

wait_for_key

print_section "Live Requests"

run_curl "Basic hello world endpoint" \
    "$BASE_URL/"

run_curl "Dynamic greeting with path parameter" \
    "$BASE_URL/hello/GlyphLang"

run_curl "Health check endpoint" \
    "$BASE_URL/health"

print_success "Hello World demo complete!"
wait_for_key

# ============================================================================
# DEMO 2: TODO API
# ============================================================================

clear
print_header "Demo 2: Todo API"

print_info "A full CRUD API with types, validation, rate limiting, and database operations."
echo ""
print_section "Key Syntax Highlights"

echo -e "${WHITE}"
cat << 'EOF'
# Type definition
: Todo {
  id: int!          # Required integer
  title: str!       # Required string
  completed: bool!  # Required boolean
  priority: str     # Optional string
}

# Route with middleware and dependency injection
@ GET /api/todos -> TodoList {
  + ratelimit(100/min)    # Rate limiting
  % db: Database          # Database injection

  $ todos = db.todos.all()
  > {todos: todos, total: todos.length()}
}
EOF
echo -e "${NC}"

wait_for_key

print_section "Live Requests"

run_curl "Get all todos" \
    "$BASE_URL/api/todos"

run_curl "Create a new todo" \
    -X POST "$BASE_URL/api/todos" \
    -H "Content-Type: application/json" \
    -d '{"title": "Demo GlyphLang at meeting", "description": "Show the team how it works", "priority": "high"}'

run_curl "Create another todo" \
    -X POST "$BASE_URL/api/todos" \
    -H "Content-Type: application/json" \
    -d '{"title": "Update LinkedIn post", "priority": "medium"}'

run_curl "Get a specific todo by ID" \
    "$BASE_URL/api/todos/1"

run_curl "Mark todo as completed" \
    -X PATCH "$BASE_URL/api/todos/1/complete"

run_curl "Get todos by priority" \
    "$BASE_URL/api/todos/priority/high"

run_curl "Update a todo" \
    -X PUT "$BASE_URL/api/todos/2" \
    -H "Content-Type: application/json" \
    -d '{"title": "Update LinkedIn post", "completed": true, "priority": "low"}'

run_curl "Delete a todo" \
    -X DELETE "$BASE_URL/api/todos/2"

run_curl "Verify deletion - get all todos" \
    "$BASE_URL/api/todos"

print_success "Todo API demo complete!"
wait_for_key

# ============================================================================
# DEMO 3: WEBSOCKET CHAT
# ============================================================================

clear
print_header "Demo 3: WebSocket Chat"

print_info "Real-time WebSocket chat in just 44 lines of code."
echo ""
print_section "The Code"

echo -e "${WHITE}"
cat << 'EOF'
# WebSocket chat endpoint
@ ws /chat {
  on connect {
    ws.join("lobby")
    ws.broadcast("User joined the chat")
  }

  on message {
    ws.broadcast(input)
  }

  on disconnect {
    ws.broadcast("User left the chat")
    ws.leave("lobby")
  }
}

# WebSocket with room parameter
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
EOF
echo -e "${NC}"

echo ""
print_info "WebSocket endpoints available at:"
echo -e "  ${CYAN}ws://localhost:3000/chat${NC}        - Global chat"
echo -e "  ${CYAN}ws://localhost:3000/chat/:room${NC}  - Room-based chat"
echo ""
print_info "To test interactively, use wscat:"
echo -e "  ${YELLOW}wscat -c ws://localhost:3000/chat${NC}"
echo ""

run_curl "Check WebSocket server status" \
    "$BASE_URL/status"

print_success "WebSocket Chat demo complete!"
wait_for_key

# ============================================================================
# DEMO 4: AUTH
# ============================================================================

clear
print_header "Demo 4: Authentication & Authorization"

print_info "JWT authentication, protected routes, and role-based access control."
echo ""
print_section "Key Syntax Highlights"

echo -e "${WHITE}"
cat << 'EOF'
# Public endpoint - no auth required
@ GET /api/health {
  + ratelimit(1000/min)
  > {status: "ok"}
}

# Protected endpoint - requires valid JWT
@ GET /api/auth/me -> UserProfile {
  + auth(jwt)
  + ratelimit(100/min)
  % db: Database

  $ userId = auth.user.id
  $ user = db.users.get(userId)
  > user
}

# Admin-only endpoint - requires JWT with admin role
@ GET /api/admin/users {
  + auth(jwt, role: admin)
  + ratelimit(50/min)
  % db: Database

  $ allUsers = db.users.all()
  > {users: allUsers, total: allUsers.length()}
}
EOF
echo -e "${NC}"

wait_for_key

print_section "Live Requests"

run_curl "Health check (public endpoint)" \
    "$BASE_URL/api/health"

run_curl "Register a new user" \
    -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"username": "demouser", "email": "demo@glyphlang.dev", "password": "securepass123"}'

print_comment "Now let's login and get a token..."
echo ""

# Login and capture token
print_command "curl -X POST $BASE_URL/api/auth/login ..."
login_response=$(curl -s -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username": "demouser", "password": "securepass123"}')

echo "$login_response" | jq . 2>/dev/null || echo "$login_response"

# Extract token (assumes jq is available)
if command -v jq &> /dev/null; then
    TOKEN=$(echo "$login_response" | jq -r '.token // empty')
fi

if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    echo ""
    print_success "Token received!"
    echo -e "${GRAY}Token: ${TOKEN:0:50}...${NC}"
    pause $PAUSE_SHORT

    run_curl "Get current user profile (protected route)" \
        "$BASE_URL/api/auth/me" \
        -H "Authorization: Bearer $TOKEN"

    run_curl "Update user profile" \
        -X PUT "$BASE_URL/api/auth/profile" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"full_name": "Demo User", "bio": "Testing GlyphLang auth"}'

    run_curl "Refresh token" \
        -X POST "$BASE_URL/api/auth/refresh" \
        -H "Authorization: Bearer $TOKEN"

    print_comment "Attempting admin endpoint (should fail - user is not admin)..."
    run_curl "Try admin-only endpoint" \
        "$BASE_URL/api/admin/users" \
        -H "Authorization: Bearer $TOKEN"

    run_curl "Logout" \
        -X POST "$BASE_URL/api/auth/logout" \
        -H "Authorization: Bearer $TOKEN"
else
    echo ""
    print_info "Could not extract token - server may not be running or returned an error"
fi

print_success "Auth demo complete!"
wait_for_key

# ============================================================================
# SUMMARY
# ============================================================================

clear
print_header "Demo Complete!"

echo -e "${WHITE}What we demonstrated:${NC}"
echo ""
echo -e "  ${GREEN}1. Hello World${NC} - Basic syntax: routes (@), variables (\$), returns (>)"
echo -e "  ${GREEN}2. Todo API${NC}    - CRUD, types (:), rate limiting (+), database (%)"
echo -e "  ${GREEN}3. WebSocket${NC}   - Real-time chat with rooms in 44 lines"
echo -e "  ${GREEN}4. Auth${NC}        - JWT authentication and role-based access control"
echo ""
echo -e "${CYAN}Key Takeaways:${NC}"
echo ""
echo -e "  • ${WHITE}23% fewer tokens${NC} than FastAPI"
echo -e "  • ${WHITE}57% fewer tokens${NC} than Java/Spring"
echo -e "  • ${WHITE}Built-in${NC} auth, rate limiting, database, WebSockets"
echo -e "  • ${WHITE}Designed for AI${NC} code generation and human review"
echo ""
echo -e "${BLUE}Learn more:${NC}"
echo -e "  Docs:   ${CYAN}https://glyphlang.dev/docs${NC}"
echo -e "  GitHub: ${CYAN}https://github.com/glyphlang/glyph${NC}"
echo ""
echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
