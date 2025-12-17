package vm

// Value represents a runtime value
type Value interface {
	Type() string
}

// NullValue represents a null value
type NullValue struct{}

func (v NullValue) Type() string { return "null" }

// IntValue represents an integer
type IntValue struct {
	Val int64
}

func (v IntValue) Type() string { return "int" }

// FloatValue represents a floating-point number
type FloatValue struct {
	Val float64
}

func (v FloatValue) Type() string { return "float" }

// StringValue represents a string
type StringValue struct {
	Val string
}

func (v StringValue) Type() string { return "string" }

// BoolValue represents a boolean
type BoolValue struct {
	Val bool
}

func (v BoolValue) Type() string { return "bool" }

// ArrayValue represents an array
type ArrayValue struct {
	Val []Value
}

func (v ArrayValue) Type() string { return "array" }

// ObjectValue represents an object (map)
type ObjectValue struct {
	Val map[string]Value
}

func (v ObjectValue) Type() string { return "object" }
