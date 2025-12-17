package interpreter

import (
	"fmt"
	"reflect"
	"strings"
)

// TypeChecker validates type compatibility and performs type checking
type TypeChecker struct {
	typeDefs  map[string]TypeDef
	functions map[string]Function
}

// NewTypeChecker creates a new TypeChecker
func NewTypeChecker() *TypeChecker {
	return &TypeChecker{
		typeDefs:  make(map[string]TypeDef),
		functions: make(map[string]Function),
	}
}

// SetTypeDefs updates the type definitions map
func (tc *TypeChecker) SetTypeDefs(typeDefs map[string]TypeDef) {
	tc.typeDefs = typeDefs
}

// SetFunctions updates the functions map
func (tc *TypeChecker) SetFunctions(functions map[string]Function) {
	tc.functions = functions
}

// CoerceNumeric converts int64 to float64 when needed for numeric operations
// Returns: (coercedLeft, coercedRight, wasCoerced)
func CoerceNumeric(left, right interface{}) (interface{}, interface{}, bool) {
	// If both are int64, no coercion needed
	if _, ok1 := left.(int64); ok1 {
		if _, ok2 := right.(int64); ok2 {
			return left, right, false // both int
		}
	}

	// If one is int64 and other is float64, coerce int to float
	if leftInt, ok := left.(int64); ok {
		if _, ok := right.(float64); ok {
			return float64(leftInt), right, true // left coerced
		}
	}
	if rightInt, ok := right.(int64); ok {
		if _, ok := left.(float64); ok {
			return left, float64(rightInt), true // right coerced
		}
	}

	// No coercion possible or needed
	return left, right, false
}

// GetRuntimeType infers the Type from a runtime value
func GetRuntimeType(value interface{}) Type {
	switch value.(type) {
	case int64:
		return IntType{}
	case string:
		return StringType{}
	case bool:
		return BoolType{}
	case float64:
		return FloatType{}
	case []interface{}:
		// For arrays, we can't determine element type from runtime value alone
		return ArrayType{ElementType: nil}
	case map[string]interface{}:
		// For objects, we return a generic named type
		return NamedType{Name: "object"}
	default:
		return nil
	}
}

// CheckType validates that a value matches an expected type annotation
func (tc *TypeChecker) CheckType(value interface{}, expectedType Type) error {
	if expectedType == nil {
		return nil // No type constraint
	}

	actualType := GetRuntimeType(value)
	if actualType == nil {
		return fmt.Errorf("cannot determine type of value: %T", value)
	}

	// Check type compatibility
	if !tc.TypesCompatible(actualType, expectedType) {
		return fmt.Errorf("type mismatch: expected %s, got %s",
			tc.TypeToString(expectedType), tc.TypeToString(actualType))
	}

	// For array types, validate element types if specified
	if arrayType, ok := expectedType.(ArrayType); ok {
		if arrayType.ElementType != nil {
			if arr, ok := value.([]interface{}); ok {
				for i, elem := range arr {
					if err := tc.CheckType(elem, arrayType.ElementType); err != nil {
						return fmt.Errorf("array element %d: %v", i, err)
					}
				}
			}
		}
	}

	// For named types, validate against TypeDef if it exists
	if namedType, ok := expectedType.(NamedType); ok {
		if typeDef, exists := tc.typeDefs[namedType.Name]; exists {
			if obj, ok := value.(map[string]interface{}); ok {
				return tc.ValidateObjectAgainstTypeDef(obj, typeDef)
			}
		}
	}

	return nil
}

// TypesCompatible checks if two types are compatible
func (tc *TypeChecker) TypesCompatible(actual, expected Type) bool {
	// Nil types are always compatible
	if actual == nil || expected == nil {
		return true
	}

	// Check for exact type match
	if reflect.TypeOf(actual) == reflect.TypeOf(expected) {
		switch a := actual.(type) {
		case IntType:
			return true
		case StringType:
			return true
		case BoolType:
			return true
		case FloatType:
			return true
		case ArrayType:
			e := expected.(ArrayType)
			// If expected has no element type constraint, any array is compatible
			if e.ElementType == nil {
				return true
			}
			// If actual has no element type, we can't validate compatibility
			if a.ElementType == nil {
				return true // Runtime check will validate elements
			}
			// Both have element types, check compatibility
			return tc.TypesCompatible(a.ElementType, e.ElementType)
		case OptionalType:
			e := expected.(OptionalType)
			return tc.TypesCompatible(a.InnerType, e.InnerType)
		case NamedType:
			e := expected.(NamedType)
			// Runtime objects have name "object", but can match any NamedType
			// The structure validation happens in CheckType
			if a.Name == "object" {
				return true
			}
			return a.Name == e.Name
		}
	}

	// Special case: int is compatible with float (auto-coercion)
	if _, ok := actual.(IntType); ok {
		if _, ok := expected.(FloatType); ok {
			return true
		}
	}

	// Special case: OptionalType is compatible with its inner type
	if optType, ok := expected.(OptionalType); ok {
		return tc.TypesCompatible(actual, optType.InnerType)
	}

	// Special case: UnionType - actual value must be compatible with at least one type in the union
	if unionType, ok := expected.(UnionType); ok {
		for _, memberType := range unionType.Types {
			if tc.TypesCompatible(actual, memberType) {
				return true // Compatible with at least one union member
			}
		}
		return false // Not compatible with any union member
	}

	// If actual is a union type, check if all members are compatible with expected
	if unionType, ok := actual.(UnionType); ok {
		for _, memberType := range unionType.Types {
			if !tc.TypesCompatible(memberType, expected) {
				return false // At least one member not compatible
			}
		}
		return true // All members compatible
	}

	return false
}

// TypeToString converts a Type to a human-readable string
func (tc *TypeChecker) TypeToString(t Type) string {
	if t == nil {
		return "any"
	}

	switch typ := t.(type) {
	case IntType:
		return "int"
	case StringType:
		return "string"
	case BoolType:
		return "bool"
	case FloatType:
		return "float"
	case ArrayType:
		if typ.ElementType != nil {
			return fmt.Sprintf("List[%s]", tc.TypeToString(typ.ElementType))
		}
		return "List"
	case OptionalType:
		return fmt.Sprintf("%s?", tc.TypeToString(typ.InnerType))
	case NamedType:
		return typ.Name
	case UnionType:
		if len(typ.Types) == 0 {
			return "never"
		}
		typeStrings := make([]string, len(typ.Types))
		for i, t := range typ.Types {
			typeStrings[i] = tc.TypeToString(t)
		}
		return strings.Join(typeStrings, " | ")
	default:
		return fmt.Sprintf("%T", t)
	}
}

// ResolveType resolves a NamedType to its TypeDef
func (tc *TypeChecker) ResolveType(typeName string) (TypeDef, error) {
	if typeDef, exists := tc.typeDefs[typeName]; exists {
		return typeDef, nil
	}
	return TypeDef{}, fmt.Errorf("undefined type: %s", typeName)
}

// ValidateTypeReference checks if a type reference is valid
func (tc *TypeChecker) ValidateTypeReference(t Type) error {
	if t == nil {
		return nil
	}

	switch typ := t.(type) {
	case NamedType:
		// Check if the named type exists
		if _, exists := tc.typeDefs[typ.Name]; !exists {
			return fmt.Errorf("undefined type: %s", typ.Name)
		}
	case ArrayType:
		// Recursively validate element type
		if typ.ElementType != nil {
			return tc.ValidateTypeReference(typ.ElementType)
		}
	case OptionalType:
		// Recursively validate inner type
		return tc.ValidateTypeReference(typ.InnerType)
	}

	return nil
}

// ValidateObjectAgainstTypeDef validates an object against a TypeDef
func (tc *TypeChecker) ValidateObjectAgainstTypeDef(obj map[string]interface{}, typeDef TypeDef) error {
	// Check required fields
	for _, field := range typeDef.Fields {
		if field.Required {
			if _, exists := obj[field.Name]; !exists {
				return fmt.Errorf("missing required field: %s", field.Name)
			}
		}
	}

	// Validate field types
	for fieldName, fieldValue := range obj {
		// Find the field definition
		var fieldDef *Field
		for i := range typeDef.Fields {
			if typeDef.Fields[i].Name == fieldName {
				fieldDef = &typeDef.Fields[i]
				break
			}
		}

		// If field is not in TypeDef, skip validation (allow extra fields)
		if fieldDef == nil {
			continue
		}

		// Validate the field value against its type
		if err := tc.CheckType(fieldValue, fieldDef.TypeAnnotation); err != nil {
			return fmt.Errorf("field %s: %v", fieldName, err)
		}
	}

	return nil
}

// ValidateArrayElements validates all elements of an array against a type
func (tc *TypeChecker) ValidateArrayElements(elements []interface{}, elementType Type) error {
	if elementType == nil {
		return nil // Untyped array, no validation needed
	}

	for i, elem := range elements {
		if err := tc.CheckType(elem, elementType); err != nil {
			return fmt.Errorf("element %d: %v", i, err)
		}
	}

	return nil
}
