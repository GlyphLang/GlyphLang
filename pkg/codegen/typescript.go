package codegen

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/ast"
	"strings"
)

// TypeScriptGenerator generates TypeScript API client code from a GlyphLang Module.
type TypeScriptGenerator struct {
	baseURL string
}

// NewTypeScriptGenerator creates a new TypeScript code generator.
func NewTypeScriptGenerator(baseURL string) *TypeScriptGenerator {
	return &TypeScriptGenerator{baseURL: baseURL}
}

// Generate produces TypeScript source code from a parsed GlyphLang module.
func (g *TypeScriptGenerator) Generate(module *ast.Module) string {
	var sb strings.Builder

	sb.WriteString("// Auto-generated TypeScript client from GlyphLang\n")
	sb.WriteString("// Do not edit manually\n\n")

	// Collect type definitions
	var typeDefs []*ast.TypeDef
	for _, item := range module.Items {
		td := getTypeDef(item)
		if td != nil {
			typeDefs = append(typeDefs, td)
		}
	}

	// Generate TypeScript interfaces
	for _, td := range typeDefs {
		g.generateInterface(&sb, td)
		sb.WriteString("\n")
	}

	// Collect routes
	var routes []*ast.Route
	for _, item := range module.Items {
		route := getRoute(item)
		if route != nil && route.Method != ast.WebSocket {
			routes = append(routes, route)
		}
	}

	// Generate API client class
	g.generateClient(&sb, routes, typeDefs)

	return sb.String()
}

// generateInterface writes a TypeScript interface for a type definition.
func (g *TypeScriptGenerator) generateInterface(sb *strings.Builder, td *ast.TypeDef) {
	fmt.Fprintf(sb, "export interface %s {\n", td.Name)
	for _, field := range td.Fields {
		tsType := glyphTypeToTS(field.TypeAnnotation)
		optional := ""
		if !field.Required {
			optional = "?"
		}
		fmt.Fprintf(sb, "  %s%s: %s;\n", field.Name, optional, tsType)
	}
	sb.WriteString("}\n")
}

// generateClient writes a TypeScript API client class.
func (g *TypeScriptGenerator) generateClient(sb *strings.Builder, routes []*ast.Route, typeDefs []*ast.TypeDef) {
	sb.WriteString("export class ApiClient {\n")
	sb.WriteString("  private baseUrl: string;\n")
	sb.WriteString("  private headers: Record<string, string>;\n\n")
	fmt.Fprintf(sb, "  constructor(baseUrl: string = %q, headers: Record<string, string> = {}) {\n", g.baseURL)
	sb.WriteString("    this.baseUrl = baseUrl.replace(/\\/$/, '');\n")
	sb.WriteString("    this.headers = { 'Content-Type': 'application/json', ...headers };\n")
	sb.WriteString("  }\n\n")

	// Generate a method for each route
	for _, route := range routes {
		g.generateMethod(sb, route, typeDefs)
	}

	// Add helper method
	sb.WriteString("  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {\n")
	sb.WriteString("    const response = await fetch(`${this.baseUrl}${path}`, {\n")
	sb.WriteString("      method,\n")
	sb.WriteString("      headers: this.headers,\n")
	sb.WriteString("      body: body ? JSON.stringify(body) : undefined,\n")
	sb.WriteString("    });\n")
	sb.WriteString("    if (!response.ok) {\n")
	sb.WriteString("      throw new Error(`API error: ${response.status} ${response.statusText}`);\n")
	sb.WriteString("    }\n")
	sb.WriteString("    return response.json();\n")
	sb.WriteString("  }\n")

	sb.WriteString("}\n")
}

// generateMethod writes a single client method for a route.
func (g *TypeScriptGenerator) generateMethod(sb *strings.Builder, route *ast.Route, typeDefs []*ast.TypeDef) {
	methodName := routeToMethodName(route)
	httpMethod := strings.ToUpper(route.Method.String())
	pathParams := extractPathParams(route.Path)
	returnType := "unknown"

	if route.ReturnType != nil {
		returnType = glyphTypeToTS(route.ReturnType)
	}

	// Build parameter list
	var params []string
	for _, p := range pathParams {
		params = append(params, fmt.Sprintf("%s: string", p))
	}

	// Add body parameter for POST/PUT/PATCH
	hasBody := httpMethod == "POST" || httpMethod == "PUT" || httpMethod == "PATCH"
	if hasBody {
		bodyType := "Record<string, unknown>"
		if route.InputType != nil {
			bodyType = glyphTypeToTS(route.InputType)
		}
		params = append(params, fmt.Sprintf("body: %s", bodyType))
	}

	paramStr := strings.Join(params, ", ")

	// Build the path with template literals
	tsPath := glyphPathToTemplate(route.Path)

	fmt.Fprintf(sb, "  async %s(%s): Promise<%s> {\n", methodName, paramStr, returnType)
	if hasBody {
		fmt.Fprintf(sb, "    return this.request<%s>(%q, `%s`, body);\n", returnType, httpMethod, tsPath)
	} else {
		fmt.Fprintf(sb, "    return this.request<%s>(%q, `%s`);\n", returnType, httpMethod, tsPath)
	}
	sb.WriteString("  }\n\n")
}

// --- Helper functions ---

// glyphTypeToTS maps a GlyphLang type to a TypeScript type string.
func glyphTypeToTS(t ast.Type) string {
	if t == nil {
		return "unknown"
	}
	switch v := t.(type) {
	case ast.NamedType:
		switch v.Name {
		case "int", "float":
			return "number"
		case "str", "string":
			return "string"
		case "bool":
			return "boolean"
		case "any":
			return "unknown"
		default:
			return v.Name
		}
	case ast.ArrayType:
		elemType := glyphTypeToTS(v.ElementType)
		return elemType + "[]"
	case ast.OptionalType:
		innerType := glyphTypeToTS(v.InnerType)
		return innerType + " | null"
	case ast.UnionType:
		var parts []string
		for _, member := range v.Types {
			parts = append(parts, glyphTypeToTS(member))
		}
		return strings.Join(parts, " | ")
	default:
		return "unknown"
	}
}

// extractPathParams extracts parameter names from a GlyphLang path like /users/:id
func extractPathParams(path string) []string {
	var params []string
	for _, seg := range strings.Split(path, "/") {
		if strings.HasPrefix(seg, ":") {
			params = append(params, seg[1:])
		}
	}
	return params
}

// glyphPathToTemplate converts /users/:id to /users/${id} for TypeScript template literals.
func glyphPathToTemplate(path string) string {
	var parts []string
	for _, seg := range strings.Split(path, "/") {
		if strings.HasPrefix(seg, ":") {
			parts = append(parts, fmt.Sprintf("${%s}", seg[1:]))
		} else {
			parts = append(parts, seg)
		}
	}
	return strings.Join(parts, "/")
}

// routeToMethodName generates a camelCase method name from a route.
func routeToMethodName(route *ast.Route) string {
	method := strings.ToLower(route.Method.String())

	// Build name from path segments
	segments := strings.Split(route.Path, "/")
	var nameParts []string
	for _, seg := range segments {
		if seg == "" || strings.HasPrefix(seg, ":") {
			continue
		}
		nameParts = append(nameParts, capitalize(seg))
	}

	if len(nameParts) == 0 {
		return method
	}

	// Prefix with HTTP method verb
	switch method {
	case "get":
		method = "get"
	case "post":
		method = "create"
	case "put":
		method = "update"
	case "delete":
		method = "delete"
	case "patch":
		method = "patch"
	}

	return method + strings.Join(nameParts, "")
}

// capitalize uppercases the first letter of a string.
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// getRoute extracts a Route from an Item, handling both value and pointer types.
func getRoute(item ast.Item) *ast.Route {
	switch v := item.(type) {
	case ast.Route:
		return &v
	case *ast.Route:
		return v
	}
	return nil
}

// getTypeDef extracts a TypeDef from an Item, handling both value and pointer types.
func getTypeDef(item ast.Item) *ast.TypeDef {
	switch v := item.(type) {
	case ast.TypeDef:
		return &v
	case *ast.TypeDef:
		return v
	}
	return nil
}
