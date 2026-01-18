#!/usr/bin/env python3
"""
AI Efficiency Benchmark - Compare token usage and generation efficiency
across Glyph, FastAPI, Flask, and Java for equivalent functionality.

This measures the token efficiency of Glyph's minimal-ceremony design.
Token counts are measured using tiktoken (OpenAI's tokenizer).

Requirements:
    pip install tiktoken
"""

import json
from dataclasses import dataclass
from typing import List, Dict

try:
    import tiktoken
    TIKTOKEN_AVAILABLE = True
except ImportError:
    TIKTOKEN_AVAILABLE = False
    print("WARNING: tiktoken not installed. Install with: pip install tiktoken")
    print("Falling back to character-based estimation.\n")

@dataclass
class CodeSample:
    name: str
    description: str
    glyph: str
    fastapi: str
    flask: str
    java: str

# Equivalent code samples in each language
SAMPLES: List[CodeSample] = [
    CodeSample(
        name="Hello World API",
        description="Simple GET endpoint returning JSON",
        glyph='''@ GET /hello {
  > {message: "Hello, World!"}
}''',
        fastapi='''@app.get("/hello")
def hello():
    return {"message": "Hello, World!"}''',
        flask='''from flask import Flask, jsonify

app = Flask(__name__)

@app.route('/hello', methods=['GET'])
def hello():
    return jsonify({"message": "Hello, World!"})''',
        java='''import org.springframework.web.bind.annotation.*;

@RestController
public class HelloController {
    @GetMapping("/hello")
    public Map<String, String> hello() {
        return Map.of("message", "Hello, World!");
    }
}'''
    ),
    CodeSample(
        name="User GET with Path Param",
        description="GET endpoint with path parameter and type",
        glyph='''@ GET /users/:id -> User {
  $ user = db.users.get(id)
  > user
}''',
        fastapi='''@app.get("/users/{id}")
def get_user(id: int) -> User:
    return db.users.get(id)''',
        flask='''@app.route('/users/<int:id>', methods=['GET'])
def get_user(id: int) -> User:
    user = db.users.get(id)
    return jsonify(user)''',
        java='''@GetMapping("/users/{id}")
public User getUser(@PathVariable Long id) {
    User user = userRepository.findById(id);
    return user;
}'''
    ),
    CodeSample(
        name="Protected Route with Auth",
        description="Endpoint with JWT authentication",
        glyph='''@ GET /api/data {
  + auth(jwt)
  > {data: "secret"}
}''',
        fastapi='''@app.get("/api/data")
def get_data(user: User = Depends(get_current_user)):
    return {"data": "secret"}''',
        flask='''@app.route('/api/data', methods=['GET'])
@jwt_required()
def get_data():
    return jsonify({"data": "secret"})''',
        java='''@GetMapping("/api/data")
@PreAuthorize("isAuthenticated()")
public Map<String, String> getData() {
    return Map.of("data", "secret");
}'''
    ),
    CodeSample(
        name="POST with Validation",
        description="Create endpoint with input validation",
        glyph='''@ POST /users -> User {
  + auth(jwt)
  < input: CreateUser
  ! validate input {
    name: str(min=1, max=100)
    email: email
  }
  $ user = db.users.create(input)
  > user
}''',
        fastapi='''class CreateUser(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    email: EmailStr

@app.post("/users")
def create_user(input: CreateUser, user: User = Depends(get_current_user)) -> User:
    return db.users.create(input)''',
        flask='''@app.route('/users', methods=['POST'])
@jwt_required()
def create_user():
    data = request.get_json()

    if not data.get('name') or len(data['name']) > 100:
        return jsonify({"error": "Invalid name"}), 400
    if not validate_email(data.get('email')):
        return jsonify({"error": "Invalid email"}), 400

    user = db.users.create(data)
    return jsonify(user), 201''',
        java='''@PostMapping("/users")
@PreAuthorize("isAuthenticated()")
public ResponseEntity<User> createUser(@Valid @RequestBody CreateUserDto input) {
    if (input.getName() == null || input.getName().length() > 100) {
        throw new ValidationException("Invalid name");
    }
    if (!EmailValidator.isValid(input.getEmail())) {
        throw new ValidationException("Invalid email");
    }

    User user = userService.create(input);
    return ResponseEntity.status(HttpStatus.CREATED).body(user);
}'''
    ),
    CodeSample(
        name="Type Definition",
        description="Define a data type/model",
        glyph=''': User {
  id: int!
  name: str!
  email: str!
  age: int
  active: bool!
}''',
        fastapi='''class User(BaseModel):
    id: int
    name: str
    email: str
    age: int | None = None
    active: bool = True''',
        flask='''from dataclasses import dataclass
from typing import Optional

@dataclass
class User:
    id: int
    name: str
    email: str
    age: Optional[int] = None
    active: bool = True''',
        java='''public class User {
    private Long id;
    private String name;
    private String email;
    private Integer age;
    private boolean active;

    // Getters and setters...
    public Long getId() { return id; }
    public void setId(Long id) { this.id = id; }
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
    public String getEmail() { return email; }
    public void setEmail(String email) { this.email = email; }
    public Integer getAge() { return age; }
    public void setAge(Integer age) { this.age = age; }
    public boolean isActive() { return active; }
    public void setActive(boolean active) { this.active = active; }
}'''
    ),
    CodeSample(
        name="CRUD API",
        description="Full CRUD operations for a resource",
        glyph=''': Todo {
  id: int!
  title: str!
  done: bool!
}

@ GET /todos -> List[Todo] {
  > db.todos.all()
}

@ GET /todos/:id -> Todo {
  > db.todos.get(id)
}

@ POST /todos -> Todo {
  < input: Todo
  > db.todos.create(input)
}

@ PUT /todos/:id -> Todo {
  < input: Todo
  > db.todos.update(id, input)
}

@ DELETE /todos/:id {
  > db.todos.delete(id)
}''',
        fastapi='''class Todo(BaseModel):
    id: int
    title: str
    done: bool

@app.get("/todos")
def list_todos() -> list[Todo]:
    return db.todos.all()

@app.get("/todos/{id}")
def get_todo(id: int) -> Todo:
    return db.todos.get(id)

@app.post("/todos", status_code=201)
def create_todo(todo: Todo) -> Todo:
    return db.todos.create(todo)

@app.put("/todos/{id}")
def update_todo(id: int, todo: Todo) -> Todo:
    return db.todos.update(id, todo)

@app.delete("/todos/{id}", status_code=204)
def delete_todo(id: int):
    db.todos.delete(id)''',
        flask='''from flask import Flask, request, jsonify

@app.route('/todos', methods=['GET'])
def list_todos():
    return jsonify(db.todos.all())

@app.route('/todos/<int:id>', methods=['GET'])
def get_todo(id):
    return jsonify(db.todos.get(id))

@app.route('/todos', methods=['POST'])
def create_todo():
    data = request.get_json()
    return jsonify(db.todos.create(data)), 201

@app.route('/todos/<int:id>', methods=['PUT'])
def update_todo(id):
    data = request.get_json()
    return jsonify(db.todos.update(id, data))

@app.route('/todos/<int:id>', methods=['DELETE'])
def delete_todo(id):
    db.todos.delete(id)
    return '', 204''',
        java='''@RestController
@RequestMapping("/todos")
public class TodoController {
    @Autowired
    private TodoRepository todoRepository;

    @GetMapping
    public List<Todo> listTodos() {
        return todoRepository.findAll();
    }

    @GetMapping("/{id}")
    public Todo getTodo(@PathVariable Long id) {
        return todoRepository.findById(id).orElseThrow();
    }

    @PostMapping
    public ResponseEntity<Todo> createTodo(@RequestBody Todo todo) {
        Todo saved = todoRepository.save(todo);
        return ResponseEntity.status(HttpStatus.CREATED).body(saved);
    }

    @PutMapping("/{id}")
    public Todo updateTodo(@PathVariable Long id, @RequestBody Todo todo) {
        todo.setId(id);
        return todoRepository.save(todo);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteTodo(@PathVariable Long id) {
        todoRepository.deleteById(id);
        return ResponseEntity.noContent().build();
    }
}'''
    ),
    CodeSample(
        name="WebSocket Handler",
        description="WebSocket endpoint with room support",
        glyph='''@ ws /chat/:room {
  on connect {
    ws.join(room)
    ws.broadcast(room, {event: "joined", user: ws.id})
  }
  on message {
    ws.broadcast(room, {user: ws.id, text: data.text})
  }
  on disconnect {
    ws.broadcast(room, {event: "left", user: ws.id})
  }
}''',
        fastapi='''rooms: dict[str, set[WebSocket]] = {}

@app.websocket("/chat/{room}")
async def chat(websocket: WebSocket, room: str):
    await websocket.accept()
    rooms.setdefault(room, set()).add(websocket)
    await broadcast(room, {"event": "joined", "user": id(websocket)})
    try:
        while True:
            data = await websocket.receive_json()
            await broadcast(room, {"user": id(websocket), "text": data["text"]})
    except WebSocketDisconnect:
        rooms[room].remove(websocket)
        await broadcast(room, {"event": "left", "user": id(websocket)})''',
        flask='''from flask_socketio import SocketIO, join_room, emit

@socketio.on('connect')
def handle_connect():
    room = request.args.get('room')
    join_room(room)
    emit('message', {'event': 'joined', 'user': request.sid}, room=room)

@socketio.on('message')
def handle_message(data):
    room = request.args.get('room')
    emit('message', {'user': request.sid, 'text': data['text']}, room=room)

@socketio.on('disconnect')
def handle_disconnect():
    room = request.args.get('room')
    emit('message', {'event': 'left', 'user': request.sid}, room=room)''',
        java='''@ServerEndpoint("/chat/{room}")
public class ChatEndpoint {
    private static Map<String, Set<Session>> rooms = new ConcurrentHashMap<>();

    @OnOpen
    public void onOpen(Session session, @PathParam("room") String room) {
        rooms.computeIfAbsent(room, k -> ConcurrentHashMap.newKeySet()).add(session);
        broadcast(room, Map.of("event", "joined", "user", session.getId()));
    }

    @OnMessage
    public void onMessage(String message, Session session, @PathParam("room") String room) {
        JsonObject data = Json.createReader(new StringReader(message)).readObject();
        broadcast(room, Map.of("user", session.getId(), "text", data.getString("text")));
    }

    @OnClose
    public void onClose(Session session, @PathParam("room") String room) {
        rooms.get(room).remove(session);
        broadcast(room, Map.of("event", "left", "user", session.getId()));
    }

    private void broadcast(String room, Map<String, String> message) {
        String json = Json.createObjectBuilder(message).build().toString();
        rooms.get(room).forEach(s -> s.getAsyncRemote().sendText(json));
    }
}'''
    ),
]


def count_tokens(text: str, encoding_name: str = "cl100k_base") -> int:
    """
    Count tokens using tiktoken (OpenAI's tokenizer).

    Args:
        text: The text to tokenize
        encoding_name: The encoding to use:
            - "cl100k_base": GPT-4, GPT-3.5-turbo, text-embedding-ada-002
            - "o200k_base": GPT-4o

    Returns:
        Actual token count from tiktoken
    """
    if TIKTOKEN_AVAILABLE:
        encoding = tiktoken.get_encoding(encoding_name)
        return len(encoding.encode(text))
    else:
        # Fallback: rough estimation (~4 chars per token)
        return len(text) // 4


def count_lines(text: str) -> int:
    """Count non-empty lines"""
    return len([l for l in text.strip().split('\n') if l.strip()])


def analyze_sample(sample: CodeSample, encoding: str = "cl100k_base") -> Dict:
    """Analyze a code sample across all languages using tiktoken"""

    glyph_tokens = count_tokens(sample.glyph, encoding)
    fastapi_tokens = count_tokens(sample.fastapi, encoding)
    flask_tokens = count_tokens(sample.flask, encoding)
    java_tokens = count_tokens(sample.java, encoding)

    return {
        'name': sample.name,
        'description': sample.description,
        'glyph': {
            'chars': len(sample.glyph),
            'lines': count_lines(sample.glyph),
            'tokens': glyph_tokens,
        },
        'fastapi': {
            'chars': len(sample.fastapi),
            'lines': count_lines(sample.fastapi),
            'tokens': fastapi_tokens,
            'vs_glyph_chars': round(len(sample.fastapi) / len(sample.glyph), 2),
            'vs_glyph_tokens': round(fastapi_tokens / glyph_tokens, 2),
        },
        'flask': {
            'chars': len(sample.flask),
            'lines': count_lines(sample.flask),
            'tokens': flask_tokens,
            'vs_glyph_chars': round(len(sample.flask) / len(sample.glyph), 2),
            'vs_glyph_tokens': round(flask_tokens / glyph_tokens, 2),
        },
        'java': {
            'chars': len(sample.java),
            'lines': count_lines(sample.java),
            'tokens': java_tokens,
            'vs_glyph_chars': round(len(sample.java) / len(sample.glyph), 2),
            'vs_glyph_tokens': round(java_tokens / glyph_tokens, 2),
        }
    }


def calculate_ai_cost(tokens: int, model: str = "gpt-4") -> Dict:
    """
    Calculate estimated AI generation cost.
    Prices as of 2024 (per 1M tokens):
    - GPT-4: $30 input, $60 output
    - GPT-3.5: $0.50 input, $1.50 output
    - Claude 3 Opus: $15 input, $75 output
    - Claude 3 Sonnet: $3 input, $15 output
    """
    prices = {
        'gpt-4': {'input': 30, 'output': 60},
        'gpt-3.5': {'input': 0.5, 'output': 1.5},
        'claude-opus': {'input': 15, 'output': 75},
        'claude-sonnet': {'input': 3, 'output': 15},
    }

    p = prices.get(model, prices['gpt-4'])
    # Assume 2:1 output:input ratio for code generation
    input_cost = (tokens * 0.33) * p['input'] / 1_000_000
    output_cost = (tokens * 0.67) * p['output'] / 1_000_000

    return {
        'model': model,
        'total_cost_usd': round(input_cost + output_cost, 6),
        'cost_per_1k_tokens': round((p['input'] * 0.33 + p['output'] * 0.67) / 1000, 4)
    }


def run_analysis():
    """Run full analysis and print results"""

    print("=" * 100)
    print("AI EFFICIENCY BENCHMARK: Glyph vs FastAPI vs Flask vs Java")
    print("=" * 100)
    print("\nMeasuring token efficiency for AI/LLM code generation")
    print("Lower tokens = faster generation, lower cost\n")

    if TIKTOKEN_AVAILABLE:
        print("Tokenizer: tiktoken (cl100k_base encoding - GPT-4/GPT-3.5)")
    else:
        print("Tokenizer: FALLBACK (character estimation - install tiktoken for accuracy)")

    print("-" * 100)

    results = []
    totals = {'glyph': {'chars': 0, 'lines': 0, 'tokens': 0},
              'fastapi': {'chars': 0, 'lines': 0, 'tokens': 0},
              'flask': {'chars': 0, 'lines': 0, 'tokens': 0},
              'java': {'chars': 0, 'lines': 0, 'tokens': 0}}

    for sample in SAMPLES:
        analysis = analyze_sample(sample)
        results.append(analysis)

        for lang in ['glyph', 'fastapi', 'flask', 'java']:
            totals[lang]['chars'] += analysis[lang]['chars']
            totals[lang]['lines'] += analysis[lang]['lines']
            totals[lang]['tokens'] += analysis[lang]['tokens']

    # Print per-sample results
    print(f"{'Sample':<25} {'Glyph':<8} {'FastAPI':<8} {'Flask':<8} {'Java':<8} {'Fast/Gly':<9} {'Flsk/Gly':<9} {'Java/Gly':<9}")
    print(f"{'(tokens)':<25} {'tok':<8} {'tok':<8} {'tok':<8} {'tok':<8} {'ratio':<9} {'ratio':<9} {'ratio':<9}")
    print("-" * 100)

    for r in results:
        print(f"{r['name']:<25} {r['glyph']['tokens']:<8} {r['fastapi']['tokens']:<8} {r['flask']['tokens']:<8} {r['java']['tokens']:<8} {r['fastapi']['vs_glyph_tokens']:<9} {r['flask']['vs_glyph_tokens']:<9} {r['java']['vs_glyph_tokens']:<9}")

    print("-" * 100)
    print(f"{'TOTAL':<25} {totals['glyph']['tokens']:<8} {totals['fastapi']['tokens']:<8} {totals['flask']['tokens']:<8} {totals['java']['tokens']:<8} {round(totals['fastapi']['tokens']/totals['glyph']['tokens'], 2):<9} {round(totals['flask']['tokens']/totals['glyph']['tokens'], 2):<9} {round(totals['java']['tokens']/totals['glyph']['tokens'], 2):<9}")

    # Summary statistics
    print("\n" + "=" * 100)
    print("SUMMARY: TOKEN EFFICIENCY")
    print("=" * 100)

    print(f"\n{'Language':<15} {'Total Chars':<15} {'Total Lines':<15} {'Total Tokens':<15}")
    print("-" * 60)
    print(f"{'Glyph':<15} {totals['glyph']['chars']:<15} {totals['glyph']['lines']:<15} {totals['glyph']['tokens']:<15}")
    print(f"{'FastAPI':<15} {totals['fastapi']['chars']:<15} {totals['fastapi']['lines']:<15} {totals['fastapi']['tokens']:<15}")
    print(f"{'Flask':<15} {totals['flask']['chars']:<15} {totals['flask']['lines']:<15} {totals['flask']['tokens']:<15}")
    print(f"{'Java':<15} {totals['java']['chars']:<15} {totals['java']['lines']:<15} {totals['java']['tokens']:<15}")

    # Token savings
    fastapi_savings = round((1 - totals['glyph']['tokens'] / totals['fastapi']['tokens']) * 100, 1)
    flask_savings = round((1 - totals['glyph']['tokens'] / totals['flask']['tokens']) * 100, 1)
    java_savings = round((1 - totals['glyph']['tokens'] / totals['java']['tokens']) * 100, 1)

    print(f"\n{'TOKEN SAVINGS WITH Glyph:':<40}")
    print(f"  vs FastAPI: {fastapi_savings}% fewer tokens")
    print(f"  vs Flask:   {flask_savings}% fewer tokens")
    print(f"  vs Java:    {java_savings}% fewer tokens")

    # AI cost comparison
    print("\n" + "=" * 100)
    print("AI GENERATION COST COMPARISON (per 1000 API calls)")
    print("=" * 100)

    models = ['gpt-4', 'claude-opus', 'claude-sonnet', 'gpt-3.5']

    print(f"\n{'Model':<18} {'Glyph':<12} {'FastAPI':<12} {'Flask':<12} {'Java':<12} {'Glyph Savings':<15}")
    print("-" * 100)

    for model in models:
        glyph_cost = calculate_ai_cost(totals['glyph']['tokens'] * 1000, model)
        fastapi_cost = calculate_ai_cost(totals['fastapi']['tokens'] * 1000, model)
        flask_cost = calculate_ai_cost(totals['flask']['tokens'] * 1000, model)
        java_cost = calculate_ai_cost(totals['java']['tokens'] * 1000, model)

        avg_savings = round((1 - glyph_cost['total_cost_usd'] / ((fastapi_cost['total_cost_usd'] + flask_cost['total_cost_usd'] + java_cost['total_cost_usd']) / 3)) * 100, 1)

        print(f"{model:<18} ${glyph_cost['total_cost_usd']:<11.2f} ${fastapi_cost['total_cost_usd']:<11.2f} ${flask_cost['total_cost_usd']:<11.2f} ${java_cost['total_cost_usd']:<11.2f} {avg_savings}%")

    # AI utilization benefits
    print("\n" + "=" * 100)
    print("AI UTILIZATION BENEFITS")
    print("=" * 100)

    print("""
+--------------------------------------------------------------------------------------------------+
| Glyph's AI-First Design Advantages:                                                              |
+--------------------------------------------------------------------------------------------------+
|                                                                                                  |
| 1. TOKEN EFFICIENCY                                                                              |
|    - No imports or framework initialization required                                             |
|    - Routes and types are the entire program                                                     |
|    - Fewer tokens than FastAPI, Flask, and Java                                                  |
|                                                                                                  |
| 2. GENERATION SPEED                                                                              |
|    - Fewer tokens = faster LLM response time                                                     |
|    - Reduced context window usage                                                                |
|    - More room for complex requirements in prompts                                               |
|                                                                                                  |
| 3. COST SAVINGS                                                                                  |
|    - Significant reduction in token usage across all comparisons                                 |
|    - Proportional savings on API costs at scale                                                  |
|                                                                                                  |
| 4. CONTEXT EFFICIENCY                                                                            |
|    - More business logic fits in context window                                                  |
|    - Better for complex multi-file operations                                                    |
|    - Enables larger refactoring tasks                                                            |
|                                                                                                  |
+--------------------------------------------------------------------------------------------------+
""")

    # JSON output
    print("\n" + "=" * 100)
    print("JSON OUTPUT")
    print("=" * 100)

    output = {
        'samples': results,
        'totals': totals,
        'savings': {
            'vs_fastapi_percent': fastapi_savings,
            'vs_flask_percent': flask_savings,
            'vs_java_percent': java_savings,
        }
    }
    print(json.dumps(output, indent=2))

    return output


if __name__ == '__main__':
    run_analysis()
