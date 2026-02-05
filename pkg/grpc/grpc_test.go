package grpc

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateProtoBasic verifies proto generation from a simple service
func TestGenerateProtoBasic(t *testing.T) {
	services := map[string]ast.GRPCService{
		"UserService": {
			Name: "UserService",
			Methods: []ast.GRPCMethod{
				{
					Name:       "GetUser",
					InputType:  ast.NamedType{Name: "GetUserRequest"},
					ReturnType: ast.NamedType{Name: "User"},
					StreamType: ast.GRPCUnary,
				},
			},
		},
	}

	typeDefs := map[string]ast.TypeDef{
		"User": {
			Name: "User",
			Fields: []ast.Field{
				{Name: "id", TypeAnnotation: ast.IntType{}},
				{Name: "name", TypeAnnotation: ast.StringType{}},
				{Name: "email", TypeAnnotation: ast.StringType{}},
			},
		},
		"GetUserRequest": {
			Name: "GetUserRequest",
			Fields: []ast.Field{
				{Name: "id", TypeAnnotation: ast.IntType{}},
			},
		},
	}

	proto := GenerateProto("myapp", services, typeDefs)
	require.NotNil(t, proto)
	assert.Equal(t, "myapp", proto.Package)
	require.Len(t, proto.Services, 1)
	require.Len(t, proto.Messages, 2)

	output := proto.Generate()
	assert.Contains(t, output, `syntax = "proto3"`)
	assert.Contains(t, output, "package myapp")
	assert.Contains(t, output, "message User")
	assert.Contains(t, output, "message GetUserRequest")
	assert.Contains(t, output, "service UserService")
	assert.Contains(t, output, "rpc GetUser (GetUserRequest) returns (User)")
}

// TestGenerateProtoWithStreaming verifies proto generation with streaming methods
func TestGenerateProtoWithStreaming(t *testing.T) {
	services := map[string]ast.GRPCService{
		"ChatService": {
			Name: "ChatService",
			Methods: []ast.GRPCMethod{
				{
					Name:       "ListMessages",
					InputType:  ast.NamedType{Name: "ListRequest"},
					ReturnType: ast.NamedType{Name: "Message"},
					StreamType: ast.GRPCServerStream,
				},
				{
					Name:       "SendMessages",
					InputType:  ast.NamedType{Name: "Message"},
					ReturnType: ast.NamedType{Name: "SendResult"},
					StreamType: ast.GRPCClientStream,
				},
				{
					Name:       "Chat",
					InputType:  ast.NamedType{Name: "Message"},
					ReturnType: ast.NamedType{Name: "Message"},
					StreamType: ast.GRPCBidirectional,
				},
			},
		},
	}

	proto := GenerateProto("chat", services, nil)
	require.NotNil(t, proto)

	output := proto.Generate()
	assert.Contains(t, output, "rpc ListMessages (ListRequest) returns (stream Message)")
	assert.Contains(t, output, "rpc SendMessages (stream Message) returns (SendResult)")
	assert.Contains(t, output, "rpc Chat (stream Message) returns (stream Message)")
}

// TestGenerateProtoFieldTypes verifies type mapping from Glyph to proto types
func TestGenerateProtoFieldTypes(t *testing.T) {
	typeDefs := map[string]ast.TypeDef{
		"Item": {
			Name: "Item",
			Fields: []ast.Field{
				{Name: "id", TypeAnnotation: ast.IntType{}},
				{Name: "name", TypeAnnotation: ast.StringType{}},
				{Name: "active", TypeAnnotation: ast.BoolType{}},
				{Name: "price", TypeAnnotation: ast.FloatType{}},
				{Name: "tags", TypeAnnotation: ast.ArrayType{ElementType: ast.StringType{}}},
			},
		},
	}

	proto := GenerateProto("", nil, typeDefs)
	require.NotNil(t, proto)
	require.Len(t, proto.Messages, 1)

	msg := proto.Messages[0]
	assert.Equal(t, "Item", msg.Name)
	require.Len(t, msg.Fields, 5)

	assert.Equal(t, "int64", msg.Fields[0].Type)
	assert.Equal(t, "string", msg.Fields[1].Type)
	assert.Equal(t, "bool", msg.Fields[2].Type)
	assert.Equal(t, "double", msg.Fields[3].Type)
	assert.Equal(t, "string", msg.Fields[4].Type)
	assert.True(t, msg.Fields[4].Repeated)
}

// TestGenerateProtoFieldNumbers verifies proto field numbering
func TestGenerateProtoFieldNumbers(t *testing.T) {
	typeDefs := map[string]ast.TypeDef{
		"Person": {
			Name: "Person",
			Fields: []ast.Field{
				{Name: "first", TypeAnnotation: ast.StringType{}},
				{Name: "last", TypeAnnotation: ast.StringType{}},
				{Name: "age", TypeAnnotation: ast.IntType{}},
			},
		},
	}

	proto := GenerateProto("", nil, typeDefs)
	require.Len(t, proto.Messages, 1)

	fields := proto.Messages[0].Fields
	assert.Equal(t, 1, fields[0].Number)
	assert.Equal(t, 2, fields[1].Number)
	assert.Equal(t, 3, fields[2].Number)
}

// TestSnakeCase verifies camelCase to snake_case conversion
func TestSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"name", "name"},
		{"firstName", "first_name"},
		{"userID", "user_i_d"},
		{"HTTPStatus", "h_t_t_p_status"},
		{"", ""},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, toSnakeCase(tt.input))
	}
}

// TestGlyphTypeToProto verifies type conversion
func TestGlyphTypeToProto(t *testing.T) {
	tests := []struct {
		input    ast.Type
		expected string
	}{
		{ast.IntType{}, "int64"},
		{ast.StringType{}, "string"},
		{ast.BoolType{}, "bool"},
		{ast.FloatType{}, "double"},
		{ast.NamedType{Name: "User"}, "User"},
		{ast.OptionalType{InnerType: ast.IntType{}}, "int64"},
		{nil, "string"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, glyphTypeToProto(tt.input))
	}
}

// TestGenerateProtoEmptyPackage verifies proto without package name
func TestGenerateProtoEmptyPackage(t *testing.T) {
	proto := GenerateProto("", nil, nil)
	require.NotNil(t, proto)
	output := proto.Generate()
	assert.Contains(t, output, `syntax = "proto3"`)
	assert.NotContains(t, output, "package")
}

// TestGRPCStreamTypeString verifies stream type string representations
func TestGRPCStreamTypeString(t *testing.T) {
	assert.Equal(t, "unary", ast.GRPCUnary.String())
	assert.Equal(t, "server_stream", ast.GRPCServerStream.String())
	assert.Equal(t, "client_stream", ast.GRPCClientStream.String())
	assert.Equal(t, "bidirectional", ast.GRPCBidirectional.String())
}

// TestGenerateProtoMultipleServices verifies multiple services in one proto file
func TestGenerateProtoMultipleServices(t *testing.T) {
	services := map[string]ast.GRPCService{
		"UserService": {
			Name: "UserService",
			Methods: []ast.GRPCMethod{
				{Name: "GetUser", InputType: ast.NamedType{Name: "GetUserRequest"}, ReturnType: ast.NamedType{Name: "User"}},
			},
		},
		"OrderService": {
			Name: "OrderService",
			Methods: []ast.GRPCMethod{
				{Name: "GetOrder", InputType: ast.NamedType{Name: "GetOrderRequest"}, ReturnType: ast.NamedType{Name: "Order"}},
			},
		},
	}

	proto := GenerateProto("app", services, nil)
	require.Len(t, proto.Services, 2)
	output := proto.Generate()
	assert.Contains(t, output, "service UserService")
	assert.Contains(t, output, "service OrderService")
}
