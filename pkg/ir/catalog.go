package ir

import (
	"fmt"
	"strings"
)

// ServiceCatalog is the Aetheros-compatible flattened representation of a GlyphLang service.
// It contains a list of named operations suitable for ingestion into an Aetheros KnowledgeIndex.
type ServiceCatalog struct {
	Service    string            `json:"service"`
	Version    string            `json:"version"`
	Operations []CatalogOp       `json:"operations"`
	Types      []CatalogType     `json:"types,omitempty"`
	Providers  []CatalogProvider `json:"providers,omitempty"`
}

// CatalogOp is a single named, callable operation in the catalog.
type CatalogOp struct {
	ID        string         `json:"id"`
	Kind      string         `json:"kind"`
	Notation  string         `json:"notation"`
	Method    string         `json:"method,omitempty"`
	Path      string         `json:"path,omitempty"`
	Params    []CatalogParam `json:"params,omitempty"`
	Returns   string         `json:"returns,omitempty"`
	Auth      *CatalogAuth   `json:"auth,omitempty"`
	Providers []string       `json:"providers,omitempty"`
}

// CatalogParam describes a single input parameter for an operation.
type CatalogParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Kind     string `json:"kind,omitempty"` // path | query | body | flag
}

// CatalogAuth describes authentication requirements for an operation.
type CatalogAuth struct {
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Roles    []string `json:"roles,omitempty"`
}

// CatalogType is a compact type definition for the catalog.
type CatalogType struct {
	Name   string         `json:"name"`
	Fields []CatalogParam `json:"fields"`
}

// CatalogProvider is a provider dependency in the catalog.
type CatalogProvider struct {
	Name         string `json:"name"`
	ProviderType string `json:"type"`
	IsStandard   bool   `json:"standard"`
}

// ToCatalog transforms a ServiceIR into an Aetheros-compatible ServiceCatalog.
func ToCatalog(svc *ServiceIR) *ServiceCatalog {
	catalog := &ServiceCatalog{
		Service:    svc.Name,
		Version:    "1.0",
		Operations: make([]CatalogOp, 0),
	}

	for _, route := range svc.Routes {
		catalog.Operations = append(catalog.Operations, routeToOp(route))
	}

	for _, fn := range svc.Functions {
		catalog.Operations = append(catalog.Operations, functionToOp(fn))
	}

	for _, cmd := range svc.Commands {
		catalog.Operations = append(catalog.Operations, commandToOp(cmd))
	}

	for _, ev := range svc.Events {
		catalog.Operations = append(catalog.Operations, eventToOp(ev))
	}

	for _, cron := range svc.CronJobs {
		catalog.Operations = append(catalog.Operations, cronToOp(cron))
	}

	for _, q := range svc.Queues {
		catalog.Operations = append(catalog.Operations, queueToOp(q))
	}

	for _, grpcSvc := range svc.GRPC {
		for _, h := range grpcSvc.Handlers {
			catalog.Operations = append(catalog.Operations, grpcHandlerToOp(grpcSvc.Name, h))
		}
	}

	for _, gql := range svc.GraphQL {
		catalog.Operations = append(catalog.Operations, graphqlToOp(gql))
	}

	if len(svc.Types) > 0 {
		catalog.Types = make([]CatalogType, len(svc.Types))
		for i, t := range svc.Types {
			catalog.Types[i] = typeSchemaToType(t)
		}
	}

	if len(svc.Providers) > 0 {
		catalog.Providers = make([]CatalogProvider, len(svc.Providers))
		for i, p := range svc.Providers {
			catalog.Providers[i] = CatalogProvider{
				Name:         p.Name,
				ProviderType: p.ProviderType,
				IsStandard:   p.IsStandard,
			}
		}
	}

	return catalog
}

func routeToOp(r RouteHandler) CatalogOp {
	method := r.Method.String()
	id := fmt.Sprintf("%s %s", method, r.Path)
	notation := fmt.Sprintf("@ %s %s", method, r.Path)

	var params []CatalogParam
	for _, pp := range r.PathParams {
		params = append(params, CatalogParam{
			Name:     pp,
			Type:     "string",
			Required: true,
			Kind:     "path",
		})
	}
	for _, qp := range r.QueryParams {
		params = append(params, CatalogParam{
			Name:     qp.Name,
			Type:     typeRefToString(&qp.Type),
			Required: qp.Required,
			Kind:     "query",
		})
	}
	if r.InputType != nil {
		params = append(params, CatalogParam{
			Name:     "body",
			Type:     typeRefToString(r.InputType),
			Required: true,
			Kind:     "body",
		})
	}

	var providerTypes []string
	for _, inj := range r.Providers {
		providerTypes = append(providerTypes, inj.ProviderType)
	}

	op := CatalogOp{
		ID:        id,
		Kind:      "route",
		Notation:  notation,
		Method:    method,
		Path:      r.Path,
		Params:    params,
		Providers: providerTypes,
	}
	if r.ReturnType != nil {
		op.Returns = typeRefToString(r.ReturnType)
	}
	if r.Auth != nil {
		op.Auth = authToAuth(r.Auth)
	}
	return op
}

func functionToOp(f FunctionDef) CatalogOp {
	var paramStrs []string
	var params []CatalogParam
	for _, p := range f.Params {
		t := typeRefToString(&p.Type)
		paramStrs = append(paramStrs, p.Name+": "+t)
		params = append(params, CatalogParam{
			Name:     p.Name,
			Type:     t,
			Required: p.Required,
		})
	}

	notation := fmt.Sprintf("! %s(%s)", f.Name, strings.Join(paramStrs, ", "))
	if f.ReturnType != nil {
		notation += ": " + typeRefToString(f.ReturnType)
	}

	op := CatalogOp{
		ID:       "fn:" + f.Name,
		Kind:     "function",
		Notation: notation,
		Params:   params,
	}
	if f.ReturnType != nil {
		op.Returns = typeRefToString(f.ReturnType)
	}
	return op
}

func commandToOp(c CommandDef) CatalogOp {
	var paramStrs []string
	var params []CatalogParam
	for _, p := range c.Params {
		t := typeRefToString(&p.Type)
		prefix := ""
		if p.IsFlag {
			prefix = "--"
		}
		req := ""
		if p.Required {
			req = "!"
		}
		paramStrs = append(paramStrs, prefix+p.Name+": "+t+req)
		kind := ""
		if p.IsFlag {
			kind = "flag"
		}
		params = append(params, CatalogParam{
			Name:     p.Name,
			Type:     t,
			Required: p.Required,
			Kind:     kind,
		})
	}

	notation := fmt.Sprintf("@ command %s %s", c.Name, strings.Join(paramStrs, " "))

	op := CatalogOp{
		ID:       "cmd:" + c.Name,
		Kind:     "command",
		Notation: strings.TrimSpace(notation),
		Params:   params,
	}
	if c.ReturnType != nil {
		op.Returns = typeRefToString(c.ReturnType)
	}
	return op
}

func eventToOp(e EventBinding) CatalogOp {
	var providerTypes []string
	for _, inj := range e.Providers {
		providerTypes = append(providerTypes, inj.ProviderType)
	}
	asyncStr := ""
	if e.Async {
		asyncStr = " async"
	}
	return CatalogOp{
		ID:        "event:" + e.EventType,
		Kind:      "event",
		Notation:  fmt.Sprintf("~%s \"%s\"", asyncStr, e.EventType),
		Providers: providerTypes,
	}
}

func cronToOp(c CronBinding) CatalogOp {
	id := "cron:" + c.Name
	if c.Name == "" {
		id = "cron:" + c.Schedule
	}
	notation := fmt.Sprintf("* \"%s\"", c.Schedule)
	if c.Name != "" {
		notation += " " + c.Name
	}

	var providerTypes []string
	for _, inj := range c.Providers {
		providerTypes = append(providerTypes, inj.ProviderType)
	}
	return CatalogOp{
		ID:        id,
		Kind:      "cron",
		Notation:  notation,
		Providers: providerTypes,
	}
}

func queueToOp(q QueueBinding) CatalogOp {
	var providerTypes []string
	for _, inj := range q.Providers {
		providerTypes = append(providerTypes, inj.ProviderType)
	}
	return CatalogOp{
		ID:        "queue:" + q.QueueName,
		Kind:      "queue",
		Notation:  fmt.Sprintf("& \"%s\"", q.QueueName),
		Providers: providerTypes,
	}
}

func grpcHandlerToOp(serviceName string, h GRPCHandlerDef) CatalogOp {
	var params []CatalogParam
	for _, p := range h.Params {
		params = append(params, CatalogParam{
			Name:     p.Name,
			Type:     typeRefToString(&p.Type),
			Required: p.Required,
		})
	}

	var providerTypes []string
	for _, inj := range h.Providers {
		providerTypes = append(providerTypes, inj.ProviderType)
	}

	op := CatalogOp{
		ID:        fmt.Sprintf("grpc:%s.%s", serviceName, h.MethodName),
		Kind:      "grpc",
		Notation:  fmt.Sprintf("grpc %s.%s", serviceName, h.MethodName),
		Params:    params,
		Providers: providerTypes,
	}
	if h.ReturnType != nil {
		op.Returns = typeRefToString(h.ReturnType)
	}
	if h.Auth != nil {
		op.Auth = authToAuth(h.Auth)
	}
	return op
}

func graphqlToOp(g GraphQLDef) CatalogOp {
	var params []CatalogParam
	for _, p := range g.Params {
		params = append(params, CatalogParam{
			Name:     p.Name,
			Type:     typeRefToString(&p.Type),
			Required: p.Required,
		})
	}

	var providerTypes []string
	for _, inj := range g.Providers {
		providerTypes = append(providerTypes, inj.ProviderType)
	}

	opStr := g.Operation.String()
	op := CatalogOp{
		ID:        fmt.Sprintf("graphql:%s:%s", opStr, g.FieldName),
		Kind:      "graphql",
		Notation:  fmt.Sprintf("graphql %s %s", opStr, g.FieldName),
		Params:    params,
		Providers: providerTypes,
	}
	if g.ReturnType != nil {
		op.Returns = typeRefToString(g.ReturnType)
	}
	if g.Auth != nil {
		op.Auth = authToAuth(g.Auth)
	}
	return op
}

func typeSchemaToType(t TypeSchema) CatalogType {
	fields := make([]CatalogParam, len(t.Fields))
	for i, f := range t.Fields {
		fields[i] = CatalogParam{
			Name:     f.Name,
			Type:     typeRefToString(&f.Type),
			Required: f.Required,
		}
	}
	return CatalogType{
		Name:   t.Name,
		Fields: fields,
	}
}

// typeRefToString converts a TypeRef to its GlyphLang source notation.
func typeRefToString(t *TypeRef) string {
	if t == nil {
		return "any"
	}
	switch t.Kind {
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float"
	case TypeString:
		return "string"
	case TypeBool:
		return "bool"
	case TypeAny:
		return "any"
	case TypeNamed, TypeProvider:
		return t.Name
	case TypeArray:
		return "[" + typeRefToString(t.Inner) + "]"
	case TypeOptional:
		return typeRefToString(t.Inner) + "?"
	case TypeFuture:
		return "Future<" + typeRefToString(t.Inner) + ">"
	case TypeUnion:
		parts := make([]string, len(t.Elements))
		for i, el := range t.Elements {
			elCopy := el
			parts[i] = typeRefToString(&elCopy)
		}
		return strings.Join(parts, " | ")
	case TypeGeneric:
		if len(t.Elements) == 0 {
			return t.Name
		}
		args := make([]string, len(t.Elements))
		for i, el := range t.Elements {
			elCopy := el
			args[i] = typeRefToString(&elCopy)
		}
		return t.Name + "<" + strings.Join(args, ", ") + ">"
	case TypeFunction:
		if len(t.Elements) == 0 {
			return "() -> " + typeRefToString(t.Inner)
		}
		params := make([]string, len(t.Elements))
		for i, el := range t.Elements {
			elCopy := el
			params[i] = typeRefToString(&elCopy)
		}
		return "(" + strings.Join(params, ", ") + ") -> " + typeRefToString(t.Inner)
	default:
		return "any"
	}
}

func authToAuth(a *AuthRequirement) *CatalogAuth {
	if a == nil {
		return nil
	}
	return &CatalogAuth{
		Type:     a.AuthType,
		Required: a.Required,
		Roles:    a.Roles,
	}
}
