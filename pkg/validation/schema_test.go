package validation

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
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
