package interpreter

import (
	"testing"
)

// TestGenericFunctionBasic tests basic generic function execution
func TestGenericFunctionBasic(t *testing.T) {
	interp := NewInterpreter()

	// Define a generic identity function: ! identity<T>(x: T): T { > x }
	identityFn := Function{
		Name: "identity",
		TypeParams: []TypeParameter{
			{Name: "T"},
		},
		Params: []Field{
			{Name: "x", TypeAnnotation: TypeParameterType{Name: "T"}, Required: true},
		},
		ReturnType: TypeParameterType{Name: "T"},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "x"}},
		},
	}

	interp.globalEnv.Define("identity", identityFn)
	interp.functions["identity"] = identityFn

	// Test with integer (type inferred)
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "identity",
		Args: []Expr{LiteralExpr{Value: IntLiteral{Value: 42}}},
	}, interp.globalEnv)

	if err != nil {
		t.Fatalf("identity(42) failed: %v", err)
	}
	if result.(int64) != 42 {
		t.Errorf("identity(42) = %v, want 42", result)
	}

	// Test with string (type inferred)
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "identity",
		Args: []Expr{LiteralExpr{Value: StringLiteral{Value: "hello"}}},
	}, interp.globalEnv)

	if err != nil {
		t.Fatalf("identity(\"hello\") failed: %v", err)
	}
	if result.(string) != "hello" {
		t.Errorf("identity(\"hello\") = %v, want \"hello\"", result)
	}
}

// TestGenericFunctionExplicitTypeArgs tests generic functions with explicit type arguments
func TestGenericFunctionExplicitTypeArgs(t *testing.T) {
	interp := NewInterpreter()

	// Define a generic identity function
	identityFn := Function{
		Name: "identity",
		TypeParams: []TypeParameter{
			{Name: "T"},
		},
		Params: []Field{
			{Name: "x", TypeAnnotation: TypeParameterType{Name: "T"}, Required: true},
		},
		ReturnType: TypeParameterType{Name: "T"},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "x"}},
		},
	}

	interp.globalEnv.Define("identity", identityFn)
	interp.functions["identity"] = identityFn

	// Test with explicit type argument <int>
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name:     "identity",
		TypeArgs: []Type{IntType{}},
		Args:     []Expr{LiteralExpr{Value: IntLiteral{Value: 100}}},
	}, interp.globalEnv)

	if err != nil {
		t.Fatalf("identity<int>(100) failed: %v", err)
	}
	if result.(int64) != 100 {
		t.Errorf("identity<int>(100) = %v, want 100", result)
	}
}

// TestGenericFunctionMultipleTypeParams tests generic functions with multiple type parameters
func TestGenericFunctionMultipleTypeParams(t *testing.T) {
	interp := NewInterpreter()

	// Define a generic function: ! pair<T, U>(first: T, second: U): object
	// Returns {first: first, second: second}
	pairFn := Function{
		Name: "pair",
		TypeParams: []TypeParameter{
			{Name: "T"},
			{Name: "U"},
		},
		Params: []Field{
			{Name: "first", TypeAnnotation: TypeParameterType{Name: "T"}, Required: true},
			{Name: "second", TypeAnnotation: TypeParameterType{Name: "U"}, Required: true},
		},
		ReturnType: nil, // Returns object
		Body: []Statement{
			ReturnStatement{Value: ObjectExpr{Fields: []ObjectField{
				{Key: "first", Value: VariableExpr{Name: "first"}},
				{Key: "second", Value: VariableExpr{Name: "second"}},
			}}},
		},
	}

	interp.globalEnv.Define("pair", pairFn)
	interp.functions["pair"] = pairFn

	// Test pair<int, string>(42, "hello")
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name:     "pair",
		TypeArgs: []Type{IntType{}, StringType{}},
		Args: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 42}},
			LiteralExpr{Value: StringLiteral{Value: "hello"}},
		},
	}, interp.globalEnv)

	if err != nil {
		t.Fatalf("pair<int, string>(42, \"hello\") failed: %v", err)
	}

	obj := result.(map[string]interface{})
	if obj["first"].(int64) != 42 {
		t.Errorf("pair.first = %v, want 42", obj["first"])
	}
	if obj["second"].(string) != "hello" {
		t.Errorf("pair.second = %v, want \"hello\"", obj["second"])
	}
}

// TestGenericFunctionTypeInference tests type inference for generic functions
func TestGenericFunctionTypeInference(t *testing.T) {
	interp := NewInterpreter()

	// Define a generic function that adds two values
	// ! first<T>(x: T, y: T): T { > x }
	firstFn := Function{
		Name: "first",
		TypeParams: []TypeParameter{
			{Name: "T"},
		},
		Params: []Field{
			{Name: "x", TypeAnnotation: TypeParameterType{Name: "T"}, Required: true},
			{Name: "y", TypeAnnotation: TypeParameterType{Name: "T"}, Required: true},
		},
		ReturnType: TypeParameterType{Name: "T"},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "x"}},
		},
	}

	interp.globalEnv.Define("first", firstFn)
	interp.functions["first"] = firstFn

	// Test type inference with two integers
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "first",
		Args: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 10}},
			LiteralExpr{Value: IntLiteral{Value: 20}},
		},
	}, interp.globalEnv)

	if err != nil {
		t.Fatalf("first(10, 20) failed: %v", err)
	}
	if result.(int64) != 10 {
		t.Errorf("first(10, 20) = %v, want 10", result)
	}

	// Test type inference with two strings
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name: "first",
		Args: []Expr{
			LiteralExpr{Value: StringLiteral{Value: "alpha"}},
			LiteralExpr{Value: StringLiteral{Value: "beta"}},
		},
	}, interp.globalEnv)

	if err != nil {
		t.Fatalf("first(\"alpha\", \"beta\") failed: %v", err)
	}
	if result.(string) != "alpha" {
		t.Errorf("first(\"alpha\", \"beta\") = %v, want \"alpha\"", result)
	}
}

// TestGenericFunctionWithConstraint tests generic functions with type constraints
func TestGenericFunctionWithConstraint(t *testing.T) {
	interp := NewInterpreter()

	// Define a generic function with Numeric constraint
	// ! double<T: Numeric>(x: T): T { > x }
	doubleFn := Function{
		Name: "double",
		TypeParams: []TypeParameter{
			{Name: "T", Constraint: NamedType{Name: "Numeric"}},
		},
		Params: []Field{
			{Name: "x", TypeAnnotation: TypeParameterType{Name: "T"}, Required: true},
		},
		ReturnType: TypeParameterType{Name: "T"},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "x"}},
		},
	}

	interp.globalEnv.Define("double", doubleFn)
	interp.functions["double"] = doubleFn

	// Test with int (satisfies Numeric)
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name:     "double",
		TypeArgs: []Type{IntType{}},
		Args:     []Expr{LiteralExpr{Value: IntLiteral{Value: 21}}},
	}, interp.globalEnv)

	if err != nil {
		t.Fatalf("double<int>(21) failed: %v", err)
	}
	if result.(int64) != 21 {
		t.Errorf("double<int>(21) = %v, want 21", result)
	}

	// Test with float (satisfies Numeric)
	result, err = interp.EvaluateExpression(FunctionCallExpr{
		Name:     "double",
		TypeArgs: []Type{FloatType{}},
		Args:     []Expr{LiteralExpr{Value: FloatLiteral{Value: 3.14}}},
	}, interp.globalEnv)

	if err != nil {
		t.Fatalf("double<float>(3.14) failed: %v", err)
	}
	if result.(float64) != 3.14 {
		t.Errorf("double<float>(3.14) = %v, want 3.14", result)
	}

	// Test with string (does NOT satisfy Numeric) - should fail
	_, err = interp.EvaluateExpression(FunctionCallExpr{
		Name:     "double",
		TypeArgs: []Type{StringType{}},
		Args:     []Expr{LiteralExpr{Value: StringLiteral{Value: "not a number"}}},
	}, interp.globalEnv)

	if err == nil {
		t.Error("double<string>(\"not a number\") should have failed constraint check")
	}
}

// TestGenericFunctionWrongTypeArgCount tests error when wrong number of type args provided
func TestGenericFunctionWrongTypeArgCount(t *testing.T) {
	interp := NewInterpreter()

	// Define a generic function with two type parameters
	pairFn := Function{
		Name: "pair",
		TypeParams: []TypeParameter{
			{Name: "T"},
			{Name: "U"},
		},
		Params: []Field{
			{Name: "first", TypeAnnotation: TypeParameterType{Name: "T"}, Required: true},
			{Name: "second", TypeAnnotation: TypeParameterType{Name: "U"}, Required: true},
		},
		ReturnType: nil,
		Body: []Statement{
			ReturnStatement{Value: ObjectExpr{Fields: []ObjectField{
				{Key: "first", Value: VariableExpr{Name: "first"}},
				{Key: "second", Value: VariableExpr{Name: "second"}},
			}}},
		},
	}

	interp.globalEnv.Define("pair", pairFn)
	interp.functions["pair"] = pairFn

	// Test with only one type argument (should fail)
	_, err := interp.EvaluateExpression(FunctionCallExpr{
		Name:     "pair",
		TypeArgs: []Type{IntType{}}, // Missing second type arg
		Args: []Expr{
			LiteralExpr{Value: IntLiteral{Value: 1}},
			LiteralExpr{Value: StringLiteral{Value: "two"}},
		},
	}, interp.globalEnv)

	if err == nil {
		t.Error("pair<int>(1, \"two\") should have failed due to wrong type arg count")
	}
}

// TestGenericTypeInstantiation tests generic type definition instantiation
func TestGenericTypeInstantiation(t *testing.T) {
	tc := NewTypeChecker()

	// Define a generic Result<T, E> type
	resultTypeDef := TypeDef{
		Name: "Result",
		TypeParams: []TypeParameter{
			{Name: "T"},
			{Name: "E"},
		},
		Fields: []Field{
			{Name: "value", TypeAnnotation: TypeParameterType{Name: "T"}, Required: false},
			{Name: "error", TypeAnnotation: TypeParameterType{Name: "E"}, Required: false},
		},
	}

	tc.SetTypeDefs(map[string]TypeDef{"Result": resultTypeDef})

	// Instantiate Result<int, string>
	instantiated, err := tc.InstantiateGenericType(resultTypeDef, []Type{IntType{}, StringType{}})
	if err != nil {
		t.Fatalf("InstantiateGenericType failed: %v", err)
	}

	// Verify the instantiated type has correct field types
	if len(instantiated.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(instantiated.Fields))
	}

	// Check value field type is int
	valueField := instantiated.Fields[0]
	if _, ok := valueField.TypeAnnotation.(IntType); !ok {
		t.Errorf("Expected value field to be IntType, got %T", valueField.TypeAnnotation)
	}

	// Check error field type is string
	errorField := instantiated.Fields[1]
	if _, ok := errorField.TypeAnnotation.(StringType); !ok {
		t.Errorf("Expected error field to be StringType, got %T", errorField.TypeAnnotation)
	}
}

// TestGenericFunctionWithArrayType tests generic functions with array type parameters
func TestGenericFunctionWithArrayType(t *testing.T) {
	interp := NewInterpreter()

	// Define: ! length<T>(arr: [T]): int
	lengthFn := Function{
		Name: "length",
		TypeParams: []TypeParameter{
			{Name: "T"},
		},
		Params: []Field{
			{Name: "arr", TypeAnnotation: ArrayType{ElementType: TypeParameterType{Name: "T"}}, Required: true},
		},
		ReturnType: IntType{},
		Body: []Statement{
			// For this test, we'll just return the actual array and check it works
			// In a real impl, we'd have a len() builtin
			ReturnStatement{Value: LiteralExpr{Value: IntLiteral{Value: 3}}},
		},
	}

	interp.globalEnv.Define("length", lengthFn)
	interp.functions["length"] = lengthFn

	// Test with array of integers
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name:     "length",
		TypeArgs: []Type{IntType{}},
		Args: []Expr{
			ArrayExpr{Elements: []Expr{
				LiteralExpr{Value: IntLiteral{Value: 1}},
				LiteralExpr{Value: IntLiteral{Value: 2}},
				LiteralExpr{Value: IntLiteral{Value: 3}},
			}},
		},
	}, interp.globalEnv)

	if err != nil {
		t.Fatalf("length<int>([1,2,3]) failed: %v", err)
	}
	if result.(int64) != 3 {
		t.Errorf("length<int>([1,2,3]) = %v, want 3", result)
	}
}

// TestTypeParameterSubstitution tests the SubstituteTypeParams function
func TestTypeParameterSubstitution(t *testing.T) {
	tc := NewTypeChecker()

	typeArgs := map[string]Type{
		"T": IntType{},
		"U": StringType{},
	}

	// Test substitution of TypeParameterType
	tParam := TypeParameterType{Name: "T"}
	result := tc.SubstituteTypeParams(tParam, typeArgs)
	if _, ok := result.(IntType); !ok {
		t.Errorf("SubstituteTypeParams(T) = %T, want IntType", result)
	}

	// Test substitution in ArrayType
	arrType := ArrayType{ElementType: TypeParameterType{Name: "T"}}
	result = tc.SubstituteTypeParams(arrType, typeArgs)
	arrResult, ok := result.(ArrayType)
	if !ok {
		t.Fatalf("SubstituteTypeParams([T]) = %T, want ArrayType", result)
	}
	if _, ok := arrResult.ElementType.(IntType); !ok {
		t.Errorf("SubstituteTypeParams([T]).ElementType = %T, want IntType", arrResult.ElementType)
	}

	// Test substitution in FunctionType
	fnType := FunctionType{
		ParamTypes: []Type{TypeParameterType{Name: "T"}},
		ReturnType: TypeParameterType{Name: "U"},
	}
	result = tc.SubstituteTypeParams(fnType, typeArgs)
	fnResult, ok := result.(FunctionType)
	if !ok {
		t.Fatalf("SubstituteTypeParams((T) -> U) = %T, want FunctionType", result)
	}
	if _, ok := fnResult.ParamTypes[0].(IntType); !ok {
		t.Errorf("SubstituteTypeParams((T) -> U).ParamTypes[0] = %T, want IntType", fnResult.ParamTypes[0])
	}
	if _, ok := fnResult.ReturnType.(StringType); !ok {
		t.Errorf("SubstituteTypeParams((T) -> U).ReturnType = %T, want StringType", fnResult.ReturnType)
	}
}

// TestTypeSatisfiesConstraint tests constraint checking
func TestTypeSatisfiesConstraint(t *testing.T) {
	tc := NewTypeChecker()

	// Test Numeric constraint
	numericConstraint := NamedType{Name: "Numeric"}
	if !tc.TypeSatisfiesConstraint(IntType{}, numericConstraint) {
		t.Error("int should satisfy Numeric constraint")
	}
	if !tc.TypeSatisfiesConstraint(FloatType{}, numericConstraint) {
		t.Error("float should satisfy Numeric constraint")
	}
	if tc.TypeSatisfiesConstraint(StringType{}, numericConstraint) {
		t.Error("string should NOT satisfy Numeric constraint")
	}
	if tc.TypeSatisfiesConstraint(BoolType{}, numericConstraint) {
		t.Error("bool should NOT satisfy Numeric constraint")
	}

	// Test Comparable constraint
	comparableConstraint := NamedType{Name: "Comparable"}
	if !tc.TypeSatisfiesConstraint(IntType{}, comparableConstraint) {
		t.Error("int should satisfy Comparable constraint")
	}
	if !tc.TypeSatisfiesConstraint(StringType{}, comparableConstraint) {
		t.Error("string should satisfy Comparable constraint")
	}
	if !tc.TypeSatisfiesConstraint(BoolType{}, comparableConstraint) {
		t.Error("bool should satisfy Comparable constraint")
	}

	// Test Any constraint
	anyConstraint := NamedType{Name: "Any"}
	if !tc.TypeSatisfiesConstraint(IntType{}, anyConstraint) {
		t.Error("int should satisfy Any constraint")
	}
	if !tc.TypeSatisfiesConstraint(StringType{}, anyConstraint) {
		t.Error("string should satisfy Any constraint")
	}

	// Test nil constraint (no constraint)
	if !tc.TypeSatisfiesConstraint(IntType{}, nil) {
		t.Error("int should satisfy nil (no) constraint")
	}
}

// TestTypeToStringGeneric tests TypeToString for generic types
func TestTypeToStringGeneric(t *testing.T) {
	tc := NewTypeChecker()

	tests := []struct {
		input    Type
		expected string
	}{
		{TypeParameterType{Name: "T"}, "T"},
		{TypeParameterType{Name: "U"}, "U"},
		{GenericType{BaseType: NamedType{Name: "List"}, TypeArgs: []Type{IntType{}}}, "List<int>"},
		{GenericType{BaseType: NamedType{Name: "Map"}, TypeArgs: []Type{StringType{}, IntType{}}}, "Map<string, int>"},
		{FunctionType{ParamTypes: []Type{IntType{}}, ReturnType: StringType{}}, "(int) -> string"},
		{FunctionType{ParamTypes: []Type{IntType{}, BoolType{}}, ReturnType: FloatType{}}, "(int, bool) -> float"},
	}

	for _, tt := range tests {
		result := tc.TypeToString(tt.input)
		if result != tt.expected {
			t.Errorf("TypeToString(%T) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestGenericFunctionNestedCalls tests calling generic functions from within other generic functions
func TestGenericFunctionNestedCalls(t *testing.T) {
	interp := NewInterpreter()

	// Define identity<T>(x: T): T
	identityFn := Function{
		Name: "identity",
		TypeParams: []TypeParameter{
			{Name: "T"},
		},
		Params: []Field{
			{Name: "x", TypeAnnotation: TypeParameterType{Name: "T"}, Required: true},
		},
		ReturnType: TypeParameterType{Name: "T"},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "x"}},
		},
	}

	interp.globalEnv.Define("identity", identityFn)
	interp.functions["identity"] = identityFn

	// Test nested generic call: identity(identity(42))
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "identity",
		Args: []Expr{
			FunctionCallExpr{
				Name: "identity",
				Args: []Expr{LiteralExpr{Value: IntLiteral{Value: 42}}},
			},
		},
	}, interp.globalEnv)

	if err != nil {
		t.Fatalf("identity(identity(42)) failed: %v", err)
	}
	if result.(int64) != 42 {
		t.Errorf("identity(identity(42)) = %v, want 42", result)
	}
}

// TestInferTypeArguments tests the type inference mechanism
func TestInferTypeArguments(t *testing.T) {
	tc := NewTypeChecker()

	// Define a generic function with single type param
	fn := Function{
		Name: "identity",
		TypeParams: []TypeParameter{
			{Name: "T"},
		},
		Params: []Field{
			{Name: "x", TypeAnnotation: TypeParameterType{Name: "T"}, Required: true},
		},
		ReturnType: TypeParameterType{Name: "T"},
	}

	// Test inference from int argument
	typeArgs, err := tc.InferTypeArguments(fn, []interface{}{int64(42)})
	if err != nil {
		t.Fatalf("InferTypeArguments for int failed: %v", err)
	}
	if len(typeArgs) != 1 {
		t.Fatalf("Expected 1 type arg, got %d", len(typeArgs))
	}
	if _, ok := typeArgs[0].(IntType); !ok {
		t.Errorf("Inferred type = %T, want IntType", typeArgs[0])
	}

	// Test inference from string argument
	typeArgs, err = tc.InferTypeArguments(fn, []interface{}{"hello"})
	if err != nil {
		t.Fatalf("InferTypeArguments for string failed: %v", err)
	}
	if _, ok := typeArgs[0].(StringType); !ok {
		t.Errorf("Inferred type = %T, want StringType", typeArgs[0])
	}

	// Test inference from bool argument
	typeArgs, err = tc.InferTypeArguments(fn, []interface{}{true})
	if err != nil {
		t.Fatalf("InferTypeArguments for bool failed: %v", err)
	}
	if _, ok := typeArgs[0].(BoolType); !ok {
		t.Errorf("Inferred type = %T, want BoolType", typeArgs[0])
	}
}

// TestGenericWithBoolType tests generics with bool type
func TestGenericWithBoolType(t *testing.T) {
	interp := NewInterpreter()

	// Define: ! not<T>(x: T): T
	notFn := Function{
		Name: "wrap",
		TypeParams: []TypeParameter{
			{Name: "T"},
		},
		Params: []Field{
			{Name: "x", TypeAnnotation: TypeParameterType{Name: "T"}, Required: true},
		},
		ReturnType: TypeParameterType{Name: "T"},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "x"}},
		},
	}

	interp.globalEnv.Define("wrap", notFn)
	interp.functions["wrap"] = notFn

	// Test with bool
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "wrap",
		Args: []Expr{LiteralExpr{Value: BoolLiteral{Value: true}}},
	}, interp.globalEnv)

	if err != nil {
		t.Fatalf("wrap(true) failed: %v", err)
	}
	if result.(bool) != true {
		t.Errorf("wrap(true) = %v, want true", result)
	}
}

// TestGenericWithFloatType tests generics with float type
func TestGenericWithFloatType(t *testing.T) {
	interp := NewInterpreter()

	identityFn := Function{
		Name: "identity",
		TypeParams: []TypeParameter{
			{Name: "T"},
		},
		Params: []Field{
			{Name: "x", TypeAnnotation: TypeParameterType{Name: "T"}, Required: true},
		},
		ReturnType: TypeParameterType{Name: "T"},
		Body: []Statement{
			ReturnStatement{Value: VariableExpr{Name: "x"}},
		},
	}

	interp.globalEnv.Define("identity", identityFn)
	interp.functions["identity"] = identityFn

	// Test with float
	result, err := interp.EvaluateExpression(FunctionCallExpr{
		Name: "identity",
		Args: []Expr{LiteralExpr{Value: FloatLiteral{Value: 3.14159}}},
	}, interp.globalEnv)

	if err != nil {
		t.Fatalf("identity(3.14159) failed: %v", err)
	}
	if result.(float64) != 3.14159 {
		t.Errorf("identity(3.14159) = %v, want 3.14159", result)
	}
}
