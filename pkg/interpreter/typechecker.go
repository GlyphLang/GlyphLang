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
	traitDefs map[string]TraitDef
	// typeScope maps type parameter names to their resolved types during generic instantiation
	typeScope map[string]Type
}

// NewTypeChecker creates a new TypeChecker
func NewTypeChecker() *TypeChecker {
	return &TypeChecker{
		typeDefs:  make(map[string]TypeDef),
		functions: make(map[string]Function),
		traitDefs: make(map[string]TraitDef),
		typeScope: make(map[string]Type),
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

// SetTraitDefs updates the trait definitions map
func (tc *TypeChecker) SetTraitDefs(traitDefs map[string]TraitDef) {
	tc.traitDefs = traitDefs
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
			return fmt.Sprintf("[%s]", tc.TypeToString(typ.ElementType))
		}
		return "[]"
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
	case TypeParameterType:
		return typ.Name
	case GenericType:
		base := tc.TypeToString(typ.BaseType)
		if len(typ.TypeArgs) > 0 {
			argStrings := make([]string, len(typ.TypeArgs))
			for i, arg := range typ.TypeArgs {
				argStrings[i] = tc.TypeToString(arg)
			}
			return fmt.Sprintf("%s<%s>", base, strings.Join(argStrings, ", "))
		}
		return base
	case FunctionType:
		paramStrings := make([]string, len(typ.ParamTypes))
		for i, pt := range typ.ParamTypes {
			paramStrings[i] = tc.TypeToString(pt)
		}
		return fmt.Sprintf("(%s) -> %s", strings.Join(paramStrings, ", "), tc.TypeToString(typ.ReturnType))
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
	// Check required fields (fields with defaults are not required)
	for _, field := range typeDef.Fields {
		if field.Required && field.Default == nil {
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

// SubstituteTypeParams substitutes type parameters in a type with their resolved types
// This is used for generic type instantiation
func (tc *TypeChecker) SubstituteTypeParams(t Type, typeArgs map[string]Type) Type {
	if t == nil {
		return nil
	}

	switch typ := t.(type) {
	case TypeParameterType:
		// Substitute the type parameter with its resolved type
		if resolved, ok := typeArgs[typ.Name]; ok {
			return resolved
		}
		// If not found in typeArgs, check tc.typeScope
		if resolved, ok := tc.typeScope[typ.Name]; ok {
			return resolved
		}
		// Return the type parameter as-is if not resolved
		return typ

	case ArrayType:
		return ArrayType{
			ElementType: tc.SubstituteTypeParams(typ.ElementType, typeArgs),
		}

	case OptionalType:
		return OptionalType{
			InnerType: tc.SubstituteTypeParams(typ.InnerType, typeArgs),
		}

	case GenericType:
		// Substitute type arguments in the generic type
		newTypeArgs := make([]Type, len(typ.TypeArgs))
		for i, arg := range typ.TypeArgs {
			newTypeArgs[i] = tc.SubstituteTypeParams(arg, typeArgs)
		}
		return GenericType{
			BaseType: tc.SubstituteTypeParams(typ.BaseType, typeArgs),
			TypeArgs: newTypeArgs,
		}

	case FunctionType:
		// Substitute type parameters in function type
		newParamTypes := make([]Type, len(typ.ParamTypes))
		for i, pt := range typ.ParamTypes {
			newParamTypes[i] = tc.SubstituteTypeParams(pt, typeArgs)
		}
		return FunctionType{
			ParamTypes: newParamTypes,
			ReturnType: tc.SubstituteTypeParams(typ.ReturnType, typeArgs),
		}

	case UnionType:
		newTypes := make([]Type, len(typ.Types))
		for i, ut := range typ.Types {
			newTypes[i] = tc.SubstituteTypeParams(ut, typeArgs)
		}
		return UnionType{Types: newTypes}

	default:
		// Primitive types and other types don't contain type parameters
		return t
	}
}

// InstantiateGenericType creates a concrete type from a generic type definition
func (tc *TypeChecker) InstantiateGenericType(typeDef TypeDef, typeArgs []Type) (TypeDef, error) {
	if len(typeArgs) != len(typeDef.TypeParams) {
		return TypeDef{}, fmt.Errorf("type %s expects %d type arguments, got %d",
			typeDef.Name, len(typeDef.TypeParams), len(typeArgs))
	}

	// Create type argument map
	typeArgMap := make(map[string]Type)
	for i, param := range typeDef.TypeParams {
		typeArgMap[param.Name] = typeArgs[i]
	}

	// Validate constraints
	for i, param := range typeDef.TypeParams {
		if param.Constraint != nil {
			if !tc.TypeSatisfiesConstraint(typeArgs[i], param.Constraint) {
				return TypeDef{}, fmt.Errorf("type argument %s does not satisfy constraint %s",
					tc.TypeToString(typeArgs[i]), tc.TypeToString(param.Constraint))
			}
		}
	}

	// Create new fields with substituted types
	newFields := make([]Field, len(typeDef.Fields))
	for i, field := range typeDef.Fields {
		newFields[i] = Field{
			Name:           field.Name,
			TypeAnnotation: tc.SubstituteTypeParams(field.TypeAnnotation, typeArgMap),
			Required:       field.Required,
			Default:        field.Default,
		}
	}

	return TypeDef{
		Name:   typeDef.Name,
		Fields: newFields,
	}, nil
}

// InstantiateGenericFunction creates a concrete function from a generic function definition
func (tc *TypeChecker) InstantiateGenericFunction(fn Function, typeArgs []Type) (Function, map[string]Type, error) {
	if len(typeArgs) != len(fn.TypeParams) {
		return Function{}, nil, fmt.Errorf("function %s expects %d type arguments, got %d",
			fn.Name, len(fn.TypeParams), len(typeArgs))
	}

	// Create type argument map
	typeArgMap := make(map[string]Type)
	for i, param := range fn.TypeParams {
		typeArgMap[param.Name] = typeArgs[i]
	}

	// Validate constraints
	for i, param := range fn.TypeParams {
		if param.Constraint != nil {
			if !tc.TypeSatisfiesConstraint(typeArgs[i], param.Constraint) {
				return Function{}, nil, fmt.Errorf("type argument %s does not satisfy constraint %s",
					tc.TypeToString(typeArgs[i]), tc.TypeToString(param.Constraint))
			}
		}
	}

	// Create new parameters with substituted types
	newParams := make([]Field, len(fn.Params))
	for i, param := range fn.Params {
		newParams[i] = Field{
			Name:           param.Name,
			TypeAnnotation: tc.SubstituteTypeParams(param.TypeAnnotation, typeArgMap),
			Required:       param.Required,
			Default:        param.Default,
		}
	}

	return Function{
		Name:       fn.Name,
		Params:     newParams,
		ReturnType: tc.SubstituteTypeParams(fn.ReturnType, typeArgMap),
		Body:       fn.Body,
	}, typeArgMap, nil
}

// TypeSatisfiesConstraint checks if a type satisfies a constraint (trait bound).
// Supports built-in constraints (Comparable, Numeric, Any, Hashable, Serializable)
// and user-defined traits by checking whether the type declares the trait in its impl list.
func (tc *TypeChecker) TypeSatisfiesConstraint(t Type, constraint Type) bool {
	if constraint == nil {
		return true // No constraint means any type is valid
	}

	// If constraint is a named type, check if it's a known trait
	if namedConstraint, ok := constraint.(NamedType); ok {
		switch namedConstraint.Name {
		case "Comparable":
			// Primitives are comparable
			switch t.(type) {
			case IntType, StringType, BoolType, FloatType:
				return true
			}
		case "Numeric":
			// Numeric types
			switch t.(type) {
			case IntType, FloatType:
				return true
			}
		case "Any":
			return true
		case "Hashable":
			// Primitives are hashable
			switch t.(type) {
			case IntType, StringType, BoolType:
				return true
			}
		}

		// Check user-defined trait: if the constraint is a known trait, check if
		// the type declares an impl for it via its TypeDef.Traits field
		if _, isTrait := tc.traitDefs[namedConstraint.Name]; isTrait {
			return tc.typeImplsTrait(t, namedConstraint.Name)
		}
	}

	// Default: types are compatible if they match
	return tc.TypesCompatible(t, constraint)
}

// typeImplsTrait checks whether a named type implements a given trait
// by looking up its TypeDef and checking the Traits field.
func (tc *TypeChecker) typeImplsTrait(t Type, traitName string) bool {
	var typeName string
	switch tt := t.(type) {
	case NamedType:
		typeName = tt.Name
	case GenericType:
		if named, ok := tt.BaseType.(NamedType); ok {
			typeName = named.Name
		}
	default:
		return false
	}

	td, ok := tc.typeDefs[typeName]
	if !ok {
		return false
	}

	for _, implTrait := range td.Traits {
		if implTrait == traitName {
			return true
		}
	}
	return false
}

// InferTypeArguments infers type arguments from function call arguments
// This is useful when type arguments are not explicitly provided
func (tc *TypeChecker) InferTypeArguments(fn Function, argValues []interface{}) ([]Type, error) {
	if len(fn.TypeParams) == 0 {
		return nil, nil // Not a generic function
	}

	typeArgs := make([]Type, len(fn.TypeParams))
	inferred := make(map[string]Type)

	// Try to infer each type parameter from the arguments
	for i, param := range fn.Params {
		if i >= len(argValues) {
			break
		}
		tc.inferFromValue(param.TypeAnnotation, argValues[i], inferred)
	}

	// Build the type arguments array
	for i, typeParam := range fn.TypeParams {
		if inferredType, ok := inferred[typeParam.Name]; ok {
			typeArgs[i] = inferredType
		} else {
			// Could not infer this type parameter
			return nil, fmt.Errorf("could not infer type for type parameter %s", typeParam.Name)
		}
	}

	return typeArgs, nil
}

// inferFromValue infers type parameter bindings from a value
func (tc *TypeChecker) inferFromValue(paramType Type, value interface{}, inferred map[string]Type) {
	if paramType == nil {
		return
	}

	switch pt := paramType.(type) {
	case TypeParameterType:
		// Direct type parameter, infer from value
		if _, exists := inferred[pt.Name]; !exists {
			inferred[pt.Name] = GetRuntimeType(value)
		}

	case ArrayType:
		// For arrays, infer element type
		if arr, ok := value.([]interface{}); ok && len(arr) > 0 {
			tc.inferFromValue(pt.ElementType, arr[0], inferred)
		}

	case FunctionType:
		// For function types, we can't easily infer from runtime values
		// This would require the value to be a callable with known signature
	}
}

// PushTypeScope pushes type bindings for a generic context
func (tc *TypeChecker) PushTypeScope(bindings map[string]Type) {
	for name, t := range bindings {
		tc.typeScope[name] = t
	}
}

// PopTypeScope removes type bindings from the scope
func (tc *TypeChecker) PopTypeScope(names []string) {
	for _, name := range names {
		delete(tc.typeScope, name)
	}
}

// GetTypeBinding returns the type bound to a type parameter name
func (tc *TypeChecker) GetTypeBinding(name string) (Type, bool) {
	t, ok := tc.typeScope[name]
	return t, ok
}
