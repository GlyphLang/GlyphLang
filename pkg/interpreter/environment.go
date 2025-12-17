package interpreter

import "fmt"

// Environment manages variable scopes and bindings
type Environment struct {
	vars   map[string]interface{}
	parent *Environment
}

// NewEnvironment creates a new environment
func NewEnvironment() *Environment {
	return &Environment{
		vars:   make(map[string]interface{}),
		parent: nil,
	}
}

// NewChildEnvironment creates a child environment with a parent scope
func NewChildEnvironment(parent *Environment) *Environment {
	return &Environment{
		vars:   make(map[string]interface{}),
		parent: parent,
	}
}

// Define adds a new variable to the current environment
func (e *Environment) Define(name string, value interface{}) {
	e.vars[name] = value
}

// Get retrieves a variable value from the environment or parent scopes
func (e *Environment) Get(name string) (interface{}, error) {
	if val, ok := e.vars[name]; ok {
		return val, nil
	}

	if e.parent != nil {
		return e.parent.Get(name)
	}

	return nil, fmt.Errorf("undefined variable: %s", name)
}

// Set updates a variable value in the environment or parent scopes
func (e *Environment) Set(name string, value interface{}) error {
	if _, ok := e.vars[name]; ok {
		e.vars[name] = value
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
