package parser

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lexProviderSource(source string) []Token {
	lexer := NewLexer(source)
	tokens, _ := lexer.Tokenize()
	return tokens
}

func lexExpandedProviderSource(source string) []Token {
	lexer := NewExpandedLexer(source)
	tokens, _ := lexer.Tokenize()
	return tokens
}

// TestProviderDefinition tests parsing a basic provider definition
func TestProviderDefinition(t *testing.T) {
	source := `provider EmailService {
  send(to: str!, subject: str!, body: str!) -> bool
  status(messageId: str!) -> str
}`
	tokens := lexProviderSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	prov, ok := module.Items[0].(*ast.ProviderDef)
	require.True(t, ok, "expected ProviderDef")
	assert.Equal(t, "EmailService", prov.Name)
	require.Len(t, prov.Methods, 2)

	assert.Equal(t, "send", prov.Methods[0].Name)
	require.Len(t, prov.Methods[0].Params, 3)
	assert.Equal(t, "to", prov.Methods[0].Params[0].Name)
	assert.Equal(t, "subject", prov.Methods[0].Params[1].Name)
	assert.Equal(t, "body", prov.Methods[0].Params[2].Name)

	assert.Equal(t, "status", prov.Methods[1].Name)
	require.Len(t, prov.Methods[1].Params, 1)
	assert.Equal(t, "messageId", prov.Methods[1].Params[0].Name)
}

// TestProviderDefinitionExpandedSyntax tests provider parsing with expanded lexer
func TestProviderDefinitionExpandedSyntax(t *testing.T) {
	source := `provider PaymentGateway {
  charge(amount: int!, currency: str!) -> str
}`
	tokens := lexExpandedProviderSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	prov, ok := module.Items[0].(*ast.ProviderDef)
	require.True(t, ok)
	assert.Equal(t, "PaymentGateway", prov.Name)
	require.Len(t, prov.Methods, 1)
	assert.Equal(t, "charge", prov.Methods[0].Name)
}

// TestProviderWithGenericParams tests provider with generic type parameters
func TestProviderWithGenericParams(t *testing.T) {
	source := `provider Cache<T> {
  get(key: str!) -> T
  set(key: str!, value: T) -> bool
  delete(key: str!) -> bool
}`
	tokens := lexProviderSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	prov, ok := module.Items[0].(*ast.ProviderDef)
	require.True(t, ok)
	assert.Equal(t, "Cache", prov.Name)
	require.Len(t, prov.TypeParams, 1)
	assert.Equal(t, "T", prov.TypeParams[0].Name)
	require.Len(t, prov.Methods, 3)
}

// TestProviderMethodNoReturnType tests provider method with no return type
func TestProviderMethodNoReturnType(t *testing.T) {
	source := `provider Logger {
  log(msg: str!)
}`
	tokens := lexProviderSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	prov, ok := module.Items[0].(*ast.ProviderDef)
	require.True(t, ok)
	require.Len(t, prov.Methods, 1)
	assert.Equal(t, "log", prov.Methods[0].Name)
	assert.Nil(t, prov.Methods[0].ReturnType)
}

// TestProviderWithTypeDefAndRoute tests provider alongside types and routes
func TestProviderWithTypeDefAndRoute(t *testing.T) {
	source := `: ChargeResult {
  id: str
  status: str
}

provider Payments {
  charge(amount: int!, currency: str!) -> ChargeResult
}

@ GET /api/charge/:id -> ChargeResult {
  % payments: Payments
  $ result = payments.charge(100, "usd")
  > result
}`
	tokens := lexProviderSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 3)

	_, ok := module.Items[0].(*ast.TypeDef)
	require.True(t, ok, "expected TypeDef")

	prov, ok := module.Items[1].(*ast.ProviderDef)
	require.True(t, ok, "expected ProviderDef")
	assert.Equal(t, "Payments", prov.Name)

	route, ok := module.Items[2].(*ast.Route)
	require.True(t, ok, "expected Route")
	require.Len(t, route.Injections, 1)
	assert.Equal(t, "payments", route.Injections[0].Name)
}

// TestProviderEmptyBody tests provider with no methods
func TestProviderEmptyBody(t *testing.T) {
	source := `provider Empty {
}`
	tokens := lexProviderSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	prov, ok := module.Items[0].(*ast.ProviderDef)
	require.True(t, ok)
	assert.Equal(t, "Empty", prov.Name)
	assert.Empty(t, prov.Methods)
}

// TestProviderMethodMultipleParams tests provider method with multiple parameters
func TestProviderMethodMultipleParams(t *testing.T) {
	source := `provider ImageProcessor {
  resize(file: str!, width: int!, height: int!, quality: int) -> str
}`
	tokens := lexProviderSource(source)
	parser := NewParser(tokens)
	module, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, module.Items, 1)

	prov, ok := module.Items[0].(*ast.ProviderDef)
	require.True(t, ok)
	require.Len(t, prov.Methods, 1)
	assert.Equal(t, "resize", prov.Methods[0].Name)
	require.Len(t, prov.Methods[0].Params, 4)
	assert.Equal(t, "file", prov.Methods[0].Params[0].Name)
	assert.Equal(t, "width", prov.Methods[0].Params[1].Name)
	assert.Equal(t, "height", prov.Methods[0].Params[2].Name)
	assert.Equal(t, "quality", prov.Methods[0].Params[3].Name)
}
