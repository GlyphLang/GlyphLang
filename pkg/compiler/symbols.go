package compiler

// SymbolScope represents the scope level of a symbol
type SymbolScope int

const (
	GlobalScope SymbolScope = iota
	RouteScope
	FunctionScope
	BlockScope
)

// SymbolSource identifies where a symbol was introduced. It mirrors the
// runtime BindingSource enum in the interpreter package so the compile-time
// diagnostics can mention path/query parameter collisions with the same
// wording (see issue #235).
type SymbolSource int

const (
	// SourceUser is a variable declared by user code (default).
	SourceUser SymbolSource = iota
	// SourcePathParam is a symbol bound from a route path parameter.
	SourcePathParam
	// SourceQueryParam is a symbol bound from a route query parameter.
	SourceQueryParam
)

// Symbol represents a variable in the symbol table
type Symbol struct {
	Name        string
	Scope       SymbolScope
	Index       int  // Index in constant pool for the name
	IsDefined   bool // Whether this symbol has been assigned a value
	ConstantIdx int  // Index of the constant if it's a constant value
	IsConstant  bool // Whether this is a compile-time constant
	IsBuiltin   bool // Whether this is a built-in variable (query, input, ws, auth)
	Source      SymbolSource
}

// SymbolTable manages variable symbols and scopes
type SymbolTable struct {
	parent  *SymbolTable
	symbols map[string]*Symbol
	scope   SymbolScope
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable(parent *SymbolTable, scope SymbolScope) *SymbolTable {
	return &SymbolTable{
		parent:  parent,
		symbols: make(map[string]*Symbol),
		scope:   scope,
	}
}

// NewGlobalSymbolTable creates a new global symbol table
func NewGlobalSymbolTable() *SymbolTable {
	return NewSymbolTable(nil, GlobalScope)
}

// Define adds a new symbol to the current scope
func (st *SymbolTable) Define(name string, nameConstantIndex int) *Symbol {
	symbol := &Symbol{
		Name:      name,
		Scope:     st.scope,
		Index:     nameConstantIndex,
		IsDefined: true,
	}
	st.symbols[name] = symbol
	return symbol
}

// DefineWithSource adds a new symbol to the current scope and tags it with
// the binding source so diagnostics can describe collisions involving
// implicitly bound variables like route path/query parameters (issue #235).
func (st *SymbolTable) DefineWithSource(name string, nameConstantIndex int, source SymbolSource) *Symbol {
	symbol := &Symbol{
		Name:      name,
		Scope:     st.scope,
		Index:     nameConstantIndex,
		IsDefined: true,
		Source:    source,
	}
	st.symbols[name] = symbol
	return symbol
}

// DefineIfAbsentWithSource defines a symbol only if no symbol with the same
// name already exists in the current scope. Returns (symbol, defined) where
// defined is true if a new symbol was created. This is used when registering
// query parameters that might collide with a same-named path parameter; path
// parameters take precedence so we skip the duplicate registration.
func (st *SymbolTable) DefineIfAbsentWithSource(name string, nameConstantIndex int, source SymbolSource) (*Symbol, bool) {
	if existing, ok := st.symbols[name]; ok {
		return existing, false
	}
	return st.DefineWithSource(name, nameConstantIndex, source), true
}

// DefineBuiltin adds a built-in symbol that can be shadowed by user declarations
func (st *SymbolTable) DefineBuiltin(name string, nameConstantIndex int) *Symbol {
	symbol := &Symbol{
		Name:      name,
		Scope:     st.scope,
		Index:     nameConstantIndex,
		IsDefined: true,
		IsBuiltin: true,
	}
	st.symbols[name] = symbol
	return symbol
}

// DefineConstant defines a compile-time constant
func (st *SymbolTable) DefineConstant(name string, nameConstantIndex int, valueConstantIndex int) *Symbol {
	symbol := &Symbol{
		Name:        name,
		Scope:       st.scope,
		Index:       nameConstantIndex,
		IsDefined:   true,
		ConstantIdx: valueConstantIndex,
		IsConstant:  true,
	}
	st.symbols[name] = symbol
	return symbol
}

// Resolve looks up a symbol in the current scope and parent scopes
func (st *SymbolTable) Resolve(name string) (*Symbol, bool) {
	// Check current scope
	if symbol, ok := st.symbols[name]; ok {
		return symbol, true
	}

	// Check parent scopes
	if st.parent != nil {
		return st.parent.Resolve(name)
	}

	return nil, false
}

// ResolveLocal looks up a symbol only in the current scope (no parent lookup)
func (st *SymbolTable) ResolveLocal(name string) (*Symbol, bool) {
	symbol, ok := st.symbols[name]
	return symbol, ok
}

// EnterScope creates a new nested scope
func (st *SymbolTable) EnterScope(scope SymbolScope) *SymbolTable {
	return NewSymbolTable(st, scope)
}

// Symbols returns all symbols in the current scope
func (st *SymbolTable) Symbols() map[string]*Symbol {
	return st.symbols
}

// Parent returns the parent symbol table
func (st *SymbolTable) Parent() *SymbolTable {
	return st.parent
}

// Scope returns the scope level
func (st *SymbolTable) Scope() SymbolScope {
	return st.scope
}
