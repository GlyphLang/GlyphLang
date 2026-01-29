package openapi

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"gopkg.in/yaml.v3"
)

// Spec represents an OpenAPI 3.0 specification.
type Spec struct {
	OpenAPI    string                `json:"openapi" yaml:"openapi"`
	Info       Info                  `json:"info" yaml:"info"`
	Paths      map[string]*PathItem  `json:"paths" yaml:"paths"`
	Components *Components           `json:"components,omitempty" yaml:"components,omitempty"`
	Security   []SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
}

// Info contains API metadata.
type Info struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string `json:"version" yaml:"version"`
}

// PathItem represents operations on a single path.
type PathItem struct {
	Get    *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Post   *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Put    *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
}

// Operation represents a single API operation.
type Operation struct {
	Summary     string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	OperationID string                `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]*Response  `json:"responses" yaml:"responses"`
	Security    []SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
	Tags        []string              `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Parameter represents a request parameter (path, query, header).
type Parameter struct {
	Name        string  `json:"name" yaml:"name"`
	In          string  `json:"in" yaml:"in"`
	Required    bool    `json:"required" yaml:"required"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      *Schema `json:"schema" yaml:"schema"`
}

// RequestBody represents a request body.
type RequestBody struct {
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool                 `json:"required" yaml:"required"`
	Content     map[string]MediaType `json:"content" yaml:"content"`
}

// MediaType represents a media type with its schema.
type MediaType struct {
	Schema *Schema `json:"schema" yaml:"schema"`
}

// Response represents a single response.
type Response struct {
	Description string               `json:"description" yaml:"description"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// Schema represents a JSON Schema object.
type Schema struct {
	Type       string             `json:"type,omitempty" yaml:"type,omitempty"`
	Format     string             `json:"format,omitempty" yaml:"format,omitempty"`
	Properties map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Required   []string           `json:"required,omitempty" yaml:"required,omitempty"`
	Items      *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	Ref        string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Nullable   bool               `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	OneOf      []*Schema          `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
}

// Components holds reusable schema definitions.
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
}

// SecurityScheme describes an auth mechanism.
type SecurityScheme struct {
	Type         string `json:"type" yaml:"type"`
	Scheme       string `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	In           string `json:"in,omitempty" yaml:"in,omitempty"`
	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
}

// SecurityRequirement maps security scheme names to scopes.
type SecurityRequirement map[string][]string

// Generator creates OpenAPI specs from GlyphLang modules.
type Generator struct {
	title   string
	version string
}

// NewGenerator creates a new OpenAPI generator with the given API title and version.
func NewGenerator(title, version string) *Generator {
	return &Generator{
		title:   title,
		version: version,
	}
}

// Generate produces an OpenAPI 3.0 spec from a parsed GlyphLang module.
func (g *Generator) Generate(module *interpreter.Module) *Spec {
	spec := &Spec{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:   g.title,
			Version: g.version,
		},
		Paths: make(map[string]*PathItem),
		Components: &Components{
			Schemas:         make(map[string]*Schema),
			SecuritySchemes: make(map[string]*SecurityScheme),
		},
	}

	// First pass: collect all type definitions for schemas
	// Note: the parser may return either value types (interpreter.TypeDef) or pointer
	// types (*interpreter.TypeDef), so getTypeDef (defined at line 474) handles both.
	for _, item := range module.Items {
		td := getTypeDef(item)
		if td != nil {
			spec.Components.Schemas[td.Name] = g.typeDefToSchema(td)
		}
	}

	// Second pass: process routes
	// Note: the parser may return either value types (interpreter.Route) or pointer
	// types (*interpreter.Route), so getRoute (defined at line 462) handles both.
	hasAuth := false
	for _, item := range module.Items {
		route := getRoute(item)
		if route == nil {
			continue
		}
		if route.Method == interpreter.WebSocket {
			continue
		}

		openAPIPath := glyphPathToOpenAPI(route.Path)

		if spec.Paths[openAPIPath] == nil {
			spec.Paths[openAPIPath] = &PathItem{}
		}

		op := g.routeToOperation(route, spec.Components.Schemas)

		switch route.Method {
		case interpreter.Get:
			spec.Paths[openAPIPath].Get = op
		case interpreter.Post:
			spec.Paths[openAPIPath].Post = op
		case interpreter.Put:
			spec.Paths[openAPIPath].Put = op
		case interpreter.Delete:
			spec.Paths[openAPIPath].Delete = op
		case interpreter.Patch:
			spec.Paths[openAPIPath].Patch = op
		}

		if route.Auth != nil {
			hasAuth = true
			g.addSecurityScheme(spec, route.Auth)
		}
	}

	// Clean up empty components
	if len(spec.Components.Schemas) == 0 && len(spec.Components.SecuritySchemes) == 0 {
		spec.Components = nil
	}
	if !hasAuth {
		spec.Security = nil
	}

	return spec
}

// ToJSON serializes the spec to JSON.
func (s *Spec) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// ToYAML serializes the spec to YAML.
func (s *Spec) ToYAML() ([]byte, error) {
	return yaml.Marshal(s)
}

func (g *Generator) routeToOperation(route *interpreter.Route, schemas map[string]*Schema) *Operation {
	op := &Operation{
		OperationID: generateOperationID(route),
		Responses:   make(map[string]*Response),
	}

	// Extract path parameters
	pathParams := extractPathParams(route.Path)
	for _, param := range pathParams {
		op.Parameters = append(op.Parameters, Parameter{
			Name:     param,
			In:       "path",
			Required: true,
			Schema:   &Schema{Type: "string"},
		})
	}

	// Add query parameters
	for _, qp := range route.QueryParams {
		p := Parameter{
			Name:     qp.Name,
			In:       "query",
			Required: qp.Required,
			Schema:   g.typeToSchema(qp.Type),
		}
		if qp.IsArray {
			p.Schema = &Schema{
				Type:  "array",
				Items: p.Schema,
			}
		}
		op.Parameters = append(op.Parameters, p)
	}

	// Handle request body for POST/PUT/PATCH
	if route.Method == interpreter.Post || route.Method == interpreter.Put || route.Method == interpreter.Patch {
		if route.InputType != nil {
			op.RequestBody = &RequestBody{
				Required: true,
				Content: map[string]MediaType{
					"application/json": {
						Schema: g.typeToSchema(route.InputType),
					},
				},
			}
		} else {
			// Default: accept any JSON body
			op.RequestBody = &RequestBody{
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{Type: "object"},
					},
				},
			}
		}
	}

	// Build response schema
	if route.ReturnType != nil {
		responseSchema := g.typeToSchema(route.ReturnType)

		// Check for union type (e.g., User | NotFound)
		if union, ok := route.ReturnType.(interpreter.UnionType); ok {
			op.Responses["200"] = &Response{
				Description: "Successful response",
				Content: map[string]MediaType{
					"application/json": {
						Schema: g.typeToSchema(union.Types[0]),
					},
				},
			}
			// Add error responses for other union types
			for i := 1; i < len(union.Types); i++ {
				statusCode := inferStatusCode(union.Types[i])
				op.Responses[statusCode] = &Response{
					Description: inferDescription(union.Types[i]),
					Content: map[string]MediaType{
						"application/json": {
							Schema: g.typeToSchema(union.Types[i]),
						},
					},
				}
			}
		} else {
			op.Responses["200"] = &Response{
				Description: "Successful response",
				Content: map[string]MediaType{
					"application/json": {
						Schema: responseSchema,
					},
				},
			}
		}
	} else {
		op.Responses["200"] = &Response{
			Description: "Successful response",
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{Type: "object"},
				},
			},
		}
	}

	// Add auth security requirement
	if route.Auth != nil {
		schemeName := authTypeToSchemeName(route.Auth.AuthType)
		op.Security = []SecurityRequirement{
			{schemeName: {}},
		}
	}

	// Derive a tag from the path
	tag := deriveTag(route.Path)
	if tag != "" {
		op.Tags = []string{tag}
	}

	return op
}

func (g *Generator) typeDefToSchema(td *interpreter.TypeDef) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}

	var required []string
	for _, field := range td.Fields {
		fieldSchema := g.typeToSchema(field.TypeAnnotation)

		if field.Required {
			required = append(required, field.Name)
		}

		schema.Properties[field.Name] = fieldSchema
	}

	if len(required) > 0 {
		sort.Strings(required)
		schema.Required = required
	}

	return schema
}

func (g *Generator) typeToSchema(t interpreter.Type) *Schema {
	if t == nil {
		return &Schema{Type: "object"}
	}

	switch typ := t.(type) {
	case interpreter.IntType:
		return &Schema{Type: "integer", Format: "int64"}
	case interpreter.FloatType:
		return &Schema{Type: "number", Format: "double"}
	case interpreter.StringType:
		return &Schema{Type: "string"}
	case interpreter.BoolType:
		return &Schema{Type: "boolean"}
	case interpreter.ArrayType:
		return &Schema{
			Type:  "array",
			Items: g.typeToSchema(typ.ElementType),
		}
	case interpreter.OptionalType:
		inner := g.typeToSchema(typ.InnerType)
		inner.Nullable = true
		return inner
	case interpreter.NamedType:
		return g.namedTypeToSchema(typ.Name)
	case interpreter.UnionType:
		var schemas []*Schema
		for _, ut := range typ.Types {
			schemas = append(schemas, g.typeToSchema(ut))
		}
		return &Schema{OneOf: schemas}
	case interpreter.GenericType:
		return g.genericTypeToSchema(typ)
	case interpreter.DatabaseType:
		return &Schema{Type: "object"}
	default:
		return &Schema{Type: "object"}
	}
}

func (g *Generator) namedTypeToSchema(name string) *Schema {
	switch strings.ToLower(name) {
	case "timestamp", "datetime":
		return &Schema{Type: "string", Format: "date-time"}
	case "date":
		return &Schema{Type: "string", Format: "date"}
	case "email":
		return &Schema{Type: "string", Format: "email"}
	case "url", "uri":
		return &Schema{Type: "string", Format: "uri"}
	case "uuid":
		return &Schema{Type: "string", Format: "uuid"}
	case "any":
		return &Schema{} // no type constraint
	default:
		return &Schema{Ref: "#/components/schemas/" + name}
	}
}

func (g *Generator) genericTypeToSchema(gt interpreter.GenericType) *Schema {
	baseName := ""
	if named, ok := gt.BaseType.(interpreter.NamedType); ok {
		baseName = named.Name
	}

	switch strings.ToLower(baseName) {
	case "list", "array":
		if len(gt.TypeArgs) > 0 {
			return &Schema{
				Type:  "array",
				Items: g.typeToSchema(gt.TypeArgs[0]),
			}
		}
		return &Schema{Type: "array", Items: &Schema{Type: "object"}}
	case "map":
		return &Schema{Type: "object"}
	default:
		// Generic named type - just reference it
		return &Schema{Ref: "#/components/schemas/" + baseName}
	}
}

func (g *Generator) addSecurityScheme(spec *Spec, auth *interpreter.AuthConfig) {
	schemeName := authTypeToSchemeName(auth.AuthType)
	if _, exists := spec.Components.SecuritySchemes[schemeName]; exists {
		return
	}

	switch strings.ToLower(auth.AuthType) {
	case "jwt", "bearer":
		spec.Components.SecuritySchemes[schemeName] = &SecurityScheme{
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		}
	case "basic":
		spec.Components.SecuritySchemes[schemeName] = &SecurityScheme{
			Type:   "http",
			Scheme: "basic",
		}
	case "api-key", "apikey":
		spec.Components.SecuritySchemes[schemeName] = &SecurityScheme{
			Type: "apiKey",
			In:   "header",
			Name: "X-API-Key",
		}
	default:
		spec.Components.SecuritySchemes[schemeName] = &SecurityScheme{
			Type:   "http",
			Scheme: auth.AuthType,
		}
	}
}

// getRoute extracts a *Route from an Item, handling both value and pointer types.
func getRoute(item interpreter.Item) *interpreter.Route {
	switch v := item.(type) {
	case interpreter.Route:
		return &v
	case *interpreter.Route:
		return v
	default:
		return nil
	}
}

// getTypeDef extracts a *TypeDef from an Item, handling both value and pointer types.
func getTypeDef(item interpreter.Item) *interpreter.TypeDef {
	switch v := item.(type) {
	case interpreter.TypeDef:
		return &v
	case *interpreter.TypeDef:
		return v
	default:
		return nil
	}
}

// glyphPathToOpenAPI converts GlyphLang path params (:id) to OpenAPI format ({id}).
func glyphPathToOpenAPI(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

// extractPathParams returns the parameter names from a GlyphLang route path.
func extractPathParams(path string) []string {
	var params []string
	for _, part := range strings.Split(path, "/") {
		if strings.HasPrefix(part, ":") {
			params = append(params, part[1:])
		}
	}
	return params
}

func generateOperationID(route *interpreter.Route) string {
	method := strings.ToLower(route.Method.String())
	parts := strings.Split(route.Path, "/")

	var segments []string
	for _, part := range parts {
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, ":") {
			segments = append(segments, "by"+capitalize(part[1:]))
		} else {
			segments = append(segments, part)
		}
	}

	if len(segments) == 0 {
		return method + "Root"
	}
	return method + capitalize(strings.Join(segments, "_"))
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func authTypeToSchemeName(authType string) string {
	switch strings.ToLower(authType) {
	case "jwt", "bearer":
		return "bearerAuth"
	case "basic":
		return "basicAuth"
	case "api-key", "apikey":
		return "apiKeyAuth"
	default:
		return authType + "Auth"
	}
}

func deriveTag(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	// Use the first two non-empty, non-parameter segments
	for _, part := range parts {
		if part != "" && !strings.HasPrefix(part, ":") && part != "api" {
			return part
		}
	}
	return ""
}

func inferStatusCode(t interpreter.Type) string {
	if named, ok := t.(interpreter.NamedType); ok {
		lower := strings.ToLower(named.Name)
		switch {
		case strings.Contains(lower, "notfound"):
			return "404"
		case strings.Contains(lower, "unauthorized"):
			return "401"
		case strings.Contains(lower, "forbidden"):
			return "403"
		case strings.Contains(lower, "badrequest") || strings.Contains(lower, "validation"):
			return "400"
		case strings.Contains(lower, "conflict"):
			return "409"
		case strings.Contains(lower, "error"):
			return "500"
		}
	}
	return "default"
}

func inferDescription(t interpreter.Type) string {
	if named, ok := t.(interpreter.NamedType); ok {
		lower := strings.ToLower(named.Name)
		switch {
		case strings.Contains(lower, "notfound"):
			return "Not Found"
		case strings.Contains(lower, "unauthorized"):
			return "Unauthorized"
		case strings.Contains(lower, "forbidden"):
			return "Forbidden"
		case strings.Contains(lower, "badrequest"):
			return "Bad Request"
		case strings.Contains(lower, "validation"):
			return "Validation Error"
		case strings.Contains(lower, "conflict"):
			return "Conflict"
		case strings.Contains(lower, "error"):
			return "Error"
		}
		return named.Name
	}
	return "Response"
}

// GenerateFromModule is a convenience function to generate an OpenAPI spec from a module.
func GenerateFromModule(module *interpreter.Module, title, version string) *Spec {
	gen := NewGenerator(title, version)
	return gen.Generate(module)
}

// FormatSpec serializes a spec in the given format ("json" or "yaml").
func FormatSpec(spec *Spec, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "json":
		return spec.ToJSON()
	case "yaml", "yml":
		return spec.ToYAML()
	default:
		return nil, fmt.Errorf("unsupported format: %s (use json or yaml)", format)
	}
}
