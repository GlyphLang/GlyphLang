package interpreter

import "fmt"

// ResultValue represents a Result[T, E] value with Ok and Err variants.
// It is the runtime representation of the Result type in GlyphLang.
type ResultValue struct {
	ok    bool
	value interface{} // Holds the Ok value or Err value
}

// NewOk creates a Result in the Ok state.
func NewOk(value interface{}) *ResultValue {
	return &ResultValue{ok: true, value: value}
}

// NewErr creates a Result in the Err state.
func NewErr(value interface{}) *ResultValue {
	return &ResultValue{ok: false, value: value}
}

// IsOk returns true if the Result is Ok.
func (r *ResultValue) IsOk() bool {
	return r.ok
}

// IsErr returns true if the Result is Err.
func (r *ResultValue) IsErr() bool {
	return !r.ok
}

// Unwrap returns the Ok value or panics if Err.
func (r *ResultValue) Unwrap() (interface{}, error) {
	if r.ok {
		return r.value, nil
	}
	return nil, fmt.Errorf("called unwrap() on an Err value: %v", r.value)
}

// UnwrapOr returns the Ok value, or the provided default if Err.
func (r *ResultValue) UnwrapOr(defaultVal interface{}) interface{} {
	if r.ok {
		return r.value
	}
	return defaultVal
}

// UnwrapErr returns the Err value or panics if Ok.
func (r *ResultValue) UnwrapErr() (interface{}, error) {
	if !r.ok {
		return r.value, nil
	}
	return nil, fmt.Errorf("called unwrapErr() on an Ok value: %v", r.value)
}

// OkValue returns the Ok value (nil if Err). Used for pattern matching.
func (r *ResultValue) OkValue() interface{} {
	if r.ok {
		return r.value
	}
	return nil
}

// ErrValue returns the Err value (nil if Ok). Used for pattern matching.
func (r *ResultValue) ErrValue() interface{} {
	if !r.ok {
		return r.value
	}
	return nil
}

// String returns a human-readable representation.
func (r *ResultValue) String() string {
	if r.ok {
		return fmt.Sprintf("Ok(%v)", r.value)
	}
	return fmt.Sprintf("Err(%v)", r.value)
}
