# Glyph Interpreter

A Go-based AST interpreter for the Glyph language.

## Overview

This interpreter executes Glyph AST nodes directly without compilation to bytecode. It supports:

- Expression evaluation (literals, variables, binary operations, field access, function calls)
- Statement execution (variable assignment, return, if statements, database queries)
- Route execution with path parameter extraction
- Environment/scope management
- User-defined functions

## Files

- `ast.go` - AST type definitions matching `glyph-core/src/ast.rs`
- `environment.go` - Variable scope management
- `evaluator.go` - Expression evaluation logic
- `executor.go` - Statement execution logic
- `interpreter.go` - Main interpreter struct and route execution
- `interpreter_test.go` - Comprehensive unit tests

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/glyphlang/glyph/pkg/interpreter"
)

func main() {
    interp := interpreter.NewInterpreter()

    // Define a simple route: @ route /hello
    route := &interpreter.Route{
        Path:   "/hello",
        Method: interpreter.Get,
        Body: []interpreter.Statement{
            interpreter.ReturnStatement{
                Value: interpreter.LiteralExpr{
                    Value: interpreter.StringLiteral{
                        Value: "Hello, World!",
                    },
                },
            },
        },
    }

    // Execute the route
    result, err := interp.ExecuteRouteSimple(route, nil)
    if err != nil {
        panic(err)
    }

    fmt.Println(result) // Output: Hello, World!
}
```

### Route with Path Parameters

```go
// @ route /greet/:name
route := &interpreter.Route{
    Path:   "/greet/:name",
    Method: interpreter.Get,
    Body: []interpreter.Statement{
        interpreter.AssignStatement{
            Target: "message",
            Value: interpreter.BinaryOpExpr{
                Op:    interpreter.Add,
                Left:  interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "Hello, "}},
                Right: interpreter.VariableExpr{Name: "name"},
            },
        },
        interpreter.ReturnStatement{
            Value: interpreter.VariableExpr{Name: "message"},
        },
    },
}

// Execute with path parameters
params := map[string]string{"name": "Alice"}
result, err := interp.ExecuteRouteSimple(route, params)
// result: "Hello, Alice"
```

### Loading Modules

```go
// Load a module with type definitions and functions
module := interpreter.Module{
    Items: []interpreter.Item{
        interpreter.TypeDef{
            Name: "User",
            Fields: []interpreter.Field{
                {Name: "id", TypeAnnotation: interpreter.IntType{}, Required: true},
                {Name: "name", TypeAnnotation: interpreter.StringType{}, Required: true},
            },
        },
        interpreter.Function{
            Name: "add",
            Params: []interpreter.Field{
                {Name: "a", TypeAnnotation: interpreter.IntType{}, Required: true},
                {Name: "b", TypeAnnotation: interpreter.IntType{}, Required: true},
            },
            ReturnType: interpreter.IntType{},
            Body: []interpreter.Statement{
                interpreter.ReturnStatement{
                    Value: interpreter.BinaryOpExpr{
                        Op:    interpreter.Add,
                        Left:  interpreter.VariableExpr{Name: "a"},
                        Right: interpreter.VariableExpr{Name: "b"},
                    },
                },
            },
        },
    },
}

err := interp.LoadModule(module)
```

## Running Tests

```bash
go test ./pkg/interpreter/... -v
```

## Test Coverage

The test suite includes:

### Environment Tests
- Variable definition and retrieval
- Undefined variable errors
- Child scope inheritance
- Variable shadowing
- Variable updates

### Expression Evaluation Tests
- Literal evaluation (int, string, bool, float)
- Variable references
- Binary operations:
  - Arithmetic: `+`, `-`, `*`, `/`
  - Comparison: `==`, `!=`, `<`, `<=`, `>`, `>=`
- String concatenation
- Division by zero handling
- Field access on objects
- Built-in function calls (`time.now`, `now`)
- User-defined function calls

### Statement Execution Tests
- Variable assignment
- Return statements
- If/else statements
- Database query statements (mocked)

### Route Execution Tests
- Simple routes
- Routes with path parameters
- Routes with multiple statements
- Complex nested expressions
- Full HTTP request/response handling

### Module Loading Tests
- Function definitions
- Type definitions
- Module validation

## Features Implemented

### Expression Evaluation
- [x] Literal values (int, string, bool, float)
- [x] Variable references
- [x] Binary operations (arithmetic and comparison)
- [x] String concatenation
- [x] Field access on maps
- [x] Function calls (built-in and user-defined)

### Statement Execution
- [x] Variable assignment
- [x] Return statements
- [x] If/else conditionals
- [x] Database queries (mocked)

### Interpreter Features
- [x] Environment/scope management
- [x] Path parameter extraction
- [x] Route execution
- [x] Module loading
- [x] User-defined functions
- [x] Type definitions

### Testing
- [x] 30+ unit tests
- [x] High test coverage
- [x] Edge case handling
- [x] Error validation

## Example: Hello World Route

This interpreter can execute the hello world example from `examples/hello-world/main.glyph`:

```glyph
# @ route /hello
#   > {text: "Hello, World!", timestamp: 1234567890}
```

Equivalent AST:

```go
route := &interpreter.Route{
    Path:   "/hello",
    Method: interpreter.Get,
    Body: []interpreter.Statement{
        interpreter.ReturnStatement{
            Value: interpreter.MapLiteral{ // (simplified for example)
                "text":      "Hello, World!",
                "timestamp": int64(1234567890),
            },
        },
    },
}
```

## Example: Greeting Route with Parameter

```glyph
# @ route /greet/:name -> Message
#   $ message = {
#     text: "Hello, " + name + "!",
#     timestamp: time.now()
#   }
#   > message
```

Equivalent AST:

```go
route := &interpreter.Route{
    Path:       "/greet/:name",
    Method:     interpreter.Get,
    ReturnType: interpreter.NamedType{Name: "Message"},
    Body: []interpreter.Statement{
        interpreter.AssignStatement{
            Target: "message",
            Value: interpreter.MapLiteral{
                "text": interpreter.BinaryOpExpr{
                    Op: interpreter.Add,
                    Left: interpreter.BinaryOpExpr{
                        Op:    interpreter.Add,
                        Left:  interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "Hello, "}},
                        Right: interpreter.VariableExpr{Name: "name"},
                    },
                    Right: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "!"}},
                },
                "timestamp": interpreter.FunctionCallExpr{
                    Name: "time.now",
                    Args: []interpreter.Expr{},
                },
            },
        },
        interpreter.ReturnStatement{
            Value: interpreter.VariableExpr{Name: "message"},
        },
    },
}
```

## Architecture

### Component Overview

```
interpreter/
├── ast.go           # AST node definitions
├── environment.go   # Scope management
├── evaluator.go     # Expression evaluation
├── executor.go      # Statement execution
└── interpreter.go   # Main interpreter logic
```

### Execution Flow

1. **Load Module**: Parse and load type definitions, functions, and routes
2. **Execute Route**:
   - Create route environment
   - Extract path parameters
   - Add parameters to environment
   - Execute statements in order
   - Handle return values
   - Build response
3. **Evaluate Expression**:
   - Recursively evaluate sub-expressions
   - Apply operators
   - Return computed value
4. **Execute Statement**:
   - Modify environment (assignments)
   - Control flow (if/else, return)
   - External operations (database queries)

## Future Enhancements

- [ ] Map/object literal expressions
- [ ] Array support
- [ ] Database integration (currently mocked)
- [ ] HTTP client for external API calls
- [ ] Middleware support
- [ ] Error handling improvements
- [ ] Type checking at runtime
- [ ] Debugger support
- [ ] Performance optimizations

## Integration

This interpreter is designed to work with:
- **Parser** (pending): Converts Glyph source code to AST
- **Runtime**: Provides HTTP server to route requests to interpreter
- **Cache**: Stores compiled/parsed modules
- **Database**: Executes database queries from routes

## Notes

- Database queries are currently mocked and return placeholder values
- Built-in functions are limited to `time.now()` and `now()`
- The interpreter assumes well-formed AST input (validation should happen in parser)
- Path parameter extraction is simplistic (no regex patterns yet)
- Error handling returns descriptive messages for debugging
