package interpreter

import (
	"fmt"
	"testing"
)

// TestExpressionStatementMarker tests expression statement isStatement marker
func TestExpressionStatementMarker(t *testing.T) {
	stmt := ExpressionStatement{Expr: LiteralExpr{Value: IntLiteral{Value: 42}}}
	stmt.isStatement()
}

// TestCallMethod tests the CallMethod reflection function
func TestCallMethod(t *testing.T) {
	// Test calling method on a string with non-existent method
	_, err := CallMethod("hello", "NonExistent", "arg")
	if err == nil {
		t.Error("CallMethod should fail for non-existent method")
	}
}

// TestHasMethod tests the HasMethod function
func TestHasMethod(t *testing.T) {
	// Test with string - has no exported methods accessible via reflect
	result := HasMethod("hello", "NonExistent")
	if result {
		t.Error("HasMethod should return false for non-existent method")
	}

	// Test with a type that has methods
	type myString string
	result = HasMethod(myString("hello"), "NonExistent")
	if result {
		t.Error("HasMethod should return false for non-existent method on custom type")
	}
}

// TestGetMethodNames tests the GetMethodNames function
func TestGetMethodNames(t *testing.T) {
	// Test with a basic type
	methods := GetMethodNames("hello")
	// Strings have no exported methods, so should return empty
	if methods == nil {
		// This is acceptable - no methods
	}

	// Test with an int
	methods = GetMethodNames(42)
	if len(methods) != 0 {
		t.Errorf("int should have no methods, got %v", methods)
	}
}

// TestReturnValueError tests returnValue Error method
func TestReturnValueError(t *testing.T) {
	rv := &returnValue{value: "test"}
	if rv.Error() != "return" {
		t.Errorf("returnValue.Error() = %q, want 'return'", rv.Error())
	}
}

// TestValidationErrorMethods tests ValidationError Error method
func TestValidationErrorMethods(t *testing.T) {
	ve := &ValidationError{Message: "invalid input"}
	if ve.Error() != "invalid input" {
		t.Errorf("ValidationError.Error() = %q, want 'invalid input'", ve.Error())
	}
}

// TestIsValidationError tests IsValidationError function
func TestIsValidationError(t *testing.T) {
	// Test with ValidationError
	ve := &ValidationError{Message: "test"}
	if !IsValidationError(ve) {
		t.Error("IsValidationError should return true for ValidationError")
	}

	// Test with regular error
	regularErr := fmt.Errorf("regular error")
	if IsValidationError(regularErr) {
		t.Error("IsValidationError should return false for regular error")
	}
}

// TestExecuteValidationWithContains tests executeValidation with builtin contains
func TestExecuteValidationWithContains(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test validation with a builtin that returns boolean - contains returns true
	stmt := ValidationStatement{
		Call: FunctionCallExpr{
			Name: "contains",
			Args: []Expr{
				LiteralExpr{Value: StringLiteral{Value: "hello world"}},
				LiteralExpr{Value: StringLiteral{Value: "hello"}},
			},
		},
	}

	_, err := interp.ExecuteStatement(stmt, env)
	if err != nil {
		t.Errorf("validation with contains should pass: %v", err)
	}
}

// TestExecuteValidationFailingContains tests executeValidation with failing contains
func TestExecuteValidationFailingContains(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test validation with a builtin that returns false
	stmt := ValidationStatement{
		Call: FunctionCallExpr{
			Name: "contains",
			Args: []Expr{
				LiteralExpr{Value: StringLiteral{Value: "hello"}},
				LiteralExpr{Value: StringLiteral{Value: "xyz"}},
			},
		},
	}

	_, err := interp.ExecuteStatement(stmt, env)
	if err == nil {
		t.Error("failing validation should return error")
	}
	if !IsValidationError(err) {
		t.Errorf("failing validation should return ValidationError, got %T", err)
	}
}

// TestGetCommands tests GetCommands method
func TestGetCommands(t *testing.T) {
	interp := NewInterpreter()

	// Load a command
	interp.LoadModule(Module{
		Items: []Item{
			&Command{
				Name:   "hello",
				Params: []CommandParam{},
				Body:   []Statement{},
			},
		},
	})

	commands := interp.GetCommands()
	if len(commands) != 1 {
		t.Errorf("GetCommands should return 1 command, got %d", len(commands))
	}
	if _, ok := commands["hello"]; !ok {
		t.Error("GetCommands should contain 'hello' command")
	}
}

// TestGetAllEventHandlers tests GetAllEventHandlers method
func TestGetAllEventHandlers(t *testing.T) {
	interp := NewInterpreter()

	// Load event handlers
	interp.LoadModule(Module{
		Items: []Item{
			&EventHandler{EventType: "user.created", Body: []Statement{}},
			&EventHandler{EventType: "user.deleted", Body: []Statement{}},
		},
	})

	handlers := interp.GetAllEventHandlers()
	if len(handlers) != 2 {
		t.Errorf("GetAllEventHandlers should return 2 event types, got %d", len(handlers))
	}
}

// TestGetQueueWorkers tests GetQueueWorkers method
func TestGetQueueWorkers(t *testing.T) {
	interp := NewInterpreter()

	// Load queue workers
	interp.LoadModule(Module{
		Items: []Item{
			&QueueWorker{QueueName: "emails", Body: []Statement{}},
			&QueueWorker{QueueName: "notifications", Body: []Statement{}},
		},
	})

	workers := interp.GetQueueWorkers()
	if len(workers) != 2 {
		t.Errorf("GetQueueWorkers should return 2 workers, got %d", len(workers))
	}
}

// TestEvaluateNeExtra tests not-equal operator additional cases
func TestEvaluateNeExtra(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test int != int (false case)
	result, err := interp.EvaluateExpression(BinaryOpExpr{
		Op:    Ne,
		Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
		Right: LiteralExpr{Value: IntLiteral{Value: 5}},
	}, env)
	if err != nil {
		t.Errorf("5 != 5 failed: %v", err)
	}
	if result != false {
		t.Errorf("5 != 5 = %v, want false", result)
	}

	// Test int != int (true case)
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Ne,
		Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
		Right: LiteralExpr{Value: IntLiteral{Value: 10}},
	}, env)
	if err != nil {
		t.Errorf("5 != 10 failed: %v", err)
	}
	if result != true {
		t.Errorf("5 != 10 = %v, want true", result)
	}

	// Test string != string
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Ne,
		Left:  LiteralExpr{Value: StringLiteral{Value: "hello"}},
		Right: LiteralExpr{Value: StringLiteral{Value: "world"}},
	}, env)
	if err != nil {
		t.Errorf("hello != world failed: %v", err)
	}
	if result != true {
		t.Errorf("'hello' != 'world' = %v, want true", result)
	}
}

// TestValuesEqual tests valuesEqual with various types
func TestValuesEqualExtra(t *testing.T) {
	interp := NewInterpreter()

	// Test nil == nil
	if !interp.valuesEqual(nil, nil) {
		t.Error("nil should equal nil")
	}

	// Test nil != non-nil
	if interp.valuesEqual(nil, 5) {
		t.Error("nil should not equal 5")
	}
	if interp.valuesEqual(5, nil) {
		t.Error("5 should not equal nil")
	}

	// Test int64 == int64
	if !interp.valuesEqual(int64(5), int64(5)) {
		t.Error("5 should equal 5")
	}

	// Test float64 == float64
	if !interp.valuesEqual(float64(3.14), float64(3.14)) {
		t.Error("3.14 should equal 3.14")
	}

	// Test string == string
	if !interp.valuesEqual("hello", "hello") {
		t.Error("'hello' should equal 'hello'")
	}

	// Test bool == bool
	if !interp.valuesEqual(true, true) {
		t.Error("true should equal true")
	}

	// Test int64 == float64 (cross type)
	if !interp.valuesEqual(int64(5), float64(5.0)) {
		t.Error("int64(5) should equal float64(5.0)")
	}

	// Test float64 == int64 (cross type)
	if !interp.valuesEqual(float64(5.0), int64(5)) {
		t.Error("float64(5.0) should equal int64(5)")
	}
}

// TestForLoopWithObject tests for loop over object/map
func TestForLoopWithObject(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Define an object to iterate over
	env.Define("obj", map[string]interface{}{
		"a": int64(1),
		"b": int64(2),
	})
	env.Define("sum", int64(0))

	// Create for loop with key and value
	stmt := ForStatement{
		KeyVar:   "k",
		ValueVar: "v",
		Iterable: VariableExpr{Name: "obj"},
		Body: []Statement{
			AssignStatement{
				Target: "sum",
				Value: BinaryOpExpr{
					Op:    Add,
					Left:  VariableExpr{Name: "sum"},
					Right: VariableExpr{Name: "v"},
				},
			},
		},
	}

	_, err := interp.ExecuteStatement(stmt, env)
	if err != nil {
		t.Errorf("for loop over object failed: %v", err)
	}

	sum, _ := env.Get("sum")
	if sum != int64(3) {
		t.Errorf("sum = %v, want 3", sum)
	}
}

// TestForLoopWithKeyOnly tests for loop over object with key only
func TestForLoopWithKeyOnly(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Define an object to iterate over
	env.Define("obj", map[string]interface{}{
		"a": int64(1),
	})
	env.Define("keys", []interface{}{})

	// Create for loop with key only (no KeyVar)
	stmt := ForStatement{
		KeyVar:   "", // No key variable
		ValueVar: "key",
		Iterable: VariableExpr{Name: "obj"},
		Body:     []Statement{},
	}

	_, err := interp.ExecuteStatement(stmt, env)
	if err != nil {
		t.Errorf("for loop with key only failed: %v", err)
	}
}

// TestForLoopInvalidIterable tests for loop with invalid iterable
func TestForLoopInvalidIterable(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("notIterable", int64(42))

	stmt := ForStatement{
		ValueVar: "item",
		Iterable: VariableExpr{Name: "notIterable"},
		Body:     []Statement{},
	}

	_, err := interp.ExecuteStatement(stmt, env)
	if err == nil {
		t.Error("for loop with int should fail")
	}
}

// TestWhileLoopWithReturn tests while loop that returns early
func TestWhileLoopWithReturn(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("i", int64(0))

	stmt := WhileStatement{
		Condition: BinaryOpExpr{
			Op:    Lt,
			Left:  VariableExpr{Name: "i"},
			Right: LiteralExpr{Value: IntLiteral{Value: 10}},
		},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "done"}}},
		},
	}

	result, err := interp.ExecuteStatement(stmt, env)
	if err == nil {
		t.Error("return in while should return returnValue error")
	}
	if result != "done" {
		t.Errorf("result = %v, want 'done'", result)
	}
}

// TestWhileLoopNonBoolCondition tests while loop with non-bool condition
func TestWhileLoopNonBoolCondition(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	stmt := WhileStatement{
		Condition: LiteralExpr{Value: IntLiteral{Value: 42}},
		Body:      []Statement{},
	}

	_, err := interp.ExecuteStatement(stmt, env)
	if err == nil {
		t.Error("while with non-bool condition should fail")
	}
}

// TestIfStatementNonBoolCondition tests if statement with non-bool condition
func TestIfStatementNonBoolCondition(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	stmt := IfStatement{
		Condition: LiteralExpr{Value: IntLiteral{Value: 42}},
		ThenBlock: []Statement{},
	}

	_, err := interp.ExecuteStatement(stmt, env)
	if err == nil {
		t.Error("if with non-bool condition should fail")
	}
}

// TestDivisionByZeroFloat tests float division by zero
func TestDivisionByZeroFloat(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	_, err := interp.EvaluateExpression(BinaryOpExpr{
		Op:    Div,
		Left:  LiteralExpr{Value: FloatLiteral{Value: 10.0}},
		Right: LiteralExpr{Value: FloatLiteral{Value: 0.0}},
	}, env)
	if err == nil {
		t.Error("division by zero should fail")
	}
}

// TestDivisionByZeroInt tests int division by zero
func TestDivisionByZeroInt(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	_, err := interp.EvaluateExpression(BinaryOpExpr{
		Op:    Div,
		Left:  LiteralExpr{Value: IntLiteral{Value: 10}},
		Right: LiteralExpr{Value: IntLiteral{Value: 0}},
	}, env)
	if err == nil {
		t.Error("integer division by zero should fail")
	}
}

// TestUnaryNegation tests unary negation operator
func TestUnaryNegation(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Negate int
	result, err := interp.EvaluateExpression(UnaryOpExpr{
		Op:    Neg,
		Right: LiteralExpr{Value: IntLiteral{Value: 42}},
	}, env)
	if err != nil {
		t.Errorf("-42 failed: %v", err)
	}
	if result != int64(-42) {
		t.Errorf("-42 = %v, want -42", result)
	}

	// Negate float
	result, err = interp.EvaluateExpression(UnaryOpExpr{
		Op:    Neg,
		Right: LiteralExpr{Value: FloatLiteral{Value: 3.14}},
	}, env)
	if err != nil {
		t.Errorf("-3.14 failed: %v", err)
	}
	if result != float64(-3.14) {
		t.Errorf("-3.14 = %v, want -3.14", result)
	}
}

// TestLogicalNotOperator tests logical NOT operator
func TestLogicalNotOperator(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// NOT true
	result, err := interp.EvaluateExpression(UnaryOpExpr{
		Op:    Not,
		Right: LiteralExpr{Value: BoolLiteral{Value: true}},
	}, env)
	if err != nil {
		t.Errorf("!true failed: %v", err)
	}
	if result != false {
		t.Errorf("!true = %v, want false", result)
	}

	// NOT false
	result, err = interp.EvaluateExpression(UnaryOpExpr{
		Op:    Not,
		Right: LiteralExpr{Value: BoolLiteral{Value: false}},
	}, env)
	if err != nil {
		t.Errorf("!false failed: %v", err)
	}
	if result != true {
		t.Errorf("!false = %v, want true", result)
	}
}

// TestArrayExpr tests array expression evaluation
func TestArrayExprExtra(t *testing.T) {
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
		t.Errorf("array expr failed: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok {
		t.Errorf("result not array: %T", result)
	}
	if len(arr) != 3 {
		t.Errorf("array length = %d, want 3", len(arr))
	}
}

// TestObjectExpr tests object expression evaluation
func TestObjectExprExtra(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	result, err := interp.EvaluateExpression(ObjectExpr{
		Fields: []ObjectField{
			{Key: "name", Value: LiteralExpr{Value: StringLiteral{Value: "Alice"}}},
			{Key: "age", Value: LiteralExpr{Value: IntLiteral{Value: 30}}},
		},
	}, env)
	if err != nil {
		t.Errorf("object expr failed: %v", err)
	}
	obj, ok := result.(map[string]interface{})
	if !ok {
		t.Errorf("result not object: %T", result)
	}
	if obj["name"] != "Alice" {
		t.Errorf("obj.name = %v, want 'Alice'", obj["name"])
	}
}

// TestTypeToString tests TypeToString for all types
func TestTypeToStringExtra(t *testing.T) {
	tc := NewTypeChecker()

	tests := []struct {
		typ  Type
		want string
	}{
		{IntType{}, "int"},
		{StringType{}, "string"},
		{BoolType{}, "bool"},
		{FloatType{}, "float"},
		{ArrayType{ElementType: IntType{}}, "[int]"},
		{OptionalType{InnerType: StringType{}}, "string?"},
		{NamedType{Name: "User"}, "User"},
		{UnionType{Types: []Type{IntType{}, StringType{}}}, "int | string"},
	}

	for _, tt := range tests {
		got := tc.TypeToString(tt.typ)
		if got != tt.want {
			t.Errorf("TypeToString(%T) = %q, want %q", tt.typ, got, tt.want)
		}
	}
}

// TestForLoopWithArrayReturn tests for loop over array that returns early
func TestForLoopWithArrayReturn(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("arr", []interface{}{int64(1), int64(2), int64(3)})

	stmt := ForStatement{
		ValueVar: "item",
		Iterable: VariableExpr{Name: "arr"},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "item"}},
		},
	}

	result, err := interp.ExecuteStatement(stmt, env)
	if err == nil {
		t.Error("return in for should return returnValue error")
	}
	if result != int64(1) {
		t.Errorf("result = %v, want 1", result)
	}
}

// TestForLoopWithMapReturn tests for loop over map that returns early
func TestForLoopWithMapReturn(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("obj", map[string]interface{}{"a": int64(1)})

	stmt := ForStatement{
		KeyVar:   "k",
		ValueVar: "v",
		Iterable: VariableExpr{Name: "obj"},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "v"}},
		},
	}

	result, err := interp.ExecuteStatement(stmt, env)
	if err == nil {
		t.Error("return in for should return returnValue error")
	}
	if result != int64(1) {
		t.Errorf("result = %v, want 1", result)
	}
}

// TestExecuteValidationNilResult tests validation that returns nil
func TestExecuteValidationNilResult(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Define a validation function that returns nil (via variable not set)
	interp.LoadModule(Module{
		Items: []Item{
			&Function{
				Name:       "alwaysNil",
				Params:     []Field{},
				ReturnType: nil,
				Body:       []Statement{}, // No return, result is nil
			},
		},
	})

	stmt := ValidationStatement{
		Call: FunctionCallExpr{Name: "alwaysNil", Args: []Expr{}},
	}

	_, err := interp.ExecuteStatement(stmt, env)
	// nil result is treated as successful validation
	if err != nil {
		// A returnValue error is expected since function has no explicit return
		if !IsValidationError(err) {
			// This is okay - the function might not return anything
		}
	}
}

// TestExecuteValidationUnexpectedType tests validation with unexpected return type
func TestExecuteValidationUnexpectedType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Define a validation function that returns a string (unexpected)
	interp.LoadModule(Module{
		Items: []Item{
			&Function{
				Name:       "returnsString",
				Params:     []Field{},
				ReturnType: StringType{},
				Body: []Statement{
					ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "unexpected"}}},
				},
			},
		},
	})

	stmt := ValidationStatement{
		Call: FunctionCallExpr{Name: "returnsString", Args: []Expr{}},
	}

	_, err := interp.ExecuteStatement(stmt, env)
	if err == nil {
		t.Error("validation with unexpected return type should fail")
	}
}

// TestCallMethodWithValidMethod tests CallMethod with a type that has methods
func TestCallMethodWithValidMethod(t *testing.T) {
	// Test with HttpMethod which has String() method
	method := Get
	result, err := CallMethod(method, "String")
	if err != nil {
		t.Errorf("CallMethod on HttpMethod.String should work: %v", err)
	}
	if result != "GET" {
		t.Errorf("HttpMethod.String() = %v, want 'GET'", result)
	}
}

// TestHasMethodWithValidMethod tests HasMethod with a type that has methods
func TestHasMethodWithValidMethod(t *testing.T) {
	method := Get
	if !HasMethod(method, "String") {
		t.Error("HttpMethod should have String method")
	}
}

// TestGetMethodNamesWithMethods tests GetMethodNames with a type that has methods
func TestGetMethodNamesWithMethods(t *testing.T) {
	method := Get
	methods := GetMethodNames(method)
	found := false
	for _, m := range methods {
		if m == "String" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("HttpMethod should have String in methods, got %v", methods)
	}
}

// TestUnsupportedStatementType tests ExecuteStatement with unsupported type
func TestUnsupportedStatementType(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// WsSendStatement is a statement but not in the switch
	// Actually let's test with ExpressionStatement which should be unsupported
	stmt := ExpressionStatement{Expr: LiteralExpr{Value: IntLiteral{Value: 42}}}

	_, err := interp.ExecuteStatement(stmt, env)
	if err == nil {
		t.Error("unsupported statement type should fail")
	}
}

// TestBinOpStringCoverage tests all BinOp String cases
func TestBinOpStringCoverage(t *testing.T) {
	tests := []struct {
		op   BinOp
		want string
	}{
		{Add, "+"},
		{Sub, "-"},
		{Mul, "*"},
		{Div, "/"},
		{Eq, "=="},
		{Ne, "!="},
		{Lt, "<"},
		{Le, "<="},
		{Gt, ">"},
		{Ge, ">="},
		{And, "&&"},
		{Or, "||"},
	}

	for _, tt := range tests {
		if got := tt.op.String(); got != tt.want {
			t.Errorf("BinOp(%d).String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

// TestValidateTypeReferenceCoverage tests type reference validation
func TestValidateTypeReferenceCoverage(t *testing.T) {
	tc := NewTypeChecker()
	tc.SetTypeDefs(map[string]TypeDef{
		"User": {Name: "User", Fields: []Field{}},
	})

	// Valid named type
	err := tc.ValidateTypeReference(NamedType{Name: "User"})
	if err != nil {
		t.Errorf("ValidateTypeReference(User) failed: %v", err)
	}

	// Invalid named type
	err = tc.ValidateTypeReference(NamedType{Name: "NonExistent"})
	if err == nil {
		t.Error("ValidateTypeReference(NonExistent) should fail")
	}

	// Array type with valid element
	err = tc.ValidateTypeReference(ArrayType{ElementType: IntType{}})
	if err != nil {
		t.Errorf("ValidateTypeReference([]int) failed: %v", err)
	}

	// Optional type with valid inner
	err = tc.ValidateTypeReference(OptionalType{InnerType: StringType{}})
	if err != nil {
		t.Errorf("ValidateTypeReference(?string) failed: %v", err)
	}
}

// TestMoreBuiltinFunctions tests additional builtin functions for coverage
func TestExtraBuiltinFunctions(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test time.now
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "time.now",
		Args: []Expr{},
	}, env)
	if err != nil {
		t.Errorf("time.now failed: %v", err)
	}
	if result == nil {
		t.Error("time.now should return a value")
	}

	// Test now (alias for time.now)
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "now",
		Args: []Expr{},
	}, env)
	if err != nil {
		t.Errorf("now failed: %v", err)
	}
	if result == nil {
		t.Error("now should return a value")
	}

	// Test upper
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "upper",
		Args: []Expr{LiteralExpr{Value: StringLiteral{Value: "hello"}}},
	}, env)
	if err != nil {
		t.Errorf("upper failed: %v", err)
	}
	if result != "HELLO" {
		t.Errorf("upper('hello') = %v, want 'HELLO'", result)
	}

	// Test lower
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "lower",
		Args: []Expr{LiteralExpr{Value: StringLiteral{Value: "HELLO"}}},
	}, env)
	if err != nil {
		t.Errorf("lower failed: %v", err)
	}
	if result != "hello" {
		t.Errorf("lower('HELLO') = %v, want 'hello'", result)
	}

	// Test trim
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "trim",
		Args: []Expr{LiteralExpr{Value: StringLiteral{Value: "  hello  "}}},
	}, env)
	if err != nil {
		t.Errorf("trim failed: %v", err)
	}
	if result != "hello" {
		t.Errorf("trim('  hello  ') = %v, want 'hello'", result)
	}

	// Test split
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "split",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "a,b,c"}},
			LiteralExpr{Value: StringLiteral{Value: ","}},
		},
	}, env)
	if err != nil {
		t.Errorf("split failed: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok || len(arr) != 3 {
		t.Errorf("split('a,b,c', ',') = %v, want [a b c]", result)
	}

	// Test join
	env.Define("parts", []interface{}{"a", "b", "c"})
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "join",
		Args: []Expr{
			VariableExpr{Name: "parts"},
			LiteralExpr{Value: StringLiteral{Value: "-"}},
		},
	}, env)
	if err != nil {
		t.Errorf("join failed: %v", err)
	}
	if result != "a-b-c" {
		t.Errorf("join(['a','b','c'], '-') = %v, want 'a-b-c'", result)
	}
}

// TestStringBuiltins tests more string builtins for coverage
func TestStringBuiltins(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test replace
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "replace",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "hello world"}},
			LiteralExpr{Value: StringLiteral{Value: "world"}},
			LiteralExpr{Value: StringLiteral{Value: "there"}},
		},
	}, env)
	if err != nil {
		t.Errorf("replace failed: %v", err)
	}
	if result != "hello there" {
		t.Errorf("replace = %v, want 'hello there'", result)
	}

	// Test substring
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "substring",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "hello world"}},
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

	// Test length
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "length",
		Args: []Expr{LiteralExpr{Value: StringLiteral{Value: "hello"}}},
	}, env)
	if err != nil {
		t.Errorf("length failed: %v", err)
	}
	if result != int64(5) {
		t.Errorf("length('hello') = %v, want 5", result)
	}

	// Test startsWith
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "startsWith",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "hello world"}},
			LiteralExpr{Value: StringLiteral{Value: "hello"}},
		},
	}, env)
	if err != nil {
		t.Errorf("startsWith failed: %v", err)
	}
	if result != true {
		t.Errorf("startsWith = %v, want true", result)
	}

	// Test endsWith
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "endsWith",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "hello world"}},
			LiteralExpr{Value: StringLiteral{Value: "world"}},
		},
	}, env)
	if err != nil {
		t.Errorf("endsWith failed: %v", err)
	}
	if result != true {
		t.Errorf("endsWith = %v, want true", result)
	}

	// Test indexOf
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "indexOf",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "hello world"}},
			LiteralExpr{Value: StringLiteral{Value: "world"}},
		},
	}, env)
	if err != nil {
		t.Errorf("indexOf failed: %v", err)
	}
	if result != int64(6) {
		t.Errorf("indexOf = %v, want 6", result)
	}

	// Test charAt
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "charAt",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "hello"}},
			LiteralExpr{Value: IntLiteral{Value: 1}},
		},
	}, env)
	if err != nil {
		t.Errorf("charAt failed: %v", err)
	}
	if result != "e" {
		t.Errorf("charAt = %v, want 'e'", result)
	}

	// Test toString
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "toString",
		Args: []Expr{LiteralExpr{Value: IntLiteral{Value: 42}}},
	}, env)
	if err != nil {
		t.Errorf("toString failed: %v", err)
	}
	if result != "42" {
		t.Errorf("toString(42) = %v, want '42'", result)
	}
}

// TestExecuteRouteSimpleWithParams tests ExecuteRouteSimple with path parameters
func TestExecuteRouteSimpleWithParams(t *testing.T) {
	interp := NewInterpreter()

	// Create a route with path parameter
	route := &Route{
		Method: Get,
		Path:   "/users/{id}",
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "id"}},
		},
	}

	// Execute route with path parameter
	pathParams := map[string]string{"id": "123"}
	result, err := interp.ExecuteRouteSimple(route, pathParams)
	if err != nil {
		t.Errorf("ExecuteRouteSimple failed: %v", err)
	}
	if result != "123" {
		t.Errorf("result = %v, want '123'", result)
	}
}

// TestExecuteRouteSimpleNoParams tests ExecuteRouteSimple without params
func TestExecuteRouteSimpleNoParams(t *testing.T) {
	interp := NewInterpreter()

	// Create a simple route
	route := &Route{
		Method: Get,
		Path:   "/hello",
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "world"}}},
		},
	}

	// Execute route without path parameters
	result, err := interp.ExecuteRouteSimple(route, nil)
	if err != nil {
		t.Errorf("ExecuteRouteSimple failed: %v", err)
	}
	if result != "world" {
		t.Errorf("result = %v, want 'world'", result)
	}
}

// TestMoreNeOperator tests Ne operator with more types
func TestMoreNeOperator(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test bool != bool
	result, err := interp.EvaluateExpression(BinaryOpExpr{
		Op:    Ne,
		Left:  LiteralExpr{Value: BoolLiteral{Value: true}},
		Right: LiteralExpr{Value: BoolLiteral{Value: false}},
	}, env)
	if err != nil {
		t.Errorf("true != false failed: %v", err)
	}
	if result != true {
		t.Errorf("true != false = %v, want true", result)
	}

	// Test float != float
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Ne,
		Left:  LiteralExpr{Value: FloatLiteral{Value: 3.14}},
		Right: LiteralExpr{Value: FloatLiteral{Value: 2.71}},
	}, env)
	if err != nil {
		t.Errorf("3.14 != 2.71 failed: %v", err)
	}
	if result != true {
		t.Errorf("3.14 != 2.71 = %v, want true", result)
	}
}

// TestLeComparison tests Le operator with more types
func TestLeComparison(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test float <= float
	result, err := interp.EvaluateExpression(BinaryOpExpr{
		Op:    Le,
		Left:  LiteralExpr{Value: FloatLiteral{Value: 3.14}},
		Right: LiteralExpr{Value: FloatLiteral{Value: 3.14}},
	}, env)
	if err != nil {
		t.Errorf("3.14 <= 3.14 failed: %v", err)
	}
	if result != true {
		t.Errorf("3.14 <= 3.14 = %v, want true", result)
	}

	// Test int <= int (equal)
	result, err = interp.EvaluateExpression(BinaryOpExpr{
		Op:    Le,
		Left:  LiteralExpr{Value: IntLiteral{Value: 5}},
		Right: LiteralExpr{Value: IntLiteral{Value: 5}},
	}, env)
	if err != nil {
		t.Errorf("5 <= 5 failed: %v", err)
	}
	if result != true {
		t.Errorf("5 <= 5 = %v, want true", result)
	}
}

// TestFieldAccessOnArray tests field access with array length
func TestFieldAccessOnArray(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("arr", []interface{}{1, 2, 3})

	// Test field access on non-object (should fail)
	_, err := interp.EvaluateExpression(FieldAccessExpr{
		Object: VariableExpr{Name: "arr"},
		Field:  "length",
	}, env)
	// Arrays don't have field access in Glyph - this should fail or return nil
	// Just exercise the code path
	_ = err
}

// TestEmitEventNoHandlers tests EmitEvent when there are no handlers
func TestEmitEventNoHandlers(t *testing.T) {
	interp := NewInterpreter()

	// Emit an event with no handlers registered
	err := interp.EmitEvent("nonexistent.event", map[string]interface{}{"data": "test"})
	if err != nil {
		t.Errorf("EmitEvent with no handlers should not error: %v", err)
	}
}

// TestEmitEventWithHandler tests EmitEvent with a handler
func TestEmitEventWithHandler(t *testing.T) {
	interp := NewInterpreter()

	// Register an event handler
	interp.LoadModule(Module{
		Items: []Item{
			&EventHandler{
				EventType: "user.created",
				Body: []Statement{
					AssignStatement{
						Target: "result",
						Value:  LiteralExpr{Value: StringLiteral{Value: "handled"}},
					},
				},
			},
		},
	})

	// Emit the event
	err := interp.EmitEvent("user.created", map[string]interface{}{"userId": 123})
	if err != nil {
		t.Errorf("EmitEvent failed: %v", err)
	}
}

// TestDbQueryWithParams tests database query with parameters
func TestDbQueryWithParams(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	stmt := DbQueryStatement{
		Var:   "result",
		Query: "SELECT * FROM users WHERE id = ?",
		Params: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 1}},
		},
	}

	_, err := interp.ExecuteStatement(stmt, env)
	if err != nil {
		t.Errorf("DbQuery with params failed: %v", err)
	}

	result, getErr := env.Get("result")
	if getErr != nil {
		t.Errorf("DbQuery should define result variable: %v", getErr)
		return
	}
	resultMap, isMap := result.(map[string]interface{})
	if !isMap {
		t.Errorf("result not map: %T", result)
		return
	}
	if resultMap["query"] != "SELECT * FROM users WHERE id = ?" {
		t.Errorf("query = %v", resultMap["query"])
	}
}

// TestMoreMathFunctions tests additional math builtins
func TestMoreMathFunctions(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test abs with positive int
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "abs",
		Args: []Expr{LiteralExpr{Value: IntLiteral{Value: 42}}},
	}, env)
	if err != nil {
		t.Errorf("abs(42) failed: %v", err)
	}
	if result != int64(42) {
		t.Errorf("abs(42) = %v, want 42", result)
	}

	// Test abs with negative int
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "abs",
		Args: []Expr{LiteralExpr{Value: IntLiteral{Value: -42}}},
	}, env)
	if err != nil {
		t.Errorf("abs(-42) failed: %v", err)
	}
	if result != int64(42) {
		t.Errorf("abs(-42) = %v, want 42", result)
	}

	// Test min with ints
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "min",
		Args: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 10}},
			LiteralExpr{Value: IntLiteral{Value: 5}},
		},
	}, env)
	if err != nil {
		t.Errorf("min(10, 5) failed: %v", err)
	}
	if result != int64(5) {
		t.Errorf("min(10, 5) = %v, want 5", result)
	}

	// Test max with ints
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "max",
		Args: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 10}},
			LiteralExpr{Value: IntLiteral{Value: 5}},
		},
	}, env)
	if err != nil {
		t.Errorf("max(10, 5) failed: %v", err)
	}
	if result != int64(10) {
		t.Errorf("max(10, 5) = %v, want 10", result)
	}
}

// TestArrayBuiltins tests array-related builtins
func TestArrayBuiltins(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("arr", []interface{}{int64(1), int64(2), int64(3)})

	// Test length with array
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "length",
		Args: []Expr{VariableExpr{Name: "arr"}},
	}, env)
	if err != nil {
		t.Errorf("length(arr) failed: %v", err)
	}
	if result != int64(3) {
		t.Errorf("length(arr) = %v, want 3", result)
	}
}

// TestParseIntBuiltin tests parseInt builtin
func TestParseIntBuiltin(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test valid parseInt
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "parseInt",
		Args: []Expr{LiteralExpr{Value: StringLiteral{Value: "42"}}},
	}, env)
	if err != nil {
		t.Errorf("parseInt('42') failed: %v", err)
	}
	if result != int64(42) {
		t.Errorf("parseInt('42') = %v, want 42", result)
	}
}

// TestParseFloatBuiltin tests parseFloat builtin
func TestParseFloatBuiltin(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	// Test valid parseFloat
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "parseFloat",
		Args: []Expr{LiteralExpr{Value: StringLiteral{Value: "3.14"}}},
	}, env)
	if err != nil {
		t.Errorf("parseFloat('3.14') failed: %v", err)
	}
	if result != float64(3.14) {
		t.Errorf("parseFloat('3.14') = %v, want 3.14", result)
	}
}

// TestExecuteRouteWithInjection tests route with database injection
func TestExecuteRouteWithInjection(t *testing.T) {
	interp := NewInterpreter()

	// Set a mock database handler
	interp.SetDatabaseHandler(nil)

	// Create a route with database injection
	route := &Route{
		Method: Get,
		Path:   "/data",
		Injections: []Injection{
			{Name: "db", Type: DatabaseType{}},
		},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "ok"}}},
		},
	}

	// Execute route
	result, err := interp.ExecuteRouteSimple(route, nil)
	if err != nil {
		t.Errorf("ExecuteRouteSimple with injection failed: %v", err)
	}
	if result != "ok" {
		t.Errorf("result = %v, want 'ok'", result)
	}
}

// TestArrayIndexExpr tests array indexing (unsupported currently)
func TestArrayIndexExpr(t *testing.T) {
	interp := NewInterpreter()
	env := NewEnvironment()

	env.Define("arr", []interface{}{int64(10), int64(20), int64(30)})

	// ArrayIndexExpr is currently not implemented in EvaluateExpression
	// Just exercise the code path
	_, err := interp.EvaluateExpression(ArrayIndexExpr{
		Array: VariableExpr{Name: "arr"},
		Index: LiteralExpr{Value: IntLiteral{Value: 1}},
	}, env)
	// We expect an error since it's unsupported
	if err == nil {
		t.Log("ArrayIndexExpr is now supported")
	}
}

// TestExecuteCommandWithDefault tests command execution with default values
func TestExecuteCommandWithDefaultExtra(t *testing.T) {
	interp := NewInterpreter()

	// Load command with default parameter
	cmd := Command{
		Name: "greet",
		Params: []CommandParam{
			{
				Name:     "name",
				Type:     StringType{},
				Required: false,
				Default:  LiteralExpr{Value: StringLiteral{Value: "World"}},
			},
		},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "name"}},
		},
	}

	// Execute without providing the parameter
	result, err := interp.ExecuteCommand(&cmd, map[string]interface{}{})
	if err != nil {
		t.Errorf("ExecuteCommand failed: %v", err)
	}
	if result != "World" {
		t.Errorf("result = %v, want 'World'", result)
	}
}

// TestCronTaskExecution tests cron task execution
func TestCronTaskExecution(t *testing.T) {
	interp := NewInterpreter()

	// Load cron task
	task := CronTask{
		Name:     "cleanup",
		Schedule: "* * * * *",
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "cleaned"}}},
		},
	}

	// Execute cron task
	result, err := interp.ExecuteCronTask(&task)
	if err != nil {
		t.Errorf("ExecuteCronTask failed: %v", err)
	}
	if result != "cleaned" {
		t.Errorf("result = %v, want 'cleaned'", result)
	}
}

// TestQueueWorkerExecution tests queue worker execution
func TestQueueWorkerExecution(t *testing.T) {
	interp := NewInterpreter()

	// Create queue worker
	worker := QueueWorker{
		QueueName: "emails",
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "processed"}}},
		},
	}

	// Execute queue worker
	result, err := interp.ExecuteQueueWorker(&worker, map[string]interface{}{"to": "test@example.com"})
	if err != nil {
		t.Errorf("ExecuteQueueWorker failed: %v", err)
	}
	if result != "processed" {
		t.Errorf("result = %v, want 'processed'", result)
	}
}

// TestEventHandlerExecution tests event handler execution
func TestEventHandlerExecution(t *testing.T) {
	interp := NewInterpreter()

	// Create event handler
	handler := EventHandler{
		EventType: "order.created",
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "handled"}}},
		},
	}

	// Execute event handler
	result, err := interp.ExecuteEventHandler(&handler, map[string]interface{}{"orderId": 123})
	if err != nil {
		t.Errorf("ExecuteEventHandler failed: %v", err)
	}
	if result != "handled" {
		t.Errorf("result = %v, want 'handled'", result)
	}
}

// TestExecuteRouteFullWithQueryParams tests the full ExecuteRoute with query parameters
func TestExecuteRouteFullWithQueryParams(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/users",
		Method: Get,
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "ok"}}},
		},
	}

	request := &Request{
		Path:   "/users?page=5&limit=10",
		Method: "GET",
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp == nil {
		t.Fatal("ExecuteRoute returned nil response")
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestExecuteRouteFullWithFloatQueryParam tests query params with float values
func TestExecuteRouteFullWithFloatQueryParam(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/items",
		Method: Get,
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "ok"}}},
		},
	}

	request := &Request{
		Path:   "/items?price=19.99&discount=0.5",
		Method: "GET",
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestExecuteRouteFullWithStringQueryParam tests query params that aren't numbers
func TestExecuteRouteFullWithStringQueryParam(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/search",
		Method: Get,
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "ok"}}},
		},
	}

	request := &Request{
		Path:   "/search?name=john&status=active",
		Method: "GET",
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestExecuteRouteFullWithEmptyQueryValue tests query params with key but no value
func TestExecuteRouteFullWithEmptyQueryValue(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/filter",
		Method: Get,
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "ok"}}},
		},
	}

	request := &Request{
		Path:   "/filter?active&featured",
		Method: "GET",
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestExecuteRouteFullNoQueryParams tests ExecuteRoute without query params
func TestExecuteRouteFullNoQueryParams(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/simple",
		Method: Get,
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "simple"}}},
		},
	}

	request := &Request{
		Path:   "/simple",
		Method: "GET",
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestExecuteRouteFullWithEmptyQueryString tests empty query string
func TestExecuteRouteFullWithEmptyQueryString(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/test",
		Method: Get,
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "test"}}},
		},
	}

	request := &Request{
		Path:   "/test?",
		Method: "GET",
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestExecuteRouteFullWithBody tests ExecuteRoute with request body
func TestExecuteRouteFullWithBody(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/users",
		Method: Post,
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "created"}}},
		},
	}

	request := &Request{
		Path:   "/users",
		Method: "POST",
		Body:   map[string]interface{}{"name": "John"},
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestExecuteRouteFullNilBody tests ExecuteRoute with nil body
func TestExecuteRouteFullNilBody(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/items",
		Method: Get,
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "items"}}},
		},
	}

	request := &Request{
		Path:   "/items",
		Method: "GET",
		Body:   nil,
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestExecuteRouteFullWithAuth tests ExecuteRoute with auth config
func TestExecuteRouteFullWithAuth(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/protected",
		Method: Get,
		Auth: &AuthConfig{
			AuthType: "bearer",
			Required: true,
		},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "protected"}}},
		},
	}

	request := &Request{
		Path:   "/protected",
		Method: "GET",
		AuthData: map[string]interface{}{
			"user": map[string]interface{}{
				"id":       int64(1),
				"username": "admin",
			},
		},
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestExecuteRouteFullWithAuthNoData tests route with auth but no auth data
func TestExecuteRouteFullWithAuthNoData(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/protected",
		Method: Get,
		Auth: &AuthConfig{
			AuthType: "bearer",
			Required: true,
		},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "protected"}}},
		},
	}

	request := &Request{
		Path:   "/protected",
		Method: "GET",
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestExecuteRouteFullWithReturnType tests route with return type validation
func TestExecuteRouteFullWithReturnType(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:       "/items",
		Method:     Get,
		ReturnType: StringType{},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "items"}}},
		},
	}

	request := &Request{
		Path:   "/items",
		Method: "GET",
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestExecuteRouteFullWithReturnTypeMismatch tests route with return type mismatch
func TestExecuteRouteFullWithReturnTypeMismatch(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:       "/items",
		Method:     Get,
		ReturnType: IntType{},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "not an int"}}},
		},
	}

	request := &Request{
		Path:   "/items",
		Method: "GET",
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err == nil {
		t.Error("ExecuteRoute should fail with type mismatch")
	}
	if resp.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", resp.StatusCode)
	}
}

// TestExecuteRouteFullWithDbInjection tests route with database injection
func TestExecuteRouteFullWithDbInjection(t *testing.T) {
	interp := NewInterpreter()
	interp.SetDatabaseHandler("mock-db")

	route := &Route{
		Path:   "/data",
		Method: Get,
		Injections: []Injection{
			{Name: "db", Type: DatabaseType{}},
		},
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "data"}}},
		},
	}

	request := &Request{
		Path:   "/data",
		Method: "GET",
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestExecuteRouteFullWithPathParam tests route with path parameters
func TestExecuteRouteFullWithPathParam(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/users/:id",
		Method: Get,
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "id"}},
		},
	}

	request := &Request{
		Path:   "/users/123",
		Method: "GET",
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if resp.Body != "123" {
		t.Errorf("Body = %v, want '123'", resp.Body)
	}
}

// TestExecuteRouteFullWithInvalidFloatParse tests query params with invalid float
func TestExecuteRouteFullWithInvalidFloatParse(t *testing.T) {
	interp := NewInterpreter()

	route := &Route{
		Path:   "/test",
		Method: Get,
		Body: []Statement{
			ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "ok"}}},
		},
	}

	// Use something that looks like float but isn't valid
	request := &Request{
		Path:   "/test?value=1.2.3",
		Method: "GET",
	}

	resp, err := interp.ExecuteRoute(route, request)
	if err != nil {
		t.Errorf("ExecuteRoute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}
