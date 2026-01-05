# GlyphLang Architecture Design

This document provides comprehensive architecture diagrams for the GlyphLang compiler and runtime system.

## Table of Contents
1. [High-Level Architecture](#high-level-architecture)
2. [Compilation Pipeline](#compilation-pipeline)
3. [Package Dependencies](#package-dependencies)
4. [Request Execution Flow](#request-execution-flow)
5. [AST Structure](#ast-structure)
6. [Virtual Machine Architecture](#virtual-machine-architecture)
7. [Server Architecture](#server-architecture)
8. [Database Integration](#database-integration)
9. [WebSocket Architecture](#websocket-architecture)
10. [JIT Compilation](#jit-compilation)

---

## High-Level Architecture

```mermaid
graph TB
    subgraph "Development"
        SRC[".glyph Source Files"]
        CLI["CLI (cmd/glyph)"]
        LSP["LSP Server"]
        IDE["IDE Integration"]
    end

    subgraph "Compilation"
        LEX["Lexer"]
        PAR["Parser"]
        TC["Type Checker"]
        COMP["Compiler"]
        OPT["Optimizer"]
        BC[".glyphc Bytecode"]
    end

    subgraph "Runtime"
        VM["Virtual Machine"]
        INTERP["Tree-Walking Interpreter"]
        JIT["JIT Compiler"]
    end

    subgraph "Server Layer"
        HTTP["HTTP Server"]
        WS["WebSocket Server"]
        MW["Middleware Chain"]
        ROUTER["Router"]
    end

    subgraph "Infrastructure"
        DB["Database (PostgreSQL)"]
        CACHE["Cache Layer"]
        METRICS["Prometheus Metrics"]
        TRACE["OpenTelemetry Tracing"]
        LOG["Structured Logging"]
    end

    SRC --> CLI
    CLI --> LEX
    IDE --> LSP
    LSP --> LEX

    LEX --> PAR
    PAR --> TC
    TC --> COMP
    COMP --> OPT
    OPT --> BC

    BC --> VM
    BC --> JIT
    PAR --> INTERP

    VM --> HTTP
    INTERP --> HTTP
    JIT --> VM

    HTTP --> ROUTER
    ROUTER --> MW
    MW --> WS

    HTTP --> DB
    HTTP --> CACHE
    HTTP --> METRICS
    HTTP --> TRACE
    HTTP --> LOG
```

---

## Compilation Pipeline

```mermaid
flowchart LR
    subgraph Input
        SOURCE["Source Code\n(.glyph)"]
    end

    subgraph Lexer
        TOK["Tokenization"]
        TOKENS["Token Stream\n@ : $ > + %"]
    end

    subgraph Parser
        RDP["Recursive Descent\nParser"]
        AST["Abstract Syntax\nTree"]
    end

    subgraph "Semantic Analysis"
        TYPE["Type Checker"]
        SYM["Symbol Table"]
    end

    subgraph "Code Generation"
        CODEGEN["Bytecode\nGenerator"]
        OPTIM["Optimizer\nLevels 0-3"]
    end

    subgraph Output
        BYTECODE["Bytecode\n(.glyphc)"]
    end

    SOURCE --> TOK
    TOK --> TOKENS
    TOKENS --> RDP
    RDP --> AST
    AST --> TYPE
    TYPE --> SYM
    SYM --> CODEGEN
    CODEGEN --> OPTIM
    OPTIM --> BYTECODE

    style SOURCE fill:#e1f5fe
    style BYTECODE fill:#c8e6c9
```

### Compilation Stages Detail

```mermaid
sequenceDiagram
    participant S as Source
    participant L as Lexer
    participant P as Parser
    participant T as TypeChecker
    participant C as Compiler
    participant O as Optimizer
    participant B as Binary

    S->>L: Read source file
    L->>L: Scan characters
    L->>P: Token stream
    P->>P: Build AST nodes
    P->>T: AST
    T->>T: Validate types
    T->>T: Build symbol table
    T->>C: Validated AST
    C->>C: Generate opcodes
    C->>C: Resolve variables
    C->>O: Raw bytecode
    O->>O: Constant folding
    O->>O: Dead code elimination
    O->>B: Optimized bytecode
    B->>B: Add magic header
    B->>B: Calculate CRC32
    B-->>S: .glyphc file
```

---

## Package Dependencies

```mermaid
graph TD
    subgraph "Entry Points"
        CMD["cmd/glyph"]
    end

    subgraph "Core Compilation"
        PARSER["pkg/parser"]
        INTERP["pkg/interpreter"]
        COMPILER["pkg/compiler"]
        VM["pkg/vm"]
    end

    subgraph "Server Infrastructure"
        SERVER["pkg/server"]
        WEBSOCKET["pkg/websocket"]
        DATABASE["pkg/database"]
    end

    subgraph "Cross-Cutting Concerns"
        SECURITY["pkg/security"]
        VALIDATION["pkg/validation"]
        ERRORS["pkg/errors"]
    end

    subgraph "Observability"
        LOGGING["pkg/logging"]
        METRICS["pkg/metrics"]
        TRACING["pkg/tracing"]
    end

    subgraph "Performance"
        JIT["pkg/jit"]
        CACHE["pkg/cache"]
        MEMORY["pkg/memory"]
    end

    subgraph "Developer Tools"
        LSP["pkg/lsp"]
        DEBUG["pkg/debug"]
        HOTRELOAD["pkg/hotreload"]
    end

    CMD --> PARSER
    CMD --> COMPILER
    CMD --> VM
    CMD --> SERVER
    CMD --> LSP

    PARSER --> INTERP
    INTERP --> ERRORS
    COMPILER --> INTERP
    COMPILER --> VM
    VM --> MEMORY

    SERVER --> INTERP
    SERVER --> VM
    SERVER --> WEBSOCKET
    SERVER --> DATABASE
    SERVER --> SECURITY
    SERVER --> VALIDATION

    WEBSOCKET --> VM

    JIT --> VM
    JIT --> COMPILER

    LSP --> PARSER
    DEBUG --> INTERP

    SERVER --> LOGGING
    SERVER --> METRICS
    SERVER --> TRACING
    SERVER --> CACHE

    HOTRELOAD --> SERVER
```

---

## Request Execution Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant S as Server
    participant R as Router
    participant MW as Middleware
    participant H as Handler
    participant I as Interpreter/VM
    participant DB as Database

    C->>S: HTTP Request
    S->>R: Match route pattern
    R->>R: Extract path params
    R->>MW: Route + Context

    loop Middleware Chain
        MW->>MW: Auth check
        MW->>MW: Rate limiting
        MW->>MW: Logging
        MW->>MW: Tracing
    end

    MW->>H: Validated request
    H->>H: Parse JSON body
    H->>H: Build environment
    H->>I: Execute route

    I->>I: Create scope
    I->>DB: Database queries
    DB-->>I: Results
    I->>I: Evaluate expressions
    I-->>H: Return value

    H->>H: Build JSON response
    H-->>C: HTTP Response
```

### Route Matching Detail

```mermaid
flowchart TD
    REQ["Incoming Request\nGET /api/users/123"]

    subgraph Router
        MATCH["Pattern Matcher"]
        ROUTES["Registered Routes"]
        PARAMS["Parameter Extractor"]
    end

    subgraph "Route Definitions"
        R1["GET /api/users"]
        R2["GET /api/users/:id"]
        R3["POST /api/users"]
    end

    REQ --> MATCH
    ROUTES --> R1
    ROUTES --> R2
    ROUTES --> R3
    MATCH --> ROUTES
    MATCH -->|"Matched: /api/users/:id"| PARAMS
    PARAMS -->|"id = 123"| CTX["Request Context"]

    style R2 fill:#c8e6c9
```

---

## AST Structure

```mermaid
classDiagram
    class Module {
        +Items []Item
        +Filename string
    }

    class Item {
        <<interface>>
    }

    class TypeDef {
        +Name string
        +Fields []Field
    }

    class Route {
        +Path string
        +Method HTTPMethod
        +Params []Param
        +ReturnType Type
        +Middleware []Middleware
        +Body []Statement
    }

    class Function {
        +Name string
        +Params []Param
        +ReturnType Type
        +Body []Statement
    }

    class Statement {
        <<interface>>
    }

    class AssignStmt {
        +Target string
        +Value Expression
    }

    class ReturnStmt {
        +Value Expression
    }

    class IfStmt {
        +Condition Expression
        +Then []Statement
        +Else []Statement
    }

    class Expression {
        <<interface>>
    }

    class Literal {
        +Value interface
        +Type LiteralType
    }

    class Variable {
        +Name string
    }

    class BinaryOp {
        +Left Expression
        +Op Operator
        +Right Expression
    }

    class FunctionCall {
        +Name string
        +Args []Expression
    }

    Module --> Item
    Item <|-- TypeDef
    Item <|-- Route
    Item <|-- Function

    Route --> Statement
    Function --> Statement

    Statement <|-- AssignStmt
    Statement <|-- ReturnStmt
    Statement <|-- IfStmt

    AssignStmt --> Expression
    ReturnStmt --> Expression
    IfStmt --> Expression

    Expression <|-- Literal
    Expression <|-- Variable
    Expression <|-- BinaryOp
    Expression <|-- FunctionCall
```

### Type System

```mermaid
classDiagram
    class Type {
        <<interface>>
    }

    class IntType {
        +Required bool
    }

    class StringType {
        +Required bool
    }

    class BoolType {
        +Required bool
    }

    class FloatType {
        +Required bool
    }

    class ArrayType {
        +ElementType Type
    }

    class OptionalType {
        +Inner Type
    }

    class NamedType {
        +Name string
    }

    class UnionType {
        +Types []Type
    }

    class DatabaseType {
        +TableName string
    }

    Type <|-- IntType
    Type <|-- StringType
    Type <|-- BoolType
    Type <|-- FloatType
    Type <|-- ArrayType
    Type <|-- OptionalType
    Type <|-- NamedType
    Type <|-- UnionType
    Type <|-- DatabaseType

    ArrayType --> Type
    OptionalType --> Type
    UnionType --> Type
```

---

## Virtual Machine Architecture

```mermaid
graph TB
    subgraph "VM State"
        PC["Program Counter"]
        STACK["Operand Stack"]
        LOCALS["Local Variables"]
        GLOBALS["Global Variables"]
        CONSTS["Constants Pool"]
    end

    subgraph "Execution"
        FETCH["Fetch Opcode"]
        DECODE["Decode Instruction"]
        EXEC["Execute"]
    end

    subgraph "Opcodes"
        direction LR
        ARITH["Arithmetic\nAdd Sub Mul Div"]
        COMPARE["Comparison\nEq Ne Lt Gt"]
        CONTROL["Control Flow\nJump JumpIf Return"]
        DATA["Data\nPush Pop Load Store"]
        COLL["Collections\nBuildArray BuildObject"]
    end

    PC --> FETCH
    FETCH --> DECODE
    DECODE --> EXEC
    EXEC --> STACK
    EXEC --> LOCALS
    EXEC --> GLOBALS
    CONSTS --> EXEC

    EXEC --> ARITH
    EXEC --> COMPARE
    EXEC --> CONTROL
    EXEC --> DATA
    EXEC --> COLL
```

### Opcode Categories

```mermaid
mindmap
    root((VM Opcodes))
        Stack Operations
            OpPush
            OpPop
        Arithmetic
            OpAdd
            OpSub
            OpMul
            OpDiv
            OpNeg
        Comparison
            OpEq
            OpNe
            OpLt
            OpGt
            OpLe
            OpGe
        Logic
            OpAnd
            OpOr
            OpNot
        Variables
            OpLoadVar
            OpStoreVar
        Control Flow
            OpJump
            OpJumpIfFalse
            OpJumpIfTrue
            OpReturn
            OpHalt
        Collections
            OpBuildArray
            OpBuildObject
            OpGetField
            OpGetIndex
        Iteration
            OpGetIter
            OpIterNext
            OpIterHasNext
        Functions
            OpCall
        HTTP
            OpHttpReturn
        WebSocket
            OpWsSend
            OpWsBroadcast
            OpWsJoinRoom
            OpWsLeaveRoom
```

---

## Server Architecture

```mermaid
graph TB
    subgraph "HTTP Layer"
        LISTENER["TCP Listener\n:8080"]
        MUX["HTTP Multiplexer"]
    end

    subgraph "Router"
        ROUTES["Route Registry"]
        MATCHER["Pattern Matcher"]
        PARAMS["Param Extractor"]
    end

    subgraph "Middleware Stack"
        AUTH["Authentication\nJWT/API Key"]
        RATE["Rate Limiter\n100 req/min"]
        CORS["CORS Handler"]
        LOG["Request Logger"]
        TRACE["Tracer"]
        METRICS_MW["Metrics Collector"]
    end

    subgraph "Handler"
        PARSE["Body Parser"]
        VALIDATE["Input Validator"]
        EXEC["Route Executor"]
        RESP["Response Builder"]
    end

    subgraph "Execution Engine"
        INTERP_E["Interpreter"]
        VM_E["Virtual Machine"]
    end

    LISTENER --> MUX
    MUX --> ROUTES
    ROUTES --> MATCHER
    MATCHER --> PARAMS

    PARAMS --> AUTH
    AUTH --> RATE
    RATE --> CORS
    CORS --> LOG
    LOG --> TRACE
    TRACE --> METRICS_MW

    METRICS_MW --> PARSE
    PARSE --> VALIDATE
    VALIDATE --> EXEC
    EXEC --> INTERP_E
    EXEC --> VM_E
    INTERP_E --> RESP
    VM_E --> RESP
```

---

## Database Integration

```mermaid
sequenceDiagram
    participant R as Route Handler
    participant I as Interpreter
    participant D as Database Layer
    participant P as Connection Pool
    participant PG as PostgreSQL

    R->>I: Execute route with DB injection
    I->>I: Evaluate db.query()
    I->>D: Query request

    D->>D: SQL injection check
    D->>D: Parameter binding
    D->>P: Get connection
    P->>P: Check pool

    alt Connection available
        P->>PG: Execute query
    else Pool exhausted
        P->>P: Wait for connection
        P->>PG: Execute query
    end

    PG-->>P: Result rows
    P-->>D: Return connection
    D->>D: Map to types
    D-->>I: Typed result
    I-->>R: Value
```

### Database Layer Structure

```mermaid
classDiagram
    class Database {
        <<interface>>
        +Query(sql, args) Rows
        +QueryRow(sql, args) Row
        +Exec(sql, args) Result
        +Begin() Tx
        +Close()
    }

    class PostgresDB {
        +pool *pgxpool.Pool
        +config PoolConfig
        +Query()
        +QueryRow()
        +Exec()
    }

    class ORM {
        +db Database
        +models map
        +Get(id) Model
        +Create(model) Model
        +Update(model) Model
        +Delete(id)
        +Where(cond) Query
    }

    class Transaction {
        +tx pgx.Tx
        +Commit()
        +Rollback()
        +Query()
    }

    Database <|-- PostgresDB
    PostgresDB --> ORM
    PostgresDB --> Transaction
```

---

## WebSocket Architecture

```mermaid
graph TB
    subgraph "WebSocket Server"
        HUB["Hub"]
        UPGRADE["HTTP Upgrader"]
    end

    subgraph "Connection Management"
        CONN["Connection"]
        READ["Read Pump"]
        WRITE["Write Pump"]
    end

    subgraph "Room System"
        ROOMS["Room Registry"]
        ROOM1["Room: chat"]
        ROOM2["Room: notifications"]
    end

    subgraph "Message Handling"
        MSG["Message Router"]
        VM_H["VM Handler"]
        BROADCAST["Broadcaster"]
    end

    UPGRADE -->|"Upgrade"| CONN
    CONN --> READ
    CONN --> WRITE
    CONN --> HUB

    HUB --> ROOMS
    ROOMS --> ROOM1
    ROOMS --> ROOM2

    READ --> MSG
    MSG --> VM_H
    VM_H -->|"OpWsSend"| WRITE
    VM_H -->|"OpWsBroadcast"| BROADCAST
    BROADCAST --> ROOM1
    BROADCAST --> ROOM2
```

### WebSocket Message Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant S as WS Server
    participant H as Hub
    participant R as Room
    participant VM as VM Handler

    C->>S: Connect (Upgrade)
    S->>H: Register client
    H->>H: Add to connections

    C->>S: Join room "chat"
    S->>H: JoinRoom(client, "chat")
    H->>R: Add client

    C->>S: Send message
    S->>VM: Execute handler
    VM->>VM: Process message

    alt Broadcast to room
        VM->>H: Broadcast("chat", msg)
        H->>R: Get clients
        R-->>H: Client list
        H->>C: Send to all
    else Direct reply
        VM->>S: Send(client, msg)
        S->>C: WebSocket frame
    end
```

---

## JIT Compilation

```mermaid
flowchart TD
    subgraph "Profiler"
        PROF["Execution Profiler"]
        COUNT["Call Counter"]
        TYPES["Type Recorder"]
    end

    subgraph "Hot Path Detection"
        THRESH["Threshold Check\n> 100 calls"]
        HOT["Hot Path Identified"]
    end

    subgraph "Specialization"
        SPEC["Type Specializer"]
        INLINE["Function Inliner"]
        FOLD["Constant Folder"]
    end

    subgraph "Recompilation"
        TIER["Tier Upgrade"]
        CACHE["Code Cache"]
        REPLACE["Replace Bytecode"]
    end

    PROF --> COUNT
    PROF --> TYPES
    COUNT --> THRESH
    THRESH -->|"Hot"| HOT
    HOT --> SPEC
    TYPES --> SPEC
    SPEC --> INLINE
    INLINE --> FOLD
    FOLD --> TIER
    TIER --> CACHE
    CACHE --> REPLACE
```

### JIT Tiers

```mermaid
graph LR
    subgraph "Tier 0: Interpreted"
        T0["Baseline\nNo optimization"]
    end

    subgraph "Tier 1: Basic JIT"
        T1["Simple optimizations\nConstant folding"]
    end

    subgraph "Tier 2: Optimized"
        T2["Type specialization\nInlining"]
    end

    subgraph "Tier 3: Aggressive"
        T3["Dead code elimination\nLoop optimization"]
    end

    T0 -->|"100 calls"| T1
    T1 -->|"1000 calls"| T2
    T2 -->|"10000 calls"| T3
```

---

## Binary Format

```mermaid
packet-beta
    0-31: "Magic: 'GLYP' (4 bytes)"
    32-39: "Version (1 byte)"
    40-47: "Flags (1 byte)"
    48-63: "Item Count (2 bytes)"
    64-95: "Constants Section"
    96-159: "Bytecode Section"
    160-191: "CRC32 Checksum (4 bytes)"
```

### Binary Structure Detail

```mermaid
graph TD
    subgraph "File Header"
        MAGIC["Magic: GLYP"]
        VER["Version: 1"]
        FLAGS["Flags: Debug/Optimized"]
        COUNT["Item Count"]
    end

    subgraph "Constants Pool"
        STRINGS["String Constants"]
        NUMS["Numeric Constants"]
        FUNCS["Function References"]
    end

    subgraph "Bytecode Sections"
        ROUTES["Route Bytecode"]
        HANDLERS["Function Bytecode"]
        TYPES["Type Definitions"]
    end

    subgraph "Footer"
        CRC["CRC32 Checksum"]
    end

    MAGIC --> VER
    VER --> FLAGS
    FLAGS --> COUNT
    COUNT --> STRINGS
    STRINGS --> NUMS
    NUMS --> FUNCS
    FUNCS --> ROUTES
    ROUTES --> HANDLERS
    HANDLERS --> TYPES
    TYPES --> CRC
```

---

## Observability Stack

```mermaid
graph TB
    subgraph "Application"
        APP["GlyphLang Server"]
    end

    subgraph "Logging"
        LOG["Structured Logger"]
        STDOUT["stdout"]
        FILE["Log File"]
    end

    subgraph "Metrics"
        PROM["Prometheus Client"]
        REQ_COUNT["request_count"]
        REQ_LATENCY["request_latency"]
        ERROR_RATE["error_rate"]
    end

    subgraph "Tracing"
        OTEL["OpenTelemetry"]
        SPANS["Span Collector"]
        JAEGER["Jaeger/OTLP"]
    end

    subgraph "External Systems"
        PROM_SRV["Prometheus Server"]
        GRAF["Grafana"]
        JAEGER_UI["Jaeger UI"]
    end

    APP --> LOG
    LOG --> STDOUT
    LOG --> FILE

    APP --> PROM
    PROM --> REQ_COUNT
    PROM --> REQ_LATENCY
    PROM --> ERROR_RATE

    APP --> OTEL
    OTEL --> SPANS
    SPANS --> JAEGER

    REQ_COUNT --> PROM_SRV
    PROM_SRV --> GRAF
    JAEGER --> JAEGER_UI
```

---

## Security Architecture

```mermaid
flowchart TD
    subgraph "Input Layer"
        REQ["HTTP Request"]
        BODY["Request Body"]
        PARAMS["URL Parameters"]
    end

    subgraph "Security Checks"
        SQL["SQL Injection\nDetector"]
        XSS["XSS\nDetector"]
        VAL["Input\nValidator"]
    end

    subgraph "Authentication"
        JWT["JWT\nValidator"]
        API["API Key\nChecker"]
        RATE["Rate\nLimiter"]
    end

    subgraph "Safe Execution"
        SAFE["Sanitized\nContext"]
        EXEC["Route\nExecution"]
    end

    REQ --> JWT
    REQ --> API
    JWT --> RATE
    API --> RATE

    BODY --> SQL
    BODY --> XSS
    PARAMS --> VAL

    SQL -->|"Safe"| SAFE
    XSS -->|"Safe"| SAFE
    VAL -->|"Valid"| SAFE
    RATE -->|"Allowed"| SAFE

    SQL -->|"Threat"| BLOCK["Block & Log"]
    XSS -->|"Threat"| BLOCK

    SAFE --> EXEC
```

---

## Development Workflow

```mermaid
flowchart LR
    subgraph "Development"
        CODE["Write .glyph"]
        SAVE["Save File"]
    end

    subgraph "Hot Reload"
        WATCH["File Watcher"]
        PARSE["Re-parse"]
        RELOAD["Hot Reload"]
    end

    subgraph "LSP"
        DIAG["Diagnostics"]
        COMP["Completions"]
        HOVER["Hover Info"]
    end

    subgraph "Testing"
        RUN["glyph run"]
        TEST["Integration Tests"]
    end

    subgraph "Production"
        BUILD["glyph compile -O3"]
        DEPLOY["Deploy .glyphc"]
    end

    CODE --> SAVE
    SAVE --> WATCH
    WATCH --> PARSE
    PARSE --> RELOAD
    PARSE --> DIAG
    PARSE --> COMP
    PARSE --> HOVER

    CODE --> RUN
    RUN --> TEST
    TEST --> BUILD
    BUILD --> DEPLOY
```

---

## Summary

GlyphLang is a complete backend language system with:

| Component | Purpose |
|-----------|---------|
| **Parser** | Tokenization and AST generation |
| **Type Checker** | Static type validation |
| **Compiler** | Bytecode generation with optimization |
| **VM** | Stack-based bytecode execution |
| **Interpreter** | Tree-walking execution |
| **Server** | HTTP/WebSocket request handling |
| **JIT** | Runtime optimization |
| **Security** | SQL injection and XSS protection |
| **Observability** | Metrics, tracing, and logging |

Performance characteristics:
- **Compilation**: ~867ns (sub-microsecond)
- **Execution**: 2.95-37.6ns per operation
- **Bytecode compression**: 94-99%
- **Test coverage**: 1696+ tests passing
