package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"fmt"
	"strconv"
	"strings"
)

// ProcessQueryParams processes raw query params according to declarations.
// It performs type conversion for declared params and auto-conversion for undeclared ones.
func ProcessQueryParams(
	rawParams map[string][]string,
	declarations []QueryParamDecl,
) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Track which params are declared
	declaredNames := make(map[string]bool)
	for _, decl := range declarations {
		declaredNames[decl.Name] = true
	}

	// Process declared parameters with type conversion
	for _, decl := range declarations {
		values, exists := rawParams[decl.Name]

		if !exists || len(values) == 0 {
			// Check for required params without defaults
			if decl.Required && decl.Default == nil {
				return nil, fmt.Errorf("required query parameter missing: %s", decl.Name)
			}
			// Default will be applied by the interpreter
			continue
		}

		if decl.IsArray {
			// Convert all values to the element type
			converted, err := convertToArray(values, decl.Type)
			if err != nil {
				return nil, fmt.Errorf("query param %s: %w", decl.Name, err)
			}
			result[decl.Name] = converted
		} else {
			// Single value - take first
			converted, err := convertValue(values[0], decl.Type)
			if err != nil {
				return nil, fmt.Errorf("query param %s: %w", decl.Name, err)
			}
			result[decl.Name] = converted
		}
	}

	// Add undeclared params with auto-conversion (backward compatibility)
	for name, values := range rawParams {
		if !declaredNames[name] {
			if len(values) == 1 {
				result[name] = autoConvert(values[0])
			} else {
				// Multiple values - keep as string array
				result[name] = values
			}
		}
	}

	return result, nil
}

// convertValue converts a string value to the specified type
func convertValue(value string, targetType Type) (interface{}, error) {
	switch targetType.(type) {
	case IntType:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer value: %s", value)
		}
		return i, nil
	case FloatType:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float value: %s", value)
		}
		return f, nil
	case BoolType:
		b, err := parseBool(value)
		if err != nil {
			return nil, err
		}
		return b, nil
	case StringType:
		return value, nil
	case ArrayType:
		// Single value for array type - wrap in slice
		at := targetType.(ArrayType)
		converted, err := convertValue(value, at.ElementType)
		if err != nil {
			return nil, err
		}
		return []interface{}{converted}, nil
	default:
		// Unknown type - return as string
		return value, nil
	}
}

// parseBool handles various boolean representations
func parseBool(s string) (bool, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off", "":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", s)
	}
}

// autoConvert auto-detects and converts value types (existing behavior)
func autoConvert(value string) interface{} {
	// Try int
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i
	}
	// Try float (only if contains decimal point)
	if strings.Contains(value, ".") {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	// Try bool
	lower := strings.ToLower(value)
	if lower == "true" || lower == "false" {
		return lower == "true"
	}
	// Return as string
	return value
}

// convertToArray converts multiple string values to typed array
func convertToArray(values []string, arrayType Type) ([]interface{}, error) {
	var elementType Type
	if at, ok := arrayType.(ArrayType); ok {
		elementType = at.ElementType
	} else {
		elementType = StringType{}
	}

	result := make([]interface{}, len(values))
	for i, v := range values {
		converted, err := convertValue(v, elementType)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		result[i] = converted
	}
	return result, nil
}

// ExtractRawQueryParams extracts all query parameter values from URL path
func ExtractRawQueryParams(path string) map[string][]string {
	result := make(map[string][]string)

	idx := strings.Index(path, "?")
	if idx == -1 {
		return result
	}

	queryString := path[idx+1:]
	if queryString == "" {
		return result
	}

	pairs := strings.Split(queryString, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		key := parts[0]
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		result[key] = append(result[key], value)
	}

	return result
}
