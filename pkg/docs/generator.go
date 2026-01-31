package docs

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/ast"
	"strings"
)

// RouteDoc holds documentation data for a single route.
type RouteDoc struct {
	Method      string
	Path        string
	InputType   string
	ReturnType  string
	PathParams  []string
	QueryParams []QueryParamDoc
	HasAuth     bool
	AuthType    string
	RateLimit   string
}

// QueryParamDoc documents a query parameter.
type QueryParamDoc struct {
	Name     string
	Type     string
	Required bool
}

// TypeDoc holds documentation data for a type definition.
type TypeDoc struct {
	Name   string
	Fields []FieldDoc
}

// FieldDoc documents a single field.
type FieldDoc struct {
	Name     string
	Type     string
	Required bool
}

// APIDoc holds all documentation data for a module.
type APIDoc struct {
	Title  string
	Routes []RouteDoc
	Types  []TypeDoc
}

// ExtractDocs extracts documentation data from a parsed module.
func ExtractDocs(module *ast.Module, title string) *APIDoc {
	doc := &APIDoc{Title: title}

	for _, item := range module.Items {
		switch v := item.(type) {
		case *ast.TypeDef:
			doc.Types = append(doc.Types, extractTypeDef(v))
		case ast.TypeDef:
			doc.Types = append(doc.Types, extractTypeDef(&v))
		case *ast.Route:
			// Skip WebSocket routes - they use a different protocol model
			if v.Method != ast.WebSocket {
				doc.Routes = append(doc.Routes, extractRoute(v))
			}
		case ast.Route:
			if v.Method != ast.WebSocket {
				doc.Routes = append(doc.Routes, extractRoute(&v))
			}
		}
	}

	return doc
}

func extractTypeDef(td *ast.TypeDef) TypeDoc {
	tdoc := TypeDoc{Name: td.Name}
	for _, f := range td.Fields {
		fdoc := FieldDoc{
			Name:     f.Name,
			Type:     typeToString(f.TypeAnnotation),
			Required: f.Required,
		}
		tdoc.Fields = append(tdoc.Fields, fdoc)
	}
	return tdoc
}

func extractRoute(r *ast.Route) RouteDoc {
	rd := RouteDoc{
		Method:     strings.ToUpper(r.Method.String()),
		Path:       r.Path,
		PathParams: extractPathParams(r.Path),
	}

	if r.ReturnType != nil {
		rd.ReturnType = typeToString(r.ReturnType)
	}
	if r.InputType != nil {
		rd.InputType = typeToString(r.InputType)
	}

	if r.Auth != nil {
		rd.HasAuth = true
		rd.AuthType = r.Auth.AuthType
	}

	if r.RateLimit != nil {
		rd.RateLimit = fmt.Sprintf("%d requests per %s", r.RateLimit.Requests, r.RateLimit.Window)
	}

	for _, qp := range r.QueryParams {
		rd.QueryParams = append(rd.QueryParams, QueryParamDoc{
			Name:     qp.Name,
			Type:     typeToString(qp.Type),
			Required: qp.Required,
		})
	}

	return rd
}

func extractPathParams(path string) []string {
	var params []string
	for _, seg := range strings.Split(path, "/") {
		if strings.HasPrefix(seg, ":") {
			params = append(params, seg[1:])
		}
	}
	return params
}

func typeToString(t ast.Type) string {
	if t == nil {
		return "any"
	}
	switch v := t.(type) {
	case ast.NamedType:
		return v.Name
	case ast.ArrayType:
		return typeToString(v.ElementType) + "[]"
	case ast.OptionalType:
		return typeToString(v.InnerType) + "?"
	case ast.UnionType:
		var parts []string
		for _, m := range v.Types {
			parts = append(parts, typeToString(m))
		}
		return strings.Join(parts, " | ")
	default:
		return "any"
	}
}
