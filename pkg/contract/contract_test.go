// Tests for pkg/contract/verify.go which implements Verify, Diff, and VerifyConsumer functions.
// Types used: Violation, VerifyResult, BreakingChange, DiffResult, ConsumerExpectation, ConsumerResult (all in verify.go).
// ast.Route implements ast.Item (ast.go:36).
// interpreter.GetContract and GetContracts (ast.go:568-575).
package contract

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeContract(name string, endpoints ...ast.ContractEndpoint) ast.ContractDef {
	return ast.ContractDef{Name: name, Endpoints: endpoints}
}

func makeEndpoint(method ast.HttpMethod, path string, returnType ast.Type) ast.ContractEndpoint {
	return ast.ContractEndpoint{Method: method, Path: path, ReturnType: returnType}
}

func makeRoute(method ast.HttpMethod, path string, returnType ast.Type) *ast.Route {
	return &ast.Route{Method: method, Path: path, ReturnType: returnType}
}

// TestVerifyPass tests that verification passes when routes match contract
func TestVerifyPass(t *testing.T) {
	ct := makeContract("UserService",
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
		makeEndpoint(ast.Post, "/users", ast.NamedType{Name: "User"}),
	)

	items := []ast.Item{
		makeRoute(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
		makeRoute(ast.Post, "/users", ast.NamedType{Name: "User"}),
	}

	result := Verify(ct, items)
	assert.True(t, result.Passed)
	assert.Empty(t, result.Violations)
	assert.Equal(t, "UserService", result.ContractName)
}

// TestVerifyMissingEndpoint tests that verification fails when an endpoint is missing
func TestVerifyMissingEndpoint(t *testing.T) {
	ct := makeContract("UserService",
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
		makeEndpoint(ast.Delete, "/users/:id", ast.NamedType{Name: "Ok"}),
	)

	items := []ast.Item{
		makeRoute(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
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
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
	)

	items := []ast.Item{
		makeRoute(ast.Get, "/users/:id", ast.NamedType{Name: "Admin"}),
	}

	result := Verify(ct, items)
	assert.False(t, result.Passed)
	require.Len(t, result.Violations, 1)
	assert.Contains(t, result.Violations[0].Message, "return type mismatch")
}

// TestVerifyUnionTypeCompatibility tests union type matching
func TestVerifyUnionTypeCompatibility(t *testing.T) {
	ct := makeContract("UserService",
		makeEndpoint(ast.Get, "/users/:id",
			ast.UnionType{Types: []ast.Type{
				ast.NamedType{Name: "User"},
				ast.NamedType{Name: "NotFound"},
			}}),
	)

	items := []ast.Item{
		makeRoute(ast.Get, "/users/:id",
			ast.UnionType{Types: []ast.Type{
				ast.NamedType{Name: "User"},
				ast.NamedType{Name: "NotFound"},
			}}),
	}

	result := Verify(ct, items)
	assert.True(t, result.Passed)
}

// TestDiffNoBreakingChanges tests that identical contracts have no breaking changes
func TestDiffNoBreakingChanges(t *testing.T) {
	old := makeContract("UserService",
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
	)
	newC := makeContract("UserService",
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
	)

	result := Diff(old, newC)
	assert.False(t, result.HasBreaking)
	assert.Empty(t, result.BreakingChanges)
}

// TestDiffRemovedEndpoint tests that removing an endpoint is a breaking change
func TestDiffRemovedEndpoint(t *testing.T) {
	old := makeContract("UserService",
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
		makeEndpoint(ast.Delete, "/users/:id", ast.NamedType{Name: "Ok"}),
	)
	newC := makeContract("UserService",
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
	)

	result := Diff(old, newC)
	assert.True(t, result.HasBreaking)
	require.Len(t, result.BreakingChanges, 1)
	assert.Equal(t, "removed", result.BreakingChanges[0].Type)
}

// TestDiffReturnTypeChanged tests that changing return type is a breaking change
func TestDiffReturnTypeChanged(t *testing.T) {
	old := makeContract("UserService",
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
	)
	newC := makeContract("UserService",
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "UserV2"}),
	)

	result := Diff(old, newC)
	assert.True(t, result.HasBreaking)
	require.Len(t, result.BreakingChanges, 1)
	assert.Equal(t, "return_type_changed", result.BreakingChanges[0].Type)
}

// TestDiffAddedEndpoint tests that adding a new endpoint is non-breaking
func TestDiffAddedEndpoint(t *testing.T) {
	old := makeContract("UserService",
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
	)
	newC := makeContract("UserService",
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
		makeEndpoint(ast.Post, "/users", ast.NamedType{Name: "User"}),
	)

	result := Diff(old, newC)
	assert.False(t, result.HasBreaking)
	require.Len(t, result.Additions, 1)
	assert.Equal(t, "added", result.Additions[0].Type)
}

// TestVerifyConsumerPass tests consumer expectations pass
func TestVerifyConsumerPass(t *testing.T) {
	ct := makeContract("UserService",
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
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
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "User"}),
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
		makeEndpoint(ast.Get, "/users/:id", ast.NamedType{Name: "Admin"}),
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
	assert.Equal(t, "int", typeString(ast.IntType{}))
	assert.Equal(t, "string", typeString(ast.StringType{}))
	assert.Equal(t, "bool", typeString(ast.BoolType{}))
	assert.Equal(t, "float", typeString(ast.FloatType{}))
	assert.Equal(t, "User", typeString(ast.NamedType{Name: "User"}))
	assert.Equal(t, "[int]", typeString(ast.ArrayType{ElementType: ast.IntType{}}))
	assert.Equal(t, "string?", typeString(ast.OptionalType{InnerType: ast.StringType{}}))
	assert.Equal(t, "User | NotFound", typeString(ast.UnionType{
		Types: []ast.Type{ast.NamedType{Name: "User"}, ast.NamedType{Name: "NotFound"}},
	}))
}

// TestInterpreterContractStorage tests that the interpreter stores contracts
func TestInterpreterContractStorage(t *testing.T) {
	interp := interpreter.NewInterpreter()

	module := ast.Module{
		Items: []ast.Item{
			&ast.ContractDef{
				Name: "UserService",
				Endpoints: []ast.ContractEndpoint{
					{Method: ast.Get, Path: "/users/:id", ReturnType: ast.NamedType{Name: "User"}},
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
