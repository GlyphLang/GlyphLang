package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test ValidateEmail function
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with dots",
			email:   "first.last@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with numbers",
			email:   "user123@example123.com",
			wantErr: false,
		},
		{
			name:    "valid email with hyphen in domain",
			email:   "user@my-domain.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
			errMsg:  "email cannot be empty",
		},
		{
			name:    "missing @ symbol",
			email:   "userexample.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "missing domain",
			email:   "user@",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "missing local part",
			email:   "@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "multiple @ symbols",
			email:   "user@@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "no dot in domain",
			email:   "user@example",
			wantErr: true,
			errMsg:  "email domain must contain at least one dot",
		},
		{
			name:    "spaces in email",
			email:   "user name@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "email too long (total)",
			email:   "verylonglocalpartverylonglocalpartverylonglocalpartverylonglocalpart@verylongdomainverylongdomainverylongdomainverylongdomainverylongdomainverylongdomainverylongdomainverylongdomainverylongdomainverylongdomainverylongdomainverylongdomainverylongdomain.com",
			wantErr: true,
			errMsg:  "invalid email format", // Fails regex before length check
		},
		{
			name:    "local part too long",
			email:   "verylonglocalpartverylonglocalpartverylonglocalpartverylonglocalpartverylonglocalpart@example.com",
			wantErr: true,
			errMsg:  "email local part too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test ValidateLength function
func TestValidateLength(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		min     int
		max     int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid length within range",
			str:     "hello",
			min:     3,
			max:     10,
			wantErr: false,
		},
		{
			name:    "valid length at minimum",
			str:     "abc",
			min:     3,
			max:     10,
			wantErr: false,
		},
		{
			name:    "valid length at maximum",
			str:     "1234567890",
			min:     3,
			max:     10,
			wantErr: false,
		},
		{
			name:    "no minimum constraint",
			str:     "hi",
			min:     0,
			max:     10,
			wantErr: false,
		},
		{
			name:    "no maximum constraint",
			str:     "this is a very long string",
			min:     5,
			max:     0,
			wantErr: false,
		},
		{
			name:    "string too short",
			str:     "ab",
			min:     5,
			max:     10,
			wantErr: true,
			errMsg:  "length must be at least 5 characters (got 2)",
		},
		{
			name:    "string too long",
			str:     "this is too long",
			min:     3,
			max:     10,
			wantErr: true,
			errMsg:  "length must be at most 10 characters (got 16)",
		},
		{
			name:    "empty string with minimum",
			str:     "",
			min:     1,
			max:     10,
			wantErr: true,
			errMsg:  "length must be at least 1 characters (got 0)",
		},
		{
			name:    "empty string no minimum",
			str:     "",
			min:     0,
			max:     10,
			wantErr: false,
		},
		{
			name:    "exact match min and max",
			str:     "12345",
			min:     5,
			max:     5,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLength(tt.str, tt.min, tt.max)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, "validation error: "+tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test ValidateRange function
func TestValidateRange(t *testing.T) {
	tests := []struct {
		name    string
		value   float64
		min     float64
		max     float64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid value within range",
			value:   5.5,
			min:     1.0,
			max:     10.0,
			wantErr: false,
		},
		{
			name:    "valid value at minimum",
			value:   1.0,
			min:     1.0,
			max:     10.0,
			wantErr: false,
		},
		{
			name:    "valid value at maximum",
			value:   10.0,
			min:     1.0,
			max:     10.0,
			wantErr: false,
		},
		{
			name:    "no minimum constraint",
			value:   -100.5,
			min:     0,
			max:     10.0,
			wantErr: false,
		},
		{
			name:    "no maximum constraint",
			value:   1000.0,
			min:     1.0,
			max:     0,
			wantErr: false,
		},
		{
			name:    "value below minimum",
			value:   0.5,
			min:     1.0,
			max:     10.0,
			wantErr: true,
			errMsg:  "value must be at least 1 (got 0.5)",
		},
		{
			name:    "value above maximum",
			value:   15.0,
			min:     1.0,
			max:     10.0,
			wantErr: true,
			errMsg:  "value must be at most 10 (got 15)",
		},
		{
			name:    "negative value below minimum",
			value:   -5.0,
			min:     -3.0,
			max:     10.0,
			wantErr: true,
			errMsg:  "value must be at least -3 (got -5)",
		},
		{
			name:    "negative value within range",
			value:   -2.0,
			min:     -5.0,
			max:     5.0,
			wantErr: false,
		},
		{
			name:    "zero value",
			value:   0.0,
			min:     -10.0,
			max:     10.0,
			wantErr: false,
		},
		{
			name:    "integer values",
			value:   5.0,
			min:     1.0,
			max:     10.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRange(tt.value, tt.min, tt.max)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, "validation error: "+tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test ValidatePattern function
func TestValidatePattern(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		pattern string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid alphanumeric pattern",
			str:     "abc123",
			pattern: "^[a-z0-9]+$",
			wantErr: false,
		},
		{
			name:    "valid phone number pattern",
			str:     "123-456-7890",
			pattern: `^\d{3}-\d{3}-\d{4}$`,
			wantErr: false,
		},
		{
			name:    "valid URL pattern",
			str:     "https://example.com",
			pattern: `^https?://[a-z0-9.-]+\.[a-z]{2,}$`,
			wantErr: false,
		},
		{
			name:    "valid UUID pattern",
			str:     "550e8400-e29b-41d4-a716-446655440000",
			pattern: `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
			wantErr: false,
		},
		{
			name:    "pattern mismatch",
			str:     "abc",
			pattern: "^[0-9]+$",
			wantErr: true,
			errMsg:  "value does not match pattern",
		},
		{
			name:    "empty pattern",
			str:     "test",
			pattern: "",
			wantErr: true,
			errMsg:  "pattern cannot be empty",
		},
		{
			name:    "invalid regex pattern",
			str:     "test",
			pattern: "[a-z",
			wantErr: true,
			errMsg:  "invalid regex pattern",
		},
		{
			name:    "case sensitive pattern",
			str:     "ABC",
			pattern: "^[a-z]+$",
			wantErr: true,
			errMsg:  "value does not match pattern",
		},
		{
			name:    "partial match fails (needs anchors)",
			str:     "abc123def",
			pattern: "[0-9]+",
			wantErr: false, // Partial matches work without anchors
		},
		{
			name:    "whitespace pattern",
			str:     "hello world",
			pattern: `^[a-z\s]+$`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePattern(tt.str, tt.pattern)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test ValidateRequired function
func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid non-empty string",
			value:   "hello",
			wantErr: false,
		},
		{
			name:    "valid number",
			value:   42,
			wantErr: false,
		},
		{
			name:    "valid zero number",
			value:   0,
			wantErr: false,
		},
		{
			name:    "valid boolean true",
			value:   true,
			wantErr: false,
		},
		{
			name:    "valid boolean false",
			value:   false,
			wantErr: false,
		},
		{
			name:    "valid non-empty array",
			value:   []interface{}{"a", "b"},
			wantErr: false,
		},
		{
			name:    "valid non-empty map",
			value:   map[string]interface{}{"key": "value"},
			wantErr: false,
		},
		{
			name:    "nil value",
			value:   nil,
			wantErr: true,
			errMsg:  "value is required",
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
			errMsg:  "value cannot be empty",
		},
		{
			name:    "empty array",
			value:   []interface{}{},
			wantErr: true,
			errMsg:  "array cannot be empty",
		},
		{
			name:    "empty map",
			value:   map[string]interface{}{},
			wantErr: true,
			errMsg:  "object cannot be empty",
		},
		{
			name:    "whitespace only string",
			value:   "   ",
			wantErr: false, // Whitespace is considered non-empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, "validation error: "+tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Validator struct and methods
func TestValidator(t *testing.T) {
	t.Run("create new validator", func(t *testing.T) {
		v := NewValidator()
		require.NotNil(t, v)
		require.NotNil(t, v.rules)
	})

	t.Run("add email rule", func(t *testing.T) {
		v := NewValidator()
		v.AddEmailRule("email")

		assert.True(t, v.HasRules("email"))
		rules := v.GetRules("email")
		assert.Len(t, rules, 1)
		assert.Equal(t, "email", rules[0].RuleType)
	})

	t.Run("add length rule", func(t *testing.T) {
		v := NewValidator()
		v.AddLengthRule("username", 3, 20)

		rules := v.GetRules("username")
		assert.Len(t, rules, 1)
		assert.Equal(t, "length", rules[0].RuleType)
		assert.Equal(t, 3, rules[0].Min)
		assert.Equal(t, 20, rules[0].Max)
	})

	t.Run("add range rule", func(t *testing.T) {
		v := NewValidator()
		v.AddRangeRule("age", 18.0, 100.0)

		rules := v.GetRules("age")
		assert.Len(t, rules, 1)
		assert.Equal(t, "range", rules[0].RuleType)
		assert.Equal(t, 18.0, rules[0].Min)
		assert.Equal(t, 100.0, rules[0].Max)
	})

	t.Run("add pattern rule", func(t *testing.T) {
		v := NewValidator()
		v.AddPatternRule("zipcode", `^\d{5}$`)

		rules := v.GetRules("zipcode")
		assert.Len(t, rules, 1)
		assert.Equal(t, "pattern", rules[0].RuleType)
		assert.Equal(t, `^\d{5}$`, rules[0].Pattern)
	})

	t.Run("add required rule", func(t *testing.T) {
		v := NewValidator()
		v.AddRequiredRule("name")

		rules := v.GetRules("name")
		assert.Len(t, rules, 1)
		assert.Equal(t, "required", rules[0].RuleType)
	})

	t.Run("add custom rule", func(t *testing.T) {
		v := NewValidator()
		customFunc := func(value interface{}) error {
			return nil
		}
		v.AddCustomRule("custom", customFunc)

		rules := v.GetRules("custom")
		assert.Len(t, rules, 1)
		assert.Equal(t, "custom", rules[0].RuleType)
		assert.NotNil(t, rules[0].CustomFunc)
	})

	t.Run("multiple rules for same field", func(t *testing.T) {
		v := NewValidator()
		v.AddRequiredRule("email")
		v.AddEmailRule("email")

		rules := v.GetRules("email")
		assert.Len(t, rules, 2)
	})

	t.Run("clear rules", func(t *testing.T) {
		v := NewValidator()
		v.AddEmailRule("email")
		v.AddRequiredRule("name")

		v.ClearRules()

		assert.False(t, v.HasRules("email"))
		assert.False(t, v.HasRules("name"))
	})

	t.Run("has rules returns false for non-existent field", func(t *testing.T) {
		v := NewValidator()
		assert.False(t, v.HasRules("nonexistent"))
	})
}

// Test Validator.Validate method
func TestValidatorValidate(t *testing.T) {
	t.Run("valid data passes all rules", func(t *testing.T) {
		v := NewValidator()
		v.AddRequiredRule("email")
		v.AddEmailRule("email")
		v.AddRequiredRule("username")
		v.AddLengthRule("username", 3, 20)
		v.AddRangeRule("age", 18.0, 100.0)

		data := map[string]interface{}{
			"email":    "user@example.com",
			"username": "johndoe",
			"age":      25.0,
		}

		err := v.Validate(data)
		assert.NoError(t, err)
	})

	t.Run("required field missing", func(t *testing.T) {
		v := NewValidator()
		v.AddRequiredRule("email")

		data := map[string]interface{}{}

		err := v.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email")
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("invalid email format", func(t *testing.T) {
		v := NewValidator()
		v.AddEmailRule("email")

		data := map[string]interface{}{
			"email": "invalid-email",
		}

		err := v.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email")
	})

	t.Run("length validation fails", func(t *testing.T) {
		v := NewValidator()
		v.AddLengthRule("username", 5, 20)

		data := map[string]interface{}{
			"username": "ab",
		}

		err := v.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "username")
		assert.Contains(t, err.Error(), "at least 5")
	})

	t.Run("range validation fails", func(t *testing.T) {
		v := NewValidator()
		v.AddRangeRule("age", 18.0, 100.0)

		data := map[string]interface{}{
			"age": 15.0,
		}

		err := v.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "age")
		assert.Contains(t, err.Error(), "at least 18")
	})

	t.Run("pattern validation fails", func(t *testing.T) {
		v := NewValidator()
		v.AddPatternRule("zipcode", `^\d{5}$`)

		data := map[string]interface{}{
			"zipcode": "ABC12",
		}

		err := v.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "zipcode")
		assert.Contains(t, err.Error(), "does not match pattern")
	})

	t.Run("custom validation fails", func(t *testing.T) {
		v := NewValidator()
		v.AddCustomRule("password", func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return &ValidationError{Message: "password must be a string"}
			}
			if len(str) < 8 {
				return &ValidationError{Message: "password must be at least 8 characters"}
			}
			return nil
		})

		data := map[string]interface{}{
			"password": "short",
		}

		err := v.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password")
		assert.Contains(t, err.Error(), "at least 8")
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		v := NewValidator()
		v.AddRequiredRule("email")
		v.AddEmailRule("email")
		v.AddRequiredRule("username")
		v.AddLengthRule("username", 5, 20)

		data := map[string]interface{}{
			"email":    "invalid",
			"username": "ab",
		}

		err := v.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email")
		assert.Contains(t, err.Error(), "username")
	})

	t.Run("optional field not validated when missing", func(t *testing.T) {
		v := NewValidator()
		v.AddEmailRule("email") // Email rule without required

		data := map[string]interface{}{} // Email not provided

		err := v.Validate(data)
		assert.NoError(t, err) // Should pass because email is optional
	})

	t.Run("type mismatch for email", func(t *testing.T) {
		v := NewValidator()
		v.AddEmailRule("email")

		data := map[string]interface{}{
			"email": 123, // Not a string
		}

		err := v.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a string")
	})

	t.Run("type mismatch for range", func(t *testing.T) {
		v := NewValidator()
		v.AddRangeRule("age", 18.0, 100.0)

		data := map[string]interface{}{
			"age": "twenty", // Not a number
		}

		err := v.Validate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a number")
	})

	t.Run("integer values in range validation", func(t *testing.T) {
		v := NewValidator()
		v.AddRangeRule("count", 1.0, 100.0)

		data := map[string]interface{}{
			"count": 50, // int instead of float64
		}

		err := v.Validate(data)
		assert.NoError(t, err)
	})
}

// Test Validator.ValidateField method
func TestValidatorValidateField(t *testing.T) {
	t.Run("validate single field success", func(t *testing.T) {
		v := NewValidator()
		v.AddEmailRule("email")

		err := v.ValidateField("email", "user@example.com")
		assert.NoError(t, err)
	})

	t.Run("validate single field failure", func(t *testing.T) {
		v := NewValidator()
		v.AddEmailRule("email")

		err := v.ValidateField("email", "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email")
	})

	t.Run("validate field with no rules", func(t *testing.T) {
		v := NewValidator()

		err := v.ValidateField("unknown", "value")
		assert.NoError(t, err)
	})

	t.Run("validate required field", func(t *testing.T) {
		v := NewValidator()
		v.AddRequiredRule("name")

		err := v.ValidateField("name", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("validate length field", func(t *testing.T) {
		v := NewValidator()
		v.AddLengthRule("username", 5, 20)

		err := v.ValidateField("username", "abc")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 5")
	})

	t.Run("validate range field", func(t *testing.T) {
		v := NewValidator()
		v.AddRangeRule("score", 0.0, 100.0)

		err := v.ValidateField("score", 150.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at most 100")
	})

	t.Run("validate pattern field", func(t *testing.T) {
		v := NewValidator()
		v.AddPatternRule("code", `^[A-Z]{3}\d{3}$`)

		err := v.ValidateField("code", "ABC123")
		assert.NoError(t, err)

		err = v.ValidateField("code", "invalid")
		assert.Error(t, err)
	})

	t.Run("validate custom field", func(t *testing.T) {
		v := NewValidator()
		v.AddCustomRule("status", func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return &ValidationError{Message: "status must be a string"}
			}
			validStatuses := []string{"active", "inactive", "pending"}
			for _, valid := range validStatuses {
				if str == valid {
					return nil
				}
			}
			return &ValidationError{Message: "invalid status value"}
		})

		err := v.ValidateField("status", "active")
		assert.NoError(t, err)

		err = v.ValidateField("status", "unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})
}

// Test ValidationError
func TestValidationError(t *testing.T) {
	t.Run("error with field", func(t *testing.T) {
		err := &ValidationError{
			Field:   "email",
			Message: "invalid format",
		}

		expected := "validation error on field 'email': invalid format"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("error without field", func(t *testing.T) {
		err := &ValidationError{
			Message: "validation failed",
		}

		expected := "validation error: validation failed"
		assert.Equal(t, expected, err.Error())
	})
}

// Test ValidationErrors
func TestValidationErrors(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		errs := &ValidationErrors{}
		assert.False(t, errs.HasErrors())
		assert.Equal(t, "validation errors occurred", errs.Error())
	})

	t.Run("single error", func(t *testing.T) {
		errs := &ValidationErrors{}
		errs.Add("email", "invalid format")

		assert.True(t, errs.HasErrors())
		assert.Contains(t, errs.Error(), "email")
		assert.Contains(t, errs.Error(), "invalid format")
	})

	t.Run("multiple errors", func(t *testing.T) {
		errs := &ValidationErrors{}
		errs.Add("email", "invalid format")
		errs.Add("username", "too short")

		assert.True(t, errs.HasErrors())
		errMsg := errs.Error()
		assert.Contains(t, errMsg, "email")
		assert.Contains(t, errMsg, "username")
		assert.Contains(t, errMsg, "; ")
	})
}

// --- Additional ValidateField tests for coverage ---

func TestValidateField_LengthTypeMismatch(t *testing.T) {
	v := NewValidator()
	v.AddLengthRule("username", 3, 20)

	// Non-string value for length validation
	err := v.ValidateField("username", 12345)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string for length validation")
}

func TestValidateField_RangeTypeMismatch(t *testing.T) {
	v := NewValidator()
	v.AddRangeRule("score", 0.0, 100.0)

	// Non-numeric value for range validation
	err := v.ValidateField("score", "not a number")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a number")
}

func TestValidateField_RangeInt64(t *testing.T) {
	v := NewValidator()
	v.AddRangeRule("count", 1.0, 100.0)

	// int64 value in range
	err := v.ValidateField("count", int64(50))
	assert.NoError(t, err)

	// int64 value out of range
	err = v.ValidateField("count", int64(200))
	assert.Error(t, err)
}

func TestValidateField_PatternTypeMismatch(t *testing.T) {
	v := NewValidator()
	v.AddPatternRule("code", `^[A-Z]+$`)

	// Non-string value for pattern validation
	err := v.ValidateField("code", 12345)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string for pattern validation")
}

func TestValidateField_EmailTypeMismatch(t *testing.T) {
	v := NewValidator()
	v.AddEmailRule("email")

	// Non-string value for email validation
	err := v.ValidateField("email", 12345)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string for email validation")
}

func TestValidateField_CustomReturnsNonValidationError(t *testing.T) {
	v := NewValidator()
	v.AddCustomRule("field", func(value interface{}) error {
		return fmt.Errorf("generic error from custom rule")
	})

	err := v.ValidateField("field", "anything")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generic error from custom rule")
}

func TestValidateField_RequiredWithNil(t *testing.T) {
	v := NewValidator()
	v.AddRequiredRule("name")

	err := v.ValidateField("name", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestValidateField_MultipleRulesOneFailure(t *testing.T) {
	v := NewValidator()
	v.AddRequiredRule("email")
	v.AddEmailRule("email")

	// Passes required but fails email format
	err := v.ValidateField("email", "not-an-email")
	assert.Error(t, err)
}

// --- Additional Validate tests for coverage ---

func TestValidate_LengthTypeMismatch(t *testing.T) {
	v := NewValidator()
	v.AddLengthRule("username", 3, 20)

	data := map[string]interface{}{
		"username": 12345, // Not a string
	}

	err := v.Validate(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string for length validation")
}

func TestValidate_PatternTypeMismatch(t *testing.T) {
	v := NewValidator()
	v.AddPatternRule("code", `^[A-Z]+$`)

	data := map[string]interface{}{
		"code": 12345, // Not a string
	}

	err := v.Validate(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string for pattern validation")
}

func TestValidate_RangeInt64(t *testing.T) {
	v := NewValidator()
	v.AddRangeRule("count", 1.0, 100.0)

	data := map[string]interface{}{
		"count": int64(50),
	}

	err := v.Validate(data)
	assert.NoError(t, err)

	data2 := map[string]interface{}{
		"count": int64(200),
	}

	err = v.Validate(data2)
	assert.Error(t, err)
}

func TestValidate_RequiredFieldEmptyString(t *testing.T) {
	v := NewValidator()
	v.AddRequiredRule("name")

	data := map[string]interface{}{
		"name": "", // Exists but empty
	}

	err := v.Validate(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestValidate_RequiredFieldNilValue(t *testing.T) {
	v := NewValidator()
	v.AddRequiredRule("name")

	data := map[string]interface{}{
		"name": nil,
	}

	err := v.Validate(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestValidate_CustomReturnsNonValidationError(t *testing.T) {
	v := NewValidator()
	v.AddCustomRule("field", func(value interface{}) error {
		return fmt.Errorf("generic error")
	})

	data := map[string]interface{}{
		"field": "anything",
	}

	err := v.Validate(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generic error")
}

func TestValidate_OptionalPatternNotValidatedWhenMissing(t *testing.T) {
	v := NewValidator()
	v.AddPatternRule("code", `^[A-Z]+$`)

	data := map[string]interface{}{} // code not provided

	err := v.Validate(data)
	assert.NoError(t, err) // Optional fields not validated when missing
}

func TestValidate_OptionalLengthNotValidatedWhenMissing(t *testing.T) {
	v := NewValidator()
	v.AddLengthRule("username", 3, 20)

	data := map[string]interface{}{} // username not provided

	err := v.Validate(data)
	assert.NoError(t, err)
}

func TestValidate_OptionalRangeNotValidatedWhenMissing(t *testing.T) {
	v := NewValidator()
	v.AddRangeRule("score", 0.0, 100.0)

	data := map[string]interface{}{} // score not provided

	err := v.Validate(data)
	assert.NoError(t, err)
}

func TestValidate_OptionalCustomNotValidatedWhenMissing(t *testing.T) {
	v := NewValidator()
	v.AddCustomRule("field", func(value interface{}) error {
		return &ValidationError{Message: "custom error"}
	})

	data := map[string]interface{}{} // field not provided

	err := v.Validate(data)
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkValidateEmail(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateEmail("user@example.com")
	}
}

func BenchmarkValidateLength(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateLength("hello world", 5, 20)
	}
}

func BenchmarkValidateRange(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateRange(50.0, 0.0, 100.0)
	}
}

func BenchmarkValidatePattern(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidatePattern("ABC123", `^[A-Z]{3}\d{3}$`)
	}
}

func BenchmarkValidatorValidate(b *testing.B) {
	v := NewValidator()
	v.AddRequiredRule("email")
	v.AddEmailRule("email")
	v.AddLengthRule("username", 3, 20)
	v.AddRangeRule("age", 18.0, 100.0)

	data := map[string]interface{}{
		"email":    "user@example.com",
		"username": "johndoe",
		"age":      25.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate(data)
	}
}
