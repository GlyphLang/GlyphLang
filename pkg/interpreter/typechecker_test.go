package interpreter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoerceNumeric_BothInt(t *testing.T) {
	left := int64(5)
	right := int64(3)

	coercedLeft, coercedRight, wasCoerced := CoerceNumeric(left, right)

	assert.Equal(t, int64(5), coercedLeft)
	assert.Equal(t, int64(3), coercedRight)
	assert.False(t, wasCoerced, "both int64, no coercion needed")
}

func TestCoerceNumeric_IntAndFloat(t *testing.T) {
	left := int64(5)
	right := float64(3.2)

	coercedLeft, coercedRight, wasCoerced := CoerceNumeric(left, right)

	assert.Equal(t, float64(5), coercedLeft)
	assert.Equal(t, float64(3.2), coercedRight)
	assert.True(t, wasCoerced, "int64 should be coerced to float64")
}

func TestCoerceNumeric_FloatAndInt(t *testing.T) {
	left := float64(5.5)
	right := int64(3)

	coercedLeft, coercedRight, wasCoerced := CoerceNumeric(left, right)

	assert.Equal(t, float64(5.5), coercedLeft)
	assert.Equal(t, float64(3), coercedRight)
	assert.True(t, wasCoerced, "int64 should be coerced to float64")
}

func TestCoerceNumeric_BothFloat(t *testing.T) {
	left := float64(5.5)
	right := float64(3.2)

	coercedLeft, coercedRight, wasCoerced := CoerceNumeric(left, right)

	assert.Equal(t, float64(5.5), coercedLeft)
	assert.Equal(t, float64(3.2), coercedRight)
	assert.False(t, wasCoerced, "both float64, no coercion needed")
}

func TestCoerceNumeric_NonNumeric(t *testing.T) {
	left := "hello"
	right := int64(5)

	coercedLeft, coercedRight, wasCoerced := CoerceNumeric(left, right)

	assert.Equal(t, "hello", coercedLeft)
	assert.Equal(t, int64(5), coercedRight)
	assert.False(t, wasCoerced, "string cannot be coerced")
}

func TestGetRuntimeType_Int(t *testing.T) {
	typ := GetRuntimeType(int64(42))
	assert.IsType(t, IntType{}, typ)
}

func TestGetRuntimeType_String(t *testing.T) {
	typ := GetRuntimeType("hello")
	assert.IsType(t, StringType{}, typ)
}

func TestGetRuntimeType_Bool(t *testing.T) {
	typ := GetRuntimeType(true)
	assert.IsType(t, BoolType{}, typ)
}

func TestGetRuntimeType_Float(t *testing.T) {
	typ := GetRuntimeType(float64(3.14))
	assert.IsType(t, FloatType{}, typ)
}

func TestGetRuntimeType_Array(t *testing.T) {
	typ := GetRuntimeType([]interface{}{1, 2, 3})
	arrayType, ok := typ.(ArrayType)
	assert.True(t, ok)
	assert.Nil(t, arrayType.ElementType, "runtime array type has no element type constraint")
}

func TestGetRuntimeType_Object(t *testing.T) {
	obj := map[string]interface{}{"name": "Alice"}
	typ := GetRuntimeType(obj)
	namedType, ok := typ.(NamedType)
	assert.True(t, ok)
	assert.Equal(t, "object", namedType.Name)
}

func TestTypeChecker_CheckType_IntMatch(t *testing.T) {
	tc := NewTypeChecker()
	err := tc.CheckType(int64(42), IntType{})
	assert.NoError(t, err)
}

func TestTypeChecker_CheckType_IntMismatch(t *testing.T) {
	tc := NewTypeChecker()
	err := tc.CheckType("hello", IntType{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type mismatch")
	assert.Contains(t, err.Error(), "expected int")
	assert.Contains(t, err.Error(), "got string")
}

func TestTypeChecker_CheckType_StringMatch(t *testing.T) {
	tc := NewTypeChecker()
	err := tc.CheckType("hello", StringType{})
	assert.NoError(t, err)
}

func TestTypeChecker_CheckType_BoolMatch(t *testing.T) {
	tc := NewTypeChecker()
	err := tc.CheckType(true, BoolType{})
	assert.NoError(t, err)
}

func TestTypeChecker_CheckType_FloatMatch(t *testing.T) {
	tc := NewTypeChecker()
	err := tc.CheckType(float64(3.14), FloatType{})
	assert.NoError(t, err)
}

func TestTypeChecker_CheckType_NilType(t *testing.T) {
	tc := NewTypeChecker()
	// Nil type means no constraint
	err := tc.CheckType(int64(42), nil)
	assert.NoError(t, err)
}

func TestTypeChecker_TypesCompatible_ExactMatch(t *testing.T) {
	tc := NewTypeChecker()

	assert.True(t, tc.TypesCompatible(IntType{}, IntType{}))
	assert.True(t, tc.TypesCompatible(StringType{}, StringType{}))
	assert.True(t, tc.TypesCompatible(BoolType{}, BoolType{}))
	assert.True(t, tc.TypesCompatible(FloatType{}, FloatType{}))
}

func TestTypeChecker_TypesCompatible_IntToFloat(t *testing.T) {
	tc := NewTypeChecker()
	// Int is compatible with float (auto-coercion)
	assert.True(t, tc.TypesCompatible(IntType{}, FloatType{}))
}

func TestTypeChecker_TypesCompatible_Mismatch(t *testing.T) {
	tc := NewTypeChecker()

	assert.False(t, tc.TypesCompatible(IntType{}, StringType{}))
	assert.False(t, tc.TypesCompatible(BoolType{}, IntType{}))
	assert.False(t, tc.TypesCompatible(FloatType{}, StringType{}))
}

func TestTypeChecker_TypesCompatible_Arrays(t *testing.T) {
	tc := NewTypeChecker()

	// Array with no element type is compatible with typed array
	untyped := ArrayType{ElementType: nil}
	typed := ArrayType{ElementType: IntType{}}
	assert.True(t, tc.TypesCompatible(untyped, typed))
	assert.True(t, tc.TypesCompatible(typed, untyped))

	// Arrays with matching element types
	intArray1 := ArrayType{ElementType: IntType{}}
	intArray2 := ArrayType{ElementType: IntType{}}
	assert.True(t, tc.TypesCompatible(intArray1, intArray2))

	// Arrays with different element types
	intArray := ArrayType{ElementType: IntType{}}
	strArray := ArrayType{ElementType: StringType{}}
	assert.False(t, tc.TypesCompatible(intArray, strArray))
}

func TestTypeChecker_TypesCompatible_NamedTypes(t *testing.T) {
	tc := NewTypeChecker()

	user1 := NamedType{Name: "User"}
	user2 := NamedType{Name: "User"}
	message := NamedType{Name: "Message"}

	assert.True(t, tc.TypesCompatible(user1, user2))
	assert.False(t, tc.TypesCompatible(user1, message))
}

func TestTypeChecker_TypeToString(t *testing.T) {
	tc := NewTypeChecker()

	assert.Equal(t, "int", tc.TypeToString(IntType{}))
	assert.Equal(t, "string", tc.TypeToString(StringType{}))
	assert.Equal(t, "bool", tc.TypeToString(BoolType{}))
	assert.Equal(t, "float", tc.TypeToString(FloatType{}))
	assert.Equal(t, "List[int]", tc.TypeToString(ArrayType{ElementType: IntType{}}))
	assert.Equal(t, "List", tc.TypeToString(ArrayType{ElementType: nil}))
	assert.Equal(t, "User", tc.TypeToString(NamedType{Name: "User"}))
	assert.Equal(t, "int?", tc.TypeToString(OptionalType{InnerType: IntType{}}))
	assert.Equal(t, "any", tc.TypeToString(nil))
}

func TestTypeChecker_ValidateArrayElements_Homogeneous(t *testing.T) {
	tc := NewTypeChecker()

	elements := []interface{}{int64(1), int64(2), int64(3)}
	err := tc.ValidateArrayElements(elements, IntType{})
	assert.NoError(t, err)
}

func TestTypeChecker_ValidateArrayElements_Mixed(t *testing.T) {
	tc := NewTypeChecker()

	elements := []interface{}{int64(1), "string", int64(3)}
	err := tc.ValidateArrayElements(elements, IntType{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "element 1")
	assert.Contains(t, err.Error(), "type mismatch")
}

func TestTypeChecker_ValidateArrayElements_NilType(t *testing.T) {
	tc := NewTypeChecker()

	// Nil element type means no validation (untyped array)
	elements := []interface{}{int64(1), "string", true}
	err := tc.ValidateArrayElements(elements, nil)
	assert.NoError(t, err)
}

func TestTypeChecker_ValidateObjectAgainstTypeDef_Valid(t *testing.T) {
	tc := NewTypeChecker()

	typeDef := TypeDef{
		Name: "User",
		Fields: []Field{
			{Name: "id", TypeAnnotation: IntType{}, Required: true},
			{Name: "name", TypeAnnotation: StringType{}, Required: true},
			{Name: "age", TypeAnnotation: IntType{}, Required: false},
		},
	}

	obj := map[string]interface{}{
		"id":   int64(1),
		"name": "Alice",
		"age":  int64(30),
	}

	err := tc.ValidateObjectAgainstTypeDef(obj, typeDef)
	assert.NoError(t, err)
}

func TestTypeChecker_ValidateObjectAgainstTypeDef_MissingRequired(t *testing.T) {
	tc := NewTypeChecker()

	typeDef := TypeDef{
		Name: "User",
		Fields: []Field{
			{Name: "id", TypeAnnotation: IntType{}, Required: true},
			{Name: "name", TypeAnnotation: StringType{}, Required: true},
		},
	}

	obj := map[string]interface{}{
		"id": int64(1),
		// "name" is missing
	}

	err := tc.ValidateObjectAgainstTypeDef(obj, typeDef)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field: name")
}

func TestTypeChecker_ValidateObjectAgainstTypeDef_TypeMismatch(t *testing.T) {
	tc := NewTypeChecker()

	typeDef := TypeDef{
		Name: "User",
		Fields: []Field{
			{Name: "id", TypeAnnotation: IntType{}, Required: true},
			{Name: "name", TypeAnnotation: StringType{}, Required: true},
		},
	}

	obj := map[string]interface{}{
		"id":   "not-a-number", // Wrong type
		"name": "Alice",
	}

	err := tc.ValidateObjectAgainstTypeDef(obj, typeDef)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field id")
	assert.Contains(t, err.Error(), "type mismatch")
}

func TestTypeChecker_ValidateObjectAgainstTypeDef_ExtraFields(t *testing.T) {
	tc := NewTypeChecker()

	typeDef := TypeDef{
		Name: "User",
		Fields: []Field{
			{Name: "id", TypeAnnotation: IntType{}, Required: true},
		},
	}

	obj := map[string]interface{}{
		"id":    int64(1),
		"extra": "allowed", // Extra fields are allowed
	}

	err := tc.ValidateObjectAgainstTypeDef(obj, typeDef)
	assert.NoError(t, err, "extra fields should be allowed")
}

func TestTypeChecker_ValidateTypeReference_ValidBuiltin(t *testing.T) {
	tc := NewTypeChecker()

	err := tc.ValidateTypeReference(IntType{})
	assert.NoError(t, err)

	err = tc.ValidateTypeReference(StringType{})
	assert.NoError(t, err)
}

func TestTypeChecker_ValidateTypeReference_ValidNamed(t *testing.T) {
	tc := NewTypeChecker()

	// Register a type definition
	tc.SetTypeDefs(map[string]TypeDef{
		"User": {Name: "User", Fields: []Field{}},
	})

	err := tc.ValidateTypeReference(NamedType{Name: "User"})
	assert.NoError(t, err)
}

func TestTypeChecker_ValidateTypeReference_UndefinedNamed(t *testing.T) {
	tc := NewTypeChecker()

	err := tc.ValidateTypeReference(NamedType{Name: "UndefinedType"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undefined type: UndefinedType")
}

func TestTypeChecker_ValidateTypeReference_ArrayWithInvalidElement(t *testing.T) {
	tc := NewTypeChecker()

	// Array with undefined named type as element
	arrayType := ArrayType{ElementType: NamedType{Name: "UndefinedType"}}

	err := tc.ValidateTypeReference(arrayType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undefined type: UndefinedType")
}

func TestTypeChecker_ResolveType_Valid(t *testing.T) {
	tc := NewTypeChecker()

	userTypeDef := TypeDef{
		Name: "User",
		Fields: []Field{
			{Name: "id", TypeAnnotation: IntType{}, Required: true},
		},
	}

	tc.SetTypeDefs(map[string]TypeDef{
		"User": userTypeDef,
	})

	resolved, err := tc.ResolveType("User")
	require.NoError(t, err)
	assert.Equal(t, "User", resolved.Name)
	assert.Len(t, resolved.Fields, 1)
}

func TestTypeChecker_ResolveType_Undefined(t *testing.T) {
	tc := NewTypeChecker()

	_, err := tc.ResolveType("UndefinedType")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undefined type: UndefinedType")
}

func TestTypeChecker_CheckType_ArrayElements(t *testing.T) {
	tc := NewTypeChecker()

	// Array of ints
	arrayType := ArrayType{ElementType: IntType{}}
	validArray := []interface{}{int64(1), int64(2), int64(3)}

	err := tc.CheckType(validArray, arrayType)
	assert.NoError(t, err)

	// Array with mixed types should fail
	invalidArray := []interface{}{int64(1), "string", int64(3)}
	err = tc.CheckType(invalidArray, arrayType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "array element 1")
}

func TestTypeChecker_CheckType_NestedObject(t *testing.T) {
	tc := NewTypeChecker()

	// Define Address type
	addressTypeDef := TypeDef{
		Name: "Address",
		Fields: []Field{
			{Name: "city", TypeAnnotation: StringType{}, Required: true},
		},
	}

	// Define User type with nested Address
	userTypeDef := TypeDef{
		Name: "User",
		Fields: []Field{
			{Name: "name", TypeAnnotation: StringType{}, Required: true},
			{Name: "address", TypeAnnotation: NamedType{Name: "Address"}, Required: true},
		},
	}

	tc.SetTypeDefs(map[string]TypeDef{
		"Address": addressTypeDef,
		"User":    userTypeDef,
	})

	// Valid nested object
	user := map[string]interface{}{
		"name": "Alice",
		"address": map[string]interface{}{
			"city": "New York",
		},
	}

	err := tc.CheckType(user, NamedType{Name: "User"})
	assert.NoError(t, err)
}
