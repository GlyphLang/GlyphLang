package parser

import (
	"testing"

	// GRPCService, GRPCHandler, and GRPCMethod types are defined in
	// pkg/interpreter/ast.go (lines 638-660). GRPCStreamType constants
	// are at lines 603-610. These were added as part of issue #38.
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lexGRPC(t *testing.T, source string) []Token {
	t.Helper()
	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	require.NoError(t, err, "Lexer should not return an error")
	return tokens
}

// TestParseGRPCServiceDefinition verifies parsing a gRPC service with methods
func TestParseGRPCServiceDefinition(t *testing.T) {
	source := `@ rpc UserService {
	GetUser(GetUserRequest) -> User
	ListUsers(ListUsersRequest) -> [User]
}`
	tokens := lexGRPC(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	svc, ok := module.Items[0].(*interpreter.GRPCService)
	require.True(t, ok, "item should be a GRPCService")
	assert.Equal(t, "UserService", svc.Name)
	require.Len(t, svc.Methods, 2)

	assert.Equal(t, "GetUser", svc.Methods[0].Name)
	inputType, ok := svc.Methods[0].InputType.(interpreter.NamedType)
	require.True(t, ok)
	assert.Equal(t, "GetUserRequest", inputType.Name)

	assert.Equal(t, "ListUsers", svc.Methods[1].Name)
}

// TestParseGRPCServerStreaming verifies parsing server streaming methods
func TestParseGRPCServerStreaming(t *testing.T) {
	source := `@ rpc EventService {
	Subscribe(SubscribeRequest) -> stream Event
}`
	tokens := lexGRPC(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	svc, ok := module.Items[0].(*interpreter.GRPCService)
	require.True(t, ok)
	require.Len(t, svc.Methods, 1)
	assert.Equal(t, interpreter.GRPCServerStream, svc.Methods[0].StreamType)
}

// TestParseGRPCClientStreaming verifies parsing client streaming methods
func TestParseGRPCClientStreaming(t *testing.T) {
	source := `@ rpc UploadService {
	Upload(stream FileChunk) -> UploadResult
}`
	tokens := lexGRPC(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	svc, ok := module.Items[0].(*interpreter.GRPCService)
	require.True(t, ok)
	require.Len(t, svc.Methods, 1)
	assert.Equal(t, interpreter.GRPCClientStream, svc.Methods[0].StreamType)
}

// TestParseGRPCBidirectionalStreaming verifies parsing bidirectional streaming
func TestParseGRPCBidirectionalStreaming(t *testing.T) {
	source := `@ rpc ChatService {
	Chat(stream Message) -> stream Message
}`
	tokens := lexGRPC(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	svc, ok := module.Items[0].(*interpreter.GRPCService)
	require.True(t, ok)
	require.Len(t, svc.Methods, 1)
	assert.Equal(t, interpreter.GRPCBidirectional, svc.Methods[0].StreamType)
}

// TestParseGRPCHandler verifies parsing a gRPC handler implementation
func TestParseGRPCHandler(t *testing.T) {
	source := `@ rpc GetUser(req: GetUserRequest) -> User {
	> req
}`
	tokens := lexGRPC(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	handler, ok := module.Items[0].(*interpreter.GRPCHandler)
	require.True(t, ok, "item should be a GRPCHandler")
	assert.Equal(t, "GetUser", handler.MethodName)
	require.Len(t, handler.Params, 1)
	assert.Equal(t, "req", handler.Params[0].Name)
	assert.Equal(t, interpreter.GRPCUnary, handler.StreamType)
}

// TestParseGRPCHandlerWithStreamReturn verifies handler with stream return type
func TestParseGRPCHandlerWithStreamReturn(t *testing.T) {
	source := `@ rpc ListUsers(req: ListRequest) -> stream User {
	> req
}`
	tokens := lexGRPC(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	handler, ok := module.Items[0].(*interpreter.GRPCHandler)
	require.True(t, ok)
	assert.Equal(t, interpreter.GRPCServerStream, handler.StreamType)
}

// TestParseGRPCHandlerWithInjection verifies dependency injection in handlers
func TestParseGRPCHandlerWithInjection(t *testing.T) {
	source := `@ rpc GetUser(req: GetUserRequest) -> User {
	% db: Database
	> req
}`
	tokens := lexGRPC(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	handler, ok := module.Items[0].(*interpreter.GRPCHandler)
	require.True(t, ok)
	require.Len(t, handler.Injections, 1)
	assert.Equal(t, "db", handler.Injections[0].Name)
}

// TestParseGRPCHandlerWithAuth verifies auth middleware in handlers
func TestParseGRPCHandlerWithAuth(t *testing.T) {
	source := `@ rpc GetUser(req: GetUserRequest) -> User {
	+ auth(jwt)
	> req
}`
	tokens := lexGRPC(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 1)

	handler, ok := module.Items[0].(*interpreter.GRPCHandler)
	require.True(t, ok)
	require.NotNil(t, handler.Auth)
	assert.Equal(t, "jwt", handler.Auth.AuthType)
}

// TestParseGRPCServiceAndHandler verifies both service definition and handler
func TestParseGRPCServiceAndHandler(t *testing.T) {
	source := `@ rpc UserService {
	GetUser(GetUserRequest) -> User
}

@ rpc GetUser(req: GetUserRequest) -> User {
	> req
}`
	tokens := lexGRPC(t, source)
	p := NewParser(tokens)
	module, err := p.Parse()
	require.NoError(t, err, "Parse should not return an error")
	require.Len(t, module.Items, 2)

	_, ok := module.Items[0].(*interpreter.GRPCService)
	require.True(t, ok, "first item should be a GRPCService")

	_, ok = module.Items[1].(*interpreter.GRPCHandler)
	require.True(t, ok, "second item should be a GRPCHandler")
}

// TestParseGRPCMissingBrace verifies error on missing service body
func TestParseGRPCMissingBrace(t *testing.T) {
	source := `@ rpc UserService`
	tokens := lexGRPC(t, source)
	p := NewParser(tokens)
	_, err := p.Parse()
	assert.Error(t, err, "should error on missing brace")
}

// TestParseGRPCHandlerMissingBrace verifies error on missing handler body
func TestParseGRPCHandlerMissingBrace(t *testing.T) {
	source := `@ rpc GetUser(req: Request) -> User`
	tokens := lexGRPC(t, source)
	p := NewParser(tokens)
	_, err := p.Parse()
	assert.Error(t, err, "should error on missing handler body brace")
}
