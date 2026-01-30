package grpc

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateProtoBasic verifies proto generation from a simple service
func TestGenerateProtoBasic(t *testing.T) {
	services := map[string]interpreter.GRPCService{
		"UserService": {
			Name: "UserService",
			Methods: []interpreter.GRPCMethod{
				{
					Name:       "GetUser",
					InputType:  interpreter.NamedType{Name: "GetUserRequest"},
					ReturnType: interpreter.NamedType{Name: "User"},
					StreamType: interpreter.GRPCUnary,
				},
			},
		},
	}

	typeDefs := map[string]interpreter.TypeDef{
		"User": {
			Name: "User",
			Fields: []interpreter.Field{
				{Name: "id", TypeAnnotation: interpreter.IntType{}},
				{Name: "name", TypeAnnotation: interpreter.StringType{}},
				{Name: "email", TypeAnnotation: interpreter.StringType{}},
			},
		},
		"GetUserRequest": {
			Name: "GetUserRequest",
			Fields: []interpreter.Field{
				{Name: "id", TypeAnnotation: interpreter.IntType{}},
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
	services := map[string]interpreter.GRPCService{
		"ChatService": {
			Name: "ChatService",
			Methods: []interpreter.GRPCMethod{
				{
					Name:       "ListMessages",
					InputType:  interpreter.NamedType{Name: "ListRequest"},
					ReturnType: interpreter.NamedType{Name: "Message"},
					StreamType: interpreter.GRPCServerStream,
				},
				{
					Name:       "SendMessages",
					InputType:  interpreter.NamedType{Name: "Message"},
					ReturnType: interpreter.NamedType{Name: "SendResult"},
					StreamType: interpreter.GRPCClientStream,
				},
				{
					Name:       "Chat",
					InputType:  interpreter.NamedType{Name: "Message"},
					ReturnType: interpreter.NamedType{Name: "Message"},
					StreamType: interpreter.GRPCBidirectional,
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
	typeDefs := map[string]interpreter.TypeDef{
		"Item": {
			Name: "Item",
			Fields: []interpreter.Field{
				{Name: "id", TypeAnnotation: interpreter.IntType{}},
				{Name: "name", TypeAnnotation: interpreter.StringType{}},
				{Name: "active", TypeAnnotation: interpreter.BoolType{}},
				{Name: "price", TypeAnnotation: interpreter.FloatType{}},
				{Name: "tags", TypeAnnotation: interpreter.ArrayType{ElementType: interpreter.StringType{}}},
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
	typeDefs := map[string]interpreter.TypeDef{
		"Person": {
			Name: "Person",
			Fields: []interpreter.Field{
				{Name: "first", TypeAnnotation: interpreter.StringType{}},
				{Name: "last", TypeAnnotation: interpreter.StringType{}},
				{Name: "age", TypeAnnotation: interpreter.IntType{}},
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
		input    interpreter.Type
		expected string
	}{
		{interpreter.IntType{}, "int64"},
		{interpreter.StringType{}, "string"},
		{interpreter.BoolType{}, "bool"},
		{interpreter.FloatType{}, "double"},
		{interpreter.NamedType{Name: "User"}, "User"},
		{interpreter.OptionalType{InnerType: interpreter.IntType{}}, "int64"},
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
	assert.Equal(t, "unary", interpreter.GRPCUnary.String())
	assert.Equal(t, "server_stream", interpreter.GRPCServerStream.String())
	assert.Equal(t, "client_stream", interpreter.GRPCClientStream.String())
	assert.Equal(t, "bidirectional", interpreter.GRPCBidirectional.String())
}

// TestGenerateProtoMultipleServices verifies multiple services in one proto file
func TestGenerateProtoMultipleServices(t *testing.T) {
	services := map[string]interpreter.GRPCService{
		"UserService": {
			Name: "UserService",
			Methods: []interpreter.GRPCMethod{
				{Name: "GetUser", InputType: interpreter.NamedType{Name: "GetUserRequest"}, ReturnType: interpreter.NamedType{Name: "User"}},
			},
		},
		"OrderService": {
			Name: "OrderService",
			Methods: []interpreter.GRPCMethod{
				{Name: "GetOrder", InputType: interpreter.NamedType{Name: "GetOrderRequest"}, ReturnType: interpreter.NamedType{Name: "Order"}},
			},
		},
	}

	proto := GenerateProto("app", services, nil)
	require.Len(t, proto.Services, 2)
	output := proto.Generate()
	assert.Contains(t, output, "service UserService")
	assert.Contains(t, output, "service OrderService")
}
