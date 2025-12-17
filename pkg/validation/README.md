# Glyph Validation Package

The validation package provides comprehensive validation rule enforcement for the Glyph language. It includes built-in validators for common validation scenarios and a flexible framework for custom validation logic.

## Features

- **Email Validation**: RFC 5322 compliant email address validation
- **Length Validation**: String length constraints (min/max)
- **Range Validation**: Numeric range validation with type checking
- **Pattern Validation**: Regex pattern matching
- **Required Validation**: Check for nil/empty values
- **Custom Validation**: Define custom validation logic
- **Multiple Rules**: Apply multiple validation rules to a single field
- **Clear Error Messages**: Detailed, user-friendly error messages

## Installation

The validation package is part of the Glyph project and can be imported as:

```go
import "github.com/glyphlang/glyph/pkg/validation"
```

## Usage Examples

### Basic Validation Functions

#### Email Validation

```go
import "github.com/glyphlang/glyph/pkg/validation"

// Valid email
err := validation.ValidateEmail("user@example.com")
if err != nil {
    fmt.Println(err) // nil
}

// Invalid email
err = validation.ValidateEmail("invalid-email")
if err != nil {
    fmt.Println(err) // validation error: invalid email format
}

// Empty email
err = validation.ValidateEmail("")
if err != nil {
    fmt.Println(err) // validation error: email cannot be empty
}
```

#### Length Validation

```go
// Validate string length (min: 3, max: 20)
err := validation.ValidateLength("hello", 3, 20)
if err != nil {
    fmt.Println(err) // nil - valid length
}

// String too short
err = validation.ValidateLength("ab", 5, 20)
if err != nil {
    fmt.Println(err) // validation error: length must be at least 5 characters (got 2)
}

// String too long
err = validation.ValidateLength("this is a very long string", 3, 10)
if err != nil {
    fmt.Println(err) // validation error: length must be at most 10 characters (got 26)
}

// No minimum constraint (min = 0)
err = validation.ValidateLength("hi", 0, 10)
if err != nil {
    fmt.Println(err) // nil - no minimum enforced
}

// No maximum constraint (max = 0)
err = validation.ValidateLength("very long string", 5, 0)
if err != nil {
    fmt.Println(err) // nil - no maximum enforced
}
```

#### Range Validation

```go
// Validate numeric range (min: 1.0, max: 100.0)
err := validation.ValidateRange(50.0, 1.0, 100.0)
if err != nil {
    fmt.Println(err) // nil - within range
}

// Value below minimum
err = validation.ValidateRange(0.5, 1.0, 100.0)
if err != nil {
    fmt.Println(err) // validation error: value must be at least 1 (got 0.5)
}

// Value above maximum
err = validation.ValidateRange(150.0, 1.0, 100.0)
if err != nil {
    fmt.Println(err) // validation error: value must be at most 100 (got 150)
}

// Negative ranges
err = validation.ValidateRange(-2.0, -5.0, 5.0)
if err != nil {
    fmt.Println(err) // nil - within range
}
```

#### Pattern Validation

```go
// Validate phone number pattern
err := validation.ValidatePattern("123-456-7890", `^\d{3}-\d{3}-\d{4}$`)
if err != nil {
    fmt.Println(err) // nil - matches pattern
}

// Validate alphanumeric
err = validation.ValidatePattern("abc123", "^[a-z0-9]+$")
if err != nil {
    fmt.Println(err) // nil - matches pattern
}

// Invalid pattern
err = validation.ValidatePattern("ABC", "^[a-z]+$")
if err != nil {
    fmt.Println(err) // validation error: value does not match pattern '^[a-z]+$'
}

// UUID validation
uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
err = validation.ValidatePattern("550e8400-e29b-41d4-a716-446655440000", uuidPattern)
if err != nil {
    fmt.Println(err) // nil - valid UUID
}
```

#### Required Validation

```go
// Valid non-empty string
err := validation.ValidateRequired("hello")
if err != nil {
    fmt.Println(err) // nil
}

// Empty string
err = validation.ValidateRequired("")
if err != nil {
    fmt.Println(err) // validation error: value cannot be empty
}

// Nil value
err = validation.ValidateRequired(nil)
if err != nil {
    fmt.Println(err) // validation error: value is required
}

// Empty array
err = validation.ValidateRequired([]interface{}{})
if err != nil {
    fmt.Println(err) // validation error: array cannot be empty
}

// Empty map
err = validation.ValidateRequired(map[string]interface{}{})
if err != nil {
    fmt.Println(err) // validation error: object cannot be empty
}

// Numbers (including zero) are valid
err = validation.ValidateRequired(0)
if err != nil {
    fmt.Println(err) // nil - zero is a valid value
}

// Booleans (including false) are valid
err = validation.ValidateRequired(false)
if err != nil {
    fmt.Println(err) // nil - false is a valid value
}
```

### Using the Validator Struct

The `Validator` struct provides a more structured approach for validating multiple fields with different rules.

#### Creating a Validator

```go
import "github.com/glyphlang/glyph/pkg/validation"

// Create a new validator
v := validation.NewValidator()
```

#### Adding Rules

```go
// Add email validation rule
v.AddEmailRule("email")

// Add length validation rule (min: 3, max: 20)
v.AddLengthRule("username", 3, 20)

// Add range validation rule (min: 18.0, max: 100.0)
v.AddRangeRule("age", 18.0, 100.0)

// Add pattern validation rule
v.AddPatternRule("zipcode", `^\d{5}$`)

// Add required validation rule
v.AddRequiredRule("name")

// Add multiple rules to the same field
v.AddRequiredRule("email")
v.AddEmailRule("email")
```

#### Custom Validation Rules

```go
// Add custom validation rule
v.AddCustomRule("password", func(value interface{}) error {
    str, ok := value.(string)
    if !ok {
        return &validation.ValidationError{Message: "password must be a string"}
    }

    // Check minimum length
    if len(str) < 8 {
        return &validation.ValidationError{Message: "password must be at least 8 characters"}
    }

    // Check for uppercase letter
    hasUpper := false
    for _, c := range str {
        if c >= 'A' && c <= 'Z' {
            hasUpper = true
            break
        }
    }
    if !hasUpper {
        return &validation.ValidationError{Message: "password must contain at least one uppercase letter"}
    }

    return nil
})
```

#### Validating Data

```go
// Validate a map of data
data := map[string]interface{}{
    "email":    "user@example.com",
    "username": "johndoe",
    "age":      25.0,
    "name":     "John Doe",
    "zipcode":  "12345",
}

err := v.Validate(data)
if err != nil {
    fmt.Println(err) // nil - all validations passed
}

// Invalid data example
invalidData := map[string]interface{}{
    "email":    "invalid-email",
    "username": "ab", // too short
    "age":      15.0, // below minimum
}

err = v.Validate(invalidData)
if err != nil {
    // Multiple validation errors
    fmt.Println(err)
    // Output: validation error on field 'email': invalid email format;
    //         validation error on field 'username': length must be at least 3 characters (got 2);
    //         validation error on field 'age': value must be at least 18 (got 15);
    //         validation error on field 'name': field is required
}
```

#### Validating Individual Fields

```go
// Validate a single field
v := validation.NewValidator()
v.AddEmailRule("email")

err := v.ValidateField("email", "user@example.com")
if err != nil {
    fmt.Println(err) // nil - valid email
}

err = v.ValidateField("email", "invalid")
if err != nil {
    fmt.Println(err) // validation error on field 'email': invalid email format
}
```

#### Managing Rules

```go
// Check if a field has rules
if v.HasRules("email") {
    fmt.Println("Email field has validation rules")
}

// Get all rules for a field
rules := v.GetRules("email")
for _, rule := range rules {
    fmt.Printf("Rule type: %s\n", rule.RuleType)
}

// Clear all rules
v.ClearRules()
```

### Complete Example: User Registration

```go
package main

import (
    "fmt"
    "github.com/glyphlang/glyph/pkg/validation"
)

func main() {
    // Create validator
    v := validation.NewValidator()

    // Define validation rules
    v.AddRequiredRule("email")
    v.AddEmailRule("email")

    v.AddRequiredRule("username")
    v.AddLengthRule("username", 3, 20)
    v.AddPatternRule("username", "^[a-zA-Z0-9_]+$") // Alphanumeric and underscore only

    v.AddRequiredRule("password")
    v.AddCustomRule("password", func(value interface{}) error {
        str, ok := value.(string)
        if !ok {
            return &validation.ValidationError{Message: "password must be a string"}
        }
        if len(str) < 8 {
            return &validation.ValidationError{Message: "password must be at least 8 characters"}
        }
        return nil
    })

    v.AddRequiredRule("age")
    v.AddRangeRule("age", 18.0, 120.0)

    v.AddPatternRule("phone", `^\d{3}-\d{3}-\d{4}$`) // Optional field

    // Valid user data
    validUser := map[string]interface{}{
        "email":    "john.doe@example.com",
        "username": "johndoe_123",
        "password": "SecurePass123",
        "age":      25.0,
        "phone":    "555-123-4567",
    }

    err := v.Validate(validUser)
    if err != nil {
        fmt.Println("Validation failed:", err)
    } else {
        fmt.Println("User registration successful!")
    }

    // Invalid user data
    invalidUser := map[string]interface{}{
        "email":    "not-an-email",
        "username": "ab", // Too short
        "password": "short", // Too short
        "age":      15.0, // Under 18
        "phone":    "555-1234", // Invalid format
    }

    err = v.Validate(invalidUser)
    if err != nil {
        fmt.Println("Validation failed:", err)
        // Output: Multiple validation errors listing all issues
    }
}
```

## Validation Error Handling

### ValidationError

Single validation error with optional field information:

```go
err := validation.ValidateEmail("invalid")
if validationErr, ok := err.(*validation.ValidationError); ok {
    fmt.Printf("Field: %s, Message: %s\n", validationErr.Field, validationErr.Message)
}
```

### ValidationErrors

Multiple validation errors:

```go
v := validation.NewValidator()
v.AddRequiredRule("email")
v.AddEmailRule("email")
v.AddLengthRule("username", 3, 20)

data := map[string]interface{}{
    "username": "ab",
}

err := v.Validate(data)
if validationErrs, ok := err.(*validation.ValidationErrors); ok {
    fmt.Printf("Total errors: %d\n", len(validationErrs.Errors))
    for _, e := range validationErrs.Errors {
        fmt.Printf("- Field '%s': %s\n", e.Field, e.Message)
    }
}
```

## Performance

The validation package includes benchmark tests for performance evaluation:

```bash
go test -bench=. ./pkg/validation
```

Example benchmark results:
- ValidateEmail: ~500 ns/op
- ValidateLength: ~50 ns/op
- ValidateRange: ~10 ns/op
- ValidatePattern: ~1000 ns/op (depends on regex complexity)
- Validator.Validate (multiple rules): ~2000 ns/op

## Best Practices

1. **Reuse Validators**: Create validators once and reuse them for better performance
2. **Order Rules Appropriately**: Place required rules first to fail fast
3. **Use Type Checking**: Always check types in custom validators
4. **Provide Clear Messages**: Custom validators should return descriptive error messages
5. **Test Edge Cases**: Thoroughly test boundary conditions
6. **Consider Optional Fields**: Not all fields need required rules

## Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./pkg/validation -v

# Run specific test
go test ./pkg/validation -v -run TestValidateEmail

# Run with coverage
go test ./pkg/validation -cover

# Run benchmarks
go test -bench=. ./pkg/validation
```

Test coverage includes:
- Valid and invalid inputs for all validators
- Boundary conditions (min/max values)
- Type mismatches
- Edge cases (empty strings, nil values, etc.)
- Multiple validation rules
- Custom validation logic

## API Reference

### Functions

- `ValidateEmail(s string) error` - Validates email address (RFC 5322)
- `ValidateLength(s string, min, max int) error` - Validates string length
- `ValidateRange(n float64, min, max float64) error` - Validates numeric range
- `ValidatePattern(s string, pattern string) error` - Validates regex pattern
- `ValidateRequired(v interface{}) error` - Validates required field

### Validator Methods

- `NewValidator() *Validator` - Create new validator instance
- `AddEmailRule(field string)` - Add email validation rule
- `AddLengthRule(field string, min, max int)` - Add length validation rule
- `AddRangeRule(field string, min, max float64)` - Add range validation rule
- `AddPatternRule(field string, pattern string)` - Add pattern validation rule
- `AddRequiredRule(field string)` - Add required validation rule
- `AddCustomRule(field string, customFunc func(interface{}) error)` - Add custom validation rule
- `Validate(data map[string]interface{}) error` - Validate all fields
- `ValidateField(field string, value interface{}) error` - Validate single field
- `ClearRules()` - Clear all validation rules
- `GetRules(field string) []ValidationRule` - Get rules for a field
- `HasRules(field string) bool` - Check if field has rules

## License

Part of the Glyph language project.
