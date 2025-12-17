package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []*ValidationError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation errors occurred"
	}

	messages := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

func (e *ValidationErrors) Add(field, message string) {
	e.Errors = append(e.Errors, &ValidationError{
		Field:   field,
		Message: message,
	})
}

func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// ValidateEmail validates an email address according to RFC 5322
func ValidateEmail(s string) error {
	if s == "" {
		return &ValidationError{Message: "email cannot be empty"}
	}

	// RFC 5322 compliant email regex (simplified but comprehensive)
	// This pattern covers most valid email formats
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	if !emailRegex.MatchString(s) {
		return &ValidationError{Message: "invalid email format"}
	}

	// Additional validation rules
	if len(s) > 254 {
		return &ValidationError{Message: "email too long (max 254 characters)"}
	}

	// Check local part (before @) length
	parts := strings.Split(s, "@")
	if len(parts) != 2 {
		return &ValidationError{Message: "invalid email format"}
	}

	if len(parts[0]) > 64 {
		return &ValidationError{Message: "email local part too long (max 64 characters)"}
	}

	// Check domain part has at least one dot
	if !strings.Contains(parts[1], ".") {
		return &ValidationError{Message: "email domain must contain at least one dot"}
	}

	return nil
}

// ValidateLength validates string length constraints
func ValidateLength(s string, min, max int) error {
	length := len(s)

	if min > 0 && length < min {
		return &ValidationError{
			Message: fmt.Sprintf("length must be at least %d characters (got %d)", min, length),
		}
	}

	if max > 0 && length > max {
		return &ValidationError{
			Message: fmt.Sprintf("length must be at most %d characters (got %d)", max, length),
		}
	}

	return nil
}

// ValidateRange validates numeric range constraints
func ValidateRange(n float64, min, max float64) error {
	if min != 0 && n < min {
		return &ValidationError{
			Message: fmt.Sprintf("value must be at least %v (got %v)", min, n),
		}
	}

	if max != 0 && n > max {
		return &ValidationError{
			Message: fmt.Sprintf("value must be at most %v (got %v)", max, n),
		}
	}

	return nil
}

// ValidatePattern validates string against a regex pattern
func ValidatePattern(s string, pattern string) error {
	if pattern == "" {
		return &ValidationError{Message: "pattern cannot be empty"}
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return &ValidationError{Message: fmt.Sprintf("invalid regex pattern: %v", err)}
	}

	if !regex.MatchString(s) {
		return &ValidationError{
			Message: fmt.Sprintf("value does not match pattern '%s'", pattern),
		}
	}

	return nil
}

// ValidateRequired validates that a value is not nil or empty
func ValidateRequired(v interface{}) error {
	if v == nil {
		return &ValidationError{Message: "value is required"}
	}

	switch val := v.(type) {
	case string:
		if val == "" {
			return &ValidationError{Message: "value cannot be empty"}
		}
	case []interface{}:
		if len(val) == 0 {
			return &ValidationError{Message: "array cannot be empty"}
		}
	case map[string]interface{}:
		if len(val) == 0 {
			return &ValidationError{Message: "object cannot be empty"}
		}
	}

	return nil
}

// ValidationRule represents a single validation rule
type ValidationRule struct {
	Field      string
	RuleType   string // "email", "length", "range", "pattern", "required", "custom"
	Min        interface{}
	Max        interface{}
	Pattern    string
	CustomFunc func(interface{}) error
}

// Validator manages validation rules and execution
type Validator struct {
	rules map[string][]ValidationRule
}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string][]ValidationRule),
	}
}

// AddRule adds a validation rule for a field
func (v *Validator) AddRule(field string, rule ValidationRule) {
	rule.Field = field
	v.rules[field] = append(v.rules[field], rule)
}

// AddEmailRule adds an email validation rule
func (v *Validator) AddEmailRule(field string) {
	v.AddRule(field, ValidationRule{
		RuleType: "email",
	})
}

// AddLengthRule adds a length validation rule
func (v *Validator) AddLengthRule(field string, min, max int) {
	v.AddRule(field, ValidationRule{
		RuleType: "length",
		Min:      min,
		Max:      max,
	})
}

// AddRangeRule adds a range validation rule
func (v *Validator) AddRangeRule(field string, min, max float64) {
	v.AddRule(field, ValidationRule{
		RuleType: "range",
		Min:      min,
		Max:      max,
	})
}

// AddPatternRule adds a pattern validation rule
func (v *Validator) AddPatternRule(field string, pattern string) {
	v.AddRule(field, ValidationRule{
		RuleType: "pattern",
		Pattern:  pattern,
	})
}

// AddRequiredRule adds a required validation rule
func (v *Validator) AddRequiredRule(field string) {
	v.AddRule(field, ValidationRule{
		RuleType: "required",
	})
}

// AddCustomRule adds a custom validation rule
func (v *Validator) AddCustomRule(field string, customFunc func(interface{}) error) {
	v.AddRule(field, ValidationRule{
		RuleType:   "custom",
		CustomFunc: customFunc,
	})
}

// Validate validates a map of values against the stored rules
func (v *Validator) Validate(data map[string]interface{}) error {
	errors := &ValidationErrors{}

	for field, rules := range v.rules {
		value, exists := data[field]

		for _, rule := range rules {
			var err error

			switch rule.RuleType {
			case "required":
				if !exists {
					errors.Add(field, "field is required")
					continue
				}
				err = ValidateRequired(value)

			case "email":
				if !exists {
					continue // Skip if not required
				}
				str, ok := value.(string)
				if !ok {
					errors.Add(field, "field must be a string for email validation")
					continue
				}
				err = ValidateEmail(str)

			case "length":
				if !exists {
					continue // Skip if not required
				}
				str, ok := value.(string)
				if !ok {
					errors.Add(field, "field must be a string for length validation")
					continue
				}
				min, max := 0, 0
				if rule.Min != nil {
					min = rule.Min.(int)
				}
				if rule.Max != nil {
					max = rule.Max.(int)
				}
				err = ValidateLength(str, min, max)

			case "range":
				if !exists {
					continue // Skip if not required
				}
				var num float64
				switch val := value.(type) {
				case float64:
					num = val
				case int:
					num = float64(val)
				case int64:
					num = float64(val)
				default:
					errors.Add(field, "field must be a number for range validation")
					continue
				}
				min, max := 0.0, 0.0
				if rule.Min != nil {
					switch m := rule.Min.(type) {
					case float64:
						min = m
					case int:
						min = float64(m)
					}
				}
				if rule.Max != nil {
					switch m := rule.Max.(type) {
					case float64:
						max = m
					case int:
						max = float64(m)
					}
				}
				err = ValidateRange(num, min, max)

			case "pattern":
				if !exists {
					continue // Skip if not required
				}
				str, ok := value.(string)
				if !ok {
					errors.Add(field, "field must be a string for pattern validation")
					continue
				}
				err = ValidatePattern(str, rule.Pattern)

			case "custom":
				if !exists {
					continue // Skip if not required
				}
				if rule.CustomFunc != nil {
					err = rule.CustomFunc(value)
				}
			}

			if err != nil {
				if validationErr, ok := err.(*ValidationError); ok {
					errors.Add(field, validationErr.Message)
				} else {
					errors.Add(field, err.Error())
				}
			}
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidateField validates a single field value against stored rules
func (v *Validator) ValidateField(field string, value interface{}) error {
	rules, exists := v.rules[field]
	if !exists {
		return nil // No rules for this field
	}

	errors := &ValidationErrors{}

	for _, rule := range rules {
		var err error

		switch rule.RuleType {
		case "required":
			err = ValidateRequired(value)

		case "email":
			str, ok := value.(string)
			if !ok {
				errors.Add(field, "field must be a string for email validation")
				continue
			}
			err = ValidateEmail(str)

		case "length":
			str, ok := value.(string)
			if !ok {
				errors.Add(field, "field must be a string for length validation")
				continue
			}
			min, max := 0, 0
			if rule.Min != nil {
				min = rule.Min.(int)
			}
			if rule.Max != nil {
				max = rule.Max.(int)
			}
			err = ValidateLength(str, min, max)

		case "range":
			var num float64
			switch val := value.(type) {
			case float64:
				num = val
			case int:
				num = float64(val)
			case int64:
				num = float64(val)
			default:
				errors.Add(field, "field must be a number for range validation")
				continue
			}
			min, max := 0.0, 0.0
			if rule.Min != nil {
				switch m := rule.Min.(type) {
				case float64:
					min = m
				case int:
					min = float64(m)
				}
			}
			if rule.Max != nil {
				switch m := rule.Max.(type) {
				case float64:
					max = m
				case int:
					max = float64(m)
				}
			}
			err = ValidateRange(num, min, max)

		case "pattern":
			str, ok := value.(string)
			if !ok {
				errors.Add(field, "field must be a string for pattern validation")
				continue
			}
			err = ValidatePattern(str, rule.Pattern)

		case "custom":
			if rule.CustomFunc != nil {
				err = rule.CustomFunc(value)
			}
		}

		if err != nil {
			if validationErr, ok := err.(*ValidationError); ok {
				errors.Add(field, validationErr.Message)
			} else {
				errors.Add(field, err.Error())
			}
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ClearRules clears all validation rules
func (v *Validator) ClearRules() {
	v.rules = make(map[string][]ValidationRule)
}

// GetRules returns all validation rules for a field
func (v *Validator) GetRules(field string) []ValidationRule {
	return v.rules[field]
}

// HasRules checks if there are any rules for a field
func (v *Validator) HasRules(field string) bool {
	return len(v.rules[field]) > 0
}
