package interpreter

import (
	"fmt"
)

// BindingSource identifies the origin of a variable binding in an Environment.
// It is used to produce better diagnostic messages when user code collides
// with an implicitly bound variable (for example, a path parameter extracted
// from a route pattern).
type BindingSource int

const (
	// BindingUser indicates a variable declared by user code (the default).
	BindingUser BindingSource = iota
	// BindingPathParam indicates a variable bound from a route path parameter.
	BindingPathParam
	// BindingQueryParam indicates a variable bound from a route query parameter.
	BindingQueryParam
)

// binding stores a variable's value alongside the source of its binding.
// Keeping both in a single map entry ensures the value and its source cannot
// drift out of sync.
type binding struct {
	value  interface{}
	source BindingSource
}

// Environment manages variable scopes and bindings.
// Environment is not safe for concurrent use; callers must synchronize access
// if an environment is shared across goroutines.
type Environment struct {
	vars   map[string]binding
	parent *Environment
}

// NewEnvironment creates a new environment
func NewEnvironment() *Environment {
	return &Environment{
		vars:   make(map[string]binding),
		parent: nil,
	}
}

// NewChildEnvironment creates a child environment with a parent scope
func NewChildEnvironment(parent *Environment) *Environment {
	return &Environment{
		vars:   make(map[string]binding),
		parent: parent,
	}
}

// Define adds a new variable to the current environment as a user-declared
// binding. For bindings that originate from the runtime (e.g. path or query
// parameters), use DefineWithSource so diagnostics can report the origin.
func (e *Environment) Define(name string, value interface{}) {
	e.DefineWithSource(name, value, BindingUser)
}

// DefineWithSource adds a new variable to the current environment and records
// where the binding originated. The source is used by diagnostics to produce
// clearer error messages when user code redeclares a variable that was
// implicitly bound (for example a path or query parameter extracted from a
// route pattern).
func (e *Environment) DefineWithSource(name string, value interface{}, source BindingSource) {
	e.vars[name] = binding{value: value, source: source}
}

// LocalSource returns the BindingSource for a variable defined in the current
// scope. If the variable is not defined locally, it returns BindingUser and
// false.
func (e *Environment) LocalSource(name string) (BindingSource, bool) {
	if b, ok := e.vars[name]; ok {
		return b.source, true
	}
	return BindingUser, false
}

// Get retrieves a variable value from the environment or parent scopes
func (e *Environment) Get(name string) (interface{}, error) {
	if b, ok := e.vars[name]; ok {
		return b.value, nil
	}

	if e.parent != nil {
		return e.parent.Get(name)
	}

	return nil, fmt.Errorf("undefined variable: %s", name)
}

// Set updates a variable value in the environment or parent scopes.
// The binding source is preserved.
func (e *Environment) Set(name string, value interface{}) error {
	if b, ok := e.vars[name]; ok {
		b.value = value
		e.vars[name] = b
		return nil
	}

	if e.parent != nil {
		return e.parent.Set(name, value)
	}

	return fmt.Errorf("undefined variable: %s", name)
}

// Has checks if a variable exists in the environment or parent scopes
func (e *Environment) Has(name string) bool {
	if _, ok := e.vars[name]; ok {
		return true
	}

	if e.parent != nil {
		return e.parent.Has(name)
	}

	return false
}

// HasLocal checks if a variable exists in the current scope only (no parent lookup)
func (e *Environment) HasLocal(name string) bool {
	_, ok := e.vars[name]
	return ok
}

// GetAll returns all variables in this environment and all parent environments.
// If a variable exists in multiple scopes, the closest scope wins.
func (e *Environment) GetAll() map[string]interface{} {
	result := make(map[string]interface{})

	// Start with parent variables (if any)
	if e.parent != nil {
		for name, val := range e.parent.GetAll() {
			result[name] = val
		}
	}

	// Override with current scope variables
	for name, b := range e.vars {
		result[name] = b.value
	}

	return result
}

// GetLocal returns all variables in the current scope only (no parent lookup).
func (e *Environment) GetLocal() map[string]interface{} {
	result := make(map[string]interface{})
	for name, b := range e.vars {
		result[name] = b.value
	}
	return result
}
