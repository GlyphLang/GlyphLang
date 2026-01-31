package validation

import (
	"github.com/glyphlang/glyph/pkg/interpreter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaFromTypeDef_Required(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "User",
		Fields: []interpreter.Field{
			{Name: "name", TypeAnnotation: interpreter.NamedType{Name: "str"}, Required: true},
			{Name: "bio", TypeAnnotation: interpreter.NamedType{Name: "str"}, Required: false},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"bio": "hello"})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"name": "John"})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_MinLen(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "name",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "minLen", Params: []interface{}{int64(3)}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"name": "ab"})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"name": "abc"})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_MaxLen(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "code",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "maxLen", Params: []interface{}{int64(5)}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"code": "toolong"})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"code": "ok"})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_Email(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "email",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "email"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"email": "notanemail"})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"email": "user@example.com"})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_Pattern(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "code",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "pattern", Params: []interface{}{"^[A-Z]{3}$"}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"code": "abc"})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"code": "ABC"})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_MinMax(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:     "age",
				Required: true,
				Annotations: []interpreter.FieldAnnotation{
					{Name: "min", Params: []interface{}{int64(0)}},
					{Name: "max", Params: []interface{}{int64(150)}},
				},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"age": float64(-1)})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"age": float64(200)})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"age": float64(25)})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_Range(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "score",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "range", Params: []interface{}{int64(1), int64(100)}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"score": float64(0)})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"score": float64(50)})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_OneOf(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "role",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "oneOf", Params: []interface{}{[]string{"admin", "user", "guest"}}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"role": "superadmin"})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"role": "admin"})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_UUID(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "id",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "uuid"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"id": "not-a-uuid"})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"id": "550e8400-e29b-41d4-a716-446655440000"})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_URL(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "website",
				Annotations: []interpreter.FieldAnnotation{{Name: "url"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"website": "not-a-url"})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"website": "https://example.com"})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_Positive(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "amount",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "positive"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"amount": float64(-5)})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"amount": float64(0)})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"amount": float64(10)})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_MinItems(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "tags",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "minItems", Params: []interface{}{int64(1)}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"tags": []interface{}{}})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"tags": []interface{}{"go"}})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_Unique(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "ids",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "unique"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"ids": []interface{}{1, 2, 1}})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"ids": []interface{}{1, 2, 3}})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_MultipleAnnotations(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "CreateUser",
		Fields: []interpreter.Field{
			{
				Name:     "name",
				Required: true,
				Annotations: []interpreter.FieldAnnotation{
					{Name: "minLen", Params: []interface{}{int64(2)}},
					{Name: "maxLen", Params: []interface{}{int64(100)}},
				},
			},
			{
				Name:     "email",
				Required: true,
				Annotations: []interpreter.FieldAnnotation{
					{Name: "email"},
				},
			},
			{
				Name: "age",
				Annotations: []interpreter.FieldAnnotation{
					{Name: "min", Params: []interface{}{int64(0)}},
					{Name: "max", Params: []interface{}{int64(150)}},
				},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	// Valid data
	err := v.Validate(map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
		"age":   float64(30),
	})
	assert.NoError(t, err)

	// Invalid: name too short and bad email
	err = v.Validate(map[string]interface{}{
		"name":  "J",
		"email": "invalid",
		"age":   float64(30),
	})
	assert.Error(t, err)
	verrs, ok := err.(*ValidationErrors)
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(verrs.Errors), 2)
}

func TestToResult(t *testing.T) {
	errs := &ValidationErrors{}
	errs.Add("name", "too short")
	errs.Add("email", "invalid format")

	result := ToResult(errs)
	assert.Equal(t, "Validation failed", result.Error)
	assert.Equal(t, "too short", result.Fields["name"])
	assert.Equal(t, "invalid format", result.Fields["email"])
}

func TestSchemaFromTypeDef_NoAnnotations(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Simple",
		Fields: []interpreter.Field{
			{Name: "name", Required: true},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"name": "anything"})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_NotEmpty(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "title",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "notEmpty"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"title": "   "})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"title": "Hello"})
	assert.NoError(t, err)
}

// --- Tests for toInt helper function ---
func TestToInt(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    int
		wantOk  bool
	}{
		{
			name:   "int64 value",
			input:  int64(42),
			want:   42,
			wantOk: true,
		},
		{
			name:   "int value",
			input:  int(99),
			want:   99,
			wantOk: true,
		},
		{
			name:   "float64 whole number",
			input:  float64(7),
			want:   7,
			wantOk: true,
		},
		{
			name:   "float64 non-whole number",
			input:  float64(3.14),
			want:   0,
			wantOk: false,
		},
		{
			name:   "string value returns false",
			input:  "hello",
			want:   0,
			wantOk: false,
		},
		{
			name:   "bool value returns false",
			input:  true,
			want:   0,
			wantOk: false,
		},
		{
			name:   "nil value returns false",
			input:  nil,
			want:   0,
			wantOk: false,
		},
		{
			name:   "int64 zero",
			input:  int64(0),
			want:   0,
			wantOk: true,
		},
		{
			name:   "int64 negative",
			input:  int64(-10),
			want:   -10,
			wantOk: true,
		},
		{
			name:   "float64 zero",
			input:  float64(0),
			want:   0,
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toInt(tt.input)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- Tests for toFloat helper function ---
func TestToFloat(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    float64
		wantOk  bool
	}{
		{
			name:   "float64 value",
			input:  float64(3.14),
			want:   3.14,
			wantOk: true,
		},
		{
			name:   "int64 value",
			input:  int64(42),
			want:   42.0,
			wantOk: true,
		},
		{
			name:   "int value",
			input:  int(99),
			want:   99.0,
			wantOk: true,
		},
		{
			name:   "string value returns false",
			input:  "hello",
			want:   0,
			wantOk: false,
		},
		{
			name:   "bool value returns false",
			input:  true,
			want:   0,
			wantOk: false,
		},
		{
			name:   "nil value returns false",
			input:  nil,
			want:   0,
			wantOk: false,
		},
		{
			name:   "float64 negative",
			input:  float64(-2.5),
			want:   -2.5,
			wantOk: true,
		},
		{
			name:   "int zero",
			input:  int(0),
			want:   0.0,
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toFloat(tt.input)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- Tests for toFloat64Value helper function ---
func TestToFloat64Value(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    float64
		wantOk  bool
	}{
		{
			name:   "float64 value",
			input:  float64(1.5),
			want:   1.5,
			wantOk: true,
		},
		{
			name:   "int64 value",
			input:  int64(10),
			want:   10.0,
			wantOk: true,
		},
		{
			name:   "int value",
			input:  int(25),
			want:   25.0,
			wantOk: true,
		},
		{
			name:   "int32 value",
			input:  int32(8),
			want:   8.0,
			wantOk: true,
		},
		{
			name:   "string value returns false",
			input:  "not a number",
			want:   0,
			wantOk: false,
		},
		{
			name:   "bool value returns false",
			input:  false,
			want:   0,
			wantOk: false,
		},
		{
			name:   "nil value returns false",
			input:  nil,
			want:   0,
			wantOk: false,
		},
		{
			name:   "int32 negative",
			input:  int32(-5),
			want:   -5.0,
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toFloat64Value(tt.input)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- Tests for paramInt helper function ---
func TestParamInt(t *testing.T) {
	tests := []struct {
		name    string
		params  []interface{}
		idx     int
		want    int
		wantOk  bool
	}{
		{
			name:   "valid int64 at index 0",
			params: []interface{}{int64(5)},
			idx:    0,
			want:   5,
			wantOk: true,
		},
		{
			name:   "valid int at index 0",
			params: []interface{}{int(10)},
			idx:    0,
			want:   10,
			wantOk: true,
		},
		{
			name:   "valid float64 whole number at index 0",
			params: []interface{}{float64(3)},
			idx:    0,
			want:   3,
			wantOk: true,
		},
		{
			name:   "index out of bounds",
			params: []interface{}{},
			idx:    0,
			want:   0,
			wantOk: false,
		},
		{
			name:   "index beyond length",
			params: []interface{}{int64(1)},
			idx:    5,
			want:   0,
			wantOk: false,
		},
		{
			name:   "non-numeric param at index",
			params: []interface{}{"hello"},
			idx:    0,
			want:   0,
			wantOk: false,
		},
		{
			name:   "float64 non-whole number",
			params: []interface{}{float64(2.7)},
			idx:    0,
			want:   0,
			wantOk: false,
		},
		{
			name:   "second param at index 1",
			params: []interface{}{int64(1), int64(20)},
			idx:    1,
			want:   20,
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := paramInt(tt.params, tt.idx)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- Tests for paramFloat helper function ---
func TestParamFloat(t *testing.T) {
	tests := []struct {
		name    string
		params  []interface{}
		idx     int
		want    float64
		wantOk  bool
	}{
		{
			name:   "valid float64 at index 0",
			params: []interface{}{float64(3.14)},
			idx:    0,
			want:   3.14,
			wantOk: true,
		},
		{
			name:   "valid int64 at index 0",
			params: []interface{}{int64(42)},
			idx:    0,
			want:   42.0,
			wantOk: true,
		},
		{
			name:   "valid int at index 0",
			params: []interface{}{int(7)},
			idx:    0,
			want:   7.0,
			wantOk: true,
		},
		{
			name:   "index out of bounds",
			params: []interface{}{},
			idx:    0,
			want:   0,
			wantOk: false,
		},
		{
			name:   "index beyond length",
			params: []interface{}{float64(1.0)},
			idx:    3,
			want:   0,
			wantOk: false,
		},
		{
			name:   "non-numeric param at index",
			params: []interface{}{"text"},
			idx:    0,
			want:   0,
			wantOk: false,
		},
		{
			name:   "second param at index 1",
			params: []interface{}{float64(1.0), float64(99.9)},
			idx:    1,
			want:   99.9,
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := paramFloat(tt.params, tt.idx)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- Tests for paramString helper function ---
func TestParamString(t *testing.T) {
	tests := []struct {
		name    string
		params  []interface{}
		idx     int
		want    string
		wantOk  bool
	}{
		{
			name:   "valid string at index 0",
			params: []interface{}{"hello"},
			idx:    0,
			want:   "hello",
			wantOk: true,
		},
		{
			name:   "index out of bounds",
			params: []interface{}{},
			idx:    0,
			want:   "",
			wantOk: false,
		},
		{
			name:   "non-string param at index",
			params: []interface{}{int64(42)},
			idx:    0,
			want:   "",
			wantOk: false,
		},
		{
			name:   "index beyond length",
			params: []interface{}{"a"},
			idx:    5,
			want:   "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := paramString(tt.params, tt.idx)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- Tests for paramStringSlice helper function ---
func TestParamStringSlice(t *testing.T) {
	tests := []struct {
		name    string
		params  []interface{}
		idx     int
		want    []string
		wantOk  bool
	}{
		{
			name:   "valid string slice at index 0",
			params: []interface{}{[]string{"a", "b", "c"}},
			idx:    0,
			want:   []string{"a", "b", "c"},
			wantOk: true,
		},
		{
			name:   "index out of bounds",
			params: []interface{}{},
			idx:    0,
			want:   nil,
			wantOk: false,
		},
		{
			name:   "non-slice param at index",
			params: []interface{}{"not a slice"},
			idx:    0,
			want:   nil,
			wantOk: false,
		},
		{
			name:   "index beyond length",
			params: []interface{}{[]string{"x"}},
			idx:    3,
			want:   nil,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := paramStringSlice(tt.params, tt.idx)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- Tests for addAnnotationRule: len annotation ---
func TestSchemaFromTypeDef_Len(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "code",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "len", Params: []interface{}{int64(5)}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	// Exactly 5 chars should pass
	err := v.Validate(map[string]interface{}{"code": "abcde"})
	assert.NoError(t, err)

	// Too short
	err = v.Validate(map[string]interface{}{"code": "abc"})
	assert.Error(t, err)

	// Too long
	err = v.Validate(map[string]interface{}{"code": "abcdefgh"})
	assert.Error(t, err)
}

// --- Tests for addAnnotationRule: negative annotation ---
func TestSchemaFromTypeDef_Negative(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "temp",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "negative"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"temp": float64(-5)})
	assert.NoError(t, err)

	err = v.Validate(map[string]interface{}{"temp": float64(0)})
	assert.Error(t, err)

	err = v.Validate(map[string]interface{}{"temp": float64(5)})
	assert.Error(t, err)

	// Non-numeric value should error
	err = v.Validate(map[string]interface{}{"temp": "cold"})
	assert.Error(t, err)
}

// --- Tests for addAnnotationRule: integer annotation ---
func TestSchemaFromTypeDef_Integer(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "count",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "integer"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	// int value passes
	err := v.Validate(map[string]interface{}{"count": int(5)})
	assert.NoError(t, err)

	// int64 value passes
	err = v.Validate(map[string]interface{}{"count": int64(10)})
	assert.NoError(t, err)

	// int32 value passes
	err = v.Validate(map[string]interface{}{"count": int32(3)})
	assert.NoError(t, err)

	// float64 whole number passes
	err = v.Validate(map[string]interface{}{"count": float64(7)})
	assert.NoError(t, err)

	// float64 non-whole number fails
	err = v.Validate(map[string]interface{}{"count": float64(3.14)})
	assert.Error(t, err)

	// string fails
	err = v.Validate(map[string]interface{}{"count": "five"})
	assert.Error(t, err)
}

// --- Tests for addAnnotationRule: maxItems annotation ---
func TestSchemaFromTypeDef_MaxItems(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "items",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "maxItems", Params: []interface{}{int64(3)}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	// Within limit
	err := v.Validate(map[string]interface{}{"items": []interface{}{"a", "b"}})
	assert.NoError(t, err)

	// At limit
	err = v.Validate(map[string]interface{}{"items": []interface{}{"a", "b", "c"}})
	assert.NoError(t, err)

	// Over limit
	err = v.Validate(map[string]interface{}{"items": []interface{}{"a", "b", "c", "d"}})
	assert.Error(t, err)

	// Non-array value
	err = v.Validate(map[string]interface{}{"items": "not an array"})
	assert.Error(t, err)
}

// --- Tests for addAnnotationRule: invalid pattern ---
func TestSchemaFromTypeDef_InvalidPattern(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "code",
				Required:    true,
				Annotations: []interpreter.FieldAnnotation{{Name: "pattern", Params: []interface{}{"[invalid"}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	// Invalid pattern should be silently ignored; no pattern rule added
	err := v.Validate(map[string]interface{}{"code": "anything"})
	assert.NoError(t, err)
}

// --- Tests for annotation rules with missing params ---
func TestSchemaFromTypeDef_MissingParams(t *testing.T) {
	// minLen with no params should not add a rule
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "name",
				Annotations: []interpreter.FieldAnnotation{{Name: "minLen", Params: []interface{}{}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)
	err := v.Validate(map[string]interface{}{"name": ""})
	assert.NoError(t, err) // No rule applied, so no error

	// maxLen with no params
	td2 := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "code",
				Annotations: []interpreter.FieldAnnotation{{Name: "maxLen", Params: []interface{}{}}},
			},
		},
	}

	v2 := SchemaFromTypeDef(td2)
	err = v2.Validate(map[string]interface{}{"code": "anything"})
	assert.NoError(t, err)

	// range with only one param
	td3 := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "score",
				Annotations: []interpreter.FieldAnnotation{{Name: "range", Params: []interface{}{int64(1)}}},
			},
		},
	}

	v3 := SchemaFromTypeDef(td3)
	err = v3.Validate(map[string]interface{}{"score": float64(500)})
	assert.NoError(t, err) // range needs both params to add rule

	// pattern with no params
	td4 := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "pat",
				Annotations: []interpreter.FieldAnnotation{{Name: "pattern", Params: []interface{}{}}},
			},
		},
	}

	v4 := SchemaFromTypeDef(td4)
	err = v4.Validate(map[string]interface{}{"pat": "anything"})
	assert.NoError(t, err)

	// oneOf with no params
	td5 := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "role",
				Annotations: []interpreter.FieldAnnotation{{Name: "oneOf", Params: []interface{}{}}},
			},
		},
	}

	v5 := SchemaFromTypeDef(td5)
	err = v5.Validate(map[string]interface{}{"role": "anything"})
	assert.NoError(t, err)
}

// --- Tests for annotation custom rules with non-numeric values ---
func TestSchemaFromTypeDef_MinWithNonNumeric(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "value",
				Annotations: []interpreter.FieldAnnotation{{Name: "min", Params: []interface{}{float64(5)}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	// Non-numeric value triggers "must be a number" error
	err := v.Validate(map[string]interface{}{"value": "not a number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a number")
}

func TestSchemaFromTypeDef_MaxWithNonNumeric(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "value",
				Annotations: []interpreter.FieldAnnotation{{Name: "max", Params: []interface{}{float64(100)}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	// Non-numeric value triggers "must be a number" error
	err := v.Validate(map[string]interface{}{"value": "not a number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a number")
}

func TestSchemaFromTypeDef_PositiveWithNonNumeric(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "amount",
				Annotations: []interpreter.FieldAnnotation{{Name: "positive"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"amount": "text"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a number")
}

// --- Tests for annotation rules with int types for toFloat64Value ---
func TestSchemaFromTypeDef_PositiveWithIntTypes(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "amount",
				Annotations: []interpreter.FieldAnnotation{{Name: "positive"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	// int value
	err := v.Validate(map[string]interface{}{"amount": int(5)})
	assert.NoError(t, err)

	// int64 value
	err = v.Validate(map[string]interface{}{"amount": int64(10)})
	assert.NoError(t, err)

	// int32 value
	err = v.Validate(map[string]interface{}{"amount": int32(3)})
	assert.NoError(t, err)

	// Negative int
	err = v.Validate(map[string]interface{}{"amount": int(-1)})
	assert.Error(t, err)
}

func TestSchemaFromTypeDef_NegativeWithIntTypes(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "temp",
				Annotations: []interpreter.FieldAnnotation{{Name: "negative"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	// Negative int value
	err := v.Validate(map[string]interface{}{"temp": int(-5)})
	assert.NoError(t, err)

	// Positive int value
	err = v.Validate(map[string]interface{}{"temp": int(5)})
	assert.Error(t, err)

	// int64 negative
	err = v.Validate(map[string]interface{}{"temp": int64(-3)})
	assert.NoError(t, err)

	// int32 negative
	err = v.Validate(map[string]interface{}{"temp": int32(-1)})
	assert.NoError(t, err)
}

// --- Tests for URL annotation with non-string value ---
func TestSchemaFromTypeDef_URLNonString(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "website",
				Annotations: []interpreter.FieldAnnotation{{Name: "url"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"website": 12345})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")
}

// --- Tests for UUID annotation with non-string value ---
func TestSchemaFromTypeDef_UUIDNonString(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "id",
				Annotations: []interpreter.FieldAnnotation{{Name: "uuid"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"id": 12345})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")
}

// --- Tests for oneOf annotation with non-string value ---
func TestSchemaFromTypeDef_OneOfNonString(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "role",
				Annotations: []interpreter.FieldAnnotation{{Name: "oneOf", Params: []interface{}{[]string{"admin", "user"}}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"role": 42})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")
}

// --- Tests for notEmpty annotation with non-string value ---
func TestSchemaFromTypeDef_NotEmptyNonString(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "title",
				Annotations: []interpreter.FieldAnnotation{{Name: "notEmpty"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	// Non-string value is allowed (returns nil from notEmpty for non-strings)
	err := v.Validate(map[string]interface{}{"title": 42})
	assert.NoError(t, err)
}

// --- Tests for minItems/maxItems annotation with non-array value ---
func TestSchemaFromTypeDef_MinItemsNonArray(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "tags",
				Annotations: []interpreter.FieldAnnotation{{Name: "minItems", Params: []interface{}{int64(1)}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"tags": "not an array"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an array")
}

// --- Tests for unique annotation with non-array value ---
func TestSchemaFromTypeDef_UniqueNonArray(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "ids",
				Annotations: []interpreter.FieldAnnotation{{Name: "unique"}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	err := v.Validate(map[string]interface{}{"ids": "not an array"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an array")
}

// --- Test ToResult with non-ValidationErrors error ---
func TestToResult_NonValidationErrors(t *testing.T) {
	err := &ValidationError{Message: "some plain error"}
	result := ToResult(err)
	assert.Equal(t, "validation error: some plain error", result.Error)
	assert.NotNil(t, result.Fields)
	assert.Empty(t, result.Fields)
}

// --- Test for min/max annotation with int values exercising toFloat64Value ---
func TestSchemaFromTypeDef_MinWithIntValue(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "value",
				Annotations: []interpreter.FieldAnnotation{{Name: "min", Params: []interface{}{float64(10)}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	// int value that passes
	err := v.Validate(map[string]interface{}{"value": int(20)})
	assert.NoError(t, err)

	// int value that fails
	err = v.Validate(map[string]interface{}{"value": int(5)})
	assert.Error(t, err)

	// int64 value
	err = v.Validate(map[string]interface{}{"value": int64(15)})
	assert.NoError(t, err)

	// int32 value
	err = v.Validate(map[string]interface{}{"value": int32(12)})
	assert.NoError(t, err)
}

func TestSchemaFromTypeDef_MaxWithIntValue(t *testing.T) {
	td := &interpreter.TypeDef{
		Name: "Input",
		Fields: []interpreter.Field{
			{
				Name:        "value",
				Annotations: []interpreter.FieldAnnotation{{Name: "max", Params: []interface{}{float64(100)}}},
			},
		},
	}

	v := SchemaFromTypeDef(td)

	// int value that passes
	err := v.Validate(map[string]interface{}{"value": int(50)})
	assert.NoError(t, err)

	// int value that fails
	err = v.Validate(map[string]interface{}{"value": int(200)})
	assert.Error(t, err)

	// int64 value
	err = v.Validate(map[string]interface{}{"value": int64(80)})
	assert.NoError(t, err)

	// int32 value
	err = v.Validate(map[string]interface{}{"value": int32(99)})
	assert.NoError(t, err)
}
