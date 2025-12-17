# Enhanced Error Messages for GLYPHLANG

This package provides comprehensive error handling with enhanced error messages, intelligent suggestions, and helpful debugging context.

## Features

### 1. Enhanced "Did you mean?" Suggestions

The error system uses advanced fuzzy matching algorithms to suggest corrections:

```go
// Variable name suggestions
availableVars := []string{"username", "userId", "userEmail"}
suggestion := GetVariableSuggestion("usrname", availableVars)
// Returns: "Did you mean 'username'? Or make sure to define the variable with '$ usrname = value'"

// Function name suggestions
availableFuncs := []string{"println", "print", "printf"}
suggestion := GetFunctionSuggestion("prnt", availableFuncs)
// Returns: "Did you mean 'print' or 'println'?"

// Type name suggestions
availableTypes := []string{"User", "Product", "Order"}
suggestion := GetTypeSuggestion("Usr", availableTypes)
// Returns: "Did you mean 'User'?"

// Route path suggestions
availableRoutes := []string{"/users", "/users/:id", "/products"}
suggestion := GetRouteSuggestion("/usr", availableRoutes)
// Returns: "Did you mean '/users'?"
```

**Common typo corrections:**
- `fucntion` → `function`
- `retrun` → `return`
- `lenght` → `length`
- `treu` → `true`
- `flase` → `false`
- And many more!

### 2. Code Snippets with Error Context

Errors show multiple lines of context with exact error position:

```go
source := `@ route /users GET -> array {
    $ userCount: int = "invalid"
    return getUsers()
}`

err := NewTypeError(
    "Type mismatch in variable declaration",
    2,
    24,
    ExtractSourceSnippet(source, 2),
    GetTypeMismatchSuggestion("int", "string", "variable declaration"),
).WithTypes("int", "string").
  WithContext("in route /users GET").
  WithFixedLine(`    $ userCount: int = 0`)
```

**Output:**
```
Type Error in route /users GET at line 2, column 24

     1 | @ route /users GET -> array {
     2 | $ userCount: int = "invalid"
       |                        ^ error here
     2 | $ userCount: int = 0 (suggested fix)
     3 | return getUsers()

Type mismatch in variable declaration

Expected: int
Actual:   string

Suggestion: Convert the string to an integer using parseInt() or ensure the value is numeric
```

### 3. Syntax Error Hints

Automatic detection of common syntax errors:

```go
// Missing brackets/braces/parentheses
DetectMissingBracket(source, line, column)
// Returns: "Missing 1 closing brace '}'" or "Unexpected closing bracket ']'"

// Unclosed strings
DetectUnclosedString(source, line)
// Returns: "Unclosed string literal (missing closing \")"

// Common syntax patterns
DetectCommonSyntaxErrors(source, line, errorMsg)
// Detects patterns like missing closing braces, unclosed strings, etc.
```

### 4. Type Error Improvements

Enhanced type error messages with clear expected vs actual types:

```go
err := NewTypeError(
    "Type mismatch in assignment",
    2,
    21,
    snippet,
    GetTypeMismatchSuggestion("int", "string", "variable assignment"),
).WithTypes("int", "string")
```

**Type-specific suggestions:**
- `int` ← `string`: "Convert the string to an integer using parseInt()"
- `string` ← `int`: "Convert the integer to a string using toString()"
- `bool` ← `int`: "Use a boolean value (true or false) or convert using a comparison"
- `float` ← `int`: "The integer will be automatically converted to float"
- `array` ← other: "Wrap the value in square brackets [] to create an array"

### 5. Runtime Error Context

Runtime errors include full execution context:

```go
err := NewRuntimeError("Division by zero").
    WithRoute("/calculate POST").
    WithExpression("result = numerator / denominator").
    WithSuggestion(GetRuntimeSuggestion("division_by_zero", nil)).
    WithScope(map[string]interface{}{
        "numerator":   10,
        "denominator": 0,
    }).
    WithStackFrame("calculateResult", "/calculate POST", 5).
    WithStackFrame("handleRequest", "main.glybc", 12)
```

**Output:**
```
Runtime Error

Division by zero

Route: /calculate POST

Expression:
  result = numerator / denominator

Variables in scope:
  numerator = 10 (int)
  denominator = 0 (int)

Stack trace:
  1. calculateResult at /calculate POST:5
  2. handleRequest at main.glybc:12

Suggestion: Add a check to ensure the divisor is not zero before dividing: if (divisor != 0) { ... }
```

**Runtime error types:**
- `division_by_zero`
- `null_reference`
- `index_out_of_bounds`
- `type_error`
- `undefined_property`
- `sql_injection`
- `xss`

## Usage Examples

### Compile Error with Suggestion

```go
err := NewCompileError(
    "Undefined variable 'totl'",
    3,
    7,
    ExtractSourceSnippet(source, 3),
    GetVariableSuggestion("totl", []string{"total", "count"}),
).WithContext("in route /users GET")

fmt.Println(err.FormatError(true)) // With colors
```

### Parse Error with Hint

```go
err := NewParseError(
    "Missing closing bracket in array literal",
    2,
    24,
    ExtractSourceSnippet(source, 2),
    "Add a closing bracket ']' to complete the array definition",
)

fmt.Println(err.FormatError(false)) // Without colors
```

### Type Error with Expected/Actual Types

```go
err := NewTypeError(
    "Type mismatch in function parameter",
    5,
    15,
    snippet,
    GetTypeMismatchSuggestion("string", "int", "function parameter"),
).WithTypes("string", "int").WithContext("in function calculateAge")
```

### Error Wrapping

```go
// Start with a simple error
baseErr := fmt.Errorf("invalid value")

// Add line information
err := WithLineInfo(baseErr, 2, 13, source)

// Add suggestion
err = WithSuggestion(err, "Check the value type")

// Add filename
err = WithFileName(err, "main.glybc")

// Format and display
fmt.Println(FormatError(err))
```

## Configuration

### Suggestion Configuration

```go
config := &SuggestionConfig{
    MaxSuggestions:      3,     // Maximum number of suggestions to return
    MaxDistance:         3,     // Maximum Levenshtein distance for matches
    MinSimilarityScore:  0.5,   // Minimum similarity score (0.0 - 1.0)
    ShowMultipleSuggestions: true, // Show multiple suggestions
}

results := FindBestSuggestions("cunt", candidates, config)
```

## Advanced Features

### Custom Suggestion Formatting

```go
results := FindBestSuggestions(target, candidates, config)
formatted := FormatSuggestions(results, true)
// Returns: "Did you mean 'count', 'counter', or 'total'?"
```

### Similarity Scoring

The similarity score takes into account:
- Levenshtein distance (edit distance)
- Common prefixes (first 3 characters)
- Common suffixes (last 2 characters)
- Substring matches
- Case-insensitive matches

### Identifier Validation

```go
if !IsValidIdentifier(name) {
    suggestion := SuggestValidIdentifier(name)
    // Returns helpful message about why it's invalid and how to fix it
}
```

## Testing

Run all tests:
```bash
go test ./pkg/errors/...
```

Run benchmarks:
```bash
go test ./pkg/errors/... -bench=. -benchmem
```

Run specific examples:
```bash
go test ./pkg/errors/... -v -run Example
```

## Performance

Benchmarks on AMD Ryzen 7 7800X3D:
- CompileError formatting: ~859 ns/op
- RuntimeError formatting: ~870 ns/op
- Levenshtein distance: ~230 ns/op
- Best suggestions: ~3964 ns/op
- Similarity score: ~84 ns/op
- Missing bracket detection: ~240 ns/op
- Unclosed string detection: ~118 ns/op

## Color Support

All error formatters support both colored and non-colored output:

```go
err.FormatError(true)  // With ANSI colors
err.FormatError(false) // Without colors (for logs, files, etc.)
```

## Integration

The error package integrates seamlessly with:
- **Parser** (`pkg/parser/errors.go`) - Parse-time error detection
- **Compiler** - Compile-time type checking and validation
- **Runtime** (`pkg/vm`) - Runtime error handling with full context
- **LSP** (`pkg/lsp`) - Language server for IDE integration

## Future Enhancements

Planned features:
- Machine learning-based suggestion ranking
- Context-aware error messages based on surrounding code
- Integration with external documentation
- Multi-language error messages
- Error recovery suggestions
