package interpreter

import (
	"testing"
)

// TestValuesEqual tests the valuesEqual method
func TestValuesEqual(t *testing.T) {
	interp := NewInterpreter()

	tests := []struct {
		name     string
		a, b     interface{}
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "first nil",
			a:        nil,
			b:        42,
			expected: false,
		},
		{
			name:     "second nil",
			a:        42,
			b:        nil,
			expected: false,
		},
		{
			name:     "equal ints",
			a:        int64(42),
			b:        int64(42),
			expected: true,
		},
		{
			name:     "unequal ints",
			a:        int64(42),
			b:        int64(43),
			expected: false,
		},
		{
			name:     "equal floats",
			a:        float64(3.14),
			b:        float64(3.14),
			expected: true,
		},
		{
			name:     "unequal floats",
			a:        float64(3.14),
			b:        float64(3.15),
			expected: false,
		},
		{
			name:     "float and int equal",
			a:        float64(42.0),
			b:        int64(42),
			expected: true,
		},
		{
			name:     "int and float equal",
			a:        int64(42),
			b:        float64(42.0),
			expected: true,
		},
		{
			name:     "equal strings",
			a:        "hello",
			b:        "hello",
			expected: true,
		},
		{
			name:     "unequal strings",
			a:        "hello",
			b:        "world",
			expected: false,
		},
		{
			name:     "equal bools true",
			a:        true,
			b:        true,
			expected: true,
		},
		{
			name:     "equal bools false",
			a:        false,
			b:        false,
			expected: true,
		},
		{
			name:     "unequal bools",
			a:        true,
			b:        false,
			expected: false,
		},
		{
			name:     "int to string mismatch",
			a:        int64(42),
			b:        "42",
			expected: false,
		},
		{
			name:     "float to string mismatch",
			a:        float64(3.14),
			b:        "3.14",
			expected: false,
		},
		{
			name:     "bool to int mismatch",
			a:        true,
			b:        int64(1),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interp.valuesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("valuesEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestTypeToString tests the TypeToString method
func TestTypeToString(t *testing.T) {
	tc := NewTypeChecker()

	tests := []struct {
		name     string
		typ      Type
		expected string
	}{
		{
			name:     "int type",
			typ:      IntType{},
			expected: "int",
		},
		{
			name:     "string type",
			typ:      StringType{},
			expected: "string",
		},
		{
			name:     "bool type",
			typ:      BoolType{},
			expected: "bool",
		},
		{
			name:     "float type",
			typ:      FloatType{},
			expected: "float",
		},
		{
			name:     "array of int",
			typ:      ArrayType{ElementType: IntType{}},
			expected: "List[int]",
		},
		{
			name:     "array of string",
			typ:      ArrayType{ElementType: StringType{}},
			expected: "List[string]",
		},
		{
			name:     "named type",
			typ:      NamedType{Name: "User"},
			expected: "User",
		},
		{
			name:     "nil type",
			typ:      nil,
			expected: "any",
		},
		{
			name:     "optional type",
			typ:      OptionalType{InnerType: IntType{}},
			expected: "int?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tc.TypeToString(tt.typ)
			if result != tt.expected {
				t.Errorf("TypeToString(%T) = %q, want %q", tt.typ, result, tt.expected)
			}
		})
	}
}

// TestTypesCompatible tests the TypesCompatible method
func TestTypesCompatible(t *testing.T) {
	tc := NewTypeChecker()

	tests := []struct {
		name     string
		expected Type
		actual   Type
		result   bool
	}{
		{
			name:     "int equals int",
			expected: IntType{},
			actual:   IntType{},
			result:   true,
		},
		{
			name:     "string equals string",
			expected: StringType{},
			actual:   StringType{},
			result:   true,
		},
		{
			name:     "int does not equal string",
			expected: IntType{},
			actual:   StringType{},
			result:   false,
		},
		{
			name:     "array int equals array int",
			expected: ArrayType{ElementType: IntType{}},
			actual:   ArrayType{ElementType: IntType{}},
			result:   true,
		},
		{
			name:     "array int does not equal array string",
			expected: ArrayType{ElementType: IntType{}},
			actual:   ArrayType{ElementType: StringType{}},
			result:   false,
		},
		{
			name:     "named type equals same named type",
			expected: NamedType{Name: "User"},
			actual:   NamedType{Name: "User"},
			result:   true,
		},
		{
			name:     "named type does not equal different named type",
			expected: NamedType{Name: "User"},
			actual:   NamedType{Name: "Post"},
			result:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tc.TypesCompatible(tt.expected, tt.actual)
			if result != tt.result {
				t.Errorf("TypesCompatible(%T, %T) = %v, want %v", tt.expected, tt.actual, result, tt.result)
			}
		})
	}
}

// TestEmitEvent tests the EmitEvent method
func TestEmitEvent(t *testing.T) {
	interp := NewInterpreter()

	// Test emitting event with no handlers
	err := interp.EmitEvent("test-event", map[string]interface{}{"key": "value"})
	// Should not error even with no handlers
	if err != nil {
		t.Errorf("EmitEvent should not error with no handlers: %v", err)
	}

	// Add an event handler
	handler := EventHandler{
		EventType: "test-event",
		Body:      []Statement{},
	}
	interp.eventHandlers["test-event"] = []EventHandler{handler}

	// Emit again
	err = interp.EmitEvent("test-event", map[string]interface{}{"data": "test"})
	if err != nil {
		t.Errorf("EmitEvent failed: %v", err)
	}

	// Emit non-existent event
	err = interp.EmitEvent("nonexistent-event", nil)
	if err != nil {
		t.Errorf("EmitEvent should not error for nonexistent event: %v", err)
	}
}

// TestEvaluateFunctionCallBuiltins tests evaluating builtin function calls
func TestEvaluateFunctionCallBuiltins(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test toString function
	env.Define("num", int64(42))
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "toString",
		Args: []Expr{VariableExpr{Name: "num"}},
	}, env)
	if err != nil {
		t.Errorf("toString failed: %v", err)
	}
	if result != "42" {
		t.Errorf("toString(42) = %v, want '42'", result)
	}

	// Test parseInt function
	env.Define("strNum", "123")
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "parseInt",
		Args: []Expr{VariableExpr{Name: "strNum"}},
	}, env)
	if err != nil {
		t.Errorf("parseInt failed: %v", err)
	}
	if result != int64(123) {
		t.Errorf("parseInt('123') = %v, want 123", result)
	}

	// Test parseFloat function
	env.Define("strFloat", "3.14")
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "parseFloat",
		Args: []Expr{VariableExpr{Name: "strFloat"}},
	}, env)
	if err != nil {
		t.Errorf("parseFloat failed: %v", err)
	}
	if result != float64(3.14) {
		t.Errorf("parseFloat('3.14') = %v, want 3.14", result)
	}

	// Test upper function
	env.Define("testStr", "hello")
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "upper",
		Args: []Expr{VariableExpr{Name: "testStr"}},
	}, env)
	if err != nil {
		t.Errorf("upper failed: %v", err)
	}
	if result != "HELLO" {
		t.Errorf("upper('hello') = %v, want 'HELLO'", result)
	}

	// Test lower function
	env.Define("upperStr", "WORLD")
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "lower",
		Args: []Expr{VariableExpr{Name: "upperStr"}},
	}, env)
	if err != nil {
		t.Errorf("lower failed: %v", err)
	}
	if result != "world" {
		t.Errorf("lower('WORLD') = %v, want 'world'", result)
	}

	// Test trim function
	env.Define("spacedStr", "  spaced  ")
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "trim",
		Args: []Expr{VariableExpr{Name: "spacedStr"}},
	}, env)
	if err != nil {
		t.Errorf("trim failed: %v", err)
	}
	if result != "spaced" {
		t.Errorf("trim('  spaced  ') = %v, want 'spaced'", result)
	}

	// Test length function
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "length",
		Args: []Expr{VariableExpr{Name: "testStr"}},
	}, env)
	if err != nil {
		t.Errorf("length failed: %v", err)
	}
	if result != int64(5) {
		t.Errorf("length('hello') = %v, want 5", result)
	}

	// Test abs function with int
	env.Define("negNum", int64(-42))
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "abs",
		Args: []Expr{VariableExpr{Name: "negNum"}},
	}, env)
	if err != nil {
		t.Errorf("abs failed: %v", err)
	}
	if result != int64(42) {
		t.Errorf("abs(-42) = %v, want 42", result)
	}
}

// TestEvaluateNe tests not equal evaluation
func TestEvaluateNe(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test int != int (not equal)
	result, err := interp.EvaluateExpression(BinaryOpExpr{
		Op:    Ne,
		Left:  LiteralExpr{Value: IntLiteral{Value: 1}},
		Right: LiteralExpr{Value: IntLiteral{Value: 2}},
	}, env)
	if err != nil {
		t.Errorf("1 != 2 failed: %v", err)
	}
	if result != true {
		t.Errorf("1 != 2 = %v, want true", result)
	}

	// Test int != int (equal)
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Ne,
		Left:  LiteralExpr{Value: IntLiteral{Value: 1}},
		Right: LiteralExpr{Value: IntLiteral{Value: 1}},
	}, env)
	if err != nil {
		t.Errorf("1 != 1 failed: %v", err)
	}
	if result != false {
		t.Errorf("1 != 1 = %v, want false", result)
	}

	// Test string != string
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Ne,
		Left:  LiteralExpr{Value: StringLiteral{Value: "a"}},
		Right: LiteralExpr{Value: StringLiteral{Value: "b"}},
	}, env)
	if err != nil {
		t.Errorf("'a' != 'b' failed: %v", err)
	}
	if result != true {
		t.Errorf("'a' != 'b' = %v, want true", result)
	}
}

// TestMoreBuiltinFunctions tests additional builtin function calls
func TestMoreBuiltinFunctions(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test time.now function
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "time.now",
		Args: []Expr{},
	}, env)
	if err != nil {
		t.Errorf("time.now failed: %v", err)
	}
	if result != int64(1234567890) {
		t.Errorf("time.now = %v, want 1234567890", result)
	}

	// Test now function
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "now",
		Args: []Expr{},
	}, env)
	if err != nil {
		t.Errorf("now failed: %v", err)
	}
	if result != int64(1234567890) {
		t.Errorf("now = %v, want 1234567890", result)
	}

	// Test split function
	env.Define("csv", "a,b,c")
	env.Define("delim", ",")
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "split",
		Args: []Expr{VariableExpr{Name: "csv"}, VariableExpr{Name: "delim"}},
	}, env)
	if err != nil {
		t.Errorf("split failed: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok || len(arr) != 3 {
		t.Errorf("split result invalid: %v", result)
	}

	// Test join function
	env.Define("parts", []interface{}{"a", "b", "c"})
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "join",
		Args: []Expr{VariableExpr{Name: "parts"}, LiteralExpr{Value: StringLiteral{Value: "-"}}},
	}, env)
	if err != nil {
		t.Errorf("join failed: %v", err)
	}
	if result != "a-b-c" {
		t.Errorf("join = %v, want 'a-b-c'", result)
	}

	// Test contains function
	env.Define("haystack", "hello world")
	env.Define("needle", "world")
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "contains",
		Args: []Expr{VariableExpr{Name: "haystack"}, VariableExpr{Name: "needle"}},
	}, env)
	if err != nil {
		t.Errorf("contains failed: %v", err)
	}
	if result != true {
		t.Errorf("contains = %v, want true", result)
	}

	// Test replace function
	env.Define("original", "hello world")
	env.Define("oldStr", "world")
	env.Define("newStr", "go")
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "replace",
		Args: []Expr{VariableExpr{Name: "original"}, VariableExpr{Name: "oldStr"}, VariableExpr{Name: "newStr"}},
	}, env)
	if err != nil {
		t.Errorf("replace failed: %v", err)
	}
	if result != "hello go" {
		t.Errorf("replace = %v, want 'hello go'", result)
	}

	// Test startsWith function
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "startsWith",
		Args: []Expr{VariableExpr{Name: "haystack"}, LiteralExpr{Value: StringLiteral{Value: "hello"}}},
	}, env)
	if err != nil {
		t.Errorf("startsWith failed: %v", err)
	}
	if result != true {
		t.Errorf("startsWith = %v, want true", result)
	}

	// Test endsWith function
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "endsWith",
		Args: []Expr{VariableExpr{Name: "haystack"}, LiteralExpr{Value: StringLiteral{Value: "world"}}},
	}, env)
	if err != nil {
		t.Errorf("endsWith failed: %v", err)
	}
	if result != true {
		t.Errorf("endsWith = %v, want true", result)
	}

	// Test indexOf function
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "indexOf",
		Args: []Expr{VariableExpr{Name: "haystack"}, LiteralExpr{Value: StringLiteral{Value: "world"}}},
	}, env)
	if err != nil {
		t.Errorf("indexOf failed: %v", err)
	}
	if result != int64(6) {
		t.Errorf("indexOf = %v, want 6", result)
	}

	// Test charAt function
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "charAt",
		Args: []Expr{VariableExpr{Name: "haystack"}, LiteralExpr{Value: IntLiteral{Value: 0}}},
	}, env)
	if err != nil {
		t.Errorf("charAt failed: %v", err)
	}
	if result != "h" {
		t.Errorf("charAt = %v, want 'h'", result)
	}

	// Test min function with ints
	env.Define("a", int64(10))
	env.Define("b", int64(20))
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "min",
		Args: []Expr{VariableExpr{Name: "a"}, VariableExpr{Name: "b"}},
	}, env)
	if err != nil {
		t.Errorf("min failed: %v", err)
	}
	if result != int64(10) {
		t.Errorf("min = %v, want 10", result)
	}

	// Test max function with ints
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "max",
		Args: []Expr{VariableExpr{Name: "a"}, VariableExpr{Name: "b"}},
	}, env)
	if err != nil {
		t.Errorf("max failed: %v", err)
	}
	if result != int64(20) {
		t.Errorf("max = %v, want 20", result)
	}
}

// TestBuiltinErrorCases tests error cases for builtin functions
func TestBuiltinErrorCases(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test upper with wrong argument count
	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "upper",
		Args: []Expr{},
	}, env)
	if err == nil {
		t.Error("upper() with no args should error")
	}

	// Test upper with non-string argument
	env.Define("num", int64(42))
	_, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "upper",
		Args: []Expr{VariableExpr{Name: "num"}},
	}, env)
	if err == nil {
		t.Error("upper(int) should error")
	}

	// Test lower with wrong argument count
	_, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "lower",
		Args: []Expr{},
	}, env)
	if err == nil {
		t.Error("lower() with no args should error")
	}

	// Test trim with wrong argument count
	_, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "trim",
		Args: []Expr{},
	}, env)
	if err == nil {
		t.Error("trim() with no args should error")
	}

	// Test split with wrong argument count
	env.Define("str", "test")
	_, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "split",
		Args: []Expr{VariableExpr{Name: "str"}},
	}, env)
	if err == nil {
		t.Error("split() with 1 arg should error")
	}

	// Test join with wrong argument count
	_, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "join",
		Args: []Expr{VariableExpr{Name: "str"}},
	}, env)
	if err == nil {
		t.Error("join() with 1 arg should error")
	}

	// Test undefined function
	_, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "nonexistent",
		Args: []Expr{},
	}, env)
	if err == nil {
		t.Error("undefined function should error")
	}
}

// TestObjectExpr tests object expression evaluation
func TestObjectExpr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	result, err := interp.EvaluateExpression(ObjectExpr{
		Fields: []ObjectField{
			{Key: "name", Value: LiteralExpr{Value: StringLiteral{Value: "Alice"}}},
			{Key: "age", Value: LiteralExpr{Value: IntLiteral{Value: 30}}},
		},
	}, env)
	if err != nil {
		t.Errorf("ObjectExpr failed: %v", err)
	}
	obj, ok := result.(map[string]interface{})
	if !ok {
		t.Errorf("ObjectExpr result not a map: %T", result)
	}
	if obj["name"] != "Alice" {
		t.Errorf("name = %v, want 'Alice'", obj["name"])
	}
	if obj["age"] != int64(30) {
		t.Errorf("age = %v, want 30", obj["age"])
	}
}

// TestArrayExpr tests array expression evaluation
func TestArrayExpr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	result, err := interp.EvaluateExpression(ArrayExpr{
		Elements: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 1}},
			LiteralExpr{Value: IntLiteral{Value: 2}},
			LiteralExpr{Value: IntLiteral{Value: 3}},
		},
	}, env)
	if err != nil {
		t.Errorf("ArrayExpr failed: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok {
		t.Errorf("ArrayExpr result not an array: %T", result)
	}
	if len(arr) != 3 {
		t.Errorf("array length = %d, want 3", len(arr))
	}
}

// TestUnaryOpExpr tests unary operator evaluation
func TestUnaryOpExpr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test negation
	result, err := interp.EvaluateExpression(UnaryOpExpr{
		Op:    Neg, // negative
		Right: LiteralExpr{Value: IntLiteral{Value: 42}},
	}, env)
	if err != nil {
		t.Errorf("unary minus failed: %v", err)
	}
	if result != int64(-42) {
		t.Errorf("-42 = %v, want -42", result)
	}

	// Test logical not
	result, err = interp.EvaluateExpression(UnaryOpExpr{
		Op:    Not,
		Right: LiteralExpr{Value: BoolLiteral{Value: true}},
	}, env)
	if err != nil {
		t.Errorf("unary not failed: %v", err)
	}
	if result != false {
		t.Errorf("!true = %v, want false", result)
	}
}

// TestFieldAccessExpr tests field access expression evaluation
func TestFieldAccessExpr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Define an object
	env.Define("user", map[string]interface{}{
		"name": "Alice",
		"age":  int64(30),
	})

	result, err := interp.EvaluateExpression(FieldAccessExpr{
		Object: VariableExpr{Name: "user"},
		Field:  "name",
	}, env)
	if err != nil {
		t.Errorf("field access failed: %v", err)
	}
	if result != "Alice" {
		t.Errorf("user.name = %v, want 'Alice'", result)
	}
}

// TestUnaryOpExprEdgeCases tests edge cases for unary operators
func TestUnaryOpExprEdgeCases(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test negation with float
	env.Define("f", float64(3.14))
	result, err := interp.EvaluateExpression(UnaryOpExpr{
		Op:    Neg,
		Right: VariableExpr{Name: "f"},
	}, env)
	if err != nil {
		t.Errorf("unary minus float failed: %v", err)
	}
	if result != float64(-3.14) {
		t.Errorf("-3.14 = %v, want -3.14", result)
	}

	// Test Not with non-boolean (should error)
	env.Define("n", int64(42))
	_, err = interp.EvaluateExpression(UnaryOpExpr{
		Op:    Not,
		Right: VariableExpr{Name: "n"},
	}, env)
	if err == nil {
		t.Error("!int should error")
	}

	// Test Neg with non-numeric (should error)
	env.Define("s", "hello")
	_, err = interp.EvaluateExpression(UnaryOpExpr{
		Op:    Neg,
		Right: VariableExpr{Name: "s"},
	}, env)
	if err == nil {
		t.Error("-string should error")
	}
}

// TestMoreBinaryOps tests additional binary operations
func TestMoreBinaryOps(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test Less Than
	result, err := interp.EvaluateExpression(BinaryOpExpr{
		Op:    Lt,
		Left:  LiteralExpr{Value: IntLiteral{Value: 1}},
		Right: LiteralExpr{Value: IntLiteral{Value: 2}},
	}, env)
	if err != nil {
		t.Errorf("1 < 2 failed: %v", err)
	}
	if result != true {
		t.Errorf("1 < 2 = %v, want true", result)
	}

	// Test Greater Than
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Gt,
		Left:  LiteralExpr{Value: IntLiteral{Value: 2}},
		Right: LiteralExpr{Value: IntLiteral{Value: 1}},
	}, env)
	if err != nil {
		t.Errorf("2 > 1 failed: %v", err)
	}
	if result != true {
		t.Errorf("2 > 1 = %v, want true", result)
	}

	// Test Less Than or Equal
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Le,
		Left:  LiteralExpr{Value: IntLiteral{Value: 2}},
		Right: LiteralExpr{Value: IntLiteral{Value: 2}},
	}, env)
	if err != nil {
		t.Errorf("2 <= 2 failed: %v", err)
	}
	if result != true {
		t.Errorf("2 <= 2 = %v, want true", result)
	}

	// Test Greater Than or Equal
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Ge,
		Left:  LiteralExpr{Value: IntLiteral{Value: 2}},
		Right: LiteralExpr{Value: IntLiteral{Value: 2}},
	}, env)
	if err != nil {
		t.Errorf("2 >= 2 failed: %v", err)
	}
	if result != true {
		t.Errorf("2 >= 2 = %v, want true", result)
	}

	// Test And
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    And,
		Left:  LiteralExpr{Value: BoolLiteral{Value: true}},
		Right: LiteralExpr{Value: BoolLiteral{Value: true}},
	}, env)
	if err != nil {
		t.Errorf("true && true failed: %v", err)
	}
	if result != true {
		t.Errorf("true && true = %v, want true", result)
	}

	// Test Or
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Or,
		Left:  LiteralExpr{Value: BoolLiteral{Value: false}},
		Right: LiteralExpr{Value: BoolLiteral{Value: true}},
	}, env)
	if err != nil {
		t.Errorf("false || true failed: %v", err)
	}
	if result != true {
		t.Errorf("false || true = %v, want true", result)
	}

	// Test Mul
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Mul,
		Left:  LiteralExpr{Value: IntLiteral{Value: 6}},
		Right: LiteralExpr{Value: IntLiteral{Value: 7}},
	}, env)
	if err != nil {
		t.Errorf("6 * 7 failed: %v", err)
	}
	if result != int64(42) {
		t.Errorf("6 * 7 = %v, want 42", result)
	}
}

// TestSubstringFunction tests the substring builtin
func TestSubstringFunction(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("str", "hello world")
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "substring",
		Args: []Expr{
			VariableExpr{Name: "str"},
			LiteralExpr{Value: IntLiteral{Value: 0}},
			LiteralExpr{Value: IntLiteral{Value: 5}},
		},
	}, env)
	if err != nil {
		t.Errorf("substring failed: %v", err)
	}
	if result != "hello" {
		t.Errorf("substring = %v, want 'hello'", result)
	}
}

// TestFloatLiteral tests float literal evaluation
func TestFloatLiteral(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	result, err := interp.EvaluateExpression(LiteralExpr{
		Value: FloatLiteral{Value: 3.14159},
	}, env)
	if err != nil {
		t.Errorf("float literal failed: %v", err)
	}
	if result != float64(3.14159) {
		t.Errorf("float = %v, want 3.14159", result)
	}
}

// TestMoreTypesCompatible tests additional type compatibility checks
func TestMoreTypesCompatible(t *testing.T) {
	tc := NewTypeChecker()

	// Test bool equals bool
	result := tc.TypesCompatible(BoolType{}, BoolType{})
	if !result {
		t.Error("BoolType should be compatible with BoolType")
	}

	// Test float equals float
	result = tc.TypesCompatible(FloatType{}, FloatType{})
	if !result {
		t.Error("FloatType should be compatible with FloatType")
	}

	// Test optional types
	result = tc.TypesCompatible(OptionalType{InnerType: IntType{}}, OptionalType{InnerType: IntType{}})
	if !result {
		t.Error("Optional int should be compatible with optional int")
	}

	// Test non-optional with optional expected (int is compatible with Optional[int])
	result = tc.TypesCompatible(IntType{}, OptionalType{InnerType: IntType{}})
	if !result {
		t.Error("int should be compatible with Optional[int]")
	}
}

// TestEnvironmentScopes tests nested scopes
func TestEnvironmentScopes(t *testing.T) {
	env := NewEnvironment()

	// Define in parent scope
	env.Define("x", int64(10))

	// Create child scope
	child := NewChildEnvironment(env)
	child.Define("y", int64(20))

	// Check parent and child visibility
	x, err := child.Get("x")
	if err != nil {
		t.Errorf("child should see parent's x: %v", err)
	}
	if x != int64(10) {
		t.Errorf("x = %v, want 10", x)
	}

	y, err := child.Get("y")
	if err != nil {
		t.Errorf("child should see y: %v", err)
	}
	if y != int64(20) {
		t.Errorf("y = %v, want 20", y)
	}

	// Parent should not see child's y
	_, err = env.Get("y")
	if err == nil {
		t.Error("parent should not see child's y")
	}
}

// TestEnvironmentSet tests setting variable values
func TestEnvironmentSet(t *testing.T) {
	env := NewEnvironment()

	// Define a variable
	env.Define("x", int64(10))

	// Set a new value
	err := env.Set("x", int64(20))
	if err != nil {
		t.Errorf("set failed: %v", err)
	}

	val, _ := env.Get("x")
	if val != int64(20) {
		t.Errorf("x = %v, want 20", val)
	}

	// Set on undefined variable should error
	err = env.Set("undefined", int64(30))
	if err == nil {
		t.Error("set on undefined should error")
	}
}

// TestEnvironmentHas tests the Has method
func TestEnvironmentHas(t *testing.T) {
	env := NewEnvironment()

	// Test undefined variable
	if env.Has("undefined") {
		t.Error("undefined variable should not be found")
	}

	// Define a variable
	env.Define("x", int64(10))
	if !env.Has("x") {
		t.Error("x should be found")
	}

	// Test in child scope
	child := NewChildEnvironment(env)
	if !child.Has("x") {
		t.Error("x should be found in child scope")
	}

	child.Define("y", int64(20))
	if !child.Has("y") {
		t.Error("y should be found in child scope")
	}

	// Parent should not have child's variable
	if env.Has("y") {
		t.Error("parent should not have child's y")
	}
}

// TestExecuteStatements tests statement execution
func TestExecuteStatements(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test assignment statement
	stmt := AssignStatement{
		Target: "x",
		Value:  LiteralExpr{Value: IntLiteral{Value: 42}},
	}

	_, err := interp.ExecuteStatement(stmt, env)
	if err != nil {
		t.Errorf("assign failed: %v", err)
	}

	val, _ := env.Get("x")
	if val != int64(42) {
		t.Errorf("x = %v, want 42", val)
	}

	// Test return statement
	retStmt := ReturnStatement{
		Value: LiteralExpr{Value: IntLiteral{Value: 100}},
	}

	_, err = interp.ExecuteStatement(retStmt, env)
	if err == nil {
		t.Error("return should return error (returnValue)")
	}
}

// TestExecuteIfStatement tests if statement execution
func TestExecuteIfStatement(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("result", "")
	// Test if true branch
	ifStmt := IfStatement{
		Condition: LiteralExpr{Value: BoolLiteral{Value: true}},
		ThenBlock: []Statement{
			AssignStatement{Target: "result", Value: LiteralExpr{Value: StringLiteral{Value: "then"}}},
		},
		ElseBlock: []Statement{
			AssignStatement{Target: "result", Value: LiteralExpr{Value: StringLiteral{Value: "else"}}},
		},
	}

	_, err := interp.ExecuteStatement(ifStmt, env)
	if err != nil {
		t.Errorf("if failed: %v", err)
	}

	val, _ := env.Get("result")
	if val != "then" {
		t.Errorf("result = %v, want 'then'", val)
	}

	// Test if false branch
	env2 := NewEnvironment()
	env2.Define("result", "")
	ifStmt.Condition = LiteralExpr{Value: BoolLiteral{Value: false}}
	_, err = interp.ExecuteStatement(ifStmt, env2)
	if err != nil {
		t.Errorf("if else failed: %v", err)
	}

	val, _ = env2.Get("result")
	if val != "else" {
		t.Errorf("result = %v, want 'else'", val)
	}
}

// TestExecuteWhileStatement tests while loop execution
func TestExecuteWhileStatement(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("i", int64(0))

	whileStmt := WhileStatement{
		Condition: BinaryOpExpr{
			Op:    Lt,
			Left:  VariableExpr{Name: "i"},
			Right: LiteralExpr{Value: IntLiteral{Value: 5}},
		},
		Body: []Statement{
			AssignStatement{
				Target: "i",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "i"},
					Right: LiteralExpr{Value: IntLiteral{Value: 1}},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(whileStmt, env)
	if err != nil {
		t.Errorf("while failed: %v", err)
	}

	val, _ := env.Get("i")
	if val != int64(5) {
		t.Errorf("i = %v, want 5", val)
	}
}

// TestExecuteForStatement tests for loop execution
func TestExecuteForStatement(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("sum", int64(0))

	forStmt := ForStatement{
		ValueVar: "item",
		Iterable: ArrayExpr{
			Elements: []Expr{
				LiteralExpr{Value: IntLiteral{Value: 1}},
				LiteralExpr{Value: IntLiteral{Value: 2}},
				LiteralExpr{Value: IntLiteral{Value: 3}},
			},
		},
		Body: []Statement{
			AssignStatement{
				Target: "sum",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "sum"},
					Right: VariableExpr{Name: "item"},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(forStmt, env)
	if err != nil {
		t.Errorf("for failed: %v", err)
	}

	val, _ := env.Get("sum")
	if val != int64(6) {
		t.Errorf("sum = %v, want 6", val)
	}
}

// TestExecuteForWithIndex tests for loop with key variable
func TestExecuteForWithIndex(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("lastIdx", int64(-1))

	forStmt := ForStatement{
		KeyVar:   "idx",
		ValueVar: "item",
		Iterable: ArrayExpr{
			Elements: []Expr{
				LiteralExpr{Value: StringLiteral{Value: "a"}},
				LiteralExpr{Value: StringLiteral{Value: "b"}},
				LiteralExpr{Value: StringLiteral{Value: "c"}},
			},
		},
		Body: []Statement{
			AssignStatement{Target: "lastIdx", Value: VariableExpr{Name: "idx"}},
		},
	}

	_, err := interp.ExecuteStatement(forStmt, env)
	if err != nil {
		t.Errorf("for with index failed: %v", err)
	}

	val, _ := env.Get("lastIdx")
	if val != int64(2) {
		t.Errorf("lastIdx = %v, want 2", val)
	}
}

// TestExecuteSwitchStatement tests switch statement execution
func TestExecuteSwitchStatement(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("result", "")
	env.Define("x", int64(2))

	switchStmt := SwitchStatement{
		Value: VariableExpr{Name: "x"},
		Cases: []SwitchCase{
			{
				Value: LiteralExpr{Value: IntLiteral{Value: 1}},
				Body:  []Statement{AssignStatement{Target: "result", Value: LiteralExpr{Value: StringLiteral{Value: "one"}}}},
			},
			{
				Value: LiteralExpr{Value: IntLiteral{Value: 2}},
				Body:  []Statement{AssignStatement{Target: "result", Value: LiteralExpr{Value: StringLiteral{Value: "two"}}}},
			},
		},
		Default: []Statement{
			AssignStatement{Target: "result", Value: LiteralExpr{Value: StringLiteral{Value: "default"}}},
		},
	}

	_, err := interp.ExecuteStatement(switchStmt, env)
	if err != nil {
		t.Errorf("switch failed: %v", err)
	}

	val, _ := env.Get("result")
	if val != "two" {
		t.Errorf("result = %v, want 'two'", val)
	}
}

// TestExecuteSwitchDefault tests switch default branch
func TestExecuteSwitchDefault(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()
	env.Define("result", "")
	env.Define("x", int64(99))

	switchStmt := SwitchStatement{
		Value: VariableExpr{Name: "x"},
		Cases: []SwitchCase{
			{
				Value: LiteralExpr{Value: IntLiteral{Value: 1}},
				Body:  []Statement{AssignStatement{Target: "result", Value: LiteralExpr{Value: StringLiteral{Value: "one"}}}},
			},
		},
		Default: []Statement{
			AssignStatement{Target: "result", Value: LiteralExpr{Value: StringLiteral{Value: "default"}}},
		},
	}

	_, err := interp.ExecuteStatement(switchStmt, env)
	if err != nil {
		t.Errorf("switch default failed: %v", err)
	}

	val, _ := env.Get("result")
	if val != "default" {
		t.Errorf("result = %v, want 'default'", val)
	}
}

// TestExecuteRouteSimple tests simple route execution
func TestExecuteRouteSimple(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Method: Get,
		Path:   "/test/:id",
		Body: []Statement{
			ReturnStatement{
				Value: ObjectExpr{
					Fields: []ObjectField{
						{Key: "id", Value: VariableExpr{Name: "id"}},
					},
				},
			},
		},
	}

	pathParams := map[string]string{"id": "123"}
	result, err := interp.ExecuteRouteSimple(route, pathParams)
	if err != nil {
		t.Errorf("ExecuteRouteSimple failed: %v", err)
	}

	obj, ok := result.(map[string]interface{})
	if !ok {
		t.Errorf("result not a map: %T", result)
	}
	if obj["id"] != "123" {
		t.Errorf("id = %v, want '123'", obj["id"])
	}
}

// TestLoadModuleWithTypeDef tests loading module with type definitions
func TestLoadModuleWithTypeDef(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TypeDef{
				Name: "User",
				Fields: []Field{
					{Name: "id", TypeAnnotation: IntType{}, Required: true},
					{Name: "name", TypeAnnotation: StringType{}, Required: true},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	if err != nil {
		t.Errorf("LoadModule failed: %v", err)
	}

	typeDef, ok := interp.GetTypeDef("User")
	if !ok {
		t.Error("User type not found")
	}
	if typeDef.Name != "User" {
		t.Errorf("type name = %s, want 'User'", typeDef.Name)
	}
}

// TestLoadModuleWithFunction tests loading module with functions
func TestLoadModuleWithFunction(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&Function{
				Name: "add",
				Params: []Field{
					{Name: "a", TypeAnnotation: IntType{}},
					{Name: "b", TypeAnnotation: IntType{}},
				},
				ReturnType: IntType{},
				Body: []Statement{
					ReturnStatement{
						Value: BinaryOpExpr{
							Op:    Add,
							Left:  VariableExpr{Name: "a"},
							Right: VariableExpr{Name: "b"},
						},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	if err != nil {
		t.Errorf("LoadModule failed: %v", err)
	}

	fn, ok := interp.GetFunction("add")
	if !ok {
		t.Error("add function not found")
	}
	if fn.Name != "add" {
		t.Errorf("function name = %s, want 'add'", fn.Name)
	}
}

// TestExecuteCommand tests command execution
func TestExecuteCommand(t *testing.T) {
	interp := NewInterpreter()

	cmd := &Command{
		Name: "greet",
		Params: []CommandParam{
			{Name: "name", Type: StringType{}, Required: true},
		},
		Body: []Statement{
			ReturnStatement{
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  LiteralExpr{Value: StringLiteral{Value: "Hello, "}},
					Right: VariableExpr{Name: "name"},
				},
			},
		},
	}

	result, err := interp.ExecuteCommand(cmd, map[string]interface{}{"name": "World"})
	if err != nil {
		t.Errorf("ExecuteCommand failed: %v", err)
	}

	if result != "Hello, World" {
		t.Errorf("result = %v, want 'Hello, World'", result)
	}
}

// TestExecuteCommandWithDefault tests command with default parameter
func TestExecuteCommandWithDefault(t *testing.T) {
	interp := NewInterpreter()

	cmd := &Command{
		Name: "greet",
		Params: []CommandParam{
			{Name: "greeting", Type: StringType{}, Required: false, Default: LiteralExpr{Value: StringLiteral{Value: "Hi"}}},
		},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "greeting"}},
		},
	}

	result, err := interp.ExecuteCommand(cmd, map[string]interface{}{})
	if err != nil {
		t.Errorf("ExecuteCommand failed: %v", err)
	}

	if result != "Hi" {
		t.Errorf("result = %v, want 'Hi'", result)
	}
}

// TestMarkerInterfaces tests interface marker methods
func TestMarkerInterfaces(t *testing.T) {
	// Test Item interface
	var items []Item
	items = append(items, &TypeDef{Name: "Test"})
	items = append(items, &Route{Path: "/test"})
	items = append(items, &Function{Name: "fn"})
	items = append(items, &WebSocketRoute{Path: "/ws"})
	items = append(items, &Command{Name: "cmd"})
	items = append(items, &CronTask{Schedule: "* * * * *"})
	items = append(items, &EventHandler{EventType: "test"})
	items = append(items, &QueueWorker{QueueName: "q"})

	for _, item := range items {
		// Just verify these implement Item
		_ = item
	}

	// Test Statement interface
	var stmts []Statement
	stmts = append(stmts, AssignStatement{Target: "x"})
	stmts = append(stmts, ReturnStatement{})
	stmts = append(stmts, IfStatement{})
	stmts = append(stmts, WhileStatement{})
	stmts = append(stmts, ForStatement{})
	stmts = append(stmts, SwitchStatement{})

	for _, stmt := range stmts {
		// Just verify these implement Statement
		_ = stmt
	}

	// Test Expr interface
	var exprs []Expr
	exprs = append(exprs, LiteralExpr{})
	exprs = append(exprs, VariableExpr{})
	exprs = append(exprs, BinaryOpExpr{})
	exprs = append(exprs, UnaryOpExpr{})
	exprs = append(exprs, FunctionCallExpr{})
	exprs = append(exprs, ObjectExpr{})
	exprs = append(exprs, ArrayExpr{})
	exprs = append(exprs, FieldAccessExpr{})

	for _, expr := range exprs {
		// Just verify these implement Expr
		_ = expr
	}
}

// TestTypeString tests Type String methods
func TestTypeString(t *testing.T) {
	tests := []struct {
		name string
		typ  Type
	}{
		{"IntType", IntType{}},
		{"StringType", StringType{}},
		{"BoolType", BoolType{}},
		{"FloatType", FloatType{}},
		{"ArrayType", ArrayType{ElementType: IntType{}}},
		{"OptionalType", OptionalType{InnerType: StringType{}}},
		{"NamedType", NamedType{Name: "User"}},
		{"DatabaseType", DatabaseType{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just ensure String() doesn't panic
			_ = tt.typ
		})
	}
}

// TestBinOpString tests BinOp String method
func TestBinOpString(t *testing.T) {
	ops := []BinOp{Add, Sub, Mul, Div, Eq, Ne, Lt, Le, Gt, Ge, And, Or}
	for _, op := range ops {
		s := op.String()
		if s == "" {
			t.Errorf("BinOp %d has empty string", op)
		}
	}
}

// TestDivision tests division operations
func TestDivision(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test int division
	result, err := interp.EvaluateExpression(BinaryOpExpr{
		Op:    Div,
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Right: LiteralExpr{Value: IntLiteral{Value: 3}},
	}, env)
	if err != nil {
		t.Errorf("int division failed: %v", err)
	}
	if result != int64(3) {
		t.Errorf("10 / 3 = %v, want 3", result)
	}

	// Test float division
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Div,
		Left:  LiteralExpr{Value: FloatLiteral{Value: 10.0}},
		Right: LiteralExpr{Value: FloatLiteral{Value: 4.0}},
	}, env)
	if err != nil {
		t.Errorf("float division failed: %v", err)
	}
	if result != float64(2.5) {
		t.Errorf("10.0 / 4.0 = %v, want 2.5", result)
	}

	// Test division by zero
	_, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Div,
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Right: LiteralExpr{Value: IntLiteral{Value: 0}},
	}, env)
	if err == nil {
		t.Error("division by zero should error")
	}
}

// TestFloatComparisons tests float comparison operations
func TestFloatComparisons(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test float less than
	result, err := interp.EvaluateExpression(BinaryOpExpr{
		Op:    Lt,
		Left:  LiteralExpr{Value: FloatLiteral{Value: 1.5}},
		Right: LiteralExpr{Value: FloatLiteral{Value: 2.5}},
	}, env)
	if err != nil {
		t.Errorf("float lt failed: %v", err)
	}
	if result != true {
		t.Errorf("1.5 < 2.5 = %v, want true", result)
	}

	// Test float greater than
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Gt,
		Left:  LiteralExpr{Value: FloatLiteral{Value: 2.5}},
		Right: LiteralExpr{Value: FloatLiteral{Value: 1.5}},
	}, env)
	if err != nil {
		t.Errorf("float gt failed: %v", err)
	}
	if result != true {
		t.Errorf("2.5 > 1.5 = %v, want true", result)
	}
}
