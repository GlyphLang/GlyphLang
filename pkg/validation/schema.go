package validation

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"math"
	"net/url"
	"regexp"
	"strings"
)

// Pre-compiled regex patterns used for validation.
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// SchemaFromTypeDef builds a Validator from a TypeDef's field annotations.
// Each field's @-annotations (e.g., @minLen(2), @email) are converted into
// the corresponding validation rules on the returned Validator.
func SchemaFromTypeDef(td *interpreter.TypeDef) *Validator {
	v := NewValidator()

	for _, field := range td.Fields {
		if field.Required {
			v.AddRequiredRule(field.Name)
		}

		for _, ann := range field.Annotations {
			addAnnotationRule(v, field.Name, ann)
		}
	}

	return v
}

func addAnnotationRule(v *Validator, field string, ann interpreter.FieldAnnotation) {
	switch ann.Name {
	case "minLen":
		if n, ok := paramInt(ann.Params, 0); ok {
			v.AddLengthRule(field, n, 0)
		}
	case "maxLen":
		if n, ok := paramInt(ann.Params, 0); ok {
			v.AddLengthRule(field, 0, n)
		}
	case "len":
		if n, ok := paramInt(ann.Params, 0); ok {
			v.AddLengthRule(field, n, n)
		}
	case "min":
		if minVal, ok := paramFloat(ann.Params, 0); ok {
			v.AddCustomRule(field, func(val interface{}) error {
				n, ok := toFloat64Value(val)
				if !ok {
					return &ValidationError{Message: "must be a number"}
				}
				if n < minVal {
					return &ValidationError{Message: fmt.Sprintf("must be at least %v", minVal)}
				}
				return nil
			})
		}
	case "max":
		if maxVal, ok := paramFloat(ann.Params, 0); ok {
			v.AddCustomRule(field, func(val interface{}) error {
				n, ok := toFloat64Value(val)
				if !ok {
					return &ValidationError{Message: "must be a number"}
				}
				if n > maxVal {
					return &ValidationError{Message: fmt.Sprintf("must be at most %v", maxVal)}
				}
				return nil
			})
		}
	case "range":
		minVal, ok1 := paramFloat(ann.Params, 0)
		maxVal, ok2 := paramFloat(ann.Params, 1)
		if ok1 && ok2 {
			v.AddRangeRule(field, minVal, maxVal)
		}
	case "pattern":
		if pat, ok := paramString(ann.Params, 0); ok {
			// Validate the regex at schema build time
			if _, err := regexp.Compile(pat); err == nil {
				v.AddPatternRule(field, pat)
			}
		}
	case "email":
		v.AddEmailRule(field)
	case "url":
		v.AddCustomRule(field, func(val interface{}) error {
			s, ok := val.(string)
			if !ok {
				return &ValidationError{Message: "must be a string for URL validation"}
			}
			u, err := url.Parse(s)
			if err != nil || u.Scheme == "" || u.Host == "" {
				return &ValidationError{Message: "must be a valid URL"}
			}
			return nil
		})
	case "notEmpty":
		v.AddCustomRule(field, func(val interface{}) error {
			s, ok := val.(string)
			if !ok {
				return nil
			}
			if strings.TrimSpace(s) == "" {
				return &ValidationError{Message: "must not be empty"}
			}
			return nil
		})
	case "uuid":
		v.AddCustomRule(field, func(val interface{}) error {
			s, ok := val.(string)
			if !ok {
				return &ValidationError{Message: "must be a string for UUID validation"}
			}
			if !uuidRegex.MatchString(s) {
				return &ValidationError{Message: "must be a valid UUID"}
			}
			return nil
		})
	case "oneOf":
		if items, ok := paramStringSlice(ann.Params, 0); ok {
			v.AddCustomRule(field, func(val interface{}) error {
				s, ok := val.(string)
				if !ok {
					return &ValidationError{Message: "must be a string for oneOf validation"}
				}
				for _, item := range items {
					if s == item {
						return nil
					}
				}
				return &ValidationError{
					Message: fmt.Sprintf("must be one of: %s", strings.Join(items, ", ")),
				}
			})
		}
	case "positive":
		v.AddCustomRule(field, func(val interface{}) error {
			n, ok := toFloat64Value(val)
			if !ok {
				return &ValidationError{Message: "must be a number"}
			}
			if n <= 0 {
				return &ValidationError{Message: "must be positive"}
			}
			return nil
		})
	case "negative":
		v.AddCustomRule(field, func(val interface{}) error {
			n, ok := toFloat64Value(val)
			if !ok {
				return &ValidationError{Message: "must be a number"}
			}
			if n >= 0 {
				return &ValidationError{Message: "must be negative"}
			}
			return nil
		})
	case "integer":
		v.AddCustomRule(field, func(val interface{}) error {
			switch v := val.(type) {
			case int, int64, int32:
				return nil
			case float64:
				// Use math.Floor comparison to avoid floating point precision issues
				if v != math.Floor(v) {
					return &ValidationError{Message: "must be an integer"}
				}
				return nil
			}
			return &ValidationError{Message: "must be an integer"}
		})
	case "minItems":
		if n, ok := paramInt(ann.Params, 0); ok {
			v.AddCustomRule(field, func(val interface{}) error {
				arr, ok := val.([]interface{})
				if !ok {
					return &ValidationError{Message: "must be an array"}
				}
				if len(arr) < n {
					return &ValidationError{
						Message: fmt.Sprintf("must have at least %d items (got %d)", n, len(arr)),
					}
				}
				return nil
			})
		}
	case "maxItems":
		if n, ok := paramInt(ann.Params, 0); ok {
			v.AddCustomRule(field, func(val interface{}) error {
				arr, ok := val.([]interface{})
				if !ok {
					return &ValidationError{Message: "must be an array"}
				}
				if len(arr) > n {
					return &ValidationError{
						Message: fmt.Sprintf("must have at most %d items (got %d)", n, len(arr)),
					}
				}
				return nil
			})
		}
	case "unique":
		v.AddCustomRule(field, func(val interface{}) error {
			arr, ok := val.([]interface{})
			if !ok {
				return &ValidationError{Message: "must be an array"}
			}
			seen := make(map[interface{}]bool)
			for _, item := range arr {
				if seen[item] {
					return &ValidationError{Message: "array items must be unique"}
				}
				seen[item] = true
			}
			return nil
		})
	}
}

// ValidationResult holds structured validation error output.
type ValidationResult struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields"`
}

// ToResult converts ValidationErrors into a structured result.
func ToResult(err error) *ValidationResult {
	verrs, ok := err.(*ValidationErrors)
	if !ok {
		return &ValidationResult{Error: err.Error(), Fields: map[string]string{}}
	}
	result := &ValidationResult{
		Error:  "Validation failed",
		Fields: make(map[string]string),
	}
	for _, e := range verrs.Errors {
		result.Fields[e.Field] = e.Message
	}
	return result
}

// paramInt safely extracts an integer parameter at the given index.
func paramInt(params []interface{}, idx int) (int, bool) {
	if idx >= len(params) {
		return 0, false
	}
	return toInt(params[idx])
}

// paramFloat safely extracts a float parameter at the given index.
func paramFloat(params []interface{}, idx int) (float64, bool) {
	if idx >= len(params) {
		return 0, false
	}
	return toFloat(params[idx])
}

// paramString safely extracts a string parameter at the given index.
func paramString(params []interface{}, idx int) (string, bool) {
	if idx >= len(params) {
		return "", false
	}
	s, ok := params[idx].(string)
	return s, ok
}

// paramStringSlice safely extracts a string slice parameter at the given index.
func paramStringSlice(params []interface{}, idx int) ([]string, bool) {
	if idx >= len(params) {
		return nil, false
	}
	s, ok := params[idx].([]string)
	return s, ok
}

func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int64:
		return int(n), true
	case int:
		return n, true
	case float64:
		// Only convert if the float is a whole number
		if n == math.Floor(n) {
			return int(n), true
		}
		return 0, false
	}
	return 0, false
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	}
	return 0, false
}

func toFloat64Value(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	}
	return 0, false
}
