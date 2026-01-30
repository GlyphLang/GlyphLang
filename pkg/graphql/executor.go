package graphql

import (
	"fmt"

	"github.com/glyphlang/glyph/pkg/interpreter"
)

// Executor handles GraphQL query execution against a schema using the Glyph interpreter.
type Executor struct {
	schema *Schema
	interp *interpreter.Interpreter
}

// NewExecutor creates a new GraphQL executor with the given schema and interpreter.
func NewExecutor(schema *Schema, interp *interpreter.Interpreter) *Executor {
	return &Executor{
		schema: schema,
		interp: interp,
	}
}

// GraphQLRequest represents an incoming GraphQL request.
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response.
type GraphQLResponse struct {
	Data   interface{}      `json:"data,omitempty"`
	Errors []GraphQLError   `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error.
type GraphQLError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

// Execute processes a GraphQL request and returns a response.
func (e *Executor) Execute(req GraphQLRequest, authData map[string]interface{}) *GraphQLResponse {
	// Parse the query string
	query, err := ParseQuery(req.Query)
	if err != nil {
		return &GraphQLResponse{
			Errors: []GraphQLError{{Message: fmt.Sprintf("parse error: %v", err)}},
		}
	}

	// Merge request variables into query
	if req.Variables != nil {
		for k, v := range req.Variables {
			query.Variables[k] = v
		}
	}

	// Determine root type
	var rootType *ObjectType
	switch query.OperationType {
	case "query":
		rootType = e.schema.Query
	case "mutation":
		rootType = e.schema.Mutation
	default:
		return &GraphQLResponse{
			Errors: []GraphQLError{{Message: fmt.Sprintf("unsupported operation type: %s", query.OperationType)}},
		}
	}

	if rootType == nil {
		return &GraphQLResponse{
			Errors: []GraphQLError{{Message: fmt.Sprintf("no %s type defined in schema", query.OperationType)}},
		}
	}

	// Resolve each top-level selection
	data := make(map[string]interface{})
	var errors []GraphQLError

	for _, sel := range query.Selections {
		resolverKey := query.OperationType + "." + sel.Name
		resolver, ok := e.schema.Resolvers[resolverKey]
		if !ok {
			errors = append(errors, GraphQLError{
				Message: fmt.Sprintf("no resolver for field: %s.%s", query.OperationType, sel.Name),
				Path:    []string{sel.EffectiveName()},
			})
			continue
		}

		// Prepare arguments with variable substitution
		args := make(map[string]interface{})
		for k, v := range sel.Args {
			args[k] = v
		}

		// Execute resolver via interpreter
		result, err := e.interp.ExecuteGraphQLResolver(&resolver, args, authData)
		if err != nil {
			errors = append(errors, GraphQLError{
				Message: err.Error(),
				Path:    []string{sel.EffectiveName()},
			})
			continue
		}

		// Apply field selections to result
		resolved := e.applySelections(result, sel.Selections)
		data[sel.EffectiveName()] = resolved
	}

	resp := &GraphQLResponse{Data: data}
	if len(errors) > 0 {
		resp.Errors = errors
	}
	return resp
}

// applySelections filters an object result to only include requested fields.
// If no sub-selections, returns the value as-is.
func (e *Executor) applySelections(value interface{}, selections []Selection) interface{} {
	if len(selections) == 0 {
		return value
	}

	switch v := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for _, sel := range selections {
			if fieldVal, ok := v[sel.Name]; ok {
				result[sel.EffectiveName()] = e.applySelections(fieldVal, sel.Selections)
			}
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, elem := range v {
			result[i] = e.applySelections(elem, selections)
		}
		return result

	default:
		return value
	}
}

// Introspect returns a map of the schema types for introspection queries.
func (e *Executor) Introspect() map[string]interface{} {
	types := make([]interface{}, 0, len(e.schema.Types))
	for _, objType := range e.schema.Types {
		fields := make([]interface{}, 0, len(objType.Fields))
		for _, field := range objType.Fields {
			fieldInfo := map[string]interface{}{
				"name": field.Name,
				"type": field.Type,
			}
			if len(field.Args) > 0 {
				argInfos := make([]interface{}, len(field.Args))
				for i, arg := range field.Args {
					argInfos[i] = map[string]interface{}{
						"name":     arg.Name,
						"type":     arg.Type,
						"required": arg.Required,
					}
				}
				fieldInfo["args"] = argInfos
			}
			fields = append(fields, fieldInfo)
		}
		types = append(types, map[string]interface{}{
			"name":   objType.Name,
			"fields": fields,
		})
	}

	result := map[string]interface{}{
		"types": types,
	}
	if e.schema.Query != nil {
		result["queryType"] = e.schema.Query.Name
	}
	if e.schema.Mutation != nil {
		result["mutationType"] = e.schema.Mutation.Name
	}
	return result
}
