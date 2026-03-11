package parser

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lexSource(t *testing.T, source string) []Token {
	t.Helper()
	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err, "Lexer should not return an error")
	return tokens
}

func lexExpandedSource(t *testing.T, source string) []Token {
	t.Helper()
	lexer := NewExpandedLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err, "ExpandedLexer should not return an error")
	return tokens
}

// TestParseGraphQLQueryResolver verifies parsing a query resolver
func TestParseGraphQLQueryResolver(t *testing.T) {
	source := `@ query user(id: int) -> User {
	> id
}`
	tokens := lexSource(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	resolver, ok := module.Items[0].(*ast.GraphQLResolver)
	require.True(t, ok, "item should be a GraphQLResolver")
	assert.Equal(t, ast.GraphQLQuery, resolver.Operation)
	assert.Equal(t, "user", resolver.FieldName)
	require.Len(t, resolver.Params, 1)
	assert.Equal(t, "id", resolver.Params[0].Name)
	assert.IsType(t, ast.IntType{}, resolver.Params[0].TypeAnnotation)

	returnType, ok := resolver.ReturnType.(ast.NamedType)
	require.True(t, ok, "return type should be NamedType")
	assert.Equal(t, "User", returnType.Name)
	assert.Len(t, resolver.Body, 1)
}

// TestParseGraphQLMutationResolver verifies parsing a mutation resolver
func TestParseGraphQLMutationResolver(t *testing.T) {
	source := `@ mutation createUser(name: string!, email: string) -> User {
	> name
}`
	tokens := lexSource(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	resolver, ok := module.Items[0].(*ast.GraphQLResolver)
	require.True(t, ok, "item should be a GraphQLResolver")
	assert.Equal(t, ast.GraphQLMutation, resolver.Operation)
	assert.Equal(t, "createUser", resolver.FieldName)
	require.Len(t, resolver.Params, 2)
	assert.Equal(t, "name", resolver.Params[0].Name)
	assert.True(t, resolver.Params[0].Required)
	assert.Equal(t, "email", resolver.Params[1].Name)
	assert.False(t, resolver.Params[1].Required)
}

// TestParseGraphQLSubscriptionResolver verifies parsing a subscription resolver
func TestParseGraphQLSubscriptionResolver(t *testing.T) {
	source := `@ subscription userCreated -> User {
	> "event"
}`
	tokens := lexSource(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	resolver, ok := module.Items[0].(*ast.GraphQLResolver)
	require.True(t, ok, "item should be a GraphQLResolver")
	assert.Equal(t, ast.GraphQLSubscription, resolver.Operation)
	assert.Equal(t, "userCreated", resolver.FieldName)
	assert.Empty(t, resolver.Params)
}

// TestParseGraphQLResolverWithInjection verifies dependency injection in resolvers
func TestParseGraphQLResolverWithInjection(t *testing.T) {
	source := `@ query users -> [User] {
	% db: Database
	> db
}`
	tokens := lexSource(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	resolver, ok := module.Items[0].(*ast.GraphQLResolver)
	require.True(t, ok, "item should be a GraphQLResolver")
	require.Len(t, resolver.Injections, 1)
	assert.Equal(t, "db", resolver.Injections[0].Name)
}

// TestParseGraphQLResolverWithAuth verifies auth middleware in resolvers
func TestParseGraphQLResolverWithAuth(t *testing.T) {
	source := `@ query me -> User {
	+ auth(jwt)
	> "user"
}`
	tokens := lexSource(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	resolver, ok := module.Items[0].(*ast.GraphQLResolver)
	require.True(t, ok, "item should be a GraphQLResolver")
	require.NotNil(t, resolver.Auth, "auth should not be nil")
	assert.Equal(t, "jwt", resolver.Auth.AuthType)
}

// TestParseGraphQLResolverNoParams verifies resolver without parameters
func TestParseGraphQLResolverNoParams(t *testing.T) {
	source := `@ query hello -> string {
	> "Hello, GraphQL!"
}`
	tokens := lexSource(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	resolver, ok := module.Items[0].(*ast.GraphQLResolver)
	require.True(t, ok, "item should be a GraphQLResolver")
	assert.Equal(t, "hello", resolver.FieldName)
	assert.Empty(t, resolver.Params)
}

// TestParseGraphQLResolverMissingBrace verifies error on missing body brace
func TestParseGraphQLResolverMissingBrace(t *testing.T) {
	source := `@ query hello -> string`
	tokens := lexSource(t, source)
	p := NewParser(tokens)
	_, err := p.Parse()
	assert.Error(t, err, "should error on missing brace")
}

// TestParseGraphQLResolverMissingFieldName verifies error on missing field name
func TestParseGraphQLResolverMissingFieldName(t *testing.T) {
	source := `@ query {`
	tokens := lexSource(t, source)
	p := NewParser(tokens)
	_, err := p.Parse()
	assert.Error(t, err, "should error on missing field name")
}

// TestParseMultipleGraphQLResolvers verifies multiple resolvers in one module
func TestParseMultipleGraphQLResolvers(t *testing.T) {
	source := `@ query user(id: int) -> User {
	> id
}

@ mutation createUser(name: string!) -> User {
	> name
}

@ subscription userCreated -> User {
	> "event"
}`
	tokens := lexSource(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 3)

	r1, ok := module.Items[0].(*ast.GraphQLResolver)
	require.True(t, ok)
	assert.Equal(t, ast.GraphQLQuery, r1.Operation)

	r2, ok := module.Items[1].(*ast.GraphQLResolver)
	require.True(t, ok)
	assert.Equal(t, ast.GraphQLMutation, r2.Operation)

	r3, ok := module.Items[2].(*ast.GraphQLResolver)
	require.True(t, ok)
	assert.Equal(t, ast.GraphQLSubscription, r3.Operation)
}

// TestParseExpandedGraphQLQuery verifies .glyphx expanded syntax parsing
func TestParseExpandedGraphQLQuery(t *testing.T) {
	source := `route query user(id: int) -> User {
	return id
}`
	tokens := lexExpandedSource(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	resolver, ok := module.Items[0].(*ast.GraphQLResolver)
	require.True(t, ok, "item should be a GraphQLResolver")
	assert.Equal(t, ast.GraphQLQuery, resolver.Operation)
	assert.Equal(t, "user", resolver.FieldName)
}
