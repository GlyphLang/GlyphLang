// Tests for pkg/contract/verify.go which implements Verify, Diff, and VerifyConsumer functions.
// Types used: Violation, VerifyResult, BreakingChange, DiffResult, ConsumerExpectation, ConsumerResult (all in verify.go).
// interpreter.Route implements interpreter.Item (ast.go:36).
// interpreter.GetContract and GetContracts (interpreter.go:568-575).
package contract

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeContract(name string, endpoints ...interpreter.ContractEndpoint) interpreter.ContractDef {
	return interpreter.ContractDef{Name: name, Endpoints: endpoints}
}

func makeEndpoint(method interpreter.HttpMethod, path string, returnType interpreter.Type) interpreter.ContractEndpoint {
	return interpreter.ContractEndpoint{Method: method, Path: path, ReturnType: returnType}
}

func makeRoute(method interpreter.HttpMethod, path string, returnType interpreter.Type) *interpreter.Route {
	return &interpreter.Route{Method: method, Path: path, ReturnType: returnType}
}

// TestVerifyPass tests that verification passes when routes match contract
func TestVerifyPass(t *testing.T) {
	ct := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
		makeEndpoint(interpreter.Post, "/users", interpreter.NamedType{Name: "User"}),
	)

	items := []interpreter.Item{
		makeRoute(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
		makeRoute(interpreter.Post, "/users", interpreter.NamedType{Name: "User"}),
	}

	result := Verify(ct, items)
	assert.True(t, result.Passed)
	assert.Empty(t, result.Violations)
	assert.Equal(t, "UserService", result.ContractName)
}

// TestVerifyMissingEndpoint tests that verification fails when an endpoint is missing
func TestVerifyMissingEndpoint(t *testing.T) {
	ct := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
		makeEndpoint(interpreter.Delete, "/users/:id", interpreter.NamedType{Name: "Ok"}),
	)

	items := []interpreter.Item{
		makeRoute(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
	}

	result := Verify(ct, items)
	assert.False(t, result.Passed)
	require.Len(t, result.Violations, 1)
	assert.Contains(t, result.Violations[0].Message, "not found")
	assert.Contains(t, result.Violations[0].Endpoint, "DELETE")
}

// TestVerifyReturnTypeMismatch tests that verification detects return type mismatches
func TestVerifyReturnTypeMismatch(t *testing.T) {
	ct := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
	)

	items := []interpreter.Item{
		makeRoute(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "Admin"}),
	}

	result := Verify(ct, items)
	assert.False(t, result.Passed)
	require.Len(t, result.Violations, 1)
	assert.Contains(t, result.Violations[0].Message, "return type mismatch")
}

// TestVerifyUnionTypeCompatibility tests union type matching
func TestVerifyUnionTypeCompatibility(t *testing.T) {
	ct := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id",
			interpreter.UnionType{Types: []interpreter.Type{
				interpreter.NamedType{Name: "User"},
				interpreter.NamedType{Name: "NotFound"},
			}}),
	)

	items := []interpreter.Item{
		makeRoute(interpreter.Get, "/users/:id",
			interpreter.UnionType{Types: []interpreter.Type{
				interpreter.NamedType{Name: "User"},
				interpreter.NamedType{Name: "NotFound"},
			}}),
	}

	result := Verify(ct, items)
	assert.True(t, result.Passed)
}

// TestDiffNoBreakingChanges tests that identical contracts have no breaking changes
func TestDiffNoBreakingChanges(t *testing.T) {
	old := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
	)
	newC := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
	)

	result := Diff(old, newC)
	assert.False(t, result.HasBreaking)
	assert.Empty(t, result.BreakingChanges)
}

// TestDiffRemovedEndpoint tests that removing an endpoint is a breaking change
func TestDiffRemovedEndpoint(t *testing.T) {
	old := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
		makeEndpoint(interpreter.Delete, "/users/:id", interpreter.NamedType{Name: "Ok"}),
	)
	newC := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
	)

	result := Diff(old, newC)
	assert.True(t, result.HasBreaking)
	require.Len(t, result.BreakingChanges, 1)
	assert.Equal(t, "removed", result.BreakingChanges[0].Type)
}

// TestDiffReturnTypeChanged tests that changing return type is a breaking change
func TestDiffReturnTypeChanged(t *testing.T) {
	old := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
	)
	newC := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "UserV2"}),
	)

	result := Diff(old, newC)
	assert.True(t, result.HasBreaking)
	require.Len(t, result.BreakingChanges, 1)
	assert.Equal(t, "return_type_changed", result.BreakingChanges[0].Type)
}

// TestDiffAddedEndpoint tests that adding a new endpoint is non-breaking
func TestDiffAddedEndpoint(t *testing.T) {
	old := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
	)
	newC := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
		makeEndpoint(interpreter.Post, "/users", interpreter.NamedType{Name: "User"}),
	)

	result := Diff(old, newC)
	assert.False(t, result.HasBreaking)
	require.Len(t, result.Additions, 1)
	assert.Equal(t, "added", result.Additions[0].Type)
}

// TestVerifyConsumerPass tests consumer expectations pass
func TestVerifyConsumerPass(t *testing.T) {
	ct := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
	)

	expectations := []ConsumerExpectation{
		{Method: "GET", Path: "/users/:id", ReturnType: "User"},
	}

	result := VerifyConsumer("OrderService", ct, expectations)
	assert.True(t, result.Passed)
	assert.Empty(t, result.Failures)
	assert.Equal(t, "OrderService", result.Consumer)
	assert.Equal(t, "UserService", result.Provider)
}

// TestVerifyConsumerMissingEndpoint tests consumer fails when expected endpoint is missing
func TestVerifyConsumerMissingEndpoint(t *testing.T) {
	ct := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "User"}),
	)

	expectations := []ConsumerExpectation{
		{Method: "DELETE", Path: "/users/:id"},
	}

	result := VerifyConsumer("OrderService", ct, expectations)
	assert.False(t, result.Passed)
	require.Len(t, result.Failures, 1)
	assert.Contains(t, result.Failures[0], "not found")
}

// TestVerifyConsumerReturnTypeMismatch tests consumer fails on return type mismatch
func TestVerifyConsumerReturnTypeMismatch(t *testing.T) {
	ct := makeContract("UserService",
		makeEndpoint(interpreter.Get, "/users/:id", interpreter.NamedType{Name: "Admin"}),
	)

	expectations := []ConsumerExpectation{
		{Method: "GET", Path: "/users/:id", ReturnType: "User"},
	}

	result := VerifyConsumer("OrderService", ct, expectations)
	assert.False(t, result.Passed)
	require.Len(t, result.Failures, 1)
	assert.Contains(t, result.Failures[0], "expected return type")
}

// TestViolationString tests the Violation.String() method
func TestViolationString(t *testing.T) {
	v := Violation{Endpoint: "GET /users", Message: "endpoint not found"}
	assert.Equal(t, "GET /users: endpoint not found", v.String())
}

// TestBreakingChangeString tests the BreakingChange.String() method
func TestBreakingChangeString(t *testing.T) {
	bc := BreakingChange{Type: "removed", Endpoint: "DELETE /users/:id", Detail: "endpoint was removed"}
	assert.Equal(t, "[removed] DELETE /users/:id: endpoint was removed", bc.String())
}

// TestTypeStringVariants tests typeString for different type kinds
func TestTypeStringVariants(t *testing.T) {
	assert.Equal(t, "int", typeString(interpreter.IntType{}))
	assert.Equal(t, "string", typeString(interpreter.StringType{}))
	assert.Equal(t, "bool", typeString(interpreter.BoolType{}))
	assert.Equal(t, "float", typeString(interpreter.FloatType{}))
	assert.Equal(t, "User", typeString(interpreter.NamedType{Name: "User"}))
	assert.Equal(t, "[int]", typeString(interpreter.ArrayType{ElementType: interpreter.IntType{}}))
	assert.Equal(t, "string?", typeString(interpreter.OptionalType{InnerType: interpreter.StringType{}}))
	assert.Equal(t, "User | NotFound", typeString(interpreter.UnionType{
		Types: []interpreter.Type{interpreter.NamedType{Name: "User"}, interpreter.NamedType{Name: "NotFound"}},
	}))
}

// TestInterpreterContractStorage tests that the interpreter stores contracts
func TestInterpreterContractStorage(t *testing.T) {
	interp := interpreter.NewInterpreter()

	module := interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.ContractDef{
				Name: "UserService",
				Endpoints: []interpreter.ContractEndpoint{
					{Method: interpreter.Get, Path: "/users/:id", ReturnType: interpreter.NamedType{Name: "User"}},
				},
			},
		},
	}

	err := interp.LoadModuleWithPath(module, "")
	require.NoError(t, err)

	c, ok := interp.GetContract("UserService")
	assert.True(t, ok)
	assert.Equal(t, "UserService", c.Name)
	assert.Len(t, c.Endpoints, 1)

	contracts := interp.GetContracts()
	assert.Len(t, contracts, 1)
}
