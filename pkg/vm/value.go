package vm

import "encoding/json"

// Value represents a runtime value
type Value interface {
	Type() string
}

// NullValue represents a null value
type NullValue struct{}

func (v NullValue) Type() string { return "null" }

func (v NullValue) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

// IntValue represents an integer
type IntValue struct {
	Val int64
}

func (v IntValue) Type() string { return "int" }

func (v IntValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Val)
}

// FloatValue represents a floating-point number
type FloatValue struct {
	Val float64
}

func (v FloatValue) Type() string { return "float" }

func (v FloatValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Val)
}

// StringValue represents a string
type StringValue struct {
	Val string
}

func (v StringValue) Type() string { return "string" }

func (v StringValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Val)
}

// BoolValue represents a boolean
type BoolValue struct {
	Val bool
}

func (v BoolValue) Type() string { return "bool" }

func (v BoolValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Val)
}

// ArrayValue represents an array
type ArrayValue struct {
	Val []Value
}

func (v ArrayValue) Type() string { return "array" }

func (v ArrayValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Val)
}

// ObjectValue represents an object (map)
type ObjectValue struct {
	Val map[string]Value
}

func (v ObjectValue) Type() string { return "object" }

func (v ObjectValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Val)
}
