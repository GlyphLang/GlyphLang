# Glyph Binary Format Specification

**Version:** 1.0
**Purpose:** AI-optimized compact representation of Glyph programs

## Design Principles

1. **Maximum compactness** - Minimal bytes, no redundancy
2. **Zero ambiguity** - Direct AST representation
3. **Fast parsing** - Binary read, no text processing
4. **Lossless** - Perfect round-trip with human format

## Format Overview

```
┌──────────────────────────────────────────────┐
│ Glyph Binary Format (.glyphc)                   │
├──────────────────────────────────────────────┤
│ [Magic Number: 4 bytes] "Glyph"               │
│ [Version: 1 byte] 0x01                       │
│ [Flags: 1 byte] compression, etc.            │
│ [Item Count: 2 bytes] number of items        │
│ ├─ [Items: variable]                         │
│ │  ├─ Type Definitions                       │
│ │  ├─ Routes                                 │
│ │  └─ Functions                              │
│ └─ [Checksum: 4 bytes] CRC32                 │
└──────────────────────────────────────────────┘
```

## Binary Encoding

### Type Codes (1 byte)
```
0x00 = Module
0x01 = TypeDef
0x02 = Route
0x03 = Function

// Types
0x10 = Int
0x11 = String
0x12 = Bool
0x13 = Float
0x14 = Array
0x15 = Optional
0x16 = Named

// Expressions
0x20 = Literal
0x21 = Variable
0x22 = BinaryOp
0x23 = FieldAccess
0x24 = FunctionCall
0x25 = Object

// Statements
0x30 = Assign
0x31 = Return
0x32 = DbQuery
0x33 = If

// Binary Operators
0x40 = Add
0x41 = Sub
0x42 = Mul
0x43 = Div
0x44 = Eq
0x45 = Ne
0x46 = Lt
0x47 = Le
0x48 = Gt
0x49 = Ge

// HTTP Methods
0x50 = GET
0x51 = POST
0x52 = PUT
0x53 = DELETE
0x54 = PATCH
```

### String Encoding
```
[Length: 2 bytes] [UTF-8 bytes: variable]
```
- Max string length: 65,535 bytes
- Compact for short identifiers

### Example: Simple Route

**Human format (.glyph):**
```glyph
@ route /hello
  > {message: "Hello, World!"}
```

**Binary format (.glyphc) - Hex dump:**
```
41 49 42 43        // Magic: "Glyph"
01                 // Version: 1
00                 // Flags: none
01 00              // Item count: 1

02                 // Type: Route
00 06 2F 68 65 6C 6C 6F  // Path: "/hello" (length=6)
50                 // Method: GET
16 00 04 76 6F 69 64     // Return: Named("void")
00                 // No auth
00                 // No rate limit
01                 // Body count: 1

31                 // Statement: Return
25                 // Expr: Object
01 00              // Field count: 1
00 07 6D 65 73 73 61 67 65  // Field: "message"
20                 // Literal
11                 // String
00 0D 48 65 6C 6C 6F 2C 20 57 6F 72 6C 64 21  // "Hello, World!"

XX XX XX XX        // CRC32 checksum
```

**Size comparison:**
- Human format: ~52 bytes (text)
- Binary format: ~45 bytes (without compression)
- **Savings: ~13%** (more savings on larger files)

## Token Efficiency

**For AI models:**
- Human format: ~15-20 tokens
- Binary format: Treated as single "blob" token or 10-15 tokens
- **Token savings: ~30-50%**

## Compression

Optional zstd compression (flag 0x01):
- Further 40-60% reduction
- Fast decompression
- Total savings: ~60-70% vs human format

## Round-Trip Guarantee

```
.glyph → parser → AST → serializer → .glyphc
.glyphc → deserializer → AST → formatter → .glyph
```

- Comments preserved in metadata
- Formatting preserved in hints
- Perfect reconstruction

## Performance Targets

| Operation | Target | Actual |
|-----------|--------|--------|
| Serialize | < 0.1ms per route | TBD |
| Deserialize | < 0.1ms per route | TBD |
| Parse .glyph | < 1ms per route | ~0.5ms |
| Format .glyph | < 1ms per route | TBD |

## Usage

```bash
# Compile human to AI format
glyph compile hello.glyph -o hello.glyphc

# Decompile AI to human format
glyph decompile hello.glyphc -o hello.glyph

# Show disassembly only (no file output)
glyph decompile --disasm hello.glyphc

# Run directly from binary
glyph run hello.glyphc

# AI generates binary directly
curl ai-api.com/generate > api.glyphc
glyph run api.glyphc
```

## Decompiler

The decompiler (`pkg/decompiler`) provides full bytecode analysis:

**Supported Constant Types:**
- `0x00` - Null
- `0x01` - Int (8 bytes, little-endian)
- `0x02` - Float (8 bytes, IEEE 754)
- `0x03` - Bool (1 byte)
- `0x04` - String (4-byte length + UTF-8 data)

**Supported Opcodes (37 total):**
- Stack: PUSH, POP
- Arithmetic: ADD, SUB, MUL, DIV
- Comparison: EQ, NE, LT, GT, LE, GE
- Logic: AND, OR, NOT, NEG
- Variables: LOAD_VAR, STORE_VAR
- Control Flow: JUMP, JUMP_IF_FALSE, JUMP_IF_TRUE
- Iteration: GET_ITER, ITER_NEXT, ITER_HAS_NEXT, GET_INDEX
- Functions: CALL, RETURN
- Data: BUILD_OBJECT, BUILD_ARRAY, GET_FIELD
- HTTP: HTTP_RETURN
- WebSocket: WS_SEND, WS_BROADCAST, WS_BROADCAST_ROOM, WS_JOIN_ROOM, WS_LEAVE_ROOM, WS_CLOSE, WS_GET_ROOMS, WS_GET_CLIENTS, WS_GET_CONN_COUNT, WS_GET_UPTIME
- Control: HALT

**Output Formats:**
- `Format()` - Pseudo-source reconstruction with route signatures
- `FormatDisassembly()` - Detailed bytecode listing with constant pool and instruction comments

## Future Optimizations

1. **Variable-length integers** - Smaller numbers use fewer bytes
2. **String interning** - Deduplicate repeated strings
3. **Schema evolution** - Version-aware parsing
4. **Streaming format** - Incremental loading
5. **Index section** - Fast random access

## Metadata Section (Optional)

For preserving comments, formatting, etc.:
```
[Metadata Flag: 0x02]
[Metadata Length: 4 bytes]
[Comments: JSON array]
[Formatting hints: JSON object]
```

