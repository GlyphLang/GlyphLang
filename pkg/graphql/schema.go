package graphql

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/ast"
	"strings"
)

// Schema represents a GraphQL schema built from Glyph type definitions and resolvers.
type Schema struct {
	Types     map[string]*ObjectType
	Query     *ObjectType
	Mutation  *ObjectType
	Resolvers map[string]ast.GraphQLResolver // key: "operation.fieldName"
}

// ObjectType represents a GraphQL object type with fields.
type ObjectType struct {
	Name   string
	Fields map[string]*FieldDef
}

// FieldDef represents a GraphQL field definition.
type FieldDef struct {
	Name       string
	Type       string // GraphQL type string (e.g., "String", "[User]", "Int!")
	Args       []ArgDef
	IsNullable bool
	IsList     bool
	ElemType   string // Element type if IsList
}

// ArgDef represents a GraphQL field argument.
type ArgDef struct {
	Name     string
	Type     string
	Required bool
}

// BuildSchema constructs a GraphQL schema from interpreter type definitions and resolvers.
func BuildSchema(typeDefs map[string]ast.TypeDef, resolvers map[string]ast.GraphQLResolver) *Schema {
	schema := &Schema{
		Types:     make(map[string]*ObjectType),
		Resolvers: resolvers,
	}

	// Convert Glyph type definitions to GraphQL object types
	for name, td := range typeDefs {
		objType := &ObjectType{
			Name:   name,
			Fields: make(map[string]*FieldDef),
		}
		for _, field := range td.Fields {
			fd := convertField(field)
			objType.Fields[fd.Name] = fd
		}
		schema.Types[name] = objType
	}

	// Build Query type from query resolvers
	queryType := &ObjectType{
		Name:   "Query",
		Fields: make(map[string]*FieldDef),
	}
	mutationType := &ObjectType{
		Name:   "Mutation",
		Fields: make(map[string]*FieldDef),
	}

	for key, resolver := range resolvers {
		fd := &FieldDef{
			Name: resolver.FieldName,
			Type: typeToGraphQL(resolver.ReturnType),
		}
		for _, param := range resolver.Params {
			fd.Args = append(fd.Args, ArgDef{
				Name:     param.Name,
				Type:     typeToGraphQL(param.TypeAnnotation),
				Required: param.Required,
			})
		}

		if strings.HasPrefix(key, "query.") {
			queryType.Fields[resolver.FieldName] = fd
		} else if strings.HasPrefix(key, "mutation.") {
			mutationType.Fields[resolver.FieldName] = fd
		}
	}

	if len(queryType.Fields) > 0 {
		schema.Query = queryType
		schema.Types["Query"] = queryType
	}
	if len(mutationType.Fields) > 0 {
		schema.Mutation = mutationType
		schema.Types["Mutation"] = mutationType
	}

	return schema
}

// GenerateSDL generates a GraphQL SDL (Schema Definition Language) string.
func (s *Schema) GenerateSDL() string {
	var b strings.Builder

	// Write custom types (skip Query and Mutation)
	for name, objType := range s.Types {
		if name == "Query" || name == "Mutation" {
			continue
		}
		writeObjectType(&b, objType)
		b.WriteString("\n")
	}

	// Write Query type
	if s.Query != nil {
		writeObjectType(&b, s.Query)
		b.WriteString("\n")
	}

	// Write Mutation type
	if s.Mutation != nil {
		writeObjectType(&b, s.Mutation)
		b.WriteString("\n")
	}

	return strings.TrimSpace(b.String())
}

func writeObjectType(b *strings.Builder, objType *ObjectType) {
	fmt.Fprintf(b, "type %s {\n", objType.Name)
	for _, field := range objType.Fields {
		if len(field.Args) > 0 {
			argStrs := make([]string, len(field.Args))
			for i, arg := range field.Args {
				argType := arg.Type
				if arg.Required {
					argType += "!"
				}
				argStrs[i] = fmt.Sprintf("%s: %s", arg.Name, argType)
			}
			fmt.Fprintf(b, "  %s(%s): %s\n", field.Name, strings.Join(argStrs, ", "), field.Type)
		} else {
			fmt.Fprintf(b, "  %s: %s\n", field.Name, field.Type)
		}
	}
	b.WriteString("}\n")
}

func convertField(field ast.Field) *FieldDef {
	fd := &FieldDef{
		Name:       field.Name,
		Type:       typeToGraphQL(field.TypeAnnotation),
		IsNullable: !field.Required,
	}
	if _, ok := field.TypeAnnotation.(ast.ArrayType); ok {
		fd.IsList = true
	}
	return fd
}

func typeToGraphQL(t ast.Type) string {
	if t == nil {
		return "String"
	}
	switch v := t.(type) {
	case ast.IntType:
		return "Int"
	case ast.StringType:
		return "String"
	case ast.BoolType:
		return "Boolean"
	case ast.FloatType:
		return "Float"
	case ast.ArrayType:
		return "[" + typeToGraphQL(v.ElementType) + "]"
	case ast.OptionalType:
		return typeToGraphQL(v.InnerType)
	case ast.NamedType:
		return v.Name
	case ast.UnionType:
		// GraphQL unions are more complex; use the first type as primary
		if len(v.Types) > 0 {
			return typeToGraphQL(v.Types[0])
		}
		return "String"
	default:
		return "String"
	}
}
