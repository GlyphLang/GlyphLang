package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTraitDefIsItem verifies TraitDef implements the Item interface
func TestTraitDefIsItem(t *testing.T) {
	trait := &TraitDef{
		Name: "Serializable",
		Methods: []TraitMethodSignature{
			{Name: "toJson", ReturnType: NamedType{Name: "string"}},
		},
	}
	var _ Item = trait // compile-time check
	assert.Equal(t, "Serializable", trait.Name)
}

// TestTraitDefIsNode verifies TraitDef implements the Node interface
func TestTraitDefIsNode(t *testing.T) {
	trait := &TraitDef{Name: "TestTrait"}
	// compile-time check - isNode() is unexported and in pkg/ast, verified via interface
	var _ Node = trait
}

// TestTypeDefTraitsField verifies the TypeDef Traits field
func TestTypeDefTraitsField(t *testing.T) {
	td := TypeDef{
		Name:   "User",
		Fields: []Field{{Name: "id", TypeAnnotation: IntType{}}},
		Traits: []string{"Serializable", "Hashable"},
		Methods: []MethodDef{
			{Name: "toJson", ReturnType: NamedType{Name: "string"}, Body: []Statement{}},
			{Name: "hash", ReturnType: IntType{}, Body: []Statement{}},
		},
	}
	assert.Len(t, td.Traits, 2)
	assert.Len(t, td.Methods, 2)
}

// TestInterpreterTraitStorage tests that the interpreter stores and retrieves traits
func TestInterpreterTraitStorage(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TraitDef{
				Name: "Serializable",
				Methods: []TraitMethodSignature{
					{Name: "toJson", ReturnType: NamedType{Name: "string"}},
					{Name: "fromJson", Params: []Field{{Name: "s", TypeAnnotation: StringType{}}}, ReturnType: NamedType{Name: "Self"}},
				},
			},
		},
	}

	err := interp.LoadModuleWithPath(module, "")
	require.NoError(t, err)

	trait, ok := interp.GetTraitDef("Serializable")
	assert.True(t, ok)
	assert.Equal(t, "Serializable", trait.Name)
	assert.Len(t, trait.Methods, 2)

	_, ok = interp.GetTraitDef("NonExistent")
	assert.False(t, ok)
}

// TestInterpreterGetTraitDefs tests GetTraitDefs returns all traits
func TestInterpreterGetTraitDefs(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TraitDef{Name: "Serializable", Methods: []TraitMethodSignature{{Name: "toJson"}}},
			&TraitDef{Name: "Hashable", Methods: []TraitMethodSignature{{Name: "hash"}}},
		},
	}

	err := interp.LoadModuleWithPath(module, "")
	require.NoError(t, err)

	traits := interp.GetTraitDefs()
	assert.Len(t, traits, 2)
	assert.Contains(t, traits, "Serializable")
	assert.Contains(t, traits, "Hashable")
}

// TestValidateTraitImplSuccess tests that ValidateTraitImpl passes for correct implementations
func TestValidateTraitImplSuccess(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TraitDef{
				Name: "Serializable",
				Methods: []TraitMethodSignature{
					{Name: "toJson", ReturnType: NamedType{Name: "string"}},
				},
			},
			&TypeDef{
				Name:   "User",
				Fields: []Field{{Name: "id", TypeAnnotation: IntType{}}},
				Traits: []string{"Serializable"},
				Methods: []MethodDef{
					{Name: "toJson", ReturnType: NamedType{Name: "string"}, Body: []Statement{}},
				},
			},
		},
	}

	err := interp.LoadModuleWithPath(module, "")
	require.NoError(t, err)

	td := interp.typeDefs["User"]
	err = interp.ValidateTraitImpl(td)
	assert.NoError(t, err)
}

// TestValidateTraitImplMissingMethod tests that ValidateTraitImpl fails when a method is missing
func TestValidateTraitImplMissingMethod(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TraitDef{
				Name: "Serializable",
				Methods: []TraitMethodSignature{
					{Name: "toJson", ReturnType: NamedType{Name: "string"}},
					{Name: "fromJson", ReturnType: NamedType{Name: "Self"}},
				},
			},
			&TypeDef{
				Name:   "User",
				Fields: []Field{{Name: "id", TypeAnnotation: IntType{}}},
				Traits: []string{"Serializable"},
				Methods: []MethodDef{
					{Name: "toJson", ReturnType: NamedType{Name: "string"}, Body: []Statement{}},
					// fromJson is missing
				},
			},
		},
	}

	err := interp.LoadModuleWithPath(module, "")
	require.NoError(t, err)

	td := interp.typeDefs["User"]
	err = interp.ValidateTraitImpl(td)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fromJson")
	assert.Contains(t, err.Error(), "Serializable")
}

// TestValidateTraitImplUnknownTrait tests that declaring a nonexistent trait is an error
func TestValidateTraitImplUnknownTrait(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&TypeDef{
				Name:   "User",
				Traits: []string{"NonExistent"},
			},
		},
	}

	err := interp.LoadModuleWithPath(module, "")
	require.NoError(t, err)

	td := interp.typeDefs["User"]
	err = interp.ValidateTraitImpl(td)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown trait")
}

// TestTypeSatisfiesUserDefinedTrait tests constraint checking with user-defined traits
func TestTypeSatisfiesUserDefinedTrait(t *testing.T) {
	tc := NewTypeChecker()

	// Register a trait
	tc.SetTraitDefs(map[string]TraitDef{
		"Serializable": {
			Name: "Serializable",
			Methods: []TraitMethodSignature{
				{Name: "toJson", ReturnType: NamedType{Name: "string"}},
			},
		},
	})

	// Register a type that implements the trait
	tc.SetTypeDefs(map[string]TypeDef{
		"User": {
			Name:   "User",
			Traits: []string{"Serializable"},
		},
		"Config": {
			Name: "Config",
			// Does NOT implement Serializable
		},
	})

	constraint := NamedType{Name: "Serializable"}

	// User implements Serializable
	assert.True(t, tc.TypeSatisfiesConstraint(NamedType{Name: "User"}, constraint))

	// Config does NOT implement Serializable
	assert.False(t, tc.TypeSatisfiesConstraint(NamedType{Name: "Config"}, constraint))

	// Primitive types do not implement user-defined traits
	assert.False(t, tc.TypeSatisfiesConstraint(IntType{}, constraint))
}

// TestBuiltinHashableConstraint tests the built-in Hashable constraint
func TestBuiltinHashableConstraint(t *testing.T) {
	tc := NewTypeChecker()
	constraint := NamedType{Name: "Hashable"}

	assert.True(t, tc.TypeSatisfiesConstraint(IntType{}, constraint))
	assert.True(t, tc.TypeSatisfiesConstraint(StringType{}, constraint))
	assert.True(t, tc.TypeSatisfiesConstraint(BoolType{}, constraint))
	assert.False(t, tc.TypeSatisfiesConstraint(FloatType{}, constraint))
}

// TestTypeImplsMultipleTraits tests that a type can implement multiple traits
func TestTypeImplsMultipleTraits(t *testing.T) {
	tc := NewTypeChecker()

	tc.SetTraitDefs(map[string]TraitDef{
		"Serializable": {Name: "Serializable", Methods: []TraitMethodSignature{{Name: "toJson"}}},
		"Hashable":     {Name: "Hashable", Methods: []TraitMethodSignature{{Name: "hash"}}},
	})

	tc.SetTypeDefs(map[string]TypeDef{
		"User": {
			Name:   "User",
			Traits: []string{"Serializable", "Hashable"},
		},
	})

	assert.True(t, tc.TypeSatisfiesConstraint(NamedType{Name: "User"}, NamedType{Name: "Serializable"}))
	assert.True(t, tc.TypeSatisfiesConstraint(NamedType{Name: "User"}, NamedType{Name: "Hashable"}))
}

// TestTraitMethodSignatureStruct tests the TraitMethodSignature struct fields
func TestTraitMethodSignatureStruct(t *testing.T) {
	sig := TraitMethodSignature{
		Name:       "serialize",
		Params:     []Field{{Name: "format", TypeAnnotation: StringType{}}},
		ReturnType: StringType{},
	}
	assert.Equal(t, "serialize", sig.Name)
	assert.Len(t, sig.Params, 1)
	assert.Equal(t, "format", sig.Params[0].Name)
}

// TestMethodDefStruct tests the MethodDef struct fields
func TestMethodDefStruct(t *testing.T) {
	method := MethodDef{
		Name:       "toJson",
		Params:     []Field{},
		ReturnType: NamedType{Name: "string"},
		Body:       []Statement{&ReturnStatement{Value: LiteralExpr{Value: StringLiteral{Value: "json"}}}},
	}
	assert.Equal(t, "toJson", method.Name)
	assert.Len(t, method.Body, 1)
}
