package contract

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"strings"
)

// Violation represents a contract violation found during verification
type Violation struct {
	Endpoint string // e.g. "GET /users/:id"
	Message  string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s: %s", v.Endpoint, v.Message)
}

// VerifyResult contains the result of contract verification
type VerifyResult struct {
	ContractName string
	Violations   []Violation
	Passed       bool
}

// BreakingChange represents a breaking change between two contract versions
type BreakingChange struct {
	Type     string // "removed", "return_type_changed", "added" (informational)
	Endpoint string
	Detail   string
}

func (b BreakingChange) String() string {
	return fmt.Sprintf("[%s] %s: %s", b.Type, b.Endpoint, b.Detail)
}

// DiffResult contains the differences between two contracts
type DiffResult struct {
	OldName         string
	NewName         string
	BreakingChanges []BreakingChange
	Additions       []BreakingChange
	HasBreaking     bool
}

// ConsumerExpectation represents what a consumer expects from a service
type ConsumerExpectation struct {
	Method     string
	Path       string
	StatusCode int
	ReturnType string
}

// ConsumerResult contains the result of consumer contract verification
type ConsumerResult struct {
	Consumer string
	Provider string
	Failures []string
	Passed   bool
}

// Verify checks if the routes in a module satisfy a contract definition.
// It compares each contract endpoint against the routes found in the module's items.
func Verify(contract interpreter.ContractDef, moduleItems []interpreter.Item) VerifyResult {
	result := VerifyResult{
		ContractName: contract.Name,
	}

	// Build a map of routes from module items
	routeMap := buildRouteMap(moduleItems)

	for _, endpoint := range contract.Endpoints {
		key := endpointKey(endpoint.Method.String(), endpoint.Path)
		route, found := routeMap[key]
		if !found {
			result.Violations = append(result.Violations, Violation{
				Endpoint: key,
				Message:  "endpoint not found in implementation",
			})
			continue
		}

		// Check return type compatibility
		if endpoint.ReturnType != nil && route.ReturnType != nil {
			if !typesCompatible(endpoint.ReturnType, route.ReturnType) {
				result.Violations = append(result.Violations, Violation{
					Endpoint: key,
					Message: fmt.Sprintf("return type mismatch: contract expects %s, implementation returns %s",
						typeString(endpoint.ReturnType), typeString(route.ReturnType)),
				})
			}
		}
	}

	result.Passed = len(result.Violations) == 0
	return result
}

// Diff compares two contracts and identifies breaking changes.
// A breaking change is a removal or return type change of an existing endpoint.
// An addition is a new endpoint not in the old contract.
func Diff(oldContract, newContract interpreter.ContractDef) DiffResult {
	result := DiffResult{
		OldName: oldContract.Name,
		NewName: newContract.Name,
	}

	oldEndpoints := make(map[string]interpreter.ContractEndpoint)
	for _, ep := range oldContract.Endpoints {
		key := endpointKey(ep.Method.String(), ep.Path)
		oldEndpoints[key] = ep
	}

	newEndpoints := make(map[string]interpreter.ContractEndpoint)
	for _, ep := range newContract.Endpoints {
		key := endpointKey(ep.Method.String(), ep.Path)
		newEndpoints[key] = ep
	}

	// Check for removed endpoints (breaking)
	for key, oldEp := range oldEndpoints {
		newEp, exists := newEndpoints[key]
		if !exists {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Type:     "removed",
				Endpoint: key,
				Detail:   "endpoint was removed",
			})
			continue
		}

		// Check return type changes
		if oldEp.ReturnType != nil && newEp.ReturnType != nil {
			if !typesCompatible(oldEp.ReturnType, newEp.ReturnType) {
				result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
					Type:     "return_type_changed",
					Endpoint: key,
					Detail: fmt.Sprintf("return type changed from %s to %s",
						typeString(oldEp.ReturnType), typeString(newEp.ReturnType)),
				})
			}
		}
	}

	// Check for added endpoints (informational, not breaking)
	for key := range newEndpoints {
		if _, exists := oldEndpoints[key]; !exists {
			result.Additions = append(result.Additions, BreakingChange{
				Type:     "added",
				Endpoint: key,
				Detail:   "new endpoint added",
			})
		}
	}

	result.HasBreaking = len(result.BreakingChanges) > 0
	return result
}

// VerifyConsumer checks if a provider contract satisfies consumer expectations.
func VerifyConsumer(consumer string, contract interpreter.ContractDef, expectations []ConsumerExpectation) ConsumerResult {
	result := ConsumerResult{
		Consumer: consumer,
		Provider: contract.Name,
	}

	endpointMap := make(map[string]interpreter.ContractEndpoint)
	for _, ep := range contract.Endpoints {
		key := endpointKey(ep.Method.String(), ep.Path)
		endpointMap[key] = ep
	}

	for _, exp := range expectations {
		key := endpointKey(exp.Method, exp.Path)
		ep, found := endpointMap[key]
		if !found {
			result.Failures = append(result.Failures, fmt.Sprintf(
				"%s: expected endpoint not found in provider contract", key))
			continue
		}

		// Check return type if specified
		if exp.ReturnType != "" && ep.ReturnType != nil {
			actual := typeString(ep.ReturnType)
			if !strings.Contains(actual, exp.ReturnType) {
				result.Failures = append(result.Failures, fmt.Sprintf(
					"%s: expected return type containing %s, got %s", key, exp.ReturnType, actual))
			}
		}
	}

	result.Passed = len(result.Failures) == 0
	return result
}

// buildRouteMap extracts routes from module items into a lookup map
func buildRouteMap(items []interpreter.Item) map[string]*interpreter.Route {
	routes := make(map[string]*interpreter.Route)
	for _, item := range items {
		if route, ok := item.(*interpreter.Route); ok {
			key := endpointKey(route.Method.String(), route.Path)
			routes[key] = route
		}
	}
	return routes
}

// endpointKey creates a canonical key for an endpoint: "GET /users/:id"
func endpointKey(method, path string) string {
	return method + " " + path
}

// typeString returns a human-readable string for a type
func typeString(t interpreter.Type) string {
	switch tt := t.(type) {
	case interpreter.IntType:
		return "int"
	case interpreter.StringType:
		return "string"
	case interpreter.BoolType:
		return "bool"
	case interpreter.FloatType:
		return "float"
	case interpreter.NamedType:
		return tt.Name
	case interpreter.ArrayType:
		return "[" + typeString(tt.ElementType) + "]"
	case interpreter.OptionalType:
		return typeString(tt.InnerType) + "?"
	case interpreter.UnionType:
		parts := make([]string, len(tt.Types))
		for i, ut := range tt.Types {
			parts[i] = typeString(ut)
		}
		return strings.Join(parts, " | ")
	default:
		return fmt.Sprintf("%T", t)
	}
}

// typesCompatible checks if two types are compatible for contract verification.
// A contract type is compatible with an implementation type if:
// - They are the same type
// - The implementation type is a union containing the contract type
// - Both are union types with matching members
func typesCompatible(contractType, implType interpreter.Type) bool {
	// Same concrete type
	if typeString(contractType) == typeString(implType) {
		return true
	}

	// Union type compatibility: contract union must be subset of impl union
	contractUnion, contractIsUnion := contractType.(interpreter.UnionType)
	implUnion, implIsUnion := implType.(interpreter.UnionType)

	if contractIsUnion && implIsUnion {
		// All contract union members must be in impl union
		for _, ct := range contractUnion.Types {
			found := false
			for _, it := range implUnion.Types {
				if typeString(ct) == typeString(it) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}

	// Impl is a union containing contract type
	if implIsUnion {
		for _, it := range implUnion.Types {
			if typeString(contractType) == typeString(it) {
				return true
			}
		}
	}

	return false
}
