# Interpreter Quick Start Guide

## Installation

The interpreter is part of the `github.com/glyphlang/glyph/pkg/interpreter` package.

```bash
# Run tests
go test ./pkg/interpreter/... -v

# Run with coverage
go test ./pkg/interpreter/... -cover

# Run examples
go test ./pkg/interpreter/... -run Example -v
```

## Basic Usage

### 1. Create an Interpreter

```go
import "github.com/glyphlang/glyph/pkg/interpreter"

interp := interpreter.NewInterpreter()
```

### 2. Execute a Simple Route

```go
route := &interpreter.Route{
    Path:   "/hello",
    Method: interpreter.Get,
    Body: []interpreter.Statement{
        interpreter.ReturnStatement{
            Value: interpreter.LiteralExpr{
                Value: interpreter.StringLiteral{Value: "Hello!"},
            },
        },
    },
}

result, err := interp.ExecuteRouteSimple(route, nil)
// result: "Hello!"
```

### 3. Route with Path Parameters

```go
route := &interpreter.Route{
    Path:   "/greet/:name",
    Method: interpreter.Get,
    Body: []interpreter.Statement{
        interpreter.ReturnStatement{
            Value: interpreter.VariableExpr{Name: "name"},
        },
    },
}

params := map[string]string{"name": "Alice"}
result, err := interp.ExecuteRouteSimple(route, params)
// result: "Alice"
```

### 4. Arithmetic Operations

```go
// Calculate: 10 + 20
expr := interpreter.BinaryOpExpr{
    Op:    interpreter.Add,
    Left:  interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
    Right: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 20}},
}

env := interpreter.NewEnvironment()
result, err := interp.EvaluateExpression(expr, env)
// result: int64(30)
```

### 5. Load Module with Functions

```go
// Define: fn double(x: int) -> int { return x * 2 }
fn := interpreter.Function{
    Name: "double",
    Params: []interpreter.Field{
        {Name: "x", TypeAnnotation: interpreter.IntType{}, Required: true},
    },
    ReturnType: interpreter.IntType{},
    Body: []interpreter.Statement{
        interpreter.ReturnStatement{
            Value: interpreter.BinaryOpExpr{
                Op:    interpreter.Mul,
                Left:  interpreter.VariableExpr{Name: "x"},
                Right: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
            },
        },
    },
}

module := interpreter.Module{
    Items: []interpreter.Item{fn},
}

err := interp.LoadModule(module)

// Now call the function
env := interpreter.NewEnvironment()
env.Define("double", fn)

callExpr := interpreter.FunctionCallExpr{
    Name: "double",
    Args: []interpreter.Expr{
        interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 21}},
    },
}

result, err := interp.EvaluateExpression(callExpr, env)
// result: int64(42)
```

## Common Patterns

### String Concatenation

```go
expr := interpreter.BinaryOpExpr{
    Op:    interpreter.Add,
    Left:  interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "Hello, "}},
    Right: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "World!"}},
}
// result: "Hello, World!"
```

### Variable Assignment

```go
stmt := interpreter.AssignStatement{
    Target: "x",
    Value:  interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
}

env := interpreter.NewEnvironment()
_, err := interp.ExecuteStatement(stmt, env)

value, _ := env.Get("x")
// value: int64(42)
```

### If/Else Logic

```go
stmt := interpreter.IfStatement{
    Condition: interpreter.BinaryOpExpr{
        Op:    interpreter.Gt,
        Left:  interpreter.VariableExpr{Name: "age"},
        Right: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 18}},
    },
    ThenBlock: []interpreter.Statement{
        interpreter.ReturnStatement{
            Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "adult"}},
        },
    },
    ElseBlock: []interpreter.Statement{
        interpreter.ReturnStatement{
            Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "minor"}},
        },
    },
}

env := interpreter.NewEnvironment()
env.Define("age", int64(25))
result, err := interp.ExecuteStatement(stmt, env)
// result: "adult"
```

### Field Access

```go
// Access user.name
expr := interpreter.FieldAccessExpr{
    Object: interpreter.VariableExpr{Name: "user"},
    Field:  "name",
}

env := interpreter.NewEnvironment()
env.Define("user", map[string]interface{}{
    "name": "Alice",
    "age":  int64(30),
})

result, err := interp.EvaluateExpression(expr, env)
// result: "Alice"
```

## AST Node Reference

### Literals

```go
// Integer
interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}}

// String
interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "hello"}}

// Boolean
interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}}

// Float
interpreter.LiteralExpr{Value: interpreter.FloatLiteral{Value: 3.14}}
```

### Variables

```go
// Reference
interpreter.VariableExpr{Name: "myVar"}
```

### Binary Operations

```go
// Operators: Add, Sub, Mul, Div, Eq, Ne, Lt, Le, Gt, Ge
interpreter.BinaryOpExpr{
    Op:    interpreter.Add,
    Left:  leftExpr,
    Right: rightExpr,
}
```

### Statements

```go
// Assignment
interpreter.AssignStatement{
    Target: "varName",
    Value:  expr,
}

// Return
interpreter.ReturnStatement{
    Value: expr,
}

// If/Else
interpreter.IfStatement{
    Condition: condExpr,
    ThenBlock: []interpreter.Statement{...},
    ElseBlock: []interpreter.Statement{...},
}

// Database Query (mocked)
interpreter.DbQueryStatement{
    Var:    "result",
    Query:  "SELECT * FROM users",
    Params: []interpreter.Expr{...},
}
```

### Routes

```go
interpreter.Route{
    Path:       "/api/users/:id",
    Method:     interpreter.Get,
    ReturnType: interpreter.NamedType{Name: "User"},
    Auth:       &interpreter.AuthConfig{AuthType: "jwt", Required: true},
    RateLimit:  &interpreter.RateLimit{Requests: 100, Window: "min"},
    Body:       []interpreter.Statement{...},
}
```

### Functions

```go
interpreter.Function{
    Name: "functionName",
    Params: []interpreter.Field{
        {Name: "param1", TypeAnnotation: interpreter.IntType{}, Required: true},
    },
    ReturnType: interpreter.IntType{},
    Body:       []interpreter.Statement{...},
}
```

### Type Definitions

```go
interpreter.TypeDef{
    Name: "User",
    Fields: []interpreter.Field{
        {Name: "id", TypeAnnotation: interpreter.IntType{}, Required: true},
        {Name: "name", TypeAnnotation: interpreter.StringType{}, Required: true},
    },
}
```

## Error Handling

The interpreter returns errors for:

- Undefined variables
- Type mismatches (e.g., adding int and string)
- Division by zero
- Missing function definitions
- Path parameter mismatches
- Field access on non-objects

```go
result, err := interp.EvaluateExpression(expr, env)
if err != nil {
    // Handle error
    log.Printf("Evaluation error: %v", err)
}
```

## Testing

### Writing Tests

```go
func TestMyRoute(t *testing.T) {
    interp := interpreter.NewInterpreter()

    route := &interpreter.Route{
        // ... define route
    }

    result, err := interp.ExecuteRouteSimple(route, params)

    require.NoError(t, err)
    assert.Equal(t, expectedValue, result)
}
```

### Running Tests

```bash
# All tests
go test ./pkg/interpreter/...

# Verbose output
go test ./pkg/interpreter/... -v

# Specific test
go test ./pkg/interpreter/... -run TestEnvironment_DefineAndGet

# Coverage
go test ./pkg/interpreter/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Best Practices

1. **Always check errors**: The interpreter returns descriptive errors
2. **Use environments for scope**: Create child environments for nested scopes
3. **Define before use**: Variables must be defined before they can be accessed
4. **Type safety**: Ensure operands match (e.g., don't add int and string)
5. **Return handling**: Return statements throw a special error - handle it properly

## Integration Example

```go
// 1. Create interpreter
interp := interpreter.NewInterpreter()

// 2. Load module (from parser)
module := parser.ParseFile("main.abc")
err := interp.LoadModule(module)

// 3. Find route
route := findRouteByPath(module, "/api/users/:id")

// 4. Execute with HTTP request
request := &interpreter.Request{
    Path:   "/api/users/123",
    Method: "GET",
    Params: map[string]string{"id": "123"},
    Body:   nil,
}

response, err := interp.ExecuteRoute(route, request)

// 5. Return HTTP response
http.StatusCode = response.StatusCode
json.Marshal(response.Body)
```

## Performance Tips

- Reuse interpreter instances (they're thread-safe for read operations)
- Cache parsed modules to avoid re-parsing
- Use `ExecuteRouteSimple` for testing (lighter than full `ExecuteRoute`)
- Profile hot paths to identify bottlenecks
- Consider implementing a bytecode VM for production use

## Troubleshooting

### "undefined variable" error
- Ensure the variable is defined before use
- Check if you're in the correct scope

### "cannot add X and Y" error
- Verify operand types match
- String concatenation requires both operands to be strings

### "division by zero" error
- Add conditional check before division
- Validate input parameters

### "not a function" error
- Ensure function is defined in module or environment
- Check function name spelling

### "field not found" error
- Verify object has the field
- Ensure field name is correct

## FAQ

**Q: Can I execute Glyph source code directly?**
A: No, you need to parse it to AST first. This interpreter works with AST nodes.

**Q: Is the interpreter thread-safe?**
A: Read operations are safe, but concurrent route executions should use separate interpreter instances or be synchronized.

**Q: How do I add custom built-in functions?**
A: Modify `evaluateFunctionCall` in `evaluator.go` to add more built-in functions.

**Q: Can I use this in production?**
A: For MVP/prototyping yes. For high-performance production, consider bytecode compilation.

**Q: How do I debug execution?**
A: Add logging in `EvaluateExpression` and `ExecuteStatement` to trace execution flow.

## Additional Resources

- Full documentation: `README.md`
- Implementation details: `IMPLEMENTATION_REPORT.md`
- Example tests: `examples_test.go`
- Unit tests: `interpreter_test.go`
- AST definitions: `ast.go`

## Support

For issues or questions:
1. Check the test files for examples
2. Review the implementation report
3. Examine the comprehensive documentation
4. Look at the AST type definitions
