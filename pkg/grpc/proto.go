package grpc

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"strings"
)

// ProtoFile represents a generated .proto file.
type ProtoFile struct {
	Package  string
	Services []ProtoService
	Messages []ProtoMessage
}

// ProtoService represents a gRPC service in proto format.
type ProtoService struct {
	Name    string
	Methods []ProtoMethod
}

// ProtoMethod represents a gRPC method in proto format.
type ProtoMethod struct {
	Name            string
	InputType       string
	OutputType      string
	ClientStreaming bool
	ServerStreaming bool
}

// ProtoMessage represents a protobuf message.
type ProtoMessage struct {
	Name   string
	Fields []ProtoField
}

// ProtoField represents a field in a protobuf message.
type ProtoField struct {
	Name     string
	Type     string
	Number   int
	Repeated bool
}

// GenerateProto generates a .proto file from Glyph service definitions and type definitions.
func GenerateProto(packageName string, services map[string]interpreter.GRPCService, typeDefs map[string]interpreter.TypeDef) *ProtoFile {
	proto := &ProtoFile{
		Package: packageName,
	}

	// Convert type definitions to proto messages
	for name, td := range typeDefs {
		msg := ProtoMessage{Name: name}
		for i, field := range td.Fields {
			pf := ProtoField{
				Name:   toSnakeCase(field.Name),
				Type:   glyphTypeToProto(field.TypeAnnotation),
				Number: i + 1,
			}
			if _, ok := field.TypeAnnotation.(interpreter.ArrayType); ok {
				pf.Repeated = true
				pf.Type = glyphTypeToProto(field.TypeAnnotation.(interpreter.ArrayType).ElementType)
			}
			msg.Fields = append(msg.Fields, pf)
		}
		proto.Messages = append(proto.Messages, msg)
	}

	// Convert service definitions
	for _, svc := range services {
		protoSvc := ProtoService{Name: svc.Name}
		for _, method := range svc.Methods {
			pm := ProtoMethod{
				Name:            method.Name,
				InputType:       typeToProtoName(method.InputType),
				OutputType:      typeToProtoName(method.ReturnType),
				ClientStreaming: method.StreamType == interpreter.GRPCClientStream || method.StreamType == interpreter.GRPCBidirectional,
				ServerStreaming: method.StreamType == interpreter.GRPCServerStream || method.StreamType == interpreter.GRPCBidirectional,
			}
			protoSvc.Methods = append(protoSvc.Methods, pm)
		}
		proto.Services = append(proto.Services, protoSvc)
	}

	return proto
}

// Generate returns the formatted .proto file content.
func (p *ProtoFile) Generate() string {
	var b strings.Builder

	b.WriteString("syntax = \"proto3\";\n\n")
	if p.Package != "" {
		fmt.Fprintf(&b, "package %s;\n\n", p.Package)
	}

	// Write messages
	for _, msg := range p.Messages {
		fmt.Fprintf(&b, "message %s {\n", msg.Name)
		for _, field := range msg.Fields {
			prefix := ""
			if field.Repeated {
				prefix = "repeated "
			}
			fmt.Fprintf(&b, "  %s%s %s = %d;\n", prefix, field.Type, field.Name, field.Number)
		}
		b.WriteString("}\n\n")
	}

	// Write services
	for _, svc := range p.Services {
		fmt.Fprintf(&b, "service %s {\n", svc.Name)
		for _, method := range svc.Methods {
			inputStr := method.InputType
			if method.ClientStreaming {
				inputStr = "stream " + inputStr
			}
			outputStr := method.OutputType
			if method.ServerStreaming {
				outputStr = "stream " + outputStr
			}
			fmt.Fprintf(&b, "  rpc %s (%s) returns (%s);\n", method.Name, inputStr, outputStr)
		}
		b.WriteString("}\n")
	}

	return strings.TrimSpace(b.String())
}

func glyphTypeToProto(t interpreter.Type) string {
	if t == nil {
		return "string"
	}
	switch v := t.(type) {
	case interpreter.IntType:
		return "int64"
	case interpreter.StringType:
		return "string"
	case interpreter.BoolType:
		return "bool"
	case interpreter.FloatType:
		return "double"
	case interpreter.ArrayType:
		return glyphTypeToProto(v.ElementType)
	case interpreter.OptionalType:
		return glyphTypeToProto(v.InnerType)
	case interpreter.NamedType:
		return v.Name
	default:
		return "string"
	}
}

func typeToProtoName(t interpreter.Type) string {
	if t == nil {
		return "Empty"
	}
	switch v := t.(type) {
	case interpreter.NamedType:
		return v.Name
	case interpreter.ArrayType:
		return typeToProtoName(v.ElementType)
	default:
		return glyphTypeToProto(t)
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteByte(byte(r - 'A' + 'a'))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
