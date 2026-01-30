package parser

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lexContractSource(source string) []Token {
	lexer := NewLexer(source)
	tokens, _ := lexer.Tokenize()
	return tokens
}

// TestContractDefinition tests parsing a basic contract definition
func TestContractDefinition(t *testing.T) {
	source := `contract UserService {
  @ GET /users -> User
  @ POST /users -> User
  @ DELETE /users/:id -> Ok
}`
	tokens := lexContractSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	contract, ok := module.Items[0].(*interpreter.ContractDef)
	require.True(t, ok, "expected ContractDef")
	assert.Equal(t, "UserService", contract.Name)
	require.Len(t, contract.Endpoints, 3)

	assert.Equal(t, interpreter.Get, contract.Endpoints[0].Method)
	assert.Equal(t, "/users", contract.Endpoints[0].Path)

	assert.Equal(t, interpreter.Post, contract.Endpoints[1].Method)
	assert.Equal(t, "/users", contract.Endpoints[1].Path)

	assert.Equal(t, interpreter.Delete, contract.Endpoints[2].Method)
	assert.Equal(t, "/users/:id", contract.Endpoints[2].Path)
}

// TestContractWithUnionReturnType tests parsing contract endpoint with union return type
func TestContractWithUnionReturnType(t *testing.T) {
	source := `contract UserService {
  @ GET /users/:id -> User | NotFound
}`
	tokens := lexContractSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	contract, ok := module.Items[0].(*interpreter.ContractDef)
	require.True(t, ok)
	require.Len(t, contract.Endpoints, 1)

	ep := contract.Endpoints[0]
	assert.Equal(t, interpreter.Get, ep.Method)
	assert.Equal(t, "/users/:id", ep.Path)

	unionType, ok := ep.ReturnType.(interpreter.UnionType)
	require.True(t, ok, "expected UnionType")
	require.Len(t, unionType.Types, 2)
}

// TestContractWithPathParams tests contract endpoints with path parameters
func TestContractWithPathParams(t *testing.T) {
	source := `contract OrderService {
  @ GET /orders/:orderId -> Order
  @ PUT /orders/:orderId -> Order
}`
	tokens := lexContractSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	contract, ok := module.Items[0].(*interpreter.ContractDef)
	require.True(t, ok)
	require.Len(t, contract.Endpoints, 2)

	assert.Equal(t, "/orders/:orderId", contract.Endpoints[0].Path)
	assert.Equal(t, interpreter.Get, contract.Endpoints[0].Method)
	assert.Equal(t, interpreter.Put, contract.Endpoints[1].Method)
}

// TestContractAllHttpMethods tests all HTTP methods in a contract
func TestContractAllHttpMethods(t *testing.T) {
	source := `contract Api {
  @ GET /a -> Ok
  @ POST /b -> Ok
  @ PUT /c -> Ok
  @ DELETE /d -> Ok
  @ PATCH /e -> Ok
}`
	tokens := lexContractSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)

	contract, ok := module.Items[0].(*interpreter.ContractDef)
	require.True(t, ok)
	require.Len(t, contract.Endpoints, 5)
	assert.Equal(t, interpreter.Get, contract.Endpoints[0].Method)
	assert.Equal(t, interpreter.Post, contract.Endpoints[1].Method)
	assert.Equal(t, interpreter.Put, contract.Endpoints[2].Method)
	assert.Equal(t, interpreter.Delete, contract.Endpoints[3].Method)
	assert.Equal(t, interpreter.Patch, contract.Endpoints[4].Method)
}

// TestContractSingleEndpoint tests contract with a single endpoint
func TestContractSingleEndpoint(t *testing.T) {
	source := `contract UserService {
  @ GET /users -> User
}`
	tokens := lexContractSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	contract, ok := module.Items[0].(*interpreter.ContractDef)
	require.True(t, ok)
	assert.Equal(t, "UserService", contract.Name)
	require.Len(t, contract.Endpoints, 1)
	assert.Equal(t, interpreter.Get, contract.Endpoints[0].Method)
}

// TestContractEmptyBody tests parsing an empty contract
func TestContractEmptyBody(t *testing.T) {
	source := `contract EmptyService {
}`
	tokens := lexContractSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	contract, ok := module.Items[0].(*interpreter.ContractDef)
	require.True(t, ok)
	assert.Equal(t, "EmptyService", contract.Name)
	assert.Empty(t, contract.Endpoints)
}

// TestContractWithTypeDefAndRoute tests contract alongside type defs and routes
func TestContractWithTypeDefAndRoute(t *testing.T) {
	source := `: User {
  id: int
  name: string
}

contract UserAPI {
  @ GET /users/:id -> User
}

@ GET /users/:id {
  > { id: 1, name: "test" }
}`
	tokens := lexContractSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 3)

	_, ok := module.Items[0].(*interpreter.TypeDef)
	require.True(t, ok)
	_, ok = module.Items[1].(*interpreter.ContractDef)
	require.True(t, ok)
	_, ok = module.Items[2].(*interpreter.Route)
	require.True(t, ok)
}
